# Go SDK

Shared Go packages for Demarch MCP servers. Module: `github.com/mistakeknot/interbase`.

## Root Package — Guards, Actions, Config

The root `interbase` package provides the same guard/action/config surface as Bash and Python, but for Go consumers. All guards are fail-open, all actions are silent no-ops without dependencies.

**Guards:**
| Function | Signature | Behavior |
|----------|-----------|----------|
| `HasIC` | `() bool` | Returns true if `ic` CLI is on PATH |
| `HasBD` | `() bool` | Returns true if `bd` CLI is on PATH |
| `HasCompanion` | `(name string) bool` | Returns true if plugin is in Claude Code cache |
| `InEcosystem` | `() bool` | Returns true if centralized interbase install exists |
| `GetBead` | `() string` | Returns `$CLAVAIN_BEAD_ID` or empty string |
| `InSprint` | `() bool` | Returns true if bead context + active ic run |

**Actions:**
| Function | Signature | Behavior |
|----------|-----------|----------|
| `PhaseSet` | `(bead, phase string, reason ...string)` | Sets phase via `bd set-state` (no-op without bd) |
| `EmitEvent` | `(runID, eventType string, payload ...string)` | Emits via `ic events emit` (no-op without ic) |
| `SessionStatus` | `() string` | Returns `[interverse] beads=... | ic=...` |
| `NudgeCompanion` | `(companion, benefit string, plugin ...string)` | Suggests missing companion install (max 2/session) |

**Config:**
| Function | Signature | Behavior |
|----------|-----------|----------|
| `PluginCachePath` | `(plugin string) string` | Returns highest-versioned cache path, or empty |
| `EcosystemRoot` | `() string` | Returns monorepo root via `$DEMARCH_ROOT` or walk-up |

**Usage:**
```go
import "github.com/mistakeknot/interbase"

if interbase.HasIC() && interbase.InSprint() {
    interbase.EmitEvent(runID, "hook.completed")
}
```

## toolerror — Structured MCP Error Contract

All Demarch MCP tool handlers should return `ToolError` instead of flat error strings, enabling agents to distinguish transient from permanent failures.

**Error types:**
| Constant | Value | Default Recoverable | Use case |
|----------|-------|---------------------|----------|
| `ErrNotFound` | `NOT_FOUND` | false | Resource doesn't exist |
| `ErrConflict` | `CONFLICT` | false | Concurrent modification |
| `ErrValidation` | `VALIDATION` | false | Invalid input/arguments |
| `ErrPermission` | `PERMISSION` | false | Access denied |
| `ErrTransient` | `TRANSIENT` | true | Temporary failure, safe to retry |
| `ErrInternal` | `INTERNAL` | false | Unexpected server error |

**Usage:**
```go
import "github.com/mistakeknot/interbase/toolerror"

// In MCP tool handler:
return mcp.NewToolResultError(toolerror.New(toolerror.ErrNotFound, "agent %q not found", name).JSON()), nil

// Convert client errors:
te := toolerror.Wrap(err)  // passthrough if already ToolError, else ErrInternal

// Add context:
toolerror.New(toolerror.ErrConflict, "version mismatch").WithRecoverable(true).WithData(map[string]any{"file": "main.go"})
```

**Consumer setup** — add to your `go.mod`:
```
require github.com/mistakeknot/interbase v0.0.0
replace github.com/mistakeknot/interbase => ../../sdk/interbase/go
```

**Adopters:** interlock (all 12 MCP tool handlers), intermap (9 MCP tool handlers)

## mcputil — MCP Tool Handler Middleware

Wraps mcp-go tool handlers with instrumentation. Register via `server.WithToolHandlerMiddleware()`.

**Features:**
- **Timing**: per-tool call duration tracking (atomic nanosecond counters)
- **Error counting**: increments per-tool error counter on Go errors and `isError` results
- **Error wrapping**: converts unhandled Go errors to structured ToolError JSON
- **Panic recovery**: catches panics, returns `ErrInternal` ToolError

**Usage:**
```go
import "github.com/mistakeknot/interbase/mcputil"

metrics := mcputil.NewMetrics()
s := server.NewMCPServer("myserver", "0.1.0",
    server.WithToolHandlerMiddleware(metrics.Instrument()),
)

// Read metrics snapshot:
stats := metrics.ToolMetrics()  // map[string]ToolStats
// ToolStats implements fmt.Stringer: "calls=N errors=N duration=Xs"
```

**Convenience helpers** (replace verbose `mcp.NewToolResultError(toolerror.New(...).JSON()), nil`):
```go
return mcputil.ValidationError("field %q is required", name)
return mcputil.NotFoundError("agent %q not found", id)
return mcputil.ConflictError("file already reserved")
return mcputil.TransientError("service unavailable")
return mcputil.WrapError(err)  // wraps any error as ErrInternal
```

**Adopters:** interlock (middleware + helpers in all 12 MCP tool handlers), intermap (middleware + helpers in 9 MCP tool handlers)

## Test Commands

```bash
cd go && go test ./...   # 19 root + 9 toolerror + 8 mcputil + conformance
```
