# EMR-493 休業期間consumer 採択技術契約

2026-10-21。親レビュー採択済みの統合技術契約。runtime未実装、試験NOT_RUN。実装・試験・DB適用・文書公開以外の外部公開を実施した記録ではない。

## 1. 範囲と既存成果
本契約は休業期間のread/create/change/history、C06結果照会、および勤務・指名適格性への接続を一つの限定consumerとして選定する。既存staff_leave_periods/staff_leave_changesを再利用し、休業業務表を二重作成しない。原対象認可を保持する最小anchor表のみ追加する。
勤務日時CRUDの既存api.yaml・EMR-514成果は維持する。starts_at/ends_at、JST勤務開始日date、off/paid_leave、休憩、unknownの既存契約を変更しない。既存Go callerの旧時刻field追従、休憩単独aggregate版、copy/CSVは後続sliceであり、本休業契約の設計完了を待たせない。
休業は在籍を保持する期間。退職・account停止・権限付与・所属生成に変換しない。期間内の新規勤務と指名を適格性から除外するが、既存勤務・予約を無断変更しない。復帰だけで勤務や指名を生成しない。

## 2. Routes・要求・read DTO
実router prefixは/api/v1（composition_runtime.goのrouter.Group("/api/v1")からprotectedへ渡す）。以下表記はapi.yaml相対pathで、/v1という実route prefixを意味しない。MasterStaff resourceを既存値のまま使用。
GET /masters/staffs/{staff_id}/leave-periods?from=YYYY-MM-DD&through=YYYY-MM-DD&after_id={id}&limit={n}
GET /masters/staffs/{staff_id}/leave-periods/{leave_id}
GET /masters/staffs/{staff_id}/leave-periods/{leave_id}/changes?before_version={v}&limit={n}
POST /masters/staffs/{staff_id}/leave-periods
PATCH /masters/staffs/{staff_id}/leave-periods/{leave_id}
GET /masters/staffs/{staff_id}/leave-operations/{operation_key}?intent={intent}
intentは必須でstaff.leave_period.create又はstaff.leave_period.changeの2値だけを許容する。欠落・未知値は入口400。認証済clinic/actor/owner_kind=staffとこのintent、operation_keyでFindBindingし、他intentへのfallbackやUUID単独検索をしない。同UUIDのcreate/changeは独立したbindingとして正しい元anchorへ解決する。queryはintent選択であり認可根拠ではない。
listは選択医院、当該staff所属と現view検査。範囲は検索対象であり勤務許可でない。一覧はid昇順・after_id排他、履歴はresulting_version降順・before_version排他。各ページでscope/staff/leave認可を再検査し、queryを別対象の認可根拠にしない。limitは省略時50、1..100を技術上限として選定する。範囲検索はfrom/through両方指定か両方省略、指定時from<=through。一覧応答は{items:LeaveView[],next_after_id:string|null}、履歴応答は{items:LeaveChangeView[],next_before_version:string|null}。次位置がnullになるまで全件取得完了とは表示しない。ページを跨ぐ更新のsnapshot保証は設けず、履歴はimmutable版順で取得し、現期間headは別GETで取得する。
IDは1..9223372036854775807の正規十進文字列。dateは厳密Gregorian YYYY-MM-DD。追加休業DTO版は十進文字列とする（既存ShiftEntry integer versionを勝手変更しない）。
LeaveView={id,clinic_id,staff_id,start_date,end_date,end_boundary,version,provenance,created_at,deleted_at}。nullableを必ず返しDBmodel直列化は使わない。履歴view={id,leave_period_id,resulting_version,before_values,after_values,operation_id,recorded_by,recorded_at}、actorはサーバ確定ID。履歴本文にclient由来actorを受けない。
Create={operation_key,start_date,end_date,end_boundary,provenance}全field必須、後3項nullable。clinic/staff/actorは本文で付替え不可。provenanceは任意補足出典で認可・人物所属・証拠確定に使用しない。新必須理由なし。
Change={operation_key,expected_version,changes:{start_date?,end_date?,end_boundary?,provenance?}}。changes最低1field、未指定維持、明示nullと欠落を別content。変更後全体で制約検査。end_dateをnullへする場合end_boundary=nullを一緒に明示。staff/clinic/id/created_at/deleted_at/version直書不可。
Delete routeは第一sliceに設けない。deleted_at存在は取消・過去履歴無効化の業務許可ではない。期間訂正/終了の既定範囲で閉じ、休業誤登録の撤回/削除可否は別設計として残す。


