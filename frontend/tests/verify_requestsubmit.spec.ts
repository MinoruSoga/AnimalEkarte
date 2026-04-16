/**
 * requestSubmit() が React 19 formAction を正しく発火させることを確認するテスト
 */
import { test, expect } from '@playwright/test';
import { loginViaForm } from './auth-helper';

test('React 19 formAction is triggered via form.requestSubmit()', async ({ page }) => {
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
  await loginViaForm(page, 'admin@example.com', 'password');

  // APIが呼ばれるまで待つ
  await page.waitForTimeout(5000);

  console.log('POST /api/v1/login called:', loginApiCalled.value ? 'YES' : 'NO');
  console.log('Final URL:', page.url());

  expect(loginApiCalled.value).toBe(true);
  expect(page.url()).not.toContain('/login');
});
