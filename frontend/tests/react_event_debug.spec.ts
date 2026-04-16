import { test } from '@playwright/test';

test('React 19 event system debug', async ({ page }) => {
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log('BROWSER:', msg.text());
    }
  });
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  // React root の確認
  const rootInfo = await page.evaluate(() => {
    const root = document.getElementById('root');
    const reactRootKey = Object.keys(root ?? {}).find(k => k.startsWith('__reactFiber') || k.startsWith('__reactContainer'));

    // formの submit イベントリスナーを確認
    const form = document.querySelector('#login-form') as HTMLFormElement;
    const formKeys = Object.keys(form ?? {}).filter(k => k.startsWith('__react'));

    // React が登録している root イベントリスナーを調査
    const allListeners = (window as typeof window & { __reactListeners?: unknown })['__reactListeners'];

    return {
      hasRoot: !!root,
      reactRootKey,
      formReactKeys: formKeys,
      hasAllListeners: !!allListeners,
    };
  });
  console.log('Root info:', JSON.stringify(rootInfo, null, 2));

  // フォームに submit イベントハンドラを注入してデバッグ
  await page.evaluate(() => {
    const form = document.querySelector('#login-form') as HTMLFormElement;
    const originalHandler = form.onsubmit;

    // submit イベントを監視
    form.addEventListener('submit', (e) => {
      console.log('FORM SUBMIT EVENT FIRED', 'defaultPrevented:', e.defaultPrevented);
    }, { capture: true });  // capture phase でも確認

    // React root でのイベントも確認
    document.addEventListener('submit', (e) => {
      console.log('DOCUMENT SUBMIT EVENT', 'target:', (e.target as HTMLElement)?.id);
    }, { capture: true });

    console.log('Event listeners added');
  });

  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');

  console.log('Clicking button...');
  await page.locator('button[type="submit"]').first().click();
  await page.waitForTimeout(2000);
  console.log('URL after click:', page.url());

  // 手動で requestSubmit も試す
  console.log('Calling requestSubmit...');
  await page.evaluate(() => {
    const form = document.querySelector('#login-form') as HTMLFormElement;
    form.requestSubmit();
  });
  await page.waitForTimeout(2000);
  console.log('URL after requestSubmit:', page.url());
});
