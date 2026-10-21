# EMR-481 一般 source binding・同期・差異解決：技術選定契約・統合全文

2026-10-21。独立レビュー補正と親採択を反映した選定全文。既決業務を変更せず接続方式を固定する。設計の採択と、公開・runtime実装・有効化・受入完了を区別する。外部書込、実装、試験なし（NOT_RUN）。

## 1. 根拠と結論

静的基準 main `0de244ba69f9287a0bacff6abeddabb50a712aac`。EMR-481 最新本文（02:48:11 UTC更新）と全コメント1件、D10 lifecycle/technical-spec、最終001・storage_finance/treatment model、billing_item_unbilled.go/update、518保存契約、484 writer契約を照合した。482/483の選定全文は同日ローカル親レビュー稿で、最新mainへ公開済みと混同しない。

一般共通bindingにsource別adapterを登録する。最初は518の明示 charge_source linkで証明できる treatment 一行の縦割り。trimmingは482の明示再適用policy、hospitalizationは405の別event正系統を維持する。vaccination/examinationの実施結果があるだけでは請求source完成としない。元不明を新規手入力や1版へ変換しない。

現 aggregateUnbilled は現在master税・当日campaignを読む旧projectionで、固定原値として採用できない。treatmentToUnbilledBillingItemは旧source列を参照し、discount_rateを無視する経路が残る。これらをmodel移行済みという理由で正系統扱いしない。

## 2. 既決の優先順位

1. 未確定・未手修正の一般プラン行は自動同期、手修正行は保護＋差異、確定済みは専用訂正。manual_override=NULLは不明でfalseにしない。
2. master更新だけでは同期しない。予定への新規採用・明示再採用で固定した値を読む。元税込額を維持し、浮動小数・割引二重計上を避ける。
3. 薬剤/処置追加削除・数量の臨床差分は462再確認。単価/割引だけは483の警告了承、医師再確認なし。手修正を保持した金額差分だけで全差異解消を新しい確定条件にしない。
4. 482 trimmingは取込後、未手修正でも明示再適用。一般autoへ巻き込まない。
5. 484のcurrent generation、明示purpose、performed/preparedを再利用。execution_source/result linkはcharge_sourceの代用にならない。会計からプランを逆更新しない。入金・会計確定・実施・証明書・在庫を自動起動しない。

## 3. 型・正系統 adapter

新wire objectはclosed。ID/版は正signed64 canonical decimal string、operation_keyはC06 UUID。金額/数量/率はexact decimal string/nullを原保存に従い保持し、0とunknownを区別する。nullable値を0/false/空配列へ補完しない。

`SourceRef` はclosed union：treatment `{kind,treatment_id}`、vaccination `{kind,vaccination_id}`、examination `{kind,examination_id}`、trimming `{kind,trimming_source_line_id}`。hospitalizationは405採択の `{kind,care_plan_item_id,hospitalization_id,event_id,event_revision}` を保存extension実装後だけ登録。treatment_plan/plan_itemを新binding kindにしない。manualはsource bindingを持たない既存手入力で、本APIのsourceにはしない。

`SourceStamp={source,version,digest}`。versionは各ownerの本当の請求根拠版。digestはschema-tagged canonical projectionのSHA256 lowercase64hex。treatmentは treatments.contract_version、trimmingはdetail_revision、hospitalizationはevent.revision。vaccination/examinationは各ownerの固定原値writer/真の版を受領するまで未登録。Version、contract_version、fulfillment_generation、plan revision、doctor clinical generationを同じ数と見做さない。

内部interfaceは `LoadCurrentForBilling(ctx, SourceRef)`、`LoadAppliedProjection(binding)`、`ClassifyChange(before,after)`、`CheckReadiness(ctx,source)`、`SyncPolicy()`。戻り値はtyped unionで、任意JSONと型推定を許さない。policyはtreatment=`auto_unmodified`、trimming=`explicit_reapply`。他kindへdefault autoを設けない。失敗は unknown/unsupported/absent を区別し、通信失敗を削除や0円としない。

### 3.1 初回 treatment subset：plan-backed charge projection

必要条件は同clinic/record/pet、現在activeの一意 charge_source link(owner_kind=treatment)、現在plan revision、実在treatment。別planへの同役割共有なし、result linkだけでは不可。master identity/quantity/単価/税/保険は518の現在採用済みplan snapshotと元projectionから取る。現在masterを再読取して過去税/価格を補わない。treatmentの既存明細割引条件は、その現在版の入力値を割合→定額のline termsへ写す。discount_rateの旧単位が証明不能なら停止し、0又は百分率へ推定しない。

