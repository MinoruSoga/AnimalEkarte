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

## ディレクトリ構造（bulletproof-react）

```
src/
├── main.tsx                # Viteエントリーポイント（ReactDOM.createRoot）
│
├── app/                    # アプリケーション層
│   ├── index.tsx           # Appコンポーネント
│   ├── provider.tsx        # プロバイダー統合（QueryClient, Toaster等）
│   └── router.tsx          # createBrowserRouter（lazy loading）
│
├── components/             # 共有コンポーネント
│   ├── ui/                 # shadcn/ui（変更禁止）
│   └── shared/             # 共有UI
│       ├── Layout/         # MainLayout, Header, Sidebar
│       ├── DataTable/      # 汎用テーブル、Pagination、Filters
│       ├── Form/           # FormField, FormError, SubmitButton
│       ├── Feedback/       # Spinner, ErrorFallback, LoadingOverlay, EmptyState
│       ├── Navigation/     # Breadcrumb, NavLink
│       ├── StatusBadge/    # ステータスバッジ
│       ├── DateRangePicker/# 日付範囲選択
│       ├── SearchBox/      # 検索ボックス
│       └── ConfirmDialog/  # 確認ダイアログ
│
├── features/               # 機能別モジュール
│   └── [feature-name]/
│       ├── api/            # データフェッチング（React Query）
│       │                   # ※ useXxx(), useCreateXxx() 等のQuery/Mutation hooks
│       ├── components/     # feature固有コンポーネント
│       │   └── XxxCard/
│       │       ├── XxxCard.tsx
│       │       ├── XxxCard.test.tsx  # テストは同階層に配置
│       │       └── index.ts
│       ├── hooks/          # ビジネスロジック・UI状態のみ
│       │                   # ※ useXxxForm(), useXxxFilters() 等
│       ├── types/          # feature固有型定義
│       ├── routes/         # ページコンポーネント
│       ├── utils/          # feature固有ユーティリティ
│       └── index.ts        # 公開API（明示的export）
│
├── hooks/                  # 共有hooks
│   ├── useDebounce.ts
│   ├── useDisclosure.ts
│   ├── useLocalStorage.ts
│   ├── usePagination.ts
│   └── useToast.ts
│
├── lib/                    # 外部ライブラリ設定
│   ├── axios.ts            # Axios設定（バックエンド接続）
│   ├── react-query.ts      # TanStack Query設定
│   ├── zod.ts              # Zodスキーマヘルパー
│   └── utils.ts            # cn()等
│
├── utils/                  # 純粋ユーティリティ関数
│   ├── format/             # date.ts, currency.ts
│   ├── validation/         # schemas.ts, validators.ts
│   └── constants/          # status.ts, routes.ts
│
├── config/                 # アプリケーション設定
│   ├── constants.ts
│   └── env.ts
│
├── stores/                 # グローバル状態（Zustand）
│   ├── authStore.ts
│   ├── themeStore.ts
│   └── sidebarStore.ts
│
├── types/                  # 共有型定義
│   ├── api.ts              # API共通型
│   └── common.ts           # 共通型
│
├── styles/                 # グローバルスタイル
│   └── globals.css         # Tailwind CSS v4
│
└── testing/                # テスト設定
    ├── setup.ts
    ├── server/             # MSW handlers
    └── utils.tsx
```

---

## アーキテクチャルール

### コードフロー（単方向）

```
shared (components/, hooks/, lib/, utils/, config/, stores/, types/)
    ↓
features/
    ↓
app/
```

### ルール

| ルール | 説明 |
|--------|------|
| Feature間import禁止 | `features/A` から `features/B` を直接importしない |
| app層で合成 | feature横断ロジックは `app/` で組み合わせる |
| `export *` 禁止 | `export * from "./xxx"` はtree-shaking阻害。明示的named exportは可 |
| 絶対パスimport | `@/` エイリアスを使用 |
| api/にデータフェッチ | React Query hooks (useQuery, useMutation) は api/ に配置 |
| hooks/にロジック | フォーム状態、フィルタ、バリデーション等は hooks/ に配置 |

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
    ↓ make codegen
src/types/generated/models.ts   ← 自動生成・直接編集禁止
    ↓ Omit / Partial / ReturnType
features/xxx/api/types.ts       ← APIリクエスト型（models.tsから導出）
src/lib/transforms/xxx.ts       ← ドメイン型（ReturnType<typeof transform>）
features/xxx/types/index.ts     ← フォームデータ型（UI専用）
```

| 型の種類 | 配置場所 |
|---------|---------|
| バックエンドモデル型 | `src/types/generated/models.ts`（自動生成・編集禁止） |
| APIリクエスト型 | `src/types/xxx.ts`（`Omit`/`Partial`で導出） |
| フロントエンドドメイン型 | `src/lib/transforms/xxx.ts`（`ReturnType`で導出） |
| フォームデータ型 | `src/types/xxx.ts`（UI専用・手書きOK） |

**原則**: 型定義はすべて `src/types/` に一元化する。`features/xxx/types/` は re-export のみ許容。

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
