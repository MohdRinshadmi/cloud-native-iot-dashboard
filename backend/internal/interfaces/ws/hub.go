// Package ws is the WebSocket transport: a tenant-aware fan-out hub
// implementing the realtime.Broadcaster port. Design follows the classic
// hub/client-pump pattern:
//
//   - the hub owns the client set on a single goroutine (no lock contention)
//   - each client gets a buffered send channel; a SLOW CLIENT IS DISCONNECTED
//     rather than allowed to backpressure the pipeline
//   - events are marshalled once per broadcast, not once per client
package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync/atomic"

	"github.com/ioss/iot-dashboard/backend/internal/application/realtime"
)

// Hub fans events out to connected clients, scoped by tenant.
type Hub struct {
	register   chan *Client
	unregister chan *Client
	events     chan realtime.Event
	clients    map[*Client]struct{}
	log        *slog.Logger
	connected  atomic.Int64
}

// NewHub constructs the hub; call Run to start it.
func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		events:     make(chan realtime.Event, 512),
		clients:    make(map[*Client]struct{}),
		log:        log,
	}
}

var _ realtime.Broadcaster = (*Hub)(nil)

// Broadcast implements realtime.Broadcaster. Non-blocking: if the hub's event
// buffer is full the event is dropped — ingest must never stall on fan-out.
func (h *Hub) Broadcast(e realtime.Event) {
	select {
	case h.events <- e:
	default:
		h.log.Warn("ws hub event buffer full, dropping event", slog.String("type", string(e.Type)))
	}
}

// Connected returns the live client count (exported to metrics in Phase 10).
func (h *Hub) Connected() int64 { return h.connected.Load() }

// Run is the hub's single-owner loop; exits when ctx is cancelled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for c := range h.clients {
				c.close()
			}
			return

		case c := <-h.register:
			h.clients[c] = struct{}{}
			h.connected.Store(int64(len(h.clients)))
			h.log.Info("ws client connected",
				slog.String("tenant_id", c.tenantID),
				slog.String("user_id", c.userID),
				slog.Int("total", len(h.clients)),
			)

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				c.close()
			}
			h.connected.Store(int64(len(h.clients)))

		case e := <-h.events:
			payload, err := json.Marshal(e)
			if err != nil {
				continue
			}
			for c := range h.clients {
				if c.tenantID != e.TenantID {
					continue // hard tenant isolation at the transport edge
				}
				select {
				case c.send <- payload:
				default:
					// Client can't keep up — cut it loose; it will reconnect.
					delete(h.clients, c)
					c.close()
					h.connected.Store(int64(len(h.clients)))
					h.log.Warn("ws client dropped (slow consumer)", slog.String("user_id", c.userID))
				}
			}
		}
	}
}