planの請求関連変更を受ける coordinator participant は、同TXで対応treatmentの請求投影（typed master、quantity、unit_price等）を一致させcontract_versionを進める。planが変更されたのにtreatment版だけ据置で同じSourceStampとしない。税/保険等treatment列にない値だけ変わる場合も、下記請求projectionを同TX更新する。既存storage_contract_version_guardはNEW/OLDからcontract_versionとupdated_atのみ除いた全row差分を比較するため、このJSON列変更も自動増分対象となる。contract_versionを直接増分指定しない。下記最小列を選ぶ。

`treatments.billing_source_projection jsonb NULL` を追加する。新writerはclosed `{schema_version,plan_item_id,plan_revision_id,plan_revision,charge_link_id,master,pricing,insurance,discount_terms,bill_discount_exclusion}` を保存。schema_version=1、masterは `{kind,id}`、kind=consultation/procedure/medicine、pricingは483 PricingInputと同じ入力単価/数量/TaxInput、insuranceは `{is_insurance_applicable,provenance}`（provenance=explicit|legacy_unset）、discount_termsは483 DiscountTerms。unknownはnullable原投影として保存するが計算には使えない。projectionの値は上記固定plan/treatment入力と下記明示採用した全体割引除外からのみ作り、client任意JSONを受け付けない。既存rowへbackfillなし。更新は既存contract_version triggerで増分、履歴原値はplan revisionおよびbinding/difference/calculation snapshotに保存する。新しい予定基盤・財務正本・税率設定表ではない。

plan snapshotのmaster税originが485/405のstaff_confirmedを借用しないこと、正規master版/採用原値の対応を検証する。masterなしother treatment、plan linkなしlegacy treatmentはこの初回登録範囲外で `source_scope_unresolved`。自由記載治療の採用済み405税確認契約を外来へ無断転用しない。これらが残ってもplan-backed縦割りの開発開始は妨げないが、対象患者の未請求集合から黙って消さない。

### 3.2 全体割引除外の採用値と由来

`bill_discount_exclusion` はclosed `{status,excluded,setting_state,master,adopted_at}`。status=`known|unknown`、excluded=boolean又はnull、setting_state=`explicit|missing|unresolved`、master=null又はclosed `{kind,id,version,source_digest}`、adopted_at=server RFC3339又はnull。master.kindは同projectionのconsultation/procedure/medicine、idは同元品目、versionは真のmaster版又はnull、source_digestは原設定を含むcanonical source projectionのSHA256。真の版がない時に1や時刻を代入しない。

knownはmaster/adopted_at非null、setting_state=explicitならexcludedは保存された明示boolean、setting_state=missingならexcluded=false。missingは正しい元masterの読取に成功し当該設定が未設定と確認した場合だけで、業務正本の設定欠落時の既決適用値である。列未接続・取得失敗・元master不明はunknown、excluded/master/adopted_at=null、setting_state=unresolvedとし、missingへ置き換えない。初回採用writerは同じ元masterの価格採用と整合するscope/版で原設定状態を凍結する。既存source snapshotに既にこの値/根拠があればその原値を再利用する。518が現在このfieldを持つと断定せず、518採用writer→481projectionへの必要な追加subsetとする。

採用時又はスタッフが対象を指定した明示再採用時だけ読み、追加後のmaster更新を同期sourceにしない。imported_valuesのprojectionおよび483/505 treatment_v1へ全量保持し、505 allocation.lines.excludedへknown.excludedを渡す。後日の現master読取で補修しない。unknownなら当該算定だけ停止する。設定欠落falseの適用で既存確定会計のsnapshot/配賦/帳票を再計算しない。除外を保険対象外・非課税・明細割引禁止に変換しない。

### 3.3 imported_values の閉じた treatment_v1

binding.imported_values=`{schema_version,source,source_version,projection,treatment_original}`。projectionは§3.1全量。treatment_originalはclosed `{item_type,consultation_id,procedure_id,medicine_id,unit_price,quantity,discount_rate,discount_amount,is_insurance,contract_version}`。nullable原値・元decimal字句を保存する。治療本文/memo等請求に不要な情報はコピーしない。元fieldを削って483 pricing_v1へ押し込まず、483 OriginalProjectionへ `treatment_v1` の閉じた枝を追加する。同じ原projectionから505の計算用PricingInput/DiscountTermsを導出し、一度だけ計算する。current master/campaignで追加割引を再探索しない。別採用済みcampaign条件が必要な行は保存済みその条件を引き継ぐ専用subsetまで保留する。

