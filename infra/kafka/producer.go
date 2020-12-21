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

// GetProducer .
func GetProducer() Producer {
	return producer
}

var producer *ProducerImpl = new(ProducerImpl)

// Producer .
type Producer interface {
	NewMsg(topic string, content []byte) *Msg
	Start(addrs []string, config *sarama.Config)
	Restart() error
}

// ProducerImpl .
type ProducerImpl struct {
	freedom.Infra
	syncProducer sarama.SyncProducer
	addrs        []string
	config       *sarama.Config
}

// Start .
func (pi *ProducerImpl) Start(addrs []string, config *sarama.Config) {
	pi.addrs = addrs
	pi.config = config
	pi.config.Producer.Return.Errors = true
	pi.config.Producer.Return.Successes = true
}

// Restart .
func (pi *ProducerImpl) Restart() error {
	if err := pi.Close(); err != nil {
		return err
	}
	return pi.dial()
}

// Booting .
func (pi *ProducerImpl) Booting(sb freedom.SingleBoot) {
	if len(pi.addrs) == 0 {
		return
	}

	sb.RegisterShutdown(func() {
		if err := pi.Close(); err != nil {
			freedom.Logger().Error(err)
		}
	})
	if err := pi.dial(); err != nil {
		panic(err)
	}
}

func (pi *ProducerImpl) dial() error {
	syncp, err := sarama.NewSyncProducer(pi.addrs, pi.config)
	if err != nil {
		return err
	}
	pi.syncProducer = syncp
	freedom.Logger().Debug("[Freedom] Producer connect servers: ", pi.addrs)
	return nil
}

// Close .
func (pi *ProducerImpl) Close() error {
	if pi.syncProducer == nil {
		return nil
	}

	defer func() {
		pi.syncProducer = nil
	}()
	return pi.syncProducer.Close()
}

// generateMessageKey
func (pi *ProducerImpl) generateMessageKey() string {
	u, _ := uuid.NewV1()
	return strings.ToUpper(strings.ReplaceAll(u.String(), "-", ""))
}

// NewMsg .
func (pi *ProducerImpl) NewMsg(topic string, content []byte) *Msg {
	return &Msg{
		Topic:   topic,
		Content: content,
	}
}
