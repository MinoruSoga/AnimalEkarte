# EMR-516/517: 保存済み検査・接種と共通予定の関連 mutation 採用契約

2026-10-21。技術方式をselectedとして固定する。文書公開とruntime実装を区別し、runtime未実装・全writer必要subset未接続・試験とDB適用は全て NOT_RUN。業務ルール、関連多重度、新roleは追加しない。現行 main 読取基準は `edf71515dbae11735a4184ff64c465e2ea51af88`。

## 1. 結論と固定境界

518の `medical_record_plan_items` / `medical_record_plan_item_revisions` / `medical_record_plan_owner_links` と `medical_record_dependency_versions.owner_kind='plan_items'` を唯一の共通予定基盤とする。516がS1/S1aのschema単一owner、517は接種consumer。現在mainで確認した001とmodelにはこの3表の実装がなく、確定設計と実装済みを区別する。

- 本契約のwriteは既存結果／実績との関連追加・関連訂正・関連解除のみ。保存済みexam/vaccinationを作成・訂正・確定・削除せず、予定stateも変更しない。実施操作は別intentとしてPlanExecutionCoordinatorに明示委譲する。関連があるだけでperformedにしない。
- PR759 §5.3のactive uniqueをそのまま適用する。同一検査予定に複数examを関連できる。一つのowner/roleを複数planへ共有しない。vaccination等primary concrete execution_sourceは一つ。charge_sourceは別制約で、本契約は一切作成しない。
- 検査linkは `link_purpose=result`、接種linkは `link_purpose=execution_source` にこのAPIが固定する。クライアントにpurposeの自由指定をさせない。接種実績を主実施元として参照する意味であり、link保存自体は臨床実施操作ではない。
- 機器受信・患者attachはplan選択ではない。患者／recordがNULL又は不一致の結果は自動補完しない。同名・同日・予約・来院・数量から対応を推定しない。
- 後日予定の追加再掲は既決どおりカルテのレポートページのみ。別カルテで元予定を消化しない。既存3タブの同一予定表示は保持する。
- 実施、請求、実在庫、証明書、次回日、タグ／通知、機器送信の自動起動なし。498のdose、500の健診owner、495の診断／目的別訂正を混ぜない。

## 2. 現存ownerの根拠と適合差

GitHub固定リンクのprefixは `https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/`。

| 対象 | 読取根拠 | 実存する契約／今回の注意 |
| --- | --- | --- |
| 共通storage | docs/work/emr/emr-518-storage-contract-20261020.md §5.3,6,7 | ownerのclinic複合FK、active時record/pet一致、単調revoke、固定版保持、共通lock順。PR759 merge f8fbf038069df5b71b792eefadb9ebd8988f966c |
| API基盤 | docs/work/emr/emr-518-package-api-contract-20261020.md | plan GET二操作、共通operations GETは公開済み契約。link mutationは未公開であり本契約の新規subset |
| 494/C06 | docs/work/emr/emr-494-edit-draft-contract-20261020.md、backend/internal/operation/{contract,dto}.go | body operation_key/edit fence、ContentとPreconditions分離、flat ResultResponse |
| 実route | backend/internal/medicalrecord/routes_records.go | examinations GET/POST/PATCH/unconfirm/DELETE、vaccinations GET/POST/PATCH/DELETE。plan link routeなし |
| 検査 | backend/internal/model/{examination_record,examination_revision}.go | examsには一般versionなし。current_revision_version nullable、実revisionにはid/version/kind・患者／record snapshot |
| 検査更新 | backend/internal/medicalrecord/examination_service_update.go | confirmed/保護版を通常更新不可、revision存在後のpet/record変更は拒否。患者付替えをlinkで代行不可 |
| 機器保存 | backend/internal/medicalrecord/lab_device_exam_persist.go | 患者紐付け時のexam生成はplan_idを持たず、medical_record_idも自動選定しない |
| 機器使用 | backend/internal/medicalrecord/lab_import_usage_tracker.go | clinical useはjob→exam lockで補償と直列化。manual_mutation receiptはambient TXで保存。noop配線は本番linkの証拠に不可 |
| 接種 | backend/internal/model/vaccination_record.go、medicalrecord/vaccination_request.go | vaccinations.Versionが真の版。旧PATCHはversion省略可だが本契約は省略不可。pet/recordはnullable |
| 接種保存 | backend/internal/medicalrecord/vaccination_service.go | 独自WithTx、旧parent draft gate、保存後タグ同期あり。本契約からCreate/Updateを呼ばない |
| 業務正本 | docs/spec/domains/{examinations,vaccinations}/lifecycle.md | 原結果、official、専用訂正、予定と実績分離。D07機器Undoは元予定関連を残したdetachを拒否 |
| 484連携 | docs/work/emr/emr-484-fulfillment-writer-contract-20261021.md（PR780 merge455ec2f4d84c5e08da568bf99fc6d825506b499d）、516最新本文 | execution_source等の変更時、実存sourceの最終contextを単一UPDATE、guard RETURNING generation。sourceなしなら生成しない |

