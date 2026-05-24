# デプロイメント・運用ドキュメント (Deployment & Operations)

> **Animal Ekarte**: ステージング・本番環境へのデプロイと安定稼働のためのガイド
> **最新更新**: 2026-05-21 | **ステータス**: 完全自動デプロイ稼働中

---

## 1. 稼働環境一覧

| 環境 | Frontend URL | API Base URL | インフラ管理 |
|:---|:---|:---|:---|
| **Staging** | [stg.noah-karte.com](https://stg.noah-karte.com) | [api.stg.noah-karte.com/api](https://api.stg.noah-karte.com/api) | AWS (us-east-1) / Vercel |
| **Production** | [noah-karte.com](https://noah-karte.com) | [api.noah-karte.com/api](https://api.noah-karte.com/api) | 同上 |

---

## 2. ドキュメント体系

目的別に関連ドキュメントを参照してください。

- **[デプロイ手順書 (CI-CD-PIPELINE.md)](./CI-CD-PIPELINE.md)**: 自動デプロイ、手動トリガー、ロールバックの手順。
- **[事前確認リスト (../../DEPLOYMENT_CHECKLIST.md)](../../DEPLOYMENT_CHECKLIST.md)**: デプロイ前の動作確認、DBリセット、疎通確認。
- **[リリースマニュアル (STG_PRE_DEPLOY_READINESS_CHECK.md)](./runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md)**: 本番反映前の最終検証ランブック。
- **[スモークテスト手順 (CRUD-SMOKE-TEST.md)](./CRUD-SMOKE-TEST.md)**: デプロイ直後の主要機能（医院/スタッフ/権限）の導線確認。
- **[混在会計スモークテスト (MIXED-PAYMENT-SMOKE-TEST.md)](./MIXED-PAYMENT-SMOKE-TEST.md)**: 混在会計 (payment_splits) の詳細動作確認。
- **[PR #49 Post-Merge Smoke Checklist (PR49_POST_MERGE_SMOKE_TEST.md)](./PR49_POST_MERGE_SMOKE_TEST.md)**: PR #49 merge 後の総合スモークチェックリスト（予約・会計・返金・CRUD・入院・健診・seed）。
- **[Lステップ Write API 一時停止メモ (LSTEP_WRITE_API_PAUSE.md)](./LSTEP_WRITE_API_PAUSE.md)**: Lステップへのタグ付与・解除・プロパティ更新を再有効化する前提条件。

---

## 3. クイック・コマンドリファレンス

運用中によく使用する AWS CLI コマンドです。

### 3.1 サービス状態の確認
```bash
export AWS_PROFILE=AnimalEkarte

# ECS サービスのデプロイ状況とタスク数を確認
aws ecs describe-services \
  --cluster animalekarte-stg-cluster \
  --services animalekarte-stg-service \
  --region us-east-1 \
  --query 'services[0].{status,runningCount,desiredCount}'
```

### 3.2 ログのリアルタイム監視
```bash
# Backend API の標準出力をフォロー
aws logs tail /ecs/animalekarte-stg --follow --region us-east-1
```

### 3.3 手動デプロイの実行
```bash
# GitHub Actions のワークフローを staging ブランチで起動
gh workflow run backend-deploy.yml --ref staging
```

---

## 4. トラブルシューティングの第一歩

1.  **ヘルスチェック確認**: `https://api.stg.noah-karte.com/health` が `"status": "ok"` を返すか。
2.  **DB接続確認**: ECS タスクが起動失敗している場合、SSM Parameter Store の認証情報が正しいか確認。
3.  **CORSエラー**: フロントエンドのドメインが `CORS_ALLOWED_ORIGIN` 環境変数に含まれているか確認。

---
