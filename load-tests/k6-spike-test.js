/**
 * K6 Load Testing - Spike Test
 *
 * スパイクテスト：突然の高負荷に対するシステムの反応を測定
 * 通常運用 → 急激な負荷増加 → 復旧パターン
 *
 * 実行: k6 run load-tests/k6-spike-test.js
 */

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend, Rate, Counter } from 'k6/metrics';

const responseDuration = new Trend('response_duration');
const errorRate = new Rate('errors');
const failedRequests = new Counter('failed_requests');

export const options = {
  stages: [
    { duration: '10s', target: 5 },    // 正常運用: 5ユーザー
    { duration: '5s', target: 100 },   // スパイク: 5秒で100ユーザーに急増
    { duration: '10s', target: 100 },  // 保持: 100ユーザーで 10秒
    { duration: '5s', target: 5 },     // 復旧: 5ユーザーまで低下
  ],
  thresholds: {
    'http_req_duration': ['p(95)<2000', 'p(99)<5000'],
    'errors': ['rate<0.2'],  // スパイク時は許容度を上げる
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const TEST_EMAIL = __ENV.TEST_EMAIL || 'admin@example.com';
const TEST_CRED = __ENV.TEST_CRED || 'test';

export default function () {
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  group('Spike Test - Appointment List', () => {
    const res = http.get(`${BASE_URL}/api/v1/appointments`, params);

    responseDuration.add(res.timings.duration);

    const ok = check(res, {
      'status is 200 or 401': (r) => r.status === 200 || r.status === 401,
      'response time < 3000ms': (r) => r.timings.duration < 3000,
    });

    if (!ok) {
      errorRate.add(1);
      failedRequests.add(1);
    }
  });

  sleep(2);
}
