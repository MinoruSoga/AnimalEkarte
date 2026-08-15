# Animal Ekarte (アニマル・カルテ)

> **Animal Ekarte**: 最新の React 19 と Go による、高機能かつ保守性の高い動物病院向け電子カルテ管理システム。

---

## 🎯 プロジェクト規約 (MUST READ)

本プロジェクトは、React 19 への完全移行およびバックエンドエラー処理の標準化を完了しています。開発および運用の際は、以下のドキュメントを最優先で参照してください。

- **[.claude/CLAUDE.md](.claude/CLAUDE.md)**: **【Single Source of Truth】** 開発規約・アーキテクチャ・最新ルールの集約地点。
- **[docs/README.md](docs/README.md)**: 各種詳細仕様書（カルテ、Lステップ、会計等）へのポータル。

---

## 🛠 技術スタック

| レイヤー | 技術 |
|:---|:---|
| **Frontend** | React 19.2 / TypeScript 6.0 / Vite 8.0 / Tailwind CSS 4 / shadcn/ui |
| **Backend** | Go 1.25.0 / Gin / GORM / Air (Hot Reload) |
| **Database** | PostgreSQL 18 (Docker: `postgres:18-alpine`) |
| **Infrastructure** | Docker Compose / AWS (ECS Fargate, RDS, S3) / Vercel |
| **Testing** | MSW (Mock Service Worker), Vitest, testify |

---

## 🔧 クイックスタート

### 1. 環境変数の準備

```bash
# ローカル開発用のテンプレートをコピー（別マシンでも同じ手順）
cp .env.example .env.local
```

`.env.example` にはローカル Docker 用のキーとダミー値が入っています。**コピーしただけで `make up` できます。**  
開発用 Make / Compose は `.env.local` を変数源にします（`.env` は使わない）。`.env.local` は Git 管理外です。STG/本番の認証情報・LINE/Vercel トークンは `.env.example` に書かず、必要な人だけ `.env.local` に追記してください。

### 2. 起動

```bash
# db / backend / frontend を起動し、ヘルスチェック完了まで待機
# DB マイグレーションは backend の起動時に自動適用
make up
```

### 3. 型定義の同期（Go モデル変更時）

```bash
# Go モデルからフロントエンド型定義を同期
make codegen
```

| サービス | ローカル URL |
|:---|:---|
| **Frontend** | [http://localhost:3003](http://localhost:3003) |
| **Backend API base** | `http://localhost:8080/api/v1` |
| **Backend health check** | [http://localhost:8080/health](http://localhost:8080/health) |
| **Database** | `localhost:5434`（接続情報は `.env.local` の `DB_NAME` / `DB_USER` / `DB_PASSWORD`） |

---

## 📖 ドキュメント体系 (詳細は [docs/README.md](docs/README.md) 参照)

| カテゴリ | 主要ドキュメント |
|:---|:---|
| **作業台帳（進行中）** | **Linear**（hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4)）· 入口 [todo.md](todo.md) / [todo-po.md](todo-po.md)（ポインタのみ）· マップ [todo-docs-linear-map](reports/todo-walk-2026-08-14/todo-docs-linear-map.md) · [docs/work/](docs/work/README.md) |
| **業務仕様** | [SPECIFICATION.md](docs/spec/specification.md) / [screens/](docs/spec/screens/) |
| **機能詳細** | [Lステップ連携](docs/spec/line/lstep-integration.md) / [会計・集計](docs/spec/cash-register.md) / [顧客分析](docs/spec/customer-aggregation.md) |
| **技術設計** | [Architecture](docs/architecture/overview.md) / [ER図](docs/architecture/erd.md)（テーブル数の正本） / [認証・認可](docs/architecture/auth.md)（RBACリソース数の正本） |
| **API** | [backend/docs/api.yaml](backend/docs/api.yaml)（contract 正本。Swagger UI 表示は `docker compose -f docker-compose.swagger.yml up`） |
| **運用・テスト** | [Deployment Hub](docs/ops/deploy/README.md) / [Manual Test Guide](docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md) |
---

## 🌐 インタラクティブ API ドキュメント

Swagger UI および Redoc をローカルで起動して、対話的な API 検証が可能です。

```bash
# ドキュメントツールの起動
docker compose -f docker-compose.swagger.yml up
```
- **Swagger UI**: [http://localhost:8081](http://localhost:8081)
- **Redoc**: [http://localhost:8082](http://localhost:8082)

---

## ライセンス
Private / Proprietary
