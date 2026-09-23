# DEV-V-OWNER-DB — 飼主DB5ケース disposable DB 実行 receipt

Execution date: 2026-09-23 (JST). Ticket: Plane `EMR-125` / unit `DEV-V-OWNER-DB`.
分類表・実行案の正本: [DEV-V-OWNER-DB.md](DEV-V-OWNER-DB.md)（本票はその §3 実行案の実行証跡）。

## 0. Result サマリ

| # | テスト完全名 | 結果 | 時間 |
|---|---|---|---|
| 1 | `TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate` | **PASS** | 0.05s |
| 2 | `TestOwnerRepository_Update_ClinicIsolation` | **PASS**（subtest 3/3 PASS） | 0.07s |
| 3 | `TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected` | **PASS** | 0.70s |
| 4 | `TestOwnerService_Update_DiscountTOCTOU_NonDiscountFieldStillOK` | **PASS** | 0.03s |
| 5 | `TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction` | **PASS** | 0.01s |

`go test` package result: `ok github.com/animal-ekarte/backend/internal/owner 0.875s`、exit 0。
`--- PASS` 行 5/5 確認。SKIP なし・FAIL なし・未出現なし。
除外ケース `TestOwnerService_Update_DiscountTOCTOU_LockedDiffWithoutPermission`（mock 経路・実DB不使用）は `-run` 列挙に含めず、実行出力にも未出現であることを確認済み。

## 1. revision（実行時点の識別）

| 欄 | 値 |
|---|---|
| worktree | `/Users/minoru/orca/workspaces/AnimalEkarte/emr-125`（branch `MinoruSoga/emr-125`） |
| worktree HEAD | `923bb99babd51a91ee031217113bbb1357d70061` |
| worktree status | clean（実行前 `git status --porcelain` 出力なし） |
| test 実行 image | `golang:1.25-alpine`（image ID `86d0141523af`、ローカル既存固定 image） |
| disposable DB image | `postgres:18-alpine`（compose `db` service と同 image 系列） |
| mount | 上記 worktree の `backend/` → `/app:ro`（read-only）、`ekarte-go-mod-cache` → `/go/pkg/mod:ro`、`--tmpfs /root/.cache/go-build` |
| 作業 dir | `/app`（container 内） |

## 2. 専用 disposable DB と承認参照

| 欄 | 内容 |
|---|---|
| 承認参照 | Orca dispatch `task_4e920808dc13` / dispatch context `ctx_ef5d2ae0bcbb`（EMR-125 対応の作業者 dispatch = operator 実行権限）。disposable DB の作成・破棄・対象は本 worker が本 run 内で実施。データ消失影響: なし（本検証専用に新規作成した使い捨て instance、実行後に破棄） |
| disposable DB 実体 | container `emr125-ownerdb`（`postgres:18-alpine`、created `2026-09-23T07:28:53Z`）、network `emr125-ownerdb-net` 上の `192.168.155.2` |
| 隔離 | `emr125-ownerdb-net` は `internal=true` の専用 network で、attach された container は `emr125-ownerdb` のみ（test 実行中は `go test` 用一時 container が追加 attach）。**共有 DB（`animalekarte-db-1`、network `ekarte-network` / `192.168.97.2`）への経路は test container に存在しない**。共有 DB への fallback なし |
| `TEST_DATABASE_URL` | 実行時に生成した使い捨て credential を含む DSN を環境変数として test container にのみ注入（値は本票・ログに記載しない。`postgres://emr125_verify:<redacted>@emr125-ownerdb:5432/ekarte_db_test?sslmode=disable`）。`DB_*` 系も同一 disposable instance を指すため、`testdb.connectTestDatabase` が行う `CREATE DATABASE ekarte_db_test` も disposable instance 上で実行された |
| `_test` suffix | test DB 名 `ekarte_db_test` は `_test` 接尾辞を満たす（`truncate.go` の非 `_test` 拒否 guard に適合） |

## 3. 接続確認（実行者による非共有 DB 確認の記録）

