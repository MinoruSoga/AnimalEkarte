# 医院 CSV カットオーバー投入（F6）

更新日: 2026-07-24

`old_db` が出力した AnimalEkarte 形状 CSV 21 テーブルを、AnimalEkarte DB へ投入する正式な consumer 手順です。`old_db` DB へは接続せず、医院・run 固定の `manifest.json` と CSV ディレクトリだけを読みます。

現行 KNJO source は未完全なため、`payments.csv` / `payment_splits.csv` は意図的にheader-onlyです。CSVの形状確認には使えますが、`status=completed` のbillingにpayment graphがないbundleはpreflightで拒否され、正式applyには使用できません。producerは正式bundleを再生成する前に、`billings.csv`へ`completed_at`を追加し、completed billingごとのpayment graphを出力する必要があります。

## 安全境界

- source は絶対パスを `/migration-input:ro` で read-only mount する。
- one-shot containerへ渡すsecretはDB接続用環境変数だけに限定し、`.env.local` 全体をcontainerへ注入しない。
- `clinic-migration-run-report.json` に記録された manifest SHA-256 を別経路で受領し、source directory 内から自己申告値を拾わない。
- manifest は `PASS`、`animalekarte_stage`、医院 code/ordinal、run ID、10M ID band、21 テーブル固定順、各 CSV SHA-256 が完全一致する場合だけ受理する。
- source directory は 0700、manifest/CSV は 0600、symlink は不可。
- parserのメモリ/時間上限としてmanifestは4MiB、各CSVは512MiB、payment親は100万件、splitは親あたり最大2件を上限とし、超過時はfail-closedする。
- target seed ID は6つとも明示する。予約種別・支払方法を表示名や先頭行から暗黙解決しない。
- apply は対象 band が21表すべて空の場合だけ実行し、既存行を削除・置換しない。
- apply は単一 transaction、advisory lock、table lockを使う。CSV が preflight 後に変わった場合は再 SHA-256 検証で全 rollback する。
- CSV/manifest は PHI を含み得る。行値をログ・report・Git・チャットへ出さない。report は aggregate count と再検証に必要な非PHIの6 seed IDのみを記録し、固定mount `/migration-reports` 直下へ0600/no-clobberで作成する。

## 事前準備

1. target DB の検証済み full backup を取得し、復元手順と担当者を確定する。
2. 旧増分002〜009を統合した現行 `001_init.sql` で target DB を再構築済みであることを確認する。統合前001が適用済みのDBはchecksum mismatchになるため `DB_RESET=true` 相当の承認済み再構築が必須で、手書きSQLによる差分適用は使わない。001内の通常の `CREATE INDEX` は対象テーブルへの書き込みを待たせ得るため、事前リハーサルで所要時間を測り、maintenance window内で適用する。
3. target DB を既存の運用経路で起動・疎通確認する。CSV Make targets は `--no-deps` で実行し、target container/service を作成・再作成しない。
4. 次の target seed ID を対象医院で確認する。
   - active clinic
   - active system-wide animal species fallback
   - 対象 clinic の active/non-deleted `exam_types.name='検査'`
   - 対象 clinic の active/non-deleted `reservation_types.category='trimming'`
   - 対象 clinic の active/non-deleted `payment_methods.system_key='cash'`（同一clinic/keyで1件）
   - 対象 clinic の active/non-deleted `payment_methods.system_key='credit_card'`（同一clinic/keyで1件）
5. producer run report と bundle を照合し、manifest SHA-256 を安全な作業票へ転記する。

共通変数の例:

```sh
export CSV_IMPORT_SOURCE_DIR=/absolute/path/to/animalekarte-csv-export/<clinic>/<run>
export CSV_MANIFEST_SHA256=<64-hex>
export CLINIC_CODE=<clinic-slug>
export CLINIC_ORDINAL=<1..50>
export MIGRATION_RUN_ID=<run-id>
export TARGET_CLINIC_ID=<id>
export FALLBACK_ANIMAL_SPECIES_ID=<id>
export FALLBACK_EXAM_TYPE_ID=<id>
export TRIMMING_RESERVATION_TYPE_ID=<id>
export PAYMENT_METHOD_CASH_ID=<id>
export PAYMENT_METHOD_CREDIT_CARD_ID=<id>
```

## Preflight（read-only）

```sh
make csv-import-preflight
```

