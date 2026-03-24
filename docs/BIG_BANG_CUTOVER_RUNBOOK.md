# EasyMail Big Bang Cutover Runbook

## 1. Preconditions

- Freeze all writes to old deployment branch.
- Backup MySQL and storage data directory.
- Export current `easymail.yaml` and postfix-related runtime config.
- Verify new binary starts with `go run ./cmd/easymailv2`.

## 2. Validation Gates

- `go test ./...` passes.
- Policy evaluator contract test passes.
- Replay check runs with zero diff:
  - `go run ./tools/replay -fixtures ./tools/replay/fixtures/policy_cases.json`
- Health checks are green on new services (`/check`).

## 3. Cutover Procedure

1. Stop old process.
2. Apply migration scripts in `scripts/migration`.
3. Start new process (`cmd/easymailv2`).
4. Validate:
   - admin login
   - webmail login
   - inbound mail path (`policy -> filter -> lmtp -> storage`)
   - outbound mail path (webmail send)
5. Monitor logs/metrics for 30 minutes.
6. Run smoke script:
   - `pwsh ./scripts/cutover/smoke.ps1`

## 4. Rollback Procedure

1. Stop new process.
2. Restore DB snapshot and storage snapshot.
3. Start old process (`cmd/easymail`).
4. Re-run smoke checks.
5. If rollback passes, keep legacy entrypoint `cmd/easymail/main.go` as active.

## 5. Smoke Checklist

- Admin routes work: `/login`, `/dashboard`, `/account/index`
- Webmail routes work: `/login`, `/mailbox/`, `/mailbox/write`
- Policy service returns expected action for valid and invalid sender/recipient.
- LMTP writes mail file and DB index.

