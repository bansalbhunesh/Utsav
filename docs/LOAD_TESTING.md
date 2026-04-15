# Load Testing Guide (k6 + Grafana)

This guide runs a realistic mixed workload against the UTSAV API and provides practical Grafana query snippets for analysis.

## 1) Prerequisites

- k6 installed locally
- Backend deployed and reachable
- Valid event slug and sub-event IDs for RSVP tests
- Guest token for RSVP submit tests
- Test OTP code enabled only in non-production environments

## 2) Run the mixed workload test

From repo root:

```bash
k6 run \
  -e BASE_URL=https://api.yourdomain.com \
  -e EVENT_SLUG=your-event-slug \
  -e SUB_EVENT_IDS=uuid1,uuid2 \
  -e GUEST_TOKEN=eyJ... \
  -e TEST_OTP_CODE=123456 \
  -e PHONE_PREFIX=+9199000 \
  loadtests/k6-mixed-workload.js
```

## 3) What this script simulates

- High read traffic on:
  - `/v1/public/events/:slug`
  - `/v1/public/events/:slug/schedule`
- Spike write traffic on:
  - `/v1/public/events/:slug/rsvp`
- Continuous OTP pressure on:
  - `/v1/auth/otp/request`
  - `/v1/auth/otp/verify`

## 4) k6 thresholds baked in

- `http_req_failed < 3%`
- `public_event` p95 < 250ms
- `public_schedule` p95 < 250ms
- `rsvp_submit` p95 < 700ms
- `otp_request` p95 < 700ms
- `otp_verify` p95 < 900ms

Tune thresholds according to your SLOs.

## 5) Grafana/Prometheus query cheat-sheet

Use these against Prometheus scraping `/metrics`.

### Request rate (RPS)

```promql
sum(rate(utsav_http_requests_total[1m]))
```

### RPS by endpoint

```promql
sum by (route) (rate(utsav_http_requests_total[1m]))
```

### Error rate (%)

```promql
100 *
sum(rate(utsav_http_requests_total{status=~"5.."}[5m])) /
sum(rate(utsav_http_requests_total[5m]))
```

### Error rate by endpoint

```promql
100 *
sum by (route) (rate(utsav_http_requests_total{status=~"5.."}[5m])) /
sum by (route) (rate(utsav_http_requests_total[5m]))
```

### p95 latency (all traffic)

```promql
histogram_quantile(
  0.95,
  sum by (le) (rate(utsav_http_request_duration_seconds_bucket[5m]))
)
```

### p95 latency by endpoint

```promql
histogram_quantile(
  0.95,
  sum by (route, le) (rate(utsav_http_request_duration_seconds_bucket[5m]))
)
```

### p99 latency by endpoint

```promql
histogram_quantile(
  0.99,
  sum by (route, le) (rate(utsav_http_request_duration_seconds_bucket[5m]))
)
```

## 6) Result interpretation quick guide

- High p95 on read endpoints + low CPU on API + high DB load:
  - likely missing cache/index efficiency.
- High RSVP p95 + rising write errors:
  - write contention/idempotency misuse/hot rows.
- High OTP latency + stable DB:
  - queue/provider bottleneck.
- Sustained 5xx after spike:
  - pool saturation or slow recovery.

## 7) Recommended test progression

1. Run once at low scale to validate env and data correctness.
2. Run baseline (30 min).
3. Run spike profile.
4. Run stress-to-failure.
5. Capture max sustainable RPS and first breaking component.
