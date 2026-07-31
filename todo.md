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
| SCEN-SEED-001 | 003_demo clinical CSV ヘッダのみ | **TASK-009**（CSV slice1 done・static verify GREEN・**適用は USER**） |
| SCEN-BROWSER-001 | scenarios 内【要実測】backlog | **TASK-010**（env READY・batch2 V04 partial・body 要実測 59） |
| SCEN-OPS-CLAIM-001 | `claim/*` 解放 | **ops-only**（USER only） |
| SCEN-OPS-COMMIT-001 | mixed commit 説明メモ | **ops-only**（rewrite しない） |
| SCEN-OPS-TREE-001 | 共有 tree concurrent WIP | **ops-only**（= R6） |
| ARCH-R2 | empty-diff COMPLETE 規律 | **ops-only**（継続） |
| ARCH-R3 | land 時 foreign 定義は `git status` 実測 | **ops note** + TASK-004 |
| POST-PULL | migrations 適用 | **ops-only** ≡ **SPEC-TOP-MIGRATE-006**（USER `make migrate`） |
| SPEC-TOP-LINE-AUDIT | `docs/spec/line/**` deep 監査 | **TASK-019 done** + **PO FINAL**（R-01 docs/test + R-06/R-07 parent RBAC landed `a1abd4db8`; R-05 Phase A+B code done / rollout+DROP HOLD） |
| SPEC-TOP-E2E-RUNTIME-84 | Playwright runtime 84 | **TASK-020**（env-forward done・runtime credentials BLOCKED） |
| SPEC-TOP-CAPABILITIES-CRUD | exclusion 面の破壊削除 | **TASK-021 Stage A**（Phase1 done; Phase2 slice1+slice2 complete; FE ZERO_IN_REPO / external UNREPORTED; CLEAN-GO/DROP HOLD） |
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

1. **TASK-009** seed 適用（USER。static green: `reports/2026-08-01-task-009-verify-seed-green.md` / reseed: `reports/2026-07-31-task-009-reseed-ops.md`）  
2. **TASK-010** 要実測残 backlog（body 59。batch2: `reports/2026-07-31-task-010-batch2.md`）  
3. **TASK-020** Playwright 93 runtime 完走（env-forward 済・要 host `E2E_LOGIN_*`。証拠: `reports/2026-07-31-task-020-env-forward.md`）  
4. **TASK-022 human residual** — S13 手動 correction + named signer + RLS runtime（agent source closeout 済）  
5. **TASK-023 human residual** — E2E_LOGIN_* 注入・5フロー通し・DB/audit・LINE/LIFF・sign-off（agent 証跡骨格 済）  
6. **TASK-024 human residual** — named documentation owner visual sign-off（agent 10/10 audit + FAQ 追記不要 済）  
7. **TASK-021 Stage A 削除**（Phase2 slice2: `reports/2026-07-31-task-021-phase2-slice2.md` — in-repo FE ZERO_IN_REPO; staff write の `excluded_type_ids` reject; exclusion routes/table KEEP。external use UNREPORTED。CLEAN-GO/DROP は USER 承認後のみ）
8. **TASK-004 / TASK-005**: 次の intentional land 時
9. **LINE follow-up（PO FINAL 済）**: `reports/2026-07-31-line-residual-po-decisions-FINAL.md`
   - High R-05 single-SoT Phase A+B — verifier cutover + reservation secret write path 撤去 done（`reports/2026-07-31-r05-single-sot-phase-a.md` / `reports/2026-07-31-r05-single-sot-phase-b.md`）。production rollout gates + column DROP は HOLD
   - R-06/R-07 parent RBAC honesty — landed `a1abd4db8`（runtime-green 未主張）
   - R-01 architecture summary + contract tests — landed `a1abd4db8`（runtime-green 未主張）
   - R-02/R-04/R-08 は ops のまま

---

## 個別タスク詳細

### TASK-004: screens-drift 意図変更セットのコミット隔離（Medium・ops）

