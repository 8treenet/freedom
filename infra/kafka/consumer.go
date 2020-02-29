package kafka

import (
	"runtime"
	"strings"

	"github.com/8treenet/freedom/general/requests"

	"github.com/8treenet/freedom"
	cluster "github.com/8treenet/freedom/infra/kafka/cluster"
	"github.com/Shopify/sarama"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, consumerPtr)
	})
}

var consumerPtr *Consumer = new(Consumer)

// Consumer .
type Consumer struct {
	saramaConsumers []*cluster.Consumer
	topicPath       map[string]string
	limiter         *Limiter
	conf            kafkaConf
}

// Booting .
func (c *Consumer) Booting(sb freedom.SingleBoot) {
	c.limiter = newLimiter(int32(runtime.NumCPU() * 2048))
	c.topicPath = sb.EventsPath(c)
	freedom.Configure(&c.conf, "infra/kafka.toml", true)
	if !c.conf.Consumer.Open {
		freedom.Logger().Debug("'infra/kafka.toml' '[[consumer.open]]' is false")
		return
	}
	if len(c.conf.Consumers) == 0 {
		freedom.Logger().Error("'infra/kafka.toml' file under '[[consumer_clients]]' error")
		return
	}
	sb.Closeing(func() {
		c.Close()
	})
	c.ReListen()
}

func (c *Consumer) ReListen() {
	topicNames := []string{}
	for topic, paths := range c.topicPath {
		topicNames = append(topicNames, topic)
		for _, path := range paths {
			freedom.Logger().Debug("Consumer listening topic:", topic, ", path:", path)
		}
	}
	for index := 0; index < len(c.conf.Consumers); index++ {
		cconf := newConsumerConfig(c.conf.Consumers[index])
		if confCallBack != nil {
			confCallBack(&cconf.Config)
		}
		instance, err := cluster.NewConsumer(c.conf.Consumers[index].Servers, c.conf.Consumers[index].GroupID, topicNames, cconf)
		if err != nil {
			panic(err)
		}
		c.saramaConsumers = append(c.saramaConsumers, instance)
		c.consume(instance, &c.conf.Consumers[index])
	}
}

func (c *Consumer) Close() {
	for _, instance := range c.saramaConsumers {
		if err := instance.Close(); err != nil {
			freedom.Logger().Error(err)
		}
	}
	c.saramaConsumers = []*cluster.Consumer{}
}

// consume
func (kc *Consumer) consume(cluster *cluster.Consumer, conf *consumerConf) {
	go func() {
		for msg := range cluster.Messages() {
			cluster.MarkOffset(msg, "")
			kc.limiter.Open(1)
			go kc.call(msg, conf)
		}
	}()

	go func() {
		for err := range cluster.Errors() {
			freedom.Logger().Error("kafkaConsumer", conf, err)
		}
	}()
}

func (kc *Consumer) call(msg *sarama.ConsumerMessage, conf *consumerConf) {
	defer func() {
		kc.limiter.Close(1)
	}()
	path, ok := kc.topicPath[msg.Topic]
	if !ok {
		freedom.Logger().Error("Undefined 'topic' :", msg.Topic, conf.Servers)
	}
	path = strings.ReplaceAll(path, ":param1", string(msg.Key))
	var request requests.Request
	if kc.conf.Consumer.ProxyHTTP2 {
		request = requests.NewH2CRequest(kc.conf.Consumer.ProxyAddr + path)
	} else {
		request = requests.NewFastRequest(kc.conf.Consumer.ProxyAddr + path)
	}
	request = request.SetBody(msg.Value)
	for index := 0; index < len(msg.Headers); index++ {
		request = request.SetHeader(string(msg.Headers[index].Key), string(msg.Headers[index].Value))
	}
	request.SetHeader("x-message-key", string(msg.Key))
	_, resp := request.Post().ToString()

	if resp.Error != nil || resp.StatusCode != 200 {
		freedom.Logger().Errorf("Call message processing failed, path:%s, topic:%s, addr:%v, body:%v, error:%v", path, msg.Topic, conf.Servers, msg.Value, resp.Error)
	}
}
