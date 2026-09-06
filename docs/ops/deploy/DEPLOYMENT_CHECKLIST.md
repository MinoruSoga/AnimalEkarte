# デプロイメント・リリース準備チェックリスト (Deployment Checklist)

> **目的**: STG/本番デプロイ前の統合チェックリストを提供する。
> **読者**: リリース担当者。
> **タイミング**: デプロイ実行前。

> **Animal Ekarte**: ステージングおよび本番環境への安全な移行手順
> **最新更新**: 2026-08-31 | **checked-in config**: STG target/workflowあり、Production workflow未実装。live provider状態はUNKNOWNで実行時receiptが必要

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
  - [ ] `make e2e`（必要なら `make e2e ARGS='e2e/<spec>.spec.ts'`）: support済みPlaywright runnerで予約〜診察〜会計を完走。

---

## 2. インフラ・構成準備（Cloudflare / Vercel）

インフラ層の構成変更、および機密情報の同期状態を確認します。現行構成は
[`../infra/architecture.md`](../infra/architecture.md) を正本とし、バックエンドは Cloudflare Workers + Containers、
DB は PlanetScale Postgres を使用します。AWS ECS/RDS は廃止済みで、切り戻し先はありません。

- [ ] **Secret names と vars を分離して検証**
  - [ ] 対象の `backend/wrangler.jsonc` または `backend/wrangler.production.jsonc` の `secrets.required` を names-only SSOT とし、値を表示せず全nameを確認する。
  - [ ] `vars` は別に検証する。`DB_PORT` / `DB_NAME` / `DB_SSL_MODE` / `S3_SHARED_BUCKET` 等の非secret varsを `wrangler secret put` へ送らない。
- [ ] **DB マイグレーションの整合性**
  - [ ] このHEADの現行 `backend/migrations/*.sql` からfresh DB schemaを構築でき、migration-key coverageが `missing=0` になることを承認済み検証で確認する。固定table数を合格条件にしない。
  - [ ] 初期マスタデータ（`backend/migrations/seeds/002_master/`）の投入準備完了。

---

## 3. リリース実行手順

1. **ブランチ管理**: 日常開発は `main`。STGはreview済みの `main -> staging` PRを使う。
   - **Production stop**: backendとfrontendのproduction approval gateが両方実装・検証されるまで、`production` へmerge/pushしない。現行 frontend workflow は `Production` Environment に bind し、production dispatch を production ref に限定する。ただし Required reviewers の有効性と backend production gate は別途確認が必要。
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
