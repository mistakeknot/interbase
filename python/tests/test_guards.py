"""Tests for interbase guard functions."""
import os
import tempfile
from unittest.mock import patch

from interbase import has_ic, has_bd, has_companion, in_ecosystem, get_bead, in_sprint


def test_has_ic_when_missing():
    with patch.dict(os.environ, {"PATH": ""}):
        assert has_ic() is False


def test_has_bd_when_missing():
    with patch.dict(os.environ, {"PATH": ""}):
        assert has_bd() is False


def test_has_companion_empty_name():
    assert has_companion("") is False


def test_has_companion_nonexistent():
    assert has_companion("this-plugin-does-not-exist-zzzz") is False


def test_get_bead_set():
    with patch.dict(os.environ, {"CLAVAIN_BEAD_ID": "iv-test123"}):
        assert get_bead() == "iv-test123"


def test_get_bead_unset():
    with patch.dict(os.environ, {"CLAVAIN_BEAD_ID": ""}, clear=False):
        assert get_bead() == ""


def test_in_ecosystem_file_exists():
    with tempfile.NamedTemporaryFile(suffix=".sh") as f:
        with patch.dict(os.environ, {"INTERMOD_LIB": f.name}):
            assert in_ecosystem() is True


def test_in_ecosystem_file_missing():
    with patch.dict(os.environ, {"INTERMOD_LIB": "/nonexistent/path/interbase.sh"}):
        assert in_ecosystem() is False


def test_in_sprint_no_bead():
    with patch.dict(os.environ, {"CLAVAIN_BEAD_ID": ""}, clear=False):
        assert in_sprint() is False


def test_in_sprint_no_ic():
    with patch.dict(os.environ, {"CLAVAIN_BEAD_ID": "iv-test", "PATH": ""}):
        assert in_sprint() is False
