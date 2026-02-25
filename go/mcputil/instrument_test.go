package mcputil_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/mistakeknot/interbase/mcputil"
	"github.com/mistakeknot/interbase/toolerror"
)

func makeReq(name string) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: name},
	}
}

func TestInstrumentSuccess(t *testing.T) {
	m := mcputil.NewMetrics()
	mw := m.Instrument()

	handler := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	})

	result, err := handler(context.Background(), makeReq("test_tool"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("result should not be error")
	}

	stats := m.ToolMetrics()
	ts, ok := stats["test_tool"]
	if !ok {
		t.Fatal("no metrics for test_tool")
	}
	if ts.Calls != 1 {
		t.Errorf("calls = %d, want 1", ts.Calls)
	}
	if ts.Errors != 0 {
		t.Errorf("errors = %d, want 0", ts.Errors)
	}
	if ts.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestInstrumentGoError(t *testing.T) {
	m := mcputil.NewMetrics()
	mw := m.Instrument()

	handler := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, errors.New("database down")
	})

	result, err := handler(context.Background(), makeReq("failing_tool"))
	if err != nil {
		t.Fatalf("middleware should convert Go errors, got: %v", err)
	}
	if !result.IsError {
		t.Fatal("result should be error")
	}

	// Verify the error is structured ToolError JSON.
	text := result.Content[0].(mcp.TextContent).Text
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("error text should be JSON: %v", err)
	}
	if parsed["type"] != toolerror.ErrInternal {
		t.Errorf("type = %v, want %v", parsed["type"], toolerror.ErrInternal)
	}

	stats := m.ToolMetrics()
	if stats["failing_tool"].Errors != 1 {
		t.Errorf("errors = %d, want 1", stats["failing_tool"].Errors)
	}
}

func TestInstrumentToolError(t *testing.T) {
	m := mcputil.NewMetrics()
	mw := m.Instrument()

	handler := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, toolerror.New(toolerror.ErrNotFound, "agent not found")
	})

	result, err := handler(context.Background(), makeReq("typed_error"))
	if err != nil {
		t.Fatalf("middleware should convert errors, got: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("error text should be JSON: %v", err)
	}
	if parsed["type"] != toolerror.ErrNotFound {
		t.Errorf("type = %v, want %v (should preserve ToolError type)", parsed["type"], toolerror.ErrNotFound)
	}
}

func TestInstrumentPanic(t *testing.T) {
	m := mcputil.NewMetrics()
	mw := m.Instrument()

	handler := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		panic("nil pointer")
	})

	result, err := handler(context.Background(), makeReq("panic_tool"))
	if err != nil {
		t.Fatalf("panic should be recovered, got error: %v", err)
	}
	if !result.IsError {
		t.Fatal("result should be error after panic")
	}

	text := result.Content[0].(mcp.TextContent).Text
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("panic error should be JSON: %v", err)
	}
	if parsed["type"] != toolerror.ErrInternal {
		t.Errorf("type = %v, want %v", parsed["type"], toolerror.ErrInternal)
	}

	stats := m.ToolMetrics()
	if stats["panic_tool"].Errors != 1 {
		t.Errorf("errors = %d, want 1", stats["panic_tool"].Errors)
	}
}

func TestInstrumentPerToolMetrics(t *testing.T) {
	m := mcputil.NewMetrics()
	mw := m.Instrument()

	ok := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	})
	fail := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return nil, errors.New("fail")
	})

	ctx := context.Background()
	ok(ctx, makeReq("tool_a"))
	ok(ctx, makeReq("tool_a"))
	ok(ctx, makeReq("tool_b"))
	fail(ctx, makeReq("tool_b"))

	stats := m.ToolMetrics()
	if stats["tool_a"].Calls != 2 {
		t.Errorf("tool_a calls = %d, want 2", stats["tool_a"].Calls)
	}
	if stats["tool_a"].Errors != 0 {
		t.Errorf("tool_a errors = %d, want 0", stats["tool_a"].Errors)
	}
	if stats["tool_b"].Calls != 2 {
		t.Errorf("tool_b calls = %d, want 2", stats["tool_b"].Calls)
	}
	if stats["tool_b"].Errors != 1 {
		t.Errorf("tool_b errors = %d, want 1", stats["tool_b"].Errors)
	}
}

func TestInstrumentIsErrorResult(t *testing.T) {
	m := mcputil.NewMetrics()
	mw := m.Instrument()

	handler := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultError("bad input"), nil
	})

	result, err := handler(context.Background(), makeReq("validation_tool"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("result should be error")
	}

	stats := m.ToolMetrics()
	if stats["validation_tool"].Errors != 1 {
		t.Errorf("isError results should be counted, errors = %d, want 1", stats["validation_tool"].Errors)
	}
}

func TestInstrumentConcurrent(t *testing.T) {
	m := mcputil.NewMetrics()
	mw := m.Instrument()

	handler := mw(func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	})

	const goroutines = 50
	const callsPerGoroutine = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				handler(context.Background(), makeReq("concurrent_tool"))
			}
		}()
	}
	wg.Wait()

	stats := m.ToolMetrics()
	want := int64(goroutines * callsPerGoroutine)
	if stats["concurrent_tool"].Calls != want {
		t.Errorf("calls = %d, want %d", stats["concurrent_tool"].Calls, want)
	}
}

func TestHelpers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() (*mcp.CallToolResult, error)
		wantType string
	}{
		{"WrapError", func() (*mcp.CallToolResult, error) { return mcputil.WrapError(errors.New("oops")) }, toolerror.ErrInternal},
		{"ValidationError", func() (*mcp.CallToolResult, error) { return mcputil.ValidationError("bad %s", "input") }, toolerror.ErrValidation},
		{"NotFoundError", func() (*mcp.CallToolResult, error) { return mcputil.NotFoundError("missing") }, toolerror.ErrNotFound},
		{"ConflictError", func() (*mcp.CallToolResult, error) { return mcputil.ConflictError("locked") }, toolerror.ErrConflict},
		{"TransientError", func() (*mcp.CallToolResult, error) { return mcputil.TransientError("busy") }, toolerror.ErrTransient},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.IsError {
				t.Fatal("result should be error")
			}
			text := result.Content[0].(mcp.TextContent).Text
			var parsed map[string]any
			if err := json.Unmarshal([]byte(text), &parsed); err != nil {
				t.Fatalf("should be JSON: %v", err)
			}
			if parsed["type"] != tt.wantType {
				t.Errorf("type = %v, want %v", parsed["type"], tt.wantType)
			}
		})
	}
}

// Verify middleware satisfies the mcp-go type.
var _ server.ToolHandlerMiddleware = mcputil.NewMetrics().Instrument()
