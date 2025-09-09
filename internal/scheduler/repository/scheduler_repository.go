package scheduler_repository

import "event_service/internal/models/message"

type Scheduler_repo interface {
	InitScheduler() error
	StartScheduler()
	StopScheduler()
	CheckEvents()
	SendMessage(msg message.Message, event_id int) error
}
