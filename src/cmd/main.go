package main

import (
	"context"

	"github.com/slava-911/test-task-0723/internal/app"
	"github.com/slava-911/test-task-0723/internal/config"
	"github.com/slava-911/test-task-0723/pkg/logging"
)

func main() {
	cfg := config.GetConfig()
	logger := logging.NewLogger(cfg.App.LogLevel)
	ctx := logging.ContextWithLogger(context.Background(), logger)
	app.LaunchApp(ctx, cfg)
}
