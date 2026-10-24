# EMR-572 空誤作成カルテの非表示・同ID復元

基準main `4b7ea65a3b692073bb41c4be52bd2c1a9be99d74`。親452、[実行票572](https://app.plane.so/baritechllc/browse/EMR-572/)、個人割当なし。[FR-medical-records-empty-record-hide-restore・4AC](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-empty-record-hide-restore)だけの選定技術契約であり、まだmain採用/実装/受入済みでない。canonical DB再構築前提で旧復元・汎用DELETE互換・予約取消cleanupなし。

根拠は[C06](../../shared/operation-participation.md)§§1.1–1.3と[495契約](../../../work/emr/emr-495-consultation-correction-contract-20261021.md)。固定main medical_record_crud.go:275–327の旧status/estimateだけのDelete・best-effort actor=nil監査を新操作へ流用しない。旧draft/finalizedを復活させない。ResourceMedicalRecordsのwireはmedical-records。

## 公開wire・現在認可

新routeは `/api/v1/medical-records/{id}/hide-empty` POST、`/{id}/restore-hidden` POST、`/{id}/visibility` GET、`/{id}/visibility-operations/{key}?intent=medical_record.hide_empty|medical_record.restore_hidden` GET。全routeは認証clinic・現在staff所属・ResourceMedicalRecords view+delete、対象clinic/pet・現在患者閲覧scopeを検査。復元にcreate/新医師確認/新権限を追加しない。unauthorized/missingは既存非開示404、grant不足403で存在詳細非公開。

POST Idempotency-Keyはcanonical lowercase nonzero UUID、body operation_key禁止。restore exact keys `{reason,expected_record_version,expected_visibility_event_id}`、hide exact `{expected_record_version,expected_visibility_event_id,edit_fence}`。fenceはnull又は `{edit_session_ref,edit_epoch}`。hide理由入力を新設しない。restore reasonはtrim後1..1000 Unicode scalarでtrim済み保存。IDは正規正数十進string、version/epoch正整数、expected eventはnull又は正規ID。未知/重複key拒否はJSON bodyとnested edit_fence objectの未知・重複memberを指す。Idempotency-Keyの同key同Content replayは許可し、同key異ContentはC06のoperation-key-conflictとする。

hideはvisibleでhead null又はrestore、restoreはhiddenでhead hideかつそのhide event ID必須。restoreはedit-right行で直列化しheldなら409 edit_right_busy、解放/奪取なし。hideは同行lockしfreeならfence=null、heldなら呼出端末の現在session/epochを照合し他端末/他者拒否。本人heldは同TXで既存free releaseとeventを記録しepoch後退なし。hidden sessionを捏造せず、可視性操作のためedit grant取得を必須にしない。UI未保存/未解決writeは解決前submit不可。

visibility GETは単一soft-deleted対象を限定readし `{record_id,record_version,visibility:visible|hidden,visibility_event_id,hidden_at,can_hide,can_restore}`。can_*は時点判定でwrite tokenでない。既存患者指定一覧GET `/api/v1/medical-records?visibility=hidden` はpet_idとview+delete必須、既存page/limit上限で最新hide eventのある対象だけ。既定visible、all modeなし、旧一般deletedは候補にしない。最小DTOと復元理由確認だけを表示しhidden本文API/汎用PATCHを開放しない。

POST/lookupは既存C06 flat ResultResponse、成功result_refs `{medical_record_id,visibility_event_id}`、committed_versions `{record_version}`。現在状態は別GETで取り過去receiptと混同しない。validation400、version-conflict/dependency-changed/edit-right-lost409、busy409、result-unavailable503。不明は同key lookup、別key補償なし。

## 不変command・anchor

owner_kind=medical_record、intent=medical_record.hide_empty|medical_record.restore_hidden。command `{schema_version:"1",intent,record_id,reason,target_hide_event_id}`。hideはtarget=null、server固定reason=empty_record_miscreation、restoreは指定hide IDと入力理由。JCS/SHA256は495同規約。Content exact `{schema_version,command_sha256,target_kind:"record",target_ref,intent}`、CAS record版/expected event/hide session-epochはPreconditions。Applyはimmutable command再hash照合後write。

新medical_record_visibility_operation_targetsはclinic_id/binding_id/medical_record_id/pet_id bigint NOT NULL、record/pet CHECK>0、PK(clinic,binding)。唯一FK(binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT。新bindingと同短claim TXで一度INSERT、UPDATE/DELETE禁止。business FKなしで臨床rowlockをclaimへ持込まない。元targetは認可済server read由来。lookup/replayは保存anchorだけで元record/petを再認可しincoming ID代用なし。metadata欠落/対象不明は非開示、URL不一致は元対象認可後409。EffectsDatabaseOnly、外部通知/予約/tag/会計/在庫write、汎用Delete/cleanup呼出なし。

## typed EmptyRecordGuardの閉包

identity/date/owner/pet/doctor/appointment/record_no、作成時刻/版、edit-right/operation/auditだけで記入済みにしない。ただしscope不整合/未取得は空でなく、入力履歴を消して空へ再分類しない。

| 検査対象 | 空として許容する条件／拒否境界 |
|:---|:---|
| inquiries | 行なし、又は全業務string厳密空＋chief_complaint_type_id/appetite/water_intake/staff_id NULL＋実入力revision/履歴なし。空白を空へ正規化しない |
| clinical_plans | 行なし、又はphysical_exam/diagnosis_details/treatment_policy厳密空、診断行/実入力revision履歴なし。旧単一診断列なし |
| record臨床参照 | recommendation_reason/conversation_summary一式/revisit_intake_target_id NULL。初回問診revision/再診回答/治療・plan item/vital/検査/健診/画像・進行upload/接種/処方・予防/addenda/訂正proposal・link/終了event/実来院根拠/医師確認成立履歴が一件でも拒否。billing_confirmation初期shellだけpending/defaultかつ未確認未記入なら可 |
| 業務依存 | 見積/請求/billing source_record/doctor_review participant/他予約source_medical_record/検査save stage/取込・帳票等の保存参照は一件でも拒否。論理削除/取消履歴も無視しない。親appointmentの正当link自体は許容するがin_consultation拒否、appointmentを書換えない |
| staff_record_drafts | 他者/未解決本文を消さず、保存draft本文/復元待ち/未解決operationは拒否して別解決操作へ。edit-right/audit/成功receipt/visibility eventは業務入力でない |

同PRで実001/採用496/516等の正本readerへ対応付け、旧列や空mockを代用しない。正規新DBだけを対象とし未知namespace/不整合/reader失敗はdependency-changed又は503、writeゼロ。全owner Done待ちでなく技術adapter閉包。

restoreも同じ空・依存・現在scopeとrecord_no/active appointmentの既存uniqueを検査。同日別recordだけで拒否せず新date/pet uniqueなし。同row deleted_at=NULLへ戻し元ID/日付/owner/pet不変。現在主飼主変更で元ownerを再帰属せず、現在pet.ownerと過去record.owner等値を新条件にしない。保存帰属と現在閲覧/操作scopeを分離。

並行順はC06 claim→appointmentあり時だけ既存予約graph Gate→binding NOWAIT→savepoint→appointment lifecycle lock→494 edit-right row→record row→依存owner既定順。全依存writerは親rowlock下で現在可視性を再検査、不足callerだけ同PR参加。hide前開始の未lock writerがhide後子commitする穴を残さず切替前inventory/fixtureで閉じる。成功replayは新Apply空判定より前に現在認可して過去receiptを返す。

## append-only保存と同一TX

既存medical_records.deleted_at/versionを更新し物理DELETE禁止、hidden汎用write拒否は親guard一元化。新medical_record_visibility_events：id bigserial PK、clinic_id/medical_record_id bigint NOT NULL、revision bigint NOT NULL CHECK>0、event_kind text hidden|restored NOT NULL、previous_event_id/restores_event_id nullable bigint、reason text NOT NULL CHECK btrim(reason)<>''、recorded_by bigint NOT NULL→staffs(id) RESTRICT、recorded_at timestamptz NOT NULL DEFAULT clock_timestamp()、operation_id bigint NOT NULL、record_version_before/after integer NOT NULL CHECK after=before+1。

UNIQUE(id, clinic_id, medical_record_id)、UNIQUE(clinic_id, medical_record_id, revision)、UNIQUE(clinic_id, operation_id)。FK(medical_record_id, clinic_id) → medical_records(id, clinic_id) RESTRICT、FK(previous_event_id, clinic_id, medical_record_id) → medical_record_visibility_events(id, clinic_id, medical_record_id) RESTRICT、FK(restores_event_id, clinic_id, medical_record_id) → medical_record_visibility_events(id, clinic_id, medical_record_id) RESTRICT、FK(operation_id, clinic_id) → operation_bindings(id, clinic_id) RESTRICT。operation_id NOT NULLは維持する。event_kind=hiddenはrestores_event_id NULL、event_kind=restoredはrestores_event_id NOT NULL。head chain/alternationとrestores_event_id=head.idをrecord lockで検査し、初回revision=1、以後revision=head.revision+1。UPDATE/DELETE禁止・既存tenant RLS、履歴trimなし。headは限定recordの最大revision、別head表/cache/statusなし。

最外TXでrecord CAS/deleted_at、hide必要release/event、visibility event、認証actor必須audit、成功receiptを保存。commit前失敗全rollback、commit後応答喪失/再読取失敗はrollbackを約束せず現在認可付き同key lookup、別key/自動補償なし。auditは本文でなくID/理由/可視性/前後版。DDL/model/ERD/索引は446/466共有hunkを単一editorへ収容。

## 一つの実装PRと受入

schema/model→guard/anchor/participant/handlers→route/OpenAPI/DTO→専用FE query/mutation→既存MedicalRecordFormActionsと患者指定hidden一覧/確認/lookupを同feature PRで閉じる。494/489/495必要subsetだけ同PR又は採用SHA、producer無しUI/mock完成不可。通常編集/終了/訂正/汎用DELETEへ混ぜない。既存接点routes_records.go/medical_record_crud.go/medical_record_repository.goとmodel/001/ERD、visibility専用service/repository/request/response/participant、公開operation helper、FE既存一覧/API。汎用Delete実体流用禁止、共有編集は個人名でなく単一窓口。

4ACのfixture（NOT_RUN）：pristine→hide→同IDrestore、入力済/clear後/全依存拒否、scope/失権/理由なし、他session/hidden stale write、child insert両順競合、active appointment unique衝突/別同日非衝突、同key replay/異Content/CASだけ更新、応答断、anchorA/URL B非開示、audit/receipt/commit前response reload失敗rollback、commit後不明はlookup/補償なし、過去hide receiptと現在restored表示分離、通常一覧除外、read/cancel zero-write、復元で開始/受付/課金/確定/通知なし。製品/試験/DB/生成/実データ操作は未実施。
