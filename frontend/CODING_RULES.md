# Frontend Coding Rules

## 概要

React 19 + TypeScript 6.0 + Vite 8 + Tailwind CSS 4 のフロントエンド開発規約。
Bulletproof-Reactアーキテクチャに準拠する。

> **正本（SSOT）**: 本ファイルは人間向けの詳細規約である。矛盾が生じた場合の優先順位は
> **① 実コード → ② `frontend/CLAUDE.md`・ネスト `CLAUDE.md`（エージェント正本） → ③ 本ファイル**。
> 本ファイルを編集する際は、必ず実コードと `frontend/CLAUDE.md` に照合して整合させること。

---

## 1. アーキテクチャ

### 1.1 ディレクトリ構成

```
src/
├── main.tsx                               # Viteエントリーポイント（ReactDOM.createRoot）
├── index.css                              # グローバルスタイル読み込み
├── vite-env.d.ts                          # Vite型定義
│
├── app/                                   # アプリケーション層
│   ├── index.tsx                          # Appコンポーネント（AppProvider → RouterProvider）
│   ├── provider.tsx                       # QueryClientProvider + Toaster
│   ├── router.tsx                         # createBrowserRouter（inline lazy パターン）
│   └── pages/                             # ★ Cross-feature合成ページ
│       └── OwnerFormPage.tsx              # owners + pets を合成する例
│
├── assets/                                # 静的アセット
│
├── features/                              # 機能別モジュール（28 features、いずれも index.ts あり）
│   ├── accounting/                        # 会計
│   ├── accounting-reports/                # 月次・帳票
│   ├── aggregation/                       # 集計
│   ├── auth/                              # 認証（ログイン・セッション管理）
│   ├── cash-register/                     # レジ締め
│   ├── checkups/                          # 健診
│   ├── clinic-settings/                   # クリニック設定
│   ├── closing-settings/                  # 締め設定
│   ├── estimates/                         # 見積
│   ├── examinations/                      # 診察
│   ├── hospitalization/                   # 入院管理
│   ├── identity-links/                    # 本人同定リンク
│   ├── inventory/                         # 在庫管理
│   ├── lab-device/                        # 検査機器
│   ├── line-reservation/                  # LINE予約
│   ├── lstep/                             # Lステップ連携
│   ├── manual/                            # マニュアル
│   ├── master/                            # マスタ設定（PATTERNS.md 参照）
│   ├── medical-records/                   # 電子カルテ
│   ├── owner-report/                      # 飼主レポート
│   ├── owners/                            # ★ ベストプラクティス参照実装
│   ├── pets/                              # ペット（CRUD API のみ）
│   ├── reception/                         # 当日の受付
│   ├── reservations/                      # 予約管理
│   ├── settings/                          # 設定
│   ├── shifts/                            # シフト管理
│   ├── trimming/                          # トリミング
│   └── vaccinations/                      # ワクチン
│
├── components/                            # 共有コンポーネント
│   ├── ui/                                # shadcn/ui（Radix UI Primitives）★変更禁止
│   ├── errors/                            # RouteErrorBoundary, RootErrorBoundary
│   └── shared/                            # アプリケーション固有の共有コンポーネント
│       ├── Layout/                        # Layout（Sidebar統合済み）
│       ├── DataTable/                     # DataTableRow, SortableDataTableRow
│       ├── SidePeek/                      # Panel, Body, Footer, TitleInput, Toolbar
│       ├── Form/                          # FormHeader, PrimaryButton
│       ├── FormFieldError/                # フィールドエラー表示
│       ├── Feedback/                      # ImageWithFallback
│       ├── ConfirmDialog/                 # 確認ダイアログ
│       ├── DateRangePicker/               # 日付範囲選択
│       ├── DatePicker/                    # 日付ピッカー
│       ├── HistoryFilterPanel/            # 履歴フィルタUI
│       ├── MasterSelectModal/             # マスタ選択モーダル
│       ├── NavigationBlocker/             # 未保存変更時のナビゲーションブロック
│       ├── PageLayout/                    # ページコンテナ
│       ├── Pagination/                    # ページネーション
│       ├── PatientInfoCard/               # 患者情報カード
│       ├── PetSelection/                  # ペット選択UI
│       ├── ReservationFormModal/          # 予約フォームモーダル
│       ├── RowActionButton/               # 行アクションボタン
│       ├── RowActionDropdown/             # 行アクションドロップダウン
│       ├── SearchBox/                     # 検索ボックス
│       ├── PropertyFilter/                # Notion風フィルタ・検索・ソートUI
│       ├── SortableHeader/                # ソート可能カラムヘッダ
│       ├── StatusBadge/                   # ステータスバッジ
│       ├── StatusPill/                    # ステータスピル
│       └── TreatmentSearchDialog/         # 処置検索ダイアログ
│
├── hooks/                                 # 共有カスタムフック（全feature横断）
│   ├── use-master-items.ts               # マスタデータ取得
│   ├── use-mobile.ts                     # モバイル判定
│   ├── use-pet-selection.ts              # ペット選択
│   ├── use-pet-selection-page.ts         # ペット選択ページロジック
│   ├── use-pet.ts                        # 単体ペット取得
│   ├── use-service-type-color-map.ts     # サービス種別→色マップ
│   ├── use-pagination.ts                 # ページネーション状態
│   ├── use-reduced-motion.ts             # アクセシビリティ
│   ├── use-sortable-list.ts              # ソータブルリスト
│   ├── use-staff-validation.ts           # スタッフ入力検証
│   ├── useTableSort.ts                   # テーブルソート状態
│   └── use-unsaved-changes.ts            # 未保存変更警告
│
├── lib/                                   # ライブラリ設定・共有ヘルパ（utils/ は廃止済み。新設禁止）
│   ├── axios.ts                           # Axiosインスタンス（baseURL, interceptors）
│   ├── react-query.ts                     # QueryClient設定（staleTime階層）
│   ├── query-keys.ts                      # React Query キーファクトリー
│   ├── zod.ts                             # Zodスキーマヘルパー
│   ├── utils.ts                           # cn() 等
│   ├── design-tokens.ts                   # デザイントークン
│   ├── handle-api-error.ts                # APIエラーハンドリング共通処理
│   ├── jst-date.ts                        # JST 日付ヘルパ
│   ├── type-utils.ts                      # TypeScriptユーティリティ型
│   ├── format/                            # date.ts, number.ts（表示フォーマット）
│   └── transforms/                        # Backend型 → Frontend型 変換ヘルパー
│       ├── pet.ts
│       ├── medicine.ts
│       └── treatment.ts
│
├── constants/                             # 共有定数（src/constants/。lib/ や feature 内での新設禁止）
│   ├── status-colors.ts
│   ├── payment-method.ts
│   └── accounting-status.ts
│
├── config/                                # アプリケーション設定
│   └── paths.ts                           # 全ルートの型安全URLマップ（getHref()付き）
│
├── stores/                                # グローバル状態（Zustand）
│   └── sidebar-store.ts                   # サイドバー開閉状態のみ
│
├── types/                                 # 共有型定義
│   ├── generated/                         # ★ 自動生成（直接編集禁止）
│   │   └── models.ts                      # make codegen（tygo）で生成
│   ├── diagnosis.ts
│   ├── medicine.ts
│   ├── owner.ts
│   ├── pet.ts
│   ├── service-type.ts
│   ├── treatment.ts
│   ├── trimming.ts
│   └── index.ts
│
├── styles/                                # グローバルスタイル
│   └── globals.css                        # Tailwind CSS v4
│
└── testing/                               # テスト設定
    ├── setup.ts
    └── server/                            # MSW handlers
```

### 1.2 コードフローの方向（単方向依存）

```
┌──────────────────────────────────────────────────────┐
│                        app/                           │
│  router.tsx        … ルート定義（inline lazy パターン）│
│  provider.tsx      … QueryClient + Toaster            │
│  pages/XxxPage.tsx … cross-feature合成（必要時のみ）  │
└──────────────────────────────────────────────────────┘
                          ↑
                          │ import可能
                          │
┌──────────────────────────────────────────────────────┐
│                     features/                         │
│  routes/   … ページコンポーネント（app/routes/は使わない）│
│  api/      … React Query hooks                        │
│  hooks/    … フォーム・フィルタ等のUIロジック          │
│  （feature 専用ヘルパは feature 内に置く）              │
│                                                        │
│  ※ feature間の直接importは禁止                        │
└──────────────────────────────────────────────────────┘
                          ↑
                          │ import可能
                          │
┌──────────────────────────────────────────────────────┐
│     shared (components/, hooks/, lib/, config/,       │
│             stores/, types/, constants/)              │
│     ※ 共有ヘルパは app 層の lib/。utils/ は廃止済み     │
└──────────────────────────────────────────────────────┘
```

**ルール:**
- `shared → features → app` の単方向のみ
- feature間の直接importは禁止（app/pages/ で合成する、または `components/shared` に移動する）
- 循環参照は絶対禁止
- cross-feature合成が必要な場合は `app/pages/XxxPage.tsx` を作成し、router.tsx から lazy import する

**bulletproof-react参照実装との意図的な差分:**

| bulletproof-react | このプロジェクト | 理由 |
|------------------|----------------|------|
| routes を `app/routes/` に配置 | routes を `features/[feature]/routes/` に配置 | feature内に完結させる方が保守性が高い |
| `clientLoader`/`clientAction`/`convert()` パターン | inline `lazy()` + 直接 `loader` 設定 | シンプルで読みやすい |
| AuthLoader を AppProvider 内に配置 | AuthProvider を router.tsx root route element にアプリ全体（/login 含む）配置 | /login 401 は route 分離でなく AuthProvider 内部の password-recovery 経路のみ restore skip（BUG-031）で防ぐ |
| `config/env.ts`（Zod検証） | なし（Vite env をそのまま使用） | 現時点では不要 |
| cross-feature合成なし（features は独立） | `app/pages/` で cross-feature 合成 | 複数 feature を使うページ専用の合成層 |

