#!/bin/sh
set -e

# swag CLIをインストール
go install github.com/swaggo/swag/cmd/swag@latest

# Swaggerドキュメントを生成（docsパッケージを先に作成）
swag init -g cmd/api/main.go -o docs

# 依存関係を解決
go mod tidy

# Airでホットリロード起動
exec air -c .air.toml
