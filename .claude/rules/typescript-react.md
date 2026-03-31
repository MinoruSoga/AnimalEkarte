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

### 3. React 19 hooks パターン（Actions）

データ更新（Mutation）は **React 19 Action** パターンを標準とする。

```typescript
// ✅ MANDATE: useActionState + <form action>
const [state, formAction, isPending] = useActionState(async (prevState, formData) => {
  try {
    const result = await updateOwner(id, formData);
    return { success: true, data: result };
  } catch (error) {
    handleApiError(error, "更新");
    return { success: false, error };
  }
}, null);

return (
  <form action={formAction}>
    <Input name="name" defaultValue={owner.name} />
    {/* ✅ MANDATE: すべてのフォームで SubmitButton を使用 */}
    <SubmitButton>保存</SubmitButton>
  </form>
);

// ❌ 禁止: useState + onSubmit での手動ローディング管理
const [isLoading, setIsLoading] = useState(false);
const handleSubmit = async (e) => {
  setIsLoading(true); // 禁止
  await api();
  setIsLoading(false);
};
```

### 4. デザイントークン（デザインシステム）

スタイリングには必ず `src/lib/design-tokens.ts` の定数を使用する。

```typescript
import { C, STYLE } from '@/lib/design-tokens';

// ✅ 正しい: トークンを使用
<div className={cn(STYLE.FLEX_BETWEEN, "p-4")} style={{ color: C.TEXT_MAIN }}>

// ❌ 禁止: Hexカラーの直接指定
<div style={{ color: '#37352F' }}>
```

### 5. 命名規則

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

// コンポーネントファイル(.tsx): PascalCase
// OwnerCard.tsx, OwnerForm.tsx
// 非コンポーネントファイル(.ts): kebab-case
// use-owner-form.ts, get-owners.ts

// Interface/Type: PascalCase
interface Owner {}
type OwnerStatus = 'active' | 'inactive';
```

### 6. 条件レンダー（`&&` 禁止）

```typescript
// ❌ 禁止: &&（0や空文字が漏れる）
{isLoading && <div>Loading...</div>}
{items.length && <List items={items} />}

// ✅ 正しい: 三項演算子
{isLoading ? <div>Loading...</div> : null}
{items.length > 0 ? <List items={items} /> : null}
```

### 7. import順序・Feature Indexing推奨

外部モジュールから feature を利用する場合、必ず `index.ts` を経由する。

```typescript
// ✅ 正しい: Feature Indexing 経由（推奨）
import { OwnerCard, useOwners } from '@/features/owners';

// ❌ 禁止: Feature 内部ファイルへの深掘り import
import { OwnerCard } from '@/features/owners/components/OwnerCard';
import { useOwners } from '@/features/owners/hooks/use-owners';

// Import順序: React → 外部lib → 内部共通 (@/) → Feature-internal → styles
import { useState } from 'react';
import axios from 'axios';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { C } from '@/lib/design-tokens';
import { getOwners } from '../api/get-owners'; // 同一feature内
import type { Owner } from '@/types';
```

### 8. フォーム管理パターン

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

### 9. memo() で最適化

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
- [ ] スタイリング = デザイントークン（`C`, `STYLE`）使用。Hexカラー禁止。
- [ ] 条件レンダー = 三項演算子（`&&` 禁止）
- [ ] 外部からの Feature 利用 = `index.ts` 経由（Feature Indexing）
- [ ] フォーム = `useActionState` + `SubmitButton` 使用
- [ ] useCallback でハンドラ安定化
- [ ] 複雑フォーム = useTransition（useState + setIsPending 禁止）
- [ ] 検索フィルタ = useDeferredValue
- [ ] API = useQuery/useMutation
- [ ] Form = useXxxForm hook
