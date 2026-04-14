# UTSAV

India's event operating system for weddings and social celebrations.

UTSAV replaces fragmented event workflows (WhatsApp groups, paper shagun logs, scattered guest data, ad hoc planner coordination) with a single product across host, organiser, and guest experiences.

---

## Why UTSAV Exists

Indian events, especially weddings, are operationally heavy:

- Guest communication is fragmented across messages/calls.
- RSVP tracking and sub-event planning are manual.
- Shagun tracking is often paper-based and error-prone.
- Media collection and post-event memory curation are inconsistent.
- Organisers manage multiple events without a unified operations layer.

UTSAV solves this by providing one backend-powered system for:

- Event setup and role-based collaboration
- Guest lifecycle (invite -> RSVP -> attendance signals)
- Shagun logging/reporting
- Gallery workflows
- Broadcast communication
- Memory-book generation
- Organiser client/event management

---

## Product Vision

UTSAV is designed as an event OS, not a single-feature tool.

- Host surface: create/manage events, guests, RSVP, finances, gallery, broadcasts
- Organiser surface: manage clients and link clients to events
- Guest surface (PWA-like web): view event page, RSVP via OTP, report shagun, browse gallery and schedule
- System goal: make event operations reliable, auditable, and repeatable

---

## Problem -> Solution Mapping

- Problem: no single source of truth
  - Solution: API-first architecture and centralized event data model.
- Problem: RSVP chaos
  - Solution: OTP-based guest auth + structured RSVP APIs per event/sub-event.
- Problem: shagun record inconsistency
  - Solution: host cash logging + guest UPI-report metadata endpoints with event-bound tokens.
- Problem: communication scatter
  - Solution: host broadcast module and public event feeds.
- Problem: post-event memory loss
  - Solution: generated memory books and public memory endpoints.

---

## Why This Product Is Strong

- India-native flow design: phone OTP, WhatsApp-share-first UX, UPI-aware shagun reporting.
- Operational depth: goes beyond discovery into real execution workflows.
- Role-aware system: owner/co-owner/organiser/contributor/vendor/guest boundaries.
- Clear API contracts: backend structured error envelopes and frontend runtime contract parsing.
- Scalable architecture: strict Handler -> Service -> Repository layering in Go backend.

### Why These Technical Choices

- **Next.js + React**: fast iteration for complex host/guest UX and SEO-friendly public pages.
- **Go + Gin API**: strong performance and low runtime overhead for high-request event windows.
- **PostgreSQL + SQL migrations**: predictable relational modeling for guests, RSVP, roles, and finance records.
- **JWT + rotating refresh tokens**: stateless access control plus recoverable session UX.
- **Zod runtime contracts**: protects frontend from malformed API payloads and integration drift.
- **React Query**: stable data-fetching, cache invalidation, and mutation ergonomics across host pages.

---

## How It Compares

UTSAV is positioned as an operations system, not just a marketplace/discovery tool.

- Discovery platforms help users find vendors.
- UTSAV helps teams run the event itself: guests, RSVP, communications, shagun records, media moderation, memory outcomes.
- This makes UTSAV complementary to vendor marketplaces and stronger in day-to-day event execution.

### Competitive Positioning Summary

- **Vendor discovery apps**: strong pre-event sourcing; weak intra-event execution tooling.
- **Generic planning/checklist apps**: partial planning support; limited India-specific shagun/guest flows.
- **UTSAV**: event execution core with host + organiser + guest surfaces under one data model.

---

## MVP Scope in This Repository

Implemented modules (core):

- Auth (host OTP + refresh; guest OTP for RSVP flow)
- Event CRUD + sub-events + members
- Guest management + CSV import
- RSVP (public flow + host visibility)
- Shagun (host cash logger + guest report)
- Gallery (presign/register/list/moderation + public listing)
- Broadcasts (host create/list + public listing)
- Memory book generation + public retrieval + export gating
- Organiser profile/clients/client-event linking
- Billing checkout/webhook stubs + tier transitions

Frontend includes:

- Host dashboard and per-event management pages
- Guest public event pages and guest actions
- Organiser console
- API client with token refresh flow
- Route protection via Next.js `proxy.ts`

---

## Architecture

### Monorepo Structure

- `app/` - Next.js app routes (host, organiser, guest pages)
- `components/` - UI and domain components
- `lib/` - API client + schema contracts + utilities
- `services/api/` - Go API service
- `db/migrations/` - PostgreSQL schema migrations
- `infra/docker/compose.yml` - local Postgres

### Backend Design

The API follows a strict layered pattern:

- Handler layer (`services/api/internal/httpserver`)
  - Parses requests, enforces HTTP semantics, returns responses.
- Service layer (`services/api/internal/service`)
  - Business rules, validation, authorization decisions, orchestration.
- Repository layer (`services/api/internal/repository`)
  - Data access only.

This separation reduces coupling, improves testability, and keeps transport logic out of domain logic.

### Frontend <-> Backend Contract

