# BUG-196: TreatmentSearchDialog の master feature への Deep Import 違反

## 概要

`components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx` が `@/features/master/api/consultations`・`@/features/master/api/procedures` を直接 import している。これらのフックは `@/features/master/index.ts`（Public API）に公開されておらず、Feature Indexing ルール違反かつ未公開 API への依存となっている。

## 再現手順

1. `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx` を確認
2. **結果**: `@/features/master/api/consultations` および `@/features/master/api/procedures` への直接 deep import が存在する

## 現状コード

### `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx:13-16`
```tsx
// ❌ Feature 内部ファイルへの deep import（禁止）
import { useGetAllConsultations } from "@/features/master/api/consultations";
import { useGetAllProcedures } from "@/features/master/api/procedures";
import { useGetAllVaccinesMaster } from "@/features/master/api/vaccines-master";
import { useGetAllCheckupTypes } from "@/features/master/api/checkup-types";
```

`useGetAllVaccinesMaster` と `useGetAllCheckupTypes` は `@/features/master/index.ts` で公開されているが、`useGetAllConsultations` と `useGetAllProcedures` は**公開されていない**（index.ts に存在しない）。

### `frontend/src/features/master/index.ts`（現状）
```ts
// ❌ useGetAllConsultations, useGetAllProcedures が未公開
export { useGetAllVaccinesMaster } from "./api/vaccines-master";
export { useGetAllCheckupTypes } from "./api/checkup-types";
// useGetAllConsultations, useGetAllProcedures は export なし
```

### 比較: 正しい実装
```tsx
// ✅ 正しい: Feature Indexing 経由
import {
  useGetAllConsultations,
  useGetAllProcedures,
  useGetAllVaccinesMaster,
  useGetAllCheckupTypes
} from "@/features/master";

// ✅ master/index.ts に公開 API として追加
export { useGetAllConsultations } from "./api/consultations";
export { useGetAllProcedures } from "./api/procedures";
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx` | 13-14 | deep import（未公開 API） | 未修正 |
| `components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx` | 15-16 | deep import（公開されているが経路が不正） | 未修正 |
| `features/master/index.ts` | — | useGetAllConsultations, useGetAllProcedures が未公開 | 未修正 |

## 修正方針

### Step 1: `master/index.ts` に不足している公開 export を追加
```ts
// features/master/index.ts
export { useGetAllConsultations } from "./api/consultations";
export { useGetAllProcedures } from "./api/procedures";
export { useGetAllVaccinesMaster } from "./api/vaccines-master";
export { useGetAllCheckupTypes } from "./api/checkup-types";
```

### Step 2: `TreatmentSearchDialog.tsx:13-16` の import を修正
```tsx
// Before
import { useGetAllConsultations } from "@/features/master/api/consultations";
import { useGetAllProcedures } from "@/features/master/api/procedures";
import { useGetAllVaccinesMaster } from "@/features/master/api/vaccines-master";
import { useGetAllCheckupTypes } from "@/features/master/api/checkup-types";

// After
import {
  useGetAllConsultations,
  useGetAllProcedures,
  useGetAllVaccinesMaster,
  useGetAllCheckupTypes,
} from "@/features/master";
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Feature Indexing
> **Deep imports from features**: `@/features/xxx/components/YYY` などの深掘り import は禁止。必ず feature の `index.ts` (Feature Indexing) を経由すること。

### `.claude/CLAUDE.md` — Public API
> Feature外部（app/等）からのインポートは必ず **`index.ts`** を経由（Deep Import禁止）

## 優先度
**Medium** — `TreatmentSearchDialog` は全 feature から参照される共有コンポーネント。内部 API 変更時に影響が検知されなくなるリスクがある。

## 関連チケット
- BUG-175: auth feature の Deep Import 違反（同パターン）

## 関連ファイル
- `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx`
- `frontend/src/features/master/index.ts`
