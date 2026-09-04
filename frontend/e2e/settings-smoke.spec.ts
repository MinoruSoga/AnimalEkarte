import { test, expect } from "@playwright/test";
import type { BrowserContext } from "@playwright/test";
import { loginAsDemoAdmin } from "./helpers/auth";

// Smoke coverage for /settings/* pages.
// Already covered by master-crud.spec.ts:
//   /settings (index navigation) → skipped
//   /settings/treatment-items (CRUD tests) → skipped
//   /settings/interview/chief-complaint → skipped
//
// Each test: navigate → assert page-specific h1 heading is visible.
// Seed prerequisites: admin@noavet.jp (system admin) has full access to all master resources.
//
// Design: each test creates a fresh page within the shared context.
// Sharing one page across 21 tests causes Chromium to accumulate state and
// become unresponsive to Playwright protocol messages on lazy-chunk navigations.
// Fresh page per test avoids state bleed while the shared context preserves the auth cookie.

// level: 1 restricts to h1 only; avoids strict-mode violations when h2 has the same text.
type SettingsPage = {
  path: string;
  heading: string;
  level?: number;
  expectTimeout?: number;
  testTimeout?: number;
};

const SETTINGS_PAGES: SettingsPage[] = [
  { path: "/settings/staff", heading: "スタッフマスタ" },
  { path: "/settings/diagnosis", heading: "診断マスタ" },
  { path: "/settings/animal-species", heading: "動物種類マスタ" },
  { path: "/settings/trimming", heading: "トリミングマスタ" },
  { path: "/settings/trimming-course-type", heading: "コース種別マスタ" },
  { path: "/settings/medicine", heading: "薬剤マスタ" },
  { path: "/settings/reservation-type", heading: "予約区分マスタ" },
  { path: "/settings/hospitalization", heading: "入院マスタ" },
  { path: "/settings/cage", heading: "ケージマスタ" },
  { path: "/settings/merchandise-items", heading: "商品マスタ" },
  { path: "/settings/insurance", heading: "保険マスタ" },
  { path: "/settings/occupations", heading: "職種マスタ" },
  { path: "/settings/permission-groups", heading: "権限グループマスタ" },
  { path: "/settings/inquiry-templates", heading: "問診テンプレートマスタ" },
  { path: "/settings/shift-templates", heading: "シフトテンプレートマスタ" },
  // closing-settings is a separate Vite lazy chunk; Chromium may pause while it compiles.
  // A higher test timeout + fresh page gives it room to load without hanging the runner.
  {
    path: "/settings/closing-time",
    heading: "締め時間設定",
    expectTimeout: 45000,
    testTimeout: 120000,
  },
  { path: "/settings/payment-methods", heading: "支払方法マスタ" },
  { path: "/settings/campaigns", heading: "割引キャンペーンマスタ" },
  // LstepSettingsPage has both h1 and h2 with the same title; use level:1 to avoid strict-mode violation.
  { path: "/settings/integrations/lstep", heading: "Lステップ連携設定", level: 1 },
  { path: "/settings/lstep/tags", heading: "Lステップタグ管理" },
  { path: "/settings/clinic", heading: "医院マスタ" },
];

test.describe("設定ページ smoke E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    // Log in once; the session cookie stays in context and is shared by all new pages.
    const loginPage = await context.newPage();
    await loginAsDemoAdmin(loginPage);
    await loginPage.close();
  });

  test.afterAll(async () => {
    await context.close();
  });

  for (const { path, heading, level, expectTimeout = 15000, testTimeout } of SETTINGS_PAGES) {
    test(`${path} — "${heading}" が表示される`, async () => {
      if (testTimeout) test.setTimeout(testTimeout);
      const page = await context.newPage();
      try {
        await page.goto(path, { waitUntil: "domcontentloaded" });
        const locator =
          level != null
            ? page.getByRole("heading", { name: heading, level })
            : page.getByRole("heading", { name: heading });
        await expect(locator).toBeVisible({ timeout: expectTimeout });
      } finally {
        await page.close();
      }
    });
  }
});
