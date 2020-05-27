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

		initiator.InjectController(func(ctx freedom.Context) (com *ProducerImpl) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

func GetDomainEventInfra() freedom.DomainEventInfra {
	return producer
}

var producer *ProducerImpl = new(ProducerImpl)
var _ Producer = new(ProducerImpl)

type Producer interface {
	NewMsg(topic string, content []byte, producerName ...string) *Msg
}

// ProducerImpl .
type ProducerImpl struct {
	saramaProducerMap map[string]sarama.SyncProducer
	defaultProducer   sarama.SyncProducer
	startUpCallBack   []func()
}

// StartUp .
func (c *ProducerImpl) StartUp(f func()) {
	c.startUpCallBack = append(c.startUpCallBack, f)
}

// Booting .
func (p *ProducerImpl) Booting(sb freedom.SingleBoot) {
	p.saramaProducerMap = make(map[string]sarama.SyncProducer)

	conf := kafkaConf{}
	freedom.Configure(&conf, "infra/kafka.toml", true)
	if !conf.Producer.Open {
		freedom.Logger().Debug("'infra/kafka.toml' '[[producer.open]]' is false")
		return
	}
	if len(conf.Producers) == 0 {
		freedom.Logger().Error("'infra/kafka.toml' file under '[[producer_clients]]' error")
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
		freedom.Logger().Debug("Producer connect servers: ", conf.Producers[index].Servers)

		if conf.Producers[index].Name == "" {
			p.defaultProducer = syncp
		}
		p.saramaProducerMap[conf.Producers[index].Name] = syncp
	}

	sb.Closeing(func() {
		p.close()
	})

	for i := 0; i < len(p.startUpCallBack); i++ {
		p.startUpCallBack[i]()
	}
}

func (p *ProducerImpl) close() {
	for _, producer := range p.saramaProducerMap {
		if err := producer.Close(); err != nil {
			freedom.Logger().Error(err)
		} else {
			freedom.Logger().Debug("Producer close complete")
		}
	}
}

// getSaramaProducer .
func (p *ProducerImpl) getSaramaProducer(name string) sarama.SyncProducer {
	if name == "" {
		return p.defaultProducer
	}

	result, ok := p.saramaProducerMap[name]
	if !ok {
		panic("The instance does not exist name:" + name)
	}
	return result
}

// generateMessageKey
func (p *ProducerImpl) generateMessageKey() string {
	u, _ := uuid.NewV1()
	return strings.ToUpper(strings.ReplaceAll(u.String(), "-", ""))
}

// NewMsg .
func (p *ProducerImpl) NewMsg(topic string, content []byte, producerName ...string) *Msg {
	pName := ""
	if len(producerName) > 0 {
		pName = producerName[0]
	}
	return &Msg{
		topic:        topic,
		key:          producer.generateMessageKey(),
		content:      content,
		producerName: pName,
	}
}

// DomainEvent .
func (p *ProducerImpl) DomainEvent(producer, topic string, data []byte, worker freedom.Worker, header ...map[string]string) {
	msg := p.NewMsg(topic, data, producer)
	if len(header) > 0 {
		msg = msg.SetHeaders(header[0])
	}
	msg.SetWorker(worker).Publish()
}
