# Azure デプロイメントガイド

## 対応サービス

| サービス | 用途 |
|---------|------|
| App Service | PaaS（推奨） |
| Container Apps | コンテナ |
| AKS | Kubernetes |
| Functions | サーバーレス |
| Azure SQL | データベース |
| Blob Storage | 静的ファイル |

## デプロイフロー

```
Local → GitHub → Azure DevOps/GitHub Actions → App Service/Container Apps
```

## 環境設定

### Azure CLI 設定

```bash
# ログイン
az login

# サブスクリプション設定
az account set --subscription SUBSCRIPTION_ID

# 確認
az account show
```

### 必要な環境変数

```bash
AZURE_SUBSCRIPTION_ID=xxx
AZURE_TENANT_ID=xxx
AZURE_CLIENT_ID=xxx
AZURE_CLIENT_SECRET=xxx
```

## App Service デプロイ

### リソース作成

```bash
# リソースグループ
az group create --name myResourceGroup --location japaneast

# App Service プラン
az appservice plan create \
  --name myPlan \
  --resource-group myResourceGroup \
  --sku B1 \
  --is-linux

# Web App
az webapp create \
  --name myApp \
  --resource-group myResourceGroup \
  --plan myPlan \
  --runtime "NODE:20-lts"
```

### デプロイコマンド

```bash
# ZIP デプロイ
az webapp deployment source config-zip \
  --resource-group myResourceGroup \
  --name myApp \
  --src app.zip

# Git デプロイ
az webapp deployment source config-local-git \
  --name myApp \
  --resource-group myResourceGroup
```

### 環境変数設定

```bash
az webapp config appsettings set \
  --name myApp \
  --resource-group myResourceGroup \
  --settings NODE_ENV=production DATABASE_URL=xxx
```

## Container Apps デプロイ

```bash
# Container Apps 環境作成
az containerapp env create \
  --name myEnv \
  --resource-group myResourceGroup \
  --location japaneast

# アプリデプロイ
az containerapp create \
  --name myApp \
  --resource-group myResourceGroup \
  --environment myEnv \
  --image myregistry.azurecr.io/app:latest \
  --target-port 3000 \
  --ingress external
```

## Azure SQL 設定

### 接続文字列

```
Server=tcp:xxx.database.windows.net,1433;Database=mydb;User ID=admin;Password=xxx;Encrypt=true;
```

### ファイアウォール設定

```bash
az sql server firewall-rule create \
  --resource-group myResourceGroup \
  --server myServer \
  --name AllowAzureServices \
  --start-ip-address 0.0.0.0 \
  --end-ip-address 0.0.0.0
```

## ロールバック

```bash
# デプロイスロット使用
az webapp deployment slot swap \
  --name myApp \
  --resource-group myResourceGroup \
  --slot staging \
  --target-slot production
```

## モニタリング

- Application Insights でログ/メトリクス
- Azure Monitor でアラート設定
- Log Analytics でクエリ分析

## コスト最適化

- 開発環境は B1 プラン使用
- 本番は Auto Scale 設定
- Reserved Instances の検討
