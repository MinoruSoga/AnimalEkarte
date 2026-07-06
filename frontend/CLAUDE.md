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

## Scoped Test Verification (MANDATORY)

- `docker compose exec frontend pnpm test:run -- <path>` は罠 — `--` 以降のパスがスコープ指定として渡らず全件実行になる。スコープ限定したい場合は必ず `docker compose exec frontend npx vitest run <path>` を使うこと（`.claude/skills/scoped-verification-gates/SKILL.md` と整合）。
- `PageLayout` に `resource` prop を渡す route の render test では `PermissionBadges → usePermission → useAuth` が呼ばれるため、`vi.mock("@/hooks/use-auth", () => ({ useAuth: () => ({ ... hasPermission: () => true }) }))` が必須（`AuthProvider` 無しだと `useAuth must be used within an AuthProvider` で全滅する）。正例: `frontend/src/features/cash-register/routes/CashRegisterHistoryPage.test.tsx`。
- 新規リーフルートを追加したら `docs/UI_DESIGN_COMPLIANCE.md` §2 のページ表を同じコミットで更新すること（C1/C3/C5 は `pnpm design-audit` で機械検証されるが、C2/C4 の `PageLayout` 使用有無は手動追跡のみのため更新漏れが唯一の防御線）。
