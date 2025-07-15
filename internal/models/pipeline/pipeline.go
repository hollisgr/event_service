package pipeline

import (
	"event_service/internal/models/event"
	"time"
)

type Query struct {
	CohortName string `json:"cohort_name"`
	Message    string `json:"message"`
	Image      string `json:"image"`
	Delay      int    `json:"delay"`
}

type PipelineTemplate struct {
	Id               int
	EventName        string
	Conditions       event.Event
	Query            Query
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
