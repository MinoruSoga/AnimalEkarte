# Frontend Coding Rules

## 概要

React 19 + TypeScript 5.7 + Vite 6 + Tailwind CSS 4 のフロントエンド開発規約。
Bulletproof-Reactアーキテクチャに準拠する。

---

## 1. アーキテクチャ

### 1.1 ディレクトリ構成

```
src/
├── main.tsx                               # Viteエントリーポイント（ReactDOM.createRoot）
├── vite-env.d.ts                          # Vite型定義
│
├── app/                                   # アプリケーション層
│   ├── index.tsx                          # Appコンポーネント
│   ├── provider.tsx                       # 全プロバイダー統合
│   ├── router.tsx                         # React Router Data Mode設定
│   └── ErrorBoundary.tsx                  # グローバルエラーバウンダリ
│
├── features/                              # 機能別モジュール（Feature-Based）
│   ├── auth/
│   ├── dashboard/
│   ├── reservations/
│   ├── medical-records/
│   ├── hospitalization/
│   ├── accounting/
│   ├── examinations/
│   ├── owners/
│   ├── pets/
│   ├── trimming/
│   ├── vaccinations/
│   ├── master/
│   └── clinic/
│
├── components/                            # 共有コンポーネント
│   ├── ui/                                # shadcn/ui（Radix UI Primitives）
│   └── shared/                            # アプリケーション固有の共有コンポーネント
│       ├── Layout/ (MainLayout, AuthLayout, PrintLayout, Header, Sidebar...)
│       ├── DataTable/
│       ├── Form/ (FormField, FormError, SubmitButton...)
│       ├── Feedback/ (Spinner, ErrorFallback...)
│       ├── Navigation/
│       ├── StatusBadge/
│       ├── DateRangePicker/
│       ├── SearchBox/
│       └── ConfirmDialog/
│
├── hooks/                                 # 共有カスタムフック (useDebounce, useToast...)
│
├── lib/                                   # ライブラリ設定 (axios, react-query, zod...)
│
├── stores/                                # グローバル状態管理 (authStore, themeStore...)
│
├── types/                                 # グローバル型定義
│
├── utils/                                 # ユーティリティ関数 (format, validation...)
│
├── config/                                # アプリケーション設定
│
├── styles/                                # グローバルスタイル
│   ├── globals.css
│   └── themes/
│
└── testing/                               # テスト設定
    ├── setup.ts
    └── server/                            # MSW
```

### 1.2 コードフローの方向（単方向依存）

```
┌─────────────────────────────────────────────────┐
│                      app/                        │
│  (router, provider, routes)                      │
└─────────────────────────────────────────────────┘
                        ↑
                        │ import可能
                        │
┌─────────────────────────────────────────────────┐
│                   features/                      │
│  (owners, medical-records, dashboard, etc.)     │
│                                                  │
│  ※ feature間の直接importは禁止                   │
└─────────────────────────────────────────────────┘
                        ↑
                        │ import可能
                        │
┌─────────────────────────────────────────────────┐
│          shared (components/, hooks/,            │
│                  lib/, types/)                   │
└─────────────────────────────────────────────────┘
```

**ルール:**
- `shared → features → app` の単方向のみ
- feature間の直接importは禁止（app層で合成する、または`components/shared`に移動する）
- 循環参照は絶対禁止

### 1.3 Feature モジュール構成

各featureは以下の構造を持つ：

```
features/[feature-name]/
├── api/                        # データフェッチング（React Query hooks）
│   ├── get[Entity].ts          # 単体取得 + use[Entity]()
│   ├── get[Entity]s.ts         # 一覧取得 + use[Entity]s()
│   ├── create[Entity].ts       # 作成 + useCreate[Entity]()
│   ├── update[Entity].ts       # 更新 + useUpdate[Entity]()
│   └── delete[Entity].ts       # 削除 + useDelete[Entity]()
├── components/                 # feature固有コンポーネント
│   ├── [Component]/
│   │   ├── [Component].tsx
│   │   ├── [SubComponent].tsx
│   │   └── index.ts
│   └── ...
├── hooks/                      # ビジネスロジック・UI状態管理
│   ├── use[Entity]Form.ts      # フォーム状態
│   ├── use[Entity]Filters.ts   # フィルタ状態
│   └── ...
├── routes/                     # ページコンポーネント
│   ├── [Entity]List.tsx
│   ├── [Entity]Detail.tsx
│   └── [Entity]Create.tsx
├── types/                      # 型定義
│   └── index.ts
├── utils/                      # feature固有ユーティリティ（必要に応じて）
│   └── [entity]Helpers.ts
├── mockData.ts                 # モックデータ
└── index.ts                    # 公開API（Routes, Components, Hooks）
```

