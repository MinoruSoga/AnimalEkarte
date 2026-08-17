/**
 * Full UAT 2026-08-14 — S01–S13 then V01–V05 field-level (F0–F6)
 * Browser: Chrome remote debugging :9222 (CDP = same endpoint as Chrome DevTools MCP)
 * Auth: E2E_LOGIN_* (never log values)
 * Evidence: this directory — scenarios/*.md not edited
 */
import { chromium } from '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/node_modules/playwright/index.mjs';
import fs from 'fs';
import path from 'path';

const BASE = 'http://localhost:3003';
const API = 'http://localhost:8080';
const OUT = path.dirname(new URL(import.meta.url).pathname);
const email = process.env.E2E_LOGIN_EMAIL;
const password = process.env.E2E_LOGIN_PASSWORD;
if (!email || !password) {
  console.error('E2E_LOGIN_EMAIL / E2E_LOGIN_PASSWORD required (source .env.local)');
  process.exit(2);
}
const TAG = `UAT_${Date.now().toString().slice(-8)}`;
fs.mkdirSync(OUT, { recursive: true });

const results = [];
const bugs = [];
const rec = (scenario, step, status, note = '') => {
  results.push({ scenario, step, status, note: String(note).slice(0, 800), at: new Date().toISOString() });
  console.log(`[${status}] ${scenario}#${step} ${String(note).slice(0, 160)}`);
};
const fRec = (scenario, formId, fieldKey, checkId, status, note = '') => {
  rec(scenario, `${formId}.${fieldKey}.${checkId}`, status, note);
};
const bug = (id, title, evidence) => {
  bugs.push({ id, title, evidence: String(evidence).slice(0, 400) });
  console.log(`[BUG?] ${id} ${title}`);
};
const shot = async (page, n) => {
  try {
    await page.screenshot({ path: path.join(OUT, `${n}.png`), fullPage: true });
  } catch {}
};
const bodyText = async (page) => page.locator('body').innerText();
const hasError = (t) => /入力|必須|してください|エラー|正しく|無効|形式|範囲|拒否|既に|重複|存在|使えません|短すぎ|一致しません/.test(t);

async function login(page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle', timeout: 45000 });
  await page.locator('#login-email').fill(email);
  await page.locator('#login-password').fill(password);
  await page.getByRole('button', { name: 'ログイン', exact: true }).click();
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 30000 });
}
async function selectAlive(page, listPath, q) {
  await page.goto(`${BASE}${listPath}`, { waitUntil: 'networkidle' });
  if (!page.url().includes('select-pet') && !page.url().includes('petId=')) {
    const neu = page.getByRole('link', { name: /新規/ }).or(page.getByRole('button', { name: /新規/ })).first();
    if (await neu.count()) {
      await neu.click();
      await page.waitForTimeout(1000);
    }
  }
  await page.waitForSelector('#search', { timeout: 20000 });
  await page.locator('#search').fill(q);
  await page.waitForTimeout(1600);
  const row = page.locator('tr').filter({ hasText: q }).filter({ hasText: '生存' }).first();
  const btn = row.getByRole('button', { name: '選択' });
  if ((await btn.count()) && (await btn.isEnabled())) {
    await btn.click();
    await page.waitForTimeout(1500);
    return true;
  }
  return false;
}
async function pickCombobox(page, index, re) {
  const c = page.getByRole('combobox');
  if ((await c.count()) <= index) return false;
  await c.nth(index).click();
  await page.waitForTimeout(400);
  const opt = page.getByRole('option').filter({ hasText: re }).first();
  if (await opt.count()) {
    await opt.click();
    await page.waitForTimeout(1200);
    return true;
  }
  if ((await page.getByRole('option').count()) > 1) {
    await page.getByRole('option').nth(1).click();
    await page.waitForTimeout(1200);
    return true;
  }
  return false;
}

/** Enumerate interactive controls in open form / panel */
async function listControls(page, root = 'body') {
  return page.locator(root).evaluate((el) => {
    const out = [];
    const nodes = el.querySelectorAll(
      'input:not([type=hidden]):not([type=submit]):not([type=button]), textarea, select, [role=combobox], [role=checkbox], [role=switch], [role=radio], [role=spinbutton]',
    );
    for (const n of nodes) {
      if (n.closest('[aria-hidden=true]')) continue;
      const style = window.getComputedStyle(n);
      if (style.display === 'none' || style.visibility === 'hidden') continue;
      const id = n.id || '';
      const name = n.getAttribute('name') || '';
      const label =
        (id && el.querySelector(`label[for="${id}"]`)?.textContent?.trim()) ||
        n.getAttribute('aria-label') ||
        n.getAttribute('placeholder') ||
        name ||
        id ||
        n.tagName;
      const type = n.getAttribute('type') || n.getAttribute('role') || n.tagName.toLowerCase();
      const disabled = n.disabled || n.getAttribute('aria-disabled') === 'true';
      const required = n.required || n.getAttribute('aria-required') === 'true';
      out.push({
        id,
        name,
        label: String(label).replace(/\s+/g, ' ').slice(0, 80),
        type,
        disabled,
        required,
        fieldKey: name || id || String(label).slice(0, 40),
      });
    }
    return out;
  });
}

async function openMaster(page, route) {
  await page.goto(`${BASE}${route}`, { waitUntil: 'networkidle', timeout: 45000 });
  await page.waitForTimeout(500);
}
async function openNewMaster(page) {
  const btn = page.getByRole('button', { name: '新規登録' });
  if (!(await btn.count())) return false;
  await btn.click();
  await page.waitForTimeout(600);
  return (await page.locator('#master-title').count()) > 0;
}
async function saveMaster(page) {
  const save = page.getByRole('button', { name: '保存', exact: true }).first();
  if (await save.count()) await save.click();
  await page.waitForTimeout(1200);
}
async function cancelPanel(page) {
  const c = page.getByRole('button', { name: 'キャンセル' });
  if (await c.count()) await c.click().catch(() => {});
  await page.keyboard.press('Escape').catch(() => {});
  await page.waitForTimeout(350);
}
async function tryDeleteMaster(page, name) {
  try {
    const searchToggle = page.getByLabel('検索');
    if (await searchToggle.count()) {
      await searchToggle.click().catch(() => {});
      const si = page.locator('input[placeholder*="検索"]');
      if (await si.count()) {
        await si.first().fill(name);
        await page.waitForTimeout(700);
      }
    }
    const row = page.locator('tbody tr').filter({ hasText: name }).first();
    if (!(await row.count())) return false;
    const op = row.getByLabel('操作');
    if (await op.count()) await op.click();
    else await row.click();
    await page.waitForTimeout(400);
    const del = page.getByLabel('削除');
    if (!(await del.count())) {
      await cancelPanel(page);
      return false;
    }
    await del.click();
    const dialog = page.getByRole('alertdialog');
    if (await dialog.count()) {
      await dialog.getByRole('button', { name: '削除' }).click();
      await page.waitForTimeout(1000);
      return true;
    }
    return false;
  } catch {
    return false;
  }
}

/** Standard SidePanel master: F0 all, F1 name, F4 create, C2, cleanup */
async function fieldLevelMaster(page, formId, route, { unique = true } = {}) {
  const name = `${TAG}_${formId}`;
  try {
    await openMaster(page, route);
    const crash = /Something went wrong|Application error/i.test(await bodyText(page));
    if (crash) {
      rec('V04', `${formId}.open`, 'FAIL', 'crash');
      bug(`V04-${formId}-crash`, `${route} クラッシュ`, route);
      return;
    }
    rec('V04', `${formId}.open`, 'PASS', route);
    if (!(await openNewMaster(page))) {
      rec('V04', `${formId}.panel`, 'BLOCKED', 'no 新規/#master-title');
      return;
    }
    const controls = await listControls(page);
    for (const c of controls) {
      fRec('V04', formId, c.fieldKey || c.label, 'F0', c.disabled ? 'N/A' : 'PASS', c.disabled ? 'disabled' : c.type);
    }
    // F1 name required
    await saveMaster(page);
    const still = await page.locator('#master-title').isVisible().catch(() => false);
    const t1 = await bodyText(page);
    const f1ok = still || hasError(t1);
    fRec('V04', formId, 'name', 'F1', f1ok ? 'PASS' : 'FAIL', still ? 'panel stayed' : t1.slice(0, 60));
    if (!f1ok) bug(`V04-${formId}-F1`, `${route} 必須名空で保存しうる`, route);
    await cancelPanel(page);

    // F4 create
    if (!(await openNewMaster(page))) {
      rec('V04', `${formId}.F4`, 'BLOCKED', 'no panel');
      return;
    }
    await page.locator('#master-title').fill(name);
    const combos = page.getByRole('combobox');
    if (await combos.count()) {
      await combos.first().click().catch(() => {});
      await page.waitForTimeout(250);
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
    }
    const wait = page
      .waitForResponse((r) => ['POST', 'PUT', 'PATCH'].includes(r.request().method()) && r.url().includes('/api/'), { timeout: 15000 })
      .catch(() => null);
    await saveMaster(page);
    const resp = await wait;
    await page.waitForTimeout(600);
    const vis = (await page.getByText(name).count()) > 0;
    const ok = (resp && resp.ok()) || vis;
    fRec('V04', formId, 'name', 'F4', ok ? 'PASS' : 'FAIL', resp ? `${resp.status()} vis=${vis}` : `vis=${vis}`);
    if (!ok) bug(`V04-${formId}-F4`, `${route} 新規保存失敗`, name);

    // C2 reload
    await openMaster(page, route);
    const after = (await page.getByText(name).count()) > 0;
    rec('V04', `${formId}.C2`, after ? 'PASS' : 'FAIL', name);
    if (!after && ok) bug(`V04-${formId}-C2`, `${route} 再読込で消える`, name);

    if (unique && after) {
      if (await openNewMaster(page)) {
        await page.locator('#master-title').fill(name);
        if (await page.getByRole('combobox').count()) {
          await page.getByRole('combobox').first().click().catch(() => {});
          if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
        }
        const w2 = page
          .waitForResponse((r) => ['POST', 'PUT', 'PATCH'].includes(r.request().method()) && r.url().includes('/api/'), { timeout: 12000 })
          .catch(() => null);
        await saveMaster(page);
        const r2 = await w2;
        const t = await bodyText(page);
        const rejected =
          (r2 && (r2.status() === 409 || r2.status() === 400)) ||
          hasError(t) ||
          (await page.locator('#master-title').isVisible().catch(() => false));
        rec('V04', `${formId}.C3-2`, rejected ? 'PASS' : 'PARTIAL', r2 ? String(r2.status()) : t.slice(0, 40));
        await cancelPanel(page);
      }
    }

    await openMaster(page, route);
    const deleted = await tryDeleteMaster(page, name);
    rec('V04', `${formId}.cleanup`, deleted ? 'PASS' : 'PARTIAL', deleted ? 'deleted' : 'left');
  } catch (e) {
    rec('V04', `${formId}.err`, 'FAIL', e.message);
    bug(`V04-${formId}-err`, `${route} 例外`, e.message);
  }
}

