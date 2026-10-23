# EMR-482 施術価格 source → 未請求・会計：技術選定契約・統合全文

> 公開時の位置づけ（2026-10-21）：親レビューで選定した保存・wire・hash・税sourceの範囲を公開する。以下の提案／調査時点の記載はその経緯として保持し、runtime実装済み・全会計契約の実装着手可能を意味しない。現行基本運用は本人訂正により外税と確定。通常外税の閉じた経路は既決D10に従って先行可能で、一律停止しない。内税を含む値引／混在／有理税抜から整数内訳への算定subsetだけは未閉鎖で、選定まで当該subsetの有効化を禁止する。実装・試験・DB適用・マージはNOT_RUN／未実施。

2026-10-21。親選定用の統合全文。採択・公開・runtime実装済みとはしない。外部書込、実金銭操作、API実装、試験なし（NOT_RUN）。

## 1. 調査結果と範囲

固定 main `edf71515dbae11735a4184ff64c465e2ea51af88`、EMR-482 本文と全コメント1件、481 本文と全コメント1件を確認。482 のコメントは PR731 merge 通知で、採用 wire の確定ではない。T4 と A2 は保存資源であり、現 request/service/未請求 reader との runtime 適合は未完。

既決は、実在0円を残す、unknownを0へ補完しない、採用元価格と手修正を区別する、行割引を二重適用しない、取込後は差分を表示してスタッフ明示再適用、確定請求は直接変更しない、である。予約構成480、開始終了、実施 readiness484、医師確認462、金額了承483、取消・再請求全体を本票で再設計しない。481全domain完成を待たず、施術source専有部分と481の施術adapterだけを対象とする。

現 T4 は以下である。
- `trimming_service_lines` は同医院/detail/course又はoptionの安定identity。course XOR option。各masterは同detailで一行。削除して同じmasterで再採番しない。
- `trimming_detail_revisions` は detail単位の連続正revision。既存insert guardがdetailをFOR UPDATEしてmax+1を検査。current pointerは存在しないので、そのdetailの最大保存revisionをcurrentとする。
- `trimming_detail_revision_lines` は親revision/source_line複合FK、master_version、nullable bigintの adopted_master_unit_price/unit_price、ExactNumeric quantity/discount_rate、bigint discount_amount、unit|lineのdiscount_stage。最終DDLは割引率の上限100を撤去済み。初期CREATEだけを見て100上限を復活させない。
- 3表ともimmutable/RLSが既存。請求sourceは `billing_source_bindings`。旧billing_items個別FK、旧detail.course_id、中間option表を復活させない。

## 2. 保存単位と型

### 2.1 価格入力と算定結果を分離

新wireのID/版は正signed64のcanonical decimal string。quantityは正、discount_rateは非負のcanonical exact decimal string。unit_price/adopted_master_unit_priceは非負signed64整数文字列又はnull、discount_amountは非負signed64整数文字列。SQL bigint以上の精度を受け付けて丸め込まない。数量・率はExactNumericを通しbinary floatを通さない。新業務金額上限を追加しない。

`PriceInput` はclosed `{unit_price,quantity,discount_rate,discount_amount,discount_stage}`。新writerのdiscount_stageは必ず `line`。これはD06 lifecycle L119–123およびD10 L496–500の既決「各course/optionの値引きは明細scope」への技術mappingであり、新しい割引単位の業務選択ではない。T4のunit enumが存在するだけではunit算式を採択した根拠にならない。rateは同じ行scopeの割合条件合計、amountは入力された定額条件合計であって計算済み割引額でない。今回UIの単一率/定額をT4へ無損失保存し、505への変換は rate>0なら非負exactで上限100を持たないrate term、amount>0ならfixed termの最大2term（順序rate→fixed、scope=line）。合算率>100も既決の超過確認/0円下限の対象で、silent clampしない。150%は単一rate value="150"で原合算値を保持し、100+50等の架空元条件へ分解しない。旧個別入力adapterのmax100を共通保存schemaへ流用しない。505のrequested/applied/excess/allocated/net/taxをT4入力へ逆代入しない。unit_priceは割引前の今回単価であり、割引済み単価と未適用条件を併送しない。

既存 `unit` 行は意味を推測してlineへ変換しない。元値を閲覧できるがtyped importは `unsupported_discount_stage` として保留。明示した新line条件で改訂するか、元ownerの意味が確定した専用変換を受領するまで、過去の計算結果を再現済みとはしない。新writerがunitを生成しない技術選定で、業務の割引scopeを変更しない。

### 2.2 採用版と在籍

