# EMR-495 B/C 採用契約：診療終了・目的別訂正・過去入力

2026-10-21。技術方式をselectedとして固定した単独で読める契約。runtime未実装・SQL未適用・全writer必要subset未接続・全試験NOT_RUN。新業務権限・理由要件は作らない。Aの複数診断readerは独立先行可。根拠main edf71515dbae11735a4184ff64c465e2ea51af88。

## 0. 適用範囲・既存構造・共通wire

medical_record_consultation_eventsの既存ended/end_undone、record単位revision、previous_event_id、source_kind=consultation/retrospectiveを再利用する。第二の終了台帳、旧status復活、reopened enumは作らない。medical_records.versionは子の完全保存版ではなく、inquiry dependency、clinical_plan版、diagnosis行版、plan_items版を独立保持する。medical_record_correction_linksは成功台帳であり、未確認proposalを入れない。

ID/revisionはcanonical正整数decimal stringかつsigned BIGINT以内。medical_records.versionなど実列INTEGERの値にはINTEGER上限も検査。JSON number/指数/負値/unknown fieldを拒否。医院・actor・recorded_by・confirmed_byは認証から決定する。通常編集/endへ新しい理由必須化なし。Cの理由は既存目的別訂正規則を継承する。

Cはbody operation_key、Bは既採用Idempotency-Keyのみ。二重要求しない。Cで単一recordの書込は494のedit_session_ref/edit_epoch/owner_expected_version、複数recordは下記record_fencesの型付き配列で対応する。Cookie/CSRFと現在権限を保持。新しいbearer、独立receipt方式は作らない。

### 0.1 B0 consultation-state read

GET /api/v1/medical-records/{id}/consultation-state のclosed responseは `{medical_record_id,record_version,consultation_head,effective_end,evidence_state}`。headはNULL又は `{event_id,revision,event_kind,source_kind,occurred_at,previous_event_id}`。event_kind=ended/end_undone、source_kind=consultation/retrospective、previous_event_idはNULL又はID。effective_endはboolean又はNULL、evidence_stateはknown/unknown。日時はRFC3339。

同clinic/recordの最大revisionをheadとする。履歴なしを過去未終了と断定しない。legacy証拠欠落、現在予約進行とevent不整合はunknown。retrospectiveを当日受付進行の証拠にしない。GETは版bootstrapやGetOrCreateを行わない。現在record viewが必要。通常owner PATCHは旧finalizedだけで拒否しないが、原snapshot/目的別保護/監査は維持。494のheader/inquiry/clinical-plan/children全writer参加前に全体切替完了としない。

### 0.2 C0/C1/C2の完全routeとbody

| method/path | intent / 用途 |
| --- | --- |
| POST /api/v1/medical-records/{id}/correction-proposals | medical_record.correction_propose |
| GET /api/v1/medical-records/{id}/correction-proposals/{proposal_id} | 保存案read |
| POST /api/v1/medical-records/{id}/correction-proposals/{proposal_id}/confirm | medical_record.correction_confirm |
| GET /api/v1/medical-records/{id}/operations/{operation_key}?intent=medical_record.correction_propose 又は medical_record.correction_confirm | 既存共通lookupへallowlist追加 |
| POST /api/v1/pets/{id}/retrospective-record-proposals | medical_record.retrospective_propose |
| GET /api/v1/pets/{id}/retrospective-record-proposals/{proposal_id} | source-less保存案read |
| POST /api/v1/pets/{id}/retrospective-record-proposals/{proposal_id}/confirm | medical_record.retrospective_confirm |
| GET /api/v1/pets/{id}/medical-record-operations/{key}?intent=medical_record.retrospective_propose 又は medical_record.retrospective_confirm | 未生成record用lookup |

query intentは上記完全名の列挙のみ。既存package等のlookup intentを置換しない。

C0 POST exact body: `operation_key,intent,source_revision,destination_record_id,clinical_date,reason,selection,source_dependency_versions,destination_dependency_versions,record_fences,supersedes_id`。intent=patient_correction/past_visit/postmortem_prior_care。clinical_date=YYYY-MM-DD。supersedes_idはNULL又は同clinic/同対象の旧proposal。sourceはpath、target_petはdestinationの認可済みserver readから導く。selected_payloadの任意本文をclientから受け取らずserverが選択元版から抽出する。

C1 confirm exact body: `operation_key,proposal_revision,proposal_digest,confirmed,record_fences,source_dependency_versions,destination_dependency_versions`。proposal IDはpathに一つだけ置き、body duplicateを受けない。confirmed=true。元先target/selectionは保存proposalから読む。

record_fencesはrecord ID昇順、重複なしの `{record_id,edit_session_ref,edit_epoch,owner_expected_version}` 配列。owner_expected_versionは当該record.version。元先が同じなら1要素。C0/C1の新writeは元先の対象履歴への参加も含めその現在fenceを照合する。source readだけの一般readerにfenceを要求する意味ではない。

