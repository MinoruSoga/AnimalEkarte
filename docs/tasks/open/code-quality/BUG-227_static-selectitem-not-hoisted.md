# BUG-227: 静的 SelectItem JSX がモジュール定数に巻き上げられていない（4箇所）

## 概要

`memo()` コンポーネント内で、定数であるはずの `<SelectItem>` リストが
インライン JSX として書かれており、レンダーのたびに同じ ReactElement を再生成している。
参照実装 (`owners/routes/OwnerForm.tsx`) では `PET_TABLE_HEADER` 等をモジュール定数として巻き上げている。

## 現状コード（4箇所 — 実コード確認済み）

### 1. `features/vaccinations/routes/VaccinationForm.tsx:221-222,294-297`

```typescript
// ❌ ワクチン種別の選択肢（定数）がコンポーネント内にインライン
<SelectContent>
  <SelectItem value="1">混合ワクチン</SelectItem>
  <SelectItem value="2">狂犬病ワクチン</SelectItem>
</SelectContent>

// ❌ 次回接種スケジュールの選択肢も同様（line 293-298）
<SelectContent>
  <SelectItem value="3weeks">3週後</SelectItem>
  <SelectItem value="4weeks">4週後</SelectItem>
  <SelectItem value="1year">1年後</SelectItem>
  <SelectItem value="custom">以外（手動）</SelectItem>
</SelectContent>
```

### 2. `features/hospitalization/components/CarePlan/CarePlanDialog.tsx:148-152,223-225`

```typescript
// ❌ ケアプランタイプ（定数）がインライン
<SelectItem value="food">食事</SelectItem>
<SelectItem value="medicine">投薬</SelectItem>
<SelectItem value="treatment">処置・検査</SelectItem>
<SelectItem value="instruction">指示・その他</SelectItem>
<SelectItem value="item">持ち物</SelectItem>

// ❌ ステータス（定数）もインライン（line 223-225）
<SelectItem value="active">実施中</SelectItem>
<SelectItem value="completed">完了</SelectItem>
<SelectItem value="discontinued">中止</SelectItem>
```

### 3. `features/accounting/routes/AccountingDetail.tsx:434-437` — `InsuranceCard` (memo) 内

```typescript
// ❌ 保険割合の選択肢（定数）がインライン
<SelectItem value="0.5">50%</SelectItem>
<SelectItem value="0.7">70%</SelectItem>
<SelectItem value="0.9">90%</SelectItem>
<SelectItem value="1.0">100%</SelectItem>
```

### 4. `features/medical-records/components/TreatmentTable.tsx:101-103`

```typescript
// ❌ 処置ステータスの選択肢（定数）がインライン
<SelectItem value="pending">未完了</SelectItem>
<SelectItem value="completed">完了</SelectItem>
<SelectItem value="not_applicable">-</SelectItem>
```

## 比較: 正しい実装（プロジェクト内参照実装）

```typescript
// features/medical-records/routes/MedicalRecordForm.tsx — モジュール定数として巻き上げ済み
const VISIT_TYPE_OPTIONS = (
  <>
    <SelectItem value="outpatient">外来</SelectItem>
    <SelectItem value="hospitalized">入院</SelectItem>
    ...
  </>
);

// features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx
const SHIFT_TYPE_OPTIONS = (
  <>
    <SelectItem value="full">全日</SelectItem>
    <SelectItem value="am">午前</SelectItem>
    ...
  </>
);
```

## 修正方針

各ファイルのコンポーネント定義の外側（モジュールスコープ）に定数を宣言する:

### VaccinationForm.tsx

```typescript
// コンポーネント定義の上（モジュールスコープ）に追加
const VACCINE_TYPE_ITEMS = (
  <>
    <SelectItem value="1">混合ワクチン</SelectItem>
    <SelectItem value="2">狂犬病ワクチン</SelectItem>
  </>
);

const NEXT_SCHEDULE_ITEMS = (
  <>
    <SelectItem value="3weeks">3週後</SelectItem>
    <SelectItem value="4weeks">4週後</SelectItem>
    <SelectItem value="1year">1年後</SelectItem>
    <SelectItem value="custom">以外（手動）</SelectItem>
  </>
);

// JSX 内で使用
<SelectContent>{VACCINE_TYPE_ITEMS}</SelectContent>
<SelectContent>{NEXT_SCHEDULE_ITEMS}</SelectContent>
```

### CarePlanDialog.tsx

```typescript
const CARE_PLAN_TYPE_ITEMS = (
  <>
    <SelectItem value="food">食事</SelectItem>
    <SelectItem value="medicine">投薬</SelectItem>
    <SelectItem value="treatment">処置・検査</SelectItem>
    <SelectItem value="instruction">指示・その他</SelectItem>
    <SelectItem value="item">持ち物</SelectItem>
  </>
);

const CARE_PLAN_STATUS_ITEMS = (
  <>
    <SelectItem value="active">実施中</SelectItem>
    <SelectItem value="completed">完了</SelectItem>
    <SelectItem value="discontinued">中止</SelectItem>
  </>
);
```

### AccountingDetail.tsx (InsuranceCard 用)

```typescript
const INSURANCE_RATIO_ITEMS = (
  <>
    <SelectItem value="0.5">50%</SelectItem>
    <SelectItem value="0.7">70%</SelectItem>
    <SelectItem value="0.9">90%</SelectItem>
    <SelectItem value="1.0">100%</SelectItem>
  </>
);
```

### TreatmentTable.tsx

```typescript
const TREATMENT_STATUS_ITEMS = (
  <>
    <SelectItem value="pending">未完了</SelectItem>
    <SelectItem value="completed">完了</SelectItem>
    <SelectItem value="not_applicable">-</SelectItem>
  </>
);
```

## 影響範囲

| ファイル | 行 | 定数の種類 |
|---------|-----|----------|
| `features/vaccinations/routes/VaccinationForm.tsx` | 221-222,294-297 | ワクチン種別 + スケジュール選択肢 |
| `features/hospitalization/components/CarePlan/CarePlanDialog.tsx` | 148-152,223-225 | ケアプランタイプ + ステータス |
| `features/accounting/routes/AccountingDetail.tsx` | 434-437 | 保険割合（InsuranceCard 内） |
| `features/medical-records/components/TreatmentTable.tsx` | 101-103 | 処置ステータス |

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — `rendering-hoist-jsx`
> 静的 JSX（Select 選択肢など）はモジュール定数に巻き上げ

### `.claude/rules/code-style.md` — Performance Rules
> `rendering-hoist-jsx`: コンポーネント外の静的 JSX（Select 選択肢など）はモジュール定数に巻き上げ

### プロジェクト内参照実装
- `features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx` — `SHIFT_TYPE_OPTIONS`
- `features/medical-records/routes/MedicalRecordForm.tsx` — `VISIT_TYPE_OPTIONS`, `MEDICAL_RECORD_TABS`

## 優先度

**Low** — レンダーコストへの影響は極めて小さい（SelectItem は単純な DOM 構造）。
ただしプロジェクトの一貫したコーディングスタイル維持のため対応すべき。
修正コストも低い（各ファイルで 10行程度の変更）。
