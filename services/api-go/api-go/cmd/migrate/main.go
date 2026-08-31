package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hcl-backend/services/api-go/internal/platform/config"
	"github.com/hcl-backend/services/api-go/internal/platform/database"
	"github.com/hcl-backend/services/api-go/internal/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(cfg.App.LogLevel)
	log.Info("Starting Migrations", "env", cfg.App.Env)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize Database connection to test credentials before migrating
	dbPool, err := database.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database for migrations", "error", err)
	}
	dbPool.Close()

	log.Info("Database connection successful. Initializing migration runner...")

	// Default to project migrations directory
	sourceURL := "file://migrations"

	m, err := migrate.New(sourceURL, cfg.Database.DSN())
	if err != nil {
		log.Fatal("Failed to initialize migration runner", "error", err)
	}

	// Parse command line args
	flag.Parse()
	args := flag.Args()
	cmd := "up"
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "up":
		log.Info("Running migrations up...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to run up migrations", "error", err)
		}
		log.Info("Up migrations complete.")
	case "down":
		log.Info("Running migrations down...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to run down migrations", "error", err)
		}
		log.Info("Down migrations complete.")
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				log.Info("No migrations applied yet.")
			} else {
				log.Fatal("Failed to get migration version", "error", err)
			}
		} else {
			log.Info("Current migration version", "version", version, "dirty", dirty)
		}
	default:
		fmt.Println("Usage: migrate [up|down|version]")
		os.Exit(1)
	}
}
