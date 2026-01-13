---
name: Deployment
description: AWS/GCP/Azureへのアプリケーションデプロイ
---

# デプロイメントスキル

## このスキルを使用するタイミング

- 新しいバージョンのデプロイ
- インフラストラクチャの更新
- 環境変数の管理

## デプロイフロー

```
1. テスト実行 → 2. ビルド → 3. ステージングデプロイ
                                    ↓
                             4. 検証
                                    ↓
                             5. 本番デプロイ
```

## 環境別コマンド

### ステージング

```bash
./scripts/deploy.sh staging
```

### 本番

```bash
# 必ず事前承認を取得
./scripts/deploy.sh production
```

## 重要な注意事項

- 必ずステージングで検証してから本番へ
- ロールバック手順を確認
- デプロイ後はモニタリングダッシュボードを確認

## プラットフォームガイド

- AWS: [aws/README.md](./aws/README.md)
- GCP: [gcp/README.md](./gcp/README.md)
- Azure: [azure/README.md](./azure/README.md)
