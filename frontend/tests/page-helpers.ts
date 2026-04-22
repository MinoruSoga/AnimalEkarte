/**
 * ページナビゲーション・要素検索ヘルパー
 * E2E テストで頻繁に使う操作をユーティリティ化
 */

import { expect, type Page, type Locator } from '@playwright/test';

/**
 * 要素が表示されるのを待ち、クリック
 */
export async function clickElement(page: Page, selector: string, timeout = 5000): Promise<void> {
  const element = page.locator(selector).first();
  await expect(element).toBeVisible({ timeout });
  await element.click();
}

/**
 * テキストを含むボタンを探してクリック
 */
export async function clickButtonByText(page: Page, text: RegExp | string, timeout = 5000): Promise<void> {
  const button = page.getByRole('button', { name: text }).first();
  await expect(button).toBeVisible({ timeout });
  await button.click();
}

/**
 * リンクをテキストで探してクリック
 */
export async function clickLinkByText(page: Page, text: RegExp | string, timeout = 5000): Promise<void> {
  const link = page.getByRole('link', { name: text }).first();
  await expect(link).toBeVisible({ timeout });
  await link.click();
}

/**
 * フォームフィールドにテキストを入力
 */
export async function fillInput(
  page: Page,
  selector: string,
  value: string,
  timeout = 5000
): Promise<void> {
  const input = page.locator(selector).first();
  await expect(input).toBeVisible({ timeout });
  await input.fill(value);
}

/**
 * テーブル内の行を選択（最初の行をクリック）
 */
export async function selectTableRow(page: Page, index = 0, timeout = 5000): Promise<void> {
  const rows = page.locator('table tbody tr, [role="grid"] [role="row"]');
  await expect(rows.nth(index)).toBeVisible({ timeout });
  await rows.nth(index).click();
}

/**
 * サイドパネルが開くのを待つ
 */
export async function waitForSidepanel(page: Page, timeout = 5000): Promise<Locator> {
  const sidepanel = page.locator('aside, [data-testid="sidepanel"], [role="complementary"]').first();
  await expect(sidepanel).toBeVisible({ timeout });
  return sidepanel;
}

/**
 * ダイアログが表示されるのを待つ
 */
export async function waitForDialog(page: Page, timeout = 5000): Promise<Locator> {
  const dialog = page.locator('[role="dialog"], [data-testid="dialog"]').first();
  await expect(dialog).toBeVisible({ timeout });
  return dialog;
}

/**
 * ダイアログのボタンをクリック（OK / キャンセル など）
 */
export async function clickDialogButton(
  page: Page,
  buttonName: RegExp | string,
  timeout = 5000
): Promise<void> {
  const dialog = page.locator('[role="dialog"]').first();
  await expect(dialog).toBeVisible({ timeout });

  const button = dialog.locator(`button:has-text("${buttonName}")`).first();
  await expect(button).toBeVisible({ timeout });
  await button.click();
}

/**
 * ナビゲーションメニュー項目をクリック
 */
export async function clickNavMenu(page: Page, text: string, timeout = 5000): Promise<void> {
  const menuItem = page.locator(
    `nav a:has-text("${text}"), nav button:has-text("${text}"), [data-testid*="nav"]:has-text("${text}")`
  ).first();

  await expect(menuItem).toBeVisible({ timeout });
  await menuItem.click();
}

/**
 * ページが指定 URL に遷移するのを待つ
 */
export async function expectNavigationToUrl(
  page: Page,
  urlPattern: RegExp | string,
  timeout = 10000
): Promise<void> {
  if (typeof urlPattern === 'string') {
    await expect(page).toHaveURL(new RegExp(urlPattern), { timeout });
  } else {
    await expect(page).toHaveURL(urlPattern, { timeout });
  }
}

/**
 * ページがログインページではないことを確認（認証確認）
 */
export async function expectAuthenticatedState(page: Page, timeout = 10000): Promise<void> {
  await expect(page).not.toHaveURL(/\/login/, { timeout });
}

/**
 * 検索フィールドにテキストを入力して検索実行
 */
export async function searchByText(
  page: Page,
  searchText: string,
  searchFieldSelector = 'input[placeholder*="検索"], input[placeholder*="Search"], [data-testid*="search"]'
): Promise<void> {
  await page.locator(searchFieldSelector).first().fill(searchText);
  await page.keyboard.press('Enter');
  await page.waitForTimeout(500);
}

/**
 * テーブル内のセルテキストを取得
 */
export async function getTableCellText(
  page: Page,
  rowIndex: number,
  cellIndex: number
): Promise<string> {
  const cell = page.locator(`table tbody tr:nth-child(${rowIndex + 1}) td:nth-child(${cellIndex + 1})`);
  return (await cell.textContent()) || '';
}

/**
 * データが表示されるのを待つ（テーブルに行があることを確認）
 */
export async function waitForDataDisplay(
  page: Page,
  timeout = 10000
): Promise<void> {
  await page.waitForSelector('table tbody tr, [role="grid"] [role="row"], [data-testid*="item"]', { timeout });
}

/**
 * ローディング表示が消えるのを待つ
 */
export async function waitForLoadingToComplete(
  page: Page,
  timeout = 10000
): Promise<void> {
  const loaders = page.locator('[data-testid*="loading"], [data-testid*="spinner"], .loader, .spinner').first();

  // ローダーが表示されている場合、消えるまで待つ
  if (await loaders.isVisible({ timeout: 1000 }).catch(() => false)) {
    await expect(loaders).not.toBeVisible({ timeout });
  }
}

/**
 * トースト・通知メッセージが表示されるのを待つ
 */
export async function expectToastMessage(
  page: Page,
  messagePattern: RegExp | string,
  timeout = 5000
): Promise<void> {
  const toast = page.locator('[data-testid*="toast"], [role="status"], .toast').filter({ hasText: messagePattern });
  await expect(toast.first()).toBeVisible({ timeout });
}

/**
 * フォームのエラーメッセージを確認
 */
export async function expectFormError(
  page: Page,
  errorPattern: RegExp | string,
  timeout = 5000
): Promise<void> {
  const error = page.locator('[role="alert"], [data-testid*="error"], .error-message').filter({
    hasText: errorPattern,
  });
  await expect(error.first()).toBeVisible({ timeout });
}

/**
 * チェックボックスを設定 (checked / unchecked)
 */
export async function setCheckboxState(
  page: Page,
  selector: string,
  checked: boolean,
  timeout = 5000
): Promise<void> {
  const checkbox = page.locator(selector).first();
  await expect(checkbox).toBeVisible({ timeout });

  const isChecked = await checkbox.isChecked();
  if (isChecked !== checked) {
    await checkbox.click();
  }
}

/**
 * セレクト要素から選択肢を選択
 */
export async function selectOption(
  page: Page,
  selectSelector: string,
  optionValue: string,
  timeout = 5000
): Promise<void> {
  const select = page.locator(selectSelector).first();
  await expect(select).toBeVisible({ timeout });
  await select.selectOption(optionValue);
}

/**
 * ページ内のすべてのテキストを取得 (デバッグ用)
 */
export async function getAllPageText(page: Page): Promise<string> {
  return (await page.locator('body').textContent()) || '';
}

/**
 * コンソールエラーをリスニング（テストコンテキスト内のエラー検出）
 */
export function listenForConsoleErrors(page: Page): Promise<string[]> {
  const errors: string[] = [];

  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      errors.push(msg.text());
    }
  });

  return Promise.resolve(errors);
}
