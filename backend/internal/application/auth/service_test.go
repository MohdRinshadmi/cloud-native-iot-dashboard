package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/application/auth"
	domauth "github.com/ioss/iot-dashboard/backend/internal/domain/auth"
	"github.com/ioss/iot-dashboard/backend/internal/domain/tenant"
	"github.com/ioss/iot-dashboard/backend/internal/domain/user"
	"github.com/ioss/iot-dashboard/backend/internal/shared/apperror"
)

// ---- fakes -----------------------------------------------------------------

type fakeUsers struct{ byID, byEmail map[string]*user.User }

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byID: map[string]*user.User{}, byEmail: map[string]*user.User{}}
}
func (f *fakeUsers) Create(_ context.Context, u *user.User) error {
	if _, dup := f.byEmail[u.Email]; dup {
		return apperror.New(apperror.CodeConflict, "email already registered")
	}
	f.byID[u.ID], f.byEmail[u.Email] = u, u
	return nil
}
func (f *fakeUsers) GetByEmail(_ context.Context, email string) (*user.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, apperror.NotFound("user not found")
}
func (f *fakeUsers) GetByID(_ context.Context, id string) (*user.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, apperror.NotFound("user not found")
}

type fakeTenants struct{ bySlug map[string]*tenant.Tenant }

func newFakeTenants() *fakeTenants { return &fakeTenants{bySlug: map[string]*tenant.Tenant{}} }
func (f *fakeTenants) Create(_ context.Context, t *tenant.Tenant) error {
	if _, dup := f.bySlug[t.Slug]; dup {
		return apperror.New(apperror.CodeConflict, "tenant exists")
	}
	f.bySlug[t.Slug] = t
	return nil
}
func (f *fakeTenants) GetByID(_ context.Context, _ string) (*tenant.Tenant, error) {
	return nil, apperror.NotFound("not implemented")
}
func (f *fakeTenants) GetBySlug(_ context.Context, slug string) (*tenant.Tenant, error) {
	if t, ok := f.bySlug[slug]; ok {
		return t, nil
	}
	return nil, apperror.NotFound("tenant not found")
}

type fakeRefresh struct {
	byHash map[string]*domauth.RefreshToken
}

func newFakeRefresh() *fakeRefresh {
	return &fakeRefresh{byHash: map[string]*domauth.RefreshToken{}}
}
func (f *fakeRefresh) Create(_ context.Context, t *domauth.RefreshToken) error {
	f.byHash[t.TokenHash] = t
	return nil
}
func (f *fakeRefresh) GetByHash(_ context.Context, h string) (*domauth.RefreshToken, error) {
	if t, ok := f.byHash[h]; ok {
		return t, nil
	}
	return nil, apperror.NotFound("token not found")
}
func (f *fakeRefresh) Revoke(_ context.Context, id string, at time.Time) error {
	for _, t := range f.byHash {
		if t.ID == id {
			t.RevokedAt = &at
			return nil
		}
	}
	return apperror.NotFound("token not found")
}
func (f *fakeRefresh) RevokeAllForUser(_ context.Context, userID string, at time.Time) error {
	for _, t := range f.byHash {
		if t.UserID == userID && t.RevokedAt == nil {
			t.RevokedAt = &at
		}
	}
	return nil
}
func (f *fakeRefresh) activeCount(userID string) int {
	n := 0
	for _, t := range f.byHash {
		if t.UserID == userID && t.RevokedAt == nil {
			n++
		}
	}
	return n
}

type passthroughTx struct{}

func (passthroughTx) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// fakeHasher does reversible "hashing" — good enough to assert flow logic.
type fakeHasher struct{}

func (fakeHasher) Hash(p string) (string, error) { return "hashed:" + p, nil }
func (fakeHasher) Compare(hash, plain string) error {
	if hash != "hashed:"+plain {
		return errors.New("mismatch")
	}
	return nil
}

type fakeIssuer struct{}