**cross-feature合成パターン（props注入による依存逆転）:**

複数の feature を組み合わせるページは `app/pages/XxxPage.tsx` で合成する。
feature コンポーネントは外部依存を props として受け取り、`app/pages/` で実装を注入する。

```
パターン: feature コンポーネントが props 型を定義 → app/pages/ が実装を注入
```

```typescript
// ✅ features/owners/routes/OwnerForm.tsx
// pets への直接依存を持たず、props 経由で受け取る（依存逆転）
interface OwnerFormProps {
  petMutations: PetMutations; // types/pet.ts に定義
}
export function OwnerForm({ petMutations }: OwnerFormProps) { ... }

// ✅ app/pages/OwnerFormPage.tsx
// app 層なので両 feature を import 可能。実装を注入する。
// feature 外からは必ず barrel（index.ts）経由 — deep import 禁止（Feature Indexing）
import { OwnerForm } from "@/features/owners";
import { createPet, useCreatePet, useUpdatePet, useDeletePet } from "@/features/pets";

export function OwnerFormPage() {
  const { mutate: createPetMutate } = useCreatePet();
  const { mutate: updatePetMutate } = useUpdatePet();
  const { mutate: deletePetMutate } = useDeletePet();

  const petMutations: PetMutations = { createPetFn: createPet, createPetMutate, ... };
  return <OwnerForm petMutations={petMutations} />;
}

// ✅ router.tsx からは app/pages/ を参照
{
  path: "new",
  lazy: async () => {
    const { OwnerFormPage } = await import("@/app/pages/OwnerFormPage");
    return { Component: OwnerFormPage };
  },
},

// ❌ NG: feature 内から別 feature を直接 import
// features/owners/routes/OwnerForm.tsx で useCreatePet() を直接 import するのは禁止
```

**判断基準:**

| ケース | 配置 |
|--------|------|
| 単一 feature のみ使うページ | `features/[feature]/routes/` から直接 import |
| 複数 feature を組み合わせるページ | `app/pages/XxxPage.tsx` で合成 |

### 1.3 Feature モジュール構成

各featureは以下の構造を持つ：

```
features/[feature-name]/
├── api/                        # データフェッチング（React Query hooks）
│   ├── get-[entity].ts         # 単体取得: raw fetch fn + queryOptions factory + useQuery hook
│   ├── get-[entity]s.ts        # 一覧取得: raw fetch fn + queryOptions factory + useQuery hook
│   ├── create-[entity].ts      # 作成: Zod schema + useMutation hook
│   ├── update-[entity].ts      # 更新: Zod schema + useMutation hook
│   ├── delete-[entity].ts      # 削除: useMutation hook
│   ├── types.ts                # APIリクエスト/レスポンス型（models.tsからOmit/Partialで導出）
│   ├── transforms.ts           # Backend型 → Frontend型 変換
│   └── index.ts                # 明示的named export
├── components/                 # feature固有コンポーネント
│   ├── [Component]/
│   │   ├── [Component].tsx
│   │   ├── [SubComponent].tsx
│   │   └── index.ts
│   └── ...
├── hooks/                      # ビジネスロジック・UI状態管理
│   ├── use[Entity]Form.ts      # フォーム状態（useActionStateで送信・pending管理）
│   └── ...
├── routes/                     # ★ ページコンポーネント（app/routes/ではなくここに配置）
│   ├── [Entity]List.tsx        # 一覧ページ（router から直接 import）
│   ├── [Entity]Form.tsx        # 作成/編集フォーム
│   │                           #   単一 feature → router から直接 import
│   │                           #   cross-feature必要 → props で受け取り app/pages/ から注入
│   ├── [Entity]Detail.tsx      # 詳細ページ
│   └── [Entity]PetSelection.tsx # ペット選択（必要な場合のみ）
├── types/                      # feature固有型定義
│   └── index.ts
├── loaders.ts                  # React Router loader（必要な場合のみ、Promise.allで並列フェッチ）
└── index.ts                    # 公開API（明示的named export のみ）
```

#### Feature構成例: reception

```
features/reception/
├── api/
│   ├── get-reception.ts        # + useGetReception()
│   ├── update-appointment-status.ts  # + useUpdateAppointmentStatus()
│   └── transforms.ts           # カラム変換ロジック
├── hooks/
│   └── use-reception-kanban.ts # カンバン状態管理
├── routes/
│   └── Reception.tsx
└── index.ts
```

#### Feature構成例: owners（★ ベストプラクティス参照実装）

```
features/owners/
├── api/
│   ├── get-owners.ts           # axios直接呼び出し（loader用）
│   ├── get-owner.ts            # 単体取得
│   ├── get-animal-species.ts   # + useGetAnimalSpecies()（マスタ）
│   ├── get-insurances.ts       # + useGetInsurances()（マスタ）
│   ├── create-owner.ts         # axios直接呼び出し
│   ├── update-owner.ts         # axios直接呼び出し
│   ├── delete-owner.ts         # axios直接呼び出し
│   ├── transforms.ts           # backend ↔ frontend 型変換
│   └── index.ts
├── components/
│   ├── PetEditModal.tsx        # lazy()でロードされる重いモーダル
│   │                           # rendering-hoist-jsx: GENDER_SELECT_ITEMS 等の定数
│   │                           # js-cache-function-results: useMemo でJSXキャッシュ
│   └── index.ts
├── hooks/
│   ├── use-owner-form.ts       # フォーム状態・ペットCRUD・保存ロジック（useActionState）
│   │                           # rerender-lazy-state-init
│   └── index.ts
├── routes/
│   ├── OwnerForm.tsx           # feature コンポーネント（router から直接 import しない）
│   │                           # props: petMutations（依存逆転 - pets feature を直接 import しない）
│   │                           # rerender-memo: OwnerInfoSection, PetTableRow, MembershipTypeButtons
│   │                           # bundle-dynamic-imports: lazy(PetEditModal)
│   │                           # rendering-hoist-jsx: PET_TABLE_HEADER
│   ├── OwnersList.tsx          # 一覧ページ（router から直接 import）
│   │                           # rerender-transitions: useDeferredValue(searchTerm)
│   │                           # rendering-conditional-render: ? null パターン
│   └── index.ts
├── types/
│   └── index.ts
├── loaders.ts                  # React Router Data Mode ローダー
│                               # async-parallel: Promise.all で全ページ並列フェッチ
└── index.ts                    # 公開API
                                # export { OwnerForm, OwnersList } from "./routes/..."
                                # export { useOwnerForm } from "./hooks/..."
                                # export { PetEditModal } from "./components/..."
# ★ OwnerForm は app/pages/OwnerFormPage.tsx で pets と合成されてから router に登録される
# ★ OwnersList は単一 feature のため router から直接 import される
```

### 1.4 主要ファイル実装例

#### app/router.tsx（React Router Data Mode）

FE-RC-053/060: `AuthProvider` はアプリ全体（`/login` を含む root route の `element`）に配置する。
`/login` での不要な `GET /v1/me` は route 分離ではなく `AuthProvider` 内部のセッション復元処理
（password-recovery 経路のみ restore skip、BUG-031）で防ぐ。以下は概念を示す簡略例であり、
実際のルート定義本体は `app/routes/app-routes.tsx` 以下に分割されている。

```typescript
import { lazy, Suspense } from "react";
import { createBrowserRouter } from "react-router";

import { Layout } from "@/components/shared/Layout";
import { RootErrorBoundary, RouteErrorBoundary } from "@/components/errors/RouteErrorBoundary";
import { AuthProvider } from "@/features/auth/provider";

/* bundle-dynamic-imports: ログインページは未認証ユーザー専用。認証済みバンドルに含めない */
const Login = lazy(() =>
  import("@/features/auth").then((m) => ({ default: m.Login })),
);

export const router = createBrowserRouter([
  // ── AuthProvider をアプリ全体（/login 含む）に配置 ─────────────────
  {
    element: (
      <AuthProvider>
        <Outlet />
      </AuthProvider>
    ),
    errorElement: <RootErrorBoundary />,
    children: [
      // ── 未認証ルート（Layout なし） ───────────────────────────────
      {
        path: "/login",
        element: (
          <Suspense fallback={null}>
            <Login />
          </Suspense>
        ),
      },

      // ── 認証済みルート（Layout 配下） ─────────────────────────────
      {
    element: <Layout />,
    errorElement: <RootErrorBoundary />,
    children: [
      // ── Reception ────────────────────────────────────────────────
      {
        path: "/",
        lazy: async () => {
          const { Reception } = await import("@/features/reception");
          return { Component: Reception };
        },
      },

      // ── Owners ───────────────────────────────────────────────────
      // ★ feature 外（app/routes・app/pages）からの import は必ず barrel（@/features/owners）経由
      // ★ lazy 動的 import も feature 外からは barrel 経由（実測: app/routes 全件 barrel・deep import 0件）
      // ★ cross-feature 合成ページ（owners + pets 等）は app/pages/ から import
      {
        path: "/owners",
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const [{ OwnersListPage }, { ownersLoader }] = await Promise.all([
                import("@/app/pages/OwnersListPage"),
                import("@/features/owners"),        // ✅ barrel から loader を取得
              ]);
              return { Component: OwnersListPage, loader: ownersLoader };
            },
          },
          {
            path: "new",
            // cross-feature合成（owners + pets）→ app/pages/ を経由
            lazy: async () => {
              const { OwnerFormPage } = await import("@/app/pages/OwnerFormPage");
              return { Component: OwnerFormPage };
            },
          },
          {
            path: ":id",
            lazy: async () => {
              const [{ OwnerFormPage }, { ownerLoader }] = await Promise.all([
                import("@/app/pages/OwnerFormPage"),
                import("@/features/owners"),        // ✅ barrel から loader を取得
              ]);
              return { Component: OwnerFormPage, loader: ownerLoader };
            },
          },
        ],
      },

      // ── Medical Records ──────────────────────────────────────────
      {
        path: "/medical-records",
        errorElement: <RouteErrorBoundary />,
        children: [
          {
            index: true,
            lazy: async () => {
              const { MedicalRecords } = await import("@/features/medical-records");
              return { Component: MedicalRecords };
            },
          },
          // ... 他のサブルート
        ],
      },
      // ... 他のルート
    ],
  },
]);
```

