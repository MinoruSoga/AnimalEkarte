# EMR-458 コース別追加option 採択技術契約

2026-10-21。親採択済みのR2を単一契約へ統合した採択技術契約。技術選定とruntime完成は別。runtime未実装、試験NOT_RUN。文書公開以外の外部writeなし。

## 1. 根拠と既決境界
固定main edf71515dbae11735a4184ff64c465e2ea51af88。
- backend/migrations/001_init.sql:7902–8007,8544–8579：T1/T2保存型・version guard・membership trigger。
- backend/internal/model/trimming_master.go：course/option contract_version、option nullable service_identity_id。storage_reservations.goにT2保存modelあり。再作成しない。
- backend/internal/trimming/routes.go：既存master-trimming view/create/edit/delete。trimming_option_request.go / service.go / repository.go はT2設定入力なし。既存FindByIDはambient TX内SHARE、Updateは汎用scoped update、Delete/Reorderも新設定lockには未参加。
- docs/work/db-contracts/unified-db-20261019.md T2：course_id NULLはコースなし専用。ruleなしunknown、明示空は許可optionなし。全コースfallbackなし。
- docs/spec/screens/settings/master-trimming.md：1option1回、内包施術二重選択不可、未設定時間を0補完しない、既存予約へmaster改定自動反映なし。
公開API snapshot api-9459cb23.yamlは既存option CRUDのみでT2専用wire未収録。本契約routeを「既にある」と報告しない。
すべてのsource URLは https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/{path} 。

## 2. 状態の最小追加（既存DBだけでは区別不能）
A. rule set：既存行の有無をそのまま使う。未作成=unconfigured、active親あり= configured、active子[]は明示空。course=NULLの親はコースなし専用。削除履歴のある親を再利用して未知へ戻さず、新親IDを作る場合はCASで旧親有無を検証する。
B. option identity：新nullable列 service_identity_assignment_state を追加、値はassigned/not_assigned、NULLはlegacy_unknown。既存非NULL identityはassignedとして読めるが、旧NULLをnot_assignedへbackfillしない。CHECK assignedならidentity非NULL、not_assignedならNULL、stateNULLでidentity非NULLは移行期間の既存assigned表現として許容。新設定writerは両方を必ず明示。旧API createは移行完了までunknown、単にcreated_atが新しいからnot_assignedにしない。
C. course内包集合：trimming_coursesに included_services_configured boolean nullable と included_services_digest bytea nullable(32bytes)追加。旧NULL=legacy_unknown。新設定はtrue+完全active集合digest、[]も真の設定済空digest。falseは採用しない（未設定を明示空へ誤変換しない）。既存内包行があってmarkerNULLなら列挙可能な既知部分にすぎず完全性unknown。GETで補完/更新なし。
B/Cの追加は未知状態の表現と既決重複検査を成立させる保存方式で、全許可・新適格性・実データmappingの決定ではない。not_assignedは「identity比較できないことが既知」でありduplicateなしの証明ではない。identity不明/未割当optionを重複検査済みに変換しない。
D. identityの論理同一性は同医院IDのみ。名称一致でmergeしない。optionへの付替えは専用CAS操作でのみ行い旧予約snapshotを変更しない。rule parent/childのPK/clinic/course/option identity直接UPDATEは禁止（既存trigger維持）。