## 4. API：初回一行単位

prefix `/api/v1`。下記は採択する追加契約で、既存route実装の説明ではない。

- GET `/billing-items/source-candidates?pet_id={id}&kind=treatment`：accounting:view。結果 `{pet_id,kind,revision,items,warnings}`。items=`{source_stamp,medical_record_id,plan_item_id,projection,readiness}`。readinessは484の現世代/用途/有効eventの固定projection。revisionは全対象source/claim/readiness/unknown警告のschema-tagged digestで196へ接続。warnings=`{source,code,blocking}`、code=source_scope_unresolved|source_unversioned|pricing_unknown|discount_semantics_unknown|readiness_unknown|adapter_unavailable。非適格sourceは候補にしないがrevisionには含める。
- POST `/accountings/{billing_id}/source-imports/preview`：read-only、body `{source_stamp}`。現在scopeと候補を検証し計算前後/必要了承を返す。
- POST `/accountings/{billing_id}/source-imports`：accounting:create、body `{operation_key,source_stamp,acknowledgement,preconditions}`。
- GET `/accountings/{billing_id}/source-differences`：accounting:view、純read。現在sourceと適用値を比較し、既存履歴+未保存current差分を表示。GETで履歴作成/同期しない。
- POST `/accountings/{billing_id}/source-resolutions/preview`：body `{billing_item_id,source_stamp,decision}`、read-only。
- POST `/accountings/{billing_id}/source-resolutions`：accounting:edit、body `{operation_key,billing_item_id,source_stamp,decision,acknowledgement,preconditions}`。decision=apply_source|keep_billing。sourceを直す選択は元owner編集画面へ遷移し、このrouteに第三writerを作らない。
- GET `/accountings/{billing_id}/source-import-operations/{operation_key}`、`source-resolution-operations` の各lookupと `/details`。accounting:view、private/no-store。

preconditions=`{billing_version,item_version,binding_version,unbilled_revision,line_versions,doctor_review_set_id,records}`。importはitem_version/binding_version=null、resolutionは正値。line_versionsは算定に参加する全現在行の `{billing_item_id,contract_version}` 昇順集合。recordsは462の現在 `{medical_record_id,clinical_generation,clinical_digest}` 完全集合。現在世代/digest unknownを確認済みへ変換しない。source_stampは採用する業務内容なのでContentへ含め、単なるCASとして除外しない。

acknowledgementはnull又は483のclosed Ack command（before/after revision/digest・影響行集合・accepted=true・超過了承）。必要な時にnullなら保存しない。preview responseは `{billing_id,source_stamp,decision,before,after,affected_item_ids,preconditions,acknowledgement_required,has_changes}`。before/after型は483をそのまま用いる。preview時点のapprovalを自動生成しない。

## 5. 同期3分岐と警告了承の境界

source ownerの保存に参加する `BillingSourceSyncParticipant` は、共通gate後に対象active bindingを発見して同一最外TXで検査する。全行転記を人へ要求せず、適用する値はsource projectionから自動生成する。

- draft、manual=false、policy=auto、必要な財務根拠既知、既存了承が不要なら、行入力/binding/505 snapshot/current参照/必要462変化/監査を同TX更新する。
- draft manual=true又はNULL、又はpolicy=explicit_reapplyなら既存会計値を守り、差異のみ保存する。NULL由来不明をfalseへ修復しない。
- 「一度でも確定済み」は業務上の意味で、現在claim_stateと確定eventを用いた既選定判定に従う。退役したever_finalized/status列を復活させない。現在draftへ戻っただけで一般autoへ戻さず、行/binding適用値を変更せず差異表示から既存理由付き訂正へ渡す。
- 確認後の金額だけ変更は483の了承を同TXで保存してから会計へ適用する。source操作のpreviewで会計ごとのbefore/afterを示し、同じtyped commandに了承を含める。確認前には新たな了承gateを設けない。単純なauto同期に人手で金額を再入力させない。

採択する限定待機方式：未確定・未手修正・auto_unmodifiedの行は原則自動同期する。医師確認後の金額だけの変更で、既決の警告了承が必要かつその有効な了承をまだ取得していない場合に限り、sourceの通常保存と適用待ち差異を同一TXで保存し、会計入力・適用source版・計算snapshotを変更しない。差異又は監査の保存に失敗した場合はその操作全体をrollbackする。通常保存の無条件成功を保証する意味ではない。

