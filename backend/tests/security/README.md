# セキュリティ境界の検証テスト

本番コードの公開APIを、合成マーカー・メモリ内のfake・`httptest`で検証する専用フォルダ。チケット番号や作業ユニットではなく、検証する境界ごとに配置する。同じ `backend/go.mod` 配下なので `backend/internal` の公開APIを利用できる。

| フォルダ | 検証内容 |
|---|---|
| `inputvalidation/` | JSON bindingの許容・拒否と比較用strict decoder |
| `browser/` | CSRF、CORS、セキュリティヘッダ |
| `uploads/` | ファイル形式検査、reader巻き戻し、保存名 |
| `middleware/` | レート制限、request ID、リクエストログ |
| `audit/` | 監査ログの不変条件とtransaction境界 |
| `reservation/` | 予約処理の順序と冪等性 |
| `medicalrecord/` | 診療記録の自動作成と冪等性 |

9つの試験ファイルと1つの比較helperはすべて `_test.go`。`inputvalidation/strict_bind_helper_test.go` はテスト専用であり、本番handlerの入力検証を変更しない。共有ファイルhandlerの通常の回帰試験は、非公開テストhelperを利用するため `internal/lstep/shared_file_content_test.go` に置く。

ネットワーク・DB・実ユーザーを使わず、Dockerでこのフォルダだけを検証する。下記はリポジトリルートで実行する（必要なimage・module cacheが既に存在すること）。

```sh
docker run --rm --pull never --network none -e GOPROXY=off \
  -v "$PWD/backend:/app:ro" \
  -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build \
  -w /app golang:1.26-alpine \
  go test -short -count=1 ./tests/security/...
```

`go test ./...` とローカルCIの明示shard一覧にも含まれる。package追加時は `scripts/ci/run-local-ci.sh` の一覧を更新し、`go list ./...` と一致させる。CI全体の起動はDB操作を伴うため、この安全な限定検証手順とは分ける。

メモリ内での合格は、実DB競合・複数インスタンス・ステージング・通知配送・本番受入の証明にならない。ケース判定と残件は [EMR-332の記録](../../../docs/work/emr/emr-332-security/index.md)、移動前のパスとの対応は [配置整理記録](../../../docs/work/emr/emr-332-security/test-layout-20261009.md) を参照。
