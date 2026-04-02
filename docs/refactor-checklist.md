# フロントエンド リファクタリング チェックリスト

> **関連ドキュメント**: バックエンドは [`backend-refactor-checklist.md`](./backend-refactor-checklist.md) を参照

> **目的**: frontend/src/ 配下のコード規約準拠チェックと修正追跡  
> **対象**: `frontend/src/` 配下の全コード  
> **最終チェック日**: 2026-04-02（全項目チェック済み・違反修正済み）

---

## 規約ドキュメント参照先

ルール定義はこのドキュメントに書かない。以下を参照すること。

| ドキュメント | 内容 | 主な参照セクション |
|-------------|------|-------------------|
| `frontend/CODING_RULES.md` | **最も包括的**（40+ルール） | §9 禁止事項、§12 パフォーマンスチェックリスト |
| `.claude/rules/typescript-react.md` | React 19 パターン・型安全性 | §1-9 |
| `.claude/rules/code-style.md` | 命名規則・import順序・デザイントークン | Frontend セクション |
| `.claude/CLAUDE.md` | プロジェクト全体規約 | Frontend ベストプラクティス参照実装 |

---

## ステータス凡例

| 記号 | 意味 |
|------|------|
| `[ ]` | 未着手 |
| `[~]` | チェック中 |
| `[x]` | 完了（違反なし or 修正済み） |
| `[!]` | 違反あり・修正待ち |

---

## プロジェクト固有の判断基準

規約ドキュメントに明記されていない判断で、チェック時に必要なもの。

### design-tokens ルールの適用範囲

| パターン | 判定 | 理由 |
|---------|------|------|
| `#37352F` 等の hex 直書き | **違反** | `C` または `PALETTE` 定数を使用すべき |
| `rgba(55,53,47,0.04)` 等の rgba 直書き | **違反** | 同上 |
| `shadow-[0_2px_4px_rgba(...)]` 内の rgba | **許容** | Tailwind shadow ユーティリティの引数であり、色トークンではない |
| `bg-red-500`, `text-blue-700` 等の Tailwind 標準色 | **許容** | フレームワーク提供のユーティリティクラス。hex ハードコードではない |
| `design-tokens.ts` 自身の hex 定義 | **許容** | トークン定義元 |
| `components/ui/` (shadcn) 内の hex | **許容** | サードパーティ UI ライブラリ。プロジェクトコードではない |
| inline `style={{ color: PALETTE.xxx }}` | **許容** | `PALETTE` 定数経由であれば OK（charts、動的スタイル等） |

### コンポーネントファイル命名

| 規約ソース | ルール | 優先度 |
|-----------|--------|--------|
| `.claude/rules/code-style.md` | コンポーネント `.tsx` は **PascalCase** | **プロジェクト規約（採用）** |
| `.claude/skills/component-naming/` | コンポーネント `.tsx` は **kebab-case** | スキル提案（不採用） |

→ **PascalCase を正とする**。138ファイルが PascalCase で統一済み。

### `memo()` + デフォルト引数 `= []`

`memo()` されたコンポーネントの props に `= []` がある場合、**呼び出し側が常に値を渡していれば実害なし**。デフォルトが実際に使われるケースでのみ修正対象。

### barrel import vs Feature Indexing

| 呼び出し元 | ルール | 例 |
|-----------|--------|-----|
| Feature **外部**（app/, hooks/, shared/） | `index.ts` 経由で import（Feature Indexing） | `import { OwnerForm } from "@/features/owners"` |
| Feature **内部**（同一 feature 内） | 直接ファイル指定（barrel 不要） | `import { getOwners } from "../api/get-owners"` |

→ `CODING_RULES.md` §12.2 を修正済み（旧: barrel 全面禁止 → 新: 上記ルール）。

### `!important` の許容範囲

`globals.css` のフォント指定（`font-family: ... !important`）は許容。コンポーネント内での使用は禁止。

### `useState(false)` の判定