会計には `awaiting_price_acknowledgement` として未適用を明示し、現在source版・before/after・影響行・CASを再提示して、明示了承後に自動導出値を適用する。金額の再入力・手転記を要求せず、manual_overrideを変更しない。古いsourceのまま確定/精算は不可。確認前・必要了承なし・既存有効了承ありへこの待機を拡大しない。手修正保持の単なる金額差異を全件解消必須にしない。待機はdurable差異＋現binding/source比較から導出し、第二の了承registryや事後の形式了承を作らない。

source保存直前の医師確認状態・変更分類を共通gate下で検査し、競合で確認参照が変わったことを了承不要の根拠にしない。価格未適用だけでは医師確認receiptを失効させず、実際の臨床差分がある場合だけ462世代/gateが変わる。sourceが後続了承までにさらに変わった時は旧afterを適用せず再previewを要求する。UI/元owner receipt detailsでは臨床保存成功と会計同期適用済みを別々に示す。差異だけcommitした成功を会計同期完了と表示しない。

source保存operationに本人が了承を添付する場合、481/483のambient参加で一つのC06 receiptを用いる。別の内部operation_keyで二重成功を作らない。483新行guardのintent allowlistは、単独price_acknowledgementに加え登録済みsource-resolution/importおよび本participantを呼ぶ具体source intentだけへ拡張する。任意intent wildcardを許さず、source owner登録時にexact intentとactor/grantを列挙する。それまで該当source保存時の即時了承枝は無効、上記保存＋差異枝のみ使う。

臨床差分と金額差分が同時なら462再確認を優先し、483のprice-only操作へ数量/identity変更を潜り込ませない。会計が再確認待ちになったことを理由に臨床sourceを巻き戻さない。

## 6. 差異履歴・解決・二重会計防止

既存billing_source_differencesをimmutable observationとして使う。changed_valuesはclosed `{schema_version,before_source,observed_source,billing_input,classification,waiting_for}`。sourceは§3.3の保存projection、billing_inputは505に渡す現在入力projection。classificationはclinical|amount|both、waiting_forはnull|price_acknowledgement|staff_resolution|dedicated_correction|calculation_contract_unavailable。削除は正規ownerのsoft-delete/tombstone版を観測した時だけsource_deleted=true、observed_source=null。取得失敗/物理欠落で架空versionを作らない。

差異INSERTはsource保存/明示submit同TXのみ。binding lock下で同binding版・source版・changed_values digestの既存完全一致を再利用し、再通知で履歴を増殖しない。版だけ進み請求入力が同値なら、実差異を捏造せず適用原値の版追従を同TX実施。元decimal字句差とexact数値差を区別し、価格が同じなら字句差だけで新了承を要求しない。plan_revisionの変化だけから臨床差分を捏造しない。manual=true時は既存原値を上書きしない。

apply_sourceは現在sourceに一致した入力を適用し、applied_source_version/imported_valuesを更新、manual_override=false。keep_billingは値とapplied版/原値を維持、manual_override=true、観測差異を解決記録する。新source版まで永久解決扱いしない。keep時に金額が変わらなければ新505 snapshotや483了承を捏造しない。unknown元値からapplyを構築しない。

解決表は482と同じ最小extension、operation_id bigint nullable、decision nullable(apply_source|keep_billing)、旧行は両null、新writerは両非null、operation compositeFK/immutable/RLS/同clinic/actor/key一致/同TX成功receipt必須。PK(clinic,difference_id)を維持し架空resolution_idを追加しない。新writer resulting_binding_versionは保存後実値。

active source uniqueは全会計に跨る既存部分uniqueを必須とし、再送時upsertで別会計へ横取りしない。releaseは既存三列を同時記録し履歴を残す。source削除を0円行へ置換しない。未手修正の削除反映は既存会計明細の正規削除participant経由で論理削除＋claim release＋462/505を同TXにし、既存理由条件を維持する。必要な削除理由/許可が不足なら差異だけ残して専用操作へ渡す。復元は旧行を無断再活性化せず、active claimが無いことと明示した元source/current版を再検査する。

## 7. 原子的写像・認可・lock

claim短TX → 最外業務TX → 既存 `reservation-graph:write:v1` gate（返されたcontextを全participantへ渡す）→ C06 binding → 関連billings ID昇順 → 既存billing→close順 → records ID昇順 → plan/source/item/bindingのowner規定順。source writerはsource行をlockする前にgateと影響会計の発見/lockを行う。source先取りからbillingへ逆順に入らない。source/明細/確認/確定に関わる全writer subsetの参加がactivation条件。同時import・release・手修正・確定も同じgateへ参加し、影響集合の発見後に新bindingが追加される窓を作らない。