## 3. 期間・unknown・重複
既存DB: end>=start、end NULL iff boundary NULL、boundary inclusive/exclusive/unknown。開始日は含む。
既知inclusiveはJST [start 00:00,end翌日00:00)、exclusiveは[start 00:00,end 00:00)。日付計算を時刻文字列比較や24h加算で代用しない。exclusive end=startは既存DDLが許す空区間として保存・表示を維持し、新しい入力拒否を加えない。
end_date NULLは終了未設定のまま保持し復帰日を推測しない。start以降の候補に休業終了済みと判定しない。boundary unknownはその終端日の包含未確認であり、確定値へ自動置換しない。曖昧性にかかる勤務/予約は未解決としてfailclosed。既存shiftwireのsource_unresolvedは維持するが、休業mutationはC06 validation/dependency-changedに分類し独自statusを加えない。既知の開始前/終了翌日以後まで無期限拒否に拡張しない。この三値評価は選定する技術契約であり実装済みではない。
勤務/予約は[開始,終了)と休業日の有効区間が交差するかで判定、日跨ぎは開始日だけで判定しない。候補と保存直前に同関数を使用。終日off/paid_leaveは勤務可能化しない。
DDLに休業重複禁止なし。重複は履歴を保持して適格性上union評価し、黙ったmerge/期間短縮/重複排除をしない。新たな排他制約や「重複は業務エラー」を発明しない。同key再送だけC06で一回化。重複のうち一件の終了が他の休業を解除しない。
終了/復帰にstaff/account/所属/capabilityを変更せず、シフトや指名を生成しない。


## 4. 認可・公開結果
POST/PATCHは既存MasterStaff:edit middlewareを維持する。新keyへのedit preflightとApply現在edit/所属再検査も必須。view-only actorはmutation retryでなくGET leave-operationsから原結果を読む。GETにはMasterStaff:view。受付後の失権をmutation middlewareで防げてもApply再検証を省略しない。
独自mutation receiptや独自statusは採用せずoperation.NewResultResponseのflat shapeのみ:
status,operation_key,result_refs?,committed_versions?,recorded_at?。
既存12値succeeded/replayed/not-succeeded/operation-key-conflict/operation-in-progress/result-unavailable/scope-not-found/not-authorized/validation/version-conflict/dependency-changed/edit-right-lostをそのまま保持。EffectSummaryは内部だけ。成功後のreaderで詳細期間/履歴を読む。既存shift専用wireへこの新休業wireを上書きしない。
format不正は既存HTTP入力error、対象・scopeの非開示は既存404/403慣例を維持。C06 owner拒否はvalidation（入力/既存勤務・保護予約不適合）またはdependency-changed（確認後変更）等既存分類のみ。理由一覧が必要なときは現在view認可付き期間/影響readerを再取得する設計であり、現在のResultResponseに任意detailsを追加した実装とはしない。影響予約は予約閲覧権限がなければ患者等を非開示とし、拒否自体だけ返す。
mutation成功refs={staff_id,leave_period_id,change_id}、versions={leave_period:正規十進版}。no-opではchange_id key省略、versionsは現版を記録。内部EffectSummary={changed:"true"|"false"}のみ。新key同値変更はCAS確認後no-op成功、history追加0。replayは原receiptそのまま。

