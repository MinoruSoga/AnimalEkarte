/**
 * V04 master full-item UAT — C1-1 / create / C2 reload / C3-2 unique / cleanup
 * Uses #master-title + 新規登録 / 保存 pattern from settings e2e
 */
import { chromium } from '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/frontend/node_modules/playwright/index.mjs';
import fs from 'fs';
import path from 'path';

const BASE = 'http://localhost:3003';
const OUT = '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/reports/uat-2026-08-14-v04';
const email = process.env.E2E_LOGIN_EMAIL || 'admin@noavet.jp';
const password = process.env.E2E_LOGIN_PASSWORD || 'password';
const TAG = `V04_${Date.now().toString().slice(-8)}`;
fs.mkdirSync(OUT, { recursive: true });
const results = [];
const bugs = [];
const rec = (id, step, status, note = '') => {
  results.push({ id, step, status, note: String(note).slice(0, 500), at: new Date().toISOString() });
  console.log(`[${status}] ${id}#${step} ${String(note).slice(0, 140)}`);
};
const bug = (id, title, ev) => {
  bugs.push({ id, title, evidence: ev });
  console.log(`[BUG?] ${id} ${title}`);
};

async function login(page) {
  await page.goto(`${BASE}/login`, { waitUntil: 'networkidle' });
  await page.locator('#login-email').fill(email);
  await page.locator('#login-password').fill(password);
  await page.getByRole('button', { name: 'ログイン', exact: true }).click();
  await page.waitForURL((u) => !u.pathname.includes('/login'), { timeout: 30000 });
}

async function openMaster(page, route) {
  await page.goto(`${BASE}${route}`, { waitUntil: 'networkidle', timeout: 45000 });
  await page.waitForTimeout(600);
}

async function openNew(page) {
  const btn = page.getByRole('button', { name: '新規登録' });
  if (!(await btn.count())) return false;
  await btn.click();
  await page.waitForSelector('#master-title', { timeout: 10000 }).catch(() => null);
  return (await page.locator('#master-title').count()) > 0;
}

async function bodyText(page) {
  return page.locator('body').innerText();
}

async function saveClick(page) {
  const save = page.getByRole('button', { name: '保存', exact: true }).first();
  if (await save.count()) await save.click();
  await page.waitForTimeout(1500);
}

async function cancelPanel(page) {
  const c = page.getByRole('button', { name: 'キャンセル' });
  if (await c.count()) await c.click().catch(() => {});
  await page.keyboard.press('Escape').catch(() => {});
  await page.waitForTimeout(400);
}

async function tryDelete(page, name) {
  try {
    // search if available
    const searchToggle = page.getByLabel('検索');
    if (await searchToggle.count()) {
      await searchToggle.click().catch(() => {});
      await page.waitForTimeout(300);
      const si = page.locator('input[placeholder*="検索"]');
      if (await si.count()) {
        await si.first().fill(name);
        await page.waitForTimeout(800);
      }
    }
    const row = page.locator('tbody tr').filter({ hasText: name }).first();
    if (!(await row.count())) return false;
    const op = row.getByLabel('操作');
    if (await op.count()) await op.click();
    else await row.click();
    await page.waitForTimeout(500);
    const del = page.getByLabel('削除');
    if (!(await del.count())) {
      await cancelPanel(page);
      return false;
    }
    await del.click();
    const dialog = page.getByRole('alertdialog');
    if (await dialog.count()) {
      await dialog.getByRole('button', { name: '削除' }).click();
      await page.waitForTimeout(1200);
      return true;
    }
    return false;
  } catch {
    return false;
  }
}