source/destination_dependency_versionsのclosed shapeは `{record_id,record_version,inquiry_revision,clinical_plan,diagnoses}`。inquiry_revisionは該当namespace正revision又はNULL、clinical_planはNULL又は `{id,version}`、diagnosesはID昇順の `{id,version}` 配列。変更owner/選択元に必要な版はNULL不可。子集合を全件確認し追加・削除も比較する。未初期化を版1として補完しない。選択元snapshot版はcommandにも固定し、単なる上書きCASと分離する。

C2 propose exact body: `operation_key,intent,clinical_date,reason,documented_content,expected_pet_version,expected_existing_target,supersedes_id`。intent=past_visit/postmortem_prior_care。expected_pet_versionは実存pets.version（INTEGER）CAS。expected_existing_targetは第一sliceではrequired NULL固定。既存recordを選ぶ場合は本source-less案を保存せず、通常owner編集又は実存sourceからのC0/C1へ操作種別を分ける。serverは現在pet/lifecycleと同clinic/petの既存record集合を取得し、提案のdependency_snapshotにimmutable保存する。confirmedなしの案保存と臨床record生成を分ける。

C2 confirm exact body: `operation_key,proposal_revision,proposal_digest,confirmed,expected_pet_version`。target petと内容はpath/保存案から、confirmed=true。元recordも新recordもまだないのでedit-sessionを捏造しない。pet版と保存案の既存record集合・既存331 lifecycle条件を同TX再照合する。集合は各recordのid/version/dateをID昇順に固定し、追加/消失/改訂時はdependency-changed。日付uniqueを新設しない。

proposal GET exact response: `{proposal_id,proposal_revision,proposal_digest,intent,source_record_id,source_revision,destination_record_id,target_pet_id,clinical_date,reason,selection,selected_payload,dependency_snapshot,proposed_by,recorded_at,consumption}`。未該当値はNULL。consumptionはNULL又は `{correction_link_id,destination_record_id}` であり、可変statusを作らない。現在元先/petのscope/viewと本文閲覧認可が必要。失権は本文を出さない。source-lessのselectionは空配列、selected_payloadはdocumented_contentの検証済み値。proposal_digestは保存command canonical hash。

### 0.3 C1/C2の許可本文と非対象

selection要素は `{owner_kind,source_row_id,source_revision,fields}`。owner_kindはinquiry/clinical_plan/diagnosisのみ。source_row_idはinquiryのみNULL（path recordで識別）、clinical_plan/diagnosisは安定ID必須。source_revisionはそれぞれinquiry namespace revision／plan版／診断行版である。inquiryはchief_complaint/chief_complaint_type_id/notes、clinical_planはphysical_exam/diagnosis_details/treatment_policyのfield列挙のみ。diagnosisは安定row IDと行版で行全体選択しfields=['row']。任意JSON path、旧単一診断field、任意ownerを拒否する。

診断行のclosed保存値は `{diagnosis_slot,diagnosis_type_id,diagnosis_name_id,diagnosis_name_snapshot,category_name_snapshot,sort_order}`。slotは既存DDLの1/2、各slotに複数行可。nullable master参照/当時ラベルを保持し現在masterから架空補完しない。転記先row IDは新規、元row IDを流用しない。inquiry IDは明示NULL、文字列clearは空文字、null可否は既存owner契約どおり。

C2 documented_contentは `{inquiry,clinical_plan,diagnoses}`。前二者は上記許可fieldのclosed object、diagnosesは上記行値配列。省略fieldは未選択、明示NULLと区別する。C1は選択済み元本文だけを転記し、新任意本文の編集は後の通常PATCHへ分離する。飼主LINE原文、実行済治療、検査、接種、画像、会計、在庫、501他院結果、plan_itemsは今回コピーしない。Aの複数診断writer contractを必要subsetとして受領し、現legacy clinical_plan_requestの単一診断fieldを正解としない。

## 1. 不変commandとC06の接続方式を選定

digestからpayloadは再構成できない。owner adapterはrequestをclosed typed commandへ検証・deep-copyし、外から変更できないprivate valueとして保持する。canonical JSONは518と同じJCS規約（UTF16 key順、array順保持、Unicode非正規化、duplicate/lone surrogate/U+0000拒否）。SHA256 lowercase hex64をcommand_sha256とする。

Content.Fieldsの正確なkey集合は schema_version='1', command_sha256, target_kind, target_ref, intent。target_kindはappointment/record/pet、target_refはURLの正ID、intentは下表のscope intentと一致。confirmだけ proposal_id と proposal_revision を追加する。省略keyの空文字補完は禁止。ContentはNewContentで作る。

