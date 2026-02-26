# interbase

Shared SDK for dual-mode plugins (standalone + ecosystem). Bash SDK for hooks, Go SDK for MCP servers, Python SDK for hooks + scripts. See `AGENTS.md` for full reference.

## Quick Commands

```bash
# Install Bash SDK
bash sdk/interbase/install.sh

# Run Bash tests
bash tests/test-guards.sh    # 16 tests
bash tests/test-nudge.sh     # 4 tests
bash tests/test-config.sh    # 5 tests

# Run Go tests
cd go && go test ./...

# Run Python tests
cd python && uv run pytest tests/ -v

# Run conformance tests (all languages)
bash tests/runners/run_bash.sh
bash tests/runners/run_go.sh
bash tests/runners/run_python.sh

# Dev testing with override
INTERMOD_LIB=/path/to/dev/interbase.sh bash your-hook.sh

# Simulate standalone mode
INTERMOD_LIB=/nonexistent bash your-hook.sh
```

## Go SDK (`go/`)

Shared Go packages for Demarch MCP servers. Import via `github.com/mistakeknot/interbase`.

- **Root package** — Guards (`HasIC`, `HasBD`, `HasCompanion`, `InEcosystem`, `GetBead`, `InSprint`), Actions (`PhaseSet`, `EmitEvent`, `SessionStatus`), Config (`PluginCachePath`, `EcosystemRoot`, `NudgeCompanion`). All fail-open.
- **`toolerror`** — Structured error contract for MCP tool handlers. Types: `NOT_FOUND`, `CONFLICT`, `VALIDATION`, `PERMISSION`, `TRANSIENT`, `INTERNAL`. Use `replace` directive in consumer go.mod: `replace github.com/mistakeknot/interbase => ../../sdk/interbase/go`
- **`mcputil`** — MCP tool handler middleware: timing metrics, error counting, panic recovery, structured error wrapping. Use `metrics.Instrument()` with `server.WithToolHandlerMiddleware()`. Also provides convenience helpers: `ValidationError()`, `NotFoundError()`, `ConflictError()`, `TransientError()`, `WrapError()`

## Python SDK (`python/`)

Shared Python package for Demarch hooks and scripts. Install via `uv pip install -e sdk/interbase/python`.

- **Guards** — `has_ic()`, `has_bd()`, `has_companion()`, `in_ecosystem()`, `get_bead()`, `in_sprint()`. All return False when deps missing.
- **Actions** — `phase_set()`, `emit_event()`, `session_status()`. All silent no-ops without deps.
- **Config** — `plugin_cache_path()`, `ecosystem_root()`, `nudge_companion()`.
- **`toolerror`** — `ToolError` exception with wire-format parity to Go. 6 error types, JSON serialization.
- **`mcputil`** — `McpMetrics` with `instrument()` for handler wrapping.

## Conformance Tests (`tests/conformance/`)

YAML-defined test cases run by thin per-language runners. Ensures Bash, Go, and Python stay in sync. MCP tests excluded for Bash.

## Design Decisions (Do Not Re-Ask)

- Guards are fail-open — `ib_has_ic`, `HasIC()`, `has_ic()` return false when tools are missing
- Actions are no-op without dependencies — safe to call in standalone mode
- Installed to `~/.intermod/interbase/` via atomic tmp+mv
- Stub ships inside each plugin's `hooks/` dir; live copy overrides at source time
- Stub must NOT set `_INTERBASE_LOADED` before sourcing live copy (would skip function definitions)
- Nudge protocol: max 2/session, auto-dismissed after 3 ignores
- interbase = code sharing; interband = data sharing
- Go actions return nothing (not error) — avoids dead error-check code at call sites
- Python uses `shutil.which()` for PATH checks, `subprocess.run()` with timeouts for CLI calls
- Session ID sanitized via regex before use in filenames
