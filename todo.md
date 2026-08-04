# Remaining work ledger (open only)

オープン residual のみを列挙する。対応済み TASK / closed 索引行は **削除済み**（2026-07-31 更新）。  
根拠・完了証拠は git 履歴と `reports/2026-07-31-*.md` を参照。

> **ID namespace**: 本ファイルの `TASK-*` はローカル連番で、現行の実装タスク番号系そのもの（claim ブランチ `claim/TASK-XXX` も同一番号）。受入テストバグは `bug.md` の `BUG-*`。`/implement` は本ファイルと `bug.md` から解決する。旧 `3-session-agent.html#ledger` 体系は 2026-07-31 廃止（同ファイルは GitHub Issue 分類ビュー）。

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
| ISSUE-249-MANUAL-LIFECYCLE | 手動検査 edit / confirmed→completed 確定解除 | **TASK-027**（DONE `046615f4b`〜`dfd653eaa`・migration 004 適用済み） |
| ISSUE-252-STANDARD-PATCH | 締め設定 standard PATCH の validation/audit | **TASK-028**（DONE `bbf82e2b8`・値投入は USER） |
| ISSUE-259-DOC-CONTRACT | Lステップ disabled 時の旧 noop 文書 | **TASK-029**（DONE・`9fc5b9ffb` push 済。残は先方 enable + USER runtime 実測） |
| ISSUE-261-TRIMMING-DECEASED | trimming 死亡ペット拒否の経路別回帰 | **TASK-030**（DONE `6e5a945ef`・runtime は USER） |
| ISSUE-249-PRINT-SNAPSHOT | 検査結果の保存 snapshot 印刷 | **TASK-031**（READY_AGENT・TASK-027 の interface freeze 完了により着手可能） |
| ISSUE-249-IMPORT-REVERT | lab import job の compensating revert | **TASK-032**（DONE code・migration 未適用・claim 解放 USER） |
| ISSUE-201-EMERGENCY-ADMIN | 構造化救急投薬記録と欠落時 fail-closed cutover | **TASK-033**（臨床承認・migration review 後） |
| ISSUE-211-CLINIC-PACKAGE-IMPORT | 健診 package の clinic-scoped import/preflight | **TASK-374**（READY_AGENT・additive migration review 必須・実値/apply は USER） |
| ISSUE-257-GOLIVE-REPLAN | 期限切れ go-live runbook の gate-driven 再計画 | **TASK-375**（READY_AGENT・docs-only） |
| ISSUE-258-DELIVERY-BOUNDARY | #258 U1〜U12 と #256 U13 の文書境界同期 | **TASK-376**（READY_AGENT・docs-only） |
| ISSUE-201-DOSE-DEVIATION-REASON | warning 逸脱理由を FE/BE・snapshot・同一 transaction audit へ接続 | **TASK-377**（READY_AGENT・臨床値は別 gate） |

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
- **状態**: **DONE**（実装 `eaa608b6a` + follow-up 是正 `db8387035`。push 済）。欠落時 runtime 変更は TASK-033 まで HOLD のまま。
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
- **状態**: **DONE**（2026-08-03 実装完了。詳細は下記「実施結果」。DB review が要求した revision→examination parent FK は tenant-safe composite FK として解決済み、migration `004_examination_revisions.sql` は適用済み）。
- **claim**: `claim/TASK-027` — **live**（実装 unit が取得。解放は USER 専権）。

#### 実装プラン（2026-08-02・adversarial revision 2 / 実行済み）
- **Ready**: 実行済み。exams の clinic-first candidate key と revision→examination parent FK を明示した上で foundation を land した。interface freeze 完了により TASK-031 は着手・完了済み。共有 route/OpenAPI を持つ TASK-032 は 2026-08-04 の plan revision で fresh DB review が RESOLVED を返し READY_AGENT へ移行済み（claim 解放待ち）。external format、clinical range、auto-commit は対象外。
- **Owner lane**: backend medicalrecord + frontend examinations + additive revision migration + dedicated permission/API contract。
- **Fixed clinical-record contract**: status は `pending / in_progress / result_entered / completed / confirmed` のみ。理由必須の `POST /v1/examinations/:id/unconfirm` は `confirmed -> completed` だけを許すが、遷移前に parent+items の official revision を append-only に保存し、新しい working revision を append する。解除後の edit も既存 revision/item を更新せず、新しい working revision+items を append して exam pointer を version CAS する。再 confirm は current working revision から新しい official version を append する。official print/history は revision store だけを読み、mutable legacy parent/current items と混ぜない。過去に一度でも confirmed version がある examination は pet を変更不可。pre-first-confirm の pet change だけを record/pet/owner/species/doctor/master 相関再検証と items assessment 再計算の同一 transaction で許可する。
- **Migration impact**: **YES**。開始時に root 番号を再測定し、append-only `examination_revisions` + `examination_revision_items`（clinic/examination/version/kind/status/pet/record/snapshot/reason/actor/timestamp/official flag）を additive migration で追加する。revision は `UNIQUE(clinic_id,examination_id,version)`、items は同じ clinic/exam/version composite FK。`exams.current_revision_version` は `(clinic_id,id,current_revision_version)` から revision `(clinic_id,examination_id,version)` への tenant-safe FK とし、`WHERE clinic_id=? AND id=? AND current_revision_version=?` の pointer CAS だけを更新可能にする。revision parent/items は official/working を問わず UPDATE/DELETE rejection triggerを持つ。両表へ project helper の `ENABLE ROW LEVEL SECURITY` + clinic `USING`/`WITH CHECK` を明示適用し、`FORCE` は role compatibility review なしに追加しない。indexes は revisions `(clinic_id,examination_id,version)` と official lookup、items `(clinic_id,examination_id,version,sort_order,id)`。既存 confirmed record は fail-closed backfill inventory を通し、official revision がない confirmed record は unconfirm/official print を拒否する。既適用 migration/seed は編集・apply しない。
- **Permission rollout**: 新 resource `examination-unconfirm` は全 default group で default-deny。fresh/demo fixture は disposable DB + sanctioned `seed-export` のみ。既存 clinic への grant は seed replay ではなく、named approver 後の permission-group API/app operation として別 evidence にする。`AllResources`、backend default table、FE label、parity test を同期する。
- **対象ファイル（path:line）**:
  - backend: `backend/internal/medicalrecord/examination_service.go:273`, `backend/internal/medicalrecord/examination_audit.go:11`, `backend/internal/medicalrecord/examination_handler.go:86`, `backend/internal/medicalrecord/routes.go:360`, `backend/internal/model/examination_record.go:10`, `backend/internal/model/audit_log.go:98`, `backend/internal/model/permission.go:6`, `backend/internal/clinic/clinic_service.go:192`, `backend/docs/api.yaml:3462`、new revision model/repository/request/response と migration。
  - frontend: `frontend/src/hooks/use-update-examination.ts:12`, `frontend/src/features/examinations/api/types.ts:27`, `frontend/src/features/examinations/hooks/use-examination-form.ts:175`, `frontend/src/features/examinations/components/ExamItemsTable.tsx:91`, `frontend/src/features/examinations/routes/ExaminationForm.tsx:38`, `frontend/src/features/master/components/permission-rule-table-model.ts:15`、new `unconfirm-examination.ts`。
- **影響 caller 全数**: confirm/update/items service+handler+routes+composition、OpenAPI/codegen、permission defaults、shared update hook、feature form/table。baseline command: `rg -l 'ConfirmExamination|UpdateExamination|ReplaceExaminationItems|ExaminationService|ExaminationHandler|ResourceExaminations|examinations/:id' backend frontend --glob '*.{go,ts,tsx,yaml}' --glob '!backend/migrations/seeds/**'`。
- **既存テストを RED から拡張**: `examination_parent_audit_test.go`, `examination_parent_audit_tx_test.go`, `examination_handler_test.go`, `routes_snapshot_test.go`, `checkup_examination_relation_write_db_test.go`, `examination_cross_tenant_master_fk_write_test.go`, `openapi_examination_mutation_contract_test.go`, `ExamItemsTable.test.tsx`, `use-examination-form.test.ts`, `ExaminationFormFields.test.tsx`, `ExaminationForm.permissions.test.tsx`, permission default/parity tests。
- **Steps**:
  1. RED: first official revision、unconfirm→working copy、新旧version isolation、old official immutability、official-readのworking混入禁止、post-confirm pet change reject、pre-confirm pet change assessment recalculation、wrong clinic/record/pet/owner/doctor/master、actor/audit rollback、permission default-deny を固定する。
  2. GREEN: ambient transaction 内で clinic-scoped exam lock→全 clinical relation 再検証→official/working revision+items append→audit→status/current-version pointer CAS の順を固定し、revision/audit/CAS failure は全 rollback。
  3. GREEN: master-selected row add/delete は working revision、pre-first-confirm pet change は pre-official current rowsだけに実装し、confirmed/ever-confirmed patient identity と全 official version は不変にする。
  4. OpenAPI/codegen、permission rollout、route snapshot、migration immutability/RLS を同期する。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test(DB_ExaminationServiceRejectsPollutedClinicalRelations|.*Examination.*(Confirm|Update|Item|Pet|Unconfirm|Revision|Audit|Rollback|Permission|CrossClinic))' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/apicontract ./internal/clinic ./internal/lintscan -run 'Test.*(Examination|Permission|DefaultPermissionRuleTable|PreloadClinicScope|MasterFKWrite)' -count=1`
  - `docker compose exec -T frontend npx vitest run src/features/examinations src/hooks/use-update-examination.test.tsx src/features/master/components/permission-rule-table-model.test.ts`
  - `make codegen-check`
