package pipeline

import (
	"context"
	"event_service/internal/db/storage"
	"event_service/internal/models/event"
	"event_service/internal/models/pipeline"
	"event_service/internal/pipeline/pipeline_interface"
	"log"
	"reflect"
)

type PipelineService struct {
	pipelineTemplates []pipeline.PipelineTemplate
	storage           storage.Storage
}

func NewPipelineService(s storage.Storage) pipeline_interface.Pipeline {
	var p PipelineService
	p.InitPipelineTemplates()
	p.storage = s
	return &p
}

func (p *PipelineService) checkConditions(e event.Event) (int, bool) {
	log.Printf("checking event: %d", e.EventId)
	ok := false
	templateId := 0

	for _, template := range p.pipelineTemplates {
		if template.EventName == e.EventType {
			t_reflect := reflect.ValueOf(template.Conditions.UserProperties)
			t_keys := t_reflect.MapKeys()
			for _, t_key := range t_keys {
				t_key_name := t_key.String()
				t_key_value := t_reflect.MapIndex(t_key).Interface()
				if e.UserProperties[t_key_name] == t_key_value {
					ok = true
				} else {
					ok = false
				}
			}
			if ok {
				templateId = template.Id
			}
		}
	}
	p.storage.EventSetStatus(context.Background(), e.EventId, "checked")
	log.Printf("event %d status updated", e.EventId)
	return templateId, ok
}

func (p *PipelineService) CheckPipelines() error {
	log.Println("pipeline check begins...")
	events, err := p.storage.EventsLoadNew(context.Background())
	if err != nil {
		return err
	}

	for _, event := range events {
		templateId, ok := p.checkConditions(event)
		if ok {
			log.Printf("pipeline condition ok, event id: %d, pipeline template: %d", event.EventId, templateId)
			newPipeline := pipeline.Pipeline{
				EventId:        event.EventId,
				UserId:         event.UserId,
				TemplateId:     templateId,
				Status:         "new",
				SendingCounter: 0,
			}
			id, err := p.storage.PipelineSave(context.Background(), newPipeline)

			if err != nil {
				log.Printf("err: %v, pipeline id: %d", err, id)
			}
		}
	}
	return nil
}

func (p *PipelineService) ExecutePipeline(pipelineId int) error {
	// c := http.Client{}
	// msg.CohortName = fmt.Sprintf("event_id %d", event_id)
	// reqAddr := s.cfg.Scheduler.Host
	// reqBody, err := json.Marshal(msg)
	// if err != nil {
	// 	log.Println("cant marshal new message")
	// 	return err
	// }
	// token := "Bearer " + s.cfg.Token

	// req, err := http.NewRequest("POST", reqAddr, bytes.NewReader(reqBody))

	// if err != nil {
	// 	log.Println("creating new request error")
	// 	return err
	// }

	// req.Header.Set("Authorization", token)
	// req.Header.Set("Content-Type", "application/json")

	// resp, err := c.Do(req)

	// if err != nil {
	// 	log.Println("sending request error")
	// 	return err
	// }

	// respByte, err := io.ReadAll(resp.Body)

	// if err != nil {
	// 	log.Println("reading responce body error")
	// 	return err
	// }

	// s.storage.EventSetStatus(context.Background(), event_id, FINISHED)
	// log.Println(string(respByte))
	return nil
}

func (p *PipelineService) InitPipelineTemplates() {

	pipelineTemplatesArr := make([]pipeline.PipelineTemplate, 0)

	newUserEvent := event.EmptyEvent()
	newUserEvent.UserProperties["status"] = "buyer"
	newUserEvent.UserProperties["language_code"] = "ru"

	newUser := pipeline.PipelineTemplate{
		Id:         1,
		EventName:  "new_user",
		Conditions: newUserEvent,
		Query: struct {
			CohortName string
			Message    string
			Image      string
			Delay      int
		}{
			CohortName: "GH-349-from-new_user-to-start_webapp_message-1",
			Message:    "<b>Привет, друг!</b> Рады что ты с нами. Не забудь обменять свои денежки. Тем более что с нами это сделать <u>супер удобно!</u> Множество способов получения и оплаты, удобный и стильный интерфейс и разнообразие курсов! Жми <b>«♻️ Начать обмен»</b>",
			Image:      "https://i.ibb.co/3s0Rzjy/Frame-354.png",
			Delay:      1,
		},
		ExitPipelineName: "start_webapp",
		NextPipelineId:   2,
		ExecuteDelay:     900,
		IsActive:         true,
	}

	pipelineTemplatesArr = append(pipelineTemplatesArr, newUser)

	newUserEvent2 := event.EmptyEvent()

	newUser2 := pipeline.PipelineTemplate{
		Id:         2,
		EventName:  "",
		Conditions: newUserEvent2,
		Query: struct {
			CohortName string
			Message    string
			Image      string
			Delay      int
		}{
			CohortName: "GH-349-from-new_user-to-start_webapp_message-2",
			Message:    "😢 - это наши партнеры в ожидании тебя, друг. Каждый из них готов помочь тебе осуществить мечту. Сделай первый шаг - Жми <b>«♻️ Начать обмен»</b>",
			Delay:      1,
		},
		ExitPipelineName: "start_webapp",
		NextPipelineId:   3,
		ExecuteDelay:     6300,
		IsActive:         true,
	}

	pipelineTemplatesArr = append(pipelineTemplatesArr, newUser2)

	newUserEvent3 := event.EmptyEvent()

	newUser3 := pipeline.PipelineTemplate{
		Id:         3,
		EventName:  "",
		Conditions: newUserEvent3,
		Query: struct {
			CohortName string
			Message    string
			Image      string
			Delay      int
		}{
			CohortName: "GH-349-from-new_user-to-start_webapp_message-2",
			Message:    "<b>Время - деньги.</b> Не теряй деньги, отправь запрос на обмен и получи наличку уже сейчас. Жми <b>«♻️ Начать обмен»</b>!\nВ случае, если есть дополнительные вопросы напишите нам @yeloo_support",
			Image:      "https://i.ibb.co/LPff9wv/Frame-303.png",
			Delay:      1,
		},
		ExitPipelineName: "start_webapp",
		ExecuteDelay:     79200,
		IsActive:         true,
	}

	pipelineTemplatesArr = append(pipelineTemplatesArr, newUser3)

	p.pipelineTemplates = pipelineTemplatesArr
}
