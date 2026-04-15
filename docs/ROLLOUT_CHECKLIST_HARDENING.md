# UTSAV Hardening Rollout Checklist

Use this checklist to roll out the latest reliability and scale hardening safely.

## 1) Pre-deploy checks

- [ ] Confirm production secrets are set in backend host:
  - `ENV=production`
  - `JWT_SECRET`
  - `DATABASE_URL`
  - `CORS_ORIGIN`
  - `AUTH_COOKIE_DOMAIN` (e.g. `.yourdomain.com` if sharing across subdomains)
  - `OTP_PROVIDER=msg91`
  - `OTP_API_KEY`
  - `OTP_SENDER_ID`
  - `UPSTASH_REDIS_REST_URL`
  - `UPSTASH_REDIS_REST_TOKEN`
  - `REDIS_URL` (required for async OTP queue worker/dispatcher)
  - `SENTRY_DSN`
- [ ] Confirm frontend env is set:
  - `NEXT_PUBLIC_API_URL`
  - `NEXT_PUBLIC_SENTRY_DSN`
- [ ] Confirm health endpoint allowed by LB:
  - `/health`
- [ ] Confirm Prometheus scrape can reach:
  - `/metrics`

## 2) Database migration order

Apply migrations in order (do not skip):

1. `000004_scale_indexes.up.sql`
2. `000005_idempotency_keys.up.sql`
3. `000006_webhook_deliveries.up.sql`

Validation queries:

```sql
SELECT to_regclass('public.idempotency_keys');
SELECT to_regclass('public.webhook_deliveries');
```

## 3) Deploy order

1. Deploy backend first.
2. Verify backend healthy and ready:
   - `GET /health` => 200
   - `GET /v1/readyz` => 200
3. Deploy frontend second (cookie auth-compatible client).

## 4) Post-deploy smoke tests

### Auth/cookie flow

- [ ] Request host OTP: `POST /v1/auth/otp/request`
- [ ] Verify host OTP: `POST /v1/auth/otp/verify`
- [ ] Confirm browser receives httpOnly cookies:
  - `utsav_access_token`
  - `utsav_refresh_token`
- [ ] Reload protected route (`/dashboard`) and confirm session persists.
- [ ] Logout endpoint clears cookies: `POST /v1/auth/logout`.

### Idempotency checks

- [ ] Repeat same request + same `Idempotency-Key`:
  - `POST /v1/public/events/:slug/rsvp`
  - `POST /v1/billing/checkout`
  - `POST /v1/events/:id/guests`
  - `POST /v1/events/:id/cash-shagun`
  - `POST /v1/events/:id/broadcasts`
- [ ] Confirm no duplicate writes.

### OTP pipeline checks

- [ ] With `REDIS_URL` set, OTP requests enqueue and send successfully.
- [ ] Simulate provider instability and confirm retries/backoff behavior.

### Webhook reliability checks

- [ ] Send same Razorpay webhook event twice:
  - first should process
  - second should dedupe safely
- [ ] Confirm event tier update still applied once.

### Metrics/Tracing checks

- [ ] `GET /metrics` returns Prometheus series:
  - `utsav_http_requests_total`
  - `utsav_http_request_duration_seconds`
- [ ] Traces visible in configured OTel pipeline (if exporter attached at infra/runtime layer).

## 5) Quick rollback plan

If severe production regression occurs:

1. Roll back backend deployment image/version.
2. Keep DB migrations in place (they are additive and safe).
3. If needed, disable async OTP by unsetting `REDIS_URL` to fall back to direct dispatch.
4. Keep idempotency protections enabled.

## 6) Final acceptance criteria

- [ ] No auth token in frontend localStorage for host session.
- [ ] API protections enforced by route middleware groups.
- [ ] Webhook replay does not duplicate side effects.
- [ ] P95 latency stable under expected load.
- [ ] Error rate and request metrics visible in monitoring stack.
