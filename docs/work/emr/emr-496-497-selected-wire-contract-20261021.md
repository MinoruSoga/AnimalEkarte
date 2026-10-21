# EMR-496 / EMR-497 採択技術契約 R5

2026-10-21。独立レビューを経て採択した技術契約。R4の保存・監査・hash契約を維持し、共通restore_contextと純読取baseline snapshotの2点を統合した公開用単一文書。既存494の採用契約を拡張する追加schema/wireであり、runtime実装、DDL適用、配備、試験は未実施（NOT_RUN）。技術選定済みと実装済みを区別する。

## 1. 現存契約と照合根拠

496の編集基盤は既に選定済み。docs/work/emr/emr-494-edit-draft-contract-20261020.md §3–5に GET/POST edit-right、共有GET /medical-records/{id}/operations/{operation_key}、既存PATCH /medical-records/{id}/inquiries、inquiry namespaceのowner CAS、acquire/take-over時baseline準備、clear/copy/Undo含む全writer参加がある。新たな問診save専用routeやCAS表・第二編集sessionを作らない。

backend/internal/medicalrecord/routes_records.go は上記PATCHの実登録を確認。inquiry_request.go現実装の受理fieldはchief_complaint、chief_complaint_type_id（三値）、notesだけ。inquiry_response.goは旧scalarを返し、initial_intake_revisions/entriesの構造化DTOはない。これは既存PATCH再利用を否定するものではなく、そのpayload/read projectionの不足。

497は同route登録ファイルに専用予防処方/投薬/帳票routeなし。検索preventive_prescriptions/PreventivePrescriptionはDDL/ERD/storage_clinical.go等、preventive-prescriptionsは0。D03technical-specのmedical-record-intake-prevention-contractは論理項目表で専用wire未選定。検索結果だけで全repo不存在と断言しない。

backend/docs/api.yaml はfetch_fileが空だったが、fetch_blobによりblob SHA9459cb23c968975417f4a44ea6f8d31ae701176f全文1,012,609文字を取得し照合済み。26408–26452行にPATCH /medical-records/{id}/inquiriesがある。現在YAMLのrequest chief_complaint_category_idは実requestのchief_complaint_type_idと不一致で、494/496共同更新時に既存実fieldへ整合する必要がある。Inquiry schema5156行、MedicalRecordInquirySummary3979行は旧scalarのみ。全文にpreventive/initial-intake/initial_intakeの専用wireはなし。494採用docsのC06拡張もこの既存PATCH YAMLには未反映。『既存wireをそのまま使えば完成』とはしない。

## 2. 共通保存・認可原則

server /api/v1。既存Cookie認証・X-Requested-With。新JSON ID/revisionは正規形の正BIGINT十進文字列。actor/clinic/recorded_at/operation_idをbodyから採用しない。現在grantと医院/患者/record対応をserver解決。

既存C06 operation_key、ResultResponseのwire意味、HTTP/status mappingを再利用する。成功/非成功の専用adapter・DTOは第18節で定義する。変更commandはtyped factoryでimmutable ContentとRequest内VerifiedCommandを同時生成し、Apply前に再計算照合。対象ref/選択した保存版/業務payload/復元目的はContent、現在epoch/期待owner版はPreconditions。別captured payload、可変pointer、別targetからApplyしない。

非変更bindinglookupが存在なら元anchor現在viewで照会/replay。不在が確定した新keyだけ現在write grant/scope preflight後に短claim。lookup失敗は不在でない。新bindingと元target anchor同TX、業務FKで共通gate前の逆順lockを取らない。race loserは元anchor再認可。未成功Applyだけ現在write grant、必要fence/CASを再検査し、業務版/必須履歴/audit/成功receipt同TX。入力不正・新key権限不成立でbinding等を作らない。

元結果不明は同key照会・同Content再送。新key作り直しや再送による二重記録なし。再認可後の比較回復は494本人draft契約を再利用する。本文・認証情報をログ/receiptに複製しない。

## 3. 公開互換とrecordless範囲

496：scalar公開名は現在BE/FE/model/DDL/回帰仕様S27が一致するchief_complaint_type_idをcanonicalとする。ただし旧公開YAML26432行はcategoryを宣言し、Inquiry5156のresponseにもcategory/type両方があるため、単にcategoryを削除しない。独立category masterは検索で確認できず、公開master26120は /masters/chief-complaint-types、同一主訴区分を指すaliasとして一時受理する。

新adapterはcategory/typeを別々にpresence追跡して読み、片方のみならその値、双方同値（双方null含む）なら1値、双方不一致ならclaim前400。省略=保持、null=解除、値=現医院の同chief_complaint_types参照。新FEはtypeのみ送る。DTO/codegenにはcategory deprecated aliasを明示し、current read responseは両fieldを同じ値で返す互換期間を設ける。別意味のcategory値を勝手にtypeへ変換しない。実利用ログ/既存consumer棚卸しなしにaliasを除去しない。このalias追加は現実装済みの主張ではなく、旧公開契約を壊さず不整合を閉じる採択技術契約。

上記のnull解除は主訴type / category aliasの三値fieldに適用する。chief_complaint / notesは現行*string DTOとrepositoryの非nil更新を維持し、JSON nullをnullable保存や解除へ読み替えない。wire→immutable command→storage→snapshotは以下の1方式に固定する。

| field | 省略 | JSON null | 空文字 | 保存 / present snapshot |
|---|---|---|---|---|
| chief_complaint / notes | 保持 | 保持（省略と同じ） | 明示的に空へ変更 | text NOT NULL。空は空のまま |
| chief_complaint_type_id / category alias | 保持 | 解除 | ID不正として拒否 | NULL又は検証済type ID |

typed factoryはchief_complaint / notesのnullを保持操作としてcanonical payloadから省略し、空文字はfieldを残す。typeのnullは解除としてpayloadへ残す。alias正規化も同factory内で行い、hashとApplyの意味を一致させる。present snapshotではchief_complaint / notesは非NULL、typeだけNULL可、absent snapshotは本文3値と名称をNULL、名称basisをnot_selectedとする。「null / 空保持」はこのfield / presence別の意味であり、presentに保存不能なNULL textを許す宣言ではない。既存inquiriesのNOT NULLを緩和するDDLは不要。新snapshot表・alias / presence decoder・OpenAPI説明との同期は未実装の依存であり、今回適用・codegenは行わない。

既存scalar IDのnumber型を今回無断で一括string化しない。JSON数値はBEで整数を厳密decodeし、現FE受理は安全整数まで。今回追加するentry/revision/operation target refsはcanonical decimal string。旧scalar拡張と全既存ID移行を同一作業に膨らませない。

497：一般的なrecordless新create/clearの業務承認は確認できない。新createは POST /medical-records/{id}/preventive-prescriptions に限定。petはrecordからserver解決。既存処方のmedical_record_id=NULLはread/history/帳票対象として保持するが、新規作成、改訂、clear、投薬writeはこのsliceでは許可しない。recordを推測補完/自動作成/後付けしない。必要なら既存専用訂正ownerの設計へ別途渡す。

したがってrecordless側に新write CASは導入しない。読取のrevision/manifest整合だけを検査する。recordbound側は処方/投薬writeすべて同recordの494session/epoch fenceへ参加し、処方・観察それぞれのexpected_revisionを独立CASとする。owner_expected_version=inquiryを予防処方へ流用しない。

## 4. 496 route/DTO

変更routeは既存 PATCH /medical-records/{id}/inquiries のまま。494の平坦operation_key/edit_session_ref/edit_epoch/owner_expected_versionに、optional initial_intake、optional restore_context（第5節の共通4 field）、history_copy時のtyped copy_sources（第11節）を追加する。既存scalarのomitted/null意味は変更しない。

initial_intake省略=構造化部分不変、object=構造化部分の完全な置換snapshot。nullは不正として曖昧な消去に使わない。空entriesは明示空回答版であり過去版削除ではない。objectのfieldはsource_kind、source_reference(nullable)、heard_at(nullable)、copied_from_revision_id(nullable)、entries配列。source_kindは既存owner_report/paper_transcription/record_reference/legacy_unknown。コピー参照はserverが元版scope・閲覧・選択一致を検証し、任意文字列だけでprovenanceを宣言させない。

各entryは既存storageのentry_key UUID、field_key、field_label、answer_state yes/no/unknown/null、detail、drug_description、symptom_description、reported_result、time_precision/event_year/event_month/event_date/time_description、reference_exam_revision_idを対応付ける。nullable欄はobject内で明示null可、未入力と回答unknownを分離する。entry_keyは同回答項目を版間で追跡し重複を拒否。自由文薬名をmedicine IDへ推論しない。D07参照は同医院/患者の確認可能な版だけ、問診writeはD07/D08/D04/LINE原文を変更しない。

