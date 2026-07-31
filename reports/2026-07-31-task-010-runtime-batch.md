# TASK-010 — Runtime browser batch (LINE / V05 hospital-side)

> **Date**: 2026-07-31  
> **Packet**: W-010 / unit TODO-MD-OPEN-REMAINING-ORCH-WAVE-20260731-V2  
> **Claim**: `claim/TASK-010` (not released by agent)  
> **Status**: **PARTIAL PASS** — env READY; Chrome batch executed; 5 marks elevated; backlog remains  
> **Scope**: Browser measurement + report. Product code / inventory / foreign claims not touched.

---

## 1. Env status (NOW READY vs prior BLOCKED)

| Check | Prior (`2026-07-31-task-010-020-runtime-blocked.md`) | This session |
|:---|:---|:---|
| docker compose healthy | Not verified | **READY** — `animalekarte-backend-1` / `db-1` / `frontend-1` Up ~4h (healthy) |
| `GET :8080/health` | Not verified | **200** `{"status":"ok"}` |
| Frontend `:3003/` | Not verified | **200** |
| Chrome DevTools MCP | Not used | **Connected** — login + multi-page navigation OK |
| Login / seed | Not verified | **OK** — demo `admin@noavet.jp` (システム管理者) → 八王子病院 当日受付 |

Prior disposition: both TASK-010 / TASK-020 **BLOCKED** on env. **Env blocker is cleared.**

---

## 2. 【要実測】 census

| Scope | Before | After | Δ |
|:---|---:|---:|---:|
| All under `docs/ops/testing/scenarios` (incl. reports/README prose) | **77** | **72** | −5 |
| Scenario bodies only (`V*.md` + `S*.md`) | **70** | **65** | −5 |
| `V05-auth-line-forms.md` alone | **13** | **8** | −5 |

Count command (re-run anytime):

```bash
rg -c '【要実測】' docs/ops/testing/scenarios
rg -c '【要実測】' docs/ops/testing/scenarios/V*.md docs/ops/testing/scenarios/S*.md
```

---

## 3. Batch attempted (Chrome DevTools MCP)

**Batch theme**: LINE hospital-side + Lステップ settings (V05-8 / V05-9 / V05-10 / V05-12 / V05-13 / V05-15) — preferred over full ~70.

**Account**: システム管理者 `admin@noavet.jp` / password demo (`password`)  
**Clinic context**: 八王子病院  
**URLs exercised**:
- `http://localhost:3003/login` → `/`
- `/line-reservation/settings`
- `/line-reservation/page-editor`
- `/line-reservation/slots?typeId=1`
- `/settings/integrations/lstep`
- `/line-reserve/` (no clinicId → error)
- `/line-reserve/1` (consumer SPA home)

### Results table

| ID | Step | Result | Observation |
|:---|:---|:---|:---|
| V05-8 #2 | 0/out-of-range booking window / months / interval | **PASS** (elevated) | FE native min: 最長≥1, 月数 1–6, 間隔≥5. Save blocked with alert「値は 1 以上にする必要があります。」API not reached. 最短受付 allows 0 (`min=0`) |
| V05-8 #4 | secret / access token fields | **PASS** (elevated) | Form inputs: channel ID + LIFF ID only. No secret/token labels or fields in DOM |
| V05-8 #1 | 受付停止 → 飼い主側 | **SKIPPED** | Would mutate clinic LINE accept flag for shared env; not toggled |
| V05-9 #4 | 1万字 long text | **PARTIAL / not elevated** | All 5 textareas `maxLength=-1` (no FE cap). Did **not** POST 10k to avoid seed pollution / BE unknown |
| V05-10 #5 | duplicate day×start | **PASS** (elevated) | Add 09:00 → POST 201; re-select 09:00 → 「この時刻は既に登録済みです」+ Add disabled. Slot then DELETE 204 (cleanup) |
| V05-12 #3 | threshold 0 / negative | **PASS** (elevated) | CPM threshold spinbuttons `min=1`; 0 → invalid + same alert; save blocked before API |
| V05-13 | same priority values | **PASS** (elevated) | UI copy: 「同値は同一優先階層として扱われます。」 Seed already holds pairs at 4, 8, 13 |
| V05-15 | duplicate auto-managed prefix | **PARTIAL** | Existing prefix `canceled_visit` re-add → **POST 409**. Toast text slightly odd (`lstep_auto_managed_prefix '' already exists`). Disease-code duplicate still 【要実測】 |
| line-reserve consumer | smoke | **OK smoke** | `/line-reserve/` → 「クリニックIDが見つかりません」; `/line-reserve/1` → ヘッダー「ノア動物病院 八王子」+ 新規予約 / 予約確認 |

