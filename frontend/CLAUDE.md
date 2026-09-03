# Frontend — React 19 / TypeScript 6.0

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
import { formatCurrency } from '@/TestUtils/format';
```

共有ヘルパの置き場所は `lib/` の1箇所のみ。`utils/` ディレクトリの新規作成を禁止する（FE7-1 で `utils/` を `lib/` へ統合済み・2026-07-18）。

## Shared Constants 配置 (MANDATORY)

```typescript
// ✅ 共有定数は src/constants/ に一元化
import { RECEPTION_STATUS_COLORS } from '@/constants/status-colors';

// ❌ 他ディレクトリ（lib/、feature 内など）での新規共有定数ファイル作成禁止
```

共有定数の置き場所は `src/constants/` の1箇所のみ（FE7-2 で `utils/constants/` を統合済み・2026-07-18）。

## Pet selector の表示範囲と選択可否 (MANDATORY)

- Accounting / Checkup / Examination / Hospitalization / Medical Record / Trimming / Vaccination の7つの直接記録入力 selector は、本人同定と死亡 sentinel の表示のため、共有 `usePetSelectionPage` から `includeDeceased: true` を要求する。行が見えることは操作可能を意味せず、選択には `pet.status === "生存"` を必須とする。
- `ReservationFormModal` の将来予約 selector は、死亡個体が将来受診候補にならないため、`includeDeceased` を意図的に指定しない。

## shared-liff 配置 (Decision Record)

`liff` / `line-reserve` は frontend の `src/` ツリーを `@/shared-liff/...` alias で参照する現行構造を維持する（FE7-3 Option B 採用・見落としではなく意図的判断・2026-07-18）。

- 判断理由: 動作・tree-shaking とも問題なし。`frontend/shared-liff/` への昇格（Option A・vite/tsconfig alias 3箇所更新）はコストに見合う便益が現時点でない（②削除原則 — 動くものを動かさない）。
- 再検討条件: `liff` / `line-reserve` に具体的な機能拡張計画が生まれた時のみ Option A を再検討する。

## Import 境界 Lint (ESLint no-restricted-imports) (MANDATORY)

`eslint.config.js` に以下4種の境界ルールが入っている。違反は ESLint で機械検出される（過去に cross-feature import 38件が人力レビューをすり抜けていた実績を受けた機械ガード化、第2期監査。FE7-0・2026-07-18）。

1. **deep import 禁止**（全域）: feature の外から `@/features/<name>/...` への直接 import を禁止。外からは `@/features/<name>`（index.ts）経由、feature 内部も相対 import を使うこと。
2. **層逆転禁止**: `src/components/`, `src/hooks/`, `src/lib/` から `@/features` への import を禁止（features は components/hooks/lib に依存してよいが逆方向は禁止）。相対パス（`../features/...`）も同様に禁止。
3. **アプリ境界**: `liff/src/`, `line-reserve/src/` から `@/features` への import を禁止。
4. **feature 間 import 禁止**（`src/features/<a>/**` → `@/features/<b>`）: barrel（index.ts）経由でも feature 間の直接 import を禁止する（FE-RC-015/060・2026-09-03。CODING_RULES.md §1.2 との不整合を解消し ESLint 側も全面禁止に統一）。cross-feature の値が必要な場合は `components/shared`・`src/hooks`・`src/lib` へ昇格するか、`app/pages/` の合成層で props 注入すること。唯一の例外は `src/features/owner-report/hooks/use-owner-clinical-briefing-data.ts`（`@/features/medical-records` の `useGetMedicalRecords` を読み取り専用で参照。`transformMedicalRecord` は診療録ドメインの正本ロジックで、cross-feature 消費のための複製は DRY 違反・drift リスクの方が大きいと判断し、具体的な問題が出るまで現状維持。`eslint.config.js` の `crossFeatureImportBanAllowlist` で管理）。

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
- 新規リーフルートを追加したら `docs/spec/ui-design-compliance.md` §2 のページ表を同じコミットで更新すること。C1/C3/C5/C6b/C7〜C19 は `pnpm design-audit`（make ci 配線済み）で機械検証される。C8 は `src/features/*/routes/*.tsx` の PageLayout / Master*Page /相対パス完全一致 allowlist（44件 = 独自page 9 + helper 35）を検査する — 正当な非 PageLayout ルートは `design-system-audit.mjs` の対応する allowlist へ同一コミットで追記すること。C18 は `TableHead` / `TableCell` 呼び出し側による非仕様 typography・vertical padding の再上書き、C19 は `DataTableRow` / `SortableDataTableRow` / `TableRow` / raw `tr` の行全体 `onClick` を禁止する。遷移・表示・選択はcell内のnative link/button、並べ替えは44px drag handleを使う。空stateで追加vertical paddingが必要な場合だけ `data-empty-state` + `colSpan` + `text-center` を明示する。

## DESIGN.md / 製品デザイン適用範囲（MANDATORY・FE11 2026-07-21）

- **色の正本**は `docs/spec/design-system.md`。brand と semantic primary はともに **`#038B94`** / active **`#027078`** を使う。認証・製品識別と汎用 CTA・選択・focus の意味役割はトークン名で分ける。臨床 semantic 色、業務 status 色、nav canvas-soft は製品判断として維持する。
- **タイポ / 形状 / 余白 / エレベーション / コンポーネント寸法の正本**は `DESIGN.md`。hairline `#E6E6E6`、caption 14px / eyebrow 12px、radius 最小 4px、`button-primary` = pill、テーブルヘッダ = `ex-data-table-cell` 様式（`STYLE.tableHeader*`）へ字義で従う。
- brand と primary は同じ色値だが、認証・ロゴ・明示的 brand CTA は brand トークン、汎用 CTA・リンク・active/focus は primary トークンを使う。臨床・業務 semantic 色を構造用途へ流用しない。

## 臨床安全境界 (MANDATORY・FE12 2026-07-28 に FE-refactor.md から移設)

臨床 sentinel・権限・日付を扱う実装では、次の3則を必ず維持する。

1. **臨床 sentinel は生成型から表示・操作境界まで欠落させない。** 死亡は明示的な positive match で遷移・mutation callback を拒否し、危険「高」は非色 cue を伴う警告として扱う。死亡 status と死亡日時が不整合なら再登録導線を出さない。
   - 実装の参照実装: `src/features/owners/components/OwnerPetsSection.tsx:108,162`（死亡なら要素自体をレンダリングしない）＋ `:113,123,133,143,153,168`（callback 側も `if (current.status === "死亡" || ...) return;` で二重に拒否）。**新しい pet 操作を追加するときはこの二重防壁に揃える。**
2. **権限は action 別の最新値を mutation 直前に再検査する。** UI の非表示・disabled・route guard だけを最終防壁にしない。view/edit 共用の唯一の detail route は read access を維持し、mutation 境界を fail-closed にする。commit 直後にも発火し得る callback の permission ref は `useLayoutEffect` で同期する。
3. **臨床 date-only は JST の厳密過去で判定する。** `YYYY-MM-DD` 契約を guard し、`todayJSTISO()`（`src/lib/jst-date.ts:26`）との文字列比較 `<`（同 `:32` の `isPastJSTDate`）を使う。現在時刻との `Date` 比較で当日を期限超過にしない。**FE12 M-05 の runtime 実測で today/future の誤検出0件を確認済み。**

## 却下済み提案 — 再提案しない (FE12 2026-07-28 に FE-refactor.md から移設)

調査のうえ「やらない」と決めた項目。**再提案する場合は、下記の再開条件を満たす新しい証拠を添えること。**

- **manual chunk の追加分割投資を行わない（2026-07-27）** — 実測 522.71 kB（gzip 145.80 kB）で 500 kB 警告に該当するが、`app/routes/operations-routes.tsx` の lazy 境界により独立 chunk として正しく分割済みで、`/manual` を開いた利用者だけが取得する。build 警告は汎用閾値であって業務上の問題の証拠ではなく、表示遅延の申告も無い。証拠なしの最適化は product-philosophy ①違反。警告閾値の引き上げによる黙らせも行わない（サイレント化は⑤の禁止事項）。**再開条件 = `/manual` の表示時間に関する具体的な業務上の申告が出た場合のみ。**
- **死亡ペットの「グレーアウト」は badge のみで合格（2026-07-28・曽我裁定）** — 現行実装（`src/features/owners/components/OwnersListTable.tsx:236-241` → `getPetStatusColor` → `src/lib/status-helpers.ts:176-180` が死亡へ `BADGE.grayHover`）を是とする。**行全体のグレーアウトは求めない。** 根拠は「死亡ペットの情報自体は正常に読めるべきであり、行全体を落とすと可読性を下げる」。
- **一覧の行アクションをペットの生死でブロックしない（2026-07-28）** — `/owners` 一覧の行アクション（`編集`/`レポート`/`削除`）は **飼主レベルの操作**である（`OwnersListTable.tsx:56-59` が `onEdit(ownerId)` / `onDeleteRequest(ownerId, ownerName)` / `onReport(ownerId, petId)` を定義）。一覧がペット行単位で描画されるため「死亡ペットの行に編集・削除が出ている」ように見えるが、対象は飼主であり飼主は生存している。**ペットの生死で飼主操作をブロックすれば正当な業務を壊す。** ペット単位の物理ブロックは上記「臨床安全境界」1のとおり `OwnerPetsSection.tsx` に実装済みで、`docs/spec/specification.md:21` の要件は達成されている。
