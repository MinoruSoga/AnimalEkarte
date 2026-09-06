# ADR-006: backend domain package 境界と許可依存グラフ

**Status**: Accepted / Implemented（2026-07-19。2026-07-24 amended: BE9 final cutoverとCSV cutover 21表契約へ同期）
**Date**: 2026-07-19（Final implementation amendment: 2026-07-24）
**Relates to**: ADR-005（Go/Gin公式ベースライン採用）、ADR-002（clinic_id完全隔離）、[`docs/product-philosophy.md`](../../product-philosophy.md)
**Deciders**: MinoruSoga（Accepted への昇格判断者）

## Context

ADR-005は「Handler → Service → Repository、Clean Architecture、layer-first/domain-firstをmandatory architectureにしない」ことと、「packageは凝集性・利用者・依存方向・変更単位で設計する」ことを決定した。しかし具体的にどのdomain package境界を採用するかは未確定のまま残されていた。

BE-refactor.md の BE9-2A（本ADRの起票元タスク）は、当時の全761 production Go source（`backend/internal/**/*.go` + `backend/cmd/**/*.go`、`_test.go`・`cmd/_archive` 除外）を分類manifestの761 rowへ固定し、codegraph（callers/callees/explore）+ grep/git log による再実測（8並列エージェント調査）を通じて、実際のRBAC resource名・route registration構造・GORM型のClinicIDフィールド・Goのreceiver型制約という一次証拠に基づき、target package境界の候補を導出した。移行後target packageの物理file数とは別指標である。

旧 BE-refactor.md §9（BE8-2、`internal/service`のみのgo/ast識別子参照実測、69ドメイン）と旧見積表（filename-prefixのみのカウント）は**正本として継承していない**——本ADRの判断は再実測（[boundary map doc](../be9-2a-boundary-map.md)）に基づく。詳細な再実測データ、per-domainの9列boundary map、§9との差分は同docを参照。**本ADRは決定のみを記録し、詳細data（分類manifest 761 source row・per-domain data・fan-in/out数値等）はboundary map docへのリンクに留める（二重管理禁止）**。

## Implementation outcome（2026-07-24）

本Decisionは実装済み。当初13 target packageは全て現行domain/capability packageへ収束し、2026-07-30に`identitylink`を加えて14 target packageとなった。旧`internal/handler`、`internal/service`、`internal/repository` directoryは**完全削除済み**（2026-07-24 recensus時点では service test-only 14 / repository test-only 50 が残っていたが、その後撤去し test residual も含め directory 自体が存在しない）。3旧layerのproduction implementationとproduction Go import edgeはいずれも0件で、期限付きfacade、巨大`Handler` / `Services` / `Repositories` aggregator、旧transaction facadeは撤去済みである。live mechanical lint gateは`backend/internal/lintscan/`に置く。

`cmd/api`は22 production Go fileへ分割した明示composition rootで、18 fileがtarget domain packageを直接importする。共有能力は実consumerに基づき`audit`、`persistence`、`scheduler`、`sharedkernel`、`textsearch`、`testdb`等へ命名して抽出し、`common`/`util`の無差別bucketは作成していない。移行後の物理file数、manifest 761 rowのprovenance、旧path消滅状況は[boundary map](../be9-2a-boundary-map.md)を正本とする。

2026-07-24のfollow-up hardeningでは、LINE webhookの全setting-secret readを受信前identity解決だけの限定例外とし、一意に署名一致したclinicへowner lookup/updateをscopeした。duplicate secretによる曖昧系はfail closed、owner未登録のtyped NotFoundだけをno-op、真のlookup/update errorはnon-2xx retryへ伝播する。follow/unfollow更新は`clinic_id + owner id + expected line_user_id`とLINE event timestampを使うCASとし、stale・duplicate・out-of-order・再連携前IDは`RowsAffected == 0`の安全なno-op、同時刻はunfollow優先とする。公開LIFF account linkはowner PIIを返さない`204 No Content`とし、LINE ID token検証はredirectを追従しない。billing confirmation/returnは認証済みstaffをactorとし、`Content-Type: application/json`（charset parameter可）以外を415、bodyを8 KiBのexact-key/string strict single-object JSON、trim後non-blankの`return_reason` 500文字、`memo` 1,000文字として境界で強制する。scheduler opsはCloudflare Access JWKSをWorker isolate内で10分cacheし、同時取得を集約、unknown `kid`/upstream failure後のrefreshを60秒cooldownしてfail closedにする。

本ADRのimplemented判定はcode/package境界についての判定であり、release readyを意味しない。fresh DB migration実適用・checksum/rollback確認、remote CI/full coverage artifact、production deploy/configuration、scheduler/observability/alert/recovery rehearsalは Linear hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) / [`todo.md`](../../../todo.md) の release gate として未実施である（旧 OPS-13〜17 節は死リンク）。

## Decision

### (a) domain-firstはproject decisionであり、Go/Gin公式の必須構成ではない

