import { test, expect } from '@playwright/test';

/**
 * Section 14: マスタ設定ブラウザテスト
 *
 * 認証: playwright.config.ts の globalSetup で事前にログインして
 * storageState (Cookie) に保存済み。beforeEach のUIログインは不要。
 *
 * BUG-001 修正後: crypto.randomUUID のフォールバックにより axios リクエストが
 * 非セキュアコンテキスト (http://frontend:3000) でも正常に動作するようになった。
 */
test.describe('Section 14: Master Settings Browser Test', () => {
  test.beforeEach(async ({ page }) => {
    // storageState の Cookie で認証済みのため、直接ダッシュボードにアクセスできる
    await page.goto('http://frontend:3000/');
    await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 });
  });

  test('14.1: マスタ共通動作 - 削除確認キャンセル', async ({ page }) => {
    // 職種マスタへ遷移
    await page.goto('http://frontend:3000/settings/occupations');

    // データが表示されるのを待つ
    await page.waitForSelector('table tbody tr', { timeout: 10000 });

    // 最初の行をクリック
    await page.locator('table tbody tr').first().click();

    // サイドパネル表示確認
    await expect(page.locator('aside')).toBeVisible();

    // 削除ボタンをクリック
    await page.getByRole('button', { name: /削除/i }).first().click();

    // キャンセルボタンをクリック
    const cancelBtn = page.getByRole('button', { name: /キャンセル/i });
    await expect(cancelBtn).toBeVisible();
    await cancelBtn.click();

    // ダイアログが閉じたことを確認
    await expect(cancelBtn).not.toBeVisible();
    // パネルは開いたまま
    await expect(page.locator('aside')).toBeVisible();
  });
});
