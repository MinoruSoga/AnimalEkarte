# EMR-464 預り金の初回受領・同口座再読込契約

選定日: 2026-10-23。対象: EMR-464 の初回backend受領・結果照会・同口座読取。

本書は選定済みの限定技術契約である。業務正本を変更せず、実装担当間の入力・出力・保存・認可・競合・受入境界を定める。共有成果の供給、編集claim、製品実装、試験、DB適用、利用開始は未完了であり、技術選定を実装済み・利用可能と読み替えない。

EMR-506の一般公開fund API・一般wire候補は未採択のまま。本書が選定する506側の範囲は、464初回受領に必要な内部同TX participantとproducer/readerの最小subsetだけである。

## 0. 対象範囲

最初の独立受入単位を **予約指定なし、同医院・同飼主、新しい実受領1件・1方法、その全額を明示了承により預りへ入れる backend API と同口座読取** とする。業務全体を狭める決定ではなく、実装順の限定である。既存fundの余剰分、複合方法、予約指定、請求充当、解除、返金、記録訂正、帳票、FE画面、締めはこの単位に含めない。

- EMR-505 の請求計算・未払計算はこの単位では使わない。全体Done待ちは不要。
- EMR-506 は新fund1件・deposit_allocate1件・同TX producer event/固定受領readerの必要subsetだけ。506候補公開route全体や既存fundの全writer CASを持ち込まない。
- EMR-465 は採択済み `ValidateNewReceiptMethodsInTx` の1方法集合で参加。別receiptは作らない。
- EMR-489 は実在C06/Q/claim metadata/lookupを再利用。ただし元飼主anchor、台帳とbindingのidentity接続、適用すべきgateのcontext伝播はまだ供給が必要。
- 現行DB/modelが存在することと、consumerへそのまま接続できることは違う。既存台帳の `(clinic,owner,operation_key)` とC06のactor/intent付き一意性のずれを、以下の限定DDL要求で閉じる。
- **本票でこれから実装するAPI/participant/readerが未実装であること自体はReady不備ではない。** 本書の範囲外で未選定のwire、又は先行供給と同PR担当範囲のどちらにも割り当てられていない共有subsetを実装担当の裁量へ残したままclaimしない。同PRでowner/hunk/型/ACまで合意した未実装subsetは、先行mergeを不必要に待たない。

本書の選定に伴う製品コード・DDL・API YAML・生成物・設定の変更はない。DB適用・実データ確認・テスト・build・codegen・runtime検証は NOT_RUN。新規routeや接続契約の選定と、既存実装の存在は以下で区別する。

## 1. 正本と既存実装の確認範囲

既存製品実装の確認基準は `04a9078d93c3899705ce9d420eae75e89709e2e3`。後続の `bd309de7b9680a46fdf3ed4a89491c198624288c` はチケット規約文書の変更であり、本書の製品source・技術仕様の根拠を変更しない。DBの既存最終形は001全文と末尾ALTER/DROPを含めて照合している。以下の固定SHAリンクは既存実装の根拠であり、本書で選定した新しい接続が実装済みであることを示さない。

主要固定参照:

