// Package rabbitmq provides RabbitMQ message broker adapters using AMQP 0-9-1.
package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/armineyvazi/common.git/pkg/ports"
)

type consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewConsumer dials RabbitMQ and opens a channel. The caller must call Close
// when done to release resources.
func NewConsumer(dsn string) (ports.MessageConsumer, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq open channel: %w", err)
	}
	return &consumer{conn: conn, ch: ch}, nil
}

// Consume declares the exchange and queue, binds the routing keys, and starts
// delivering messages to resultChan. The goroutine stops when ctx is cancelled
// or the AMQP channel is closed.
func (c *consumer) Consume(ctx context.Context, queue, exchangeName string, routeKeys []string, resultChan chan ports.Delivery) error {
	if err := c.ch.ExchangeDeclare(exchangeName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %q: %w", exchangeName, err)
	}

	q, err := c.ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue %q: %w", queue, err)
	}

	for _, key := range routeKeys {
		if err := c.ch.QueueBind(q.Name, key, exchangeName, false, nil); err != nil {
			return fmt.Errorf("bind queue %q with key %q: %w", q.Name, key, err)
		}
	}

	msgs, err := c.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("start consume on %q: %w", q.Name, err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				resultChan <- ports.Delivery{
					Body: msg.Body,
					Tag:  msg.DeliveryTag,
				}
			}
		}
	}()

	return nil
}

func (c *consumer) Ack(tag uint64) error {
	return c.ch.Ack(tag, false)
}

func (c *consumer) Requeue(tag uint64) error {
	return c.ch.Nack(tag, false, true)
}

func (c *consumer) Reject(tag uint64) error {
	return c.ch.Reject(tag, false)
}

func (c *consumer) Close() error {
	if err := c.ch.Close(); err != nil {
		return fmt.Errorf("close rabbitmq channel: %w", err)
	}
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("close rabbitmq connection: %w", err)
	}
	return nil
}
