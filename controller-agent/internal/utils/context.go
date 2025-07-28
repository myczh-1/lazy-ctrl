package utils

import (
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/myczh-1/lazy-ctrl-agent/internal/common"
)

// GetRequestID extracts or generates a request ID from the context
func GetRequestID(c *gin.Context) string {
	// Try to get from header first
	requestID := c.GetHeader(common.HeaderXRequestID)
	if requestID != "" {
		return requestID
	}
	
	// Try to get from context
	if id, exists := c.Get(string(common.ContextKeyRequestID)); exists {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	
	// Generate new one
	requestID = uuid.New().String()
	c.Set(string(common.ContextKeyRequestID), requestID)
	return requestID
}

// SetRequestID sets the request ID in the context
func SetRequestID(c *gin.Context, requestID string) {
	c.Set(string(common.ContextKeyRequestID), requestID)
	c.Header(common.HeaderXRequestID, requestID)
}

// GetUserIP extracts the user's IP address from the context
func GetUserIP(c *gin.Context) string {
	// Try to get from context first
	if ip, exists := c.Get(string(common.ContextKeyUserIP)); exists {
		if userIP, ok := ip.(string); ok {
			return userIP
		}
	}
	
	// Extract from headers or client IP
	ip := c.GetHeader(common.HeaderXRealIP)
	if ip == "" {
		ip = c.GetHeader(common.HeaderXForwardedFor)
	}
	if ip == "" {
		ip = c.ClientIP()
	}
	
	c.Set(string(common.ContextKeyUserIP), ip)
	return ip
}

// GetUserAgent extracts the user agent from the context
func GetUserAgent(c *gin.Context) string {
	// Try to get from context first
	if ua, exists := c.Get(string(common.ContextKeyUserAgent)); exists {
		if userAgent, ok := ua.(string); ok {
			return userAgent
		}
	}
	
	userAgent := c.GetHeader(common.HeaderUserAgent)
	c.Set(string(common.ContextKeyUserAgent), userAgent)
	return userAgent
}

// GetStartTime extracts the request start time from the context
func GetStartTime(c *gin.Context) time.Time {
	if startTime, exists := c.Get(string(common.ContextKeyStartTime)); exists {
		if t, ok := startTime.(time.Time); ok {
			return t
		}
	}
	
	// Return current time as fallback
	return time.Now()
}

// SetStartTime sets the request start time in the context
func SetStartTime(c *gin.Context, startTime time.Time) {
	c.Set(string(common.ContextKeyStartTime), startTime)
}

// GetCurrentTimestamp returns the current timestamp in RFC3339 format
func GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// GetRequestDuration calculates the duration since the request started
func GetRequestDuration(c *gin.Context) time.Duration {
	startTime := GetStartTime(c)
	return time.Since(startTime)
}