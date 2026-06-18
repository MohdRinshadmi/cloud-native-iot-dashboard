<div align="center">

# Cloud-Native IoT Analytics Dashboard

**Real-time telemetry, AI insights and fleet monitoring — engineered for 1,000,000+ devices.**

Think Tesla fleet monitoring · Datadog · Grafana · AWS IoT Console.

`React 19` · `TypeScript` · `Go + Gin` · `PostgreSQL` · `Redis` · `MQTT` · `WebSockets` · `Docker` · `Kubernetes` · `Terraform` · `AWS`

</div>

---

## What this is

A production-grade, multi-tenant IoT analytics platform: devices stream telemetry
over **MQTT**, a **Go (Gin)** backend ingests/processes/scores it, and a
**React 19** command-center dashboard renders it live over **WebSockets** —
with a pluggable AI layer for anomaly detection and predictive maintenance.

Everything is **free, open-source and self-hostable**. The entire stack runs
locally with Docker Compose at **$0**.

## Architecture (local development)

```
┌─────────────┐  MQTT   ┌────────────┐        ┌──────────────────────────┐
│ IoT Devices │ ───────▶│ Mosquitto  │───────▶│  Go + Gin API            │
│ (simulated) │  1883   │  broker    │ consume│  ┌────────────────────┐  │
└─────────────┘         └────────────┘        │  │ domain (DDD)       │  │
                                              │  │ application        │  │
       browser  ◀──── WebSockets ────────────▶│  │ infrastructure     │  │
┌──────────────┐                              │  │ interfaces (HTTP)  │  │
│  React 19    │ ◀──── REST /api/v1 ─────────▶│  └────────────────────┘  │
│  dashboard   │                              └─────┬──────────┬─────────┘
└──────────────┘                                    │          │
                                              ┌─────▼───┐ ┌────▼────┐
                                              │Postgres │ │  Redis  │
                                              │ (truth) │ │  (hot)  │
                                              └─────────┘ └─────────┘
            observability: Prometheus (scrape /metrics) ─▶ Grafana
```

Production target (Phase 13+): AWS IoT Core → EventBridge/SQS → EKS-hosted
services → DynamoDB hot path / S3 cold path, provisioned with Terraform.

## Repository layout

```
.
├── backend/                  # Go + Gin — Clean Architecture / DDD
│   ├── cmd/api/              #   composition root (main.go, DI wiring)
│   ├── configs/              #   reference configuration
│   ├── internal/
│   │   ├── domain/           #   entities, value objects, ports (pure Go)
│   │   ├── application/      #   use-cases / services
│   │   ├── infrastructure/   #   config, logging, server, DB/MQTT adapters
│   │   ├── interfaces/       #   HTTP transport: handlers, middleware, router
│   │   └── shared/           #   cross-layer primitives (apperror)
│   └── pkg/                  #   public reusable helpers
├── frontend/                 # React 19 + Vite + TypeScript (strict)
│   └── src/
│       ├── app/              #   app shell: providers, router
│       ├── routes/           #   TanStack Router route modules
│       ├── features/         #   feature-based modules (UI + logic per domain)
│       ├── shared/           #   design-system components, lib utilities
│       ├── services/api/     #   typed API client layer
│       ├── hooks/            #   cross-feature hooks (TanStack Query)
│       ├── stores/           #   Zustand global state (auth/RBAC)
│       ├── types/ utils/     #   shared types & helpers
│       └── assets/
├── infra/                    # local infra configs (mosquitto, prometheus, grafana)
├── docker-compose.yml        # full local stack
├── Makefile                  # developer command center
└── .env.example              # single source of env truth
```

## Quick start

Prereqs: Docker Desktop, Go ≥ 1.22, Node ≥ 20.

```bash
# 1. one-time setup — env file + dependencies
make setup

# 2. start backing services (Postgres, Redis, Mosquitto, Prometheus, Grafana)
make infra-up

# 3. run the API (terminal 1)
make be-run            # → http://localhost:8080

# 4. run the web app (terminal 2)
make fe-run            # → http://localhost:5173
```

