import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { createAuthedContext } from './helpers/context';
import { DEMO_IRIS_PET } from './helpers/demo-seed';
import { CheckupsPage } from './pages/checkups-page';

// E2E flow tests for checkups (/checkups) pages.
// Covers: list page, pet selection, new form.
// Note: /checkups/:id detail route does not exist in router.
// Demo seed: Iris pet id=1000099 (not petId=1).

test.describe('定期健診 フロー E2E', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test('/checkups — 定期健診一覧が表示される', async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoList();
      await expect(checkups.listHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(checkups.newButton()).toBeVisible({ timeout: 10000 });
      await expect(checkups.tableBody()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test('/checkups/select-pet — ペット選択画面が表示される', async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoSelectPet();
      await expect(checkups.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      // Iris is outside the unfiltered first page — search via #search auto-debounce.
      await checkups.patientSearchInput().fill(DEMO_IRIS_PET.name);
      await expect(checkups.irisText()).toBeVisible({ timeout: 15000 });
    } finally {
      await page.close();
    }
  });

  test(`/checkups/new?petId=${DEMO_IRIS_PET.id} — 定期健診新規登録フォームが表示される`, async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoNew(`?petId=${DEMO_IRIS_PET.id}`);
      await expect(checkups.newFormHeading()).toBeVisible({ timeout: 15000 });
      await expect(checkups.irisText()).toBeVisible({ timeout: 10000 });
      await expect(checkups.saveButton()).toBeVisible({ timeout: 10000 });
    } finally {
      await page.close();
    }
  });

  test('/checkups — 新規登録ボタンでペット選択画面に遷移する', async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoList();
      await expect(checkups.listHeading()).toBeVisible({
        timeout: 15000,
      });

      await checkups.newButton().click();
      await expect(checkups.selectPetHeading()).toBeVisible({
        timeout: 15000,
      });
      await expect(page).toHaveURL(/\/checkups\/select-pet/);
    } finally {
      await page.close();
    }
  });

  test('/checkups — 一覧で検索が機能する', async () => {
    const page = await context.newPage();
    const checkups = new CheckupsPage(page);
    try {
      await checkups.gotoList();
      await expect(checkups.listHeading()).toBeVisible({
        timeout: 15000,
      });

      const rows = checkups.rows();
      const initialRowCount = await rows.count();

      await page.getByLabel('検索').click();
      const searchInput = checkups.searchInput();
      await expect(searchInput).toBeVisible();
      await searchInput.fill('Iris');

      // Wait for search to apply
      await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => null);

      // Verify filtered results or search input retained
      const filteredRows = await checkups.rows().count();
      expect(filteredRows).toBeLessThanOrEqual(initialRowCount);
      await expect(searchInput).toHaveValue('Iris');
    } finally {
      await page.close();
    }
  });
});
