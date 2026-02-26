"""Companion nudge protocol — rate-limited install suggestions."""

from __future__ import annotations

import json
import os
import re
import sys
from pathlib import Path

from interbase.guards import has_companion


def nudge_companion(
    companion: str, benefit: str, plugin: str = "unknown"
) -> None:
    """Suggest installing a missing companion. Silent no-op if rate-limited."""
    if not companion:
        return
    if has_companion(companion):
        return

    # Sanitize session ID for safe filenames
    sid = re.sub(r"[^a-zA-Z0-9_-]", "", os.environ.get("CLAUDE_SESSION_ID", "unknown"))
    state_dir = Path(
        os.environ.get("XDG_CONFIG_HOME", os.path.expanduser("~/.config"))
    ) / "interverse"
    session_file = state_dir / f"nudge-session-{sid}.json"
    state_file = state_dir / "nudge-state.json"

    # Session budget
    count = _read_session_count(session_file)
    if count >= 2:
        return

    # Durable dismissal
    if _is_dismissed(state_file, plugin, companion):
        return

    # Atomic dedup via mkdir — matches Bash pattern. First caller wins.
    state_dir.mkdir(parents=True, exist_ok=True)
    flag = state_dir / f".nudge-{sid}-{plugin}-{companion}"
    try:
        flag.mkdir()  # atomic: fails if already exists
    except FileExistsError:
        return  # another hook already emitted this nudge

    # Emit nudge
    print(
        f"[interverse] Tip: run /plugin install {companion} for {benefit}.",
        file=sys.stderr,
    )

    # Record state
    _write_session_count(session_file, count + 1)
    _record_nudge(state_file, plugin, companion)


def _read_session_count(path: Path) -> int:
    try:
        data = json.loads(path.read_text())
        return int(data.get("count", 0))
    except (FileNotFoundError, json.JSONDecodeError, ValueError):
        return 0


def _write_session_count(path: Path, count: int) -> None:
    try:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps({"count": count}))
    except OSError:
        pass


def _is_dismissed(state_file: Path, plugin: str, companion: str) -> bool:
    try:
        data = json.loads(state_file.read_text())
        key = f"{plugin}:{companion}"
        entry = data.get(key, {})
        return entry.get("dismissed", False) is True
    except (FileNotFoundError, json.JSONDecodeError):
        return False


def _record_nudge(state_file: Path, plugin: str, companion: str) -> None:
    try:
        state_file.parent.mkdir(parents=True, exist_ok=True)
        key = f"{plugin}:{companion}"
        try:
            data = json.loads(state_file.read_text())
        except (FileNotFoundError, json.JSONDecodeError):
            data = {}
        entry = data.get(key, {"ignores": 0, "dismissed": False})
        entry["ignores"] = entry.get("ignores", 0) + 1
        if entry["ignores"] >= 3:
            entry["dismissed"] = True
        data[key] = entry
        state_file.write_text(json.dumps(data))
    except OSError:
        pass
