# EMR-490 / 492 採択技術契約・runtime未実装

2026-10-21 公開用統合版。既決業務を継承し、追加wire/adapter方式を技術契約として選定する。本文の選定はruntime実装・OpenAPI反映・DB適用・試験・運用開始を意味しない。491への外部更新は行わない。

## 0. 照合範囲・既存選定

- EMR-490 本文および全コメント0件。EMR-492 本文および全コメント1件（0e7c1f1a-a8f0-4c70-902a-e97447269e8f）。最新 READY-AUDIT-B を旧記述より優先。
- EMR-491 本文の初回 slice は既存自院bindingを使う GET/PATCH /owners/:id と owner.basic_revision.save。新binding・登録・pet saveまで受領済みとはしない。
- GitHub main 読取基準 19855942f762f31f0403b5de0ddbc61b761fd94e。既存 api.yaml blob 9459cb23c968975417f4a44ea6f8d31ae701176f。
- [owners正本](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/spec/domains/owners-pets/lifecycle.md)、[owners設計](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/design/domains/owners-pets/technical-spec.md)、[区分正本](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/spec/screens/settings/master-reservation-type.md)、[予約設計§3.2/6.4](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/design/domains/reservations/technical-spec.md)を再利用。
- 区分停止の個別有効意図・usage_resume_pending、名称revision＋予約snapshot、全writer gateは既選定。ここで新規選択しない。

## 1. 490 操作表（今回選定する新route/DTO）

|操作|選定route/既存再利用|既存grantと用途|
|---|---|---|
|同院連絡先候補|POST /owners/contact-candidates/query、body {phone?,email?}、少なくとも一方|owners:view。選択中の認可医院のみ。GET query/logへ生連絡先を増やさない|
|common候補|POST /common-patients/candidates/query、oneOf {owner_name} / {animal_species_id,pet_name}|491/common患者の現在認可を対象bindingから個別検査（§4）。463一覧scopeを認可そのものとして流用しない。院外診療/金銭情報を返さない|
|新飼主＋複数pet|既存 POST /owners に operation_key と登録契約を追加|既存 owners:create。候補readの view を createから暗黙付与しない|
|既存自院飼主へ新pet|既存 POST /pets の owners:create を再利用。単一petの既存契約から変更しない|対象ownerの現医院・bindingを検査。複数pet一括追加を勝手に追加しない|
|確認済common患者の自院利用|POST /common-patients/use-in-clinic|必要local owner/petを作る既存 owners:create と元候補への現在view。新grantなし。これ自体で基本情報editをしない|
|登録結果lookup|GET /owner-registration-operations/{key}、intentを既知familyに限定|元clinic/actor/family scopeと現在の元操作grant。register-with-petsはowners:createによるreceipt限定回復を許容し、use-in-clinicは既定のcreate＋元候補viewを維持。成功時は結果ownerの同clinic/当該operation対応を検査|

contact queryとcommon queryを別DTO・別検索処理にする。contactは同院警告だけでcommon一致の証明ではない。名前一致のみで新規停止しない。既存 identity-links/owners/search はD19権限/group用途であり、common利用APIの代用にしない。新しい正規化・あいまい閾値・候補表示項目の公開範囲は選ばない。候補DTOは既存owner/pet許可projectionの§5.1の列挙項目だけ。現在許可医院を減らす変更は次回保存/readでも再検査する。

### 1.1 新規登録DTO・原子性

既存createOwnerRequestとpets[]を移行adapterで再利用する。追加 envelope は operation_key(UUID)、必要な identity_decisions[]。患者項目の上限・null/省略・gender・DOB検証は既存DTO/491型対応をそのまま用い、保存storage型をAPIへ公開しない。追加の臨床必須項目や自由文理由を設けない。

再送identityはC06 operation_keyだけとする。第二registration_request_id・dedup registryは設けない。同key/同Contentは同じ成功owner/pet IDsへreplayし、異Contentはconflict。in-progress/unknown中は新key救済を禁止し、元keyの現在認可付きlookupを継続する。確定拒否後のCASだけの変更はC06既存規則、業務入力の変更は旧結果確定後の新keyへ。異なるkeyの独立した新規登録を電話/メール/氏名一致で同一操作に潰さない。different_peopleとして同連絡先の複数登録は許可し、スタッフsame_person確認による新規停止は維持する。これでC06再送に必要な二重実行防止を満たすが、誤って独立した新操作を起こした場合まで自動名寄せで防ぐ保証は加えない。

identity_decisionsはスタッフが実際に候補確認した場合の候補ref・確認時basic revision・same_person/different_peopleを保持。連絡先警告が出たというだけで追加必須確認を課さない。same_personなら新owner createは止めて既存owner/new petまたはuse-in-clinicへ戻す。different_peopleは通常形式検証へ進む。

全pet検証→最外TXでowner common subject/基本revision/local owner/immutable clinic binding→pet writerのambient参加で全pet subject/revision/local/binding→既存で必要な確認履歴・receiptを一度にcommit。一匹でも失敗なら全部rollback。成功receiptにowner_id・入力順client_pet_refからpet_idへの対応・各common subject/revision IDを保存する。client_pet_refは当該request内の一意なopaque tokenで、動物名から対応を復元しない。一般登録へ新たな必須監査規則は足さない（既存PATH-OWNER-PET-REGISTERは一般audit必須ではない）。既存で必須の監査とC06成功receiptの原子性は維持する。

### 1.2 confirmation storageの実際の限界

