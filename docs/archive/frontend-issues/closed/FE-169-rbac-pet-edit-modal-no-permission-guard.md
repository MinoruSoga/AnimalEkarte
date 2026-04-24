# FE-169: PetEditModal — usePermission 完全欠落・保存ボタン常時表示・全フィールド常時編集可

## 概要

`OwnerForm.tsx` 内の `PetEditModal`（ペット追加・編集モーダル）は `usePermission` を一切呼び出していない。`canEdit=false` ユーザーが飼主フォームを開きペット行をクリックすると、保存ボタンが表示され、全フィールドが編集可能な状態でモーダルが開く。

## 根本原因

```tsx
// PetEditModal.tsx — usePermission なし ❌
export const PetEditModal = memo(function PetEditModal({
  isOpen,
  petId,
  ownerId,
  clinicId,
  onClose,
  onChangeOwner,
  // ...
}: PetEditModalProps) {
  // usePermission 呼び出しなし

  // 行 604-609 — 保存ボタンに権限ガードなし ❌
  <SubmitButton>
    {petId ? "更新" : "登録"}
  </SubmitButton>

  // 全フォームフィールドに disabled={!canEdit} なし ❌
  // 名前・種別・品種・性別・毛色・生年月日・体重等 常時編集可能
})
```

`OwnerForm.tsx` 側でペット行クリックを制御しているものの（`canEdit` で PetEditModal 開閉を条件付けしているかは未確認）、モーダル自体に権限ガードがない。もし呼び出し側のガードが不完全な場合、直接モーダルが開かれると無条件に編集・保存ができる。

## 影響

`canEdit=false` / `owners`（ペット設定に関連するリソース）権限のないユーザーがペット編集モーダルを開くと：
1. 名前・種別・品種・性別・毛色・生年月日・体重等の全フィールドが入力可能
2. 「登録」「更新」ボタンが表示される → クリックすると POST/PATCH → 403

## 修正方針

```tsx
// PetEditModal.tsx
import { usePermission } from "@/features/auth/hooks/use-permission";

export const PetEditModal = memo(function PetEditModal({...}: PetEditModalProps) {
  // 1. usePermission を追加（ペットは owners リソースで制御）
  const { canEdit } = usePermission("owners");

  return (
    <>
      {/* 2. フォームフィールドを fieldset でラップして一括 disable */}
      <fieldset disabled={!canEdit}>
        {/* ...全フォームフィールド... */}
      </fieldset>

      {/* 3. 保存ボタンを canEdit でガード */}
      {canEdit ? (
        <SubmitButton>{petId ? "更新" : "登録"}</SubmitButton>
      ) : null}
    </>
  );
});
```

## 優先度

**HIGH** — ペット情報（種別・品種・生年月日等）は診療・会計に直結するデータ。`canEdit=false` ユーザーが変更を試みると API エラーが発生する。また、呼び出し元でのガードが不完全な場合（OwnersList 行クリックから遷移）に直撃する（FE-159 参照）。

## 関連ファイル

- `frontend/src/features/owners/components/PetEditModal.tsx` (行 604-609: 保存ボタン)
- `frontend/src/features/owners/routes/OwnerForm.tsx` (PetEditModal 呼び出し元)
- 発見日: 2026-04-08（RBAC Phase 2/3 テスト中）
- 関連: FE-158（OwnerForm フィールド disabled 漏れ）、FE-159（OwnersList 行クリック）
