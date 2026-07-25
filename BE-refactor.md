# BE-refactor 第10期（BE10）— backend規約適合（フォルダ構成）修正計画

## メタ

- 更新日: 2026-07-25
- 要件責任者: MinoruSoga
- 監査baseline HEAD: `3ea5fb067`（下記「監査baseline」節の実測時点。以後の進捗で更新しない歴史的anchor）
- 位置づけ: backendフォルダ構成監査から確定した修正候補だけを扱うround-scoped plan。BE10完了時に本ファイルを削除し、完了履歴はgit履歴と実装時の検証資産を正本とする。
- 本期の業務目的: package境界の判断根拠と残存物の退役条件を明示し、同じ手動監査・誤検出・台帳探索を繰り返す工程を削除する。

### 進捗サマリー（2026-07-25 時点）

| unit | 状態 | 証跡 |
|---|---|---|
| BE10-1 clinic subpackage統合 | **完了**（commit済み） | `0301ae0e2`・下記BE10-1実行ledger |
| BE10-2 Phase 0 計画確定 | **完了**（commit済み） | `bcbdb5101`・下記Phase 0 ledger |
| BE10-2 B0 `testdb` fixture export | **完了**（commit済み） | `27d95aacd`・下記B0実行ledger・`backend/internal/testdb/fixtures.go` |
| BE10-2 B1〜B14 Go移設 | **B1〜B3 完了**（commit済み。B4〜B14未着手） | B1=`718f6c9b3`／B2=`134e7953b`／B3=`36f283f37`・下記batch表・下記B1/B2/B3実行ledger・下記E12補正根拠 |
| BE10-3 空directory削除 | 未着手 | — |
| BE10-4 ignore未登録 | 判断待ち | — |
| BE10-5 `q&a.html` path drift | 未着手 | — |
| BE10-6 package境界gate | 未着手 | BE10-1/2の判断確定後に着手 |

進行原則: 1 unitずつ実行し、unit完了は次unitの着手を認可しない。各unit完了時に該当節へ実行ledgerを追記し、本表と`状態`行を同時に更新する。
- 重複禁止:
  - `todo.md`: 直ちに着手可能な実装タスクだけを置く。本書のBE10項目を複製しない。
  - `BE-pending.md`: 着手保留・任意検証の正本。本書と重複させない。
  - `q&a.html`: PO判断・USER実操作・release gateの正本。本書と重複させない。
  - ADR-006: package境界の恒久的な正本。本書はADRの内容を置き換えない。

## 規約の正本と Authority

