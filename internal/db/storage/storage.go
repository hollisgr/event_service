package storage

import (
	"context"
	"event_service/internal/models/event"
	"event_service/internal/models/pipeline"
)

type Storage interface {
	EventSave(ctx context.Context, e event.Event) (int, error)
	EventsLoadNew(ctx context.Context) ([]event.Event, error)
	EventLoad(ctx context.Context, event_id int) (event.Event, error)
	EventSetStatus(ctx context.Context, event_id int, status string) error

	PipelineSave(ctx context.Context, p pipeline.Pipeline) (int, error)
	PipelineLoad(ctx context.Context, pipeline_id int) (pipeline.Pipeline, error)
	PipelinesLoad(ctx context.Context) ([]pipeline.Pipeline, error)
	PipelineSetStatus(ctx context.Context, pipeline_id int, status string) error
	PipelineIncreaseSendCounter(ctx context.Context, pipeline_id int) error

	PipelineTemplateSave(ctx context.Context, data pipeline.PipelineTemplateDTO) (int, error)
	PipelineTemplatesLoad(ctx context.Context) ([]pipeline.PipelineTemplateDTO, error)
	PipelineTemplateLoad(ctx context.Context, templateId int) (pipeline.PipelineTemplate, error)
}
