package kafka

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/general/requests"
	cluster "github.com/8treenet/freedom/infra/kafka/cluster"
	"github.com/Shopify/sarama"
)

const (
	defaultRetrySec = 45
	waitSec         = 2
	XRetryCount     = "x-retry-count"
)

func newRetryHandle(topicPath map[string]string, conf kafkaConf, limiter *Limiter) *retryHandle {
	handle := new(retryHandle)
	handle.limiter = limiter
	handle.topicPath = topicPath
	handle.consumerMap = make(map[*consumerConf]*cluster.Consumer)
	handle.producerMap = make(map[*consumerConf]sarama.SyncProducer)
	handle.kafkaconf = conf
	return handle
}

type retryHandle struct {
	topicPath       map[string]string
	limiter         *Limiter
	consumerMapLock sync.Mutex
	isClose         bool

	consumerMap map[*consumerConf]*cluster.Consumer
	producerMap map[*consumerConf]sarama.SyncProducer
	kafkaconf   kafkaConf
}

func (c *retryHandle) StartProducer(conf *consumerConf) {
	cf := newProducerConfig(producerConf{Servers: conf.Servers, Name: ""})
	if confCallBack != nil {
		confCallBack(cf)
	}

	syncp, err := sarama.NewSyncProducer(conf.Servers, cf)
	if err != nil {
		panic(err)
	}
	freedom.Logger().Debug("Retry producer connect servers: ", conf.Servers)

	c.producerMap[conf] = syncp
}

func (c *retryHandle) StartConsumer(conf *consumerConf, topicNames []string) {
	newTopicNames := []string{}
	for i := 0; i < len(topicNames); i++ {
		newTopicNames = append(newTopicNames, conf.RetryPrefix+topicNames[i])
		freedom.Logger().Debug("Retry consumer listening topic:", conf.RetryPrefix+topicNames[i], ", path:", c.topicPath[topicNames[i]])
	}

	clusterConfig := newConsumerConfig(*conf)
	if confCallBack != nil {
		confCallBack(&clusterConfig.Config)
	}
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	clusterConfig.Consumer.Offsets.AutoCommit.Enable = false

	sec := defaultRetrySec
	if conf.RetryIntervalSec > sec {
		sec = conf.RetryIntervalSec
	}

	for {
		var isClose bool
		c.consumerMapLock.Lock()
		isClose = c.isClose
		c.consumerMapLock.Unlock()
		if isClose {
			return
		}

		begin := time.Now()
		err := c.loopConsumer(conf, newTopicNames, clusterConfig, sec)
		if err != nil {
			freedom.Logger().Error("Retry consumer error: ", err)
		}
		durationMs := time.Now().Sub(begin).Milliseconds()
		sleepMs := int64(sec*1000/2) - durationMs
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}
}

func (c *retryHandle) loopConsumer(conf *consumerConf, topicNames []string, clusterConfig *cluster.Config, sec int) error {
	instance, err := cluster.NewConsumer(conf.Servers, conf.RetryGroupID, topicNames, clusterConfig)
	if err != nil {
		return err
	}

	c.consumerMapLock.Lock()
	c.consumerMap[conf] = instance
	c.consumerMapLock.Unlock()
	freedom.Logger().Debug("Retry consumer connect servers: ", conf.Servers)

	go func() {
		for err := range instance.Errors() {
			freedom.Logger().Error("kafkaConsumer", conf, err)
		}
	}()

	front := time.Now().Add(time.Duration(-sec) * time.Second)
	for {
		breakFor := false
		msgChan := instance.Messages()
		select {
		case msg := <-msgChan:
			if msg == nil {
				breakFor = true
				break
			}
			if msg.Timestamp.After(front) {
				instance.ResetOffset(msg, "")
				continue
			}
			freedom.Logger().Debug("Retry consume topic: ", msg.Topic)
			instance.MarkOffset(msg, "")
			c.limiter.Open(1)
			go c.call(msg, conf)
		case <-time.After(3 * time.Second):
			breakFor = true
		}
		if breakFor {
			break
		}
	}

	instance.Close()
	freedom.Logger().Debug("Retry consumer close complete")
	c.consumerMapLock.Lock()
	delete(c.consumerMap, conf)
	c.consumerMapLock.Unlock()
	return nil
}

