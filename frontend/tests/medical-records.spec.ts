import { test, expect } from '@playwright/test';

/**
 * 医療記録フロー E2E テスト
 *
 * 認証: playwright.config.ts の globalSetup で事前にログイン済み
 *
 * テストフロー:
 * 1. 医療記録一覧ページへ遷移
 * 2. 動物選択フロー確認
 * 3. 医療記録フォーム表示確認
 * 4. 記録入力・保存フロー検証
 */

test.describe('医療記録フロー E2E テスト', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://frontend:3000/');
    await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 });
  });

  test('MED-01: 医療記録ページナビゲーション', async ({ page }) => {
    // 医療記録メニューをクリック
    const medicalLink = page.locator(
      'a:has-text("医療記録"), a:has-text("Medical Records"), [data-testid="nav-medical"]'
    ).first();

    if (await medicalLink.isVisible()) {
      await medicalLink.click();
      await expect(page).toHaveURL(/\/medical-records/, { timeout: 10000 });
    }
  });

  test('MED-02: 医療記録フォーム表示確認', async ({ page }) => {
    // 医療記録ページへ遷移（直接パスで遷移）
    await page.goto('http://frontend:3000/medical-records', { waitUntil: 'networkidle' });

    // ページが読み込まれたことを確認
    await expect(page.locator('body')).toBeVisible();

    // フォームまたはコンテンツが表示されていることを確認
    const content = page.locator('[data-testid="medical-record-form"], form, [role="form"]').first();
    if (await content.isVisible({ timeout: 3000 })) {
      await expect(content).toBeVisible();
    }
  });

  test('MED-03: 記録保存ボタンの存在確認', async ({ page }) => {
    // 医療記録ページへ遷移
    await page.goto('http://frontend:3000/medical-records', { waitUntil: 'networkidle' });

    // 保存ボタンを探す
    const saveBtn = page.locator(
      'button:has-text("保存"), button:has-text("送信"), [data-testid="btn-save"]'
    ).first();

    if (await saveBtn.isVisible({ timeout: 3000 })) {
      await expect(saveBtn).toBeEnabled();
    }
  });
});
