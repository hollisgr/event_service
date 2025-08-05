package pipeline

import (
	"time"
)

type Query struct {
	CohortName string `json:"cohort_name"`
	Message    string `json:"message"`
	Image      string `json:"image"`
	Delay      int    `json:"delay"`
}

type PipelineTemplate struct {
	Id               int            `json:"id"`
	EventName        string         `json:"event_name"`
	ExecuteType      string         `json:"execute_type"`
	Conditions       map[string]any `json:"conditions"`
	Query            map[string]any `json:"query"`
	ExitPipelineName string         `json:"exit_pipeline_name"`
	NextPipelineId   int            `json:"next_pipeline_id"`
	ExecuteDelay     int            `json:"execute_delay"`
	IsActive         bool           `json:"is_active"`
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

func EmptyQuery() Query {
	return Query{
		CohortName: "null",
		Message:    "null",
		Image:      "null",
		Delay:      0,
	}
}
