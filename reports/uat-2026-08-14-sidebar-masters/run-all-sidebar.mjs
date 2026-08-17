/**
 * Fresh full sidebar マスタ設定 UAT — all pages + form fields
 * 2026-08-14 re-run
 */
import { chromium } from '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/node_modules/playwright/index.mjs';
import fs from 'fs';
import path from 'path';

const BASE = 'http://localhost:3003';
const OUT = path.dirname(new URL(import.meta.url).pathname);
const email = process.env.E2E_LOGIN_EMAIL || 'admin@noavet.jp';
const password = process.env.E2E_LOGIN_PASSWORD || 'password';
const TAG = `SB_${Date.now().toString().slice(-8)}`;
const only = process.env.UAT_ONLY || ''; // e.g. "batch1" | "batch2" | ""
const results = [];
const bugs = [];
const rec = (id, step, status, note = '') => {
  results.push({ id, step, status, note: String(note).slice(0, 400), at: new Date().toISOString() });
  console.log(`[${status}] ${id}#${step} ${String(note).slice(0, 120)}`);
};
const bug = (id, title, ev) => {
  bugs.push({ id, title, evidence: String(ev).slice(0, 250) });
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
  await page.waitForTimeout(550);
}
async function openNew(page) {
  const b = page.getByRole('button', { name: '新規登録' });
  if (!(await b.count())) return false;
  await b.click();
  await page.waitForTimeout(600);
  return true;
}
async function saveBtn(page) {
  return page.getByRole('button', { name: '保存', exact: true }).first();
}
async function save(page) {
  const s = await saveBtn(page);
  if (await s.count()) {
    if (await s.isEnabled()) await s.click();
  }
  await page.waitForTimeout(1100);
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
async function hasTitle(page) {
  return (await page.locator('#master-title').count()) > 0;
}

/** Safe extras per kind — do not spam all numbers */
async function fillFields(page, name, kind) {
  if (kind === 'shift') {
    await page.getByLabel('テンプレート名').fill(name);
    const combo = page.getByRole('combobox');
    for (let i = 0; i < Math.min(await combo.count(), 2); i++) {
      await combo.nth(i).click().catch(() => {});
      await page.waitForTimeout(150);
      const o = page.getByRole('option');
      if (await o.count()) {
        const full = o.filter({ hasText: '全日' });
        if (await full.count()) await full.first().click();
        else await o.first().click();
      }
    }
    await page.getByLabel('開始時刻').fill('09:00').catch(() => {});
    await page.getByLabel('終了時刻').fill('18:00').catch(() => {});
    await page.getByLabel('メモ').fill('SBメモ').catch(() => {});
    return;
  }
  if (await hasTitle(page)) await page.locator('#master-title').fill(name);
  if (kind === 'campaigns') {
    const d = page.locator('input[type=date]');
    if ((await d.count()) >= 2) {
      await d.nth(0).fill('2026-08-01');
      await d.nth(1).fill('2026-12-31');
    }
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
    if (await ta.count()) await ta.fill('SB本文').catch(() => {});
  } else if (kind === 'medicine') {
    const price = page.locator('input[type=number]').first();
    if (await price.count()) await price.fill('100').catch(() => {});
  } else if (kind === 'treatment') {
    const n = page.locator('input[type=number]').first();
    if (await n.count()) await n.fill('100').catch(() => {});
  } else if (kind === 'insurance') {
    const n = page.locator('input[type=number]').first();
    if (await n.count()) await n.fill('80').catch(() => {});
  } else if (kind === 'diagnosis-name') {
    const c = page.getByRole('combobox').first();
    if (await c.count()) {
      await c.click();
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click();
    }
  } else if (kind === 'cage') {
    const c = page.getByRole('combobox');
    for (let i = 0; i < Math.min(await c.count(), 2); i++) {
      await c.nth(i).click().catch(() => {});
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
    }
  } else {
    const cat = page.getByLabel('カテゴリ');
    if (await cat.count()) await cat.fill('SBカテゴリ').catch(() => {});
    const c = page.getByRole('combobox').first();
    if (await c.count()) {
      await c.click().catch(() => {});
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
    }
    // optional description textarea
    const ta = page.locator('textarea').first();
    if (await ta.count() && await ta.isVisible().catch(() => false)) await ta.fill('SB desc').catch(() => {});
  }
}

async function masterFlow(page, id, route, { unique = true, kind = 'generic' } = {}) {
  const name = `${TAG}_${id}`;
  try {
    await go(page, route);
    const body = await page.locator('body').innerText();
    if (/Something went wrong|Application error/i.test(body)) {
      rec(id, 'open', 'FAIL', 'crash');
      bug(`SB-${id}-crash`, `${route} crash`, route);
      return;
    }
    rec(id, 'open', 'PASS', route);

    if (!(await openNew(page))) {
      rec(id, 'panel', 'BLOCKED', 'no 新規登録');
      return;
    }

    // C1-1
    if (kind === 'shift') {
      const s = await saveBtn(page);
      const disabled = await s.isDisabled().catch(() => false);
      rec(id, 'C1-1', disabled ? 'PASS' : 'PARTIAL', disabled ? 'save disabled empty' : 'enabled');
    } else if (await hasTitle(page)) {
      await save(page);
      const stay = await page.locator('#master-title').isVisible().catch(() => false);
      rec(id, 'C1-1', stay ? 'PASS' : 'PARTIAL', stay ? 'empty blocked' : 'panel closed');
      await cancel(page);
      if (!(await openNew(page))) {
        rec(id, 'create', 'BLOCKED', 'no panel reopen');
        return;
      }
    } else {
      rec(id, 'C1-1', 'PARTIAL', 'no master-title');
    }

    await fillFields(page, name, kind);
    // wait save enable for shift
    if (kind === 'shift') {
      const s = await saveBtn(page);
      for (let i = 0; i < 20 && (await s.isDisabled().catch(() => false)); i++) await page.waitForTimeout(150);
    }
    const w = page.waitForResponse(
      (r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()),
      { timeout: 15000 },
    ).catch(() => null);
    await save(page);
    const resp = await w;
    await go(page, route);
    await searchName(page, name);
    const vis = (await page.getByText(name).count()) > 0;
    const ok = (resp && resp.status() >= 200 && resp.status() < 300) || vis;
    rec(id, 'create', ok ? 'PASS' : 'FAIL', resp ? `${resp.status()} vis=${vis}` : `vis=${vis}`);
    if (!ok) {
      let t = '';
      try { t = resp ? (await resp.text()).slice(0, 100) : ''; } catch {}
      bug(`SB-${id}-create`, `${route} 新規失敗`, t || name);
      return;
    }
    rec(id, 'C2-reload', vis ? 'PASS' : 'FAIL', name);
    if (!vis) bug(`SB-${id}-C2`, `${route} 再読込で一覧に出ない`, name);

    // reopen
    if (vis) {
      const row = page.locator('tbody tr, tr').filter({ hasText: name }).first();
      if (await row.count()) {
        const op = row.getByLabel('操作');
        if (await op.count()) await op.click();
        else await row.locator('button').last().click().catch(() => {});
        await page.waitForTimeout(500);
        if (kind === 'shift') {
          const v = await page.getByLabel('テンプレート名').inputValue().catch(() => '');
          rec(id, 'C2-reopen', v.includes(name) ? 'PASS' : 'PARTIAL', v.slice(0, 40));
          const memo = await page.getByLabel('メモ').inputValue().catch(() => '');
          rec(id, 'field-memo', memo.includes('SB') ? 'PASS' : 'PARTIAL', memo);
        } else if (await hasTitle(page)) {
          const v = await page.locator('#master-title').inputValue().catch(() => '');
          rec(id, 'C2-reopen', v.includes(name) ? 'PASS' : 'PARTIAL', v.slice(0, 40));
        } else rec(id, 'C2-reopen', 'PARTIAL', 'no title field');
        await cancel(page);
      }
    }

    // C3-2
    if (unique && vis) {
      if (await openNew(page)) {
        await fillFields(page, name, kind);
        if (kind === 'shift') {
          const s = await saveBtn(page);
          for (let i = 0; i < 15 && (await s.isDisabled().catch(() => false)); i++) await page.waitForTimeout(150);
        }
        const w2 = page.waitForResponse(
          (r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()),
          { timeout: 12000 },
        ).catch(() => null);
        await save(page);
        const r2 = await w2;
        const rej = r2 && (r2.status() === 409 || r2.status() >= 400);
        if (!rej && r2 && r2.ok()) {
          rec(id, 'C3-2', 'FAIL', `dup ${r2.status()}`);
          bug(`SB-${id}-unique`, `${route} 同名重複可`, name);
        } else rec(id, 'C3-2', rej ? 'PASS' : 'PARTIAL', r2 ? String(r2.status()) : 'noresp');
        await cancel(page);
      }
    } else if (!unique) rec(id, 'C3-2', 'SKIP', 'no unique');
  } catch (e) {
    rec(id, 'err', 'FAIL', e.message);
    bug(`SB-${id}-err`, `${route} 例外`, e.message);
  }
}

// --- specials ---
async function clinicFlow(page) {
  try {
    await go(page, '/settings/clinic');
    const t = await page.locator('body').innerText();
    rec('clinic', 'open', /医院|インボイス|法人|クリニック/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 50));
    const n = await page.locator('input,textarea,select,button').count();
    rec('clinic', 'fields_present', n > 5 ? 'PASS' : 'PARTIAL', `controls=${n}`);
    const inputs = page.locator('input[type=text]');
    let did = false;
    for (let i = 0; i < Math.min(await inputs.count(), 20); i++) {
      const aria = ((await inputs.nth(i).getAttribute('aria-label')) || '') + ((await inputs.nth(i).getAttribute('placeholder')) || '');
      const val = await inputs.nth(i).inputValue().catch(() => '');
      if (/インボイス|登録番号/.test(aria) || val.startsWith('T')) {
        const old = val;
        await inputs.nth(i).fill('T9999999999999');
        const saveB = page.getByRole('button', { name: /保存|更新/ }).first();
        if (await saveB.count()) {
          const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['PUT', 'PATCH', 'POST'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
          await saveB.click();
          const r = await w;
          rec('clinic', 'invoice-save', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : 'noresp');
          await inputs.nth(i).fill(old || '');
          if (await saveB.count()) await saveB.click().catch(() => {});
          did = true;
          break;
        }
      }
    }
    if (!did) rec('clinic', 'invoice-save', 'PARTIAL', 'heuristic miss');
  } catch (e) { rec('clinic', 'err', 'FAIL', e.message); }
}

async function staffFlow(page) {
  try {
    await go(page, '/settings/staff');
    rec('staff', 'open', /スタッフ/.test(await page.locator('body').innerText()) ? 'PASS' : 'PARTIAL', '');
    const neu = page.getByRole('button', { name: /新規|追加|招待/ });
    rec('staff', 'new_btn', (await neu.count()) > 0 ? 'PASS' : 'PARTIAL', `n=${await neu.count()}`);
    if (await neu.count()) {
      await neu.first().click();
      await page.waitForTimeout(500);
      const inv = await page.evaluate(() => [...document.querySelectorAll('input,textarea,select,[role=combobox]')].filter((e) => {
        const s = window.getComputedStyle(e);
        return s.display !== 'none' && s.visibility !== 'hidden';
      }).map((e) => ({ tag: e.tagName, type: e.getAttribute('type'), aria: e.getAttribute('aria-label'), ph: e.getAttribute('placeholder'), id: e.id })).slice(0, 30));
      rec('staff', 'field_list', inv.length ? 'PASS' : 'PARTIAL', JSON.stringify(inv).slice(0, 280));
      // fill each text/email empty check
      const saveB = page.getByRole('button', { name: /保存|登録|作成|招待/ }).first();
      if (await saveB.count()) {
        await saveB.click();
        await page.waitForTimeout(600);
        const t = await page.locator('body').innerText();
        rec('staff', 'C1-1', /入力|必須|メール|してください/.test(t) || (await page.locator('input:invalid').count()) > 0 ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 50));
      }
      // inventory: try fill name-like fields without submitting real invite
      if (await hasTitle(page)) {
        await page.locator('#master-title').fill(`${TAG}_staff_draft`);
        rec('staff', 'fill-title', 'PASS', 'draft name');
      }
      await cancel(page);
    }
  } catch (e) { rec('staff', 'err', 'FAIL', e.message); }
}

