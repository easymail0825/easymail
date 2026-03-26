# EasyMail Architecture V2

## Modular Monolith Layout

- `internal/platform`: bootstrap, config, runtime orchestration
- `internal/identity`: auth/account/domain/session semantics
- `internal/policy`: policy evaluator and policy server adapter
- `internal/filtering`: milter filter server adapter
- `internal/delivery`: lmtp delivery adapter
- `internal/storage`: storage abstraction adapter
- `internal/adminapi`: admin transport adapter
- `internal/webmailapi`: webmail transport adapter

## Boot Entry

- Default entry: `cmd/easymail/main.go` (bootstrap/runtime)

## Migration Rule

- Existing behavior is preserved by adapters first.
- New business logic must be added in module packages, not old controllers/models directly.

