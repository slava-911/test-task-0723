package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/slava-911/test-task-0723/internal/config"
	"github.com/slava-911/test-task-0723/internal/controller/http/handler"
	"github.com/slava-911/test-task-0723/internal/domain/service"
	"github.com/slava-911/test-task-0723/internal/jwt"
	"github.com/slava-911/test-task-0723/internal/storage"
	"github.com/slava-911/test-task-0723/pkg/cache/freecache"
	"github.com/slava-911/test-task-0723/pkg/logging"
	"github.com/slava-911/test-task-0723/pkg/metric"
	"github.com/slava-911/test-task-0723/pkg/postgresql"
)

func LaunchApp(ctx context.Context, cfg *config.Config) {
	logger := logging.LoggerFromContext(ctx)

	logger.Info("postgresql connection initialization")
	dbDSN := getDatabaseConnectionString(cfg)
	dbClient, err := postgresql.NewClient(ctx, dbDSN, cfg.DB.MaxAttempts, cfg.DB.ConnectionTimeout)
	if err != nil {
		logger.Fatal("failed to connect to database")
	}

	logger.Info("Running PostgreSQL migrations")
	if err = runMigrations(cfg.App.MigrationsPath, dbDSN); err != nil {
		logger.WithError(err).Fatal("failed to run PostgreSQL migrations")
	}

	logger.Info("echo instance initialization")
	e := echo.New()
	loggerConfigurationForEcho(e, logger)

	logger.Println("cache initialization")
	refreshTokenCache := freecache.NewCacheRepo(104857600) // 100MB

	logger.Println("helpers initialization")
	jwtHelper := jwt.NewHelper(refreshTokenCache, logger)
	validateInst := validator.New()

	logger.Info("setup handlers and routes")
	metric.Register(e, cfg.App.Name)

	userStorage := storage.NewUserStorage(dbClient, logger)
	userService := service.NewUserService(userStorage, logger)
	userHandler := handler.NewUserHandler(userService, jwtHelper, validateInst, logger)
	userHandler.Register(e)

	orderStorage := storage.NewOrderStorage(dbClient, logger)
	orderService := service.NewOrderService(orderStorage, logger)
	orderHandler := handler.NewOrderHandler(orderService, logger)
	orderHandler.Register(e)

	productStorage := storage.NewProductStorage(dbClient, logger)
	productService := service.NewProductService(productStorage, logger)
	productHandler := handler.NewProductHandler(productService, logger)
	productHandler.Register(e)

	// Rate Limiter
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodPatch, http.MethodDelete},
		AllowHeaders: []string{"Authorization", "Location", "Charset", "Access-Control-Allow-Origin", "Content-Type",
			"content-type", "Origin", "Accept", "Content-Length", "Accept-Encoding", "X-CSRF-Token"},
		ExposeHeaders:    []string{"Location", "Authorization", "Content-Disposition"},
		AllowCredentials: true,
	}))

	logger.WithFields(map[string]any{
		"IP":   cfg.HTTP.IP,
		"Port": cfg.HTTP.Port,
	}).Info("HTTP Server initializing")

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.HTTP.IP, cfg.HTTP.Port))
	if err != nil {
		logger.WithError(err).Fatal("failed to create listener")
	}

	httpServer := &http.Server{
		Handler:      e,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
	}

	logger.Info("application completely initialized and started")

	go handleGracefulShutdown(ctx, dbClient, httpServer)

	if err = httpServer.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Warn("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}

// getDatabaseConnectionString returns database connection string (DSN, URI, URL) from config
func getDatabaseConnectionString(cfg *config.Config) string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		cfg.DB.Type, cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
}

func runMigrations(migrationsPath, dbDSN string) error {
	m, err := migrate.New(migrationsPath, dbDSN+"?sslmode=disable")
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func loggerConfigurationForEcho(e *echo.Echo, logger *logging.Logger) {
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.WithFields(map[string]any{
					"URI":    v.URI,
					"status": v.Status,
				}).Info("request")
			} else {
				logger.WithFields(map[string]any{
					"URI":    v.URI,
					"status": v.Status,
					"error":  v.Error,
				}).Error("request error")
			}
			return nil
		},
	}))
}

func handleGracefulShutdown(ctx context.Context, db postgresql.Client, httpServer *http.Server) {
	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	logger := logging.LoggerFromContext(ctx)
	signals := []os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)
	sig := <-sigChan
	logger.Infof("Caught signal %s. Shutting down...", sig)

	defer db.Close()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal(err)
	}
}
