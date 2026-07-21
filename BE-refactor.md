# BE-refactor — Go/Gin公式ベースラインへのコード移行

> **ACTIVE (2026-07-19)**: 実行対象は下記BE9のみ。現行正本 = [`.claude/rules/go-gin-backend-guidelines.md`](.claude/rules/go-gin-backend-guidelines.md)、review正本 = [`.claude/refs/go-gin-backend-review.md`](.claude/refs/go-gin-backend-review.md)。
> **進捗 (2026-07-21)**: BE9-0 → BE9-2A → BE9-1 → BE9-2B（pilot=manualarticle・`d00c72a93`）→ **sub-batch①（`538cdb34`）→ ②（`14f00f6c`）→ ③（lab saga・`75c55c48`）→ ④a（vital/clinical_plan/medical_record_image・`e3eb253e`）→ ④b（treatment+dose kernel・Phase1=`cd8fd984d` WithTx化 / Phase2=`6508faab0` 縦移動）完遂**。checkup_sync系はlstep batchへ先送り（論点#7）。medicalrecord domainの残 = **⑤（hospitalization/discharge-with-billing）→（②③④の枠外: medical_record本体/examination/addendum等の残留高riskパスは⑤後に再定義）**。共有カーネル昇格batch = **完遂（`f93299f1c`）**。⑤ = **完遂（Phase1=`d4e227cf8` WithTx化／Phase2=`f024b09e7` 4エンティティ縦移動・72file）** — **BE9-2Dのmedicalrecord定義済みsub-batch①〜⑤は全消化**。repos.Transaction機構のproduction consumer=0（機構削除=BE9-2F）。BUG-418（DischargeWithBilling監査なし・既存負債）はtodo.mdバグ台帳起票済み。⑥残マスタ群 = **完遂（`a21977e91`・90file・2026-07-21）** — ④b⑤の暫定債務（transitional alias/carrier/Services.Hospitalization残置/handler carrier）を全解消・enum validator 4本sharedkernel昇格・discountガードpackage-level共用化。⑦ medical_record本体+addendum+examination = **完遂（`d4d7ef068`・84file・2026-07-21）** — **BE9-2D medicalrecord domain単位完了条件を充足**（旧layer package上のmedicalrecord系残実装ゼロ・残=facade群のみ→撤去はBE9-2F）。reservation Phase 0（論点#1案A実装=staffs/shift_entries書込のstaff domain一本化・クロステナントtest 5本mutation実証+fail-closedガード）= **完遂（2026-07-21）** — ADR-006論点#1へ実装記録追記済み。R①（reservation_typeマスタ群・`c4c95698d`・86file）= **完遂** — 縦移動sub-batch定義R①〜R⑥は「BE9-2D: medicalrecord sub-batch定義」節末尾参照。R⑥完遂（`0ee22c180`）= **reservation domain完了**。billing前提=BUG-417是正完了（`2634f58fe`）。**billing domain完了（Phase0+B①〜B⑥・`2634f58fe`〜`24420376c`）**。次 = **lstep（106file・⑧checkup_sync合流・論点#7の帰属裁定=lstep確定）→ BE9-2E（残る中小domain）→ BE9-2F（facade撤去）→ BE9-3 → BE9-4**。分割根拠と着手前ゲートは「現在地」節参照。push未実施。
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
| 1 | medicalrecord | 185 | 最大domain・**進行中**。①=済（`538cdb34`）→②=済（`14f00f6c`・checkup_syncは論点#7で除外）→③lab(saga)=済（`75c55c48`）→④a=済（`e3eb253e`）→④b treatment+dose kernel=済（`cd8fd984d`+`6508faab0`）→**共有カーネル昇格batch（次）**→⑤hospitalization/billing接続の順（詳細定義 = 本doc「BE9-2D: medicalrecord sub-batch定義」） |
| 2 | lstep | 106 | read/config/tag系から開始し、delivery/background writeは後段。line/liff内部分割は論点#2裁定済み（2026-07-20・単一`internal/lstep`確定、実消費境界出現時のみ再評価） |
| 3 | reservation | 77 | query/master系から開始。ADR-006論点#1は**裁定済み（2026-07-20・案A=staffを唯一の書き込み者に一本化）** — 着手ブロック解除。書き込み一本化sliceを移行の初手に置く |
| 4 | billing | 65 | read/calculationから開始。**着手時に`billing_item_repository.go`のUpdate/Delete防御ギャップ是正+クロステナントtest追加が必須前提（ADR-006論点#6・todo.md バグ台帳 BUG-417）** |
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

**進捗（2026-07-19・第1 slice完遂）**: largest-ready = medicalrecord（185 file最大。reservation=論点#1未決裁でハードブロック、billing=論点#6是正が前提でより小規模、lstepはDAG最上位で早期抽出不利のため対象外）。boundary map §3.7 sub-batch①（out-dom=0純粋master-CRUD: diagnosis type/name, examination type, chief complaint type）を`internal/medicalrecord`へ移動。

- **repository層**: BE8-4既存subpackage 4個（diagnosistype/diagnosisname/examtype/chiefcomplaint、いずれも旧42 subpackageの一部）をroll-up（削除しmedicalrecordへ統合）。generic `Repository`/`New`をentity-specific名へ改名（`DiagnosisTypeRepository`等）——外部から見える`internal/repository`側facade名は不変のため呼び出し側影響なし。`paginate()`ヘルパーの重複定義（diagnosistype/diagnosisname双方が独立複製していたもの）を1箇所へ統合。
- **service層**: 診断カテゴリ・診断名・検査種別・主訴種別の4サービス、DTO名は完全不変で移動。
- **handler層**: `DiagnosisHandler`/`ExamTypeHandler`/`ChiefComplaintHandler`の3 struct + 単一の`medicalrecord.Handler.RegisterRoutes`エントリポイント。**設計上の発見**: 当初per-entity複数`RegisterRoutes`を検討したが、`openapi_route_drift_test.go`の`buildFuncsFromDir`がbare名でfunc mapを構築するため、同名メソッドが複数struct上にあると2/3のroute setがdrift検知から静かに脱落する。単一エントリポイント必須と判明し設計を修正（BE9-2D以降も同じ制約に従うこと）。
- **共有ヘルパーの扱い**: `internal/service`の共有validation kernel（validators.go/validators_name.go/update_fields.go、9ドメインfan-in、boundary map §4.2「維持」）はmedicalrecordがimport禁止のため、pure/statelessな部分のみ`internal/medicalrecord/validators.go`へ複製（意図的な既知debt——boundary map §4.3のaudit kernel昇格と同型の解消を推奨、昇格先ができたら両コピーを統合）。`internal/handler/query_helpers.go`（parseIDParam/parsePagination/parseOptionalUint64Query/parseUUIDParam）はboundary map §3.13で既にtarget:httpapi認定済みだがBE9-2B pilotが未使用のため未移動だった箇所——本batchが初のconsumerとなり`internal/httpapi/query_helpers.go`+`pagination_response.go`として正式に切り出し完了（handler側は薄いfacade化）。
- **cross-package依存の新規発見**: `internal/handler/clinical_plan_response.go`（ClinicalPlan、sub-batch④・BE9-2D対象、本batch対象外）が`diagnosisTypeResponse`/`diagnosisNameResponse`をunqualified参照しており、medicalrecord内での想定外の密結合が判明。`DiagnosisTypeResponse`/`DiagnosisNameResponse`をexportし、`clinical_plan_response.go`が`internal/medicalrecord`をimportする形で解消（DTOの二重定義を回避）。**申し送り**: sub-batch②以降で他domainの旧handler fileが同様にmedicalrecordの response/request 型を直接参照していないか、着手時に確認すること。
- **旧facade方針の例外**: fan-in 0のservice/handler層は**facade化せず完全削除**（manualarticle先例のrepository facade完全削除と同型——型aliasで延命する理由がない場合はそうする）。fan-in>0のrepository層facade（clinical_plan_service.go/medical_record_service.go/examination_service.go/lab_import_examination_service.go/inquiry_service.goが依存）のみ型alias+delegate constructorとして維持。
- **test**: 旧`internal/service/{diagnosis,exam_type,chief_complaint}_service_test.go`と`internal/handler/{diagnosis,exam_type,chief_complaint}_{handler,request,response}_test.go`はmedicalrecordへ完全移動。旧`repository/{diagnosistype,diagnosisname}`のtestも移動。examtype/chiefcomplaintのフラット直下test（`internal/repository/exam_type_repository_test.go`・`chief_complaint_repository_test.go`）は**両方medicalrecordへ完全移動**——ただし`makeExamTypeMaster`/`makeExaminationRec`（`examination_repository_test.go`/`examination_repository_tx_atomicity_test.go`が共有、BE9-2D対象で本batch対象外）は旧ファイルにローカルコピーとして残置、`makeChiefComplaintType`は当初「`inquiry_repository_test.go`が共有」と誤判定（実際はコメント記述のみで実呼び出しなし、`golangci-lint`のunused検出で発覚・訂正）し旧`chief_complaint_repository_test.go`ごと削除（medicalrecord側に一本化）。`cross_tenant_master_fk_write_test.go`（複数domain共有の巨大test file）はexam_typeの2 FK guard testのみ切り出し（`NewExamTypeService`等がinternal/serviceから消えるため必須）、`mockDiagnosisTypeRepository`/`mockDiagnosisNameRepository`/`mockChiefComplaintTypeRepository`/`mockExamTypeRepository`の4モック構造体は他domain（ClinicalPlanService/InquiryService/MedicalRecordService/ExaminationService/LabImportExaminationService、いずれもBE9-2D対象）が`ok*Repo()`/`reject*Repo()`経由で使い続けるため定義元ファイルにコメント付きで残置。
- **gate追随**: `route_snapshot.golden`（25行削除）、`openapi_route_drift_test.go`の`migratedDomainRoutePackages`（`{dir: "../medicalrecord", prefix: "/api/v1"}`追加）、`master_fk_write_inventory_lint_test.go`の`serviceWriteRolePackagePrefixes`（`"medicalrecord/"`追加。副作用として`internal/medicalrecord`のhandler層が`*gin.Context`引数を伴いスキャン対象へ入り`knownSafeParamQualifiers`へ`"gin"`追加が必要になった——将来のBE9-2C/2D/2Eドメインも同じ制約に当たるため以後追加不要）+ allowlist該当4エントリのコメント更新——全gate green確認済み。他4 lint（preload_clinic_scope/audit_tx/dbortx/n1）はこの4ドメインに既存allowlist entryなし、basename→相対パス更新は不要だった（無違反ファイルのため）。
- **検証**: `internal/medicalrecord`・`internal/repository`（42 subpackage含む）・`internal/service`・`internal/handler`・`internal/httpapi`・`internal/apicontract`の全パッケージが単独実行でPASS（`gofmt -l`/`go vet`もclean、pre-existing debtとは無関係）。並行エージェント多数が同一`animalekarte-db-1`へ同時アクセスした状態での再実行時のみ、Owner/Treatment/TreatmentPlan等（本batch対象外domain）で`ERROR: deadlock detected`・行数不一致の非決定論的flakeを観測——単独実行で100%再現しないため§5 row 6と同型の環境要因と判定（詳細はCompletion Report）。
- 次 = BE9-2C sub-batch②（checkup/vaccine/prescription/inquiry等、非確定処理）または他domainのBE9-2C、あるいはmedicalrecordのBE9-2D（高riskパス）。