Applyはimmutable commandを再serializeしhashを再計算してreq.Content.Fieldsのcommand_sha256とconstant-time比較し、schema/target/intent/proposal値も一致検査。不一致はvalidation、writeゼロ。保存proposalを使うconfirmでは現在認可後にimmutable proposalを読み、command内のproposal_digestと再計算値も一致させる。mutable request/別lookup結果を無検査で捕捉しない。C06自体のfingerprint生成は変更しない。

commandに入れるのは業務対象・内容・理由・了承したsnapshot。Preconditionsの正確なkeysは record.{id}.edit_session_ref, record.{id}.edit_epoch, record.{id}.version, inquiry.{record_id}.revision, clinical_plan.{plan_id}.version, diagnosis.{row_id}.version, appointment.{id}.version, consultation.{record_id}.head_id, consultation.{record_id}.revision。headなしだけ最後2値はliteral 'none'。必要ownerだけ入り、欠落required key拒否。確認した転記元版/plan snapshotはcommandにも入り、書込CAS-only更新と区別する。

## 2. Bの新専用route・headerを一方式に固定

POST /api/v1/reservations/{id}/end-consultation と /undo-consultation-end は**新専用routeとしてflat C06のみ**を公開する。既存汎用PATCH /reservations/{id}のresponse shapeは変更しない。予約設計で選定済みのIdempotency-Key（canonical lowercase非zero UUID）を新routeでも使用し、body operation_keyは拒否。旧operation_result wrapperを新routeには出さない。これで既公開API互換を維持し、新URLの機械契約をC06へ閉じる。共有statusライブラリの変更なし。

scope owner_kind='reservation'; intentはreservation.consultation_end / reservation.consultation_end_undo。lookup GET /api/v1/reservation-operations/{key}?intent=上記完全名。元appointment＋recordをclaim metadataへ固定し、現在リクエストtargetから元targetを推定しない。

end bodyの全key: expected_version, confirmed(true), record_ref, expected_consultation_head, plan_dependency_versions, warning_receipt。
undo bodyの全key: expected_version, confirmed(true), record_ref, expected_consultation_head, ended_event_id, reason。
record_ref={owner:'medical_record',kind:'consultation',id,revision}。expected_consultation_head=nullまたは{event_id,revision}。

end commandはappointment_id,record_id,confirmed,acknowledged_plan_snapshot。acknowledged_plan_snapshotのexact shapeは{plan_dependency_versions,warning_receipt}で、以下の二つのclosed DTOをそのまま含む。undo commandはappointment_id,record_id,ended_event_id,reason,confirmed。expected_version/header版/head CASはPreconditions。別ended_eventのUndoは同内容でないため新command。

### 2.1 516/518の終了警告subset

518のGET plan-itemsのcollection_version/version_stateとitems[].id/version/executionをそのまま参照する。新clinical_plan診断revisionを予定版に代用しない。

plan_dependency_versionsのclosed shape:
{record_id, collection_version, items:[{id,version,execution_revision_id,execution_kind,execution_state}]}
collection_versionはplan_items namespaceの正revision、version_state=readyの取得結果だけ利用可。itemsは全件取得した同recordの集合、ID昇順、重複不可。execution_kind=clinical/commercial。clinical stateはnot_performed/performed/cancelled/deferred、commercialはselectedのみ。revision_idは518 execution.revision_id。DB pending→wire not_performedを保持。commercialを未実施警告へ数えず、clinical not_performed/deferredを警告対象とする。null/uninitialized/unresolved・不完全取得は「全件実施済み」にしない。全件0はreadyなcollectionだけ空を認める。

warning_receiptはnullまたは{schema_version:'1',snapshot_sha256,acknowledged_item_ids:decimalString[]}。snapshot_sha256は上のcanonical snapshotのSHA256、item IDsは警告対象全件を昇順列挙。警告なしはnull、ありは一致必須。これは認可token/証明署名ではなく**確認対象のecho**であり、serverが現在plan集合・版・状態を同TX再計算し一致する時だけ確認を認める。clientだけのhashやconfirmedで現在依存検査を免除しない。scope/対象/確認snapshotはcommandに結合する。変化ならdependency-changed、入力保持・再表示・再確認。

Bは本文変更なし、既存reservations:editとrecord view。現在edit-right rowで直列化するが、本文を編集しない予約終了に新たな編集許可要件を作らない。ended/end_undone追加、予約遷移、必須audit、receiptを同TX。通常編集・自動保存成功とは別。Undoは既存billing finalization history=neverのみ。event時刻はserver、旧status再導入なし。

### 2.2 B target保存・TX・再開