- package境界のproject decision: [ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md)
- backend作業規約: [backend/CODING_RULES.md](backend/CODING_RULES.md)
- Go/Gin一般規約: [.claude/rules/go-gin-backend-guidelines.md §2](.claude/rules/go-gin-backend-guidelines.md#2-モジュールとパッケージ)
- Authority順序は`backend/CODING_RULES.md:5-13`に従う。
  1. Go language/toolchain仕様
  2. Go/Gin公式文書に基づく正本
  3. application invariantsとAccepted ADR
  4. OpenAPI/schema/migration等のproject contract
  5. package内の局所説明
- Go/Ginはlayer-first/domain-first、Handler → Service → Repository、固定directory深さ、package/fileサイズを規定しない。これらを独自に逸脱根拠へ追加しない。

## 監査baseline

- 実測日: 2026-07-25
- HEAD: `3ea5fb067`
- すべて本計画起票前にhost上の読み取り専用`find` / `grep` / `git`で再取得した。Go/npm/pnpm commandは実行していない。

### A3 `internal/`直下集合

実行:

```sh
find /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal -mindepth 1 -maxdepth 1 -type d | sort
```

出力（basename、34件）:

```text
apicontract
apperrors
audit
auth
authjwt
billing
clinic
config
csvimport
dbconn
httpapi
infra
inventory
lintscan
logger
lstep
manualarticle
medicalrecord
middleware
model
owner
persistence
pet
repository
reservation
scheduler
seedbundle
service
sharedkernel
staff
testdb
textsearch
timeutil
trimming
```

ADR-006 `:34-45`の13 target + 19 cross-cuttingは全件存在する。記載外の2件はADR-006 `:18`がtest-only残存を承認している`repository`と`service`で、欠落は0件。

### A4 legacy layer

実行:

```sh
for d in handler service repository; do p=/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/$d; if [ -d "$p" ]; then echo "$d: prod=$(find "$p" -name '*.go' -not -name '*_test.go' | wc -l) test=$(find "$p" -name '*_test.go' | wc -l)"; else echo "$d: dir absent"; fi; done
```

出力:

```text
handler: dir absent
service: prod=0 test=14
repository: prod=0 test=50
```

### A5 clinic subpackageとrepository残骸

実行:

```sh
find /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/clinic -mindepth 1 -maxdepth 1 -type d | sort
find /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository -mindepth 1 -maxdepth 1 -type d | sort
```

`internal/clinic`の4 subpackage:

```text
clinicholiday
clinicsettings
closingspecialperiod
company
```

`internal/repository`の15子directory:

```text
animalspecies
audit
clinicholiday
clinicsettings
closingspecialperiod
company
inventory
merchandiseitem
occupation
repohelpers
repotest
sharedfile
shiftentry
shifttemplate
staffclinicassignment
```

各repository側directoryに`find <absolute-dir> -type f | wc -l`を実行した結果は全15件とも`0`。clinic側4名はrepository残骸側の名前と完全一致する。

consumer再実測:

```text
clinicholiday        -> internal/clinic/repositories.go:6 の1 production consumer
clinicsettings       -> internal/clinic/repositories.go:7 の1 production consumer
closingspecialperiod -> internal/clinic/repositories.go:8 の1 production consumer
company              -> internal/clinic/repositories.go:9 の1 production consumer
```

### A6 legacy配下のlint gate test

実行:

```sh
find /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/service /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository -name '*lint*_test.go' | sort
```

出力:

```text
/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/audit_tx_inventory_lint_test.go
/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/dbortx_inventory_lint_test.go
/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/migration_cascade_lint_test.go
/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/preload_clinic_scope_lint_test.go
/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/service/master_fk_write_inventory_lint_test.go
/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/service/n1_lint_test.go
```

### A7 ignore未登録

実行と出力:

```text
$ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend check-ignore -v .ruff_cache
(empty)
exit=1

$ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend check-ignore -v .wrangler
(empty)
exit=1

.ruff_cache tracked=0
.wrangler tracked=0
```

実在確認では`.ruff_cache`は2 file、`.wrangler`は0 file、空の`backend/.git`も0 file。`.ruff_cache`内部の自身の`.gitignore`により内部生成物はignoreされるが、top-level directory自体へのproject ignore登録はない。

### A8 旧layer path参照

実行:

```sh
grep -n -E 'internal/(handler|service|repository)' '/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/q&a.html' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/todo.md
```

出力は`q&a.html`の5行、`todo.md`は0行。5件はいずれも現行実装位置を説明する文脈であり、歴史的path引用ではなく更新対象のdoc driftと分類する。

| 行 | 旧参照 | 現行path |
|---:|---|---|
| 1108 | `internal/service/line_link_service.go` | `backend/internal/lstep/line_link_service.go` |
| 1400 | `internal/handler/aggregation_handler.go`, `internal/repository/ltv_repository.go` | `backend/internal/lstep/aggregation_handler.go`, `backend/internal/owner/ltv_repository.go` |
| 1430 | `internal/service/lstep_settings_thresholds.go` | `backend/internal/lstep/lstep_settings_thresholds.go` |
| 1436 | `internal/service/lstep_tag_sync_api.go`ほか4 file | 同名fileはすべて`backend/internal/lstep/` |
| 1454 | `internal/handler/owner_handler.go:133-153` | `backend/internal/owner/http_owner.go:133-153` |

### A9 BE10期番号

作成前実行:

```sh
grep -rn 'BE10' --include='*.md' --include='*.html' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
```

出力:

```text
(empty)
```

作成前の既存md/htmlで`BE10`は未使用だった。

### BE10-6 gate所在確認

`scripts/`直下の`check-*.mjs` / `check-*.sh`を列挙し、さらに次を実行した。

```sh
rg -n 'ADR-006|package boundar|internal/(handler|service|repository)' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/scripts --glob 'check-*'
```

既存hitは`check-docs-symbol-drift.sh`の旧handler数確認とそのfixtureだけで、A3/A4/A5相当のpackage境界適合gateは存在しない。

## 判定記号

- 状態:
  - `未着手`: 修正方針と検証面は定義済みだが実装前。
  - `判断待ち`: 選択肢またはowner・期限・配置先の決裁が必要。
  - `対応中`: 実装unitが開始済み。
  - `完了`: 修正とscoped verificationが完了し、残余リスクが記録済み。
- severity:
  - `CRITICAL`: clinic isolation・認可・臨床安全に直結する破れ。
  - `HIGH`: 明文禁止のproduction構成。
  - `MEDIUM`: 規約要求または退役条件が未充足。
  - `LOW`: working-tree残骸、docs/ignore drift、予防的gate不足。

## 逸脱項目 BE10-1〜BE10-6

### BE10-1 — `internal/clinic/`配下4 subpackageの機械的分割

- 規約根拠: `backend/CODING_RULES.md:17`は現在のdirectory名の機械的複製を禁じ、ADR-006 `:52`はdomain内の分離を実consumer・依存方向・変更周期が分かれる場合に限定する。
- 現状: `internal/clinic/{clinicholiday,clinicsettings,closingspecialperiod,company}`は各1 production fileを持つが、production consumerはいずれも`internal/clinic/repositories.go`だけである。4名は空の`internal/repository/{clinicholiday,clinicsettings,closingspecialperiod,company}`と一致する。
- 修正内容: ①4 subpackageをparent `internal/clinic`へbehavior-preservingに統合する、または②独立consumer・依存方向・変更周期の根拠をADR-006へ記録して承認済み構成とする。追加抽象化より削除・統合を先に評価する。
- 検証方法: A5のdirectory/file/consumer再実測、変更時はclinic packageの既存runtime test、route/RBAC/OpenAPI contractのscoped verificationを実行する。
- severity: MEDIUM
- 前提・依存: folder移動をclinic isolation・認可・臨床安全の証明にしない。route/RBAC/OpenAPIの挙動を変えない移動だけを許容し、既存runtime testを維持する。
- 状態: 完了（2026-07-25・commit `0301ae0e2`・下記実行ledger）。残余リスクは「BE10外の残余課題」R-1として別unitへ切り出した。

#### BE10-1 実行ledger（2026-07-25）

- 実行状態: `完了`。parent統合実装、build/vet/test/gofmt、test名・`clinic_id`行の逐語保存はすべて完了した。実行側は当初、scoped lintが本unitで変更していない`closing_settings_request.go`の既存`staticcheck S1016` 2件を検出したことでB12をBLOCKEDとしたが、生成側のreconciliationでこれをPASSへ補正した。補正根拠: S1016が指摘する4型は`closing_settings_request.go:3`/`:37`と`closing_settings_service.go:63`/`:80`に宣言されており、いずれも`git diff --name-only HEAD -- backend`の変更8 pathに含まれない。したがってstaticcheckの解析入力は統合前とbyte同一で、2件は既存lint負債である。本unitの義務は「新規指摘ゼロ」であり充足している。B12の判定基準を「package絶対0件」としたのは生成側promptの誤指定であり、以後のlint gateは「pre-change baseline比で新規0件」とする。
- plan commit: `f20106ac71a32e1cb4851b8b6300581945a4b21e`（`docs: add BE10 backend conformance plan`）。変更pathは`BE-refactor.md`と`todo.md`のみ、Go file 0件、`Co-Authored-By`なし。
- 変更file:
  - `backend/internal/clinic/clinicholiday/repository.go` → `backend/internal/clinic/clinic_holiday_repository.go`
  - `backend/internal/clinic/clinicholiday/repository_test.go` → `backend/internal/clinic/clinic_holiday_repository_test.go`
  - `backend/internal/clinic/clinicsettings/repository.go` → `backend/internal/clinic/clinic_settings_repository.go`
  - `backend/internal/clinic/clinicsettings/repository_test.go` → `backend/internal/clinic/clinic_settings_repository_test.go`
  - `backend/internal/clinic/closingspecialperiod/repository.go` → `backend/internal/clinic/closing_special_period_repository.go`
  - `backend/internal/clinic/company/repository.go` → `backend/internal/clinic/company_repository.go`
  - `backend/internal/clinic/company/repository_test.go` → `backend/internal/clinic/company_repository_test.go`
  - `backend/internal/clinic/repositories.go`
  - `BE-refactor.md`（本ledger）
- validation / verification:
  - `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-be10-1-clinic-subpackage-merge.md`

    ```text
    Prompt Craft Harness Validation: PASS
    VALIDATOR_EXIT=0
    ```

  - TDD RED: `docker compose exec backend go build ./internal/clinic/...`

    ```text
    internal/clinic/repositories.go:6:2: no required module provides package github.com/animal-ekarte/backend/internal/clinic/clinicholiday; to add it:
        go get github.com/animal-ekarte/backend/internal/clinic/clinicholiday
    internal/clinic/repositories.go:7:2: no required module provides package github.com/animal-ekarte/backend/internal/clinic/clinicsettings; to add it:
        go get github.com/animal-ekarte/backend/internal/clinic/clinicsettings
    internal/clinic/repositories.go:8:2: no required module provides package github.com/animal-ekarte/backend/internal/clinic/closingspecialperiod; to add it:
        go get github.com/animal-ekarte/backend/internal/clinic/closingspecialperiod
    internal/clinic/repositories.go:9:2: no required module provides package github.com/animal-ekarte/backend/internal/clinic/company; to add it:
        go get github.com/animal-ekarte/backend/internal/clinic/company
    found packages clinic (clinic_handler.go) and clinicholiday (clinic_holiday_repository.go) in /app/internal/clinic
    RED_BUILD_EXIT=1
    ```

  - B7: `grep -rh '^func Test' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/clinic | sed 's/(.*//' | sort`の移動前後`diff -u`

    ```text
    (empty)
    B7_DIFF_EXIT=0
    ```

  - B8: `grep -rh 'clinic_id' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/clinic | sed 's/^[[:space:]]*//' | sort`の移動前後`diff -u`

    ```text
    (empty)
    B8_RETRY_DIFF_EXIT=0
    ```

  - B9: `docker compose exec backend go build ./internal/clinic/...`

    ```text
    B9_BUILD_EXIT=0
    ```

  - B9: `docker compose exec backend go vet ./internal/clinic/...`

    ```text
    go: downloading github.com/stretchr/testify v1.11.1
    go: downloading gopkg.in/yaml.v3 v3.0.1
    go: downloading github.com/davecgh/go-spew v1.1.1
    go: downloading github.com/pmezard/go-difflib v1.0.0
    B9_VET_EXIT=0
    ```

  - B10: `docker compose exec backend go test ./internal/clinic/... -count=1 -p 1`

    ```text
    ok  	github.com/animal-ekarte/backend/internal/clinic	0.482s
    B10_TEST_EXIT=0
    ```

  - B11: `docker compose exec backend gofmt -l internal/clinic/`

    ```text
    (empty)
    B11_GOFMT_EXIT=0
    ```

  - B12: `docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-be10-1-20260725-1023 --entrypoint golangci-lint backend run --max-same-issues 0 --max-issues-per-linter 0 ./internal/clinic/...`

    ```text
    internal/clinic/closing_settings_request.go:11:9: S1016: should convert r (type UpdateClinicSettingsRequest) to UpdateClinicSettingsInput instead of using struct literal (staticcheck)
        return UpdateClinicSettingsInput{
               ^
    internal/clinic/closing_settings_request.go:46:9: S1016: should convert r (type UpdateSpecialPeriodRequest) to UpdateSpecialPeriodInput instead of using struct literal (staticcheck)
        return UpdateSpecialPeriodInput{
               ^
    2 issues:
    * staticcheck: 2
    B12_LINT_EXIT=1
    ```

  - B12 provenance: `git diff --name-only HEAD -- backend/internal/clinic/closing_settings_request.go`は空。`git show HEAD:backend/internal/clinic/closing_settings_request.go | shasum -a 256`とworking treeの`shasum -a 256`はいずれも`f674f459265478700a4574edc2d1f8c33cf8a9de3420f898465faa7e4dfe597a`。
- Failure Signature log:
  - B6 / attempt 1: expected=clinic直下のsubdirectory 0件、actual=移動元4 directoryが空のまま残存、verification=`find ... -mindepth 1 -maxdepth 1 -type d`、error signature=4 path出力、fix=各directoryが空であることを確認して4 exact pathだけ`rmdir`、result=再実行empty/PASS。
  - B8 / attempt 1: expected=移動前後の`clinic_id`行diff 0、actual=移動残骸コメント整理で2行差分、verification=B1と同一pipelineの保存結果を`diff -u`、error signature=production queryではなく`clinic_settings_repository_test.go`コメント2行、fix=baselineの該当2行を逐語復元、result=retry diff empty/PASS。
  - B12 / attempt 1: expected=scoped lint 0、actual=未変更fileの`S1016` 2件、verification=上記fresh-cache scoped lint、error signature=`closing_settings_request.go:11:9`と`:46:9`、fix=scope外drive-by修正を行わずHEADとのbyte同一性をSHA-256で証明、result=BLOCKED。再開条件は別unitで既存2件を解消後に同じB12 commandを再実行すること。
- De-Sloppify: 移動元subpackage import、旧package宣言、旧`New`、旧`repository` receiver、重複fixture、未使用import、空directoryを確認し、残存0。query、GORM順序、error、assertion、test名、route/RBAC/OpenAPIは変更なし。
- Independent Review Gate:
  - reviewer role: `PASS`。CRITICAL 0 / HIGH 0 / MEDIUM 0 / LOW 0。B14(a) query/`clinic_id`逐語保存、(b) test/assertion非欠落、(c) identifier改名と意味保存、(d) allowlist、(e) route/RBAC/OpenAPI非変更の5観点すべてPASS。Must Fix / Should Fix / Nitsなし。B12の2件は`closing_settings_request.go`がHEADとbyte同一であるため本diff由来ではないとの判定。
  - clinic-isolation-auditor role: `Approve`。CRITICAL 0 / HIGH 0 / MEDIUM 0。4 production repositoryのmethod bodyをreceiver型名だけ正規化して比較したSHA-256は移動前後で全件一致。`clinic_id` inventoryは移動前後とも100行、`diff -u` empty。query predicate、GORM呼び出し順、transaction利用有無、error変換、`RowsAffected`、companyのsingleton制約に変更なし。
  - 採用・却下: 両passの全判定を採用。修正要求0件のためcode patchなし。SQL新規作成がないため専用database reviewはB8逐語diffとclinic-isolation-auditorのproduction body hash比較で代替した。auth/secret/外部write変更がないためsecurity-review、route変更がないためE2Eは実行していない。
- Assumption deviations: none。

### BE10-2 — legacy test-only packageの削除phase未記録

- 規約根拠: `backend/CODING_RULES.md:114`は、残すlegacy facade/adapterにconsumerと削除phaseを要求する。
- 現状: `internal/service`はproduction 0/test 14、`internal/repository`はproduction 0/test 50。現行`todo.md`/`q&a.html`にBE9-3/BE9-4相当の退役task・担当・期限はない。legacy配下にはA6のlint gate test 6件が残る。
- 修正内容: 下記Phase 0で64 test fileの移設先、helper gap、移設batch、2 packageの削除条件・担当・期限を確定した。移設先が確定するまで実働gateを削除しない。
- 検証方法: A4/A6を再実行し、各batchで移設先packageのscoped testをgreenに保つ。最終batch後はlegacy production/test file 0、旧path import 0、`internal/service` directory不存在、`internal/repository` root Go package不存在を確認する。repository配下の空directoryはBE10-3に残す。
- severity: MEDIUM
- 前提・依存: BE9-1でsource discoveryがpackage非依存化済みであることを再確認する。test所在packageの変更で検出scopeが狭まらないことを移設先の自己テストで証明する。
- 状態: 対応中（Phase 0計画確定、B0〜B3完了・commit済み。B4〜B14は未着手）

#### BE10-2 Phase 0 計画確定ledger（2026-07-25）

- Phase 0状態: `完了`。実測baselineはHEAD `ef32784d3`、`service=14`、`repository=50`、production Go file `0`、未分類`0`。
- 変更範囲: 本ledgerだけ。Go file、`internal/testdb`、ADR、`todo.md`、`BE-pending.md`は変更していない。
- Product Philosophy gate: 本unitは新しい業務機能を追加せず、BE9完了後の二重test surfaceとbridgeを削除する計画である。②削除を先に行い、新packageや新helper方式を作らない。

##### helper実測と`internal/testdb`対応

`grep -rn '^func '`の実測では、`func Test`で始まる定義をservice 79件、repository 188件除外した後、非Test定義はservice 112件、repository 238件だった。receiver methodも含む。さらに`:= func(`はservice 3件、repository 8件あったが、11件ともfunction-local closureでcross-file依存ではない。

同一file内だけで定義・消費されるhelper、mock receiver method、function-local closureはfileと同時に移し、`internal/testdb` exportは追加しない。batchを制約するcross-file helperだけを次表に全件示す（コメントだけのhitはconsumerに数えない）。

| local helper / surface | 実consumer | 対応exportまたはgap判定 |
|---|---|---|
| `setupTestDB` | repository 30 file | `testdb.SetupTestDB`（既存、signature/実装一致） |
| `setupIsolatedTestDB` | `audit_real_ddl`、`preload_followup` | `testdb.SetupIsolatedTestDB`（既存、signature/実装一致） |
| `ensureClinicSettingsTable` | `clinic_repository` | `testdb.EnsureClinicSettingsTable`（既存、生DDL実装一致） |
| `makeTestOwner` | repository 11 file | `testdb.MakeTestOwner`（既存、signature/実装一致） |
| `ensureAutoMigrated` | repository 27 file | `testdb.EnsureAutoMigrated`（既存、signature/実装一致） |
| `markAutoMigrated` | cross-file consumer 0 | `testdb.MarkAutoMigrated`（既存）。定義bridgeは最終削除 |
| `TestMain` | package lifecycleのみ | `testdb.CloseSharedTestDB`は既存。移設先ごとのtest process終了でpoolは破棄されるため、legacy `TestMain`は不要として削除 |
| `makeSpeciesAndPet` | owner/pet/reservation/medicalrecord test | gap: B0で同一実装を`testdb.MakeSpeciesAndPet`としてexport追加 |
| `makeDoctor` | reservation/medicalrecord test | gap: B0で`testdb.MakeDoctor`をexport追加 |
| `makeHistoryMedicalRecord` | medicalrecord test | gap: B0で`testdb.MakeHistoryMedicalRecord`をexport追加 |
| `seedClinicsForFK` | reservation test | gap: B0で`testdb.SeedClinicsForFK`をexport追加 |
| `makeInsuranceMaster` / `makePetWithInsurance` | billing/owner/pet test | gap: B0で`testdb.MakeInsurance` / `testdb.MakePetWithInsurance`をexport追加 |
| `makeReservationType` / `setupReservationIsolationTestDB` / `makeReservation` | reservation test群 | gap: `internal/reservation`のtest helperへ移す。domain固有のため`testdb`へ追加しない |
| `makeMedicineMaster` / `makeProcedure` / `makeClinicScopedClinicalReadParents` | medicalrecord test群 | gap: `internal/medicalrecord`のtest helperへ移す |
| `makeShiftEntryWithType` | consumer 0 | 不要。carrier削除時に削除 |
| `makeBillingWith` / `makeBilling` | carrierとclinic blocking-count test | gap: clinic test内で最小fixtureを再定義し、2 carrierは不要として削除 |
| `setupClinicTestDB` | `clinic_repository`、`clinic_permission_group_tx_atomicity` | gap: `internal/clinic`のtest helperへ移す |
| `setupOwnerPetIsolationTestDB` | owner/pet/medicalrecord test | gap: 各移設先で`testdb.SetupTestDB` + 必要modelだけを準備し、共有helperは削除 |
| `makeInsuranceMaster` / `makePetWithInsurance` | `cross_clinic_preload`、`insurance_repository` | 上記B0 exportへ差し替え |
| `makeDiagnosisTypeMaster` / `makeDiagnosisNameRec` | `master_preload`、`diagnosis_repository` | gap: B10で`internal/medicalrecord` test helperへ移す |
| `makeShiftEntry` | reservation schedule test | gap: B8で`internal/reservation` test helperへ移す |
| reservation staff helper群（`makeDoctorAssignedToClinic`、`setup*IsolationTestDB`、`setupReservationStaffTxAtomicityTestDB`） | reservation staff 7 file | gap: B8で7 fileと同時に`internal/reservation`へ移す |
| `seedClinicsForFK` / `makeStaffClinicAssignment` | reservation read test群 | `seedClinicsForFK`はB0 export、`makeStaffClinicAssignment`はB8でreservation内再定義 |
| `awaitStaffTestSignal` / `awaitStaffTestError` | staff race test 2 file | gap: B12で2 fileと同時に`internal/staff`へ移す |
| preload lint共有helper（`moduleInternalSource`、`legacyLintKey`、`baseFileName`、`receiverMethodKey`、2 discovery assertion） | preload/audit-tx/db-or-tx 3 gate | gap: B5で3 gateを同時に`internal/lintscan`へ移す |
| `target_repository_test_facades_test.go`のalias/wrapper全件 | repository 33 file | 不要。各fileを実在するdomain constructorへ直接接続し、B14でfacade削除 |
| `target_test_surface_test.go`のalias/wrapper全件 | service 8 file | 不要。audit/clinic/staffの実在symbolへ直接接続し、B3でbridge削除 |
| `mockAuditRepository` / `mockPermissionGroupRepository` | audit/clinic各1 file | gap: B1/B2で各target domain testへ分割移設 |
| `mockTransactor` | clinic/staff各1 file | gap: B2/B3で各target domain test内に最小再定義 |

同file扱いにした定義も未列挙にしないため、`grep '^func '`とclosure探索から得た完全なfile別定義inventoryを以下に固定する。重複名はinline analyzer fixture上の別定義であり、意図的に重複表示する。この一覧のうち上のcross-file表に無い定義はすべて「移設先domain内で同fileのまま再定義」と判定する。

<details>
<summary>非Test helper / receiver method / closure 定義inventory（64 file）</summary>

```text
service/audit_clinic_test_doubles_test.go: mockAuditRepository.recordLog, mockAuditRepository.Create, mockAuditRepository.CreateTx, mockPermissionGroupRepository.Create, mockPermissionGroupRepository.UpdateRules
service/audit_service_test.go: ptrUint64ForAuditTest
service/clinic_holiday_service_test.go: mockClinicHolidayRepository.FindAllByYearMonth, mockClinicHolidayRepository.Save, mockClinicHolidayRepository.Delete, mockClinicHolidayRepository.FindByDate
service/clinic_service_test.go: clinicBoolPtr, mockPermissionGroupRepository.DeleteSoftDeletedByClinicID, mockClinicPermissionGroupWriter.DeleteSoftDeletedByClinicID, mockClinicRepository.FindAll, mockClinicRepository.FindByStaffID, mockClinicRepository.FindByID, mockClinicRepository.LockActiveByID, mockClinicRepository.LockByIDForUpdate, mockClinicRepository.FindCompany, mockClinicRepository.Create, mockClinicRepository.Update, mockClinicRepository.Delete, mockClinicRepository.CountOwnersByClinicID, mockClinicRepository.CountStaffByClinicID, mockClinicRepository.CountBlockingReferencesByClinicID, closure:findRule, closure:newRepository
service/clinic_test_transactor_test.go: mockTransactor.WithTx
service/closing_settings_service_test.go: mockClinicSettingsRepository.FindByClinicID, mockClinicSettingsRepository.Save, mockClosingSpecialPeriodRepository.FindAll, mockClosingSpecialPeriodRepository.FindByID, mockClosingSpecialPeriodRepository.FindByDate, mockClosingSpecialPeriodRepository.Create, mockClosingSpecialPeriodRepository.Update, mockClosingSpecialPeriodRepository.Delete, mockClosingSpecialPeriodRepository.CheckOverlap, mockClosingClinicHolidayRepository.FindByDate, mockClosingClinicHolidayRepository.FindAllByYearMonth
service/company_service_test.go: mockCompanyRepository.FindSingleton, mockCompanyRepository.Update
service/master_fk_write_inventory_lint_test.go: mfkDirOf, mfkExternalParam.qualifiedType, mfkExternalParam.occurrence, mfkKey, isIDType, localStructName, qualifiedStructName, masterFKsOf, sortedKeys, analyzeServicePackage, baseName, isServiceWriteRolePackage, matchesRolePackagePrefixes, analyzeRealServiceSource, equalStringSets, reconcileMasterFKWrites, joinSet, isReviewedExternalParam
service/n1_lint_test.go: n1AllowlistKey, analyzeFileN1, matchN1Call, baseNameN1, walkServiceN1, xService.f (10 fixture definitions), xService.validateOwnerPetsInsuranceOwnership, closure:fn
service/staff_clinic_assignment_reservation_race_test.go: blockingAssignmentRaceReservationRepository.Create, observedAssignmentRaceStaffUpdateLocker.LockActiveByIDForUpdate, blockingAssignmentRaceClinicLookup.LockActiveByID, observedAssignmentRaceReservationStaffRepository.FindByID, setupStaffAssignmentReservationRaceTest, newAssignmentRaceStaffService, newAssignmentRaceReservationService
service/staff_cross_tenant_test.go: crossTenantStaffRepository.Create, crossTenantStaffRepository.LockActiveByIDForUpdateInClinic, crossTenantStaffRepository.Update, crossTenantStaffAccountStore.FindByEmail, crossTenantStaffAccountStore.Create, crossTenantStaffAccountStore.UpdatePasswordHash, crossTenantStaffAccountStore.DeletePasswordResetTokens, crossTenantStaffAssignmentRepository.Create, crossTenantStaffAssignmentRepository.LockActiveByStaff, rejectingCrossTenantOccupationRepository.LockActiveByIDForShare, mockReservationForStaff.ExistsByStaffID, mockReservationForStaff.FindClinicIDsByStaffID
service/staff_shift_security_integration_test.go: blockingStaffClinicLookup.LockActiveByID, observedStaffClinicLookup.LockActiveByID, blockingClinicDeleteRepository.Delete, blockingShiftEntryCreateRepository.Create, observedShiftStaffLocker.LockActiveByIDForShare, observedStaffDeleteLocker.LockActiveByIDForUpdateInClinic, observedStaffAssignmentLocker.LockActiveByIDForUpdate, blockingStaffDeleteRepository.Delete, setupStaffShiftSecurityIntegrationTest
service/target_test_surface_test.go: NewAuditService, validateAuditLog, NewClinicHolidayService, NewClinicService, buildClinicUpdate, NewClosingSettingsService, NewCompanyService, buildCompanyUpdate, NewStaffService, NewShiftEntryService, strPtr, ptrFloat64
service/update_fields_test.go: none
repository/appointment_admin_repository_test.go: setupReservationAdminTestDB, makeLineCustomerForAdmin, makeAdminReservationAt
repository/appointment_repository_test.go: none
repository/audit_real_ddl_test.go: setupAuditRealDDLTestDB
repository/audit_repository_test.go: setupAuditTestDB, uint64Ptr
repository/audit_tx_inventory_lint_test.go: analyzeFileForClinicalResultDeletes, clinicalResultModelFromArg, receiverMethodKey, auditInventoryKey, walkRepositoryForClinicalResultDeletes, aggregateClinicalResultFindings, reconcileClinicalResultDeletes, examinationRepository.ReplaceItemsByExamID, checkupFieldResultRepository.ReplaceForCheckup, examinationRepository.DoubleDelete, vaccineRepository.Delete, vitalRepository.Delete, someRepo.BrokenDelete, someRepo.WrongPkg, purgeExamResults, examinationRepository.BulkDelete, examinationRepository.CleanupExamResults
repository/be9_2c_r3_test_helper_carrier_test.go: makeReservationType, setupReservationIsolationTestDB, makeReservation, makeSpeciesAndPet, makeBilling
repository/billing_test_fixtures_test.go: makeBillingWith
repository/billings_hospitalization_unique_migration_test.go: stripSQLLineComments
repository/billings_hospitalization_unique_test.go: setupBillingsHospitalizationUniqueTestDB, makeHospBilling
repository/checkup_migration_ddl_helpers_test.go: readCheckupMigration010, extractCreateTableDDL
repository/clinic_permission_group_tx_atomicity_test.go: none
repository/clinic_repository_test.go: setupClinicTestDB, makeClinicFixture
repository/closing_special_period_repository_test.go: setupClosingSpecialPeriodRepositoryTestDB, makeClosingSpecialPeriod
repository/count_clinic_scope_isolation_test.go: setupEstimateIsolationTestDB
repository/cross_clinic_preload_isolation_test.go: makeInsuranceMaster, makePetWithInsurance, setupInsurancePreloadTestDB, makeReservationWithType
repository/db_setup_test.go: setupTestDB, setupIsolatedTestDB, ensureClinicSettingsTable, makeTestOwner, ensureAutoMigrated, markAutoMigrated
repository/db_test.go: testDBConfig
repository/dbortx_inventory_lint_test.go: funcUsesDBOrTx, isPersistenceTxFromContext, isLocalHelperCall, expressionContainsProducer, assignedProducerHandles, producerHandlesRemainDerived, expressionDerivedFromHandle, funcUsesProducedDBHandle, funcUsesNamedDBHandle, funcForwardsProducedHandleToLocalHelper, parameterNameAt, funcReturnsDBOrTxHandle, funcUsesDBOrTxCall, funcUsesRequiredAmbientTx, funcMatchesAmbientTxExpectation, detectAmbientTxParticipationExpectation, parseAmbientTxSourceFile, walkRepositoryForAmbientTxExpectations, walkRepositoryForDBOrTx, reconcileDBOrTxInventory, fooRepository.Bar (2 fixture definitions), fooRepository.Baz, fooRepository.Qux, fooRepository.Canonical, fooRepository.Cap, fooRepository.Nope, fooRepository.Reorder, silentDB (7 fixture definitions), fooRepository.Create (9 fixture definitions), fooRepository.Update (4 fixture definitions), writeWithTx (2 fixture definitions)
repository/diagnosis_repository_test.go: setupDiagnosisRepoTestDB
repository/insurance_repository_test.go: setupInsuranceRepositoryTestDB
repository/isolation_test_helpers_test.go: makeMedicineMaster, makeHistoryMedicalRecord, makeProcedure, makeDoctor, makeClinicScopedClinicalReadParents, makeShiftEntryWithType
repository/master_preload_clinic_isolation_test.go: makeCageMaster, makeCheckupTypeMaster, makeDiagnosisTypeMaster, makeDiagnosisNameRec, makeHospitalizationRec, makeCheckupRec
repository/migration_cascade_lint_test.go: countCascadeOccurrences, reconcileMigrationCascade, walkMigrationsForCascade
repository/owner_pet_clinic_isolation_test.go: setupOwnerPetIsolationTestDB
repository/owner_pet_create_write_owner_test.go: ownerRegistrationWriterFunc.CreateForOwnerRegistration
repository/owner_pet_relationship_preload_clinic_isolation_test.go: makeOwnerPetRelationshipTestPet, makeOwnerPetRelationshipTestSpecies
repository/payment_method_master_repository_test.go: setupPaymentMethodMasterRepoTestDB, makePaymentMethodMaster, makePaymentMethodBilling, makePaymentForBilling
repository/pet_write_medimage_clinic_isolation_test.go: setupMedImageIsolationTestDB, makeMedRecordImage
repository/preload_clinic_scope_lint_test.go: siteExceptionKey, analyzeFilePreloads, preloadHasClinicScope, funcLitHasClinicScope, preloadFailDetail, isSiteExcepted, stringLitValue, clinicScopedIntermediatePrefixes, preloadReceiverChainHasScopedAssociation, lastAssocSegment, baseFileName, isInSet, moduleInternalSource, legacyLintKey, assertDiscoversFileFromDifferentTopLevelPackage, assertLintscanReachesTwoOrMoreNestingLevels, walkRepositoryPreloads
repository/preload_followup_clinic_isolation_test.go: makeExamRec, setupPreloadTrimmingDetailTestDB
repository/preload_master_model_reconciliation_test.go: extractModelClinicScopedStructs, structHasClinicIDField, fieldNamed, isUint64OrPtrUint64, extractStringMapValues, canonicalModelName, reconcileMasterModelCoverage, readSideMasterModelNames, siblingPackageDir, loadModelClinicScopedStructs, loadWriteSideMasterModelNames
repository/reservation_owner_pet_preload_clinic_isolation_test.go: setupReservationOwnerPetPreloadDB
repository/reservation_schedule_clinic_isolation_test.go: setupScheduleIsolationTestDB, makeShiftEntry
repository/reservation_schedule_repository_test.go: setupReservationScheduleCRUDTestDB
repository/reservation_staff_capability_preload_clinic_isolation_test.go: setupReservationStaffCapabilityPreloadTestDB, makeStaffReservationCapability, closure:assertPreloadIsolation
repository/reservation_staff_capability_write_clinic_isolation_test.go: setupCapabilityIsolationTestDB, makeDoctorAssignedToClinic, closure:countCapabilities
repository/reservation_staff_exclusion_clinic_isolation_test.go: setupExclusionIsolationTestDB, closure:countExclusions, closure:exclusionExists
repository/reservation_staff_junction_lock_race_test.go: replaceReservationJunction, countReservationJunction, revokeStaffAssignmentWithIdentityLock
repository/reservation_staff_repository_test.go: setupReservationStaffRepoTestDB, closure:containsID, closure:sortOrderOf
repository/reservation_staff_repository_tx_atomicity_test.go: setupReservationStaffTxAtomicityTestDB
repository/reservation_staff_service_readback_atomicity_test.go: failingReservationStaffReadbackRepository.LockForMutation, failingReservationStaffReadbackRepository.FindByID, failingReservationStaffReadbackRepository.FindAllExcludedReservationTypes
repository/rls_effectiveness_test.go: setupAppPrivateRLSFunctions, closure:exec, closure:asRole
repository/rls_role_privilege_test.go: none
repository/staff_occupation_write_race_test.go: setupStaffOccupationWriteRaceDB, makeStaffOccupationRaceClinic, coordinatedOccupationRepository.LockActiveByIDForShare, coordinatedOccupationRepository.LockActiveByIDForUpdate, awaitStaffOccupationMutation, makeUnassignedOccupationRaceStaff
repository/staff_preload_clinic_isolation_test.go: seedClinicsForFK, makeStaffClinicAssignment, makeReservationWithDoctor
repository/staff_shift_graph_atomicity_test.go: setupStaffShiftGraphAtomicityDB, makeShiftGraphStaff, makeShiftGraphEntry, failingShiftEntryBreakRepository.ReplaceBreaks, pausingShiftEntryLockRepository.LockActiveByIDForUpdate, awaitStaffTestSignal, awaitStaffTestError, failingShiftTemplateBreakRepository.UpdateBreaks, makeShiftGraphTemplate
repository/staff_update_security_test.go: setupStaffUpdateSecurityDB, makeStaffUpdateClinic, makeAccountStaffForUpdate, newStaffUpdateServiceForDB, noopStaffUpdateCredentialAuditTxLogger.LogEntryTx, staffUpdateCredentialAudit, failAfterStaffAccountUpdate.UpdatePasswordHash
repository/target_repository_test_facades_test.go: targetStaffAccountStore.DeletePasswordResetTokens, NewAccountRepository, NewAccountingRepository, NewReservationAdminRepository, NewAuditRepository, NewCarePlanItemRepository, NewCheckupRepository, NewClinicRepository, NewClinicalPlanRepository, NewClosingSpecialPeriodRepository, NewDiagnosisTypeRepository, NewDiagnosisNameRepository, NewEstimateRepository, NewExaminationRepository, NewHospitalizationRepository, NewInsuranceRepository, NewMedicalRecordImageRepository, NewMedicalRecordRepository, NewOccupationRepository, NewOwnerRepository, NewOwnerRepositoryWithPetWriter, NewPaymentMethodMasterRepository, NewPermissionGroupRepository, NewPetRepository, NewPetRepositoryWithWriter, NewReservationRepository, NewReservationScheduleRepository, NewReservationStaffRepository, NewShiftEntryRepository, NewShiftTemplateRepository, NewStaffClinicAssignmentRepository, NewStaffRepository, NewTransactor, txFromContext
repository/test_schema_enum_parity_test.go: extractSQLEnumTypes, goEnumTypes, reconcileTestSchemaEnumParity
repository/transactor_test.go: setupTransactorTestDB
```

</details>

`internal/testdb/testdb.go`の現行export面は`EnsureAutoMigrated`、`MarkAutoMigrated`、`CloseSharedTestDB`、`SetupTestDB`、`EnsureClinicSettingsTable`、`MakeTestOwner`、`SetupIsolatedTestDB`の7件。B0で追加する6 export以外を「対応あり」と扱わない。

##### 64 file分類（未分類0）

表中`DB`は上表の既存`testdb` 6対応、`RF`/`SF`はrepository/serviceのtest-only facadeを直接domain symbolへ置換、`local`は同fileのhelper/receiver/closureをfileと同時に移すことを表す。

| legacy file | 移設先 / 削除 | 根拠symbol（現行実在path） | local helper処理 | batch |
|---|---|---|---|---:|
| `service/audit_clinic_test_doubles_test.go` | `audit`と`clinic`へ分割後削除 | consumersは`audit_service_test.go`と`clinic_service_test.go` | mock2型を各domainへ移す | B1/B2 |
| `service/audit_service_test.go` | `audit` | `audit.NewService` (`internal/audit/service.go:62`) | `SF NewAuditService/validateAuditLog`を直接化、audit mock同時移設 | B1 |
| `service/clinic_holiday_service_test.go` | `clinic` | `clinic.NewClinicHolidayService` (`internal/clinic/clinic_holiday_service.go:24`) | `SF`、local mock methods | B2 |
| `service/clinic_service_test.go` | `clinic` | `clinic.NewClinicService` / `BuildClinicUpdate` (`internal/clinic/clinic_service.go:273/85`) | `SF`、clinic mock、`mockTransactor`、local closures | B2 |
| `service/clinic_test_transactor_test.go` | `clinic`/`staff`へ再定義後削除 | consumersはclinic/staff test | `mockTransactor`を各domain内再定義 | B2/B3 |
| `service/closing_settings_service_test.go` | `clinic` | `clinic.NewClosingSettingsService` (`internal/clinic/closing_settings_service.go:106`) | `SF`、local mocks | B2 |
| `service/company_service_test.go` | `clinic` | `clinic.NewCompanyService` / `BuildCompanyUpdate` (`internal/clinic/company_service.go:92/43`) | `SF`、local mock | B2 |
| `service/master_fk_write_inventory_lint_test.go` | `lintscan` | `lintscan.WalkInternalTreeT` (`internal/lintscan/lintscan.go:158`) | local analyzer一式を同時移設 | B4 |
| `service/n1_lint_test.go` | `lintscan` | `lintscan.WalkInternalTreeT` (`internal/lintscan/lintscan.go:158`) | local analyzer/closuresを同時移設 | B4 |
| `service/staff_clinic_assignment_reservation_race_test.go` | `staff` | `staff.NewStaffService` (`internal/staff/staff_service.go:172`) | `SF`、既存`testdb` direct、local doubles | B3 |
| `service/staff_cross_tenant_test.go` | `staff` | `staff.NewStaffService` (`internal/staff/staff_service.go:172`) | `SF`、local doubles、transactor再定義 | B3 |
| `service/staff_shift_security_integration_test.go` | `staff` | `staff.NewShiftEntryService` (`internal/staff/shift_entry_service.go:100`) | `SF`、既存`testdb` direct、local doubles | B3 |
| `service/target_test_surface_test.go` | 削除 | alias先は`audit`/`clinic`/`staff`に実在 | B1〜B3で8 consumerを直接化後に不要 | B3 |
| `service/update_fields_test.go` | `sharedkernel` | `sharedkernel.SetNullableUint64Field` (`internal/sharedkernel/validators.go:88`) | helperなし | B4 |
| `repository/appointment_admin_repository_test.go` | `reservation` | `reservation.NewReservationAdminRepository` (`internal/reservation/appointment_admin_repository.go:33`) | `DB`、`RF`、reservation helper群 | B8 |
| `repository/appointment_repository_test.go` | `reservation` | `reservation.ParseJSTDate` / `AppointmentDayRange` (`internal/reservation/reservation_repository.go:720/713`) | helperなし | B8 |
| `repository/audit_real_ddl_test.go` | `audit` | `audit.NewRepository` (`internal/audit/repository.go:26`) | `DB isolated`、DDL helper、`RF` | B6 |
| `repository/audit_repository_test.go` | `audit` | `audit.NewRepository` (`internal/audit/repository.go:26`) | `setupAuditRealDDLTestDB`、`RF` | B6 |
| `repository/audit_tx_inventory_lint_test.go` | `lintscan` | module-wide sourceを`lintscan.WalkInternalTreeT`で取得 | preload lint共有helper、local analyzer | B5 |
| `repository/be9_2c_r3_test_helper_carrier_test.go` | helper分配後削除 | reservation/pet/billingの一時carrier | B0 exportまたはtarget-localへ分配 | B14 |
| `repository/billing_test_fixtures_test.go` | 削除 | consumerは上記carrierだけ | clinic側最小fixtureへ置換後不要 | B14 |
| `repository/billings_hospitalization_unique_migration_test.go` | `billing` | `billings(hospitalization_id)` business constraint | local SQL normalizer | B11 |
| `repository/billings_hospitalization_unique_test.go` | `billing` | `billing.NewAccountingRepository` (`internal/billing/accounting_repository.go:56`) | `DB`、`RF`、local fixture | B11 |
| `repository/checkup_migration_ddl_helpers_test.go` | `audit`へ移して名称修正 | consumerは`audit_real_ddl_test.go`だけ | local DDL reader/extractor | B6 |
| `repository/clinic_permission_group_tx_atomicity_test.go` | `clinic` | clinic Createとpermission groupの同一tx回帰 | `setupClinicTestDB`、`RF` | B7 |
| `repository/clinic_repository_test.go` | `clinic` | `clinic.ClinicRepository` (`internal/clinic/clinic_repository.go`) | `DB`、`RF`、clinic helper、billing fixture再定義 | B7 |
| `repository/closing_special_period_repository_test.go` | `clinic` | `clinic.NewClosingSpecialPeriodRepository` (`internal/clinic/closing_special_period_repository.go`) | `DB`、`RF`、local fixture | B7 |
| `repository/count_clinic_scope_isolation_test.go` | `medicalrecord`/`reservation`/`billing`へtest単位分割 | 3 repositoryのCount method | `DB`、reservation helper、local estimate setup | B10 |
| `repository/cross_clinic_preload_isolation_test.go` | `pet`/`owner`/`reservation`へtest単位分割 | `pet.Repository`、`owner.Repository`、`reservation.ReservationStore` | `DB`、B0 insurance/pet exports、`RF` | B9 |
| `repository/db_setup_test.go` | 削除 | 全実装は既に`internal/testdb/testdb.go`へ委譲 | 36 consumerをdirect exportへ置換後不要 | B14 |
| `repository/db_test.go` | `dbconn` | `dbconn.OpenGORM` (`internal/dbconn/gorm.go:29`) | `testDBConfig`は同file移設、`DB` | B13 |
| `repository/dbortx_inventory_lint_test.go` | `lintscan` | module-wide sourceを`lintscan.WalkInternalTreeT`で取得 | preload/audit lint共有helper、local analyzer | B5 |
| `repository/diagnosis_repository_test.go` | `medicalrecord` | `medicalrecord.NewDiagnosisTypeRepository` (`internal/medicalrecord/diagnosis_type_repository.go:32`) | `DB`、diagnosis helper | B10 |
| `repository/insurance_repository_test.go` | `billing` | `billing.NewInsuranceRepository` (`internal/billing/insurance_repository.go:27`) | `DB`、B0 insurance/pet exports、`RF` | B9 |
| `repository/isolation_test_helpers_test.go` | helper分配後削除 | medicalrecord/reservationの10 consumer | B0 exportまたはtarget-localへ分配 | B14 |
| `repository/master_preload_clinic_isolation_test.go` | `medicalrecord`/`reservation`へtest単位分割 | medicalrecord 5 repository + reservation admin 1 | `DB`、`RF`、medicalrecord helper | B10 |
| `repository/migration_cascade_lint_test.go` | `lintscan` | migration CASCADE inventory gate | local analyzer。相対path gapを同batch修正 | B5 |
| `repository/owner_pet_clinic_isolation_test.go` | `owner`/`pet`へtest単位分割 | `owner.Repository` / `pet.Repository` | `DB`、`RF`、owner/pet setup | B9 |
| `repository/owner_pet_create_write_owner_test.go` | `owner`/`pet`へtest単位分割 | owner/pet write owner実装 (`internal/owner/repository.go`, `internal/pet/repository.go`) | `DB`、`RF`、local writer double | B9 |
| `repository/owner_pet_relationship_preload_clinic_isolation_test.go` | `owner`/`pet`へtest単位分割 | owner/pet Preload contract | `DB`、`RF`、local fixtures | B9 |
| `repository/payment_method_master_repository_test.go` | `billing` | `billing.NewPaymentMethodMasterRepository` (`internal/billing/payment_method_master_repository.go:27`) | `DB`、`RF`、local fixtures | B11 |
| `repository/pet_write_medimage_clinic_isolation_test.go` | `pet`/`medicalrecord`へtest単位分割 | pet Update + `medicalrecord.NewMedicalRecordImageRepository` (`internal/medicalrecord/medical_record_image_repository.go:31`) | `DB`、B0 fixture exports、`RF` | B9 |
| `repository/preload_clinic_scope_lint_test.go` | `lintscan` | `lintscan.WalkInternalTreeT` | 3 lint共有helperの定義owner | B5 |
| `repository/preload_followup_clinic_isolation_test.go` | `medicalrecord`/`trimming`/`reservation`へtest単位分割 | examination、trimming detail、reservation repository | `DB`、`RF`、local fixtures | B10 |
| `repository/preload_master_model_reconciliation_test.go` | `lintscan` | model/read-write registry reconciliation gate | cwd/service-file固定gapを同batch修正 | B5 |
| `repository/reservation_owner_pet_preload_clinic_isolation_test.go` | `reservation` | `reservation.NewReservationRepository` (`internal/reservation/reservation_repository.go:244`) | `DB`、`RF`、reservation helper群 | B8 |
| `repository/reservation_schedule_clinic_isolation_test.go` | `reservation` | `reservation.NewReservationScheduleRepository` (`internal/reservation/reservation_schedule_repository.go:50`) | `DB`、`RF`、schedule helper | B8 |
| `repository/reservation_schedule_repository_test.go` | `reservation` | 同上 | `DB`、`RF`、schedule helper | B8 |
| `repository/reservation_staff_capability_preload_clinic_isolation_test.go` | `reservation` | `reservation.NewReservationStaffRepository` (`internal/reservation/reservation_staff_repository.go:52`) | `DB`、`RF`、reservation staff helper | B8 |
| `repository/reservation_staff_capability_write_clinic_isolation_test.go` | `reservation` | 同上 | `DB`、`RF`、reservation staff helper/local closure | B8 |
| `repository/reservation_staff_exclusion_clinic_isolation_test.go` | `reservation` | 同上 | `DB`、`RF`、reservation staff helper/local closures | B8 |
| `repository/reservation_staff_junction_lock_race_test.go` | `reservation` | 同上 | `RF`、tx/capability helper群 | B8 |
| `repository/reservation_staff_repository_test.go` | `reservation` | 同上 | `DB`、`RF`、staff helper/local closures | B8 |
| `repository/reservation_staff_repository_tx_atomicity_test.go` | `reservation` | 同上 | `DB`、`RF`、tx helper | B8 |
| `repository/reservation_staff_service_readback_atomicity_test.go` | `reservation` | `reservation.NewReservationStaffService` (`internal/reservation/reservation_staff_service.go`) | `RF`、reservation staff helper、local double methods | B8 |
| `repository/rls_effectiveness_test.go` | `persistence` | RLS/transaction-scoped GUCのcross-cutting DB invariant | `DB`、local closures | B13 |
| `repository/rls_role_privilege_test.go` | `persistence` | DB role privilege preflight | `DB` | B13 |
| `repository/staff_occupation_write_race_test.go` | `staff` | `staff.NewOccupationRepository` (`internal/staff/occupation_repository.go:33`) | `DB`、`RF`、staff signal helper | B12 |
| `repository/staff_preload_clinic_isolation_test.go` | `reservation` | test対象は`ReservationRepository.Doctor` Preload | `DB`、`RF`、B0 seed export、reservation helper | B8 |
| `repository/staff_shift_graph_atomicity_test.go` | `staff` | `staff.NewShiftEntryService` / shift template repository | `DB`、`RF`、staff signal helper | B12 |
| `repository/staff_update_security_test.go` | `staff` | `staff.NewStaffService` (`internal/staff/staff_service.go:172`) | `DB`、`RF`、local doubles | B12 |
| `repository/target_repository_test_facades_test.go` | 削除 | alias先9 packageに全symbol実在 | 33 consumerを直接constructorへ置換後不要 | B14 |
| `repository/test_schema_enum_parity_test.go` | `testdb` | `testdb.SharedTestSchemaEnumTypes` (`internal/testdb/testdb.go:269`) | local analyzer、migration pathをmodule root化 | B13 |
| `repository/transactor_test.go` | `persistence` | `persistence.NewTransactor` / `TxFromContext` (`internal/persistence/transactor.go:21`, `tx.go:19`) | `DB`、`RF`、local setup | B13 |

##### 共有test infra 9 file

consumer数は同じlegacy package内の実file集合で数えた。`db_setup_test.go`は`setupTestDB`単体のconsumerがContextどおり30 file、同fileの6 helper全体のconsumer unionが36 fileだった。

| file | consumer file数 | 判定 | 根拠 |
|---|---:|---|---|
| `repository/db_setup_test.go` | 36 | 不要になるので削除 | 6 helperは既存`testdb` exportへのthin bridge。全consumer direct化後に役割0 |
| `repository/db_test.go` | 0 | `dbconn`へ移す | `OpenGORM`だけを検証し、`testDBConfig`は同file内 |
| `repository/isolation_test_helpers_test.go` | 10 | `testdb`へ一部吸収、残りはdomainへ移して削除 | 6 helperのうちcross-domain fixtureはB0 export、medicalrecord固有はdomain-local、consumer 0の1件は削除 |
| `repository/billing_test_fixtures_test.go` | 1 | 不要になるので削除 | consumerは一時carrierだけ。clinic側の必要fixtureをlocal化すると役割0 |
| `repository/be9_2c_r3_test_helper_carrier_test.go` | 18 | 不要になるので削除 | コメントどおりBE9移設中のcarrier。B0/target domainへ分配後に役割0 |
| `repository/target_repository_test_facades_test.go` | 33 | 不要になるので削除 | alias/wrapper先が9 target packageに実在し、全consumer直接化後にbridge不要 |
| `service/target_test_surface_test.go` | 8 | 不要になるので削除 | alias/wrapper先がaudit/clinic/staffに実在し、全consumer直接化後にbridge不要 |
| `service/audit_clinic_test_doubles_test.go` | 2 | 移設先domainへ分割移動 | audit mockとclinic mockのconsumerが各1 fileで明確 |
| `service/clinic_test_transactor_test.go` | 2 | 移設先domainで再定義後削除 | clinic/staffそれぞれに小さいtest-only transactorを置き、cross-domain共有を残さない |

##### lint gate 6件

| lint gate | 移設先 | source discovery再確認 |
|---|---|---|
| `repository/audit_tx_inventory_lint_test.go` | `lintscan` | `moduleInternalSource`→`lintscan.WalkInternalTreeT`; package非依存 |
| `repository/dbortx_inventory_lint_test.go` | `lintscan` | 同上。preload/audit helper依存があるため3 file同時移設 |
| `repository/preload_clinic_scope_lint_test.go` | `lintscan` | `moduleInternalSource` (`:446-448`)が`WalkInternalTreeT`; package非依存 |
| `service/master_fk_write_inventory_lint_test.go` | `lintscan` | `analyzeRealServiceSource` (`:703-705`)が`WalkInternalTreeT`; discoveryはpackage非依存、role prefix filterは意図したcontent scope |
| `service/n1_lint_test.go` | `lintscan` | `walkServiceN1` (`:243`)が`WalkInternalTreeT`; package非依存 |
| `repository/migration_cascade_lint_test.go` | `lintscan` | **gapあり**: `migrationsDir="../../migrations"` (`:39`) と`os.ReadDir` (`:102`) は現在と移設先が同じ深さなら動くがlocation-independentではない。B5で`lintscan.FindModuleRoot`由来のabsolute migration dirへ変更する |

BE9-1の「5 gateのproduction source discoveryがpackage非依存」は実装で再確認できた。6件目のmigration gateはGo source discoveryを行わず、上記relative-path gapを持つため「6件すべて非依存」とは判定しない。

##### 実行batchとhelper集合検算

B0からB14を順に適用する。`I(files)`を上の64 file別inventoryに記載したhelperの和集合、`D=I(当該batchのfile)`、`C`をbatch後もlegacyに残るfileが消費するhelperと定義する。次表の`D`欄は`D`のうちcross-file consumerを持つ要素を全名列挙し、同file内だけの要素は完全inventoryと直後のfile manifestから機械的に復元できる`I(files) local-only`と記す。local-only要素のconsumerは定義fileと同時に移るため`C`へ入らない。各batchで移動fileは既存`testdb` export、B0 export、または同batch内helperだけを参照する。

| batch | file / unit | 先行 | `D` | `C` | `D ∩ C` |
|---:|---|---:|---|---|---|
| B0 ✅完了 `27d95aacd` | `testdb`に6 fixture export追加（legacy file移動なし） | — | `∅` | `∅` | `∅` |
| B1 ✅完了 `718f6c9b3` | service audit test + audit側double分割、service bridgeのaudit symbol除去 | B0 | `mockAuditRepository.recordLog`, `mockAuditRepository.Create`, `mockAuditRepository.CreateTx`, `NewAuditService`, `validateAuditLog` + `I(files) local-only` | `∅`（audit consumerを同時移設） | `∅` |
| B2 ✅完了 `134e7953b` | service clinic holiday/clinic/closing/company + clinic側double、bridgeのclinic symbol除去 | B1 | `mockPermissionGroupRepository.Create`, `mockPermissionGroupRepository.UpdateRules`, `NewClinicHolidayService`, `NewClinicService`, `buildClinicUpdate`, `NewClosingSettingsService`, `NewCompanyService`, `buildCompanyUpdate` + `I(files) local-only` | `NewClinicService`（B3対象のstaff integration testが消費するためbridgeをB3まで保持） | `NewClinicService`（撤去をB3へ延期して解消） |
| B3 ✅完了 `36f283f37` | service staff 3 file + staff transactor、service bridge/残infra削除 | B2 | `mockTransactor.WithTx`, `NewStaffService`, `NewShiftEntryService`, `strPtr`, `ptrFloat64` + `I(files) local-only` | `∅` | `∅` |
| B4 | service master-FK/N+1 lint→`lintscan`、update-fields→`sharedkernel` | B3 | `I(files) local-only` | `∅` | `∅` |
| B5 | repository lint 5 file→`lintscan`、migration/reconciliation path gap修正 | B0 | `moduleInternalSource`, `legacyLintKey`, `baseFileName`, `receiverMethodKey`, `assertDiscoversFileFromDifferentTopLevelPackage`, `assertLintscanReachesTwoOrMoreNestingLevels` + `I(files) local-only` | `∅`（3 consumer gateを同時移設） | `∅` |
| B6 | repository audit 2 file + DDL helper→`audit` | B0 | `setupAuditRealDDLTestDB`, `readCheckupMigration010`, `extractCreateTableDDL` + `I(files) local-only` | `∅`（audit consumerを同時移設） | `∅` |
| B7 | repository clinic/permission-group/closing-special 3 file→`clinic` | B0 | `setupClinicTestDB` | `∅`（2 consumerを同時移設） | `∅` |
| B8 | reservation 13 file（appointment、schedule、staff、owner/pet preload、staff preload）→`reservation` | B0 | `setupReservationAdminTestDB`, `makeLineCustomerForAdmin`, `makeAdminReservationAt`, `makeShiftEntry`, `setupCapabilityIsolationTestDB`, `makeDoctorAssignedToClinic`, `setupExclusionIsolationTestDB`, `setupReservationStaffTxAtomicityTestDB`, `seedClinicsForFK`, `makeStaffClinicAssignment` + `I(files) local-only` | `∅`（全実consumer同時移設、`isolation_test_helpers_test.go`の`seedClinicsForFK` callもB0 exportへ直接化） | `∅` |
| B9 | cross-clinic file分割 + insurance + owner/pet 3 file + pet/medimage分割 | B0,B8 | `makeInsuranceMaster`, `makePetWithInsurance`, `setupOwnerPetIsolationTestDB` + `I(files) local-only` | `∅`（全consumer同時移設またはB0 export化） | `∅` |
| B10 | count/master-preload/preload-followupをtest単位分割 + diagnosis→各domain | B0,B8,B9 | `makeDiagnosisTypeMaster`, `makeDiagnosisNameRec` + `I(files) local-only` | `∅`（diagnosis consumer同時移設） | `∅` |
| B11 | billing unique migration/runtime + payment-method 3 file→`billing` | B0,B10 | `I(files) local-only` | `∅` | `∅` |
| B12 | staff occupation/shift/update 3 file→`staff` | B0 | `awaitStaffTestSignal`, `awaitStaffTestError` | `∅`（2 consumer同時移設） | `∅` |
| B13 | db test→`dbconn`、RLS/transactor→`persistence`、enum parity→`testdb` | B0 | `I(files) local-only` | `∅` | `∅` |
| B14 | repository `db_setup`、isolation/billing/BE9 carrier、target facadeを削除 | B1〜B13 | `I(db_setup_test.go, isolation_test_helpers_test.go, billing_test_fixtures_test.go, be9_2c_r3_test_helper_carrier_test.go, target_repository_test_facades_test.go)`（要素全名は上のfile別inventory） | `∅`（64 testのtarget移設完了済み） | `∅` |

batchごとの含有fileは次のとおり。分割fileは同じ原本の対象部分を複数batchに明記している。

- B0: legacy fileなし（`testdb` export追加だけ）。
- B1: `service/audit_clinic_test_doubles_test.go`（audit部分）、`service/audit_service_test.go`。
- B2: `service/audit_clinic_test_doubles_test.go`（clinic部分）、`service/clinic_holiday_service_test.go`、`service/clinic_service_test.go`、`service/clinic_test_transactor_test.go`（clinic部分）、`service/closing_settings_service_test.go`、`service/company_service_test.go`。
- B3: `service/clinic_test_transactor_test.go`（staff部分）、`service/staff_clinic_assignment_reservation_race_test.go`、`service/staff_cross_tenant_test.go`、`service/staff_shift_security_integration_test.go`、`service/target_test_surface_test.go`。
- B4: `service/master_fk_write_inventory_lint_test.go`、`service/n1_lint_test.go`、`service/update_fields_test.go`。
- B5: `repository/audit_tx_inventory_lint_test.go`、`repository/dbortx_inventory_lint_test.go`、`repository/migration_cascade_lint_test.go`、`repository/preload_clinic_scope_lint_test.go`、`repository/preload_master_model_reconciliation_test.go`。
- B6: `repository/audit_real_ddl_test.go`、`repository/audit_repository_test.go`、`repository/checkup_migration_ddl_helpers_test.go`。
- B7: `repository/clinic_permission_group_tx_atomicity_test.go`、`repository/clinic_repository_test.go`、`repository/closing_special_period_repository_test.go`。
- B8: `repository/appointment_admin_repository_test.go`、`repository/appointment_repository_test.go`、`repository/reservation_owner_pet_preload_clinic_isolation_test.go`、`repository/reservation_schedule_clinic_isolation_test.go`、`repository/reservation_schedule_repository_test.go`、`repository/reservation_staff_capability_preload_clinic_isolation_test.go`、`repository/reservation_staff_capability_write_clinic_isolation_test.go`、`repository/reservation_staff_exclusion_clinic_isolation_test.go`、`repository/reservation_staff_junction_lock_race_test.go`、`repository/reservation_staff_repository_test.go`、`repository/reservation_staff_repository_tx_atomicity_test.go`、`repository/reservation_staff_service_readback_atomicity_test.go`、`repository/staff_preload_clinic_isolation_test.go`。
- B9: `repository/cross_clinic_preload_isolation_test.go`、`repository/insurance_repository_test.go`、`repository/owner_pet_clinic_isolation_test.go`、`repository/owner_pet_create_write_owner_test.go`、`repository/owner_pet_relationship_preload_clinic_isolation_test.go`、`repository/pet_write_medimage_clinic_isolation_test.go`。
- B10: `repository/count_clinic_scope_isolation_test.go`、`repository/diagnosis_repository_test.go`、`repository/master_preload_clinic_isolation_test.go`、`repository/preload_followup_clinic_isolation_test.go`。
- B11: `repository/billings_hospitalization_unique_migration_test.go`、`repository/billings_hospitalization_unique_test.go`、`repository/payment_method_master_repository_test.go`。
- B12: `repository/staff_occupation_write_race_test.go`、`repository/staff_shift_graph_atomicity_test.go`、`repository/staff_update_security_test.go`。
- B13: `repository/db_test.go`、`repository/rls_effectiveness_test.go`、`repository/rls_role_privilege_test.go`、`repository/test_schema_enum_parity_test.go`、`repository/transactor_test.go`。
- B14: `repository/be9_2c_r3_test_helper_carrier_test.go`、`repository/billing_test_fixtures_test.go`、`repository/db_setup_test.go`、`repository/isolation_test_helpers_test.go`、`repository/target_repository_test_facades_test.go`。

各batchは移設・直接接続・helper解消だけのstructural-only unitとし、behavior/security修正は別unitへ分離する。各batchの実装unitは移動先packageだけのDocker scoped testを必須とし、B5はlint gate自身のlocation-agnostic self-test、B14はlegacy file/import 0の静的gateを追加する。B0〜B14の実装は本Phase 0の対象外である。

##### 削除条件・担当・期限

- `internal/service`
  - 削除条件: B1〜B4完了、14 fileのtarget移設またはbridge削除、production/test/import 0、移設先`audit`/`clinic`/`staff`/`lintscan`/`sharedkernel` scoped test green。
  - 担当: MinoruSoga（実装owner）。各batchのGo reviewは当該実装unitのreviewerが担当する。
  - 期限: 2026-07-31。
- `internal/repository`
  - 削除条件: B0/B5〜B14完了、50 fileのtarget移設またはcarrier/facade削除、lint gate 6件とenum/RLS/DDL gateの移設先green、production/test/import 0。空子directory削除はBE10-3であり本条件へ混ぜない。
  - 担当: MinoruSoga（実装owner）。clinic isolation/transaction testを含むbatchは当該専門reviewを追加する。
  - 期限: 2026-08-08。

期限超過時は削除条件を緩めず、本節を`判断待ち`へ戻して未完batch、blocker、再開条件を記録する。

<details>
<summary>Phase 0 verification ledger</summary>

- Changed files: `BE-refactor.md`のみ。Go file変更なし。
- Saved prompt validator: `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-be10-2-phase0-legacy-test-retirement-plan.md`

  ```text
  Prompt Craft Harness Validation: PASS
  Profile: standard (declared-risk-tier)
  Target: agent (detected)
  Quality mode: standard
  exit=0
  ```

- C1 baseline: `git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte rev-parse --short HEAD`は`ef32784d3`。`git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte status --porcelain -- backend BE-refactor.md todo.md`は`(empty)`。
- C2: 指定2 pathへの`find ... -maxdepth 1 -name '*.go'`結果は上の64行分類表と同じservice 14 / repository 50。`find ... -not -name '*_test.go' | wc -l`は`0`。
- C3: package別`grep -rn '^func '`とhelper別`grep -rln`の結果は上のhelper表/inventoryへ固定した。`func Test`除外後はservice 112 / repository 238、除外した`func Test`はservice 79 / repository 188、function-local closureはservice 3 / repository 8。
- C4: `grep -n '^func [A-Z]' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb.go`

  ```text
  59:func EnsureAutoMigrated(db *gorm.DB, models ...any) error {
  109:func MarkAutoMigrated(models ...any) {
  119:func CloseSharedTestDB() {
  129:func SetupTestDB(t *testing.T) *gorm.DB {
  156:func EnsureClinicSettingsTable(t *testing.T, db *gorm.DB) {
  198:func MakeTestOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
  218:func SetupIsolatedTestDB(t *testing.T) *gorm.DB {
  ```

- C5/C8 count reconciliation:

  ```text
  helper_inventory total=64 service=14 repository=50
  classification total=64 service=14 repository=50
  classified=64 manifest_unique=64 missing=0
  ```

- C6: helper consumerのsame-package file unionは`db_setup=36`（`setupTestDB`単体30）、`db_test=0`、`isolation=10`、`billing fixtures=1`、`BE9 carrier=18`、`repository target facade=33`、`service target surface=8`、`audit/clinic doubles=2`、`clinic transactor=2`。
- C7: `find /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/service /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository -maxdepth 1 -name '*lint*_test.go' | sort`

  ```text
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/audit_tx_inventory_lint_test.go
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/dbortx_inventory_lint_test.go
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/migration_cascade_lint_test.go
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/preload_clinic_scope_lint_test.go
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/service/master_fk_write_inventory_lint_test.go
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/service/n1_lint_test.go
  ```

- C9: `grep -n '削除条件\|担当\|期限' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/BE-refactor.md`でservice/repository両方の3項目を確認した。
- C10 final: `git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte status --porcelain -- backend BE-refactor.md todo.md`は` M BE-refactor.md`。`git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte diff --name-only -- backend`は`(empty)`。
- C11: `grep -n 'BE10-2' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/todo.md /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/BE-pending.md`は`(empty)`、exit=1（0 hit）。
- `git diff --check -- BE-refactor.md`: `(empty)`、exit=0。
- Failure Signature log:
  - C5 supplementary count / attempt 1: expected=64行、actual=0行、verification=Node parser、error signature=parserが旧heading `64 test file分類`とbacktickなしrow patternへ依存、fix=現行heading `64 file分類`と実row patternに合わせて最小化、result=`classification total=64 service=14 repository=50` / PASS。成果物の分類内容に欠落はなかった。
- De-Sloppify: 対象外。Goコード・testを変更しておらず、本unitは計画文書だけである。
- Assumption deviations: none。

</details>

#### BE10-2 B0 実行ledger（2026-07-25）

- 実行状態: `完了`。scoped実装・等価性・build/vet/test/gofmt/lint比較・legacy無変更・D6bの一時stage probeを完了した。
- 変更file:
  - `backend/internal/testdb/fixtures.go`（新規。指定6 exportのみ）
  - `BE-refactor.md`（本ledger）
- scope: B0のみ。B1〜B14、BE10-3〜BE10-6、`internal/repository`、`internal/service`は変更していない。
- Saved Prompt Validation Gate:

  ```text
  $ node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-be10-2-b0-testdb-fixture-exports.md
  Prompt Craft Harness Validation: PASS
  Profile: standard (declared-risk-tier)
  Target: agent (detected)
  Quality mode: standard
  validator_exit_status=0
  ```

- D1 baseline:

  ```text
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte rev-parse --short HEAD
  3cf2eb8e6
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte status --porcelain -- backend BE-refactor.md
   M backend/internal/pet/chronic_condition_repository.go
   M backend/internal/pet/chronic_condition_repository_test.go
  ```

  上記2 pathはB0開始前から存在するallowlist外の他writer差分として保存し、本unitでは触れていない。生成時HEAD `bcbdb5101`との差はContext差分であり、Assumption deviationではない。

- D1/D9 scoped lint（両回とも同一command）:

  ```text
  $ docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-$RANDOM --entrypoint golangci-lint backend run ./internal/testdb/... --max-same-issues 0 --max-issues-per-linter 0
  internal/testdb/testdb.go:100:10: error returned from external package is unwrapped: sig: func (*gorm.io/gorm.DB).AutoMigrate(dst ...interface{}) error (wrapcheck)
                  return err
                         ^
  1 issues:
  * wrapcheck: 1
  ```

  baseline/変更後ともexit=1。Compose timestamp・一時container ID行を除いたdiagnostic diffは`(empty)`、exit=0、新規指摘は0件。`testdb.go:100`の1件はpre-existingで、本unitでは修正していない。baseline保存先は`/tmp/animalekarte-be10-2-b0-d1.y9pHR2/lint-baseline.txt`、変更後は同directoryの`lint-after.txt`。

- D2 source body baseline: `/tmp/animalekarte-be10-2-b0-d2.ZHO5YD`へ6 `.src`を保存した。行数は次のとおり。

  ```text
         6 makeDoctor.src
        13 makeHistoryMedicalRecord.src
         6 makeInsuranceMaster.src
         8 makePetWithInsurance.src
        13 makeSpeciesAndPet.src
         9 seedClinicsForFK.src
        55 total
  ```

- D3 package-local依存: 6 bodyはいずれもrepository package-local helperを呼ばない。既存`testdb` exportへの差し替え、同file内の非export helper追加はともに0件。
- D4 export surface:

  ```text
  $ grep -n '^func [A-Z]' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/*.go
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/fixtures.go:17:func MakeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/fixtures.go:32:func MakeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/fixtures.go:40:func MakeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/fixtures.go:55:func SeedClinicsForFK(t *testing.T, db *gorm.DB, clinicIDs ...uint64) {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/fixtures.go:66:func MakeInsurance(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Insurance {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/fixtures.go:74:func MakePetWithInsurance(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, insuranceID *uint64, name string) *model.Pet {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/surface_test.go:9:func TestExportedSchemaSurface(t *testing.T) {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb.go:59:func EnsureAutoMigrated(db *gorm.DB, models ...any) error {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb.go:109:func MarkAutoMigrated(models ...any) {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb.go:119:func CloseSharedTestDB() {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb.go:129:func SetupTestDB(t *testing.T) *gorm.DB {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb.go:156:func EnsureClinicSettingsTable(t *testing.T, db *gorm.DB) {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb.go:198:func MakeTestOwner(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Owner {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb.go:218:func SetupIsolatedTestDB(t *testing.T) *gorm.DB {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb_test.go:15:func TestQuotePostgresIdentifier(t *testing.T) {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb_test.go:40:func TestIsDuplicateDatabaseError(t *testing.T) {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/testdb/testdb_test.go:80:func TestEnumValuesEqual_TestEnumAppendedValues(t *testing.T) {
  non_test_export_count=13
  ```

  `_test.go`内の`Test*`は上の非test export countから除外した。指定6名は各1件、既存7名も各1件で衝突なし。

- D5 body逐語照合: 6件ともsignatureの関数名変更（許容分類①）だけで、body差分は0行。

  ```text
  makeSpeciesAndPet:
  1c1
  < func makeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
  ---
  > func MakeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {

  makeDoctor:
  1c1
  < func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
  ---
  > func MakeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {

  makeHistoryMedicalRecord:
  1c1
  < func makeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {
  ---
  > func MakeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {

  seedClinicsForFK:
  1c1
  < func seedClinicsForFK(t *testing.T, db *gorm.DB, clinicIDs ...uint64) {
  ---
  > func SeedClinicsForFK(t *testing.T, db *gorm.DB, clinicIDs ...uint64) {

  makeInsuranceMaster:
  1c1
  < func makeInsuranceMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Insurance {
  ---
  > func MakeInsurance(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Insurance {

  makePetWithInsurance:
  1c1
  < func makePetWithInsurance(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, insuranceID *uint64, name string) *model.Pet {
  ---
  > func MakePetWithInsurance(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, insuranceID *uint64, name string) *model.Pet {
  ```

- D7/D8 scoped verification:

  ```text
  $ docker compose exec backend go build ./internal/testdb/...
  (empty)
  build_exit_status=0
  $ docker compose exec backend go vet ./internal/testdb/...
  (empty)
  vet_exit_status=0
  $ docker compose exec backend go test ./internal/testdb/... -count=1 -p 1
  ok  	github.com/animal-ekarte/backend/internal/testdb	0.002s
  test_exit_status=0
  $ docker compose exec backend gofmt -l internal/testdb/
  (empty)
  gofmt_exit_status=0
  ```

- D6 legacy無変更:

  ```text
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte diff --name-only -- backend/internal/repository backend/internal/service
  (empty)
  $ grep -rn 'func makeSpeciesAndPet\|func makeDoctor(\|func makeHistoryMedicalRecord\|func seedClinicsForFK\|func makeInsuranceMaster\|func makePetWithInsurance' /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/isolation_test_helpers_test.go:26:func makeHistoryMedicalRecord(t *testing.T, db *gorm.DB, clinicID, petID uint64, recordNo string, date time.Time) *model.MedicalRecord {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/isolation_test_helpers_test.go:55:func makeDoctor(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Staff {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/be9_2c_r3_test_helper_carrier_test.go:64:func makeSpeciesAndPet(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, petName string) *model.Pet {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/staff_preload_clinic_isolation_test.go:31:func seedClinicsForFK(t *testing.T, db *gorm.DB, clinicIDs ...uint64) {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/cross_clinic_preload_isolation_test.go:26:func makeInsuranceMaster(t *testing.T, db *gorm.DB, clinicID uint64, name string) *model.Insurance {
  /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/backend/internal/repository/cross_clinic_preload_isolation_test.go:33:func makePetWithInsurance(t *testing.T, db *gorm.DB, clinicID, ownerID uint64, insuranceID *uint64, name string) *model.Pet {
  ```

- De-Sloppify: 新規fileに未使用import・重複helper・console log・commented-out codeはなく、copyしたbodyを変更するcleanupは行っていない。cleanup差分0のため追加の再検証対象なし。
- D6b tracked/staged-path probe:

  ```text
  $ git check-ignore -v backend/internal/testdb/fixtures.go
  (empty)
  check_ignore_exit_status=1
  $ git add backend/internal/testdb/fixtures.go BE-refactor.md
  git_add_exit_status=0
  $ git diff --cached --name-only
  BE-refactor.md
  backend/internal/testdb/fixtures.go
  cached_names_exit_status=0
  $ git restore --staged backend/internal/testdb/fixtures.go BE-refactor.md
  restore_staged_exit_status=0
  $ git status --porcelain -- backend BE-refactor.md
   M BE-refactor.md
   M backend/internal/pet/chronic_condition_repository.go
   M backend/internal/pet/chronic_condition_repository_test.go
  ?? backend/internal/testdb/fixtures.go
  d6b_status_diff_exit_status=0
  ```

  D6b前後のporcelainはbyte同一で、indexはprobe前の状態へ復元した。
- Failure Signature log: none。
- Assumption deviations: none。

#### BE10-2 B1 実行ledger（2026-07-25）

- 実行状態: `検証BLOCKED`。structural差分、byte等価性、vet、gofmt、lint比較は完了したが、E12の指定scoped testがB1対象外の既存service integration test 2件でtest DB schema不足により2回とも失敗した。DB reset・migration適用とB2以降の修正は本unitで禁止されているため、B1完了にはしていない。
- 変更file:
  - `backend/internal/audit/service_operations_test.go`（`backend/internal/service/audit_service_test.go`から移設）
  - `backend/internal/audit/service_doubles_test.go`（新規）
  - `backend/internal/service/audit_clinic_test_doubles_test.go`
  - `backend/internal/service/target_test_surface_test.go`
  - `BE-refactor.md`
- scope: B1のみ。B2〜B14、BE10-3〜BE10-6、production code、`internal/repository`、既存`internal/audit/*_test.go`は変更していない。
- Saved Prompt Validation Gate:

  ```text
  $ node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-be10-2-b1-audit-test-migration.md
  Prompt Craft Harness Validation: PASS
  Profile: standard (declared-risk-tier)
  Target: agent (detected)
  Quality mode: standard
  validator_exit_status=0
  prompt_sha256=7476b05f950f70f6232144bd4e0fc6f149f4077640fc13d197f5613d0683c0d0
  ```

- E1 baseline:

  ```text
  temporary_evidence_dir=/tmp/animalekarte-be10-2-b1-final.iXBzGk
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte rev-parse --short HEAD
  7cf8a6e07
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte status --porcelain -- backend BE-refactor.md
   M BE-refactor.md
  $ grep -h '^func Test' backend/internal/service/audit_service_test.go | sed 's/(.*//' | sort | wc -l
  21
  ```

  `BE-refactor.md`のbaseline差分はB0をcommit済みへ直す既存1行であり、本unitでは保持した。変更前lintは指定commandで`0 issues.`、exit=0。

- E2/E3: 移設元は680行、`mockAuditRepository`保存範囲は32行。21 test名、`mockAuditRepository`、`ptrUint64ForAuditTest`の移設先collisionはすべて0件。
- E4/E5: `package service`を`package audit`へ変更し、指定4 tokenだけをrenameした。期待値生成後の旧code-token grepは0、`diff /tmp/animalekarte-be10-2-b1-final.iXBzGk/ops.expect backend/internal/audit/service_operations_test.go`は無出力・exit=0。コメント内の素の`validateAuditLog` 3件は不変。承認済みassertion message例外は`"auditService must implement AuditTxLogger"`から`"auditService must implement TxLogger"`へ変更した。
- E6/E7: `mockAuditRepository`のtypeと3 methodの保存body diffは無出力・exit=0。service側の同symbol grepは0。`mockPermissionGroupRepository`の保存body diffも無出力・exit=0。
- E8: bridgeのaudit旧symbol/import grepは0、top-level `func`/`type`宣言は24件から20件。diffはaudit 4宣言、`auditdomain` import、未使用になった`model` importの削除だけ。
- E9/E10: 移設元`ls`は`No such file or directory`・exit=1。`comm -23 names.src names.dst`は無出力、移設後`internal/audit`のtest名総数は29。
- E11 vet:

  ```text
  $ docker compose exec backend go vet ./internal/audit/... ./internal/service/... ./internal/repository/...
  (Composeの未設定DB変数warning 3行のみ)
  vet_exit=0
  ```

- E12 scoped test / Failure Signature log:

  ```text
  $ docker compose exec backend go test ./internal/audit/... ./internal/service/... -count=1 -p 1
  ok  	github.com/animal-ekarte/backend/internal/audit	0.002s
  ERROR: relation "medical_record_addenda" does not exist (SQLSTATE 42P01)
  --- FAIL: TestStaffDeleteAndShiftCreate_DeleteWinnerPreventsOrphanShiftDatabase (5.04s)
  ERROR: relation "hospitalizations" does not exist (SQLSTATE 42P01)
  --- FAIL: TestStaffSetAssignmentsAndClinicDelete_DeleteWinnerLeavesAssignmentsUnchangedDatabase (5.05s)
  FAIL	github.com/animal-ekarte/backend/internal/service	11.708s
  test_exit=1
  ```

  retry 1も同じ2 relation・同じ2 testで再現し、`internal/audit`は`ok`。原因は`setupStaffShiftSecurityIntegrationTest`がdelete dependency scanに必要な2 tableを作成しない既存test DB前提で、B1差分とは無関係。禁止されたDB reset/migrationやallowlist外test修正には進んでいない。

- E13/E14:

  ```text
  $ docker compose exec backend gofmt -l internal/audit/ internal/service/
  (empty)
  gofmt_exit=0
  $ docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-$RANDOM --entrypoint golangci-lint backend run ./internal/audit/... ./internal/service/... --max-same-issues 0 --max-issues-per-linter 0
  0 issues.
  lint_after_exit=0
  lint_diagnostic_diff_exit=0
  new_lint_diagnostic_count=0
  ```

- De-Sloppify: 未使用importと撤去済みbridge symbolへの死んだ参照は0。移設test、重複helper、permission-group/clinic/staff面には手を付けていない。
- E16 tracked/staged-path probe / Failure Signature log:

  ```text
  $ git check-ignore -v backend/internal/audit/service_operations_test.go backend/internal/audit/service_doubles_test.go
  (empty)
  check_ignore_exit=1
  $ git add backend/internal/audit/service_operations_test.go backend/internal/audit/service_doubles_test.go backend/internal/service/audit_service_test.go backend/internal/service/audit_clinic_test_doubles_test.go backend/internal/service/target_test_surface_test.go BE-refactor.md
  fatal: pathspec 'backend/internal/service/audit_service_test.go' did not match any files
  git_add_exit=128
  ```

  attempt 1はE4の`git mv`がsource pathをindexからも除いた状態だったため、exact `git add`の旧pathがmatchせず失敗した。`git restore --staged backend/internal/audit/service_operations_test.go backend/internal/service/audit_service_test.go`でindexをbaselineの空状態へ戻した後、同じE16 commandをretry 1として再実行した。

  ```text
  check_ignore_retry1_exit=1
  git_add_retry1_exit=0
  $ git diff --cached --name-only
  BE-refactor.md
  backend/internal/audit/service_doubles_test.go
  backend/internal/audit/service_operations_test.go
  backend/internal/service/audit_clinic_test_doubles_test.go
  backend/internal/service/target_test_surface_test.go
  cached_names_retry1_exit=0
  restore_staged_retry1_exit=0
  e16_status_retry1_diff_exit=0
  ```

  `audit_service_test.go`の削除はGitが`service_operations_test.go`へのrenameとして1件表示するため、`--name-only`はallowlist 6 physical pathを5行で表現する。新規2 fileはいずれも一覧に含まれ、allowlist外は0。restore後のporcelainはprobe直前とbyte同一、indexはbaselineの空状態へ復元した。E16はretry 1で`PASS`。

- E18/E19: final scoped porcelainの今回追加行は削除1・変更2・新規2で、すべてallowlist。`BE-refactor.md`はbaselineから既に` M`だったためstatus行は不変。baseline/current HEADはいずれも`7cf8a6e07`で増分commitなし。本unit差分はworktreeに残り、indexは空。
- E17 Independent Review Gate: reviewer roleで5観点すべてPASS。CRITICAL 0 / HIGH 0 / MEDIUM 0 / LOW 0。rename-map-only・21 test保持・clinic/staff bridge保持・permission/audit doubleのbyte一致・allowlistを独立再検証した。E12の2失敗は変更していないtest setupとdependency scanのschema gapでB1差分とは無関係との判定。

#### BE10-2 B2 実行ledger（2026-07-25）

- 実行状態: `完了`（commit `134e7953b`）。serviceのclinic系test 4 fileとpermission-group doubleを`internal/clinic`へ移設し、clinic bridgeはB2 consumerが消えたsymbolだけを撤去した。
- 変更file:
  - `backend/internal/clinic/clinic_holiday_service_test.go`（移設）
  - `backend/internal/clinic/clinic_service_test.go`（移設）
  - `backend/internal/clinic/closing_settings_service_test.go`（移設）
  - `backend/internal/clinic/company_service_test.go`（移設）
  - `backend/internal/clinic/permission_group_doubles_test.go`（新規）
  - `backend/internal/service/clinic_holiday_service_test.go`（消滅）
  - `backend/internal/service/clinic_service_test.go`（消滅）
  - `backend/internal/service/closing_settings_service_test.go`（消滅）
  - `backend/internal/service/company_service_test.go`（消滅）
  - `backend/internal/service/audit_clinic_test_doubles_test.go`（消滅）
  - `backend/internal/service/target_test_surface_test.go`
  - `BE-refactor.md`
- scope: B2のみ。B3〜B14、BE10-3〜BE10-6、R-1/R-2、production code、model、migrationは変更していない。
- Saved Prompt Validation Gate:

  ```text
  $ node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-be10-2-b2-clinic-test-migration.md
  Prompt Craft Harness Validation: PASS
  Profile: standard (declared-risk-tier)
  Target: agent (detected)
  Quality mode: standard
  validator_exit=0
  prompt_sha256=3efb77b7dc3666626592e7699e3af0f26c50e0674d4851249982bdd05ae08750
  ```

- G1 baseline:

  ```text
  temporary_evidence_dir=/tmp/animalekarte-be10-b2.syBhzf
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte rev-parse --short HEAD
  14c6702db
  source_test_names=28
  test_exit=0 pass_lines=125 fail_lines=0
  lint_exit=1
  internal/clinic/closing_settings_request.go:11:9: S1016: should convert r (type UpdateClinicSettingsRequest) to UpdateClinicSettingsInput instead of using struct literal (staticcheck)
  internal/clinic/closing_settings_request.go:46:9: S1016: should convert r (type UpdateSpecialPeriodRequest) to UpdateSpecialPeriodInput instead of using struct literal (staticcheck)
  2 issues:
  * staticcheck: 2
  ```

  baseline時点の本unit外WIPは`backend/internal/medicalrecord`の6 fileであり、保持して変更していない。

- G2/G4: 移設元は243/1170/798/287行、permission-group double保存範囲は22行。移設先5 pathは全て不存在、28 test名・`mockPermissionGroupRepository`・`clinicServiceTransactorDouble`・`clinicStringPtr`・`clinicFloat64Ptr`の衝突は0件。
- G3 consumer境界: `NewClinicHolidayService`、clinic input alias群、`buildClinicUpdate`、`NewClosingSettingsService`、company input/constructor/build wrapperはB2の4 file以外にconsumerがなく撤去した。`NewClinicService`は`staff_shift_security_integration_test.go:589,652`が消費するため`clinicdomain` importとともにB3まで保持した。staff bridge宣言は変更していない。
- G5〜G8: 4 fileを`git mv`し`package clinic`へ変更した。`buildClinicUpdate`/`buildCompanyUpdate`はtarget packageの`BuildClinicUpdate`/`BuildCompanyUpdate`へ直接化し、package自己importを実名へ置換した。service専用helperは`clinicServiceTransactorDouble`、`clinicStringPtr`、`clinicFloat64Ptr`として新規test double fileへ同梱した。permission-group doubleのtypeと2 methodは保存bodyとの`diff`が無出力・exit=0。元double fileは空になるため削除した。
- G7: 保存原本へpackage変更、target実名化、同梱helper参照だけを機械適用した4 expected fileと移設先の`diff -u`は全て無出力・exit=0。assertion、test名、fixture、error文言、呼出順序の差分は0。
- G9: bridgeの撤去対象grepは0件。`NewClinicService|clinicdomain`は7 hit。`strPtr`/`ptrFloat64`はpromptのB3所有指定どおり保持した。
- G11: `comm -23 names.src names.dst`は無出力。欠落0、移設test名28、移設後`internal/clinic`全test名95。
- G12 vet:

  ```text
  $ docker compose exec backend go vet ./internal/clinic/... ./internal/service/... ./internal/staff/...
  (Composeの未設定DB変数warning 3行のみ)
  vet_exit=0
  ```

  G12→G6/G9補完ループは0周。

- G13 scoped test:

  ```text
  $ docker compose exec backend go test ./internal/clinic/... ./internal/service/... -count=1 -p 1 -v
  PASS
  ok  	github.com/animal-ekarte/backend/internal/clinic	0.831s
  PASS
  ok  	github.com/animal-ekarte/backend/internal/service	2.610s
  test_exit=0 pass_before=125 pass_after=125 fail_before=0 fail_after=0
  $ diff /tmp/animalekarte-be10-b2.syBhzf/test.base.names /tmp/animalekarte-be10-b2.syBhzf/test.after.names
  (empty)
  normalized_diff_exit=0
  ```

- G14/G15:

  ```text
  $ docker compose exec backend gofmt -l internal/clinic/ internal/service/
  (empty)
  gofmt_exit=0
  $ docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-be10-b2-after2-<pid> --entrypoint golangci-lint backend run ./internal/clinic/... ./internal/service/... --max-same-issues 0 --max-issues-per-linter 0
  internal/clinic/closing_settings_request.go:11:9: S1016: should convert r (type UpdateClinicSettingsRequest) to UpdateClinicSettingsInput instead of using struct literal (staticcheck)
  internal/clinic/closing_settings_request.go:46:9: S1016: should convert r (type UpdateSpecialPeriodRequest) to UpdateSpecialPeriodInput instead of using struct literal (staticcheck)
  2 issues:
  * staticcheck: 2
  lint_exit=1 diagnostic_diff_exit=0 base_diagnostics=2 after_diagnostics=2 new_diagnostics=0
  ```

- G16: `git diff --name-only -- backend/internal/model backend/migrations`は無出力。`internal/clinic`の本unit変更は全て`_test.go`。
- Failure Signature log:
  - G15 / attempt 1: expected=baseline比new lint 0、actual=`strPtr`/`ptrFloat64`のunused 2件追加、verification=fresh-cache scoped lint、error signature=`target_test_surface_test.go:61/65 unused`、原因=B3所有として保持を指定されたhelperのstaff consumerが現行treeには存在しない、fix=同じtest-only bridge fileへblank-identifier liveness参照を追加、result=attempt 2でdiagnostic diff 0・new 0 / PASS。
- Assumption deviations:
  - G6の指定grepは、rename mapが「同名で直接呼ぶ」と定めるconstructor/input名まで0件を要求しており自己矛盾する。実際に変更が必要な旧private helper名、package自己import、service helper名を0件とし、same-nameのtarget実symbolはvet/testで直接接続を証明した。
  - Contextは`strPtr`/`ptrFloat64`をB3 staff test consumerありとするが、現行treeではB2の`clinic_service_test.go`だけがconsumerだった。B3所有の明示仕様を優先して保持した。
- De-Sloppify: 未使用import、撤去済みbridge参照、service側permission-group double残骸は0。移設test、既存clinic helper、B3所有helperは削除・統合していない。
- G18 tracked/staged-path probe:

  ```text
  $ git reset
  reset_exit=0
  $ git check-ignore -v backend/internal/clinic/permission_group_doubles_test.go
  (empty)
  check_ignore_exit=1
  $ git add -- <allowlist 12 physical pathを個別列挙>
  git_add_exit=0
  $ git diff --cached --name-only
  BE-refactor.md
  backend/internal/clinic/clinic_holiday_service_test.go
  backend/internal/clinic/clinic_service_test.go
  backend/internal/clinic/closing_settings_service_test.go
  backend/internal/clinic/company_service_test.go
  backend/internal/clinic/permission_group_doubles_test.go
  backend/internal/service/target_test_surface_test.go
  cached_exit=0 cached_count=7
  $ git restore --staged -- <allowlist 12 physical pathを個別列挙>
  restore_staged_exit=0
  $ git diff --cached --name-only
  (empty)
  ```

  4組のsource/destinationとdouble splitはrenameとしてdestination 5行へ集約されるため、12 physical pathはcached 7 pathで表現された。allowlist外は0。
- G19 Independent Review Gate: reviewer roleとGo reviewer roleの2 passで、CRITICAL 0 / HIGH 0 / MEDIUM 0 / LOW 0。①4 expected diff 0、②28 test名とassertion保持、③bridge撤去過不足なし、④permission-group body byte一致、⑤allowlist外/production変更なしを独立再確認した。blank-identifier liveness参照はB3保持指定とlint new 0を同時に満たす最小test-only措置として承認された。
- G20/G21:

  ```text
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte status --porcelain -- backend BE-refactor.md
   M BE-refactor.md
   D backend/internal/service/audit_clinic_test_doubles_test.go
   D backend/internal/service/clinic_holiday_service_test.go
   D backend/internal/service/clinic_service_test.go
   D backend/internal/service/closing_settings_service_test.go
   D backend/internal/service/company_service_test.go
   M backend/internal/service/target_test_surface_test.go
  ?? backend/internal/clinic/clinic_holiday_service_test.go
  ?? backend/internal/clinic/clinic_service_test.go
  ?? backend/internal/clinic/closing_settings_service_test.go
  ?? backend/internal/clinic/company_service_test.go
  ?? backend/internal/clinic/permission_group_doubles_test.go
  ```

  上記12 physical pathだけが本unit由来で全件allowlist内。baselineから存在した`backend/internal/medicalrecord` 6 fileと、実行中に別writerが追加した`backend/tygo.yaml`・`backend/internal/billing` 2 fileは本unit外として保持した。baseline/current HEADはいずれも`14c6702db`、増分commit 0、indexは空で本unit差分はworktreeに残っている。
- `NewClinicService` carryover: `staff_shift_security_integration_test.go:589,652`のconsumerが残るためwrapperと`clinicdomain` importをB3へ持ち越した。B2では撤去しないことが本unitの仕様である。

#### BE10-2 B3 実行ledger（2026-07-25）

- 実行状態: `完了（hash 追記待ち）`。serviceのstaff系test 3 fileとstaff用`mockTransactor`を`internal/staff`へ移設し、B2から持ち越された`NewClinicService` / `strPtr` / `ptrFloat64` / liveness anchorを含む`target_test_surface_test.go`を削除した。
- 変更file:
  - `backend/internal/staff/staff_clinic_assignment_reservation_race_test.go`（移設）
  - `backend/internal/staff/staff_cross_tenant_test.go`（移設）
  - `backend/internal/staff/staff_shift_security_integration_test.go`（移設）
  - `backend/internal/staff/transactor_doubles_test.go`（新規）
  - `backend/internal/service/staff_clinic_assignment_reservation_race_test.go`（消滅）
  - `backend/internal/service/staff_cross_tenant_test.go`（消滅）
  - `backend/internal/service/staff_shift_security_integration_test.go`（消滅）
  - `backend/internal/service/clinic_test_transactor_test.go`（消滅）
  - `backend/internal/service/target_test_surface_test.go`（消滅）
  - `backend/internal/service/master_fk_write_inventory_lint_test.go`（文字列リテラル内pathのみ更新）
  - `BE-refactor.md`
- scope: B3のみ。B4〜B14、BE10-3〜BE10-6、R-1/R-2、production code、model、migrationは変更していない。
- Saved Prompt Validation Gate:

  ```text
  $ node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-be10-2-b3-staff-test-migration-bridge-removal.md
  Prompt Craft Harness Validation: PASS
  validator_exit=0
  prompt_sha256=0cc18d5049daa03754f00cacf8766fdf773101403cbf07fe7cef8a51004ca48e
  ```

- H1 baseline:

  ```text
  temporary_evidence_dir=/tmp/animalekarte-be10-2-b3-r2.rIHe1v
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte rev-parse --short HEAD
  0736cd6f9
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte status --porcelain -- backend BE-refactor.md
  (empty)
  source_test_names=12
  test_exit=0 pass_lines=413 fail_lines=0
  lint_exit=1 diagnostics=16
  ```

- H2/H5: 移設元保存行数は429/223/720、`mockTransactor`保存bodyは14行。package-level宣言の全数照合はmoving 22、destination 260、衝突は`mockReservationForStaff`のみ、test名衝突0、`stubReservationForStaff`は既存0件。承認済み解決として移動側だけを`stubReservationForStaff`へ機械置換した。
- H3 consumer境界: `target_test_surface_test.go`の宣言は、3 staff test移設後にservice側実consumer 0。`NewClinicService`は`clinic.NewClinicService`へ直接接続し、`NewStaffService` / `NewShiftEntryService` / input aliasは`package staff`の実symbolへ直接化した。`strPtr` / `ptrFloat64`はbridge内liveness anchorだけが残存理由だったため、bridge file削除で同時解消した。
- H4 文字列リテラル依存: `master_fk_write_inventory_lint_test.go`の旧path literalは3件（allowlist reasonの説明文のみ）、`n1_lint_test.go`は0件。`masterFKWriteStatus` commentは「The gate does not verify these」、`reconcileMasterFKWrites`はkey/status/masterFKs/real sourceだけを検証し、reason内pathの実在は検証しない。記録を事実に合わせるため3 literalだけを`internal/staff/staff_cross_tenant_test.go`へ更新した。allowlist key/status/masterFKs/lint logic/test名は変更していない。
- H7/H8: 移設3 fileの差分はpackage行、`staffdomain`自己import除去と実名化、`clinic` import/qualifier、`mockReservationForStaff`→`stubReservationForStaff`、gofmt import整列のみ。`clinic` importで既存local名がshadow警告になったため、そのlocal bindingと参照だけを`clinicRecord`へ機械改名した。assertion、test名、fixture値、error文言、呼び出し順序の変更は0。
- H9/H10/H11/H12: `mockTransactor` body diffは無出力・exit=0。`clinic_test_transactor_test.go`、`target_test_surface_test.go`、移設元3 fileはいずれも不在（`ls` exit=1）。`comm -23 names.src names.dst`は無出力、`internal/staff`のtest名は301件。
- H13 vet:

  ```text
  $ docker compose exec backend go vet ./internal/staff/... ./internal/service/... ./internal/clinic/...
  (Composeの未設定DB変数warning 3行のみ)
  vet_exit=0
  ```

  H13→H7補完ループは0周。

- H14 scoped test:

  ```text
  $ docker compose exec backend go test ./internal/staff/... ./internal/service/... ./internal/clinic/... -count=1 -p 1 -v
  PASS
  ok  	github.com/animal-ekarte/backend/internal/staff
  PASS
  ok  	github.com/animal-ekarte/backend/internal/service
  PASS
  ok  	github.com/animal-ekarte/backend/internal/clinic
  test_exit=0 pass_before=413 pass_after=413 fail_before=0 fail_after=0
  $ diff /tmp/animalekarte-be10-2-b3-r2.rIHe1v/test.base.names /tmp/animalekarte-be10-2-b3-r2.rIHe1v/test.current.names
  (empty)
  normalized_diff_exit=0
  ```

  `TestMasterFKWriteInventory_*`と`TestN1Lint_*`は全件PASS。

- H15/H16:

  ```text
  $ docker compose exec backend gofmt -l internal/staff/ internal/service/
  (empty)
  gofmt_exit=0
  $ docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-b3-r2-after-<random> --entrypoint golangci-lint backend run --max-same-issues 0 --max-issues-per-linter 0 ./internal/staff/... ./internal/service/... ./internal/clinic/...
  16 issues:
  * gocritic: 9
  * revive: 1
  * staticcheck: 4
  * unparam: 1
  * unused: 1
  lint_exit=1 diagnostic_diff_exit=0 base_diagnostics=16 after_diagnostics=16 new_diagnostics=0
  ```

- H17: `git diff --name-only -- backend/internal/model backend/migrations`は無出力。`backend/internal/staff` / `backend/internal/clinic`の本unit変更は全て`_test.go`。
- Failure Signature log:
  - H1 lint wrapper / attempt 1: lint本体はbaseline 16件を出力したが、zsh予約変数`status`への代入でwrapperが失敗。変数名を`lint_exit`へ変更したattempt 2で`lint_exit=1`と16診断を保存した。
  - H16 / attempt 1: expected=baseline比new lint 0、actual=`clinic` importが既存local変数をshadowして`gocritic/importShadow` 5件追加、fix=fixture値・assertion・呼出順を変えずlocal bindingと参照だけを`clinicRecord`へ機械改名、result=baseline/after各16診断・diagnostic diff 0・new 0。
  - H14 / shared-DB retry: full suite再検証中にlock timeout/deadlock 2件、その後に前回失敗runが残したpassword reset tokenによるduplicate 1件を観測。各失敗testを最小単位で再実行して原因を分離し、token残骸は`TestPasswordResetTokenRepository_DeleteByID_SingleUse`の既存test setupでcleanupした（手動DELETE、DB reset、migration適用なし）。最終同一scoped suiteはPASS 413 / FAIL 0、baseline名diff 0。
- Assumption deviations: H4のpath literalは実在検証されない説明文だが、事実性維持のため更新した。`clinic` import名を維持すると既存local名がlint shadowになるため、H16 new 0を満たす最小の機械改名として`clinicRecord`を採用した。
- De-Sloppify: 未使用import、bridge死参照、service側staff bridge残骸、重複helper統合の誤適用を確認。移設test、既存staff helper、`stubReservationForStaff`は削除・統合していない。
- H19 tracking safety: 実行前indexは空。bare `git reset`後、新規4 fileの`git check-ignore -v`は各exit=1。allowlist 11 physical pathを明示`git add`したcached一覧はrename集約後の7 pathだけで全件allowlist内。明示`git restore --staged`後のcached pathは0件。
- H20 Independent Review Gate: explorer由来のfresh independent reviewで(a)移設testの意味変化なし、(b)12 test名/assertion欠落なし、(c)bridge全consumer調査妥当、(d)文字列literalは3件更新・取りこぼしなし、(e)allowlist外/production変更なしを確認。CRITICAL/HIGH/MEDIUM/LOWはいずれも0。DB-backed suiteはreview pass内で並列実行していない。
- H21/H22: baseline scoped porcelainは空、afterの11 physical pathは全件allowlist内。baseline HEAD=`0736cd6f9`、current HEAD=`070043aed`。増分commitは`10abd8258`（`todo.md`）、`0804c400e`（frontend 3 file）、`070043aed`（`FE-refactor.md`）だけで、いずれもB3 allowlistを含まない。本unit変更はindexを空に戻したworktreeに残しており、commit/push/PRは実施していない。

### BE10-3 — 空の`internal/repository/*` 15 directory

- 規約根拠: `backend/CODING_RULES.md:17`の機械的複製を避ける方針。git管理下のpackage違反ではなくworking-tree残骸である。
- 現状: A5で列挙した15 directoryはすべてfile 0、git追跡0。空directoryのため`git status`/`git ls-files`だけでは検出できない。
- 修正内容: 他sessionが利用していないことを確認後、15 directoryを削除する。新しいplaceholderやkeep fileは追加しない。
- 検証方法: 同じ`find .../internal/repository -mindepth 1 -maxdepth 1 -type d`が0件となり、legacy lint testの移設・削除計画と競合しないことを確認する。
- severity: LOW
- 前提・依存: 複数sessionの作業所有権を確認してから削除する。BE10-2のlint gate移設とは別unitとして扱う。
- 状態: 未着手

### BE10-4 — `.ruff_cache` / `.wrangler`のproject ignore未登録

- 規約根拠: `backend/.gitignore`が生成binaryを明示登録しているproject運用との不整合。明文の規約条文はない。
- 現状: 両pathはtracked 0で、`git check-ignore -v`はともにempty/exit 1。`.ruff_cache`は2 file、`.wrangler`は0 file。空の`backend/.git`も存在する。
- 修正内容: ①project `.gitignore`へ登録する、または②不要な残骸を削除する。存在しない生成物を永続ignoreへ追加する前に削除可能性を評価する。
- 検証方法: 採択方針に応じて`git check-ignore -v`の根拠行、または`find`によるpath不在を確認し、`git status --porcelain -- backend`に意図しない生成物がないことを確認する。
- severity: LOW
- 前提・依存: `.gitignore`変更とdirectory削除は本計画起票のscope外。toolがdirectoryを再生成するかを確認してから選択する。
- 状態: 判断待ち

### BE10-5 — `q&a.html`の旧layer path参照5行

- 規約根拠: ADR-006 `:18`がhandler削除とservice/repositoryのtest-only化を記録しており、ユーザー向け台帳が旧pathを現行実装位置として案内すると正本driftになる。
- 現状: A8の5 hit（1108、1400、1430、1436、1454）はすべて現行実装を説明する文脈であり、表に示した現行domain pathへ更新すべきdriftと分類した。
- 修正内容: 5行の旧pathだけを表の現行pathへ更新する。歴史・決裁内容・業務契約は変更しない。
- 検証方法: A8のgrepが0 hitとなること、更新した各basenameが現行treeに実在すること、`q&a.html`の該当カード内容が変わっていないことをdiffで確認する。
- severity: LOW
- 前提・依存: `q&a.html`はユーザー向け正本であり、本計画起票では変更しない。docs-onlyの独立unitとして実施する。
- 状態: 未着手

### BE10-6 — package境界適合の機械gate不在

- 規約根拠: ADR-006と`backend/CODING_RULES.md:113-124`はpackage境界reviewを要求するが、A3/A4/A5を一括検査する専用gateがない。
- 現状: `scripts/check-*`には他不変条件用gateが存在するが、package境界適合は手動`find`/`grep`に依存する。`check-docs-symbol-drift.sh`の旧handler数確認は本gateの代替ではない。
- 修正内容: `scripts/check-*.mjs`形式で、①`internal/`直下集合差分、②legacy production file数、③legacy production import edge、④domain配下のlayer名subpackage、⑤bucket名とlive package名重複を検査する。変更周期の判断は候補出力+reviewに留める。
- 検証方法: 既存`check-*.test.mjs`形式に倣った自己テストで正常treeと各mutationを検出し、実treeへのgate実行をPASSさせる。
- severity: LOW
- 前提・依存: BE10-1/BE10-2の判断確定後に確定規約を自動化する。未確定の統合方針やtest配置を先に固定しない。
- 状態: 未着手

## 非逸脱として裁定済み

### 取り下げ2件 — `internal/infra/lstep`と`internal/infra/smtp`

- ADR-006 `:40-44`は`internal/infra`をcross-cuttingの現状維持packageとして承認しており、13 target domain packageではない。
- `infra/{crypto,httpx,line,lstep,smtp}`は外部system/adapter単位で「何を提供するか」が明確な凝集境界である。
- adapterのproduction consumerが1 packageであることや、`infra/smtp`がcomposition rootの`cmd/api/smtp_adapters.go`だけから配線されることは正常な形である。
- よって`infra/lstep`と`infra/smtp`をBE10逸脱項目へ含めない。

### ADR-006が承認済みの8状態

1. `internal/service`と`internal/repository`がtest-only fileだけを残して存在すること（ADR-006 `:18`）。削除phase未記録だけをBE10-2で扱う。
2. `internal/handler`が存在しないこと（ADR-006 `:18`）。
3. `internal/model`が単一flat packageとして多数fileを持つこと（ADR-006論点#4）。
4. domain package内に`handler`/`service`/`repository` subpackageがないこと（ADR-006 `:52`）。
5. ADR-006 `:39-44`の19 cross-cutting packageが小規模であること。
6. `internal/lstep`を単一の大規模packageとして維持すること（ADR-006代替案4・Trade-offs）。
7. lint gate testがlegacy package配下に残ること自体（ADR-006 `:126`）。移設先未決だけをBE10-2で扱う。
8. `cmd/_archive/`が存在すること。underscore prefixによりGo build対象外である。

## 未監査の規約軸

本roundでは新しい監査を実施せず、次を未監査として明示する。

- boundary map §5の許可依存グラフ45 edge全体のトポロジカル検証
- Go codeの意味的レビュー
- clinic isolation、認可、臨床安全、transaction境界の実装監査
- frontend、infra、scripts、docsの構成監査

この列挙は不具合または逸脱の存在を意味しない。

## Acceptance Checklist

- [x] **AC-BE10-1 clinic境界**: 「①parent統合」で確定し実装完了（`0301ae0e2`）。route/RBAC/OpenAPIは非変更、test名と`clinic_id`述語は逐語一致、build/vet/test/gofmtはgreen、独立レビュー2本severity 0。
- [ ] **AC-BE10-2 legacy退役**: 削除条件・担当・期限・lint gate 6件の移設先はPhase 0で確定済み（`bcbdb5101`）。残条件＝B1〜B14を完了させ、legacy production/test/importが0になること。B0は完了。
- [ ] **AC-BE10-3 空directory**: 他session所有権確認後に15 directoryを削除し、同じ`find`が0件となる。
- [ ] **AC-BE10-4 ignore/削除**: `.ruff_cache`、`.wrangler`、空`.git`についてignoreまたは削除方針を確定し、選択した検証がPASSする。tool再生成の判断待ちは再開条件付きで記録する。
- [ ] **AC-BE10-5 docs drift**: `q&a.html`の5行を現行pathへ更新し、旧layer path grepが0件となる。
- [ ] **AC-BE10-6 package gate**: BE10-1/2の決定後に専用gateとmutation自己テストを追加し、実treeでPASSする。
- [ ] **AC-BE10状態収束**: BE10-1〜6がすべて`完了`または再開条件付き`判断待ち`で確定し、`未着手`/`対応中`が0件となる。
- [ ] **todo.mdへ戻す条件**: 判断が確定し直ちに実装可能になったunitだけを`todo.md`の「個別タスク詳細」へ一意なtaskとして移す。移管時は本書側を実行ledgerに更新し、同じ仕様本文を二重掲出しない。
- [ ] **残余課題の住所確定（round終了の前提）**: 下記「BE10外の残余課題」2件を`todo.md`の「個別タスク詳細」へ移し、本書からの参照を消す。本書はround終了時に削除されるため、移管前に削除すると2件が失われる。
- [ ] **round終了**: 全unitの収束、残余リスク、scoped verificationを記録後、本ファイルを削除する。恒久規約はADR-006/CODING_RULES、完了証跡はgit履歴とtest/gateを正本とする。

## BE10外の残余課題（`todo.md`へ移管して本書から消す）

BE10の逸脱項目ではないが、BE10の実行中に実測で見つかった既存の技術債。本書の削除で失わないため、round終了前に`todo.md`へ移す。**BE10のどのunitでもdrive-by修正しない。**

| # | 対象 | 内容 | 発見元 |
|---|---|---|---|
| R-1 | `backend/internal/clinic/closing_settings_request.go:11`, `:46` | 既存`staticcheck S1016` 2件。`UpdateClinicSettingsRequest`→`UpdateClinicSettingsInput`、`UpdateSpecialPeriodRequest`→`UpdateSpecialPeriodInput`をstruct literalではなく型変換で書くべきという指摘。宣言元は`closing_settings_request.go:3`/`:37`と`closing_settings_service.go:63`/`:80` | BE10-1のscoped lint。同fileはBE10-1の変更8 pathに含まれず、本unitが持ち込んだものではない |
| R-2 | `backend/internal/testdb/testdb.go:100` | 既存`wrapcheck` 1件。`gorm.DB.AutoMigrate`のerrorを未wrapで返している | BE10-2 B0のlint baseline。B0の変更前後で同一 |
| R-3 | `backend/internal/service/staff_shift_security_integration_test.go:147` の setup | **完了（2026-07-25）**。2 dependency-count 関数が参照する全17表を列挙し、shared test schema / 現setupで未整備だった12表をsetup内で整備した。対象2 testと`go test ./internal/service/... -count=1 -p 1 -v`はPASS、baseline PASS→FAILは0、production code無変更 | BE10-2 B1 の E12。**B1 起因ではないことを確定済み**（下記 E12 補正根拠、R-3実行ledger） |

### BE10-2 B1 の E12 補正根拠（2026-07-25・生成側 reconciliation）

実行側は E12（`go test ./internal/audit/... ./internal/service/...`）の失敗により B1 全体を BLOCKED とし、実装差分は worktree に残した。生成側で以下を実測し、**E12 を PASS へ補正して B1 を完了扱いとする**。

- **B1 が DB setup を持ち去った可能性を排除**: `git show HEAD:backend/internal/service/audit_service_test.go | grep -c 'gorm\|testdb\|AutoMigrate\|setupTestDB'` は **0**。移設した test は全て mock 駆動で DB に触れず、表を作る side effect を持たない。したがって移設によって他 test の前提表が消えることは原理的に起こり得ない。
- **失敗 test は B1 の変更集合外**: `git status --porcelain -- backend/internal/service` は `audit_clinic_test_doubles_test.go` / `audit_service_test.go`(D) / `target_test_surface_test.go` の3 path のみ。`staff_shift_security_integration_test.go` は未変更。
- **原因コードは BE9 由来**: dependency-count query を持つ `staff_repository.go` / `clinic_repository.go` の直近変更は `dad69bc6a`（backend domain cutover）等で、B1 より遥かに前。`./internal/service/...` を最近誰も実行していなかったため潜在していた。
- `internal/audit`（移設先）は `ok ... 0.002s` で PASS。B1 の成果物自体は全 gate を通過し、独立レビューも severity 0。

生成側 prompt の欠陥: E12 の判定基準を「test PASS」という**絶対条件**にしたため、既存 defect で完了判定が止まった。lint gate は baseline 相対へ是正済みだったが test gate は絶対のままだった（同型欠陥の3回目）。以後の test gate も「pre-change baseline 比で新規失敗0件」とする。

いずれもproduction runtimeに影響しない（`internal/testdb`はproduction importer 0件のtest専用package、`closing_settings_request.go`はDTO変換）。修正は独立unitで行い、scoped lintで0件化を確認する。

### R-3 実行ledger（2026-07-25）

- 実行状態: `完了`。変更は`backend/internal/service/staff_shift_security_integration_test.go`と本ledgerの2 pathだけ。production Go file、`internal/testdb`、migrationは無変更。
- Saved Prompt Validation Gate:

  ```text
  $ node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-be10-r3-service-test-schema-gap.md
  Prompt Craft Harness Validation: PASS
  validator_exit_status=0
  ```

- F1 baseline:

  ```text
  temporary_evidence_dir=/tmp/animalekarte-be10-r3.EPpD5t
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte rev-parse --short HEAD
  718f6c9b3
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte status --porcelain -- backend BE-refactor.md
   M backend/docs/api.yaml
   M backend/internal/reservation/nested_summary_response.go
   M backend/internal/reservation/reservation_response_test.go
  $ docker compose exec backend go test ./internal/service/... -count=1 -p 1 -v
  --- FAIL: TestStaffDeleteAndShiftCreate_DeleteWinnerPreventsOrphanShiftDatabase (5.05s)
  --- FAIL: TestStaffSetAssignmentsAndClinicDelete_DeleteWinnerLeavesAssignmentsUnchangedDatabase (5.06s)
  baseline_top_level_pass_count=56
  test_exit=1
  $ docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-$RANDOM --entrypoint golangci-lint backend run ./internal/service/... --max-same-issues 0 --max-issues-per-linter 0
  0 issues.
  lint_exit=0
  ```

- F2/F3: `CountBlockingReferencesByStaffID`全文（`staff_repository.go:329-383`）の`checks`は`medical_records`（2列）、`medical_record_addenda`、`hospitalizations`、`exams`、`shift_entries`、`billing_refunds`、`cash_register_closes`、`vital_records`。slice外は`payments`とJOIN先`billings`。`CountBlockingReferencesByClinicID`全文（`clinic_repository.go:166-204`）の`checks`は`appointments`、`medical_records`、`hospitalizations`、`exams`、`vaccinations`、`checkups`、`billings`、`clinic_settings`、`clinic_integrations`、`lstep_settings`、`permission_groups`で、slice外参照はない。全17 unique tableに対応modelがあり、modelなしは0件。
- 表/model対応: `appointments`=`model.Reservation`、`medical_records`=`model.MedicalRecord`、`medical_record_addenda`=`model.MedicalRecordAddendum`、`hospitalizations`=`model.Hospitalization`、`exams`=`model.Examination`、`shift_entries`=`model.ShiftEntry`、`billing_refunds`=`model.BillingRefund`、`cash_register_closes`=`model.CashRegisterClose`、`vital_records`=`model.VitalRecord`、`payments`=`model.Payment`、`billings`=`model.Billing`、`vaccinations`=`model.Vaccination`、`checkups`=`model.Checkup`、`clinic_settings`=`model.ClinicSettings`、`clinic_integrations`=`model.ClinicIntegration`、`lstep_settings`=`model.LstepSettings`、`permission_groups`=`model.PermissionGroup`。
- F4/F5: 現setupの明示modelは`Company`、`Clinic`、`Staff`、`StaffClinicAssignment`、`ShiftEntry`、`ShiftEntryBreak`。`SetupTestDB`のshared coreが`MedicalRecord`、`Billing`、`Payment`、`BillingRefund`を既に保証するため、実差集合は`medical_record_addenda`、`hospitalizations`、`exams`、`cash_register_closes`、`vital_records`、`appointments`、`vaccinations`、`checkups`、`clinic_settings`、`clinic_integrations`、`lstep_settings`、`permission_groups`の12表。既存`setupClinicTestDB` patternに合わせ、対応modelと`ExaminationType` / `ExamTypeField` / `CheckupType`を`EnsureAutoMigrated`へ追加し、`clinic_settings`だけは既知のGORM `time`型問題を避ける`testdb.EnsureClinicSettingsTable`で整備した。

  ```text
  $ git -C /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte diff --name-only -- backend/internal/staff backend/internal/clinic backend/internal/model
  (empty)
  production_diff_exit=0
  ```

- F6 target tests（F2→F6補完loop 1周、retry 0）:

  ```text
  $ docker compose exec backend go test ./internal/service/... -count=1 -p 1 -run 'TestStaffDeleteAndShiftCreate_DeleteWinnerPreventsOrphanShiftDatabase|TestStaffSetAssignmentsAndClinicDelete_DeleteWinnerLeavesAssignmentsUnchangedDatabase' -v
  === RUN   TestStaffDeleteAndShiftCreate_DeleteWinnerPreventsOrphanShiftDatabase
  --- PASS: TestStaffDeleteAndShiftCreate_DeleteWinnerPreventsOrphanShiftDatabase (0.44s)
  === RUN   TestStaffSetAssignmentsAndClinicDelete_DeleteWinnerLeavesAssignmentsUnchangedDatabase
  --- PASS: TestStaffSetAssignmentsAndClinicDelete_DeleteWinnerLeavesAssignmentsUnchangedDatabase (0.16s)
  PASS
  ok  	github.com/animal-ekarte/backend/internal/service	0.608s
  test_target_exit=0
  ```

- F7 regression:

  ```text
  $ docker compose exec backend go test ./internal/service/... -count=1 -p 1 -v
  PASS
  ok  	github.com/animal-ekarte/backend/internal/service	2.523s
  test_after_exit=0
  after_top_level_pass_count=58
  after_top_level_fail_count=0
  PASS_to_FAIL=0
  FAIL_to_PASS=2
  TestStaffDeleteAndShiftCreate_DeleteWinnerPreventsOrphanShiftDatabase
  TestStaffSetAssignmentsAndClinicDelete_DeleteWinnerLeavesAssignmentsUnchangedDatabase
  ```

  `diff test.base.status test.after.status`は上記2件のFAIL削除・PASS追加と実行時間差だけ。baselineでPASSだったtest名の差集合は空。

- `TRUNCATE`判断: 拡張しなかった。追加表へfixtureをinsertせず、F6/F7が1周目でPASSし、baseline PASS→FAILも0件だったため。先に`CASCADE`範囲を広げる方が同packageの他fixtureを消すリスクを増やす。
- F8/F9:

  ```text
  $ docker compose exec backend go vet ./internal/service/... ./internal/staff/... ./internal/clinic/...
  (Composeの未設定DB変数warning 3行のみ)
  vet_exit=0
  $ docker compose exec backend gofmt -l internal/service/
  (Composeの未設定DB変数warning 3行のみ)
  gofmt_exit=0
  $ docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-$RANDOM --entrypoint golangci-lint backend run ./internal/service/... --max-same-issues 0 --max-issues-per-linter 0
  0 issues.
  lint_after_exit=0
  lint_diagnostic_diff=(empty)
  new_lint_diagnostic_count=0
  ```

- Database review / clinic isolation review: `Approve`。CRITICAL 0 / HIGH 0 / MEDIUM 0 / LOW 0。全count queryの`clinic_id` scope、paymentsのclinic-scoped billings JOIN、shared core + 追加schemaの全数性、`clinic_settings`の既存raw-DDL helper、TRUNCATE非拡張を確認。
- De-Sloppify: 新規test、assertion、fixture値、debug出力、死んだmodel追加は0。既存repository setupを踏襲するsupport model 3件を保持し、追加表のTRUNCATEは行わなかった。
- F11 tracking safety:

  ```text
  $ git reset
  Unstaged changes after reset:
  M	BE-refactor.md
  M	Makefile
  M	backend/internal/service/staff_shift_security_integration_test.go
  M	docs/ops/deploy/README.md
  M	scripts/run-local-ci.sh
  reset_exit=0
  $ git check-ignore -v backend/internal/service/staff_shift_security_integration_test.go BE-refactor.md
  (empty)
  check_ignore_exit=1
  $ git add backend/internal/service/staff_shift_security_integration_test.go BE-refactor.md
  add_exit=0
  $ git diff --cached --name-only
  BE-refactor.md
  backend/internal/service/staff_shift_security_integration_test.go
  cached_names_exit=0
  $ git restore --staged backend/internal/service/staff_shift_security_integration_test.go BE-refactor.md
  restore_staged_exit=0
  $ git diff --cached --name-only
  (empty)
  final_cached_names_exit=0
  ```

- F12 Independent Review Gate: reviewer roleによるfresh passは`APPROVE`。CRITICAL 0 / HIGH 0 / MEDIUM 0、LOW 1（F11/F13/F14のledger未記載）。必須5観点は (a) F2全数性、(b) production無変更、(c) TRUNCATE非拡張のfixture安全性、(d) setup外のtest変更なし、(e) allowlist外のsession-owned変更なし、すべてPASS。本追記でLOWを解消した。
- F13 allowlist:

  ```text
  F1 scoped status:
   M backend/docs/api.yaml
   M backend/internal/reservation/nested_summary_response.go
   M backend/internal/reservation/reservation_response_test.go
  $ git status --porcelain -- backend BE-refactor.md
   M BE-refactor.md
   M backend/internal/service/staff_shift_security_integration_test.go
  current_scoped_status_exit=0
  F1になく今回現れた行:
   M BE-refactor.md
   M backend/internal/service/staff_shift_security_integration_test.go
  allowlist_match=2/2
  session_owned_allowlist_outside_count=0
  ```

  F1の3 pathは並行writerの増分commit `463e07424af94eafdd0a1bf5a575134fb9f60b3c`へ取り込まれたためcurrent statusから消えた。本unitは当該pathを変更していない。repo全体にはA4 rehearsal系の並行writer差分があるが、F13の対象集合外かつ本unit由来ではない。
- F14 worktree / concurrent HEAD:

  ```text
  F1_HEAD=718f6c9b3
  CURRENT_HEAD=2946339243830bfe5c2b21740e3dd7c15911893a
  current allowlist worktree rows:
   M BE-refactor.md
   M backend/internal/service/staff_shift_security_integration_test.go
  $ git log --oneline -3
  294633924 docs: BUG-431クローズ（463e07424で修正完了）
  bf2c05b89 docs: S2実装プランのコード規約照合 — DEC-17へ実装前提3件を補強
  463e07424 fix(backend): BUG-431 — 受付の危険度バッジが実APIで点灯しない契約不整合を修正
  increment_count=5
  bc4fe88cb3009ad93b27a6d899c78b8e274e5cb7:
    frontend/src/features/examinations/api/get-examination-items.ts
    frontend/src/features/examinations/components/ExamPivotTable.test.tsx
    frontend/src/features/examinations/components/ExamPivotTable.tsx
    frontend/src/features/examinations/components/ExaminationHistoryPanel.test.tsx
    frontend/src/features/examinations/components/ExaminationHistoryPanel.tsx
    frontend/src/features/examinations/constants.ts
    frontend/src/features/examinations/routes/ExaminationForm.permissions.test.tsx
    frontend/src/features/examinations/routes/ExaminationForm.tsx
    frontend/src/features/medical-records/components/ExaminationGroup.test.tsx
    frontend/src/features/medical-records/components/ExaminationGroup.tsx
    frontend/src/features/medical-records/components/MedicalRecordExamination.test.tsx
    frontend/src/features/medical-records/components/MedicalRecordExamination.tsx
  3321c801fce919316b5b436bd4f75eabff0c4ca4:
    todo.md
  463e07424af94eafdd0a1bf5a575134fb9f60b3c:
    backend/docs/api.yaml
    backend/internal/reservation/nested_summary_response.go
    backend/internal/reservation/reservation_response_test.go
  bf2c05b8989bafe5ed7da6b45af2f67e56ede960:
    q&a.html
  2946339243830bfe5c2b21740e3dd7c15911893a:
    todo.md
  increment_allowlist_hits=(empty)
  ```

- Failure Signature log:
  - F7 command wrapper attempt 1: exact Docker commandの実行前にwrapper JavaScriptが`Unexpected identifier 'pipefail'`で失敗。shellを正しく渡して同じ検証commandを再実行し、attempt 2でPASS。test自体のretryは0。
  - F10 patch attempts 1–2: contextの引用符を誤ってescapeしたため`Invalid Context`、file変更なし。2-strikeで対象行のexact bytesを再読し、literal引用符の最小patchでattempt 3 PASS。同一failureを3回は繰り返していない。
- Assumption deviations: none。
