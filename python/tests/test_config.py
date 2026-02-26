"""Tests for interbase config functions."""
import os
from unittest.mock import patch

from interbase import plugin_cache_path, ecosystem_root


def test_plugin_cache_path_empty():
    assert plugin_cache_path("") == ""


def test_ecosystem_root_env_override():
    with patch.dict(os.environ, {"DEMARCH_ROOT": "/test/demarch"}):
        assert ecosystem_root() == "/test/demarch"


def test_ecosystem_root_unset():
    with patch.dict(os.environ, {"DEMARCH_ROOT": ""}, clear=False):
        # Should return something or empty — just shouldn't raise
        ecosystem_root()