## 原対象anchorと限定storage差分
既存operation_bindingsはfingerprintとscopeのみ。Content原文/対象IDを保持せず、失敗bindingにstaff_leave_changesもない。成功refsだけでは失敗/unknown/key-conflictのoriginal-target認可を閉じられない。scope IntentへstaffIDを埋める案はキーnamespaceを変え、別target同key競合を検出できなくするので不採用。
選定追加表 staff_leave_operation_targets:
clinic_id bigint NOT NULL、binding_id bigint NOT NULL、staff_id bigint NOT NULL、leave_period_id bigint NULL。
PK(clinic_id,binding_id)、即時FK(binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT、staff_id→staffs(id) RESTRICT、nullable複合FK(leave_period_id,clinic_id)→staff_leave_periods(id,clinic_id) RESTRICT、全非NULL IDは正数CHECK。
createのleave_period_idはNULL（未採番targetはstaffを認可rootにする）、changeは既存leaveID必須。binding owner_kind="staff"、intent="staff.leave_period.create"|"staff.leave_period.change"と型の対応をclaim時/resolve時検証。update/delete禁止と既存clinic RLS。対象の現在在籍を原anchorが証明するとはしない。
既採択EMR502 ClaimMetadataWriter型の限定拡張を再利用する。同短claim TXでbinding新規winnerだけcanonicalContent staff_id/leave_period_idからanchor INSERT。callback失敗は両方rollback。既存binding loserはanchorを再作成・補修しない。欠落anchorはfailclosed、原targetをURL/新payload/成功historyから推測しない。create成功後もanchorをleaveIDで書き換えず、原staffrootのまま成功refsのleaveを照合する。
Resolve: trusted clinic+actor+owner+intent+keyのbinding→同binding anchor→原staff/leaveのclinicと関係を検査→現在MasterStaff:view→URLstaff/leave一致。認可前のmismatch詳細を返さない。削除・所属終了後も現在閲覧可能な原履歴rootが解決できれば原結果取得、解決不能はnot-authorized。結果照会を編集grantにすり替えない。
create前のstaff所属/actor/edit preflightを通過した内容のみclaim。Apply前に所属喪失してもclaimは原targetを保持し拒否結果を照会可能。この追加は休業業務表の重複でなく操作認可anchorのみ。

## lock順を選択
実コード:
operation/runner.go109–111: ClaimBindingはowner業務TXより前に独立claim。
同297–300: owner Gate→binding NOWAIT→receipt/unknown→Apply。
staff/shift_entry_repository.go237–263: staff UPDATE→対象assignment UPDATE。
reservation/reservation_staff_repository.go75–103: 予約のambient TX読取はstaff SHARE→assignment SHARE。mutation109–138はstaff UPDATE→assignment UPDATE。capability junctionも同順。
採択順:
(0) 認証/選択clinic/現在edit/actor実体/対象staff所属preflight。
(1) 短claim TXでbinding＋原anchorをcommit。業務lockを保持したままclaimしない。
(2) 新業務TXのGateで原anchorのstaff行FOR UPDATE NOWAIT→対象staff_clinic_assignment行FOR UPDATE NOWAIT。複数対象の将来bulkではstaff ID昇順だが本sliceは一staff。
(3) binding FOR UPDATE NOWAIT。
(4) receipt/unknownを読み原anchor現在viewで開示。成功replayに新CASを適用しない。
(5) Applyで現在edit/grant fence・staff/assignment有効性再検証→対象leave FOR UPDATE NOWAIT（change）→expected_version。
(6) 当該staffの休業集合と勤務・保護予約を再読取して候補状態に対する適格性を確認→period/history/audit/receipt→commit。
staff先行exclusiveが予約staff SHARE・勤務staff UPDATEと同じ競合点になる。既存lower helperをGateでblocking取得するのでなく同順NOWAITで取得し、55P03はoperation.Busy→operation-in-progress。独自gate table/advisory keyなし。
予約行をleave側が後からFOR UPDATE取得すると、予約writerのappointment→staff順との循環を作り得るので、このsliceのleave側は予約を更新/lockしない。staff gate下で現在保護集合を読む。現在予約の状態変更writerについて、指名/時刻/再受付等「新たに保護集合へ入る変更」は必ずstaff SHAREを取得してから変更する必要がある。これをconsumer接続の実装義務として列挙し、無関係な全467完了を要求しない。staff SHAREを通らない新規保護集合writerを放置して有効化しない。
既存read取込だけでこの全writer閉包が実装済みとはしないが、順序自体はこの案で固定する。過去予約・他院変更の無差別lockや再実装を行わない。

## 履歴/監査exact shape
before_values/after_valuesはschemaタグ付きobject:
{schema:"staff-leave-state/v1",staff_id:string,start_date:date,end_date:date|null,end_boundary:enum|null,provenance:string|null}
create before_values={}、afterは保存値。changeはlock取得後の実旧値/実保存値。client提供beforeを監査値として信頼しない。
staff_leave_changesのrecorded_byは認証accountに紐づく実staff、recorded_atはDB、operation_idはbinding.id、resulting_versionはRETURNING版。期間・provenance変更の通常理由は必須追加なし。
既存audit sinkへ同TXでaction=staff.leave_period.create または staff.leave_period.change（本契約で選定する技術action名）、resource=staff_leave_period、resource_id=保存leaveID、clinic_id=scope、actor_type=staff、actor_id=認証staff、metadata={operation_id:string,change_id:string,previous_version:string|null,resulting_version:string}。before/afterとprovenanceは業務historyに保持し、auditへ重複私的内容を載せない。no-opはhistory/auditの架空変更イベントを作らず成功receiptだけ、内部changed=false。001:2596–2615の実列resource/actor_id/metadataを使用し、old_value/new_valueはNULL（詳細は同TXのhistory参照）、created_atはDB既定。ip_address/user_agentは既存trusted audit contextから取得可能な範囲だけ、client本文を採用しない。新audit table不要。
初期版は1、change版はentity_version_guard、一回のみ。履歴immutable/deferred成功receipt/RLSを保持。409等の失敗で業務historyを残さず共通attempt outcomeに委ねる。

## immutable Contentとhash commitment
C06 Content各値上限を尊重し、原文provenanceを直接string mapへ押し込まない。Content={staff_id,leave_period_id:create時"none"|id,command_digest}。
command_digest=SHA256(UTF8(JCS({schema:"staff-leave-command/v1",intent,staff_id,leave_period_id:null|id,values:明示field map})))のlowerhex。期間/provenance文字列とnull/presenceを保持。typed immutable commandが同projectionを再計算しContentとの一致をApplyで確認後のみ書込。actor/clinic/key/CAS/時刻はprojectionへ入れずscope/preconditionsへ。
newclaim metadata callbackはimmutable command再計算とContent IDsを照合してanchorを作成する。任意captured payloadへfingerprintを書き替えない。


## 後続sliceと非対象
- 休業DELETE/取消は本契約に含めない。調査したD14/493全コメントに休業自体の削除許可は確認できないため、期間短縮・境界変更・履歴保持を先行する。削除が必要になった場合のみその意味を別途確認する。
- 勤務日時BE/514実API結合、446/466の休憩単独aggregate版は別subset。休業の初回read/create/change/history設計と混ぜない。
- copyは同医院・選択staff・元期間/先開始日・相対日付・元snapshot・明示置換・確認範囲atomicの既決条件を維持する。既存shift_bulk_operationsを再利用する後続技術契約とし、通常単日APIのループで部分commitしない。
- CSVの列/識別列/encoding/日時文法/複数休憩表現/サイズ上限/原ファイル保管は別設計。氏名から人物推測、空欄削除、staff自動登録、正常行自動部分保存をしない。共通normalized rows validator/preview基盤だけ先行可能。CSV完成は主張しない。
- 新権限、外部通知、予約自動取消/移動/指名解除、勤務自動復帰、実環境データ変更は追加しない。

## 受入ケースと完了境界
すべてNOT_RUN。inclusive/exclusive/同日空区間/unknown/null、日跨ぎ/境界接触、重複union、同値no-opとhistoryunique、CASrace、休業作成と予約同時変更、claim+anchor原子性/競合winner、別target同key、失権/actor未解決/他院非開示、commit応答喪失、deferred history/audit/receipt失敗rollback、復帰による勤務作成0、他医院不変をfixture化する。
実装時に必要な指名・勤務writerを列挙し、staff gate前提と現在休業のsave-time再検査へ対応づける。院内/LINE予約作成・変更・再受付・指名変更、通常/トリミングの対象指名owner、勤務create/change、SaveByStaffDate、所属/能力変更の該当経路を含む。未接続writerを残して機能活性化・全適格性完成とはしない。必要subsetの連携を閉じれば、無関係な467全機能や後続CSV完成を一律待たない。

## 根拠と確認範囲
固定main:19855942f762f31f0403b5de0ddbc61b761fd94e。backend/docs/api.yaml blob9459cb23c968975417f4a44ea6f8d31ae701176fと取得済コピーの同一性を確認。
GitHub基点:https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/
- backend/migrations/001_init.sql:2596–2615 audit、8959–8974休業、10222–10229版、10596–10607履歴、10709成功receipt制約。
- backend/internal/model/storage_reservations.go:126–168既存model。
- backend/internal/staff/handler.go:210–237既存MasterStaff権限。
- backend/internal/staff/shift_entry_repository.go:237–263 staff/assignment UPDATE順。
- backend/internal/reservation/reservation_staff_repository.go:75–138 SHARE/UPDATE順。
- backend/internal/operation/{contract,dto,runner,store}.go:C06原対象resolver、flat結果、独立claim→Gate→binding順。
- docs/spec/domains/staff-workforce/README.md:staff-source-adoption、staff-leave-booking-handoff。
- docs/design/domains/staff-workforce/technical-spec.md:shift-bulk-operation-contract。
- docs/spec/screens/24-shift-calendar.md:shift-copy-csv-open-details。
- Plane EMR-493/514本文。勤務日時APIは既選定だが休業routeは未収録、現handlerにも未接続。
本契約により新しい業務質問を追加しない。ローカル資料照合のみで、実行・runtime安全確認・公開済みを意味しない。

## 変更履歴
R1からR2で、mutation入口は既存editへ統一、view-only取得はGET、独自wireを廃してC06 flat結果へ統一、原対象anchorの限定追加を明示、exclusive同日空区間を維持。R2の追補形式をやめ、本書を単一の採択技術契約へ統合した。
