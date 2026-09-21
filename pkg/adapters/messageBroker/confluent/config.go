// Package confluent provides ports.KafkaProducer and ports.KafkaConsumer
// implementations backed by github.com/confluentinc/confluent-kafka-go/v2,
// which wraps librdkafka. It is a drop-in replacement for the Sarama-based
// adapter and supports the full range of librdkafka client configuration.
package confluent

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

// SecurityProtocol controls transport-layer and authentication security.
type SecurityProtocol string

const (
	SecurityProtocolPlaintext     SecurityProtocol = "PLAINTEXT"
	SecurityProtocolSASLPlaintext SecurityProtocol = "SASL_PLAINTEXT"
	SecurityProtocolSSL           SecurityProtocol = "SSL"
	SecurityProtocolSASLSSL       SecurityProtocol = "SASL_SSL"
)

// SASLMechanism selects the SASL authentication mechanism.
type SASLMechanism string

const (
	SASLMechanismPlain       SASLMechanism = "PLAIN"
	SASLMechanismSCRAMSHA256 SASLMechanism = "SCRAM-SHA-256"
	SASLMechanismSCRAMSHA512 SASLMechanism = "SCRAM-SHA-512"
	SASLMechanismGSSAPI      SASLMechanism = "GSSAPI"
	SASLMechanismOAuthBearer SASLMechanism = "OAUTHBEARER"
)

// CompressionType selects the producer message compression algorithm.
type CompressionType string

const (
	CompressionNone   CompressionType = "none"
	CompressionGZIP   CompressionType = "gzip"
	CompressionSnappy CompressionType = "snappy"
	CompressionLZ4    CompressionType = "lz4"
	CompressionZSTD   CompressionType = "zstd"
)

// AutoOffsetReset controls where the consumer starts when no committed offset exists.
type AutoOffsetReset string

const (
	AutoOffsetResetEarliest AutoOffsetReset = "earliest"
	AutoOffsetResetLatest   AutoOffsetReset = "latest"
	AutoOffsetResetNone     AutoOffsetReset = "none"
)

// BaseConfig holds fields shared by both the producer and the consumer.
type BaseConfig struct {
	// Brokers is a comma-separated list of host:port pairs.
	// Example: "broker1:9092,broker2:9092"
	Brokers string

	// SecurityProtocol defines the protocol used to communicate with brokers.
	// Defaults to PLAINTEXT when empty.
	SecurityProtocol SecurityProtocol

	// SASL authentication fields. Ignored when SecurityProtocol is PLAINTEXT or SSL.
	SASLMechanism SASLMechanism
	SASLUsername  string
	SASLPassword  string

	// SSL/TLS fields. Only used when SecurityProtocol is SSL or SASL_SSL.
	SSLCALocation   string // path to CA certificate file
	SSLCertLocation string // path to client certificate file
	SSLKeyLocation  string // path to client private key file
	SSLKeyPassword  string
	// DisableSSLCertVerification skips broker certificate verification.
	// Set true only in development or when using self-signed certificates.
	// The default (false) leaves librdkafka's secure default (verification on) intact.
	DisableSSLCertVerification bool

	// SocketTimeoutMs is the TCP socket timeout in milliseconds. Default 60000.
	SocketTimeoutMs int
	// MetadataMaxAgeMs is how often the client refreshes cluster metadata. Default 300000.
	MetadataMaxAgeMs int
}

// ProducerConfig holds all producer-specific settings on top of BaseConfig.
type ProducerConfig struct {
	BaseConfig

	// Acks controls the number of broker acknowledgements required before
	// considering a produce request complete. "all" is the safest default.
	// Valid values: "0", "1", "all" (or "-1").
	Acks string

	// EnableIdempotence guarantees exactly-once delivery within a session.
	// Requires Acks="all". Defaults to true.
	EnableIdempotence bool

	// CompressionType selects the compression codec. Defaults to none.
	CompressionType CompressionType

	// Retries is the number of retries before a message is considered failed. Default 5.
	Retries int
	// RetryBackoffMs is the backoff between retries in milliseconds. Default 100.
	RetryBackoffMs int
	// MessageTimeoutMs is the total time a message may block before being dropped. Default 300000.
	MessageTimeoutMs int

	// BatchNumMessages is the maximum number of messages batched per produce request. Default 10000.
	BatchNumMessages int
	// LingerMs delays sending to allow batching in milliseconds. Default 5.
	LingerMs int
	// QueueBufferingMaxMessages is the local queue capacity before blocking. Default 100000.
	QueueBufferingMaxMessages int
	// MaxMessageBytes is the maximum allowed request size. Default 1000000.
	MaxMessageBytes int
}

