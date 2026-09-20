package confluent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/armineyvazi/common.git/pkg/ports"
)

// Consumer wraps a confluent-kafka-go consumer and implements
// ports.KafkaConsumer. It supports both auto-commit and manual-commit modes.
type Consumer struct {
	log      ports.LoggerWithTraceID
	cfg      ConsumerConfig
	consumer *kafka.Consumer
}

// NewConsumer creates a confluent Kafka consumer from cfg.
// The consumer does not subscribe to any topics until Consume is called.
func NewConsumer(log ports.LoggerWithTraceID, cfg ConsumerConfig) (*Consumer, error) {
	if cfg.Brokers == "" {
		return nil, fmt.Errorf("confluent consumer: Brokers must not be empty")
	}
	if cfg.GroupID == "" {
		return nil, fmt.Errorf("confluent consumer: GroupID must not be empty")
	}
	cfg = consumerDefaults(cfg)

	cm := buildBaseConfigMap(cfg.BaseConfig)
	cm["group.id"] = cfg.GroupID
	cm["auto.offset.reset"] = string(cfg.AutoOffsetReset)
	cm["enable.auto.commit"] = cfg.EnableAutoCommit
	cm["auto.commit.interval.ms"] = cfg.AutoCommitIntervalMs
	cm["session.timeout.ms"] = cfg.SessionTimeoutMs
	cm["heartbeat.interval.ms"] = cfg.HeartbeatIntervalMs
	cm["max.poll.interval.ms"] = cfg.MaxPollIntervalMs
	cm["fetch.min.bytes"] = cfg.FetchMinBytes
	cm["fetch.max.bytes"] = cfg.FetchMaxBytes

	c, err := kafka.NewConsumer(&cm)
	if err != nil {
		return nil, fmt.Errorf("confluent consumer: create: %w", err)
	}

	return &Consumer{
		log:      log,
		cfg:      cfg,
		consumer: c,
	}, nil
}

// Consume subscribes to the comma-separated topics list and starts delivering
// messages to resultChan. It returns immediately after the subscription is
// confirmed and continues consuming in a background goroutine until ctx is
// cancelled or the consumer is closed.
//
// Messages are delivered as JSON-encoded ports.KafkaMessage values, matching
// the contract of the Sarama-based adapter.
func (c *Consumer) Consume(ctx context.Context, topics, groupID string, resultChan chan []byte) error {
	topicList := strings.Split(topics, ",")
	for i, t := range topicList {
		topicList[i] = strings.TrimSpace(t)
	}

	if err := c.consumer.SubscribeTopics(topicList, c.rebalanceCb(ctx)); err != nil {
		return fmt.Errorf("confluent consumer: subscribe %v: %w", topicList, err)
	}

	c.log.Info(ctx, "confluent consumer: subscribed", "topics", topicList, "group", groupID)

	go c.pollLoop(ctx, resultChan)
	return nil
}

// CommitMessage manually commits the offset for msg. Only relevant when
// EnableAutoCommit is false.
func (c *Consumer) CommitMessage(msg *kafka.Message) error {
	_, err := c.consumer.CommitMessage(msg)
	if err != nil {
		return fmt.Errorf("confluent consumer: commit: %w", err)
	}
	return nil
}

// Assignment returns the current partition assignment.
func (c *Consumer) Assignment() ([]kafka.TopicPartition, error) {
	return c.consumer.Assignment()
}

// Close unsubscribes and closes the underlying consumer.
func (c *Consumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("confluent consumer: close: %w", err)
	}
	return nil
}

// ServiceName satisfies the infrastructure service interface.
func (c *Consumer) ServiceName() string {
	return fmt.Sprintf("confluent_kafka_consumer_%s", c.cfg.Brokers)
}

// IsHealthy returns true when the consumer is initialised.
func (c *Consumer) IsHealthy(_ context.Context) bool {
	return c.consumer != nil
}

// pollLoop is the long-running read loop. It exits when ctx is done or a
// fatal kafka.Error is received.
func (c *Consumer) pollLoop(ctx context.Context, resultChan chan []byte) {
	for {
		select {
		case <-ctx.Done():
			c.log.Info(ctx, "confluent consumer: context cancelled, stopping poll loop")
			return
		default:
		}

		ev := c.consumer.Poll(100) // 100 ms timeout; checks ctx frequently
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			c.handleMessage(ctx, e, resultChan)

		case kafka.Error:
			if e.IsFatal() {
				c.log.Error(ctx, "confluent consumer: fatal error, stopping",
					"code", e.Code(), "err", e)
				return
			}
			if e.Code() != kafka.ErrUnknownTopicOrPart {
				c.log.Warn(ctx, "confluent consumer: transient error",
					"code", e.Code(), "err", e)
			}

		case kafka.AssignedPartitions:
			c.log.Info(ctx, "confluent consumer: partitions assigned",
				"partitions", e.Partitions)

		case kafka.RevokedPartitions:
			c.log.Info(ctx, "confluent consumer: partitions revoked",
				"partitions", e.Partitions)

		case kafka.OffsetsCommitted:
			if e.Error != nil {
				c.log.Warn(ctx, "confluent consumer: offset commit failed", "err", e.Error)
			}
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg *kafka.Message, resultChan chan []byte) {
	km := ports.KafkaMessage{
		Value:  msg.Value,
		Offset: int64(msg.TopicPartition.Offset),
		Topic:  *msg.TopicPartition.Topic,
	}

	b, err := json.Marshal(km)
	if err != nil {
		c.log.Error(ctx, "confluent consumer: marshal message", "err", err)
		return
	}

	select {
	case resultChan <- b:
		c.log.Info(ctx, "confluent consumer: message dispatched",
			"topic", km.Topic,
			"offset", km.Offset)
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
		c.log.Warn(ctx, "confluent consumer: result channel full, dropping message",
			"topic", km.Topic, "offset", km.Offset)
	}
}

func (c *Consumer) rebalanceCb(ctx context.Context) kafka.RebalanceCb {
	return func(_ *kafka.Consumer, event kafka.Event) error {
		switch e := event.(type) {
		case kafka.AssignedPartitions:
			c.log.Info(ctx, "confluent consumer: rebalance — assigned",
				"partitions", e.Partitions)
		case kafka.RevokedPartitions:
			c.log.Info(ctx, "confluent consumer: rebalance — revoked",
				"partitions", e.Partitions)
		}
		return nil
	}
}
