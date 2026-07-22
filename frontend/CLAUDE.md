# Frontend — React 19 / TypeScript 6.0

## Stack

React 19 / TypeScript 6.0 / Vite 8 / Tailwind CSS 4 / shadcn/ui

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

`lib/design-tokens.ts`（805行）は 800行ファイルサイズ規約の documented exception。トークンテーブルの分割は可読性を下げるだけで得られる便益がないため、分割しない（FE7-4 判断・2026-07-18）。

## Feature Indexing (MANDATORY)

```typescript
// ✅ index.ts 経由
import { OwnerCard, useOwners } from '@/features/owners';

// ❌ deep import 禁止
import { OwnerCard } from '@/features/owners/components/OwnerCard';
```

詳細は `src/features/CLAUDE.md` を参照。

## Shared Helper 配置 (MANDATORY)

```typescript
// ✅ 共有ヘルパは lib/ に一元化
import { formatCurrency } from '@/lib/format/number';

// ❌ utils/ ディレクトリの新設禁止
import { formatCurrency } from '@/utils/format';
```

共有ヘルパの置き場所は `lib/` の1箇所のみ。`utils/` ディレクトリの新規作成を禁止する（FE7-1 で `utils/` を `lib/` へ統合済み・2026-07-18）。

## Shared Constants 配置 (MANDATORY)

```typescript
// ✅ 共有定数は src/constants/ に一元化
import { RECEPTION_STATUS_COLORS } from '@/constants/status-colors';

// ❌ 他ディレクトリ（lib/、feature 内など）での新規共有定数ファイル作成禁止
```

共有定数の置き場所は `src/constants/` の1箇所のみ（FE7-2 で `utils/constants/` を統合済み・2026-07-18）。

## shared-liff 配置 (Decision Record)

`liff` / `line-reserve` は frontend の `src/` ツリーを `@/shared-liff/...` alias で参照する現行構造を維持する（FE7-3 Option B 採用・見落としではなく意図的判断・2026-07-18）。

- 判断理由: 動作・tree-shaking とも問題なし。`frontend/shared-liff/` への昇格（Option A・vite/tsconfig alias 3箇所更新）はコストに見合う便益が現時点でない（②削除原則 — 動くものを動かさない）。
- 再検討条件: `liff` / `line-reserve` に具体的な機能拡張計画が生まれた時のみ Option A を再検討する。

## Import 境界 Lint (ESLint no-restricted-imports) (MANDATORY)

`eslint.config.js` に以下3種の境界ルールが入っている。違反は ESLint で機械検出される（過去に cross-feature import 38件が人力レビューをすり抜けていた実績を受けた機械ガード化、第2期監査。FE7-0・2026-07-18）。

1. **deep import 禁止**（全域）: feature の外から `@/features/<name>/...` への直接 import を禁止。外からは `@/features/<name>`（index.ts）経由、feature 内部も相対 import を使うこと。
2. **層逆転禁止**: `src/components/`, `src/hooks/`, `src/lib/` から `@/features` への import を禁止（features は components/hooks/lib に依存してよいが逆方向は禁止）。
3. **アプリ境界**: `liff/src/`, `line-reserve/src/` から `@/features` への import を禁止。

## Prohibited Commands (must NOT auto-execute)

```bash
docker compose exec frontend pnpm lint
docker compose exec frontend pnpm test:run
docker compose exec frontend pnpm build
docker compose exec frontend pnpm type-check
docker compose exec frontend pnpm install
```

## Test 配置 (MANDATORY)

- テストはテスト対象ファイルに隣接配置する（`Foo.tsx` → `Foo.test.tsx`、同一ディレクトリ）。
- `__tests__/` ディレクトリは新設禁止（FE5-23 で全廃済み）。

## Scoped Test Verification (MANDATORY)

- `docker compose exec frontend pnpm test:run -- <path>` は罠 — `--` 以降のパスがスコープ指定として渡らず全件実行になる。スコープ限定したい場合は必ず `docker compose exec frontend npx vitest run <path>` を使うこと（`.claude/skills/scoped-verification-gates/SKILL.md` と整合）。
- `tsc --noEmit`（`pnpm type-check`）は `tsconfig.json` の `exclude` によりテストファイルを検証しない。import 改名・移行の検証罠（3アプリ全域grep・vitest実証手順）は `.claude/skills/scoped-verification-gates/SKILL.md`「検証の罠」節を参照。
- `PageLayout` に `resource` prop を渡す route の render test では `PermissionBadges → usePermission → useAuth` が呼ばれるため、`vi.mock("@/hooks/use-auth", () => ({ useAuth: () => ({ ... hasPermission: () => true }) }))` が必須（`AuthProvider` 無しだと `useAuth must be used within an AuthProvider` で全滅する）。正例: `frontend/src/features/cash-register/routes/CashRegisterHistoryPage.test.tsx`。
- 新規リーフルートを追加したら `docs/spec/ui-design-compliance.md` §2 のページ表を同じコミットで更新すること。C1/C3/C5/C6b/C7〜C19 は `pnpm design-audit`（make ci 配線済み）で機械検証される。C8 は `src/features/*/routes/*.tsx` の PageLayout / Master*Page /相対パス完全一致 allowlist（14件 = 独自page 9 + helper 5）を検査する — 正当な非 PageLayout ルートは `design-system-audit.mjs` の対応する allowlist へ同一コミットで追記すること。C18 は `TableHead` / `TableCell` 呼び出し側による非仕様 typography・vertical padding の再上書き、C19 は `DataTableRow` / `SortableDataTableRow` / `TableRow` / raw `tr` の行全体 `onClick` を禁止する。遷移・表示・選択はcell内のnative link/button、並べ替えは44px drag handleを使う。空stateで追加vertical paddingが必要な場合だけ `data-empty-state` + `colSpan` + `text-center` を明示する。

## DESIGN.md / 製品デザイン適用範囲（MANDATORY・FE11 2026-07-21）

- **色の正本**は `docs/spec/design-system.md`。brand **`#0075DE`**、臨床 semantic 色、業務 status 色、nav canvas-soft は製品判断として維持する。旧 teal `#038B94` は legacy（audit C1 で再導入禁止）。
- **タイポ / 形状 / 余白 / エレベーション / コンポーネント寸法の正本**は `DESIGN.md`。hairline `#E6E6E6`、caption 14px / eyebrow 12px、radius 最小 4px、`button-primary` = pill、テーブルヘッダ = `ex-data-table-cell` 様式（`STYLE.tableHeader*`）へ字義で従う。
- brand は CTA・リンク・active/focus に限定し、装飾へ流用しない。semantic 色を構造用途へ流用しない。