既存 owner_identity_confirmations は owner_id と candidate_owner_id が同clinic ownersへのFK、両ID非同一、operation_id必須。未作成ownerの事前確認や他院candidateをそのままINSERTできない。成功した別人登録の同院候補確認のみ、作成ownerと候補ownerを指定して同TX記録できる。same_personで新規停止した事実のために空ownerを生成しない。他院common確認は既存common clinic bindingのconfirmed_by/at＋operation provenanceへ保存し、候補subject/basic revision commitmentをoperationに結ぶ。将来別の永続確認履歴が必要ならDDL ownerへ独立要求し、現表で出来ると記載しない。

### 1.3 use-in-clinic DTO/CASと別key重複

body = operation_key、common_owner_id、expected_owner_version、expected_owner_basic_revision_id、pets:[common_pet_id,expected_pet_version,expected_pet_basic_revision_id]、staff_confirmation:true。確認の意思を自院利用へ明示するが、新たな理由入力は要求しない。新petの基本情報をこのDTOでcreate/editしない。既存owner→pet関係をserverで解決し、候補の氏名だけで関係を生成しない。

現在認可→common owner→pet UUID安定順でlockし、対象版・関係・既存自院bindingを再照合。immutable bindingの(clinic,common subject)と(clinic,local ID)uniqueを再利用。同subject別keyは既存binding/local IDsへ収束し、新local患者を増やさない。mapping不明/複数は停止し旧医院別基本列へfallbackしない。診療・請求・LINE/groupは移管/統合しない。

### 1.4 C06、元target、lock、491必要subset

新規登録の元targetは生成前owner IDでなく、認証clinic＋actor＋owner.register-with-pets＋operation_keyで一意のC06 bindingそのもの。claimと同TXのimmutable owner metadataでこれを固定し、lookupは成功receipt不在でも元scopeを解決できる。use-in-clinicはcommon owner UUID＋pet集合digestを固定。actor/clinic/owner_kindはbodyから採用しない。C06 contract.go のOwnerKindはstringで固定enumではない。本契約はowner_kind=owner、intent=owner.register-with-pets / owner.use-in-clinicを選ぶ（既存491 owner.basic_revision.saveとは別family）。既存POST /petsの接続はpet write-ownerを維持し、本契約でownerへ移管しない。

shared RequestのVerifiedCommand限定拡張によりimmutable typed commandを単一factoryで生成し、Contentはtarget refsとcanonical payload commitment、Preconditionsはexpected versionsに分離。PII payload fingerprintはowners/予約設計の秘密鍵付きcanonical HMAC＋鍵versionを再利用する。生電話/住所をContent/log/cacheへ複製しない。claim metadataのみの短TXの後、業務TX→graph gate→receipt→§8の具体的なowner/pet/common順→binding/改訂。共通UUID改訂だけの491初回sliceへgraph gateを一律追加しないが、登録/local FKや予約関連に影響するwriterは予約§6.4へ参加する。逆順が生じる場合、行取得後のgate追加ではなく最初からやり直す。

491から必要なのは(1)認可付きcommon候補readerとcurrent basic revision reader、(2)新subject＋最初のtyped revision＋immutable clinic bindingをambient TXで返すparticipant、(3)既存subject利用の一意binding解決、(4)DOB/Cygramの同revision組成とnull規則、(5)型付きOwner/PetResponse変換、(6)deferred common_basic_publishとのreceipt版整合。初回既存owner saveだけの公開は(2)(3)の完成を意味しない。deferred triggerとappの二重current pointer/version更新は禁止。491への外部writeはせず、490側依存表として引き渡す。

## 2. 492 設定wire（今回選定する追加差分）

既存 /masters/reservation-types と /masters/reservation-type-groups のGET/POST/PATCH/DELETE/reorderを再利用。既存grant master-reservation-type の view/create/edit/delete を各現routeどおり維持。追加restoreは POST `/masters/reservation-types/{id}/restore` と POST `/masters/reservation-type-groups/{id}/restore` の2 routeを明示し、いずれも master-reservation-type:edit。2026-10-21 03:55 UTC、本人 Sentinel_fc62ffe7af188191b2e001d0f3d15036 の「ok」による新たな業務権限確定であり、旧routeに存在した契約と偽らない。復元後無効・利用再開は別操作を維持し、新grantは増やさない。既存LINE管理 /clinics/{clinic_id}/reservation-types も同じwriterへadapter接続し、旧category/color/occupation経由の別正本を残さない。

全writeにoperation_key、更新/削除/restoreにexpected_version。PATCHの変更項目だけContentに束縛し、省略保持/null消去を既存型に沿って区別。削除/restoreは専用intent、DELETEはversion/keyをJSON bodyとしてOpenAPIに明記する。reorderは順序付きID＋期待版集合を受け、当該scope集合との一致も検査する。

種別create: name、business_kind(general/trimming/inpatient/hotel)、duration_minutes、individual_active、group_id(nullable)、§5の明示allowlist。durationは必須5..480の5刻み。旧category・種別colorは受理しない。group create: name/color/is_active/既存sort_order。色値/presetを新たに発明しない。

種別PATCH changes: name、duration_minutes、individual_active、group_id、§5の明示allowlist。business_kind変更と未設定補完は別intentへ分ける: POST /masters/reservation-types/{id}/business-kind/change と /business-kind/confirm、body operation_key/expected_version/business_kind。changeは過去未来いずれか予約が一件でもあれば拒否。confirmは未設定のみで既存の確認補完条件を検査し、既存分類変更の抜け道にしない。

