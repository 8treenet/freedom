package kafka

import (
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

func NewMiddleware() ProducerHandler {
	return func(msg *Msg) {
		fmt.Println("Middleware begin")
		fmt.Println("msg", msg.Topic, msg.GetMessageKey(), msg.GetHeader(), string(msg.Content))
		msg.Next()
		fmt.Println("Middleware begin")
	}
}

func NewStopMiddleware() ProducerHandler {
	return func(msg *Msg) {
		fmt.Println("StopMiddleware begin")
		msg.Next() //msg.Stop()
		fmt.Println("StopMiddleware end", msg.GetExecution())
	}
}

func TestProducer(t *testing.T) {
	initTestProducer()
	InstallMiddleware(NewMiddleware())
	InstallMiddleware(NewStopMiddleware())
	producer.NewMsg("event-sell", []byte("hello")).Publish()

	time.Sleep(2 * time.Second)
}

func initTestProducer() {
	kconf := kafkaConf{}
	kconf.Producer.Open = true
	kconf.Producers = append(kconf.Producers, producerConf{
		Servers: []string{":9092"},
	})
	producer.saramaProducerMap = make(map[string]sarama.SyncProducer)
	producer.dial(kconf)
}