## 3. selected route・closed DTO（既存prefix /api/v1）
既存master-trimming grantを再利用。新grantなし。医院は認証済selected clinicから取得しbody clinic_id禁止。新ID/版は正規形正十進string、nullableは明示、legacy数値wireを変更しない。
1. GET /masters/trimming-service-identities：view。items[{id,display_name,deleted}]。同医院のみ、未知名を推測しない。
2. POST /masters/trimming-service-identities：create。{operation_key,display_name}、display_name nullable既存保存許容に合わせる。作成はスタッフの明示操作だけ、option名から自動生成なし。初期sliceで改名/削除routeは追加しない。
3. GET /masters/trimming-options/{id}/service-identity：view。{option_id,option_version,state,service_identity_id}。state=legacy_unknown/not_assigned/assigned。
4. POST /masters/trimming-options/{id}/service-identity：edit。{operation_key,expected_option_version,assignment:{state,service_identity_id}}。assignmentはassigned+ID又はnot_assigned+null、legacy_unknown書込禁止、省略/nullassignment禁止。解除はnot_assigned明示。旧PATCHはこのfieldを受け付けない。
5. GET /masters/trimming-courses/{id}/included-services：view。{course_id,course_version,state,service_identity_ids,digest}。unknownではdigest=null、見えている旧行を完全集合と称しない。
6. POST /masters/trimming-courses/{id}/included-services：edit。{operation_key,expected_course_version,service_identity_ids}。配列必須、null/省略不可、[]は明示空、duplicateID拒否。全集合置換・同医院実在identityのみ。新false/unknown操作はなし。
7. GET /masters/trimming-option-rule-sets?scope_kind=course&course_id={id} 又は ?scope_kind=no_course：view。courseの場合ID必須、no_courseの場合course_id禁止。{scope,configuration_state,rule_set_id,contract_version,option_ids}。unconfiguredはid/version=nullかつoption_ids=[]（状態を消さない）。configured空も[]だがid/version非NULL。
8. POST /masters/trimming-option-rule-sets：edit。{operation_key,scope:{kind,course_id},expected_rule_set:{id,version},option_ids}。scope course+ID又はno_course+null、両memberrequired。expected両nullは確定未作成のみ、両nonnullは同じactive親CAS。[]明示空、null/省略不可。scope変更は別資源、既存親のcourse_id付替えなし。親存在だけではcreate grantを新設しない。これは既存masterの対応設定edit。
9. GET /masters/trimming-option-candidates?scope_kind=...：master viewの設定preview専用。予約/LIFF公開の新grantを兼ねない。出力{scope,rule_set_id,rule_set_version,configuration_state,course_version,included_services_state,included_services_digest,candidates:[{option_id,option_version,service_identity_id,identity_state,decision,reason_codes}]}。価格/数量/時間再計算は返さず、既存ownerの属性readerへ接続する。decision eligible/ineligible/unknownはT2対応/内包の限定判定と明示し、全予約適格性の証明にしない。
10. 各新POSTの既存C06結果照会を同一consumerとして接続。専用新read grantなし。lookupは原targetのmaster view、未成功Applyは対象create/edit。operation key UUIDを別キーへ複製しない。

## 4. not_assignedは新しい業務禁止ではない
既決は「コース内包施術の二重選択不可」であり、全optionにidentity登録を必須とする新しい業務規則ではない。configuredの内包集合が完全な空、又はscope=no_courseなら、内包重複という論点自体がないためidentity未割当だけで不適格にしない。完全な非空内包集合とidentity未割当／旧不明optionの組合せは比較不能=unknown（重複なしと推定不可）。assignedの場合だけ同医院identityの集合membershipで判定する。他の対応集合・既存active/時間/併用条件は独立。異なるoption同士を同identityという理由だけで新たに禁止しない（既決の同option1回とは別）。旧NULL不明と新明示無割当の表示は区別するが、それだけで新適格性の差は作らない。