- **問題**: intentional と foreign を同一 commit に混ぜない。foreign 定義は land 直前の `git status` / `git diff` が正本。
- **修正方針**: land 直前に porcelain 実測 → path-scoped `git add`（`git add -A` 禁止）。foreign は触らない・捨てない。
- **受け入れ条件**: staged ⊆ intentional; foreign 非 stage; 破棄しない。
- **状態**: **ops 手順 open**（再発・次 land 用）。前回実測: `reports/2026-07-31-task-004-005-land-proc.md`。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: ops-only
- **Blockers (today)**: checklist 自体は開始可能。実 land は intentional path set と commit owner が未確定で、shared tree に foreign WIP があるため未開始。
- **Preconditions**: land owner が `git status --porcelain=v1` と staged/unstaged diff を再実測し、intentional allowlist を明記する。`claim/TASK-004` が新たに存在する場合は編集を止める。
- **Code anchors** (file:line or path globs from live tree): `reports/2026-07-31-task-004-005-land-proc.md:39-47`, `scripts/check-docs-symbol-drift.sh`, `git status` / staged diff。
- **Convention anchors** (rule doc paths): `AGENTS.md` packet claim protocol、`.claude/rules/git-worktree-safety.md`、`.claude/CLAUDE.md` scoped verification / prohibited commands。
- **Steps**:
  1. land 直前の status と staged/unstaged path を採取し、packet owner が intentional / foreign を分類する。
  2. intentional path のみを path-scoped stage し、cached path が allowlist の部分集合であることを確認する。
  3. foreign WIP が unstaged のまま保存されていることを確認し、commit は USER の明示承認後だけ行う。
- **Verification** (scoped only):
  - `git status --short --branch`
  - `git status --porcelain=v1`
  - `git diff --name-status && git diff --cached --name-status`
- **Non-actions / HOLD**: `git add -A`、foreign WIP の stage/discard、history rewrite、force-push、claim 削除、製品コード変更は行わない。
- **Exit criteria for close**: 対象 land ごとに staged ⊆ intentional、foreign 非 stage、破棄なしを verbatim evidence で示し、USER が commit/統合結果を確認する。
- **Evidence sources read**: `todo.md`, `reports/2026-07-31-task-004-005-land-proc.md`, current `git status`, current claim list。

### TASK-005: closed packs 回帰のコミット前検証ゲート（Medium・ops）

- **問題**: land 前に doc/code 整合と inventory / hospitalization を機械確認する手順。
- **修正方針**: land 直前: `bash scripts/check-docs-symbol-drift.sh`; scoped hospitalization / route-inventory tests。結果は reports に記録。
- **受け入れ条件**: ゲート PASS; inventory 84 維持; hospitalization unit PASS。
- **状態**: **ops 手順 open**（land 都度）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: ops-only
- **Blockers (today)**: scoped gate は定義済み。実行対象となる intentional staged set と land window は未確定。
- **Preconditions**: TASK-004 の path 分離を完了し、Docker stack が利用可能で、foreign WIP が staged されていないことを確認する。
- **Code anchors** (file:line or path globs from live tree): `scripts/check-docs-symbol-drift.sh:28-34`, `scripts/check-docs-symbol-drift.test.sh:16-19`, `frontend/src/app/routes/route-inventory.test.tsx:45-56`, `backend/internal/medicalrecord/routes_snapshot_test.go:80-118`。
- **Convention anchors** (rule doc paths): `.claude/CLAUDE.md` scoped verification exception、`frontend/CLAUDE.md` scoped Vitest rule、`backend/CLAUDE.md` Docker-only rule。
- **Steps**:
  1. docs symbol drift とその self-test を実行し、exit code と出力を land note に残す。
  2. route inventory 84、hospitalization route/service、必要な LINE R6/R7 tests を対象ファイルだけで再実行する。
  3. gate 後に staged path が増えていないことを再確認し、FAIL は修正 packet へ戻す。
- **Verification** (scoped only):
  - `bash scripts/check-docs-symbol-drift.sh && bash scripts/check-docs-symbol-drift.test.sh`
  - `docker compose exec -T frontend npx vitest run src/app/routes/route-inventory.test.tsx`
  - `docker compose exec -T backend go test ./internal/medicalrecord/ -run 'Hospitalization|TestRegisterRoutes_Snapshot' -count=1`
  - `docker compose exec -T frontend npx vitest run src/features/hospitalization`
- **Non-actions / HOLD**: full test/lint/build、DB reset/migrate、dependency install、commit、claim 削除は自動実行しない。gate PASS を runtime 全体 green と呼ばない。
- **Exit criteria for close**: land 対象の docs drift、inventory 84、hospitalization scoped gate が PASS し、結果と対象 SHA/path が report に記録される。
- **Evidence sources read**: `reports/2026-07-31-task-004-005-land-proc.md`, live scripts/tests, `AGENTS.md`, `.claude/CLAUDE.md`。

### TASK-009: 003_demo clinical CSV ヘッダのみ — seed 再投入（High）

