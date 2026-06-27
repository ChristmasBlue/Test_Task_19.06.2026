package application

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"test_task/tools/config"
	"test_task/tools/logger"
	"test_task/tools/metrics"
	"time"
)

type BaseApplication struct {
	starters        []func() error
	waiters         []func(ctx context.Context) error
	shutdownTimeout time.Duration
}

type Builder struct {
	cfg config.Config
	app *BaseApplication
}

func NewBuilder(cfg config.Config) *Builder {
	return &Builder{
		cfg: cfg,
		app: &BaseApplication{},
	}
}

func (b *Builder) WithLogger() *Builder {
	logger.InitLogger(b.cfg)

	return b
}

func (b *Builder) WithMetricsPort() *Builder {
	server, err := metrics.NewServer(b.cfg.MetricsPort(), b.cfg.MetricsTimeout())
	if err != nil {
		panic(err)
	}

	b.app.starters = append(b.app.starters, server.Start)
	b.app.waiters = append(b.app.waiters, server.Stop)

	return b
}

func (b *Builder) AddStarters(starters ...func() error) *Builder {
	b.app.starters = append(b.app.starters, starters...)

	return b
}

func (b *Builder) AddWaiters(waiters ...func(context.Context) error) *Builder {
	b.app.waiters = append(b.app.waiters, waiters...)

	return b
}

func (b *Builder) Build() *BaseApplication {
	b.app.shutdownTimeout = b.cfg.GracefullShutdownTimeout()
	return b.app
}

func (app *BaseApplication) Start() error {
	for _, starter := range app.starters {
		if err := starter(); err != nil {
			slog.Error("shutdown error", "err", err)
			return err
		}
	}

	return nil
}

func (app *BaseApplication) Stop(ctx context.Context) error {
	var lastErr error
	for _, waiter := range app.waiters {
		if err := waiter(ctx); err != nil {
			slog.Error("shutdown error", "err", err)
			lastErr = err
		}
	}
	return lastErr
}

func (app *BaseApplication) Run() error {
	if err := app.Start(); err != nil {
		return err
	}
	slog.Info("application started")

	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)
	<-chSig
	slog.Info("application sign stopped")

	ctx, cancel := context.WithTimeout(context.Background(), app.shutdownTimeout)
	defer cancel()

	return app.Stop(ctx)
}
