# デプロイメント・運用ドキュメント (Deployment & Operations)

> **目的**: デプロイメントハブ(環境URL・主要コマンド・関連ドキュメント索引)を提供する。
> **読者**: 全運用者。
> **タイミング**: デプロイ運用開始時。

> **Animal Ekarte**: ステージング・本番環境へのデプロイと安定稼働のためのガイド
> **最新更新**: 2026-06-12 | **ステータス**: 完全自動デプロイ稼働中

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
- **[ECS ロールバックランブック (BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)](./runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)**: 旧 AWS ECS/RDS 経路への緊急ロールバック手順（通常運用では不要）。
- **[スモークテスト手順 (CRUD-SMOKE-TEST.md)](./CRUD-SMOKE-TEST.md)**: デプロイ直後の主要機能（医院/スタッフ/権限）の導線確認。
- **[混在会計スモークテスト (MIXED-PAYMENT-SMOKE-TEST.md)](./MIXED-PAYMENT-SMOKE-TEST.md)**: 混在会計 (payment_splits) の詳細動作確認。
- **[Lステップ Write API 一時停止メモ (LSTEP_WRITE_API_PAUSE.md)](./LSTEP_WRITE_API_PAUSE.md)**: Lステップへのタグ付与・解除・プロパティ更新を再有効化する前提条件。
- **[STG デモデータライフサイクル (STG-DEMO-DATA-LIFECYCLE.md)](./STG-DEMO-DATA-LIFECYCLE.md)**: Seed/Demo/Smoke テストデータの分類、作成元、Cleanup 方針、DB_RESET 機構。
- **[STG 継続運用チェックリスト (STG-CONTINUOUS-OPERATIONS.md)](./STG-CONTINUOUS-OPERATIONS.md)**: 日次/週次/月次の STG 環境監視・検査・メンテナンス。
- **[Vercel フロントエンド検証手順 (VERCEL-FRONTEND-STAGING-TEST.md)](./VERCEL-FRONTEND-STAGING-TEST.md)**: デプロイ後の UI・ログイン・API 連携検証。
- **[休憩時間データ形状監査 (BREAK-HOURS-SHAPE-AUDIT.md)](./BREAK-HOURS-SHAPE-AUDIT.md)**: R1-3 デプロイ前の STG/本番 break_hours 形状監査手順。
- **[Delete / Soft Delete 設計パターン (../../DELETE_SOFT_DELETE_PATTERNS.md)](../../DELETE_SOFT_DELETE_PATTERNS.md)**: Hard Delete と Soft Delete の使い分け、FK 制約との関係、実装パターン、STG-001 教訓。

> PR #49 Post-Merge Smoke Checklist（PR固有の使い切りチェックリスト）・CRUD スモーク自動化戦略
> （§3.4 に記載の通り自動化自体が撤去済みで計画倒れとなった歴史的記録）は特定PR/時点のスナップショット
> のため `docs/archive/` へ退役済み。

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

### 3.4 自動スモークテストの実行 (手動トリガー)
```bash
# STG 疎通確認（/health）
gh workflow run stg-smoke.yml
```
> 旧 stg-health-check / stg-readonly-smoke / stg-crud-smoke の3本は `stg-smoke.yml` に統合後、
> login/readonly/CRUD は `STG_DEMO_*` secret 未設定で約1年間機能していなかったためデッドコードとして撤去
> （現状は health 疎通のみ）。CRUD の正しさは backend unit/integration テスト + FE route-guard テストでカバー。
> CRUD smoke を復活させる場合は `STG_DEMO_EMAIL`/`STG_DEMO_PASSWORD` を設定し git 履歴 `281a561e` を参照。

---

## 4. デプロイ後のロールバック判定フレームワーク

### 4.1 ヘルスチェック手順

デプロイ完了直後、以下の順序でシステム稼働状態を確認してください。

1.  **API ヘルスチェック** (`/health` エンドポイント):
    ```bash
    curl -s https://api.stg.noah-karte.com/health | jq '.status'
    # 期待: "ok"
    ```
    **失敗時アクション**: HTTP 200 が返らない場合、§4.2 のロールバック判定へ。

