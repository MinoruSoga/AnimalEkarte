# BUG: Playwright (Headless Chrome) 環境でのログイン不具合

## 事象
Docker コンテナ内の Playwright (Ubuntu 24.04) からフロントエンド (`http://frontend:3000/login`) にアクセスし、ログインを試行すると成功しない。

## 詳細
- **挙動**: ログインボタンをクリックしても、ネットワークログに `/api/v1/login` への POST リクエストが発生しない。
- **エラー**: ブラウザコンソールにはエラーは表示されていないが、URL が `/login` から遷移しない。
- **推測原因**:
  - React 19 の `useActionState` (formAction) が、Playwright の `click()` イベントによるネイティブのフォーム送信を正しくインターセプトできていない。
  - Vite dev サーバーの `allowedHosts` 制限（修正済みだが影響が残っている可能性）。
  - Docker ネットワーク内での CORS / Cookie (SameSite=None) の制約。

## 再現手順
1. `docker compose up -d playwright`
2. `docker compose exec playwright npx playwright test tests/master.spec.ts`

## 期待動作
1. ログインが成功し、ダッシュボードまたはホームページに遷移すること。
2. 以降のマスタ設定テストが続行できること。

## 暫定対応
- 統合テスト (Vitest/JSDOM) またはコード精査による機能検証。
- Playwright テスト内での API 直接ログインによる Cookie 注入（未完成）。