- **Non-actions / HOLD**: print、lab import revert、clinical range、external format、auto-commit、migration apply、Issue close、claim 削除を行わない。
- **Exit criteria for close**: original confirmed snapshot と patient identity が不変で、unconfirm/re-edit/reconfirm が version/permission/status/actor/audit/correlation contract を満たし、TASK-026、permission rollout、clinic isolation、migration/API regression が green。

#### 実施結果（2026-08-03・TASK-027 saved-prompt run）

- **Outcome**: **COMPLETE / IMPLEMENTED_UNMERGED**。Slice A/B/C と final safety repair を sequential TDD と独立 review で実装し、全指定 test・codegen・go vet は green。migration は未適用、push/merge/Issue close は未実施。
- **Slice reach / commits**: Slice A `046615f4bc923869f189c4e104e27d0539d8c88d`、Slice B `1dd1cf04e77fa7adef38b0230a1b824e4f9abff6`、Slice C `c161baffb2372b8da3195a3b3474f4824f23ada6`、repair `fb0cf9c910aef842fdde1a0206bb5546163096c3`。append-only official/working revision、理由必須 unconfirm、default-deny permission、初回確定前患者変更、結果行 add/delete UI、死亡患者への create/rebind 拒否、items query/readiness と record-local state 分離まで実装済み。
- **Evidence**: medicalrecord revision gate `ok ... 5.834s`、lintscan `ok ... 1.549s`、unconfirm/item/pet gate `ok ... 4.117s`、apicontract/clinic `ok ... 0.029s / 0.369s`、`make codegen-check` exit 0、frontend `160 + 2` tests PASS、repair race/DB/ambient-tx gates PASS、go vet exit 0。最終 Go・clinic isolation・React/a11y・security・clinical safety review は全て APPROVE（CRITICAL/HIGH/MEDIUM 0）。
- **Lint truth**: `GOLANGCI_LINT_VERSION := v2.11.4` の pinned imageで完全件数を実測し、package baseline 62 issues / `EXIT_CODE=1`。Slice B 時点と同数・同分類で新規 finding はなく **baseline-red / delta-green**。repo-wide lint green とは扱わない。
- **Report**: `reports/2026-08-03-task-027-examination-revisions.md`。
- **Git / claim**: `claim/TASK-027` は live のまま。解放は main 統合後の USER 専権。着手時の foreign WIP は本 unitが触れないまま、`ReservationFormModal` 2 files が別ownerの `617f6f9bf`、`bug.md` が別ownerの `fc1cc5a8e` へ入った。
- **HOLD / manual action**: 001 は RLS baseline の該当箇所だけを read-only 確認し、001/002/003 は未変更、seeds は未アクセス。004 は自動適用していない。統合・pull 後にユーザーが `make migrate` を手動実行する。

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
- **状態**: **DONE (agent 2026-08-04 session-a)**。`GET /v1/examinations/:id/print-snapshot` + FE PrintPortal 配線を実装。claim `claim/TASK-031` 取得済・未解放。Issue #249 open のまま。migration なし。scoped tests green。codegen-check は pre-existing models.ts drift（ClinicalPlan audit）のため uncommitted regen あり。
- **claim**: `claim/TASK-031` — **not live**（2026-08-01 USER 解放済み。起票時の過剰取得を是正したもので、本タスクは未着手）。

#### 実装プラン（2026-08-02・adversarial revision 2）
- **Ready**: READY_AGENT_AFTER_TASK-027_INTERFACE_FREEZE。manual unconfirm/revision foundation の API/schema interface だけを先に固定する。clinical range と lab import revert は対象外。
- **Owner lane**: backend examination print snapshot API + frontend print presentation。
- **Fixed contract**: 新 read-only `GET /v1/examinations/:id/print-snapshot?version=` が parent/items/status/version を一つの DTO として単一 SQL snapshot（または read-only repeatable-read transaction）から返す。official owner/other-clinic output は confirmed revision だけ。current non-confirmed は院内用 `DRAFT / 未確定` watermark、confirmed→unconfirm 後は旧 official version と current draft を混ぜない。FE `formItems`、未保存 edit、range/status 再計算、内部 `danger_reason` は含めない。
- **Migration impact**: 本 task 固有の migration は **NO**（TASK-027 revision schema を消費）。read endpoint の OpenAPI/codegen 影響は **YES**。
- **対象ファイル（path:line）**: `backend/internal/medicalrecord/examination_repository.go:84`, `backend/internal/medicalrecord/examination_response.go:36`, `backend/internal/medicalrecord/examination_handler.go:68`, `backend/internal/medicalrecord/routes.go:360`, `backend/docs/api.yaml:3462`, `frontend/src/components/shared/PrintPortal.tsx:45`, `frontend/src/features/examinations/components/ExamPivotTable.tsx:224`, `frontend/src/features/examinations/routes/ExaminationForm.tsx:63`、new print query/DTO/component/model/tests。
- **影響 caller 全数**: examination repository/service/handler/routes/OpenAPI/codegen、`PrintPortal`、`ExaminationForm`、`ExamPivotTable`、feature export。baseline command: `rg -l 'PrintPortal|ExamPivotTable|GetExamination|ListExaminationItems|ExaminationHandler|examinations/:id' backend frontend --glob '*.{go,ts,tsx,yaml}' --glob '!backend/migrations/seeds/**'`。
- **既存テストを RED から拡張**: `examination_repository_test.go`, `examination_handler_test.go`, `routes_snapshot_test.go`, `openapi_examination_mutation_contract_test.go`, `PrintPortal.test.tsx`, `ExamPivotTable.test.tsx`, `ExaminationForm.permissions.test.tsx`、new backend print snapshot tests と `ExaminationPrintArea.test.tsx`。
- **Steps**:
  1. RED: same-clinic confirmed official、other-clinic 404、draft watermark、unconfirmed old-version isolation、parent/items atomic version、unperformed/missing/qualitative/unit/date、no unsaved values を固定する。
  2. GREEN: versioned single DTO と key-scoped query hook、pure print model、`PrintPortal` surface を追加する。
  3. official/draft label と横長 matrix landscape を固定し、owner-facing output から内部 field を除外する。
  4. clean-demo browser/physical print sign-off は USER/QA evidence とし、source green と混同しない。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Examination.*(PrintSnapshot|CrossClinic|Revision)' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/apicontract -run 'Test.*Examination.*Print' -count=1`
  - `docker compose exec -T frontend npx vitest run src/components/shared/PrintPortal.test.tsx src/features/examinations/components/ExamPivotTable.test.tsx src/features/examinations/components/ExaminationPrintArea.test.tsx src/features/examinations/routes/ExaminationForm.permissions.test.tsx`
  - `make codegen-check`
- **Non-actions / HOLD**: clinical assessment recalc、新 range、unconfirm mutation、lab import、実データ screenshot、Issue close、claim 削除を行わない。
- **Exit criteria for close**: print DTO が clinic-scoped atomic revision snapshot で、official/draft/superseded boundary、FE non-recalculation、unsaved-value exclusion、backend other-clinic rejection、scoped tests が green。human sign-off は別 evidence。

### TASK-032: #249 lab import job の compensating revert を examination unconfirm と分離する（Critical / Clinical record integrity）

