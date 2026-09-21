package confluent_test

import (
	"context"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/armineyvazi/common.git/pkg/adapters/messageBroker/confluent"
)

// nopLogger satisfies ports.LoggerWithTraceID without importing the full adapter stack.
type nopLogger struct{}

func (n *nopLogger) Debug(ctx context.Context, msg string, params ...any)  {}
func (n *nopLogger) Info(ctx context.Context, msg string, params ...any)   {}
func (n *nopLogger) Warn(ctx context.Context, msg string, params ...any)   {}
func (n *nopLogger) Error(ctx context.Context, msg string, params ...any)  {}
func (n *nopLogger) Panic(ctx context.Context, msg string, params ...any)  {}
func (n *nopLogger) Errorw(ctx context.Context, msg string, params ...any) {}
func (n *nopLogger) Warnw(ctx context.Context, msg string, params ...any)  {}
func (n *nopLogger) Flush() error                                          { return nil }

func TestNewProducer_InvalidBroker_ReturnsError(t *testing.T) {
	_, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: ""},
	})
	if err == nil {
		t.Fatal("expected error for empty broker list, got nil")
	}
}

func TestProducer_MockCluster_Produce(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer p.Close()

	p.Produce("test-topic", []byte(`{"event":"test"}`))
	// Flush gives the broker time to accept the message.
	remaining := p.Flush(5000)
	if remaining != 0 {
		t.Errorf("Flush: %d messages still pending", remaining)
	}
}

func TestProducer_MockCluster_ProduceWithKey(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer p.Close()

	p.ProduceWithKey("keyed-topic", []byte("key-1"), []byte(`{"id":1}`))
	if remaining := p.Flush(5000); remaining != 0 {
		t.Errorf("Flush: %d messages pending", remaining)
	}
}

func TestProducer_MockCluster_ProduceWithHeaders(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer p.Close()

	headers := map[string][]byte{
		"x-trace-id": []byte("abc-123"),
		"x-source":   []byte("order-service"),
	}
	p.ProduceWithHeaders("header-topic", nil, []byte(`{"msg":"hello"}`), headers)
	if remaining := p.Flush(5000); remaining != 0 {
		t.Errorf("Flush: %d messages pending", remaining)
	}
}

func TestProducer_MockCluster_HighVolume(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig:        confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		BatchNumMessages:  500,
		LingerMs:          1,
		EnableIdempotence: false, //nolint:staticcheck
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer p.Close()

	const n = 500
	for i := 0; i < n; i++ {
		p.Produce("bulk-topic", []byte(`{"i":1}`))
	}
	if remaining := p.Flush(10000); remaining != 0 {
		t.Errorf("Flush: %d messages still pending after high-volume produce", remaining)
	}
}

func TestProducer_ServiceName(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer p.Close()

	if p.ServiceName() == "" {
		t.Error("ServiceName must not be empty")
	}
}

func TestProducer_IsHealthy(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	defer p.Close()

	if !p.IsHealthy(context.Background()) {
		t.Error("expected IsHealthy=true after successful construction")
	}
}

func TestProducer_Close_Idempotent(t *testing.T) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		t.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig: confluent.BaseConfig{Brokers: mc.BootstrapServers()},
	})
	if err != nil {
		t.Fatalf("NewProducer: %v", err)
	}
	p.Close()
	p.Close() // must not panic
}

// BenchmarkProducer_Produce measures enqueue throughput against a mock cluster.
func BenchmarkProducer_Produce(b *testing.B) {
	mc, err := kafka.NewMockCluster(1)
	if err != nil {
		b.Skipf("MockCluster unavailable: %v", err)
	}
	defer mc.Close()

	p, err := confluent.NewProducer(&nopLogger{}, confluent.ProducerConfig{
		BaseConfig:        confluent.BaseConfig{Brokers: mc.BootstrapServers()},
		EnableIdempotence: false,
		BatchNumMessages:  10000,
		LingerMs:          5,
	})
	if err != nil {
		b.Fatalf("NewProducer: %v", err)
	}
	defer p.Close()

	payload := []byte(`{"event":"bench","id":1}`)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p.Produce("bench-topic", payload)
	}

	// Drain after the benchmark loop.
	done := make(chan struct{})
	go func() {
		p.Flush(10000)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		b.Error("flush timed out")
	}
}
