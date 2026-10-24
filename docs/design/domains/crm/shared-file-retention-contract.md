# EMR-567 共有ファイル原本保持の限定技術契約

2026-10-24、基準main `c0ee0914e723791c59cfa971cbb1ad84cff171db`。[D15保持方針](../../../spec/line/lstep-integration.md#crm-shared-file-retention)と[C05](../../../spec/shared/retention-audit.md)を実現する技術選定。Phase 1最後の優先順位を維持する。retention_class、deleted_at、15分/1時間URL TTL、[C06/Q/ClaimMetadataWriter](../../shared/operation-participation.md)は既存供給だが、以下の専用接続は未実装。storage先行delete・経過時間だけの抽出は適合証拠ではない。

## 保存形

全FKはclinicを含めON DELETE RESTRICT、IDはopaque安定ID、版は正整数bigint。監査・anchorのcascade deleteは禁止。以下は実DDL変更ではなく限定供給契約。

| 保存先 | 追加内容・制約 |
|:---|:---|
| shared_files | version/reference_revision bigint default 1、retention_evidence_id/object_generation nullable、cleanup_state default none（none/reserved/deleting/deleted/unknown） |
| shared_file_retention_evidence | append-only。id/clinic_id/file_id/file_version_at_check、retention_class/classification_reason、original_id/version nullable、verified_by_kind/ref/at、proof_kind/ref、content_digest、version。temporaryには原本一致と用途証拠必須 |
| shared_file_references | id/clinic/file、ref_kind/id/version、requires_copy、delivery_state nullable、active、created_at/ended_at。UNIQUE(clinic,file,ref_kind,ref_id)。履歴は削除しない |
| shared_file_operation_targets | binding_id＋clinicをPK、file_id/authorization_root_id。claimで保存し論理削除後も保持。コピー準備では後述の検証metadataも保存 |
| shared_file_cleanup_jobs | id/clinic/file/file_key/object_generation、expected_file_version/reference_revision、evidence_id/version、original_id/version、pin_held/pin_released_at、state、attempt_id/provider_receipt nullable/last_error、requested_at/started_at/finished_at。state=reserved/deleting/succeeded/failed/cancelled/blocked/unknown。1file/generationに1active job |
| audit_logs再利用 | resource=shared_file、resource_id=file_id。old/newに版・論理状態、metadataにoperation binding/classification/evidence/original refs/reference_revision/reason全文。既存actor/created_atで必須同TX INSERT |

原本pinはjob内の同院shared_files FKとheld列へ統合する。汎用原本registry・別pin表・専用audit表は作らない。初版allowlistはsource_kind=shared_file_original、source_ref=別shared_files.id、source_version=その版だけ。自己参照・同objectは禁止。classやdigestだけ、外部URL、client文字列は保持証明でない。

**任意Uploadはoriginal_file_id/purpose/同じbytesが指定されてもunclassifiedのまま**。分類unclassifiedと保持評価unknownは別。新しい人間の分類画面・権限・汎用外部source frameworkは設けない。

## server所有の配送コピーproducer

固定mainの接続点は `backend/internal/lstep/line_send_service.go` の `dispatchLineMessage` pdf_url/image_url→GetSignedURLForDelivery→PushFileURL/PushImageURL。567同一実装PRでURL準備を内部 `PrepareVerifiedDeliveryCopy(ctx,AuthScope,preparationKey,sourceFileID,expectedSourceVersion,deliveryAttemptRef)` へ接続し、`{copyFileID,evidenceID,version,preparationStatus}` を返す。一般HTTP/任意purposeから呼べず、配送callerだけが別objectを作る。既存Send認可・同院検査・通知停止/目的/本人条件を保持し、新しい送信目的や権限を作らない。

sourceはbusiness_original/required_history又はcleanup対象外の保持されたunclassified実体。後者を業務原本と改称しない。temporary多段コピーは初版対象外。同院・版・不変世代・実体読取・保持fenceをserverが検証し、新copyの別object/同一bytesに加えserver配送用途と外部Push前に永続化したdeliveryAttemptRefをevidenceへ固定する。client class/proof/digestは採用せず、元fileは送信しても分類を変えない。

preparationKeyはC06登録結果の識別子、verification_idはserver採番の外部copy試行ID。準備開始時に永続化し同準備は同keyを使う。operation_targetsは元source/clinic/verification_id、段階prepared/verifying/verified/unknown/failed、予定copy key/世代、確認digestを保存。claim hookはmetadataだけを書く。

1. claimと不変anchorを保存。
2. source file lock下で現在認可・版を検査し、referencesへprovenance_verificationとdelivery_pendingをpre-pin。
3. TX外で別不変objectへcopyしbytes/世代を検証。
4. source/file/ref/世代をlock下で再照合し、C06 Applyがcopy行・temporary evidence・audit・成功receiptを同一TX commit。
5. 検証pinだけを終了し配送参照を保持、そのcopyの1時間署名URLをdispatchへ返す。登録前は一般fileとして公開せず、失敗時はPushへ進まない。

Contentはschema=shared_file.prepare_delivery_copy.v1、source_file_id、delivery_attempt_ref、intent=prepare_delivery_copy。期待source/ref版・verification_idはPreconditions。evidence.operation_binding_idとdeliveryAttemptRefにuniqueを置く。成功結果はcopy_file_id/evidence_id/committed_version/retention_class=temporary。同key成功はreceipt再利用、異内容はconflict、verified metadataだけでは業務成功でない。

コピー準備receiptは外部送信receiptではない。504の外部送信確定/unknown/再送契約を維持し、準備成功からPush再実行許可を導かない。copy/登録応答不明は同verification_id/key照合まで停止、新object・新key再登録・Push再実行なし。未登録objectはmetadataで追跡し自動削除しない。確定失敗も元分類と必要参照を保持して未配送を示し、外部不明の元file参照は解除しない。

## 504との供給・統合順（旧caller復活なし）

固定mainのdispatchLineMessageは上記の出発点の実在根拠だけである。Draft #853 head `fd291a54d0c5f76a9d247168c1fb5ce42440ffc3` は同関数を除去しSendを422 line_dispatch_pendingへ置換している。未mergeの候補headをmain採用とせず、採用後に旧関数を復活・fallbackしない。

正規接続先は[504 C契約§4/5](../../../work/emr/emr-504-c-dispatch-adapter-contract-20261021.md)のmanual pdf_url/image_url限定CreateDispatch→claim済attempt→start前内部準備。504がsource/目的/現在認可/owner-file対応、immutable file_id/version/digest、intent_id/attempt_id/execution fenceを供給し、PrepareVerifiedDeliveryCopyのdeliveryAttemptRefには**実attempt_id**を使う。コピー/evidence/参照commit後、504 start TXがfile対応・digest・現在gateを再照合してから外部開始。copy ID/URLは配送artifact mappingでありimmutable semantic payloadを差替えない。準備不明はstart不可、準備receiptと送信receiptは別。

保護DDL/manual hide/read/URL/cleanupは並行準備可能。確認済temporary成功経路にはmanual-file C1/C2 source/intent/attempt/start-fence subsetを同一実装PRで504 ownerと共同収容するか、固定先行採用SHAを受領して接続する。全504・全trigger・C3臨床・C4運用は待たない。未供給時は「保持機能実装可、実配送producer未接続」と明示し567全体を即時実装Ready/全完了としない。#853採用後の対象hunkを再照合し、根拠のない別Push経路を作らない。

## 参照fence・単一保持判定

- `ResolveRetainedOriginal(clinic,sourceKind,sourceRef,sourceVersion)` → stable original_id、retained/unknown、original_version、object_ref/generation、digest、proof_ref。消費時に現在保持を再検査。
- `AcquireRetentionPin(tx,original_id,expected_version,cleanup_job_id)` → job.pin_held。原本の削除/置換/保持解除も同原本行をlockしheld pinなら拒否。保証できない原本はcleanup対象外。
- `EvaluateFileRetention(tx,clinic,fileID,expected)` → manual_hide_allowed、cleanup_eligible、保護理由、現在file/ref/evidence/original版。manualとcleanupが共用。
- `AttachFileReference(tx,fileID,refKind,refID,refVersion,requiresCopy,deliveryState)` / `EndFileReference(...)` は業務記録・送信claim・URL要求・不明送信を保護する。外部処理より先にDB参照をcommitし、送信済みだけで履歴参照を消さない。

全参照増減・分類/evidence変更・manual hide・reservationは同file行をlockしreference_revision/versionを増加。通常参照追加はcleanup_state=reserved/deleting/unknown/deletedで拒否。reserved解除は外部未開始が確定した時だけ。既存参照が未移行/閉包不明ならunknown。カウンタだけでなく同TX current referencesを検査しrequires_copy=true/未解決送信/読取失敗はcleanup不可。

lock順は対象file＋原本fileを一つの集合としてID昇順→reference/evidence/job。C06 Gateがfile gates→binding row→Apply内の残りを取る。原本ref変更は依存競合で停止し下位fileを後取りしない。原本削除側の逆順も禁止。同内容再試行は最新preconditionで再確認できるが古い版を成功させない。

## manual hide・C06

| /v1 route | 契約 |
|:---|:---|
| GET /shared-files/:id/deletion-preview | 現在管理read/scope。file_version/reference_revision/evidence_version/original_version、retention_class、manual_effect=hide_only、cleanup_eligible/reasons |
| POST /shared-files/:id/hide | operation_key UUID、reason必須、expected file/ref/evidence/original版。operation status/receipt refs/committed version/logical_state/physical_state |
| GET /shared-file-operations/:operation_key | 元target anchorの現在認可後C06 Lookup。成功なしbindingも解決し、コピー準備pending/unknownと登録成功も区別 |

旧DELETE callerはhide intentへ置換しstorage.Deleteを呼ばない。list/新署名URLは現在deleted_atでhiddenへの新URLを拒否。manual hideは現在hospital-settings:deleteを再検査しuploadから管理権限へ昇格させない。

Scopeはserver clinic、owner_kind=shared-file、認証staff、intent=shared_file.hide。Contentはschema=shared_file.hide.v1、file_id、effect=hide_only、command_sha256。VerifiedCommandに固定schema `{file_id,effect,reason}` のUTF-8内容を保持し、同じ正規化のprojectorがSHA-256 hexへ写像する。reason全文は短Content mapの256byteへ入れずQからauditへ保存。trim後非空を検証し未知key拒否、新しい理由長の業務制限は作らず既存transport上限で超過を切捨てず拒否。Preconditionsは期待各版・現grant fenceでdigestへ混ぜず、stale後に同key同内容＋最新preconditionを渡せる。

ClaimMetadataWriterはbinding・元file/clinic/認可rootだけ。Applyはgate/binding保持下で現在認可/scope/版/参照/原本proofを検査し、deleted_at・file.version増加・必須audit・receiptを同一DB TX保存。監査失敗・競合は全rollbackしobject/file_key/参照/evidence保持。receiptはfile_id/audit_id/logical_state=hidden/physical_effect=none/committed file_version/reference_revision/日時。replay/lookupも元target現在認可、異内容operation-key-conflict、不明はLookup、DB/object不在だけで成功にしない。

原本不明のmanual hideは許可し、original_version不存在をproof不足として拒否しない。NULLから別値への変化は競合として検出。original/history/unclassifiedの自動cleanupはdeleted_atも変更しない。

## 物理cleanupはDB-only Apply外

確認済みtemporaryだけに既存143時間を適用しURL TTLと分ける。hide成功は物理削除許可/成功でなく、D15 jobの別結果。storage.DeleteをC06 Applyへ入れない。

1. file lock下でclass/proof/別原本保持/参照/143hを再検査し、適格時だけjob＋原本pin＋reserved＋必須audit/reservation receiptを同TX保存。不適格zero-write。
2. Worker直前TXで全fence再照合。変更なら外部未開始のままcancelled/blocked、file.none＋version増加・pin解除、既存manual deleted_at保持。適格ならdeleting＋attempt_idをcommitし参照/分類変更を拒否、pin保持。
3. TX外でexact file_key＋object_generationを削除。adapterは条件付き削除又は不変key保証を提供し上書き/再利用禁止。保証できない旧objectはunknown。
4. provider確定証拠とattempt_idでjob.succeeded/file.deleted・論理非表示、必要履歴/audit/元参照を保存後pin解除。確定未成立失敗はjob.failed/file.none＋version増加・pin解除。timeoutはjob/file.unknown＋version増加・pin保持。hide receiptを物理成功へ書換えない。
5. 同job/attempt/generationのprovider receipt等で照合。DB/object欠落だけで成功にせず、unknownの自動再delete/key変更/再送なし。外部未実行/確定失敗が証明できた時だけ同job再判定。それまでpin/fence保持。

発行済URLは15分/1時間内で利用可能な場合があり即時失効を保証しない。旧storage先行deleteの片側欠落はpartial/unknownで元参照と原本照合を保持し復旧済みとしない。scheduler確認・停止・実行を本docs作業で行わない。

## 同一実装PRと受入

567 PRへ4表/追加列、元原本adapter、PrepareVerifiedDeliveryCopy、参照caller、Q adapter、preview/hide/lookup・UI conflict/recovery、list/URL、cleanup reservation/worker/reconciliationを共同収容。446/466は限定DDL/model/生成供給、489は既存Q/Runner subset。567が保持方針と全consumer接続を所有し489全体Done待ちにしない。source adapter未供給なら保護に留まるsource_kindを明示しtemporaryを捏造しない。hideだけ・永久cleanup停止だけで完了としない。

Upload/Createはunclassified、FindAll/FindByIDは版付きprojection、GetSignedURL/ForDeliveryは現在状態と配送保護参照、DeleteはC06 hide、CleanupExpired/FindExpiredは適格候補→jobへ置換。closed契約と同PR責務が設計Ready、実装SHA・caller接続・動作PASSは別。技術実装不足を新しい業務承認待ちにしない。

fixtureは任意upload保護、original/history/unclassified手動hideの実体保持と自動除外の論理状態保持、証拠ありtemporary143h、ref/proof競合、失権、監査rollback、commit後不明・同key照会、同key異内容、外部timeout/片側欠落、URL TTLとobject保持の独立。全て定義のみ。製品/API YAML/DDL/生成・試験/build/DB/Docker/STG/cleanupはNOT_RUN。
