// Package realtime defines the event contract between the ingest pipeline and
// realtime consumers (the WebSocket hub today; SNS/EventBridge in the AWS
// topology). Events are tenant-tagged so transports can enforce isolation.
package realtime

import (
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/domain/telemetry"
)

// EventType discriminates the payload shape on the wire.
type EventType string

const (
	EventTelemetry    EventType = "telemetry"
	EventDeviceStatus EventType = "device_status"
)

// Event is the envelope broadcast to clients. TenantID routes the event and is
// never serialized to the client payload.
type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`

	TenantID string `json:"-"`
}

// StatusData is the device_status payload.
type StatusData struct {
	DeviceID   string     `json:"device_id"`
	Status     string     `json:"status"`
	LastSeenAt *time.Time `json:"last_seen_at"`
}

// NewTelemetryEvent wraps a reading for broadcast (tenant stripped from wire).
func NewTelemetryEvent(r telemetry.Reading) Event {
	return Event{Type: EventTelemetry, TenantID: r.TenantID, Data: r}
}

// NewStatusEvent wraps a device status transition for broadcast.
func NewStatusEvent(tenantID string, d StatusData) Event {
	return Event{Type: EventDeviceStatus, TenantID: tenantID, Data: d}
}

// Broadcaster is the fan-out port implemented by the WebSocket hub. It must
// be non-blocking: a slow consumer can never stall the ingest pipeline.
type Broadcaster interface {
	Broadcast(e Event)
}

// NopBroadcaster discards events (tests, tooling).
type NopBroadcaster struct{}

func (NopBroadcaster) Broadcast(Event) {}
