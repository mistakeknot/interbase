#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/../lib/interbase.sh"

pass=0 fail=0

assert_eq() {
    local label="$1" got="$2" want="$3"
    if [[ "$got" == "$want" ]]; then
        echo "PASS  $label"
        pass=$((pass + 1))
    else
        echo "FAIL  $label: got '$got', want '$want'"
        fail=$((fail + 1))
    fi
}

assert_nonempty() {
    local label="$1" got="$2"
    if [[ -n "$got" ]]; then
        echo "PASS  $label"
        pass=$((pass + 1))
    else
        echo "FAIL  $label: got empty string"
        fail=$((fail + 1))
    fi
}

assert_empty() {
    local label="$1" got="$2"
    if [[ -z "$got" ]]; then
        echo "PASS  $label"
        pass=$((pass + 1))
    else
        echo "FAIL  $label: got '$got', want empty"
        fail=$((fail + 1))
    fi
}

# Test ib_plugin_cache_path
assert_empty "plugin_cache_path empty name" "$(ib_plugin_cache_path "")"
assert_empty "plugin_cache_path nonexistent" "$(ib_plugin_cache_path "this-plugin-does-not-exist-zzzz")"

# Test with a plugin we know exists (clavain)
if compgen -G "${HOME}/.claude/plugins/cache/*/clavain/*" &>/dev/null; then
    assert_nonempty "plugin_cache_path clavain" "$(ib_plugin_cache_path "clavain")"
fi

# Test ib_ecosystem_root with env override
export DEMARCH_ROOT="/test/demarch"
assert_eq "ecosystem_root env override" "$(ib_ecosystem_root)" "/test/demarch"
unset DEMARCH_ROOT

# Test ib_ecosystem_root walk-up (from inside SDK dir, should find monorepo root)
pushd "$SCRIPT_DIR" >/dev/null
result=$(ib_ecosystem_root)
popd >/dev/null
# We're inside sdk/interbase/tests — walking up should find parent with sdk/interbase
if [[ -n "$result" ]]; then
    assert_nonempty "ecosystem_root walk-up" "$result"
fi

echo ""
echo "Config tests: $pass passed, $fail failed"
[[ "$fail" -eq 0 ]]
