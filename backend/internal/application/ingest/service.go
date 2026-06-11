// Package ingest is the telemetry event-processing pipeline:
//
//	MQTT consumer → Enqueue() → bounded queue → worker pool → {Postgres,
//	Redis latest, heartbeat state machine, WebSocket broadcast}
//
// Design properties:
//   - BACKPRESSURE: the queue is bounded; when full, messages are dropped and
//     counted rather than blocking the broker connection (load-shedding).
//   - THROTTLED WRITES: heartbeats update the devices row at most once per
//     interval per device — 1M devices at 1msg/s must not mean 1M row
//     updates/s. Status *transitions* always write and broadcast immediately.
//   - The package depends only on ports; MQTT/Gin/GORM never appear here.
package ingest

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/application/realtime"
	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/domain/telemetry"
)

// Kind discriminates inbound message types (mapped from MQTT topic suffix).
type Kind string

const (
	KindTelemetry Kind = "telemetry"
	KindStatus    Kind = "status"
)

// Message is one raw inbound MQTT message, minimally decoded at the edge.
type Message struct {
	DeviceID   string
	Kind       Kind
	Payload    []byte
	ReceivedAt time.Time
}

// DeviceRegistry is the narrow device port the pipeline needs (interface
// segregation — implemented by persistence.DeviceRepository).
type DeviceRegistry interface {
	// FindByID is tenant-unscoped: device identity is established by the
	// broker connection, not a user session. Ingest-path only.
	FindByID(ctx context.Context, id string) (*device.Device, error)
	Update(ctx context.Context, d *device.Device) error
	// MarkOfflineBefore flips online/degraded devices with stale heartbeats
	// to offline, returning the transitions for broadcast.
	MarkOfflineBefore(ctx context.Context, cutoff time.Time) ([]device.StatusChange, error)
}

// Config tunes the pipeline.
type Config struct {
	Workers           int           // concurrent processors
	QueueSize         int           // bounded inbox
	HeartbeatInterval time.Duration // min gap between last_seen row writes
	OfflineAfter      time.Duration // heartbeat silence ⇒ offline
}

// Service is the pipeline. Construct with NewService, run with Start.
type Service struct {
	cfg         Config
	queue       chan Message
	history     telemetry.Repository
	latest      telemetry.LatestStore
	devices     DeviceRegistry
	broadcaster realtime.Broadcaster
	log         *slog.Logger
	now         func() time.Time

	lastSeenWrite sync.Map // deviceID → time.Time of last persisted heartbeat
	dropped       atomic.Int64
	processed     atomic.Int64
	wg            sync.WaitGroup
}

// NewService wires the pipeline.
func NewService(
	cfg Config,
	history telemetry.Repository,
	latest telemetry.LatestStore,
	devices DeviceRegistry,
	broadcaster realtime.Broadcaster,
	log *slog.Logger,
	now func() time.Time,
) *Service {
	if cfg.Workers <= 0 {
		cfg.Workers = 4
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1024
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = 30 * time.Second
	}
	if cfg.OfflineAfter <= 0 {
		cfg.OfflineAfter = 90 * time.Second
	}
	return &Service{
		cfg:         cfg,
		queue:       make(chan Message, cfg.QueueSize),
		history:     history,
		latest:      latest,
		devices:     devices,
		broadcaster: broadcaster,
		log:         log,
		now:         now,
	}
}

// Enqueue hands a message to the pipeline. Non-blocking: returns false when
// the queue is full (message shed) so the MQTT callback never stalls.
func (s *Service) Enqueue(m Message) bool {
	select {
	case s.queue <- m:
		return true
	default:
		if n := s.dropped.Add(1); n%100 == 1 {
			s.log.Warn("ingest queue full, shedding load", slog.Int64("dropped_total", n))
		}
		return false
	}
}

// Start launches the worker pool and the offline sweeper; both exit when ctx
// is cancelled. Returns immediately.
func (s *Service) Start(ctx context.Context) {
	for i := 0; i < s.cfg.Workers; i++ {
		s.wg.Add(1)
		go s.worker(ctx)
	}
	s.wg.Add(1)
	go s.offlineSweeper(ctx)
	s.log.Info("ingest pipeline started",
		slog.Int("workers", s.cfg.Workers),
		slog.Int("queue_size", s.cfg.QueueSize),
		slog.Duration("offline_after", s.cfg.OfflineAfter),
	)
}

// Wait blocks until all workers have exited (call after cancelling ctx).
func (s *Service) Wait() { s.wg.Wait() }

// Stats exposes counters (Prometheus exports these in Phase 10).
func (s *Service) Stats() (processed, dropped int64) {
	return s.processed.Load(), s.dropped.Load()
}

func (s *Service) worker(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-s.queue:
			s.dispatch(ctx, m)
		}
	}
}

