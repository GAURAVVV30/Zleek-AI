package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hcl-backend/services/api-go/internal/platform/cache"
	"github.com/hcl-backend/services/api-go/internal/platform/config"
	"github.com/hcl-backend/services/api-go/internal/platform/database"
	"github.com/hcl-backend/services/api-go/internal/platform/events"
	"github.com/hcl-backend/services/api-go/internal/platform/logger"

	notifApp "github.com/hcl-backend/services/api-go/internal/notifications/application"
	notifInfra "github.com/hcl-backend/services/api-go/internal/notifications/infrastructure"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(cfg.App.LogLevel)
	log.Info("Starting Worker service", "env", cfg.App.Env)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize Database
	dbPool, err := database.NewPool(ctx, cfg.Database.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer dbPool.Close()

	// Initialize Redis Cache
	redisClient, err := cache.NewClient(ctx, cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatal("Failed to connect to redis", "error", err)
	}
	defer redisClient.Close()

	// Initialize Notifications Module
	notificationsRepo := notifInfra.NewPostgresRepository(dbPool)
	notificationsEventHandler := notifApp.NewEventHandler(notificationsRepo)

	// Initialize Event Bus
	redisBus := events.NewRedisBus(redisClient)
	notificationsEventHandler.Register(redisBus)

	log.Info("Worker ready to process background jobs")

	// Start listening for domain events
	go func() {
		eventTypes := []string{
			events.EventTypeCompetencyUpdated,
			events.EventTypeConceptWeak,
			events.EventTypeGoalAchieved,
			events.EventTypeResourceFlagged,
		}
		if err := redisBus.StartListening(ctx, eventTypes); err != nil {
			log.Error("Event bus listener stopped", "error", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down worker...")
	log.Info("Worker stopped")
}