**bulletproof-react との Router差分:**

| bulletproof-react | このプロジェクト |
|---|---|
| `convert(queryClient)` HOF でルートモジュールを変換 | `lazy: async () => { ... return { Component, loader } }` のインライン形式 |
| `clientLoader(queryClient)` をルートファイル内に定義 | `loaders.ts` を features/ 内の別ファイルに分離 |
| loader が `queryClient.getQueryData ?? fetchQuery` でキャッシュ統合 | loader が直接 axios 呼び出し → `useLoaderData()` で受け取る（React Query 非経由） |
| Default export でルートコンポーネントを export | Named export（`export function XxxList()`） |

#### app/index.tsx

```typescript
import { RouterProvider } from "react-router";
import { AppProvider } from "./provider";
import { router } from "./router";

export function App() {
  return (
    <AppProvider>
      <RouterProvider router={router} />
    </AppProvider>
  );
}
```

#### app/provider.tsx

```typescript
import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "@/components/ui/sonner";
import { queryClient } from "@/lib/react-query";

// AuthProvider はここに置かない。
// /login でも GET /v1/me が実行され 401 が発生するため router.tsx の保護ルートに配置する。
interface AppProviderProps {
  children: React.ReactNode;
}

export function AppProvider({ children }: AppProviderProps) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <Toaster />
    </QueryClientProvider>
  );
}
```

#### lib/axios.ts（Golangバックエンド接続）

**認証方式: httpOnly Cookie（`withCredentials: true`）**
localStorage への token 保存は禁止。httpOnly Cookie はブラウザが自動送信する。

```typescript
import Axios, { type InternalAxiosRequestConfig } from "axios";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

function authRequestInterceptor(config: InternalAxiosRequestConfig) {
  // httpOnly Cookie は withCredentials: true でブラウザが自動送信するため
  // Authorization ヘッダへの手動注入は不要
  config.headers = config.headers || {};
  config.headers.Accept = "application/json";
  config.headers["X-Request-ID"] = crypto.randomUUID(); // トレーシング用
  return config;
}

export const axios = Axios.create({
  baseURL: API_URL,
  timeout: 60000,
  withCredentials: true, // ★ httpOnly Cookie を cross-origin リクエストで送信するために必須
  headers: { "Content-Type": "application/json" },
});

axios.interceptors.request.use(authRequestInterceptor);
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (
      error.response?.status === 401 &&
      window.location.pathname !== "/login" // 無限リダイレクトループ防止
    ) {
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);
```

**禁止パターン:**
```typescript
// ❌ localStorage に token を保存（XSS で盗まれる）
const token = localStorage.getItem("token");
config.headers.Authorization = `Bearer ${token}`;
```

#### lib/react-query.ts

```typescript
import { QueryClient, type DefaultOptions } from "@tanstack/react-query";

const queryConfig: DefaultOptions = {
  queries: {
    // デフォルト: 中程度の変更頻度（医療記録、検査等）
    staleTime: 5 * 60 * 1000,
    gcTime: 15 * 60 * 1000,
    refetchOnWindowFocus: false,
    retry: 1,
  },
};

export const queryClient = new QueryClient({
  defaultOptions: queryConfig,
});

// リソース別キャッシング戦略（useQuery の staleTime/gcTime に渡す）
export const QUERY_STALE_TIMES = {
  STATIC: 30 * 60 * 1000,  // 30分: 飼主・ペット・マスタ等（低頻度変更）
  MEDIUM: 5 * 60 * 1000,   // 5分: 医療記録、検査、会計等（デフォルト）
  REALTIME: 2 * 60 * 1000, // 2分: 予約、Kanban等（高頻度）
};

export const QUERY_GC_TIMES = {
  LONG: 30 * 60 * 1000,     // 30分: マスタデータ等
  STANDARD: 15 * 60 * 1000, // 15分: ほとんどのデータ
  SHORT: 5 * 60 * 1000,     // 5分: 一時的なUI状態
};
```

#### React Query パターン

##### Query パターン（データ取得）

ファイル名は **kebab-case + verb-noun**: `get-xxx.ts`, `get-xxxs.ts`
フック名は **`useGet` + PascalCase**: `useGetXxx`, `useGetXxxs`

```typescript
// features/xxx/api/get-xxx.ts
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Xxx } from "@/types/xxx";            // 共有型
import type { Xxx as BackendXxx } from "@/types/generated/models"; // 自動生成型
import { transformXxx } from "./transforms";        // Backend → Frontend 変換

interface XxxResponse {
  data: BackendXxx;
}

export const getXxx = async (id: string): Promise<Xxx> => {
  const { data } = await axios.get<XxxResponse>(`/v1/xxx/${id}`);
  return transformXxx(data.data);
};

export const useGetXxx = (id: string) => {
  return useQuery({
    queryKey: ["xxx", id],
    queryFn: () => getXxx(id),
    enabled: !!id,
  });
};
```

**一覧取得パターン:**

```typescript
// features/xxx/api/get-xxxs.ts
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Xxx } from "@/types/xxx";
import type { Xxx as BackendXxx } from "@/types/generated/models";
import { transformXxx } from "./transforms";

interface XxxsResponse {
  data: BackendXxx[];
  total: number;
  page: number;
  limit: number;
}

export const getXxxs = async (): Promise<Xxx[]> => {
  const { data } = await axios.get<XxxsResponse>("/v1/xxx");
  return data.data.map(transformXxx);
};

export const useGetXxxs = () => {
  return useQuery({
    queryKey: ["xxxs"],
    queryFn: getXxxs,
  });
};
```

##### Mutation パターン（データ作成・更新・削除）

**パターン A（シンプル）**: hooks が浅い場合は API ファイルに `useMutation` hook を置く

```typescript
// features/pets/api/create-pet.ts
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Pet as BackendPet } from "@/types/generated/models";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";

export const createPet = async (data: CreatePetRequest): Promise<Pet> => {
  const { data: responseData } = await axios.post<BackendPet>("/v1/pets", data);
  return transformBackendPetToFrontend(responseData);
};

export const useCreatePet = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createPet,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["pets"] });
    },
  });
};
```

**パターン B（複雑なフォーム）**: フォームフックは API ファイルに生関数のみを置き、送信は hooks/ の `useActionState` に集約

```typescript
// features/owners/api/create-owner.ts — 生関数のみ export
import { axios } from "@/lib/axios";
import type { Owner as BackendOwner } from "@/types/generated/models";
import { transformOwner } from "./transforms";

export const createOwner = async (data: CreateOwnerRequest): Promise<Owner> => {
  const { data: responseData } = await axios.post<BackendOwner>("/v1/owners", data);
  return transformOwner(responseData);
};
// ※ 送信は hooks/use-owner-form.ts が useActionState 内で呼び出す
```

##### loaders.ts パターン（React Router Data Mode）

loader は **直接 axios を呼び出す**（`queryClient.prefetchQuery` は使わない）。
返したデータはルートコンポーネントが `useLoaderData()` で受け取る。

```typescript
// features/owners/loaders.ts
import { axios } from "@/lib/axios";

export interface OwnersLoaderData {
  pets: Pet[];
}

export const ownersLoader = async (): Promise<OwnersLoaderData> => {
  try {
    // 総件数確認 → 残りページを並列フェッチ（async-parallel）
    const { data: firstPage } = await axios.get<PetsResponse>("/v1/pets", {
      params: { page: 1, limit: PER_PAGE },
    });
    const totalPages = Math.ceil(firstPage.total / PER_PAGE);
    const remainingPages = await Promise.all(
      Array.from({ length: totalPages - 1 }, (_, i) =>
        axios.get<PetsResponse>("/v1/pets", { params: { page: i + 2, limit: PER_PAGE } })
          .then(r => r.data)
      )
    );
    const allPets = [firstPage, ...remainingPages]
      .flatMap(page => page.data.map(transformBackendPetToFrontend));
    return { pets: allPets };
  } catch {
    throw new Response("データの取得に失敗しました", { status: 500 });
  }
};
```

**bulletproof-react との loader 差分:**

| bulletproof-react | このプロジェクト |
|---|---|
| `queryClient.getQueryData(key) ?? queryClient.fetchQuery(queryOptions)` | `axios.get()` で直接フェッチ |
| React Query キャッシュに載る（コンポーネントは `useQuery` で取得） | React Router の `useLoaderData()` で取得（React Query を経由しない） |
| loader と hook が `queryOptions` factory を共有 | loader と hook は独立（二重フェッチを避けるため loader のデータは `useLoaderData`） |

##### api/ vs hooks/ の区別

| 配置場所 | 用途 | 例 |
|---------|------|-----|
| `api/get-xxx.ts` | React Query の useQuery hook + 生フェッチ関数 | `getOwners`, `useGetOwners` |
| `api/create-xxx.ts` | 生 mutation 関数（+ 簡単な場合は useMutation hook） | `createOwner`, `useCreatePet` |
| `hooks/use-xxx-form.ts` | フォーム状態・useActionState 送信・ビジネスロジック | `useOwnerForm`, `useHospitalizationForm` |

#### Feature API実装例（features/owners/api/get-owners.ts）

