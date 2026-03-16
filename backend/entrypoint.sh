#!/bin/sh
set -e

# 依存関係を解決
go mod tidy

# Airでホットリロード起動
exec air -c .air.toml
