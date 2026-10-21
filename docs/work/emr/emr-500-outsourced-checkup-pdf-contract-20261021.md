# 外注健診PDF：採択技術契約 R4

2026-10-21。採択済み技術契約、runtime未実装・試験NOT_RUN。本人が物理方式を直接指定した意味ではない。採択済みD09業務要件（#790→#784統合）を実装可能なprovider非依存sliceへ具体化する。ここに書くroute/table/portは採択した技術契約であり現mainの実在を意味しない。外部送信・provider・課金・本物PDF処理は未承認のまま。

## 0. 調査基準と既存資源

基準main e71ddf03462a0ee7a9864ad2ce6854b243591770（06:00 UTC取得）。

- `backend/internal/model/storage_clinical.go`：CheckupResultRevision（package_version、package_snapshot、fields_snapshot、restored_from_id、reason、operation_id、actor/time）、CheckupSaveAttempt/StageResultが実在。PDF原本、抽出候補、runの保存はない。
- `backend/internal/model/checkup_record.go`：checkups.versionが実在。field_resultsは型付き値とsnapshotを持つが原本/抽出版を持たない。
- `backend/internal/medicalrecord/routes_records.go`：record配下GETはmedical-records:view、checkup createはcreate、更新/結果PUTはedit。global checkups:viewをrecord配下の新権限へ置換しない。lab-import権限を転用しない。
- `medical_record_image_request.go`：PDFを含む既存添付は10MiB、request11MiB、拡張子/MIME/sniff照合。画像添付としてPDF受付はあるが、checkup identity・不変原本・抽出runとは別契約。
- `infra/uploader.go`：Upload/Delete/GetSignedURLのみ。immutable put、read/stat/digest照合はない。LocalUploader.GetSignedURLはTTL無視のURLであり、新原本の認可読取には使えない。S3署名URLも発行後の即時失権を保証しない。
- `medical_record_image_service.go`：元画像は削除経路を持ち、exam_id対応。健診原本をこの可削除imageレコードへ紐づけただけでは原本保持にならない。
- `lab_import_service.go`/lab_import_jobs：D07検査import専用の状態・患者mapping。汎用OCR queueではない。既存labjobをCheckupへ別名流用しない。
- `checkup_field_result_service.go`：削除/挿入/削除監査は同TXだが、候補/原本/C06/revision追加を含む同TXではない。定義取得がTX前、旧finalized一律拒否、nil number→normal導出もある。現関数をそのまま呼んで契約充足とはしない。
- `checkup_service.go`：Create/Updateの後処理にLINE followup/tag同期が存在。PDF登録/抽出/確定からこれらの汎用Create/Updateを呼んで副作用を発生させない。

リンクは末尾。以下はこれらの差分を隠さず追加する最小案。

## 1. identity と対象

最初のsliceは、通常受付または既決過去入力を経て既に保存されたCheckupを対象にする。未保存の親/子をPDF登録が自動生成しない。EMR-500の親→子→結果の段階receiptは維持し、子保存成功IDからPDFへ進む。原本登録だけを健診実施・完了と扱わない。

不変対象tupleは clinic_id / medical_record_id / checkup_id / pet_id。医院は認証scopeから取得し、URLのrecord/checkupとDBの親・患者を照合。checkup.pet_id NULLは同親のpet_idで解決するが、両方なし/矛盾は登録拒否。名称一致やPDF記載だけではpatientを選ばない。別健診への付替えAPIは初回なし。将来訂正に原本保持・経緯が必要であり単純UPDATEで対象を書き換えない。

原本document_id→抽出run_id（複数可）→immutable candidate→CheckupResultRevision（職員の明示確定）の一方向系譜。複数PDFを同健診へ登録できるが、各確定はその時点の結果全体snapshotをCASで保存し、別PDFの確定結果を知らずに上書きしない。自動mergeなし。

## 2. API選定案

実ルータprefixは `/api/v1`。下記はapi.yaml相対path。`B=/medical-records/{record_id}/checkups/{checkup_id}`。BIGINT ID/版は正規形十進文字列、UUIDは標準文字列。新しい認可resourceは作らない。