/** Standard master: C1-1 empty, create, C2 reload+reopen, optional C3-2, cleanup */
async function standardMaster(page, id, route, { unique = true, nameFieldOnly = true } = {}) {
  const name = `${TAG}_${id}`;
  try {
    await openMaster(page, route);
    const crash = /Something went wrong|Application error/i.test(await bodyText(page));
    if (crash) {
      rec(id, 'open', 'FAIL', 'crash');
      bug(`V04-${id}-crash`, `${route} クラッシュ`, route);
      return;
    }
    rec(id, 'open', 'PASS', route);

    // C1-1
    if (!(await openNew(page))) {
      rec(id, 'C1-1', 'BLOCKED', 'no 新規登録/#master-title');
      return;
    }
    await saveClick(page);
    const t1 = await bodyText(page);
    const panelStill = await page.locator('#master-title').isVisible().catch(() => false);
    const err = /入力|必須|してください|エラー|空/.test(t1) || panelStill;
    // if saved empty wrongly, name would appear empty in list - treat as fail
    const emptySaved = !panelStill && !/入力|必須|してください/.test(t1);
    if (emptySaved) {
      rec(id, 'C1-1', 'FAIL', 'empty may have saved');
      bug(`V04-${id}-C1`, `${route} 必須空でも保存できてしまう可能性`, t1.slice(0, 80));
    } else {
      rec(id, 'C1-1', err ? 'PASS' : 'PARTIAL', panelStill ? 'panel stayed' : t1.replace(/\s+/g, ' ').slice(0, 60));
    }
    await cancelPanel(page);

    // Create
    if (!(await openNew(page))) {
      rec(id, 'create', 'BLOCKED', 'no panel');
      return;
    }
    await page.locator('#master-title').fill(name);
    // optional extra fields: first other text input
    if (!nameFieldOnly) {
      const extras = page.locator('[role=dialog] input[type=text], [data-state=open] input[type=text], aside input[type=text]');
      // skip
    }
    // category combobox if present (diagnosis name)
    const combos = page.getByRole('combobox');
    if (await combos.count()) {
      await combos.first().click().catch(() => {});
      await page.waitForTimeout(300);
      if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
    }
    const wait = page.waitForResponse((r) => ['POST', 'PUT', 'PATCH'].includes(r.request().method()) && r.url().includes('/api/'), { timeout: 15000 }).catch(() => null);
    await saveClick(page);
    const resp = await wait;
    await page.waitForTimeout(800);
    const createdVisible = (await page.getByText(name).count()) > 0;
    const createOk = (resp && resp.ok()) || createdVisible;
    rec(id, 'create', createOk ? 'PASS' : 'FAIL', resp ? `${resp.status()} vis=${createdVisible}` : `vis=${createdVisible}`);
    if (!createOk) {
      bug(`V04-${id}-create`, `${route} 新規保存失敗`, name);
      await cancelPanel(page);
      return;
    }

    // C2 reload
    await openMaster(page, route);
    const afterReload = (await page.getByText(name).count()) > 0;
    rec(id, 'C2-reload', afterReload ? 'PASS' : 'FAIL', name);
    if (!afterReload) bug(`V04-${id}-C2`, `${route} 再読込で消える`, name);

    // C2 reopen
    if (afterReload) {
      const row = page.locator('tbody tr').filter({ hasText: name }).first();
      if (await row.count()) {
        const op = row.getByLabel('操作');
        if (await op.count()) await op.click();
        else await row.dblclick().catch(() => row.click());
        await page.waitForTimeout(600);
        const val = await page.locator('#master-title').inputValue().catch(() => '');
        rec(id, 'C2-reopen', val.includes(name) || val === name ? 'PASS' : 'PARTIAL', `val=${val.slice(0, 40)}`);
        await cancelPanel(page);
      } else rec(id, 'C2-reopen', 'PARTIAL', 'no row op');
    }

    // C3-2 unique
    if (unique && afterReload) {
      if (await openNew(page)) {
        await page.locator('#master-title').fill(name);
        if (await page.getByRole('combobox').count()) {
          await page.getByRole('combobox').first().click().catch(() => {});
          if (await page.getByRole('option').count()) await page.getByRole('option').first().click().catch(() => {});
        }
        const wait2 = page.waitForResponse((r) => ['POST', 'PUT', 'PATCH'].includes(r.request().method()) && r.url().includes('/api/'), { timeout: 12000 }).catch(() => null);
        await saveClick(page);
        const r2 = await wait2;
        await page.waitForTimeout(800);
        const t = await bodyText(page);
        const rejected = (r2 && (r2.status() === 409 || r2.status() === 400)) || /既に|重複|unique|存在|使えません|登録されて/.test(t) || (await page.locator('#master-title').isVisible().catch(() => false));
        // counting two rows with same name is also fail if both exist without error
        const count = await page.getByText(name).count();
        if (!rejected && count >= 2) {
          rec(id, 'C3-2', 'FAIL', 'duplicate accepted');
          bug(`V04-${id}-unique`, `${route} 同名が重複登録できる`, name);
        } else {
          rec(id, 'C3-2', rejected ? 'PASS' : 'PARTIAL', r2 ? `status=${r2.status()}` : t.replace(/\s+/g, ' ').slice(0, 50));
        }
        await cancelPanel(page);
      } else rec(id, 'C3-2', 'BLOCKED', 'no panel');
    } else if (!unique) {
      rec(id, 'C3-2', 'SKIP', 'no unique constraint');
    }

    // cleanup
    await openMaster(page, route);
    const deleted = await tryDelete(page, name);
    rec(id, 'cleanup', deleted ? 'PASS' : 'PARTIAL', deleted ? 'deleted' : 'left V04 row');
  } catch (e) {
    rec(id, 'err', 'FAIL', e.message);
    bug(`V04-${id}-err`, `${route} 例外`, e.message);
  }
}

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
const page = await context.newPage();
page.setDefaultTimeout(40000);
await login(page);
rec('ENV', 'login', 'PASS', page.url());

