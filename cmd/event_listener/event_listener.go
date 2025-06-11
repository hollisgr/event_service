package main

import (
	"context"
	"event_service/internal/cfg"
	"event_service/internal/db"
	"event_service/internal/http/service"
	"event_service/pkg/logger"
	"event_service/pkg/postgres"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	logger := logger.GetLogger()

	cfg := cfg.GetConfig()

	r := gin.Default()

	validate := validator.New(validator.WithRequiredStructEnabled())

	pgxPool, err := postgres.NewClient(context.Background(), 3, *cfg)

	if err != nil {
		logger.Fatalln(err)
	}

	err = pgxPool.Ping(context.Background())

	if err != nil {
		logger.Fatalln(err)
	}

	storage := db.NewStorage(pgxPool, logger)

	handler := service.NewHandler(storage, cfg, validate)

	handler.Register(r)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nInterrupt signal received. Exiting...")
		pgxPool.Close()
		os.Exit(0)
	}()

	r.Run(cfg.Listen.BindIP + ":" + cfg.Listen.Port)
}
