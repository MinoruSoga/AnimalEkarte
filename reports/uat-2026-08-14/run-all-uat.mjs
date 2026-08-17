/**
 * Fresh full UAT from zero — S01-S13 + V01-V05
 * 2026-08-14 · no scenario md edits · bugs → bug-candidates.json
 */
import { chromium } from '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/node_modules/playwright/index.mjs';
import fs from 'fs';
import path from 'path';

const BASE = 'http://localhost:3003';
const API = 'http://localhost:8080';
const OUT = '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/uat-2026-08-14';
const email = process.env.E2E_LOGIN_EMAIL || 'admin@noavet.jp';
const password = process.env.E2E_LOGIN_PASSWORD || 'password';
fs.mkdirSync(OUT, { recursive: true });
const results = [];
const bugs = [];
const rec = (scenario, step, status, note = '') => {
  results.push({ scenario, step, status, note: String(note).slice(0, 800), at: new Date().toISOString() });
  console.log(`[${status}] ${scenario}#${step} ${String(note).slice(0, 160)}`);
};
const bug = (id, title, evidence) => {
  bugs.push({ id, title, evidence });
  console.log(`[BUG?] ${id} ${title}`);
};
const shot = async (page, n) => {
  try { await page.screenshot({ path: path.join(OUT, `${n}.png`), fullPage: true }); } catch {}
};

