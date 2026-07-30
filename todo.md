# Screens-spec drift — remaining work ledger

本台帳は screens-spec 実装ドリフト監査 + fix-all（Mode 3 COMPLETE）後に残ったオープン項目のみを列挙する。
Pack A–F の doc honesty 出荷済み分は再掲しない（再スキャンで回帰が確認されたもののみ再オープン）。

## 索引 / サマリー

| Inv | 内容 | 処置 |
|-----|------|------|
| R1 | 入院 create + nested treatment-plan POSTs が非原子的 | **TASK-001-BE / TASK-001-FE** |
| R2 | 編集モード治療プランが read-only のまま | **TASK-002**（要PO判断） |
| R3 | フォーム一括割引が非永続 | **TASK-003**（要PO判断） |
| R4 | screens-drift 意図変更のコミット隔離 | **TASK-004**（stage plan 済み・未 stage / 未 commit） |
| R5 | コミット前の closed pack 回帰ゲート | **TASK-005**（scoped gates 実行済み・land 前に再実行可） |
| R6 | マルチエージェント共有 tree thrash | **ops-only**（下記）— 製品 TASK なし。既存 `git-worktree-safety` を正本とし worktree 隔離を徹底 |
| R7 | Writer が diff なしで成功宣言した harness 問題 | **ops-only**（下記）— 製品 TASK なし。任意で CLAUDE/Agents に「empty-diff ≠ success」1行を後続で検討 |
| R8-D | `40-identity-links.md` 欠落 + README 索引不全 | **TASK-006 done**（docs 作成済） |
| R8-B03 | `03-owners-list.md` が client-side `/owners` のまま | **TASK-007 done** |
| R8-board | ボードカード open が `edit` 必須 vs 仕様未記載 | **TASK-008 done** |
| R8-README | README「全40」 | **TASK-006 done に包含** |
| 一時帰宅 | ドキュメント honesty 済み / BE enum 一致 | **WONTFILE** — mismatch ではない |

### Ops-only notes（製品コード TASK にしない）

- **R6**: 並行 Grok/Claude は共有 working tree 禁止。`.claude/rules/git-worktree-safety.md` 準拠で worktree 分離。repo への追記は任意。
- **R7**: harness が empty-diff 成功を許す問題。prompt-craft / writer ゲート側。本 repo の機能 TASK ではない。

### 推奨実装順

1. TASK-004 path-scoped stage → 再 TASK-005 → ユーザー承認後 commit で screens-drift land  
2. ~~TASK-006 / 007 / 008~~（docs residual 完了）  
3. TASK-001-BE → TASK-001-FE（原子性、臨床 integrity）  
4. TASK-002 / TASK-003（PO 判断待ち）

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

- **問題**: 共有 tree に screens-drift 修正と foreign WIP（billing / examinations / scenarios / architecture / package 等）が混在。一括 stage すると異レーンが同一 commit に混入する。
- **根拠**: intentional に `docs/spec/screens/*`（+ 新規 `40-identity-links.md`）、Pack C hospitalization form/write、`route-inventory.test.tsx`、`master-settings-index-model(.test).ts` を含む。Foreign は billing/csvimport/examinations/scenarios/architecture/package/symbol-drift scripts/codex-security 等。identity-links FE / ExamTypeFieldsEditor は ambiguous（既定 OUT）。
- **修正方針**: path-scoped `git add` のみ（`git add -A` 禁止）。Completion Report の stage コマンドを使用。foreign は触らない・捨てない。
- **受け入れ条件**: ① `git diff --cached --name-only` が intentional 集合 ⊆ のみ。② foreign パスが staged に含まれない。③ commit 後も foreign WIP が working tree に残る（破棄しない）。
- **状態**: Medium/ops。**stage plan 準備済み（未 stage）**。commit はユーザー承認境界。

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
- **状態**: **done**（docs 作成・README 更新・symbol-drift PASS）。commit は TASK-004 と同時 land。

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

---
