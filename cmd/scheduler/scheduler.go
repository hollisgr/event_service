package main

import (
	"context"
	"event_service/internal/cfg"
	"event_service/internal/db"
	"event_service/internal/pipeline"
	"event_service/internal/scheduler"
	"event_service/pkg/postgres"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := cfg.GetConfig()

	pgxPool, err := postgres.NewClient(context.Background(), 3, *cfg)

	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	err = pgxPool.Ping(context.Background())

	if err != nil {
		log.Fatalln(err)
	}

	storage := db.NewStorage(pgxPool)

	pipelineService := pipeline.NewPipelineService(storage, *cfg)

	s := scheduler.NewScheduler(storage, cfg, pipelineService)

	err = s.InitScheduler()

	if err != nil {
		log.Fatal(err)
	}

	s.StartScheduler()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nInterrupt signal received. Exiting...")
		pgxPool.Close()
		os.Exit(0)
	}()

	select {}
}
