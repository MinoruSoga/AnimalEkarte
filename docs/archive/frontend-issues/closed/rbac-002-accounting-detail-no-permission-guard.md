# BUG: AccountingDetail — 閲覧のみユーザーに「物販追加」「返金登録」「会計確定」ボタンが表示される

## 優先度
**HIGH** — セキュリティ / RBAC バイパス

## 再現手順
1. 権限グループを「閲覧のみ」（can_view=true, can_edit=false, can_delete=false）に設定
2. `/accounting/:id` に直接アクセス（URL バイパス）
3. 「会計精算」ページが開く
4. 以下のボタンが閲覧のみユーザーに表示される：
   - 「+ 物販・その他追加」ボタン
   - 「会計を確定する」SubmitButton
   - 「+ 返金を登録」ボタン

## 期待動作
- `canEdit=false` のとき「物販・その他追加」「会計を確定する」を非表示
- `canCreate=false` または `canEdit=false` のとき「+ 返金を登録」を非表示

## 実際の動作
- `AccountingDetail` に `usePermission` 呼び出しが存在しない
- 全ての操作ボタンが閲覧のみユーザーにも表示される
- ボタンを押すと API は 403 で弾くが、UI レイヤーの防御が欠如

## 影響ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx`
  - 「物販・その他追加」ボタン (line ~207)
  - 「会計を確定する」SubmitButton (line ~566)
  - 「返金を登録」ボタン (line ~640)

## 修正方針
```tsx
import { usePermission } from "@/features/auth";
// ...
const { canEdit, canCreate } = usePermission("accounting");
// ...
// 物販・その他追加ボタン
{canEdit ? (
  <Button ...>物販・その他追加</Button>
) : null}
// 会計を確定する
{canEdit ? (
  <SubmitButton ...>会計を確定する</SubmitButton>
) : null}
// 返金を登録
{canEdit ? (
  <Button ...>+ 返金を登録</Button>
) : null}
```

## Layer 2 ルートガード欠如
`/accounting/:id` のルートに `RequirePermission` ラッパーがない。