既存public import=create、resolution=edit、read=viewを保持。現在clinic/staff/会計owner/pet、source同患者/record、現在grantをApplyで検査。実際に割引条件を編集する時だけ既存discount:edit、同条件を引き継ぐ派生額再計算だけに新grantを付けない。source内部autoは正規source編集許可に基づく既決の派生更新で、公開された会計権限迂回endpointは作らない。483本人了承を伴う枝は現在accounting:editを別途要求する。sourceだけ閲覧可でも会計詳細を開示しない。

import同TX：source/current projection+readiness再検証 → billing_item採番 → active binding →505 snapshot/lines→billings.current_calculation_id→462必要世代/digest→監査→成功receipt。source保存同TX：source/plan projection+version →差異又は上記会計同期→元owner監査+会計監査→元operation一つのreceipt。resolution同TX：差異観測→選択→行/binding→必要505/462/483→resolution/audit→receipt。いずれか失敗は全rollback。既存serviceが内側WithTxでcommitしないambient関数へ分離する。

監査sinkは既存audit_logs。action=`billing.source_import|billing.source_sync|billing.source_resolution`、resource=`billing_item`、resource_id=対象行、old_value/new_valueは `{billing_id,item_version,binding_version,applied_source_version,source,pricing,manual_override,calculation_id}` の固定前後。metadata=`{schema_version,operation_id,operation_key,source_stamp,difference_id,decision,parent_intent}`。source保存側syncは元operation参照。実際の変更なしで架空old/newや成功監査を作らない。actorは認証staff、IP/UAの新収集なし。監査失敗もrollback。

## 8. 共通operation・照会

専用scope owner_kind=billing、intent=`billing.source_import|billing.source_resolution`。Content exact keysは `schema_version,billing_id,billing_item_id,source_kind,source_id,source_version,source_digest,decision,acknowledgement_digest`。schema_version文字列1、importのitem/decisionは規定sentinel `none`、ackなしもnone。source_idはsource別stable ID、hospitalization等複合は別adapterのcanonical source commitmentをsource_digestへ含む。各値256byte以内、長いJSONをmap値へ入れない。immutable typed commandをdeep-copyしApply前に同じContentへ再hash照合する。hashからpayloadを復元できるとは扱わない。CAS/fenceはPreconditions。

