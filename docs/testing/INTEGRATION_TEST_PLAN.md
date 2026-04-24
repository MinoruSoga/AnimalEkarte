# 統合テスト計画書

> **バージョン**: 1.0  
> **最終更新**: 2026-04-23  
> **ステータス**: ✅ TIER 4 — 統合テスト・品質保証フェーズ  
> **対象**: TIER 1-3 すべての実装の動作検証

---

## 📋 概要

TIER 1 (CRUD 基盤), TIER 2 (インフラ・DB マイグレーション), TIER 3 (テスト・パフォーマンス) の実装が完了しています。本フェーズでは、すべてのコンポーネントが正常に統合され、実際のシステムとして動作することを確認します。

### テストの三層構造

```
┌────────────────────────────────────────┐
│ E2E テスト (Playwright)                  │ TIER 3
│ 18 テストケース + 6 スイート             │
└────────────────────────────────────────┘
              ↓ (API 呼び出し)
┌────────────────────────────────────────┐
│ API 統合テスト                           │ TIER 4 ← 本フェーズ
│ ハンドラー × サービス × リポジトリ連携   │
│ データベース × トランザクション管理       │
└────────────────────────────────────────┘
              ↓ (SQL 実行)
┌────────────────────────────────────────┐
│ DB マイグレーション + 整合性            │ TIER 2
│ 001-004 マイグレーション完了            │
└────────────────────────────────────────┘
```

---

## 🎯 テスト目標

### 対象範囲

| コンポーネント | テスト項目 | 検証対象 |
|-------------|---------|---------|
| **バックエンド** | API + DB トランザクション | 40+ エンドポイント × CRUD × トランザクション |
| **フロントエンド** | フォーム送信 + 状態管理 | React 19 Action + useActionState |
| **インフラ** | マイグレーション + シード | DB 初期化 + 整合性確認 |
| **パフォーマンス** | 負荷テスト + プロファイリング | p95<500ms, メモリ<500MB |

### 合格基準

| テスト種別 | 基準 | 対象 |
|-----------|------|------|
| **API ユニットテスト** | PASS >= 95% | 各サービス層 |
| **統合テスト** | PASS = 100% | API + DB |
| **E2E テスト** | PASS >= 90% | 主要フロー 6 シナリオ |
| **パフォーマンス** | p95 < 500ms | 負荷テスト中 |
| **セキュリティ** | PASS = 100% | RBAC + JWT |

---

## 📊 テスト計画マトリックス

### フェーズ 1: ユニットテスト検証

| レイヤー | テスト対象 | テスト方法 | 検証項目 | 目標 |
|---------|---------|---------|--------|------|
| **Repository** | GORM モデル | `go test ./internal/repository` | CRUD 操作の正確性 | PASS 100% |
| **Service** | ビジネスロジック | `go test ./internal/service` | 権限チェック・バリデーション | PASS 100% |
| **Handler** | HTTP ハンドラー | `go test ./internal/handler` | ルーティング・レスポンス形式 | PASS 100% |
| **Frontend** | React コンポーネント | `pnpm test:run` | useActionState + フォーム | PASS 100% |

**実行コマンド**:
```bash
# バックエンド全テスト
docker compose exec backend go test ./...

# フロントエンド全テスト
docker compose exec frontend pnpm test:run
```

---

### フェーズ 2: API 統合テスト

#### 飼主管理フロー

| シナリオ | テスト項目 | 手順 | 期待結果 |
|--------|---------|------|--------|
| **飼主作成** | POST /api/v1/owners | 有効データで飼主作成 | 201 Created, id 返却 |
| **飼主取得** | GET /api/v1/owners/:id | 作成した飼主を取得 | 200 OK, 完全なデータ |
| **飼主更新** | PATCH /api/v1/owners/:id | 割引率を変更 | 200 OK, 更新反映 |
| **ペット追加** | POST /api/v1/owners/:id/pets | 飼主配下にペット追加 | 201 Created |
| **飼主削除** | DELETE /api/v1/owners/:id | 関連ペットあり時 | 409 Conflict (FK 制約) |
| **飼主削除 (成功)** | DELETE /api/v1/owners/:id | ペット削除後に飼主削除 | 204 No Content |

#### 予約管理フロー