```typescript
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Owner } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { Owner as BackendOwner } from "@/types/generated/models";

interface OwnersResponse {
  data: BackendOwner[];
  total: number;
  page: number;
  limit: number;
}

export const getOwners = async (): Promise<Owner[]> => {
  const { data } = await axios.get<OwnersResponse>("/v1/owners");
  return data.data.map(transformOwner);
};

export const useGetOwners = () => {
  return useQuery({
    queryKey: ["owners"],
    queryFn: getOwners,
  });
};
```

#### Feature公開API例（features/owners/index.ts）

```typescript
// Routes（ページコンポーネント）
// ★ 外部からの import は必ずこの index.ts（barrel）経由で行う（Feature Indexing — MANDATORY）
//    import { OwnersList } from "@/features/owners"                     ← 正しい（barrel import）
//    import { OwnersList } from "@/features/owners/routes/OwnersList"  ← 禁止（deep import）
export { OwnerForm } from "./routes/OwnerForm";
export { OwnersList } from "./routes/OwnersList";

// Hooks
export { useOwnerForm } from "./hooks/useOwnerForm";

// Components（外部公開が必要なもののみ）
export { PetEditModal } from "./components/PetEditModal";

// loaders も同じ index.ts から公開し、呼び出し元（router.tsx 等）は barrel 経由で import する
export { ownersLoader, ownerLoader } from "./loaders";
// import { ownersLoader, ownerLoader } from "@/features/owners";  ← router.tsx 側の呼び出し例
```

#### main.tsx

```typescript
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./app";
import "./styles/globals.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
```

---

## 2. React 19 ルール

### 2.1 コンポーネント定義

```typescript
// ✅ 正しい: 関数宣言 + 明示的Props型
interface PatientCardProps {
  patient: Patient;
  onSelect?: (id: string) => void;
  className?: string;
  ref?: React.Ref<HTMLDivElement>;  // React 19: ref as prop
}

export function PatientCard({
  patient,
  onSelect,
  className,
  ref,
}: PatientCardProps) {
  return (
    <div ref={ref} className={className}>
      {/* ... */}
    </div>
  );
}

// ❌ 禁止: FC型（React 19では不要）
export const PatientCard: FC<Props> = ({ patient }) => {};

// ❌ 禁止: forwardRef（React 19では不要）
export const PatientCard = forwardRef<HTMLDivElement, Props>((props, ref) => {});

// ✅ React.memo: 親の別 state 変化で不要な再レンダーが発生するコンポーネントに適用
// 対象: 重いフォームセクション、テーブル行、独立したサブセクション
// 前提: props として渡す関数は useCallback で安定化すること
const OwnerInfoSection = memo(function OwnerInfoSection({ data, onChange }: Props) {
  return <div>...</div>;
});

// ❌ 軽量でシンプルなコンポーネントへの過剰適用は不要
export const PatientCard = memo(({ patient }) => {});  // ❌ (props が毎回変わるなら無意味)
```

### 2.2 フォーム管理：useActionState（フォーム標準）+ useTransition（非フォーム非同期）

このプロジェクトの**フォームは複雑・単純を問わず `useActionState` を標準**とする（`frontend/CLAUDE.md` MANDATORY）。
多フィールドの制御フォームでも、フィールド状態は `useState`、送信・バリデーション・pending は `useActionState` で管理する。
参照実装は `features/owners/hooks/use-owner-form.ts`（飼主 + ペット CRUD の複雑フォーム）。

`useTransition` は**フォーム送信には使わない**。リスト再取得・ナビゲーション・削除など、フォーム外の非ブロッキング非同期更新の pending 管理に限定する（例: `OwnersList.tsx`, `use-reception-kanban.ts`, `use-master-crud.ts`）。

#### 標準: `useState`（制御フィールド）+ `useActionState`（送信・バリデーション・pending）

```typescript
// features/xxx/hooks/use-xxx-form.ts
import { useState, useActionState, useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

export function useXxxForm(id?: string, initialData?: Xxx) {
  const queryClient = useQueryClient();

  // ✅ 制御フィールドは useState + lazy init（mapToFormData は初回レンダーのみ）
  const [formData, setFormData] = useState<XxxFormData>(
    () => initialData ? mapToFormData(initialData) : DEFAULT_DATA
  );

  // ✅ functional setState で deps なし → useCallback が安定
  const handleInputChange = useCallback((field: string, value: string | boolean) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  }, []);

  // ✅ 送信・バリデーション・pending は useActionState に集約
  const [formState, formAction, isPending] = useActionState(
    async (_prev: ActionState, _formData: FormData): Promise<ActionState> => {
      const errors: Record<string, string> = {};
      if (!formData.name.trim()) errors.name = "名前を入力してください";
      if (Object.keys(errors).length > 0) {
        return { success: false, fieldErrors: errors, timestamp: Date.now() };
      }
      try {
        if (id) {
          await updateXxx(id, buildUpdateRequest(formData));
        } else {
          await createXxx(buildCreateRequest(formData));
        }
        await queryClient.invalidateQueries({ queryKey: ["xxxs"] });
        toast.success("保存しました");
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  return { formData, isPending, handleInputChange, formAction, formState };
}
```

```typescript
// 呼び出し側（route/component）— <form action={formAction}> で駆動
<form action={formAction}>
  {/* ✅ ? : null （&& 禁止） */}
  {formState.fieldErrors?.name ? <p className="text-red-600">{formState.fieldErrors.name}</p> : null}
  <Input name="name" value={formData.name} onChange={(e) => handleInputChange("name", e.target.value)} />
  <SubmitButton>保存</SubmitButton>  {/* useFormStatus で pending を取得（§2.3） */}
</form>
```

#### 単純フォーム（非制御）も `useActionState`

単一フィールドの作成フォーム等は、`useState` を省き `FormData` から直接読み取る。手法は同一（`useActionState`）。

```typescript
const [state, formAction, isPending] = useActionState<ActionState, FormData>(
  async (_, formData) => {
    const name = formData.get("name") as string;
    if (!name) return { success: false, fieldErrors: { name: "名前は必須です" }, timestamp: Date.now() };
    await createXxx({ name });
    return { success: true, timestamp: Date.now() };
  },
  INITIAL_ACTION_STATE
);
```

**使い分けガイド:**

| ケース | 手法 |
|---|---|
| フォーム送信全般（多フィールド制御フォーム含む） | `useActionState`（制御フィールドは `useState` 併用） |
| 単一フィールド・非制御の作成フォーム | `useActionState`（`FormData` 読み取り） |
| フォーム外の非同期更新（リスト再取得・ナビゲーション・削除の pending） | `useTransition` |

> ⚠️ `useTransition` + カスタム hook でフォーム送信を管理するのは**旧パターン**であり禁止。
> 既存コードは全て `useActionState` へ移行済み（`use-owner-form.ts` 他 9 フォームフック）。

### 2.3 useFormStatus（送信ボタン）

```typescript
import { useFormStatus } from "react-dom";
import { Button } from "@/components/ui/button";

interface SubmitButtonProps {
  children: React.ReactNode;
  loadingText?: string;
}

export function SubmitButton({
  children,
  loadingText = "処理中...",
}: SubmitButtonProps) {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" disabled={pending}>
      {pending ? loadingText : children}
    </Button>
  );
}

// 使用例
<form action={formAction}>
  <SubmitButton>保存する</SubmitButton>
</form>
```

### 2.4 useOptimistic（楽観的更新）

```typescript
import { useOptimistic } from "react";

interface Pet {
  id: string;
  name: string;
}

export function usePetList(initialPets: Pet[]) {
  const [optimisticPets, addOptimisticPet] = useOptimistic(
    initialPets,
    (state, newPet: Pet) => [...state, newPet]
  );

  const addPet = async (pet: Omit<Pet, "id">) => {
    // 楽観的に即座にUIに反映
    const tempPet = { ...pet, id: `temp-${Date.now()}` };
    addOptimisticPet(tempPet);

    // 実際のAPI呼び出し
    const created = await createPet(pet);
    return created;
  };

  return { pets: optimisticPets, addPet };
}
```

### 2.5 use()（Promise/Context読み取り）

```typescript
import { Suspense, use } from "react";

// Promise を直接読み取る
function PatientList({ patientsPromise }: { patientsPromise: Promise<Patient[]> }) {
  const patients = use(patientsPromise);

  return (
    <ul>
      {patients.map((patient) => (
        <li key={patient.id}>{patient.name}</li>
      ))}
    </ul>
  );
}

// 使用時は必ず Suspense でラップ
function PatientPage() {
  const patientsPromise = fetchPatients();

  return (
    <Suspense fallback={<div>読み込み中...</div>}>
      <PatientList patientsPromise={patientsPromise} />
    </Suspense>
  );
}

// Context を直接読み取る
function ThemeButton() {
  const theme = use(ThemeContext);
  return <button style={{ color: theme.primary }}>Click</button>;
}
```

### 2.6 Context（Provider省略）

```typescript
import { createContext, useContext, useState } from "react";

interface AuthContextValue {
  user: User | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

// Context作成
export const AuthContext = createContext<AuthContextValue | null>(null);

// カスタムhook
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthContext");
  }
  return context;
}

// Provider（React 19: Context を直接使用）
function AppProvider({ children }: { children: React.ReactNode }) {
  const authValue = useAuthStore();

  return (
    // React 19: <AuthContext.Provider value={...}> ではなく
    <AuthContext value={authValue}>
      {children}
    </AuthContext>
  );
}
```

---

## 3. TypeScript ルール

### 3.1 型定義

