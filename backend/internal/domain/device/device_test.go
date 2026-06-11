package device_test

import (
	"testing"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
)

func TestNewDevice_EnforcesInvariants(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		id      string
		tenant  string
		device  string
		wantErr bool
	}{
		{"valid", "dev-1", "tenant-1", "Sensor A", false},
		{"missing id", "", "tenant-1", "Sensor A", true},
		{"missing tenant", "dev-1", "", "Sensor A", true},
		{"missing name", "dev-1", "tenant-1", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, err := device.NewDevice(tc.id, tc.tenant, tc.device, "model-x", now)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Status != device.StatusProvisioning {
				t.Errorf("new device should be provisioning, got %s", d.Status)
			}
		})
	}
}

func TestDevice_MarkSeen_TransitionsOnline(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d, err := device.NewDevice("dev-1", "tenant-1", "Sensor A", "model-x", now)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	beat := now.Add(time.Minute)
	d.MarkSeen(beat)

	if d.Status != device.StatusOnline {
		t.Errorf("expected online after heartbeat, got %s", d.Status)
	}
	if !d.IsOnline(beat.Add(30*time.Second), time.Minute) {
		t.Error("device should be online within window")
	}
	if d.IsOnline(beat.Add(2*time.Minute), time.Minute) {
		t.Error("device should be offline outside window")
	}
}
