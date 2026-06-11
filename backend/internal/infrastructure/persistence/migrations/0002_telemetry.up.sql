-- 0002_telemetry: append-only telemetry history (the "hot" relational store).
-- At 1M-device scale this table moves to a time-series store (Timescale /
-- DynamoDB+S3); the Repository port isolates that swap from business code.

CREATE TABLE telemetry (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID        NOT NULL,
    device_id   UUID        NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    temperature DOUBLE PRECISION,
    battery     DOUBLE PRECISION,
    voltage     DOUBLE PRECISION,
    cpu         DOUBLE PRECISION,
    memory      DOUBLE PRECISION,
    signal      DOUBLE PRECISION,
    lat         DOUBLE PRECISION,
    lng         DOUBLE PRECISION,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Chart queries are always "one device, recent window, newest first".
CREATE INDEX idx_telemetry_device_ts ON telemetry (device_id, ts DESC);
-- Fleet-wide analytics scan by tenant + window.
CREATE INDEX idx_telemetry_tenant_ts ON telemetry (tenant_id, ts DESC);