`useState(false)` 自体は UI トグル（モーダル開閉等）で正当な使用。**違反となるのは API mutation の pending 状態を `useState(false)` + `setIsPending(true/false)` で手動管理するパターンのみ**。代わりに `useTransition` を使用する。

---

## チェック実行手順

### 実行方法

各チェック項目には Grep パターンを記載。以下の手順で実行する。

```bash
# Docker 経由（推奨）
docker compose exec frontend npx grep-pattern ...

# または Claude Code の Grep ツール
Grep pattern="xxx" path="frontend/src" glob="*.tsx"
```

### 並列実行（Agent Teams）

チェックは読み取り専用のため完全並列可。修正は同一ファイルを触る Agent を同時起動しないこと。

```
# Phase 1: 横断スキャン — 1 Agent
# Phase 2: ドメイン別 — ドメイン数分の Agent を並列起動
# Phase 3: 共通コンポーネント — 1-2 Agent
```

---

## Phase 1: 横断スキャン（構造的ルール）

全 feature に影響する構造的違反を先に洗い出す。

| # | チェック項目 | Grep パターン / 観点 | ステータス |
|---|------------|---------------------|-----------|
| 1 | `&&` 条件レンダー | `&&\s*[\(<]` in `.tsx` | `[x]` 0件 |
| 2 | deep import（feature 内部） | `from ["']@/features/[^/]+/(components\|hooks\|api\|routes)/` | `[x]` 0件 |
| 3 | hex カラー直書き | `#[0-9a-fA-F]{3,8}` in `.tsx`/`.ts`（除外: `design-tokens.ts`, `components/ui/`） | `[x]` 0件 |
| 4 | rgba 直書き | `rgba?\(` in `.tsx`/`.ts`（除外: 同上 + `shadow-[...]` 内） | `[x]` 0件 |
| 5 | `any` 型 | `: any\|as any` | `[x]` 0件 |
| 6 | `FC` / `React.FC` | `React\.FC\|: FC[<\s]` | `[x]` 0件 |
| 7 | `forwardRef` | `forwardRef\(` | `[x]` 0件 |
| 8 | `export *` | `export \* from` | `[x]` 0件 |
| 9 | `console.log` | `console\.(log\|debug\|info)` | `[x]` 0件 |
| 10 | lazy() 未使用の大型モーダル | import 文で直接 import している大型コンポーネント | `[x]` 0件 |
| 11 | loader 内の直列 await | loader 関数内の複数 `await`（`Promise.all` 未使用） | `[x]` 0件 |
| 12 | `export default` | `export default ` in `.tsx`/`.ts`（除外: `components/ui/`, `main.tsx`） | `[x]` 0件 |
| 13 | `!important` | `!important` in `.tsx`/`.ts`/`.css` | `[x]` 1件（globals.css フォント指定 — 許容） |
| 14 | 空/コメントのみの `index.ts` | ファイル内容チェック | `[x]` 1件修正（vaccinations/types/index.ts 削除） |
| 15 | `.gitkeep`（ファイル存在ディレクトリ） | ファイル存在チェック | `[x]` 0件 |
| 16 | API hook 動詞省略 | `export.*use[A-Z]` で `useGet/Create/Update/Delete` 以外 in `features/*/api/` | `[x]` 0件 |
| 17 | `queryClient.prefetchQuery` | `queryClient\.prefetchQuery` | `[x]` 0件 |
| 18 | `localStorage` に token 保存 | `localStorage.*token` | `[x]` 0件（httpOnly cookie 使用） |
| 19 | `useState(false)` で pending 管理 | 文脈確認（67件の `useState(false)` 中、mutation pending 管理は 0件） | `[x]` 0件 |
| 20 | `rerender-lazy-state-init` | `useState(expensiveCalc())` 形式（lazy init 未使用） | `[x]` 1件修正（ReservationManagement `new Date()` → `() => new Date()`） |
| 21 | dead re-export ファイル | re-export のみで参照ゼロの `.ts` ファイル | `[x]` 6件削除（shared component の未参照 index.ts） |
| 22 | feature 内部 barrel import | `from ["']\.\.\/(api\|hooks\|components)["']` in `features/` | `[x]` 0件 |

