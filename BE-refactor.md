# BE-refactor — Go/Gin公式ベースラインへのコード移行

> **ACTIVE (2026-07-19)**: 実行対象は下記BE9のみ。現行正本 = [`.claude/rules/go-gin-backend-guidelines.md`](.claude/rules/go-gin-backend-guidelines.md)、review正本 = [`.claude/refs/go-gin-backend-review.md`](.claude/refs/go-gin-backend-review.md)。
> **進捗 (2026-07-19)**: BE9-0 → BE9-2A → BE9-1 → BE9-2B（pilot=manualarticle）まで完遂。[ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md) Accepted。**成果物は全て未コミット** — `internal/lintscan`・`internal/httpapi`・`internal/manualarticle`の新設と旧layer側の修正が相互依存するため**部分コミット禁止**（lintscanを欠くとlint testがbuild不能になる）。次 = BE9-2C。着手前ゲートは「実装順序とbatch境界」の現在地節を参照。
> **BE8 SUPERSEDED**: 固定layer・層優先subpackage・repository→service→handler移行は [ADR-005](docs/architecture/adr/005-go-gin-backend-guidelines.md) により廃止。BE8-4/5/6/7の残作業は実行しない。旧本文は未コミット履歴の保全目的で残す。

## Active task: BE9 — 新コード規約をbackend実装・自動検査へ適用する（High）

### 目的と非目的

- **目的**: Go/Gin公式ベースラインを、今後の実装・review・CIで実効性のある状態にする。巨大なlayer packageをdomain/resource単位の凝集packageへ段階移行し、package境界、consumer-side interface、Context、Gin HTTP境界、error/security、server lifecycleをsemanticに検証する。
- **非目的**: 外部templateのfolder名を根拠なくコピーすること、全constructor/interfaceを一括renameすること、既存のGORM/apperrors/helperを理由なく置換すること。大規模なfolder変更自体は本タスクの明示的なscopeに含む。
- **不変条件**: clinic/owner/pet/staff分離、認可、監査、transaction atomicity、OpenAPI互換性を維持する。正本は [backend application invariants](.claude/refs/backend-application-invariants.md) とADR-002。

### 目標folder/package構成（project decision）

Go/Gin公式はdomain-firstやClean Architectureを指定しない。以下は、Go公式の「server codeは`internal`」「packageは利用者から見て凝集したAPI」「不要なinterfaceを先行作成しない」と、Gin公式のresource別route registrationをAnimalEkarteへ適用する**project固有の目標**である。**BE9-2Aで依存グラフを再実測し、[ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md)としてAccepted済み（2026-07-19）**。境界の正本 = ADR-006、per-domain詳細data（9列boundary map・許可依存グラフ・cycle解消方針）の正本 = [be9-2a-boundary-map.md](docs/architecture/be9-2a-boundary-map.md)。以下はAccepted版の要約である。

```text
backend/
  cmd/
    api/                    # executable entry point
    <other-command>/
  internal/
    # 13 target package（ADR-006 Accepted）:
    owner/ pet/ staff/ auth/ reservation/ trimming/
    medicalrecord/ billing/ inventory/ lstep/ clinic/
    manualarticle/          # BE9-2B pilotで移行完了済み
    httpapi/                # cross-resource Gin response/error helper（BE9-2Bで新設済み）
    # 現状維持（既存の凝集cross-cutting package）:
    config/ dbconn/ middleware/ infra/ model/
    timeutil/ seedbundle/ logger/ csvimport/ authjwt/ apperrors/ apicontract/
    lintscan/               # BE9-1で新設したlint共通source scanner
```

- domain package内では、HTTP、application logic、persistenceを必ず別subpackageにしない。同じ利用者・変更単位なら1packageの複数fileで構成する。
- resource routeは各domainが`Register...Routes(*gin.RouterGroup)`等で登録し、巨大な`handler.Handler`/`service.Services`/`repository.Repositories`集約への新規追加を停止する。
- domain間依存はconsumer側の最小interfaceで表し、implementation constructorはconcrete typeを返すことを基本とする。
- `internal/handler`、`internal/service`、`internal/repository`はmigration中のcompatibility facadeとしてのみ残し、BE9完了時にproduction implementationを0件にする。
- `model`の一括分割は先行条件にしない。GORM associationによるcycleを実測し、domain ownershipを安全に移せるtypeだけを同じbatchで移す。

### 実測根拠（当初根拠と解消状況）

- ~~`repoSourceFS`/`serviceSourceFS`の固定globが旧layer直下しか走査しない~~ → **BE9-1で解消済み**。lint 5本（preload_clinic_scope / audit_tx_inventory / dbortx_inventory / master_fk_write_inventory / n1）は`internal/lintscan`の`WalkInternalTreeT`へ委譲し、`internal/**`全treeを走査する。go:embed globは撤去済み。
- ~~active Go codeに旧architecture番号（P13/P14/P7等）の記述が残る~~ → **BE9-0で解消済み**。P<n>を合否条件に使うactive gateは改修前後とも0件（全て装飾ラベル）で、enforcing gate内の37件を意味化改名。残148件の分類正本 = [be9-0-legacy-gate-inventory.md](docs/architecture/be9-0-legacy-gate-inventory.md)。
- `cmd/api/main.go`は`gin.New()`、trusted proxies、security middleware、`http.Server` timeout、SIGINT/SIGTERM、timeout付き`Shutdown`、worker drainを実装済み。ここは全面再実装せず、公式review checklistとの差分だけを修正する（BE9-3で監査・現在も有効な前提）。

### 対応優先順位（大規模domain優先）

方針は**大きな課題を先に狙い、変更batchは小さくする**。旧`handler`、`service`、`repository`をlayerごとに順番に移すのではなく、1つのdomain/resourceについてHTTP・application logic・persistence・testを縦に移す。BE9-0、BE9-2A、BE9-1はfolder移動前の安全gate、BE9-2Bは移行手順を証明するpilot 1件だけであり、小規模domainを先に片付けるphaseにはしない。

規模は**BE9-2Aで再実測済み**（全761 production Go file、未分類0件。正本 = [be9-2a-classification-manifest.csv](docs/architecture/be9-2a-classification-manifest.csv) / boundary map）。旧filename-prefixベースの暫定集計は正本として継承しない — 最大の乖離はmedicalrecord（旧見積96 file → 実測185 file）。