#### BE9-2D: 大規模domain内の高risk pathを段階移行する（BE9-2Cと反復）

1. LSTEPはread/config/tag処理を先に移し、outbound delivery、background job、外部writeを後段にする。
2. medical recordは非確定処理を先に移し、clinical finalize lock、addendum、audit fail-closedを最後にする。
3. reservationはquery/master系を先に移し、状態遷移、通知、appointment統合、transaction pathを後段にする。
4. billing/accountingはread/calculationを先に移し、write、締め、refund、atomicity pathを後段にする。
5. cycleは共有`common/util/interfaces` packageで逃げず、consumer interface、type ownership、event/value parameterのいずれかで解消する。`model` typeを移す場合はGORM association、migration/OpenAPI/codegen、全consumerを同一batchで検証する。

**BE9-2Dのdomain単位完了条件**: 選択domainの全高risk pathがADR-006の許可依存graph内から提供され、該当するclinical finalize lock、audit fail-closed、billing atomicity、clinic isolationのruntime integration testがPASSする。選択domainのproduction implementationとtestがtarget packageへ移り、旧package側には作業中batchの期限付きfacade以外を残さない。

**BE9-2C/2Dの全体完了条件**: 1domainごとにBE9-2C→BE9-2Dを完了してから次のlargest-ready domainへ進み、LSTEP/LINE、medical record/clinical、reservation、billing/accountingの4大規模domainがすべてdomain単位完了条件を満たす。

##### BE9-2D: medicalrecord sub-batch定義（boundary map §3.7①-⑤の詳細化・2026-07-19）

sub-batch①（master-CRUD: diagnosis type/name, examination type, chief complaint type）はBE9-2Cで完遂済み（`internal/medicalrecord`）。以下は②-⑤の高risk path一覧・baseline test・rollback単位。**定義のみ・実装はBE9-2D本体（別unit）**。file一覧は代表例——完全なfile集合は着手時に`docs/architecture/be9-2a-classification-manifest.csv`のtarget:medicalrecordエントリを当該token（checkup/vaccin/prescription/inquiry等）でgrepして確定すること（本docは一覧の網羅を保証しない）。

**②checkup/vaccine/prescription/inquiry（非確定処理・lstep通知はinterface経由のため抽出を妨げない）**

- 対象file代表例: `internal/service/{checkup,checkup_field_result,checkup_type,vaccine,vaccination,prescription,inquiry,inquiry_template}_service.go` + 対応`internal/repository/*_repository.go`（vaccine/vaccinationは既にsubpackage化されていない場合あり、着手時に`internal/repository/vaccine/`等の存在を確認）+ `internal/handler/{checkup,vaccination,prescription,inquiry}_*.go`。manifest上のcheckup系29 file・vaccin系13 file・prescription系6 file・inquiry系11 fileが母集団（handler/service/repository/test合算、正確な内訳は着手時にmanifestで再集計）。
- **baseline test（移動前green必須）**:
  - finalize-lock保護: `TestCheckupService_Create_FinalizedRejection` / `TestCheckupService_Update_FinalizedRejection` / `TestCheckupService_Delete_FinalizedRejection`（`internal/service/checkup_service_test.go`）、`TestPrescriptionService_Create_FinalizedRejected` / `_Update_FinalizedRejected` / `_Delete_FinalizedRejected`（`internal/service/prescription_service_test.go`）。**vaccinationはガードなし**（boundary map §3.7既知）——`TestVitalService_*`同様、移動時に新規finalize-lock guardを追加してはならない（behavior-preserving、機能追加は別issue）。
  - clinic isolation: `TestVaccinationRepository_FindByID_VaccinePreloadClinicIsolation` / `_SameClinicVaccinePreloaded`（`internal/repository/vaccination_master_preload_clinic_isolation_test.go`）。checkup/prescription/inquiryの専用clinic-isolation test fileは未確認——`preload_clinic_scope_lint_test.go`（package非依存walk、§3.7既存記載）のgreenが機械的代替網。
  - master-FK-write fail-open防止: `internal/service/master_fk_write_inventory_lint_test.go`の`inquiryService.Save`エントリ（test: `TestInquiryService_Save_RejectsCrossClinicChiefComplaintType`、`internal/service/cross_tenant_master_fk_write_test.go:2224`）——inquiry移動時は`serviceWriteRolePackagePrefixes`の`"medicalrecord/"`が既にBE9-2Cで追加済みのため新規prefix追加不要、allowlist keyの受信先コメント更新のみ。
  - TOCTOUギャップ既知（§3.7既存記載）: checkupは`lockDraftMedicalRecord`未使用のTOCTOU gapが移動前から存在——移動でこのgapを拡大/縮小しないことをbaseline確認の対象とする（新規安全化はBE9-2D本体の別issue、本batch=移動のみでは変更しない）。
- **rollback単位**: (i) composition切替batch — `internal/medicalrecord`への実装+test移動、旧`internal/service|handler|repository`側は期限付きfacade化（fan-in 0なら完全削除、BE9-2C同様）。(ii) facade剥がしbatch — 残る呼び出し側（存在すれば）の直接import切替、BE9-2Fへ持ち越し。2つは別コミットとし、(i)のみでrevert可能な状態を維持する。
- **前提条件**: なし（reservation/billing依存なし、ready frontier上は即着手可）。inquiryのChiefComplaintTypeID FK guardは既にBE9-2Cで`internal/medicalrecord`のChiefComplaintTypeRepositoryへ依存済みのため、移動時はimport pathの単純化のみ。

