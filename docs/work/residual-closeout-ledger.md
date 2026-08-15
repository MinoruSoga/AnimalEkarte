# Residual Closeout Ledger

> **Role:** Live residual-closeout run state only (not a second product backlog).  
> **SoT for product residual:** root `todo.md` + `todo-po.md`.  
> **Policy:** `docs/work/decisions/fable-po-recommendation.md`.  
> **Docs snapshot:** 2026-08-13 — SoT を todo / todo-po に統合。  

## Frozen unit order (do not re-order without USER)

| Unit | Scope | External write? | Status |
|------|--------|-----------------|--------|
| **U0** | Tree hygiene + claim release report | No (local write) | **COMPLETE** (2026-08-08) |
| **U1** | PO-01 · GitHub #98 close path | Yes | **COMPLETE** (2026-08-08) — #98 CLOSED |
| **U2** | PO-02 · GitHub #99 close path | Yes | **COMPLETE** (2026-08-08) — #99 CLOSED |
| **U3** | PO-03 · #252 ↔ #257 go-live gate Yes/No | Yes (or STATUS one-liner) | **COMPLETE** (2026-08-08) — gate include #252=YES (STATUS + #257 comment) |
| **U4** | PO-06 · TASK-023 / #254 five-flow UAT | Yes (evidence) | **COMPLETE** (2026-08-08) — RECORD_ONLY overall=PARTIAL; #254 not closed |
| **U5** | TASK-024 / #256 screenshot-FAQ sign-off | Yes (evidence) | **COMPLETE** (2026-08-08) — dual SIGNED_OFF overall=PASS; #256 not closed |
| **U6** | TASK-022 / #239 S13 / RLS evidence | Yes (evidence) | **COMPLETE** (2026-08-09) — LEDGER_PO s13=NOT_RUN rls=N_A_APP_LAYER_ONLY; #239 not reopened |
| **PO-07** | TASK-021-B client registry inventory | No (local BOTH_LOCAL) | **COMPLETE** (2026-08-09) — declaration=NO_KNOWN_EXTERNAL_CONSUMERS; B/C/D DROP not performed |
| **UAT-R1** | TASK-023 / #254 residual f3–f5 re-record | No (local STATUS_LEDGER) | **COMPLETE** (2026-08-09) — RECORD_ONLY overall=PARTIAL; f3–f5 unchanged from U4; #254 not closed |

**HOLD** (separate decision required): TASK-021 B/C/D product DROP (registry evidence recorded; B still needs PO approve before delete; C needs PO-08 access log), LINE-R05 DROP, TASK-033/#201 clinical, #249 clinical ranges.

**Residual series U0–U6:** frozen order complete (no U7). **Post-series:** PO-07 client registry recorded 2026-08-09; **UAT-R1** residual f3–f5 re-record 2026-08-09 (does not invent U7 product unit).

---

## U0 — Tree hygiene + claim release report

| Field | Value |
|-------|--------|
| Date | 2026-08-08 |
| Branch / HEAD | `main` @ `264777c882c86d6ad5f9d7c3d22f047588fc23fc` |
| Run status | **COMPLETE** |
| Next unit | **U1 = PO-01 · #98** (do not auto-chain) |

### Working-tree decisions

| Path | Classification | Action |
|------|----------------|--------|
| `.gitignore` (+ `.code-review-graph/`) | Keep — local tooling ignore; not on `origin/main` | local commit (no push) |
| `backend/internal/billing/accounting_complete_appointments_test.go` | Noise — import order only (`testdb`/`reservation` swap) | path-scoped `git restore` |
| `backend/internal/billing/accounting_repository_tx_atomicity_test.go` | Noise — same import order only | path-scoped `git restore` |
| `reports/` (incl. `reports/bug-md-2agent-loop/`) | Local loop artifacts | leave **untracked**; do not stage; not gitignored (optional future ignore is PO-side memo only) |

### Claim ancestry (re-measured)

| Claim | Tip | `merge-base --is-ancestor … main` | Release |
|-------|-----|-------------------------------------|---------|
| `claim/ARCH-A4` | `26380f0b2` | exit **0** (ancestor) | USER: `git branch -D claim/ARCH-A4` |
| `claim/ARCH-A7` | `4b349d796` | exit **0** (ancestor) | USER: `git branch -D claim/ARCH-A7` |

Agent **did not** delete claim branches.

### Gate commands (verbatim evidence in Completion Report)

1. `git status -sb` / `git diff --stat` — inventory
2. Path-scoped restore of two billing tests → empty diff for those paths
3. Claim ancestry exit 0 for A4/A7
4. Post-commit: `git status -sb` shows only `?? reports/` (and no claim deletes)

### Notes for PO

