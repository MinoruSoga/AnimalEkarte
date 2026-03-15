# Frontend - React 19 / TypeScript 5.7（bulletproof-react準拠）

## ⚠️ コマンド実行ルール

**npmコマンドはローカル実行禁止。必ずDocker経由で実行する。**

```bash
# ❌ NG
npm run build

# ✅ OK
docker compose exec frontend npm run build
```

## コマンド一覧

| タスク | コマンド |
|--------|---------|
| ビルド | `docker compose exec frontend npm run build` |
| Lint | `docker compose exec frontend npm run lint` |
| テスト | `docker compose exec frontend npm run test:run` |
| テスト(watch) | `docker compose exec frontend npm test` |
| カバレッジ | `docker compose exec frontend npm run test:coverage` |

---

## 技術スタック

| 技術 | バージョン |
|------|-----------|
| React | 19 |
| TypeScript | 5.7 |
| Vite | 6 |
| Tailwind CSS | 4 |
| shadcn/ui | Radix UI |
| React Router | 7 (Data Mode) |
| TanStack Query | v5 |
| Axios | HTTP Client |
| Vitest | Testing Library |
| Zustand | 状態管理 |

---

## ディレクトリ構造（bulletproof-react準拠・プロジェクト固有拡張あり）

```
src/
├── main.tsx                # Viteエントリーポイント（ReactDOM.createRoot）
├── index.css               # グローバルスタイル読み込み
├── vite-env.d.ts           # Vite型定義
│
├── app/                    # アプリケーション層
│   ├── index.tsx           # Appコンポーネント（AppProvider → RouterProvider）
│   ├── provider.tsx        # QueryClientProvider + Toaster（AuthProviderはここに置かない）
│   ├── router.tsx          # createBrowserRouter（inline lazy パターン）
│   └── pages/             # ★ Cross-feature合成ページ（複数featureを跨ぐ場合のみ）
│       └── OwnerFormPage.tsx  # owners + pets を合成する例
│
├── assets/                 # 静的アセット（画像・アイコン等）
│
├── components/             # 共有コンポーネント
│   ├── ui/                 # shadcn/ui（Radix UI Primitives）★変更禁止
│   ├── errors/             # RouteErrorBoundary, RootErrorBoundary
│   └── shared/             # アプリ固有共有UI
│       ├── Layout/             # Layout（Sidebar統合済み）
│       ├── DataTable/          # DataTableRow, SortableDataTableRow
│       ├── SidePeek/           # SidePeekPanel, Body, Footer, TitleInput, Toolbar
│       ├── Form/               # FormHeader, PrimaryButton
│       ├── FormFieldError/     # フィールドエラー表示
│       ├── Feedback/           # ImageWithFallback
│       ├── ConfirmDialog/      # 確認ダイアログ
│       ├── DateRangePicker/    # 日付範囲選択
│       ├── NotionDatePicker/   # 日付ピッカー
│       ├── HistoryFilterPanel/ # 履歴フィルタUI
│       ├── MasterSelectModal/  # マスタ選択モーダル
│       ├── NavigationBlocker/  # 未保存変更時のナビゲーションブロック
│       ├── PageLayout/         # ページコンテナ
│       ├── Pagination/         # ページネーション
│       ├── PatientInfoCard/    # 患者情報カード
│       ├── PetSelection/       # ペット選択UI（PetSelection, SearchForm, ResultsTable）
│       ├── ReservationFormModal/ # 予約フォームモーダル
│       ├── RowActionButton/    # 行アクションボタン
│       ├── RowActionDropdown/  # 行アクションドロップダウン
│       ├── SearchBox/          # 検索ボックス
│       ├── SearchFilterBar/    # 検索+フィルタバー
│       ├── SortableHeader/     # ソート可能カラムヘッダ
│       ├── StatusBadge/        # ステータスバッジ
│       ├── StatusPill/         # ステータスピル
│       └── TreatmentSearchDialog/ # 処置検索ダイアログ
│
├── features/               # 機能別モジュール（18 features）
│   ├── auth/               # 認証（ログイン・セッション管理）
│   ├── dashboard/          # ダッシュボード
│   ├── owners/             # ★ ベストプラクティス参照実装
│   ├── pets/               # ペット（CRUD API のみ）
│   ├── reservations/       # 予約管理
│   ├── medical-records/    # 電子カルテ
│   ├── hospitalization/    # 入院管理
│   ├── examinations/       # 診察
│   ├── accounting/         # 会計
│   ├── vaccinations/       # ワクチン
│   ├── trimming/           # トリミング
│   ├── inventory/          # 在庫管理
│   ├── estimates/          # 見積
│   ├── shifts/             # シフト管理
│   ├── master/             # マスタ設定（PATTERNS.md 参照）
│   └── hospital-settings/  # 病院設定（クリニックマスタ）
│
│   └── [feature-name]/     # 各featureの構造
│       ├── api/            # React Query hooks（get/create/update/delete）
│       │   ├── get-xxx.ts      # useQuery + queryOptions factory
│       │   ├── create-xxx.ts   # useMutation + Zod schema
│       │   ├── update-xxx.ts
│       │   ├── delete-xxx.ts
│       │   ├── types.ts        # APIリクエスト/レスポンス型
│       │   ├── transforms.ts   # Backend ↔ Frontend 変換
│       │   └── index.ts
│       ├── components/     # feature固有UI
│       ├── hooks/          # useXxxForm, useXxxFilters 等
│       ├── routes/         # ページコンポーネント（★ app/routes/ でなくここに置く）
│       ├── types/          # feature固有型
│       ├── loaders.ts      # React Router loader（必要な場合のみ）
│       └── index.ts        # 公開API（明示的named export）
│
├── hooks/                  # 共有カスタムフック（全feature横断）
│   ├── use-master-items.ts         # マスタデータ取得
│   ├── use-mobile.ts               # モバイル判定
│   ├── use-pet-selection.ts        # ペット選択
│   ├── use-pet-selection-page.ts   # ペット選択ページロジック
│   ├── use-pet.ts                  # 単体ペット取得
│   ├── use-service-type-color-map.ts
│   ├── usePagination.ts            # ページネーション状態
│   ├── useReducedMotion.ts         # アクセシビリティ
│   ├── useSortableList.ts          # ソータブルリスト
│   ├── useStaffValidation.ts       # スタッフ入力検証
│   ├── useTableSort.ts             # テーブルソート状態
│   └── useUnsavedChanges.ts        # 未保存変更警告
│
├── lib/                    # ライブラリ設定・ユーティリティ
│   ├── axios.ts            # Axiosインスタンス（baseURL, interceptors）
│   ├── react-query.ts      # QueryClient設定（staleTime階層）
│   ├── zod.ts              # Zodスキーマヘルパー
│   ├── utils.ts            # cn() 等
│   ├── design-tokens.ts    # デザイントークン
│   ├── handle-api-error.ts # APIエラーハンドリング共通処理
│   ├── type-utils.ts       # TypeScriptユーティリティ型
│   └── transforms/         # 型変換ヘルパー（Backend型 → Frontend型）
│       ├── pet.ts
│       ├── medicine.ts
│       └── treatment.ts
│
├── config/                 # アプリケーション設定
│   └── paths.ts            # 全ルートの型安全URLマップ（getHref() 付き）
│
├── stores/                 # グローバル状態（Zustand）
│   └── sidebar-store.ts    # サイドバー開閉状態のみ
│
├── types/                  # 共有型定義
│   ├── generated/          # ★ 自動生成（直接編集禁止）
│   │   └── models.ts       # make codegen（tygo）で生成
│   ├── diagnosis.ts
│   ├── medicine.ts
│   ├── owner.ts
│   ├── pet.ts
│   ├── service-type.ts
│   ├── treatment.ts
│   ├── trimming.ts
│   └── index.ts
│
├── utils/                  # 純粋ユーティリティ関数
│   ├── format/             # date.ts, number.ts
│   ├── constants/          # 定数
│   ├── validation/         # バリデーション
│   └── status-helpers.ts   # ステータス変換ヘルパー
│
├── styles/                 # グローバルスタイル
│   └── globals.css         # Tailwind CSS v4
│
└── testing/                # テスト設定
    ├── setup.ts
    └── server/             # MSW handlers
```

