---
description: TypeScript/React 19 coding standards (type safety, React 19 patterns)
alwaysApply: true
globs: ["frontend/src/**/*.{ts,tsx}"]
---

# TypeScript / React 19 Rules

React 19 + TypeScript 6.0 project standards.

## Core Rules

### 1. Type Definition (any prohibited, type safety first)

```typescript
// ❌ Prohibited: any
const handleChange = (e: any) => {};
const data: any = response.data;

// ✅ Correct
interface ChangeEvent {
  target: { name: string; value: string };
}

const handleChange = (e: ChangeEvent) => {};

// unknown + type guard
const parseData = (data: unknown): Owner | null => {
  if (data && typeof data === 'object' && 'id' in data) {
    return data as Owner;
  }
  return null;
};
```

### 2. Component Definition (React 19)

```typescript
// ✅ React 19 - Function declaration + Props type
interface OwnerCardProps {
  owner: Owner;
  onSelect?: (id: number) => void;
  ref?: React.Ref<HTMLDivElement>;
}

export function OwnerCard({ owner, onSelect, ref }: OwnerCardProps) {
  return <div ref={ref} onClick={() => onSelect?.(owner.id)}>{owner.name}</div>;
}

// ❌ Prohibited: FC type (unnecessary in React 19)
export const OwnerCard: FC<Props> = () => {};

// ❌ Prohibited: forwardRef (React 19 uses ref as prop)
export const OwnerCard = forwardRef(({ owner }, ref) => <div ref={ref} />);
```

### 3. React 19 Hooks Pattern (Actions)

Use **React 19 Action** pattern for data mutations (standard).

```typescript
// ✅ MANDATE: useActionState + <form action>
const [state, formAction, isPending] = useActionState(async (prevState, formData) => {
  try {
    const result = await updateOwner(id, formData);
    return { success: true, data: result };
  } catch (error) {
    handleApiError(error, "update");
    return { success: false, error };
  }
}, null);

return (
  <form action={formAction}>
    <Input name="name" defaultValue={owner.name} />
    {/* ✅ MANDATE: Use SubmitButton for all forms */}
    <SubmitButton>Save</SubmitButton>
  </form>
);

// ❌ Prohibited: useState + onSubmit manual loading management
const [isLoading, setIsLoading] = useState(false);
const handleSubmit = async (e) => {
  setIsLoading(true); // prohibited
  await api();
  setIsLoading(false);
};
```

### 4. Design Tokens

Always use constants from `src/lib/design-tokens.ts` for styling.

```typescript
import { C, STYLE } from '@/lib/design-tokens';

// ✅ Correct: Use tokens
<div className={cn(STYLE.FLEX_BETWEEN, "p-4")} style={{ color: C.TEXT_MAIN }}>

// ❌ Prohibited: Direct hex color specification
<div style={{ color: '#37352F' }}>
```

### 5. Naming Conventions

```typescript
// Components: PascalCase
export function OwnerCard() {}
export function OwnerForm() {}

// Functions/variables: camelCase
const getOwners = async () => {};
const isLoading = false;

// Constants: UPPER_SNAKE_CASE
const API_BASE_URL = 'http://localhost:8080';
const MAX_OWNERS_PER_PAGE = 20;

// API Hook: useGet/useCreate/useUpdate/useDelete + entity name
const useGetOwners = () => useQuery(...);
const useGetOwner = (id) => useQuery(...);
const useCreateOwner = () => useMutation(...);

// Form Hook: use + entity name + Form
const useOwnerForm = (initialValues?: Owner) => {
  const [formData, setFormData] = useState(initialValues);
  return { formData, setFormData };
};

// Component file (.tsx): PascalCase
// OwnerCard.tsx, OwnerForm.tsx
// Non-component file (.ts): kebab-case
// use-owner-form.ts, get-owners.ts

// Interface/Type: PascalCase
interface Owner {}
type OwnerStatus = 'active' | 'inactive';
```

### 6. Conditional Render (`&&` prohibited)

```typescript
// ❌ Prohibited: && (zero or empty string leak)
{isLoading && <div>Loading...</div>}
{items.length && <List items={items} />}

// ✅ Correct: Ternary
{isLoading ? <div>Loading...</div> : null}
{items.length > 0 ? <List items={items} /> : null}
```

### 7. Import Order / Feature Indexing

Always use `index.ts` when importing from external features.

```typescript
// ✅ Correct: Feature Indexing (recommended)
import { OwnerCard, useOwners } from '@/features/owners';

// ❌ Prohibited: Deep import from feature internals
import { OwnerCard } from '@/features/owners/components/OwnerCard';
import { useOwners } from '@/features/owners/hooks/use-owners';

// Import order: React → external lib → internal shared (@/) → Feature-internal → styles
import { useState } from 'react';
import axios from 'axios';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { C } from '@/lib/design-tokens';
import { getOwners } from '../api/get-owners'; // same feature only
import type { Owner } from '@/types';
```

### 8. Form Management Pattern

```typescript
// useCallback for handler stabilization (memo prerequisite)
const handleInputChange = useCallback((field: string, value: string) => {
  setFormData(prev => ({ ...prev, [field]: value }));
}, []);

// lazy init for deferred initial state
const [formData, setFormData] = useState<Owner>(() => {
  return initialOwner ?? defaultOwner;
});

// useDeferredValue for search filter delay
const deferredSearchTerm = useDeferredValue(searchTerm);
const filteredOwners = useMemo(
  () => owners.filter(o => o.name.includes(deferredSearchTerm)),
  [owners, deferredSearchTerm]
);
```

### 9. memo() Optimization

```typescript
// ✅ Split large forms into sections
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

// Memoize section children
const OwnerInfoSection = memo(function OwnerInfoSection({ data, onChange }: Props) {
  return <div>...</div>;
});
```

## Checklist

- [ ] No `any` types (use unknown + type guard)
- [ ] No `FC` type, forwardRef (React 19 patterns)
- [ ] Components = PascalCase
- [ ] Hook = camelCase
- [ ] Functions = camelCase
- [ ] Constants = UPPER_SNAKE_CASE
- [ ] Styling = Design tokens (`C`, `STYLE`). No hex colors.
- [ ] Conditional render = Ternary (NOT `&&`)
- [ ] Feature imports = via `index.ts` (Feature Indexing)
- [ ] Forms = `useActionState` + `SubmitButton`
- [ ] useCallback for handler stabilization
- [ ] Complex forms = useTransition (NOT useState + setIsPending)
- [ ] Search filter = useDeferredValue
- [ ] API = useQuery/useMutation
- [ ] Forms = useXxxForm hook
