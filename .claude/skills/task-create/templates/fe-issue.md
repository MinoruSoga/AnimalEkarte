# FE-XXX: イシュータイトル

**Status**: Open
**Priority**: High / Medium / Low
**Affects**: 影響する機能・コンポーネント
**Date Created**: YYYY-MM-DD
**Related**: TASK-XXX, BE-XXX（関連イシュー）

## Summary

1-2行で問題・実装内容を説明。

## 現状のコード

**実際のコードを読んで** 現在の実装を記載（推測禁止）。

```typescript
// frontend/src/features/xxx/yyy.tsx:行番号
// 現在のコード（関連部分のみ抜粋）
```

## 必要な変更

### 1. 型定義（該当する場合）

```typescript
// frontend/src/features/xxx/api/types.ts
// models.ts からの導出型を追加・修正
```

### 2. API hooks（該当する場合）

```typescript
// frontend/src/features/xxx/api/get-xxx.ts or create-xxx.ts
// 追加・修正する API 関数・hook
```

### 3. コンポーネント変更

```typescript
// frontend/src/features/xxx/components/XxxComponent.tsx
// or frontend/src/features/xxx/routes/XxxPage.tsx
// 追加・修正する UI（Before → After の差分）
```

### 4. hooks 変更（該当する場合）

```typescript
// frontend/src/features/xxx/hooks/use-xxx-form.ts（hooks ファイルは kebab-case 命名）
// フォーム状態・バリデーション等の変更
```

## UI 操作フロー

1. ユーザーが「〜」画面を開く
2. 「〜」ボタンをクリック
3. 〜が表示される
4. ...

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] feature 外部からの import は `features/xxx/index.ts`（barrel）経由 / feature 内部での自己 barrel import なし（正本: frontend/src/features/CLAUDE.md）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] フォームは `useActionState`（isPending 内蔵）、非フォームの遷移は `useTransition`。`useState(false)` + `setIsPending` 禁止（正本: frontend/CLAUDE.md）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）

## 依存関係

- BE-XXX が先に完了している必要がある（API エンドポイントが必要）
- `make codegen` で `models.ts` が更新されている必要がある

## 完了条件

- [ ] 型エラーなし（`pnpm build` パス — 全体コマンドはユーザー手動実行）
- [ ] ESLint エラーなし（`pnpm lint` パス — 同上）
- [ ] UI が期待通りに動作
- [ ] 既存機能に影響なし
