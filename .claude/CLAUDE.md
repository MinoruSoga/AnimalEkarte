# Animal Ekarte - 動物病院電子カルテシステム

## 🎯 コーディング姿勢 (Senior Engineer - Flat Thinking)

- **Flat Thinking**: 社交辞令を排除。論理と事実に基き、不適切なパターンやリスクを率直に指摘する。
- **型安全性最優先**: Go/TypeScript 共に `any` を禁止し、厳格な型定義を行う。
- **アーキテクチャ遵守**: handler → service → repository の軽量レイヤードを徹底。

---

## 📋 技術スタック

| 項目 | 内容 |
|------|------|
| Frontend | React 19, TypeScript 5.7, Vite 6, Tailwind CSS 4, shadcn/ui |
| Backend | Go 1.25, Gin, GORM |
| Database | PostgreSQL 17 (Docker: postgres:17-alpine) |
| Testing | MSW (Mock Service Worker), Vitest, testify |

---

## 🔧 運用ルール

### ⚠️ コマンド実行: ALWAYS USE DOCKER
**npm/goコマンドはローカル実行禁止。必ずDocker Compose経由で実行。**

```bash
docker compose exec frontend npm run <command>
docker compose exec backend go test ./...
```

### 開発ツール
- `make codegen`: Goモデル → TypeScript型生成 (`models.ts`)
- `make lint-front`: フロントエンドの型・規約チェック
- `make test-front`: フロントエンドのテスト実行

---

## 🏗 フロントエンド・アーキテクチャ (React 19)

### 1. Feature-Based + Public API
- `src/features/[feature]/index.ts` を **Public API (Barrel)** として必ず整備する。
- **重要**: feature 外部（`app/` レイヤー等）からは、必ずこの `index.ts` を経由してインポートする。**内部ファイルへの Deep Import は厳禁。**

### 2. Dependency Inversion (Synthesis in app/pages)
- Feature 間の直接インポートは禁止。
- 複数 feature を組み合わせる場合は、`src/app/pages/` で合成し、props 経由で依存を注入する。

### 3. React 19 Action パターン (MANDATORY)
- フォーム送信は原則 **`useActionState`** と **`<form action={formAction}>`** を使用する。
- 送信ボタンは必ず **`SubmitButton`** (shared/Form) を使用し、`useFormStatus` による自動制御を行う。
- `setIsPending` による手動の状態管理は廃止。

### 4. デザインシステム (Notion-like)
- **絶対ルール**: 色やスタイルは必ず `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。`#37352F` 等のハードコードは禁止。

---

## 🏗 バックエンド・アーキテクチャ (Clean Architecture)

### 1. エラー処理の統一
- **Repository**: GORM エラーは必ず `apperrors.FromGORM(err, "resource", id)` で変換する。
- **Service**: 内部エラーは `apperrors.Wrap(err, "message")` でラッピングする。
- **Handler**: `RespondError(c, err)` を使用して一貫したエラーレスポンスを返却する。

### 2. Context & Logging
- すべてのメソッドの第一引数は `context.Context`。
- 構造化ログ `log/slog` を使用し、`InfoContext`, `ErrorContext` でコンテキストを伝播させる。

---

## 📚 詳細規約
- `frontend/CODING_RULES.md`: フロントエンドの詳細ルール
- `backend/CLAUDE.md`: バックエンドの詳細ルール
- `.gemini/styleguide.md`: Gemini 固有の補足（本ファイルと同期済み）
