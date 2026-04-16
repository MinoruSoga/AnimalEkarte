import { test } from '@playwright/test';

test('Diagnose React 19 formAction login', async ({ page }) => {
  const requests: string[] = [];
  page.on('request', request => {
    const url = request.url();
    if (!url.includes('fonts') && !url.includes('.woff') && !url.includes('gstatic')) {
      requests.push(`${request.method()} ${url}`);
    }
  });
  page.on('console', msg => console.log('BROWSER:', msg.text()));
  page.on('pageerror', err => console.log('PAGE ERROR:', err.message));

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  // React の controlled input への入力
  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');

  const emailVal = await page.inputValue('#login-email');
  const passVal = await page.inputValue('#login-password');
  console.log('Email value after fill:', emailVal);
  console.log('Password set:', passVal.length > 0 ? 'yes' : 'no');

  // formAction の属性を確認（React 19 では "javascript:void(0)" や空になる可能性）
  const formActionAttr = await page.$eval('#login-form', (el: HTMLFormElement) => el.getAttribute('action'));
  console.log('Form action attribute:', formActionAttr);

  const submitBtnText = await page.locator('button[type="submit"]').first().textContent();
  console.log('Submit button text:', submitBtnText);

  console.log('Clicking submit button...');
  await page.locator('button[type="submit"]').first().click();

  await page.waitForTimeout(5000);
  console.log('URL after click:', page.url());

  const loginApiCalled = requests.some(r => r.includes('/login') && r.startsWith('POST'));
  console.log('POST /login called:', loginApiCalled ? 'YES' : 'NO');
  console.log('All non-asset requests:', requests.filter(r => !r.includes('.ts') && !r.includes('.js') && !r.includes('.css')).join('\n'));
});