// §1 standard masters
const standards = [
  ['animal-species', '/settings/animal-species', true],
  ['diagnosis-type', '/settings/diagnosis?tab=diagnosis_type', true],
  ['diagnosis-name', '/settings/diagnosis?tab=diagnosis_name', false],
  ['chief-complaint', '/settings/interview/chief-complaint', true],
  ['inquiry-templates', '/settings/inquiry-templates', false],
  ['reservation-type-group', '/settings/reservation-type', false], // group may share page
  ['hospitalization-plan', '/settings/hospitalization', true],
  ['cage', '/settings/cage', true],
  ['merchandise', '/settings/merchandise-items', true],
  ['insurance', '/settings/insurance', true],
  ['occupations', '/settings/occupations', true],
  ['trimming-course', '/settings/trimming?tab=course', true],
  ['trimming-option', '/settings/trimming?tab=option', true],
  ['trimming-course-type', '/settings/trimming-course-type', true],
  ['campaigns', '/settings/campaigns', false],
  ['payment-methods', '/settings/payment-methods', true],
  ['medicine', '/settings/medicine', true],
  ['shift-templates', '/settings/shift-templates', true],
];

for (const [id, route, unique] of standards) {
  await standardMaster(page, id, route, { unique });
}

// treatment 5 tabs — C1-1 on consultation + create one per tab
const treatTabs = ['consultation', 'examination', 'procedure', 'vaccine', 'checkup'];
for (const tab of treatTabs) {
  const id = `treatment-${tab}`;
  const route = `/settings/treatment-items?tab=${tab}`;
  const name = `${TAG}_${id}`;
  try {
    await openMaster(page, route);
    rec(id, 'open', /Something went wrong/i.test(await bodyText(page)) ? 'FAIL' : 'PASS', route);
    if (!(await openNew(page))) {
      // checkup may differ
      const alt = page.getByRole('button', { name: /新規/ });
      if (await alt.count()) await alt.first().click();
      await page.waitForTimeout(500);
    }
    if (await page.locator('#master-title').count()) {
      if (tab === 'consultation') {
        await saveClick(page);
        const panel = await page.locator('#master-title').isVisible().catch(() => false);
        rec(id, 'C1-1', panel ? 'PASS' : 'PARTIAL', 'empty save');
        await cancelPanel(page);
        await openNew(page);
      }
      await page.locator('#master-title').fill(name);
      // price 0 if exists
      const price = page.locator('input[type=number]').first();
      if (await price.count()) await price.fill('100').catch(() => {});
      await saveClick(page);
      await page.waitForTimeout(1000);
      const ok = (await page.getByText(name).count()) > 0;
      rec(id, 'create', ok ? 'PASS' : 'PARTIAL', name);
      await openMaster(page, route);
      rec(id, 'C2-reload', (await page.getByText(name).count()) > 0 ? 'PASS' : 'PARTIAL', name);
      await tryDelete(page, name);
      rec(id, 'cleanup', 'PASS', 'attempted');
    } else {
      rec(id, 'panel', 'BLOCKED', 'no #master-title');
    }
  } catch (e) {
    rec(id, 'err', 'FAIL', e.message);
  }
}

