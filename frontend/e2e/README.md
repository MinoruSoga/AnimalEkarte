# Playwright E2E Tests

## Overview

End-to-end tests for Animal Ekarte using Playwright.

### Test Coverage

| File | Tests |
|------|-------|
| `owners-search.spec.ts` | /owners kana search — unauthenticated redirect + ぴ/ピ→ピーター (かな非区別) |
| `owners-flow.spec.ts` | /owners list/create/search/detail/edit 主要操作フロー |
| `accounting-smoke.spec.ts` | /accounting タブ smoke + いりす/イリス→Iris かな非区別検索 + /accounting?tab=unpaid + /accounting/reports |
| `accounting-flow.spec.ts` | /accounting 行クリック→詳細遷移・Iris 検索→詳細・/accounting/reports セレクタ・会計精算フォーム確定ボタン表示 |
| `reservations-smoke.spec.ts` | /reservations auth guard + カレンダーナビ smoke |
| `reservation-patient-search.spec.ts` | 新規予約作成モーダル PatientSelectionTable: 先頭20件外の患者1003298（SPANKY）を `#search` の自動デバウンスでpet-name検索し選択 |
| `medical-records-patient-search.spec.ts` | usePetSelectionPage代表面: `include_deceased=true` の先頭20件外から患者1003298（SPANKY）をpet-name検索し、accessible name付き選択ボタンを確認 |
| `clinical-smoke.spec.ts` | 受付/顧客集計/カルテ管理/入院管理/トリミング/検査/予防接種/定期健診 各ページ smoke |
| `clinical-flows.spec.ts` | カルテ管理 一覧/検索/行クリック詳細・ペット選択画面 |
| `medical-records-create.spec.ts` | /medical-records/new?petId=1 直接 URL アクセスで新規カルテ入力フォーム表示確認 |
| `medical-records-pagination-sort.spec.ts` | /medical-records ページネーション(page=2 遷移)・列ソート(飼主名 desc→asc, URL 状態)・ステータスフィルタ(確定済, PropertyFilter) — `c80d9dc1` follow-up (AC-3) |
| `trimming-flow.spec.ts` | /trimming 一覧表示・新規登録ペット選択・/trimming/new?petId=1 フォーム表示 |
| `hospitalization-flow.spec.ts` | /hospitalization 一覧リストビュー・新規登録遷移・ステータスタブ切り替え |
| `vaccinations-flow.spec.ts` | /vaccinations 一覧表示・検索フィルタ・新規登録遷移・行クリック詳細遷移 |
| `business-smoke.spec.ts` | 業務系ページ（ダッシュボード/CPM/在庫/日次集計等）smoke |
| `operations-smoke.spec.ts` | 受付/トリミング/会計 等 主要ページ一覧/操作 smoke |
| `inventory-crud.spec.ts` | /inventory 在庫 CRUD フロー |
| `settings-smoke.spec.ts` | /settings/* 全設定ページ smoke (21 ページ) |
| `settings-crud.spec.ts` | /settings/* 設定マスタ CRUD フロー |
| `master-crud.spec.ts` | 処置マスタ CRUD (A-D) 4 ケース |
| `auth-flows.spec.ts` | /login 表示・入力・ログイン成功/失敗・パスワード表示切替 + /forgot-password・/reset-password アクセス/リダイレクト確認 |
| `examinations-flow.spec.ts` | /examinations 一覧・/examinations/select-pet ペット選択・/examinations/new フォーム・/:id 詳細・検索機能 |
| `checkups-flow.spec.ts` | /checkups 一覧・/checkups/select-pet ペット選択・/checkups/new フォーム・新規ボタン遷移・検索機能 |
| `estimates-flow.spec.ts` | /estimates 一覧・/estimates/new 新規フォーム・/:id 詳細・/:id/edit 編集フォーム・新規ボタン・検索 |
| `shifts-flow.spec.ts` | /shifts カレンダー表示・ナビゲーション矢印・スタッフセレクタ・カレンダー移動・フィルタ機能 |
| `lstep-flow.spec.ts` | /lstep/checkup-sync 抽出ページ + /lstep/delivery-monitor 監視ページ・フィルタ + /lstep/analytics 分析ページ・セレクタ |
| `line-reservation-flow.spec.ts` | /line-reservation 基本設定フォーム + /line-reservation/page-editor 編集フォーム + /line-reservation/slots 枠設定ページ/説明表示 |
| `manual-flow.spec.ts` | /manual リダイレクト・サイドバー・カテゴリ切替・検索 + /manual/:category/:slug 記事ページ・リンク遷移 |

## Execution Model

Playwright runs in the **official `mcr.microsoft.com/playwright` Docker image**, not the project's
Alpine-based frontend container. The Alpine container cannot run Playwright's Chromium (glibc
dependency). The test container connects to the running frontend through the host-published
`http://localhost:3003` port as `http://host.docker.internal:3003`.

```
┌─────────────────────────────────────┐
│  Host / docker compose              │
│  ┌──────────┐   ┌──────────────┐   │
│  │ frontend │   │   backend    │   │
│  │ :3003    │   │   :8080      │   │
│  └──────────┘   └──────────────┘   │
│         ▲ host.docker.internal      │
│  ┌──────┴───────────────────────┐  │
│  │ playwright:v1.60.0-jammy     │  │
│  │ (Chromium 1223, glibc-based) │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

## Prerequisites

### App

The app must be reachable from the host at `http://localhost:3003`.

```bash
docker compose up -d   # if not already running
```

### Seed data assumed by E2E tests

| Spec | Required data | Source |
|------|--------------|--------|
| `owners-search.spec.ts` | pet name `ピーター` (name_kana=`ぴーたー`), owner 5 (佐藤 花子), clinic 1 | `003_seed_demo.sql` |
| `accounting-smoke.spec.ts` | owner 1 (林 文明, はやし ふみあき) with completed billing for pet 1 (`Iris(イリス)`, name_kana=`いりす`) | `003_seed_demo.sql` |
| `accounting-flow.spec.ts` | same as `accounting-smoke.spec.ts` | `003_seed_demo.sql` |
| `reservations-smoke.spec.ts` | admin user at clinic 1 with reservations permission | `003_seed_demo.sql` |
| `reservation-patient-search.spec.ts` | clinic 1 の pet id=1003298 `SPANKY` が `/v1/pets?page=1&limit=20` の外に存在し、`search=SPANKY` で返る | `003_demo/pets.csv` |
| `medical-records-patient-search.spec.ts` | 同じ pet id=1003298 `SPANKY`; 初期一覧と検索の双方が `include_deceased=true` | `003_demo/pets.csv` |
| `master-crud.spec.ts` | treatment procedure items incl. `注射` (root with children) | `003_seed_demo.sql` |
| `hospitalization-flow.spec.ts` | 1+ active hospitalization records at clinic 1 | `003_seed_demo.sql` |
| `vaccinations-flow.spec.ts` | 1+ vaccination records; owner `林 文明` with pet `林 文明` | `003_seed_demo.sql` |
| `medical-records-pagination-sort.spec.ts` | clinic 1 に PAGE_SIZE(20) 超（開発環境では20,000件超）の medical_records が必要（page=2 到達用）。件数が少ない環境ではページ2ボタンが表示されず fail する | 開発DBの既存データ量に依存（seed 追加不要な環境が大半） |

If seed data is missing, run:

```bash
make reset   # resets and re-applies all migrations + seeds
```

### Auth

Credentials are **env-injected only** (SEC-CS2-F01). There is no in-repository password fallback.

| Variable | Required | Description |
|----------|----------|-------------|
| `E2E_LOGIN_EMAIL` | yes (for authenticated specs) | Admin account email present in the target DB (local demo seed) |
| `E2E_LOGIN_PASSWORD` | yes (for authenticated specs) | Matching password (never commit; inject via shell/CI secrets) |
| `E2E_AUTH_STATE_PATH` | no | Cached storage-state path (default `/tmp/animal-ekarte-demo-admin-storage-state.json`) |

Login is handled automatically via `helpers/auth.ts`; no manual pre-auth step is needed once the env vars are set.

## Running Tests

### Recommended: Docker script (avoids browser version mismatch)

```bash
# All tests
./scripts/run-e2e.sh

# Specific file
./scripts/run-e2e.sh e2e/owners-search.spec.ts
```

This script mounts only spec/config files and installs a fresh `@playwright/test@1.60.0`
inside the playwright container. This avoids the chromium-1217 vs chromium-1223 version
conflict caused by the host node_modules (pnpm-lock pins 1217, Docker image ships 1223).
Override the target with `PLAYWRIGHT_TEST_BASE_URL` when needed.
When set on the host, `E2E_LOGIN_EMAIL`, `E2E_LOGIN_PASSWORD`, and `E2E_AUTH_STATE_PATH`
are forwarded into the Playwright container (name-only `-e`; unset vars are not injected).

### Alternative: macOS native (if pnpm and playwright browsers are installed on host)

```bash
cd frontend
PLAYWRIGHT_TEST_BASE_URL=http://localhost:3003 pnpm test:e2e
```

**Note**: If you see `Executable doesn't exist at .../chromium_headless_shell-1217/...`,
your local playwright-core (pnpm-lock) expects chromium-1217 but only 1223 is installed.
Use the Docker script above instead.

### UI mode (interactive)

```bash
cd frontend
pnpm test:e2e:ui
```

## Authentication

All specs (except those that test unauthenticated redirect) log in automatically via
`helpers/auth.ts:loginAsDemoAdmin`. The helper reads `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD`
(required; throws if unset), navigates to `/login`, waits for the login API to succeed, then
stores the authenticated storage state in `/tmp` (or `E2E_AUTH_STATE_PATH`).
Later specs restore that state instead of repeating UI login, avoiding the backend login
rate limit during full-suite runs.

```bash
export E2E_LOGIN_EMAIL='…'      # local demo admin email
export E2E_LOGIN_PASSWORD='…'   # never commit
./scripts/run-e2e.sh
```

`master-crud.spec.ts` shares a `BrowserContext` across all tests via `beforeAll` (same pattern
used in `owners-search.spec.ts`, `accounting-smoke.spec.ts`, and `reservations-smoke.spec.ts`).

## Coverage Notes

### Reservation cancelled / no_show filter

`filterCalendarAppointments` (filters out `cancelled`, keeps `no_show`) is unit-tested in:

```
frontend/src/features/reservations/routes/__tests__/reservation-management.filter.test.ts
```

E2E validation of this filter would require seeding appointments for the current calendar week —
a date-dependent setup that is intentionally excluded to avoid flakiness.
`reservations-smoke.spec.ts` provides a page-load smoke test; the filter guarantee is the unit test.

### Patient server-side search

The reservation modal and the shared `usePetSelectionPage` selection flow both
prove that exact runtime pet `1003298` (`SPANKY`) is absent from the unfiltered
first 20 rows, returned by a debounced pet-name backend `search`
predicate, rendered, and exposed through an exact accessible select-button name. Kana normalization
for other list surfaces remains covered by their dedicated specs and the
`normalizeKana` unit tests.

## Architecture Support

| Environment | Support |
|-------------|---------|
| Linux x86_64 (Docker script) | ✅ Fully supported |
| Linux arm64/aarch64 (Apple Silicon Docker) | ✅ Supported via the official Playwright image |
| macOS arm64 (native pnpm) | ✅ Works if playwright browsers installed |

## Troubleshooting

### `Executable doesn't exist at .../chromium_headless_shell-1217/...`

Your host's `playwright-core` (from pnpm-lock) pins Chromium 1217, but the available
Playwright image uses Chromium 1223. Use `./scripts/run-e2e.sh` which runs in the correct
environment with no host node_modules interference.

### Apple Silicon / arm64

Use `./scripts/run-e2e.sh`. The official Playwright Docker image can launch Chromium on
Linux arm64, so the tests should execute rather than skip.

### `connect EPERM 127.0.0.1:9222`

This was a previous Chrome DevTools (CDP) connection error. The new test setup does NOT
use CDP — Playwright launches its own Chromium. This error should not appear anymore.

### Tests fail with "Failed to navigate"

```bash
curl http://localhost:3003  # verify frontend is up
curl http://localhost:8080/health  # verify backend is up
```

### Tests fail with "Element not found"

Update selectors in the spec file. Run with `--reporter=html` and open
`frontend/playwright-report/index.html` for a detailed breakdown.
