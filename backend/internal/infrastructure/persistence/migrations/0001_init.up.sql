-- 0001_init: core multi-tenant schema.
-- Conventions: UUID PKs (generated in the app), TIMESTAMPTZ everywhere,
-- ON DELETE CASCADE inside a tenant boundary.

CREATE TABLE tenants (
    id          UUID PRIMARY KEY,
    name        TEXT        NOT NULL,
    slug        TEXT        NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE users (
    id            UUID PRIMARY KEY,
    tenant_id     UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email         TEXT        NOT NULL UNIQUE,          -- global login identifier
    name          TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL,                 -- admin | operator | viewer
    created_at    TIMESTAMPTZ NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_users_tenant ON users (tenant_id);

CREATE TABLE devices (
    id           UUID PRIMARY KEY,
    tenant_id    UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name         TEXT        NOT NULL,
    model        TEXT        NOT NULL DEFAULT '',
    firmware     TEXT        NOT NULL DEFAULT '',
    status       TEXT        NOT NULL,                  -- provisioning|online|offline|degraded|decommissioned
    last_seen_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL
);
-- Every device query is tenant-scoped; status is the hot filter.
CREATE INDEX idx_devices_tenant        ON devices (tenant_id);
CREATE INDEX idx_devices_tenant_status ON devices (tenant_id, status);

CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,             -- sha-256 hex; raw token never stored
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);
