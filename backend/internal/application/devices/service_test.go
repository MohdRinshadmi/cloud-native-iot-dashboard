package devices_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/application/devices"
	"github.com/ioss/iot-dashboard/backend/internal/domain/device"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// fakeRepo is an in-memory device.Repository for use-case tests.
type fakeRepo struct {
	items map[string]*device.Device
}

func newFakeRepo() *fakeRepo { return &fakeRepo{items: map[string]*device.Device{}} }

func (f *fakeRepo) Create(_ context.Context, d *device.Device) error {
	f.items[d.ID] = d
	return nil
}

func (f *fakeRepo) GetByID(_ context.Context, tenantID, id string) (*device.Device, error) {
	d, ok := f.items[id]
	if !ok || d.TenantID != tenantID {
		return nil, apperror.NotFound("device not found")
	}
	return d, nil
}

func (f *fakeRepo) Update(_ context.Context, d *device.Device) error {
	f.items[d.ID] = d
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, tenantID, id string) error {
	d, ok := f.items[id]
	if !ok || d.TenantID != tenantID {
		return apperror.NotFound("device not found")
	}
	delete(f.items, id)
	return nil
}

func (f *fakeRepo) CountByStatus(_ context.Context, tenantID string) (map[device.Status]int64, error) {
	out := map[device.Status]int64{}
	for _, d := range f.items {
		if d.TenantID == tenantID {
			out[d.Status]++
		}
	}
	return out, nil
}

func (f *fakeRepo) List(_ context.Context, tenantID string, fl device.Filter) ([]*device.Device, int64, error) {
	var all []*device.Device
	for _, d := range f.items {
		if d.TenantID != tenantID {
			continue
		}
		if fl.Q != "" && !strings.Contains(strings.ToLower(d.Name), strings.ToLower(fl.Q)) {
			continue
		}
		if fl.Status != "" && d.Status != fl.Status {
			continue
		}
		all = append(all, d)
	}
	return all, int64(len(all)), nil
}

func fixedNow() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) }

func TestService_Create_SetsProvisioningAndPersists(t *testing.T) {
	repo := newFakeRepo()
	svc := devices.NewService(repo, fixedNow)

	d, err := svc.Create(context.Background(), devices.CreateInput{
		TenantID: "t1", Name: "Pump A", Model: "PX-100", Firmware: "1.0.0",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if d.Status != device.StatusProvisioning {
		t.Errorf("status = %s, want provisioning", d.Status)
	}
	if d.ID == "" {
		t.Error("expected generated id")
	}
	if len(repo.items) != 1 {
		t.Errorf("persisted %d, want 1", len(repo.items))
	}
}

func TestService_Create_RejectsEmptyName(t *testing.T) {
	svc := devices.NewService(newFakeRepo(), fixedNow)
	_, err := svc.Create(context.Background(), devices.CreateInput{TenantID: "t1"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestService_Update_PartialPatch(t *testing.T) {
	repo := newFakeRepo()
	svc := devices.NewService(repo, fixedNow)
	d, _ := svc.Create(context.Background(), devices.CreateInput{
		TenantID: "t1", Name: "Pump A", Model: "PX-100",
	})

	fw := "2.1.0"
	got, err := svc.Update(context.Background(), "t1", d.ID, devices.UpdateInput{Firmware: &fw})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Firmware != "2.1.0" {
		t.Errorf("firmware = %q, want 2.1.0", got.Firmware)
	}
	if got.Name != "Pump A" || got.Model != "PX-100" {
		t.Error("unrelated fields must be preserved on partial update")
	}
}

func TestService_TenantIsolation(t *testing.T) {
	repo := newFakeRepo()
	svc := devices.NewService(repo, fixedNow)
	d, _ := svc.Create(context.Background(), devices.CreateInput{TenantID: "t1", Name: "Pump A"})

	if _, err := svc.Get(context.Background(), "t2", d.ID); err == nil {
		t.Error("cross-tenant read must fail")
	}
	if err := svc.Delete(context.Background(), "t2", d.ID); err == nil {
		t.Error("cross-tenant delete must fail")
	}
}

func TestService_List_RejectsUnknownStatus(t *testing.T) {
	svc := devices.NewService(newFakeRepo(), fixedNow)
	_, _, err := svc.List(context.Background(), devices.ListInput{TenantID: "t1", Status: "warp-speed"})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestService_Summary_CountsAndZeroFills(t *testing.T) {
	repo := newFakeRepo()
	svc := devices.NewService(repo, fixedNow)
	ctx := context.Background()

	// Two devices: one stays provisioning, one goes online.
	_, _ = svc.Create(ctx, devices.CreateInput{TenantID: "t1", Name: "a"})
	d, _ := svc.Create(ctx, devices.CreateInput{TenantID: "t1", Name: "b"})
	d.Status = device.StatusOnline
	repo.items[d.ID] = d
	// Other tenant device must not leak into the summary.
	_, _ = svc.Create(ctx, devices.CreateInput{TenantID: "t2", Name: "c"})

	sum, err := svc.Summary(ctx, "t1")
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if sum.Total != 2 {
		t.Errorf("total = %d, want 2 (tenant-scoped)", sum.Total)
	}
	if sum.Online != 1 {
		t.Errorf("online = %d, want 1", sum.Online)
	}
	// Zero-fill: a status with no devices must still be present.
	if _, ok := sum.ByStatus[device.StatusDecommissioned]; !ok {
		t.Error("ByStatus must zero-fill every status key")
	}
}
