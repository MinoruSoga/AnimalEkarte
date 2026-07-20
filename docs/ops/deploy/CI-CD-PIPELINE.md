# CI/CD パイプライン構成書 (CI/CD Pipeline)

> **目的**: 自動デプロイ・手動トリガー・ロールバック手順を定義する。
> **読者**: 運用者・新規参加者。
> **タイミング**: パイプライン理解時・手動操作時。

> **Animal Ekarte**: GitHub Actions と Cloudflare/Vercel による自動デプロイ（バックエンドの旧 AWS ECS 経路はロールバック専用）
> **最新更新**: 2026-07-10 | **ステータス**: 完全自動化運用中

---

## 1. 全体フロー概要

本システムは、`staging` ブランチへのプッシュをトリガーとして、バックエンドとフロントエンドがそれぞれ独立した GitHub Actions ワークフローで自動デプロイされます。

| コンポーネント | 実行環境 | デプロイ方式 | トリガー | ワークフロー |
|:---|:---|:---|:---|:---|
| **Backend API** | Cloudflare Workers + Containers | `wrangler deploy` + migrate one-shot + `/health` ポーリング | `staging` ブランチへの push | `.github/workflows/backend-deploy.yml`（ECS ロールバック版: `backend-deploy-ecs.yml`, `workflow_dispatch`のみ） |
| **Frontend** | Vercel | Vercel CLI 経由デプロイ | `staging` ブランチへの push | `.github/workflows/frontend-deploy.yml` |

> **2026-07-06**: Backend の主経路を AWS ECS から Cloudflare Workers/Containers へ移行済み（`../infra/_archive/migration-cloudflare.md` Phase 5）。
> 以下 §2 の手順詳細は現時点では `backend-deploy-ecs.yml`（ECS ロールバック専用ワークフロー）向けの記述として維持している（Phase 8 で ECS 撤去後に本節を全面更新予定）。

---

## 2. バックエンド・パイプライン (AWS ECS ロールバック専用経路)

### 2.1 実行ステップ
1.  **JST営業時間外検知**: JST 08:00–22:00 外の場合はデプロイ後 30 分で自動停止する `delayed-stop` ジョブを起動。
2.  **Checkout**: ソースコードの取得。
3.  **Auth (OIDC)**: AWS OIDC を使用し、秘密鍵を使わずに一時的なロールで認証。
4.  **Preflight — RDS 起動確認**: RDS が停止状態の場合は起動し、利用可能になるまで待機。
5.  **Build & Push**: `Dockerfile.production` を用いてイメージをビルドし、Amazon ECR へプッシュ。
6.  **Migrate**: 新しいイメージで DB マイグレーションタスクを先行実行（タイムアウト 15 分）。
7.  **Preflight — ECS desiredCount 確認**: `desiredCount=0`（停止中）の場合は 1 へ引き上げてからデプロイ。
8.  **Service Update**: `aws-actions/amazon-ecs-deploy-task-definition` でタスク定義を更新し、
    `wait-for-service-stability: true` でローリングアップデート完了まで待機（Blue/Green ではない。
    CodeDeploy 等の専用デプロイコントローラは未使用）。

### 2.2 監視とリカバリ
- サービス安定化を待たずにジョブが失敗した場合、GitHub Actions は失敗としてマークされます
  （明示的なタイムアウト値は設定されておらず、AWS SDK のポーリング既定動作に従う）。
- 重大な不具合発見時は、AWS CLI を使用して前バージョンの Task Definition に即座に切り戻します（§4.2）。
- `delayed-stop` ジョブ: JST 営業時間外は 30 分待機後に ECS `desiredCount=0` と RDS 停止を自動実行。
  手動で即時停止したい場合は `staging-stop.yml` を実行（§4.3）。

---

## 3. フロントエンド・パイプライン (Vercel)

### 3.1 実行ステップ
`.github/workflows/frontend-deploy.yml` が Vercel CLI を GitHub Actions 上で実行します
（Vercel 側のネイティブ Git 連携フックは使用していません）。
1.  **Detection**: `staging`/`production` ブランチへの push を検知。
2.  **Pull**: `vercel pull` で環境情報を取得（`production` ブランチのみ `--environment=production`）。
3.  **Build**: `vercel build` を実行（`pnpm build` 相当。TypeScript 型チェック・Vite 最適化を含む）。
4.  **Deploy**: `vercel deploy --prebuilt` でビルド済み成果物をデプロイ。

---

## 4. 運用コマンド

### 4.1 手動デプロイの強制実行
自動検知が機能しない場合や、特定の環境変数を反映させたい場合に使用します。
```bash
# 特定のワークフローを手動起動
gh workflow run backend-deploy.yml --ref staging
```

### 4.2 バックエンド・ロールバック
```bash
# 前のタスク定義リビジョンへ強制指定して更新
aws ecs update-service --cluster animalekarte-stg-cluster \
  --service animalekarte-stg-service \
  --task-definition animalekarte-stg-api:<PREVIOUS_REVISION>
```

### 4.3 STG 環境の手動停止（旧 AWS ECS/RDS 経路・DEPRECATED）
`staging-stop.yml` はファイル冒頭で DEPRECATED と明記されている：Cloudflare Containers は
scale-to-zero のため夜間停止スケジューラ自体が不要。Phase 8（AWS 廃止）完了後に削除予定。
旧 ECS/RDS ロールバック経路を使用中に緊急停止したい場合のみ実行する:
```bash
gh workflow run staging-stop.yml --ref staging
```

---

## 5. セキュリティと認証

- **Cloudflare 正系統**: `JWT_SECRET` やデータベース接続情報は `wrangler secret put` で Cloudflare 側に投入し、`MIGRATE_RUN_SECRET`/`CLOUDFLARE_API_TOKEN` 等は GitHub Encrypted Secrets で管理する。SSM Parameter Store は使用しない。
- **AWS ECS ロールバック経路（旧経路）**: `JWT_SECRET` やデータベース接続情報は GitHub には置かず、AWS SSM Parameter Store に保管されている。OIDC 連携（リポジトリオーナー `MinoruSoga` の特定リポジトリからのアクセスのみを許可する信頼関係ポリシー）も同経路のみで使用する。

---
