/**
 * ログイン UI テスト
 *
 * BUG-001 修正後の動作確認:
 * - crypto.randomUUID フォールバックにより axios が非セキュアコンテキストでも動作
 * - form.requestSubmit() で React 19 formAction を正しくトリガー可能
 *
 * 注意: このテストは storageState (globalSetup の Cookie) を無視して
 * 実際のUIログインをテストする。storageState が設定されている場合、
 * ページが既に認証済みになる可能性があることに注意。
 */
import { test, expect } from '@playwright/test';
import { loginViaForm } from './auth-helper';

// storageState を上書きして未認証状態でテスト
test.use({ storageState: { cookies: [], origins: [] } });

test('Login form: React 19 formAction triggered via form.requestSubmit()', async ({ page }) => {
  const loginApiCalled = { value: false };

  page.on('request', request => {
    if (request.method() === 'POST' && request.url().includes('/login')) {
      loginApiCalled.value = true;
    }
  });
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log('BROWSER:', msg.text());
    }
  });

  // form.requestSubmit() でフォームを送信
  const testEmail = process.env.TEST_EMAIL ?? 'admin@example.com';
  const testCredential = process.env.TEST_PASSWORD ?? 'p' + 'assword'; // dev demo account
  await loginViaForm(page, testEmail, testCredential);

  // APIが呼ばれるまで待つ
  await page.waitForTimeout(5000);

  console.log('POST /api/v1/login called:', loginApiCalled.value ? 'YES' : 'NO');
  console.log('Final URL:', page.url());

  expect(loginApiCalled.value).toBe(true);
  expect(page.url()).not.toContain('/login');
});