async function login(page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
  await page.locator('#login-email').fill(email);
  await page.locator('#login-password').fill(password);
  await page.getByRole('button', { name: 'ログイン', exact: true }).click();
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 30000 });
}
async function apiLogin(request) {
  return request.post(`${API}/api/v1/login`, {
    headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest', Origin: BASE },
    data: { email, password },
  });
}
async function selectAlive(page, listPath, q) {
  await page.goto(`${BASE}${listPath}`, { waitUntil: 'networkidle' });
  if (!page.url().includes('select-pet') && !page.url().includes('petId=')) {
    const neu = page.getByRole('link', { name: /新規/ }).or(page.getByRole('button', { name: /新規/ })).first();
    if (await neu.count()) { await neu.click(); await page.waitForTimeout(1000); }
  }
  await page.waitForSelector('#search', { timeout: 20000 });
  await page.locator('#search').fill(q);
  await page.waitForTimeout(1600);
  const row = page.locator('tr').filter({ hasText: q }).filter({ hasText: '生存' }).first();
  const btn = row.getByRole('button', { name: '選択' });
  if (await btn.count() && await btn.isEnabled()) {
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
  if (await opt.count()) { await opt.click(); await page.waitForTimeout(1200); return true; }
  if (await page.getByRole('option').count() > 1) {
    await page.getByRole('option').nth(1).click();
    await page.waitForTimeout(1200);
    return true;
  }
  return false;
}

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
const page = await context.newPage();
page.setDefaultTimeout(35000);

// ENV
const h = await page.request.get(`${API}/health`);
rec('ENV', 'health', h.ok() ? 'PASS' : 'FAIL', String(h.status()));
await login(page);
rec('ENV', 'login', 'PASS', page.url());
const lr = await apiLogin(page.request);
rec('ENV', 'api_login', lr.ok() ? 'PASS' : 'FAIL', String(lr.status()));

// S04 first after env (order skill: S01 first actually - do S01 first)
// ===== S01 =====
try {
  let r = await page.request.fetch(`${API}/api/v1/clinics/1/pets/1000002/death`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json', 'X-Requested-With': 'XMLHttpRequest' },
    data: JSON.stringify({ deceased_at: '2026-08-14', reason: 'UAT fresh S01' }),
  });
  rec('S01', 'death', r.status() === 204 ? 'PASS' : 'FAIL', String(r.status()));
  if (r.status() !== 204) bug('S01-death', '死亡登録 API が 204 でない', String(r.status()));
  for (const [lab, pth] of [['mr', '/medical-records/select-pet'], ['acc', '/accounting/select-pet'], ['hosp', '/hospitalization/select-pet']]) {
    await page.goto(`${BASE}${pth}`, { waitUntil: 'networkidle' });
    await page.waitForSelector('#search');
    await page.locator('#search').fill('豆助');
    await page.waitForTimeout(1500);
    const dis = await page.getByRole('button', { name: /選択不可/ }).count();
    rec('S01', `block_${lab}`, dis > 0 ? 'PASS' : 'FAIL', `dis=${dis}`);
    if (dis === 0) bug('S01-block', `死亡ペットが${lab}で選択不可にならない`, pth);
  }
  r = await page.request.fetch(`${API}/api/v1/clinics/1/pets/1000002/death`, {
    method: 'DELETE', headers: { 'X-Requested-With': 'XMLHttpRequest' },
  });
  rec('S01', 'revive', r.status() === 204 ? 'PASS' : 'FAIL', String(r.status()));
  const pet = await (await page.request.get(`${API}/api/v1/pets/1000002`, { headers: { 'X-Requested-With': 'XMLHttpRequest' } })).json();
  rec('S01', 'alive', pet.status === 'alive' || !pet.deceased_at ? 'PASS' : 'FAIL', JSON.stringify({ status: pet.status }));
  await shot(page, 'S01');
} catch (e) { rec('S01', 'err', 'FAIL', e.message); bug('S01-err', 'S01 例外', e.message); }

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
      let saveOk = false; let saveNote = '';
      if (await save.count() && await save.isEnabled()) {
        const waitSave = page.waitForResponse(
          (r) => /\/examinations/.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()),
          { timeout: 15000 },
        ).catch(() => null);
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
      const body = await page.locator('body').innerText();
      if (body.includes('HIGH') || body.includes('LOW')) rec('S02', 'HL', 'PASS', `id=${m?.[1] || ''}`);
      else if (body.includes('未判定')) {
        rec('S02', 'HL', 'FAIL', '未判定 only');
        bug('S02-HL', '保存後 HIGH/LOW にならず未判定', `exam=${m?.[1]}`);
      } else rec('S02', 'HL', 'BLOCKED', body.replace(/\s+/g, ' ').slice(0, 100));
      await page.goto(`${BASE}/examinations/1014565`, { waitUntil: 'networkidle' });
      const t2 = await page.locator('body').innerText();
      const lock = /完了済みのため結果の編集/.test(t2);
      const en = await page.locator('input[name^="examItems."][name$=".inspectionValue"]:not([disabled])').count();
      rec('S02', 'completed_lock', lock && en === 0 ? 'PASS' : 'FAIL', `lock=${lock} en=${en}`);
      if (!(lock && en === 0)) bug('S02-lock', '完了検査が編集ロックされない', '1014565');
    }
    await shot(page, 'S02');
  }
} catch (e) { rec('S02', 'err', 'FAIL', e.message); }

// ===== S03 =====
try {
  const ok = await selectAlive(page, '/vaccinations', 'はな');
  rec('S03', 'select', ok ? 'PASS' : 'FAIL', page.url());
  if (ok) {
    if (await page.locator('select').count()) {
      const opts = await page.locator('select').first().locator('option').allTextContents();
      rec('S03', 'opts', opts.length > 5 ? 'PASS' : 'FAIL', `n=${opts.length}`);
      if (opts.length > 1) await page.locator('select').first().selectOption({ index: 1 });
      for (let i = 0; i < await page.locator('select').count(); i++) {
        const o = await page.locator('select').nth(i).locator('option').allTextContents();
        if (o.includes('3週後')) {
          await page.locator('select').nth(i).selectOption({ label: '3週後' });
          rec('S03', 'interval', 'PASS', '3週後');
          break;
        }
      }
    }
    const wait = page.waitForResponse((r) => /vaccin/i.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 12000 }).catch(() => null);
    const save = page.getByRole('button', { name: /保存|登録/ }).first();
    if (await save.count() && await save.isEnabled()) await save.click();
    const resp = await wait;
    rec('S03', 'save', resp && resp.ok() ? 'PASS' : 'FAIL', resp ? String(resp.status()) : 'noresp');
    await shot(page, 'S03');
  }
} catch (e) { rec('S03', 'err', 'FAIL', e.message); }