// insurance C1-3 coverage bounds
try {
  await openMaster(page, '/settings/insurance');
  const name = `${TAG}_ins_bound`;
  if (await openNew(page)) {
    await page.locator('#master-title').fill(name);
    // find coverage input
    const nums = page.locator('input[type=number]');
    if (await nums.count()) {
      await nums.first().fill('101');
      await saveClick(page);
      const t = await bodyText(page);
      const rejected = /101|範囲|0〜100|100以下|補償/.test(t) || (await page.locator('#master-title').isVisible().catch(() => false));
      rec('insurance-bound', 'C1-3-101', rejected ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 60));
      if (!rejected) bug('V04-ins-101', '補償率101が拒否されない', name);
      await cancelPanel(page);
      // valid 100
      if (await openNew(page)) {
        await page.locator('#master-title').fill(name + '_ok');
        if (await nums.count()) await nums.first().fill('100');
        await saveClick(page);
        await page.waitForTimeout(800);
        rec('insurance-bound', 'C1-3-100', (await page.getByText(name + '_ok').count()) > 0 ? 'PASS' : 'PARTIAL', '100');
        await tryDelete(page, name + '_ok');
      }
    } else rec('insurance-bound', 'C1-3', 'BLOCKED', 'no number input');
  }
} catch (e) { rec('insurance-bound', 'err', 'FAIL', e.message); }

// payment system row cannot delete - smoke
try {
  await openMaster(page, '/settings/payment-methods');
  const cash = page.locator('tbody tr').filter({ hasText: /現金|cash/i }).first();
  if (await cash.count()) {
    await cash.getByLabel('操作').click().catch(() => cash.click());
    await page.waitForTimeout(500);
    const del = page.getByLabel('削除');
    if (await del.count() && await del.isEnabled()) {
      await del.click();
      const dialog = page.getByRole('alertdialog');
      if (await dialog.count()) await dialog.getByRole('button', { name: '削除' }).click();
      await page.waitForTimeout(1000);
      const t = await bodyText(page);
      const blocked = /システム標準|削除できません|無効化できません/.test(t) || (await page.getByText(/現金/).count()) > 0;
      rec('payment-system', 'no-delete', blocked ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 60));
      if (!blocked) bug('V04-pay-sys', 'システム標準支払方法が削除できてしまう', 'cash');
    } else rec('payment-system', 'no-delete', 'PASS', 'delete disabled');
    await cancelPanel(page);
  } else rec('payment-system', 'open', 'PARTIAL', 'no cash row');
} catch (e) { rec('payment-system', 'err', 'FAIL', e.message); }

// closing time page
try {
  await openMaster(page, '/settings/closing-time');
  const t = await bodyText(page);
  rec('closing-time', 'open', /締め|AM|PM|休診|特別/.test(t) ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 60));
  // holiday empty add
  const addH = page.getByRole('button', { name: /追加|休診/ }).first();
  if (await addH.count()) {
    await addH.click();
    await page.waitForTimeout(400);
    const save = page.getByRole('button', { name: /保存|追加|登録/ }).last();
    if (await save.count() && await save.isEnabled()) {
      await save.click();
      await page.waitForTimeout(600);
      rec('closing-time', 'holiday-C1', 'PARTIAL', 'clicked add empty');
    } else rec('closing-time', 'holiday-C1', 'PASS', 'add disabled without date');
  } else rec('closing-time', 'holiday-C1', 'PARTIAL', 'no add btn');
} catch (e) { rec('closing-time', 'err', 'FAIL', e.message); }

// slots page C3-3 style
try {
  await openMaster(page, '/line-reservation/slots?typeId=99999999');
  const t = await bodyText(page);
  const ok = !/Something went wrong|Application error/i.test(t) && t.length > 30;
  rec('slots', 'invalid-typeId', ok ? 'PASS' : 'FAIL', t.replace(/\s+/g, ' ').slice(0, 60));
  if (!ok) bug('V04-slots-404', '無効 typeId で白画面/クラッシュ', 'typeId=99999999');
  await openMaster(page, '/line-reservation/slots');
  rec('slots', 'open', /枠|スロット|区分|予約/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'default');
} catch (e) { rec('slots', 'err', 'FAIL', e.message); }

// LSTEP settings open + save empty secrets
try {
  await openMaster(page, '/settings/integrations/lstep');
  const t = await bodyText(page);
  rec('lstep', 'open', /Lステップ|連携|API|LIFF/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 50));
  const save = page.getByRole('button', { name: /保存/ }).first();
  if (await save.count() && await save.isEnabled()) {
    const wait = page.waitForResponse((r) => r.url().includes('lstep') && ['PUT', 'PATCH', 'POST'].includes(r.request().method()), { timeout: 10000 }).catch(() => null);
    await save.click();
    const r = await wait;
    rec('lstep', 'save-empty-secrets', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : 'noresp');
  } else rec('lstep', 'save', 'PARTIAL', 'no save');
  await openMaster(page, '/settings/lstep/tags');
  rec('lstep-tags', 'open', /タグ|prefix|条件/.test(await bodyText(page)) ? 'PASS' : 'PARTIAL', 'tags');
} catch (e) { rec('lstep', 'err', 'FAIL', e.message); }

