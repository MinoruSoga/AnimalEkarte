# EMR-506 初回受領・方法訂正・lookupの採択技術契約

基準main `6e3a3a30a207b84bef334e2515373355ad659997`。mainマージは技術設計採用でありAPI/保存/helperの実在、製品接続/activation/受入済みとは別。進行状態はPlane。対象A=初回実受領＋明示配賦、B=方法のみ誤記訂正、両family lookup。fund/event/correction保存モデルと[金銭業務正本](../../../spec/domains/accounting/lifecycle.md)を再選定しない。505 exact/481 binding/462確認/508提示返答の設計採択と必要subset未受領を分ける。新本人業務判断待ちなし。

## A receiveとB correct_method

AはPOST `/api/v1/payment-funds`、intent=billing.fund.receive。一操作一受領、明示配賦だけ。route前段は認証/医院scope/strict decode、write-only middlewareでreplayを遮断せず初回/未成功Applyで現在accounting:create。closed request `{operation_key,owner_id,payment_method_id,received_amount,received_at,cash,allocations,posthumous_context,post_close_reason}`。clinic/actor body禁止、未知field/重複member拒否。ID/額は正数canonical decimal string/int64上限、cash金額のみ0可、先頭0/符号/空白/小数/指数/overflow拒否、新fund0不可。

cash必須key、非現金null、現金新受領は `{tendered_amount,change_given_amount}` 両額必須でnet=受領額。両不明null許可はB限定。received_at必須offset RFC3339・小数秒0–6桁、年0001–9999、leap second/-00:00拒否、UTC YYYY-MM-DDTHH:mm:ss.ffffffZ固定6桁へ正確変換。6桁超拒否で丸め/切捨てなし、同瞬間offset違いは同内容。now補完なし、recorded_at別、表示は既存医院zone。

allocationsは `[{billing_id,amount,expected_billing_version,expected_billing_settlement_token}]`。重複billing拒否、正額、checked加算又はbigintで総額overflow拒否、総額≤受領額。配列は意味順なし、ID数値昇順fingerprint/ordinal。古い請求日/同日作成順はUI初期案だけで送信後再配分なし。非現金5000/配賦3000/未配賦2000可、自動deposit/refundなし。

posthumous_context必須key、通常null、死亡後 `{reason,source_refs:[{kind,id,revision}]}`。参照は生前証明自体でなく正規readerで患者/元owner/billing/sourceを検査。allowlistは供給済通常診療subsetだけ、未供給は理由付き停止し他source fallbackなし。source revision/選択/理由はContent、期待版/life/source CASはPreconditions。A同TXはfund→allocate events(operation_id,ordinal)→audit→receipt、再読失敗を成功0残へしない。

BはPOST `/api/v1/payment-funds/{fund_id}/record-corrections`、intent=billing.fund.correct_method。closed request `{operation_key,expected_fund_token,payment_method_id,cash,reason,posthumous_context,post_close_reason}`。額/received_at/owner/original_source_reference入力禁止、有効値保持。mutationは現在accounting-post-close-edit:editを限定再利用、新grantなし。前段write一律要求なし、成功replayはview。既資金訂正へ465新受領active要求なし。

現金へ変更はcash object又は承認済両不明null。objectは両非NULL/各0以上、tendered-change=**最新effective受領額**（既往金額訂正含む）。非現金へ変更は明示cash:null、旧値はbefore_values/原fund/既訂正版へ不変保存。不明を0にせず受領額/最新effective実日時を保持しcorrection append、消費event追加0。日時変更は別後続intent billing.fund.correct_received_atとして独立契約、同route discriminatorへ混ぜない。金額/refund全familyの採択ではない。

## Q・bounded Content・厳密Preconditions

schemaはA billing-fund-receive-v1、B billing-fund-correct-method-v1。UTF-8 RFC8785 JCS、全ID/額/日時canonical string、allocation ID数値順、source_refsはkind ASCII→id数値→revision数値順、同(kind,id)重複拒否。必須key/明示null保持。reasonはUTF-8原値保持、空白のみ拒否でtrim改変なし。post_close_reasonはA/B必須key（通常null、既存締め後理由要求時非空）、死亡後理由への代用なし。

