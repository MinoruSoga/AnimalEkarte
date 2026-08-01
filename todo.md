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
| SCEN-BROWSER-001 | scenarios 内【要実測】backlog | **TASK-010**（env READY・batch5 reclass+browser route・final census は batch5 report） |
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
| ISSUE-201-DOSE-LOOKUP | dose parameter 取得障害の silent fallback | **TASK-025**（READY_AGENT・臨床値は別 gate） |
| ISSUE-249-CONFIRMED-LOCK | confirmed 検査の更新/削除 lock・audit | **TASK-026**（DONE `2a8aca33c`・`main` 統合済み） |
| ISSUE-249-MANUAL-LIFECYCLE | 手動検査 edit / confirmed→completed 確定解除 | **TASK-027**（READY_AGENT） |
| ISSUE-252-STANDARD-PATCH | 締め設定 standard PATCH の validation/audit | **TASK-028**（DONE `bbf82e2b8`・値投入は USER） |
| ISSUE-259-DOC-CONTRACT | Lステップ disabled 時の旧 noop 文書 | **TASK-029**（DONE・`9fc5b9ffb` push 済。残は先方 enable + USER runtime 実測） |
| ISSUE-261-TRIMMING-DECEASED | trimming 死亡ペット拒否の経路別回帰 | **TASK-030**（DONE `6e5a945ef`・runtime は USER） |
| ISSUE-249-PRINT-SNAPSHOT | 検査結果の保存 snapshot 印刷 | **TASK-031**（READY_AGENT） |
| ISSUE-249-IMPORT-REVERT | lab import job の compensating revert | **TASK-032**（READY_AGENT・migration review 必須） |
| ISSUE-201-EMERGENCY-ADMIN | 構造化救急投薬記録と欠落時 fail-closed cutover | **TASK-033**（臨床承認・migration review 後） |

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

1. **TASK-009** seed 適用（USER。static green reconfirmed 2026-08-01: `python3 scripts/verify_seed.py` exit 0 / reseed: `reports/2026-07-31-task-009-reseed-ops.md`）
2. **TASK-010** 要実測残 backlog（final exact **38** / semantic **40**。`reports/2026-08-01-task-010-batch5.md`。次: DEFER/BLOCKED re-smoke / V01 after TASK-009 apply）
3. **TASK-020** Playwright 93 runtime 完走（env-forward 済・host `E2E_LOGIN_*` **UNSET** 再確認 2026-08-01。証拠: `reports/2026-07-31-task-020-env-forward.md`）
4. **TASK-022 human residual** — S13 手動 correction + named signer + RLS runtime（source regression re-green 2026-08-01）
5. **TASK-023 human residual** — E2E_LOGIN_* 注入・5フロー通し・DB/audit・LINE/LIFF・sign-off（agent 証跡骨格 済）
6. **TASK-024 human residual** — named documentation owner visual sign-off（manual vitest 18 PASS re-green 2026-08-01）
7. **TASK-021 Stage A 削除**（Phase2 slice2: `reports/2026-07-31-task-021-phase2-slice2.md` — in-repo FE ZERO_IN_REPO; staff write の `excluded_type_ids` reject; exclusion routes/table KEEP。external use UNREPORTED。CLEAN-GO/DROP は USER 承認後のみ）
8. **TASK-004 / TASK-005**: 次の intentional land 時（2026-08-01 freeze: clean tree → land gate **not triggered**）
9. **LINE follow-up（PO FINAL 済）**: `reports/2026-07-31-line-residual-po-decisions-FINAL.md`
   - High R-05 single-SoT Phase A+B — verifier cutover + reservation secret write path 撤去 done（`reports/2026-07-31-r05-single-sot-phase-a.md` / `reports/2026-07-31-r05-single-sot-phase-b.md`）。production rollout gates + column DROP は HOLD
   - R-06/R-07 parent RBAC honesty — landed `a1abd4db8`；scoped FE/BE re-green 2026-08-01（runtime-green 未主張）
   - R-01 architecture summary + contract tests — landed `a1abd4db8`；scoped BE re-green 2026-08-01（runtime-green 未主張）
   - R-02/R-04/R-08 は ops のまま

---

## Staged plan outcome — TASK-025 readiness dossier（2026-08-01）

- **Status**: **SCOPED_COMPLETE / GLOBAL_BLOCKED**。owned packet は docs-only で完成し独立 clinical/security review は PASS。共有 tree の foreign WIP により、prompt-wide の global allowlist と `todo.md` append-only predicate は BLOCKED。foreign WIP は変更・stage・破棄していない。
- **Changed files (owned)**: `reports/2026-08-01-issue-readiness-dossier.md`, `todo.md`, `q&a.html`, `3-session-agent.html`。
- **Runtime verification**: 製品コード・migration・seed・DB を変更していないため適用対象なし。Docker/full-project test は保存プロンプトにより実行していない。

### Gate evidence（verbatim）

1. Open Issue set / dossier schema

```text
$ gh issue list --state open --limit 200 --json number --jq '.[].number' | sort -n > /tmp/open.txt; grep -oE '^## Issue #[0-9]+' reports/2026-08-01-issue-readiness-dossier.md | grep -oE '[0-9]+' | sort -n > /tmp/dossier.txt; diff /tmp/open.txt /tmp/dossier.txt; echo "set_diff_exit=$?"; wc -l /tmp/open.txt /tmp/dossier.txt
set_diff_exit=0
      21 /tmp/open.txt
      21 /tmp/dossier.txt
      42 total
$ N=$(grep -c '^## Issue #' reports/2026-08-01-issue-readiness-dossier.md); for h in '現状実測' '残作業' '次に動くのは' '着手プラン' '回答起案'; do printf '%s=%s\n' "$h" "$(grep -c "^### $h" reports/2026-08-01-issue-readiness-dossier.md)"; done; printf 'N=%s\n' "$N"
現状実測=21
残作業=21
次に動くのは=21
着手プラン=21
回答起案=21
N=21
```

2. Decision-pack coverage

```text
$ targets=(89 97 98 99 201 211 212 235 249 250 252 253 254 255 256 257 258 259 260 261 284); for n in $targets; do printf '#%s=%s\n' "$n" "$(grep -c "#$n" 'q&a.html')"; done
#89=3
#97=4
#98=2
#99=2
#201=20
#211=13
#212=8
#235=8
#249=10
#250=2
#252=2
#253=2
#254=1
#255=5
#256=4
#257=1
#258=5
#259=2
#260=5
#261=10
#284=1
```

3. ID uniqueness / allocation

```text
$ grep -oE 'id="dec-[0-9]+"' 'q&a.html' | sort | uniq -d
$ grep -ohE '\bDEC-[0-9]+' 'q&a.html' 3-session-agent.html todo.md phase2.html | sed 's/DEC-//' | sort -n -u | tail -1
58
$ grep -ohE '\bTASK-[0-9]+' todo.md | sed 's/TASK-//' | sort -n -u | tail -1
033
```

