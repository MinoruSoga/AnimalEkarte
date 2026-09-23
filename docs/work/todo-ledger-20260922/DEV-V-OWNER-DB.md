# DEV-V-OWNER-DB — owner 実DBテスト5ケース 分類表・実行案

Campaign `todo-ledger-20260922` revision 1. Unit `PREP-OWNER-DB`. Attempt `att-ownerdb-20260922-001`. Claim `claim/PREP-OWNER-DB`. Prompt SHA-256 `e310635f05cdf6b1f18ec7b01fcb551e39e0ad7249e506252e92f1e38b54229e`. Acceptance SHA-256 `eb29c3e114aa906e8e1eb8592f00131b11412904ef1d601f5ebbbe48431c4d1d`.

Sheet date: 2026-09-22. Worktree HEAD `1a2a4fe4e6ac5a401a637172a365b2d081393c17` (branch `claim/PREP-OWNER-DB`)。本票は**過去 receipt と現行ソースの読取照合のみ**。実DB実行・`TEST_DATABASE_URL` 実値の取得・Docker での `go test` 実行・共有DB接続は本 unit の範囲外であり、本票はそれらの**実行案**を定める。分類は過去 receipt の実在証跡のみに基づき、証跡なしは未実行とする。未実行・SKIP を PASS にしない・パッケージの exit 0 を充足にしない。

## 0. Verdict サマリ

| 分類 | 件数 | 内訳 |
|---|---|---|
| PASS | 0 | 過去 receipt に実行成功の証跡なし |
| FAIL | 0 | 過去 receipt に実行失敗の証跡なし |
| SKIP | 0 | 過去 receipt に SKIP 記録の証跡なし |
| 未実行 | 5 | 全ケース。`remaining-local-reconciliation-20260917` で DEV-V-OWNER-DB は `status: blocked` / `attempts: []` のままクローズされており、以降の `docs/work/`・`.planning/`・`reports/` に `./internal/owner` の `go test` 実行証跡が存在しない |

**結論**: 5ケース全て **未実行**。前回(2026-09-17 campaign)の blocker は「Disposable `TEST_DATABASE_URL` not confirmed」で、現時点でも専用 disposable DB の承認参照・実行証跡は未照合のまま(todo-verification.md L127 `UNKNOWN(追加証拠未照合)`)。§3 の実行案の外部承認が揃うまで実DB実行には進まない。

## 1. 対象5ケース(完全名・現行ソース実在確認)

| # | テスト完全名 | 定義箇所(file:line) | DB 経路 | 実DB前提 |
|---|---|---|---|---|
| 1 | `TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate` | `backend/internal/owner/repository_update_atomicity_test.go:15` | `setupOwnerPetIsolationTestDB` → `testdb.SetupTestDB` | 要(更新+reload の rollback 原子性。GORM callback で強制 reload 失敗) |
| 2 | `TestOwnerRepository_Update_ClinicIsolation` | `backend/internal/owner/repository_clinic_isolation_test.go:79` | `setupOwnerPetIsolationTestDB` → `testdb.SetupTestDB` | 要(clinic 越境 Update が NotFound・行不変・正規 clinic 成功) |
| 3 | `TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected` | `backend/internal/owner/owner_discount_toctou_test.go:126` | `setupOwnerDiscountTOCTOUDB` → `testdb.SetupTestDB` | 要(2 goroutine・`FOR UPDATE` ロック直列化・stale 割引0拒否。ctx timeout 8s) |
| 4 | `TestOwnerService_Update_DiscountTOCTOU_NonDiscountFieldStillOK` | `backend/internal/owner/owner_discount_toctou_test.go:197` | `setupOwnerDiscountTOCTOUDB` → `testdb.SetupTestDB` | 要(非割引フィールド更新は権限なしでも成功し割引値を保持) |
| 5 | `TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction` | `backend/internal/owner/owner_discount_toctou_test.go:238` | `setupOwnerDiscountTOCTOUDB` → `testdb.SetupTestDB` | 要(ambient tx なしの LockByIDForUpdate がエラー「ambient transaction」を返す) |

除外: `TestOwnerService_Update_DiscountTOCTOU_LockedDiffWithoutPermission`(`owner_discount_toctou_test.go:216`)は **mock repository 経路で実DBを使わない**ため対象外(todo-verification.md L134 の `LockedDiffWithoutPermission を除く` と一致)。

## 2. 既存結果の分類(過去 receipt 照合)

