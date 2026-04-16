# BUG-236: hospitalization・medical-records コンポーネントで `bg-white` をデザイントークン未使用

## 概要

`features/hospitalization/components/` の 7 ファイル（11 箇所）と `features/medical-records/components/MedicalRecordBillCheck.tsx`（2 箇所）で、Tailwind クラス `bg-white` をデザイントークン `C.bgWhite` に通さず直接使用している。BUG-228（vaccinations/examinations/inventory）・BUG-233（master DragOverlay）と同種の違反。

## 再現手順

（ランタイム動作に変化なし。コードの一貫性違反）

## 期待する動作

- `bg-white` → `${C.bgWhite}` に統一

## 現状コード

### hospitalization コンポーネント（11 箇所）

```
HospitalizationBasicInfo.tsx:28,96,117
HospitalizationBoard.tsx:68
HospitalizationCostSummary.tsx:27
HospitalizationNoteCard.tsx:25,35
HospitalizationTabbedView.tsx:53,65
HospitalizationTreatmentTable.tsx:24
HospitalizationExpandedView.tsx:41,53
```

#### 代表例 `HospitalizationBasicInfo.tsx:28`
```tsx
// ❌
<div className={`bg-white rounded-lg shadow-sm border ${C.borderMedium} ...`}>
```

### medical-records コンポーネント（2 箇所）

```
medical-records/components/MedicalRecordBillCheck.tsx:138,163
```

#### `MedicalRecordBillCheck.tsx:138`
```tsx
// ❌
<div className={`flex flex-col items-center justify-center p-12 bg-white rounded-lg border border-dashed ${C.text40}`}>
```

## 影響範囲

| 対象 | 行 | 状態 |
|------|-----|------|
| `features/hospitalization/components/HospitalizationBasicInfo.tsx` | 28, 96, 117 | 未修正 |
| `features/hospitalization/components/HospitalizationBoard.tsx` | 68 | 未修正 |
| `features/hospitalization/components/HospitalizationCostSummary.tsx` | 27 | 未修正 |
| `features/hospitalization/components/HospitalizationNoteCard.tsx` | 25, 35 | 未修正 |
| `features/hospitalization/components/HospitalizationTabbedView.tsx` | 53, 65 | 未修正 |
| `features/hospitalization/components/HospitalizationTreatmentTable.tsx` | 24 | 未修正 |
| `features/hospitalization/components/HospitalizationExpandedView.tsx` | 41, 53 | 未修正 |
| `features/medical-records/components/MedicalRecordBillCheck.tsx` | 138, 163 | 未修正 |

## 修正方針

各ファイルで `bg-white` → `${C.bgWhite}` に一括置換する。

```tsx
// Before
<div className={`bg-white rounded-lg border ...`}>

// After
<div className={`${C.bgWhite} rounded-lg border ...`}>
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず **`C`**, **`STYLE`** 定数を使用

`C.bgWhite = "bg-white"` がトークンとして定義済み（`design-tokens.ts:239`）。

## 優先度
**Low** — 視覚的変化なし。BUG-228・BUG-233 の修正と同時に対応推奨（統一修正バッチとして）。

## 関連チケット
- BUG-228: vaccinations/examinations/inventory の同種違反（14 箇所）
- BUG-233: master DragOverlay の同種違反（3 箇所）

## 関連ファイル
- `frontend/src/features/hospitalization/components/` — 7 ファイル
- `frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx`
