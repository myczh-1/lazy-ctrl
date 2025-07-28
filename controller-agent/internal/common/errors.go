package common

import (
	"errors"
	"fmt"
)

// Domain error definitions
var (
	// Command errors
	ErrCommandNotFound      = errors.New("command not found")
	ErrCommandAlreadyExists = errors.New("command already exists")
	ErrCommandInvalidID     = errors.New("invalid command ID")
	ErrCommandInvalidConfig = errors.New("invalid command configuration")
	
	// Security errors
	ErrInvalidPin         = errors.New("invalid PIN")
	ErrPinRequired        = errors.New("PIN required")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrCommandNotAllowed  = errors.New("command not allowed")
	
	// Execution errors
	ErrExecutionFailed    = errors.New("command execution failed")
	ErrExecutionTimeout   = errors.New("command execution timeout")
	ErrPlatformNotSupported = errors.New("platform not supported")
	
	// Configuration errors
	ErrConfigNotFound     = errors.New("configuration not found")
	ErrConfigInvalid      = errors.New("invalid configuration")
	ErrConfigLoadFailed   = errors.New("failed to load configuration")
	
	// Repository errors
	ErrRepositoryNotFound = errors.New("resource not found in repository")
	ErrRepositoryConflict = errors.New("resource conflict in repository")
	ErrRepositoryInternal = errors.New("internal repository error")
)

// DomainError represents a domain-specific error with additional context
type DomainError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *DomainError) WithContext(key string, value interface{}) *DomainError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Error constructors
func NewCommandNotFoundError(id string) *DomainError {
	return &DomainError{
		Code:    "COMMAND_NOT_FOUND",
		Message: fmt.Sprintf("command with ID '%s' not found", id),
		Cause:   ErrCommandNotFound,
		Context: map[string]interface{}{"commandId": id},
	}
}

func NewCommandAlreadyExistsError(id string) *DomainError {
	return &DomainError{
		Code:    "COMMAND_ALREADY_EXISTS",
		Message: fmt.Sprintf("command with ID '%s' already exists", id),
		Cause:   ErrCommandAlreadyExists,
		Context: map[string]interface{}{"commandId": id},
	}
}

func NewSecurityError(message string, cause error) *DomainError {
	return &DomainError{
		Code:    "SECURITY_ERROR",
		Message: message,
		Cause:   cause,
	}
}

func NewExecutionError(command string, cause error) *DomainError {
	return &DomainError{
		Code:    "EXECUTION_ERROR",
		Message: fmt.Sprintf("failed to execute command '%s'", command),
		Cause:   cause,
		Context: map[string]interface{}{"command": command},
	}
}

func NewConfigurationError(message string, cause error) *DomainError {
	return &DomainError{
		Code:    "CONFIGURATION_ERROR",
		Message: message,
		Cause:   cause,
	}
}

func NewRepositoryError(operation, resource string, cause error) *DomainError {
	return &DomainError{
		Code:    "REPOSITORY_ERROR",
		Message: fmt.Sprintf("failed to %s %s", operation, resource),
		Cause:   cause,
		Context: map[string]interface{}{
			"operation": operation,
			"resource":  resource,
		},
	}
}

// IsErrorOfType checks if an error is of a specific domain error type
func IsErrorOfType(err error, target error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return errors.Is(domainErr.Cause, target)
	}
	return errors.Is(err, target)
}

// GetErrorCode extracts the error code from a domain error
func GetErrorCode(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return "UNKNOWN_ERROR"
}

// GetErrorContext extracts context from a domain error
func GetErrorContext(err error) map[string]interface{} {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Context
	}
	return nil
}