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
4. **TASK-022** #239 Phase 1 closeout と代表手動 correction gate
5. **TASK-023** #254 5業務フロー UAT 統合証跡（TASK-009/010/020 を再利用）
6. **TASK-024** #256 現行 screenshot / FAQ finalization（FAQ は TASK-023 後）
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
- **問題**: reverse lookup OpenAPI の着地 commit で Issue は閉じられたが、親 owner group 全医院に対する actor 認可と、代表データでの link → history → unlink → relink の回復証跡は無い。Issue の閉鎖を security/runtime 完了証拠へ読み替えず、Phase 2 の auto-link・merge・追加 surface を開始しない。
- **現況の実測根拠**: `gh issue view 239 --json state,closedAt,body` は CLOSED と未消込の受け入れ条件を同時に返す。reverse lookup は `backend/docs/api.yaml:23559-23621` と `backend/internal/apicontract/openapi_route_drift_test.go:455-457` で parity が閉じている。一方、`backend/internal/identitylink/service.go:481-486` は親 group の anchor clinic を actor が持たない場合に一部 member の可視性だけで継続し得て、この exact parent-anchor/all-owner-member rejection regression が無い。DEC-46 は代表 surface の runtime 実測前に Phase 2 を開始しない。
- **要求挙動**: owner/pet/group の全 ID を認証 actor の clinic set へ相関し、view/edit を分離して mutation 直前に再認可する。hidden・不存在・他院・mixed ID は同じ fail-closed 応答で全件を原子的に拒否し、business write と非 PHI audit は同一 transaction、audit failure は rollback とする。決定的 lock order・CAS/version・unique/idempotency を保ち、RLS は実 application role/session context で強制されることを証明する。代表 correction が不足を示すまで auto-link・merge・追加 surface・新 DDL・record 移動を行わない。
- **Owned paths**: `backend/internal/identitylink/`、`docs/spec/screens/40-identity-links.md`、新規 `docs/ops/testing/scenarios/S13-identity-links-manual-correction.md`、新規 `reports/2026-07-31-task-022-identity-link-closeout.md`。
- **Reference paths（read-only）**: `backend/docs/api.yaml`、`backend/internal/apicontract/openapi_route_drift_test.go`、`backend/migrations/001_init.sql`、`frontend/src/features/identity-links/`、`q&a.html#dec-46`、Issue #239 本文/Phase 0 comment。氏名・メール・電話・診療内容・credential を report/audit へ記録しない。
- **依存**: USER が本起票 run の `claim/TASK-022` を統合後に解放すること。代表2医院データ、named operator、named operational/clinical signer、実 application role の RLS 証跡と、必要なら外部 Issue state の再整理は USER 境界。migration/reset は実行しない。
- **実行契約**: 独立 worktree・single writer。上記 claim の USER 解放後、実装者は `git branch --list 'claim/TASK-022'` が空であることを確認し、`git branch claim/TASK-022` を取得してから最初の編集を行う。開始時に現 HEAD を再測定し、別 branch にのみある修正を重複実装しない。
- **検証コマンド**:
  - `docker compose exec -T backend go test -p 1 ./internal/identitylink ./internal/apicontract -count=1`
  - `git diff --check -- backend/internal/identitylink docs/spec/screens/40-identity-links.md docs/ops/testing/scenarios/S13-identity-links-manual-correction.md reports/2026-07-31-task-022-identity-link-closeout.md`
  - 編集前: `bash scripts/check-docs-symbol-drift.sh > /tmp/task-022-docs-drift.before 2>&1 || test $? -eq 1`
  - 編集前: `grep '^FAIL  ' /tmp/task-022-docs-drift.before | LC_ALL=C sort > /tmp/task-022-docs-drift.before.failures`
  - 編集後: `bash scripts/check-docs-symbol-drift.sh > /tmp/task-022-docs-drift.after 2>&1 || test $? -eq 1`
  - 編集後: `grep '^FAIL  ' /tmp/task-022-docs-drift.after | LC_ALL=C sort > /tmp/task-022-docs-drift.after.failures`
  - `diff -u /tmp/task-022-docs-drift.before.failures /tmp/task-022-docs-drift.after.failures`
- **完了条件**: source-level で閉じた OpenAPI parity regression を維持し、全親医院認可の exact regression が green。hidden/mixed/no-audit failure で write/audit が0件、競合時に partial graph が0件。named operator が代表2医院 owner+pet の link → scoped history → unlink → relink を1回実施し、named signer が回復性・工程数/時間・誤 link・追加 capability の要否を承認する。RLS runtime が未実測なら本 task は UNREPORTED のままにし、Phase 2 へ進まない。

