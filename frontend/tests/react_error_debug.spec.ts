import { test } from '@playwright/test';

test('React 19 formAction error detection', async ({ page }) => {
  const errors: string[] = [];

  page.on('console', msg => {
    const text = msg.text();
    if (!text.includes('[vite]') && !text.includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, text);
    }
    if (msg.type() === 'error') errors.push(text);
  });
  page.on('pageerror', err => {
    console.log('PAGE ERROR:', err.message);
    errors.push(err.message);
  });

  // XHR/fetch をモンキーパッチして API 呼び出しを確認
  await page.addInitScript(() => {
    const origFetch = window.fetch;
    window.fetch = async (...args) => {
      console.log('FETCH called:', args[0]);
      const result = await origFetch(...args);
      console.log('FETCH response:', result.status, args[0]);
      return result;
    };

    // axios は XMLHttpRequest も使う
    const OrigXHR = window.XMLHttpRequest;
    class TrackedXHR extends OrigXHR {
      open(method: string, url: string | URL, ...rest: Parameters<XMLHttpRequest['open']> extends [string, string | URL, ...infer R] ? R : never) {
        console.log('XHR open:', method, url.toString());
        return super.open(method, url.toString(), ...(rest as [boolean?, string?, string?]));
      }
    }
    window.XMLHttpRequest = TrackedXHR as unknown as typeof XMLHttpRequest;
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');

  console.log('Submitting form...');
  await page.locator('button[type="submit"]').first().click();

  await page.waitForTimeout(5000);
  console.log('Final URL:', page.url());
  console.log('Errors:', errors);
});
