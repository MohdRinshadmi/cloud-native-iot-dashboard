// Package command is the remote-command bounded context: the audit trail and
// status lifecycle for operations dispatched to devices (reboot, config push,
// OTA firmware update). The command is the record; transport (MQTT) lives in
// infrastructure.
package command

import (
	"context"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// Type enumerates the supported remote operations.
type Type string

const (
	TypeReboot      Type = "reboot"
	TypeConfigPush  Type = "config_push"
	TypeSetFirmware Type = "set_firmware" // OTA
)

// Valid reports whether t is a known command type.
func (t Type) Valid() bool {
	switch t {
	case TypeReboot, TypeConfigPush, TypeSetFirmware:
		return true
	default:
		return false
	}
}

// Status is the command lifecycle. queued → sent → acked | failed.
type Status string

const (
	StatusQueued Status = "queued"
	StatusSent   Status = "sent"
	StatusAcked  Status = "acked"
	StatusFailed Status = "failed"
)

// Command is the persisted record of one dispatched operation.
type Command struct {
	ID        string
	TenantID  string
	DeviceID  string
	Type      Type
	Payload   map[string]any
	Status    Status
	Result    string
	IssuedBy  string
	CreatedAt time.Time
	UpdatedAt time.Time
	AckedAt   *time.Time
}

// New constructs a queued command, validating type-specific payload.
func New(id, tenantID, deviceID, issuedBy string, t Type, payload map[string]any, now time.Time) (*Command, error) {
	switch {
	case id == "":
		return nil, apperror.InvalidInput("command id is required")
	case tenantID == "" || deviceID == "":
		return nil, apperror.InvalidInput("tenant and device are required")
	case !t.Valid():
		return nil, apperror.InvalidInput("unknown command type")
	}
	if payload == nil {
		payload = map[string]any{}
	}
	// OTA requires a target version.
	if t == TypeSetFirmware {
		if v, ok := payload["version"].(string); !ok || v == "" {
			return nil, apperror.InvalidInput("set_firmware requires a non-empty version")
		}
	}
	return &Command{
		ID: id, TenantID: tenantID, DeviceID: deviceID, IssuedBy: issuedBy,
		Type: t, Payload: payload, Status: StatusQueued,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

// MarkSent records that the command was published to the broker.
func (c *Command) MarkSent(now time.Time) {
	c.Status = StatusSent
	c.UpdatedAt = now
}

// Ack applies a device acknowledgement, moving to acked/failed.
func (c *Command) Ack(success bool, result string, now time.Time) {
	if success {
		c.Status = StatusAcked
	} else {
		c.Status = StatusFailed
	}
	c.Result = result
	c.AckedAt = &now
	c.UpdatedAt = now
}

// FirmwareVersion extracts the OTA target from a set_firmware payload.
func (c *Command) FirmwareVersion() string {
	if v, ok := c.Payload["version"].(string); ok {
		return v
	}
	return ""
}

// Repository is the persistence port for commands.
type Repository interface {
	Create(ctx context.Context, c *Command) error
	Update(ctx context.Context, c *Command) error
	// GetByID is tenant-unscoped: ACK handling resolves a command from an
	// untrusted device message by id, then trusts the stored tenant/device.
	GetByID(ctx context.Context, id string) (*Command, error)
	ListByDevice(ctx context.Context, tenantID, deviceID string, limit int) ([]*Command, error)
}