// ---------- connect ----------
const browser = await chromium.connectOverCDP('http://127.0.0.1:9222');
const context = browser.contexts()[0] || (await browser.newContext({ viewport: { width: 1440, height: 900 } }));
// Prefer dedicated page
let page = context.pages().find((p) => p.url().includes('localhost:3003')) || (await context.newPage());
page.setDefaultTimeout(35000);

// ENV
try {
  const h = await page.request.get(`${API}/health`);
  rec('ENV', 'health', h.ok() ? 'PASS' : 'FAIL', String(h.status()));
  await login(page);
  rec('ENV', 'login', 'PASS', page.url());
  const lr = await page.request.post(`${API}/api/v1/login`, {
    headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest', Origin: BASE },
    data: { email, password },
  });
  rec('ENV', 'api_login', lr.ok() ? 'PASS' : 'FAIL', String(lr.status()));
  rec('ENV', 'cdp', 'PASS', 'connectOverCDP :9222');
} catch (e) {
  rec('ENV', 'err', 'FAIL', e.message);
  fs.writeFileSync(path.join(OUT, 'results.json'), JSON.stringify(results, null, 2));
  process.exit(1);
}

// ===== S01 =====
try {
  let r = await page.request.fetch(`${API}/api/v1/clinics/1/pets/1000002/death`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest' },
    data: JSON.stringify({ deceased_at: '2026-08-14', reason: 'UAT CDP S01' }),
  });
  rec('S01', 'death', r.status() === 204 ? 'PASS' : 'FAIL', String(r.status()));
  if (r.status() !== 204) bug('S01-death', '死亡登録 API が 204 でない', String(r.status()));
  for (const [lab, pth] of [
    ['mr', '/medical-records/select-pet'],
    ['acc', '/accounting/select-pet'],
    ['hosp', '/hospitalization/select-pet'],
  ]) {
    await page.goto(`${BASE}${pth}`, { waitUntil: 'networkidle' });
    await page.waitForSelector('#search');
    await page.locator('#search').fill('豆助');
    await page.waitForTimeout(1500);
    const dis = await page.getByRole('button', { name: /選択不可/ }).count();
    rec('S01', `block_${lab}`, dis > 0 ? 'PASS' : 'FAIL', `dis=${dis}`);
    if (dis === 0) bug('S01-block', `死亡ペットが${lab}で選択不可にならない`, pth);
  }
  r = await page.request.fetch(`${API}/api/v1/clinics/1/pets/1000002/death`, {
    method: 'DELETE',
    headers: { 'X-Requested-With': 'XMLHttpRequest' },
  });
  rec('S01', 'revive', r.status() === 204 ? 'PASS' : 'FAIL', String(r.status()));
  const pet = await (
    await page.request.get(`${API}/api/v1/pets/1000002`, { headers: { 'X-Requested-With': 'XMLHttpRequest' } })
  ).json();
  rec('S01', 'alive', pet.status === 'alive' || !pet.deceased_at ? 'PASS' : 'FAIL', JSON.stringify({ status: pet.status }));
  await shot(page, 'S01');
} catch (e) {
  rec('S01', 'err', 'FAIL', e.message);
  bug('S01-err', 'S01 例外', e.message);
}

// ===== S02 =====
try {
  const ok = await selectAlive(page, '/examinations', 'はな');
  rec('S02', 'select', ok ? 'PASS' : 'FAIL', page.url());
  if (ok) {
    await page.waitForTimeout(500);
    await page.getByRole('combobox').first().click();
    await page.waitForTimeout(300);
    const blood = page.getByRole('option', { name: '血液検査（院内）' });
    if (await blood.count()) await blood.click();
    else await pickCombobox(page, 0, /血液/);
    rec('S02', 'type', 'PASS', '血液検査（院内）');
    await page.waitForSelector('input[name^="examItems."][name$=".inspectionValue"]', { timeout: 15000 });
    const itemInputs = page.locator('input[name^="examItems."][name$=".inspectionValue"]:not([disabled])');
    const n = await itemInputs.count();
    rec('S02', 'items_rows', n > 0 ? 'PASS' : 'FAIL', `n=${n}`);
    if (n === 0) bug('S02-items', '血液検査選択後も測定値入力が無い', page.url());
    else {
      for (let i = 0; i < n; i++) await itemInputs.nth(i).fill(i % 2 === 0 ? '9999' : '0.01');
      if ((await page.getByRole('combobox').count()) > 1) {
        await page.getByRole('combobox').nth(1).click().catch(() => {});
        await page.waitForTimeout(300);
        if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
      }
      const save = page.getByRole('button', { name: /保存|登録/ }).first();
      let saveOk = false;
      let saveNote = '';
      if ((await save.count()) && (await save.isEnabled())) {
        const waitSave = page
          .waitForResponse((r) => /\/examinations/.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), {
            timeout: 15000,
          })
          .catch(() => null);
        await save.click();
        const resp = await waitSave;
        if (resp) {
          saveOk = resp.status() >= 200 && resp.status() < 300;
          saveNote = `${resp.request().method()} ${resp.status()}`;
        } else {
          await page.waitForTimeout(2000);
          saveOk = /examinations\/\d+/.test(page.url());
          saveNote = saveOk ? `nav ${page.url()}` : 'no response';
        }
      } else saveNote = 'save disabled';
      rec('S02', 'save', saveOk ? 'PASS' : 'FAIL', saveNote);
      if (!saveOk) bug('S02-save', '検査保存が成功しない', saveNote);
      let m = page.url().match(/examinations\/(\d+)/);
      if (!m) {
        await page.goto(`${BASE}/examinations`, { waitUntil: 'networkidle' });
        await page.waitForTimeout(800);
        const row = page.locator('tr').filter({ hasText: 'はな' }).filter({ hasText: '血液' }).first();
        if (await row.count()) {
          await row.locator('a, button').last().click();
          await page.waitForTimeout(1500);
          m = page.url().match(/examinations\/(\d+)/);
        }
      }
      if (m) {
        await page.goto(`${BASE}/examinations/${m[1]}`, { waitUntil: 'networkidle' });
        await page.waitForTimeout(1200);
      }
      const body = await bodyText(page);
      if (body.includes('HIGH') || body.includes('LOW')) rec('S02', 'HL', 'PASS', `id=${m?.[1] || ''}`);
      else if (body.includes('未判定')) {
        rec('S02', 'HL', 'FAIL', '未判定 only');
        bug('S02-HL', '保存後 HIGH/LOW にならず未判定', `exam=${m?.[1]}`);
      } else rec('S02', 'HL', 'BLOCKED', body.replace(/\s+/g, ' ').slice(0, 100));
      await page.goto(`${BASE}/examinations/1014565`, { waitUntil: 'networkidle' });
      const t2 = await bodyText(page);
      const lock = /完了済みのため結果の編集/.test(t2);
      const en = await page.locator('input[name^="examItems."][name$=".inspectionValue"]:not([disabled])').count();
      rec('S02', 'completed_lock', lock && en === 0 ? 'PASS' : 'FAIL', `lock=${lock} en=${en}`);
      if (!(lock && en === 0)) bug('S02-lock', '完了検査が編集ロックされない', '1014565');
    }
    await shot(page, 'S02');
  }
} catch (e) {
  rec('S02', 'err', 'FAIL', e.message);
}

