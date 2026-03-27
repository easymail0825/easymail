$ErrorActionPreference = 'Stop'

function Ensure-Dir([string]$Path) {
  if (-not (Test-Path -LiteralPath $Path)) {
    New-Item -ItemType Directory -Path $Path | Out-Null
  }
}

$dirs = @(
  '.github/workflows', '.github/ISSUE_TEMPLATE', '.github/PULL_REQUEST_TEMPLATE',
  'cmd/api', 'cmd/worker', 'cmd/scheduler', 'cmd/migrate',
  'internal/app/api/handler', 'internal/app/api/handler/middleware', 'internal/app/api/router',
  'internal/app/service', 'internal/app/service/impl',
  'internal/app/domain/model', 'internal/app/domain/repository', 'internal/app/domain/service', 'internal/app/domain/event',
  'internal/app/cmd/cli',
  'internal/pkg/config',
  'internal/pkg/database/mysql', 'internal/pkg/database/redis', 'internal/pkg/database/mongodb',
  'internal/pkg/cache',
  'internal/pkg/queue/kafka', 'internal/pkg/queue/rabbitmq', 'internal/pkg/queue/nsq',
  'internal/pkg/logger',
  'internal/pkg/auth/jwt', 'internal/pkg/auth/oauth2', 'internal/pkg/auth/casbin',
  'internal/pkg/validator', 'internal/pkg/tracer', 'internal/pkg/metrics', 'internal/pkg/health',
  'internal/pkg/errors', 'internal/pkg/constants', 'internal/pkg/types', 'internal/pkg/utils',
  'internal/infrastructure/persistence/mysql', 'internal/infrastructure/persistence/redis', 'internal/infrastructure/persistence/elasticsearch',
  'internal/infrastructure/messaging', 'internal/infrastructure/http',
  'pkg/api', 'pkg/errors', 'pkg/utils/common',
  'api/proto/v1', 'api/openapi', 'api/asyncapi', 'api/jsonschema',
  'deployments/docker',
  'deployments/k8s/base', 'deployments/k8s/overlays/dev', 'deployments/k8s/overlays/staging', 'deployments/k8s/overlays/prod', 'deployments/k8s/helm/myapp',
  'deployments/terraform', 'deployments/ansible/playbooks',
  'scripts/build', 'scripts/dev', 'scripts/db/migrations', 'scripts/db/seeds', 'scripts/db/schema', 'scripts/release', 'scripts/tools',
  'test/integration', 'test/mocks', 'test/fixtures', 'test/testdata', 'test/testutil',
  'web/src', 'web/public',
  'docs/api', 'docs/architecture', 'docs/deployment', 'docs/development'
)

foreach ($d in $dirs) { Ensure-Dir $d }

# cmd: easymail -> api
if (Test-Path -LiteralPath 'cmd/easymail/main.go') {
  Move-Item -LiteralPath 'cmd/easymail/main.go' -Destination 'cmd/api/main.go' -Force
}
if (Test-Path -LiteralPath 'cmd/easymail/easymail.yaml') {
  Copy-Item -LiteralPath 'cmd/easymail/easymail.yaml' -Destination 'configs/config.example.yaml' -Force
  Move-Item -LiteralPath 'cmd/easymail/easymail.yaml' -Destination 'configs/config.yaml' -Force
}
if ((Test-Path -LiteralPath 'configs/easymail.yaml') -and (-not (Test-Path -LiteralPath 'configs/config.dev.yaml'))) {
  Copy-Item -LiteralPath 'configs/easymail.yaml' -Destination 'configs/config.dev.yaml' -Force
}

# internal: models -> domain/model
if (Test-Path -LiteralPath 'internal/model') {
  Ensure-Dir 'internal/app/domain/model'
  Get-ChildItem -LiteralPath 'internal/model' -Force | ForEach-Object {
    Move-Item -LiteralPath $_.FullName -Destination 'internal/app/domain/model' -Force
  }
  Remove-Item -LiteralPath 'internal/model' -Force -Recurse
}

