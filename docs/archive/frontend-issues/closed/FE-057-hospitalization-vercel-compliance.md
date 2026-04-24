# FE-057: 入院 — memo/hoisted JSX/barrel index 準拠

**Status**: Open
**Priority**: Medium
**Affects**: `features/hospitalization/`
**Date Created**: 2026-03-18
**Related**: TASK-013

## Summary

入院 feature の Vercel React Best Practices 違反を修正する。
主な修正: 8コンポーネントの memo() 化、static JSX の巻き上げ、barrel index 除去、lazy state init 修正。

## 現状のコード

### 1. memo() 未適用の大型コンポーネント（8件）

```typescript
// routes/HospitalizationForm.tsx:29 — 227行、memo() なし
export function HospitalizationForm() { ... }

// components/HospitalizationDesktopLayout.tsx:38 — 107行、memo() なし
export function HospitalizationDesktopLayout({ ... }) { ... }

// components/HospitalizationMobileLayout.tsx:39 — 115行、memo() なし
export function HospitalizationMobileLayout({ ... }) { ... }

// components/HospitalizationTreatmentTable.tsx:14 — 111行、memo() なし
export function HospitalizationTreatmentTable({ ... }) { ... }

// components/HospitalizationCostSummary.tsx:17 — 114行、memo() なし
export function HospitalizationCostSummary({ ... }) { ... }

// components/HospitalizationBoard.tsx:118 — 177行、memo() なし + CageCard 内部コンポーネントも未 memo
export function HospitalizationBoard({ ... }) { ... }

// components/CarePlan/CarePlanDialog.tsx:40 — 230行、memo() なし
export function CarePlanDialog({ ... }) { ... }

// components/DailyRecord/DailyRecordSection.tsx:23 — 138行、memo() なし
export function DailyRecordSection({ ... }) { ... }
```

### 2. static JSX 未巻き上げ（3件）

```typescript
// components/HospitalizationBasicInfo.tsx:42-58
// RadioGroup items がインライン定義
<RadioGroup value={formData.hospitalizationType}>
  <div><RadioGroupItem value="入院" id="type-hospitalization" /><Label>入院</Label></div>
  <div><RadioGroupItem value="ホテル" id="type-hotel" /><Label>ホテル</Label></div>
</RadioGroup>

// components/HospitalizationTreatmentTable.tsx:41-52
// テーブルヘッダーがインライン
<thead><tr>
  <th>治療内容</th><th>メモ</th>...
</tr></thead>

// components/CarePlan/CarePlanDialog.tsx:173-186
// タイミングボタンがインライン
{["morning", "noon", "night"].map((time) => (...))}
```

### 3. barrel index（HospitalizationDetail.tsx）

```typescript
// routes/HospitalizationDetail.tsx:9-14
import {
  DischargeAlertDialog,
  HospitalizationDetailActions,
  HospitalizationDesktopLayout,
  HospitalizationMobileLayout
} from "../components";  // ❌ barrel index 経由
```

### 4. lazy state init 不備

```typescript
// components/DailyRecord/DailyRecordSection.tsx:24
const [selectedDate, setSelectedDate] = useState(new Date());
// ❌ new Date() が毎レンダーで評価される（実害は小さいが規約違反）
```

### 5. useMemo 不足

```typescript
// components/CarePlan/CarePlanSection.tsx:53-60
{plans.map(plan => (
  <CarePlanItemRow key={plan.id} plan={plan} onEdit={handleOpenEdit} onDelete={onDelete} />
))}
// ❌ CarePlanItemRow が memo() なし、map() が useMemo なし

// components/HospitalizationTreatmentTable.tsx:55-104
{treatmentPlans.map((plan) => (
  <tr key={plan.id}>...</tr>  // 9セルの入力フィールド
))}
// ❌ useMemo なし
```

## 必要な変更

### 1. memo() 追加（8コンポーネント）

各コンポーネントの export を memo() で囲む:

```typescript
// 例: HospitalizationTreatmentTable.tsx
export const HospitalizationTreatmentTable = memo(function HospitalizationTreatmentTable({
  treatmentPlans, onUpdate, onRemove, onAdd
}: Props) {
  // ...existing implementation...
});
```

対象:
- `HospitalizationDesktopLayout`
- `HospitalizationMobileLayout`
- `HospitalizationTreatmentTable`
- `HospitalizationCostSummary`
- `HospitalizationBoard`（+ 内部 CageCard も memo 化）
- `CarePlanDialog`
- `DailyRecordSection`
- `CarePlanItemRow`

### 2. static JSX 巻き上げ

```typescript
// HospitalizationBasicInfo.tsx — モジュールレベルに巻き上げ
const HOSPITALIZATION_TYPE_OPTIONS = [
  { value: "入院", id: "type-hospitalization", label: "入院" },
  { value: "ホテル", id: "type-hotel", label: "ホテル" },
] as const;

// HospitalizationTreatmentTable.tsx
const TREATMENT_TABLE_HEADER = (
  <thead className="bg-[#F7F6F3]">
    <tr>
      <th className="text-left px-3 py-2">治療内容</th>
      <th className="text-left px-3 py-2">メモ</th>
      {/* ... */}
    </tr>
  </thead>
);

// CarePlanDialog.tsx
const TIMING_OPTIONS = [
  { id: "morning", label: "朝" },
  { id: "noon", label: "昼" },
  { id: "night", label: "夜" },
] as const;
```

### 3. barrel index 除去

```typescript
// HospitalizationDetail.tsx — 直接 import に変更
import { DischargeAlertDialog } from "../components/DischargeAlertDialog";
import { HospitalizationDetailActions } from "../components/HospitalizationDetailActions";
import { HospitalizationDesktopLayout } from "../components/HospitalizationDesktopLayout";
import { HospitalizationMobileLayout } from "../components/HospitalizationMobileLayout";
```

`components/index.ts` を削除（存在する場合）。

### 4. lazy state init 修正

```typescript
// DailyRecordSection.tsx:24
const [selectedDate, setSelectedDate] = useState(() => new Date());
```

### 5. useMemo 追加

```typescript
// CarePlanSection.tsx
const planRows = useMemo(() =>
  plans.map(plan => (
    <CarePlanItemRow key={plan.id} plan={plan} onEdit={handleOpenEdit} onDelete={onDelete} />
  )),
  [plans, handleOpenEdit, onDelete]
);
```

### 6. ConfirmDialog の lazy 化（HospitalizationForm.tsx）

```typescript
// Before
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";

// After
const ConfirmDialog = lazy(() =>
  import("@/components/shared/ConfirmDialog/ConfirmDialog").then(m => ({ default: m.ConfirmDialog }))
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

- [ ] 8コンポーネントが memo() で囲まれている
- [ ] static JSX が3箇所でモジュールレベルに巻き上げられている
- [ ] barrel index import が直接 import に変更されている
- [ ] DailyRecordSection の useState が lazy init
- [ ] CarePlanSection の plans.map() が useMemo でキャッシュされている
- [ ] ConfirmDialog が lazy() + Suspense で読み込まれている
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] 入院フォーム・一覧・詳細画面の操作が正常動作
