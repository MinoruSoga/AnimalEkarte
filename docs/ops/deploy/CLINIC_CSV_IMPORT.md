# 医院 CSV カットオーバー投入（F6）

更新日: 2026-07-31

`old_db` が出力した AnimalEkarte 形状 CSV 21 テーブルを、AnimalEkarte DB へ投入する正式な consumer 手順です。`old_db` DB へは接続せず、医院・run 固定の `manifest.json` と CSV ディレクトリだけを読みます。このF6経路は対象DBへのcutoverであり、`backend/migrations/seeds/` のCSVやmanifestは生成・更新しません。

現行 KNJO source は未完全なため、`payments.csv` / `payment_splits.csv` は意図的にheader-onlyです。CSVの形状確認には使えますが、既知の未確定支払候補を持つbillingはproducer側で `needs_review` に隔離され、正式manifestは支払・分割支払がどちらも正件数でなければpreflightで拒否されます。producer契約は `billings.csv` の `completed_at` と completed billingごとのpayment graphに対応済みですが、完全かつ検証済みのKNJO sourceから新しいformal bundleを生成するまでapplyはBLOCKEDです。

## Issue #250 受け入れ条件との対応（consumer 側）

| AC | consumer 実装 | 状態 |
|---|---|---|
| 全対象 table の source→target mapping | 下表 + `csvimport.CutoverMappingCoverage()` / `CutoverTableSpecs()` | 21表 formal 固定。**業務上の個人責任者名は USER ops が run sheet で確定** |
| dry-run と代表データ照合 | `make csv-import-preflight`（source+target read-only）。DB 書き込みなし | 実装済み。代表データ手動照合は USER |
| clinic/owner/pet/staff 越境・FK・金額 | band 検査、clinic_id isolation、validated FK、payment graph | 実装+unit。実 DB 照合は rehearsal |
| rerun が二重作成しない / 部分 commit なし | 非空 band は `CUTOVER_REF_BAND_OCCUPIED` で fail-closed。単一 transaction | 実装+unit |
| stop/rollback・非PHI audit | report は status/timestamp、manifest digest、clinic/run/target metadata、ID band、aggregate count、6 seed ID、failure stage のみ。エラーは table/行番号/`CUTOVER_REF_*` のみ（CSV cell/氏名等なし） | 実装+unit / F8 rehearsal |
| production cutover | #253/#254/#255 gate 後に USER 実施 | **本 lane では実施しない** |

### 21 表 source→target mapping（formal cutover v1）

正本は `backend/internal/csvimport.CutoverTableSpecs()`。以下は operator 向け要約（PHI 列値は載せない）。

| # | target table | CSV | clinic 列 | isolation | consumer |
|---:|---|---|---|---|---|
| 1 | staffs | staffs.csv | yes | clinic_id_column | formal_cutover_v1 |
| 2 | procedures | procedures.csv | yes | clinic_id_column | formal_cutover_v1 |
| 3 | merchandise_items | merchandise_items.csv | yes | clinic_id_column | formal_cutover_v1 |
| 4 | owners | owners.csv | yes | clinic_id_column | formal_cutover_v1 |
| 5 | pets | pets.csv | yes | clinic_id_column | formal_cutover_v1 |
| 6 | medical_records | medical_records.csv | yes | clinic_id_column | formal_cutover_v1 |
| 7 | inquiries | inquiries.csv | no | id_band_and_parent_fk | formal_cutover_v1 |
| 8 | clinical_plans | clinical_plans.csv | no | id_band_and_parent_fk | formal_cutover_v1 |
| 9 | vital_records | vital_records.csv | yes | clinic_id_column | formal_cutover_v1 |
| 10 | appointments | appointments.csv | yes | clinic_id_column | formal_cutover_v1 |
| 11 | appointment_trimming_details | appointment_trimming_details.csv | yes | clinic_id_column | formal_cutover_v1 |
| 12 | billings | billings.csv | yes | clinic_id_column | formal_cutover_v1 |
| 13 | billing_items | billing_items.csv | no | id_band_and_parent_fk | formal_cutover_v1 |
| 14 | payments | payments.csv | yes | clinic_id_column | formal_cutover_v1 |
| 15 | payment_splits | payment_splits.csv | yes | clinic_id_column | formal_cutover_v1 |
| 16 | estimates | estimates.csv | yes | clinic_id_column | formal_cutover_v1 |
| 17 | estimate_items | estimate_items.csv | no | id_band_and_parent_fk | formal_cutover_v1 |
| 18 | exams | exams.csv | yes | clinic_id_column | formal_cutover_v1 |
| 19 | exam_results | exam_results.csv | no | id_band_and_parent_fk | formal_cutover_v1 |
| 20 | vaccines | vaccines.csv | yes | clinic_id_column | formal_cutover_v1 |
| 21 | vaccinations | vaccinations.csv | yes | clinic_id_column | formal_cutover_v1 |

