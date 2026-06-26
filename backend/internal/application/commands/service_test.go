package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/application/commands"
	"github.com/ioss/iot-dashboard/backend/internal/application/realtime"
	"github.com/ioss/iot-dashboard/backend/internal/domain/command"
	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// ---- fakes -----------------------------------------------------------------

type fakeCommands struct{ byID map[string]*command.Command }

func newFakeCommands() *fakeCommands { return &fakeCommands{byID: map[string]*command.Command{}} }
func (f *fakeCommands) Create(_ context.Context, c *command.Command) error {
	cp := *c
	f.byID[c.ID] = &cp
	return nil
}
func (f *fakeCommands) Update(_ context.Context, c *command.Command) error {
	cp := *c
	f.byID[c.ID] = &cp
	return nil
}
func (f *fakeCommands) GetByID(_ context.Context, id string) (*command.Command, error) {
	if c, ok := f.byID[id]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, apperror.NotFound("command not found")
}
func (f *fakeCommands) ListByDevice(_ context.Context, _, deviceID string, _ int) ([]*command.Command, error) {
	var out []*command.Command
	for _, c := range f.byID {
		if c.DeviceID == deviceID {
			out = append(out, c)
		}
	}
	return out, nil
}

type fakeDevices struct{ byID map[string]*device.Device }

func (f *fakeDevices) Create(context.Context, *device.Device) error { return nil }
func (f *fakeDevices) GetByID(_ context.Context, tenantID, id string) (*device.Device, error) {
	d, ok := f.byID[id]
	if !ok || d.TenantID != tenantID {
		return nil, apperror.NotFound("device not found")
	}
	cp := *d
	return &cp, nil
}
func (f *fakeDevices) Update(_ context.Context, d *device.Device) error {
	cp := *d
	f.byID[d.ID] = &cp
	return nil
}
func (f *fakeDevices) Delete(context.Context, string, string) error { return nil }
func (f *fakeDevices) List(context.Context, string, device.Filter) ([]*device.Device, int64, error) {
	return nil, 0, nil
}
func (f *fakeDevices) CountByStatus(context.Context, string) (map[device.Status]int64, error) {
	return nil, nil
}

type capturePublisher struct {
	published []*command.Command
	failNext  bool
}

func (p *capturePublisher) Publish(_ context.Context, _ string, c *command.Command) error {
	if p.failNext {
		return errors.New("broker down")
	}
	p.published = append(p.published, c)
	return nil
}

func fixedNow() time.Time { return time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC) }

func newSvc(t *testing.T) (*commands.Service, *fakeCommands, *fakeDevices, *capturePublisher) {
	t.Helper()
	cmds := newFakeCommands()
	devs := &fakeDevices{byID: map[string]*device.Device{
		"dev-1": {ID: "dev-1", TenantID: "t1", Name: "pump", Firmware: "1.0.0", Status: device.StatusOnline},
	}}
	pub := &capturePublisher{}
	svc := commands.NewService(cmds, devs, pub, realtime.NopBroadcaster{}, fixedNow)
	return svc, cmds, devs, pub
}

// ---- tests -----------------------------------------------------------------

func TestIssue_PersistsPublishesAndMarksSent(t *testing.T) {
	svc, _, _, pub := newSvc(t)

	cmd, err := svc.Issue(context.Background(), commands.IssueInput{
		TenantID: "t1", DeviceID: "dev-1", IssuedBy: "u1", Type: command.TypeReboot,
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if cmd.Status != command.StatusSent {
		t.Errorf("status = %s, want sent", cmd.Status)
	}
	if len(pub.published) != 1 {
		t.Errorf("published %d, want 1", len(pub.published))
	}
}

func TestIssue_ForeignDeviceRejected(t *testing.T) {
	svc, _, _, _ := newSvc(t)
	_, err := svc.Issue(context.Background(), commands.IssueInput{
		TenantID: "t2", DeviceID: "dev-1", Type: command.TypeReboot,
	})
	if err == nil {
		t.Fatal("cross-tenant issue must fail")
	}
}

func TestIssue_SetFirmwareRecordsTarget(t *testing.T) {
	svc, _, devs, _ := newSvc(t)
	_, err := svc.Issue(context.Background(), commands.IssueInput{
		TenantID: "t1", DeviceID: "dev-1", Type: command.TypeSetFirmware,
		Payload: map[string]any{"version": "2.0.0"},
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if devs.byID["dev-1"].TargetFirmware != "2.0.0" {
		t.Errorf("target firmware = %q, want 2.0.0", devs.byID["dev-1"].TargetFirmware)
	}
}

func TestIssue_PublishFailureMarksFailed(t *testing.T) {
	svc, cmds, _, pub := newSvc(t)
	pub.failNext = true
	_, err := svc.Issue(context.Background(), commands.IssueInput{
		TenantID: "t1", DeviceID: "dev-1", Type: command.TypeReboot,
	})
	if err == nil {
		t.Fatal("expected dispatch error")
	}
	// The command record should exist and be marked failed.
	var found *command.Command
	for _, c := range cmds.byID {
		found = c
	}
	if found == nil || found.Status != command.StatusFailed {
		t.Errorf("expected a failed command record, got %+v", found)
	}
}

func TestHandleAck_OTAAppliesFirmware(t *testing.T) {
	svc, _, devs, _ := newSvc(t)
	cmd, _ := svc.Issue(context.Background(), commands.IssueInput{
		TenantID: "t1", DeviceID: "dev-1", Type: command.TypeSetFirmware,
		Payload: map[string]any{"version": "2.0.0"},
	})

	ack := []byte(`{"command_id":"` + cmd.ID + `","success":true,"firmware":"2.0.0"}`)
	if err := svc.HandleAck(context.Background(), "dev-1", ack); err != nil {
		t.Fatalf("ack: %v", err)
	}

	dev := devs.byID["dev-1"]
	if dev.Firmware != "2.0.0" {
		t.Errorf("firmware = %q, want 2.0.0 (OTA applied)", dev.Firmware)
	}
	if dev.TargetFirmware != "" {
		t.Errorf("target should clear after apply, got %q", dev.TargetFirmware)
	}
}

func TestHandleAck_DeviceMismatchRejected(t *testing.T) {
	svc, _, _, _ := newSvc(t)
	cmd, _ := svc.Issue(context.Background(), commands.IssueInput{
		TenantID: "t1", DeviceID: "dev-1", Type: command.TypeReboot,
	})
	ack := []byte(`{"command_id":"` + cmd.ID + `","success":true}`)
	if err := svc.HandleAck(context.Background(), "someone-else", ack); err == nil {
		t.Fatal("ack from a different device must be rejected")
	}
}

func TestHandleAck_Idempotent(t *testing.T) {
	svc, _, _, _ := newSvc(t)
	cmd, _ := svc.Issue(context.Background(), commands.IssueInput{
		TenantID: "t1", DeviceID: "dev-1", Type: command.TypeReboot,
	})
	ack := []byte(`{"command_id":"` + cmd.ID + `","success":true}`)
	if err := svc.HandleAck(context.Background(), "dev-1", ack); err != nil {
		t.Fatalf("first ack: %v", err)
	}
	if err := svc.HandleAck(context.Background(), "dev-1", ack); err != nil {
		t.Errorf("duplicate ack must be a no-op, got %v", err)
	}
}
