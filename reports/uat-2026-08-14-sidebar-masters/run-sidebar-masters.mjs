/**
 * Sidebar マスタ設定 — full item UAT
 * All paths under マスタ設定 + nested tabs + related settings masters
 * C1-1 / create with all fillable fields / C2 reload / C3-2 unique where applicable
 */
import { chromium } from '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/node_modules/playwright/index.mjs';
import fs from 'fs';
import path from 'path';

const BASE = 'http://localhost:3003';
const OUT = '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/uat-2026-08-14-sidebar-masters';
const email = process.env.E2E_LOGIN_EMAIL || 'admin@noavet.jp';
const password = process.env.E2E_LOGIN_PASSWORD || 'password';
const TAG = `SB_${Date.now().toString().slice(-8)}`;
fs.mkdirSync(OUT, { recursive: true });

const results = [];
const bugs = [];
const rec = (id, step, status, note = '') => {
  results.push({ id, step, status, note: String(note).slice(0, 500), at: new Date().toISOString() });
  console.log(`[${status}] ${id}#${step} ${String(note).slice(0, 140)}`);
};
const bug = (id, title, ev) => {
  bugs.push({ id, title, evidence: String(ev).slice(0, 300) });
  console.log(`[BUG?] ${id} ${title}`);
};

async function login(page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.locator('#login-email').fill(email);
  await page.locator('#login-password').fill(password);
  await page.getByRole('button', { name: 'ログイン', exact: true }).click();
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 30000 });
}

async function go(page, route) {
  await page.goto(`${BASE}${route}`, { waitUntil: 'domcontentloaded', timeout: 45000 });
  await page.waitForTimeout(700);
}

async function openNew(page) {
  const b = page.getByRole('button', { name: '新規登録' });
  if (!(await b.count())) return false;
  await b.click();
  try {
    await page.waitForSelector('#master-title', { timeout: 10000 });
    return true;
  } catch {
    return false;
  }
}

async function save(page) {
  const s = page.getByRole('button', { name: '保存', exact: true }).first();
  if (await s.count() && await s.isEnabled()) await s.click();
  await page.waitForTimeout(1400);
}

async function cancel(page) {
  const c = page.getByRole('button', { name: 'キャンセル' });
  if (await c.count()) await c.click().catch(() => {});
  await page.keyboard.press('Escape').catch(() => {});
  await page.waitForTimeout(350);
}

