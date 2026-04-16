/**
 * Playwright Global Setup
 *
 * React 19 の useActionState (formAction) はネイティブ form.submit() とは互換性がなく、
 * Playwright の button.click() がネイティブ submit をトリガーするため
 * /api/v1/login への POST リクエストが発生しない問題がある。
 *
 * 解決策: globalSetup でバックエンド API に直接 POST してセッション Cookie を取得し、
 * storageState.json に保存する。テストはこの認証済み状態から開始する。
 *
 * 参考: https://playwright.dev/docs/auth#basic-shared-account-in-all-tests
 */

import { chromium, type FullConfig } from '@playwright/test';

const BACKEND_URL = process.env.BACKEND_URL || 'http://backend:8080';
const STORAGE_STATE_PATH = '/tmp/playwright-auth-state.json';

export { STORAGE_STATE_PATH };

export default async function globalSetup(_config: FullConfig) {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  // バックエンドに直接ログイン
  const response = await context.request.post(`${BACKEND_URL}/api/v1/login`, {
    data: {
      email: 'admin@example.com',
      password: 'password',
    },
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok()) {
    const body = await response.text();
    throw new Error(`Login failed: ${response.status()} ${body}`);
  }

  // フロントエンドへアクセスして Cookie を正しいドメインに設定する
  // (Set-Cookie は backend ドメインに設定されているため、
  //  Vite proxy 経由でフロントエンドの origin から送信させる必要がある)
  const frontendUrl = process.env.BASE_URL || 'http://frontend:3000';

  // バックエンドから返ってきた Cookie をフロントエンドのコンテキストに追加
  const rawCookies = response.headers()['set-cookie'];
  if (rawCookies) {
    const cookieLines = rawCookies.split('\n');
    const cookies = cookieLines
      .map(raw => {
        const parts = raw.split(';').map(s => s.trim());
        const [nameValue, ...attrs] = parts;
        if (!nameValue?.includes('=')) return null;
        const eqIdx = nameValue.indexOf('=');
        const name = nameValue.slice(0, eqIdx).trim();
        const value = nameValue.slice(eqIdx + 1).trim();

        // path を attrs から取得
        const pathAttr = attrs.find(a => a.toLowerCase().startsWith('path='));
        const path = pathAttr ? pathAttr.split('=')[1] : '/';

        // HttpOnly かどうか
        const httpOnly = attrs.some(a => a.toLowerCase() === 'httponly');

        return {
          name,
          value,
          domain: 'frontend',
          path,
          httpOnly,
          secure: false,
          sameSite: 'Lax' as const,
        };
      })
      .filter((c): c is NonNullable<typeof c> => c !== null);

    await context.addCookies(cookies);
  }

  // storageState を保存（Cookie が含まれる）
  await context.storageState({ path: STORAGE_STATE_PATH });

  // 動作確認: /v1/me が 200 を返すことを確認
  const meResponse = await page.request.get(`${frontendUrl}/api/v1/me`);
  if (!meResponse.ok()) {
    console.warn(
      `[globalSetup] Warning: /v1/me returned ${meResponse.status()}. ` +
      'Cookie injection may need adjustment.'
    );
  } else {
    console.log('[globalSetup] Authentication state saved successfully.');
  }

  await browser.close();
}
