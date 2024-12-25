package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/redis/go-redis/v9"

	log "queue-bite/internal/config/logger"
	"queue-bite/internal/platform/eventbus"
)

var REDIS_EVENTBUS = "eventbus/redis"

type redisEventBus struct {
	logger   log.Logger
	client   *redis.Client
	sub      *redis.PubSub
	mu       *sync.RWMutex
	handlers map[string][]eventbus.Handler
	registry *eventbus.EventRegistry
}

func NewRedisEventBus(logger log.Logger, client *redis.Client, registry *eventbus.EventRegistry) eventbus.EventBus {
	bus := &redisEventBus{
		logger:   logger,
		client:   client,
		sub:      client.Subscribe(context.Background()),
		mu:       &sync.RWMutex{},
		registry: registry,
		handlers: make(map[string][]eventbus.Handler),
	}
	go bus.startSubscriptionLoop()
	return bus
}

func (bus *redisEventBus) Publish(ctx context.Context, event eventbus.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		bus.logger.LogErr(REDIS_EVENTBUS, err, "failed to parse event payload", "event", event)
		return err
	}

	if err := bus.client.Publish(ctx, event.Topic(), data).Err(); err != nil {
		bus.logger.LogErr(REDIS_EVENTBUS, err, "failed to publish event to topic", "topic", event.Topic(), "event", event)
		return err
	}
	bus.logger.LogDebug(REDIS_EVENTBUS, "publish event to topic", "topic", event.Topic(), "event", event)
	return nil
}

func (bus *redisEventBus) Subscribe(topic string, handler eventbus.Handler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.handlers[topic] = append(bus.handlers[topic], handler)

	if len(bus.handlers[topic]) == 1 {
		if err := bus.sub.Subscribe(context.Background(), topic); err != nil {
			bus.logger.LogErr(REDIS_EVENTBUS, err, "subscribe to topic of event bus", "topic", topic)
			return err
		}
	}

	return nil
}

func (bus *redisEventBus) Unsubscribe(topic string, handler eventbus.Handler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	handlers := bus.handlers[topic]
	for i, h := range handlers {
		if &h == &handler {
			bus.handlers[topic] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	if len(bus.handlers[topic]) == 0 {
		if err := bus.sub.Unsubscribe(context.Background(), topic); err != nil {
			bus.logger.LogErr(REDIS_EVENTBUS, err, "unsubscribe to topic of event bus", "topic", topic)
			return err
		}
		delete(bus.handlers, topic)
	}

	return nil
}

func (bus *redisEventBus) startSubscriptionLoop() {
	ch := bus.sub.Channel()

	for msg := range ch {
		bus.mu.RLock()
		handlers, exists := bus.handlers[msg.Channel]
		bus.mu.RUnlock()

		if !exists {
			continue
		}

		eventType, ok := bus.registry.GetEventType(msg.Channel)
		if !ok {
			bus.logger.LogDebug(REDIS_EVENTBUS, "unknown event, check event registry configuration", "topic", msg.Channel)
			continue
		}

		event := eventType.NewEvent()
		if err := json.Unmarshal([]byte(msg.Payload), event); err != nil {
			bus.logger.LogErr(REDIS_EVENTBUS, err, "failed to parse the payload of event", "topic", msg.Channel)
			continue
		}

		for _, handler := range handlers {
			go func(h eventbus.Handler) {
				if err := h(context.Background(), event); err != nil {
					bus.logger.LogErr(REDIS_EVENTBUS, err, "failed to handle event")
				}
			}(handler)
		}
	}
}
