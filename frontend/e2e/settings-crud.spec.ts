import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { SettingsMasterPage } from './pages/settings-master-page';

// CRUD flow tests for master settings pages.
// Each test creates test data with a timestamp prefix for isolation.
// Where possible, the test also deletes the created data to keep the DB clean.
//
// Design: fresh page per test within shared context to avoid Chromium
// state accumulation across many navigations.

test.describe('設定マスタ CRUD E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('動物種類マスタ: 新規作成 → 一覧表示確認 → 削除', async () => {
    test.setTimeout(60000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const speciesName = `E2E種_${Date.now()}`;
    try {
      await settings.open('/settings/animal-species');
      await expect(settings.heading('動物種類マスタ')).toBeVisible();

      // 新規作成パネルを開く
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible();
      await settings.masterTitleInput().fill(speciesName);
      await settings.saveButton().click();

      // 保存成功後: パネルが閉じて一覧に表示される
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
      await expect(page.getByText(speciesName)).toBeVisible({ timeout: 10000 });

      // ページをリロードしてクリーンな状態にしてから削除操作
      // (React state が残っている場合のパネル再オープン問題を回避)
      await settings.open('/settings/animal-species');
      await expect(settings.heading('動物種類マスタ')).toBeVisible();

      // PropertyFilter: 検索トグルを開いてから入力
      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('動物種類名で検索...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill(speciesName);
      await expect(page.getByText(speciesName)).toBeVisible({ timeout: 10000 });

      // 行の「操作」ボタンをクリックしてパネルを開く
      await settings.rowActionButton(speciesName).click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });

      // 削除ボタン (toolbar の aria-label="削除") をクリック
      await settings.deleteButton().click();

      // AlertDialog の「削除」ボタンをクリック
      const dialog = settings.deleteDialog();
      await expect(dialog).toBeVisible();
      await dialog.getByRole('button', { name: '削除' }).click();

      // 削除後に一覧から消える
      await expect(page.getByText(speciesName)).not.toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('動物種類マスタ: 検索フィルタが機能する', async () => {
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    try {
      await settings.open('/settings/animal-species');
      await expect(settings.heading('動物種類マスタ')).toBeVisible();

      // PropertyFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('動物種類名で検索...');
      await expect(searchInput).toBeVisible();
      // 検索で絞り込み: seed にある「犬」に含まれる文字で検索
      await searchInput.fill('犬');
      await expect(searchInput).toHaveValue('犬');
      await expect(page.getByText('犬').first()).toBeVisible({ timeout: 10000 });
      // 検索クリア
      await searchInput.clear();
      await expect(searchInput).toHaveValue('');
    } finally {
      await page.close();
    }
  });

  test('薬剤マスタ: 新規作成パネルが開き保存できる', async () => {
    test.setTimeout(60000);
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    const medicineName = `E2E薬_${Date.now()}`;
    try {
      await settings.open('/settings/medicine');
      await expect(settings.heading('薬剤マスタ')).toBeVisible();

      // 新規登録ボタンをクリック
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });

      // 品名を入力して保存
      await settings.masterTitleInput().fill(medicineName);
      await settings.saveButton().click();

      // 保存成功後パネルが閉じる
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('診断マスタ: 新規登録パネルが開く', async () => {
    const page = await context.newPage();
    const settings = new SettingsMasterPage(page);
    try {
      await settings.open('/settings/diagnosis');
      await expect(settings.heading('診断マスタ')).toBeVisible();

      // 「新規登録」ボタンをクリック → パネルが開く
      await settings.newButton().click();
      await expect(settings.masterTitleInput()).toBeVisible({ timeout: 10000 });

      // キャンセルボタンで閉じる
      await settings.cancelButton().click();
      await expect(settings.masterTitleInput()).not.toBeVisible({ timeout: 5000 });
    } finally {
      await page.close();
    }
  });
});
