import { expect, test } from "@playwright/test";
import { loginAsDemoAdmin } from "./helpers/auth";

// UAT re-verification for merged fixes (2026-09-24). Covers Plane items in the
// UAT state: EMR-64, EMR-76, EMR-205, EMR-206, EMR-207, EMR-208, EMR-209.
// Requires the docker compose stack (localhost:3003/8080); the E2E account is
// bound to clinic 2.
const CLINIC_ID = 2;
const RESERVATION_TYPE_ID = 2;
const DOCTOR_ID = 20000003;
const MEDICAL_RECORD_ID = 11000001;
const CLINIC_HEADERS = {
  "X-Clinic-ID": String(CLINIC_ID),
  "X-Requested-With": "XMLHttpRequest",
};

// Tests are independent; do not use serial mode so one failure does not skip the rest.
test.describe("UAT residuals 2026-09-24", () => {
  test.beforeEach(async ({ page }) => {
    await loginAsDemoAdmin(page);
  });

  test("EMR-76: overlapping reservation returns 409 not 500", async ({ page }) => {
    // Unique slot per run so leftover fixtures never collide with the first create.
    const base = `2031-${String(1 + (Math.floor(Date.now() / 86400000) % 12)).padStart(2, "0")}-${String(1 + (Math.floor(Date.now() / 3600000) % 28)).padStart(2, "0")}`;
    const body = {
      start_time: `${base}T10:00:00+09:00`,
      end_time: `${base}T11:00:00+09:00`,
      reservation_type_id: RESERVATION_TYPE_ID,
      doctor_id: DOCTOR_ID,
      is_designated: true,
      notes: "uat-residual-emr76",
    };
    const first = await page.request.post(`/api/v1/clinics/${CLINIC_ID}/reservations`, {
      data: body,
      headers: CLINIC_HEADERS,
    });
    expect([200, 201], `first create: ${first.status()} ${await first.text()}`).toContain(
      first.status(),
    );
    const created = (await first.json()) as { id?: number };
    expect(created.id).toBeTruthy();

    const overlap = await page.request.post(`/api/v1/clinics/${CLINIC_ID}/reservations`, {
      data: {
        ...body,
        start_time: `${base}T10:30:00+09:00`,
        end_time: `${base}T11:30:00+09:00`,
      },
      headers: CLINIC_HEADERS,
    });
    const status = overlap.status();
    const text = await overlap.text();
    expect(status, `expected 409 for overlapping reservation, got ${status}: ${text}`).toBe(409);

    const del = await page.request.delete(
      `/api/v1/clinics/${CLINIC_ID}/reservations/${created.id}`,
      { headers: CLINIC_HEADERS },
    );
    expect([200, 204]).toContain(del.status());
  });

  test("EMR-205: print portal pages render content under print media", async ({ page }) => {
    // 2026-07-25 has 72 completed billings on scheduled_date; reports returns
    // data even for empty months so its print portal always mounts.
    // Examination print needs a saved revision (none exist in this DB) — out of scope.
    const targets: { url: string; selector: string }[] = [
      {
        url: "/accounting?tab=daily&daily_date=2026-07-25",
        selector: '[data-testid="daily-print-area"]',
      },
      { url: "/accounting/reports", selector: "[data-print-portal]" },
    ];
    for (const { url, selector } of targets) {
      await page.goto(url, { waitUntil: "domcontentloaded" });
      await page.waitForLoadState("networkidle", { timeout: 60000 }).catch(() => undefined);
      const portal = page.locator(selector).first();
      await expect(portal, `${url}: ${selector} missing`).toBeAttached({ timeout: 30000 });
      await page.emulateMedia({ media: "print" });
      const state = await portal.evaluate((el) => ({
        display: getComputedStyle(el).display,
        textLength: (el.textContent ?? "").trim().length,
      }));
      expect(state.display, `${url}: ${selector} display under print`).not.toBe("none");
      expect(state.textLength, `${url}: ${selector} is blank`).toBeGreaterThan(0);
      await page.emulateMedia({ media: "screen" });
    }
  });

  test("EMR-207: shared Button shows visible indicator on keyboard focus", async ({ page }) => {
    await page.goto("/", { waitUntil: "domcontentloaded" });
    await page.getByRole("heading", { name: "当日の受付" }).waitFor({ timeout: 60000 });

    // Tab until a <button> is focused (shared Button uses buttonVariants classes).
    let found = false;
    for (let i = 0; i < 80 && !found; i += 1) {
      await page.keyboard.press("Tab");
      found = await page.evaluate(() => {
        const el = document.activeElement;
        if (!el || el.tagName !== "BUTTON") return false;
        const style = getComputedStyle(el);
        const visible =
          (style.boxShadow !== "none" && style.boxShadow !== "") ||
          (style.outlineStyle !== "none" && style.outlineWidth !== "0px");
        return visible;
      });
    }
    expect(found, "no focused <button> ever showed a visible indicator").toBe(true);
  });

  test("EMR-64: dialog close restores focus to the trigger", async ({ page }) => {
    await page.goto(`/medical-records/${MEDICAL_RECORD_ID}`, {
      waitUntil: "domcontentloaded",
    });
    // The 治療 tab mounts the treatment section with the dialog trigger.
    const tab = page.getByRole("tab", { name: "治療", exact: true });
    await expect(tab).toBeVisible({ timeout: 60000 });
    await tab.click();
    const trigger = page.getByRole("button", { name: "マスタから追加" }).first();
    await expect(trigger).toBeVisible({ timeout: 30000 });
    await trigger.focus();
    await trigger.click();
    const dialog = page.getByRole("dialog").first();
    await expect(dialog).toBeVisible({ timeout: 30000 });
    await page.keyboard.press("Escape");
    await expect(dialog).not.toBeVisible({ timeout: 15000 });
    const active = await page.evaluate(() => {
      const el = document.activeElement;
      return {
        tag: el?.tagName,
        text: el?.textContent?.trim() ?? "",
        isBody: el === document.body,
      };
    });
    expect(active.isBody, `focus fell to body after Escape (active=${active.tag})`).toBe(false);
    expect(active.text, `focus did not return to trigger (active=${active.text})`).toContain(
      "マスタから追加",
    );
  });

  async function openReservationTypePanel(page: import("@playwright/test").Page) {
    await page.goto("/settings/reservation-type", { waitUntil: "domcontentloaded" });
    const opener = page.getByRole("button", { name: /詳細: 予約区分 トリミング/ }).first();
    await expect(opener, "reservation-type row opener").toBeVisible({ timeout: 60000 });
    await opener.click();
  }

  test("EMR-208: available-slot add issues POST and persists", async ({ page }) => {
    await openReservationTypePanel(page);

    // The add control lives inside the 520px side-peek panel (shadow-panel class).
    // Two type=button "追加" exist (予約可能枠 first, 予約不可時間 second) plus a
    // type=submit panel-save "追加" — the available-slots one is the first match.
    const panel = page.locator("div.shadow-panel").first();
    const addButton = panel.locator('button[type="button"]', { hasText: "追加" }).first();
    await expect(addButton).toBeVisible({ timeout: 30000 });

    // Vary the start time per run so a previously persisted slot cannot collide (409).
    const timeSelect = addButton.locator("xpath=..").getByRole("combobox");
    await timeSelect.click();
    const options = page.getByRole("option");
    const optionCount = await options.count();
    await options.nth(Math.floor(Date.now() / 60000) % optionCount).click();

    const postDone = page.waitForResponse(
      (res) =>
        res.url().includes("/api/v1/masters/reservation-types/") &&
        res.url().includes("available-slots") &&
        res.request().method() === "POST",
      { timeout: 30000 },
    );
    await addButton.click();
    const res = await postDone;
    expect(
      [200, 201],
      `available-slots POST status ${res.status()}: ${await res.text()}`,
    ).toContain(res.status());
    await expect(page.getByText(/毎週月曜日|月曜/).first()).toBeVisible({ timeout: 15000 });

    // Cleanup: remove the slot created by this run.
    const created = (await res.json()) as { id?: number };
    if (created.id) {
      await page.request.delete(
        `/api/v1/masters/reservation-types/${RESERVATION_TYPE_ID}/available-slots/${created.id}`,
        { headers: CLINIC_HEADERS },
      );
    }
  });

  test("EMR-209: linking an occupation renders its badge", async ({ page }) => {
    // Seed one occupation (unique per run; table had none for this clinic).
    const occName = `UAT職種209-${Date.now() % 1000000}`;
    const create = await page.request.post("/api/v1/masters/occupations", {
      data: { name: occName, description: "uat residual check" },
      headers: CLINIC_HEADERS,
    });
    expect([200, 201], `occupation create: ${create.status()} ${await create.text()}`).toContain(
      create.status(),
    );
    const occ = (await create.json()) as { id?: number };
    expect(occ.id).toBeTruthy();

    // Verify the fixed envelope contract directly.
    const list = await page.request.get(
      `/api/v1/masters/reservation-types/${RESERVATION_TYPE_ID}/occupations`,
      { headers: CLINIC_HEADERS },
    );
    expect(list.status()).toBe(200);
    const listBody = (await list.json()) as { data?: unknown };
    expect(Array.isArray(listBody.data), "occupations GET must return {data:[...]}").toBe(true);

    await openReservationTypePanel(page);
    // The occupations select is the last combobox in the panel (after 紐付け職種,
    // before キャンセル/保存). Its placeholder text is not exposed to the a11y tree.
    const panel = page.locator("div.shadow-panel").first();
    const select = panel.getByRole("combobox").last();
    await expect(select).toBeVisible({ timeout: 30000 });
    await select.click();
    await page.getByRole("option", { name: occName }).click();
    await expect(page.getByText(occName).first()).toBeVisible({ timeout: 15000 });
  });

  test("EMR-206: LIFF vaccine dates render as dates, not RFC3339", async ({ page }) => {
    test.setTimeout(120000);
    // Mock customer for clinic 2 is linked to an owner with vaccination rows.
    await page.goto(`/liff/?clinic_id=${CLINIC_ID}`, { waitUntil: "domcontentloaded" });
    // Vite dev server keeps an HMR socket open — networkidle never settles.
    // Wait for a terminal state instead: vaccine table, empty-state, or error page.
    await page
      .locator(
        "text=/ワクチン記録|ペット情報はありません|LINE認証に失敗しました|ログイン情報が取得できませんでした|健康記録の取得/",
      )
      .first()
      .waitFor({ state: "visible", timeout: 90000 });
    // Real vaccine rows must be present for the check to be meaningful.
    await expect(
      page.getByText("接種日").first(),
      "vaccine table header missing — no rows rendered",
    ).toBeVisible();
    // RFC3339 leak pattern: YYYY-MM-DDT in any visible cell text.
    const rfc3339 = page.locator("text=/\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}/");
    await expect(rfc3339, "vaccine dates still render RFC3339 raw values").toHaveCount(0);
  });
});
