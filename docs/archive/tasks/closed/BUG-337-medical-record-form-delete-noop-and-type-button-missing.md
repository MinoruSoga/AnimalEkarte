# BUG-337: カルテフォームの削除確認ダイアログが実際に削除しない + フローティングボタンが form submit を発火する

**Status**: OPEN  
**Priority**: Critical  
**Discovery**: コード静的解析 Section 4 カルテ管理 (2026-04-12)

## 概要

`/medical-records/:id` のカルテフォームで「削除」ボタンをクリックすると確認ダイアログが表示されるが、「削除する」ボタンをクリックしても**カルテは削除されない**（ダイアログが閉じるだけ）。加えて、フォーム内の「削除」「バイタル記録」「印刷」ボタンに `type="button"` がないため、クリックするたびに `<form action={formAction}>` が submit されカルテが保存される。

## 再現手順

### Bug 1: 削除確認ダイアログが削除を実行しない

1. 管理者アカウントでログイン
2. `/medical-records/:id` でカルテを開く（「問診」タブ）
3. 右下の赤い「削除」ボタンをクリック → 確認ダイアログが表示される
4. 「削除する」をクリック
5. **結果**: ダイアログが閉じ、カルテは削除されない（`/medical-records` 一覧に遷移もしない）
6. **期待**: DELETE API が呼ばれ、「カルテを削除しました」トースト後に一覧に遷移する

### Bug 2: 「削除」ボタンクリックでフォームが保存される

1. カルテを編集状態にする（何か変更する）
2. 「削除」ボタンをクリック（`type="button"` がないため form submit 発生）
3. **結果**: 削除ダイアログが表示されると同時に `formAction` が実行されカルテが保存される
4. **期待**: ダイアログのみ表示され、フォームは submit されない

### Bug 3: 「バイタル記録」「印刷」ボタンでも同様

クリックすると `type="button"` がないため form submit が発火し、カルテが保存される。

## 現状コード

### `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:264`
```tsx
<form action={formAction} className={LAYOUT.fullHeight}>
```

### `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:415-491`
```tsx
{/* フローティングボタン — すべて type="button" なし → form submit 発火 */}
<Button
  variant="ghost-danger"
  onClick={() => setIsDeleteConfirmOpen(true)}   // ← type="button" なし → form submit
  className={...}
>
  削除
</Button>
<Button
  variant="outline"
  onClick={() => setIsVitalsOpen(true)}           // ← type="button" なし → form submit
>
  バイタル記録
</Button>
<Button
  variant="outline"
  onClick={() => { window.print(); }}             // ← type="button" なし → form submit
>
  印刷
</Button>

{/* Delete confirm dialog — 実際の削除ロジックがない */}
{isDeleteConfirmOpen ? (
  <div ...>
    <Button variant="outline" onClick={() => setIsDeleteConfirmOpen(false)}>
      キャンセル     {/* ← type="button" なし + ダイアログ閉じるだけ */}
    </Button>
    <Button onClick={() => setIsDeleteConfirmOpen(false)}>
      削除する       {/* ← type="button" なし + ダイアログ閉じるだけ、削除API呼ばない */}
    </Button>
  </div>
) : null}
```

### `frontend/src/features/medical-records/api/delete-medical-record.ts:9-19`
```ts
// ← 実装は存在するが MedicalRecordForm.tsx で使用されていない
export const useDeleteMedicalRecord = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteMedicalRecord,
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["medical-records"] }); },
    onError: (error) => handleApiError(error, "カルテ削除"),
  });
};
```

### 比較: 正しい実装（プロジェクト内参照実装）

