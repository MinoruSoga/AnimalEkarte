import { readFile } from 'node:fs/promises';
import { expect, type BrowserContext, type Page } from '@playwright/test';

// SEC-CS2-F01: credentials must be injected via env. No repository-known
// password fallback — privileged demo secrets must not live in source.
const AUTH_STATE_PATH = process.env.E2E_AUTH_STATE_PATH ?? '/tmp/animal-ekarte-demo-admin-storage-state.json';

function requireE2ECredentials(): { email: string; password: string } {
  const email = process.env.E2E_LOGIN_EMAIL?.trim() ?? '';
  const password = process.env.E2E_LOGIN_PASSWORD ?? '';
  if (!email || !password) {
    throw new Error(
      'E2E_LOGIN_EMAIL and E2E_LOGIN_PASSWORD must be set (no in-repo credential fallback; see frontend/e2e/README.md)',
    );
  }
  return { email, password };
}

type StorageState = Awaited<ReturnType<BrowserContext['storageState']>>;

async function waitForAuthenticatedHome(page: Page) {
  await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 60000 });
  await expect(page).not.toHaveURL(/\/login/);
  await expect(page.getByRole('heading', { name: '当日の受付' })).toBeVisible({ timeout: 60000 });
}

async function restoreCachedAuth(page: Page): Promise<boolean> {
  try {
    const rawStorageState = await readFile(AUTH_STATE_PATH, 'utf8');
    const storageState = JSON.parse(rawStorageState) as StorageState;
    if (storageState.cookies.length === 0) return false;

    await page.context().addCookies(storageState.cookies);
    await waitForAuthenticatedHome(page);
    return true;
  } catch {
    return false;
  }
}

export async function loginAsDemoAdmin(page: Page) {
  if (await restoreCachedAuth(page)) return;

  const { email, password } = requireE2ECredentials();

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

  await emailInput.fill(email);
  await page.locator('#login-password').fill(password);

  const loginResponsePromise = page.waitForResponse(
    (response) => response.url().includes('/v1/login') && response.request().method() === 'POST',
    { timeout: 60000 },
  );
  await page.getByRole('button', { name: 'ログイン' }).click();
  const loginResponse = await loginResponsePromise;
  expect(loginResponse.status()).toBe(200);

  // The app uses httpOnly cookies. After the login response is committed,
  // navigate explicitly so a fresh document restores auth from the cookie.
  // This avoids depending on a React Router navigation event, which can be
  // missed late in the full E2E suite even though /v1/login succeeded.
  await waitForAuthenticatedHome(page);
  await page.context().storageState({ path: AUTH_STATE_PATH });
}
