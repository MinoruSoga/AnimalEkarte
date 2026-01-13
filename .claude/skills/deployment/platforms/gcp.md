# GCP デプロイメントガイド

## 対応サービス

| サービス | 用途 |
|---------|------|
| Cloud Run | コンテナ（推奨） |
| GKE | Kubernetes |
| App Engine | PaaS |
| Cloud Functions | サーバーレス |
| Cloud SQL | データベース |
| Cloud Storage | 静的ファイル |

## デプロイフロー

```
Local → GitHub → Cloud Build → Cloud Run/GKE
```

## 環境設定

### gcloud CLI 設定

```bash
# 認証
gcloud auth login

# プロジェクト設定
gcloud config set project PROJECT_ID

# 確認
gcloud config list
```

### 必要な環境変数

```bash
GOOGLE_CLOUD_PROJECT=xxx
GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
```

## Cloud Run デプロイ

### Dockerfile

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 8080
CMD ["npm", "start"]
```

### デプロイコマンド

```bash
# イメージビルド&プッシュ
gcloud builds submit --tag gcr.io/PROJECT_ID/app

# Cloud Run デプロイ
gcloud run deploy app \
  --image gcr.io/PROJECT_ID/app \
  --platform managed \
  --region asia-northeast1 \
  --allow-unauthenticated
```

### 環境変数設定

```bash
gcloud run services update app \
  --set-env-vars="NODE_ENV=production,DATABASE_URL=xxx"
```

## Cloud SQL 設定

### 接続

```bash
# Cloud SQL Proxy 使用
cloud_sql_proxy -instances=PROJECT:REGION:INSTANCE=tcp:5432

# 接続文字列
postgresql://user:password@localhost:5432/dbname
```

### Cloud Run からの接続

```bash
gcloud run services update app \
  --add-cloudsql-instances=PROJECT:REGION:INSTANCE \
  --set-env-vars="DATABASE_URL=/cloudsql/PROJECT:REGION:INSTANCE"
```

## ロールバック

```bash
# リビジョン一覧
gcloud run revisions list --service app

# 特定リビジョンにトラフィック切り替え
gcloud run services update-traffic app --to-revisions=app-00001-xxx=100
```

## モニタリング

- Cloud Logging でログ確認
- Cloud Monitoring でメトリクス監視
- Cloud Trace でトレーシング

## コスト最適化

- 最小インスタンス数を 0 に設定
- CPU スロットリングの活用
- Committed Use Discounts の検討