新規 GET /medical-records/{id}/inquiries を現在版の全問診読取として追加する。既存record一覧/詳細のMedicalRecordInquirySummaryは軽量のまま保持し、全構造化entriesを一覧へ展開しない。clinic/pet/record、owner_version nullable＋version_state、intake_revision_id/revision nullable、source/entries/countを含む。未作成と読取失敗は別。既存scalar読取を撤去しない。新規 GET /medical-records/{id}/inquiries/revisions/{revision_id} を固定版読取用にだけ追加。構造化の履歴選択は既存同患者カルテ履歴からこの固定版を引き、scalarを含むowner版全体は第5.1節のowner-version readerを使う。ページング全史APIを無根拠に必須化しない。

## 5. 496 CAS/復元write

同一PATCHで旧scalarと構造化が変わる場合、494のinquiry owner_versionを1回だけ増分。scalar-onlyでもstructured-onlyでも、新owner版には全scalar＋当時intake参照の不変snapshotを第5.1節の表へ保存する。initial_intakeのrevisionは構造化snapshotの改訂順で、inquiry owner_versionと等しいと仮定しない。scalarだけ変更なら構造化revisionを捏造しない。構造化が変化する時だけ新revision＋全entriesを同TXで追加し、以前の版は不変。初回revision1は初回成功構造化write、494のbaseline1とは別。

restore_contextは496・497共通のoptional closed object {kind,baseline_ref,source_operation_key,selected_entry_keys}とし、指定時は4 fieldすべて必須。kindはtab_open_clear/history_copy/copy_undo、baseline_refは確立済み本人baselineのUUID、source_operation_keyはcopy_undoのみ元成功copy UUID（他kindはnull）、selected_entry_keysはUUID昇順・重複なし配列（entry選択なしは空配列）。省略は通常保存、nullは不正。tab_open_clearはtab_open基準、history_copy/copy_undoは対応copy_before基準を使う。旧案のbase_intake_revision_id/source_revision_idをrestore_contextへ混在させない。コピー元は第11節のtyped copy_sourcesだけで指定し、復元基準とコピー元のidentityを混同しない。復元値は読み取った固定版からserverが解決し、入力bodyの任意復元値を信用しない。baselineの保存edit_session_ref/edit_epochが現在保持中のfenceと完全一致することを同TXで検査する。release後の再acquire・別tenureの旧baselineは同accountでも使用不可。期待inquiry版、元自操作列、復元対象の後続変更を照合し、他者編集/対象外後続変更があれば競合停止。復元は新履歴で過去削除なし。

### 5.1 inquiry owner版の限定append-only snapshot

新表inquiry_owner_snapshotsを採択する。列はclinic_id / medical_record_id / owner_version（正bigint、NOT NULL）、inquiry_id bigint nullable、presence_state text NOT NULL CHECK IN('present','absent')、chief_complaint text nullable、chief_complaint_type_id bigint nullable、chief_complaint_type_name text nullable、chief_complaint_type_name_basis text NOT NULL CHECK IN('recorded','legacy_unknown','not_selected')、input_operation_kind text NOT NULL、copy_source_refs jsonb NOT NULL、notes text nullable、intake_revision_id bigint nullable、snapshot_basis text NOT NULL CHECK IN('observed_baseline','saved_change')、operation_id bigint NOT NULL、recorded_by bigint NOT NULL、recorded_at timestamptz NOT NULL（server時刻）。PKは(clinic_id,medical_record_id,owner_version)。既存494のowner_versionだけを使い、別row ID / 第二counter / 別CASを作らない。

record FK(medical_record_id,clinic_id)→medical_records(id,clinic_id)、operation FK(operation_id,clinic_id)→operation_bindings(id,clinic_id)、staff FK(recorded_by)→staffs(id)、intake FK(intake_revision_id,clinic_id,medical_record_id)→第9節で選定済みのinitial_intake_revisions同UNIQUEを全RESTRICTとする。inquiriesにはclinic列がないためUNIQUE(id,medical_record_id)だけを追加し、FK(inquiry_id,medical_record_id)→inquiries(id,medical_record_id)で同recordを保証する。mutableな現在owner版へ過去版FKを張らず、版一致は同TXの既存owner CASで検査する。type IDと第20.4節の名称/basisは当時のsnapshot値であり、現在master変更で過去値を書き換えない。

CHECKでpresentは実inquiry_idと非NULL chief_complaint / notesを必須としtype IDはNULL可、absentはinquiry_idと本文3 scalarとnameをすべてNULL、name_basisをnot_selectedとする。intake_revision_idは当時の構造化head参照として独立に保持し、存在しないintake版を捏造しない。通常invokerのUPDATE / DELETE拒否trigger、repository INSERT / SELECTのみ。RLS ENABLE、既存apply_rls_policyのFOR ALL USING / WITH CHECKにhas_clinic_access(clinic_id)。repositoryのtrusted selected clinic、現在record view / 同患者検査も維持し、FORCE / bypass / 新roleは追加しない。本人private baselineのRLSを臨床履歴の閲覧権へ流用しない。

全inquiry writerは既存gate / fence / record lock / owner CAS下で変更前の既知snapshotを照合する。既存行があれば臨床値・presence・当時intake参照を完全比較し、不一致では上書きせず全体停止する。既存行の元operation / actor / 時刻は保持し、新操作のmetadataへ置換しない。初期化又は初回実変更で旧owner版のsnapshotが未記録なら、今観測できた値だけをobserved_baselineとして保存する。変更後はscalar-only / structured-only / 両変更 / 検証済復元の各新owner版にsaved_changeの完全snapshotを1行appendし、実変更前後の既知snapshot・owner版増分・必須監査・成功receiptを同TXで確定する。失敗は全rollback。no-op / replayでは観測行も変更行も増やさず、GETは初期化・履歴作成を行わない。未取得の過去版はbackfillしない。

observed_baselineのrecorded_by / recorded_at / operation_idは「このsnapshotを観測した操作」を表し、過去の入力者・元保存日時・元操作ではない。saved_changeだけが当該owner版を保存した変更操作を表す。この意味区分をsnapshot_basis付きDTOで明示し、過去actor / 時刻を推測補完しない。

固定readerはGET /medical-records/{id}/inquiries/owner-revisions/{owner_version}。現在record viewで同clinic / record / petを認可し、副作用なし・Cache-Control:no-store。closed DTOは{medical_record_id,owner_version,inquiry_id,presence_state,scalar_snapshot,intake_revision_id,input_operation_kind,copy_source_refs,snapshot_basis,recorded_by,recorded_at,operation_key,snapshot_sha256}。新ID / 版は正十進string、nullable refsはnull。scalar_snapshotは第20.4節の名称固定2 fieldを含むclosed {chief_complaint,chief_complaint_type_id,notes,chief_complaint_type_name,chief_complaint_type_name_basis}（null / 空保持、新固定DTOのtype IDはstring|null）。既存scalar routeのnumber互換は変更しない。intakeは当時の固定refから既存版readerで全entriesを再読し、現在headを混ぜない。未観測版や取得失敗を空snapshotへ置換しない。既存/inquiries/revisions/{revision_id}はintake版専用のまま、owner_versionへ読み替えない。

nullableなIDも非NULL時は正bigintを要求する。presence / snapshot_basis / payload相互条件はtable-local CHECKとtyped decoderで検証し、未知値・部分snapshotを保存しない。

snapshot_sha256はlowerhex SHA256(UTF8(JCS({schema:"inquiry-owner-snapshot-v1",medical_record_id,owner_version,inquiry_id,presence_state,scalar_snapshot,intake_revision_id,input_operation_kind,copy_source_refs})))。観測/保存actor・時刻はコピー臨床値に混ぜず、固定値とrefだけをcommitする。private本人baselineはこの臨床履歴readerの代用にしない。

fixture計画（全NOT_RUN）：連続scalar-only保存の各owner版を再読、structured-only版で同scalar＋当時intake参照を再読、scalar-onlyでintake版増分なし、変更前snapshot不一致の全rollback、observed_baselineの観測actor表示と過去actor非推測、GET / no-op / replayによる履歴増殖なし、別clinic / record / pet非開示。3 branchの未知field / ref混在 / 欠落 / hash不一致と、owner版をintake版へ偽装したコピーを拒否し、manifestと処方provenanceが同じ固定sourceへ戻ることも確認する。

## 6. 496 UI/変更候補

治療プラン内スタッフ問診区画、既存自動保存adapterを拡張。独立tab/保存buttonを作らない。tab開始基準と複製直前基準を別保持し、遅延save成功で基準を上書きしない。原文は既存read-only別区画。

候補変更はinquiry_request/response/handler/service/repository.go、既存api/inquiries.tsとactive-tab保存hook。新private initial_intake_{projection,participant}.go。494共有router/runner/anchor部分は必要subsetだけowner受渡し。

