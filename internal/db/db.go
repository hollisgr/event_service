package db

import (
	"context"
	"encoding/json"
	"event_service/internal/db/storage"
	"event_service/internal/dto"
	"event_service/internal/models/event"
	"event_service/internal/models/message"
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

func (r *repository) Save(ctx context.Context, e event.Event) (int, error) {
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
				version_name
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			RETURNING event_id
			`
	// r.logger.Traceln("SQL Query:", formatQuery(query))
	r.client.QueryRow(ctx, query, e.DeviceCarrier, e.DeviceFamily, e.DeviceId, e.DeviceType, e.DisplayName, e.DMA, e.EventId, eventPropertiesBytes, e.EventTime, e.EventType, e.UserId, userPropertiesBytes, e.UUID, e.VersionName).Scan(&eventId)
	if eventId == 0 {
		return eventId, fmt.Errorf("failed to save event")
	}
	return eventId, nil
}

func (r *repository) LoadEvents(ctx context.Context) ([]dto.FromDbToMessagesDTO, error) {
	query := `
	SELECT 
		user_id,
		event_id,
		sending_attempts,
		is_planned
	FROM events
	WHERE 
		is_sended = false
		AND 
		sending_attempts < 3
	`
	rows, err := r.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	dtoArr := make([]dto.FromDbToMessagesDTO, 0)

	for rows.Next() {
		tempDTO := dto.FromDbToMessagesDTO{}
		tempDTO.Message = message.NewMessage()

		err = rows.Scan(&tempDTO.Message.UserId, &tempDTO.Event.EventId, &tempDTO.Event.SendingAttempts, &tempDTO.Event.IsPlanned)
		if err != nil {
			return nil, err
		}
		dtoArr = append(dtoArr, tempDTO)
	}
	return dtoArr, nil
}

func (r *repository) IncreaseEventSendCounter(ctx context.Context, event_id int) error {
	query := `
	UPDATE 
		events
	Set
		sending_attempts = sending_attempts + 1
	WHERE 
		event_id = $1
		RETURNING event_id
	`

	// r.logger.Traceln("SQL Query:", formatQuery(query))
	row := r.client.QueryRow(ctx, query, event_id)
	err := row.Scan(&event_id)
	if err != nil || event_id == 0 {
		return err
	}

	return nil
}

func (r *repository) MarkEventIsPlanned(ctx context.Context, event_id int) error {
	query := `
	UPDATE 
		events
	Set
		is_planned = true
	WHERE 
		event_id = $1
		RETURNING event_id
	`

	// r.logger.Traceln("SQL Query:", formatQuery(query))
	row := r.client.QueryRow(ctx, query, event_id)
	err := row.Scan(&event_id)
	if err != nil || event_id == 0 {
		return err
	}

	return nil
}

func (r *repository) MarkEventIsSended(ctx context.Context, event_id int) error {
	query := `
	UPDATE 
		events
	Set
		is_sended = true
	WHERE 
		event_id = $1
		RETURNING event_id
	`

	// r.logger.Traceln("SQL Query:", formatQuery(query))
	row := r.client.QueryRow(ctx, query, event_id)
	err := row.Scan(&event_id)
	if err != nil || event_id == 0 {
		return err
	}

	return nil
}
