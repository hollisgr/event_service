package db

import (
	"context"
	"encoding/json"
	"event_service/internal/db/storage"
	"event_service/internal/models/event"
	"event_service/internal/models/pipeline"
	"event_service/pkg/logger"
	"event_service/pkg/postgres"
	"fmt"
	"strings"
)

type repository struct {
	client postgres.Client
	logger *logger.Logger
}

func formatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}

func NewStorage(client postgres.Client, logger *logger.Logger) storage.Storage {
	return &repository{
		client: client,
		logger: logger,
	}
}

func (r *repository) PipelinesLoad(ctx context.Context) ([]pipeline.Pipeline, error) {
	query := `
	SELECT 
		id,
		parent_id,
		event_id,
		user_id,
		template_id,
		status,
		created_at
	FROM pipelines
	`
	rows, err := r.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	pipelineArr := make([]pipeline.Pipeline, 0)

	for rows.Next() {
		tempPipeline := pipeline.Pipeline{}

		err = rows.Scan(&tempPipeline.Id, &tempPipeline.ParentId, &tempPipeline.EventId, &tempPipeline.UserId, &tempPipeline.TemplateId, &tempPipeline.Status, &tempPipeline.CreatedAt)
		if err != nil {
			return nil, err
		}
		pipelineArr = append(pipelineArr, tempPipeline)
	}
	return pipelineArr, nil
}

func (r *repository) PipelineSetStatus(ctx context.Context, pipeline_id int, status string) error {
	query := `
	UPDATE 
		pipelines
	Set
		status = $2
	WHERE 
		id = $1
		RETURNING id
	`
	// r.logger.Traceln("SQL Query:", formatQuery(query))
	row := r.client.QueryRow(ctx, query, pipeline_id, status)
	err := row.Scan(&pipeline_id)
	if err != nil || pipeline_id == 0 {
		return err
	}
	return nil
}

func (r *repository) PipelineSave(ctx context.Context, p pipeline.Pipeline) (int, error) {
	pipelineId := 0

	query := `
			INSERT INTO pipelines (
				parent_id,
				event_id,
				user_id,
				template_id,
				status
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
			`
	// r.logger.Traceln("SQL Query:", formatQuery(query))
	r.client.QueryRow(ctx, query, p.ParentId, p.EventId, p.UserId, p.TemplateId, p.Status).Scan(&pipelineId)
	if pipelineId == 0 {
		return pipelineId, fmt.Errorf("failed to save pipeline")
	}
	return pipelineId, nil
}

func (r *repository) PipelineIncreaseSendCounter(ctx context.Context, pipeline_id int) error {
	query := `
	UPDATE 
		pipelines
	Set
		sending_counter = sending_counter + 1
	WHERE 
		id = $1
		RETURNING id
	`

	// r.logger.Traceln("SQL Query:", formatQuery(query))
	row := r.client.QueryRow(ctx, query, pipeline_id)
	err := row.Scan(&pipeline_id)
	if err != nil || pipeline_id == 0 {
		return err
	}

	return nil
}

func (r *repository) EventSave(ctx context.Context, e event.Event) (int, error) {
	eventId := 0

	eventPropertiesBytes, err := json.Marshal(e.EventProperties)
	if err != nil {
		return eventId, fmt.Errorf("failed to marshall event properties")
	}

	userPropertiesBytes, err := json.Marshal(e.UserProperties)
	if err != nil {
		return eventId, fmt.Errorf("failed to marshall user properties")
	}

	query := `
			INSERT INTO events (
				device_carrier, 
				device_family, 
				device_id, 
				device_type, 
				display_name, 
				dma, 
				event_id, 
				event_properties,
				event_time, 
				event_type, 
				user_id, 
				user_properties,
				uuid, 
				version_name,
				status
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			RETURNING event_id
			`
	// r.logger.Traceln("SQL Query:", formatQuery(query))
	r.client.QueryRow(ctx, query,
		e.DeviceCarrier, e.DeviceFamily, e.DeviceId,
		e.DeviceType, e.DisplayName, e.DMA,
		e.EventId, eventPropertiesBytes, e.EventTime,
		e.EventType, e.UserId, userPropertiesBytes,
		e.UUID, e.VersionName, e.Status).Scan(&eventId)
	if eventId == 0 {
		return eventId, fmt.Errorf("failed to save event")
	}
	return eventId, nil
}

func (r *repository) EventsLoadNew(ctx context.Context) ([]event.Event, error) {
	query := `
	SELECT 
		device_carrier, 
		device_family, 
		device_id, 
		device_type, 
		display_name, 
		dma, 
		event_id, 
		event_properties,
		event_time, 
		event_type, 
		user_id, 
		user_properties,
		uuid, 
		version_name,
		status
	FROM events
	WHERE 
		status = 'new'
	`
	rows, err := r.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	eventArr := make([]event.Event, 0)

	for rows.Next() {
		tempEvent := event.Event{}
		var tempEventProperties []byte
		var tempUserProperties []byte

		err = rows.Scan(&tempEvent.DeviceCarrier, &tempEvent.DeviceFamily, &tempEvent.DeviceId,
			&tempEvent.DeviceType, &tempEvent.DisplayName, &tempEvent.DMA,
			&tempEvent.EventId, &tempEventProperties, &tempEvent.EventTime,
			&tempEvent.EventType, &tempEvent.UserId, &tempUserProperties,
			&tempEvent.UUID, &tempEvent.VersionName, &tempEvent.Status)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(tempEventProperties, &tempEvent.EventProperties)
		json.Unmarshal(tempUserProperties, &tempEvent.UserProperties)
		eventArr = append(eventArr, tempEvent)
	}
	return eventArr, nil
}

func (r *repository) EventLoad(ctx context.Context, event_id int) (event.Event, error) {
	query := `
		SELECT 
			device_carrier,
			device_family,
			device_id ,
			device_type,
			display_name,
			dma,
			event_id,
			event_time,
			event_type,
			user_id,
			uuid,
			version_name,
			status
		FROM
			events
		WHERE event_id = $1
		`
	e := event.Event{}
	row := r.client.QueryRow(ctx, query, event_id)
	err := row.Scan(
		&e.DeviceCarrier, &e.DeviceFamily, &e.DeviceId,
		&e.DeviceType, &e.DisplayName, &e.DMA,
		&e.EventId, &e.EventProperties, &e.EventTime,
		&e.EventType, &e.UserId, &e.UUID,
		&e.Status, &e.VersionName)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (r *repository) EventSetStatus(ctx context.Context, event_id int, status string) error {
	query := `
	UPDATE 
		events
	Set
		status = $2
	WHERE 
		event_id = $1
		RETURNING event_id
	`

	// r.logger.Traceln("SQL Query:", formatQuery(query))
	row := r.client.QueryRow(ctx, query, event_id, status)
	err := row.Scan(&event_id)
	if err != nil || event_id == 0 {
		return err
	}

	return nil
}
