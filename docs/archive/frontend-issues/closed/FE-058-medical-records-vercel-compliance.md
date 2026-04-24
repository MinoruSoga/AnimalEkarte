# FE-058: カルテ — memo/hoisted static data 準拠

**Status**: Open
**Priority**: Medium
**Affects**: `features/medical-records/`
**Date Created**: 2026-03-18
**Related**: TASK-013

## Summary

カルテ feature の Vercel React Best Practices 違反を修正する。
主な修正: 9コンポーネントの memo() 化、4箇所の static data 巻き上げ、TreatmentTable の useMemo 追加。

## 現状のコード

### 1. memo() 未適用の大型コンポーネント（9件）

```typescript
// routes/MedicalRecordForm.tsx:34 — 402行、9タブのフォーム親
export function MedicalRecordForm() { ... }
// 問題: 任意のタブの state 変更で全9タブが再レンダーされる

// components/MedicalRecordInterview.tsx — 99行、3カラムレイアウト
export function MedicalRecordInterview({ ... }) { ... }

// components/MedicalRecordDiagnosisPlan.tsx:37 — 221行、治療テーブル + 診断ヘッダ
export function MedicalRecordDiagnosisPlan({ ... }) { ... }

// components/MedicalRecordExamination.tsx — 66行
export function MedicalRecordExamination({ ... }) { ... }

// components/MedicalRecordVaccination.tsx — 67行
export function MedicalRecordVaccination({ ... }) { ... }

// components/MedicalRecordImage.tsx — 63行
export function MedicalRecordImage({ ... }) { ... }

// components/MedicalRecordEstimate.tsx:9 — 137行
export function MedicalRecordEstimate({ ... }) { ... }

// components/MedicalRecordBillCheck.tsx:19 — 182行
export function MedicalRecordBillCheck({ ... }) { ... }

// components/TreatmentTable.tsx:43 — 285行、9カラム×N行のテーブル
export function TreatmentTable({ ... }) { ... }
```

**注**: 以下は既に memo() 適用済み（変更不要）:
- `VaccinationForm.tsx:41` — ✅ memo()
- `DiagnosisHeader.tsx:24` — ✅ memo()
- `StaffSelectionModal.tsx:18` — ✅ memo()

### 2. static data がコンポーネント内部で定義（4件）

```typescript
// components/MedicalRecordDiagnosisPlan.tsx:57-62
export function MedicalRecordDiagnosisPlan({...}) {
  const templates = [
    { label: "定期検診", text: "# 定期検診\n特に異常なし。食欲・元気あり。" },
    { label: "ワクチン", text: "# 混合ワクチン接種\n体調良好。" },
    // ...
  ];
  // ❌ 毎レンダーで再生成

// components/MedicalRecordVaccination.tsx:21-35
  const historyItems = [
    { id: 1, name: "フィラリア薬", date: "24/4/6", next: "24/4/6" },
    // ...
  ];

// components/MedicalRecordExamination.tsx:15-39
  const examGroups = isNewRecord ? [] : [
    { id: 1, date: "2025/10/10 10:15", machine: "血液検査機器 A", ... },
    // ...
  ];

// components/MedicalRecordImage.tsx:16-34
  const imageGroups = isNewRecord ? [] : [
    { id: 1, date: "2025/10/10 10:10:10", images: [ ... ] },
    // ...
  ];
```

### 3. TreatmentTable — useMemo なし

```typescript
// components/TreatmentTable.tsx:80-183
{items.map((item) => (
  <div key={item.id} className={...}>
    {/* 9カラムの入力フィールド行 */}
  </div>
))}
// ❌ 60行以上の JSX を毎レンダーで再生成
```

## 必要な変更

### 1. memo() 追加（9コンポーネント）

```typescript
// 例: TreatmentTable.tsx
export const TreatmentTable = memo(function TreatmentTable({
  items, onUpdate, onRemove, onOpenSearch, onAddRow, showStatus = false
}: TreatmentTableProps) {
  // ...existing implementation...
});
```

対象リスト:
- `MedicalRecordForm` — ルートフォーム（タブ切り替え最適化）
- `MedicalRecordInterview`
- `MedicalRecordDiagnosisPlan`
- `MedicalRecordExamination`
- `MedicalRecordVaccination`
- `MedicalRecordImage`
- `MedicalRecordEstimate`
- `MedicalRecordBillCheck`
- `TreatmentTable`

### 2. static data モジュールレベル巻き上げ

```typescript
// MedicalRecordDiagnosisPlan.tsx — コンポーネント外に移動
const DIAGNOSIS_TEMPLATES = [
  { label: "定期検診", text: "# 定期検診\n特に異常なし。食欲・元気あり。" },
  { label: "ワクチン", text: "# 混合ワクチン接種\n体調良好。" },
  // ...
] as const;

// MedicalRecordVaccination.tsx
const MOCK_HISTORY_ITEMS = [
  { id: 1, name: "フィラリア薬", date: "24/4/6", next: "24/4/6" },
  // ...
] as const;

// MedicalRecordExamination.tsx
const MOCK_EXAM_GROUPS = [
  { id: 1, date: "2025/10/10 10:15", machine: "血液検査機器 A", ... },
  // ...
] as const;

// MedicalRecordImage.tsx
const MOCK_IMAGE_GROUPS = [
  { id: 1, date: "2025/10/10 10:10:10", images: [ ... ] },
  // ...
] as const;
```

**注**: `isNewRecord` 条件で空配列を返すケースは `useMemo` で対応:
```typescript
const examGroups = useMemo(
  () => isNewRecord ? [] : MOCK_EXAM_GROUPS,
  [isNewRecord]
);
```

### 3. TreatmentTable の useMemo 追加

```typescript
// TreatmentTable.tsx
const tableRows = useMemo(() =>
  items.map((item) => (
    <div key={item.id} className={...}>
      {/* 9カラムの入力フィールド行 */}
    </div>
  )),
  [items, onUpdate, onRemove, showStatus]
);

return (
  <div>
    {TREATMENT_TABLE_HEADER}
    {tableRows}
  </div>
);
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理
- [ ] 型は `models.ts` から導出

## 依存関係

- 依存なし（独立して着手可能）

## 完了条件

- [ ] 9コンポーネントが memo() で囲まれている
- [ ] 4箇所の static data がモジュールレベルに巻き上げられている
- [ ] TreatmentTable の items.map() が useMemo でキャッシュされている
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] カルテフォーム全9タブの操作が正常動作