- **問題**: clinical CSV がヘッダのみでシナリオ前提データが揃わない可能性。
- **修正方針**: 設計 `reports/2026-07-31-task-009-seed-design.md` に従い USER が seed 適用。エージェントは migrate/seed auto-apply しない。
- **受け入れ条件**: 対象 CSV がヘッダのみでなくなる; シナリオ前提を満たす; 適用手順が1箇所で辿れる; 適用は USER。
- **状態**: **CSV slice1 committed（authoring done）/ static verifier GREEN（2026-08-01）/ 適用は USER**。slice1: hospitalizations + treatment_plans + daily_records + care_plan_items（G1 medical_records は既存 dump で充足）。証拠: `reports/2026-07-31-task-009-slice1.md`。**static gate**: `python3 scripts/verify_seed.py` → OK（imported clinical graph 認識: empty treatments + large medical_records; high-id appointments の business-hours 免除; RV/他院 placeholder allowlist）。証跡: `reports/2026-08-01-task-009-verify-seed-green.md`。**USER reseed 手順**: `reports/2026-07-31-task-009-reseed-ops.md`（既適用 DB は checksum mismatch → `make reset` が正。agent は auto wipe しない）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER（static green 後の DB 適用のみ残る）
- **Owner lane**: USER（apply）/ agent は static gate のみ
- **Blockers (today)**: DB 適用証跡なし。static `verify_seed.py` は 2026-08-01 に green（imported-graph 適応）。
- **Preconditions**: USER が local DB の適用履歴、退避要否、reset のデータ損失受容を確認する。static は `python3 scripts/verify_seed.py` で exit 0 を再確認。
- **Code anchors** (file:line or path globs from live tree): `backend/migrations/seeds/003_demo/{hospitalizations,treatment_plans,daily_records,care_plan_items}.csv`, `backend/migrations/seeds/003_demo/manifest.json`, `backend/cmd/migrate/main.go:41-136`, `backend/cmd/migrate/csvbundle.go:81-188`, `backend/internal/seedbundle/manifest.go:34-55`。
- **Convention anchors** (rule doc paths): `backend/migrations/CLAUDE.md`, `.agents/skills/migration-seed-safety/SKILL.md`, `reports/2026-07-31-task-009-reseed-ops.md`。
- **Steps**:
  1. 4 CSV の行数と manifest 掲載を再確認し、`verify_seed.py` の全 failure を slice1 内/外へ分類する。
  2. slice1 外の failure は別 claim/packet で seed-export 正規経路から修正し、CSV を手編集せず static verifier を green にする。
  3. verifier green 後、USER が未適用 DB なら `make migrate`、既適用 local DB なら reseed runbook に従い破壊性を確認して `make reset` を選ぶ。
  4. USER 適用後に入院ボード、入院詳細、カルテ一覧の smoke evidence を記録する。
- **Verification** (scoped only):
  - `wc -l backend/migrations/seeds/003_demo/{hospitalizations,treatment_plans,daily_records,care_plan_items}.csv`
  - `python3 scripts/verify_seed.py`
  - USER only: `make migrate` または、local のデータ損失を明示受容した場合だけ `make reset`
- **Non-actions / HOLD**: agent による migrate/seed/reset/DB_RESET/direct psql、STG/PROD seed 操作、CSV 手編集、claim 削除を行わない。現行 static RED のまま適用完了を主張しない。
- **Exit criteria for close**: static verifier exit 0、USER apply 証跡、対象4 CSVのDB反映とシナリオ前提 smoke が揃う。
- **Evidence sources read**: `reports/2026-07-31-task-009-seed-design.md`, `reports/2026-07-31-task-009-slice1.md`, `reports/2026-07-31-task-009-reseed-ops.md`, live seed tree, current verifier output, commit `c286bfe0a`。

### TASK-010: scenarios【要実測】一括実測バックログ（Medium）

- **問題**: scenarios に【要実測】残存。
- **修正方針**: browser-test レーンで実測。記録は `reports/`。
- **受け入れ条件**: 要実測 0 または PO/BUG 振分; reports に実行記録。
- **状態**: **env READY / batch2 partial**（2026-07-31 next orch wave）。docker healthy + `:8080/health` 200 + `:3003/` 200。batch1 V05: 5 件（証拠: `reports/2026-07-31-task-010-runtime-batch.md`）。**batch2 V04**: 6 件 elevate（要実測 body **65→59** / V04 11→5）。証拠: `reports/2026-07-31-task-010-batch2.md`。残 backlog open。claim: `claim/TASK-010`（USER 解放）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: BLOCKED(`claim/TASK-010`・seed/runtime 前提)
- **Owner lane**: AGENT→USER
- **Blockers (today)**: live claim が存在する。exact `【要実測】` は59件、S08/S09の装飾variantを含む semantic prefix `【要実測` は62件。seed依存シナリオはTASK-009 apply未了で false negative の恐れがある。
- **Preconditions**: current claim owner が継続するか、統合/放棄後にUSERがclaimを解放する。実行windowで health と必要なseed/role/credentialを値非出力で確認する。
- **Code anchors** (file:line or path globs from live tree): `docs/ops/testing/scenarios/{V*.md,S*.md}`, `docs/ops/testing/scenarios/README.md:46-52`, `reports/2026-07-31-task-010-runtime-batch.md`, `reports/2026-07-31-task-010-batch2.md`。
- **Convention anchors** (rule doc paths): `.agents/skills/browser-test/SKILL.md`, `docs/ops/testing/CLAUDE.md`, `.claude/CLAUDE.md` security/runtime evidence boundary。
- **Steps**:
  1. exact marker 59 と semantic prefix 62 を同じscopeで再採取し、S08/S09の3 decorated marksを別欄で固定する。
  2. 非seed依存 batch を先行し、seed依存 batch はTASK-009 USER apply後へ分け、browser-test laneで小分け実測する。
  3. 各 mark を PASS/FAIL/BLOCKED/DEFER と証拠へ置換し、FAILはBUG、仕様不明は要POへID付きで振り分ける。
  4. 一時データを安全に後始末し、batch report と残数を更新する。