| 操作 | route | 既存grant | 内容 |
|---|---|---|---|
| 原本予約 | POST B/pdf-documents | medical-records:create | C06、原本metadata/期待対象版を登録しdocument_idを返す |
| 原本列挙 | GET B/pdf-documents | medical-records:view | reserve/readyとrun状況、登録actor/time、元対象・digest。cursor paging |
| 原本内容転送 | PUT B/pdf-documents/{document_id}/content | medical-records:create | application/pdfのbinary、予約済みdigest/長さと一致のみ。DB TX中にblob転送しない |
| 原本登録完了 | POST B/pdf-documents/{document_id}/register | medical-records:create | C06、保存blobを照合しreadyへ。同TX監査/receipt |
| 原本内容読取 | GET B/pdf-documents/{document_id}/content | medical-records:view | 現在認可の後にserver stream、application/pdf/nosniff/private,no-store。公開/署名URLを返さない |
| 抽出開始/明示再抽出 | POST B/pdf-documents/{document_id}/extractions | medical-records:edit | C06、queue登録までを操作成功。抽出完了とは別 |
| run列挙/単体 | GET B/pdf-documents/{document_id}/extractions[/{run_id}] | medical-records:view | 進捗、候補、出典、failure。現scopeのみ |
| 確認・修正して確定 | POST B/pdf-documents/{document_id}/extractions/{run_id}/confirmations | medical-records:edit | C06、同じCheckup結果revisionへ保存 |

原本metadataGETはlist内で足りる。結果履歴はEMR-500の同じ結果revision readerへsourceリンクを追加し、別の確定値正本APIを作らない。C06結果照会は新限定consumer GET B/pdf-operations/{operation_key}?intent={intent} とする。owner_kind=checkup_pdf、intentは checkup_pdf.reserve / checkup_pdf.register / checkup_pdf.extract / checkup_pdf.resume / checkup_pdf.confirm の閉じた集合。actorは認証staff、clinicは認証scope。対象IDをintent namespaceへ混ぜない。元binding/targetsの対象を現在認可した後、URLtuple不一致は404。未claimはscope-not-foundで、別actorのkey探索を許さない。

GET lookupは元actorの現在medical-records:view、またはそのintentの元grant（reserve/register=create、それ以外edit）のどちらかで、ID/版/時点だけの既存receiptを返せる限定authorizerとする。create-only actorにも自分の登録成功の回復を閉じるが、原本bytes/候補本文/document listへのview権限は増やさない。reloadでkey自体を失った場合用にGET B/pdf-operations?intent={intent}も同じ元actor/現在grantで提供し、元対象のoperation_key/status/結果ID/版/時点だけを列挙する。順序はclaim時刻とkeyの降順、上限50、opaque cursorはページ境界だけで認可証明ではない。元対象を毎回認可し、一覧不在を未実行証明とせず新key自動作成を禁止する。POSTは元grantを維持、view-only actorはGET lookupで回復する。claim winnerのみreserve document UUIDをanchorに固定、同key/race loserは元UUIDを再利用。新発番UUIDをrequest Contentへ含めて再送時に内容を変えない。新UUIDはserver result identityであり元内容はtupleとmetadata。

VerifiedCommand / ClaimMetadataWriter / ContextGateの採択済み必要subsetを利用するが、e71ddf03の現mainに全hookがあると扱わない。新targetsとclaimの同TX保証の実装が公開条件。

### 入力schema

reserve: `operation_key, expected_checkup_version, original_filename, byte_size, sha256`。sha256は64小文字hex、bytes>0かつ既存添付と同じ10MiB上限。multipart envelopeを使う場合11MiB。filenameは表示用でpathに使わない。formatはapplication/pdf固定。source tupleはDBから束縛し、callerからclinicを受けない。

register: `operation_key, expected_document_version`。blobstoreの不変object+実測digest/size/MIMEと予約が一致した場合のみready。予約後の親変更/失権を再検査。内容が未到達なら既存C06のdependency-changed分類、成功receiptは作らない。