- **対応 Issue**: GitHub Issue #249 F-3c(a)。
- **問題**: persisted import job を取消す endpoint/状態がなく、手動 examination の確定解除と混ぜると status・permission・audit・rollback の意味が不定になる。
- **状態**: **DONE (agent 2026-08-04)** — code + tests + OpenAPI/codegen + composition + 2 migrations land。**migration 未適用**（USER が `make migrate`）。**claim 未解放**（USER が `git branch -D claim/TASK-032`）。Issue close / push は USER 専権。
- **claim**: `claim/TASK-032` — **live**（実装セッションが再取得・保持。解放は USER 専権。main 統合後に USER が削除）。

#### 実装プラン（2026-08-04・FK-supporting index revision / adversarial revision 2 継承）
- **Ready**: **READY_AGENT / migration review required**。FK/index 欠陥補正 + independent DB review CRITICAL/HIGH ゼロ確認済み（2026-08-04）。MEDIUM は preflight 対象の events 明示・新 table FK 列 NOT NULL・clinic_id NOT NULL の implementer 記述を同 revision で吸収済み。external format/crosswalk と auto-commit enable は対象外。
- **Owner lane**: backend medicalrecord / lab import compensation + additive migration/API contract。
- **Fixed contract**: manual unconfirm と別 endpoint。`persisted -> reverted` は terminal compensation。job を `(clinic_id,id)` で status 非依存 lock 後、active linked exams を ID 順 lockし、exam type、pet-owner、medical-record clinic/pet、doctor active assignment、master FK を fail-closed 検証する。confirmed、finalized medical-record、manual mutation、downstream clinical use、usage unknown のいずれかがあれば全体 409。安全な import-created draft parentだけを条件付き soft deleteし、各 exam の parent+items immutable retraction snapshot と actor/reason を残す。child result hard deleteはしない。
- **Authoritative downstream-use contract**: migration 後に import と同一 transaction で `usage_tracking_started` を記録した job だけ revert eligibility を持つ。legacy job/marker 不在は `usage_unknown` 409。`GetExamination`、`ListExaminationItems`、`GetLabExamReport`、TASK-031 print/export は clinical payload を返す前に append-only usage receipt を記録し、receipt failure は response を返さない。examination/item update、confirm/unconfirm、manual row replace は同一 mutation transaction で `manual_mutation` receipt を記録する。`medical_record_images.exam_id` 等の durable relation も conflict predicate に含める。revert transaction は receipt、application audit、current revision/version、durable relationを lock/readし、errorまたは未計測 consumer は zero-write 409。import 時 snapshot/version と一致し、receipt/relation が全て absent の新規 draftだけを safe とする。
- **Transaction/retry contract**: job/event/exam/retraction/usage-receipt/revert-receipt/audit repository は全て `DBOrTx`、lock method は `TxFromContext` 不在を拒否。順序は job status-independent lock→`(clinic_id,idempotency_key)` revert receipt lookup→既存なら persisted job/payload 比較して same=success・different=409→未処理なら status=`persisted` gate→linked exams→relation/downstream receipt locks→retraction snapshots→RowsAffected照合付き soft delete→CAS job→job event→revert receipt→application audit。同一 key+canonical request は event/audit を増やさず、異 payload は409。
- **Migration impact**: **YES（最低2ファイル）**。開始時に current max+1 を再測定する。第1 migration は `ALTER TYPE ... ADD VALUE IF NOT EXISTS 'reverted'` だけとし同 transaction 内で新値を参照しない。第2 migration は次を追加する。
  - parents/keys: 新 table は全て `clinic_id bigint NOT NULL`。jobs `UNIQUE(clinic_id,id)`（複合 job FK の参照側）、exams は既存 `UNIQUE(clinic_id,id)`（`004_examination_revisions.sql`）に加え `UNIQUE(clinic_id,id,job_id)`（exam+job 複合 FK の参照側）、retractions `UNIQUE(clinic_id,id,job_id,exam_id)`。
  - FKs（全て `ON DELETE RESTRICT`。子側 soft delete は app 概念であり、PG 参照整合性検査は `deleted_at` を見ず物理行を対象とする。**nullable は exams.job_id のみ** — retractions / retraction items / usage receipts / revert receipts の複合 FK 列は全て `NOT NULL`。MATCH SIMPLE で一部 NULL の行が RI をすり抜けないようにする）:
    1. **FK-E1** events `(clinic_id,job_id) -> jobs(clinic_id,id)`
    2. **FK-X1** exams `(clinic_id,job_id) -> jobs(clinic_id,id)`（現行 single-col `job_id ON DELETE SET NULL` を置換）
    3. **FK-R1** retractions `(clinic_id,job_id) -> jobs(clinic_id,id)`
    4. **FK-R2** retractions `(clinic_id,exam_id,job_id) -> exams(clinic_id,id,job_id)`
    5. **FK-RI1** retraction items `(clinic_id,retraction_id,job_id,exam_id) -> retractions(clinic_id,id,job_id,exam_id)`
    6. **FK-U1** usage receipts `(clinic_id,job_id) -> jobs(clinic_id,id)`
    7. **FK-U2** usage receipts `(clinic_id,exam_id,job_id) -> exams(clinic_id,id,job_id)`
    8. **FK-V1** revert receipts `(clinic_id,job_id) -> jobs(clinic_id,id)`
  - receipts: append-only usage receipts と revert receipts。revert receipt は `UNIQUE(clinic_id,idempotency_key)` と FK-V1 + canonical request/result fields、usage receipt は FK-U1/U2 + use-kind/actor/time を持つ。
  - RLS: jobs/events/exams/retractions/retraction-items/receipts 全表で project helper による `ENABLE ROW LEVEL SECURITY` と clinic predicate の `USING`/`WITH CHECK` を明示適用・検証する。既存 project posture に合わせ `FORCE ROW LEVEL SECURITY` は role compatibility review なしに追加せず、全 repository clinic predicate を必須のまま維持する。
  - **indexes（FK-supporting と query を分離。FK-supporting は soft-deleted child を除外しない）**:
    - **RI 原則**: PostgreSQL の `ON DELETE/UPDATE RESTRICT` 検査は child の物理行を対象にする。soft delete（`deleted_at IS NOT NULL`）行も FK 値が非 NULL なら検査対象。よって FK-supporting index に `WHERE deleted_at IS NULL` を付けない。nullable FK のみ `WHERE <fk> IS NOT NULL` を許容する（MATCH SIMPLE で全 FK 列 NULL の行は検査対象外のため、検査対象を除外しない）。
    - **FK-supporting（必須・1対1）**:
      1. FK-X1 exams: `(clinic_id, job_id) WHERE job_id IS NOT NULL` — soft-deleted linked exams を含む。現行 `idx_exams_clinic_job` と同形を複合 FK 置換後も RI 用として維持/置換する。**`deleted_at` 述語は付けない**。
      2. FK-E1 events: `(clinic_id, job_id, created_at, id)` 非部分（events は append-only・`deleted_at` 無し想定）。
      3. FK-R1 retractions: `(clinic_id, job_id, exam_id, id)` 非部分 — leading `(clinic_id, job_id)` が R1 を支える。
      4. FK-R2 retractions: `(clinic_id, exam_id, job_id)` 非部分 — **FK 列順と一致**（R1 用 index の列順入れ替えでは left-prefix にならないため別 index）。
      5. FK-RI1 retraction items: `(clinic_id, retraction_id, job_id, exam_id)` 非部分 — 複合 FK 全列。list 用に `(clinic_id, retraction_id, id)` を追加してもよいが RI の代替にしない。
      6. FK-U1 usage receipts: `(clinic_id, job_id, exam_id, created_at, id)` 非部分 — leading `(clinic_id, job_id)`。
      7. FK-U2 usage receipts: `(clinic_id, exam_id, job_id)` 非部分 — **FK 列順と一致**（U1 用 index とは別）。
      8. FK-V1 revert receipts: `(clinic_id, job_id, created_at, id)` 非部分。
    - **query/active 用（FK 検査用ではない）**:
      - active exams `(clinic_id, job_id, id) WHERE deleted_at IS NULL AND job_id IS NOT NULL` — 業務 list/lock 用。`deleted_at IS NULL` は active 行だけを引く意図であり、**RI は FK-X1 の index が支える**ため本部分 index の述語は FK 検査対象を除外しても問題ない。
    - retractions/receipts は UPDATE/DELETE rejection trigger（append-only）。
  - fail-closed preflight（composite FK 追加前）: 複合 job/exam FK を得る**既存 child 全表**を対象にする — 最低限 `exams` と `lab_import_events`（現行 single-col `job_id` FK のため `child.clinic_id ≠ job.clinic_id` や orphan `job_id` が残り得る）。不整合時は migration を停止する。新 table は legacy 行なし。既存 exams の `job_id ON DELETE SET NULL` は preflight 通過後に composite/RESTRICT へ置換する。enum removal は通常 rollback 不可のため application rollback + forward corrective migration。既適用 migration/seed は編集せず、agent は apply しない。
