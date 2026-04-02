# デプロイメント・ドキュメント

Animal Ekarte のデプロイ手順・CI/CD・運用ガイド。

**ステータス:** ✅ Backend・Frontend 自動デプロイ完全稼働中

---

## アクセスURL（Stg環境）

| サービス | URL |
|---------|-----|
| Frontend | https://stg.noah-karte.com |
| Backend API | https://api.stg.noah-karte.com/api |
| ALB（直接・ヘルスチェック用） | http://animalekarte-stg-alb-1915768826.us-east-1.elb.amazonaws.com |
| API仕様 | `backend/docs/api.yaml`（OpenAPI） |

---

## ドキュメント

| ドキュメント | 内容 |
|-----------|------|
| [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) | 自動デプロイパイプライン、手動デプロイ、ロールバック、トラブルシューティング |
| [DEPLOYMENT-CHECKLIST.md](./DEPLOYMENT-CHECKLIST.md) | デプロイ前チェックリスト、デプロイ手順、DB作り直し |
| [docs_infra_architecture.md](../docs_infra_architecture.md) | AWSリソース一覧、インフラ構成、セキュリティ、コスト |

---

## クイックリファレンス

### デプロイフロー

| コンポーネント | トリガー | 方式 |
|--------------|---------|------|
| Backend API | `backend/**` → staging push | GitHub Actions → ECR → ECS |
| Frontend | `frontend/**` → staging push | Vercel GitHub Integration |

### よく使うコマンド

```bash
export AWS_PROFILE=AnimalEkarte

# API ヘルスチェック（ALB 直接）
curl http://animalekarte-stg-alb-1915768826.us-east-1.elb.amazonaws.com/health | jq .

# API ヘルスチェック（CloudFront 経由・本番同等）
curl https://api.stg.noah-karte.com/health | jq .

# ECS デプロイ状態確認
aws ecs describe-services \
  --cluster animalekarte-stg-cluster \
  --services animalekarte-stg-service \
  --region us-east-1 \
  --query 'services[0].{status,runningCount,desiredCount}'

# ECS ログ確認
aws logs tail /ecs/animalekarte-stg --follow --region us-east-1

# Backend 手動デプロイ
gh workflow run backend-deploy.yml --ref staging
```