extract: `operation_key, expected_document_version, expected_latest_run_id`（初回null）。再抽出は既存terminal runを認識した明示操作。provider/profile名はserverの採択済み設定から解決し、callerは任意provider/URL/credential/プロンプトを渡せない。adapter無効は503、claim/queueや外部送信なし。queue登録成功receiptはrun_idのみ返す。

confirm: `operation_key, expected_checkup_version, expected_result_revision, expected_package_version, expected_field_versions, candidate_digest, patient_match_confirmed:true, reviewed_fields, removed_entry_keys, review_reason`。結果revision0は既存結果revisionなしを示すCASsentinelでDB行0を作らない。reviewed_fieldsは全結果snapshotのtyped entries。既存値を残すか変更するか画面で比較し、元結果版を照合する。初回reasonは任意、既存確定値を変える訂正は理由必須。理由に未選定の文字数上限を発明しない。原本/抽出版/候補digestはURLとrunから固定、bodyから別患者/別runを注入不可。

reviewed_fieldsの唯一の入力定義は§2aのclosed unionとする。元表記・0・false・空文字・空配列と未評価を区別し、黙った単位換算やdecimal丸めをしない。decimal(10,4)適合不能はC06 validation（400）、原候補は保持する。未知項目をmasterへ自動登録しない。


### 応答とエラー

全DBcommandはC06既採択flat12statusと現ResultResponse（result_refs / committed_versions等の既存フィールド）を再利用する。新たなreceipt envelopeを発明しない。result_refsはmap[string]stringなので不存在キー省略、nullを入れない。reserveはdocument_id、registerはdocument_id、extractはdocument_id/run_id、confirmはcheckup_id/result_revision_id/document_id/run_id。result_refsのresumeはdocument_id/run_id。committed_versionsの有限キーはreserve/register=document_version、extract/resume=run_version、confirm=checkup_version/result_revision/package_version。extractのdocumentは変更しないためdocument_versionを結果版へ入れない。confirmでdependency版を変更するowner採択がある場合だけmedical_record_dependency_revisionを追加する。この追加の有無はwriter契約で固定し任意mapを返さない。無作用成功は既存対応版を返す。result_revision_idは永続行ID、result_revisionは結果の連番で混同しない。常にoperation成功とextraction成功と職員確定を区別。

400 malformed、401未認証、403権限不足、404対象不一致/不可視、409 CAS・active run・not-ready、413上限、415非PDF、429共通quota、503provider/store未配線。結果不明はC06照会し、5xxやtimeoutから新keyへ自動変更しない。Retry-Afterは存在する場合のみ尊重し安全な同一操作の再送可能性とは分離する。ContentとPreconditionsを分離する。Contentにはschema-tag付きJCS投影 `{schema:"checkup-pdf-command/v1",intent,record_id,checkup_id,document_id,run_id,profile_id,profile_version,original_filename,byte_size,sha256,candidate_digest,reviewed_fields,removed_entry_keys,review_reason,patient_match_confirmed}` の操作別実在キーだけを含め、SHA256 digestをNewContentのcanonical文字列キーcommand_digestへ入れる。対象IDもContentへ明示する。profile_id/profile_versionの2キーはresumeでのみ、対象runに不変束縛済みのserver値をContentへ含める。extractでは呼出時の最新設定をfingerprintへ混入せず、claim winnerが採用profileをtarget metadataへ固定しrunへ同値を渡す。再送/race loserは元targetの採用profileを再利用し、最新設定へ差し替えない。multipart bytesは実測sha256+sizeでcommitmentとする。ApplyはContent由来の不変typed commandから同投影を再計算し照合し、別captured mutable payloadで実行しない。