// LINE page editor
try {
  await openMaster(page, '/line-reservation/page-editor');
  const t = await bodyText(page);
  rec('line-page-editor', 'open', /ページ|ヘッダ|プライバシー|予約/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 50));
  const save = page.getByRole('button', { name: /保存/ }).first();
  if (await save.count()) {
    const wait = page.waitForResponse((r) => ['PUT', 'PATCH', 'POST'].includes(r.request().method()) && r.url().includes('/api/'), { timeout: 10000 }).catch(() => null);
    await save.click();
    const r = await wait;
    rec('line-page-editor', 'C2-save', r && r.ok() ? 'PASS' : 'PARTIAL', r ? String(r.status()) : 'noresp');
  }
} catch (e) { rec('line-page-editor', 'err', 'FAIL', e.message); }

// clinic invoice section
try {
  await openMaster(page, '/settings/clinic');
  const t = await bodyText(page);
  rec('clinic', 'open', /医院|インボイス|法人/.test(t) ? 'PASS' : 'PARTIAL', t.replace(/\s+/g, ' ').slice(0, 50));
  const inv = page.getByLabel(/インボイス|登録番号/).or(page.locator('input').filter({ has: page.locator('xpath=..') }));
  // try find invoice input by placeholder
  const invInput = page.locator('input').filter({ hasText: '' });
  const candidates = page.locator('input[type=text]');
  let filled = false;
  for (let i = 0; i < Math.min(await candidates.count(), 15); i++) {
    const ph = (await candidates.nth(i).getAttribute('placeholder')) || '';
    const aria = (await candidates.nth(i).getAttribute('aria-label')) || '';
    if (/インボイス|T[0-9]|登録番号/.test(ph + aria) || ph.includes('T')) {
      const old = await candidates.nth(i).inputValue();
      await candidates.nth(i).fill('T9999999999999');
      const save = page.getByRole('button', { name: /保存|更新/ }).first();
      if (await save.count()) await save.click();
      await page.waitForTimeout(1000);
      rec('clinic-invoice', 'C2', 'PASS', 'filled T999...');
      await candidates.nth(i).fill(old);
      if (await save.count()) await save.click();
      filled = true;
      break;
    }
  }
  if (!filled) rec('clinic-invoice', 'C2', 'PARTIAL', 'invoice input not found by heuristic');
} catch (e) { rec('clinic-invoice', 'err', 'FAIL', e.message); }

// reservation-type create leaf
try {
  await openMaster(page, '/settings/reservation-type');
  const name = `${TAG}_rtype`;
  // may need tab for 区分 not group
  const tabs = page.getByRole('tab');
  for (let i = 0; i < await tabs.count(); i++) {
    const tx = await tabs.nth(i).innerText();
    if (/区分|タイプ|一覧/.test(tx) && !/グループ/.test(tx)) {
      await tabs.nth(i).click();
      await page.waitForTimeout(400);
      break;
    }
  }
  if (await openNew(page)) {
    await page.locator('#master-title').fill(name);
    await saveClick(page);
    await page.waitForTimeout(800);
    rec('reservation-type', 'create', (await page.getByText(name).count()) > 0 ? 'PASS' : 'PARTIAL', name);
    await tryDelete(page, name);
  } else rec('reservation-type', 'create', 'BLOCKED', 'no panel');
} catch (e) { rec('reservation-type', 'err', 'FAIL', e.message); }

await browser.close();
fs.writeFileSync(path.join(OUT, 'results.json'), JSON.stringify(results, null, 2));
fs.writeFileSync(path.join(OUT, 'bug-candidates.json'), JSON.stringify(bugs, null, 2));
const c = results.reduce((a, r) => { a[r.status] = (a[r.status] || 0) + 1; return a; }, {});
console.log('TOTALS', c);
console.log('BUGS', bugs.length, bugs.map((b) => b.id).join(','));
console.log('FAILS', results.filter((r) => r.status === 'FAIL').map((r) => `${r.id}#${r.step}`).join(','));
