# Conformance Tests

YAML-defined test cases in `tests/conformance/` ensure all three SDKs (Bash, Go, Python) implement the same behavior. Each YAML file defines a `domain` (guards, actions, config, mcp) and a list of test cases with setup, call, args, and expected results.

## Domains

| File | Tests | Languages |
|------|-------|-----------|
| `guards.yaml` | 9 cases | Bash, Go, Python |
| `actions.yaml` | 3 cases | Bash, Go, Python |
| `config.yaml` | 3 cases | Bash, Go, Python |
| `mcp.yaml` | 6 cases | Go, Python (Bash excluded — hooks don't run MCP servers) |

## Running Conformance Tests

```bash
bash tests/runners/run_bash.sh     # parses YAML, calls ib_* functions
bash tests/runners/run_go.sh       # runs go test with conformance_test.go
bash tests/runners/run_python.sh   # runs pytest test_conformance.py
```

## Adding New Conformance Tests

1. Add a test case to the appropriate YAML file in `tests/conformance/`
2. If the test calls a new function, add the function dispatch to all three runners
3. Run all three runners to verify cross-language consistency
