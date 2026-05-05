package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
	"github.com/armineyvazi/common.git/pkg/ports"
)

type kafkaProducer struct {
	log      ports.LoggerWithTraceID
	dsn      string
	producer sarama.AsyncProducer
}

type producerProvider struct {
	producerProvider func() sarama.AsyncProducer
}

func (kp *kafkaProducer) ServiceName() string {
	return fmt.Sprintf("kafka_%s", kp.dsn)
}

func (kp *kafkaProducer) IsHealthy(ctx context.Context) bool {
	return true
}

func NewProducer(
	log ports.LoggerWithTraceID,
	dsn,
	username,
	password string,
) ports.KafkaProducer {

	version, err := sarama.ParseKafkaVersion(sarama.DefaultVersion.String())
	if err != nil {
		panic(fmt.Sprintf("Error parsing Kafka version: %v", err))
	}

	config := sarama.NewConfig()
	config.Version = version
	config.Producer.Idempotent = true
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	config.Net.MaxOpenRequests = 1
	config.Net.SASL.Enable = true
	config.Net.SASL.User = username
	config.Net.SASL.Password = password
	config.Net.SASL.Handshake = true
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512

	producerProvider := newProducerProvider(strings.Split(dsn, ","), config)
	producer := producerProvider.producerProvider()

	ctx := context.Background()

	// Goroutine to listen to successes
	go func() {
		for success := range producer.Successes() {
			log.Debug(ctx, fmt.Sprintf("Message produced successfully: topic(%s)/partition(%d)/offset(%d)",
				success.Topic, success.Partition, success.Offset))
		}
	}()

	// TODO: add handler function
	// Goroutine to listen to errors
	go func() {
		for err := range producer.Errors() {
			log.Error(ctx, fmt.Sprintf("Failed to produce message: %v", err))
		}
	}()

	return &kafkaProducer{
		log:      log,
		dsn:      dsn,
		producer: producer,
	}
}

func (kp *kafkaProducer) Produce(topic string, value []byte) {
	kp.producer.Input() <- &sarama.ProducerMessage{Topic: topic, Key: nil, Value: sarama.ByteEncoder(value)}
}

func newProducerProvider(brokers []string, config *sarama.Config) *producerProvider {
	provider := &producerProvider{}
	provider.producerProvider = func() sarama.AsyncProducer {
		producer, err := sarama.NewAsyncProducer(brokers, config)
		if err != nil {
			panic(err)
		}
		return producer
	}
	return provider
}
