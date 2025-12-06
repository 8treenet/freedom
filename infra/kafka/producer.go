package kafka

import (
	"strings"

	"github.com/8treenet/freedom"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, producer)
	})
}

// GetProducer Gets an instance of the producer.
func GetProducer() Producer {
	return producer
}

var producer *ProducerImpl = new(ProducerImpl)

// Producer The producer's interface definition.
type Producer interface {
	// Create a new message
	NewMsg(topic string, content []byte) *Msg
	// Start pass in the relevant address, configuration.
	Start(addrs []string, config *sarama.Config)
	// Restart the connection.
	Restart() error
}

// ProducerImpl The realization of the producer.
type ProducerImpl struct {
	freedom.Infra
	syncProducer sarama.SyncProducer
	addrs        []string
	config       *sarama.Config
}

// Start pass in the relevant address, configuration.
func (pi *ProducerImpl) Start(addrs []string, config *sarama.Config) {
	pi.addrs = addrs
	pi.config = config
	pi.config.Producer.Return.Errors = true
	pi.config.Producer.Return.Successes = true
}

// Restart the connection.
func (pi *ProducerImpl) Restart() error {
	if err := pi.Close(); err != nil {
		return err
	}
	return pi.dial()
}

// Booting The method of overriding the component .
// The single-case component initiates a callback.
func (pi *ProducerImpl) Booting(bootManager freedom.BootManager) {
	if len(pi.addrs) == 0 {
		return
	}

	bootManager.RegisterShutdown(func() {
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
	u := uuid.New()
	return strings.ToUpper(strings.ReplaceAll(u.String(), "-", ""))
}

// NewMsg  Create a new message.
func (pi *ProducerImpl) NewMsg(topic string, content []byte) *Msg {
	return &Msg{
		Topic:   topic,
		Content: content,
	}
}
