"""Action functions — silent no-ops when dependencies are absent."""

from __future__ import annotations

import subprocess
import sys

from interbase.guards import has_bd, has_ic


def phase_set(bead: str, phase: str, reason: str = "") -> None:
    """Set the phase on a bead via bd. Silent no-op without bd."""
    if not has_bd():
        return
    try:
        subprocess.run(
            ["bd", "set-state", bead, f"phase={phase}"],
            capture_output=True,
            timeout=10,
        )
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError) as exc:
        print(f"[interbase] bd set-state failed: {exc}", file=sys.stderr)


def emit_event(run_id: str, event_type: str, payload: str = "{}") -> None:
    """Emit an event via ic. Silent no-op without ic."""
    if not has_ic():
        return
    try:
        subprocess.run(
            ["ic", "events", "emit", run_id, event_type, f"--payload={payload}"],
            capture_output=True,
            timeout=10,
        )
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError) as exc:
        print(f"[interbase] ic events emit failed: {exc}", file=sys.stderr)


def session_status() -> str:
    """Return the ecosystem status string."""
    parts: list[str] = []

    if has_bd():
        parts.append("beads=active")
    else:
        parts.append("beads=not-detected")

    if has_ic():
        try:
            result = subprocess.run(
                ["ic", "run", "current", "--project=."],
                capture_output=True,
                timeout=5,
            )
            if result.returncode == 0:
                parts.append("ic=active")
            else:
                parts.append("ic=not-initialized")
        except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
            parts.append("ic=not-initialized")
    else:
        parts.append("ic=not-detected")

    return f"[interverse] {' | '.join(parts)}"
