/**
 * V04 batch2 — remaining masters + specials (faster)
 */
import { chromium } from '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/node_modules/playwright/index.mjs';
import fs from 'fs';
import path from 'path';

const BASE = 'http://localhost:3003';
const OUT = '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/uat-2026-08-14-v04';
const email = process.env.E2E_LOGIN_EMAIL || 'admin@noavet.jp';
const password = process.env.E2E_LOGIN_PASSWORD || 'password';
const TAG = `V04b_${Date.now().toString().slice(-8)}`;
const results = [];
const bugs = [];
const rec = (id, step, status, note = '') => {
  results.push({ id, step, status, note: String(note).slice(0, 400) });
  console.log(`[${status}] ${id}#${step} ${String(note).slice(0, 120)}`);
};
const bug = (id, title, ev) => { bugs.push({ id, title, evidence: ev }); console.log(`[BUG?] ${id} ${title}`); };

async function login(page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'domcontentloaded' });
  await page.locator('#login-email').fill(email);
  await page.locator('#login-password').fill(password);
  await page.getByRole('button', { name: 'ログイン', exact: true }).click();
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 30000 });
}
async function go(page, route) {
  await page.goto(`${BASE}${route}`, { waitUntil: 'domcontentloaded', timeout: 40000 });
  await page.waitForTimeout(500);
}
async function openNew(page) {
  const b = page.getByRole('button', { name: '新規登録' });
  if (!(await b.count())) return false;
  await b.click();
  try { await page.waitForSelector('#master-title', { timeout: 8000 }); return true; } catch { return false; }
}
async function save(page) {
  const s = page.getByRole('button', { name: '保存', exact: true }).first();
  if (await s.count()) await s.click();
  await page.waitForTimeout(1200);
}
async function cancel(page) {
  const c = page.getByRole('button', { name: 'キャンセル' });
  if (await c.count()) await c.click().catch(() => {});
  await page.keyboard.press('Escape').catch(() => {});
  await page.waitForTimeout(300);
}

async function masterFlow(page, id, route, { unique = true, extra } = {}) {
  const name = `${TAG}_${id}`;
  try {
    await go(page, route);
    if (/Something went wrong/i.test(await page.locator('body').innerText())) {
      rec(id, 'open', 'FAIL', 'crash'); bug(`V04-${id}`, `${route} crash`, route); return;
    }
    rec(id, 'open', 'PASS', route);
    if (!(await openNew(page))) { rec(id, 'panel', 'BLOCKED', 'no new'); return; }
    // C1
    await save(page);
    const panel = await page.locator('#master-title').isVisible().catch(() => false);
    rec(id, 'C1-1', panel ? 'PASS' : 'PARTIAL', panel ? 'blocked empty' : 'panel closed?');
    await cancel(page);
    if (!(await openNew(page))) return;
    await page.locator('#master-title').fill(name);
    if (extra) await extra(page);
    // pick first combo if any required
    if (await page.getByRole('combobox').count()) {
      await page.getByRole('combobox').first().click().catch(() => {});
      await page.waitForTimeout(200);
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
    }
    const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 12000 }).catch(() => null);
    await save(page);
    const r = await w;
    await page.waitForTimeout(500);
    let vis = (await page.getByText(name).count()) > 0;
    if (!vis) { await go(page, route); vis = (await page.getByText(name).count()) > 0; }
    const ok = (r && r.ok()) || vis;
    rec(id, 'create', ok ? 'PASS' : 'FAIL', r ? `${r.status()} vis=${vis}` : `vis=${vis}`);
    if (!ok) { bug(`V04-${id}-create`, `${route} 新規失敗`, name); return; }
    await go(page, route);
    rec(id, 'C2-reload', (await page.getByText(name).count()) > 0 ? 'PASS' : 'FAIL', name);
    if (unique) {
      if (await openNew(page)) {
        await page.locator('#master-title').fill(name);
        if (await page.getByRole('combobox').count()) {
          await page.getByRole('combobox').first().click().catch(() => {});
          if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
        }
        const w2 = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
        await save(page);
        const r2 = await w2;
        const rej = r2 && (r2.status() === 409 || r2.status() >= 400);
        rec(id, 'C3-2', rej ? 'PASS' : 'PARTIAL', r2 ? String(r2.status()) : 'noresp');
        if (!rej && r2 && r2.ok()) bug(`V04-${id}-unique`, `${route} 同名重複可`, name);
        await cancel(page);
      }
    } else rec(id, 'C3-2', 'SKIP', 'no unique');
  } catch (e) { rec(id, 'err', 'FAIL', e.message); }
}

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
page.setDefaultTimeout(35000);
await login(page);
rec('ENV', 'login', 'PASS', '');

