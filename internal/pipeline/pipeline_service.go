package pipeline

import (
	"context"
	"encoding/json"
	"event_service/internal/cfg"
	"event_service/internal/db/storage"
	"event_service/internal/dto"
	"event_service/internal/factory"
	"event_service/internal/models/event"
	"event_service/internal/models/pipeline"
	"event_service/internal/pipeline/pipeline_interface"
	"fmt"
	"log"
	"reflect"
)

type PipelineService struct {
	pipelineTemplates []pipeline.PipelineTemplate // Array of pipeline templates
	storage           storage.Storage             // Interface for data storage
	cfg               cfg.Cfg                     // Configuration settings
}

// Creates a new instance of the PipelineService object and initializes its internal structures.
func NewPipelineService(s storage.Storage, cfg cfg.Cfg) pipeline_interface.Pipeline {
	var p PipelineService
	p.storage = s
	p.cfg = cfg
	p.InitPipelineTemplates()
	return &p
}

// CheckPipelines verifies the state of existing pipelines and processes new events.
// It retrieves new events from storage, checks conditions for creating new pipelines,
// generates appropriate pipeline records, and updates current operations' statuses.
// Returns a list of processed pipelines and possible error.
func (p *PipelineService) CheckPipelines() ([]dto.PipelineDTO, error) {
	log.Println("pipeline check begins...")
	counter := 0
	dtoArr := make([]dto.PipelineDTO, 0)
	events, err := p.storage.EventsLoadNew(context.Background())
	if err != nil {
		return nil, err
	}

	for _, event := range events {
		template, ok := p.CheckConditions(event)
		if ok {
			counter++
			p.NewPipeline(event, template, &dtoArr)
		}
		p.CheckCancelState(event)
	}
	p.CheckSended(&dtoArr)
	if counter == 0 {
		log.Println("have no new pipelines!")
	}
	return dtoArr, nil
}

// CheckCancelState examines whether any active pipelines should be cancelled based on incoming events.
// It iterates over all stored pipelines, comparing their user IDs with the event's user ID.
// If a matching pipeline exists and its exit condition matches the event type, the pipeline's status is set to 'cancelled'.
// No value is returned since it's intended purely for side effects.
func (p *PipelineService) CheckCancelState(e event.Event) {
	pArr, _ := p.storage.PipelinesLoad(context.Background())
	for _, pipeline := range pArr {
		if pipeline.UserId == e.UserId {
			t, _ := p.storage.PipelineTemplateLoad(context.Background(), pipeline.TemplateId)
			if t.ExitPipelineName == e.EventType {
				p.storage.PipelineSetStatus(context.Background(), pipeline.Id, "cancelled")
				log.Println("cancel pipeline with id: ", pipeline.Id)
			}
		}
	}
}

// CheckConditions determines whether an event meets the conditions for starting a pipeline.
// It compares the event type with available pipeline templates and checks if the event satisfies the specified conditions.
// Returns a matched pipeline template along with a boolean indicating success.
// If no suitable template is found, an empty template and false are returned.
// Additionally, sets the event status to "checked" after processing.
func (p *PipelineService) CheckConditions(e event.Event) (pipeline.PipelineTemplate, bool) {
	template := pipeline.PipelineTemplate{}
	log.Printf("checking event: %d", e.EventId)
	ok := false

	for _, t := range p.pipelineTemplates {
		if e.EventType == t.EventName {
			e_map := p.EventToMap(e)
			ok = p.CompareConditions(t.Conditions, e_map)
			log.Println("name ok", e.EventId)
		}
		if ok {
			template = t
			break
		}
	}
	p.storage.EventSetStatus(context.Background(), e.EventId, "checked")
	log.Printf("event %d status updated", e.EventId)
	return template, ok
}

// CheckPlanned checks for planned pipelines in the storage and appends them to the provided slice of PipelineDTO.
// It logs the beginning of the check and counts the number of planned pipelines with a SendingCounter less than 3.
// If no planned pipelines are found, it logs a message indicating that there are no planned pipelines.
//
// Parameters:
//   - dtoArr: A pointer to a slice of PipelineDTO where the found planned pipelines will be appended.
//
// Returns:
//   - This function does not return a value. It modifies the provided dtoArr slice directly.
func (p *PipelineService) CheckSended(dtoArr *[]dto.PipelineDTO) {
	log.Println("planned pipelines check begins...")
	counter := 0
	pArr, err := p.storage.PipelinesLoad(context.Background())
	if err != nil {
		return
	}

	for _, pipeline := range pArr {
		if pipeline.Status == "send_err" && pipeline.SendingCounter < 3 {
			counter++
			t, _ := p.storage.PipelineTemplateLoad(context.Background(), pipeline.TemplateId)
			tempDTO := dto.PipelineDTO{
				Pipeline:         pipeline,
				PipelineTemplate: t,
			}
			*dtoArr = append(*dtoArr, tempDTO)
		}
	}
	if counter == 0 {
		log.Println("have no planned pipelines!")
	}
}

