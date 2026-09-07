# PlanetScale STG migration / master-seed runbook

> **目的**: fresh/rebuilt STGでcurrent DDLと`002_master`を同じ`cmd/migrate` pathから適用し、検証する。
> **安全境界**: 本書を根拠にagentがDBをquery/write/resetしたりdeployしたりしない。shared STG操作はdata ownerとapproved operatorの明示承認が必要。

## 1. Current state

- `cmd/migrate`はtop-level `backend/migrations/*.sql`を昇順適用した後、`BundleOrderForEnv(APP_ENV)`を適用する。
- current CSV bundle orderは**全environmentで`002_master`だけ**。`003_demo` / `004_staging`は削除済みで復元対象ではない。
- フェーズ3 が LoginForm と同じ合成デモログインを upsert する。パスワードはコード定数（全デモ共通）。production の API は受け付けない。
- `002_master/manifest.json`がtable inventory/load orderのSSOT（現在12 table）。固定checksum、固定row count、固定table countをrunbookへ複製しない。
- fresh DBのexpected historyはcurrent DDL filename keys + `seeds/002_master` + ログイン seed 適用時の `seeds/003_login`。`Migration key coverage missing=0`を一次判定にする。
- Cloudflare backend workflowはdeploy、`POST /_internal/migrate`、post-migrate healthの順。path filterによりbackend対象変更だけが自動起動する。

STGの画面デモログインは migrate フェーズ3が合成 `stg-staff-*@example.test`（権限「一般」）を upsert する。CSV に account は載せない。操作用の個別 account は approved provisioning、21表 clinical dataはapproved `make stg-uat-handoff`（`_old_db_handoff` の REHEARSAL_ONLY を含む）または formal cutover を使う。`cmd/migrate` は 21 CSV を読まない。PlanetScale の user-defined role は table owner / `BYPASSRLS` ではないため、RLS 付き 21 表への直接 `COPY FROM` は `0A000` で拒否される。handoff importer が TEMP COPY + バッチ `INSERT SELECT` で回避し、STG UAT は表ごとに commit する（長時間の単一 transaction は backend 切断になる）。`pscale role reset-default` で app `postgres` role のパスワードを回さない。

## 2. Pre-deploy stop gates

次の1つでも満たさなければdeploy/rebuildを止める。

- target database/branch、data owner、operator、maintenance windowをrun sheetに固定した。
- 保存対象、verified backup/restore、rollback decisionを確認した。
- target Wranglerの`secrets.required` namesとvarsを値を表示せず確認した。
- current artifactにtop-level DDLと`002_master/manifest.json`/全listed CSVが揃う。
- legacy keys がある場合は、現行 master-only translation と対象 DB の master 完全性を照合し、不整合時の reviewed recovery plan がある。
- backend production/STG gateとbilling状態など、該当release gateがgreen。

### `public` schema prerequisite

過去に`DROP SCHEMA public CASCADE`だけが実行されたDBでは`public`自体が存在しない。`ensureMigrationsTable`は`no schema has been selected to create in`でfailする。これは自動seed failureではなくschema prerequisite failureである。

Approved operatorがdeploy前にnames-only target verificationとread-only schema presence checkを行う。`public`が無い場合は、同じchange record、target confirmation、backup/rollback、write approvalの下で作成する。agentや通常deployはこのdirect DB stepを行わない。

Direct connectionを使う場合はapproved short-lived roleを用いる。migrationはadvisory lockを使うためHyperdriveをmanual migration pathの代わりにしない。credential値をlog/file/chatへ残さない。

## 3. Normal apply

Approved operatorがreviewed `main -> staging` deliveryまたはbackend workflow dispatchを行う。通常pathは必ずrepositoryの`cmd/migrate`を使う。

Expected log shape:

```text
Migration completed file=<each current top-level SQL filename>
Migration summary applied=N skipped=M total=T
Seed bundle loaded bundle=002_master
Seed bundle summary applied=1 skipped=0 total=1
Migration key coverage missing=0 extra=X expected=E recorded=R
```

既適用なら`applied/skipped`は変わり得る。固定値ではなく、current plan、coverage `missing=0`、checksum mismatchなし、workflow exit statusを確認する。

## 4. Verification

Repository artifactsからexpected inventoryを導出する。

1. top-level `backend/migrations/*.sql` filenameを列挙する。
2. `backend/migrations/seeds/002_master/manifest.json`のtable/file順を読み、listed fileが全て存在することを確認する。
3. migrate logでcurrent DDLと`seeds/002_master`のcoverage `missing=0`を確認する。
4. approved read-only DB verificationでは`schema_migrations` keysをexpected planと比較し、manifestの各tableについてcurrent CSVのrequired master rowsを検証する。checksum値やsecretをrun sheetへ転記しない。
5. `/health` `200`をlivenessとして確認する。その後、provisioned account/login/permissionsと必要なhandoff/importを別gateで確認する。

`002_master`の現在のclinicsはmanifest/CSVから導出し、`is_active` contractで確認する。「demo clinicsが3件」「system_admin group ID 1」等の旧期待を使わない。

## 5. Failure handling

- `public` missing: deployを止め、§2のapproved prerequisite procedureへ戻る。
- checksum mismatch: file/historyを手修正せず、artifact/target取り違えとapproved recovery planを調査する。
- migration coverage `missing>0`: release stop。healthだけを成功証跡にしない。
- legacy key translation failure: [SEED_MIGRATION_OPERATIONS.md](./SEED_MIGRATION_OPERATIONS.md#2-legacy-seed-keys) の現行 translation 契約と対象 DB を照合する。削除済み bundle 参照の旧不具合とは切り分け、manual baseline や reset 不要宣言をしない。
- partial/unknown outcome: rerunやschema dropを重ねず、run ID、commit、target、last completed phaseを記録し、approved read-only verificationで状態を確定する。

### Manual seed fallback is prohibited

`\copy`や個別INSERTでmanifestの一部だけをロードし、固定checksumを`schema_migrations`へ書くfallbackは削除した。transaction、全12 table、sequence、advisory lock、recordingを同一に再現できず、`ON CONFLICT DO NOTHING`はmismatchを隠す。

障害を診断して**同じ`cmd/migrate` pathを再実行**する。direct recoveryが不可避なら、別のreviewed transactional tool、scoped tests、rollback/rehearsalを先に実装する。runbook内の手書きSQLで代替しない。

## 6. Rebuild / rollback boundary

Schema rebuildは通常deployやcleanupではない。target、data loss、backup/restore、downtime、operator、approvalが揃うまで実行しない。AWS/RDSは退役済みでrollback先ではない。

Approved rebuildでは`DROP SCHEMA`と`CREATE SCHEMA`を対で扱い、`public` missing stateを残さない。その後は§3のnormal `cmd/migrate` pathへ戻る。partial table repair、retired demo bundle restore、old checksum insertionは行わない。

## 7. Deferred blockers

- Legacy seed-key translation は現行 `002_master` の記録だけを扱う。対象 DB の既存 master が満たされることは別途確認が必要。
- STG/productionのlive state、backup restore readiness、secret values、PlanetScale CLI behaviorはrepository docsだけでは検証済みとしない。
- Demo/clinical dataはmaster seedとは別のapproved provisioning/cutoverが必要。
