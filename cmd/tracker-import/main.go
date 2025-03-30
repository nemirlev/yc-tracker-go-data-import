package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nemirlev/yc-tracker-go-data-import/internal/config"
	"github.com/nemirlev/yc-tracker-go-data-import/internal/repository"
	"github.com/nemirlev/yc-tracker-go-data-import/internal/service"
	"github.com/nemirlev/yc-tracker-go-data-import/pkg/database"
	"github.com/nemirlev/yc-tracker-go-data-import/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Set up logging
	logger.SetupLogging(cfg.App.LogLevel)

	// Connect to database
	db, err := database.Connect(cfg.GetDSN())
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run migrations
	migrationsPath := "migrations"
	if err := database.MigrateDB(cfg.GetMigrateDSN(), migrationsPath); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Create repository
	repositoryService := repository.NewService(db)

	// Create main service
	svc := service.NewService(cfg, repositoryService)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run sync operation
	if err := svc.Sync(ctx); err != nil {
		slog.Error("Failed to sync data", "error", err)
		os.Exit(1)
	}

	slog.Info("Application completed successfully")
}

// Handler for Yandex Cloud Function
func Handler(ctx context.Context) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return err
	}

	// Set up logging
	logger.SetupLogging(cfg.App.LogLevel)

	// Connect to database
	db, err := database.Connect(cfg.GetDSN())
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return err
	}
	defer db.Close()

	// Run migrations
	migrationsPath := "migrations"
	if err := database.MigrateDB(cfg.GetMigrateDSN(), migrationsPath); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		return err
	}

	// Create repository
	repositoryService := repository.NewService(db)

	// Create main service
	svc := service.NewService(cfg, repositoryService)

	// Run sync operation
	if err := svc.Sync(ctx); err != nil {
		slog.Error("Failed to sync data", "error", err)
		return err
	}

	slog.Info("Application completed successfully")
	return nil
}