| 暫定優先 | target domain | 実測file数 | 着手方針 |
|---|---|---:|---|
| 1 | medicalrecord | 185 | 最大domain。master CRUD→非確定処理→lab(saga)→finalize lock中核→hospitalization/billing接続の順（sub-batch案 = boundary map §3.7） |
| 2 | lstep | 106 | read/config/tag系から開始し、delivery/background writeは後段。line/liff内部分割はADR-006論点#2（BE9-2D着手直前に再確認） |
| 3 | reservation | 77 | query/master系から開始。**着手前にADR-006論点#1（reservation↔staff共有テーブル二重書き込み）の決裁が必須** |
| 4 | billing | 65 | read/calculationから開始。**着手時に`billing_item_repository.go`のUpdate/Delete防御ギャップ是正+クロステナントtest追加が必須前提（ADR-006論点#6）** |
| 5 | staff 31 / auth 25 / clinic 25 / trimming 23 / pet 21 / owner 13 / inventory 12 | 145計 | 大規模targetのcycle解消・依存解錠に必要な場合だけ先行し、解除後は大規模targetへ戻る |
| 済 | manualarticle 6 / httpapi 12 | 18計 | BE9-2B pilotで移行完了 |

暫定表を機械的な固定順にはしない。各batch開始時に、①target依存graphがacyclic、②tenant/認可/transaction/clinical safetyのbaseline testが存在、③route/API/SQL互換性とrollback単位が定義済み、の3条件を満たす**ready frontier**を作る。その中からproduction行数、file数、旧aggregator/call-site削減量が最大のdomain/subdomainを選ぶ（largest-ready）。小規模domainは、大規模targetのcycle解消・直接依存解除に必要な場合だけ先行し、解除後は直ちに大規模targetへ戻る。

### BE9-0: 旧規約の実効面をinventory化する

1. `.go`、`*_test.go`、CI、scriptから旧P1–P18、固定layer、特定helper名を合否条件にする箇所を`rg`で列挙する。
2. 各項目を`official Go/Gin`、`application safety invariant`、`project implementation detail`、`historical label`へ分類する。
3. security/behaviorを守るtestは削除せずsemanticな名前・messageへ変更する。folder形状・定義順・logging場所・interface数だけを強制するgateは廃止または根拠あるproject policyへ分離する。

**完了条件**: inventoryに対象file、現行の検出範囲、維持/改修/廃止判断、代替testが記録され、旧番号だけを根拠にfailするactive gateが0件。

**完遂（2026-07-19）**: 正本 = [be9-0-legacy-gate-inventory.md](docs/architecture/be9-0-legacy-gate-inventory.md)。scan 185件→148件（enforcing gate内の37件を意味化改名、検出ロジックは不変）。P<n>を条件分岐・比較に使うコードは機械探索で0件 — 全て装飾ラベルであり、**旧番号だけを根拠にfailするactive gateは改修前後とも0件**。machine-enforcedなshape-only gate（P13定義順・P14 layering等）も現行コードに存在しないことを確認。残148件の分類（Infra移行タスクID・PRレビューID等の別名前空間を含む）とBE9-4の残基準は同docに記録。

### BE9-1: safety lintをpackage非依存へ変更する

1. `repoSourceFS`/`serviceSourceFS`の固定globを、module配下の全production Go packageを発見できる共通scannerへ置換する。`go:embed **`が使えるという前提を置かず、`go/packages`またはmodule root基準のfilesystem walkを比較して選ぶ。
2. preload、hard-delete audit、ambient transaction、master-FK write、N+1 inventoryを、directory名でなくAST/type/data-pathの意味で判定する。
3. raw SQL、bare scalar FK、deep subpackage、background jobを静的検出できない場合は、blind spotをtest出力へ明記し、cross-tenant/atomicity runtime testを必須化する。部分的な静的解析を完全保証と表現しない。
4. scanner自身にroot package、1階層、2階層以上、別名package、testdata除外、raw SQL/bare scalarのfixtureを追加する。

**完了条件**: 同じ違反fixtureを`internal/repository`、`internal/service`、任意の`internal/<cohesive-package>`へ置いても同じ判定になり、新packageが旧folder外という理由で監査を回避できない。

**完遂（2026-07-19）**: `internal/lintscan`（`WalkInternalTreeT`）を新設し、lint 5本（preload_clinic_scope / audit_tx_inventory / dbortx_inventory / master_fk_write_inventory / n1）のsource discoveryを固定go:embed globから`internal/**`全tree walkへ置換。既存allowlistキーとの後方互換を維持（repository配下は旧repoSourceFS相対キー、他packageはinternal/相対キー）。scanner自体のfixture testは`lintscan_test.go`（testdata/vendor/`_test.go`除外を含む）。walk/AST破損時にvacuously greenで通過しない防御（最低検出数floor）を判定側へ追加。BE9-2Bで新設した`internal/manualarticle`等の新packageも走査対象に入ることを実地確認済み。

### BE9-2: domain/resource packageへ大規模移行する

#### BE9-2A: boundary mapとADRを確定する

1. production Go fileをpackage/import/call graphで再計測し、変更頻度、fan-in/out、transaction、route、tenant boundaryをdomainごとに記録する。旧BE8のfile prefixだけの分類結果をそのまま正本にしない。
2. `owner`、`pet`、`reservation`、`medicalrecord`、`billing`、`inventory`、`lstep`等について、target import path、owned types/routes/queries、consumer、許可する依存を一覧化する。
3. domain-firstはGo/Gin公式の必須構成ではなくproject decisionであること、代替案、cycle/security risk、段階移行方法をADR-006へ記録する。
4. migration batchごとにbaseline test、移動対象、call-site、compatibility facade、削除条件を固定する。

**完了条件**: 全production Go fileが「target package」「現状維持」「削除」のいずれかに割り当てられ、未分類fileが0件。target package間の許可依存graphがacyclicで、ADR-006がAccepted。

**完遂（2026-07-19）**: [ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md) Accepted。boundary map正本 = [be9-2a-boundary-map.md](docs/architecture/be9-2a-boundary-map.md)、分類マニフェスト = [be9-2a-classification-manifest.csv](docs/architecture/be9-2a-classification-manifest.csv)（761 file・未分類0・削除0）。13 target package間の許可依存グラフは45エッジでacyclic（機械検証済み）。生cycle 10組中9組は既存規約のconsumer-side interfaceで解消可能。**reservation↔staff共有テーブル二重書き込みの1組だけはinterface逆転で解決できず、設計決裁が必要**（ADR論点#1）。santa dual-review 3ラウンドで収束。副産物として`billing_item_repository.go`のUpdate/Delete防御ギャップを新規検出（GORMの`Joins()`がUPDATE/DELETEへ伝播しない罠。service層事前checkで現状は生きた漏洩ではないが、defense-in-depth不全+クロステナントtestゼロ。**未修正** — ADR論点#6としてbilling着手前ゲート化）。

