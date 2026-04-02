# フロントエンド リファクタリング チェックリスト

> **関連ドキュメント**: バックエンドは [`backend-refactor-checklist.md`](./backend-refactor-checklist.md) を参照

> **目的**: Vercel React Best Practices 規約への準拠確認・違反修正の進捗管理  
> **対象**: `frontend/src/` 配下の全コード  
> **基準**: `.claude/rules/typescript-react.md` + `.claude/skills/vercel-react-best-practices/` + `.claude/skills/component-naming/`  
> **更新**: 各ドメインのチェック完了時に更新する

---

## 推奨実行戦略 — Agent Teams による並列処理

チェック・修正は **Agent Teams で並列実行**することで大幅に高速化できる。
各 Phase の指示例を以下に示す。

### Phase 1（横断スキャン）— 1 Agent で完結

```
Explore エージェントを 1 つ起動し、
docs/refactor-checklist.md の「Phase 1: 横断スキャン」全項目を実行せよ。
発見した違反を「発見済み違反ログ」テーブルに追記し、Phase 1 ステータスを更新すること。
```

### Phase 2（ドメイン別）— High/Medium/Low を並列起動

High 優先度の 3 ドメインを同時にチェックさせる例：

```
# 1 メッセージで 3 Agent を同時起動
Agent 1（Explore）: features/accounting/ を docs/refactor-checklist.md の
  accounting セクションのルールに従ってチェックし、発見内容を発見済み違反ログに追記せよ。

Agent 2（Explore）: features/hospitalization/ を docs/refactor-checklist.md の
  hospitalization セクションのルールに従ってチェックし、発見内容を発見済み違反ログに追記せよ。

Agent 3（Explore）: features/estimates/ を docs/refactor-checklist.md の
  estimates セクションのルールに従ってチェックし、発見内容を発見済み違反ログに追記せよ。
```

Medium 5 ドメインも同様に 1 メッセージで 5 Agent 並列起動できる。

### Phase 2（修正）— ドメイン毎に implementer Agent

チェック完了後、違反ログを元に修正を並列実行：

```
# 修正も並列
Agent 1（implementer）: 発見済み違反ログの accounting 行を修正せよ。
Agent 2（implementer）: 発見済み違反ログの hospitalization 行を修正せよ。
```

### Phase 3（共通 + 命名）— 最後に 2 Agent 並列

```
Agent 1（Explore）: components/shared/ を Phase 3 ルールに従ってチェックせよ。
Agent 2（Explore）: frontend/src/ 全体の component-naming 横断スキャンを実行せよ。
```

> **注意**: 複数 Agent が同一ファイルを同時編集すると競合する。
> チェック（読み取り専用）は完全並列可。修正は同一ファイルを触る Agent を同時起動しないこと。

---

## ステータス凡例

| 記号 | 意味 |
|------|------|
| `[ ]` | 未着手 |
| `[~]` | チェック中 |
| `[x]` | 完了（違反なし or 修正済み） |
| `[!]` | 違反あり・修正待ち |

---

## チェック対象ルール一覧

### Critical（必ず修正）

| ルールID | 内容 | チェック観点 |
|---------|------|------------|
| `rendering-conditional-render` | `&&` 禁止 → 三項演算子 | JSX 内の `{condition &&` パターン |
| `async-parallel` | 独立フェッチは `Promise.all` 並列化 | loader / useEffect 内の直列 `await` |
| `bundle-dynamic-imports` | 重いモーダル・ダイアログは `lazy()` + `Suspense` | import文で直接importしている大型コンポーネント |
| `bundle-feature-indexing` | feature外からの deep import 禁止 | `@/features/xxx/components/YYY` 形式のimport |

### Medium（パフォーマンス改善）

