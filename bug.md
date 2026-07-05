# クローズ済み Issue 仕様適合監査レポート → 未対応 backlog

> **本書の役割**: 2026-07-02 に実施したクローズ済み Issue 45件の仕様適合監査レポートを土台に、対応済み項目を随時削除して**未対応・残タスクのみの backlog**として運用している。以下の「調査日／対象／方法／判定サマリ」は監査実施時点の統計情報（履歴として保持）であり、実際に対応すべき項目は「対応状況」以降の各セクションのみを参照すること。

- **調査日**: 2026-07-02
- **対象**: 直近2週間（2026-06-18〜2026-07-02）にクローズされた GitHub Issue 45件（Search API と全件ローカルフィルタの2経路で突合し取りこぼしゼロを確認、全件 COMPLETED クローズ）
- **方法**: 各 Issue の本文・コメント・クローズ根拠を取得し、HEAD (main) の実コードと静的突合。CRITICAL/HIGH 全件＋主要 MEDIUM（M-4, M-6, M-8, M-9 および #181/#182/#188 の裏取り）は本体セッションで実コードを再検証済み。テスト実行・DB照会は未実施（静的検証のみ）。

## 判定サマリ（監査時点の統計・履歴参照用）

| 判定 | 件数 | 内訳 |
|---|---|---|
| MATCH（仕様充足） | 24 | #124 #125 #126 #127 #128 #151 #152 #153 #158 #159 #161 #179 #180 #182 #183 #184 #187 #188 #190 #191 #192 #198 #123 #193 |
| PARTIAL(一部未達) | 7 | #150 #155 #189 #194 #195 #196 #197 |
| MISMATCH(不一致) | 1 | #154 |
| CLOSED-AS-DECISION(実装なし・設計判断/リスク受容) | 13 | #156 #160 #178 #181 #185 #212 #89 #91 #96 #97 #98 #99 #109 |

## 対応状況（2026-07-05 更新 — 対応済み項目は本書から削除済み）

対応済み・削除済み: H-1/H-2 (`a45da439`) / H-3 (`4774666a`) / H-4 (`a620bdfc`) / H-6 (`aa9c0a5d`) / M-1/M-2 (`9d3df80c`) / M-8 (`5b0ac22d`) / M-9 (`ba8cecea`) / M-12 (`3f836cda`)、C-2（park 原本6 Issue #89 #91 #97 #98 #99 #109 の再オープンで代替解決）、仕様未充足 Issue 11件の再オープン、M-1 最終仕様化／M-3 越日EMG (`9ab95845`)／M-5 #178 保留クローズ／M-6 孤児API削除／M-7 税率 exact-match／M-10 `.gitignore` 明文化 (`830ee8e9`)。

