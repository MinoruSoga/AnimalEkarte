# BUG-222: useCallback deps にオブジェクト・配列を直接渡している（9箇所）

## 概要

`useCallback` の依存配列にオブジェクトや配列を渡している箇所が 9箇所ある。
オブジェクトは毎レンダーで参照が変わるため、`useCallback` のメモ化が毎回無効化される。
結果として `memo()` でラップされた子コンポーネントに新しい関数参照が伝わり、意図した最適化が機能しない。

## 現状コード（9箇所 — 実コード確認済み）

### 1. `features/reception/routes/Reception.tsx:122,165` — `columns`（配列）

```typescript
const handleDragOver = useCallback((event: DragOverEvent) => {
  const sourceColumn = columns.find(col => ...);  // columns を参照
  moveCard(...);
}, [columns, moveCard]);  // ❌ columns は配列オブジェクト

const handleDragEnd = useCallback((event: DragEndEvent) => {
  const sourceColumn = columns.find(col => ...);  // columns を参照
  ...
}, [columns, moveCard]);  // ❌ 同上
```

### 2. `features/reception/routes/Reception.tsx:203` — `selectedAppointment`（オブジェクト）

```typescript
const handleAdvanceStatus = useCallback(() => {
  if (!selectedAppointment) return;
  advanceStatus(selectedAppointment);
}, [selectedAppointment, advanceStatus]);  // ❌ selectedAppointment はオブジェクト
```

### 3. `features/reception/routes/Reception.tsx:250` — `editingAppointment`（オブジェクト）

```typescript
const handleEditSave = useCallback((data, selectedPets) => {
  if (!editingAppointment?.id || !data.start) return;
  ...
}, [editingAppointment, updateAppointment]);  // ❌ editingAppointment はオブジェクト
```

### 4. `features/reception/routes/Reception.tsx:265` — `cancelTarget`（オブジェクト）

```typescript
const executeCancel = useCallback(() => {
  if (!cancelTarget) return;
  cancelAppointment(cancelTarget.id);
}, [cancelTarget, cancelAppointment]);  // ❌ cancelTarget はオブジェクト
```

### 5. `features/hospitalization/routes/HospitalizationDetail.tsx:42` — `hospitalization`（オブジェクト）

```typescript
const handleDischarge = useCallback(() => {
  if (!hospitalization) return;
  // hospitalization.id を使用
  dischargeHospitalization(...);
}, [dischargeHospitalization, navigate, hospitalization]);  // ❌ hospitalization はオブジェクト
```

### 6. `features/estimates/routes/EstimateList.tsx:178` — `deleteModal`（オブジェクト）

```typescript
const handleDeleteConfirm = useCallback(() => {
  if (deleteModal.item == null) return;
  deleteEstimate(deleteModal.item);
  deleteModal.close();
}, [deleteModal, deleteEstimate]);  // ❌ deleteModal はオブジェクト
```

### 7. `features/master/routes/MedicineSettings.tsx:692` — `formData`, `selectedMedicine`（オブジェクト）

```typescript
const handleSave = useCallback(() => {
  // formData, selectedMedicine を使用
}, [formData, selectedMedicine, updateMutation, createMutation, handleCloseEdit]);
// ❌ formData, selectedMedicine は両方オブジェクト
```

### 8. `features/master/routes/DiagnosisSettings.tsx:144` — `formData`（オブジェクト）

```typescript
const handleSave = useCallback(() => {
  if (!formData.name) { ... }
  onSave(formData);
}, [formData, onSave]);  // ❌ formData はオブジェクト
```

## 比較: 正しい実装（プロジェクト内参照実装）

```typescript
// features/owners/routes/OwnersList.tsx — primitive deps パターン
const handleDeleteClick = useCallback((ownerId: number) => {
  setPendingDeleteOwnerId(ownerId);  // ID（primitive）を状態に保存
}, []);  // deps 空

const handleDeleteConfirm = useCallback(() => {
  if (pendingDeleteOwnerId === null) return;  // number（primitive）を使用
  startDeleteTransition(() => {
    deleteOwnerFn(String(pendingDeleteOwnerId), { ... });
  });
}, [pendingDeleteOwnerId, deleteOwnerFn]);  // ✅ primitive deps のみ
```

## 修正方針

### パターン A: オブジェクトから primitive を抽出（推奨）

```typescript
// Before (BUG-222-4: cancelTarget オブジェクトが deps に)
const executeCancel = useCallback(() => {
  if (!cancelTarget) return;
  cancelAppointment(cancelTarget.id);
}, [cancelTarget, cancelAppointment]);

// After: ID だけを state に持ち、useCallback deps を primitive にする
const [cancelTargetId, setCancelTargetId] = useState<string | null>(null);

const executeCancel = useCallback(() => {
  if (!cancelTargetId) return;
  cancelAppointment(cancelTargetId);
}, [cancelTargetId, cancelAppointment]);  // ✅ primitive
```

### パターン B: useRef でトランジェント値を保持（高頻度変化する場合）

```typescript
// formData のように頻繁に変わるオブジェクトの場合
const formDataRef = useRef(formData);
useEffect(() => { formDataRef.current = formData; }, [formData]);

const handleSave = useCallback(() => {
  onSave(formDataRef.current);
}, [onSave]);  // ✅ deps からオブジェクトを除外
```

### パターン C: useActionState でオブジェクト deps を廃止

フォーム系は useActionState に移行することで `formData` を deps から排除できる:

```typescript
const [, formAction] = useActionState(async (_, fd: FormData) => {
  const name = fd.get("name") as string;
  await onSave({ name });
}, null);

<form action={formAction}>
  <Input name="name" defaultValue={formData.name} />
  <SubmitButton>保存</SubmitButton>
</form>
```

## 影響範囲

| ファイル | 行 | 問題の deps |
|---------|-----|------------|
| `features/reception/routes/Reception.tsx` | 122,165 | `columns`（配列） |
| `features/reception/routes/Reception.tsx` | 203 | `selectedAppointment` |
| `features/reception/routes/Reception.tsx` | 250 | `editingAppointment` |
| `features/reception/routes/Reception.tsx` | 265 | `cancelTarget` |
| `features/hospitalization/routes/HospitalizationDetail.tsx` | 42 | `hospitalization` |
| `features/estimates/routes/EstimateList.tsx` | 178 | `deleteModal` |
| `features/master/routes/MedicineSettings.tsx` | 692 | `formData`, `selectedMedicine` |
| `features/master/routes/DiagnosisSettings.tsx` | 144 | `formData` |

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — `rerender-dependencies`
> `useCallback` deps にオブジェクトを入れない — primitive を抽出して使う

### プロジェクト内参照実装
`features/owners/routes/OwnersList.tsx` — `pendingDeleteOwnerId`（number）のみを deps に渡す

## 優先度

**Medium** — `memo()` による最適化が意図通りに機能していない。ただし機能的な不具合はなし。

## 関連チケット

- BUG-221: mutation の useTransition 漏れ（同じファイルを修正する際に合わせて対応推奨）
