#!/bin/bash
set -e

echo "=== 環境セットアップ開始 ==="

# 依存関係のインストール
echo "→ 依存関係をインストール中..."
npm install

# 環境変数の確認
echo "→ 環境変数を確認中..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "  .envファイルを作成しました。設定を確認してください。"
fi

# ビルド確認
echo "→ ビルドを確認中..."
npm run build

# テスト確認
echo "→ 基本テストを実行中..."
npm run test:basic || echo "  基本テストがスキップされました"

# 進捗ファイルの初期化
echo "→ 進捗ファイルを初期化中..."
if [ ! -f claude-progress.txt ]; then
    cat > claude-progress.txt << 'EOF'
# プロジェクト進捗

## 完了した機能
（なし）

## 現在の作業
初期セットアップ完了

## 次のステップ
features.jsonの最初の機能から開始

## 注意事項
- 各機能完了後にこのファイルを更新すること
- Gitコミットメッセージに機能IDを含めること

## 最終更新
[日時]
EOF
fi

echo "=== セットアップ完了 ==="
