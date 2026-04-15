import http from "k6/http";
import { check, sleep } from "k6";
import { randomItem } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const EVENT_SLUG = __ENV.EVENT_SLUG || "demo-wedding";
const GUEST_TOKEN = __ENV.GUEST_TOKEN || "";
const OTP_CODE = __ENV.TEST_OTP_CODE || "123456";
const PHONE_PREFIX = __ENV.PHONE_PREFIX || "+9199000";
const SUB_EVENT_IDS = (__ENV.SUB_EVENT_IDS || "").split(",").filter(Boolean);

if (SUB_EVENT_IDS.length === 0) {
  throw new Error("SUB_EVENT_IDS env var is required, comma-separated UUIDs");
}

export const options = {
  scenarios: {
    public_read_normal: {
      executor: "ramping-arrival-rate",
      startRate: 30,
      timeUnit: "1s",
      preAllocatedVUs: 100,
      maxVUs: 800,
      stages: [
        { duration: "5m", target: 120 },
        { duration: "10m", target: 120 },
        { duration: "5m", target: 20 },
      ],
      exec: "publicReadScenario",
    },
    rsvp_spike: {
      executor: "ramping-arrival-rate",
      startRate: 10,
      timeUnit: "1s",
      preAllocatedVUs: 200,
      maxVUs: 2000,
      stages: [
        { duration: "2m", target: 40 },
        { duration: "1m", target: 300 },
        { duration: "8m", target: 300 },
        { duration: "2m", target: 20 },
      ],
      exec: "rsvpScenario",
    },
    otp_flow: {
      executor: "ramping-arrival-rate",
      startRate: 5,
      timeUnit: "1s",
      preAllocatedVUs: 100,
      maxVUs: 1000,
      stages: [
        { duration: "3m", target: 25 },
        { duration: "7m", target: 60 },
        { duration: "3m", target: 5 },
      ],
      exec: "otpScenario",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.03"],
    "http_req_duration{endpoint:public_event}": ["p(95)<250"],
    "http_req_duration{endpoint:public_schedule}": ["p(95)<250"],
    "http_req_duration{endpoint:rsvp_submit}": ["p(95)<700"],
    "http_req_duration{endpoint:otp_request}": ["p(95)<700"],
    "http_req_duration{endpoint:otp_verify}": ["p(95)<900"],
  },
};

function requestPublicEvent() {
  const res = http.get(`${BASE_URL}/v1/public/events/${EVENT_SLUG}`, {
    tags: { endpoint: "public_event", flow: "read" },
  });
  check(res, { "public event status 200": (r) => r.status === 200 });
}

function requestPublicSchedule() {
  const res = http.get(`${BASE_URL}/v1/public/events/${EVENT_SLUG}/schedule`, {
    tags: { endpoint: "public_schedule", flow: "read" },
  });
  check(res, { "public schedule status 200": (r) => r.status === 200 });
}

function submitRSVP() {
  if (!GUEST_TOKEN) {
    throw new Error("GUEST_TOKEN env var is required for RSVP scenario");
  }

  const payload = JSON.stringify({
    items: [
      {
        sub_event_id: randomItem(SUB_EVENT_IDS),
        status: randomItem(["yes", "no", "maybe"]),
        meal_pref: randomItem(["veg", "non_veg", ""]),
        dietary: "",
        accommodation_needed: Math.random() < 0.2,
        travel_mode: randomItem(["self", "bus", "train", ""]),
        plus_one_names: "",
      },
    ],
  });

  const res = http.post(`${BASE_URL}/v1/public/events/${EVENT_SLUG}/rsvp`, payload, {
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${GUEST_TOKEN}`,
      "Idempotency-Key": `${__VU}-${__ITER}-${Date.now()}`,
    },
    tags: { endpoint: "rsvp_submit", flow: "write" },
  });

  check(res, {
    "rsvp accepted": (r) => [200, 201, 409].includes(r.status),
  });
}

function phoneForVU() {
  const suffix = String(__VU).padStart(5, "0");
  return `${PHONE_PREFIX}${suffix}`;
}

function runOTPFlow() {
  const phone = phoneForVU();

  const reqRes = http.post(
    `${BASE_URL}/v1/auth/otp/request`,
    JSON.stringify({ phone }),
    {
      headers: { "Content-Type": "application/json" },
      tags: { endpoint: "otp_request", flow: "auth" },
    }
  );
  check(reqRes, {
    "otp request response expected": (r) => [200, 202, 429].includes(r.status),
  });

  if (reqRes.status === 200 || reqRes.status === 202) {
    const verifyRes = http.post(
      `${BASE_URL}/v1/auth/otp/verify`,
      JSON.stringify({ phone, code: OTP_CODE }),
      {
        headers: { "Content-Type": "application/json" },
        tags: { endpoint: "otp_verify", flow: "auth" },
      }
    );
    check(verifyRes, {
      "otp verify response expected": (r) => [200, 401, 429].includes(r.status),
    });
  }
}

export function publicReadScenario() {
  requestPublicEvent();
  requestPublicSchedule();
  sleep(0.2);
}

export function rsvpScenario() {
  submitRSVP();
  sleep(0.3);
}

export function otpScenario() {
  runOTPFlow();
  sleep(0.5);
}