| ルールID | 内容 | チェック観点 |
|---------|------|------------|
| `rerender-memo` | 大型コンポーネントは `memo()` + `useCallback` | memo未使用の重いコンポーネント |
| `rerender-functional-setstate` | `setState` は `prev =>` 形式 | `useCallback` 内で直接 state 参照の setState |
| `rerender-dependencies` | `useCallback` deps にオブジェクト禁止 | deps配列に `{}` / `[]` / オブジェクト変数 |
| `rerender-lazy-state-init` | 高コスト初期化は `useState(() => ...)` | `useState(expensiveCalc())` パターン |
| `rerender-transitions` | 検索は `useDeferredValue`、API書き込みは `useTransition` | `useState` で isLoading を手動管理 |
| `rendering-hoist-jsx` | 静的JSX（Select選択肢等）はモジュール定数に巻き上げ | コンポーネント内で毎回生成される静的配列・JSX |
| `js-cache-function-results` | APIリスト由来のJSX生成は `useMemo` | `.map()` が依存なしで毎回再実行 |

### Low（最適化）

| ルールID | 内容 | チェック観点 |
|---------|------|------------|
| `bundle-barrel-imports` | バレルファイル経由の重い import を直接 import に | `index.ts` 経由で多数を一括 import |
| `rerender-memo-with-default-value` | デフォルト非プリミティブ props はモジュール定数に hoist | `defaultProps = {}` / デフォルト引数 `= []` |
| `rerender-derived-state-no-effect` | effect 不要の派生 state は render 中に計算 | `useEffect` → `setState` で派生値を更新 |
| `js-set-map-lookups` | 繰り返し検索は `Set` / `Map` で O(1) 化 | ループ内の `Array.find` / `Array.includes` |

---

## コンポーネント命名規則チェック（component-naming）

> **基準**: `.claude/skills/component-naming/`  
> **横断チェック推奨**: ドメイン個別より全体 Grep の方が効率的

### 命名ルール一覧

| ルールID | 内容 | チェック観点 | 例 |
|---------|------|------------|-----|
| `naming-pascal-case` | コンポーネントは PascalCase 必須 | `export const` / `export function` の名前 | `trendChart` → `TrendChart` |
| `naming-domain-role` | ドメイン + ロール の組み合わせ | 汎用名単体（`Card`, `List`, `Item`, `Data`）を禁止 | `Card` → `PatientCard` |
| `naming-no-bad-suffix` | 禁止サフィックス | `-Container`, `-Wrapper`, `-Component`, `-Element` | `CardContainer` → `MetricCard` |
| `naming-no-layout-desc` | レイアウト・見た目の名前禁止 | Left/Right/Top/Bottom/Big/Small/Red/TwoColumn 等 | `LeftSidebar` → `NavigationSidebar` |
| `naming-file-kebab` | ファイル名は kebab-case | `.tsx` ファイル名がコンポーネント名と対応しているか | `TrendChart.tsx` → `trend-chart.tsx` |
| `naming-compound` | 関連サブコンポーネントは Compound pattern | `HeroPostImage` / `HeroPostTitle` 等の flat 命名 | `FormLabel` → `Form.Label` |

### 横断スキャンチェック項目

| チェック項目 | 観点 | ステータス | 発見数 |
|------------|------|-----------|--------|
| 汎用名コンポーネント | `export (const\|function) (Card\|List\|Item\|Data\|Table\|Form\b)` | `[x]` | 0 |
| 禁止サフィックス | `(Container\|Wrapper\|Component\|Element)['\s{]` | `[x]` | 0 |
| レイアウト記述名 | `(Left\|Right\|Top\|Bottom\|Big\|Small)(Panel\|Section\|Column\|Bar)` | `[x]` | 0 |
| PascalCase 違反 | `.tsx` ファイル内の `export const [a-z]` | `[x]` | 0 |
| ファイル名 PascalCase | `features/`, `components/` 配下に `.tsx` が PascalCase になっているか | `[x]` | 138（プロジェクト規約 `.claude/rules/code-style.md` が PascalCase を規定。component-naming スキルの kebab-case とは競合するがプロジェクト規約優先） |

### ドメイン別命名チェック

