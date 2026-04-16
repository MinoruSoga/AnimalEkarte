import { test } from '@playwright/test';

test('React 19 root container event debug', async ({ page }) => {
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log('BROWSER:', msg.text());
    }
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  // React root container のイベントリスナーを調査
  const rootEventInfo = await page.evaluate(() => {
    const root = document.getElementById('root');
    if (!root) return { error: 'no root' };

    const containerKey = Object.keys(root).find(k => k.startsWith('__reactContainer'));
    if (!containerKey) return { error: 'no container key' };

    // React の内部 fiber を取得してイベントハンドラを確認
    const fiber = (root as Record<string, unknown>)[containerKey];

    // getEventListeners は Chrome DevTools API なので使えないが、
    // React の internal props を確認する
    const form = document.querySelector('#login-form') as HTMLFormElement;
    const formPropsKey = Object.keys(form).find(k => k.startsWith('__reactProps'));
    if (!formPropsKey) return { error: 'no form props key' };

    const formProps = (form as Record<string, unknown>)[formPropsKey] as Record<string, unknown>;

    return {
      containerKey,
      formPropKeys: Object.keys(formProps),
      formAction: typeof formProps['action'],
      hasOnSubmit: 'onSubmit' in formProps,
      reactVersion: (window as typeof window & { React?: { version: string } }).React?.version,
    };
  });
  console.log('Root event info:', JSON.stringify(rootEventInfo, null, 2));

  // React の submit イベント処理タイミングを確認するため、bubble phase 後に確認
  await page.evaluate(() => {
    const form = document.querySelector('#login-form') as HTMLFormElement;

    // bubble phase の最後（window）で確認
    window.addEventListener('submit', (e) => {
      console.log('WINDOW SUBMIT (bubble):', 'defaultPrevented:', e.defaultPrevented);
    });

    // capture phase の最初（window）で確認
    window.addEventListener('submit', (e) => {
      console.log('WINDOW SUBMIT (capture):', 'defaultPrevented:', e.defaultPrevented);
    }, { capture: true });

    // root container で確認
    const root = document.getElementById('root');
    root?.addEventListener('submit', (e) => {
      console.log('ROOT SUBMIT (bubble):', 'defaultPrevented:', e.defaultPrevented);
    });
    root?.addEventListener('submit', (e) => {
      console.log('ROOT SUBMIT (capture):', 'defaultPrevented:', e.defaultPrevented);
    }, { capture: true });
  });

  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');

  await page.locator('button[type="submit"]').first().click();
  await page.waitForTimeout(3000);
  console.log('Final URL:', page.url());
});
