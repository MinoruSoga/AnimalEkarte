# FE-103: LoadingFallback / ErrorFallback / EmptyStateFallback 共通化

**Status**: Closed
**Priority**: Medium
**Affects**: trimming, medical-records, estimates, checkups（リストページ）
**Date Created**: 2026-03-24
**Related**: TASK-024

## Summary

複数のリストページで isLoading / isError 時のフォールバック UI が独立実装されており、スピナーのスタイル・エラーメッセージ文言・構造が不統一。共有コンポーネントとして切り出す。

## 現状のコード

```typescript
// frontend/src/features/trimming/routes/TrimmingList.tsx:295-300（完全重複）
if (isLoading) return (
  <div className="flex justify-center items-center p-8">
    <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
  </div>
);
if (error) return <div className="p-4 text-red-600">データの取得に失敗しました</div>;
```

```typescript
// frontend/src/features/medical-records/routes/MedicalRecords.tsx:121-126（同じパターン）
if (isLoading) return (
  <div className="flex justify-center items-center p-8">
    <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
  </div>
);
if (isError) return <div className="p-4 text-red-600">データの取得に失敗しました</div>;
```

```typescript
// frontend/src/features/checkups/routes/CheckupsList.tsx:68-79（Table 内バリエーション）
{isLoading ? (
  <TableRow>
    <TableCell colSpan={7} className="h-24 text-center text-sm text-muted-foreground">
      読み込み中...
    </TableCell>
  </TableRow>
) : error ? (
  <TableRow>
    <TableCell colSpan={7} className="h-24 text-center text-sm text-destructive">
      データの取得に失敗しました
    </TableCell>
  </TableRow>
) : ...}
```

## 必要な変更

### 1. DataStates コンポーネント作成

```typescript
// frontend/src/components/shared/DataStates/DataStates.tsx（新規作成）

// ローディングスピナー
export function LoadingFallback({ className }: { className?: string }) {
  return (
    <div className={`flex justify-center items-center p-8 ${className ?? ""}`}>
      <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
    </div>
  );
}

// エラー表示
export function ErrorFallback({
  message = "データの取得に失敗しました",
  className,
}: {
  message?: string;
  className?: string;
}) {
  return (
    <div className={`p-4 text-red-600 ${className ?? ""}`}>
      {message}
    </div>
  );
}

// 空状態表示
export function EmptyStateFallback({
  icon,
  message,
  className,
}: {
  icon?: React.ReactNode;
  message: string;
  className?: string;
}) {
  return (
    <div className={`flex flex-col items-center gap-2 p-8 text-muted-foreground ${className ?? ""}`}>
      {icon ? <div className="opacity-40">{icon}</div> : null}
      <span className="text-sm">{message}</span>
    </div>
  );
}
```

```typescript
// frontend/src/components/shared/DataStates/index.ts（新規作成）
export { LoadingFallback, ErrorFallback, EmptyStateFallback } from "./DataStates";
```

### 2. TrimmingList.tsx の置き換え

```typescript
// Before: インライン実装
// After:
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates/DataStates";

if (isLoading) return <LoadingFallback />;
if (error) return <ErrorFallback />;
```

### 3. MedicalRecords.tsx も同様に置き換え

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] barrel index 経由 import なし（`DataStates/DataStates` を直接 import）
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `FC` / `forwardRef` なし

## 依存関係

- Backend 変更なし。他の FE イシューとも独立。

## 完了条件

- [ ] `frontend/src/components/shared/DataStates/DataStates.tsx` が作成されている
- [ ] TrimmingList / MedicalRecords のインラインローディング・エラー表示が削除されている
- [ ] 各画面のローディング・エラー表示が変化なし
- [ ] `pnpm lint` エラーなし
- [ ] `pnpm build` エラーなし

## クローズ情報

- **Closed At**: 2026-03-24
- **変更ファイル**:
  - `frontend/src/components/shared/DataStates/DataStates.tsx` — 新規作成（LoadingFallback / ErrorFallback / EmptyStateFallback）
  - `frontend/src/features/trimming/routes/TrimmingList.tsx` — インライン実装を LoadingFallback / ErrorFallback に置き換え
  - `frontend/src/features/medical-records/routes/MedicalRecords.tsx` — 同上
