# Animal Ekarte (アニマル・カルテ)

> **Animal Ekarte**: 最新の React 19 と Go による、高機能かつ保守性の高い動物病院向け電子カルテ管理システム。

---

## 🎯 プロジェクト Town 規約 (MUST READ)

本プロジェクトは、React 19 への完全移行およびバックエンドエラー処理の標準化を完了しています。開発および運用の際は、以下のドキュメントを最優先で参照してください。

- **[.claude/CLAUDE.md](.claude/CLAUDE.md)**: **【Single Source of Truth】** 開発規約・アーキテクチャ・最新ルールの集約地点。
- **[GEMINI.md](GEMINI.md)**: Gemini CLI 向けの最適化されたコンテキスト。
- **[docs/README.md](docs/README.md)**: 各種詳細仕様書（カルテ、Lステップ、会計等）へのポータル。

---

## 🛠 技術スタック

| レイヤー | 技術 |
|:---|:---|
| **Frontend** | React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 / shadcn/ui |
| **Backend** | Go 1.25 / Gin / GORM / Air (Hot Reload) |
| **Database** | PostgreSQL 18 (Docker: postgres:18-alpine) |
| **Infrastructure** | Docker Compose / AWS (ECS Fargate, RDS, S3) / Vercel |
| **Testing** | MSW (Mock Service Worker), Vitest, testify |

---

## 🔧 クイックスタート

### 1. 準備
```bash
# 環境変数のコピー
cp .env.example .env
```

### 2. 起動
```bash
# 全コンテナの起動（初期化スクリプト自動実行）
make up

# フロントエンド型定義の同期
make codegen
```

| サービス | ローカル URL |
|:---|:---|
| **Frontend** | [http://localhost:3000](http://localhost:3000) |
| **Backend API** | [http://localhost:8080/api/v1](http://localhost:8080/api/v1) |
| **Database** | `localhost:5434` (User: admin / Pass: password) |

---

## 📖 ドキュメント体系 (92 Tables / 31 Resources)

| カテゴリ | 主要ドキュメント |
|:---|:---|
| **業務仕様** | [SPECIFICATION.md](docs/SPECIFICATION.md) / [screens/](docs/screens/) |
| **機能詳細** | [Lステップ連携](docs/line/lstep-integration.md) / [会計・集計](docs/CASH_REGISTER_SPEC.md) / [顧客分析](docs/CUSTOMER_AGGREGATION_SPEC.md) |
| **技術設計** | [Architecture](docs/architecture.md) / [ER図 (v31.17)](docs/ERD.md) / [認証・認可](docs/AUTH.md) |
| **API** | [API_SPEC.md (v2.3)](docs/API_SPEC.md) / [openapi.yaml](docs/openapi.yaml) |
| **運用・テスト** | [Deployment Hub](docs/infra/deploy/README.md) / [Manual Test Guide](docs/testing/SECTION_14_MANUAL_TEST_GUIDE.md) |

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