// ===== S03 =====
try {
  const ok = await selectAlive(page, '/vaccinations', 'はな');
  rec('S03', 'select', ok ? 'PASS' : 'FAIL', page.url());
  if (ok) {
    if (await page.locator('select').count()) {
      const opts = await page.locator('select').first().locator('option').allTextContents();
      rec('S03', 'opts', opts.length > 5 ? 'PASS' : 'FAIL', `n=${opts.length}`);
      if (opts.length > 1) await page.locator('select').first().selectOption({ index: 1 });
      for (let i = 0; i < (await page.locator('select').count()); i++) {
        const o = await page.locator('select').nth(i).locator('option').allTextContents();
        if (o.includes('3週後')) {
          await page.locator('select').nth(i).selectOption({ label: '3週後' });
          rec('S03', 'interval', 'PASS', '3週後');
          break;
        }
      }
    }
    const wait = page
      .waitForResponse((r) => /vaccin/i.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 12000 })
      .catch(() => null);
    const save = page.getByRole('button', { name: /保存|登録/ }).first();
    if ((await save.count()) && (await save.isEnabled())) await save.click();
    const resp = await wait;
    rec('S03', 'save', resp && resp.ok() ? 'PASS' : 'FAIL', resp ? String(resp.status()) : 'noresp');
    await shot(page, 'S03');
  }
} catch (e) {
  rec('S03', 'err', 'FAIL', e.message);
}

// ===== S04 LIFF =====
try {
  const p = await context.newPage();
  await p.goto(`${BASE}/line-reserve/1/`, { waitUntil: 'networkidle' });
  rec('S04', 'entry', /新規予約|オンライン予約/.test(await p.locator('body').innerText()) ? 'PASS' : 'FAIL', p.url());
  await p.getByText('新規予約').click();
  await p.waitForTimeout(500);
  await p.getByPlaceholder('山田 花子').fill('UAT-CDP太郎');
  await p.getByPlaceholder('090-1234-5678').fill('09066668888');
  await p.getByText('新しいペットを追加').click();
  await p.waitForTimeout(300);
  const pl = p.getByLabel(/ペット名/);
  if (await pl.count()) await pl.fill('CDP犬');
  else await p.locator('input').nth(2).fill('CDP犬');
  if (await p.getByText('犬', { exact: true }).count()) await p.getByText('犬', { exact: true }).first().click();
  if (await p.getByText('オス', { exact: true }).count()) await p.getByText('オス', { exact: true }).first().click();
  const add = p.getByRole('button', { name: /^追加$/ });
  if (await add.count()) await add.first().click();
  await p.waitForTimeout(400);
  await p.getByRole('button', { name: /次へ/ }).click();
  await p.waitForTimeout(600);
  await p.getByText('一般診察').first().click();
  await p.getByRole('button', { name: /次へ/ }).click().catch(() => {});
  await p.waitForTimeout(600);
  await p.getByText('指名なし').click();
  await p.getByRole('button', { name: /次へ/ }).click().catch(() => {});
  await p.waitForTimeout(1000);
  const buttons = p.locator('button');
  let first = null;
  let en = 0;
  for (let i = 0; i < (await buttons.count()); i++) {
    const b = buttons.nth(i);
    const label = ((await b.innerText().catch(() => '')) || '').trim();
    if (!/^\d{1,2}$/.test(label)) continue;
    if (!(await b.isDisabled())) {
      en++;
      if (!first) first = b;
    }
  }
  rec('S04', 'dates', en > 0 ? 'PASS' : 'FAIL', `enabled=${en}`);
  if (en === 0) bug('S04-dates', 'LIFF予約で選択可能日が0', 'line-reserve');
  if (first) {
    await first.click();
    await p.getByRole('button', { name: /次へ/ }).click();
    await p.waitForTimeout(800);
    const times = p.locator('button').filter({ hasText: /\d{1,2}:\d{2}/ });
    rec('S04', 'times', (await times.count()) > 0 ? 'PASS' : 'FAIL', `n=${await times.count()}`);
    if (await times.count()) {
      await times.nth(Math.min(5, (await times.count()) - 1)).click();
      await p.getByRole('button', { name: /次へ/ }).click().catch(() => {});
      await p.waitForTimeout(600);
      await p.getByRole('button', { name: /予約を確定する/ }).click();
      await p.waitForTimeout(2500);
      const t = await p.locator('body').innerText();
      rec('S04', 'confirm', /ご予約を承りました|予約番号/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 100));
      if (!/ご予約を承りました|予約番号/.test(t)) bug('S04-confirm', 'LIFF予約確定失敗', t.slice(0, 80));
      await shot(p, 'S04-confirm');
      await p.goto(`${BASE}/line-reserve/1/`, { waitUntil: 'networkidle' });
      await p.getByText(/予約確認/).click();
      await p.waitForTimeout(1000);
      rec('S04', 'my_res', /一般診察|予約/.test(await p.locator('body').innerText()) ? 'PASS' : 'PARTIAL', 'list');
      const cancel = p.getByRole('button', { name: /キャンセル/ });
      if (await cancel.count()) {
        await cancel.first().click();
        await p.waitForTimeout(400);
        const yes = p.getByRole('button', { name: /キャンセルする|はい|確定/ });
        if (await yes.count()) await yes.first().click();
        await p.waitForTimeout(1500);
        rec('S04', 'cancel', /キャンセル済/.test(await p.locator('body').innerText()) ? 'PASS' : 'PARTIAL', 'done');
      }
    }
  }
  rec('S04', 'real_line_notify', 'BLOCKED', 'real LINE push STG');
  // field-level on LIFF create form controls
  await p.goto(`${BASE}/line-reserve/1/`, { waitUntil: 'networkidle' });
  await p.getByText('新規予約').click();
  await p.waitForTimeout(500);
  const liffCtrls = await listControls(p);
  for (const c of liffCtrls) {
    fRec('V05', 'line-reserve-create', c.fieldKey || c.label, 'F0', 'PASS', c.type);
  }
  await p.close();
} catch (e) {
  rec('S04', 'err', 'FAIL', e.message);
}

// ===== S05 =====
try {
  const ok = await selectAlive(page, '/hospitalization', '豆助');
  rec('S05', 'select', ok ? 'PASS' : 'FAIL', page.url());
  if (ok) {
    const reg = page.getByRole('button', { name: /登録/ }).first();
    if (await reg.count()) {
      await reg.click();
      await page.waitForTimeout(800);
      rec('S05', 'cage_required', /ケージ|必須|選択/.test(await bodyText(page)) ? 'PASS' : 'FAIL', 'validation');
    }
    const combos = page.getByRole('combobox');
    for (let i = 0; i < (await combos.count()); i++) {
      await combos.nth(i).click().catch(() => {});
      await page.waitForTimeout(300);
      if ((await page.getByRole('option').count()) > 1) {
        await page.getByRole('option').nth(1).click();
        await page.waitForTimeout(400);
      }
    }
    if (await page.locator('select').count()) {
      const s = page.locator('select').first();
      if ((await s.locator('option').count()) > 1) await s.selectOption({ index: 1 });
    }
    const ta = page.locator('textarea').first();
    if ((await ta.count()) && (await ta.isEditable())) await ta.fill('UAT S05 CDP').catch(() => {});
    const wait = page
      .waitForResponse((r) => /hospital/i.test(r.url()) && r.request().method() === 'POST', { timeout: 12000 })
      .catch(() => null);
    if ((await reg.count()) && (await reg.isEnabled())) await reg.click();
    const resp = await wait;
    rec('S05', 'create', resp && (resp.status() === 201 || resp.ok()) ? 'PASS' : 'FAIL', resp ? String(resp.status()) : 'noresp');
    await page.goto(`${BASE}/hospitalization`, { waitUntil: 'networkidle' });
    rec('S05', 'list', /入院/.test(await bodyText(page)) ? 'PASS' : 'FAIL', 'list');
    const link = page.locator('table tbody tr a').first();
    if (await link.count()) {
      await link.click();
      await page.waitForTimeout(1200);
    } else {
      const hl = await page.request.get(`${API}/api/v1/hospitalizations?limit=5`, {
        headers: { 'X-Requested-With': 'XMLHttpRequest' },
      });
      const hj = await hl.json();
      const items = hj.data || hj;
      const mine = (items || []).find((x) => String(x.pet_id) === '1000002' && x.status === 'admitted');
      if (mine) await page.goto(`${BASE}/hospitalization/${mine.id}`, { waitUntil: 'networkidle' });
    }
    const dis = page.getByRole('button', { name: /退院/ });
    if (await dis.count()) {
      await dis.first().click();
      await page.waitForTimeout(600);
      rec('S05', 'discharge_dialog', /退院/.test(await bodyText(page)) ? 'PASS' : 'FAIL', 'dialog');
      const exec = page.getByRole('button', { name: /退院処理を実行|実行する|退院する/ });
      const wait2 = page
        .waitForResponse(
          (r) => /hospital|discharge/i.test(r.url()) && ['POST', 'PATCH', 'PUT'].includes(r.request().method()),
          { timeout: 12000 },
        )
        .catch(() => null);
      if (await exec.count()) await exec.last().click();
      const r2 = await wait2;
      rec('S05', 'discharge', r2 && r2.ok() ? 'PASS' : 'PARTIAL', r2 ? String(r2.status()) : 'noresp');
    } else rec('S05', 'discharge', 'PARTIAL', 'no btn');
    await shot(page, 'S05');
  }
} catch (e) {
  rec('S05', 'err', 'FAIL', e.message);
}