### TASK-023: #254 5業務フロー UAT 統合証跡（High）

- **対応 Issue**: GitHub Issue #254。
- **問題**: demo stack は起動できるが、既存証跡は部分実施で、scenario 本文に `【要実測】` が残る。authenticated E2E の host credential は未設定で、全5フローの PASS/FAIL、目視、DB/audit、実機 LINE/LIFF、FAIL 処置が1つの受け入れ証跡へ統合されていない。
- **現況の実測根拠**: `docs/ops/testing/scenarios/reports/2026-07-28-local.md` は prior failure の再確認に限定し、`reports/2026-07-31-task-010-batch2.md` が current residual set を記録する。`frontend/e2e/README.md:93-103` は `E2E_LOGIN_*` の env 注入を要求し、TASK-020 は未設定/BLOCKED。自動 E2E は目視・DB persistence・実機通知の代替ではない。
- **要求挙動**: 既存 asset を再利用し、①外来/検査/会計/締め ②予約→受付→再予約 ③trimming＋診察会計 ④LINE予約→記録 ⑤月次集計/report を1回通す。各 step に PASS/FAIL/BLOCKED、executor、環境、秘密/患者情報を含まない evidence、owner/disposition を記録し、agent 観測と human 観測を分離する。未承認 GitHub 書込は行わない。
- **Owned paths**: 新規 `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md` のみ。
- **Reference paths（read-only）**: `docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md`、`docs/ops/testing/scenarios/README.md`、同配下 S01〜S12/V01〜V05 と既存 reports、`frontend/e2e/`、`todo.md` の TASK-009/010/020、`reports/2026-07-31-task-010-batch2.md`、`reports/2026-07-31-task-010-runtime-batch.md`（検出済み credential 値を読取・転記しない）。
- **依存**: USER が本起票 run の `claim/TASK-023` を統合後に解放すること。既存 credential の redaction・到達環境確認・必要な失効/rotation、secret channel からの `E2E_LOGIN_*` 注入、必要時の local reset/seed、scenario 指定 DB/audit 観測、実機 LINE/LIFF、delivery owner の FAIL 処置は USER 境界。TASK-010/TASK-020 の owned files は変更しない。
- **実行契約**: 独立 worktree・single writer。authoring claim の USER 解放後、`git branch --list 'claim/TASK-023'` が空であることを確認し、`git branch claim/TASK-023` を取得してから report を作る。scenario/E2E の不具合候補は report 内で owner/disposition を付け、外部 Issue 化は別の明示承認を待つ。
- **検証コマンド**:
  - `docker compose ps`
  - `curl -fsS http://127.0.0.1:8080/health`
  - `curl -fsS -o /dev/null http://127.0.0.1:3003/`
  - `test -n "${E2E_LOGIN_EMAIL:-}" && test -n "${E2E_LOGIN_PASSWORD:-}"`
  - `cd frontend && ./scripts/run-e2e.sh e2e/clinical-flows.spec.ts e2e/examinations-flow.spec.ts e2e/accounting-flow.spec.ts e2e/reservations-smoke.spec.ts e2e/trimming-flow.spec.ts e2e/line-reservation-flow.spec.ts`
  - `git diff --check -- docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md`
- **完了条件**: 5フローと対応 scenario ID が1 report で追跡でき、全 step に状態・executor・環境・evidence がある。未waive FAIL は0件とし、残る FAIL は named owner・処置と delivery owner が承認した waiver/post-delivery disposition を持つ。human-only の DB/audit・LINE/LIFF・操作性 sign-off が揃わない限り #254 を完了扱いしない。credential/account/password/token/session 値は report・diff・commit に0件。

### TASK-024: #256 現行 screenshot / FAQ finalization（Medium）

