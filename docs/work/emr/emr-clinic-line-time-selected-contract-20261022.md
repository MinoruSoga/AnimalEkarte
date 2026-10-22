# 共通医院時間・LINE受付・担当者別開始時刻：限定選定契約（A02 / A03）

2026-10-22技術選定済み。固定調査基準は `9eef73feac19c25571724372b0b19da160ba5cfb`。本稿は既決業務の目標設計であり、API正本収録・DDL/model/ERD・製品実装・実医院値確認・配備・受領・試験の完了を示さない。今回docsのみ、生成/build/test/Docker/DB操作はNOT_RUN。A01/A06は[別の選定契約](emr-master-update-selected-contract-20261022.md)を維持する。

要件判断窓口はMinoru Soga。共通医院基準、LINE内側制限、担当者別登録時刻を分け、レジ日界や既存予約を自動変更しない。既存設定をコピーして確認工程を省く方式、全staffへの共通slot複製、汎用schedule基盤・第二fence・独自監査台帳は採用しない。計画上の測定は保存scope/操作/権限/現在版/旧writer境界の対応が追えることとし、新しい業務KPIや実績を捏造しない。現5セッションのscope/担当を広げず、後続単一claimで受領する。

## 0. 採用事実と目標の区別

固定mainでは、`line_reservation_settings`の既存全量PUTと、医院＋予約区分単位の`reservation_type_available_slots`が存在する。後者にstaff/assignment scopeはない。`clinic_holidays`はshifts/closingの二入口で保存され、レジと予約双方のconsumerが参照する。これらの存在は共通医院時間・担当者別設定が完成した証拠ではない。492の既選定契約は従来LINE開始枠を拡張しないため、本稿が限定追加を選ぶ。

