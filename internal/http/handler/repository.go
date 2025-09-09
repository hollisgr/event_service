package handler

import (
	"github.com/gin-gonic/gin"
)

type Handler interface {
	// Register(mux *http.ServeMux)
	Register(r *gin.Engine)
}
