package eventbus

import "context"

type Event interface {
	Topic() string
	NewEvent() Event
}

type Handler func(ctx context.Context, event Event) error

type EventBus interface {
	Publish(ctx context.Context, event Event) error

	Subscribe(topic string, handler Handler) error

	Unsubscribe(topic string, handler Handler) error
}