**完遂（2026-07-20・3batch順次実行: A=repository→B=service→C=handler+配線、各batch間もコンパイル可能な中間状態を維持）**:

- **スコープ確定時の判断2件**: (1) **checkup_sync系は除外しlstep batchへ先送り**（service 4 file+handler 3 file+repository checkupsync/。依存の実質がlstep domain: `lstep.Client`直import・CPM純関数kernel（複製はコード内コメントが明示的に禁止）・owner/pet/tagCache repo・route自体が`/clinics/:clinic_id/lstep/checkup-sync`。本docの②代表file一覧にも元々不在——論点#7として現在地節に登録）。(2) `internal/service/lstep_tag_sync_vaccine.go`はmanifest分類誤り（`*lstepTagSyncService`のmethod分割file・兄弟3 fileは再分類済み）→ target:lstepへ訂正。この2判断でowner/pet/tagCache/lstepSettings依存が消滅し、cross-domain contractは5系statement（medicalRecordFinder/Locker・Transactor・AuditTxLogger・lstep tag/trigger narrow interface群）に収束。
- **移動実績**: service 8本（checkup/checkup_field_result/checkup_type/vaccine/vaccination/prescription/inquiry/inquiry_template、receiver名・DTO名不変）、repository 8本（subpackage 5個roll-up+flat 3本、`dbOrTx`→`repohelpers.DBOrTx`置換）、handler 8 entity群（checkup_fieldはCheckupHandlerへ統合）、37 route（RBAC parity逐語転記: vaccination per-route=BUG-125、checkup-types=ResourceCheckups例外、/medical-records配下=ResourceMedicalRecords=BUG-133）、test群（cross_tenant 5 entity分はmedicalrecord単一fileへ集約——6 helperがunit test共有のためper-entity分割せず）。
- **契約と配線**: consumer-side契約=`medicalrecord/service_deps.go`。`cmd/api/main.go`が8 serviceを直接構築（aggregator非経由・Transactor新インスタンス・`svcs.Audit`のcomma-ok assert=fail-fast起動検証・audit adapterはmanualarticle先例）。graceful shutdownの`svcs.Checkup.Wait()`はmain.goローカル`checkupSvc.Wait()`へ（同一インスタンス・順序不変）。
- **gate追随**: route golden 447→410行、medicalrecord snapshot 25→62 route、dbortx/audit_tx allowlistキー`medicalrecord/`化、master_fk reason文字列path更新（キーはreceiver名維持で不変）、**openapi_date_format_drift_test.goのscan dirを`../medicalrecord`へ拡張**（responseScanDirs新設——response file移動でdrift entryがstale化する罠、route driftのmigratedDomainRoutePackagesと同型の恒久対処）。docs-symbol-drift green。
- **一時的gate弱体化の復元を確認**: Batch B中間状態で`knownSafeParamQualifiers`へ追加した`"medicalrecord"`はBatch Cで除去済み・除去後lint green実測（中間状態でgateを緩めた場合は復元をbatch完了条件に含める——本batchで確立した規律）。
- **検証**: baseline 9本の移動前green→移動後同名green（`=== RUN`確認）、変更5 package（medicalrecord/repository/service/handler/apicontract）のDB-backed全数test `-p 1`でgreen、build/vet（`./...`全体）/gofmt/`git diff --check` clean。敵対レビュー（4レンズ+反証検証）の結果はコミットメッセージ参照。
- **残置**: mock carrier=`internal/service/be9_2d_mock_carriers_test.go`（残留lstep/liff/vital等のtestが使う6 mock。carrier解消は各domainのBE9-2C/2D時）。repository facade alias 8本（fan-in>0のため。削除=BE9-2F）。

**③lab_import/lab_report（sagaパターン・単一tx wrappingではなくper-row partial success）**

- 対象file代表例: `internal/service/{lab_import_examination,lab_result_import,lab_import,lab_report_query,lab_audit_logger}_service.go` + `internal/repository/lab_import_duplicate_checker*.go`等。manifest上lab_import系6 file・lab_report系3 fileが母集団。
- **baseline test（移動前green必須・saga挙動の証拠）**:
  - 部分失敗耐性: `TestLabResultImportService_Commit_RowError` / `TestLabResultImportService_Commit_AllFailed` / `TestLabResultImportService_Commit_WithDuplicate` / `_AllDuplicate`（`internal/service/lab_result_import_service_test.go`）。
  - clinic isolation: `TestLabResultImportService_Commit_ClinicScopeEnforced`。
  - キャンセル・補償遷移: `TestLabResultImportService_Commit_ContextCancelledDuringPersist` / `_CompensationTransitionAlsoFails`（saga補償ロジックの中核、移動時に最優先でgreen確認）。
  - 出所制限（manual source blocked等、#安全ガード）: `TestLabResultImportService_Commit_NonFixtureBlocked_DrWan` / `_NonFixtureBlocked_Manual`、`internal/service/lab_audit_logger_test.go`の`TestLabAuditLogger_LogSourceBlocked_*`系7 test。
  - master-FK-write: `labImportExaminationService.PersistBatch`（allowlist、test: `TestLabImportExaminationService_PersistBatch_RejectsCrossClinicExamType`）・`labResultImportService.Commit`（test: `TestLabResultImportService_Commit_RejectsCrossClinicExamType`）。
- **rollback単位**: sub-batch②と同型（composition切替→facade剥がしを別コミット）。lab importはexam_type（sub-batch①で既にmedicalrecord内）に依存するため、①完了後は追加の外部依存なし。
- **前提条件**: sub-batch①（済）。reservation/billing依存なし。

**完遂（2026-07-20・コミット`75c55c48`・3batch順次A/B/C）**: lab 19 file（service 5+repository 1+handler 5+test）をmedicalrecordへ縦移動。**labはleaf domainでfacade残置ゼロ**（外部fan-inがservice.go/repositories.go/handler.goのaggregator配線3点に集約・移動で全消滅→完全削除。sub-batch②の8 facade残置とは対照的）。**確立/再確認した知見**: ①leaf domainでもrepository層は中間状態で一時facade必須（Batch A単独ではservice/handlerがまだrepository.LabImport*参照→build不能。Batch Bで参照消滅と同時に完全削除しleaf原則をdischarge）②saga補償（per-row partial success・context.WithoutCancel補償遷移・孤児exam Delete掃除P2-7）はtx primitive非依存のためconsumer-side interface（examinationImportRepoにDelete method含む=補償の消失防止）越しの通常呼び出しで挙動不変移植可③非tx監査（LogEntry best-effort）用にAuditTxLogger（tx版）とは別の非tx AuditLogger view新設+main.go adapter④中間状態でlint緩めた"medicalrecord" qualifierはBatch Cで除去しゲート復元⑤computeExamResultStatus複製で共有helper複製が積み上がる（**ただし④a inventory実測でコピー数は2＝rule-of-three(3コピー)は誤り。昇格の真の根拠はcaller恒久ドメイン境界跨ぎ。共有カーネル昇格は④に含めず④後の専用batchへ**——判断詳細は「現在地」節の④分割決定を参照）。cross_tenant lab 5本はmedicalrecordの集約testへ（persistExam型assertionで実装同時移動）。ptr[T]はaccounting/cash_register test依存のためinternal/serviceに残置コピー。検証: baseline 27本green→移動後medicalrecord側同名green（=== RUN確認）・変更5 package DB-backed全数test -p 1 green・build/vet(./...)/gofmt/docs-symbol-drift/git diff --check clean・敵対レビュー4レンズ+反証。
- **前提条件**: sub-batch①（済）。reservation/billing依存なし。

**④treatment/vital/clinical_plan（`lockDraftMedicalRecord` row-lock中核・抽出時にA全体="診断/検査/処方/lab"が先に必要）**

