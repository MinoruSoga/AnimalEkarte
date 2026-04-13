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
| [architecture.md](architecture.md) | バックエンドアーキテクチャ（レイヤード構造・エラーハンドリング） |
| [ERD.md](ERD.md) | データベース設計（テーブル定義・リレーション） |
| [AUTH.md](AUTH.md) | 認証・認可の仕組み（RBAC・JWT・マルチクリニック） |
| [data-flow.md](data-flow.md) | リクエスト〜レスポンスのデータフロー |
| [DESIGN_SYSTEM.md](DESIGN_SYSTEM.md) | UIデザイン規約・デザイントークン |

---

## 📱 画面・機能仕様

- **[SPECIFICATION.md](SPECIFICATION.md)**: システム全体の機能要件と主要ビジネスフロー。
- **[screens/](screens/)**: 各画面の詳細仕様（項目定義・コンポーネント構成）。
- **[spec-medical-record-flow.md](spec-medical-record-flow.md)**: カルテ登録・編集の複雑なライフサイクル仕様。

---

## 💬 LINE予約システム

- **[line-reseavation.md](line-reseavation.md)**: LINE予約システムの機能要件と UI 仕様。
- **[LINE-RESERVATION-ARCHITECTURE.md](LINE-RESERVATION-ARCHITECTURE.md)**: LINE連携のアーキテクチャ全体像。
- **[LINE-SETUP.md](LINE-SETUP.md)**: LINE公式アカウントおよび Messaging API のセットアップ手順。

---

## 📦 運用・インフラ

- **[OPERATIONS.md](OPERATIONS.md)**: 開発環境のセットアップ・運用コマンド（Make）。
- **[infra/docs_infra_architecture.md](infra/docs_infra_architecture.md)**: AWS ステージング環境のインフラ構成図と構築手順。

---

## 🗄️ タスク管理

- **[tasks/closed/](tasks/closed/)**: 完了済みバグ修正・機能開発タスク。
- **[tasks/open/](tasks/open/)**: 未着手タスク。

---

**最終更新**: 2026-04-13