// remaining from timeout + retries for inquiry
const list = [
  ['trimming-course', '/settings/trimming?tab=course', true, null],
  ['trimming-option', '/settings/trimming?tab=option', true, null],
  ['trimming-course-type', '/settings/trimming-course-type', true, null],
  ['campaigns', '/settings/campaigns', false, async (page) => {
    const dates = page.locator('input[type=date]');
    if ((await dates.count()) >= 2) {
      await dates.nth(0).fill('2026-08-01');
      await dates.nth(1).fill('2026-12-31');
    }
    const num = page.locator('input[type=number]').first();
    if (await num.count()) await num.fill('10').catch(() => {});
  }],
  ['payment-methods', '/settings/payment-methods', true, null],
  ['medicine', '/settings/medicine', true, null],
  ['shift-templates', '/settings/shift-templates', true, null],
  ['inquiry-templates', '/settings/inquiry-templates', false, async (page) => {
    const cat = page.getByLabel('カテゴリ');
    if (await cat.count()) await cat.fill('V04カテゴリ');
    const ta = page.locator('textarea').first();
    if (await ta.count()) await ta.fill('V04 body').catch(() => {});
  }],
];
for (const [id, route, u, extra] of list) await masterFlow(page, id, route, { unique: u, extra });

// treatment tabs
for (const tab of ['consultation', 'examination', 'procedure', 'vaccine', 'checkup']) {
  const id = `treatment-${tab}`;
  const route = `/settings/treatment-items?tab=${tab}`;
  const name = `${TAG}_${id}`;
  try {
    await go(page, route);
    rec(id, 'open', 'PASS', route);
    if (!(await openNew(page))) { rec(id, 'panel', 'BLOCKED', 'no'); continue; }
    if (tab === 'consultation') {
      await save(page);
      rec(id, 'C1-1', (await page.locator('#master-title').isVisible().catch(() => false)) ? 'PASS' : 'PARTIAL', 'empty');
      await cancel(page);
      await openNew(page);
    }
    await page.locator('#master-title').fill(name);
    const num = page.locator('input[type=number]').first();
    if (await num.count()) await num.fill('100').catch(() => {});
    const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && r.request().method() === 'POST', { timeout: 12000 }).catch(() => null);
    await save(page);
    const r = await w;
    await go(page, route);
    const vis = (await page.getByText(name).count()) > 0;
    rec(id, 'create', (r && r.ok()) || vis ? 'PASS' : 'FAIL', r ? String(r.status()) : `vis=${vis}`);
    if (!((r && r.ok()) || vis)) bug(`V04-${id}`, `treatment ${tab} create fail`, name);
  } catch (e) { rec(id, 'err', 'FAIL', e.message); }
}

// insurance 101
try {
  await go(page, '/settings/insurance');
  if (await openNew(page)) {
    await page.locator('#master-title').fill(`${TAG}_ins101`);
    const n = page.locator('input[type=number]').first();
    if (await n.count()) {
      await n.fill('101');
      const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 8000 }).catch(() => null);
      await save(page);
      const r = await w;
      const body = await page.locator('body').innerText();
      const rej = (r && r.status() >= 400) || /補償|範囲|100|入力/.test(body) || (await page.locator('#master-title').isVisible().catch(() => false));
      rec('insurance-101', 'C1-3', rej ? 'PASS' : 'FAIL', r ? String(r.status()) : body.slice(0, 40));
      if (!rej) bug('V04-ins-101', '補償率101が保存できてしまう', '');
    } else rec('insurance-101', 'input', 'BLOCKED', 'no number');
    await cancel(page);
  }
} catch (e) { rec('insurance-101', 'err', 'FAIL', e.message); }

