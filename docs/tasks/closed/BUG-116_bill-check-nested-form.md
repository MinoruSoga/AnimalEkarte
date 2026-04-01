# BUG-116: 「チェック完了」ボタンが機能しない（nested form 違反）

## 概要

カルテ編集「会計(医師確認)」タブの「チェック完了」ボタンをクリックしても、
billing-review API が呼ばれず、ステータスが 未確認 のまま変化しない。

React コンソールに以下のエラーが出力される：

```
In HTML, <form> cannot be a descendant of <form>. This will cause a hydration error.
Uncaught Error: A React form was unexpectedly submitted.
```

## 症状

1. カルテ編集 → 「会計(医師確認)」タブをクリック
2. 「チェック完了」ボタンをクリック
3. 期待: `POST /api/v1/medical-records/1/billing-review/confirm` → 200 / ステータス「確認済み」に変化
4. 実際:
   - API コールが一切発生しない
   - ステータスが「未確認」のまま
   - コンソールに nested form エラー（msgid=512, 513）

## 根本原因

`MedicalRecordBillCheck.tsx` の L201 が `<form action={formAction}>` を使用しているが、
このコンポーネントは `MedicalRecordForm` 内部の `<form action={outerFormAction}>` の子として
レンダリングされる。

HTML 仕様上 `<form>` のネストは禁止されており、ブラウザは内側の `<form>` を無視する。
React 19 はこれを検出し `"A React form was unexpectedly submitted"` をスローする。
結果、`formAction`（billing-review confirm）は一度も実行されない。

```tsx
// MedicalRecordForm.tsx（外側）
<form action={outerFormAction}>
  ...
  <MedicalRecordBillCheck />  {/* ← 内側に別 <form> を持つ */}
</form>

// MedicalRecordBillCheck.tsx L201（内側：問題箇所）
<form action={formAction}>       // ❌ nested form 違反
  <SubmitButton>チェック完了</SubmitButton>
</form>
```

## 修正方針

`<form action={formAction}>` を削除し、`useTransition` + ボタンの `onClick` に置き換える。

```tsx
// ✅ 修正後
const [isPending, startTransition] = useTransition();

const handleConfirm = useCallback(() => {
  startTransition(async () => {
    try {
      await confirmMutation.mutateAsync({
        confirmed_by: Number(user?.id ?? 0),
        memo: "医師確認済み",
      });
      toast.success("会計確認を完了しました");
    } catch {
      // handleApiError(error, "会計確認") を呼ぶ
    }
  });
}, [confirmMutation, user?.id]);

// JSX
<Button
  type="button"
  onClick={handleConfirm}
  disabled={isPending || items.length === 0}
>
  チェック完了
</Button>
```

`useActionState` と内側 `<form>` は完全に削除する（L46-60 + L201-211）。

## 影響ファイル

- `frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx`
  - L46-60: `useActionState` 定義 → 削除
  - L201-211: `<form action={formAction}><SubmitButton>...</SubmitButton></form>` → `useTransition` + `Button` に置換

## 優先度

High（会計確認ワークフローが全滅）

## 関連

- BUG-115: 見積書「行を追加」も同じ nested form パターン（別タブ）
- BUG-101: chiefComplaint バリデーション（外側 form action の誤発火原因）
- テスト確認日: 2026-04-01（ローカル環境）
- コンソールエラー msgid=510-513（`<form> cannot be a descendant of <form>`）