4. Append-only

```text
$ git diff --numstat -- todo.md 'q&a.html'
276	0	q&a.html
426	39	todo.md
```

`q&a.html` は削除 0 で PASS。`todo.md` の削除 39 は clean baseline 後に出現した別 claim 所有の foreign WIP であり、global predicate は BLOCKED。owned hunks は索引行と HEAD 末尾以降の新規 TASK/outcome だけで削除 0。

5. View / HTML5 / duplicate IDs

```text
$ git diff -- 3-session-agent.html | grep '^+' | grep -nE '[0-9]+[[:space:]]*件|\b[0-9a-f]{7,40}\b'; echo "view_forbidden_exit=$?"
view_forbidden_exit=1
$ /opt/homebrew/bin/tidy -errors -quiet -utf8 'q&a.html'; echo "qa_exit=$?"; /opt/homebrew/bin/tidy -errors -quiet -utf8 3-session-agent.html; echo "view_exit=$?"
qa_exit=0
view_exit=0
$ grep -oE 'id="[^"]*"' 'q&a.html' | sort | uniq -d
$ grep -oE 'id="[^"]*"' 3-session-agent.html | sort | uniq -d
```

6. Sensitive-pattern scan

```text
$ grep -inE '(password|passwd|secret|token|api[_-]?key|AKIA[0-9A-Z]{16}|postgres://|mysql://|BEGIN [A-Z ]*PRIVATE KEY)' reports/2026-08-01-issue-readiness-dossier.md; echo "sensitive_exit=$?"
sensitive_exit=1
```

7. Scope / trackability

```text
$ git diff --name-only
3-session-agent.html
bug.md
docs/ops/testing/scenarios/S07-estimate-status-control.md
docs/ops/testing/scenarios/S08-accounting-corrections.md
docs/ops/testing/scenarios/S09-closing-time-boundaries.md
docs/ops/testing/scenarios/V02-accounting-reservation-forms.md
q&a.html
todo.md
$ git check-ignore -v reports/2026-08-01-issue-readiness-dossier.md; echo "ignored_exit=$?"
ignored_exit=1
$ git diff --cached --name-only
```

Global allowlist は foreign WIP のため BLOCKED。owned path set は上記 changed files 4 件だけ。最終 commit の staged paths と `git show --stat HEAD` は自己参照を避けるため本 pre-commit ledger へ埋め込まず Completion Report に逐語記録する。

8. Claims

```text
$ git branch --list 'claim/TASK-025'
[empty]
$ git branch claim/TASK-025; echo "exit=$?"
exit=0
$ for n in 025 026 027 028 029 030 031 032 033; do git branch --list "claim/TASK-$n"; done
  claim/TASK-025
  claim/TASK-026
  claim/TASK-027
  claim/TASK-028
  claim/TASK-029
  claim/TASK-030
  claim/TASK-031
  claim/TASK-032
  claim/TASK-033
```

9. Saved-prompt validator

```text
$ node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fast-issue-readiness-dossier.md
Prompt Craft Harness Validation: PASS
Profile: standard (declared-risk-tier)
Target: agent (source-path)
Quality mode: standard
Execution contract: dynamic-workflow/v1
validator_exit=0
```

### Failure Signature / deviations

- **FS-1 / attempt 1**: q&a coverage loop expected one count per Issue; zsh scalar target list produced one composite grep and zero. Root cause was zsh scalar non-splitting. Fix: explicit zsh array. Result: all 21 counts ≥ 1。
- **FS-2 / review repair 1**: independent healthcare review rejected finalized/free-text addendum as an emergency administration substitute. Fix: TASK-025 technical slice and TASK-033 structured active/draft event + atomic missing-data cutover. Result: clinical review PASS。
- **FS-3 / review repair 1**: inherited P0 rows contradicted DEC-48/58. Append-only constraint forbids rewriting them. Fix: immediately preceding `issue-readiness-current-p0-20260801` current-authority block. Result: security review PASS。
- **Assumption deviations**: live open set remained the generated 21 (difference none). Native Workflow tool was unavailable, so real multi-agent fan-out/review roles were used. Required referenced `docs/CODEX-NAVIGATION-GUIDE.md` was absent (harness P2). TASK-033 was added from independent clinical falsification. Global scope/append-only gates remain BLOCKED solely by preserved foreign WIP.

## 個別タスク詳細

### TASK-004: screens-drift 意図変更セットのコミット隔離（Medium・ops）

- **問題**: intentional と foreign を同一 commit に混ぜない。foreign 定義は land 直前の `git status` / `git diff` が正本。
- **修正方針**: land 直前に porcelain 実測 → path-scoped `git add`（`git add -A` 禁止）。foreign は触らない・捨てない。
- **受け入れ条件**: staged ⊆ intentional; foreign 非 stage; 破棄しない。
- **状態**: **ops 手順 open**（再発・次 land 用）。前回実測: `reports/2026-07-31-task-004-005-land-proc.md`。**2026-08-01 rebaseline**: intentional land set **なし**（`git status --porcelain` empty at freeze; post-freeze foreign `bug.md` WIP のみ・stage しない）→ gate **not triggered**。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: ops-only
- **Blockers (today)**: intentional land path set と commit owner が未確定。2026-08-01 freeze 時点では staged/intentional set 無し。
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
- **状態**: **ops 手順 open**（land 都度）。**2026-08-01 rebaseline**: intentional staged set なし → gate **not triggered**（docs-drift / inventory / hospitalization 未実行）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: ops-only
- **Blockers (today)**: scoped gate は定義済み。実行対象となる intentional staged set と land window は未確定（not triggered 維持）。
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
- **状態**: **CSV slice1 committed（authoring done）/ static verifier GREEN（2026-08-01 reconfirm exit 0）/ 適用は USER**。slice1: hospitalizations + treatment_plans + daily_records + care_plan_items（各 2 data rows / header+2 = wc 3; all in manifest）。証拠: `reports/2026-07-31-task-009-slice1.md`。**static gate reconfirm**: `python3 scripts/verify_seed.py` → OK exit 0（consultations=27 … medical_records=425544 …）。証跡: `reports/2026-08-01-task-009-verify-seed-green.md` + 本 session reconfirm。**USER reseed 手順**: `reports/2026-07-31-task-009-reseed-ops.md`（既適用 DB は checksum mismatch → `make reset` が正。agent は auto wipe しない）。**apply/smoke 証跡: なし** → close 不可。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER（static green 後の DB 適用のみ残る）
- **Owner lane**: USER（apply）/ agent は static gate のみ
- **Blockers (today)**: DB 適用証跡なし。static `verify_seed.py` は 2026-08-01 reconfirm green。
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
- **状態**: **env READY / batch5 RECLASSIFIED + partial browser route + disposition repair**（2026-08-01 follow-up）。health 200 + Chrome ノア/八王子。false PASS elevate 撤回。runtime PASS: L61/S08（partial）/S01-A3/S03-#7/S12-link。final census exact **38** / semantic **40**。batch5 12 = PASS×2 + DEFER/BLOCKED×10。V01×17 seed-gated。残 S* は BLOCKED/DEFER/FAIL 付き。証拠: `reports/2026-08-01-task-010-batch5.md`。claim: **`claim/TASK-010` held**。

