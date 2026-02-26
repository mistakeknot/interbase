#!/usr/bin/env bash
# Conformance test runner for Bash interbase SDK.
# Reads YAML test cases and executes against lib/interbase.sh.
#
# SECURITY: Uses allowlisted env vars instead of eval. YAML values are
# never executed as shell code.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SDK_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
CONFORMANCE_DIR="$SCRIPT_DIR/../conformance"

# Save original PATH for python3 invocations
ORIG_PATH="$PATH"

source "$SDK_ROOT/lib/interbase.sh"

pass=0
fail=0
skip=0

# Allowlisted env vars that setup blocks can set
declare -A ALLOWED_SETUP_VARS=([PATH]=1 [CLAVAIN_BEAD_ID]=1 [INTERMOD_LIB]=1 [DEMARCH_ROOT]=1)

apply_setup() {
    # Parse setup JSON using python3 on the original PATH
    local setup_json="$1"
    if [[ -z "$setup_json" || "$setup_json" == "{}" ]]; then
        return
    fi
    local pairs
    pairs=$(PATH="$ORIG_PATH" python3 -c "
import json, sys
d = json.loads(sys.argv[1])
for k, v in d.items():
    print(f'{k}\t{v}')
" "$setup_json")
    while IFS=$'\t' read -r skey sval; do
        if [[ -n "${ALLOWED_SETUP_VARS[$skey]+x}" ]]; then
            export "$skey=$sval"
        fi
    done <<< "$pairs"
}

run_test() {
    local yaml_file="$1"

    # Skip MCP tests for Bash
    local languages
    languages=$(grep '^languages:' "$yaml_file" 2>/dev/null || echo "")
    if [[ -n "$languages" ]] && [[ "$languages" != *"bash"* ]]; then
        echo "SKIP  $yaml_file (not for bash)"
        return
    fi

    # Parse ALL tests from YAML (single python3 invocation)
    # Uses \x1f (Unit Separator) as delimiter to avoid bash read collapsing empty fields
    # Format: name<US>call<US>arg0<US>setup_json<US>expect_type<US>expect_value
    local parsed
    parsed=$(PATH="$ORIG_PATH" python3 -c "
import yaml, json, sys

SEP = '\x1f'

with open(sys.argv[1]) as f:
    data = yaml.safe_load(f)

for t in data.get('tests', []):
    name = t.get('name', '')
    call = t.get('call', '')
    args = t.get('args', [])
    arg0 = str(args[0]) if args else ''
    setup = json.dumps(t.get('setup', {}))

    if 'expect' in t:
        val = t['expect']
        # Normalize booleans to lowercase for bash comparison
        if isinstance(val, bool):
            etype, evalue = 'exact', str(val).lower()
        else:
            etype, evalue = 'exact', str(val)
    elif 'expect_error' in t:
        etype, evalue = ('no_error', '') if not t['expect_error'] else ('error', '')
    elif 'expect_contains' in t:
        etype, evalue = 'contains', t['expect_contains']
    elif 'expect_no_error' in t:
        etype, evalue = 'no_error', ''
    else:
        etype, evalue = 'skip', ''
    fields = [name, call, arg0, setup, etype, evalue]
    print(SEP.join(fields))
" "$yaml_file")

    while IFS=$'\x1f' read -r name call arg0 setup_json expect_type expect_value; do
        [[ -n "$name" ]] || continue

        # Save env
        local old_path="${PATH}" old_bead="${CLAVAIN_BEAD_ID:-}" old_intermod="${INTERMOD_LIB:-}" old_demarch="${DEMARCH_ROOT:-}"

        # Apply setup
        apply_setup "$setup_json"

        # Execute — use if/else to capture exit codes under set -e
        local result=""
        case "$call" in
            has_ic)        if ib_has_ic; then result="true"; else result="false"; fi ;;
            has_bd)        if ib_has_bd; then result="true"; else result="false"; fi ;;
            has_companion) if ib_has_companion "$arg0"; then result="true"; else result="false"; fi ;;
            get_bead)      result=$(ib_get_bead) ;;
            in_ecosystem)  if ib_in_ecosystem; then result="true"; else result="false"; fi ;;
            in_sprint)     if ib_in_sprint; then result="true"; else result="false"; fi ;;
            phase_set)     ib_phase_set "bead-123" "planned"; result="no_error" ;;
            emit_event)    ib_emit_event "run-123" "test-event"; result="no_error" ;;
            session_status) result=$(ib_session_status 2>&1) ;;
            plugin_cache_path) result=$(ib_plugin_cache_path "$arg0" 2>/dev/null) || true ;;
            ecosystem_root) result=$(ib_ecosystem_root 2>/dev/null) || true ;;
            *) echo "SKIP  $name (unknown: $call)"; skip=$((skip + 1))
               export PATH="$old_path" CLAVAIN_BEAD_ID="$old_bead" INTERMOD_LIB="$old_intermod" DEMARCH_ROOT="$old_demarch"
               continue ;;
        esac

        # Restore env
        export PATH="$old_path" CLAVAIN_BEAD_ID="$old_bead" INTERMOD_LIB="$old_intermod" DEMARCH_ROOT="$old_demarch"

        # Assert
        case "$expect_type" in
            exact)
                if [[ "$result" == "$expect_value" ]]; then
                    echo "PASS  $name"; pass=$((pass + 1))
                else
                    echo "FAIL  $name: got '$result', expected '$expect_value'"; fail=$((fail + 1))
                fi ;;
            contains)
                if [[ "$result" == *"$expect_value"* ]]; then
                    echo "PASS  $name"; pass=$((pass + 1))
                else
                    echo "FAIL  $name: '$result' missing '$expect_value'"; fail=$((fail + 1))
                fi ;;
            no_error) echo "PASS  $name"; pass=$((pass + 1)) ;;
            skip) skip=$((skip + 1)) ;;
        esac
    done <<< "$parsed"
}

for yaml_file in "$CONFORMANCE_DIR"/*.yaml; do
    run_test "$yaml_file"
done

echo ""
echo "Results: $pass passed, $fail failed, $skip skipped"
[[ "$fail" -eq 0 ]]