- 対象file代表例: `internal/service/{treatment,vital,clinical_plan,medical_record_lock,medical_record_image}_service.go` + `internal/repository/{treatment,vital,clinical_plan}_repository.go` + `internal/handler/{treatment,vital,clinical_plan}_*.go`。manifest上treatment系11 file・vital系5 file・clinical_plan系5 fileが母集団。
- **baseline test（移動前green必須・row-lock中核のため最重要）**:
  - `lockDraftMedicalRecord`本体: `TestLockDraftMedicalRecord_NilParentFailsClosed`（`internal/service/medical_record_lock_test.go`）——fail-closed契約の唯一の直接test、移動前後で必ずPASS。
  - finalize-lock保護（per-domain）: `TestClinicalPlanService_Update_RejectsFinalizedParent` / `_Delete_RejectsFinalizedParent`（`internal/service/clinical_plan_service_test.go`）、`TestMedicalRecordImageService_FinalizedGuard`（`internal/service/medical_record_image_service_test.go`）。**treatment/vitalには対応する`Finalized`/`Lock`named testが存在しない**（grep実測でゼロ件確認、§3.7既存記載と整合——finalize-lock保護は`lockDraftMedicalRecord`呼び出し自体で担保され専用regression testが無い状態。移動時に新規追加してはならない=behavior-preserving、追加は別issue）。
  - clinic isolation: `TestTreatmentRepository_FindByMedicalRecordID_MasterPreloadClinicIsolation` / `_SameClinicMasterPreloaded`（`internal/repository/treatment_master_preload_clinic_isolation_test.go`）。
  - master-FK-write: `medicalRecordService.CreateSubRecords`のChiefComplaintTypeID/Diagnosis1-2系allowlistエントリ（`internal/service/master_fk_write_inventory_lint_test.go`、test: `TestMedicalRecordService_CreateSubRecords_RejectsCrossClinicChiefComplaintType` / `_RejectsCrossClinicDiagnosisFK`）——この移動でmedicalrecordパッケージ内のvalidateCreateSubRecordDiagnosisFKs相当ヘルパーも追随させる必要あり（診断FK 4スロット全て）。
- **rollback単位**: sub-batch②③と同型。ただし`lockDraftMedicalRecord`は複数domain（medicalrecord A+B, billing）からfan-in=20/8file/3domainのため、単一batchでの移動はrisk大——**推奨分割**: (i) `medical_record_lock.go`本体+`TestLockDraftMedicalRecord_NilParentFailsClosed`のみ先行移動しfacade化、(ii) treatment/vital/clinical_planの各serviceを個別batchで追随（それぞれ独立してrevert可能）。billing（billing_confirmation_service.go/estimate_service.go）からの呼び出しは論点#6是正前提のため本sub-batchでは触れない。
- **前提条件**: sub-batch①②③（全domain内依存解決後）。billing側の`lockDraftMedicalRecord`呼び出し元（billing_confirmation_service.go/estimate_service.go）は移動しない（billing domain着手時=論点#6是正後）——A domain内移動のみで billing→medicalrecord逆依存を新設しない。

**完遂（④a=2026-07-20・`e3eb253e`／④b=2026-07-21・Phase1=`cd8fd984d`+Phase2=`6508faab0`）**: ④aの実績詳細は「現在地」節。④b（treatment+treatment_dose_save+dose kernel）は2フェーズ2コミット:

- **Phase 1（`cd8fd984d`・behavior-preserving refactor・移動なし）**: treatmentServiceの`*repository.Repositories`集約+repo-swap tx機構（`repos.Transaction`/`NewRepositories(tx)`）を、他service同様のctx-txKey機構（`Transactor.WithTx`+各repoのdbOrTx参加）へ挙動保存で変換。実消費repo 8本+Transactor+AuditTxLoggerの個別注入化、treatment_repository Create/Update/DeleteのdbOrTx化、**medicalrecord vitalRepository.FindByMedicalRecordIDのDBOrTx化**（dose体重解決が保存tx内から読むread——旧機構ではtx-bound cloneで暗黙にtx参加していた分の等価維持=#201 TOCTOU防止）、逸脱監査のtxRepos.Audit.CreateTx直呼び→AuditTxLogger.LogEntryTx統一（R1-2 (D1)・medicine/dose-param同型・監査行内容field単位で不変）。X-11並行性証明テスト4本をproduction実機構（WithTx）へ追随——旧コメントの「ctx-txKey機構とtreatmentRepositoryの混在は自己デッドロック」はdbOrTx化で解消され、**本テストがWithTxでgreenであること自体がtreatment書込の同一tx参加の実DB証明**。
- **Phase 2（`6508faab0`・機械的縦移動）**: 3batch順次（A=repository→B=service→C=handler+配線）を1コミット集約。**dose kernel（dose_calc/dose_revalidation/dose_validators）もmedicalrecordへ移動**——dose_validators.goが`eligibleMedicineUnitsForPerWeight`（per_weight適格単位の安全マップ）をdose_calc.goと共有しマップ複製は許可集合ドリフト源のため分割不可、medicine帰属=medicalrecord（論点#4）とも整合。残留medicine/medicine_dose_param serviceは`medicalrecord.`修飾消費。cross-package依存はservice_deps.goのconsumer-side view 6本、逸脱監査はAuditEntry+main.goの既存medicalRecordAuditTxAdapter。discountガード（BUG-372）はPermissionChecker（bool判定view）注入+`handler.HasPermission` transitional accessor新設で移植。route 6本（medical-records配下5+`/pets/:id/treatment-history`）はRBAC逐語転記で単一RegisterRoutesへ。**master-FK write gate防御の判断**: `*medicalrecord.MedicineDoseParamInput`引数でNoUnknownCrossPackageParamが発火した際、`"medicalrecord"`のqualifier包括allowlist化を却下（medicalrecordにはCreateTreatmentInput等master FKを運ぶ型があり恒久弱体化）——service側transitional type aliasで解決。facade残置=repository.TreatmentRepository（billing_item read消費・削除=BE9-2F）。carrier追加=fptr/errAuditWriteFailed/mockMedicineDoseParamRepository（残留medicine dose test用・解消=medicine移行時）。
- 検証: 両フェーズともbaseline同名green・変更5package DB-backed全数test -p 1 green・build/vet(./...)/gofmt/docs-drift/diff-check clean・敵対レビュー各2レンズ（Phase1: CRITICAL/HIGH 0/MEDIUM 1受容=LogEntryTx内部ラップ1段追加、Phase2: CRITICAL/HIGH 0/MEDIUM 1是正=discountガードfork化にtreatment_discount_permission_test.go新設）。

**⑤hospitalization/discharge-with-billing（billingとの実エッジを含む最終段）**

- 対象file代表例: `internal/service/{hospitalization,hospitalization_plan,daily_record,care_plan_item}_service.go` + `internal/repository/hospitalization*.go` + `internal/handler/hospitalization*.go`。manifest上hospitalization系10 fileが母集団。
- **baseline test（移動前green必須・billing atomicity中核）**:
  - DischargeWithBilling一連: `TestHospitalizationService_DischargeWithBilling_NotFound` / `_AlreadyDischarged` / `_UpdateFails` / `_WithoutAccounting` / `_CarePlanItemsFetchError` / `_BillingCreateError` / `_WithCarePlanItems` / `_BillingItemCreateError` / `_UpdateBillingTotalsError` / `_ConcurrentDoubleDischarge_ReturnsNotFoundWithoutAccounting`（全10 test、`internal/service/hospitalization_service_test.go`）——billing atomicity（tx境界・二重会計防止）の直接証拠。
  - clinic isolation（owner/pet混入防止）: `TestHospitalizationService_DischargeWithBilling_DoesNotPropagateForeignOwnerPet` / `_RejectsContaminatedOwnerPetAfterOuterFind` / `_WithoutAccounting_RejectsForeignOwnerPet` / `_RejectsInvalidOwnerPetLinks`（`internal/service/hospitalization_owner_pet_clinic_isolation_test.go`）。
  - master-FK-write: `hospitalizationService.Create` / `Update`のCageID guard（allowlist、test: `TestHospitalizationService_Create_RejectsCrossClinicCageFK` / `_Update_RejectsCrossClinicCageFK`）。
