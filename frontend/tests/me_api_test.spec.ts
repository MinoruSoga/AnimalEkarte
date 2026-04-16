import { test, expect } from '@playwright/test';

test('Cookie set then /api/v1/me call', async ({ page, context }) => {
  page.on('request', req => {
    if (req.url().includes('/v1/') || req.url().includes('/api/')) {
      console.log('REQUEST:', req.method(), req.url(), 'cookies:', req.headers()['cookie'] ? 'YES' : 'NO');
    }
  });
  page.on('response', res => {
    if (res.url().includes('/v1/') || res.url().includes('/api/')) {
      console.log('RESPONSE:', res.status(), res.url());
    }
  });

  // ログイン
  const loginResponse = await context.request.post('http://frontend:3000/api/v1/login', {
    data: { email: 'admin@example.com', password: 'password' },
    headers: { 'Content-Type': 'application/json' },
  });
  expect(loginResponse.ok()).toBe(true);

  // /me を手動で呼んでCookieが効くか確認
  const meResponse = await context.request.get('http://frontend:3000/api/v1/me');
  console.log('/me via context.request:', meResponse.status());

  // ページから fetch で /me を呼んでも確認
  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  const meFromPage = await page.evaluate(async () => {
    try {
      const resp = await fetch('/api/v1/me', { credentials: 'include' });
      return { status: resp.status };
    } catch (e) {
      return { error: String(e) };
    }
  });
  console.log('/me via page.evaluate fetch:', meFromPage);

  // ページから axios 的な XHR でも確認
  const meFromXHR = await page.evaluate(() => {
    return new Promise<{ status: number }>((resolve) => {
      const xhr = new XMLHttpRequest();
      xhr.open('GET', '/api/v1/me');
      xhr.withCredentials = true;
      xhr.onload = () => resolve({ status: xhr.status });
      xhr.onerror = () => resolve({ status: -1 });
      xhr.send();
    });
  });
  console.log('/me via page.evaluate XHR:', meFromXHR);
});
