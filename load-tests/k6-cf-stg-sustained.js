/**
 * K6 Load Testing - Cloudflare STG Sustained Load (P4-9・試行12)
 *
 * Cloudflare Workers + Containers (STG) に対する10分間の低VU持続負荷。
 * setup() で1回だけログインしてセッションを確立し（下記「設計上の注意」参照）、
 * default() ループでは /health（認証不要）と GET /clinics（認証付き・確立済みCookie再利用）を
 * 繰り返して、Container のコールドスタート・スケール挙動下でのレイテンシとエラー率を観測する。
 * CPU課金実測との比較は Cloudflare Dashboard 側で手動確認する（本スクリプトの範囲外）。
 *
 * 設計上の注意（試行12で発見・修正）:
 *   当初は default() ループ内で毎イテレーションログインしていたが、`POST /api/v1/login` には
 *   IPベースのレートリミット(5回/分・バースト5、BUG-130ブルートフォース対策
 *   `backend/internal/handler/handler.go` 参照)が掛かっており、VU3の同一ソースIPから
 *   連続ログインすると数秒でレート制限(429)に達し、ログイン失敗率が急増した
 *   (試行12初回実行: 失敗率55.86%・successful_logins 54/333)。
 *   これは Cloudflare Containers 側の性能問題ではなく、意図した既存のブルートフォース対策が
 *   正しく機能した結果であるため、本番相当の使い方(1ユーザー1ログイン、以降はセッション継続)に
 *   合わせてログインを setup() で1回だけ行う設計に変更した。
 *
 * 実行（ローカルにk6が無い場合はDocker経由）:
 *   BASE_URL=https://animalekarte-stg-api.baritech-soga.workers.dev \
 *   STG_DEMO_EMAIL=... STG_DEMO_PASSWORD=... \
 *   k6 run load-tests/k6-cf-stg-sustained.js
 *
 *   # Docker経由（k6未インストール時。カレントディレクトリを/load-testsにマウントしwork-dir指定）:
 *   docker run --rm --network host \
 *     -e BASE_URL -e STG_DEMO_EMAIL -e STG_DEMO_PASSWORD \
 *     -v "$(pwd)/load-tests:/load-tests" -w /load-tests \
 *     grafana/k6 run k6-cf-stg-sustained.js
 */

import http from "k6/http";
import { check, sleep, group } from "k6";
import { Rate, Trend } from "k6/metrics";

const errorRate = new Rate("errors");
const healthDuration = new Trend("health_duration");
const clinicsDuration = new Trend("clinics_duration");

// Container のコールドスタート（インスタンス0→1起動）を許容するため、
// 通常APIのp95閾値（500ms級）よりゆるめに設定する。
export const options = {
  stages: [
    { duration: "30s", target: 3 }, // ramp-up
    { duration: "9m", target: 3 }, // sustained（低VU・10分弱の主要区間）
    { duration: "30s", target: 0 }, // ramp-down
  ],
  thresholds: {
    http_req_duration: ["p(95)<3000"],
    http_req_failed: ["rate<0.05"],
    errors: ["rate<0.05"],
  },
};

const BASE_URL =
  __ENV.BASE_URL || "https://animalekarte-stg-api.baritech-soga.workers.dev";
const STG_DEMO_EMAIL = __ENV.STG_DEMO_EMAIL;
const STG_DEMO_PASSWORD = __ENV.STG_DEMO_PASSWORD;

// setup() は全VUで共有される1回のみの実行(k6仕様)。ログインもここで1回だけ行い、
// 発行された Cookie を全VU・全イテレーションで再利用する(実際のユーザー挙動に近い)。
export function setup() {
  if (!STG_DEMO_EMAIL || !STG_DEMO_PASSWORD) {
    throw new Error(
      "STG_DEMO_EMAIL / STG_DEMO_PASSWORD must be set via env (not hardcoded)",
    );
  }
  const res = http.post(
    `${BASE_URL}/api/v1/login`,
    JSON.stringify({ email: STG_DEMO_EMAIL, password: STG_DEMO_PASSWORD }),
    {
      headers: {
        "Content-Type": "application/json",
        "X-Requested-With": "XMLHttpRequest",
      },
    },
  );
  if (res.status !== 200 || !res.headers["Set-Cookie"]) {
    throw new Error(`setup login failed: status=${res.status}`);
  }
  return { cookie: res.headers["Set-Cookie"] };
}

export default function (data) {
  group("Health", () => {
    const res = http.get(`${BASE_URL}/health`);
    healthDuration.add(res.timings.duration);
    const healthOk = check(res, {
      "health status 200": (r) => r.status === 200,
    });
    errorRate.add(healthOk ? 0 : 1);
  });

  sleep(1);

  group("Clinics", () => {
    const res = http.get(`${BASE_URL}/api/v1/clinics`, {
      headers: { Cookie: data.cookie, "X-Requested-With": "XMLHttpRequest" },
    });
    clinicsDuration.add(res.timings.duration);
    const clinicsOk = check(res, {
      "clinics status 200": (r) => r.status === 200,
      "clinics response < 3000ms": (r) => r.timings.duration < 3000,
    });
    errorRate.add(clinicsOk ? 0 : 1);
  });

  sleep(2);
}

export function handleSummary(data) {
  return {
    stdout: textSummary(data),
    "results-cf-stg-sustained.json": JSON.stringify(data),
  };
}

function textSummary(data) {
  const m = data.metrics || {};
  let s = "\n=== Cloudflare STG Sustained Load Summary (P4-9) ===\n";
  if (m.http_reqs) s += `Total requests: ${m.http_reqs.values.count}\n`;
  if (m.http_req_failed)
    s += `Failed rate: ${(m.http_req_failed.values.rate * 100).toFixed(2)}%\n`;
  if (m.http_req_duration) {
    s += `p95 duration: ${m.http_req_duration.values["p(95)"]?.toFixed(0)}ms\n`;
    s += `avg duration: ${m.http_req_duration.values.avg?.toFixed(0)}ms\n`;
  }
  return s;
}
