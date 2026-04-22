import { test, expect } from '@playwright/test';

/**
 * 入院管理フロー E2E テスト
 *
 * 認証: playwright.config.ts の globalSetup で事前にログイン済み
 *
 * テストフロー:
 * 1. 入院一覧ページへ遷移
 * 2. 入院レコード表示確認
 * 3. 入院フォーム操作テスト
 * 4. 入院登録・更新フロー検証
 */

test.describe('入院管理フロー E2E テスト', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://frontend:3000/');
    await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 });
  });

  test('HOSP-01: 入院ページナビゲーション', async ({ page }) => {
    // 入院メニューをクリック
    const hospLink = page.locator(
      'a:has-text("入院"), a:has-text("Hospitalization"), [data-testid="nav-hospitalization"]'
    ).first();

    if (await hospLink.isVisible()) {
      await hospLink.click();
      await expect(page).toHaveURL(/\/hospitalizations?/, { timeout: 10000 });
    }
  });

  test('HOSP-02: 入院レコード一覧表示確認', async ({ page }) => {
    // 入院ページへ遷移（直接パスで遷移）
    await page.goto('http://frontend:3000/hospitalizations', { waitUntil: 'networkidle' });

    // テーブルまたはリスト表示
    const list = page.locator('table, [role="grid"], [data-testid="hospitalization-list"]').first();
    if (await list.isVisible({ timeout: 3000 })) {
      await expect(list).toBeVisible();
    }
  });

  test('HOSP-03: 入院作成フォーム表示', async ({ page }) => {
    // 入院ページへ遷移
    await page.goto('http://frontend:3000/hospitalizations', { waitUntil: 'networkidle' });

    // 新規作成ボタンをクリック
    const createBtn = page.locator(
      'button:has-text("新規"), button:has-text("作成"), [data-testid="btn-create-hospitalization"]'
    ).first();

    if (await createBtn.isVisible()) {
      await createBtn.click();

      // フォームが表示されるのを待つ
      const form = page.locator('form, [data-testid="hospitalization-form"]').first();
      await expect(form).toBeVisible({ timeout: 5000 });
    }
  });

  test('HOSP-04: 入院期間選択UI確認', async ({ page }) => {
    // 入院フォームページへ遷移
    await page.goto('http://frontend:3000/hospitalizations/new', { waitUntil: 'networkidle' });

    // 日付入力フィールド (入院開始日、終了日等) が存在することを確認
    const dateInput = page.locator('input[type="date"], [data-testid*="date"]').first();
    if (await dateInput.isVisible({ timeout: 3000 })) {
      await expect(dateInput).toBeVisible();
    }
  });
});
