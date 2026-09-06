/**
 * K6 Load Testing - API Endpoints
 *
 * 主要 API エンドポイントの負荷テスト。
 * setup() で1回だけ LOAD_TEST_LOGIN_* 認証し、Cookie を全 VU で再利用する。
 * login 非200 / cookie 欠落 / protected 非200 / 0 request|check|iteration|successful_logins は非0終了。
 * パスワード・body・cookie・token 値はログに出さない。
 *
 * 実行:
 *   BASE_URL=http://localhost:8080 \
 *   LOAD_TEST_LOGIN_EMAIL=... LOAD_TEST_LOGIN_PASSWORD=... \
 *   k6 run load-tests/k6-api-endpoints.js
 */

import http from "k6/http";
import { check, sleep, group } from "k6";
import { Rate, Trend, Counter } from "k6/metrics";

const errorRate = new Rate("errors");
const loginDuration = new Trend("login_duration");
const appointmentListDuration = new Trend("appointment_list_duration");
const medicalRecordDuration = new Trend("medical_record_duration");
const permissionGroupDuration = new Trend("permission_group_duration");
const successfulLogins = new Counter("successful_logins");
const apiErrors = new Counter("api_errors");

export const options = {
  stages: [
    { duration: "30s", target: 10 },
    { duration: "1m30s", target: 10 },
    { duration: "30s", target: 50 },
    { duration: "1m", target: 50 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<500", "p(99)<1000"],
    http_req_failed: ["rate<0.1"],
    errors: ["rate<0.1"],
    // fail-closed: 認証済み負荷が実際に走ったことを aggregate で証明する
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

function authHeaders(cookie) {
  return {
    headers: {
      "Content-Type": "application/json",
      "X-Requested-With": "XMLHttpRequest",
      Cookie: cookie,
    },
  };
}

// setup() は全 VU 共有で1回のみ。login レートリミット回避のためここで認証する。
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

  loginDuration.add(loginRes.timings.duration);

  if (loginRes.status !== 200) {
    // status のみ。body / cookie / token は出さない
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

  const params = authHeaders(data.cookie);

  group("Appointment List", () => {
    const appointmentRes = http.get(`${BASE_URL}/api/v1/reservations`, params);
    appointmentListDuration.add(appointmentRes.timings.duration);
    const appointmentOk = check(appointmentRes, {
      "appointment list status 200": (r) => r.status === 200,
      "appointment response time < 1000ms": (r) => r.timings.duration < 1000,
    });
    if (!appointmentOk) {
      errorRate.add(1);
      apiErrors.add(1);
    }
  });

  sleep(1);

  group("Medical Records", () => {
    const medicalRes = http.get(`${BASE_URL}/api/v1/medical-records`, params);
    medicalRecordDuration.add(medicalRes.timings.duration);
    const medicalOk = check(medicalRes, {
      "medical records status 200": (r) => r.status === 200,
      "medical records response time < 1500ms": (r) =>
        r.timings.duration < 1500,
    });
    if (!medicalOk) {
      errorRate.add(1);
      apiErrors.add(1);
    }
  });

  sleep(1);

  group("Permission Groups", () => {
    const permRes = http.get(
      `${BASE_URL}/api/v1/masters/permission-groups`,
      params,
    );
    permissionGroupDuration.add(permRes.timings.duration);
    const permOk = check(permRes, {
      "permission groups status 200": (r) => r.status === 200,
      "permission groups response time < 1000ms": (r) =>
        r.timings.duration < 1000,
    });
    if (!permOk) {
      errorRate.add(1);
      apiErrors.add(1);
    }
  });

  sleep(1);
}

// summary JSON は CI の --summary-export が正本。ここでは stdout のみ。
export function handleSummary(data) {
  return {
    stdout: textSummary(data, { indent: " " }),
  };
}

function metricCount(metric) {
  if (!metric || !metric.values) return 0;
  if (typeof metric.values.count === "number") return metric.values.count;
  // Rate metrics (checks): passes + fails を活動量として数える
  const passes = metric.values.passes;
  const fails = metric.values.fails;
  if (typeof passes === "number" || typeof fails === "number") {
    return (passes || 0) + (fails || 0);
  }
  return 0;
}

function textSummary(data, options) {
  const indent = (options && options.indent) || "";
  const m = (data && data.metrics) || {};
  let summary = "\n=== Load Test Results (API Endpoints) ===\n";

  summary += `\n${indent}HTTP Requests (http_reqs): ${metricCount(m.http_reqs)}\n`;
  summary += `${indent}Iterations: ${metricCount(m.iterations)}\n`;
  summary += `${indent}Checks: ${metricCount(m.checks)}\n`;
  summary += `${indent}Successful logins: ${metricCount(m.successful_logins)}\n`;

  if (m.http_req_failed && m.http_req_failed.values) {
    const rate = m.http_req_failed.values.rate;
    if (typeof rate === "number") {
      summary += `${indent}Failed rate: ${(rate * 100).toFixed(2)}%\n`;
    }
  }
  if (m.http_req_duration && m.http_req_duration.values) {
    const avg = m.http_req_duration.values.avg;
    if (typeof avg === "number") {
      summary += `${indent}Avg Duration: ${Math.round(avg)}ms\n`;
    }
  }

  // fail-closed 表示用（exit は k6 thresholds が担当）
  for (const name of [
    "http_reqs",
    "iterations",
    "checks",
    "successful_logins",
  ]) {
    if (metricCount(m[name]) <= 0) {
      summary += `${indent}FAIL: ${name} is zero or missing\n`;
    }
  }

  return summary;
}
