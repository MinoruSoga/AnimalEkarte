# BE-refactor 第10期（BE10）— backend規約適合 active plan

## メタとcurrent snapshot

- 更新日: 2026-07-26
- 要件責任者: MinoruSoga
- execution snapshot: `2026-07-26T14:18:26+09:00`、branch `main`、full HEAD `b6fddfc20994de125fa7c10eec3a0fe8557bada8`
- 位置づけ: backendフォルダ構成監査で残った未完了unitだけを扱うround-scoped plan。BE10全体のround closeまでは本書を維持し、close後に削除する。
- 業務目的: package境界の判断根拠、legacy testの移設順、残存物の退役条件を一意にし、同じ手動監査・誤検出・台帳探索を繰り返す工程を削除する。

execution snapshotのlegacy root package実測:

| path | production Go | test Go | 状態 |
|---|---:|---:|---|
| `backend/internal/handler` | — | — | directory absent |
| `backend/internal/service` | 0 | 0 | Go surfaceは空だがdirectory退役条件は未充足 |
| `backend/internal/repository` | 0 | 25 | B9〜B14のactive manifestと一致 |

既存B8b relocation WIPは、移設元
`backend/internal/repository/reservation_owner_pet_preload_clinic_isolation_test.go`
が不存在、移設先
`backend/internal/trimming/reservation_owner_pet_preload_clinic_isolation_test.go`
が存在するworktree-applied/uncommitted状態である。このprovenanceからcommit hashは推定しない。

## AuthorityとSSOT ownership

- BE10のlive実行計画: **本書**
- 恒久的なpackage decision: [ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md)
- package boundaryの実測/provenance: [boundary map](docs/architecture/be9-2a-boundary-map.md)と[classification inventory](docs/architecture/be9-2a-classification-manifest.csv)
- backend作業規約: [backend/CODING_RULES.md](backend/CODING_RULES.md)
- 直ちに着手可能なtask: [todo.md](todo.md)
- 着手保留・任意検証: [BE-pending.md](BE-pending.md)
- PO/USER判断とrelease gate: [q&a.html](q&a.html)
- 完了証跡: git history、またはcurrent treeで検証されたuncommitted WIP

同じ仕様や完了記録を複数台帳へ複製しない。`3-session-agent.html`はguide viewでありSSOTではない。
Go/Ginはlayer-first/domain-first、Handler → Service → Repository、固定directory深さを規定しない。

## Active summary

| unit | current state | next condition |
|---|---|---|
| BE10-2 legacy test-only package退役 | 完了(2026-07-26) | B9〜B14全batch終了。service退役とrepository close gate(production/test/import 0・carrier 4不存在)を実測済み |
| BE10-3 空directory | 完了(2026-07-26) | 15 directory削除済み・指定`find` 0件を実測済み |
| BE10-4 ignore未登録残骸 | 決裁済み・USER実行待ち | 3残骸とも「削除」を採択(根拠と手動コマンドは同unit節)。実行後にpath不存在で検証 |
| BE10-5 `q&a.html` path drift | 完了(2026-07-26) | 5 hitを現行pathへ修正済み・旧layer path hit 0件を実測済み |
| BE10-6 package境界gate | 対応中 | 実装prompt `~/.claude/prompt-craft-runs/agent-be10-6.md` 生成済み・external agent実行待ち |
| round close | 未着手 | BE10-4のUSER実行検証とBE10-6完了後、SSOT routingを確認し本書削除 |

次の実行単位は **BE10-6**。B9〜B14は2026-07-26に全batch完了
(証跡=worktree WIPと各Completion Report・coordinator独立照合済み。履歴への反映は未実施でUSER判断待ち)。

## BE10-2 — legacy test-only package退役

### 実行境界と依存

- 全batchはmove、test単位split、実在domain APIへの直接接続、helper解消だけを行うstructural-only unitとする。behavior/security defectは同じbatchで直さず、`todo.md`等の正しい住所へ別taskとしてroutingする。
- B9は既存の`internal/testdb` fixture APIと、現行reservation/trimmingのpostconditionを前提にする。
- B10はB9の分割・helper解消後に実行する。
- B11はB10後に実行し、migration/runtime/payment-methodの既存gateを維持する。
- B12とB13は現行`internal/testdb`を前提にできるが、並列化せず先行unitの終了後に順番どおり実行する。
- B14はB9〜B13の全file移設、全consumer直接化、helper/facade consumer 0の後だけ実行する。
- 各batchでsource packageと全target packageをgateし、変更前後のtest-name集合とFAIL集合を比較する。clinic isolation、write owner、transaction、RLSに触れるbatchは該当専門reviewを追加する。

