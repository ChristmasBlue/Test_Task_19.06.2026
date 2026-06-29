package main

import (
	"test_task/internal/adapters/config"
	"test_task/pkg/application"
)

var (
	configPath = "D:\\go_projects\\Test_task3\\configs\\config.yaml"
)

func main() {
	cfg := config.NewConfig(configPath)

	app := application.New(cfg)
	app.Run()
}
