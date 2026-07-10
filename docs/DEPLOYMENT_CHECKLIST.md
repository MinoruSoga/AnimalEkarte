# デプロイメント・リリース準備チェックリスト (Deployment Checklist)

> **目的**: STG/本番デプロイ前の統合チェックリストを提供する。
> **読者**: リリース担当者。
> **タイミング**: デプロイ実行前。

> **Animal Ekarte**: ステージングおよび本番環境への安全な移行手順
> **最新更新**: 2026-07-10 | **ステータス**: Production Ready (108 Tables Sync)

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

インフラ層の構成変更、および機密情報の同期状態を確認します（バックエンドは Cloudflare Workers + Containers。AWS ECS/RDS は Phase 8 完了までのロールバック専用経路）。

- [ ] **環境変数（Secret）の完全同期**
  - [ ] `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` / `DB_SSL_MODE`: 正しい本番/STG PlanetScale Postgres インスタンスを指しているか（`wrangler secret put` で投入。ECS ロールバック経路のみ SSM Parameter Store）。
  - [ ] `JWT_SECRET`: 本番用（32文字以上のランダム文字列）の設定。
  - [ ] `INTEGRATION_ENCRYPTION_KEY`: 病院別 API キー保護用の AES-256 キー。
  - [ ] `S3_SHARED_BUCKET`: 共有ファイル（LINE連携用、実体は Cloudflare R2）のバケット準備。
- [ ] **DB マイグレーションの整合性**
  - [ ] 全 108 テーブルのスキーマが、ローカルのマイグレーション（`001_init.sql`。旧増分`005`〜`012`は統合済み、詳細は`docs/ERD.md` §4.3）と完全一致しているか。
  - [ ] 初期マスタデータ（`002_seed_master.sql`）の投入準備完了。

---

## 3. リリリース実行手順

1.  **ブランチ管理**: 開発完了機能を `staging` または `production` ブランチへマージ（バックエンドは `staging` push で `backend-deploy.yml` が自動起動。本番向けバックエンド自動デプロイワークフローは未整備）。
2.  **デプロイ監視**: 
    - GitHub Actions `backend-deploy.yml`（Cloudflare Workers + Containers デプロイ）の進捗監視。
    - Vercel ダッシュボードにて、フロントエンドのビルド成功とエッジ配信を確認。
3.  **DB更新**: 通常は `backend-deploy.yml` のマイグレーションステップ（`infra/scripts/cf-run-migrate.sh`）に任せる。`db_reset=true` オプションは AWS ECS ロールバック経路（`backend-deploy-ecs.yml` の `workflow_dispatch` 入力）のみで利用可能で、Cloudflare 経路には無い（`001_init.sql` 変更時など DB 再作成が必要な場合は ECS 経路を使う）。

---

## 4. リリース後検証 (Post-Deploy Smoke Test)

デプロイ直後に、実際の環境で以下の「クリティカル・パス」を確認します。

- [ ] **システム疎通**: `/health` エンドポイントが `200 OK` を返すか。
- [ ] **認証基盤**: 管理者および一般スタッフで正常にログインし、セッションが維持されるか。
- [ ] **マルチテナント隔離**: 他院のデータが混入していないことを実機で再確認。
- [ ] **外部連携テスト**: Lステップ設定画面から「疎通テスト」を実行し、成功（Green）するか。

---

## 5. 緊急時の切り戻し（Rollback）

- **即時ロールバック**: 致命的な不具合発見時、`backend-deploy-ecs.yml`（`workflow_dispatch` 専用の AWS ECS ロールバック経路）でタスク定義を前のリビジョンに即座に戻す準備があるか。
- **ログ分析**: Cloudflare 経路は Workers Logs、AWS ECS ロールバック経路は `aws logs tail` により、起動エラーやランタイムエラーの有無を 10 分間監視。

---
