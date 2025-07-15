package pipeline_interface

type Pipeline interface {
	InitPipelineTemplates()
	CheckPipelines() error
	ExecutePipeline(pipelineId int) error
}
