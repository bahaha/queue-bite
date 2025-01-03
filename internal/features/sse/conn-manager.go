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

// ServerSentEvents manages real-time event streaming to connected clients.
// Handles SSE connections and routes events to appropriate clients.
type ServerSentEvents interface {
	// RegisterClient establishes SSE connection with client browser.
	// Sets up required headers and begins streaming for specified party.
	RegisterClient(w http.ResponseWriter, partyID d.PartyID)

	// UnregisterClient removes client connection and cleans up resources.
	// Called when client disconnects or connection times out.
	UnregisterClient(partyID d.PartyID)

	// HandleNotifyPartyReady processes ready status events.
	// Streams notification to client when their party becomes ready.
	HandleNotifyPartyReady(ctx context.Context, event eventbus.Event) error

	// HandleNotifyPartyQueueStatusUpdate processes queue updates.
	// Streams queue position and wait time updates to connected clients.
	HandleNotifyPartyQueueStatusUpdate(ctx context.Context, event eventbus.Event) error
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

func NewServerSentEvent(logger log.Logger, eventbus eventbus.EventBus) ServerSentEvents {
	svc := &sse{
		logger:   logger,
		eventbus: eventbus,
		clients:  make(map[d.PartyID]*Client),
		mu:       sync.RWMutex{},
	}

	svc.subscribeToEvents()
	return svc
}

func (s *sse) subscribeToEvents() {
	s.eventbus.Subscribe(TopicNotifyPartyReady, s.HandleNotifyPartyReady)
	s.eventbus.Subscribe(TopicNotifyPartyQueueStatusUpdate, s.HandleNotifyPartyQueueStatusUpdate)
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