### Active manifest

current repository root test集合は、以下のB9=6、B10=4、B11=3、B12=3、B13=5、B14=4、合計25 fileと重複・欠落なく一致する。

#### B9 — cross-clinic / owner / pet / insurance（6）

| source file | target / action |
|---|---|
| `cross_clinic_preload_isolation_test.go` | test単位で`pet` / `owner` / `reservation`へ分割 |
| `insurance_repository_test.go` | `billing`へ移設 |
| `owner_pet_clinic_isolation_test.go` | test単位で`owner` / `pet`へ分割 |
| `owner_pet_create_write_owner_test.go` | test単位で`owner` / `pet`へ分割 |
| `owner_pet_relationship_preload_clinic_isolation_test.go` | test単位で`owner` / `pet`へ分割 |
| `pet_write_medimage_clinic_isolation_test.go` | test単位で`pet` / `medicalrecord`へ分割 |

`testdb.MakeInsurance` / `testdb.MakePetWithInsurance`へ直接化する。
`setupOwnerPetIsolationTestDB`はtargetごとに必要modelだけを準備するlocal setupへ分割する。
同file local helperは対象testと一緒に移し、repository facadeは実在constructorへ直接接続する。
clinic-isolation predicate、owner/pet write owner、preload/write契約を専門reviewする。

#### B10 — count / diagnosis / preload（4）

| source file | target / action |
|---|---|
| `count_clinic_scope_isolation_test.go` | test単位で`medicalrecord` / `reservation` / `billing`へ分割 |
| `diagnosis_repository_test.go` | `medicalrecord`へ移設 |
| `master_preload_clinic_isolation_test.go` | test単位で`medicalrecord` / `reservation`へ分割 |
| `preload_followup_clinic_isolation_test.go` | test単位で`medicalrecord` / `trimming` / `reservation`へ分割 |

diagnosis helper（`makeDiagnosisTypeMaster` / `makeDiagnosisNameRec`）はconsumerと同じunitで
`medicalrecord`へ移す。各testを対象domainへ分割し、local helperとconstructorを直接化する。
source discoveryはfile locationに依存させない。clinic-isolation predicateとtransaction境界を専門reviewする。

#### B11 — billing integrity（3）

| source file | target / action |
|---|---|
| `billings_hospitalization_unique_migration_test.go` | `billing`へ移設 |
| `billings_hospitalization_unique_test.go` | `billing`へ移設 |
| `payment_method_master_repository_test.go` | `billing`へ移設 |

hospitalization uniquenessのmigration gateとruntime gate、payment-method master gateを維持する。
SQL normalizer、DB setup、fixtureはconsumerと同じ`billing` test surfaceへlocalに移し、実在constructorへ直接接続する。

#### B12 — staff concurrency / security（3）

| source file | target / action |
|---|---|
| `staff_occupation_write_race_test.go` | `staff`へ移設 |
| `staff_shift_graph_atomicity_test.go` | `staff`へ移設 |
| `staff_update_security_test.go` | `staff`へ移設 |

`awaitStaffTestSignal` / `awaitStaffTestError`は2 consumerと同じunitで`staff`へ移す。
transaction atomicity、race ordering、credential/audit failure contract、clinic/staff isolationを専門reviewする。

#### B13 — DB / persistence / test schema（5）

| source file | target / action |
|---|---|
| `db_test.go` | `dbconn`へ移設 |
| `rls_effectiveness_test.go` | `persistence`へ移設 |
| `rls_role_privilege_test.go` | `persistence`へ移設 |
| `test_schema_enum_parity_test.go` | `testdb`へ移設 |
| `transactor_test.go` | `persistence`へ移設 |

