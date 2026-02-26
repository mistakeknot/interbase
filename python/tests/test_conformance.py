"""Conformance test runner -- reads YAML test cases, runs against Python SDK."""

from __future__ import annotations

import json
import os
from pathlib import Path
from unittest.mock import patch

import pytest
import yaml

import interbase
from interbase.toolerror import ToolError

CONFORMANCE_DIR = Path(__file__).resolve().parent.parent.parent / "tests" / "conformance"


def load_suites():
    """Load all YAML conformance suites."""
    suites = []
    for f in sorted(CONFORMANCE_DIR.glob("*.yaml")):
        data = yaml.safe_load(f.read_text())
        languages = data.get("languages", [])
        if languages and "python" not in languages:
            continue
        for tc in data.get("tests", []):
            suites.append(pytest.param(tc, id=tc["name"]))
    return suites


@pytest.mark.parametrize("tc", load_suites())
def test_conformance(tc):
    setup = tc.get("setup", {})
    env_patch = {}
    for k, v in setup.items():
        env_patch[k] = str(v)

    with patch.dict(os.environ, env_patch, clear=False):
        call = tc["call"]
        args = tc.get("args", [])

        if call in ("has_ic", "has_bd", "has_companion", "in_ecosystem",
                     "get_bead", "in_sprint", "phase_set", "emit_event",
                     "session_status", "plugin_cache_path", "ecosystem_root"):
            func = getattr(interbase, call)
            result = func(*args)
        elif call == "toolerror_new":
            te = ToolError(args[0], args[1])
            result = json.loads(te.json())
        elif call == "toolerror_new_with_data":
            te = ToolError(args[0], args[1]).with_data(**tc.get("data", {}))
            result = json.loads(te.json())
        elif call == "toolerror_str":
            te = ToolError(args[0], args[1])
            result = str(te)
        elif call == "toolerror_wrap_tool_error":
            te = ToolError(args[0], args[1])
            result = json.loads(ToolError.wrap(te).json())
        elif call == "toolerror_wrap_generic":
            result = json.loads(ToolError.wrap(ValueError(args[0])).json())
        else:
            pytest.skip(f"unknown function: {call}")
            return

        # Assert
        if "expect" in tc:
            expected = tc["expect"]
            if isinstance(expected, bool):
                assert result is expected, f"got {result!r}, want {expected!r}"
            else:
                assert result == expected, f"got {result!r}, want {expected!r}"
        if "expect_error" in tc and not tc["expect_error"]:
            pass  # no error raised = pass
        if "expect_contains" in tc:
            assert tc["expect_contains"] in str(result)
        if "expect_json" in tc:
            expected = tc["expect_json"]
            for k, v in expected.items():
                if k == "expect_message_contains":
                    assert v in result.get("message", "")
                else:
                    assert result.get(k) == v, f"key {k}: got {result.get(k)!r}, want {v!r}"