- **Verification** (scoped only):
  - `rg -n '【要実測】' docs/ops/testing/scenarios/V*.md docs/ops/testing/scenarios/S*.md`
  - `rg -n '【要実測' docs/ops/testing/scenarios/V*.md docs/ops/testing/scenarios/S*.md`
  - `docker compose ps`
- **Non-actions / HOLD**: claim を無視した編集、証拠なしのmark昇格、secret/PII記録、full E2E、DB reset/migrate、本番LINE操作、claim削除を行わない。
- **Exit criteria for close**: exact/semantic両censusが0、または残り全件がID付きPO/BUG/明示BLOCKEDへ振分済みで、全batchの実行記録が存在する。
- **Evidence sources read**: `todo.md`, two TASK-010 reports, live scenario files, current exact/prefix census, current claim list。

### TASK-019: docs/spec/line/** deep 監査 follow-up（Medium / 任意）

- **問題**: line 仕様 vs 実装の deep 突合が partial のまま。
- **根拠**: 初回記録 `reports/2026-07-31-task-019-line-audit.md`。
- **修正方針**: deep pass で差分を docs/BUG/要PO/ops に振分。秘密・本番 webhook 操作は対象外。
- **受け入れ条件**: deep 結果1回記録; 新規 open は ID 付きまたは残差なし。
- **状態**: **done**（deep: `reports/2026-07-31-task-019-line-deep-audit.md`）。**PO FINAL**: `reports/2026-07-31-line-residual-po-decisions-FINAL.md`（`3d448ec5e`）。**R-01** binding B（code/tests SoT + architecture 要約）— docs/test landed `a1abd4db8`（runtime-green 未主張）。**R-05** binding A-CI（SoT=`clinic_integrations`）— Phase A verifier cutover + **Phase B reservation secret write/read 撤去 done**（`reports/2026-07-31-r05-single-sot-phase-a.md` / `reports/2026-07-31-r05-single-sot-phase-b.md`）；production rollout gates + column DROP HOLD。**R-06/R-07** original child residual close; parent-container / `/lstep` wrapper honesty landed `a1abd4db8`（runtime-green 未主張）。**R-02/R-04/R-08** ops、**R-03**→TASK-010。claim: `claim/LINE-R-FIX` / `claim/LINE-PO-R01-R05` / `claim/LINE-PARENT-RBAC` / `claim/R-05-SOT` / `claim/R-05-PHASE-B`（USER 解放）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: ops-only
- **Blockers (today)**: R-02/R-04/R-08はproduction/deploy evidence待ち。R-05はlegacy presence inventory未実施でrollout/DROP HOLD。LINE系live claim 3本が残る。
- **Preconditions**: ops owner、対象環境、secret-safe evidence channelを指定する。agent editが必要ならlive claimsのownerを確認し、値/hash/暗号文を成果物へ残さない。
- **Code anchors** (file:line or path globs from live tree): `backend/internal/lstep/line_link_service.go:295-447`, `backend/cmd/api/composition_reservation_test.go:96-130`, `backend/internal/model/line_reservation_setting.go:32-36`, `frontend/src/components/shared/Layout/{sidebar-menu.tsx,SidebarItems.tsx}`, `frontend/liff/src/lib/liff-config.ts`, `frontend/line-reserve/src/App.tsx:61-77`。
- **Convention anchors** (rule doc paths): `docs/spec/line/CLAUDE.md`, `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`, `docs/product-philosophy.md`, `.claude/CLAUDE.md` secret/production boundaries。
- **Steps**:
  1. R-01/R-06/R-07はlanded sourceをscoped testsで再確認し、runtime-greenとは分離して記録する。Phase B reportのcomposition residualは`fac8c86b2`で解消済みとして扱う。
  2. USER/opsがR-02 webhook/provisioning、R-04 Write API再有効化、R-08 LIFF ID一致をrunbook順で実測する。
  3. R-05は値を出さずclinicごとの`empty/equal/mismatch`だけinventoryし、presence/mismatchゼロとruntime evidence後にのみ別packetでDROPを提案する。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/lstep ./internal/reservation ./cmd/api -run 'TestLineLinkService_HandleWebhook|TestVerifySignatureAnyClinic|TestLineReservationSetting|TestNewReservationComposition_InjectsLineCredentialCipherClosures' -count=1`
  - `docker compose exec -T frontend npx vitest run src/components/shared/Layout/sidebar-menu.lstep-nav.test.tsx src/components/shared/Layout/SidebarItems.test.tsx src/app/routes/settings-routes.lstep-tags.test.tsx src/app/routes/operations-routes.lstep-delivery-monitor.test.tsx`
- **Non-actions / HOLD**: production LINE/L-step実送信、credential読取/記録、DB直接操作、`make migrate`、DROP author/apply、rollout、claim削除をagentは行わない。
- **Exit criteria for close**: R-02/R-04/R-08のnamed ops evidence、R-05 inventory/runtime gate、R-01/R-06/R-07 scoped evidenceが揃う。column DROPは別承認・別packet・USER applyまでcloseしない。
- **Evidence sources read**: `reports/2026-07-31-task-019-line-deep-audit.md`, `reports/2026-07-31-line-residual-po-decisions-FINAL.md`, R-05 Phase A/B reports, live code/tests, commits `a1abd4db8` / `fac8c86b2`。

### TASK-020: ui-design-compliance Playwright 再 runtime（84）（Low / 任意）

- **問題**: inventory 84 静的更新後の full runtime 未実施。
- **修正方針**: env 可なら `ui-design-compliance-readonly.spec.ts` workers=1。結果を reports へ。
- **状態**: **env-forward done / runtime credentials BLOCKED**（2026-07-31 next orch）。`run-e2e.sh` が host に設定時のみ `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`（+ optional `E2E_AUTH_STATE_PATH`）を name-only `-e` で Playwright docker へ転送。証拠: `reports/2026-07-31-task-020-env-forward.md`。prior runtime: 4p/3f/86 DNR（`reports/2026-07-31-task-020-runtime.md`）。host が EMAIL_UNSET/PASSWORD_UNSET のため再 runtime 未実施。full green 未達。claim: `claim/TASK-020` + `claim/W-020-ENV`（USER 解放）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: USER
- **Blockers (today)**: hostの認証env未注入。`claim/TASK-020` / `claim/W-020-ENV` がlive。前回は93 selected中4 passed / 3 failed / 86 did not runでfull green未達。
- **Preconditions**: USERがsecret channelからhost/CIへ必要envを注入し、値を出さず非空だけ確認する。claim holderまたはUSERが実行windowとevidence ownerを決める。
- **Code anchors** (file:line or path globs from live tree): `frontend/scripts/run-e2e.sh:28-46`, `frontend/e2e/helpers/auth.ts:4-17`, `frontend/playwright.config.ts:3-26`, `frontend/e2e/ui-design-compliance-readonly.spec.ts:447-646`, `frontend/src/app/routes/route-inventory.test.tsx:46-55`。
- **Convention anchors** (rule doc paths): `frontend/e2e/README.md`, `frontend/CLAUDE.md`, `.claude/CLAUDE.md` secret/Docker/scoped rules。
- **Steps**:
  1. USERが認証envを安全に注入し、値非表示のpreflightを通す。
  2. canonical `run-e2e.sh` で対象specだけをworkers=1で実行し、93 tests選択を確認する。
  3. passed/failed/did-not-run/blockedをreportへ記録し、FAILはBUG、環境要因はBLOCKEDへ振り分ける。
- **Verification** (scoped only):
  - `test -n "${E2E_LOGIN_EMAIL:-}" && test -n "${E2E_LOGIN_PASSWORD:-}"`
  - `cd frontend && ./scripts/run-e2e.sh e2e/ui-design-compliance-readonly.spec.ts --workers=1`
  - `rg -n 'E2E_LOGIN|DOCKER_ENV' frontend/scripts/run-e2e.sh`
- **Non-actions / HOLD**: credentialの生成・推測・表示・git保存、誤ったfull-suite runner、本番操作、claim削除、4 public/static passesだけでのgreen宣言を行わない。
- **Exit criteria for close**: canonical runnerが93 testsを選択し、全結果がreport化され、未実行/FAILが0またはID付き処分済みとなる。
- **Evidence sources read**: `reports/2026-07-31-task-020-runtime.md`, `reports/2026-07-31-task-020-env-forward.md`, live runner/auth/config/spec, current env name-only status, current claims。

### TASK-021 Stage A: exclusion 面の破壊的撤去（Medium・PO決裁済・inventory 済）

- **問題**: Stage B で facade 化済み。exclusion route/payload/model/table の最終撤去が残る。
- **修正方針**: **consumer inventory + 破壊変更の明示承認後**に Stage A（FINAL 参照）。新 endpoint は追加しない。`available-staffs` は WONTFILE。
- **受け入れ条件**: exclusion production surface 削除; migration あり; Stage B 互換 consumer が無いこと inventory で証明。
- **状態**: **Phase1 FE residual SAFE-CLEANUP done / Phase2 slice1+slice2 COMPLETE / in-repo FE ZERO_IN_REPO / external use UNREPORTED / CLEAN-GO·DROP·migrate HOLD**。Stage B: `e9dddd921`。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md` + LINE residual FINAL（021 Phase2）。inventory: `reports/2026-07-31-task-021-stage-a-inventory.md`。Phase1: `reports/2026-07-31-task-021-phase1-consumer-prep.md`。Phase2 slice1: `reports/2026-07-31-task-021-phase2-slice1.md`。Phase2 slice2: `reports/2026-07-31-task-021-phase2-slice2.md`（staff Create/Update で `excluded_type_ids` hard-reject；exclusion routes/table/response KEEP）。次は USER external inventory または破壊変更承認後の CLEAN-GO のみ。claim: `claim/TASK-021` + `claim/W-021-P1` + `claim/TASK-021-P2` + `claim/TASK-021-S2`（USER 解放）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: BLOCKED(external use UNREPORTED・破壊承認未取得・`claim/TASK-021`)
- **Owner lane**: AGENT→USER
- **Blockers (today)**: legacy endpoint外部利用ゼロが未証明。CLEAN-GO/DROP承認がなく、live claimsがある。route/OpenAPI/model/table/RLS/seed/export/test surfaceは現HEADで生存。
- **Preconditions**: USER/opsがaccess log、client registry、既知利用者でdeprecation終了を証明し、CLEAN-GOとDROP authorを別々に承認する。claim ownerを確定する。
- **Code anchors** (file:line or path globs from live tree): `backend/internal/staff/handler.go:209-212`, `backend/internal/staff/{staff_handler.go,staff_service_permissions.go,ports.go}`, `backend/internal/reservation/{reservation_staff_request.go,reservation_staff_service.go,reservation_staff_repository.go}`, `backend/internal/model/staff_reservation_exclusion.go`, `backend/docs/api.yaml:6544-6547`, `backend/migrations/001_init.sql:2655-2663`, `backend/cmd/seed-export/tables.go:40-47`。
- **Convention anchors** (rule doc paths): `reports/2026-07-31-line-residual-po-decisions-FINAL.md:129-147`, `backend/migrations/CLAUDE.md`, `.agents/skills/migration-seed-safety/SKILL.md`, `.claude/rules/go-gin-backend-guidelines.md`, `docs/product-philosophy.md`。
- **Steps**:
  1. USER evidenceで`GET|PUT /masters/staffs/:id/excluded-reservation-types`、`excluded_type_ids`、`excluded_courses`の外部利用ゼロ/deprecation終了を確定する。
  2. 承認後のagent packetでlegacy request/response/route/handler/service/port/model/OpenAPI/generated/tests/docsをcapabilities-onlyへ削除し、positive capability surfaceと`available-staffs` banを維持する。
  3. seed/exportからexclusion physical surfaceを外し、新規max+1 numbered migrationでtable/RLS DROPをauthorする。既存migrationは編集しない。
  4. inventoryとscoped testsをgreenにした後、USERだけが`make migrate`を実行して適用証跡を残す。