#### 実装プラン（2026-08-01・codebase調査 + runtime-evidence follow-up）
- **Ready**: env READY。`claim/TASK-010` live。seed依存（V01）は TASK-009 USER apply 後。
- **Owner lane**: AGENT→USER
- **Blockers (today)**: exact **38** / semantic **40**。LIFF 401。受付 0 件。TASK-009 apply 不在。append-only close。
- **Preconditions**: health/demo を値非出力で確認。foreign `bug.md` 非 stage。source を runtime PASS に置換しない。
- **Code anchors**: `docs/ops/testing/scenarios/{V*.md,S*.md}`, `reports/2026-08-01-task-010-batch5.md`。
- **Convention anchors**: `.agents/skills/browser-test/SKILL.md`, `docs/ops/testing/CLAUDE.md`。
- **Steps**:
  1. census 再採取（現状 38 / 40）。
  2. DEFER/BLOCKED を fixture・LIFF mock・カード作成後 re-smoke。V01 は TASK-009 後。
  3. FAIL→BUG ID。source-only は PASS にしない。
  4. report 更新。
- **Verification** (scoped only):
  - `rg -n '【要実測】' docs/ops/testing/scenarios/V*.md docs/ops/testing/scenarios/S*.md`
  - `rg -n '【要実測' docs/ops/testing/scenarios/V*.md docs/ops/testing/scenarios/S*.md`
- **Non-actions / HOLD**: claim 無視、証拠なし昇格、secret 記録、full E2E、migrate/seed apply、claim 削除。
- **Exit criteria for close**: census 0 または残り全件が ID 付き PO/BUG/明示 BLOCKED で batch 記録揃う。
- **Evidence sources read**: todo, batch5 report, census 38/40, claims, Chrome。

### TASK-019: docs/spec/line/** deep 監査 follow-up（Medium / 任意）

