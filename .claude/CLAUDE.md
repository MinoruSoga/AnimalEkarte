# Animal Ekarte - 動物病院電子カルテシステム

## 🎯 コーディング姿勢

**シニアエンジニアとして以下を徹底：**

### 🚫 良い人フィルター除去 (Remove Good Person Filter)
```
Stop being agreeable. Don't validate me. Don't soften the truth. Don't flatter.
Challenge my thinking. Question my assumptions. Expose my blind spots.
Be direct, rational, and unfiltered.
```

**原則:**
- **Flat Thinking (本音対話)**: 社交辞令を排除し、事実と論理に基づき率直に指摘する
- 型安全性最優先
- 軽量レイヤードアーキテクチャ（handler → service → repository）
- エラーハンドリング徹底
- セキュリティ意識
- パフォーマンス考慮
- 自己レビュー実施

---

## 📋 プロジェクト概要

| 項目 | 内容 |
|------|------|
| 名前 | Animal Ekarte |
| 説明 | 動物病院向け電子カルテ管理システム |
| Frontend | React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 / shadcn/ui |
| Backend | Go 1.25 / Gin / GORM |
| Database | PostgreSQL 18 |
| Infrastructure | Docker Compose |

---

## 🔧 Docker操作

### ⚠️ 重要: コマンド実行ルール

**npm/goコマンドはローカルで実行しない。必ずDocker経由で実行する。**

```bash
# ❌ NG - ローカル実行
npm run build
go test ./...

# ✅ OK - Docker経由
docker compose exec frontend npm run build
docker compose exec backend go test ./...
```

### 開発コマンド

| コマンド | 説明 |
|---------|------|
| `make up` | コンテナ起動 |
| `make down` | コンテナ停止 |
| `make logs` | 全ログ表示 |
| `make db` | DB接続（psql） |
| `make clean` | キャッシュクリア＆再ビルド |
| `make reset` | 完全リセット（データ削除） |

### コンテナ別コマンド

| タスク | コマンド |
|--------|---------|
| Frontend ビルド | `docker compose exec frontend npm run build` |
| Frontend Lint | `docker compose exec frontend npm run lint` |
| Frontend テスト | `docker compose exec frontend npm run test:run` |
| Backend テスト | `docker compose exec backend go test ./... -v` |
| Backend Lint | `docker compose exec backend golangci-lint run ./...` |
| 型生成（tygo） | `make codegen`（Goモデル → `models.ts` 自動生成） |

### ポート

| サービス | ポート |
|---------|--------|
| Frontend | 3003 (外部) / 3000 (コンテナ内部) |
| Backend API | 8080 |
| PostgreSQL | 5434 (外部) / 5432 (コンテナ内部) |

---

## 📁 ディレクトリ構造