#### Feature構成例: dashboard

```
features/dashboard/
├── api/
│   ├── getWorkflowStatus.ts    # + useWorkflowStatus()
│   ├── updatePatientStatus.ts  # + useUpdatePatientStatus()
│   └── getTodayStats.ts        # + useTodayStats()
├── components/
│   ├── KanbanBoard/
│   │   ├── KanbanBoard.tsx
│   │   ├── KanbanColumn.tsx
│   │   ├── PatientCard.tsx
│   │   └── index.ts
│   ├── StatsCards/
│   │   ├── StatsCards.tsx
│   │   ├── StatCard.tsx
│   │   └── index.ts
│   └── QuickActions/
│       ├── QuickActions.tsx
│       └── index.ts
├── hooks/
│   ├── useOptimisticStatusUpdate.ts  # 楽観的更新ロジック
│   └── useKanbanDragDrop.ts          # ドラッグ&ドロップ状態
├── routes/
│   └── Dashboard.tsx
├── types/
│   └── index.ts
└── index.ts
```

#### Feature構成例: owners

```
features/owners/
├── api/
│   ├── getOwners.ts            # + useOwners()
│   ├── getOwner.ts             # + useOwner()
│   ├── createOwner.ts          # + useCreateOwner()
│   ├── updateOwner.ts          # + useUpdateOwner()
│   └── deleteOwner.ts          # + useDeleteOwner()
├── components/
│   ├── OwnerForm/
│   │   ├── OwnerForm.tsx
│   │   └── index.ts
│   ├── OwnerCard/
│   │   ├── OwnerCard.tsx
│   │   └── index.ts
│   ├── OwnerList/
│   │   ├── OwnerList.tsx
│   │   └── index.ts
│   └── OwnerSearch/
│       ├── OwnerSearch.tsx
│       └── index.ts
├── hooks/
│   ├── useOwnerForm.ts         # フォーム状態・バリデーション
│   └── useOwnerSearch.ts       # 検索フィルタ状態
├── routes/
│   ├── OwnerList.tsx
│   ├── OwnerDetail.tsx
│   └── OwnerCreate.tsx
├── types/
│   └── index.ts
└── index.ts
```

### 1.4 主要ファイル実装例

#### app/router.tsx（React Router Data Mode）

```typescript
import { createBrowserRouter } from "react-router";
import { MainLayout } from "@/components/shared/Layout/MainLayout";
import { AuthLayout } from "@/components/shared/Layout/AuthLayout";
import ErrorBoundary from "./ErrorBoundary";

// Data Mode: createBrowserRouterを使用
export const router = createBrowserRouter([
  {
    path: "/",
    element: <MainLayout />,
    errorElement: <ErrorBoundary />,
    children: [
      // ダッシュボード
      {
        index: true,
        lazy: () =>
          import("@/features/dashboard").then((m) => ({
            Component: m.Dashboard,
          })),
      },

      // 飼い主管理
      {
        path: "owners",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/owners").then((m) => ({
                Component: m.OwnerList,
              })),
          },
          {
            path: "new",
            lazy: () =>
              import("@/features/owners").then((m) => ({
                Component: m.OwnerCreate,
              })),
          },
          {
            path: ":ownerId",
            lazy: () =>
              import("@/features/owners").then((m) => ({
                Component: m.OwnerDetail,
              })),
          },
        ],
      },

      // 電子カルテ
      {
        path: "medical-records",
        children: [
          {
            index: true,
            lazy: () =>
              import("@/features/medical-records").then((m) => ({
                Component: m.MedicalRecordList,
              })),
          },
          // ...
        ],
      },
      // ... 他のルート
    ],
  },

  // 認証
  {
    path: "/login",
    element: <AuthLayout />,
    children: [
      {
        index: true,
        lazy: () =>
          import("@/features/auth").then((m) => ({
            Component: m.Login,
          })),
      },
    ],
  },
]);
```