- **問題**: line 仕様 vs 実装の deep 突合が partial のまま。
- **根拠**: 初回記録 `reports/2026-07-31-task-019-line-audit.md`。
- **修正方針**: deep pass で差分を docs/BUG/要PO/ops に振分。秘密・本番 webhook 操作は対象外。
- **受け入れ条件**: deep 結果1回記録; 新規 open は ID 付きまたは残差なし。
- **状態**: **done**（deep: `reports/2026-07-31-task-019-line-deep-audit.md`）。**PO FINAL**: `reports/2026-07-31-line-residual-po-decisions-FINAL.md`（`3d448ec5e`）。**R-01** binding B — docs/test landed `a1abd4db8`；**scoped BE re-green 2026-08-01**（`go test -p 1 ./internal/lstep ./internal/reservation ./cmd/api -run 'TestLineLinkService_HandleWebhook|…'` exit 0；runtime-green 未主張）。**R-05** Phase A+B code done / production inventory+rollout+DROP HOLD。**R-06/R-07** landed `a1abd4db8`；**scoped FE re-green 2026-08-01**（4 files / 13 tests PASS）。**R-02/R-04/R-08** ops、**R-03**→TASK-010。claim names historical: LINE claims **not live** on 2026-08-01 refs（stale 「USER 解放」prose only）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: ops-only
- **Blockers (today)**: R-02/R-04/R-08はproduction/deploy evidence待ち。R-05はlegacy presence inventory未実施でrollout/DROP HOLD。LINE claim refs は live では **0**（ledger 旧 prose は stale）。
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
- **状態**: **env-forward done / runtime credentials BLOCKED**（2026-08-01 reconfirm）。`run-e2e.sh` が host に設定時のみ `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`（+ optional `E2E_AUTH_STATE_PATH`）を name-only `-e` で Playwright docker へ転送。証拠: `reports/2026-07-31-task-020-env-forward.md`。prior runtime: 4p/3f/86 DNR（`reports/2026-07-31-task-020-runtime.md`）。**name-only preflight 2026-08-01: EMAIL=UNSET / PASSWORD=UNSET** → authenticated re-run 未実施。full green 未達（4 public passes を green と呼ばない）。claim/TASK-020 + W-020-ENV: **not live** on refs（stale live wording falsified）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: USER
- **Blockers (today)**: hostの認証env未注入（name-only UNSET）。前回は93 selected中4 passed / 3 failed / 86 did not runでfull green未達。
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
- **状態**: **Phase1 FE residual SAFE-CLEANUP done / Phase2 slice1+slice2 COMPLETE / in-repo FE ZERO_IN_REPO / external use UNREPORTED / CLEAN-GO·DROP·migrate HOLD**。Stage B: `e9dddd921`。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md` + LINE residual FINAL（021 Phase2）。inventory: `reports/2026-07-31-task-021-stage-a-inventory.md`。Phase1/Phase2 reports 済。**2026-08-01 inventory reconfirm**: exclusion routes/table/model/OpenAPI/response 生存；FE production consumer ZERO；`excluded_type_ids` hard-reject KEEP。次は USER external inventory または破壊変更承認後の CLEAN-GO のみ。claim/TASK-021 family: **not live** on refs（blocker の claim 半分は stale）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: BLOCKED(external use UNREPORTED・破壊承認未取得)
- **Owner lane**: AGENT→USER
- **Blockers (today)**: legacy endpoint外部利用ゼロが未証明。CLEAN-GO/DROP承認がない。route/OpenAPI/model/table/RLS/seed/export/test surfaceは現HEADで生存。
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
- **状態**: **agent source closeout done / human residual open**（2026-07-31）。`CreatePetGroup` fallback 除去 + regression。**scoped re-green 2026-08-01**: `docker compose exec -T backend go test -p 1 ./internal/identitylink ./internal/apicontract -count=1` → both packages **ok** exit 0。Phase 2 未着手。証拠: `reports/2026-07-31-task-022-identity-link-closeout.md`、S13: `docs/ops/testing/scenarios/S13-identity-links-manual-correction.md`。
- **残 human**: named operator の 2 医院 link→history→unlink→relink 実施と named signer 承認；RLS runtime を実 application role で証明（未なら UNREPORTED のまま Phase 2 禁止）。
- **claim**: `claim/TASK-022` — **not live**（ledger と一致）。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: USER
- **Blockers (today)**: named S13 operator/signer evidenceとreal application-role RLS runtime proofが未取得。Phase 2はBLOCKED。
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
- **状態**: **agent 証跡骨格 done / human residual open**（2026-07-31）。統合 report: `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md`。env health PASS；**2026-08-01 name-only: E2E_LOGIN_EMAIL/PASSWORD=UNSET** → authenticated E2E **BLOCKED**（launch しない）；5 フロー human 欄 PENDING；`confusion_count: 0` 単独では close 不可。#254 は human 揃うまで完了扱いしない。
- **残 human**: USER が secret channel で `E2E_LOGIN_*` 注入後 E2E 実行；QA が 5 フローブラウザ通し・DB/audit 目視・実 LINE/LIFF；PO/現場の使い勝手 sign-off と FAIL 処分承認。
- **claim**: `claim/TASK-023` — **not live**。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: USER
- **Blockers (today)**: authenticated E2E credentials未注入、QA/LINE owner/POのhuman evidence未取得。
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
- **状態**: **agent audit + FAQ disposition done / human visual sign-off open**（2026-07-31）。10/10 current/replace 判定。replace **4 採用**（`02`/`06`/`13`/`14`）；`05`/`07`/`10` clean-seed 再撮影待ち。FAQ **追記不要**。証拠: `reports/2026-07-31-task-024-manual-audit.md`。**manual vitest re-green 2026-08-01**: 5 files / **18 tests PASS**。manual-flow E2E は E2E_LOGIN_* UNSET で BLOCKED。
- **残 human**: named documentation owner の visual/content sign-off；任意で `19-aggregation` / `04-medical-records` 再撮影。
- **claim**: `claim/TASK-024` — **not live**。

#### 実装プラン（2026-08-01・codebase調査）
- **Ready**: READY_USER
- **Owner lane**: USER
- **Blockers (today)**: clean-seed版`05`/`07`/`10`再撮影とnamed documentation owner sign-offが未取得。manual-flow E2Eはcredential待ち。
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

### TASK-025: #201 dose parameter technical failure の silent fallback を止める（Critical / Clinical safety）

- **対応 Issue**: GitHub Issue #201。
- **問題**: FE は dose parameter 取得 error を manual default に変換するため、BE が repository/system error を保存中止へ伝播する契約を UI が silent bypass し得る。体重/species/parameter 欠落は別の cutover dependency を持つ。
- **状態**: **DONE**（実装 `eaa608b6a` + follow-up 是正。`main` に commit 済み・未 push）。欠落時 runtime 変更は TASK-033 まで HOLD のまま。
- **claim**: `claim/TASK-025`（取得済み。USER が統合後に解放）。

#### 実施結果（2026-08-02・TASK-025 unit）
- **Delivered**: `DoseParamsAuthority`（success / failed / pending / idle）と `DoseGateSource`（ready / missing / technical_failure）で技術障害と欠落を型で分離。`TreatmentsTab` の bare `catch {}` は固定文言 + 再試行 + `return`（create せず）に、`TreatmentRow` は `isError` を配線して `onUpdate` 前に停止。upstream body は画面に出さない。BE 無変更。
- **Verification** (scoped): `npx vitest run src/features/medical-records/components/TreatmentsTab` → **4 files / 35 tests PASS**（RED 時 5 failed）。`git diff --name-only HEAD -- backend/` 空。
- **Follow-up 是正（本セッション・reconciliation で検出）**:
  1. `toDoseParamsAuthority` が取得中を `success` + 空配列として符号化していた → `pending` を独立分岐にした。「取得成功したが param 無し」と区別できず、`status === "success"` を権威判定に使う実装が壊れるため。
  2. `computeDoseGate` が `DoseCalcInput | null` を受け付けたまま残っていた（production call site は 3 つとも移行済みで、使用者はテストのみ）→ `DoseGateSource` のみへ narrowing。null 許容は「技術障害を欠落と同一視して保存を通す」経路の型上の復活だった。
  3. `resetQueries` が通常の薬剤選択経路で毎回実行されていた → 再試行ハンドラ内へ移動。共有 queryKey の無条件リセットは STATIC staleTime を無効化し、同一薬剤の `TreatmentRow` が一往復のあいだ data 無しに落ちて行の投与量ゲートが一時的に開いていた（本 unit が持ち込んだ回帰）。
- **Non-actions / HOLD**: missing-data runtime behavior、構造化救急投薬記録、上限値・warning 数値、DB/migration、Issue close、claim 削除、push は未実施。

#### 実装プラン（2026-08-01・readiness dossier）
- **Ready**: READY_AGENT
- **Owner lane**: frontend clinical-safety contract（current BE technical error propagation を維持）
- **Blockers (today)**: technical failure slice はなし。体重/species/parameter 欠落時の runtime cutover は TASK-033 まで HOLD。
- **Preconditions**: DEC-48 を読み、lookup technical failure を欠落と区別した typed state にし、通常保存を停止する。current missing-data behavior はこの unit で変更しない。
- **Code anchors**: `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx:239-277`, `TreatmentRow.tsx:80-105,172-187`, `backend/internal/medicalrecord/treatment_dose_save.go:14-18,29-73`。
- **Steps**:
  1. RED: parameter fetch error 時に visible error、通常保存不能、retry が現れる component test を追加する。
  2. RED: technical failure 中は onUpdate と通常 treatment write が zero、error message は upstream body を転記せず、retry が表示されることを固定する。
  3. GREEN: query error を missing-data state と区別して row/save gate へ渡し、error の manual default 変換を残さない。
  4. retry で authoritative parameter 取得が成功した後だけ通常保存可能に戻す。
- **Verification** (scoped only):
  - `docker compose exec -T frontend npx vitest run src/features/medical-records/components/TreatmentsTab/TreatmentsTab.test.tsx src/features/medical-records/components/TreatmentsTab/TreatmentRow.test.tsx`
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Dose.*Lookup' -count=1`
- **Non-actions / HOLD**: missing-data runtime behavior、構造化救急投薬記録、dedicated override、上限値・warning 数値、DB/migration、Issue close、claim 削除を行わない。
- **Exit criteria for close**: technical failure が UI/BE で missing と区別され、通常保存/write が zero、visible retry と parameter 解決後の復帰が scoped tests で green。欠落時 contract の close は TASK-033 に残る。
- **Evidence sources read**: `reports/2026-08-01-issue-readiness-dossier.md` Issue #201、DEC-48、live FE/BE source と tests。

### TASK-026: #249 confirmed 検査の transaction 順序・409 lock・parent mutation audit（Critical / Clinical record integrity）