## 5. C06 exact Content・intent・元target
Content.Fieldsはmap[string]string、exact keysはschema_version/resource_kind/target_ref/payload_sha256。schema_version="1"、resource_kind=service_identity|option_identity|course_included|option_rule_set。target_refはidentitycreateのみ"new"、option/courseは正十進ID、ruleは"course:<id>"又は"no_course"。payload_sha256は閉じたpayloadのJCS UTF-8 SHA256 lowercase64hex、配列は数値ID昇順（lexical順でない）でunique。nullable display_nameはpayloadJSON内にのみ残す。期待版/expected_rule_setはPreconditionsで、hash Contentへ混ぜない。Runnerはclinic/intent/keyとのbindingを既存通り検証する。
intentはtrimming.service_identity.create / trimming.option_identity.replace / trimming.course_included.replace / trimming.option_rules.replaceの4つ。新private typed factoryがContentとpayloadを同時生成しApplyがhash/target/4fieldsを再照合。createのtarget_ref="new"は結果ID予測ではなく、元操作bindingの対象型を表す。
短claim metadata用にtrimming_configuration_operation_targetsを追加する。PK(clinic_id,binding_id)、FK(binding_id,clinic_id)→operation_bindings(id,clinic_id)。operation_key uuid NOT NULLはbindingと同値を検査し、clinic/key単独uniqueを新設しない。columns intent text NOTNULL, resource_kind text NOTNULL, scope_kind text NOTNULL, target_ref text NOTNULL, target_id bigint NULL, created_at timestamptz NOTNULL defaultclock_timestamp()。scope_kind=clinic|course|no_course|option。identitycreateはclinic+new+target_idNULL、optionidentityはoption+ID、courseincludedはcourse+ID、rulesはcourse+ID又はno_course+NULL。CHECKで4intentの有限shapeを固定。医院RLS・UPDATE/DELETE禁止。対象business FKは置かない。短claimでmaster/course/option/identity行をlockせず、存在/状態/版は本Applyで照合する。
既存binding発見時はこの元anchorをロードし、現在selected clinic＋master-trimming:viewで結果開示可否を判断する。bodyの別targetを使って認可しない。新keyは現create/edit preflight、短claim+anchor、未成功Applyで再create/edit。replayにcreate/editを再要求しない。削除済みtargetでも同医院元anchorと現在master viewで元receipt参照可、現在業務適格性を再実行してreceiptを書換えない。
lookupは GET /masters/trimming-configuration-operations/{operation_key}?intent={4種}。現在認証/医院/master view必須、未知key（binding未発見）は既存Runner.Lookupに合わせscope-not-found/404。not-succeeded/200は既存bindingのsettledStateによる未成功結果であり、未知keyへ流用しない、未読取/DB失敗を不在扱い禁止。新grantなし。既存RunnerのResolveLookupTarget error／元anchorのbinding・clinic・identity不一致はnot-authorized/403へ閉じる。これを503化する専用adapterは作らない。bodyのimmutable Content/hash conflictによるoperation-key-conflict/409とは区別する。全成功replayは元内容・版のみ。

## 6. exact status / HTTP / 成功shape
12status：succeeded200、replayed200、validation400、version-conflict409、dependency-changed409、edit-right-lost409、operation-key-conflict409、operation-in-progress202、scope-not-found404、not-authorized403、result-unavailable503、not-succeeded200。既存494 MedicalRecordOperationResult mappingと同じで、新enumを増やさない。202は同key照会、503/transport500はunknown停止、新key再実行禁止、Retry-Afterがあれば尊重。not-succeededは成功でも無条件再実行許可でもなく、現Preconditions再確認を必要とする。入口invalidJSON/未認証/現scope拒否は既存ErrorResponseの400/401/403/404。
成功required: status,operation_key,result_refs,committed_versions,recorded_at。result_refs/versionsはmap[string]stringだけ。identitycreate refs={resource_kind:"service_identity",service_identity_id} versions={}、optionidentity refs={resource_kind:"option_identity",option_id} versions={option:version}、courseincluded refs={resource_kind:"course_included",course_id} versions={course:version}、rules refs={resource_kind:"option_rule_set",rule_set_id,scope_kind:"course",course_id}又は{resource_kind:"option_rule_set",rule_set_id,scope_kind:"no_course"} versions={rule_set:version}。no_courseでcourse_id=null/"null"/"0"を返さず有限branchで省略。identitycreateの{}を必ず保持する専用successDTOを使い、genericconverterは変更しない。非成功DTOはstatus/operation_keyだけ。

