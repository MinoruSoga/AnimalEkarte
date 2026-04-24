# FE-149: OwnerForm — ペット行ドットメニューの新規作成ショートカットに canCreate ガードなし

## 概要

`OwnerForm.tsx` のペット行 DropdownMenu に「予約作成」「カルテ作成」「トリミング」「入院登録」「会計登録」のナビゲーションショートカットが存在するが、`canCreate` チェックがない。閲覧のみユーザーでも DropdownMenuTrigger（MoreHorizontal ボタン）が表示され、これらのメニュー項目が見えてしまう。

## 影響範囲

- ファイル: `frontend/src/features/owners/routes/OwnerForm.tsx`
- 行: 136〜163 あたりのペット行 DropdownMenu
- 権限: `can_create = false` のユーザー

## 現状の挙動

閲覧のみユーザーがオーナー編集画面でペット行の「…」をクリックすると、以下のメニュー項目が表示される：

- 予約作成 → `/reservations/new?petId=...`
- カルテ作成 → `/medical-records/new?petId=...`
- トリミング → `/trimming/new?petId=...`
- 入院登録 → `/hospitalization/new?petId=...`
- 会計登録 → `/accounting/new?petId=...`

遷移先のルートは `RequirePermission action="create"` でガードされているため実際の作成は不可能だが、メニュー項目が見えること自体がUX上の問題。

## 期待する挙動

`canCreate` が false の場合：
1. DropdownMenuTrigger（「…」ボタン）自体を非表示にするか、
2. 新規作成系のメニュー項目のみ非表示にする

注: 「詳細・編集」（canEdit）と「削除」（canDelete）はすでにガードされている。

## 修正方針

```tsx
// OwnerForm.tsx — usePermission でリソース横断 canCreate を取得
const { canCreate: canCreateReservation } = usePermission("reservations");
// ... 他のリソースも同様

// または owners の canCreate を転用 (owners 作成 = ペット行追加も同権限と見なすなら)
const { canCreate } = usePermission("owners");

// DropdownMenu 内:
{canCreate ? (
  <>
    <DropdownMenuItem onClick={() => navigate(`/medical-records/new?petId=${pet.id}`)}>カルテ作成</DropdownMenuItem>
    ...
  </>
) : null}
```

## 優先度

LOW — 遷移先でルートガードが機能するため実害はないが、UX が不整合

## 関連

- `frontend/src/features/owners/routes/OwnerForm.tsx`
- BUG-RBAC テスト 2026-04-07