---

## アーキテクチャルール

### コードフロー（単方向）

```
shared (components/, hooks/, lib/, utils/, config/, stores/, types/)
    ↓
features/
    ↓
app/          ← ルーター定義 + cross-feature合成（app/pages/）
```

### ルール

| ルール | 説明 |
|--------|------|
| Feature間import禁止 | `features/A` から `features/B` を直接importしない |
| cross-feature合成は `app/pages/` | 複数featureのAPIを組み合わせる必要がある場合は `app/pages/XxxPage.tsx` を作成し、router.tsx から lazy import する |
| Routesは `features/` に置く | bulletproof-reactの `app/routes/` パターンは採用しない。各featureの `routes/` にページコンポーネントを配置する |
| `export *` 禁止 | `export * from "./xxx"` はtree-shaking阻害。明示的named exportは可 |
| 絶対パスimport | `@/` エイリアスを使用 |
| api/にデータフェッチ | React Query hooks (useQuery, useMutation) は api/ に配置 |
| hooks/にロジック | フォーム状態、フィルタ、バリデーション等は hooks/ に配置 |
| RouterでAuthProvider | AuthProviderはrouter.tsx内の保護ルートにのみ配置（provider.tsxに置かない）。/loginページで不要な GET /v1/me が発生するのを防ぐ |
| `config/paths.ts` でURL管理 | ハードコードされたURLパス文字列は禁止。`paths.owners.list.getHref()` 等を使用する |

