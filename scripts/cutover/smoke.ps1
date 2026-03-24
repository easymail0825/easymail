Param(
  [string]$AdminBase = "http://127.0.0.1:10088",
  [string]$WebmailBase = "http://127.0.0.1:10089"
)

$ErrorActionPreference = "Stop"

function Check-Endpoint([string]$url) {
  $resp = Invoke-WebRequest -Uri $url -Method GET -TimeoutSec 8
  if ($resp.StatusCode -lt 200 -or $resp.StatusCode -ge 300) {
    throw "endpoint check failed: $url code=$($resp.StatusCode)"
  }
  Write-Host "[smoke] ok: $url"
}

Check-Endpoint "$AdminBase/check"
Check-Endpoint "$WebmailBase/check"
Check-Endpoint "$AdminBase/login"
Check-Endpoint "$WebmailBase/login"

Write-Host "[smoke] basic checks passed"

