# Playwright E2E Tests

## Overview

End-to-end tests for the Animal Ekarte master pages (Master CRUD) using Playwright.

### Test Coverage

- **Test A**: Chief Complaint navigation from settings
- **Test B**: Treatment Plan - Procedure tab and parent category selector
- **Test C**: Root items with children cannot change parent
- **Test D**: All 5 treatment plan tabs display correctly

## Prerequisites

### 1. Running Services

Before executing E2E tests, start the development environment:

```bash
docker compose up -d frontend backend db
```

Verify that:
- Frontend is accessible at http://localhost:3003
- Backend API is running on http://localhost:8080
- Database is initialized

### 2. Authentication

These tests assume the user is already **authenticated via a demo/seed session**.

If you encounter redirect to `/login`, you need to:

1. Add a Playwright auth setup (beforeEach hook) to automate login
2. Or manually log in via the browser before running tests

Current implementation does NOT include explicit login automation.

### 3. Architecture Support

**⚠️ Important**: Playwright E2E tests are **NOT supported on arm64 Docker environments** (Apple Silicon, etc.).

- **Linux x86_64**: Fully supported ✅
- **Linux arm64/aarch64**: Tests are automatically skipped
- **macOS arm64**: May work with native Playwright (not Docker)

If you run `pnpm test:e2e` on arm64 Docker:
```
✓ All tests skipped: "Playwright Chromium is not supported in this Docker arm64 runtime"
```

This is expected behavior. For CI automation, use x86_64 runners.

## Running Tests

### Local Development

```bash
# Run all E2E tests (headless)
pnpm test:e2e

# Run with UI (interactive mode)
pnpm test:e2e:ui

# Run specific test file
pnpm test:e2e -- e2e/master-crud.spec.ts
```

### Environment Variables

Customize the base URL:

```bash
PLAYWRIGHT_TEST_BASE_URL=http://localhost:3003 pnpm test:e2e
```

Default: `http://localhost:3003`

### Test Results

After running tests:
- **HTML Report**: `frontend/playwright-report/`
- **Test Results**: `frontend/test-results/`

These directories are added to `.gitignore` and won't be committed.

## Troubleshooting

### Tests are skipped with "Chromium not supported"

✅ Expected on arm64 Docker. This is not an error.

To run E2E tests, use:
- GitHub Actions (Linux x86_64)
- Native x86_64 Docker on Linux
- Local macOS with native Playwright

### Tests fail with "Failed to navigate"

1. Verify frontend is running: `curl http://localhost:3003`
2. Verify authentication: Check if redirected to login page
3. Check browser compatibility: Requires Chromium (installed via `playwright install chromium`)

### Tests fail with "Element not found"

1. Check if the UI structure has changed (e.g., button text)
2. Update selectors in test file
3. Run with `--ui` flag to debug: `pnpm test:e2e:ui`

## Future Enhancements

- [ ] Add explicit login automation (beforeEach hook)
- [ ] Integrate into CI/CD (separate E2E job for x86_64)
- [ ] Add API mocking for isolated component testing
- [ ] Add visual regression testing
- [ ] Increase test coverage (CRUD operations with data cleanup)
