# AWS デプロイメントガイド

## 対応サービス

| サービス | 用途 |
|---------|------|
| EC2 | 仮想サーバー |
| ECS/Fargate | コンテナ |
| Lambda | サーバーレス |
| RDS | データベース |
| S3 | 静的ファイル |
| CloudFront | CDN |

## デプロイフロー

```
Local → GitHub → CodePipeline → CodeBuild → ECS/EC2
```

## 環境設定

### AWS CLI 設定

```bash
# 認証情報設定
aws configure

# プロファイル確認
aws sts get-caller-identity
```

### 必要な環境変数

```bash
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
AWS_REGION=ap-northeast-1
```

## ECS デプロイ

### タスク定義

```json
{
  "family": "app-task",
  "containerDefinitions": [
    {
      "name": "app",
      "image": "xxx.dkr.ecr.ap-northeast-1.amazonaws.com/app:latest",
      "portMappings": [
        { "containerPort": 3000 }
      ],
      "environment": [
        { "name": "NODE_ENV", "value": "production" }
      ]
    }
  ]
}
```

### デプロイコマンド

```bash
# ECR にイメージプッシュ
aws ecr get-login-password | docker login --username AWS --password-stdin xxx.dkr.ecr.ap-northeast-1.amazonaws.com
docker build -t app .
docker tag app:latest xxx.dkr.ecr.ap-northeast-1.amazonaws.com/app:latest
docker push xxx.dkr.ecr.ap-northeast-1.amazonaws.com/app:latest

# ECS サービス更新
aws ecs update-service --cluster my-cluster --service my-service --force-new-deployment
```

## RDS 設定

### 接続文字列

```
postgresql://user:password@xxx.ap-northeast-1.rds.amazonaws.com:5432/dbname
```

### セキュリティグループ

- ECS タスクからのインバウンド許可（ポート 5432）
- VPC 内部のみアクセス可能に設定

## ロールバック

```bash
# 前のタスク定義に戻す
aws ecs update-service --cluster my-cluster --service my-service --task-definition app-task:123
```

## モニタリング

- CloudWatch Logs でログ確認
- CloudWatch Metrics でメトリクス監視
- X-Ray でトレーシング

## コスト最適化

- Fargate Spot の活用
- Auto Scaling の設定
- Reserved Instances の検討
