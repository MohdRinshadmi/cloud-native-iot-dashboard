// Package device is the Device bounded context. It contains pure domain code:
// entities, value objects, invariants and ports. It imports NOTHING from
// infrastructure or transport — the dependency rule points strictly inward.
package device

import (
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// Status is the lifecycle/connectivity state of a device.
type Status string

const (
	StatusProvisioning   Status = "provisioning" // created, not yet seen
	StatusOnline         Status = "online"       // heartbeat within threshold
	StatusOffline        Status = "offline"      // heartbeat lapsed
	StatusDegraded       Status = "degraded"     // online but unhealthy telemetry
	StatusDecommissioned Status = "decommissioned"
)

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	switch s {
	case StatusProvisioning, StatusOnline, StatusOffline, StatusDegraded, StatusDecommissioned:
		return true
	default:
		return false
	}
}

// Device is the aggregate root of this context. It belongs to exactly one
// tenant (multi-tenancy is enforced at the domain boundary, not bolted on).
type Device struct {
	ID         string
	TenantID   string
	GroupID    *string // optional fleet/group membership
	Name       string
	Model      string
	Firmware   string  // currently-running version
	TargetFirmware string // desired version (set on OTA, cleared once applied)
	Status     Status
	LastSeenAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewDevice constructs a valid Device in the provisioning state, enforcing
// invariants at creation time. `now` is injected to keep the domain
// deterministic and testable (no hidden time.Now()).
func NewDevice(id, tenantID, name, model string, now time.Time) (*Device, error) {
	if id == "" {
		return nil, apperror.InvalidInput("device id is required")
	}
	if tenantID == "" {
		return nil, apperror.InvalidInput("tenant id is required")
	}
	if name == "" {
		return nil, apperror.InvalidInput("device name is required")
	}
	return &Device{
		ID:        id,
		TenantID:  tenantID,
		Name:      name,
		Model:     model,
		Status:    StatusProvisioning,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Rename updates the device's descriptive fields, enforcing invariants.
// Empty-string arguments mean "leave unchanged" for model/firmware; the name
// may never become empty.
func (d *Device) Rename(name, model, firmware string, now time.Time) error {
	if name == "" {
		return apperror.InvalidInput("device name cannot be empty")
	}
	d.Name = name
	if model != "" {
		d.Model = model
	}
	if firmware != "" {
		d.Firmware = firmware
	}
	d.UpdatedAt = now
	return nil
}

// AssignGroup moves the device into a group (or nil to ungroup).
func (d *Device) AssignGroup(groupID *string, now time.Time) {
	d.GroupID = groupID
	d.UpdatedAt = now
}

// RequestFirmware records the desired OTA target version.
func (d *Device) RequestFirmware(version string, now time.Time) error {
	if version == "" {
		return apperror.InvalidInput("target firmware version is required")
	}
	d.TargetFirmware = version
	d.UpdatedAt = now
	return nil
}

// ApplyFirmware marks an OTA complete: the running version becomes the target
// and the pending target clears. Called when the device ACKs the update.
func (d *Device) ApplyFirmware(version string, now time.Time) {
	d.Firmware = version
	if d.TargetFirmware == version {
		d.TargetFirmware = ""
	}
	d.UpdatedAt = now
}

// MarkSeen records a heartbeat, transitioning the device online.
func (d *Device) MarkSeen(at time.Time) {
	d.LastSeenAt = &at
	if d.Status == StatusProvisioning || d.Status == StatusOffline {
		d.Status = StatusOnline
	}
	d.UpdatedAt = at
}

// IsOnline reports whether the last heartbeat is within the given window.
func (d *Device) IsOnline(now time.Time, window time.Duration) bool {
	if d.LastSeenAt == nil {
		return false
	}
	return now.Sub(*d.LastSeenAt) <= window
}
