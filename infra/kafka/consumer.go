package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/8treenet/freedom/infra/requests"

	"github.com/8treenet/freedom"
	"github.com/IBM/sarama"
	"go.uber.org/ratelimit"
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

// ConsumerConfig 消费者配置结构体
type ConsumerConfig struct {
	// Kafka 服务器地址
	Addrs []string
	// 消费者组ID
	GroupID string
	// Sarama 配置
	Config *sarama.Config
	// 代理地址
	ProxyAddr string
	// 是否使用 H2C 代理
	ProxyH2C bool
	// HTTP 请求超时时间
	RequestTimeout time.Duration
	// 速率限制（并行模式下每秒处理的消息数，默认800）
	RateLimit int
}

// Consumer Kafka Consumer interface definition.
type Consumer interface {
	// Start 使用配置结构体启动消费者
	Start(config *ConsumerConfig)
	// Restart the connection.
	Restart() error
	// Close the connection.
	Close() error
}

// ConsumerImpl Kafka Consumer implementation.
type ConsumerImpl struct {
	freedom.Infra
	topicPath       map[string]string
	topicSequential map[string]bool // 存储每个topic的串行/并行配置
	config          *sarama.Config
	client          sarama.ConsumerGroup
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	proxyH2C        bool
	proxyAddr       string
	addrs           []string
	groupID         string
	rateLimit       int
	proxyTimeout    time.Duration
	h2cClient       requests.Client
	httpClient      requests.Client
	// 批处理配置
	limiter ratelimit.Limiter
}

// Start 使用配置结构体启动消费者
func (c *ConsumerImpl) Start(config *ConsumerConfig) {
	c.addrs = config.Addrs
	c.groupID = config.GroupID
	c.config = config.Config
	c.proxyAddr = config.ProxyAddr
	c.proxyH2C = config.ProxyH2C
	c.proxyTimeout = config.RequestTimeout
	if c.proxyTimeout <= 0 {
		c.proxyTimeout = 60 * time.Second
	}

	c.rateLimit = config.RateLimit
	if c.rateLimit <= 0 {
		c.rateLimit = 800 // 默认值
	}

	c.config.Consumer.Return.Errors = false
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
	c.limiter = ratelimit.New(c.rateLimit)

	c.topicPath = bootManager.EventsPath(c)
	c.topicSequential = bootManager.EventsSequential(c)
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
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29

	for message := range claim.Messages() {
		// 检查当前topic的串行/并行配置，如果没有配置则使用全局默认配置
		sequential, exists := consumerHandle.consumer.topicSequential[message.Topic]
		if !exists {
			sequential = true // 使用默认配置
		}

		if sequential {
			// 串行处理
			if err := consumerHandle.consumer.do(message); err != nil {
				continue
			}
			session.MarkMessage(message, "")
			continue
		}

		// 并行处理
		consumerHandle.consumer.limiter.Take()
		session.MarkMessage(message, "")
		go consumerHandle.consumer.do(message)
	}
	return nil
}