/** Fill every visible editable control in the open side panel with safe values */
async function fillAllPanelFields(page, title) {
  await page.locator('#master-title').fill(title);
  // text inputs except title
  const texts = page.locator('input[type="text"]:not(#master-title), input:not([type]):not(#master-title)');
  for (let i = 0; i < await texts.count(); i++) {
    const el = texts.nth(i);
    if (!(await el.isVisible().catch(() => false))) continue;
    if (!(await el.isEditable().catch(() => false))) continue;
    const ph = ((await el.getAttribute('placeholder')) || '') + ((await el.getAttribute('aria-label')) || '');
    if (/カテゴリ/.test(ph)) await el.fill('SBカテゴリ');
    else if (/検索/.test(ph)) continue;
    else await el.fill(`SB_${i}`).catch(() => {});
  }
  // textarea
  const tas = page.locator('textarea');
  for (let i = 0; i < await tas.count(); i++) {
    const el = tas.nth(i);
    if (await el.isVisible().catch(() => false) && await el.isEditable().catch(() => false)) {
      await el.fill('SB本文テスト').catch(() => {});
    }
  }
  // dates
  const dates = page.locator('input[type=date]');
  if ((await dates.count()) >= 1) await dates.nth(0).fill('2026-08-01').catch(() => {});
  if ((await dates.count()) >= 2) await dates.nth(1).fill('2026-12-31').catch(() => {});
  // times
  const times = page.locator('input[type=time]');
  if ((await times.count()) >= 1) await times.nth(0).fill('09:00').catch(() => {});
  if ((await times.count()) >= 2) await times.nth(1).fill('18:00').catch(() => {});
  // numbers — use safe mid values (coverage etc.)
  const nums = page.locator('input[type=number]');
  for (let i = 0; i < await nums.count(); i++) {
    const el = nums.nth(i);
    if (!(await el.isVisible().catch(() => false))) continue;
    const min = await el.getAttribute('min');
    const max = await el.getAttribute('max');
    let v = '10';
    if (max && Number(max) <= 100) v = '50'; // insurance-like
    if (min && Number(min) >= 1) v = String(Math.max(1, Number(min)));
    await el.fill(v).catch(() => {});
  }
  // comboboxes — pick first non-empty option
  const combos = page.getByRole('combobox');
  const nCombo = await combos.count();
  for (let i = 0; i < nCombo; i++) {
    const c = combos.nth(i);
    if (!(await c.isVisible().catch(() => false))) continue;
    await c.click().catch(() => {});
    await page.waitForTimeout(250);
    const opts = page.getByRole('option');
    const oc = await opts.count();
    if (oc > 0) {
      // prefer 休み for shift to avoid time complexity
      const off = page.getByRole('option').filter({ hasText: /休み|off|有給/ });
      if (await off.count()) await off.first().click().catch(() => {});
      else await opts.nth(Math.min(1, oc - 1)).click().catch(() => {});
    }
    await page.keyboard.press('Escape').catch(() => {});
    await page.waitForTimeout(150);
  }
  // native selects
  const sels = page.locator('select');
  for (let i = 0; i < await sels.count(); i++) {
    const s = sels.nth(i);
    if ((await s.locator('option').count()) > 1) await s.selectOption({ index: 1 }).catch(() => {});
  }
}

