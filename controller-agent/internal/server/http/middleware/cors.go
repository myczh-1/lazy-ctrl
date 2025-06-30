package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 添加CORS支持
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许的源
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		// 设置CORS头
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Pin")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24小时

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}