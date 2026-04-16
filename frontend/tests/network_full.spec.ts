import { test } from '@playwright/test';

test('Full network monitoring', async ({ page }) => {
  const allRequests: string[] = [];

  // Playwright の request/response イベントは XHR/fetch を全てキャプチャする
  page.on('request', req => {
    allRequests.push(`${req.method()} ${req.url()}`);
  });
  page.on('response', res => {
    const url = res.url();
    if (url.includes('/api/') || url.includes('backend')) {
      console.log(`RESPONSE: ${res.status()} ${url}`);
    }
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  const apiCallsOnLoad = allRequests.filter(r =>
    r.includes('/api/') || r.includes('backend') || r.includes('v1')
  );
  console.log('API calls on load:', apiCallsOnLoad);

  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');

  await page.locator('button[type="submit"]').first().click();
  await page.waitForTimeout(5000);

  const apiCallsAfterSubmit = allRequests.filter(r =>
    r.includes('/api/') || r.includes('backend') || r.includes('v1')
  );
  console.log('API calls after submit:', apiCallsAfterSubmit);
  console.log('Final URL:', page.url());
});
