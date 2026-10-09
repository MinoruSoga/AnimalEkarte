# 検証コード・関連文書の配置整理（2026-10-09）

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

`golangci-lint run ./tests/security/... --max-same-issues 0 --max-issues-per-linter 0` は exit 0 / 0 issues（既存 `golangci/golangci-lint:v2.11.4` image、network none、backend read-only、`GOPROXY=off`）。初回のgoimports 1件はimportグループ整形で修正した。Gitleaksの所有ファイル検査はexit 0 / no leaks。別々の登録済みworktreeでGo/security reviewerが読み取り独立レビューし、両者APPROVE・actionable findings 0。10ファイルのコード/import集合、57/169試験、66package CI登録を独自に突合した。コンテナ検証はrootの実測ログ確認であり、reviewerによる試験の独立再実行は行っていない。先行のテスト移動は配置整理であり、既存の防御／検出判定や未検証範囲を引き上げない。ステージング実測、実DB、複数インスタンス、通知配送、C14脆弱性状態は従来のUNKNOWN/BLOCKEDのまま。


## 後続の全体配置調査と修復

調査基準は、チケット／agent unitだけを配置名にすること、製品コードと検証専用補助の役割が名前から分からないこと、移動後の参照・CI登録が追随していないこと。調査対象はGit追跡済みのbackend、frontend、scripts、docs。main `e2b1a40386a4508c53a4b346218f323365ac80ed` と修復前task HEAD `e8071f175a96ab1e7c3388b0cf3e36ff4645cc5c` を調べ、fetch後もorigin/mainが同じであることを確認した。共有mainの他者WIPは保全し、このtask worktreeでだけ変更した。

docsの33ファイルを内容別に移動し、作業unit番号のフォルダとrootファイル名を除去した。root `docs/work/emr-332-security/` はチケットの補助証跡を保つ既存のdocs/work用途に従って維持する。C01–C18は判定matrixのstable case ID、Unit/Attempt/Receiverは初回作業の履歴識別子として維持する。

| 現行フォルダ | 内容 |
|---|---|
| `baseline/` | 安全境界、route/middleware map、テスト分類、判定template、source inventory |
| `identity-access/` | 認証、token、認可、clinic scope、fail-closed。生ログは `evidence/` |
| `input-output/` | 入力検証、SQL棚卸し、XSS/render |
| `browser/` | CSRF、transport headers、CORS、cache。生ログは `evidence/` |
| `egress/` | 通信先／SSRF検証 |
| `uploads/` | アップロード検証 |
| `dependencies/` | 依存ライブラリ／供給網 |
| `workflow-audit/` | 業務順序／冪等性、rate limit、検出／監査 |
| `assessment/` | 初回再確認と残件 |

現行文書間参照と7つのGo試験ソースの参照コメントを追随させ、docs/workの入口にも登録した。生ログ9ファイルは移動前Git blobのSHA-256と一致し、Markdown内の実行出力5ブロックも同一。文書リンク21件を確認し、欠落0。過去コマンドや出力に現れる旧package名は履歴として保持し、現行手順とは区別する。

| 追加発見 | 対応 | 限定検証 |
|---|---|---|
| `OwnerForm.bug373.test.tsx` | 同じroutes配下の `OwnerForm.pet-owner-change.test.tsx` へ改名。機能名のコメント／describeに変更し、BUG-373は経緯注記に保持 | Vitest改名前後8件PASS・同じcase集合。ESLint／Prettier PASS。過去snapshot JSONの旧pathは書き換えない |
| `billing/list_query_helpers_test_free.go` | 実requestで使う製品コードなので `list_query_helpers.go` へ改名 | ファイル内容byte同一、cmd/api build PASS |
| `cmd/auth-d5-dualprocess` のscript呼出先不整合 | 実在する `scripts/tools/auth-d5-dualprocess.sh` を参照。呼出先計算を小関数へ抽出 | 新規path contractは旧pathでENOENTのRED、修正後GREEN。既存DSN許容／拒否6ケースを内容変更なしで維持し全PASS。main／script／DB harnessは未実行 |
| `infra/lstep/client.go` 内の共有テストconstructor | `client_test_support.go` へ分離。関数・export・runtime guardは同一 | 既存consumerとhardened dialの24関数／37 PASS eventsが抽出前後で一致。build／vet／lint PASS |

LSTEPの共有補助は別packageの `_test.go` からimportされるため、infra側の `_test.go` へ移すとconsumerから見えなくなる。非公開client型の別support package移設には新しい公開seamが必要になる。今回は専用ファイルへの分離に留め、通常 `.go` としてコンパイルされる理由を明記した。本番factoryは `NewClient` とhardened transportを保持し、補助の非テスト時panic guardはそのまま。このファイル分離を「本番compileから除外」とは主張しない。

通常の隣接回帰試験、共有 `testdb`／`lintscan`、cmd-only `clinicale2e`、実consumerと安全guardがあるsynthetic／rehearsal、業務割引のcampaignは用途を確認し配置変更対象にしなかった。EMR-332以外の作業unit名フォルダは確認されなかった。

後続検証はnetwork none、取得済みimage/moduleのみ、candidate source read-onlyで行った。frontendは既存node_modules volumeをread-onlyにし、Vitestの `--configLoader runner --cache false` を使用。Go CLIの試験ではbackendとscriptsを同じ `/repo` 配下へread-only mountし、scriptの実在・実行permissionを調べるだけで実行しなかった。

- Go: CLI限定試験、cmd/apiとlauncher build、変更scopeのvet／lint（各0 issues）、gofmt。
- LSTEP: `go test -short -p 1 -count=1 -json ./internal/lstep ./internal/infra/lstep -run '^Test(CheckupSyncService_CreateCheckupSync_|GetOwnerTags_Client|IsForbiddenDialIP$|HardenedDialContext_)'`。
- frontend: `vitest run src/features/owners/routes/OwnerForm.pet-owner-change.test.tsx --configLoader runner --cache false`、当該ファイルのESLint／Prettier。
- CI package一覧は66件で既存shard表と完全一致、重複0。新package追加／CI除外の追加なし。

Go、TypeScript、security reviewerがそれぞれ専用worktreeで読み取り独立レビューし、全体APPROVE。Critical/High/Mediumなし、security/docのLow2件（入口の存在しないrange参照・抽出helperの旧引用）は修正しreadback確認済み、残指摘0。提供ログを確認し、コード／既存試験／raw証跡の同一性を独自に突合した。reviewerによる試験の独立再実行なし。所有ファイルのGitleaksはexit 0 / no leaks。frontendの型・import・試験本体は同一であり、今回typecheck結果は提示していない。full CI、実DB、staging、通知配送、frontend E2Eは未実行で、元のセキュリティ残件と受入状態は変わらない。ロールバックが必要なら通常のレビュー付きGit revertを使い、データ・監査ログの削除は行わない。