## 7. 497 A/B/C routeと認可

A固定処方：GET /pets/{pet_id}/preventive-prescriptions、POST /medical-records/{id}/preventive-prescriptions、GET /preventive-prescriptions/{id}/revisions/{revision_id}、POST /preventive-prescriptions/{id}/revisions。
B投薬確認：GET /preventive-prescriptions/{id}/medication-observations、POST同、POST /preventive-medication-observations/{id}/revisions。POSTは保存された関連recordが存在する場合のみ、その元recordの現在fenceを要求する。
C固定reader：GET /pets/{pet_id}/preventive-reports/prescription?prescription_id=&revision_id=、GET /pets/{pet_id}/preventive-reports/annual?year=&prescription_ids=。readのみで保存/発行/配信operationを起こさない。

新write入口MedicalRecords view、未成功createは既存MedicalRecords create、改訂はedit。grant新設/職員割当変更なし。D21ページと専用出力はMedicalRecords view＋Owners view、現在同owner/pet認可。recordlessはreadのみ。旧storageのNULLをrecord自動作成/推測付替えで埋めない。

## 8. 497 DTO/版

A create body：operation_key、source_treatment_id nullable、content。medical_record_idはpathから、pet_idはrecordからserverが解決する。改訂body：operation_key、expected_revision、content。すべてのwriteでedit_session_ref/edit_epochを平坦に追加。A/Bのcreate・改訂DTOすべてにoptional restore_contextを第5節と同じclosed 4 fieldで明示し、history_copyを行う処方writeには第11節のtyped copy_sourcesを持たせる。既存投薬事実に対するclear禁止等の安全境界は第9節のまま、field追加で操作を新たに解禁しない。public contentの利用者入力はprescription_title、prescribed_date、medicine_id/medicine_name_snapshot、administration_conditions、period_descriptionと開始終了partial date、remarks、responsible_staff_idのみ。copied_history_record_id / copied_history_revision / history_snapshotはpublic contentとpayload hashから除外し、指定されたら拒否する。検証済copy_sourcesだけからserverが固定元参照・snapshotを導出し、その由来はsource_sha256で束縛する。記録actor/時刻/operation_idもserver決定。copy元参照とsnapshot一致はserver検証、当前master名で旧版を書き換えない。

B create body：operation_key、prescription_revision_id、display_year、content。改訂body：operation_key、expected_revision、prescription_revision_id、display_year、content。display_yearは通常create/改訂で必須、第20.5節の版別表示年に保存する。content=state(planned/taken/not_taken/unknown)、source_kind(performed_record/owner_report/legacy_unknown/other)、source_reference、confirmed_at、partial time、notes、change_reason。stateや頻度を処方から自動生成しない。未入力は記録不存在でunknown回答とは別。display_yearは1..9999、出来事の年月不明を1月へ補完しない。

新規・実編集のmedication writeは、親lock下で参照prescription_revision_idの同clinic/prescription対応とpresence_state='present'を必須検査し、absent処方版を参照した新しい投薬事実は拒否する。第9節の検証済clear/Undoによるabsent復元は新しい投薬事実ではなく、元prescription_revision参照を保持する別経路とする。この区別で依存する実投薬の保護を解除せず、absent復元を通常create/改訂の抜け道にしない。

recordbound親row lockでrevision順を直列化し、期待revision一致を検査してappend-only revisionを作る。headは同親のMAX(revision)をlock内で解決してよいがMAX(id)を版として使わない。新createはparent pet/recordの安定lockで競合を直列化、source_treatment_idの既存uniqueに従い別key重複も検出。処方に結び付く投薬revisionは同clinic/prescriptionの正確なprescription_revision_idを保持、最新処方へ自動付替えしない。

## 9. append-only存在状態

比較：A）通常レコードを物理DELETEは履歴/元ID消失で不可。B）unknown/not_takenへ書換は臨床事実改変で不可。C）別sidecarだけのtombstoneは全readerで二重head合成が必要で、固定revisionの意味が曖昧になる。D）各immutable revisionに明示present/absentを持たせる方式を採択する。

追加対象はinitial_intake_revisions、preventive_prescription_revisions、preventive_medication_revisionsの3表に限定。presence_state text NOT NULL CHECK IN('present','absent')。既存revisionは旧入力が保存された事実としてpresentへ技術移行（臨床値の推論や旧行の削除はしない）。通常新規/改訂はpresent、既決の開始時クリア/複製取消により元が不存在だった場合だけabsentを新revisionとしてappendする。通常payloadでpresence_stateを書かせず、serverの検証済restore_contextから導く。

復元self参照は同じ業務identityまでDB制約に含める。initial_intake_revisionsへUNIQUE(id,clinic_id,medical_record_id)、preventive_medication_revisionsへUNIQUE(id,clinic_id,observation_id)を追加し、restores_revision_idはそれぞれ(restores_revision_id,clinic_id,medical_record_id)／(restores_revision_id,clinic_id,observation_id)のself FKとする。preventive_prescription_revisionsは既存UNIQUE(id,clinic_id,prescription_id)を再利用し同3列self FK。全FKはRESTRICT。別カルテ/観察/処方の版を同clinicというだけで復元元にしない。

各revisionに restoration_kind nullable enum(tab_open_clear,copy_undo)、restores_revision_id nullable、restores_operation_id nullable、restoration_reason nullableを追加する。self FKは上記のidentity付き複合FK、operationは同clinic bindingへのFK。presence=absentならrestoration_kindと元自操作refを必須とし、対象が基準時不存在だったことを後述manifestで検証する。present復元も同じmetadataを保持する。これらは臨床stateではなく入力状態の復帰を表す。

history_copyで生成したrevisionはrestoration_kindと復元専用restores_revision_id / restores_operation_id / restoration_reasonをnullとし、復元enumを増やさない。revision又はinquiry owner snapshotのoperation_idから同clinicの成功copy operation、operation_kind='history_copy'のrestore manifest、同operationのauditへ結び、entry provenance又は処方history_sourcesから固定コピー元を辿る。scalar-only copyの臨床readerは第20.1節のclinic-scoped input_operation_kind/copy_source_refsを読み、private restore manifestを読取条件にしない。

initial_intake absentはentry_count=0、entriesなし。以前のsource_kind/source_reference/copied_from等は過去present revisionから読めるため、absentに新しいowner回答unknownを生成しない。absent revisionの既存source_kindはrestoration元の記録出所を維持し、回答内容と扱わない。source_kind=legacy_unknownを自動埋めする方式にはしない。

preventive prescription absentは元identity/record/petと復元metadataを保持し、利用者入力の臨床payloadは全NULL。medication absentも元observation/prescription/prescription_revision参照と復元metadataは保持するが、stateをNULLにできるDDLへ限定変更し、CHECK((presence='present' AND state IN('planned','taken','not_taken','unknown')) OR (presence='absent' AND state IS NULL))を追加。state=NULLは未回答や未投薬でなく『入力状態が不存在へ復帰』。source_kindは元版の出所を保持して新しい確認事実にしない。presentでstate NULLは禁止。absentに臨床値を混ぜるwriteは拒否。

現在readerは最新revisionのpresenceを返す。absentは通常一覧に投薬/処方として描かず、履歴では復元操作と以前のpresent版を辿れる。固定過去revision readerは当時のpresenceを返す。absentから通常再入力は同identityへ新present revisionをappendし、別IDを作らない。

安全境界：既に実投薬の証拠・別owner効果に使われた保存内容をclear/Undoで不存在にしない。検証済の自入力/自コピーの復帰だけ、履歴・current CAS・元自操作列を照合する。依存する投薬/請求/出庫/提示版等が存在するなら影響を表示して既存専用訂正へ渡し、元事実やそのreceiptは不変。月セル実績の取消を新しい汎用clearとして解禁しない。

## 10. entry由来・不変baseline

比較：revision.source_kind単一のままでは複数sourceコピーを正確に保持できない。自由文source_referenceへJSON詰込みは参照制約/照合を失う。entriesへsource列を複製すると1entryに複数由来がある場合に再不足する。最小immutable mapping表を採択する。

新initial_intake_entry_provenance：clinic_id、intake_revision_id、entry_key、source_ordinal（非負、PK全4列）、source_kind（既存4値）、source_reference nullable、source_record_id nullable、source_intake_revision_id nullable、source_entry_key nullable、heard_at nullable、copied_at nullable。FK(clinic_id,intake_revision_id,entry_key)→initial_intake_entries、元entry参照は同clinicの完全triple FK。元record/pet一致はserver resolverで検査。参照tupleは全NULL又は全指定、record_referenceなら実元tuple必須。元別患者/未取得版/存在しないentryを禁止。

