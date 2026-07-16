import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// E2E flow tests for LINE reservation pages:
// /line-reservation, /line-reservation/page-editor, /line-reservation/slots
// Covers: page load, basic navigation, form/editor interaction.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe('LINE予約 フロー E2E', () => {
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

  test('/line-reservation — LINE予約基本設定ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/line-reservation', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '基本設定', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText('稼働状態')).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: '設定を保存' })).toBeVisible({ timeout: 10000 });
      await expect(page).toHaveURL(/\/line-reservation/);
    } finally {
      await page.close();
    }
  });

  test('/line-reservation — ページタブが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/line-reservation', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: '基本設定', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText('営業時間・定休日')).toBeVisible({ timeout: 10000 });
      await expect(page.getByText('予約枠設定')).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/line-reservation/page-editor — ページエディタが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/line-reservation/page-editor', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'ページ編集', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByLabel('ヘッダーテキスト')).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: '変更を保存' })).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/line-reservation/page-editor — ヘッダーテキスト入力フィールドが機能する', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/line-reservation/page-editor', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'ページ編集', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      const headerTextarea = page.getByLabel('ヘッダーテキスト');
      await expect(headerTextarea).toBeVisible({ timeout: 10000 });
      await headerTextarea.fill('E2E ヘッダーテキスト');
      await expect(headerTextarea).toHaveValue('E2E ヘッダーテキスト');
    } finally {
      await page.close();
    }
  });

  test('/line-reservation/slots — LINE予約枠設定ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/line-reservation/slots', { waitUntil: 'domcontentloaded' });
      // Use level:1 and more specific pattern to avoid strict locator
      await expect(page.getByRole('heading', { name: 'LINE予約枠', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      await expect(page.getByText('LINE予約で選択できる開始時刻を日別に設定します')).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });

  test('/line-reservation/slots — ページが表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/line-reservation/slots', { waitUntil: 'domcontentloaded' });
      await expect(page).toHaveURL(/\/line-reservation\/slots/);
      // SD-7 (348cae41) で加算モード仕様の文言に更新済み
      await expect(page.getByText('営業時間から自動生成される予約可能枠に追加されます')).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/line-reservation/slots — 予約枠の説明が表示される', async () => {
    const page = await context.newPage();
    try {
      await page.goto('/line-reservation/slots', { waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('heading', { name: 'LINE予約枠', level: 1 })).toBeVisible({
        timeout: 15000,
      });
      // SD-7 (348cae41) で加算モード仕様の文言に更新済み
      await expect(page.getByText('枠が未登録の日も営業時間から自動生成されます')).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });
});
