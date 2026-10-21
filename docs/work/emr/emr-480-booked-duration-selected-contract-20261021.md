# EMR-480 採用時間・本人actor・削除guard 採択技術契約

2026-10-21。親採択済みのR2を単一契約へ統合した採択技術契約。技術選定とruntime完成は別。runtime未実装。文書公開以外の外部更新、DDL適用、コード実装、試験なし、全fixture NOT_RUN。

## 1. 固定根拠と非変更境界

照合SHA: `edf71515dbae11735a4184ff64c465e2ea51af88`。
- [予約の時間8点合意](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/spec/domains/reservations/pages/reservation-management.md#booking-course-option-duration-20261017)
- [既存T3保存契約](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/design/domains/reservations/technical-spec.md#booked-duration-storage-20261019)
- `backend/migrations/001_init.sql` のT3、`app_private.trimming_booking_snapshot_complete`、後段の `trimming_booking_options_complete`、immutable/RLS。
- `backend/internal/model/storage_reservations.go`、`docs/architecture/erd-detail/trimming_booking_revisions.md`。
- `backend/internal/trimming/trimming_course_repository.go` のCountUsage/Delete。
- `backend/internal/reservation/reservation_service.go` Create/CreateBatch/Update、`reservation_validators_create.go` createLineReservationInTx、`reservation_type_repository.go` FindByID。
- 458統合契約の設定意味・T2 lock・内包識別は `emr-458-option-assignment-selected-contract-20261021.md` の公開時固定版へ接続する。独立して再設計しない。

コース0or1、コースありはその時間、なしは予約区分時間、各optionは一度・数量なし、option明示0可、未設定拒否を維持。料金・実施済み・会計追加へ変換しない。予約区分の既存利用条件はコース選択でも免除しない。手動終了保持は維持し、コース/option変更時の再計算だけ既決どおり明示確認する。

## 2. 実欠落

T3 courseなしではcourse_duration_minutesもNULL。現header/childの両complete triggerが `sum(options)+coalesce(course_duration,0)` をtotalと比較するため、区分30+option10を40として保存できない。列だけ足してtriggerを片方だけ直すのは不十分。
recorded_byはstaff FK NOT NULLでLIFF本人を表せない。予約担当トリマーを作者へ転用しない。
現LIFFは `createLineReservationInTx` で旧detail.CourseID/SetOptionsをwriteし得る。これはT3採用保存・T4実施snapshot成立の証拠でなく、置換必要なcaller。既存T3 modelの検索結果はstorage mapping/tests中心であり全writer不存在の証明とはしない。

## 3. 最小DDL/model差分

既存T3に次を追加する。既存PK/FK/同appointment連続revision/immutable/RLSは維持。
- `snapshot_schema_version smallint NULL CHECK (snapshot_schema_version IS NULL OR snapshot_schema_version=1)`。NULLは既存未分類、1は新writerの完全形。DEFAULT 1は付けず、古いwriterを誤認しない。
- `reservation_type_id bigint NULL`、`reservation_type_duration_minutes integer NULL`、`reservation_type_version bigint NULL CHECK (reservation_type_version > 0)`。複合FK `(reservation_type_id,clinic_id) -> reservation_types(id,clinic_id) ON DELETE RESTRICT`。
- schema1ではreservation_type_id必須。courseなしでは区分duration>0と既存reservation_types.version必須。courseありでは区分durationはNULL（時間基礎として不採用）、区分versionは選択適格性の根拠として必須。course列の既存3列pair制約はそのまま維持。
- schemaNULLは追加区分3列NULLのまま。既存course有りでcourse/base/totalが読めることと、全snapshotの完全性証明は区別する。旧courseなしのbaseをtotal-optionsや現masterから逆算・補完しない。
- `recorded_by`をnullableにする（既存staff FK維持）、`actor_kind text NULL CHECK (actor_kind IN ('staff','line_customer'))`、`recorded_by_line_customer_id bigint NULL` を追加。LINE複合FK `(recorded_by_line_customer_id,clinic_id) -> line_customers(id,clinic_id) ON DELETE RESTRICT`。
- actor制約: 旧schemaNULLはactor_kindNULL、既存recorded_by非NULL、LINE列NULL。schema1は staff+recorded_by非NULL+LINE NULL または line_customer+staff NULL+LINE非NULL。staff 0、予約担当者、body指定actorを許さない。既存staff値を書換えてactorを推測backfillしない。
- modelはpointer列とclosed enum、`RecordedBy *uint64`、新LINE列 `*uint64`、区分ID/分数/版/スキーマ版を対応。全てstorage JSON非公開。ERDは生成元DDLと同じPRで更新し、手編集だけで整合したとしない。

schema1の新INSERTは選択基礎 `CASE WHEN course_id IS NOT NULL THEN course_duration_minutes ELSE reservation_type_duration_minutes END` を用い、headerとchild両deferred triggerでcount・合計・正のtotalを照合する。schemaNULLの旧既存行を再計算/更新しない。子の遅延追加も同じ検査で拒否、header/childのimmutableを解除しない。現在の旧形INSERTは旧writer退役まで互換として残るが新consumerは必ずschema1を送る。schemaNULLが新完全採用成果として返らないことをfixture化する。

## 4. 時間解決と旧snapshot

private `booking_duration_resolver.go` に同じ型付き計算結果を一つ持つ。入力は認可済みclinic/pet/区分/開始、course nullable、distinct option IDs、現在採用revisionまたは新規、明示selection-change flag。出力はbase kind、採用ID/値/観測版、options、total、proposed_end、T2適格性の結果と根拠。現masterを読み直す単純helperを各画面へ散在させない。
新規/選択変更はT2検証後にmaster値を固定し、候補→空き枠→保存へ同じ算定値を渡す。ただし候補表示は予約確保でない。保存TXで再取得・比較し、観測版/値/適格性が変化した場合は再表示確認へ戻し、表示した終了と別値を黙って採用しない。reservation_typesは001後段に既存versionがありmodelも対応済み。row SHARE lock中のid/version/durationを保存する。versionだけでなく採用値もdigestに含め、現writerのversion増分未接続を隠さない。新しい同義版列は作らない。
選択を変えない既存編集は保存済み時間を保持し、master改定だけで新revisionを作らない。開始/区分変更や手動終了は既存ownerの方針へ従い、ここで一律再計算・上書きを発明しない。コース/option明示変更は既決の再計算を行う。旧snapshot欠落の予約は既存start/end表示を維持し、採用内訳はunknown。単なる閲覧/notes変更でsnapshotを補完しない。
総所要時間と実予約start/end差分を常時同値CHECKにはしない。既決の手動終了を破壊するため。自動算定の作成/選択変更時だけ同一算定結果でproposed_endと採用totalを確定し、その後の明示手動終了の意味は既存予約契約へ残す。

## 5. actor・scope・原子性・再送

院内は既存reservations:create/edit、閲覧はview、認証staffと医院scope。LIFFは既存本人token→server customer/clinic→現在owner/pet grant、対象予約の現在利用可否を再検査。master編集grantを顧客へ要求・追加しない。現在のcustomer.owner_idだけで成立させず、472/519の既存selected current-grant adapterの必要subsetへ接続する。LINE sub/subject hash/内部actor IDを表示DTOへ公開しない。
予約の最外TX開始→467 graph gate（returned ctx伝播）→既存操作binding/現在認可/booking・owner/pet・appointment保護→458 T2 SHARED→区分row SHARE→course/rule/options/identityの458順→再検証→予約変更とT3 header/options→既存必要監査/成功receiptを同TX。snapshot INSERT guardはappointmentをFOR UPDATEするので、T2取得後に初めて予約行を逆向き取得しない。新規は予約ID確保/行INSERTをTX内で行い、失敗なら全てrollbackする。候補のreadはwrite graph gateや予約作成を伴わない。
現在LIFFはbooking lock後にmaster所有を読む一方、院内Createには区分readがTX外にもある。TX外readは表示/preflightのみ、採用根拠にはせず上記順のauthoritative readを追加する。458設定writerはappointmentへ戻らない。区分masterの既存UPDATEはrow lockで競合を閉じ、T2 master用lockだけで区分更新も保護済みと誤認しない。
T3は予約commandのparticipantとし、別の独立snapshot保存endpoint/operation identityを作らない。予約の同operation成功再送は固定結果を返しT3を増やさない。型付きadopted snapshot hashを外側予約command content/receiptへ束縛し、同key異選択を拒否。結果不明で新keyを発行しない。snapshot/option/監査/receiptのどれか失敗で予約もrollbackする。現在旧予約経路のreceipt実装済みとはしない。Bの必須subsetは該当する予約create/update/tentative吸収commandの既存C06接続で、480だけのreceipt基盤を新設しない。
LINE actorのT3・外側appointment_change_events・operation bindingの一致は本書§8で選定する。実装者へ未決として戻さず、同じPRの必要schema/model/ERD subsetへ含める。

同様にT3.pet_idは必須。現LIFFのpet未解決/best-effort後付けへ無理にT3を挿入せず、519の1予約1pet確定subsetへ接続する。未解決を架空petや名前一致で埋めず、480から全予約に新たな患者必須業務を課さない。これは対象経路の保存契約統合依存として明記する。

## 6. Pure preview と canonical adoption digest

院内 `POST /api/v1/reservations/trimming-duration-preview`、LIFF `POST /api/liff/{clinicId}/trimming-duration-preview` を純read操作として追加する案を選ぶ。GET queryへの配列拡張案は採らない。予約・target・customer・operation metadataを作成せず、previewを予約確保としない。院内は既存reservations:viewと医院/pet行scope、LIFFは既存本人/現在患者利用認可。previewをmaster編集grantへ接続しない。

requestはclosedで全キー必須: `{appointment_id, pet_id, reservation_type_id, start_time, course_id, option_ids}`。appointment_id/course_idはstringまたはnull、他IDは正のsigned-int64 canonical decimal string、option_idsは同string配列（重複400）、start_timeは既存RFC3339契約。appointment指定時は対象医院/患者/既存現在版を照合。新規時のpet_idは519の確定local pet identityを使い、名前/customer_fieldsから推測しない。courseなし/optionなしはnull/[]で表す。owner ID/actor/分数/価格はrequestへ追加しない。

成功200はclosed `{state:'ready', adoption, adoption_digest, proposed_end_time}`。`adoption`のexact keysと型は以下。optionはIDの数値昇順。各objectはclosed、nullも省略しない。
- `schema_version`:整数1
- `clinic_id`,`pet_id`:canonical ID string（clinicはtrusted context）
- `appointment_id`:string|null
- `reservation_type`: `{id:string,version:string,duration_minutes:integer, business_kind:string}`。durationは現在値でありcourseありでも適格性/表示観測の入力、totalへ二重加算しない。business_kindは既存確定非NULL値、legacy category推測不可。
- `course`:null または `{id:string,version:string,duration_minutes:integer,included_services_digest:string}`。digestは458の完全included集合SHA256下位hex64、未設定/不完全はnull。optionが空なら内包比較を必要とせず、nullを既知空集合へ置換しないままready可。選択optionがあり必要な458適格性がunknownならその契約のunknown応答へ。
- `rule_set`:null または `{id:string,version:string,scope_kind:'course'|'no_course',course_id:string|null}`。nullは設定不存在であり空設定とは別。optionが空なら458の既決に従い不要なrule設定を予約全体の新必須条件にしない。
- `options`: `{id:string,version:string,duration_minutes:integer,service_identity_id:string|null,assignment_state:'assigned'|'not_assigned'|'legacy_unknown'}` の配列。未知identityをfake IDへ変換しない。設定と適格性は458統合契約を適用し、本contractで強めない。
- `base_kind`:'course'|'reservation_type'
- `base_duration_minutes`:正整数
- `total_duration_minutes`:正整数
- `start_time`:UTCに正規化した既存精度のRFC3339 string。`proposed_end_time`はstart+totalで同精度。

`adoption_digest = lowerhex(SHA256(UTF8("trimming-adoption-v1\n") || JCS(adoption)))`。ここでprefix末尾は実LF一文字。ID/版は文字列、分数は正確な整数、文字列の勝手なtrim/改名なし。arrayの並びは上記一意順、object順はJCS。hashは公開整合tokenで認可証明ではない。新secret/署名鍵は不要。患者ID等は既存認可範囲だけへ返し、内部LINE sub/hash/actor/customer IDはadoptionに含めない。

保存requestのnew `trimming_selection` はclosed **4 keys** `{course_id,option_ids,expected_booking_revision_id,expected_adoption_digest}`。expected_adoption_digestはhex64必須、expected_booking_revision_idはnewnull/編集時実保存ID、旧snapshot不存在はnull。clientはpreviewで表示したdigestをそのままechoする。body hashだけでは認可・適格性成立とせず、保存TXでtrusted clinic/pet/currenttype/startと再読したmaster/configから同adoptionを計算し、constant-time等の通常安全比較で一致確認。違えば409 `TRIMMING_ADOPTION_CHANGED`、write/成功receiptなし。最新previewを取り直して終了時刻を見せる。確認対象adoption_digestが変わる場合は元operationの結果を確定させた後、新しい明示操作・新keyで送る。unknown/busy中は別keyを作らない。同じdigestのまま現在状態CASだけを再確認する場合に限り同key再送できる。自動確認/黙った置換は禁止。

編集時のexpected booking refは対象appointment/clinicの最新T3 IDと比較。別予約IDや古いrefは409 `TRIMMING_BOOKING_REVISION_CONFLICT`、対象不可視は既存404。startやpet/区分を変更したら旧previewは使えない。master改定と無関係なnotes編集でselectionを省略した場合は、この再採用・digestチェックを要求しない。既存手動終了の非上書きも維持。

## 7. 旧client/旧field切替と有効化条件

新previewと保存対応を非活性で先に配備→新clientがpreview/4key selection/operation_keyを送る→旧trimming編集画面を更新/再読込させ、旧形保存をbackendで受け付けない切替→対象経路を有効化する。旧clientが残るのに新producerへ同時に切り替わったとみなさない。
院内batchは既存pets配列の各要素をclosed `{owner_id,pet_id,trimming_selection}` へ拡張し、各petについて独立previewの4key selectionを必須にする。batchのtop-level trimming_selectionは拒否し、別petのdigestを使い回さない。既存同一時間/同一担当の共有batch契約を守り、各算定結果がその共通区間と整合しない場合は予約を分ける判断を自動実行せず、既存入力検証エラーとして全体不保存。非trimming既存batchの入力は変更しない。

旧 `trimming_course_id` / `trimming_option_ids` は移行adapterの入力に限定し、名前/現在masterからexpected digestをserverで作って付与しない。移行期間にこの旧pairを受け付ける場合も **expected_adoption_digest とoperation_keyの新追加欄を必須** にし、4key typed commandへ正規化、同じ認可/T2/receiptへ必ず通す。新objectと旧pairの両指定400。digest/keyがない旧trimming選択writeは400 `TRIMMING_PREVIEW_REQUIRED` と明示再読込案内、旧detail.CourseID/SetOptionsへfallthroughしない。HTTP/メッセージを根拠に自動送信しない。
新規trimming予約で「選択全省略」を旧詳細writerへの抜け道にしない。有効化対象のtrimming作成はcourseなし/optionなしでもpreview済み明示selectionが必要。これは既決時間を記録する通信契約でありコース必須・新料金・業務適格性の追加ではない。既存予約の選択非変更notes等は省略保持を許す。通常の非trimming予約を本sliceの新field必須へ巻き込まない。

## 8. 外側予約operation・actor・receiptの最小選定

既存C06 operation_bindings/resultsを利用する。owner_kind=`reservation`、intents=`reservation.create` / `reservation.update` / `reservation.create_batch`、actor_kindはstaff/customer、actor_refはtrusted staff/customer IDのcanonical string。既に同scopeで採択/実装した予約commandがある場合はそのidentityへparticipant統合し、同じ予約変更に第二keyを作らない。新clientは外側requestのoperation_key（canonical lowercase nonnil UUID）必須。院内/LIFFの既存create/update payloadを保持しnewselectionを追加する、任意body actor/clinicは採らない。

new mutation wrapper Content.Fields は exact4 string keys: `{schema_version:'1',resource_kind:'reservation',target_ref:<'new'|appointment ID|'batch'>,payload_sha256:<hex64>}`。digest projectionは正規化済みの**採択予約業務payload全体**（既存start/end/区分/患者/指名/notes等のowner字段と選択course/option IDs）で、key/actor/expected_version/expected_booking_revision_idを除く。expected_adoption_digestは名前がexpectedでも確認済みsnapshotの意味commitmentなのでContent投影へ含める。現在状態のexpected_version/expected_booking_revision_idだけをPreconditionsへ分離し、同じ確認snapshotのままの未成功CAS再確認だけ同keyで許す。成功後の同keyは元receipt固定、変更内容は別操作。外側ownerが既に全payload hash型を持つ場合はそれにselectionを追加し、別hash方式を併設しない。新wrapperのhash projectionはclosed `{schema_version:1,intent,target_ref,booking_changes,trimming_selection}`。trimming_selectionはclosed `{course_id,option_ids,expected_adoption_digest}`（optionsは数値順、digestは確認済みhex64）。booking_changesは既存routeの明示入力だけから作るclosed PATCH-mapで、許容キーを `start_time,end_time,owner_id,pet_id,visit_type,reservation_type_id,doctor_id,is_designated,status,notes,source,reservation_route,request_text,trimming_style_request,customer_fields,pets` に限定する。未入力キーは省略、明示nullはnull、IDはcanonical文字列、時刻はUTC正規化、bool/enum/textは既存検証後の値を保持、customer_fieldsは既存認めるJSON値をJCS化し原文改変なし。petsは院内batchのowner_id/pet_idおよびtyped trimming_selection（course_id/option_ids/expected_adoption_digest）をpet ID数値順に並べた配列。各petの現在状態CAS値（expected_booking_revision_id等）だけをPreconditionsのpet ID付きmapに分離する。各petのexpected_adoption_digestは当該petのContentへ保持する。batch top-level hash projectionのtrimming_selectionはnull固定。各routeが元から許可しないキーはこの共通mapで許可を追加せず400。現在状態CASをbooking_changesへ紛れ込ませず、確認snapshotのdigestは上記trimming_selectionだけへ一意に格納する。保存時の既存値mergeはfingerprintの前に勝手に行わず、明示入力commandをhashする。サーバー確定値はadoption比較/receiptへ保持し、Content内の確認digestとApply再計算値の一致を必須とする。digestを変更して同keyを再送すればoperation-key-conflictであり、確認内容が変わった操作をCAS変更に偽装しない。

本節のimmutable typed commandは[採択済みQの共通契約](../../design/shared/operation-participation.md#verified-command-selected-20261021)に従い、Applyがreq.VerifiedCommandだけから再構成する。Qはrequest-memoryのみでowner-private耐久保管とは別、CAS/fenceはPreconditions、Contentは既定の短いstring mapを維持する。489のQ必要subsetは未実装依存であり、別command/DTO/pointerのcaptureで迂回せず、既存Content-only consumerは変更しない。


createの短claim targetは業務FKを取らない既存C06型付きmetadata pattern: `reservation_operation_targets` PK `(clinic_id,binding_id)`、FK `(binding_id,clinic_id)->operation_bindings(id,clinic_id)`、`target_kind text` new/appointment/batch、`appointment_id bigint NULL`、`line_subject_hash bytea NULL`（customerのみ32bytes）をimmutable/RLSで保存する。appointment target以外IDNULL、staff subjectNULL。既存同義target表があれば再利用、並行表を作らない。短claimはbusinessrowlockなし。新createの成功実IDはreceiptにだけ記録。lookupは現在元actor/clinic/subjectと元targetのview scopeを検査し、globalkey検索しない。customer actor一致だけで現在grant不検査にしない。

T3へ `operation_id bigint NULL` と `appointment_change_event_id bigint NULL` を追加し、既存binding/eventの複合FK(clinic付き)へ接続。schema1では両方必須、旧NULLはそのまま。UNIQUE `(clinic_id,appointment_id,operation_id)` により同予約同操作のsnapshot二重作成を拒否。eventのappointment/clinicとT3を一致検査し、成功receipt必須deferred guardを同TX適用。batchは同operationに複数appointment可。新batch専用receipt refs exact `{appointments:<JCS配列文字列>}`、配列はappointment ID数値順で各要素 `{appointment_id,appointment_change_event_id,booking_revision_id}`、versions exact `{appointments:<JCS配列文字列>}`、要素 `{appointment_id,appointment_version,booking_revision}` とする。全ID/版文字列、単数キーを上書きしない。既存batchreceiptがある場合はその同義固定結果へ統合し第二receiptを作らない。

appointment_change_eventsは既存recorded_byをnullableへ変更し、`actor_kind text NULL`、`recorded_by_line_customer_id bigint NULL` を追加。旧actor_kindNULLはstaff記録値維持、newstaffはstaff非NULL/customerNULL、newcustomerは逆。customer FK `(id,clinic_id)`、staff FK維持、actorの排他CHECKを置く。new480 commandはactor_kind明示必須。既存staffイベントの更新/作者再帰属はしない。eventとbindingのclinic/actor_refが一致し、storage actor_kind staffはbinding staff、storage line_customerはbinding customerへ対応することをdeferred guardで検証。storageのactor_kind有限値はT3と同じstaff/line_customer、NULLは旧staff行のみ。LINE sub再紐付けの扱いはtargetの元subjectと現在本人認可で検査し、ID再利用推測をしない。model/ERDのnullable型も同じPRで更新する。T3とeventのrecorded_by/LINE actorは一致させる。

単数成功C06 wireはrequired `{status,operation_key,result_refs,committed_versions,recorded_at}`。refs exact `{appointment_id,appointment_change_event_id,booking_revision_id}` string、versions exact `{appointment:<appointments.version>,booking:<T3.revision>}` string。legacy outercreateの201/bodyが必要なら既存route adapterが固定receiptから201互換応答を返し、C06専用lookupは200。初回201とreplay200を混在させない。本案は新mutation endpointを増やさず既存予約routeを拡張するので、createは初回/replayとも既存201adapter、既存更新成功は既存200adapter、C06 lookupは200に固定する。共通converterを変更せず専用adapterでrequired mapsを保持。非成功は既存C0612statusとHTTP mapping（458統合契約§6）を再利用し、未知keyは既存Runner.Lookupどおりscope-not-found/404、既存bindingのlookup not-succeeded/200も無条件の新key再実行許可ではない。上の2つの409業務codeは既存error形のdetailsへ対応させる。成功後current stateとreceipt固定値を混ぜない。

必要監査はこの同TX appointment_change_events（before/afterは予約採用値とT3参照、原文の余分な複写なし）と既存必要audit sinkを維持し、同じ意味の第二監査を新設しない。event/receipt/option/headerのどれか失敗で全rollback。staff認可actor guardとcustomercurrentgrantを元経路どおり使い、全staff権限をcustomerへ転用しない。

## 9. 必要subsetの閉じ方

519/520の公開済みID wire/current pet readerは参照して再設計しない。ただしそのreader公開だけで未解決pet作成経路の確定が保証されるとはしない。480 newconsumerへ渡すのは実appointmentと一致する確定local petのみ。元予約ownerのcreate binding adapterを接続する技術subsetであり、491全体・全LINE移管の完成をgateにしない。
Bの必要変更はこのsnapshot/eventactor/target/receipt adapter＋preview/digest＋旧field廃止guard＋typed bookingparticipantで限定。必要カラムは選定済みとして同一実装ownerへ渡す。実装時は既存outercommand/targetとの重複照合を行い同一ownerで統合する。AのT4削除guardは482の同T2共有lock参加が必要という前提を維持し、A単体で全race解消済みとはしない。
追加fixture NOT_RUN: preview後区分version/値改定、rule親版改定、samecourse含有digest改定、option並び違い同digest、別pet/clinicのdigest流用、未署名hashで認可迂回不可、旧pair+digest移行・旧pairのみ拒否、trimmingcreate省略拒否/非trimming互換、LINE event/binding/T3 actor一致、同key復送1event1T3、batch同key各appointment一版、receipt途中失敗全rollback。

## 10. operation lookup の exact route / status

新read-only lookupは院内 `GET /api/v1/reservations/operations/{operationKey}?intent=reservation.create|reservation.update|reservation.create_batch`、LIFF `GET /api/liff/{clinicId}/reservation-operations/{operationKey}?intent=reservation.create`。動的 `/:id` よりstatic routeを先に登録する。LINEで存在しない更新/一括業務をこのlookupから追加しない。unknown query400、no-store。元bindingのactor/clinic/intent/subjectで照合し、成功refsに対する現在view scopeを検査する。binding未発見はscope-not-found/404。既存RunnerのResolveLookupTarget error／anchorのbinding・clinic・identity不一致、元targetの現在認可拒否はnot-authorized/403。bodyのimmutable Content/hash conflictによるoperation-key-conflict/409とは区別し、anchor取得失敗を503化する専用adapterは作らない。別actorキーの存在を公開しない。
C06 transportは succeeded/replayed=200、validation=400、version-conflict/dependency-changed/edit-right-lost/operation-key-conflict=409、operation-in-progress=202、scope-not-found=404、not-authorized=403、result-unavailable=503、not-succeeded=200。認証前401/strict body400は既存error形。非成功はstatus/keyのみ、refs/版を漏らさない。TRIMMING_ADOPTION_CHANGEDはdependency-changed、TRIMMING_BOOKING_REVISION_CONFLICTはversion-conflictへ対応。既存予約作成201 adapterはsucceeded/replayedの成功bodyだけを元shapeへ写す。Retry-Afterがあれば尊重し、202/5xx/不明/lookup not-succeededを自動新key作成条件としない。

## 11. A: course削除guard修正PR

現CountUsageとDelete.NOT EXISTSは同じscope predicateを一つのSQL builderへ寄せ、退役済detail.course_idを参照しない。T4 `trimming_detail_revisions` の同clinic/detail MAX(revision)を一つ選び、`trimming_detail_revision_lines` → `trimming_service_lines` のcourse_idを照合する。detail→appointmentsは同clinicかつappointments.deleted_at IS NULLを維持。COUNTはdistinct detail単位（同detailに複数matchinglineで増やさない）。DeleteのNOT EXISTSも同じ式。
履歴の旧revisionだけの利用を現在利用とみなさず、source_lineだけ存在して最新revisionにないものをcurrentへ数えない。履歴は保持、course softdelete後も既存FK/旧表示参照を壊さない。新たにT3予約履歴全件で削除禁止へ広げない。このPRは既存の『施術記録で使用中』の保存先置換である。
同時追加対抗は458 T2 EXCLUSIVE→course UPDATEでDelete、T4新採用writerはT2 SHARED→course SHARE→新line/revision。旧T4 writerがこの保護を使わない間は単にSQLだけ直して競合閉包済みとしない。Aに必要なT4 master利用lock adapter subsetを含めるか、B/482の接続依存として限定公開する。既存course画像のcleanup/auditとretentionを変更しない。
A: repository/sharedpredicate＋限定fixture。Bの時間schemaや全予約UIを待つ必要はないが、利用中guardの新保存先とrace条件を明記する。

## 12. B: 採用snapshot・時間consumer PR

B1共有schema/model/ERD＋両deferred triggers＋typed resolver/participant＋readDTO fixtureを同一責任ownerにまとめる。001の実適用は別承認、本作業で行わない。458のconfig公開とmasterlock adapter、457時間属性を再実装しない。
B2院内Create/CreateBatch/Update/ConfirmTentativeとLIFFValidateAndCreate/createLineReservationInTx、候補dates/times、FE採用表示/選択変更終了確認を必要subsetとして接続する。同一機能1PRで閉じられるならB1/B2を統合可能。ただし全patient/master/LINE移管を全面gateとしない。院内batchは各appointmentに独立T3、外側batch原子性・既存同時予約例外は保持。
旧LIFF detail.CourseID/SetOptionsの予約選択保存をT3へ置換し、T4実施内容へ自動昇格しない。style_request等専用記録の既存入力を一括消去しない。対象caller inventoryとcomposition adapterを成果物に含める。名前検索だけで全writer網羅済みとせず、直接GORM/SQLのT3/旧detailコース書込を確認する。
公開前の明示必要subset: 予約outer operation/receiptとLINE actorイベント対応、519の確定pet、458 T2 verifier/lock、482 T4採用lock、446 schema ownerの既存割当。業務再質問でなく技術統合残として担当を固定する。

## 13. fixture（全NOT_RUN）

course60+10+0=70、no-course区分30+10=40、0option明示保存/未設定拒否、重複option400、T2別コース/内包除外、未知T2と設定空集合、master同時改定で再確認、旧snapshot欠落unknown/notes編集不補完、予約区分改定後旧採用不変、手動終了保持、選択変更延長の既決空き枠拒否、LIFF本人actor/staff代入拒否、認可剥奪/別医院、同key同選択再送1revision・異選択拒否、header/child/audit/receipt途中失敗allrollback、遅延option追加拒否、削除guardCount/Delete同predicate・最新T4採用あり/履歴のみ・deletedappointment・同時採用、既存wire数値互換。



## 14. read DTO、単一shapeと変更履歴
既存院内/LIFF予約GETと保存応答に追加するbooked_durationはnullまたはclosed `{state,booking_revision_id,revision,base,options,total_duration_minutes}`。state=recorded/legacy_unknown。recordedはID/版string、baseはcourseの `{kind,id,version,duration_minutes}` またはreservation_typeの同4keys、optionsは `{id,version,duration_minutes}` 配列、total正整数。legacy_unknownは実在するID/版/totalのみ返し不明null、base/optionsはnull。unknownを0/[]/現masterへ置換しない。actorは公開readへ出さない。既存予約IDの数値/string移行は519/520の選定済みrolloutに従う。
変更履歴: R1のtype版不存在は最終DDL/modelの既存versionに訂正。selectionはexpected_adoption_digest付き4keysへ統一、旧fieldの無条件fallbackは撤回。外側actor/receiptを必要schemaへ固定。未確認caller閉包は実装証拠の残であり、全経路runtime完了ではない。