initial_intake_entry_provenanceはRLS ENABLE、既存app_private.apply_rls_policyでFOR ALL USING / WITH CHECKにapp_private.has_clinic_access(clinic_id)を適用する。repositoryもtrusted selected clinicで絞り、現在元target viewと同患者照合を行う。既存clinic境界を維持し、FORCE / bypass / SECURITY DEFINER / roleは新設しない。

この契約を有効化した後の新規revisionの各entryは少なくとも1由来mappingを同TXで持つ。既存revisionへ架空mappingをbackfillしない。既存版はprovenance_basis=legacy_revision_levelとして元revisionの実在source_kind/source_reference/copied_from等だけをread projectionへ返し、entry単位が確認済みとは表示しない。旧版のコピー時も元entry由来不明を保持して旧revision-level参照を付け、推測補完しない。通常入力は聞取/紙等の実sourceを保持、選択コピーは元immutable mappingを再利用して直接元entry参照を追加する。複数sourceを1つへ偽装しない。revision-level source_kindはrevision生成操作の由来を示す互換summaryとし、entry由来を上書きする正本ではない。履歴readerは新規版ではmapping、既存版ではlegacy_revision_levelと実在revision-level sourceを返し、旧mapping欠如を新規版の検査免除に広げない。sourceを信用して単なる回答を確認済み検査へ昇格しない。

復元基準はclientの『初期不存在』を信用せず、変更送信前にserverの同TX session/epoch/owner版・全集合digest照合で確立する。tab_openとcopy_beforeは別基準で再fetch/保存による上書きなし。旧scalarもprivate不変snapshotに保持する。正確な保存表・PK/FK/RLS・wire・再試行は第16節、C06 receipt/hashは第18節を唯一の定義とする。

## 11. 処方用copy mapping・source重複・lock順

既往歴copyは下記typed copy_sourcesの固定refとselected entry refsをserver readerで実読し、同clinic/pet・現在view・保存版・必要全項目を検証する。bodyのhistory_snapshotは採用せずserverが固定元から構成する。copied_history_revisionがどのowner版かを曖昧にせず、構造化496由来はinitial_intake revision IDを型付きrefとして下記処方専用mappingに持つ。旧record revisionだけの不明参照を現在版へ読み替えない。scalarを含む新規臨床履歴選択はinquiry_owner_snapshotのrecord_id / owner_versionを第5.1節readerで実読する。従来のinquiry_snapshotは明示選択された本人baseline_ref UUID＋owner_versionだけを扱う限定branchとして維持し、本人認可を要する。臨床履歴が欠ける時にprivate baselineや現在値へfallbackせず、未観測の過去owner版をコピー可能としない。intake revision IDやrecord revision番号でowner版を代用しない。

copy_sourcesはhistory_copy時に必須の非空typed配列で、通常保存/clear/undo時は省略する。各要素はdiscriminated closed object。kind=initial_intakeは {kind,record_id,intake_revision_id,entry_keys}、kind=inquiry_snapshotは {kind,record_id,baseline_ref,owner_version}、kind=inquiry_owner_snapshotは {kind,record_id,owner_version} のみ。record/intake revision/owner版は正十進string、baseline_refはUUID、entry_keysは選択した実entry UUIDの昇順・重複なし配列。inquiry_snapshotのbaselineはinquiry scopeの保存済み完全snapshotであり、owner_versionはその不変保存版と一致することを要求する。現在owner版へ読み替えず、現在元record viewと本人baseline認可は別途必要。initial_intake branchにbaseline_ref/owner_versionを、inquiry_snapshot branchにintake_revision_id/entry_keysを、inquiry_owner_snapshot branchにbaseline_ref/intake_revision_id/entry_keysを混ぜない。inquiry_owner_snapshotは同clinic / 同患者の実snapshotと元record現在viewを要求し、private baselineの本人認可を臨床履歴認可へ流用しない。scalar sourceからはscalarだけを復元し、構造化entriesやentry provenanceの元tupleを推論生成しない。旧base_intake_revision_idは復元baselineのmanifest内固定版へ、旧source_revision_idはcopy_sourcesのinitial_intake branchだけへ対応し、公開DTOに旧fieldを残さない。既存copied_from_revision_id等のstorage summaryは検証済みtyped sourceから導く互換値であり、別コピー元を宣言する第二経路にしない。

history_copyは保存済み内容の取り込みに限定し、absent sourceは選択不可とする。initial_intakeは選択revisionのpresence_state='present'、inquiry_snapshotは本人baseline内の選択inquiry itemのpresence='present'、inquiry_owner_snapshotは固定owner snapshotのpresence_state='present'を必須とする。未作成・実absent版ともcopy source validationで拒否し、存在をunknown回答や空内容へ変換しない。server source reader / factoryでclaim前に検査し、Applyでも同固定ref / digest / presence / 現在認可を再検査する。manifest / provenanceへの記録もこの検証済present sourceだけに限定する。absent版の履歴読取は維持し、本人のclear / Undoが開始時不存在へ戻す別経路は第9節の復元metadata・安全条件に従う。history_copyでabsent復元やrestoration enumを新設しない。

処方用に別のimmutable preventive_prescription_history_sourcesを選定する。列はclinic_id,prescription_revision_id,prescription_id,source_ordinal（PKはclinic_id,prescription_revision_id,source_ordinal）、source_kind='initial_intake'|'inquiry_snapshot'|'inquiry_owner_snapshot'、source_record_id、source_intake_revision_id nullable、source_inquiry_baseline_ref nullable、source_owner_revision nullable、copied_at、copied_text。FK(prescription_revision_id,clinic_id,prescription_id)は既存処方revision UNIQUEへ。initial_intake branchは実intake revisionと第20.2節child tableの選択entry全集合への参照を要求し、inquiry_snapshot branchは実baselineの固定scalar snapshot＋owner版を参照する。inquiry_owner_snapshot branchは(clinic_id,source_record_id,source_owner_revision)→inquiry_owner_snapshotsのPKへRESTRICT FKで参照し、source_inquiry_baseline_ref / source_intake_revision_idはNULLとする。新しい列や第二owner版は作らず既存source_owner_revisionを使う。CHECKで3 branch固有refの混在/欠落を拒否する。initial_intake_entry_provenanceは引き続き実intake entryの完全tupleを要求し、owner snapshotをintake版や元entryへ偽装しない。entry表へのmappingを処方に流用しない。元record/current clinic/pet・版/entry存在/現在viewはresolverで検証し、copied_textはその固定sourceからserverが構成する。複数sourceを順序付き複数行として保持し、元版/原文は不変。旧処方revisionは既存copied_history_record_id/revision/history_snapshotの実在値だけをlegacy projectionし、この新tableを架空backfillしない。

preventive_prescription_history_sourcesもRLS ENABLE、既存app_private.apply_rls_policyでFOR ALL USING / WITH CHECKにapp_private.has_clinic_access(clinic_id)を適用する。repositoryはtrusted selected clinicで絞り、全branchで現在元target view / 同患者を検査し、inquiry_snapshot branchだけ追加で本人baseline認可を検査する。clinic境界を越える由来・copied_text読取は許さず、FORCE / bypass / SECURITY DEFINER / roleは新設しない。

source_treatment_id付き新keycreateは同clinic/record/petの実sourceを非変更preflightし、後段で必要gate→lifecycle→edit_right→record→対象treatment ID順→既存prescription ID順→observation ID順→revision/metadataの順でlock。既存lifecycleに予約gateが必要な時だけ先行取得し、全予防readへ予約lockを広げない。親のrecord lockで当該recordの新prescription作成を直列化する。

同source unique(clinic_id,source_treatment_id)の既存prescriptionがある新keyは、自動的に第二処方を作らず、かつ既存処方内容を上書きしない。現viewで既存IDを照会可能にしたうえでdependency-changed(409)として同identity編集/再確認へ誘導。同key再送なら元receiptのみ。source処方の内容一致だけで別keyを成功replay扱いにしない。NULL sourceはこのuniqueを重複対策に使えず、明示operation identityがcreate重複防止の単位。

既存recordboundの改訂はrecord/親prescriptionの順にlockし、必要observation/revisionを安定ID順へ。expected_revisionとMAX(revision)は同親lock下で比較。新revision、provenance/restore metadata、必須audit、receiptは同TX。immutable factoryが読み取った元snapshot/digestをContentへ結合し、claim後に違うcopy元や最新masterを採用しない。競合した新読取/選択変更は再確認して新keyとする。

## 12. 専用2帳票の固定reader

reader response：clinic/owner/pet、report_kind、year又はprescription_id、保存版manifest（処方id/revision_id、各observation id/revision_id）、scope_digest、complete、items、unknown/未入力凡例。serverの同一read snapshotで対象集合を確定。年間全件を100件で切らず、必要ページがある場合はmanifest固定cursorで全取得、欠落・版変化ならcomplete=falseで印刷不可。

