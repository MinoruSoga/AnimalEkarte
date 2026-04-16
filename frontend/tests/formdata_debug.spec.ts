import { test } from '@playwright/test';

test('FormData value check after fill', async ({ page }) => {
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log('BROWSER:', msg.text());
    }
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  // submit イベントを傍受して FormData の内容を確認
  await page.evaluate(() => {
    const root = document.getElementById('root');
    root?.addEventListener('submit', (e: Event) => {
      const submitEvent = e as SubmitEvent;
      if (submitEvent.target instanceof HTMLFormElement) {
        const formData = new FormData(submitEvent.target);
        console.log('FormData entries:');
        for (const [key, value] of formData.entries()) {
          console.log(`  ${key} = ${key.includes('password') ? '***' : value}`);
        }
      }
    });
  });

  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');

  // 入力後の DOM value を確認
  const domValues = await page.evaluate(() => {
    const email = document.querySelector('#login-email') as HTMLInputElement;
    const pass = document.querySelector('#login-password') as HTMLInputElement;
    const form = document.querySelector('#login-form') as HTMLFormElement;

    // FormData を作成して確認
    const fd = new FormData(form);
    return {
      emailDomValue: email.value,
      passLength: pass.value.length,
      fdEmail: fd.get('login-email'),
      fdPass: fd.get('login-password') ? '***set***' : '***empty***',
    };
  });
  console.log('DOM values before submit:', JSON.stringify(domValues, null, 2));

  await page.locator('button[type="submit"]').first().click();
  await page.waitForTimeout(3000);
  console.log('Final URL:', page.url());
});