実行時に取得した実測値（実値パスワードは除く）:

- test container 内 `getent hosts emr125-ownerdb` → `192.168.155.2 emr125-ownerdb`（専用 container の IP のみ解決）
- `docker network inspect emr125-ownerdb-net` → `internal=true`、containers=`emr125-ownerdb(192.168.155.2/24)` のみ
- `docker inspect animalekarte-db-1` → network `ekarte-network` / `192.168.97.2`（別 container・別 subnet・別 network。共有 DB server とは別実体であることを確認）
- disposable server 上の DB 一覧（実行後）: `ekarte_db`, `ekarte_db_test`, `postgres` — `ekarte_db_test` は本実行で disposable instance 内に作成された
- test double schema 着陸確認: `ekarte_db_test` の public schema に table 42 件、`pg_type` enum 54 件（`testdb.SharedTestSchemaEnumTypes` の 54 型と一致）、`owners` 行あり — test 負荷が disposable instance に実際に到達したことを確認
- `-short` 未指定、`GOPROXY=off`（内部 network のみ、外部 fetch なし）で実行

## 4. 実行コマンド（再現用・秘密値は除去済み）

```bash
# disposable DB（内部 network のみ、port 非公開）
docker network create --internal emr125-ownerdb-net
docker run -d --name emr125-ownerdb --network emr125-ownerdb-net \
  --security-opt no-new-privileges \
  -e POSTGRES_USER=emr125_verify -e POSTGRES_PASSWORD=<generated> \
  -e POSTGRES_DB=ekarte_db -e TZ=Asia/Tokyo \
  postgres:18-alpine -c timezone=Asia/Tokyo

# 対象 worktree を mount した一時 container で5ケース列挙実行
docker run --rm --network emr125-ownerdb-net --security-opt no-new-privileges \
  -v <worktree>/backend:/app:ro \
  -v ekarte-go-mod-cache:/go/pkg/mod:ro \
  --tmpfs /root/.cache/go-build -w /app \
  -e DB_HOST=emr125-ownerdb -e DB_PORT=5432 \
  -e DB_USER=emr125_verify -e DB_PASSWORD=<generated> -e DB_NAME=ekarte_db \
  -e TEST_DATABASE_URL='postgres://emr125_verify:<generated>@emr125-ownerdb:5432/ekarte_db_test?sslmode=disable' \
  -e GOPROXY=off -e GOFLAGS=-mod=mod -e GOMODCACHE=/go/pkg/mod -e TZ=Asia/Tokyo \
  golang:1.25-alpine \
  go test ./internal/owner -count=1 -v -run '^(TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate|TestOwnerRepository_Update_ClinicIsolation|TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected|TestOwnerService_Update_DiscountTOCTOU_NonDiscountFieldStillOK|TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction)$'
```

## 5. cleanup 結果

| 項目 | 結果 |
|---|---|
| disposable container `emr125-ownerdb` | `docker rm -f` 済み（volume 未使用のため PGDATA も container と共に消滅。`docker ps -a --filter name=emr125` 空を確認） |
| 専用 network `emr125-ownerdb-net` | `docker network rm` 済み（`docker network ls --filter name=emr125` 空を確認） |
| 一時 secret | 実行中のみ shell 環境変数と `/tmp/.emr125_pw`(chmod 600) に保持し、実行後 `rm -f` 済み（不存在を確認）。receipt・ログ・worktree への実値記載なし |
| per-test TRUNCATE | `testdb.SetupTestDB` 内の core 表 TRUNCATE と owner helper の `insurances`/`animal_species` TRUNCATE に委譲済み（全て disposable instance 内） |
| 共有 DB | `animalekarte-db-1` は実行前後を通じて running・再起動なし（started `2026-09-22T18:35:59Z` のまま）。共有 DB への接続・書込み・TRUNCATE は一切行っていない |

## 6. 判定

**全5ケース PASS（実DB・専用 disposable instance 上で実行済み）**。コード変更・migration 適用・共有 DB 接続なし。
未完了項: なし。