Plane読取根拠: https://app.plane.so/baritechllc/browse/EMR-516/ と https://app.plane.so/baritechllc/browse/EMR-517/ の本文・全コメント。516の2026-10-21依存handoffを旧起票時の未選定記述より優先する。

## 3. Exact route / intent（6操作、6path）

以下はAPI YAML上のpath。実routerの既存 `/v1` prefixを二重に追加しない。

| method/path | operationId | C06 intent |
| --- | --- | --- |
| POST /medical-records/{record_id}/plan-items/{plan_item_id}/examination-links | linkPlanExamination | medical_record.plan_examination_link |
| POST /medical-records/{record_id}/plan-items/{plan_item_id}/examination-links/{link_id}/replace | replacePlanExaminationLink | medical_record.plan_examination_link_replace |
| POST /medical-records/{record_id}/plan-items/{plan_item_id}/examination-links/{link_id}/unlink | unlinkPlanExamination | medical_record.plan_examination_unlink |
| POST /medical-records/{record_id}/plan-items/{plan_item_id}/vaccination-links | linkPlanVaccination | medical_record.plan_vaccination_link |
| POST /medical-records/{record_id}/plan-items/{plan_item_id}/vaccination-links/{link_id}/replace | replacePlanVaccinationLink | medical_record.plan_vaccination_link_replace |
| POST /medical-records/{record_id}/plan-items/{plan_item_id}/vaccination-links/{link_id}/unlink | unlinkPlanVaccination | medical_record.plan_vaccination_unlink |

`replace` は関連だけの誤り訂正または参照版更新。実績本文の訂正APIではない。新ownerは同じrecord/petのみ。別planへの移動はこのrouteでは扱わない。旧linkをrevokeし新linkを同TXで追加する。一旦unlinkを成功させてから別createを自動実行する二段方式を採らない。

共通lookup `GET /medical-records/{record_id}/operations/{operation_key}` に上記6intentのallowlistを追加する。既存package等のintentを置換せず、link専用lookupを増設しない。lookup実装は494/489単一ownerへ差分要求する。

## 4. DTOと厳密な識別

全bodyに必須 `operation_key, edit_session_ref, edit_epoch, owner_expected_version, plan_expected_version`。operation_key/sessionは既存canonical UUID、全ID/版は先頭0なし正整数decimal stringかつ1..9223372036854775807をserver検査する。owner_expected_versionは `plan_items` aggregate、plan_expected_versionは個別plan.version。clinic/pet/actor/日時/purpose/clinical_stateのbody指定、unknown field、JSON duplicate keyは拒否する。Cookie/既存CSRFを継承し新Idempotency-Keyヘッダを足さない。

### 検査link

追加必須 `examination_id, examination_guard, source_revision`、任意nullable `source_component_path`。

- examination_guardはclosed union。
  - revisioned: `{kind:'revisioned', current_revision_version:Version}`。
  - unversioned: `{kind:'unversioned', source_digest:lowercase_sha256_hex}`。真にCurrentRevisionVersion=NULLの保存済みlegacy/未revision結果のみ。owner_versionはNULLのまま。
- source_revisionはclosed union。
  - revision: `{kind:'revision', revision_id:Id, revision_version:Version}`。同clinic/examの実存immutable revisionを要求する。current guardは現在headのCASであり、参照する固定版と別。旧officialを参照する場合は旧版と表示し、最新版へ偽装しない。
  - unversioned: `{kind:'unversioned', source_digest:Hash}`。guardと同digestのみ。架空revision ID/versionを生成しない。link revisionのprovenanceへ物理exams/id、digest、非versionedである事実を保存し、歴史表示の値が必要なら不変の参照時projectionを同revision snapshotへ保存する。編集可能な結果本文の複製にはしない。