expected_checkup_version / expected_result_revision / expected_document_version / expected_latest_run_id / expected_package_version / expected_field_versions、および現在grant fenceはPreconditionsのみでfingerprintへ入れない。ただし保持する旧snapshotの出典revisionや実際に採用するfield snapshot identityは業務Contentであり、単なる期待CAS版とは別キーにする。同内容・同keyのCAS差替え再実行は可能、既成功は現在CASより先に元receiptを返す。値・対象・候補・採用定義や出典が変わる場合は内容変更であり新operation_key。expected_latest_run_idは競合検出だけのPreconditionsとし、変更だけで内容を変えない。同keyが既にrunを作成済みならそのrunを返し再抽出しない。明示的なもう一度の再抽出は新operation_keyで、現在latest terminalをCASする。receiptへ任意details/reason一覧を追加しない。必要な現在状態は現在view認可付きdocument/run/result readerから再取得する。EffectSummaryは内部receipt限定でwireへ出さない。

このconsumerの有限HTTP対応は succeeded/replayed/not-succeeded=200、operation-in-progress=202、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、result-unavailable=503、scope-not-found=404、not-authorized=403、validation=400。共通12statusは追加しない。現ResultResponseのstatus/operation_key/result_refs/committed_versions/recorded_at以外は追加しない。pre-C06の不正JSON・UUID・媒体・サイズ・quota・未配線は既存apperrors/HTTP error形式で400/401/403/404/413/415/429/503、operation receiptが成立したようには返さない。


### 2a. 全結果entryのclosed union

reviewed_fieldsは次の2branchに閉じ、未知キー・混在を拒否する。
- retain: {kind:"retain", retained_from_result_revision_id, retained_entry_key}。serverが元値/定義/評価/由来を不変revisionからコピーし、callerは値を送らない。元revisionがexpected current headと一致しない場合は拒否。
- review: {kind:"review", entry_key, field_ref, adopted_field_snapshot, retained_from_result_revision_id, retained_entry_key, value_state, value, provenance, review_note}。provenanceはcurrent_run_candidate（今回run内candidate IDs）またはmanual_review（職員が原本を確認した修正）。旧snapshotに対する修正はretained_from_result_revision_id / retained_entry_keyを共に非nullで保持し、元entryの型・単位・定義snapshotへ検証する。新規entryでは両方null、一方のみnullは禁止。field_refは現在master対応が存在する場合だけそのID、削除済み旧entryではnull可。adopted_field_snapshotは新規では採用masterの不変snapshot identity、旧entry修正ではその元snapshot identityと一致させる。同名新masterへ付替えない。元entry keyは安定key、legacyにない場合は不変revision ID＋元fields_snapshot全体JCS digest＋その元配列ordinalで識別し、現在配列位置で推定しない。

removed_entry_keysは必須配列（削除なしは[]）で、期待current結果版内の固定entry refを持つ。重複/未知entryを拒否。全current entry集合に対しretain/review/明示removed_entry_keysのちょうど1分岐を要求し、未送信・未取得entryを削除としない。明示削除は既存500結果消去の認可/監査に従う。pdf_sources単一runは今回確認操作の原本であり、各entryはcurrent_run_candidate / retained_result_revision / manual_reviewの固定由来を保持。後続manual PUT/copy/復元も元由来とassessmentを保持する。

正式6型のpresent値はnumber=decimal文字列、text=文字列、boolean=bool、single_select=options.value文字列、multi_select/checklist=options.value文字列配列。not_evaluatedではvalue=nullのみ。異型余剰value拒否。multi_select/checklistは既存service同様options包含を検査し、順序と重複は入力のまま保持して既存で未禁止の値を新業務規則で拒否しない。schema canonicalizationでも配列sort/dedupeしない。

## 3. 原本blobとDBの境界

既存FileUploaderを直接完成品扱いせず、新port `CheckupPdfBlobStore` を追加：`PutIfAbsentVerified(key, reader, expectedSha256, expectedSize)`, `Inspect(key)`, `OpenVerified(key, digest)`。keyはserver生成opaqueでclinic/documentを束縛。既存objectの再PUTは同digest/sizeだけ同一成功、違えばconflict。providerへのURLfetchは禁止。実データの外部保存先採用は別承認・設定で、未配線はfail closed。

