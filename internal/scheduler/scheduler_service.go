package scheduler

import (
	"event_service/internal/cfg"
	"event_service/internal/db/storage"
	"event_service/internal/pipeline/pipeline_interface"
	"event_service/internal/scheduler/scheduler_interface"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type Scheduler struct {
	scheduler       gocron.Scheduler
	storage         storage.Storage
	pipelineService pipeline_interface.Pipeline
	cfg             *cfg.Cfg
}

func NewScheduler(storage storage.Storage, cfg *cfg.Cfg, p pipeline_interface.Pipeline) scheduler_interface.Scheduler {
	s, err := gocron.NewScheduler()

	if err != nil {
		log.Println("cant create new scheduler")
		return nil
	}

	return &Scheduler{
		scheduler:       s,
		storage:         storage,
		cfg:             cfg,
		pipelineService: p,
	}
}

func (s *Scheduler) InitScheduler() error {
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(
			time.Duration(s.cfg.Scheduler.CheckEventsTimeOutSec)*time.Second,
		),
		gocron.NewTask(s.pipelineService.CheckPipelines),
	)
	if err != nil {
		log.Println("cant create new job")
		return err
	}
	return nil
}

func (s *Scheduler) StartScheduler() {
	s.scheduler.Start()
	log.Println("Scheduler started!")
}

func (s *Scheduler) StopScheduler() {
	s.scheduler.Shutdown()
	log.Println("Scheduler stopped!")
}

func (s *Scheduler) PlanningPipeline(pipelineId int, delay int) error {
	log.Println("Planning pipeline id:", pipelineId)

	_, err := s.scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(time.Duration(delay)*time.Second)),
		),
		gocron.NewTask(
			func() {
				err := s.pipelineService.ExecutePipeline(pipelineId)
				if err != nil {
					log.Println(err)
					return
				}
			},
		),
	)
	if err != nil {
		return err
	}
	return nil
}
