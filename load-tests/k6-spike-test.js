/**
 * K6 Load Testing - Spike Test
 *
 * スパイクテスト：突然の高負荷に対するシステムの反応を測定。
 * setup() で LOAD_TEST_LOGIN_* 認証し、Cookie 付きで protected API を叩く。
 * login 非200 / cookie 欠落 / protected 非200 / 0 request|check|iteration|successful_logins は非0終了。
 * パスワード・body・cookie・token 値はログに出さない。
 *
 * 実行:
 *   BASE_URL=http://localhost:8080 \
 *   LOAD_TEST_LOGIN_EMAIL=... LOAD_TEST_LOGIN_PASSWORD=... \
 *   k6 run load-tests/k6-spike-test.js
 */

import http from "k6/http";
import { check, sleep, group } from "k6";
import { Trend, Rate, Counter } from "k6/metrics";

const responseDuration = new Trend("response_duration");
const errorRate = new Rate("errors");
const failedRequests = new Counter("failed_requests");
const successfulLogins = new Counter("successful_logins");

export const options = {
  stages: [
    { duration: "10s", target: 5 },
    { duration: "5s", target: 100 },
    { duration: "10s", target: 100 },
    { duration: "5s", target: 5 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<2000", "p(99)<5000"],
    errors: ["rate<0.2"],
    // fail-closed: 認証済みスパイクが実際に走ったことを aggregate で証明する
    http_reqs: ["count>0"],
    iterations: ["count>0"],
    checks: ["rate>0"],
    successful_logins: ["count>0"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const LOAD_TEST_LOGIN_EMAIL = __ENV.LOAD_TEST_LOGIN_EMAIL;
const LOAD_TEST_LOGIN_PASSWORD = __ENV.LOAD_TEST_LOGIN_PASSWORD;

function requireCredentials() {
  if (!LOAD_TEST_LOGIN_EMAIL || !LOAD_TEST_LOGIN_PASSWORD) {
    throw new Error(
      "LOAD_TEST_LOGIN_EMAIL / LOAD_TEST_LOGIN_PASSWORD must be set via env (no fallback)",
    );
  }
}

export function setup() {
  requireCredentials();

  const loginRes = http.post(
    `${BASE_URL}/api/v1/login`,
    JSON.stringify({
      email: LOAD_TEST_LOGIN_EMAIL,
      password: LOAD_TEST_LOGIN_PASSWORD,
    }),
    {
      headers: {
        "Content-Type": "application/json",
        "X-Requested-With": "XMLHttpRequest",
      },
    },
  );

  if (loginRes.status !== 200) {
    throw new Error(`setup login failed: status=${loginRes.status}`);
  }

  const cookie = loginRes.headers["Set-Cookie"];
  if (!cookie) {
    throw new Error("setup login failed: missing Set-Cookie");
  }

  successfulLogins.add(1);
  return { cookie };
}

export default function (data) {
  if (!data || !data.cookie) {
    throw new Error("missing session cookie from setup");
  }

  const params = {
    headers: {
      "Content-Type": "application/json",
      "X-Requested-With": "XMLHttpRequest",
      Cookie: data.cookie,
    },
  };

  group("Spike Test - Appointment List", () => {
    const res = http.get(`${BASE_URL}/api/v1/reservations`, params);
    responseDuration.add(res.timings.duration);

    const ok = check(res, {
      "status is 200": (r) => r.status === 200,
      "response time < 3000ms": (r) => r.timings.duration < 3000,
    });

    if (!ok) {
      errorRate.add(1);
      failedRequests.add(1);
    }
  });

  sleep(2);
}

// summary JSON は CI の --summary-export が正本。ここでは stdout のみ。
export function handleSummary(data) {
  return {
    stdout: textSummary(data),
  };
}

function metricCount(metric) {
  if (!metric || !metric.values) return 0;
  if (typeof metric.values.count === "number") return metric.values.count;
  const passes = metric.values.passes;
  const fails = metric.values.fails;
  if (typeof passes === "number" || typeof fails === "number") {
    return (passes || 0) + (fails || 0);
  }
  return 0;
}

function textSummary(data) {
  const m = (data && data.metrics) || {};
  let s = "\n=== Spike Test Results ===\n";
  s += `HTTP Requests (http_reqs): ${metricCount(m.http_reqs)}\n`;
  s += `Iterations: ${metricCount(m.iterations)}\n`;
  s += `Checks: ${metricCount(m.checks)}\n`;
  s += `Successful logins: ${metricCount(m.successful_logins)}\n`;
  if (
    m.http_req_failed &&
    m.http_req_failed.values &&
    typeof m.http_req_failed.values.rate === "number"
  ) {
    s += `Failed rate: ${(m.http_req_failed.values.rate * 100).toFixed(2)}%\n`;
  }
  if (
    m.http_req_duration &&
    m.http_req_duration.values &&
    typeof m.http_req_duration.values.avg === "number"
  ) {
    s += `Avg duration: ${Math.round(m.http_req_duration.values.avg)}ms\n`;
  }
  for (const name of [
    "http_reqs",
    "iterations",
    "checks",
    "successful_logins",
  ]) {
    if (metricCount(m[name]) <= 0) {
      s += `FAIL: ${name} is zero or missing\n`;
    }
  }
  return s;
}