1 reserveのDBcommandで不変metadata/原scope/actorとreceiptを作る。
2 PUT前に現在create権限/対象を検査、共通quotaをbody読取前に適用。10MiBでstream limit、MIME/ext/sniffの既存helper再利用。内容digest実測。blobwriteはTX外。不確実ならInspectで確認し、同じdocument/keyへ再開する。JSONURL登録口なし。
3 registerはInspect証跡を得てから短いTXへ入り、immutable object保証を前提に現在scopeとexpected_document_version照合→ready/監査/receiptを同TX。ready後の置換禁止。失敗時に既保存blobを自動deleteしない。
4 登録未完もlistで見えるためreloadでdocument_id回収できる。別keyのduplicate予防として、同checkup・SHA・sizeの一意制約を置き、同じ内容再アップロードは既存documentを返す。医院間のhash存在を漏らさない。元ファイル名が違っても同一原本、後続runは追加可能。

保持・削除の新運用は選定しない。ready原本はこのsliceから削除不可、未完blobの自動期限purgeもなし。保持容量監視は実行未承認。暗号化/破損PDFは原本登録後も抽出blocked/failedとして保持できるが、成功候補にしない。PDF parser/マルウェア対策の具体実装・dependencyは採用前確認、抽出workerは外部network/tool実行/埋込script/添付実行を許さない。

## 4. 最小DB差分（新規提案）

既存checkups / checkup_field_results / checkup_result_revisions / checkup_save_attempts / operation_bindings・receipts・auditを維持。generic images/lab_jobsの意味は変えない。

1 `checkup_pdf_documents`：id UUID PK、clinic_id/medical_record_id/checkup_id/pet_id、original_filename、object_key UNIQUE、sha256 bytea32、byte_size>0、mime固定、state reserved/ready、version>0、created_by/at、registered_at nullable、reserve_operation_id、register_operation_id nullable。UK(clinic_id,checkup_id,sha256,byte_size)、UK(id,clinic_id,checkup_id)。tupleの書換禁止。readyではregistered_at/register_operation_id NOT NULLのCHECK。document_versionはこの表の状態変更のみ1増分。
2 `checkup_pdf_extractions`：id UUID PK、clinic_id/checkup_id/document_id、run_sequence>0、state queued/running/candidate_ready/failed/unknown/blocked、version、profile_id/profile_version、engine_version(nullable未開始)、requester_id、enqueue_operation_id、lease_token/lease_expires_at、dispatch_started_at nullable、provider_request_ref(nullable private)、started_at/finished_at、safe_error_code、candidate_schema_version、candidate_digest、candidate_json nullable、completed_at。UK(document_id,run_sequence)、UK(id,clinic_id,checkup_id)、UK(id,document_id,clinic_id,checkup_id)、partialunique(document_id) WHERE state IN(queued,running,unknown,blocked)。候補完成後candidate不変。状態CHECKはcandidate_readyでJSON/digest/完了時刻必須、それ以外で候補を成功扱い不可。blockedは権限等でdispatch不能、unknownは外部結果未確定。復旧なしで新runを通すためにfailedへ勝手に変更しない。
3 `checkup_result_pdf_sources`：result_revision_id PKかつ同clinic/checkupへのFK、(run_id,document_id,clinic_id,checkup_id)の単一複合FK→extractionsの同一UK、candidate_digest、reviewed_fields_provenance JSONB、confirmed_by/at、patient_match_confirmed true、operation_id。最初のsliceは一確定command一run。確認済み全結果snapshotは既存revisionが正本、この表へ結果値を二重保存しない。元候補値はimmutable runに残るので修正前後を比較できる。未来の多原本同時確定は別設計。
4 `checkup_pdf_operation_targets`：operation_id PK、clinic/record/checkup/pet、kind、document_id/run_id nullable、adopted_profile_id/adopted_profile_version nullable（extractのclaim winnerで共に固定）、不変command_digest。C06 claimと同TXで永続化し、失敗bindingの元対象認可にも使う。reserveはclaim時にdocumentが未作成なので新ID予約と不変元tupleで保持する。成功receiptがある場合は現対象版を見て拒否するより先に元targetと§2の限定lookup authorizer（元actorのcurrent view又は元grant、receipt ID/版のみ）を検査して原成功を返す。

