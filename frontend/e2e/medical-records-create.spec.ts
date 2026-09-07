import { expect, test, type BrowserContext } from "@playwright/test";

import { SYNTHETIC_CLINICAL_SCENARIOS, SYNTHETIC_PET } from "./fixtures/ui-design-clinical";
import { assertClinicalSuiteReady } from "./helpers/clinical-suite";
import { createAuthedContext } from "./helpers/context";
import { installSyntheticApiInterceptor, isBusinessNonGet } from "./helpers/synthetic-api";

const E2E_ORIGIN = new URL(process.env.PLAYWRIGHT_TEST_BASE_URL ?? "http://localhost:3003").origin;

test.describe("カルテ新規作成フォーム E2E", () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    assertClinicalSuiteReady();
    context = await createAuthedContext(browser);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test("mount時auto-createをローカルfulfillしsynthetic detailへ遷移する", async () => {
    const scenario = SYNTHETIC_CLINICAL_SCENARIOS.medicalRecordCreate;
    const page = await context.newPage();
    const interceptor = await installSyntheticApiInterceptor(page, scenario.endpoints, {
      expectedOrigin: E2E_ORIGIN,
      expectedClinicId: scenario.currentClinicId,
    });

    try {
      await page.addInitScript(({ key, value }) => window.localStorage.setItem(key, value), {
        key: "auth_current_clinic:v1",
        value: scenario.currentClinicId,
      });
      await page.clock.setFixedTime(new Date(scenario.fixedTime));
      await page.goto(scenario.entryPath, { waitUntil: "domcontentloaded" });
      await expect(page).toHaveURL(scenario.renderedPath, { timeout: 30000 });
      await expect(page.getByRole("heading", { name: "カルテ編集" })).toBeVisible();
      await expect(
        page.getByText(SYNTHETIC_PET.name, { exact: false }).filter({ visible: true }).first(),
      ).toBeVisible();
      await page.waitForTimeout(350);
      await page.close();

      const locallyFulfilledBusinessNonGet =
        interceptor.ledger.locallyFulfilled.filter(isBusinessNonGet);
      const continuedBusinessNonGet =
        interceptor.ledger.continuedToBackend.filter(isBusinessNonGet);
      expect(locallyFulfilledBusinessNonGet).toEqual(scenario.expectedLocalBusinessNonGet);
      expect(interceptor.ledger.blocked).toEqual([]);
      expect(interceptor.ledger.validationFailures).toEqual([]);
      expect(continuedBusinessNonGet).toEqual([]);
      expect(interceptor.ledger.continuedToBackend).toEqual([]);
      expect(interceptor.ledger.attempted.length).toBe(
        interceptor.ledger.locallyFulfilled.length +
          interceptor.ledger.blocked.length +
          interceptor.ledger.continuedToBackend.length,
      );
    } finally {
      if (!page.isClosed()) await page.close();
      await interceptor.dispose();
    }
  });
});