Qは業務値のみ。Aはrequestからoperation_keyとallocation expected版/tokenを除く全fields。Bはfund_id/解決済owner_id/payment_method_id/cash/reason/posthumous_context/post_close_reason。operation_keyはRequest.Key。Preconditions map[string]stringはA exact key billing_preconditionsのJCS array値 `[{billing_id,billing_version,billing_settlement_token}]`（ID数値順）、B exact key fund_tokenにexpected_fund_token。Applyでstrict typed復元しAのID集合はQ.allocationsと完全一致。Q/digestにCASを混ぜない。現contract.go:65–67/runner:492–510のContent64fields/256bytes制限をPreconditionsへ誤適用しない。

Content exact5keys `{schema,command_digest,owner_id,target_kind,target_id}`。digest=SHA256(canonical Q bytes) lowercase64hex、A target owner/owner_id、B fund/fund_id、全value≤256 bytes/map≤64fields。pure projectorでschema/strict Q検証し同5fields生成、NewVerifiedCommand(schema,Q,Content,projector)使用。Applyはreq.VerifiedCommand.Command()だけでDTO復元、captured DTO/raw payloadなし。長い理由/配列は短mapに詰めず既存JCS衝突耐性で全入力を束縛、claimは短anchorのみ。

## fund tokenとbilling settlement token

fund.version/第二counterを作らない。fund_token=`v1:`＋lowercase SHA256(UTF-8 JCS)。exact preimage:

```text
{schema:"billing-fund-state-v1",original:{id,clinic_id,owner_id},correction:null|{id,revision,corrected_method_id,corrected_received_at,corrected_received_amount,corrected_tendered_amount,corrected_change_given_amount},events:[{id,event_kind,amount,billing_id,source_allocation_id,actual_refund_method_id,source_deposit_entry_id,occurred_at,operation_id,ordinal}]}
```

original3IDは不変資金参照。correction最新revision又はnull、events全5種/全履歴ID数値順（deposit_allocate/release必須）。ID/額/revision/ordinal canonical strings、nullable必須null|string、日時UTC固定6桁。全bytesをhashし先頭page/activeのみ/残額だけ/correction単独/MAX(event)単独不可。ABAでもevent集合が変わる。不変性破損を補修tokenにしない。同fund row lock下で再生成/照合しreceive/allocation/release/reallocate/refund/correction/deposit全writerが共有、請求/口座CASは別。A新fundは事前token不要、B以降既fund mutation必須。readと成功receiptへcommitted token、未参加writerで有効化なし。

billings no-op UPDATEでは既存triggerは版を増やさない。金銭eventだけのため実値/updated_atを捏造しない。Aはexpected_billing_version＋expected_billing_settlement_token。settlement token=`v1:`＋SHA256(JCS) exact:

```text
{schema:"billing-settlement-state-v1",billing_id,clinic_id,owner_id,billing_version,claim_state,current_calculation_id,events:[{id,fund_id,event_kind,amount,source_allocation_id,operation_id,ordinal}]}
```

当該billingの全fund allocate/release全履歴をevent ID数値順、release元は同clinic/fund/billing allocateへ正規解決。deposit橋渡しはallocate/releaseで含むためdeposit_allocate/deposit_release/refundを二重加算しない。nullable source_allocation_id=null、他ID/額/版canonical string、claim_state enum。current_calculation_id/505現在請求額不明はA停止、0補完なし。請求版と別fund配賦両方を捕捉。全allocate/release/cancel/deposit-apply/reallocateが同graph/billing lockで最新計算/元owner/有効配賦を再読。新allocate/reallocate先は非取消/最新未払上限、release/cancelは取消元許可＋元allocation/未解除上限で、cancel+release同TXを禁止しない。fund tokenで請求CASを代用せず通常billing実値triggerは不変。

## TX・life/source・close参加順

共通prefixはmetadata短claim commit→最外physical TX→reservation-graph:write:v1固定namespace/key→binding NOWAIT→成功replay/unknown→Apply。gate前に業務row/close/policy/staff/advisory保持なし、未取得Busy、nested claim/TXなし。権限/close/method/source/pet/資金writerはgraph参加又は切替停止、旧未参加並走なし。採用上位gateの参加で新rankを発明しない。

A下位順：現在認可→選択billing ID昇順FOR UPDATE（状態/元owner/505現在計算/配賦再読）→received_atだけから正規close resolver実区間→由来close lock→owner/pet/life/source根拠再読→465 method参照lock/現在active→fund INSERT→events→settlement token再計算→manifest/audit/receipt。recorded_at/請求日をclose対象に追加しない。

