"""Tests for interbase.mcputil — metrics middleware."""
from interbase.mcputil import McpMetrics


def test_instrument_counts_calls():
    metrics = McpMetrics()
    handler = lambda x: x * 2
    wrapped = metrics.instrument("double", handler)

    assert wrapped(5) == 10
    assert wrapped(3) == 6

    stats = metrics.tool_metrics()["double"]
    assert stats.calls == 2
    assert stats.errors == 0


def test_instrument_counts_errors():
    metrics = McpMetrics()

    def bad_handler():
        raise ValueError("boom")

    wrapped = metrics.instrument("bad", bad_handler)

    try:
        wrapped()
    except ValueError:
        pass

    stats = metrics.tool_metrics()["bad"]
    assert stats.calls == 1
    assert stats.errors == 1


def test_instrument_tracks_duration():
    metrics = McpMetrics()
    import time

    def slow_handler():
        time.sleep(0.01)  # 10ms

    wrapped = metrics.instrument("slow", slow_handler)
    wrapped()

    stats = metrics.tool_metrics()["slow"]
    assert stats.total_duration_ns > 0
    assert stats.total_duration_s > 0.005  # at least 5ms
