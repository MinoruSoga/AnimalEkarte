# BUG-225: StaffSelectionModal が lazy() でロードされていない

## 概要

`MedicalRecordForm.tsx` で `VitalsModal` と `OwnerSearchModal` は `lazy()` でロードされているが、
同じファイルで使われる `StaffSelectionModal` だけ静的インポートのまま。
モーダルは初期レンダー時には不要なため、遅延ロードすべき。

## 現状コード

### `features/medical-records/routes/MedicalRecordForm.tsx:27-31`

```typescript
// ❌ StaffSelectionModal は静的インポート
import { StaffSelectionModal } from "../components/StaffSelectionModal";

// ✅ 同じファイル内で VitalsModal, OwnerSearchModal は lazy
const VitalsModal = lazy(() =>
  import("../components/VitalsModal").then((m) => ({ default: m.VitalsModal }))
);
const OwnerSearchModal = lazy(() =>
  import("../components/OwnerSearchModal").then((m) => ({ default: m.OwnerSearchModal }))
);
```

`StaffSelectionModal` (91行) は `CommandDialog`/`CommandInput`/`CommandList` (Radix UI / cmdk) を
インポートしており、他のモーダルと同様に初回レンダー時には不要なコードが含まれる。

## 修正方針

```typescript
// StaffSelectionModal を他モーダルと同じパターンで lazy 化
const StaffSelectionModal = lazy(() =>
  import("../components/StaffSelectionModal").then((m) => ({
    default: m.StaffSelectionModal,
  }))
);
```

既存の `Suspense` ブロック（line 491-499, 511-518）を利用するか、
`StaffSelectionModal` の呼び出し箇所を `Suspense fallback={null}` でラップする:

```tsx
<Suspense fallback={null}>
  <StaffSelectionModal
    isOpen={isStaffSelectionOpen}
    onClose={...}
    onSelect={...}
  />
</Suspense>
```

## 影響範囲

| ファイル | 行 | 内容 |
|---------|-----|------|
| `features/medical-records/routes/MedicalRecordForm.tsx` | 27 | static import を lazy() に変更 |
| `features/medical-records/routes/MedicalRecordForm.tsx` | ~502 | Suspense でラップ |

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — `bundle-dynamic-imports`
> 重いモーダル・ダイアログは `lazy()` + `Suspense` で遅延ロード

### Vercel React Best Practices — `bundle-dynamic-imports`
> Use dynamic imports for components not needed on initial render.

### プロジェクト内参照実装
同一ファイル `MedicalRecordForm.tsx:28-33` — `VitalsModal`, `OwnerSearchModal` の lazy パターン

## 優先度

**Low** — `StaffSelectionModal` は 91行と小規模。ただし一貫性のためと、
`CommandDialog` (cmdk) バンドルを分離する観点から対応すべき。