- **dry-run**: `make csv-import-preflight` が source digest / 6 seed binding / 空 band / FK catalog を read-only 検証する。apply は別コマンド。
- **idempotency**: apply 後の再 preflight/apply は `CUTOVER_REF_BAND_OCCUPIED` で拒否（既存行の置換・削除なし）。
- **verify 再実行**: `make csv-import-verify` は read-only で何度でも可（非空 band をエラーにしない）。
- **非PHI エラー ID**: `CUTOVER_REF_BAND_OCCUPIED` / `CUTOVER_REF_CLINIC_ISOLATION` / `CUTOVER_REF_ROW_COUNT` と table・CSV 行番号のみ。セル値は出さない。

`REHEARSAL_ONLY` / `PARTIAL` bundleを
`backend/migrations/seeds/_old_db_handoff/<clinic>/` に置くことは、
保管場所であり `cmd/migrate` / `make seed` は読みません。正式 `make csv-import`
preflight はこれを拒否します。共有 STG へ載せる場合は
`STG_UAT_CSV_IMPORT_ALLOW_REHEARSAL=YES_I_UNDERSTAND` と target 確認付きの
`make stg-uat-handoff`（全医院。seed ID / manifest SHA は
`scripts/stg-uat-old-db-handoff.sh` が `_old_db_handoff` と `002_master` から埋め、
接続は同スクリプトの export と gitignored `scripts/stg-uat-old-db-handoff.local.env`）
を使います。importer が受理する code/ordinal/clinic_id は
hachioji=1、jouto=2、shikishima=3、hakobuneco=4 だけです。
apply report が既に `PASS` の医院は wrapper が skip し、残りへ進みます。
低レベル fallback は `make stg-uat-import`。医院別の配置コマンドは
`make old-db-handoff-stage`（手順:
[OLD_DB_HANDOFF_LOCAL.md](./OLD_DB_HANDOFF_LOCAL.md)）。
21 CSVを `003_demo` へ直接コピーしてseedとして扱うことは禁止します。実行可能seedへの
変換は現行コードでは未実装です。[SEED_MIGRATION_OPERATIONS.md](./SEED_MIGRATION_OPERATIONS.md)
のBLOCKED境界に従い、21表専用adapterを実装・検証するまでは migrate seed にはしません。

## 安全境界

- source は絶対パスを `/migration-input:ro` で read-only mount する。
- one-shot containerへ渡すsecretはDB接続用環境変数だけに限定し、`.env.local` 全体をcontainerへ注入しない。
- `clinic-migration-run-report.json` に記録された manifest SHA-256 を別経路で受領し、source directory 内から自己申告値を拾わない。
- manifest は `PASS`、`animalekarte_stage`、医院 code/ordinal、run ID、10M ID band、21 テーブル固定順、各 CSV SHA-256 が完全一致する場合だけ受理する。さらに manifest schema version、承認済みstage mapping bundle（`020_canonical.sql` + `030_stage.sql`）/CSV contract SHA-256、`TRUSTED_CANDIDATE`、完全性 `PASS`、検証済み source identity、明示的な空の不完全table一覧、stage build UUID、DB1/DB2/DB3 summary digestと生成順、base-load/KNJO evidence digest、支払・分割支払の正件数を固定構造で検証し、未知fieldも拒否する。
- source directory は 0700、manifest/CSV は 0600、symlink は不可。
- parserのメモリ/時間上限としてmanifestは4MiB、各CSVは512MiB、payment親は100万件、splitは親あたり最大2件を上限とし、超過時はfail-closedする。
- target seed ID は6つとも明示する。予約種別・支払方法を表示名や先頭行から暗黙解決しない。
- apply は対象 band が21表すべて空の場合だけ実行し、既存行を削除・置換しない。
- apply は単一 transaction、advisory lock、table lockを使う。CSV が preflight 後に変わった場合は再 SHA-256 検証で全 rollback する。`make stg-uat-handoff` だけは PlanetScale が長時間の未コミット INSERT で接続を切るため、表ごとに commit する。成功済み prefix（件数・clinic isolation が manifest と一致）は再実行で skip し、一致しない非空 band は fail-closed のまま上書きしない。医院単位では apply report が `PASS` なら wrapper がその医院の `make stg-uat-import` を skip する。`STARTED` / 失敗 report は no-clobber のため、再実行前に失敗日時付きへリネームする。
- CSV/manifest は PHI を含み得る。行値をログ・report・Git・チャットへ出さない。report は status/timestamp、manifest digest、clinic/run/target metadata、ID band、aggregate count、再検証に必要な非PHIの6 seed ID、failure stage を記録し、CSV cell値は記録せず、固定mount `/migration-reports` 直下へ0600/no-clobberで作成する。

