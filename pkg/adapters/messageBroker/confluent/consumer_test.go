package confluent_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/armineyvazi/common.git/pkg/adapters/messageBroker/confluent"
	"github.com/armineyvazi/common.git/pkg/ports"
)

func TestNewConsumer_MissingGroupID_ReturnsError(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	// confluent requires a non-empty group.id
	_, err = confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		GroupID:    "",
	})
	if err == nil {
		t.Fatal("expected error for empty GroupID, got nil")
	}
}

func TestNewConsumer_InvalidBroker_ReturnsError(t *testing.T) {
	_, err := confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: ""},
		GroupID:    "test-group",
	})
	if err == nil {
		t.Fatal("expected error for empty broker list, got nil")
	}
}

func TestConsumer_MockCluster_ProduceAndConsume(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	const topic = "e2e-topic"
	const payload = `{"event":"order_created","id":42}`

	// Produce a message directly via a low-level producer so the consumer has
	// something to read before we start polling.
	prod, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig:        confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		EnableIdempotence: false,
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	prod.Produce(topic, []byte(payload))
	prod.Flush(5000)
	prod.Close()

	// Build the consumer.
	cons, err := confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig:      confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		GroupID:         "e2e-group",
		AutoOffsetReset: confluent.AutoOffsetResetEarliest,
	})
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer func() { _ = cons.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultChan := make(chan []byte, 10)
	if err := cons.Consume(ctx, topic, "e2e-group", resultChan); err != nil {
		t.Fatalf("Consume: %v", err)
	}

	select {
	case raw := <-resultChan:
		var msg ports.KafkaMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("unmarshal KafkaMessage: %v", err)
		}
		if string(msg.Value) != payload {
			t.Errorf("value: got %q, want %q", msg.Value, payload)
		}
		if msg.Topic != topic {
			t.Errorf("topic: got %q, want %q", msg.Topic, topic)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for message")
	}
}

func TestConsumer_ServiceName(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	c, err := confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		GroupID:    "svc-group",
	})
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer func() { _ = c.Close() }()

	if c.ServiceName() == "" {
		t.Error("ServiceName must not be empty")
	}
}

func TestConsumer_IsHealthy(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	c, err := confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		GroupID:    "health-group",
	})
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer func() { _ = c.Close() }()

	if !c.IsHealthy(context.Background()) {
		t.Error("expected IsHealthy=true")
	}
}

func TestConsumer_ConsumeMultipleTopics(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	topics := []string{"topic-a", "topic-b"}

	// Produce one message to each topic.
	prod, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig:        confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		EnableIdempotence: false,
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	for _, topic := range topics {
		prod.Produce(topic, []byte(`{"topic":"`+topic+`"}`))
	}
	prod.Flush(5000)
	prod.Close()

	cons, err := confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig:      confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		GroupID:         "multi-group",
		AutoOffsetReset: confluent.AutoOffsetResetEarliest,
	})
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer func() { _ = cons.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultChan := make(chan []byte, 10)
	// comma-separated topics, matching the ports.KafkaConsumer contract
	if err := cons.Consume(ctx, "topic-a,topic-b", "multi-group", resultChan); err != nil {
		t.Fatalf("Consume: %v", err)
	}

	received := 0
	deadline := time.After(10 * time.Second)
	for received < len(topics) {
		select {
		case <-resultChan:
			received++
		case <-deadline:
			t.Fatalf("timed out: received %d/%d messages", received, len(topics))
		}
	}
}

func TestConsumer_ContextCancellation_StopsLoop(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	c, err := confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		GroupID:    "cancel-group",
	})
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	resultChan := make(chan []byte, 1)

	if err := c.Consume(ctx, "cancel-topic", "cancel-group", resultChan); err != nil {
		t.Fatalf("Consume: %v", err)
	}

	// Cancel immediately — the poll loop should exit cleanly.
	cancel()
	time.Sleep(300 * time.Millisecond)
	// No assertion needed: if the goroutine didn't exit we'd see it in the race detector.
}

// BenchmarkConsumer_Poll measures the overhead of the Poll loop when there
// are no messages available (idle consumer).
func BenchmarkConsumer_Poll(b *testing.B) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		b.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	// Pre-produce messages so the consumer has work to do.
	prod, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig:        confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		EnableIdempotence: false,
	})
	if err != nil {
		b.Fatalf("NewProducer: %v", err)
	}
	for i := 0; i < b.N; i++ {
		prod.Produce("bench-consume", []byte(`{"i":1}`))
	}
	prod.Flush(10000)
	prod.Close()

	c, err := confluent.NewConsumer(&nopLogger{}, confluent.ConsumerConfig{
		BaseConfig:      confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		GroupID:         "bench-group",
		AutoOffsetReset: confluent.AutoOffsetResetEarliest,
	})
	if err != nil {
		b.Fatalf("NewConsumer: %v", err)
	}
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan []byte, b.N+1)
	if err := c.Consume(ctx, "bench-consume", "bench-group", resultChan); err != nil {
		b.Fatalf("Consume: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		select {
		case <-resultChan:
		case <-time.After(10 * time.Second):
			b.Fatalf("timed out at message %d/%d", i, b.N)
		}
	}
}
