import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  testMatch: "**/*.spec.ts",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: "html",
  // Vite dev mode serves many unbundled ES modules; each requires a separate
  // HTTP request through the Docker network bridge (host.docker.internal).
  // 60 s for navigation and 90 s per test comfortably covers that overhead.
  timeout: 90000,
  // Lazy route chunks load after the main bundle; bump assertion timeout so
  // expect(locator).toBeVisible() survives the extra Vite download round-trip.
  expect: { timeout: 15000 },
  use: {
    baseURL: process.env.PLAYWRIGHT_TEST_BASE_URL || "http://localhost:3003",
    trace: "on-first-retry",
    navigationTimeout: 60000,
    actionTimeout: 30000,
  },
  // The frontend/backend dev servers must be started separately via docker compose.
  // This project uses Docker Compose for local app startup.
  webServer: undefined,
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