- 実revisionのcaptured clinic/record/petも現在ownerとtargetに一致すること。現在ownerの患者が一致するだけで、別患者時代の固定版を参照可能にしない。revisionが指すexam IDも必ず一致させる。
- source_component_pathはNULL又はD07が返した安定component識別子を完全一致検査する。自由入力、項目名、配列indexを認めない。実存D07 adapterが子識別を公開しない対象はNULLでexam全体を参照する。component pathによって518 storage契約§5.3のowner単位uniqueを迂回しない。
- source_digest算出はserverの版固定projection `exam-link-source-v1` のJCS SHA256。exact keys: `schema,clinic_id,examination_id,medical_record_id,pet_id,current_revision_version,status,exam_type_id,doctor_id,date,performed_date_precision,result_summary,machine,job_id,updated_at,items`。itemsはid昇順、各要素 `id,exam_type_field_id,name,inspection_value,normal_value,result,unit,reference_value,ref_min,ref_max,is_abnormal,status,sort_order,qualitative_min,qualitative_max,updated_at`。各既存NULLはNULL、IDはcanonical string、dateはDB dateのYYYY-MM-DD（NULLはNULL）、updated_atはUTC RFC3339の小数9桁固定（Z）、ExactNumericは既存の正規decimal string、sort_orderは整数／NULL、boolはbool／NULL、原値文字列はそのまま。unknown dataを空値で補完しない。digestは実装時にDTO/modelと同版公開する。child writerもexam親lockへ参加し、読取とCASの途中でitemsだけ変わる経路を残さない。

### 接種link

追加必須 `vaccination_id, vaccination_expected_version`。参照元のvaccination.Versionを必須化し、実際のID/版をlinkへ保存する。source_component_pathはNULL固定。owner_revision_refは `{kind:'vaccination',vaccination_id:Id,version:Version}`。これを実在しない接種revision表のIDと解釈しない。参照時projectionはlink revisionのimmutable snapshotで履歴保護し、現在の接種本文はownerから読む。取得した版が変わったら現在版更新ありと区別する。

### replace / unlink

- replaceは各link bodyに `expected_link_revision_id, reason` を追加。path link_idが現在activeで、同plan/owner-kind/purpose/clinicとexpected_link_revision_idが一致すること。新owner/新版は上記linkと同じ検証。reasonは既存誤関連訂正理由に相当しtrim後非空。D07 unconfirmの500文字条件を関連訂正へ無根拠にコピーしない。
- unlinkは共通body＋ `expected_link_revision_id, owner_guard, reason`。owner_guardは対象型の上記現在版/legacy digest。new ownerフィールドは受けない。元linkが既にrevokedなら新keyではvalidation、元keyでは保存済みreceipt replay。削除されているownerは歴史anchorを検証できるscoped lookupで扱い、原本文を消したり復活させない。
- 新規linkと同じactive関係が既に存在する場合、新keyではvalidationとして既存関係を再取得する。成功に見せてrevisionを増やすno-opや二重receiptを作らない。same-key replayは通常どおり。

## 5. C06 Content / CAS / target

Scopeはtrusted context由来のclinic、owner_kind=`medical_record`、actor_kind=`staff`、actor_refと上記intent。path record/plan/linkとbody owner IDはcanonical化してContentへ入れる。短claim anchorは494のimmutable record anchorを使い、新business FKをclaimへ追加しない。

### 5.1 Immutable typed commandとcommitment

bodyを検証後、全nested struct、slice、nullable stringをdeep-copyして変更不可のconsumer commandとする。commandのexact keysは `schema_version,action,record_id,plan_item_id,owner_kind,new_source,old_link,reason`。schema_version='1'、action=link/replace/unlink。new_sourceはlink/replaceで `{owner_id,reference,source_component_path}`、unlinkではNULL。referenceは§4のtyped source_revision又は接種 `{kind:'vaccination',vaccination_id,version}`。old_linkはreplace/unlinkで `{link_id,plan_revision_id,owner_id,reference,source_component_path}`、linkではNULL。old_linkは現在認可付きserver readから生成し、bodyのexpected_link_revision_idと一致検査する。reasonはlinkでNULL、replace/unlinkでは原入力文字列を保持しtrim判定だけで非空を検査。原文を256byteへ切り捨てない。new/oldとも参照に必要なNULLを明記しunknown fieldを受けない。

