# FE-097: 会計(医師確認)タブ - 「行を追加」ボタンがno-op

## 概要
会計(医師確認)タブの「行を追加」ボタンをクリックしても何も起きない。`TreatmentTable` に `onAddRow` も `onOpenSearch` も渡されていないため。

## 対象ファイル
- `frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx`
- `frontend/src/features/medical-records/components/TreatmentTable.tsx`

## 問題箇所
```tsx
// MedicalRecordBillCheck.tsx
<TreatmentTable
  items={items}
  onUpdate={handleUpdateItem}
  onRemove={handleRemoveItem}
  showStatus={false}
  // ❌ onAddRow / onOpenSearch が未渡し
/>

// TreatmentTable.tsx
<Button onClick={onOpenSearch || onAddRow}>
  行を追加
</Button>
// onOpenSearch=undefined, onAddRow=undefined → onClick=undefined → no-op
```

## 期待動作
- 「行を追加」クリック → 処置検索ダイアログを開く（`TreatmentSearchDialog`を利用）
- 選択した処置を会計明細に追加する（`useTreatments` の追加ミューテーションを呼ぶ）

## 修正案
```tsx
// MedicalRecordBillCheck.tsx
const [isSearchOpen, setIsSearchOpen] = useState(false);
const createTreatmentMutation = useCreateTreatment(medicalRecordId);

const handleAddTreatment = useCallback((treatment) => {
  createTreatmentMutation.mutate({ ...treatment, selected: true });
}, [createTreatmentMutation]);

<TreatmentTable
  ...
  onOpenSearch={() => setIsSearchOpen(true)}
/>
```

## 優先度
Medium

## 発見日
2026-03-24（機能テスト中）
