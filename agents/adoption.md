# Adoption & Installation

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