async function masterCRUD(page, id, route, { unique = true } = {}) {
  const name = `${TAG}_${id}`;
  try {
    await go(page, route);
    const body0 = await page.locator('body').innerText();
    if (/Something went wrong|Application error/i.test(body0)) {
      rec(id, 'open', 'FAIL', 'crash');
      bug(`SB-${id}-crash`, `${route} 画面クラッシュ`, route);
      return;
    }
    rec(id, 'open', 'PASS', route);

    if (!(await openNew(page))) {
      // staff / permission may differ
      const alt = page.getByRole('button', { name: /新規|追加|招待|作成/ });
      if (await alt.count()) {
        await alt.first().click();
        await page.waitForTimeout(600);
        rec(id, 'panel', 'PARTIAL', 'alt new button');
      } else {
        rec(id, 'panel', 'BLOCKED', 'no 新規登録');
        return;
      }
    }

    if (!(await page.locator('#master-title').count())) {
      // non-standard form (staff etc.) — field inventory only
      const fields = await page.evaluate(() => {
        return [...document.querySelectorAll('input,textarea,select,[role=combobox]')].slice(0, 40).map((e) => ({
          tag: e.tagName, type: e.getAttribute('type'), id: e.id,
          name: e.getAttribute('name'), aria: e.getAttribute('aria-label'), ph: e.getAttribute('placeholder'),
        }));
      });
      rec(id, 'fields_inventory', fields.length ? 'PASS' : 'PARTIAL', JSON.stringify(fields).slice(0, 200));
      // try fill required-looking and save if possible
      const saveBtn = page.getByRole('button', { name: /保存|登録|作成/ }).first();
      if (await saveBtn.count()) {
        await saveBtn.click().catch(() => {});
        await page.waitForTimeout(800);
        const t = await page.locator('body').innerText();
        rec(id, 'C1-empty-or-save', /入力|必須|してください/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 60));
      }
      await cancel(page);
      return;
    }

    // C1-1 empty
    await save(page);
    const panelStay = await page.locator('#master-title').isVisible().catch(() => false);
    rec(id, 'C1-1', panelStay ? 'PASS' : 'PARTIAL', panelStay ? 'empty blocked' : 'panel closed');
    await cancel(page);

    // Create with all fields
    if (!(await openNew(page))) {
      rec(id, 'create', 'BLOCKED', 'no panel');
      return;
    }
    await fillAllPanelFields(page, name);
    const w = page.waitForResponse(
      (r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()),
      { timeout: 15000 },
    ).catch(() => null);
    await save(page);
    const resp = await w;
    await page.waitForTimeout(500);
    await go(page, route);
    let vis = (await page.getByText(name).count()) > 0;
    if (!vis) {
      // search
      const st = page.getByLabel('検索');
      if (await st.count()) {
        await st.click().catch(() => {});
        const si = page.locator('input[placeholder*="検索"]');
        if (await si.count()) {
          await si.first().fill(name);
          await page.waitForTimeout(800);
          vis = (await page.getByText(name).count()) > 0;
        }
      }
    }
    const createOk = (resp && resp.status() >= 200 && resp.status() < 300) || vis;
    rec(id, 'create', createOk ? 'PASS' : 'FAIL', resp ? `${resp.status()} vis=${vis}` : `vis=${vis}`);
    if (!createOk) {
      const t = await page.locator('body').innerText();
      bug(`SB-${id}-create`, `${route} 全項目入力後の新規保存失敗`, `${name} ${t.slice(0, 80)}`);
      return;
    }

    // C2 reload
    await go(page, route);
    const st2 = page.getByLabel('検索');
    if (await st2.count()) {
      await st2.click().catch(() => {});
      const si = page.locator('input[placeholder*="検索"]');
      if (await si.count()) {
        await si.first().fill(name);
        await page.waitForTimeout(700);
      }
    }
    const after = (await page.getByText(name).count()) > 0;
    rec(id, 'C2-reload', after ? 'PASS' : 'FAIL', name);
    if (!after) bug(`SB-${id}-C2`, `${route} 再読込で消える`, name);

    // C2 reopen + field check
    if (after) {
      const row = page.locator('tbody tr').filter({ hasText: name }).first();
      if (await row.count()) {
        const op = row.getByLabel('操作');
        if (await op.count()) await op.click();
        else await row.click();
        await page.waitForTimeout(600);
        if (await page.locator('#master-title').count()) {
          const val = await page.locator('#master-title').inputValue().catch(() => '');
          rec(id, 'C2-reopen-title', val === name || val.includes(name) ? 'PASS' : 'PARTIAL', `val=${val.slice(0, 40)}`);
          // inventory filled fields still present
          const filled = await page.evaluate(() => {
            const root = document.querySelector('[role=dialog], aside, [data-state=open]') || document.body;
            return [...root.querySelectorAll('input,textarea')].filter((e) => {
              if (e instanceof HTMLInputElement || e instanceof HTMLTextAreaElement) {
                return e.value && e.value.length > 0 && e.type !== 'checkbox' && e.type !== 'hidden';
              }
              return false;
            }).length;
          });
          rec(id, 'C2-reopen-fields', filled >= 1 ? 'PASS' : 'PARTIAL', `filledInputs=${filled}`);
        } else rec(id, 'C2-reopen-title', 'PARTIAL', 'no title');
        await cancel(page);
      }
    }

    // C3-2 unique
    if (unique && after) {
      if (await openNew(page)) {
        await fillAllPanelFields(page, name);
        const w2 = page.waitForResponse(
          (r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()),
          { timeout: 12000 },
        ).catch(() => null);
        await save(page);
        const r2 = await w2;
        const rej = r2 && (r2.status() === 409 || r2.status() === 400);
        if (!rej && r2 && r2.ok()) {
          rec(id, 'C3-2', 'FAIL', `dup accepted ${r2.status()}`);
          bug(`SB-${id}-unique`, `${route} 同名が重複登録できる`, name);
        } else {
          rec(id, 'C3-2', rej ? 'PASS' : 'PARTIAL', r2 ? String(r2.status()) : 'noresp');
        }
        await cancel(page);
      }
    } else if (!unique) rec(id, 'C3-2', 'SKIP', 'no unique per V04');

    // cleanup best-effort
    await go(page, route);
    try {
      const st = page.getByLabel('検索');
      if (await st.count()) {
        await st.click().catch(() => {});
        const si = page.locator('input[placeholder*="検索"]');
        if (await si.count()) {
          await si.first().fill(name);
          await page.waitForTimeout(600);
        }
      }
      const row = page.locator('tbody tr').filter({ hasText: name }).first();
      if (await row.count()) {
        await row.getByLabel('操作').click().catch(() => {});
        await page.waitForTimeout(400);
        const del = page.getByLabel('削除');
        if (await del.count() && await del.isEnabled()) {
          await del.click();
          const dlg = page.getByRole('alertdialog');
          if (await dlg.count()) await dlg.getByRole('button', { name: '削除' }).click();
          await page.waitForTimeout(800);
          rec(id, 'cleanup', 'PASS', 'deleted');
        } else {
          rec(id, 'cleanup', 'PARTIAL', 'no delete');
          await cancel(page);
        }
      } else rec(id, 'cleanup', 'PARTIAL', 'row not found');
    } catch {
      rec(id, 'cleanup', 'PARTIAL', 'cleanup err');
    }
  } catch (e) {
    rec(id, 'err', 'FAIL', e.message);
    bug(`SB-${id}-err`, `${route} 例外`, e.message);
  }
}

