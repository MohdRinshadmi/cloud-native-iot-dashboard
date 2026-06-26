package persistence

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/datatypes"

	"github.com/ioss/iot-dashboard/backend/internal/domain/command"
)

type commandModel struct {
	ID        string `gorm:"primaryKey"`
	TenantID  string
	DeviceID  string
	Type      string
	Payload   datatypes.JSON
	Status    string
	Result    string
	IssuedBy  *string
	CreatedAt time.Time
	UpdatedAt time.Time
	AckedAt   *time.Time
}

func (commandModel) TableName() string { return "commands" }

func commandToModel(c *command.Command) *commandModel {
	payload, _ := json.Marshal(c.Payload)
	var issuedBy *string
	if c.IssuedBy != "" {
		issuedBy = &c.IssuedBy
	}
	return &commandModel{
		ID: c.ID, TenantID: c.TenantID, DeviceID: c.DeviceID,
		Type: string(c.Type), Payload: datatypes.JSON(payload),
		Status: string(c.Status), Result: c.Result, IssuedBy: issuedBy,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, AckedAt: c.AckedAt,
	}
}

func (m *commandModel) toDomain() *command.Command {
	payload := map[string]any{}
	if len(m.Payload) > 0 {
		_ = json.Unmarshal(m.Payload, &payload)
	}
	issuedBy := ""
	if m.IssuedBy != nil {
		issuedBy = *m.IssuedBy
	}
	return &command.Command{
		ID: m.ID, TenantID: m.TenantID, DeviceID: m.DeviceID,
		Type: command.Type(m.Type), Payload: payload,
		Status: command.Status(m.Status), Result: m.Result, IssuedBy: issuedBy,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, AckedAt: m.AckedAt,
	}
}

// CommandRepository is the GORM adapter for command.Repository.
type CommandRepository struct{ db *gorm.DB }

func NewCommandRepository(db *gorm.DB) *CommandRepository { return &CommandRepository{db: db} }

var _ command.Repository = (*CommandRepository)(nil)

func (r *CommandRepository) Create(ctx context.Context, c *command.Command) error {
	return translateError(dbFrom(ctx, r.db).Create(commandToModel(c)).Error, "command not found")
}

func (r *CommandRepository) Update(ctx context.Context, c *command.Command) error {
	m := commandToModel(c)
	res := dbFrom(ctx, r.db).Model(&commandModel{}).Where("id = ?", c.ID).
		Updates(map[string]any{
			"status": m.Status, "result": m.Result, "updated_at": m.UpdatedAt, "acked_at": m.AckedAt,
		})
	return translateError(res.Error, "command not found")
}

func (r *CommandRepository) GetByID(ctx context.Context, id string) (*command.Command, error) {
	var m commandModel
	err := dbFrom(ctx, r.db).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, translateError(err, "command not found")
	}
	return m.toDomain(), nil
}

func (r *CommandRepository) ListByDevice(ctx context.Context, tenantID, deviceID string, limit int) ([]*command.Command, error) {
	var models []commandModel
	err := dbFrom(ctx, r.db).
		Where("tenant_id = ? AND device_id = ?", tenantID, deviceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "command not found")
	}
	out := make([]*command.Command, len(models))
	for i := range models {
		out[i] = models[i].toDomain()
	}
	return out, nil
}
