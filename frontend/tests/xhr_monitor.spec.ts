import { test } from '@playwright/test';

test('XHR/Axios monitoring', async ({ page }) => {
  // addInitScript は page navigate 前に登録する必要がある
  await page.addInitScript(() => {
    // axios は XMLHttpRequest を使用する
    const OrigXHR = window.XMLHttpRequest;
    const xhrLog: string[] = [];

    function TrackedXHR(this: XMLHttpRequest) {
      const xhr = new OrigXHR() as XMLHttpRequest & { _url?: string; _method?: string };

      const origOpen = xhr.open.bind(xhr);
      xhr.open = function(method: string, url: string | URL, async?: boolean, user?: string | null, password?: string | null) {
        const urlStr = url.toString();
        xhrLog.push(`XHR.open: ${method} ${urlStr}`);
        console.log('XHR.open:', method, urlStr);
        (window as typeof window & { __xhrLog: string[] }).__xhrLog = xhrLog;
        return origOpen(method, url, async ?? true, user, password);
      };

      return xhr;
    }
    TrackedXHR.prototype = OrigXHR.prototype;
    window.XMLHttpRequest = TrackedXHR as unknown as typeof XMLHttpRequest;
    console.log('XHR monkey-patched');
  });

  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, msg.text());
    }
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  // 初期ロード時に /api/v1/me が呼ばれるはず
  const xhrLogAfterLoad = await page.evaluate(() => (window as typeof window & { __xhrLog?: string[] }).__xhrLog ?? []);
  console.log('XHR log after load:', xhrLogAfterLoad);

  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');

  await page.locator('button[type="submit"]').first().click();
  await page.waitForTimeout(3000);

  const xhrLogAfterSubmit = await page.evaluate(() => (window as typeof window & { __xhrLog?: string[] }).__xhrLog ?? []);
  console.log('XHR log after submit:', xhrLogAfterSubmit);
  console.log('Final URL:', page.url());
});
