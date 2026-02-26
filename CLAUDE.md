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

## SDK APIs

Three SDKs (Bash, Go, Python) with identical semantics — Guards (fail-open), Actions (no-op without deps), Config. See [AGENTS.md](./AGENTS.md) for full API reference per language.

- **Go:** `import github.com/mistakeknot/interbase` — uses `replace` directive in consumer go.mod
- **Python:** `uv pip install -e sdk/interbase/python`
- **Conformance tests:** `tests/conformance/` — YAML-defined, run by per-language runners

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
