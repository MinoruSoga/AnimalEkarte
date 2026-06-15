import { expect, type Page } from '@playwright/test';

const DEMO_EMAIL = 'admin@noavet.jp';
const DEMO_PASSWORD = 'password';

export async function loginAsDemoAdmin(page: Page) {
  // Use domcontentloaded to avoid waiting for all Vite ES module requests (which
  // can be slow through the Docker host-gateway bridge). React renders after the
  // modules load; we wait explicitly for the email input instead.
  await page.goto('/login', { waitUntil: 'domcontentloaded' });

  // Wait for the login form to render (React suspends until initialAuthPromise
  // resolves — GET /v1/me → 401 → null, typically <1 s but can be slow in Docker).
  // Fresh browser contexts download all Vite ES modules through the Docker bridge,
  // which can take 35-50 s — 60 s covers the worst case.
  const emailInput = page.locator('#login-email');
  await emailInput.waitFor({ state: 'visible', timeout: 60000 });

  await emailInput.fill(DEMO_EMAIL);
  await page.locator('#login-password').fill(DEMO_PASSWORD);

  // waitForURL + click in parallel so we don't miss a fast navigation.
  // Use waitUntil: 'commit' so the promise resolves as soon as React Router's
  // replaceState fires — not after the home-page lazy chunks finish loading.
  // Default 'load' would wait for Vite to serve all home-page modules (30-50 s
  // through Docker), making the 30 s timeout unreliable for cold contexts.
  await Promise.all([
    page.waitForURL((url) => !url.pathname.startsWith('/login'), {
      timeout: 60000,
      waitUntil: 'commit',
    }),
    page.getByRole('button', { name: 'ログイン' }).click(),
  ]);

  // Sanity check: confirm we actually left /login.
  await expect(page).not.toHaveURL(/\/login/);
}