- **Verification** (scoped only):
  - `rg -n 'staff_reservation_exclusions|StaffReservationExclusion|ExcludedTypeIDs|excluded_type_ids|excluded_courses|excluded-reservation-types' backend frontend docs --glob '!**/node_modules/**'`
  - `docker compose exec -T backend go test -p 1 ./internal/reservation ./internal/staff ./internal/apicontract ./internal/model -run 'ReservationStaff|Capability|Excluded|AvailableStaffs|OpenAPI|RLS|Schema' -count=1`
  - `docker compose exec -T frontend npx vitest run src/hooks/use-reservation-types.test.ts src/components/shared/ReservationFormModal/filter-staff-candidates.test.ts`
- **Non-actions / HOLD**: UNREPORTEDのままCLEAN-GOしない。`001_init.sql`編集、DROP author/apply、seed/RLS削除、`make migrate`、`available-staffs`追加、維持対象route/capabilities削除、claim削除は承認前に行わない。
- **Exit criteria for close**: external-use証明、production exclusion surface zero、capabilities-only tests/contract、new migration、USER apply evidenceがすべて揃う。
- **Evidence sources read**: TASK-021 inventory/Phase1/Phase2 slice1/slice2 reports, PO FINAL reports, live source/migration/seed/export tree, reachable commits `e9dddd921` / `a06c12965` / `8a97a5696`。