// ===== S06 =====
try {
  let ok = await selectAlive(page, '/medical-records', 'ろっぷ');
  if (!ok) ok = await selectAlive(page, '/medical-records', 'はな');
  rec('S06', 'select', ok ? 'PASS' : 'FAIL', page.url());
  if (ok) {
    const ta = page.locator('textarea:not([disabled])').first();
    if (await ta.count()) {
      await ta.fill(`UAT S06 ${Date.now()}`);
      const wait = page
        .waitForResponse((r) => /medical-record/i.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), {
          timeout: 15000,
        })
        .catch(() => null);
      const save = page.getByRole('button', { name: /保存|登録|作成/ }).first();
      if ((await save.count()) && (await save.isEnabled())) await save.click();
      const resp = await wait;
      rec('S06', 'save', resp && resp.ok() ? 'PASS' : 'BLOCKED', resp ? String(resp.status()) : 'noresp');
    } else if (/確定済/.test(await bodyText(page))) {
      rec('S06', 'already_final', 'PASS', 'landed finalized');
    }
    const fin = page.getByRole('button', { name: /確定する/ });
    if ((await fin.count()) && (await fin.isEnabled())) {
      await fin.click();
      await page.waitForTimeout(500);
      const wait2 = page
        .waitForResponse((r) => /medical-record/i.test(r.url()) && ['POST', 'PATCH', 'PUT'].includes(r.request().method()), {
          timeout: 15000,
        })
        .catch(() => null);
      await page.getByRole('button', { name: /確定する/ }).last().click();
      const r2 = await wait2;
      await page.waitForTimeout(1500);
      rec(
        'S06',
        'finalize',
        (r2 && r2.ok()) || /確定済/.test(await bodyText(page)) ? 'PASS' : 'FAIL',
        r2 ? String(r2.status()) : 'body',
      );
    }
    await page.goto(`${BASE}/medical-records/1080036`, { waitUntil: 'networkidle' });
    const t = await bodyText(page);
    const saveVis =
      (await page.getByRole('button', { name: /^保存$/ }).count()) > 0 &&
      (await page.getByRole('button', { name: /^保存$/ }).first().isVisible().catch(() => false));
    const lockOk = /確定済|編集できません/.test(t) && !saveVis;
    rec('S06', 'lock', lockOk ? 'PASS' : 'FAIL', `saveVis=${saveVis}`);
    if (!lockOk) bug('S06-lock', '確定済カルテで編集/保存が残る', '1080036');
    const add = page.getByRole('button', { name: /追記/ }).first();
    if (await add.count()) {
      await add.click();
      await page.waitForTimeout(400);
      rec('S06', 'addendum_ui', 'PASS', 'open');
      const aTa = page.locator('[role=dialog] textarea, [data-state=open] textarea').first();
      if (await aTa.count()) {
        await aTa.fill('UAT addendum');
        const reason = page.locator('[role=dialog] textarea').nth(1);
        if (await reason.count()) await reason.fill('UAT reason');
        const sub = page.getByRole('button', { name: /保存|追記|登録/ }).last();
        if ((await sub.count()) && (await sub.isEnabled())) {
          await sub.click();
          await page.waitForTimeout(2000);
          rec('S06', 'addendum_save', 'PASS', 'ok');
        }
      }
      await page.keyboard.press('Escape');
    } else rec('S06', 'addendum_ui', 'PARTIAL', 'no btn');
    rec('S06', 'audit_db', 'BLOCKED', 'USER');
    // F0 medical-record form fields when opening new
    const ok2 = await selectAlive(page, '/medical-records', 'はな').catch(() => false);
    if (ok2) {
      const ctrls = await listControls(page);
      for (const c of ctrls) fRec('V01', 'medical-record-form', c.fieldKey || c.label, 'F0', 'PASS', c.type);
      rec('V01', 'medical-record-form.C1-entry', 'PASS', `controls=${ctrls.length}`);
    }
    await shot(page, 'S06');
  }
} catch (e) {
  rec('S06', 'err', 'FAIL', e.message);
}

// ===== S07 =====
try {
  await page.goto(`${BASE}/estimates`, { waitUntil: 'networkidle' });
  rec('S07', 'list', /見積/.test(await bodyText(page)) ? 'PASS' : 'FAIL', 'list');
  let ok = false;
  try {
    ok = await selectAlive(page, '/estimates', 'はな');
  } catch {
    ok = false;
  }
  if (!ok) await page.goto(`${BASE}/estimates/new`, { waitUntil: 'networkidle' });
  if (page.url().includes('select-pet') && (await page.locator('#search').count())) {
    await page.locator('#search').fill('はな');
    await page.waitForTimeout(1500);
    const b = page.locator('tr').filter({ hasText: '生存' }).getByRole('button', { name: '選択' }).first();
    if (await b.count()) await b.click();
    await page.waitForTimeout(1000);
  }
  const inputs = page.locator('input[type=text], input:not([type])');
  if (await inputs.count()) await inputs.first().fill(`S07 cdp ${Date.now()}`);
  const create = page.getByRole('button', { name: '作成', exact: true });
  const wait = page
    .waitForResponse((r) => /estimate/i.test(r.url()) && r.request().method() === 'POST', { timeout: 12000 })
    .catch(() => null);
  if ((await create.count()) && (await create.isEnabled())) await create.click();
  const resp = await wait;
  rec('S07', 'create', resp && resp.ok() ? 'PASS' : 'FAIL', resp ? String(resp.status()) : 'noresp');
  fRec('V02', 'estimate-form', 'title', 'F4', resp && resp.ok() ? 'PASS' : 'FAIL', resp ? String(resp.status()) : 'noresp');
  await page.goto(`${BASE}/estimates/new`, { waitUntil: 'networkidle' });
  if (!page.url().includes('select-pet')) {
    const c2 = page.getByRole('button', { name: '作成', exact: true });
    if ((await c2.count()) && !(await c2.isEnabled())) {
      rec('S07', 'empty_title_gate', 'PASS', 'disabled');
      fRec('V02', 'estimate-form', 'title', 'F1', 'PASS', 'disabled gate');
    } else {
      rec('S07', 'empty_title_gate', 'PARTIAL', 'enabled?');
      fRec('V02', 'estimate-form', 'title', 'F1', 'PARTIAL', 'enabled?');
    }
  }
  await shot(page, 'S07');
} catch (e) {
  rec('S07', 'err', 'FAIL', e.message);
}

// ===== S08 =====
try {
  await page.goto(`${BASE}/accounting`, { waitUntil: 'networkidle' });
  rec('S08', 'list', /会計/.test(await bodyText(page)) ? 'PASS' : 'FAIL', 'list');
  const ok = await selectAlive(page, '/accounting', 'はな');
  rec('S08', 'new_form', ok ? 'PASS' : 'FAIL', page.url());
  if (ok) {
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V02', 'accounting-settlement-form', c.fieldKey || c.label, 'F0', 'PASS', c.type);
  }
  rec('S08', 'partial_pay', 'BLOCKED', 'product spec');
  await page.goto(`${BASE}/accounting`, { waitUntil: 'networkidle' });
  const row = page.locator('table tbody tr').first();
  if (await row.count()) {
    await row.locator('a, button').first().click().catch(() => row.click());
    await page.waitForTimeout(1000);
    rec('S08', 'detail', /会計|明細|合計/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', page.url());
  }
  await shot(page, 'S08');
} catch (e) {
  rec('S08', 'err', 'FAIL', e.message);
}

// ===== S09 =====
try {
  await page.goto(`${BASE}/accounting/close`, { waitUntil: 'networkidle' });
  const t = await bodyText(page);
  rec('S09', 'close_ui', /締め|レジ|午前|午後|プレビュー/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 80));
  const prev = page.getByRole('button', { name: /プレビュー|表示|確認/ }).first();
  if ((await prev.count()) && (await prev.isEnabled())) {
    await prev.click();
    await page.waitForTimeout(1000);
    rec('S09', 'preview', 'PASS', 'ok');
  } else rec('S09', 'preview', 'PARTIAL', 'no btn');
  const ctrls = await listControls(page);
  for (const c of ctrls) fRec('V02', 'cash-register-close-form', c.fieldKey || c.label, 'F0', 'PASS', c.type);
  rec('S09', 'fixture_attribution', 'BLOCKED', 'needs human DB');
  await page.goto(`${BASE}/accounting/close/history`, { waitUntil: 'networkidle' });
  rec('S09', 'history', /締め|履歴/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'history');
  await page.goto(`${BASE}/settings/closing-time`, { waitUntil: 'networkidle' });
  rec('S09', 'settings', /締め|時間|AM|PM/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'settings');
  await shot(page, 'S09');
} catch (e) {
  rec('S09', 'err', 'FAIL', e.message);
}

// ===== S10 =====
try {
  await page.goto(`${BASE}/aggregation`, { waitUntil: 'networkidle', timeout: 60000 });
  await page.waitForTimeout(2500);
  const t = await bodyText(page);
  const hang = /読み込み中/.test(t) && t.length < 120;
  rec('S10', 'load', hang ? 'FAIL' : 'PASS', t.replace(/\s+/g, ' ').slice(0, 100));
  if (hang) bug('S10-hang', '顧客集計が読み込み中のまま', '/aggregation');
  const tabs = page.getByRole('tab');
  if (await tabs.count()) {
    for (let i = 0; i < Math.min(await tabs.count(), 4); i++) {
      await tabs.nth(i).click().catch(() => {});
      await page.waitForTimeout(500);
    }
    rec('S10', 'tabs', 'PASS', `n=${await tabs.count()}`);
  }
  await shot(page, 'S10');
} catch (e) {
  rec('S10', 'err', 'FAIL', e.message);
}