// ConsumerConfig holds all consumer-specific settings on top of BaseConfig.
type ConsumerConfig struct {
	BaseConfig

	// GroupID is the consumer group this consumer belongs to.
	GroupID string

	// AutoOffsetReset controls the starting offset when no committed offset exists.
	// Defaults to earliest.
	AutoOffsetReset AutoOffsetReset

	// EnableAutoCommit commits offsets automatically in the background. Default true.
	EnableAutoCommit bool
	// AutoCommitIntervalMs is the interval between automatic offset commits. Default 5000.
	AutoCommitIntervalMs int

	// SessionTimeoutMs is the consumer group session timeout. Default 45000.
	SessionTimeoutMs int
	// HeartbeatIntervalMs is how often heartbeats are sent to the broker. Default 3000.
	HeartbeatIntervalMs int
	// MaxPollIntervalMs is the maximum time between polls before the consumer
	// is considered dead and kicked from the group. Default 300000.
	MaxPollIntervalMs int

	// FetchMinBytes is the minimum amount of data the broker should return. Default 1.
	FetchMinBytes int
	// FetchMaxBytes is the maximum data per fetch request. Default 52428800 (50 MB).
	FetchMaxBytes int
}

// producerDefaults fills in zero values with sensible production defaults.
func producerDefaults(cfg ProducerConfig) ProducerConfig {
	if cfg.Acks == "" {
		cfg.Acks = "all"
	}
	if !cfg.EnableIdempotence {
		cfg.EnableIdempotence = true
	}
	if cfg.Retries == 0 {
		cfg.Retries = 5
	}
	if cfg.RetryBackoffMs == 0 {
		cfg.RetryBackoffMs = 100
	}
	if cfg.MessageTimeoutMs == 0 {
		cfg.MessageTimeoutMs = 300000
	}
	if cfg.BatchNumMessages == 0 {
		cfg.BatchNumMessages = 10000
	}
	if cfg.LingerMs <= 0 {
		cfg.LingerMs = 5
	}
	if cfg.QueueBufferingMaxMessages == 0 {
		cfg.QueueBufferingMaxMessages = 100000
	}
	if cfg.MaxMessageBytes == 0 {
		cfg.MaxMessageBytes = 1000000
	}
	return cfg
}

// consumerDefaults fills in zero values with sensible production defaults.
func consumerDefaults(cfg ConsumerConfig) ConsumerConfig {
	if cfg.AutoOffsetReset == "" {
		cfg.AutoOffsetReset = AutoOffsetResetEarliest
	}
	if !cfg.EnableAutoCommit {
		cfg.EnableAutoCommit = true
	}
	if cfg.AutoCommitIntervalMs == 0 {
		cfg.AutoCommitIntervalMs = 5000
	}
	if cfg.SessionTimeoutMs == 0 {
		cfg.SessionTimeoutMs = 45000
	}
	if cfg.HeartbeatIntervalMs == 0 {
		cfg.HeartbeatIntervalMs = 3000
	}
	if cfg.MaxPollIntervalMs == 0 {
		cfg.MaxPollIntervalMs = 300000
	}
	if cfg.FetchMinBytes == 0 {
		cfg.FetchMinBytes = 1
	}
	if cfg.FetchMaxBytes == 0 {
		cfg.FetchMaxBytes = 52428800
	}
	return cfg
}

// buildBaseConfigMap returns a kafka.ConfigMap with all BaseConfig fields applied.
func buildBaseConfigMap(b BaseConfig) kafka.ConfigMap {
	cm := kafka.ConfigMap{
		"bootstrap.servers": b.Brokers,
	}

	if b.SecurityProtocol != "" {
		cm["security.protocol"] = string(b.SecurityProtocol)
	}
	if b.SASLMechanism != "" {
		cm["sasl.mechanisms"] = string(b.SASLMechanism)
	}
	if b.SASLUsername != "" {
		cm["sasl.username"] = b.SASLUsername
	}
	if b.SASLPassword != "" {
		cm["sasl.password"] = b.SASLPassword
	}
	if b.SSLCALocation != "" {
		cm["ssl.ca.location"] = b.SSLCALocation
	}
	if b.SSLCertLocation != "" {
		cm["ssl.certificate.location"] = b.SSLCertLocation
	}
	if b.SSLKeyLocation != "" {
		cm["ssl.key.location"] = b.SSLKeyLocation
	}
	if b.SSLKeyPassword != "" {
		cm["ssl.key.password"] = b.SSLKeyPassword
	}
	if b.DisableSSLCertVerification {
		cm["enable.ssl.certificate.verification"] = false
	}
	if b.SocketTimeoutMs > 0 {
		cm["socket.timeout.ms"] = b.SocketTimeoutMs
	}
	if b.MetadataMaxAgeMs > 0 {
		cm["metadata.max.age.ms"] = b.MetadataMaxAgeMs
	}

	return cm
}
