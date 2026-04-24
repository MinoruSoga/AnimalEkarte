# BUG: TrimmingForm — 閲覧のみユーザーに「保存」「削除」ボタンが表示される

## 優先度
**HIGH** — セキュリティ / RBAC バイパス

## 再現手順
1. 権限グループを「閲覧のみ」（can_view=true, can_edit=false, can_delete=false）に設定
2. `/trimming/:id` に直接アクセス（URL バイパス）
3. 「トリミング編集」フォームが開く
4. 右上に「削除」ボタン（赤）と「保存」ボタン（青）が表示される

## 期待動作
- `canEdit=false` のとき「保存」ボタンを非表示
- `canDelete=false` のとき「削除」ボタンを非表示

## 実際の動作
- `TrimmingForm` に `usePermission` 呼び出しが存在しない
- 「保存」「削除」ボタンが閲覧のみユーザーにも表示される
- ボタンを押すと API は 403 で弾くが、UI レイヤーの防御が欠如

## 影響ファイル
- `frontend/src/features/trimming/routes/TrimmingForm.tsx` (lines 541-559)

## 修正方針
```tsx
import { usePermission } from "@/features/auth";
// ...
const { canEdit, canDelete } = usePermission("trimming");
// ...
headerAction={
  <div className="flex gap-2">
    {mode === "edit" && canDelete ? (
      <Button variant="ghost-danger" ...>削除</Button>
    ) : null}
    {canEdit ? (
      <SubmitButton>保存</SubmitButton>
    ) : null}
  </div>
}
```

## Layer 2 ルートガード欠如
`/trimming/:id` のルートに `RequirePermission` ラッパーがない。
`edit` アクションに対するルートガードも不足している。
