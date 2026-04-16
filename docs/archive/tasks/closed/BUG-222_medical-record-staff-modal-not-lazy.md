# BUG-222: MedicalRecordForm の StaffSelectionModal が lazy() 未使用

## 概要

`MedicalRecordForm.tsx:27` で `StaffSelectionModal` を直接 import している。同ファイルの `VitalsModal`・`OwnerSearchModal` は既に `lazy()` + `Suspense` でロードされており、モーダル系コンポーネントの扱いに不統一がある。`bundle-dynamic-imports` ルールへの準拠が必要。

## 再現手順

（ランタイム動作は変わらないが、バンドル分析で確認可能）
1. `npm run build` を実行しバンドルサイズを確認
2. MedicalRecordForm の初期バンドルに StaffSelectionModal が含まれる
3. **結果**: モーダルが使用されない場合でもバンドルに含まれる

## 期待する動作

- StaffSelectionModal は VitalsModal・OwnerSearchModal と同様に `lazy()` でロードされること

## 現状コード

### `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:27`
```tsx
// ❌ 直接 import（VitalsModal・OwnerSearchModal と不統一）
import { StaffSelectionModal } from "../components/StaffSelectionModal";

// ✅ 既に lazy になっているもの
const VitalsModal = lazy(() =>
  import("../components/VitalsModal").then((m) => ({ default: m.VitalsModal }))
);
const OwnerSearchModal = lazy(() =>
  import("@/components/shared/OwnerSearchModal/OwnerSearchModal").then((m) => ({ default: m.OwnerSearchModal }))
);
```

### 比較: 正しい実装（同ファイル内）
```tsx
const VitalsModal = lazy(() =>
  import("../components/VitalsModal").then((m) => ({ default: m.VitalsModal }))
);
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `features/medical-records/routes/MedicalRecordForm.tsx:27` | StaffSelectionModal 直接 import | 未修正 |

## 修正方針

### `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`

```tsx
// Before（line 27）
import { StaffSelectionModal } from "../components/StaffSelectionModal";

// After
const StaffSelectionModal = lazy(() =>
  import("../components/StaffSelectionModal").then((m) => ({ default: m.StaffSelectionModal }))
);
```

また、`StaffSelectionModal` の使用箇所を `<Suspense fallback={null}>` で囲むこと（`VitalsModal` と同じ pattern）:
```tsx
<Suspense fallback={null}>
  <StaffSelectionModal
    open={isStaffModalOpen}
    // ...
  />
</Suspense>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — `bundle-dynamic-imports`
> 重いモーダル・ダイアログは `lazy()` + `Suspense` で遅延ロード

### `.claude/rules/code-style.md` — Performance Rules
> `bundle-dynamic-imports`: 重いモーダル・ダイアログは `lazy()` + `Suspense` で遅延ロード

### プロジェクト内参照実装
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:28-33` — VitalsModal, OwnerSearchModal の正しい lazy パターン
- `frontend/src/features/owners/routes/OwnerForm.tsx` — `PetEditModal = lazy(...)` パターン

## 優先度
**Low** — バンドルサイズへの影響は軽微（91行のシンプルなモーダル）。ただし同ファイル内の一貫性のため修正すべき。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:27-33`
- `frontend/src/features/medical-records/components/StaffSelectionModal.tsx`
