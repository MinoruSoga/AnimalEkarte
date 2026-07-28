import { expect, test } from '@playwright/test';
import type { BrowserContext, Page } from '@playwright/test';

import { loginAsDemoAdmin } from './helpers/auth';

const REPORT_OWNER_ID = process.env.OWNER_REPORT_E2E_OWNER_ID ?? '1';
const REPORT_PET_ID = process.env.OWNER_REPORT_E2E_PET_ID ?? '1';
const REPORT_URL = `/owners/${REPORT_OWNER_ID}/report?petId=${REPORT_PET_ID}`;
const SHOT_DIR = process.env.SHOT_DIR ?? 'test-results/owner-report';
const SCROLL_TOLERANCE = 2;

const PANEL_NAMES = [
  '診療前の確認',
  '今日の来院',
  '次の行動',
  '前回診療',
  '基本情報',
  '種類別履歴',
] as const;

interface PanelMetric {
  overflowY: string;
  clientHeight: number;
  scrollHeight: number;
}

async function gotoReport(context: BrowserContext, width: number, height: number) {
  const page = await context.newPage();
  await page.setViewportSize({ width, height });
  await page.goto(REPORT_URL, { waitUntil: 'domcontentloaded' });
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 30000 });
  await expect(page.getByRole('region', { name: '診療前の確認' })).toBeVisible({ timeout: 30000 });
  return page;
}

async function panelMetrics(page: Page): Promise<PanelMetric[]> {
  return page.evaluate(() => {
    return Array.from(document.querySelectorAll<HTMLElement>('[data-owner-report-scroll]')).map(
      (element) => {
        const style = getComputedStyle(element);
        return {
          overflowY: style.overflowY,
          clientHeight: element.clientHeight,
          scrollHeight: element.scrollHeight,
        };
      },
    );
  });
}

async function expectFixedViewport(page: Page) {
  const documentMetric = await page.evaluate(() => {
    window.scrollTo(0, 5000);
    return {
      scrollHeight: document.scrollingElement?.scrollHeight ?? 0,
      scrollWidth: document.scrollingElement?.scrollWidth ?? 0,
      innerHeight: window.innerHeight,
      innerWidth: window.innerWidth,
      scrollY: window.scrollY,
      scrollX: window.scrollX,
    };
  });
  expect(documentMetric.scrollHeight).toBeLessThanOrEqual(
    documentMetric.innerHeight + SCROLL_TOLERANCE,
  );
  expect(documentMetric.scrollWidth).toBeLessThanOrEqual(
    documentMetric.innerWidth + SCROLL_TOLERANCE,
  );
  expect(documentMetric.scrollY).toBe(0);
  expect(documentMetric.scrollX).toBe(0);
}

test.describe('#158 飼主レポート 1画面ブリーフィング', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    const loginPage = await context.newPage();
    await loginAsDemoAdmin(loginPage);
    await loginPage.close();
  });

  test.afterAll(async () => {
    await context?.close();
  });

  for (const viewport of [
    { name: '1440x900', width: 1440, height: 900 },
    { name: '1280x720', width: 1280, height: 720 },
    { name: '390x844', width: 390, height: 844 },
    { name: '844x390', width: 844, height: 390 },
    { name: '568x320', width: 568, height: 320 },
    { name: '320x480', width: 320, height: 480 },
  ]) {
    test(`${viewport.name}: 6領域をページスクロールなしで同時提示する`, async () => {
      const page = await gotoReport(context, viewport.width, viewport.height);
      try {
        await expectFixedViewport(page);

        for (const panelName of PANEL_NAMES) {
          await expect(page.getByRole('region', { name: panelName })).toBeVisible();
          await expect(page.getByRole('heading', { name: panelName })).toBeVisible();
        }
        await expect(page.getByRole('tablist')).toHaveCount(0);
        await expect(page.getByRole('tab')).toHaveCount(0);
        await expect(page.getByRole('tabpanel')).toHaveCount(0);

        const metrics = await panelMetrics(page);
        expect(metrics).toHaveLength(6);
        for (const metric of metrics) {
          expect(['auto', 'scroll']).toContain(metric.overflowY);
          expect(metric.clientHeight).toBeGreaterThan(0);
          expect(metric.clientHeight).toBeLessThanOrEqual(viewport.height);
        }

        await page.screenshot({ path: `${SHOT_DIR}/owner-report-${viewport.name}.png` });
      } finally {
        await page.close();
      }
    });
  }

  test('履歴は縦を種類、横を日付にし、薬・予防接種・処置を別行表示する', async () => {
    const page = await gotoReport(context, 1440, 900);
    try {
      const history = page.getByRole('table', {
        name: '診療履歴を種類別に分け、日付の新しい順に左から表示',
      });
      await expect(history).toBeVisible();
      for (const kind of ['診療', '検査', '薬・処方', '予防接種', '処置', 'ケア']) {
        await expect(history.getByRole('rowheader', { name: new RegExp(kind) })).toBeVisible();
      }
    } finally {
      await page.close();
    }
  });

  test('短い画面で内容が溢れても内部だけスクロールし、ページは動かない', async () => {
    const page = await gotoReport(context, 568, 320);
    try {
      const before = await panelMetrics(page);
      const overflowingIndex = before.findIndex(
        (metric) => metric.scrollHeight > metric.clientHeight + 1,
      );
      expect(overflowingIndex).toBeGreaterThanOrEqual(0);

      const result = await page.evaluate((index) => {
        const panels = Array.from(
          document.querySelectorAll<HTMLElement>('[data-owner-report-scroll]'),
        );
        const panel = panels[index];
        panel.scrollTop = panel.scrollHeight;
        return { panelScrollTop: panel.scrollTop, windowScrollY: window.scrollY };
      }, overflowingIndex);

      expect(result.panelScrollTop).toBeGreaterThan(0);
      expect(result.windowScrollY).toBe(0);
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test('ペット選択はタブを使わず、ページ遷移なしでpetIdを同期する', async () => {
    const page = await gotoReport(context, 1440, 900);
    try {
      const selector = page.getByRole('combobox', { name: 'ペット切替' });
      const options = selector.getByRole('option');
      const count = await options.count();
      if (count < 2) {
        test.info().annotations.push({
          type: 'note',
          description: `対象飼主のペットは${count}件のため、切替はRTLテストで担保する。`,
        });
        return;
      }

      const pathname = new URL(page.url()).pathname;
      const nextPetId = await options.nth(1).getAttribute('value');
      if (!nextPetId) throw new Error('2頭目のpetIdを取得できません');
      await selector.selectOption(nextPetId);

      await expect.poll(() => new URL(page.url()).pathname).toBe(pathname);
      await expect.poll(() => new URL(page.url()).searchParams.get('petId')).toBe(nextPetId);
      for (const panelName of PANEL_NAMES) {
        await expect(page.getByRole('region', { name: panelName })).toBeVisible();
      }
    } finally {
      await page.close();
    }
  });
});
