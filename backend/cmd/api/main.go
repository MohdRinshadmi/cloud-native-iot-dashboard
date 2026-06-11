// Command api is the entrypoint for the Cloud-Native IoT Analytics backend.
//
// main.go is the COMPOSITION ROOT: the one place allowed to know about every
// layer. It loads config, connects infrastructure (Postgres, Redis), runs
// migrations, constructs services and adapters, wires them together (manual
// dependency injection), and runs the server with graceful shutdown.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	appauth "github.com/ioss/iot-dashboard/backend/internal/application/auth"
	appdevices "github.com/ioss/iot-dashboard/backend/internal/application/devices"
	"github.com/ioss/iot-dashboard/backend/internal/application/health"
	"github.com/ioss/iot-dashboard/backend/internal/application/ingest"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/authtoken"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/cache"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/config"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/hash"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/logger"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/mqtt"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/persistence"
	"github.com/ioss/iot-dashboard/backend/internal/infrastructure/server"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/handler"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/http/router"
	"github.com/ioss/iot-dashboard/backend/internal/interfaces/ws"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "0.1.0-dev"

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	startedAt := time.Now()
	ctx := context.Background()

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

	// 3. Infrastructure: Postgres (+ migrations) and Redis.
	db, err := persistence.Connect(cfg.Postgres, log)
	if err != nil {
		return err
	}
	if err := persistence.Migrate(db, log); err != nil {
		return err
	}

	redisClient, err := cache.New(ctx, cfg.Redis)
	if err != nil {
		return err
	}
	defer func() { _ = redisClient.Close() }()

	// 4. Adapters.
	deviceRepo := persistence.NewDeviceRepository(db)
	userRepo := persistence.NewUserRepository(db)
	tenantRepo := persistence.NewTenantRepository(db)
	refreshRepo := persistence.NewRefreshTokenRepository(db)
	txManager := persistence.NewTxManager(db)
	hasher := hash.NewBcryptHasher()
	jwtManager := authtoken.NewManager(cfg.JWT.AccessSecret, cfg.JWT.AccessTTL, time.Now)
	limiter := cache.NewRateLimiter(redisClient)

	// 5. Application services.
	healthSvc := health.New(version, startedAt)
	healthSvc.Register(persistence.NewHealthChecker(db))
	healthSvc.Register(cache.NewHealthChecker(redisClient))

	deviceSvc := appdevices.NewService(deviceRepo, time.Now)
	authSvc := appauth.NewService(
		userRepo, tenantRepo, refreshRepo, txManager,
		hasher, jwtManager, cfg.JWT.RefreshTTL, time.Now,
	)

	// 6. Dev convenience: demo tenant, users and fleet (never in prod).
	if cfg.Env == config.EnvDevelopment {
		if err := persistence.SeedDev(ctx, db, hasher, log); err != nil {
			return err
		}
	}

	// 7. Real-time pipeline: WS hub ← ingest workers ← MQTT consumer.
	// runCtx cancellation tears the whole pipeline down in reverse order.
	runCtx, cancelRun := context.WithCancel(ctx)
	defer cancelRun()

	hub := ws.NewHub(log)
	go hub.Run(runCtx)

	telemetryRepo := persistence.NewTelemetryRepository(db)
	latestStore := cache.NewLatestStore(redisClient)
	ingestSvc := ingest.NewService(
		ingest.Config{
			Workers:           cfg.Ingest.Workers,
			QueueSize:         cfg.Ingest.QueueSize,
			HeartbeatInterval: cfg.Ingest.HeartbeatInterval,
			OfflineAfter:      cfg.Ingest.OfflineAfter,
		},
		telemetryRepo, latestStore, deviceRepo, hub, log, func() time.Time { return time.Now().UTC() },
	)
	ingestSvc.Start(runCtx)

	consumer := mqtt.NewConsumer(cfg.MQTT, ingestSvc, log)
	if err := consumer.Start(); err != nil {
		return err
	}

	// 8. Transport wiring.
	engine := router.New(router.Deps{
		Config:    cfg,
		Logger:    log,
		Health:    handler.NewHealthHandler(healthSvc),
		Auth:      handler.NewAuthHandler(authSvc, cfg.IsProduction()),
		Devices:   handler.NewDeviceHandler(deviceSvc),
		Telemetry: handler.NewTelemetryHandler(deviceSvc, telemetryRepo, latestStore),
		WS:        ws.NewHandler(hub, jwtManager, cfg.HTTP.AllowedOrigins, log),
		Verifier:  jwtManager,
		Limiter:   limiter,
	})
	srv := server.New(cfg.HTTP, engine, log)

	// 9. Run server; block until SIGINT/SIGTERM; drain gracefully in order:
	// stop intake (MQTT) → drain workers → drain HTTP/WS.
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()

	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-signalCtx.Done():
		consumer.Stop() // 1. no new messages
		cancelRun()     // 2. workers + hub wind down
		ingestSvc.Wait()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil { // 3. drain HTTP
			return err
		}
		processed, dropped := ingestSvc.Stats()
		log.Info("shutdown complete",
			slog.Int64("messages_processed", processed),
			slog.Int64("messages_dropped", dropped),
		)
		return nil
	}
}
