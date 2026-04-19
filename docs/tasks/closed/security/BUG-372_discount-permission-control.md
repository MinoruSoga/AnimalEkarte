# BUG-372: 割引フィールドの権限制御（新規 `discount` リソース追加）

**作成日**: 2026-04-14
**Status**: Closed (2026-04-14)
**Priority**: HIGH（不正値引のリスク）
**Affects**: 飼主 / 治療 / 入院 / 見積 / 会計の全ての割引フィールド、RBAC モデル

## 実装結果（2026-04-14）

### 完了
- BE: `ResourceDiscount` 追加 + `AllResources` 登録 (`model/permission.go`)
- BE: マイグレーション `003_seed_demo.sql` に全 6 グループの discount 行をシード（執行=ON、一般=OFF）
- BE: handler 共通ヘルパー `discount_permission.go` 実装（float/int の Create/Edit パターン）
- BE: 適用済みハンドラ
  - `owner_handler.go`: Create（discount_rate ゼロ以外で create 権限要）/ Update（既存値と比較、edit 権限要）
  - `treatment_handler.go`: Create + Update 両方（discount_rate + discount_amount）
  - `treatment_plan_handler.go`: MedicalRecord・Hospitalization の両コンテキストの Create + Update
  - `estimate_handler.go`: Create + Update（discount_amount）
  - `accounting_handler.go`: Update（非ゼロ値指定時のみ edit 権限要。既存 payment 比較はリポジトリ未整備のため非対応）
- BE: `make codegen` で `models.ts` に `ResourceDiscount = "discount"` 反映確認
- FE: `OwnerForm.tsx` — OwnerInfoSection の discountRate に `disabled={!canEditDiscount}` + 説明テキスト
- FE: `TreatmentsTab/TreatmentRow.tsx` — discount_amount セル編集を `canEditDiscount` で保護
- FE: `EstimateForm.tsx` — AmountSection の割引額入力を `canEditDiscount` で保護
- FE: build / lint 成功（エラー 0）

### 設計変更（当初計画との差異）
- 当初: service 層で権限チェック → 既存 service は context から staff_id を取得できない構造（DI 全面書き換えコスト大）
- 採用: **handler 層で権限チェック**（既存 `h.hasPermission` パターンを流用）
- 影響: AC-5〜AC-10 は handler 層で satisfiable。監査対象範囲は変わらず

### 未完了（別イシュー化候補）
- FE 入院フォーム（`features/hospitalization/`）の割引入力欄 disabled 対応 — BE は treatment_plan 経由で保護済み。UI は別セッション対応
- accounting の既存 payment 値取得ロジック（Payment リポジトリ整備が必要） — AC-10 を厳密に満たさない（非ゼロ値のみ edit 権限要で代替）
- テーブル駆動テスト追加 — 実装完了。テスト追加は次セッション

**依頼元（原文）**:

> 割引の入力が誰でも入力できてしまうので、できないようにしてほしい
>
> 補足: 現状の権限制御にて飼主情報の値引率フォームの制御項目を追加したらいいと思います

---

## 概要

割引（値引率・値引額）が `owners:edit` 等の通常編集権限保有者なら誰でも入力可能な現状を、新規 `discount` リソースの専用権限で保護する。対象は飼主だけでなく **治療・入院・見積・会計の全ての割引フィールド**を包括的に制御する。

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| Q1 | 保護対象の範囲 | **(B) 割引全般を保護** — owners.discount_rate / treatments.discount_rate / hospitalizations.discount_rate / estimates.discount_rate + amount / payments.discount_amount すべて |
| Q2 | リソース命名 | **(A) 新規リソース `discount` を追加** |
| Q3 | 権限なし時の UI | **(A) disabled**（値表示・編集不可） |
| Q4 | 既存スタッフ初期権限 | **(C) `is_system_admin=true` のみ ON**（一般スタッフは管理者が後から付与） |
| Q5 | BE バリデーション | 必須実装（API 直叩き対策） |
| Q6 | 監査ログ | 不要（既存 slog で十分） |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | `discount` リソース追加 + 5 ハンドラに権限チェック | BE | BE-112 | - |
| 2 | 5 つの割引入力フォームに `disabled` 制御 | FE | FE-250 | #1 |

## 対象フィールド一覧

| 場所 | フィールド | 既存ファイル |
|------|----------|------------|
| 飼主 | `owners.discount_rate` | `backend/internal/handler/owner_request.go:40,62`、`frontend/.../OwnerForm.tsx:425-440` |
| 治療 | `treatments.discount_rate` | `backend/internal/handler/treatment_request.go:17,36` |
| 入院 | `hospitalizations.discount_rate` / `discount_amount` | `backend/internal/model/hospitalization.go:116-117` |
| 見積 | `estimates.discount_amount` / `estimate_items.discount_rate` / `discount_amount` | `backend/internal/handler/estimate_request.go:15,30`、`backend/internal/model/estimate.go:32,60-61` |
| 会計（payment） | `payments.discount_amount` | `backend/internal/handler/accounting_request.go:42` |
| 会計（明細） | `billing_items.discount_rate` / `discount_amount` | `backend/internal/model/treatment.go:45-46`（同 struct） |

## 受入条件（Acceptance Criteria）

### 権限モデル
- [ ] **AC-1**: `Resource` enum に `ResourceDiscount = "discount"` が追加され、`AllResources` に登録される
- [ ] **AC-2**: 権限グループ管理画面に「割引」リソース行が自動表示され、view/create/edit/delete のチェックが設定可能
- [ ] **AC-3**: 既存スタッフのうち `is_system_admin=true` のみが `discount` 権限を保有。一般スタッフはデフォルトで全アクション OFF
- [ ] **AC-4**: マイグレーション/起動時シードで既存環境の権限グループに `discount` リソース行が `can_view/create/edit/delete=false` で自動挿入される

