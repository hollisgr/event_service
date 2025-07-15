package pipeline

import (
	"event_service/internal/models/event"
	"time"
)

type PipelineTemplate struct {
	Id         int
	EventName  string
	Conditions event.Event
	Query      struct {
		CohortName string
		Message    string
		Image      string
		Delay      int
	}
	ExitPipelineName string
	NextPipelineId   int
	ExecuteDelay     int
	IsActive         bool
}

type Pipeline struct {
	Id             int
	ParentId       int
	EventId        int
	UserId         string
	TemplateId     int
	Status         string
	SendingCounter int
	CreatedAt      time.Time
}
