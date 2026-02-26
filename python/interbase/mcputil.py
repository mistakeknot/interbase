"""MCP tool handler middleware — timing, error wrapping, metrics."""

from __future__ import annotations

import time
from collections import defaultdict
from dataclasses import dataclass, field
from typing import Any, Callable


@dataclass
class ToolStats:
    """Snapshot of metrics for a single tool."""

    calls: int = 0
    errors: int = 0
    total_duration_ns: int = 0

    @property
    def total_duration_s(self) -> float:
        return self.total_duration_ns / 1_000_000_000

    def __str__(self) -> str:
        return f"calls={self.calls} errors={self.errors} duration={self.total_duration_s:.3f}s"


@dataclass
class McpMetrics:
    """Collects per-tool call metrics for MCP servers."""

    _tools: dict[str, ToolStats] = field(default_factory=lambda: defaultdict(ToolStats))

    def tool_metrics(self) -> dict[str, ToolStats]:
        """Return a snapshot of metrics for all tools."""
        return dict(self._tools)

    def instrument(self, tool_name: str, handler: Callable) -> Callable:
        """Wrap a tool handler with timing, error counting, and error wrapping.

        Args:
            tool_name: The MCP tool name for metric grouping.
            handler: The original handler callable.

        Returns:
            Wrapped handler that collects metrics.
        """
        stats = self._tools[tool_name]

        def wrapper(*args: Any, **kwargs: Any) -> Any:
            stats.calls += 1
            start = time.monotonic_ns()
            try:
                result = handler(*args, **kwargs)
                return result
            except Exception:
                stats.errors += 1
                raise
            finally:
                stats.total_duration_ns += time.monotonic_ns() - start

        return wrapper