#### BE9-2B: compositionと新規追加先を切り替える

1. 新規resource実装を`internal/handler|service|repository`へ追加することを停止し、target domain packageへ追加する。
2. resourceごとのroute registrationとconstructorをdomain packageへ用意し、巨大な`handler.Handler`、`service.Services`、`repository.Repositories`から順次切り離す。
3. `cmd/api`または必要最小限の`internal/server`でdomain moduleを組み立てる。DIを`main.go`だけに限定せず、型安全なcompositionを維持する。
4. package boundaryを跨ぐinterfaceはconsumer側へ置き、mock目的だけのimplementation interfaceを増やさない。

**完了条件**: 少なくとも1つの低結合domainが旧3package集約を経由せず起動・route登録・DB処理・testを完結し、以降の新規codeが同じ構成へ追加できる。

**完遂（2026-07-19）**: pilot=manualarticle（out-dom=0）。`internal/manualarticle`（repository+service+handler+routes、同一test名で緑）を新設。前提として`internal/httpapi`（target:httpapi 6ファイル: context_helpers/response/response_pg/bind_errors/time_response/slice_helpers）も切り出し、`internal/handler`側は269ファイルの既存呼び出し互換のため薄いdelegating facadeを残した（削除はBE9-2F）。`handler.Handler.RegisterRoutes`が`protected *gin.RouterGroup`を返すよう変更し、`cmd/api/main.go`がmanualarticleを`handler.Handler`/`service.Services`/`repository.Repositories`を経由せず直接組み立てて配線（Audit依存はconsumer-side `AuditLogger`interface+main.goのadapterで、Permission依存は`PermissionMiddleware`関数型注入で解消——auth domain未移行のため）。ADR-006をAcceptedへ昇格（6論点中5件はOpenのまま後続フェーズへ紐付けて記録）。次=BE9-2C/2D（reservation↔staffのcycle論点#1を先に決裁）。

#### BE9-2C: dependency-readyな大規模domainを優先migrationする

1. BE9-2Aのboundary mapからready frontierを作り、production行数、file数、旧aggregator/call-site削減量が最大のdomain/subdomainを選ぶ。低結合domainの全消化をcore着手条件にしない。
2. 選択domainのhandler/service/repository実装とtestを、route/use case/transaction単位の小batchで同じtarget packageへ縦移動する。複数の無関係な大規模domainを同一batchで変更しない。
3. file名は役割説明に使ってよいが、fileごとにsubpackageを作らない。package APIとして不要なsymbolをunexportする。
4. request-scoped処理は`c.Request.Context()`からDB/外部APIへ伝播し、error chainを`%w`と`errors.Is/As`で保持する。
5. composition切替と旧facade削除は別々にrevert可能なbatchにする。facadeはcall-site移行中だけ型alias/薄いdelegateとして許可し、削除期限を持たせる。

**BE9-2Cのdomain単位完了条件**: ready frontierから選択した1つの大規模domainについて、低〜中riskのvertical sliceがtarget packageへ移り、そのsliceの旧実装は期限付きfacadeだけになっている。旧packageに残るproduction implementationはBE9-2D対象として一覧化した高risk pathだけとする。API/SQL/tenant behaviorのbefore/after testが同一で、BE9-2Dのbaseline testとrollback単位が定義されている。

#### BE9-2D: 大規模domain内の高risk pathを段階移行する（BE9-2Cと反復）

1. LSTEPはread/config/tag処理を先に移し、outbound delivery、background job、外部writeを後段にする。
2. medical recordは非確定処理を先に移し、clinical finalize lock、addendum、audit fail-closedを最後にする。
3. reservationはquery/master系を先に移し、状態遷移、通知、appointment統合、transaction pathを後段にする。
4. billing/accountingはread/calculationを先に移し、write、締め、refund、atomicity pathを後段にする。
5. cycleは共有`common/util/interfaces` packageで逃げず、consumer interface、type ownership、event/value parameterのいずれかで解消する。`model` typeを移す場合はGORM association、migration/OpenAPI/codegen、全consumerを同一batchで検証する。

**BE9-2Dのdomain単位完了条件**: 選択domainの全高risk pathがADR-006の許可依存graph内から提供され、該当するclinical finalize lock、audit fail-closed、billing atomicity、clinic isolationのruntime integration testがPASSする。選択domainのproduction implementationとtestがtarget packageへ移り、旧package側には作業中batchの期限付きfacade以外を残さない。

**BE9-2C/2Dの全体完了条件**: 1domainごとにBE9-2C→BE9-2Dを完了してから次のlargest-ready domainへ進み、LSTEP/LINE、medical record/clinical、reservation、billing/accountingの4大規模domainがすべてdomain単位完了条件を満たす。

#### BE9-2E: 残る中小domainを規模順にmigrationする

1. 大規模domainのために先行移行済みの小規模dependencyを除き、残存production行数、file数、facade/call-site削減量が大きいdomainから処理する。
2. BE9-2Cと同じ縦移動、Context/error、API/SQL/tenant behavior、revert可能性のgateを適用する。
3. cross-cutting packageは実際の複数consumerがある場合だけ維持し、所有者が1domainへ収束したcodeはそのdomainへ移す。

**完了条件**: boundary mapで「target package」へ分類した全中小domainが移行済みで、未分類または移行期限のないfacadeが0件。

#### BE9-2F: 旧layer packageとfacadeを撤去する

1. 全call-siteをtarget domain packageへ変更し、期限切れfacade、巨大aggregator、旧layer専用helperを削除する。
2. shared helperは実際の複数consumerがあるものだけ、用途を表すpackageへ置く。`common`、`util`、`interfaces` packageを作らない。
3. docs、lint allowlist、test fixture、DI、route snapshot、OpenAPI symbol referenceを新pathへ同期する。

**完了条件**: `internal/handler`、`internal/service`、`internal/repository`にproduction implementationが0件。残すdirectory/fileがある場合は、domain packageへ置けない具体的consumer理由とADR-006の例外記録が必要。全target packageは単独test可能でimport cycleが0件。

### BE9-3: Gin HTTP境界・production lifecycleを公式checklistで監査する

1. route group/middleware scope、public/authenticated/authorized境界、`ShouldBind*` error処理、typed input validation、ownership、response/error contractをresource単位で監査する。
2. dependencyはclosureまたはstructで型安全に注入し、package global/untyped context injectionを新設しない。`main.go`だけを唯一のDI場所として強制しない。
3. trusted proxy失敗時の扱い、CORS/CSRF/cookie/rate limit/body limit、timeout値、`http.ErrServerClosed`、shutdown順序、goroutineの終了/cancelをdeployment前提と突合する。
4. `httptest`でbinding、validation、authn/authz、ownership、unknown 500、middleware abort/orderを検証する。

