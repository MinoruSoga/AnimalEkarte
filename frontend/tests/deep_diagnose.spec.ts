import { test } from '@playwright/test';

test('Deep diagnose React 19 event handling', async ({ page }) => {
  const requests: string[] = [];
  page.on('request', request => {
    const url = request.url();
    if (!url.includes('fonts') && !url.includes('.woff') && !url.includes('gstatic')) {
      requests.push(`${request.method()} ${url}`);
    }
  });
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log('BROWSER:', msg.text());
    }
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');

  // React state の値を確認（controlled input の状態）
  const reactState = await page.evaluate(() => {
    const emailEl = document.querySelector('#login-email') as HTMLInputElement;
    const passEl = document.querySelector('#login-password') as HTMLInputElement;
    const form = document.querySelector('#login-form') as HTMLFormElement;

    // React internal fiber を確認
    const emailFiber = Object.keys(emailEl).find(k => k.startsWith('__reactFiber'));

    return {
      emailValue: emailEl?.value,
      passLength: passEl?.value?.length,
      formAction: form?.action,
      formActionAttr: form?.getAttribute('action'),
      hasReactFiber: !!emailFiber,
    };
  });
  console.log('React state:', JSON.stringify(reactState, null, 2));

  // requestSubmit を試す
  console.log('Trying form.requestSubmit()...');
  try {
    await page.evaluate(() => {
      const form = document.querySelector('#login-form') as HTMLFormElement;
      form.requestSubmit();
    });
    console.log('requestSubmit() called successfully');
  } catch (e) {
    console.log('requestSubmit() error:', e);
  }

  await page.waitForTimeout(2000);
  const urlAfterRequestSubmit = page.url();
  console.log('URL after requestSubmit:', urlAfterRequestSubmit);

  // dispatchEvent で submit を試す
  if (urlAfterRequestSubmit.includes('/login')) {
    console.log('requestSubmit did not work, trying dispatchEvent...');
    await page.evaluate(() => {
      const form = document.querySelector('#login-form') as HTMLFormElement;
      const submitEvent = new SubmitEvent('submit', {
        bubbles: true,
        cancelable: true,
        submitter: form.querySelector('button[type="submit"]') as HTMLElement
      });
      form.dispatchEvent(submitEvent);
    });
    await page.waitForTimeout(2000);
    console.log('URL after dispatchEvent:', page.url());
  }

  // React event simulation
  if (page.url().includes('/login')) {
    console.log('Trying React synthetic event simulation...');
    // React 19 のevent delegation は document ではなく root containerに登録される
    await page.evaluate(() => {
      const btn = document.querySelector('button[type="submit"]') as HTMLButtonElement;
      // click の代わりに pointer events を順番に送信
      btn.dispatchEvent(new PointerEvent('pointerdown', { bubbles: true }));
      btn.dispatchEvent(new MouseEvent('mousedown', { bubbles: true }));
      btn.dispatchEvent(new PointerEvent('pointerup', { bubbles: true }));
      btn.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));
      btn.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    });
    await page.waitForTimeout(2000);
    console.log('URL after manual click events:', page.url());
  }

  const loginApiCalled = requests.some(r => r.includes('/login') && r.startsWith('POST'));
  console.log('POST /login called:', loginApiCalled ? 'YES' : 'NO');
});