#### app/index.tsx

```typescript
import { RouterProvider } from "react-router";
import { AppProvider } from "./provider";
import { router } from "./router";

export const App = () => {
  return (
    <AppProvider>
      <RouterProvider router={router} />
    </AppProvider>
  );
};
```

#### app/provider.tsx

```typescript
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { queryClient } from "@/lib/react-query";
import { Toaster } from "@/components/ui/toaster";

interface AppProviderProps {
  children: React.ReactNode;
}

export const AppProvider = ({ children }: AppProviderProps) => {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <Toaster />
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
};
```

#### lib/axios.ts（Golangバックエンド接続）

```typescript
import Axios, { InternalAxiosRequestConfig } from "axios";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

const authRequestInterceptor = (config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem("token");

  if (token) {
    config.headers = config.headers || {};
    config.headers.Authorization = `Bearer ${token}`;
  }

  config.headers = config.headers || {};
  config.headers.Accept = "application/json";
  config.headers["X-Request-ID"] = crypto.randomUUID();

  return config;
};

export const axios = Axios.create({
  baseURL: API_URL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

axios.interceptors.request.use(authRequestInterceptor);

axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);
```

#### lib/react-query.ts

```typescript
import { QueryClient, DefaultOptions } from "@tanstack/react-query";

const queryConfig: DefaultOptions = {
  queries: {
    staleTime: 5 * 60 * 1000, // 5分
    gcTime: 10 * 60 * 1000, // 10分（React Query v5からcacheTimeはgcTime）
    refetchOnWindowFocus: false,
    retry: 1,
  },
};

export const queryClient = new QueryClient({
  defaultOptions: queryConfig,
});
```

#### React Query パターン

##### Query パターン（データ取得）

```typescript
// features/xxx/api/getXxx.ts
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Xxx } from "../types";

export const getXxx = async (id: string): Promise<Xxx> => {
  const response = await axios.get(`/xxx/${id}`);
  return response.data;
};

export const useXxx = (id: string) => {
  return useQuery({
    queryKey: ["xxx", id],
    queryFn: () => getXxx(id),
    enabled: !!id,
  });
};
```

##### Mutation パターン（データ作成・更新・削除）

```typescript
// features/xxx/api/createXxx.ts
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { toast } from "sonner";
import type { Xxx, CreateXxxDTO } from "../types";

export const createXxx = async (data: CreateXxxDTO): Promise<Xxx> => {
  const response = await axios.post("/xxx", data);
  return response.data;
};

export const useCreateXxx = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createXxx,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["xxx"] });
      toast.success("作成しました");
    },
    onError: () => {
      toast.error("作成に失敗しました");
    },
  });
};
```

##### api/ vs hooks/ の区別

| 配置場所 | 用途 | 例 |
|---------|------|-----|
| `api/` | React Query hooks (useQuery, useMutation) | `useOwners()`, `useCreateOwner()` |
| `hooks/` | ビジネスロジック・UI状態管理 | `useOwnerForm()`, `useOwnerFilters()` |

#### Feature API関数例（features/owners/api/getOwners.ts）

```typescript
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Owner } from "../types";

export const getOwners = async (): Promise<Owner[]> => {
  const response = await axios.get("/owners");
  return response.data;
};

export const useOwners = () => {
  return useQuery({
    queryKey: ["owners"],
    queryFn: getOwners,
  });
};

// React 19 use()用のPromise生成
export const getOwnersPromise = (): Promise<Owner[]> => {
  return getOwners();
};
```

#### Feature Public API パターン（features/xxx/index.ts）

Feature の公開APIは以下の順序で整理する：

