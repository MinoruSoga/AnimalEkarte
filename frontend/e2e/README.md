# Playwright E2E Tests

## Overview

End-to-end tests for Animal Ekarte using Playwright.

### Test Coverage

| File | Tests |
|------|-------|
| `master-crud.spec.ts` | Settings page master CRUD (A-D) |
| `owners-search.spec.ts` | /owners kana search — unauthenticated redirect + ぴ/ピ→ピーター (かな非区別) |
| `accounting-smoke.spec.ts` | /accounting タブ smoke + いりす/イリス→Iris かな非区別検索 + /accounting?tab=unpaid + /accounting/reports |
| `reservations-smoke.spec.ts` | /reservations auth guard + カレンダーナビ smoke (cancelled filtering は unit test 担保) |

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
| `reservations-smoke.spec.ts` | admin user at clinic 1 with reservations permission | `003_seed_demo.sql` |
| `master-crud.spec.ts` | treatment procedure items incl. `注射` (root with children) | `003_seed_demo.sql` |

If seed data is missing, run:

```bash
make reset   # resets and re-applies all migrations + seeds
```

### Auth

All specs use `admin@noavet.jp` / `password` (demo seed, clinic 1, `is_system_admin=true`).
Login is handled automatically via `helpers/auth.ts`; no manual pre-auth step is needed.

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
`helpers/auth.ts:loginAsDemoAdmin`. The helper navigates to `/login`, fills the demo credentials,
and waits for the URL to leave `/login`.

- Email: `admin@noavet.jp` / Password: `password`

`master-crud.spec.ts` shares a `BrowserContext` across all tests via `beforeAll` (same pattern
used in `owners-search.spec.ts`, `accounting-smoke.spec.ts`, and `reservations-smoke.spec.ts`).

## Coverage Notes

### Reservation cancelled / no_show filter

`filterCalendarAppointments` (filters out `cancelled`, keeps `no_show`) is unit-tested in:

```
frontend/src/features/reservations/routes/__tests__/ReservationManagement.filter.test.ts
```

E2E validation of this filter would require seeding appointments for the current calendar week —
a date-dependent setup that is intentionally excluded to avoid flakiness.
`reservations-smoke.spec.ts` provides a page-load smoke test; the filter guarantee is the unit test.

### Kana non-distinction search

Tested on two surfaces:
- `/owners` — pet name `ピーター` matched by `ぴ` (hiragana) and `ピ` (katakana)
- `/accounting` — pet name `Iris(イリス)` matched by `いりす` (hiragana) and `イリス` (katakana)

Both use the shared `normalizeKana` utility (`src/lib/normalize-kana.ts`) which converts katakana
to hiragana before comparison. The unit tests for `normalizeKana` live in
`src/lib/normalize-kana.test.ts`.

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
