# interbase — Shared Integration SDK

Multi-language SDK enabling Interverse plugins to work in both standalone (Claude Code marketplace) and integrated (Clavain/Intercore ecosystem) modes. Bash SDK for hooks, Go SDK for MCP servers, Python SDK for hooks + scripts.

## File Structure

```
sdk/interbase/
  lib/
    interbase.sh    — core Bash SDK (installed to ~/.intermod/interbase/)
    VERSION         — semver for installed copy (currently 2.0.0)
  templates/
    interbase-stub.sh   — shipped inside each plugin
    integration.json    — schema template for plugin integration manifests
  tests/
    test-guards.sh      — guard function + stub fallback tests (16 assertions)
    test-nudge.sh       — nudge protocol tests (4 assertions)
    test-config.sh      — config function tests (5 assertions)
    conformance/
      guards.yaml       — cross-language guard test cases
      actions.yaml      — cross-language action test cases
      config.yaml       — cross-language config test cases
      mcp.yaml          — MCP contract tests (Go + Python only)
    runners/
      run_bash.sh       — conformance runner for Bash
      run_go.sh         — conformance runner for Go
      run_python.sh     — conformance runner for Python
  go/
    go.mod              — Go module: github.com/mistakeknot/interbase (Go 1.23, mcp-go v0.43.2)
    interbase.go        — root package: guards, actions, config/discovery
    interbase_test.go   — 19 tests
    conformance_test.go — YAML-driven conformance tests
    README.md           — standalone Go SDK reference
    toolerror/
      toolerror.go      — structured error contract for MCP servers
      toolerror_test.go — 9 tests
    mcputil/
      instrument.go      — tool handler middleware (timing, errors, panics, metrics)
      instrument_test.go — 8 tests
  python/
    pyproject.toml      — Python package config (hatchling, requires-python >=3.11)
    interbase/
      __init__.py       — public API re-exports, __version__ = "2.0.0"
      guards.py         — fail-open guard functions
      actions.py        — silent no-op action functions
      config.py         — config + discovery functions
      nudge.py          — companion nudge protocol
      toolerror.py      — structured MCP error contract (wire-compatible with Go)
      mcputil.py        — MCP metrics middleware
    tests/
      test_guards.py    — guard function tests
      test_actions.py   — action function tests
      test_config.py    — config function tests
      test_toolerror.py — toolerror tests
      test_mcputil.py   — metrics middleware tests
      test_conformance.py — YAML-driven conformance tests
  install.sh            — deploy Bash SDK to ~/.intermod/interbase/
```

## Function Reference

### Guards (fail-open)
| Function | Signature | Behavior |
|----------|-----------|----------|
| `ib_has_ic` | `()` | Returns 0 if `ic` CLI is on PATH |
| `ib_has_bd` | `()` | Returns 0 if `bd` CLI is on PATH |
| `ib_has_companion` | `(name)` | Returns 0 if plugin `name` is in Claude Code cache |
| `ib_in_ecosystem` | `()` | Returns 0 if sourced via live copy (not stub) |
| `ib_get_bead` | `()` | Echoes `$CLAVAIN_BEAD_ID` or empty string |
| `ib_in_sprint` | `()` | Returns 0 if bead context + active ic run |

### Actions (no-op without dependencies)
| Function | Signature | Behavior |
|----------|-----------|----------|
| `ib_phase_set` | `(bead, phase, [reason])` | Sets phase via `bd set-state` (no-op without bd) |
| `ib_emit_event` | `(run_id, event_type, [payload])` | Emits via `ic events emit` (no-op without ic) |
| `ib_session_status` | `()` | Prints `[interverse] beads=... \| ic=...` to stderr |
| `ib_nudge_companion` | `(companion, benefit, [plugin])` | Suggests missing companion install (max 2/session) |

### Internal helpers (prefixed `_ib_`)
- `_ib_nudge_state_dir`, `_ib_nudge_state_file`, `_ib_nudge_session_file` — path helpers
- `_ib_nudge_session_count`, `_ib_nudge_session_increment` — session budget tracking
- `_ib_nudge_is_dismissed`, `_ib_nudge_record` — durable dismissal (3 ignores = dismissed)