```ts
// frontend/src/features/owners/routes/OwnersList.tsx:306-318 — 正しい削除パターン
const handleConfirmDelete = useCallback(() => {
    if (!pendingDeleteOwnerId) return;
    startDeleteTransition(async () => {
        try {
            await deleteOwner(pendingDeleteOwnerId);
            toast.success("飼主を削除しました");
            closeDeleteModal();
            revalidator.revalidate();
        } catch (error) {
            handleApiError(error, "削除");
        }
    });
}, [pendingDeleteOwnerId, closeDeleteModal, revalidator]);
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `MedicalRecordForm.tsx:415-422` | 「削除」フローティングボタン — `type="button"` なし | 要修正 |
| `MedicalRecordForm.tsx:425-434` | 「バイタル記録」ボタン — `type="button"` なし | 要修正 |
| `MedicalRecordForm.tsx:437-451` | 「印刷」ボタン — `type="button"` なし | 要修正 |
| `MedicalRecordForm.tsx:479-487` | 削除確認ダイアログ内ボタン — `type="button"` なし + 削除ロジックなし | 要修正 |
| `MedicalRecordForm.tsx` | `useDeleteMedicalRecord` が未使用 | 要修正（接続が必要） |

## 修正方針

### 1. フローティングボタンに `type="button"` を追加

```tsx
{/* 削除 */}
<Button
  type="button"    // ← 追加
  variant="ghost-danger"
  onClick={() => setIsDeleteConfirmOpen(true)}
  className={...}
>
  削除
</Button>

{/* バイタル記録 */}
<Button
  type="button"    // ← 追加
  variant="outline"
  onClick={() => setIsVitalsOpen(true)}
  disabled={isNewRecord}
>
  バイタル記録
</Button>

{/* 印刷 */}
<Button
  type="button"    // ← 追加
  variant="outline"
  onClick={() => { setIsPrinting(true); setTimeout(() => { window.print(); setIsPrinting(false); }, 100); }}
>
  印刷
</Button>
```

### 2. 削除確認ダイアログに実際の削除ロジックを追加

```tsx
// コンポーネントトップに追加
const navigate = useNavigate();
const { mutate: deleteRecord, isPending: isDeleting } = useDeleteMedicalRecord();

// 削除ハンドラ
const handleDeleteConfirm = useCallback(() => {
    if (!recordId) return;
    deleteRecord(recordId, {
        onSuccess: () => {
            toast.success("カルテを削除しました");
            navigate(paths.medicalRecords.getHref());
        },
        onError: (error) => {
            handleApiError(error, "カルテ削除");
        },
    });
}, [recordId, deleteRecord, navigate]);

// ダイアログのボタン
<Button
  type="button"    // ← 追加
  variant="outline"
  onClick={() => setIsDeleteConfirmOpen(false)}
  disabled={isDeleting}
>
  キャンセル
</Button>
<Button
  type="button"    // ← 追加
  className={`${STYLE.btnDanger}`}
  onClick={handleDeleteConfirm}  // ← 実際の削除処理に変更
  disabled={isDeleting}
>
  {isDeleting ? "削除中..." : "削除する"}
</Button>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — React 19 Action パターン

> **React 19 Action**: 原則 `useActionState` と `<form action={formAction}>` を使用

`<form action={...}>` を使う場合、フォーム内の全 `<button>` に `type="button"` を明示しないと全ボタンがデフォルトで form を submit してしまう。

### `.claude/rules/error-handling.md` — handleApiError 必須

> **すべての `catch` ブロックで `handleApiError` を呼び出す。**

### プロジェクト内参照実装

- `frontend/src/features/owners/routes/OwnerForm.tsx` — `type="button"` 明示例
- `frontend/src/features/owners/routes/OwnersList.tsx:306-318` — 削除の正しいパターン

## 優先度

**Critical** — カルテ削除ボタンが存在するにもかかわらず、クリックしてもデータは削除されない（UIとロジックが完全に切断されている）。また、意図しないカルテ保存が複数のボタンで発生する。

## 関連チケット

- BUG-333: 会計詳細の同一パターン（`type="button"` 欠如）

## 関連ファイル

- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:415-491` — 修正対象（全ボタン）
- `frontend/src/features/medical-records/api/delete-medical-record.ts` — 実装済み（接続が必要）