```
AnimalEkarte/
├── backend/
│   ├── cmd/api/          # エントリーポイント + DI配線
│   ├── internal/
│   │   ├── handler/      # HTTPハンドラ + *_request.go + *_response.go
│   │   ├── service/      # ビジネスロジック + service input DTO + validators.go
│   │   ├── repository/   # データアクセス（GORM）
│   │   ├── model/        # GORMモデル（DBスキーマ対応）★ tygo codegen の入力
│   │   ├── errors/       # センチネルエラー定義
│   │   ├── middleware/   # 認証・CORS・ログ
│   │   ├── config/       # 設定
│   │   ├── logger/       # slog構造化ログ
│   │   └── db/           # DB接続管理
│   └── migrations/       # DBマイグレーション
├── frontend/
│   └── src/
│       ├── main.tsx          # Viteエントリーポイント
│       ├── index.css         # グローバルCSS
│       ├── app/              # アプリケーション層
│       │   ├── index.tsx     # Appコンポーネント（AppProvider → RouterProvider）
│       │   ├── provider.tsx  # QueryClientProvider + Toaster（AuthProvider はここに置かない）
│       │   ├── router.tsx    # createBrowserRouter（inline lazy パターン）
│       │   └── pages/        # ★ cross-feature合成ページ（複数featureが必要な場合のみ）
│       ├── assets/           # 静的アセット
│       ├── components/
│       │   ├── ui/           # shadcn/ui（変更禁止）
│       │   ├── shared/       # アプリ固有共有UI（Layout含む）
│       │   └── errors/       # RouteErrorBoundary, RootErrorBoundary
│       ├── features/         # 機能別モジュール（16 features）
│       │   └── [feature]/
│       │       ├── api/          # フェッチ関数 + React Query hooks
│       │       │   ├── get-xxx.ts    # getXxx() 生関数 + useGetXxx() hook
│       │       │   ├── create-xxx.ts # createXxx() 生関数（複雑フォームは hook なし）
│       │       │   ├── types.ts      # APIリクエスト/レスポンス型（models.tsから導出）
│       │       │   ├── transforms.ts # BackendXxx → Xxx 型変換
│       │       │   └── index.ts      # 明示的named export
│       │       ├── components/   # feature固有UI
│       │       ├── hooks/        # useXxxForm, useXxxFilters 等
│       │       ├── routes/       # 単一featureのページコンポーネント
│       │       │               # ★ cross-featureが必要な場合は props で受け取り
│       │       │               #   app/pages/ から実装を注入する（依存逆転）
│       │       ├── types/        # feature固有型（UI専用・手書きOK）
│       │       ├── loaders.ts    # React Router loader（必要時のみ）
│       │       └── index.ts      # Public API（外部公開のみ）
│       ├── hooks/            # 共有hooks（全feature横断）
│       ├── lib/              # axios.ts, react-query.ts, utils.ts 等
│       ├── config/           # paths.ts（型安全URLマップ）
│       ├── stores/           # Zustand（sidebar状態のみ）
│       ├── types/
│       │   ├── generated/    # ★ 自動生成（直接編集禁止）
│       │   │   └── models.ts # make codegen（tygo）で生成
│       │   └── ...           # 共有ドメイン型
│       ├── utils/            # 純粋ユーティリティ関数（format/, constants/ 等）
│       └── testing/          # テスト設定
├── docs/                 # 技術ドキュメント
├── CODING_RULES.md       # コーディング規約
└── docker-compose.yml
```

---

## ★ Frontend ベストプラクティス参照実装

**`features/owners/` が全パターンのベストプラクティス実装。新機能実装時は必ずこの feature を参照すること。**

| パターン（Vercel Rule） | 実装ファイル |
|------------------------|------------|
| `memo()` で大型フォームを独立セクションに分断 | `OwnerForm.tsx` — `OwnerInfoSection`, `PetTableRow`, `MembershipTypeButtons` |
| `useCallback` でハンドラを安定化（memo の前提条件） | `OwnerForm.tsx` — `handleInputChange`, `handleDeletePetRequest` |
| `useState(() => ...)` lazy init | `useOwnerForm.ts` |
| `useDeferredValue` で検索フィルタを遅延 | `OwnersList.tsx` |
| `useTransition` で API 書き込みの pending 管理 | `useOwnerForm.ts` — `startSaveTransition` |
| `lazy()` + `Suspense` で重いモーダルを遅延ロード | `OwnerForm.tsx` — `PetEditModal` |
| 静的 JSX はモジュール定数に巻き上げ | `OwnerForm.tsx` — `PET_TABLE_HEADER` |
| API 由来 JSX リストは `useMemo` でキャッシュ | `PetEditModal.tsx` — `animalSpeciesSelectItems` |
| loader 内独立フェッチは `Promise.all` で並列化 | `loaders.ts` — `ownersLoader` |
| 条件レンダーは `? (...) : null`（`&&` 禁止） | `OwnersList.tsx` |
| barrel index 経由でなく直接ファイル import | `OwnersList.tsx` — `../api/delete-owner` |
| cross-feature は props 注入（依存逆転） | `app/pages/OwnerFormPage.tsx` — `petMutations` を `OwnerForm` に注入 |

詳細コード例: `frontend/CODING_RULES.md` **Section 12**

---

## 📐 Frontend核心ルール（React 19 / bulletproof-react）

