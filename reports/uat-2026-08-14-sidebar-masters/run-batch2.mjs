/**
 * Sidebar masters batch2 — remaining + retests (safe field fill)
 */
import { chromium } from '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/node_modules/playwright/index.mjs';
import fs from 'fs';
import path from 'path';

const BASE = 'http://localhost:3003';
const OUT = '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/uat-2026-08-14-sidebar-masters';
const email = process.env.E2E_LOGIN_EMAIL || 'admin@noavet.jp';
const password = process.env.E2E_LOGIN_PASSWORD || 'password';
const TAG = `SB2_${Date.now().toString().slice(-8)}`;
const results = [];
const bugs = [];
const rec = (id, step, status, note = '') => {
  results.push({ id, step, status, note: String(note).slice(0, 400) });
  console.log(`[${status}] ${id}#${step} ${String(note).slice(0, 120)}`);
};
const bug = (id, title, ev) => { bugs.push({ id, title, evidence: String(ev).slice(0, 200) }); console.log(`[BUG?] ${id}`); };

async function login(page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.locator('#login-email').fill(email);
  await page.locator('#login-password').fill(password);
  await page.getByRole('button', { name: 'ログイン', exact: true }).click();
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 30000 });
}
async function go(page, route) {
  await page.goto(`${BASE}${route}`, { waitUntil: 'domcontentloaded', timeout: 40000 });
  await page.waitForTimeout(600);
}
async function openNew(page) {
  const b = page.getByRole('button', { name: '新規登録' });
  if (!(await b.count())) return false;
  await b.click();
  try { await page.waitForSelector('#master-title', { timeout: 8000 }); return true; } catch { return false; }
}
async function save(page) {
  await page.getByRole('button', { name: '保存', exact: true }).first().click().catch(() => {});
  await page.waitForTimeout(1200);
}
async function cancel(page) {
  const c = page.getByRole('button', { name: 'キャンセル' });
  if (await c.count()) await c.click().catch(() => {});
  await page.keyboard.press('Escape').catch(() => {});
  await page.waitForTimeout(300);
}
async function searchName(page, name) {
  const st = page.getByLabel('検索');
  if (await st.count()) {
    await st.click().catch(() => {});
    const si = page.locator('input[placeholder*="検索"]');
    if (await si.count()) {
      await si.first().fill(name);
      await page.waitForTimeout(700);
    }
  }
}
async function fillSafe(page, name, kind) {
  await page.locator('#master-title').fill(name);
  if (kind === 'campaigns') {
    const d = page.locator('input[type=date]');
    if ((await d.count()) >= 2) { await d.nth(0).fill('2026-08-01'); await d.nth(1).fill('2026-12-31'); }
    const n = page.locator('input[type=number]').first();
    if (await n.count()) await n.fill('10').catch(() => {});
  } else if (kind === 'inquiry' || kind === 'interview') {
    const cat = page.getByLabel('カテゴリ');
    if (await cat.count()) await cat.fill('SBカテゴリ');
    else {
      const t = page.locator('input[placeholder*="カテゴリ"]');
      if (await t.count()) await t.first().fill('SBカテゴリ');
    }
    const ta = page.locator('textarea').first();
    if (await ta.count()) await ta.fill('本文').catch(() => {});
  } else if (kind === 'medicine') {
    const price = page.getByLabel(/単価|価格/).or(page.locator('input[type=number]')).first();
    if (await price.count()) await price.fill('100').catch(() => {});
  } else if (kind === 'treatment') {
    const n = page.locator('input[type=number]').first();
    if (await n.count()) await n.fill('100').catch(() => {});
  } else if (kind === 'shift') {
    const c = page.getByRole('combobox').first();
    if (await c.count()) {
      await c.click();
      const off = page.getByRole('option').filter({ hasText: /休み|off|有給/ });
      if (await off.count()) await off.first().click();
      else if (await page.getByRole('option').count()) await page.getByRole('option').first().click();
    }
    const times = page.locator('input[type=time]');
    if ((await times.count()) >= 2) {
      await times.nth(0).fill('09:00');
      await times.nth(1).fill('18:00');
    }
  } else if (kind === 'diagnosis-name') {
    const c = page.getByRole('combobox').first();
    if (await c.count()) {
      await c.click();
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click();
    }
  } else if (kind === 'insurance') {
    const n = page.locator('input[type=number]').first();
    if (await n.count()) await n.fill('80').catch(() => {});
  } else if (kind === 'cage') {
    const c = page.getByRole('combobox');
    for (let i = 0; i < Math.min(await c.count(), 2); i++) {
      await c.nth(i).click().catch(() => {});
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
    }
  } else {
    // generic: category + first combo only
    const cat = page.getByLabel('カテゴリ');
    if (await cat.count()) await cat.fill('SBカテゴリ').catch(() => {});
    const c = page.getByRole('combobox').first();
    if (await c.count()) {
      await c.click().catch(() => {});
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
    }
  }
}

