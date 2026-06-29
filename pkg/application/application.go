package application

import (
	"test_task/internal/adapters/storage/ms"
	"test_task/internal/cases"
	"test_task/internal/ports"
	"test_task/internal/ports/http/public"
	baseApp "test_task/tools/application"
	"test_task/tools/config"
	"test_task/tools/port"
)

type App struct {
	cfg     config.Config
	storage cases.Repository
	service ports.Service

	server port.Port
}

func New(cfg config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (app *App) Run() {
	app.initStore()
	app.initService()
	app.initServer()

	baseApp := baseApp.NewBuilder(app.cfg).WithLogger().WithMetricsPort().AddStarters(app.server.Start).AddWaiters(app.server.Stop, app.storage.Stop).Build()

	if err := baseApp.Run(); err != nil {
		panic(err)
	}
}

func (app *App) initStore() {
	st, err := ms.NewStorage(app.cfg)
	if err != nil {
		panic(err)
	}
	app.storage = st
}

func (app *App) initService() {
	serv, err := cases.NewService(app.storage, app.cfg)
	if err != nil {
		panic(err)
	}

	app.service = serv
}

func (app *App) initServer() {
	srv, err := public.NewServer(app.service, app.cfg)
	if err != nil {
		panic(err)
	}

	app.server = srv
}
