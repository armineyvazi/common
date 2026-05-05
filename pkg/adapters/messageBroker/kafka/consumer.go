package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/IBM/sarama"

	"github.com/armineyvazi/common.git/pkg/ports"
)

type kafkaConsumer struct {
	log      ports.LoggerWithTraceID
	dsn      string
	username string
	password string
}

func (kc *kafkaConsumer) ServiceName() string {
	return fmt.Sprintf("kafka_%s", kc.dsn)
}

func (kc *kafkaConsumer) IsHealthy(ctx context.Context) bool {
	return true
}

func NewConsumer(log ports.LoggerWithTraceID, dsn, username, password string) ports.KafkaConsumer {
	return &kafkaConsumer{
		log:      log,
		dsn:      dsn,
		username: username,
		password: password,
	}
}

func (kc *kafkaConsumer) Consume(ctx context.Context, topics, groupId string, resultChan chan []byte) error {

	kc.log.Info(ctx, "Starting a new consumer")

	version, err := sarama.ParseKafkaVersion(sarama.DefaultVersion.String())
	if err != nil {
		kc.log.Panic(ctx, fmt.Sprintf("Error parsing Kafka version: %v", err))
	}

	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Net.SASL.Enable = true
	config.Net.SASL.User = kc.username
	config.Net.SASL.Password = kc.password
	config.Net.SASL.Handshake = true
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := consumerGroupHandler{
		log:        kc.log,
		ready:      make(chan bool),
		resultChan: resultChan,
	}

	client, err := sarama.NewConsumerGroup(strings.Split(kc.dsn, ","), groupId, config)
	if err != nil {
		kc.log.Panic(ctx, fmt.Sprintf("Error creating consumer group client: %v", err))
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {

			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					kc.log.Error(ctx, err.Error())
					return
				}
				kc.log.Panic(ctx, fmt.Sprintf("Error from consumer: %v", err))
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				kc.log.Error(ctx, ctx.Err().Error())
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	kc.log.Info(ctx, "consumer up and running!...")

	return nil
}

func (kc *kafkaConsumer) Close() error {
	return nil
}

type consumerGroupHandler struct {
	log        ports.LoggerWithTraceID
	resultChan chan []byte
	ready      chan bool
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				h.log.Error(session.Context(), "message channel was closed")
				return nil
			}

			msg := ports.KafkaMessage{
				Value:  message.Value,
				Offset: message.Offset,
				Topic:  message.Topic,
			}

			msgB, err := json.Marshal(msg)
			if err != nil {
				h.log.Error(session.Context(), err.Error())
				continue
			}

			h.resultChan <- msgB
			h.log.Debug(session.Context(), fmt.Sprintf("offset: %d - topic: %s", message.Offset, message.Topic))
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