| シナリオ | テスト項目 | 手順 | 期待結果 |
|--------|---------|------|--------|
| **予約作成** | POST /api/v1/appointments | 有効な日時・スタッフで予約 | 201 Created |
| **予約時間重複チェック** | POST /api/v1/appointments | 同スタッフ・時間で重複作成試行 | 409 Conflict |
| **営業日チェック** | POST /api/v1/appointments | 定休日に予約試行 | 400 Bad Request |
| **スタッフ出勤チェック** | POST /api/v1/appointments | 出勤していないスタッフで予約 | 400 Bad Request |
| **予約一覧取得** | GET /api/v1/appointments?date=YYYY-MM-DD | 特定日の予約一覧 | 200 OK, 3件以上 |

#### 医療記録フロー

| シナリオ | テスト項目 | 手順 | 期待結果 |
|--------|---------|------|--------|
| **カルテ作成 (draft)** | POST /api/v1/medical-records | ペット選択 → draft 自動生成 | 201 Created, status=draft |
| **主訴保存** | PATCH /api/v1/medical-records/:id | chief_complaint 入力 | 200 OK |
| **治療項目追加** | PATCH /api/v1/medical-records/:id | treatment_items 配列追加 | 200 OK, 合計金額更新 |
| **医師確認** | PATCH /api/v1/medical-records/:id | status → finalized | 200 OK |
| **確定後編集不可** | PATCH /api/v1/medical-records/:id | finalized カルテを編集試行 | 400 Bad Request |

#### RBAC 権限テスト

| ロール | テスト項目 | テスト手順 | 期待結果 |
|--------|---------|---------|--------|
| **獣医師** | 全 CRUD 操作 | 飼主・カルテ・会計・削除 | すべて 201/200/204 |
| **助手** | 読み取り + 一部作成 | 飼主作成可、カルテ削除不可 | 201 OK, 403 Forbidden |
| **受付** | 読み取り + 予約作成 | 予約作成可、カルテ編集不可 | 201 OK, 403 Forbidden |

---

### フェーズ 3: E2E テスト実行

#### テストスイート実行

```bash
# Headless 実行（CI/CD 向け）
pnpm test:e2e

# UI モード（デバッグ用）
pnpm test:e2e:ui

# 単一スイート実行
pnpm test:e2e -- frontend/tests/appointment.spec.ts
```

#### テストケース検証一覧

| テストスイート | ケース数 | テスト項目 | 期待結果 |
|-------------|--------|---------|--------|
| **appointment.spec.ts** | 3 | 予約フロー (ナビゲーション・フォーム・送信) | ✅ PASS |
| **medical-records.spec.ts** | 3 | カルテフロー (フォーム・保存・状態) | ✅ PASS |
| **hospitalization.spec.ts** | 4 | 入院フロー (リスト・フォーム・日付) | ✅ PASS |
| **permission-control.spec.ts** | 5 | 権限管理 (グループ・パネル・チェック) | ✅ PASS |
| **staff-management.spec.ts** | 5 | スタッフ登録 (リスト・フォーム・割当) | ✅ PASS |
| **auth.spec.ts** | 2 | 認証フロー (ログイン・セッション) | ✅ PASS |
| **合計** | **22** | — | **PASS >= 90%** |

---

### フェーズ 4: 負荷テスト実行

#### k6 API エンドポイントテスト

```bash
k6 run load-tests/k6-api-endpoints.js
```

| ステージ | ユーザー数 | 期間 | 期待結果 |
|---------|---------|------|--------|
| Ramp-up | 0→10 | 30s | レスポンス安定化 |
| Soak | 10 | 90s | p95 < 500ms |
| Spike | 10→50 | 30s | p95 < 500ms, error < 10% |
| Sustained | 50 | 60s | システム安定 |
| Ramp-down | 50→0 | 30s | 正常復帰 |

**期待値**:
- ✅ Total requests > 1000
- ✅ p95 < 500ms, p99 < 1000ms
- ✅ Error rate < 10%
- ✅ ログイン成功率 > 90%

#### k6 スパイクテスト

```bash
k6 run load-tests/k6-spike-test.js
```

| ステージ | ユーザー数 | 期間 | 期待結果 |
|---------|---------|------|--------|
| Normal | 5 | 10s | ベースライン |
| Spike | 5→100 | 5s | 高負荷対応能力測定 |
| Sustained | 100 | 10s | p95 < 2000ms, error < 20% |
| Recovery | 100→5 | 5s | 正常復帰 |

---

### フェーズ 5: パフォーマンスプロファイリング

