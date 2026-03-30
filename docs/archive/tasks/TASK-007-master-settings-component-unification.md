# TASK-007: マスタ設定画面の共通コンポーネント統一 + Vercel Best Practices 準拠

**作成日**: 2026-03-17
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

マスタ設定画面17ページ（5,461行）で重複している CRUD ハンドラ・レイアウト・検索パターンを共通 hook/コンポーネントに抽出し、Vercel React Best Practices に完全準拠させる。レガシーの `Settings.tsx`（旧 master_items API）も廃止する。

## 依頼内容（原文）

> フロントエンドの設定マスタのすべてのページにて、できる限り同じコンポーネントを使用するようにしてください。
> vercel-react-best-practicesのベストプラクティスなコード規約に準拠する実装にして。

## 現状分析

### 重複コードの規模
- CRUD ハンドラ: ~440行（11ページ同一パターン）
- ページレイアウト外枠: ~220行（11ページ同一）
- 検索フィルタ: ~91行（13ページ同一）
- state 宣言: ~44行（11ページ同一）
- **合計 ~795行の重複**

### レガシーコード
- `Settings.tsx`（416行）+ `useMasterItems` hook + 旧 master_items API（4ファイル）
- 3ページ（Insurance, JobTitle, InquiryTemplate）がこのレガシーテンプレートを使用
- `master_items` STI は DB レベルで既に廃止済み（45テーブル分割済み）

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 |
|---|----------|------|---------|------|
| 1 | `useMasterCRUD<T>` カスタム hook 作成 | FE | FE-026 | - |
| 2 | `MasterListPage` レイアウトコンポーネント作成 | FE | FE-027 | - |
| 3 | パターンA/B ページ（7ページ）を hook + レイアウトに移行 | FE | FE-028 | #1, #2 |
| 4 | パターンC ページ（Diagnosis, Trimming）を移行 | FE | FE-029 | #1, #2 |
| 5 | レガシー Settings.tsx 廃止 + Insurance/JobTitle 専用ページ化 | FE | FE-030 | #1, #2 |

## 影響範囲

### DB / Backend
- 変更なし（Insurance, JobTitle の専用 API は既に実装済みのはず）

### Frontend
- `frontend/src/features/master/hooks/use-master-crud.ts` — 新規 hook
- `frontend/src/features/master/components/MasterListPage.tsx` — 新規レイアウト
- `frontend/src/features/master/routes/*.tsx` — 全17ページの改修
- `frontend/src/features/master/hooks/use-master-items.ts` — 廃止
- `frontend/src/features/master/api/get-master-items.ts` 等 — 廃止

## 実装順序

1. FE-026 + FE-027（共通 hook + レイアウト — 並行可）
2. FE-028（パターンA/B 7ページ移行）
3. FE-029（パターンC 2ページ移行）
4. FE-030（レガシー廃止 + 専用ページ化）

## 関連イシュー

- [FE-026: useMasterCRUD hook](../frontend/issues/open/FE-026-use-master-crud-hook.md)
- [FE-027: MasterListPage コンポーネント](../frontend/issues/open/FE-027-master-list-page-component.md)
- [FE-028: パターンA/B ページ移行](../frontend/issues/open/FE-028-pattern-ab-pages-migration.md)
- [FE-029: パターンC ページ移行](../frontend/issues/open/FE-029-pattern-c-pages-migration.md)
- [FE-030: レガシー Settings.tsx 廃止](../frontend/issues/open/FE-030-legacy-settings-removal.md)