> Phase 2 のドメインチェック時に合わせて実施する

| ドメイン | ステータス | 発見内容 |
|---------|-----------|---------|
| `accounting` | `[x]` | |
| `hospitalization` | `[x]` | |
| `estimates` | `[x]` | |
| `reservations` | `[x]` | |
| `examinations` | `[x]` | |
| `trimming` | `[x]` | |
| `master` | `[x]` | |
| `dashboard` | `[x]` | |
| `auth` | `[x]` | |
| `checkups` | `[x]` | |
| `pets` | `[x]` | |
| `vaccinations` | `[x]` | |
| `inventory` | `[x]` | |
| `hospital-settings` | `[x]` | |
| `shifts` | `[x]` | |
| `components/shared` | `[x]` | |

---

## Phase 1: 横断スキャン（Critical ルール）

> **目的**: 全 feature に影響する構造的違反を先に洗い出す  
> **方法**: Grep でコードベース全体を機械的にスキャン

| チェック項目 | コマンド / 観点 | ステータス | 発見数 |
|------------|----------------|-----------|--------|
| `&&` 条件レンダー | `grep -r "{\s*\w\+\s*&&"` | `[x]` | 再検証: 0件（全て三項演算子使用済み） |
| deep import（feature内部） | `grep -r "@/features/.*/components/"` | `[x]` | 再検証: 0件（cross-feature deep import なし） |
| lazy() 未使用の大型モーダル | import文でModalを直接参照 | `[x]` | 再検証: 0件 |
| loader内の直列 await | `useQuery` / loader 関数内の複数 await | `[x]` | 再検証: 0件 |
| design-tokens hex ハードコード | 全 hex カラー直接使用 | `[x]` | 再検証: **78箇所・42ファイル追加発見**（`#2EAADC` 13件, `#F7F6F3` 31件, `#2383E2` 14件, `#038B94` 8件, その他 12件）→ 全て `C` / `PALETTE` 定数に修正済み。新トークン: `medicalBlue`, `bgMedicalBlue/5/90`, `textMedicalBlue`, `borderLMedicalBlue`, `ringMedicalBlue`, `focusRingMedicalBlue`, `hoverBorderMedicalBlue50`, `textBrand`, `bgBrand10`, `hoverBgBrand`, `focusRingBrand`, `borderBrand`, `focusBorderAccent`, `dataCheckedBg/Border Accent/Brand`, `bgMuted` |

---

## Phase 2: ドメイン別チェック

### 優先度の根拠

| 優先度 | ドメイン | 理由 |
|--------|---------|------|
| **Skip** | `owners`, `medical-records` | プロジェクト参照実装。規約準拠済み前提 |
| **High** | `accounting`, `hospitalization`, `estimates` | 集計・重いフォーム・複合UI → rerender・bundle 問題が出やすい |
| **Medium** | `reservations`, `examinations`, `trimming`, `master`, `dashboard` | 一覧+モーダル構成。dynamic import 漏れ多め |
| **Low** | `auth`, `checkups`, `pets`, `vaccinations`, `inventory`, `hospital-settings`, `shifts` | シンプルCRUD。致命的問題は少ない想定 |

---

### [High] accounting

**対象**: `features/accounting/`  
**注目点**: 請求集計・明細テーブル → memo最適化、リスト由来JSX

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `components/` | `rerender-memo` | `[x]` | ItemListCard/InsuranceCard/PaymentCard/RefundSection 全て memo() 済み |
| `components/` | `rendering-hoist-jsx` | `[x]` | CATEGORY_LABELS 等モジュール定数化済み |
| `components/` | `js-cache-function-results` | `[x]` | `AccountingDetail` L.328/L.712 の `.map()` を useMemo 化済み |
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `routes/` | `bundle-dynamic-imports` | `[x]` | AccountingDocument を lazy() 化済み |
| `app/pages/AccountingDetailPage.tsx` | `async-parallel` | `[x]` | 単一 hook のみ |

---

### [High] hospitalization