- **rollback単位**: sub-batch②③④と同型。**billingとの実エッジ（`DischargeWithBilling`がBilling/BillingItem行を作成）は逆依存を作らない**——medicalrecordがbillingのrepository/serviceをconsumer-side interfaceで受ける既存パターン（ADR-006 §5 cycle解消方式）を維持し、billing側の型をimportしない。billing domain自体がBE9-2Cで未着手のため、hospitalization移動時点でもbilling側は旧`internal/service`のまま——facadeでなくconsumer interfaceでの分離を維持すること。
- **前提条件**: sub-batch①②③④。billing domainのBE9-2C未着手（billing_item_repository.go是正=論点#6が前提、todo.md バグ台帳 BUG-417）でも本sub-batch自体は着手可——medicalrecord→billingは既にinterfaceで逆転済みのため billing側の状態に非依存。ただしbilling側のrepository実装がBUG-417の防御ギャップを抱えたままである点は、hospitalization移動が新たな依存を追加しないことの確認事項として残す（是正はbilling domain着手時）。

**完遂（2026-07-21・Phase1=`d4e227cf8`／Phase2=`f024b09e7`）**: ④bと同型のrefactor-then-move。Phase1=hospitalizationServiceのWithTx+個別注入8本化（dbOrTx化3本=hospitalization Lock/UpdateIfNotDischarged+carePlanItem FindByHospitalizationID・Q2-C並行性テスト2本をWithTx追随・敵対レビューApprove/CRITICAL-HIGH 0・MEDIUM=BUG-418起票）。Phase2=hospitalization/hospitalization_plan/daily_record/care_plan_itemの4エンティティ72fileを3batch順次で縦移動（dailyrecord subpackageはroll-up・facade 4本・route 22本RBAC逐語転記・snapshot両側追随85→107）。**追加のkernel昇格2本**: validateReservationOwnerPetLinks（AUD-001・reservation系と恒久跨ぎ）+DefaultTaxRate。**billing実エッジの決着**: accountingCreator/billingItemWriter view（*model.Billing/BillingItem=model帰属のためADR-006逆依存禁止に非抵触）。**残置**: Services.Hospitalization field（treatment_plan_handler④外の入院所有権検証が唯一のconsumer・medicalrecord型でmain.go注入・削除=treatment-plan移行時）。敵対レビュー統合レンズApprove・CRITICAL/HIGH 0・MEDIUM（allowlist reason stale）は同コミット内是正。LOW参考=旧handler残置treatment-planとの/hospitalizationsグループgin path merge共存に統合起動自動テストなし（staging実起動で確認推奨）。

**⑥〜⑧: 残留パスの再定義（2026-07-21・⑤完遂後の全数棚卸し）**

manifest target:medicalrecordの旧package残存83 fileを実測分類した結果、実体は以下の3グループ+facade群（treatment/vital/checkup系等の完了batch由来alias=BE9-2F対象）。①〜⑤と同じbatch規律（3batch順次・RBAC逐語転記・snapshot両側追随・同名test green・敵対レビュー）を適用する。

**⑥ 残マスタ群（clean move・低risk）= consultation / procedure / medicine(+dose_param) / cage / treatment_plan（計28 file+subpackage 3個roll-up: cage/ consultation/ procedure/）**
- medicine/medicine_dose_paramは既にWithTx+auditTxLogger注入済み（R1-2）、他はrepo単発注入 — Phase 1不要で①同型のclean move。dose kernel（ValidateMedicineDoseConfig等）は④bでmedicalrecord移動済みのため、medicine移動でservice側のtransitional alias（MedicineDoseParamInput）とcarrier（fptr/errAuditWriteFailed/mockMedicineDoseParamRepository）が解消できる。
- treatment_plan移動でServices.Hospitalization残置field（⑤の唯一consumer=treatment_plan_handlerの入院所有権検証4箇所）とhandler側carrier（mockHospitalizationService）も解消。
- baseline: 各service test（consultation/procedure/medicine/cage/treatment_plan_service_test.go）+ medicine系cross-tenant/dose config test群 + master_fk allowlistのmedicineService系エントリ。
- 論点#4済み: medicine帰属=medicalrecord確定。

**⑥ 完遂（2026-07-21・`a21977e91`・90file）**: 3batch順次を1コミット集約。repository flat 3+subpackage 3 roll-up（facade 6本・dbortx 12キー再prefix）→service 6本（medicineInventoryRepo view新設・enum validator 4本[TaxType=billing恒久跨ぎ]をsharedkernel昇格・監査はAuditEntry+adapter統一）→handler 18本+route 35本逐語転記（snapshot 107→142）。**解消した暫定債務**: MedicineDoseParamInput alias・service carrier 3本（fptr/errAuditWriteFailed/mockMedicineDoseParamRepository）・Services.Hospitalization残置field・handler側mockHospitalizationService carrier。discountガード（BUG-372）はpackage-level関数化しtreatment/treatment_plan共用（旧EffectivePermission経由判定はmain.goのh.HasPermission注入で同一チェーン・テストはdeny/grant写像で等価変換）。敵対レビューApprove・CRITICAL/HIGH/MEDIUM 0（LOW=allowlist reason旧パス表記は同コミット内是正）。新carrier=strPtr/uint64Ptr/mockMedicine/Procedure/Consultation（残留cross_tenant builder用・解消=各domain移行時）。medicine_inventory_tx_atomicity_testはinventory未移行依存のためrepository残置（注記付き・移設=inventory移行時）。

**⑦ medical_record本体+addendum+examination（最高risk・finalize/audit中核・medicalrecord unit完了条件）= 完遂（2026-07-21・`d4d7ef068`・84file）**: 3batch順次を1コミット集約。
- Batch A: repository 4本移動（medical_record[owner_visit=メソッド拡張のためfacadeはDTO alias 2本のみ]/addendum/examination）+**kana_normalize.goをrepohelpersへhoist**（owner/pet/ltv/medical_recordの4者共有・BE8§6文書化済みltv解錠・repository側delegate・byte等価）。dbortx 13キー+audit_tx examinationエントリ再prefix。X-11 finalize_lock並行性・examination tx atomicityテストはwithTx/testTransactor化（④b先例=gate追随自体が新機構の実DB証明）。count_clinic_scope_isolationはEstimate/Reservation跨ぎのためrepository残置（注記付き）。
- Batch B: service 9本移動。新view 5本（mrLineCustomerRepo/mrReservationRepo[sharedkernel.OwnerPetLinkVerifier埋込]/mrDeliveryTrigger/mrTagSyncer/mrAuditLogger）。監査はLogMedicalRecordChange/LogAddendumCreateの**具象直渡し**（vital先例・signature=primitives+model型のみ）。examination/addendumのmedRec依存はmedicalRecordLocker/medicalRecordFinderへnarrow化。**計画通り自己解消2件**=computeExamResultStatus複製統合・audit_diff.goのestimate系diffをservice側へ分離（billing残留）。**発見=dead依存**: 旧constructorのownerRepo/petRepoは代入のみ全file使用ゼロ→除去（53 test call site調整・レビューでgrep実証+位置ズレ個別確認）。Services.MedicalRecord=medicalrecord型で残置（reservation_handler consumer・⑤先例）。
- Batch C: handler 9本+route 15本逐語転記（MR CRUD 6+addenda 2+examinations 7。billing-confirmationは旧側group残置・gin path merge共存）。snapshot golden −15行/medicalrecord側142→157。403権限ゲートtest 1本はlab/vital先例で非移行。
- 検証: 4パッケージ全数test -p 1 green（medicalrecord 1140 subtests）・build/vet/gofmt/docs-drift/diff-check clean・敵対レビュー（clinic-isolation統合レンズ・byte比較+機械リネーム正規化diff）Approve CRITICAL/HIGH 0・LOW 2件は同コミット是正。**BE9-2D medicalrecord domain単位完了条件（高riskパス全提供+runtime test green）を充足。旧layer上のmedicalrecord系残実装ゼロ（残=facade群→BE9-2F）**。

**⑧ checkup_sync（8 file）= 論点#7通りlstep帰属変更が既定** — lstep domainのBE9-2C着手時にmanifest訂正の上lstepへ移動（medicalrecordへは移さない）。

#### reservation domain sub-batch定義（2026-07-21 inventory実測: production 77 file = handler 30+service 31+repository 16）

Phase 0（論点#1案A書込一本化）完遂済み（`3dc35694e`）。縦移動はmedicalrecord playbook（3batch A/B/C・RBAC逐語・snapshot両側追随・敵対レビュー1統合レンズ）を踏襲し、以下R①〜R⑥の順で実施:

