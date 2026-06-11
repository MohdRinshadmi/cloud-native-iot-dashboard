// Command api is the entrypoint for the Cloud-Native IoT Analytics backend.
//
// main.go is the COMPOSITION ROOT: the one place allowed to know about every
// layer. It loads config, builds the logger, constructs services and adapters,
// wires them together (manual dependency injection), and runs the server with
// signal-driven graceful shutdown. Everything else depends on abstractions.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ioss/iot-dashboard/backend/internal/application/health"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/config"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/logger"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/server"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/handler"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/router"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "0.1.0-dev"

func main() {
	if err := run(); err != nil {
		// Logger may not exist yet if config failed; use stderr as a last resort.
		slog.Error("fatal", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	startedAt := time.Now()

	// 1. Configuration (fail fast on bad input).
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 2. Structured logging.
	log := logger.New(string(cfg.Env), cfg.LogLevel)
	log.Info("starting iot-api",
		slog.String("version", version),
		slog.String("env", string(cfg.Env)),
	)

	// 3. Application services. (Phase 3+ registers DB/Redis/MQTT checkers.)
	healthSvc := health.New(version, startedAt)

	// 4. Transport wiring.
	healthHandler := handler.NewHealthHandler(healthSvc)
	engine := router.New(router.Deps{
		Config: cfg,
		Logger: log,
		Health: healthHandler,
	})
	srv := server.New(cfg.HTTP, engine, log)

	// 5. Run server in a goroutine; main goroutine waits for a stop signal.
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		// 6. Graceful shutdown with a bounded deadline.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		log.Info("shutdown complete")
		return nil
	}
}
