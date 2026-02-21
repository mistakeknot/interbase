# interbase — Vision and Philosophy

## What interbase Is

interbase is the shared integration SDK that lets Interverse plugins work in two modes: standalone (installed from the Claude Code marketplace, no ecosystem) and integrated (running inside Clavain with beads, intercore, and companion plugins). A single `source` call gives every plugin access to guards, phase tracking, event emission, and companion discovery — or graceful no-ops when those systems aren't present.

The SDK has one job: make dual-mode plugins trivial to write. Plugin authors call `ib_*` functions without caring whether the ecosystem is there. The SDK resolves that question at load time and makes the right thing happen.

## Core Convictions

### 1. Fail-open, always

Every `ib_*` function degrades to a safe default when its dependency is missing. `ib_phase_set` without `bd`? No-op. `ib_emit_event` without `ic`? No-op. `ib_has_companion` without the companion? Returns false. No function ever throws, exits, or prints an error in standalone mode. The standalone user's experience must be indistinguishable from "this plugin doesn't know about the ecosystem."

### 2. No hard dependencies

interbase is never in a plugin's `package.json`, `pyproject.toml`, or install manifest. Plugins ship a stub (`interbase-stub.sh`) that defines inline no-ops. The live copy at `~/.intermod/interbase/` upgrades those no-ops to real implementations — but the stub alone is sufficient for standalone operation.

### 3. Stub-first design

The stub is the plugin's ground truth. It must work without the SDK installed, without network access, without any state directory. The live copy is a progressive enhancement, not a requirement. This means the stub's API surface is the contract — the live copy must be a superset that never breaks stub callers.

### 4. Nudge, don't require

When a standalone plugin would benefit from a companion it doesn't have, interbase suggests — once or twice, then stops. The nudge protocol has a session budget (max 2), durable dismissal (3 ignores = permanent), and atomic dedup. It never blocks, never repeats, and never makes the user feel pressured. Discovery is a service, not a sales pitch.

## Scope

interbase explicitly covers:

- **Guards**: Capability detection (`ib_has_ic`, `ib_has_bd`, `ib_has_companion`, `ib_in_ecosystem`, `ib_in_sprint`)
- **Phase tracking**: Setting sprint phase via beads (`ib_phase_set`)
- **Event emission**: Structured events to intercore runs (`ib_emit_event`)
- **Session status**: Ecosystem awareness diagnostic (`ib_session_status`)
- **Companion nudging**: Discovery protocol for missing companions (`ib_nudge_companion`)
- **Stub template**: Ready-to-copy fallback for new plugin adoptions

interbase explicitly does **not** cover:

- **Plugin framework**: It's not a build system, project scaffolder, or plugin generator. Use `interdev` for that.
- **Data sharing**: Cross-plugin state and channels are `interband`'s job. interbase shares code, interband shares data.
- **Runtime orchestration**: Dispatch, scheduling, and multi-agent coordination belong to Clavain and `intermute`.
- **Version management**: Plugin publishing is `interbump`/`interpub` territory.

## Where It Fits

```
Plugin author writes code
    ↓
interbase-stub.sh (shipped in plugin)
    ↓ sources (if present)
~/.intermod/interbase/interbase.sh (installed SDK)
    ↓ calls
bd, ic, companion cache (ecosystem tools)
```

The resolution pattern mirrors `interband`: environment override (`INTERMOD_LIB`) takes precedence over the installed copy, which takes precedence over the stub's inline no-ops. Three tiers, same order everywhere.

## What interbase Is Not

- **Not a framework.** It's a bag of functions. No lifecycle hooks, no plugin registration, no mandatory init. Source it and call what you need.
- **Not a build artifact.** It's plain Bash, installed by copying a file. No compilation, no transpilation, no package manager.
- **Not coupled to Clavain.** Any tool that puts `bd` or `ic` on PATH makes interbase's guards light up. The SDK detects capabilities, not brands.

## 2026-02-20 Context

Version `1.0.0` shipped as the implementation of the dual-mode plugin architecture (bead `iv-gcu2`). interflux is the first plugin to adopt, serving as the reference implementation. The SDK lives at `sdk/interbase/` in the Interverse monorepo, signaling its role as a developer-facing library rather than internal infrastructure.

Current focus: expanding the adopter base beyond interflux, extending the API surface for gate participation and dispatch tracking, and writing a plugin author getting-started guide.