- **R① reservation_typeマスタ群 = 完遂（2026-07-21・`c4c95698d`・86file）**: `internal/reservation`新設。repository 6本（subpackage 4 roll-up改名+flat 2）+service 9本+handler 9本+route 28本逐語転記+test 26本。昇格3件（pg_errors→repohelpers・ValidateTimeRange→sharedkernel[trimming跨ぎ]・ParseDate→httpapi[owner/pet/lstep/inventory跨ぎ]）・dead依存除去（liff resAdminRepo）・master_fk prefixへ"reservation/"追加。**転記漏れ1本（LIFF image upload）をsnapshot gateが検出→復元**。レビューApprove CRITICAL/HIGH 0（byte等価機械diff・RBAC 28本全数一致・Test名143→152突合消失ゼロ）・MEDIUM 1件（service側alias消費者ゼロ化）は同コミット削除。残置=handler側alias（reservation_response.go[R③]消費）・mock carrier（R④⑤解消）・filterApplicableUnavailableTimes複製（R⑤自己解消）。
- **R② reservation_staff + reservation_schedule = 完遂（2026-07-21・`227792859`）**: repo 2+service 3+handler 6+route 10逐語+test 8移動。Phase 0 delegateをconsumer-side view化（staffsWriter/shiftWriter・facade ctorが具象注入=論点#1最終形）。昇格2件（junction_helpers 5本→repohelpers[auth跨ぎ]・シフト時刻検証3本→sharedkernel[staff跨ぎ]）。repo test 7本は意図的repository残置（facade=production配線の実DB検証・count_clinic先例）。snapshot 28→38。レビューApprove CRITICAL/HIGH 0（書き込み者複線化なしgrep実証・tx_atomicity 6本green）。
- **R③ reservation本体 + capacity + timeslot_engine + available_dates = 完遂（2026-07-21・`00afe3898`）**: コア予約フロー（X-9 booking lock/AUD-001/FKガード3本）。medicalRecordAutoCreator view（svcs.MedicalRecord具象直渡し=topo逆行回避）+liffAvailability/staffAssignmentFinder view。前倒し移動2件（liff business hours+engine 2関数）。Paginate hoist。master_fk帰属変更（liffService→reservationValidators・監視継続実証）。**openapi drift walkerはindirect route登録を走査不能（phantom 7本検出→inline展開で解消）**。snapshot 38→45。レビューApprove CRITICAL/HIGH 0（Test名74本消失ゼロ）。Services.MedicalRecordはhandler側残留consumer（medical_record_ownership.go等）があるため残置継続。
- **R④ appointment_admin + notification = 完遂（2026-07-21・`94bdcb94b`）**: 墓標3件削除。notificationのlstep/auth跨ぎ依存はclosure注入化（decryptCredential/LinePusher/sendMail=service集約が旧実装を包んで渡す・topo逆行import回避）。レビューApprove・HIGH 1件（SMTP closure写像未検証）はsmtpSendAdapter抽出+field-mappingテストで是正。R③分離のadmin系test 2節合流。snapshot 45→48。
- **R⑤ liff系 = 完遂（2026-07-21・`de5f0d348`・最大スライス~7,400L）**: service 10+handler 4+test 15移動。LIFF公開API 13本をRegisterLiffRoutesへ逐語転記（middleware循環→injected closure化=liffAuth/rate limit factory/linkLiffAccount）。**openapi drift walkerへrootFn拡張**（LIFF絶対パス用の追加root機構・gate強化）。ReservationLimitError 409フォールバックを計画移設・R①③複製解消・transitional alias全削除・scoped golangci-lint 0件。レビューApprove CRITICAL 0（HIGH/MEDIUM/LOW全是正）。key契約テスト新設（liff_customer_id）。snapshot 48→61。
- **R⑥ line_reservation_setting = 完遂（2026-07-21・`0ee22c180`）**: reservation最終スライス。encrypt/decryptLineCredentialはclosure注入化（R④先例）+実配線の貫通統合テスト新設（レビューMEDIUM是正）。route 2本逐語・snapshot 61→63。line-customers 2本はlstep帰属で残置注記。**BE9-2C reservation domain単位完了条件を充足**（旧layer上のtarget:reservation残実装ゼロ・残=facade群→BE9-2F）。

各batchでdbortx/audit_tx/preload/master_fk allowlist再prefix+route snapshot±追随+全数test -p 1 green+敵対レビューApproveを完了条件とする。sub-batch内の依存詳細（narrow view設計・mock carrier方向）は各batch着手時にinventory実測で確定。

**論点#4の紐付け（2026-07-20裁定済み）****論点#4の紐付け（2026-07-20裁定済み）**: `model/medicine.go`/`vaccine.go`は**medicalrecord帰属で確定**（ADR-006委任裁定）——sub-batch②のvaccine/vaccination移動は既定通りmedicalrecordへ（inventoryへの付け替えなし）。`internal/model/line_reservation_setting.go`は**reservation帰属で確定**（概念タグのみ・ファイル移動なし、reservation着手時にタグ通り扱う）。

#### billing domain sub-batch定義（2026-07-21 inventory実測: production 65 file = handler 27+repository 21+service 17）

前提=BUG-417是正完了（`2634f58fe`・EXISTS subquery化+分離テスト4本mutation実証）。`internal/billing`新設。playbook=reservation R①〜R⑥と同型（3batch A/B/C・敵対レビュー1統合レンズ・snapshot/lint/allowlist追随・昇格判定はシンボル単位）:

- **B① マスタ群 = 完遂（2026-07-21・`22b2094e1`）**: `internal/billing`新設。subpackage 3 roll-up+service 4+handler 9+route 18逐語（resource 3種）+snapshot新設。lint canary 2箇所再ポイント（nested討査到達性維持を再parse実証）。master_fk prefixへ"billing/"追加。レビューApprove CRITICAL/HIGH 0・MEDIUM 3是正（dead delegate削除等）。scoped lint 0件。
- **B② estimate + billing_confirmation = 完遂（2026-07-21・`9a1e8bad7`）**: X-11/AUD-005/SD-2ガード群。kernel直呼び化でservice側lock delegate解消（レビューMEDIUM=dead化を同コミット削除）。**discountガード（BUG-372）をhttpapiへ昇格**（3domain目恒久跨ぎ・medicalrecordはdelegate化）。billingAuditAdapter（AuditEntry写像・gate回避はtype alias規則）。route 8本逐語・snapshot 18→26。レビューApprove CRITICAL/HIGH 0。
- **B③ billing_item + refund + billing計算 = 完遂（2026-07-21・`d2e01da75`）**: 金銭中核。#211 fail-closed（tx原子性テスト同一txKey機構でgreen）・BUG-417是正込み移動。treatmentBillingReader view化。adapter 9field写像テスト新設（R④教訓の恒久化）。route 8本逐語・snapshot 26→34。レビューApprove CRITICAL/HIGH 0・MEDIUM 3是正。
- **B④ accounting core+reports = 完遂（2026-07-21・`7fc7649f5`）**: accounting_service全5file+repo 8file（reports系は同一struct上methodのためB⑤から前倒し合流）。fail-closed監査3経路byte等価。機械置換事故2件（delegate自己再帰stack overflow/optionalStaffIDの401副作用）を検出是正・レビューで原本byte等価実証。DTO alias 18型（cash_register=B⑤残留用）。route 10本逐語・snapshot 34→44。レビューApprove CRITICAL/HIGH 0。
- **B⑤+B⑥ = 完遂（2026-07-21・`24420376c`）**: cash_register(+close)/accounting_report の repo/service/handler移動+**残置解消**。DaySchedule を sharedkernel 昇格。同一package化で顕在化した型名衝突を CloseBillingDetailRow 改名で解消（API出力DTOは無変更）。route 6本逐語・snapshot 44→50。alias棚卸し（service 7型+handler file削除・repository DTO aliasはlstep系test実消費のため残置しコメント訂正）。レビューApprove CRITICAL/HIGH/MEDIUM 0。**BE9-2C billing domain単位完了条件を充足**（manifest target:billingの旧layer production実装ゼロ・残=facade群→BE9-2F）。
- **B⑥ 残置解消 = B⑤に合流して完遂**（上記）。

各batch完了条件=従来gate+scoped golangci-lint（billing全域）0件（R⑤知見④の恒久化）。

#### lstep domain sub-batch定義（2026-07-21 inventory実測: production 107 file = handler 40+service 52+repository 15）