source 契約（completed billingの`completed_at`、payment親子、method placeholder、split算術、billing/payment total一致、保険比率0〜1、保険額の符号、割引額の非負を含む）に加えて、target の6 seed binding、全 migrated ID/FK 列の BIGINT、21 sequence、会計FK、`payments.billing_id` UNIQUE、`payment_methods(clinic_id, system_key)` partial UNIQUE、`payment_splits(billing_id)` index、対象 band が空であることを検証します。completed billingにpayment 1件とsplit 1〜2件が揃わない場合を含め、1件でも不一致なら apply へ進みません。

## Apply

`TARGET_DB_NAME` は `.env.local` / 実行環境の `DB_NAME` と完全一致する値を明示します。

```sh
make csv-import TARGET_DB_NAME=<exact-db-name>
```

Make target は次の確認を CLI へ渡します。

run sheet上でbackup取得・復元手順・担当者を再確認したoperatorだけがこのtargetを実行し、target実行そのものを次の明示確認として扱います。

- target write を承認済み
- 復元確認済み backup が存在
- target host が `DB_HOST` と完全一致
- target database が `DB_NAME` と完全一致
- aggregate count + 6 seed IDだけを含むreportの新規パス（既存 report を上書きしない）

apply 後、各 table の件数・clinic isolation・会計親子/支払方法/分割金額の整合・completed timestamp・sequence floor/max ID を transaction 内で検証してから commit します。21 sequence は既存値を下げず、次回 application ID が `1,000,000,000` 以上かつ現行`max(id)`超になるよう進めます。

## Verify（read-only）

```sh
make csv-import-verify
```

単一の REPEATABLE READ snapshot内で、manifest 件数、医院割当、6 seed binding、BIGINT/sequence契約、`payments.billing_id`を親キーとするpayment_splits論理親子、method seed binding、payment金額、completed billing/payment対応、split合計/received/changeを再検証します。seedの並行変更を防ぐ`FOR SHARE`を使うためPostgreSQL transaction自体は`READ ONLY`指定にしませんが、このverify経路はtarget dataを変更しません。preflight/apply/verify はCSV列が依存する全FKについてtarget上にvalidated制約が存在することも検証し、その制約をCOPY時にPostgreSQLが適用するため、orphanが1件でもあればapply transactionはcommitされません。

## 失敗・rollback

- transaction を開始できなかった場合、target data は未変更でreportは `FAILED_BEFORE_TRANSACTION` になります。接続・DB状態を確認し、原因を解消して新しい作業確認を取るまで再実行しません。
- transaction 内の失敗はデータ行を自動 rollback し、report は `FAILED_DATA_ROLLED_BACK` になります。同じ source digest でも、原因を解消して新しい作業確認を取るまで再実行しません。PostgreSQL sequence はtransaction rollback対象外ですが、処理は値を下げずapplication予約域へ進めるだけなので、失敗時に残り得るのは安全な番号飛びだけです。
- commit応答が失われた場合は、commit済みかrollback済みかを断定せずreportを `COMMIT_OUTCOME_UNKNOWN` とします。再実行・backup restore・運用開始をすべて止め、同じmanifest/seedでread-onlyの `make csv-import-verify` を実行してDB管理者が結果を照合するまで状態変更を行いません。
- process crashや強制終了後にapply reportがmissing、malformed、または `STARTED` のままの場合もcommit結果を証明できないため、`COMMIT_OUTCOME_UNKNOWN` と同じ未確定状態として扱います。reportを作り直す再実行やbackup restoreへ進まず、targetを隔離してread-only verifyとDB照合を先に行います。
- commit 後の rollback は、後続 application row を cascade delete する危険があるため importer に削除コマンドを持たせません。メンテナンス状態を維持し、事前に検証した full backup を復元します。
- band が既に占有されている場合、`make csv-import` は置換せず fail-closed します。手書き DELETE や旧 `make stage-import` へ迂回しません。

## Scoped verification

```sh
docker compose exec backend go test ./internal/csvimport -count=1
docker compose exec backend go test ./cmd/csv-import -count=1
docker compose exec backend go test ./cmd/migrate -count=1
```

実PostgreSQL上のcatalog/payment SQL構文と実行計画は、共有DBの排他leaseを取得した独立セッションで `CSVIMPORT_DB_INTEGRATION=1` を付け、`go test ./internal/csvimport -run TestCutoverPaymentTargetSQLAgainstPostgres -count=1 -p 1` を実行します。このテストはtransaction-localな一時テーブルに正常/敵対fixtureを作成し、必ずrollbackします。

実 DB apply、DB reset、STG/PROD 操作はこのテスト手順に含みません。
