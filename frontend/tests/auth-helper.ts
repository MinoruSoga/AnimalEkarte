/**
 * Playwright Auth Helper
 *
 * React 19 useActionState (formAction) と Playwright の互換性問題を回避する
 * ユーティリティ関数。
 *
 * 問題:
 *   React 19 は <form action={asyncFn}> の action 属性を
 *   "javascript:throw new Error('A React form was unexpectedly submitted...')" に設定し、
 *   ネイティブの form.submit() をブロックする。
 *   Playwright の button[type="submit"].click() はブラウザのネイティブ submit を
 *   発火させるため、React の formAction が実行されない。
 *
 * 解決策:
 *   - テストで認証が必要な場合: storageState (globalSetup で保存した Cookie) を使う
 *   - ログインフォーム自体をテストする場合: form.requestSubmit() を使う
 */

import type { Page } from '@playwright/test';

/**
 * React 19 formAction に対応したフォーム送信。
 * form.requestSubmit() は React の submit ハンドラを正しくトリガーする。
 *
 * @param page Playwright Page
 * @param formSelector フォームのセレクタ (デフォルト: '#login-form')
 */
export async function submitReactForm(
  page: Page,
  formSelector: string = '#login-form'
): Promise<void> {
  await page.evaluate((selector) => {
    const form = document.querySelector(selector) as HTMLFormElement | null;
    if (!form) throw new Error(`Form not found: ${selector}`);
    // requestSubmit() は submit イベントを発火させ、React の formAction が処理する
    // form.submit() (native) は React がブロックするため使用不可
    form.requestSubmit();
  }, formSelector);
}

/**
 * ログインページでメールとパスワードを入力し、formAction を発火させる。
 * 認証が必要なテストには storageState を使うこと。
 * このヘルパーはログインフォーム自体のテスト専用。
 *
 * @param page Playwright Page
 * @param email メールアドレス
 * @param password パスワード
 */
export async function loginViaForm(
  page: Page,
  email: string,
  password: string
): Promise<void> {
  await page.goto('/login');
  await page.waitForLoadState('networkidle');

  await page.fill('#login-email', email);
  await page.fill('#login-password', password);

  // React 19 formAction を正しく発火させる
  await submitReactForm(page);
}
