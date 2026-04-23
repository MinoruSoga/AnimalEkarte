/**
 * K6 Load Testing - API Endpoints
 *
 * 主要 API エンドポイントの負荷テスト
 * 並行ユーザー数、スループット、レスポンスタイムを測定
 *
 * 実行: k6 run load-tests/k6-api-endpoints.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// カスタムメトリクス
const errorRate = new Rate('errors');
const loginDuration = new Trend('login_duration');
const appointmentListDuration = new Trend('appointment_list_duration');
const medicalRecordDuration = new Trend('medical_record_duration');
const permissionGroupDuration = new Trend('permission_group_duration');
const successfulLogins = new Counter('successful_logins');
const apiErrors = new Counter('api_errors');

// テストオプション
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp-up: 10ユーザーまで 30秒で増加
    { duration: '1m30s', target: 10 }, // Stay: 10ユーザーで 1分30秒保持
    { duration: '30s', target: 50 },   // Spike: 50ユーザーに急増
    { duration: '1m', target: 50 },    // Stay: 50ユーザーで 1分保持
    { duration: '30s', target: 0 },    // Ramp-down: 0ユーザーまで減少
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],      // 95%レスポンス < 500ms, 99% < 1000ms
    'http_req_failed': ['rate<0.1'],                        // エラー率 < 10%
    'errors': ['rate<0.1'],                                 // カスタムエラー率 < 10%
  },
};

// テスト用認証情報（環境変数から取得）
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TEST_EMAIL = __ENV.TEST_EMAIL || 'admin@example.com';
const TEST_CRED = __ENV.TEST_CRED || 'test';

let authToken = '';

export function setup() {
  // グローバル認証トークン取得
  const loginRes = http.post(`${BASE_URL}/api/v1/login`, {
    email: TEST_EMAIL,
    cred: TEST_CRED,
  });

  // レスポンスから認証情報を抽出（通常 Cookie に設定される）
  if (loginRes.status !== 200) {
    console.error('Setup: Login failed', loginRes.status, loginRes.body);
    throw new Error('Failed to login during setup');
  }

  return { token: loginRes.headers['Set-Cookie'] || '' };
}

export default function (data) {
  // 認証ヘッダーを準備
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Cookie': data.token,
    },
  };

  // ========== ログインテスト ==========
  group('Login Flow', () => {
    const loginRes = http.post(`${BASE_URL}/api/v1/login`, {
      email: TEST_EMAIL,
      cred: TEST_CRED,
    });

    loginDuration.add(loginRes.timings.duration);
    const loginOk = check(loginRes, {
      'login status is 200': (r) => r.status === 200,
      'login has Set-Cookie': (r) => r.headers['Set-Cookie'] !== undefined,
    });

    if (loginOk) {
      successfulLogins.add(1);
    } else {
      errorRate.add(1);
      apiErrors.add(1);
    }
  });

  sleep(1);

  // ========== 診察一覧取得 ==========
  group('Appointment List', () => {
    const appointmentRes = http.get(`${BASE_URL}/api/v1/appointments`, params);

    appointmentListDuration.add(appointmentRes.timings.duration);
    const appointmentOk = check(appointmentRes, {
      'appointment list status 200 or 401': (r) => r.status === 200 || r.status === 401,
      'appointment response time < 1000ms': (r) => r.timings.duration < 1000,
    });

    if (!appointmentOk) {
      errorRate.add(1);
      apiErrors.add(1);
    }
  });

  sleep(1);

  // ========== 医療記録取得 ==========
  group('Medical Records', () => {
    const medicalRes = http.get(`${BASE_URL}/api/v1/medical-records`, params);

    medicalRecordDuration.add(medicalRes.timings.duration);
    const medicalOk = check(medicalRes, {
      'medical records status 200 or 401': (r) => r.status === 200 || r.status === 401,
      'medical records response time < 1500ms': (r) => r.timings.duration < 1500,
    });

    if (!medicalOk) {
      errorRate.add(1);
      apiErrors.add(1);
    }
  });

  sleep(1);

  // ========== 権限グループ取得 ==========
  group('Permission Groups', () => {
    const permRes = http.get(`${BASE_URL}/api/v1/permission-groups`, params);

    permissionGroupDuration.add(permRes.timings.duration);
    const permOk = check(permRes, {
      'permission groups status 200 or 401': (r) => r.status === 200 || r.status === 401,
      'permission groups response time < 1000ms': (r) => r.timings.duration < 1000,
    });

    if (!permOk) {
      errorRate.add(1);
      apiErrors.add(1);
    }
  });

  sleep(1);
}

export function handleSummary(data) {
  // テスト結果をコンソールとファイルに出力
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'load-tests/results.json': JSON.stringify(data),
  };
}

/**
 * テキストサマリー生成 (簡易版)
 */
function textSummary(data, options) {
  const indent = options.indent || '';
  let summary = '\n=== Load Test Results ===\n';

  if (data.metrics) {
    summary += `\n${indent}HTTP Requests:\n`;
    if (data.metrics.http_requests) {
      summary += `${indent}  Total: ${data.metrics.http_requests.value}\n`;
    }
    if (data.metrics.http_req_failed) {
      summary += `${indent}  Failed: ${data.metrics.http_req_failed.value}\n`;
    }
    if (data.metrics.http_req_duration) {
      summary += `${indent}  Avg Duration: ${Math.round(data.metrics.http_req_duration.value)}ms\n`;
    }
  }

  return summary;
}
