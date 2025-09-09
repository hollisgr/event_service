package storage

import (
	"context"
	"event_service/internal/dto"
	"event_service/internal/models/event"
)

type Storage interface {
	Save(ctx context.Context, e event.Event) (int, error)
	LoadEvents(ctx context.Context) ([]dto.FromDbToMessagesDTO, error)
	IncreaseEventSendCounter(ctx context.Context, event_id int) error
	MarkEventIsPlanned(ctx context.Context, event_id int) error
	MarkEventIsSended(ctx context.Context, event_id int) error
}
