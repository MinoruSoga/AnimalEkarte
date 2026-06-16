import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// CRUD flow tests for master settings pages.
// Each test creates test data with a timestamp prefix for isolation.
// Where possible, the test also deletes the created data to keep the DB clean.
//
// Design: fresh page per test within shared context to avoid Chromium
// state accumulation across many navigations.

test.describe('設定マスタ CRUD E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    const loginPage = await context.newPage();
    await loginAsDemoAdmin(loginPage);
    await loginPage.close();
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('動物種類マスタ: 新規作成 → 一覧表示確認 → 削除', async () => {
    test.setTimeout(60000);
    const page = await context.newPage();
    const speciesName = `E2E種_${Date.now()}`;
    try {
      await page.goto('/settings/animal-species', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '動物種類マスタ' })).toBeVisible();

      // 新規作成パネルを開く
      await page.getByRole('button', { name: '新規登録' }).click();
      await expect(page.locator('#master-title')).toBeVisible();
      await page.locator('#master-title').fill(speciesName);
      await page.getByRole('button', { name: '保存' }).click();

      // 保存成功後: パネルが閉じて一覧に表示される
      await expect(page.locator('#master-title')).not.toBeVisible({ timeout: 10000 });
      await expect(page.getByText(speciesName)).toBeVisible({ timeout: 10000 });

      // ページをリロードしてクリーンな状態にしてから削除操作
      // (React state が残っている場合のパネル再オープン問題を回避)
      await page.goto('/settings/animal-species', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '動物種類マスタ' })).toBeVisible();

      // NotionFilter: 検索トグルを開いてから入力
      await page.getByLabel('検索').click();
      const searchInput = page.getByPlaceholder('動物種類名で検索...');
      await expect(searchInput).toBeVisible();
      await searchInput.fill(speciesName);
      await expect(page.getByText(speciesName)).toBeVisible({ timeout: 10000 });

      // 行の「操作」ボタンをクリックしてパネルを開く
      await page.locator('tbody tr').filter({ hasText: speciesName }).getByLabel('操作').click();
      await expect(page.locator('#master-title')).toBeVisible({ timeout: 10000 });

      // 削除ボタン (toolbar の aria-label="削除") をクリック
      await page.getByLabel('削除').click();

      // AlertDialog の「削除」ボタンをクリック
      const dialog = page.getByRole('alertdialog');
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
    try {
      await page.goto('/settings/animal-species', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '動物種類マスタ' })).toBeVisible();

      // NotionFilter: 検索トグルボタンをクリックして入力欄を表示
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
    const medicineName = `E2E薬_${Date.now()}`;
    try {
      await page.goto('/settings/medicine', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '薬剤マスタ' })).toBeVisible();

      // 新規登録ボタンをクリック
      await page.getByRole('button', { name: '新規登録' }).click();
      await expect(page.locator('#master-title')).toBeVisible({ timeout: 10000 });

      // 品名を入力して保存
      await page.locator('#master-title').fill(medicineName);
      await page.getByRole('button', { name: '保存' }).click();

      // 保存成功後パネルが閉じる
      await expect(page.locator('#master-title')).not.toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('診断病名マスタ: 新規登録パネルが開く', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/settings/diagnosis', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '診断病名マスタ' })).toBeVisible();

      // 「新規登録」ボタンをクリック → パネルが開く
      await page.getByRole('button', { name: '新規登録' }).click();
      await expect(page.locator('#master-title')).toBeVisible({ timeout: 10000 });

      // キャンセルボタンで閉じる
      await page.getByRole('button', { name: 'キャンセル' }).click();
      await expect(page.locator('#master-title')).not.toBeVisible({ timeout: 5000 });
    } finally {
      await page.close();
    }
  });
});
