# Load Testing Guide

Animal Ekarte の負荷テストは k6 を使用しています。API エンドポイントのパフォーマンス、スケーラビリティ、スパイク対応能力を測定します。

## テストシナリオ

| テスト | ファイル | 目的 | ユーザー数 |
|--------|---------|------|----------|
| API エンドポイント | `k6-api-endpoints.js` | 通常運用負荷 | 10 → 50 |
| スパイテスト | `k6-spike-test.js` | 急激な負荷増加対応 | 5 → 100 → 5 |

## 前提条件

k6 がインストール済み（Docker 推奨）

```bash
# Docker で k6 実行
docker run -i grafana/k6 run - < load-tests/k6-api-endpoints.js
```

## テスト実行

### API エンドポイント負荷テスト

```bash
# 標準実行（ローカル k6）
k6 run load-tests/k6-api-endpoints.js

# 環境指定
BASE_URL=http://localhost:8080 \
TEST_EMAIL=admin@example.com \
TEST_CRED=test \
k6 run load-tests/k6-api-endpoints.js

# Docker 経由
docker run -v $(pwd):/scripts -e BASE_URL=http://host.docker.internal:8080 \
  grafana/k6 run /scripts/load-tests/k6-api-endpoints.js
```

### スパイクテスト

```bash
k6 run load-tests/k6-spike-test.js

# リモート結果送信（Grafana Cloud）
k6 run --out cloud load-tests/k6-spike-test.js
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
- POST `/api/v1/login` — ログイン
- GET `/api/v1/appointments` — 診察一覧
- GET `/api/v1/medical-records` — 医療記録
- GET `/api/v1/permission-groups` — 権限グループ

**パフォーマンス閾値**:
- レスポンスタイム: p95 < 500ms, p99 < 1000ms
- エラー率: < 10%

### k6-spike-test.js

**段階別負荷**:
1. **Normal** (0-10s): 5ユーザー
2. **Spike** (10-15s): 100ユーザーに急増
3. **Sustained** (15-25s): 100ユーザーで保持
4. **Recovery** (25-30s): 5ユーザーに復帰

**パフォーマンス閾値**:
- レスポンスタイム: p95 < 2000ms, p99 < 5000ms
- エラー率: < 20% (スパイク時は許容)

## 結果分析

### 生成ファイル

```
load-tests/results.json — JSON形式の詳細結果
```

### メトリクス解釈

| メトリクス | 説明 | 目標値 |
|----------|------|-------|
| `http_req_duration` | レスポンスタイム | p95 < 500ms |
| `http_req_failed` | リクエスト失敗率 | < 10% |
| `errors` | カスタムエラー率 | < 10% |
| `successful_logins` | 成功ログイン数 | 最大化 |
| `api_errors` | API エラー数 | 最小化 |

### 例: テスト結果の読み方

```
=== Load Test Results ===

HTTP Requests:
  Total: 500
  Failed: 5
  Avg Duration: 450ms

✓ http_req_duration: p95=480ms, p99=950ms
✓ http_req_failed: rate=1%
✓ errors: rate=1%
```

**解釈**:
- ✅ パフォーマンス目標達成（p95 < 500ms）
- ✅ エラー率低い（1%）
- ✅ 500リクエスト中 5失敗（許容範囲）

## トラブルシューティング

### エラー: "Connection refused"
- バックエンドが起動していることを確認
- BASE_URL が正しいことを確認

```bash
docker compose ps   # Docker Compose 確認
curl http://localhost:8080/health   # API ヘルスチェック
```

### エラー: "Too many open files"
- k6 が開けるファイル数制限に達した
- システムリソース制限を増やす

```bash
# macOS / Linux
ulimit -n 4096

# Docker の場合
docker run --ulimit nofile=65536:65536 grafana/k6 run ...
```

### スパイクテスト失敗
- システムが高負荷に対応できていない可能性
- エンドポイント最適化を検討
- キャッシング戦略を実装

## パフォーマンス最適化ガイド

### 1. データベースクエリ最適化

```go
// 例: N+1 問題を避ける
appointments, _ := repo.FindAllWithPreload(ctx)  // ✅ プリロード
appointments, _ := repo.FindAll(ctx)             // ❌ N+1 クエリ
```

### 2. API レスポンス キャッシング

```go
// Redis キャッシング例
if cached, ok := cache.Get("appointments"); ok {
    return cached, nil
}
appointments, _ := repo.FindAll(ctx)
cache.Set("appointments", appointments, 5*time.Minute)
return appointments, nil
```

### 3. コネクションプール調整

```go
// GORM コネクションプール設定
db.DB().SetMaxIdleConns(10)
db.DB().SetMaxOpenConns(100)
```

### 4. インデックス追加

```sql
-- 頻繁なクエリ対象にインデックスを追加
CREATE INDEX idx_clinic_id ON appointments(clinic_id);
CREATE INDEX idx_staff_id ON staff_clinic_assignments(staff_id);
```

## CI/CD 統合

### GitHub Actions での負荷テスト実行

`.github/workflows/load-test.yml`:

```yaml
name: Load Tests
on:
  schedule:
    - cron: '0 2 * * *'  # 毎日 02:00 実行
  workflow_dispatch

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: grafana/setup-k6-action@v1
      - run: docker compose up -d
      - run: k6 run load-tests/k6-api-endpoints.js
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: load-test-results
          path: load-tests/results.json
```

## 参考資料

- [k6 公式ドキュメント](https://k6.io/docs/)
- [k6 API リファレンス](https://k6.io/docs/javascript-api/)
- [Grafana Cloud k6](https://grafana.com/products/cloud/k6/)

## パフォーマンスベンチマーク目標

| 指標 | 目標 | 現状 | 状態 |
|------|------|------|------|
| ログイン レスポンス | < 300ms | TBD | 🔄 |
| 診察一覧 取得 | < 500ms | TBD | 🔄 |
| 医療記録 取得 | < 1000ms | TBD | 🔄 |
| 権限グループ 取得 | < 500ms | TBD | 🔄 |
| **全体エラー率** | < 5% | TBD | 🔄 |
| **同時接続数** | 100+ | TBD | 🔄 |

---

**最終更新**: 2026-04-23  
**担当**: Claude Code (TIER 3 Load Testing)
