# Animal Ekarte

動物病院向け電子カルテシステム

## ⚠️ 重要: Docker環境

**このプロジェクトはDocker環境で動作する。npm/goコマンドのローカル実行は禁止。**

```bash
# ✅ 正しい実行方法
docker compose exec frontend npm run <command>
docker compose exec backend go <command>
```

## クイックリファレンス

| タスク | コマンド |
|--------|---------|
| 起動 | `make up` |
| 停止 | `make down` |
| ログ | `make logs` |
| DB接続 | `make db` |
| Frontend ビルド | `docker compose exec frontend npm run build` |
| Frontend Lint | `docker compose exec frontend npm run lint` |
| Backend テスト | `docker compose exec backend go test ./... -v` |
| Backend Lint | `docker compose exec backend golangci-lint run ./...` |

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| Frontend | React 19, TypeScript 5.7, Vite 6, Tailwind CSS 4, shadcn/ui |
| Backend | Go 1.22+, Gin, GORM |
| Database | PostgreSQL 18 |
| Infrastructure | Docker Compose |

## アーキテクチャ

### Frontend (bulletproof-react準拠)
```
src/
├── app/           # ルーター、プロバイダ
├── components/    # ui/, shared/, layouts/, errors/
├── features/      # 機能別モジュール
├── hooks/         # 共有hooks
├── lib/           # ユーティリティ
├── types/         # 共有型定義
└── testing/       # テスト設定
```

### Backend (クリーンアーキテクチャ)
```
internal/
├── handler/       # HTTPハンドラ
├── service/       # ビジネスロジック
├── repository/    # データアクセス
├── model/         # ドメインモデル
├── errors/        # センチネルエラー
└── config/        # 設定
```

## 核心ルール

### Frontend
- **React 19**: `FC`禁止、`forwardRef`禁止、ref as prop
- **bulletproof-react**: feature間import禁止、barrel file禁止
- **TypeScript**: `any`禁止、型安全性優先

### Backend
- **Context伝播**: 全関数の第一引数に`context.Context`
- **エラーハンドリング**: センチネルエラー + ラッピング
- **ログ**: slog構造化ログ

## ドキュメント

| ドキュメント | 説明 |
|-------------|------|
| [プロジェクト詳細](.claude/CLAUDE.md) | 包括的なプロジェクト設定 |
| [コーディング規約](CODING_RULES.md) | 全体のコーディングルール |
| [Frontend規約](frontend/CODING_RULES.md) | React 19 / TypeScript詳細 |
| [Backend規約](backend/CODING_RULES.md) | Go / Gin / GORM詳細 |
| [ERD](docs/ERD.md) | データベース設計 |
| [API仕様](http://localhost:8080/swagger/index.html) | Swagger UI |
| [仕様定義書](spec.md) | システム仕様 |
