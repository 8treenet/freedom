package kafka

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Shopify/sarama"
)

// Msg .
type Msg struct {
	httpHeader http.Header
	Topic      string
	key        string
	Content    []byte
	header     map[string]interface{}
	stop       bool
	nextIndex  int
	sendErr    error
}

// Publish .
func (msg *Msg) Publish() error {
	msg.Next()
	return msg.sendErr
}

// SetHeader .
func (msg *Msg) SetHeader(head map[string]interface{}) *Msg {
	if msg.header == nil {
		msg.header = head
		return msg
	}

	for key, value := range head {
		msg.header[key] = value
	}
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
func (msg *Msg) GetHeader() map[string]interface{} {
	return msg.header
}

func (msg *Msg) do() error {
	if msg.key == "" {
		msg.key = producer.generateMessageKey()
	}
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
		saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: []byte(key), Value: []byte(fmt.Sprint(value))})
	}

	_, _, err := producer.syncProducer.SendMessage(saramaMsg)
	return err
}