**本セッション（2026-07-05）で追加対応・確認済み（コミット未実施・push 未実施）**:
- **B-1 完全体**: `medical_record_service_test.go` の `FindAll` mock を `repository.MedicalRecordListFilters` 受け取りに更新し、影響を受けた全 service テストファイルのモックも追随修正。`go test ./internal/service/...` `./internal/handler/...` `./internal/repository/...` 全 green。
- **B-1 follow-up 列ソート**: `sort`/`order` query（許可4キー: date/owner_name/pet_name/status）を handler で検証・repository で動的 `ORDER BY`（JOIN含む）化。FE `MedicalRecords.tsx` の列ヘッダクリックで server ソート、URL 状態（`useSearchParams`）に反映。
- **B-1 follow-up E2E**: `frontend/e2e/medical-records-pagination-sort.spec.ts` 新規追加（ページネーション page=2 遷移／列ソート URL 状態／ステータスフィルタ適用）。Docker Playwright で3件 PASS 確認済み。既存 `clinical-flows.spec.ts` 等の回帰なし確認済み。
- **M-10 本番デモアカウント非表示**: `vite.config.ts` の `define` で `__VERCEL_ENV__`（Vercel の `VERCEL_ENV` を build-time 定数化）を注入し、`LoginForm.tsx` の `SHOW_DEMO` をインライン式化（`__VERCEL_ENV__ !== "production" && (...)`）してビルド時に tree-shake。`VERCEL_ENV=production` での scoped production build で `grep` によりデモアカウント文字列（メール等）が dist に含まれないことを確認済み。ロジックの単体テストは `show-demo-accounts.test.ts`（`computeShowDemoAccounts` として分離）。
- **H-7 カバレッジ**: scoped `go test -cover` で `internal/handler` 94.6%／`internal/service` 93.3%／`internal/repository` 73.1%（medical_record_repository の CRUD＋clinic_id 隔離テスト追加込み）を確認。`backend/.coverage-baseline`（89.9%, 2026-07-03 arm）／`frontend/.coverage-baseline`（43.78%, 2026-07-04 arm）は既に実測値で armed 済み、`docs/coverage-policy.md` のプレースホルダも解消済みであることを確認（追加更新不要）。
- **H-5（コード側）**: `.github/workflows/backend-deploy.yml` を `SSM_SECRET_PARAM_MAP` 経由の `valueFrom`（SSM SecureString）方式に変更し、`DB_PASSWORD`/`JWT_SECRET`/`INTEGRATION_ENCRYPTION_KEY` を `environment`（平文）から除外。`infra/terraform/modules/ecs/main.tf` に ECS task execution role 向け `ssm:GetParameters`/`kms:Decrypt` の IAM ポリシーを追加。**AWS への SSM 実登録・Terraform apply・デプロイはユーザー承認待ち**（手順は下記ランブック参照）。
- **C-1（コード側）**: CI に `secret-scan`（gitleaks/gitleaks-action@v3, push/PR 全件・fetch-depth 0）ジョブを追加。既知の非シークレット誤検知（`backend/docs/api-examples.md` 等4ファイル）用に `.gitleaks.toml` allowlist を追加（`.env.staging` は意図的に allowlist 対象外のまま維持）。**ローテーション・`.env.staging` 追跡解除・Issue #97 実値削除はユーザー承認待ち**（手順は下記ランブック参照）。
- **M-4/M-11/C-1 外部操作ランブック**: [`docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md`](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) を新規作成。実行順序・exact コマンド（プレースホルダ値）・完了記録テンプレートを含む。

**残っているのは「外部操作（AWS/GitHub への書き込み）」と「PO 判断待ち（#189）」のみ**。詳細は下記 BLOCKED セクション参照。**全コミット push 未実施**。

### 残項目（すべて外部操作 or PO 判断待ち — コードでは完結しない）

| 残項目 | 状態 | 手順書 |
|---|---|---|
| C-1 ローテーション／#97 実値削除／.env.staging 追跡解除 | ⏳ **ユーザー実施要**（AWS RDS/JWT/LINE + `gh issue edit`） | [外部操作ランブック §1, §3](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| H-5 SSM 実登録＋Terraform apply＋デプロイ検証 | ⏳ **ユーザー実施要**（コードは実装済み） | [同 §2](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| M-4 STG db_reset=true デプロイ | ⏳ **ユーザー承認要**（破壊的操作。Checkup系スキーマは統合済み `001_init.sql` に含まれるため反映には db_reset が必須） | [同 §4](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| M-11 GitHub Secrets 登録 | ⏳ **ユーザー実施要**（`CI_TEST_EMAIL`/`CI_TEST_PASSWORD`、workflow 側は参照実装済み） | [同 §5](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) |
| #189 締め後訂正の可視化バッジ | 🚫 **BLOCKED（PO 未決）** — デフォルト非実装。ユーザーが明示的に「バッジ要」と指示した場合のみ実装 | 下記「保留 (PO Decision Pending)」参照 |

---

## 🚫 保留 (PO Decision Pending)

### #189. 締め後訂正の可視化バッジ

**事象**: レジ締め後に金額が訂正された場合、締め詳細画面・月次集計レポートにその事実を示す
UI 表示（バッジ等）が存在しない。FE 側の警告表示自体は既に対応済み（`0dad744e`）だが、
「締め後に訂正がある」ことを事後的に画面上で可視化する要否は PO 未決のまま。

