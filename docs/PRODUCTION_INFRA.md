# UTSAV Production Infrastructure Protocol

This guide defines a deploy-now, production-grade setup for UTSAV with managed services, explicit environment variables, and clear operational guardrails.

## 1) Recommended production stack

### Frontend
- Hosting: **Vercel**
- Domain + DNS/CDN/WAF: **Cloudflare**

### Backend API
- Hosting: **Render Web Service** (or Fly.io equivalent)
- Runtime: Go service from `services/api/cmd/server`

### Database
- **Neon PostgreSQL** (production branch + connection pooling)

### Media/Object Storage
- **Cloudflare R2** (S3-compatible bucket + public CDN URL)

### Auth / OTP
- OTP provider: **MSG91** (India-first) or **Twilio** fallback
- Important: keep `DEV_OTP_CODE` disabled in production

### Payments
- **Razorpay** (webhook configured)

### Monitoring / Logs / Uptime
- Error tracking: **Sentry**
- Structured logs: **Logtail** or **Datadog**
- Uptime checks: **Better Stack** or **Pingdom**

### Rate limiting / shared cache (phase 1.5)
- **Upstash Redis** for distributed rate limiting

---

## 2) Mandatory environment variables

## Frontend (Vercel)

File reference: `.env.example`

```bash
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

```bash
NEXT_PUBLIC_SENTRY_DSN=https://<frontend-public-dsn>
```

## Backend (Render / Fly)

File reference: `services/api/.env.example`

### Required

```bash
HTTP_PORT=8080
DATABASE_URL=postgres://USER:PASSWORD@HOST/DB?sslmode=require
MIGRATIONS_PATH=../../db/migrations
JWT_SECRET=<secure-random-min-32-char-secret>
CORS_ORIGIN=https://app.yourdomain.com
RUN_MIGRATIONS=true
ENV=production
```

### Strongly recommended

```bash
DEV_OTP_CODE=
OTP_MAX_ATTEMPTS=5
OBJECT_STORE_PUBLIC_BASE_URL=https://cdn.yourdomain.com
OBJECT_STORE_BUCKET=utsav-prod
OBJECT_STORE_REGION=auto
RAZORPAY_KEY_ID=
RAZORPAY_WEBHOOK_SECRET=
LOG_LEVEL=info
```

### Observability and reliability

```bash
SENTRY_DSN=https://<backend-dsn>
BETTERSTACK_HEARTBEAT_URL=https://uptime.betterstack.com/api/v1/heartbeat/...
LOGTAIL_SOURCE_TOKEN=
UPSTASH_REDIS_REST_URL=https://<upstash-endpoint>
UPSTASH_REDIS_REST_TOKEN=<upstash-token>
RATE_LIMIT_WINDOW=900
AUTH_OTP_REQUEST_LIMIT=5
AUTH_OTP_VERIFY_LIMIT=10
RSVP_OTP_REQUEST_LIMIT=10
RSVP_OTP_VERIFY_LIMIT=20
PUBLIC_RSVP_SUBMIT_LIMIT=30
```

---

## 3) Service accounts you must create

1. **Vercel**
   - Project for Next.js app
   - Environment groups: preview + production
2. **Render**
   - Web Service for Go API
   - Health check configured to `/v1/healthz`
3. **Neon**
   - Production DB + pooled connection endpoint
   - Automated backups enabled
4. **Cloudflare**
   - DNS + SSL/TLS + WAF
   - (Optional) CDN/proxy in front of API
5. **Cloudflare R2**
   - Bucket: `utsav-prod`
   - Public base URL (or CDN custom domain)
6. **MSG91 or Twilio**
   - Production OTP sender
7. **Razorpay**
   - Key + webhook secret
8. **Sentry + Logtail/Datadog + BetterStack**
   - Error, logs, uptime/alerting
9. **Upstash Redis** (phase 1.5)
   - Centralized distributed rate limiting

---

## 4) Deployment steps (Vercel + Render + Neon)

## A. Neon setup

1. Create Neon project + production branch.
2. Copy pooled `DATABASE_URL` with SSL required.
3. Enable backups and branch protection.

## B. Render API deployment

1. Create new Web Service from this repo.
2. Root directory: `services/api`
3. Build command:

```bash
go build -o bin/api ./cmd/server
```

4. Start command:

```bash
./bin/api
```

5. Add backend env vars from `services/api/.env.example` (production values).
6. Set health check path: `/v1/healthz`
7. Confirm readiness endpoint: `/v1/readyz`
8. Validate CORS with frontend domain only.

## C. Vercel frontend deployment

1. Import repo on Vercel.
2. Root directory: repository root.
3. Add env var:

```bash
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

