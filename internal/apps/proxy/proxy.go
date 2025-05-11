package proxy

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/ASsssker/proxy/internal/config"
	"github.com/ASsssker/proxy/internal/mq"
	v1 "github.com/ASsssker/proxy/internal/rest/v1"
	"github.com/ASsssker/proxy/internal/services"
	"github.com/ASsssker/proxy/internal/storage/postgres"
	"github.com/ASsssker/proxy/internal/validation"
	"github.com/gin-gonic/gin"
)

var (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type ProxyApp struct {
	log     *slog.Logger
	service *services.ProxyService
	srv     *http.Server
}

func MustNewProxyApp(ctx context.Context, cfg config.Config) ProxyApp {
	log := setupLogger(cfg.Env)
	log.InfoContext(ctx, "starting proxy service", slog.String("env", cfg.Env))
	log.DebugContext(ctx, "credential proxy service", slog.String("host", cfg.ProxyHost), slog.String("port", cfg.ProxyPort))

	taskProvider, err := postgres.NewPostgresDB(ctx, log, cfg.PostgresDNS())
	if err != nil {
		panic(err)
	}
	log.InfoContext(ctx, "successful connection to the database")

	msgSender, err := mq.NewRabbitMQ(cfg, log)
	if err != nil {
		panic(err)
	}
	log.InfoContext(ctx, "successful connection to the mq")

	validator, err := validation.NewValidator()
	if err != nil {
		panic(err)
	}

	service := services.NewProxyService(log, taskProvider, msgSender, validator)

	if cfg.Env == envProd {
		gin.SetMode(gin.ReleaseMode)
	}

	handler := gin.Default()
	v1.Register(handler, log, service)

	srv := http.Server{
		Handler:      handler,
		Addr:         net.JoinHostPort(cfg.ProxyHost, cfg.ProxyPort),
		ReadTimeout:  cfg.ProxyHTTPReadTimeout,
		WriteTimeout: cfg.ProxyHTTPWriteTimeout,
		IdleTimeout:  cfg.ProxyHTTPIdleTimeout,
	}

	return ProxyApp{
		log:     log,
		service: service,
		srv:     &srv,
	}
}

func (p ProxyApp) MustRun(ctx context.Context) {
	if err := p.srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			p.log.ErrorContext(ctx, "failed to run http server", slog.String("error", err.Error()))
			panic(err)
		}
	}
}

func (p ProxyApp) Stop(ctx context.Context) {
	p.log.InfoContext(ctx, "start stopping server")
	if err := p.srv.Shutdown(ctx); err != nil {
		p.log.ErrorContext(ctx, "failed to stopping server", slog.String("error", err.Error()))
	}

	if err := p.service.Close(ctx); err != nil {
		p.log.ErrorContext(ctx, "failed to stopping service", slog.String("error", err.Error()))
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
