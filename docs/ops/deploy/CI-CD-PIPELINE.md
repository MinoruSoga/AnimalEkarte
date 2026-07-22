# CI/CD パイプライン構成書 (CI/CD Pipeline)

> **目的**: 自動デプロイ・手動トリガー・ロールバック手順を定義する。
> **読者**: 運用者・新規参加者。
> **タイミング**: パイプライン理解時・手動操作時。

> **Animal Ekarte**: GitHub Actions と Cloudflare/Vercel による自動デプロイ（バックエンドは Cloudflare Workers + Containers が標準運用、AWS ECS はロールバック専用）
> **最新更新**: 2026-07-22 | **ステータス**: 標準化済み（Cloudflare 本運用 + ECS ロールバック）

---

## 1. 全体フロー概要

本システムは、`staging` ブランチへの push をトリガーとして、バックエンドとフロントエンドがそれぞれ独立した GitHub Actions ワークフローで自動デプロイされます。

| コンポーネント | 実行環境 | デプロイ方式 | トリガー | ワークフロー |
|:---|:---|:---|:---|:---|
| **Backend API** | Cloudflare Workers + Containers | `wrangler deploy` + migrate one-shot + `/health` ポーリング | `staging` ブランチへの push | `.github/workflows/backend-deploy.yml`（ロールバック時のみ: `backend-deploy-ecs.yml`, `workflow_dispatch`） |
| **Frontend** | Vercel | Vercel CLI 経由デプロイ | `staging` ブランチへの push | `.github/workflows/frontend-deploy.yml` |

> **2026-07-06**: Backend の主経路を AWS ECS から Cloudflare Workers/Containers へ移行済み（`migration-cloudflare.md` Phase 5）。
> 以下 §2 は **Cloudflare 正規フロー**と**AWS ECS ロールバック専用フロー**を分離して記載している。

---

## 2. バックエンド・パイプライン（Cloudflare 正規経路）

### 2.1 実行ステップ
1. **Checkout**: ソースコードの取得。
2. **Auth / Tooling**: pnpm・Node のセットアップと依存関係インストール。
3. **認証チェック**: `CLOUDFLARE_API_TOKEN` 未設定は即時失敗。`wrangler whoami` で検証。
4. **Deploy**: `npx wrangler deploy` で Worker と Container を同時デプロイ。
5. **DB migrate**: `MIGRATE_RUN_SECRET` を使って `POST /_internal/migrate`（`cf-run-migrate.sh`）を実行。
   旧 `docker compose run` 方式（`db_reset=true`）は正規経路に存在しない。
6. **ヘルスチェック**: `WORKER_URL/health` を 200/`status: ok` までポーリング。
7. **スモーク（任意）**: `STG_DEMO_EMAIL/STG_DEMO_PASSWORD` が設定されている場合のみ `cf-crud-smoke.sh` を実行。

### 2.2 監視とリカバリ
- ジョブは「デプロイ → migrate → `/health` ポーリング」の順で成功条件を満たすと完了扱い。
- 既知の既存制約として「デプロイ直後〜migrate完了間」に古いバージョントラフィック到達が残る可能性があるため、重大障害時は AWS ECS ロールバック経路に即時切替する。

---

## 3. バックエンド・パイプライン（AWS ECS ロールバック専用）

### 3.1 実行ステップ（運用時は例外時のみ）
1. `.github/workflows/backend-deploy-ecs.yml` を `workflow_dispatch` で起動。
2. 既定どおり OIDC 認証、ECR Build、task-definition 更新、Service 安定化待機を実行。
3. 必要に応じて `workflow_dispatch` の `db_reset=true` で ECS 側の再初期化を実施（Cloudflare 経路には該当オプションなし）。
4. 緊急停止が必要な場合のみ `staging-stop.yml`（`desiredCount=0` 系）を利用。

### 3.2 運用上の位置づけ
- **通常運用**: 参照しない。
- **移行期間の緊急時回避策**: `backend-deploy-ecs.yml` を停止手段/ロールバック手段として維持（Phase 8 方針に従って順次縮退予定）。

---

## 4. フロントエンド・パイプライン (Vercel)

### 4.1 実行ステップ
`.github/workflows/frontend-deploy.yml` が Vercel CLI を GitHub Actions 上で実行します
（Vercel 側のネイティブ Git 連携フックは使用していません）。
1. **Detection**: `staging`/`production` ブランチへの push を検知。
2. **Pull**: `vercel pull` で環境情報を取得（`production` ブランチのみ `--environment=production`）。
3. **Build**: `vercel build` を実行（`pnpm build` 相当。TypeScript 型チェック・Vite 最適化を含む）。
4. **Deploy**: `vercel deploy --prebuilt` でビルド済み成果物をデプロイ。

---

## 5. 運用コマンド

### 5.1 手動デプロイの強制実行
自動検知が機能しない場合や、特定の環境変数を反映させたい場合に使用します。
```bash
# 特定のワークフローを手動起動
gh workflow run backend-deploy.yml --ref staging
```

### 5.2 バックエンド・ロールバック（ECS）
```bash
aws ecs update-service --cluster animalekarte-stg-cluster \
  --service animalekarte-stg-service \
  --task-definition animalekarte-stg-api:<PREVIOUS_REVISION>
```

### 5.3 STG 環境の手動停止（旧 AWS ECS/RDS 経路）
`staging-stop.yml` は DEPRECATED（Cloudflare Containers は scale-to-zero 自動）であり、
ECS ロールバック経路維持中に緊急停止が必要な場合のみ実行する:
```bash
gh workflow run staging-stop.yml --ref staging
```

---

## 6. セキュリティと認証

- **Cloudflare 正系統**: `JWT_SECRET` / DB 接続情報 / `MIGRATE_RUN_SECRET` / `CLOUDFLARE_API_TOKEN` は `wrangler secret put` / GitHub Encrypted Secrets で管理。`SSM Parameter Store` は未使用。
- **AWS ECS ロールバック経路（旧経路）**: `JWT_SECRET` / DB 接続情報は `AWS SSM Parameter Store` 管理、OIDC は同経路のみで利用。

---
