# BUG-282: SidePeek/ に barrel index.ts が欠落 — master 全16設定画面が deep import 違反

## 概要
`components/shared/SidePeek/` に `index.ts` が存在しないため、`features/master/routes/` 配下の16設定画面ファイルが全て各サブコンポーネントへ直接パスを指定する deep import 違反を犯している。

## 再現手順
1. `ls frontend/src/components/shared/SidePeek/` → index.ts が存在しない（10コンポーネントファイルのみ）
2. `grep -r "from.*SidePeek/" frontend/src --include="*.tsx" --include="*.ts" -l` → 19ファイル（うち3ファイルはSidePeek内部の自己参照）

**実際の import（違反例）:**
```typescript
// frontend/src/features/master/routes/ServiceTypeSettings.tsx
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropertyInput } from "@/components/shared/SidePeek/PropertyInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
//                                                          ^^^^^^^^^^^^^^^^^^^
//                                                          サブコンポーネントへの直接パス指定（deep import）
```

## 期待する動作
```typescript
// barrel index.ts 経由でまとめてimport
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
```

## 現状コード

### `frontend/src/components/shared/SidePeek/` — index.ts が存在しない
```
SidePeek/
├── MasterSidePanel.tsx
├── MoneyInput.tsx
├── PropertyInput.tsx
├── PropertyRow.tsx
├── SidePeekBody.tsx
├── SidePeekFooter.tsx
├── SidePeekPanel.tsx
├── SidePeekTitleInput.tsx
├── SidePeekToolbar.tsx
└── StatusToggleButton.tsx
（index.ts が存在しない → 全てのファイルに直接パス指定が必要な状態）
```

### 比較: 正しい実装（プロジェクト内参照実装）
```typescript
// components/shared/ConfirmDialog/index.ts（正しい barrel 例）
export { ConfirmDialog } from './ConfirmDialog';
```

## 影響範囲

| 対象ファイル | 詳細 | 状態 |
|------------|------|------|
| `features/master/routes/ServiceTypeSettings.tsx` | PropertyRow/StatusToggleButton/PropertyInput/MasterSidePanel deep import | 要修正 |
| `features/master/routes/InterviewTemplateSettings.tsx` | PropertyRow/StatusToggleButton/MasterSidePanel deep import | 要修正 |
| `features/master/routes/MerchandiseItemSettings.tsx` | PropertyRow/StatusToggleButton/MoneyInput/MasterSidePanel deep import | 要修正 |
| `features/master/routes/HospitalizationSettings.tsx` | deep import | 要修正 |
| `features/master/routes/PermissionGroupSettings.tsx` | deep import | 要修正 |
| `features/master/routes/AnimalSpeciesSettings.tsx` | deep import | 要修正 |
| `features/master/routes/ChiefComplaintSettings.tsx` | deep import | 要修正 |
| `features/master/routes/CageSettings.tsx` | deep import | 要修正 |
| `features/master/routes/CompanySettings.tsx` | deep import | 要修正 |
| `features/master/routes/InsuranceSettings.tsx` | deep import | 要修正 |
| `features/master/routes/StaffSettings.tsx` | deep import | 要修正 |
| `features/master/routes/DiagnosisSettings.tsx` | deep import | 要修正 |
| `features/master/routes/OccupationSettings.tsx` | deep import | 要修正 |
| `features/master/routes/MedicineSettings.tsx` | deep import | 要修正 |
| `features/master/routes/TrimmingSettings.tsx` | deep import | 要修正 |
| `features/master/routes/TreatmentPlanMaster.tsx` | deep import | 要修正 |
| **合計: 16ファイル** | | |

## 修正方針

### 1. barrel ファイル作成 — `frontend/src/components/shared/SidePeek/index.ts`
```typescript
export { MasterSidePanel } from './MasterSidePanel';
export { MoneyInput } from './MoneyInput';
export { PropertyInput } from './PropertyInput';
export { PropertyRow } from './PropertyRow';
export { SidePeekBody } from './SidePeekBody';
export { SidePeekFooter } from './SidePeekFooter';
export { SidePeekPanel } from './SidePeekPanel';
export { SidePeekTitleInput } from './SidePeekTitleInput';
export { SidePeekToolbar } from './SidePeekToolbar';
export { StatusToggleButton } from './StatusToggleButton';
```

### 2. 全16ファイルの import を barrel 経由に変更

**修正前（ServiceTypeSettings.tsx の例）:**
```typescript
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { StatusToggleButton } from "@/components/shared/SidePeek/StatusToggleButton";
import { PropertyInput } from "@/components/shared/SidePeek/PropertyInput";
import { MasterSidePanel } from "@/components/shared/SidePeek/MasterSidePanel";
```

**修正後:**
```typescript
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Feature Indexing
> **Deep imports from features**: `@/features/xxx/components/YYY` などの深掘り import は禁止。必ず feature の `index.ts` (Feature Indexing) を経由すること。

同じ原則が `components/shared/` のサブディレクトリにも適用される。

### プロジェクト内参照実装
- `components/shared/ConfirmDialog/index.ts` — barrel パターンの正しい実装

## 優先度
**Medium** — 動作への影響はないが、16ファイルがルール違反状態。SidePeek のサブコンポーネントをリファクタリングした場合に全16ファイルが破壊されるリスクがある。

## 関連チケット
- BUG-281: DataStates barrel index.ts 欠落（同様の問題）
- BUG-283: deprecated use-master-items 依存残存

## 関連ファイル
- `frontend/src/components/shared/SidePeek/` — barrel 欠落のディレクトリ
- `frontend/src/features/master/routes/` — 影響を受ける全設定画面