## 事前準備

1. target DB の検証済み full backup を取得し、復元手順と担当者を確定する。
2. 対象を、この HEAD の現行 `backend/migrations/*.sql`（現在は統合済み `001_init.sql`）で再構築済みであることを確認する。異なる内容の001が適用済みのDBはchecksum mismatchになるため `DB_RESET=true` 相当の承認済み再構築が必須で、手書きSQLによる差分適用は使わない。001内の通常の `CREATE INDEX` は対象テーブルへの書き込みを待たせ得るため、事前リハーサルで所要時間を測り、maintenance window内で適用する。
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

source 契約（completed billingの`completed_at`、payment親子、billing completionとpayment/split timestampの完全一致、non-completed billingへのpayment禁止、method placeholder、split算術、billing/payment total一致、保険比率0〜1、保険額の符号、割引額の非負を含む）に加えて、target の6 seed binding、全 migrated ID/FK 列の BIGINT、21 sequence、会計FK、`payments.billing_id` UNIQUE、`payment_methods(clinic_id, system_key)` partial UNIQUE、`payment_splits(billing_id)` index、対象 band が空であることを検証します。completed billingにpayment 1件とsplit 1〜2件が揃わない場合を含め、1件でも不一致なら apply へ進みません。

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
- status/timestamp、manifest digest、clinic/run/target metadata、ID band、aggregate count、6 seed ID、failure stage だけを含みCSV cell値を含まないreportの新規パス（既存 report を上書きしない）

apply 後、各 table の件数・clinic isolation・会計親子/支払方法/分割金額の整合・completed timestamp・sequence floor/max ID を transaction 内で検証してから commit します。21 sequence は既存値を下げず、次回 application ID が `1,000,000,000` 以上かつ現行`max(id)`超になるよう進めます。

PostgreSQL は RLS 有効テーブルへの `COPY FROM` を、superuser / `BYPASSRLS` / （FORCE RLS なしの）table owner 以外に拒否します（SQLSTATE `0A000`）。`app.bypass_rls` は `has_clinic_access()` 用 GUC であり、この COPY 制限は解除しません。importer は RLS で COPY が拒否される role では、transaction 内 TEMP テーブルへ COPY したあと `INSERT ... SELECT` を ctid バッチで本表へ載せます。STG UAT はさらに表ごとに commit し、PlanetScale が数百万行の未コミット transaction で backend を切断するのを避けます。table owner による正式 cutover（ローカル docker）は従来どおり本表へ直接 COPY し、21 表を単一 transaction で commit します。

## Verify（read-only）

```sh
make csv-import-verify
```

単一の REPEATABLE READ snapshot内で、manifest 件数、医院割当、6 seed binding、BIGINT/sequence契約、`payments.billing_id`を親キーとするpayment_splits論理親子、method seed binding、payment金額、completed billing/payment対応、split合計/received/changeを再検証します。seedの並行変更を防ぐ`FOR SHARE`を使うためPostgreSQL transaction自体は`READ ONLY`指定にしませんが、このverify経路はtarget dataを変更しません。preflight/apply/verify はCSV列が依存する全FKについてtarget上にvalidated制約が存在することも検証し、その制約をCOPY時にPostgreSQLが適用するため、orphanが1件でもあればapply transactionはcommitされません。

## 失敗・rollback

