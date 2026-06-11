package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ioss/iot-dashboard/backend/internal/domain/telemetry"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// latestKeyPrefix namespaces hot telemetry; TTL keeps the cache self-pruning
// for decommissioned devices.
const (
	latestKeyPrefix = "telemetry:latest:"
	latestTTL       = 24 * time.Hour
)

// LatestStore is the Redis adapter for telemetry.LatestStore — O(1) access to
// each device's newest reading, powering live panels without touching Postgres.
type LatestStore struct{ client *redis.Client }

func NewLatestStore(client *redis.Client) *LatestStore { return &LatestStore{client: client} }

var _ telemetry.LatestStore = (*LatestStore)(nil)

// storedReading persists the tenant alongside the wire fields (the Reading's
// json tags hide TenantID, which the server needs back on reads).
type storedReading struct {
	telemetry.Reading
	TenantID string `json:"tenant_id"`
}

func (s *LatestStore) SetLatest(ctx context.Context, r telemetry.Reading) error {
	b, err := json.Marshal(storedReading{Reading: r, TenantID: r.TenantID})
	if err != nil {
		return err
	}
	return s.client.Set(ctx, latestKeyPrefix+r.DeviceID, b, latestTTL).Err()
}

func (s *LatestStore) GetLatest(ctx context.Context, deviceID string) (*telemetry.Reading, error) {
	b, err := s.client.Get(ctx, latestKeyPrefix+deviceID).Bytes()
	if err == redis.Nil {
		return nil, apperror.NotFound("no telemetry for device")
	}
	if err != nil {
		return nil, apperror.Internal(err)
	}
	var sr storedReading
	if err := json.Unmarshal(b, &sr); err != nil {
		return nil, apperror.Internal(err)
	}
	r := sr.Reading
	r.TenantID = sr.TenantID
	return &r, nil
}
