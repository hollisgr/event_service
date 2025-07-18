package handler_service

import (
	"context"
	"event_service/internal/cfg"
	"event_service/internal/db/storage"
	"event_service/internal/listener/handler_interface"
	"event_service/internal/models/event"
	"event_service/pkg/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type handler_service struct {
	logger   *logger.Logger
	storage  storage.Storage
	cfg      *cfg.Cfg
	validate *validator.Validate
}

func NewHandler(s storage.Storage, cfg *cfg.Cfg, validate *validator.Validate) handler_interface.Handler {
	return &handler_service{
		logger:   logger.GetLogger(),
		storage:  s,
		cfg:      cfg,
		validate: validate,
	}
}

func (h *handler_service) Register(r *gin.Engine) {
	r.Use(AuthMiddleware(h.cfg.Token))
	r.POST("/events", h.Event)
	r.GET("/events", h.GetEvents)
}

func (h *handler_service) Event(c *gin.Context) {

	newEvent := event.NewEvent()

	err := c.BindJSON(&newEvent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request", "message": "req body data is incorrect"})
		return
	}

	err = h.validate.Struct(newEvent)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "val error", "message": fmt.Sprintf("%v", err)})
		return
	}

	eventId, err := h.storage.EventSave(context.Background(), newEvent)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "database error", "message": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "event id": eventId})
}

func (h *handler_service) GetEvents(c *gin.Context) {

	events, err := h.storage.EventsLoadNew(context.Background())

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "database error", "message": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "events": events})
}
