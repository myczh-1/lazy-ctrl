package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// HandleHealth returns the health status of the service
//
//	@Summary		Health check
//	@Description	Get the health status of the lazy-ctrl-agent service
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Param			X-Pin	header		string	false	"PIN for authentication (if required)"
//	@Success		200		{object}	gin.H{status=string,timestamp=int}
//	@Failure		401		{object}	gin.H	"Unauthorized"
//	@Failure		429		{object}	gin.H	"Too Many Requests"
//	@Security		PinAuth
//	@Router			/health [get]
func (h *SystemHandler) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	})
}