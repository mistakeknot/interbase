# Interbase SDK Migration Guide

How to adopt interbase in existing plugins. Covers Go MCP servers, Python hooks,
and Python standalone scripts.

## Go MCP Server -- Adding toolerror + mcputil

### 1. Add the dependency

In your plugin's `go.mod`:

```
require github.com/mistakeknot/interbase v0.0.0
replace github.com/mistakeknot/interbase => ../../../sdk/interbase/go
```

Run `go mod tidy` after adding.

### 2. Add metrics middleware

```go
import "github.com/mistakeknot/interbase/mcputil"

metrics := mcputil.NewMetrics()
s := server.NewMCPServer("your-plugin", version,
    server.WithToolHandlerMiddleware(metrics.Instrument()),
)
```

This wraps every tool handler with timing, error counting, panic recovery,
and structured error wrapping -- zero changes to individual handlers needed.

### 3. Replace flat error strings

Before:
```go
return mcp.NewToolResultError("project not found"), nil
```

After:
```go
return mcputil.NotFoundError("project %q not found", name)
```

Available helpers:
- `mcputil.NotFoundError(format, args...)` -- `NOT_FOUND`
- `mcputil.ValidationError(format, args...)` -- `VALIDATION`
- `mcputil.ConflictError(format, args...)` -- `CONFLICT`
- `mcputil.TransientError(format, args...)` -- `TRANSIENT`
- `mcputil.WrapError(err)` -- wraps any error as `INTERNAL`

### 4. Add guards + config (optional)

```go
import "github.com/mistakeknot/interbase"

if interbase.HasIC() {
    interbase.EmitEvent(runID, "tool-called")
}

root := interbase.EcosystemRoot()
```

All guards return false when tools are missing. All actions are silent no-ops.

## Python Hook -- Using guards + actions

### Pattern: Thin Bash wrapper

The Bash hook file sources `interbase-stub.sh`, then delegates to Python:

```bash
#!/usr/bin/env bash
source "$(dirname "$0")/interbase-stub.sh"
python3 "$(dirname "$0")/my-hook.py" "$@"
```

### Pattern: Fail-open import

In the Python hook, use a try/except import so the hook works without the SDK:

```python
try:
    import interbase
except ImportError:
    exit(0)  # standalone mode -- exit cleanly

bead = interbase.get_bead()
if bead:
    interbase.phase_set(bead, "hook-fired")
```

### Installing the Python SDK

From the monorepo root:

```bash
cd sdk/interbase/python
uv pip install -e .
```

Or add to your plugin's dependencies if using `pyproject.toml`.

## Python Standalone Script -- Using guards + config

For scripts that aren't hooks (analysis tools, bridges, etc.):

```python
import interbase

# Guards -- check what's available
if interbase.in_ecosystem():
    root = interbase.ecosystem_root()
    cache = interbase.plugin_cache_path("intermap")

# Actions -- safe without deps
if interbase.get_bead():
    interbase.phase_set(interbase.get_bead(), "analyzing")

# Status
status = interbase.session_status()
```

## Common Patterns

### Conditional behavior based on ecosystem

```python
import interbase

if interbase.in_ecosystem() and interbase.has_companion("interlock"):
    # Rich behavior with coordination
    pass
else:
    # Standalone fallback
    pass
```

### Nudging companion installs

```python
interbase.nudge_companion(
    "interlock",
    "file coordination across agents",
    plugin="my-plugin"
)
```

Rate-limited: max 2 per session, auto-dismissed after 3 ignores.

### MCP error handling in Python

```python
from interbase.toolerror import ToolError, ERR_NOT_FOUND, ERR_VALIDATION
from interbase.mcputil import McpMetrics

metrics = McpMetrics()

@metrics.instrument("my-tool", my_handler)
def my_handler(request):
    if not found:
        raise ToolError(ERR_NOT_FOUND, f"resource {name!r} not found")
    if not valid:
        raise ToolError(ERR_VALIDATION, "invalid input")
    return result
```
