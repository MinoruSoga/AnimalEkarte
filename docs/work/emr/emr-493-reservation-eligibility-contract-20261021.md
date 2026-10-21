# EMR-493 医院別指名停止・個人対応区分：①選定契約

2026-10-21。親レビュー採択の技術契約。文書のみ。API schema登録、DDL/model変更、runtime実装・DB適用・試験は未実施（NOT_RUN）。main 755f77abd2e6d63a2fc608fcfe6a518a61afcfffの確定業務仕様へ合わせ、現コードの動作を業務仕様へ昇格させない。

## 0. 根拠と既存基盤

正本 [スタッフ設定](../../spec/screens/settings/master-staff.md)の positive-reservation-capabilities / capability-booking-guard / clinic-departure-booking-guard。A院指名停止の前にA院の未来/現在診療中の保護予約を手動解消、B院勤務/予約/過去履歴保持、スタッフ全体無効化/所属削除/ログイン失効/自動予約変更なし。職種/権限/能力を混同しない。
既存model/account.go StaffClinicAssignment: nomination_paused nullable bool(json:"-")、capabilities_configured_atあり、版なし。staff_reservation_capabilitiesのpositive junctionは既存。staff_service_permissions.goは能力Get/Setを既存reservation repositoryへ委譲。handler.goのGET/PUT /masters/staffs/:id/capable-reservation-typesはMasterStaff:view/editで実在する。staff_request.goはreservation_type_ids配列のみ、最大50、key/CASなし。
既存staff_clinic_assignment_service.goの所属集合置換は、この医院別指名停止とは別機能。所属解除を代用しない。現NominationPaused参照はmodelだけで公開DTO/保存経路の実装済み根拠はない。原稿調査時api.yaml全体fetchは空応答だったためAPI全体不在を断言しない。本稿は実在handler/DTOと追加API登録義務を区別する。

## 1. 選定route/DTO

実prefix /api/v1、以下相対path。選択clinicは認証context、本文clinic/actor不可。既存MasterStaff:view/editを使い新grantなし。
GET /masters/staffs/{staff_id}/reservation-eligibility
PATCH /masters/staffs/{staff_id}/reservation-eligibility
GET /masters/staffs/{staff_id}/reservation-eligibility-operations/{operation_key}
GETは既存staff所属/医院scopeと現在viewを検査。PATCHはedit、対象staffの当該assignment現在所属を毎回確認。
View exact={staff_id,clinic_id,assignment_id,version,nomination_paused,capabilities_configured,reservation_type_ids}。ID/版は正規十進文字列1..int64最大。nomination_pausedはtrue/false/nullをそのまま返す。capabilities_configuredはconfigured_at非NULLのboolean、日時自体を版にしない。IDsは当該医院の保存済みpositive集合の数値昇順。未設定falseと設定済み空集合true/[]を区別し、未設定は予約不可。既存nullの指名状態はunknownとして表示しfalseへ初期化しない。未知値の適格性は既決failclosedと同じ現在readerへ渡す。
PATCH exact={operation_key,expected_version,changes}、changes exact optional keys={nomination_paused,reservation_type_ids}、最低1field。nomination_paused指定時bool必須、null書込不可。IDs指定時配列必須、null拒否、空配列は明示的な設定済み空集合、最大50は既存入力上限を維持、重複/未知/別医院ID拒否。未指定は保存値維持、configured_atの解除は本slice非対象。能力集合を書けばconfigured=trueとなる。型/人数/職種/所属から全許可を生成しない。
既存休止/削除typeが保存集合にある場合、未編集既存IDは保持する。新規追加するIDだけ現在選択可能な同医院区分を要求し、旧無効IDを黙って削除したり全編集を不当に止めたりしない。rename/group表示は492 readerを再利用する。

## 2. 同一aggregate CASの最小DDL案