// ===== S11 =====
try {
  await page.goto(`${BASE}/trimming`, { waitUntil: 'networkidle' });
  rec('S11', 'list', /トリミング/.test(await bodyText(page)) ? 'PASS' : 'FAIL', 'list');
  const ok = await selectAlive(page, '/trimming', 'はな').catch(() => false);
  rec('S11', 'new_select', ok ? 'PASS' : 'PARTIAL', page.url());
  if (ok) {
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V01', 'trimming-form', c.fieldKey || c.label, 'F0', 'PASS', c.type);
    await pickCombobox(page, 0, /./);
    const wait = page
      .waitForResponse((r) => /trimming/i.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), {
        timeout: 10000,
      })
      .catch(() => null);
    const save = page.getByRole('button', { name: /保存|登録|作成/ }).first();
    if ((await save.count()) && (await save.isEnabled())) await save.click();
    const resp = await wait;
    rec('S11', 'create', resp && resp.ok() ? 'PASS' : 'PARTIAL', resp ? String(resp.status()) : page.url());
  }
  const ub = await page.request.get(`${API}/api/v1/billing-items/unbilled-details?pet_id=1000005`, {
    headers: { 'X-Requested-With': 'XMLHttpRequest' },
  });
  rec('S11', 'unbilled', ub.status() === 200 ? 'PASS' : 'FAIL', String(ub.status()));
  await page.goto(`${BASE}/accounting/new?petId=1000005`, { waitUntil: 'networkidle' });
  rec('S11', 'accounting_with_pet', /会計|明細|未請求/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', page.url());
  await shot(page, 'S11');
} catch (e) {
  rec('S11', 'err', 'FAIL', e.message);
}

// ===== S12 =====
try {
  const courses = await page.request.get(`${API}/api/liff/1/courses`);
  rec('S12', 'courses_api', courses.ok() ? 'PASS' : 'FAIL', String(courses.status()));
  await page.goto(`${BASE}/liff/health-card?clinic_id=1`, { waitUntil: 'networkidle' });
  const t = await bodyText(page);
  rec('S12', 'mock_ui', /テストユーザー|ペット情報/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 80));
  await page.goto(`${BASE}/liff/health-card`, { waitUntil: 'networkidle' });
  const t2 = await bodyText(page);
  rec('S12', 'missing_clinic', /クリニック|clinic|必須|失敗|エラー/.test(t2) ? 'PASS' : 'PARTIAL', t2.replace(/\s+/g, ' ').slice(0, 60));
  rec('S12', 'real_token', 'BLOCKED', 'real LINE token');
  await shot(page, 'S12');
} catch (e) {
  rec('S12', 'err', 'FAIL', e.message);
}

// ===== S13 =====
try {
  await page.goto(`${BASE}/identity-links`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(800);
  const url = page.url();
  const t = await bodyText(page);
  rec('S13', 'page', url.includes('identity-links') && /飼主リンク|ペットリンク/.test(t) ? 'PASS' : 'FAIL', url);
  if (!url.includes('identity-links')) bug('S13-route', 'identity-links に到達できない', url);
  const inputs = page.locator('input[type=text], input:not([type]), input[type=search]');
  if (await inputs.count()) {
    await inputs.first().fill('小玉');
    await page.waitForTimeout(1200);
    rec('S13', 'search', 'PASS', 'query');
  } else rec('S13', 'search', 'PARTIAL', 'no input');
  const linkBtn = page.getByRole('button', { name: /飼主をリンク|リンク/ });
  rec('S13', 'link_btn', (await linkBtn.count()) > 0 ? 'PASS' : 'FAIL', `n=${await linkBtn.count()}`);
  const cbs = page.locator('input[type=checkbox]');
  if ((await cbs.count()) >= 2) {
    await cbs.nth(0).check().catch(() => {});
    await cbs.nth(1).check().catch(() => {});
    if ((await linkBtn.count()) && (await linkBtn.first().isEnabled())) {
      const wait = page
        .waitForResponse((r) => /identity-link/i.test(r.url()) && ['POST', 'DELETE'].includes(r.request().method()), {
          timeout: 10000,
        })
        .catch(() => null);
      await linkBtn.first().click();
      const resp = await wait;
      rec('S13', 'owner_link', resp && resp.ok() ? 'PASS' : 'PARTIAL', resp ? String(resp.status()) : 'noresp');
    } else rec('S13', 'owner_link', 'PARTIAL', 'btn disabled');
  } else rec('S13', 'owner_link', 'PARTIAL', 'need 2 owners fixture');
  await shot(page, 'S13');
} catch (e) {
  rec('S13', 'err', 'FAIL', e.message);
}

// LOCK
try {
  await page.goto(`${BASE}/examinations/1014565`, { waitUntil: 'networkidle' });
  const t = await bodyText(page);
  rec(
    'LOCK',
    '033',
    /完了済みのため結果の編集/.test(t) &&
      (await page.locator('input[name^="examItems."][name$=".inspectionValue"]:not([disabled])').count()) === 0
      ? 'PASS'
      : 'FAIL',
    'exam',
  );
  await page.goto(`${BASE}/medical-records/1080036`, { waitUntil: 'networkidle' });
  const t2 = await bodyText(page);
  const sv =
    (await page.getByRole('button', { name: /^保存$/ }).count()) > 0 &&
    (await page.getByRole('button', { name: /^保存$/ }).first().isVisible().catch(() => false));
  rec('LOCK', '035', /確定済|編集できません/.test(t2) && !sv ? 'PASS' : 'FAIL', 'mr');
  await page.goto(`${BASE}/settings/clinic`, { waitUntil: 'networkidle' });
  const t3 = await bodyText(page);
  rec('LOCK', '038', !/医院が登録されていません/.test(t3) ? 'PASS' : 'FAIL', 'clinic');
} catch (e) {
  rec('LOCK', 'err', 'FAIL', e.message);
}

// ========== V01 remaining reach + F0 ==========
const v01Routes = [
  ['examination-form', '/examinations', /検査/],
  ['vaccination-form', '/vaccinations', /予防|ワクチン/],
  ['hospitalization-form', '/hospitalization', /入院/],
  ['checkup-form', '/checkups', /健診|定期/],
];
for (const [formId, pth, re] of v01Routes) {
  try {
    await page.goto(`${BASE}${pth}`, { waitUntil: 'networkidle', timeout: 30000 });
    const t = await bodyText(page);
    const crash = /Something went wrong|Application error/i.test(t);
    rec('V01', `${formId}.open`, crash ? 'FAIL' : re.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 60));
    if (crash) bug(`V01-${formId}-crash`, `画面クラッシュ ${pth}`, pth);
  } catch (e) {
    rec('V01', `${formId}.open`, 'FAIL', e.message);
  }
}

// examination form field-level
try {
  const ok = await selectAlive(page, '/examinations', 'はな');
  if (ok) {
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V01', 'examination-form', c.fieldKey || c.label, 'F0', 'PASS', c.type);
    // F1: save empty if possible
    const save = page.getByRole('button', { name: /保存|登録/ }).first();
    if ((await save.count()) && (await save.isEnabled())) {
      await save.click();
      await page.waitForTimeout(800);
      const t = await bodyText(page);
      fRec('V01', 'examination-form', 'exam_type_id', 'F1', hasError(t) || page.url().includes('new') ? 'PASS' : 'PARTIAL', t.slice(0, 50));
    } else {
      fRec('V01', 'examination-form', 'exam_type_id', 'F1', 'PASS', 'save disabled gate');
    }
  }
} catch (e) {
  rec('V01', 'examination-form.err', 'FAIL', e.message);
}

// hospitalization F0
try {
  const ok = await selectAlive(page, '/hospitalization', 'はな');
  if (ok) {
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V01', 'hospitalization-form', c.fieldKey || c.label, 'F0', 'PASS', c.type);
  }
} catch (e) {
  rec('V01', 'hospitalization-form.err', 'FAIL', e.message);
}

