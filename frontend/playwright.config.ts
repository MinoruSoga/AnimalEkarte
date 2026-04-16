import { defineConfig, devices } from '@playwright/test';

/**
 * BUG-001 修正サマリー:
 *
 * 根本原因:
 *   `src/lib/axios.ts` の requestInterceptor が `crypto.randomUUID()` を呼んでいたが、
 *   `http://frontend:3000` (Docker 内ホスト名) は非セキュアコンテキスト (non-secure context)
 *   のため `crypto.randomUUID` が存在せず TypeError が発生し、全ての axios リクエストが
 *   無音で失敗していた。
 *
 * 修正内容:
 *   `src/lib/axios.ts`: `crypto.randomUUID` の存在チェックを追加し、フォールバックを実装。
 *
 * 認証戦略:
 *   - globalSetup で `context.request.post('/api/v1/login')` を使いバックエンドに直接ログイン
 *   - Cookie を storageState として保存し、認証済みのブラウザコンテキストをテストに提供
 *   - これにより React 19 formAction と Playwright の click() の互換性問題も同時に回避
 *
 * React 19 formAction の問題 (参考情報):
 *   React 19 の <form action={asyncFn}> は、フォームの action 属性を
 *   "javascript:throw new Error('A React form was unexpectedly submitted.')" に設定し、
 *   ネイティブ form.submit() をブロックする。Playwright の button.click() は
 *   ブラウザのネイティブ submit を発火させ、React の formAction を経由しない。
 *   → ログインフォームの UI テストには `form.requestSubmit()` を使うこと
 *      (auth-helper.ts の `submitReactForm()` を参照)
 */

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'list',
  globalSetup: './tests/global-setup.ts',
  use: {
    baseURL: process.env.BASE_URL || 'http://frontend:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    // globalSetup で保存した認証済み Cookie を使用
    storageState: '/tmp/playwright-auth-state.json',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