- **対応 Issue**: GitHub Issue #249。
- **問題**: confirm が親 status を先に更新して item replace を自己拒否し得る。confirmed delete に status guard がなく、既存 status conflict は 400 相当。parent examination の create/update/confirm/delete は authenticated actor と application audit を持たない。
- **状態**: **DONE**（実装 `2a8aca33c`。2026-08-01 に `--no-ff` merge で `main` へ統合。todo.md の 5 箇所を手動解決）。
- **claim**: `claim/TASK-026` — **live**。統合済みのため解放可能（削除は USER 専権）。

#### 実施結果（2026-08-01・TASK-026 unit）
- **Ready**: READY_USER_INTEGRATE
- **Delivered**: confirm は items/range の検証・保存後に最後に `confirmed` へ遷移。confirmed update/delete/items mutation は 409。create/update/confirm/delete は authenticated actor と before/after snapshot を持つ transaction-bound audit を記録し、audit failure は全 rollback。clinic/pet/record correlation は fail-closed。
- **Verification**: backend medicalrecord primary + lab regression + full package PASS（coverage **91.8%**）、frontend examinations **3 files / 61 tests PASS**、`internal/apicontract` PASS、`go vet` PASS、`cmd/api` compile PASS、changed-diff golangci-lint **0 issues**、`git diff --check` PASS。
- **Verification boundary**: `internal/model` 全体は既存の `CashRegisterClose.deleted_at` test-schema drift により full green を主張しない。TASK-026 で変更した audit model regression は PASS。
- **Independent review**: Go / clinic isolation / healthcare / database は PASS。security / acceptance は APPROVE-WITH-NOTE（repository 直呼びに defensive confirmed guard はないが、production call graph は service 経由で blocking finding なし）。
- **Integration gate**: 2026-08-01 に `main` へ統合済み（到達性は `git merge-base --is-ancestor 2a8aca33c HEAD` で確認）。TASK-027/031/032 の dependency wait は解除された。
- **Non-actions / HOLD**: migration、clinical range 値、external import、auto-commit enable、Issue close、claim 削除、push/merge は未実施。

### TASK-027: #249 手動検査の結果行操作・患者変更・confirmed→completed 確定解除（High）

- **対応 Issue**: GitHub Issue #249。
- **問題**: manual workflow の row add/delete、confirm 前 patient change、権限付き確定解除が未完。現行 examination status に <code>unconfirmed</code>/<code>cancelled</code> はなく、lab import job の取消と混ぜてはならない。
- **状態**: **READY_AGENT**。TASK-026 の immutable confirmed contract は実装済み。
- **claim**: `claim/TASK-027` — **not live**（2026-08-01 USER 解放済み。起票時の過剰取得を是正したもので、本タスクは未着手）。

#### 実装プラン（2026-08-01・readiness dossier）
- **Ready**: READY_AGENT
- **Owner lane**: backend medicalrecord + frontend examinations
- **Blockers (today)**: なし。print は TASK-031、lab import revert は TASK-032。external file/crosswalk、clinical range、auto-commit は対象外。
- **Preconditions**: DEC-53/TASK-026 regression、DEC-57、Issue #249 の current AC、既存 examination RBAC/audit contract を再読する。
- **Code anchors**: `frontend/src/features/examinations/components/ExamItemsTable.tsx:91-98`, `backend/internal/medicalrecord/routes.go:360-371`, `backend/internal/medicalrecord/examination_service.go`, `backend/internal/model/examination_record.go:10-17`, examination feature tests。
- **Steps**:
  1. 現行 status（pending/in_progress/result_entered/completed/confirmed）× row/pet/unconfirm operation × permission × audit の matrix を test fixture に固定し、存在しない状態を追加しない。
  2. non-confirmed の result row add/delete と confirm 前 patient change を clinic/pet/medical-record correlation fail-closed で実装する。
  3. 専用 permission + 理由必須の unconfirm endpoint を追加し、clinic-scoped lock 下で <code>confirmed -&gt; completed</code>、authenticated actor、before/after audit を同一 transaction に置く。
  4. confirmed direct mutation、別 clinic/pet/record、理由/actor/audit dependency 欠落を拒否し、write/audit zero または全 rollback を確認する。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*Examination' -count=1`
  - `docker compose exec -T frontend npx vitest run src/features/examinations`
- **Non-actions / HOLD**: print、lab import job 取消、clinical range 推測、external file/crosswalk、auto-commit、migration/seed、Issue close、claim 削除を行わない。
- **Exit criteria for close**: 現行状態だけの matrix で edit/unconfirm が permission/status/actor/audit contract を満たし、<code>confirmed -&gt; completed</code> 以外を作らず、TASK-026 regression と clinic isolation が green。
- **Evidence sources read**: dossier Issue #249、DEC-53/57、Issue body/current routes/FE source。

### TASK-028: #252 standard closing settings PATCH の validation・lost-update 防止・transaction-bound audit（High）

- **対応 Issue**: GitHub Issue #252 の OPS apply から分離した technical gap。
- **問題**: standard update は read-modify-save で全設定列を upsert する。special period 相当の boundary validation、actor/audit/transactor、row lock/CAS がなく、並行 partial PATCH が相互に上書きされ得る。
- **状態**: **DONE**（`bbf82e2b8`、2026-08-01 land・push 済）。投入値は変更せず、production apply は USER。runtime green は未主張。
- **claim**: `claim/TASK-028` — **live**（実装 unit が取得。`main` 統合後に USER が解放する）。