previewと出力直前に同一scope・manifestを照合する。新規観察追加も検知する必要があるため、既知ID版だけ比較せず対象集合全体digestを再計算。全件不足、権限失効、他患者切替は出力停止。管理入口だけD03へ渡し、D21はwriteを持たない。再印刷/閉じる/印刷結果不明は状態変更なし。

## 13. C06結果wire/HTTP

496・497の変更consumerは494採用文書§3.2のMedicalRecordOperationResult mappingを同一採用する。bodyはstatus,operation_key,result_refs?,committed_versions?,recorded_at?、wrapperなし。

- succeeded / replayed：200。
- validation：400。
- version-conflict / dependency-changed / edit-right-lost / operation-key-conflict：409。
- operation-in-progress：202。同keylookup、送信結果の推測禁止。
- scope-not-found：404。
- not-authorized：403。
- result-unavailable：503。結果不明停止、照会とRetry-After尊重。
- not-succeeded：200。ただし保存成功ではなく、再確認後に同key再送できる候補。

入口JSON/alias不一致/認証/現在scope拒否は既存ErrorResponse{error:string}の400/401/403/404。C06成功のように扱わない。unexpected transport/500も成功不明で新key作成禁止。read GETは通常200のprojection、取得失敗は非2xxで空entries/complete=trueへ置換しない。C06 enum新設やgenericRunner全consumer変更をしない。

共通lookupは496=既存494 /medical-records/{id}/operations/{operation_key}?intent=medical_record.inquiry.save。497 recordbound=同lookupにmedical_record.preventive_prescription.create/revise、medical_record.preventive_medication.create/reviseを限定追加し、既存medical_record_operation_targetsを再利用する。pet operation anchor/lookupは追加しない。recordlessはread-onlyのため新operation anchor不要。履歴readには現在元target viewを要求し、旧epochをsuccess replay条件にしない。

## 14. 実装分割・残る業務境界

496-A：494既存writerへ構造化payload/projection＋entry provenanceを接続、旧alias互換、通常自動保存。496-B：baseline/restore metadataとclear/copy/undo。双方計画を同票に保持し、Aだけで全AC完了とはしない。
497-A：recordbound固定処方・改訂・読取、既存recordless read-only。497-B：投薬確認の独立append revision、存在復帰保護。497-C：専用2帳票固定manifest readerと既存D21導線。A/B/CそれぞれNOT_RUN、master/provider/物理印刷/実配信なし。

recordless writeの未採択境界は497の全体ACから削除する意味でも497完了の意味でもない。read-only sliceと残る業務入口確認を票に分離し、今回の部分接続を全機能完了へ昇格させない。

新しい臨床値/頻度/適応の質問は追加不要。recordless新createを本sliceへ含める要求が将来出た場合だけ、現在未確認の業務入口/承認を親へ提出する。ここでは安全なread-only境界で閉じる。

## 15. 固定根拠

https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/work/emr/emr-494-edit-draft-contract-20261020.md
https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/backend/internal/medicalrecord/routes_records.go
https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/backend/internal/medicalrecord/inquiry_request.go
https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/spec/domains/medical-records/pages/initial-intake.md
https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/spec/domains/medical-records/pages/preventive-medication.md
https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/design/domains/medical-records/technical-spec.md#medical-record-intake-prevention-contract
ERD6表：同SHA docs/architecture/erd-detail/initial_intake_revisions.md、initial_intake_entries.md、preventive_prescriptions.md、preventive_prescription_revisions.md、preventive_medication_observations.md、preventive_medication_revisions.md。

## 16. Baseline / restore metadata の閉じた保存契約

追加2表は以下の選定で固定する。これはschema/model ownerへ渡す具体定義であり、現存schemaだと主張しない。前節の曖昧な『schema ownerが合わせる』を置換する。

medical_record_input_baselines：clinic_id bigint NOT NULL、operation_id bigint NOT NULL、baseline_ref uuid NOT NULL、medical_record_id bigint NOT NULL、principal_account_id bigint NOT NULL、actor_staff_id bigint NOT NULL、edit_session_ref uuid NOT NULL、edit_epoch bigint NOT NULL CHECK>0、baseline_kind text NOT NULL CHECK IN('tab_open','copy_before')、scope_kind text NOT NULL CHECK IN('inquiry','preventive_prescriptions','preventive_medications')、owner_version bigint NULL CHECK>0、manifest jsonb NOT NULL、manifest_digest bytea NOT NULL CHECK octet_length=32、created_at timestamptz NOT NULL DEFAULT clock_timestamp()。PK(clinic_id,operation_id)、UNIQUE(clinic_id,principal_account_id,baseline_ref)、UNIQUE(clinic_id,baseline_ref,medical_record_id,principal_account_id)。baseline_refはこのbaseline作成のoperation_keyと同じUUIDであり、別の操作keyを追加しない。operation FK(operation_id,clinic_id)→operation_bindings(id,clinic_id)、record FK(medical_record_id,clinic_id)→medical_records(id,clinic_id)、account FK→accounts(id)、staff FK(actor_staff_id)→staffs(id)。全RESTRICT。baseline_kind/target/actorはbinding/Content一致もserver検証する。これら業務FKはclaimではなく、gate/fence後の業務TX内INSERTで検査する。

actor_staff_idの単列FKは履歴actorの実在・削除保護を担う。保存時に現在actor/account対応、有効assignmentと選択医院の現在grant、既存admin例外guardを検査し、record / 本人baseline認可 / fenceも維持する。live assignmentへのFKやstaff主所属医院との一致制約は追加せず、権限を拡張しない。これは既存複数医院staff契約への整合であり、guard / DDL実装済みの宣言ではない。

medical_record_input_restore_manifests：clinic_id bigint NOT NULL、operation_id bigint NOT NULL、medical_record_id bigint NOT NULL、principal_account_id bigint NOT NULL、baseline_ref uuid NOT NULL、operation_kind text NOT NULL CHECK IN('history_copy','tab_open_clear','copy_undo')、manifest jsonb NOT NULL、manifest_digest bytea NOT NULL CHECK octet_length=32、created_at timestamptz NOT NULL DEFAULT clock_timestamp()。PK(clinic_id,operation_id)。FK(operation_id,clinic_id)→operation_bindings(id,clinic_id)、FK(clinic_id,baseline_ref,medical_record_id,principal_account_id)→baselineの同UNIQUE、record/account FKはbaseline側で保証。全RESTRICT。元copyの成功subset/作成ID/直前値参照をこの表に保存する。失敗claimだけではこの表もbaselineも作られない。

両表は通常invokerのBEFORE UPDATE OR DELETE拒否trigger、repository INSERT/SELECTのみ、runtime TRUNCATE/DDL不可。任意cleanup/expiryは追加しない。RLS ENABLE、FOR ALL USING/WITH CHECKに既存app_private.has_clinic_access(clinic_id) AND principal_account_id = NULLIF(current_setting('app.current_account_id',true),'')::bigintを使う。trusted認証staff→accountをTXごとSET LOCAL、他actor/不存在accountでは停止。FORCE/SECURITY DEFINER/bypass新設なし。table ownerの絶対防御は主張しない。RLSに加え現在record view/edit、元staff/account、目的、target、fenceをアプリ層で検査する。

baseline manifestはclosed object {schema,record_id,scope_kind,owner_version,complete,items}。schema='medical-record-input-baseline-v1'、record_idは正十進string、owner_versionはinquiry版string又は未初期化時null、complete=trueのみ保存。items各要素は全6 field必須のclosed {kind,id,presence,revision_id,revision,scalar_snapshot}。kind順はinquiry→initial_intake→preventive_prescription→preventive_medicationに固定し、同kindの実IDは数値昇順。inquiry scopeはinquiryとinitial_intake singletonを各1件、予防scopeは対象親identityの全集合を含む。各fieldは次の閉表だけを許容し、代替表現・省略・0・空文字の偽IDを拒否する。

| kind / 状態 | id | revision_id | revision | scalar_snapshot |
|---|---|---|---|---|
| inquiry present | 実inquiry ID | null（別revision行IDなし） | manifest.owner_versionと同値 | 第20.4節を含むclosed scalar。textは非NULL、typeはstring又はnull |
| inquiry absent（未作成又は入力状態復帰） | null | null | manifest.owner_versionと同値 | 第20.4節を含むclosed scalar。absentでは本文3値/name=null、name_basis=not_selected |
| initial_intake 未作成absent | null | null | null | null |
| initial_intake 実revision present / absent | medical_record_id（singletonの実identity） | 当該intake revision行ID | 当該intake revision番号 | null |
| preventive_prescription 実revision present / absent | 実prescription ID | 当該処方revision行ID | 当該処方revision番号 | null |
| preventive_medication 実revision present / absent | 実observation ID | 当該投薬revision行ID | 当該投薬revision番号 | null |

