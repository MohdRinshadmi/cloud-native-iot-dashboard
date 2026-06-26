// Package commands implements remote-command use-cases: issuing operations to
// devices (persist → publish over MQTT → mark sent) and processing device
// acknowledgements (status update + OTA firmware application). It depends only
// on ports — the MQTT transport is injected.
package commands

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/ioss/iot-dashboard/backend/internal/application/realtime"
	"github.com/ioss/iot-dashboard/backend/internal/domain/command"
	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// Publisher dispatches a command to a device over the wire (MQTT adapter).
type Publisher interface {
	Publish(ctx context.Context, deviceID string, c *command.Command) error
}

// Service orchestrates command issuance and acknowledgement.
type Service struct {
	commands command.Repository
	devices  device.Repository
	pub      Publisher
	bc       realtime.Broadcaster
	now      func() time.Time
}

// NewService wires the service.
func NewService(
	commands command.Repository,
	devices device.Repository,
	pub Publisher,
	bc realtime.Broadcaster,
	now func() time.Time,
) *Service {
	return &Service{commands: commands, devices: devices, pub: pub, bc: bc, now: now}
}

// IssueInput describes a command to dispatch.
type IssueInput struct {
	TenantID string
	DeviceID string
	IssuedBy string
	Type     command.Type
	Payload  map[string]any
}

// Issue validates ownership, persists the command, publishes it and marks it
// sent. For OTA (set_firmware) it also records the device's target version so
// the UI can show "updating → vX" before the ACK lands.
func (s *Service) Issue(ctx context.Context, in IssueInput) (*command.Command, error) {
	dev, err := s.devices.GetByID(ctx, in.TenantID, in.DeviceID)
	if err != nil {
		return nil, err // 404 for foreign/unknown device
	}

	cmd, err := command.New(uuid.NewString(), in.TenantID, in.DeviceID, in.IssuedBy, in.Type, in.Payload, s.now())
	if err != nil {
		return nil, err
	}

	if cmd.Type == command.TypeSetFirmware {
		if err := dev.RequestFirmware(cmd.FirmwareVersion(), s.now()); err != nil {
			return nil, err
		}
		if err := s.devices.Update(ctx, dev); err != nil {
			return nil, err
		}
	}

	if err := s.commands.Create(ctx, cmd); err != nil {
		return nil, err
	}

	if err := s.pub.Publish(ctx, in.DeviceID, cmd); err != nil {
		// Publish failed: record it so the operator sees the command didn't go out.
		cmd.Ack(false, "publish failed: "+err.Error(), s.now())
		_ = s.commands.Update(ctx, cmd)
		return nil, apperror.Wrap(apperror.CodeUnavailable, "could not dispatch command", err)
	}

	cmd.MarkSent(s.now())
	if err := s.commands.Update(ctx, cmd); err != nil {
		return nil, err
	}
	s.broadcast(cmd)
	return cmd, nil
}

// IssueMany dispatches the same command to many devices (group fan-out, e.g.
// a fleet-wide OTA rollout). Each device acks independently, so the rollout
// progresses device-by-device. Returns how many were dispatched and any
// per-device errors keyed by device id.
func (s *Service) IssueMany(ctx context.Context, tenantID, issuedBy string, deviceIDs []string, t command.Type, payload map[string]any) (issued int, errs map[string]string) {
	errs = map[string]string{}
	for _, id := range deviceIDs {
		_, err := s.Issue(ctx, IssueInput{
			TenantID: tenantID, DeviceID: id, IssuedBy: issuedBy, Type: t, Payload: payload,
		})
		if err != nil {
			errs[id] = err.Error()
			continue
		}
		issued++
	}
	return issued, errs
}

// History returns recent commands for a device (tenant-scoped).
func (s *Service) History(ctx context.Context, tenantID, deviceID string, limit int) ([]*command.Command, error) {
	// Ownership check (404 for foreign devices).
	if _, err := s.devices.GetByID(ctx, tenantID, deviceID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.commands.ListByDevice(ctx, tenantID, deviceID, limit)
}

// ackPayload is the device → server acknowledgement shape, received on
// devices/{id}/commands/ack.
type ackPayload struct {
	CommandID string `json:"command_id"`
	Success   bool   `json:"success"`
	Result    string `json:"result"`
	Firmware  string `json:"firmware"` // applied version (set_firmware only)
}

// HandleAck processes a device acknowledgement (called from the MQTT adapter).
// It is idempotent and trusts only the stored command's tenant/device — the
// inbound deviceID merely has to match the recorded one.
func (s *Service) HandleAck(ctx context.Context, deviceID string, raw []byte) error {
	var ack ackPayload
	if err := json.Unmarshal(raw, &ack); err != nil {
		return err
	}
	if ack.CommandID == "" {
		return apperror.InvalidInput("ack missing command_id")
	}

	cmd, err := s.commands.GetByID(ctx, ack.CommandID)
	if err != nil {
		return err
	}
	// The ack must come from the device the command targeted.
	if cmd.DeviceID != deviceID {
		return apperror.New(apperror.CodeForbidden, "ack device mismatch")
	}
	// Idempotent: ignore duplicate acks for an already-finalized command.
	if cmd.Status == command.StatusAcked || cmd.Status == command.StatusFailed {
		return nil
	}

	cmd.Ack(ack.Success, ack.Result, s.now())
	if err := s.commands.Update(ctx, cmd); err != nil {
		return err
	}

	// OTA completion: apply the firmware to the device record.
	if ack.Success && cmd.Type == command.TypeSetFirmware {
		dev, err := s.devices.GetByID(ctx, cmd.TenantID, cmd.DeviceID)
		if err == nil {
			version := ack.Firmware
			if version == "" {
				version = cmd.FirmwareVersion()
			}
			dev.ApplyFirmware(version, s.now())
			_ = s.devices.Update(ctx, dev)
		}
	}

	s.broadcast(cmd)
	return nil
}

// commandEventData is the realtime payload for command lifecycle changes.
type commandEventData struct {
	ID       string `json:"id"`
	DeviceID string `json:"device_id"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Result   string `json:"result,omitempty"`
}

func (s *Service) broadcast(cmd *command.Command) {
	s.bc.Broadcast(realtime.Event{
		Type:     realtime.EventCommand,
		TenantID: cmd.TenantID,
		Data: commandEventData{
			ID: cmd.ID, DeviceID: cmd.DeviceID, Type: string(cmd.Type),
			Status: string(cmd.Status), Result: cmd.Result,
		},
	})
}
