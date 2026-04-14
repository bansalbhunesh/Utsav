# Run UTSAV locally

## 1. Start Docker Desktop (Windows)

Postgres runs in Compose. If you see `docker_engine: The system cannot find the file specified`:

1. Install **Docker Desktop** and start it.
2. Ensure **WSL2** backend is enabled if prompted.
3. Retry:

```bash
cd utsav
docker compose -f infra/docker/compose.yml up -d
```

Wait until `pg_isready` passes (Compose healthcheck).

## 2. API (Go)

From repo root, paths are relative to `services/api`:

**Git Bash / macOS / Linux:**

```bash
cd services/api
export MIGRATIONS_PATH=../../db/migrations
export DATABASE_URL=postgres://utsav:utsav@127.0.0.1:5432/utsav?sslmode=disable
export HTTP_PORT=8080
export JWT_SECRET=dev-change-me
export CORS_ORIGIN=http://localhost:3000
go run ./cmd/server
```

**PowerShell:**

```powershell
cd services/api
$env:MIGRATIONS_PATH="..\..\db\migrations"
$env:DATABASE_URL="postgres://utsav:utsav@127.0.0.1:5432/utsav?sslmode=disable"
$env:HTTP_PORT="8080"
go run ./cmd/server
```

You should see: `utsav api listening on :8080`

Smoke test: `curl http://127.0.0.1:8080/v1/healthz`

## 3. Web (Next.js)

```bash
cd apps/web
npm run dev
```

Open **http://localhost:3000** — `/v1/*` is rewritten to the API (default `http://127.0.0.1:8080`).

## 4. Happy path

1. **Login** → OTP request → verify with `123456`.
2. **New event** → note slug (e.g. `demo-wedding`).
3. Open **Guest link** `/e/<slug>` → RSVP → Shagun (after RSVP guest token is set).
4. **Guests** → paste CSV (`name,phone` header) or **Import CSV** — duplicates by phone upsert per event.

If OTP endpoints return **429**, in-memory rate limits tripped (15 minutes): host login **5/IP**, RSVP OTP request **10/(IP+event+phone)**.

## 5. Tests

### API unit tests

```bash
cd services/api
go test ./...
```

### API integration (testcontainers + Postgres)

Requires Docker Desktop running. This test is behind a build tag so normal `go test` stays fast/light.

```bash
cd services/api
go test -tags=integration ./internal/httpserver -run TestMigrationsAgainstPostgresContainer -v
```

### Web e2e smoke (Playwright)

Start API + web locally first, then:

```bash
cd apps/web
npx playwright install
npm run test:e2e
```

Optional env vars:

- `E2E_BASE_URL` (default `http://127.0.0.1:3000`)
- `E2E_PHONE` (default `+919876543210`)
- `E2E_OTP` (default `123456`)
- `E2E_EVENT_SLUG` / `E2E_EVENT_TITLE`