async function closingFlow(page) {
  try {
    await go(page, '/settings/closing-time');
    const t = await page.locator('body').innerText();
    rec('closing-time', 'open', /締め|AM|PM|休診|特別/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 50));
    const times = page.locator('input[type=time]');
    rec('closing-time', 'time_fields', (await times.count()) >= 1 ? 'PASS' : 'PARTIAL', `n=${await times.count()}`);
    const saveB = page.getByRole('button', { name: /保存/ }).first();
    if (await saveB.count() && await saveB.isEnabled()) {
      const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['PUT', 'PATCH', 'POST'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
      await saveB.click();
      const r = await w;
      rec('closing-time', 'save', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : 'noresp');
    }
    // holiday empty
    const add = page.getByRole('button', { name: /追加/ });
    if (await add.count()) {
      await add.first().click();
      await page.waitForTimeout(400);
      const date = page.locator('input[type=date]').first();
      if (await date.count()) {
        // leave empty, try save
        const s2 = page.getByRole('button', { name: /保存|追加|登録/ }).last();
        if (await s2.count() && !(await s2.isEnabled())) rec('closing-time', 'holiday-C1', 'PASS', 'disabled');
        else if (await s2.count()) {
          await s2.click();
          await page.waitForTimeout(400);
          rec('closing-time', 'holiday-C1', 'PARTIAL', 'clicked empty');
        }
      }
    }
  } catch (e) { rec('closing-time', 'err', 'FAIL', e.message); }
}

