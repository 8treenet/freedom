package kafka

import (
	"fmt"
	"net/http"
	"time"

	"github.com/8treenet/freedom"
	"github.com/Shopify/sarama"
)

// Msg .
type Msg struct {
	httpHeader   http.Header
	Topic        string
	key          string
	Content      []byte
	header       map[string]string
	producerName string
	stop         bool
	nextIndex    int
	sendErr      error
}

// SetWorker .
func (msg *Msg) SetWorker(worker freedom.Worker) *Msg {
	freedom.HandleBusMiddleware(worker)
	msg.httpHeader = worker.Bus().Header
	return msg
}

// Publish .
func (msg *Msg) Publish() {
	go func() {
		msg.Next()
	}()
}

// SetHeader .
func (msg *Msg) SetHeader(head map[string]string) *Msg {
	if msg.header == nil {
		msg.header = head
		return msg
	}

	for key, value := range head {
		msg.header[key] = value
	}
	return msg
}

// SelectClient .
func (msg *Msg) SelectClient(producer string) *Msg {
	msg.producerName = producer
	return msg
}

// Next .
func (msg *Msg) Next() {
	if len(middlewares) == 0 {
		msg.sendErr = msg.do()
		return
	}
	if msg.IsStopped() {
		return
	}
	if msg.nextIndex == len(middlewares) {
		msg.sendErr = msg.do()
		return
	}
	msg.nextIndex = msg.nextIndex + 1
	middlewares[msg.nextIndex-1](msg)
}

// IsStopped .
func (msg *Msg) IsStopped() bool {
	return msg.stop
}

// Stop .
func (msg *Msg) Stop() *Msg {
	msg.stop = true
	return msg
}

// GetExecution .
func (msg *Msg) GetExecution() error {
	return msg.sendErr
}

// GetMessageKey .
func (msg *Msg) GetMessageKey() string {
	return msg.key
}

// SetMessageKey .
func (msg *Msg) SetMessageKey(key string) *Msg {
	msg.key = key
	return msg
}

// GetHeader .
func (msg *Msg) GetHeader() map[string]string {
	return msg.header
}

func (msg *Msg) do() error {
	saramaMsg := &sarama.ProducerMessage{
		Topic:     msg.Topic,
		Key:       sarama.StringEncoder(msg.key),
		Value:     sarama.StringEncoder(msg.Content),
		Timestamp: time.Now(),
	}
	for key := range msg.httpHeader {
		saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: []byte(key), Value: []byte(msg.httpHeader.Get(key))})
	}

	for key, value := range msg.header {
		saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: []byte(key), Value: []byte(value)})
	}

	syncProducer := producer.getSaramaProducer(msg.producerName)
	if syncProducer == nil {
		errMsg := fmt.Sprintf("This '%s', no producer found, please check 'infra/kafka.toml'.", msg.Topic)
		freedom.Logger().Error("[Freedom] " + errMsg)
		return nil
	}
	_, _, err := syncProducer.SendMessage(saramaMsg)
	freedom.Logger().Debug("[Freedom] Produce topic: ", saramaMsg.Topic)
	if err == nil {
		return nil
	}
	freedom.Logger().Error("[Freedom] Failed to send message,", "topic:"+msg.Topic, "content:"+string(msg.Content), "error:"+err.Error())
	return err
}
