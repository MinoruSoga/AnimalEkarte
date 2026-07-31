# TASK-024 — Manual screenshot audit + FAQ gate

| Field | Value |
|-------|--------|
| Date | 2026-07-31 |
| Branch | `task-024-impl-20260731` |
| Claim | `claim/TASK-024` (held; not released) |
| Phase A | 10 screen PNGs + judgment table (below) |
| Phase B | FAQ gate from TASK-023 confusions — **追記不要** (see §Phase B) |
| Live app | frontend `http://127.0.0.1:3003`, backend `http://127.0.0.1:8080` (shared compose) |
| Capture viewport | 1440×900 CSS px (within 1280–1920 width) |
| Auth | Shared session already authenticated (Playwright). No credentials invented or read from reports. |

## Method

1. Read corresponding screen docs under `frontend/src/features/manual/content/screens/`.
2. Verify each PNG exists, is valid, and is still referenced as `images/<same-filename>.png` in the matching screen MD.
3. Capture live UI via Playwright against the shared app (authenticated shell reachable).
4. Visually compare existing PNG vs live UI chrome (layout, tabs, columns, toolbar controls) — not seed data dates alone.
5. **replace** only when chrome mismatch is structural or the existing image is clearly broken; otherwise **current** with explicit human sign-off slot.
6. PHI: live shared DB rows include owner/pet names (demo/seed-like). Replacements overwrite same filename only; no orphan `_screenshot-*` added under `content/images/`.

## Limitations (agent-limited)

| Item | Detail |
|------|--------|
| Aggregation page | Live `/aggregation` remained on 「読み込み中...」 after multi-second wait (heavy clinic data). Structural chrome (CPM chips) observed via a11y text, but clean full-page capture of loaded table **not** obtained. |
| Multi-clinic filter | Live list screens show 「表示拠点」 chips for multi-clinic users. Existing clean seed screenshots lack this chrome; replacing would ship dense real-looking rows (year anomalies / volume). Left **current** where list chrome still matches docs without being broken. |
| Sidebar chrome drift | Live nav includes 締め履歴 / LINE / Lステップ / マスタ / 取扱説明書 more consistently than older PNGs. Not alone a replace trigger when main content still matches. |
| E2E env | `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD` **unset** in agent environment → manual-flow e2e portion **BLOCKED** (env check only; no secret values). |

## 10/10 judgment table

| # | Image | Screen MD | Route(s) | Judgment | Rationale |
|---|-------|-----------|----------|----------|-----------|
| 1 | `02-reception.png` | `02-reception.md` | `/` | **replace** | Existing lacked header **受付テレメトリ** (本日受付 / 平均待ち / 最長待ち) documented in MD §画面の見方. Live matches 5-column kanban + telemetry. Overwrote same filename (1440×900). |
| 2 | `04-medical-records.png` | `04-medical-records.md` | `/medical-records` | **current** | Existing list chrome (columns 診療日…操作, 作成中/確定済, 会計 link, 新規カルテ登録) still matches doc and is intact. Live adds multi-clinic 「表示拠点」 and dense rows (incl. anomalous dates) — not suitable blind replace. **PENDING visual sign-off by named documentation owner** for 表示拠点 chrome. |
| 3 | `05-accounting.png` | `05-accounting.md` | `/accounting` | **replace** | Existing showed 2 tabs only (会計一覧 / 未納者一覧). Live + MD document **当日会計** as third tab. Overwrote same filename. |
| 4 | `06-reservations.png` | `06-reservations.md` | `/reservations` | **replace** | Live week view exposes explicit **5日 / 7日** toggle next to 週表示 (doc: 表示日数は 5 日／7 日切替). Existing lacked that control. Overwrote same filename. |
| 5 | `07-examinations.png` | `07-examinations.md` | `/examinations` | **replace** | Existing rows had **empty 飼主名 / ペット名 / 検査種別** while showing results — incomplete / misleading for manual. Live fills those columns. Overwrote same filename. |
| 6 | `10-trimming.png` | `10-trimming.md` | `/trimming` | **replace** | Doc lists **犬種** column; existing PNG omitted 犬種. Live list includes 犬種 (+ pet id under name). Overwrote same filename. Note: live sample sparse on スタイル希望 values. |
| 7 | `13-cash-register.png` | `13-cash-register.md` | `/accounting/close` | **replace** | Doc requires 区分 **午前 / 午後 / 緊急**. Existing showed only 午前/午後. Live includes **緊急**. Overwrote same filename. |
| 8 | `14-accounting-reports.png` | `14-accounting-reports.md` | `/accounting/reports` | **replace** | Existing was **broken for docs**: red 「データの取得に失敗しました」 with no table. Live shows 集計単位 + 日次明細 (AM締/PM締) + 印刷/PDF・CSV. Overwrote same filename. |
| 9 | `16-line-reservation.png` | `16-line-reservation.md` | `/line-reservation/settings` | **current** | Live basic settings (稼働状態, 営業時間・定休日, 予約枠設定) matches existing PNG chrome. No structural mismatch warranting replace. **PENDING visual sign-off by named documentation owner** for minor nav chrome only. |
| 10 | `19-aggregation.png` | `19-aggregation.md` | `/aggregation` | **current** | Existing intact (tabs 売上ランキング/来院回数/最終来院, filters, ranking table). Live shell shows **CPMセグメント** chips (doc §CPM) but table stayed loading — agent could not replace with a clean loaded capture. **PENDING visual sign-off by named documentation owner** to recapture with CPM chips + loaded rows. |

### Summary counts

