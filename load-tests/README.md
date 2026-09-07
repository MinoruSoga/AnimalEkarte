# Load Testing Guide

Animal Ekarte の負荷テストは k6 を使用しています。API エンドポイントのパフォーマンス、スケーラビリティ、スパイク対応能力を測定します。

## テストシナリオ

| テスト             | ファイル                 | 目的                     | ユーザー数  |
| ------------------ | ------------------------ | ------------------------ | ----------- |
| API エンドポイント | `k6-api-endpoints.js`    | 通常運用負荷             | 10 → 50     |
| スパイテスト       | `k6-spike-test.js`       | 急激な負荷増加対応       | 5 → 100 → 5 |
| CF STG 持続負荷    | `k6-cf-stg-sustained.js` | STG Container 低 VU 観測 | 3           |

## 前提条件

- k6 がインストール済み（Docker 推奨）
- ローカル API エンドポイント／スパイクテストの認証は **`LOAD_TEST_LOGIN_EMAIL` / `LOAD_TEST_LOGIN_PASSWORD` のみ**（literal fallback なし）
- GitHub Actions の1つの `load-test` job 内の2つの k6 step では同名 env が必須（未設定は fail-closed）

```bash
# Docker で k6 実行（認証 env 必須）
docker run --rm -i \
  -e BASE_URL=http://host.docker.internal:8080 \
  -e LOAD_TEST_LOGIN_EMAIL \
  -e LOAD_TEST_LOGIN_PASSWORD \
  -v "$(pwd)/load-tests:/scripts" \
  grafana/k6 run /scripts/k6-api-endpoints.js
```

## 認証契約（#109 / TASK-606）

| 項目      | 契約                                                                         |
| --------- | ---------------------------------------------------------------------------- |
| 環境変数  | `LOAD_TEST_LOGIN_EMAIL`, `LOAD_TEST_LOGIN_PASSWORD` のみ                     |
| 禁止      | 旧 CI テスト用 secret 名・汎用 TEST_* env・ハードコード demo 認証の fallback |
| login     | `POST /api/v1/login` が非 200、または `Set-Cookie` 欠落 → 非 0 終了          |
| protected | 認証付き GET が非 200 → check 失敗（401 合格扱いは禁止）                     |
| aggregate | `http_reqs` / `iterations` / `checks` / `successful_logins` が 0 → 非 0 終了 |
| 秘匿      | パスワード・body・cookie・token 値を log / summary / artifact に出さない     |

ローカル API エンドポイント／スパイクテストは `setup()` で 1 回だけログインし、Cookie を再利用する（login レートリミット回避）。

## テスト実行

### API エンドポイント負荷テスト

```bash
BASE_URL=http://localhost:8080 \
LOAD_TEST_LOGIN_EMAIL="$LOAD_TEST_LOGIN_EMAIL" \
LOAD_TEST_LOGIN_PASSWORD="$LOAD_TEST_LOGIN_PASSWORD" \
k6 run load-tests/k6-api-endpoints.js \
  --summary-export load-tests/results-endpoints-summary.json
```

### スパイクテスト

```bash
BASE_URL=http://localhost:8080 \
LOAD_TEST_LOGIN_EMAIL="$LOAD_TEST_LOGIN_EMAIL" \
LOAD_TEST_LOGIN_PASSWORD="$LOAD_TEST_LOGIN_PASSWORD" \
k6 run load-tests/k6-spike-test.js \
  --summary-export load-tests/results-spike-summary.json
```

## テストパラメータ

### k6-api-endpoints.js

**段階別負荷**:

1. **Ramp-up** (0-30s): 10ユーザーに増加
2. **Soak** (30-90s): 10ユーザーで保持
3. **Spike** (90-120s): 50ユーザーに急増
4. **Sustained** (120-180s): 50ユーザーで保持
5. **Ramp-down** (180-210s): 0ユーザーに減少

**テスト対象 API**:

- POST `/api/v1/login` — setup で 1 回（`successful_logins` 計上）
- GET `/api/v1/reservations` — 診察一覧（status 200 必須）
- GET `/api/v1/medical-records` — 医療記録（status 200 必須）
- GET `/api/v1/masters/animal-species` — 動物種類（status 200 必須）

