# Features — Feature Indexing Pattern

## ディレクトリ構造 (MANDATORY)

各 feature は `index.ts`（公開 API）を必須とし、`api/` `components/` `hooks/` `routes/` `types/` のサブディレクトリに分ける（実例は `ls features/owners` 参照）。`loaders.ts`（react-router ローダー）は `owners` のみが使う例外的パターン。

## index.ts の責務

```typescript
// ✅ features/owners/index.ts — 公開するものだけ export
export { OwnerCard } from "./components/OwnerCard";
export { OwnerForm } from "./components/OwnerForm";
export { useOwners } from "./hooks/use-owners";
export type { Owner } from "./types";

// 内部実装（api/hooks の詳細）は export しない
```

## Import ルール

```typescript
// ✅ 外部から feature を使う場合は必ず index.ts 経由
import { OwnerCard, useOwners } from "@/features/owners";

// ✅ feature 内部から同 feature 内の別ファイルは相対 import OK
import { getOwners } from "../api/get-owners";

// ❌ deep import 禁止（外部から feature 内部への直接アクセス）
import { OwnerCard } from "@/features/owners/components/OwnerCard";
import { useOwners } from "@/features/owners/hooks/use-owners";
```

## 命名規則

| 種別                   | 規則                                          | 例                                 |
| ---------------------- | --------------------------------------------- | ---------------------------------- |
| コンポーネントファイル | PascalCase.tsx                                | `OwnerCard.tsx`                    |
| フック・ユーティリティ | kebab-case.ts                                 | `use-owner-form.ts`                |
| API 関数               | kebab-case.ts                                 | `get-owners.ts`, `create-owner.ts` |
| API フック             | `useGet/useCreate/useUpdate/useDelete` + 名詞 | `useGetOwners`, `useCreateOwner`   |
| フォームフック         | `use` + 名詞 + `Form`                         | `useOwnerForm`                     |

## 命名例外 (FE-RC-055・2026-09-03)

`useGet` 命名は「バックエンドを直接 fetch する query hook」に適用する。以下は対象外:

- **派生値 facade**: `useAnimalSpecies`（`src/hooks/use-animal-species.ts`）・`useClinicTaxRates`（`src/hooks/use-clinic-tax-rates.ts`）・`useCurrentClinicName`（`src/hooks/use-current-clinic-name.ts`）は、内部で `useQuery`/`useAuth` を呼びラベル付与・フォールバック計算・整形などの派生値を返す facade であり、生の fetch wrapper ではない。`useGet` は「取得のみ」を意味するため、派生値を返す facade には付けない。内部の生 fetch 部分（例: `useAnimalSpecies` 内の `useGetAnimalSpecies`）には `useGet` を付けてよい。
- **mutation 動詞**: `useCreate/useUpdate/useDelete` 以外にも業務動詞が固有な mutation hook は動詞そのものを使ってよい（例: `usePutLabDeviceWait`、`useClearLabDeviceWait`、`useReceiveLabDeviceFrames`、`useAttachLabDeviceJob`、`useDetachLabDeviceJob`、`useUnconfirmExamination`、`useReplacePetSubOwners`）。`useCreate/useUpdate/useDelete` に強引に当てはめて業務語彙を失わない。