各revisionはその時点の実施source集合を全量保存する。current revisionに行がないことが今回の削除/失効で、stable source行のDELETEではない。再追加は同じsource_line_id。予約booking revisionの集合や現在master価格で過去の実施集合を再建しない。

`MasterRef` はclosed `{kind,id}`、kind=`course|option`。`AdoptedMaster` はclosed `{version,unit_price}`。版は正、価格はnullable。最初の新規採用/明示再採用だけ、scope付きmaster SHARE lock下で現在versionと表示済み価格を照合して保存する。既存採用を保持する時は現在masterの価格/版へ置換しない。非active masterの既存履歴は失わず、新規選択の既存active/適用規則は維持する。

unit_price=nullは未設定。保存は可能だが請求取込不可。0は既知の実在単価で取込可能。adopted_master_unit_price=nullでもスタッフが今回unit_priceを明示入力した事実は保存できる。これを当時のmaster価格が既知だったことにしない。旧sourceに採用版がない場合は `legacy_unversioned`。現在の明示採用として新revisionを作ることと、過去の履歴backfillは別である。

## 3. 施術側 API（追加候補）

prefix `/api/v1`。既存 `/trimmings/{id}` のidはappointment IDでありdetail IDへ意味変更しない。

- GET `/trimmings/{appointment_id}/price-source`：現在保存projectionと編集CAS。trimming:view。
- POST `/trimmings/{appointment_id}/price-revisions`：既存detailの実施source集合/価格入力を同時改訂。trimming:edit。予約作成・日時変更・施術開始終了を兼ねない。
- GET `/trimmings/{appointment_id}/price-operations/{operation_key}` と `/details`：元操作の結果/固定revision。trimming:view。

GET responseはclosed `{appointment_id,detail_id,appointment_version,detail_revision,source_digest,lines,provenance}`。detail_revisionは未作成時null。provenance=`versioned|legacy_unversioned`。linesはsource_line_id順のclosed `{source_line_id,master,adopted_master,input,tax_input}`。source_digestはschema-tagged全保存projectionのJCS UTF-8 SHA256 lowercase64hex。GETは採用/修復/改訂をしない。旧未接続構造だけからlinesを捏造しない。

POST bodyはclosed `{operation_key,lines,preconditions}`。linesは今回の全source集合で、各要素closed `{source_line_id,master,adoption,adopted_master,input,tax_adoption,tax_input}`。source_line_idは新規master初採用時のみnull、既存はcanonical ID。adoption=`retain|adopt_current`。retainは直前の同source採用値と完全一致、adopt_currentは指定master現在版/価格と完全一致。重複master/source、複数course、同ID別master、clinic/detail不一致、未知fieldは拒否。empty linesは施術記録の部分保存として許可し、請求可能な架空0円行を作らない。preconditionsはclosed `{appointment_version,detail_revision,source_digest}`。CAS-only前提と、入力に明示した採用元master版という業務内容を区別する。

新規source行採番、全revision lines、detail_values、audit、receiptは同TX。source集合の変更を旧SetOptionsと別TXに分けない。新source_line_idは下記の固定receipt details readerで返す。全入力が現保存値と同じなら成功receiptを保存し `changed=false`、同じcurrent revisionを返す。no-opで新revision/業務auditを増やさず、再送は同receipt。旧snapshotなしの新たな明示採用はno-op扱いしない。

`detail_values` はT4の閉じたschema=1 projectionとして `{schema_version,appointment_id,owner_id,pet_id,style_request,body_weight,bw_unit,body_temperature,used_shampoo,used_ribbon,remarks,style_image,completed_image}` を保存する。IDはnullableを現在model通り保持、本文/画像は保存済みstring、身体数値はnullable exact decimal、bw_unitは既存BodyWeightUnit enum又はnull。POST価格bodyから本文は更新しない。対応本文writerも同じT4保存関数で全projectionと価格行を同TX改訂するので、価格更新が旧本文へ巻き戻さない。

