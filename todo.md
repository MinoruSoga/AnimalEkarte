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
| SCEN-SEED-001 | 003_demo clinical CSV ヘッダのみ | **TASK-009**（CSV slice1 done・**適用は USER** / reseed ops 文書化済） |
| SCEN-BROWSER-001 | scenarios 内【要実測】backlog | **TASK-010**（env READY・batch2 V04 partial・body 要実測 59） |
| SCEN-OPS-CLAIM-001 | `claim/*` 解放 | **ops-only**（USER only） |
| SCEN-OPS-COMMIT-001 | mixed commit 説明メモ | **ops-only**（rewrite しない） |
| SCEN-OPS-TREE-001 | 共有 tree concurrent WIP | **ops-only**（= R6） |
| ARCH-R2 | empty-diff COMPLETE 規律 | **ops-only**（継続） |
| ARCH-R3 | land 時 foreign 定義は `git status` 実測 | **ops note** + TASK-004 |
| POST-PULL | migrations 適用 | **ops-only** ≡ **SPEC-TOP-MIGRATE-006**（USER `make migrate`） |
| SPEC-TOP-LINE-AUDIT | `docs/spec/line/**` deep 監査 | **TASK-019 done** + **PO FINAL**（R-01/R-05 binding; R-06/R-07 child close + parent follow-up 実装中） |
| SPEC-TOP-E2E-RUNTIME-84 | Playwright runtime 84 | **TASK-020**（env-forward done・runtime credentials BLOCKED） |
| SPEC-TOP-CAPABILITIES-CRUD | exclusion 面の破壊削除 | **TASK-021 Stage A**（Phase1 done; **Phase2 start approved**; CLEAN-GO/DROP HOLD） |
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

1. **TASK-009** seed 適用（USER。reseed ops: `reports/2026-07-31-task-009-reseed-ops.md`）  
2. **TASK-010** 要実測残 backlog（body 59。batch2: `reports/2026-07-31-task-010-batch2.md`）  
3. **TASK-020** Playwright 93 runtime 完走（env-forward 済・要 host `E2E_LOGIN_*`。証拠: `reports/2026-07-31-task-020-env-forward.md`）  
4. **TASK-022 human residual** — S13 手動 correction + named signer + RLS runtime（agent source closeout 済）  
5. **TASK-023 human residual** — E2E_LOGIN_* 注入・5フロー通し・DB/audit・LINE/LIFF・sign-off（agent 証跡骨格 済）  
6. **TASK-024 human residual** — named documentation owner visual sign-off（agent 10/10 audit + FAQ 追記不要 済）  
7. **TASK-021 Stage A 削除**（Phase1 FE residual: `reports/2026-07-31-task-021-phase1-consumer-prep.md` — BE/OpenAPI/seed/DB consumer 撤去後 + 破壊変更承認）
8. **TASK-004 / TASK-005**: 次の intentional land 時
9. **LINE follow-up（PO FINAL 済）**: `reports/2026-07-31-line-residual-po-decisions-FINAL.md`
   - High R-05 single-SoT cutover（`clinic_integrations`）
   - High R-06/R-07 parent RBAC honesty（本 session で着手）
   - Medium R-01 architecture summary + contract tests（本 session で着手）
   - R-02/R-04/R-08 は ops のまま

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
- **状態**: **CSV slice1 committed（authoring done）/ 適用は USER**。slice1: hospitalizations + treatment_plans + daily_records + care_plan_items（G1 medical_records は既存 dump で充足）。証拠: `reports/2026-07-31-task-009-slice1.md`。**USER reseed 手順**: `reports/2026-07-31-task-009-reseed-ops.md`（既適用 DB は checksum mismatch → `make reset` が正。agent は auto wipe しない）。claim: `claim/TASK-009`（USER が統合後に解放）。

### TASK-010: scenarios【要実測】一括実測バックログ（Medium）

- **問題**: scenarios に【要実測】残存。
- **修正方針**: browser-test レーンで実測。記録は `reports/`。
- **受け入れ条件**: 要実測 0 または PO/BUG 振分; reports に実行記録。
- **状態**: **env READY / batch2 partial**（2026-07-31 next orch wave）。docker healthy + `:8080/health` 200 + `:3003/` 200。batch1 V05: 5 件（証拠: `reports/2026-07-31-task-010-runtime-batch.md`）。**batch2 V04**: 6 件 elevate（要実測 body **65→59** / V04 11→5）。証拠: `reports/2026-07-31-task-010-batch2.md`。残 backlog open。claim: `claim/TASK-010`（USER 解放）。

### TASK-019: docs/spec/line/** deep 監査 follow-up（Medium / 任意）