表の実ID / 版は正BIGINT十進string、nullは明示null。inquiryのrevisionだけは純読取GET時のowner未初期化に限りmanifest.owner_versionとともにnull可であり、POST確立は初期化後の正owner版を要求する。inquiryのpresenceは現在owner入力状態を使い、保持中の過去行から存在を復活させない。scalarの型 / null / 空は第3節の保存正規化に従う。実revisionのpresenceはその固定行のpresence_stateと一致し、absent行でもID / 版をnullにしない。予防集合の未作成identityにはplaceholder itemを作らず、完全な空集合はitems=[]。親identityが実在してhead revisionを取得できない場合は未作成absentへ縮退せず取得失敗とする。未知key・重複・相互条件・同clinic / record / pet・全集合取得をtable-local検証関数とtyped decoder双方で検証する。この一意形のJCSだけをmanifest_digestとし、未掲載予防IDはその完全snapshot時点に存在しなかったことを表す。

fixture計画（NOT_RUN）：scalarの省略 / null / 空のfield別wire→storage→snapshot→hash一致、3 copy branchの未作成 / 実absent source拒否とclear / Undo分離、全kindの未作成 / 実present / 実absent manifest形状・順序・digest一致、0 / 空ID / 不正null組合せ / head取得失敗の拒否を確認する。

restore manifestはclosed {schema,record_id,scope_kind,baseline_ref,source_operation_key,copy_sources,targets}、schema='medical-record-input-restore-v1'。source_operation_keyはcopy_undoだけ元成功copy UUID、他はnull。copy_sourcesはhistory_copyだけ第18節の検証済typed sources（snapshot_sha256付き）を同じ安定順の非空配列で固定し、clear / undoではnullとする。3 branchのclosed union / FK / source hash一致を検査し、任意client snapshotを保存しない。targets各要素はclosed {kind,id,before_revision_id,after_revision_id,before_owner_version,after_owner_version,before_presence,after_presence,selected_entry_keys}。inquiry singletonのidはmedical_record_id、before_owner_version / after_owner_versionは不変snapshotのowner版string、before_revision_id / after_revision_idはnullとし、他kindは実ID・実revision IDを使ってowner版欄をnullとする。inquiryの旧版snapshotが存在する時だけbefore_owner_versionを設定し、未観測版は作らない。owner版参照は(clinic_id,record_id,owner_version)で第5.1節の実snapshotへ照合する。版はstring/null、存在状態はpresent/absent、entry keysはUUID昇順・重複なし。任意文章をmanifestへ追加しない。scalarの実復帰値は本人baseline内snapshot又は既存固定版参照、直接manifestに二重コピーしない。全targetのscope/current dependencyを同TXで検証し、部分成功を全成功にしない。

### 16.1 Baseline POST / GET / 再試行

純読取 GET /medical-records/{id}/input-baselines?scope_kind=inquiry|preventive_prescriptions|preventive_medications を定義する。scope_kindは必須で未知値/重複値を拒否。現在record view、trusted clinic/account/staff scopeを検査し、同一read snapshotで対象の全集合と現在版を取得する。成功200のclosed responseは {record_id,scope_kind,owner_version,manifest_digest,manifest} の5 fieldのみ。manifestは第16節の同じschema/並び順・complete=true・不変scalar表現、manifest_digestはそのUTF8 JCSのSHA256小文字64桁hex。inquiryのowner_versionは現在版、未初期化時はnullであり、494のacquire/take-overによる初期化前にはPOSTへ進めない。予防集合はowner_version=nullで既存head全集合digestを使う。

このGETは観測だけでbaseline_ref/operation_keyを発番せず、binding、baseline表、owner版、session、audit、receiptを作成・更新しない。GETで受け取ったsnapshotは保存基準として未確立。確立は下記POSTが同TXの現在fence/owner版/全集合digestを照合して成功した時だけ。不完全取得・失権・scope不一致・読取失敗では非2xxとし、成功manifest_digestやcomplete=trueを返さない。空集合が正当に完全取得できた場合と失敗を区別する。Cache-Control:no-store。読み取り後の状態変化はPOST側409で検出し、GET観測を無条件で保存しない。

POST /medical-records/{id}/input-baselines のrequestはclosed、全field必須：operation_key(UUID)、baseline_kind(tab_open|copy_before)、scope_kind上記3値、edit_session_ref(UUID)、edit_epoch(正十進string)、expected_owner_version(string|null)、expected_manifest_digest(64桁小文字hex)。初期GETで受領した完全な現在manifest digestを送る。本文snapshotをclientに送らせない。inquiryはowner_version必須、予防集合はnullと各既存head集合digestでCASを代替する。通常writeの親lock閉包を使い、snapshot変化はdependency-changed409で基準を作らない。

baseline作成は既存C06、owner_kind=medical_record、intent=medical_record.input_baseline.capture。認可入口view、未成功Applyはedit＋current held session/epoch＋scope/owner版/digest検査。既存494record anchorをClaimMetadataWriterで新binding時だけ保存、baseline本文はclaim metadataではない。494のanchorは結果を開示する元targetのみ、本人baselineは成功した入力基準とprivate scalarを保持する別責務。共有Execute内にExecuteをネストしない。

同key同Contentの成功再送は同baselineだけ、現在のsnapshotへ置換しない。expected_manifest_digestは復元先identityとして第18節payloadへ束縛し、単なる可変Preconditionsにはしない。同key別baseline_kind/scope/record/manifest digestはoperation-key-conflict。異keyは新しい明示的基準取得として別baselineになるが、FEはtab_openをその編集タブの開いた世代で1回だけ確立し、通信再試行で新keyへ変えない。copy_beforeは各明示copyごと1回。追加tab_session IDや別idempotency keyを作らない。基準取得未確定中は当該変更送信を停止。snapshot変化によるdigest変更は再確認後の新しい明示基準・新keyを必要とし、旧key/refの基準をすり替えない。新copy/タブ再開始を意図した場合のみ新key。単なる再fetchでtab_open refを更新しない。

POST成功responseは第18節C06閉shape。本文取得はGET /medical-records/{id}/input-baselines/{baseline_ref}、closed response {baseline_ref,record_id,baseline_kind,scope_kind,edit_epoch,owner_version,manifest_digest,manifest,created_at}。現在元record view＋同principalを必要とし、baseline_refは必ずclinic＋認証principalでqualified lookupし、他staff/別accountへscalarを公開しない。GETは基準作成/延命なし、Cache-Control:no-store。コピー元が別recordの場合は元recordの現viewをコピーApplyで別途再検査する。

全処理はDB-onlyで非同期jobを作らない。通信切断/202/503/500は既存GET /medical-records/{id}/operations/{operation_key}?intent=medical_record.input_baseline.captureで照会。確定未成功なら同key/Contentで現Preconditionsを再確認して再試行、unknownは停止、成功なら元baselineをGETする。本人draftにはrefと入力世代を暗号化して保持できるが、認証token/session権限証明を本文へ保存しない。baseline再取得だけで旧epochを現在権へ昇格しない。

## 17. 必須audit sinkとgrantの実route根拠

既存audit_logs（backend/internal/model/audit_log.go、DDL2597行以降）を同ambient TXで使う。本書で追加するaction/resourceラベルは以下に固定し、独立audit表を作らない。actor_type='staff'、actor_id=trusted staff、clinic_id=trusted selected clinic、created_at=server時刻、resource_idは記載対象ID。未知actorを0/systemへ代用しない。ip_address/user_agentをこのconsumerの新たな必須入力にしない。

- baseline：action='medical_record.input_baseline.capture'、resource='medical_record'、resource_id=record ID。
- 496通常/復元：action='medical_record.inquiry.save'、resource='medical_record'、resource_id=record ID。
- 497処方：action='preventive_prescription.create'又は'preventive_prescription.revise'、resource='preventive_prescription'、resource_id=prescription ID。
- 497投薬：action='preventive_medication.create'又は'preventive_medication.revise'、resource='preventive_medication'、resource_id=observation ID。

全metadataはclosed {schema,operation_id,operation_key,medical_record_id,baseline_ref,restore_kind,changed_fields}。schema='clinical-input-audit-v1'、operation ID/record IDは正十進string、baseline_ref UUID/null、restore_kind null又は上記3操作、changed_fieldsは許可field path昇順配列。old_value/new_valueはclosed {revision_id,revision,presence}又は未作成時null。inquiry scalar更新はrevision値にinquiry owner版、revision_idはintake版が存在する場合のみ、それ以外null。本文/薬量/患者氏名の無条件複製なし。差分本文は既存clinical immutable版/本人baselineへ参照で保持。

audit INSERT失敗は版/entry/provenance/restore manifest/成功receipt全rollback。成功no-opは同operation1件のaudit＋receipt、版は不変、replayでauditを増やさない。GET/プレビューは臨床変更auditを作らない。best-effortゴルーチン監査で代用しない。