// NewPipeline creates a new pipeline based on the provided event and pipeline template.
// It initializes the pipeline with the event ID, user ID, template ID, and sets the status to "new".
// The new pipeline is saved to the storage, and if there is a next pipeline template,
// it recursively creates the next pipeline using the same event and the next template.
//
// Parameters:
// - e: An event.Event object containing details about the event.
// - t: A pipeline.PipelineTemplate object representing the template for the new pipeline.
// - dtoArr: A pointer to a slice of dto.PipelineDTO where the created pipeline DTO will be appended.
//
// Returns:
// - This function does not return a value. It modifies the dtoArr slice to include the new pipeline DTO.
func (p *PipelineService) NewPipeline(e event.Event, t pipeline.PipelineTemplate, dtoArr *[]dto.PipelineDTO) {
	log.Printf("pipeline condition ok, event id: %d, pipeline template: %d", e.EventId, t.Id)
	newPipeline := pipeline.Pipeline{
		EventId:        e.EventId,
		UserId:         e.UserId,
		TemplateId:     t.Id,
		Status:         "new",
		SendingCounter: 0,
	}
	id, err := p.storage.PipelineSave(context.Background(), newPipeline)
	if err != nil {
		log.Printf("err: %v, pipeline id: %d", err, id)
	}
	newPipeline.Id = id

	tempDTO := dto.PipelineDTO{
		Pipeline:         newPipeline,
		PipelineTemplate: t,
	}
	*dtoArr = append(*dtoArr, tempDTO)

	if t.NextPipelineId != 0 {
		NextTemplate, _ := p.storage.PipelineTemplateLoad(context.Background(), t.NextPipelineId)
		p.NewPipeline(e, NextTemplate, dtoArr)
	}
}

// ExecutePipeline executes a pipeline after checking its state and limits.
// Sends the task to an executor, updates counters and finishes the pipeline on success.
// Returns errors in case of loading issues, exceeded retries, cancellation, or execution failures.
func (p *PipelineService) ExecutePipeline(dto dto.PipelineDTO) error {

	pipeline, err := p.storage.PipelineLoad(context.Background(), dto.Pipeline.Id)

	if err != nil {
		return err
	}

	if pipeline.SendingCounter >= 3 {
		err = fmt.Errorf("cant exec pipeline, sending counter >= 3")
		return err
	}

	if pipeline.Status == "cancelled" {
		err = fmt.Errorf("cant exec pipeline, status == cancelled")
		return err
	}

	log.Println("execute pipeline, id :", pipeline.Id, "counter: ", pipeline.SendingCounter)

	e := factory.ExecutorFactory(dto.PipelineTemplate.ExecuteType)

	p.storage.PipelineIncreaseSendCounter(context.Background(), pipeline.Id)

	err = e.Send(dto)

	if err != nil {
		log.Println(err)
		p.storage.PipelineSetStatus(context.Background(), pipeline.Id, "send_err")
		return err
	}

	p.storage.PipelineSetStatus(context.Background(), pipeline.Id, "finished")
	return nil
}

// InitPipelineTemplates initializes pipeline templates by loading them from storage.
// In case of failure during template loading, logs an error message but does not return any value.
func (p *PipelineService) InitPipelineTemplates() {
	templates, err := p.storage.PipelineTemplatesLoad(context.Background())
	if err != nil {
		log.Println("init pipeline templates err, cant load templates")
		return
	}
	p.pipelineTemplates = templates
}

// CompareConditions recursively compares two maps to determine if an event satisfies given conditions.
// It walks through the keys of the first map ('t') and ensures they exist in the second map ('e')
// and have equivalent values according to DeepEqual comparison.
// Nested maps are handled recursively, ensuring deep equality between nested key-value pairs.
// Returns true if all conditions are met; otherwise, returns false.
func (p *PipelineService) CompareConditions(t, e map[string]any) bool {
	empty_map := p.EventToMap(event.EmptyEvent())
	for k, v := range t {
		evVal, ok := e[k]
		if !ok {
			return false // ключа нет в event
		} else {
			emptyVal := empty_map[k]
			tVal := v
			if !reflect.DeepEqual(emptyVal, tVal) {
				switch vTyped := v.(type) {
				case map[string]any:
					evMap, ok := evVal.(map[string]any)
					if !ok {
						return false
					}
					if !p.CompareConditions(vTyped, evMap) {
						return false
					}
				default:
					if !reflect.DeepEqual(v, evVal) {
						return false
					}
				}
			}
		}
	}
	return true
}

// EventToMap converts an event into a generic map representation.
// It marshals the event into JSON format and then unmarshals it back into a map[string]interface{},
// allowing easy access to event properties as a dynamic structure.
// Any unmarshaling errors are ignored, assuming valid input.
func (p *PipelineService) EventToMap(e event.Event) map[string]any {
	e_data, _ := json.Marshal(e)
	e_map := make(map[string]any)
	json.Unmarshal(e_data, &e_map)
	return e_map
}