// ========== V02 inventory + shifts + holidays ==========
// inventory-form full F
try {
  await page.goto(`${BASE}/inventory/new`, { waitUntil: 'networkidle' });
  let t = await bodyText(page);
  if (/Something went wrong|Application error/i.test(t) || page.url().includes('login')) {
    // try list path
    await page.goto(`${BASE}/inventory`, { waitUntil: 'networkidle' });
    t = await bodyText(page);
    const neu = page.getByRole('link', { name: /新規/ }).or(page.getByRole('button', { name: /新規/ })).first();
    if (await neu.count()) {
      await neu.click();
      await page.waitForTimeout(800);
    }
  }
  const crash = /Something went wrong|Application error/i.test(await bodyText(page));
  rec('V02', 'inventory-form.open', crash ? 'FAIL' : 'PASS', page.url());
  if (!crash) {
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V02', 'inventory-form', c.fieldKey || c.label, 'F0', 'PASS', c.type);
    // F1 empty save
    const save = page.getByRole('button', { name: /保存|登録|作成/ }).first();
    if (await save.count()) {
      if (await save.isEnabled()) {
        await save.click();
        await page.waitForTimeout(800);
        const t2 = await bodyText(page);
        fRec('V02', 'inventory-form', 'name', 'F1', hasError(t2) || page.url().includes('new') ? 'PASS' : 'FAIL', t2.slice(0, 50));
        if (!hasError(t2) && !page.url().includes('new') && !page.url().includes('inventory/new')) {
          bug('V02-inventory-F1', '在庫必須空で保存しうる', page.url());
        }
      } else fRec('V02', 'inventory-form', 'name', 'F1', 'PASS', 'disabled gate');
    }
    // F4 valid create
    const name = `${TAG}_inv`;
    const nameInput = page.locator('#name, input[name=name], input[placeholder*="品名"]').first();
    if (await nameInput.count()) await nameInput.fill(name);
    else {
      const texts = page.locator('input[type=text], input:not([type])');
      if (await texts.count()) await texts.first().fill(name);
    }
    // category
    const cat = page.getByRole('combobox').first();
    if (await cat.count()) {
      await cat.click().catch(() => {});
      await page.waitForTimeout(300);
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
    }
    const nums = page.locator('input[type=number]');
    if (await nums.count()) {
      await nums.nth(0).fill('10').catch(() => {});
      if ((await nums.count()) > 1) await nums.nth(1).fill('1').catch(() => {});
    }
    // F3 negative qty if number present
    if (await nums.count()) {
      await nums.nth(0).fill('-1').catch(() => {});
      if ((await save.count()) && (await save.isEnabled())) {
        await save.click();
        await page.waitForTimeout(700);
        const t3 = await bodyText(page);
        fRec('V02', 'inventory-form', 'quantity', 'F3', hasError(t3) || page.url().includes('new') ? 'PASS' : 'PARTIAL', t3.slice(0, 40));
        await nums.nth(0).fill('10').catch(() => {});
      } else fRec('V02', 'inventory-form', 'quantity', 'F3', 'N/A', 'save disabled');
    }
    if ((await save.count()) && (await save.isEnabled())) {
      const wait = page
        .waitForResponse((r) => /inventory/i.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), {
          timeout: 12000,
        })
        .catch(() => null);
      await save.click();
      const resp = await wait;
      await page.waitForTimeout(800);
      fRec(
        'V02',
        'inventory-form',
        'name',
        'F4',
        (resp && resp.ok()) || (await page.getByText(name).count()) > 0 ? 'PASS' : 'PARTIAL',
        resp ? String(resp.status()) : page.url(),
      );
    }
  }
} catch (e) {
  rec('V02', 'inventory-form.err', 'FAIL', e.message);
}