## 7. audit exact sink
既存audit_logsに同ambientTXで1成功operation1行、actor_type=staff、actor_id=trusted current staff、clinic_id=trusted scope、created_at=server、IPAddress/UserAgentの新必須化なし。action/resource/resource_id：
- trimming.service_identity.create / trimming_service_identity / 新identityID
- trimming.option_identity.replace / trimming_option / optionID
- trimming.course_included.replace / trimming_course / courseID
- trimming.option_rules.replace / trimming_option_rule_set / ruleSetID
いずれも既存varchar50以内。metadata exact={schema:"trimming-config-audit-v1",operation_id,operation_key,scope_kind,target_ref,changed_fields}。operation_id正十進string、operation_keyUUID、changed_fieldsはidentity_assignment / included_services / allowed_options / display_nameの該当path昇順配列。old_value/new_valueはexact {state,contract_version,member_ids,digest}。state=legacy_unknown|not_assigned|assigned|unconfigured|configured、versionはstring又はidentitycreateのnull、member_idsはIDstring配列（identityassignment最大1）、digestはnullable64hex。create old_value=null。本文/価格/患者情報を複製しない。no-op成功でもaudit1行＋receipt、版不変。replayは追加auditなし。audit失敗は設定/marker/digest/receiptすべてrollback。既存画像参照auditは別意味なので維持し、同じT2auditをservice/repo双方で二重記録しない。

## 8. marker・digestのDDL整合
option state列CHECKは(state IS NULL) OR (state='assigned' AND service_identity_id IS NOT NULL) OR (state='not_assigned' AND service_identity_id IS NULL)。既存nonNULL+stateNULLはreaderでassigned既知とするが旧NULLはunknown。新専用writerはstate必須、legacyCRUDで割当fieldを変更不可。任意SQL書換が保証外という理由で正規writerを省略しない。
course CHECKは(configured IS NULL AND digest IS NULL) OR (configured IS TRUE AND digest IS NOT NULL AND octet_length(digest)=32)。digest IS NOT NULLを明示し、CHECKのNULL三値論理による通過を防ぐ。course/clinicのimmutableidentityは既存FK境界維持。
digestの正確bytesはUTF-8のprefix "trimming-included-v1\\n"（仕様表記の\\nは実LF1byte）に、active service_identity_idを数値昇順の十進文字列＋LFで連結。空集合はprefixのみ。PostgreSQL側は同形式string_aggのCOALESCE('')を使いpgcrypto.digest(...,'sha256')。JCSのSQL再実装は不要。client計算値を信頼しない。
DEFERRABLE INITIALLY DEFERRED constraint triggerをcourseのINSERT/UPDATEとincluded child INSERT/UPDATE/DELETEに置く。OLD/NEWの影響courseを必要なら両方列挙し、最終course row＋同医院active子集合を読み、configured=trueならdigest完全一致、unknownならdigestNULLを検証。子identity/course/clinicの直接付替えは禁止しretire/insert。triggerで未知をconfiguredへ変更しない。
子差分は必ずcourse親UPDATE lock後。専用serviceは最終子集合を計算して親marker/digestを一度更新、既存storage_contract_version_guardによって実差分時に版増分。正規writer以外の子変更だけでdigestを更新しない場合はdeferred検査でcommit拒否。同集合のsoft-retire→restoreの無意味往復は禁止し差分だけ更新。rule子は既存AFTERtriggerの親版増分を維持し手動+1を重ねない。

