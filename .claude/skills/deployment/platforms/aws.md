# AWS デプロイメントガイド（AnimalEkarte 実態）

> このプロジェクトの AWS デプロイは **GitHub Actions ワークフロー**（`.github/workflows/backend-deploy.yml`）が実行する。
> CodePipeline / CodeBuild は使用していない。手動での `aws` CLI 操作は調査・障害対応目的に限定する。

## 使用サービス

| サービス | 用途 |
|---------|------|
| ECS/Fargate | バックエンド API コンテナ実行 |
| ECR | Docker イメージレジストリ |
| RDS (PostgreSQL) | データベース（ステージング: `animalekarte-stg-db`） |
| CloudWatch Logs | ログ確認（ロググループ `/ecs/animalekarte-stg`） |

フロントエンドは AWS ではなく **Vercel** にデプロイされる（`.github/workflows/frontend-deploy.yml` 参照）。本ファイルはバックエンド（ECS）のみを対象とする。

## デプロイフロー（実態）

```
git push (staging ブランチ, backend/** 変更)
  → GitHub Actions (backend-deploy.yml)
    → OIDC で AWS 認証 (aws-actions/configure-aws-credentials, role-to-assume)
    → ECR ログイン・Docker イメージビルド & push
    → RDS 停止状態なら起動 (preflight)
    → .env.staging → ECS タスク定義の environment に変換
    → migrate タスク定義を登録し ECS RunTask でマイグレーション実行・完了待ち
    → ECS service の desiredCount を確認（0 なら 1 に戻す）
    → amazon-ecs-deploy-task-definition で API タスク定義を更新・デプロイ
    → runningCount を検証
  → (時間外デプロイの場合) 30分後に ECS desiredCount=0 / RDS 停止
```

## 認証（OIDC・キー不要）

`aws configure` や `AWS_ACCESS_KEY_ID` は使用しない。GitHub Actions の OIDC + IAM ロールで認証する。

```yaml
permissions:
  id-token: write
  contents: read

- uses: aws-actions/configure-aws-credentials@v6.1.0
  with:
    role-to-assume: arn:aws:iam::698109622668:role/animalekarte-stg-github-ecs-deploy-role
    aws-region: us-east-1
```

## デプロイ方法

デプロイは **GitHub Actions で自動実行**。手動での `aws ecs update-service` によるデプロイは行わない。

```bash
# ステージング: staging ブランチへの push で backend/** 変更時に自動デプロイ
git push origin staging

# 手動トリガー（DB リセットが必要な場合など）
# GitHub Actions → backend-deploy.yml → Run workflow → db_reset: true/false

# CI ステータス確認
gh run list --workflow=backend-deploy.yml
gh run watch
```

## 主要ステップの内訳

### 1. イメージビルド & ECR push

```bash
docker buildx build --platform linux/amd64 \
  -f backend/Dockerfile.production \
  -t <ECR_REGISTRY>/animalekarte-api:$GITHUB_SHA \
  -t <ECR_REGISTRY>/animalekarte-api:latest \
  --push ./backend
```

### 2. マイグレーション（デプロイ前必須）

`animalekarte-stg-migrate` タスク定義を ECS RunTask (Fargate) で実行し、`STOPPED` になるまでポーリング。`exitCode != 0` は即座にログを取得して fail。

### 3. API デプロイ

`aws-actions/amazon-ecs-deploy-task-definition` で `animalekarte-stg-service` を更新。`wait-for-service-stability: true` でロールアウト完了を待機。

## ロールバック

```bash
# 直前のタスク定義リビジョンに戻す
aws ecs update-service --cluster animalekarte-stg-cluster \
  --service animalekarte-stg-service \
  --task-definition animalekarte-stg-api:<直前のリビジョン番号>
```

## モニタリング

```bash
# マイグレーション/API ログ確認
aws logs get-log-events \
  --log-group-name /ecs/animalekarte-stg \
  --log-stream-name <stream-name>

# サービス状態確認
aws ecs describe-services \
  --cluster animalekarte-stg-cluster \
  --services animalekarte-stg-service \
  --query 'services[0].{status:status,runningCount:runningCount,desiredCount:desiredCount}'
```

## ステージングのコスト最適化（実装済み）

- 営業時間外 (JST 08:00-22:00 の外) にデプロイした場合、デプロイ完了 30 分後に ECS `desiredCount=0` / RDS 停止を自動実行（`backend-deploy.yml` の `delayed-stop` job）。
- デプロイ時に RDS が停止していれば自動起動（`rds-preflight` step）。

## 注意事項

- 本番環境への直接 push は禁止（`main` → `staging` は PR 経由、`production` への直接 push 禁止は `.claude/CLAUDE.md` 参照）。
- `.env.staging` の内容がそのまま ECS タスク定義の `environment` に展開されるため、秘密情報を平文で置かない（Secrets Manager 未導入は既知の技術的負債）。
