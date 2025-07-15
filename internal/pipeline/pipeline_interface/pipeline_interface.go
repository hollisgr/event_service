package pipeline_interface

import "event_service/internal/dto"

type Pipeline interface {
	InitPipelineTemplates()
	CheckPipelines() ([]dto.PipelineDTO, error)
	ExecutePipeline(dto.PipelineDTO) error
}
