package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// DeviceRepository is the GORM adapter for device.Repository.
// Every query filters by tenant_id — isolation is structural.
type DeviceRepository struct{ db *gorm.DB }

// NewDeviceRepository wires the adapter.
func NewDeviceRepository(db *gorm.DB) *DeviceRepository { return &DeviceRepository{db: db} }

var _ device.Repository = (*DeviceRepository)(nil)

func (r *DeviceRepository) Create(ctx context.Context, d *device.Device) error {
	err := dbFrom(ctx, r.db).Create(deviceToModel(d)).Error
	return translateError(err, "device not found")
}

func (r *DeviceRepository) GetByID(ctx context.Context, tenantID, id string) (*device.Device, error) {
	var m deviceModel
	err := dbFrom(ctx, r.db).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		First(&m).Error
	if err != nil {
		return nil, translateError(err, "device not found")
	}
	return m.toDomain(), nil
}

func (r *DeviceRepository) Update(ctx context.Context, d *device.Device) error {
	res := dbFrom(ctx, r.db).
		Where("tenant_id = ? AND id = ?", d.TenantID, d.ID).
		Updates(deviceToModel(d))
	if res.Error != nil {
		return translateError(res.Error, "device not found")
	}
	if res.RowsAffected == 0 {
		return apperror.NotFound("device not found")
	}
	return nil
}

func (r *DeviceRepository) Delete(ctx context.Context, tenantID, id string) error {
	res := dbFrom(ctx, r.db).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&deviceModel{})
	if res.Error != nil {
		return translateError(res.Error, "device not found")
	}
	if res.RowsAffected == 0 {
		return apperror.NotFound("device not found")
	}
	return nil
}

// FindByID is the tenant-UNSCOPED lookup used exclusively by the ingest
// pipeline, where identity comes from the broker connection rather than a
// user session. Never expose this through a user-facing handler.
func (r *DeviceRepository) FindByID(ctx context.Context, id string) (*device.Device, error) {
	var m deviceModel
	err := dbFrom(ctx, r.db).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, translateError(err, "device not found")
	}
	return m.toDomain(), nil
}

// MarkOfflineBefore flips online/degraded devices whose heartbeat is older
// than cutoff to offline, returning each transition for broadcast. One UPDATE
// with RETURNING — the sweep stays O(1) round-trips regardless of fleet size.
func (r *DeviceRepository) MarkOfflineBefore(ctx context.Context, cutoff time.Time) ([]device.StatusChange, error) {
	type row struct {
		ID         string
		TenantID   string
		LastSeenAt *time.Time
	}
	var rows []row
	err := dbFrom(ctx, r.db).Raw(`
		UPDATE devices
		   SET status = 'offline', updated_at = now()
		 WHERE status IN ('online', 'degraded')
		   AND (last_seen_at IS NULL OR last_seen_at < ?)
		RETURNING id, tenant_id, last_seen_at`, cutoff).
		Scan(&rows).Error
	if err != nil {
		return nil, translateError(err, "device not found")
	}
	out := make([]device.StatusChange, len(rows))
	for i, rw := range rows {
		out[i] = device.StatusChange{
			DeviceID: rw.ID, TenantID: rw.TenantID,
			Status: device.StatusOffline, LastSeenAt: rw.LastSeenAt,
		}
	}
	return out, nil
}

func (r *DeviceRepository) List(ctx context.Context, tenantID string, f device.Filter) ([]*device.Device, int64, error) {
	q := dbFrom(ctx, r.db).Model(&deviceModel{}).Where("tenant_id = ?", tenantID)

	if f.Q != "" {
		like := "%" + f.Q + "%"
		q = q.Where("(name ILIKE ? OR model ILIKE ?)", like, like)
	}
	if f.Status != "" {
		q = q.Where("status = ?", string(f.Status))
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, translateError(err, "device not found")
	}

	page := f.Page.Normalize()
	var models []deviceModel
	err := q.Order("created_at DESC, id").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, translateError(err, "device not found")
	}

	out := make([]*device.Device, len(models))
	for i := range models {
		out[i] = models[i].toDomain()
	}
	return out, total, nil
}