**パフォーマンス閾値**:

- レスポンスタイム: p95 < 500ms, p99 < 1000ms
- エラー率: < 10%
- fail-closed: `http_reqs` / `iterations` / `checks` / `successful_logins` > 0

### k6-spike-test.js

**段階別負荷**:

1. **Normal** (0-10s): 5ユーザー
2. **Spike** (10-15s): 100ユーザーに急増
3. **Sustained** (15-25s): 100ユーザーで保持
4. **Recovery** (25-30s): 5ユーザーに復帰

**認証**: setup で `LOAD_TEST_LOGIN_*` ログイン → Cookie 付き GET `/api/v1/reservations`（status 200 必須）

**パフォーマンス閾値**:

- レスポンスタイム: p95 < 2000ms, p99 < 5000ms
- エラー率: < 20% (スパイク時は許容)
- fail-closed: `http_reqs` / `iterations` / `checks` / `successful_logins` > 0

## 結果分析

### 生成ファイル

```
load-tests/results-endpoints-summary.json — k6 --summary-export（API endpoints）
load-tests/results-spike-summary.json     — k6 --summary-export（spike）
```

CI は time-series の `--out json` も artifact に残すが、合格判定は **summary-export の aggregate** のみ（`http_reqs` 等）。`metrics.http_requests` のような誤キーや silent parse は使わない。

### メトリクス解釈

| メトリクス          | 説明                 | 目標値                   |
| ------------------- | -------------------- | ------------------------ |
| `http_req_duration` | レスポンスタイム     | p95 < 500ms（endpoints） |
| `http_req_failed`   | リクエスト失敗率     | < 10%                    |
| `http_reqs`         | 総リクエスト数       | > 0（必須）              |
| `iterations`        | VU イテレーション数  | > 0（必須）              |
| `checks`            | check 実行           | rate > 0 かつ活動量 > 0  |
| `successful_logins` | setup ログイン成功数 | > 0（必須）              |
| `errors`            | カスタムエラー率     | < 10%                    |

## トラブルシューティング

### エラー: "LOAD_TEST_LOGIN_EMAIL / LOAD_TEST_LOGIN_PASSWORD must be set"

- ローカル環境変数または CI job の env が未設定
- literal fallback は廃止済み。値を export して再実行

### エラー: "setup login failed: status=..."

- 認証情報が対象環境のアカウントと一致しているか確認
- status 以外（body/cookie 値）はログに出ない仕様

### エラー: "Connection refused"

- バックエンドが起動していることを確認
- BASE_URL が正しいことを確認

```bash
docker compose ps
curl http://localhost:8080/health
```

### エラー: "Too many open files"

```bash
ulimit -n 4096
docker run --ulimit nofile=65536:65536 grafana/k6 run ...
```

## CI/CD 統合

`.github/workflows/performance-tests.yml`:

- 両 k6 step は **required**（`continue-on-error` なし）
- env: `LOAD_TEST_LOGIN_EMAIL` / `LOAD_TEST_LOGIN_PASSWORD`（fallback なし）
- `--summary-export load-tests/results-endpoints-summary.json` / `results-spike-summary.json`
- 後段 step が aggregate（`http_reqs`, `iterations`, `checks`, `successful_logins`）を fail-closed 検証
- Lighthouse のみ既存どおり `continue-on-error: true` 可

## 承認済み STG 持続テスト（別契約）

`k6-cf-stg-sustained.js` はローカル API エンドポイント／スパイクテストおよび CI job とは別の、承認済み STG 持続負荷テストである。このスクリプトだけが `STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` を要求する。ローカル手順へ再利用せず、実行には個別の承認を要する。

## 参考資料

- [k6 公式ドキュメント](https://k6.io/docs/)
- [k6 API リファレンス](https://k6.io/docs/javascript-api/)
- `load-tests/k6-cf-stg-sustained.js` — 承認済み STG 持続負荷の別契約

---

**最終更新**: 2026-07-30
**担当**: TASK-606 (#109 performance-tests auth / fail-closed)