**対象**: `features/hospitalization/`  
**注目点**: タブ構成（CarePlan・DailyRecords） → dynamic import、memo分割

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `components/CarePlan/` | `rerender-memo` | `[x]` | CarePlanItemRow/CarePlanSection/CarePlanDialog memo() + useCallback 済み |
| `components/CarePlanTab/` | `bundle-dynamic-imports` | `[x]` | |
| `components/DailyRecord/` | `rerender-memo` | `[x]` | DailyRecordSection/DailyRecordTimeline memo() + useCallback 済み |
| `components/DailyRecordsTab/` | `bundle-dynamic-imports` | `[x]` | |
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `routes/` | `rerender-transitions` | `[x]` | useTransition 使用済み |

---

### [High] estimates

**対象**: `features/estimates/`  
**注目点**: 明細行（EstimateLineItems） → memo、hoist-jsx

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `components/EstimateLineItems/` | `rerender-memo` | `[x]` | memo() + useMemo 修正済み（EstimateLineItems.tsx） |
| `components/EstimateLineItems/` | `rendering-hoist-jsx` | `[x]` | CATEGORY_LABELS モジュール定数化済み |
| `components/EstimateStatusBadge/` | `rendering-hoist-jsx` | `[x]` | （小コンポーネント・対象外） |
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `routes/` | `bundle-dynamic-imports` | `[x]` | Vite ルート分割済み |

---

### [Medium] reservations

**対象**: `features/reservations/`  
**注目点**: カレンダー UI → heavy component の遅延ロード

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `routes/` | `bundle-dynamic-imports` | `[x]` | ReservationFormModal / MonthView / WeekView / ReservationDetailModal は lazy() 化済み |
| `routes/` | `rerender-transitions` | `[x]` | |

---

### [Medium] examinations

**対象**: `features/examinations/`  

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `components/` | `rerender-memo` | `[x]` | FormFieldsSection memo() + useCallback 済み |
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `routes/` | `bundle-dynamic-imports` | `[x]` | |
| `routes/` | `js-cache-function-results` | `[x]` | ExaminationForm.tsx:190-191 — examTypes/staffList .map() に useMemo なし → useMemo([examTypesRaw]) / useMemo([staffListRaw]) で修正済み |

---

### [Medium] trimming

**対象**: `features/trimming/`  

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `routes/` | `bundle-dynamic-imports` | `[x]` | MasterSelectModal / ConfirmDialog を lazy() 化済み |
| `routes/` | `rerender-memo` | `[x]` | LeftColumn / MiddleColumn / RightColumn memo() 済み |
| `hooks/` | `js-cache-function-results` | `[x]` | use-trimming-form.ts:400-401 coursesRaw/optionsRaw .map() を useMemo で修正済み |

---

### [Medium] master

**対象**: `features/master/`  
**注目点**: 多数のマスタカテゴリ → 静的選択肢の hoist、共通パターン

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `components/` | `rendering-hoist-jsx` | `[x]` | CATEGORY_CONFIG モジュール定数化済み |
| `hooks/` | `rerender-functional-setstate` | `[x]` | use-master-crud.ts 全ハンドラ useCallback 済み |
| `routes/` | `rendering-conditional-render` | `[x]` | |

---

### [Medium] dashboard

**対象**: `features/dashboard/`  
**注目点**: 複数データソースの並列フェッチ

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | useGetDashboard + useGetStaffs は React Query が自動並列化。loader パターン不使用のため実質並列 |
| `components/` | `rerender-memo` | `[x]` | AppointmentCard memo() 済み |
| `routes/` | `rendering-conditional-render` | `[x]` | AppointmentCard.tsx:184 は `cond1 && cond2 ? ... : null` 形式。&& は条件式内の論理演算子で違反なし |

---

### [Low] auth

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `hooks/` | `rerender-functional-setstate` | `[x]` | |

---

### [Low] checkups

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `routes/` | `rendering-conditional-render` | `[x]` | |

---