```typescript
// ✅ interface: オブジェクト型に使用
interface Owner {
  id: string;
  name: string;
  email: string;
  pets: Pet[];
}

// ✅ type: Union、Intersection、関数型に使用
type Status = "active" | "inactive" | "pending";
type OwnerWithPets = Owner & { totalPets: number };
type FetchFn<T> = (id: string) => Promise<T>;

// ✅ Props型命名: ComponentNameProps
interface OwnerCardProps {
  owner: Owner;
  onSelect?: (id: string) => void;
}

// ❌ 禁止: I prefix
interface IOwner {}  // ❌
interface Owner {}   // ✅

// ❌ 禁止: any
const data: any = fetchData();  // ❌
const data: unknown = fetchData();  // ✅ → 型ガードで絞り込む
```

### 3.2 型ガード

```typescript
// カスタム型ガード
function isOwner(value: unknown): value is Owner {
  return (
    typeof value === "object" &&
    value !== null &&
    "id" in value &&
    "name" in value
  );
}

// 使用例
if (isOwner(data)) {
  console.log(data.name);  // Owner型として扱える
}

// Discriminated Union
type ApiResponse<T> =
  | { status: "success"; data: T }
  | { status: "error"; error: string };

function handleResponse<T>(response: ApiResponse<T>) {
  if (response.status === "success") {
    return response.data;  // T型
  } else {
    throw new Error(response.error);  // string
  }
}
```

### 3.3 Generics

```typescript
// API レスポンス型
interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
}

// フェッチ関数
async function fetchPaginated<T>(
  url: string,
  page: number
): Promise<PaginatedResponse<T>> {
  const response = await fetch(`${url}?page=${page}`);
  return response.json();
}

// 使用例
const owners = await fetchPaginated<Owner>("/api/owners", 1);
//    ^? PaginatedResponse<Owner>
```

### 3.4 Import / Export

```typescript
// ✅ 名前付きexport（tree-shaking対応）
export function OwnerCard() {}
export function OwnerList() {}
export type { Owner, Pet };

// ❌ 禁止: default export（IDE補完が弱い）
export default function OwnerCard() {}

// ❌ 禁止: wildcard re-export（tree-shaking阻害）
// components/index.ts
export * from "./OwnerCard";
export * from "./OwnerList";

// ✅ 許可: 明示的なre-export（feature公開API用）
// features/owners/api/index.ts
export { getOwners } from "./get-owners";
export { createOwner } from "./create-owner";

// ✅ feature 外からは barrel（index.ts）経由で named import
import { OwnerCard } from "@/features/owners";
```

### 3.5 Import順序

```typescript
// 1. React / Framework
import { useState, useEffect, Suspense } from "react";
import { useNavigate, useParams } from "react-router";

// 2. 外部ライブラリ
import { format } from "date-fns";
import { z } from "zod";

// 3. 共有モジュール (@/)
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useDebounce } from "@/hooks/useDebounce";
import { formatDate } from "@/lib/format/date";

// 4. feature内部（相対パス、同一feature内のみ）
import { OwnerCard } from "../components/OwnerCard";
import { useOwnerForm } from "../hooks/useOwnerForm";

// 5. 型（type keyword付き）
import type { Owner, Pet } from "@/types";
import type { OwnerFormData } from "../types";
```

### 3.6 バックエンド型の扱い（generated models.ts）

#### 概要

```
backend/internal/model/*.go
    ↓ make codegen (tygo)
src/types/generated/models.ts   ← 自動生成・直接編集禁止
    ↓ Omit / Partial / ReturnType
features/xxx/api/types.ts       ← APIリクエスト型（導出）
src/lib/transforms/xxx.ts       ← フロントエンドドメイン型（ReturnType導出）
features/xxx/types/index.ts     ← フォームデータ型（UI専用）
```

**重要ルール:**
- `src/types/generated/models.ts` は **直接編集禁止**。`make codegen` で再生成される
- Goモデル変更 → `make codegen` → `models.ts` 更新 → 型エラーをフロントエンドで修正

---

#### APIリクエスト型の導出パターン

`interface` を手書きせず、`models.ts` の型から `Omit`/`Partial`/`Required` で導出する。
参照実装: `src/types/pet.ts`, `src/features/trimming/api/types.ts`

```typescript
// features/xxx/api/types.ts
import type { XxxRecord as BackendXxxRecord } from "@/types/generated/models";

export type BackendXxx = BackendXxxRecord;  // alias（transforms.ts で使用）

/**
 * サーバー側で自動生成されるフィールド（リクエストに含めない）
 */
type ServerFields =
  | "id"
  | "clinic_id"
  | "created_at"
  | "updated_at"
  | "some_relation";  // リレーション（外部キーで代替）

/**
 * リクエストで送信可能なフィールド
 * Goモデル変更 → make codegen → models.ts 更新で自動追従
 */
type XxxWritable = Omit<BackendXxxRecord, ServerFields>;

/**
 * 作成リクエスト: 必須フィールドを Required で明示、残りはoptional
 */
export type CreateXxxRequest =
  Required<Pick<XxxWritable, "required_field_a" | "required_field_b">> &
  Partial<Omit<XxxWritable, "required_field_a" | "required_field_b">> & {
    extra_ids?: number[];  // リレーションIDなどmodels.tsに存在しないフィールドは手動追加
  };

/**
 * 更新リクエスト（PATCH: 全フィールドoptional）
 */
export type UpdateXxxRequest = Partial<XxxWritable> & {
  extra_ids?: number[];
};
```

**禁止パターン:**

```typescript
// ❌ 手書きinterfaceはGoモデルとの乖離を生む
export interface CreateXxxRequest {
  field_a: string;
  field_b?: number;
  // Goモデルにfiled_cが追加されても気づかない
}
```

---

#### フロントエンドドメイン型の導出パターン（ReturnType）

transform関数の戻り値から型を自動導出する。手動でinterfaceを維持しない。
参照実装: `src/lib/transforms/pet.ts`

```typescript
// lib/transforms/xxx.ts
import type { XxxRecord as BackendXxx } from "@/types/generated/models";

// BackendXxx（snake_case）→ フロントエンド型（camelCase）変換
export const transformBackendXxxToFrontend = (x: BackendXxx) => ({
  id: String(x.id ?? 0),
  // snake_case → camelCase
  someField: x.some_field ?? "",
  // enumマッピング（APIの英語値 → 日本語表示値）
  status: STATUS_MAP[x.status ?? ""] ?? x.status ?? "",
  // リレーション
  relatedName: x.related?.name ?? "",
});

/**
 * Xxx フロントエンド型 — transformの戻り値から自動導出
 * 手動管理せず BackendXxx（models.ts）と常に同期
 */
export type Xxx = ReturnType<typeof transformBackendXxxToFrontend>;
```

**禁止パターン:**

```typescript
// ❌ interfaceを手動管理するとtransformとの乖離が生じる
export interface Xxx {
  id: string;
  someField: string;
  // transformに追加したフィールドをここにも追加し忘れる
}
```

---

#### フォームデータ型（UI専用）

フォームの入力状態はUIに最適化した型として `features/xxx/types/index.ts` に定義する。
バックエンドモデルと1対1対応しない（File型、UI専用フラグ等を含む）。

```typescript
// features/xxx/types/index.ts

// UI表示用の定数（as const パターン）
export const XXX_STATUS_VALUES = ["予約", "施術中", "完了"] as const;
export type XxxStatus = (typeof XXX_STATUS_VALUES)[number];

// フォーム入力データ（UIに最適化、バックエンド型と必ずしも一致しない）
export interface XxxFormData {
  // 文字列で保持（数値もstringで管理してUIの空文字を許容）
  someField: string;
  // File型など、バックエンドに存在しないUI専用フィールド
  imageFile: File | null;
  // UI専用フラグ
  isDirty: boolean;
}
```

---

#### 型の配置ルール まとめ

| 型の種類 | 配置場所 | 例 |
|---------|---------|-----|
| バックエンドモデル型 | `src/types/generated/models.ts`（自動生成・編集禁止） | `TrimmingRecord`, `Pet` |
| APIリクエスト型 | `src/types/xxx.ts` | `CreateTrimmingRequest`, `CreatePetRequest` |
| フロントエンドドメイン型 | `src/lib/transforms/xxx.ts`（ReturnType導出） | `Pet = ReturnType<...>` |
| フォームデータ型 | `src/types/xxx.ts` | `TrimmingFormData`, `PetFormData` |
| DI用インターフェース型 | `src/types/xxx.ts` | `PetMutations` |

> **原則**: feature をまたぐかどうかに関わらず、全ての型定義は `src/types/` に配置する。
> `features/xxx/types/index.ts` は `src/types/xxx.ts` への re-export のみ許容（直接型定義を書かない）。
>
> **例外(FA9)**: transform 関数の ReturnType を型の正本とする場合に限り、`src/types/index.ts` から feature barrel の型を re-export してよい。

---

#### モデル変更時の手順

```bash
# 1. Goモデルを編集
# backend/internal/model/xxx.go

# 2. 型生成
make codegen  # tygo → src/types/generated/models.ts を更新

# 3. フロントエンドの型エラーを修正
#    - api/types.ts の ServerFields を必要に応じて更新
#    - transforms.ts に新フィールドのマッピングを追加
#    - types/index.ts の FormData を必要に応じて更新
docker compose exec frontend pnpm build
```

---

## 4. コンポーネント設計

### 4.1 Props設計

```typescript
// ✅ 必須propsと任意propsを明確に
interface ButtonProps {
  // 必須
  children: React.ReactNode;

  // 任意（デフォルト値あり）
  variant?: "primary" | "secondary" | "danger";
  size?: "sm" | "md" | "lg";
  disabled?: boolean;

  // コールバック
  onClick?: () => void;

  // DOM属性継承
  className?: string;
  ref?: React.Ref<HTMLButtonElement>;
}

// ✅ ComponentPropsWithoutRef を活用
import { ComponentPropsWithoutRef } from "react";

interface InputProps extends ComponentPropsWithoutRef<"input"> {
  label?: string;
  error?: string;
}
```

