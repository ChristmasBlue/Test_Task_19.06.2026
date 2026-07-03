package main

// @title           Task Manager API
// @version         1.0
// @description     API для управления задачами в командах
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in              header
// @name            Authorization

import (
	"os"
	"test_task/internal/adapters/config"
	"test_task/pkg/application"
)

func main() {

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	cfg := config.NewConfig(configPath)

	app := application.New(cfg)
	app.Run()
}
