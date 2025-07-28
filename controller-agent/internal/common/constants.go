package common

import "time"

// Application constants
const (
	// Application information
	AppName    = "lazy-ctrl-agent"
	AppVersion = "2.0.0"
	
	// Default timeouts
	DefaultCommandTimeout = 10 * time.Second
	DefaultHTTPTimeout    = 30 * time.Second
	DefaultShutdownTimeout = 30 * time.Second
	
	// Rate limiting
	DefaultRateLimitPerMinute = 60
	RateLimitCleanupInterval  = time.Minute
	
	// Security
	DefaultPinRequired = false
	PinMinLength      = 4
	PinMaxLength      = 32
	
	// Configuration
	DefaultConfigPath    = "configs/config.yaml"
	DefaultCommandsPath  = "configs/commands.json"
	CommandsVersion      = "3.0"
	
	// Server defaults
	DefaultHTTPPort = 7070
	DefaultGRPCPort = 7071
	DefaultMQTTPort = 1883
	
	// Logging
	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"
	
	// Command execution
	MaxCommandOutputSize = 1024 * 1024 // 1MB
	CommandBufferSize    = 1024
	
	// Platform support
	PlatformWindows = "windows"
	PlatformLinux   = "linux"
	PlatformDarwin  = "darwin"
	
	// Command types
	CommandTypeShell     = "shell"
	CommandTypeScript    = "script"
	CommandTypeSequence  = "sequence"
	CommandTypeTemplate  = "template"
)

// HTTP Status messages
const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusStarting  = "starting"
	StatusStopping  = "stopping"
)

// Error codes
const (
	ErrorCodeValidation    = "VALIDATION_ERROR"
	ErrorCodeNotFound      = "NOT_FOUND"
	ErrorCodeConflict      = "CONFLICT"
	ErrorCodeUnauthorized  = "UNAUTHORIZED"
	ErrorCodeForbidden     = "FORBIDDEN"
	ErrorCodeRateLimit     = "RATE_LIMIT_EXCEEDED"
	ErrorCodeInternal      = "INTERNAL_ERROR"
	ErrorCodeTimeout       = "TIMEOUT"
	ErrorCodeUnavailable   = "SERVICE_UNAVAILABLE"
)

// Header names
const (
	HeaderContentType     = "Content-Type"
	HeaderAuthorization   = "Authorization"
	HeaderXPin            = "X-Pin"
	HeaderXRequestID      = "X-Request-ID"
	HeaderXRealIP         = "X-Real-IP"
	HeaderXForwardedFor   = "X-Forwarded-For"
	HeaderUserAgent       = "User-Agent"
)

// Content types
const (
	ContentTypeJSON = "application/json"
	ContentTypeText = "text/plain"
	ContentTypeHTML = "text/html"
)

// Context keys
type ContextKey string

const (
	ContextKeyRequestID ContextKey = "request_id"
	ContextKeyUserIP    ContextKey = "user_ip"
	ContextKeyUserAgent ContextKey = "user_agent"
	ContextKeyStartTime ContextKey = "start_time"
)

// MQTT topics
const (
	MQTTTopicPrefix   = "lazy-ctrl"
	MQTTTopicCommand  = "command"
	MQTTTopicExecute  = "execute"
	MQTTTopicStatus   = "status"
	MQTTTopicResponse = "response"
)

// Default positions for UI layout
const (
	DefaultPositionX      = 0
	DefaultPositionY      = 0
	DefaultPositionWidth  = 1
	DefaultPositionHeight = 1
)

// Command categories
const (
	CategorySystem     = "system"
	CategoryNetwork    = "network"
	CategoryMedia      = "media"
	CategoryDevelopment = "development"
	CategoryUtility    = "utility"
	CategoryCustom     = "custom"
)

// Security levels
const (
	SecurityLevelPublic     = "public"
	SecurityLevelProtected  = "protected"
	SecurityLevelPrivate    = "private"
	SecurityLevelAdmin      = "admin"
)

// Execution results
const (
	ExecutionStatusSuccess   = "success"
	ExecutionStatusFailed    = "failed"
	ExecutionStatusTimeout   = "timeout"
	ExecutionStatusCancelled = "cancelled"
)