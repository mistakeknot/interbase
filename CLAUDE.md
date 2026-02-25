# interbase

Shared SDK for dual-mode plugins (standalone + ecosystem). Bash SDK for hooks + Go SDK for MCP servers. See `AGENTS.md` for full reference.

## Quick Commands

```bash
# Install Bash SDK
bash sdk/interbase/install.sh

# Run Bash tests
bash tests/test-guards.sh    # 16 tests
bash tests/test-nudge.sh     # 4 tests

# Run Go tests
cd go && go test ./...

# Dev testing with override
INTERMOD_LIB=/path/to/dev/interbase.sh bash your-hook.sh

# Simulate standalone mode
INTERMOD_LIB=/nonexistent bash your-hook.sh
```

## Go SDK (`go/`)

Shared Go packages for Demarch MCP servers. Import via `github.com/mistakeknot/interbase`.

- **`toolerror`** — Structured error contract for MCP tool handlers. Types: `NOT_FOUND`, `CONFLICT`, `VALIDATION`, `PERMISSION`, `TRANSIENT`, `INTERNAL`. Use `replace` directive in consumer go.mod: `replace github.com/mistakeknot/interbase => ../../sdk/interbase/go`
- **`mcputil`** — MCP tool handler middleware: timing metrics, error counting, panic recovery, structured error wrapping. Use `metrics.Instrument()` with `server.WithToolHandlerMiddleware()`. Also provides convenience helpers: `ValidationError()`, `NotFoundError()`, `ConflictError()`, `TransientError()`, `WrapError()`

## Design Decisions (Do Not Re-Ask)

- Guards are fail-open — `ib_has_ic`, `ib_has_bd` return 1 (not error) when tools are missing
- Actions are no-op without dependencies — safe to call in standalone mode
- Installed to `~/.intermod/interbase/` via atomic tmp+mv
- Stub ships inside each plugin's `hooks/` dir; live copy overrides at source time
- Stub must NOT set `_INTERBASE_LOADED` before sourcing live copy (would skip function definitions)
- Nudge protocol: max 2/session, auto-dismissed after 3 ignores
- interbase = code sharing; interband = data sharing
