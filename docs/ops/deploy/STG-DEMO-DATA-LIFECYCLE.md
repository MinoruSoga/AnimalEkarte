# STG データライフサイクル

> **目的**: master seed、運用アカウント、smoke/investigation dataの作成・cleanup境界を定義する。
> **現行contract**: 全 `APP_ENV` でseed bundleは `002_master` のみ。`003_demo` / `004_staging` は退役済み。

## 1. 分類

| category | source | retention |
|---|---|---|
| Master seed | `backend/migrations/seeds/002_master/` | environment lifetime。通常cleanupしない |
| Operation-provisioned account | [STAFF_ACCOUNT_PROVISIONING.md](./STAFF_ACCOUNT_PROVISIONING.md) | owner/expiryをrun sheetに記録 |
| Smoke data | [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) | 同じrunでcleanup |
| Investigation data | approved task/run sheet | ownerと期限を固定し、終了時cleanup |

`002_master` は医院骨格と参照masterを提供する。現在のclinic CSVにはID 1/2があり、clinic-scoped permission groupsが含まれる。ID 1やgroup名を「system adminとして削除不可」と扱わない。privileged demo accountはseedされない。

ローカルを含む全environmentで `BundleOrderForEnv(APP_ENV)` は現在 `002_master` だけを返す。臨床/demo dataはseed bundleへ復元せず、ローカルは `_old_db_handoff`、STG cutover/UATは承認済みF6経路を使う。

## 2. 作成ルール

- master seedは`cmd/migrate`の同じ経路だけで適用する。CSVや`schema_migrations`を手編集しない。
- operation accountはapproved provisioningで作る。既知passwordをseedや文書へ置かない。
- smoke/investigation dataはclinic scopeとownerを決め、作成時に返ったIDをrun sheetへ記録する。
- cookieはbrowserのHttpOnly sessionまたは権限0600のcookie jarで扱う。header/cookie/token値を文書へ貼らない。
- API mutationが一律に`audit_logs`へ書くとは仮定しない。[CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) のroute matrixに従う。

## 3. Smoke cleanup

現在のsmoke手順は既存clinicを編集・復元し、test permission groupとtest staffを作る。`.related_count` のような存在しないresponse fieldに依存しない。

1. run sheetのknown IDsとlist/detail APIで対象を再確認する。
2. このrunのtest permission groupを削除する。active assignmentがあれば停止し、勝手にdetachしない。
3. このrunのtest staffを削除する。child dependencyによる`409`なら停止する。
4. clinicは削除せず、保存した元値へ復元してGETで確認する。もし別のapproved caseでtest clinicを作った場合だけ最後に削除する。
5. list/detailでstateを確認する。permission-groupの成功mutationだけは明示contractに従いauditも確認する。

依存関係が異なるcaseではAPIの`409`を正として停止し、順序を推測してdirect SQLへ切り替えない。

## 4. Direct DB operation

通常cleanupでは禁止する。API全断などのincidentで承認された場合だけ、次を事前に固定する。

- target environment/databaseとclinic scope
- exact IDsとdependency graph
- verified backup/restoreとrollback条件
- single transactionとpost-check
- operator、時刻、incident/change ID

manual SQLの後にapplication由来を装う`audit_logs` rowをINSERTしない。incident evidenceとして別に実施記録を残す。

## 5. 残置判定

Smoke dataは原則残置しない。investigation dataを一時保持する場合はowner、purpose、clinic scope、expiry、cleanup APIがすべて明示され、通常UI/operationを妨げないことを確認する。tenant isolation違反、visibleな不要data、次回smokeを妨げる重複はrelease stopとする。

## 6. DB再構築境界

`DB_RESET=true` は `public` schemaを破棄する。通常deployやcleanupの手段ではない。

- local disposable DB: [LOCAL_DB_RESET.md](./LOCAL_DB_RESET.md) のuser-owned destructive stepだけを使う。
- shared STG/production: data owner、backup/restore、downtime、target、approvalが確定するまで実行しない。
- current Cloudflare workflowに`db_reset` inputはない。AWSは退役済みでrollback/reset先ではない。
- 再構築後も適用するseedは `BundleOrderForEnv(APP_ENV)`、現在は `002_master` のみ。

health `200` はlivenessだけを示す。運用account、permissions、必要なhandoff/import、corrected CRUD casesを別に確認する。

## 7. Historical note

過去に提案された `stg-smoke-cleanup.yml` / `stg-cleanup-smoke-test-data.sh` は実装されていないhistorical proposalであり、実行手順ではない。存在しないautomationをrelease gateにしない。cleanupは本書とcorrected CRUD runbookのmanual API procedureを使う。

## 参考

- [CRUD smoke](./CRUD-SMOKE-TEST.md)
- [Seed/migration operations](./SEED_MIGRATION_OPERATIONS.md)
- [STG seed runbook](./STG_PLANETSCALE_SEED_RUNBOOK.md)
- [Release readiness](./runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md)
