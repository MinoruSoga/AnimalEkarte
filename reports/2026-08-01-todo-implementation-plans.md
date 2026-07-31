# todo.md 実装プラン・マスターマトリクス（2026-08-01）

## 実行単位

- Unit: `TODO-MD-IMPLEMENTATION-PLANS-ORCH-20260801`
- Baseline HEAD at packet start: `8a97a5696c1c1360c99743e2b22a87e3d314e131`
- Verification HEAD: `239a8a736ffb8ba41288b6626c730b2639cb7e07`。調査中にforeign ownerの2 commitが`main`へlandしたため、成果物はこのHEADでも再検証する。
- Primary ledger: `todo.md`
- Scope: planning / investigation only。製品コード、migration、seed、DB、外部LINE、credentialは変更しない。
- Readyの意味: `READY_USER` は「USERが最初のpreconditionから開始できるrunbookがある」であり、完了・runtime greenを意味しない。

## Live truth corrections

| Surface | Historical text | Live result | Planning effect |
|---|---|---|---|
| TASK-009 static seed gate | slice1 reportではstatic PASS | 4対象CSVは各3行だが、`python3 scripts/verify_seed.py` は既存`treatments`欠落、appointment時刻分布、vaccination参照でexit 1 | DB apply前に別packetでstatic REDを解消。TASK-009はBLOCKED |
| TASK-010 census | exact body `【要実測】` 59 | exact markerは59、S08/S09の装飾variantを含むsemantic prefix `【要実測` は62 | 両censusを固定し、3 decorated marksを取りこぼさない |
| TASK-019 Phase B report | composition residual記載あり | reachable commit `fac8c86b2` と現testで解消済み | residualから除外し、rollout/inventory/DROPだけHOLD |
| claim text | TASK-009/022/023/024等のclaim記載あり | liveに存在するtask claimはTASK-010/020/021。TASK-009/022/023/024は absent | historical textは保持し、開始判断はlive branchを使う |
| CODEX navigation guide | supplementが参照 | `docs/CODEX-NAVIGATION-GUIDE.md` はlive treeに存在しない | missing pathを発明せず、本repoのAGENTS/CLAUDE/rulesを使用 |
| shared-tree status | packet開始時に複数foreign WIP、途中で一度docs-onlyへ収束 | verification中もstatusが変動し、owned `todo.md` / report以外のforeign WIP（観測例: `backend/migrations/seeds/003_demo/estimates.csv`, `bug.md`）を検出 | final statusを正本とし、foreign pathは非stage・非変更のまま保持。land直前にも再実測 |

## TASK master matrix

| ID | Ready | Owner lane | Top blocker today | First command / action | Detailed plan |
|---|---|---|---|---|---|
| TASK-004 | READY_USER | ops-only | intentional land set未確定、foreign WIPあり | `git status --porcelain=v1` | `todo.md` TASK-004 plan |
| TASK-005 | READY_USER | ops-only | land対象/staged set未確定 | `bash scripts/check-docs-symbol-drift.sh` | `todo.md` TASK-005 plan |
| TASK-009 | BLOCKED(`verify_seed.py` FAIL・USER apply) | AGENT→USER | seed全体static RED、DB apply証跡なし | `python3 scripts/verify_seed.py` | `todo.md` TASK-009 plan |
| TASK-010 | BLOCKED(`claim/TASK-010`・seed/runtime前提) | AGENT→USER | live claim、seed依存、59/62 census | `git branch --list 'claim/TASK-010'` | `todo.md` TASK-010 plan |
| TASK-019 | READY_USER | ops-only | production/deploy evidence、R-05 inventory | `git branch --list 'claim/LINE-*'` | `todo.md` TASK-019 plan |
| TASK-020 | READY_USER | USER | credential未注入、live claims | `test -n "${E2E_LOGIN_EMAIL:-}" && test -n "${E2E_LOGIN_PASSWORD:-}"` | `todo.md` TASK-020 plan |
| TASK-021 | BLOCKED(external use UNREPORTED・破壊承認・claim) | AGENT→USER | external利用ゼロ/破壊承認なし | `git branch --list 'claim/TASK-021'` | `todo.md` TASK-021 plan |
| TASK-022 | READY_USER | USER | S13 named sign-off、real app-role RLS proof | S13手順1-8をnamed operatorへ割当 | `todo.md` TASK-022 plan |
| TASK-023 | READY_USER | USER | E2E credentials、QA/LINE/PO evidence | credentialをsecret channelで注入 | `todo.md` TASK-023 plan |
| TASK-024 | READY_USER | USER | clean-seed 05/07/10再撮影、named sign-off | clean-seed撮影windowを割当 | `todo.md` TASK-024 plan |

Ready summary: **READY_AGENT 0 / READY_USER 7 / BLOCKED 3**。

## Open index / ops-only matrix

