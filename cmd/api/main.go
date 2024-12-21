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
	hostdesk_repo "queue-bite/internal/features/hostdesk/repository"
	hostdesk "queue-bite/internal/features/hostdesk/service"
	seatmanager "queue-bite/internal/features/seatmanager/service"
	servicetime "queue-bite/internal/features/servicetime/service"
	"queue-bite/internal/features/sse"
	waitlist_redis "queue-bite/internal/features/waitlist/repository/redis"
	waitlist "queue-bite/internal/features/waitlist/service"
	"queue-bite/internal/platform"
	eb "queue-bite/internal/platform/eventbus"
	eb_redis "queue-bite/internal/platform/eventbus/redis"
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
	eventbus := eb_redis.NewRedisEventBus(logger, redis.Client, eventRegistry)

	waitlist := waitlist.NewWaitlistService(
		logger,
		waitlist_redis.NewRedisWaitlistRepository(logger, redis.Client, cfg.Waitlist.EntityTTL, cfg.Waitlist.ScanChunkSize),
		servicetime.NewFixedRateEstimator(cfg.ServiceEstimator.FixedRateUnit),
		eventbus,
	)

	instantHost := hostdesk.NewInstantServeHostDesk(
		logger,
		cfg.HostDesk.InstantServeHostDeskSeatCapacity,
		hostdesk_repo.NewInMemoryHostDeskRepository(logger),
		eventbus,
	)

	seatManager := seatmanager.NewSeatManager(
		logger,
		eventbus,
		waitlist,
		instantHost,
		seatmanager.NewOrderedSeatingStrategy(waitlist),
	)

	sseManager := sse.NewServerSentEvent(logger, eventbus)

	server := server.NewServer(
		cfg,
		logger,
		redis,
		eventRegistry,
		eventbus,
		waitlist,
		instantHost,
		seatManager,
		sseManager,
	)
	seatManagerError := make(chan error, 1)
	serverError := make(chan error, 1)

	go func() {
		if err := seatManager.WatchSeatVacancy(ctx); err != nil {
			logger.LogErr(log.Server, err, "failed to start seat manager")
			seatManagerError <- fmt.Errorf("failed to start seat manager: %w", err)
			return
		}

		seatManagerError <- nil
	}()

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
	case err := <-seatManagerError:
		return err
	case err := <-serverError:
		seatManager.UnwatchSeatVacancy(ctx)
		return err
	case <-ctx.Done():
		logger.LogInfo(log.Server, "shutting down server gracefully, press <C-c> again to force")

		seatManager.UnwatchSeatVacancy(ctx)
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
