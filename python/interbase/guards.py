"""Guard functions — fail-open capability detection."""

from __future__ import annotations

import glob
import os
import shutil
import subprocess


def has_ic() -> bool:
    """Return True if the ic (Intercore) CLI is on PATH."""
    return shutil.which("ic") is not None


def has_bd() -> bool:
    """Return True if the bd (Beads) CLI is on PATH."""
    return shutil.which("bd") is not None


def has_companion(name: str) -> bool:
    """Return True if the named plugin is in the Claude Code cache."""
    if not name:
        return False
    home = os.path.expanduser("~")
    pattern = os.path.join(home, ".claude", "plugins", "cache", "*", name, "*")
    return len(glob.glob(pattern)) > 0


def in_ecosystem() -> bool:
    """Return True if the centralized interbase install exists."""
    path = os.environ.get("INTERMOD_LIB", "")
    if not path:
        home = os.path.expanduser("~")
        path = os.path.join(home, ".intermod", "interbase", "interbase.sh")
    return os.path.isfile(path)


def get_bead() -> str:
    """Return the current bead ID from $CLAVAIN_BEAD_ID, or empty string."""
    return os.environ.get("CLAVAIN_BEAD_ID", "")


def in_sprint() -> bool:
    """Return True if there is an active sprint context (bead + ic run)."""
    if not get_bead():
        return False
    if not has_ic():
        return False
    try:
        result = subprocess.run(
            ["ic", "run", "current", "--project=."],
            capture_output=True,
            timeout=5,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return False