- Frontend uses `lib/api.ts` for all API calls.
- Structured backend error envelopes are parsed into typed `ApiError`.
- Runtime response validation is handled via Zod contract parsers in `lib/contracts`.

---

## Core User Flows

### 1) Host Flow

1. Login with OTP
2. Create event + configure sub-events
3. Add/import guests
4. Share guest-facing links
5. Track RSVPs and shagun
6. Manage gallery and broadcasts
7. Generate memory book

### 2) Guest Flow

1. Open event link
2. Request/verify RSVP OTP
3. Submit RSVP with preferences
4. Report shagun payment (if enabled and invited)
5. View schedule/gallery/broadcasts

### 3) Organiser Flow

1. Create organiser profile
2. Create clients
3. Link clients to accessible events
4. Track organiser-scoped event/client data

---

## API Surface (High-Level)

All routes are mounted under `/v1`.

- Health: `/healthz`, `/readyz`
- Auth: `/auth/otp/request`, `/auth/otp/verify`, `/auth/refresh`, `/me`
- Events (host): create/list/get/patch + sub-events + members
- Guests/Vendors/Shagun/RSVP (host): event-scoped operational APIs
- Gallery/Broadcast/Memory (host): event-scoped creation/moderation/generation
- Public: event details, schedule, gallery, broadcasts, RSVP OTP/submit, shagun report, memory book
- Organiser: profile, clients, client-event linking
- Billing: checkout/list/webhook

For the exact route list, see `services/api/internal/httpserver/router.go`.

---

## Security and Reliability

- OTP-authenticated host and guest flows.
- JWT-based host access; guest token bound to event+phone context.
- Event-level RBAC checks in backend.
- Structured API error envelopes for consistent client handling.
- Rate limiting windows for OTP endpoints.
- Refresh-token rotation with hashed token persistence.

---

## Local Development

### Prerequisites

- Node.js 20+
- npm
- Go 1.25+
- Docker Desktop (for local Postgres)

### 1) Start Postgres

```bash
make docker-up
```

### 2) Run API

```bash
cd services/api
MIGRATIONS_PATH=../../db/migrations \
DATABASE_URL=postgres://utsav:utsav@127.0.0.1:5432/utsav?sslmode=disable \
HTTP_PORT=8080 \
JWT_SECRET=dev-insecure-change-me \
go run ./cmd/server
```

### 3) Run Web

```bash
cd ../..
npm install
npm run dev
```

Open `http://localhost:3000`.

Set frontend API base (if needed):

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## Build, Lint, Test

### Frontend

```bash
npm run lint
npm run build
```

### Backend

```bash
cd services/api
go test ./...
```

Integration tests exist under `services/api/internal/httpserver/integration_test.go` (Docker/testcontainers-based).

---

## Environment Configuration

### Frontend

- `NEXT_PUBLIC_API_URL` (default: `http://localhost:8080`)

### Backend (key vars)

- `HTTP_PORT` (default `8080`)
- `DATABASE_URL` (default local Postgres DSN)
- `MIGRATIONS_PATH` (default `../../db/migrations`)
- `JWT_SECRET` (set strong secret in production)
- `DEV_OTP_CODE` (dev only)
- `CORS_ORIGIN` (default `http://localhost:3000`)
- `RUN_MIGRATIONS` (default `true`)
- `OBJECT_STORE_PUBLIC_BASE_URL`
- `RAZORPAY_KEY_ID`
- `RAZORPAY_WEBHOOK_SECRET`

---

## Product Impact and Business Potential

- Consumer-side pain reduction: less coordination overhead for families and hosts.
- Planner productivity: organiser workflows become repeatable and auditable.
- Social trust layer: structured RSVP + shagun records improve event confidence.
- Distribution leverage: every event exposes future hosts/organisers to the system.
- Long-term moat potential: cross-event relationship and transaction metadata.

### Monetization Direction (From Product Strategy)

- **B2C per-event tiers**: free entry + paid unlocks for higher-scale and advanced controls.
- **B2B organiser subscriptions**: recurring revenue for multi-event professionals.
- **Future expansion**: vendor marketplace commissions, premium templates, memory export add-ons.

### GTM Direction (From Product Strategy)

- Start with real events and organiser-led adoption.
- Use memory-book sharing and guest exposure as an organic growth loop.
- Expand city-by-city once organiser and vendor density reaches critical thresholds.

---

## Current Product Status

- Core host/guest/organiser workflows are implemented.
- Frontend build and lint pass.
- API architecture is converged to layered design.
- Remaining production readiness tasks are operational:
  - CI/CD hardening and deployment environments
  - Production secrets management
  - Observability dashboards/alerts
  - Broader automated test coverage across all domains

---

## Roadmap (Post-MVP)

- Advanced organiser analytics
- Vendor marketplace depth
- Expanded invitation/memory premium features
- AI-assisted planning and recommendations
- Cross-event intelligence enhancements

---

## License

Proprietary (unless explicitly changed by repository owner).
