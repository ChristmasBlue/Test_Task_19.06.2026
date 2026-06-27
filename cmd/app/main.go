package main

import (
	"test_task/internal/adapters/config"
	"test_task/pkg/application"
)

var (
	configPath = ""
)

func main() {
	cfg := config.NewConfig(configPath)

	app := application.New(cfg)
	app.Run()
}
