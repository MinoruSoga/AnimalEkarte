import { test, expect } from '@playwright/test';

test('Debug Login Console', async ({ page }) => {
  page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
  page.on('pageerror', err => console.log('BROWSER ERROR:', err.message));

  await page.goto('/login');
  
  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');
  
  console.log('Clicking login button...');
  const loginBtn = page.getByRole('button', { name: "ログイン" });
  await loginBtn.click();
  
  await page.waitForTimeout(5000);
  console.log('Current URL after click:', page.url());
});