### TASK-022: #239 Phase 1 closeout と代表手動 correction gate（High）

- **対応 Issue**: GitHub Issue #239（live state は CLOSED。未充足の受け入れ条件を local New Work として追跡）。
- **状態**: **agent source closeout done / human residual open**（2026-07-31）。`CreatePetGroup` の any-member fallback を除去し、親 owner-group の anchor + 全 active member clinic を actor に要求する regression を green（`go test -p 1 ./internal/identitylink ./internal/apicontract -count=1`）。Phase 2 未着手。証拠: `reports/2026-07-31-task-022-identity-link-closeout.md`、S13: `docs/ops/testing/scenarios/S13-identity-links-manual-correction.md`。security review PASS。
- **残 human**: named operator の 2 医院 link→history→unlink→relink 実施と named signer 承認；RLS runtime を実 application role で証明（未なら UNREPORTED のまま Phase 2 禁止）。
- **claim**: `claim/TASK-022`（USER が main 統合確認後に解放）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: USER
- **Blockers (today)**: named S13 operator/signer evidenceとreal application-role RLS runtime proofが未取得。Phase 2はBLOCKED。ledger記載の`claim/TASK-022`はlive branchに存在しない。
- **Preconditions**: non-production/local-STG data、2医院を扱えるnamed operator、named signer、RLS実測可能なOps/DBAと証跡様式を確保する。
- **Code anchors** (file:line or path globs from live tree): `backend/internal/identitylink/service.go:443-486`, `backend/internal/identitylink/service.go:864-882`, `backend/internal/identitylink/service_test.go:495-638`, `backend/internal/model/identity_link_rls_migration_test.go:10-13`, `backend/internal/persistence/{rls_effectiveness_test.go,rls_role_privilege_test.go}`。
- **Convention anchors** (rule doc paths): `.claude/refs/backend-application-invariants.md`, `backend/CLAUDE.md`, `docs/ops/testing/scenarios/S13-identity-links-manual-correction.md`。
- **Steps**:
  1. agent source closeoutのscoped regressionを再確認し、source proofとruntime proofを分離する。
  2. named operatorがS13のlink→history→unlink→relinkを2医院条件で実施し、named signerが署名する。
  3. Ops/DBAが実application roleでcross-clinic遮断を実測し、secret/PIIなしでrole・環境・期待・結果を記録する。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/identitylink ./internal/apicontract -count=1`
  - manual: `docs/ops/testing/scenarios/S13-identity-links-manual-correction.md` 手順1-8 + named sign-off
- **Non-actions / HOLD**: S13/RLS evidence前のPhase 2、auto-link/merge/candidate UI/DDL/record move、DB/migrate、Issue close、claim削除を行わない。
- **Exit criteria for close**: source regression PASS、S13 named operator実測、named signer承認、real application-role RLS proofの4点が揃う。
- **Evidence sources read**: `reports/2026-07-31-task-022-identity-link-closeout.md`, S13 scenario, live identitylink/RLS tests, live GitHub #239 state。

### TASK-023: #254 5業務フロー UAT 統合証跡（High）

- **対応 Issue**: GitHub Issue #254。
- **状態**: **agent 証跡骨格 done / human residual open**（2026-07-31）。統合 report: `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md`。env health PASS；`E2E_LOGIN_*` 未注入で authenticated E2E **BLOCKED**（EXIT:1、値非出力）；5 フロー step 全行に status/executor/env/evidence/owner；agent と human 欄分離；`confusion_count: 0`。#254 は human 揃うまで完了扱いしない。
- **残 human**: USER が secret channel で `E2E_LOGIN_*` 注入後 E2E 実行；QA が 5 フローブラウザ通し・DB/audit 目視・実 LINE/LIFF；PO/現場の使い勝手 sign-off と FAIL 処分承認。
- **claim**: `claim/TASK-023`（USER 解放）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: USER
- **Blockers (today)**: authenticated E2E credentials未注入、QA/LINE owner/POのhuman evidence未取得。ledger記載の`claim/TASK-023`はlive branchに存在しない。
- **Preconditions**: USER secret injection、QA lead、LINE setting owner、PO/現場責任者、非PII evidence場所と同一実行windowを確保する。
- **Code anchors** (file:line or path globs from live tree): `frontend/scripts/run-e2e.sh:30-46`, `frontend/e2e/helpers/auth.ts:8-17`, `frontend/e2e/{clinical-flows,examinations-flow,accounting-flow,reservations-smoke,trimming-flow,line-reservation-flow}.spec.ts`, `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md`。
- **Convention anchors** (rule doc paths): `frontend/e2e/README.md`, `docs/ops/testing/CLAUDE.md`, `.claude/CLAUDE.md` secret/runtime boundaries。
- **Steps**:
  1. USERが認証envをsecret channelから注入し、値を表示せずcanonical runnerを起動する。
  2. QAが5業務flowをブラウザで通し、DB/audit、実LINE/LIFFを担当ownerと確認する。
  3. PO/現場責任者がusabilityをsign-offし、FAILを納品前修正/後対応/棄却へ処分する。
- **Verification** (scoped only):
  - `cd frontend && ./scripts/run-e2e.sh e2e/clinical-flows.spec.ts e2e/examinations-flow.spec.ts e2e/accounting-flow.spec.ts e2e/reservations-smoke.spec.ts e2e/trimming-flow.spec.ts e2e/line-reservation-flow.spec.ts`
  - human checklist: `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md:98-211`
- **Non-actions / HOLD**: credentialsの生成/推測/表示、agentによるDB直参照・real LINE/LIFF操作、human欄の代筆、Issue #254 close、claim削除を行わない。
- **Exit criteria for close**: authenticated E2E、5 flow、DB/audit、real LINE/LIFF、PO/現場sign-offとFAIL処分が全てnamed evidenceで揃う。
- **Evidence sources read**: Issue #254 integrated report, live runner/auth helpers/specs, live GitHub #254 state。

### TASK-024: #256 現行 screenshot / FAQ finalization（Medium）

- **対応 Issue**: GitHub Issue #256。
- **状態**: **agent audit + FAQ disposition done / human visual sign-off open**（2026-07-31）。10/10 current/replace 判定。replace 7 枚のうち **4 枚のみ採用**（`02` / `06` / `13` / `14`）。`05` / `07` / `10` はフルシード由来の飼主氏名・ペット ID が写り込んだため受領検証で差し戻し（`254fdc2f3`）— 該当 3 画面の文書不一致は未解消で、クリーンシード環境での再撮影が必要。FAQ は TASK-023 `confusion_count: 0` に基づき **追記不要**（`10-troubleshooting.md` 変更なし）。証拠: `reports/2026-07-31-task-024-manual-audit.md`。vitest manual 18 tests PASS。manual-flow E2E は env 未注入で BLOCKED。
- **残 human**: named documentation owner の visual/content sign-off；任意で `19-aggregation` / `04-medical-records` 再撮影。
- **claim**: `claim/TASK-024`（USER 解放）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: USER
- **Blockers (today)**: clean-seed版`05`/`07`/`10`再撮影とnamed documentation owner sign-offが未取得。manual-flow E2Eはcredential待ち。ledger記載の`claim/TASK-024`はlive branchに存在しない。
- **Preconditions**: full seed/PII-bearing DBを使わないclean seed環境、承認済み撮影者、named documentation owner、必要ならE2E secret injectionを確保する。
- **Code anchors** (file:line or path globs from live tree): `frontend/src/features/manual/content/screens/{05-accounting,06-reservations,07-examinations,10-trimming,13-cash-register,14-accounting-reports,19-aggregation}.md`, `frontend/e2e/manual-flow.spec.ts`, manual feature tests。
- **Convention anchors** (rule doc paths): `frontend/src/features/manual/CLAUDE.md`, `frontend/CLAUDE.md`, `docs/product-philosophy.md`, `.claude/CLAUDE.md` privacy boundary。
- **Steps**:
  1. clean seed環境で`05-accounting`、`07-examinations`、`10-trimming`を再撮影し、PII/secret混入を受領検査する。
  2. 10画像を現行UI/本文参照と突合し、named documentation ownerがvisual/content sign-offする。
  3. credentials注入後にmanual scoped Vitest/E2Eを実行する。TASK-023で新規observed confusionがない限りFAQはno-addを維持する。
- **Verification** (scoped only):
  - `docker compose exec -T frontend npx vitest run src/features/manual/api/get-manual-articles.test.tsx src/features/manual/components/ManualSidebar.test.tsx src/features/manual/components/manual-content.test.ts src/features/manual/lib/parse-frontmatter.test.ts src/features/manual/routes/ManualPage.test.tsx`
  - `cd frontend && ./scripts/run-e2e.sh e2e/manual-flow.spec.ts`
- **Non-actions / HOLD**: PII-bearing screenshot採用、full seed撮影、推測FAQ追加、browser/credential/DB操作のagent代行、Issue #256 close、claim削除を行わない。
- **Exit criteria for close**: clean-seed 3枚再撮影、10/10 named visual sign-off、manual scoped tests/E2E evidence、FAQ no-add判断が揃う。
- **Evidence sources read**: `reports/2026-07-31-task-024-manual-audit.md`, manual screen refs/tests, TASK-023 confusion evidence, live GitHub #256 state。

---
