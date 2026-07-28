import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { DEMO_IRIS_PET } from './helpers/demo-seed';
import { ExaminationsPage } from './pages/examinations-page';

// E2E flow tests for examinations (/examinations) pages.
// Covers: list page, pet selection, new form, detail form.
// Demo seed: Iris pet id=1000099 (not petId=1).

test.describe('検査管理 フロー E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/examinations — 検査管理一覧が表示される', async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    try {
      await examinations.gotoList();
      // Use level:1 to avoid strict locator violation with multiple headings
      await expect(examinations.listHeadingPattern()).toBeVisible({ timeout: 15000 });
      // Verify page loaded even if list is empty
      await expect(page).toHaveURL(/\/examinations/);
    } finally {
      await page.close();
    }
  });

  test('/examinations/select-pet — ペット選択画面が表示される', async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    try {
      await examinations.gotoSelectPet();
      await expect(examinations.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      await examinations.patientSearchInput().fill(DEMO_IRIS_PET.name);
      await expect(examinations.irisText()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test(`/examinations/new?petId=${DEMO_IRIS_PET.id} — 検査登録フォームが表示される`, async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    try {
      await examinations.gotoNew(`?petId=${DEMO_IRIS_PET.id}`);
      await expect(examinations.newFormHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(examinations.irisText()).toBeVisible({ timeout: 10000 });
      await expect(examinations.saveButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/examinations/:id — 検査詳細フォームが表示される', async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    try {
      // Navigate to list first to find an examination ID
      await examinations.gotoList();
      await expect(examinations.listHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(examinations.firstRow()).toBeVisible({ timeout: 15000 });

      // DataTableRow is non-interactive; open via DataTableRowLink.
      await examinations.firstDetailLink().click();

      await expect(examinations.detailHeading()).toBeVisible({ timeout: 15000 });
      await expect(page).toHaveURL(/\/examinations\/\d+/);
      await expect(examinations.saveButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/examinations — 検査一覧で検索が機能する', async () => {
    const page = await context.newPage();
    const examinations = new ExaminationsPage(page);
    try {
      await examinations.gotoList();
      await expect(examinations.listHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(examinations.firstRow()).toBeVisible({ timeout: 15000 });
      const firstTestType = (await examinations.firstRowTestTypeCell().textContent())?.trim();
      expect(firstTestType).toBeTruthy();

      await page.getByLabel('検索').click();
      const searchInput = examinations.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill(firstTestType ?? '');
      await expect(examinations.firstRow()).toContainText(firstTestType ?? '', { timeout: 10000 });
    } finally {
      await page.close();
    }
  });
});
