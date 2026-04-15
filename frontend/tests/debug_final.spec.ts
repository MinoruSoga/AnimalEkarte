import { test, expect } from '@playwright/test';

test('Debug Final State Fixed', async ({ page, request, context }) => {
  const response = await request.post('http://backend:8080/api/v1/login', {
    data: { email: 'admin@example.com', password: 'password' }
  });
  const headers = response.headers();
  const setCookies = headers['set-cookie'].split('\n'); // 実際には配列で来るはず
  
  const cookies = [];
  for (const raw of setCookies) {
    const mainPart = raw.split(';')[0];
    const [name, value] = mainPart.split('=');
    cookies.push({
      name: name.trim(),
      value: value.trim(),
      domain: 'frontend',
      path: name.trim() === 'refresh_token' ? '/api/v1/auth/refresh' : '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax'
    });
  }
  
  await context.addCookies(cookies);

  console.log('Navigating to dashboard...');
  await page.goto('/');
  await page.waitForTimeout(5000);
  
  console.log('Current URL:', page.url());
  
  if (page.url().includes('/login')) {
    console.log('FAIL: Still on login page.');
  } else {
    console.log('SUCCESS: Beyond login page. Path:', page.url());
  }
});
