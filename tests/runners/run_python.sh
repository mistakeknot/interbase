#!/usr/bin/env bash
# Conformance test runner for Python interbase SDK.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SDK_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$SDK_ROOT/python"
uv run pytest tests/test_conformance.py -v 2>&1 || exit 1
echo "Python conformance: PASS"
