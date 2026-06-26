# AnimalEkarte — Project TODO

> 更新: 2026-06-26（lab_import BE 自動検証完了: gofmt PASS / scoped test PASS / lint 0 issues。format+lint 修正 5 ファイル未コミット。handler/router/FE が残作業）
> 前回: #193/#194/#195/#196/#185/#197/#198/#159/#160/#179/#190/#180/#178 全 CLOSED 確認 → 除去。lab_import Phase 0+1 コミット済み（c224ec81/73340eac）。ADR-003 Proposed→Decided 反映
> 監査範囲: backend / frontend / CI-CD / migrations / docs / GitHub Issues / git untracked
> **PR マージ・push・外部書き込みはユーザー所有アクション。**

---

## P0 — ブロッカー / セキュリティ即対応

- [ ] **[USER] 平文シークレット 7 件のローテーション（本番デプロイ前必須）**
  - Issues #89/#91/#96-99/#109 で STG 限定 park 済み
  - 内容: GitHub PAT 他 7 件を AWS Secrets Manager 等へ移行 + 旧値ローテーション
  - 根拠: memory `issue_plaintext_secrets_stg_park_20260621.md`
  - **実施者: ユーザー**（credential 変更は安全境界外）

- [ ] **[USER] PR #186 を main へマージ**
  - CI: Detect changes / Backend / Frontend / Codegen Sync / Playwright E2E / ShellCheck / actionlint 全 SUCCESS
  - Vercel: SUCCESS / qlty check: SUCCESS（2026-06-26 06:13 確認）
  - 根拠: `gh pr view 186`
  - **実施者: ユーザー**

