# Remaining work ledger (open only)

オープン residual のみを列挙する。対応済み TASK / closed 索引行は **削除済み**（2026-07-31 更新）。  
根拠・完了証拠は git 履歴と `reports/2026-07-31-*.md` を参照。

> **ID namespace**: 本ファイルの `TASK-*` はローカル連番。`3-session-agent.html#ledger` 体系外。`/implement` は正本 ledger からのみ解決。

## 索引 / サマリー

| Inv | 内容 | 処置 |
|-----|------|------|
| R4 | screens-drift 意図変更のコミット隔離 | **TASK-004**（ops 手順・land は USER） |
| R5 | コミット前の closed pack 回帰ゲート | **TASK-005**（ops 手順・land 前再実行） |
| R6 | マルチエージェント共有 tree thrash | **ops-only** |
| R7 | empty-diff 成功宣言 harness | **ops-only** |
| SCEN-SEED-001 | 003_demo clinical CSV ヘッダのみ | **TASK-009**（CSV slice1 done・**適用は USER**） |
| SCEN-BROWSER-001 | scenarios 内【要実測】backlog | **TASK-010**（BLOCKED env） |
| SCEN-OPS-CLAIM-001 | `claim/*` 解放 | **ops-only**（USER only） |
| SCEN-OPS-COMMIT-001 | mixed commit 説明メモ | **ops-only**（rewrite しない） |
| SCEN-OPS-TREE-001 | 共有 tree concurrent WIP | **ops-only**（= R6） |
| ARCH-R2 | empty-diff COMPLETE 規律 | **ops-only**（継続） |
| ARCH-R3 | land 時 foreign 定義は `git status` 実測 | **ops note** + TASK-004 |
| POST-PULL | migrations 適用 | **ops-only** ≡ **SPEC-TOP-MIGRATE-006**（USER `make migrate`） |
| SPEC-TOP-LINE-AUDIT | `docs/spec/line/**` deep 監査 | **TASK-019 done**（残差 R-01..R-08 のみ） |
| SPEC-TOP-E2E-RUNTIME-84 | Playwright runtime 84 | **TASK-020**（BLOCKED env） |
| SPEC-TOP-CAPABILITIES-CRUD | exclusion 面の破壊削除 | **TASK-021 Stage A**（Stage B 実装済） |
| SPEC-TOP-CLAIM-RELEASE | claim 解放 | **SCEN-OPS-CLAIM-001** |

### 対応済み（削除済み・再掲しない）

TASK-001-BE/FE, TASK-002/003（WONTFIX + UI follow-up 実装済）, TASK-006/007/008, TASK-011, TASK-012/013/014（Wave1 実装済）, TASK-015/016/017, TASK-018, TASK-019 deep, TASK-021 Stage B, ARCH-DONE, SPEC-TOP-G1-G12, SPEC-TOP-FOOTER-115, SPEC-TOP-CAP-SOT-DOC, SPEC-TOP-AVAILABLE-STAFFS（WONTFILE）, R1–R3, R8-\*, SCEN-S11-COPY-001, SCEN-AUDIT-MED-001, ARCH-R1, ISSUE-261 P0 deceased-pet write guards（`79fe62265`）。

### Ops-only notes（製品コード TASK にしない）

- **R6 / SCEN-OPS-TREE-001**: 並行エージェントは worktree 隔離。共有 tree は 1 編集セッションのみ。
- **R7 / ARCH-R2**: 受け入れは `git diff` / `git status` の実 diff 必須。empty-diff COMPLETE 禁止。
- **ARCH-R3 / TASK-004**: land 直前の `git status --porcelain` で intentional / foreign を定義。台帳に dirty 一覧を書かない。
- **POST-PULL / SPEC-TOP-MIGRATE-006**: USER が `make migrate`。エージェントは auto-apply しない。migrations `002`/`003` は local 適用済みの可能性あり — 他環境は再確認。
- **SCEN-OPS-CLAIM-001**: claim 解放は USER only（統合後）。
- **SCEN-OPS-COMMIT-001**: mixed history の説明用。history rewrite / force-push しない。

### 推奨実装順（open のみ）

1. **TASK-009** seed 適用（USER。設計: `reports/2026-07-31-task-009-seed-design.md` / slice1: `reports/2026-07-31-task-009-slice1.md`）  
2. **TASK-010** browser 要実測（env 後。seed 後が理想）  
3. **TASK-020** Playwright 84 runtime（env 後）  
4. **TASK-021 Stage A**（consumer inventory + 破壊変更承認後）  
5. **TASK-004 / TASK-005**: 次の intentional land 時  

---

## 個別タスク詳細