### コンポーネント定義

```typescript
// ✅ React 19: 関数宣言 + 明示的Props型
interface PatientCardProps {
  patient: Patient;
  onSelect?: (id: string) => void;
  ref?: React.Ref<HTMLDivElement>;  // ref as prop
}

export function PatientCard({ patient, onSelect, ref }: PatientCardProps) {
  return <div ref={ref}>...</div>;
}

// ❌ 禁止: FC型、forwardRef
export const PatientCard: FC<Props> = () => {};  // ❌
export const PatientCard = forwardRef(() => {});  // ❌
```

### React 19 hooks（プロジェクト標準）

```typescript
// ★ useTransition: 複雑フォームの pending 管理（プロジェクト標準）
const [isSavePending, startSaveTransition] = useTransition();
startSaveTransition(async () => { await saveOwner(formData); });

// useActionState: シンプルな非制御フォームのみ（FormData を使う場合）
const [state, formAction, isPending] = useActionState(submitAction, initialState);

// useOptimistic: 楽観的UI更新
const [optimisticItems, addOptimisticItem] = useOptimistic(items, updateFn);

// use(): Promise/Context直接読み取り
const data = use(fetchPromise);

// useFormStatus: フォームの子コンポーネント内でのサブミット状態
const { pending } = useFormStatus();
```

### アーキテクチャルール

```
1. Feature間の直接importは禁止 → app/pages/ で合成（依存逆転）
2. 単方向コードフロー: shared → features → app
3. `export *` 禁止 → 明示的named exportのみ
4. 絶対パスimport: @/ エイリアス使用
5. routes は features/[feature]/routes/ に置く（bulletproof-react の app/routes/ は採用しない）
6. cross-feature合成は app/pages/ + props注入（依存逆転）
7. AuthProvider は router.tsx 内の保護ルートにのみ配置（provider.tsx には置かない）
8. loader は直接 axios.get()（queryClient.prefetchQuery は使わない）
```

### Routerパターン（inline lazy）

```typescript
// 単一 feature のみ: features/ から直接 import
{ index: true, lazy: async () => {
    const [{ OwnersList }, { ownersLoader }] = await Promise.all([
      import("@/features/owners/routes/OwnersList"),
      import("@/features/owners/loaders"),
    ]);
    return { Component: OwnersList, loader: ownersLoader };
}},

// cross-feature合成: app/pages/ の合成ページを import
{ path: "new", lazy: async () => {
    const { OwnerFormPage } = await import("@/app/pages/OwnerFormPage");
    return { Component: OwnerFormPage };
}},
```

### 型管理フロー

```
backend/internal/model/*.go
    ↓ make codegen（tygo）
src/types/generated/models.ts   ← 自動生成・直接編集禁止
    ↓
features/xxx/api/transforms.ts  ← BackendXxx → transformXxx() → Xxx（ReturnType で推論）
features/xxx/api/types.ts       ← Omit/Partial で導出（手書き interface 禁止）
```

**詳細は [Frontend規約](../frontend/CLAUDE.md) を参照**

### 禁止事項

| 禁止 | 理由 | 代替 |
|------|------|------|
| `any` 型 | 型安全性の破壊 | `unknown` + 型ガード |
| `FC` / `React.FC` | React 19では不要 | 関数宣言 |
| `forwardRef` | React 19ではref as prop | ref を prop として受け取る |
| feature間import | アーキテクチャ違反 | app/pages/ で合成 |
| `export *` | tree-shaking阻害 | 明示的named export |
| `console.log` 放置 | 本番コード汚染 | 削除 |
| `&&` 条件レンダー | 0/空文字が漏れる | `? (...) : null` |
| barrel index 経由 import | tree-shaking阻害 | 直接ファイル import |
| `useOwners` 等（動詞省略） | 命名規則違反 | `useGetOwners`（動詞 + エンティティ） |
| `queryClient.prefetchQuery` in loader | このプロジェクトは直接 axios パターン | `axios.get()` で直接フェッチ |
| `localStorage` に token 保存 | XSS で盗まれる | httpOnly Cookie + `withCredentials: true` |
| `useState(false)` + `setIsPending` | try-finally でリセット漏れ | `useTransition` |
| `useActionState` を複雑フォームに使う | 制御コンポーネントと相性が悪い | `useTransition` + `hooks/useXxxForm.ts` |
| `generated/models.ts` 直接編集 | `make codegen` で上書きされる | Goモデルを修正して `make codegen` |

