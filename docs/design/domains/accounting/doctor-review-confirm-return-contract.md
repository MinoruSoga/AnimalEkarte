# 既存会計の医師確認・差戻し契約

| 項目 | 内容 |
|---|---|
| 文書ID | design.accounting.doctor-review-confirm-return |
| 技術選定 | 2026-10-24 JST / 2026-10-23 UTC、root・独立レビューを経た既存billing限定契約 |
| 基準main | `7dad9adc7f066fab5cc734dc8ff7e125a80d67d4` |
| 状態 | 技術契約選定済み。保存/API・共通gate・writer適合・実装・利用開始は未供給/未完。runtime Readyではない |
| 検証 | tests/build/codegen/実DB/Docker/browserはNOT_RUN。製品・API YAML・実DDL・Plane変更なし |

既存FR/ACを拡張せず、既存の確認集合・差戻し・operationの技術接続を定める。[業務正本](../../../spec/domains/accounting/lifecycle.md)と[作業計画](../../../work/emr/doctor-review-contract-plan-20261023.md#doctor-review-plan-owner-handoffs)に接続する。

<a id="review-actions-scope"></a>
## 1. 有限範囲と正本

一つの既存billingに対する現在集合読取、認証actorによる明示Confirm、現在参照されたsetへの理由付きReturn、不変履歴/C06結果読取、旧確認writer迂回排除を閉じる。複数record可、plan-backed treatmentの完全cohortで部分knownなし。

[採択済み内部契約](doctor-review-internal-projection.md#review-internal-scope)をそのまま使用。失効対象は安定行/source・薬剤/処置identity・数量・record集合のみ。metadata/正当な技術link-onlyは世代/digest不変。481 last-imported snapshotは現在採用identityと一致を証明できる行のみ再利用、current数量はbilling_items。現在treatmentで過去値を補修しない。新current projection列は追加しない。

Confirmは484のprepared/performed生成や履行完了を要求せず、505/506・金銭・在庫・請求確定を発火しない。billing作成、新規billing Complete接続、生前分C、退院takeover Dは別packet。担当医以外の医院が認めた医師の代行可、Return主体、price-onlyの既決を再質問しない。

<a id="review-actions-wire"></a>
## 2. 共通wire

以下のroute表記はprefixを省略する。公開時は既存 /api/v1 配下とし、別の /v1 公開面を追加しない。ID/世代/版は正signed64 canonical十進string、digestはlowercase64hex、operation_keyは既存ParseOperationKeyのlowercase hyphenated非nil UUID。unknown/duplicate JSON key、型不一致、非正規IDを拒否し黙って正規化しない。POSTはapplication/json・8KiB上限、超過は413で全要求拒否、record切捨てなし。既存cookie/CSRF/認証を維持し新bearer/Idempotency-Key headerなし。

memoはstring省略可、null拒否、省略は空文字、Go strings.TrimSpace後1000 Unicode rune以下。return_reasonはstring必須、同TrimSpace後非空500 rune以下。Unicode正規化や内部空白圧縮なし。

<a id="review-actions-current"></a>
### 2.1 現在集合
GET /accountings/{billing_id}/doctor-review

現在accounting:view＋元billing同院scope。同一consistent read snapshot、書込/lock/baseline生成なし、Cache-Control:no-store。
exact200: {billing_id,billing_version,current_review_set_id,projection,review_state}。
current_review_set_idはpositive-ID string|null。projectionは採択内部契約のKnown/Unresolved/Unsupportedそのまま。Known exact: {state:"known",billing_id,billing_generation,billing_digest,records,billing_clinical_items}。record object: {medical_record_id,clinical_generation,clinical_digest,clinical_items}。record item: {source_kind:"treatment",treatment_id,item_kind,master_id,quantity}。billing item: {billing_item_id,source_kind:"treatment",source_id,medical_record_id,item_kind,master_id,quantity}。item_kind=medicine|procedure。quantityはcanonical decimal string。失敗projectionは{state,reason}だけで臨床集合の部分payloadなし。

review_state:
- none: 現在参照NULL。
- valid: current knownと保存modern snapshotのclinic/billing・完全record集合・両側generation/digestが一致。
- stale: 確定的不一致。
- unresolved: current又は保存snapshotの完全性を証明不能。
過去確認医師の後日の退職等を、このreadで新しい失効条件にしない。

<a id="review-actions-confirm"></a>
### 2.2 Confirm
POST /accountings/{billing_id}/doctor-review/confirm
closed body: {operation_key,expected_billing_version,expected_current_review_set_id,reviewed,memo?}。
reviewed exact: {billing_generation,billing_digest,records}。
records各要素exact: {medical_record_id,clinical_generation,clinical_digest}。ID数値昇順へcanonical化、重複拒否。現在参照CASはpositive-ID|null。

reviewedは明示確認対象のimmutable Content。CASだけ更新しても同keyのreviewedは変えない。任意臨床行をclientから書かせず、同lock下でserverが全量signatureを再構成しreviewedの全集合/世代/digestと一致したpreimageを保存する。
新keyで現在同一集合への有効確認が既存ならvalidation（無断履歴量産/memo修正にしない）。同key成功は先にreplay。stale/unresolved旧参照でもcurrentが完全knownかつ明示した現在参照CAS一致なら新setを作り旧set不変。

<a id="review-actions-return"></a>
### 2.3 Return
POST /accountings/{billing_id}/doctor-review/return
closed body: {operation_key,expected_billing_version,review_set_id,return_reason,memo?}。
review_set_idは元setのimmutable Content。現在billings.doctor_review_set_idと一致＋billing CASを検査。現在projectionがunknown/staleでも現在参照されたsetを理由付き解除可能。新医師資格・臨床再確認条件なし。
成功はReturn履歴追記と現在参照NULL化のみ。旧set/record snapshot/confirmation_note/clinical世代・digest不変。別key二重Returnは現在参照不一致409、同key成功はreplay。

<a id="review-actions-lookup"></a>
### 2.4 元operation照会
GET /accountings/{billing_id}/doctor-review/operations/{action}/{operation_key}
action=confirm|return（closed）。認証clinic・同actor・owner/intent/keyでbinding検索→保存anchorの元billing現在accounting:view→URL一致検査。不一致404。URLを元target証明にしない。確定的なbinding不在だけscope-not-found。anchor欠損/現在view拒否/元target解決不能は非開示、GET修復なし。

<a id="review-actions-stored"></a>
### 2.5 保存set読取
GET /accountings/{billing_id}/doctor-review/sets/{review_set_id}
現在accounting:viewと元set clinic/billing一致。modern exact200:
{state:"stored",billing_id,review_set_id,billing_generation,billing_digest,billing_signature,records,reviewed_by,reviewed_at,confirmation_note,return_event}。
recordsは保存doctor_review_recordsから {medical_record_id,clinical_generation,clinical_digest,clinical_items} の4keyだけを投影し、採択内部型とID数値昇順を使用する。保存側のclinic/review_set/billing列や現在sourceを混ぜない。reviewed_byはpositive-ID string、confirmation_noteは保存値string|null。reviewed_at/returned_atは保存instantをUTC RFC3339で表現し、PostgreSQLの実精度を維持する（最大microsecond、架空小数桁や補完日時なし）。return_event=null又は{ return_id,returned_by,returned_at,return_reason,memo }でIDはpositive-ID string、reason/memoは保存済string。
legacy不足exact: {state:"legacy-unavailable",billing_id,review_set_id,reason:"missing-billing-signature"}。不存在の会計当時内容を生成せず、modernの部分推定payloadを返さない。

<a id="review-actions-c06"></a>
## 3. C06・認可・response

scope: owner_kind=billing、actor_kind=staff、actor_ref=認証staff canonicalID、clinic=trusted。
intent=billing.doctor_review.confirm / billing.doctor_review.return。
Confirm canonical command exact: {schema:"billing-doctor-review-confirm-command-v1",billing_id,reviewed,memo}。
Return: {schema:"billing-doctor-review-return-command-v1",billing_id,review_set_id,return_reason,memo}。
JCS SHA256 lowerhex command_digest。
Confirm Content string map exact: schema_version="1",billing_id,reviewed_billing_generation,reviewed_billing_digest,command_digest。
Return Content: schema_version="1",billing_id,review_set_id,command_digest。
集合/理由/memoはcommand_digestへ束縛。Q=NewVerifiedCommand、Applyはreq.VerifiedCommandのみから復元し別capture DTO禁止。Qはrequest-memoryで未成功raw commandの汎用耐久保存は追加しない。別fingerprint/receipt基盤なし。
Preconditions: expected_billing_version、Confirmのみexpected_current_review_set_id（null内部sentinel none）。Returnは元Contentのreview_set_idと現在参照一致を検査。技術source.version/link版は確認Contentへ混入しない。

ResultResponse direct body、Outcome/Binding/EffectSummary非公開。
Confirm成功refs {billing_id,doctor_review_set_id},versions {billing,clinical_set_generation}。
Return成功refs {billing_id,doctor_review_set_id,doctor_review_return_id},versions {billing}。
すべてmap値string、nullなし。成功recorded_at必須、既存time.Time RFC3339Nano serialization。
succeeded/replayed=200; lookup not-succeeded=200; operation-in-progress=202; 記録済unknown/result-unavailable=503; key/version/dependency/edit-right conflict=409;validation=400;scope-not-found=404;not-authorized=403。非成功C06 bodyはstatus/operation_keyのみ。

C06前の不正要求400 {error:"invalid-request"}、未認証401、認可拒否403。現在認可後のstore/provider/serializer障害500 {error:"internal-error"}。既存ErrorResponse {error:string} と httpapi.RespondError の経路を維持し、新error_code envelopeを追加しない。provider未配線を記録済unknownと偽らず、500/接続断から成功不存在を推定しない。

新Confirm/Return effectはaccounting:edit、read/lookup/保存成功replayはaccounting:view。無条件edit middlewareでview-onlyになった本人の成功replayを妨げない。
1. 認証/selected clinic/view、同scope/key非変更lookup。
2. 既存bindingは元anchor現在viewでsuccess/unknownをCAS/医師/readinessより先に処理。
3. 確定不在newkeyだけeditとtargetscopeをread-only preflight後claim。
4. loserは保存anchorで再認可、incomingtarget追記なし。
5. 未成功Applyで現在edit/scope/適格性/CAS再検査。

ConfirmのみT03共用AssertMedicalRecordDoctorInClinic port：認証actor、active/nondeleted doctor分類・同院nondeleted assignment、mandatory ambient・SHARE NOWAIT adapter、designation/legacy/default/import根拠の受領。client doctor_id/担当医ID/免許文字列/職種名を代替にせず新role/grantなし。unconfiguredは当該confirm停止、通常編集/read停止なし。現在適格性と明示確認行為は別で、command+actor+保存signature+時刻+operation+同TXsuccessが後者の証拠。Return医師限定なし。

<a id="review-actions-storage"></a>
## 4. 選定した保存差分（未供給・未実装）

<a id="review-actions-set-storage"></a>
### 4.1 既存doctor_review_sets
billing_signature jsonb NULL、operation_id bigint NULLを追加。
両NULL又は両非NULL、signature非NULLはobject、operation正値、(operation_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT、非NULLの(clinic_id,operation_id)一意。既存clinic/billing FK・record_count・immutable・RLS維持。legacy両NULL保持、空object/backfill/旧immutable更新なし。
新producerINSERT guardは両非NULL必須、新setへoperation_success_required() deferred receipt条件。
BillingSignature *json.RawMessage、OperationID *uint64を選定する。ConfirmationNote列はそのまま。

signature exact preimage: {schema:"doctor-review-billing-v1",clinic_id,billing_id,records:[{medical_record_id,clinical_generation,clinical_digest}],billing_clinical_items:[限定billing item]}。
typed producer/readerがexactkey/type/orderとJCS SHA256対billing_clinical_digest、clinic/billing、payload集合対doctor_review_records/record_count、record側clinical_items対digestを検証。SQL jsonb text hashをJCS代用にしない。deferred集合整合は既存membership guardへ必要比較を追加する単一owner差分。metadata/price/技術linkは追加しない。

<a id="review-actions-return-storage"></a>
### 4.2 doctor_review_returns
append-only履歴を選定する。id bigint identity PK; clinic_id,billing_id,review_set_id,operation_id,returned_by bigint NOT NULL; returned_at timestamptz NOT NULL default clock_timestamp(); return_reason text NOT NULL; memo text NOT NULL default ''。
ID正値、(review_set_id,clinic_id,billing_id)→doctor_review_sets(id,clinic_id,billing_id) RESTRICT、(operation_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT。clinic_id REFERENCES clinics(id) ON DELETE RESTRICT、returned_by REFERENCES staffs(id) ON DELETE RESTRICT を既存doctor_review_setsと同じ規約で選定。UNIQUE(clinic_id,operation_id)、UNIQUE(clinic_id,review_set_id)。同setは一度Return、再確認は新set。
reason非空/500文字以下、memo1000以下、serviceは先行TrimSpace/rune。既存contract_history_immutableでUPDATE/DELETE拒否、clinic RLS/全repo clinic predicate、deferred success-required。新counterなし。
Return事実だけで第二確認集合ではない。金融billing_change_eventsへreview_returnを混ぜない。legacy setにもReturn可、失われたsignatureを要求しない。

<a id="review-actions-anchor"></a>
### 4.3 billing_review_operation_targets
clinic_id bigint NOT NULL、operation_id bigint NOT NULL、billing_id bigint NOT NULL CHECK>0、PK(clinic_id,operation_id)、sole FK(operation_id,clinic_id)→operation_bindings(id,clinic_id)RESTRICT。immutable/clinic RLS。business billing FKなし、claim短TXで業務lock先取りなし。
ClaimMetadataWriterのwinnerだけ同claimTX INSERT、失敗時bindingもrollback。resolverはbinding+clinic+owner/intentを照合。GET修復なし。同等anchor供給が来たらsemanticを統合し二表なし。

<a id="review-actions-locks"></a>
## 5. TX/lock順

metadata-only claim短TX終了後、最外業務TX開始直後に既存reservation-graph:write:v1 try gate→binding UPDATE NOWAIT→元anchor現在view/success/unknown→未成功savepoint/Apply。
Apply:
1. 現在grant/clinic/認証staff保護。新set/Return INSERTのclinic FK参照先をKEY SHARE NOWAITで先行取得し、必要staff-global等の既存保護はstaff row前に取得。Confirmはstaff/assignment SHARE NOWAITで共用医師predicate。Returnも既存LockActiveStaffAssignment相当のactive/nondeleted staff・同院assignment保護をSHARE NOWAITで維持するがdoctor分類は要求しない。後段FKが未宣言のclinic/staff lockを初めて取る形にしない。
2. billing UPDATE NOWAIT、参照/版/scope再読。
3. Confirmのみcurrentrecord集合再発見→record ID昇順→plan rows ID昇順→plan_items dependency→typed treatment ID昇順→billing_items ID昇順→activebindings決定順。同ambientprovider。
4. Confirmはcurrentsignature/baseline/Content一致。Returnは元set/現在参照一致、currentsource全量不要。
5. set+records又はReturnhistory→billing参照→audit→C06receipt同TXcommit。
集合変化で下位取得後に未宣言上位lockが要る場合は最外TXをやり直す。強いlockの後付け/別TX/commit先行なし。

481 gate→binding→billing→必要close→records→plan/sourceと484 source-only gate→binding→editrights→record→plan→dependency→treatmentを再定義しない。billing影響source変更は先に影響billing列挙するclosureが必要。Complete close→billingとrefund/item billing→closeは共通gateの上位直列化で保護、全owner単一FK順への反転なし。Confirm/Returnはclose/fund/method/appointmentや506資金participantを不要に追加しない。全writer/auth/staff/membership参加証拠なしにgateだけで安全としない。
runtime必要subsetは489 optionalContextGate相当の同ambientcontext伝搬、467同namespace trygate、医師NOWAITadapter、cohort writer閉包/currentprovider。採択464のhookを共用し第二実装なし。現Runner Gate(ctx)errorを供給済ContextGateにしない。

Confirm/Returnはclinical世代/digest不変。doctor_review_set_id変更のbillings.versionは既存entity_version_guard（固定001_init.sql L10236–10243、billings_storage_version L10904）が一回増分、adapterのversion+1二重書込なし。CAS WHERE後RETURNINGで保存後版。失敗時set/record/history/audit/参照/receipt全rollback。commit後配送失敗をrollback済みと呼ばない。

<a id="review-actions-bypass"></a>
## 6. 監査/旧入口

action=billing.doctor_review.confirm又はreturn;resource=billing;resource_id元billing;actor認証staff。
old/new exact {doctor_review_set_id:positiveID|null}。
Confirmmetadata {schema_version:"1",operation_id,operation_key,doctor_review_set_id,billing_generation,billing_digest}。
Returnmetadata {schema_version:"1",operation_id,operation_key,doctor_review_set_id,doctor_review_return_id}。
理由/memoはdomain不変履歴へ、無条件二重複製なし。必須JSON失敗をbest-effortで落とさずSQL失敗も全rollback。

旧 GET /medical-records/{id}/billing-confirmation、POST同/confirm・/returnにはbilling/集合/C06keyがないため、安全な新writeへ推定変換しない。利用開始cohortでは旧GET GetOrCreate INSERTを除去し、旧GET/POSTは現在認可後409 {error:"billing-review-contract-required"}を選定する。対応UIはbilling-scoped GETへ同切替。唯一らしいbilling、架空confirmed/pending ID時刻を生成しない。未切替cohortのDROP済旧table callerを互換正常としない。旧status PATCH/Complete/直接repoが確認を作る迂回もcohort台帳で閉じる。全sourceを無断新gateへ移さない。

<a id="review-actions-complete-dependency"></a>
## 7. 別の準備billing/新規Complete packet

既存billing前提は内部subsetで、製品のbillingなしComplete禁止ではない。462統合ownerが /accountings作成と/complete について、準備billing+全binding作成再読、操作key/重複防止/未請求版/同TX初期baseline、準備identityと現BillingID/takeover互換、初回201/再送200/既存Billing整数wire/原請求日を維持するadapterを閉じる。他billing確認流用・自動/架空確認なし。ここが未供給の間Complete新gateを有効化せず、billing_id必須を無断追加しない。

<a id="review-actions-acceptance"></a>
## 8. 有限受入とowner

fixture期待: canonical/unknown/duplicate/null/上限/正規化;複数record完全性;481currentidentity証明可否;signature違い拒否とmetadata/link-only不変;別key既確認拒否/同keyreplay;CASのみretry対Content変更;source変更対Confirm両順序;Return対newConfirm両順序;unknown/stale/legacy Returnと二重拒否;actor≠担当医/未設定/失権/reader欠落/Returndoctor制限なし;成功後view-onlyreplay/view失効非開示;anchor/URL/lookup障害;claim winnerloserrollback;gate/binding/domain競合/異TXContextGate拒否;billingversion一回/clinical不変/overflowzero-write;全必須保存失敗rollback;commit後応答喪失同key照会;確認当時billingsignature再読;旧GETzero-write/旧POST拒否/迂回閉鎖;新規Complete未接続で旧wire無断変更なし。全NOT_RUN。

owner:462統合、481currentidentity/sourceclosure、518/484plan/gateclosure、sharedstaff/reservationdesignation/predicate、489/467hook/trygate、446/466+model保存、445/API exactwire/旧入口切替、audit同TX。個人claim/実行ID/供給SHA/共有hunk合意は未受領で、担当済としない。技術契約選定とruntime供給を分離する。Readyは計画/契約状態と実行前提の明示で、実装完了ではない。

<a id="review-actions-evidence"></a>
## 固定根拠と確認限界

固定mainの以下のsourceを根拠とする。実DBのtrigger/FK適用やruntime供給を検証したものではない。保存差分は文書選定であり、既存列/実装が存在するという主張ではない。

- [docs/design/domains/accounting/doctor-review-internal-projection.md](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/docs/design/domains/accounting/doctor-review-internal-projection.md)
- [backend/internal/billing/routes.go#L100-L106](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/billing/routes.go#L100-L106)
- [backend/internal/billing/billing_confirmation_request.go](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/billing/billing_confirmation_request.go)
- [backend/internal/operation/contract.go](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/operation/contract.go)
- [backend/internal/operation/dto.go](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/operation/dto.go)
- [backend/internal/reservation/reservation_intent_repository.go#L222-L239](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/reservation/reservation_intent_repository.go#L222-L239)
- [docs/design/domains/reservations/technical-spec.md#L420-L444](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/docs/design/domains/reservations/technical-spec.md#L420-L444)
- [docs/work/emr/emr-481-source-binding-contract-20261021.md#L100-L118](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/docs/work/emr/emr-481-source-binding-contract-20261021.md#L100-L118)
- [docs/work/emr/emr-484-fulfillment-writer-contract-20261021.md#L68-L80](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/docs/work/emr/emr-484-fulfillment-writer-contract-20261021.md#L68-L80)
- [docs/design/domains/accounting/owner-deposit-initial-receipt.md#L208-L240](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/docs/design/domains/accounting/owner-deposit-initial-receipt.md#L208-L240)
- [backend/migrations/001_init.sql#L8160-L8192](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/migrations/001_init.sql#L8160-L8192)
- [backend/migrations/001_init.sql#L8821-L8839](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/migrations/001_init.sql#L8821-L8839)
- [backend/migrations/001_init.sql#L9456-L9475](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/migrations/001_init.sql#L9456-L9475)
- [backend/migrations/001_init.sql#L10235-L10244](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/migrations/001_init.sql#L10235-L10244)
- [backend/migrations/001_init.sql#L10901-L10904](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/migrations/001_init.sql#L10901-L10904)
