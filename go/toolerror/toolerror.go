// Package toolerror defines the shared structured error contract for Demarch MCP servers.
//
// All Demarch MCP tool handlers should return ToolError instead of flat error strings.
// This enables agents to distinguish transient from permanent failures and make
// informed retry decisions.
//
// Usage:
//
//	return toolerror.New(toolerror.ErrNotFound, "agent %q not registered", agentName)
//	return toolerror.New(toolerror.ErrTransient, "database busy").WithRecoverable(true)
package toolerror

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Error type constants. Use these as the Type field in ToolError.
const (
	ErrNotFound   = "NOT_FOUND"   // Resource doesn't exist
	ErrConflict   = "CONFLICT"    // Concurrent modification conflict
	ErrValidation = "VALIDATION"  // Invalid input or arguments
	ErrPermission = "PERMISSION"  // Access denied
	ErrTransient  = "TRANSIENT"   // Temporary failure, safe to retry
	ErrInternal   = "INTERNAL"    // Unexpected server error
)

// ToolError is a structured error for MCP tool handlers.
// It carries enough context for agents to make retry and fallback decisions.
type ToolError struct {
	Type        string         `json:"type"`
	Message     string         `json:"message"`
	Recoverable bool           `json:"recoverable"`
	Data        map[string]any `json:"data,omitempty"`
}

// New creates a ToolError with the given type and formatted message.
// Recoverable defaults to true for ErrTransient, false for all others.
func New(errType string, format string, args ...any) *ToolError {
	return &ToolError{
		Type:        errType,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: errType == ErrTransient,
	}
}

// Error implements the error interface.
func (e *ToolError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// WithRecoverable sets the Recoverable flag and returns the error for chaining.
func (e *ToolError) WithRecoverable(r bool) *ToolError {
	e.Recoverable = r
	return e
}

// WithData sets the Data field and returns the error for chaining.
func (e *ToolError) WithData(data map[string]any) *ToolError {
	e.Data = data
	return e
}

// JSON returns the error as a JSON string for embedding in MCP tool results.
func (e *ToolError) JSON() string {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf(`{"type":"%s","message":"%s","recoverable":%t}`, e.Type, e.Message, e.Recoverable)
	}
	return string(b)
}

// FromError extracts a *ToolError from an error chain.
// Returns nil if the error is not a ToolError.
func FromError(err error) *ToolError {
	var te *ToolError
	if errors.As(err, &te) {
		return te
	}
	return nil
}

// Wrap converts any error into a ToolError. If the error is already a ToolError,
// it is returned as-is. Otherwise, it's wrapped as ErrInternal.
func Wrap(err error) *ToolError {
	if te := FromError(err); te != nil {
		return te
	}
	return New(ErrInternal, "%s", err.Error())
}