| Judgment | Count | Files |
|----------|------:|-------|
| replace | 7 | `02-reception`, `05-accounting`, `06-reservations`, `07-examinations`, `10-trimming`, `13-cash-register`, `14-accounting-reports` |
| current | 3 | `04-medical-records`, `16-line-reservation`, `19-aggregation` |

## Named documentation owner — visual sign-off

| Slot | Status | Owner |
|------|--------|-------|
| Full 10-image visual sign-off against product UI | **PENDING** | **named documentation owner** (human — not agent) |
| Priority re-check | **PENDING** | `04-medical-records` (表示拠点), `19-aggregation` (CPM chips + loaded data), any PHI redaction preference on replaced list screens |

> Agents must not mark this slot complete. Human documentation owner signs after reviewing the 7 replacements and 3 kept images in-product.

## Integrity checks (all 10)

| Check | Result |
|-------|--------|
| PNG valid / non-zero | Pass (all 10) |
| Screen MD still references same path `images/<file>.png` | Pass (all 10) |
| Manual article routing still expects these screen slugs | Pass (articles present; e2e references `05-accounting` path pattern) |
| Orphan `_screenshot-*` created in owned images dir | **None** (phase A) |
| FAQ / `10-troubleshooting.md` edited | **No** — Phase B 追記不要 (see §Phase B) |

## Image dimensions after phase A

| File | W×H | Action |
|------|-----|--------|
| `02-reception.png` | 1440×900 | replaced |
| `04-medical-records.png` | 2880×1800 | current |
| `05-accounting.png` | 1440×900 | replaced |
| `06-reservations.png` | 1440×900 | replaced |
| `07-examinations.png` | 1440×900 | replaced |
| `10-trimming.png` | 1440×900 | replaced |
| `13-cash-register.png` | 1440×900 | replaced |
| `14-accounting-reports.png` | 1440×900 | replaced |
| `16-line-reservation.png` | 2880×1800 | current |
| `19-aggregation.png` | 1440×900 | current |

## Verification

### Vitest (scoped manual suite)

Command (shared compose frontend service):

```bash
docker compose -f /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/docker-compose.yml exec -T frontend \
  npx vitest run src/features/manual/api/get-manual-articles.test.tsx \
  src/features/manual/components/ManualSidebar.test.tsx \
  src/features/manual/components/manual-content.test.ts \
  src/features/manual/lib/parse-frontmatter.test.ts \
  src/features/manual/routes/ManualPage.test.tsx
```

Result: **PASS** — 5 files / 18 tests passed (vitest v4.1.8, ~1.3s). Shared frontend mount used; tests do not depend on binary PNG content.

### E2E manual-flow

```bash
cd frontend && ./scripts/run-e2e.sh e2e/manual-flow.spec.ts
```

| Check | Result |
|-------|--------|
| `E2E_LOGIN_EMAIL` | unset |
| `E2E_LOGIN_PASSWORD` | unset |
| E2E execution | **BLOCKED** (login env missing; env check only — no secrets printed; e2e not started) |

### Diff check

```bash
git diff --check -- reports/2026-07-31-task-024-manual-audit.md \
  frontend/src/features/manual/content/images/02-reception.png \
  frontend/src/features/manual/content/images/05-accounting.png \
  frontend/src/features/manual/content/images/06-reservations.png \
  frontend/src/features/manual/content/images/07-examinations.png \
  frontend/src/features/manual/content/images/10-trimming.png \
  frontend/src/features/manual/content/images/13-cash-register.png \
  frontend/src/features/manual/content/images/14-accounting-reports.png
```

Result: **PASS** (`git diff --check` exit 0).

## Phase B — FAQ (`10-troubleshooting.md`) gate

### Input (authoritative)

| Field | Value |
|-------|--------|
| Source report | `docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md` |
| Ledger task (source) | TASK-023 / #254 |
| Section | §5 TASK-024 向け — confusion / staff-pain manifest |
| `confusion_count` | **0** |
| `confusions` | `[]` (empty) |
| Agent note in source | Authenticated UI not exercised (`E2E_LOGIN_*` missing); do not invent FAQ from residual 【要実測】 lists |

### Decision: **追記不要**

| Mapping | Result |
|---------|--------|
| TASK-023 observed staff confusions | **0** |
| Speculative FAQ entries from residual / backlog | **禁止**（source report notes + phase B required behavior） |
| Action on `frontend/src/features/manual/content/workflows/10-troubleshooting.md` | **変更なし**（追記しない） |
| 1:1 rationale | zero confusions from TASK-023 → no FAQ append |

> **追記不要**: TASK-023 レポート（`docs/ops/testing/scenarios/reports/2026-07-31-local-issue-254.md`）の `confusion_count: 0` / `confusions: []` を正本とし、観測ゼロのため FAQ にエントリを追加しない。E2E 未実行・human UAT 未了による未観測を、推測で埋めてはならない。

### Integrity after Phase B

| Check | Result |
|-------|--------|
| `10-troubleshooting.md` content modified | **No** |
| Speculative FAQ entries added | **None** |
| Human visual sign-off (Phase A images) | **PENDING** (unchanged; agent must not complete) |

## Remains for human

1. **Human sign-off** of all 10 images (table above) — still **PENDING**.
2. Optional recapture: `19-aggregation` with loaded CPM chips; `04-medical-records` if multi-clinic chrome must appear in manual.
3. Optional: re-shoot replaced PNGs at 2× device scale for parity with older 2880×1800 assets.
4. Run e2e / human UAT once login env is available; if new confusions are observed later, open a follow-up FAQ update from that evidence (not invent from residual lists).
5. USER releases `claim/TASK-024` only after main integration or explicit abandon (`git branch -D claim/TASK-024`).
