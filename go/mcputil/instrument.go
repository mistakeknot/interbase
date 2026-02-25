// Package mcputil provides shared middleware for Demarch MCP servers.
//
// The primary entry point is [NewMetrics], which creates a metrics collector,
// and [Metrics.Instrument], which returns a [server.ToolHandlerMiddleware]
// that wraps every tool handler with timing, error counting, automatic
// ToolError wrapping, and panic recovery.
//
// Usage:
//
//	metrics := mcputil.NewMetrics()
//	s := server.NewMCPServer("my-server", "1.0",
//	    server.WithToolHandlerMiddleware(metrics.Instrument()),
//	)
package mcputil

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mistakeknot/interbase/toolerror"
)

// toolCounters holds atomic counters for a single tool.
type toolCounters struct {
	calls      atomic.Int64
	errors     atomic.Int64
	durationNs atomic.Int64
}

// Metrics collects per-tool call metrics.
type Metrics struct {
	mu    sync.RWMutex
	tools map[string]*toolCounters
}

// ToolStats is a snapshot of metrics for a single tool.
type ToolStats struct {
	Calls    int64         `json:"calls"`
	Errors   int64         `json:"errors"`
	Duration time.Duration `json:"total_duration"`
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{
		tools: make(map[string]*toolCounters),
	}
}

// countersFor returns the counters for a tool, creating them if needed.
func (m *Metrics) countersFor(name string) *toolCounters {
	m.mu.RLock()
	c, ok := m.tools[name]
	m.mu.RUnlock()
	if ok {
		return c
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// Double-check after acquiring write lock.
	if c, ok = m.tools[name]; ok {
		return c
	}
	c = &toolCounters{}
	m.tools[name] = c
	return c
}

// ToolMetrics returns a snapshot of metrics for all tools.
func (m *Metrics) ToolMetrics() map[string]ToolStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]ToolStats, len(m.tools))
	for name, c := range m.tools {
		out[name] = ToolStats{
			Calls:    c.calls.Load(),
			Errors:   c.errors.Load(),
			Duration: time.Duration(c.durationNs.Load()),
		}
	}
	return out
}

// Instrument returns a ToolHandlerMiddleware that wraps every tool handler
// with timing, error counting, ToolError wrapping, and panic recovery.
//
// Pass it to server.WithToolHandlerMiddleware when creating the MCP server.
func (m *Metrics) Instrument() server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, req mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
			name := req.Params.Name
			c := m.countersFor(name)
			c.calls.Add(1)
			start := time.Now()

			// Panic recovery — convert to structured ErrInternal.
			defer func() {
				if r := recover(); r != nil {
					te := toolerror.New(toolerror.ErrInternal, "panic in %s: %v", name, r)
					result = mcp.NewToolResultError(te.JSON())
					err = nil
					c.errors.Add(1)
				}
				c.durationNs.Add(int64(time.Since(start)))
			}()

			result, err = next(ctx, req)

			// If handler returned a Go error, wrap it as ToolError.
			if err != nil {
				te := toolerror.Wrap(err)
				result = mcp.NewToolResultError(te.JSON())
				err = nil
				c.errors.Add(1)
				return
			}

			// Count tool-level errors (isError results).
			if result != nil && result.IsError {
				c.errors.Add(1)
			}

			return
		}
	}
}

// WrapError is a standalone helper that converts a Go error into an MCP tool
// error result with structured ToolError JSON. Useful for handlers that want
// explicit error conversion without the full middleware.
func WrapError(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(toolerror.Wrap(err).JSON()), nil
}

// ValidationError returns an MCP tool error result for invalid input.
func ValidationError(format string, args ...any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(toolerror.New(toolerror.ErrValidation, format, args...).JSON()), nil
}

// NotFoundError returns an MCP tool error result for missing resources.
func NotFoundError(format string, args ...any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(toolerror.New(toolerror.ErrNotFound, format, args...).JSON()), nil
}

// ConflictError returns an MCP tool error result for concurrent modification conflicts.
func ConflictError(format string, args ...any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(toolerror.New(toolerror.ErrConflict, format, args...).JSON()), nil
}

// TransientError returns an MCP tool error result for temporary failures.
func TransientError(format string, args ...any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(toolerror.New(toolerror.ErrTransient, format, args...).JSON()), nil
}

// Ensure the type satisfies the interface at compile time.
var _ fmt.Stringer = (*ToolStats)(nil)

func (s *ToolStats) String() string {
	return fmt.Sprintf("calls=%d errors=%d duration=%s", s.Calls, s.Errors, s.Duration)
}