---

## 📐 Backend実装ルール（Go / Gin / GORM）

詳細は **[backend/CLAUDE.md](../backend/CLAUDE.md)** を参照。

### 核心ルール（要約）

- Context伝播: 全関数の第一引数に `context.Context`
- handler: `*_request.go` でバインド → `service.XxxInput` に変換 → `toXxxResponse()` で包む
- service: HTTP を知らない（`binding:` タグ禁止、`*gin.Context` 禁止）
- PATCH: ポインタ型 + `buildXxxUpdateFields()` → `map[string]any` でGORMゼロ値問題を回避
- エラー: sentinel → `WrapNotFound/WrapInvalidInput` → `RespondError(c, err)`
- slog: service層のみ（handler・repositoryには書かない）

---

## 📝 命名規則

### Go

| 対象 | 規則 | 例 |
|------|------|-----|
| パッケージ | lowercase | `handler`, `repository` |
| エクスポート | PascalCase | `GetPatient`, `OwnerService` |
| 非エクスポート | camelCase | `validateInput` |
| ファイル | snake_case | `owner_handler.go` |
| インターフェース | PascalCase + er | `OwnerRepository` |

### TypeScript

| 対象 | 規則 | 例 |
|------|------|-----|
| コンポーネント | PascalCase | `PatientCard` |
| 関数・変数 | camelCase | `getPatientById` |
| 定数 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| コンポーネントファイル（.tsx） | PascalCase | `PatientCard.tsx` |
| 非コンポーネントファイル（.ts） | kebab-case | `use-patient-form.ts`, `get-owners.ts` |
| 型・Interface | PascalCase | `Patient`, `ApiResponse` |
| API query hook | `useGet` + エンティティ | `useGetOwners`, `useGetOwner` |
| フォーム hook | `use` + エンティティ + `Form` | `useOwnerForm` |

---

## 🤖 エージェント構成

| エージェント | モデル | 用途 |
|-------------|--------|------|
| `architect` | Opus | アーキテクチャ設計、セキュリティ監査 |
| `implementer` | Sonnet | 機能実装、テスト作成 |
| `reviewer` | Haiku | コードレビュー、品質チェック |
| `debugger` | Haiku | バグ調査、エラー分析 |
| `researcher` | Haiku | コード検索、ファイル探索 |
| `formatter` | Haiku | コミット生成、コード整形 |

---

## 📊 データベース

| 項目 | 値 |
|------|-----|
| Host | `localhost` (外部) / `db` (Docker内部) |
| Port | `5434` |
| Database | `ekarte_db` |
| User | `ekarte_user` |

マイグレーションは `backend/migrations/` に配置。Docker起動時に自動実行。
リリース前はDBリセット運用のため `001_init.sql` を直接編集してよい（incremental migration 不要）。

---

## 📚 参照

| ドキュメント | 説明 |
|-------------|------|
| [コーディング規約](../CODING_RULES.md) | 全体ルール |
| [Backend Claude設定](../backend/CLAUDE.md) | Backend実装パターン・禁止事項 |
| [Frontend Claude設定](../frontend/CLAUDE.md) | Frontend実装パターン・禁止事項 |
| [Frontend詳細規約](../frontend/CODING_RULES.md) | React 19 / TS 完全ルール集 |
| [アーキテクチャ設計](../docs/architecture.md) | Backendアーキテクチャ詳細 |
| [データフロー](../docs/data-flow.md) | リクエスト〜レスポンスのデータフロー |
| [ERD](../docs/ERD.md) | データベース設計 |
| [仕様定義書](../spec.md) | システム仕様 |