既存staff_clinic_assignmentsへ eligibility_version bigint NOT NULL DEFAULT 1 CHECK>0、eligibility_digest bytea NOT NULL CHECK octet_length=32を追加する。別staff/能力業務台帳は作らない。旧行の初期digestは実保存値からのみ算出し、nomination null/未設定/positive IDsを変換しない。稼働データ適用は別承認。
DigestProjection exact={schema:"staff-reservation-eligibility/v1",staff_id,clinic_id,assignment_id,assignment_deleted,nomination_paused,capabilities_configured,reservation_type_ids}。assignment_deletedはdeleted_at非NULLのboolean。UTF8 JCSのSHA256。IDsは完全な保存集合の数値昇順、日時・updated_at・職種・権限を含めない。parent版はこの集合状態の変更ごと一回だけ増分、同値は増分ゼロ。configured=falseからtrueは集合空でも変更。
writerはstaff→assignment lock下で全子を変更後に最終digest/版を一度更新する。BEFORE UPDATE guardはdigest変化ならnew.version=old+1、非変化なら同版を要求し、手動跳躍/減少を拒否。assignmentとcapability子のdeferred constraint triggerはcommit時の実集合から再計算したdigest一致を検査する（子削除/insert/一括置換/soft delete/restoreも対象）。これにより旧writerが版更新を省略すればrollback。アプリとtriggerで二重増分しない。勤務版やpermission版と混用しない。原anchor保存のための業務テーブル専用uniqueや複合FKは追加しない。
新assignment作成は実値に対応するdigestと版1を初期化。configured未設定/nomination nullを許可の既定へ変換しない。DDL/model/概要・生成ERDと全必要writerを一つの実装subsetとして渡す。

## 3. 正確な保存条件・予約保護

停止true又は能力削除によって未来/現在診療中（予定終了超過含む）の既指名が不適格になる場合は拒否。予約ownerの既存判定時間/引継ぎ後区間を使い過去全体勤務を要求しない。影響予約の理由・日時/患者/区分・編集導線は現在予約viewでのみ開示。権限がなければ安全な拒否だけ、取得不調は保存停止。過去完了のみで永久拒否しない。
falseへの明示変更はこの属性の停止解除だけで、能力/勤務/休憩/休業/医院条件を免除しない。勤務/予約を生成せず、未知の過去設定を推測復元しない。A院操作はA院assignmentと能力だけ、他院・staff/account・権限group不変。
同値PATCHも現在認可と期待版を先に検査し、古版をno-op成功にしない。意味値が同じで新keyならreceiptのみ、版/変更auditゼロ。成功同keyreplayは原結果、後日の現版で再実行しない。

## 4. C06・元anchor・lookup

