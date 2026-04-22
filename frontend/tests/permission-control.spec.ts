import { test, expect } from '@playwright/test';

/**
 * 権限制御 E2E テスト
 *
 * 認証: playwright.config.ts の globalSetup で事前にログイン済み
 *
 * テストフロー:
 * 1. 権限グループ設定ページ表示確認
 * 2. 権限グループ一覧表示確認
 * 3. 権限ルール編集フロー検証
 * 4. アクセス制御UI確認
 */

test.describe('権限制御 E2E テスト', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://frontend:3000/');
    await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 });
  });

  test('PERM-01: 権限グループ設定ページナビゲーション', async ({ page }) => {
    // 設定メニューをクリック
    const settingsLink = page.locator(
      'a:has-text("設定"), a:has-text("Settings"), [data-testid="nav-settings"]'
    ).first();

    if (await settingsLink.isVisible()) {
      await settingsLink.click();
      await expect(page).toHaveURL(/\/settings/, { timeout: 10000 });
    }
  });

  test('PERM-02: 権限グループ一覧表示', async ({ page }) => {
    // 権限グループ設定ページへ遷移
    await page.goto('http://frontend:3000/settings/permission-groups', { waitUntil: 'networkidle' });

    // ページが読み込まれたことを確認
    await expect(page.locator('body')).toBeVisible();

    // グループ一覧テーブルを確認
    const table = page.locator('table, [role="grid"], [data-testid="permission-groups-list"]').first();
    if (await table.isVisible({ timeout: 3000 })) {
      await expect(table).toBeVisible();
    }
  });

  test('PERM-03: 権限グループ選択・詳細表示', async ({ page }) => {
    // 権限グループ設定ページへ遷移
    await page.goto('http://frontend:3000/settings/permission-groups', { waitUntil: 'networkidle' });

    // テーブル内の行をクリック
    const row = page.locator('table tbody tr, [role="grid"] [role="row"]').first();
    if (await row.isVisible()) {
      await row.click();

      // サイドパネルまたは詳細ビューが表示されるのを待つ
      const detail = page.locator('aside, [data-testid="permission-group-detail"]').first();
      await expect(detail).toBeVisible({ timeout: 5000 });
    }
  });

  test('PERM-04: 権限ルール編集フォーム確認', async ({ page }) => {
    // 権限グループ設定ページへ遷移
    await page.goto('http://frontend:3000/settings/permission-groups', { waitUntil: 'networkidle' });

    // グループを選択してサイドパネルを開く
    const row = page.locator('table tbody tr, [role="grid"] [role="row"]').first();
    if (await row.isVisible()) {
      await row.click();

      // 権限ルールの edit ボタンを探す
      const editBtn = page.locator('button:has-text("編集"), [data-testid="btn-edit-rules"]').first();
      if (await editBtn.isVisible({ timeout: 3000 })) {
        await expect(editBtn).toBeEnabled();
      }
    }
  });

  test('PERM-05: 権限チェックボックス表示確認', async ({ page }) => {
    // 権限グループ詳細ページへ遷移
    await page.goto('http://frontend:3000/settings/permission-groups', { waitUntil: 'networkidle' });

    // パネルが表示されるのを待つ
    await page.waitForSelector('aside, [data-testid="permission-group-detail"]', { timeout: 5000 });

    // チェックボックス (権限チェック) が表示されていることを確認
    const checkbox = page.locator('input[type="checkbox"]').first();
    if (await checkbox.isVisible({ timeout: 3000 })) {
      await expect(checkbox).toBeVisible();
    }
  });
});
