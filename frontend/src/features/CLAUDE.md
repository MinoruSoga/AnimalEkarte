# Features — Feature Indexing Pattern

## ディレクトリ構造 (MANDATORY)

```
features/
└── owners/
    ├── index.ts          ← 公開 API (必須)
    ├── api/
    │   ├── get-owners.ts
    │   └── create-owner.ts
    ├── components/
    │   ├── OwnerCard.tsx
    │   └── OwnerForm.tsx
    ├── hooks/
    │   └── use-owners.ts
    ├── routes/
    │   └── OwnerForm.tsx   ← ルートコンポーネント（react-router の要素）
    ├── loaders.ts          ← react-router ローダー（owners のみが使う例外的パターン）
    └── types/
        └── index.ts
```

## index.ts の責務

```typescript
// ✅ features/owners/index.ts — 公開するものだけ export
export { OwnerCard } from './components/OwnerCard';
export { OwnerForm } from './components/OwnerForm';
export { useOwners } from './hooks/use-owners';
export type { Owner } from './types';

// 内部実装（api/hooks の詳細）は export しない
```

## Import ルール

```typescript
// ✅ 外部から feature を使う場合は必ず index.ts 経由
import { OwnerCard, useOwners } from '@/features/owners';

// ✅ feature 内部から同 feature 内の別ファイルは相対 import OK
import { getOwners } from '../api/get-owners';

// ❌ deep import 禁止（外部から feature 内部への直接アクセス）
import { OwnerCard } from '@/features/owners/components/OwnerCard';
import { useOwners } from '@/features/owners/hooks/use-owners';
```

## 命名規則

| 種別 | 規則 | 例 |
|------|------|---|
| コンポーネントファイル | PascalCase.tsx | `OwnerCard.tsx` |
| フック・ユーティリティ | kebab-case.ts | `use-owner-form.ts` |
| API 関数 | kebab-case.ts | `get-owners.ts`, `create-owner.ts` |
| API フック | `useGet/useCreate/useUpdate/useDelete` + 名詞 | `useGetOwners`, `useCreateOwner` |
| フォームフック | `use` + 名詞 + `Form` | `useOwnerForm` |
