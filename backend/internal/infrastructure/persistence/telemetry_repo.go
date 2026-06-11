package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ioss/iot-dashboard/backend/internal/domain/telemetry"
)

type telemetryModel struct {
	ID          int64 `gorm:"primaryKey;autoIncrement"`
	TenantID    string
	DeviceID    string
	TS          time.Time
	Temperature *float64
	Battery     *float64
	Voltage     *float64
	CPU         *float64 `gorm:"column:cpu"`
	Memory      *float64
	Signal      *float64
	Lat         *float64
	Lng         *float64
	CreatedAt   time.Time
}

func (telemetryModel) TableName() string { return "telemetry" }

func readingToModel(r telemetry.Reading) telemetryModel {
	return telemetryModel{
		TenantID: r.TenantID, DeviceID: r.DeviceID, TS: r.TS,
		Temperature: r.Temperature, Battery: r.Battery, Voltage: r.Voltage,
		CPU: r.CPU, Memory: r.Memory, Signal: r.Signal, Lat: r.Lat, Lng: r.Lng,
	}
}

func (m *telemetryModel) toDomain() telemetry.Reading {
	return telemetry.Reading{
		TenantID: m.TenantID, DeviceID: m.DeviceID, TS: m.TS,
		Temperature: m.Temperature, Battery: m.Battery, Voltage: m.Voltage,
		CPU: m.CPU, Memory: m.Memory, Signal: m.Signal, Lat: m.Lat, Lng: m.Lng,
	}
}

// TelemetryRepository is the GORM adapter for telemetry.Repository.
type TelemetryRepository struct{ db *gorm.DB }

func NewTelemetryRepository(db *gorm.DB) *TelemetryRepository { return &TelemetryRepository{db: db} }

var _ telemetry.Repository = (*TelemetryRepository)(nil)

func (r *TelemetryRepository) InsertBatch(ctx context.Context, readings []telemetry.Reading) error {
	if len(readings) == 0 {
		return nil
	}
	models := make([]telemetryModel, len(readings))
	for i, rd := range readings {
		models[i] = readingToModel(rd)
	}
	return translateError(dbFrom(ctx, r.db).CreateInBatches(models, 500).Error, "telemetry")
}

func (r *TelemetryRepository) ListRecent(ctx context.Context, tenantID, deviceID string, since time.Time, limit int) ([]telemetry.Reading, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	var models []telemetryModel
	err := dbFrom(ctx, r.db).
		Where("tenant_id = ? AND device_id = ? AND ts >= ?", tenantID, deviceID, since).
		Order("ts DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "telemetry")
	}
	out := make([]telemetry.Reading, len(models))
	for i := range models {
		out[i] = models[i].toDomain()
	}
	return out, nil
}