実route根拠：19855942のroutes_records.goはPATCH /:id/inquiriesにResourceMedicalRecords edit、POST /:id/treatmentsとPOST /:id/checkupsに同create、各PATCHに同edit、GET record/childに同view。予防処方createへcreate、改訂へeditの対応はこの既存resource/actionを再利用する新consumer選定で、現予防route実装済みの証拠ではない。D21 OwnerReport.tsxはMedicalRecords viewとOwners viewを使う。閲覧grantをwriteへ昇格しない。

## 18. C06成功refs/versionsと正確なhash projection

公開ResultResponseの成功時required keysはstatus,operation_key,result_refs,committed_versions,recorded_atだけ。result_refs/versionsはmap[string]string、余分keyなし。非成功時はresult_refs/versions/recorded_atを出さない。recorded_atは保存済みserver timestamp。同scope replayは同じmap。

この新consumerの成功responseには専用adapter / response DTOを定義し、success時requiredのresult_refs / committed_versions / recorded_atを明示出力する。baseline成功のcommitted_versionsは空object `{}` を保証し、既存generic DTOのomitemptyや空map→nil変換は使用しない。他consumerのDTOやshared converterを一括変更しない。非成功はstatus / operation_keyだけの分離DTOを使い、空成功mapを成功欠落・nullに変えない。これは採択技術契約であり、adapter / DTO実装済みとは記載しない。

- baseline result_refs={medical_record_id,baseline_ref,manifest_digest}、committed_versions={}。
- inquiry result_refsはmedical_record_idを必須とし、inquiryが実在又は今回新規作成された時だけinquiry_id、今回又は現在のintakeが存在するときだけintake_revision_id、restore metadataを作った場合だけrestore_operation_idを加える。committed_versions={inquiry}にintake版が存在するときだけinitial_intake。
- prescription result_refs={medical_record_id,prescription_id,prescription_revision_id}にrestore時だけrestore_operation_id、committed_versions={preventive_prescription}。
- medication result_refs={medical_record_id,prescription_id,prescription_revision_id,observation_id,medication_revision_id}にrestore時だけrestore_operation_id、committed_versions={preventive_medication}。参照処方revisionはresult_refに固定し、処方が今回改訂されたようなcommitted_versionsを返さない。

intakeの「存在」はrevision行の実在を指し、臨床payloadの有無とは区別する。presence_state='absent'でもclear / Undoがappendした実revision行のIDをintake_revision_id、当該revisionをcommitted_versions.initial_intakeへ返す。成功replayと固定版readerも同一ID / 版を使い、実revision行がない場合だけこれらrefs / 版を省略する。no-opで現在のabsent版が実在する場合も同じ既存行の規則を使い、新revisionを作らない。

fixture計画（NOT_RUN）：baseline成功responseは必須keyを保持し、`committed_versions: {}` を明示出力する。省略・null・失敗shapeへの変換を拒否し、非成功はstatus / operation_keyのみであることを確認する。

optionalはこの有限列挙条件のみ。null/空文字/0のIDをmapへ入れない。inquiry no-opで未作成ならinquiry_id keyも省略し、body省略が空inquiry自動生成を起こさない。保存後はread projectionを取り、receiptを本文代用にしない。

Content.Fieldsのexact keysは全consumerでschema,medical_record_id,target_kind,target_ref,payload_sha256,source_sha256,restore_sha256。schema='clinical-input-command-v1'。target_kindはinquiry/input_baseline/preventive_prescription/preventive_medication。target_refはinquiryでは未作成/作成済みにかかわらずmedical_record_idの正十進stringを常に使う。その他既存targetは正十進string、新create/baselineはoperation_key UUID（発番前のidentity）。source/hash/restoreが無い場合もそれぞれJCS nullのSHA256を使い、keyを省略しない。Contentにactor/session/epoch/current expected_versionは含めず既存Scope/Preconditionsで扱う。

payload_sha256=SHA256(UTF8(JCS(payload)))の小文字hex。payloadはintent別closed typed DTO：baselineは{baseline_kind,scope_kind,expected_manifest_digest}、inquiryは{chief_complaint,chief_complaint_type_id,notes,initial_intake}の第3節正規化後のfieldのみ（textのnullは保持として省略、typeのnullは解除として保持、空textは値として保持、aliasはcanonical typeへ正規化）、prescriptionは第8節public content＋source_treatment_id（copied_history_record_id / copied_history_revision / history_snapshotは除外）、medicationは第8節content＋prescription_revision_id＋display_year（create/改訂とも）。server-only recorded/actor/operation列を混ぜない。initial_intake entriesはentry_key順、provenanceはsource ordinal順でcanonical化し、画面配列順を意味と取り違えない。

source_sha256はcopy時のclosed {sources:[typed source]}、その他null。typed sourceは第11節copy_sourcesをserverで検証しsnapshot_sha256を加えた形だけとする。initial_intake branchのexact keysはkind,record_id,intake_revision_id,entry_keys,snapshot_sha256、inquiry_snapshot branchはkind,record_id,baseline_ref,owner_version,snapshot_sha256、inquiry_owner_snapshot branchはkind,record_id,owner_version,snapshot_sha256。汎用revision_idへ両identityを押し込まない。sourcesはkind、record_id、branch別固定ref、entry_keysのtupleで安定順、entry keysは昇順。snapshot_sha256は第20.3節のbranch別closed projectionで計算し、inquiry_owner_snapshotは第5.1節の正確なprojectionを使う。Applyで同じrecord / owner版 / snapshot digestを再照合し、3 branchのrefをrestore manifest / 処方history_sourcesへ一貫保存する。restore_sha256は第5節と同一のclosed restore_context {kind,baseline_ref,source_operation_key,selected_entry_keys}又は省略時JCS null。selected IDs/refsはすべてcanonical型、clientのhashだけは信用しない。

typed constructorは後記の共有VerifiedCommand限定拡張によりContentとRequest内deep-copied commandを一度に作り、ClaimMetadataWriter/Apply双方で上記7 keysと3hashを再計算一致確認。field map取り出しと別captured未検証payloadで書かない。current epoch/CASが変わっても同Content再試行可、copy対象版/選択/本文/復元基準が変われば新keyが必要。

## 19. ContextGateの限定引渡し

PR #784のroot採択R4契約に合わせ、共有runnerへoptional ContextGate interfaceの GateContext(context.Context) (context.Context,error) を追加するsubsetを再利用する。公開初期head509fe2c73da2effc1f0f9a388b292896af814988には未反映で、設計採択とruntime実装を区別する。現contract.goのGate(ctx) errorだけではreservation gateのcontext markerがApplyへ伝播した証拠にならない。

Runnerの最外WithTx先頭でContextGateを実装するparticipantならGateContextだけを呼び、なければ旧Gateを呼ぶ。returned contextはnon-nilかつ同じambient TX pointerであることを検査する。そのcontextをbinding lock、元anchor解決/現在認可、成功receipt/disclosure、savepoint、Apply、audit、InsertResult、同TX unknown記録まで一貫伝播する。rollback後の別TXで再利用しない。

496/497の各write/baseline participantは、既存D03 lifecycleがreservation graph gateを必要とする場合にGateContextを実装し、AcquireReservationGraphWriteGateから返るcontextをそのまま返す。Apply/repositoryの既存AssertReservationGraphWriteGateを維持し、参加する全writerも同じambient chainへ統合する。gate不要の独立readは新たに取得しない。旧interface破壊、ctxをmutable fieldへ捕捉、marker自作、DetachTx、後段再gateによる順序修復は禁止。

必要共有subsetは489/494へ依頼し、実装ownerが独立でrunnerを並行変更しない。ContextGate未統合のconsumerはmock/pure adapterまでに留め、単独文書公開でruntime保存可能とは扱わない。

監査sinkとoperation現interfaceの固定根拠：
https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/backend/internal/model/audit_log.go
https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/architecture/erd-detail/audit_logs.md
https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/backend/internal/operation/contract.go


## 20. 統合レビューで採択した限定補正

### 20.1 問診current presenceと公開copy provenance
inquiriesへpresence_state text NOT NULL CHECK IN('present','absent')を追加する。既存実行の存在はpresentとして技術移行するが未作成行は作らない。検証済復元でbaselineが不存在なら既存行のpresenceだけabsentにし、非NULLscalarは保持したまま臨床current readでは返さない。同owner CAS・owner snapshot・必須audit・receiptと同TX。再入力は同実inquiry IDをpresentへ戻し、新rowを重複生成しない。詳細/一覧summary/現scalar読取/コピー元選択/自動保存全callerがpresenceを尊重するまでactivationしない。履歴snapshotは不変。過去rowの物理削除で代用しない。