Or run **everything in containers**:

```bash
make up                # api + web + all infra
```

| Service     | URL                                | Notes                       |
| ----------- | ---------------------------------- | --------------------------- |
| Web app     | http://localhost:5173              | dark-theme command center   |
| API         | http://localhost:8080/api/v1/health| Gin, versioned REST         |
| Liveness    | http://localhost:8080/livez        | k8s-style probe             |
| Readiness   | http://localhost:8080/readyz       | dependency rollup           |
| Grafana     | http://localhost:3001              | admin / admin               |
| Prometheus  | http://localhost:9090              |                             |
| MQTT        | localhost:1883 (ws: 9001)          | Mosquitto                   |

Try the MQTT pipeline right now:

```bash
docker exec iot-mosquitto mosquitto_pub \
  -t 'devices/dev-001/telemetry' \
  -m '{"temperature": 42.5, "battery": 87}'
```

## Engineering decisions (Phase 1)

| Decision | Rationale |
| --- | --- |
| **Clean Architecture + DDD (backend)** | `domain` has zero framework imports; Gin lives only in `interfaces/`, GORM will live only in `infrastructure/`. Swappable edges, testable core. |
| **Manual DI at the composition root** | `cmd/api/main.go` wires everything explicitly — no magic container, dependency graph readable in one file. |
| **`slog` for logging** | Structured JSON in prod (Loki-ready), human text in dev, zero third-party dependency. |
| **Probes split: `/livez` vs `/readyz`** | Liveness never touches dependencies (restart signal); readiness fans out concurrent dependency checks (traffic signal). Kubernetes-native from day one. |
| **Feature-based frontend architecture** | Each domain feature owns its UI + queries + state; `shared/` holds the design system. Scales to dozens of features without import spaghetti. |
| **Typed API client + Zod-validated env** | Transport errors are structured (`ApiError` with status/code/request-id); a misconfigured deploy fails at boot, not mid-session. |
| **Strict TS (`exactOptionalPropertyTypes`, `noUncheckedIndexedAccess`)** | The compiler catches the bug classes that hurt most in dashboards: undefined indexing and optional drift. |
| **Distroless, non-root API image** | ~20 MB attack surface, no shell, CVE-resistant. |
| **Explicit CORS allow-list** | Credentials + wildcard is a browser error and a security smell; origins are configured per environment. |

## Development workflow

```bash
make help        # list every command
make check       # vet + tests + lint + typecheck (run before pushing)
make be-test     # Go tests with -race
make fe-test     # Vitest
make infra-logs  # tail infra containers
make down        # stop everything, remove volumes
```

## Roadmap

- [x] **Phase 1 — Foundation**: monorepo, Gin API skeleton (clean architecture), React 19 + Vite app, Docker Compose stack, health probes, CI-ready quality gates
- [x] **Phase 2 — Frontend foundation**: app shell, design system, code-split routing, ⌘K palette
- [x] **Phase 3 — Backend foundation**: GORM + Postgres, repositories, embedded migrations, device CRUD, Redis
- [x] **Phase 4 — Authentication**: JWT + refresh rotation w/ reuse detection, RBAC, login UI, route guards
- [x] **Phase 5 — Real-time**: MQTT consumer, worker-pool ingest pipeline, WS broadcaster, offline sweep, device simulator (`make sim`)
- [x] **Phase 6 — Dashboard**: fleet-summary aggregate endpoint, Recharts time-series/donut/bar, D3 sparklines + radial gauges, streaming device charts (range picker + live append), Analytics page, live throughput
- [ ] Phase 7 — Device management: groups, firmware, OTA simulation, commands
- [ ] Phase 8 — Maps: Leaflet + OpenStreetMap, clustering, geofencing
- [ ] Phase 9 — AI layer: anomaly detection, health scoring, predictive maintenance
- [ ] Phase 10 — Monitoring: Prometheus metrics, Grafana dashboards, Loki, OTel
- [ ] Phase 11 — Docker hardening · Phase 12 — Kubernetes · Phase 13 — Terraform/AWS · Phase 14 — CI/CD

## License

MIT