## 9. 全writerのlockとread consistency
未作成no-course集合はrow lockできないため、医院単位T2設定用transaction advisory lockを1個追加する。設定writerはexclusive、候補の整合read/予約captureはshared。同医院しか対象にせず、hash collisionは直列化増加だけでscope漏洩を起こさない。全writer閉包前にはこのlockだけで安全と報告しない。
設定POST：C06の対象consumerに必要な既存ContextGate（設定だけの操作へ予約graph gateを新必須化しない）→binding→必要actor/current grant→医院T2 EXCLUSIVE→course昇順→rule親昇順→option昇順→identity昇順→子差分→audit→receipt。新scope作成も同lock内で有無再確認。設定writerはappointment/予約detailを取得しない。
予約/施術採用側：既存C06/booking advisory/appointment/detailの既定順を維持→医院T2 SHARED→course/rule親/option/identityの同順SHARE→完全集合・版再検査→採用snapshot。設定writerは予約親を後追いしないため、この順で設定↔予約の逆順循環を作らない。予約の既存master SHAREが医院T2より先に実行される入口は、活性化前に順序を修正する必要subset。
設定GET preview：短いread TXで医院T2 SHARED→同順SHARE＋同一snapshot。GETは未設定親やmarkerを作らない。permission/read failureは成功空集合へ変換しない。
必須参加writer：既存course create/update/休止/delete/reorder、option create/update/休止/delete/reorder、新identity作成・割当解除/付替え、course内包編集、rule set作成/子追加/復帰/無効化、将来bulk/import。既存service/repositoryの各書込がすべて医院T2 EXCLUSIVEをmaster row前に取得。画像metadataがcourse UPDATEする経路も同じguardへ。storage upload自体をDB lock内に移さない。
identity削除/改名を初期sliceで公開しないが、既存raw/internalwriterがあれば同閉包か明示停止が必要。DB権限外の任意SQL操作まで安全保証を拡張しない。

## 10. 実caller閉包：確認済み入口と必要subset
固定mainの実コードを確認した範囲：
- routes.goのCreate/Update/Delete/ReorderTrimmingCourse → trimming_course_service.go Create/Update/Delete/Reorder → trimming_course_repository.go Create/Update/Delete/Reorder。Create/Update/DeleteはWithTx、Reorderは直接repo。全て医院T2 EXCLUSIVEを最初のmaster／course-type／image lock前へ追加する必要がある。Updateのtx外FindByIDはfastcheckのみ、確定値はtx内UPDATE lock。既存画像更新はLockForImageUpdate→image keyの順、T2→course→imageへ拡張し逆転しない。Create画像もT2→course-type→image→INSERT、既存record writerとのimage順は維持。
- routes.goのCreate/Update/Delete/ReorderTrimmingOption → trimming_option_service.go同名 → trimming_option_repository.go。Create/Update/Reorderは現状ambientTX保証なし、全体WithTxでT2取得必須。Createのfalse補正2UPDATEも同TXへ包む。Deleteの内側Transactionは外側ambientを使うsavepointとして統合、独立TXcommit禁止。
- 上記repoが直接利用されるconsumer：reservation/liff_service_catalog.go GetTrimmingCourses/GetTrimmingOptionsはFindAllのみ、reservation/reservation_validators.goはfinder、trimming/trimming_service_validate.go と trimming_service_mutate.goはFK/採用reader、billing/billing_item_service_create.goは所有権reader。これらはmaster writerと誤分類しない。予約/施術採用readerのT2 shared＋版captureは480/482必要subset、billing所有権readに新予約適格性を要求しない。
- composition_runtime.goが2repositoryを構築し、composition_billing.go等へfinderとして注入。新lock依存をserviceconstructorに必須注入し、nilなら書込fail-closed。service/privatewrite経路以外のbulk/importを現在存在すると推定しない。ただし実装PRの静的呼出一覧で直接GORM Table/Model書込が見つかれば同閉包へ取り込むまで活性化不可。
- course image upload/discard自体はblob準備／清掃でcourse rowの設定writeとは区別。course参照変更は確認済みCreate/Updateにある。画像cleanupがT2/courseを取る新逆順を作らず、既存keylock→参照readはnonlocking readのまま。未知別writerの不存在までは検索結果だけで証明していない。
この表は実装指示の必要subsetで、runtime全caller参加済みという判定ではない。Aで既存8CRUD mutation families＋新4mutation families＋画像参照変更経路を静的対応表にし、Bで院内/LIFF予約・施術のcapture接続を証明する。全体gateを他の無関係course画像機能や料金機能へ広げない。