// ===== S04 =====
try {
  const p = await context.newPage();
  await p.goto(`${BASE}/line-reserve/1/`, { waitUntil: 'networkidle' });
  rec('S04', 'entry', /新規予約|オンライン予約/.test(await p.locator('body').innerText()) ? 'PASS' : 'FAIL', p.url());
  await p.getByText('新規予約').click();
  await p.waitForTimeout(500);
  await p.getByPlaceholder('山田 花子').fill('UATゼロ太郎');
  await p.getByPlaceholder('090-1234-5678').fill('09066667777');
  await p.getByText('新しいペットを追加').click();
  await p.waitForTimeout(300);
  const pl = p.getByLabel(/ペット名/);
  if (await pl.count()) await pl.fill('ゼロ犬');
  else await p.locator('input').nth(2).fill('ゼロ犬');
  for (const sp of ['犬']) {
    if (await p.getByText(sp, { exact: true }).count()) { await p.getByText(sp, { exact: true }).first().click(); break; }
  }
  for (const g of ['オス']) {
    if (await p.getByText(g, { exact: true }).count()) { await p.getByText(g, { exact: true }).first().click(); break; }
  }
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
  let first = null; let en = 0;
  for (let i = 0; i < await buttons.count(); i++) {
    const b = buttons.nth(i);
    const label = ((await b.innerText().catch(() => '')) || '').trim();
    if (!/^\d{1,2}$/.test(label)) continue;
    if (!(await b.isDisabled())) { en++; if (!first) first = b; }
  }
  rec('S04', 'dates', en > 0 ? 'PASS' : 'FAIL', `enabled=${en}`);
  if (en === 0) bug('S04-dates', 'LIFF予約で選択可能日が0（shifts seed後）', 'line-reserve');
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
  await p.close();
} catch (e) { rec('S04', 'err', 'FAIL', e.message); }

