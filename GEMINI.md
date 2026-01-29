# Animal Ekarte - 動物病院電子カルテシステム (Gemini Context)

## 🎯 コーディング姿勢 (Gemini Agent Guidelines)

**シニアエンジニアとして以下を徹底：**
- **型安全性最優先**: TypeScriptとGoの型システムを最大限活用する。
- **SOLID原則・クリーンアーキテクチャ**: 依存性の逆転、責任の分離を意識する。
- **エラーハンドリング徹底**: エラーを無視せず、適切にラップして伝播させる。
- **セキュリティ意識**: SQLインジェクション対策、機密情報の扱いに注意する。
- **パフォーマンス考慮**: N+1問題の回避、不要なレンダリングの抑制。
- **自己レビュー実施**: 生成コードの品質を担保する。

---

## 📋 プロジェクト概要

**名前:** Animal Ekarte
**説明:** 動物病院向け電子カルテ管理システム
**日付:** 2026-01-14

---

## 🛠️ 技術スタック

### Backend
- **言語:** Go (Golang)
- **フレームワーク:** Gin
- **ホットリロード:** Air
- **ロギング:** slog (構造化ログ)
- **Linter:** golangci-lint

### Frontend
- **言語:** TypeScript 5.7
- **フレームワーク:** React 19
- **ルーティング:** React Router (Data Mode)
- **ビルドツール:** Vite 6
- **スタイル:** Tailwind CSS 4
- **UIライブラリ:** shadcn/ui (Radix UIベース)
- **状態管理:** TanStack Query, Zustand, React Hooks, Context API
- **アイコン:** lucide-react

### Infrastructure
- **データベース:** PostgreSQL 18
- **コンテナ:** Docker Compose
- **マイグレーション:** SQL files (backend/migrations/)

---

## 📁 ディレクトリ構造

```
AnimalEkarte/
├── backend/              # Go (Gin)
├── frontend/
│   ├── src/
│   │   ├── main.tsx
│   │   ├── vite-env.d.ts
│   │   ├── app/          # App entry, providers, router, ErrorBoundary
│   │   ├── features/     # Feature-based modules (auth, dashboard, reservations, etc.)
│   │   ├── components/   # Shared components (ui/, shared/)
│   │   │   ├── ui/       # shadcn/ui
│   │   │   └── shared/   # App-specific shared UI (Layout, Form, DataTable, Feedback...)
│   │   ├── hooks/        # Global shared hooks
│   │   ├── lib/          # Library config (axios, queryClient, etc.)
│   │   ├── stores/       # Global state (auth, theme, etc.)
│   │   ├── types/        # Shared types
│   │   ├── utils/        # Global utilities (format, validation, etc.)
│   │   └── testing/      # Test setup & MSW (Mock Service Worker)
│   ├── vite.config.ts
│   └── ...
├── docker-compose.yml
├── Makefile
└── .env
```

---

## 🚀 開発コマンド (重要)

**npm/goコマンドはローカルで実行せず、必ずDocker経由で実行してください。**

| タスク | コマンド |
|--------|---------|
| コンテナ起動 | `make up` |
| コンテナ停止 | `make down` |
| 全ログ表示 | `make logs` |
| DB接続 (psql) | `make db` |
| Frontend ビルド | `docker compose exec frontend npm run build` |
| Frontend Lint | `docker compose exec frontend npm run lint` |
| Frontend テスト | `docker compose exec frontend npm run test:run` |
| Backend テスト | `docker compose exec backend go test ./... -v` |
| Backend Lint | `docker compose exec backend golangci-lint run ./...` |

---

## 📝 コーディング規約

詳細なコーディング規約とスタイルガイドは **[.gemini/styleguide.md](.gemini/styleguide.md)** を参照してください。

### 主なポイント
- **Go (Backend):** パッケージ名は小文字、ExportはPascalCase。Context必須。
- **TypeScript (Frontend):** コンポーネントはPascalCase。Feature-based Architecture。
- **共通:** 型安全性最優先。

---

## 🔐 環境変数とセキュリティ
- `.env` ファイルで管理 (`DB_USER`, `DB_PASSWORD`, etc.)。
- シークレットはコードにコミットしない。

---

## 📐 重要な実装パターン

### Go: Repository Pattern
```go
type PatientRepository interface {
    FindByID(ctx context.Context, id string) (*Patient, error)
}
```

### Go: Error Handling
```go
if err != nil {
    return nil, fmt.Errorf("failed to find patient: %w", err)
}
```

### React: Feature Structure
```
features/owners/
  ├── api/         # API calls (+ hooks)
  ├── components/  # Feature-specific UI
  ├── hooks/       # Logic
  ├── routes/      # Route components
  ├── types/       # Types
  └── index.ts     # Public API
```

---

## ✅ チェックリスト (実装時)

1.  **既存コードの確認:** 似た機能の実装方法を真似る。
2.  **型定義:** 最初に型を定義する。
3.  **テスト:** 必要に応じてテストを追加/修正する。
4.  **Lint/Format:** `make lint` 相当のコマンドでチェックする。
5.  **Docker:** 動作確認は必ずDockerコンテナ内で行うか、`make` コマンドを使用する。