### TASK-004: screens-drift 意図変更セットのコミット隔離（Medium・ops）

- **問題**: intentional と foreign を同一 commit に混ぜない。foreign 定義は land 直前の `git status` / `git diff` が正本。
- **修正方針**: land 直前に porcelain 実測 → path-scoped `git add`（`git add -A` 禁止）。foreign は触らない・捨てない。
- **受け入れ条件**: staged ⊆ intentional; foreign 非 stage; 破棄しない。
- **状態**: **ops 手順 open**（再発・次 land 用）。前回実測: `reports/2026-07-31-task-004-005-land-proc.md`。

### TASK-005: closed packs 回帰のコミット前検証ゲート（Medium・ops）

- **問題**: land 前に doc/code 整合と inventory / hospitalization を機械確認する手順。
- **修正方針**: land 直前: `bash scripts/check-docs-symbol-drift.sh`; scoped hospitalization / route-inventory tests。結果は reports に記録。
- **受け入れ条件**: ゲート PASS; inventory 84 維持; hospitalization unit PASS。
- **状態**: **ops 手順 open**（land 都度）。

### TASK-009: 003_demo clinical CSV ヘッダのみ — seed 再投入（High）

- **問題**: clinical CSV がヘッダのみでシナリオ前提データが揃わない可能性。
- **修正方針**: 設計 `reports/2026-07-31-task-009-seed-design.md` に従い USER が seed 適用。エージェントは migrate/seed auto-apply しない。
- **受け入れ条件**: 対象 CSV がヘッダのみでなくなる; シナリオ前提を満たす; 適用手順が1箇所で辿れる; 適用は USER。
- **状態**: **CSV slice1 committed（authoring done）/ 適用は USER**。slice1: hospitalizations + treatment_plans + daily_records + care_plan_items（G1 medical_records は既存 dump で充足）。証拠: `reports/2026-07-31-task-009-slice1.md`。claim: `claim/TASK-009`（USER が統合後に解放）。

### TASK-010: scenarios【要実測】一括実測バックログ（Medium）

- **問題**: scenarios に【要実測】残存。
- **修正方針**: browser-test レーンで実測。記録は `reports/`。
- **受け入れ条件**: 要実測 0 または PO/BUG 振分; reports に実行記録。
- **状態**: **BLOCKED（env）**。`reports/2026-07-31-task-010-020-runtime-blocked.md`。

### TASK-019: docs/spec/line/** deep 監査 follow-up（Medium / 任意）

- **問題**: line 仕様 vs 実装の deep 突合が partial のまま。
- **根拠**: 初回記録 `reports/2026-07-31-task-019-line-audit.md`。
- **修正方針**: deep pass で差分を docs/BUG/要PO/ops に振分。秘密・本番 webhook 操作は対象外。
- **受け入れ条件**: deep 結果1回記録; 新規 open は ID 付きまたは残差なし。
- **状態**: **done**（deep: `reports/2026-07-31-task-019-line-deep-audit.md`）。残差 ID のみ: **R-01** 要PO（webhook follow/unfollow 契約を architecture に書くか）、**R-02** ops（本番 webhook / line_bot_user_id）、**R-03**→TASK-010 要実測、**R-04** ops/USER（Write dual-gate 再有効化）、**R-05** 要PO（LINE credential 二重ストア SoT）、**R-06** 要PO（delivery-monitor nav 有無）、**R-07** 要PO（tags sidebar vs route RBAC）、**R-08** ops（LIFF ID 二重 bootstrap）。docs honesty 最小修正済（reservation-spec / architecture / lstep-integration / README / setup / screen37）。

### TASK-020: ui-design-compliance Playwright 再 runtime（84）（Low / 任意）

- **問題**: inventory 84 静的更新後の full runtime 未実施。
- **修正方針**: env 可なら `ui-design-compliance-readonly.spec.ts` workers=1。結果を reports へ。
- **状態**: **BLOCKED（env）**。`reports/2026-07-31-task-010-020-runtime-blocked.md`。

### TASK-021 Stage A: exclusion 面の破壊的撤去（Medium・PO決裁済・未着手）

- **問題**: Stage B で facade 化済み。exclusion route/payload/model/table の最終撤去が残る。
- **修正方針**: **consumer inventory + 破壊変更の明示承認後**に Stage A（FINAL 参照）。新 endpoint は追加しない。`available-staffs` は WONTFILE。
- **受け入れ条件**: exclusion production surface 削除; migration あり; Stage B 互換 consumer が無いこと inventory で証明。
- **状態**: **Stage A remaining**。Stage B: `e9dddd921`。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md`。

---