read DTOは version, business_kind, duration_minutes, individual_active, usage_resume_pending, deleted_at, group_id, group_version, effective_active, name_revision を既存許可projectionへ追加。effective_activeはserver導出でclientは書かない。group readはgroup version/is_active/deleted_at。削除履歴を読む管理routeは既存viewを現在検査し通常候補へ削除行を戻さない。

### 2.1 group/個別意図の既選定遷移

- group無効化はgroup is_activeだけ変更、子のindividual_activeを上書きしない。
- group復元はdeleted_at解除かつis_active=false。子の削除解除/個別一括有効化なし。
- group明示有効化時は子の既存個別意図を尊重。group無効中にも個別意図変更を保存でき、有効化時に全子の再クリックを要求しない。
- 無効/削除groupから所属変更した種別はusage_resume_pendingを立てる。通常利用可能な種別の所属変更では新たに立てない。所属変更のみで再開しない。
- 種別復元は個別無効。明示individual_active=trueで再開待ち解除。ただし無効group・削除等の他gateは残る。子削除をgroup操作で自動復元しない。
- effective_active = 個別有効 && !種別削除 && !再開待ち && (groupなし || group有効かつ非削除)。旧unknown flagは勝手にfalse補完しない。

### 2.2 renameのT/S・予約接続公開interface

rename familyはPATCH changesにnameキーが存在する場合に固定してdispatchする。保存済現在名との比較でfamilyを切り替えない。同key再送時に現在値が既に新名でも元familyを維持する。nameが現在値と同じ新規操作はname revision/snapshotを増やさず、同command内の他changesだけを更新する。bodyにTや対象予約リストを持たせない。gate/receipt/master lock取得、現在認可/CAS後にDB clock_timestampを一度読みTとする。Tは成功した名称改訂の論理的な変更確定時刻として採る技術選定で、HTTP受信・TX開始・実際のDB commit時刻のいずれとも同義ではない。既存正本§3.2は同gate内の変更確定時刻を要求するがclock取得位置までは選定していない。gate取得後の最終認可/CAS成功直後に一度採時し、S>Tの集合と名称改訂へ同じ値を用い、全体commit成功によって初めてこのTが有効になる。rollback時には確定時刻も改訂も存在しない。Tとcommitの間にSを迎える予約もS>Tなら新名になる点を明記し、commit直前の壁時計を境界とする仕様へ無断に読み替えない。この線形化定義は今回の追加技術選定であり既決の業務境界S>T/S<=Tを変更しない。master name_revisionを増分し immutable name revision(previous_name/display_name/changed_at=T/operation/actor)を追加する。

同じambient TXで予約owner participant ApplyTypeRename(tx, clinic_id, type_id, previous_revision, new_revision, new_name, effective_at=T, operation_binding_id)を呼ぶ。participantは同gate内で当該typeの予定開始S>Tを検索し、appointment ID安定順に保護しsnapshot/採用revision/出所と予約version・必要変更履歴を更新する。S=T/S<Tは変更しない。集合をclient previewだけで固定せず同gate再検索。master+全対象snapshot+必要監査+成功receiptを原子commit。途中失敗でmasterだけ成功にしない。

大量対象でも今回の既選定同期契約をasyncの部分完了へ勝手に変えない。負荷・timeout時は全rollback/同key照会、実測による変更は別技術レビューとする。閲覧readerは保存snapshotを返し、現在名JOINで過去を再構成/巻戻ししない。

予約の許可された未来への日時変更は既存468更新route/DTOを使い、新rename専用予約APIを増やさない。予約owner内部 ResolveCurrentTypeName(tx,clinic_id,type_id) が現在master name_revision/nameを取得し、その更新で採用し前後snapshotと時刻を履歴化。旧予定開始のrevisionを時間検索して復元しない。日時変更自体の許可条件・臨床/公開条件は468既決のままで新しく制限しない。

### 2.3 C06/元target/lock/段階引渡し

settingsの更新元targetはclinic＋type/group ID。新規作成は認証clinic/actor/intent/keyのC06 bindingを元targetとして固定し、receipt不在でもlookupできるようにする。lookup GET /reservation-master-operations/{key}?intent=既知enum、現在master view、binding原targetを解決してからURL整合/結果開示。POST/PATCH再送は元write grantを維持。別keyで同じrenameを繰り返してrevisionを増やす救済はしない。

lockは既選定: metadata-only claim TX終了→業務TX→reservation-graph:write:v1→receipt→§8のgroup/type順→対象appointment ID安定順。group所属変更ではold/new group ID安定順を含む。現在認可/CAS・classification使用有無・group依存versionをgate内再照合。別キー並行renameやgroup変更はCAS失敗を返し、再取得・比較後の新操作にする。

PR-A設定: type/group CRUD/restore/reorder・分類intent・version・個別意図とread DTO・C06 lookup・名称revisionを公開する。予約snapshot更新participantが未受領の間は予約に作用するrenameを成功させる独立writeにしない。interface/mock/型作業は進められるが利用開始はcontract stateを保留。空のparticipant成功stubは禁止。

PR-B予約接続: 上記ambient participant、予約read snapshot、新規予約snapshot、許可未来リスケcurrentname、全writer gate/切替、旧admin/LINE route adaptersを接続。受領API schemaとcommit SHAをPR間の入力に固定する。設定PRの存在だけで予約接続完了としない。

### 2.4 新規初期値と旧LINE管理writerの切替

