# Animal Ekarte

動物病院向け電子カルテ管理システム

## 技術スタック

| レイヤー | 技術 |
|---------|------|
| Frontend | React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 / shadcn/ui |
| Backend | Go / Gin / Air (Hot Reload) |
| Database | PostgreSQL 18 |
| Infrastructure | Docker Compose |

## 前提条件

- Docker / Docker Compose
- Make

## セットアップ

```bash
# リポジトリをクローン
git clone <repository-url>
cd AnimalEkarte

# .env ファイルを作成
cp .env.example .env
# DB_USER, DB_PASSWORD, DB_NAME を設定

# コンテナ起動
make up
```

起動後のアクセス先：

| サービス | URL |
|---------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| PostgreSQL | localhost:5432 |

## 開発コマンド

```bash
make up              # コンテナ起動
make build           # コンテナ起動（ビルド付き）
make down            # コンテナ停止
make logs            # 全ログ表示
make logs-api        # APIログ表示
make logs-front      # フロントエンドログ表示
make ps              # コンテナ状態確認
make db              # DB接続（psql）
make restart-api     # API再起動
make restart-front   # フロントエンド再起動
make clean           # キャッシュクリア＆再ビルド
make reset           # 完全リセット（データ削除）
```

### 品質管理

```bash
make lint            # Go リンター実行
make lint-fix        # Go リンター実行（自動修正）
make test            # Go テスト実行
make test-cover      # Go テスト実行（カバレッジ付き）
make swagger         # Swagger ドキュメント生成
```

> **⚠️ npm / go コマンドはローカルで直接実行しないこと。必ず Docker 経由で実行する。**
>
> ```bash
> # ✅ 正しい実行方法
> docker compose exec frontend npm run build
> docker compose exec backend go test ./...
> ```

## プロジェクト構成

```
AnimalEkarte/
├── backend/
│   ├── cmd/              # エントリーポイント
│   ├── internal/         # 内部パッケージ
│   │   ├── handler/      # HTTP ハンドラ
│   │   ├── service/      # ビジネスロジック
│   │   ├── repository/   # データアクセス
│   │   ├── model/        # ドメインモデル
│   │   ├── errors/       # エラー定義
│   │   ├── config/       # 設定
│   │   └── logger/       # ロガー
│   ├── migrations/       # DB マイグレーション
│   └── docs/             # Swagger ドキュメント
├── frontend/
│   └── src/
│       ├── app/          # アプリケーション層（ルーター、プロバイダ）
│       ├── components/   # 共通コンポーネント（ui/, shared/, layouts/）
│       ├── features/     # 機能別モジュール（bulletproof-react準拠）
│       ├── hooks/        # 共有hooks
│       ├── lib/          # ユーティリティ
│       ├── types/        # 型定義
│       └── testing/      # テスト設定
├── docs/                 # 技術ドキュメント
├── docker-compose.yml
├── Makefile
└── .env
```

## ドキュメント

- [コーディング規約](CODING_RULES.md)
- [Frontend コーディング規約](frontend/CODING_RULES.md)
- [Backend コーディング規約](backend/CODING_RULES.md)
- [ERD（データベース設計）](docs/ERD.md)
- [API 設計ロードマップ](docs/API-ROADMAP.md)
- [仕様定義書](spec.md)

## ライセンス

Private