**方針**: **デフォルトでは実装しない**。要件（①要件を疑う）を満たさないまま UI を追加するのは
`docs/PRODUCT_PHILOSOPHY.md` の 5 ステップ原則に反するため、PO が「バッジ要」と明示するまで
着手しない。

**要実装となった場合の対応イメージ**（PO 承認後の参考のみ・未着手）:
- 対象: 締め詳細画面 + 月次集計レポート
- 表示: 該当レコードに「締め後訂正あり」の最小バッジ（アイコン + tooltip 程度）
- 出典: `audit_logs` の締め後書き込みイベントを検出

---

## 🔴 CRITICAL

### C-1. [#97/#89] 平文シークレットが現在も露出継続 — ローテーション未実施（コード側の再発防止は対応済み）

**現状（2026-07-05 更新）**: `.env.staging` は依然 git 追跡されたまま平文の `DB_PASSWORD` /
`JWT_SECRET` / `INTEGRATION_ENCRYPTION_KEY` を含む。Issue #97 本文にも実値記載が残る。
ローテーションは未実施。**コード側の受け皿（H-5 の SSM valueFrom 化・CI gitleaks scan）は
本セッションで実装済み**のため、あとは実際のローテーション・登録・追跡解除を行うだけの状態。

**残タスク**: [外部操作ランブック §1（ローテーション）・§3（追跡解除・#97編集）](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) の exact コマンドに従って実施。**ユーザー承認必須**（AWS RDS 変更・LINE 再発行・`gh issue edit` はいずれも外部書き込み）。

**検証方法**: ランブック §1-3 の確認コマンド参照。最終的に `git ls-files | grep -E '\.env'` で `.env.staging` が消えること。

---

## 🟠 HIGH

### H-5. [#99] backend-deploy.yml が secrets を ECS environment に平文展開 — コード対応済み・実登録待ち

**現状（2026-07-05 更新）**: `.github/workflows/backend-deploy.yml` を `SSM_SECRET_PARAM_MAP`
経由の `valueFrom`（SSM SecureString）方式へ変更済み。`DB_PASSWORD`/`JWT_SECRET`/
`INTEGRATION_ENCRYPTION_KEY` は `environment`（平文）から除外され、`secrets[].valueFrom` で
SSM ARN を参照する形になっている。`infra/terraform/modules/ecs/main.tf` に ECS task
execution role 向け `ssm:GetParameters`/`kms:Decrypt` の IAM ポリシーを追加済み。

**残タスク**: [外部操作ランブック §2](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) — Terraform apply（IAM反映）→ SSM へ新値登録（C-1 ローテーション後）→ デプロイ → `aws ecs describe-task-definition` で平文露出が無いことを確認。**ユーザー承認必須**。

### H-7. [#212] カバレッジ拡充 — 完了確認済み

**現状（2026-07-05 確認）**: scoped `go test -cover` で `internal/handler` 94.6%／
`internal/service` 93.3%／`internal/repository` 73.1%。`backend/.coverage-baseline`
（89.9%, 2026-07-03 arm）・`frontend/.coverage-baseline`（43.78%, 2026-07-04 arm）は
実測値で armed 済み、`docs/coverage-policy.md` のプレースホルダも解消済み。**追加対応不要**。

---

## 🟡 MEDIUM

### M-4. [#160 vs #211] 健診系統 — STG db_reset 待ちのみ

**現状**: Checkup 系（#211, ADR-004）の実装・スキーマは統合済み `001_init.sql` に含まれている
（旧稿で言及していた「migration 010/011」という個別番号は migration 統合前の呼称で、現在は
実体が存在しない — #179 と同種の文書ドリフト）。STG の DB が統合前の状態のままであれば、
反映には **db_reset=true が必須**（個別 ALTER の追い migration は作成しない方針のため）。

**残タスク**: [外部操作ランブック §4](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)。**破壊的操作のためユーザー承認必須**（STG データ全損）。

**検証方法**: fresh DB apply 後、健診記録の入力導線が Checkup 系1系統のみであることを UI で確認。

### M-10. [#91] frontend/.env.production — 完了確認済み

**現状（2026-07-05 更新）**: `vite.config.ts` の `define` で `__VERCEL_ENV__` を build-time
定数として注入し、`LoginForm.tsx` の `SHOW_DEMO` をインライン式（
`__VERCEL_ENV__ !== "production" && (import.meta.env.DEV || ... === "true")`）に変更。
Vite/esbuild が本番ビルド時に式全体を `false` へ定数畳み込みし、`DEMO_ACCOUNTS` 配列ごと
tree-shake される。`VERCEL_ENV=production` での scoped production build を実施し、
dist にデモアカウントのメールアドレス等が含まれないことを `grep` で確認済み。**追加対応不要**。

### M-11. [#109] CI テスト認証情報 — Secrets 登録のみ残タスク

**現状**: workflow 側は `secrets.CI_TEST_EMAIL` / `secrets.CI_TEST_PASSWORD` を参照する実装のまま変更なし。

**残タスク**: [外部操作ランブック §5](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) — `gh secret set` で登録。**ユーザー実施要**。

---

## 🟢 LOW（記録のみ・対応は任意）

| Issue | 内容 | 状態 / 推奨対応 |
|---|---|---|
| #194 | 受入基準「現行ベースライン数値の記録」未達だった `docs/coverage-policy.md:49` のプレースホルダ | ✅ **解消済み**（H-7 確認と同一作業。プレースホルダは既に実測値で置換済み） |
| #196 | pet CRUD テストが R/D 止まり（Create/Update なし）。medical_record_image_repository 未カバーだった懸念 | ✅ **解消済み**（`pet_repository_test.go` に `TestPetRepository_Create`/`_Update`、`pet_write_medimage_clinic_isolation_test.go` に clinic_id 隔離テストが存在することを確認。`medical_record_image_repository` も `TestMedicalRecordImageRepository_FindByID_ClinicIsolation`/`_Delete_ClinicIsolation` が green） |
| #182 | クローズコメント本文「本Issueはクローズしない（OPEN維持）」と実際の CLOSED 状態が矛盾 | 未投稿・承認待ち。投稿コマンド: `gh issue comment 182 --body "子 #188/#189 完了に伴いクローズ"` |
| #179 | クローズコメントの migration 番号（010/008）が統合後の実体（001_init.sql）と乖離 | 文書ドリフトのみ。未投稿・承認待ち。投稿コマンド: `gh issue comment 179 --body "migration は 001_init.sql へ統合済み。クローズコメント中の 010/008 表記は統合前の呼称（実害なし・記録目的のコメント）"` |
| #184 | 実ブラウザ印刷の視覚検証は JSDOM 制約で未実施 | 未対応。`frontend/e2e/medical-records-pagination-sort.spec.ts` で Playwright 基盤の実利用実績ができたため、印刷ビューのビジュアルリグレッションも同基盤で追加可能（任意 follow-up） |

---

## 本監査の対象外とした既知の横断事項

- **監査ログ書込がコードベース全体で tx 外の best-effort**（`auditRepository.Create` が dbOrTx 非使用）: #189 等の個別 Issue の受入条件とは独立した横断的既知事項のため、各 Issue の判定には含めていない。#211 系の follow-up 作業（audit-tx 原子化: refund/checkup_field_results は対応済み、残りは横断展開中）で別途追跡されている。
- **検証手法の限界**: 静的コードリーディングのみ。テスト実行・DB照会・実ブラウザ検証は未実施。「テストが存在する」ことは確認済みだが「テストが通る」ことは未確認。#212 のカバレッジ実数も実測していない（2026-07-05 セッションで scoped `go test -cover` により実測値を確認済み。上記 H-7 参照）。

---

## 既知の実装課題（Open Bugs）

**B-1（カルテ一覧 server-side pagination）は 2026-07-05 セッションで完全に解消済み**（本体・
mock signature 修正・列ソート server 化・E2E 追加まで完了、上記「対応状況」参照）。
現時点で残っている実装課題はない。
