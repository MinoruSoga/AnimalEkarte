# FE-226: Feature Indexing 違反 — shared コンポーネントが feature 内部を直接 import

## 概要

`frontend/src/components/shared/` 配下の2ファイルが、
feature の `index.ts` を経由せずに feature 内部ファイルを直接 import している。
プロジェクト規約「外部からは必ず `index.ts` を通す（Deep Import 禁止）」に違反。

## 違反箇所

### `frontend/src/components/shared/PermissionBadges/PermissionBadges.tsx:2`

```ts
// Before: feature 内部ファイルへの Deep Import
import { usePermission } from "@/features/auth/hooks/use-permission";

// After: feature index を経由
import { usePermission } from "@/features/auth";
```

### `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx:13-16`

```ts
// Before: master feature の 4 つの内部ファイルを直接 import
import { useGetAllConsultations } from "@/features/master/api/consultations";
import { useGetAllProcedures }    from "@/features/master/api/procedures";
import { useGetAllVaccinesMaster } from "@/features/master/api/vaccines-master";
import { useGetAllCheckupTypes }  from "@/features/master/api/checkup-types";

// After: master feature の index.ts から re-export して経由
import {
  useGetAllConsultations,
  useGetAllProcedures,
  useGetAllVaccinesMaster,
  useGetAllCheckupTypes,
} from "@/features/master";
```

## 対応手順

1. `frontend/src/features/auth/index.ts` に `usePermission` を追加（未 export の場合）
2. `frontend/src/features/master/index.ts` に 4 フックを追加（未 export の場合）
3. 上記 2 ファイルの import を index 経由に変更

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Public API
> Feature 外部（app/ 等）からのインポートは必ず **`index.ts`** を経由（Deep Import 禁止）

### `frontend/CLAUDE.md` — 禁止事項
> **Deep imports from features**: `@/features/xxx/components/YYY` などの深掘り import は禁止。
> 必ず feature の `index.ts` (Feature Indexing) を経由すること。

## 優先度
**Medium** — アーキテクチャ違反。現時点では動作するが、feature 内部の refactor 時に影響を受ける。
特に `TreatmentSearchDialog` は 4 つの master 内部ファイルに直接依存しており、リスクが高い。

## 関連ファイル
- `frontend/src/components/shared/PermissionBadges/PermissionBadges.tsx`
- `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx`
- `frontend/src/features/auth/index.ts`
- `frontend/src/features/master/index.ts`
