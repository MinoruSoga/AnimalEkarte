# Animal Ekarte プロジェクト・ドキュメント

本ディレクトリには、Animal Ekarte の設計、仕様、および開発プロセスに関するドキュメントが格納されています。

## ⚠️ 開発ガイドライン (最優先)

プロジェクトの最新のコーディング規約、アーキテクチャ、および開発コマンドについては、以下のファイルを**必ず**参照してください。

- **[.claude/CLAUDE.md](../.claude/CLAUDE.md)**: **【Single Source of Truth】** 全エージェントおよび開発者向けの統合ルール。
- **[GEMINI.md](../GEMINI.md)**: Gemini CLI 向けの最適化されたコンテキスト。
- **[.gemini/styleguide.md](../.gemini/styleguide.md)**: スタイルガイド・実装パターンの詳細。

---

## 🏗 システム設計

| ドキュメント | 内容 |
|:---|:---|
| [architecture.md](architecture.md) | 全体アーキテクチャ・データフロー（更新継続中） |
| [ERD.md](ERD.md) | データベース設計（ER図） |
| [AUTH.md](AUTH.md) | 認証・認可の仕組み（RBAC） |
| [data-flow.md](data-flow.md) | システム間のデータフロー図 |

---

## 📱 画面・機能仕様

- **[SPECIFICATION.md](SPECIFICATION.md)**: システム全体の機能要件。

---

## 📦 運用・インフラ

- **[OPERATIONS.md](OPERATIONS.md)**: 開発環境のセットアップ・運用手順。
- **[infra/](infra/)**: インフラ構成 (Terraform) とデプロイ手順。

---

## 🗄️ タスク管理

- **[tasks/closed/](tasks/closed/)**: 完了済みバグ修正・機能開発タスク。
- **[tasks/open/](tasks/open/)**: 未着手タスク。