**完了条件**: [Go/Gin backend review](.claude/refs/go-gin-backend-review.md)の全項目にPASSまたは根拠付きN/Aがあり、cross-tenant requestと内部error非漏洩のnegative testがある。

### BE9-4: verification・移行完了

- scoped gate: 新scanner/packageと変更対象packageをDocker経由でtest/race/vetする。full `go test ./...`とfull lintは自動実行せず、最終gateとしてユーザー手動実行を依頼する。
- `rg -n 'go-package-conventions|gin-architecture-compliance|golang-gin-clean-arch' .`が0件。
- active code/testの旧P1–P18参照は、同名の別project phaseやhistorical fixtureを除き0件。例外にはsemanticな説明を付ける。
- `bash .claude/scripts/sync-agents-skills.sh`後に`.claude`と`.agents`のrules/skills差分が0件。
- `bash scripts/check-docs-symbol-drift.sh`、local Markdown link check、`git diff --check`がPASS。
- `BE-refactor.md`のBE9を完了化し、旧BE8本文は履歴として削除またはarchiveするかを別途判断する。

### 実装順序とbatch境界

`BE9-0 → BE9-2A → BE9-1 → BE9-2B（pilot 1件）→ {BE9-2C ↔ BE9-2Dを大規模domainごとにlargest-ready方式で反復} → BE9-2E → BE9-2F → BE9-3 → BE9-4`。BE9-1は新target packageを監査できる状態にしてからproduction migrationを開始する。BE9-3のresource監査は各BE9-2 batchでも反復する。大規模domainを先に狙うが、各batchはroute/use case/transaction単位とし、behavior-preservingな移動と機能変更を混在させない。security invariantを変更する必要が出た場合は本タスク内で推測せず、ADRとruntime isolation testを先に更新する。

### 現在地と着手前ゲート（2026-07-19）

**BE9-0 → BE9-2A → BE9-1 → BE9-2B まで完遂（全て未コミット）。次 = BE9-2C**。ゲートの正本 = ADR-006「未解決論点」節。以下は要約:

| # | ゲート | 発火タイミング |
|---|---|---|
| 論点#1 | reservation↔staff共有テーブル二重書き込み — staffを唯一の書き込み者にする案／reservation固有カラム分離案の**ユーザー決裁** | BE9-2Cでreservationまたはstaffへ着手する前（必須） |
| 論点#2 | lstepのline/liff内部分割の再確認（現決定=単一`internal/lstep`） | BE9-2D lstep内部分割の実装直前 |
| 論点#3 | clinic↔reservation/trimmingの営業時間制約依存が実測で確認できなかった件のドメインオーナー確認 | reservation/trimming着手時 |
| 論点#4 | `model/medicine.go`・`vaccine.go`のinventory vs medicalrecord帰属、`line_reservation_setting.go`のlstep vs reservation帰属 | 該当domain着手時に併せて確定（低影響） |
| 論点#6 | `billing_item_repository.go`のUpdate/Delete防御ギャップ — subquery形式是正+クロステナント分離test追加 | billing domain着手時（必須前提） |

（論点#5=間接isolation 3件はBE9-2A内で検証完了・決裁事項から除外済み）

### BE9-2B実績からの申し送り（以降の各batchで踏む）

- **handler側固定gateの追随を各batchに含める**: domain移行はroute snapshot（`handler/testdata/route_snapshot.golden`）とOpenAPI route drift（`apicontract/openapi_route_drift_test.go`）の更新を必ず伴う（BE9-2Bで実地確認）。batch計画時にこれら固定gateを事前列挙する。
- **cross-domain依存の解消パターン**: Audit依存 = consumer-side interface（`AuditLogger`）+ `main.go`のadapter、Permission依存 = middleware関数型注入（auth domain未移行のため）。BE9-2C以降も同型を使う。
- **旧layer側は薄いdelegating facadeで互換維持**（呼び出し側無変更）。facade削除はBE9-2Fまで持ち越し、削除期限を持たせる。
- **共有テストDBフレーク**（旧BE8 §5 row 6）はBE9でも生きている: `go test -p 1 ./internal/repository/...`で本batchが触れないファイルの赤は退行でない（pre-batch再現を確認して続行）。恒久対処=該当テストの`setupIsolatedTestDB`化はfollow-upのまま未着手。

---

## Superseded history: BE8（以下は実行禁止）

## 0. 要約

backend/ 全体を実測評価した結果、**問題は service / repository / handler の3層フラット肥大に集中**しており、他の 12 internal パッケージ・cmd/・worker/・migrations/・直下ファイルは健全（§1.5 で個別に判定根拠を明示 — 触らないこと自体が決定事項）。方針は「層優先 × ドメインサブパッケージ」を正式規約とし、strangler 方式で repository → service → handler の順に段階統一する。**一斉移動は禁止**。

**進捗（2026-07-19 時点）**: BE8-0/1/2/3/8（lint 網羅性固定・規約明文化・service 依存グラフ実測=§9・repotest 基盤・errors→apperrors）は完了。**BE8-4（repository 分割）は batch28 まで実施済み（サブパッケージ 42 ドメイン）。ただし残るフラット未分割実装 52 本のうち安全な葉はほぼ枯渇** — 大半が forbidden クラスタ（reservation/medical_record/accounting/LSTEP=40 本）・名指し除外（6 本）・強結合（pet/owner/staff/vital=4 本）で、境界マップ確定なしに着手できない。確実に残る安全な葉は `vaccination` ほぼ 1 本。BE8-5（service 202f）・BE8-7（handler 269f）は未着手で、いずれもクラスタ境界マップ確定を要する大物。

---

## 1. 現状実測

```bash
# 再実測コマンド（着手時に必ず実行）
cd backend
for d in internal/*/; do
  n=$(find "$d" -maxdepth 1 -name "*.go" | grep -vc _test); t=$(find "$d" -maxdepth 1 -name "*.go" | grep -c _test)
  echo "$d impl=$n test=$t subdirs=$(find "$d" -mindepth 1 -type d | wc -l) lines=$(find "$d" -name '*.go' | xargs cat | wc -l)"
done
```

**internal/ 主要パッケージ（行数は 2026-07-17 実測・パッケージ数/ファイル数は 2026-07-18 実測）:**

