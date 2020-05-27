package kafka

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/8treenet/freedom/infra/requests"

	"github.com/8treenet/freedom"
	"github.com/Shopify/sarama"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, consumerPtr)
	})
}

var consumerPtr *Consumer = new(Consumer)

// Consumer .
type Consumer struct {
	topicPath       map[string]string
	limiter         *Limiter
	conf            kafkaConf
	startUpCallBack []func()
	consumerGroups  []sarama.ConsumerGroup
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// StartUp .
func (c *Consumer) StartUp(f func()) {
	c.startUpCallBack = append(c.startUpCallBack, f)
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

	c.Listen()
	for i := 0; i < len(c.startUpCallBack); i++ {
		c.startUpCallBack[i]()
	}
}

func (c *Consumer) Listen() {
	topicNames := []string{}
	for topic, path := range c.topicPath {
		topicNames = append(topicNames, topic)
		freedom.Logger().Debug("Consumer listening topic:", topic, ", path:", path)
	}
	var ctx context.Context

	ctx, c.cancel = context.WithCancel(context.Background())
	for index := 0; index < len(c.conf.Consumers); index++ {
		addrConf := c.conf.Consumers[index]
		cconf := newConsumerConfig(addrConf)
		if confCallBack != nil {
			confCallBack(cconf, c.conf.Other)
		}
		cconf.Consumer.Return.Errors = false
		client, err := sarama.NewConsumerGroup(addrConf.Servers, addrConf.GroupID, cconf)
		if err != nil {
			freedom.Logger().Fatal(err)
		}
		freedom.Logger().Debug("Consumer connect servers: ", addrConf.Servers)
		c.consumerGroups = append(c.consumerGroups, client)
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for {
				// `Consume` should be called inside an infinite loop, when a
				// server-side rebalance happens, the consumer session will need to be
				// recreated to get the new claims
				if err := client.Consume(ctx, topicNames, &consumerHandle{
					consumer: c,
					conf:     &addrConf,
				}); err != nil {
					freedom.Logger().Errorf("Error from consumer: %v", err)
					time.Sleep(5 * time.Second)
				}
				// check if context was cancelled, signaling that the consumer should stop
				if ctx.Err() != nil {
					return
				}
			}
		}()
	}
}

func (c *Consumer) Close() {
	c.cancel()
	c.wg.Wait()
	for _, item := range c.consumerGroups {
		if err := item.Close(); err != nil {
			freedom.Logger().Error(err)
		} else {
			freedom.Logger().Debug("Consumer close complete")
		}
	}
	c.consumerGroups = []sarama.ConsumerGroup{}
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
		request = requests.NewHttpRequest(kc.conf.Consumer.ProxyAddr + path)
	}
	request = request.SetBody(msg.Value)
	for index := 0; index < len(msg.Headers); index++ {
		request = request.AddHeader(string(msg.Headers[index].Key), string(msg.Headers[index].Value))
	}
	request.AddHeader("x-message-key", string(msg.Key))
	_, resp := request.Post().ToString()

	if resp.Error != nil || resp.StatusCode != 200 {
		freedom.Logger().Errorf("Call message processing failed, path:%s, topic:%s, addr:%v, body:%v, error:%v", path, msg.Topic, conf.Servers, string(msg.Value), resp.Error)
	}
}

type consumerHandle struct {
	consumer *Consumer
	conf     *consumerConf
}

func (consumerHandle *consumerHandle) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumerHandle *consumerHandle) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumerHandle *consumerHandle) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		consumerHandle.consumer.limiter.Open(1)
		session.MarkMessage(message, "")
		go consumerHandle.consumer.call(message, consumerHandle.conf)
	}
	return nil
}