parent_idは本canonical create/PATCHのallowlistに含めず、指定は400。親子分類は未決であり、既存保存parent_idの削除・再解釈・新UIによる編集を今回採用しない。

新typeの保存はusage_resume_pending=false、name_revision=1を明示し、同TXで名称revision1（previous_name=NULL、display_name=新name、同operation/actor/採時）を作る。type createのcommitted_versionsは{master,name_revision}。既存unknownフラグ/名称版に架空初期値をbackfillしない。最終001ではcategory/colorはDROP済みであり、旧defaultを新列へ復活させない。

LINE管理の/clinics/{clinic_id}/reservation-typesは旧wireの無変換互換を約束しない。新管理FEはcanonical masters routeへ移行し、create business_kind/key、全更新key/CAS、DELETE JSON body、status→PATCH changes.individual_active、sort-order direction→最新GETから得た完全items[{id,expected_version}]のreorderを送る。serverはkey/期待版/分類/並び順を推測生成しない。旧管理write endpointは切替時にwriteなしの既存非対応エラーへ閉じ、再読込・新FEへ誘導する。新handlerと旧caller停止を同じactivation条件にし、旧writeが迂回できる状態では有効化しない。旧readと通常LIFF飼主予約routeは本切替の対象外。

### 2.5 旧患者coverageの有効化条件

contact candidateのrequired common refsは既存患者のbinding移行が検証済みのscopeでのみ成立する。最終001では旧owners基本列もDROP済みのため、未bindingを旧氏名/phoneへfallbackしたり架空revisionを生成しない。491から必要な限定受領物に、対象医院の全非削除local ownerに対する一意common binding・読取可能な固定基本版・連絡先候補索引のcoverage検証を追加する。未binding/重複/取得不能が残るscopeは候補読取を完全成功として返さず、新登録activationを保留する。空候補として黙って除外しない。これは確認済み移行実行や全491完了を意味せず、必要scopeの受領条件である。

## 3. 共通HTTP・未実施受入

新C06 consumerは既公開mappingを再利用: succeeded/replayed/not-succeeded200、operation-in-progress202、result-unavailable503、key/version/dependency/edit-right409、validation400、scope-not-found404、not-authorized403。非成功はstatus/keyのみ、pre-C06のみ既存error:string。既存他routeを一括変更しない。business ruleの通常validationとscope leak抑止を混同しない。応答不明は元intent/keyでlookup、結果不可から新keyで業務を再実行しない。

未実施fixture: 490の二匹目失敗/応答断/同電話別人/same_person停止/同subject二院利用・同院競合/確認後版変更/失権/旧値fallback禁止。492のS=T・同日S>T・S<T/後日read/未来再予定/renameと予約作成両順/group停止中個別変更/restore→明示enable/所属変更再開待ち/異key競合/途中snapshot失敗。全て仕様・試験設計のみ、実行合格を記載しない。

### 3.1 metadataの選定形

490: owner_registration_operation_anchors(clinic_id bigint, operation_binding_id bigint, cygram_rule_version text)、owner_use_operation_anchors(clinic_id bigint, operation_binding_id bigint, common_owner_id uuid, common_pet_ids uuid[], pet_set_digest text)。492: reservation_master_operation_anchors(clinic_id bigint, operation_binding_id bigint, target_kind text type/group, target_id bigint nullable)。新規familyではtarget_id=nullでC06 bindingが元scope、既存対象familyではtarget_id必須。各表PKはoperation_binding_id、binding+clinicへのimmediate複合FK、clinic RLS、UPDATE/DELETE拒否でimmutable。target業務行FKをclaim時に追加してlock順を壊さない。生成後IDsはreceiptへ保存しanchorを上書きしない。

claim winner hookだけがbindingとanchorを同じ短TXに書く。loserは原anchorをreadし欠損を入力targetから補修しない。業務target読取/lockはclaim hook内で行わない。第二identity・request registryは作らない。

492 owner_kind=reservation、familyはreservation-type.create/update/rename/delete/restore/reorder/business-kind-change/business-kind-confirm、reservation-type-group.create/update/delete/restore/reorder。本表は今回の技術選定値で、既存runtimeに実装済みという意味ではない。PATCHでnameと他項目を同時変更した場合はrename familyの単一commandへ全changesを含め、一部だけ別TXに分割しない。lookupのintent enumからこれらへserver mapする。既存wireの型は維持し、clinic_idを含む既存numberのIDを一括string移行しない。正規decimal stringはhash projectionとC06のstring mapでの表現であり、公開DTO全体の型ではない。UUIDはcanonical文字列、request objectはclosedとする。

## 4. common候補の現在認可とContextGate

既決根拠はowners lifecycle「担当者・権限・前提」「common-owner-pet-20261017」「AC-owners-pets-other-clinic-use/common-denied」。所属医院・操作権限・対象患者認可を維持し、権限内の他院患者だけ利用する。491第一sliceは自院immutable binding→common UUIDの認可であり、全院common search権限を公開していない。463のowners一覧医院ID集合、共通UUIDの存在、生成医院anchorだけで候補開示を許可しない。

選定adapter: CanReadCommonCandidate(principal, subject)は、対象subjectのimmutable clinic bindingのうち少なくとも一つについて、現在の当該clinic所属＋owners:view＋既存local owner/pet対象scopeの認可がすべて成立するときだけ許可する。pet候補はそのpet/ownerの対応をserverで検証する。clientのclinic_idsやcurrent basic revisionのsource_clinicだけから許可しない。結果は許可済common basic projectionだけで、他院の診療/金銭や他bindingの一覧を付随開示しない。save/lookupにもこの個別判定を再適用する。これは既決原則を実装する技術predicateであって、新しいcross-clinic閲覧grantではない。491第一sliceを全common検索契約と偽らない。このpredicateの拡張は本契約に含まれない。

