# Next steps (engineering)

Ordered for the spec’s “core guest path” and hackathon demo depth.

1. **Ops**: Keep [LOCAL_DEV.md](./LOCAL_DEV.md) accurate; add production `Dockerfile` for API + web when you pick a host.
2. **Auth hardening**: ~~In-memory rate limits on host + guest RSVP OTP request (5/15m per IP; 10/15m per IP+slug+phone).~~ Next: MSG91 adapter, shorter JWT TTL + ADR.
3. **Guest list**: ~~`POST /v1/events/:id/guests/import` + Guests page CSV paste/file.~~ Next: async job + larger files; duplicate merge UI.
4. **Invites**: ~~`/e/[slug]/invite` animated shell + host copy/WhatsApp on event detail; `inviteShare` helpers.~~ Next: richer art (tier assets), per-guest deep links.
5. **Gallery**: ~~Presign upload + moderation queue UI (`/events/:id/gallery`) with object store URL adapter.~~ Next: signed PUT via S3 SDK (R2/MinIO creds) + thumbnail pipeline.
6. **Memory Book**: ~~Aggregate payload generation + tier-gated PDF export stub with host UI (`/events/:id/memory-book`).~~ Next: themed renderer + persisted PDF URL.
7. **Broadcasts**: ~~Segment builder UI + host broadcasts list (`/events/:id/broadcasts`) matching `audience` JSON.~~ Next: SMS provider adapter and delivery tracking.
8. **Organiser**: ~~Client CRUD + link events endpoints + web console (`/organiser`).~~ Next: team-member roles and client activity timeline.
9. **Billing**: ~~Checkout order payload + Razorpay webhook signature verify + event tier entitlement update.~~ Next: real Razorpay Orders API call + retries/idempotency log.
10. **Tests**: ~~API unit + integration (`-tags=integration`) + Playwright smoke with CI jobs (`api-integration`, `e2e-smoke`).~~ Next: stabilize flaky retries/reporting and add seeded fixtures.