async function flow(page, id, route, { unique = true, kind = 'generic' } = {}) {
  const name = `${TAG}_${id}`;
  try {
    await go(page, route);
    if (/Something went wrong/i.test(await page.locator('body').innerText())) {
      rec(id, 'open', 'FAIL', 'crash'); bug(`SB-${id}-crash`, `${route} crash`, route); return;
    }
    rec(id, 'open', 'PASS', route);
    if (!(await openNew(page))) { rec(id, 'panel', 'BLOCKED', 'no new'); return; }
    await save(page);
    rec(id, 'C1-1', (await page.locator('#master-title').isVisible().catch(() => false)) ? 'PASS' : 'PARTIAL', 'empty');
    await cancel(page);
    if (!(await openNew(page))) return;
    await fillSafe(page, name, kind);
    const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 15000 }).catch(() => null);
    await save(page);
    const r = await w;
    await go(page, route);
    await searchName(page, name);
    const vis = (await page.getByText(name).count()) > 0;
    const ok = (r && r.status() >= 200 && r.status() < 300) || vis;
    rec(id, 'create', ok ? 'PASS' : 'FAIL', r ? `${r.status()} vis=${vis}` : `vis=${vis}`);
    if (!ok) {
      let body = '';
      try { body = r ? (await r.text()).slice(0, 120) : ''; } catch {}
      bug(`SB-${id}-create`, `${route} 新規失敗`, body || name);
      return;
    }
    rec(id, 'C2-reload', vis ? 'PASS' : 'FAIL', name);
    if (!vis) bug(`SB-${id}-C2`, `${route} 一覧に出ない`, name);
    if (unique && vis) {
      if (await openNew(page)) {
        await fillSafe(page, name, kind);
        const w2 = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 12000 }).catch(() => null);
        await save(page);
        const r2 = await w2;
        const rej = r2 && (r2.status() === 409 || r2.status() >= 400);
        if (!rej && r2 && r2.ok()) {
          rec(id, 'C3-2', 'FAIL', 'dup ok');
          bug(`SB-${id}-unique`, `${route} 同名重複可`, name);
        } else rec(id, 'C3-2', rej ? 'PASS' : 'PARTIAL', r2 ? String(r2.status()) : 'noresp');
        await cancel(page);
      }
    } else if (!unique) rec(id, 'C3-2', 'SKIP', 'no unique');
  } catch (e) { rec(id, 'err', 'FAIL', e.message); }
}

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
page.setDefaultTimeout(35000);
await login(page);
rec('ENV', 'login', 'PASS', '');

// retest treatments + medicine
for (const tab of ['consultation', 'examination', 'procedure', 'vaccine', 'checkup']) {
  await flow(page, `treatment-${tab}`, `/settings/treatment-items?tab=${tab}`, { unique: true, kind: 'treatment' });
}
await flow(page, 'medicine', '/settings/medicine', { unique: true, kind: 'medicine' });

// remaining from timeout
const rest = [
  ['cage', '/settings/cage', true, 'cage'],
  ['trimming-course', '/settings/trimming?tab=course', true, 'generic'],
  ['trimming-option', '/settings/trimming?tab=option', true, 'generic'],
  ['trimming-course-type', '/settings/trimming-course-type', true, 'generic'],
  ['permission-groups', '/settings/permission-groups', true, 'generic'],
  ['occupations', '/settings/occupations', true, 'generic'],
  ['insurance', '/settings/insurance', true, 'insurance'],
  ['merchandise', '/settings/merchandise-items', true, 'generic'],
  ['payment-methods', '/settings/payment-methods', true, 'generic'],
  ['campaigns', '/settings/campaigns', false, 'campaigns'],
  ['shift-templates', '/settings/shift-templates', true, 'shift'],
  ['inquiry-templates', '/settings/inquiry-templates', false, 'inquiry'],
  ['interview-templates', '/settings/interview/templates', false, 'interview'],
];
for (const [id, route, u, k] of rest) await flow(page, id, route, { unique: u, kind: k });

// staff special
try {
  await go(page, '/settings/staff');
  rec('staff', 'open', /スタッフ/.test(await page.locator('body').innerText()) ? 'PASS' : 'PARTIAL', '');
  const neu = page.getByRole('button', { name: /新規|追加|招待/ });
  rec('staff', 'new_btn', (await neu.count()) > 0 ? 'PASS' : 'PARTIAL', `n=${await neu.count()}`);
  if (await neu.count()) {
    await neu.first().click();
    await page.waitForTimeout(500);
    const fields = await page.locator('input:visible,textarea:visible,select:visible').count();
    rec('staff', 'panel_fields', fields > 0 ? 'PASS' : 'FAIL', `fields=${fields}`);
    // inventory each field type
    const inv = await page.evaluate(() => [...document.querySelectorAll('input,textarea,select,[role=combobox]')].filter(e => {
      const s = window.getComputedStyle(e);
      return s && s.display !== 'none' && s.visibility !== 'hidden';
    }).map(e => ({ tag: e.tagName, type: e.getAttribute('type'), aria: e.getAttribute('aria-label'), ph: e.getAttribute('placeholder'), id: e.id })).slice(0, 25));
    rec('staff', 'field_list', inv.length ? 'PASS' : 'PARTIAL', JSON.stringify(inv).slice(0, 250));
    const saveBtn = page.getByRole('button', { name: /保存|登録|作成|招待/ }).first();
    if (await saveBtn.count()) {
      await saveBtn.click();
      await page.waitForTimeout(600);
      const t = await page.locator('body').innerText();
      rec('staff', 'C1-1', /入力|必須|メール|してください/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 50));
    }
    await cancel(page);
  }
} catch (e) { rec('staff', 'err', 'FAIL', e.message); }