新B routeのclaim targetを保存する補助metadata `reservation_consultation_operation_targets` を使用する。columnsはclinic_id/binding_id/appointment_id/medical_record_id/pet_id（全bigint NOT NULL、各ID正）、command_sha256（lowercase64hex）とする。PK(clinic_id,binding_id)、唯一FK(binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT。business FKなし、INSERT後UPDATE/DELETE不可、既存tenant RLS。新binding作成時だけ同短claim TXでINSERTする。元targetは認可済み予約とrecordの明示対応から導き、lookupでincoming recordを補完に使わない。metadata不在は結果非開示で照合担当へ、GETで生成しない。

claim終了→Runner Gate→binding NOWAIT→現在認可付きreplay/unknown→savepoint→予約lifecycle→record edit-right row→record/plan/依存ownerの既存順。現在clinic/actor/予約edit/record view、予約record/pet対応、appointment版、record版、event head、警告集合を再検査。本文fenceを持たないスタッフに本文writeを行わせない。

endはin_consultation→accountingとended revision+1を同TX。undoは会計確定履歴neverをlock下確認し、対象endedをheadとしてend_undoneへ連鎖、予約はin_consultationへ戻す。返金済み/取消/残額0をneverとみなさない。勤務外を新undo禁止条件にせず既存の継続前手動担当修正を維持。初回event revision=1、以降head+1、同key replayは追加ゼロ。event/予約版増分/必須audit/receiptを一体commit。occurred_at/recorded_atはserver時刻。新たな再開recordを作らない。

reservation lookupは保存metadataの予約/record/petを現在scopeと予約edit/record viewで再検査する。保存成功は元receipt、現在進行はB0等の別readで示す。新routeのflat C06採用を旧PATCHのresponse変更や調整待ちにしない。

## 3. Cproposal保存と消費の物理採用契約

新表medical_record_correction_proposals（未作成）:
- id BIGSERIAL PK; clinic_id bigint NOT NULL FK clinics(id) RESTRICT; UNIQUE(id,clinic_id)
- proposal_revision bigint NOT NULL CHECK=1。不変案の改訂は新IDであり、既存臨床revisionとは別。
- intent text NOT NULL CHECK IN(patient_correction,past_visit,postmortem_prior_care)
- target_pet_id bigint NOT NULL; FK(clinic_id,target_pet_id)→pets(clinic_id,id) RESTRICT
- source_record_id bigint NULL; source_revision bigint NULL; destination_record_id bigint NULL
- source pair CHECK: 両NULL、または両非NULLかつsource_revision>0
- FK(source_record_id,clinic_id)→medical_records(id,clinic_id) RESTRICT、FK(destination_record_id,clinic_id)→medical_records(id,clinic_id) RESTRICT
- clinical_date date NOT NULL; reason text NOT NULL CHECK btrim<>''
- command_json jsonb NOT NULL CHECK object; command_sha256 bytea NOT NULL CHECK octet_length=32
- dependency_snapshot jsonb NOT NULL CHECK object。owner版のtyped集合で任意keyを受理しない。
- proposed_by bigint NOT NULL FK staffs(id) RESTRICT **単列FK**
- proposal_operation_id bigint NOT NULL; FK(proposal_operation_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT; UNIQUE(clinic_id,proposal_operation_id)
- supersedes_id bigint NULL; FK(supersedes_id,clinic_id)→同表(id,clinic_id) RESTRICT。置換案も旧案の履歴を消さない。
- recorded_at timestamptz NOT NULL DEFAULT clock_timestamp()

全列UPDATE/DELETE禁止のimmutable trigger（移行専用手続きの勝手な例外なし）。DB triggerでJSON/hash自己整合を検査するかは実装選定でなく、writerが必ず同一canonical bytesから両方を生成し、read/confirmでも再hashする契約を固定。UPDATE権限を通常roleへ付けない。INSERT/SELECTは既存app_private.apply_rls_policyのtenant_clinic_id_isolation、USING/WITH CHECK app_private.has_clinic_access(clinic_id)。RLSを現在grant代用にしない。

staffはグローバル主体なのでstaffs(id,clinic_id)複合FKを新設しない。current staff→account、staff_clinic_assignmentsの当該clinicかつdeleted_at IS NULL、既存grant、既存確認医師policyを業務TXで再検査。過去提案者が所属を失ってもFKを削除して履歴を消さない。confirmed_by/recorded_byも既存単列staff FKのまま、現在所属は実行時チェック。

correction_linksへproposal_id bigint NULL（既存履歴維持）を追加する採用設計。新CconfirmはNOT NULLをwriterと専用guardで必須化。FK(proposal_id,clinic_id)→proposal(id,clinic_id)、UNIQUE(clinic_id,proposal_id)（NULL複数は旧履歴のみ）。同一proposalのconfirm並行2keyはproposal row lock→既存link照会、一方だけ消費。既存linkが別keyで存在すればdependency-changed、既存成功refを新key成功として複製しない。旧key照会を案内する安全なreadは別途current認可付き。

C0は**proposal作成/read/lookupまで**。C0でconfirm APIを有効にしない。C1/C2がconfirmを提供し、医師確認＋実際の転記/新規＋correction_link＋audit＋receiptを一つの成功TXで行う。「確認だけ成功→後で転記」の第二confirm/第二成功receiptは作らない。

## 4. Cのexact immutable command / scope / lookup

scope owner_kind=medical_record。
C0 intent=medical_record.correction_propose。command={schema_version:'1',intent,source_record_id,source_revision,destination_record_id,target_pet_id,clinical_date,reason,selection,selected_payload,source_owner_versions,destination_owner_versions,supersedes_id}。NULLも明示、unknown key拒否。target_kind=record、target_ref=source ID。
C1 intent=medical_record.correction_confirm。command={schema_version:'1',proposal_id,proposal_revision,proposal_digest,source_record_id,destination_record_id,target_pet_id,confirmed:true}。具体selected_payloadは保存proposalから読み、再hash一致後のみ適用。proposal_digestは確認snapshotに結合するためContent hash対象。
C2 intents=medical_record.retrospective_propose / medical_record.retrospective_confirm。target_kind=pet、target_ref=対象pet。source pairはNULL、新規先recordはまだNULL。confirm commandは{schema_version:'1',proposal_id,proposal_revision,proposal_digest,target_pet_id,confirmed:true}（未生成record IDを想像しない）。

C2 propose command exact shapeは `{schema_version:'1',intent,source_record_id:null,source_revision:null,destination_record_id:null,target_pet_id,clinical_date,reason,selection:[],selected_payload,source_owner_versions:null,destination_owner_versions:null,pet_snapshot,existing_record_snapshot,supersedes_id}`。selected_payloadは検証済みdocumented_content、pet_snapshotは `{pet_id,version}`、existing_record_snapshotは§0.2の全件集合。死亡後生前入力の既存331 policy判定もproposalのdependency_snapshotへserver側で保持し、confirmで同policyを再評価する。新たな臨床判定enum/条件をこのsnapshotに作らない。expected_pet_versionはPreconditionsの `pet.{id}.version` に加えるが、確認した元pet_snapshotはcommandにも固定される。

C0 proposeも既存record内容を公開/保存するのでsource+destinationの現在scope/viewと提案に必要な既存editを照合。confirmは書込grantと既存確認医師policy。医師認可を提案者booleanや指定doctor IDから推測しない。

新metadata表medical_record_correction_operation_targets:
clinic_id bigint, binding_id bigint, root_kind text CHECK record|pet, root_ref bigint>0, source_record_id bigint NULL, destination_record_id bigint NULL, target_pet_id bigint NOT NULL, proposal_id bigint NULL。PK(clinic_id,binding_id)。唯一FKは(binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT。record/pet/proposal/clinic直接FKは置かずclaim時の業務row暗黙lockを避ける。root_kind=recordならroot_ref=source、source非NULL。root_kind=petならroot_ref=target_pet、source/destinationはNULL（新規生成前）。全値はContentとimmutablecommandの整合検査済み値からderive。新bindingだけ同短claimTXでINSERT、変更/後付け禁止。既存494 single-record anchorは変更しない。

lookupはbindingに固定されたmetadataを唯一のtarget源とし、URL record/pet一致と同clinic/actor/intentを検査。record型は元先両record＋petの現在scope/viewと元操作grant、pet型は現pet scope＋元create/edit/確認grantを検査。成功後の新recordもreceiptのdestination_record_idと現在実record/pet/clinicを照合してから結果を返す。未成功pet型はrecord未生成でも現在pet認可で照会可。必要metadata欠落・元先不明/アクセス失効なら結果非開示、scope漏出なし。current request targetで欠落anchor補完しない。

unknown attemptはC06ルールどおり再実行停止。lookupでnot-succeededを確認しても自動新keyや別患者へ切替しない。propose再送は元proposalだけ、confirm再送は元link/recordだけ。別actorは元操作のkeyを代理照会しない。別医師の確認は自分のscopeのconfirmとして既存proposalを認可readする。

### 4.1 C1の実writeと同TX条件

source_revisionはsource medical_records.versionの確認値であり、子の完全snapshotを意味しない。子版/selection/元先row対応をtransferred_fieldsのschema_version=1へ保存する。取得できない過去snapshotを現在masterから再生成しない。患者取り違えではsourceのpet/owner/date/bodyを変更しない。destinationは明示選択した同clinicの既存recordで、target_pet/実診療日/元先対応を照合する。

選択した本文だけをdestinationへ反映し、非選択fieldと既存診断行は消さない。診断は新IDで追加し元row対応を保持。全対象現在認可/既存医師確認→編集権row ID順→record ID順→owner/dependency/子行安定順→proposal lock/未消費確認→選択元/先の版集合照合→destination ownerのambient write→所定owner版増分→correction_link＋proposal一度消費＋audit＋receiptを同TX。outer Runner Gate/binding/unknown/savepointは省かない。proposal等を先に独立TXで消費してから転記しない。監査やlink失敗は全rollback。

依存の改訂/追加/削除、死亡状態変化、保護会計の関連変化、grant失効を旧確認内容で通さない。元snapshotの意味が変われば再確認案、新command。CASだけの拒否と内容変化を区別する。訂正影響readから受付/新請求/返金/実在庫/実行済み治療を自動起動しない。


### C0/C2 proposeの耐久commandと同key再開

bodyのsource_dependency_versions/destination_dependency_versionsは、了承した選択snapshotとしてcommandのsource_owner_versions/destination_owner_versionsへ写像する。単なる現在実行CASはPreconditionsにのみ置く。C1では固定proposal_digestから了承snapshotを復元し、保存された選択元/先と固定版を検証する。受信dependency_versionsは現在実行CASとして別途照合し、それで保存snapshotを置換しない。現在fence/CASをcommandへ混入しない。了承snapshotの変更は別内容、CASだけの更新は同keyである。

owner-private新表 `medical_record_correction_commands` を採用する。columnsは `clinic_id bigint NOT NULL, binding_id bigint NOT NULL, schema_version smallint NOT NULL CHECK (=1), request_sha256 char(64) NOT NULL, command_sha256 char(64) NOT NULL, canonical_request text NOT NULL, canonical_command text NOT NULL, recorded_at timestamptz NOT NULL DEFAULT now()`。PK(clinic_id,binding_id)、唯一FK(binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT。両hashはlowercase64hex。canonical列は§1 JCSのUTF-8全文で、重複key/不正Unicode/NULL文字を拒否し再canonical化一致と各hash一致を保存前・読取後に検証する。本文が必要なowner専用保存であり、C06短Content・EffectSummary・receipt・ログにPII本文を入れない。tenant RLS、INSERT後UPDATE/DELETE不可、business FKなし。private repositoryだけから現在元targetの本文閲覧認可後に読む。DB保存可を一般公開APIの本文開示許可としない。

canonical_requestのC0 exact keysは `{schema_version:'1',intent,source_record_id,source_revision,destination_record_id,clinical_date,reason,selection,supersedes_id}`。source_record_idはpath、その他は検証済みbody。C2 exact keysは `{schema_version:'1',intent,target_pet_id,clinical_date,reason,documented_content,expected_existing_target:null,supersedes_id}`。target_pet_idはpath。bodyのrecord_fences/dependency_versions/expected_pet_versionは再開時の実行前提なのでcanonical_requestへ入れず、元に捕捉した了承snapshotはcanonical_commandに保持する。source_revision/selection内の固定版を変えればrequest hashも変わる。再開時のdependency更新によって保存commandのsource_owner_versions/destination_owner_versions/pet_snapshotを上書きしない。固定snapshotが今も適用可能かはApplyで検査し、snapshot不一致をCAS-only更新で通さない。canonical_commandは§4のexact shapeを保存する。内部で捕捉するVerifiedCommand等のmemory値はこの耐久保管の代替ではない。共通Runnerは変更せずconsumerのclaim hookとprivate repositoryが担う。

新keyは現在認可付き読取で元snapshotを捕捉し、commandを検証してからbinding・target metadata・owner-private commandを同じ短claim TXにINSERTする。claimは業務本文へFK/row lockを取りに行かず、Applyで通常lock順に現在版/対象/保存commandを再照合する。snapshot取得後にsourceが変わればdependency-changed、業務writeゼロ。command保管だけを提案保存成功としない。owner-private保管失敗はbindingもrollbackする。

既存keyは保存targetで現在認可を先に検査し、保存canonical_requestとincomingの意味的入力を照合する。CAS-only更新は許すが意味的変更はoperation-key-conflict。保存commandを再hashしてC06 Contentを再構成し、現在sourceから別snapshotを採らない。成功済みなら元proposal/receiptをreplayする。未成功なら同じ耐久commandでApplyを再試行し、現在preconditionsをlock下検査する。依存snapshotが変わった場合は元操作結果を確定させた上で新案の再確認へ進み、unknownのまま新keyで迂回しない。command欠落/破損はresult-unavailableか既存invariant errorで停止し、GET修復/現在本文補完は禁止。

raw requestとserver抽出後canonical_commandの各1MiB上限をclaim前に検査する（既存JSON上限と同じ技術上限）。超過は既存413 ErrorResponse、claim/writeゼロ、切捨て無し。内部callerも同上限。snapshotの一部を無言で省いて小さくしない。DDL/model/ERD/private repository/claim hookと再起動後replay fixtureがC0/C2 proposeの必須引渡しsubsetである。

### 既存先・baseline・実来院factの境界

C2第一sliceのexpected_existing_targetはrequired NULL固定。non-NULLは案保存前にvalidationとし、既存recordへの任意documented_contentをC1へ流さない。既存recordの新任意本文は通常owner編集、実存sourceからの選択転記だけがC0/C1。画面入力は保持し架空proposalを作らない。

診察終了のbaseline欠落はdependency-changed。第一sliceは既存494 acquire/takeoverで編集権限のある担当が準備済みのrecordに限定する。受付担当へrecord editを追加せず、GETやend操作でbaselineを生成しない。未準備recordは「編集担当による準備未完了」と表示し、全legacy recordの終了が実装可能/対応済みとは記載しない。後続のbaseline整備は518の必要subsetとして別に残す。

destination clinical_plan=NULLは選択先ownerの明示不存在CASとする。record/fence lock下で不存在を再確認し、owner既存初期値＋選択field＋初期版＋audit＋receiptを同TX作成する。source選択ownerのNULLは拒否する。GETで版1を補完せず、未選択fieldの上書きや診断空集合との混同をしない。

511 §6の適格条件を満たす明示過去来院確認と患者/日付訂正は、correction_link_id・確認実日・当時owner/pet・根拠版・同bindingを既存actual_visit_factsへ、訂正とreceiptの同outer TXで追記する。source kind/原fact参照は511の既存closed契約へ写像し、元factを保持して訂正chainで解決する。単なる転記・外部資料・医師確認を来院事実へ変換せず、終了eventとも区別する。このwriter接続を必要subsetに含め、失敗は訂正全体rollback。

## 5. source-less guardの全条件

correction_linksのsource_record_id/source_revisionをNULL可にするが、以下をDB CHECKでANDする:
1. sourceの両NULLまたは両非NULL＋revision>0。片NULL禁止。
2. intent=patient_correctionまたはaddendumでは両非NULL。
3. source両NULLならintent IN(past_visit,postmortem_prior_care)、destination_record_id非NULL、clinical_date非NULL、proposal_id非NULL。
4. 全case reason trim非空、operation_id/recorded_by非NULL（既存維持）。新Cconfirmはconfirmed_by非NULL。
5. sourceなしのtransferred_fields.schema_version=1、source_kind='new_retrospective'を専用guardで強制。sourceありはsource_kind='existing_record'。

同TXのguard/participantでproposal intent/date/reason/target_pet/source pair一致、destination.pet_id=target_pet、全record同clinic、operation_binding owner/intent/actor、confirm actorとconfirmed_by、proposal未消費を照合。source-lessはdestinationが当該操作で生成されたことを作成participantの戻り値と同TX行で確認し、他の既存recordをNULLsourceの新規として偽装しない。既存recordへの追記はsourceありC1へ。

新record=version1、date実診療日、appointment_id NULL、entered_by認証actor。元source=self禁止。pet単位既存lifecycle直列化、確認済み既存対象集合を再検査し、同日候補が増えた場合は再選択。日付unique新設なし。死亡flag維持・生前事実確認は331既存policy。時点不明を生前確定にしない。

新recordがまだない間のeditfenceは存在しないため、pet/create current認可と直列化で守り、新record通常編集開始時に494 acquire。既存recordを使う場合はそのcurrentfenceを必須にする。

新record作成成功だけでretrospective ended eventを自動生成しない。過去の診療終了自体を記録する追加内容は別の明示操作契約が必要で、このC2 bodyへ暗黙に足さない。過去入力が成功したことと終了済みレポートへの掲載条件は別である。受付/予約を架空作成せず、今日の日付やsource=selfでsource-lessを偽装しない。

## 6. audit / receiptの正確なkeys

必須audit_logsは既存表を再利用。actor_type='staff', actor_id=認証staff、clinic_id=trustedclinic、resource='medical_record'、resource_id=対象record（source-less proposeのみpet対象なのでresource='pet', resource_id=pet）。actionは下記:
- medical_record.consultation_end
- medical_record.consultation_end_undo
- medical_record.correction_propose
- medical_record.correction_confirm
- medical_record.retrospective_propose
- medical_record.retrospective_confirm

metadataの全key: schema_version:'1', operation_binding_id, operation_key, intent, proposal_id(null可), source_record_id(null可), destination_record_id(null可), source_revision(null可), confirmed_by(null可), reason(null可), changed_fields（限定field path配列）, before_versions, after_versions。旧/new_valueは変更ownerの許可subsetだけ（選択本文の前後）、end/proposeは本文を複製せずNULL。IP/UAをこの設計で新たに必須収集しない。audit書込失敗は業務TX rollback。

result_refs / committed_versionsのexact集合:
- B end/undo refs={appointment_id,medical_record_id,consultation_event_id}; versions={appointment,consultation_event}。値は全てdecimal strings。
- C propose refs={proposal_id}; versions={proposal:'1'}。
- C1 confirm refs={proposal_id,source_record_id,destination_record_id,correction_link_id}; versionsはdestination_recordを必須、実変更ownerだけinquiry/clinical_planを追加。複数診断行版を単一版へ潰さず必要なら診断後readで取得し、未定義dynamic receipt keyを作らない。
- C2 confirm refs={proposal_id,destination_record_id,correction_link_id}; versions={destination_record:'1'}＋初期作成したinquiry/clinical_planの版だけ。

既存record.versionを実変更で増分しないowner実装を使う場合、destination_recordは変更後の観測版と明記し、子更新はそのowner版がCAS根拠。receiptを子全体snapshotへ拡張しない。変更なしのownerは版増分なし。sourceの元本文/元患者は不変。

12 statusとHTTPは全新route共通:
succeeded200 / replayed200 / not-succeeded200 / operation-in-progress202 / result-unavailable503 / operation-key-conflict409 / scope-not-found404 / not-authorized403 / validation400 / version-conflict409 / dependency-changed409 / edit-right-lost409。
非成功はstatus,operation_keyだけ。成功だけrefs/versions/recorded_atを付ける。entry JSON/authエラーは既存ErrorResponse。既存12以外をC06 enumへ足さない。raw fingerprint/command/patient本文はエラーやreceiptに出さない。

## 7. 実装PRの閉包

A reader先行。B0 read→B1専用end/undo。C0 proposal persistence/read/lookup（confirmなし）→C1既存recordの限定転記confirm→C2source-lesscreateconfirm。C0の成功は「案保存」であり、カルテ訂正完了とは表示しない。C1/C2のみ業務変更完了。共通api.yamlと489metadata hookの編集ownerを一名にし、各slice必要subsetで渡す。どの段階も試験実行NOT_RUN、隔離環境なし。

根拠追補: backend/internal/operation/{contract,dto}.go、backend/docs/fixtures/emr518/plan-list.response.json、plan-detail.response.json、docs/work/emr/emr-518-package-api-contract-20261020.md、001_init.sqlのstaff_clinic_assignments / app_private.apply_rls_policy。根拠一覧は本書末尾へ統合。これらを実装済みB/Cの証拠にしない。

### 保存guardの補足

proposal自体のshapeも閉じる。patient_correctionはsource/destination両record必須、両recordが異なること、target_pet=destination.petを照合する。past_visit/postmortem_prior_careのsource-less proposalはsource pairとdestinationともNULL、target_pet必須。既存record向けはsource/destination両方が明示済みで、元先が同一か別かに関係なく元snapshotと目的を照合する。同一recordの後日追記は原患者付替えではなく既存record対象の専用訂正として扱う。

JSON guardの必須keyが欠落/JSON null/型不正のときSQL CHECKのUNKNOWNで通さない。schema_version/source_kind等の条件全体はIS TRUE相当で真のみ許可する。DB proposal/link間照合は同clinic対象をlockして既存C06業務TX内で行い、JSONだけの自己申告に依存しない。

追加claim metadata表も既存tenant RLS（has_clinic_accessのUSING/WITH CHECK）とINSERT後UPDATE/DELETE不可を適用する。単一FK(binding_id,clinic_id)以外の業務FKは付けない。proposal表をclaim時にFK参照して共通gateより前にproposal業務lockを取ることも避け、後段で実proposal/record/petを再照合する。

## 8. 根拠と公開境界

- https://app.plane.so/baritechllc/browse/EMR-495/ の本文・全3コメント、494 handoffとA/B/C境界。
- docs/work/emr/emr-494-edit-draft-contract-20261020.md §§2–6（wire/claim/current auth/inquiry CAS）。
- docs/spec/domains/medical-records/lifecycle.md（screen-consultation-end、agreement-first-five、empty-delete-reviewed-correction、retrospective-no-card-no-billing、通常編集/自動保存）。
- docs/design/domains/medical-records/technical-spec.md（CL-R05/doctor-confirm-gate/owner impact）、docs/design/domains/reservations/technical-spec.md §§4.5/4.7/5.1。
- backend/migrations/001_init.sql の medical_records / medical_record_dependency_versions / medical_record_correction_links / medical_record_consultation_events / clinical_plan_diagnoses。
- backend/internal/operation/{contract,dto,store}.go、backend/internal/model/{audit_log,pet}.go、backend/internal/medicalrecord/{routes_records,inquiry_request,clinical_plan_request}.go、backend/internal/reservation/routes.go。
- docs/work/emr/emr-518-package-api-contract-20261020.md、backend/docs/fixtures/emr518/{plan-list.response,plan-detail.response}.json。

repo根拠の固定prefix: https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/ 。旧実装が残る箇所は採用仕様との適合差であり、そのまま流用を許す根拠にしない。新MTG健診PDF/RV4文書の別owner要件はこの診療終了／訂正sliceへ混ぜない。公開は技術契約の固定のみ。runtime登録、DDL/model/ERD適合、全writer必要subset、試験とDB適用は未完・NOT_RUN。