Go公式（[Organizing a Go module](https://go.dev/doc/modules/layout)、[Package names](https://go.dev/blog/package-names)）とGin公式（[API design patterns](https://gin-gonic.com/en/docs/routing/api-design/)）は、domain-firstかlayer-firstか、あるいはどちらでもない構成かを規定しない。本ADRが採用するdomain-first境界（下記14 target package）は、AnimalEkarteプロジェクト固有の設計判断であり、[go-gin-backend-guidelines.md](../../../.claude/rules/go-gin-backend-guidelines.md) §2「公式未規定」の範囲内で、ADR-005が示す4基準（凝集性・利用者・依存方向・変更単位）を根拠に確定した。

**採用するtarget package構成（14 target package + 既存cross-cutting packageの現状維持）**:

```text
backend/internal/
  owner/ pet/ staff/ auth/ reservation/ trimming/
  medicalrecord/ billing/ inventory/ lstep/
  clinic/ manualarticle/ httpapi/ identitylink/
  # 現状維持（既存の凝集cross-cutting package）:
  config/ dbconn/ middleware/ infra/ model/
  timeutil/ seedbundle/ seedlogin/ logger/ csvimport/
  labdeviceagent/  # cmd/lab-device-agent 専用のローカル検査機器 agent。domain ではない
  authjwt/ apperrors/ apicontract/ lintscan/
  audit/ persistence/ scheduler/ sharedkernel/
  textsearch/ testdb/
```

> 2026-07-30 amendment: `identitylink/` を #239 Phase 1 の vertical slice として target domain に追加（cross-clinic owner/pet identity 連結）。依存は `apperrors` / `audit` / `httpapi` / `model` / `persistence` / `textsearch` のみ（owner/pet package への Go import は無し）。

> 2026-08-22 amendment: `labdeviceagent/` を keep-tier に追加する。Mac ローカル検査機器の serial-port agent（ADR-008）であり、consumer は `cmd/lab-device-agent` のみ。14 target domain には含めない。

> 2026-09-05 amendment: `seedlogin/` を keep-tier に追加する。migrate フェーズ3の合成デモログイン upsert。runtime の `staffs` write owner は `staff` のまま。14 target domain には含めない。

### Product philosophyに基づく運用境界（project decision）

[`docs/product-philosophy.md`](../../product-philosophy.md)はproductのWHAT / WHYと判断順序を定める文書であり、このfolder tree自体を要求しない。本ADRは、業務workflowの一貫性と変更cycle timeをpackage境界へ反映するため、上記構成を**domain/capability-firstのmodular monolith**として運用する。

- route、use case、transaction、persistence、testをdomainごとのvertical sliceとして計画・review・rollbackする。
- domain内に`handler`、`service`、`repository` subpackageを機械的に作らない。実際のconsumer、依存方向、変更周期が分かれた場合だけ分離する。
- 1つのbusiness factには1つのsource of truthとwrite ownerを置く。`appointments`とそのlifecycleのwrite ownerは`reservation`、`staffs`と`shift_entries`のwrite ownerは`staff`とする。他domainはbusiness intentを表すconsumer-side interfaceまたは明示的orchestrationを通して操作し、owner外へ任意fieldを変更できるgeneric update APIを公開しない。
- cross-domain writeはownerとtransaction境界を明示し、owner外に独立したpersistence実装を作らない。移行中のcompatibility facadeは薄いdelegate/type aliasに限定し、consumer移行後の削除条件を持たせる。
- BE9-2B以降、旧`internal/handler|service|repository`を未移行実装と期限付きcompatibility codeのmigration surfaceとして扱い、新規production実装を追加しない方針で移行した。BE9完了時点で production implementation 0件となり、その後3旧layer directoryは**完全削除済み**である。新規実装は本ADRのtarget domain packageまたは実consumerを持つ命名済みcross-cutting packageへ置く。
- 自動化は安全な手動pathと同じuse caseを再利用し、停止手段、失敗通知、監査、手動fallback、idempotencyまたは明示的retry policyを備える。
- clinical safety、clinic isolation、authorization、auditabilityは効率化より優先する。package配置だけを安全性の証拠にせず、runtime testとapplication invariantで検証する。
- `internal/csvimport` は医院カットオーバー専用のcross-domain例外として、固定21表・固定列契約だけを単一transactionで扱う。`payments.billing_id`一意性と同じ`billing_id`を使うpayment_splits論理親子、completed billingのpayment/completed_at、cash/credit-cardの明示seed binding、split整合もcommit前とread-only REPEATABLE READ verifyで検証する。通常applicationから再利用できる汎用write APIは公開せず、manifest digest、clinic band、6つの明示seed binding、全件検証を満たすoperator commandからのみ呼ぶ。

**Historical landing status (2026-07-22; final outcomeは上記Implementation outcome)**: `staffs`/`shift_entries`と`appointments`のwrite owner一本化は完了した。`appointments`は`de15c7903`で独立したowner外writeとgeneric field-update APIを撤去し、[`appointment_write_owner_lint_test.go`](../../../backend/internal/reservation/appointment_write_owner_lint_test.go)のAST gateで回帰を防ぐ。

L⑥（core `849c27524` / final composition `962ce70e3`）でLSTEPのproduction compositionをtarget package側のtyped `lstep.Application`へ収束した。`cmd/api`が`lstep.Dependencies`からapplicationを組み立て、旧`service.NewServices` / `service.Services`とroot `repository.Repositories`はLSTEP/SharedFileを所有しない。legacy domainへはtyped resultの必要最小限だけを渡し、owner/pet lifecycleはconsumer-side intent interface、legacy DTO/audit変換はcomposition root adapterで接続する。このcutoverでconsumer 0のroot facadeと旧service adapterを削除し、期限付きcompatibility surfaceだけをBE9-2E/2Fへ残した。

SharedFileはroute・use case・persistence・testを単一`internal/lstep` vertical sliceへ移した。4 route、status、storage/error/OpenAPI contractを維持し、POST authorizationは`owners:edit` / `medical-records:create` / `medical-records:edit`のtyped OR要件とする。clinic/staff scopeはJWT由来のみを保存pathへ渡し、URL/body/query値を認可根拠にしない。package移動自体を安全性の根拠とせず、RBAC tuple、敵対clinic test、route snapshot、OpenAPI drift gateで固定する。

Historical landing note: BE9-2Eの最初のdomainとしてtrimmingの23 source rowを単一`internal/trimming` vertical sliceへ収束した（2026-07-23、code tip `297a23fc7`）。handler behavior、request/response、use case、persistence、domain-owned testをtargetへ移し、当時はcentral route/composition/tygoの実consumerを持つcompatibility surface 13件（thin facade/alias 12 + tygo codegen carrier 1）をBE9-2F期限で残した。`appointments`のwrite ownerは引き続き`internal/reservation`であり、trimmingのwriteは`CreateForTrimming` / `LockTrimmingByID` / `UpdateForTrimming` / `DeleteForTrimming`のconsumer-side intentへ限定する。同interfaceはtrimmingに必要なread/validation/booking-lock capabilityも持つが、generic appointment writerや独立persistenceを公開しない。route/RBAC/OpenAPI、clinic isolation、transaction/lock、status/error contractを維持した。13件のcompatibility surfaceとlegacy central consumerは2026-07-24のBE9-2Fで撤去済みである。

#### `appointments` write ownerの実装決定（BE9-2E-0）

`internal/reservation`は`appointments`のpersistenceとlifecycle invariantを所有する。他domainは利用側packageで必要最小限のinterfaceを宣言し、次のbusiness intentだけを呼ぶ。

| owner operation | consumer | transaction境界 | invariant |
|---|---|---|---|
| `CompleteForAccounting` | billing | billingが開始したambient transactionを必須とし、reservationはそのtransactionへ参加。返却用reloadもcommit前に行い、transaction欠落はfail-closed | clinicを固定し、同日同一owner/petの`accounting`または同clinic medical recordに紐づく非terminal appointmentだけを完了化。commit後のreload失敗によるerror-after-commitを作らない |
| `BackfillForMedicalRecord` | medicalrecord | appointment-linked Createでは欠損の有無にかかわらずambient transaction内でrow lockし、appointment直引きのduplicate判定とINSERTまで保持 | 同clinicの`general`予約だけを許可。欠損contextだけを補完し、既存値との不一致、cross-clinic owner/pet/doctor、owner-pet不整合を拒否。通常カルテは予約1件につき有効1件まで、`date`はappointmentのJST日付へ正規化して固定し、`appointment_id`再紐付けも禁止。duplicate lookupと必須依存はfail-closed |
| `PrepareForMedicalRecordFinalization` | medicalrecord | appointment-linked finalized Create/Update transaction内でlifecycle advisory lock→row lockをcommitまで保持。finalized Createではbackfill lockより先に取得 | `no_show`/`cancelled` appointmentのカルテ確定を拒否し、`MarkNoShow`との同時実行を同じlock順序で直列化 |
| `MarkNoShow` | lstep | 候補1件ごとのtransaction内でreservation lifecycle lock、CAS、system監査を同時commit/rollback | clinic/id、`pending`/`confirmed`、終了4時間経過、同clinicの確定済みカルテ不存在を条件とする。stale候補・再実行はno-op。実遷移時は直前status、rule version、評価時刻、batch run IDを監査し、監査失敗時は遷移もrollback |
| `CreateForTrimming` / `LockTrimmingByID` / `UpdateForTrimming` | trimming | ambient transactionを必須化し、advisory lock→appointment row→予約区分共有ロック後にslot/capacityを再検証。staff所属・対応可能種別、course/optionも同じtransactionで共有ロックし、appointment/detail/optionsを同時commit/rollback | 新規Createは同clinicのactiveなtrimming予約区分だけを許可。petから同clinic ownerを導出して`owner_id`を保存し、既存予約の欠損ownerはappointment fieldを変更しないdetail-only writeでもtyped updateにより補完する。pet/owner整合、doctor所属・対応可能種別、course/optionのclinic・active状態、medical record紐付け後のidentity変更禁止、terminal本体/詳細変更禁止を検証し、許可fieldだけをtyped commandで渡す |
| `DeleteForTrimming` | trimming | ambient transactionを必須化し、advisory lock→appointment row→予約区分共有ロック→medical record dependency check→soft delete | 同clinic・非削除かつ非terminalのtrimmingを対象とし、inactiveな既存履歴にも同じguardを適用。medical recordがあればconflictし、同時Createとはrow lockで直列化。一般診療appointmentは同clinicでもNotFound |

LIFF予約はtransactor・reservation repositoryを必須とし、public/activeな予約区分、LINE顧客のclinic所有、明示staffの所属・対応可能種別・`is_active=true`・`reservation_visible=true`、activeなcourse/optionをwrite transaction内で再検証する。具体repositoryの`FOR SHARE`をappointment/detail/optionsのcommitまで保持し、必須repository欠落またはいずれかの保存失敗はfail-closedとして予約本体を含めてrollbackする。

通常カルテのsoft deleteは対象カルテを`FOR UPDATE`したtransaction内で見積依存を再確認し、`clinic_id + id + status=draft`を単一DELETE条件に含める。見積Createも親カルテ行を先にlockするため両者は直列化され、見積が先なら削除をConflict、削除が先なら後続見積を拒否する。確定処理が先行した場合や既に非draftの場合もConflictとし、確定済みカルテを削除しない。

予約キャンセル後のdraftカルテcleanupもこの通常削除経路へ委譲し、旧repository bypassは持たない。一方、予約更新とカルテcleanupは既存contractどおり別transactionのbest-effortであり、部分成功時の再収束方式はBE9-2Eまでのcross-domain orchestration判断としてgit履歴（旧`BE-refactor.md`・2026-07-24退役）に記録済みである。これはowner外の`appointments` write例外ではない。

締め後の会計編集はtransaction内監査を必須とし、監査dependency欠落または監査write失敗時は編集自体をrollbackする。

owner内の汎用`update(map[string]any)`は非公開primitiveとし、owner外へ任意field更新能力を公開しない。`appointment_write_owner_lint_test.go`はproduction tree全体をAST走査し、`FirstOrCreate`を含むGORM mutation、query変数・typed引数・typed parameter（slice/arrayとnamed/alias typeを含む）またはfree function/receiver method戻り値由来の`model.Reservation`、cross-fileおよび宣言戻り型で解決したpackage-qualified free/receiver factory、直接または変数代入した`TableName()`、package/local定数、table alias、schema-qualified table、静的helper戻り値またはschema-qualified tableのraw SQLを使うwrite、広いrepository依存、`map[string]any`/`map[string]interface{}`のnamed/aliasを含むgeneric appointment mutation capabilityを拒否する。clinic isolation（一覧JOIN/search JOINのpet/owner一致とnested Preloadの中間detail scopeを含む）、master-FK、権限resource境界、LIFFのinactive・非表示staff拒否、ambient transaction必須、owner/line customer/予約区分/staff assignment/capability/course/option共有ロック、通常カルテのgeneral category・JST date・予約1件1カルテ、カルテ確定・見積作成・draft削除の競合、slot/capacityのlock後再検証、terminal detail/削除不変、no-showとカルテ確定の双方の先行順序、detail-only writeを含むtrimming owner永続化、必須validator・監査依存のfail-closed、idempotency、監査を含むrollbackはruntime testで検証し、package配置だけを成立根拠にしない。

分類manifest 761 source rowの割り当て（`target package` / `現状維持` / `削除`）は [be9-2a-classification-manifest.csv](../be9-2a-classification-manifest.csv) に記録（未分類0件、削除0件）。per-domainの9列data（owned types/routes/queries、consumers、change frequency、fan-in/out、transaction有無、route有無、tenant boundary）は [boundary map doc §3](../be9-2a-boundary-map.md#3-per-domain-boundary-map9列) を参照。

### (b) 代替案

1. **layer-first維持（ADR-006採択時点の状態、2026-07-19 snapshot）**: 当時の`internal/handler`(269 file)、`internal/service`(202 file)、`internal/repository`(154 file、42 subpackage含む)という3巨大packageを維持する案。却下理由: ADR-005は固定layerをGo/Gin公式要件として扱うことを廃止し、本ADRは実測した凝集性と変更単位を根拠にlayer-first維持を却下した。当時は`handler.go`、`service.go`、`repositories.go`が巨大なcomposition rootとなり、1機能の変更単位とpackage境界が一致していなかった。件数は現在値ではなく採択時snapshotである。
2. **repository層のみADR-006採択時点の42 subpackage粒度をhandler/serviceへも機械的に踏襲**: repository/checkup, repository/account等の当時の42分割を単純にhandler/serviceへも同数複製。却下理由: 42は「エンティティ単位」の粒度でありRBAC resource・route構造から見た「業務ドメイン」単位（例: checkup/diagnosis/examination/vaccine/prescription等18サブトークンが全てmedicalrecordという1業務ドメインに属する）とは一致しない。42 packageに分割すると、本ADRが検出した10組の生cycleがそのまま42×42の組み合わせで再現され、収拾がつかなくなる。
3. **§9の69ドメイン粒度をそのまま採用**: 却下理由。§9は`internal/service`のみの実測であり、handler/repository/modelを含まない。また69ドメインは「抽出時のコンパイル安全性」の単位であって「業務境界」の単位ではない（§9自身が"抽出安全性の主軸はoutgoing"と明記、業務的凝集は別軸）。boundary map §6の差分表が示す通り、69ドメインの多くは実際には1つの業務ドメイン（medicalrecord等）に集約される。
4. **line/liff/lstepをドメイン境界で最初から3分割**: 却下。BE9-2A時点では実測不足のため保留していたが、2026-07-20の論点#2で**単一target package `internal/lstep`**に確定した。sub-batchは同一package内の変更・review単位として使い、外部consumerが限定subsetだけを継続利用する実測が得られた場合のみ、本ADRを改訂してsubpackage分離を再評価する。

### (c) cycle・security risk

実測で10組の生の双方向依存（Go import cycle相当）を検出した（当初7組として起票したが、santa dual-review round 1でarchitectレビュアーが追加3組を検出——lstepが所有する`LstepTagSyncService`interfaceがmedicalrecordだけでなくowner/pet/billingの計4ドメインからfieldとして直接保持されていたため。詳細は[boundary map §5](../be9-2a-boundary-map.md#5-許可依存グラフdagと生cycleの解消方式)の追記参照）。うち9組は本プロジェクト既存の規約（[go-gin-backend-guidelines.md §3](../../../.claude/rules/go-gin-backend-guidelines.md)「interfaceは利用側で定義」）に基づくconsumer-side interfaceで解消可能——利用側packageが必要最小限のinterfaceを自分のpackage内で宣言し、所有側packageの具象型がそれを構造的に満たし、DIルート（現`service.go`の後継）で配線する。この方式では利用側packageは所有側packageをGo importしない（コンパイル時cycle解消）。

| # | cycle | 解消方式 | 詳細 |
|---|---|---|---|
| 1 | staff↔auth | staffがAccountCreator相当のinterfaceを宣言 | [boundary map §5](../be9-2a-boundary-map.md#5-許可依存グラフdagと生cycleの解消方式) |
| 2 | owner↔billing | ownerがInsuranceExistenceChecker相当のinterfaceを宣言 | 同上 |
| 3 | medicalrecord↔billing | billingがMedicalRecordOwnershipChecker/DraftLocker相当のinterfaceを宣言 | 同上 |
| 4 | reservation↔trimming | reservationがmasterFKOwnershipChecker相当のinterfaceを宣言 | 同上 |
| 5 | clinic↔auth | clinicがPermissionGroupWriter相当のinterfaceを宣言 | 同上 |
| 6 | medicalrecord↔lstep | medicalrecordがTagSyncNotifier相当のinterfaceを宣言 | 同上 |
| 8 | owner↔lstep | ownerがTagSyncNotifier相当のinterfaceを宣言 | 同上（round1レビューで追加検出） |
| 9 | pet↔lstep | petがTagSyncNotifier相当のinterfaceを宣言 | 同上（round1レビューで追加検出） |
| 10 | billing↔lstep | billingがTagSyncNotifier相当のinterfaceを宣言 | 同上（round1レビューで追加検出） |

**7組目（reservation↔staff、model.Staff/model.ShiftEntryへの共有テーブル二重書き込み）はinterface逆転だけでは解決しない真の依存cycleとして検出された**（[boundary map §7.1](../be9-2a-boundary-map.md#be9-2a-reservation-staff)）。2026-07-20に`staffs`/`shift_entries`のwrite ownerを`staff`へ一本化する案(A)を採用し、2026-07-21のreservation Phase 0で実装・clinic-isolation testまで完了した。履歴と却下案は「論点の解決記録」#1に残す。

**#8-10追加検出の影響評価（round1 architectレビュー）**: この見落としはDAG構造自体（13 package taxonomy、トポロジカル順序）を無効化しない——3組とも#6と同一の解消方式で解消し、lstepはトポロジカル順序で最上位のまま変わらない。Goはimport cycleを物理的に拒否するため、この見落としが実害化する経路は「BE9-2Cでowner/pet/billingへ着手した際のコンパイルエラー」であり、「サイレントなtenant isolation不具合」ではない——census網羅性の是正であり、構造的な誤りの訂正ではない。

**tenant boundary上のセキュリティリスク**として以下を明記する（ADR-002との整合確認）:
- `POST /api/line/webhook`（lstepドメイン）は`clinic_id`なしで受信するため、署名検証段階で全`LineReservationSetting`のLINE channel secretを読む意図的なcross-clinic-identity readを行う。この例外は受信前identity解決だけに限定し、ownerを全clinic横断で検索・更新する権限へ拡張しない。異なるclinicで同じsecretが一致する曖昧系はfail closedとし、一意に一致したclinic IDを以後の`FindByLineUserID`とfollow/unfollow updateへ必須scopeとして渡す。owner未登録のtyped NotFoundだけをno-op、真のlookup/update errorはnon-2xx retryへ伝播する。更新CASは`clinic_id + owner id + expected line_user_id`と正数かつ受信時刻+5分以内のLINE event timestampを必須とし、followは保存済みfollow/unfollowの両時刻より新しい場合だけ、unfollowは保存済みfollowと同時刻以上かつ保存済みunfollowより新しい場合だけ適用する。したがってstale・duplicate・out-of-order・再連携前IDは`RowsAffected == 0`の安全なno-op、同時刻はunfollow優先となる。これをADR-002「no unscoped read」不変条件に対する**限定された証拠ベースの例外**として承認する。
- `BillingConfirmation`/`BillingItem`/`Payment`（billing）、`Inquiry`/`CarePlanItem`/`Treatment`/`ClinicalPlan`/`MedicalRecordImage`（medicalrecord）は自前`clinic_id`を持たず、親レコード（billing/medical_record/hospitalization）経由の間接isolation——親側のownership検証が壊れると連鎖的にisolationが崩れる構造的リスクとして明記する。BE9-2B以降のリファクタでこの間接isolationパターンを壊さないこと（当初`Inquiry`のみと記載していたが、reviewで同型resourceを追加検出——詳細は[boundary map §7.3](../be9-2a-boundary-map.md#be9-2a-indirect-isolation)）。`StaffReservationExclusion`も同型（自前clinic_idなし）だが、こちらは新規発見ではなく既に`preload_clinic_scope_lint_test.go`でsite-exception追跡済みの既知の低severity読み取りギャップ。`ExamResult`/`ExamTypeField`/`StaffNote`も検証済み——生きた漏洩なし（詳細はboundary map §7.3）。
- **起票時の記録（2026-07-21に是正済み）**: `BillingItem`は間接isolationの中で唯一、repository層の防御が実質的にno-opだった（[boundary map §7.4](../be9-2a-boundary-map.md#be9-2a-bug-417)）。`billing_item_repository.go`のUpdate/DeleteがGORMの`Joins()`をUPDATE/DELETE SQLへ伝播しない罠に該当したが、当時もservice層の事前check（`FindByID`）でgateされており生きた漏洩ではなかった。BE9-2Aではmeasurement/document-onlyのため修正せず、2026-07-21のbilling Phase 0でsubquery形式への是正とクロステナントtest追加を完了した。詳細は「論点の解決記録」#6。
- `AnimalSpecies`（pet）、`Company`（clinic、シングルトン）、`LstepAutoManagedPrefix`/`LstepConditionTagMapping`/`LstepSendPurposeTagPrefix`（lstep）はglobal-master。BE9-1でlintのsource discoveryはpackage非依存化済みで、preload lintは`AnimalSpecies`を明示的なglobal exemptionとして持つ。他のglobal-masterも新しいassociation/preload pathを追加する際にclinic predicateを誤強制しないよう、この分類をreview根拠にする。

master-FK-write lint（実装fileは`backend/internal/lintscan/master_fk_write_inventory_lint_test.go`）はapplication write-roleのreview-coverageゲートであり、上記global-master allowlist（clinic-id-isolation/preload lint向け）とは対象・目的が異なる。GORM column definitionとHTTP DTOはwrite operationではないため意図的にscope外とし、package layerを理由には除外しない。role filterはwrite logicを持つdomain packageへ広げる単一の拡張ポイントを持つ。

### (d) 段階移行方法

`BE9-0(完了) → BE9-2A(本ADR) → BE9-1(lintのpackage非依存化) → BE9-2B(pilot 1件) → {BE9-2C↔BE9-2Dをlargest-ready方式で反復} → BE9-2E-0(write owner収束) → BE9-2E → BE9-2F → BE9-3 → BE9-4` の順序で固定する（BE-refactor.mdより）。

移行はstrangler pattern（一斉移動禁止）とし、各batchは以下を満たす：
1. target依存graphがacyclic（本ADRの許可依存グラフに従う。consumer-side interfaceで解消する9組に加え、reservation↔staffはwrite owner一本化済み。owner/pet/billingへ着手する場合は#8-10のlstep notifier interfaceも忘れずに解消すること——見落とすとコンパイル時import cycleで即座に失敗する）
2. tenant/認可/transaction/clinical safetyのbaseline testが存在
3. route/API/SQL互換性とrollback単位が定義済み

上記3条件を満たす**ready frontier**からproduction行数・file数・旧aggregator/call-site削減量が最大のdomainを選ぶ（largest-ready方式）。boundary map §8のlstep内部sub-batch案が示す通り、大規模domain内部もsink-first（out-dom=0のsub-batchから）で段階抽出する。

### (e) migration batchごとのbaseline test・移動対象・call-site・facade削除条件

以下はBE9-2C/2D着手時に各batchが固定すべき項目のテンプレート（本ADRの時点では値を確定しない——boundary map §5の許可依存グラフとトポロジカル順序が入力）：

- **トポロジカル順序**（本ADRの許可依存グラフをacyclicity検証した結果、[boundary map §5](../be9-2a-boundary-map.md#5-許可依存グラフdagと生cycleの解消方式)のedge解消後）: `httpapi → clinic → inventory → manualarticle → owner → pet → staff → auth → reservation → trimming → billing → medicalrecord → lstep`（13 node、45 edge、cycle 0件。round2 santa dual-reviewの指摘を受け、実際の検証コマンド`python3 be92a_toposort.py`とその出力をboundary map §5.0に明示——「機械検証済み」の根拠を可視化した）。
- **baseline test**: 移行対象domainの既存test（`*_test.go`）をそのまま維持し、移動後も同一test名で green を維持する。tenant isolation runtime test（clinic-scoped/global-master/cross-clinic-identity、boundary map §3の分類）は移行前後で挙動が変わらないことを追加確認する。
- **移動対象**: [classification manifest](../be9-2a-classification-manifest.csv)の`target:<domain>`行。ただし10組のcycleのうち解消方式(c)を適用するファイル（例: `reservation_validators.go`, `owner_service.go`, `clinicService.CreateClinic`等）は、interfaceの宣言・配線変更を伴うため通常のfile移動より慎重なレビューが必要。
- **call-site（BE9-2A時点のsnapshot）**: 当時の`internal/service/service.go`（108 field）、`internal/repository/repositories.go`（95 field）、`internal/handler/handler.go`（44 route registration呼び出し + 95 fileの共有receiver）を移行入力として記録した。これらの数値は現在値ではなく、現行consumerは各batch開始時に`rg`で再実測し、BE9-2Fで集約を最終解体する。
- **compatibility facade / 削除条件**: 移行中は型alias/薄いdelegateとして許可し、call-site移行完了後に削除期限を設ける（BE-refactor.md BE9-2Cの既定方針）。`internal/repository/base.go`/`helpers.go`/`transactor.go`の非exportヘルパーは`repository/repohelpers`の同等export版が既に存在するため、対応するflat root packageの分割完了時点で削除する（boundary map §4.2）。
- **business/data ownership**: 変更対象のbusiness fact、source of truth、write owner、owner外の全write call-siteを列挙する。owner外の直接writeはowner APIへ収束させるか、例外の理由・transaction・削除条件をADRへ記録する。
- **product/automation gate**: user workflowの入口から完了までを同じsliceで検証する。自動処理を含む場合は停止、失敗通知、監査、手動fallback、idempotency/retryをbaseline testまたは運用contractに含める。

## 論点の解決記録（起票時の着手前ゲート）

BE9-2B完了時点では後続phaseの着手前ゲートとして残していたが、2026-07-21までに以下6項目はすべて裁定、検証または是正済みである。起票時の根拠と実装条件を履歴として残し、新しい未解決事項は[`todo.md`](../../../todo.md) / [`todo-po.md`](../../../todo-po.md) へ登録する。

> **2026-07-20 委任裁定／2026-07-22状態同期**: ユーザー（MinoruSoga）がPO判断をAIへ委任したため、アーキテクチャ判断である論点#1〜#4を裁定した。論点#6は当時Openの技術的是正ゲートとして残したが、2026-07-21のbilling Phase 0で是正済み。臨床安全に関わる判断（#201等）は本委任の対象外であり、このADRでは裁定していない。

1. **Resolved (2026-07-20・委任裁定): 案(A) staffを唯一の書き込み者とする** — **reservation↔staffの共有テーブル二重書き込み**（[boundary map §7.1](../be9-2a-boundary-map.md#be9-2a-reservation-staff)）。`staffs`テーブル（reservation固有カラム`staff_type`/`reservation_visible`/`reservation_comment`/`sort_order`含む）と`shift_entries`テーブルの書き込みをstaff domainのexportedメソッドへ一本化し、reservationはconsumer-side interfaceで呼ぶ（reservation→staffエッジは許可依存グラフのtopo順（staffがreservationより上流）に整合し、cycleを生まない）。**案(B)（reservation固有カラムの別テーブル切り出し）を却下した根拠**: ①DDL・データ移行・OpenAPI/codegen再同期・既存DBのDB_RESETを要し、BE9の「behavior-preservingな移動と機能変更を混在させない」原則に反する ②staff行の属性を2テーブルへ分離するとライフサイクル同期（作成・削除・クリニック移動）の二重管理を新設する（product-philosophy ②違反）③案(A)は将来案(B)へ移行する選択肢を閉じない。**実装条件**: (i) shift_entriesも同一方針（`reservation_schedule_repository.go`の書き込みはstaff domain経由へ）(ii) 一本化後のstaff側write pathに既存のclinic-scope検証（`validateClinicScopedMasterIDs`同等）を維持し、クロステナント分離testを追加 (iii) read pathは移動対象外（preload lintが継続監査）。 **実装完了（2026-07-21・BE9-2C reservation Phase 0）**: staffs 4 write（Create/Update/Delete/UpdateSortOrder）を`staff_repository.go`の`*ForReservation`メソッドへ、shift_entries 2 write（Save/Delete）を`repository/shiftentry`の`SaveByStaffDate`/`DeleteByStaffDate`へbody移動（byte等価・セマンティクス統合なし）。reservation側2 repoは1-line delegate化（interface無変更）。条件(ii)のクロステナントtestは`staff_reservation_write_clinic_isolation_test.go`+`shiftentry/reservation_write_clinic_isolation_test.go`で移動先メソッド直叩き追加。
2. **Resolved (2026-07-20・委任裁定): 単一`internal/lstep`で確定** — **lstepのline/liff内部分割**（(b)代替案4）。根拠: ①§2是正でLIFF系9 fileが既にreservationへ再分類され、3分割の主要動機（liff境界）が縮小済み ②sub-batch構造（tag_sync/health_tag/delivery/settings/batch/csv）で運用粒度は確保でき、package境界の先行分割は10組cycle検出時と同型の結合表面を実装知見なしに増やす（不要なinterfaceを先行作成しない原則）。**再評価トリガー**: BE9-2D lstep実装中に「外部packageがline専用subsetのみをimportする実消費境界」が出現した場合のみ分割を再検討（本ADR再改訂で記録）。
3. **Resolved (2026-07-20・委任裁定): 依存エッジを追加しない（実測を正とする）** — **clinic↔reservation/trimmingの想定依存が実測で確認できなかった**（[boundary map §7.2](../be9-2a-boundary-map.md#72-clinic--reservationtrimming-の消費関係が実測で確認できなかった)）。2026-07-20の再grep でClinicHoliday/ClosingSpecialPeriodの消費者はclinic domain CRUD（holiday/closing_settings系）とaccounting_report_service（billing→clinicエッジ、topo順に整合）のみと確認——reservation/trimmingからの参照ゼロは測定漏れではなく実装の事実。存在しない依存を想定で許可依存グラフへ追加しない。**残る製品論点（BE9スコープ外・裁定対象外）**: 休診日・臨時休業が予約可否判定に反映されていない可能性は、シフト駆動の空き枠設計（シフトが無い日=枠が無い）で意図的にカバーされている可能性があり、要件化するならproduct-philosophy ①に従い責任者名付きで別Issue起票する。BE9はbehavior-preservingのため実装しない。
4. **Resolved (2026-07-20・委任裁定): medicine/vaccine=medicalrecord帰属、line_reservation_setting=reservation帰属で確定**（いずれも概念タグのみ・model一括現状維持方針は不変でファイル移動なし）。根拠: medicine(+dose_param)/vaccineは薬量計算・予防接種という臨床ワークフローが主消費者で、manifest上も対応するhandler/service/repositoryが全てtarget:medicalrecord（在庫観点はmerchandise_item/inventoryが別途担う）。line_reservation_settingは§2是正でLIFF予約フロントエンドがreservation帰属となった以上、その設定モデルもreservation所有が整合的（boundary map §4.4の再検討推奨の通り）。
5. ~~ExamResult/ExamTypeField/StaffNoteの間接isolation保護パターン未検証~~ — **本タスク内で検証完了**（[boundary map §7.3](../be9-2a-boundary-map.md#be9-2a-indirect-isolation)）。3件とも生きた漏洩なし（ExamResultはpre-check+tx方式、ExamTypeFieldは読み取り専用master、StaffNoteは追記専用でupdate/delete method自体が存在しない）。決裁事項からは除外。
6. **Resolved (2026-07-21・BE9-2C billing Phase 0): `billing_item_repository.go`のUpdate/Delete防御ギャップ**（[boundary map §7.4](../be9-2a-boundary-map.md#be9-2a-bug-417)、santa dual-review round 2で検出）— 起票時はservice層の事前checkでgateされており生きた漏洩ではなかったが、repository層の防御が実質no-opだった。Update/DeleteをEXISTS subquery述語へ是正し、クロステナント分離テスト4本（`billing_item_write_clinic_isolation_test.go`・述語削除mutationでRED化実証）を追加してbilling domain着手前提を充足した。

## Consequences

### Positive

- 分類manifest 761 source rowの帰属が実測（RBAC resource名・route構造・GORM ClinicIDフィールド・Goのreceiver型制約という一次証拠）に基づき機械的に検証可能な形で確定した。旧BE8/§9のfilename-prefixのみの分類の限界（receiver-vs-filename不一致等）を後続batchでも訂正できる形にした。
- 10組の生cycleのうち9組は、新規abstractionを発明せず、プロジェクト既存のconsumer-side interface規約（ADR-005/go-gin-backend-guidelines.md §3）だけで解消可能であることを実証した——BE9-2B以降で「解決策が見つからず立ち往生する」リスクを大幅に下げた。
- medicalrecordドメインのBE9-2A初回分類が185 source row（旧見積96の約2倍）と判明し、batchサイジングを補正できた。checkup_sync再分類後の現行manifestは175 source row、2026-07-24 live treeの`internal/medicalrecord`はproduction Go 180 fileであり、これらは別指標である。同様にbillingはmanifest 65 source rowに対して現行production Go 71 fileで、snapshot provenanceと移行後の物理file数を混同しない。

### Trade-offs

- write ownerを一本化すると、owner packageのAPIとtransaction境界がcross-domain変更の調整点になる。特に`staffs`/`shift_entries`は`staff`、`appointments`は`reservation`を経由するため、直接table writeより依存関係が明示的になる一方、owner APIの互換性管理が必要になる。
- lstepを単一packageとして維持するため、内部の機能群はfile/型とvertical sliceで整理する必要がある。外部consumerが限定されたsubsetだけを継続利用する実測が得られた場合は、subpackage分離を再評価する。
- consumer-side interfaceによるcycle解消は、実装時に新規interfaceの命名・配置・DI配線という設計判断を都度必要とし、単純なfile移動より工数がかかる。

## References

- [BE9-2A boundary map（本ADRの実測データ正本）](../be9-2a-boundary-map.md)
- [BE9-2A classification manifest（761 source rowの分類）](../be9-2a-classification-manifest.csv)
- [ADR-005: Go/Gin公式ベースラインとpackage architecture](005-go-gin-backend-guidelines.md)
- [ADR-002: マルチテナント設計 — clinic_id完全隔離](002-multitenancy-clinic-id-isolation.md)
- [go-gin-backend-guidelines.md](../../../.claude/rules/go-gin-backend-guidelines.md)
- 旧BE-refactor.md BE9-2A（2026-07-24退役・経緯はgit履歴）

## 現行実装への補足（2026-09-06）

本文の BE9 measurement・file 数・移行時の tenant 分類は履歴として保持する。現行 contract は以下の source と照合する。

- `internal/lintscan/package_boundary_gate_test.go` は **35 top-level package / 14 domain** を pin する。`seedlogin` は `cmd/migrate` の非本番デモ upsert に加え、`auth/auth_service.go` の catalog 限定非本番認証補助からも使われる。cmd-only とは分類しない（[例外 package 規律](../exception-package-discipline.md)）。
- §(c) の「`Payment` / `BillingItem` / `ExamTypeField` は自前 clinic なし」は採用時の記録である。現行 model では `Payment.ClinicID` / `BillingItem.ClinicID` / `ExamTypeField.ClinicID` が存在し、DDL では `billing_items` / `treatments` / `appointment_trimming_options` の clinic が親から複製される。GORM field の有無と DDL 列の有無は別指標。現行の複合 FK / RLS は [ERD](../erd.md) と `001_init.sql` を参照する。
- [GitHub #249](https://github.com/MinoruSoga/AnimalEkarte/issues/249) の Phase 2 には `exam_type_fields` の direct clinic scope への移行要求がある。現行の同表、`exam_types` / `exam_reference_ranges` の複合 FK は DDL に実装されている。Issue は取得時点で OPEN であり、臨床 range 承認などの受入まで完了したとは扱わない。
- nested owner/pet 登録は `owner.PetRegistrar` → `pet.CreateForOwnerRegistration` が同じ ambient transaction に参加する。owner 外の独立した pet insert 経路を作らない（[cross-domain catalog](../cross-domain-orchestration-catalog.md) の `PATH-OWNER-PET-REGISTER`）。
- 上記は HEAD `7c6592f9f` のコードと DDL の静的照合であり、実 DB 適用、STG/PROD の release gate の判定ではない。