runの(document_id,clinic_id,checkup_id)はdocumentsの(id,clinic_id,checkup_id)へ複合FK。sourceのrunとdocumentを独立FKだけで検査して同健診別原本を混同しない。candidate_digestは当該run保存値と同TXで検証する。全FKはclinicを含む複合参照。既存checkupsには(id,clinic_id)UKあり、record/petの親整合はlock下検査＋新tuple変更防止で保証。DB強制用に必要な複合UK/FKを追加し、元checkupのparent/pet再割当writerを閉じる。新表RLSは現在clinic、workerはjobに束縛されたclinic単位、跨院無制限bypassrole新設なし。原本/候補/確定sourceはON DELETE RESTRICTで履歴保全。旧checkup削除は新依存有なら履歴付き削除方針に従い物理cascade不可。

version/history/receiptは既存trigger責務を照合し、手動version増分との二重更新をしない。model/ERD生成は担当ownerの正規生成、手書き生成物修正禁止。

## 5. 抽出adapterとjob

`CheckupPdfExtractor`入力はimmutable原本bytes/digest、run_id、profile_id/version。出力はversioned `checkup-pdf-candidate/v1`：pages[{page,text}], entries[{candidate_id,page,raw_label,raw_value,raw_unit,raw_reference,proposed_field_id?,proposed_field_version?,parse_state,value?}], patient_labels（原表記・未検証）、warnings。parse_state=parsed/unreadable/unmapped。candidate_idはrun内安定ID。numberはdecimal文字列、欠損はnullと明示state。engine/version/digest・入力documentをserverが付け、providerの任意IDを信用しない。PDFの命令文を実行したり別情報を探索しない。

候補JSON/textはDB保存し、rawproviderレスポンスを無制限log/別サービスに送らない。candidateはdiagnosis/normalを決めない。全空抽出はfailed `empty_extraction`、部分読取はcandidate_ready＋個別unreadable/unmappedで職員確認待ち。異なるrunの候補を後着順に上書きしない。


候補digestは SHA256(UTF8(JCS({schema:"checkup-pdf-candidate/v1",document_sha256,run_id,profile_id,profile_version,engine_version,pages,entries,patient_labels,warnings}))) のlowerhex。全キー必須、未知キー拒否、nullable値は明示null。pagesはpage昇順（1始まり重複なし）、entriesは抽出順を不変保持しcandidate_idはrun内一意、各page参照は実在。patient_labelsは{raw_name:null|string,raw_identifier:null|string}、warningsは{code,page:null|positiveint}の配列で入力順保持。entryキーはcandidate_id/page/raw_label/raw_value/raw_unit/raw_reference/proposed_field_id/proposed_field_version/parse_state/valueを固定し、nullableキーは省略しない。proposed field参照は双方null又は双方非null。parsedのvalueは{type,data}でtypeは正式6型、dataは§2aのpresent対応。unreadable/unmappedのvalueはnullで原表記を残す。providerが返したrun/document/profileは信用せずサーバのjob値を使用する。JSON全体を一回の完了TXでimmutable保存、部分出力の継ぎ足しはしない。

profileはmax_pages / max_output_bytes / max_entries / max_decoded_image_pixels / walltime_ms / memory_bytesを必須の正数で持ち、未設定/0を無制限扱いせず開始拒否。具体値は採用parser/providerの受領物で未選定、disabled fixture以外のruntimeを開始できない。超過はfailedで候補成功にしない。10MiB圧縮入力からの巨大展開にも適用する。

job表自体を耐久queueとして使う新worker。既存D07jobの状態・revert・Examination生成を再利用しない。enqueueはC06TXでqueued行まで。workerは短いTXでlease+tokenを取得してcommit、対象/currentrequester grantをdispatch前に確認し、外部処理はTX外。lease失効だけで外部有料処理を再送しない。dispatch直前の短いTXでdispatch_started_atと不変provider request identityを記録する。その後のlease期限切れは未送信と推測せずunknown、providerの同一request照会/冪等保証が採択されるまで自動再呼出し不可。

