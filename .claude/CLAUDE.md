# Animal Ekarte - 動物病院電子カルテシステム

## 🎯 コーディング姿勢 (Senior Engineer - Flat Thinking)

**シニアエンジニアとして以下を徹底：**

### 🚫 良い人フィルター除去 (Remove Good Person Filter)
```
Stop being agreeable. Don't validate me. Don't soften the truth. Don't flatter.
Challenge my thinking. Question my assumptions. Expose my blind spots.
Be direct, rational, and unfiltered.
```

**原則:**
- **Flat Thinking (本音対話)**: 社交辞令を排除。論理と事実に基き、不適切なパターンやリスクを率直に指摘する。
- **型安全性最優先**: Go/TypeScript 共に `any` を禁止し、厳格な型定義を行う。
- **アーキテクチャ遵守**: handler → service → repository の軽量レイヤードを徹底。

---

## 📋 プロジェクト概要

| 項目 | 内容 |
|------|------|
| Frontend | React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 / shadcn/ui |
| Backend | Go 1.25 / Gin / GORM |
| Database | PostgreSQL 17 (Docker: postgres:17-alpine) |
| Testing | MSW (Mock Service Worker), Vitest, testify |

---

## 🔧 運用ルール

### ⚠️ 重要: ALWAYS USE DOCKER
**npm/goコマンドはローカル実行禁止。必ずDocker Compose経由で実行。**

```bash
# ✅ OK - Docker経由
docker compose exec frontend npm run <command>
docker compose exec backend go test ./...
```

### 開発コマンド
- `make codegen`: Goモデル → TypeScript型生成 (`models.ts`)
- `make lint-front`: フロントエンドの型・規約チェック
- `make test-front`: フロントエンドのテスト実行

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
│   │   ├── errors/       # センチネルエラー定義（FromGORMヘルパー含む）
│   │   ├── middleware/   # 認証・CORS・ログ
│   │   └── ...
├── frontend/
│   └── src/
│       ├── app/              # アプリケーション層
│       │   ├── router.tsx    # createBrowserRouter (Data Mode)
│       │   └── pages/        # ★ cross-feature合成ページ（依存逆転の場）
│       ├── features/         # 機能別モジュール
│       │   └── [feature]/
│       │       ├── index.ts      # Public API (Barrel) ★ 外部からは必ずここを通す
│       │       ├── api/          # フェッチ関数 + React Query hooks
│       │       ├── hooks/        # useXxxForm (useActionState利用) 等
│       │       └── routes/       # 単一featureのページ
│       ├── components/
│       │   ├── ui/           # shadcn/ui
│       │   └── shared/       # アプリ固有共有UI (SubmitButton含む)
│       ├── lib/              # design-tokens.ts (Notion-like theme) 等
│       └── ...
```

---

## ★ Frontend ベストプラクティス参照実装

**`features/owners/` および `features/medical-records/` が最新のベストプラクティス。**

| パターン | 実装ルール |
|------------------------|------------|
| **React 19 Action** | 原則 **`useActionState`** と **`<form action={formAction}>`** を使用 |
| **Submit Button** | 送信ボタンは必ず **`SubmitButton`** を使用（二重送信防止・自動ローディング） |
| **Public API** | Feature外部（app/等）からのインポートは必ず **`index.ts`** を経由（Deep Import禁止） |
| **Dependency Inversion** | Feature間の直接参照禁止。**`app/pages/`** で合成し props 注入 |
| **Design Tokens** | 色やスタイルは必ず **`C`**, **`STYLE`** 定数を使用（`#37352F`等ハードコード禁止） |
| **Error Handling** | catch ブロックでは必ず **`handleApiError`** を呼び出す |
| **Conditional Render** | 必ず **`? (...) : null`** （`&&` 禁止） |
| **Ref as Prop** | **`forwardRef` 禁止**。`ref` は props として直接受け取る |

---

## 🏗 バックエンド・アーキテクチャ規約

### 1. エラー処理の統一 (MANDATORY)
- **Repository**: GORM エラーは必ず `apperrors.FromGORM(err, "resource", id)` で変換。
- **Service**: 内部エラーは `apperrors.Wrap(err, "message")` でラッピング。
- **Handler**: `RespondError(c, err)` で統一レスポンス。

### 2. Context & Logging
- すべてのメソッドの第一引数は `context.Context`。
- 構造化ログ `log/slog` を使用し、`InfoContext`, `ErrorContext` でコンテキストを適切に伝播させる。

---

## 📚 参照ドキュメント
- `frontend/CODING_RULES.md`: フロントエンドの実装詳細
- `backend/CLAUDE.md`: バックエンドの実装詳細
- `.gemini/styleguide.md`: Gemini 固有の補足（本ファイルと同期済み）