### [Low] pets

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `routes/` | `rendering-conditional-render` | `[x]` | |

---

### [Low] vaccinations

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `routes/` | `rendering-conditional-render` | `[x]` | |

---

### [Low] inventory

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `routes/` | `rendering-conditional-render` | `[x]` | |

---

### [Low] hospital-settings

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `api/` | `async-parallel` | `[x]` | |
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `hooks/` | `rerender-functional-setstate` | `[x]` | |

---

### [Low] shifts

| レイヤー | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `routes/` | `rendering-conditional-render` | `[x]` | |
| `routes/` | `bundle-dynamic-imports` | `[x]` | ShiftFormDialog lazy() 化済み |

---

## Phase 3: 共通コンポーネント・app層

### components/shared/

> **注目点**: 全 feature から呼ばれる → ここの修正インパクトが最大

| コンポーネント | ルール | ステータス | 発見内容 |
|--------------|--------|-----------|---------|
| `DataTable/` | `rerender-memo`, `js-cache-function-results` | `[x]` | memo() 追加済み。renderRow は呼び出し側が useCallback で安定化 |
| `FormDialog/` | `bundle-dynamic-imports`（呼び出し元） | `[x]` | 各 feature routes/ から lazy() 経由で import 済み |
| `MasterSelectModal/` | `bundle-dynamic-imports`（呼び出し元） | `[x]` | |
| `OwnerSearchModal/` | `bundle-dynamic-imports`（呼び出し元） | `[x]` | |
| `TreatmentSearchDialog/` | `bundle-dynamic-imports`（呼び出し元） | `[x]` | |
| `SearchBox/` | `rerender-transitions`（useDeferredValue） | `[x]` | 検索 deferral は呼び出し側責務（OwnersList.tsx 等で useDeferredValue 使用済み） |
| `Pagination/` | `rerender-memo` | `[x]` | memo() 追加済み |
| `SidePeek/` | `rerender-memo`, `bundle-dynamic-imports` | `[x]` | SidePeekPanel memo() 追加済み |
| `NotionFilter/` | `rendering-hoist-jsx`, `rerender-memo` | `[x]` | DATE_PRESETS モジュール定数化済み。memo() 追加済み |
| 全体 | `rendering-conditional-render` | `[x]` | &&パターンなし確認済み |
| `ConfirmDialog/` | `design-tokens` | `[x]` | `#37352F` ハードコード → `C.bgPrimary` / `C.hoverBgPrimaryDark` に修正済み |
| `Layout/` | `bundle-feature-indexing` | `[x]` | Layout.tsx + Sidebar.tsx の auth deep import → index.ts 経由に修正済み |
| `Layout/Sidebar.tsx` | `rendering-conditional-render` | `[x]` | `{!collapsed && <p>}` 3箇所 → 三項演算子に修正済み |

### app/pages/

| ファイル | ルール | ステータス | 発見内容 |
|---------|--------|-----------|---------|
| `AccountingDetailPage.tsx` | `async-parallel` | `[x]` | ローダーなし。単一 hook のみ |
| `OwnerFormPage.tsx` | `async-parallel`, `rerender-memo` | `[x]` | ローダーなし。app 層での mutation 注入のみ |
| `OwnersListPage.tsx` | `async-parallel` | `[x]` | ownersLoader で Promise.all 並列フェッチ実装済み |

---

## 発見済み違反ログ

> チェック中に発見した違反をここに蓄積する

