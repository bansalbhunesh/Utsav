# UTSAV

[![CI](https://github.com/bansalbhunesh/Utsav/actions/workflows/ci.yml/badge.svg)](https://github.com/bansalbhunesh/Utsav/actions/workflows/ci.yml)
[![Next.js](https://img.shields.io/badge/Next.js-16-black)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Production-blue)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/license-Proprietary-lightgrey)](#license)

UTSAV is a production-grade event operating system for high-stakes celebrations.  
It combines event operations (OTP auth, RSVP, guest, shagun, gallery, broadcast, billing) with an intelligence layer that helps hosts make better decisions under real-world traffic and time pressure.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Scalability Features](#scalability-features)
- [Key Innovations](#key-innovations)
- [Setup Instructions](#setup-instructions)
- [Validation Commands](#validation-commands)
- [API Examples](#api-examples)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)
- [License](#license)

## Overview

UTSAV is built as a full-stack monorepo with a Next.js frontend and Go backend, designed for:

- Reliable host and guest authentication using OTP with secure cookies.
- Strong consistency for critical event operations (RSVP, guest updates, payments).
- Production observability for incident response (logs, metrics, traces, Sentry).
- Infrastructure-level safety (distributed rate limits, idempotency, webhook dedupe).
- Decision support via intelligence modules (Relationship Priority Score live; RSVP Risk and Shagun Signal coming).

Monorepo structure:

- `app/` - Next.js App Router UI (host, organiser, public guest)
- `services/api/` - Go API service (`handler -> service -> repository`)
- `db/migrations/` - SQL schema and infra migrations
- `lib/` - API clients, contracts, shared utilities
- `.github/workflows/ci.yml` - CI pipeline

## Architecture

### System Design

```text
Users (Host / Guest / Organiser)
        |
        v
Next.js Frontend (Vercel, frontend root)
        |
        v
Go API (Gin, stateless)
  |        |          |
  |        |          +--> Sentry + Prometheus + OTel
  |        |
  |        +--> Redis (Upstash): distributed rate limits + async queue broker
  |
  +--> PostgreSQL (Neon): source of truth, transactional writes, indexed reads
```

### Backend layering

- `httpserver` handles protocol, auth middleware, request parsing, error envelopes.
- `service` contains business and intelligence logic.
- `repository` isolates SQL, transactions, and persistence patterns.

### Runtime model

- Public endpoints (`/v1/public/...`) are traffic-heavy and protected with persistent limits.
- Authenticated host endpoints (`/v1/events/:id/...`) use event-scoped RBAC.
- Non-critical external side effects (OTP dispatch) run via async queue patterns.

## Tech Stack

### Frontend

- Next.js 16 (App Router), React 19, TypeScript
- Tailwind CSS v4
- TanStack Query for server state
- Zod contracts for runtime safety
- Sentry (`@sentry/nextjs`)

### Backend

- Go 1.25+, Gin
- PostgreSQL via `pgx`
- SQL migrations via `golang-migrate`
- Redis / Upstash (rate limiting + queue infra)
- Asynq for background OTP dispatch
- JWT + bcrypt auth
- Prometheus metrics + OpenTelemetry middleware + `sentry-go`

### Infrastructure

- Frontend deploy: Vercel (`frontend` root in project config)
- Backend deploy: Render/Fly (`render.yaml`, `fly.toml`)
- Media: Cloudflare R2 (integration-ready)
- CI: GitHub Actions

## Scalability Features

UTSAV includes practical scale controls needed for high-concurrency wedding/event workloads:

- **Distributed rate limiting:** OTP request/verify and public RSVP routes use Redis-backed limits.
- **Idempotent writes:** critical POST flows require `Idempotency-Key` to prevent duplicate execution.
- **Async external calls:** OTP provider delivery is queue-dispatched with retry/circuit-breaker behavior.
- **Webhook dedupe:** replay-safe delivery keys for payment webhooks.
- **Pagination everywhere:** host list surfaces (`guests`, `rsvps`, `shagun`) support `limit/offset`.
- **Scale-oriented indexing:** focused indexes on RSVP, OTP, guests, events, shagun query paths.
- **Stateless API:** horizontal scaling via multiple backend instances with shared Redis/Postgres state.

## Key Innovations

### 1) Relationship Priority Score (live)

A production intelligence feature that ranks guests by operational importance for event success:

- Weighted model with normalized score (0-100)
- Recency decay + uncertainty handling for sparse data
- Tiers: `Critical`, `Important`, `Optional`
- Demo output: top guests to personally call, guests needing attention, colored tier cards

### 2) Reliability-first infra decisions

- Persistent distributed controls over in-memory shortcuts
- Hard fail-safe behavior in production config
- Observability-first instrumentation for fast incident triage

### 3) Decision-system direction

UTSAV is built to evolve from a workflow app into an intelligence-backed operating system:

- RSVP Risk Predictor (coming)
- Shagun Signal Intelligence (coming)

## Setup Instructions

### Prerequisites

- Node.js 20+
- npm
- Go 1.25+
- Docker Desktop

### 1) Start local dependencies

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

## Validation Commands

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

### Request OTP

```bash
curl -X POST "http://localhost:8080/v1/auth/otp/request" \
  -H "Content-Type: application/json" \
  -d '{"phone":"+919876543210"}'
```

### Verify OTP

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

### Relationship Priority Overview

```bash
curl -X GET "http://localhost:8080/v1/events/<event_id>/intelligence/relationship-priority" \
  -H "Authorization: Bearer <host_access_token>"
```

## Deployment

- Frontend: Vercel (configured to deploy frontend app root)
- Backend: Render/Fly with health check on `/health`
- Migrations: backend startup-managed (`RUN_MIGRATIONS=true`) or release phase

Production docs:

- `docs/PRODUCTION_INFRA.md`
- `docs/ROLLOUT_CHECKLIST_HARDENING.md`
- `docs/INFRA_VERIFICATION_CHECKLIST.md`

## Troubleshooting

- **Go not found after install**
  - Open a new terminal session and run `go version`.
- **Frontend cannot call API**
  - Verify `NEXT_PUBLIC_API_URL` and backend `CORS_ORIGIN`.
- **Auth loops**
  - Verify `AUTH_COOKIE_DOMAIN` and cookie settings for your host.
- **OTP not delivered**
  - Validate `OTP_PROVIDER`, `OTP_API_KEY`, `OTP_SENDER_ID`, `REDIS_URL`.
- **Rate limits appear non-persistent**
  - Validate `UPSTASH_REDIS_REST_URL` and `UPSTASH_REDIS_REST_TOKEN`.

## License

Proprietary (unless explicitly changed by repository owner).
