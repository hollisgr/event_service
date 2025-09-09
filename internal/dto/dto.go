package dto

import (
	"event_service/internal/models/event"
	"event_service/internal/models/message"
)

type FromDbToMessagesDTO struct {
	Message message.Message
	Event   event.Event
}