| # | ファイル | ルールID | 内容 | 優先度 | ステータス |
|---|---------|---------|------|--------|-----------|
| 1 | `components/shared/ConfirmDialog/`, `FormDialog/` 他 | `bundle-dynamic-imports` | **Vite ルート分割済み。routes/ は既に別チャンク。** `router.tsx` の全ルートが `lazy: async () => { ... }` パターンで遅延ロード済み。`ConfirmDialog` / `FormDialog` / `TreatmentSearchDialog` / `ReservationFormModal` はすべて `features/routes/` or `features/components/` 内から使われており、それぞれのルートチャンクに閉じている（初期バンドル混入なし）。`Dashboard.tsx` の `ReservationFormModal` / `DashboardDetailModal` も既に `lazy()` 化済み。`app/pages/` 3 ファイルと `Layout.tsx` にはダイアログ import なし。追加の `lazy()` 対応は不要。 | Critical | `[x]` |
| 2 | `features/`, `components/shared/` 配下 全 .tsx | `naming-file-kebab` | **プロジェクト規約優先・違反ではない**: `.claude/rules/code-style.md` が「コンポーネントファイル（.tsx）: PascalCase」と明示。`component-naming` スキルの kebab-case ルールと競合するが、プロジェクト固有規約が優先される。 | Medium | `[x]` |
| 3 | `features/accounting/routes/AccountingDetail.tsx:328,712` | `js-cache-function-results` | `.map()` に useMemo なし → useMemo([items]) / useMemo([refunds]) で修正済み | Medium | `[x]` |
| 4 | `features/estimates/components/EstimateLineItems/EstimateLineItems.tsx` | `rerender-memo` | `memo()` なし + `.slice().sort().map()` に useMemo なし → memo() + useMemo([items]) で修正済み | Medium | `[x]` |
| 5 | `features/examinations/routes/ExaminationForm.tsx:190-191` | `js-cache-function-results` | examTypes/staffList の inline .map() に useMemo なし → useMemo([examTypesRaw]/[staffListRaw]) で修正済み | Medium | `[x]` |
| 6 | `features/trimming/routes/TrimmingForm.tsx:400-401` | `js-cache-function-results` | coursesRaw/optionsRaw の inline .map() に useMemo なし → useMemo([coursesRaw]/[optionsRaw]) で修正済み | Medium | `[x]` |
| 7 | `components/shared/DataTable/`, `Pagination/`, `SidePeek/`, `NotionFilter/` | `rerender-memo` | 4 共有コンポーネントに memo() なし → memo() 追加済み | Medium | `[x]` |
| 8 | `components/shared/Layout/Sidebar.tsx:271,288,301` | `rendering-conditional-render` | `{!collapsed && <p>}` パターン3箇所 → 三項演算子 `? ... : null` に修正済み | Critical | `[x]` |
| 9 | `Sidebar.tsx`, `Layout.tsx`, `RequirePermission.tsx` + 18 features ファイル | `bundle-feature-indexing` | auth feature への deep import 21箇所 → `@/features/auth` index.ts 経由に修正。`usePermission` エクスポート追加 | Critical | `[x]` |
| 10 | `hooks/use-pet.ts`, `use-owner.ts`, `use-master-items.ts`, `use-pet-selection-page.ts` | `bundle-feature-indexing` | shared hooks から pets/owners/master feature への deep import 5箇所 → index.ts 経由に修正。owners/index.ts に `getOwner`/`useGetOwner` エクスポート追加 | Critical | `[x]` |
| 11 | `reservations/routes/ReservationManagement.tsx`, `components/ReservationDetailModal.tsx` | `bundle-feature-indexing` | master feature への deep import 2箇所 → `@/features/master` index.ts 経由に修正 | Critical | `[x]` |
| 12 | `components/shared/ConfirmDialog/ConfirmDialog.tsx:49` | `design-tokens` | `bg-[#37352F]` / `hover:bg-[#37352F]/90` ハードコード → `C.bgPrimary` / `C.hoverBgPrimaryDark` に修正済み | Medium | `[x]` |
| 13 | frontend/src/ 全体（72ファイル・339箇所） | `design-tokens` | `#37352F` hex ハードコード全箇所を `C` / `PALETTE` 定数に一括修正。design-tokens に `borderBPrimary`, `dataActiveBorderB`, `dataActiveText` 新トークン追加 | Medium | `[x]` |
| 14 | `hospitalization/components/HospitalizationBasicInfo.tsx`, `HospitalizationNoteCard.tsx`, `DailyRecord/DailyRecordTimeline.tsx` | `rerender-memo` | 大型コンポーネント3件に memo() 未適用 → memo() 追加済み | Medium | `[x]` |
| 15 | `hospital-settings/routes/ClinicMasterSettings.tsx` (PropertyRow) | `rerender-memo` | ループ内レンダリングされる内部コンポーネントに memo() 未適用 → memo() 追加済み | Medium | `[x]` |
| 16 | `NotionFilter/FilterRuleRow.tsx`, `NotionFilter/SortPopover.tsx` | `design-tokens` | `#F1F1EF`/`#E8E7E4` hex ハードコード4箇所 → `C.bgMutedBadge`/`C.hoverBgMutedBadge` に修正。`hoverBgMutedBadge` 新トークン追加 | Medium | `[x]` |
| 17 | `medical-records/components/MedicalRecordVaccination.tsx`, `CheckupsTab.tsx`, `StaffSelectionModal.tsx` | `bundle-feature-indexing` | medical-records → master への deep import 4箇所 → `@/features/master` index.ts 経由に修正。master/index.ts に `useGetAllVaccinesMaster`/`useGetAllCheckupTypes`/`useGetStaffs` エクスポート追加 | Critical | `[x]` |
| 18 | `components/shared/RequirePermission.tsx` | `bundle-feature-indexing` | `@/features/auth/types` → `@/features/auth` に修正 | Critical | `[x]` |
| 19 | `master/routes/StaffSettings.tsx`, `PermissionGroupSettings.tsx`, `ServiceTypeSettings.tsx`, `PermissionGroupSelector.tsx` | `design-tokens` | `#6B7280`/`#3B82F6`/`#e5e7eb` hex 9箇所 → `PALETTE.defaultGray`/`defaultBlue`/`borderUnselected` に修正 | Medium | `[x]` |
| 20 | `medical-records/components/VitalsTab/VitalsGraph.tsx` | `design-tokens` | `#E07B54`/`#9C6EDE`/`#4CAF82`/`#e8e6e3`/`#9B9B97` チャート色9箇所 → `PALETTE.chart*` 定数に修正 | Medium | `[x]` |
| 21 | 42ファイル（78箇所） | `design-tokens` | `#2EAADC`(13件), `#F7F6F3`(31件), `#2383E2`(14件), `#038B94`(8件) 等の残存 hex ハードコード → 全て `C`/`PALETTE` 定数に修正。新トークン19個追加（medical-blue系, brand系, accent-focus系, data-checked系, bgMuted） | Critical | `[x]` |
| 22 | `medical-records/components/VaccinationHistory.tsx:122` | `rendering-conditional-render` | `{!isLoading && filteredItems.map(...)}` → `? ... : null` に修正 | Critical | `[x]` |
| 23 | `hospitalization/hooks/use-hospitalization-list.ts`, `use-hospitalizations.ts` | `async-parallel` | 独立した `updateHospitalization` 2件の直列 await → `Promise.all` で並列化 | Medium | `[x]` |
| 24 | `medical-records/routes/MedicalRecordForm.tsx` | `rendering-hoist-jsx` | `VISIT_TYPE_OPTIONS` + `tabs` 静的配列をモジュール定数に巻き上げ | Medium | `[x]` |
| 25 | `hospitalization/components/DailyRecordsTab/DailyDateNav.tsx` | `rendering-hoist-jsx` | `weekDays` 静的配列をモジュール定数 `WEEK_DAYS` に巻き上げ | Medium | `[x]` |
| 26 | `hospitalization/components/CarePlanTab/CarePlanTab.tsx` | `rendering-hoist-jsx` | `["morning"]` 初期値をモジュール定数 `INITIAL_TIMING` に巻き上げ | Medium | `[x]` |
| 27 | `hospitalization/hooks/use-hospitalizations.ts` | `rerender-transitions` | 手動 `isLoading` 管理の死コード → ファイル削除（`useGetHospitalizations` React Query フックが既存） | Medium | `[x]` |
| 32 | `utils/constants/status-colors.ts:45-57` | `design-tokens` | `VISIT_TYPE_COLORS` の badge 系プロパティに `#D3E5EF`/`#183B56`/`#B8D4E3`/`#F7F6F3`/`#37352F`/`rgba(55,53,47,0.09)` ハードコード 6箇所 → `C.bgAccentLight60`/`textAccentDark90`/`borderAccentBadge50`/`bgPage60`/`text90`/`borderLight50` 新トークン 6 個追加して修正 | Medium | `[x]` |
| 33 | `features/inventory/routes/InventoryForm.tsx:59,133,207` | `design-tokens` | `border-[rgba(55,53,47,0.16)]` ハードコード 3箇所 → `C.borderMedium` に修正 | Medium | `[x]` |