| Index ID | Ready | Owner | Top blocker / trigger | First command or checklist | Plan disposition |
|---|---|---|---|---|---|
| R4 | READY_USER | ops-only | next intentional land | `git status --porcelain=v1` | TASK-004 |
| R5 | READY_USER | ops-only | staged set確定後 | `bash scripts/check-docs-symbol-drift.sh` | TASK-005 |
| R6 | READY_USER | ops-only | shared treeの複数editor | `git worktree list` | separate worktree / one shared-tree writer |
| R7 | READY_USER | ops-only | empty-diff誤宣言 | `git diff --name-only` | actual diff必須 |
| SCEN-SEED-001 | BLOCKED | AGENT→USER | static seed FAIL・USER apply | `python3 scripts/verify_seed.py` | TASK-009 |
| SCEN-BROWSER-001 | BLOCKED | AGENT→USER | `claim/TASK-010`・runtime prerequisites | `git branch --list 'claim/TASK-010'` | TASK-010 |
| SCEN-OPS-CLAIM-001 | READY_USER | ops-only | integration/abandon evidence | `git branch --list 'claim/*'` | USER only release。agent delete禁止 |
| SCEN-OPS-COMMIT-001 | READY_USER | ops-only | mixed history説明時 | `git status --porcelain=v1` | memoのみ。rewrite/force-push禁止 |
| SCEN-OPS-TREE-001 | READY_USER | ops-only | concurrent WIP | `git worktree list` | R6と同じ |
| ARCH-R2 | READY_USER | ops-only | empty-diff success | `git diff --name-status` | R7と同じ |
| ARCH-R3 | READY_USER | ops-only | land直前のforeign定義 | `git status --porcelain=v1` | TASK-004 |
| POST-PULL | READY_USER | ops-only | migration変更を含むpull | USERが変更有無を確認後`make migrate` | agent auto-apply禁止 |
| SPEC-TOP-LINE-AUDIT | READY_USER | ops-only | R-02/R-04/R-05/R-08 ops evidence | `git branch --list 'claim/LINE-*'` | TASK-019 |
| SPEC-TOP-E2E-RUNTIME-84 | READY_USER | USER | credential未注入 | value-free env preflight | TASK-020（93-test runtime） |
| SPEC-TOP-CAPABILITIES-CRUD | BLOCKED | AGENT→USER | external use UNREPORTED / approval | `git branch --list 'claim/TASK-021'` | TASK-021 Stage A |
| SPEC-TOP-CLAIM-RELEASE | READY_USER | ops-only | main integration/abandon evidence | `git branch --list 'claim/*'` | SCEN-OPS-CLAIM-001 |

## Common HOLD / non-actions

- `make migrate`, `make reset`, `make db`, `DB_RESET`, direct `psql`はagent実行禁止。
- production LINE/L-step、real LIFF、credential/secret値の読取・記録はUSER/ops境界。
- `001_init.sql`や適用済みmigrationの直接編集、CSV手編集、証拠前のDROP/CLEAN-GOは禁止。
- `git add -A`, history rewrite, force-push, foreign WIP discard、claim branch deleteは禁止。
- full repository test/lint/buildは自動実行せず、各planのDocker-scoped commandだけを使う。

## Evidence and command inventory

- Prompt validation: `node ~/.claude/scripts/prompt-craft-harness-validate.js --json --target agent --require-dynamic-workflow /Users/minoru/.claude/prompt-craft-runs/agent-fast-todo-implementation-plans.md` → exit 0、failed checksなし。
- Scenario census: exact `【要実測】` = 59、semantic prefix `【要実測` = 62。
- Seed static gate: `python3 scripts/verify_seed.py` → exit 1。TASK-009 apply greenではない。
- Relevant live claims: `claim/TASK-010`, `claim/TASK-020`, `claim/W-020-ENV`, `claim/TASK-021`, `claim/W-021-P1`, `claim/LINE-R-FIX`, `claim/LINE-PO-R01-R05`, `claim/LINE-PARENT-RBAC`, `claim/TODO-MD-IMPLEMENTATION-PLANS-ORCH-20260801`。
- All paths directly linked by current TASK sections were found. Referenced commits `c286bfe0a`, `79fe62265`, `3d448ec5e`, `a1abd4db8`, `e9dddd921`, `254fdc2f3`, `e17e749ec`, `fac8c86b2` are locally reachable.

## Manual wave ledger

Native Workflow was not available. Execution used multi-agent read-only fan-out, joined draft packets, one coordinator writer, then scoped verification/review.

| Wave | Labels | Ownership | Status | Integration decision |
|---|---|---|---|---|
| Wave 0 | RO-PARSE / RO-OPS / RO-SEED / RO-RUNTIME / RO-LINE / RO-021 / RO-HUMAN / RO-CONV | read-only evidence clusters; no file ownership | joined | all evidence reconciled; 59/62 distinction and static seed RED retained |
| Wave 1 | PL-OPS / PL-ENV / PL-LINE-021 / PL-H22 | read-only plan drafts; no file ownership | joined | root normalized template, Ready flags, and scoped commands |
| Wave 2 | W-LEDGER / W-REPORT / V-SCOPE | root only: `todo.md`, this report。V-SCOPEはread-only | joined / verified | structure、matrix、anchor、Ready、tracking、non-action gatesはPASS。whole-tree `git diff --name-only` allowlistだけはforeign WIPのためBLOCKEDとして保持。V-READY-2指摘は再実測して変動snapshotを反映し、先行V-READYは明示cancel済み |

Loop-status capability was unavailable; this table is the required manual wave ledger.