### Routerパターン（プロジェクト固有）

bulletproof-reactの `clientLoader`/`clientAction`/`convert()` パターンは使用しない。以下の inline lazy パターンを使う：

```typescript
// ✅ このプロジェクトの標準パターン
{
  path: "/owners",
  lazy: async () => {
    const [{ OwnersList }, { ownersLoader }] = await Promise.all([
      import("@/features/owners/routes/OwnersList"),
      import("@/features/owners/loaders"),
    ]);
    return { Component: OwnersList, loader: ownersLoader };
  },
},

// cross-feature合成が必要な場合は app/pages/ を使う
{
  path: "new",
  lazy: async () => {
    const { OwnerFormPage } = await import("@/app/pages/OwnerFormPage");
    return { Component: OwnerFormPage };
  },
},
```

---

## React 19 コーディングルール

> コード例の詳細は [`CODING_RULES.md` Section 2](./CODING_RULES.md) を参照。
> コンポーネント定義・新hooks・フォームパターンの実装例がすべて記載されている。

**最重要禁止事項**（詳細は下記「禁止事項」テーブルにも記載）:
- `FC` / `React.FC` → `export function Xxx(props: XxxProps)` に置き換える
- `forwardRef` → ref を通常の prop として受け取る（React 19 では不要）

---

## TypeScriptルール

### Import順序

```typescript
// 1. React/Framework
import { useState, useEffect } from "react";
import { useNavigate } from "react-router";

// 2. 外部ライブラリ
import { format } from "date-fns";

// 3. 共有モジュール (@/)
import { Button } from "@/components/ui/button";

// 4. feature内部（相対パス）
import { OwnerCard } from "../components/OwnerCard";

// 5. 型（type keyword付き）
import type { Owner } from "@/types";
```

### 命名規則

| 対象 | 規則 | 例 |
|------|------|-----|
| コンポーネント | PascalCase | `PatientCard` |
| 関数・変数 | camelCase | `getPatientById` |
| 定数 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| ファイル | kebab-case | `patient-card.tsx` |
| 型・Interface | PascalCase | `Patient` |
| hooks | use + camelCase | `usePatientForm` |

### バックエンド型の扱い

> 詳細は `CODING_RULES.md` **Section 3.6** を参照。

```
backend/internal/model/*.go
    ↓ make codegen（tygo）
src/types/generated/models.ts   ← 自動生成・直接編集禁止
    ↓ Omit / Partial / ReturnType
features/xxx/api/types.ts       ← APIリクエスト型（models.tsから導出）
src/lib/transforms/xxx.ts       ← ドメイン型変換（ReturnType<typeof transform>で型推論）
src/types/xxx.ts                ← 共有ドメイン型（複数featureで使う場合）
features/xxx/types/index.ts     ← feature固有型（UI専用・手書きOK）
```

| 型の種類 | 配置場所 |
|---------|---------|
| バックエンドモデル型 | `src/types/generated/models.ts`（自動生成・編集禁止） |
| APIリクエスト/レスポンス型 | `features/xxx/api/types.ts`（`Omit`/`Partial`で導出） |
| フロントエンドドメイン型 | `src/lib/transforms/xxx.ts`（`ReturnType<typeof transform>`で推論） |
| 共有ドメイン型 | `src/types/xxx.ts`（複数featureで共有するもの） |
| フォームデータ型 | `features/xxx/types/index.ts`（UI専用・手書きOK） |

