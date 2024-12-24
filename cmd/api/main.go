package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"queue-bite/internal/config"
	"queue-bite/internal/config/logger"
	hdimpl "queue-bite/internal/features/hostdesk/repository"
	hd "queue-bite/internal/features/hostdesk/service"
	sm "queue-bite/internal/features/seatmanager/service"
	st "queue-bite/internal/features/servicetime/service"
	wimpl "queue-bite/internal/features/waitlist/repository/redis"
	"queue-bite/internal/platform"
	eb "queue-bite/internal/platform/eventbus"
	ebimpl "queue-bite/internal/platform/eventbus/redis"
	"queue-bite/internal/server"
	_ "queue-bite/pkg/env/autoload"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.NewConfig(getenv)
	if err != nil {
		return err
	}

	logger := log.NewZerologLogger(stdout, cfg.Dev)
	redis := platform.NewRedis(cfg, logger)
	eventRegistry := eb.NewEventRegistry()
	eventbus := ebimpl.NewRedisEventBus(logger, redis.Client, eventRegistry)
	instantHost := hd.NewInstantServeHostDesk(logger,
		cfg.HostDesk.InstantServeHostDeskSeatCapacity,
		hdimpl.NewRedisHostDeskRepository(logger, redis.Client),
		// hdimpl.NewInMemoryHostDeskRepository(logger),
		eventbus,
		hd.NewLinearServiceTimer(logger, cfg.HostDesk.LinearServiceTimerDurationPerGuest))

	server := server.NewServer(
		cfg,
		logger,
		redis,
		eventRegistry,
		eventbus,
		st.NewFixedRateEstimator(cfg.ServiceEstimator.FixedRateUnit),
		wimpl.NewRedisWaitlistRepository(logger, redis.Client, cfg.Waitlist.EntityTTL, cfg.Waitlist.ScanChunkSize),
		instantHost,
		sm.NewFairOrderStrategy(),
		sm.NewOrderedSeatingStrategy,
	)
	serverError := make(chan error, 1)

	go func() {
		logger.LogInfo(log.Server, "starting server", "addr", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogErr(log.Server, err, "could not start server")
			serverError <- fmt.Errorf("could not start server: %w", err)
			return
		}

		serverError <- nil
	}()

	select {
	case err := <-serverError:
		return err
	case <-ctx.Done():
		logger.LogInfo(log.Server, "shutting down server gracefully, press <C-c> again to force")
		// seatManager.UnwatchSeatVacancy(ctx)
		sctx, stop := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer stop()

		if err := server.Shutdown(sctx); err != nil {
			logger.LogErr(log.Server, err, "server forced to shutdown")
			return fmt.Errorf("server forced to shutdown with error: %w", err)
		}
		logger.LogInfo(log.Server, "server exiting")
		return nil
	}
}