func (kc *retryHandle) call(msg *sarama.ConsumerMessage, conf *consumerConf) {
	defer func() {
		kc.limiter.Close(1)
	}()

	if !strings.HasPrefix(msg.Topic, conf.RetryPrefix) {
		freedom.Logger().Error("HasPrefix 'topic' :", msg.Topic, conf.RetryPrefix)
		return
	}

	topic := string(bytes.TrimPrefix([]byte(msg.Topic), []byte(conf.RetryPrefix)))
	path, ok := kc.topicPath[topic]
	if !ok {
		freedom.Logger().Error("Undefined 'topic' :", topic, conf.Servers)
	}

	path = strings.ReplaceAll(path, ":param1", string(msg.Key))
	var request requests.Request
	if kc.kafkaconf.Consumer.ProxyHTTP2 {
		request = requests.NewH2CRequest(kc.kafkaconf.Consumer.ProxyAddr + path)
	} else {
		request = requests.NewHttpRequest(kc.kafkaconf.Consumer.ProxyAddr + path)
	}
	request = request.SetBody(msg.Value)
	for index := 0; index < len(msg.Headers); index++ {
		request = request.SetHeader(string(msg.Headers[index].Key), string(msg.Headers[index].Value))
	}
	request.SetHeader("x-message-key", string(msg.Key))
	_, resp := request.Post().ToString()

	if resp.Error != nil || resp.StatusCode != 200 {
		freedom.Logger().Errorf("Call message processing failed, path:%s, topic:%s, addr:%v, body:%v, error:%v", path, msg.Topic, conf.Servers, string(msg.Value), resp.Error)
		kc.RetryMsg(topic, msg.Value, msg.Key, msg.Headers, conf)
	}
}

func (kc *retryHandle) RetryMsg(topic string, value []byte, msgKey []byte, header []*sarama.RecordHeader, conf *consumerConf) {
	syncProducer := kc.producerMap[conf]
	newTopic := conf.RetryPrefix + topic
	saramaMsg := &sarama.ProducerMessage{
		Topic:     newTopic,
		Key:       sarama.StringEncoder(msgKey),
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now(),
	}

	noRetryCount := true
	for _, hvalue := range header {
		if string(hvalue.Key) != XRetryCount {
			saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: hvalue.Key, Value: hvalue.Value})
			continue
		}
		noRetryCount = false
		i, err := strconv.Atoi(string(hvalue.Value))
		if err != nil {
			freedom.Logger().Error("Failed to retry send message,", "topic:"+newTopic, "content:"+string(value), "error: x-retry-count atoi")
			return
		}
		if i == conf.RetryCount {
			saramaMsg.Topic = conf.RetryFailPrefix + topic
			saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: hvalue.Key, Value: hvalue.Value})
			continue
		}
		i += 1
		saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: []byte(XRetryCount), Value: []byte(strconv.Itoa(i))})
	}
	if noRetryCount {
		saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: []byte(XRetryCount), Value: []byte("1")})
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				freedom.Logger().Error(err)
			}
		}()
		_, _, err := syncProducer.SendMessage(saramaMsg)
		if err == nil {
			freedom.Logger().Debug("Retry produce topic: ", saramaMsg.Topic)
			return
		}
		freedom.Logger().Error("Failed to send message,", "topic:"+newTopic, "content:"+string(value), "error:"+err.Error())
	}()
}

func (c *retryHandle) Close() {
	c.consumerMapLock.Lock()
	c.isClose = true
	for _, instance := range c.consumerMap {
		if err := instance.Close(); err != nil {
			freedom.Logger().Error(err)
		}
	}
	c.consumerMapLock.Unlock()
	for _, instance := range c.producerMap {
		if err := instance.Close(); err != nil {
			freedom.Logger().Error(err)
		} else {
			freedom.Logger().Debug("Retry producer close complete")
		}
	}
}
