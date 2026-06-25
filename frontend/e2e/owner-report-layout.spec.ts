import { test, expect } from '@playwright/test';
import type { BrowserContext } from '@playwright/test';
import { loginAsDemoAdmin } from './helpers/auth';

// #158 飼主レポート 高密度レイアウト回帰テスト。
// 目的: デスクトップでページ全体をスクロールさせず、各履歴パネルが境界内スクロールすることを
//       実ブラウザの DOM 計測で証明する（jsdom ではレイアウトが無いため検証できない）。
//
// Seed: owner id=1 「林 文明」, pet id=1 「Iris(イリス)」 at clinic_id=1 (003_seed_demo.sql)。
// 認証: admin@noavet.jp / password (system_admin)。
//
// スクリーンショット出力先は SHOT_DIR 環境変数で差し替え可能（リポジトリを汚さないため /tmp 等にマウント）。

const REPORT_URL = '/owners/1/report?petId=1';
const SHOT_DIR = process.env.SHOT_DIR ?? 'test-results/owner-report';

// 上部固定バーを除いた本文がページスクロールしないことを許容する誤差(px)。
const SCROLL_TOLERANCE = 2;

interface PanelMetric {
  overflowY: string;
  clientHeight: number;
  scrollHeight: number;
}

async function gotoReport(context: BrowserContext, width: number, height: number) {
  const page = await context.newPage();
  await page.setViewportSize({ width, height });
  await page.goto(REPORT_URL, { waitUntil: 'domcontentloaded' });
  // 飼主名(h1) と最初のパネル見出しが出るまで待つ。
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 30000 });
  await expect(page.getByRole('heading', { name: 'ペット詳細' })).toBeVisible({ timeout: 30000 });
  return page;
}

async function panelMetrics(page: import('@playwright/test').Page): Promise<PanelMetric[]> {
  return page.evaluate(() => {
    const bodies = Array.from(document.querySelectorAll('main [tabindex="0"]'));
    return bodies.map((el) => {
      const cs = getComputedStyle(el);
      return {
        overflowY: cs.overflowY,
        clientHeight: (el as HTMLElement).clientHeight,
        scrollHeight: (el as HTMLElement).scrollHeight,
      };
    });
  });
}