## Go SDK

Shared Go packages for Demarch MCP servers. Module: `github.com/mistakeknot/interbase`.

### Root Package — Guards, Actions, Config

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

### toolerror — Structured MCP Error Contract

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

### mcputil — MCP Tool Handler Middleware

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

### Test Commands

```bash
cd go && go test ./...   # 19 root + 9 toolerror + 8 mcputil + conformance
```

## Python SDK

Shared Python package for Demarch hooks and scripts. Module: `interbase` (install via `uv pip install -e sdk/interbase/python`).

### Guards (fail-open)
| Function | Signature | Behavior |
|----------|-----------|----------|
| `has_ic` | `() -> bool` | Returns True if `ic` CLI is on PATH (via `shutil.which`) |
| `has_bd` | `() -> bool` | Returns True if `bd` CLI is on PATH |
| `has_companion` | `(name: str) -> bool` | Returns True if plugin is in Claude Code cache |
| `in_ecosystem` | `() -> bool` | Returns True if centralized interbase install exists |
| `get_bead` | `() -> str` | Returns `$CLAVAIN_BEAD_ID` or empty string |
| `in_sprint` | `() -> bool` | Returns True if bead context + active ic run |

### Actions (no-op without dependencies)
| Function | Signature | Behavior |
|----------|-----------|----------|
| `phase_set` | `(bead: str, phase: str, reason: str = "")` | Sets phase via `bd set-state` (no-op without bd) |
| `emit_event` | `(run_id: str, event_type: str, payload: str = "{}")` | Emits via `ic events emit` (no-op without ic) |
| `session_status` | `() -> str` | Returns `[interverse] beads=... | ic=...` |
| `nudge_companion` | `(companion: str, benefit: str, plugin: str = "unknown")` | Suggests missing companion install (max 2/session) |

### Config + Discovery
| Function | Signature | Behavior |
|----------|-----------|----------|
| `plugin_cache_path` | `(plugin: str) -> str` | Returns highest-versioned cache path, or empty |
| `ecosystem_root` | `() -> str` | Returns monorepo root via `$DEMARCH_ROOT` or walk-up |

### toolerror — Structured MCP Error Contract

Wire-format compatible with Go's `toolerror.ToolError`. Same 6 error types, same JSON serialization.

**Error type constants:** `ERR_NOT_FOUND`, `ERR_CONFLICT`, `ERR_VALIDATION`, `ERR_PERMISSION`, `ERR_TRANSIENT`, `ERR_INTERNAL`

**Usage:**
```python
from interbase.toolerror import ToolError, ERR_NOT_FOUND, ERR_TRANSIENT

# Create
te = ToolError(ERR_NOT_FOUND, "agent not found")

# With metadata
te = ToolError(ERR_NOT_FOUND, "gone").with_data(file="main.go")

# Override recoverable
te = ToolError(ERR_NOT_FOUND, "gone").with_recoverable(True)

# Serialize to wire format
json_str = te.json()  # matches Go's encoding/json output

# Convert any exception
te = ToolError.wrap(exc)        # passthrough if already ToolError, else ERR_INTERNAL
te = ToolError.from_error(exc)  # returns None if not a ToolError
```

### mcputil — MCP Metrics Middleware

**Usage:**
```python
from interbase.mcputil import McpMetrics

metrics = McpMetrics()
wrapped = metrics.instrument("tool_name", original_handler)

# Read metrics snapshot
for name, stats in metrics.tool_metrics().items():
    print(f"{name}: {stats}")  # "calls=N errors=N duration=Xs"
```

### Test Commands

```bash
cd python && uv run pytest tests/ -v
```

### Install

```bash
uv pip install -e sdk/interbase/python            # editable install
uv pip install -e "sdk/interbase/python[test]"     # with test deps (pytest, pyyaml)
```

## Conformance Tests

