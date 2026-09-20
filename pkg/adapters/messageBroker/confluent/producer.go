package confluent

import (
	"context"
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/armineyvazi/common.git/pkg/ports"
)

// Producer wraps a confluent-kafka-go async producer and implements
// ports.KafkaProducer. Delivery reports are processed in a background
// goroutine; call Close to flush and shut it down cleanly.
type Producer struct {
	log      ports.LoggerWithTraceID
	cfg      ProducerConfig
	producer *kafka.Producer
	wg       sync.WaitGroup
	once     sync.Once
}

// NewProducer creates a confluent Kafka producer from cfg.
// It returns an error rather than panicking so callers can handle
// misconfiguration gracefully at startup.
func NewProducer(log ports.LoggerWithTraceID, cfg ProducerConfig) (*Producer, error) {
	if cfg.Brokers == "" {
		return nil, fmt.Errorf("confluent producer: Brokers must not be empty")
	}
	cfg = producerDefaults(cfg)

	cm := buildBaseConfigMap(cfg.BaseConfig)
	cm["acks"] = cfg.Acks
	cm["enable.idempotence"] = cfg.EnableIdempotence
	cm["retries"] = cfg.Retries
	cm["retry.backoff.ms"] = cfg.RetryBackoffMs
	cm["message.timeout.ms"] = cfg.MessageTimeoutMs
	cm["batch.num.messages"] = cfg.BatchNumMessages
	cm["linger.ms"] = cfg.LingerMs
	cm["queue.buffering.max.messages"] = cfg.QueueBufferingMaxMessages
	cm["message.max.bytes"] = cfg.MaxMessageBytes
	if cfg.CompressionType != "" {
		cm["compression.type"] = string(cfg.CompressionType)
	}

	p, err := kafka.NewProducer(&cm)
	if err != nil {
		return nil, fmt.Errorf("confluent producer: create: %w", err)
	}

	pr := &Producer{
		log:      log,
		cfg:      cfg,
		producer: p,
	}
	pr.startDeliveryLoop()
	return pr, nil
}

// Produce enqueues msg for asynchronous delivery to topic.
// It satisfies ports.KafkaProducer.
func (p *Producer) Produce(topic string, value []byte) {
	p.ProduceWithKey(topic, nil, value)
}

// ProduceWithKey enqueues a message with an explicit partition key.
func (p *Producer) ProduceWithKey(topic string, key, value []byte) {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   key,
		Value: value,
	}
	if err := p.producer.Produce(msg, nil); err != nil {
		p.log.Error(context.Background(), "confluent producer: enqueue failed",
			"topic", topic, "err", err)
	}
}

// ProduceWithHeaders enqueues a message with Kafka headers attached.
func (p *Producer) ProduceWithHeaders(topic string, key, value []byte, headers map[string][]byte) {
	kHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kHeaders = append(kHeaders, kafka.Header{Key: k, Value: v})
	}
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:     key,
		Value:   value,
		Headers: kHeaders,
	}
	if err := p.producer.Produce(msg, nil); err != nil {
		p.log.Error(context.Background(), "confluent producer: enqueue failed",
			"topic", topic, "err", err)
	}
}

// Flush waits for all enqueued messages to be delivered or for timeoutMs to
// elapse. It returns the number of messages still in the queue.
func (p *Producer) Flush(timeoutMs int) int {
	return p.producer.Flush(timeoutMs)
}

// Close flushes outstanding messages (up to 30 s) and shuts down the
// delivery goroutine. It is safe to call Close more than once.
func (p *Producer) Close() {
	p.once.Do(func() {
		p.producer.Flush(30000)
		p.producer.Close()
		p.wg.Wait()
	})
}

// ServiceName satisfies the infrastructure service interface used for health checks.
func (p *Producer) ServiceName() string {
	return fmt.Sprintf("confluent_kafka_producer_%s", p.cfg.Brokers)
}

// IsHealthy returns true when the underlying producer is running.
func (p *Producer) IsHealthy(_ context.Context) bool {
	return p.producer != nil
}

// startDeliveryLoop reads from the producer's Events channel and logs
// delivery reports and errors. The goroutine exits when the channel closes.
func (p *Producer) startDeliveryLoop() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ctx := context.Background()
		for e := range p.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					p.log.Error(ctx, "confluent producer: delivery failed",
						"topic", *ev.TopicPartition.Topic,
						"partition", ev.TopicPartition.Partition,
						"err", ev.TopicPartition.Error)
				} else {
					p.log.Info(ctx, "confluent producer: message delivered",
						"topic", *ev.TopicPartition.Topic,
						"partition", ev.TopicPartition.Partition,
						"offset", ev.TopicPartition.Offset)
				}
			case kafka.Error:
				p.log.Error(ctx, "confluent producer: kafka error",
					"code", ev.Code(), "err", ev)
			}
		}
	}()
}