- **対象ファイル（path:line）**: `backend/internal/model/lab_import.go:11`, `backend/internal/medicalrecord/lab_import_service.go:21`, `backend/internal/medicalrecord/lab_import_repository.go:3`, `backend/internal/medicalrecord/lab_import_handler.go:136`, `backend/internal/medicalrecord/routes.go:373`, `backend/internal/model/examination_record.go:32`, `backend/cmd/api/composition_medicalrecord_repositories.go:126`, `backend/cmd/api/composition_medicalrecord_services.go:126`, `backend/cmd/api/composition_medicalrecord.go:126`, `backend/docs/api.yaml:7481`, `backend/internal/apicontract/openapi_examination_mutation_contract_test.go:1`, `backend/internal/lintscan/dbortx_inventory_lint_test.go:1`、次番号 migration と test DB enum fixture。
- **影響 caller 全数**: job/event/examination persistence、usage/revert receipt、examination detail/items、lab report、TASK-031 print/export、medical-record image relation、application audit、service/handler/routes/API composition、transition/retraction response、OpenAPI、route snapshot、test DB/RLS/migration/lint inventory。baseline command: `rg -l 'LabImportJobStatus|LabImportService|LabImportRepository|LabImportExamination|TransitionStatus|NewLabImport|lab-import|job_id|GetExamination|ListExaminationItems|GetLabExamReport|GetExamReport|PrintSnapshot|exam_id' backend frontend --glob '*.{go,ts,tsx,yaml}' --glob '!backend/migrations/seeds/**'`。downstream anchors: `backend/internal/medicalrecord/examination_handler.go:68`, `backend/internal/medicalrecord/examination_handler.go:143`, `backend/internal/medicalrecord/lab_report_handler.go:48`, `backend/internal/medicalrecord/lab_report_query_service.go:32`, `backend/internal/medicalrecord/medical_record_image_service.go:149`, examination update/items/confirm services、TASK-031 print endpoint。
- **既存テストを RED から拡張**: `backend/internal/medicalrecord/lab_import_service_test.go`, `backend/internal/medicalrecord/lab_import_repository_test.go`, `backend/internal/medicalrecord/lab_import_handler_test.go`, `backend/internal/medicalrecord/lab_import_examination_service_test.go`, `backend/internal/medicalrecord/examination_parent_audit_tx_test.go`, `backend/internal/medicalrecord/examination_handler_test.go`, `backend/internal/medicalrecord/lab_report_query_service_test.go`, `backend/internal/medicalrecord/medical_record_image_service_test.go`, `backend/internal/medicalrecord/routes_snapshot_test.go`, `backend/internal/apicontract/openapi_examination_mutation_contract_test.go`, `backend/internal/lintscan/dbortx_inventory_lint_test.go`。新規 transaction/integration test で atomic revert、clinic isolation、usage receipt failure、legacy/unknown use、manual mutation、detail/items/report/print exposure、durable relation conflict を固定する。
- **Steps**:
  1. RED: same-clinic instrumented-unused success、wrong/corrupt clinic-record-pet-doctor-assignment-master、same job UUID in other clinic、orphan/mismatched existing job link、confirmed/finalized/manual mutation/detail-items-report-print exposure/durable relation/legacy usage-unknown conflict、receipt failure zero-response/zero-revert、idempotent/conflicting replay、concurrent persist/read/revert、audit/retraction failure rollback を固定する。
  2. Competing persist path を含む全 repository を ambient transaction/lock orderへ統一し、base DB fallbackを拒否する。
  3. 2段 additive migration、RLS/composite FK/index/idempotency/retraction snapshot、status/event/API response を同期する。
  4. `POST /v1/lab-imports/:id/revert` を `lab-import:edit` + reason + idempotency key で追加し、conditional RowsAffected と audit atomicity を固定する。
  5. OpenAPI、route/test DB/RLS/migration/lint inventory を同期し、retracted fact は履歴で可視、confirmed/finalized record と child result は変更ゼロを確認する。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*LabImport.*(Revert|Transition|Audit|Rollback|Clinic|Concurrent)' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/apicontract ./internal/lintscan -count=1`
  - `docker compose exec -T backend go test -p 1 ./cmd/api -run 'Test.*LabImport' -count=1`
- **Non-actions / HOLD**: confirmed examination の自動解除、child result hard delete、external format/crosswalk、auto-commit enable、seed edit、migration apply、Issue close、claim 削除を行わない。
- **Exit criteria for close**: caller census と番号が再計測済みで、manual unconfirm と別 endpoint/status/permission、confirmed/finalized/manual/downstream/unknown conflict 409、clinical response 前 usage receipt、retraction snapshot、status-independent idempotent retry、DBOrTx/lock/RLS/FK/RowsAffected/audit atomicity、clinic isolation、migration/API regression が green。
- **Plan revision log (2026-08-04)**: FK-supporting index 欠陥を補正。active-only `(clinic_id,job_id,id) WHERE deleted_at IS NULL` を RI 用と誤認しないよう、FK-X1 に soft-deleted 行を含む `(clinic_id,job_id) WHERE job_id IS NOT NULL` を必須化。FK-R2/U2 に FK 列順一致 index、FK-RI1 に複合 FK 全列 index を追加。independent review: database-reviewer CRITICAL=0 HIGH=0 MEDIUM=2（events preflight / 新 table FK NOT NULL）→ 同 revision で吸収; clinic-isolation-auditor CRITICAL=0 HIGH=0 MEDIUM=2（clinic_id NOT NULL 明示 / FORCE RLS  defer は project posture として許容）→ clinic_id NOT NULL を parents/keys に明記。他設計（transaction/idempotency/RLS/receipt）は adversarial revision 2 を継承。状態を READY_AGENT / migration review required へ更新。

#### 実施記録（2026-08-04 implementation unit）
- **Delivered**:
  - migrations: `007_lab_import_job_status_reverted.sql`（enum ADD VALUE only）、`008_lab_import_revert_compensation.sql`（複合 FK/index/receipts/retractions/RLS/preflight）
  - endpoint: `POST /api/v1/lab-imports/:job_id/revert`（`lab-import:edit` + reason + Idempotency-Key）
  - state machine: `persisted → reverted` terminal compensation（TransitionStatus からは到達不可）
  - usage receipts before clinical payload（detail/items/lab report/print snapshot）+ manual mutation in mutation tx
  - atomic terminal: production composition で status transition + `usage_tracking_started` を同一 `WithTx`
  - OpenAPI `api.yaml` + codegen `frontend/src/types/generated/models.ts` + route smoke 504
- **Scoped gates**（agent 実行）:
  - `go test -p 1 ./internal/medicalrecord -run 'Test.*LabImport.*(Revert|Transition|...)'` → PASS
  - `go test -p 1 ./cmd/api` → PASS（route count 504）
  - `TestDBOrTxInventory_MatchesAllowlist` → PASS
  - `make codegen-check`（models staged）→ PASS
  - full `./internal/lintscan` は pre-existing ERD marker 115/116 drift と seed CSV drift（checkup_types/exams）で FAIL — TASK-032 起因ではない
- **Independent reviews**（round ≤2）:
  - database / clinic-isolation / healthcare / security: PASS（CRIT=0 HIGH=0）— plan/implementation session
  - go-reviewer round1: FAIL HIGH=1（marker/tx + clinical TOCTOU）→ fix
  - go-reviewer recheck: **PASS CRIT=0 HIGH=0**（atomic terminal + LogRevertSucceeded + usage lock order）
- **USER handoff**:
  1. `make migrate`（007→008 の順で適用）
  2. main 統合後 `git branch -D claim/TASK-032`
  3. push / Issue close は USER 判断
- **Non-actions honored**: migration apply なし、seed 編集なし、unconfirm 経路変更なし、child result hard delete なし、TASK-033 未着手

### TASK-033: #201 active/draft 構造化救急投薬記録 + 欠落時 fail-closed cutover（Critical / Clinical safety）

