Write-Host "[cutover] stopping old service (manual/system specific)"
Write-Host "[cutover] run migration scripts (manual)"
Write-Host "[cutover] starting new service"
go run ./cmd/easymailv2

