# Interbase Go SDK

Go packages for Demarch plugin integration.

## Packages

- **`interbase`** (root) — Guards, actions, config/discovery. All fail-open.
- **`toolerror`** — Structured MCP error contract with 6 error types.
- **`mcputil`** — MCP handler middleware with timing, error counting, panic recovery.

## Usage

```go
import "github.com/mistakeknot/interbase"
import "github.com/mistakeknot/interbase/toolerror"
import "github.com/mistakeknot/interbase/mcputil"
```

In consumer `go.mod`:
```
require github.com/mistakeknot/interbase v0.0.0
replace github.com/mistakeknot/interbase => ../../sdk/interbase/go
```

See `sdk/interbase/docs/migration-guide.md` for adoption patterns.
