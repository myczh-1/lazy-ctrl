package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		logger.WithFields(logrus.Fields{
			"method":     params.Method,
			"path":       params.Path,
			"status":     params.StatusCode,
			"latency":    params.Latency,
			"client_ip":  params.ClientIP,
			"user_agent": params.Request.UserAgent(),
		}).Info("HTTP request")
		return ""
	})
}