- **対応 Issue**: GitHub Issue #256。
- **問題**: static manual text は同期済みだが、主要10画面の numbered screenshot は現行 UI との一致が未検証。FAQ は #254 の実測前には対象を確定できず、推測で増やすと誤案内と二重資産を作る。
- **現況の実測根拠**: 主要10件の `frontend/src/features/manual/content/screens/` 文書は同番号の既存 image を参照するが、`frontend/e2e/manual-flow.spec.ts` は navigation/content を検証するだけで画像の現行性を検証しない。`docs/delivery/OPERATION_MANUAL.md` には training plan draft があり、別の owner inquiry FAQ も既存のため新設しない。
- **要求挙動**: 10画像を現行 UI と比較し、不一致だけを同名で置換する。FAQ は TASK-023 が実測した staff の混乱だけを `workflows/10-troubleshooting.md` へ追加し、該当0件なら「追記不要」を監査 report に記録する。orphan `_screenshot-*`、別 FAQ、新固定値、秘密/患者情報を作らない。named documentation owner が画像と文言を1回で承認し、判断と署名を report に残す。
- **Owned paths**: `frontend/src/features/manual/content/workflows/10-troubleshooting.md`、`frontend/src/features/manual/content/images/02-reception.png`、`frontend/src/features/manual/content/images/04-medical-records.png`、`frontend/src/features/manual/content/images/05-accounting.png`、`frontend/src/features/manual/content/images/06-reservations.png`、`frontend/src/features/manual/content/images/07-examinations.png`、`frontend/src/features/manual/content/images/10-trimming.png`、`frontend/src/features/manual/content/images/13-cash-register.png`、`frontend/src/features/manual/content/images/14-accounting-reports.png`、`frontend/src/features/manual/content/images/16-line-reservation.png`、`frontend/src/features/manual/content/images/19-aggregation.png`、新規 `reports/2026-07-31-task-024-manual-audit.md`。
- **Reference paths（read-only）**: `frontend/src/features/manual/content/screens/02-reception.md`、`frontend/src/features/manual/content/screens/04-medical-records.md`、`frontend/src/features/manual/content/screens/05-accounting.md`、`frontend/src/features/manual/content/screens/06-reservations.md`、`frontend/src/features/manual/content/screens/07-examinations.md`、`frontend/src/features/manual/content/screens/10-trimming.md`、`frontend/src/features/manual/content/screens/13-cash-register.md`、`frontend/src/features/manual/content/screens/14-accounting-reports.md`、`frontend/src/features/manual/content/screens/16-line-reservation.md`、`frontend/src/features/manual/content/screens/19-aggregation.md`、`docs/delivery/OPERATION_MANUAL.md`、TASK-023 の final report、`frontend/e2e/manual-flow.spec.ts`、`frontend/src/features/manual/CLAUDE.md`。
- **依存**: USER が本起票 run の `claim/TASK-024` を統合後に解放すること。screenshot audit は先行可能だが、FAQ 編集と #256 完了は TASK-023 の confusion manifest 後。named documentation owner の目視 sign-off が必要。
- **実行契約**: 独立 worktree・single writer。上記 claim の USER 解放後、`git branch --list 'claim/TASK-024'` が空であることを確認し、`git branch claim/TASK-024` を取得してから編集する。Owned paths 外の manual text/画像は触らず、他 writer の変更を revert しない。
- **検証コマンド**:
  - `docker compose exec -T frontend npx vitest run src/features/manual/api/get-manual-articles.test.tsx src/features/manual/components/ManualSidebar.test.tsx src/features/manual/components/manual-content.test.ts src/features/manual/lib/parse-frontmatter.test.ts src/features/manual/routes/ManualPage.test.tsx`
  - `cd frontend && ./scripts/run-e2e.sh e2e/manual-flow.spec.ts`
  - `git diff --check -- frontend/src/features/manual/content/workflows/10-troubleshooting.md frontend/src/features/manual/content/images/02-reception.png frontend/src/features/manual/content/images/04-medical-records.png frontend/src/features/manual/content/images/05-accounting.png frontend/src/features/manual/content/images/06-reservations.png frontend/src/features/manual/content/images/07-examinations.png frontend/src/features/manual/content/images/10-trimming.png frontend/src/features/manual/content/images/13-cash-register.png frontend/src/features/manual/content/images/14-accounting-reports.png frontend/src/features/manual/content/images/16-line-reservation.png frontend/src/features/manual/content/images/19-aggregation.png reports/2026-07-31-task-024-manual-audit.md`
- **完了条件**: `reports/2026-07-31-task-024-manual-audit.md` に10/10画像の current/replace 判定、TASK-023 confusion 全件の既存案内/FAQ追記/追記不要 disposition、named documentation owner の visual/content sign-off がある。置換画像は current UI・1280〜1920px・mask済み・同名参照維持、speculative FAQ は0件で、targeted tests/E2E が green。

---
