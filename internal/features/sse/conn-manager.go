package sse

import (
	"context"
	"net/http"
	log "queue-bite/internal/config/logger"
	d "queue-bite/internal/domain"
	"queue-bite/internal/platform/eventbus"
	"sync"
)

var SSE = "sse"

type ServerSentEvents interface {
	RegisterClient(w http.ResponseWriter, partyID d.PartyID)
	UnregisterClient(partyID d.PartyID)

	HandleNotifyPartyReady(ctx context.Context, event eventbus.Event) error
}

type sse struct {
	logger   log.Logger
	eventbus eventbus.EventBus
	clients  map[d.PartyID]*Client
	mu       sync.RWMutex
}

type Client struct {
	PartyID d.PartyID
	Writer  http.ResponseWriter
	Done    chan struct{}
}

func NewServerSentEvent(logger log.Logger, eventbus eventbus.EventBus) *sse {
	svc := &sse{
		logger:   logger,
		eventbus: eventbus,
		clients:  make(map[d.PartyID]*Client),
		mu:       sync.RWMutex{},
	}

	return svc
}

func (s *sse) subscribeToEvents() {
	s.eventbus.Subscribe(TopicNotifyPartyReady, s.HandleNotifyPartyReady)
}

func (s *sse) RegisterClient(w http.ResponseWriter, partyID d.PartyID) {
	client := &Client{
		PartyID: partyID,
		Writer:  w,
		Done:    make(chan struct{}),
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[partyID] = client
}

func (s *sse) UnregisterClient(partyID d.PartyID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, partyID)
}

func (s *sse) getClient(partyID d.PartyID) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	client, exists := s.clients[partyID]
	if !exists {
		return nil
	} else {
		return client
	}
}