**禁止**: `interface CreateXxxRequest { ... }` の手書き → `models.ts` からの導出に統一

---

## 禁止事項

| 禁止 | 理由 | 代替 |
|------|------|------|
| `any` 型 | 型安全性の破壊 | `unknown` + 型ガード |
| `FC` / `React.FC` | React 19では不要 | 関数宣言 |
| `forwardRef` | React 19では不要 | ref as prop |
| feature間import | アーキテクチャ違反 | app層で合成 |
| `export *` | tree-shaking阻害 | 明示的named export |
| `console.log` 放置 | 本番コード汚染 | 削除 |
| default export | IDE補完が弱い | 名前付きexport |
| `&&` 条件レンダー | 0/空文字が漏れる | `? ... : null` |
| barrel index 経由 import | tree-shaking 阻害 | 直接ファイル import（feature 外からも同様） |
| `generated/models.ts` 直接編集 | `make codegen` で上書きされる | Goモデルを修正して `make codegen` |
| APIリクエスト型を `interface` で手書き | Goモデルとの乖離 | `models.ts` から `Omit`/`Partial` で導出 |
| コメントのみ・空の `index.ts` を残す | 死ファイル、混乱の原因 | 削除する |
| 実ファイルがあるフォルダの `.gitkeep` を残す | 不要 | 削除する |

---

## パフォーマンスパターン（Vercel React Best Practices準拠）

> **参照実装**: `features/owners/` の実装がすべてのパターンのベストプラクティス。
> 詳細コード例は `frontend/CODING_RULES.md` Section 12 を参照。

### 必須パターン一覧

| パターン | ルール | 実装場所 |
|---------|--------|---------|
| 独立したフォームセクションを `memo()` で分断 | `rerender-memo` | `OwnerForm.tsx` — `OwnerInfoSection`, `PetTableRow`, `MembershipTypeButtons` |
| `memo()` に渡すハンドラは `useCallback` で安定化 | `rerender-functional-setstate` | `OwnerForm.tsx` — `handleInputChange`, `handleDeletePetRequest` |
| `useState` 初期値に高コスト関数は lazy init | `rerender-lazy-state-init` | `useOwnerForm.ts` — `useState(() => ...)` |
| 検索フィルタは `useDeferredValue` で遅延 | `rerender-transitions` | `OwnersList.tsx` — `useDeferredValue(searchTerm)` |
| API ミューテーションは `useTransition` | `rerender-transitions` | `useOwnerForm.ts` — `startSaveTransition` |
| 重いモーダルは `lazy()` + `Suspense` | `bundle-dynamic-imports` | `OwnerForm.tsx` — `const PetEditModal = lazy(...)` |
| 静的 JSX はモジュール定数に巻き上げ | `rendering-hoist-jsx` | `OwnerForm.tsx` — `PET_TABLE_HEADER`; `PetEditModal.tsx` — `GENDER_SELECT_ITEMS` |
| API 由来の JSX リストは `useMemo` | `js-cache-function-results` | `PetEditModal.tsx` — `animalSpeciesSelectItems` |
| 条件レンダーは必ず `? (...) : null` | `rendering-conditional-render` | `OwnersList.tsx` — `{pet.status ? ... : null}` |
| loader で独立フェッチは `Promise.all` | `async-parallel` | `loaders.ts` — `ownersLoader` |
| `useCallback` deps にはオブジェクトでなく primitive | `rerender-dependencies` | `OwnersList.tsx` — `pendingDeleteOwnerId` |

---

## ESLint

- **エラー**: 0件を維持
- **Warning**: 6件（shadcn/ui由来、許容）

```bash
docker compose exec frontend npm run lint
```

---

## 参照

| ドキュメント | 説明 |
|-------------|------|
| [詳細コーディング規約](./CODING_RULES.md) | 完全なルール集 |
| [プロジェクト全体](../CODING_RULES.md) | 共通ルール |
| [React 19](https://react.dev/blog/2024/12/05/react-19) | 公式リリースノート |
| [bulletproof-react](https://github.com/alan2207/bulletproof-react) | アーキテクチャ参照 |