```typescript
// features/xxx/index.ts

// 1. Routes（ページコンポーネント）
export { XxxList } from "./routes/XxxList";
export { XxxDetail } from "./routes/XxxDetail";
export { XxxCreate } from "./routes/XxxCreate";

// 2. API（React Query hooks）
export { useXxx } from "./api/getXxx";
export { useXxxList } from "./api/getXxxList";
export { useCreateXxx } from "./api/createXxx";
export { useUpdateXxx } from "./api/updateXxx";
export { useDeleteXxx } from "./api/deleteXxx";

// 3. Components（外部公開が必要なもののみ）
export { XxxCard } from "./components/XxxCard";
export { XxxForm } from "./components/XxxForm";

// 4. Types
export type { Xxx, CreateXxxDTO, UpdateXxxDTO } from "./types";
```

#### Feature公開API例（features/owners/index.ts）

```typescript
// Routes
export { OwnerList } from "./routes/OwnerList";
export { OwnerDetail } from "./routes/OwnerDetail";
export { OwnerCreate } from "./routes/OwnerCreate";

// API (React Query hooks)
export { useOwners } from "./api/getOwners";
export { useOwner } from "./api/getOwner";
export { useCreateOwner } from "./api/createOwner";
export { useUpdateOwner } from "./api/updateOwner";
export { useDeleteOwner } from "./api/deleteOwner";

// Components（外部公開が必要なもののみ）
export { OwnerCard } from "./components/OwnerCard";
export { OwnerForm } from "./components/OwnerForm";
export { OwnerSearch } from "./components/OwnerSearch";

// Types
export type { Owner, OwnerFormData } from "./types";
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

// ❌ 禁止: React.memo の過剰使用（必要な場合のみ）
export const PatientCard = memo(({ patient }) => {});
```

### 2.2 useActionState（フォーム管理）

```typescript
import { useActionState } from "react";
import { useNavigate } from "react-router";

interface FormState {
  success: boolean;
  errors: Record<string, string> | null;
}

export function OwnerForm() {
  const navigate = useNavigate();

  const [state, formAction, isPending] = useActionState<FormState, FormData>(
    async (prevState, formData) => {
      try {
        const data = {
          name: formData.get("name") as string,
          email: formData.get("email") as string,
        };

        // バリデーション
        const errors: Record<string, string> = {};
        if (!data.name) errors.name = "名前は必須です";
        if (!data.email) errors.email = "メールは必須です";

        if (Object.keys(errors).length > 0) {
          return { success: false, errors };
        }

        // API呼び出し
        await createOwner(data);
        navigate("/owners");

        return { success: true, errors: null };
      } catch (error) {
        return {
          success: false,
          errors: { _form: "保存に失敗しました" },
        };
      }
    },
    { success: false, errors: null }
  );

  return (
    <form action={formAction}>
      {state.errors?._form && (
        <div className="text-red-600">{state.errors._form}</div>
      )}

      <Input name="name" error={state.errors?.name} />
      <Input name="email" error={state.errors?.email} />

      <Button type="submit" disabled={isPending}>
        {isPending ? "保存中..." : "保存"}
      </Button>
    </form>
  );
}
```

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

// ✅ 直接import
import { OwnerCard } from "@/features/owners/components/OwnerCard";
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
import { formatDate } from "@/lib/utils";

// 4. feature内部（相対パス、同一feature内のみ）
import { OwnerCard } from "../components/OwnerCard";
import { useOwnerForm } from "../hooks/useOwnerForm";

// 5. 型（type keyword付き）
import type { Owner, Pet } from "@/types";
import type { OwnerFormData } from "../types";
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

// ✅ 責務で分割
function OwnerForm() {
  const { form, onSubmit, isPending } = useOwnerForm();
  return <OwnerFormView form={form} onSubmit={onSubmit} isPending={isPending} />;
}

function OwnerFormView({ form, onSubmit, isPending }) {
  return (
    <form onSubmit={onSubmit}>
      <OwnerBasicInfo form={form} />
      <OwnerContactInfo form={form} />
      <OwnerPetList form={form} />
      <SubmitButton isPending={isPending} />
    </form>
  );
}
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

