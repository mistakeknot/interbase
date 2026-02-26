"""Structured MCP error contract — wire-format compatible with Go toolerror."""

from __future__ import annotations

import json
from typing import Any

# Error type constants — must match Go wire values exactly.
ERR_NOT_FOUND = "NOT_FOUND"
ERR_CONFLICT = "CONFLICT"
ERR_VALIDATION = "VALIDATION"
ERR_PERMISSION = "PERMISSION"
ERR_TRANSIENT = "TRANSIENT"
ERR_INTERNAL = "INTERNAL"

_DEFAULT_RECOVERABLE = {
    ERR_NOT_FOUND: False,
    ERR_CONFLICT: False,
    ERR_VALIDATION: False,
    ERR_PERMISSION: False,
    ERR_TRANSIENT: True,
    ERR_INTERNAL: False,
}


class ToolError(Exception):
    """Structured error for MCP tool handlers.

    Carries enough context for agents to make retry and fallback decisions.
    Wire format matches Go's toolerror.ToolError exactly.
    """

    def __init__(self, err_type: str, message: str, **data: Any) -> None:
        super().__init__(message)
        self.type: str = err_type
        self.message: str = message
        self.recoverable: bool = _DEFAULT_RECOVERABLE.get(err_type, False)
        self.data: dict[str, Any] = dict(data) if data else {}

    def with_recoverable(self, recoverable: bool) -> ToolError:
        """Override the recoverable flag. Returns self for chaining."""
        self.recoverable = recoverable
        return self

    def with_data(self, **kwargs: Any) -> ToolError:
        """Set data fields. Returns self for chaining."""
        self.data.update(kwargs)
        return self

    def json(self) -> str:
        """Serialize to JSON wire format matching Go's encoding/json output."""
        obj: dict[str, Any] = {
            "type": self.type,
            "message": self.message,
            "recoverable": self.recoverable,
        }
        if self.data:
            obj["data"] = self.data
        return json.dumps(obj, separators=(",", ":"))

    def __str__(self) -> str:
        return f"[{self.type}] {self.message}"

    @classmethod
    def from_error(cls, exc: BaseException) -> ToolError | None:
        """Extract a ToolError from an exception. Returns None if not one."""
        if isinstance(exc, ToolError):
            return exc
        return None

    @classmethod
    def wrap(cls, exc: BaseException) -> ToolError:
        """Convert any exception to ToolError. Passthrough if already one."""
        if isinstance(exc, ToolError):
            return exc
        return cls(ERR_INTERNAL, str(exc))
