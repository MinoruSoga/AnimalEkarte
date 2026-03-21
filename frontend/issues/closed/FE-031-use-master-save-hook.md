# FE-031: useMasterSave hook — handleSave パターン抽出

**Status**: Open
**Priority**: High
**Affects**: 全マスタ設定ページ（9ページ）
**Date Created**: 2026-03-18
**Related**: TASK-008, FE-032

## Summary

全マスタ設定ページで繰り返されている `handleSave` コールバック（~30行/ページ）を汎用 hook に抽出する。バリデーション → `startSaveTransition` → create/update 分岐 → toast 通知のパターンを型安全に共通化する。

## 現状のコード

**9ページすべてで以下のパターンが重複している:**

```typescript
// frontend/src/features/master/routes/JobTitleSettings.tsx:121-162
const handleSave = useCallback(
  (data: JobTitleFormData) => {
    if (!data.name.trim()) {
      toast.error("名称は必須です");
      return;
    }
    crud.startSaveTransition(() => {
      if (crud.editTarget !== null && crud.editTarget !== "new") {
        const req: UpdateJobTitleRequest = {
          name: data.name,
          description: data.description || undefined,
          is_active: data.isActive,
        };
        updateMutation.mutate(
          { id: crud.editTarget.id, req },
          {
            onSuccess: () => { toast.success("更新しました"); crud.handleClose(); },
            onError: () => toast.error("更新に失敗しました"),
          },
        );
      } else {
        const req: CreateJobTitleRequest = { ... };
        createMutation.mutate(req, {
          onSuccess: () => { toast.success("登録しました"); crud.handleClose(); },
          onError: () => toast.error("登録に失敗しました"),
        });
      }
    });
  },
  [crud.editTarget, updateMutation, createMutation, crud.handleClose, crud.startSaveTransition],
);
```

**ページ間の差分（これだけが異なる）:**
1. バリデーション関数（`data.name.trim()` が基本、Staff は email/password 追加チェック）
2. `toCreateRequest(data)` — FormData → CreateRequest 変換
3. `toUpdateRequest(data)` — FormData → UpdateRequest 変換

## 必要な変更

### 1. hook 作成

```typescript
// frontend/src/features/master/hooks/use-master-save.ts

import { useCallback } from "react";
import { toast } from "sonner";
import type { UseMutationResult } from "@tanstack/react-query";
import type { UseMasterCRUDReturn } from "@/features/master/hooks/use-master-crud";

interface MasterEntity {
  id: string;
}

interface UseMasterSaveOptions<T extends MasterEntity, TForm, TCreate, TUpdate> {
  crud: UseMasterCRUDReturn<T>;
  createMutation: UseMutationResult<unknown, Error, TCreate>;
  updateMutation: UseMutationResult<unknown, Error, { id: string; req: TUpdate }>;
  /** バリデーション。エラーメッセージを返すと toast.error で表示。null なら通過。 */
  validate: (data: TForm) => string | null;
  /** FormData → CreateRequest 変換 */
  toCreateRequest: (data: TForm) => TCreate;
  /** FormData → UpdateRequest 変換 */
  toUpdateRequest: (data: TForm) => TUpdate;
}

interface UseMasterSaveReturn<TForm> {
  handleSave: (data: TForm) => void;
}

export function useMasterSave<T extends MasterEntity, TForm, TCreate, TUpdate>({
  crud,
  createMutation,
  updateMutation,
  validate,
  toCreateRequest,
  toUpdateRequest,
}: UseMasterSaveOptions<T, TForm, TCreate, TUpdate>): UseMasterSaveReturn<TForm> {
  const handleSave = useCallback(
    (data: TForm) => {
      const error = validate(data);
      if (error) {
        toast.error(error);
        return;
      }

      crud.startSaveTransition(() => {
        if (crud.editTarget !== null && crud.editTarget !== "new") {
          const req = toUpdateRequest(data);
          updateMutation.mutate(
            { id: crud.editTarget.id, req },
            {
              onSuccess: () => {
                toast.success("更新しました");
                crud.handleClose();
              },
              onError: () => toast.error("更新に失敗しました"),
            },
          );
        } else {
          const req = toCreateRequest(data);
          createMutation.mutate(req, {
            onSuccess: () => {
              toast.success("登録しました");
              crud.handleClose();
            },
            onError: () => toast.error("登録に失敗しました"),
          });
        }
      });
    },
    [crud.editTarget, crud.handleClose, crud.startSaveTransition, createMutation, updateMutation, validate, toCreateRequest, toUpdateRequest],
  );

  return { handleSave };
}
```

### 2. UseMasterCRUDReturn 型のエクスポート

`use-master-crud.ts` から `UseMasterCRUDReturn<T>` 型を export する（現状未エクスポートの場合）。

## 使用例（移行後の JobTitleSettings）

```typescript
const { handleSave } = useMasterSave<JobTitle, JobTitleFormData, CreateJobTitleRequest, UpdateJobTitleRequest>({
  crud,
  createMutation,
  updateMutation,
  validate: (data) => (!data.name.trim() ? "名称は必須です" : null),
  toCreateRequest: (data) => ({
    name: data.name,
    description: data.description || undefined,
    is_active: data.isActive,
  }),
  toUpdateRequest: (data) => ({
    name: data.name,
    description: data.description || undefined,
    is_active: data.isActive,
  }),
});
```

## プロジェクトルール遵守チェック

- [x] `any` 型なし — ジェネリクスで型安全
- [x] `FC` / `forwardRef` なし — hook なので該当なし
- [x] barrel index 経由 import なし
- [x] `useTransition` で pending 管理（`crud.startSaveTransition` 経由）
- [x] `useCallback` でハンドラ安定化
- [x] Vercel Best Practices: `rerender-functional-setstate`, `rerender-dependencies` 準拠

## 完了条件

- [ ] `use-master-save.ts` 作成
- [ ] `UseMasterCRUDReturn<T>` 型が export されている
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend npm run lint` パス）
- [ ] 既存の9ページから呼び出し可能（FE-033/034 で移行）
