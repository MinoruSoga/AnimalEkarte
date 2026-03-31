# FE-077: カルテ編集ヘッダーに飼主変更モーダル追加

**Status**: Open
**Priority**: High
**Affects**: medical-records feature、shared components
**Date Created**: 2026-03-19
**Related**: TASK-022, BE-047, FE-078

## Summary

カルテ編集ページ（MedicalRecordForm）のヘッダー PatientInfoCard の飼主名をクリックすると、飼主検索モーダルが開き、別の飼主を選択してカルテの owner_id を変更できるようにする。共有コンポーネント `OwnerSearchModal` を新規作成し、FE-078（PetEditModal）でも再利用する。

## 現状のコード

### PatientInfoCard（既に onOwnerClick prop が存在）

```typescript
// frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx:26
onOwnerClick?: () => void;

// L68-76: onOwnerClick が渡されればクリック可能な飼主名を表示
{onOwnerClick ? (
  <button type="button" onClick={onOwnerClick}
    className="text-base font-medium text-[#37352F] hover:underline decoration-dotted underline-offset-2 cursor-pointer">
    {ownerName}
  </button>
) : (
  <span className="text-base font-medium text-[#37352F]">{ownerName}</span>
)}
```

### MedicalRecordForm（現在 onOwnerClick 未使用）

```typescript
// frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:195-211
<PatientInfoCard
  ownerName={selectedPet.ownerName}
  petName={...}
  // ... onOwnerClick は渡していない
/>
```

### use-medical-record-form.ts（飼主変更ロジックなし）

```typescript
// frontend/src/features/medical-records/hooks/use-medical-record-form.ts:60-68
const resolvedPetId = isNewRecord ? (petId ?? "") : (existingRecord?.petId ?? "");
const { pet: selectedPet, isLoading: isPetLoading } = usePetInfo(resolvedPetId);
const resolvedOwnerId = selectedPet?.ownerId ?? "";
const { owner } = useOwnerInfo(resolvedOwnerId);
const ownerDiscountRate = owner?.discountRate ?? 0;
```

### 飼主一覧 API（既存）

```typescript
// frontend/src/features/owners/api/get-owners.ts:14-17
export const getOwners = async (): Promise<Owner[]> => {
  const { data } = await axios.get<OwnersResponse>("/v1/owners");
  return data.data.map(transformOwner);
};
```

### カルテ更新 API（owner_id 送信可能）

```typescript
// frontend/src/features/medical-records/api/types.ts:38-40
export type UpdateMedicalRecordRequest = Partial<
  Omit<ApiMedicalRecord, ...>
>;
// owner_id?: number が含まれる
```

## 必要な変更

### 1. 共有コンポーネント `OwnerSearchModal` 新規作成

```typescript
// frontend/src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx

interface OwnerSearchModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (owner: { id: string; name: string }) => void;
  currentOwnerName?: string;  // 現在の飼主名（確認ダイアログ用）
}

export function OwnerSearchModal({ open, onOpenChange, onSelect, currentOwnerName }: OwnerSearchModalProps) {
  // 1. 飼主名・飼主No・電話番号で検索
  // 2. 結果テーブル表示
  // 3. 行クリックで確認ダイアログ表示
  //    「飼主を {currentOwnerName} → {selectedOwner.name} に変更します。よろしいですか？」
  // 4. 確定で onSelect(selectedOwner) を呼ぶ
}
```

**UI仕様は Figma デザインなしのため、既存の `PetSelectionSearchForm` と `StaffSelectionModal` のパターンを踏襲する。**

検索フィールド:
- 飼主No（ownerId）
- 飼主名（ownerName）
- 電話番号（phone）

結果テーブルカラム:
- 飼主No / 飼主名 / 電話番号 / 住所

### 2. MedicalRecordForm に飼主変更モーダル追加

```typescript
// frontend/src/features/medical-records/routes/MedicalRecordForm.tsx

// import 追加
const OwnerSearchModal = lazy(() =>
  import("@/components/shared/OwnerSearchModal/OwnerSearchModal").then((m) => ({ default: m.OwnerSearchModal }))
);

// state 追加
const [isOwnerSearchOpen, setIsOwnerSearchOpen] = useState(false);

// PatientInfoCard に onOwnerClick を渡す（編集時のみ）
<PatientInfoCard
  ownerName={selectedPet.ownerName}
  // ...
  onOwnerClick={!isNewRecord ? () => setIsOwnerSearchOpen(true) : undefined}
/>

// モーダル追加
<Suspense fallback={null}>
  {!isNewRecord && recordId ? (
    <OwnerSearchModal
      open={isOwnerSearchOpen}
      onOpenChange={setIsOwnerSearchOpen}
      currentOwnerName={selectedPet.ownerName}
      onSelect={handleChangeOwner}
    />
  ) : null}
</Suspense>
```

### 3. use-medical-record-form.ts に飼主変更ハンドラ追加

```typescript
// frontend/src/features/medical-records/hooks/use-medical-record-form.ts

// 飼主変更ハンドラ
const handleChangeOwner = useCallback((newOwner: { id: string; name: string }) => {
  if (!recordId) return;
  startSaveTransition(async () => {
    try {
      await updateMutation.mutateAsync({
        id: recordId,
        req: { owner_id: Number(newOwner.id) },
      });
      toast.success(`飼主を ${newOwner.name} に変更しました`);
      // selectedPet の再取得 or owner 情報の再取得が必要
    } catch (error) {
      handleApiError(error, "飼主変更");
    }
  });
}, [recordId, updateMutation, startSaveTransition]);

// return に追加
return {
  // ... 既存
  handleChangeOwner,
};
```

### 4. OwnerSearchModal の index.ts

```typescript
// frontend/src/components/shared/OwnerSearchModal/index.ts
export { OwnerSearchModal } from "./OwnerSearchModal";
```

## UI 操作フロー

1. ユーザーがカルテ編集ページ（`/medical-records/:id`）を開く
2. ヘッダーの PatientInfoCard で飼主名（下線付きリンクスタイル）をクリック
3. OwnerSearchModal が開く
4. 飼主名 or 飼主No or 電話番号で検索
5. 結果テーブルから飼主を選択
6. 確認ダイアログ: 「飼主を {旧飼主名} → {新飼主名} に変更します。よろしいですか？」
7. 「変更する」ボタンで確定 → PATCH API 実行
8. 成功 toast → ヘッダーの飼主名が更新される

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）
- [ ] OwnerSearchModal は `components/shared/` に配置（cross-feature import 違反回避）
- [ ] `lazy()` + `Suspense` でモーダル遅延ロード
- [ ] `useCallback` でハンドラ安定化

## 依存関係

- BE-047 が先に完了している必要がある（owner_id バリデーション）
- `features/owners/api/get-owners.ts` の飼主一覧 API を OwnerSearchModal 内から利用（shared → features 方向のため、OwnerSearchModal 内で直接 axios 呼び出しにする or props で検索関数を注入する）

## 完了条件

- [ ] 型エラーなし（`npm run build` パス）
- [ ] ESLint エラーなし（`npm run lint` パス）
- [ ] カルテ編集ページで飼主名クリック → モーダル → 検索 → 選択 → 確認 → 変更が動作する
- [ ] 変更後にヘッダーの飼主名と割引率が更新される
- [ ] 新規カルテ作成時は飼主変更ボタンが表示されない
- [ ] 既存機能に影響なし