2026-10-23の[開始・終了契約](../../design/domains/trimming/trimming-start-end-contract.md#trimming-start-end-actual-performer)で、schema1を保持したままschema2（上記全key＋nullable `actual_performer_staff_id`）を選定。actor/予約doctorから推測backfillせず、本文/価格writerは同じT4全projection・source digestで実担当者を保持する。旧schema1専用writerは切替scopeで属性を落として保存しない。明示保存・同医院staff検査・reader/DTO/UI/共有関数の供給は未完であり、価格/税/会計取込全体を終了前提にしない。


通常mutable状態だけをこのPOSTで扱い、完了後/死亡後の生前訂正は既決専用訂正ownerから同保存関数へ渡す。通常PATCHの存在を専用訂正の許可にしない。専用訂正のroute全体を482で新設しないが、同じsource保存/監査関数を通す契約は必須。

## 4. 481施術adapter：reader と import / resolution

既存numeric DTO全体の破壊を避け、次のtyped経路を追加する候補。

- GET `/billing-items/trimming-source-candidates?pet_id={id}`：accounting:view。queryはpet_idだけ。結果closed `{pet_id,revision,items,warnings}`。revisionは安定source集合、current revision、適格状態、active claim集合、warningsのschema-tagged digest。196のcomplete時未請求更新検出にもこの構成要素を含める。
- POST `/accountings/{billing_id}/trimming-source-imports`：accounting:create（既存明細create入口のgrant）。
- POST `/accountings/{billing_id}/trimming-source-resolutions`：accounting:edit。
- GET `/accountings/{billing_id}/trimming-source-import-operations/{operation_key}` と `/details`、および `trimming-source-resolution-operations` 同形：accounting:view。

candidate itemはclosed `{source,appointment_id,detail_id,owner_id,pet_id,master,adopted_master,input,tax_input,accounting_context}`。sourceはclosed `{kind,line_id,detail_revision}`、kind=`trimming`固定。kindとidの組がidentityでありloop index/appointment IDだけで代用しない。accounting_contextは§5でshapeを閉じたnullable typed値。税根拠が未設定の実データはnullを保持する。warnings要素はclosed `{source_kind,code,count,blocking}`、source_kind=`trimming`、code=`legacy_unversioned|unknown_unit_price|unsupported_discount_stage|accounting_context_unavailable|source_scope_unresolved`、count正整数、blocking=true。価格unknownを候補から黙って落とさず警告に残す。非適格状態と既active claimは通常候補に含めないが、完了確認用revisionには含める。

candidateはcurrent revisionに存在するsource、医院/owner/pet一致、既存D06会計pull適格状態、未active claimを検査する。masterの現在active/価格>0を採用済行の存在条件にしない。開始終了ownerの新適格状態を推測しない。現accounting状態判定を利用する間も任意status PATCHを完成の代用として増設しない。

import bodyはclosed `{operation_key,source,accounting_context,preconditions}`、一操作一行。価格/割引値は受信せず、sourceの指定した保存revisionから読む。preconditionsはclosed `{billing_version,appointment_version,unbilled_revision}`。source.detail_revisionは採用する業務版なのでContentへ拘束する。適用時にそれが現在版でなくなればdependency-changed。新しい版を採用し直す時は明示新内容/新keyとし、同keyの採用版をすり替えない。

resolution bodyはclosed `{operation_key,billing_item_id,source,decision,preconditions}`。decision=`apply_source|keep_billing`。preconditionsはclosed `{billing_version,item_version,binding_version,appointment_version,unbilled_revision}`。apply_sourceは表示した差分の現在source入力を適用、keep_billingは現会計値を保持し差分を解決記録する。どちらも source.detail_revision をContentに含める。会計側手修正済み/unknownのmanual_overrideをfalseへ補完しない。

apply_sourceはbilling入力を元source入力に置換し、binding.applied_source_version/imported_valuesを今回sourceにし、manual_override=false、contract_versionを既存guardで進める。keep_billingはapplied_source_version/imported_valuesを変更せずmanual_override=true、resolutionに観測source版と結果binding版を記録する。keepは将来の新source変更まで永久黙認する指示ではない。missing/deleted sourceへのapply_sourceは拒否し、source削除を会計金額0へ変換しない。削除差分のkeepは既存会計を明示保持するだけで新請求を生成しない。

いずれも `billing_source_differences` と `billing_source_difference_resolutions` の既存historyを用いる。resolutionのPKは(clinic_id,difference_id)であり独立resolution_idは新設しない。GETで履歴INSERTをしない。submit TXで差分を観測しclosed changed_values `{schema_version,before_source,observed_source,billing_input}`（source不在はobserved_source=null）を保存し、選択を記録する。根拠不足の旧bindingを現在値から逆算しない。

取込後の施術変更は、会計が未編集draftでも自動同期しない。D06固有の明示再適用が一般3分岐の自動同期より具体的な選定である。draft手修正は差分確認付き明示選択、finalized/ever_finalizedは本経路で上書きせず既存訂正へ渡す。

## 5. 505入力・税adapter：master-backed の最小選定

### 5.1 根拠区分と選択

既決D10第五群1は「明細に対応するmaster由来の会計時点の税区分/率を保持」、税込入力契約は元入力額/入力税区分の保持を要求する。405採択文書の staff_confirmed は free-text treatment / manual other care の具体枝であり、master-backed 施術へ無条件に許可済みとは読まない。従って本案の技術選択は **trimming_courses / trimming_options の既存masterに同型のnullable税profileを追加し、明示入力値を継承する方式**。新税率・新税分類・新権限を決めず、course type等の非課金分類masterへ広げない。会計行に架空masterやstaff_confirmed根拠を作らない。

unknownの非推定・取込停止は405の既決保護とD10の根拠保持を施術adapterにも適用する技術選定。405固有422 errorはC06へ持ち込まず、本操作のvalidation=400/候補警告を使用する。0円でもtax unknownをexemptへ変更しない。

### 5.2 最小保存・master DTO

両masterへ `tax_type text NULL` と `tax_rate numeric NULL` を追加、defaultなし。CHECKは「両NULL」又は「type in included/excluded/exempt、rate非NULLかつ非負有限numeric、exemptならrate=0」。課税0%はincluded/excludedの0でありexemptと別。旧行をUPDATE/backfillしない。既存contract_version triggerが税変更も増分し、scope/RLS/既存master変更監査を維持する。税率の上限や医院の具体値を本案では作らない。

既存 GET `/masters/trimming-courses/{id}` / `/masters/trimming-options/{id}` に `tax_profile` と `contract_version` を追加。tax_profileはclosed `{status,type,rate}`、status=`known|unknown`。knownならtypeは上記enum、rateはcanonical exact decimal、unknownならtype/rate両null。JSON null profileと片fieldだけの更新は受け付けない。

既存POST/PATCHの既知outer fieldsに任意 `tax_profile` を追加する（nestedはclosed）。create未指定はunknown、PATCH未指定は保持。指定時はclosed `{type,rate}`のknown値を全体置換する。PATCH税変更時は追加field `expected_contract_version` 必須。master row lock内で現在 `master-trimming:edit` と医院/版を照合し、税pair・既存version増分・既存変更auditを同TX。createは既存master-trimming:create。既存他fieldsの許可集合をnested税2fieldだけへ狭めない。一般PATCHによるknown→unknown clearは追加しない。名前/画像/価格だけの変更で税を初期化しない。GETのmaster-trimming:view、その他既存grantを保持する。

新たな独立税履歴表/税操作registryは不要。masterの現値は編集可能、採用した履歴は下記T4/A2/505 immutable snapshotに固定する。master PATCHの結果断は元IDの再読取とCASで照合し、同内容の再送で版だけ増やさない。税更新だけに独立C06 familyや公開routeを新設しない。

### 5.3 入力価格の税基準と会計時点税を分離

元価格が税込か税抜かを後日master変更から逆推定しないため、T4 `trimming_detail_revision_lines.tax_input jsonb NULL` を追加する。NULLは既存契約以前のunknown。新writerは必ずclosed `{schema_version,status,type,rate,master}` を保存。schema_version=1。unknownならtype/rate/master=null、knownならtype/rateは上記profile、masterはclosed `{kind,id,version}`、kind=`trimming_course|trimming_option`。同sourceの実masterと一致することをinsert guardで検査。既存line immutable/RLSのままで、現在税cacheという第二正本を作らない。

§3のGET/POSTに列挙した `tax_input` とPOST `tax_adoption` は本節のshapeを使う。tax_adoption=`retain|adopt_current`。retainは直前同sourceのtax_inputを全値保持、adopt_currentは明示したprofileとmaster版を現在master SHARE lock下で照合して保存。新source初採用はadopt_current必須。profileがunknownならunknown保存でよく、価格保存・途中記録は妨げない。tax_inputのmaster版は価格のadopted_master.versionと別であり、税だけの再確認で採用元価格を現在master価格へ置換しない。POSTのline digestは追加2fieldも完全拘束する。元unit_priceはそのtax_input.typeで入力された額のまま保存し、includedを税抜換算した額へ置換しない。

税不明の旧行は後日の明示adopt_currentによる新revisionで既知化する。現在確認した事実であって、過去価格の税基準が最初から既知だったとはしない。通常mutable外は既存専用訂正の理由/履歴を通す。新sourceの既知税も価格手修正時に黙って消さずretainする。

### 5.4 accounting_context のclosed shapeと現在版

candidateのaccounting_contextはclosed `{tax_type,tax_rate,tax_category,is_insurance_applicable,provenance}` 又はnull。税根拠が揃った時のみobject。tax_typeは既存enum、tax_rateはcanonical exact decimal、tax_categoryはexemptならexempt、それ以外taxable。is_insurance_applicableは既存D10保険入力のbooleanをそのまま保持し、税profileから推定しない。現施術readerのfalseは既存挙動として維持し、本税adapterで保険対象を拡張・変更しない。個別保険確認の既存境界と新税根拠保存を混同せず、税改訂だけのために別の全保険adapter完成gateを追加しない。

provenanceはclosed `{kind,master,input_detail_revision}`、kind=`master`固定、masterはclosed `{kind,id,version,tax_type,tax_rate}`、kind=`trimming_course|trimming_option`。input_detail_revisionはsource.detail_revisionと一致。税sourceは必ずそのsource_lineのmasterであり、別courseの税を流用しない。serverが現master税実値/contract_versionを読み、このaccounting_contextを返す。新しい会計取込で税profileがunknownならnull。source.tax_inputがunknownでもnull。

source入力税のtype/rateと会計時点のmaster税値が異なる場合、元額を自動換算/再解釈せずcandidateをaccounting_context_unavailableとし、スタッフが価格入力の基準を確認して新tax_input revisionを保存してから取込する。master版だけが価格/名称変更で進み税実値が同一なら、元価格/入力税はそのまま、会計時点版をprovenanceへ別保存する。master inactiveだけで保存済価格行を消さず、同clinicの保持済master税実値を読める。現在masterが解決不能なら不明扱いではなく適切なscope/data integrity拒否にする。

import bodyのaccounting_contextは表示済み上記object全量。Contentのaccounting_context_digestは全fieldを拘束。Applyはmaster SHARE lock下で現在版/type/rateと完全照合し、違えばdependency-changed。現在master価格は読み直してunit_priceを再計算しない。入力/税profile/取込CASのどれかが変われば新たな表示確認へ戻し、同keyの業務内容を差し替えない。税profile writerは同rowのUPDATE lockを取得するので取込中のすり抜けを防ぐ。

取込後のsource再適用は既存会計の保存済accounting_contextを保持する。施術料金の再適用を税master再適用の承認へ広げない。新source.tax_inputの基準が保存済contextと異なる場合は本price/source resolutionで自動変換せず、既存会計の税訂正経路へ渡す。確定済みの税snapshotは常に不変。

### 5.5 原値と505算定結果のexact mapping

binding.imported_valuesはclosed `{schema_version,source,appointment_id,detail_id,master,adopted_master,input,tax_input,accounting_context}`。inputは未計算の元入力、tax_inputは価格入力時の基準、accounting_contextは今回採用した会計税根拠。初回取込/明示source再適用ごとの保存値を残し、mutable bindingの現imported_valuesだけを過去版の唯一保存先にしない。505 line.source_snapshot.original_projection={kind:"trimming_v1",values:このprojection全量} と計算用applied_pricingを分離して固定し、旧snapshot/監査beforeを残す。

T4 rate/amountは505 discount_conditionsのline scope入力に写し、requested/applied/excess/allocated/net/taxを元入力へ逆代入しない。会計全体割引は505 conditionsに一度だけ置く。再計算で既存飼主/キャンペーン自動割引を同じsource調整へ重ねず、別の明示会計条件は別scopeとして保持する。T4 unit_priceがincludedなら元税込額を保持した505入力として渡す。

master税2field・T4 tax_input・現在版照合・imported_values/505固定projectionが施術の必要adapter subset。405の入院source全体や481全domainの完了は待たない。価格writer/typed readerはunknown表示を備えて先行できるが、importの有効化はこの具体subsetが実装された後。

## 6. 現在認可、同TX、lock、全writer

price POSTはtrimming:edit、GET/lookupはtrimming:view。importはaccounting:create、resolutionはaccounting:edit、会計lookupはaccounting:view。既存sourceの割引を転記するだけで新discount:editを増やさず、実際の割引条件の新規編集/変更を行うwriterは既存discount permission規則の対象とする。単価だけ変更して派生割引額が変わる場合に割引変更grantを追加要求しない。具体的trimming専用grantを新設しない。Applyで現在grant・clinic/owner/pet・元対象・状態を再検証する。POSTはwrite middlewareのためview-only replayはGET。

既存489のnonbreaking ContextGate選定を再利用し、graph gate取得で返るcontextをbinding lock/Apply以下へ伝播する。Gate(errorのみ)へctxを捨ててassertを迂回しない。短いclaim TX→最外業務TX→共通 `reservation-graph:write:v1` gate→C06 binding→会計経路ではbilling rowと既存close境界→既存booking lockが必要ならそれ→appointment→detail→master SHARE（kind/id順）→source/billing item/binding（ID順）→履歴/監査/receipt。価格だけ経路はbilling/closeを取得しない。予約ownerのbooking→appointment順、既存会計ownerのbilling→close順を反転しない。共通graph gateの参加が欠けるwriterを残してこの順序だけで安全と主張しない。

importはitem・binding（manual_override=false）・505 snapshot/lines・billings.current_calculation_id・必須audit・receiptを同一TX。取込元には新計算結果を書き戻さない。releaseされていない既存bindingがあれば同key成功はreplay、別keyはdependency-changedとして二行目を作らない。最終DDLのuq_active_billing_trimming_sourceを最終防壁とする。unique違反を無条件成功へ変換しない。

resolutionはitem/保護状態・binding・difference/resolution・必要な新505 snapshot・audit・receiptを同TX。keep_billingで価格条件/結果が変わらなければ新calculation snapshotを作らず、差分選択とauditだけを残す。確定済請求のsource取込版/金額をsource改訂から変えない。

旧DELETE、取消、再請求、generic item create/update、complete一括作成、source専用訂正、appointment owner/pet/状態変更、施術本文保存/選択変更、master更新の各入口を列挙して参加させる。billing_items.source=trimmingの全create経路はこのadapterを使い、同TXbinding必須deferred検査を通す。非対応の旧個別FK payloadをそのまま受理する経路は有効化時に明示拒否し、旧値を安全なtyped identityへ一意変換できる場合だけ互換adapterを使う。loop indexから復元しない。

claim releaseは481/506所有の既存3field（released_at/operation_id/reason）を一括保存する専用処理だけ。billing item soft-deleteや親cancelledだけをreleaseとみなさない。解放後だけ再importでき、旧binding/itemは保持する。release未接続の間は当該旧削除/取消経路からclaimを黙って残す・解除するどちらも許さず、このsourceを扱う該当操作を閉じたエラーへ戻す。全domainのrelease完成は待たないが、施術の削除/取消adapter適合はcutover条件。

## 7. C06 Content / anchor / receipt / audit

owner_kindは価格=`trimming`、会計=`billing`。intentは `trimming.price_revision`、`billing.trimming_source_import`、`billing.trimming_source_resolution` の3つ。actor_kind=staff、clinic/actorは認証context。operation_keyはC06 canonical UUID。scopeからtarget IDを推測しない。

Content.Fieldsのexactキー：
- price: `{schema,appointment_id,lines_digest}`。
- import: `{schema,billing_id,source_line_id,source_revision,accounting_context_digest}`。
- resolution: `{schema,billing_id,billing_item_id,source_line_id,source_revision,decision}`。
schemaはそれぞれ `trimming-price-revision-v1` / `trimming-source-import-v1` / `trimming-source-resolution-v1`。digestは全closed commandの対応部分をJCS UTF-8 SHA256化。全string valueはC06 256byte以内。入力配列はmaster kind/id順へ正規化し、重複を除去して受理せず拒否する。lines_digestはsource_line_id/master/adoption/adopted_master/input/tax_adoption/tax_inputの全fieldを含む。Applyがimmutable commandからContentを再生成し照合後に書く。CASはPreconditionsに置くが、明示採用source版はContentから落とさない。

本節のimmutable typed commandは[採択済みQの共通契約](../../design/shared/operation-participation.md#verified-command-selected-20261021)に従い、Applyがreq.VerifiedCommandだけから再構成する。Qはrequest-memoryのみでowner-private耐久保管とは別、CAS/fenceはPreconditions、Contentは既定の短いstring mapを維持する。489のQ必要subsetは未実装依存であり、別command/DTO/pointerのcaptureで迂回せず、既存Content-only consumerは変更しない。


最小metadata新表候補 `trimming_billing_operation_targets(clinic_id,operation_id,appointment_id,billing_id)`。clinic_id/operation_idはNOT NULL正bigint、appointment_id/billing_idはnullable正bigint、PK(clinic_id,operation_id)、binding複合FK RESTRICT、price intentはappointment_idのみ、billing intentsはbilling_idのみというCHECK/insert guard。immutable、既存apply_rls_policy clinic scope、scoped repo。業務targetへのFKを張らず短いclaimで業務lockを先取りしない。489 ClaimMetadataWriterのwinnerだけ同claim TXで保存し、失敗はclaimもrollback。第二operation registry/別fingerprintは作らない。元target消失/失権/別target再送は元anchorでfail-closedし、最新request targetで認可を代替しない。

price成功result_refs `{appointment_id,detail_id,detail_revision_id}`、committed_versions `{detail,appointment}`。import成功result_refs `{billing_id,billing_item_id,calculation_id}`、committed_versions `{billing,item,binding,source,calculation}`。resolution成功result_refs `{billing_id,billing_item_id,difference_id,calculation_id}`、committed_versionsはimportと同形。calculation_idはkeepの場合現在保存ID。成功recorded_at必須。詳細は固定保存revision/binding履歴/監査から返し、current値を過去receiptへ代入しない。price no-opは専用detailsのchanged=false、それ以外true。financial_effect/inventory_effectはnone（請求額改訂と実金銭移動は別）。

現C06 ResultResponse直body、12status/HTTPは succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。非成功はstatus/operation_keyだけ。pre-C06構文拒否だけ既存ErrorResponse。unknown/busyで新key再発行・再取込・自動補償しない。成功receiptは後続source失効でも過去成功のまま。現在認可を満たすlookupで原結果を再取得する。

必須audit_logsは同TX、actorは認証staff、IP/UA新収集なし。
- action=`trimming_price_revision`、resource=`appointment_trimming_details`、resource_id=detail ID、old/new=`{revision_id,revision,source_digest}`（旧未作成は3値null）、metadata=`{schema_version,operation_id,operation_key,appointment_id,source_line_ids}`。
- action=`trimming_source_import`、resource=`billings`、resource_id=billing ID、old/new=`{billing_version,calculation_id}`、metadata=`{schema_version,operation_id,operation_key,billing_item_id,source_line_id,source_revision,binding_version}`。
- action=`trimming_source_resolution`、resource=`billings`、old/new同上、metadataはimport同形に`decision,difference_id`を追加。
必要な既表拡張として trimming_detail_revisions.operation_id nullable bigint と billing_source_difference_resolutions.operation_id nullable bigint / decision nullable text(apply_source|keep_billing)を追加する。既存行はnullを保持、新writerは必須、operation_idはclinic付きbinding FK RESTRICT、actor/key/intentの一致guard、新枝だけ成功receipt必須のdeferred検査を設ける。既存immutable/RLSは保持し、no-opは既存revisionのoperation_idを書き換えず今回receiptだけを追加する。独立resolution_idや第二履歴表は不要。新historyの成功operation参照はreceiptと同TXを検証し、audit失敗もrollback。業務no-opだけ新auditを作らない。Contentと監査の大きい本文を二重fingerprint正本にしない。

### 7.1 price成功detailsのclosed response

GET `/trimmings/{appointment_id}/price-operations/{operation_key}/details` はintent=`trimming.price_revision`固定、元C06 scope・保存anchor・現在trimming:viewを検査する。通常の結果GETとPOSTは既存ResultResponseのままとし、details/EffectSummaryを共通DTOへ追加しない。非成功は§7の既存status/HTTPとstatus/operation_keyのみ。成功時の専用response exact keysは `{status,operation_key,result_refs,committed_versions,recorded_at,changed,lines}`。statusはsucceeded又はreplayed、refs/versions/recorded_atは元成功receipt保存値、changedはboolean。linesは元receiptのdetail_revision_idが指すimmutable revisionの全行で、各要素exact `{source_line_id,master}`、master exact `{kind,id}`（kind=trimming_course|trimming_option、IDはcanonical正十進文字列）、master kind/id順。入力の新規source_line_id=nullも、request内で一意なmaster kind/idにより採番結果へ照合でき、現在revisionや入力順から推測しない。

price Applyは既存C06 Result.EffectSummaryのallowlistに `changed` を追加し、値を文字列 `true|false` として成功receiptと同TXで保存する。これは内部保存のみで、専用readerがbooleanへ検証変換する。no-opでも元current revisionのIDとchanged=falseを今回receiptへ固定し、後日currentの比較でchangedを再計算しない。固定revision・行集合・保存changedの欠落/不整合は現在認可後503 result-unavailableで止め、falseや空linesへ補完しない。新永続表/第二receiptは作らない。fixtureは新規masterのnull→ID対応、並行source追加、no-op後の別改訂でも元changed/lines不変、失権/別scope非開示、内部EffectSummaryが共通ResultResponseへ漏れないこと。NOT_RUN。

## 8. 最小実装sliceと未決の限定

A（482）：T4 writer、stable source/current projection、unknown/0、price C06/anchor/監査、旧writerの必要参加。予約作成の全面C06化や480全体は別。既存detail向け価格routeから開始可能。

B（481→482）：typed candidates/import/resolutionの施術subset、active claim/release参加、sameTX505接続。全domain binding/保険/医師確認を一度に実装しない。ただし取り扱うsourceと同会計を変更する必要writerを未参加のまま有効化しない。

税根拠の保存元/closed DTO/現在版/immutable snapshotは§5の施術master-backed subsetに閉じた。技術採択後の実装依存であり、実装者が税率やstaff_confirmed許可を推測する残件にしない。上記route/schemaは本案の候補で、現API実装済みではない。

後続受入候補：0/unknown、master改定後の採用価格不変、rate+fixed一度適用、削除再追加同source、候補→import保存→再読込一致、同key応答断、別key同source競合、source変更と手修正の両順、失権、確定後不変、release全fieldと再import、unknown停止、audit失敗全rollback。今回すべてNOT_RUN。

### 505共通型との統合・未閉鎖subset

505 TaxInput.master.kindには今回追加するtrimming_course/trimming_optionを明記する。現mainに税列が既存とは扱わず、482のmaster2列/T4 tax_input/版検査が実装された後だけwriter有効化。staff_confirmedは各ownerの採択済み入力事実を読むだけで、405のfree-text treatment/manual other care許可を施術へ転用しない。

505原projection保存はclosed `{kind,values}` のtrimming_v1枝で本書§5.5の全値を保持し、input/tax_input/accounting_contextを縮約しない。計算用PricingInputはclosed `{schema_version,input_unit_price,quantity,tax}`、schema_version=1。単価/数量はinputから、taxはaccounting_contextから写し、価格入力時tax_inputは原根拠に残す。TaxInputはclosed `{status,type,rate,origin,master,confirmed_by,confirmed_at}`でknown/master、masterは `{kind,id,version,updated_at,tax_type,tax_rate}`（updated_at=null、版あり）、確認actor/timeは元にないためnull。import actor/timeはoperation監査に保持する。元値や税込額を計算結果へ上書きしない。

税込有理税抜基準を整数の割引基準/配賦基準/net/subtotalへ対応させる505算定契約は未閉鎖。ExactRatio型を定義しただけで完了とはしない。既決元税込額保持を変える新丸め/算式を実装者が発明せず、505の段階別対応/合成fixture選定までは内税を含む値引／混在／有理税抜から整数内訳への当該import/再適用の算定subsetを無効化する。通常外税の閉じた経路まで一律停止しない。価格writer/read/unknown保存、税根拠snapshot、通常外税の閉じた経路は別sliceとして先行可能。この残りは試験NOT_RUNとは別の設計残で、税率の業務決定や481全domain完了待ちではない。

## 9. 固定根拠

- [T4/A2最終DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/migrations/001_init.sql#L8087-L8230)、[ExactNumeric変更](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/migrations/001_init.sql#L9394-L9396)、[active claim](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/migrations/001_init.sql#L9481-L9494)、[最終binding guard](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/migrations/001_init.sql#L11017-L11031)。
- [T4 model](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/model/storage_reservations.go)、[現detail model](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/model/trimming.go)。
- [未接続request](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/trimming/trimming_request.go)、[現update/lock](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/trimming/trimming_service_update.go)、[旧未請求reader](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/billing/billing_item_repository_unbilled.go)。
- [既存billing DTO](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/billing/billing_item_request.go)、[現billing grants](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/billing/routes.go)、[現trimming grants](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/trimming/routes.go)。
- [D06価格・再適用](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/design/domains/trimming/technical-spec.md)、[D10 adapter](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/design/domains/accounting/technical-spec.md#accounting-trimming-adjustment-adapter)、[現C06](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/contract.go)。

- [D10 master由来の会計時点税・計算順](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/spec/domains/accounting/lifecycle.md#L78-L85)、[元税込入力保持](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/spec/domains/accounting/lifecycle.md#L623-L631)、[405採択税入力と限定staff_confirmed枝](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/work/emr/emr-405-tax-input-contract-20261020.md#3-入院入力契約-v1)、[D06行scope](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/spec/domains/trimming/lifecycle.md#L119-L123)。

## 10. 関連する原projection adapterの追補

入院sourceの405既存 `source_snapshot.pricing` と今回の505版付き原値保存形の対応は、[483のadapter追補](emr-483-price-acknowledgement-selected-20261021.md#hospitalization-405-pricing-adapter)に従う。施術は本書§5.5のtrimming_v1全量を使い、入院枝へ置換しない。旧405文書は変更せず、原値全量保存・元字句とnullable／field presenceの保持、applied_pricingとの分離を維持する。

### 2026-10-21本人訂正：外税の基本運用と未閉鎖範囲

本人 Sentinel_9d2e828eb9308191be6b57221ddd8f02「間違えてました、外税ですね」により、現行の基本運用は外税と確定する。通常外税の閉じたpreview／submit／import経路は既決D10に従って先行可能であり、内税edgeの未閉鎖を全会計の停止条件へ広げない。

未閉鎖は内税を含む値引／混在／有理税抜から整数内訳への算定subsetに限定する。既存の「入力した税込額を維持する」要件は撤回しない。親確認では会計フロー本文に税／値引式はなく、旧動画では2行外税100%値引→0円だけを確認したため、この資料から内税edgeの式を選定したとは扱わない。新しい数式・丸め・税率は追加せず、段階別対応の選定責任と既決の有効化条件を維持する。