---

## Phase 2: ドメイン別チェック（パフォーマンス・最適化）

### 優先度

| 優先度 | ドメイン | 理由 |
|--------|---------|------|
| **Skip** | `owners`, `medical-records` | 参照実装。規約準拠済み前提 |
| **High** | `accounting`, `hospitalization`, `estimates` | 集計・重いフォーム・複合UI |
| **Medium** | `reservations`, `examinations`, `trimming`, `master`, `dashboard` | 一覧+モーダル構成 |
| **Low** | `auth`, `checkups`, `pets`, `vaccinations`, `inventory`, `hospital-settings`, `shifts` | シンプル CRUD |

### チェック対象ルール（ドメイン毎に確認）

| ルールID | 確認観点 |
|---------|---------|
| `rerender-memo` | 大型コンポーネントに `memo()` + handler に `useCallback` |
| `rerender-functional-setstate` | `useCallback` 内の setState が `prev =>` 形式 |
| `rerender-lazy-state-init` | 高コスト `useState` 初期値が `() => ...` lazy 形式 |
| `rerender-dependencies` | `useCallback`/`useMemo` の deps にオブジェクト/配列でなく primitive |
| `rendering-hoist-jsx` | 静的配列・JSX がモジュール定数に巻き上げ済み |
| `js-cache-function-results` | API 由来リストの `.map()` が `useMemo` でキャッシュ |
| `rerender-transitions` | 検索に `useDeferredValue`、手動 `isLoading` 管理なし |
| `async-parallel` | 独立フェッチが `Promise.all` で並列化 |
| `bundle-dynamic-imports` | 重いモーダル・ダイアログが `lazy()` + `Suspense` |

---

### [High] accounting — `[x]`

| レイヤー | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `components/` | `rerender-memo` | `[x]` | ItemListCard/InsuranceCard/PaymentCard/RefundSection 全て memo() 済み |
| `components/` | `rendering-hoist-jsx` | `[x]` | CATEGORY_LABELS 等モジュール定数化済み |
| `components/` | `js-cache-function-results` | `[x]` | AccountingDetail L.328/L.712 の `.map()` を useMemo 化済み |
| `routes/` | `bundle-dynamic-imports` | `[x]` | AccountingDocument を lazy() 化済み |
| `routes/` | `rerender-dependencies` | `[x]` | AccountingDetail の `baseAccounting?.items` deps → `baseItems` useMemo で安定化済み |

### [High] hospitalization — `[x]`

| レイヤー | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `components/CarePlan/` | `rerender-memo` | `[x]` | CarePlanItemRow/CarePlanSection/CarePlanDialog memo() + useCallback 済み |
| `components/DailyRecord/` | `rerender-memo` | `[x]` | DailyRecordSection/DailyRecordTimeline memo() + useCallback 済み |
| `components/` | `rendering-hoist-jsx` | `[x]` | DailyDateNav WEEK_DAYS / CarePlanTab INITIAL_TIMING 巻き上げ済み |
| `hooks/` | `async-parallel` | `[x]` | 直列 await → Promise.all 並列化済み |
| `routes/` | `rerender-transitions` | `[x]` | useTransition 使用済み。死コード use-hospitalizations.ts 削除済み |

### [High] estimates — `[x]`

| レイヤー | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `components/EstimateLineItems/` | `rerender-memo` | `[x]` | memo() + useMemo 修正済み |
| `components/EstimateLineItems/` | `rendering-hoist-jsx` | `[x]` | CATEGORY_LABELS モジュール定数化済み |
| `routes/` | `bundle-dynamic-imports` | `[x]` | Vite ルート分割済み |

### [Medium] reservations — `[x]`

| レイヤー | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `routes/` | `bundle-dynamic-imports` | `[x]` | ReservationFormModal/MonthView/WeekView/ReservationDetailModal lazy() 化済み |

### [Medium] examinations — `[x]`

