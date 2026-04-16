import { test } from '@playwright/test';

test('Intercept all routes', async ({ page }) => {
  // 全URLをインターセプト
  await page.route('**/*', async (route) => {
    const req = route.request();
    const url = req.url();
    const method = req.method();

    // APIリクエストをログ
    if (url.includes('/v1/') || url.includes('/api/')) {
      console.log('INTERCEPTED:', method, url);
    }

    // 全リクエストを通す
    await route.continue();
  });

  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, msg.text().substring(0, 300));
    }
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(3000);

  console.log('Final URL:', page.url());
});