// closing already partial - quick save
try {
  await go(page, '/settings/closing-time');
  rec('closing-time', 'open', /締め|AM|PM/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', '');
  const saveBtn = page.getByRole('button', { name: /保存/ }).first();
  if (await saveBtn.count() && await saveBtn.isEnabled()) {
    const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['PUT', 'PATCH', 'POST'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
    await saveBtn.click();
    const r = await w;
    rec('closing-time', 'save', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : '');
  }
} catch (e) { rec('closing-time', 'err', 'FAIL', e.message); }

// insurance 101
try {
  await go(page, '/settings/insurance');
  if (await openNew(page)) {
    await page.locator('#master-title').fill(`${TAG}_i101`);
    const n = page.locator('input[type=number]').first();
    if (await n.count()) {
      await n.fill('101');
      const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 8000 }).catch(() => null);
      await save(page);
      const r = await w;
      const stay = await page.locator('#master-title').isVisible().catch(() => false);
      rec('insurance-bound', 'C1-3-101', (r && r.status() >= 400) || stay ? 'PASS' : 'FAIL', r ? String(r.status()) : 'stay');
      if (!((r && r.status() >= 400) || stay)) bug('SB-ins-101', '補償率101保存可', '');
    }
    await cancel(page);
  }
} catch (e) { rec('insurance-bound', 'err', 'FAIL', e.message); }

// payment system
try {
  await go(page, '/settings/payment-methods');
  const row = page.locator('tbody tr').filter({ hasText: /現金/ }).first();
  if (await row.count()) {
    await row.getByLabel('操作').click().catch(() => {});
    await page.waitForTimeout(300);
    const del = page.getByLabel('削除');
    rec('payment-system', 'no-delete', !(await del.count()) || !(await del.isEnabled()) ? 'PASS' : 'PARTIAL', '');
    await cancel(page);
  }
} catch (e) { rec('payment-system', 'err', 'FAIL', e.message); }

// merge batch1 log
let batch1 = [];
try {
  const log = fs.readFileSync(path.join(OUT, 'run.log'), 'utf8');
  for (const line of log.split('\n')) {
    const m = line.match(/^\[(PASS|FAIL|PARTIAL|BLOCKED|SKIP)\] ([^#]+)#(\S+)\s*(.*)$/);
    if (m) batch1.push({ id: m[2], step: m[3], status: m[1], note: m[4] || '' });
  }
} catch {}
// drop superseded treatment/medicine fails from batch1
const drop = new Set([
  'treatment-consultation#C2-reload', 'treatment-examination#C2-reload', 'treatment-procedure#C2-reload',
  'treatment-vaccine#C2-reload', 'treatment-checkup#C2-reload', 'medicine#create',
]);
batch1 = batch1.filter((r) => !drop.has(`${r.id}#${r.step}`));
// also drop treatment create if we're redoing all steps - keep open/C1 from b1 or b2
const map = new Map();
for (const r of batch1.concat(results)) map.set(`${r.id}#${r.step}`, r);
const final = [...map.values()];
// clear false bugs - only keep if still FAIL create/C2 in final
const realFails = final.filter((r) => r.status === 'FAIL');
const realBugs = [];
for (const f of realFails) {
  if (f.step === 'create' || f.step === 'C2-reload' || f.step === 'open' || f.step === 'C3-2') {
    realBugs.push({ id: `SB-${f.id}-${f.step}`, title: `${f.id} ${f.step} FAIL`, evidence: f.note });
  }
}
fs.writeFileSync(path.join(OUT, 'results.json'), JSON.stringify(final, null, 2));
fs.writeFileSync(path.join(OUT, 'bug-candidates.json'), JSON.stringify(realBugs, null, 2));
const c = final.reduce((a, r) => { a[r.status] = (a[r.status] || 0) + 1; return a; }, {});
console.log('TOTALS', c, 'steps', final.length);
console.log('FAILS', realFails.map((f) => `${f.id}#${f.step}`).join(','));
console.log('BUGS', realBugs.map((b) => b.id).join(','));
await browser.close();
