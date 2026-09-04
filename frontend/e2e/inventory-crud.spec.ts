import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { InventoryPage } from "./pages/inventory-page";

// E2E tests for inventory (/inventory) pages.
// Tests new item creation via /inventory/new form.
// Note: there is no delete button in the inventory UI; cleanup is not done.
// Unique names (timestamp suffix) minimize risk of collisions across test runs.
//
// Test order is intentional: fast read/filter tests run first so the slow
// form-submission test (which can take up to 60s) doesn't starve them.

test.describe("在庫管理 E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/inventory — 在庫管理一覧が表示される", async () => {
    const page = await context.newPage();
    const inventory = new InventoryPage(page);
    try {
      await inventory.gotoList();
      await expect(inventory.listHeading()).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test("/inventory/new — 在庫登録フォームが表示される", async () => {
    const page = await context.newPage();
    const inventory = new InventoryPage(page);
    try {
      await inventory.gotoNew();
      await expect(inventory.newFormHeading()).toBeVisible({ timeout: 15000 });
      await expect(inventory.nameInput()).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test("/inventory — 検索フィルタが機能する", async () => {
    const page = await context.newPage();
    const inventory = new InventoryPage(page);
    try {
      await inventory.gotoList();
      await expect(inventory.listHeading()).toBeVisible();

      // PropertyFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel("検索").click();
      const searchInput = inventory.searchInput();
      await expect(searchInput).toBeVisible();
      // 存在しない品名で検索 — 入力が受け付けられることを確認
      await searchInput.fill("存在しない品名_XXXXXXXXXXXXXXXX");
      await expect(searchInput).toHaveValue("存在しない品名_XXXXXXXXXXXXXXXX");
      // 検索クリア
      await searchInput.clear();
      await expect(searchInput).toHaveValue("");
    } finally {
      await page.close();
    }
  });

  test("/inventory — 在庫一覧から編集画面に遷移する", async () => {
    const page = await context.newPage();
    const inventory = new InventoryPage(page);
    try {
      await inventory.gotoList();
      await expect(inventory.listHeading()).toBeVisible();
      // 一覧に1件以上あれば行をクリックして在庫編集画面に遷移
      await expect(inventory.firstRow()).toBeVisible({ timeout: 15000 });
      await inventory.firstRow().click();
      await expect(inventory.editHeading()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/inventory\/\d+/);
    } finally {
      await page.close();
    }
  });

  test("/inventory/new — フォームに入力して登録すると一覧に表示される", async () => {
    test.setTimeout(60000);
    const page = await context.newPage();
    const inventory = new InventoryPage(page);
    const itemName = `E2E在庫_${Date.now()}`;
    try {
      await inventory.gotoNew();
      await expect(inventory.newFormHeading()).toBeVisible({ timeout: 15000 });

      // 品名を入力
      await inventory.nameInput().fill(itemName);
      // 単位を入力
      await inventory.unitInput().fill("個");
      // 数量を入力
      await inventory.quantityInput().fill("10");

      // 「登録」ボタンをクリック (form 属性でフォームに紐付け; 複数該当を避けるため属性セレクタ使用)
      await inventory.submitButton().click();

      // 成功後に在庫一覧にリダイレクトされる
      await expect(page).toHaveURL("/inventory", { timeout: 15000 });
      await expect(inventory.listHeading()).toBeVisible();

      // 登録した品名が一覧に表示される
      // PropertyFilter: 検索トグルボタンをクリックして入力欄を表示
      await page.getByLabel("検索").click();
      const searchInput = inventory.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill(itemName);
      await expect(inventory.itemText(itemName)).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
