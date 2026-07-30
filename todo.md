# Screens-spec drift — remaining work ledger

本台帳は screens-spec 実装ドリフト監査 + fix-all（Mode 3 COMPLETE）後に残ったオープン項目のみを列挙する。
Pack A–F の doc honesty 出荷済み分は再掲しない（再スキャンで回帰が確認されたもののみ再オープン）。

## 索引 / サマリー

| Inv | 内容 | 処置 |
|-----|------|------|
| R1 | 入院 create + nested treatment-plan POSTs が非原子的 | **TASK-001-BE / TASK-001-FE** |
| R2 | 編集モード治療プランが read-only のまま | **TASK-002**（要PO判断） |
| R3 | フォーム一括割引が非永続 | **TASK-003**（要PO判断） |
| R4 | screens-drift 意図変更のコミット隔離 | **TASK-004** |
| R5 | コミット前の closed pack 回帰ゲート | **TASK-005** |
| R6 | マルチエージェント共有 tree thrash | **ops-only**（下記）— 製品 TASK なし。既存 `git-worktree-safety` を正本とし worktree 隔離を徹底 |
| R7 | Writer が diff なしで成功宣言した harness 問題 | **ops-only**（下記）— 製品 TASK なし。任意で CLAUDE/Agents に「empty-diff ≠ success」1行を後続で検討 |
| R8-D | `40-identity-links.md` 欠落 + README 索引不全 | **TASK-006** |
| R8-B03 | `03-owners-list.md` が client-side `/owners` のまま | **TASK-007** |
| R8-board | ボードカード open が `edit` 必須 vs 仕様未記載 | **TASK-008** |
| R8-README | README「全40」 | **TASK-006 に包含** |
| 一時帰宅 | ドキュメント honesty 済み / BE enum 一致 | **WONTFILE** — mismatch ではない |

### Ops-only notes（製品コード TASK にしない）

- **R6**: 並行 Grok/Claude は共有 working tree 禁止。`.claude/rules/git-worktree-safety.md` 準拠で worktree 分離。repo への追記は任意。
- **R7**: harness が empty-diff 成功を許す問題。prompt-craft / writer ゲート側。本 repo の機能 TASK ではない。

### 推奨実装順

1. TASK-004（隔離）→ TASK-005（検証）で screens-drift を安全に land  
2. TASK-006 / TASK-007 / TASK-008（docs residual、低リスク）  
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
- **根拠**: 現 tree（再スキャン）: intentional 例 — `docs/spec/screens/*`（00/03/07/09/10/11/14/15/18/20/22/25/26/31/34/39/README/settings/master-examinations）、hospitalization form hooks/routes + `treatment-plans-write(.test).ts`、`route-inventory.test.tsx`、`master-settings-index-model(.test).ts`、任意で identity-links / ExamTypeFieldsEditor / permission-rule-table-model。Foreign 例 — billing、csvimport、examinations、`docs/ops/testing/scenarios/*`、`docs/architecture/*`、`docs/spec/specification.md` 等、`package.json`/`pnpm-lock.yaml`、symbol-drift scripts、`codex-security-output/`。`40-identity-links.md` は **欠落**（TASK-006）。
- **修正方針**: path-scoped `git add` のみ（`git add -A` 禁止）。可能なら worktree で intentional のみ land。ambiguous（README / identity-links / master-settings-index-model / ExamTypeFieldsEditor）は hunk 確認後。foreign は触らない・捨てない。
- **受け入れ条件**: ① `git diff --cached --name-only` が intentional 集合 ⊆ のみ。② foreign パスが staged に含まれない。③ commit 後も foreign WIP が working tree に残る（破棄しない）。
- **状態**: Medium/ops。TASK-005 とセットで land 前に実施。コミット自体はユーザー承認境界。

### TASK-005: closed packs 回帰のコミット前検証ゲート（Medium・ops）

