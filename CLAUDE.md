# interbase

Shared Bash SDK for dual-mode plugins (standalone + ecosystem). See `AGENTS.md` for full function reference, stub pattern, and nudge protocol.

## Quick Commands

```bash
# Install SDK
bash sdk/interbase/install.sh

# Run tests
bash tests/test-guards.sh    # 16 tests
bash tests/test-nudge.sh     # 4 tests

# Dev testing with override
INTERMOD_LIB=/path/to/dev/interbase.sh bash your-hook.sh

# Simulate standalone mode
INTERMOD_LIB=/nonexistent bash your-hook.sh
```

## Design Decisions (Do Not Re-Ask)

- Guards are fail-open — `ib_has_ic`, `ib_has_bd` return 1 (not error) when tools are missing
- Actions are no-op without dependencies — safe to call in standalone mode
- Installed to `~/.intermod/interbase/` via atomic tmp+mv
- Stub ships inside each plugin's `hooks/` dir; live copy overrides at source time
- Stub must NOT set `_INTERBASE_LOADED` before sourcing live copy (would skip function definitions)
- Nudge protocol: max 2/session, auto-dismissed after 3 ignores
- interbase = code sharing; interband = data sharing