| パッケージ | impl + test | 行数 | 判定 |
|---|---|---|---|
| **service** | 202 + 202 | **131,093** | **是正対象（BE8-5・未着手）** — 完全フラット |
| **handler** | 269 + 206 | **95,040** | **是正対象（BE8-7・未着手）** — フラット（サブ dir は testdata のみ） |
| **repository** | 93(flat) + 164 | **53,221** | **是正対象（BE8-4・進行中）** — ドメインサブパッケージ **42 個** + repohelpers + repotest。フラット直下 93 本 = facade 41 + 未分割実装 52（内訳・残候補は §6 BE8-4）。**安全な葉はほぼ枯渇** |
| model | 85 + 18 | 5,751 | 現状維持（§8 — GORM モデルは FK 相互参照で分割すると cycle 不可避） |
| middleware | 9 + 8 | 2,343 | 健全 |
| infra | 7 + 2 | 1,450 | 健全（既にサブ dir 4 個で目標形） |
| apicontract | 1 + 2 | 1,232 | 健全（単一責務） |
| config | 2 + 2 | 525 | 健全 |
| **apperrors** | 1 + 1 | 447 | **健全（BE8-8 完了・2026-07-18 に errors→apperrors リネーム済み）** |
| csvimport / dbconn / logger / seedbundle / timeutil / authjwt | 各 1〜2 | 12〜257 | 健全（小さく単一責務。timeutil は用途限定名で `util` 禁止則に非抵触） |

repository サブパッケージ（2026-07-19 実測 42 ドメイン）: `account, animalspecies, audit, cage, campaign, checkup, checkupsync, checkuptype, chiefcomplaint, clinicholiday, clinicsettings, closingspecialperiod, company, consultation, dailyrecord, diagnosisname, diagnosistype, examtype, inquirytemplate, insurance, inventory, manualarticle, merchandiseitem, occupation, passwordreset, paymentmethod, prescription, procedure, refund, reservationtype, reservationtypeavailableslot, reservationtypegroup, reservationtypeunavailabletime, sharedfile, shiftentry, shifttemplate, staffclinicassignment, tokenblacklist, trimmingcourse, trimmingcoursetype, trimmingoption, vaccine` + `repohelpers`（scope.go / tx.go / junction.go）+ `repotest`（テスト基盤・BE8-3）。設計意図の先例 = `paymentmethod/repository.go` 冒頭コメント／real-DDL 例外の先例 = `audit/`／tenant-scope 置換の先例 = `consultation/`。

個々のファイルは最大 617 行で 800 行規約内。**問題はファイルサイズではなくパッケージ粒度**。

### 1.5 対象外と評価した領域（触らないことも決定事項）

| 領域 | 実測 | 判定理由 |
|---|---|---|
| `cmd/`（api, migrate, lstep-migrate, seed-export, stage-import, coverage-ratchet） | 6 バイナリ + `_archive`（underscore prefix で Go ビルド対象外） | 公式レイアウト準拠。変更不要 |
| `worker/`（index.ts ほか） | TypeScript の Cloudflare Worker ラッパ | Go スコープ外。配置は妥当（backend デプロイ単位に同梱） |
| `migrations/` | 001_init.sql + seeds | 独自規約あり（migrations/CLAUDE.md）。本計画のスコープ外 |
| backend/ 直下のバイナリ・成果物（`api`, `migrate`, `lstep-migrate`, `seed-old-db`, `stage-import`, `coverage.out`, `tmp/`） | **全て gitignored を確認済み**（git ls-files / check-ignore 実測） | 衛生問題なし。`seed-old-db` は対応する cmd/ が既に無い stale ローカルバイナリ — 見つけたら手元で消してよい（git 影響なし） |
| `backend/docs/`（api.yaml） | OpenAPI 正本 | 変更不要 |
| Dockerfile.dev / .production / entrypoint.sh / tygo.yaml / CODING_RULES.md | 設定・規約ファイル | 公式レイアウト上、非 Go ファイルのルート配置は正当 |

---

## 2. 調査結果（Option B 採用の根拠・§8 の再提案禁止決定が依拠する）

