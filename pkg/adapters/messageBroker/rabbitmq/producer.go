package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/armineyvazi/common.git/pkg/ports"
)

type producer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewProducer dials RabbitMQ and opens a channel ready for publishing.
// The caller must call Close when done.
func NewProducer(dsn string) (ports.MessageProducer, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq open channel: %w", err)
	}
	return &producer{conn: conn, ch: ch}, nil
}

// Publish JSON-encodes payload and publishes it to exchangeName. The first
// element of routeKeys is used as the routing key; queueName is used as a
// fallback when routeKeys is empty. duplicatedMessageKey is set as the AMQP
// MessageId for idempotency checks at the consumer.
func (p *producer) Publish(ctx context.Context, payload interface{}, duplicatedMessageKey, exchangeName, queueName string, routeKeys []string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	routingKey := queueName
	if len(routeKeys) > 0 {
		routingKey = routeKeys[0]
	}

	err = p.ch.PublishWithContext(ctx, exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    duplicatedMessageKey,
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("publish to exchange %q: %w", exchangeName, err)
	}
	return nil
}

func (p *producer) Close() error {
	if err := p.ch.Close(); err != nil {
		return fmt.Errorf("close rabbitmq channel: %w", err)
	}
	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("close rabbitmq connection: %w", err)
	}
	return nil
}