## 11. 最小PRと引渡し
PR-A（458専有）：state/digest最小DDL+model、T2 private config service/repository/handler/DTO、新route、C06 adapter・audit、設定UI往復、fixture定義。新private候補 backend/internal/trimming/option_configuration_{request,response,service,repository,handler}.go と option_configuration_locks.go。既存routes.go/api.yaml/model/trimming_master.goの差分は457担当と編集順を合意。既存CRUD全writerlock参加はAの必須subsetで、後送りして安全完成としない。
PR-B（480/482側）：shared verifier adapterを予約/LIFF/施術採用writerへ接続。458は候補readerとValidateAndCaptureConfiguration(ctx,scope,expectedVersions,selectedOptionIDs)を提供、出力は対応根拠・identity・版だけ。480が時間算定・予約version・LIFF actor・予約snapshot、482が実施明細/会計sourceを所有。A単独で本番予約適格性を活性化しない。
457受領必須subset：course医院scope/有効削除/契約版/曜日条件、共有file編集順、course writerlock参加。画像全機能完成はgateにしない。
480採択技術契約 `emr-480-booked-duration-selected-contract-20261021.md` が区分基礎時間snapshot、本人actor、退役済course参照削除guard、preview commitmentを定義する。458は重複定義せず、その必要adapter subsetへ引き渡す。
477はカルテ再診問診readerでありoption料金ownerではない。料金変更/保険/見積/請求の範囲へ拡大しない。

## 12. fixture設計（すべてNOT_RUN）
- A→X、B→Y、no_course→Z。NULL集合がA/Bへfallbackしない。
- 未設定親／configured空／configured非空をGET→保存→再GETで区別。取得失敗を空successにしない。
- 旧option NULL=unknown、新明示not_assigned、assigned別IDを区別。新無割当をduplicateなし扱いしない。
- course内包unknown/明示空/非空。同名異identityは統合せず、別optionID同内包identityは重複採用不可。
- crossclinic／inactive／deleted／既存の時間NULLと明示0・曜日条件は既決verifierと合成し、新条件を推測しない。
- 同scope初回作成race、親先行child更新、集合複数差分のversion最終値、membership no-op無更新。
- option割当変更 vs candidate/capture、course内包編集 vs capture、master休止/delete/reorder vs capture、grant loss／CAS conflict／C06samekeyreplay／unknown停止。
- audit/receipt/子更新いずれか失敗は設定全rollback、旧予約時間/料金/会計snapshot不変。
- 全writer一覧を実ファイル/functionと対応させる静的適合表。未参加入口をテスト省略によって完了扱いしない。

## 13. 追加fixtures
course内包空/no_courseで未割当を一律拒否しない、内包非空でunknownをeligibleへ変換しない。Content4値が全部string、no_course refsにcourse_id無し、identitycreateversions{}、新claim時にbusinessFKlock無し、別targetbodyで元key結果を読めない、現在view失効でreplay開示不可。rule複数行差分で最終版が+Nでも正しく返る。course子だけ変更→digest不整合rollback、旧unknownにGETbackfill無し。画像変更T2→course→imageとcapture並行の順序fixture。すべて定義のみNOT_RUN。

## 14. 480との接続と変更履歴
480 previewは458のT2結果を構成要素として同一snapshotからdigest化する。458のmaster設定preview自体は予約の開始/患者/時間を含まず、480のadoption_digestと同一ではない。course内包不明でも選択optionが空なら新たな必須設定を課さない。optionありで必要比較がunknownならunknownをeligibleへ変換しない。option無割当それ自体の一律禁止はしない。
480/482 adoptionは予約/施術の既存親lockを先に取得してT2 SHARED、その後のcourse/rule/option/identity順を共通化。course削除はT2 EXCLUSIVEのままT4使用有無を非locking queryで読む。detail/appointmentを削除側からFOR UPDATEしない。予約側がT2前にmaster lockを持つ旧入口は接続前に修正する。
変更履歴: R1の数値Content、nullable refs、未固定status/audit、無割当一律不可と読める表現を置換。operation anchorは実際のbinding PK/FKへ統合。全writer参加の実装証拠は未受領で、選定完了と混同しない。