func (fakeIssuer) Issue(u *user.User) (string, time.Duration, error) {
	return "access-for-" + u.ID, 15 * time.Minute, nil
}

// ---- harness ---------------------------------------------------------------

type harness struct {
	svc     *auth.Service
	users   *fakeUsers
	refresh *fakeRefresh
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	users, tenants, refresh := newFakeUsers(), newFakeTenants(), newFakeRefresh()
	now := func() time.Time { return time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) }
	svc := auth.NewService(users, tenants, refresh, passthroughTx{}, fakeHasher{}, fakeIssuer{}, 7*24*time.Hour, now)
	return &harness{svc: svc, users: users, refresh: refresh}
}

func (h *harness) register(t *testing.T) *auth.Session {
	t.Helper()
	s, err := h.svc.Register(context.Background(), auth.RegisterInput{
		TenantName: "Acme Corp", Email: "admin@acme.io", Name: "Ada", Password: "Password123!",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return s
}

// ---- tests -----------------------------------------------------------------

func TestRegister_CreatesTenantAdminAndSession(t *testing.T) {
	h := newHarness(t)
	s := h.register(t)

	if s.User.Role != user.RoleAdmin {
		t.Errorf("first user role = %s, want admin", s.User.Role)
	}
	if s.AccessToken == "" || s.RefreshToken == "" {
		t.Error("session must include both tokens")
	}
}

func TestRegister_RejectsShortPassword(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.Register(context.Background(), auth.RegisterInput{
		TenantName: "Acme", Email: "a@b.co", Name: "A", Password: "short",
	})
	if err == nil {
		t.Fatal("expected password policy error")
	}
}

func TestLogin_SuccessAndGenericFailure(t *testing.T) {
	h := newHarness(t)
	h.register(t)

	if _, err := h.svc.Login(context.Background(), "admin@acme.io", "Password123!"); err != nil {
		t.Fatalf("valid login failed: %v", err)
	}

	_, errWrongPw := h.svc.Login(context.Background(), "admin@acme.io", "nope")
	_, errNoUser := h.svc.Login(context.Background(), "ghost@acme.io", "nope")
	if errWrongPw == nil || errNoUser == nil {
		t.Fatal("invalid logins must fail")
	}
	// Same message for both — no account enumeration.
	if errWrongPw.Error() != errNoUser.Error() {
		t.Error("login failures must be indistinguishable")
	}
}

func TestRefresh_RotatesToken(t *testing.T) {
	h := newHarness(t)
	s := h.register(t)

	s2, err := h.svc.Refresh(context.Background(), s.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if s2.RefreshToken == s.RefreshToken {
		t.Error("refresh must rotate the token")
	}
	if got := h.refresh.activeCount(s.User.ID); got != 1 {
		t.Errorf("active tokens = %d, want 1 (old revoked, new active)", got)
	}
}

func TestRefresh_ReuseDetectionRevokesAllSessions(t *testing.T) {
	h := newHarness(t)
	s := h.register(t)

	// Legitimate rotation...
	if _, err := h.svc.Refresh(context.Background(), s.RefreshToken); err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	// ...then an attacker replays the OLD token.
	if _, err := h.svc.Refresh(context.Background(), s.RefreshToken); err == nil {
		t.Fatal("replayed token must be rejected")
	}
	if got := h.refresh.activeCount(s.User.ID); got != 0 {
		t.Errorf("after reuse detection active tokens = %d, want 0 (all revoked)", got)
	}
}

func TestLogout_RevokesAndIsIdempotent(t *testing.T) {
	h := newHarness(t)
	s := h.register(t)

	if err := h.svc.Logout(context.Background(), s.RefreshToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if got := h.refresh.activeCount(s.User.ID); got != 0 {
		t.Errorf("active tokens after logout = %d, want 0", got)
	}
	if err := h.svc.Logout(context.Background(), s.RefreshToken); err != nil {
		t.Errorf("repeat logout must be a no-op, got %v", err)
	}
}