- **対応 Issue**: GitHub Issue #201。
- **問題**: current addendum は finalized medical record 専用の自由記述で、薬剤、実投与量・単位、投与時刻を構造化せず、active/draft の救急・既実施投薬を通常治療履歴と handoff に残す代替経路ではない。代替経路なしに体重/species/parameter 欠落時の通常保存だけを止めると、救急記録を失う。
- **状態**: **BLOCKED_CLINICAL_INPUT_AND_DECISION_SOT_RECONCILIATION_AND_DATABASE_REVIEW / migration review required**。最終 fail-closed 契約と cutover 順序は DEC-48。臨床責任者の未記入欄、`q&a.html#issue-readiness-current-p0-20260801` の #201 current row に対する authorized append-only correction、treatment parent FK/child index と history/FK-supporting indexes の plan 補正が全て揃うまで実装を開始しない。
- **claim**: `claim/TASK-033` — **not live**（2026-08-01 USER 解放済み。起票時の過剰取得を是正したもので、本タスクは未着手）。

#### 実装プラン（2026-08-02・adversarial revision 2）
- **Ready**: BLOCKED_CLINICAL_INPUT_AND_DECISION_SOT_RECONCILIATION_AND_DATABASE_REVIEW。下記 conditional contract、caller census、migration、cutover、tests は draft 済みだが、final DB review が treatment→medical-record FK/child index、treatment history index、event medicine/actor FK indexes を未充足と判定した。臨床責任者の一行承認、decision SoT に TASK-025 実装済み範囲/TASK-033 HOLD を併記する authorized correction、bounded plan revision と fresh DB review の全てが揃った後だけ `READY_AGENT / migration review required` へ移す。本 unit は `q&a.html` 編集禁止のため correction を代行しない。
- **Owner lane**: backend medicalrecord + frontend TreatmentsTab/history + identity-link handoff + dedicated permission + migration/API contract。
- **未確定入力（臨床責任者だけ）**: 対象とする救急・既実施投薬ケース、既知 medicine master を必須にするか限定的な未同定薬 snapshot を許すか、route vocabulary、dose/strength/concentration の unit と requiredness、投与時点の体重/species snapshot policy、reason taxonomy または bounded free-text rule、訂正対象と rationale、create permission を受ける role/group、臨床出典、承認者、発効日。値・語彙・ケースは本台帳にも実装にも補完しない。
- **Fixed event superset**: `emergency_medication_administrations` は `id, clinic_id, medical_record_id, pet_id, medicine_id(nullable), medicine_name_snapshot(nonblank), actual_dose(positive), dose_unit, route, administered_at, reason, strength/concentration snapshot value+unit(nullable), weight/species snapshot(nullable), performed_attestation(true), actor_id, source_treatment_id(nullable), correction_of_id(nullable), correction_reason, idempotency_key, created_at` を持つ append-only fact。臨床承認が未同定薬を禁じる場合は service が `medicine_id` 必須を強制し、許す場合も bounded case と name snapshot を必須にする。未来時刻は service で reject し、volatile `now()` CHECK は置かない。`updated_at/deleted_at`、UPDATE/DELETE route/repository は作らない。
- **API/permission**: record-scoped GET、initial POST、correction POST を別 route にする。client は clinic/pet/actor を送らず handler が authenticated actor を必須取得する。GET は `medical-records:view`、POST は既存 medical-record permission と新しい `emergency-medication-administrations:create` の双方を要求する。専用 resource は全 default group で default-deny。fresh/demo fixture は disposable DB + sanctioned `seed-export` のみ、既存 clinic grant は named approver 後に permission-group API/app で行い、seed replay しない。
- **Tenant/relational DDL**: candidate key と child FK は clinic-first の同一順序で固定する。
  - parents: medical records `UNIQUE(clinic_id,id,pet_id)`、medicines `UNIQUE(clinic_id,id)`、staff assignments `UNIQUE(clinic_id,staff_id)`、treatments `UNIQUE(clinic_id,id,medical_record_id,pet_id)`、events `UNIQUE(clinic_id,id,medical_record_id,pet_id)`。
  - treatment backfill: current treatments に `clinic_id`/`pet_id` を additive に追加し、medical record から fail-closed backfill・相関 preflight 後に NOT NULL/composite unique/RLS を付ける。全 treatment create/update/delete/reorder は client 値でなく locked record から clinic/pet を設定・照合する。
  - event FKs: `(clinic_id,medical_record_id,pet_id) -> medical_records`、nullable `(clinic_id,medicine_id) -> medicines`、`(clinic_id,actor_id) -> staff_clinic_assignments`、nullable `(clinic_id,source_treatment_id,medical_record_id,pet_id) -> treatments`、nullable `(clinic_id,correction_of_id,medical_record_id,pet_id) -> events`。全て `ON DELETE RESTRICT`。
  - RLS/immutability: treatment/event と新規表に project helper の `ENABLE ROW LEVEL SECURITY` + clinic `USING`/`WITH CHECK` を明示適用し、repository predicate も必須。current project posture と role compatibility review を経ず `FORCE` は追加しない。events は UPDATE/DELETE rejection trigger、`UNIQUE(clinic_id,idempotency_key)`、`UNIQUE(clinic_id,correction_of_id) WHERE correction_of_id IS NOT NULL`、positive/nonblank/correction parity/no-self-correction CHECK を持つ。
- **Transaction/idempotency**: 全 repository は `DBOrTx`、transaction-required method は `TxFromContext` 不在を拒否する。全 competing path の global lock order は medical record→treatments（ID昇順）→active staff assignment→medicine（指定時）→event correction target。既存 treatment create/update/delete/reorder も record を先に lockし、source treatment を後から lockするよう同 unit で揃える。clinic/draft/pet-owner/actor/medicine/treatment/correction の全相関を検証し、event と mandatory application audit を atomic にする。`ON CONFLICT DO NOTHING RETURNING` 後に persisted normalized fields を比較し、same clinic/key+same payload は既存成功・audit zero、異 payload は 409。Preload した current medicine を履歴 snapshot の代用にしない。
- **History/handoff**: existing mutable `/treatments` の billing/inventory/total/update/delete へ混ぜず、`source_treatment_id` で order/billing fact と performed-event fact を明示的に対応させ、臨床表示の二重計上を防ぐ。訂正後の effective projection は superseded event を current clinical fact から除外し、監査 view では replacement relationship を返す。通常 pet history と identity-link history は同一 CTE/`UNION ALL`（discriminator 付き）を count/page の唯一の入力にし、各 branch で `(clinic_id, pet_id)` を相関し、stable keyset `(occurred_at,id,kind)` で bounded pagination する。indexes は event record history `(clinic_id,medical_record_id,administered_at DESC,id DESC)`、pet history `(clinic_id,pet_id,administered_at DESC,id DESC)`、source link `(clinic_id,source_treatment_id,id)`、effective correction lookup `(clinic_id,correction_of_id,id)`。owner-facing output に内部 danger reason を出さない。
- **Missing-data cutover**: backend decision を `ready / not_applicable / missing_medicine / missing_weight / missing_species / missing_parameter` に型分離し、technical failure は error のまま。`missing_medicine` は上記 clinical medicine-identity policy が明示的に許す場合だけ event path を提示し、それ以外は zero-write で block。missing は理由別 409 と zero treatment/inventory/audit write、pending と technical failure も FE save を止める。非薬剤/非体重式だけ not_applicable。event create/read/correction、permission、history、audit、clinical approval が全て green になった同一 TASK-033 unit でだけ反転し、partial cutover を禁止する。
- **Migration impact**: **YES**。実装開始時に root inventory を再測定して次番号の additive migration を追加し、treatment clinic/pet fail-closed backfill、上記 clinic-first candidate keys/FKs、RLS `USING`/`WITH CHECK`、explicit indexes、trigger、idempotency/correction constraints を含める。permission catalog/fixture 影響も **YES**。既適用 migration/seed の手編集・自動 apply はしない。pull 後は USER が `make migrate`。
- **対象ファイル（path:line）**:
  - backend core: `backend/internal/model/medical_record_addendum.go:5`, `backend/internal/medicalrecord/medical_record_addendum_service.go:84`, `backend/internal/medicalrecord/treatment_dose_save.go:14`, `backend/internal/medicalrecord/treatment_service.go:1`, `backend/internal/medicalrecord/treatment_repository.go:1`, `backend/internal/medicalrecord/treatment_handler.go:75`, `backend/internal/medicalrecord/treatment_response.go:72`, `backend/internal/medicalrecord/routes.go:272`, `backend/internal/model/permission.go:6`, `backend/internal/model/audit_log.go:98`, `backend/cmd/api/composition_medicalrecord_repositories.go:126`, `backend/cmd/api/composition_medicalrecord_services.go:126`, `backend/cmd/api/composition_medicalrecord.go:126`, `backend/docs/api.yaml:1`。
  - history/isolation: `backend/internal/identitylink/repository.go:533`, `backend/internal/identitylink/service.go:761`, `backend/internal/identitylink/handler.go:307`, `backend/internal/identitylink/routes.go:27`, `backend/internal/identitylink/types.go:1`, `backend/internal/clinic/clinic_service.go:1`。
  - frontend: `frontend/src/features/medical-records/components/MedicalRecordFormPanels.tsx:387`, `frontend/src/features/medical-records/components/MedicalRecordTreatment.tsx:17`, `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx:1`, `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx:1`, `frontend/src/features/medical-records/components/TreatmentsTab/treatment-row-dose-gate.ts:20`, `frontend/src/features/owner-report/api/get-pet-treatment-history.ts:18`, `frontend/src/features/owner-report/components/TreatmentHistorySection.tsx:1`, `frontend/src/lib/query-keys.ts:1`, `frontend/src/features/master/components/permission-rule-table-model.ts:15`。
  - new: next-number migration、event model、request/response/repository/service/handler、effective-history projection、focused tests、FE API/component/tests。