| # | テスト完全名 | 分類 | 根拠 receipt / 証跡 |
|---|---|---|---|
| 1 | `TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate` | **未実行** | 証跡なし。`.planning/agent-fast-campaign/remaining-local-reconciliation-20260917/units.json`(DEV-V-OWNER-DB, `status:"blocked"`, `attempts:[]`)、同 `INVENTORY.md:41`(external_prerequisite, "Disposable TEST_DATABASE_URL not confirmed")、todo-verification.md L127(UNKNOWN) |
| 2 | `TestOwnerRepository_Update_ClinicIsolation` | **未実行** | 証跡なし。同上 |
| 3 | `TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected` | **未実行** | 証跡なし。同上。テスト定義は `57e3d4f55`(fix: SEC-CS-F15 recheck owner discount under FOR UPDATE lock)で導入されたが実行 receipt は存在しない |
| 4 | `TestOwnerService_Update_DiscountTOCTOU_NonDiscountFieldStillOK` | **未実行** | 証跡なし。同上(同コミット由来) |
| 5 | `TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction` | **未実行** | 証跡なし。同上(同コミット由来) |

### 照合した証跡の範囲

- `rg -l "DEV-V-OWNER-DB|TOCTOU|LockByIDForUpdate|ClinicIsolation|ReloadFailureRollsBackUpdate" docs/work/` → `todo-campaign-20260918/UAT-R2-EXCLUSIVE-LOCK.md`(治療 TOCTOU の命名規約のみ、owner 実行 receipt ではない)、`todo-campaign-20260919-ready17/SLACK-VACCINE-MULTI.md`・`SLACK-CAMERA.md`(medicalrecord のロック参照、無関係)。
- `rg "go test.*internal/owner|TestOwner"` over `docs/`・`.planning/`・`reports/` → 実行 receipt 0件。
- `docs/work/docs-perfection/EVIDENCE/astra-loop-baseline.json` L1701/L1706/L1715 → 3テストファイルの hash baseline(**実在**の証跡であり実行結果ではない)。
- `.planning/agent-fast-campaign/remaining-local-reconciliation-20260917/units.json` DEV-V-OWNER-DB → `unblock.kind:"external_prerequisite"`, `detail:"Disposable TEST_DATABASE_URL not confirmed"`, `operator_action:"Operator supplies approved disposable DB and runs scoped owner cases"`, `attempts:[]`(**実行ゼロ**の証跡)。
- todo-verification.md L127 → `UNKNOWN(追加証拠未照合)`「前回は disposable DB URL 未設定」。

## 3. 実行案(専用 disposable DB・共有DB非使用)

### 3.1 前提条件(外部承認・実行前に揃える)

| 条件 | 内容 | 根拠 |
|---|---|---|
| 専用 disposable DB | 共有 DB ではなく本検証専用に provisioned された使い捨て PostgreSQL(専用 container/専用 instance)。**共有 DB への fallback は禁止** | todo-verification.md L31「共有 DB は不可」・L142、remaining-local-reconciliation-20260917 INVENTORY.md:41 |
| `TEST_DATABASE_URL` の秘密管理供給 | 接続文字列は秘密管理(保護された環境変数/シークレットストア)から供給し、**実値・接続文字列・パスワードを本票・receipt・ログに書かない** | todo-verification.md L140「秘密管理から供給」、本 prompt Constraints |
| 実行者確認 | 実行者が接続先の実体(`SELECT current_database()`・host/instance 識別)が**共有 DB でないこと**を実行時に確認し receipt に記録する。testdb は DB 名 `_test` 接尾辞を強制するが(`truncate.go` L29-31・L90-92 非 `_test` DB への TRUNCATE/terminate を fatal 拒否)、同名の `*_test` DB が共有 server 上に存在しうるため、suffix ガードは disposable 性の証明にならない | todo-verification.md L140「接続先の実体が共有 DB でないことを実行者が確認」 |
| 専用 DB 承認参照 | disposable DB の作成・破棄と対象・データ消失影響を示した承認記録を成果物に引用する | scoped-verification-gates §6(承認済み disposable 環境のみ)、todo-verification.md L142「専用 DB の承認参照」 |

### 3.2 実行形態(候補 mount・Docker)