本節のimmutable typed commandは[採択済みQの共通契約](../../design/shared/operation-participation.md#verified-command-selected-20261021)に従い、Applyがreq.VerifiedCommandだけから再構成する。Qはrequest-memoryのみでowner-private耐久保管とは別、CAS/fenceはPreconditions、Contentは既定の短いstring mapを維持する。489のQ必要subsetは未実装依存であり、別command/DTO/pointerのcaptureで迂回せず、既存Content-only consumerは変更しない。


old_link再構成はrevoked履歴も読む。linkの固定owner/ref/component/plan_revision_idはrevokeで変更しないため、同keyのreplace/unlink再送でも元commandを再構成できる。activeであることは新Applyだけで検査し、replay前段でrevokedを拒否しない。既存bindingなら保存metadataを先に現在認可付きで照会し、現在ownerの内容や現在版を要求し直さず元payload hashと不変link参照を照合する。新bindingだけに次節の新規target preflightを適用する。現在のmaster/owner更新を理由に認可済み過去成功receiptを再実行へ回さない。

command_sha256はこの全commandのJCS SHA256。reason_sha256はJCS(reason string又はNULL)のSHA256、component_sha256はJCS `{old:old_link.source_component_path又はNULL,new:new_source.source_component_path又はNULL}` のSHA256。いずれもlowercase64hex。JCS UTF-8、Unicode文字列の勝手な正規化なし、NULLと空文字は区別。componentはowner返却安定identityであり、hash化は意味の変更や任意文字列の許可ではない。

C06 Contentのexact string keysは **10個**: `schema_version='1',intent,record_id,plan_item_id,owner_kind,action,command_sha256,reason_sha256,component_sha256,target_kind='record'`。各値は固定enum/intent、最大19桁ID、64桁hashで256byte以内。reason/componentの原文をContentに入れない。NewContentの既存制限とfingerprint実装は変更しない。

Applyはimmutable commandを再JCS化しcommand/reason/component hashを再計算しContentとconstant-time比較する。さらにContentのtarget/owner/action/intentとcommandを厳密一致検査する。不一致はvalidation/writeゼロ。以後のwriteはこの検査済みcommandのみから行い、mutable requestを閉包で参照しない。hashからpayloadを復元できるとは記載しない。commandのsource参照は実ownerと固定版をlock後再検証する。

Preconditions exact keysは `edit_session_ref,edit_epoch,owner_expected_version,plan_expected_version,owner_guard_kind,owner_guard_version,owner_guard_digest`。guardの非該当値だけ空文字。replaceは旧ownerも `old_owner_guard_kind,old_owner_guard_version,old_owner_guard_digest` を追加しwire必須 `old_owner_guard` で受ける。unlinkのowner_guardは旧owner、linkは新owner。CAS/fenceはhash対象外だが選択固定版・旧link revision・理由はcommand hash対象。CASだけ再確認は同key、選択版や理由の変更は旧結果確定後に新明示操作。

### 5.2 未成功lookupの永続target（最小補助metadata）

既存 `operation/store.go` のClaimBindingはscope、key、fingerprintのみを保存する。Content本文もplan/owner targetも永続化しない。494 record anchorだけでは未成功bindingのowner認可ができない。成功receiptがまだない時にincoming URL/bodyを元対象とみなしてはいけない。

既存494 record anchorは変更せず、516 S1に補助表 **medical_record_plan_operation_targets** を追加する採用設計。これは予定基盤や新receiptではなくclaim時target metadataで、PR759の既決clinical3表を置き換えない。columns:

- `clinic_id bigint NOT NULL, binding_id bigint NOT NULL`、PK(clinic_id,binding_id)、唯一FK `(binding_id,clinic_id)` → operation_bindings(id,clinic_id) RESTRICT。
- `schema_version smallint NOT NULL CHECK=1, record_id bigint NOT NULL CHECK>0, plan_item_id bigint NOT NULL CHECK>0, pet_id bigint NOT NULL CHECK>0`。
- `owner_kind text NOT NULL CHECK IN ('examination','vaccination'), action text NOT NULL CHECK IN ('link','replace','unlink')`。
- `old_owner_id bigint NULL CHECK>0, new_owner_id bigint NULL CHECK>0, old_link_id bigint NULL CHECK>0, old_link_revision_id bigint NULL CHECK>0`。
- `command_sha256 char(64) NOT NULL CHECK lowercase64hex`。
- shape CHECK: linkならold三列NULL/new NOTNULL、replaceなら全四列NOTNULL、unlinkならold三列NOTNULL/new NULL。旧新ownerが同IDの版更新は可。

新binding作成時のみbinding＋494 anchor＋このmetadataを同じ短claim TXでINSERTし、一つでも失敗なら全rollback。既存bindingに後付け・上書き・upsert更新しない。business record/plan/pet/ownerへのFKを付けずclaimでbusiness lockを取らない。全列immutable（UPDATE拒否、通常DELETEなし）、既存clinic RLS helper/tenant scopeを適用し新GRANT/FORCE/roleを作らない。

record/plan/pet/old ownerは認可済みscoped preflight readから、新ownerは検証したcommandから導く。新ownerの同clinic/record/petを非lock readで確認してからclaimし、Applyで全てlock下再検査する。clinical identityが不一致に変わったら業務write拒否。metadataと494 record anchorのrecord、command hash、Scope.intentとactionは一致必須。claim時に読んだbusiness情報を権限永続化とみなさない。

lookupはtrusted clinic/actorと既存6intent allowlistでbindingを特定し、保存metadataのrecordがURLと一致することを照合。metadataからplan/pet/旧新ownerを復元し現在record/owner viewとscopeを検査する。replaceは旧新双方。未成功にも同じ処理。planのsoft deleteやlink revokeは元metadataを破棄しない。現在ownerのpatientが変わり元対象の履歴認可を立証できなければnot-authorizedで非開示。認可anchorのmetadata欠落・不整合・解決不能は現Runner同様not-authorized/403で非開示として修復担当へ渡し、incoming targetやsuccess refsで補完・作成しない。結果照会で元write grantを要求せず、viewによる開示だけを判定する。元targetと現在開示認可が確定した後の結果読取失敗だけresult-unavailable/503とする。

## 6. 現在認可・整合・lock・同一TX

権限根拠と技術対応を分ける。既存D03 lifecycle権限表はカルテ内treatment PATCH/並替えPUTを `medical-records:edit` とし、518 §7は既存item変更の既存edit/deleteと、実施／結果変更のowner権限を分ける。routes_records.goのexaminations/vaccinationsのviewは実存するが、「owner本文不変のplan linkにもowner:edit必須」という既決条項は見つからなかった。owner本文不変の関連操作にowner:editを追加要求しない。

本契約の技術割当は **record view＋medical-records:edit＋該当examinations:view又はvaccinations:view＋現在edit fence**。link/replace/unlinkは共通plan関係を編集するだけで、実owner本文、実施、削除、解除、患者関係、証明書を変更できない。owner本文／状態を実際に変える別経路は引き続きそのownerのcreate/edit/delete/unconfirm等を必須にし、同時保存をこのlink-only grantで代行しない。実ownerの専用保護はgrantとは別に維持。unlinkはownerやplanを削除する操作ではないためowner/deleteも追加しない。master管理grant、新roleなし。lookup/readは現在record/owner viewだけを要求し、edit失権でもviewが残る本人は元結果を照会できる。

1. selected clinic固定→認証/前段scope→C06短claim（binding＋既存494 anchor＋§5.2のtarget metadataのみ）。Contentからrecord targetを導く。
2. Runner Gate→binding NOWAIT→現在認可付きreplay/unknown照会→savepoint→Apply。replay/lookupでplan/owner/dependencyの初期化・repairなし。
3. Applyはlifecycle→edit-right→record→関連plan ID昇順→plan_items dependency。expected aggregate/plan/fenceをlock下検証。baseline欠落はdependency-changed、494 acquire/takeover以外でINSERTしない。
4. owner lockは共通順で取得。検査のjob付きは関連job UUIDの固定順→exam ID昇順→revision/child ID昇順。接種はvaccination ID昇順。replaceの旧新ownerも同順。同jobのdetach/revertもowner lock前に同coordinatorへ参加する必要があり、既存job→exam順を逆にしない。事前readでjob候補を集め、lock後にJobIDを再検査し変化時はrollback/retryで再計画する。
5. current grants、同clinic/record/pet、active/deleted、plan kind、現在版／選択固定版／digest、既存uniqueを検査。検査はexam plan、接種はvaccine planだけ。recordが診療終了済みであることだけで旧draft gateを復活させない。D07公式版／死亡後専用処理等の既存保護を解除しない。
6. plan revisionをappend（link/replace=owner_link、unlink=owner_unlink）。replaceはsnapshotに旧新link双方と理由を記録。旧linkをrevoke、新linkに新revisionを結ぶ。plan.versionとplan_items aggregateを各1回増分。clinical_state、実owner.version、実結果本文は不変。
7. D07機器由来なら既存 `lab_import_usage_receipts` の `examination_detail` として、参照元結果を照合したclinical useをambient TXで保存する。既存model/lab_import.goが定義する種別を使い、本文変更していないのにmanual_mutationを偽記録しない。新しいuse_kind/enumを追加しない。job→examを取得済みのambient participantを正規trackerに設け、現在のRecordClinicalUseが開く別WithTxをネストさせない。receiptの失敗は関連もrollback。同一operation replayでは再挿入なし。旧usage receiptをunlinkで消さず、unlinkだけで機器Undoが許可されるとは扱わない。
8. execution_source変更に実存484sourceが依存する場合、authoritative最終contextと既存qualifying fields/digestを単一source UPDATEにまとめRETURNING generationを受領。replaceの旧新sourceを各1回。同一sourceなら1回。sourceのないlinkからtreatment/event/digestを生成しない。結果linkのみで484の意味が変わらなければ更新しない。
9. 関連＋revision＋aggregate＋必要usage/source generation＋audit＋成功C06 receiptを同TX commit。外部I/O、別WithTx service呼出しなし。既存owner create/updateのpostcommit通知／タグを呼ばない。

### 6.1 audit_logsのexact保存形

既存 `backend/internal/model/audit_log.go` / `audit_logs` を再利用し、新audit台帳を作らない。`actor_type='staff'`、actor_id=認証staff、clinic_id=trusted clinic、resource=`medical_record_plan_item`、resource_id=plan_item_id、created_at=server DB時刻。actionは新consumer定数 `medical_record_plan.link` / `medical_record_plan.link_replace` / `medical_record_plan.unlink` の3値（既存varchar50内）。既存実装にある定数と偽らず、このconsumerで追加するラベルと明記する。

old_value/new_valueはそれぞれ関連snapshotまたはNULL。exact keysは `link_id,plan_revision_id,owner_kind,owner_id,owner_reference,link_purpose,source_component_path,revoked`。linkではold=NULL/new=追加関係、replaceではold=旧active/new=新active（旧revokeはmetadataで辿る）、unlinkではold=旧active/new=同関係revoked=true。owner_referenceは§4のtyped参照のみ。実結果全文、料金、PDFは含めない。

metadataのexact keys: `schema_version:'1',operation_binding_id,operation_key,intent,record_id,pet_id,plan_item_id,previous_plan_revision_id,new_plan_revision_id,revoked_owner_link_id,created_owner_link_id,reason,command_sha256,before_versions,after_versions`。リンクID非該当はNULL、理由はlinkのみNULL、before_versions/after_versionsは `{plan_items:Version,plan_item:Version}`。ID/版はdecimal string、他は上述typed値。reasonは原入力を保存しhashだけで監査原文を捨てない。IP/UAは既存任意列の既存扱いを維持し今回新たに収集必須化しない。audit必須保存失敗は関連・成功receiptごとrollback。revoke列の四要素NULL整合とFK/RLSは518を保持する。§5.2のclaim metadata以外の新表なし。

## 7. 成功receipt・照会・エラー

flat `operation.ResultResponse` を維持する。`result_refs` exact keys:

- link: `record_id,plan_item_id,plan_revision_id,owner_link_id,owner_kind,owner_id`。
- replace: 上記＋`revoked_owner_link_id`。
- unlink: `record_id,plan_item_id,plan_revision_id,revoked_owner_link_id,owner_kind,owner_id`。

`committed_versions`は `plan_items,plan_item` の2key。実ownerを変更していないのでowner_versionをcommitted扱いで足さない。source generationは内部EffectSummaryと既存source receipt契約に従い、ここへ未定義可変keyを出さない。recorded_atは成功保存時刻。同key replayは元refs/versions/timeを返し最新表示を偽装しない。

HTTP mapping: succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。12状態を既存綴りのまま使う。not-succeededは成功ではない。非成功にrefs/versions/timeを出さない。owner target不在等はC06 validation、binding不在scope-not-foundと混同しない。前段認証失敗401は既存error envelope。

lookupは既存494 record anchorと§5.2の保存target metadataを照合し、現在viewで別患者・別医院へのreceipt開示を防ぐ。元plan/旧新ownerをScopeやfingerprintから復元できるとは扱わない。参照ownerの現在本文が失われても既存scoped歴史readerで元対象の現在認可が成立すれば保存receipt自体は照会できる。失権なら403でrefsを出さない。送信タイムアウト／parse失敗／500後は元key照会。結果未確定で新key・自動unlink・反対操作を行わない。

## 8. 読取projection・警告の閉包

既存plan GETのowner_linksへ同identity/refを返す。未読込や失権をempty links/resultsに潰さない。共通DTO統合は516単一owner。

guardの取得経路をFE推測にしない。既存 `GET /examinations/{id}` の成功DTOへ加算的に `plan_link_source` を追加し、exact fieldsを `examination_id,clinic_id,medical_record_id,pet_id,guard,references,component_paths` とする。guardは§4のunion、referencesは現在認可で閲覧できる同一examの実revision参照と、真にunversionedの場合のdigest参照だけ。component_pathsはD07が保証できる安定IDのみ、未提供なら空配列。NULL identityもそのまま返し、結果未取得やサービス失敗をNULL readinessに擬装しない。既存 `GET /vaccinations/{id}` へ `plan_link_source={vaccination_id,clinic_id,medical_record_id,pet_id,version}` を加算する。取得は各ownerの既存view認可で、リンク実行のedit認可とは別。新しい検索範囲や一覧全件取得は作らず、既存候補選択後のdetail取得を使う。両DTO追加とmutationが同版の機械契約／fixtureを持つまでFE独自guardを生成しない。

追加のread-only `link_readiness` はclosed enum `ready,identity_unresolved,identity_mismatch,owner_changed,reference_unavailable,owner_forbidden,owner_unavailable,plan_unavailable`。これは臨床状態・承認段階ではない。ready以外でリンクを送信しないが、他の予定や独立結果の閲覧を止めない。`owner_changed`は元固定参照を保持して再取得・人の確認へ。現在版と固定版を同じものと表示しない。

臨床stateは518のpending→not_performed、performed/deferred/cancelled同値、commercial→selectedを維持。result_stateは既存schemaを維持し、link_readinessで置換しない。結果ありかつnot_performed、performedかつ結果待ち、旧officialと作業版、意味不明／取得失敗は別表示。結果があるだけで実施や正常を断定しない。deferred/cancelledのプランにも過去事実の関連は保持できるが、実施へ復帰させない。removedプランへの新規linkは拒否し、既存関連訂正／unlinkの履歴経路は保持する。

warning receiptを新規承認トークンとして要求しない。関連訂正理由と現在版確認は本文・Contentに結び、診療終了warning_receiptは495/Bの別契約のまま。後日予定レポートは終了済みnormal recordを対象とし再開中を除外する既決条件を守る。

## 9. 必要subset別PRと未実装境界

1. **516 S1/S1a**: 既決共通3表＋application header/複合FK、model/ERD、RLS、plan read、coordinator・dependency・494 anchor。今回link consumer追加分として§5.2のtarget metadataを同schema ownerへ渡す。521/522の別schema/runtimeを奪わず共通model単一writer。これ自体を検査全機能待ちにしない。
2. **516 E-link**: 上記3検査route＋schema/fixture、固定版／unversioned digest reader、owner ambient lock参加、lab usage／detach guard参加、同identity既存examへのlink/replace/unlinkとFE。既存独立機器受信は停止しない。受信だけのrecord=NULLはidentity_unresolvedとして閲覧を維持。
3. **517 V-link**: 上記3接種route＋fixture、実版mandatory、owner ambient lock参加、link/replace/unlinkとFE。同じ実績への関連だけで証明書や次回日を更新しない。証明書内容訂正をこのPRに混ぜない。
4. **owner writer閉包**: 直接PATCH/delete/reparent、revision更新、job detach/revert、既存owner保存からの関連影響を同coordinator順に参加させる。link APIの本番有効化には当該ownerの閉包だけ必要。全領域Done待ちにしない。独立examにrecordを付ける操作は既存D07の適法な関連変更のみ再利用でき、revisionありの変更拒否を解除する新routeは本契約に含めない。対応不明結果の正しいrecord修復は既存目的別訂正担当へ渡し、リンクの便利機能で迂回しない。

本契約は「結果の新規作成＋実施」を一ボタンにする契約を追加しない。将来同じ明示操作で実owner保存と関連を原子的に行う場合も、同じcoordinatorのambient participantとして既存owner検証と副作用境界を閉じる必要がある。既存Createを呼んだ後にこのlink POSTを自動実行して同一成功と偽装しない。

実装受入fixture候補（全NOT_RUN）: 複数exam/一接種primary、same-key replay、differentpayload conflict、固定旧版と現在新版、legacy no fake version、NULLrecord受信、別患者/別医院/同名別予定、失権/旧fence、link後detach拒否、replace原子性/監査失敗、旧新source generation各1回、成功応答断→lookup、unlink後原結果/証明書/実施状態不変。実行・Docker・隔離環境・DB適用は今回行わない。

## 10. 別要件との境界

新MTGの健診PDFと狂犬病関連4文書はEMR-500/486/487側の別要件。本契約のlink-onlyは結果・実績identityを関連付けるだけで、自動PDF生成、OCR、証明書発行を追加しない。帳票が4文書になることと、同じ接種実績のprimary execution_source link数は無関係であり、§1と518 §5.3の既決制約は変更しない。文書作成ごとに同じ接種の主linkを4本作成しない。



### 8.1 owner_linksとgeneration読取のclosed必要subset

既存MedicalRecordPlanOwnerLink.idは既にlink IDであり削除/再採番しない。既存id/owner_kind/owner_id/owner_version/fixed_revision_id/statusの互換を維持し、required fields `plan_revision_id,link_purpose,source_component_path,owner_reference,fulfillment_sources` を加算する。plan_revision_idはlinkが作成された不変plan revisionの正IDで、現在plan headではない。link_purposeはexecution_source/result/charge_sourceの既存3値、source_component_pathはrequired nullable string（§4の安定IDのみ）。owner_referenceは§4の検査revision/legacy digest/接種versionの同じclosed unionであり、unknownをNULLや架空revisionにしない。current guardはowner detailのplan_link_sourceから取得し、owner_referenceと混同しない。

fulfillment_sourcesはrequired配列で、要素のexact keysは `{record_id,plan_item_id,owner_link_id,treatment_id}`（全canonical正ID）。authoritative関係が§6.8で影響させる実存484 treatment sourceだけを列挙し、tuple順・重複無し。無関係なら空配列、未読込/失権なら成功空配列にせずerror/既存unknown扱い。exam/vaccinationに新generationを設けず、結果linkからtreatmentを生成しない。履歴linkでも現在認可された同じ元sourceを辿れなければ推測しない。

成功/replay後は操作前後の旧新fulfillment_sourcesを別々に保持し、各tupleの既存484 GET /medical-records/{record_id}/plan-items/{plan_item_id}/fulfillment を再取得する。返却active owner_link/treatment identityがtupleと一致する場合だけ、そのfulfillment_generationを当該sourceの現在source_generationとして次の明示操作へ渡す。source変更/失権/unknownなら次writeを停止し再選択する。両sourceが同じtupleならGETは一回。これは現在値であり先のlink receiptのcommit時値ではない。内部EffectSummaryを公開値とせずcommitted_versionsの既存2keyを維持する。484のtreatment-only範囲をexam等の実施操作へ拡張しない。必要なsource関係resolverと現在認可を同じreader引渡しsubsetに含める。

非空fixtureのexact JSONは `{"id":"81","owner_kind":"examination","owner_id":"61","owner_version":"4","fixed_revision_id":"71","status":"active","plan_revision_id":"23","link_purpose":"result","source_component_path":null,"owner_reference":{"kind":"revision","revision_id":"71","revision_version":"4"},"fulfillment_sources":[]}`。対応するexam61/revision71(version4)、plan revision23が同clinic/record/petに存在するfixtureを併置する。replace/unlinkのexpected_link_revision_id='23'が現在plan.versionではなく当該link作成版から組み立つことを固定する。さらに旧新の異なるtreatment tupleを持つfixtureでは2回の484 GETと各返却generationを混ぜずに保持する。owner_reference/link_purposeを任意文字列を許すschemaにしない。no-source case、失権、同source重複、現在owner差替えもnegative fixtureへ含める。

reasonは原文保持、trimは非空判定だけ。既存JSON raw body 1MiB制限をparse/deep-copy/hash/claim前に当該routeへ適用し、誤Content-Typeのbinary上限へ逃がさない。内部callerも同上限、超過413既存ErrorResponse、切捨て無し。新しい500文字業務制限は追加しない。