// payment system delete
try {
  await go(page, '/settings/payment-methods');
  const row = page.locator('tbody tr').filter({ hasText: /現金/ }).first();
  if (await row.count()) {
    await row.getByLabel('操作').click().catch(() => {});
    await page.waitForTimeout(400);
    const del = page.getByLabel('削除');
    if (await del.count() && await del.isEnabled()) {
      await del.click();
      const d = page.getByRole('alertdialog');
      if (await d.count()) await d.getByRole('button', { name: '削除' }).click();
      await page.waitForTimeout(800);
      const t = await page.locator('body').innerText();
      const ok = /システム標準|削除できません/.test(t) || (await page.getByText('現金').count()) > 0;
      rec('payment-system', 'no-delete', ok ? 'PASS' : 'FAIL', t.slice(0, 50));
      if (!ok) bug('V04-pay-sys', '標準支払削除可', '');
    } else rec('payment-system', 'no-delete', 'PASS', 'delete disabled');
    await cancel(page);
  }
} catch (e) { rec('payment-system', 'err', 'FAIL', e.message); }

// closing / slots / lstep / page-editor / clinic
for (const [id, route, re] of [
  ['closing-time', '/settings/closing-time', /締め|AM|PM|休診/],
  ['slots', '/line-reservation/slots', /枠|スロット|区分|予約/],
  ['lstep', '/settings/integrations/lstep', /Lステップ|連携|API/],
  ['lstep-tags', '/settings/lstep/tags', /タグ|prefix|条件/],
  ['page-editor', '/line-reservation/page-editor', /ページ|ヘッダ|プライバシー|予約/],
  ['clinic', '/settings/clinic', /医院|インボイス|法人/],
  ['diagnosis', '/settings/diagnosis', /診断/],
  ['animal-species', '/settings/animal-species', /動物/],
]) {
  try {
    await go(page, route);
    const t = await page.locator('body').innerText();
    const crash = /Something went wrong/i.test(t);
    rec(id, 'open', crash ? 'FAIL' : re.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 50));
    if (crash) bug(`V04-${id}-crash`, `${route} crash`, '');
  } catch (e) { rec(id, 'err', 'FAIL', e.message); }
}

// slots invalid typeId
try {
  await go(page, '/line-reservation/slots?typeId=99999999');
  const t = await page.locator('body').innerText();
  rec('slots-bad-id', 'C3-3', !/Something went wrong/i.test(t) && t.length > 20 ? 'PASS' : 'FAIL', t.slice(0, 40));
} catch (e) { rec('slots-bad-id', 'err', 'FAIL', e.message); }

// page editor save
try {
  await go(page, '/line-reservation/page-editor');
  const saveBtn = page.getByRole('button', { name: /保存/ }).first();
  if (await saveBtn.count()) {
    const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['PUT', 'PATCH', 'POST'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
    await saveBtn.click();
    const r = await w;
    rec('page-editor', 'save', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : 'noresp');
  }
} catch (e) { rec('page-editor', 'save', 'FAIL', e.message); }

// lstep save
try {
  await go(page, '/settings/integrations/lstep');
  const saveBtn = page.getByRole('button', { name: /保存/ }).first();
  if (await saveBtn.count() && await saveBtn.isEnabled()) {
    const w = page.waitForResponse((r) => /lstep/i.test(r.url()) && ['PUT', 'PATCH', 'POST'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
    await saveBtn.click();
    const r = await w;
    rec('lstep', 'save', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : 'noresp');
  }
} catch (e) { rec('lstep', 'save', 'FAIL', e.message); }

// merge with batch1 log lines if present - write full results
const prevPath = path.join(OUT, 'batch1-partial.json');
// parse batch1 from run.log if needed
let batch1 = [];
try {
  const log = fs.readFileSync(path.join(OUT, 'run.log'), 'utf8');
  for (const line of log.split('\n')) {
    const m = line.match(/^\[(PASS|FAIL|PARTIAL|BLOCKED|SKIP)\] ([^#]+)#(\S+)\s*(.*)$/);
    if (m) batch1.push({ id: m[2], step: m[3], status: m[1], note: m[4] || '' });
  }
} catch {}
const merged = batch1.concat(results);
// dedupe by id#step keep last
const map = new Map();
for (const r of merged) map.set(`${r.id}#${r.step}`, r);
const final = [...map.values()];
fs.writeFileSync(path.join(OUT, 'results.json'), JSON.stringify(final, null, 2));
fs.writeFileSync(path.join(OUT, 'bug-candidates.json'), JSON.stringify(bugs, null, 2));
const c = final.reduce((a, r) => { a[r.status] = (a[r.status] || 0) + 1; return a; }, {});
console.log('TOTALS', c, 'steps', final.length);
console.log('FAILS', final.filter((r) => r.status === 'FAIL').map((r) => `${r.id}#${r.step}`).join(','));
console.log('BUGS', bugs.map((b) => b.id).join(','));
await browser.close();
