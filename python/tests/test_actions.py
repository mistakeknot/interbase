"""Tests for interbase action functions."""
import os
from unittest.mock import patch

from interbase import phase_set, emit_event, session_status


def test_phase_set_no_bd():
    with patch.dict(os.environ, {"PATH": ""}):
        # Should succeed silently (fail-open)
        phase_set("bead-123", "planned")


def test_emit_event_no_ic():
    with patch.dict(os.environ, {"PATH": ""}):
        emit_event("run-123", "test-event")


def test_session_status_format():
    status = session_status()
    assert status.startswith("[interverse]")
    assert "beads=" in status
    assert "ic=" in status