B下位順：現在認可→元fund/sourceの保存済close/区分＋訂正差分対象resolver→由来close lock→元owner/pet/life/source再読→fund FOR UPDATE→token/最新effective訂正版→method存在/院/参照保護（active不要）→correction/D12差分participant→manifest/audit/receipt。最新effective received_atを保持し未訂正日時へ戻さない。Bはbilling段空、過去配賦billingのlock/計算/締め検査なし、fund全event CASは保持。

正規resolverはcash-register/business-rules.md:30/39/42/61/120とfinance-operations/technical-spec.md:145/149/150/265の実event→JST営業日AM/PM/EMG実区間、元source保存close/区分/訂正参照を使う。B未締sourceは最新effective実日時で解決。日付単位lockは区間由来の調整keyで、拒否/理由要求は当該実区間の締め状態だけ。日全体閉鎖・今日/記録日/無関係な請求日を根拠にしない。506/509必要subsetはtyped resolver、元close/差分期待版、正規source publication/late差分同TX participant。履歴/差分/理由/締め後認可を保持し方法誤記を実授受と区別。未供給/不明は当該操作だけ停止、架空日付補修なし。事前read対gate後source/区間集合変更は後取りせずwhole rollback→再試行。正規reader既存lock/owner participantを共用し独立TX/上位再取得なし。

462 Completeはgraph→receipt→既存close→FK/record/stay→billingを保持。506はambient TXの下位fund/eventsのみ、Runner/claim再実行なし。必要source実区間根拠はbilling取得前の既存close段で解決/受領し、下位から上位close追加取得や検査省略なし。不足は停止、nested Runner/別日付fallbackなし。既存refund/item/cancelはgraph→receipt（利用時）→billing→close→fundを保持し全ownerへclose→billingを強制しない。

死亡登録/訂正・主飼主移管/source訂正もgraph参加。元院/患者/元owner/source版/生前事実再検査、life reader未配線/障害fail closed。死亡日時を会計日補完せず491/source writerを506が実装しない。回収/Bへ新医師確認を追加せず、未請求確定は462 doctor-review/readiness/483 gateを維持。

## anchor・lookup・不変成功manifest

GET `/api/v1/accounting-operations/{intent}/{operation_key}`、intentは採択activation済family allowlistのみ。認証clinic/同actor/owner_kind=billing、request owner/target禁止。新billing_operation_targets closed `{binding_id,clinic_id,schema_version:1,owner_id,target_kind,target_id}`、PK(binding_id)、binding_id+clinic_id複合FK、business owner/fund/billing FKなし。ID正bigint、target CHECK owner|fund、A owner target_id=owner_id、B fund_id、schema=1。多対象の認可rootは元owner。A ownerはContent、B ownerは正規事前fund readerで解決しQ/Content束縛/gate後一致、取得不能claim前停止、B body owner変更なし。

A/B前段はshape/scope/元owner正規解決でwrite一律gateなし。既存binding再送は保存anchorでAuthorizeLookupへ到達、write失権でもview保持なら成功receipt。未成功Applyはwrite/close/source等・新受領だけactive必須。新candidateの不存在/不整合を正常anchor化しないが、成功anchorを今回candidate readで阻止/上書きしない。

既存CreateClaimMetadata(ctx,b,req)が短Content anchor fieldsだけから同claim TX INSERT。business writesなし、失敗binding/metadata同rollback、loser callbackなし、anchor上書きなし。変動billing/source集合はmetadataでなくmanifestへ。未成功でも元owner/rootとB fund同院元ownerを正規解決。ResolveLookupTargetはbinding+clinic保存anchor→LookupTarget{BindingID,ClinicID,Kind:"billing_owner",Ref:owner_id}。AuthorizeLookupは元owner同院/同actor/現在accounting:view、B fund欠落/owner不一致も非開示。URL/body/receiptだけ/現在主飼主fallbackなし。489 contract:218–231/runnerの既存hookを再利用し新hook/receipt storeを作らない。

A result_refs={fund_id,operation_id,fund_token,manifest_id}、committed_versions={}。B refs={fund_id,correction_id,operation_id,fund_token,manifest_id}、versions={correction_revision}。tokenはrefs、versionsはdecimal専用。複数請求/event配列をstring mapに詰めずoperation_id不変aggregate＋manifestへ。当時結果と現在残readerを分離。

