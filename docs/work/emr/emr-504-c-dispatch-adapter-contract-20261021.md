# EMR-504 C：caller目的・現在権限・provider照合の採択技術契約

2026-10-21。採択契約・runtime未実装・NOT_RUN。Phase1の最後優先の技術slice Cを具体化する。既存A/B・start_time > server now・GET/POST lock再検査は変更しない。外部送信、個人情報転送、資格登録、料金契約、権限付与、DB適用、試験の実行許可ではない。runtime未実装、NOT_RUN。

## 1. 根拠と限定範囲

main固定 `e71ddf03462a0ee7a9864ad2ce6854b243591770` の以下を読取した。

- [D15正本](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/spec/line/lstep-integration.md)：DM目的別、確認付きfollowup、unknown保留。
- [実route/grant](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/lstep/routes.go)：manual send/健診sync owners:edit、owner履歴/preview owners:view、設定hospital-settings、global tag mutation system-admin。
- [現手動send](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/lstep/line_send_service.go)：旧owner LINE列、送信後best-effort log/audit、errorのみprovider。これは目標適合済みではない。
- [LINE Client](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/infra/line/client.go)、[LSTEP Client](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/infra/lstep/client.go)：両方errorのみ、受理ID／attempt照会の型付き結果なし。LSTEPはGetUser/Tagsを持つが配送照会ではない。
- [保存model](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/model/storage_messaging.go) と [intent](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/line_dispatch_intents.md)／[attempt](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/line_dispatch_attempts.md)／[result](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/line_dispatch_results.md)／[confirmation](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/followup_confirmations.md)：result succeededはexternal_id必須。confirmation purposeはgeneral_guidance/medical_followup、dispatch purposeは別enum。
- [既採択A/B＋Cの入口](https://github.com/MinoruSoga/AnimalEkarte/blob/5a1e00204457084b159346b765cf0002cc8ed164/docs/work/emr/emr-503-504-selected-wire-contract-20261021.md#4-emr-504phase1内の技術整備ab-と後続技術研究c)。rootの固定17件補正のsupport execution lifecycleを同じ意味でLINEへ適用する。support表をLINEへ流用しない。

本書はCのcaller/権限/dispatch/providerと、§3Aの医師要フォロー指定・スタッフ確認writerを新たな技術選定として閉じる。既存doctor列や未実装routeを実装済みとはしない。選定とruntime適合は別であり、必要sliceの実装前に対象callerを有効化しない。マーケ全体除外やA/B停止にはしない。

## 2. purposeの決定方式と全caller境界

永続purposeは既存6値だけ：reservation_confirmation / reservation_change / reservation_reminder / guidance / promotion / medical_followup。unknownという新purposeを保存して送らない。目的未特定は判定結果blockedでありdispatch生成0。

caller adapterが実sourceと確認済みtemplate/scenario定義からpurposeを導出する。clientの文字列purposeやtag名で予約例外を自己申告できない。source無関係な自由文は予約例外にせず、一般案内／販促／医療案内の明示分類と根拠を要求する。分類はDM抑止を回避する新権限ではない。

| caller集合 | 固定する分類・追加条件 |
|---|---|
| 予約作成確認／変更通知／確定予約リマインド | 実appointment＋元操作／既存通知stageが存在し、内容が当該通知に対応する場合だけ各reservation_*。販促の混入で例外へ拡張しない |
| first_visit_welcome / first_visit_followup_3d / first_visit_followup_7d / checkup_followup | followup confirmation必須。general_guidance→guidance、medical_followup→同名。一般確認を医療目的へ流用しない |
| next_visit_tentative_reminder_3d / _0d | 未予約意図であり予約例外不可。既存意図日/stageを元事象とし、確認済み内容に従うguidance（医療案内なら医師指定等の契約が必要） |
| dormant_prevention_365d / _240d / _210d / _180d、birthday_message | 予約例外不可。挨拶／一般来院案内はguidance、販促内容ならpromotion。採用済み実templateとの対応が未確認ならblocked |
| filaria_alert / flea_tick_alert、vaccine_deadline_60d / _30d、food_refill_reminder | 予約例外不可。実内容が一般案内／販促／個別医療フォローのどれかを版付きmappingで識別。名称だけの自動分類なし、既存source安全条件維持 |
| next_visit_reminder | 退役。履歴のみ、intent/claim/外部発火0 |
| 手動LINE sendと全alias | 既存payload制約を保持し、同じ内部dispatch境界を通す。purposeだけで予約例外を許さず元appointment/通知事象の証拠を要する。pdf/imageは同clinicだけでなく対象owner/file原本対応を検証 |
| 健診対象者一括sync／シナリオ発火tag／手動tag付与 | メッセージ発火につながる操作は同じ目的判定とDM・確認条件を通す。bulk全体を1人のreceiptにせず、owner別intentと結果を追跡。部分結果は明示 |
| ordinary CPM/VISIT等の状態同期・プロパティ更新 | 配送と別の既存冪等同期。ただし外部設定上メッセージを発火するmappingならordinaryと呼んで迂回不可。発火有無不明の外部writeは保留、内部計算/readは可 |
| Webhook follow/unfollow | 既存署名/医院解決＋469 binding writer。イベント受信を営業送信許可にせず、新replyを追加しない |

上表は全17登録triggerと主要caller族を包含する。実template本文・scenario配線・provider契約の意味を捏造しない。deployment manifestはcaller_code、purpose、source_schema、template/scenario revision、configuration_digest、effect_class（ordinary_state / message_trigger / direct_message）を閉じた対応表として保持する。曖昧なmappingだけblockedとし、一般に機能をPhase2へ除外しない。既存医院設定の管理権限を保持し、このmanifestに新しい運用UIや新grantを要求しない。

DM必要を要求するのはguidance/promotion/medical_followup。reservation_*だけDM不要／未設定を理由に除外しない。他の独立停止・本人/宛先・医院・死亡由来・既存配信gate・daily claimは全purposeで保持。DM読取不明もblocked。既存臨床判断・配信日数・頻度・CPM閾値は変更しない。

## 3. 認可・source adapter

- 手動send、followupスタッフ確認保存、owner配下の管理照合／retryは既存owners:edit。GET context/intent/結果照会はowners:view。いずれも元targetの現在医院/owner scope、参照するmedical-records/checkups/reservationsの既存viewを追加照合し、読み取れない臨床本文を返さない。確認保存は臨床本文編集を兼ねない。
- 既存手動tag DELETEのowners:delete、医院設定のhospital-settings:view/edit/delete、global tag mutationのsystem-adminを維持。監視view・analytics:view・接続test成功をsend権へ昇格しない。
- 医師指定のwriterは既存臨床ownerのwrite認可と医師actor条件が必要。スタッフ確認payloadにdoctor_staff_id/doctor_atを受理して指定事実を作らない。現在担当医IDや職種名だけを指定証拠にしない。
- scheduled workerは人のgrantを偽造しない。既存job実行scope＋医院設定＋保存済みstaff確認の有効性・現在source条件を検査。手動操作の元actorが必要な経路は現在assignment/grantも再照合し、失権で新start不可。
- Pure GET内self-repairなし。preflightに加え、C06 Apply内とdispatch-start TX内で元anchorから現在認可を再検査。POST replayも現在edit、GET原receiptは同actor/clinic＋現在read。

Source adapterの出力はimmutable typed projection：schema、source_kind（reservation / first_visit / checkup / vaccination / birthday / tentative_visit / owner_segment / food_purchase / manual_message）、source_id、event_key、source_version、source_digest、clinic_id、owner_id、pet_id nullable、performed_confirmation_id nullable、doctor_designation_ref nullable。ID/版は正signed64文字列、digestはUTF-8 JCS SHA256 lowercase hex。event_keyはcallerの既存事象identity/stageを復元し、現在時刻や新operation keyを代用品にしない。臨床本文をdispatch用コピーとして無目的に増殖させない。

followupでは既存followup_confirmationsのrecord必須・任意checkup・同患者制約を守る。現DDLのcheckup自身もrecord必須であり、独立画面のPath Bは既存record保存へ接続する。架空recordを作らず、保存途中・scope不明は停止する。具体source／医師原本／確認wireは次節で選定する。

## 3A. C3の臨床原本writerとスタッフ確認wire（今回の新技術選定）

### 根拠・所有

D15 [FR-crm-followup-confirmation](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/spec/line/lstep-integration.md#requirement-fr-crm-followup-confirmation) は「医師の要フォロー指定」と「スタッフの実施済み・送信対象確認」を別事実とする。指定の保存route/tableは既存にあるとは扱わず、本節で追加する。保存ownerはmedicalrecord、配送ownerはlstep。根拠：[実際のrecord/健診route](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/routes_records.go)、[record DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/medical_records.md)、[健診DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/checkups.md)、[健診項目DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/checkup_field_results.md)、[staff DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/staffs.md)。健診のmedical_record_idはNOT NULL、record配下の健診編集／項目保存はmedical-records:editでありcheckups:editではない。

これはカルテ全体の医師確定、確認済み診療ロック、会計承認、検査結果確定ではない。医師確定ボタンを復活させず、保存済みの該当record/健診の「要フォロー」属性を医師本人が明示変更する操作だけを用意する。新しい医療判断閾値・送信時刻・必要な診療行為を追加しない。一般案内に医師指定は不要。過去入力・訂正・転記・復元は指定や確認を自動複製しない。

### 臨床sourceのclosed projection

schema='clinical-followup-source-v1' の全keys：clinic_id, owner_id, pet_id, source_kind（first_visit|checkup）, source_id, medical_record_id, record_version, clinical_set_generation, clinical_set_digest, checkup。前6IDとrecord_versionは正十進文字列。generation/digestは既存494の確立済み値を必須とし、未初期化をGETや本writerで勝手に生成しない。checkupはfirst_visitならnull、checkupなら次のclosed object：id, version, medical_record_id, pet_id nullable, checkup_type_id, date（YYYY-MM-DD）, next_date nullable, doctor_id nullable, result（保存text）, field_results。field_resultsはidの数値昇順の配列で、各objectは id, checkup_type_field_id nullable, field_name, field_type, unit, value_number nullable, value_text, value_bool nullable, value_list, ref_min nullable, ref_max nullable, is_abnormal, status, sort_order。型とenumは固定参照DDLの列型を保持、decimalはExactNumeric canonical文字列、整数sort_orderは既存int32、text配列の順序は保持。timestampsや現在master表示名を足さない。削除済みsourceはprojectionを生成せず非適格。NULLのcheckup.pet_idは親record.pet_idから対象を検証できる場合だけ許容し、非NULLなら親と完全一致を要する。

source_digest=SHA256(JCS(projection))。source_versionはfirst_visitではrecord.version、checkupではcheckup.version。両者を合成した架空数値やtimestampを版にしない。digestはrecord/clinical_set版と健診の全保存結果を束縛し、checkup.versionだけに頼らない。指定履歴・スタッフ確認自体はこのprojectionに含めず自己無効化しない。内容を元へ戻しても494 generationとsource版の進行を巻き戻さない。first_visitの適格性は既存初診source判定の結果を用い、単に任意recordを初診と申告できない。

この完全projectionは内部immutable snapshotとして保存し、公開contextでは本文を別権限へ漏らさず既存臨床読取権限でのみ取得できる。source ownerは全対象writer（record編集/削除/復元/移管、494臨床集合変更、checkup編集/削除/結果置換/取込）で既存版進行と同じ親record lockを守る。外部取込を含む子の直書きで親lockを迂回しない。価格だけの別domain更新から新しい医師指定を要求する規則は追加しない。

### 医師指定：新しいappend-only原本

medicalrecord所有のclinical_followup_designationsを追加する。columns：id BIGSERIAL PK, clinic_id/owner_id/pet_id/medical_record_id BIGINT NOT NULL, checkup_id BIGINT NULL, revision BIGINT NOT NULL, required BOOLEAN NOT NULL, source_version BIGINT NOT NULL, source_digest BYTEA NOT NULL, source_snapshot JSONB NOT NULL, actor_staff_id BIGINT NOT NULL, recorded_at TIMESTAMPTZ NOT NULL server時刻, operation_id BIGINT NOT NULL。ID/revision/source_version>0、digest長32、source_snapshotは上記closed schemaで関係列と一致。UNIQUE(id,clinic_id)、UNIQUE(operation_id,clinic_id)、対象ごとrevisionのuniqueはcheckup_id NULL用(clinic_id,medical_record_id,revision) WHERE checkup_id IS NULLと非NULL用(clinic_id,medical_record_id,checkup_id,revision) WHERE checkup_id IS NOT NULLの2本。FK owner/clinic、record/clinic/pet、checkup/clinic、actor_staff/clinic、operation/clinicは既存複合uniqueを参照しRESTRICT。checkupのrecord/pet一致は同lock下のwriter検証＋制約trigger、更新削除禁止trigger、apply_rls_policy、医院scope repository必須。履歴を現在ownerへ付け替えない。

同対象親record lock下でrevision=max+1（初回1）、overflow拒否。最新revisionのrequired=trueで、snapshot/digestと現在sourceが一致する時だけ有効。falseは明示取消履歴であり物理削除しない。source変更後に古いtrueを自動復活させない。既存指定行があり同じrequired・同じsourceで別keyの反復は変更なしreceipt、指定event/revision/auditの新規生成なし。同key成功は元receipt。変更なしでも新操作のC06結果は保存する。既存doctor_id／staff_typeだけは指定事実の代替にならない。

新route prefix /api/v1：
- GET /medical-records/{record_id}/followup-context?source_kind=first_visit|checkup&checkup_id=...。first_visitでcheckup_id禁止、checkupでは必須。現在medical-records:view。response closed：source（上記projection）, source_version, source_digest, designation（null又はid/revision/required/actor_staff_id/recorded_at/source_digest）, designation_valid:boolean, eligible:boolean, blocking_reason（null|source_unavailable|not_first_visit|source_changed|partial_save|not_supported）。GETはno-store、同一整合snapshot、補完writeなし。
- POST /medical-records/{record_id}/followup-designations。closed body：operation_key UUID, source_kind, checkup_id nullable, required boolean, acknowledged_source（上記projection）, expected_source_digest, expected_designation_revision nullable, edit_session_id, edit_epoch。edit_session_id/epochの型・意味は既存494。acknowledged_sourceは確認した保存結果としてimmutable Content、期待現在digest/指定revision/edit値はPreconditions。Applyでacknowledged_sourceとdigestを再計算し現在sourceと一致照合。要求のactor/clinic/owner/pet/doctor日時独立fieldは受理しない。
- GET /medical-records/{record_id}/operations/{operation_key}?intent=medical_record.followup_designate。既存record operation readerへintentを追加し、元anchor・同actor/clinic・現在medical-records:viewで原receipt取得。新POSTも成功replayには現在medical-records:edit middlewareが必要、viewのみはGET。

医師指定POSTはmedical-records:edit＋494現在edit holder/fence＋同医院の現在有効staffでstaff_type='doctor'、authenticated accountに対応するactor本人を要求する。occupation名・client doctor_id・system-adminだけから医師本人を作らない。これは既決医師による指定のactor検証で、新RBAC grantを作らない。preflightとApplyで現在assignment/actor条件を再検査する。履歴の医師が後に退職しても履歴を削除しない。新送信で必要な現在source/指定条件との関係は§3の現在検査を適用する。

### スタッフ確認：既存表の限定拡張

followup_confirmationsを再利用し、新しい別確認registryを作らない。追加columns：source_snapshot JSONB NOT NULL（新規行）、recipient_binding_revision BIGINT NOT NULL（新規行、正値）、doctor_designation_id BIGINT NULL＋FK(id,clinic_id)→clinical_followup_designations RESTRICT。既存source_version/digestは上記projectionの版/digest。medical_followupでは同対象・同sourceの現在有効なrequired=true指定ref必須、doctor_designated_by/atは当該原本actor/recorded_atからserver転記して一致検証。general_guidanceでは3医師関連列すべてNULL。confirmed_byとperformed_confirmed_atは認証staff/server記録時刻であり、診療実施日時を捏造しない。旧行は架空snapshot/refでbackfillせずlegacy未検証として新配信適格にしない。新行のclosed snapshot、列対応、同clinic/owner/pet/record/checkup整合、append-only trigger、RLS、operation同clinic対応を必須にする。運用済み旧行保全のためnullable追加→新writer制約の順を許容し、既存実DB状態を本書から断定しない。

- POST /owners/{owner_id}/followup-confirmations。closed body：operation_key, medical_record_id, checkup_id nullable, purpose（general_guidance|medical_followup）, acknowledged_source, acknowledged_binding_revision, doctor_designation_id nullable, performed_confirmed（trueのみ）, recipient_confirmed（trueのみ）, expected_source_digest, expected_designation_revision nullable, expected_binding_revision。binding各版はcanonical正十進文字列、acknowledged版はContentへ束縛して保存し、expected版はCAS。両方を現在469 bindingと一致検証し、未連携・本人不一致は確認保存不可。source_kindはacknowledged_sourceから導出し二重入力なし。owners:edit＋medical-records:view、元owner/患者/record現在scope、同医院active staff本人を必須とし、臨床本文編集権や医師職種は一般確認に要求しない。falseは未確認であり保存成功へ変換しない。新しい診療実施eventを書かず、スタッフが当該保存結果と送信対象を確認した事実だけを記録する。
- GET /owners/{owner_id}/followup-confirmation-context?medical_record_id=...&checkup_id=...&purpose=... を既存候補から採用する。record/purpose必須、checkup省略はfirst_visit、未知/重複query拒否。owners:view＋medical-records:viewで§3A臨床contextの全keysにbinding_revision nullable、binding_state、purposeを加えたclosed responseを返す。binding_stateは469の既存enum。現在本人/対象scopeを検査し、secret/raw LINE IDは返さない。
- GET /owners/{owner_id}/line-operations/{operation_key}?intent=line.followup_confirm を§7allowlistへ追加。owners:view＋元record現在view、同actor/clinic。POST replayはowners:edit、view-only回復はGET。

Path Bは既存段階保存を全面atomic化しない。UIは保存段階にunknown/busyが残れば確認POSTを出さず、既存結果照会を先に解決する。serverはロック下でacknowledged_source全体と現在保存結果を照合し、途中版を送信許可にしない。pending C06臨床writerがある場合はその既存bindingを解決するまで停止し、snapshot一致だけで未確定writeを不存在と推定しない。単なる記録作成・保存・要フォロー指定は確認POSTを代行しない。確認自体はqueue/start/provider HTTPを実行しない。

### C06・監査・通知接続

指定はowner_kind=medical_record。スタッフ確認はLINE operation consumerのowner_kind=lineに統一する新技術選定とし、既採択504のnotification_resume/reconcile/retryも同じconsumer定数を参照する。operation_bindingsのowner_kindは非空textでDB enumではないため、新DB enumは作らない。既存bindingを別kindへ改名・再bindingせず、既存採用定数がlineと異なる実装は互換adapterで旧lookupを維持する。Content.Fields exact keys：指定はschema='followup-designation-v1', record_ref, checkup_ref（first_visitは文字列'none'）, required（'true'|'false'）, command_digest。確認はschema='followup-confirmation-v1', owner_ref, record_ref, checkup_ref, purpose, command_digest。全値string、各256byte以内。command_digestは全immutable commandのJCS SHA256。指定commandはschema/source_kind/checkup_id/required/acknowledged_source、確認commandはschema/medical_record_id/checkup_id/purpose/acknowledged_source/acknowledged_binding_revision/doctor_designation_id/performed_confirmed/recipient_confirmedで閉じ、CAS/operation key/生成ID/現在時刻は含めない。doctor refはContentであり、別指定への置換は新明示操作。Applyは同じimmutable commandから全Contentを再計算照合。C06の12status/HTTP、claim前元anchor metadata、ContextGate、元targetlookup、失敗同key照会は既採択503/504・489を再利用する。

指定のlock順は494/518臨床順（Gate→binding→現在認可→edit権→record→必要子→指定履歴）。確認保存とdispatchのsource再読は504 graph/owner/binding側の既存順の後、record→checkup→field_resultsをID昇順→指定履歴の順にする。臨床writerから後置で配送owner lockを取りに行かない。指定は配送queueに触れず、配送は後の独立判定にするため逆順を作らない。全関与writerの親record lock参加がcutover条件。現在sourceと指定はclaim前とstart TXの両方で照合する。

成功refs：指定はrecord_id, designation_id（変更なしは既存id）, designation_revision、確認はowner_id, record_id, confirmation_id。成功committed_versionsは指定={designation:正十進文字列revision}、確認={confirmation:"1"}とする（immutable idを版代替にしない）。新確認で既存同一owner/record/checkup/purpose/source_digest/binding revision/doctor refの確認が存在する場合は元行refを返す変更なしreceipt、確認行重複なし。親record lockにより並行重複を直列化し、完全snapshot比較も行う。別目的は別確認だが送信重複抑止を解除しない。

必須audit action=medical_record.followup_designate / line.followup_confirm。resource_typeはそれぞれmedical_record / owner、resource_idは元anchor。old/new closed：designation_id nullable, designation_revision nullable, required nullable, confirmation_id nullable, source_digest。metadata closed：schema='clinical-followup-audit-v1', operation_id, record_id, checkup_id nullable, purpose nullable, actor_staff_id。旧状態がないときold=null、指定でconfirmation_id/purpose=null、確認で該当指定がなければdesignation関連null。snapshot本文は監査へ複製しない。原本／確認・必須監査・C06成功結果を同TX。no-op/replayは業務監査を重複追加しない。必須監査・CAS・認可失敗は保存0、外部作用0。

§3一般projectionのperformed_confirmation_idはこのconfirmation_id、doctor_designation_refはnull又はclosed {id, revision, source_digest}。source_idは元record又はcheckup ID、event_keyはfirst_visit:{record_id}又はcheckup:{checkup_id}（stageはdelivery_stageへ分離）。再確認の新confirmation ID・指定改版・臨床版変更をevent_keyへ混ぜて同一診療の自動再送を発明しない。既存3日/7日/welcome/健診stageのみへ接続する。確認後source変更・指定取消・binding revision変更・対象移管・必要読取不能は古い確認を無効として保留、歴史行と過去succeededを保持する。一般確認を医療目的へ変換しない。必要な再確認後も同semantic成功・unknownを越えて新送信しない。

## 4. 内部CreateDispatch・同一事象dedup

内部command closed keys：schema='line-dispatch-command-v1', source（§3projection）, purpose, followup_confirmation_id nullable, destination, payload, delivery_stage。destination closed keys：provider（line_messaging / lstep_trigger）, integration_ref, recipient_binding_revision, configuration_digest。payloadは既存text/pdf_url/image_url/tag_triggerの判別union。textはtype/text、fileはtype/file_id/file_version/file_digest/file_name、tagはtype/tag_code/scenario_ref/scenario_versionのみ。自由URL、credential、raw LINE IDをpublic commandに受理しない。file URLは対応fileからserver生成し原本digestを照合、署名URL自体はsemantic hashへ入れない。

provider、recipient、template/scenarioとpayload commitmentを内部snapshotへ固定し、C06 Contentにはschema＋source identity/event_key/purpose/delivery_stage＋canonical command digestを全て束縛する。現在CAS・期待binding/source版の再検査値はPreconditions、確認した固定source snapshot版はContent。typed commandからApply前に再計算しunchecked payloadを送らない。

line_dispatch_intentsに新規semantic_key bytea(32相当CHECK)とsemantic_payload jsonbを追加し、UNIQUE(clinic_id,semantic_key)。projectionはschema='line-dispatch-semantic-v1', owner_id, source_kind, source_id, event_key, purpose, delivery_stage。operation key/actor/現在binding/template変更を含めない。同じ元事象を別key・再link・テンプレート編集で再送しない。完全projection一致を比較しhashだけに依存しない。既存daily priority/claimは併用。

同semantic既存intentがあればそのimmutable payloadとの差を検査。同じ内容でも別keyの新intentを作らずdependency-changed409として既存intent読取へ案内し、別keyの成功receiptを偽造しない。内容が変わった場合も同元事象の自動再送はしない。元sourceに既決の別通知事象が生じた場合だけ別event_keyを使う。新しい再通知業務をこの技術規則から作らない。legacy復元不能はsemantic不明のまま保護し、新INSERT用NOT NULL化に架空keyをbackfillしない。旧経路の同一事象claimを照合できないcallerはcutover不可。

## 5. claim/start/result lifecycle：support17補正との一致

line_dispatch_attempt_executionsをowner-private新設。columns：clinic_id bigint NOT NULL、attempt_id bigint NOT NULL、intent_id bigint NOT NULL、state text NOT NULL、fence bigint NOT NULL DEFAULT 1、started_at timestamptz NULL、finished_at timestamptz NULL。PK(clinic_id,attempt_id)。FK(attempt_id,clinic_id,intent_id)→line_dispatch_attempts(id,clinic_id,intent_id) RESTRICT。state=claimed/started/finished/revoked、fence>0。claimed/revokedは両時刻NULL、startedはstarted_atのみ、finishedは両非NULLかつfinished_at>=started_at。許可遷移claimed→started、claimed→revoked、started→finishedだけ、expected state/fence CASでfenceを増分。overflow停止、DELETE禁止、tenant RLS/apply_rls_policy、scope repository。attempt/result履歴のimmutable triggerと別に制御行の許可遷移triggerを置く。

取得順は504 Bのgraph→C06/現在認可→owner→recipient→binding→必要pet/appointment/eligibilityを維持し、その後source ownerの既存lock順→intent→execution。source依存の共通取得順を参加writerで固定できないcallerは有効化しない。設定writerも非secret integration identityの共通同期規約へ参加する。DB gateと外部HTTPがatomicだとは保証しない。

1. claim TX：現在条件・既存claim・最新attempt・全歴成功なしを検査。attempt＋execution(claimed)＋line.dispatch_claim監査を同TX commit。失敗はHTTP0。
2. start TX：同lock下で現在source/確認/医師指定/binding/DM/停止/gate/configuration_digest/payloadを再検査。claimed→started CASとline.dispatch_start監査を同TX commit。このcommitが開始の線形化点。失敗はHTTP0。以後固定した宛先・payloadを一度だけ外部call。開始後に変化しても外部効果回収を保証しない。
3. 確定応答後：started→finished＋result＋line.dispatch_result監査を同じ結果TXへ。DB/監査失敗は外部効果不明でblocking unknown。startedのままでも二重claim不可。finishedは応答処理完了で、配送到達や不在証明ではない。
4. 照合：未開始claimedは同lockでrevokedへCASして旧workerのstartを永続拒否。started/finishedは元呼出しが今後効果を作れない終結証拠＋当該attemptの不在証拠の両方が必要。timeout、lease、DB result=NULL、fence変更、単発検索不在、cancel要求だけでは不足。証明不能ならunknown維持、再送0。
5. 一度でも同intentのsucceededがあればterminal。provider側の後日削除でverified_absentへ降格させない。retryは最新verified_absentと元attempt一致をlock/CAS照合して同intent次attemptのみ。definite_failure単独では再送不可。既存DB retry guardがより緩い場合は同条件へ補正するまで有効化しない。

既存attemptにexecution行がないことを未開始とみなさない。移行証拠不足はunknown、GET補完禁止。HTTP自動retryはeffectful callについて原則0。providerが同attempt_keyの同一payloadを重複効果なしで受理する証拠のあるadapterだけ同一keyのtransport retryを許せるが、期限切れ後は新keyにせずunknown。既存clientのmaxRetries=3を無条件継承しない。

## 6. provider能力契約と結果の意味

provider capabilityはversioned manifest（server-owned）。closed fields：schema='line-provider-capabilities-v1', provider, adapter_version, effect_scope（message_acceptance / scenario_trigger_acceptance）, acceptance_evidence_supported:boolean, idempotent_retry_supported:boolean, reconcile_supported:boolean, terminal_absence_supported:boolean, evidence_reference:string。未知値はfalse。evidence_referenceは確認した契約/版への非機密参照で、ユーザー提供本文や任意URLを実行しない。能力宣言とその実装を満たさないadapterは対応操作をblocked。公式応答契約をまだ確認していないadapterは未実証と記録し、receipt ID必須という架空条件で永続的に成功不能とは決めない。

| adapter | 初回sendの条件 | succeededの意味 | 照合／verified_absent |
|---|---|---|---|
| line_messaging | 同じbindingから解決した宛先、現在認可、固定payload、開始fence。typed responseで当該requestに相関する公式契約上の受理応答証拠を取得できること（提供protocolの必須IDを含む） | providerが当該message requestを受理。端末到達・既読保証ではない | 照会能力・終結/不在証拠が確認できる場合だけ。現error-only Clientからは保証しない |
| lstep_trigger | versioned scenario/tag mappingが目的と送信対象に対応し、発火作用が特定されること。二重gate必須。公式契約上の発火request受理を確認できるtyped evidence能力が必要（external IDは任意） | 当該発火requestの受理。シナリオ内全メッセージ配送成功とは別 | GetUser/GetTagsやtag現在存在／不在は当該attemptの発火・配送証明ではない。能力未確認ならunknown、retry不可 |
| ordinary_state | 既存冪等状態同期契約、発火しないという設定対応が確認できること | dispatch結果を作らない。状態同期結果として既存経路で報告 | メッセージ不存在証明へ転用しない |

現DDLの `CHECK(state<>'succeeded' OR external_id IS NOT NULL)` は、外部IDを返さない正式成功protocolを表現できないため限定変更する。external_idのnullable列型は維持するが、全resultでexternal_id IS NOT DISTINCT FROM evidence.provider_receipt_id（JSON nullをSQL NULLとして比較）を保存検証する。現在の成功枝line_push_200_v1では両者非NULLかつprotocol_evidence.request_idと完全一致を必須とし、将来確認済みのno-ID成功枝を閉じたunionへ追加した場合だけsucceededの両NULLを許す。succeededは evidence.schema='line-provider-evidence-v1'、evidence_code='accepted_response'、protocol_evidence内のresponse_contract_ref非空、当該protocolで成功を示すresponse_status／必要fieldの検証済み証拠を要求する保存検証へ置換する。DB側はclosed shapeと必要項目・型、service adapterは公式応答意味とrequest相関を検証する。任意2xxやclient申告を成功へ昇格しない。providerの正式受理を示すresponse bodyが必要な場合は、その非機密型付きprojectionをadapter版の証拠契約で固定し、raw bodyを保存しない。schema/model/repository/fixtureの変更がC2の利用開始条件。support側のexternal ID要件はこのLINE限定補正から変更しない。

内部adapterは、SendもReconcileも保存と同じ `ProviderEvidenceV1` 型をそのまま返す。Send closed shapeは {outcome:accepted|rejected|unknown, provider_observed_at:null|RFC3339, evidence:ProviderEvidenceV1}。Reconcile closed shapeは {outcome:accepted|verified_absent|unknown, evidence:ProviderEvidenceV1}。外部ID/応答status/契約ref/必須protocol fieldsを別のcaptured変数から足さない。outcomeとevidenceの判別枝の整合を保存前とDB制約で検査する。occurred_atはSendの信頼できるprovider_observed_atのみ（無ければNULL）、evidence.observed_atはserver観測時刻であって発生日時ではない。

ProviderEvidenceV1 の共通closed keys：schema='line-provider-evidence-v1', provider（line_messaging|lstep_trigger）, adapter_version（非空string）, effect_scope（message_acceptance|scenario_trigger_acceptance）, evidence_code, provider_receipt_id（null|非空string）, attempt_key（元attempt UUID）, payload_digest/configuration_digest（各64桁lowercase hex）, observed_at（server RFC3339）, original_execution_terminal:boolean, absence_proven:boolean, protocol_evidence。protocol_evidenceは次のclosed discriminated unionだけを許す。raw response、任意map、credential、quote token、本文、宛先を保存しない。

- `line_push_200_v1`：{kind:'line_push_200_v1', response_status:200, response_contract_ref:'https://developers.line.biz/en/reference/messaging-api/#send-push-message', request_id:非空string, sent_message_ids:string[]}。公式[Push message応答](https://developers.line.biz/en/reference/messaging-api/nojs/#send-push-message)を2026-10-21に読取確認：200とsentMessagesを返すが受信者到達は保証しない。adapterは当該POST /v2/bot/message/pushの実応答からのみ構築し、sentMessages各idを損失のないopaque文字列として保持する（JSON numberなら元整数tokenを厳密解析、JS float経由禁止）。要求message数と対応する非空配列、1～5件、空ID拒否。request_idは公式共通response header x-line-request-idから保持し、この提供protocolで欠落した場合はunknownとする。provider_receipt_idと保存external_idは両方非NULLでそのrequest_idと完全一致し、メッセージID配列から架空単一receiptを作らない。evidence_code=accepted_response、provider=line_messaging、effect_scope=message_acceptance、terminal=true、absence=false。Send outcome=accepted、保存succeeded。Reconcileのacceptedは同attemptについて既に保存したこの権威ある実応答を回収できた場合だけで、GetTags/現在状態から生成不可。
- `unresolved_v1`：{kind:'unresolved_v1', response_status:null|integer(100..599), response_contract_ref:null, reason:'transport_unknown'|'lookup_unavailable'|'unsupported_protocol'}。evidence_codeはtransport_unknown又はlookup_unavailable、provider_receipt_id=null、保存external_id=null、absence=false。terminalは当該元呼出しの実証に従い、不明はfalse。outcome=unknown、保存unknown。2xxだけでacceptedへ変換しない。
- `local_revoked_v1`：{kind:'local_revoked_v1', response_status:null, response_contract_ref:null, attempt_id:正十進string, revoked_fence:正十進string}。元attempt/executionを同TXでclaimed→revokedとし、未開始で旧workerも開始不能と証明した場合だけ。evidence_code=local_revoked_before_start、provider_receipt_id=null、保存external_id=null、terminal=true、absence=true、Reconcile outcome=verified_absent。DBのexecution行/fenceとの一致を検証し、clientがこの枝をsubmitするAPIはない。

本書で確認済みの外部成功枝はline_push_200_v1だけ。LSTEP独自APIの正式成功／拒否／照合／不在証明の応答契約は未実証なので、その枝を捏造しない。LSTEPは未知能力falseで新send開始不可、既存unknownは保留。将来、公式契約を確認したadapterごとにresponse_status/response_contract_refと必須protocol fieldsを列挙した新しいunion枝・DB CHECK・fixtureを同じ変更で追加する。receipt IDを返さない正式protocolもその必須成功証拠で表現できるため、全provider共通のexternal ID必須制約は戻さない。現unionにはdefinitive_rejection/外部terminal_absence枝がないためSend rejectedは生成不可、Reconcile verified_absentはlocal_revoked枝のみ。既存歴史definite_failure/verified_absentは保持するが新規証拠をこの枝へ偽変換しない。公式仕様から同attempt idempotent retryや照会能力まで確認済みと拡張解釈しない。

succeeded保存検証は上記evidence全体のclosed shape、accepted_response、対応する成功protocol枝、request相関とexternal_id/evidence.provider_receipt_idのNULL含む同値を要求する。現LINE成功枝は両非NULL=request_idであり、ID欠落を両NULLの成功へ変換しない。将来の確認済みno-ID成功枝以外にsucceeded NULLを許可しない。先述のresponse_status/response_contract_refはprotocol_evidence内のフィールドを指す。戻り値・永続evidence・監査で異なる証拠形式を作らない。一般HTTP400/500、nil error、外部IDの有無だけでは成功／不在を認定しない。

## 7. 管理照合route・receipt・監査

新public管理routeはowner配下に限定する。prefix /api/v1。
- POST /owners/{owner_id}/line-dispatch-intents/{intent_id}/reconciliations：closed body operation_key UUID、expected_attempt_id string、expected_result_revision required nullable string。nullはresult不存在CAS。owners:edit＋元source read。照会HTTPはC06 TX外、保存時に同attempt/fence/結果版再照合。送信0。
- POST 同prefix /retries：closed body operation_key、expected_attempt_id、authorized_retry_result_id。§5条件のみ。新attemptの準備receiptであり外部成功ではない。
- GET /owners/{owner_id}/line-operations/{operation_key}?intent=line.dispatch_reconcile|line.dispatch_retry：元anchorの現在read、同actor/clinicの原receipt。既存line.notification_resumeと§3Aのline.followup_confirmを別intentとしてallowlistをowner契約に追記する。
- GET /owners/{owner_id}/line-dispatch-intents/{intent_id}：owners:view＋元source read。closed response intent_id, owner_id, purpose, effect_scope, latest_attempt_id nullable, execution_state（not_claimed/claimed/started/finished/revoked/unknown）, latest_result（null又はresult_id/revision/state/external_id nullable/recorded_at/occurred_at nullable）, blocking_reason nullable。result.stateは既存4値。execution_stateのunknownはattemptが存在するのに制御行欠落・読取矛盾等で実行状態を解決できない場合。not_claimedはattempt不存在を権威ある読取で確認した場合のみ。保存execution.stateは従来のclaimed/started/finished/revokedの4値を維持し、readerのunknownを保存stateへ加えない。未試行やresult無しを未送信確定にしない。

### C用の元target anchor拡張（新技術選定）

既採択line_notification_operation_targetsのB用metadata（schema/appointment_ids）へrecordやintentを詰め込まない。同じowner-private表をclosed unionへ限定拡張し、Bの既存枝は完全保持する。追加target_kindはowner_line_followup / owner_line_dispatch、target_refは元owner ID。列/PKは既採択のclinic_id,operation_id,target_kind,target_ref,metadata、PK(clinic_id,operation_id)、正BIGINT CHECK、即時FK(operation_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICTのみ。owner/record/checkup/intent/attemptへの業務FKを置かず、短いclaim TXで業務row lockを先取りしない。

- intent=line.followup_confirm：target_kind=owner_line_followup、metadata exact keys {schema:'line-followup-target-v1', medical_record_id, checkup_id:null|string}。IDsはcanonical正signed64 string。record/患者/owner対応はpreflightで解決しimmutable command/Contentへ束縛。lookupは元ownerのowners:view＋元recordのmedical-records:viewと現在同医院・行scopeを検査し、checkupがある場合は元recordとの対応も検証する。失敗bindingにconfirmation行が無くてもこのanchorから認可できる。対象を解決できない場合は別recordへfallbackしない。
- intent=line.dispatch_reconcile|line.dispatch_retry：target_kind=owner_line_dispatch、metadata exact keys {schema:'line-dispatch-target-v1', intent_id, attempt_id, source}。sourceはclosed {kind,id,medical_record_id:null|string,checkup_id:null|string,appointment_id:null|string}。kindは§3の9値のみ。first_visitはmedical_record_id=id/他null、checkupはcheckup_id=id/medical_record_id必須/appointment null、reservationはappointment_id=id/他null、他6種は3補助IDすべてnull。IDsはcanonical正signed64 string。preflightでimmutable intent/sourceから作りContentのcommand_digestへ全束縛する。lookupは元owner現在view、原intent/attemptの医院・owner関係、source現在readを検査。clinicalは元record view、reservationは元appointmentのreservations:view、vaccinationはvaccinations:view、その他はowner view＋当該source ownerの既存readを検査し、読めない場合は開示しない。incoming bodyや現在別sourceへ認可先を差替えない。

DBのtable-local CHECK/検証関数はtarget_kindとschemaの対応、exact key集合、JSON型、nullable形、ID範囲、source kind-shapeを検査する。bindingのowner_kind/intentとの一致はdomain reader/writerでも検査。既採択のUPDATE/DELETE拒否trigger、INSERT/SELECTのみrepository、app_private.apply_rls_policy/has_clinic_access、trusted clinic scopeを維持する。新role、SECURITY DEFINER、修復・補完GETを追加しない。

489のClaimMetadataWriterを利用し、binding INSERT winnerだけが同claim TXで上記anchorを保存、失敗ならbindingもrollback。verified immutable commandとcanonical Contentの一致済み値だけを使い、callback内に業務row読取/lock/外部作用を持込まない。既存bindingにはcallbackを再実行せず原anchorを保持する。認可anchorの欠落/不整合/解決不能とresolver内の認可根拠読取失敗は現Runner同様403 not-authorizedで非開示。元targetと現在開示認可が確定した後の結果読取故障だけ503 result-unavailableとしてfail-closed。医師指定はmedical_record_operation_targetsの元record anchorを使い現在record viewを検査し、record配下checkupの追加write権をlookupで要求しない。医師指定POSTの新executionでは元checkup関係もApplyで検証する。

新writeは[現contract.go](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/operation/contract.go)と[dto.go](https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/operation/dto.go)のC06 12status/ResultResponseをそのまま使用する。HTTPは既採択503/504どおり、succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。busyは説明語でwire statusには使わない。wire keysはstatus,operation_key,result_refs,committed_versions,recorded_atであり後3つはsucceeded/replayedのみ、非成功で成功refs/版/時刻を返さない。pre-C06 ErrorResponse/HTTPも既採択503/504 §7を再利用。Contentはintent_id/attempt_id＋immutable operation内容、CASはPreconditions。failed bindingにも元owner/intent/source anchorをClaimMetadataWriterで同claim TX保存し、本文差替えでlookup認可先を変更しない。success refsはowner_id,intent_id,attempt_id,result_id（reconcile）又はowner_id,intent_id,attempt_id,authorized_retry_result_id（retry）、committed_versionsはreconcile={result:正十進文字列revision}、retry={attempt:正十進文字列attempt_number}。receiptは内部処理成功のみ。

本節のimmutable typed commandは[採択済みQの共通契約](../../design/shared/operation-participation.md#verified-command-selected-20261021)に従い、Applyがreq.VerifiedCommandだけから再構成する。共通constructorとRunnerによる既定Content全key照合・再検証を通し、別command/DTO/pointerのcaptureでwriteしない。Qはrequest-memoryのみで耐久保管とは別、CAS/fenceはPreconditions、Contentは既定の短いstring mapを維持する。489のQ必要subsetは未実装依存であり、既存Content-only consumerは変更しない。


audit_logs actionはline.dispatch_create / line.dispatch_claim / line.dispatch_start / line.dispatch_result / line.dispatch_reconcile / line.dispatch_retry / line.dispatch_revoke。resource_type=owner、resource_id=元owner_id、trusted clinic/actor。old/newはintent_id/attempt_id/execution_state/fence/result_id/result_revision/result_state（未存在null）のclosed shape。metadataはschema='line-dispatch-audit-v1', operation_id nullable, intent_id, attempt_id nullable, source_kind, source_id, purpose。jobは既存job actor表現を用い架空staffなし。必須監査と対象DB遷移は同TX、外部実行後の失敗だけはunknown保持。proof raw本文は監査へ書かない。

## 8. 実装slice・未完境界

C1：caller registry＋purpose/DM/currentauth＋immutable内部command/dedup、既存manual/alias/bulk/tag発火の迂回遮断。C2：execution/typed provider/結果照合/retry＋fake provider fixture定義。C3：臨床source・医師指定・確認保存の具体adapterとUI接続。C4：全callerの切替対応表・運用受入。C1/2をC3未完成で全配送公開せず、C3が必要なcallerだけblockedにする。

臨床source concrete wireは§3Aで選定済み・未実装。実template/scenario対応、provider能力の証拠、実環境registration/受入はなお実環境依存の未完subset。本書が選定した意味境界を後戻りさせず、存在しないprovider照会や未実装doctor指定をあると偽らない。本人業務質問は現在不要。資料照合で特定callerの意味がどうしても決まらない場合だけ、当該本文/元事象を示して確認する。料金上限や新時刻の数値を独断で追加しない。

fixture定義（実行NOT_RUN）：DM3値×6purpose、全17trigger分類、別owner file拒否、未知scenario blocked、全alias迂回0、同semantic別key、失権、binding変更、監査失敗、claim→revokedと遅延start、started後unknown、nil成功のみ拒否／外部IDなしの正式受理応答、tag不在偽absence、provider後日削除で過去成功terminal、結果TX失敗、結果なしnullCAS、再試行前source変更、ordinary同期と配送分離。実LINE・個人情報送信・費用発生の許可を含めない。