- **影響 caller 全数**: treatment create/update/dose gate、`ListPetTreatmentHistory`、`ListLinkedTreatmentHistory` の handler/service/repository/response、medicalrecord/identity-link routes、API composition、OpenAPI/codegen、permission defaults/labels、TreatmentsTab mount、owner report mapper/table。implementation baseline で再実行する: `rg -l 'TreatmentsTab|useGetTreatments|useCreateTreatment|useUpdateTreatment|useDeleteTreatment|useReorderTreatments|FindHistoryByPetID|ListPetTreatmentHistory|ListLinkedTreatmentHistory|NewTreatmentServiceWithAudit|TreatmentRepository|DoseGateSource' frontend/src backend/internal backend/cmd --glob '*.{go,ts,tsx}' --glob '!backend/migrations/seeds/**'`。
- **既存テストを RED から拡張**: `backend/internal/medicalrecord/treatment_dose_save_test.go`, `backend/internal/medicalrecord/treatment_service_test.go`, `backend/internal/medicalrecord/treatment_repository_test.go`, `backend/internal/medicalrecord/treatment_handler_test.go`, `backend/internal/medicalrecord/routes_snapshot_test.go`, `backend/internal/identitylink/repository_integration_test.go`, `backend/internal/identitylink/service_test.go`, `backend/internal/clinic/clinic_service_test.go`, `backend/internal/apicontract/openapi_examination_mutation_contract_test.go`, `backend/internal/lintscan/dbortx_inventory_lint_test.go`, `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.test.tsx`, `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.test.tsx`, `frontend/src/features/medical-records/components/TreatmentsTab/treatment-row-dose-gate.test.ts`, `frontend/src/features/owner-report/api/get-pet-treatment-history.test.ts`。
- **Cutover order**:
  1. 臨床責任者が未確定入力を一行承認し、authorized owner が decision SoT の TASK-025/TASK-033 状態を append-only correction で調停する。両方の evidence 後に実装側は `claim/TASK-033` と isolated worktree を取得する。
  2. RED: treatment backfill/preflight、clinic-first schema/RLS/FK/immutability、same/wrong clinic-record-pet、active assignment、primary-clinic mismatch+valid assignment、medicine policy、future time、actor/audit rollback、idempotent replay/conflict、source-treatment lock order、concurrent treatment/event mutation、correction/effective projection、normal/identity-link count+page isolation、permission default-deny、missing zero-write を固定する。
  3. Event migration/backend/API/UI/history/permission を current missing behavior のまま実装し、migration/database/clinic-isolation/healthcare/Go/React/TypeScript review を通す。
  4. Event create/read/correction、source-treatment reconciliation、通常・identity-link handoff を green にする。
  5. 同一 TASK-033 unit 内でのみ FE/BE missing state を理由別 fail-closed へ反転し、全 scoped gate と no-partial-cutover reconciliation を行う。
  6. agent は migration を適用せず、USER へ `make migrate` を handoff する。
- **Verification** (scoped only):
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*(EmergencyMedicationAdministration|DoseMissing|DoseSpecies|DoseParameter|PetTreatmentHistory|StaffAssignment|EffectiveProjection)' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/identitylink -run 'Test.*LinkedTreatmentHistory' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/clinic -run 'Test(DefaultPermissionRuleTable|DemoPermissionGroupRules)' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/model -run 'Test.*(EmergencyMedicationAdministration|RLS|CompositeFK)' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/apicontract ./internal/lintscan -count=1`
  - `docker compose exec -T frontend npx vitest run src/features/medical-records/components/TreatmentsTab src/features/owner-report/api/get-pet-treatment-history.test.ts`
  - `make codegen-check`
- **Non-actions / HOLD**: 臨床上限・warning 値、route/unit/case の発明、臨床承認前または構造化経路 green 前の missing cutover、generic free-text/addendum 代用、既適用 migration/seed edit、自動 migration apply、DB 操作、Issue close、claim 削除を行わない。
- **Exit criteria for close**: 上記 clinical input と decision-SoT correction が承認済みで、event が必須 field、clinic/pet/record/medicine-or-approved-snapshot/active-actor/treatment/correction 相関、dedicated permission、audit atomicity、immutable correction/effective projection、bounded and count-consistent history/handoff を満たし、missing-data 通常 write が理由別 zero、partial cutover がなく全 scoped gate が green。

### TASK-374: #211 健診 package の versioned clinic-scoped import/preflight（High / Tenant safety）

- **対応 Issue**: GitHub Issue #211。
- **問題**: `checkup_types` と `checkup_type_fields` は clinic-scoped だが、provisional 定義は shared environment で禁止された `003_demo` に属する。table 単位の `seed-export` は承認済み subset を安全な別 bundle へ出せず、臨床承認だけでは実 clinic へ原子的に反映できない。
- **状態**: **DONE (agent 2026-08-04 session-a / synthetic)**。migration `006_checkup_package_import.sql` 作成（未適用）、manifest/canonicalizer、preview/apply endpoint、default-deny permission、provenance+fail-closed audit を実装。claim `claim/TASK-374` 取得済・未解放。Issue #211 open のまま。migration apply / real manifest は USER-only。scoped tests green。
- **claim**: 実装セッションが編集前に `claim/TASK-374` を確認・取得する。本裁定 run では claim を取得しない。
- **Owner lane**: backend checkup master import/preflight + transaction-bound audit + scoped docs/tests。実データ、migration apply、GitHub write は USER-only。
- **独立 gate**: DR-CLINICAL は臨床値/単位、出典、臨床承認者 role、発効日、opaque clinical-row/approval reference だけを所有する。DR-OPS は target clinic authorization、environment、DB history、operator role、dry-run/apply/rollback の結果 enum と opaque restricted reference だけを所有する。manifest/stable key、実 identity、receipt/audit 本文は repo 外に置き、片方の承認を他方へ流用しない。

