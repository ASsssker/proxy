package requester

import (
	"context"
	"log/slog"
	"os"

	"github.com/ASsssker/proxy/internal/config"
	"github.com/ASsssker/proxy/internal/mq"
	"github.com/ASsssker/proxy/internal/services"
	"github.com/ASsssker/proxy/internal/storage/postgres"
)

var (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type RequesterApp struct {
	log     *slog.Logger
	service *services.RequesterService
}

func MustNewRequester(ctx context.Context, cfg config.Config) *RequesterApp {
	log := setupLogger(cfg.Env)
	log.InfoContext(ctx, "starting requester service", slog.String("env", cfg.Env))

	taskUpdater, err := postgres.NewPostgresDB(ctx, log, cfg.PostgresDNS())
	if err != nil {
		panic(err)
	}
	log.InfoContext(ctx, "successful connection to the database")

	msgReceiver, err := mq.NewNatsMQ(cfg, log)
	if err != nil {
		panic(err)
	}
	log.InfoContext(ctx, "successful connection to the mq")

	taskExecutor := services.NewRequestExecutor(cfg, log)
	service := services.NewRequesterService(log, cfg, taskUpdater, msgReceiver, taskExecutor)

	return &RequesterApp{
		log:     log,
		service: service,
	}
}

func (r RequesterApp) MustRun(ctx context.Context) {
	if err := r.service.Run(ctx); err != nil {
		r.log.ErrorContext(ctx, "failed to run requester app", slog.String("error", err.Error()))
		panic(err)
	}
}

func (r RequesterApp) Stop(ctx context.Context) {
	if err := r.service.Close(ctx); err != nil {
		r.log.ErrorContext(ctx, "failed to stop requester app", slog.String("error", err.Error()))
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