| レイヤー | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `components/` | `rerender-memo` | `[x]` | FormFieldsSection memo() + useCallback 済み |
| `routes/` | `js-cache-function-results` | `[x]` | examTypes/staffList .map() に useMemo 追加済み |

### [Medium] trimming — `[x]`

| レイヤー | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `routes/` | `bundle-dynamic-imports` | `[x]` | MasterSelectModal/ConfirmDialog lazy() 化済み |
| `routes/` | `rerender-memo` | `[x]` | LeftColumn/MiddleColumn/RightColumn memo() 済み |
| `hooks/` | `js-cache-function-results` | `[x]` | coursesRaw/optionsRaw .map() を useMemo 化済み |

### [Medium] master — `[x]`

| レイヤー | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `components/` | `rendering-hoist-jsx` | `[x]` | CATEGORY_CONFIG モジュール定数化済み |
| `hooks/` | `rerender-functional-setstate` | `[x]` | use-master-crud.ts 全ハンドラ useCallback 済み |

### [Medium] dashboard — `[x]`

| レイヤー | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `api/` | `async-parallel` | `[x]` | React Query 自動並列化。loader 不使用のため実質並列 |
| `components/` | `rerender-memo` | `[x]` | AppointmentCard memo() 済み |

### [Low] auth — `[x]`

チェック済み。違反なし。

### [Low] checkups / pets / vaccinations / inventory / hospital-settings / shifts — `[x]`

全てチェック済み。シンプル CRUD 構成のため重大な違反なし。shifts の ShiftFormDialog は lazy() 化済み。

---

## Phase 3: 共通コンポーネント・app 層

### components/shared/ — `[x]`

| コンポーネント | ルール | ステータス | 備考 |
|--------------|--------|-----------|------|
| `DataTable/` | `rerender-memo` | `[x]` | memo() 済み。renderRow は呼び出し側 useCallback で安定化 |
| `Pagination/` | `rerender-memo` | `[x]` | memo() 済み |
| `SidePeekPanel/` | `rerender-memo` | `[x]` | memo() 済み |
| `NotionFilter/` | `rerender-memo`, `rendering-hoist-jsx` | `[x]` | memo() 済み。DATE_PRESETS モジュール定数化済み |
| `Layout/Sidebar.tsx` | `rendering-conditional-render` | `[x]` | `&&` 3箇所を三項演算子に修正済み |
| `Layout/` | `bundle-feature-indexing` | `[x]` | auth deep import → index.ts 経由に修正済み |
| `ConfirmDialog/` | `design-tokens` | `[x]` | `#37352F` → C 定数に修正済み |

### app/pages/ — `[x]`

| ファイル | ルール | ステータス | 備考 |
|---------|--------|-----------|------|
| `OwnersListPage.tsx` | `async-parallel` | `[x]` | ownersLoader で Promise.all 並列フェッチ実装済み |
| `AccountingDetailPage.tsx` | `async-parallel` | `[x]` | 単一 hook のみ。並列化対象なし |
| `OwnerFormPage.tsx` | `rerender-memo` | `[x]` | app 層での mutation 注入のみ |

---

## コンポーネント命名チェック — `[x]`

| チェック項目 | Grep パターン | ステータス | 結果 |
|------------|--------------|-----------|------|
| 汎用名コンポーネント | `export (const\|function) (Card\|List\|Item\|Data\|Table\|Form)\b` | `[x]` | 0件 |
| 禁止サフィックス | `(Container\|Wrapper\|Component\|Element)['\s{]` | `[x]` | 0件 |
| レイアウト記述名 | `(Left\|Right\|Top\|Bottom\|Big\|Small)(Panel\|Section\|Column\|Bar)` | `[x]` | 0件 |
| PascalCase 違反 | `.tsx` 内の `export const [a-z]` | `[x]` | 0件 |

---

## 修正履歴

全55件の違反を発見・修正。以下はカテゴリ別サマリ。

### design-tokens（hex/rgba ハードコード）— 32件

