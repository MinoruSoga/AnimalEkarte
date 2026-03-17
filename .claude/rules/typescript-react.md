---
description: TypeScript/React 19コーディング規約（型安全性、React 19パターン）
alwaysApply: true
globs: ["frontend/src/**/*.{ts,tsx}"]
---

# TypeScript / React 19 Rules

React 19 + TypeScript 5.7 プロジェクト標準ルール。

## 核心ルール

### 1. 型定義（any禁止、型安全性優先）

```typescript
// ❌ 禁止: any
const handleChange = (e: any) => {};
const data: any = response.data;

// ✅ 正しい
interface ChangeEvent {
  target: { name: string; value: string };
}

const handleChange = (e: ChangeEvent) => {};

// unknown + 型ガード
const parseData = (data: unknown): Owner | null => {
  if (data && typeof data === 'object' && 'id' in data) {
    return data as Owner;
  }
  return null;
};
```

### 2. コンポーネント定義（React 19）

```typescript
// ✅ React 19 - 関数宣言 + Props型
interface OwnerCardProps {
  owner: Owner;
  onSelect?: (id: number) => void;
  ref?: React.Ref<HTMLDivElement>;
}

export function OwnerCard({ owner, onSelect, ref }: OwnerCardProps) {
  return <div ref={ref} onClick={() => onSelect?.(owner.id)}>{owner.name}</div>;
}

// ❌ 禁止: FC型（React 19では不要）
export const OwnerCard: FC<Props> = () => {};

// ❌ 禁止: forwardRef（React 19ではref as prop）
export const OwnerCard = forwardRef(({ owner }, ref) => <div ref={ref} />);
```

### 3. React 19 hooks パターン

```typescript
// ★ useTransition: 複雑フォーム pending 管理（プロジェクト標準）
const [isSavePending, startSaveTransition] = useTransition();

const handleSave = () => {
  startSaveTransition(async () => {
    await saveOwner(formData);
  });
};

// useActionState: シンプルな非制御フォーム（FormData使用時のみ）
const [state, formAction, isPending] = useActionState(submitAction, null);

// useOptimistic: 楽観的UI更新
const [optimisticOwners, addOptimisticOwner] = useOptimistic(owners, (state, newOwner) => [
  ...state,
  newOwner,
]);

// use(): Promise/Context直接読み取り
const ownerData = use(ownerPromise);

// useFormStatus: フォーム子コンポーネント内
const { pending } = useFormStatus();
```

### 4. 命名規則

```typescript
// コンポーネント: PascalCase
export function OwnerCard() {}
export function OwnerForm() {}

// 関数・変数: camelCase
const getOwners = async () => {};
const isLoading = false;

// 定数: UPPER_SNAKE_CASE
const API_BASE_URL = 'http://localhost:8080';
const MAX_OWNERS_PER_PAGE = 20;

// API Hook: useGet/useCreate/useUpdate/useDelete + エンティティ名
const useGetOwners = () => useQuery(...);
const useGetOwner = (id) => useQuery(...);
const useCreateOwner = () => useMutation(...);

// Form Hook: use + エンティティ名 + Form
const useOwnerForm = (initialValues?: Owner) => {
  const [formData, setFormData] = useState(initialValues);
  return { formData, setFormData };
};

// ファイル: kebab-case
// owner-card.tsx, owner-form.tsx, use-owner-form.ts

// Interface/Type: PascalCase
interface Owner {}
type OwnerStatus = 'active' | 'inactive';
```

### 5. 条件レンダー（`&&` 禁止）

```typescript
// ❌ 禁止: &&（0や空文字が漏れる）
{isLoading && <div>Loading...</div>}
{items.length && <List items={items} />}

// ✅ 正しい: 三項演算子
{isLoading ? <div>Loading...</div> : null}
{items.length > 0 ? <List items={items} /> : null}
```

### 6. import順序・barrel index禁止

```typescript
// ❌ 禁止: barrel index経由import
import { OwnerCard } from '@/features/owners';
import { usePets } from '@/features/pets';

// ✅ 正しい: 直接ファイル import（tree-shaking優先）
import { OwnerCard } from '@/features/owners/components/OwnerCard';
import { usePets } from '@/features/pets/hooks/use-pets';
import { Owner } from '@/types';
import { cn } from '@/lib/utils';

// Import順序: React → 外部lib → 内部modules → types → styles
import { useState } from 'react';
import axios from 'axios';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { getOwners } from '../api/get-owners';
import type { Owner } from '@/types';
import styles from './owner-card.module.css';
```

### 7. フォーム管理パターン

```typescript
// useCallback でハンドラ安定化（memo の前提条件）
const handleInputChange = useCallback((field: string, value: string) => {
  setFormData(prev => ({ ...prev, [field]: value }));
}, []);

// lazy init で初期状態を遅延初期化
const [formData, setFormData] = useState<Owner>(() => {
  return initialOwner ?? defaultOwner;
});

// useDeferredValue で検索フィルタ遅延
const deferredSearchTerm = useDeferredValue(searchTerm);
const filteredOwners = useMemo(
  () => owners.filter(o => o.name.includes(deferredSearchTerm)),
  [owners, deferredSearchTerm]
);
```

### 8. memo() で最適化

```typescript
// ✅ 大型フォームをセクションに分割
export const OwnerForm = memo(function OwnerForm({ owner, onSave }: Props) {
  const [formData, setFormData] = useState(owner);
  const handleChange = useCallback(..., []);

  return (
    <form>
      <OwnerInfoSection data={formData} onChange={handleChange} />
      <PetTableSection pets={formData.pets} />
    </form>
  );
});

// memo() でセクション子をメモ化
const OwnerInfoSection = memo(function OwnerInfoSection({ data, onChange }: Props) {
  return <div>...</div>;
});
```

## チェックリスト

- [ ] `any` 型なし（unknown + 型ガード）
- [ ] `FC` 型、forwardRef なし（React 19パターン）
- [ ] コンポーネント = PascalCase
- [ ] Hook = camelCase
- [ ] 関数 = camelCase
- [ ] 定数 = UPPER_SNAKE_CASE
- [ ] 条件レンダー = 三項演算子（`&&` 禁止）
- [ ] import は直接ファイル（barrel禁止）
- [ ] useCallback でハンドラ安定化
- [ ] 複雑フォーム = useTransition（useState + setIsPending 禁止）
- [ ] 検索フィルタ = useDeferredValue
- [ ] API = useQuery/useMutation
- [ ] Form = useXxxForm hook
