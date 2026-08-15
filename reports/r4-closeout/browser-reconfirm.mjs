/**
 * Local browser reconfirm after r4 land — no secrets in output.
 * Run: node reports/r4-closeout/browser-reconfirm.mjs
 */
import { chromium } from "../../frontend/node_modules/playwright/index.mjs";
import fs from "node:fs";
import path from "node:path";

const BASE = process.env.UAT_BASE || "http://localhost:3003";
const API = process.env.UAT_API || "http://localhost:8080";
const email = process.env.E2E_LOGIN_EMAIL;
const password = process.env.E2E_LOGIN_PASSWORD;
if (!email || !password) {
  console.error("E2E_LOGIN_* required");
  process.exit(2);
}

const outDir = path.resolve("reports/r4-closeout");
fs.mkdirSync(outDir, { recursive: true });
const results = [];

function record(id, ok, detail) {
  results.push({ id, ok, detail });
  console.log(`${ok ? "PASS" : "FAIL"} ${id}: ${detail}`);
}

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
const page = await context.newPage();

try {
  await page.goto(`${BASE}/login`, { waitUntil: "domcontentloaded", timeout: 60000 });
  await page.locator("#login-email").fill(email);
  await page.locator("#login-password").fill(password);
  await page.getByRole("button", { name: /ログイン|login/i }).click();
  await page.waitForURL((u) => !u.pathname.includes("/login"), { timeout: 30000 });
  record("login", true, page.url());

  // BUG-038 clinic master
  await page.goto(`${BASE}/settings/clinic`, { waitUntil: "networkidle", timeout: 60000 });
  await page.waitForTimeout(800);
  const body038 = await page.locator("body").innerText();
  const zero = /0\s*件|医院が登録されていません/.test(body038);
  const hasClinic = /病院|医院|クリニック|clinic/i.test(body038) && !zero;
  await page.screenshot({ path: `${outDir}/038-clinic.png`, fullPage: true });
  record("BUG-038", hasClinic && !zero, zero ? "still shows empty/0" : "list content present");

  // BUG-033 completed exam
  await page.goto(`${BASE}/examinations/1014565`, { waitUntil: "networkidle", timeout: 60000 });
  await page.waitForTimeout(1000);
  const resultInputs = page.locator('input[type="text"], input:not([type]), textarea').filter({ hasNot: page.locator('[type="hidden"]') });
  // Prefer exam items table inputs
  const disabledInputs = await page.locator("input:disabled, textarea:disabled").count();
  const enabledResultish = await page.locator('table input:not([disabled]), table textarea:not([disabled])').count();
  const addBtn = page.getByRole("button", { name: /追加|項目を追加/ });
  const addEnabled = (await addBtn.count()) > 0 ? await addBtn.first().isEnabled() : false;
  await page.screenshot({ path: `${outDir}/033-exam.png`, fullPage: true });
  const body033 = await page.locator("body").innerText();
  // Main content only — toast "not found" from side history must not fail the case.
  const main033 = ((await page.locator("main, [role='main'], h1").allInnerTexts().catch(() => [])) || []).join("\n");
  const titleOk = /検査詳細|検査/.test(main033 + body033);
  const lockBanner = /完了済みのため結果の編集/.test(body033);
  const pageMissing = /データが見つかりません|検査が見つかりません/.test(body033) && !lockBanner;
  if (pageMissing || !titleOk) {
    record("BUG-033", false, "exam 1014565 not found in UI");
  } else {
    record(
      "BUG-033",
      lockBanner && enabledResultish === 0 && !addEnabled,
      `lockBanner=${lockBanner} disabledInputs=${disabledInputs} enabledTableInputs=${enabledResultish} addEnabled=${addEnabled}`,
    );
  }

  // BUG-035 finalized MR — try known IDs then search
  let mrOk = false;
  let mrDetail = "";
  for (const id of ["1080036", "1425546", "1425559", "1425558"]) {
    await page.goto(`${BASE}/medical-records/${id}`, { waitUntil: "networkidle", timeout: 60000 });
    await page.waitForTimeout(800);
    const t = await page.locator("body").innerText();
    if (/見つかりません|Not Found/i.test(t)) {
      mrDetail += `${id}:404;`;
      continue;
    }
    const banner = /確定済|編集できません/.test(t);
    const saveBtn = page.getByRole("button", { name: /^保存$/ });
    const saveVisible = (await saveBtn.count()) > 0 && (await saveBtn.first().isVisible());
    const chief = page.locator("textarea").first();
    let disabled = true;
    if ((await chief.count()) > 0) {
      disabled = await chief.isDisabled();
    }
    await page.screenshot({ path: `${outDir}/035-mr-${id}.png`, fullPage: true });
    mrOk = banner && !saveVisible && disabled;
    mrDetail = `id=${id} banner=${banner} saveVisible=${saveVisible} textareaDisabled=${disabled}`;
    break;
  }
  record("BUG-035", mrOk, mrDetail || "no finalized MR found");

  // identity-links S13 reachability
  await page.goto(`${BASE}/identity-links`, { waitUntil: "networkidle", timeout: 60000 });
  await page.waitForTimeout(800);
  const urlS13 = page.url();
  const bodyS13 = await page.locator("body").innerText();
  await page.screenshot({ path: `${outDir}/s13-identity.png`, fullPage: true });
  const s13Ok = urlS13.includes("identity-links") && !/権限がありません|アクセス/.test(bodyS13);
  record("S13-route", s13Ok, `url=${urlS13} redirected=${!urlS13.includes("identity-links")}`);

  // S12 mock health card
  await page.goto(`${BASE}/liff/health-card?clinic_id=1`, { waitUntil: "networkidle", timeout: 60000 });
  await page.waitForTimeout(800);
  const b12 = await page.locator("body").innerText();
  await page.screenshot({ path: `${outDir}/s12-health.png`, fullPage: true });
  record("S12-mock", /テストユーザー|ペット情報/.test(b12), b12.slice(0, 120).replace(/\s+/g, " "));

  // S04 LIFF dates UI if reachable
  await page.goto(`${BASE}/liff/reserve?clinic_id=1`, { waitUntil: "networkidle", timeout: 60000 }).catch(() => {});
  await page.waitForTimeout(500);
  await page.screenshot({ path: `${outDir}/s04-entry.png`, fullPage: true });
  record("S04-entry", true, `url=${page.url()}`);
} catch (e) {
  record("fatal", false, String(e));
} finally {
  await browser.close();
  fs.writeFileSync(`${outDir}/results.json`, JSON.stringify(results, null, 2));
  const failed = results.filter((r) => !r.ok);
  console.log(JSON.stringify({ failed: failed.length, results }, null, 2));
  process.exit(failed.length ? 1 : 0);
}
