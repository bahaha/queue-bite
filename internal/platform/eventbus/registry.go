package eventbus

import "sync"

type EventRegistry struct {
	eventTypes map[string]Event
	mu         sync.RWMutex
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		eventTypes: map[string]Event{},
	}
}

func (r *EventRegistry) Register(topic string, event Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.eventTypes[topic] = event
}

func (r *EventRegistry) GetEventType(topic string) (Event, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	et, ok := r.eventTypes[topic]
	return et, ok
}