1. **Go 公式**（[go.dev/doc/modules/layout](https://go.dev/doc/modules/layout)）: サーバプロジェクトはロジックを `internal/` 配下に置き、公式例は `internal/auth/`, `internal/metrics/`, `internal/model/` の**ドメイン名パッケージ**。cmd/ にバイナリ。→ 本リポジトリは internal/ + cmd/ は準拠済み。巨大フラットパッケージは公式例の姿ではない。
2. **Google Go スタイルガイド**（[best-practices](https://google.github.io/styleguide/go/best-practices)）:
   - パッケージ名は機能を表すドメイン名。`util`/`helper`/`common` は不可。
   - **識別子でパッケージ名を繰り返さない**（stutter 禁止: `paymentmethod.NewRepository` であって `paymentmethod.NewPaymentMethodRepository` ではない）。
   - 分割基準 = 概念的に独立した機能は小さな専用パッケージへ。逆に「両方 import しないと使えない」なら統合が正。
3. **Gin/コミュニティ実勢**: `internal/{handler,service,repository,domain,middleware}` の層構成 + 小さく焦点の合ったパッケージ・consumer 側 interface 定義・非循環依存が合意。

---

## 3. 目標構成（決定）

**採用 = Option B: 層優先を維持し、各層内をドメインサブパッケージ化**

```
backend/internal/
  repository/
    repohelpers/        # 共有: clinicScope・DBOrTx（既存）
    repotest/           # テスト基盤（BE8-3・深さ1）
    reservation/        # ドメイン単位。package reservation
    accounting/
    ...
  service/
    servicehelpers/     # 必要になった時点で新設（先行して作らない — YAGNI）
    reservation/        # package reservation（service 側）
    ...
  handler/              # BE8-7: service 完了後に同方式で分割（testdata/ は共有のまま）
  model/                # 現状維持（§8 — 分割しない決定）
```

**旧命名規約（現行正本 = go-gin-backend-guidelines.md）**:
- パッケージ名 = 単数形・全小文字・アンダースコアなしのドメイン名。`util`/`common`/`helpers` 単独名は禁止（`repohelpers` は既存例外として存置）。
- 新規型は stutter 禁止（`reservation.Repository` / `reservation.NewRepository`）。**既存型の公開リネームは移動と同時にやらない**（呼び出し側全変更で diff が爆発する）— 移動時は型名維持、リネームは別コミットで機械的に。
- service ↔ service のドメイン間参照は **consumer 側で interface を定義**して受ける（import cycle の構造的回避）。

---

## 4. 移行方式 = strangler（一斉移動禁止）

- **新規ドメインは必ずサブパッケージで作る**（BE8-1 で規約化済み・即日発効中）。
- **既存フラットファイルは、そのドメインに実装変更が入るときに移す**。移動だけの巨大 PR を作らない。ただし BE8-4/5 の計画バッチ（葉ドメインから 5〜10 ファイル単位）は例外として許可。
- 完了の定義 = フラット直下に .go が残っていない状態（facade を最終的に剥がす）。期限は設けない。

---

## 5. 制約・地雷（着手前に必読・残バッチが踏む生きた制約）

| # | 地雷 | 実測根拠 | 対処 |
|---|------|---------|------|
| 1 | **自作 lint の走査範囲は「1階層まで」** — `repoSourceFS` の embed は `//go:embed *.go */*.go`。`repository/<domain>/*.go`（1階層サブパッケージ）は**カバー済み**。audit_tx / dbortx の 2 lint も同じ repoSourceFS を共有。**盲点 = ①2階層目以降（`<domain>/<subdir>/*.go` は 3 lint 全てから不可視）②service 側を走査する lint は存在しない** | `preload_clinic_scope_lint_test.go:35`。BE8-0 でライブ実証済み | **旧計画上の制約。現行のpackage設計ルールではない**。lintの実装上の走査範囲としてのみ扱う |
| 1-2 | **【生きた運用手順】3 lint の allowlist キーは「root=basename・1階層サブパッケージ=相対パス」に統一済み**（BE8-0 で preload/audit_tx を dbortx 方式へ統一）。**キーが統一されたことは「バッチがもうキーに触れない」を意味しない** — フラットファイルをサブパッケージへ移すと、そのファイルのキーが `staff_repository.go` → `staff/repository.go` のように **basename→相対パスへ必ず変わる** | dbortx の既存 allowlist に `reservationtype/repository.go\|repository.FindAll` 等の相対パスキーが実在 | **BE8-4/5 の各バッチで、移動したファイルの該当 allowlist エントリを basename→相対パスへ手動更新する**（移動を戻す必要はなくキーを更新すればよい）。この手順を BE8-4 手順テンプレ ④ に恒久保持 |
| 2 | **サブパッケージからテスト基盤が使える（BE8-3 で解決済み）** — `setupTestDB` 系は当初 `_test.go` 内定義でパッケージ外 import 不能だったが、`repository/repotest`（深さ1）へ抽出して export 済み | grep 実測・commit aa0dd6804 | 新規サブパッケージのテストは `repotest.SetupTestDB` 等を使う。先例雛形 = `paymentmethod/repository_test.go` |
| 3 | **DI 配線 = `cmd/api/main.go`**（`service.New*` を直接呼ぶ。他に `handler/auth_session.go`・`cmd/lstep-migrate/main.go` が service を参照）。移動バッチごとに main.go の import/呼び出しが変わる | grep 実測 | 各バッチで main.go を必ず diff 確認。コンストラクタは stutter 命名のまま — 移動時はリネームしない（§3） |
| 4 | **同一パッケージ内のドメイン間参照は import に現れない** — 分割して初めて cycle がコンパイルエラー化する | Go 言語仕様 | §9 の依存グラフで葉から抽出。cycle は consumer 側 interface で切る — **in-repo 先例**: `reservation_service.go:125` の `typeRepo reservationTypeFinder`（小文字ローカル interface）。この形を標準とする |
| 5 | パス参照の追随対象: directory CLAUDE.md・`docs/architecture/overview.md`・`.claude/refs/go-gin-backend-review.md`・scoped 検証規約・**top-level `backend/*.md`（README.md / CODING_RULES.md）**。**ci.yml は `backend/**` 一括フィルタのため追随不要**（確認済み） | `ci.yml:46`。BE8-8 で top-level backend/*.md の漏れが顕在化 | **旧計画の履歴**。現行review正本は `go-gin-backend-review.md` |
| 6 | **共有テスト DB フレーク** — `TestBillingItemRepository_FindUnbilledTrimmingItemsByPetID`（`appointment_trimming_options` FK 違反）等が累積テスト DB 状態で非決定論的に赤くなり、`go test -p 1 ./internal/repository/...` が単発で緑にならないことがある | batch26-28 で顕在化。pre-batch コードでも再現。repository/CLAUDE.md「共有テスト DB」節記載のフレーク級 | **本バッチが触れないファイルの赤は退行でない**（pre-batch A/B 比較 + 連続再実行で確認して続行）。恒久対処 = 該当テストの `setupIsolatedTestDB` 化（別タスク・follow-up。検証ゲートを毎バッチ毀損するため優先度高） |

---

## 6. タスク分割（この順で実行）

### 完了済み（実施ログは git 履歴が正本。生きた制約は §5・§9 へ移設済み）

- **BE8-0/1/2/3/8 ✅（履歴）** — lint 網羅性固定／旧規約明文化（現在は `.claude/rules/go-gin-backend-guidelines.md` に置換）／service 依存グラフ実測／repotest 基盤／`internal/errors`→`internal/apperrors` リネーム。詳細は git 履歴。

### BE8-4: repository 残りフラットファイルの段階分割 — **進行中（batch28 まで実施済み・安全な葉ほぼ枯渇）**

- **実施済み（batch1-28・commit は git 履歴）**: 計 42 ドメインをサブパッケージ化（一覧 = §1）。直近 batch20-28 = trimmingoption / procedure / audit / account / checkup / consultation / trimmingcourse / diagnosistype / diagnosisname。**罠の実例**: `account`（ログインアカウント・葉）と `accounting`（会計・forbidden）は別物／`trimmingcourse`（コースマスタ・葉）と `trimming`（AppointmentTrimmingDetail・reservation cluster）は別物。名前で判断せず実装を見る。
- **手順テンプレ（確定・各バッチで踏む）**:
  1. 新規サブパッケージへ実装+テストを新規作成（`Repository`/`repository`/`New` の**非 stutter 命名**）。テストは `repotest.SetupTestDB` 等を使う。**real-DDL 例外**: テストが 001_init.sql の実 DDL（`inet` 等 AutoMigrate が再現しない型）に依存する場合はフラットの real-DDL ハーネスに残置してよい（先例 = audit batch22 の `ip_address inet`・X-3 ドリフト回避。実装は facade 経由でフラットテストから実行されカバレッジ欠落なし）。
  2. `repohelpers.X` 直接呼び出し。フラット非公開ヘルパ置換（`medicalRecordTenantScope`→`repohelpers.MedicalRecordTenantScope` 等）は **package 修飾子のみで同一 SQL を発行**することを diff で確認。**paginate は共有ヘルパ未整備のため local-copy**（procedure/inventory/checkup/diagnosistype/diagnosisname に既存。rule-of-three 超過 → `repohelpers.Paginate` hoist は別バッチ・follow-up）。
  3. **旧フラットファイルは型名維持の facade 化**（`type XxxRepository = <domain>.Repository`・必要なら `var Fn = <domain>.Fn` 再エクスポート）で service/handler 呼び出し側を無変更に保つ。
  4. **【必須・§5 row 1-2】3 lint 該当エントリの allowlist キーを basename→相対パスへ手動更新**（例: `staff_repository.go|...` → `staff/repository.go|...`）。該当は移動ファイルが該当パターンを含む場合のみ。
  5. scoped 検証: `go test -p 1 ./internal/repository/...`（**`-p 1` 必須**）+ 3 lint（`-run` 関数名 prefix・`Lint` 単独不可）+ golangci-lint scoped（`docker compose run --rm --no-deps --entrypoint golangci-lint backend run --max-same-issues 0 --max-issues-per-linter 0 ./internal/repository/...`・**フレッシュキャッシュ必須**）。**判定は「本バッチ変更ファイルに新規 issue 0」**（既存 pre-existing debt 5 件=gofmt×4+wrapcheck×1 は対象外）。**§5 row 6 フレーク注意**: 本バッチが触れないファイルの赤は退行でない（pre-batch 再現を確認して続行）。
- **残 = 未分割実装 52 本。ただし安全な葉はほぼ枯渇**（内訳・2026-07-19 実測）:
  - **forbidden クラスタ 40 本**（repository/CLAUDE.md「Forbidden in drive-by tasks」・**境界マップ確定まで着手不可**）: reservation 系 8（reservation/schedule/staff/type_liff/type_occupation/appointment/appointment_admin/trimming）・medical_record 系 13（medical_record ±addendum/image/owner_visit・care_plan_item・clinical_plan・treatment±plan・examination・hospitalization±plan・inquiry・pet_chronic_condition）・accounting/billing 系 5（accounting/billing_confirmation/billing_item/cash_register_close/estimate）・LSTEP/line 系 14（lstep_* 10 + line_* 4）。
  - **名指し除外 6 本**: `ltv`（`kana_normalize.go` を owner/pet/medical_record と共有し切り出すと真の import cycle。unblock = `kana_normalize.go` を repohelpers へ hoist・別バッチ判断）・`lab_import`（3 コンストラクタ）・`permission_group`（5 ファイルで閾値超）・`medicine`+`medicine_dose_param`（#201 薬量計算隣接）・`clinic`（`isUniqueConstraintErr` 非公開ヘルパで move-not-modify 不成立・repohelpers hoist 別ステップ要）。
  - **強結合 4 本**（batch26-28 で却下）: `pet`/`owner`/`staff`/`vital`（medical_record/line/billing クラスタに結合）。
  - **統合推奨 1 本**: `checkup_field` は既分割 `checkup/` へ統合すべき（単独分割非推奨）。
  - **確実に残る安全な葉 = `vaccination` ほぼ 1 本**（batch28 で検証済み・次バッチ着手可）。他は上記のいずれかに該当。**即着手できる clean-leaf 在庫は事実上尽きた** — 以降の価値は上記ゲート付き作業（cluster 境界マップ・kana_normalize/paginate/isUniqueConstraintErr の hoist）と facade 剥がしに移る。
- **facade 剥がし方針**: 未着手・申し送り継続（41 facade をフラットから最終的に空にする段階。呼び出し側全変更の churn を要するため責任者判断）。

### BE8-5: service の段階分割（BE8-4 完了後・未着手）

- BE8-4 と同じ手順テンプレ。追加事項: ドメイン間参照は consumer 側 interface（§3）で切ってから移動する。cycle が出たら **移動を戻すのではなく interface 抽出で解決**する。抽出順は §9（sinks-first / reverse-topological）。**service を走査する自作 lint は存在しない**（§5 row 1）ため、BE8-5 開始時に「service にも同種 lint が必要か」を判断事項として起票する。
- **完了条件**: service フラット直下が空になる。

### BE8-6: ドキュメント同期（各フェーズ末に反復）

- `docs/architecture/overview.md`・`go-gin-backend-review.md`・`docs/spec/screens/` の該当 doc のパッケージパス記述を更新。`scripts/check-docs-symbol-drift.sh` green を確認する、という旧計画上の記録。

### BE8-7: handler 層の分割（BE8-5 完了後・未着手）

- **規模**: 269 + 206 test = 475 files / 9.5万行（3層で最多ファイル数）。サブ dir は `testdata/` のみ — 分割後も `testdata/` は共有位置に残す。
- **作業**: BE8-4/5 と同じ手順テンプレ。handler 固有の追加確認: ①ルート登録（`handler.go`・`master_routes.go`・`reservation_line_routes.go`）が全ハンドラを参照するため、バッチごとに登録側の import 更新が必須 ②handler 内 lint（`medical_record_image_handler_test.go`・`lab_report_handler_test.go` 等の allowlist 型テスト）の走査範囲を BE8-0 方式で先に確認 ③P5（RequirePermission 必須）等の handler 系 P ルールの検査機構がパッケージ分割に耐えるか確認。
- **完了条件**: handler フラット直下がルート登録ファイル・lint テスト・testdata のみになる。
- **判断ゲート**: BE8-5 完了時点の実測（ビルド時間・見つけにくさの実感）で「handler は分割せず現状維持」へ倒すことも許可する。その場合は §8 へ理由付きで移す。

---

## 7. 凍結条件（解除済み・履歴）

- **凍結解除済み（2026-07-18）**: Go-live 完了（ユーザー明示確認）+ PR #186 MERGED + main CI green の 3 条件成立。以後 BE8-3/4/8 の本番コードパスに触れる作業を実施。
- BE8-1（規約）・BE8-2（依存グラフ）は本番バイナリに触れないため凍結中（2026-07-17）に先行実施済み。

## 8. やらないこと（決定済み・再評価しない）

- **Option A（ドメイン優先の全面転換: `internal/reservation/{handler,service,repository}`）** — 理由: ①層別 CLAUDE.md・P1-P18 lint 体系・scoped 検証規約がすべて層パスを前提としており波及が桁違い ②repository の先行分割が Option B 形であり方向転換は二重の手戻り。再提案しない。
- **pkg/ ディレクトリ新設** — self-contained server binary であり公式ガイダンス上 `internal/` で完結（§2-1）。
- **model の分割** — GORM モデル 85 files は FK・Preload で相互参照しており、ドメイン分割すると model 間 import cycle が不可避。単一 `model` パッケージは go.dev 公式例（`internal/model/`）とも整合。5,751 行と軽量で実害なし。
- **§1.5 の健全領域への変更**（cmd/・worker/・migrations/・小規模パッケージ）— 触らないことが決定事項。改善提案が出たら①要件から検証する。
- **移動と同時の公開型リネーム** — diff 爆発防止。リネームは別コミット。

---

## 9. service ドメイン依存グラフと抽出順（2026-07-17 go/ast 実測・BE8-2 の出力／BE8-5 の入力）

### 9.1 実測手法

- 使い捨て go/ast スクリプト（scratchpad のみ・リポジトリ非コミット）で `internal/service/*.go`（実装 202 files）を 2 パス解析。
  - **Pass 1**: 全ファイルのトップレベル識別子（`Recv==nil` の func・type・var/const）→ 定義ドメインのマップを構築。**メソッドは名前衝突（`Create`/`Update` 等）で汚染するため除外**。
  - **Pass 2**: 各ファイルの USE 参照（`ast.Ident`）を走査し、宣言サイト（FuncDecl/TypeSpec/Field/ValueSpec 名）と `SelectorExpr.Sel`（`x.Foo` の `.Foo`）を除外。マップに存在し **かつ定義ドメイン ≠ 自ファイルドメイン** の参照のみをドメイン間エッジとして計上。
- ドメイン境界 = ファイル名 prefix。`reservation_type`/`reservation_staff`/`checkup_sync` は first-token だと親（reservation/checkup）に吸収されるため明示分離。残りは第 1 underscore トークン。全 69 ドメイン・773 エッジ。

### 9.2 計測の限界（正直な明記）

- **メソッド経由・interface 経由の依存は検出されない**（`f.svc.DoThing()` や §5 の `reservationTypeFinder` ローカル interface）。これは欠陥ではなく設計通り — 測っているのは「`git mv` でサブパッケージ化したとき**コンパイルを壊す構文的識別子結合**」であり、実行時配線ではない。**`out-dom=0` のドメインは、他ドメインへの残依存があってもそれは既に interface 化されている ⟹ 抽出しても cycle にならない**。
- 逆に **incoming（in-ref）はローカル変数名の衝突で過大計上され得る**。incoming は cycle 安全性ではなく「抽出時に import 追随が必要なファイル数の目安（churn）」として使う。

### 9.3 コンポジションルートの発見（抽出順の主軸が変わる根拠）

`service.go`（`Service` 集約 struct + `NewService`）は **59 ドメインの `NewXxxService` を参照する out-dom=59 の合成ルート**。このため実ドメインは全て「root からの incoming エッジ 1 本」を持ち、**被参照ゼロの純粋な葉は `service` 自身しか存在しない**。表中の `in-dom=1` の大半はこの root 配線であり、ドメイン間結合ではない。

→ 抽出安全性の主軸は **outgoing（依存の向き）**。抽出後 `service.go` は必ず `service/<D>` を import する（配線エッジ）。**D が親パッケージ残留の識別子を参照する（out-dom>0）と D→親 の逆 import が生じ cycle 化**する。従って **`out-dom=0` のシンクから逆トポロジカル順に抽出する**。

### 9.4 抽出段階（reverse-topological・sinks-first）

| 段階 | ドメイン | files | in(dom/ref) | out(dom/ref) | 根拠 |
|------|---------|:---:|:---:|:---:|------|
| **①雛形** | **daily** | 1 | 1 / 2 | **0 / 0** | 純粋シンク・in は root のみ・churn ゼロ。テンプレ実証用 |
| **①雛形** | **inventory** | 1 | 1 / 2 | **0 / 0** | 同上 |
| **①雛形** | **manual** | 1 | 1 / 2 | **0 / 0** | 同上 |
| ②共有カーネル | audit | 2 | 16 / 86 | **0 / 0** | out=0 で常時 cycle-safe。高 fan-in — 先に抜くと多数の依存元がシンク化し②以降を解錠 |
| ②共有カーネル | update (update_fields) | 1 | 7 / 8 | **0 / 0** | 同上（共有 update ヘルパ） |
| ②共有カーネル | timeslot | 1 | 4 / 64 | **0 / 0** | 同上（in-ref は概算） |
| ②共有カーネル | dose | 3 | 2 / 19 | **0 / 0** | 用量計算の純粋シンク |
| ②その他シンク | reservation_staff, token, species, account, shared, smtp, go | 各1〜2 | 低 | **0 / 0** | いずれも out=0・機械的に移せる葉 |
| ③解錠後の準シンク | validators (out-dom=2), company / chronic / clinical / care / refund / vaccination / prescription ほか out-dom=1 群 | 各1〜2 | 低〜中 | 1〜2 / 低 | ②抽出で残依存が消えるとシンク化。**validators は in-dom=37 の最大 fan-in — 残 2 依存を先に解決 or interface 化してから** |
| ④結合コア（最後尾） | lstep | 48 | 12 / 62 | 3 / 15 | out-dom=3 の準シンクだが **48 files — 単一バッチ不可。tag_sync / health_tag / delivery / settings / batch / csv でサブバッチ必須** |
| ④結合コア | liff | 11 | 3 / 4 | 5 / **80** | 高 outgoing（80 refs）。依存先を先に抽出 |
| ④結合コア | accounting(6) / reservation_type(9) / treatment(3) / trimming(4) / medicine(2) / estimate / appointment(3) | — | 中 | 4〜6 / 中 | 相互結合。依存解決後に |
| ④結合コア（**厳守最後尾**） | reservation | 4 | 9 / 34 | 5 / 15 | 高結合。in/out 双方大 |
| ④結合コア（**厳守最後尾**） | medical_record | 9 | 8 / 24 | 4 / 23 | finalized ガード（142f5ebe）の臨床安全コア。最後に移す |

> **注記**: `service`（`service.go`, in-dom=0）は合成ルートであり**親パッケージに残す — 抽出対象ではない**。表の `in-dom=1` は全て root 配線（依存結合ではない）。

### 9.5 最初の 3 バッチ（BE8-5 着手順・確定）

1. **daily**（`daily_record_service.go`）— out=0・被参照は root のみ。サブパッケージ雛形（BE8-3 の repotest 基盤）を service 側で実証。
2. **inventory**（`inventory_service.go`）— 同型の純粋シンク。手順テンプレの反復確認。
3. **manual**（`manual_article_service.go`）— 同上。3 本で「1 ドメイン = git mv + package 宣言 + service.go の import 1 行 + scoped test」の型を固める。

3 本完了後、**②共有カーネル（audit → update → timeslot → dose）** を抜いて多数の準シンクを解錠 → ③ → ④結合コアの順。lstep(48f) と medical_record(臨床安全) は最後尾。