- transaction を開始できなかった場合、target data は未変更でreportは `FAILED_BEFORE_TRANSACTION` になります。接続・DB状態を確認し、原因を解消して新しい作業確認を取るまで再実行しません。
- 正式 `make csv-import` の transaction 内失敗はデータ行を自動 rollback し、report は `FAILED_DATA_ROLLED_BACK` になります。`make stg-uat-handoff` は表ごと commit するため、失敗した表だけ rollback し report は `FAILED_TABLE_ROLLED_BACK` です。成功済みの表は残るので、原因解消後に同じ command を再実行すると一致した prefix を skip します。一致しない非空 band は上書きしません。PostgreSQL sequence は transaction rollback 対象外ですが、処理は値を下げず application 予約域へ進めるだけなので、失敗時に残り得るのは安全な番号飛びだけです。report は no-clobber のため、失敗後の再実行前に `sensitive-local/csv-import-reports/<clinic>-<run>-stg-uat-apply.json`（または formal の `*-apply.json`）を失敗日時付きへリネームします。
- commit応答が失われた場合は、commit済みかrollback済みかを断定せずreportを `COMMIT_OUTCOME_UNKNOWN` とします。再実行・backup restore・運用開始をすべて止め、同じmanifest/seedでread-onlyの `make csv-import-verify` を実行してDB管理者が結果を照合するまで状態変更を行いません。
- process crashや強制終了後にapply reportがmissing、malformed、または `STARTED` のままの場合もcommit結果を証明できないため、`COMMIT_OUTCOME_UNKNOWN` と同じ未確定状態として扱います。reportを作り直す再実行やbackup restoreへ進まず、targetを隔離してread-only verifyとDB照合を先に行います。
- one-shot STG UAT importのreportが `PASS` になるのは、apply後の最終verify成功時だけです。commit済みapplyの最終verifyが失敗した場合は `FAILED_POST_COMMIT_VERIFY` / failure stage `verify` とし、targetを隔離して原因を照合します。
- commit 後の rollback は、後続 application row を cascade delete する危険があるため importer に削除コマンドを持たせません。メンテナンス状態を維持し、事前に検証した full backup を復元します。
- band が既に占有されている場合、`make csv-import` は置換せず fail-closed します。手書き DELETE や別経路への迂回はしません。

## Scoped verification

```sh
docker compose exec backend go test ./internal/csvimport -count=1
docker compose exec backend go test ./cmd/csv-import -count=1
docker compose exec backend go test ./cmd/migrate -count=1
```

実PostgreSQL上のcatalog/payment SQL構文と実行計画は、共有DBの排他leaseを取得した独立セッションで `CSVIMPORT_DB_INTEGRATION=1` を付け、`go test ./internal/csvimport -run TestCutoverPaymentTargetSQLAgainstPostgres -count=1 -p 1` を実行します。このテストはtransaction-localな一時テーブルに正常/敵対fixtureを作成し、必ずrollbackします。

実 DB apply、DB reset、STG/PROD 操作はこのテスト手順に含みません。

F8 の失敗側リハーサルは、通常 importer に fault injection を追加せず、専用の
[F8 G4 synthetic failure rehearsal](F8_G4_FAILURE_REHEARSAL.md) を使用します。

## rehearsal 前提（#250 / BRT-42 · 2026-08-20）

**cutover 実行と架空 COMPLETE bundle は禁止。** 本節は手順の不足を埋めるだけ。

| 前提 | 状態 | 根拠 |
|---|---|---|
| formal COMPLETE producer bundle 受領 | **未記入**（受領記録なし） | Linear BRT-42 / #250。KNJO source 未完全のため apply は本文どおり BLOCKED |
| `payments.csv` / `payment_splits.csv` が正件数 | 現行 KNJO は header-only（本文） | 正件数になるまで preflight 拒否 |
| F8 G4 synthetic failure rehearsal | 手順あり（専用 compose。本番 CSV 不可） | [F8_G4_FAILURE_REHEARSAL.md](F8_G4_FAILURE_REHEARSAL.md) |
| 代表データ手動照合 | **未記入** | USER |
| production cutover | **しない** | #253/#254/#255 gate 後の USER |

COMPLETE 受領後の rehearsal 順（実行は USER。本セッションでは走らせない）:

1. bundle を repo 外に置き、PHI を git / Issue に載せない
2. `make csv-import-preflight`（write 0）
3. 隔離 DB での F8 G4（本番 CSV を渡さない）
4. 隔離 rehearsal apply（共有 STG/PROD は承認後のみ）
5. `make csv-import-verify`（read-only）
6. production cutover は gate 後の別作業