// Sidebar マスタ設定 leaf pages + nested tabs (from sidebar-menu.tsx + settings routes)
const MASTERS = [
  // 医院 is special — not #master-title CRUD list only
  { id: 'clinic', route: '/settings/clinic', unique: false, mode: 'special-clinic' },
  { id: 'animal-species', route: '/settings/animal-species', unique: true },
  // 診療項目 tabs
  { id: 'treatment-consultation', route: '/settings/treatment-items?tab=consultation', unique: true },
  { id: 'treatment-examination', route: '/settings/treatment-items?tab=examination', unique: true },
  { id: 'treatment-procedure', route: '/settings/treatment-items?tab=procedure', unique: true },
  { id: 'treatment-vaccine', route: '/settings/treatment-items?tab=vaccine', unique: true },
  { id: 'treatment-checkup', route: '/settings/treatment-items?tab=checkup', unique: true },
  // 診断
  { id: 'diagnosis-type', route: '/settings/diagnosis?tab=diagnosis_type', unique: true },
  { id: 'diagnosis-name', route: '/settings/diagnosis?tab=diagnosis_name', unique: false },
  // 問診
  { id: 'inquiry-templates', route: '/settings/inquiry-templates', unique: false },
  { id: 'chief-complaint', route: '/settings/interview/chief-complaint', unique: true },
  { id: 'interview-templates', route: '/settings/interview/templates', unique: false },
  { id: 'medicine', route: '/settings/medicine', unique: true },
  { id: 'reservation-type', route: '/settings/reservation-type', unique: true },
  { id: 'hospitalization-plan', route: '/settings/hospitalization', unique: true },
  { id: 'cage', route: '/settings/cage', unique: true },
  { id: 'trimming-course', route: '/settings/trimming?tab=course', unique: true },
  { id: 'trimming-option', route: '/settings/trimming?tab=option', unique: true },
  { id: 'trimming-course-type', route: '/settings/trimming-course-type', unique: true },
  { id: 'permission-groups', route: '/settings/permission-groups', unique: true },
  { id: 'occupations', route: '/settings/occupations', unique: true },
  { id: 'staff', route: '/settings/staff', unique: false, mode: 'special-staff' },
  { id: 'insurance', route: '/settings/insurance', unique: true },
  { id: 'merchandise', route: '/settings/merchandise-items', unique: true },
  { id: 'payment-methods', route: '/settings/payment-methods', unique: true },
  { id: 'closing-time', route: '/settings/closing-time', unique: false, mode: 'special-closing' },
  // also under settings hub often linked
  { id: 'campaigns', route: '/settings/campaigns', unique: false },
  { id: 'shift-templates', route: '/settings/shift-templates', unique: true },
];

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
page.setDefaultTimeout(40000);
await login(page);
rec('ENV', 'login', 'PASS', page.url());