`testDBConfig`、local setup、analyzerはconsumerと同fileで移す。
migration pathは同深度相対path(`../../migrations`)を維持する(B5b方式・test内nolintコメントが正本。module root解決の旧規定は2026-07-26 coordinator裁定で撤回 — B11/B13の移設で同方式のPASSを実証済み)。
RLS effectiveness/role privilege、ambient transaction、commit/rollback contractを専門reviewする。

#### B14 — carrier / facade retirement（4）

次の4 fileは、B9〜B13終了後、全consumerがtarget-domain APIまたは`testdb` APIを直接使用し、
constructor/facade consumerが0になったことを確認してから削除する。

- `be9_2c_r3_test_helper_carrier_test.go`
- `db_setup_test.go`
- `isolation_test_helpers_test.go`
- `target_repository_test_facades_test.go`

B14のclose gateはrepository rootのproduction/test/import 0、上記4 file不存在、target package gate greenに加え、
`internal/service`の退役条件も同時に再検証する。

### 退役条件・担当・期限

#### `backend/internal/service`

- current Go状態: production=0、test=0、import=0。
- directory未退役surface:
  - tracked `backend/internal/service/CLAUDE.md`
  - ignored `backend/internal/service/.DS_Store`
  - `CLAUDE.md` line 3の「残存する14 file」説明
- 退役条件: 上記2 fileとstale説明を明示的にdisposition/removalし、その後path不存在を確認する。Go count 0だけでは退役扱いにしない。
- owner: MinoruSoga
- deadline: 2026-07-31

#### `backend/internal/repository`

- 退役条件:
  - B9〜B14を順番どおり終了する。
  - 全target gate（enum / RLS / DDLを含む）がgreenである。
  - production/test/importが0である。
  - B14の4 carrier/facadeが不存在である。
  - root test packageが不存在である。空のchild directoryは別unit BE10-3で扱う。
- owner: MinoruSoga
- deadline: 2026-08-08

期限を超えた場合も条件を弱めない。対象を再開条件付き`判断待ち`へ戻し、blocker、未実行batch、再開条件を記録する。

## BE10-3 — 空の`internal/repository/*` 15 directory

current stateは次の15 immediate child directoryがemptyである。

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

shared-sessionの所有権と利用中pathを確認してから削除し、placeholderは追加しない。
同じ`find backend/internal/repository -mindepth 1 -maxdepth 1 -type d`が0件になることをclose gateとする。

## BE10-4 — ignore未登録のlocal artifact

execution snapshot:

| target | state | regular files | tracked | ignored |
|---|---|---:|---:|---|
| `backend/.ruff_cache` | present | 2 | 0 | false |
| `backend/.wrangler` | present（immediate child=`tmp/`） | 0 | 0 | false |
| repository-root `.wrangler` | absent（本unitのtargetではない） | — | 0 | false |
| `backend/.git` | present | 0 | 0 | false |

project `.gitignore`へ登録するか、不要な残骸として削除するかを決裁する。
本unitでは編集・削除しない。採択後は`git check-ignore -v`の根拠行、または`find`によるpath不存在で検証する。

2026-07-26 決裁(coordinator代理・3残骸とも削除を採択):

- `backend/.ruff_cache` = Go backendに残ったPython lint cache(2026-07-19生成の残骸)。ignore登録は残骸の固定化であり削除が正。
- `backend/.wrangler` = 空`tmp/`のみ。wrangler実行の正位置はworker側であり削除が正。
- `backend/.git` = regular file 0の迷子directory。nested repo誤認の危険源であり削除が正。
- 実行はUSER手動(AI実行環境は削除系および`.git` path操作が権限拒否): まず `find backend/.git -type f | wc -l` が0であることを確認した上で `rm -rf backend/.ruff_cache backend/.wrangler backend/.git`
- 検証: `ls -d backend/.ruff_cache backend/.wrangler backend/.git` が3件とも "No such file or directory"

## BE10-5 — `q&a.html`旧layer path drift

execution snapshotのhitはexactly 5件: lines `249`, `547`, `577`, `583`, `601`。
docs-onlyの独立unitで、業務契約や歴史を変えず次へ更新する。