- **問題**: line 仕様 vs 実装の deep 突合が partial のまま。
- **根拠**: 初回記録 `reports/2026-07-31-task-019-line-audit.md`。
- **修正方針**: deep pass で差分を docs/BUG/要PO/ops に振分。秘密・本番 webhook 操作は対象外。
- **受け入れ条件**: deep 結果1回記録; 新規 open は ID 付きまたは残差なし。
- **状態**: **done**（deep: `reports/2026-07-31-task-019-line-deep-audit.md`）。**PO FINAL**: `reports/2026-07-31-line-residual-po-decisions-FINAL.md`（`3d448ec5e`）。**R-01** binding B（code/tests SoT + architecture 要約）— follow-up docs/test。**R-05** binding A-CI（SoT=`clinic_integrations`）— High cutover 未実装。**R-06/R-07** original child residual close; parent-container / `/lstep` wrapper honesty follow-up。**R-02/R-04/R-08** ops、**R-03**→TASK-010。claim: `claim/LINE-R-FIX` / `claim/LINE-PO-R01-R05` / `claim/LINE-PARENT-RBAC`（USER 解放）。

### TASK-020: ui-design-compliance Playwright 再 runtime（84）（Low / 任意）

- **問題**: inventory 84 静的更新後の full runtime 未実施。
- **修正方針**: env 可なら `ui-design-compliance-readonly.spec.ts` workers=1。結果を reports へ。
- **状態**: **env-forward done / runtime credentials BLOCKED**（2026-07-31 next orch）。`run-e2e.sh` が host に設定時のみ `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`（+ optional `E2E_AUTH_STATE_PATH`）を name-only `-e` で Playwright docker へ転送。証拠: `reports/2026-07-31-task-020-env-forward.md`。prior runtime: 4p/3f/86 DNR（`reports/2026-07-31-task-020-runtime.md`）。host が EMAIL_UNSET/PASSWORD_UNSET のため再 runtime 未実施。full green 未達。claim: `claim/TASK-020` + `claim/W-020-ENV`（USER 解放）。

### TASK-021 Stage A: exclusion 面の破壊的撤去（Medium・PO決裁済・inventory 済）

- **問題**: Stage B で facade 化済み。exclusion route/payload/model/table の最終撤去が残る。
- **修正方針**: **consumer inventory + 破壊変更の明示承認後**に Stage A（FINAL 参照）。新 endpoint は追加しない。`available-staffs` は WONTFILE。
- **受け入れ条件**: exclusion production surface 削除; migration あり; Stage B 互換 consumer が無いこと inventory で証明。
- **状態**: **Phase1 FE residual SAFE-CLEANUP done / Phase2 START APPROVED / CLEAN-GO·DROP·migrate HOLD**。Stage B: `e9dddd921`。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md` + LINE residual FINAL（021 Phase2）。inventory: `reports/2026-07-31-task-021-stage-a-inventory.md`。Phase1: `reports/2026-07-31-task-021-phase1-consumer-prep.md`。Phase2 = §6.1–§6.3 consumer/BE/OpenAPI のみ。claim: `claim/TASK-021` + `claim/W-021-P1`（USER 解放）。

### TASK-022: #239 Phase 1 closeout と代表手動 correction gate（High）

- **対応 Issue**: GitHub Issue #239（live state は CLOSED。未充足の受け入れ条件を local New Work として追跡）。
- **状態**: **agent source closeout done / human residual open**（2026-07-31）。`CreatePetGroup` の any-member fallback を除去し、親 owner-group の anchor + 全 active member clinic を actor に要求する regression を green（`go test -p 1 ./internal/identitylink ./internal/apicontract -count=1`）。Phase 2 未着手。証拠: `reports/2026-07-31-task-022-identity-link-closeout.md`、S13: `docs/ops/testing/scenarios/S13-identity-links-manual-correction.md`。security review PASS。
- **残 human**: named operator の 2 医院 link→history→unlink→relink 実施と named signer 承認；RLS runtime を実 application role で証明（未なら UNREPORTED のまま Phase 2 禁止）。
- **claim**: `claim/TASK-022`（USER が main 統合確認後に解放）。

### TASK-023: #254 5業務フロー UAT 統合証跡（High）

- **対応 Issue**: GitHub Issue #254。
- **状態**: **agent 証跡骨格 done / human residual open**（2026-07-31）。統合 report: `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md`。env health PASS；`E2E_LOGIN_*` 未注入で authenticated E2E **BLOCKED**（EXIT:1、値非出力）；5 フロー step 全行に status/executor/env/evidence/owner；agent と human 欄分離；`confusion_count: 0`。#254 は human 揃うまで完了扱いしない。
- **残 human**: USER が secret channel で `E2E_LOGIN_*` 注入後 E2E 実行；QA が 5 フローブラウザ通し・DB/audit 目視・実 LINE/LIFF；PO/現場の使い勝手 sign-off と FAIL 処分承認。
- **claim**: `claim/TASK-023`（USER 解放）。

### TASK-024: #256 現行 screenshot / FAQ finalization（Medium）

- **対応 Issue**: GitHub Issue #256。
- **状態**: **agent audit + FAQ disposition done / human visual sign-off open**（2026-07-31）。10/10 current/replace 判定と 7 画像同名置換。FAQ は TASK-023 `confusion_count: 0` に基づき **追記不要**（`10-troubleshooting.md` 変更なし）。証拠: `reports/2026-07-31-task-024-manual-audit.md`。vitest manual 18 tests PASS。manual-flow E2E は env 未注入で BLOCKED。
- **残 human**: named documentation owner の visual/content sign-off；任意で `19-aggregation` / `04-medical-records` 再撮影。
- **claim**: `claim/TASK-024`（USER 解放）。

---
