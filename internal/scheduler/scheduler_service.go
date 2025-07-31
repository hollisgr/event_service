package scheduler

import (
	"context"
	"event_service/internal/cfg"
	"event_service/internal/db/storage"
	"event_service/internal/dto"
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
		gocron.NewTask(s.PlanningPipelines),
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

func (s *Scheduler) NewJobExecutePipeline(dto dto.PipelineDTO, timeout int) {
	s.scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(time.Duration(timeout)*time.Second)),
		),
		gocron.NewTask(
			func() {
				err := s.pipelineService.ExecutePipeline(dto)
				if err != nil {
					log.Println(err)
					return
				}
			},
		),
	)
}

func (s *Scheduler) PlanningPipelines() {

	dtoArr, err := s.pipelineService.CheckPipelines()

	if err != nil {
		log.Println("have no available pipelines")
		return
	}

	for _, dto := range dtoArr {
		if dto.Pipeline.Status == "new" {
			s.storage.PipelineSetStatus(context.Background(), dto.Pipeline.Id, "planned")
			log.Println("Planning pipeline id:", dto.Pipeline.Id)
			s.NewJobExecutePipeline(dto, dto.PipelineTemplate.ExecuteDelay)
		} else {
			s.NewJobExecutePipeline(dto, cfg.GetConfig().Scheduler.SendMessageTimeOutSec)
		}
	}
}
