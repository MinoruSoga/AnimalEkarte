# CI/CD パイプライン構成書 (CI/CD Pipeline)

> **Animal Ekarte**: GitHub Actions と AWS/Vercel による高速かつ安全な自動デプロイ
> **最新更新**: 2026-06-12 | **ステータス**: 完全自動化運用中

---

## 1. 全体フロー概要

本システムは、ブランチへのプッシュをトリガーとして、バックエンドとフロントエンドがそれぞれの最適化された経路で自動デプロイされます。

| コンポーネント | 実行環境 | デプロイ方式 | トリガー |
|:---|:---|:---|:---|
| **Backend API** | AWS ECS (Fargate) | GitHub Actions + ECR | `staging` ブランチへの push |
| **Frontend** | Vercel | Vercel GitHub Hook | `staging` ブランチへの push |

---

## 2. バックエンド・パイプライン (AWS ECS)

### 2.1 実行ステップ
1.  **Checkout**: ソースコードの取得。
2.  **Auth (OIDC)**: AWS OIDC を使用し、秘密鍵を使わずに一時的なロール（`animalekarte-stg-github-ecs-deploy-role`）で認証。
3.  **Build & Push**: `Dockerfile.production` を用いてイメージをビルドし、Amazon ECR へプッシュ。
4.  **Migrate**: 新しいイメージを使用して、DB マイグレーションタスクを先行実行。
5.  **Service Update**: ECS サービスを更新し、Blue/Green 形式で段階的にタスクを入れ替え。

### 2.2 監視とリカバリ
- サービスが 10 分以内に `STABLE` 状態にならない場合、GitHub Actions は失敗としてマークされます。
- 重大な不具合発見時は、AWS CLI を使用して前バージョンの Task Definition に即座に切り戻します。

---

## 3. フロントエンド・パイプライン (Vercel)

### 3.1 実行ステップ
1.  **Detection**: `frontend/` ディレクトリ内の変更を検知。
2.  **Build**: `pnpm build` を実行。TypeScript の型チェックと Vite による最適化（Minify/Tree-shaking）を実施。
3.  **Edge Deploy**: Vercel のグローバルエッジネットワークへ即座に配信。

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

---

## 5. セキュリティと認証

- **秘密情報の分離**: `JWT_SECRET` やデータベース接続情報は GitHub には置かず、AWS SSM Parameter Store に厳重に保管されています。
- **OIDC 連携**: リポジトリオーナー `MinoruSoga` の特定リポジトリからのアクセスのみを許可する信頼関係ポリシーを適用済みです。

---
