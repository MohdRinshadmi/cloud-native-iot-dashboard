package ingest_test

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/application/ingest"
	"github.com/ioss/iot-dashboard/backend/internal/application/realtime"
	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/domain/telemetry"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// ---- fakes -----------------------------------------------------------------

type fakeHistory struct {
	mu   sync.Mutex
	rows []telemetry.Reading
}

func (f *fakeHistory) InsertBatch(_ context.Context, rs []telemetry.Reading) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows = append(f.rows, rs...)
	return nil
}
func (f *fakeHistory) ListRecent(context.Context, string, string, time.Time, int) ([]telemetry.Reading, error) {
	return nil, nil
}
func (f *fakeHistory) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.rows)
}

type fakeLatest struct {
	mu sync.Mutex
	m  map[string]telemetry.Reading
}

func newFakeLatest() *fakeLatest { return &fakeLatest{m: map[string]telemetry.Reading{}} }
func (f *fakeLatest) SetLatest(_ context.Context, r telemetry.Reading) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.m[r.DeviceID] = r
	return nil
}
func (f *fakeLatest) GetLatest(_ context.Context, id string) (*telemetry.Reading, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.m[id]; ok {
		return &r, nil
	}
	return nil, apperror.NotFound("no reading")
}

type fakeRegistry struct {
	mu      sync.Mutex
	devices map[string]*device.Device
	updates int
}

func (f *fakeRegistry) FindByID(_ context.Context, id string) (*device.Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if d, ok := f.devices[id]; ok {
		cp := *d
		return &cp, nil
	}
	return nil, apperror.NotFound("device not found")
}
func (f *fakeRegistry) Update(_ context.Context, d *device.Device) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *d
	f.devices[d.ID] = &cp
	f.updates++
	return nil
}
func (f *fakeRegistry) MarkOfflineBefore(_ context.Context, cutoff time.Time) ([]device.StatusChange, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []device.StatusChange
	for _, d := range f.devices {
		if (d.Status == device.StatusOnline || d.Status == device.StatusDegraded) &&
			d.LastSeenAt != nil && d.LastSeenAt.Before(cutoff) {
			d.Status = device.StatusOffline
			out = append(out, device.StatusChange{
				DeviceID: d.ID, TenantID: d.TenantID, Status: d.Status, LastSeenAt: d.LastSeenAt,
			})
		}
	}
	return out, nil
}

type captureBroadcaster struct {
	mu     sync.Mutex
	events []realtime.Event
}

func (b *captureBroadcaster) Broadcast(e realtime.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, e)
}
func (b *captureBroadcaster) byType(t realtime.EventType) []realtime.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []realtime.Event
	for _, e := range b.events {
		if e.Type == t {
			out = append(out, e)
		}
	}
	return out
}

// ---- harness ---------------------------------------------------------------

func newPipeline(t *testing.T) (*ingest.Service, *fakeHistory, *fakeLatest, *fakeRegistry, *captureBroadcaster, context.CancelFunc) {
	t.Helper()
	history := &fakeHistory{}
	latest := newFakeLatest()
	now := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	seen := now.Add(-time.Hour)
	reg := &fakeRegistry{devices: map[string]*device.Device{
		"dev-1": {ID: "dev-1", TenantID: "t1", Name: "pump", Status: device.StatusOffline, LastSeenAt: &seen},
	}}
	bc := &captureBroadcaster{}
	svc := ingest.NewService(
		ingest.Config{Workers: 2, QueueSize: 16, HeartbeatInterval: 30 * time.Second, OfflineAfter: time.Hour},
		history, latest, reg, bc, testLogger(), func() time.Time { return now },
	)
	ctx, cancel := context.WithCancel(context.Background())
	svc.Start(ctx)
	return svc, history, latest, reg, bc, cancel
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met within deadline")
}

// ---- tests -----------------------------------------------------------------