- `reports/` remains untracked by design for this unit; adding a permanent `.gitignore` rule is optional and not required for U0.
- Also present local branch (not in U0 release set): `chore/bugmd-loop-driver`.
- U0 does **not** authorize starting U1 in the same session.

---

## U1 — PO-01 · GitHub #98 close path

| Field | Value |
|-------|--------|
| Date | 2026-08-08 |
| Branch / HEAD | `main` @ `77979604d` (local ahead 1 of origin/main) |
| Run status | **COMPLETE** (close executed after USER approval) |
| Next unit | **U2 = PO-02 · #99** (do not auto-chain) |

### Pre-write evidence (re-measured this session)

| Check | Result |
|-------|--------|
| `gh issue view 98 --json number,title,state` | `{"number":98,"state":"OPEN","title":"🔴 CRITICAL: 旧RDS credential履歴の残余リスクと廃止スクリプト撤去"}` |
| Comment count | 7; latest 2026-07-31 PARTIAL — residual risk USER judgment still open; **no** Fable close one-liner (enum + opaque_ref) posted |
| Fable F-098 | USER-adopted 2026-08-06: `ACCEPT_RESIDUAL_RISK` with provider invalidation non-secret one-liner then close |
| PO-todo PO-01 | still `- [ ]` (unchecked) |
| STATUS #98 row | still open / USER 専権 · ACCEPT_RESIDUAL_RISK path |
| Tree: `scripts/stg-db-tunnel.sh` | **MISSING** (exit 1 on `test -f`); no `*tunnel*` under scripts |
| Password pattern scan | scripts/infra hits classified as env names / guards / fixtures only; **no** plaintext password reintroduction reported (values not printed) |

### Close comment draft (placeholders only — not posted)

```text
PO-01 / F-098 close: enum=<ENUM> opaque_ref=<OPAQUE_REF>
Policy: ACCEPT_RESIDUAL_RISK (fable-po-recommendation F-098). Role split: rotation execution remains #89/#97; this issue records residual-history risk acceptance after non-secret provider/operator confirmation. No secrets in this comment.
```

Allowed enums: `PROVIDER_INVALIDATED` | `RDS_INSTANCE_GONE` | `ACCEPT_RESIDUAL_RISK_WITH_OPAQUE_REF`  
opaque_ref: non-secret ticket/console check/dated operator attestation id only.

### Blocker (required USER input)

1. **Provider confirmation enum** (one of allowed set above)
2. **opaque_ref** (non-empty, non-secret)
3. After draft review: **explicit approval sentence** to run `gh issue comment 98` + `gh issue close 98`

Agent did **not** invent enum/ref, did **not** post, did **not** close, did **not** rotate credentials, did **not** rewrite history, did **not** push.

### Gate commands (this session)

1. `gh issue view 98 --json number,title,state` → OPEN (verbatim above)
2. `gh issue view 98 --json ... comments` → 7 comments; no close one-liner
3. `test -f scripts/stg-db-tunnel.sh` → exit 1 (absent)
4. Probes A/B/C fan-out joined (main agent supplied live `gh` for Probe A capability gap)
5. Independent Review Gate on **posted** payload: **deferred** until filled draft exists (no external write)

### Out of scope honored

- No #89/#97 rotation, no filter-repo, no U2 start, no product code change, no secrets in ledger

### Notes for PO / next session

- Resume: USER supplies enum + opaque_ref → security-reviewer/santa on exact body → USER approval → post+close → flip U1 COMPLETE; then optional PO-todo/STATUS one-liners
- **U2 must not start** until U1 is COMPLETE or USER explicitly re-scopes

### Resume attempt (2026-08-08) — `fast-residual-closeout-u1-resume-issue98.md`

| Field | Value |
|-------|--------|
| Prompt | `~/.claude/prompt-craft-runs/fast-residual-closeout-u1-resume-issue98.md` |
| Re-measure `gh issue view 98 --json number,title,state` | `{"number":98,"state":"OPEN","title":"🔴 CRITICAL: 旧RDS credential履歴の残余リスクと廃止スクリプト撤去"}` |
| Tunnel script | still absent (`test -f` exit 1) |
| USER enum | **missing** (not in session message) |
| USER opaque_ref | **missing** |
| USER external-write approval sentence | **missing** |
| Filled body | not produced (would invent values — forbidden) |
| Independent Review Gate | not run on post body (no fill) |
| `gh issue comment` / `gh issue close` | **not executed** |
| U1 status after resume | still **BLOCKED** |
| Next unit | still **U2** (do not auto-start) |

**Exact missing fields to unblock (paste in next session):**
1. `enum=` one of `PROVIDER_INVALIDATED` | `RDS_INSTANCE_GONE` | `ACCEPT_RESIDUAL_RISK_WITH_OPAQUE_REF`
2. `opaque_ref=` non-secret id only
3. After agent shows filled draft + review PASS: explicit sentence approving `gh issue comment 98` + `gh issue close 98`

