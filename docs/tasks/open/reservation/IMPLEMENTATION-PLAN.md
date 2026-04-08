# LINE予約システム 実装計画

> **仕様書**: `docs/line-reseavation.md` セクション15
> **タスク一覧**: `docs/tasks/open/reservation/00-OVERVIEW.md`
> **作成日**: 2026-04-08

---

## ブランチ戦略

```
main（開発統合ブランチ）
  ↑ PR merge（Phase完了ごと）
feature/line-reservation（メイン作業ブランチ）
  ├── feature/line-reservation/phase1-db
  ├── feature/line-reservation/phase2-admin-api
  ├── feature/line-reservation/phase3-liff-api
  ├── feature/line-reservation/phase4-admin-fe
  ├── feature/line-reservation/phase5-liff-app
  ├── feature/line-reservation/phase6-line
  └── feature/line-reservation/phase7-deploy
```

**運用ルール**:
- `feature/line-reservation` を `main` から切り出す
- Phase ごとにサブブランチを切り、完了後 `feature/line-reservation` にマージ
- Phase 1〜3 完了時点で中間PR を `main` に出す（バックエンドのみ先行マージ）
- 全Phase完了後に最終PRを `main` にマージ

---

## Phase別実装計画

### Phase 1: DB・モデル基盤（1日目）

**ブランチ**: `feature/line-reservation/phase1-db`