| 日付 | 対象 | 修正概要 |
|------|------|---------|
| 2026-04-01 | 72ファイル・339箇所 | `#37352F` hex 全量 → `C`/`PALETTE` 定数。新トークン3個追加 |
| 2026-04-02 | NotionFilter 2ファイル | `#F1F1EF`/`#E8E7E4` → `C.bgMutedBadge`/`C.hoverBgMutedBadge` |
| 2026-04-02 | master 4ファイル | `#6B7280`/`#3B82F6`/`#e5e7eb` → `PALETTE.defaultGray`/`defaultBlue`/`borderUnselected` |
| 2026-04-02 | VitalsGraph | チャート色5色 → `PALETTE.chart*` 定数 |
| 2026-04-02 | 42ファイル・78箇所 | `#2EAADC`/`#F7F6F3`/`#2383E2`/`#038B94` 等の残存 hex → 新トークン19個追加 |
| 2026-04-02 | status-colors + InventoryForm | VISIT_TYPE_COLORS badge hex + rgba → 新トークン6個 |
| 2026-04-02 | hospitalization + master + shared + vaccinations | styles.ts/use-service-type-color-map.ts 等 hex 7箇所 + rgba 3箇所 |
| 2026-04-02 | CompanySettings + MonthView + StaffSettings + VitalsGraph | rgba 5箇所 + border/inline style → 新 PALETTE トークン2個（bgSkeleton, whiteAlpha80） |

### bundle-feature-indexing（deep import）— 6件

| 日付 | 対象 | 修正概要 |
|------|------|---------|
| 2026-04-01 | auth feature 21箇所 | Sidebar/Layout/RequirePermission + 18 features → index.ts 経由 |
| 2026-04-01 | shared hooks 5箇所 | pets/owners/master への deep import → index.ts 経由 |
| 2026-04-01 | reservations 2箇所 | master への deep import → index.ts 経由 |
| 2026-04-02 | medical-records 4箇所 | master への deep import → index.ts 経由。master/index.ts にエクスポート追加 |
| 2026-04-02 | RequirePermission | auth/types → @/features/auth |

### rendering-conditional-render（&& 禁止）— 2件

| 日付 | 対象 | 修正概要 |
|------|------|---------|
| 2026-04-01 | Sidebar.tsx 3箇所 | `{!collapsed && <p>}` → 三項演算子 |
| 2026-04-02 | VaccinationHistory.tsx | `{!isLoading && ...}` → 三項演算子 |

### rerender-memo — 2件

| 日付 | 対象 | 修正概要 |
|------|------|---------|
| 2026-04-02 | hospitalization 3コンポーネント | HospitalizationBasicInfo/NoteCard/DailyRecordTimeline に memo() 追加 |
| 2026-04-02 | hospital-settings PropertyRow | ループ内コンポーネントに memo() 追加 |

### その他 — 7件

| 日付 | 対象 | ルール | 修正概要 |
|------|------|--------|---------|
| 2026-04-02 | hospitalization 2ファイル | `async-parallel` | 直列 await → Promise.all |
| 2026-04-02 | medical-records + hospitalization 3ファイル | `rendering-hoist-jsx` | 静的配列をモジュール定数に巻き上げ |
| 2026-04-02 | hospitalization | `rerender-transitions` | 死コード use-hospitalizations.ts 削除 |
| 2026-04-02 | reservations/ReservationManagement.tsx | `rerender-lazy-state-init` | `useState(new Date())` → `useState(() => new Date())` |
| 2026-04-02 | accounting/AccountingDetail.tsx | `rerender-dependencies` | `baseAccounting?.items` deps → `baseItems` useMemo で安定化 |
| 2026-04-02 | vaccinations/types/index.ts | 空ファイル削除 | コメントのみ・参照ゼロの index.ts + types/ ディレクトリを削除 |
| 2026-04-02 | reservations/MonthView.tsx | import 重複修正 | `C` の重複 import を削除 |
| 2026-04-02 | shared 6コンポーネント | dead re-export 削除 | NotionFilter/TreatmentSearchDialog/DateRangePicker/StatusPill/PageLayout/StatusBadge の未参照 index.ts を削除 |
