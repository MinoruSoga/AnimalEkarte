# BUG-281: DataStates/ に barrel index.ts が欠落 — 17ファイルが deep import 違反

## 概要
`components/shared/DataStates/` に `index.ts` が存在しないため、このコンポーネントを利用する17ファイルが全て `DataStates.tsx` へ直接パスを指定する deep import 違反を犯している。Feature Indexing ルール違反。

## 再現手順
1. `ls frontend/src/components/shared/DataStates/` → `DataStates.tsx` のみ、`index.ts` が存在しない
2. `grep -r "DataStates/DataStates" frontend/src --include="*.tsx" --include="*.ts" -l` → 17ファイルが直接パス指定

**実際の import（違反例）:**
```typescript
// frontend/src/features/accounting/routes/AccountingList.tsx:29
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates/DataStates";
//                                                                              ^^^^^^^^^^^
//                                                                              ファイル直接指定（deep import）
```

## 期待する動作
```typescript
// barrel index.ts 経由でimport
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
```

## 現状コード

### `frontend/src/components/shared/DataStates/` — index.ts が存在しない
```
DataStates/
└── DataStates.tsx   ← ファイル名まで指定しないとimportできない状態
（index.ts が存在しない）
```

### 比較: 正しい実装（プロジェクト内参照実装）
```typescript
// frontend/src/components/shared/ConfirmDialog/index.ts（正しい barrel 例）
export { ConfirmDialog } from './ConfirmDialog';
```

## 影響範囲

| 対象ファイル | 詳細 | 状態 |
|------------|------|------|
| `features/accounting/routes/AccountingList.tsx:29` | deep import | 要修正 |
| `features/accounting/routes/AccountingDetail.tsx:28` | deep import | 要修正 |
| `features/trimming/routes/TrimmingForm.tsx:24` | deep import | 要修正 |
| `features/trimming/routes/TrimmingList.tsx:33` | deep import | 要修正 |
| `features/estimates/routes/EstimateDetail.tsx:2` | deep import | 要修正 |
| `features/estimates/routes/EstimateForm.tsx:2` | deep import | 要修正 |
| `features/estimates/routes/EstimateList.tsx:2` | deep import | 要修正 |
| `features/checkups/routes/CheckupsList.tsx:20` | deep import | 要修正 |
| `features/medical-records/components/ClinicalPlanSection/ClinicalPlanSection.tsx:7` | deep import | 要修正 |
| `features/medical-records/routes/MedicalRecordForm.tsx:15` | deep import | 要修正 |
| `features/medical-records/routes/MedicalRecords.tsx:23` | deep import | 要修正 |
| `features/inventory/routes/InventoryList.tsx` | deep import | 要修正 |
| `features/hospitalization/routes/HospitalizationList.tsx` | deep import | 要修正 |
| `features/hospitalization/routes/HospitalizationDetail.tsx` | deep import | 要修正 |
| `features/hospitalization/routes/HospitalizationForm.tsx` | deep import | 要修正 |
| （他2ファイル） | deep import | 要修正 |
| **合計: 17ファイル** | | |

## 修正方針

### 1. barrel ファイル作成 — `frontend/src/components/shared/DataStates/index.ts`
```typescript
export { LoadingFallback, ErrorFallback } from './DataStates';
```

### 2. 全17ファイルの import パスを修正

**修正前:**
```typescript
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates/DataStates";
```

**修正後:**
```typescript
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
```

修正対象ファイルは全て同パターンのため、一括置換で対応可能:
```bash
# 確認
grep -r "DataStates/DataStates" frontend/src --include="*.tsx" --include="*.ts" -l

# 置換（各ファイルで実施）
# "@/components/shared/DataStates/DataStates"
# → "@/components/shared/DataStates"
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Feature Indexing
> **Deep imports from features**: `@/features/xxx/components/YYY` などの深掘り import は禁止。必ず feature の `index.ts` (Feature Indexing) を経由すること。

同じ原則が `components/shared/` にも適用される。ディレクトリにはBarrel `index.ts` を置くことで外部から実装詳細を隠蔽すべきである。

### プロジェクト内参照実装
- `components/shared/ConfirmDialog/index.ts` — barrel パターンの正しい実装
- `components/shared/NotionDatePicker/index.ts` — 同様

## 優先度
**Medium** — 動作への影響はないが、17ファイルがルール違反状態。`DataStates.tsx` をリネームした際に全17ファイルが壊れるリスクがある。

## 関連チケット
- BUG-280: 安全削除対象コンポーネント
- BUG-282: SidePeek barrel index.ts 欠落（同様の問題）

## 関連ファイル
- `frontend/src/components/shared/DataStates/DataStates.tsx:1` — barrel 欠落の根本
