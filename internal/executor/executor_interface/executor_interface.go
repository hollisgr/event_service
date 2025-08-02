package executor_interface

import "event_service/internal/dto"

type Executor interface {
	Send(dto dto.PipelineDTO) error
}
