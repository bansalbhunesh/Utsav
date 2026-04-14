# UTSAV monorepo

Production-shaped scaffold: **Next.js** (`apps/web`) + **Go/Gin API** (`services/api`) + **PostgreSQL** (`db/migrations`, `infra/docker`).

Product source of truth (copy into repo for portability): `../UTSAV_PRODUCT_SPEC_v1.5.1.md` on this machine.

**Step-by-step local run (including when Docker fails on Windows):** see [docs/LOCAL_DEV.md](docs/LOCAL_DEV.md).

**Windows quick start** (Docker Desktop running): `pwsh -File scripts/start-local.ps1` from repo root — opens API + web in separate windows.

## Prereqs

- Docker (for Postgres)
- Go 1.22+
- Node 20+ (for Next.js)

After freeing disk space, `go build ./cmd/server` and `npm run build` should succeed locally (output binary: `services/api/bin/utsav-api.exe` if you use the same build command).

## Run Postgres

```bash
docker compose -f infra/docker/compose.yml up -d
```

## Run API

```bash
cd services/api
export MIGRATIONS_PATH=../../db/migrations
export DATABASE_URL=postgres://utsav:utsav@127.0.0.1:5432/utsav?sslmode=disable
export HTTP_PORT=8080
go mod tidy
go run ./cmd/server
```

Windows PowerShell:

```powershell
cd services/api
$env:MIGRATIONS_PATH="..\..\db\migrations"
$env:DATABASE_URL="postgres://utsav:utsav@127.0.0.1:5432/utsav?sslmode=disable"
$env:HTTP_PORT="8080"
go mod tidy
go run ./cmd/server
```

## Run Web

```bash
cd apps/web
npm install
npm run dev
```

`next.config.mjs` rewrites `/v1/*` to `NEXT_PUBLIC_API_URL` (default `http://127.0.0.1:8080`).

## Dev OTP

API reads `DEV_OTP_CODE` (default `123456`). OTP challenges use bcrypt hashes of that code in development.

## CI

See `.github/workflows/ci.yml` (runs on GitHub where disk is available).

[![CI](https://github.com/bhune/utsav/actions/workflows/ci.yml/badge.svg)](https://github.com/bhune/utsav/actions/workflows/ci.yml)

Key jobs in the workflow:
- `api` (unit/build)
- `api-integration` (testcontainers migration check)
- `web` (lint/build)
- `e2e-smoke` (Playwright login -> create event -> RSVP)

### Re-run CI checks locally

```bash
# API unit
cd services/api
go test ./...

# API integration (Docker required)
go test -tags=integration ./internal/httpserver -run TestMigrationsAgainstPostgresContainer -v

# Web lint/build
cd ../../apps/web
npm run lint
npm run build

# Playwright smoke (API + web running locally)
npx playwright install
npm run test:e2e
```