業務正本は[D17共通基準](../../spec/domains/clinic-settings/README.md#clinic-hours-common-basis-20261020)、[LINE分類別開始時刻](../../spec/line/reservation-spec.md#requirement-fr-line-reservations-business-start-times)、警告/現在認可は[予約技術仕様§6.1](../../design/domains/reservations/technical-spec.md)、C06は[共有操作参加/Q](../../design/shared/operation-participation.md#verified-command-selected-20261021)と[492契約](emr-490-492-selected-wire-contract-20261021.md)を再利用する。実装に仕様を縮小しない。

## 1. 共通wire・現在認可・時間境界

以下のpathは実prefix `/api/v1` を除く。新subresourceのID/版は正規十進文字列 `1..9223372036854775807`、operation_keyは既存非zero小文字UUID。既存APIの全ID型を変更しない。全DTOをclosedとし、未知/重複key、不正UTF-8、複数JSON値を拒否する。全read/preview/結果照会はno-store、previewはread-onlyでclaim0/write0。医院/actorは認証contextとURL/元対象の一致を検査し、bodyから採用しない。

共通医院時間はJST暦日・厳密HH:MM。startは00:00..23:59、endは00:01..24:00、同日start<end、通常区間は[開始,終了)。通常予約の予定区間終了=閉店は可。日跨ぎは隣接曜日/日の2区間を明示入力し、片側から自動生成しない。24:00は専用endpoint変換で翌JST暦日00:00へ解釈し、旧HHMM parserへ2400を無変換で渡さず、レジ営業日に付替えない。日別特別営業時間・業務有効期間・RRULEは追加しない。

## 2. A03 共通医院時間aggregate

### 2.1 保存・Settings

小さな `clinic_operating_hours` aggregateをclinic_id PKで新設する。正bigint version、schema_version=`clinic-operating-hours/v1`、weekly_hours、closed_dates、national_holiday_closedを保存する。clinics基本情報、clinic_settingsレジ日界、LINE内側設定とは別。医院scope/RLS・現在医院predicateと既存clinic参照を維持し、DDL/model/ERDへの実反映は①受領残とする。

Settings exact=`{weekly_hours,closed_dates,national_holiday_closed}`。

- weekly_hours=`[{weekday,intervals:[{start,end}]}]`。曜日0..6各1件、曜日昇順。各日の区間は開始昇順・非重複、空配列は明示休診。
- closed_dates=`[{date,reason}]`。厳密Gregorian YYYY-MM-DD、一意昇順。reasonは空可、既存休診理由validationに整合する。
- national_holiday_closedは明示boolean。初回は全field明示、実時刻/曜日/祝日方針のdefaultなし。

行なしを全日営業/全日休診へ補完しない。集合は全置換、未知key/null拒否。正常な空と未設定、legacy未照合、JSON不正、取得失敗を別状態/エラーで扱う。

### 2.2 完全設定のAPI

| method/path | request・意味 | 現在権限 |
|---|---|---|
| GET /clinics/{clinic_id}/operating-hours | exact=`{schema_version,clinic_id,state,version,settings}`。state=configured/unconfigured。configuredは版/Settings必須、unconfiguredは版/Settings null | 現在医院所属＋hospital_settings:view |
| POST /clinics/{clinic_id}/operating-hours | initialize。exact=`{operation_key,settings,warning_receipt?}`、行不在要求 | hospital_settings:edit |
| PATCH /clinics/{clinic_id}/operating-hours | update。exact=`{operation_key,expected_version,changes,warning_receipt?}`。changesはSettingsの1field以上 | hospital_settings:edit |
| POST /clinics/{clinic_id}/operating-hours/preview | mode=initialize/update、対応settings/changesとexpected_version（initializeはnull、updateは正版）。同じ設定/警告計算のread-only照会 | hospital_settings:view、元医院所属 |

障害503をunconfiguredへ変換しない。初回commonはclinic行lockとPK uniqueで別keyの初期化競合を閉じる。更新/no-opとも現在認可とCASを検査し、真の変更時だけ版を一回増分する。

owner_kind=`clinic`。intentsは次の6個だけ。lookupの短名を一意にmapする。

| lookup短名 | intent | 原familyの現在view |
|---|---|---|
| initialize | clinic.operating-hours.initialize | hospital_settings:view |
| update | clinic.operating-hours.update | hospital_settings:view |
| holiday-set-from-shifts | clinic.operating-hours.holiday-set-from-shifts | shifts:view |
| holiday-unset-from-shifts | clinic.operating-hours.holiday-unset-from-shifts | shifts:view |
| holiday-set-from-closing | clinic.operating-hours.holiday-set-from-closing | closing-settings:view＋既存shared handlerの追加条件 |
| holiday-unset-from-closing | clinic.operating-hours.holiday-unset-from-closing | closing-settings:view＋既存shared handlerの追加条件 |

GET `/clinics/{clinic_id}/operating-hours/operations/{operation_key}?intent=<短名>` は元familyの現在viewと元clinic/actor/targetを確認するC06照会。別intentへ変えて別権限で元結果を読まない。success refs=`{clinic_id}`、versions=`{clinic_operating_hours}`。新familyの共有registry/API/lookup allowlistは446/489の必要subsetとして受領し、収録済みとしない。

### 2.3 個別休診の権限限定adapter

| method/path | request/response | 既存grant |
|---|---|---|
| POST /clinic-holidays/common | set exact=`{operation_key,expected_version,date,reason,warning_receipt?}` | shifts:create |
| DELETE /clinic-holidays/common/{date} | unset exact=`{operation_key,expected_version}` | shifts:delete |
| GET /clinic-holidays/common?year_month=YYYY-MM | 月内holidayだけ＋common版、全営業時間は返さない | shifts:view |
| POST /clinic-holidays/common/preview | set/unsetのaction・対象日・期待版、set時reasonのread-only確認 | shifts:view |
| POST /closing-settings/holidays/common | 上記setと同shape | closing-settings:create |
| DELETE /closing-settings/holidays/common/{date} | 上記unsetと同shape | closing-settings:delete |
| GET /closing-settings/holidays/common?year_month=YYYY-MM | 月内holidayだけ＋common版 | closing-settings:view＋現shared handlerのshifts:view条件 |
| POST /closing-settings/holidays/common/preview | 上記と同じ限定preview | closing-settings:view＋現shared handlerのshifts:view条件 |

日付限定actionをserviceへも渡し、各権限の一律ORで全営業時間/曜日/祝日方針を書けるようにしない。common未設定なら初期化へ案内し、他fieldを生成しない。同日同理由set/既にない日unsetは現在CAS検査後no-op、同日異理由setは拒否し無言upsertしない。closing側の既存追加shifts:view条件は権限棚卸しなしに除去しない。新しいgrantは作らない。

### 2.4 旧休日とレジを巻き込まない切替

旧clinic_holidaysは出所別に棚卸しし、名前だけで共通正本へ昇格させない。共通初期値は権限ある医院側の明示確認からだけ保存する。旧履歴/IDは削除せず、dual-writeをしない。

common利用開始の医院では旧shifts `/clinic-holidays` mutationを409 ErrorResponse code=`contract_upgrade_required`で停止し、新adapterへ移す。予約/common readerが旧clinic_holidaysを参照しなくなったことを切替条件にする。

旧 `/closing-settings/holidays` は既存grant・既存保存意味の**レジ専用legacy read/writeとして維持**する。新commonへ書かず、新common日付adapterとは別操作。common/予約readerは旧表を参照しない。別途合意までレジ休日編集停止、レジ日界やclosing_special_periodsの共通転用、既存値の自動上書きは行わない。

## 3. A03 LINE acceptanceの限定契約

### 3.1 保存版・保護field

既存 `line_reservation_settings` を別保存のまま、acceptance_versionとreviewed_acceptance_versionだけを限定追加する。旧行は版1のmetadataだけを置き、reviewed=nullで未照合を維持する。旧実値の正しさを認定しない。confirm/updateで意味値または確認状態が変わると一回だけ版を進め、reviewed=新版。真のno-opは同版、通知/表示変更で版を進めず、triggerとアプリで二重増分しない。

保護field Pは以下の12個に閉じ、各既存型/validationを維持する。

`status, closed_weekdays, closed_dates, national_holiday_closed, business_hours, business_hours_by_weekday, break_hours, booking_window_max_days, booking_window_min_days, time_slot_mode, time_slot_interval_minutes, show_no_staff_option`

min<=max、壊れたJSONにdefaultを入れない。business_hours_by_weekdayのnullによる曜日override全解除以外はnull不可。共通SettingsのHH:MMを既存LINEのHHMM wireへ機械転用せず、境界adapterで既存型を検証してJST区間へ解釈する。trimmingに通常診療用の生成開始時刻を加算しない。

非P表示案内Nは `header_text,reservation_notice,cancel_notice,privacy_policy,calendar_months,phone_number,notification_email,request_example`。additional_fieldsは登録フォーム専有。line_channel_id/liff_id/access token/channel secret等の連携/秘密は新DTO/receipt/logへ入れない。旧no_staff_mode自動指名と撤回済みdaily_limit/monthly_limitをP/新consumerへ復活させず、旧値はlegacy保持する。

### 3.2 APIと状態

以下のbaseは `/clinics/{clinic_id}/line-reservation-settings/acceptance`。現在hospital_settings:view/editと現在医院所属を使う。

| method/path | closed入力/出力 |
|---|---|
| GET base | exact=`{schema_version:"line-acceptance/v1",clinic_id,state,acceptance_version,settings,common_version}`。state=unconfigured/legacy-unreviewed/configured。未設定はacceptance版/設定null、旧未照合は実保存版と検証可能なPを返し確認済みとしない。JSON不正/読取不能は失敗で空成功にしない。common未設定はcommon_version null |
| POST base | initialize。line行なし、exact=`{operation_key,expected_common_version,settings}`、settingsはP全field明示 |
| POST base/confirm | 旧未照合から確認。exact=`{operation_key,expected_acceptance_version,expected_common_version,settings}`、P全field明示 |
| PATCH base | configuredのみ。exact=`{operation_key,expected_acceptance_version,expected_common_version,changes}`、Pの1field以上 |
| POST base/preview | initialize/confirm/updateのmode、対応期待版とsettings/changes。read-onlyで共通との包含/実効集合を確認 |
| GET base/operations/{operation_key}?intent=initialize\|confirm\|update | 元clinic/actorと現在viewを検査し、下記の固定intentへmapするC06結果 |

writeはedit、GET/preview/lookupはview。owner_kind=`reservation`、intents=`line-acceptance.initialize/confirm/update`。success refs=`{clinic_id}`、versions=`{line_acceptance,clinic_operating_hours}`。common版は検査した版であり共通更新を意味しない。common未設定のまま受付確認/新設定を有効化せず、初期化へ案内する。保存値を成功receiptへ混ぜずGETで再取得する。

### 3.3 内側制限・旧writer閉包

確認済み共通C(d)と確認済み既存LINE内側L(d)のintersectionを実効値として読む。実効値を保存へ書き戻さない。共通縮小時もLINE保存値は保持し、管理previewでmask差を示す。共通拡大時にLINE保存値を拡大しない。legacy未照合値を確認済みへ昇格しない。新LINE変更は共通内を検証し、切詰め保存せず拒否する。

common/LINE未設定・legacy未照合・正常空・JSON不正・取得失敗を区別し、LINE判定不能を受付可へfallbackしない。院内では既存警告契約へ接続し、医院時間未設定という理由を分けるが、院内全面拒否は新設しない。

活性化医院の旧全量PUTはP同値でも409 `contract_upgrade_required`/zero-write。切替前に既存base `/clinics/{clinic_id}/line-reservation-settings` のPATCH `/presentation`（exact=`{changes:N}`、Nの変更fieldだけ）とPATCH `/registration-fields`（additional_fieldsだけ）を用意する。既存grant/validationを維持し、P/秘密を再保存せず、非Pへ新C06/版必須化を広げない。未切替医院と旧readを一律停止しない。Pの旧直接UPDATEも版を迂回できないよう、必要writer閉包/限定guardを①②で受領する。旧repositoryの独立TXを外側へネストしただけで同TX化完了としない。

## 4. A02 担当者別trimming開始policy

### 4.1 保存・日付置換

`line_staff_start_policies` を新設し、PK/uniqueは `(clinic_id,reservation_type_id,assignment_id)`。staff_idはそのassignmentのstaffと照合する。正version、schema_version=`trimming-registered-starts/v1`、weekly_times/date_overrides/deleted_atを持つ。勤務/能力/所要時間/定員は複製しない。医院scope/RLSと設定writerの現在所属・type/staff/assignmentの関係検査を維持し、未設定/明示空/取得失敗を分け、他院/他担当へfallbackしない。休業/指名停止を理由に将来設定編集まで禁止する新規則を加えない。

Policy exact=`{weekly_times,date_overrides}`。weekly_times=`[{weekday,times}]` は0..6各1件、曜日昇順、timesは一意昇順のHH:MM 00:00..23:59。date_overrides=`[{date,times}]` は厳密日付一意昇順。日別overrideがあれば週次を置換し、times=[]は当日0、例外削除で週次へ戻る。旧weekly+specific unionと明示的に異なる新契約。固定15分/5分刻み/未来有効期間を加えない。旧共通slotを全staffへ複製しない。

### 4.2 API・削除・再設定

base=`/masters/reservation-types/{type_id}/staffs/{staff_id}/line-start-policy`。

| method/path | 入力・意味 | 既存grant |
|---|---|---|
| GET base | exact=`{schema_version,clinic_id,reservation_type_id,staff_id,assignment_id,state,version,policy,deleted_at}` | master-reservation-type:view |
| PUT base | exact=`{operation_key,mode,expected_version,expected_assignment_id,policy}`、mode=create/replace/reconfigure | master-reservation-type:edit |
| DELETE base | exact=`{operation_key,expected_version,expected_assignment_id}` | master-reservation-type:delete |
| POST base/preview | mode/期待版/assignment/policy（delete以外）/明示dates、read-only | master-reservation-type:view |
| GET /line-start-policy-operations/{operation_key}?intent=create\|replace\|reconfigure\|delete | 元scope/actor/assignmentの現在view、厳密intent mapでC06結果照会 | master-reservation-type:view |

GET state=unconfiguredはversion/policy/deleted_at null、configuredは版/Policy必須・deleted_at null、deletedは現版/削除時刻/保存Policyを管理用に返す。現在の元所属を認可不能なら非開示。予約可能booleanを返さず、staff selectorは既存staff readerの現在認可を維持し、読めない氏名を新projectionで漏らさない。

createはexpected_version null＋行なし。replaceは現版＋非削除。reconfigureは現版＋tombstone＋全Policy明示、同PK/同staff/同assignmentのままdeleted_at解除・版+1。入力なしrestoreや別assignment付替え、createによるtombstone迂回なし。DELETEはtombstone＋版増分、現版で既削除ならno-op、古版はconflict。物理削除しない。

owner_kind=`reservation`、intents=`line-start-policy.create/replace/reconfigure/delete`。typed target exact=`{kind:"line_staff_start_policy",clinic_id,reservation_type_id,staff_id,assignment_id}`。lookupは短名からこれらへ固定mapし、別familyへ切替えない。

previewは設定上の登録集合、common/LINEでmaskした集合と各版を返す。勤務/休業/能力/競合は未評価と明示し、保存保証/予約可能tokenを返さない。

### 4.3 未成功lookupの原assignment

既存492 `reservation_master_operation_anchors` の明示後続差分として、target_kind=`line_start_policy`、target_id=type ID、staff_id/assignment_idの2列を追加する。旧type/groupは2列NULL、新kindだけ正ID必須。既存binding FK/RLS/immutable、winner-only ClaimMetadataWriterによるbinding＋anchorの短い同TXを維持する。業務staff/assignment/type FKを短claimへ追加せず、原関係はpreflight/現在認可/Applyで検査する。新業務台帳/dedup registryなし。492/446/489の型/DDL/lookup allowlist受領を必要subsetとする。

incoming targetで原anchorを補修しない。元assignmentの現在認可不能・anchor欠落/解決不能は非開示、認可成立後の結果読取不能は既存result-unavailableで停止。現staffの別assignmentや成功履歴だけで未成功の元対象を推定しない。

## 5. C06・同TX・警告確認

### 5.1 commandとlock

既存C06/Q/ContextGateを再利用する。Contentはschema＋command commitment＋typed targetの短いstring map、期待版とwarning_receiptはPreconditions、assignmentはimmutable targetとする。Qでdeep-copy/claim前とApply前に検証し、Applyはreq.VerifiedCommandだけから入力を再構成する。別DTO/pointerをcaptureしてwriteしない。Qはrequest-memoryで耐久保管を兼ねず、既存元target解決/ClaimMetadataWriterの必要subsetを受領する。

短いclaim後の業務TXはgraph gate→binding NOWAIT→Apply行lock（common→LINE等の既存順）→CAS/現在認可/警告根拠再検査→業務write＋既存audit＋成功receiptの同TX。初期commonはclinic行/PK unique、policyは元assignmentを含む現在target検査で競合を閉じる。新global gate/独自fence/全owner Done待ちを作らず、旧LINE repositoryの独立TXのまま外側成功receiptだけ先commitしない。

既存492のflat ResultResponse/12status・HTTPを維持する。succeeded/replayed/not-succeeded=200、operation-in-progress=202、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、result-unavailable=503、scope-not-found=404、not-authorized=403、validation=400。独自details/reasonを共通結果へ足さない。legacy停止/strict decodeは既存ErrorResponse。success refs/versionsは各節の限定集合、設定値はGET再取得。auditはtrusted actor/clinic/対象と実before/afterを既存sinkへ同TX、失敗全rollback、no-op/replayで版/業務変更監査を増やさない。

同key replayは元結果、unknownは元key照会、busyも通常再送/新key救済にしない。確定拒否後の同command・Preconditions再確認と業務意味値変更は既存C06規則で分ける。失敗を別keyで迂回せず、保存済成功を現在設定で再実行しない。

### 5.2 既存warning proof

予約technical-spec§6.1のcode/severity/fields/dependency_revision/opaque warning_receiptを再利用し、C06成功receiptと分ける。common previewは現在設定・提案・実効LINE差分・影響既存予約をread-only計算する。休診追加は既存予約一覧を警告し、確認すれば保存可。全予約解消を条件にせず、予約/勤務/記録を連鎖変更しない。

予約一覧は現在reservations:viewの開示範囲だけ。権限不足は内容を漏らさず確認不能、取得不調も確認済みにしない。取消はwrite APIを呼ばずclaim0/write0。LINE側へ警告overrideは設けない。

proofを医院/actor/family/target/command commitment/common版/必要LINE版/影響予約ID・版・日時・状態の正規集合へ結ぶ。既存warning証明adapterを再利用し、新registryを作らない。保存時は同じ関数で再評価し、対象集合変更は409で再preview/再確認、CAS違いも409、現在失権は拒否。warning受領が現在server検査を免除することはない。

## 6. ホテルの閉店点時刻（本人承認）

LINEホテル予約の入室・退室予定は閉店時刻ちょうどを許可する。9:00〜18:00なら18:00可、18:01不可。休日・明示休憩・LINE受付条件・ホテル定員等の独立条件は免除しない。

営業元区間の終端と、休憩/休日/LINE差集合から生じる端点を区別する。休憩開始点を「閉店なので可」と再許可しない。24:00は実日時の翌JST暦日00:00で判定し、翌暦日の休診例外をこの承認から作らない。宿泊途中の休診で全滞在を拒否する条件を追加しない。通常予約の予定区間終了=閉店の包含と、ホテル点時刻の境界は別fixtureにする。

## 7. ①②③④の引渡しと配備

| 段階/既存owner | 必要受領subset | 完了と混同しない範囲 |
|---|---|---|
| ①445統合/446466保存型/489共有操作/527本人正規生成受領 | common aggregate、LINE限定版とguard、staff policy、492 anchor後続差分、API/registry/lookup/型、Q/ContextGate/同TX adapter | 文書選定でAPI/DDL/model/生成・runtime提供済みにしない。新migration番号予約なし |
| ②445医院設定 | common設定UI/現在read/write、警告preview・確認/取消・CAS、旧レジ休日との分離 | 現Aの共有作業へFEを追加せず、後続単一claim |
| ②448/492 LINE設定単一writer | acceptance確認と部分更新、非P分離、担当者policy/元assignment/版/削除再設定、現在read | 現Dへ追加せず、通常勤務/能力の複製なし。493は現在所属readerの境界 |
| ③450 | 既存pending/error/未保存/競合/失権等の共通UI契約へ接続 | 共通部品再実装・全UI完成ではない |
| ④468 | ResolveClinicAcceptanceWindowとResolveRegisteredStartsのstate/scope/版/集合を利用 | 新consumer実装開始や全予約Doneを①②gateへ戻さない |

通常診療の自動生成は維持する。trimming明示指名は当該staffの登録集合、指名なしは「当該時刻を登録済み」を各staff適格性の辺へ入れる。単純union＋空き人数で受入せず、仮候補を実担当へ保存しない。候補/確定前は同resolver＋最新設定/勤務/休業/能力/全体予定重複を読む。既存予約占有を新start設定で遡及消失させない。

①契約公開/API設計は先行可。必要schema適用・model/reader/guard整合確認前のruntime有効化禁止、BE前の新FE公開禁止。①②設定保存の完了を④全体Done待ちにしないが、実予約へ利用開始する時は対応consumer切替・全必要writer閉包・実値照合が必須。既存医院値を自動コピー/統合/上書きせず、実DB適用は今回しない。

A02/A03の技術選定は本稿で閉じる。実医院値/移行元確認、後続claim、API/DDL/model/生成/製品実装・consumer接続・受領・実試験は残る。全チケット/全AC/①〜③の実装完了を宣言しない。

## 8. fixture（すべてNOT_RUN）

- common未設定/正常空/週次複数区間/休日/祝日方針、24:00・日跨ぎ明示2区間、旧LINE HHMM adapter、JSON破損/取得失敗の非補完。
- initialize同時競合、update CAS/no-op、holiday二namespaceのaction限定・同日同理由/異理由・元family lookup、無権限/別院非開示、audit/receipt失敗の全rollback。
- 共通縮小mask/拡大時LINE保持、既存LINE未照合→confirm、P更新版/非P不増分、秘密非投影、旧全量PUT停止/非P救済経路、未切替医院保持、旧P writer迂回拒否。
- common利用後の旧shifts停止、新commonと旧closingレジ休日の独立編集、予約/common旧表非参照、旧ID/履歴保持、dual-writeなし。
- staff policyの週次/日別置換/明示空/例外削除、未設定/取得失敗/他staff fallbackなし、tombstone create拒否/reconfigure同identity、assignment変更/失権、元anchor lookup/異intent/異command/unknown/newkey禁止。
- previewと保存間の予約集合/版/日時/状態変更→再確認、警告了承保存/取消write0、view不足の非開示・確認不能、LINE overrideなし、予約を自動変更しない。
- 通常予約の終了=閉店、ホテル18:00可/18:01不可、休憩開始端点/休日/LINE制限/定員は独立拒否、24:00の実JST日判定、宿泊途中休診の全滞在拒否を新設しない。
- ④staff登録集合と現在適格性の同時評価、指名なしのmatching、既存占有保持、仮候補を実担当へ保存しない。fixture準備/実行/実機受入は別判定。

## 固定根拠（9eef73fe、採用コードは目標実装の証拠にしない）

- [D17本人決定](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/spec/domains/clinic-settings/README.md#L88)
- [LINE正本：分類別・担当者別/ホテル](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/spec/line/reservation-spec.md#L58)
- [予約警告/根拠再評価](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/design/domains/reservations/technical-spec.md#L208)
- [既存LINE model](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/model/line_reservation_setting.go#L8)
- [既存LINE request](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/reservation/line_reservation_setting_request.go#L1)
- [旧LINE repositoryの独立TX](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/reservation/line_reservation_setting_repository.go#L86)
- [既存医院LINE route/grant](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/reservation/routes.go#L139)
- [旧共通slot request](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/reservation/reservation_type_request.go#L133)
- [旧slot保存scope](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/model/reservation_type.go#L95)
- [旧slot reader](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/reservation/reservation_type_available_slot_repository.go#L31)
- [旧shift休日route](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/clinic/clinic_holiday_handler.go#L94)
- [旧closing休日route](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/clinic/closing_settings_handler.go#L151)
- [元target metadataの既選定](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md#L115)