4. Deploy and test authenticated/guest flows.

## D. Cloudflare domain + TLS

1. Point `app.yourdomain.com` -> Vercel.
2. Point `api.yourdomain.com` -> Render.
3. Enforce HTTPS and add baseline WAF rules.

---

## 5) Production hardening checklist

## Security
- [ ] `JWT_SECRET` is strong and private.
- [ ] `DEV_OTP_CODE` is blank in production.
- [ ] `CORS_ORIGIN` is exact frontend URL (no wildcard).
- [ ] TLS enforced on app/api domains.
- [ ] Webhook secrets validated (Razorpay).

## Reliability
- [ ] Health checks configured.
- [ ] Auto-restart on failure enabled.
- [ ] DB backups and restore drill documented.
- [ ] Runbook for OTP/payment provider downtime.

## Scalability
- [ ] API stateless deployment (horizontal scale ready).
- [ ] Neon pooled connection string used.
- [ ] Media served via CDN/public object URL.
- [ ] Redis distributed limiter planned/implemented.

## Observability
- [ ] Sentry captures backend + frontend exceptions.
- [ ] Structured logs centralized.
- [ ] Uptime probes alert on app/api/critical endpoints.

---

## 6) Load and resilience validation plan

Run pre-launch tests for:

- 1000+ concurrent public event page views.
- OTP request bursts and verification spikes.
- RSVP write contention (same event, many guests).
- Gallery upload registration bursts.
- Dashboard query bursts during live event.

Success criteria:

- No 5xx spikes under expected peak.
- p95 latency acceptable for public pages and RSVP writes.
- No data consistency regressions (guest list, RSVP state, shagun entries).
- No auth/session corruption under refresh load.

---

## 7) Infra status and closure

Implemented in this repository:

1. **OTP provider path**
   - MSG91 sender integration added (`services/api/internal/otp/provider.go`).
   - Production guard added: `DEV_OTP_CODE` must be disabled in production.
   - OTP generation + bcrypt hashing + expiry/attempt checks are enforced in services/repositories.
2. **Distributed rate limiting path**
   - Upstash REST limiter added (`services/api/internal/ratelimit/distributed.go`).
   - Wired for OTP request/verify in auth + RSVP services.
   - In-memory fallback remains for local development when Upstash env is not set.
3. **Observability baseline**
   - Backend Sentry middleware wiring added in API startup (`sentry-go`).
   - Structured JSON request logging in middleware with `request_id`.
4. **Deployment manifests**
   - `render.yaml`, `fly.toml`, `services/api/Dockerfile`, `vercel.json` added.

Closed in-repo:

- Frontend Sentry is wired via `@sentry/nextjs`, startup instrumentation, and `app/error.tsx` capture.
- Backend structured JSON logs include `request_id`, `user_id`, `guest_id`, endpoint path, and error code context.
- OTP expiration is 5 minutes with persisted attempt lockout (`OTP_MAX_ATTEMPTS`) and production guard for `DEV_OTP_CODE`.
- Distributed rate limiting is enforced via Upstash for auth OTP request/verify, RSVP OTP request/verify, and public RSVP submit.
- Deployment manifests include health checks, baseline horizontal scale, and required env injection.

---

## 8) Quick go-live minimum (if launching now)

If you need to launch immediately with minimal risk:

- Use Neon + Render + Vercel + Cloudflare.
- Set strong secrets and strict CORS.
- Keep `RUN_MIGRATIONS=true` initially, then move to controlled migration job.
- Keep payment webhooks off until end-to-end tested in staging.
- Enable uptime checks and alerting before public launch.

This gets you to a robust "works in production" baseline while preserving a clean path to scale.
