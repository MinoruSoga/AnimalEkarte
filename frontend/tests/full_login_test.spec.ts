import { test, expect } from '@playwright/test';

test('Full login flow via context.request', async ({ page, context }) => {
  // Step 1: Playwright の context.request でログイン
  const loginResponse = await context.request.post('http://frontend:3000/api/v1/login', {
    data: { email: 'admin@example.com', password: 'password' },
    headers: { 'Content-Type': 'application/json' },
  });

  expect(loginResponse.ok()).toBe(true);

  const cookies = await context.cookies();
  console.log('Cookies after login:', cookies.map(c => c.name));

  // Step 2: ページにアクセス
  await page.goto('/');
  await page.waitForTimeout(5000);

  console.log('URL after goto /:', page.url());

  // /login に飛ばされていないはず
  const finalUrl = page.url();
  console.log('Final URL:', finalUrl);
  console.log('Login success:', !finalUrl.includes('/login'));
});
