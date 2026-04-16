import { test } from '@playwright/test';

test('Print ALL requests', async ({ page }) => {
  const allRequests: string[] = [];
  page.on('request', req => {
    allRequests.push(`${req.method()} ${req.url()}`);
  });
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, msg.text().substring(0, 200));
    }
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(3000);

  // 全リクエストをそのまま出力
  console.log('=== ALL REQUESTS ===');
  for (const r of allRequests) {
    console.log(r);
  }
  console.log('=== END ===');
  console.log('Total requests:', allRequests.length);
});
