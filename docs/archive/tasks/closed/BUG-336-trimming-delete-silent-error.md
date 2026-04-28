# BUG-336: トリミング削除失敗時にエラートーストが表示されない

**Status**: CLOSED  
**Priority**: Medium  
**Discovery**: コード静的解析 Section 9 トリミング管理 (2026-04-12)

## 概要

`/trimming/:id` の詳細フォームで削除確認ダイアログを承認した際、削除 API が失敗しても「エラーが発生しました」トーストが表示されない。`handleDelete` に `onError` コールバックがなく、`useDeleteTrimming` にも `onError` がないため、エラーはサイレントに無視される。

## 再現手順

（再現にはバックエンドがエラーを返す状況が必要: 接続切断・認可エラー等）

1. 管理者アカウントでログイン
2. `/trimming/:id` でトリミング記録を開く
3. 削除ボタン → 確認ダイアログ → 「削除」クリック（バックエンドエラーを起こした状態で）
4. **結果**: ダイアログが閉じ、エラートーストが表示されない。データは削除されていないが、ユーザーは削除されたと思い込む
5. **期待**: 「サーバーエラーが発生しました」エラートーストが表示される

## 現状コード

### `frontend/src/features/trimming/hooks/use-trimming-form.ts:251-261`
```ts
const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    startDeleteTransition(() => {
        deleteMutation.mutate(id, {
            onSuccess: () => {
                toast.success("トリミング情報を削除しました");
                onSuccess?.();
            },
            // onError がない → エラーは useDeleteTrimming の onError に依存するが
            // useDeleteTrimming にも onError がない → サイレントに無視
        });
    });
}, [isEdit, id, deleteMutation, startDeleteTransition]);
```

### `frontend/src/features/trimming/api/delete-trimming.ts:8-17`
```ts
export const useDeleteTrimming = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: deleteTrimming,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["trimmings"] });
        },
        // onError がない
    });
};
```

### 比較: 正しい実装（プロジェクト内参照実装）

```ts
// frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:281-291
const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    deleteVaccinationFn(id, {
        onSuccess: () => {
            toast.success("予防接種情報を削除しました");
            onSuccess?.();
        },
        onError: (error) => {
            handleApiError(error, "削除");  // ✅ onError あり
        },
    });
}, [isEdit, id, deleteVaccinationFn]);
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/trimming/hooks/use-trimming-form.ts:251-261` | `handleDelete` — onError なし | 要修正 |
| `frontend/src/features/trimming/api/delete-trimming.ts:8-17` | フック自体 — onError なし | 要修正 |

## 修正方針

### 1. `use-trimming-form.ts:251-261` — onError を追加

```ts
const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    startDeleteTransition(() => {
        deleteMutation.mutate(id, {
            onSuccess: () => {
                toast.success("トリミング情報を削除しました");
                onSuccess?.();
            },
            onError: (error) => {
                handleApiError(error, "削除");  // ← 追加
            },
        });
    });
}, [isEdit, id, deleteMutation, startDeleteTransition]);
```

### 2. `delete-trimming.ts:8-17` — onError を追加（フォールバック）

```ts
export const useDeleteTrimming = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: deleteTrimming,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["trimmings"] });
        },
        onError: (error) => {
            handleApiError(error, "削除");  // ← 追加
        },
    });
};
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/error-handling.md` — handleApiError 必須

> **すべての `catch` ブロックで `handleApiError` を呼び出す。**

useMutation の `onError` コールバックも同様にエラーハンドリングが必要。

### プロジェクト内参照実装

- `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:281-291` — `onError` + `handleApiError` の正しいパターン

## 優先度

**Medium** — 削除失敗時にユーザーが操作の失敗を認識できず、データは残っているが削除されたと思い込む可能性がある。

## 関連チケット

- BUG-335: 予防接種一覧の削除でも同一パターン

## 関連ファイル

- `frontend/src/features/trimming/hooks/use-trimming-form.ts:251-261` — 修正対象
- `frontend/src/features/trimming/api/delete-trimming.ts:8-17` — 修正対象