### 4.2 コンポーネント分割基準

```typescript
// ❌ 巨大コンポーネント（500行超）
function OwnerForm() {
  // フォームロジック 200行
  // バリデーション 100行
  // UI 200行
}

// ✅ 責務で分割 + memo() で再レンダー境界を設定（OwnerForm.tsx の実装パターン）
// ページコンポーネント: フォーム状態を管理し、子に個別propsを渡す
export function OwnerForm() {
  const { formData, handleInputChange, handleSubmit, isPending } = useOwnerForm();
  return (
    <form onSubmit={handleSubmit}>
      <OwnerInfoSection
        formData={formData}
        onChange={handleInputChange}
      />
      <PetSection
        pets={formData.pets}
        onAdd={...}
      />
    </form>
  );
}

// セクションコンポーネント: memo() で囲み、他セクション変更時の再レンダーを防ぐ
const OwnerInfoSection = memo(function OwnerInfoSection({ formData, onChange }) {
  return (
    <div>
      <Input value={formData.name} onChange={e => onChange("name", e.target.value)} />
      <Input value={formData.email} onChange={e => onChange("email", e.target.value)} />
    </div>
  );
});
```

### 4.3 条件付きレンダリング

```typescript
// ✅ 早期return
function UserProfile({ user }: { user: User | null }) {
  if (!user) {
    return <div>ログインしてください</div>;
  }

  return <div>{user.name}</div>;
}

// ✅ 三項演算子（シンプルな場合）
function Status({ isActive }: { isActive: boolean }) {
  return <span>{isActive ? "有効" : "無効"}</span>;
}

// ❌ && 演算子（rendering-conditional-render 違反）
// number が 0 のとき "0" がレンダリングされる危険がある
function ErrorMessage({ error }: { error?: string }) {
  return error && <p className="text-red-600">{error}</p>;  // ❌
}

// ✅ 三項演算子を常に使用（ternary + null パターン）
function ErrorMessage({ error }: { error?: string }) {
  return error ? <p className="text-red-600">{error}</p> : null;  // ✅
}

// ❌ ネストした三項演算子
function Status({ status }) {
  return status === "active" ? "有効" : status === "pending" ? "保留" : "無効";
}

// ✅ オブジェクトマップ
const STATUS_LABELS: Record<Status, string> = {
  active: "有効",
  pending: "保留",
  inactive: "無効",
};

function Status({ status }: { status: Status }) {
  return <span>{STATUS_LABELS[status]}</span>;
}
```

---

## 5. Hooks ルール

### 5.1 カスタムhook設計

命名規則: `use + 動詞/名詞`（例: `useOwnerForm`, `useHospitalizationList`）

**実プロジェクトのフォーム hook パターン（`useActionState` 使用）:**

制御フィールドは `useState`、送信・バリデーション・pending は `useActionState` に集約する（詳細な参照実装は §2.2 と `features/owners/hooks/use-owner-form.ts`）。

```typescript
// features/owners/hooks/use-owner-form.ts
import { useState, useActionState, useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

export function useOwnerForm(id?: string, initialOwner?: Owner) {
  const queryClient = useQueryClient();

  // ✅ lazy init: 高コストなマッピングは初回のみ実行
  const [ownerData, setOwnerData] = useState<OwnerData>(
    () => initialOwner ? mapOwnerToFormData(initialOwner) : DEFAULT_OWNER_DATA
  );

  // ✅ functional setState → deps なし → useCallback が stable
  const handleInputChange = useCallback((field: string, value: string | boolean) => {
    setOwnerData(prev => ({ ...prev, [field]: value }));
  }, []);

  // ✅ 送信・バリデーション・pending は useActionState に集約
  const [formState, formAction, isPending] = useActionState(
    async (_prev: ActionState, _formData: FormData): Promise<ActionState> => {
      try {
        if (id) {
          await updateOwner(id, buildUpdateRequest(ownerData));
        } else {
          await createOwner(buildCreateRequest(ownerData));
        }
        await queryClient.invalidateQueries({ queryKey: ["owners"] });
        toast.success("保存しました");
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

  return { ownerData, formState, isPending, handleInputChange, formAction };
}
```

**`useState` + `setIsPending` を使わない理由:**

| 手法 | 問題点 |
|---|---|
| `useState(false)` + `setIsPending(true/false)` | try-finally で手動管理が必要、エラー時の reset 漏れリスク |
| `useActionState`（フォーム送信）/ `useTransition`（フォーム外の非同期） | React が pending 状態を自動管理、reset 漏れが構造的に起きない |

### 5.2 依存配列

```typescript
// ✅ 必要な依存のみ含める
useEffect(() => {
  fetchData(id);
}, [id]);  // id が変わったときのみ実行

// ✅ useCallback の依存
const handleClick = useCallback(() => {
  onSelect(item.id);
}, [item.id, onSelect]);

// ❌ オブジェクト全体を依存に含めない
useEffect(() => {
  console.log(user.name);
}, [user]);  // ❌ userオブジェクトの参照が変わるたびに実行

useEffect(() => {
  console.log(user.name);
}, [user.name]);  // ✅ user.name が変わったときのみ
```

### 5.3 副作用の分離

```typescript
// ❌ 複数の副作用を1つのuseEffectに
useEffect(() => {
  fetchUser();
  trackPageView();
  setupWebSocket();
}, []);

// ✅ 副作用ごとに分離
useEffect(() => {
  fetchUser();
}, []);

useEffect(() => {
  trackPageView();
}, []);

useEffect(() => {
  const ws = setupWebSocket();
  return () => ws.close();
}, []);
```

---

## 6. スタイリング（Tailwind CSS）

### 6.1 クラス順序

```tsx
// 推奨順序:
// 1. レイアウト (display, position, flex/grid)
// 2. ボックスモデル (width, height, margin, padding)
// 3. 視覚 (background, border, shadow)
// 4. タイポグラフィ (font, text)
// 5. その他 (cursor, transition, animation)

<div className="
  flex items-center justify-between gap-4
  w-full p-4
  bg-white border border-gray-200 rounded-lg shadow-sm
  text-sm font-medium text-gray-900
  cursor-pointer transition-colors hover:bg-gray-50
">
```

### 6.2 条件付きクラス

```typescript
import { cn } from "@/lib/utils";

interface ButtonProps {
  variant?: "primary" | "secondary";
  size?: "sm" | "md" | "lg";
  className?: string;
}

function Button({ variant = "primary", size = "md", className }: ButtonProps) {
  return (
    <button
      className={cn(
        // ベーススタイル
        "inline-flex items-center justify-center rounded-md font-medium transition-colors",
        // バリアント
        {
          "bg-blue-600 text-white hover:bg-blue-700": variant === "primary",
          "bg-gray-200 text-gray-900 hover:bg-gray-300": variant === "secondary",
        },
        // サイズ
        {
          "px-3 py-1.5 text-sm": size === "sm",
          "px-4 py-2 text-base": size === "md",
          "px-6 py-3 text-lg": size === "lg",
        },
        // カスタムクラス
        className
      )}
    >
      {/* ... */}
    </button>
  );
}
```

### 6.3 レスポンシブ

```tsx
// モバイルファースト
<div className="
  flex flex-col gap-2
  md:flex-row md:gap-4
  lg:gap-6
">

// ブレークポイント
// sm: 640px
// md: 768px
// lg: 1024px
// xl: 1280px
// 2xl: 1536px
```

---

## 7. エラーハンドリング

### 7.1 ErrorBoundary

```typescript
// components/errors/RouteErrorBoundary.tsx
import { useRouteError, isRouteErrorResponse, Link } from "react-router";

export function RouteErrorBoundary() {
  const error = useRouteError();

  if (isRouteErrorResponse(error)) {
    if (error.status === 404) {
      return (
        <div className="flex flex-col items-center justify-center h-full">
          <h1 className="text-2xl font-bold">ページが見つかりません</h1>
          <Link to="/">ダッシュボードへ戻る</Link>
        </div>
      );
    }
  }

  return (
    <div className="flex flex-col items-center justify-center h-full">
      <h1 className="text-2xl font-bold">エラーが発生しました</h1>
      <p>予期せぬエラーが発生しました。再度お試しください。</p>
    </div>
  );
}
```

### 7.2 API エラー

`axios` インスタンスは `lib/axios.ts` で一元管理（`withCredentials: true` + 401 自動リダイレクト）。
詳細は **Section 1.4 の `lib/axios.ts` 例** を参照。

```typescript
// features/xxx/api/get-xxx.ts での API エラーハンドリング例
import { axios } from "@/lib/axios";
import { AxiosError } from "axios";
import { toast } from "sonner";

try {
  const { data } = await axios.get<XxxResponse>(`/v1/xxx/${id}`);
} catch (error) {
  if (error instanceof AxiosError) {
    if (error.response?.status === 404) {
      toast.error("データが見つかりません");
    } else {
      toast.error("通信エラーが発生しました");
    }
  }
}

// loaders.ts でのエラー処理: throw new Response でルートエラーバウンダリに委譲
export const xxxLoader = async (): Promise<XxxLoaderData> => {
  try {
    const { data } = await axios.get<XxxResponse>("/v1/xxx");
    return { items: data.data };
  } catch {
    throw new Response("データの取得に失敗しました", { status: 500 });
  }
};
```

---

## 8. テスト

### 8.1 ファイル配置

テストファイルは対象ファイルと**同階層**に配置する：

