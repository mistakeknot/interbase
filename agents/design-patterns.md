# Design Patterns

## Load Guard Pattern

The stub must NOT set `_INTERBASE_LOADED` before attempting to source the live copy — otherwise the live copy's own guard would short-circuit, skipping all function definitions. The guard is only set in the fallback (stub) path.

## Nudge Protocol

- Max 2 nudges per session (tracked in `~/.config/interverse/nudge-session-*.json`)
- After 3 ignores of the same companion, permanently dismissed
- Durable state in `~/.config/interverse/nudge-state.json`
- Atomic dedup via `mkdir` (prevents parallel duplicate nudges)
- Only fires from live copy; stubs have a no-op implementation

## Relationship to interband

interband (`core/interband/`) provides data sharing between plugins (key-value state, channels). interbase provides code sharing (SDK functions). Both use the same resolution pattern: env override -> monorepo path -> home directory fallback.
