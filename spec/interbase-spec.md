# Interbase SDK Interface Specification

Version: 2.0.0
Date: 2026-02-26

## Conventions

- **Fail-open:** Every guard function returns false/0/empty when its dependency
  is missing. Never raises an exception. Never blocks.
- **Silent no-op:** Every action function succeeds silently when its dependency
  is missing. Errors from underlying tools are logged to stderr but never
  propagated to the caller.
- **Environment variables:** Functions read from environment. They never write
  to environment (no side effects on env).

## Domain 1: Guards

### has_ic

Checks whether the `ic` (Intercore) CLI binary is available on PATH.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_has_ic` | exit code 0 (found) or 1 (missing) |
| Go | `func HasIC() bool` | true if found |
| Python | `def has_ic() -> bool` | True if found |

**Behavior:**
- Checks `$PATH` for an executable named `ic`
- Does NOT execute `ic` — only checks existence
- Returns false/1 if PATH is empty or ic is not found

### has_bd

Checks whether the `bd` (Beads) CLI binary is available on PATH.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_has_bd` | exit code 0/1 |
| Go | `func HasBD() bool` | bool |
| Python | `def has_bd() -> bool` | bool |

**Behavior:** Same as `has_ic` but for `bd` binary.

### has_companion

Checks whether a named companion plugin is installed in the Claude Code cache.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_has_companion NAME` | exit code 0/1 |
| Go | `func HasCompanion(name string) bool` | bool |
| Python | `def has_companion(name: str) -> bool` | bool |

**Behavior:**
- Scans `~/.claude/plugins/cache/*/NAME/*` for any matching directory
- Returns false if `name` is empty
- Does NOT check plugin version or health — just existence

### in_ecosystem

Returns true if the SDK was loaded from the centralized install (not a stub).

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_in_ecosystem` | exit code 0/1 |
| Go | `func InEcosystem() bool` | bool |
| Python | `def in_ecosystem() -> bool` | bool |

**Behavior:**
- Bash: checks `_INTERBASE_SOURCE == "live"`
- Go/Python: checks that the centralized install exists at
  `~/.intermod/interbase/interbase.sh` (file existence, not sourcing)

### get_bead

Returns the current bead ID from the environment.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_get_bead` | stdout (bead ID or empty) |
| Go | `func GetBead() string` | string (empty if unset) |
| Python | `def get_bead() -> str` | str (empty if unset) |

**Behavior:** Reads `$CLAVAIN_BEAD_ID` environment variable. Returns empty
string if unset or empty.

### in_sprint

Returns true if there is an active sprint context (bead + ic run).

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_in_sprint` | exit code 0/1 |
| Go | `func InSprint() bool` | bool |
| Python | `def in_sprint() -> bool` | bool |

**Behavior:**
- Returns false if `$CLAVAIN_BEAD_ID` is empty
- Returns false if `ic` is not on PATH
- Executes `ic run current --project=.` and returns true if exit code is 0
- Stderr/stdout from `ic` are suppressed

## Domain 2: Actions

### phase_set

Sets the phase on a bead via `bd set-state`.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_phase_set BEAD PHASE [REASON]` | exit code 0 (always) |
| Go | `func PhaseSet(bead, phase string, reason ...string)` | (no return) |
| Python | `def phase_set(bead: str, phase: str, reason: str = "") -> None` | None |

**Behavior:**
- If `bd` is not on PATH: silent no-op, return success
- Executes: `bd set-state BEAD "phase=PHASE"`
- If `bd` returns non-zero: log to stderr, return success anyway
- `reason` parameter is currently unused but reserved for future use

### emit_event

Emits an event via `ic events emit`.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_emit_event RUN_ID EVENT_TYPE [PAYLOAD]` | exit code 0 (always) |
| Go | `func EmitEvent(runID, eventType string, payload ...string)` | (no return) |
| Python | `def emit_event(run_id: str, event_type: str, payload: str = "{}") -> None` | None |

**Behavior:**
- If `ic` is not on PATH: silent no-op, return success
- Executes: `ic events emit RUN_ID EVENT_TYPE --payload=PAYLOAD`
- Default payload: `"{}"` (empty JSON object)
- If `ic` returns non-zero: log to stderr, return success anyway

### session_status

Prints ecosystem status to stderr.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_session_status` | exit code 0 (always), output on stderr |
| Go | `func SessionStatus() string` | status string |
| Python | `def session_status() -> str` | status string |

**Behavior:**
- Probes `bd` and `ic` availability
- If `ic` is available, probes `ic run current --project=.` for active run
- Bash: prints `[interverse] beads=active|not-detected | ic=active|not-initialized|not-detected` to stderr
- Go/Python: returns the same formatted string (caller decides where to print)

## Domain 3: Config + Discovery

### plugin_cache_path

Returns the filesystem path to a plugin's Claude Code cache directory.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_plugin_cache_path PLUGIN` | stdout (path or empty) |
| Go | `func PluginCachePath(plugin string) string` | string |
| Python | `def plugin_cache_path(plugin: str) -> str` | str |

**Behavior:**
- Scans `~/.claude/plugins/cache/*/PLUGIN/` directories
- Returns the path to the highest-versioned directory found
- Returns empty string if plugin not found or name is empty
- Does NOT validate the directory contents

### ecosystem_root

Returns the Demarch monorepo root directory.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_ecosystem_root` | stdout (path or empty) |
| Go | `func EcosystemRoot() string` | string |
| Python | `def ecosystem_root() -> str` | str |

