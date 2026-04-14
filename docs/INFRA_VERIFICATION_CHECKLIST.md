# UTSAV Infra Verification Checklist

Use this checklist after staging/production deploys.

## 1) OTP Provider (MSG91) Validation

- [ ] `ENV=production` and `DEV_OTP_CODE=` are set.
- [ ] `OTP_PROVIDER=msg91`, `OTP_API_KEY`, and `OTP_SENDER_ID` are configured.
- [ ] `POST /v1/auth/otp/request` sends OTP to a real phone number.
- [ ] OTP expires after 5 minutes.
- [ ] OTP verify fails after `OTP_MAX_ATTEMPTS`.
- [ ] OTP values are hashed in DB (`phone_otp_challenges.code_hash`, `rsvp_otp_challenges.code_hash`).

## 2) Distributed Rate Limiting Validation

- [ ] `UPSTASH_REDIS_REST_URL` and `UPSTASH_REDIS_REST_TOKEN` are configured.
- [ ] Auth OTP request/verify are limited by IP+phone.
- [ ] Public RSVP OTP request/verify are limited by IP+phone.
- [ ] Public RSVP submit endpoint is limited.
- [ ] Restart API process and confirm limits still hold (Redis-backed, not process memory).

Load smoke command (adjust URL and payload):

```bash
for i in $(seq 1 50); do curl -s -o /dev/null -w "%{http_code}\n" \
  -X POST "https://api.yourdomain.com/v1/auth/otp/request" \
  -H "content-type: application/json" \
  -d '{"phone":"919999999999"}'; done
```

## 3) Observability Validation

- [ ] Backend Sentry receives captured API exceptions (`SENTRY_DSN`).
- [ ] Frontend Sentry receives client exception from `app/error.tsx`.
- [ ] API logs are JSON and include `request_id`, `user_id`/`guest_id`, `endpoint`, `error_code`.
- [ ] `/health` and `/v1/readyz` return expected status.
- [ ] BetterStack monitor (or equivalent) is configured against `/health`.
- [ ] `BETTERSTACK_HEARTBEAT_URL` (if used) receives heartbeat pings.

## 4) Deployment Reproducibility Validation

- [ ] Render/Fly deploy is done from repo config (`render.yaml`/`fly.toml`) without manual runtime edits.
- [ ] Vercel deploy uses `.env.example` keys including `NEXT_PUBLIC_API_URL` and `NEXT_PUBLIC_SENTRY_DSN`.
- [ ] API service starts with migrations and passes health checks.
- [ ] Horizontal scale baseline enabled (Render `numInstances: 2`).
- [ ] New developer can deploy from zero by following `docs/PRODUCTION_INFRA.md`.

## 5) Wedding-Day Reliability Smoke

- [ ] Burst test OTP requests (50 in short window) and verify rate limits trigger.
- [ ] Submit multiple RSVP payloads concurrently for same event; data remains consistent.
- [ ] Trigger one backend error intentionally and verify Sentry alert + log entry.
- [ ] Restart backend during test traffic; rate limiting and auth remain intact.
