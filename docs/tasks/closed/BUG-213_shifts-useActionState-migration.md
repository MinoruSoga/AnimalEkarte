# BUG-191: ShiftFormDialog が useActionState + SubmitButton 未使用

| 項目 | 内容 |
|------|------|
| 優先度 | **Medium** |
| カテゴリ | React 19 Action パターン |

## 概要

ShiftFormDialog が `onSubmit` + `useTransition` の旧パターンを使用。
プロジェクト規約では `useActionState` + `<form action={formAction}>` + `SubmitButton` が必須。

## 現状コード

### `frontend/src/features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx:90,181`
```typescript
const [isPending, startSaveTransition] = useTransition();
// ...
<form onSubmit={handleSubmit} className="space-y-4">
```

## 修正方針

```typescript
const [state, formAction, isPending] = useActionState(async (_, formData) => {
  // ... save logic
}, null);

<form action={formAction}>
  {/* ... form fields ... */}
  <SubmitButton>保存</SubmitButton>
</form>
```

## 参照実装

`features/owners/hooks/use-owner-form.ts` — useActionState + SubmitButton の正しいパターン。
