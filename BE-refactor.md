# BE-refactor 第10期（BE10）— backend規約適合（フォルダ構成）修正計画

## メタ

- 更新日: 2026-07-25
- 要件責任者: MinoruSoga
- baseline HEAD: `3ea5fb067`
- 位置づけ: backendフォルダ構成監査から確定した修正候補だけを扱うround-scoped plan。BE10完了時に本ファイルを削除し、完了履歴はgit履歴と実装時の検証資産を正本とする。
- 本期の業務目的: package境界の判断根拠と残存物の退役条件を明示し、同じ手動監査・誤検出・台帳探索を繰り返す工程を削除する。
- 本期は計画起票のみ。BE10-1〜BE10-6の実装、Go code、`.gitignore`、`q&a.html`、ADR、scriptsの変更は行わない。
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
- 状態: 完了（2026-07-25・下記実行ledger）。残余リスクは`closing_settings_request.go`の既存`staticcheck S1016` 2件だが、これは本unitが持ち込んだものではなく別unitで扱う。

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
- 修正内容: 2 packageの削除条件、担当、期限を決め、lint gate test 6件の移設先を確定する。移設先が確定するまで実働gateを削除しない。
- 検証方法: A4/A6を再実行し、移設後はlegacy production/test file 0、旧path import 0、移設先のscoped test PASSを確認する。
- severity: MEDIUM
- 前提・依存: BE9-1でsource discoveryがpackage非依存化済みであることを再確認する。test所在packageの変更で検出scopeが狭まらないことを移設先の自己テストで証明する。
- 状態: 判断待ち

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

- [ ] **AC-BE10-1 clinic境界**: BE10-1が「parent統合」または「ADR根拠追加」のどちらかで判断確定し、実装した場合はroute/RBAC/OpenAPI/runtime testを維持して`完了`となる。実装不能な外部判断が残る場合は、判断者・必要入力・再開条件を記録して`判断待ち`とする。
- [ ] **AC-BE10-2 legacy退役**: 削除条件・担当・期限・lint gate 6件の移設先が確定し、移設完了後にlegacy production/test/importが0である。未決なら必要なowner決裁と再開条件を記録して`判断待ち`とする。
- [ ] **AC-BE10-3 空directory**: 他session所有権確認後に15 directoryを削除し、同じ`find`が0件となる。
- [ ] **AC-BE10-4 ignore/削除**: `.ruff_cache`、`.wrangler`、空`.git`についてignoreまたは削除方針を確定し、選択した検証がPASSする。tool再生成の判断待ちは再開条件付きで記録する。
- [ ] **AC-BE10-5 docs drift**: `q&a.html`の5行を現行pathへ更新し、旧layer path grepが0件となる。
- [ ] **AC-BE10-6 package gate**: BE10-1/2の決定後に専用gateとmutation自己テストを追加し、実treeでPASSする。
- [ ] **AC-BE10状態収束**: BE10-1〜6がすべて`完了`または再開条件付き`判断待ち`で確定し、`未着手`/`対応中`が0件となる。
- [ ] **todo.mdへ戻す条件**: 判断が確定し直ちに実装可能になったunitだけを`todo.md`の「個別タスク詳細」へ一意なtaskとして移す。移管時は本書側を実行ledgerに更新し、同じ仕様本文を二重掲出しない。
- [ ] **round終了**: 全unitの収束、残余リスク、scoped verificationを記録後、本ファイルを削除する。恒久規約はADR-006/CODING_RULES、完了証跡はgit履歴とtest/gateを正本とする。
