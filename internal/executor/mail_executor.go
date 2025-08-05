package executor

import (
	"event_service/internal/dto"
	"event_service/internal/executor/executor_interface"
)

type MailExecutor struct {
}

func NewMailExecutor() executor_interface.Executor {
	e := MailExecutor{}
	return &e
}

func (e *MailExecutor) Send(dto dto.PipelineDTO) error {
	// logic for mail executor
	return nil
}
