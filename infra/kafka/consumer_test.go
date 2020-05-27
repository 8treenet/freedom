package kafka

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

func TestConsumer(t *testing.T) {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Consumer.Retry.Backoff = 500 * time.Millisecond
	client, err := sarama.NewConsumerGroup([]string{":9092"}, "hahahah", config)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			fmt.Println("client.Consume")
			if err := client.Consume(ctx, []string{"aa"}, &mockConsumerHandle{}); err != nil {
				fmt.Printf("Error from consumer: %v", err)
				time.Sleep(3 * time.Second)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				fmt.Println("ctx.Err() != nil", ctx.Err())
				return
			}
			fmt.Println("client.Consume end")
		}
	}()

	time.Sleep(15 * time.Second)
	cancel()
	client.Close()
}

type mockConsumerHandle struct {
}

func (consumerHandle *mockConsumerHandle) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumerHandle *mockConsumerHandle) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumerHandle *mockConsumerHandle) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		session.MarkMessage(message, "")
		fmt.Println(message)
	}
	return nil
}
