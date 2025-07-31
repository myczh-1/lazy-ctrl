package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/common"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"requestId,omitempty"`
}

// ErrorInfo represents error information in API responses
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse sends a successful API response
func SuccessResponse(c *gin.Context, data interface{}) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: GetCurrentTimestamp(),
		RequestID: GetRequestID(c),
	}

	c.JSON(http.StatusOK, response)
}

// CreatedResponse sends a created (201) API response
func CreatedResponse(c *gin.Context, data interface{}) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: GetCurrentTimestamp(),
		RequestID: GetRequestID(c),
	}

	c.JSON(http.StatusCreated, response)
}

// ErrorResponse sends an error API response
func ErrorResponse(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: GetCurrentTimestamp(),
		RequestID: GetRequestID(c),
	}

	c.JSON(statusCode, response)
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string, details map[string]interface{}) {
	ErrorResponse(c, http.StatusBadRequest, common.ErrorCodeValidation, message, details)
}

// NotFoundError sends a not found error response
func NotFoundError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, common.ErrorCodeNotFound, message, nil)
}

// ConflictError sends a conflict error response
func ConflictError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusConflict, common.ErrorCodeConflict, message, nil)
}

// UnauthorizedError sends an unauthorized error response
func UnauthorizedError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, common.ErrorCodeUnauthorized, message, nil)
}

// ForbiddenError sends a forbidden error response
func ForbiddenError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, common.ErrorCodeForbidden, message, nil)
}

// RateLimitError sends a rate limit error response
func RateLimitError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusTooManyRequests, common.ErrorCodeRateLimit, message, nil)
}

// InternalError sends an internal server error response
func InternalError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, common.ErrorCodeInternal, message, nil)
}

// TimeoutError sends a timeout error response
func TimeoutError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusRequestTimeout, common.ErrorCodeTimeout, message, nil)
}

// ServiceUnavailableError sends a service unavailable error response
func ServiceUnavailableError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusServiceUnavailable, common.ErrorCodeUnavailable, message, nil)
}

// HandleDomainError converts domain errors to appropriate HTTP responses
func HandleDomainError(c *gin.Context, err error) {
	code := common.GetErrorCode(err)
	context := common.GetErrorContext(err)

	switch code {
	case "COMMAND_NOT_FOUND":
		NotFoundError(c, err.Error())
	case "COMMAND_ALREADY_EXISTS":
		ConflictError(c, err.Error())
	case "SECURITY_ERROR":
		if common.IsErrorOfType(err, common.ErrInvalidPin) {
			UnauthorizedError(c, err.Error())
		} else if common.IsErrorOfType(err, common.ErrRateLimitExceeded) {
			RateLimitError(c, err.Error())
		} else {
			ForbiddenError(c, err.Error())
		}
	case "EXECUTION_ERROR":
		if common.IsErrorOfType(err, common.ErrExecutionTimeout) {
			TimeoutError(c, err.Error())
		} else {
			InternalError(c, err.Error())
		}
	case "CONFIGURATION_ERROR":
		ValidationError(c, err.Error(), context)
	case "REPOSITORY_ERROR":
		InternalError(c, err.Error())
	default:
		InternalError(c, "An unexpected error occurred")
	}
}

// responseWriter wraps gin.ResponseWriter to intercept JSON responses
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return len(b), nil
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return len(s), nil
}

// ResponseFormatterMiddleware wraps gin responses with standard format
func ResponseFormatterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip formatting for Swagger/docs paths
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/swagger") || 
		   strings.HasPrefix(path, "/docs") ||
		   strings.Contains(path, "swagger") {
			c.Next()
			return
		}

		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		c.Next()

		// Only format successful JSON responses
		statusCode := c.Writer.Status()
		if statusCode >= 200 && statusCode < 300 && w.body.Len() > 0 {
			var originalData interface{}
			if err := json.Unmarshal(w.body.Bytes(), &originalData); err == nil {
				// Skip if already formatted (contains success field)
				if dataMap, ok := originalData.(map[string]interface{}); ok {
					if _, hasSuccess := dataMap["success"]; hasSuccess {
						// Already formatted, write original
						w.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
						w.ResponseWriter.Write(w.body.Bytes())
						return
					}
				}

				// Format the response
				response := APIResponse{
					Success:   true,
					Data:      originalData,
					Timestamp: GetCurrentTimestamp(),
					RequestID: GetRequestID(c),
				}

				w.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
				responseBytes, _ := json.Marshal(response)
				w.ResponseWriter.Write(responseBytes)
				return
			}
		}

		// Write original response for non-JSON or error responses
		w.ResponseWriter.Write(w.body.Bytes())
	}
}
