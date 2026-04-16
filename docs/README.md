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
- **[screens/99-medical-record-flow.md](screens/99-medical-record-flow.md)**: カルテ登録・編集の複雑なライフサイクル仕様。

---

## 💬 LINE予約システム

- **[line/reservation-spec.md](line/reservation-spec.md)**: LINE予約システムの機能要件と UI 仕様。
- **[line/architecture.md](line/architecture.md)**: LINE連携のアーキテクチャ全体像。
- **[line/setup.md](line/setup.md)**: LINE公式アカウントおよび Messaging API のセットアップ手順。

---

## 🧪 テスト・検証

- **[testing/HANDLER_TEST_DOCUMENTATION_STATUS.md](testing/HANDLER_TEST_DOCUMENTATION_STATUS.md)**: API ハンドラーのテストカバレッジ・ステータス。
- **[testing/SECTION_14_MANUAL_TEST_GUIDE.md](testing/SECTION_14_MANUAL_TEST_GUIDE.md)**: 手動ブラウザテストの実行ガイド。
- **[FUNCTIONAL_TEST_REPORT.md](FUNCTIONAL_TEST_REPORT.md)**: 機能テスト・検証レポートの集約版。
- **[archive/](archive/)**: 過去の機能テストレポート、トリアージ記録、および古い設計資料。

---

## 📦 運用・インフラ

- **[OPERATIONS.md](OPERATIONS.md)**: 開発環境のセットアップ・運用コマンド（Make）。
- **[infra/architecture.md](infra/architecture.md)**: AWS ステージング環境のインフラ構成図と構築手順。

---

## 🗄️ タスク管理

- **[tasks/closed/](tasks/closed/)**: 完了済みバグ修正・機能開発タスク。
- **[tasks/open/](tasks/open/)**: 未着手タスク。

---

**最終更新**: 2026-04-15
