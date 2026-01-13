# ドキュメント目次

Animal Ekarte（動物病院電子カルテシステム）の技術ドキュメントです。

## ドキュメント一覧

| ドキュメント | 説明 |
|-------------|------|
| [ERD.md](./ERD.md) | データベース設計書（ER図、テーブル定義、ステータス定義） |
| [MIGRATION.md](./MIGRATION.md) | マイグレーションシステム（DB作成・管理方法） |
| [API-ROADMAP.md](./API-ROADMAP.md) | 未実装API設計書（将来実装予定のエンドポイント） |

## API仕様

### 実装済みAPI（Swagger）

実装済みAPIの仕様は Swagger UI で確認できます：

- **Swagger UI:** http://localhost:8080/swagger/index.html
- **OpenAPI仕様:** `backend/docs/swagger.yaml`

### 未実装API（設計のみ）

- [API-ROADMAP.md](./API-ROADMAP.md) - 将来実装予定のエンドポイント一覧

## クイックリンク

### データベース設計

- [ER図（Mermaid）](./ERD.md#er図-mermaid)
- [テーブル一覧](./ERD.md#テーブル一覧)
- [ステータス定義](./ERD.md#ステータス定義)
- [インデックス設計](./ERD.md#インデックス設計)

## 関連ドキュメント

- [プロジェクト概要](../CLAUDE.md)
- [仕様定義書](../spec.md)
- [デザインシステム](../DESIGN_SYSTEM.md)
- [Backend README](../backend/README.md)