YAML-defined test cases in `tests/conformance/` ensure all three SDKs (Bash, Go, Python) implement the same behavior. Each YAML file defines a `domain` (guards, actions, config, mcp) and a list of test cases with setup, call, args, and expected results.

### Domains

| File | Tests | Languages |
|------|-------|-----------|
| `guards.yaml` | 9 cases | Bash, Go, Python |
| `actions.yaml` | 3 cases | Bash, Go, Python |
| `config.yaml` | 3 cases | Bash, Go, Python |
| `mcp.yaml` | 6 cases | Go, Python (Bash excluded — hooks don't run MCP servers) |

### Running Conformance Tests

```bash
bash tests/runners/run_bash.sh     # parses YAML, calls ib_* functions
bash tests/runners/run_go.sh       # runs go test with conformance_test.go
bash tests/runners/run_python.sh   # runs pytest test_conformance.py
```

### Adding New Conformance Tests

1. Add a test case to the appropriate YAML file in `tests/conformance/`
2. If the test calls a new function, add the function dispatch to all three runners
3. Run all three runners to verify cross-language consistency

## Install (Bash SDK)

```bash
bash sdk/interbase/install.sh
# Installs to ~/.intermod/interbase/interbase.sh
```

Uses atomic temp-then-mv to prevent partial reads by concurrent hooks.

## Plugin Adoption Pattern

1. Copy `templates/interbase-stub.sh` into plugin's `hooks/` directory
2. Create `integration.json` in `.claude-plugin/` using `templates/integration.json` as schema
3. Source the stub in session-start hook
4. Call `ib_*` functions — they're no-ops in standalone, functional in ecosystem

## Dev Testing

Override the live copy path without modifying `~/.intermod/`:
```bash
INTERMOD_LIB=/path/to/dev/interbase.sh bash your-hook.sh
```

Simulate standalone mode:
```bash
INTERMOD_LIB=/nonexistent bash your-hook.sh
```

## Test Commands

```bash
# Bash unit tests
bash sdk/interbase/tests/test-guards.sh   # 16 tests
bash sdk/interbase/tests/test-nudge.sh     # 4 tests
bash sdk/interbase/tests/test-config.sh    # 5 tests

# Go unit tests
cd sdk/interbase/go && go test ./...

# Python unit tests
cd sdk/interbase/python && uv run pytest tests/ -v

# Conformance tests (all languages)
bash sdk/interbase/tests/runners/run_bash.sh
bash sdk/interbase/tests/runners/run_go.sh
bash sdk/interbase/tests/runners/run_python.sh
```

## Adopters

### Bash SDK
| Plugin | Functions used |
|--------|---------------|
| interflux | `ib_session_status` |
| intermem | `ib_session_status`, `ib_nudge_companion` |
| intersynth | `ib_session_status`, `ib_nudge_companion` |
| interline | `ib_session_status` |

All four ship `interbase-stub.sh` in their `hooks/` directory and source it from `session-start.sh`.

### Go SDK
| Module | Scope |
|--------|-------|
| interlock | All 12 MCP tool handlers (toolerror + mcputil middleware + helpers) |
| intermap | 9 MCP tool handlers (toolerror + mcputil middleware + helpers) |

## Load Guard Pattern

The stub must NOT set `_INTERBASE_LOADED` before attempting to source the live copy — otherwise the live copy's own guard would short-circuit, skipping all function definitions. The guard is only set in the fallback (stub) path.

## Nudge Protocol

- Max 2 nudges per session (tracked in `~/.config/interverse/nudge-session-*.json`)
- After 3 ignores of the same companion, permanently dismissed
- Durable state in `~/.config/interverse/nudge-state.json`
- Atomic dedup via `mkdir` (prevents parallel duplicate nudges)
- Only fires from live copy; stubs have a no-op implementation

## Relationship to interband

interband (`core/interband/`) provides data sharing between plugins (key-value state, channels). interbase provides code sharing (SDK functions). Both use the same resolution pattern: env override → monorepo path → home directory fallback.
