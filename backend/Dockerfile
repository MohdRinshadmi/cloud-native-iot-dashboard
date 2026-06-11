# =============================================================================
# Go + Gin API — multi-stage, distroless, non-root. Final image ~15-20 MB.
# =============================================================================

# ---- build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache deps separately from source for fast incremental builds.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Static, stripped binary. CGO off => runs on scratch/distroless.
ARG VERSION=0.1.0
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/api ./cmd/api

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/api /api

EXPOSE 8080
USER nonroot:nonroot

# Distroless has no shell; the orchestrator probes /livez & /readyz over HTTP.
ENTRYPOINT ["/api"]
