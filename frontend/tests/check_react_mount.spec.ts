import { test } from '@playwright/test';

test('Check React mount and API calls', async ({ page }) => {
  const allRequests: string[] = [];
  page.on('request', req => {
    allRequests.push(req.method() + ' ' + req.url());
  });
  page.on('console', msg => {
    if (!msg.text().includes('[vite]') && !msg.text().includes('React DevTools')) {
      console.log(`BROWSER [${msg.type()}]:`, msg.text());
    }
  });
  page.on('pageerror', err => console.log('PAGEERROR:', err.message));

  await page.goto('/login');

  // Vite のコンパイル完了を待つ
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(3000);

  // root の中身を確認
  const rootContent = await page.evaluate(() => {
    const root = document.getElementById('root');
    return {
      hasChildren: (root?.children?.length ?? 0) > 0,
      firstChildTag: root?.children[0]?.tagName,
      innerHTML: root?.innerHTML?.substring(0, 300),
    };
  });
  console.log('Root content:', JSON.stringify(rootContent, null, 2));

  // 全リクエストのうちAPIっぽいもの
  const apiReqs = allRequests.filter(r =>
    !r.includes('.ts') && !r.includes('.js') && !r.includes('.css') &&
    !r.includes('.woff') && !r.includes('fonts') && !r.includes('gstatic') &&
    !r.includes('googleapis')
  );
  console.log('Non-asset requests:', apiReqs);
});
