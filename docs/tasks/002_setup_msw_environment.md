# Task: Setup MSW Testing Environment

## Status
- [x] Install `msw` in frontend (via Docker)
- [x] Create `frontend/src/testing/mocks/handlers.ts`
- [x] Create `frontend/src/testing/mocks/node.ts`
- [x] Create `frontend/src/testing/mocks/browser.ts`
- [x] Update `frontend/src/testing/setup.ts` to start/stop MSW
- [x] Verify MSW setup with a test case (`src/testing/msw.test.ts`)

## Description
The testing environment is now equipped with MSW (Mock Service Worker). This allows for reliable frontend integration tests by mocking API responses without requiring a running backend.

## Usage
Add feature-specific handlers to `src/testing/mocks/handlers.ts` and use them in your Vitest tests.
