---
description: デプロイ準備チェック（staging / production）
---

# デプロイ準備

$ARGUMENTS 環境へのデプロイを準備してください。

## 事前チェック

1. 現在のブランチを確認
2. 未コミットの変更がないことを確認
3. すべてのテストがパスすることを確認
4. ビルドが成功することを確認

## デプロイ手順

1. 環境: $ARGUMENTS
2. ビルド確認: `docker compose exec frontend pnpm build`
3. デプロイ: GitHub Actions `backend-deploy.yml` ワークフローをトリガー
   ```bash
   # main push で自動実行、または手動トリガー
   gh workflow run backend-deploy.yml
   gh run watch
   ```

## 注意事項

- 本番デプロイには確認が必要
- ロールバック手順を表示
- デプロイ後のチェックリストを提示

## 出力

- デプロイ準備状況
- 実行するコマンド
- 確認事項
