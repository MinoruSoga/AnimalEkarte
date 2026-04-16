import { test, expect } from '@playwright/test';

test('Auth state when Cookie is pre-set', async ({ page, context }) => {
  const apiRequests: string[] = [];

  page.on('request', req => {
    const url = req.url();
    if (url.includes('/api/v1') || url.includes('/v1/')) {
      console.log('API REQUEST:', req.method(), url);
      apiRequests.push(url);
    }
  });
  page.on('response', res => {
    const url = res.url();
    if (url.includes('/api/v1') || url.includes('/v1/')) {
      console.log('API RESPONSE:', res.status(), url);
    }
  });
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, msg.text().substring(0, 300));
    }
  });
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));

  // Step 1: ログインしてCookieを設定
  const loginResp = await context.request.post('http://frontend:3000/api/v1/login', {
    data: { email: 'admin@example.com', password: 'password' },
    headers: { 'Content-Type': 'application/json' },
  });
  console.log('Login status:', loginResp.status());

  const cookiesBeforeNav = await context.cookies();
  console.log('Cookies before navigation:', cookiesBeforeNav.map(c => c.name));

  // Step 2: ページに移動
  console.log('Navigating to /...');
  await page.goto('/');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(5000);

  console.log('URL after navigation:', page.url());
  console.log('API requests during navigation:', apiRequests);
});
