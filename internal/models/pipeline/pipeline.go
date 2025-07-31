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
	Id               int         `json:"id"`
	EventName        string      `json:"event_name"`
	Conditions       event.Event `json:"conditions"`
	Query            Query       `json:"query"`
	ExitPipelineName string      `json:"exit_pipeline_name"`
	NextPipelineId   int         `json:"next_pipeline_id"`
	ExecuteDelay     int         `json:"execute_delay"`
	IsActive         bool        `json:"is_active"`
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

type PipelineTemplateDTO struct {
	Id               int            `json:"id"`
	EventName        string         `json:"event_name"`
	Conditions       map[string]any `json:"conditions"`
	Query            map[string]any `json:"query"`
	ExitPipelineName string         `json:"exit_pipeline_name"`
	NextPipelineId   int            `json:"next_pipeline_id"`
	ExecuteDelay     int            `json:"execute_delay"`
	IsActive         bool           `json:"is_active"`
}

func EmptyQuery() Query {
	return Query{
		CohortName: "null",
		Message:    "null",
		Image:      "null",
		Delay:      0,
	}
}
