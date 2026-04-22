# E2E テスト実行ガイド

Animal Ekarte の E2E テストは Playwright を使用しており、以下のテストスイートで構成されています。

## テストスイート一覧

| ファイル | テスト内容 | ステータス |
|---------|----------|-----------|
| `login.spec.ts` | ログインフォーム UI テスト | ✅ 実装完了 |
| `master.spec.ts` | マスタ設定共通動作テスト | ✅ 実装完了 |
| `appointment.spec.ts` | 診察フロー E2E テスト | ✅ 新規実装 |
| `medical-records.spec.ts` | 医療記録フロー E2E テスト | ✅ 新規実装 |
| `hospitalization.spec.ts` | 入院管理フロー E2E テスト | ✅ 新規実装 |
| `permission-control.spec.ts` | 権限制御フロー E2E テスト | ✅ 新規実装 |
| `staff-management.spec.ts` | スタッフ管理フロー E2E テスト | ✅ 新規実装 |

## テスト実行方法

### 前提条件
- Docker Compose が起動していること
- バックエンド・フロントエンドが実行中であること
- テスト用アカウント (admin@example.com / password) がデータベースに存在すること

### 全テスト実行

```bash
# テスト用のデモ環境をリセット
docker compose exec db psql -U postgres animal_ekarte < /docker-entrypoint-initdb.d/001_init.sql
make reset

# フロントエンドテスト実行
docker compose exec frontend npm run test:e2e
```

### 特定のテストスイートのみ実行

```bash
# ログインテストのみ
docker compose exec frontend npx playwright test tests/login.spec.ts

# 診察フロー テストのみ
docker compose exec frontend npx playwright test tests/appointment.spec.ts

# 医療記録テストのみ
docker compose exec frontend npx playwright test tests/medical-records.spec.ts

# 入院管理テストのみ
docker compose exec frontend npx playwright test tests/hospitalization.spec.ts

# 権限制御テストのみ
docker compose exec frontend npx playwright test tests/permission-control.spec.ts

# スタッフ管理テストのみ
docker compose exec frontend npx playwright test tests/staff-management.spec.ts
```

### ブラウザモード (対話的実行)

```bash
# UI モードで実行（テストを視覚的に確認）
docker compose exec frontend npx playwright test --ui
```

### テスト結果確認

```bash
# HTMLレポート表示
docker compose exec frontend npx playwright show-report
```

## 認証フロー

### globalSetup 認証戦略

- `globalSetup.ts` が全テスト前に一度実行される
- テスト用アカウント (admin@example.com / password) でバックエンドにログイン
- Cookie を `/tmp/playwright-auth-state.json` に保存
- 全テストが認証済み状態から開始する

### ログインテスト (login.spec.ts)

ログインフォーム自体をテストするため、意図的に `storageState` を上書きして未認証状態で実行：

```typescript
test.use({ storageState: { cookies: [], origins: [] } });
```

## React 19 formAction の互換性

Playwright の `button.click()` はブラウザのネイティブ submit を発火させ、React の formAction を経由しません。
そのため、ログインテストでは `form.requestSubmit()` ヘルパーを使用します：

```typescript
await submitReactForm(page, '#login-form');
```

詳細は `auth-helper.ts` を参照してください。

## トラブルシューティング

### エラー: "Login failed: 401"
- テスト用アカウント (admin@example.com) がデータベースに存在することを確認
- パスワードがデフォルト値 "password" であることを確認
- バックエンドが起動していることを確認

```bash
docker compose ps
```

### エラー: "access_token Cookie not found"
- Vite proxy が正しく Set-Cookie を転送しているか確認
- `vite.config.ts` の proxy 設定を確認
- バックエンド Cookie 設定を確認 (`HttpOnly` が正しく設定されているか)

### テストがタイムアウト
- ページが読み込まれるのに時間がかかる場合、タイムアウト時間を増やす
- ネットワーク遅延がある場合、`waitForLoadState('networkidle')` を使用

### スクリーンショット・トレース確認

テストが失敗した場合、自動的にスクリーンショット・トレースが保存されます：

```bash
# 失敗レポートを表示
docker compose exec frontend npx playwright show-report
```

## テストケース追加方法

### 新しいテストスイートを追加する場合

1. `tests/` ディレクトリに `feature-name.spec.ts` を作成
2. 認証ブロック設定：
   ```typescript
   test.beforeEach(async ({ page }) => {
     await page.goto('http://frontend:3000/');
     await expect(page).not.toHaveURL(/\/login/, { timeout: 10000 });
   });
   ```
3. テストケースを実装

### テストケース命名規則

- テストグループ名: 機能別（例: "診察フロー E2E テスト"）
- テストケース名: `{テストグループ短縮コード}-{番号}: {日本語説明}`
  - 例: `APT-01: ダッシュボード表示確認`

### セレクタ戦略

以下の優先順位でセレクタを選択：

1. `data-testid` 属性（推奨）
2. `aria` ロール属性 (`role="button"` など)
3. テキスト検索 (`:has-text()`)
4. 最後の手段：CSS セレクタ / XPath

```typescript
// 推奨例
const btn = page.locator('[data-testid="btn-save"]');

// 許容例
const btn = page.getByRole('button', { name: /保存/i });

// 避けるべき例
const btn = page.locator('div > button:nth-child(3)');
```

## CI/CD 統合

### GitHub Actions での実行

`.github/workflows/e2e.yml`:

```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:18-alpine
      # ... その他のサービス定義
    steps:
      - uses: actions/checkout@v3
      - run: docker compose up -d
      - run: docker compose exec frontend npm run test:e2e
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: frontend/playwright-report/
```

## パフォーマンス目標

| メトリクス | 目標 | 現状 |
|----------|------|------|
| 全テスト実行時間 | < 60秒 | TBD |
| テスト失敗率 | < 1% | TBD |
| カバレッジ（メインフロー） | 80% + | TBD |

## 参考資料

- [Playwright 公式ドキュメント](https://playwright.dev/docs/intro)
- [Playwright 認証ガイド](https://playwright.dev/docs/auth)
- [BUG-001 修正レポート](../docs/BUG-001-crypto-randomuuid-fix.md)