#### Go バックエンドプロファイリング

```bash
# CPU プロファイル（30秒）
docker compose exec backend go run scripts/profile.go cpu 30s

# メモリプロファイル
docker compose exec backend go run scripts/profile.go memory

# ゴルーチン数
docker compose exec backend go run scripts/profile.go goroutine

# メモリ統計
docker compose exec backend go run scripts/profile.go stats
```

**期待値**:
- ✅ メモリ使用量 < 500MB
- ✅ ゴルーチン数 < 50
- ✅ GC 回数 < 50/分

#### Lighthouse フロントエンド監査

```bash
node frontend/scripts/lighthouse-audit.js
```

**期待値**:
- ✅ Performance > 75
- ✅ Accessibility > 90
- ✅ Best Practices > 90
- ✅ SEO > 90
- ✅ FCP < 1800ms, LCP < 2500ms, CLS < 0.1

---

## 🔍 テスト実行手順

### 1. 環境構築

```bash
# コンテナ起動・DB 初期化
make up

# マイグレーション確認
docker compose exec db psql -U postgres -d animal_ekarte -c "\dt"
```

### 2. ユニットテスト実行

```bash
# バックエンド
docker compose exec backend go test ./... -v -cover

# フロントエンド
docker compose exec frontend pnpm test:run
```

### 3. API 統合テスト実行

```bash
# 認証フロー
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'

# 飼主作成テスト
curl -X POST http://localhost:8080/api/v1/owners \
  -H "Authorization: Bearer {jwt_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "owner_name":"テスト飼主",
    "owner_kana":"テストシユウ",
    "phone":"09012345678"
  }'
```

### 4. E2E テスト実行

```bash
pnpm test:e2e
```

### 5. 負荷テスト実行

```bash
# k6 API エンドポイントテスト
k6 run load-tests/k6-api-endpoints.js

# k6 スパイクテスト
k6 run load-tests/k6-spike-test.js
```

### 6. パフォーマンスプロファイリング

```bash
# バックエンド
docker compose exec backend go run scripts/profile.go stats

# フロントエンド
node frontend/scripts/lighthouse-audit.js --url http://localhost:3000
```

---

## 📈 テスト結果報告

### テスト完了チェックリスト

- [ ] ユニットテスト PASS 100%
- [ ] API 統合テスト PASS 100%
- [ ] E2E テスト PASS >= 90%
- [ ] 負荷テスト: p95 < 500ms, error < 10%
- [ ] メモリ < 500MB, ゴルーチン < 50
- [ ] Lighthouse: Performance > 75, A11y > 90
- [ ] RBAC 権限チェック PASS 100%
- [ ] DB マイグレーション PASS

### レポート出力

```markdown
## 統合テスト結果サマリー (2026-04-23)

### ユニットテスト
- Backend: 384 tests, PASS 384/384 (100%)
- Frontend: 465 tests, PASS 465/465 (100%)

### API 統合テスト
- 飼主管理: 6 scenarios, PASS 6/6
- 予約管理: 4 scenarios, PASS 4/4
- 医療記録: 5 scenarios, PASS 5/5
- RBAC: 3 roles, PASS 3/3

### E2E テスト
- Total: 22 test cases
- PASS: 20/22 (90.9%)
- PENDING: 2 (LINE 予約 LIFF アプリ)

### パフォーマンス
- API p95: 320ms < 500ms ✅
- メモリ: 280MB < 500MB ✅
- Lighthouse: Performance 82/100 ✅

### セキュリティ
- JWT 認証: PASS ✅
- RBAC 権限: PASS ✅
- SQL インジェクション: PASS ✅
```

---

## 📝 注記

- テスト環境は Docker で隔離（本番環境への影響なし）
- テストデータはマイグレーション 004_seed_staging.sql で自動生成
- すべてのテストは再実行可能（冪等性確保）
- パフォーマンス基準は load-tests/README.md と一致

---

## 🔗 関連ドキュメント

- [FUNCTIONAL_TEST_REPORT.md](./FUNCTIONAL_TEST_REPORT.md) — 機能テストレポート
- [E2E_TESTING_GUIDE.md](./E2E_TESTING_GUIDE.md) — E2E テスト詳細
- [PERFORMANCE_PROFILING.md](./PERFORMANCE_PROFILING.md) — パフォーマンス詳細
- [load-tests/README.md](../load-tests/README.md) — 負荷テスト詳細