func TestPipeline_TelemetryFlow(t *testing.T) {
	svc, history, latest, reg, bc, cancel := newPipeline(t)
	defer func() { cancel(); svc.Wait() }()

	ok := svc.Enqueue(ingest.Message{
		DeviceID:   "dev-1",
		Kind:       ingest.KindTelemetry,
		Payload:    []byte(`{"temperature": 42.5, "battery": 87}`),
		ReceivedAt: time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC),
	})
	if !ok {
		t.Fatal("enqueue rejected")
	}

	waitFor(t, func() bool { return history.count() == 1 })

	// History row persisted with tenant enrichment.
	if history.rows[0].TenantID != "t1" {
		t.Errorf("tenant = %q, want t1", history.rows[0].TenantID)
	}
	// Latest cache updated.
	r, err := latest.GetLatest(context.Background(), "dev-1")
	if err != nil || r.Temperature == nil || *r.Temperature != 42.5 {
		t.Error("latest store not updated correctly")
	}
	// Offline device transitioned online (heartbeat) + both events broadcast.
	waitFor(t, func() bool { return len(bc.byType(realtime.EventDeviceStatus)) == 1 })
	reg.mu.Lock()
	status := reg.devices["dev-1"].Status
	reg.mu.Unlock()
	if status != device.StatusOnline {
		t.Errorf("device status = %s, want online", status)
	}
	if len(bc.byType(realtime.EventTelemetry)) != 1 {
		t.Error("expected one telemetry broadcast")
	}
}

func TestPipeline_UnknownDeviceDropped(t *testing.T) {
	svc, history, _, _, bc, cancel := newPipeline(t)
	defer func() { cancel(); svc.Wait() }()

	svc.Enqueue(ingest.Message{
		DeviceID: "ghost", Kind: ingest.KindTelemetry,
		Payload: []byte(`{"temperature": 1}`), ReceivedAt: time.Now(),
	})

	time.Sleep(150 * time.Millisecond)
	if history.count() != 0 || len(bc.events) != 0 {
		t.Error("unknown device must be dropped entirely")
	}
}

func TestPipeline_QueueShedsWhenFull(t *testing.T) {
	// No workers started — queue fills up.
	history := &fakeHistory{}
	svc := ingest.NewService(
		ingest.Config{Workers: 1, QueueSize: 2},
		history, newFakeLatest(),
		&fakeRegistry{devices: map[string]*device.Device{}},
		realtime.NopBroadcaster{}, testLogger(), time.Now,
	)
	// Not calling Start: nothing drains the queue.
	m := ingest.Message{DeviceID: "d", Kind: ingest.KindTelemetry, Payload: []byte(`{}`), ReceivedAt: time.Now()}
	if !svc.Enqueue(m) || !svc.Enqueue(m) {
		t.Fatal("first two enqueues must fit")
	}
	if svc.Enqueue(m) {
		t.Error("third enqueue must be shed (queue full)")
	}
	_, dropped := svc.Stats()
	if dropped != 1 {
		t.Errorf("dropped = %d, want 1", dropped)
	}
}

func TestPipeline_StatusOfflineAnnouncement(t *testing.T) {
	svc, _, _, reg, bc, cancel := newPipeline(t)
	defer func() { cancel(); svc.Wait() }()

	// Bring it online first.
	svc.Enqueue(ingest.Message{
		DeviceID: "dev-1", Kind: ingest.KindStatus,
		Payload: []byte(`{"status":"online"}`), ReceivedAt: time.Now(),
	})
	waitFor(t, func() bool {
		reg.mu.Lock()
		defer reg.mu.Unlock()
		return reg.devices["dev-1"].Status == device.StatusOnline
	})

	// LWT announces offline.
	svc.Enqueue(ingest.Message{
		DeviceID: "dev-1", Kind: ingest.KindStatus,
		Payload: []byte(`{"status":"offline"}`), ReceivedAt: time.Now(),
	})
	waitFor(t, func() bool {
		reg.mu.Lock()
		defer reg.mu.Unlock()
		return reg.devices["dev-1"].Status == device.StatusOffline
	})

	if got := len(bc.byType(realtime.EventDeviceStatus)); got != 2 {
		t.Errorf("status broadcasts = %d, want 2 (online + offline)", got)
	}
}