owner_kind=staff、intent=staff.reservation_eligibility.change。Content exact={schema:"staff-eligibility-command/v1",staff_id,assignment_id,command_sha256}。commandはclosed {schema,intent,staff_id,assignment_id,changes}、presence/nullを区別するJCS SHA256、256byteを超えるID集合をmapへ直入れしない。[共有Q](../../design/shared/operation-participation.md#verified-command-selected-20261021) deep-copy/再hashを使いApplyはreq.VerifiedCommandのみ。Preconditions={eligibility_version:expected_version}、現在認可/保護をclient stampへ委譲しない。
限定metadata表 staff_eligibility_operation_targets: clinic_id/binding_id/staff_id/assignment_id bigint NOT NULL、PK(clinic_id,binding_id)、FKは(binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICTだけとする。staff_id/assignment_idは業務FKを持たないimmutable anchor値として保存する。全ID>0。claim前preflight、元anchorの現在認可を行うlookup、業務TX内Applyの各段階で、実staff/assignmentの存在・同clinic・assignmentのstaff一致を検査する。短claim内では業務実体のFK検査や業務lockを実行せず、preflight後の変更はlookup/Apply時に検知して非開示または拒否し、incoming対象への置換で回復しない。UPDATE/DELETE禁止、clinic RLS、通常公開なし。
新keyは現在edit/対象関係を確認したcommandだけ短claimへ。ClaimMetadataWriter winner-onlyでbinding/anchor同TX、失敗両rollback、業務lockをclaimへ持ち込まない。loserは原anchorを再利用、incoming対象で補修しない。
lookupはtrusted clinic+actor+owner+intent+keyでbinding→原anchor→現在MasterStaff:view/元staff所属とassignment対応→URLstaff一致。元actor一致必須、別actor key探査なし。原anchorなし/解決不能は非開示、成功refsだけで失敗binding認可を推測しない。
成功refs exact={staff_id,assignment_id}、versions={staff_eligibility:resulting_version}、C06 flat ResultResponseと既存12status/HTTPを維持。succeeded/replayedは元receipt、details配列なし。非成功は既存標準shape。保存失敗validation、版不一致version-conflict、確認後依存変更dependency-changed、unknown/result-unavailableは新key救済しない。現在viewのGETで最新状態を再読し、元receiptと混ぜない。

### 非成功後の復旧・再送遷移

すべて元keyの現在認可付きlookupで結果を確認してから判断する。通信断/503から未実行を推測しない。入力案はscope/失権規則に従い保持し、lookup結果と現在Viewを混ぜない。

| 元操作の確認状態 | 次の操作 |
|---|---|
| validationの確定拒否 | 現在View/保護予約を再取得して原因を解消する。changesの意味値を修正するなら、旧操作が未成功で確定した後に内容を再確認し新operation_key。外部で保護予約だけが解消され意味値が同じなら、現在依存/CASを再確認した同keyで再試行可。無条件の自動再送なし |
| version-conflictの確定拒否 | 最新Viewと差分を読んで再確認し、同一changes/元targetを維持する場合だけexpected_versionを更新して同keyで再試行する。changes/targetが変われば旧操作確定後の新command/newkey。成功済みなら現在CASで再実行せず原receipt |
| dependency-changedの確定拒否 | assignment identity、所属/区分/保護予約の現在状態を再取得して再確認する。同じ元target・changesが有効な場合だけ現在前提で同key再試行可。元assignmentの付替えや内容変更は旧操作確定後の別command/newkeyであり、旧anchorを補修しない |
| unknown / result-unavailable、通信断、照会不能 | 通常PATCH再送・新key作成を止め、同key GET lookupだけを継続する。現在状態一致/時間経過/版変化から解決扱いしない。権威ある照会で成功なら原receipt、未成功が確定した場合だけ該当する上記遷移へ。未知が残る限り保留 |
| operation-in-progress | busy中は同key照会のみ。別key/自動Applyで競合を迂回せず、確定結果を得てから上記へ |

## 5. TX・全writer閉包・監査

短claim後の業務TXは既存reservation graph gate（必要な派生ctxをContextGateで保持）→C06 binding NOWAIT→成功receipt/unknownの再読取→savepoint→未成功Apply内でstaff FOR UPDATE NOWAIT→当該assignment FOR UPDATE NOWAIT→capabilityと区分の既定順SHARE。元anchor現在viewでreplay/unknownを開示分類し、未成功Applyはedit/所属・assignment identity・CAS・保護集合を再検査する。Participant.Gateはgraph gateだけを担い、staff/assignment等のcommand依存業務rowをbindingより前に先取りしない。予約rowの後取りlock/予約変更なし。全staffの複数対象操作はstaff ID昇順。
staff_service_permissions.go→UpdateReservationCapabilities、旧excluded facade、reservation staff create seed、所属restore/delete、直接junction writerは同aggregate更新に参加しdeferred guardを通る必要subset。旧excludedのactive universe逆算で新しい全許可を作らない。分類/職種gate退役は492/468の既決へ、既存職種master自体は残す。
同TXでassignment/pause/集合/版→audit_logs→成功receipt。audit action=staff.reservation_eligibility.change、resource=staff_clinic_assignment、resource_id=assignmentID、actor_type=staff、actor_id=trusted staff、clinic_id=scope。old_value/new_valueはDigestProjection＋versionの実before/after、metadata={operation_id,changed_fields}、changed_fieldsはnomination_paused/capabilitiesの変更した方だけ。既存audit sinkの失敗は全rollback、no-opは変更auditなし、replay重複なし。ip/UAは既存trusted contextのみ、新収集なし。

## 6. 旧API切替と1PR

既存GET capable-reservation-typesは互換readとして同集合から返し、UIの肯定形/group表示を維持する。既存PUTのbodyにserverが勝手に現在版/新keyを補って新契約成功としない。新GET/PATCHへ当該FE writerを移し、旧PUT・excluded mutationのクライアントが残るなら切替前にcaller一覧で把握する。活性化時は旧CASなしmutationを409 ErrorResponse code=contract_upgrade_required（新たなC06 statusではない）で停止して明示更新を促す。移行窓口の文書/互換エラーをapi.yamlへ同PRで追加し、認可を迂回する新routeを作らない。既存の関係ないstaff profile/permission APIは変更しない。
①本契約のDB/API/共通保護interfaceを確定、②マスタ設定PRは新DTO/保存/現在read/FE/履歴/receiptと既存予約保護の必要最小接続で閉じる。③以降・④新予約consumer全体をgateにしない。④の新予約入口へ医院別pause/positive判定を拡張する接続と受入は別成果。ただし既存予約を不適格にする変更が無保護で通る状態で②writeをactivateしない。
493単一統合owner、250既存capability/467保護/489Qとmetadata/446schemaの必要subsetを回収。休業/CSV/勤務日時CRUDは再実装しない。
fixtureは未設定/明示空/nullpaused、Aのみ停止/B保持、保護未解消/解消後、過去のみ、no-op古版、capability子だけ変更、旧writer bypass rollback、別target同key、失権lookup、audit/receipt失敗、group追加を自動許可しない、旧API停止とFE切替、全必要callerを定義。試験NOT_RUN。
