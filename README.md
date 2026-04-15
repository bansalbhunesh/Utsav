# UTSAV

UTSAV is an event operations platform for weddings and social celebrations.  
It unifies host, guest, and organiser workflows into one system: OTP auth, RSVP, guest operations, shagun tracking, gallery, broadcasts, and memory-book generation.

[![CI](https://github.com/bansalbhunesh/Utsav/actions/workflows/ci.yml/badge.svg)](https://github.com/bansalbhunesh/Utsav/actions/workflows/ci.yml)
[![Next.js](https://img.shields.io/badge/Next.js-16-black)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Proprietary-lightgrey)](#license)

## Table of Contents

- [Quick Start](#quick-start)
- [What’s in this Repository](#whats-in-this-repository)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Production Hardening Included](#production-hardening-included)
- [Local Validation Commands](#local-validation-commands)
- [API Examples](#api-examples)
- [Environment Variables](#environment-variables)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)
- [License](#license)

## Quick Start

### Prerequisites

- Node.js 20+
- npm
- Go 1.25+
- Docker Desktop

### 1) Start local Postgres

```bash
make docker-up
```

### 2) Start backend API

```bash
cd services/api
cp .env.example .env
go run ./cmd/server
```

### 3) Start frontend

```bash
cd ../..
cp .env.example .env.local
npm install
npm run dev
```

Open `http://localhost:3000`.

## What’s in this Repository

### Core product modules

- Host auth (OTP + refresh)
- Guest RSVP auth (OTP)
- Event CRUD + sub-events + members
- Guest management + CSV import
- RSVP (public submit + host visibility)
- Shagun (host logging + public report)
- Gallery (presign/register/moderation/public)
- Broadcasts (host/public)
- Memory book (generate/public/export-gated)
- Organiser profile + clients + event linking
- Billing checkout + webhook pipeline

### Monorepo structure

- `app/` - Next.js routes (host, organiser, public guest)
- `components/` - UI/domain components
- `lib/` - API client, contracts, shared utils
- `providers/` - app-level providers
- `store/` - client state stores
- `services/api/` - Go API service
- `db/migrations/` - SQL migrations
- `.github/workflows/ci.yml` - CI pipeline

## Architecture

### Backend layering

- Handler: `services/api/internal/httpserver`
- Service: `services/api/internal/service`
- Repository: `services/api/internal/repository`

### API shape

- Versioned routes under `/v1`
- Health endpoints: `/health`, `/v1/healthz`, `/v1/readyz`
- Public routes: `/v1/public/...`
- Host/organiser routes protected by auth middleware

### Auth model

- Host session: OTP -> httpOnly cookies (`utsav_access_token`, `utsav_refresh_token`)
- Guest session: event-scoped guest token for RSVP/shagun flows
- Event-scoped RBAC checks for host operations

### System architecture diagram

```text
                           +----------------------+
                           |   Users (Web/Mobile) |
                           +----------+-----------+
                                      |
                                      v
                           +----------------------+
                           |  CDN / Edge / WAF    |
                           |  (Cloudflare)        |
                           +----------+-----------+
                                      |
                    +-----------------+-----------------+
                    |                                   |
                    v                                   v
       +--------------------------+        +---------------------------+
       | Next.js Frontend         |        | Public Asset Delivery     |
       | (Vercel / Web Tier)      |        | (Cloudflare R2 + CDN)     |
       +------------+-------------+        +---------------------------+
                    |
                    | HTTPS API Calls
                    v
       +--------------------------+      +-----------------------------+
       | API Gateway / LB         |----->| Observability               |
       | (Ingress, routing)       |      | (Logs, Metrics, Traces)     |
       +------------+-------------+      +-----------------------------+
                    |
                    v
       +--------------------------+        +---------------------------+
       | Go API Service (Gin)     |<------>| Redis Cluster             |
       | Stateless App Pods       |        | - Cache                   |
       | (horizontal autoscale)   |        | - Rate Limit Counters     |
       +------+---------+---------+        | - Queue Broker            |
              |         |                  +-------------+-------------+
              |         |                                |
              |         |                                v
              |         |                     +-------------------------+
              |         +-------------------->| Async Workers           |
              |                               | (OTP, media, fanout,    |
              |                               | webhooks, exports)       |
              |                               +-----------+-------------+
              |                                           |
              v                                           v
   +--------------------------+                 +------------------------+
   | PostgreSQL (Primary DB)  |<----------------| Write/Update Jobs      |
   | - Transactions           |                 +------------------------+
   | - Core relational data   |
   +------------+-------------+
                |
                v
   +--------------------------+
   | Read Replicas (optional) |
   | for heavy read endpoints |
   +--------------------------+
```

### Data flow and scale notes

- Read-heavy public endpoints use Redis caching before Postgres and serve media directly from R2/CDN.
- Write paths (RSVP, guest operations) persist canonical data in Postgres and enqueue non-critical side effects.
- OTP flow uses Redis for rate limiting and queue-backed async dispatch for provider calls.
- Scaling is independent per tier: stateless API pods, worker pool by queue depth, and optional read replicas for DB-heavy reads.

## Tech Stack

### Frontend

- Next.js 16, React 19, TypeScript
- Tailwind CSS v4
- TanStack Query
- Zod + React Hook Form
- Zustand
- Sentry (`@sentry/nextjs`)

### Backend

- Go + Gin
- PostgreSQL (`pgx`)
- SQL migrations (`golang-migrate`)
- JWT + bcrypt
- Upstash Redis limiter + Redis-backed async OTP queue (Asynq)
- Sentry (`sentry-go`)
- Prometheus metrics + OTel Gin middleware

## Production Hardening Included

- Distributed rate limiting for OTP + RSVP flows
- Async OTP dispatch with retry/circuit-breaker wrapper
- Idempotency-key enforcement for critical write endpoints
- Webhook replay dedupe for Razorpay events
- Structured JSON logs with request/user/error context
- Prometheus endpoint: `/metrics`
- Route-group auth middleware to reduce handler-level auth drift

## Local Validation Commands

### Frontend

```bash
npm run lint
npm run build
npm run test:e2e
```

### Backend

```bash
cd services/api
go test ./...
go vet ./...
```

## API Examples

### Request OTP (host login)

```bash
curl -X POST "http://localhost:8080/v1/auth/otp/request" \
  -H "Content-Type: application/json" \
  -d '{"phone":"+919876543210"}'
```

### Verify OTP (host login)

```bash
curl -X POST "http://localhost:8080/v1/auth/otp/verify" \
  -H "Content-Type: application/json" \
  -d '{"phone":"+919876543210","code":"123456"}' \
  -i
```

### Public RSVP submit (idempotent)

```bash
curl -X POST "http://localhost:8080/v1/public/events/<slug>/rsvp" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <guest_access_token>" \
  -H "Idempotency-Key: rsvp-submit-001" \
  -d '{"items":[{"sub_event_id":"<uuid>","status":"yes"}]}'
```

## Environment Variables

Use:

- Frontend template: `.env.example`
- Backend template: `services/api/.env.example`

Critical production values include:

- `DATABASE_URL`
- `JWT_SECRET`
- `CORS_ORIGIN`
- `AUTH_COOKIE_DOMAIN`
- `OTP_PROVIDER`, `OTP_API_KEY`, `OTP_SENDER_ID`
- `UPSTASH_REDIS_REST_URL`, `UPSTASH_REDIS_REST_TOKEN`
- `REDIS_URL` (async OTP queue)
- `SENTRY_DSN`, `NEXT_PUBLIC_SENTRY_DSN`
- `NEXT_PUBLIC_API_URL`

## Deployment

- Frontend: Vercel (project root `./`)
- Backend: Render or Fly (`render.yaml` / `fly.toml`)
- Migrations: run in deploy startup via backend config (`RUN_MIGRATIONS=true`)

For full rollout steps, use:

- `docs/ROLLOUT_CHECKLIST_HARDENING.md`
- `docs/PRODUCTION_INFRA.md`
- `docs/INFRA_VERIFICATION_CHECKLIST.md`

## Contributing

1. Create a feature branch from `main`.
2. Keep changes scoped (one concern per PR).
3. Run local checks before opening PR:
   - `npm run lint`
   - `npm run build`
   - `cd services/api && go test ./...`
4. Document infra-impacting changes in `docs/`.
5. For API changes, update contracts and examples in the same PR.

## Troubleshooting

- **`go: command not found`**
  - Open a new terminal/session after installing Go and run `go version`.
- **Frontend cannot reach API**
  - Verify `NEXT_PUBLIC_API_URL` and CORS (`CORS_ORIGIN`) match the frontend origin.
- **Auth loops on protected pages**
  - Confirm backend sets `utsav_access_token` and `AUTH_COOKIE_DOMAIN` is correct for your domain setup.
- **`npm run test:e2e` fails locally**
  - Ensure frontend/API are running, then set `E2E_BASE_URL` (default `http://127.0.0.1:3000`).
- **OTP not sending in production**
  - Check `OTP_PROVIDER`, `OTP_API_KEY`, `OTP_SENDER_ID`, and `REDIS_URL` (for async queue mode).
- **Rate limiting behaves like memory-only**
  - Ensure `UPSTASH_REDIS_REST_URL` and `UPSTASH_REDIS_REST_TOKEN` are set.

## License

Proprietary (unless explicitly changed by repository owner).