const BATCH1 = [
  ['animal-species', '/settings/animal-species', true, 'generic'],
  ['treatment-consultation', '/settings/treatment-items?tab=consultation', true, 'treatment'],
  ['treatment-examination', '/settings/treatment-items?tab=examination', true, 'treatment'],
  ['treatment-procedure', '/settings/treatment-items?tab=procedure', true, 'treatment'],
  ['treatment-vaccine', '/settings/treatment-items?tab=vaccine', true, 'treatment'],
  ['treatment-checkup', '/settings/treatment-items?tab=checkup', true, 'treatment'],
  ['diagnosis-type', '/settings/diagnosis?tab=diagnosis_type', true, 'generic'],
  ['diagnosis-name', '/settings/diagnosis?tab=diagnosis_name', false, 'diagnosis-name'],
  ['inquiry-templates', '/settings/inquiry-templates', false, 'inquiry'],
  ['chief-complaint', '/settings/interview/chief-complaint', true, 'generic'],
  ['interview-templates', '/settings/interview/templates', false, 'interview'],
  ['medicine', '/settings/medicine', true, 'medicine'],
  ['reservation-type', '/settings/reservation-type', true, 'generic'],
  ['hospitalization-plan', '/settings/hospitalization', true, 'generic'],
];

const BATCH2 = [
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
];

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
page.setDefaultTimeout(35000);
await login(page);
rec('ENV', 'login', 'PASS', page.url());
await go(page, '/settings');
rec('settings-hub', 'open', /マスタ|設定/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', '/settings');

if (!only || only === 'batch1' || only === 'all') {
  await clinicFlow(page);
  for (const [id, route, u, k] of BATCH1) await masterFlow(page, id, route, { unique: u, kind: k });
}
if (!only || only === 'batch2' || only === 'all') {
  for (const [id, route, u, k] of BATCH2) await masterFlow(page, id, route, { unique: u, kind: k });
  await staffFlow(page);
  await closingFlow(page);

  // insurance 101
  try {
    await go(page, '/settings/insurance');
    if (await openNew(page) && await hasTitle(page)) {
      await page.locator('#master-title').fill(`${TAG}_i101`);
      const n = page.locator('input[type=number]').first();
      if (await n.count()) {
        await n.fill('101');
        const w = page.waitForResponse((r) => /\/api\//.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 8000 }).catch(() => null);
        await save(page);
        const r = await w;
        const stay = await page.locator('#master-title').isVisible().catch(() => false);
        const rej = (r && r.status() >= 400) || stay;
        rec('insurance-bound', 'C1-3-101', rej ? 'PASS' : 'FAIL', r ? String(r.status()) : 'stay');
        if (!rej) bug('SB-ins-101', '補償率101が保存できる', '');
      }
      await cancel(page);
    }
  } catch (e) { rec('insurance-bound', 'err', 'FAIL', e.message); }

  // payment system delete
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
}

await browser.close();

// merge if batch split
const outFile = path.join(OUT, only ? `results-${only}.json` : 'results-fresh.json');
const bugFile = path.join(OUT, only ? `bugs-${only}.json` : 'bugs-fresh.json');
fs.writeFileSync(outFile, JSON.stringify(results, null, 2));
fs.writeFileSync(bugFile, JSON.stringify(bugs, null, 2));
const c = results.reduce((a, r) => { a[r.status] = (a[r.status] || 0) + 1; return a; }, {});
console.log('FILE', outFile);
console.log('TOTALS', c, 'steps', results.length);
console.log('FAILS', results.filter((r) => r.status === 'FAIL').map((r) => `${r.id}#${r.step}`).join(',') || '(none)');
console.log('BUGS', bugs.map((b) => b.id).join(',') || '(none)');
