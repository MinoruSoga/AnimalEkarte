import { type Browser, type BrowserContext } from "@playwright/test";

import {
  assertClinicalAppEnv,
  assertClinicalBaseURL,
  assertClinicalTeardownRegistered,
} from "./clinical-env";
import { requireClinicalFixture, type ClinicalFixture } from "./clinical-fixture";
import { createAuthedContext } from "./context";

export function assertClinicalSuiteReady(): ClinicalFixture {
  assertClinicalAppEnv(process.env.APP_ENV);
  assertClinicalBaseURL(process.env.PLAYWRIGHT_TEST_BASE_URL);
  assertClinicalTeardownRegistered(process.env.E2E_CLINICAL_TEARDOWN);
  return requireClinicalFixture();
}

export async function createClinicalContext(browser: Browser): Promise<{
  context: BrowserContext;
  fixture: ClinicalFixture;
}> {
  const fixture = assertClinicalSuiteReady();
  const context = await createAuthedContext(browser);
  return { context, fixture };
}