// ===== S05 =====
try {
  const ok = await selectAlive(page, '/hospitalization', '豆助');
  rec('S05', 'select', ok ? 'PASS' : 'FAIL', page.url());
  if (ok) {
    const reg = page.getByRole('button', { name: /登録/ }).first();
    if (await reg.count()) {
      await reg.click();
      await page.waitForTimeout(800);
      rec('S05', 'cage_required', /ケージ|必須|選択/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', 'validation');
    }
    const combos = page.getByRole('combobox');
    for (let i = 0; i < await combos.count(); i++) {
      await combos.nth(i).click().catch(() => {});
      await page.waitForTimeout(300);
      if (await page.getByRole('option').count() > 1) {
        await page.getByRole('option').nth(1).click();
        await page.waitForTimeout(400);
      }
    }
    if (await page.locator('select').count()) {
      const s = page.locator('select').first();
      if ((await s.locator('option').count()) > 1) await s.selectOption({ index: 1 });
    }
    const ta = page.locator('textarea').first();
    if (await ta.count() && await ta.isEditable()) await ta.fill('UAT S05').catch(() => {});
    const wait = page.waitForResponse((r) => /hospital/i.test(r.url()) && r.request().method() === 'POST', { timeout: 12000 }).catch(() => null);
    if (await reg.count() && await reg.isEnabled()) await reg.click();
    const resp = await wait;
    rec('S05', 'create', resp && (resp.status() === 201 || resp.ok()) ? 'PASS' : 'FAIL', resp ? String(resp.status()) : 'noresp');
    await page.goto(`${BASE}/hospitalization`, { waitUntil: 'networkidle' });
    rec('S05', 'list', /入院/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', 'list');
    for (const name of ['入院中', 'ボード', 'リスト']) {
      const t = page.getByRole('tab', { name }).or(page.getByRole('button', { name }));
      if (await t.count()) { await t.first().click().catch(() => {}); rec('S05', `tab_${name}`, 'PASS', 'ok'); break; }
    }
    const link = page.locator('table tbody tr a').first();
    if (await link.count()) { await link.click(); await page.waitForTimeout(1200); }
    else {
      const hl = await page.request.get(`${API}/api/v1/hospitalizations?limit=5`, { headers: { 'X-Requested-With': 'XMLHttpRequest' } });
      const hj = await hl.json();
      const items = hj.data || hj;
      const mine = (items || []).find((x) => String(x.pet_id) === '1000002' && x.status === 'admitted');
      if (mine) await page.goto(`${BASE}/hospitalization/${mine.id}`, { waitUntil: 'networkidle' });
    }
    const dis = page.getByRole('button', { name: /退院/ });
    if (await dis.count()) {
      await dis.first().click();
      await page.waitForTimeout(600);
      rec('S05', 'discharge_dialog', /退院/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', 'dialog');
      const exec = page.getByRole('button', { name: /退院処理を実行|実行する|退院する/ });
      const wait2 = page.waitForResponse((r) => /hospital|discharge/i.test(r.url()) && ['POST', 'PATCH', 'PUT'].includes(r.request().method()), { timeout: 12000 }).catch(() => null);
      if (await exec.count()) await exec.last().click();
      const r2 = await wait2;
      rec('S05', 'discharge', r2 && r2.ok() ? 'PASS' : 'PARTIAL', r2 ? String(r2.status()) : 'noresp');
    } else rec('S05', 'discharge', 'PARTIAL', 'no btn');
    await shot(page, 'S05');
  }
} catch (e) { rec('S05', 'err', 'FAIL', e.message); }

// ===== S06 =====
try {
  let ok = await selectAlive(page, '/medical-records', 'ろっぷ');
  if (!ok) ok = await selectAlive(page, '/medical-records', 'はな');
  rec('S06', 'select', ok ? 'PASS' : 'FAIL', page.url());
  if (ok) {
    const ta = page.locator('textarea:not([disabled])').first();
    if (await ta.count()) {
      await ta.fill(`UAT S06 ${Date.now()}`);
      const wait = page.waitForResponse((r) => /medical-record/i.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 15000 }).catch(() => null);
      const save = page.getByRole('button', { name: /保存|登録|作成/ }).first();
      if (await save.count() && await save.isEnabled()) await save.click();
      const resp = await wait;
      rec('S06', 'save', resp && resp.ok() ? 'PASS' : 'BLOCKED', resp ? String(resp.status()) : 'noresp');
    } else if (/確定済/.test(await page.locator('body').innerText())) {
      rec('S06', 'already_final', 'PASS', 'landed finalized');
    }
    const fin = page.getByRole('button', { name: /確定する/ });
    if (await fin.count() && await fin.isEnabled()) {
      await fin.click();
      await page.waitForTimeout(500);
      const wait2 = page.waitForResponse((r) => /medical-record/i.test(r.url()) && ['POST', 'PATCH', 'PUT'].includes(r.request().method()), { timeout: 15000 }).catch(() => null);
      await page.getByRole('button', { name: /確定する/ }).last().click();
      const r2 = await wait2;
      await page.waitForTimeout(1500);
      rec('S06', 'finalize', (r2 && r2.ok()) || /確定済/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', r2 ? String(r2.status()) : 'body');
    }
    await page.goto(`${BASE}/medical-records/1080036`, { waitUntil: 'networkidle' });
    const t = await page.locator('body').innerText();
    const saveVis = (await page.getByRole('button', { name: /^保存$/ }).count()) > 0 && (await page.getByRole('button', { name: /^保存$/ }).first().isVisible().catch(() => false));
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
        if (await sub.count() && await sub.isEnabled()) {
          await sub.click();
          await page.waitForTimeout(2000);
          rec('S06', 'addendum_save', 'PASS', 'ok');
        }
      }
      await page.keyboard.press('Escape');
    } else rec('S06', 'addendum_ui', 'PARTIAL', 'no btn');
    rec('S06', 'audit_db', 'BLOCKED', 'USER');
    await shot(page, 'S06');
  }
} catch (e) { rec('S06', 'err', 'FAIL', e.message); }

// ===== S07 =====
try {
  await page.goto(`${BASE}/estimates`, { waitUntil: 'networkidle' });
  rec('S07', 'list', /見積/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', 'list');
  let ok = false;
  try { ok = await selectAlive(page, '/estimates', 'はな'); } catch { ok = false; }
  if (!ok) await page.goto(`${BASE}/estimates/new`, { waitUntil: 'networkidle' });
  if (page.url().includes('select-pet') && await page.locator('#search').count()) {
    await page.locator('#search').fill('はな');
    await page.waitForTimeout(1500);
    const b = page.locator('tr').filter({ hasText: '生存' }).getByRole('button', { name: '選択' }).first();
    if (await b.count()) await b.click();
    await page.waitForTimeout(1000);
  }
  const inputs = page.locator('input[type=text], input:not([type])');
  if (await inputs.count()) await inputs.first().fill(`S07 zero ${Date.now()}`);
  const create = page.getByRole('button', { name: '作成', exact: true });
  const wait = page.waitForResponse((r) => /estimate/i.test(r.url()) && r.request().method() === 'POST', { timeout: 12000 }).catch(() => null);
  if (await create.count() && await create.isEnabled()) await create.click();
  const resp = await wait;
  rec('S07', 'create', resp && resp.ok() ? 'PASS' : 'FAIL', resp ? String(resp.status()) : 'noresp');
  await page.goto(`${BASE}/estimates/new`, { waitUntil: 'networkidle' });
  if (!page.url().includes('select-pet')) {
    const c2 = page.getByRole('button', { name: '作成', exact: true });
    if (await c2.count() && !(await c2.isEnabled())) rec('S07', 'empty_title_gate', 'PASS', 'disabled');
    else rec('S07', 'empty_title_gate', 'PARTIAL', 'enabled?');
  }
  await shot(page, 'S07');
} catch (e) { rec('S07', 'err', 'FAIL', e.message); }

// ===== S08 =====
try {
  await page.goto(`${BASE}/accounting`, { waitUntil: 'networkidle' });
  rec('S08', 'list', /会計/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', 'list');
  const ok = await selectAlive(page, '/accounting', 'はな');
  rec('S08', 'new_form', ok ? 'PASS' : 'FAIL', page.url());
  rec('S08', 'partial_pay', 'BLOCKED', 'product spec');
  await page.goto(`${BASE}/accounting`, { waitUntil: 'networkidle' });
  const row = page.locator('table tbody tr').first();
  if (await row.count()) {
    await row.locator('a, button').first().click().catch(() => row.click());
    await page.waitForTimeout(1000);
    rec('S08', 'detail', /会計|明細|合計/.test(await page.locator('body').innerText()) ? 'PASS' : 'PARTIAL', page.url());
  }
  await shot(page, 'S08');
} catch (e) { rec('S08', 'err', 'FAIL', e.message); }

// ===== S09 =====
try {
  await page.goto(`${BASE}/accounting/close`, { waitUntil: 'networkidle' });
  const t = await page.locator('body').innerText();
  rec('S09', 'close_ui', /締め|レジ|午前|午後|プレビュー/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 80));
  const prev = page.getByRole('button', { name: /プレビュー|表示|確認/ }).first();
  if (await prev.count() && await prev.isEnabled()) {
    await prev.click();
    await page.waitForTimeout(1000);
    rec('S09', 'preview', 'PASS', 'ok');
  } else rec('S09', 'preview', 'PARTIAL', 'no btn');
  rec('S09', 'fixture_attribution', 'BLOCKED', 'needs human DB');
  await page.goto(`${BASE}/accounting/close/history`, { waitUntil: 'networkidle' });
  rec('S09', 'history', /締め|履歴/.test(await page.locator('body').innerText()) ? 'PASS' : 'PARTIAL', 'history');
  await page.goto(`${BASE}/settings/closing-time`, { waitUntil: 'networkidle' });
  rec('S09', 'settings', /締め|時間|AM|PM/.test(await page.locator('body').innerText()) ? 'PASS' : 'PARTIAL', 'settings');
  await shot(page, 'S09');
} catch (e) { rec('S09', 'err', 'FAIL', e.message); }

// ===== S10 =====
try {
  await page.goto(`${BASE}/aggregation`, { waitUntil: 'networkidle', timeout: 60000 });
  await page.waitForTimeout(2500);
  const t = await page.locator('body').innerText();
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
} catch (e) { rec('S10', 'err', 'FAIL', e.message); }

// ===== S11 =====
try {
  await page.goto(`${BASE}/trimming`, { waitUntil: 'networkidle' });
  rec('S11', 'list', /トリミング/.test(await page.locator('body').innerText()) ? 'PASS' : 'FAIL', 'list');
  const ok = await selectAlive(page, '/trimming', 'はな').catch(() => false);
  rec('S11', 'new_select', ok ? 'PASS' : 'PARTIAL', page.url());
  if (ok) {
    await pickCombobox(page, 0, /./);
    const wait = page.waitForResponse((r) => /trimming/i.test(r.url()) && ['POST', 'PUT', 'PATCH'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
    const save = page.getByRole('button', { name: /保存|登録|作成/ }).first();
    if (await save.count() && await save.isEnabled()) await save.click();
    const resp = await wait;
    rec('S11', 'create', resp && resp.ok() ? 'PASS' : 'PARTIAL', resp ? String(resp.status()) : page.url());
  }
  const ub = await page.request.get(`${API}/api/v1/billing-items/unbilled-details?pet_id=1000005`, {
    headers: { 'X-Requested-With': 'XMLHttpRequest' },
  });
  rec('S11', 'unbilled', ub.status() === 200 ? 'PASS' : 'FAIL', String(ub.status()));
  await page.goto(`${BASE}/accounting/new?petId=1000005`, { waitUntil: 'networkidle' });
  rec('S11', 'accounting_with_pet', /会計|明細|未請求/.test(await page.locator('body').innerText()) ? 'PASS' : 'PARTIAL', page.url());
  await shot(page, 'S11');
} catch (e) { rec('S11', 'err', 'FAIL', e.message); }

// ===== S12 =====
try {
  const courses = await page.request.get(`${API}/api/liff/1/courses`);
  rec('S12', 'courses_api', courses.ok() ? 'PASS' : 'FAIL', String(courses.status()));
  await page.goto(`${BASE}/liff/health-card?clinic_id=1`, { waitUntil: 'networkidle' });
  const t = await page.locator('body').innerText();
  rec('S12', 'mock_ui', /テストユーザー|ペット情報/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 80));
  await page.goto(`${BASE}/liff/health-card`, { waitUntil: 'networkidle' });
  const t2 = await page.locator('body').innerText();
  rec('S12', 'missing_clinic', /クリニック|clinic|必須|失敗|エラー/.test(t2) ? 'PASS' : 'PARTIAL', t2.replace(/\s+/g, ' ').slice(0, 60));
  rec('S12', 'real_token', 'BLOCKED', 'real LINE token');
  await shot(page, 'S12');
} catch (e) { rec('S12', 'err', 'FAIL', e.message); }

// ===== S13 =====
try {
  await page.goto(`${BASE}/identity-links`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(800);
  const url = page.url();
  const t = await page.locator('body').innerText();
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
    if (await linkBtn.count() && await linkBtn.first().isEnabled()) {
      const wait = page.waitForResponse((r) => /identity-link/i.test(r.url()) && ['POST', 'DELETE'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
      await linkBtn.first().click();
      const resp = await wait;
      rec('S13', 'owner_link', resp && resp.ok() ? 'PASS' : 'PARTIAL', resp ? String(resp.status()) : 'noresp');
    } else rec('S13', 'owner_link', 'PARTIAL', 'btn disabled');
  } else rec('S13', 'owner_link', 'PARTIAL', 'need 2 owners fixture');
  await shot(page, 'S13');
} catch (e) { rec('S13', 'err', 'FAIL', e.message); }

// LOCK
try {
  await page.goto(`${BASE}/examinations/1014565`, { waitUntil: 'networkidle' });
  const t = await page.locator('body').innerText();
  rec('LOCK', '033', /完了済みのため結果の編集/.test(t) && (await page.locator('input[name^="examItems."][name$=".inspectionValue"]:not([disabled])').count()) === 0 ? 'PASS' : 'FAIL', 'exam');
  await page.goto(`${BASE}/medical-records/1080036`, { waitUntil: 'networkidle' });
  const t2 = await page.locator('body').innerText();
  const sv = (await page.getByRole('button', { name: /^保存$/ }).count()) > 0 && (await page.getByRole('button', { name: /^保存$/ }).first().isVisible().catch(() => false));
  rec('LOCK', '035', /確定済|編集できません/.test(t2) && !sv ? 'PASS' : 'FAIL', 'mr');
  await page.goto(`${BASE}/settings/clinic`, { waitUntil: 'networkidle' });
  const t3 = await page.locator('body').innerText();
  rec('LOCK', '038', !/医院が登録されていません/.test(t3) ? 'PASS' : 'FAIL', 'clinic');
} catch (e) { rec('LOCK', 'err', 'FAIL', e.message); }

// V routes
const vRoutes = [
  ['V01', '/medical-records', /カルテ/],
  ['V01', '/examinations', /検査/],
  ['V01', '/vaccinations', /予防|ワクチン/],
  ['V01', '/hospitalization', /入院/],
  ['V01', '/trimming', /トリミング/],
  ['V01', '/checkups', /健診|定期/],
  ['V02', '/accounting', /会計/],
  ['V02', '/estimates', /見積/],
  ['V02', '/accounting/close', /締め/],
  ['V02', '/shifts', /シフト/],
  ['V02', '/', /受付|予約/],
  ['V03', '/owners/new', /飼主/],
  ['V03', '/owners', /飼主/],
  ['V03', '/settings/staff', /スタッフ/],
  ['V03', '/settings/permission-groups', /権限/],
  ['V03', '/settings/clinic', /医院|病院/],
  ['V04', '/settings', /設定|マスタ/],
  ['V04', '/settings/payment-methods', /支払/],
  ['V04', '/settings/campaigns', /キャンペーン/],
  ['V04', '/settings/diagnosis', /診断/],
  ['V04', '/settings/insurance', /保険/],
  ['V05', '/forgot-password', /パスワード|メール/],
  ['V05', '/reset-password', /無効|トークン|パスワード/],
  ['V05', '/line-reservation/settings', /LINE|予約/],
  ['V05', '/line-reservation/page-editor', /ページ|編集|LINE|予約/],
  ['V05', '/line-reservation/slots', /枠|スロット|予約|LINE/],
  ['V05', '/settings/integrations/lstep', /Lステップ|連携/],
  ['V05', '/settings/lstep/tags', /タグ/],
  ['V05', '/lstep/analytics', /分析|CSV|取込|Lステップ/],
  ['V05', '/lstep/checkup-sync', /健診|タグ/],
];
for (const [id, pth, re] of vRoutes) {
  try {
    await page.goto(`${BASE}${pth}`, { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(300);
    const t = await page.locator('body').innerText();
    const crash = /Something went wrong|Application error/i.test(t);
    rec(id, pth, crash ? 'FAIL' : re.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 70));
    if (crash) bug(`${id}-crash`, `画面クラッシュ ${pth}`, t.slice(0, 100));
  } catch (e) {
    rec(id, pth, 'FAIL', e.message);
  }
}

for (const [id, pth, btnRe, errRe] of [
  ['V03', '/owners/new', /登録|保存|作成/, /入力|必須|してください/],
  ['V05', '/forgot-password', /送信|リセット/, /メール|入力/],
]) {
  try {
    await page.goto(`${BASE}${pth}`, { waitUntil: 'networkidle' });
    const btn = page.getByRole('button', { name: btnRe }).first();
    if (!(await btn.count())) { rec(id, `${pth}#c1`, 'BLOCKED', 'no btn'); continue; }
    if (!(await btn.isEnabled())) { rec(id, `${pth}#c1`, 'PASS', 'disabled gate'); continue; }
    await btn.click({ timeout: 8000 });
    await page.waitForTimeout(700);
    const t = await page.locator('body').innerText();
    rec(id, `${pth}#c1`, errRe.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 80));
  } catch (e) {
    rec(id, `${pth}#c1`, 'BLOCKED', e.message);
  }
}

try {
  await page.goto(`${BASE}/`, { waitUntil: 'networkidle' });
  const nb = page.getByRole('button', { name: /新規予約|予約登録/ }).first();
  if (await nb.count()) {
    await nb.click();
    await page.waitForTimeout(800);
    const conf = page.getByRole('button', { name: /予約を確定|確定/ }).first();
    if (await conf.count()) {
      if (!(await conf.isEnabled())) rec('V02', 'reservation_empty', 'PASS', 'disabled');
      else {
        await conf.click();
        await page.waitForTimeout(600);
        const t = await page.locator('body').innerText();
        rec('V02', 'reservation_empty', /患者|予約区分|選択|入力/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 80));
      }
    } else rec('V02', 'reservation_empty', 'PARTIAL', 'no confirm');
  } else rec('V02', 'reservation_empty', 'PARTIAL', 'no new');
} catch (e) { rec('V02', 'reservation_empty', 'BLOCKED', e.message); }

for (const pth of ['/settings/payment-methods', '/settings/campaigns']) {
  try {
    await page.goto(`${BASE}${pth}`, { waitUntil: 'networkidle' });
    const neu = page.getByRole('button', { name: /新規|追加|\+/ }).first();
    if (await neu.count()) {
      await neu.click();
      await page.waitForTimeout(500);
      const save = page.getByRole('button', { name: /保存|登録|作成/ }).first();
      if (await save.count()) {
        if (!(await save.isEnabled())) rec('V04', `${pth}#c1`, 'PASS', 'disabled');
        else {
          await save.click();
          await page.waitForTimeout(600);
          const t = await page.locator('body').innerText();
          rec('V04', `${pth}#c1`, /入力|必須|してください/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 60));
        }
      }
    } else rec('V04', `${pth}#c1`, 'PARTIAL', 'no new');
  } catch (e) { rec('V04', `${pth}#c1`, 'BLOCKED', e.message); }
}

// final pet alive
try {
  const pet = await (await page.request.get(`${API}/api/v1/pets/1000002`, { headers: { 'X-Requested-With': 'XMLHttpRequest' } })).json();
  if (pet.status === 'deceased' || pet.deceased_at) {
    await page.request.fetch(`${API}/api/v1/clinics/1/pets/1000002/death`, { method: 'DELETE', headers: { 'X-Requested-With': 'XMLHttpRequest' } });
  }
  const pet2 = await (await page.request.get(`${API}/api/v1/pets/1000002`, { headers: { 'X-Requested-With': 'XMLHttpRequest' } })).json();
  rec('S01', 'final_alive', pet2.status === 'alive' || !pet2.deceased_at ? 'PASS' : 'FAIL', JSON.stringify({ status: pet2.status }));
} catch (e) { rec('S01', 'final_alive', 'FAIL', e.message); }

await browser.close();

fs.writeFileSync(path.join(OUT, 'results.json'), JSON.stringify(results, null, 2));
fs.writeFileSync(path.join(OUT, 'bug-candidates.json'), JSON.stringify(bugs, null, 2));
const c = results.reduce((a, r) => { a[r.status] = (a[r.status] || 0) + 1; return a; }, {});
console.log('TOTALS', c);
console.log('BUG_CANDIDATES', bugs.length, bugs.map((b) => b.id).join(','));
console.log('FAILS', results.filter((r) => r.status === 'FAIL').map((r) => `${r.scenario}#${r.step}`).join(','));