### Hard-gate attempt (2026-08-08) — `fast-residual-closeout-u1-hardgate-issue98.md`

| Field | Value |
|-------|--------|
| Hard gate enum | **PASS** — `ACCEPT_RESIDUAL_RISK_WITH_OPAQUE_REF` (USER-supplied, allowed) |
| Hard gate opaque_ref | **PASS** — `2026-08-08-PO-attestation-F098` (non-secret attestation shape) |
| Issue re-measure | still OPEN (`gh issue view 98` → state OPEN) |
| Filled body | prepared (see below); not posted |
| Independent Review (Santa B+C / security-reviewer) | both **PASS** (zero CRITICAL/HIGH) |
| External-write approval sentence | **MISSING** in launch message |
| `gh issue comment` / `gh issue close` | **not executed** (safety boundary) |
| U1 status | still **BLOCKED** on approval only |
| Next unit | still **U2** (do not auto-start) |

**Filled body (review PASS; awaiting USER approval to post):**
```text
PO-01 / F-098 close: enum=ACCEPT_RESIDUAL_RISK_WITH_OPAQUE_REF opaque_ref=2026-08-08-PO-attestation-F098
Policy: ACCEPT_RESIDUAL_RISK (fable-po-recommendation F-098). Role split: rotation execution remains #89/#97; this issue records residual-history risk acceptance after non-secret provider/operator confirmation. No secrets in this comment.
```

**Unblock (one sentence only):**
`I approve posting the non-secret close comment and closing GitHub Issue #98.`

### Close execution (2026-08-08) — USER approval received

| Field | Value |
|-------|--------|
| Approval | USER: `I approve posting the non-secret close comment and closing GitHub Issue #98.` |
| `gh issue comment 98` | exit 0 → https://github.com/MinoruSoga/AnimalEkarte/issues/98#issuecomment-5225913483 |
| `gh issue close 98` | exit 0 · reason completed |
| Re-measure | `{"number":98,"state":"CLOSED","title":"🔴 CRITICAL: 旧RDS credential履歴の残余リスクと廃止スクリプト撤去"}` |
| last_comment | matches filled body (enum + opaque_ref non-secret) |
| U1 status | **COMPLETE** |
| Next unit | **U2 = PO-02 · #99** — not started |
| Local docs | PO-todo PO-01 checked; STATUS #98 row removed (close-table rule) |

**Posted body (verbatim):**
```text
PO-01 / F-098 close: enum=ACCEPT_RESIDUAL_RISK_WITH_OPAQUE_REF opaque_ref=2026-08-08-PO-attestation-F098
Policy: ACCEPT_RESIDUAL_RISK (fable-po-recommendation F-098). Role split: rotation execution remains #89/#97; this issue records residual-history risk acceptance after non-secret provider/operator confirmation. No secrets in this comment.
```

---

## U2 — PO-02 · GitHub #99 close path

| Field | Value |
|-------|--------|
| Date | 2026-08-08 |
| Branch / HEAD | `main` (local ahead of origin; residual docs dirty) |
| Run status | **COMPLETE** (close executed after USER approval) |
| Next unit | **U3 = PO-03** (do not auto-start) |
| Prompt | `~/.claude/prompt-craft-runs/fast-residual-closeout-u2-po02-issue99.md` |

### Hard gate (launch message)

| Field | Value |
|-------|--------|
| enum | **PASS** — `WORKFLOW_ABSENT` (allowed) |
| opaque_ref | **PASS** — `2026-08-08-PO-attestation-F099` (non-secret dated attestation) |
| external-write approval sentence | **PASS** — USER: `I approve posting the non-secret close comment and closing GitHub Issue #99.` |

### Pre-write evidence (re-measured before close)

| Check | Result |
|-------|--------|
| `gh issue view 99 --json number,title,state` | was OPEN: `{"number":99,"state":"OPEN","title":"🔴 HIGH: 廃止予定ECS deploy経路の撤去と現行rollback手順の一本化"}` |
| `gh issue view 253 --json number,title,state` | `{"number":253,"state":"OPEN","title":"[DELIVERY] 本番環境整備 — CI/CD・監視・DB backup/restore gate"}` (SoT; not closed) |
| `test -f .github/workflows/backend-deploy-ecs.yml; echo $?` | exit **1** (absent) |
| active workflows | CF-oriented names; **no** `*ecs*` workflow filename |
| Fable F-099 | APPROVE path-zero one-liner + rollback SoT=#253 → close; Phase-8 mass delete separate |
| Independent Review | Santa B+C **PASS** (NICE) on exact filled body |

### Close execution (2026-08-08) — USER approval received