### #784 ContextGate接続を必要subsetへ

PR #784 の追加補正（review 4332795869、共有R4採択契約）に従い、489のnonbreaking optional ContextGate.GateContext(ctx)(context.Context,error)を、graph gateが必要な490登録/自院利用・492全master writerの依存にする。現error-only GateでAcquireReservationGraphWriteGateのderived ctxを捨てない。RunnerはContextGateがあればそれだけ、なければ従来Gateだけを1回実行し、nil ctx/ambient Tx pointer不一致を拒否する。returned ctxをbinding/receipt/認可/savepoint/Apply/ambient owner-pet又はrename participant/監査/結果INSERT/同TX unknown記録へ一貫して渡す。下位Assertを維持し、marker自作・mutable fieldにctx保存・nested TX・DetachTx・終了済ctxのlookup流用は禁止。491初回のgraph不要sliceへ不要なgateを拡張しない。489実装受領前に成立済みとしない。

### 根拠区分

既存選定: 同院連絡先警告とcommon候補の分離、別人登録可/同一確認停止、原子登録、C06元key再送、T/S厳密境界、名前snapshot、個別意図/復元無効。
本人追加確定: 492復元は既存master編集権限（2026-10-21 03:55 UTC上記回答）。
今回追加技術選定: route/closed DTO、common候補個別認可predicate、metadata anchor、rename Tの線形化位置、ambient participant、以下のhash/receipt/audit/lock。仕様選定と実装未完了を区別する。

## 5. closed DTO・canonical commitment

以下は本契約が接続するwireの全allowlist。未知field・未知enum・重複JSON key・複数JSON値・不正UTF8を400で拒否する。application/jsonのみ。既存患者fieldのvalidationを緩めず、新臨床必須条件・新公開fieldを増やさない。scope/actorはserverで確定。登録に限り既存clinic_id任意指定は保持し、現在所属と対象clinicのcreateをserverで検証してからC06 scopeを作る。bodyのclinic_idを認可として信用しない。

### 5.1 490 DTO

- contact query: phone?, email?のみ。null/空値だけの要求は拒否。既存phone/email validatorを再利用。返却rowはowner_id, common_owner_id, owner_name, owner_name_kana, subject_version, basic_revision_idだけ。連絡先自体をechoしない。
- common query: owner_nameのvariant、またはanimal_species_id＋pet_nameのvariantのみ。返却rowはcommon_owner_id, owner_name, owner_name_kana, owner_version, owner_basic_revision_id, pets[]。pet rowはcommon_pet_id, pet_name, animal_species_id, pet_version, pet_basic_revision_id。追加住所/電話/診療/LINE情報は返さない。pet検索時には一致したpetだけ、owner検索時のpetは個別認可済だけを返す。
- POST /owners: operation_key, identity_decisions?, clinic_id?, owner_name, owner_name_kana?, birth_date?, company?, postal_code?, address1?, address2?, home_postal_code?, home_address1?, home_address2?, phone?, company_phone?, email?, remarks?, is_dangerous?, discount_rate?, membership_type?, dm_preference?, pets[]。
- pets各要素: client_pet_ref（UUID、request内一意）, name, animal_species_id, name_kana?, breed?, color?, blood_type?, microchip_number?, gender?, status?, birth_date?, weight?, neutered_date?, neutered_status?, acquisition_type?, danger_level?, danger_reason?, food?, environment?, insurance_id?, remarks?, name_origin?, meeting_story?。旧createPetForOwnerRequestのfieldを列挙したもので、新しい生死変更ルートを設けず、死亡/復活の既存専用規則を迂回しない。
- identity_decisions各要素: candidate_owner_id, expected_basic_revision_id, determination(same_person/different_people)。同院候補だけを扱い、同一candidate重複は拒否。他院common確認はuse-in-clinicで行い、この配列へ他院local IDを送らない。expected_basic_revision_idはPreconditionsへ、candidate ID/determinationはContentへ入る。
- use-in-clinic: operation_key, common_owner_id, expected_owner_version, expected_owner_basic_revision_id, pets[], staff_confirmation(trueのみ)。pets要素はcommon_pet_id, expected_pet_version, expected_pet_basic_revision_id。common_pet_id集合の重複は拒否。ownerだけの自院利用は空集合可、petの新規基本情報は受理しない。
- GET lookup queryはintent=register-with-pets/use-in-clinicのみ。POST /petsは既存単独追加契約のまま、この仕様から未指定の一括APIを派生させない。

### 5.2 492 DTO