前提=論点#2裁定済み（単一`internal/lstep`・内部3分割しない）+論点#7裁定（checkup_syncはlstep帰属＝⑧を本domainで合流）。playbook=medicalrecord/reservation/billing と同型（3batch A/B/C・敵対レビュー1統合レンズ・snapshot/lint/allowlist追随・昇格判定はシンボル単位）:

- **L① 設定・認証基盤**: lstep_settings(+connection/credentials/thresholds/update)/lstep_sync_settings/line_credentials/lstep_sync_error_counter + repo 4本。cipher系closure注入はB④/B⑥先例（decrypt/encryptLineCredential）を踏襲。
- **L② LINE 送信基盤**: line_messaging/line_send/line_send_log/line_link(+token)/line_customer + repo 4本。**reservation側が消費するLinePusher/decryptCredential closureの供給元**（R④注入の解消先）。
- **L③ タグ同期コア**: lstep_tag_sync 群（tag_cache/tag_code_mapping/tag_config/trigger_priority repo + health_tag_sync 系 6file + CPM kernel）。**⑧checkup_sync（service 4+handler 3+repository/checkupsync）を本batchで合流**（論点#7）。
- **L④ 配信トリガ・監視**: lstep_delivery_trigger 群（batch/client/methods/state/suppression/log）+ lstep_delivery_monitor + lstep_lifecycle。
- **L⑤ バッチ・分析・CSV**: lstep_batch(+delivery/dormant/noshow/segmentation)/lstep_analytics/lstep_csv_import(+helpers)/lstep_friend_attribute_snapshot/aggregation。
- **L⑥ 残置解消**: repository DTO alias（accounting_reports_dto_aliases.go=lstep系test consumer）・mock carrier群・Services field棚卸し。

各batch完了条件=従来gate+scoped golangci-lint（lstep全域）0件。

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

### 現在地と着手前ゲート（2026-07-20）

**BE9-0 → BE9-2A → BE9-1 → BE9-2B（`d00c72a93`）→ ①（`538cdb34`）→ ②（`14f00f6c`）→ ③（`75c55c48`）→ ④a（`e3eb253e`）→ ④b（`cd8fd984d`+`6508faab0`）→ 共有カーネル昇格（`f93299f1c`）→ ⑤（`d4e227cf8`+`f024b09e7`）→ ⑥（`a21977e91`）→ ⑦（`d4d7ef068`・2026-07-21）まで完遂 = **medicalrecord domain完了（⑧checkup_syncはlstep帰属）**。reservation Phase 0（論点#1案A・書込一本化）完遂。次 = reservation縦移動本体**。

**sub-batch④の重要決定（2026-07-20 inventory+advisor）**:
- **④を④a/④bに分割する**。**④a（clean move）= vital + clinical_plan + medical_record_image → 完遂（2026-07-20・コミット`e3eb253e`）**。①②③と同じnarrow consumer-side interfaceパターンで移動。**④a実績の非自明点**: (1) advisor指摘3点を織り込み——lockDraftMedicalRecordのmedicalrecordコピー（本体byte-identical確認済・コメント差のみ）にtest新設（`medical_record_lock_test.go`・nil-parent-fails-closed parity・従来カバレッジ0）、verifyMedicalRecordOwnershipを残留`internal/handler/medical_record_ownership.go`へ抽出（④外treatment_plan_handlerが消費）、vital LogVitalChangeは**adapter不要の具象直渡し**（signature=primitive+map[string]anyのみでservice.AuditServiceと完全一致——labAuditAdapter型変換不要でfield parity達成）。(2) Batch Cのfork委譲がエージェント調整の錯綜（相互peer認識・fork完了待ちループ）を招き、統括が引き取り単一finisherで解消。残留漏れ3件（medical_record_handler_testのServices.ClinicalPlan field参照×2・simple_settingsのmisfiled TestUpdateClinicalPlanRequest）を統括が直接是正。(3) 親medical_record handlerはh.svc.ClinicalPlanをproductionで元々不使用（HEAD確認）のためfield注入除去はbehavior-preserving。検証: 変更5 package DB-backed全数test -p 1 green・build/vet(./...)/gofmt/docs-drift/diff-check green・敵対レビュー4レンズ+反証で指摘0（vital監査byte-identical専用レンズ含む）。**④b（refactor-then-move）= treatment（+treatment_dose_save+dose kernel）→ 完遂（2026-07-21・Phase1=`cd8fd984d`/Phase2=`6508faab0`・ユーザーの「BE-refactor.md対応」指示をgo-aheadとして実行）**。計画通り2フェーズで実施——詳細は「BE9-2D: medicalrecord sub-batch定義」④の完遂記録参照。計画時に見えていなかった発見: (1) dose kernelのservice側残留consumer（dose_validators.go←medicine master書込検証）が安全マップ共有でkernelごと移動が必要だった。(2) X-11並行性テストが旧repo-swap機構を「productionと同じ」とpinしており、gate追随（WithTx化）自体が新機構のtx参加の実DB証明になった。(3) master-FK gateのcross-package qualifier検出はtype aliasで型同一のまま回避可能（qualifier包括allowlist化は恒久弱体化のため禁じ手）。
- **共有カーネル昇格batch = 完遂（2026-07-21・`f93299f1c`）**: `internal/sharedkernel`新設（import面={apperrors,model,stdlib}・acyclic）。LockDraftMedicalRecord+MedicalRecordLocker/GoSafe/AuditActorTypeFor/validators family（RequiredName/OptionalName/OwnedMasterFK(+FKs)/SetNullableUint64Field/NonNegativePrice/DiscountRate）+共有ErrMsg定数5本を単一実装化。service/medicalrecord両側は既存呼び出し面互換の1行delegate+定数alias（呼び出し40+箇所無変更・delegate解消=各domain移行時）。scope除外は計画通り（logReplaceDeletionTx/computeExamResultStatus=examination移行時自己解消）。検証=高リスク3本体のHEAD原本byte-identical機械証明+3パッケージ全数green+lock testの三重化（sharedkernel直+両delegate経由同名）。
- **（履歴）共有カーネル昇格は④に含めない・④後の専用batchへ**。決定打: 昇格は④内でfile 0・新規重複0を解決する（medicalrecord側コピーはsub-batch②から既存）。逆に昇格はbilling_confirmation/estimate/examination/accounting/lstep/medicine等④外callerに波及し最高リスクスライスを膨張させる。**実測でコピー数は2（service+medicalrecord）でrule-of-three(3コピー)ではない** — 昇格の真の根拠は「literal count」ではなく「callerがmedicalrecord子孫 vs billing/accounting/lstep/masterという恒久ドメイン境界を跨ぐ（domain移行で決して消えない）」こと。専用batchはpure/leaf依存helperのみscope: lockDraftMedicalRecord(+medicalRecordLocker)/goSafe/auditActorTypeFor/validateNonNegativePrice/validators family。**lockDraftMedicalRecordを最初に昇格**（X-11 finalize-race臨床安全guard・純関数・{apperrors,model}依存のみ）。logReplaceDeletionTx・computeExamResultStatusは除外（examination移行時に自己解消・audit-sink型結合あり）。kernel package import面={apperrors,model}=検証済みacyclic。

**sub-batch④以降の前提**: ⑤hospitalization/discharge-with-billingは④の後。billing側の`lockDraftMedicalRecord`呼び出し元（billing_confirmation/estimate）は触れない（論点#6=billing着手時）。ゲートの正本 = ADR-006「未解決論点」節。以下は要約:

**2026-07-20: 論点#1〜#4はユーザーのPO判断委任に基づき裁定済み（Resolved）** — 裁定内容・根拠・実装条件の正本 = ADR-006「未解決論点」節（同日追記）。要約:

