package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"event_service/internal/cfg"
	"event_service/internal/db/storage"
	"event_service/internal/models/message"
	scheduler_repository "event_service/internal/scheduler/repository"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-co-op/gocron/v2"
)

type Scheduler struct {
	scheduler gocron.Scheduler
	storage   storage.Storage
	cfg       *cfg.Cfg
}

func NewScheduler(storage storage.Storage, cfg *cfg.Cfg) scheduler_repository.Scheduler_repo {
	s, err := gocron.NewScheduler()

	if err != nil {
		log.Println("cant create new scheduler")
		return nil
	}

	return &Scheduler{
		scheduler: s,
		storage:   storage,
		cfg:       cfg,
	}
}

func (s *Scheduler) InitScheduler() error {
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(
			time.Duration(s.cfg.Scheduler.CheckEventsTimeOutSec)*time.Second,
		),
		gocron.NewTask(s.CheckEvents),
	)
	if err != nil {
		log.Println("cant create new job")
		return err
	}
	return nil
}

func (s *Scheduler) StartScheduler() {
	s.scheduler.Start()
	log.Println("Scheduler started!")
}

func (s *Scheduler) StopScheduler() {
	s.scheduler.Shutdown()
	log.Println("Scheduler stopped!")
}

func (s *Scheduler) CheckEvents() {
	log.Println("Checking events")
	msgCounter := 0
	messages, err := s.storage.LoadEvents(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	for _, msg := range messages {
		if msg.Event.SendingAttempts == 0 && !msg.Event.IsPlanned {
			msgCounter++
			log.Println("Planning event id:", msg.Event.EventId)
			_, _ = s.scheduler.NewJob(
				gocron.OneTimeJob(
					gocron.OneTimeJobStartDateTime(time.Now().Add(time.Duration(s.cfg.Scheduler.SendMessageTimeOutSec)*time.Second)),
				),
				gocron.NewTask(
					func() {
						err = s.storage.MarkEventIsPlanned(context.Background(), msg.Event.EventId)
						if err != nil {
							log.Println(err)
							return
						}
						err = s.storage.IncreaseEventSendCounter(context.Background(), msg.Event.EventId)
						if err != nil {
							log.Println(err)
							return
						}
						err = s.SendMessage(msg.Message, msg.Event.EventId)
						if err != nil {
							log.Println(err)
							return
						}
					},
				),
			)
		}
		if msg.Event.SendingAttempts < 3 && msg.Event.IsPlanned {
			msgCounter++
			log.Printf("Resend event_id: %d, count: %d", msg.Event.EventId, msg.Event.SendingAttempts)
			_, _ = s.scheduler.NewJob(
				gocron.OneTimeJob(
					gocron.OneTimeJobStartDateTime(time.Now().Add(time.Duration(1)*time.Second)),
				),
				gocron.NewTask(
					func() {
						err = s.storage.IncreaseEventSendCounter(context.Background(), msg.Event.EventId)
						if err != nil {
							log.Println(err)
							return
						}
						err = s.SendMessage(msg.Message, msg.Event.EventId)
						if err != nil {
							log.Println(err)
							return
						}
					},
				),
			)
		}
	}
	if msgCounter == 0 {
		log.Println("have no avaible events!")
	}
}

func (s *Scheduler) SendMessage(msg message.Message, event_id int) error {
	c := http.Client{}
	msg.CohortName = fmt.Sprintf("event_id %d", event_id)
	reqAddr := s.cfg.Scheduler.Host
	reqBody, err := json.Marshal(msg)
	if err != nil {
		log.Println("cant marshal new message")
		return err
	}
	token := "Bearer " + s.cfg.Scheduler.Token

	req, err := http.NewRequest("POST", reqAddr, bytes.NewReader(reqBody))

	if err != nil {
		log.Println("creating new request error")
		return err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)

	if err != nil {
		log.Println("sending request error")
		return err
	}

	respByte, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Println("reading responce body error")
		return err
	}

	s.storage.MarkEventIsSended(context.Background(), event_id)
	log.Println(string(respByte))
	return err
}