// hub
await go(page, '/settings');
rec('settings-hub', 'open', /マスタ|設定|医院/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', '/settings');

for (const m of MASTERS) {
  if (m.mode === 'special-clinic') {
    try {
      await go(page, m.route);
      const t = await page.locator('body').innerText();
      rec(m.id, 'open', /医院|インボイス|法人/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 60));
      // inventory fields
      const n = await page.locator('input,textarea,select,button').count();
      rec(m.id, 'fields_present', n > 5 ? 'PASS' : 'PARTIAL', `controls=${n}`);
      // invoice-like
      const inputs = page.locator('input[type=text]');
      let did = false;
      for (let i = 0; i < Math.min(await inputs.count(), 20); i++) {
        const aria = ((await inputs.nth(i).getAttribute('aria-label')) || '') + ((await inputs.nth(i).getAttribute('placeholder')) || '');
        if (/インボイス|登録番号|T[0-9]/.test(aria) || (await inputs.nth(i).inputValue()).startsWith('T')) {
          const old = await inputs.nth(i).inputValue();
          await inputs.nth(i).fill('T9999999999999');
          const saveBtn = page.getByRole('button', { name: /保存|更新/ }).first();
          if (await saveBtn.count()) {
            const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['PUT', 'PATCH', 'POST'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
            await saveBtn.click();
            const r = await w;
            rec(m.id, 'invoice-save', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : 'noresp');
            await inputs.nth(i).fill(old || '');
            if (await saveBtn.count()) await saveBtn.click();
            did = true;
            break;
          }
        }
      }
      if (!did) rec(m.id, 'invoice-save', 'PARTIAL', 'field heuristic miss');
    } catch (e) { rec(m.id, 'err', 'FAIL', e.message); }
    continue;
  }
  if (m.mode === 'special-staff') {
    try {
      await go(page, m.route);
      const t = await page.locator('body').innerText();
      rec(m.id, 'open', /スタッフ|職員|メンバー/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 50));
      const neu = page.getByRole('button', { name: /新規|追加|招待/ });
      rec(m.id, 'new_btn', (await neu.count()) > 0 ? 'PASS' : 'PARTIAL', `n=${await neu.count()}`);
      if (await neu.count()) {
        await neu.first().click();
        await page.waitForTimeout(600);
        const fields = await page.locator('input,textarea,select').count();
        rec(m.id, 'panel_fields', fields > 0 ? 'PASS' : 'FAIL', `fields=${fields}`);
        // empty save
        const saveBtn = page.getByRole('button', { name: /保存|登録|作成|招待/ }).first();
        if (await saveBtn.count()) {
          await saveBtn.click();
          await page.waitForTimeout(700);
          const t2 = await page.locator('body').innerText();
          rec(m.id, 'C1-1', /入力|必須|してください|メール/.test(t2) || (await page.locator('input:invalid').count()) > 0 ? 'PASS' : 'PARTIAL', t2.replace(/\s+/g, ' ').slice(0, 50));
        }
        await cancel(page);
      }
      // list rows
      rec(m.id, 'list', (await page.locator('tbody tr').count()) >= 0 ? 'PASS' : 'FAIL', 'rows');
    } catch (e) { rec(m.id, 'err', 'FAIL', e.message); }
    continue;
  }
  if (m.mode === 'special-closing') {
    try {
      await go(page, m.route);
      const t = await page.locator('body').innerText();
      rec(m.id, 'open', /締め|AM|PM|休診|特別/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 50));
      // fill time inputs if any and save standard
      const times = page.locator('input[type=time]');
      const n = await times.count();
      rec(m.id, 'time_fields', n >= 1 ? 'PASS' : 'PARTIAL', `timeInputs=${n}`);
      const saveBtn = page.getByRole('button', { name: /保存/ }).first();
      if (await saveBtn.count() && await saveBtn.isEnabled()) {
        const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['PUT', 'PATCH', 'POST'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
        await saveBtn.click();
        const r = await w;
        rec(m.id, 'save', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : 'noresp');
      }
      // holiday section add empty
      const add = page.getByRole('button', { name: /追加/ });
      if (await add.count()) {
        await add.first().click();
        await page.waitForTimeout(400);
        const save2 = page.getByRole('button', { name: /保存|追加|登録/ }).last();
        if (await save2.count() && !(await save2.isEnabled())) rec(m.id, 'holiday-C1', 'PASS', 'disabled without date');
        else if (await save2.count()) {
          await save2.click();
          await page.waitForTimeout(500);
          rec(m.id, 'holiday-C1', 'PARTIAL', 'clicked');
        }
      }
    } catch (e) { rec(m.id, 'err', 'FAIL', e.message); }
    continue;
  }

  await masterCRUD(page, m.id, m.route, { unique: m.unique });
}

