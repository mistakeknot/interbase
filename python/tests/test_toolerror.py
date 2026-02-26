"""Tests for interbase.toolerror — wire format parity with Go."""
import json

from interbase.toolerror import ToolError, ERR_NOT_FOUND, ERR_TRANSIENT, ERR_INTERNAL


def test_new_error_has_correct_type():
    te = ToolError(ERR_NOT_FOUND, "agent 'x' not found")
    assert te.type == "NOT_FOUND"
    assert te.message == "agent 'x' not found"
    assert te.recoverable is False


def test_transient_default_recoverable():
    te = ToolError(ERR_TRANSIENT, "db busy")
    assert te.recoverable is True


def test_with_recoverable_override():
    te = ToolError(ERR_NOT_FOUND, "msg").with_recoverable(True)
    assert te.recoverable is True


def test_with_data():
    te = ToolError(ERR_NOT_FOUND, "msg").with_data(file="main.go")
    assert te.data == {"file": "main.go"}


def test_json_wire_format():
    te = ToolError(ERR_NOT_FOUND, "agent 'fd-safety' not registered")
    parsed = json.loads(te.json())
    assert parsed["type"] == "NOT_FOUND"
    assert parsed["message"] == "agent 'fd-safety' not registered"
    assert parsed["recoverable"] is False
    # data should be omitted when empty (match Go omitempty)
    assert "data" not in parsed or parsed["data"] == {}


def test_json_with_data():
    te = ToolError(ERR_NOT_FOUND, "msg").with_data(file="main.go")
    parsed = json.loads(te.json())
    assert parsed["data"] == {"file": "main.go"}


def test_str_format():
    te = ToolError(ERR_NOT_FOUND, "agent gone")
    assert str(te) == "[NOT_FOUND] agent gone"


def test_wrap_regular_exception():
    exc = ValueError("bad value")
    te = ToolError.wrap(exc)
    assert te.type == "INTERNAL"
    assert "bad value" in te.message


def test_wrap_tool_error():
    original = ToolError(ERR_NOT_FOUND, "gone")
    wrapped = ToolError.wrap(original)
    assert wrapped is original


def test_from_error_found():
    te = ToolError(ERR_NOT_FOUND, "gone")
    found = ToolError.from_error(te)
    assert found is te


def test_from_error_not_found():
    exc = ValueError("nope")
    assert ToolError.from_error(exc) is None
