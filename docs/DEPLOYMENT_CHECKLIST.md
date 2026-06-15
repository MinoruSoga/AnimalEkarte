# デプロイメント・リリース準備チェックリスト (Deployment Checklist)

> **Animal Ekarte**: ステージングおよび本番環境への安全な移行手順
> **最新更新**: 2026-06-12 | **ステータス**: Production Ready (103 Tables Sync)

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

## 2. インフラ・構成準備（AWS / Vercel）

インフラ層の構成変更、および機密情報の同期状態を確認します。

- [ ] **環境変数（Secret）の完全同期**
  - [ ] `DATABASE_URL`: 正しい本番/STG RDS インスタンスを指しているか。
  - [ ] `JWT_SECRET`: 本番用（32文字以上のランダム文字列）の設定。
  - [ ] `INTEGRATION_ENCRYPTION_KEY`: 病院別 API キー保護用の AES-256 キー。
  - [ ] `S3_SHARED_BUCKET`: 共有ファイル（LINE連携用）のバケット準備。
- [ ] **DB マイグレーションの整合性**
  - [ ] 全 103 テーブルのスキーマが、ローカルの `001_init.sql` と完全一致しているか。
  - [ ] 初期マスタデータ（`002_seed_master.sql`）の投入準備完了。

---

## 3. リリリース実行手順

1.  **ブランチ管理**: 開発完了機能を `staging` または `production` ブランチへマージ。
2.  **デプロイ監視**: 
    - GitHub Actions `backend-deploy.yml` の進捗監視。
    - Vercel ダッシュボードにて、フロントエンドのビルド成功とエッジ配信を確認。
3.  **DB更新**: 大規模変更時は `db_reset=true` を指定（初回のみ）、通常は自動マイグレーションに任せる。

---

## 4. リリース後検証 (Post-Deploy Smoke Test)

デプロイ直後に、実際の環境で以下の「クリティカル・パス」を確認します。

- [ ] **システム疎通**: `/health` エンドポイントが `200 OK` を返すか。
- [ ] **認証基盤**: 管理者および一般スタッフで正常にログインし、セッションが維持されるか。
- [ ] **マルチテナント隔離**: 他院のデータが混入していないことを実機で再確認。
- [ ] **外部連携テスト**: Lステップ設定画面から「疎通テスト」を実行し、成功（Green）するか。

---

## 5. 緊急時の切り戻し（Rollback）

- **即時ロールバック**: 致命的な不具合発見時、AWS ECS タスク定義を前のリビジョンに即座に戻す準備があるか。
- **ログ分析**: `aws logs tail` により、起動エラーやランタイムエラーの有無を 10 分間監視。

---
