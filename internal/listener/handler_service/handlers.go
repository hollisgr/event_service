package handler_service

import (
	"context"
	"event_service/internal/cfg"
	"event_service/internal/db/storage"
	"event_service/internal/listener/handler_interface"
	"event_service/internal/models/event"
	"event_service/internal/models/pipeline"
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

// NewHandler creates a new handler instance.
// It takes dependencies like storage, configuration, and validator, and initializes a Logger.
// Returns a concrete implementation of the Handler interface.
func NewHandler(s storage.Storage, cfg *cfg.Cfg, validate *validator.Validate) handler_interface.Handler {
	return &handler_service{
		logger:   logger.GetLogger(),
		storage:  s,
		cfg:      cfg,
		validate: validate,
	}
}

// Register registers routes for the web server using Gin framework.
// Adds middleware for authentication and defines endpoints for saving and retrieving events/pipeline templates.
func (h *handler_service) Register(r *gin.Engine) {
	r.Use(AuthMiddleware(h.cfg.Token))
	r.POST("/events", h.SaveEvent)
	r.GET("/events", h.GetEvents)
	r.POST("/pipeline_templates", h.SavePipelineTemplate)
	r.GET("/pipeline_templates", h.GetPipelineTemplates)
}

// GetPipelineTemplates fetches all pipeline templates from storage.
// On success, responds with a list of pipeline templates.
// On database error, returns Bad Request response with error details.
func (h *handler_service) GetPipelineTemplates(c *gin.Context) {

	templates, err := h.storage.PipelineTemplatesLoad(context.Background())

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "database error", "message": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "pipeline_templates": templates})
}

// SavePipelineTemplate saves a new pipeline template to storage.
// Binds JSON payload to PipelineTemplateDTO, validates, and stores it.
// Responds with the saved pipeline template's ID on success, or bad request error on validation/database issues.
func (h *handler_service) SavePipelineTemplate(c *gin.Context) {
	var newTemplate pipeline.PipelineTemplate
	err := c.BindJSON(&newTemplate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request", "message": "req body data is incorrect"})
		return
	}

	if newTemplate.ExecuteType == "" {
		newTemplate.ExecuteType = "default"
	}

	id, err := h.storage.PipelineTemplateSave(context.Background(), newTemplate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "bad request", "message": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "pipeline_template id": id})
}

// SaveEvent persists a new event in storage.
// Validates the event data and saves it.
// Responds with the saved event's ID on success, or error messages on validation or database problems.
func (h *handler_service) SaveEvent(c *gin.Context) {

	newEvent := event.EmptyEvent()

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

// GetEvents retrieves all new events from storage.
// On success, returns a list of events.
// On database error, returns Bad Request response with error details.
func (h *handler_service) GetEvents(c *gin.Context) {

	events, err := h.storage.EventsLoadNew(context.Background())

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "database error", "message": fmt.Sprintf("%v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "events": events})
}