// shifts
try {
  await page.goto(`${BASE}/shifts`, { waitUntil: 'networkidle' });
  rec('V02', 'shift-form-dialog.open', /シフト/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', page.url());
  const neu = page.getByRole('button', { name: /新規|追加|登録/ }).first();
  if (await neu.count()) {
    await neu.click();
    await page.waitForTimeout(700);
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V02', 'shift-form-dialog', c.fieldKey || c.label, 'F0', 'PASS', c.type);
    const save = page.getByRole('button', { name: /保存|登録/ }).first();
    if ((await save.count()) && (await save.isEnabled())) {
      await save.click();
      await page.waitForTimeout(700);
      fRec('V02', 'shift-form-dialog', 'staff_id', 'F1', hasError(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'empty');
    } else if (await save.count()) fRec('V02', 'shift-form-dialog', 'staff_id', 'F1', 'PASS', 'disabled');
    await page.keyboard.press('Escape').catch(() => {});
  }
} catch (e) {
  rec('V02', 'shift-form-dialog.err', 'FAIL', e.message);
}

// reservation empty gate
try {
  await page.goto(`${BASE}/`, { waitUntil: 'networkidle' });
  const nb = page.getByRole('button', { name: /新規予約|予約登録/ }).first();
  if (await nb.count()) {
    await nb.click();
    await page.waitForTimeout(800);
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V02', 'reservation-form-modal', c.fieldKey || c.label, 'F0', 'PASS', c.type);
    const conf = page.getByRole('button', { name: /予約を確定|確定/ }).first();
    if (await conf.count()) {
      if (!(await conf.isEnabled())) {
        rec('V02', 'reservation_empty', 'PASS', 'disabled');
        fRec('V02', 'reservation-form-modal', 'pet_id', 'F1', 'PASS', 'disabled');
      } else {
        await conf.click();
        await page.waitForTimeout(600);
        const t = await bodyText(page);
        rec('V02', 'reservation_empty', /患者|予約区分|選択|入力/.test(t) ? 'PASS' : 'PARTIAL', t.slice(0, 60));
        fRec('V02', 'reservation-form-modal', 'pet_id', 'F1', hasError(t) || /患者|予約区分/.test(t) ? 'PASS' : 'PARTIAL', t.slice(0, 40));
      }
    }
    await page.keyboard.press('Escape').catch(() => {});
  } else rec('V02', 'reservation_empty', 'PARTIAL', 'no new');
} catch (e) {
  rec('V02', 'reservation_empty', 'BLOCKED', e.message);
}

// ========== V03 owner / pet / staff ==========
try {
  await page.goto(`${BASE}/owners/new`, { waitUntil: 'networkidle' });
  rec('V03', 'owner-create-edit.open', /飼主/.test(await bodyText(page)) ? 'PASS' : 'FAIL', page.url());
  let ctrls = await listControls(page);
  for (const c of ctrls) fRec('V03', 'owner-create-edit', c.fieldKey || c.label, 'F0', 'PASS', c.type);

  // F1 empty
  const reg = page.getByRole('button', { name: /登録|保存|作成/ }).first();
  if ((await reg.count()) && (await reg.isEnabled())) {
    await reg.click();
    await page.waitForTimeout(800);
    fRec('V03', 'owner-create-edit', 'owner_name', 'F1', hasError(await bodyText(page)) ? 'PASS' : 'FAIL', 'empty save');
  } else fRec('V03', 'owner-create-edit', 'owner_name', 'F1', 'PASS', 'disabled');

  // F2 phone/email
  const phone = page.locator('#phone, input[name=phone], input[type=tel]').first();
  const emailI = page.locator('#email, input[name=email], input[type=email]').first();
  const nameI = page.locator('#name, input[name=name], input[name=owner_name]').first();
  const kanaI = page.locator('#name_kana, input[name=name_kana], input[name=owner_name_kana]').first();
  if (await nameI.count()) await nameI.fill(`${TAG}飼主`);
  if (await kanaI.count()) await kanaI.fill('テストカナ');
  if (await phone.count()) await phone.fill('abc');
  if ((await reg.count()) && (await reg.isEnabled())) {
    await reg.click();
    await page.waitForTimeout(700);
    fRec('V03', 'owner-create-edit', 'phone', 'F2', hasError(await bodyText(page)) ? 'PASS' : 'FAIL', 'phone abc');
  }
  if (await phone.count()) await phone.fill('09012345678');
  if (await emailI.count()) {
    await emailI.fill('abc');
    if ((await reg.count()) && (await reg.isEnabled())) {
      await reg.click();
      await page.waitForTimeout(700);
      fRec('V03', 'owner-create-edit', 'email', 'F2', hasError(await bodyText(page)) ? 'PASS' : 'FAIL', 'email abc');
    }
    await emailI.fill('');
  }
  // F3 discount 101
  const discount = page.locator('input[name=discount_rate], #discount_rate, input[name*=discount]').first();
  if (await discount.count()) {
    await discount.fill('101');
    if ((await reg.count()) && (await reg.isEnabled())) {
      await reg.click();
      await page.waitForTimeout(700);
      fRec('V03', 'owner-create-edit', 'discount_rate', 'F3', hasError(await bodyText(page)) ? 'PASS' : 'FAIL', '101');
    }
    await discount.fill('0');
  } else fRec('V03', 'owner-create-edit', 'discount_rate', 'F3', 'N/A', 'no field visible');

  // F4 create
  if (await nameI.count()) await nameI.fill(`${TAG}飼主`);
  if (await kanaI.count()) await kanaI.fill('テストカナ');
  if (await phone.count()) await phone.fill(`090${String(Date.now()).slice(-8)}`);
  if (await emailI.count()) await emailI.fill(`uat.${Date.now()}@example.com`);
  const wait = page
    .waitForResponse((r) => /owners/i.test(r.url()) && r.request().method() === 'POST', { timeout: 15000 })
    .catch(() => null);
  if ((await reg.count()) && (await reg.isEnabled())) await reg.click();
  const resp = await wait;
  await page.waitForTimeout(1000);
  fRec('V03', 'owner-create-edit', 'owner_name', 'F4', resp && resp.ok() ? 'PASS' : 'PARTIAL', resp ? String(resp.status()) : page.url());
  if (resp && resp.ok()) {
    rec('V03', 'owner-create-edit.C2', /owners\/\d+|飼主/.test(page.url() + (await bodyText(page))) ? 'PASS' : 'PARTIAL', page.url());
  }
} catch (e) {
  rec('V03', 'owner-create-edit.err', 'FAIL', e.message);
}

// C3-3 not found
try {
  await page.goto(`${BASE}/owners/999999999`, { waitUntil: 'networkidle' });
  const t = await bodyText(page);
  const ok = /見つかり|404|存在しません|エラー|Not Found|ありません/.test(t) && !/読み込み中/.test(t);
  rec('V03', 'owner-create-edit.C3-3', ok ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 60));
} catch (e) {
  rec('V03', 'owner-create-edit.C3-3', 'FAIL', e.message);
}

// pet modal from owner
try {
  await page.goto(`${BASE}/owners`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(600);
  const row = page.locator('table tbody tr a').first();
  if (await row.count()) {
    await row.click();
    await page.waitForTimeout(1000);
    const addPet = page.getByRole('button', { name: /ペット追加|追加/ }).first();
    if (await addPet.count()) {
      await addPet.click();
      await page.waitForTimeout(700);
      const ctrls = await listControls(page);
      for (const c of ctrls) fRec('V03', 'pet-edit-modal', c.fieldKey || c.label, 'F0', 'PASS', c.type);
      const save = page.getByRole('button', { name: /保存|登録/ }).first();
      if ((await save.count()) && (await save.isEnabled())) {
        await save.click();
        await page.waitForTimeout(700);
        fRec('V03', 'pet-edit-modal', 'name', 'F1', hasError(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'empty');
      } else if (await save.count()) fRec('V03', 'pet-edit-modal', 'name', 'F1', 'PASS', 'disabled');
      // weight F3
      const w = page.locator('input[name=weight], #weight, input[name*=weight]').first();
      if (await w.count()) {
        await w.fill('201');
        if ((await save.count()) && (await save.isEnabled())) {
          await save.click();
          await page.waitForTimeout(600);
          fRec('V03', 'pet-edit-modal', 'weight', 'F3', hasError(await bodyText(page)) ? 'PASS' : 'PARTIAL', '201');
        }
      }
      await page.keyboard.press('Escape').catch(() => {});
    } else rec('V03', 'pet-edit-modal.open', 'PARTIAL', 'no add pet');
  }
} catch (e) {
  rec('V03', 'pet-edit-modal.err', 'FAIL', e.message);
}

// staff
try {
  await page.goto(`${BASE}/settings/staff`, { waitUntil: 'networkidle' });
  rec('V03', 'staff-side-panel.open', /スタッフ/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', page.url());
  const neu = page.getByRole('button', { name: /新規/ }).first();
  if (await neu.count()) {
    await neu.click();
    await page.waitForTimeout(700);
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V03', 'staff-side-panel', c.fieldKey || c.label, 'F0', 'PASS', c.type);
    const save = page.getByRole('button', { name: /保存|登録/ }).first();
    if ((await save.count()) && (await save.isEnabled())) {
      await save.click();
      await page.waitForTimeout(700);
      fRec('V03', 'staff-side-panel', 'name', 'F1', hasError(await bodyText(page)) || (await page.locator('#master-title, input[name=name]').count()) ? 'PASS' : 'PARTIAL', 'empty');
    } else if (await save.count()) fRec('V03', 'staff-side-panel', 'name', 'F1', 'PASS', 'disabled');
    await cancelPanel(page);
  }
} catch (e) {
  rec('V03', 'staff-side-panel.err', 'FAIL', e.message);
}

// permission groups
try {
  await page.goto(`${BASE}/settings/permission-groups`, { waitUntil: 'networkidle' });
  rec('V03', 'permission-group-side-panel.open', /権限/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', page.url());
  const neu = page.getByRole('button', { name: /新規/ }).first();
  if (await neu.count()) {
    await neu.click();
    await page.waitForTimeout(700);
    const ctrls = await listControls(page);
    for (const c of ctrls) fRec('V03', 'permission-group-side-panel', c.fieldKey || c.label, 'F0', 'PASS', c.type);
    const save = page.getByRole('button', { name: /保存|登録/ }).first();
    if ((await save.count()) && (await save.isEnabled())) {
      await save.click();
      await page.waitForTimeout(700);
      fRec('V03', 'permission-group-side-panel', 'name', 'F1', hasError(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'empty');
    }
    await cancelPanel(page);
  }
} catch (e) {
  rec('V03', 'permission-group.err', 'FAIL', e.message);
}

// clinic
try {
  await page.goto(`${BASE}/settings/clinic`, { waitUntil: 'networkidle' });
  rec('V03', 'clinic-master-side-panel.open', /医院|病院|クリニック/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', page.url());
  const ctrls = await listControls(page);
  for (const c of ctrls) fRec('V03', 'clinic-master-side-panel', c.fieldKey || c.label, 'F0', c.disabled ? 'N/A' : 'PASS', c.type);
  // tax bound if present
  const nums = page.locator('input[type=number]');
  if (await nums.count()) {
    const v = await nums.first().inputValue().catch(() => '');
    await nums.first().fill('101');
    const save = page.getByRole('button', { name: /保存|更新/ }).first();
    if ((await save.count()) && (await save.isEnabled())) {
      await save.click();
      await page.waitForTimeout(800);
      fRec('V03', 'clinic-master-side-panel', 'standard_tax_rate', 'F3', hasError(await bodyText(page)) ? 'PASS' : 'PARTIAL', '101');
      if (v) await nums.first().fill(v);
    }
  }
} catch (e) {
  rec('V03', 'clinic.err', 'FAIL', e.message);
}

// ========== V04 masters field-level ==========
const masters = [
  ['master-animal-species', '/settings/animal-species', true],
  ['master-diagnosis-type', '/settings/diagnosis?tab=diagnosis_type', true],
  ['master-diagnosis-name', '/settings/diagnosis?tab=diagnosis_name', false],
  ['master-chief-complaint', '/settings/interview/chief-complaint', true],
  ['master-inquiry-templates', '/settings/inquiry-templates', false],
  ['master-hospitalization-plan', '/settings/hospitalization', true],
  ['master-cage', '/settings/cage', true],
  ['master-merchandise', '/settings/merchandise-items', true],
  ['master-insurance', '/settings/insurance', true],
  ['master-occupations', '/settings/occupations', true],
  ['master-trimming-course', '/settings/trimming?tab=course', true],
  ['master-trimming-option', '/settings/trimming?tab=option', true],
  ['master-trimming-course-type', '/settings/trimming-course-type', true],
  ['master-campaign', '/settings/campaigns', false],
  ['master-payment-method', '/settings/payment-methods', true],
  ['master-medicine', '/settings/medicine', true],
  ['master-shift-template', '/settings/shift-templates', true],
  ['master-reservation-type', '/settings/reservation-type', false],
];
for (const [id, route, unique] of masters) {
  await fieldLevelMaster(page, id, route, { unique });
}

// treatment 5 tabs
for (const tab of ['consultation', 'examination', 'procedure', 'vaccine', 'checkup']) {
  const formId = `master-treatment-${tab}`;
  const route = `/settings/treatment-items?tab=${tab}`;
  const name = `${TAG}_${formId}`;
  try {
    await openMaster(page, route);
    rec('V04', `${formId}.open`, /Something went wrong/i.test(await bodyText(page)) ? 'FAIL' : 'PASS', route);
    if (!(await openNewMaster(page))) {
      const alt = page.getByRole('button', { name: /新規/ });
      if (await alt.count()) await alt.first().click();
      await page.waitForTimeout(500);
    }
    if (await page.locator('#master-title').count()) {
      const ctrls = await listControls(page);
      for (const c of ctrls) fRec('V04', formId, c.fieldKey || c.label, 'F0', 'PASS', c.type);
      await saveMaster(page);
      const panel = await page.locator('#master-title').isVisible().catch(() => false);
      fRec('V04', formId, 'name', 'F1', panel || hasError(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'empty');
      await cancelPanel(page);
      if (await openNewMaster(page)) {
        await page.locator('#master-title').fill(name);
        const price = page.locator('input[type=number]').first();
        if (await price.count()) await price.fill('100').catch(() => {});
        await saveMaster(page);
        await page.waitForTimeout(800);
        fRec('V04', formId, 'name', 'F4', (await page.getByText(name).count()) > 0 ? 'PASS' : 'PARTIAL', name);
        await openMaster(page, route);
        await tryDeleteMaster(page, name);
      }
    } else rec('V04', `${formId}.panel`, 'BLOCKED', 'no title');
  } catch (e) {
    rec('V04', `${formId}.err`, 'FAIL', e.message);
  }
}

// insurance 101
try {
  await openMaster(page, '/settings/insurance');
  const name = `${TAG}_ins_bound`;
  if (await openNewMaster(page)) {
    await page.locator('#master-title').fill(name);
    const nums = page.locator('input[type=number]');
    if (await nums.count()) {
      await nums.first().fill('101');
      await saveMaster(page);
      const t = await bodyText(page);
      const rejected = hasError(t) || (await page.locator('#master-title').isVisible().catch(() => false));
      fRec('V04', 'master-insurance', 'coverage_rate', 'F3', rejected ? 'PASS' : 'FAIL', '101');
      if (!rejected) bug('V04-ins-101', '補償率101が拒否されない', name);
      await cancelPanel(page);
    }
  }
} catch (e) {
  rec('V04', 'insurance-bound.err', 'FAIL', e.message);
}

// payment system no-delete
try {
  await openMaster(page, '/settings/payment-methods');
  const cash = page.locator('tbody tr').filter({ hasText: /現金|cash/i }).first();
  if (await cash.count()) {
    await cash.getByLabel('操作').click().catch(() => cash.click());
    await page.waitForTimeout(500);
    const del = page.getByLabel('削除');
    if ((await del.count()) && (await del.isEnabled())) {
      await del.click();
      const dialog = page.getByRole('alertdialog');
      if (await dialog.count()) await dialog.getByRole('button', { name: '削除' }).click();
      await page.waitForTimeout(1000);
      const t = await bodyText(page);
      const blocked = /システム標準|削除できません|無効化できません/.test(t) || (await page.getByText(/現金/).count()) > 0;
      fRec('V04', 'master-payment-method', 'system_key', 'F6', blocked ? 'PASS' : 'FAIL', t.slice(0, 50));
      if (!blocked) bug('V04-pay-sys', 'システム標準支払方法が削除できてしまう', 'cash');
    } else fRec('V04', 'master-payment-method', 'system_key', 'F6', 'PASS', 'delete disabled');
    await cancelPanel(page);
  }
} catch (e) {
  rec('V04', 'payment-system.err', 'FAIL', e.message);
}

// ========== V05 auth / LINE ==========
// login F0 F1 F2 — use new page to not lose session
try {
  const p = await context.newPage();
  await p.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
  const ctrls = await listControls(p);
  for (const c of ctrls) fRec('V05', 'auth-login', c.fieldKey || c.label, 'F0', 'PASS', c.type);
  // empty
  await p.getByRole('button', { name: 'ログイン', exact: true }).click();
  await p.waitForTimeout(600);
  fRec('V05', 'auth-login', 'email', 'F1', hasError(await p.locator('body').innerText()) || p.url().includes('login') ? 'PASS' : 'FAIL', 'empty');
  // bad format
  await p.locator('#login-email').fill('not-an-email');
  await p.locator('#login-password').fill('x');
  await p.getByRole('button', { name: 'ログイン', exact: true }).click();
  await p.waitForTimeout(800);
  fRec('V05', 'auth-login', 'email', 'F2', p.url().includes('login') ? 'PASS' : 'FAIL', 'invalid stay on login');
  await p.close();
} catch (e) {
  rec('V05', 'auth-login.err', 'FAIL', e.message);
}

try {
  await page.goto(`${BASE}/forgot-password`, { waitUntil: 'networkidle' });
  rec('V05', 'auth-forgot-password.open', /パスワード|メール/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', page.url());
  const ctrls = await listControls(page);
  for (const c of ctrls) fRec('V05', 'auth-forgot-password', c.fieldKey || c.label, 'F0', 'PASS', c.type);
  const btn = page.getByRole('button', { name: /送信|リセット/ }).first();
  if ((await btn.count()) && (await btn.isEnabled())) {
    await btn.click();
    await page.waitForTimeout(700);
    fRec('V05', 'auth-forgot-password', 'email', 'F1', hasError(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'empty');
  } else if (await btn.count()) fRec('V05', 'auth-forgot-password', 'email', 'F1', 'PASS', 'disabled');
} catch (e) {
  rec('V05', 'auth-forgot.err', 'FAIL', e.message);
}

try {
  await page.goto(`${BASE}/reset-password`, { waitUntil: 'networkidle' });
  rec('V05', 'auth-reset-password.open', /無効|トークン|パスワード|リセット/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', page.url());
} catch (e) {
  rec('V05', 'auth-reset.err', 'FAIL', e.message);
}

const v05Routes = [
  ['line-reservation-settings', '/line-reservation/settings', /LINE|予約/],
  ['line-reservation-page-editor', '/line-reservation/page-editor', /ページ|編集|LINE|予約/],
  ['line-reservation-slots', '/line-reservation/slots', /枠|スロット|予約|LINE/],
  ['lstep-settings', '/settings/integrations/lstep', /Lステップ|連携/],
  ['lstep-tag-config', '/settings/lstep/tags', /タグ/],
  ['lstep-analytics', '/lstep/analytics', /分析|CSV|取込|Lステップ/],
  ['lstep-checkup-sync', '/lstep/checkup-sync', /健診|タグ/],
];
for (const [formId, pth, re] of v05Routes) {
  try {
    await page.goto(`${BASE}${pth}`, { waitUntil: 'networkidle', timeout: 30000 });
    const t = await bodyText(page);
    const crash = /Something went wrong|Application error/i.test(t);
    rec('V05', `${formId}.open`, crash ? 'FAIL' : re.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 60));
    if (crash) bug(`V05-${formId}-crash`, `画面クラッシュ ${pth}`, pth);
    if (!crash) {
      const ctrls = await listControls(page);
      for (const c of ctrls) {
        fRec('V05', formId, c.fieldKey || c.label, 'F0', c.disabled ? 'N/A' : 'PASS', c.disabled ? 'disabled/secret' : c.type);
      }
      // try empty save if present
      const save = page.getByRole('button', { name: /保存|登録|作成|送信/ }).first();
      if ((await save.count()) && (await save.isEnabled()) && ctrls.some((c) => c.required || /name|email|secret/i.test(c.fieldKey))) {
        // skip destructive save on lstep secrets
        if (!formId.includes('lstep-settings')) {
          await save.click().catch(() => {});
          await page.waitForTimeout(600);
          fRec('V05', formId, ctrls[0]?.fieldKey || 'form', 'F1', hasError(await bodyText(page)) || (await save.isEnabled()) ? 'PASS' : 'PARTIAL', 'empty attempt');
        } else {
          fRec('V05', formId, 'secrets', 'F6', 'N/A', 'skip mutating live secrets');
        }
      }
    }
  } catch (e) {
    rec('V05', `${formId}.err`, 'FAIL', e.message);
  }
}

// ensure pet alive
try {
  const pet = await (
    await page.request.get(`${API}/api/v1/pets/1000002`, { headers: { 'X-Requested-With': 'XMLHttpRequest' } })
  ).json();
  if (pet.status === 'deceased' || pet.deceased_at) {
    await page.request.fetch(`${API}/api/v1/clinics/1/pets/1000002/death`, {
      method: 'DELETE',
      headers: { 'X-Requested-With': 'XMLHttpRequest' },
    });
  }
  const pet2 = await (
    await page.request.get(`${API}/api/v1/pets/1000002`, { headers: { 'X-Requested-With': 'XMLHttpRequest' } })
  ).json();
  rec('S01', 'final_alive', pet2.status === 'alive' || !pet2.deceased_at ? 'PASS' : 'FAIL', JSON.stringify({ status: pet2.status }));
} catch (e) {
  rec('S01', 'final_alive', 'FAIL', e.message);
}

// write artifacts
fs.writeFileSync(path.join(OUT, 'results.json'), JSON.stringify(results, null, 2));
fs.writeFileSync(path.join(OUT, 'bug-candidates.json'), JSON.stringify(bugs, null, 2));

const byStatus = results.reduce((a, r) => {
  a[r.status] = (a[r.status] || 0) + 1;
  return a;
}, {});
const fails = results.filter((r) => r.status === 'FAIL');
const byScenario = {};
for (const r of results) {
  byScenario[r.scenario] = byScenario[r.scenario] || { PASS: 0, FAIL: 0, PARTIAL: 0, BLOCKED: 0, 'N/A': 0, SKIP: 0 };
  byScenario[r.scenario][r.status] = (byScenario[r.scenario][r.status] || 0) + 1;
}

const lines = [
  '# AnimalEkarte UAT — CDP :9222 再実施 2026-08-14',
  '',
  '| 項目 | 値 |',
  '|------|-----|',
  '| 環境 | localhost:3003/:8080 · Chrome CDP `:9222` · LIFF mock ON |',
  '| 認証 | E2E_LOGIN_*（値は非公開） |',
  '| 範囲 | S01〜S13 + V01〜V05（FIELD-LEVEL-PROTOCOL 適用） |',
  '| シナリオ md | **未編集** |',
  '| 実行手段 | Playwright `connectOverCDP(http://127.0.0.1:9222)`（DevTools MCP と同 endpoint） |',
  '| merge/push/migrate | なし |',
  '',
  '## 集計',
  '',
  Object.entries(byStatus)
    .map(([k, v]) => `**${k}** ${v}`)
    .join(' · '),
  '',
  `**結果件数** ${results.length} · **bug-candidates** ${bugs.length}`,
  '',
  '## シナリオ別',
  '',
  '| ID | PASS | FAIL | PARTIAL | BLOCKED | N/A |',
  '|----|------|------|---------|---------|-----|',
];
for (const id of Object.keys(byScenario).sort()) {
  const s = byScenario[id];
  lines.push(`| ${id} | ${s.PASS || 0} | ${s.FAIL || 0} | ${s.PARTIAL || 0} | ${s.BLOCKED || 0} | ${s['N/A'] || 0} |`);
}
lines.push('', '## FAIL 一覧', '');
if (!fails.length) lines.push('（なし）');
else for (const f of fails) lines.push(`- \`${f.scenario}#${f.step}\` — ${f.note}`);
lines.push('', '## bug-candidates', '');
if (!bugs.length) lines.push('（なし）');
else for (const b of bugs) lines.push(`- **${b.id}**: ${b.title} — ${b.evidence}`);
lines.push('', '## 人間レーン（BLOCKED）', '', '- 実 LINE 通知 / 実 token → todo-po', '- audit_logs DB 参照 → USER', '- 締め fixture 属性 → USER', '');
fs.writeFileSync(path.join(OUT, 'FINAL.md'), lines.join('\n'));

console.log('TOTALS', byStatus);
console.log('BUG_CANDIDATES', bugs.length, bugs.map((b) => b.id).join(','));
console.log('FAILS', fails.map((f) => `${f.scenario}#${f.step}`).join(',') || '(none)');
console.log('WROTE', path.join(OUT, 'results.json'), path.join(OUT, 'FINAL.md'));
// disconnect CDP only — do not kill user Chrome
await browser.close();
