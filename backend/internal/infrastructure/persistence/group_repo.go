package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ioss/iot-dashboard/backend/internal/domain/group"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

type groupModel struct {
	ID          string `gorm:"primaryKey"`
	TenantID    string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (groupModel) TableName() string { return "device_groups" }

func groupToModel(g *group.Group) *groupModel {
	return &groupModel{
		ID: g.ID, TenantID: g.TenantID, Name: g.Name, Description: g.Description,
		CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
}

func (m *groupModel) toDomain() *group.Group {
	return &group.Group{
		ID: m.ID, TenantID: m.TenantID, Name: m.Name, Description: m.Description,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

// GroupRepository is the GORM adapter for group.Repository.
type GroupRepository struct{ db *gorm.DB }

func NewGroupRepository(db *gorm.DB) *GroupRepository { return &GroupRepository{db: db} }

var _ group.Repository = (*GroupRepository)(nil)

func (r *GroupRepository) Create(ctx context.Context, g *group.Group) error {
	return translateError(dbFrom(ctx, r.db).Create(groupToModel(g)).Error, "group not found")
}

func (r *GroupRepository) GetByID(ctx context.Context, tenantID, id string) (*group.Group, error) {
	var m groupModel
	err := dbFrom(ctx, r.db).Where("tenant_id = ? AND id = ?", tenantID, id).First(&m).Error
	if err != nil {
		return nil, translateError(err, "group not found")
	}
	return m.toDomain(), nil
}

func (r *GroupRepository) Update(ctx context.Context, g *group.Group) error {
	res := dbFrom(ctx, r.db).Model(&groupModel{}).
		Where("tenant_id = ? AND id = ?", g.TenantID, g.ID).
		Updates(map[string]any{"name": g.Name, "description": g.Description, "updated_at": g.UpdatedAt})
	if res.Error != nil {
		return translateError(res.Error, "group not found")
	}
	if res.RowsAffected == 0 {
		return apperror.NotFound("group not found")
	}
	return nil
}

func (r *GroupRepository) Delete(ctx context.Context, tenantID, id string) error {
	res := dbFrom(ctx, r.db).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&groupModel{})
	if res.Error != nil {
		return translateError(res.Error, "group not found")
	}
	if res.RowsAffected == 0 {
		return apperror.NotFound("group not found")
	}
	return nil
}

func (r *GroupRepository) List(ctx context.Context, tenantID string) ([]*group.Group, error) {
	var models []groupModel
	err := dbFrom(ctx, r.db).Where("tenant_id = ?", tenantID).Order("name").Find(&models).Error
	if err != nil {
		return nil, translateError(err, "group not found")
	}
	out := make([]*group.Group, len(models))
	for i := range models {
		out[i] = models[i].toDomain()
	}
	return out, nil
}

func (r *GroupRepository) DeviceCounts(ctx context.Context, tenantID string) (map[string]int64, error) {
	type row struct {
		GroupID string
		Count   int64
	}
	var rows []row
	err := dbFrom(ctx, r.db).Model(&deviceModel{}).
		Select("group_id, count(*) as count").
		Where("tenant_id = ? AND group_id IS NOT NULL", tenantID).
		Group("group_id").
		Scan(&rows).Error
	if err != nil {
		return nil, translateError(err, "group not found")
	}
	out := make(map[string]int64, len(rows))
	for _, rw := range rows {
		out[rw.GroupID] = rw.Count
	}
	return out, nil
}
