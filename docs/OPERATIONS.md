# 運用・開発ガイド (OPERATIONS)

本ドキュメントは、システムの開発、ビルド、運用、およびデプロイの手順を網羅します。

---

## 1. 開発環境のセットアップ

本プロジェクトは Docker Compose をベースとしており、ローカルマシンへのランタイム（Go/Node.js）の直接インストールは不要です。

### 基本コマンド
- **起動**: `make up`
- **停止**: `make down`
- **再構築**: `make build`
- **全ログ確認**: `make logs`

---

## 2. 開発ワークフロー

### 2.1 型安全な開発 (Codegen)
バックエンドの Go モデル (`internal/model/*.go`) を **Single Source of Truth** とし、フロントエンドの TypeScript 型を自動生成します。

- **実行**: `make codegen`
- **生成先**: `frontend/src/types/generated/`
- **仕組み**: `tygo` (Go compiled to TS) を使用。モデル変更後は必ず実行してください。

### 2.2 データベース・マイグレーション
`backend/migrations/` 配下の SQL ファイルが自動的に適用されます。

---

## 3. 運用コマンド

### ログの監視
- **APIログ**: `make logs-api`
- **フロントエンドログ**: `make logs-front`

### メンテナンス
- **DB直接操作**: `make db` (psqlを起動)
- **完全リセット**: `make reset` (ボリュームを破棄して初期状態に戻す)

---

## 4. デプロイメント

デプロイ手順は [CI-CD-PIPELINE.md](./infra/deploy/CI-CD-PIPELINE.md) を参照。
インフラ構成は [docs_infra_architecture.md](./infra/docs_infra_architecture.md) を参照。

---

## 5. セキュリティ運用

- **秘密情報の管理**: ローカルでは `.env`、AWS では **SSM Parameter Store** (`/animalekarte/test/db/*`) を使用します。
- **環境変数**: `.env.example` をコピーして必要な値を設定してください。機密情報は `git` にコミットしないでください。