### Side effects cleaned

- Temporary slot `一般診察` / 2026-07-31 / 09:00: created then deleted (DELETE 204).
- No permanent settings save of invalid zeros (FE blocked).
- No 受付停止 toggle; no 10k page-editor save.

### Scenario marks changed

| File | Change |
|:---|:---|
| `docs/ops/testing/scenarios/V05-auth-line-forms.md` | Elevated expected results; removed 【要実測】 on V05-8 #2, V05-8 #4, V05-10 #5, V05-12 #3, V05-13; narrowed V05-15 to disease-code only |

**No other scenario files edited.**

---

## 4. Remaining backlog (65 body marks)

| Pack | Remaining 要実測 |
|:---|---:|
| V01 clinical forms | 17 |
| V04 settings master | 11 |
| V02 accounting / reservation | 9 |
| V05 auth / LINE (residual) | 8 |
| V03 owner / pet / staff | 4 |
| S01, S02, S11 | 3 each |
| S03, S12 | 2 each |
| S04, S05, S07 | 1 each |

**V05 residual (need next batch)**:
1. V05-6 #4 phone non-digits (line-reserve create — needs LIFF/customer journey)
2. V05-7 #4 same-day cancel window
3. V05-8 #1 accept-off consumer UX (requires deliberate settings mutation)
4. V05-9 #4 10k char save (FE uncapped; need controlled BE probe + restore)
5. V05-14 empty/invalid mapping entries
6. V05-15 disease-code duplicate
7. V05-16 bad CSV
8. V05-17 bulk tag remove observation (Write API stopped)

---

## 5. Exact next USER commands (full backlog)

```bash
# A) Stack (already healthy as of this report — re-check if cold)
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
docker compose ps
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8080/health
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:3003/

# B) Mark census
rg -n '【要実測】' docs/ops/testing/scenarios --glob '!**/reports/**' --glob '!README.md'

# C) Continue browser batches (Chrome DevTools MCP /browser-test skill)
# Prefer small batches by pack, e.g.:
#   - V05 residual (line-reserve create/cancel + accept-off)
#   - V01 clinical (needs seed pets/records)
#   - V02 accounting/reservation
#   - V04 masters
# Record each batch under reports/YYYY-MM-DD-*.md
# Elevate 【要実測】 only on clear PASS; cleanup any temporary data

# D) Optional TASK-020 (separate): UI design compliance Playwright
# docker compose exec frontend pnpm test:e2e -- e2e/ui-design-compliance-readonly.spec.ts --workers=1
```

Do **not** auto-run full `pnpm test:run`, full e2e suite, `docker compose down`, or DB reset from agents.

---

## 6. Disposition

| Item | Status |
|:---|:---|
| TASK-010 env blocker | **CLEARED** |
| TASK-010 mark → 0 | **OPEN** (65 body remaining) |
| This batch | **5 elevated**, evidence in this report + V05 scenario body |
| Claim `claim/TASK-010` | Held — **USER** releases after integrate/abandon |

### Related reports

| Report | Relation |
|:---|:---|
| `reports/2026-07-31-task-010-020-runtime-blocked.md` | Prior BLOCKED (env) |
| `reports/2026-07-31-task-009-seed-design.md` | Seed for clinical browser cases |
| `reports/2026-07-31-task-019-line-audit.md` | LINE audit companion |