- **問題**: screens-drift land 前に Pack A–F 相当の doc/code 整合と inventory 84・hospitalization 保存パスが壊れていないことを機械確認するゲートが手順化されていない。
- **根拠**: 意図パスに hospitalization 保存・route inventory・screens docs が含まれる。symbol-drift スクリプトは foreign dirty の可能性あり（`scripts/check-docs-symbol-drift.*`）— ゲート実行時はスクリプト自体の foreign 変更に注意。
- **修正方針**: land 直前に（ユーザー/Docker）: `bash scripts/check-docs-symbol-drift.sh`（意図 docs が stage 済みの状態）; `docker compose exec frontend pnpm test:run -- src/features/hospitalization src/app/routes/route-inventory.test.tsx src/features/master/routes/master-settings-index-model.test.ts`。失敗時は TASK-004 の stage 集合を見直す。
- **受け入れ条件**: ① 上記コマンドが PASS。② inventory が 84 product pages を維持。③ hospitalization create/plan 関連 unit が PASS。
- **状態**: Medium/ops。TASK-004 の直後。エージェントは full-suite を自動実行しない（scoped のみ）。

### TASK-006: identity-links 画面仕様の新設と screens README 索引更新（High）

- **問題**: 製品ルート `/identity-links` と inventory 84 がある一方、`docs/spec/screens/40-identity-links.md` が欠落。README は「全40」のまま identity-links 行なし。Pack D ドキュメント完了の再スキャンで回帰。
- **根拠**: `docs/spec/screens/40-identity-links.md` 不存在; `docs/spec/screens/README.md` L3 付近「全40」; `route-inventory.test.tsx` が 84 pages; `IdentityLinksPage.tsx` が `ResourceIdentityLinks` view/edit。影響: docs のみ（コード変更不要が原則）。
- **修正方針**: 実装（`IdentityLinksPage.tsx` + routes）を読んで `40-identity-links.md` を他 screens と同型で作成。README に索引行を追加し件数表記を現実（40 ファイル前提なら identity-links を含めた正しい数え方、または「画面仕様ファイル数」と product leaf 84 の関係を混同しない文言）に修正。symbol-drift トークンを満たす。
- **受け入れ条件**: ① `40-identity-links.md` が存在し権限・主 API・主要操作がコードと一致。② README から到達可能。③ `bash scripts/check-docs-symbol-drift.sh` が identity-links 関連で fail しない。
- **状態**: High（docs residual）。TASK-004 の intentional set に含めて land 可。

### TASK-007: 飼主一覧仕様を GET /v1/pets サーバページネーションに合わせて修正（High）

- **問題**: `03-owners-list.md` が `GET /api/v1/owners` + クライアント側フィルタ/ソート、`useDeferredValue` を記載。実装は `ownersLoader` が `GET /v1/pets` で page/search/species/include_deceased をサーバ転送。owners feature に `useDeferredValue` なし。
- **根拠**: doc L58–67 付近; code `frontend/src/features/owners/loaders.ts`（ownersLoader, `/v1/pets`）。影響: docs。
- **修正方針**: 仕様を pets 行粒度・サーバサイド pagination/search に書き換え。API 表を実エンドポイント/権限に合わせて更新。削除・レポート等の残アクションはコード突合のうえ維持/修正。
- **受け入れ条件**: ① 仕様の list 取得が `/v1/pets`（または現行正本）と一致。② 「クライアント側で全件フィルタ」主張が消える。③ symbol-drift が通る。
- **状態**: High（docs residual / Pack B 回帰）。コード変更は不要（実装が正）。

### TASK-008: 入院ボードのカード open が edit 必須であることの仕様正直化（Medium）

- **問題**: 仕様は「カードまたはリスト行クリックで詳細へ」と view 相当に読めるが、ボードは `canOpenCard = occupant && !deceased && canEdit` のため view-only ではカードから詳細遷移できない（リスト側は別経路）。
- **根拠**: `HospitalizationBoard.tsx` L43–44, L80–81; 遷移先は `use-hospitalization-list.ts` の detail href; `07-hospitalization-list.md` L35。影響: docs（または PO が view でも open させたいなら FE 1 行級変更 — **要PO**）。
- **修正方針**: 既定は **docs honesty**: ボードカード open / DnD は `hospitalization:edit`、一覧番号リンク等の view 経路は別と明記。PO が view-open を要求する場合のみ `canOpenCard` を view に緩和する FE タスクに分岐。
- **受け入れ条件**: ① 仕様が board の edit ゲートを明示。② view-only 手動/テストで期待どおり（docs 案ならクリック不可が仕様通り）。③ リスト経路の記述が board と混線しない。
- **状態**: Medium。既定 docs-only。要PO判断は view-open 変更時のみ。

---
