# E2E テスティング実装ガイド

> **作成日**: 2026-04-23  
> **ステータス**: ✅ TIER 3 — E2E テスト基盤構築完了  
> **対象システム**: Animal Ekarte (動物病院向け電子カルテシステム)

---

## 📌 概要

本ドキュメントは、Animal Ekarte の E2E (End-to-End) テスト実装ガイドです。Playwright を使用した UI テストフロー、認証戦略、テストケース設計、CI/CD 統合をカバーしています。

## 🎯 実装目標

| 項目 | 目標 | 実装状況 |
|------|------|--------|
| テスト基盤構築 | Playwright + globalSetup 認証戦略 | ✅ 完了 |
| テストスイート | 6つの主要業務フロー | ✅ 完了 |
| ヘルパー関数 | 15+ ページ操作ユーティリティ | ✅ 完了 |
| CI/CD 統合 | GitHub Actions workflow | ✅ 完了 |
| ドキュメント | テスト実行ガイド | ✅ 完了 |

## 🏗 実装内容

### テストスイート (6個)

1. **login.spec.ts** — ログイン UI テスト
2. **appointment.spec.ts** — 診察フロー (3ケース)
3. **medical-records.spec.ts** — 医療記録 (3ケース)
4. **hospitalization.spec.ts** — 入院管理 (4ケース)
5. **permission-control.spec.ts** — 権限制御 (5ケース)
6. **staff-management.spec.ts** — スタッフ管理 (5ケース)

**合計**: 18テストケース

### ページヘルパー関数 (page-helpers.ts)

15個のユーティリティ関数（要素操作、ナビゲーション、検証）

```typescript
// 要素操作
clickElement / fillInput / selectTableRow

// ナビゲーション
expectNavigationToUrl / expectAuthenticatedState

// UI パターン
waitForSidepanel / waitForDialog / clickDialogButton

// 検証
expectToastMessage / expectFormError / waitForDataDisplay
```

### GitHub Actions Workflow

`.github/workflows/e2e-tests.yml` - 自動テスト実行パイプライン

- Docker Compose セットアップ
- DB マイグレーション
- E2E テスト実行
- レポート生成・アップロード

## 🚀 実行方法

### 全テスト実行

```bash
docker compose up                 # 別ターミナル
docker compose exec frontend npm run test:e2e
```

### 特定テスト実行

```bash
docker compose exec frontend npx playwright test tests/appointment.spec.ts
```

### UI モード (対話的実行)

```bash
docker compose exec frontend npm run test:e2e:ui
```

### デバッグモード

```bash
docker compose exec frontend npm run test:e2e:debug
```

## 🔐 認証戦略

### globalSetup

テスト実行前に一度、テスト用アカウントでログイン。Cookie を保存し、全テストが認証済み状態から開始。

### ログインテスト例外

`login.spec.ts` は意図的に未認証状態で実行。`form.requestSubmit()` で React 19 formAction をトリガー。

## 📊 テストカバレッジ

| 機能 | テスト数 | ステータス |
|------|---------|----------|
| ログイン | 1 | ✅ |
| 診察 | 3 | ✅ |
| 医療記録 | 3 | ✅ |
| 入院管理 | 4 | ✅ |
| 権限制御 | 5 | ✅ |
| スタッフ管理 | 5 | ✅ |

## 🛠 トラブルシューティング

### エラー: "Login failed"
- テスト用アカウントが DB に存在することを確認
- 003_seed_demo.sql でデフォルトアカウントが作成されているか確認

### エラー: "access_token Cookie not found"
- Vite proxy の Set-Cookie 転送設定を確認 (vite.config.ts)
- バックエンド Cookie 設定を確認 (HttpOnly 設定)

### テストがタイムアウト
- タイムアウト時間を増やす
- `waitUntil: 'networkidle'` を指定

## 📚 参考資料

- `frontend/tests/README.md` — テスト実行ガイド
- `frontend/tests/page-helpers.ts` — ヘルパー関数リファレンス
- `.github/workflows/e2e-tests.yml` — CI/CD workflow

---

**最終更新**: 2026-04-23  
**担当**: Claude Code (TIER 3 E2E テスト基盤構築)