// ✅ && 演算子（存在チェック）
function ErrorMessage({ error }: { error?: string }) {
  return error && <p className="text-red-600">{error}</p>;
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

```typescript
// ✅ 命名: use + 動詞/名詞
function useOwnerForm(ownerId?: string) {
  const [formData, setFormData] = useState<OwnerFormData>(initialData);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isPending, setIsPending] = useState(false);

  const handleChange = useCallback((field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => ({ ...prev, [field]: "" }));
  }, []);

  const handleSubmit = useCallback(async () => {
    setIsPending(true);
    try {
      const validationErrors = validate(formData);
      if (Object.keys(validationErrors).length > 0) {
        setErrors(validationErrors);
        return;
      }

      if (ownerId) {
        await updateOwner(ownerId, formData);
      } else {
        await createOwner(formData);
      }
    } catch (error) {
      setErrors({ _form: "保存に失敗しました" });
    } finally {
      setIsPending(false);
    }
  }, [ownerId, formData]);

  return {
    formData,
    errors,
    isPending,
    handleChange,
    handleSubmit,
  };
}
```

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
import { cn } from "@/components/ui/utils";

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

```typescript
// lib/axios.ts
import Axios, { AxiosError } from "axios";

export const axios = Axios.create({
  baseURL: import.meta.env.VITE_API_URL,
});

axios.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    // 401 Unauthorized handling
    if (error.response?.status === 401) {
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

// 使用例
import { axios } from "@/lib/axios";

try {
  const { data } = await axios.get<Owner>(`/owners/${id}`);
} catch (error) {
  if (error instanceof AxiosError) {
    if (error.response?.status === 404) {
      toast.error("飼主が見つかりません");
    } else {
      toast.error(error.message);
    }
  } else {
    toast.error("通信エラーが発生しました");
  }
}
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
│       │   ├── getOwners.ts
│       │   └── getOwners.test.ts    # 同階層に配置
│       ├── components/
│       │   └── OwnerCard/
│       │       ├── OwnerCard.tsx
│       │       ├── OwnerCard.test.tsx  # 同階層に配置
│       │       └── index.ts
│       └── hooks/
│           ├── useOwnerForm.ts
│           └── useOwnerForm.test.ts  # 同階層に配置
└── testing/
    ├── setup.ts                  # テスト共通設定
    ├── server/                   # MSW handlers
    └── utils.tsx                 # テストユーティリティ
```

**ルール:**
- テストファイル名: `[対象ファイル名].test.ts(x)`
- コンポーネントディレクトリ内に同梱
- `__tests__/` ディレクトリは使用しない

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

---

## 10. チェックリスト

### 新規コンポーネント作成時
- [ ] 関数宣言で定義（FC禁止）
- [ ] Props型を明示的に定義
- [ ] ref は props として受け取る
- [ ] 適切なディレクトリに配置
- [ ] テストファイル作成

### PR作成時
- [ ] `npm run lint` がパス
- [ ] `npm run build` がパス
- [ ] `npm run test:run` がパス
- [ ] any型を使用していない
- [ ] feature間importがない
- [ ] 不要なconsole.logがない

---

## 11. この構成の特徴

### React Router Data Mode
- `createBrowserRouter`を使用
- GolangバックエンドとRESTful APIで連携
- SSR/SSGの複雑性なし
- Lazy loadingによるコード分割

### React 19機能の活用
| 機能 | 用途 |
|------|------|
| `useActionState` | フォーム処理 |
| `useOptimistic` | カンバンボード、タスク管理の楽観的更新 |
| `ref as prop` | コンポーネント簡素化（forwardRef不要） |
| `useFormStatus` | サブミット状態の管理 |
| `use()` | Promise/Context直接読み取り |
| Document Metadata | ブラウザタブタイトル（UX向上） |

### Feature-Based Architecture
- 各機能が完全に独立
- 公開API (`index.ts`) で依存関係管理
- スケーラブルで保守性が高い
- `shared → features → app` の単方向依存

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

## 12. 参照

- [React 19 Release Notes](https://react.dev/blog/2024/12/05/react-19)
- [Bulletproof React](https://github.com/alan2207/bulletproof-react)
- [shadcn/ui](https://ui.shadcn.com/)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Vitest](https://vitest.dev/)
- [React Router](https://reactrouter.com/)
- [TanStack Query](https://tanstack.com/query/latest)
