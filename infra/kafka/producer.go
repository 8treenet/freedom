package kafka

import (
	"strings"

	"github.com/8treenet/freedom"
	"github.com/Shopify/sarama"
	uuid "github.com/iris-contrib/go.uuid"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, producer)

		initiator.InjectController(func(ctx freedom.Context) (com *Producer) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

var producer *Producer = new(Producer)

// Producer .
type Producer struct {
	saramaProducerMap map[string]sarama.SyncProducer
	defaultProducer   sarama.SyncProducer
}

// Booting .
func (p *Producer) Booting(sb freedom.SingleBoot) {
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
			confCallBack(c)
		}
		syncp, err := sarama.NewSyncProducer(conf.Producers[index].Servers, c)
		if err != nil {
			panic(err)
		}
		if len(conf.Producers) > 1 && conf.Producers[index].Name == "" {
			panic("An instance name is required under multiple instances")
		}

		if len(conf.Producers) == 1 {
			p.defaultProducer = syncp
			break
		}
		p.saramaProducerMap[conf.Producers[index].Name] = syncp
	}
}

// getSaramaProducer .
func (p *Producer) getSaramaProducer(name string) sarama.SyncProducer {
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
func (p *Producer) generateMessageKey() string {
	u, _ := uuid.NewV1()
	return strings.ToUpper(strings.ReplaceAll(u.String(), "-", ""))
}

// NewMsg .
func (p *Producer) NewMsg(topic string, content []byte, producerName ...string) *Msg {
	return &Msg{
		topic:        topic,
		key:          producer.generateMessageKey(),
		content:      content,
		producerName: "",
	}
}
