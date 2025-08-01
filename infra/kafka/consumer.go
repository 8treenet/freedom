package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/ratelimit"

	"github.com/8treenet/freedom/infra/requests"

	"github.com/8treenet/freedom"
	"github.com/Shopify/sarama"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, consumerPtr)
	})
}

// GetConsumer Returns the Consumer instance.
func GetConsumer() Consumer {
	return consumerPtr
}

var consumerPtr *ConsumerImpl = new(ConsumerImpl)

// Consumer Kafka Consumer interface definition.
type Consumer interface {
	// Start pass in the relevant address, configuration.
	Start(addrs []string, groupID string, config *sarama.Config, proxyAddr string, proxyH2C bool)
	// Restart the connection.
	Restart() error
	// Close the connection.
	Close() error
	// Set the limit flow per second.
	SetRateLimit(rate int)
	// Set the HTTP agent to time out.
	SetProxyTimeout(time.Duration)
	// SetSerializable .
	SetSerializable()
}

// ConsumerImpl Kafka Consumer implementation.
type ConsumerImpl struct {
	freedom.Infra
	topicPath    map[string]string
	config       *sarama.Config
	client       sarama.ConsumerGroup
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	proxyH2C     bool
	proxyAddr    string
	addrs        []string
	groupID      string
	limiter      ratelimit.Limiter
	serializable bool
	rate         int
	proxyTimeout time.Duration
	h2cClient    requests.Client
	httpClient   requests.Client
}

// Start pass in the relevant address, configuration.
func (c *ConsumerImpl) Start(addrs []string, groupID string, config *sarama.Config, proxyAddr string, proxyH2C bool) {
	c.addrs = addrs
	c.groupID = groupID
	c.config = config
	c.proxyAddr = proxyAddr
	c.proxyH2C = proxyH2C
	c.config.Consumer.Return.Errors = false
	c.rate = 800
	c.proxyTimeout = 60 * time.Second
}

// SetRateLimit Set the limit flow per second.
func (c *ConsumerImpl) SetRateLimit(rate int) {
	c.rate = rate
	return
}

// SetSerializable .
func (c *ConsumerImpl) SetSerializable() {
	c.serializable = true
}

// SetProxyTimeout Set the HTTP agent to time out.
func (c *ConsumerImpl) SetProxyTimeout(timeout time.Duration) {
	c.proxyTimeout = timeout
	return
}

// Restart the connection.
func (c *ConsumerImpl) Restart() error {
	if err := c.Close(); err != nil {
		return err
	}
	return c.listen()
}

// Booting The method of overriding the component .
// The single-case component initiates a callback.
func (c *ConsumerImpl) Booting(bootManager freedom.BootManager) {
	if len(c.addrs) == 0 {
		return
	}
	c.h2cClient = requests.NewH2CClient(c.proxyTimeout, 5*time.Second)
	c.httpClient = requests.NewHTTPClient(c.proxyTimeout, 5*time.Second)
	c.limiter = ratelimit.New(c.rate)

	c.topicPath = bootManager.EventsPath(c)
	bootManager.RegisterShutdown(func() {
		if err := c.Close(); err != nil {
			freedom.Logger().Error(err)
		}
	})

	if err := c.listen(); err != nil {
		panic(err)
	}
}

// listen .
func (c *ConsumerImpl) listen() error {
	topicNames := []string{}
	for topic, path := range c.topicPath {
		topicNames = append(topicNames, topic)
		freedom.Logger().Debug("[Freedom] Consumer listening topic:", topic, ", path:", path)
	}
	var ctx context.Context

	ctx, c.cancel = context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(c.addrs, c.groupID, c.config)
	if err != nil {
		return err
	}
	freedom.Logger().Debug("[Freedom] Consumer connect servers: ", c.addrs)
	c.client = client
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, topicNames, &consumerHandle{
				consumer: c,
			}); err != nil {
				freedom.Logger().Errorf("[Freedom] Error from consumer: %v", err)
				time.Sleep(5 * time.Second)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()
	return nil
}

// Close the connection.
func (c *ConsumerImpl) Close() error {
	if c.client == nil {
		return nil
	}

	c.cancel()
	c.wg.Wait()
	defer func() {
		c.cancel = nil
		c.client = nil
	}()
	return c.client.Close()
}

func (c *ConsumerImpl) do(msg *sarama.ConsumerMessage) (e error) {
	defer func() {
		if err := recover(); err != nil {
			freedom.Logger().Error(err)
			e = fmt.Errorf("%v", err)
			return
		}
	}()

	path, ok := c.topicPath[msg.Topic]
	if !ok {
		freedom.Logger().Error("[Freedom] Undefined 'topic' :", msg.Topic)
		return
	}
	var request requests.Request
	if c.proxyH2C {
		request = requests.NewH2CRequest(c.proxyAddr + path).SetClient(c.h2cClient)
	} else {
		request = requests.NewHTTPRequest(c.proxyAddr + path).SetClient(c.httpClient)
	}

	request = request.SetBody(msg.Value)
	for index := 0; index < len(msg.Headers); index++ {
		request = request.SetHeaderValue(string(msg.Headers[index].Key), string(msg.Headers[index].Value))
	}
	request.SetHeaderValue("x-message-key", string(msg.Key))
	_, resp := request.Post().ToString()

	if resp.Error != nil || resp.StatusCode != 200 {
		freedom.Logger().Errorf("[Freedom] Call message processing failed, path:%s, topic:%s, addr:%v, body:%v, error:%v", path, msg.Topic, c.addrs, string(msg.Value), resp.Error)
		e = fmt.Errorf("http code:%d, err:%w", resp.StatusCode, resp.Error)
	}
	return
}

type consumerHandle struct {
	consumer *ConsumerImpl
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
		if consumerHandle.consumer.serializable {
			if err := consumerHandle.consumer.do(message); err != nil {
				continue
			}
			session.MarkMessage(message, "")
			continue
		}

		consumerHandle.consumer.limiter.Take()
		session.MarkMessage(message, "")
		go consumerHandle.consumer.do(message)
	}
	return nil
}
