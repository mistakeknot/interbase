"""Config + discovery functions."""

from __future__ import annotations

import glob
import os


def plugin_cache_path(plugin: str) -> str:
    """Return the cache path for a named plugin. Empty if not found."""
    if not plugin:
        return ""
    home = os.path.expanduser("~")
    pattern = os.path.join(home, ".claude", "plugins", "cache", "*", plugin, "*")
    matches = sorted(glob.glob(pattern))
    return matches[-1] if matches else ""


def ecosystem_root() -> str:
    """Return the Demarch monorepo root. Checks $DEMARCH_ROOT then walks up."""
    root = os.environ.get("DEMARCH_ROOT", "")
    if root:
        return root
    try:
        d = os.getcwd()
    except OSError:
        return ""
    while True:
        if os.path.isdir(os.path.join(d, "sdk", "interbase")):
            return d
        parent = os.path.dirname(d)
        if parent == d:
            break
        d = parent
    return ""
