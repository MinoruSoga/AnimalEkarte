# FE-078: PetEditModal に飼主変更機能追加

**Status**: Open
**Priority**: Medium
**Affects**: owners feature（PetEditModal）
**Date Created**: 2026-03-19
**Related**: TASK-022, FE-077

## Summary

PetEditModal（ペット編集モーダル）に「飼主変更」ボタンを追加し、OwnerSearchModal（FE-077 で作成）を使って飼主を変更できるようにする。飼主変更時は確認ダイアログを表示し、確定すると即時 PATCH API で `owner_id` を更新する。

## 現状のコード

### PetEditModal（飼主変更機能なし）

```typescript
// frontend/src/features/owners/components/PetEditModal.tsx:53-59
interface PetEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  ownerName?: string;       // ← 読み取り専用で表示のみ
  petData?: PetFormData;
  onSave: (data: PetFormData) => void;
}

// L180-181: タイトルに ownerName を表示
<DialogTitle className={`text-sm font-bold ${C.text}`}>
  {isEdit ? `${ownerName}のペット情報編集` : `${ownerName}のペット新規登録`}
</DialogTitle>
```

### PetFormData（owner_id フィールドなし）

```typescript
// frontend/src/features/owners/types/index.ts:22-49
export interface PetFormData {
  id: string;
  isPending?: boolean;
  petNumber: string;
  petName: string;
  // ... owner_id フィールドなし
}
```

### ペット更新 API（owner_id 送信可能）

```typescript
// frontend/src/features/pets/api/update-pet.ts
// PATCH /v1/pets/{id} — UpdatePetRequest = Partial<PetWritable>
// PetWritable には owner_id が含まれる
```

### Backend（owner_id バリデーション済み）

```go
// backend/internal/service/pet_service.go:226-231
if input.OwnerID != nil {
	if _, err := s.ownerRepo.FindByID(ctx, clinicID, *input.OwnerID); err != nil {
		return nil, apperrors.WrapInvalidInput("owner not found in this clinic")
	}
}
```

## 必要な変更

### 1. PetEditModal に props 追加

```typescript
// frontend/src/features/owners/components/PetEditModal.tsx

interface PetEditModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  ownerName?: string;
  ownerId?: string;                                    // 追加: 現在の飼主ID
  petData?: PetFormData;
  onSave: (data: PetFormData) => void;
  onChangeOwner?: (newOwner: { id: string; name: string }) => void;  // 追加: 飼主変更コールバック
}
```

### 2. PetEditModal に飼主変更ボタン追加

```typescript
// DialogHeader 内、タイトルの横に「飼主変更」ボタンを追加
// 既存ペット編集時（isEdit === true）のみ表示

<DialogHeader>
  <div className="flex items-center justify-between">
    <div>
      <DialogTitle>...</DialogTitle>
      <DialogDescription>...</DialogDescription>
    </div>
    {isEdit && onChangeOwner ? (
      <Button
        variant="outline"
        size="sm"
        onClick={() => setIsOwnerSearchOpen(true)}
        className={`h-8 text-xs ${C.borderMedium}`}
      >
        飼主変更
      </Button>
    ) : null}
  </div>
</DialogHeader>

// OwnerSearchModal を lazy で追加
<Suspense fallback={null}>
  {isOwnerSearchOpen ? (
    <OwnerSearchModal
      open={isOwnerSearchOpen}
      onOpenChange={setIsOwnerSearchOpen}
      currentOwnerName={ownerName}
      onSelect={handleOwnerChange}
    />
  ) : null}
</Suspense>
```

### 3. 飼主変更ハンドラ

```typescript
// PetEditModal 内
const [isOwnerSearchOpen, setIsOwnerSearchOpen] = useState(false);

const handleOwnerChange = useCallback((newOwner: { id: string; name: string }) => {
  setIsOwnerSearchOpen(false);
  onChangeOwner?.(newOwner);
}, [onChangeOwner]);
```

### 4. OwnerForm/OwnerFormPage から onChangeOwner を注入

PetEditModal は `features/owners/` 内にあるため、ペット更新 API（`features/pets/api/update-pet.ts`）を直接 import できない（cross-feature import 禁止）。

`app/pages/OwnerFormPage.tsx` で `updatePet` を注入するか、PetEditModal の `onChangeOwner` コールバックを通じて親コンポーネントに委譲する。

```typescript
// 親コンポーネント（OwnerForm.tsx or OwnerFormPage.tsx）で:
const handlePetChangeOwner = useCallback(async (petId: string, newOwner: { id: string; name: string }) => {
  // PATCH /v1/pets/{petId} with { owner_id: Number(newOwner.id) }
  // 成功時: toast + ペットリスト再取得
}, []);
```

## UI 操作フロー

1. ユーザーが飼主フォームで既存ペットの「編集」をクリック → PetEditModal が開く
2. モーダルヘッダーの「飼主変更」ボタンをクリック
3. OwnerSearchModal（FE-077 で作成済み）が開く
4. 飼主を検索・選択
5. 確認ダイアログ: 「このペットは {旧飼主名} の管理下から外れます。よろしいですか？」
6. 確定 → PATCH API 実行 → toast「飼主を {新飼主名} に変更しました」
7. PetEditModal が閉じる or 飼主名が更新される

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）
- [ ] cross-feature import なし（onChangeOwner は props 経由で注入）
- [ ] `lazy()` + `Suspense` で OwnerSearchModal を遅延ロード

## 依存関係

- FE-077 が先に完了している必要がある（OwnerSearchModal が作成済みであること）
- Backend のペット更新 API は既に owner_id 変更をサポート済み（バリデーション付き）

## 完了条件

- [ ] 型エラーなし（`pnpm build` パス）
- [ ] ESLint エラーなし（`pnpm lint` パス）
- [ ] PetEditModal で「飼主変更」ボタン → 検索 → 選択 → 確認 → 変更が動作する
- [ ] 新規ペット登録時は「飼主変更」ボタンが表示されない
- [ ] 飼主変更後にペット一覧の表示が更新される
- [ ] 既存のペット編集・保存機能に影響なし
