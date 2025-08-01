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

// NewStorage creates a new instance of repository implementing the storage.Storage interface.
// It takes a PostgreSQL client and a logger as parameters and returns a pointer to the repository struct.
func NewStorage(client postgres.Client, logger *logger.Logger) storage.Storage {
	return &repository{
		client: client,
		logger: logger,
	}
}

// PipelineLoad retrieves a single pipeline record from the database by its unique identifier.
// It uses SQL query to fetch the pipeline data from the 'pipelines' table where the primary key equals the provided pipeline_id.
// Scans the resulting row into a Pipeline struct and returns it along with any encountered errors.
// If no pipeline is found or other database errors occur, the error is propagated upwards.
func (r *repository) PipelineLoad(ctx context.Context, pipeline_id int) (pipeline.Pipeline, error) {
	var pipeline pipeline.Pipeline
	query := `
	SELECT 
		id,
		parent_id,
		event_id,
		user_id,
		template_id,
		status,
		created_at,
		sending_counter
	FROM pipelines
	WHERE id = $1
	`
	row := r.client.QueryRow(ctx, query, pipeline_id)
	err := row.Scan(&pipeline.Id, &pipeline.ParentId, &pipeline.EventId, &pipeline.UserId, &pipeline.TemplateId, &pipeline.Status, &pipeline.CreatedAt, &pipeline.SendingCounter)
	if err != nil {
		return pipeline, err
	}
	return pipeline, nil
}

// PipelinesLoad retrieves all pipeline records from the database.
// Executes a SELECT query on the 'pipelines' table without filters, fetching all columns.
// Each retrieved row is scanned into a Pipeline struct, which is appended to a slice.
// Returns a slice containing all pipelines and any encountered errors.
// If no results are found or database errors occur, the error is propagated upward.
func (r *repository) PipelinesLoad(ctx context.Context) ([]pipeline.Pipeline, error) {
	query := `
	SELECT 
		id,
		parent_id,
		event_id,
		user_id,
		template_id,
		status,
		created_at,
		sending_counter
	FROM pipelines
	`
	rows, err := r.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	pipelineArr := make([]pipeline.Pipeline, 0)

	for rows.Next() {
		tempPipeline := pipeline.Pipeline{}

		err = rows.Scan(&tempPipeline.Id, &tempPipeline.ParentId, &tempPipeline.EventId, &tempPipeline.UserId, &tempPipeline.TemplateId, &tempPipeline.Status, &tempPipeline.CreatedAt, &tempPipeline.SendingCounter)
		if err != nil {
			return nil, err
		}
		pipelineArr = append(pipelineArr, tempPipeline)
	}
	return pipelineArr, nil
}

// PipelineSetStatus updates the status field of a pipeline identified by its unique ID.
// Uses an UPDATE statement to change the status column of the corresponding entry in the 'pipelines' table.
// Verifies the update was successful by scanning the affected pipeline ID back from the RETURNING clause.
// Returns an error if the operation failed or if the pipeline wasn't properly updated.
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

// PipelineSave inserts a new pipeline record into the database.
// Constructs an INSERT query to add a new entry to the 'pipelines' table, populating parent_id, event_id, user_id, template_id, and status fields.
// Retrieves the generated pipeline ID using the RETURNING clause and scans it into the pipelineId variable.
// Returns the inserted pipeline ID and an error if insertion fails.
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

// PipelineIncreaseSendCounter increments the send counter for a pipeline identified by its unique ID.
// Performs an atomic update operation, increasing the 'sending_counter' field by 1.
// Confirms the update by scanning the pipeline ID back from the RETURNING clause.
// Returns an error if the update fails or if the pipeline isn't properly updated.
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

// EventSave inserts a new event record into the database.
// Serializes event and user properties to JSON bytes and includes them in the INSERT query.
// Populates multiple fields of the 'events' table, returning the auto-generated event ID.
// Handles serialization errors and reports them appropriately.
// Returns the event ID and an error if insertion fails.
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

// EventsLoadNew retrieves all new events from the database.
// Executes a SELECT query to fetch events whose status is marked as 'new', selecting various event attributes.
// Deserializes JSON-encoded event and user properties into structured types.
// Returns a slice of Event structs and any encountered errors.
// If no results are found or database errors occur, the error is propagated upward.
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