# internal: application services
if (Test-Path -LiteralPath 'internal/application/auth') { Move-Item -LiteralPath 'internal/application/auth' -Destination 'internal/app/service/auth' -Force }
if (Test-Path -LiteralPath 'internal/application/session') { Move-Item -LiteralPath 'internal/application/session' -Destination 'internal/app/service/session' -Force }
if (Test-Path -LiteralPath 'internal/identity') { Move-Item -LiteralPath 'internal/identity' -Destination 'internal/app/service/identity' -Force }

# internal: service layer
if (Test-Path -LiteralPath 'internal/service/interface.go') { Move-Item -LiteralPath 'internal/service/interface.go' -Destination 'internal/app/service/interface.go' -Force }
if (Test-Path -LiteralPath 'internal/service/admin') { Move-Item -LiteralPath 'internal/service/admin' -Destination 'internal/app/api/handler/admin' -Force }
if (Test-Path -LiteralPath 'internal/service/webmail') { Move-Item -LiteralPath 'internal/service/webmail' -Destination 'internal/app/api/handler/webmail' -Force }
if (Test-Path -LiteralPath 'internal/service/dovecot') { Move-Item -LiteralPath 'internal/service/dovecot' -Destination 'internal/app/service/dovecot' -Force }
if (Test-Path -LiteralPath 'internal/service/milter') { Move-Item -LiteralPath 'internal/service/milter' -Destination 'internal/app/service/milter' -Force }
if (Test-Path -LiteralPath 'internal/service/storage') { Move-Item -LiteralPath 'internal/service/storage' -Destination 'internal/app/service/storage' -Force }

# internal: storage wrapper
if (Test-Path -LiteralPath 'internal/storage') { Move-Item -LiteralPath 'internal/storage' -Destination 'internal/app/service/storage_legacy' -Force }

# internal: easylog -> logger
if (Test-Path -LiteralPath 'internal/easylog') { Move-Item -LiteralPath 'internal/easylog' -Destination 'internal/pkg/logger/easylog' -Force }

# internal: database split
if (Test-Path -LiteralPath 'internal/database/mysql.go') { Move-Item -LiteralPath 'internal/database/mysql.go' -Destination 'internal/pkg/database/mysql/mysql.go' -Force }
if (Test-Path -LiteralPath 'internal/database/redis.go') { Move-Item -LiteralPath 'internal/database/redis.go' -Destination 'internal/pkg/database/redis/redis.go' -Force }
if (Test-Path -LiteralPath 'internal/database/init.go') { Move-Item -LiteralPath 'internal/database/init.go' -Destination 'internal/pkg/database/init.go' -Force }
if (Test-Path -LiteralPath 'internal/database/application.go') { Move-Item -LiteralPath 'internal/database/application.go' -Destination 'internal/pkg/database/application.go' -Force }
if (Test-Path -LiteralPath 'internal/database') {
  $count = (Get-ChildItem -LiteralPath 'internal/database' -Recurse -Force | Measure-Object).Count
  if ($count -eq 0) { Remove-Item -LiteralPath 'internal/database' -Force -Recurse }
}

# internal: preprocessing -> utils
if (Test-Path -LiteralPath 'internal/preprocessing') { Move-Item -LiteralPath 'internal/preprocessing' -Destination 'internal/pkg/utils/preprocessing' -Force }

# internal: observability -> tracer/health
if (Test-Path -LiteralPath 'internal/observability/health') { Move-Item -LiteralPath 'internal/observability/health' -Destination 'internal/pkg/health/health_legacy' -Force }
if (Test-Path -LiteralPath 'internal/observability/sessiontrace') { Move-Item -LiteralPath 'internal/observability/sessiontrace' -Destination 'internal/pkg/tracer/sessiontrace' -Force }
if (Test-Path -LiteralPath 'internal/observability') {
  $count = (Get-ChildItem -LiteralPath 'internal/observability' -Recurse -Force | Measure-Object).Count
  if ($count -eq 0) { Remove-Item -LiteralPath 'internal/observability' -Force -Recurse }
}

# internal: infrastructure repository -> persistence
if (Test-Path -LiteralPath 'internal/infrastructure/repository/account_auth.go') {
  Move-Item -LiteralPath 'internal/infrastructure/repository/account_auth.go' -Destination 'internal/infrastructure/persistence/mysql/account_auth.go' -Force
}

Write-Host 'Reorg script finished.'