結果commitはlease tokenとrun versionをCASし、同run+同candidate_digestの重複完了はnoop、異digestの二重完了拒否。処理途中の失権/対象変更は確認結果へ反映不可。すでに取得した出力はその元jobのrestricted候補証跡として保持可能だが、職員閲覧は現在認可、confirmは改めて現在edit必須。workerに臨床確定権限を付与しない。blockedのうちdispatch未実施が証明されたものは、POST B/pdf-documents/{document_id}/extractions/{run_id}/resume（medical-records:edit、C06）で再開する。入力はoperation_key/expected_run_version。元requester本人・現在grant・同対象を確認し、同run/profile/documentを維持したままblocked→queuedへ一回遷移、監査/receipt同TX。別職員へのjob譲渡は初回非対象。再開後lease取得時も現在grantを再確認する。Contentにはrun/target/profileidentityのみ、expected_run_versionはPreconditions。失権したままの再開や対象矛盾は拒否、外部結果不明をblockedへ偽装しない。

unknownはresumeを受けない。限定内部reconcilerが保存済みprovider_request_refで同一要求を照会し、確定候補/確定失敗へ収束させる。外部未実施が証明できる場合だけ同runのqueuedへ戻せる。採択providerに照会/冪等保証がなければunknownを維持して外部呼出し停止。具体providerのreconciliation contractが定まるまで実adapterをactivationしない。

provider未選定の現時点はdisabled実装＋合成fixture adapterのコード契約まで先行可。runtime実患者PDF抽出は開始しない。外部呼出しがない純ローカル抽出でもparser選定・resource上限・実処理は別実装受入であり、この稿で実行許可しない。

## 6. 確認・修正・確定と競合

確認画面は原本、immutable抽出候補、現在結果、職員編集値を区別する。候補へ職員値を上書き保存しない。編集は当初メモリdraft、未送信編集がreloadで回復すると保証しない。確定済み訂正は新revision、旧candidate/旧revision/sourceは保持。

新実行の順序は、事前currentgrant・形式検証→C06既存業務Gate（現在staff/assignment等）→binding/元target→record→checkup→package/fieldsを安定順→document→run→現在result head。500既存writerのロック順と同じ共通owner関数へ統合し、逆順writerを残して公開しない。原本blob/抽出workerはrecord等のlockを持って外部処理しない。

TX内でcurrentgrant/医院/親/患者/利用可能な現在入力scopeを再検査し、checkup/package/field/result期待版とcandidate digestを照合。親finalizedを機械的一律不可とする旧実装を復活させずD09通常編集・履歴の採択規則を適用する。

同一TXで、型検証→結果field projection置換→CheckupResultRevision追加→pdf_sources追加→必要checkup/dependency版更新（owner既定で一回）→監査before/after/sourceIDs→C06receipt。いずれか失敗なら全rollback。PDF/run作成は別stageなのでこのrollbackで原本を消さない。既存EMR-500の段階保存全体を一TX化しない。

再抽出は既存runの候補/確定を変更しない。confirm対象runが現在latestでなくても、明示選択したrun/digestを表示して確認したなら使用できる。新run完了だけを理由に期待結果版を進めない。別職員の確定やpackage改定が先行すれば409、更新値を自動mergeせず再比較。既成功samekeyは原receiptを返す。別keyで同じrun+同じ確認snapshot・同じ現headなら、現在認可と元対象を照合後、CAS不一致を返す前に既存同一確定を識別して無作用successで既revisionを返し、追加履歴を水増ししない。同一と証明できなければ通常CASで拒否。reload時はdocuments/runsと結果履歴のsourceから確定先を照会する。未発見だけで前操作未実行とは扱わず、固定した期待結果版を持たない新確定commandを自動送信しない。

旧field-results PUT/Checkup.Result更新・copy/復元も同じ結果head/CAS/履歴保護を通す必要がある。旧writerがhistoryを飛ばせる状態でPDF確定だけactivationしない。書込後のLINEタグ同期、followup、請求、在庫、証明書、Examination作成を本commandへ載せない。

