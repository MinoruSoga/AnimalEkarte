# BUG-335: 予防接種一覧の削除操作でユーザーフィードバックがない

**Status**: CLOSED  
**Priority**: Medium  
**Discovery**: コード静的解析 Section 8 予防接種管理 (2026-04-12)

## 概要

`/vaccinations` の一覧画面で削除確認ダイアログを承認すると、削除は実行されるが「削除しました」トーストが表示されない。また削除が失敗した場合もエラートーストが表示されない（エラーはサイレントに無視される）。

## 再現手順

1. 管理者アカウントでログイン
2. `/vaccinations` 一覧で既存の予防接種記録の削除ボタンをクリック
3. 確認ダイアログで「削除」ボタンをクリック
4. **結果**: ダイアログが閉じ、一覧から該当行が消える。トーストは表示されない
5. **期待**: 「予防接種記録を削除しました」成功トーストが表示される

## 現状コード

### `frontend/src/features/vaccinations/routes/VaccinationList.tsx:144-151`
```ts
const handleDeleteConfirm = useCallback(() => {
    if (!pendingDeleteId) return;
    startDeleteTransition(() => {
        deleteVaccinationFn(pendingDeleteId, {
            onSuccess: () => setPendingDeleteId(null),  // ← toast.success がない
            // onError もない → エラーは無視される
        });
    });
}, [pendingDeleteId, deleteVaccinationFn]);
```

### `frontend/src/features/vaccinations/api/delete-vaccination.ts:8-17`
```ts
export const useDeleteVaccination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteVaccination,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vaccinations"] });
      // toast.success がない
    },
    // onError がない → エラーが無視される
  });
};
```

### 比較: 正しい実装（プロジェクト内参照実装）

```ts
// frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:281-291 — フォーム側は正しく実装済み
const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    deleteVaccinationFn(id, {
        onSuccess: () => {
            toast.success("予防接種情報を削除しました");  // ✅ toast あり
            onSuccess?.();
        },
        onError: (error) => {
            handleApiError(error, "削除");  // ✅ エラー処理あり
        },
    });
}, [isEdit, id, deleteVaccinationFn]);
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/vaccinations/routes/VaccinationList.tsx:144-151` | 削除確認後のコールバック | 要修正 |
| `frontend/src/features/vaccinations/api/delete-vaccination.ts:8-17` | フック自体の onError 欠如 | 要修正 |

## 修正方針

### 1. `VaccinationList.tsx:144-151` — toast + onError を追加

```ts
const handleDeleteConfirm = useCallback(() => {
    if (!pendingDeleteId) return;
    startDeleteTransition(() => {
        deleteVaccinationFn(pendingDeleteId, {
            onSuccess: () => {
                toast.success("予防接種記録を削除しました");  // ← 追加
                setPendingDeleteId(null);
            },
            onError: (error) => {
                handleApiError(error, "削除");  // ← 追加
            },
        });
    });
}, [pendingDeleteId, deleteVaccinationFn]);
```

### 2. `delete-vaccination.ts:8-17` — onError を追加（フォールバック）

```ts
export const useDeleteVaccination = () => {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: deleteVaccination,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["vaccinations"] });
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

useMutation の `onError` も同様に必ず `handleApiError` を呼び出す必要がある。

### プロジェクト内参照実装

- `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:281-291` — 同一フィーチャー内のフォーム側で正しく実装済み
- `frontend/src/features/estimates/api/delete-estimate.ts:13-22` — `onSuccess` + `toast.success` + `onError` + `handleApiError` を全て含む正しいパターン

## 優先度

**Medium** — 削除操作は成功しているが、ユーザーへのフィードバックが欠如し、エラー時には無言で失敗する。

## 関連チケット

なし

## 関連ファイル

- `frontend/src/features/vaccinations/routes/VaccinationList.tsx:144-151` — 修正対象
- `frontend/src/features/vaccinations/api/delete-vaccination.ts:8-17` — 修正対象
