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

// NewScheduler creates a new Scheduler instance responsible for scheduling tasks related to pipelines.
// It takes a Storage object for persistent data, Cfg configuration, and a Pipeline interface for pipeline-related operations.
// In case of failure initializing the underlying scheduler, it logs an error and returns nil.
// Otherwise, it constructs and returns a fully configured Scheduler instance.
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

// InitScheduler configures and starts a periodic task in the scheduler.
// It schedules a recurring job that runs at intervals defined by the configuration timeout (in seconds).
// The job executes the PlanningPipelines task, which presumably handles pipeline planning activities.
// Returns an error if the job cannot be successfully added to the scheduler.
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

// StartScheduler initiates the scheduler and starts executing scheduled jobs.
// This function triggers the actual start of the background scheduler process.
// It also logs a confirmation message upon successful startup.
func (s *Scheduler) StartScheduler() {
	s.scheduler.Start()
	log.Println("Scheduler started!")
}

// StopScheduler stops the running scheduler gracefully.
// It shuts down the scheduler, halting all scheduled jobs and preventing future executions.
// Upon stopping, it logs a confirmation message.
func (s *Scheduler) StopScheduler() {
	s.scheduler.Shutdown()
	log.Println("Scheduler stopped!")
}

// NewJobExecutePipeline schedules a one-time job to execute a pipeline after a specified delay.
// It accepts a PipelineDTO describing the pipeline details and a timeout specifying the delay before execution.
// The job is triggered once at the calculated time (current time plus the timeout duration).
// Errors during pipeline execution are logged but do not affect subsequent scheduling.
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

// PlanningPipelines plans and schedules pipeline executions.
// First, it retrieves a list of pipelines requiring action by calling CheckPipelines().
// Then, for each pipeline:
//   - If the pipeline status is "new," marks it as planned and schedules immediate execution.
//   - Otherwise, schedules deferred execution with a configurable timeout.
//
// If no pipelines are available, it logs a message stating so.
func (s *Scheduler) PlanningPipelines() {

	dtoArr, err := s.pipelineService.CheckPipelines()

	if err != nil {
		log.Println("have no available pipelines")
		return
	}

	for _, dto := range dtoArr {
		if dto.Pipeline.Status == "new" {
			s.storage.PipelineSetStatus(context.Background(), dto.Pipeline.Id, "planned")
			log.Println("Planning new pipeline id:", dto.Pipeline.Id, "delay: ", dto.PipelineTemplate.ExecuteDelay)
			s.NewJobExecutePipeline(dto, dto.PipelineTemplate.ExecuteDelay)
		}
		if dto.Pipeline.Status == "send_err" {
			timeout := cfg.GetConfig().Scheduler.SendMessageTimeOutSec
			s.NewJobExecutePipeline(dto, timeout)
			log.Println("Re-planning pipeline id:", dto.Pipeline.Id, "delay: ", timeout)
		}
	}
}
