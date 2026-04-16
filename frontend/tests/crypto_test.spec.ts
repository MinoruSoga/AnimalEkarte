import { test } from '@playwright/test';

test('crypto.randomUUID availability in non-secure context', async ({ page }) => {
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, msg.text());
    }
  });
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  const cryptoInfo = await page.evaluate(() => {
    return {
      hasCrypto: typeof crypto !== 'undefined',
      hasRandomUUID: typeof crypto?.randomUUID === 'function',
      isSecureContext: window.isSecureContext,
      origin: window.location.origin,
      protocol: window.location.protocol,
      // 実際に呼んでみる
      uuidResult: (() => {
        try {
          return crypto.randomUUID();
        } catch (e) {
          return `ERROR: ${String(e)}`;
        }
      })(),
    };
  });
  console.log('Crypto info:', JSON.stringify(cryptoInfo, null, 2));
});
