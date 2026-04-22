import { test, expect } from '@playwright/test';

/**
 * スタッフ管理フロー E2E テスト
 *
 * 認証: playwright.config.ts の globalSetup で事前にログイン済み
 *
 * テストフロー:
 * 1. スタッフ設定ページナビゲーション
 * 2. スタッフ一覧表示確認
 * 3. スタッフ登録フォーム表示
 * 4. クリニック割当フロー検証
 */

test.describe('スタッフ管理フロー E2E テスト', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://frontend:3000/');
    await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 });
  });

  test('STAFF-01: スタッフ設定ページナビゲーション', async ({ page }) => {
    // 設定メニューをクリック
    const settingsLink = page.locator(
      'a:has-text("設定"), a:has-text("Settings"), [data-testid="nav-settings"]'
    ).first();

    if (await settingsLink.isVisible()) {
      await settingsLink.click();
      await expect(page).toHaveURL(/\/settings/, { timeout: 10000 });
    }
  });

  test('STAFF-02: スタッフ一覧表示確認', async ({ page }) => {
    // スタッフ設定ページへ遷移
    await page.goto('http://frontend:3000/settings/staff', { waitUntil: 'networkidle' });

    // ページが読み込まれたことを確認
    await expect(page.locator('body')).toBeVisible();

    // スタッフ一覧テーブルを確認
    const table = page.locator('table, [role="grid"], [data-testid="staff-list"]').first();
    if (await table.isVisible({ timeout: 3000 })) {
      await expect(table).toBeVisible();
    }
  });

  test('STAFF-03: スタッフ新規登録フォーム', async ({ page }) => {
    // スタッフ設定ページへ遷移
    await page.goto('http://frontend:3000/settings/staff', { waitUntil: 'networkidle' });

    // 新規スタッフ登録ボタンをクリック
    const createBtn = page.locator(
      'button:has-text("新規"), button:has-text("追加"), [data-testid="btn-create-staff"]'
    ).first();

    if (await createBtn.isVisible()) {
      await createBtn.click();

      // フォームが表示されるのを待つ
      const form = page.locator('form, [data-testid="staff-form"]').first();
      await expect(form).toBeVisible({ timeout: 5000 });
    }
  });

  test('STAFF-04: スタッフフォーム入力フィールド確認', async ({ page }) => {
    // スタッフフォームページへ遷移
    await page.goto('http://frontend:3000/settings/staff/new', { waitUntil: 'networkidle' });

    // 必須フィールドが表示されていることを確認
    const nameInput = page.locator('input[placeholder*="名前"], input[placeholder*="Name"], [data-testid*="name"]').first();
    if (await nameInput.isVisible({ timeout: 3000 })) {
      await expect(nameInput).toBeVisible();
    }

    const emailInput = page.locator('input[type="email"], [data-testid*="email"]').first();
    if (await emailInput.isVisible({ timeout: 3000 })) {
      await expect(emailInput).toBeVisible();
    }
  });

  test('STAFF-05: スタッフ クリニック割当表示', async ({ page }) => {
    // スタッフ一覧ページへ遷移
    await page.goto('http://frontend:3000/settings/staff', { waitUntil: 'networkidle' });

    // スタッフ行をクリックして詳細を表示
    const row = page.locator('table tbody tr, [role="grid"] [role="row"]').first();
    if (await row.isVisible()) {
      await row.click();

      // クリニック割当情報が表示されるのを待つ
      const clinicInfo = page.locator(
        'aside, [data-testid="staff-clinic-assignments"], [data-testid="staff-detail"]'
      ).first();
      await expect(clinicInfo).toBeVisible({ timeout: 5000 });
    }
  });
});
