package service

import (
	"context"
	"event_service/internal/cfg"
	"event_service/internal/db/storage"
	"event_service/internal/http/handler"
	"event_service/internal/models/event"
	"event_service/pkg/logger"
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

func NewHandler(s storage.Storage, cfg *cfg.Cfg, validate *validator.Validate) handler.Handler {
	return &handler_service{
		logger:   logger.GetLogger(),
		storage:  s,
		cfg:      cfg,
		validate: validate,
	}
}

func (h *handler_service) Register(r *gin.Engine) {
	r.POST("/events", h.Event)
	r.GET("/auth", h.GetAuthToken)
}

func (h *handler_service) GetAuthToken(c *gin.Context) {
	token := CreateToken(h.cfg.JwtSecretKey)
	c.JSON(http.StatusOK, gin.H{"status": "OK", "access token": token})
}

func (h *handler_service) Event(c *gin.Context) {

	auth := c.GetHeader("Authorization")
	err := CheckHeader(auth)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "denied", "message": "access token is incorrect"})
		return
	}

	newEvent := event.NewEvent()

	err = c.BindJSON(&newEvent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request", "message": "req body data is incorrect"})
		return
	}

	err = h.validate.Struct(newEvent)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "val error", "message": "data validation error"})
		return
	}

	if newEvent.EventType == "start_webapp" {
		eventId, err := h.storage.Save(context.Background(), newEvent)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "database error", "message": "cant save data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "event id": eventId})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "event type is not 'start web'"})
	}
}
