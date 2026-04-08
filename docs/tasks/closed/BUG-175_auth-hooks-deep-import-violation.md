# BUG-175: `@/features/auth/hooks/` への Deep Import 違反（26ファイル）

## 概要

26 ファイルが `@/features/auth/hooks/use-permission` または `@/features/auth/hooks/use-auth` を直接 import している。プロジェクト規約では Feature 外部からの参照は必ず `index.ts` 経由（`@/features/auth`）が必須。`use-permission` は `auth/index.ts` に既にエクスポートされているが、多くのファイルが内部ファイルを直接参照しており、Feature Indexing 規約に違反している。

## 再現手順

1. 以下のファイルを任意に開く（例: `features/checkups/routes/CheckupsList.tsx`）
2. import 文を確認する
3. **結果**: `@/features/auth/hooks/use-permission` という深掘り import が使われている

## 期待する動作

```tsx
// ✅ 正しい: index.ts 経由
import { usePermission } from "@/features/auth";

// ❌ 違反: hooks ファイルへの直接参照
import { usePermission } from "@/features/auth/hooks/use-permission";
```

## 現状コード

### Deep Import を使用している 26 ファイル一覧

| ファイルパス | 行番号 | 違反 import |
|---|---|---|
| `features/checkups/routes/CheckupsList.tsx` | 23 | `@/features/auth/hooks/use-permission` |
| `features/reception/routes/Reception.tsx` | 22 | `@/features/auth/hooks/use-permission` |
| `features/medical-records/components/ExaminationFilter.tsx` | 13 | `@/features/auth/hooks/use-permission` |
| `features/medical-records/components/InterviewTreatmentPolicy.tsx` | 10 | `@/features/auth/hooks/use-permission` |
| `features/medical-records/components/MedicalRecordEstimate.tsx` | 11 | `@/features/auth/hooks/use-permission` |
| `features/medical-records/components/DiagnosisHeader.tsx` | 5 | `@/features/auth/hooks/use-permission` |
| `features/medical-records/components/InterviewChiefComplaint.tsx` | 15 | `@/features/auth/hooks/use-permission` |
| `features/master/components/MasterCRUDPage.tsx` | 4 | `@/features/auth/hooks/use-permission` |
| `features/master/components/MasterListPage.tsx` | 15 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/TreatmentPlanMaster.tsx` | 45 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/ServiceTypeSettings.tsx` | 21 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/MerchandiseItemSettings.tsx` | 30 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/PermissionGroupSettings.tsx` | 10 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/AnimalSpeciesSettings.tsx` | 18 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/MasterSettingsIndex.tsx` | 14 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/MasterSettingsIndex.tsx` | 15 | `@/features/auth/hooks/use-auth` |
| `features/master/routes/DiagnosisSettings.tsx` | 48 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/MedicineSettings.tsx` | 48 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/CageSettings.tsx` | 24 | `@/features/auth/hooks/use-permission` |
| `features/master/routes/CompanySettings.tsx` | 17 | `@/features/auth/hooks/use-permission` |

**※ hospitalization, accounting, estimates, owners は正しく `@/features/auth` を使用しており違反なし。**

## 影響範囲

| 対象 | 件数 | 状態 |
|---|---|---|
| Deep import 違反ファイル | 20ファイル | 未修正 |
| `checkups`, `reception`, `medical-records`, `master` feature | 全体的に違反 | 未修正 |

## 修正方針

### 全 20 ファイルで一括置換

```bash
# 確認コマンド
grep -r "from \"@/features/auth/hooks/" frontend/src/features/

# 置換: @/features/auth/hooks/use-permission → @/features/auth
# 置換: @/features/auth/hooks/use-auth → @/features/auth
```

各ファイルの import を以下に変更:
```tsx
// Before
import { usePermission } from "@/features/auth/hooks/use-permission";

// After
import { usePermission } from "@/features/auth";
```

`@/features/auth/index.ts` に `usePermission` と `useAuth` が既にエクスポートされていることを確認済みのため、動作変更なし。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Feature Indexing
> **PROHIBITED**: `@/features/xxx/components/YYY` などの深掘り import は禁止。必ず feature の `index.ts` (Feature Indexing) を経由すること。

### `.claude/CLAUDE.md` — Public API
> Feature外部（app/等）からのインポートは必ず **`index.ts`** を経由（Deep Import禁止）

### プロジェクト内参照実装
- `features/hospitalization/routes/HospitalizationForm.tsx:15` — `import { useAuth, usePermission } from "@/features/auth";` ✅ 正しい実装

## 優先度
**Medium** — 機能的な問題はないが、`auth` feature の内部構造変更（ファイル名変更・移動）が行われると 20 ファイルすべてで import エラーが発生する。一括置換で解決できるため早期対応を推奨。

## 関連チケット
なし

## 関連ファイル
- `frontend/src/features/auth/index.ts` — 公開 API 定義
- 上記 20 ファイル全て
