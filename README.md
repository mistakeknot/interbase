# interbase

Shared multi-language SDK for the Demarch ecosystem. Two layers:

- **Bash SDK** — fail-open guards and no-op actions for plugin hooks. If ecosystem tools (Beads, Intercore) are present, everything lights up; if absent, everything still works.
- **Go SDK** — shared packages for MCP servers. Currently provides structured error contracts so agents can distinguish retriable from permanent failures.

## Who this is for

- **Plugin authors** building Interverse-compatible Claude Code plugins (Bash SDK)
- **MCP server developers** building Go services in the Demarch ecosystem (Go SDK)

End users don't interact with interbase directly.

## Bash SDK

Each plugin ships a thin `interbase-stub.sh` that checks for the centralized SDK at `~/.intermod/interbase/interbase.sh`. If found, the full SDK loads. If not, inline no-op stubs activate and every `ib_*` function returns a safe default.

```bash
# Install
bash install.sh

# Run tests
bash tests/test-guards.sh    # 16 tests
bash tests/test-nudge.sh     # 4 tests
```

See `AGENTS.md` for the full function reference, stub pattern, and adoption guide.

## Go SDK

Shared Go packages for Demarch MCP servers. See [`go/README.md`](go/README.md) for the full reference.

```go
import "github.com/mistakeknot/interbase/toolerror"

// Return structured errors from MCP tool handlers
return mcp.NewToolResultError(
    toolerror.New(toolerror.ErrNotFound, "agent %q not found", name).JSON(),
), nil
```

```bash
# Run tests
cd go && go test ./...
```

See [`docs/sdk-toolerror.md`](../../docs/sdk-toolerror.md) for the design rationale and wire format specification.
