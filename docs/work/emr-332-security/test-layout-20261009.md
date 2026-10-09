# 検証専用テストの配置整理（2026-10-09）

対象は Draft PR #617 の `feat/emr-332-security-defense`。移動前HEADは `7b48be17a8acb6293c737267a15aa92fa0812362`。mainへの統合・デプロイ・ステージング実測は行っていない。

作業チケット／ユニット名で分けた `backend/internal/**/emr332_u*/` を削除し、同じGo module内の `backend/tests/security/` に内容別で配置した。9つの試験ファイルの試験内容は維持し、package名・packageコメント・1ファイルのimport整形だけを調整した。未使用の `strict_bind.go` は `strict_bind_helper_test.go` に変更し、比較検証時だけコンパイルされる。本番handlerへの配線はなく、C05の修復を意味しない。

| 移動前（`backend/`からの相対パス） | 現行フォルダ | ファイル |
|---|---|---|
| `internal/middleware/emr332_u03/` | `tests/security/inputvalidation/` | `bind_boundary_test.go`, `strict_bind_test.go`, `strict_bind_helper_test.go`（旧 `strict_bind.go`） |
| `internal/middleware/emr332_u04/` | `tests/security/browser/` | `browser_edge_test.go` |
| `internal/imageupload/emr332_u05/` | `tests/security/uploads/` | `sniff_contract_test.go` |
| `internal/middleware/emr332_u06/` | `tests/security/middleware/` | `rate_limit_test.go`, `logging_requestid_test.go` |
| `internal/audit/emr332_u06/` | `tests/security/audit/` | `audit_validate_test.go` |
| `internal/reservation/emr332_u06/` | `tests/security/reservation/` | `order_idempotency_test.go` |
| `internal/medicalrecord/emr332_u06/` | `tests/security/medicalrecord/` | `auto_create_idempotency_test.go` |

共有ファイルhandlerの通常の回帰試験 `backend/internal/lstep/shared_file_content_test.go` は、同packageの非公開test helperを使うためそのまま配置する。製品コードの修復も変更しない。

CIは `scripts/ci/run-local-ci.sh` の既存shard表に7 packageを登録した。Dockerの `go list ./...` とshard一覧は66 packageで完全一致し、重複割当0。`go test ./...` と明示shard経路の両方に含まれる。実DBを再作成するfull local CIは未実行。

現在のソース参照を更新し、初回campaign・修復時の実行コマンド／verbatimログは書き換えていない。旧パスは当時のrevisionの証跡として本表でたどれる。

## 限定検証

ネットワークなし・module取得なし・backend read-only mountで実行。DB、実ユーザー、個人情報、永続化、負荷試験は使わない。テストにはメモリ内fakeまたはhttptestを用いる。

共通コンテナ条件（リポジトリルートで実行）:

```sh
docker run --rm --pull never --network none -e GOPROXY=off \
  -v "$PWD/backend:/app:ro" \
  -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build \
  -w /app golang:1.26-alpine \
  go test -short -count=1 -json ./tests/security/...
```

| 確認 | 結果 |
|---|---|
| 移動前7 packageの `go test -short -count=1 -json` | exit 0、57 test functions、169 pass events（subtest含む）、FAIL/SKIP 0 |
| 現行 `go test -short -count=1 -json ./tests/security/...` | exit 0、同じ57関数／169 pass events。旧→現行package対応で正規化したtest名の多重集合が完全一致 |
| 10 Goファイルの移動前後比較 | package・コメント・import整形を除くコードとimport集合が同一 |
| `go test -race -short -count=1 ./tests/security/...` | 7 package PASS、exit 0。既存 `golangci/golangci-lint:v2.11.4` image、Go 1.26.1、`CGO_ENABLED=1`、gccで実行 |
| race初回試行（Alpine Go image） | exit 2: cgoなしでrace非対応。上記の既存compiler入りimageで再検証してPASS |
| `go build -o /tmp/emr332-api ./cmd/api` | exit 0、使われないhelperを `_test.go` に移しても製品build成立 |
| `go vet ./tests/security/...` | exit 0 |
| `gofmt -l tests/security` | 出力なし |
| `bash -n scripts/ci/run-local-ci.sh` | exit 0 |
| `go list ./...` とshard一覧の集合比較 | 66 package完全一致、重複0 |

`golangci-lint run ./tests/security/... --max-same-issues 0 --max-issues-per-linter 0` は exit 0 / 0 issues（既存 `golangci/golangci-lint:v2.11.4` image、network none、backend read-only、`GOPROXY=off`）。初回のgoimports 1件はimportグループ整形で修正した。Gitleaksの所有ファイル検査はexit 0 / no leaks。別々の登録済みworktreeでGo/security reviewerが読み取り独立レビューし、両者APPROVE・actionable findings 0。10ファイルのコード/import集合、57/169試験、66package CI登録を独自に突合した。コンテナ検証はrootの実測ログ確認であり、reviewerによる試験の独立再実行は行っていない。今回の変更は配置整理であり、既存の防御／検出判定や未検証範囲を引き上げない。ステージング実測、実DB、複数インスタンス、通知配送、C14脆弱性状態は従来のUNKNOWN/BLOCKEDのまま。
