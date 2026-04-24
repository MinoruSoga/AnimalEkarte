---
name: Deployment
description: AWS ECS（GitHub Actions経由）へのアプリケーションデプロイ
---

# デプロイメントスキル

> AnimalEkarte は **AWS ECS** にデプロイ。GitHub Actions ワークフロー経由で自動化。

## このスキルを使用するタイミング

- 新しいバージョンのデプロイ
- インフラストラクチャの更新
- 環境変数の管理

## デプロイフロー（実際）

```
git push → GitHub Actions (.github/workflows/backend-deploy.yml)
    ↓
Docker Build → ECR Push → ECS Deploy
    ↓
Verify deployment
```

## デプロイ方法

デプロイは **GitHub Actions で自動実行**（`scripts/deploy.sh` は存在しない）。

```bash
# ステージング: main ブランチへの push で自動デプロイ
git push origin main

# 手動トリガー (workflow_dispatch)
# GitHub Actions → backend-deploy.yml → Run workflow

# CI ステータス確認
gh run list --workflow=backend-deploy.yml
gh run watch
```

## CI パイプライン（ci.yml）

```
PR push → Backend (go build / go test / golangci-lint / schema-check)
        → Frontend (pnpm install / lint / build)
        → Codegen check (make codegen-check)
```

## 重要な注意事項

- 必ずステージングで検証してから本番へ
- ECS タスク定義のロールバックは AWS Console または `aws ecs update-service`
- デプロイ後はモニタリングダッシュボードを確認

## プラットフォーム参照

- AWS ECS: [platforms/aws.md](./platforms/aws.md)
- GCP (参考): [platforms/gcp.md](./platforms/gcp.md)