| # | 裁定（2026-07-20） | 実装時の条件 |
|---|---|---|
| 論点#1 | **案(A) staffを唯一の書き込み者に一本化**（`staffs`・`shift_entries`両テーブル。reservationはconsumer-side interfaceでstaffのexportedメソッドを呼ぶ）。案(B)別テーブル切り出しはDDL/データ移行/DB_RESET波及とライフサイクル二重管理のため却下 | 一本化後のstaff側write pathにclinic-scope検証を維持+クロステナントtest追加。read pathは対象外。**reservation/staffのBE9-2C着手ブロックは解除済み** |
| 論点#2 | **単一`internal/lstep`で確定**（LIFF 9 fileのreservation再分類済みで分割動機が縮小） | BE9-2D lstep実装中に実消費境界が出現した場合のみ再評価（ADR再改訂） |
| 論点#3 | **clinic→reservationエッジは追加しない**（2026-07-20再grepで消費者=clinic CRUD+accounting_reportのみと確定） | 休診日の予約反映は製品論点としてBE9外（要件化は責任者名付き別Issue） |
| 論点#4 | **medicine/vaccine=medicalrecord、line_reservation_setting=reservation帰属で確定**（概念タグのみ・ファイル移動なし） | sub-batch②のvaccine移動は既定通りmedicalrecordへ |
| 論点#6 | **裁定対象外（決裁事項ではなく技術的是正ゲート）** — `billing_item_repository.go`のsubquery是正+クロステナントtest追加（todo.md バグ台帳 BUG-417） | billing domain着手時の必須前提。BE9外でこのファイルへ触れる場合もその場で是正 |

（論点#5=間接isolation 3件はBE9-2A内で検証完了・決裁事項から除外済み。#201等の臨床安全判断は委任対象外で未裁定のまま）

**論点#7（Open・2026-07-20 sub-batch②で新規登録）**: checkup_sync系（`internal/service/checkup_sync_service{,_create,_metadata,_preview}.go`+handler 3 file+`internal/repository/checkupsync/`）のdomain帰属 — manifestはtarget:medicalrecordだが、依存の実質はlstep domain（`lstep.Client`直import・CPM純関数kernel `CalculateCPMStage`への依存でこのkernelは複製禁止コメント付き・owner/pet/tagCache repo・LstepSettings・route=`/clinics/:clinic_id/lstep/checkup-sync`・権限=ResourceOwners）。sub-batch②では移動せず`internal/service`残留。**発火タイミング=lstep domainのBE9-2C着手時**: (a) lstepへ帰属変更（推奨・manifest訂正）か (b) CPM kernelを独立packageへ先行抽出してmedicalrecordへ移すかを決める。

### BE9-2B/2C実績からの申し送り（以降の各batchで踏む・sub-batch②分は末尾に追記）

- **handler側固定gateの追随を各batchに含める**: domain移行はroute snapshot（`handler/testdata/route_snapshot.golden`）とOpenAPI route drift（`apicontract/openapi_route_drift_test.go`の`migratedDomainRoutePackages`への新domain登録）の更新を必ず伴う（BE9-2B/2Cで実地確認）。batch計画時にこれら固定gateを事前列挙する。
- **domain packageのroute登録は単一エントリポイント必須**: `openapi_route_drift_test.go`の`buildFuncsFromDir`はbare名でfunc mapを構築するため、同名メソッドが複数struct上にあるとroute setがdrift検知から**静かに脱落**する。per-entity複数`RegisterRoutes`は禁止、`<domain>.Handler.RegisterRoutes` 1本に集約する（BE9-2C sub-batch①で発見・設計修正済み）。
- **master-FK-write lintの`knownSafeParamQualifiers`への`"gin"`追加は対応済み**（sub-batch①）。以後のdomain移行でhandler層がスキャン対象に入っても追加作業は不要。`serviceWriteRolePackagePrefixes`への新domain prefix追加は初回のみ必要。
- **fan-in 0の旧実装はfacade化せず完全削除が既定**（manualarticle/sub-batch①の先例）。型aliasで延命するのはfan-in>0（他domainの旧実装が依存）の場合のみで、削除期限を持たせる。
- **共有validation kernelの複製debt**: `internal/medicalrecord/validators.go`は`internal/service/validators*.go`のpure部分の意図的複製（medicalrecordから旧serviceへのimportを禁止するため）。共有カーネルをcross-cutting packageへ昇格した時点で両コピーを統合する。以後のdomainで同じ複製が3個目に達したら昇格を必須化する（rule-of-three）。
- **cross-domain依存の解消パターン**: Audit依存 = consumer-side interface（`AuditLogger`）+ `main.go`のadapter、Permission依存 = middleware関数型注入（auth domain未移行のため）。BE9-2C以降も同型を使う。
- **旧layer側は薄いdelegating facadeで互換維持**（呼び出し側無変更）。facade削除はBE9-2Fまで持ち越し、削除期限を持たせる。
- **docs数値ゲートの追随**: `scripts/check-docs-symbol-drift.sh`の「ハンドラー数」チェック（`internal/handler/*_handler.go`のfile数）はBE9のhandler分散で測定基盤が溶解したため、docs側の宣言（`docs/spec/specification.md`）を削除して恒久解消済み（2026-07-20。宣言が復活しない限り3cチェックは発火しない）。**各batch完了時に同スクリプトを実行して他の数値宣言のドリフトも確認する**（sub-batch①ではこの1件が漏れていた）。
- **共有テストDBフレーク**（旧BE8 §5 row 6）はBE9でも生きている: `go test -p 1 ./internal/repository/...`で本batchが触れないファイルの赤は退行でない（pre-batch再現を確認して続行）。恒久対処=該当テストの`setupIsolatedTestDB`化はfollow-upのまま未着手。
- **（sub-batch②）移動test fileが定義するmockを残留testが共有している場合はmock carrier fileを残す**: `internal/service/be9_2d_mock_carriers_test.go`先例（6 mock・残置理由コメント付き）。carrier解消は当該残留consumerのdomain移行時。
- **（sub-batch②）中間状態でgateを弱めたら復元をbatch完了条件に含める**: Batch中間状態のための一時的なlint緩和（例: `knownSafeParamQualifiers`への追加）はコード内に"REMOVE"マーカーを書き、最終batchの完了条件＝除去+green再実測とする。
- **（sub-batch②）response file移動時はopenapi_date_format_drift_test.goのscan dir拡張を確認**: drift entryはbasename keyingのため、移動でscan対象から外れると「drift解消」の誤シグナルになる。`responseScanDirs`へ移行先dirを追加する（route driftの`migratedDomainRoutePackages`と同型の追随gate）。
- **（sub-batch②）大規模移行は3batch順次（repository→service→handler+配線）が安定**: 各batch間もコンパイル可能に保つには、中間状態でService集約structのfield型をmedicalrecord型へ付け替える（service→medicalrecord importは方向的に合法）。旧handlerは method set 同一のため無変更で通る。
- **（④b）tx機構の変換は移動と別コミットに分離する（refactor-then-move）**: repo-swap機構（repos.Transaction）依存のserviceは、先にWithTx+dbOrTx化をbehavior-preservingで完遂・レビュー・コミットしてから移す。旧機構でtx-bound cloneにより「暗黙に」tx参加していたreadは、WithTx化で明示的にdbOrTx化しないとtx外へ漏れる（vital FindByMedicalRecordIDで実際に必要になった）——変換時は閉包内の全read/writeについてdbOrTx対応を実測すること。
- **（④b）並行性証明テストはproduction実機構をpinしている**: tx機構を変えたら、その機構を使う実DB並行性テスト（X-11系）も同じbatchで追随させる。追随後のgreenが新機構の挙動証明を兼ねる（旧機構のまま残すと「productionが使っていない機構の証明」に劣化する）。
- **（④b）master-FK write gateのcross-package qualifierはtype aliasで回避する**: 移動済みdomainの型をservice層シグネチャに晒すとNoUnknownCrossPackageParamが発火する。`knownSafeParamQualifiers`へのqualifier追加は当該package全型の包括safe化＝恒久弱体化なので禁じ手。service側に`type X = medicalrecord.X`のtransitional aliasを置き無修飾へ戻す（型同一・gate実効維持・当該domain移行時に解消）。
- **（④b）安全許可集合マップ（eligibleMedicineUnitsForPerWeight等）は複製禁止**: 共有シンボルが「安全側の許可集合」を定義する場合、複製はドリフト＝安全性劣化源。消費者が複数domainに跨るなら定義側kernelごと帰属domainへ移し、残留消費者は修飾importで参照する。
- **（sub-batch②）rule-of-three進行状況**: 共有kernel複製は validators系（sub-batch①）+ lockDraftMedicalRecord/goSafe/logReplaceDeletionTx/validateNonNegativePrice（sub-batch②・いずれも原本service+複製medicalrecordの2箇所目）。**3個目の複製が発生するdomain（reservation/billing着手時が有力）で共有kernel package昇格を必須化**。

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
