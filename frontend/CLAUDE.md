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

### コンポーネント定義

```typescript
// ✅ 正しい: 関数宣言 + 明示的Props型
interface PatientCardProps {
  patient: Patient;
  onSelect?: (id: string) => void;
  ref?: React.Ref<HTMLDivElement>;  // React 19: ref as prop
}

export function PatientCard({ patient, onSelect, ref }: PatientCardProps) {
  return <div ref={ref}>...</div>;
}

// ❌ 禁止
export const PatientCard: FC<Props> = () => {};        // FC禁止
export const PatientCard = forwardRef(() => {});       // forwardRef禁止
```

### React 19 新hooks

```typescript
// useActionState: フォームアクション管理
import { useActionState } from "react";
const [state, formAction, isPending] = useActionState(
  async (prevState, formData: FormData) => {
    // 処理
    return { success: true, errors: null };
  },
  { success: false, errors: null }
);

// useOptimistic: 楽観的UI更新
import { useOptimistic } from "react";
const [optimisticItems, addOptimisticItem] = useOptimistic(
  items,
  (state, newItem) => [...state, newItem]
);

// use(): Promise/Context直接読み取り
import { use } from "react";
const data = use(fetchPromise);  // Suspense必須
const theme = use(ThemeContext);

// useFormStatus: フォーム送信状態（react-dom）
import { useFormStatus } from "react-dom";
const { pending } = useFormStatus();
```

### フォーム送信ボタン例

```typescript
export function SubmitButton({ children }: { children: React.ReactNode }) {
  const { pending } = useFormStatus();
  return (
    <Button type="submit" disabled={pending}>
      {pending ? "処理中..." : children}
    </Button>
  );
}
```

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
