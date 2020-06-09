package kafka

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/8treenet/freedom"
	"github.com/Shopify/sarama"
)

// Msg .
type Msg struct {
	httpHeader   http.Header
	topic        string
	key          string
	content      []byte
	headers      map[string]string
	producerName string
}

// SetWorker .
func (msg *Msg) SetWorker(worker freedom.Worker) *Msg {
	freedom.HandleBusMiddleware(worker)
	msg.httpHeader = worker.Bus().Header
	return msg
}

// Publish .
func (msg *Msg) Publish() {
	saramaMsg := &sarama.ProducerMessage{
		Topic:     msg.topic,
		Key:       sarama.StringEncoder(msg.key),
		Value:     sarama.StringEncoder(msg.content),
		Timestamp: time.Now(),
	}
	for key := range msg.httpHeader {
		saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: []byte(key), Value: []byte(msg.httpHeader.Get(key))})
	}

	for key, value := range msg.headers {
		saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{Key: []byte(key), Value: []byte(value)})
	}

	go func() {
		now := time.Now()
		syncProducer := producer.getSaramaProducer(msg.producerName)
		if syncProducer == nil {
			errMsg := fmt.Sprintf("This '%s', no producer found, please check 'infra/kafka.toml'.", msg.topic)
			freedom.Logger().Error(errMsg)
			freedom.Prometheus().KafkaProducerWithLabelValues(msg.topic, errors.New(errMsg), now)
			return
		}
		_, _, err := syncProducer.SendMessage(saramaMsg)
		freedom.Logger().Debug("Produce topic: ", saramaMsg.Topic)
		freedom.Prometheus().KafkaProducerWithLabelValues(msg.topic, err, now)
		if err == nil {
			return
		}
		freedom.Logger().Error("Failed to send message,", "topic:"+msg.topic, "content:"+string(msg.content), "error:"+err.Error())
	}()
}

// SetHeaders .
func (msg *Msg) SetHeaders(headers map[string]string) *Msg {
	if msg.headers == nil {
		msg.headers = headers
		return msg
	}

	for key, value := range headers {
		msg.headers[key] = value
	}
	return msg
}

// SelectClient .
func (msg *Msg) SelectClient(producer string) *Msg {
	msg.producerName = producer
	return msg
}