// EventLoad retrieves a single event record from the database by its unique identifier.
// Executes a SELECT query to fetch the event details from the 'events' table where the primary key matches the provided event_id.
// Scans the resulting row into an Event struct and returns it along with any encountered errors.
// If no event is found or other database errors occur, the error is propagated upward.
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

// EventSetStatus updates the status of an event identified by its unique ID.
// Issues an UPDATE query to modify the status column of the corresponding entry in the 'events' table.
// Verifies the update was successful by scanning the event ID back from the RETURNING clause.
// Returns an error if the operation failed or if the event wasn't properly updated.
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

// PipelineTemplateSave inserts a new pipeline template into the database.
// Marshals conditions and query data into JSON byte arrays and includes them in the INSERT query.
// Inserts a new entry into the 'pipeline_templates' table, capturing the autogenerated template ID.
// Reports marshaling errors or insertion failures explicitly.
// Returns the template ID and an error if insertion fails.
func (r *repository) PipelineTemplateSave(ctx context.Context, data pipeline.PipelineTemplateDTO) (int, error) {
	id := 0
	conditionsBytes, err := json.Marshal(data.Conditions)
	if err != nil {
		return id, fmt.Errorf("failed to marshall tempalate conditions")
	}

	queryBytes, err := json.Marshal(data.Query)
	if err != nil {
		return id, fmt.Errorf("failed to marshall template query")
	}
	query := `
	INSERT INTO pipeline_templates (
		event_name,
		conditions,
		query,
		exit_pipeline_name,
		next_pipeline_id,
		execute_delay,
		is_active
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`
	// r.logger.Traceln("SQL Query:", formatQuery(query))
	r.client.QueryRow(ctx, query, data.EventName, conditionsBytes, queryBytes, data.ExitPipelineName, data.NextPipelineId, data.ExecuteDelay, data.IsActive).Scan(&id)
	if id == 0 {
		return id, fmt.Errorf("pipeline template save error")
	}
	return id, nil
}

// PipelineTemplatesLoad retrieves all pipeline templates from the database.
// Executes a SELECT query to fetch all entries from the 'pipeline_templates' table.
// Each row is scanned into a PipelineTemplateDTO struct and appended to a slice.
// Returns a slice of PipelineTemplateDTO instances and any encountered errors.
// If no results are found or database errors occur, the error is propagated upward.
func (r *repository) PipelineTemplatesLoad(ctx context.Context) ([]pipeline.PipelineTemplateDTO, error) {
	query := `
	SELECT 
		id,
		event_name,
		conditions,
		query,
		exit_pipeline_name,
		next_pipeline_id,
		execute_delay,
		is_active
	FROM pipeline_templates
	`
	rows, err := r.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	tArr := make([]pipeline.PipelineTemplateDTO, 0)

	for rows.Next() {
		t := pipeline.PipelineTemplateDTO{}

		err = rows.Scan(&t.Id, &t.EventName, &t.Conditions, &t.Query, &t.ExitPipelineName, &t.NextPipelineId, &t.ExecuteDelay, &t.IsActive)
		if err != nil {
			return nil, err
		}
		tArr = append(tArr, t)
	}
	return tArr, nil
}

// PipelineTemplateLoad retrieves a single pipeline template from the database by its unique identifier.
// It selects all relevant fields from the 'pipeline_templates' table where the primary key matches the provided templateId.
// Scans the resulting row into a PipelineTemplate struct and returns it along with any encountered errors.
// If no template is found or other database errors occur, the error is propagated upward.
func (r *repository) PipelineTemplateLoad(ctx context.Context, templateId int) (pipeline.PipelineTemplate, error) {
	var tempalate pipeline.PipelineTemplate
	query := `
	SELECT 
		id,
		event_name,
		conditions,
		query,
		exit_pipeline_name,
		next_pipeline_id,
		execute_delay,
		is_active
	FROM pipeline_templates
	WHERE id = $1
	`
	row := r.client.QueryRow(ctx, query, templateId)
	err := row.Scan(&tempalate.Id, &tempalate.EventName, &tempalate.Conditions, &tempalate.Query, &tempalate.ExitPipelineName, &tempalate.NextPipelineId, &tempalate.ExecuteDelay, &tempalate.IsActive)
	if err != nil {
		return tempalate, err
	}
	return tempalate, nil
}
