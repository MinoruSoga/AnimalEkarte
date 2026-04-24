# Frontend — React 19 / TypeScript 5.7

## Stack

React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 / shadcn/ui

## Type Safety (MANDATORY)

```typescript
// ❌ 禁止
const data: any = response.data;

// ✅ unknown + 型ガード
const parseData = (data: unknown): Owner | null => {
  if (data && typeof data === 'object' && 'id' in data) return data as Owner;
  return null;
};
```

## React 19 Patterns (MANDATORY)

```typescript
// ✅ Forms: useActionState + <form action> + SubmitButton
const [state, formAction, isPending] = useActionState(async (prevState, formData) => {
  const result = await updateOwner(id, formData);
  return { success: true, data: result };
}, null);

return (
  <form action={formAction}>
    <Input name="name" defaultValue={owner.name} />
    <SubmitButton>Save</SubmitButton>
  </form>
);

// ❌ 禁止: useState + manual isLoading
const [isLoading, setIsLoading] = useState(false);

// ✅ Components: function declaration (no FC, no forwardRef)
export function OwnerCard({ owner, ref }: { owner: Owner; ref?: React.Ref<HTMLDivElement> }) {}

// ❌ 禁止
export const OwnerCard: FC<Props> = () => {};
export const OwnerCard = forwardRef(...);
```

## Conditional Render (MANDATORY)

```typescript
// ✅
{isLoading ? <Spinner /> : null}

// ❌ && は数値 0 / 空文字をレンダリングする
{isLoading && <Spinner />}
```

## Design Tokens (MANDATORY)

```typescript
import { C, STYLE } from '@/lib/design-tokens';

// ✅
<div className={STYLE.FLEX_BETWEEN} style={{ color: C.TEXT_MAIN }}>

// ❌ hex 直書き禁止
<div style={{ color: '#37352F' }}>
```

## Feature Indexing (MANDATORY)

```typescript
// ✅ index.ts 経由
import { OwnerCard, useOwners } from '@/features/owners';

// ❌ deep import 禁止
import { OwnerCard } from '@/features/owners/components/OwnerCard';
```

詳細は `src/features/CLAUDE.md` を参照。

## Prohibited Commands (must NOT auto-execute)

```bash
docker compose exec frontend pnpm lint
docker compose exec frontend pnpm test:run
docker compose exec frontend pnpm build
docker compose exec frontend pnpm type-check
docker compose exec frontend pnpm install
```
