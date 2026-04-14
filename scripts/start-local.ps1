# Run from repo root after Docker Desktop is running:
#   pwsh -File scripts/start-local.ps1

$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)

Write-Host "Starting Postgres..."
docker compose -f infra/docker/compose.yml up -d

Write-Host "Starting API in new window..."
$apiDir = Join-Path (Get-Location) "services\api"
Start-Process pwsh -ArgumentList @(
  "-NoExit", "-Command",
  "cd '$apiDir'; `$env:MIGRATIONS_PATH='..\..\db\migrations'; `$env:DATABASE_URL='postgres://utsav:utsav@127.0.0.1:5432/utsav?sslmode=disable'; `$env:HTTP_PORT='8080'; `$env:CORS_ORIGIN='http://localhost:3000'; go run ./cmd/server"
)

Start-Sleep -Seconds 3

Write-Host "Starting web in new window..."
$webDir = Join-Path (Get-Location) "apps\web"
Start-Process pwsh -ArgumentList @("-NoExit", "-Command", "cd '$webDir'; npm run dev")

Write-Host "Open http://localhost:3000 when Next shows ready."
