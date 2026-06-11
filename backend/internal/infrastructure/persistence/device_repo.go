package persistence

import (
	"context"

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