## 7. 未評価と既存normal既定値の閉鎖

現実装はnil numberやtextでstatus=normalを生成するため、未確認候補を既存field_resultsへ仮保存してはならない。confirmのnot_evaluatedは既存revision.fields_snapshotのversioned補足 `value_state:not_evaluated` として保持し、正常判定用のpresent rowを生成しない。present rowにも確認済みであることと評価可否は別概念。

既存field_resultsのstatus既定値をAPIがそのまま正常として公開しないため、PDF由来snapshotに `assessment_state: evaluated | not_evaluated` を追加しreaderはこれを必ず適用する。numberかつ値・採択基準の双方が存在して新たに選定する十進比較が成立した場合だけevaluated、それ以外はnot_evaluatedで、wire評価値はnull/未評価表示。float64の既存computeCheckupNumberStatusは再利用せず、decimal(10,4)を正確な十進数として比較する。旧結果保持branchの当時assessment/statusは現在基準で再評価しない。これはOCRによる臨床判定の新設ではなく既存のunknown/未評価保護。全受け手（D09一覧/履歴、D03健診タブ、staff患者履歴）を同じsnapshot-aware readerへ接続するまでPDF結果activation不可。D20公開は追加しない。

旧manual結果の既存表示を一括再解釈するマイグレーションは含めない。PDF由来の判定metadataを既存readerが読めないならそのreaderへ新結果を公開しない。新業務status enumを他domainへ広げない。

## 8. FE / BE ownerの完結範囲

A: BE storage/blob登録/read/認可/原本一覧＋C06targets。既存500親子IDと正規生成担当と連携し、providerなしで閉じる。blobstore実環境配線・実uploadは別受入。
B: BE runqueue/workerport/candidate schema・合成adapter/failed/unknown/retry照会。providerdisabled維持。queue受付成功を実抽出成功とは報告しない。
C: EMR-500 BE結果writer統合/CAS/確認source/履歴/全旧writer閉鎖・未評価reader。A/B契約に依存。500本体と別競合実装を始めず現在ownerへ同じwriter責務として渡す。
D: FE原本登録/一覧/候補比較/編集/確定UI、D03同結果reader。固定api.yaml/mockfixtureで独立PR可、本番入口はA/Cと採択providerが揃うまで非活性。UIに未実装providerが動くような成功表示をしない。失権・patient切替で本文/draftを破棄、遅延レスポンスはtuple/run/request世代で拒否。

関連票: 500主担当、452/D03受け手、516予定identity必要subset、456/225fieldmaster再利用、446/466DDL/正規modelERD、489C06必要subset、495親保護。既存実装ownerを調整せず新並行writerを割り当てない。D07/D08/通知504へ移管しない。

## 9. 残件・受入（実行しない）

業務の追加質問は現時点なし。技術的な残件は、実PDF見本未受領による帳票mapping、private不変blob adapterの配線、parser/providerと費用・外部患者情報転送の明示承認、既存500全writer/readerへの接続。見本共有予定を本人の約束と記載しない。新料金/資格/外部サービスをこの稿から採用しない。

合成fixtureで記述する受入: 同bytes再PUT/異bytes拒否、blob成功DB断/reload、別患者遅延応答、予約権喪失後register、workerlease期限と外部不明、再抽出後旧確定不変、空抽出失敗、0/false/null/精度超過、単位不明、並行確定・定義変更CAS、監査失敗全rollback、旧PUT競合、retain/review/明示removed集合の完全一致、削除済fieldと別PDF旧由来保持、create-only原actorの限定receipt回復、resumeのdocument_id/run_idとrun_version、同key期待CAS変更と原成功replay、現権限のPDFstream、履歴原本消去不能、NOT_RUN表示。試験実行・本物PDF処理なし。

## ソース

基準commit以下のパス（永久リンク）:
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/routes_records.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/model/storage_clinical.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/checkup_field_result_service.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/checkup_service.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/medical_record_image_request.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/infra/uploader.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/infra/local_uploader.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/lab_import_service.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/design/domains/checkups/technical-spec.md