新billing_financial_operation_manifestsはPK(id)、UNIQUE(binding_id)、clinic_id/schema_version=1/intent/fund_id/correction_id nullable/fund_token/committed_at/payload jsonb。payload exact:

```text
{events:[{event_id,ordinal,kind,billing_id,amount}],billings:[{billing_id,calculation_id,billing_version,billing_settlement_token,claim_state,allocated_amount,remaining_amount}],effective_fund:{correction_revision,received_amount,available_amount}}
```

nullableはevent.billing_id（非請求event）だけ、A calculation_id必須確定refで不明停止。B専用payloadはevents=[]/billings=[]/同effective_fundに固定、無関係505不足で停止しない。元fund correction_revision="0"、訂正版は正数、他canonical strings/enum。events ordinal順、billings ID数値順。INSERT-only/UPDATE DELETE禁止、binding/clinic/receipt同TX整合を制約/participantで強制し成功receiptなしmanifest commit不可。最終再読値をfreezeしてaudit/receipt同commit。B金銭移動/請求更新なし。lookupはmanifestだけから当時結果、manifest_id→binding/clinic一致検査。

## read・HTTP・責任境界・受領順

GET `/api/v1/payment-funds/{fund_id}` は現在accounting:view、original/effective/corrections/allocations/releases/actual_refunds/available_amount/fund_token分離。master現在名はcurrent label、保存system_keyと偽らず、返金必要/実返金/未配賦を同一視しない。請求未払は505最新額＋有効配賦で506再算定なし。

本family選定HTTP：初回成功201、replay/成功lookup200、processing202、記録unknown503、validation400、CAS/dependency/key-conflict409、現在認可不可403（結果なし）、同scope key不存在404。認可後result store障害generic500はunknown別code。source根拠不明validation400 source_unresolved、reader故障500、停止方法409 dependency_changed。共通全family HTTP変更ではない。書込/audit/receipt失敗全rollback、commit応答不明は既存C06再確認/同keylookup、unknownで新key受領/自動補償/再返金なし。外部決済送信なし。

505 cancel/rebill/計算版/元番号、506 fund release/reallocate/refund、464 deposit復元利用、462 Complete最外統合。cancel+release+必要deposit復元はowner一TX、再配賦は別明示操作、元snapshot/請求日保持。元資金不明は依存refund/復元だけ停止。refundは正額event/非空理由/実方法/元fundを分離しpending/cancelledをperformedにせずperformedだけactual_refund_event_id。方法別旧上限でなく元利用可能額。日時/金額/refund全familyは後続個別採択、同CAS/anchor/TXへ参加後接続。

月次未払の請求cutoff/実受領日/非入金調整、既決4列＋調整を保持。配賦/解除/預り/請求訂正を当月入金にせずevent一回、印刷/画面同scope-cutoff、新返金相殺/表示方針なし。

受領順：契約→489既存metadata/current auth＋446/466 anchor/manifest＋API単一owner型→506-A（465参照lock/505未払CAS/auditを同TX、旧Payment/Split閉鎖）→506-B（方法だけ、全消費writer CAS）→462通常Completeと生前source一種（T06/481/484/483必要subset、退院takeover別）。後続日時/release/reallocate/refund/505 cancel/464 depositは既存票、507/509へimmutable source IDs。全票Done待ち/保存モデル再設計なし。

受入定義NOT_RUN：部分複数配賦/同額ABA/資金競合/停止対受領両順/現金訂正両null/offset等値・6桁超拒否/死亡登録両順/source不明/owner変更/同key異内容・CASだけ更新/失権lookup/監査receipt rollback。実装・SQL/model/生成/API YAML/DB/実金銭/試験/build/browser/外部送信は行わない。

根拠は固定mainのbilling/routes.go:120–141、storage_finance.go、operation contract/runner/dto、[会計技術仕様](technical-spec.md)、[C06](../../shared/operation-participation.md)、[死亡後限定参加](../../../work/emr/emr-540-308-462-506-selected-handoff-20261022.md)、[旧候補と採択表](../../../work/emr/emr-506-fund-wire-design-candidates-20261022.md)。506コメント0def0d64-b336-4267-b2cb-d4a403c8c5a4、d4eea33f-53af-4142-8bed-3393b667fa1f、f3dd7d1f-4eb2-4e31-9677-7ce3fcc93f8fの境界を継承する。