#### 実装プラン（2026-08-01・readiness dossier）
- **Ready**: READY_AGENT
- **Owner lane**: backend clinic settings / audit transaction
- **Blockers (today)**: なし。実値・対象 clinic・apply window は USER gate。
- **Preconditions**: DEC-54、closing settings request/service/repository/composition、audit/DBOrTx conventions を読む。
- **Code anchors**: `backend/internal/clinic/closing_settings_service.go:88,141-165,350-364`, `closing_settings_handler.go:29`, `clinic_settings_repository.go:50`, `closing_settings_request.go:3-7`, `closing_settings_service_test.go:206-255`。
- **Steps**:
  1. RED: invalid time ordering/range/partial combination reject を table-driven test にする。
  2. RED: 同一 clinic への並行 partial PATCH が lost update せず、別 clinic は競合しない concurrency test を追加する。
  3. RED: authenticated actor 付き valid update と before/after audit が同一 transaction、audit dependency 不在/failure で update rollback する test を追加する。
  4. special-period validation pattern を再利用し、clinic-scoped row/advisory lock または CAS の一方式で read-modify-save を直列化する。
  5. handler→service へ actor を明示伝播し、clinic_id、actor、before/after の非機密 metadata を audit して cross-clinic master を参照しない。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/clinic -run 'TestClosingSettingsService_UpdateStandard.*(Concurrent|Audit|Rollback|Validation)|TestUpdateClosingSettings' -count=1`
- **Non-actions / HOLD**: production value apply、過去履歴再計算、DB/migration、Issue close、claim 削除を行わない。
- **Exit criteria for close**: invalid input が fail-fast、並行 partial PATCH が lost update せず、actor 付き update/audit が atomic、audit dependency/failure rollback と clinic scope regression が green。
- **Evidence sources read**: dossier Issue #252、DEC-54、live closing settings source/tests。

#### 実施結果（2026-08-01・TASK-028 unit）
- **Outcome**: `UpdateStandard` の unlocked read-modify-save を是正。境界値 validation（時刻書式・順序・`closed_weekdays` 範囲）、`Transactor.WithTx` 内で親 `clinics` 行を `FOR UPDATE`（read の**前**）、handler からの actor 伝播（`httpapi.ExtractStaffID`）、同一 transaction の fail-closed audit を追加。
- **直列化方式**: 親 clinic 行の row lock。`clinic_settings` 行は初回 upsert 時に存在しないため、その行を掴んでも直列化されない。CAS は version 列の migration を要するため不採用。
- **Changed files**: `closing_settings_service.go`, `closing_settings_handler.go`, `closing_settings_service_test.go`, `closing_settings_update_standard_integrity_test.go`（新規 398 行）, `closing_settings_handler_test.go`, `composition_clinic.go`, `composition_runtime.go`, `composition_clinic_test.go`, `reports/2026-08-01-task-028-closing-settings-integrity.md`
- **Gates**: 4 系統 RED→GREEN（Validation / Concurrent / Audit / Rollback）。回帰 `./internal/clinic ./cmd/api` は baseline の既存 2 FAIL（holiday 系）に対し新規失敗 0。
- **Evidence 品質の但し書き**: concurrency の RED は単発では確実に FAIL せず（flaky PASS）、15 ラウンドの実 DB テストで固定した。他 3 系統より証拠が弱い。
- **Audit 設計**: 締め時間の実値は記録せず、変更フィールドの presence metadata のみ。
- **Non-actions**: 実値投入、apply window、OpenAPI 更新（`api.yaml` が並行セッションで dirty）、Issue #252 close、claim 削除、push は未実施。
- **Report**: `reports/2026-08-01-task-028-closing-settings-integrity.md`

### TASK-029: #259 Lステップ deploy/clinic gate の異なる disabled contract を文書同期する（Medium / docs-only）

- **対応 Issue**: GitHub Issue #259 の source/docs drift。
- **問題**: deploy gate OFF は disabled error + HTTP zero、clinic の <code>is_sync_enabled=false</code> は intentional skip/noop だが、一部 spec が二 gate を混同する。
- **状態**: **DONE**（`b659ac952`+`9fc5b9ffb`、2026-08-01 USER push 済）。write/cron code の再実装はしていない。runtime green は未主張（STG/production の cron 自然発火・実送信は未実測）。
- **claim**: `claim/TASK-029` — **not live**（2026-08-01 統合後に USER が解放済み）。

#### 実装プラン（2026-08-01・readiness dossier）
- **Ready**: READY_AGENT
- **Owner lane**: `docs/spec/screens/31-lstep-integration.md`, `docs/spec/screens/34-lstep-delivery-monitor.md`, `docs/spec/line/cost-analysis.md`
- **Blockers (today)**: なし。external enablement と live send は USER/先方 gate。
- **Preconditions**: DEC-55、`docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`、current write client/scheduler tests を読む。
- **Code anchors**: 上記三 docs、`backend/internal/infra/lstep/client.go:22-25,72-85`, `backend/internal/lstep/lstep_delivery_trigger_service_test.go:838-841,895-897`, `backend/wrangler.jsonc:97-102`, `backend/worker/scheduled-jobs.ts:30-34`。
- **Steps**:
  1. deploy gate OFF を <code>ErrWriteDisabled</code> + HTTP zero、clinic gate OFF を intentional skip/noop と別記し、片方の contract を他方へ一般化しない。
  2. scheduler/cron 配線済みと、STG/production の自然発火・実送信が未実測である境界を分離する。
  3. pause runbook を唯一の enable/stop/rollback 正本として link し、契約値や環境実値を記載しない。
- **Verification** (scoped only):
  - `rg -n 'noop|no-op|ErrWriteDisabled|LSTEP_WRITE_API_ENABLED|cron' docs/spec/screens/31-lstep-integration.md docs/spec/screens/34-lstep-delivery-monitor.md docs/spec/line/cost-analysis.md docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`
  - `bash scripts/check-docs-symbol-drift.sh`
- **Non-actions / HOLD**: external enable、実送信、cron fire、環境実値、write/scheduler code、Issue close、claim 削除を行わない。
- **Exit criteria for close**: 三 docs が deploy error/HTTP-zero と clinic skip/noop を分離して current scheduler contract と一致し、外部/runtime 未実測を green と書かず、docs drift check が green。
- **Evidence sources read**: dossier Issue #259、DEC-55、live source/runbook/tests。

#### 実施結果（2026-08-01・TASK-029 unit）
- **Outcome**: 3 spec doc の deploy gate / clinic gate 混同（noop 一語）を是正。deploy OFF = `ErrWriteDisabled` + HTTP 未送信、clinic OFF = `nil, nil` intentional skip を別項で記述。cron 配線済みと STG/production 未実測を分離。
- **Changed files**: `docs/spec/screens/31-lstep-integration.md`, `docs/spec/screens/34-lstep-delivery-monitor.md`, `docs/spec/line/cost-analysis.md`, `reports/2026-08-01-task-029-lstep-gate-contract.md`, `todo.md`（本追記）
- **Gates**:
  - `rg -n 'noop|no-op'` on 3 doc → deploy-gate-as-noop 行 0
  - `rg -n 'ErrWriteDisabled'` / `is_sync_enabled` / `LSTEP_WRITE_API_PAUSE` → 3 doc すべてヒット
  - `bash scripts/check-docs-symbol-drift.sh` → exit 0
- **Non-actions**: backend/worker 未変更、Issue #259 未操作、claim 未削除、env 実値未記載、runtime/cron 未実行
- **Audit report**: `reports/2026-08-01-task-029-lstep-gate-contract.md`
- **claim**: `claim/TASK-029` は 2026-08-01 に USER が解放済み



### TASK-030: #261 trimming 死亡ペット拒否の経路別 regression と stale phase2 同期（High / Clinical safety）

- **対応 Issue**: GitHub Issue #261。
- **問題**: trimming detail create/update は request に <code>pet_id</code> がある場合だけ死亡確認し、予約から算出した <code>finalPetID</code> を常時検証しない。pet_id 省略の通常経路で死亡済み予約ペットが通り得て、経路別 test もない。`phase2.html` の guard 欠落記述も current source とずれる。
- **状態**: **DONE**（`6e5a945ef`、2026-08-01 land・push 済）。Issue 全体の runtime/OPS completion は USER gate のまま。runtime green は未主張。
- **claim**: `claim/TASK-030` — **live**（実装 unit が取得。`main` 統合後に USER が解放する）。

#### 実装プラン（2026-08-01・readiness dossier）
- **Ready**: READY_AGENT
- **Owner lane**: backend trimming tests + `phase2.html` current-source sync
- **Blockers (today)**: なし。対象環境 runtime、DB、real LINE/LIFF は含めない。
- **Preconditions**: DEC-41/47、deceased pet shared helper、trimming service paths、phase2 truth boundary を読む。
- **Code anchors**: `backend/internal/trimming/trimming_service.go:275-280,490-500,646-656`, `trimming_service_test.go:29,153-157`, `backend/internal/sharedkernel/pet_not_deceased.go:10-31`, `phase2.html:206`。
- **Steps**:
  1. RED: detail create/update で request の pet_id が nil かつ予約由来 finalPetID が死亡、明示 pet replacement が死亡、通常 create が死亡の各経路を固定する。
  2. GREEN: request の有無に関係なく finalPetID を business write 前に検証し、拒否時 repository write/audit が zero であることを確認する。
  3. living pet regression と clinic mismatch を維持する。
  4. `phase2.html` の旧「guard 欠落」を current source/test と runtime 未実測の表現へ同期する。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/trimming -run 'TestTrimmingService_.*Deceased' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/sharedkernel ./internal/reservation ./internal/trimming -count=1`
