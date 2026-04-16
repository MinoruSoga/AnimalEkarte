import { test } from '@playwright/test';

test('Direct axios call test', async ({ page }) => {
  const capturedRequests: string[] = [];

  await page.route('**/*', async (route) => {
    const req = route.request();
    capturedRequests.push(`${req.method()} ${req.url()}`);
    await route.continue();
  });

  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, msg.text().substring(0, 500));
    }
  });
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));

  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(2000);

  // axios のモジュールが利用可能か確認しつつ、/v1/me へのリクエスト
  await page.evaluate(async () => {
    console.log('Testing XMLHttpRequest directly...');
    const xhr = new XMLHttpRequest();
    xhr.open('GET', '/v1/me', false); // synchronous
    try {
      xhr.send();
      console.log('XHR /v1/me status:', xhr.status);
    } catch (e) {
      console.log('XHR error:', String(e));
    }

    console.log('Testing XMLHttpRequest /api/v1/me...');
    const xhr2 = new XMLHttpRequest();
    xhr2.open('GET', '/api/v1/me', false); // synchronous
    try {
      xhr2.send();
      console.log('XHR /api/v1/me status:', xhr2.status);
    } catch (e) {
      console.log('XHR error:', String(e));
    }
  });

  await page.waitForTimeout(1000);

  const v1Requests = capturedRequests.filter(r => r.includes('/v1/'));
  console.log('v1 requests captured:', v1Requests);
  console.log('All captured count:', capturedRequests.length);
});
