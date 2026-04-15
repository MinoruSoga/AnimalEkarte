import { test, expect } from '@playwright/test';

test('Debug Login Page', async ({ page }) => {
  await page.goto('/login');
  
  // 初期状態の保存
  console.log('--- Initial Page State ---');
  console.log(await page.content());

  await page.fill('#login-email', 'admin@example.com');
  await page.fill('#login-password', 'password');
  
  console.log('--- After filling form ---');
  await page.click('button:has-text("ログイン")');
  
  // 3秒待機して状態を確認
  await page.waitForTimeout(3000);
  
  console.log('--- After clicking login (3s wait) ---');
  console.log('Current URL:', page.url());
  
  const errorMsg = await page.locator('#login-error').textContent();
  if (errorMsg) {
    console.log('Error Message found:', errorMsg);
  } else {
    console.log('No error message found.');
  }

  await page.screenshot({ path: 'debug_login.png' });
});