```
src/
├── lib/
│   ├── utils.ts
│   └── utils.test.ts              # 同階層に配置
├── components/
│   └── shared/
│       └── StatusBadge/
│           ├── StatusBadge.tsx
│           ├── StatusBadge.test.tsx  # 同階層に配置
│           └── index.ts
├── features/
│   └── owners/
│       ├── api/
│       │   ├── get-owners.ts
│       │   └── get-owners.test.ts    # 同階層に配置
│       ├── components/
│       │   └── OwnerCard/
│       │       ├── OwnerCard.tsx
│       │       ├── OwnerCard.test.tsx  # 同階層に配置
│       │       └── index.ts
│       └── hooks/
│           ├── use-owner-form.ts
│           └── use-owner-form.test.ts  # 同階層に配置
└── testing/
    ├── setup.ts                  # テスト共通設定
    ├── mocks/                    # MSW handlers
    └── utils.tsx                 # テストユーティリティ (createTestWrapper 等)
```

**ルール:**
- テストファイル名: `[対象ファイル名].test.ts(x)`
- 新規テストは同階層への co-located 配置を推奨する
- `__tests__/` は FE5-23 で全廃済み。テストは対象ファイル隣接配置とし、`__tests__/` の新設は禁止（正本: `frontend/CLAUDE.md`）

### 8.2 テスト例

```typescript
// utils.test.ts
import { describe, it, expect } from "vitest";
import { formatDate, calculateAge } from "./utils";

describe("formatDate", () => {
  it("should format date in Japanese format", () => {
    const date = new Date("2024-01-15");
    expect(formatDate(date)).toBe("2024年1月15日");
  });

  it("should handle invalid date", () => {
    expect(() => formatDate(new Date("invalid"))).toThrow();
  });
});

describe("calculateAge", () => {
  it("should return correct age", () => {
    const birthDate = new Date("2020-01-15");
    // 現在日付をモック
    vi.setSystemTime(new Date("2024-06-01"));
    expect(calculateAge(birthDate)).toBe(4);
  });
});
```

### 8.3 コンポーネントテスト

```typescript
// Button.test.tsx
import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { Button } from "./button";

describe("Button", () => {
  it("should render children", () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText("Click me")).toBeInTheDocument();
  });

  it("should call onClick when clicked", () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click me</Button>);

    fireEvent.click(screen.getByRole("button"));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it("should be disabled when disabled prop is true", () => {
    render(<Button disabled>Click me</Button>);
    expect(screen.getByRole("button")).toBeDisabled();
  });
});
```

---

## 9. 禁止事項一覧

| 禁止 | 理由 | 代替 |
|------|------|------|
| `any` 型 | 型安全性の破壊 | `unknown` + 型ガード |
| `FC` / `React.FC` | React 19では不要 | 関数宣言 |
| `forwardRef` | React 19では不要 | ref as prop |
| feature間import | アーキテクチャ違反 | app層で合成 |
| `export *` re-export | tree-shaking阻害 | 明示的export or 直接import |
| `console.log` 放置 | 本番コード汚染 | 削除またはlogger使用 |
| ハードコード値 | 保守性低下 | 定数化 |
| default export | IDE補完が弱い | 名前付きexport |
| インラインスタイル | 一貫性欠如 | Tailwind CSS |
| `!important` | 詳細度問題 | クラス設計見直し |
| feature内部から自 feature の barrel（`../api` 等）を経由 | 不要な間接参照、tree-shaking 阻害 | 直接ファイル指定（`../api/delete-owner` 等） |
| feature外から feature 内部ファイルを deep import（`@/features/xxx/api/get-xxx` 等） | Feature Indexing 違反、内部構造への依存 | feature の `index.ts` 経由（`@/features/xxx`） |
| コメントのみ・空の index.ts | 死ファイル、混乱の原因 | 削除する |
| re-exportのみで自身のロジックを持たないファイル | 参照ゼロなら死ファイル | 削除する |
| 実ファイルが存在するフォルダの `.gitkeep` | 不要 | 削除する |
| API query hook を `useOwners` と命名（動詞省略） | 命名規則違反 | `useGetOwners`（`useGet` + エンティティ名） |
| loader 内で `queryClient.prefetchQuery` を使う | このプロジェクトは直接 axios + `useLoaderData` パターン | `axios.get()` で直接フェッチして返す |
| `localStorage` に token を保存 | XSS で盗まれる | httpOnly Cookie + `withCredentials: true` |
| フォーム送信の pending を `useState(false)` + `setIsPending` で管理 | try-finally でのリセット漏れリスク | `useActionState` の `isPending` を使う（フォーム外の非同期は `useTransition`） |
| `useTransition` + カスタム hook でフォーム送信を管理 | 旧パターン。`useActionState` に全移行済み | `useActionState`（制御フィールドは `useState` 併用・§2.2） |

---

## 10. チェックリスト

### 新規コンポーネント作成時
- [ ] 関数宣言で定義（FC禁止）
- [ ] Props型を明示的に定義
- [ ] ref は props として受け取る
- [ ] 適切なディレクトリに配置
- [ ] テストファイル作成

### PR作成時
- [ ] `docker compose exec frontend pnpm lint` がパス
- [ ] `docker compose exec frontend pnpm build` がパス
- [ ] `docker compose exec frontend pnpm test:run` がパス
- [ ] any型を使用していない
- [ ] feature間importがない
- [ ] 不要なconsole.logがない
- [ ] barrel index 経由の import がない（feature 外からも含む）
- [ ] コメントのみ・空の index.ts / .gitkeep が残っていない
- [ ] 参照ゼロの re-export ファイルが残っていない

---

## 11. この構成の特徴

### React Router Data Mode
- `createBrowserRouter`を使用
- GolangバックエンドとRESTful APIで連携
- SSR/SSGの複雑性なし
- Lazy loadingによるコード分割

### React 19機能の活用
| 機能 | 用途 | 備考 |
|------|------|------|
| `useActionState` | **全フォーム送信の標準**（複雑な制御フォーム含む） | `hooks/use-xxx-form.ts`。制御フィールドは `useState` 併用 |
| `useTransition` | フォーム外の非同期更新の pending 管理（リスト再取得・ナビゲーション・削除） | フォーム送信には使わない |
| `useOptimistic` | 楽観的UI更新（カンバン・タスク管理など） | |
| `ref as prop` | コンポーネント簡素化（`forwardRef` 不要） | |
| `useFormStatus` | サブミット状態の管理（フォームの子コンポーネント内） | |
| `use()` | Promise/Context 直接読み取り | |
| Document Metadata | ブラウザタブタイトル（UX向上） | |

### Feature-Based Architecture
- 各機能が `api/`, `components/`, `hooks/`, `routes/`, `types/` に完結
- `index.ts` は公開 API のカタログ（import 先には使わない — 直接ファイルを参照）
- cross-feature 合成は `app/pages/` + props 注入（依存逆転）
- `shared → features → app` の単方向依存（feature 間の直接 import 禁止）

### 認証システム向け最適化
- SEO機能は不要
- 認証後のみアクセス可能
- SPA構成

### 動物病院システム特性対応
- SOAPS形式カルテ対応
- カンバンボードワークフロー
- 入院管理（ケアプラン・デイリーログ）
- トリミング・ワクチン管理

---

## 12. パフォーマンス最適化（Vercel React Best Practices準拠）

> **参照実装**: `features/owners/` が全パターンのベストプラクティス実装。
> 新機能を実装する際はこの feature を手本にすること。

### 12.1 Re-render 最適化

#### `rerender-memo` — コンポーネント境界で再レンダーを分断する

大きなフォームやページは、独立した責務ごとに `memo()` コンポーネントに分割する。
**前提**: props として渡すハンドラはすべて `useCallback` で安定化すること。

```typescript
// ✅ 飼主フォームの参照実装（OwnerForm.tsx）
// ペット操作・モーダル開閉では ownerData/fieldErrors が変わらないため
// 17フィールド全体の再レンダーを防ぐ
const OwnerInfoSection = memo(function OwnerInfoSection({
  ownerData,
  fieldErrors,
  onChange,
  onClearError,
}: OwnerInfoSectionProps) {
  return <div className="grid grid-cols-4 gap-4">...</div>;
});

// ✅ テーブル行の参照実装 — N行×Mハンドラのインライン関数生成を排除
const PetTableRow = memo(function PetTableRow({
  pet,
  onEdit,
  onDeleteRequest,
}: PetTableRowProps) {
  return <TableRow>...</TableRow>;
});

// ✅ 安定したハンドラ（useCallback）を memo コンポーネントに渡す
const handleDeletePetRequest = useCallback((id: string, name: string) => {
  setDeletePetTarget({ id, name });
}, []);  // deps なし = stable

const clearFieldError = useCallback((field: string) => {
  setFieldErrors((prev) => {
    const next = { ...prev };
    delete next[field];
    return next;
  });
}, []);  // functional setstate なので deps なし
```

#### `rerender-functional-setstate` — 安定したコールバックのための関数型 setState

```typescript
// ✅ setState に関数形式を使うと deps から state を外せる → useCallback が安定
const handleInputChange = useCallback((field: string, value: string | boolean | number) => {
  setOwnerData(prev => ({ ...prev, [field]: value }));  // prev 参照 → 依存不要
  markDirty();
}, [setOwnerData, markDirty]);  // どちらも stable な setter

// ❌ 直接 state を参照すると deps に追加が必要 → useCallback が不安定になる
const handleInputChange = useCallback((field: string, value: string) => {
  setOwnerData({ ...ownerData, [field]: value });  // ownerData が dep に必要
}, [ownerData]);  // ownerData が変わるたびに新しい関数参照 → memo 無効
```

#### `rerender-lazy-state-init` — 高コストな初期化を lazy に

```typescript
// ✅ 関数を渡すと初回レンダー時のみ実行される
const [ownerData, setOwnerData] = useState<OwnerData>(
  () => initialOwner ? mapOwnerToFormData(initialOwner) : DEFAULT_OWNER_DATA
);

// ❌ 直接値を渡すと毎レンダーで mapOwnerToFormData が実行される
const [ownerData, setOwnerData] = useState<OwnerData>(
  initialOwner ? mapOwnerToFormData(initialOwner) : DEFAULT_OWNER_DATA  // ❌
);
```

