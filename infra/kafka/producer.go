package kafka

import (
	"strings"

	"github.com/8treenet/freedom"
	"github.com/Shopify/sarama"
	uuid "github.com/iris-contrib/go.uuid"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, producer)
	})
}

var producer *ProducerImpl = new(ProducerImpl)
var _ Producer = (*ProducerImpl)(nil)

// Producer .
type Producer interface {
	NewMsg(topic string, content []byte, producerName ...string) *Msg
}

// ProducerImpl .
type ProducerImpl struct {
	freedom.Infra
	saramaProducerMap map[string]sarama.SyncProducer
	defaultProducer   sarama.SyncProducer
	startUpCallBack   []func()
}

// Booting .
func (pi *ProducerImpl) Booting(sb freedom.SingleBoot) {
	pi.saramaProducerMap = make(map[string]sarama.SyncProducer)

	conf := kafkaConf{}
	if err := freedom.Configure(&conf, "infra/kafka.toml"); err != nil {
		panic(err)
	}
	pi.dial(conf)

	sb.RegisterShutdown(func() {
		pi.close()
	})

	for i := 0; i < len(pi.startUpCallBack); i++ {
		pi.startUpCallBack[i]()
	}
}

func (pi *ProducerImpl) dial(conf kafkaConf) {
	if !conf.Producer.Open {
		freedom.Logger().Debug("[Freedom] 'infra/kafka.toml' '[[producer.open]]' is false")
		return
	}
	if len(conf.Producers) == 0 {
		freedom.Logger().Error("[Freedom] 'infra/kafka.toml' file under '[[producer_clients]]' error")
		return
	}
	for index := 0; index < len(conf.Producers); index++ {
		c := newProducerConfig(conf.Producers[index])
		if confCallBack != nil {
			confCallBack(c, conf.Other)
		}
		syncp, err := sarama.NewSyncProducer(conf.Producers[index].Servers, c)
		if err != nil {
			panic(err)
		}
		freedom.Logger().Debug("[Freedom] Producer connect servers: ", conf.Producers[index].Servers)

		if conf.Producers[index].Name == "" {
			pi.defaultProducer = syncp
		}
		pi.saramaProducerMap[conf.Producers[index].Name] = syncp
	}
}

func (pi *ProducerImpl) close() {
	for _, producer := range pi.saramaProducerMap {
		if err := producer.Close(); err != nil {
			freedom.Logger().Error(err)
		} else {
			freedom.Logger().Debug("[Freedom]Producer close complete")
		}
	}
}

// getSaramaProducer .
func (pi *ProducerImpl) getSaramaProducer(name string) sarama.SyncProducer {
	if name == "" {
		return pi.defaultProducer
	}

	result, ok := pi.saramaProducerMap[name]
	if !ok {
		panic("[Freedom] The instance does not exist name:" + name)
	}
	return result
}

// generateMessageKey
func (pi *ProducerImpl) generateMessageKey() string {
	u, _ := uuid.NewV1()
	return strings.ToUpper(strings.ReplaceAll(u.String(), "-", ""))
}

// NewMsg .
func (pi *ProducerImpl) NewMsg(topic string, content []byte, producerName ...string) *Msg {
	pName := ""
	if len(producerName) > 0 {
		pName = producerName[0]
	}
	return &Msg{
		Topic:        topic,
		Content:      content,
		producerName: pName,
	}
}