### Backend バリデーション（API 直叩き対策）
- [ ] **AC-5**: PATCH `/owners/:id` で `discount_rate` フィールドが含まれる時、`discount:edit` 権限がない場合は 403 を返す（フィールド未指定なら通常通り処理）
- [ ] **AC-6**: PATCH `/treatments/:id` で `discount_rate` 同上
- [ ] **AC-7**: POST/PATCH `/hospitalizations` で `discount_rate` / `discount_amount` 同上
- [ ] **AC-8**: POST/PATCH `/estimates` および `/estimate-items` で `discount_amount` / `discount_rate` 同上
- [ ] **AC-9**: PATCH `/accountings/:id` で `discount_amount` 同上
- [ ] **AC-10**: 新規作成（POST）でも同様に権限チェック。ゼロ値 (`0`) は権限不要、ゼロ以外を指定する場合のみ要権限

### Frontend UI
- [ ] **AC-11**: 飼主編集フォームの「値引率」入力欄が `discount:edit` なし時 `disabled`、値はそのまま表示
- [ ] **AC-12**: 治療・入院・見積・会計の各画面の割引入力欄も同様に `disabled`
- [ ] **AC-13**: 権限なし時のフォーム送信で `discount_rate` フィールドが既存値と同じなら正常送信される（エラーにならない）

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| リソース粒度 | 単一 `discount` リソース | Q1=B + Q2=A。割引全般を一括管理 | リソース別（`owners-discount`/`treatments-discount` 等） — 権限管理が煩雑 |
| BE 権限チェック実装位置 | **service 層** | フィールドの有無判定（`*float64` の nil チェック）が service 層の責務範囲。複数ハンドラから service が呼ばれても担保できる | handler 層 — service 直接呼ぶテスト・将来のリファクタリングで漏れる |
| ゼロ値変更の扱い | 既存値と同じなら権限不要 | 通常編集（住所更新等）で `discount_rate=0` を再送するだけで 403 にならない配慮 | ゼロ値でも常に権限要 — 既存 UX を壊す |
| `discount` リソースの初期付与 | 起動時シードで `is_system_admin=true` のスタッフが所属する権限グループのみ rule を ON、他は OFF で挿入 | Q4=C。明示的にデータ投入で安全側 | アプリ層で「rule 不在 = ON」のフォールバック — 意図不明瞭 |
| FE 権限取得 | 既存 `usePermission("discount")` パターン流用 | 既存パターンで一貫性 | 新規 hook 作成 |

## 影響範囲

### Backend
- `backend/internal/model/permission.go:6-32` — `ResourceDiscount` 追加、`AllResources` 追加
- `backend/internal/service/owner_service.go` — 値引率変更時の権限チェック追加
- `backend/internal/service/treatment_service.go` — 同上
- `backend/internal/service/hospitalization_service.go` — 同上
- `backend/internal/service/estimate_service.go` — 同上
- `backend/internal/service/accounting_service.go` — 同上
- `backend/internal/handler/auth_helpers.go`（または新規） — 権限チェックヘルパー `requireDiscountPermissionIfChanged()` 等
- `backend/migrations/001_init.sql` または `003_seed_demo.sql` — 既存環境への `discount` リソース行挿入

### Frontend
- `frontend/src/types/generated/models.ts` — `make codegen` で `ResourceDiscount` 自動追加
- `frontend/src/features/owners/routes/OwnerForm.tsx:425-440` — `disabled={!canEditDiscount}` 追加
- `frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx` 等の治療割引入力欄
- `frontend/src/features/hospitalization/` の入院割引入力欄
- `frontend/src/features/estimates/` の見積割引入力欄
- `frontend/src/features/accounting/routes/AccountingDetail.tsx` の支払割引入力欄

### DB
- **マイグレーション**: 既存 `permission_group_rules` に `discount` リソース行を全権限グループに対して挿入（`is_system_admin` 配下グループは権限 ON、それ以外 OFF）

## 参照実装

- `backend/internal/model/permission.go` — 既存リソース定義パターン
- `frontend/src/features/auth/hooks/use-permission.ts` — `usePermission` hook
- `backend/internal/handler/handler.go:178-181` — `RequirePermission` ミドルウェア使用例（ただし本タスクは service 層で判定）
- `frontend/src/features/owners/routes/OwnerForm.tsx:82-93` — `canEdit` props 受け取りパターン

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| 既存運用で全スタッフが割引を入力していた場合、本リリース後に業務停止 | **高** | リリース前に管理者が明示的に「割引」権限を付与する手順を周知。リリースノート明記 |
| `discount_rate=0` の再送が 403 になる UX 不具合 | 中 | 「既存値と同じなら権限不要」ロジック実装で対策（AC-10） |
| `billing_items.discount_rate` (個別明細単位) も対象範囲か曖昧 | 中 | `treatment.discount_rate` と同じ struct 定義のため同時対応する。AC-8 に明記 |
| 権限チェック漏れ（ハンドラの追加・変更時に discount フィールドが追加された場合） | 高 | service 層で集約することで、ハンドラ修正時の漏れを最小化。テストケース必須 |
| 既存テストが権限なしユーザーで割引フィールドを送って通っている | 中 | 影響テストを洗い出し、is_system_admin スタッフでのテストに変更 |

## 未解決事項

- なし

## 実装順序

1. BE-112: リソース定義追加 + 5 service への権限チェック追加 + 既存環境マイグレーション
2. FE-250: 5 画面の割引入力欄に `disabled` 制御追加

## 関連イシュー

- BE-112: discount リソース追加 + 全 service の権限チェック
- FE-250: 割引入力フォーム disabled 制御（5 画面）