test.describe('#158 飼主レポート 高密度レイアウト', () => {
  let context: BrowserContext;

  test.beforeAll(async ({ browser }) => {
    context = await browser.newContext();
    const loginPage = await context.newPage();
    await loginAsDemoAdmin(loginPage);
    await loginPage.close();
  });

  test.afterAll(async () => {
    await context.close();
  });

  for (const vp of [
    { name: '1440x900', width: 1440, height: 900 },
    { name: '1920x1080', width: 1920, height: 1080 },
  ]) {
    test(`デスクトップ ${vp.name}: ページ全体はスクロールせず、8 パネルが境界内スクロールになる`, async () => {
      const page = await gotoReport(context, vp.width, vp.height);
      try {
        // --- AC1: ページ(body)レベルの縦スクロールが発生しない ---
        const doc = await page.evaluate(() => {
          // ページスクロールを試みても動かないことも確認する。
          window.scrollTo(0, 5000);
          return {
            scrollH: document.scrollingElement?.scrollHeight ?? 0,
            innerH: window.innerHeight,
            scrollYAfterAttempt: window.scrollY,
          };
        });
        expect(doc.scrollH).toBeLessThanOrEqual(doc.innerH + SCROLL_TOLERANCE);
        expect(doc.scrollYAfterAttempt).toBe(0);

        // --- AC2: 8 パネルそれぞれが overflow:auto かつビューポート内に収まる境界高さを持つ ---
        const metrics = await panelMetrics(page);
        expect(metrics).toHaveLength(8);
        for (const m of metrics) {
          expect(['auto', 'scroll']).toContain(m.overflowY);
          expect(m.clientHeight).toBeGreaterThan(0);
          expect(m.clientHeight).toBeLessThanOrEqual(doc.innerH);
        }

        // --- AC4: 8 セクション見出しが最初のビューポートに同時提示される ---
        for (const title of [
          'ペット詳細',
          '予防接種履歴',
          '健康診断（検査）履歴',
          '投薬履歴',
          '麻酔処置履歴',
          '手術処置履歴',
          '治療履歴',
          'トリミング履歴',
        ]) {
          await expect(page.getByRole('heading', { name: title })).toBeVisible();
        }

        await page.screenshot({ path: `${SHOT_DIR}/owner-report-${vp.name}.png` });
      } finally {
        await page.close();
      }
    });
  }

  test('短いビューポートで履歴が溢れても、内部スクロールに閉じ込められページは動かない (AC2/AC3)', async () => {
    // 高さを意図的に詰めて、データ量に依存せず必ずパネル内オーバーフローを起こす。
    const page = await gotoReport(context, 1280, 520);
    try {
      const before = await panelMetrics(page);
      // 少なくとも 1 つのパネルが内容過多でスクロール可能(scrollHeight > clientHeight)になる。
      const overflowingIndex = before.findIndex((m) => m.scrollHeight > m.clientHeight + 1);
      expect(overflowingIndex).toBeGreaterThanOrEqual(0);

      const result = await page.evaluate((idx) => {
        const bodies = Array.from(document.querySelectorAll('main [tabindex="0"]')) as HTMLElement[];
        const el = bodies[idx];
        el.scrollTop = el.scrollHeight; // パネル内を最下部までスクロール
        return { panelScrollTop: el.scrollTop, windowScrollY: window.scrollY };
      }, overflowingIndex);

      // パネル内はスクロールした / ページ(window)は動いていない。
      expect(result.panelScrollTop).toBeGreaterThan(0);
      expect(result.windowScrollY).toBe(0);

      // AC3: 内部スクロール後も飼主コンテキスト(飼主名 h1)は上部バーに残り視認できる。
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test('ペット切替はページ遷移せず ?petId= を同期する (AC5)', async () => {
    const page = await gotoReport(context, 1440, 900);
    try {
      const tabs = page.getByRole('tab');
      const count = await tabs.count();
      if (count < 2) {
        test.info().annotations.push({
          type: 'note',
          description: `飼主1 のペットは ${count} 件のため切替はライブ検証不可。RTL テストで担保。`,
        });
        return;
      }
      const pathBefore = new URL(page.url()).pathname;
      await tabs.nth(1).click();
      await expect(async () => {
        const url = new URL(page.url());
        expect(url.pathname).toBe(pathBefore); // ルート遷移していない
        expect(url.searchParams.get('petId')).not.toBeNull();
      }).toPass({ timeout: 5000 });
      // 切替後も全パネルは健在（末尾のトリミング履歴まで残る）
      await expect(page.getByRole('heading', { name: 'トリミング履歴' })).toBeVisible();
    } finally {
      await page.close();
    }
  });

  test('モバイル 390x844: 飼主と全セクションが利用可能（ページスクロール許容） (AC8)', async () => {
    const page = await gotoReport(context, 390, 844);
    try {
      // 飼主名 + 8 セクション見出しが（スクロールすれば）到達可能であること。
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
      for (const title of [
        'ペット詳細',
        '予防接種履歴',
        '健康診断（検査）履歴',
        '投薬履歴',
        '麻酔処置履歴',
        '手術処置履歴',
        '治療履歴',
        'トリミング履歴',
      ]) {
        await expect(page.getByRole('heading', { name: title })).toBeAttached();
      }

      // モバイルでは root(div) 自身がスクロールコンテナ（html/body は overflow:hidden 固定）。
      // 内容がビューポートを超えてスクロール可能で、末尾セクションまで実際に到達できることを確認する。
      const reach = await page.evaluate(() => {
        const main = document.querySelector('main') as HTMLElement;
        const root = main.parentElement as HTMLElement; // h-dvh の root = スクロールコンテナ
        root.scrollTop = root.scrollHeight; // 末尾までスクロール
        const sections = Array.from(document.querySelectorAll('main section')) as HTMLElement[];
        const last = sections[sections.length - 1];
        const rect = last.getBoundingClientRect();
        return {
          rootScrollable: root.scrollHeight > root.clientHeight + 1,
          lastVisibleInViewport: rect.top < window.innerHeight && rect.bottom > 0,
        };
      });
      // 内容過多でも root スクロールで全コンテンツに到達可能（祖先クリップで埋もれない）。
      expect(reach.rootScrollable).toBe(true);
      expect(reach.lastVisibleInViewport).toBe(true);

      // AC3(モバイル): スクロール後も飼主名は sticky ヘッダーでビューポート内に残る。
      await expect(page.getByRole('heading', { level: 1 })).toBeInViewport();

      // スクリーンショットは先頭に戻して撮る。
      await page.evaluate(() => {
        const main = document.querySelector('main') as HTMLElement;
        (main.parentElement as HTMLElement).scrollTop = 0;
      });
      await page.screenshot({ path: `${SHOT_DIR}/owner-report-390x844.png`, fullPage: true });
    } finally {
      await page.close();
    }
  });
});
