package ports

import "context"

type Delivery struct {
	Body []byte
	Tag  uint64
}

type MessageConsumer interface {
	Consume(ctx context.Context, queue, exchangeName string, routeKeys []string, resultChan chan Delivery) error
	Ack(tag uint64) error
	Requeue(tag uint64) error
	Reject(tag uint64) error
	Close() error
}

type MessageProducer interface {
	Publish(ctx context.Context, payload interface{}, duplicatedMessageKey string, exchangeName, queueName string, routeKeys []string) error
	Close() error
}
