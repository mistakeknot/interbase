# interbase

Interverse plugins need shared tools (Beads, Intercore) without hard-depending on them. interbase is the shim: present, everything lights up; absent, everything still works.

## What this does

Each plugin ships a thin `interbase-stub.sh` that checks for the centralized SDK at `~/.intermod/interbase/interbase.sh`. If found, the full SDK loads. If not, inline no-op stubs activate and every `ib_*` function returns a safe default. Plugin authors call `ib_has_bd`, `ib_register`, or `ib_nudge` without guarding against missing dependencies. Guards are fail-open by design.

## Who this is for

Plugin authors building Interverse-compatible Claude Code plugins. End users don't interact with interbase directly; it installs as a shared library that plugins source automatically.

## Install

```bash
bash install.sh
```

Installs to `~/.intermod/interbase/` via atomic tmp+mv.

## For plugin authors

See `AGENTS.md` for the full function reference, stub pattern, and adoption guide.