- [ ] **STG リリース前チェックリスト完了 (#123)**
  - P1 全項目の実機 STG 検証が未完了 → STG リリース BLOCKED
  - 手順: PR #186 マージ後 → STG デプロイ (`db_reset=true` dispatch) → Issue #123 の P1 チェックリスト実施
  - 注: `billing_status` の正値は `'waiting'`（Issue 本文の `'unpaid'` は誤り。訂正済み `docs/release-checks/ISSUE-123-release-check-corrections.md`）
  - 根拠: Issue #123 OPEN（2026-06-26 時点で唯一の open issue）

---

## P1 — 進行中 WIP（コミット待ち・実装続行）

### lab_import 外部検査連携（Dr.Wan Phase 0+1 scaffold）

**現在の実装状態:**

| ファイル | git 状態 | 規約チェック（2026-06-26） |
|---------|---------|------------|
| `backend/internal/service/lab_import_service.go` | committed (73340eac) + 未コミット修正 | ✅ gofmt/lint PASS |
| `backend/internal/service/lab_import_service_test.go` | committed (73340eac) + 未コミット修正 | ✅ gofmt PASS |
| `backend/internal/service/lab_import_examination_service.go` | committed (c224ec81) + 未コミット修正 | ✅ lint PASS (wrapcheck fix) |
| `backend/internal/service/lab_import_examination_service_test.go` | committed (c224ec81) + 未コミット修正 | ✅ lint PASS (gocritic fix) |
| `backend/internal/model/lab_import.go` | committed (73340eac) + 未コミット修正 | ✅ gofmt PASS |
| `backend/internal/repository/lab_import_repository.go` | committed (73340eac) + 未コミット修正 | ✅ lint PASS (LabImportDuplicateCheckerDB 追加) |
| `backend/migrations/005_add_lab_import_tables.sql` | committed (73340eac) | ✅ 規約適合 |
| `backend/migrations/006_add_exam_results_exam_id_index.sql` | **untracked** | ✅ 規約適合（CREATE INDEX IF NOT EXISTS） |
| `backend/internal/handler/lab_import_handler.go` | **未作成** | — |
| frontend lab_import UI | **未着手** | — |
| router 配線 | **未接続** | — |

**BE 自動検証結果（2026-06-26 Docker 実行）:**
- gofmt: ✅ PASS（全 6 ファイル clean）
- scoped go test: ✅ PASS（`ok github.com/animal-ekarte/backend/internal/service 0.801s`）
- golangci-lint: ✅ PASS（0 issues — gocritic/gofmt/wrapcheck 5 件修正後）

**修正内容（未コミット — format+lint cleanup）:**
- `model/lab_import.go`: gofmt タブ幅アライメント修正（LabInboundBatch + LabImportPreviewResponse struct）
- `service/lab_import_service.go`: gocritic rangeValCopy → index ベースに変更（ResultRows ループ）
- `service/lab_import_service_test.go`: gofmt 単行関数を複数行展開（newStubJobRepo）
- `service/lab_import_examination_service.go`: wrapcheck — ctx.Err() を apperrors.Wrap でラップ
- `service/lab_import_examination_service_test.go`: gocritic rangeValCopy（items ループ） + paramTypeCombine（stubDupChecker.IsDuplicate）

**着手タスク（優先順）:**

- [ ] **[AGENT] format/lint 修正 + DuplicateCheckerDB + migration をコミット（ユーザー承認後）**
  - 対象: `model/lab_import.go` / `service/lab_import_service.go` / `service/lab_import_service_test.go` / `service/lab_import_examination_service.go` / `service/lab_import_examination_service_test.go` / `repository/lab_import_repository.go` / `migrations/006_add_exam_results_exam_id_index.sql`
  - 根拠: gofmt PASS + scoped test PASS + lint 0 issues 確認済み

- [ ] **[AGENT] `lab_import_handler.go` 作成**
  - エンドポイント案:
    - `POST /v1/lab-import/preview` → `PreviewBatch`
    - `POST /v1/lab-import/jobs` → `CreateJob`
    - `GET  /v1/lab-import/jobs` → `ListJobs`
    - `GET  /v1/lab-import/jobs/:id` → `GetJob`
  - P5 準拠: `RequirePermission` 付与必須
  - `source_type=drwan/manual` は `PreviewBatch` で `blocked_reasons` 返却済み（handler でも 400 返却を検討）
  - 前提: WIP コミット後

- [ ] **[AGENT] router へ配線**
  - `NewLabImportJobRepository` / `NewLabImportEventRepository` / `NewLabImportJobService` の DI
  - 前提: handler 作成後

- [ ] **[PO 確認後] Frontend lab_import UI**
  - Phase 0: フィクスチャ入力のみ。Dr.Wan MDB / manual はサービス層でブロック済み
  - UX 仕様（操作者・操作画面・入力形式）は PO 未確認 → FE 着手前に確認要
  - 根拠: migration コメント「Phase BLOCKED — MDB スキーマ未確認」

---

## P2 — STG/CI 品質

- [ ] **STG スモークテスト login/CRUD の復元**
  - 現状: `stg-smoke.yml` は `/health` のみ（`STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` secrets 未設定で無効）
  - 手順: `gh secret set STG_DEMO_EMAIL` / `STG_DEMO_PASSWORD` 設定後、login/CRUD smoke ステップを復活
  - 根拠: `.github/workflows/stg-smoke.yml:5-9`（コメントに明記）
  - **実施者: ユーザー**（secret 設定は credential 操作）

---

## P3 — インフラ / パフォーマンス

- [ ] **P2 Terraform（internal ALB + VPC Origin）本番適用**
  - **BLOCKED**: `infra/terraform/terraform.tfvars` 不在 → full plan 実行不可
  - 前提: `db_password` 供給 + `alb_internal = true` 設定後 `terraform plan` 実行
  - 注: VPC Origin は ALB SG が Service-SG を source 参照必須（CIDR では 504）
  - 根拠: `docs/infra/P2_TERRAFORM_PLAN_RUNBOOK.md`
  - **実施者: ユーザー**（terraform apply は本番影響・破壊的操作）

- [ ] **N+1 クエリ解消 — LIFF 日付一覧 capacity チェック**
  - 対象: `backend/internal/service/liff_service_availability.go:62,73`
  - 内容: 日付一覧取得時に capacity を 1 件ずつクエリ → IN 句バッチ取得で改善
  - 根拠: 同ファイル内 TODO コメント実在

- [ ] **STG 費用最適化プラン実施**
  - 根拠: `docs/tasks/open/STG-COST-OPTIMIZATION-PLAN-2026-06-01.md`

---

## PO 決裁待ち（実装着手禁止）

| 内容 | 決裁待ち事項 | 根拠 |
|------|------------|------|
| ADR-003 論点1: payment_methods DB TRIGGER 検討 | 独立 Issue 要起票後 PO 判断 | `docs/adr/003-*.md` Status=Decided、論点1のみ保留 |
| lab_import Frontend UI | Phase 0 の操作 UX 仕様（操作者・画面・入力形式） | migration コメント + service Phase 0 定義 |

---

## 完了済み（2026-06-25〜26 で CLOSED 確認）

| # | 内容 |
|---|------|
| #193 | CI: actionlint ジョブ追加 → CLOSED |
| #194 | CI: カバレッジ計測ポリシー + 除外設定 → CLOSED |
| #195 | CI: setup-node@v4 → @v6 統一 → CLOSED |
| #196 | repository 層 clinic_id 隔離テスト補強 → CLOSED |
| #185 | ADR-003 payment_method 残論点 全論点確定 → CLOSED（論点1は保留 Issue 要起票） |
| #197 | payment_methods.system_key 列追加 → CLOSED |
| #198 | representativeMethod bank_transfer 分岐修正 → CLOSED |
| #159 | カルテレポート 麻酔処置・手術処置 + is_surgery フラグ → CLOSED |
| #160 | 健康診断カテゴリー化 既存 Examination で充足確認 → CLOSED |
| #179 | 月次集計レポート（#191/#192/#190 全 CLOSED） → CLOSED |
| #190 | 帳票レイアウト設定 全層実装 → CLOSED |
| #180 | CPM セグメント別人数・一覧 → CLOSED |
| #178 | カルテページ飼主コハビタントペット → CLOSED |

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

## 証跡サマリー（検査対象）

| 検査対象 | 結果 |
|---------|------|
| GitHub Issues（open） | #123 のみ OPEN（2026-06-26 確認） |
| GitHub Issues（closed 直近 30 件） | #159/#160/#178-#198 全 CLOSED 確認 |
| PR #186 CI ステータス | 全 SUCCESS（actionlint 含む、Vercel + qlty 含む） |
| `git status --short` | `M todo.md` のみ（lab_import 5 件は untracked のまま） |
| BE 規約チェック（lab_import） | P4/P8/P11/P16 修正済み。gofmt/scoped test は Docker 停止で未実行 |
| FE 規約チェック | lab_import 関連 FE ファイルなし → 対象外（根拠: git status 変更なし・FE lab_import UI 未着手） |
| migration 規約チェック | `005_add_lab_import_tables.sql` 適合（RESTRICT・clinic_id・インデックス・命名） |
| model 規約チェック | `lab_import.go` 適合（PHI コメント・型定義・TableName 実装） |
| `backend/internal/handler/` | lab_import ハンドラ不在確認 |
| `backend/internal/router/` | lab_import 配線なし確認（`grep LabImport` 0 件） |
| `backend/internal/service/lab_import_service_test.go` | 10+ テストケース実装済み（untracked）確認 |
| `docs/adr/003-*.md` | Status=Decided（論点1のみ保留） |
| `.github/workflows/stg-smoke.yml:5-9` | login/CRUD smoke 無効（secrets 未設定コメント実在） |
| `liff_service_availability.go:62,73` | N+1 TODO コメント実在確認 |