- create type: operation_key, name, business_kind, duration_minutes, individual_active?, group_id?, expected_group_versions[], description?, sort_order?, reservation_display_name?, max_concurrent?, short_name?, show_short_name?, reservation_visible?, reservation_comment?, reservation_image_url?, reservation_day_option?, is_internal?。
- PATCH type: operation_key, expected_version, expected_group_versions[], changes。changesはcreateのname〜is_internalのうちbusiness_kindを除いたfieldのみ、空changesは拒否。group/max_concurrentは明示null消去、他fieldのnullは既存nullable契約以外拒否。旧clear_group_id/clear_max_concurrentは新wireで受理しない。旧adapterがある場合は存在フラグ付きnullへ一回だけ変換する。
- create group: operation_key,name,color?,sort_order?,is_active?。PATCH group: operation_key,expected_version,changes{name?,color?,sort_order?,is_active?}、空changes拒否。
- type/group DELETEとrestore: operation_key,expected_versionのみ。restore payloadからis_activeを受け付けない。
- business-kind change/confirm: operation_key,expected_version,business_kindのみ、権限はmaster-reservation-type:edit。business_kindはgeneral/trimming/inpatient/hotelだけ。
- reorder: operation_key,items[{id,expected_version}]。配列順が保存順、ID重複不可。clinicの非削除type/group集合をgate内で再検索し同集合でなければdependency-changed。意図した順序をID順ソートしない。
- expected_group_versionsは今回参照する旧/新groupの{id,version}集合、ID昇順、重複不可。create typeのgroup_idが非NULLならexpected_group_versionsは選択groupの{id,version}厳密1件を必須とし、groupなしでは空集合を必須とする。既存group版が変われば再確認する。groupの状態・権限を版が同じだけで省略しない。
- 個別activeのcreate省略は既存is_activeのtrue defaultへ正規化する。group createも同様。予約公開用fieldは既存意味のままで、本変更からLINE公開範囲/開始枠ルールを拡張しない。category/種別color/occupation関連fieldは拒否。

### 5.3 hash / Preconditions / captured payload

command factoryは後記の共有VerifiedCommand限定拡張を前提とし、strict decode・既存validation・存在/nullの区別を終えたtyped valueをdeep copyし、Request内のimmutable VerifiedCommand、Content、Preconditionsを一度に作る。同keyの保存済Content fingerprintとincoming commandの不一致はoperation-key-conflictでzero-write。fingerprint一致後、Apply冒頭でもreq.VerifiedCommandからcommitmentを再検査し、request内payloadの不一致又は業務依存の変化はdependency-changedでzero-write。期待版だけの不一致はversion-conflictとする。これらの詳細は元targetの現在認可を通した後だけ返す。mutable map/slice/後から再parseしたpayloadは使用しない。

payload commitmentはlowerhex HMAC-SHA256(K_version, UTF8(JCS(projection)))。鍵の実生成/設定/交換は本作業に含めない。payload commitment用のkey versionはcommand_schema単位で固定し、再送では同schemaの固定versionを使う。稼働中keyの自動選択で同内容digestが変わる実装は禁止。C06外側fingerprintの鍵version管理とは区別し、旧schemaのbindingが残る間は対応commitment鍵を照合可能に保つ。鍵廃棄/変更は別の移行設計なしに行わない。projectionはschema=owner-register-v1 / owner-use-v1 / reservation-master-v1、intent、検証済対象refs、全変更fieldのpresence/null/value。ID/版は正規decimal文字列、UUIDはcanonical小文字、日付はYYYY-MM-DD、時刻はUTC RFC3339Nano。十進小数は入力十進値を正規文字列化しbinary floatの丸め値をdigestにしない。文字列は既存validatorが規定する変換だけ、空白や漢字の独自正規化なし。集合はID順、pets登録入力順・reorder順は順序を保持する。

operation_key/actor/clinic/server時刻/CAS値はprojectionに入れずC06 scope/Preconditionsへ分離。登録で入力したcandidate決定・serverが選定し固定したCygram rule_versionはContentの意味に含む。Cygram rule_versionは初回claimのowner_registration_operation_anchors.cygram_rule_versionにNOT NULLで保存し、bindingと同じ短TXで確定する。DOB省略/nullでもこのschemaの選定rule版を記録するが、結果は既存DOB契約どおり未設定とする。birth_date派生結果自体はclient入力でない。Contentはcommand_schema, command_digest, digest_key_version と必要root refだけのstring map。owner-useはcommon_owner_id, pet_set_digestを追加、master既存操作はtarget_kind/target_id、新規はtarget_kind、reorderはtarget_kind/ordered_ids_digestを追加する。大配列・PII原文をContentに置かない。C06の小field長制約を維持する。

Preconditionsは登録candidate expected_basic_revision、use-in-clinic owner/pet期待版、master expected_versionとgroup期待版、reorder各期待版。成功済同keyは古いCASで再writeせず保存結果を返す。確定拒否後の同Content/CAS再確認と業務内容変更を区別する。

## 6. success refs / 失敗時anchor / RLS

保存内部C06 Result.ResultRefs/CommittedVersions/EffectSummaryはstring mapの既存型を維持する。公開ResultResponse allowlistは既存operation/dto.goどおりstatus, operation_key, result_refs, committed_versions, recorded_atだけ。result_refs/committed_versions/recorded_atはsucceeded/replayed時だけ（recorded_atは保存値がある場合）、非成功はstatus/operation_keyだけとする。EffectSummaryは保存内部専用でnever sent。以下のsummary表記は内部保存値であり公開bodyには含めない。以下のindexは0始まりdecimal。外部payloadにPIIを複製しない。

