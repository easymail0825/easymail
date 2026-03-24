Param(
  [string]$ConfigPath = "cmd/easymail/easymail.yaml"
)

Write-Host "[cutover] verify config path: $ConfigPath"
if (!(Test-Path $ConfigPath)) {
  Write-Error "config not found: $ConfigPath"
  exit 1
}

Write-Host "[cutover] running unit tests..."
go test ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "[cutover] running replay checks..."
go run ./tools/replay -fixtures ./tools/replay/fixtures/policy_cases.json
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "[cutover] all checks passed."

