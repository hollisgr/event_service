package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"event_service/internal/cfg"
	"event_service/internal/db/storage"
	"event_service/internal/dto"
	"event_service/internal/models/event"
	"event_service/internal/models/pipeline"
	"event_service/internal/pipeline/pipeline_interface"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
)

type PipelineService struct {
	pipelineTemplates []pipeline.PipelineTemplate
	storage           storage.Storage
	cfg               cfg.Cfg
}

func NewPipelineService(s storage.Storage, cfg cfg.Cfg) pipeline_interface.Pipeline {
	var p PipelineService
	p.storage = s
	p.cfg = cfg
	p.InitPipelineTemplates()
	return &p
}

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
	p.CheckPlanned(&dtoArr)
	if counter == 0 {
		log.Println("have no new pipelines!")
	}
	return dtoArr, nil
}

func (p *PipelineService) CheckCancelState(e event.Event) {
	pArr, _ := p.storage.PipelinesLoad(context.Background())
	for _, pipeline := range pArr {
		if pipeline.UserId == e.UserId {
			t, _ := p.storage.PipelineTemplateLoad(context.Background(), pipeline.TemplateId)
			if t.ExitPipelineName == e.EventType {
				p.storage.PipelineSetStatus(context.Background(), pipeline.Id, "cancelled")
				log.Println("canceling pipeline with id: ", pipeline.Id)
			}
		}
	}
}

func (p *PipelineService) CheckConditions(e event.Event) (pipeline.PipelineTemplate, bool) {
	template := pipeline.PipelineTemplate{}
	log.Printf("checking event: %d", e.EventId)
	ok := false

	for _, t := range p.pipelineTemplates {
		if e.EventType == t.EventName {
			e_map := p.EventToMap(e)
			t_map := p.EventToMap(t.Conditions)
			ok = p.CompareConditions(t_map, e_map)
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
func (p *PipelineService) CheckPlanned(dtoArr *[]dto.PipelineDTO) {
	log.Println("planned pipelines check begins...")
	counter := 0
	pArr, err := p.storage.PipelinesLoad(context.Background())
	if err != nil {
		return
	}

	for _, pipeline := range pArr {
		if pipeline.Status == "planned" && pipeline.SendingCounter < 3 {
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

// ExecutePipeline executes a pipeline based on the provided PipelineDTO.
// It loads the pipeline from storage and checks the sending counter and status.
// If the sending counter is 3 or more, or if the status is "cancelled",
// it returns an error. If checks pass, it increments the sending counter,
// prepares an HTTP request to the scheduler with the pipeline template query,
// and sends the request. Finally, it updates the pipeline status to "finished".
//
// Parameters:
//   - dto: A PipelineDTO containing the details of the pipeline to execute.
//
// Returns:
//   - An error if any step fails; otherwise, returns nil.
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

	p.storage.PipelineIncreaseSendCounter(context.Background(), pipeline.Id)

	c := http.Client{}
	reqAddr := p.cfg.Scheduler.Host
	reqBody, err := json.Marshal(dto.PipelineTemplate.Query)
	if err != nil {
		return err
	}

	token := "Bearer " + p.cfg.Token

	req, err := http.NewRequest("POST", reqAddr, bytes.NewReader(reqBody))

	if err != nil {
		return err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)

	if err != nil {
		return err
	}

	respByte, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	p.storage.PipelineSetStatus(context.Background(), pipeline.Id, "finished")
	log.Println(string(respByte))
	return nil
}

// InitPipelineTemplates initializes the pipeline templates by loading them from the storage.
// It retrieves the pipeline templates, and for each template, it creates a corresponding
// PipelineTemplate object by unmarshalling the conditions and query from JSON bytes.
// If there is an error during the loading process, it logs an error message and exits the function.
// The initialized templates are then appended to the pipelineTemplates slice of the PipelineService.
func (p *PipelineService) InitPipelineTemplates() {
	pipelineTemplatesArr, err := p.storage.PipelineTemplatesLoad(context.Background())
	if err != nil {
		log.Println("init pipeline templates err, cant load templates")
		return
	}

	for _, t := range pipelineTemplatesArr {
		tempEvent := event.EmptyEvent()
		tempQuery := pipeline.EmptyQuery()

		conditionsBytes, _ := json.Marshal(t.Conditions)
		queryBytes, _ := json.Marshal(t.Query)

		json.Unmarshal(conditionsBytes, &tempEvent)
		json.Unmarshal(queryBytes, &tempQuery)

		tempTemplate := pipeline.PipelineTemplate{
			Id:               t.Id,
			EventName:        t.EventName,
			Conditions:       tempEvent,
			Query:            tempQuery,
			ExitPipelineName: t.ExitPipelineName,
			NextPipelineId:   t.NextPipelineId,
			ExecuteDelay:     t.ExecuteDelay,
			IsActive:         t.IsActive,
		}

		p.pipelineTemplates = append(p.pipelineTemplates, tempTemplate)
	}
}

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

func (p *PipelineService) EventToMap(e event.Event) map[string]any {
	e_data, _ := json.Marshal(e)
	e_map := make(map[string]any)
	json.Unmarshal(e_data, &e_map)
	return e_map
}