- register: refs={owner_id,common_owner_id,owner_basic_revision_id,pet_0_id,pet_0_common_id,pet_0_basic_revision_id,pet_0_client_ref,...}、versions={owner_subject,pet_0_subject,...}、summary={pet_count}。pet順は入力順。...は同じ定義のi=0..pet_count-1だけであり任意fieldではない。
- use: refsはregisterと同じowner/pet ID・revisionキー（client_refなし）、petsはcommon UUID昇順、versionsは同じsubjectキー、summary={pet_count}。既存binding再利用でも同じlocal IDへ収束。対象版は実際に確認した版、common基本情報を更新しない。
- type/group単独: refs={reservation_type_id}又は{reservation_type_group_id}、versions={master}。type createはversionsにname_revision=1を必須追加する。renameはversionsにname_revision、summaryにsnapshot_updated_countを追加。renameが更新した全appointment IDを大きなreceiptへ詰めず、操作紐付けのaudit行から照合する。
- reorder: refs={target_0_id,...}、versions={target_0,...}、summary={target_count}。定義域はi=0..target_count-1のみ。
- committed subject versionは491のdeferred publish遷移と同じ値を用い、commit前の古いpointerを採らない。commit/deferred trigger失敗時はResultもrollback。result mapの実helper制約と最大requestサイズの適合は受領試験対象であり、未対応時に結果を切り捨てない。

metadataは§3.1に従いbindingと同短TXで生成し、失敗/unknown/未成功にも残す。operation_binding_id/clinic_idはNOT NULL、operation_binding_id PK、UNIQUE(operation_binding_id,clinic_id)、FOREIGN KEY(operation_binding_id,clinic_id)→operation_bindings(id,clinic_id) ON DELETE RESTRICT、CHECK clinic_id>0。owner_useのcommon_owner_id/common_pet_ids/pet_set_digestはNOT NULL、pet配列はUUID昇順・重複なし（空配列可）、digestはlowerhex64で当該配列commitmentと一致させる。pet IDsを不可逆digestだけにせず元target metadataへ保存し、未成功lookupでも各petの現在認可を再検査できるようにする。master target_kindはtype/group、target_idはNULLまたは正値。familyがcreate/reorderならtarget_id=NULL、他familyなら非NULLの整合をclaim factoryで検査する。reorder原targetは認証clinicの当該master集合で、成功ID集合から後付けで推測しない。

RLSはENABLE/FORCEし、USING/WITH CHECKは同じambient transactionで検証設定したclinic_idとの一致。API実行roleにBYPASSRLSを付与せず、HTTP bodyでRLS contextを設定しない。immutable triggerでUPDATE/DELETEを拒否し、通常writerにはINSERT/SELECTだけ。claim hook以外からINSERTしない。本人/権限の現在検査はRLSのclinic一致だけで代用しない。cross-clinic common読取は各認可済bindingの既存readerを使用し、anchorのRLSを全院へ広げない。

lookupはscopeのbinding→保存anchor→元targetの現在認可（新規登録は元actorの現在owners:createによる限定receipt、その他は既定view）→保存receipt。新規未成功は元認証clinic/actor/familyが原target。成功時は結果IDが同clinicに実在し当該operationの作成/利用結果と一致することも検査する。useはanchorの原common ownerと全common_pet_idsの現在候補認可、masterは元type/group（削除済でも権限検査可能な原行）又はclinic集合の現在master:viewを検査。別URL/別対象で元結果を漏らさず、pre-C06不一致は現在認可後の既存409 ErrorResponse、C06保存済Content fingerprint不一致はoperation-key-conflict、fingerprint一致後のApply依存変化はdependency-changed、期待版CAS不一致はversion-conflict。欠落anchor/解決不能は非開示で止め、成功receiptから別targetを捏造しない。

## 7. audit sinkと失敗時原子性

一般owner登録へ新たな全件audit義務は追加しない。different_people同院確認があるときだけ既存owner_identity_confirmationsへ、clinic_id/新owner_id/candidate_owner_id/determination/operation_id/confirmed_by/confirmed_atを保存。evidence={schema:owner-identity-confirm-v1,candidate_basic_revision_id}。same_personは登録拒否で空owner/成功receiptを作らない。use-in-clinicはimmutable bindingのconfirmed_by/confirmed_at/operation_idが一次provenance。

既存設計で監査を要求するuse-in-clinicと492変更、およびrename/reschedule snapshot履歴のsinkはaudit_logsへ選定する。staff操作は認証clinic_id,actor_type=staff,actor_id、created_at=DB時刻。metadata={schema:owner-use-v1又はreservation-master-v1,operation_id,operation_key,intent}。actionはowner.use-in-clinic、または§3.1のmaster intentと同じ文字列。resourceはowner / reservation_type / reservation_type_group、resource_idはそのlocal ID。old_value/new_valueはIDsと前後version・変更fieldだけ（owner useはbinding ID集合とcommon UUID集合、基本情報原文は含めない）。type/group削除復元には前後deleted_at/is_active/usage_resume_pendingを含める。名称変更の前後名は履歴要件として保存し、他の患者情報を混在させない。

rename participantは変更したappointmentごとにaction=reservation.type_name_snapshot_change, resource=appointment, resource_id=appointment_id、old_value/new_value={display_name,name_revision,source,appointment_version}、metadataにoperation_id/type_id/effective_at=Tを持つ監査行を同TXで追加する。許可未来リスケもこの同じactionで元予約operation_idに束縛し、前後予定開始をmetadataへ保存する。type name revisionの既存previous_name/display_name/changed_at/changed_by/operation_idも同TXで追記。これによりreceiptの件数と操作紐付け監査集合を照合できる。

reorderは各対象IDの前後sort_order/versionを1対象1audit行として同じoperationへ束縛。監査が必須のoperationはsink INSERT失敗でdomainとreceiptすべてrollback。GET/replayはaudit行を追加しない。claim metadataや失敗ログを業務成功監査へ見せない。外部HTTP・LINE通知・tag syncをこのTXから起動しない。

