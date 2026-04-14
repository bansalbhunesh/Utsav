# 90-second demo script (Phase 0–1)

1. **Show stack**: `docker compose -f infra/docker/compose.yml up -d`, then run API + web per `README.md`.
2. **Health**: open web home, click **Ping API** — expect `API health: 200` when rewrites hit `GET /v1/healthz`.
3. **Auth**: `POST /v1/auth/otp/request` then `POST /v1/auth/otp/verify` with dev code `123456` — receive `access_token`.
4. **Create event**: `POST /v1/events` with `{ "slug": "...", "title": "...", "host_upi_vpa": "host@upi" }` — slug appears in `GET /v1/public/events/:slug`.
5. **Narrate roadmap**: guests, RSVP guest JWT, cash logger, Memory Book generation — all tables exist; wire UI incrementally.