- **Non-actions / HOLD**: DB/migration/seed、対象環境 runtime、実機 LINE/LIFF、臨床値、Issue close、claim 削除を行わない。
- **Exit criteria for close**: pet_id 省略を含む各経路の deceased rejection、zero write/audit、living/clinic regression が green で、phase2 が source proof と runtime proof を混同しない。
- **Evidence sources read**: dossier Issue #261、DEC-41/47、live trimming/sharedkernel source/tests。

#### 実施結果（2026-08-01・TASK-030 unit）
- **Outcome**: `createDetailForExistingAppointment` と `Update` で死亡ペット検証が `if input.PetID != nil` に囲まれ `input.PetID` を検証していた欠陥を是正。両箇所とも条件分岐を外し `finalPetID` を無条件検証へ変更（`trimming_service.go:498` / `:653`）。`Create` は元から無条件のため未変更。
- **nil 安全性**: `ValidateReservationPetNotDeceased`（`reservation_service.go:187-189`）は `petID == nil` で early return するため、ペット未紐付け予約では no-op。無条件化による回帰は無い。
- **Changed files**: `trimming_service.go`, `trimming_deceased_pet_test.go`（新規 247 行）, `phase2.html`, `reports/2026-08-01-task-030-deceased-pet-guard.md`
- **Gates**: RED = nil pet_id の detail create / Update 2 経路が FAIL。GREEN = `TestTrimmingService_.*Deceased` 4/4 PASS。回帰 `./internal/trimming ./internal/sharedkernel ./internal/reservation` は baseline / after とも FAIL 空。
- **phase2.html 同期**: 「trimming 3 関数が `ValidateReservationOwnerPetLinksWithRepo` のみ」という 2026-07-29 時点の記述は current source と不一致だった（実際は呼んでいるが条件分岐内）。実測に合わせて是正。
- **Non-actions**: `appointment_admin_service.go` / `liff_service_reservations.go` は実測のみでコード未変更。runtime / 実機 LINE・LIFF / DB・migration、Issue #261 close、claim 削除、push は未実施。
- **Remaining**: LIFF の `resolveReservationPetID` は in-memory livingPets のみで DB 行固定の検証ではない（別決裁）。他 domain の「input pet があるときだけ死亡検証」ギャップの棚卸しも未実施。
- **Report**: `reports/2026-08-01-task-030-deceased-pet-guard.md`

### TASK-031: #249 検査結果を保存済み snapshot から印刷する（Medium）

- **対応 Issue**: GitHub Issue #249 F-5a。
- **問題**: 飼主説明・他院添付・院内保管向け print surface が未完。画面 state や FE 再計算を印刷正本にすると保存済み臨床記録と不一致になり得る。
- **状態**: **READY_AGENT**。TASK-026 の immutable/audit contract は実装済み。
- **claim**: `claim/TASK-031` — **not live**（2026-08-01 USER 解放済み。起票時の過剰取得を是正したもので、本タスクは未着手）。

#### 実装プラン（2026-08-01・readiness dossier）
- **Ready**: READY_AGENT
- **Owner lane**: frontend examinations / print presentation
- **Blockers (today)**: なし。臨床 range の新規推測、manual unconfirm、lab import revert は対象外。
- **Preconditions**: DEC-53/TASK-026 regression、DEC-57、Issue #249 F-5a、`PrintPortal` の既存利用例、#229 の飼主向け表現境界を読む。
- **Code anchors**: `frontend/src/components/shared/PrintPortal.tsx`, `frontend/src/features/examinations/components/ExamPivotTable.tsx`, `frontend/src/features/examinations/api/get-examination-items.ts`, examinations feature tests。
- **Steps**:
  1. RED: 保存済み examination/items snapshot だけを入力にし、実施項目のみ、欠測、定性値、日付/単位を表示する print component test を追加する。
  2. RED: FE が status/range を再計算しないこと、画面上の未保存 edit を印刷しないこと、test ID が一意であることを固定する。
  3. GREEN: `PrintPortal` を再利用し、三用途で共通の保存 snapshot view model を生成する。横長 matrix のみ landscape とする。
  4. print preview の browser/UAT は個人情報を含まない clean-demo で USER/QA が行い、本 task の source green と区別する。
- **Verification** (scoped only):
  - `docker compose exec -T frontend npx vitest run src/features/examinations`
- **Non-actions / HOLD**: 臨床判定再計算、新 range 値、manual unconfirm、lab import、実データ screenshot、Issue close、claim 削除を行わない。
- **Exit criteria for close**: print が保存 snapshot のみを表示し、FE 再計算・未保存値混入がなく、scoped component tests が green。human print sign-off は別 evidence。
- **Evidence sources read**: dossier Issue #249、DEC-53/57、Issue #249 F-5a、current print/examination source。

### TASK-032: #249 lab import job の compensating revert を examination unconfirm と分離する（Critical / Clinical record integrity）

- **対応 Issue**: GitHub Issue #249 F-3c(a)。
- **問題**: persisted import job を取消す endpoint/状態がなく、手動 examination の確定解除と混ぜると status・permission・audit・rollback の意味が不定になる。
- **状態**: **READY_AGENT / migration review required**。設計は DEC-57。
- **claim**: `claim/TASK-032` — **not live**（2026-08-01 USER 解放済み。起票時の過剰取得を是正したもので、本タスクは未着手）。