本節のimmutable typed commandは[採択済みQの共通契約](../../design/shared/operation-participation.md#verified-command-selected-20261021)に従い、Applyがreq.VerifiedCommandだけから再構成する。Qはrequest-memoryのみでowner-private耐久保管とは別、CAS/fenceはPreconditions、Contentは既定の短いstring mapを維持する。489のQ必要subsetは未実装依存であり、別command/DTO/pointerのcaptureで迂回せず、既存Content-only consumerは変更しない。


最小claim metadata `billing_source_operation_targets(clinic_id,operation_id,billing_id,billing_item_id,source_kind,source_id)`、itemはimport時null、他NOT NULL。PK(clinic,operation)、binding複合FK、immutable/RLS。業務FKは付けず短claimが業務lockを先取りしない。489 callbackでwinnerだけ同claim TXに保存、失敗はclaimrollback。このwinner-only同TX hookは必要実装依存であり、静的契約から現main実装済みとは扱わない。別TXでanchorを後付けしない。incoming sourceを元binding認可へ流用しない。lookupは保存anchorの元会計で現在view確認しURL一致、source履歴の開示は当該元sourceの現在scopeも検査。欠落・不整合・解決不能anchorは現Runner同様not-authorized/403で非開示、GET修復なし。元targetと現在開示認可が確定した後の結果読取失敗だけresult-unavailable/503。

C06内部ResultRefs/CommittedVersionsはmap[string]string。null値を保存せず、対象が存在しないkeyは省略する。wireの既存nullable DTOへの写像が必要ならシリアライザだけで行い、内部null map仕様を新設しない。refsのallowlist=`billing_id,billing_item_id,calculation_id,difference_id`、versionsのallowlist=`billing,item,binding,calculation`。値は実際の保存済みcanonical文字列であり、現在値から過去receiptを補完しない。

| 保存枝 | 必須refs | 条件付きrefs | 必須versions | 条件付きversions |
|:---|:---|:---|:---|:---|
| import成功 | billing_id,billing_item_id,calculation_id | 保存した差異があればdifference_id | billing,item,binding,calculation | なし |
| apply_source・会計入力変更 | billing_id,billing_item_id,calculation_id,difference_id | なし | billing,item,binding,calculation | なし |
| keep_billing・会計入力不変 | billing_id,billing_item_id,difference_id | なし | billing,item,binding（保存後実値、未変更版は同値） | なし |
| no-op/版追従のみ | billing_id,billing_item_id | 今回固定した差異があればdifference_id | billing,item,binding | 今回計算snapshotを保存した時だけcalculation。対応calculation_idも同時に保存 |
| source削除反映・正規明細削除 | billing_id,billing_item_id,calculation_id | 今回保存した差異があればdifference_id | billing,item,binding,calculation | なし |
| source保存＋差異だけ | 元ownerのreceiptを使用 | 元owner detailsに固定したbilling_id/billing_item_id/difference_idを返す | 元ownerの実保存版 | 会計版は元owner detailsの観測値として保持し、会計適用成功版にしない |

複数会計に影響するsource保存のdetailsは `billing_sync_results` 配列で `{billing_id,billing_item_id,difference_id,applied,observed_item_version,observed_binding_version}` を固定保存し、applied=falseは未適用を意味する。これは元owner detailsのtyped extensionでC06共通mapへ任意配列を入れない。対象なしは配列空、差異なしのdifference_idはwireでnull可だが内部receipt mapへ入れない。適用結果は元操作TXの監査/差異/保存結果から構築し、GET時の現会計を代入しない。

12状態は succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。非成功にrefs/versions/timeなし。結果不明・照会失敗は新key救済/自動再実行しない。

## 9. 算定の限定未閉鎖・切出しと公開条件

483 §8.5/505の税込有理基準から整数配賦/集計へ至る対応は設計未閉鎖のsubsetで、NOT_RUNという表示だけで隠さない。税込を含む当該算定のpreview/import/apply_source/同期による計算適用は、そのsubset契約が閉じ実装接続されるまで有効化しない。原projection保存・unknown表示・設計の閉じた税抜通常subsetは独立して進められ、税込の未閉鎖を全481/外税通常の着手停止にしない。税/率/割引/除外の由来不明を現在master・0・10%・falseで救済しない。

通常source保存にこの算定未閉鎖が影響する場合は、source＋durable差異＋監査を同TX保存し会計未適用を明示する。waiting_for enumへ `calculation_contract_unavailable` を追加し、価格了承待ちと区別する。これは不明計算結果を仮保存する方式ではなく、該当会計のsource整合/確定gateで停止する。全体rollback要件は同じ。計算が閉じた後も最新sourceと再previewを照合し、古い差異から無断再実行しない。


A：一般型/typed reader＋plan-backed treatment projection保存subset＋unknown表示。既存未請求revision/原値を揃える。runtime activation前の文書/fixture PRを独立可。
B：一行import＋binding/505/currentreadiness/receipt。all-domain completion不要、必要な462/484/505/489 subsetを明記。
C：source writer同TX参加・3分岐差異・resolution・483必要参加。古いwriter迂回が残る対象cohortへ有効化しない。
D：482のtrimming adapter接続と405 hospitalization adapter。A/Bの全票完了待ちにせず各source閉じた必要subsetで進める。

検証項目は両先行順（手修正/同期、医師確認/臨床変更、確定/source変更）、同key別内容、active source二重import、未知税/率単位、nullable/0、source削除復元、会計権限失効、監査rollback、同receipt再取得、差異だけの成功/未適用表示。差異/監査失敗rollback、確定歴ありdraft、同source同時import、設定missingと読取unknown、C06内部null非使用、差異だけ成功の表示も対象とする。全てNOT_RUN。§5の限定待機境界は親採択済み。税込算定の限定未閉鎖、必要subsetのruntime接続、全writer参加、受入は残件として分離する。本人への税率・新臨床ルールの質問は増やさない。

## 10. 一次参照

- https://app.plane.so/baritechllc/browse/EMR-481/
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/docs/spec/domains/accounting/lifecycle.md#doctor-review-sync-billing-20261017
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/docs/design/domains/accounting/technical-spec.md#review-version-sync-proposal
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/migrations/001_init.sql#L8150
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/billing/billing_item_unbilled.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/model/treatment.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/docs/work/emr/emr-518-storage-contract-20261020.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/docs/work/emr/emr-484-fulfillment-writer-contract-20261021.md
