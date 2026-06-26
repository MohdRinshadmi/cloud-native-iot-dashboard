-- 0003_groups_commands: device fleets/groups, firmware targets, and the
-- remote-command audit trail (reboot / config push / OTA firmware update).

CREATE TABLE device_groups (
    id          UUID PRIMARY KEY,
    tenant_id   UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL,
    UNIQUE (tenant_id, name)
);
CREATE INDEX idx_device_groups_tenant ON device_groups (tenant_id);

-- A device belongs to at most one group; firmware target drives OTA tracking.
ALTER TABLE devices
    ADD COLUMN group_id        UUID REFERENCES device_groups(id) ON DELETE SET NULL,
    ADD COLUMN target_firmware TEXT NOT NULL DEFAULT '';
CREATE INDEX idx_devices_group ON devices (group_id);

-- Every remote command is recorded for audit + status tracking.
CREATE TABLE commands (
    id         UUID PRIMARY KEY,
    tenant_id  UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    device_id  UUID        NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    type       TEXT        NOT NULL,        -- reboot | config_push | set_firmware
    payload    JSONB       NOT NULL DEFAULT '{}',
    status     TEXT        NOT NULL,        -- queued | sent | acked | failed
    result     TEXT        NOT NULL DEFAULT '',
    issued_by  UUID,                        -- user who issued it (audit)
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    acked_at   TIMESTAMPTZ
);
CREATE INDEX idx_commands_device ON commands (device_id, created_at DESC);
CREATE INDEX idx_commands_tenant ON commands (tenant_id, created_at DESC);
