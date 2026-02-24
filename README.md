# interbase

Shared Bash SDK for Interverse plugins. Enables dual-mode operation: plugins work standalone via the Claude Code marketplace and gain additional features when the Interverse ecosystem is present.

## What This Does

Interverse plugins need to call shared tools (Beads for tracking, Intercore for coordination) without hard-depending on them. interbase provides a stub pattern: each plugin ships a thin `interbase-stub.sh` that checks for the centralized SDK at `~/.intermod/interbase/interbase.sh`. If found, the full SDK loads. If not, inline no-op stubs activate — every `ib_*` function returns a safe default, and the plugin works standalone.

This means plugin authors call `ib_has_bd`, `ib_register`, or `ib_nudge` without guarding against missing dependencies. Guards are fail-open by design.

## Who This Is For

Plugin authors building Interverse-compatible Claude Code plugins. End users don't interact with interbase directly — it installs as a shared library that plugins source automatically.

## Install

```bash
bash install.sh
```

Installs to `~/.intermod/interbase/` via atomic tmp+mv.

## For Plugin Authors

See `AGENTS.md` for the full function reference, stub pattern, and adoption guide.
