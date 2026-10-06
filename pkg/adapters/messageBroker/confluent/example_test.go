package confluent_test

import (
	"context"
	"fmt"
	"time"

	"github.com/armineyvazi/common.git/pkg/adapters/messageBroker/confluent"
)

// ExampleNewProducer shows constructing a Confluent Kafka producer and
// publishing messages with different levels of metadata.
func ExampleNewProducer() {
	log := &nopLogger{}

	p, err := confluent.NewProducer(log, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{
			Brokers:          "broker1:9092,broker2:9092",
			SecurityProtocol: confluent.SecurityProtocolSASLSSL,
			SASLMechanism:    confluent.SASLMechanismSCRAMSHA512,
			SASLUsername:     "app-producer",
			SASLPassword:     "secret",
		},
		Acks:              "all",
		EnableIdempotence: true,
		CompressionType:   confluent.CompressionSnappy,
		Retries:           5,
	})
	if err != nil {
		fmt.Println("new producer:", err)
		return
	}
	defer p.Close()

	// Simple publish — fire and forget.
	p.Produce("orders", []byte(`{"id":1,"status":"pending"}`))

	// Publish with a partition key so that all events for the same order
	// go to the same partition (preserves order within a partition).
	p.ProduceWithKey("orders", []byte("order-42"), []byte(`{"id":42,"status":"shipped"}`))

	// Publish with headers for distributed tracing.
	p.ProduceWithHeaders("orders", []byte("order-7"), []byte(`{"id":7}`), map[string][]byte{
		"x-trace-id": []byte("abc-def-123"),
		"x-source":   []byte("checkout-service"),
	})

	// Flush ensures all enqueued messages are delivered before shutdown.
	remaining := p.Flush(5000)
	if remaining != 0 {
		fmt.Printf("%d messages not delivered\n", remaining)
	}
}

// ExampleNewProducer_healthCheck shows how to check whether the producer
// is alive, useful in a readiness probe handler.
func ExampleNewProducer_healthCheck() {
	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: "localhost:9092"},
	})
	if err != nil {
		fmt.Println("init:", err)
		return
	}
	defer p.Close()

	if p.IsHealthy(context.Background()) {
		fmt.Println("producer ready")
	}
}

// ExampleNewConsumer shows subscribing to topics and consuming messages
// from the result channel in a dedicated goroutine.
func ExampleNewConsumer() {
	log := &nopLogger{}

	c, err := confluent.NewConsumer(log, confluent.ConsumerConfig{
		BaseConfig: confluent.BaseConfig{
			Brokers:          "broker1:9092",
			SecurityProtocol: confluent.SecurityProtocolSASLPlaintext,
			SASLMechanism:    confluent.SASLMechanismPlain,
			SASLUsername:     "app-consumer",
			SASLPassword:     "secret",
		},
		GroupID:              "order-processor",
		AutoOffsetReset:      confluent.AutoOffsetResetEarliest,
		EnableAutoCommit:     true,
		AutoCommitIntervalMs: 5000,
	})
	if err != nil {
		fmt.Println("new consumer:", err)
		return
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// resultChan receives JSON-encoded ports.KafkaMessage values.
	resultChan := make(chan []byte, 100)

	// Subscribe to multiple topics (comma-separated).
	if err := c.Consume(ctx, "orders,payments", c.ServiceName(), resultChan); err != nil {
		fmt.Println("consume:", err)
		return
	}

	// Process messages until the context times out or the app shuts down.
	for {
		select {
		case msg := <-resultChan:
			_ = msg // unmarshal into ports.KafkaMessage and process
		case <-ctx.Done():
			return
		}
	}
}

// ExampleNewConsumer_manualCommit shows disabling auto-commit and committing
// offsets only after successfully processing each message — guaranteeing
// at-least-once delivery.
func ExampleNewConsumer_manualCommit() {
	c, err := confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig:       confluent.BaseConfig{Brokers: "localhost:9092"},
		GroupID:          "invoice-generator",
		EnableAutoCommit: false, // manual commit
	})
	if err != nil {
		fmt.Println("new consumer:", err)
		return
	}
	defer c.Close()

	// After processing a kafka.Message, commit its offset explicitly:
	//   if err := c.CommitMessage(msg); err != nil { ... }
	fmt.Println("manual-commit consumer ready")
}
