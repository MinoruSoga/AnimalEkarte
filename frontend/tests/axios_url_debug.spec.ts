import { test } from '@playwright/test';

test('Axios URL resolution debug', async ({ page }) => {
  const allRequests: { method: string; url: string }[] = [];
  page.on('request', req => {
    allRequests.push({ method: req.method(), url: req.url() });
  });
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, msg.text());
    }
  });

  await page.goto('/login');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(3000);

  // axios の実際の URL 解決を確認
  const axiosUrlTest = await page.evaluate(async () => {
    // window から axios を取得できるか確認
    // (Vite の module system ではグローバルに expose されていない)
    const result = {
      baseURL: '',
      testUrl: '',
    };
    return result;
  });
  console.log('Axios URL test:', axiosUrlTest);

  // 全てのリクエストを出力
  const nonAssetReqs = allRequests.filter(r =>
    !r.url.includes('.ts') && !r.url.includes('.js') && !r.url.includes('.css') &&
    !r.url.includes('.woff') && !r.url.includes('fonts') && !r.url.includes('gstatic') &&
    !r.url.includes('googleapis') && !r.url.includes('vite') && !r.url.includes('react-refresh') &&
    !r.url.includes('node_modules') && !r.url.includes('vite/dist')
  );
  console.log('All meaningful requests count:', nonAssetReqs.length);
  console.log('Requests:', nonAssetReqs.map(r => `${r.method} ${r.url}`));

  // fetch を直接呼んで確認
  const fetchResult = await page.evaluate(async () => {
    try {
      const resp = await fetch('/api/v1/me');
      return { status: resp.status, url: resp.url };
    } catch (e) {
      return { error: String(e) };
    }
  });
  console.log('Direct fetch /api/v1/me:', fetchResult);
});
