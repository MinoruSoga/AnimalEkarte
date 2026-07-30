# Screens-spec drift — remaining work ledger

本台帳は screens-spec 実装ドリフト監査 + fix-all（Mode 3 COMPLETE）後に残ったオープン項目のみを列挙する。
Pack A–F の doc honesty 出荷済み分は再掲しない（再スキャンで回帰が確認されたもののみ再オープン）。

加えて **SCENARIOS-ALIGN-REMEDIATE-WAVE1**（scenarios doc↔code 監査・修正 wave + Mode 3）後の residual を同一ファイルに併記する（索引 ID は `SCEN-*`、TASK は 009 以降）。screens-spec 節（R1–R8 / TASK-001–008）は保持する。

さらに **docs/spec top-level audit→fix→Mode3 連鎖**（SPEC-TOP residual）の未カバー follow-up を 2026-07-31 に併記する（索引 ID は `SPEC-TOP-*`、TASK は **018 以降**）。G1–G12 機能クローズと ERD/spec **115** 同期は **再実装 TASK にしない**（coverage matrix / ARCH-DONE のみ）。

## 索引 / サマリー

| Inv | 内容 | 処置 |
|-----|------|------|
| R1 | 入院 create + nested treatment-plan POSTs が非原子的 | **TASK-001-BE / TASK-001-FE** |
| R2 | 編集モード治療プランが read-only のまま | **TASK-002**（要PO判断） |
| R3 | フォーム一括割引が非永続 | **TASK-003**（要PO判断） |
| R4 | screens-drift 意図変更のコミット隔離 | **TASK-004**（open・手順。現状 intentional は ledger 用 `todo.md` のみ可。screens 本体 land は dirty 再計測時） |
| R5 | コミット前の closed pack 回帰ゲート | **TASK-005**（scoped gates 実行済み・land 前に再実行可） |
| R6 | マルチエージェント共有 tree thrash | **ops-only**（下記）— 製品 TASK なし。既存 `git-worktree-safety` を正本とし worktree 隔離を徹底 |
| R7 | Writer が diff なしで成功宣言した harness 問題 | **ops-only**（下記）— 製品 TASK なし。任意で CLAUDE/Agents に「empty-diff ≠ success」1行を後続で検討 |
| R8-D | `40-identity-links.md` 欠落 + README 索引不全 | **TASK-006 done**（docs 作成済・tree にファイル PRESENT） |
| R8-B03 | `03-owners-list.md` が client-side `/owners` のまま | **TASK-007 done** |
| R8-board | ボードカード open が `edit` 必須 vs 仕様未記載 | **TASK-008 done** |
| R8-README | README「全40」 | **TASK-006 done に包含** |
| 一時帰宅 | ドキュメント honesty 済み / BE enum 一致 | **WONTFILE** — mismatch ではない |
| SCEN-SEED-001 | 003_demo clinical CSV がヘッダのみ | **TASK-009**（High / seed rebuild 方針・適用は USER） |
| SCEN-BROWSER-001 | scenarios 内【要実測】backlog | **TASK-010**（Med / browser-test レーン一括） |
| SCEN-S11-COPY-001 | S11 手順2 文言 residual | **TASK-011**（Low / docs 微修正） |
| SCEN-PROD-ESTIMATE-001 | estimate unlock vs 26§2.1 | **TASK-012**（要PO判断） |
| SCEN-PROD-CLOSING-001 | closing reverse boundaries | **TASK-013**（要PO判断） |
| SCEN-PROD-PAYMENT-KEY-001 | payment system_key | **TASK-014**（要PO判断） |
| SCEN-AUDIT-MED-001 | 監査 MEDIUM 残渣の再スキャン | **TASK-015**（Low / 任意 gate） |
| SCEN-OPS-CLAIM-001 | `claim/*` 解放（旧 `claim/SCENARIOS-ALIGN-REMEDIATE-WAVE1` 含む） | **ops-only**（下記）— USER only。2026-07-31 再計測: live `claim/*` **空**（解放済み or 未作成）。残存時のみ `git branch -D` |
| SCEN-OPS-COMMIT-001 | wave が mixed commit に同梱済み | **ops-only**（下記）— history hygiene 説明のみ。rewrite しない |
| SCEN-OPS-TREE-001 | 共有 tree に hospitalization/identity-links 等 concurrent WIP | **ops-only**（下記）— R6 と統合。worktree 隔離徹底 |
| SCR-ID-LINKS-WIP | 生成時点 snapshot で `40-identity-links.md` D + screens README M が dirty だった件 | **ops-only / closed thrash**（下記）— identity wipe 再オープンの製品 TASK は起票しない。tree: file PRESENT + README No.40 + TASK-006 done 維持 |
| ARCH-DONE | architecture docs Mode 3 同期 + 2026-07-31 追記（ERD `TABLE_COUNT=115` = 001:110 + identity 4 + `medical_record_image_upload_quota` 1、`check-docs-symbol-drift` 3a 正本、FE `RESOURCE_LABELS` 35=`AllResources`、narrative）。115 同期は `07e2fb945`（`docs/architecture/erd.md` + `docs/spec/specification.md` のみ path-scoped） | **closed / done** — 再実装 TASK にしない。Mode 3 本体は mixed 履歴（例: `1d8d2e59d`）に混入していたが、**115 追記は path-scoped で完了**。索引フッタ 110 残は ARCH-R4 / TASK-017（目標 **115**） |
| ARCH-R1 | ADR-003 References の歴史的 `internal/service/...` パス | **TASK-016**（Low / docs） |
| ARCH-R2 | architecture fix-all 第1回: empty-diff / worktree 未 merge で COMPLETE 宣言 | **ops-only**（R7 拡張）— 製品 TASK なし。empty-diff COMPLETE 拡張は **COMPLETE**（規律継続・製品 TASK なし） |
| ARCH-R3 | architecture が mixed preserve commit に混入。screens-drift land 時の foreign 定義 | **ops note** + **TASK-004 文言 amend**。最新 115 同期 `07e2fb945` は path-scoped isolation の好例（ERD+spec のみ） |
| ARCH-R4 | docs-symbol-drift が README/screens 等フッタを 3a 対象外にした経緯。110→**115** 手動同期の残と gate 再対象化の方針 | **TASK-017**（Low / 任意）≡ **SPEC-TOP-FOOTER-115** |
| POST-PULL | migrations `005` / `006` pull 後のローカル適用 | **ops-only**（下記）≡ **SPEC-TOP-MIGRATE-006** — USER が `make migrate`。エージェントは auto-apply しない |
| SPEC-TOP-G1-G12 | docs/spec top-level G1–G12 機能クローズ（reservation honesty / cash-register / customer-aggregation / IdentityLinks PageLayout / ExamTypeFieldsEditor / e2e+route inventory 84 / design-audit / symbol-drift@115 等） | **already done** — 再 open しない（coverage matrix のみ） |
| SPEC-TOP-AVAILABLE-STAFFS | `available-staffs` API | **WONTFILE** — `docs/spec/reservation-to-record-flow.md` で「未実装・導入しない」。PO 再オープン時のみ再評価 |
| SPEC-TOP-CAP-SOT-DOC | capabilities write SoT 叙述（dual surface narrative の doc 固定） | **already done**（doc）。残 write 面は **TASK-021** |
| SPEC-TOP-FOOTER-115 | 非 gate フッタ（docs README / screens README / DEPLOYMENT_CHECKLIST）が 110 のまま | **TASK-017**（目標 115。旧 114 文言は破棄） |
| SPEC-TOP-SCREENS-AUDIT | `docs/spec/screens/**` 全量 vs code 再監査 | **TASK-018**（Med / 任意 full pass） |
| SPEC-TOP-LINE-AUDIT | `docs/spec/line/**` 全量 vs code 再監査 | **TASK-019**（Med / 任意 full pass） |
| SPEC-TOP-E2E-RUNTIME-84 | inventory 84 更新後の ui-design-compliance Playwright 再 runtime | **TASK-020**（Low / 任意 ops） |
| SPEC-TOP-CAPABILITIES-CRUD | staff 対応可能種別: capabilities write SoT と exclusion 面の残 dual surface | **TASK-021**（Med・要PO判断） |
| SPEC-TOP-MIGRATE-006 | migration 006 未適用環境 | **POST-PULL / ops-only**（USER `make migrate`） |
| SPEC-TOP-CLAIM-RELEASE | parallel claim/* 解放 | **SCEN-OPS-CLAIM-001 / ops-only**（USER only） |

### Ops-only notes（製品コード TASK にしない）

- **R6 / SCEN-OPS-TREE-001**: 並行 Grok/Claude は共有 working tree 禁止。`.claude/rules/git-worktree-safety.md` 準拠で worktree 分離。hospitalization / identity-links 等の concurrent WIP がある共有 tree では 1 編集セッションのみ。repo への追記は任意。
- **R7**: harness が empty-diff 成功を許す問題。prompt-craft / writer ゲート側。本 repo の機能 TASK ではない。
- **ARCH-R2（R7 拡張）**: architecture docs fix-all 第1回で、マルチエージェントが empty-diff または worktree 未 merge のまま COMPLETE を宣言した事例あり。受け入れ判定は **`git diff -- <path>` / `git status` で実 diff を必須**とし、宣言文面だけで PASS にしない（harness/ops 側の規律。製品コード TASK にしない）。本拡張は ops 規律として **COMPLETE**（継続適用・追加製品 TASK なし）。
- **ARCH-R3**: architecture Mode 3 本体は `1d8d2e59d chore: preserve current main work` 等の **mixed preserve commit** に混入し得る。一方 **115 追記は `07e2fb945 docs: sync ERD and specification table count to 115`**（`docs/architecture/erd.md` / `docs/spec/specification.md` のみ）で path-scoped isolation 済み — この形を screens-drift land の参考にする。将来 land（TASK-004）時は **その時点の working tree / `git diff` を再計測**して intentional / foreign を定義し直す。architecture を「未コミット foreign WIP」前提に固定しない。
- **SCR-ID-LINKS-WIP**（黙殺しない）: 台帳生成時点の snapshot では uncommitted `D docs/spec/screens/40-identity-links.md` と `M docs/spec/screens/README.md`（41→40 索引）が観測された。**2026-07-31 再計測**: `40-identity-links.md` **PRESENT**（tracked 例: `6f83aae9f`）、README No.40 行あり、製品ルート `/identity-links` と整合。**既定 disposition は restore**だが既に揃っているため **identity wipe 再オープン用の製品 TASK は起票しない**（TASK-006 done の再オープン禁止。TASK-018 は screens 全量監査用に別用途）。再発時は restore（route 存続）を既定とし、feature 削除は PO 証拠がある場合のみ。
- **POST-PULL（migrations 005/006）≡ SPEC-TOP-MIGRATE-006**: `005_line_webhook_bot_user_id.sql`（カラムのみ）および `006_medical_record_image_upload_quota.sql`（`medical_record_image_upload_quota` 1 テーブル）を pull した開発者は、更新アプリ利用前に **USER が手動で `make migrate`**。エージェントは migration を auto-apply しない（Agents.md / CLAUDE.md）。
- **SCEN-OPS-CLAIM-001 ≡ SPEC-TOP-CLAIM-RELEASE**: claim 解放は **USER only**。**2026-07-31 再計測** `git branch --list 'claim/*'` は **空**（解放済み or この clone に未作成）。残存が再出現した場合のみ merge 後 `git branch -D claim/<TASK-ID>`。エージェントは claim を削除しない（Agents.md packet claim プロトコル）。
- **SCEN-OPS-COMMIT-001**: scenarios wave の product/docs 成果は mixed commit（例: `1d8d2e59d chore: preserve current main work`）に unrelated パスと共に同梱済み。将来の history hygiene / 説明用メモ。**history rewrite / force-push はしない**。
- **SPEC-TOP closed（再 file 禁止）**: G1–G12 機能クローズ、symbol-drift@115、route/e2e inventory 84 静的ゲート、design-audit 緑、available-staffs「導入しない」、capabilities SoT の **doc 叙述固定**は already done / WONTFILE。再 open は回帰証拠がある場合のみ。

### 推奨実装順

1. TASK-004: intentional dirty が **screens-drift land 用**に再出現したとき path-scoped stage → 再 TASK-005 → ユーザー承認後 commit（**foreign 定義は ARCH-R3 / その時点の diff**）。現状 `todo.md` のみ dirty は ledger unit として許容（screens land 対象外）  
2. ~~TASK-006 / 007 / 008~~（docs residual 完了・identity-links は commit 済み・PRESENT）  
3. ~~ARCH-DONE / SPEC-TOP-G1-G12~~（architecture Mode 3 + **115** + top-level fix wave 完了・再開禁止。`07e2fb945`）  
4. TASK-001-BE → TASK-001-FE（原子性、臨床 integrity）  
5. TASK-002 / TASK-003（PO 判断待ち — 決定を発明しない）  
6. **ops**: SCEN-OPS-TREE-001（隔離）を維持 → claim 残存時のみ SCEN-OPS-CLAIM-001（USER）。POST-PULL / SPEC-TOP-MIGRATE-006 は USER が `make migrate`  
7. **TASK-009** seed 方針（High; 適用は USER / エージェントは migrate・seed 自動適用しない）  
8. **TASK-012 / TASK-013**（PO 決裁・High）→ 決定後に実装 follow-up  
9. **TASK-010** 【要実測】browser-test バックログ（Med; seed 後が理想）  
10. **TASK-014** payment system_key（PO・Med）  
11. **TASK-011** S11 文言（Low・即着手可）  
12. **TASK-015** 監査 MEDIUM 再突合（Low・任意）  
13. **TASK-016 / TASK-017** architecture 残債（Low・任意順; TASK-017 = SPEC-TOP-FOOTER-115、目標 **115**。ARCH-DONE の再オープンではない）  
14. **TASK-018 / TASK-019**（任意 full audit: screens/** / line/**）  
15. **TASK-021** capabilities dual surface（PO）→ 決定後 follow-up  
16. **TASK-020** e2e runtime 84 再確認（Low・任意; inventory 静的 84 は完了済み）

---

## 個別タスク詳細

### TASK-001-BE: 入院作成と治療プランの原子的永続化（High）

- **問題**: 入院親レコード POST 成功後に治療プランを逐次 POST しており、途中失敗時に親（および先行プラン）が残る。臨床グラフの部分成功は product philosophy の fail-closed に反する。
- **根拠**: FE `use-hospitalization-form.ts` が `createHospitalization` 後に `createTreatmentPlanForHospitalization` を loop（nested POST; 単一 DB TX ではない）。BE は親 `POST /v1/hospitalizations` と `POST /v1/hospitalizations/:id/treatment-plans` が別 handler（`backend/internal/medicalrecord/`）。Create リクエスト型に nested plans なし。影響: DB/BE。
- **修正方針**: 案A（推奨）親 create と nested plans を同一 TX の BE 契約に載せる（nested body または dedicated composite endpoint）。案B 補償削除 API（親削除 or plans ロールバック）を FE 失敗時に呼ぶ。案A を優先。clinic scope・audit・権限は既存 hospitalization create/edit と同等。`make codegen` が必要なら同タスクに含める。
- **受け入れ条件**: ① プラン N 件中 k 件目で強制失敗させても hospitalization 行が残らない（TX rollback）または補償後に残らない。② 全成功時は親+全プランが同一 clinic で読める。③ スコープ: `docker compose exec backend go test ./internal/medicalrecord/...`（該当 package）。
- **状態**: P1/High。TASK-001-FE 依存元。仕様確認: 案A/B の選択は実装時に BE 契約最小変更を優先（nested create が既存 graph と整合）。未解決: 公開 API shape の最終決定。

### TASK-001-FE: 入院フォーム保存を原子的契約へ接続（High）

- **問題**: 現状 FE が親作成→プラン逐次 POST を直列実行し、途中失敗を toast エラーにしても親が残り得る。
- **根拠**: `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts` create 分岐; `api/treatment-plans-write.ts` POST only; 仕様 honesty は `docs/spec/screens/09-hospitalization-form.md`。影響: FE（TASK-001-BE 契約後）。
- **修正方針**: TASK-001-BE の契約に合わせて create path を 1 回の成功境界に寄せる。案A なら composite/nested body 1 リクエスト。案B なら失敗時補償呼び出し + 失敗 UX を「部分作成済み」ではなく復旧済みに揃える。テスト: `use-hospitalization-form.test.ts` / `treatment-plans-write.test.ts` を更新。
- **受け入れ条件**: ① プラン POST 失敗シミュレーションで orphan 親が残らない（モック+契約）。② 成功時は登録 toast と一覧/詳細でプラン件数一致。③ `docker compose exec frontend pnpm test:run -- src/features/hospitalization`。
- **状態**: P1/High。**TASK-001-BE 完了後**着手。仕様確認ログ: BE 契約確定を待つ。

### TASK-002: 編集モード治療プランの更新アンロック（Medium・要PO判断）

- **問題**: 編集 UI は治療プラン参照のみ。BE には PATCH/DELETE があるが FE クライアント未配線。現場で明細変更が入院フォームからできない。
- **根拠**: `HospitalizationForm.tsx` `readOnly={isEdit}` と説明文; edit save は親のみ（`use-hospitalization-form.ts`）; BE `routes.go` に PATCH/DELETE treatment-plans あり; FE write は create POST のみ。影響: FE（必要なら BE 権限制御の再確認）。
- **修正方針**: **要PO判断** 案A: フォーム edit で PATCH/DELETE + GET 同期（`treatment-plans-write` 拡張）。案B: 永続 RO を受け入れ、詳細/ケア導線のみと仕様に固定（現状 honesty 維持）。案C: 別画面を唯一の write owner と明記。実装は PO 決定後。
- **受け入れ条件**: 案A なら ① edit でプラン追加/変更/削除が保存され再読込で一致 ② view-only 権限では操作不可。案B なら ① 仕様と UI が「意図的 RO」と検証可能に一致し追加実装なし。スコープ FE テスト + 該当 BE テスト。
- **状態**: Medium。要PO判断（unlock vs WONTFIX honesty）。TASK-001 と独立だが同一 feature。

### TASK-003: 入院フォーム一括割引の永続化可否（Medium・要PO判断）

- **問題**: 一括割引（%/円）は常に表示専用。操作可能に見えて保存されないリスクは honesty 文言で緩和済みだが、製品として永続化するか未決。
- **根拠**: `HospitalizationForm.tsx` CostSummary `readOnly` + 非保存説明; model/types に discount フィールドなし; `09-hospitalization-form.md` L36 相当 honesty。影響: 案により DB/BE/FE。
- **修正方針**: **要PO判断** 案A: BE カラム + API + FE 永続（請求整合を会計とすり合わせ）。案B: 現状の表示専用を受け入れ済みとし追加実装しない（docs を accepted decision に格上げ）。会計二重管理禁止（product philosophy）に触れないこと。
- **受け入れ条件**: 案A なら create/update 往復で割引値が一致し会計と矛盾しない契約。案B なら 受け入れ記録（本 TASK クローズ理由）と UI honesty が残る。PO 未決のまま実装開始しない。
- **状態**: Medium。要PO判断。未解決: 会計ドメインとの source of truth。

### TASK-004: screens-drift 意図変更セットのコミット隔離（Medium・ops）

- **問題**: screens-drift 意図変更と、それ以外のレーン変更を同一 commit に混ぜない。**foreign 定義は land 直前の `git status` / `git diff` が正本**（固定リストに依存しない）。
- **根拠（2026-07-31 amend / ARCH-R3 + re-measure）**: architecture docs は local main に commit 済み。ERD/specification の **115** 同期は path-scoped `07e2fb945`（`erd.md` + `specification.md` のみ — isolation の好例）。Mode 3 本体は mixed 例 `1d8d2e59d` にも混入し得る。**2026-07-31 再計測（ledger unit 後）**: screens / product intentional は **dirty なし**。許容 dirty は ledger 用 **`todo.md` のみ**（本 unit の allowlist）。`40-identity-links.md` は PRESENT/tracked（例: `6f83aae9f`）。過去の intentional 候補（`docs/spec/screens/*`、Pack C hospitalization form/write、`route-inventory.test.tsx`、`master-settings-index-model(.test).ts` 等）は **現状 dirty ではない**。foreign は **その時点で dirty な意図外パス**のみ。identity-links FE / ExamTypeFieldsEditor は ambiguous（既定 OUT）。**次の screens land 直前に必ず再計測**。
- **修正方針**: screens intentional dirty が無い間は screens stage no-op。dirty 再出現時のみ path-scoped `git add`（`git add -A` 禁止）。foreign は触らない・捨てない。`todo.md` 単独 commit は ledger 用に別途可（screens-drift land と混ぜない）。
- **受け入れ条件**: ① dirty がある land では `git diff --cached --name-only` が intentional 集合 ⊆ のみ。② foreign が staged に含まれない。③ commit 後も foreign WIP が残る（破棄しない）。④ screens intentional が無いときは「screens stage 対象なし」を Completion Report に明示してよい。
- **状態**: Medium/ops。**open（手順・再発用）**。screens 即時 stage 不要。commit はユーザー承認境界。architecture 未コミット前提のブロッカー文言は撤回。

### TASK-005: closed packs 回帰のコミット前検証ゲート（Medium・ops）

- **問題**: screens-drift land 前に Pack A–F 相当の doc/code 整合と inventory 84・hospitalization 保存パスが壊れていないことを機械確認するゲートが手順化されていない。
- **根拠**: 意図パスに hospitalization 保存・route inventory・screens docs が含まれる。symbol-drift スクリプトは foreign dirty の可能性あり（`scripts/check-docs-symbol-drift.*`）— ゲート実行時はスクリプト自体の foreign 変更に注意。
- **修正方針**: land 直前に（ユーザー/Docker）: `bash scripts/check-docs-symbol-drift.sh`（意図 docs が stage 済みの状態）; `docker compose exec frontend pnpm test:run -- src/features/hospitalization src/app/routes/route-inventory.test.tsx src/features/master/routes/master-settings-index-model.test.ts`。失敗時は TASK-004 の stage 集合を見直す。
- **受け入れ条件**: ① 上記コマンドが PASS。② inventory が 84 product pages を維持。③ hospitalization create/plan 関連 unit が PASS。
- **状態**: Medium/ops。**scoped symbol-drift + vitest を本セッションで実行**。land 直前に再実行推奨。full-suite は自動実行しない。

### TASK-006: identity-links 画面仕様の新設と screens README 索引更新（High）

- **問題**: 製品ルート `/identity-links` と inventory 84 がある一方、`docs/spec/screens/40-identity-links.md` が欠落。README は「全40」のまま identity-links 行なし。Pack D ドキュメント完了の再スキャンで回帰。
- **根拠**: （実装前）`40-identity-links.md` 不存在; README「全40」; `IdentityLinksPage` + `ResourceIdentityLinks`。
- **修正方針**: `IdentityLinksPage` / `identity-links-api` / routes を読んで仕様作成。README に No.40 行 + 「全41画面」（番号付き md ファイル数）と product leaf 84 の数え分けを明記。
- **受け入れ条件**: ① ファイル存在・権限/API 一致 ② README リンク ③ symbol-drift OK。
- **状態**: **done**（docs 作成・README 更新・symbol-drift PASS）。`40-identity-links.md` / README 索引は main に commit 済み（例: `6f83aae9f`）。SCR-ID-LINKS-WIP の一時 dirty は thrash として ops 記録済み — 本 TASK は再オープンしない。

### TASK-007: 飼主一覧仕様を GET /v1/pets サーバページネーションに合わせて修正（High）

- **問題**: `03-owners-list.md` が `GET /api/v1/owners` + クライアント側フィルタ、`useDeferredValue` を誤記載。
- **根拠**: `ownersLoader` → `GET /v1/pets` サーバ page/search 等。owners に `useDeferredValue` なし（300ms debounce → URL）。
- **修正方針**: API 表・§3.1・列ソート撤去をコードに合わせて書き換え。
- **受け入れ条件**: `/v1/pets` 正本・クライアント全件フィルタ主張なし・symbol-drift OK。
- **状態**: **done**（`03-owners-list.md` 更新済み）。

### TASK-008: 入院ボードのカード open が edit 必須であることの仕様正直化（Medium）

- **問題**: 仕様が「カード/行クリック→詳細」と一般化し、board の `canEdit` ゲートを欠く。
- **根拠**: `HospitalizationBoard` `canOpenCard`/`canDrag` が `canEdit`; リスト No リンクは view。
- **修正方針**: docs honesty で board open/DnD=edit、list No=view を明記（FE 変更なし）。
- **受け入れ条件**: 仕様が edit ゲートを明示し board/list を混線しない。
- **状態**: **done**（`07-hospitalization-list.md` 更新済み）。

### TASK-009: 003_demo clinical CSV ヘッダのみ — seed 再投入方針（High）

- **問題**: `003_demo` 系 clinical CSV がヘッダのみで、シナリオが想定する臨床データが seed 再投入後に揃わない可能性がある。本 wave はシナリオ前提側を修正済みだが、seed 実体の再投入は未実施。
- **根拠**: residual **SCEN-SEED-001**。scenarios 実行環境は seed 003_demo / STG 004_staging。本 wave はシナリオ前提修正済み。seed 再投入は別 unit。
- **修正方針**: (1) 対象 CSV と最小 fixture 行を文書化。(2) **migration-seed-safety** および seed 規約に従い設計（clinic_id・クロステナント衝突・checksum）。(3) エージェントは migration / seed 適用を自動実行しない。USER が `make migrate` 等を手動実行。(4) シナリオ前提との差分を最小データで埋める。紙/Excel 忠実複製はしない。
- **受け入れ条件**: ① 方針で定めた対象 clinical CSV がヘッダのみでなくなる。② S 系シナリオが seed 003_demo で前提検索条件を満たせる。③ 適用手順が docs または seed README に1箇所で辿れる。④ 適用は USER 手動。
- **状態**: open / High。依存なし（シナリオ前提は wave 済み）。実行タイミングは USER。

### TASK-010: scenarios【要実測】一括実測バックログ（Medium）

- **問題**: scenarios 内に【要実測】が残存。仕様文書に期待値が無い挙動の初回実測→期待結果昇格が未完。コード確定 Class E は本 wave で昇格済み。残りは実測レーン。
- **根拠**: residual **SCEN-BROWSER-001**。README 方針: 【要実測】は初回実行で実測し正しければ期待結果へ昇格。AI 実行は browser-test スキル粒度。
- **修正方針**: **1 本の backlog TASK** に集約（項目ごと 70 TASK は作らない）。残件一覧は `rg '【要実測】' docs/ops/testing/scenarios` で再取得。browser-test レーンで実測→正しいものはマーク除去＋期待結果明記、バグは BUG 起票、仕様不明は 要PO へ分岐。実行記録はシナリオ本文に書かず `reports/YYYY-MM-DD-<env>.md`。
- **受け入れ条件**: ① 【要実測】の残件が 0、または「PO待ち / BUG-xxx」に振分済みで残件表が台帳から辿れる。② PASS 分はマーク除去済み。③ reports に実行記録あり。
- **状態**: open / Medium。理想は TASK-009（seed）後。seed 無しでも検索可能な項目から先行可。

### TASK-011: S11 手順2 叙述の受付済-only 整合（Low）

- **問題**: S11 手順2: expected は受付済-only だが、手順叙述が「受付済→診療中→open」と読めて紛らわしい。
- **根拠**: residual **SCEN-S11-COPY-001**。実装/期待は wave で揃済み、**文言 residual のみ**。対象: `docs/ops/testing/scenarios/S11-trimming-combined-accounting.md`。
- **修正方針**: S11 の手順2 叙述のみ修正。expected と操作の対応を受付済-only に統一。製品コード・seed 変更なし。
- **受け入れ条件**: ① 手順叙述と期待結果が矛盾しない。② レビューで「診療中必須」と誤読されない。
- **状態**: open / Low。依存なし。即着手可。

### TASK-012: 見積 unlock と 26§2.1 の仕様整合（High・要PO判断）

- **問題**: estimate unlock 挙動と仕様 26§2.1 の解釈が未確定。実装を独断で変えると臨床/会計安全を損なう。
- **根拠**: residual **SCEN-PROD-ESTIMATE-001**。S07 等に unlock 導線は現行実装に無い（要仕様決定）の honesty あり。product philosophy: 責任者と目的が無い変更は要件ではない。
- **修正方針**: **実装開始条件: PO決定**。論点: unlock 可否・権限・監査・確定後ロックとの関係・26§2.1 の正本解釈。決定後に実装 TASK へ分割可。PO 未決のままコード変更しない。
- **受け入れ条件**: ① PO が unlock の許可範囲を文書化（specs または決定記録）。② 実装変更は決定後 follow-up で受け入れ再定義。③ 決定前にコード変更しない。
- **状態**: open / High・要PO判断。依存: PO 決裁。

### TASK-013: 締め reverse 境界の仕様確定（High・要PO判断）

- **問題**: 締め（closing）の reverse 境界（AM/PM/EMG・越日・取消/訂正の許容範囲）が PO 未決。S09 と実装の差分リスク。
- **根拠**: residual **SCEN-PROD-CLOSING-001**。締め後会計編集は fail-closed 監査・同一 TX 対象になり得る。
- **修正方針**: **実装開始条件: PO決定**。境界表（どの status/時間帯で reverse 可/不可）を仕様正本へ。決定後にコード・シナリオを一括追随。決定前に reverse 緩和を独断実装しない。
- **受け入れ条件**: ① reverse 境界が仕様に表形式で固定。② 決定前コード変更なし。③ 決定後 follow-up で S09 と実装が一致。
- **状態**: open / High・要PO判断。依存: PO 決裁、S09。

### TASK-014: payment system_key の仕様・運用方針（Medium・要PO判断）

- **問題**: payment `system_key` の扱い（固定キー・表示名・マスタ編集可否・互換）が PO 未決。
- **根拠**: residual **SCEN-PROD-PAYMENT-KEY-001**。会計マスタ変更は精算・締め・訂正に波及。関連: S08 / V04 会計マスタ。
- **修正方針**: **実装開始条件: PO決定**。system_key の不変性、UI 露出、seed/マスタ初期値、既存データ移行要否を決裁。決定前にキー体系を変える PR を出さない。
- **受け入れ条件**: ① system_key 方針が1文書に記載。② 決定前に破壊的キー変更をしない。③ 実装 follow-up の要否が明確。
- **状態**: open / Medium・要PO判断。依存: PO 決裁。

### TASK-015: 監査 MEDIUM 残渣の scenarios 再突合 gate（Low）

- **問題**: 監査 MEDIUM 残渣の再スキャン余地がある。wave で主要 S/V は DOC_ALIGNED 済みだが、取りこぼし防止の gate が未手順化。
- **根拠**: residual **SCEN-AUDIT-MED-001**。必須ではなく任意だが抜け防止に有効。pre-WAVE1 open_questions の未 SCEN 化分（V05 full form audit 等）は本 gate で「残差なし」または別 ID 起票に振り分ける。
- **修正方針**: 短時間の再突合チェックリスト（scenarios 索引 × 監査メモ）を1回実行。乖離があれば BUG/要PO/docs に振り分け。乖離ゼロなら本 TASK を done で閉じる。大規模再監査はしない。
- **受け入れ条件**: ① 再突合結果が reports または台帳コメントに1回記録。② 新規 open が residual ID 付きで起票されるか「残差なし」と明示。
- **状態**: open（任意）/ Low。依存: 本 wave DOC_ALIGNED 完了後。TASK-010 と並列可。

### TASK-016: ADR-003 References の歴史的 `internal/service` パス掃除（Low / docs）

- **問題**: ADR-003 本文 References が削除済みの `backend/internal/service/...` を指したまま残り、現行 write owner（`internal/billing`）と読み手が誤認しうる。Decision 本文の歴史記述ではなく **References の生存パス整合**が残債（非ブロック）。
- **根拠**: residual **ARCH-R1**。`docs/architecture/adr/003-payment-method-identity-and-consistency.md` L182–183 が `backend/internal/service/accounting_service_builders.go` / `cash_register_service.go` を列挙。`backend/internal/service` ディレクトリは不在。現行ファイルは `backend/internal/billing/accounting_service_builders.go` / `cash_register_service.go`（同 ADR 内 live 参照と不整合）。architecture Mode 3 本体は **ARCH-DONE**。
- **修正方針**: References を現行 `internal/billing/*` パスへ更新するか、「旧 path・historical」と明示して読者が現存コードを辿れるようにする。ADR Decision / Status の意味変更はしない。BE9 inventory や他 historical 文書は本 TASK 対象外。
- **受け入れ条件**: ① References 上の path が tree に存在するか historical 明示がある。② `internal/service` を現行パスとして読ませない。③ 製品コード変更なし。
- **状態**: open / Low（docs debt）。ARCH-DONE の再オープンではない。SCEN TASK-009–015 とは独立。

### TASK-017: 索引フッタ TABLE_COUNT と docs-symbol-drift 3a スコープ方針（Low / 任意）

- **問題**: テーブル数の正本は ERD / specification で **115**（`<!-- ERD:TABLE_COUNT=115 -->`、migrations 行頭 `CREATE TABLE` 合算 001:110+004:4+006:1）だが、一部索引フッタが **110** のまま。`check-docs-symbol-drift.sh` 3a は ERD+specification のみ強制し、README/screens/overview 系フッタを意図的に対象外にしているため、フッタ drift がゲートをすり抜ける。
- **根拠**: residual **ARCH-R4**。`scripts/check-docs-symbol-drift.sh` L192 コメント「docs/README.md / screens/README.md / overview の索引フッタは別 unit で追随」。実測: `docs/README.md:40`・`docs/spec/screens/README.md:90`・`docs/ops/deploy/DEPLOYMENT_CHECKLIST.md:39` が **110**。ERD/specification **115** は ARCH-DONE（追記 commit `07e2fb945`、path-scoped）。フッタの **115** 手動同期は **未完了がディスク証拠**（旧 114 目標文言は破棄し 115 に合わせる）。
- **修正方針**: 案A（推奨最小）: 上記 3 ファイルの 110→**115** のみ同期（Resources 35 は維持。DEPLOYMENT_CHECKLIST は「全テーブル=001_init のみ 110」と読める曖昧さを、物理総数 115 と 001=110 の区別が分かる表現へ直す）。案B: 3a の `check_number` 対象に該当フッタを戻し機械強制。案C: フッタを out-of-scope のまま accepted とし台帳/コメントで固定（同期は任意）。handler 3c の dead branch は本 TASK で触らない。
- **受け入れ条件**: 案A なら ① 3 ファイルが **115**（または 001=110 と総数 115 の区別が正しい）② ERD/specification と矛盾なし ③ 既存 3a PASS。案B なら フッタ不一致で FAIL。案C なら out-of-scope が 1 箇所に文書固定。製品コード・migration 変更なし。**115 同期の再実装（ERD/spec の再編集）はしない**。
- **状態**: open / Low・任意。ARCH-DONE（ERD/gate/FE labels/narrative/115 path-scoped）の再実装ではない。residual **SPEC-TOP-FOOTER-115** を本 TASK に merge（重複 TASK を作らない）。

### TASK-018: docs/spec/screens/** 全量 vs code 再監査（Medium / 任意）

- **問題**: screens-spec Mode 3 / Pack A–F 後も、`docs/spec/screens/**` 全体の再フルパス監査は未手順化。既存 open は R1–R3（TASK-001–003）に限定され、他画面の silent drift を拾う lane が無い。
- **根拠**: residual **SPEC-TOP-SCREENS-AUDIT**。索引: `docs/spec/screens/README.md`（番号付き 41 + settings/）。製品 leaf 84 は route-inventory / e2e inventory と別カウント。Mode 3 COMPLETE は「既知 residual のみ open」方針だが、全量再パスは optional として台帳に明示する必要がある。
- **修正方針**: **1 TASK に集約**（画面ごと 70 TASK は作らない）。(1) README 索引 × route-inventory 84 の突合。(2) 差分は docs honesty / BUG / 要PO / WONTFILE に振り分け、既存 open TASK に吸収できるものは吸収。(3) 製品コード変更は本 TASK では開始条件を満たした follow-up のみ。紙/Excel 忠実複製の新規要件を発明しない。
- **受け入れ条件**: ① 監査結果が reports または台帳コメントに1回記録。② 新規 open が residual/TASK ID 付き、または「残差なし」。③ inventory 84 / 番号付き md 41 の数え分けが結果に明記。
- **状態**: open（任意）/ Medium。依存: なし。TASK-001–003 と独立（重複発見時は既存 TASK にリンク）。

### TASK-019: docs/spec/line/** 全量 vs code 再監査（Medium / 任意）

- **問題**: `docs/spec/line/**`（LIFF / L-step / setup 等）と実装のフル突合が本 ledger に未掲載。LINE 面の silent drift を拾う lane が無い。
- **根拠**: residual **SPEC-TOP-LINE-AUDIT**。tree: `docs/spec/line/`（README / architecture / reservation-spec / lstep-integration / setup 等）。`docs/spec/README.md` 索引あり。todo 索引に line 専用 residual が無かった。
- **修正方針**: **1 TASK に集約**。(1) line 仕様索引 × BE webhook / FE LIFF / L-step 経路を突合。(2) 差分を docs / BUG / 要PO / ops に振り分け。(3) 秘密情報・本番 webhook 操作は対象外。migration 適用は USER。
- **受け入れ条件**: ① 監査結果が1回記録。② 新規 open が ID 付きまたは「残差なし」。③ 製品コードは決定後 follow-up のみ。
- **状態**: open（任意）/ Medium。依存: なし。SCEN / screens TASK と独立。

### TASK-020: ui-design-compliance Playwright 再 runtime（84 製品ページ）（Low / 任意）

- **問題**: e2e inventory は 84 製品ページへ静的更新済みだが、`/identity-links` を含む **runtime 全数再確認**は doc 上 deferred のまま。
- **根拠**: residual **SPEC-TOP-E2E-RUNTIME-84**。`docs/spec/ui-design-compliance.md` L41 / L132: 最終 full runtime は 2026-07-23（83 製品 / 92 tests）。在庫 84 後の期待は 93 tests。静的 inventory: `frontend/e2e/ui-design-compliance-readonly.spec.ts`、route-inventory `expect(pages).toHaveLength(84)`。
- **修正方針**: 任意 ops。`docker compose` 環境で `frontend/e2e/ui-design-compliance-readonly.spec.ts` を workers=1 再実行。結果を reports または ui-design-compliance の runtime 日付へ反映。失敗時は BUG / inventory 修正 TASK へ分岐。製品機能の新規実装は本 TASK の範囲外。
- **受け入れ条件**: ① 再 runtime 結果（PASS 件数 or 失敗一覧）が1箇所に記録。② inventory 84 と矛盾しない。③ 失敗は silent にしない。
- **状態**: open（任意）/ Low。静的 inventory 84 は already done。TASK-010（scenarios 要実測）とは別 lane。

### TASK-021: staff 対応可能種別 dual surface の収束可否（Medium・要PO判断）

- **問題**: write SoT は `staff_reservation_capabilities` と doc 固定済みだが、exclusion 形 API/UI 名・候補フィルタが残存し dual surface のまま。全面 CRUD 排除か、exclusion を互換 facade として残すかが PO 未決。
- **根拠**: residual **SPEC-TOP-CAPABILITIES-CRUD**。doc: `docs/spec/reservation-to-record-flow.md`（capabilities SoT + dual residual 明示）。BE: `backend/internal/staff/handler.go` に excluded / capable 両系。FE: staff は capable API 書込でも section 名が `StaffExcludedReservationTypesSection`、予約候補が `excluded_courses` 依存の経路あり（例: ReservationFormFields）。available-staffs は **導入しない**（SPEC-TOP-AVAILABLE-STAFFS / WONTFILE）— 本 TASK で available-staffs を実装しない。
- **修正方針**: **実装開始条件: PO決定**。論点: (1) exclusion API の削除/readonly/互換維持 (2) FE 命名と候補フィルタを capabilities 正本へ寄せるか (3) master-staff 仕様の最終 SoT。決定前に破壊的 API 削除をしない。product philosophy: 二重管理を増やさない。
- **受け入れ条件**: ① PO が dual surface の最終方針を文書化。② 決定前にコードで exclusion を独断削除しない。③ 実装 follow-up の要否が明確。
- **状態**: open / Medium・要PO判断。SPEC-TOP-CAP-SOT-DOC（doc 固定）は already done。依存: PO 決裁。

---