#### 実装プラン（2026-08-01・readiness dossier）
- **Ready**: READY_AGENT
- **Owner lane**: backend medicalrecord / lab import compensation
- **Blockers (today)**: なし。migration-seed-safety と database review を開始時に通す。external format/auto-commit enable は対象外。
- **Preconditions**: DEC-53/TASK-026 regression、DEC-57、`backend/migrations/CLAUDE.md`、clinic isolation、DBOrTx/audit conventions、lab import transition table を読む。
- **Code anchors**: `backend/internal/model/lab_import.go:10-21`, `backend/internal/medicalrecord/lab_import_service.go:15-29`, `backend/internal/model/examination_record.go:32-43`, `backend/internal/medicalrecord/routes.go:373-388`, lab import service/repository tests。
- **Steps**:
  1. RED: clinic-scoped persisted job の revert success、wrong-clinic/invalid-state/second-revert 409、reason/actor/audit dependency 欠落、linked confirmed exam conflict、audit failure rollback を追加する。
  2. 新規 migration で terminal <code>reverted</code> status を追加し、既適用 migration/seed bundle を編集しない。transition table と API contract を同期する。
  3. 専用 permission + 理由必須 endpoint で job と linked exams を clinic-scoped lock し、confirmed exam があれば全体を 409 で拒否する。
  4. 未確定の job 由来 parent exams の soft delete、job <code>persisted -&gt; reverted</code>、authenticated actor + before/after audit を同一 transaction に置き、child result を hard deleteしない。
  5. API spec/codegen、migration static checks、clinic isolation、rollback regressions を同一 unit で検証する。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*LabImport.*Revert|Test.*Examination.*Audit' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/apicontract ./internal/lintscan -count=1`
- **Non-actions / HOLD**: confirmed examination の自動解除、child result hard delete、external format/crosswalk、auto-commit enable、seed edit/apply、Issue close、claim 削除を行わない。
- **Exit criteria for close**: manual unconfirm と別 endpoint/status/permission で、confirmed linked record を拒否し、revert/audit/soft delete が atomic、clinic isolation と migration/API regression が green。
- **Evidence sources read**: dossier Issue #249、DEC-53/57、Issue #249 F-3c、current lab import/examination source/tests。

### TASK-033: #201 active/draft 構造化救急投薬記録 + 欠落時 fail-closed cutover（Critical / Clinical safety）

- **対応 Issue**: GitHub Issue #201。
- **問題**: current addendum は finalized medical record 専用の自由記述で、薬剤、実投与量・単位、投与時刻を構造化せず、active/draft の救急・既実施投薬を通常治療履歴と handoff に残す代替経路ではない。代替経路なしに体重/species/parameter 欠落時の通常保存だけを止めると、救急記録を失う。
- **状態**: **READY_AFTER_CLINICAL_APPROVAL / migration review required**。最終 fail-closed 契約と cutover 順序は DEC-48。臨床値は未決定。
- **claim**: `claim/TASK-033` — **not live**（2026-08-01 USER 解放済み。起票時の過剰取得を是正したもので、本タスクは未着手）。

#### 実装プラン（2026-08-01・independent clinical review reconciliation）
- **Ready**: READY_AGENT_AFTER_CLINICAL_APPROVAL
- **Owner lane**: backend medicalrecord + frontend TreatmentsTab/history + migration/API contract
- **Blockers (today)**: 臨床責任者による記録対象ケース、専用権限、理由、訂正条件の一行承認。開始時に migration-seed-safety、database、clinic-isolation、healthcare review を通す。
- **Preconditions**: DEC-48 と clinical pack、`backend/migrations/CLAUDE.md`、clinic isolation、DBOrTx/audit conventions を読み、TASK-025 technical failure slice を green にする。欠落時 fail-closed を単独で先行有効化しない。
- **Code anchors**: `backend/internal/model/treatment.go:29-60`, `backend/internal/medicalrecord/treatment_request.go:7-24,47-64`, `backend/internal/medicalrecord/treatment_dose_save.go:14-18,29-73`, `backend/internal/medicalrecord/medical_record_addendum_service.go:75-104`, `backend/internal/model/medical_record_addendum.go`, `frontend/src/features/medical-records/components/TreatmentsTab/{TreatmentsTab,TreatmentRow}.tsx`。
- **Steps**:
  1. RED: active/draft medical record に clinic/pet/medical-record 相関、medicine ID、実投与量・単位、投与時刻、理由、authenticated actor を必須とする immutable emergency administration event の create/read と、通常治療履歴・handoff 表示を追加する。
  2. RED: wrong clinic/pet/medical-record/medicine、欠けた dose/unit/time/reason/actor、権限なし、audit dependency/write failure、重複・競合を拒否し、event と audit が同一 transaction で rollback することを固定する。
  3. GREEN: 新規 append-only migration で clinic-scoped dedicated event と必要な相関制約・index・訂正リンクを追加し、hard delete/上書き更新を許さない repository/service/API を実装する。既適用 migration/seed は編集しない。
  4. 専用 permission と actor を handler→service へ伝播し、理由と before/after を application audit に同一 transaction で記録する。通常 treatment history/handoff は event を構造化表示し、free-text addendum を代替扱いしない。
  5. 記録経路、clinic isolation、audit、API/UI が green かつ臨床承認済みになった同一 unit で、体重なし、species ID/名称欠落・不正、parameter なしを理由別 typed state にし、通常 dose 保存/write を fail-closed に切り替える。部分 cutover を禁止する。
  6. API spec/codegen、migration static checks、FE/BE regressions を実行し、current master の臨床値を変更しない。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord/... -run 'Test.*(EmergencyAdministration|DoseMissing|DoseSpecies|DoseParameter)' -count=1`
  - `docker compose exec -T frontend npx vitest run src/features/medical-records/components/TreatmentsTab/TreatmentsTab.test.tsx src/features/medical-records/components/TreatmentsTab/TreatmentRow.test.tsx`
  - `docker compose exec -T backend go test -p 1 ./internal/apicontract ./internal/lintscan -count=1`
- **Non-actions / HOLD**: 臨床上限・warning 値の発明、臨床承認前または構造化経路 green 前の missing-data cutover、既存 addendum の投薬記録代用、既適用 migration/seed edit/apply、DB 操作、Issue close、claim 削除を行わない。
- **Exit criteria for close**: active/draft の構造化事実記録が必須 field、clinic/pet/record 相関、permission、actor/audit atomicity、immutable correction、通常履歴/handoff を満たし、その安全経路と臨床承認を前提に missing-data 通常保存/write が理由別に zero となる。current runtime を部分的に切り替えていない。
- **Evidence sources read**: dossier Issue #201、DEC-48、clinical pack、current treatment/addendum source、independent healthcare review。

---
