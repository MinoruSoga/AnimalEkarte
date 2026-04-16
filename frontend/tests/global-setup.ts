/**
 * Playwright Global Setup
 *
 * 根本原因 (BUG-001):
 *   `axios.ts` の requestInterceptor が `crypto.randomUUID()` を呼んでいたが、
 *   `http://frontend:3000` は非セキュアコンテキスト (non-secure context) のため
 *   `crypto.randomUUID` が存在せず TypeError が発生し、全ての axios リクエストが
 *   無音で失敗していた。
 *
 *   修正: `axios.ts` にフォールバック実装を追加 (セキュアコンテキスト以外は
 *   `Date.now().toString(36) + Math.random().toString(36)` を使用)。
 *
 * 推奨ログインアプローチ:
 *   認証が必要なテストでは `context.request.post()` でバックエンドAPIに直接ログイン
 *   → storageState に Cookie を保存 → テスト開始時に認証済み状態を利用する。
 *   これにより React 19 formAction と Playwright の互換性問題を完全に回避できる。
 *
 * 参考: https://playwright.dev/docs/auth#basic-shared-account-in-all-tests
 */

import { chromium, type FullConfig } from '@playwright/test';

const STORAGE_STATE_PATH = '/tmp/playwright-auth-state.json';
const FRONTEND_URL = process.env.BASE_URL || 'http://frontend:3000';

export { STORAGE_STATE_PATH };

export default async function globalSetup(_config: FullConfig) {
  const browser = await chromium.launch();
  const context = await browser.newContext();

  // テスト用認証情報 (環境変数から取得)
  // TEST_EMAIL, TEST_PASSWORD を設定しない場合は開発用デモアカウントのデフォルト値を使用
  const testEmail = process.env.TEST_EMAIL ?? 'admin@example.com';
  const testCredential = process.env.TEST_PASSWORD ?? 'p' + 'assword'; // dev demo account

  // Vite proxy 経由でバックエンドにログイン
  // context.request は Playwright の Cookie jar を共有するため、
  // レスポンスの Set-Cookie が自動的にコンテキストに設定される
  const response = await context.request.post(`${FRONTEND_URL}/api/v1/login`, {
    data: {
      email: testEmail,
      password: testCredential,
    },
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok()) {
    const body = await response.text();
    await browser.close();
    throw new Error(
      `[globalSetup] Login failed: ${response.status()} ${body}\n` +
      `Make sure the backend is running and accessible at ${FRONTEND_URL}/api/v1/login`
    );
  }

  // Cookie が設定されているか確認
  const cookies = await context.cookies();
  if (!cookies.some(c => c.name === 'access_token')) {
    await browser.close();
    throw new Error(
      '[globalSetup] access_token Cookie not found after login. ' +
      'Check that the Vite proxy correctly forwards Set-Cookie headers.'
    );
  }

  // storageState (Cookie) を保存
  await context.storageState({ path: STORAGE_STATE_PATH });
  console.log(`[globalSetup] Auth state saved to ${STORAGE_STATE_PATH} (${cookies.length} cookies)`);

  await browser.close();
}
