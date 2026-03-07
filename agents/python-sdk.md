# Python SDK

Shared Python package for Demarch hooks and scripts. Module: `interbase` (install via `uv pip install -e sdk/interbase/python`).

## Guards (fail-open)
| Function | Signature | Behavior |
|----------|-----------|----------|
| `has_ic` | `() -> bool` | Returns True if `ic` CLI is on PATH (via `shutil.which`) |
| `has_bd` | `() -> bool` | Returns True if `bd` CLI is on PATH |
| `has_companion` | `(name: str) -> bool` | Returns True if plugin is in Claude Code cache |
| `in_ecosystem` | `() -> bool` | Returns True if centralized interbase install exists |
| `get_bead` | `() -> str` | Returns `$CLAVAIN_BEAD_ID` or empty string |
| `in_sprint` | `() -> bool` | Returns True if bead context + active ic run |

## Actions (no-op without dependencies)
| Function | Signature | Behavior |
|----------|-----------|----------|
| `phase_set` | `(bead: str, phase: str, reason: str = "")` | Sets phase via `bd set-state` (no-op without bd) |
| `emit_event` | `(run_id: str, event_type: str, payload: str = "{}")` | Emits via `ic events emit` (no-op without ic) |
| `session_status` | `() -> str` | Returns `[interverse] beads=... | ic=...` |
| `nudge_companion` | `(companion: str, benefit: str, plugin: str = "unknown")` | Suggests missing companion install (max 2/session) |

## Config + Discovery
| Function | Signature | Behavior |
|----------|-----------|----------|
| `plugin_cache_path` | `(plugin: str) -> str` | Returns highest-versioned cache path, or empty |
| `ecosystem_root` | `() -> str` | Returns monorepo root via `$DEMARCH_ROOT` or walk-up |

## toolerror — Structured MCP Error Contract

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

## mcputil — MCP Metrics Middleware

**Usage:**
```python
from interbase.mcputil import McpMetrics

metrics = McpMetrics()
wrapped = metrics.instrument("tool_name", original_handler)

# Read metrics snapshot
for name, stats in metrics.tool_metrics().items():
    print(f"{name}: {stats}")  # "calls=N errors=N duration=Xs"
```

## Test Commands

```bash
cd python && uv run pytest tests/ -v
```

## Install

```bash
uv pip install -e sdk/interbase/python            # editable install
uv pip install -e "sdk/interbase/python[test]"     # with test deps (pytest, pyyaml)
```