| current line | target mapping |
|---:|---|
| 249 | `backend/internal/lstep/line_link_service.go` |
| 547 | `backend/internal/lstep/aggregation_handler.go` + `backend/internal/owner/ltv_repository.go` |
| 577 | `backend/internal/lstep/lstep_settings_thresholds.go` |
| 583 | related LSTEP tag files under `backend/internal/lstep/` |
| 601 | `backend/internal/owner/http_owner.go` |

本unitでは`q&a.html`を編集しない。実装unitでは旧layer path hit 0、各basenameの現行tree実在、該当cardの意味不変を確認する。

## BE10-6 — package境界専用gate

current treeにこのscopeの専用gateはない。`scripts/check-docs-symbol-drift.sh`はdocs参照のdrift gateであり代替ではない。

提案gateは次を検査する。

1. `internal/` top-level set
2. legacy production fileとproduction import edge
3. domain配下のlayer名subpackage
4. bucket名とlive package名の重複
5. 正常treeと各違反mutationを使うself-test

gateはADR-006のaccepted topologyをencodeし、承認済みのinfra adapter、flat `model`、
cross-cutting package、大規模`lstep`、`cmd/_archive`をfalse positiveにしない。
着手待ちはactive BE10-2 legacy退役だけであり、過去のBE10-1を前提に戻さない。

## 未監査の規約軸

本roundでは次を未監査として残す。

- boundary mapの許可依存graph全体
- Go codeの意味的review
- clinic isolation、認可、臨床安全、transaction境界の実装review
- frontend、infra、scripts、docsの構成

未監査は不具合または逸脱の存在を意味しない。

## Normative execution rules

1. 移設batchはsource packageと全target packageをgateし、変更前後のtest-name集合とFAIL集合を列挙する。shared DB suiteは`-p 1`で実行する。
2. PASS条件はsession-owned pathへscopeする。他writer所有のproduction状態やworktree全体のdiffへ束縛しない。
3. unit開始時にhelper consumerと`D ∩ C`を再scanする。raw string/comment内の擬似宣言・文字列pathを識別子consumerとして数えない。
4. 移設先packageの同名helper collisionとtest-local conventionを事前確認する。同一実装は既存helperを使い、異なる場合だけ局所的に改名する。
5. 最後のconsumerが去るtest-only bridge/carrier/facadeは同じunitで削除する。
6. third-party domain importからimport cycleが生じないか確認する。package import名がlocal変数をshadowする場合は、test変数の一括改名ではなく明示aliasで`importShadow`を避ける。
7. gate移設時は`Makefile`と`scripts/run-local-ci.sh`の明示package/`-run`配線を同じunitで更新し、0 testの`-run`成功をPASSにしない。
8. behavior/security/clinical findingや他ownerのFAILは移設へ混ぜず、一意な別taskへroutingする。

## Acceptance Checklist

- [ ] **AC-BE10-2 legacy退役**: B9〜B14を順番どおり終了し、service/repositoryの全退役条件を満たす。
- [ ] **AC-BE10-3 空directory**: shared-session所有権確認後に15 directoryを削除し、指定`find`を0件にする。
- [ ] **AC-BE10-4 ignore/削除**: 3つのbackend artifactについてignoreまたは削除方針を確定し、選択したgateをgreenにする。
- [ ] **AC-BE10-5 docs drift**: `q&a.html`の5 hitを現行pathへ直し、旧layer path hitを0件にする。
- [ ] **AC-BE10-6 package gate**: ADR-006 accepted topologyの専用gateとmutation self-testを追加し、実treeでgreenにする。
- [ ] **AC-BE10状態収束**: active unitがすべて終了、またはblockerと再開条件を持つ`判断待ち`となり、`未着手`/`対応中`を0件にする。
- [ ] **todo一意routing**: 判断が確定して直ちに実装可能なunitだけを`todo.md`へ一意に移し、移管した仕様本文を本書から削除する。
- [ ] **round close / file deletion**: 全unitの収束、残余risk、scoped verification、SSOT routingを確認後に本書を削除する。

## Round close

close条件は、Acceptance Checklistの全項目が満たされ、未解決事項が正しいSSOTへ一意にroutingされ、
active planと完了証跡の二重管理がないこと。恒久規約はADR-006/CODING_RULES、release gateは`q&a.html`、
完了証跡はgit history/current verified WIPをauthorityとし、条件充足後に本ファイルを削除する。
