# デプロイメント・ドキュメント

Animal Ekarte のデプロイ手順・CI/CD・運用ガイド。

**ステータス:** ✅ Backend・Frontend 自動デプロイ完全稼働中

---

## アクセスURL（Test環境）

| サービス | URL |
|---------|-----|
| Frontend | https://frontend-r0m0pyiaf-minorusogas-projects.vercel.app |
| Backend API | http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com |
| Swagger UI | http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/swagger/index.html |

---

## ドキュメント

| ドキュメント | 内容 |
|-----------|------|
| [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) | 自動デプロイパイプライン、手動デプロイ、ロールバック、トラブルシューティング |
| [DEPLOYMENT-CHECKLIST.md](./DEPLOYMENT-CHECKLIST.md) | デプロイ前チェックリスト、デプロイ手順、DB作り直し |
| [deployment-status.md](./deployment-status.md) | AWSリソース一覧、既知の問題、コスト、次のアクション |

---

## クイックリファレンス

### デプロイフロー

| コンポーネント | トリガー | 方式 |
|--------------|---------|------|
| Backend API | `backend/**` → main push | GitHub Actions → ECR → ECS |
| Frontend | `frontend/**` → main push | Vercel GitHub Integration |

### よく使うコマンド

```bash
export AWS_PROFILE=AnimalEkarte

# API ヘルスチェック
curl http://animalekarte-test-alb-1778215308.us-east-1.elb.amazonaws.com/health | jq .

# ECS デプロイ状態確認
aws ecs describe-services \
  --cluster animalekarte-test-cluster \
  --services animalekarte-test-service \
  --region us-east-1 \
  --query 'services[0].{status,runningCount,desiredCount}'

# ECS ログ確認
aws logs tail /ecs/animalekarte-test --follow --region us-east-1

# Backend 手動デプロイ
gh workflow run backend-deploy.yml --ref main
```