- 対象 worktree(`claim/PREP-OWNER-DB`・実行時点 HEAD)を container へ mount し、`backend/` を作業ディレクトリとして `go test` を実行する。`docs/ops/agent-harness.md` の隔離条件(network 制御・読み取り専用 mount・capability なし一時 container)に従い、**到達可能なネットワーク先は専用 disposable DB のみ**とする(共有 DB へ到達可能な経路を残さない)。
- image・mount・対象 diff・コマンド・終了コード・テスト件数を証跡に結ぶ(agent-harness.md L40)。
- DB 依存のため **`-short` を付けない**(`testdb.SetupTestDB` は `-short` で `t.Skip` → 全件 SKIP になる)。
- testdb 接続系の実装前提(`backend/internal/testdb/testdb.go` L526-591): `TEST_DATABASE_URL` が非空なら test DSN をその値で上書き(L576-578)。未設定時は `DB_HOST`/`DB_PORT`/`DB_USER`/`DB_PASSWORD`/`DB_NAME` から `<DB_NAME>_test` を組み立て、main 接続で `CREATE DATABASE <name>_test`(advisory lock 付き・L593-617)を試みる。disposable 供給は `TEST_DATABASE_URL` を使い、DB 名が `_test` 接尾辞を満たすことを確認する。
- schema は GORM AutoMigrate による test double であり `001_init.sql` を適用しない(testdb.go L39-41)。本5ケースはこの schema 上の契約(rollback 原子性・clinic 隔離・FOR UPDATE 直列化・ambient tx 必須)を検証する。
- テスト間分離は `SetupTestDB` 内の `TRUNCATE ... RESTART IDENTITY CASCADE`(core 表: billing_refunds/payments/billings/medical_records/owners、`truncate.go` L23)と owner helper の `insurances`/`animal_species` TRUNCATE(`repository_test_helpers_test.go` L24-36)による。**破壊的 cleanup が前提**であることが共有 DB 禁止の実装上の根拠。

### 3.3 実行コマンド(5件完全名列挙)

```bash
cd backend
go test ./internal/owner -count=1 -v -run '^(TestOwnerRepository_UpdateAndFind_ReloadFailureRollsBackUpdate|TestOwnerRepository_Update_ClinicIsolation|TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected|TestOwnerService_Update_DiscountTOCTOU_NonDiscountFieldStillOK|TestOwnerRepository_LockByIDForUpdate_RequiresAmbientTransaction)$'
```

- `-run` は上記5件の完全名を `|` 列挙した anchored 正規表現。**`TestOwnerService_Update_DiscountTOCTOU_LockedDiffWithoutPermission` は含めない**(mock 経路・対象外)。`*` や前方一致でTOCTOU群を束ねない(LockedDiffWithoutPermission を巻き込むため)。
- `-v` で各ケースの `--- PASS` / `--- FAIL` / `--- SKIP` 行を採取し、**パッケージの exit 0 を個別ケースの充足にしない**。ケースが SKIP/未出現なら全体 exit 0 でも未完了(todo-verification.md L142)。
- `-count=1` で Go test cache を無効化する。
- `TestOwnerService_Update_DiscountTOCTOU_StaleZeroRejected` は goroutine 間の `FOR UPDATE` 直列化を検証する実並行テスト(ctx timeout 8s)であり、実DB接続とロック取得が前提。

### 3.4 cleanup

- 各テスト内のデータ分離: `testdb` の per-test TRUNCATE(上記)に任せる。
- 実行後の cleanup: 専用 disposable DB の **DROP/廃棄**(container/instance 削除)を行い、cleanup 実施結果を receipt に記録する。`TEST_DATABASE_URL` 等の一時秘密は環境から除去する。
- cleanup 未完了も receipt に明記する(未完を隠さない)。

### 3.5 成果物要件(実行 receipt に必須の欄)

| 欄 | 内容 |
|---|---|
| revision | 実行時点の worktree HEAD(`claim/PREP-OWNER-DB` 上の commit SHA)と image/mount 識別 |
| 専用 DB 承認参照 | disposable DB 作成・破棄の承認記録への参照(承認者・対象・消失影響) |
| ケース別結果 | 5件完全名それぞれの 実行/PASS/FAIL/SKIP(`-v` の `---` 行を証拠化) |
| cleanup 結果 | disposable DB の廃棄・秘密除去の実施有無と証跡 |
| 接続確認 | 実行者による「接続先が共有 DB でない」確認の記録(実値は伏せる) |

### 3.6 実DB実行へ進むために必要な外部承認一覧

1. 専用 disposable DB の provisioning 承認(作成・破棄・対象・消失影響を明示したもの)。
2. `TEST_DATABASE_URL` の秘密管理からの供給(実値は本票・成果物に記載しない)。
3. 実行者(指名 operator)による接続先非共有の確認と、Docker による対象 worktree mount 実行の承認。
4. 上記未充足の間は DEV-V-OWNER-DB は実行不可 — 本 unit は分類表・実行案の作成まで。

## 4. 範囲外メモ

- 本票の作成にあたり実DB実行・`TEST_DATABASE_URL` 実値取得・Docker `go test`・共有DB接続・コード変更・migration 適用は一切行っていない。
- `backend/internal/owner/**`・`backend/internal/testdb/**` は読み取り専用で参照のみ。
- 未実行・SKIP を PASS にしない。パッケージ exit 0 は個別ケース充足の証拠にならない。