| 順序 | タスク | 作業内容 | 成果物 |
|------|--------|---------|--------|
| 1-1 | TASK-RES-001 | マイグレーションSQL作成 | `backend/migrations/003_line_reservation.sql` |
| 1-2 | TASK-RES-002 | Goモデル拡張 + 新規定義 | 既存model/*.go 修正 + 新規4ファイル |
| 1-3 | TASK-RES-003 | シードデータ作成 | `backend/migrations/003_line_reservation_seed.sql` |
| 1-4 | — | codegen 実行 | `make codegen` → models.ts 更新確認 |

**完了確認**:
```bash
docker compose exec backend psql -c "\dt reservation_*"
docker compose exec backend psql -c "\d+ service_types"  # 新カラム確認
docker compose exec backend psql -c "\d+ staffs"         # 新カラム確認
docker compose exec backend psql -c "\d+ reservation_appointments"  # 新カラム確認
make codegen  # models.ts にTypeScript型が生成される
```

**コミット**: `feat(reservation): add LINE reservation DB schema and models`

---

### Phase 2: バックエンド管理者API（2〜4日目）

**ブランチ**: `feature/line-reservation/phase2-admin-api`

| 順序 | タスク | 作業内容 | 目安 |
|------|--------|---------|------|
| 2-1 | TASK-RES-010 | 基本設定API（GET/PUT） | 0.5日 |
| 2-2 | TASK-RES-011 | コースCRUD API（7エンドポイント） | 0.5日 |
| 2-3 | TASK-RES-012 | スタッフCRUD API（7エンドポイント） | 0.5日 |
| 2-4 | TASK-RES-013 | スタッフ個人設定API（GET/PUT/DELETE） | 0.5日 |
| 2-5 | TASK-RES-014 | 予約管理API（月/日表示 + 手動入力 + キャンセル） | 0.5日 |
| 2-6 | TASK-RES-015 | 顧客管理API（一覧 + オーナー紐付け） | 0.25日 |
| 2-7 | TASK-RES-016 | ルーティング登録・DI配線 | 0.25日 |

**実装パターン**（全API共通）:
```
handler/reservation_xxx_handler.go   — HTTPハンドラ
handler/reservation_xxx_request.go   — リクエストバインド
handler/reservation_xxx_response.go  — レスポンス変換
service/reservation_xxx_service.go   — ビジネスロジック
repository/reservation_xxx_repository.go — GORM操作
```

**完了確認**:
```bash
# 全エンドポイントの疎通テスト（curl or httpie）
docker compose exec backend go test ./internal/handler/... -run Reservation -v
docker compose exec backend go test ./internal/service/... -run Reservation -v
```

**コミット（タスクごと）**:
```
feat(reservation): add reservation settings API
feat(reservation): add reservation courses CRUD API
feat(reservation): add reservation staffs CRUD API
feat(reservation): add staff schedule API
feat(reservation): add reservation management API
feat(reservation): add reservation customers API
feat(reservation): register all reservation routes
```

---

### Phase 3: バックエンド公開API + コアロジック（5〜8日目）

**ブランチ**: `feature/line-reservation/phase3-liff-api`

| 順序 | タスク | 作業内容 | 目安 |
|------|--------|---------|------|
| 3-1 | TASK-RES-020 | LIFF認証ミドルウェア | 0.5日 |
| 3-2 | TASK-RES-022 | **時間枠生成エンジン**（★最重要） | 1.5日 |
| 3-3 | TASK-RES-023 | 空き日付計算 | 0.5日 |
| 3-4 | TASK-RES-024 | 予約制限チェック + 楽観ロック | 0.5日 |
| 3-5 | TASK-RES-025 | 指名なし委譲ロジック | 0.25日 |
| 3-6 | TASK-RES-021 | 公開予約フローAPI（9エンドポイント） | 0.75日 |

**TASK-RES-022（時間枠生成エンジン）の実装順序**:
```
1. テストケースを先に書く（TDD）
   - timeslot_engine_test.go に10パターン
2. 基本ロジック実装（営業時間内で枠生成）
3. 休憩時間除外
4. 既存予約除外
5. 個人設定（shift_entries）の反映
6. minimize_gaps モード実装
7. 指名なしの全スタッフ統合
```

**完了確認**:
```bash
# 時間枠エンジンのユニットテスト
docker compose exec backend go test ./internal/service/... -run TimeSlot -v -count=1

# LIFF API疎通テスト（LIFF認証はモックで）
docker compose exec backend go test ./internal/handler/... -run Liff -v
```

**コミット（タスクごと）**:
```
feat(reservation): add LIFF auth middleware
feat(reservation): implement time slot generation engine
feat(reservation): add available dates calculation
feat(reservation): add reservation validators with optimistic locking
feat(reservation): add no-staff delegation logic
feat(reservation): add LIFF public API endpoints
```

**★ ここで中間PR → staging マージ（バックエンド完成）**

---

### Phase 4: 管理画面フロントエンド（9〜13日目）

**ブランチ**: `feature/line-reservation/phase4-admin-fe`

| 順序 | タスク | 作業内容 | 目安 |
|------|--------|---------|------|
| 4-1 | TASK-RES-030 | Feature scaffolding + ルーティング + サイドメニュー | 0.5日 |
| 4-2 | TASK-RES-033 | 基本設定画面 | 1日 |
| 4-3 | TASK-RES-031 | コース設定画面（CRUD + is_internal + 並び順） | 1日 |
| 4-4 | TASK-RES-032 | スタッフ設定画面（CRUD + 非対応コース + staff_type） | 1日 |
| 4-5 | TASK-RES-034 | 個人設定画面（ガントチャート + モーダル） | 1日 |
| 4-6 | TASK-RES-035 | **予約状況画面**（★最も複雑：月/日表示 + 手動入力 + キャンセル） | 1.5日 |
| 4-7 | TASK-RES-036 | ページ編集画面 | 0.5日 |

**実装順序の理由**:
- 基本設定（4-2）を先に作る → コース/スタッフの設定基盤になる
- コース（4-3）→ スタッフ（4-4）の順 → スタッフの非対応コース選択にコースデータが必要
- 個人設定（4-5）はスタッフ画面の後 → スタッフ選択ドロップダウンが必要
- 予約状況（4-6）が最後 → 全設定が完了した状態でテストできる

**コミット（タスクごと）**:
```
feat(reservation): scaffold reservations feature module
feat(reservation): add reservation settings page
feat(reservation): add reservation courses management page
feat(reservation): add reservation staffs management page
feat(reservation): add staff schedule management page
feat(reservation): add reservation calendar page (month/day view)
feat(reservation): add reservation page editor
```

---

### Phase 5: LIFF App（14〜19日目）

**ブランチ**: `feature/line-reservation/phase5-liff-app`

| 順序 | タスク | 作業内容 | 目安 |
|------|--------|---------|------|
| 5-1 | TASK-RES-040 | プロジェクトセットアップ + Docker + LIFF SDK | 1日 |
| 5-2 | TASK-RES-049 | トップページ（アコーディオン + CTA） | 0.5日 |
| 5-3 | TASK-RES-041 | STEP 1: お客様情報入力（5フィールド） | 0.5日 |
| 5-4 | TASK-RES-042 | STEP 2: コース選択 | 0.5日 |
| 5-5 | TASK-RES-043 | STEP 3: スタッフ選択 | 0.5日 |
| 5-6 | TASK-RES-044 | STEP 4: 日付選択（カレンダー）★ | 1日 |
| 5-7 | TASK-RES-045 | STEP 5: 時間選択 | 0.5日 |
| 5-8 | TASK-RES-046 | STEP 6: 要望入力 | 0.25日 |
| 5-9 | TASK-RES-047 | STEP 7: 確認画面 | 0.5日 |
| 5-10 | TASK-RES-048 | STEP 8: 完了 + エラー + メンテナンス画面 | 0.5日 |
| 5-11 | TASK-RES-049 | マイ予約一覧 + キャンセル | 0.5日 |

**実装順序**: トップページ → STEP 1〜8 を順番に → マイ予約。フローの順にやるのが最も効率的。

**コミット（画面ごと）**:
```
feat(liff-app): initialize LIFF app project with Vite + React 19 + Tailwind
feat(liff-app): add top page with accordion sections
feat(liff-app): add customer info input (step 1)
feat(liff-app): add course selection (step 2)
feat(liff-app): add staff selection (step 3)
feat(liff-app): add date selection calendar (step 4)
feat(liff-app): add time selection (step 5)
feat(liff-app): add request input (step 6)
feat(liff-app): add confirmation page (step 7)
feat(liff-app): add completion, error, and maintenance pages (step 8)
feat(liff-app): add my reservations and cancellation
```

---

### Phase 6: LINE連携（20日目）

**ブランチ**: `feature/line-reservation/phase6-line`

| 順序 | タスク | 作業内容 | 目安 |
|------|--------|---------|------|
| 6-1 | TASK-RES-060 | LINE Messaging API Push通知 | 0.5日 |
| 6-2 | TASK-RES-061 | メール通知 | 0.5日 |

**コミット**:
```
feat(reservation): add LINE push notification service
feat(reservation): add email notification service
```

---

### Phase 7: 結合・デプロイ（21〜22日目）

**ブランチ**: `feature/line-reservation/phase7-deploy`

| 順序 | タスク | 作業内容 | 目安 |
|------|--------|---------|------|
| 7-1 | TASK-RES-069 | CORS設定 | 0.25日 |
| 7-2 | TASK-RES-070 | Docker Compose更新 | 0.25日 |
| 7-3 | TASK-RES-071 | 結合テスト（テスト用LINEアカウントで実施） | 1日 |
| 7-4 | TASK-RES-072 | stagingデプロイ | 0.5日 |

**結合テストの実施手順**:
```
1. テスト用LINE公式アカウントAでLIFF App動作確認
   - 予約フロー全8ステップ通し
   - Push通知受信確認
   - キャンセル確認
2. 管理画面で予約状況確認
   - source='line' の予約が表示される
   - 手動予約入力
   - 月/日表示の切替
3. テスト用LINE公式アカウントBでマルチクリニック確認
4. エッジケース確認
   - 同日予約制限
   - メンテナンス中表示
   - 時間枠競合（2端末で同時予約）
```

**コミット**:
```
feat(reservation): add CORS configuration for LIFF app
feat(reservation): add liff-app to docker-compose
test(reservation): add integration test scenarios
feat(reservation): configure staging deployment
```

**★ 最終PR → staging マージ**

---

## スケジュールサマリ

| Phase | 日数 | 累計 | マイルストーン |
|-------|------|------|-------------|
| Phase 1: DB | 1日 | 1日 | テーブル作成・モデル定義完了 |
| Phase 2: 管理者API | 3日 | 4日 | 管理者APIフル稼働 |
| Phase 3: LIFF API | 4日 | 8日 | **バックエンド完成 → 中間PR** |
| Phase 4: 管理画面FE | 5日 | 13日 | 管理画面で予約管理可能 |
| Phase 5: LIFF App | 6日 | 19日 | LINE予約フロー動作 |
| Phase 6: LINE連携 | 1日 | 20日 | Push通知・メール通知稼働 |
| Phase 7: 結合・デプロイ | 2日 | **22日** | **staging デプロイ完了** |

**総工数: 約22営業日（約1ヶ月）**

---

## PR戦略

| PR | タイミング | ベース | 内容 |
|----|----------|--------|------|
| PR #1（中間） | Phase 3 完了後 | main | Phase 1〜3（DB + バックエンドAPI全量） |
| PR #2（最終） | Phase 7 完了後 | main | Phase 4〜7（フロントエンド + LIFF + 連携 + デプロイ） |

**2つのPRに分ける理由**:
- バックエンドを先にマージすることで、フロントエンド開発中にAPI仕様の手戻りを防ぐ
- レビュー範囲を分割し、レビュー品質を担保する
- Phase 4〜5 はバックエンドAPIが main で動いている状態で開発・テストできる

---

## 開始手順

```bash
# 1. main を最新に
git checkout main
git pull origin main

# 2. メイン作業ブランチを切る
git checkout -b feature/line-reservation

# 3. Phase 1 サブブランチを切る
git checkout -b feature/line-reservation/phase1-db

# 4. 実装開始（TASK-RES-001 から）
```