2.  **ECS サービス稼働確認** (desired count == running count):
    ```bash
    aws ecs describe-services \
      --cluster animalekarte-stg-cluster \
      --services animalekarte-stg-service \
      --region us-east-1 \
      --query 'services[0].{desiredCount,runningCount}'
    ```
    **期待**: `desiredCount` と `runningCount` が同じ値。不一致時は §4.2 へ。

3.  **CloudWatch エラーログ監視**:
    ```bash
    aws logs tail /ecs/animalekarte-stg --region us-east-1 --follow | grep -i "error\|fatal"
    ```
    **期待**: ERROR/FATAL ログが 5 分間で 3 件以下。多発時は §4.2 へ。

4.  **(オプション) ALB ターゲット健全性確認**:
    ```bash
    aws elbv2 describe-target-health \
      --target-group-arn <TARGET_GROUP_ARN> \
      --region us-east-1
    ```
    **期待**: 全タスクが `healthy` 状態。

---

### 4.2 ロールバック 要否判定基準

以下の 6 つのいずれかに該当した場合、**即座にロールバックを実行してください**。

| # | 症状 | 判定方法 | ロールバック判定 |
|---|------|--------|-----------------|
| 1 | `/health` エンドポイント非応答 | HTTP 200 か確認 | **即ロールバック** |
| 2 | ECS desired count ≠ running count | describe-services 確認 | **即ロールバック** |
| 3 | CloudWatch ERROR/FATAL ログ多発 | 5 分間で 3 件以上 | **即ロールバック** |
| 4 | CRUD スモークテスト想定外エラー | 想定外 400/500 レスポンス | **即ロールバック** |
| 5 | FK 保護が 409 ではなく 4xx/5xx 返す | DELETE 試行時ステータス確認 | **即ロールバック** |
| 6 | データ破損・想定外削除・テナント隔離破綻 | スモークテスト後に手動確認 | **即ロールバック** |

**ロールバック実行フロー**:
1. 上記症状のいずれかを発見したら、まず **チーム内に通知** 
2. [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) §ロールバック手順に従い自動ロールバック実行
3. ロールバック後、前回デプロイ版に戻ったことを確認
4. 復帰後、インシデント報告書作成（原因分析、再発防止策記載）

---

### 4.3 ロールバック不要の条件

以下の 3 つがすべて成立した場合、デプロイ成功と判定し、運用へ移行します。

| 条件 | 確認方法 | 判定 |
|------|--------|------|
| **ヘルスチェック PASS** | §4.1 §1-3 をすべて通過 | ✅ |
| **CRUD スモークテスト PASS** | [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) を完全実行し、全ステータスコードが期待値 | ✅ |
| **テストデータ削除完了・記録済み** | §4.1 の cleanup 処理完了、削除レコード数・操作者・タイムスタンプをログ記録 | ✅ |

**3 つすべて ✅ の場合**: デプロイは本番リリース候補として承認。

---

### 4.4 認証情報保護ポリシー

デプロイ検証時、以下のいずれも **本ドキュメント・成果物・ログ出力に記録してはいけません**：

- `password` （パスワード）
- `access_token` （アクセストークン）
- `refresh_token` （リフレッシュトークン）
- Cookie の `Set-Cookie` ヘッダ値
- demo アカウントの token / cookie 値

**実装方法**:
- API 検証時は、ブラウザ DevTools の Network タブで Cookie を確認し、スクリプト/出力には含めない
- curl 実行時は `${TOKEN}` 等の環境変数を使用し、実トークン値を可視化しない
- CloudWatch ログは自動サニタイズされていることを確認（パスワード値が出力されていないこと）

---

### 4.5 参考資料

- [デプロイ手順書 (CI-CD-PIPELINE.md)](./CI-CD-PIPELINE.md)：自動デプロイ・手動トリガー・ロールバック手順
- [スモークテスト手順 (CRUD-SMOKE-TEST.md)](./CRUD-SMOKE-TEST.md)：CRUD 全操作・FK保護検証・権限テスト
- AWS CloudWatch: `/ecs/animalekarte-stg` ロググループ
- SSM Parameter Store: API 認証情報（team lead のみアクセス可）

---