| Field | Value |
|-------|--------|
| Approval | USER: `I approve posting the non-secret close comment and closing GitHub Issue #99.` |
| `gh issue comment 99` | exit 0 → https://github.com/MinoruSoga/AnimalEkarte/issues/99#issuecomment-5226073407 |
| `gh issue close 99` | exit 0 · reason completed |
| Re-measure | `{"number":99,"state":"CLOSED","title":"🔴 HIGH: 廃止予定ECS deploy経路の撤去と現行rollback手順の一本化"}` |
| last_comment | matches filled body (enum + opaque_ref + rollback_sot=#253) |
| U2 status | **COMPLETE** |
| Next unit | **U3 = PO-03** — not started |
| Local docs | PO-todo PO-02 checked; STATUS #99 row removed (close-table rule) |

**Posted body (verbatim):**
```text
PO-02 / F-099 close: enum=WORKFLOW_ABSENT opaque_ref=2026-08-08-PO-attestation-F099 rollback_sot=#253
Policy: APPROVE path-zero confirmation (fable-po-recommendation F-099). Rollback SoT is Issue #253 (single source; no second rollback SSOT). Role split: #99 confirms old ECS path is not an executable deploy route; Phase-8 mass delete remains separate if still needed. No secrets in this comment.
```

### Out of scope honored

- No Phase-8 ECS/terraform mass delete, no ECS path revival, no #98 re-open, no #89/#97 rotation, no U3 start, no secrets in artifacts, no push

---

## U3 — PO-03 · #252 ↔ #257 go-live gate Yes/No

| Field | Value |
|-------|--------|
| Date | 2026-08-08 |
| Branch / HEAD | `main` (local residual docs dirty; ahead of origin) |
| Run status | **COMPLETE** (STATUS + #257 comment after USER approval; no issue close) |
| Next unit | **U4 = PO-06 · TASK-023 / #254** (do not auto-start) |
| Prompt | `~/.claude/prompt-craft-runs/fast-residual-closeout-u3-po03-gate252.md` |

### Hard gate (launch message)

| Field | Value |
|-------|--------|
| decision | **PASS** — `INCLUDE_252_IN_GOLIVE_GATE` → DECISION_WORD=`YES` |
| write_target | **PASS** — `BOTH` |
| opaque_ref | **PASS** — `2026-08-08-PO-attestation-F257-gate252` (non-secret) |
| external-write approval sentence | **PASS** — USER: `I approve posting the non-secret PO-03 decision comment on GitHub Issue #257.` |

### Pre-write evidence (re-measured)

| Check | Result |
|-------|--------|
| `gh issue view 252 --json number,title,state` | `{"number":252,"state":"OPEN","title":"[OPS] 各院の締め時間設定値の投入 — 全院を城東と同値で確定投入（PO裁定 2026-07-15）"}` |
| `gh issue view 257 --json number,title,state` | `{"number":257,"state":"OPEN","title":"[OPS] 本番切替（Go-live 2026-07-27）— 切替手順書・切り戻し基準・直後サポート体制"}` |
| F-257 | HOLD new window; #252 gate add = USER Yes/No (this unit freezes YES) |
| PO-03 | complete = #257 or STATUS one-liner; final Yes/No = USER |
| Independent Review | Santa B+C **PASS** (NICE) on exact filled body |

### Comment execution (2026-08-08) — USER approval received

| Field | Value |
|-------|--------|
| Approval | USER: `I approve posting the non-secret PO-03 decision comment on GitHub Issue #257.` |
| `gh issue comment 257` | exit 0 → https://github.com/MinoruSoga/AnimalEkarte/issues/257#issuecomment-5226172248 |
| Re-measure #252 | still OPEN (not closed by this unit) |
| Re-measure #257 | still OPEN (not closed by this unit) |
| last_comment | matches filled body (go-live_gate_include_#252=YES + opaque_ref) |
| U3 status | **COMPLETE** |
| Next unit | **U4 = PO-06 · TASK-023 / #254** — not started |
| Local docs | STATUS §2 #257 one-liner; PO-todo PO-03 checked; PO-20 unchanged |

**Posted body (verbatim):**
```text
PO-03 / F-257: go-live_gate_include_#252=YES opaque_ref=2026-08-08-PO-attestation-F257-gate252
Policy: Fable F-257 HOLD on new go-live window (not set in this unit). This line only freezes whether #252 (clinic closing-time settings) is a go-live prerequisite gate. #252/#257 remain open for their own execution work. No secrets in this comment.
```

### Out of scope honored

- No new go-live window date, no #252 value inject, no #252/#257 close, no U4 start, no push, no secrets

---

## U4 — PO-06 · TASK-023 / #254 five-flow UAT

| Field | Value |
|-------|--------|
| Date | 2026-08-08 |
| Branch / HEAD | `main` (local residual docs dirty; ahead of origin) |
| Run status | **COMPLETE** (RECORD_ONLY result line + #254 comment after USER approval; no issue close) |
| Next unit | **U5 = TASK-024 / #256** (do not auto-start) |
| Prompt | `~/.claude/prompt-craft-runs/fast-residual-closeout-u4-po06-task023-uat.md` |

### Hard gate (launch message)

| Field | Value |
|-------|--------|
| mode | **PASS** — `RECORD_ONLY` |
| overall | **PASS** — `PARTIAL` (consistent with f3=FAIL + f4=UNTESTED + f5=BLOCKED) |
| f1–f5 | **PASS** — f1=PASS f2=PASS f3=FAIL f4=UNTESTED f5=BLOCKED |
| opaque_ref | **PASS** — `2026-08-08-PO-uat-TASK-023` (non-secret) |
| write_target | **PASS** — `BOTH` |
| memo | stack local demo; f3 trimming fail TBD; f4/f5 not run |
| external-write approval sentence | **PASS** — USER: `I approve posting the non-secret PO-06 UAT result comment on GitHub Issue #254.` |

### Environment re-measure (non-secret)

| Probe | Result |
|-------|--------|
| `curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:3003/` | http_code=`000` curl_exit=`7` → **DOWN/unreachable** |
| `curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/health` | http_code=`000` curl_exit=`7` → **DOWN/unreachable** |
| `E2E_LOGIN_*` presence (values not printed) | EMAIL=UNSET · PASSWORD=UNSET in this shell |
| mode | RECORD_ONLY — no assisted browser UAT; USER matrix accepted |

### Result matrix

| ID | Flow | Result |
|----|------|--------|
| f1 | 受付→診察→検査→会計→締め | PASS |
| f2 | 予約→来院→再予約 | PASS |
| f3 | トリミング受付→実施→精算 | FAIL |
| f4 | LINE予約→カルテ反映 | UNTESTED |
| f5 | 月次集計→帳票出力 | BLOCKED |
| overall | — | PARTIAL |

### Prior report

| Path | Disposition |
|------|-------------|
| `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md` | Exists; historical only; #254 complete?=NO — **not** auto-PASS |

### Comment execution (2026-08-08) — USER approval received

| Field | Value |
|-------|--------|
| Approval | USER: `I approve posting the non-secret PO-06 UAT result comment on GitHub Issue #254.` |
| `gh issue comment 254` | exit 0 → https://github.com/MinoruSoga/AnimalEkarte/issues/254#issuecomment-5226213497 |
| Re-measure #254 | still OPEN (not closed by this unit; overall≠PASS) |
| last_comment | matches filled body (overall=PARTIAL + f1–f5 + opaque_ref) |
| U4 status | **COMPLETE** |
| Next unit | **U5 = TASK-024 / #256** — not started |
| Local docs | PO-todo PO-06 checked; STATUS TASK-023 + §2 #254 one-liner; #254 row kept |

**Posted body (verbatim):**
```text
PO-06 / TASK-023 / #254: overall=PARTIAL f1=PASS f2=PASS f3=FAIL f4=UNTESTED f5=BLOCKED opaque_ref=2026-08-08-PO-uat-TASK-023 memo=stack local demo; f3 trimming fail TBD; f4/f5 not run
Policy: five-flow UAT record only. No secrets/PHI. #254 close not performed in this unit unless separately approved.
```

### Out of scope honored

- No inventing PASS; no #254 close; no U5 start; no secrets/PHI; no push; no assisted UAT while stack DOWN

---

## U5 — PO-14 · TASK-024 / #256 screenshot-FAQ sign-off

| Field | Value |
|-------|--------|
| Date | 2026-08-08 |
| Branch / HEAD | `main` (local residual docs dirty; ahead of origin) |
| Run status | **COMPLETE** (dual SIGNED_OFF record + #256 comment after USER approval; no issue close; no history rewrite) |
| Next unit | **U6 = TASK-022 / #239** (do not auto-start) |
| Prompt | `~/.claude/prompt-craft-runs/fast-residual-closeout-u5-task024-issue256.md` |

### Hard gate (launch message)

| Field | Value |
|-------|--------|
| privacy_disposition | **PASS** — `SIGNED_OFF` |
| repo_disposition | **PASS** — `SIGNED_OFF` |
| overall | **PASS** — dual SIGNED_OFF ⇒ `PASS` |
| opaque_ref | **PASS** — `2026-08-08-PO-signoff-TASK-024` (non-secret) |
| write_target | **PASS** — `BOTH` |
| memo | visual FAQ+manual screenshots reviewed off-repo; no PII in comment |
| external-write approval sentence | **PASS** — USER: `I approve posting the non-secret PO-14 TASK-024 sign-off comment on GitHub Issue #256.` |

### Pre-write evidence (re-measured)

| Check | Result |
|-------|--------|
| `gh issue view 256 --json number,title,state` | `{"number":256,"state":"OPEN","title":"[OPS] 操作マニュアル・手順書の整備 — 操作研修は納品後実施（PO裁定 2026-07-15）"}` |
| `test -d frontend/src/features/manual` | **EXISTS** (api, components, content, hooks, index.ts, …) — no images dumped to chat |
| F-256 | **RATIFY** — DEC-61 no-rewrite; TASK-024 必須残; DEFER rejected |
| DEC-61 | no history rewrite; non-secret enum + opaque ref only on ledger |
| Independent Review | Santa B+C **PASS** (NICE) on exact filled body |

### Disposition matrix

| Owner | Disposition |
|-------|-------------|
| Privacy | SIGNED_OFF |
| Repository | SIGNED_OFF |
| overall | PASS |

### Comment execution (2026-08-08) — USER approval received

| Field | Value |
|-------|--------|
| Approval | USER: `I approve posting the non-secret PO-14 TASK-024 sign-off comment on GitHub Issue #256.` |
| `gh issue comment 256` | exit 0 → https://github.com/MinoruSoga/AnimalEkarte/issues/256#issuecomment-5226653658 |
| Re-measure #256 | still OPEN (close approval not given; dual SIGNED_OFF alone does not close) |
| last_comment | matches filled body (privacy/repo SIGNED_OFF + overall=PASS + opaque_ref) |
| U5 status | **COMPLETE** |
| Next unit | **U6 = TASK-022 / #239** — not started |
| Local docs | PO-todo PO-14 checked; STATUS TASK-024 + §2 #256 one-liner; #256 row kept |

**Posted body (verbatim):**
```text
PO-14 / TASK-024 / #256: privacy=SIGNED_OFF repo=SIGNED_OFF overall=PASS opaque_ref=2026-08-08-PO-signoff-TASK-024 memo=visual FAQ+manual screenshots reviewed off-repo; no PII in comment
Policy: F-256 RATIFY + DEC-61 no-rewrite. Visual sign-off record only; no screenshots/PII in this comment. #256 close not performed unless dual SIGNED_OFF and separately approved.
```

### Out of scope honored

- No history rewrite; no screenshots/PII in repo or comment; no #256 close without separate approval; no U6 start; no push; no secrets

---

## U6 — PO-15 · TASK-022 / S13 + RLS residual evidence

| Field | Value |
|-------|--------|
| Date | 2026-08-09 |
| Branch / HEAD | `main` (local residual docs dirty; ahead of origin) |
| Run status | **COMPLETE** (LEDGER_PO record only; no ISSUE_239 comment; no reopen/close) |
| Next unit | **none** (U0–U6 frozen series complete; do not invent U7) |
| Prompt | `~/.claude/prompt-craft-runs/fast-residual-closeout-u6-task022-s13-rls.md` |

### Hard gate (launch message)

| Field | Value |
|-------|--------|
| s13_overall | **PASS** — `NOT_RUN` (not claiming scenario PASS; memo admits unrun steps 1–8) |
| rls_runtime | **PASS** — `N_A_APP_LAYER_ONLY` (S13: DB RLS runtime out of scenario) |
| env | **PASS** — `UNKNOWN` |
| opaque_ref | **PASS** — `2026-08-09-PO-TASK-022-residual` (non-secret) |
| write_target | **PASS** — `LEDGER_PO` |
| memo | S13 human 1-8 not executed this session; RLS runtime out of S13; app-layer identitylink tests only |
| external-write approval | **N/A** — LEDGER_PO only (no ISSUE_239) |

### Pre-write evidence (re-measured)

| Check | Result |
|-------|--------|
| `gh issue view 239 --json number,title,state` | `{"number":239,"state":"CLOSED","title":"[FEAT] 医院別レコードを残す同一owner/petリンクと所属院内の統合履歴（Q&A No.28①/No.30）"}` |
| S13 HUMAN table | remains agent-unfilled by design (all PENDING); residual record is ledger / PO-todo |
| S13 RLS runtime | 本シナリオ外; `N_A_APP_LAYER_ONLY` honest |
| `test -d backend/internal/identitylink` | **EXISTS** |
| `test -d frontend/src/features/identity-links` | **EXISTS** |
| Independent Review (#239 body) | **N/A** — write_target=LEDGER_PO; secret/PHI self-check on one-liner only |

### Result matrix

| Field | Value |
|-------|--------|
| s13_overall | NOT_RUN |
| rls_runtime | N_A_APP_LAYER_ONLY |
| env | UNKNOWN |

### Local docs

| Path | Action |
|------|--------|
| `todo-po.md` PO-15 | checkbox checked + 完了 one-liner |
| `STATUS.md` TASK-022 | optional residual one-liner; remains open (human S13 still PENDING in scenario) |
| ISSUE_239 | **skipped** (write_target=LEDGER_PO); #239 stays CLOSED |
| S13 HUMAN table | **unchanged** (agent does not fill) |

**Recorded body (verbatim):**
```text
PO-15 / TASK-022 / S13: s13=NOT_RUN rls_runtime=N_A_APP_LAYER_ONLY env=UNKNOWN opaque_ref=2026-08-09-PO-TASK-022-residual memo=S13 human 1-8 not executed this session; RLS runtime out of S13; app-layer identitylink tests only
Policy: human residual evidence only. No PHI. #239 may already be CLOSED; no reopen/close games in this unit. App-layer identitylink tests ≠ DB RLS runtime unless rls_runtime says so.
```

### Out of scope honored

- No inventing S13 PASS; no filling S13 HUMAN table; no #239 reopen/close/comment; no PHI/secrets; no production exercise; no push; no U7 invented

---

## PO-07 — TASK-021-B client registry inventory

| Field | Value |
|-------|--------|
| Date | 2026-08-09 |
| Branch / HEAD | `main` (local residual docs; claim=`claim/PO-07-TASK-021-CLIENT-REGISTRY`) |
| Run status | **COMPLETE** (BOTH_LOCAL: STATUS + PO-todo + ledger; no gh issue comment; no B/C/D DROP) |
| Next unit | **PO-08** access-log 90d counts (USER ops / STG-prod) — do not auto-chain product delete |
| Prompt | `~/.claude/prompt-craft-runs/fast-po07-task021-client-registry.md` |
| Orchestration | native Workflow `po07-task021-client-registry-scan` (probes A/B/C parallel explore) |

### Hard gate (launch message)

| Field | Value |
|-------|--------|
| declaration | **PASS** — `NO_KNOWN_EXTERNAL_CONSUMERS` |
| opaque_ref | **PASS** — `2026-08-09-PO-TASK-021-registry` |
| write_target | **PASS** — `BOTH_LOCAL` |
| inventory_start | **PASS** — `2026-08-09` (F-021-X / PO-09 clock) |
| memo | in-repo FE/LIFF only known consumers of exclusion surfaces; no partner/mobile clients known |
| external-write approval | **N/A** — BOTH_LOCAL only (no STATUS_AND_ISSUE) |

### In-repo client registry (paths only; no secrets)

| Surface | Path (representative) | Kind | Notes |
|---------|----------------------|------|-------|
| FE test compat | `frontend/src/hooks/use-reservation-types.test.ts:29,50,61,81,93,113` | test_compat | Asserts `excluded_courses` **not** on projected data |
| FE capable API | `frontend/src/features/master/api/staff-reservation-types.ts:19,39` | production_consumer (capable-side) | GET/PUT `capable-reservation-types` only — **not** exclusion route |
| FE UI (name legacy) | `frontend/src/features/master/components/StaffSidePanelSections.tsx:32` | other | `StaffExcludedReservationTypesSection` uses `capableIdSet` |
| FE generated type | `frontend/src/types/generated/models.ts` (StaffReservationExclusion) | type_def / unused_ref | No production import of exclusion routes |
| LIFF | `frontend/liff`, `frontend/src/shared-liff`, `frontend/line-reserve` | **absence** | **0 hits** for `excluded_courses` / `excluded-reservation-types` |
| OpenAPI field | `backend/docs/api.yaml:9119` | openapi_surface | `excluded_courses` deprecated:true |
| OpenAPI route | `backend/docs/api.yaml:14431` | openapi_surface | GET/PUT `/masters/staffs/{id}/excluded-reservation-types` |
| Handler routes | `backend/internal/staff/handler.go:209-210` | handler_route | GET/PUT still registered |
| Handler impl | `backend/internal/staff/staff_handler.go:314,329` | handler_route | Get/Set still live |
| Response DTO | `backend/internal/reservation/reservation_staff_response.go:20` | model_dto | `excluded_courses` JSON facade |
| Spec note | `docs/spec/screens/settings/master-staff.md:60-61` | doc | FE は capable 側で永続化のため exclusion 未呼出 |
| Policy | `docs/work/decisions/fable-po-recommendation.md:23,74` | policy | F-021-B HOLD + client registry mandatory |

**USER declaration:** `NO_KNOWN_EXTERNAL_CONSUMERS` — agent scan supports in-repo FE/LIFF-only known consumers; **not** absolute cryptographic proof of zero out-of-repo clients.

### Local docs

| Path | Action |
|------|--------|
| `STATUS.md` TASK-021 table + detail | one-liner + registry summary; B/C/D remain HOLD |
| `todo-po.md` PO-07 | checkbox checked + 完了 one-liner |
| `todo-po.md` PO-09 | checkbox checked + inventory_start=`2026-08-09` |
| Issue comment | **skipped** (write_target=BOTH_LOCAL) |

**Recorded body (verbatim):**
```text
PO-07 / TASK-021-B client_registry: declaration=NO_KNOWN_EXTERNAL_CONSUMERS opaque_ref=2026-08-09-PO-TASK-021-registry inventory_start=2026-08-09 memo=in-repo FE/LIFF only known consumers of exclusion surfaces; no partner/mobile clients known
In-repo scan: FE/LIFF/OpenAPI/handler cites summarized in ledger/STATUS (paths only). B/C/D DROP not performed. F-021-X clock starts only if inventory_start set.
```

### Out of scope honored

- No TASK-021-B response field delete; no C route DROP; no D migrate DROP; no PO-08 STG/prod log access; no push/PR; no secrets/IP/UA/token; no gh issue comment

---

## UAT-R1 — TASK-023 / #254 residual f3–f5 re-record

| Field | Value |
|-------|--------|
| Date | 2026-08-09 |
| Branch / HEAD | `main` (local residual docs dirty; `?? reports/`) |
| Run status | **COMPLETE** (RECORD_ONLY STATUS_LEDGER; no #254 comment; no issue close) |
| Next unit | **S13 human steps 1–8** (recommended order #2) — do not auto-start; PO-08 still deferred |
| Prompt | `~/.claude/prompt-craft-runs/fast-uat-residual-f3-f5.md` |
| Orchestration | native Workflow `uat-r1-residual-record` + parallel explore probes B/C; Probe A env curls on main |

### Hard gate (launch message)

| Field | Value |
|-------|--------|
| mode | **PASS** — `RECORD_ONLY` |
| overall | **PASS** — `PARTIAL` (consistent with f3=FAIL + f4=UNTESTED + f5=BLOCKED) |
| f1–f5 | **PASS** — f1=PASS(INHERIT_U4) f2=PASS(INHERIT_U4) f3=FAIL f4=UNTESTED f5=BLOCKED |
| opaque_ref | **PASS** — `2026-08-09-PO-uat-R1` (non-secret) |
| write_target | **PASS** — `STATUS_LEDGER` |
| memo | no re-run yet; stack may be down; residual f3-f5 unchanged from U4 |
| external-write approval | **N/A** — STATUS_LEDGER only (no ISSUE_254 / BOTH) |

### Environment re-measure (non-secret)

| Probe | Result |
|-------|--------|
| `curl -s -o /dev/null -w '%{http_code}' --connect-timeout 3 http://127.0.0.1:3003/` | http_code=`000` → **DOWN/unreachable** |
| `curl -s -o /dev/null -w '%{http_code}' --connect-timeout 3 http://127.0.0.1:8080/health` | http_code=`000` → **DOWN/unreachable** |
| mode | RECORD_ONLY — no assisted browser UAT; USER matrix accepted without inventing PASS |

### Result matrix

| ID | Flow | Result | Notes |
|----|------|--------|-------|
| f1 | 受付→診察→検査→会計→締め | PASS | INHERIT_U4 |
| f2 | 予約→来院→再予約 | PASS | INHERIT_U4 |
| f3 | トリミング受付→実施→精算 | FAIL | residual unchanged; no re-run |
| f4 | LINE予約→カルテ反映 | UNTESTED | residual unchanged; no re-run |
| f5 | 月次集計→帳票出力 | BLOCKED | residual unchanged; no re-run |
| overall | — | PARTIAL | not all PASS/waived SKIP |

### Local docs

| Path | Action |
|------|--------|
| `STATUS.md` TASK-023 table | UAT-R1 residual one-liner; remains open |
| `STATUS.md` §1 TASK-022/023/024 | UAT-R1 bullet |
| `STATUS.md` §2 #254 | UAT-R1 one-liner; #254 OPEN |
| `todo-po.md` PO-06 | residual UAT-R1 line under checked PO-06 |
| Issue #254 comment | **skipped** (write_target=STATUS_LEDGER; no approval sentence) |

**Recorded body (verbatim):**
```text
UAT-R1 / TASK-023 / #254 residual f3–f5: overall=PARTIAL f1=PASS f2=PASS f3=FAIL f4=UNTESTED f5=BLOCKED opaque_ref=2026-08-09-PO-uat-R1 memo=no re-run yet; stack may be down; residual f3-f5 unchanged from U4
Policy: local demo residual fill after U4 PARTIAL. No secrets/PHI. No prod. PO-08 STG/prod logs out of scope. #254 close not performed unless separately approved.
```

### Out of scope honored

- No inventing f3–f5 PASS; no assisted UAT while stack DOWN; no #254 comment/close; no PO-08 STG/prod log counts; no TASK-021 B/C/D DROP; no S13 execution; no U7 product unit; no push/PR; no secrets/PHI
