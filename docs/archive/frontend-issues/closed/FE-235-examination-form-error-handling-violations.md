# FE-235: use-examination-form のエラーハンドリング違反（toast.error 直呼び + delete に try/catch なし）

## 概要

`frontend/src/features/examinations/hooks/use-examination-form.ts` で2つのエラーハンドリング問題がある。

## 問題1: catch ブロックで toast.error を直接呼び出し

### `use-examination-form.ts:128-130`

```ts
// Before: handleApiError を使わずに固定文字列の toast
} catch {
  toast.error("保存に失敗しました");  // ← handleApiError でない
}

// After
} catch (error) {
  handleApiError(error, "診察記録の保存");
}
```

`handleApiError` を使わないと、401/403/409 等のステータスコード別メッセージ分岐が機能せず、
バックエンドの具体的なエラーメッセージも表示されない。

## 問題2: handleDelete に try/catch がない

### `use-examination-form.ts:151-158`

```ts
// Before: try/catch なし
const handleDelete = () => {
  startDeleteTransition(async () => {
    await deleteExamination(id);  // 失敗してもサイレント
    navigate(paths.examinations.list.getHref());
  });
};

// After
const handleDelete = () => {
  startDeleteTransition(async () => {
    try {
      await deleteExamination(id);
      navigate(paths.examinations.list.getHref());
    } catch (error) {
      handleApiError(error, "診察記録の削除");
    }
  });
};
```

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックで `handleApiError(error, "コンテキスト")` を呼び出す。

### 関連チケット
- FE-218: toast.error 直呼びの同種問題（OwnersList, CompanySettings）

## 優先度
**Medium** — 診察記録の保存・削除エラーが適切にユーザーに通知されない。

## 関連ファイル
- `frontend/src/features/examinations/hooks/use-examination-form.ts`
- `frontend/src/lib/handle-api-error.ts`