#### `rerender-transitions` + `useDeferredValue` — 非緊急更新の遅延

```typescript
// ✅ 大量データのフィルタリング（OwnersList.tsx の参照実装）
const [searchTerm, setSearchTerm] = useState("");
const deferredSearchTerm = useDeferredValue(searchTerm);

// searchTerm: 入力は即座に反映（入力ブロッキングなし）
// deferredSearchTerm: ブラウザがアイドル時にフィルタリングを遅延実行
const filteredPets = useMemo(() => {
  if (!deferredSearchTerm) return pets;
  return pets.filter(pet => pet.ownerName.toLowerCase().includes(deferredSearchTerm));
}, [pets, deferredSearchTerm]);

// フィルタ遅延中の視覚フィードバック
const isFiltering = searchTerm !== deferredSearchTerm;
// <div className={isFiltering ? "opacity-60" : ""}>

// ✅ API ミューテーションに useTransition
const [isSavePending, startSaveTransition] = useTransition();
startSaveTransition(async () => {
  await saveData();
});
```

#### `rerender-dependencies` — useCallback の deps にオブジェクトを使わない

```typescript
// ✅ オブジェクトからプリミティブを抽出して deps に使う
const pendingDeleteOwnerId = pendingDeleteOwner?.id ?? null;  // string | null

const handleConfirmDelete = useCallback(() => {
  if (!pendingDeleteOwnerId) return;
  deleteOwner(pendingDeleteOwnerId);
}, [pendingDeleteOwnerId]);  // ✅ primitive

// ❌ オブジェクト参照を deps に入れると毎回新しい関数が生成される
const handleConfirmDelete = useCallback(() => {
  if (!pendingDeleteOwner?.id) return;
  deleteOwner(pendingDeleteOwner.id);
}, [pendingDeleteOwner]);  // ❌ オブジェクト参照
```

---

### 12.2 Bundle 最適化

#### `bundle-dynamic-imports` — 重いモーダルを lazy load

```typescript
// ✅ 初回 open 時のみチャンクをフェッチ（OwnerForm.tsx の参照実装）
const PetEditModal = lazy(() =>
  import("../components/PetEditModal").then(m => ({ default: m.PetEditModal }))
);

// Suspense でラップ必須
<Suspense fallback={null}>
  <PetEditModal open={petModalOpen} ... />
</Suspense>
```

#### `bundle-feature-indexing` — feature の外からは必ず `index.ts` 経由、内部は直接ファイル指定

import ルールは **呼び出し元が feature の外か内か** で使い分ける。

```typescript
// ── feature 外からの import（app/, hooks/, components/shared/ など） ──

// ✅ feature の index.ts 経由（Feature Indexing — MANDATORY）
import { useGetOwners, OwnerCard } from "@/features/owners";

// ❌ deep import 禁止（index.ts を迂回）
import { useGetOwners } from "@/features/owners/api/get-owners";  // ❌
import { OwnerCard } from "@/features/owners/components/OwnerCard"; // ❌

// ── feature 内部からの import（同一 feature 内） ──

// ✅ 直接ファイルを指定（tree-shaking が効く）
import { deleteOwner } from "../api/delete-owner";
import { formatDate } from "@/lib/format/date";

// ❌ feature 内部で自 feature の index.ts を経由するのは不要な迂回
import { deleteOwner } from "../api";  // ❌ feature 内では直接ファイル指定
```

**整理**:
- **feature 外 → feature**: `index.ts` 経由（Public API として公開されたものだけを使う）
- **feature 内部**: 直接ファイル指定（不要なモジュールを bundle に含めない）
- `features/xxx/api/index.ts` 等の barrel ファイルは **外部公開 API のカタログ** として維持し、feature 内での自己参照には使わない

---

### 12.3 レンダリングパフォーマンス

#### `rendering-hoist-jsx` — 静的な JSX をモジュールレベル定数に巻き上げる

```typescript
// ✅ 静的な JSX はコンポーネント外に定数として定義（OwnerForm.tsx の参照実装）
// → コンポーネントが何回レンダーされても JSX ノードを再生成しない
const PET_TABLE_HEADER = (
  <TableHeader>
    <TableRow>
      <TableHead>ペット番号</TableHead>
      <TableHead>ペット名</TableHead>
      {/* ... */}
    </TableRow>
  </TableHeader>
);

// Select の選択肢も同様（enum 値から生成）
const GENDER_SELECT_ITEMS = PET_GENDER_VALUES.map((g) => (
  <SelectItem key={g} value={g}>{g}</SelectItem>
));

// API 由来のデータは useMemo（React Query キャッシュ）
const animalSpeciesSelectItems = useMemo(() =>
  animalSpeciesList.map((s) => (
    <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>
  )),
  [animalSpeciesList]
);
```

#### `js-cache-function-results` — API 由来の JSX リストは `useMemo` でキャッシュ

静的な定数（`rendering-hoist-jsx`）と異なり、API から取得したデータ由来の JSX リストは
`useMemo` でキャッシュする。これにより不要な JSX ノードの再生成を防ぐ。

```typescript
// ✅ API データ由来の JSX リスト（PetEditModal.tsx の参照実装）
// animalSpeciesList が変わらない限り SelectItem を再生成しない
const animalSpeciesSelectItems = useMemo(() =>
  animalSpeciesList.map((s) => (
    <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>
  )),
  [animalSpeciesList]
);

// ❌ レンダーのたびに SelectItem を再生成する
return (
  <SelectContent>
    {animalSpeciesList.map((s) => (           // ← レンダーごとに新しいノード
      <SelectItem key={s.id} value={String(s.id)}>{s.name}</SelectItem>
    ))}
  </SelectContent>
);
```

**静的定数 vs useMemo の使い分け:**

| データ元 | 手法 |
|---------|------|
| コンパイル時に確定する定数（enum 値等） | モジュールレベル定数 `const FOO = [...]` |
| API / React Query から取得するリスト | `useMemo([list])` |

#### `rendering-conditional-render` — 条件付きレンダリングは必ず三項演算子

```typescript
// ✅ 常に ternary + null
{pet.status ? <StatusBadge>{pet.status}</StatusBadge> : null}
{pagination.totalPages > 1 ? <Pagination ... /> : null}

// ❌ && 演算子（数値 0 や空文字を意図せずレンダリングする危険）
{count && <span>{count}</span>}         // count=0 のとき "0" が表示される ❌
{pagination.totalPages > 1 && <Pagination />}  // boolean なら問題ないが不統一 ❌
```

---

### 12.4 非同期・データローディング

#### `async-parallel` — loader での並列フェッチ

```typescript
// ✅ React Router loader での並列ページフェッチ（loaders.ts の参照実装）
export const ownersLoader = async (): Promise<OwnersLoaderData> => {
  // page 1 で総件数を確認
  const { data: firstPage } = await axios.get<PetsResponse>("/v1/pets", {
    params: { page: 1, limit: PER_PAGE },
  });

  const totalPages = Math.ceil(firstPage.total / PER_PAGE);

  // 残りのページを Promise.all で並列フェッチ
  const remainingPages = await Promise.all(
    Array.from({ length: totalPages - 1 }, (_, i) =>
      axios.get<PetsResponse>("/v1/pets", { params: { page: i + 2, limit: PER_PAGE } })
        .then(r => r.data)
    )
  );

  return {
    pets: [firstPage, ...remainingPages].flatMap(page =>
      page.data.map(transformBackendPetToFrontend)
    ),
  };
};

// ✅ 複数の独立した非同期処理も Promise.allSettled で並列実行
const results = await Promise.allSettled(
  pendingPets.map(pet => createPet(transformCreatePetRequest(pet)))
);
const failedCount = results.filter(r => r.status === "rejected").length;
```

---

### 12.5 パフォーマンスチェックリスト

新規ページ・大型コンポーネント実装時の確認事項：

- [ ] フォーム/テーブル内の独立したセクションを `memo()` で分断しているか
- [ ] `memo()` コンポーネントに渡すハンドラは `useCallback` で安定化しているか
- [ ] `useCallback` の deps にオブジェクト全体を入れていないか（primitive を抽出）
- [ ] `useState` の初期値は高コストな場合 `() => ...` の lazy init 形式か
- [ ] 検索/フィルタは `useDeferredValue` で UI ブロッキングを防いでいるか
- [ ] フォーム送信は `useActionState`（`isPending`）、フォーム外の非同期更新（リスト再取得・削除・ナビゲーション）は `useTransition` で pending 管理しているか
- [ ] 重いモーダル/ダイアログは `lazy()` + `Suspense` で遅延ロードしているか
- [ ] static な JSX（enum 由来の SelectItem 一覧等）はモジュール定数に巻き上げているか（`rendering-hoist-jsx`）
- [ ] API 由来の JSX リストは `useMemo([list])` でキャッシュしているか（`js-cache-function-results`）
- [ ] 条件付きレンダリングはすべて `? (...) : null` 形式か（`&&` は使わない）
- [ ] feature 間の import は `index.ts`（barrel）経由か（deep import 禁止・§1.3）。barrel には re-export のみを置き実装ロジックを混ぜない（tree-shaking を阻害しない）
- [ ] loader 内で独立したフェッチは `Promise.all` で並列化しているか

---

## 13. 参照

- [React 19 Release Notes](https://react.dev/blog/2024/12/05/react-19)
- [Bulletproof React](https://github.com/alan2207/bulletproof-react)
- [shadcn/ui](https://ui.shadcn.com/)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Vitest](https://vitest.dev/)
- [React Router](https://reactrouter.com/)
- [TanStack Query](https://tanstack.com/query/latest)
