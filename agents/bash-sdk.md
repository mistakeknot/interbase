# Bash SDK API

## Guards (fail-open)
| Function | Signature | Behavior |
|----------|-----------|----------|
| `ib_has_ic` | `()` | Returns 0 if `ic` CLI is on PATH |
| `ib_has_bd` | `()` | Returns 0 if `bd` CLI is on PATH |
| `ib_has_companion` | `(name)` | Returns 0 if plugin `name` is in Claude Code cache |
| `ib_in_ecosystem` | `()` | Returns 0 if sourced via live copy (not stub) |
| `ib_get_bead` | `()` | Echoes `$CLAVAIN_BEAD_ID` or empty string |
| `ib_in_sprint` | `()` | Returns 0 if bead context + active ic run |

## Actions (no-op without dependencies)
| Function | Signature | Behavior |
|----------|-----------|----------|
| `ib_phase_set` | `(bead, phase, [reason])` | Sets phase via `bd set-state` (no-op without bd) |
| `ib_emit_event` | `(run_id, event_type, [payload])` | Emits via `ic events emit` (no-op without ic) |
| `ib_session_status` | `()` | Prints `[interverse] beads=... \| ic=...` to stderr |
| `ib_nudge_companion` | `(companion, benefit, [plugin])` | Suggests missing companion install (max 2/session) |

## Internal helpers (prefixed `_ib_`)
- `_ib_nudge_state_dir`, `_ib_nudge_state_file`, `_ib_nudge_session_file` — path helpers
- `_ib_nudge_session_count`, `_ib_nudge_session_increment` — session budget tracking
- `_ib_nudge_is_dismissed`, `_ib_nudge_record` — durable dismissal (3 ignores = dismissed)