## 8. lockの具体参加順・未実装境界

全操作のclaimは短いmetadata-only TXで完了し、claim lockを保持して業務gateを後取りしない。業務TX先頭でContextGateを取得→同ctxでC06 binding NOWAIT→現在grant/原target再検査。490登録/利用、492masterはreservation-graph:write:v1へ参加する。491既存basic改訂のみのno-op gateを変更しない。

490新規は同院candidate owner ID昇順→そのcommon owner UUID昇順→必要species/insurance FKを各table内ID昇順に確認→新owner/local-binding/revision→入力順pet writer。useは既存local owner/petをowner ID昇順→pet ID昇順で確認し、common owner→common pets UUID昇順→clinic bindingsのowner/pet順。未存在local rowをlockしたことにせず、common subjectの保護と既存uniqueで同院二重作成を防ぐ。既存subject基本改訂とのlock整合は491 participantが同じ順序を受領する。異なる順序を要求する既存writerを発見した場合はgate後の下位順を勝手に反転せず統合調整し、未接続入口の利用開始を止める。

492はold/new group ID昇順→対象type ID昇順→appointment ID昇順。group操作はgroup ID→所属type ID昇順で状態/集合を検査。reorderは対象kindのID昇順で保護し、保存順序は入力順を用いる。通常予約のstaff/clinic/pet/lifecycle上位資源をmaster側で後取りせず、graph gate下の既存master→appointment participant順を使う。rename participantは別TX/再claim/外部作用なし。全writerがgateへ参加するまで「競合安全実装済み」としない。

未実装の受領物は、489 claim metadata hook/ContextGate/結果DTO、490/491のcommon readerとambient登録/利用writer、492のmaster handler/reader/rename participant、型生成とOpenAPI更新、既存DDLに対するanchor追加の統合、実001/RLS/deferred trigger/並行性/応答断試験。これは技術契約の選定残ではなくruntime実装・検証残。試験NOT_RUN、外部書込なし、491更新なし。

## 9. 初回claimと再構成・公開結果の最終補正

既存bindingがある登録要求では、認証clinic/actor/intent/keyから元anchorをreadし現在認可後、そのcygram_rule_versionでcommandを再構成する。初回だけserverの選定ruleを候補にし、claim winnerがbinding＋anchorへ同時保存する。競合loserはwinner metadataを読み直し、その版でContent/fingerprintを再構成して比較する。候補版で先に異内容と判定してはならない。claim hookはimmutable metadata保存だけで、業務rowのlock/基本改訂保存/独立TXを行わない。新しいregistryは追加しない。

未成功binding・プロセス再起動・現行rule更新後も元版を使う。元版の算法/固定表を利用できない場合は処理を止める（result-unavailable、結果詳細なし）。最新ruleへ黙って置換した新revisionや新key救済を行わない。rule版の変更を業務内容変更として新操作にするのは旧結果が確定し、改めてその操作を開始するときだけ。初回claim metadata取得を含むfactory/489受領interfaceはこの動作を実現する必要があり、現在の実装成立を宣言しない。

公開型の区別: 既存createOwnerRequestのclinic_id・animal_species_id・insurance_id、既存masterのgroup_id/sort_order/duration_minutes/max_concurrentは既存number型を維持する。新expected_versionは対応する公開readerのversion型に合わせる。hash projectionだけID/版/十進値をcanonical decimal string化し、C06 result mapは既存string値のまま。payload型変更や全APIのID string化を本契約に含めない。

判定順: writeは認証＋入口の現在write grant→binding/元anchor解決→元target現在認可→保存済Content fingerprint比較（不一致operation-key-conflict）→同Contentならreplay/unknown判定→Applyの現在依存照合（変化dependency-changed）→期待版CAS（不一致version-conflict）。実取得順はC06/graph lock契約を守る。lookupは前述family別の現在認可（登録は同actorの現在owners:createによるreceipt限定）を使い、POST再送は入口のwrite grantを省略しない。認可失効をkey/CAS詳細で漏らさない。EffectSummaryのpet_count/target_count/snapshot_updated_countは内部でreceipt整合を検査する値で、公開ResultResponseへ自動marshalしない。


## 統合レビュー共通補正：shared Request の限定 immutable payload

現operation.Participantはreq.Content.Fields由来のwriteのみ許し、private captured commandを許さない。このconsumerには全文を短いContent mapへ格納できないため、共有Requestへoptional VerifiedCommandを追加する限定deltaを必要subsetとして選定する。VerifiedCommandはoperation package所有のimmutable値で、constructorでschemaとcanonical command bytesをdeep-copyし、schema別pure projectorが求めたContent全keysとの一致を検査する。getterはdetached copyのみを返す。Runnerはclaim/fingerprint前とApply直前に同じrequest bytesから再検証し、Applyはreq.VerifiedCommandだけからtyped入力を再構成する。participantへのDTO capture、mutable pointer、別parse結果によるwriteは禁止。CASはPreconditionsのままである。

payloadはrequest memoryのみ、binding/receipt/logへ保存しない。既存consumerはVerifiedCommandなしのContent-only方式を維持する。HMAC/各digestのschemaと鍵版は本文既定どおりで新正規化をしない。489で共通contract/runnerの限定interfaceを一箇所だけ実装するまで、このconsumerを現C06のまま実装可能と扱わない。これは親による技術選定であり、本人がAPI/物理方式を直接承認したとの記載ではない。実装/試験NOT_RUN。
