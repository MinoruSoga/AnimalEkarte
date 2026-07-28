---
description: Go ビルドエラー・go vet・lint エラーの解決（Docker経由）
argument-hint: [--vet | --lint | blank for build only]
---

# /go-build

Go ビルドエラーを診断し `go-build-resolver` エージェントで最小差分修正します。

## 使用法

```bash
/go-build           # go build ./... のエラーのみ解決
/go-build --vet     # go build + go vet まで解決
/go-build --lint    # go build + go vet + golangci-lint まで解決
```

## Step 1: ビルドエラー収集

以下を **ユーザーに実行してもらい**、出力を貼り付けてもらう:

```bash
docker compose exec backend go build ./...
```

`$ARGUMENTS` が `--vet` または `--lint` の場合は追加で:

```bash
docker compose exec backend go vet ./...
# --lint の場合
docker compose exec backend golangci-lint run ./...
```

## Step 2: エラー解析

エラー出力が提供された場合、`go-build-resolver` エージェントを起動して解決する。

解決フロー:
1. エラーメッセージからファイル・行番号を特定
2. 該当ファイルを読んでコンテキスト理解
3. 最小差分で修正（アーキテクチャ変更禁止）
4. 修正後コマンドをユーザーに提示して再確認依頼

## 注意事項

- **Docker 必須**: `go build` はローカル実行禁止
- **最小差分**: リファクタリング・アーキテクチャ変更は行わない
- **apperrors パターン**: `apperrors.FromGORM()` / `apperrors.Wrap()` を維持
- **同一エラー3回継続**: 解決不能と判断し報告する

## 出力形式

```
[FIXED] backend/internal/handler/user.go:42
エラー: undefined: UserService
修正: import "project/internal/service" を追加

ビルドステータス: SUCCESS/FAILED | 修正エラー数: N | 変更ファイル: リスト
```

## 使用エージェント

`go-build-resolver` (Sonnet)
