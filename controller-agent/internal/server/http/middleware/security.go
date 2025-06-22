package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/security"
)

func SecurityMiddleware(config *config.Config, securityService *security.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成客户端ID用于限流
		clientID := securityService.GetClientID(c.ClientIP(), c.Request.UserAgent())
		
		// 检查限流
		if err := securityService.CheckRateLimit(clientID); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// PIN验证（如果需要）
		if config.Security.PinRequired {
			pin := c.GetHeader("X-Pin")
			if pin == "" {
				pin = c.Query("pin")
			}
			
			if !securityService.ValidatePin(pin) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid or missing PIN",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}