#### 固定契約
- `003_demo`、`002_master`、既適用 bundle/CSV を変更しない。shared environment に `003_demo` を load しない。
- repo 外の versioned manifest は namespace/version、type/field の stable key と関係、opaque clinical approval reference を表す。manifest に clinic ID・actor ID を持たせず、repository fixture は synthetic value だけを使う。
- **additive migration は必須**: type/field に nullable import namespace/key を追加し、`(clinic_id, import_namespace, import_key)` の partial UNIQUE と clinic-first index を持たせる。clinic-scoped import provenance/receipt は namespace/version、canonical content digest、actor、status、件数、作成 resource ID mapping を access-controlled DB 内に耐久化し、`UNIQUE (clinic_id, namespace, version)`、RLS `USING`/`WITH CHECK`、clinic-first FK/index を持つ。digest/mapping は operator output、application log、Git 台帳へ返さない。
- 同 migration で `checkup_types(parent_id, clinic_id) -> checkup_types(id, clinic_id)` の複合 FK へ置換し、`ON DELETE SET NULL (parent_id)` と `(clinic_id, parent_id)` index を維持する。field→type の既存複合 FK と順序を合わせる。
- apply は authenticated request context の clinic/actor を唯一の authority とし、manifest/CLI の ID を認可根拠にしない。actor の active clinic assignment、既存 `ResourceCheckups` create/edit、default-deny の専用 import permission を transaction 内で確認する。unauthorized と foreign-clinic は同じ非漏洩 error surface で write zero とする。
- dry-run は domain write zero。apply は prior dry-run を信用せず、transaction 内で clinic row lock → authorization → receipt/stable-key collision → active/soft-deleted name collision → parent/type/field clinic correlation →全 field validation を再実行してから type→field の決定的順序で write する。DB UNIQUE/FK も通常 CRUD との競合を拒否し、type/field、receipt、audit のいずれかが失敗すれば全 write を rollback する。
- manifest は unknown field を拒否する厳格 schema とし、stable-key 順、UTF-8/trim、decimal(10,4) 文字列表現、null/empty、options の順序・重複規則を固定して canonical digest を計算する。number は options 空かつ min≤max、select/checklist は非空で value 一意、boolean/text は min/max/options 空を検証する。
- version は immutable import artifact とする。同じ clinic/namespace/version/content は no-op、同 version の異内容、既存 stable key の異内容、旧 version の自動更新/置換、同名 supersession は conflict とし、別 DEC なしに mutation しない。apply 対象 field は `is_provisional=false` と clinical approval reference を必須にし、true/missing は拒否する。
- post-commit rollback 実行は scope 外だが、作成 resource IDs を provenance に残し、read-only rollback preflight が患者記録依存を報告できるようにする。使用済み type/field/result を hard-delete しない。
- **sink 分離**: access-controlled provenance DB は internal actor/clinic ID、namespace/version/digest、resource mapping、status/counts を持つ。別 schema の access-controlled application audit は internal actor/clinic ID と before/after/resource を持つ。operator receipt DTO は opaque receipt ID・件数・結果のみ、application log は receipt ID・結果のみ、Git 台帳は非機密結果 enum・role label・opaque restricted-evidence reference のみを持つ。

#### 実装・検証
1. RED: invalid schema/version、unknown/foreign clinic、wrong-clinic parent、別院/inactive/permission不足 actor、auth-clinic 不一致、duplicate key/name、invalid field matrix、partial/audit/receipt failure、same/different replay、concurrent CRUD/apply、dry-run zero domain write を固定する。receipt/log の allowlist test で actor/clinic/digest/resource mapping/before/after が外部出力されないことも固定する。
2. manifest parser/canonicalizer/validator と DB-independent preflight を小さな package に置き、repository/service は `persistence.DBOrTx` と明示 clinic predicate を必須にする。
3. authenticated preview/apply endpoint、dedicated default-deny permission、apply transaction、audit/provenance/opaque receipt を composition root へ配線する。apply endpoint は明示 action と actor context を必須にし、値入り sample は作らない。
4. 次番号の additive migration と migration inventory/test を追加し、migration-seed-safety、database、clinic-isolation、healthcare、Go/security review を通す。既適用 migration/seed は編集せず、agent は適用しない。
- **Scoped verification**:
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*CheckupPackageImport' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*CheckupPackageImport.*(CrossClinic|Actor|Permission|Concurrent|Rollback)' -count=1`（PostgreSQL-backed、2 clinic・同一 stable key を含む）
  - `docker compose exec -T backend go test -p 1 ./cmd/api -run 'Test.*CheckupPackageImport' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/model ./internal/lintscan -count=1`
- **Non-actions / HOLD**: 実 row/clinic/price/range の決定、real manifest の repository 保存、shared `003_demo` load、CSV 手編集、既適用 bundle/migration edit、DB apply/reset、Issue close、claim 解放を行わない。
- **Exit criteria**: synthetic manifest の canonicalization、dry-run/apply/replay/conflict、actor-clinic authorization、2-clinic isolation、concurrency、audit/provenance rollback、migration/FK/index/RLS tests が green で、実値なしの operator runbook が完成し、USER に real manifest・target authorization・apply/rollback window の restricted inputだけを handoff できる。

### TASK-375: #257 期限切れ go-live runbook の gate-driven 再計画（High / Docs-only）

- **対応 Issue**: GitHub Issue #257。
- **問題**: `docs/delivery/GOLIVE_RUNBOOK.md` は失効した 2026-08-03 window と 2026-07-18 timeline を併存させ、open prerequisite と未確定 authority/support があるのに当日手順として読める。
- **状態**: **DONE (docs sync 2026-08-04 session-b)**。`docs/delivery/GOLIVE_RUNBOOK.md` を historical No-Go（2026-08-03）+ 相対 T-/T+ timeline + 新 window 一箇所記入欄 + fail-closed prerequisites（#89/#97/#98/#99/#250/#253/#254/#255 + authority/support/rollback）へ同期。値（日付・人名・contact）は空欄のまま。claim `claim/TASK-375` 取得済・未解放。Issue #257 は open のまま。
  - claim: `git branch --list 'claim/TASK-375'` → empty → `git branch claim/TASK-375` exit 0
  - gates: `rg -n '2026-08-03|No-Go|2026-07-18|T-[0-9]|T\+[0-9]|#89|#97|#98|#99|#250|#253|#254' docs/delivery/GOLIVE_RUNBOOK.md` 目視 PASS; `git diff --check -- docs/delivery/GOLIVE_RUNBOOK.md` exit 0
  - Assumption: 新 window 記入欄は冒頭 1 箇所へ集約（既存分散 blank は §4/§5 の 確定待ち を残しつつ冒頭表を SSOT 化）
- **claim**: 実装セッションが編集前に `claim/TASK-375` を確認・取得する。本裁定 run では claim を取得しない。**実装 session-b が claim 取得済み。解放は USER 専権。**
- **Owner lane**: `docs/delivery/GOLIVE_RUNBOOK.md` と、その日付/ownershipを直接参照する delivery docs の最小同期。
- **Steps**:
  1. 失効 window を historical No-Go と明記し、実行可能な current window として表示しない。
  2. timeline を `T-` / `T+` 相対時刻へ変換し、具体日時は全 prerequisite の named evidence と Go/No-Go authority が揃った後だけ USER が一箇所へ記入する。
  3. #89/#97/#98/#99、#250、#253、#254、#255、authority/support/rollback owner を fail-closed prerequisite として参照し、旧 AWS 系を rollback 先にしない。
  4. credential、contact、日付、人名、provider 実値を補完せず、外部実行・deploy・DB 操作を行わない。
- **Scoped verification**: `rg -n '2026-08-03|2026-07-18|T-[0-9]|T\+[0-9]|Go/No-Go|rollback' docs/delivery/GOLIVE_RUNBOOK.md` の目視突合、`git diff --check -- docs/delivery/GOLIVE_RUNBOOK.md`。
- **Non-actions / HOLD**: 新 window/owner/contact の決定、production/deploy/DB/credential/GitHub 操作、旧 infrastructure の復活、Issue close、claim 解放を行わない。
- **Exit criteria**: stale window が実行指示から除外され、相対 timeline と全 gate が一意で、USER が新 window と named owners を一行記入するまで fail-closed と読める。

### TASK-376: #258 U1〜U12 と #256 U13 の delivery 文書境界同期（Medium / Docs-only）

- **対応 Issue**: GitHub Issue #258（U13 owner は #256）。
- **問題**: `q&a.html`/current view は #258 を U1〜U12、U13 を #256 の納品後研修とする一方、`docs/delivery/DELIVERY_PACKAGE.md` の冒頭は U1〜U13 全てを #258 final approval 条件として読める。二重 blocker と ownership drift が残る。
- **状態**: **DONE (docs sync 2026-08-04 session-b)**。`DELIVERY_PACKAGE.md` の #258 completion/approval を U1〜U12 に限定し U13 行を除去。U13 日程・形式・参加者・実施 receipt は `OPERATION_MANUAL.md` §10/#256 に一意集約。値は空欄のまま。claim `claim/TASK-376` 取得済・未解放。Issue #258 は open のまま。
  - claim: `git branch --list 'claim/TASK-376'` → empty → `git branch claim/TASK-376` exit 0
  - gates: `rg -n 'U1|U12|U13|#256|#258|研修' docs/delivery/DELIVERY_PACKAGE.md docs/delivery/OPERATION_MANUAL.md` 目視 PASS; `git diff --check -- docs/delivery/DELIVERY_PACKAGE.md docs/delivery/OPERATION_MANUAL.md` exit 0; `U1–U13` 残存 0
  - Assumption: OPERATION_MANUAL §10 は既存。新規節は作らず §10 に記入欄（日程/形式/参加者/receipt）を追加
- **claim**: 実装セッションが編集前に `claim/TASK-376` を確認・取得する。本裁定 run では claim を取得しない。**実装 session-b が claim 取得済み。解放は USER 専権。**
- **Owner lane**: `docs/delivery/DELIVERY_PACKAGE.md`、`docs/delivery/OPERATION_MANUAL.md` と直接参照する delivery docs の最小同期。
- **Steps**:
  1. #258 の document completion/approval input を U1〜U12 に限定し、各値の正本を既存表のまま維持する。
  2. U13 の日程・形式・参加者・実施 receipt を #256 / operation manual の納品後研修へ一意に移し、#258 close blocker として二重管理しない。
  3. credential/API key、価格、契約、billing、実人名、日付を補完せず、空欄と secret-manager 状態参照を維持する。
  4. q&a/view の受入条件を docs へ複製せず、Issue/DR anchor と owner 境界だけを同期する。
- **Scoped verification**: `rg -n 'U1|U12|U13|#256|#258|研修' docs/delivery/DELIVERY_PACKAGE.md docs/delivery/OPERATION_MANUAL.md` の目視突合、`git diff --check -- docs/delivery/DELIVERY_PACKAGE.md docs/delivery/OPERATION_MANUAL.md`。
- **Non-actions / HOLD**: U1〜U13 の値決定、credential/価格/契約/実 identity の記録、外部サービス操作、Issue close、claim 解放を行わない。
- **Exit criteria**: #258=U1〜U12、#256=U13 が全参照で一意になり、契約責任者/Client と training owner の一行入力先が分離される。

### TASK-377: #201 warning 逸脱理由を FE/BE・snapshot・同一 transaction audit へ接続する（Critical / Clinical safety）

- **対応 Issue**: GitHub Issue #201。
- **問題**: live Issue は warning 範囲で inline warning と必須逸脱理由、FE bypass 時の BE 同一境界を要求する。しかし `computeDoseGate` が返す `requiresConfirm/reason` を `TreatmentRow` が消費せず、上限内の著しい乖離は表示なし・理由なしで保存される。BE request/service に理由 input がなく、既存 deviation audit も clinician-entered reason を持たない。
- **状態**: **DONE / no migration（2026-08-04）**。claim `claim/TASK-377` 取得済み（解放は USER 専権）。設計正本は DEC-65。20% 閾値は未変更。上限値・warning 帯・taxonomy・TASK-033 は臨床 gate のまま。
- **claim**: `claim/TASK-377`（実装 run が取得。解放は USER）。
- **Owner lane**: medicalrecord treatment create/update request・service・dose snapshot/audit、TreatmentsTab quantity editor、OpenAPI/types、focused tests。TASK-033、dose master 値、migration、GitHub write は所有しない。

#### 実施記録（2026-08-04）
- **理由契約**: JSON field `dose_deviation_reason`、trim 後 1〜500 Unicode、error は固定文言 `InvalidInput` / technical は actor・nil audit・strict marshal。snapshot に `deviates_from_computed` + `dose_deviation_reason`。audit NewValue に flags + reason + actor。露出は staff treatment snapshot / audit のみ（owner history・validation error・slog に本文なし）。
- **BE**: `RequiresDeviationReason`、write 前 `ensureDoseDeviationAuditReady`、reason-required は strict marshal、safe は best-effort 維持、stale reason 除去。
- **FE**: `requiresConfirm` → `requiresDeviationReason`、inline 理由 UI（modal なし）、空理由 mutation 0、idempotent 1 回送信、create 経路は reason-required で create 停止。
- **Scoped gates**:
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Treatment.*Dose.*(DeviationReason|Audit|Rollback|Clinic)' -count=1` → ok
  - `docker compose exec -T backend go test -p 1 ./internal/apicontract -run 'Test.*Treatment.*Dose.*Reason' -count=1` → ok
  - FE vitest TreatmentRow + dose-gate → 23 passed
  - `make codegen-check` → FAIL（pre-existing TASK-374 models drift。本 unit は request field を手書き型 + OpenAPI で同期。generated models は未変更コミット）