---

## 修正完了サマリ

> 修正完了した違反の要約（PR/commit 参照）

| 日付 | 対象ドメイン | 修正ルール | 修正内容 | commit |
|------|------------|-----------|---------|--------|
| 2026-04-01 | shared/Layout | `rendering-conditional-render` | Sidebar.tsx の `&&` 条件レンダー3箇所を三項演算子に修正 | - |
| 2026-04-01 | 全 feature + shared | `bundle-feature-indexing` | auth/pets/owners/master への deep import 28箇所を index.ts 経由に修正。3つの index.ts にエクスポート追加 | - |
| 2026-04-01 | 全 feature + shared (72ファイル) | `design-tokens` | `#37352F` hex ハードコード339箇所を `C` / `PALETTE` 定数に一括修正。3つの新トークン追加 | - |
| 2026-04-02 | hospitalization + hospital-settings | `rerender-memo` | HospitalizationBasicInfo/NoteCard/DailyRecordTimeline/PropertyRow に memo() 追加 | - |
| 2026-04-02 | shared/NotionFilter | `design-tokens` | FilterRuleRow/SortPopover の `#F1F1EF`/`#E8E7E4` hex → C 定数。`hoverBgMutedBadge` トークン追加 | - |
| 2026-04-02 | medical-records → master | `bundle-feature-indexing` | 4箇所の cross-feature deep import → index.ts 経由に修正。master/index.ts にエクスポート追加 | - |
| 2026-04-02 | shared/RequirePermission | `bundle-feature-indexing` | auth/types → @/features/auth に修正 | - |
| 2026-04-02 | master + VitalsGraph | `design-tokens` | `#6B7280`/`#3B82F6`/`#e5e7eb`/チャート色 18箇所 → PALETTE 定数に修正。7つの新トークン追加 | - |
| 2026-04-02 | 全42ファイル (78箇所) | `design-tokens` | `#2EAADC`/`#F7F6F3`/`#2383E2`/`#038B94` 等の残存 hex 全量修正。19個の新 C トークン追加（medical-blue, brand, accent-focus, data-checked, bgMuted） | - |
| 2026-04-02 | medical-records | `rendering-conditional-render` | VaccinationHistory.tsx の `&&` → 三項演算子 | - |
| 2026-04-02 | hospitalization | `async-parallel` | ケージ移動の直列 await → Promise.all 並列化（2ファイル） | - |
| 2026-04-02 | medical-records + hospitalization | `rendering-hoist-jsx` | MedicalRecordForm/DailyDateNav/CarePlanTab の静的配列をモジュール定数に巻き上げ | - |
| 2026-04-02 | hospitalization | `rerender-transitions` | 死コード use-hospitalizations.ts 削除（React Query フック既存） | - |
| 2026-04-02 | status-colors + inventory | `design-tokens` | status-colors.ts の VISIT_TYPE_COLORS badge hex 6箇所 + InventoryForm.tsx の rgba 3箇所 → C トークンに修正。新トークン 6 個追加（text90, bgPage60, bgAccentLight60, textAccentDark90, borderAccentBadge50, borderLight50） | - |
