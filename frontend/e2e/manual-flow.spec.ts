import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { createAuthedContext } from "./helpers/context";
import { ManualPage } from "./pages/manual-page";

// E2E flow tests for manual pages:
// /manual (index with sidebar navigation)
// /manual/:category/:slug (dynamic article pages)
// Covers: page load, sidebar category navigation, article display.
// Seed data: admin@noavet.jp is system_admin with full access.

test.describe("マニュアル フロー E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("/manual — マニュアルページが表示される", async () => {
    const page = await context.newPage();
    const manual = new ManualPage(page);
    try {
      await manual.gotoIndex();
      await expect(page).toHaveURL(/\/manual\/screens\/00-overview/);
      await expect(manual.overviewHeading()).toBeVisible({
        timeout: 15000,
      });
    } finally {
      await page.close();
    }
  });

  test("/manual — サイドバーが表示される", async () => {
    const page = await context.newPage();
    const manual = new ManualPage(page);
    try {
      await manual.gotoIndex();
      const sidebar = manual.sidebar();
      await expect(sidebar).toBeVisible({ timeout: 15000 });

      const navLinks = manual.sidebarLinks();
      const linkCount = await navLinks.count();
      expect(linkCount).toBeGreaterThanOrEqual(1);
    } finally {
      await page.close();
    }
  });

  test("/manual — サイドバーのカテゴリをクリックできる", async () => {
    const page = await context.newPage();
    const manual = new ManualPage(page);
    try {
      await manual.gotoIndex();
      await manual.categoryTab("業務フロー").click();
      await expect(manual.categoryTab("業務フロー")).toHaveAttribute("aria-selected", "true");
      await expect(manual.sidebarLink("新規飼主の登録から初診会計まで")).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });

  test("/manual/:category/:slug — マニュアル記事ページが表示される", async () => {
    const page = await context.newPage();
    const manual = new ManualPage(page);
    try {
      await manual.gotoArticle("screens/03-owners");
      await expect(page).toHaveURL(/\/manual\/screens\/03-owners/);
      await expect(manual.heading("飼主・ペット管理", 1)).toBeVisible({
        timeout: 15000,
      });
    } finally {
      await page.close();
    }
  });

  test("/manual — コンテンツが表示される", async () => {
    const page = await context.newPage();
    const manual = new ManualPage(page);
    try {
      await manual.gotoIndex();
      await expect(page).toHaveURL(/\/manual\/screens\/00-overview/);
      await expect(manual.searchInput()).toBeVisible({ timeout: 10000 });
      await manual.searchInput().fill("会計");
      await expect(manual.firstSidebarLink()).toBeVisible({
        timeout: 10000,
      });
    } finally {
      await page.close();
    }
  });

  test("/manual — ナビゲーションリンクが機能する", async () => {
    const page = await context.newPage();
    const manual = new ManualPage(page);
    try {
      await manual.gotoIndex();
      await manual.sidebarLink("会計").click();
      await expect(page).toHaveURL(/\/manual\/screens\/05-accounting/);
      await expect(manual.heading("会計", 1)).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });
});