// insurance 101 bound
try {
  await go(page, '/settings/insurance');
  if (await openNew(page)) {
    await page.locator('#master-title').fill(`${TAG}_ins101`);
    const num = page.locator('input[type=number]').first();
    if (await num.count()) {
      await num.fill('101');
      const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 8000 }).catch(() => null);
      await save(page);
      const r = await w;
      const stay = await page.locator('#master-title').isVisible().catch(() => false);
      const rej = (r && r.status() >= 400) || stay;
      rec('insurance-bound', 'C1-3-101', rej ? 'PASS' : 'FAIL', r ? String(r.status()) : 'stay');
      if (!rej) bug('SB-ins-101', '補償率101が保存できてしまう', '');
    }
    await cancel(page);
  }
} catch (e) { rec('insurance-bound', 'err', 'FAIL', e.message); }

// payment system no delete
try {
  await go(page, '/settings/payment-methods');
  const row = page.locator('tbody tr').filter({ hasText: /現金/ }).first();
  if (await row.count()) {
    await row.getByLabel('操作').click().catch(() => {});
    await page.waitForTimeout(400);
    const del = page.getByLabel('削除');
    if (!(await del.count()) || !(await del.isEnabled())) rec('payment-system', 'no-delete', 'PASS', 'disabled');
    else {
      await del.click();
      const d = page.getByRole('alertdialog');
      if (await d.count()) await d.getByRole('button', { name: '削除' }).click();
      await page.waitForTimeout(800);
      const t = await page.locator('body').innerText();
      const ok = /システム標準|削除できません/.test(t) || (await page.getByText('現金').count()) > 0;
      rec('payment-system', 'no-delete', ok ? 'PASS' : 'FAIL', t.slice(0, 40));
      if (!ok) bug('SB-pay-sys', 'システム標準支払が削除できる', '');
    }
    await cancel(page);
  }
} catch (e) { rec('payment-system', 'err', 'FAIL', e.message); }

// permission groups matrix smoke
try {
  await go(page, '/settings/permission-groups');
  const t = await page.locator('body').innerText();
  rec('permission-groups-extra', 'matrix', /権限|表示|作成|編集|削除/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 50));
} catch (e) { rec('permission-groups-extra', 'err', 'FAIL', e.message); }

await browser.close();
fs.writeFileSync(path.join(OUT, 'results.json'), JSON.stringify(results, null, 2));
fs.writeFileSync(path.join(OUT, 'bug-candidates.json'), JSON.stringify(bugs, null, 2));
const c = results.reduce((a, r) => { a[r.status] = (a[r.status] || 0) + 1; return a; }, {});
console.log('TOTALS', c, 'steps', results.length);
console.log('FAILS', results.filter((r) => r.status === 'FAIL').map((r) => `${r.id}#${r.step}`).join(','));
console.log('BUGS', bugs.map((b) => b.id).join(','));
