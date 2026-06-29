package public

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"test_task/internal/entities"
	"test_task/internal/ports"
	"test_task/tools/config"
	"test_task/tools/metrics"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const basePath = "/api/v1"

type Server struct {
	router  *http.Server
	service ports.Service
	config  config.Config
}

func NewServer(service ports.Service, cfg config.Config) (*Server, error) {
	if service == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "service not set")
	}

	if cfg == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "config not set")
	}

	return &Server{
		router: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.HTTPPort()),
			ReadTimeout:  cfg.HTTPTimeout(),
			WriteTimeout: cfg.HTTPTimeout(),
		},
		service: service,
		config:  cfg,
	}, nil
}

func (s *Server) Start() error {
	s.registerRoutes()

	go func() {
		if err := s.router.ListenAndServe(); err != nil {
			slog.Error("server stopped", "err", err.Error())
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	slog.Info("server stopped")

	if err := s.router.Shutdown(ctx); err != nil {
		slog.Error("server shutdown", "err", err.Error())
		return err
	}

	return nil
}

func (s *Server) registerRoutes() {
	router := gin.New()

	router.Use(metrics.HTTPMetricsMiddleware())

	router.POST(fmt.Sprintf("%s%s", basePath, "/login"), s.Login)
	router.POST(fmt.Sprintf("%s%s", basePath, "/register"), s.Register)

	api := router.Group(basePath)
	api.Use(AuthMiddleware(s.config.TokenSecret()))

	api.GET("/teams", s.GetTeams)
	api.POST("/teams", s.CreateTeam)
	api.GET("/teams/:id", s.GetTeamByID)
	api.POST("/teams/:id/invite", s.AddMember)
	api.GET("/teams/:id/members", s.GetTeamMembers)
	api.POST("/tasks", s.CreateTask)
	api.GET("/tasks/:id", s.GetTaskByID)
	api.PUT("/tasks/:id", s.UpdateTaskByID)
	api.GET("/tasks/:id/history", s.GetHistory)
	api.GET("/teams/:id/tasks", s.GetTasksByTeam)
	api.POST("/tasks/:id/comments", s.AddComment)
	api.GET("/tasks/:id/comments", s.GetCommentsByTask)
	s.router.Handler = router
}
