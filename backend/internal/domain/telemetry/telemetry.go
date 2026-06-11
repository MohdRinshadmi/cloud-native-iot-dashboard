// Package telemetry is the time-series bounded context: device readings and
// the ports for storing/querying them. Hot path (latest value) and history
// (rows) are separate ports because their backing stores differ (Redis vs
// Postgres) — and later DynamoDB/S3 in the AWS topology.
package telemetry

import (
	"context"
	"time"
)

// Reading is one telemetry sample. Metric fields are pointers: devices report
// different subsets and "absent" must be distinguishable from zero.
type Reading struct {
	TenantID    string    `json:"-"`
	DeviceID    string    `json:"device_id"`
	TS          time.Time `json:"ts"`
	Temperature *float64  `json:"temperature,omitempty"`
	Battery     *float64  `json:"battery,omitempty"`
	Voltage     *float64  `json:"voltage,omitempty"`
	CPU         *float64  `json:"cpu,omitempty"`
	Memory      *float64  `json:"memory,omitempty"`
	Signal      *float64  `json:"signal,omitempty"`
	Lat         *float64  `json:"lat,omitempty"`
	Lng         *float64  `json:"lng,omitempty"`
}

// Normalize clamps obviously-corrupt values and defaults a missing timestamp.
// Devices in the field send garbage; the pipeline must not amplify it.
func (r *Reading) Normalize(now time.Time) {
	if r.TS.IsZero() || r.TS.After(now.Add(5*time.Minute)) {
		r.TS = now
	}
	clampPct(&r.Battery)
	clampPct(&r.CPU)
	clampPct(&r.Memory)
}

func clampPct(p **float64) {
	if *p == nil {
		return
	}
	v := **p
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	*p = &v
}

// Repository is the history-storage port (Postgres now; cold storage later).
// InsertBatch takes a slice so a batching writer can slot in without an API
// change — workers currently flush per-message.
type Repository interface {
	InsertBatch(ctx context.Context, readings []Reading) error
	// ListRecent returns readings for one device since `since`, newest first,
	// tenant-scoped at the query level.
	ListRecent(ctx context.Context, tenantID, deviceID string, since time.Time, limit int) ([]Reading, error)
}

// LatestStore is the hot-path port: O(1) read of a device's newest reading.
type LatestStore interface {
	SetLatest(ctx context.Context, r Reading) error
	GetLatest(ctx context.Context, deviceID string) (*Reading, error)
}
