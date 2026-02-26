"""Interbase — Shared integration SDK for Demarch Interverse plugins.

All guard functions are fail-open: they return False when their dependency
is missing. All action functions are silent no-ops when dependencies are
absent.
"""

from interbase.guards import (
    has_ic,
    has_bd,
    has_companion,
    in_ecosystem,
    get_bead,
    in_sprint,
)
from interbase.actions import phase_set, emit_event, session_status
from interbase.config import plugin_cache_path, ecosystem_root
from interbase.nudge import nudge_companion
from interbase.toolerror import ToolError, ERR_NOT_FOUND, ERR_CONFLICT, ERR_VALIDATION, ERR_PERMISSION, ERR_TRANSIENT, ERR_INTERNAL
from interbase.mcputil import McpMetrics, ToolStats

__version__ = "2.0.0"

__all__ = [
    "has_ic",
    "has_bd",
    "has_companion",
    "in_ecosystem",
    "get_bead",
    "in_sprint",
    "phase_set",
    "emit_event",
    "session_status",
    "plugin_cache_path",
    "ecosystem_root",
    "nudge_companion",
    "ToolError",
    "ERR_NOT_FOUND",
    "ERR_CONFLICT",
    "ERR_VALIDATION",
    "ERR_PERMISSION",
    "ERR_TRANSIENT",
    "ERR_INTERNAL",
    "McpMetrics",
    "ToolStats",
]
