import { test } from '@playwright/test';

test('Login via Vite proxy and check Cookie', async ({ page, context }) => {
  // Playwright の context.request でViteプロキシ経由ログイン
  const loginResponse = await context.request.post('http://frontend:3000/api/v1/login', {
    data: { email: 'admin@example.com', password: 'password' },
    headers: { 'Content-Type': 'application/json' },
  });

  console.log('Login status:', loginResponse.status());
  const headers = loginResponse.headers();
  console.log('Response headers:', JSON.stringify(headers, null, 2));

  // Playwright のcontext.cookiesを確認
  const cookies = await context.cookies();
  console.log('Cookies after login:', cookies.map(c => `${c.name}=${c.value.substring(0, 20)}... (domain: ${c.domain}, path: ${c.path})`));
});