**Behavior:**
- Reads `$DEMARCH_ROOT` if set
- Otherwise walks up from CWD looking for a directory containing
  `sdk/interbase/` as a heuristic
- Returns empty string if not found

### nudge_companion

Suggests installing a missing companion plugin. Rate-limited.

| Language | Signature | Return |
|----------|-----------|--------|
| Bash | `ib_nudge_companion COMPANION BENEFIT [PLUGIN]` | exit code 0 (always) |
| Go | `func NudgeCompanion(companion, benefit string, plugin ...string)` | (no return) |
| Python | `def nudge_companion(companion: str, benefit: str, plugin: str = "unknown") -> None` | None |

**Behavior:**
- If `companion` is empty: silent no-op
- If companion is already installed (`has_companion`): silent no-op
- If session budget exhausted (>=2 nudges this session): silent no-op
- If companion permanently dismissed (>=3 ignores): silent no-op
- Otherwise: print `[interverse] Tip: run /plugin install COMPANION for BENEFIT.` to stderr
- Increment session counter and record ignore in durable state
- Session state: `~/.config/interverse/nudge-session-${CLAUDE_SESSION_ID}.json`
- Durable state: `~/.config/interverse/nudge-state.json`
- Atomic dedup via `mkdir` (Bash/Python) or equivalent (Go)
- Session ID sanitized: strip non-alphanumeric characters except `-` and `_`

## Domain 4: MCP Contracts (Go + Python only)

### ToolError

Structured error type for MCP tool handlers.

**Wire format (JSON):**
```json
{
  "type": "NOT_FOUND",
  "message": "agent 'fd-safety' not registered",
  "recoverable": false,
  "data": {}
}
```

**Error types:**
| Constant | Wire value | Default recoverable |
|----------|-----------|-------------------|
| ErrNotFound | `"NOT_FOUND"` | false |
| ErrConflict | `"CONFLICT"` | false |
| ErrValidation | `"VALIDATION"` | false |
| ErrPermission | `"PERMISSION"` | false |
| ErrTransient | `"TRANSIENT"` | true |
| ErrInternal | `"INTERNAL"` | false |

**API (Go):**
```go
toolerror.New(errType, format, args...) *ToolError
te.WithRecoverable(bool) *ToolError
te.WithData(map[string]any) *ToolError
te.JSON() string
te.Error() string  // "[TYPE] message"
toolerror.FromError(err) *ToolError
toolerror.Wrap(err) *ToolError
```

**API (Python):**
```python
ToolError(err_type, message, **data)
te.with_recoverable(bool) -> ToolError
te.with_data(**kwargs) -> ToolError
te.json() -> str
str(te) -> "[TYPE] message"
ToolError.from_error(exc) -> ToolError | None
ToolError.wrap(exc) -> ToolError
```

### Metrics Middleware

Handler middleware for MCP tool handlers providing timing, error counting,
error wrapping, and panic/exception recovery.

**Go:** `mcputil.NewMetrics()` + `metrics.Instrument()` -> `server.ToolHandlerMiddleware`
**Python:** `McpMetrics()` + `metrics.instrument()` -> decorator/middleware callable
