# BUG-240: rerender-memo — ReceptionDetailModal の RelatedPages / ActionButtons が memo() 未適用

## 概要
`ReceptionDetailModal`（482行）は `memo()` でラップされているが、内部のサブコンポーネント `RelatedPages`（70行、7コールバック props）と `ActionButtons`（150行以上、12コールバック props）は `memo()` なし。

`currentStatus` が変わって `ReceptionDetailModal` が再レンダーされた場合、`RelatedPages` は `currentStatus` を参照していないのに必ず再レンダーされる。親が `memo()` 化された意図（不要な再レンダー防止）を、子のサブコンポーネントが実現できていない。

## 現状コード

### `features/reception/components/ReceptionDetailModal.tsx:52,128`
```typescript
// ❌ memo() なし — 親が再レンダーされるたびに必ず再レンダー
function RelatedPages({
  isTrimming,
  onCreateMedicalRecord,
  onCreateTrimming,
  onCreateAccounting,
  onCreateHospitalization,
  canCreateMedicalRecord = false,
  canCreateAccounting = false,
  canCreateHospitalization = false,
}: RelatedPagesProps) { ... }

// ❌ 同様に memo() なし
function ActionButtons({
  currentStatus,
  appointment,
  isTrimming,
  isHospitalization,
  isMedical,
  onConfirm,
  onEdit,
  onCancel,
  onOpenOwnerDetail,
  onCreateMedicalRecord,
  onCreateTrimming,
  onCreateAccounting,
  onCreateHospitalization,
}: ActionButtonsProps) { ... }
```

### 親の `ReceptionDetailModal` は memo 済み
```typescript
// line 304
export const ReceptionDetailModal = memo(function ReceptionDetailModal({ ... }) { ... });
```

### コールバックは useCallback で安定化済み（lines 323-355）
```typescript
const handleCreateMedicalRecord = useCallback((tab?: string) => { ... }, [petId, appointmentId, navigateAndClose]);
const handleCreateTrimming = useCallback(() => ..., [petId, navigateAndClose]);
// ...など全コールバックが useCallback 済み
```

## 問題の具体例

1. `ReceptionDetailModal` に渡される `currentStatus`（"受付予約" → "受付済"）が変化
2. `ReceptionDetailModal` が再レンダーされる（memo の効果で必然）
3. `RelatedPages` は `currentStatus` を受け取らず、変化に無関係 → **本来再レンダー不要**
4. しかし `memo()` がないため、`RelatedPages` も必ず再レンダーされる

## 修正方針

```typescript
// ✅ RelatedPages を memo でラップ
const RelatedPages = memo(function RelatedPages({
  isTrimming,
  onCreateMedicalRecord,
  onCreateTrimming,
  onCreateAccounting,
  onCreateHospitalization,
  canCreateMedicalRecord = false,
  canCreateAccounting = false,
  canCreateHospitalization = false,
}: RelatedPagesProps) {
  return (
    // ...既存の JSX そのまま
  );
});

// ✅ ActionButtons も memo でラップ
const ActionButtons = memo(function ActionButtons({
  currentStatus,
  appointment,
  // ... 他 props
}: ActionButtonsProps) {
  // ...既存の実装そのまま
});
```

**注意**: コールバック props はすでに `useCallback` で安定化済み（lines 323-355）のため、`memo()` 追加だけで効果が出る。

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — rerender-memo
> `memo()` でコンポーネント再レンダー防止。必ず props ハンドラを `useCallback` で安定化すること。（コールバックは安定化済み ✅）

### プロジェクト内参照実装
`features/owners/routes/OwnerForm.tsx` — `OwnerInfoSection`, `PetTableRow` が `memo()` + `useCallback` ハンドラのパターン

## 優先度
**Medium** — 受付画面はユーザーが頻繁に操作する（ステータス変更ごとに再レンダー）。`RelatedPages` は `currentStatus` と無関係なため効果が大きい。修正は10分。

## 関連ファイル
- `frontend/src/features/reception/components/ReceptionDetailModal.tsx:52,128,304,323-355,447,462`
