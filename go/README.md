# interbase Go SDK

Shared Go packages for Demarch MCP servers.

```
module github.com/mistakeknot/interbase
```

## Packages

### toolerror

Structured error contract for MCP tool handlers. Replaces flat `fmt.Errorf` strings with typed, machine-parseable JSON that agents can act on.

```go
import "github.com/mistakeknot/interbase/toolerror"
```

#### Error types

| Constant | Wire value | Recoverable | When to use |
|----------|-----------|-------------|-------------|
| `ErrNotFound` | `NOT_FOUND` | false | Resource doesn't exist |
| `ErrConflict` | `CONFLICT` | false | Concurrent modification (e.g. reservation held by another agent) |
| `ErrValidation` | `VALIDATION` | false | Invalid input or arguments |
| `ErrPermission` | `PERMISSION` | false | Access denied |
| `ErrTransient` | `TRANSIENT` | true | Temporary failure, safe to retry |
| `ErrInternal` | `INTERNAL` | false | Unexpected server error |

#### Creating errors

```go
// Basic — type + formatted message
te := toolerror.New(toolerror.ErrNotFound, "agent %q not registered", agentName)

// With metadata
te := toolerror.New(toolerror.ErrConflict, "file locked").
    WithRecoverable(true).
    WithData(map[string]any{"holder": "agent-2", "file": "main.go"})

// Convert any error to ToolError (passthrough if already ToolError, else ErrInternal)
te := toolerror.Wrap(err)

// Extract ToolError from error chain (returns nil if not a ToolError)
te := toolerror.FromError(err)
```

#### Using in MCP handlers

The `.JSON()` method serializes to the wire format that agents parse:

```go
func myHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    result, err := doSomething()
    if err != nil {
        return mcp.NewToolResultError(toolerror.Wrap(err).JSON()), nil
    }
    return jsonResult(result)
}
```

For HTTP client wrappers, build a mapping function (see interlock's `toToolError()` for the reference implementation):

```go
func toToolError(err error) *mcp.CallToolResult {
    var ce *client.ConflictError
    if errors.As(err, &ce) {
        te := toolerror.New(toolerror.ErrConflict, "%v", ce).WithRecoverable(true)
        return mcp.NewToolResultError(te.JSON())
    }
    // ... map other domain errors to appropriate types
    return mcp.NewToolResultError(toolerror.Wrap(err).JSON())
}
```

#### The error interface

`ToolError` implements Go's `error` interface, so it works with `errors.As`, `errors.Is`, `fmt.Errorf("%w", te)`, and standard error chains.

```go
te := toolerror.New(toolerror.ErrTransient, "database busy")
fmt.Println(te.Error()) // "[TRANSIENT] database busy"

wrapped := fmt.Errorf("handler failed: %w", te)
recovered := toolerror.FromError(wrapped) // non-nil, Type == "TRANSIENT"
```

### mcputil

MCP tool handler middleware. Wraps mcp-go handlers with timing, error counting, panic recovery, and structured error conversion.

```go
import "github.com/mistakeknot/interbase/mcputil"
```

#### Middleware

Register with `server.WithToolHandlerMiddleware()` to instrument all tool calls:

```go
metrics := mcputil.NewMetrics()
s := server.NewMCPServer("myserver", "0.1.0",
    server.WithToolHandlerMiddleware(metrics.Instrument()),
)

// Later — read per-tool metrics:
for name, stats := range metrics.ToolMetrics() {
    fmt.Printf("%s: %d calls, %d errors, %s total\n", name, stats.Calls, stats.Errors, stats.Duration)
}
```

The middleware applies to every handler automatically:
- **Timing**: accumulates `time.Since(start)` per tool (atomic nanosecond counter)
- **Error wrapping**: if handler returns `(nil, error)`, converts to `(ToolError.JSON(), nil)` — agents always see structured errors
- **Error counting**: increments per-tool counter on both Go errors and `isError` results
- **Panic recovery**: catches panics, returns `ErrInternal` ToolError

#### Convenience helpers

Replace verbose `mcp.NewToolResultError(toolerror.New(...).JSON()), nil` with one-liners:

```go
return mcputil.ValidationError("field %q is required", name)
return mcputil.NotFoundError("agent %q not found", id)
return mcputil.ConflictError("file already reserved")
return mcputil.TransientError("service unavailable: %v", err)
return mcputil.WrapError(err)  // wraps any error as ErrInternal
```

All helpers return `(*mcp.CallToolResult, error)` — the standard MCP handler return signature.

## Setup

Since interbase lives in the monorepo and isn't published to a Go module proxy, consumers use a `replace` directive:

```
// go.mod
require github.com/mistakeknot/interbase v0.0.0

replace github.com/mistakeknot/interbase => ../../sdk/interbase/go
```

Adjust the relative path based on your module's location in the monorepo.

## Tests

```bash
go test ./...
```

## Adopters

| Module | Scope |
|--------|-------|
| interlock | All 12 MCP tool handlers (toolerror + mcputil middleware + helpers) |

## Design

See [docs/sdk-toolerror.md](../../docs/sdk-toolerror.md) for the design rationale, wire format specification, and agent-side parsing guidance.
