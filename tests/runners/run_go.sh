#!/usr/bin/env bash
# Conformance test runner for Go interbase SDK.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SDK_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$SDK_ROOT/go"
go test -v -run TestConformance ./... -tags conformance 2>&1 || exit 1
echo "Go conformance: PASS"
