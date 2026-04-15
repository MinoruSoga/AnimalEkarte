import { test, expect } from '@playwright/test';

test.describe('Section 14: Master Settings Browser Test (UI Login)', () => {
  test.beforeEach(async ({ page }) => {
    // 1. ログイン画面へ。Vite のポート 3000 にアクセス。
    await page.goto('http://frontend:3000/login');
    
    // 2. デモアカウント (安田 希恵) をクリック。
    const demoAccount = page.locator('button:has-text("安田 希恵")');
    await demoAccount.waitFor({ state: 'visible', timeout: 5000 });
    await demoAccount.click();
    
    // 3. ログインボタンをクリック。
    // React 19 の formAction を発火させるため、少し待ってからクリック。
    await page.waitForTimeout(500);
    await page.getByRole('button', { name: "ログイン" }).click();
    
    // 4. ダッシュボードへの遷移を待つ。
    // タイムアウトを長めに設定し、Vite のコンパイル待ちを許容。
    await expect(page).not.toHaveURL(/\/login/, { timeout: 20000 });
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
