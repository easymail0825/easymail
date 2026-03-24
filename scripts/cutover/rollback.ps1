Param(
  [string]$LegacyEntry = "./cmd/easymail/main.go"
)

Write-Host "[rollback] stopping v2 service (manual/system specific)"
Write-Host "[rollback] restoring DB and storage snapshots (manual/system specific)"
Write-Host "[rollback] starting legacy service"
go run $LegacyEntry

