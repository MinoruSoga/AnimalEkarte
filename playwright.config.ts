import { defineConfig, devices } from '@playwright/test';
import path from 'path';

const STORAGE_STATE_PATH = '/tmp/playwright-auth-state.json';
const FRONTEND_URL = process.env.BASE_URL || 'http://localhost:3003'; // docker-compose port mapping

export default defineConfig({
  testDir: './frontend/tests',
  testMatch: '**/*.spec.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html'],
    ['json', { outputFile: 'playwright-report/results.json' }],
    ['junit', { outputFile: 'playwright-report/junit.xml' }],
  ],

  use: {
    baseURL: FRONTEND_URL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        // globalSetup で保存された storageState (Cookie) を読み込む
        storageState: STORAGE_STATE_PATH,
      },
    },
  ],

  // globalSetup: テスト開始前に一度だけ実行 → Cookie/storageState を保存
  globalSetup: require.resolve('./frontend/tests/global-setup.ts'),

  webServer: undefined, // docker-compose で既に起動しているため不要

  timeout: 30 * 1000, // 30s per test
});
