import { test, expect } from '@playwright/test';

test('Debug Login Network', async ({ page }) => {
  // ネットワークリクエストの監視
  page.on('request', request => console.log('>>', request.method(), request.url()));
  page.on('response', response => console.log('<<', response.status(), response.url()));

  await page.goto('/login');
  
  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');
  
  const loginBtn = page.getByRole('button', { name: "ログイン" });
  await loginBtn.click();
  
  await page.waitForTimeout(5000);
  console.log('Final URL:', page.url());
});
