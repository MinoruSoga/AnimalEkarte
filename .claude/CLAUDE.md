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
| Swagger生成 | `docker compose exec backend swag init -g cmd/api/main.go` |

### ポート

| サービス | ポート |
|---------|--------|
| Frontend | 3000 |
| Backend API | 8080 |
| PostgreSQL | 5432 |

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
│   │   ├── model/        # GORMモデル（DBスキーマ対応）
│   │   ├── errors/       # センチネルエラー定義
│   │   ├── middleware/   # 認証・CORS・ログ
│   │   ├── config/       # 設定
│   │   ├── logger/       # slog構造化ログ
│   │   └── db/           # DB接続管理
│   └── migrations/       # DBマイグレーション
├── frontend/
│   └── src/
│       ├── main.tsx      # Viteエントリーポイント
│       ├── index.css     # グローバルCSS
│       ├── app/          # アプリケーション層
│       │   ├── index.tsx     # Appコンポーネント
│       │   ├── provider.tsx  # プロバイダー統合
│       │   ├── router.tsx    # ルーター設定
│       │   └── routes/       # ルート定義
│       ├── components/
│       │   ├── ui/       # shadcn/ui
│       │   ├── shared/   # 共有UI
│       │   ├── layouts/  # Layout, Sidebar
│       │   └── errors/   # ErrorBoundary
│       ├── features/     # 機能別モジュール
│       │   └── [feature]/
│       │       ├── api/
│       │       ├── components/
│       │       ├── hooks/
│       │       ├── types/
│       │       └── routes/
│       ├── hooks/        # 共有hooks
│       ├── lib/          # ユーティリティ
│       ├── types/        # 共有型定義
│       └── testing/      # テスト設定
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
| `useTransition` で API 書き込みの pending 管理 | `useOwnerForm.ts` |
| `lazy()` + `Suspense` で重いモーダルを遅延ロード | `OwnerForm.tsx` — `PetEditModal` |
| 静的 JSX はモジュール定数に巻き上げ | `OwnerForm.tsx` — `PET_TABLE_HEADER` |
| API 由来 JSX リストは `useMemo` でキャッシュ | `PetEditModal.tsx` — `animalSpeciesSelectItems` |
| loader 内独立フェッチは `Promise.all` で並列化 | `loaders.ts` — `ownersLoader` |
| 条件レンダーは `? (...) : null`（`&&` 禁止） | `OwnersList.tsx` |
| barrel index 経由でなく直接ファイル import | `OwnersList.tsx` — `../api/delete-owner` |

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

### React 19 新hooks

```typescript
// useActionState: フォームアクション管理
const [state, formAction, isPending] = useActionState(submitAction, initialState);

// useOptimistic: 楽観的UI更新
const [optimisticItems, addOptimisticItem] = useOptimistic(items, updateFn);

// use(): Promise/Context直接読み取り
const data = use(fetchPromise);
const theme = use(ThemeContext);

// useFormStatus: フォーム送信状態
const { pending } = useFormStatus();
```

### アーキテクチャルール

```
1. Feature間の直接importは禁止 → app層で合成
2. 単方向コードフロー: shared → features → app
3. `export *` 禁止 → 明示的named exportは可
4. 絶対パスimport: @/ エイリアス使用
```

### Feature構成パターン

```
features/[feature]/
├── api/                # API呼び出し + React Query hooks
│   ├── get-xxx.ts      # useQuery hooks
│   ├── create-xxx.ts   # useMutation hooks
│   ├── types.ts        # APIリクエスト/レスポンス型
│   ├── transforms.ts   # Backend ↔ Frontend 変換
│   └── index.ts        # 明示的named export
├── components/         # Feature固有UI
├── hooks/              # ビジネスロジック・UI状態
├── routes/             # ページコンポーネント
├── types/              # ドメイン型定義
└── index.ts            # Public API（外部公開のみ）
```

**詳細は [Frontend規約](../frontend/CLAUDE.md) を参照**

### 禁止事項

| 禁止 | 理由 |
|------|------|
| `any` 型 | 型安全性の破壊 |
| `FC` / `React.FC` | React 19では不要 |
| `forwardRef` | React 19ではref as prop |
| feature間import | アーキテクチャ違反 |
| `export *` | tree-shaking阻害 |
| `console.log` 放置 | 本番コード汚染 |

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
| ファイル | kebab-case | `patient-card.tsx` |
| 型・Interface | PascalCase | `Patient`, `ApiResponse` |

---

## 🤖 エージェント構成

| エージェント | モデル | 用途 |
|-------------|--------|------|
| `architect` | Opus | アーキテクチャ設計、セキュリティ監査 |
| `implementer` | Sonnet | 機能実装、テスト作成 |
| `reviewer` | Sonnet | コードレビュー、品質チェック |
| `debugger` | Sonnet | バグ調査、エラー分析 |
| `researcher` | Haiku | コード検索、ファイル探索 |
| `formatter` | Haiku | コミット生成、コード整形 |

---

## 📊 データベース

| 項目 | 値 |
|------|-----|
| Host | `localhost` (外部) / `db` (Docker内部) |
| Port | `5432` |
| Database | `ekarte_db` |
| User | `ekarte_user` |

マイグレーションは `backend/migrations/` に配置。Docker起動時に自動実行。

---

## 📚 参照

| ドキュメント | 説明 |
|-------------|------|
| [コーディング規約](../CODING_RULES.md) | 全体ルール |
| [Backend Claude設定](../backend/CLAUDE.md) | Backend実装パターン・禁止事項 |
| [Frontend Claude設定](../frontend/CLAUDE.md) | Frontend実装パターン・禁止事項 |
| [アーキテクチャ設計](../docs/architecture.md) | Backendアーキテクチャ詳細 |
| [データフロー](../docs/data-flow.md) | リクエスト〜レスポンスのデータフロー |
| [ERD](../docs/ERD.md) | データベース設計 |
| [仕様定義書](../spec.md) | システム仕様 |