- **Review**: healthcare/security/go → CRITICAL/HIGH 0。react HIGH（Enter+blur 二重送信）を idempotency + dirty-only reason UI で修正後 FE green。typescript CRITICAL/HIGH 0。
- **migration 本数**: 6（開始時と同一）。

#### 固定契約
- 上限超過は従来どおり hard reject、technical failure は通常保存停止、missing data は TASK-033 cutover まで現行 contract を維持する。上限内で `BelowMinSaved || DeviatesFromComputed` の時だけ、trim 後 1〜500 Unicode 文字の generic free-text `dose_deviation_reason` を create/update の必須入力にする。これは transport の技術上限であり、臨床 taxonomy の決定ではない。modal と confirm-to-bypass は作らない。
- FE は `currentGate.reason` を色だけに依存しない inline warning として表示し、理由が空なら quantity mutation を送らない。誤解を招く `requiresConfirm` は reason-required contract として型・呼出し・test を同期し、著しい乖離で `warning="none"` でも理由 UI を表示する。
- BE は FE 判定を信用せず、既存 `EvaluateSavedDose` の locked effective state で同じ条件を再評価する。直接 API、create/update、空白理由・500文字超過理由を同じ validation に通し、理由不足、authenticated actor 欠落、snapshot serialization、audit のいずれかが失敗したら treatment/inventory/audit write を同一 transaction で zero/rollback にする。reason-required 保存では optional actor、`auditTx == nil` を technical failure とし、後方互換の feature flag にしない。既存 best-effort snapshot marshal は reason-required 経路で error-returning に変え、serialization 失敗を黙って成功させない。
- normalized reason と `deviates_from_computed` を既存 `dose_param_snapshot` に値で固定し、access-controlled dose-deviation audit に actor・評価 flags・reason を残す。safe dose への再評価で stale reason を残さない。application log、owner-facing history、Git 台帳へ理由本文を出さない。既存 JSONB 列を使うため schema migration は追加しない。
- audit fields から warning 件数と理由記録率を算出可能にするが、本 task で外部 telemetry、臨床 threshold、reason taxonomy、例外 override を追加しない。

#### 実装・検証
1. RED: 上限内の著しい乖離と下限割れで inline reason が表示され、空理由では mutation zero、理由付きで1回だけ送信される RTL を追加する。上限超過、technical failure、missing、safe dose の既存契約も固定する。
2. RED: create/update の理由欠落・blank・500文字境界/超過・direct API bypass、理由付き成功、snapshot/audit 内容、missing authenticated actor、nil audit dependency、audit failure rollback、forced snapshot marshal failure、safe re-evaluation の stale reason 除去、clinic isolation を table-driven test で固定する。missing actor、nil audit、marshal failure は treatment/inventory/audit write が全て zero であることを直接検証する。sentinel 理由文字列が owner history、identity-link response、validation error、`slog` に出ない negative contract test も追加する。
3. request/service/snapshot/audit と FE type/editor/API を最小接続し、backend を最終 authority にする。OpenAPI と生成型に field を同期し、既存 treatment response 以外へ理由を露出しない。
4. healthcare、security、Go、React、TypeScript の独立 review を通し、理由本文が非 audit log/owner surface に漏れないことと、TASK-033 contract を変更していないことを再照合する。
- **Scoped verification**:
  - `docker compose exec -T backend go test -p 1 ./internal/medicalrecord -run 'Test.*Treatment.*Dose.*(DeviationReason|Audit|Rollback|Clinic)' -count=1`
  - `docker compose exec -T backend go test -p 1 ./internal/apicontract -run 'Test.*Treatment.*Dose.*Reason' -count=1`
  - `docker compose exec -T frontend npx vitest run src/features/medical-records/components/TreatmentsTab/TreatmentRow.test.tsx src/features/medical-records/components/TreatmentsTab/treatment-row-dose-gate.test.ts`
  - `make codegen-check`
- **Non-actions / HOLD**: 上限値・warning threshold・taxonomy・救急適用条件の決定、TASK-033 missing-data cutover、通常 mutation override、migration/DB apply、Issue close、claim 解放を行わない。
- **Exit criteria**: warning/deviation の全 create/update 経路が理由なし zero-write、理由付き時だけ authenticated actor、snapshot、transaction-bound audit を伴って成功し、FE は inline で理由を要求して modal を使わず、直接 API bypass・missing actor・nil audit dependency・audit failure・snapshot serialization failure・clinic crossing の regression が green。残る委任外判断は DR-CLINICAL の #201 bundle 1 行だけ。

---
