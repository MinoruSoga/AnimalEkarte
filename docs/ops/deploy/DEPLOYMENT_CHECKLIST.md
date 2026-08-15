# デプロイメント・リリース準備チェックリスト (Deployment Checklist)

> **目的**: STG/本番デプロイ前の統合チェックリストを提供する。
> **読者**: リリース担当者。
> **タイミング**: デプロイ実行前。

> **Animal Ekarte**: ステージングおよび本番環境への安全な移行手順
> **最新更新**: 2026-07-23 | **ステータス**: STG 稼働中 / Production 未構築

---

## 1. 事前検証（開発環境・CI）

デプロイを開始する前に、ローカルおよび CI 環境で以下の品質基準を全て満たしていることを確認します。

- [ ] **コード品質の完遂**
  - [ ] `make lint`: バックエンド静的解析（golangci-lint）で指摘ゼロ。
  - [ ] `make lint-front`: フロントエンド ESlint/Prettier 検証完了。
  - [ ] `docker compose exec frontend pnpm type-check`: TypeScript の型整合性（0 errors）。
- [ ] **自動テストの PASS**
  - [ ] `make test`: バックエンド単体・結合テスト（PASS 100%）。
  - [ ] `make test-front`: フロントエンド Vitest（PASS 100%）。
  - [ ] `docker compose exec frontend pnpm test:e2e`: Playwright による、予約〜診察〜会計の完走。

---

## 2. インフラ・構成準備（Cloudflare / Vercel）

インフラ層の構成変更、および機密情報の同期状態を確認します。現行構成は
[`../infra/architecture.md`](../infra/architecture.md) を正本とし、バックエンドは Cloudflare Workers + Containers、
DB は PlanetScale Postgres を使用します。AWS ECS/RDS は廃止済みで、切り戻し先はありません。

- [ ] **環境変数（Secret）の完全同期**
  - [ ] `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` / `DB_SSL_MODE`: 対象環境の PlanetScale Postgres を指しているか（`wrangler secret put` で投入）。
  - [ ] `JWT_SECRET`: 本番用（32文字以上のランダム文字列）の設定。
  - [ ] `INTEGRATION_ENCRYPTION_KEY`: 病院別 API キー保護用の AES-256 キー。
  - [ ] `S3_SHARED_BUCKET`: 共有ファイル（LINE連携用、実体は Cloudflare R2）のバケット準備。
- [ ] **DB マイグレーションの整合性**
  - [ ] 全 115 テーブルのスキーマが、ローカルのマイグレーション（`001_init.sql`。旧増分`002`〜`009`および2026-07-27夕統合の旧`002`〜`004`の原文・元コミット・SHA-256は末尾のアーカイブ節へ統合済み、詳細は`docs/architecture/erd.md` §4.3）と完全一致しているか。
  - [ ] 初期マスタデータ（`backend/migrations/seeds/002_master/`）の投入準備完了。

---

## 3. リリリース実行手順

1.  **ブランチ管理**: 開発完了機能を `staging` または `production` ブランチへマージ（バックエンドは `staging` push で `backend-deploy.yml` が自動起動。本番向けバックエンド自動デプロイワークフローは未整備）。
2.  **デプロイ監視**: 
    - GitHub Actions `backend-deploy.yml`（Cloudflare Workers + Containers デプロイ）の進捗監視。
    - Vercel ダッシュボードにて、フロントエンドのビルド成功とエッジ配信を確認。
3.  **DB更新**: `backend-deploy.yml` のマイグレーションステップ（`infra/scripts/cf-run-migrate.sh`）に任せる。現行 workflow に `db_reset` 入力はない。共有 STG の再作成が必要な場合は、破壊的操作として別途明示承認を得て [PlanetScale STG シード投入 Runbook](./STG_PLANETSCALE_SEED_RUNBOOK.md) に従う。

---

## 4. リリース後検証 (Post-Deploy Smoke Test)

デプロイ直後に、実際の環境で以下の「クリティカル・パス」を確認します。

- [ ] **システム疎通**: `/health` エンドポイントが `200 OK` を返すか。
- [ ] **認証基盤**: 管理者および一般スタッフで正常にログインし、セッションが維持されるか。
- [ ] **マルチテナント隔離**: 他院のデータが混入していないことを実機で再確認。
- [ ] **外部連携テスト**: Lステップ設定画面から「疎通テスト」を実行し、成功（Green）するか。

---

## 5. 障害時の復旧判断

- **切り戻し先なし**: AWS は廃止済みで、ホットスタンバイや AWS への自動切り戻しは存在しない。
- **初動**: [`../infra/staging/runbook.md`](../infra/staging/runbook.md) に従い、実 URL / workers.dev の `/health`、ローリング更新の旧イメージ残留、PlanetScale 接続枯渇、Cloudflare 障害情報を切り分ける。
- **復旧**: Cloudflare 側の修正・再デプロイを基本とし、基盤喪失時はスナップショットと現行 IaC から再建する。Workers Logs / Containers の観測結果と GitHub Actions の失敗ステップを記録する。

---
