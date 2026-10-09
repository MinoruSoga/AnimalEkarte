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


## 追加監査: 作業段階名と試験専用API、過去UATの発見境界

候補HEAD `e5b1cce805d19fe7ca8e4ad794606aa5f03afecc` の追跡済みbackend、frontend、scripts、docsを再調査した。frontendの静的監査は2,363パス、非テストコード1,374ファイル。backendでは他のunit/ticketフォルダと同内容Goファイル重複も確認した。共有main `d40a0b0d1525955ba26012a04415fee788f0652f` の他者コミットを取り込まず、task worktreeだけで変更する。

| 追加発見 | 対応 | 検証 |
|---|---|---|
| 認可テストの `permission_group_service_mutate_d2_test.go` / `permission_group_service_rules_d2_test.go` | 同じauth packageの `permission_group_service_update_with_rules_tx_test.go` / `permission_group_service_update_rules_tx_test.go` へ改名。auth設計の現行参照も更新 | 内容byte同一、Dockerで改名前後とも14ケースPASS・同じ集合。D2が入る過去の識別子は保持。実DB原子性の証明ではない |
| 継続して使うRadix reset設計がチケット名、frontend規則からの参照が移動前の不存在パス | `docs/architecture/radix-controlled-form-reset.md` へ内容同一で改名、索引とfrontend規則の相対リンクを更新 | 3つの改名ファイルはbyte同一、変更対象設計文書・規則のローカルリンク26件に欠落0 |
| `query-keys.ts` に `_test.fe516Fail/fe516Silent` が混在 | 専用キーは唯一のconsumer `react-query.test.ts` のreadonlyローカル定数へ移動。製品registryから削除 | Vitest改修前後2ケースPASS、同じキー値／ケース。ESLint・Prettier PASS |
| `clinic-selection-recovery.ts` に同義の `resetClinicSelectionRecoveryForTests()` | wrapperを削除し、4つの試験で既存 `clearClinicSelectionRecovery()` を直接使用 | 4ファイルのVitest改修前後53ケースPASS・同じ集合。既存clearの処理、abort・世代管理・spy観測区間は変更なし。ESLint・Prettier PASS |
| seed医院ID固定の過去UAT 2本が通常 `*.spec.ts` として発見される | `frontend/e2e/_archive/seeded-clinic-master-forms.archived.ts` / `seeded-clinic-uat-residuals.archived.ts` へ保全。相対importと停止注記だけ補正し、Playwrightも `_archive` を除外。現在の再実行案内を停止へ訂正 | Playwright `--list` のみ: 整理前31ケースを発見、整理後archiveは0（No tests foundの期待どおりexit 1）、通常V04は44ケース発見・exit 0。ブラウザ／DB／live要求0 |

過去UAT本文は停止注記と相対import以外は同一。`.archived.ts` によりPlaywright、page-consumerの列挙、agent検証の `rglob('*.spec.ts')` から自然に外れる。settings master pageには通常のsettings-crudとV04のconsumerが残る。既存 `--v04` のallowlist、fixture、workflowは変更しない。archiveはE2E用tsconfigの型検査対象からも外れるため、型検査済みとは主張しない。

ページconsumer契約23ケースPASS、実settings master pageのactive consumerも確認した。agent検証のE2E関連契約15ケースは限定実行でPASS。active spec列挙は40ファイル、archive 0。初回に当該Pythonファイル全81ケースを試した際は、Docker imageにgitが無くgit fixtureの2ケースがERRORとなった。今回の変更に関連するE2E契約へ絞って再確認し、未確認のgit fixture検証はPASSへ置き換えない。Prettier初回のquery-keys空行指摘は削除して再検証PASS。新規の攻撃ペイロード、live probe、DB変更、migration、E2E実行は行っていない。

改善プラン: 過去UATで必要なケースは使い捨てfixtureの通常specへ移植し、承認済み対象環境・合成アカウント・隔離医院・固定ID不使用・失敗時も完了するcleanupを確認してから別途実測する。archiveの拡張子復帰や独自configでの直接実行はしない。過去の24/24等の実測とFAILを保持し、今回の一覧確認をUAT成功と扱わない。命名・配置変更でC05/C09や他のUNKNOWN/BLOCKEDを引き上げない。

追加監査の候補20ファイル（改名5件を含む）は、別々の登録済みworktreeでGo・TypeScript・security reviewerが読み取り独立レビューし、全員APPROVE・actionable findings 0。JSONケース集合、改名ファイル、archive本文の保持を独自比較した。独立した試験再実行は行っていない。所有ファイルのGitleaksはexit 0 / no leaks、staged diff checkはPASS。
