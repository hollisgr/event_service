package dto

import "event_service/internal/models/pipeline"

type PipelineDTO struct {
	Pipeline         pipeline.Pipeline
	PipelineTemplate pipeline.PipelineTemplate
}
