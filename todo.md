# AnimalEkarte — Project TODO

> 作成: 2026-06-26 | 監査範囲: backend / frontend / CI-CD / migrations / docs / GitHub Issues
> 旧: 2026-06-17 E2E 検証レポート（内容は E2E が 136/136 PASS で完了済み、今後の実施事項に置換）
>
> **PR #186 マージはユーザー所有アクション**（このリストに含めない）

---

## P0 — ブロッカー / セキュリティ即対応

- [ ] **[USER] 平文シークレット 7 件のローテーション（本番デプロイ前必須）**
  - Issues #89/#91/#96-99/#109 で STG 限定 park 済み
  - 内容: GitHub PAT 他 7 件を AWS Secrets Manager 等へ移行 + 旧値ローテーション
  - 根拠: memory `issue_plaintext_secrets_stg_park_20260621.md`
  - **実施者: ユーザー**（credential 変更は安全境界外）

- [ ] **[USER] PR #186 を main へマージ**
  - CI: Detect changes / Backend / Frontend / Codegen Sync / Playwright E2E / ShellCheck 全 SUCCESS
  - 根拠: `gh pr view 186`
  - **実施者: ユーザー**

- [ ] **STG リリース前チェックリスト完了 (#123)**
  - P1 全項目の実機 STG 検証が未完了 → STG リリース BLOCKED
  - 手順: PR #186 マージ後 → STG デプロイ (`db_reset=true` dispatch) → Issue #123 の P1 チェックリスト実施
  - 注: `billing_status` の正値は `'waiting'`（Issue 本文の `'unpaid'` は誤り。訂正済み `docs/release-checks/ISSUE-123-release-check-corrections.md`）
  - 根拠: Issue #123 OPEN

---

## P1 — セキュリティ / 高リスク欠落

- [ ] **[#196] repository 層テスト補強 — clinic_id 隔離不変条件（セキュリティ最重要）**
  - 現状: `backend/internal/repository/` に source 101 / test 9（≈9%）
  - **完了**: `owner_pet_clinic_isolation_test.go` 追加 (2026-06-26)
    - `OwnerRepository.FindByID` / `Update` / `Delete` — clinic_id 隔離 5 テスト PASS
    - `PetRepository.FindByID` / `Delete` — clinic_id 隔離テスト PASS
  - **残**: `reservation_repository.go` / `accounting_repository*.go` の FindByID / Update / Delete 隔離テスト未追加
  - 根拠: Issue #196 + `find backend/internal/repository -name '*_test.go'`（9 件）

- [ ] **[#185] ADR-003 payment_method 残論点 — PO 決裁取得**
  - エンジニア案は `docs/adr/003-payment-method-identity-and-consistency.md` に提示済み
  - 論点: DB 制約・system_key 導入・レガシー NULL 行・命名統一（4 件）
  - 決裁確定後に個別実装 Issue を起票して着手（本 Issue での実装は禁止）
  - 根拠: Issue #185 OPEN + ADR-003 (Proposed)

---

## P2 — CI/CD 品質

- [ ] **[#193] actionlint ジョブ追加**
  - 現状: `grep -rn actionlint .github/workflows` → 0 件。YAML 構文エラーは push 後まで検知不可
  - スコープ: `.github/workflows/**` 変更 PR のみ起動、所要 <1 分
  - 根拠: Issue #193

- [ ] **[#195] `frontend-deploy.yml:45` の `setup-node@v4` を `@v6` へ統一**
  - ドリフト: `frontend-deploy.yml` = `@v4`、`ci.yml:233` / `performance-tests.yml:44,231` = `@v6`
  - 合わせてピン記法ポリシー（メジャータグ vs SHA ピン）を全 8 ワークフローで統一
  - 根拠: Issue #195 + `grep -rn setup-node .github/workflows/`

- [ ] **[#194] カバレッジ計測ポリシー定義 + 段階的しきい値導入**
  - Phase 1: 除外対象明文化（`cmd/` / `migrations/` / `*_mock.go` / codegen 出力）
  - Phase 2: PR コメントへ数値表示 + 低下時 warn
  - Phase 3: patch coverage しきい値（warn 先行、fail は段階的導入）
  - 現状: backend ≈46.2%（非ゲート）、`frontend/vite.config.ts` に `coverage.thresholds` 未設定
  - 根拠: Issue #194 + `ci.yml:174` (非ゲートコメント) + `vite.config.ts:140`

- [ ] **STG スモークテスト login/CRUD の復元**
  - 現状: `stg-smoke.yml` は `/health` のみ（`STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` secrets 未設定で約1年無効）
  - 手順: `gh secret set STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` 設定後、`commit 281a561e` のステップを戻す
  - 根拠: `.github/workflows/stg-smoke.yml:5-10`

---

## P2 — 機能実装（仕様確定済み）

- [ ] **[#159] 麻酔処置履歴（スキーマ変更なし・即着手可能）**
  - 定義: `Treatment JOIN Procedure WHERE anesthesia != 'none'`（`Procedure.Anesthesia` enum 既存）
  - BE: 既存 `GET /v1/pets/:id/treatment-history` の cross-pet 集約パターン適用
  - FE: `OwnerReport.tsx` に `showAnesthesia=true` セクション追加
  - 根拠: Issue #159 + `backend/internal/model/procedure.go:9-15`

- [ ] **[#159] 手術処置履歴（migration 009 必要）**
  - `procedures` に `is_surgery BOOLEAN DEFAULT false` 追加（新規 migration 009、additive、後方互換）
  - BE: `treatment-history` に `is_surgery` フィルタ追加（handler / service / repository）
  - FE: 手術フィルタ済みセクション追加
  - 根拠: Issue #159 + `backend/internal/model/procedure.go:18-34`（`is_surgery` 未存在）

- [ ] **[#160] 健康診断カテゴリー化（マスタ投入のみ、コード変更不要）**
  - 既存 `検査(Examination)` でカテゴリ→項目→結果階層が実装済み
  - 新規 migration で `exam_types`（健診カテゴリ・`parent_id` 階層）+ `exam_type_fields` 投入
  - 根拠: Issue #160 + `backend/migrations/003_seed_demo.sql:585-644`（Examination 種別実例）

- [ ] **[#190] 明細兼領収書 帳票レイアウト設定 UI 実装**
  - migration 008 でスキーマ済み（`show_logo` / `show_registration_warning` / `show_item_category` / `footer_note`）
  - 残件: (a) 設定 UI 全セクショントグル (b) 表示順 (c) ロゴ画像アップロード
  - **ブロッカー**: クライアントの「別紙（現行の明細書）」仕様が未提供（設定機構のみ先行着手可能）
  - 根拠: Issue #190 + `backend/migrations/008_add_accounting_document_layout_settings.sql`

---

## P3 — インフラ / パフォーマンス

- [ ] **P2 Terraform（internal ALB + VPC Origin）本番適用**
  - **BLOCKED**: `infra/terraform/terraform.tfvars` 不在 → full plan 実行不可
  - 前提: `db_password` 供給 + `alb_internal = true` 設定後 `terraform plan` 実行
  - 注: VPC Origin は ALB SG が Service-SG を source 参照必須（CIDR では 504）
  - 根拠: `docs/infra/P2_TERRAFORM_PLAN_RUNBOOK.md`

- [ ] **N+1 クエリ解消 — LIFF 日付一覧 capacity チェック**
  - 対象: `backend/internal/service/liff_service_availability.go:62,73`
  - 内容: 日付一覧取得時に capacity を 1 件ずつクエリ → バッチ取得（IN 句等）で改善
  - 根拠: `liff_service_availability.go:62-73`（TODO コメント実在）

- [ ] **STG 費用最適化プラン実施**
  - 根拠: `docs/tasks/open/STG-COST-OPTIMIZATION-PLAN-2026-06-01.md`

---

## PO 決裁待ち（実装着手禁止）

| # | 内容 | 決裁待ち事項 |
|---|------|------------|
| #185 | payment_method system_key 導入・DB 制約・命名統一 | ADR-003 論点 4 件 |
| #159 | 手術処置の `is_surgery` スキーマ | 手術定義・分類方針 |
| #160 | 健診カテゴリマスタ内容 | カテゴリ名・階層構成 |
| #190 | 帳票レイアウト仕様詳細 | クライアント別紙提供待ち |
| #179 | 月次集計レポート残件 (#191/#192 CLOSED、#190 追跡中) | 帳票設定との整合 |
| #159/#160 | 麻酔処置ページ / 手術処置ページの UX 詳細 | PO UX 確認待ち |

---

## ユーザー所有アクション一覧

| アクション | 根拠 |
|-----------|------|
| GitHub PAT + 7 件シークレット ローテーション | 本番前必須、credential 変更は安全境界外 |
| PR #186 の main へのマージ | 外部 write 操作 |
| STG `db_reset=true` デプロイ dispatch | 本番影響作業 |
| `gh secret set STG_DEMO_EMAIL / STG_DEMO_PASSWORD` | secrets 設定 |
| Terraform tfvars 用意 + `terraform apply` 承認 | インフラ破壊的変更 |

---

## 証跡サマリー（検査対象一覧）

| 検査対象 | 内容 |
|---------|------|
| GitHub Issues（open 全 10 件） | #123 / #159 / #160 / #179 / #185 / #190 / #193 / #194 / #195 / #196 |
| PR #186 | CI 全チェック SUCCESS、E2E SUCCESS 確認 |
| `.github/workflows/` 全 8 ファイル | action バージョン・coverage ゲート・secrets 参照確認 |
| `backend/internal/repository/` | test 密度実測（source 101 / test 8 ≈ 8%） |
| `backend/migrations/` 001-008 | 次番 009 未作成確認 |
| `docs/adr/003-*.md` | Proposed / PO 待ち確認 |
| `frontend/vite.config.ts:140` | `coverage.thresholds` 未設定確認 |
| `liff_service_availability.go:62,73` | N+1 TODO 実存確認 |
| `docs/tasks/open/` | STG コスト最適化・デプロイ準備ドキュメント確認 |
| `stg-smoke.yml` | login/CRUD smoke 無効化確認 |