- [業務正本 accounting/lifecycle](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/spec/domains/accounting/lifecycle.md): 飼主口座、明示了承、実受領と配分の分離、現金純額、新受領時の入力条件。
- [accounting technical-spec L45–55, L195–228](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/design/domains/accounting/technical-spec.md#L195-L228): 預り独立intent、原子的履歴/receipt、認可mapping、予約なし預りの元clinic/owner/fund保護。
- [finance-operations technical-spec L111–119](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/design/domains/finance-operations/technical-spec.md#L111-L119): 操作/結果/現在read権限/unknown/監査。
- [465採択済み限定契約](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/work/emr/emr-449-459-465-payment-master-contract-20261021.md): guard入力、Q、method先行lock、停止writer、465-M/465-C境界。
- [506未採択候補](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/work/emr/emr-506-fund-wire-design-candidates-20261022.md): 新wire候補と既存Complete互換、fund全体versionなし。
- [C06参加契約](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/design/shared/operation-participation.md#L33-L70)、[contract.go](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/operation/contract.go)、[runner.go](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/operation/runner.go)、[store.go](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/operation/store.go)。古いgeneric説明の「claim metadata未接続」より現在の実装済みhookを優先する。
- [最終保存model L519–590](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/model/storage_finance.go#L519-L590)、[既存billing route](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/billing/routes.go)。専用預りroute/consumerは存在しない。

### 選定済み事項と未供給成果

|区分|内容|
|---|---|
|既決|実受領はfund、預りは同fundのdeposit_allocate。別受領を作らない。元医院/元飼主を維持。任意予約は業務として可能。自動失効/無断繰越なし。履歴/必須監査/成功receiptは同TX。停止後の成功replayは再検査で失敗へ変えない。|
|既決|C06 binding scope、HMAC/Q、現在元対象認可、unknown時に新keyで救済しない。465のmethod行NOWAIT→domain lock順。|
|本書で選定する限定技術契約|初回予約なし・1新fund全額限定、以下の3route/DTO/HTTP、owner_kind/intent、台帳binding接続、空口座revision表現、source eventのproducer shape、限定共有API。|
|現行との具体的接続差分|台帳key範囲とC06 scope差、未成功anchorなし、Gateのcontext返却非対応、C06 no-waitと現在graph gate helperのblocking差、専用reader/consumerなし。新464 callerが未実装なので「稼働中464の不具合を再現した」とは述べない。|
|未受領の共有成果|446/466の必要DDL契約受領、489/467の限定gate adapter受領、465 guard実装受領、506内部participant受領、編集owner/実行ID対応。下記で必要成果を指定し、関連親票全体のDoneにはしない。|

## 2. 初回単位・権限・operation family

### 2.1 新規3route

|method/path|用途・権限|
|---|---|
|`POST /api/v1/owner-deposit-accounts/{owner_id}/receipts`|本初回受領だけ。認証/選択医院/shape検査後C06へ。新規Apply内で現在 `accounting:create`。pre-routeのcreate一律拒否でread許可済replayを妨げない。|
|`GET /api/v1/owner-deposit-accounts/{owner_id}`|同医院・元飼主の現在口座。現在 `accounting:view`。完全read-only。|
|`GET /api/v1/owner-deposit-receipt-operations/{operation_key}`|このfamily専用の同actor結果照会。現在 `accounting:view` を保存anchorの元飼主に対して評価。URLに別ownerを渡さない。|

以下は本書で選定する新規routeであり、確認基準の既存APIには未登録である。既存 `POST /api/v1/accountings/complete` の整数wire・初回201/再送200・Billing responseに変更なし。旧accounting PATCHや既存refund routeを預り受領として呼ばない。

C06 Scopeは認証済み選択 `clinic_id`、`owner_kind="accounting"`、`actor_kind="staff"`、認証staffのcanonical decimal `actor_ref`、`intent="owner-deposit.receive-new"` に固定する。clientにclinic/actor/intent/owner_kindの上書き口を与えない。別actor・別医院は同じUUIDでも別scopeであり、回復のためactor/医院を変える操作は禁止。別actorのreceiptを共通key検索で発見・返却しない。

本書で既存grantの操作mappingを選定する。新職種・新grant・全スタッフ許可は追加しない。利用開始には後述の実装・供給・認可・競合の条件を満たす必要がある。認証中のselected clinic membershipも現在検査する。system_admin等の既存明示的bypassを独自に増減しない。

### 2.2 create/viewの分岐

|現在権限|未成功の新規Apply|成功済POST再送/GET lookup|GET口座|
|---|---|---|---|
|create + view|可|可|可|
|createのみ|可。初回write成功receiptはそのwrite応答として返す|不可|不可|
|viewのみ|不可|元対象の現在認可を満たす成功のみ可|可|
|両方なし/医院非所属|不可|非開示|不可|

このcreate/view分岐を本書の権限mappingとして選定する。現コードの動作だけを業務許可の根拠にしない。採択finance-operations technical-spec:116は権限失効後の再送結果を現在read権限で判定し、:119は既存操作grantと依存write grantを要求する一方、createがviewを内包するとは定めない。accounting technical-spec:220–225も新預りのmapping不足時は停止を要求する。この正本の境界に従い、新受領は既存accounting:create、保存結果と口座の読取はaccounting:viewへ対応付ける。createにviewを暗黙継承させない。現C06が新規成功で `AuthorizeLookup` を通らず、replayでは通る（runner.go:333–335,405）のは、この選択を実現できる実装根拠に過ぎない。FE接続は後続だが、受領から再読込までの画面導線はcreate/viewと既存method候補読取権限を全て満たす場合にのみ提供する。POSTにcreate middlewareを無条件再利用して、create失効後/view保持の成功replayまで403にしない。失敗/statusの開示はC06の元対象現在read検査を保持する。

createのみで初回成功応答が失われた場合、操作したstaffは同key lookup/replayを現在view不足で拒否される。この権限境界により、同actorのview権限が通常の権限管理で回復するまで結果確認を保留し、権限を自動付与せず、別actor/別keyで再受領しない。新規Applyへview必須を暗黙追加せず、既存grantの意味を変更しない。

### 2.3 現在認可readerの実在接点と同PRの新adapter

[auth/http_permission.go:25–60](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/auth/http_permission.go#L25-L60) の `HasPermissionInClinic` 経路はrequest memoを使う。HTTP admission用として存在するが、gate後の保存直前現在grant検査をこのmemoだけで済ませない。

[auth/permission_group_repository.go:370–397](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/auth/permission_group_repository.go#L370-L397) の `FindAllEffectivePermissionsByStaffID(ctx, staffID, clinicID) ([]model.PermissionGroupRule,error)` は実在し、DBOrTxで同ambient TXを使用、active/nondeleted groups/rulesをclinic限定で集約する。create/viewの独立flagも実在する。このreaderはstaff/account有効性や現在所属を自動で保証しないため、その再検査は別に必要。

本PRで実装する新adapterは `CheckOwnerDepositAuthority(ctx, authenticatedIdentity, clinicID, ownerID, action) (allowed bool,error)`。actionはcreate/viewのみ、clinic/actorはserver認証identityとの一致必須。入力identityのIDだけを参照し、認証accountとstaffの現在対応、staff/accountのactive/nondeleted、選択clinicのactive/所属、元ownerのclinicを現在DBから再読する。通常staffの所属はstaff_clinic_assignmentsを使い、古いstaffs.clinic_idへ推測fallbackしない。system_admin bypassは既存方針と同じ対象に限り、現在accounts.is_system_adminを再検査する。許可ruleは前記fresh readerで判定、memo/保持boolだけを再利用しない。SQL/依存障害はerrorでfail-closed、明示拒否と区別し、C06の開示契約に従う。

Applyからはgate取得済みambient TXを必須にして呼び、auth/staff/ownerに加え、選択clinicの有効/無効・削除writerを含む第7.4節の参加閉包で現在状態をfenceする。clinics.is_activeは非key列なのでFK用KEY SHAREだけでは更新を防げない。保存側は第7.3節でclinicにSHARE NOWAITを取得してactiveを再読し、共通gate参加と既存row fenceを両方保持する。lookup/readでは現在認可readerとして呼び、口座GETでは認可対象とprojectionを同一read snapshotで検査する。新規受領は非削除ownerを要求する一方、保存済み履歴のsoft-deleted ownerは既存履歴閲覧規則に従い、現在のpet帰属や一覧非表示だけで元receiptを別ownerへ移さない。履歴閲覧の対象が解決できない場合はfail-closed。この新adapterは**未実装の本PR作成物**である。auth側のfresh reader interface/既存membership例外の受渡しhunkは同PRowner合意対象とし、新grantや追加owners:viewを勝手に要求しない。

## 3. closed write requestとcanonical command

新routeに限ってID/金額/版をcanonical decimal **string**とする。既存APIを一括文字列化しない。整数円のstorage意味を維持し、小数額や新料金規則を導入しない。

POST bodyは次の8キーを全て必須とし、未知キー/重複JSONキー/異型を400で拒否する。URL ownerも含め、missing/empty/0の意味を混ぜない。

|field|closed型・規則|C06分類|
|---|---|---|
|operation_key|lowercase hyphenated nonzero UUID。既存 `ParseOperationKey` を使用|Scope key。Content外|
|expected_account_revision|string `[1-9][0-9]*`、DB bigint内。成功で+1可能な範囲|Preconditionsのみ|
|received_amount|string正整数、1..9223372036854775807。最終balanceのoverflowも拒否|Content/Q|
|payment_method_id|string正整数、1..DB bigint最大。nameから推定不可|Content/Q|
|received_at|string、実授受時刻。UTC `YYYY-MM-DDTHH:mm:ss.ffffffZ` の厳密6桁microsecond。実在暦時刻、年0001..9999。offset/過剰精度/暗黙丸め/now補完は不受理|Content/Q|
|tendered_amount|string非負整数又は明示null。非nullはcanonical `0`又は正整数、bigint内|Content/Q。nullはcanonical sentinel `null`|
|change_given_amount|上記と同じ|Content/Q|
|deposit_consent|JSON boolean trueのみ。既決の明示了承を機械可読で受ける|Content/Qは`true`|

URL owner_idも正のcanonical decimal stringでDB bigint内。C06 Contentのキーを `owner_id,received_amount,payment_method_id,received_at,tendered_amount,change_given_amount,deposit_consent` の7個に固定する。期待版だけを `Preconditions["owner_deposit_account_revision"]` に入れる。actor/clinicは認証Scope、operation keyは専用key。依存確認用の新しいmethod版/fund版を捏造しない。

465採択契約との整合から、Qを使用する: `NewVerifiedCommand` のschemaは `owner-deposit.receive-new/v1`。Qは上記7業務fieldのみのcanonical typed payload。pure projectorが同じContentを生成する。Applyと506/465への入力は `req.VerifiedCommand` の再検証済みcopyからだけ導出し、handler DTO/pointerの別捕捉をしない。expected revisionはQに含めずreq.Preconditionsから読む。Qはrequest-memoryであり、再読後にcommandを再構築できる耐久保存と偽らない。

methodを465でlock後、安定system_keyがcashなら両現金field非null、tendered >= change、received=tendered-changeを検査する。非cashなら両nullを要求する。DBの「両null許容」を新現金受領の入力免除に使わない。金額誤記訂正での現金対不明null許可は今回の新受領へ転用しない。方法の種別が安定IDへ解決できなければvalidation。名称や現在active候補の欠落を現金へfallbackしない。

appointment_id/billing_id/pet_id/existing_fund_id/allocations/reason/自由metadata等はこの初回DTOに存在しない。送信された場合は未知キーとして拒否し、黙って落とさない。保存appointment/billing/source_entry/correction参照はNULL。これは将来の任意予約・余剰預り業務を禁止する仕様追加ではない。新しい金額上限、期限制限、追加承認や理由必須を新造しない。

## 4. closed success/read/error wire（新adapterだけ）

### 4.1 成功receipt

operation応答はC06 `NewResultResponse` の安全なfieldを基礎にし、**明示whitelistのowner receipt**を追加する。Outcome/Binding/EffectSummaryをそのままJSON公開しない。

成功必須field:
- `status`: POST初回 `succeeded` / POST同内容再送 `replayed` / GET保存成功 `succeeded`。
- `operation_key`:要求key。
- `result_refs`: 全値stringの固定object。`owner_id`, `deposit_entry_id`, `payment_fund_id`, `fund_consumption_event_id`, `actual_source_event_id`, `allocation_source_event_id`。
- `committed_versions`: `owner_deposit_account_revision` だけのstring map。fundに存在しないversionを返さない。
- `recorded_at`: C06が保存・再読したreceipt時刻。初回/再送で変えない。
- `receipt`: `schema="owner-deposit.receive-new/v1"`, `intent`, `clinic_id`, `owner_id`, `performed_by`, `received_amount`, `payment_method_id`, `payment_method_system_key_at_receipt`, `received_at`, `tendered_amount`, `change_given_amount`, `deposit_consent=true`, `account_created:boolean`, `account_revision_before`, `account_revision_after`, `balance_before`, `balance_after`, `actual_movement_amount`, `deposit_allocated_amount`, `fund_unallocated_amount="0"`, `billing_movement_amount="0"`, `inventory_movement="none"`。ID/金額/版はdecimal string、現金対はstring/null。intentとinventory_movementは固定値。

これらのreceipt値は同TX保存したC06 EffectSummaryの固定キー（内部string map）又はそのbindingで不変の元受領参照からのみ復元する。`inventory_movement` はEffectSummaryにも必須キーとしてstring `none` を保存し、成功whitelist receiptへその固定値だけを返す。欠損/別値を推測でnoneに補完しない。採択finance-operations technical-spec:116の副作用分類に従う明示であり、inventory participant・在庫write・在庫APIは追加しない。初回値を現在master/現在口座で再計算しない。snapshotのmethods system_keyは当時lock下で確認・保存した値と明示し、現在masterを歴史snapshotと偽らない。required field欠損/不整合はgeneric internal errorであり、不完全成功bodyを返さない。

時刻のexact wireを分ける:
- request.received_at / receipt.received_at / source.occurred_at: 厳密UTC、6桁小数付き `YYYY-MM-DDTHH:mm:ss.ffffffZ`。requestから検証した実時刻を保持する。
- operation応答のrecorded_at: 既存C06 `NewResultResponse` の `*time.Time` をそのまま標準JSON encoderで返す。Go `time.Time.MarshalJSON` のRFC3339Nano形式、UTCならZ、他offsetなら±HH:MM、小数は必要な桁のみ（0..9桁、末尾0省略）。本routeのために6桁UTCへ書き換えない。保存DB由来の精度はmicrosecondだがDTOの型の保証を勝手に狭めない。再送は保存された同じinstant/値を使用する。
- account GET.read_at: この新DTOではDB statement時刻をUTCへ表示変換し6桁小数で返す。口座snapshotの観測時刻でありreceipt commit時刻ではない。
- source digest用recorded_at: producerが同TXで一度取得したDB clock_timestampのmicrosecond値を明示保存し、UTC6桁へcanonical化する。C06 recorded_atと同じ時刻だとは主張しない。

`recorded_at`は保存記録時刻、`received_at`は実授受時刻。DB default `clock_timestamp` は厳密な物理commit瞬間ではない。既存deposit operationの `committed_at` もその列名だけで厳密commit timestampと説明しない。

### 4.2 口座GET

200の必須object: `clinic_id:string`, `owner_id:string`, `account_exists:boolean`, `revision:string`, `balance:string`, `read_at:UTC timestamp`。一覧/全履歴/請求残/ペット別残はこのreaderに追加しない。元受領詳細はoperation lookupの固定receiptから得る。

GETはwriteを含まないREPEATABLE READ/read-onlyの最外読取TXを新readerが所有し、fresh認可と口座projectionを同じsnapshotで検査する。既存C06 ExecuteをGETから呼ばず、口座初期化もしない。口座が存在すれば、同一statementでrevisionと全有効entryの符号付SUMを読む。read_atはそのprojection statement_timestampを使う。符号は現DDLと同じ `receive/release/correction_credit:+`、`apply/refund/correction_debit:-`。SUMはnumeric等のoverflowしない中間値で検査し、許容bigintへのcastを最後に行う。整合性としてrevision=entry数+1、balance非負・bigint範囲、必要source bridgeの存在を検査し、破損/由来欠落を0へ変えない。query失敗・RLS非表示・認可失敗を「空口座」と扱わない。

存在しない場合は、同医院ownerを現在認可・実在確認し、entryもないことを同一snapshotで確認できた場合だけ `account_exists=false, revision="1", balance="0"`。これは**未作成の論理初期状態**でありpersisted revision0ではない。GETで口座をINSERTしない。口座行だけ存在しentry0件の場合は `account_exists=true, revision="1", balance="0"` と区別する。

account GETはreceiptと別の現在projection。成功response後のGETが他の成功を含みrevisionが進むのは正常。古いreceiptのbalance_afterを書き換えない。GET失敗時は「受領成功、残高再読込だけ未完了」であり、再受領や新key作成を促さない。client接続契約はclinic/owner/keyの対応を保持し、遅着した別owner/古いrevisionの応答を現在画面へ当てない。

### 4.3 HTTPとエラー

C06 outcomeの場合は常に `status, operation_key` を返す。成功以外では result_refs/committed_versions/recorded_at/receipt を**省略**し、空の成功情報を出さない。以下は本routeの選定契約であり、506の一般wireや既存APIへ一般化しない。

|outcome|HTTP|clientの扱い|
|---|---:|---|
|POST succeeded / replayed|201 / 200|固定receiptを保持し、別GETで現在口座を読む|
|GET lookup succeeded|200|保存receipt。再実行しない|
|not-succeeded|200（lookup）|現helperがlock下で未成功と判断。自動再実行不可。人が前提確認後、同内容・同keyで再送可|
|operation-in-progress|202|§2.2どおり現在viewがある同actorだけ、同医院/key/元targetでlookup。create-onlyは同actorの現在view権限回復まで結果未確定のままkey/元targetを保持して保留。pending専用の閲覧例外なし。別actor/新key/再受領で救済しない|
|result-unavailable|503|sticky unknown又はC06保守的停止。自動補償/再受領/新key不可|
|operation-key-conflict|409|元target現在認可後、receipt非開示。成功済み内容の変更に使わない|
|version-conflict / dependency-changed|409|変更なし。現在口座/依存再読後、人が再確認|
|validation|400|C06内の確定業務拒否。generic messageのみ。事実と入力を確認|
|not-authorized / edit-right-lost|403|非開示。元targetを現要求ownerで置き換えない|
|scope-not-found|404|この正確なscopeに未検出。実授受失敗の証明ではない。自動新key不可|

C06前のJSON/UUID/字句不正は400、未認証401、非所属403。これらは別closed error object `{error_code:"invalid-request"|"authentication-required"|"not-authorized"}`、成功fieldなし。operation keyが構文不正ならechoしない。予期しないstore/reader/serialization障害は500 `{error_code:"internal-error"}`。存在確認前のDB障害を404へ変えない。500/接続断も成功不存在の証拠ではなく、既知の同keyで照会する。GET口座の不在ownerは認可済scope内の404 `{error_code:"target-not-found"}`、口座未作成は上記200。raw SQL/他医院owner/内部原因は返さない。

Rejectionは現C06がStatusだけを保存するので、実装者へ永続field-error詳細を仮定させない。固定enum以外の理由コードを後付けしない。C06 Lookupは最後の拒否詳細を再現するAPIではなく、not-succeeded/unknown等を返すものと明示する。

## 5. replay、新規受領、CAS

1. metadata-only claimを先に完了し、同key同内容は既存bindingを使う。HMAC key versionは保存版で比較。成功ありなら元対象の現在view後に固定receiptだけ返し、method active/CAS/new receipt validationを再実行しない。
2. 新規未成功だけApplyへ。現在create・元clinic/owner・465・account CASを保存直前に検査する。同keyでmethod/金額/時刻/owner/了承を変更するとContent conflict。期待口座revisionだけはfingerprint外であり、確定拒否後に人が再読・再確認した同内容の再送では同keyのまま更新できる。
3. 初回空口座はownerを保護した業務TX内でrevision1の口座を作成し、entry1件により2になる。口座を先に単独commitしない。既存口座も期待版一致が必須。
4. account lock下で現balanceを確認し、`balance_after=balance_before+received_amount`, `resulting_revision=expected_revision+1`, expected_entry_count=1。overflowは丸めずvalidation。実際のrevision incrementは既存entry triggerが行う。applicationが二重incrementしない。
5. 同account・別keyの2要求が期待版1なら、片方が2へ成功し、他方は先行中busy又は先行後version-conflict。自動で最新revisionへ置換して両方受領しない。同key競合はC06へ収束する。
6. 本単位は新fundをこのTXで初めて作るため、他TXが既に消費する既存fundのversion tokenは不要。fund.amount全額を同TXで預りへ配分しunallocated=0。将来の余剰預り/既存fund利用にこのCAS不要条件を流用しない。
7. 同keyでも別actorはC06上別操作。UUID一致だけで別actorのreceiptをreplayしない。実授受事実の不明をactor切替で再記録しない。

## 6. 最小DDL/model要求（446/466単一owner）

現存: owner_deposit_accounts/operations/entries、payment_funds、fund_consumption_events、finance_source_events、C06 binding/results、必要RLS/遅延制約。deposit operation/entry immutable triggerも **既存**（001:8736–8740）であり新規重複作成しない。

本節は必要なschema差分・型/NULL/制約/triggerとmodel/ERDの受入を定義するだけで、migrationファイルの配置・採番・生成・適用を指示しない。001は固定SHAの既存最終形を照合する根拠であり、**001へ無条件追記する指示ではない**。001統合後の新DDL配置/採番は最新migration規約と446/466単一ownerへ引き渡す。削除済み003–010を復活させない。実ownerが選定した配置、必要モデル差分、ERD/生成詳細ERDの受入一覧を供給成果に記録し、consumerが独自に別migrationを生成しない。

### D1: 元target anchor

新規 `owner_deposit_operation_targets` の契約:
- `clinic_id bigint NOT NULL`, `binding_id bigint NOT NULL`, `owner_id bigint NOT NULL CHECK>0`。
- PK `(clinic_id,binding_id)`、FK `(binding_id,clinic_id)`→operation_bindings `(id,clinic_id)`、同clinic RLS、INSERT/SELECTのみ、UPDATE/DELETE拒否。
- **claim時owner/account FKを付けない。** FKのKEY SHAREが共通gate前の業務lockになるため。現在のmanual_article_operation_targetsが同じ理由でtarget FKを持たない設計例（001:11409–11436）。ownerの実在/scopeはApply/lookup authで検証する。
- claim winnerだけが `CreateClaimMetadata(ctx,binding,req)` 内で、認証Scope・fingerprinted owner_idから作る。敗者は既存anchorを変更しない。callback失敗はbindingとanchor両方rollback。口座/entry/fund/audit/successはclaimで作らない。
- resolverはbindingID+clinic+owner_kind+intentをJOIN検証し、`LookupTarget{Kind:"owner-deposit-account",Ref:canonical owner_id}` を返す。anchor欠損/別family/不整合はfail-closed。現在URL/bodyから補修しない。

### D2: 台帳identityをC06 bindingへ一致

`owner_deposit_operations` に `operation_id bigint NOT NULL` を追加し、主keyを `(clinic_id,owner_id,operation_id)`、さらに `(clinic_id,operation_id)` UNIQUE、binding複合FKへ。既存 `(clinic,owner,operation_key)` のUNIQUE/PKを残さない。残すと別actor/intentが同UUIDを使う正当な別bindingが衝突する。元 `operation_key uuid` は保存metadataとして残しbinding.operation_keyとの一致を検査する。

`owner_deposit_entries` へ同じ `operation_id bigint NOT NULL`、operation FKとoperation ordinal uniqueを `(clinic_id,owner_id,operation_id[,ordinal])` へ変更。既存account revision unique・entry scope・source entry FKは維持する。entry_guard/operation_complete等の**全operation単位検索**をUUIDからbinding IDへ一括変更する。UUID比較だけ残す半接続は禁止。

`request_hash` は削除せず、locked bindingの保存済み `fingerprint` 32byteの**非権威コピー**に固定する。Bindingの公開structにはfingerprintがないため同TXのINSERT SELECT/厳密JOINで取得する。現在HMAC鍵で再計算してcopyしない。key versionの権威はbindingに残す。request_hashで別の再送判定をせず、CAS期待版をdigestへ混入しない。

owner-operation guardでclinic、anchor owner、operation key、owner_kind/intent、staff actor/performed_by、stored fingerprintを照合。entryのoperation_idは親と一致。成功receiptのbinding参照を遅延検証し、台帳だけ別operationでcommitさせない。操作/明細の既存不変triggerとRLSは保持する。

既存DBに実データがあるかは未確認。既存行のbindingが一意に照合できる根拠なしにseed/backfillを推測しない。immutable既存履歴を通常UPDATEする設計ではない。446/466が列/制約/trigger/ERD/model対応と旧値mapping可否を受領成果に記載し、適用計画・DB適用は別承認。個別464 DB PRを作らない。

### D3: 初回intent限定の完全性

receive-new成功について、同clinic/owner/bindingでoperation1件・receive entry1件・new fund1件・deposit_allocate1件が対応し、entry.amount=event.amount=fund.received_amount、event.source_deposit_entry_id=entry.id、entry/event/fund operation_id一致をdeferred検査する。entryのappointment/billing/source_entry/correctionはNULL。source event2件とC06 refsが実行の同じproducer集合を指すこともowner保存受入で検証する。

**一般payment_fundsへ UNIQUE(operation_id) を追加しない。** 複合方法の複数fundや、別operationの既存fund一部を預りにする将来業務を壊すため。全額/同binding fund条件はowner_kind+intentを確認した本familyに限定する。既存fundを使う他intentではeventとその預りentryの操作対応だけを保ち、元fund作成operationとの一致を強制しない。

## 7. 全lock順・same TX・gate供給subset

### 7.1 なぜ予約なしでもgateか

採択accounting technical-spec:228は「全writer gate／取得順へ従い」、かつ「元clinic／owner／資金で予約なし預りも保護」を要求する。reservations technical-spec:425の共通prefix、:430のowner/FK、:438のauth/policy、:440のleaf除外条件も該当する。新受領はowner/method/staff/clinicへのFKと権限に依存するので、appointment=NULLだけではleafにならない。共有helperが存在するだけを根拠に予約機能全体を依存へ足すわけではない。

### 7.2 現C06とgate APIの未供給接点

現 `Participant.Gate(ctx) error`（contract.go:201）に対し、`AcquireReservationGraphWriteGate(ctx) (context.Context,error)` は私有gate状態入りの新contextを返す（persistence/reservation_write_gate.go:87–112）。その戻りcontextを捨てる接続ではApplyのAssertが失敗する。また既存helperは `pg_advisory_xact_lock` でblocking、C06参加契約§1.2は業務gate/binding no-waitを要求する。**既存464が不具合動作しているのではなく、新consumer接続前に解消すべき適合差分。** 全C06 consumerへgraph gateを新規強制する契約ではない。

必要な限定供給:
- 489へoptional `ContextGate` 相当のhook `GateContext(ctx) (context.Context,error)` を追加。実装participantではこれを一度だけ呼び、旧Gateと二重呼出ししない。戻りctxのambient TXが元と同一であることをrunnerが検査し、binding読み/lookup auth/Apply/result writeへ渡す。既存Gateのみconsumerは互換維持。contextをparticipantの可変fieldへ隠して使い回さない。
- 467/persistenceへ **同じnamespace** `reservation-graph:write:v1` のtry-acquire variantを追加。既存first-lock検査・同TX state・Assertを保持し、`pg_try_advisory_xact_lock` falseをBusyへadapterする。既存blocking APIの意味は変更しない。新しいadvisory namespace/row versionを作らない。
- C06の短いclaim INSERTは現状ON CONFLICT待ちになり得る（store.go:33–108）。本契約も「全要求が絶対nonblocking」と約束しない。no-wait対象は業務gate/binding/明示domain lock。

### 7.3 本familyの具体的順序

|順|処理・lock|所有/制限|
|---:|---|---|
|0|認証・selected clinic・厳密request/Q生成。ここで方法active/口座CASを拒否しない|handler。成功replayを先行業務検査で潰さない|
|1|短いclaim TX: binding INSERT/既存読取 → winner anchor INSERT → commit|489/464。既存bindingのclinic FK KEY SHAREはこのmetadata TXに元から存在し、このTX終了で解放される。新anchorにowner/account FKを足さない。他TXをこの中から起動しない|
|2|最外業務TX開始直後、共通graph try gate。業務table read/lock、policy/staff/close等より前|489/467接点。取得不可Busy、owner effectsなし|
|3|binding `FOR UPDATE NOWAIT` → receipt/unknown再読 → 元target現在viewによる成功replay/停止判定|C06。ここで成功なら以降へ進まない|
|4|Apply savepoint → 同gate TXで現在staff/所属/clinic/create/anchor元owner scopeを再検査|464認可reader。キャッシュだけの判定禁止。readはこの段。独自policy advisoryを追加しない|
|5|465へ新fund methods集合 `[payment_method_id]`、ID昇順 `FOR UPDATE NOWAIT`。同clinic/nondeleted/stable key/activeと現金対検査|465。method rowをfund/accountより後に追加しない|
|6|clinic行を `FOR SHARE NOWAIT` で取得してis_activeを再読 → 認証staff行を `FOR KEY SHARE NOWAIT` → owner行 `FOR UPDATE NOWAIT`。同scope/非削除/有効性を再読|464＋clinic owner。method→clinic→staff→owner。clinic SHAREは非key active UPDATEとも競合するrow fence。既存LockActiveByIDのSHARE意味を保持しNOWAIT variantを同PR又は先行subsetで供給。今回は各1件。gateが上位なので他ownerの下位順を全域反転しない|
|7|account行 `FOR UPDATE NOWAIT`。なければ既に保護したowner下でINSERT revision1。同TXでactual revision/balance再読・期待版照合|464。同owner全participantの順と初期化競合をfixture化。重複作成を成功と取り違えない|
|8|deposit operation headerを保存（expected/result+1/count1/balance後）。receive entry ordinal1を保存。entry triggerのaccount FOR UPDATEは取得済み同rowの再入・revision更新|464/DDL。accountの二重increment禁止|
|9|506内部participant: new fund INSERT → deposit_allocate INSERT。fund_consumption_guardのfund FOR UPDATEは自TXで新規作成した同row。method/owner/accountを後から取得しない|506。既存fundを読む/消費する道はこのAPIにない|
|10|同TXの実受領source1件・非金銭allocation source1件 → fail-closed監査 → C06成功result INSERT/保存時刻再読|506 producer/464 audit/489 receipt。外部通知・ファイル生成なし|
|11|既存/追加のdeferred FK・complete・bridge・success receipt検査が全部成立して最外commit|失敗はowner write全rollback。結果不明はC06決着確認へ|

DBが行うFK KEY SHAREもlock graphの対象。上記でmethod/clinic/staff/owner/accountを先に保護し、bindingは段3、新fund/entry/eventは自TX作成済みを参照する。後から未知の既存source/billing/appointmentをロックする余地をDTOから除く。trigger内FOR UPDATEを消してgateだけに置換しない。clinic有効/無効・削除、policy変更/所属変更/owner削除等も採択graph gateに参加する必要があり、未参加の既存writerを「gateがあるから安全」と見なさない。

### 7.4 activationに必要な狭いwriter閉包

本sliceの切替時に必要な証跡は、**選択clinicのis_active有効/無効変更及びclinic削除**、対象ownerの通常変更/削除、staff/account/所属/permission policyの対象grant変更、method停止/削除と新参照、及び当口座の他writerである。予約/受付/全464全506完成は要求しない。採択正本のgate参加又は適法な既存切替拒否の証跡が未供給なら、新routeの有効化は停止する。本書は新しい運用停止switchの導入・作動を許可しない。

clinic側の必要subsetを明記する。採択[reservations technical-spec:437–440](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/design/domains/reservations/technical-spec.md#L437-L440)は医院設定・利用停止等の最終判定writerをgateへ参加させ、既存row lockを除去しない契約である。現[clinic_service_update.go:42–43](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/clinic/clinic_service_update.go#L42-L43)はIsActiveをis_activeへ更新map化し、[clinic_service.go:358–391](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/clinic/clinic_service.go#L358-L391)の所有TXから[clinic_repository.go:175–195](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/clinic/clinic_repository.go#L175-L195)のUPDATEへ到達する。この所有TX本文には明示的graph gate取得がなく、呼出し元/直接writerを含む閉包が既に完成したとは未確認。

必要な供給はclinic owner＋467の **active変更/削除の実物理TX先頭で共通gateへ参加するhunkと入口一覧、及びclinic read fence** に限る。直接管理writer/import等が同列へ触れるなら同じ参加又は既存適法な切替拒否を証跡化する。clinic全設定/全機能Done待ちは不要。既存 [LockActiveByID:81–95](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/clinic/clinic_repository.go#L81-L95) はambient TXを要求しFOR SHAREを使う実在fenceだがblockingなので、そのままC06 no-wait適合としない。464用NOWAIT variantは同じactive条件/SHARE強度の限定adapterとして供給し、既存blocking APIは変更しない。KEY SHAREをactiveのfenceと誤記しない。

activation証拠は入口→物理TX→gate→clinic UPDATE/DELETEの順、464側method→clinic SHARE NOWAIT→active再読の順、無効化先行なら新受領拒否/受領先行なら成立済みreceipt保持、両順の競合fixtureを含む。実行NOT_RUNはそのまま残す。これは既存active条件の競合保護を閉じる技術契約であり、新しいbusiness権限・医院停止ルールを追加しない。

現読取したauth permission_group_service_mutate、pet_owner_repository、payment_method_master_repository等の抜粋だけでは全入口のgate参加を証明できない。これらのsourceにhelperが見つからないことだけで全auth未対応と断定もしない。467/該当ownerから本閉包の入口→実物理TX→gate→下位lock→拒否/切替の対応を必要subsetとして受領する。未受領をruntime PASSとは扱わない。

## 8. 506内部producer、source、監査、保存結果

506へ渡す内部participantの入力は「認可済みclinic/staff、binding ID、Qから検証した新受領値、同TXで作成済みreceive entry ID」。返却はnew fund ID、deposit_allocate ID、実受領/非金銭source ID、固定元受領snapshot。外部HTTP呼出し/独立transaction/内部commitなし。要求entryを同clinic/owner/bindingで再検査し、method guard実施済みTX participationを検査する。guard実施をclient booleanで申告させない。

全額関係: new fund.received_amount=A、entry.receive.amount=A、event.deposit_allocate.amount=A、event.source_deposit_entry_id=entry.ID。新fundの `original_source_reference` はNULL。本列はimport由来の保存用であり同意/適当metadataを詰めない。

### producer source契約（509の設計は対象外）

同TXでfinance_source_eventsを2件だけ作る:
1. 実受領source: source_owner=`billing`、typed ref=fund_idだけ、source_version=1、actual_method_id=保存方法ID、actual_movement_amount=A、cash_movement_amount=当時安定key cashならA/他なら0、receivable_delta=0。
2. 非金銭預り配分source: source_owner=`billing`、typed ref=fund_consumption_event_idだけ、source_version=1、actual_method_id=NULL、actual/cash/receivable_deltaは全0。

deposit entryを第3の実受領eventにしない。第2eventのcontextにentry/account/binding/元fundを型付きdecimal stringで対応させ、預りの由来を追えるようにする。actual sourceのcontextも同じ関係と受領時method system_keyを保存。source_event_keyは各行server生成UUIDを一度だけ保存する。request UUIDを横流ししてactor別bindingで衝突させない。replayで再生成・再insertしない。各typed sourceの既存uniqueも維持する。

#### source digest/contextのexact schema

両sourceのcontext_snapshotは次の**全14キーが常に存在**するobject。追加キー/省略は禁止: `account_owner_id`, `operation_binding_id`, `owner_deposit_entry_id`, `payment_fund_id`, `deposit_allocation_event_id`, `payment_method_id`, `payment_method_system_key_at_receipt`, `received_amount`, `received_at`, `tendered_amount`, `change_given_amount`, `deposit_consent`, `account_revision_before`, `account_revision_after`。ID/amount/revisionはcanonical decimal string、received_atはUTC6桁、method keyは既存安定識別のASCII文字列、現金対はstring/null、consentはJSON true。actual/allocationとも同じcontextを保存する。現金対のnullもkeyは省略しない。任意自由text/name/院名/飼主名は含まない。

hash対象は次の**全22キーが常に存在**するobject:
`schema`, `clinic_id`, `source_owner`, `source_event_key`, `source_version`, `billing_change_event_id`, `fund_id`, `fund_consumption_event_id`, `deposit_entry_id`, `cash_adjustment_event_id`, `fund_record_correction_id`, `corrected_event_id`, `occurred_at`, `actual_method_id`, `actual_movement_amount`, `cash_movement_amount`, `receivable_delta`, `context_snapshot`, `recorded_at`, `recorded_by`, `reason`, `source_kind`。

- schema=`owner-deposit-source/v1`、source_owner=`billing`、source_version=`1`（string）。source_kindはactualで`actual-receipt`、allocationで`deposit-allocation`。source_kind/schemaはdigest schemaに属し、保存列の新設ではない。
- actualはfund_idだけ非null、他の5typed refは全てnull。allocationはfund_consumption_event_idだけ非null、他の5typed refは全てnull。corrected_event_idは両方null。実際のtyped ref列は6種類なので、非null以外のtyped refは5つである。
- actual_method_idはactualだけ方法ID string、allocationではnull。actual_movement_amount/cash_movement_amount/receivable_deltaは符号を含めcanonical integer string（本単位の値は非負又は0）。recorded_byは認証staff ID string。reasonは両方null（新理由必須ルールを足さない）。
- occurred_atは実授受時刻UTC6桁。recorded_atは段10で一度DB clock_timestampを取得して両source行に明示設定した同じmicrosecond値のUTC6桁。defaultに任せてhash後に時刻が変わる形を禁止する。source_event_keyは行ごとに保存したlowercase UUID。各行自身のbigserial idとcontent_digestはhash対象外。
- canonical JSONはUTF-8/BOMなし、各objectのキーをASCII byte昇順、区切りは`:`と`,`のみ、空白/改行なし。全key/string値は固定ASCIIまたは上述canonical ASCII字句なのでquote/backslash/control/非ASCIIを受理しない。全stringをASCII二重引用符で囲み、null/trueはJSON literal、数値JSON token・array・falseは本schemaにはない。nested contextにも同じ規則を適用する。sha256はこのexact byte列へ一度適用したraw32byteで保存する。hex文字列32byteとはしない。

このdomain限定encodingを他source producerへ一般化せず、実装PRでactual/cash、actual/noncash、allocationの固定byte列＋期待digest golden fixtureを同じ共有ownerで作る。これはsource内容照合用でありC06 HMACの代替ではない。未知値0補完はしない。締めの営業日、close、opening balance、ingestion、締め後差分は**ここで設計・書込しない**。既存consumerが未対応ならproducer引渡し未接続と明示する。

### mandatory audit

既存 `audit.TxLogger.LogEntryTx`（audit/service.go:29–31,201–217 / repository.go:40–50）が実在する。通常 `LogEntry/Create` はbase DBへ出るので使わない。新action/resource名は `owner_deposit.receive` / `owner_deposit_account`、resourceIDはowner ID。clinic/認証staff、operation key/binding、entry/fund/event refs、before/after revisionとbalance、実額/方法/実日時、明示了承、記録時刻を同TXに保存する。

required metadataは固定typed structで事前に厳密serialize/必須field確認し、失敗ならApplyを失敗させる。既存 `MarshalJSON` はbest-effortで失敗値を落とすため、必須金銭根拠を落として監査成功としない。監査writer失敗でfund/entry/source/receiptを全rollback。465 guardは別の金銭audit/receiptを作らない。

## 9. 所有・必要subsetと「1 PR」の境界

**実行単位: `464-R1 backend receipt-and-account-read`。これはscope識別名であり、新しいチケットIDではない。** 適切な既存実行票があれば利用し、親464のみをPR完了対象にして全AC完了へ自動移動させない。claim前に親/実行ID/範囲/独立AC/残ACを対応付ける。この文書の選定時点では実行IDとの対応は未成立であり、claim前の残件として管理する。

|owner|必要成果・責任|不要な全体待ち|
|---|---|---|
|464|本family API/DTO/Q/anchor reader/participant/口座CAS・ledger・固定receipt/current account reader/監査/fixture。新domain codeはbilling owner内。|全預りUI、適用/解除/返金/訂正完了|
|506|第8節の内部新fund・deposit_allocate・2source・固定受領snapshot。464と同一ambient TX、必要hunkを同PRに提供又は先行供給。|一般公開fund route、全配賦/refund/correction/CAS writer|
|465|既採択guard、method row lockと停止/削除参照保護、必要fixture。|全consumer接続・全7AC runtime PASS|
|489|既存C06/Q/claim hook、保存HMAC keyring注入契約、context-returning optional gate subset。|他domain consumer全完了|
|467/該当writer owner|同namespace try-gate/context/assertと第7.4節の本対象閉包の参加/切替証跡。clinic有効/無効/削除writerを明示的に含む。|予約/受付/全writerの製品全体完成|
|clinic owner|active変更/削除のgate参加hunk、実在SHARE fenceの464用NOWAIT adapter、対象入口と競合fixture。|clinic全設定/全機能Done|
|446/466|D1–D3のDDL/trigger/RLS/ERD・既存row mapping可否・model handoffを単一DB統合ownerで受領。|全DB backlog、DB適用・reset|
|API/routes/model単一編集owner|新3route定義・厳密DTO、composition root登録、storage modelのoperation_id/PK対応、生成対象変更の受入。共有hunkを競合編集しない。|旧Completeや全API変更|
|505|本初回は依存なし。将来の請求充当時に計算/未払subsetを別受領。|505 Done|

新route、error shape、family/grant mapping、D1–D3、gate adapter、source shapeは本書で選定済みである。供給担当はこの契約に対応する成果と編集範囲を受領し、未担当の共有差分を実装担当の裁量へ残さない。claim条件を次の2経路に分ける:
- 先行供給: 実際に先行完成が必要と合意したsubsetは供給commit SHA/型/受入契約を受領し、consumer PRに重複実装しない。
- 同PR供給: 464/506/465/489/467等の必要subsetを同じ実行PRで実装する場合、未実装でも構わない。開始前に具体的担当、変更path/hunk、型/lock/AC、統合順を合意し、そのPRのscopeへ明記する。先行merge済みでなければclaim不可とはしない。担当又は必要契約が未定の部分だけを開始前に解消する。
DDLは446/466単一統合の既定方針を維持し、機能PRへ勝手に同梱しない。DBの必要subset契約が固定すればコード準備は進められる。schema未供給の間は結合/有効化が未成立と明記し、全DB完成を待たない。

### R1の独立受入条件

- 指定のbackend範囲と供給subsetが閉じた1本のconsumer PRで、初回の受領保存→成功結果照会→同口座現在読取まで型・route・caller・同TX・必要監査が接続される。
- 共有API/model/routeの編集ownerとDB handoff commitが固定。506/465/489 library存在だけで接続済みにしない。
- strict request/immutable receipt/lookup/current readのcontract fixtureと下記拒否・競合fixtureが作成される。
- 残ACとしてFE、予約、既存資金余剰、他intent、consumer後続を明記する。backend限定成果を製品UI利用可能や464全体Doneと呼ばない。
- testコード/合成fixtureの準備と実行を分離。プロジェクトの後日mainローカル実行方針を維持し、未実行はNOT_RUN。必須CIの無効化や迂回をしない。実DB適用・deployは別承認。

## 10. 独立受入matrix（作成対象、実行NOT_RUN）

|case|期待する閉じ方|
|---|---|
|空口座受領→再読|論理revision1/0円から一回だけrevision2/A円、1fund/1entry/1allocate/2source/1監査/1成功receipt。同じ元owner。GETはwriteなし|
|既存口座B円→受領A|CAS一致でrevision+1/B+A。履歴を書換えず同口座current read|
|同key同内容並行・応答喪失|1回だけcommit、同refs/時刻/receipt、元owner認可。同key lookup、別key救済なし|
|同key別owner/金額/方法/時刻/同意|元anchorの現在view後conflict。別request targetによる認可代用なし。owner writeなし|
|CASだけ変更|確定拒否後、人の再確認で同key同Contentの期待版更新可。成功済みにはCAS再適用なし|
|別keyで同account/空account競合|busy又はversion-conflict。一方だけ期待版を満たす。初期化raceで孤立口座/片側成功なし|
|別actorが同UUID|独立bindingとして扱い、台帳UUID key衝突なし。別actorreceipt開示なし。回復目的でactor変更しない|
|停止と受領/削除と参照競合|method先行同row lock。停止先行なら拒否、受領先行なら成立fund保持。成功後停止replayは元receipt|
|clinic無効化/削除と受領|対象writerのgate先行、464のSHARE NOWAITとactive再読。無効化先行commitなら新受領拒否、受領先行commitなら元receipt/履歴保持。KEY SHAREだけでは非key更新防止にならないことをfixtureで区別|
|現在grant喪失/医院違い/owner違い|create/view matrixどおり、元target現在認可、部分データ漏洩なし|
|createのみ初回成功の応答喪失|現在view不足でlookup/replay非開示。同actorのview回復まで保留、別actor/別keyによる再受領なし。権限自動付与なし。回復後同keyで元receiptを取得|
|不正shape/0/負/小数/桁超過/現金差不一致|丸め/既定方法/now/NULL→0補完なし。金銭writeなし|
|graph/binding/method/owner/account lock busy|必要段でNOWAIT/try。C06 in-progress、attemptなし。claim UNIQUE待ちは別に測定|
|gate context/TX混入|元TXと異なるcontext、入れ子TX、下位lock後gate、独立audit/fund TXを拒否|
|anchor失敗/claim後crash|binding+anchor同時成立か両方なし。claim後にfund/口座がない状態でも元target lookup可|
|source/audit/receipt/deferred制約失敗|全owner effects rollback、成功receiptなし。unknown/commit不明はC06決着照会を保持|
|既存trigger・新binding bridge|revision二重加算なし、operation key/actor/owner不一致拒否、別binding event拒否。本intent全額条件を他intentへ適用しない|
|成功後GET失敗/遅着GET|受領成功を失敗にしない。key/clinic/ownerを保持し読取だけ再試行。古い結果で別口座を上書きしない|
|producer二重計上防止|実移動Aはfund source1件だけ。非金銭source0、請求残0、再送でsource増加なし。締め書込なし|
|副作用分類round-trip|初回success→同keyreplay→GET lookupの全receiptでinventory_movement=none。保存EffectSummaryも同値。欠損/別値を補完せず不整合検出。在庫participant/API/writeなし|

## 11. 選定状態・供給・利用開始条件

### 本書で選定済み

1. backend限定・予約なし・1新fund全額のR1境界と、3route/strict decimal-string/UTC microsecond wire。
2. family名、Q、create/view分岐、HTTP/error/固定receipt/current account read契約。
3. binding-key台帳接続・claim anchor・stored fingerprint copy・限定deferred bridgeのDB要求。
4. 489 optional context-returning hook、467同namespace try-gate、clinic SHARE NOWAITの限定adapter。既存blocking consumerとの互換を維持する。
5. 506内部同TX participant・2source producer・mandatory audit・readerの引渡し。

これらは技術契約の選定であり、464のhandler/participant/readerや各共有subsetの実装完了を示さない。EMR-506の一般公開route/一般wire、既存fund全writer CAS、EMR-509の締め設計は本書で選定しない。

### claim前に対応付ける事項

共有subsetごとに先行供給又は同PR担当/hunk/型/ACを明記し、親票と実行ID、部分scope、独立AC、残ACを対応付ける。本書の選定済み契約に基づき同PR担当が実装する成果は、未実装であることだけを理由に先行merge待ちへ戻さない。

### 結合・利用開始前に確認する事項

必要schema/model/ERDの供給、旧値mappingの可否、gate writer閉包と切替証跡、HMAC keyring/DI配線、現在認可、同TX保存・監査・receiptの接続、必要検証の受入を確認する。未成立の部分は結合/利用開始条件として残す。全親票・全DB・全clinic機能のDone待ちへ広げない。DB適用・deployは別承認。

実DBの現行行/適用状態、実RLS/遅延制約/並行実行、全auth/owner/clinic/master入口の実際のgate参加、画面の操作性、現在の有効HMAC配線は未確認。静的な契約選定から成功を推定せず、実行していない検証はNOT_RUNと記録する。
