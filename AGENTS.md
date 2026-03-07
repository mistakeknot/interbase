# interbase — Shared Integration SDK

## Canonical References
1. [`PHILOSOPHY.md`](../../PHILOSOPHY.md) — direction for ideation and planning decisions.
2. `CLAUDE.md` — implementation details, architecture, testing, and release workflow.

## Philosophy Alignment Protocol
Review [`PHILOSOPHY.md`](../../PHILOSOPHY.md) during:
- Intake/scoping
- Brainstorming
- Planning
- Execution kickoff
- Review/gates
- Handoff/retrospective

For brainstorming/planning outputs, add two short lines:
- **Alignment:** one sentence on how the proposal supports the module's purpose within Demarch's philosophy.
- **Conflict/Risk:** one sentence on any tension with philosophy (or 'none').

If a high-value change conflicts with philosophy, either:
- adjust the plan to align, or
- create follow-up work to update `PHILOSOPHY.md` explicitly.


Multi-language SDK enabling Interverse plugins to work in both standalone (Claude Code marketplace) and integrated (Clavain/Intercore ecosystem) modes. Bash SDK for hooks, Go SDK for MCP servers, Python SDK for hooks + scripts.

## File Structure

```
sdk/interbase/
  lib/interbase.sh      — core Bash SDK (installed to ~/.intermod/interbase/)
  templates/            — interbase-stub.sh (ships in plugins) + integration.json schema
  go/                   — Go SDK (interbase.go, toolerror/, mcputil/) — 36 tests
  python/               — Python SDK (guards, actions, config, nudge, toolerror, mcputil) — 6 test files
  tests/                — Bash tests (25 assertions) + conformance/ (YAML-driven, 3 runners)
  install.sh            — deploy to ~/.intermod/interbase/
```

## Topic Guides

| Topic | File | Covers |
|-------|------|--------|
| Bash SDK | [agents/bash-sdk.md](agents/bash-sdk.md) | Guards, actions, internal helpers (`ib_*` / `_ib_*`) |
| Go SDK | [agents/go-sdk.md](agents/go-sdk.md) | Guards, actions, config, toolerror, mcputil, test commands |
| Python SDK | [agents/python-sdk.md](agents/python-sdk.md) | Guards, actions, config, toolerror, mcputil, install |
| Conformance Tests | [agents/conformance-tests.md](agents/conformance-tests.md) | YAML-driven cross-language test suite, runners |
| Adoption | [agents/adoption.md](agents/adoption.md) | Install, plugin adoption pattern, dev testing, test commands, adopters |
| Design Patterns | [agents/design-patterns.md](agents/design-patterns.md) | Load guard, nudge protocol, relationship to interband |

## Quick Reference

```bash
# Install Bash SDK
bash sdk/interbase/install.sh

# Run all tests
bash tests/test-guards.sh && bash tests/test-nudge.sh && bash tests/test-config.sh
cd go && go test ./...
cd python && uv run pytest tests/ -v

# Conformance (all languages)
bash tests/runners/run_bash.sh
bash tests/runners/run_go.sh
bash tests/runners/run_python.sh

# Dev testing
INTERMOD_LIB=/path/to/dev/interbase.sh bash your-hook.sh
```