func (s *Service) dispatch(ctx context.Context, m Message) {
	// Per-message budget: one stuck dependency must not wedge a worker.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var err error
	switch m.Kind {
	case KindTelemetry:
		err = s.processTelemetry(ctx, m)
	case KindStatus:
		err = s.processStatus(ctx, m)
	}
	if err != nil {
		s.log.Warn("ingest message failed",
			slog.String("device_id", m.DeviceID),
			slog.String("kind", string(m.Kind)),
			slog.String("error", err.Error()),
		)
		return
	}
	s.processed.Add(1)
}

// processTelemetry: validate → enrich with tenant → persist history + latest
// → heartbeat → broadcast.
func (s *Service) processTelemetry(ctx context.Context, m Message) error {
	var r telemetry.Reading
	if err := json.Unmarshal(m.Payload, &r); err != nil {
		return err
	}
	// The topic, not the payload, is authoritative for identity.
	r.DeviceID = m.DeviceID
	r.Normalize(m.ReceivedAt)

	d, err := s.devices.FindByID(ctx, m.DeviceID)
	if err != nil {
		return err // unknown device — drop (rogue publisher or deleted device)
	}
	r.TenantID = d.TenantID

	if err := s.history.InsertBatch(ctx, []telemetry.Reading{r}); err != nil {
		return err
	}
	if err := s.latest.SetLatest(ctx, r); err != nil {
		// Hot-path cache failure is non-fatal; history already persisted.
		s.log.Warn("latest store write failed", slog.String("error", err.Error()))
	}

	s.heartbeat(ctx, d, r.TS)
	s.broadcaster.Broadcast(realtime.NewTelemetryEvent(r))
	return nil
}

// heartbeat records liveness. Transitions (→online) write + broadcast
// immediately; steady-state heartbeats persist at most once per interval.
func (s *Service) heartbeat(ctx context.Context, d *device.Device, at time.Time) {
	wasStatus := d.Status

	if last, ok := s.lastSeenWrite.Load(d.ID); ok {
		if t, ok := last.(time.Time); ok && at.Sub(t) < s.cfg.HeartbeatInterval && wasStatus == device.StatusOnline {
			return // throttled: recent write and no transition pending
		}
	}

	d.MarkSeen(at)
	if err := s.devices.Update(ctx, d); err != nil {
		s.log.Warn("heartbeat persist failed", slog.String("device_id", d.ID), slog.String("error", err.Error()))
		return
	}
	s.lastSeenWrite.Store(d.ID, at)

	if d.Status != wasStatus {
		s.broadcaster.Broadcast(realtime.NewStatusEvent(d.TenantID, realtime.StatusData{
			DeviceID:   d.ID,
			Status:     string(d.Status),
			LastSeenAt: d.LastSeenAt,
		}))
	}
}

// statusPayload is the devices/{id}/status message shape (incl. LWT).
type statusPayload struct {
	Status string `json:"status"`
}

// processStatus handles explicit online/offline announcements.
func (s *Service) processStatus(ctx context.Context, m Message) error {
	var p statusPayload
	if err := json.Unmarshal(m.Payload, &p); err != nil {
		return err
	}

	d, err := s.devices.FindByID(ctx, m.DeviceID)
	if err != nil {
		return err
	}

	wasStatus := d.Status
	switch p.Status {
	case "online":
		d.MarkSeen(m.ReceivedAt)
	case "offline":
		d.Status = device.StatusOffline
		d.UpdatedAt = m.ReceivedAt
	default:
		return nil // unknown announcement — ignore
	}

	if d.Status == wasStatus {
		return nil
	}
	if err := s.devices.Update(ctx, d); err != nil {
		return err
	}
	s.broadcaster.Broadcast(realtime.NewStatusEvent(d.TenantID, realtime.StatusData{
		DeviceID:   d.ID,
		Status:     string(d.Status),
		LastSeenAt: d.LastSeenAt,
	}))
	return nil
}

// offlineSweeper periodically flips silent devices offline. The sweep runs at
// a third of the threshold so detection lag stays bounded.
func (s *Service) offlineSweeper(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.cfg.OfflineAfter / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cutoff := s.now().Add(-s.cfg.OfflineAfter)
			changes, err := s.devices.MarkOfflineBefore(ctx, cutoff)
			if err != nil {
				s.log.Warn("offline sweep failed", slog.String("error", err.Error()))
				continue
			}
			for _, ch := range changes {
				s.broadcaster.Broadcast(realtime.NewStatusEvent(ch.TenantID, realtime.StatusData{
					DeviceID:   ch.DeviceID,
					Status:     string(ch.Status),
					LastSeenAt: ch.LastSeenAt,
				}))
			}
			if len(changes) > 0 {
				s.log.Info("offline sweep", slog.Int("transitioned", len(changes)))
			}
		}
	}
}