inquiry_owner_snapshotsへinput_operation_kind（observed_baseline/ordinary_save/history_copy/tab_open_clear/copy_undo）とcopy_source_refs jsonbを追加する。history_copyだけ非空配列、それ以外空配列。これはclinic-scoped immutable clinical projectionで、private baseline本文/restore manifest RLSは変更しない。copy_source_refsはsource順を保持し、closed union: initial_intake={kind,record_id,intake_revision_id,entry_keys,snapshot_sha256}; inquiry_owner_snapshot={kind,record_id,owner_version,snapshot_sha256}; inquiry_snapshot={kind,record_id,owner_version,snapshot_sha256}。最後のbranchはprivate baseline UUID/本文を開示しない。本人baselineが参照する臨床版を捏造せず、source_kindで本人baseline由来と表示する。record viewの別staffもこのprojectionからcopyを識別でき、private restore権限は得ない。固定DTOとsnapshot hashへ両fieldを同じ意味で追加する。

### 20.2 処方sourceのentry配列
§11のpreventive_prescription_history_sources.source_entry_key単列は削除し、source_ordinalをsource group identityとして保持する。新child preventive_prescription_history_source_entriesはclinic_id,prescription_revision_id,source_ordinal,entry_ordinal（0始まり）,source_intake_revision_id,source_entry_keyを全NOT NULL、PK(clinic_id,prescription_revision_id,source_ordinal,entry_ordinal)、UNIQUE(clinic_id,prescription_revision_id,source_ordinal,source_entry_key)。親groupへの完全FKと(clinic_id,source_intake_revision_id,source_entry_key)→initial_intake_entriesの完全FK、RESTRICT。child intake revisionは親groupの同列と一致をconstraint triggerで検証。initial_intake branchだけ選択entry全件をentry_key昇順の連番で同TX INSERTし、他branchはchild0件。group単位で配列をlosslessに再構成し、同revisionの異subsetを別source ordinalのまま保持する。immutable trigger、clinic RLS、現在元record viewは親と同じ。新保存でchild欠落/飛び番/重複/余剰は失敗rollback。

### 20.3 コピー元hashのexact projection
すべてUTF8 JCSのSHA256小文字64hex。ID/版は正十進string、UUID小文字、nullは明示、日付YYYY-MM-DD、時刻UTC RFC3339Nano。未掲載keyは拒否する。
initial_intake branchはclosed {schema,record_id,intake_revision_id,presence_state,revision_source,provenance_basis,selected_entry_keys,entries,entry_provenance}、schema='initial-intake-copy-v1'、presence_state='present'。revision_source={source_kind,source_reference,heard_at,copied_from_revision_id}は保存値/null。provenance_basisはentry_mapping/legacy_revision_level。selected_entry_keysは非空UUID昇順重複なし、entriesはその完全一致集合だけentry_key順。entryはclosed {entry_key,field_key,field_label,answer_state,detail,drug_description,symptom_description,reported_result,time_precision,event_year,event_month,event_date,time_description,reference_exam_revision_id}、型/nullは§4/保存schemaどおり。entry_provenanceはselected entryの全mappingをentry_key/source_ordinal順で含み、closed {entry_key,source_ordinal,source_kind,source_reference,source_record_id,source_intake_revision_id,source_entry_key,heard_at,copied_at}。legacy_revision_levelでは空配列としrevision_sourceの実在値のみを使い、架空mappingなし。
inquiry_snapshot branchはclosed {schema,record_id,baseline_ref,owner_version,presence_state,scalar_snapshot}、schema='inquiry-baseline-copy-v1'。保存済本人baselineのinquiry itemからだけ取得、present必須。scalar_snapshotは主訴/主訴type/notesと下記当時label契約のclosed形。baseline全体digestで予防集合等の無関係値を混ぜず、選択scalarだけをcommitする。inquiry_owner_snapshotは§5.1の固定projectionに従う。copy source権限と実在snapshotをApplyでも照合し、hashだけを信頼しない。

### 20.4 責任職員scopeと当時主訴名
新規又は責任職員変更時は、同clinicのstaff_clinic_assignments.deleted_at IS NULLと対象staff実在をtrusted resolverで検証し、既存staff/assignment共通lock順に参加してcommitまで保護する。IDだけの単列FKで他院staffを選定しない。未変更の歴史担当は保持し、退職/assignment退役で過去参照を強制解除しない。予約用nomination_paused/休業/能力gateを処方の新制限にしない。
inquiry owner snapshotとbaseline scalar_snapshotへchief_complaint_type_name（text|null）とchief_complaint_type_name_basis（recorded/legacy_unknown/not_selected）を追加し、fixedDTO/hash/source copyにも含む。新規選択又は変更時だけscope検証済masterから名前を固定、同ID未変更なら既存固定名を保持する。選択なしはnot_selectedかつname=NULL。旧版に名前がない場合legacy_unknownかつname=NULLで、現在master名を過去値として埋めない。observed_baselineの観測値を過去入力時点の名前と主張しない。既存scalar numeric ID wireとcategory aliasは保持する。

## 統合レビュー共通補正：shared Request の限定 immutable payload

現operation.Participantはreq.Content.Fields由来のwriteのみ許し、private captured commandを許さない。このconsumerには全文を短いContent mapへ格納できないため、共有Requestへoptional VerifiedCommandを追加する限定deltaを必要subsetとして選定する。VerifiedCommandはoperation package所有のimmutable値で、constructorでschemaとcanonical command bytesをdeep-copyし、schema別pure projectorが求めたContent全keysとの一致を検査する。getterはdetached copyのみを返す。Runnerはclaim/fingerprint前とApply直前に同じrequest bytesから再検証し、Applyはreq.VerifiedCommandだけからtyped入力を再構成する。participantへのDTO capture、mutable pointer、別parse結果によるwriteは禁止。CASはPreconditionsのままである。

payloadはrequest memoryのみ、binding/receipt/logへ保存しない。既存consumerはVerifiedCommandなしのContent-only方式を維持する。HMAC/各digestのschemaと鍵版は本文既定どおりで新正規化をしない。489で共通contract/runnerの限定interfaceを一箇所だけ実装するまで、このconsumerを現C06のまま実装可能と扱わない。これは親による技術選定であり、本人がAPI/物理方式を直接承認したとの記載ではない。実装/試験NOT_RUN。

### 20.5 表示年訂正の版別保存

表示年は医療実施日/医療stateではなく専用年次表の帰属年であり、誤入力訂正を既存の投薬改訂で行う技術方式を採択する。本人がAPI/物理方式を直接承認したという意味ではない。

preventive_medication_revisionsへeffective_display_year integer NULL CHECK (effective_display_year BETWEEN 1 AND 9999) とdisplay_year_basis text NOT NULL CHECK IN('recorded','legacy_unknown')を追加する。CHECK ((display_year_basis='recorded' AND effective_display_year IS NOT NULL) OR (display_year_basis='legacy_unknown' AND effective_display_year IS NULL)) を明示する。旧revisionはlegacy_unknown/NULLのままで、親の現在年を過去各版へbackfillしない。新通常create/改訂は必須display_yearをrecordedとして固定する。親observation.display_yearはidentity作成時の元値として保持し、訂正で上書きしない。

現在年次表はhead revisionにrecorded年があればそれを採用する。旧headがlegacy_unknownの場合だけ、実在する親年をlegacy_current_observedのread basisで表示し、版別確定値とはしない。固定revision readerは当該版のeffective_display_yearとbasisだけを返し、legacy_unknownでは年NULL、現在親年へのfallbackなし。出力DTOへdisplay_year（integer|null）とdisplay_year_basis（recorded/legacy_unknown/legacy_current_observed。ただし固定版に最後の値は禁止）を明示する。年次検索はheadと年basisを同snapshotで取得し、欠落を別年や0へ移さない。

訂正は既存recordのcurrent session/epochとobservation expected_revisionのCASを通し、新revision・必須audit・C06 receiptを同TXで確定する。同key replayは元版、年変更だけでも新revisionを作り、実施日/state/処方参照を自動変更しない。年変更を含む正規化payloadを既定Content digestへ束縛する。表示年訂正のため別observationを作る救済は禁止する。

検証済clear/Undoはbaseline manifestで固定した当該revisionのeffective年/basisをそのまま新復元revisionへ保存する。元版legacy_unknownはそのまま保持し、clientのdisplay_yearで復元元を補完しない。restore_contextありではdisplay_yearは元版の年又はNULLを送る復元専用variantとし、通常writeの必須1..9999 variantとはdiscriminated decoderで分離し、不一致を拒否する。absenceは入力状態であり、表示年の推論を起こさない。copy/帳票readerも固定版のこの2列を使う。

NOT_RUN fixture：誤年の改訂後にcurrent年次表だけが新年へ移る、旧固定revisionは元年又はlegacy_unknownを保持、同key応答断で二重改訂なし、年度変更競合409、年だけの変更で実施日/state不変、legacy年不明を過去確定値へ変換しない。
