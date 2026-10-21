# EMR-368 保存時氏名の採択技術契約 R4

2026-10-21。root採択済みR4・第10節の接続契約を統合した公開版。調査基準は2026-10-21 04:07 UTC、固定SHAは第2節のとおり。物理氏名event＋projection、正規化差分によるno-op、新方式空createのみunentered、旧legacy_unknown保持、単一staff FKと現在assignment認可、staff→lifecycleのlock順、成功replay不変更を採択する。設計採択であり、runtime実装・DDL適用・試験合格・運用開始を意味しない。今回の公開は本書1文書だけで、コード・API・DB・設定・実施・マージを含まない。試験は NOT_RUN。票本文と全8コメントの再取得は提供された調査履歴であり、最新コメントのPR729文書等merge通知を、本方式の既実装の証拠にしない。

## 1. 既決業務と採択技術方式

保存時点の業務要件は確定済み。D03 medical-record-form.md L536–548 / L584–592 に、初回実入力の正常保存時のID/氏名を不変保持、最後の成功実編集時のID/氏名を表示、改名だけでは変えず、改名後の実編集では最終名を更新するとある。担当ID変更だけで再snapshotする方式はこのACに反する。同じ職員IDでも別の実編集成功ならその保存時名が必要。担当医変更を行った場合も保存actorと新担当医は別であり、担当医名を記録者名にしない。

D03 lifecycle L223–227 は本体・ヘッダー・問診・治療プラン・治療/処方・バイタル・関連子入力・追記・許可された削除/過去入力の成功編集ログを要求。氏名表示を理由に編集可否や目的別訂正条件を変更しない。受付/空カルテ生成、読取、印刷、編集権取得は実入力としない。

調査時点の票では物理snapshotか厳密immutable staff revisionかが未選定だったが、本契約では操作単位の物理氏名event＋projectionに固定する。staff_revisions検索結果0のみで全repo不存在とは断言しないが、確認したStaffモデルにはname、UpdatedAt等の可変列しかなく、staff_repository.Updateは既存rowを更新する。既存不変staff revisionへの保存操作単位のFKを採用した正本は確認できなかった。これは調査根拠であり、現在の方式選定を未決に戻す意味ではない。

## 2. 現実装で確認した具体不足

固定参照 `5f6ac6c03500e63d18165b14c22631e586f044ed`。以下は固定SHAでの観察であり、採択方式の実装完了を示さない。

- model/medical_record.go：entered_by IDあり、初回名/最終名/最終actor/初回未入力判定列なし。現在Statusモデル列は既に削除済みなので古いコメントのモデルStatus残存を再掲しない。一方repository updateには旧status条件が残り、別ownerのlifecycle整合を必要subsetとして扱う。
- medical_record_repository.go L189–195,309：EnteredByStaff preloadは現在staff row、soft-delete除外。改名後の当時名や削除後の名を保証しない。
- medical_record_crud.go L163–178,266：create/updateの監査はpost-commit/best-effort。L332–378 推奨理由更新はactor=nilでbest-effort。削除L322–325もactor=nil。氏名projectionだけを別TXで足してはいけない。
- inquiry_service.go：UpsertInquiryInputにactorがなく、SaveByMedicalRecordID保存のみ。494/496の新writer接続時に同じ成功単位へ参加が必要。
- clinical_plan_service.go L318–378：ActorID必須・同TX auditUpdateTxの既存経路あり。ただし保存時氏名snapshotは未確認で、既存監査を二重化せず同ambient TXへ本体帰属を接続する。
- medical_record_entered_by_actor.go：現在staff/所属をSHARE lockし認証actor検証する既存create guardあり。主所属clinic一致ではなくassignmentまたは検証済system-adminを扱う。snapshotにもこの認可意味を継承し、誤ったstaff(id,clinic_id)複合FKを導入しない。

## 3. 採択方式：操作単位の物理氏名snapshot＋本体projection

比較したstaff name履歴の全rename経路/不変revisionを新設して各clinical保存から厳密に参照する方式は、職員マスタ・取込・修復全writerへ依存し、本票だけでは既存採用根拠がないため採用しない。採択方式はカルテ実編集の成功操作に名前をsnapshotしたappend-only eventを保持し、初回/最終を同TXで投影する。現在名JOIN不要、失敗・再送・改名・退職に対する証明が局所で閉じる。既存audit_logsを代替/削除しない。

新 medical_record_author_events：id bigint PK、clinic_id、medical_record_id、operation_id、actor_staff_id、actor_name_snapshot text NOT NULL、recorded_at timestamptz、edit_kind text、changes jsonb NOT NULL（closed arrayの各要素owner_kind/owner_ref/committed_owner_version/change_kind、後述の安定順）。UNIQUE(clinic_id,medical_record_id,operation_id)。FK(medical_record_id,clinic_id)→medical_records(id,clinic_id)、FK(operation_id,clinic_id)→operation_bindings(id,clinic_id)、actor FK→staffs(id)単列RESTRICT。氏名eventはUPDATE/DELETE拒否、runtime INSERT/SELECTのみ、既存clinic RLS。1操作で複数子を保存しても同recordの帰属eventは1件。複数recordならrecord別1件。氏名をclient body/doctor/現在のresponseから受け取らない。

新 medical_record_author_projections：PK(clinic_id,medical_record_id)、first_state enum(unentered,recorded,legacy_unknown)、first_event_id nullable、last_event_id nullable。各event FKは同clinic/recordを含む複合FKとし、event側UNIQUE(id,clinic_id,medical_record_id)を置く。recordedだけfirst_event必須、unentered/legacy_unknownではfirst_eventNULL。projectionはrecord lock下で同TXのみ更新。旧entered_byはその既存由来のIDとして保持し、NULL判定だけで新初回を決めない。正本の初回actorはeventから返す。状態相互制約の詳細は第10.4節に固定する。

新規の空カルテ生成時にだけ、同create TXでunentered projectionを明示作成する。既存rowのmigrationはlegacy_unknown、projection欠落もlegacy_unknownとして読む。既存entered_byがあっても当時名証拠なしならknown legacy ID＋name不明。現在staff名のbackfillは禁止。確実な旧保存操作・不変名の証拠が存在する別importは別途producer契約で扱い、現行日時から推測しない。

最初の成功実編集でunentered→recorded＋first_event、常にlast_eventを同eventへ。legacy_unknownの初回は以後もunknownのまま、成功実編集後のlast_eventだけ記録。別actorも同actor改名後も同じ規則。eventのserver名を読み取ったstaff rowはcommitまでSHARE lockし、同時renameとの順序を固定する。氏名取得/認可失敗時に旧名やunknownを入れて成功させない。新スタッフ保存でnameNULLを許す意味ではない。

## 4. 成功単位・再送・読取

帰属対象consumerの選定順序は、既存必要graph gate→binding/元targetの現在開示認可と保存結果分岐→未成功Applyに限るactor staff・必要assignment/account SHARE→既存lifecycle→edit session/epoch→record→対象owner lockとする。第10.3節のactor先行順へ統一し、旧案のlifecycle後actor取得は採用しない。staff rename側とのlock閉包は実装ownerが既存guardと合わせる必要subsetとして明示。対象の実差分とactor名を確定後、業務版・帰属event・projection・必要な既存監査・C06成功receiptを同ambient TXで保存する。どれか失敗なら全rollback。監査のpost-commitへの退避禁止。

C06同key成功replayは元event/receiptを参照し、氏名を再取得/last更新しない。成功replayへ新しいedit権・現在在職・氏名取得を追加要求しない。競合・fence喪失・不成立ではeventなし。結果不明は既存lookup、別keyで復活させない。operation_idなし旧writerに架空operationを足して完成扱いにせず、既存C06統合ownerへ必要subsetとして渡す。

採択済みの技術上のno-op判定はserverの正規化済み業務payloadと現在保存値の同等性で行う。真の無変更自動保存、版読取、baseline作成だけは実編集イベントを作らず、projectionも変更しない。revisionが作られたことだけで実編集とみなさず、clear/Undoでも有効な状態が変わった成功保存なら記録する。このno-op境界は「実編集」ACを実現する採択技術方式であり、新たな臨床内容や編集許可を追加しない。

GET既存medical-record detailに追加 author_attribution object：first {state,staff_id,name_at_save,saved_at,operation_key}、last同形。stateはrecorded/unentered/legacy_unknown。recordedはID/名/時刻の保存event根拠必須、legacy_unknownは確実な旧IDのみnullableで保持。取得障害は通常非2xx/既存error、legacy_unknownへ落として成功応答にしない。新refは正規形十進string、既存entered_byの公開型はこの作業で破壊しない。現在Staffの氏名や在職状態をこの欄へ混ぜない。親カルテview認可内の履歴snapshotのみ公開し、staff profile全体の閲覧権を広げない。FE MedicalRecordStickyHeader.tsxの担当者欄とは独立した初回/最終表示、詳細query再取得で更新し楽観的actor表示はしない。closed DTOの詳細は第10.4節へ従う。

## 5. writer対応表と実装引継ぎ

確認したroutes_records.goの登録と現保存経路を開始点とし、各owner全writerを同一境界へ接続する。下表は調査/接続計画であり全writer実装確認済み表ではない。

| 保存群 | 実入口/経路 | 本票の接続点・境界 |
|---|---|---|
| 本体/ヘッダー | POST/PATCH /medical-records、medical_record_crud.go | 空作成はunentered、実入力を伴う保存のみ帰属event。create監査post-commitから必要監査同TXへ |
| 推奨理由 | PATCH /medical-records/:id/recommendation-reason | actor=nil解消、差分成功の同TXへ |
| 問診 | PATCH /:id/inquiries、inquiry_service.Save | 494/496 immutable participantの成功内に接続、LINE原文readerは除外 |
| 診察/診断/方針 | PATCH/DELETE /:id/clinical-plan、clinical_plan_service | 既存fail-closed監査を保持し同TX帰属追加、GetOrCreateのみは実入力としない |
| バイタル/治療 | /:id/vitals POST/PATCH/DELETE、/:id/treatments POST/PATCH/DELETE/PUT | 各子writerの実差分成功、bulkは同record操作1event。担当医でなく認証actor |
| 検査/健診/接種 | /examinations CRUD・items置換・unconfirm、/:id/checkups CRUD・field-results、/vaccinations CRUD | 記録に属するスタッフ成功編集のみ。機器受信や独立記録を無条件に本体staff最終へ昇格しない。D07/D08目的別訂正ownerへ必要hook引渡し |
| 予防処方/投薬 | 497の新recordbound write | 保存自体が実編集なら同TX。recordless readと固定帳票GET除外 |
| 追記/画像/許可削除 | /:id/addenda POST、images upload/delete、本体DELETE | 各既存grant/監査/履歴保持を維持し帰属を接続。326追記者表示を再実装せず、空作成を初回にしない |
| clear/copy/Undo/別内部入口 | 494/496/各owner manifest参加writer | 元保存版を保持、現在actorと今回実差分のみ。同じ医療事実を再生して元actorへ偽装しない |
| 取込/システム | csvimport/予約自動作成/機器連携 | 初回履歴証拠を捏造しない。現在スタッフ名で旧補完なし。systemをstaff0へ変換しない |

## 6. 明示的に残す確認境界

- 所属や担当IDの変更時だけsnapshotする方式は棄却し、本人への再質問不要。
- recordの実編集として分類される全子writerのhook位置とC06全参加閉包は実装準備の残り。上記入口表だけで全保存経路確認済みとはしない。
- systemの仮想的な本体変更を本人質問へ追加しない。追加確認した実在経路の対象外境界は第8節のとおり。現時点で本人回答を要する具体的system表示案件は挙げない。
- no-op判定の上記技術方式は採択済み。氏名時点や新しい臨床事実の判断待ちへ戻さない。
- 物理差分は446/schema ownerの既存受付へ提出、001凍結/追加migrationの本人決定を維持。別実装PRの既存ownerへhookを引渡し、直接変更しない。

## 7. 固定fixture設計（実行なし）

A初回実入力成功→初回/最終A、A改名のみ→旧表示保持、同A再編集成功→初回旧名/最終新名、B編集→初回A/最終B、担当医変更actorB→新担当者名を帰属にしない、空自動create→unentered、旧NULL＋後日B編集→first unknown/lastB、既知旧ID＋名欠落→IDのみ保持、name取得失敗/業務失敗/必須監査失敗/帰属INSERT失敗→全部不成立、競合敗者/同key再送→帰属不変、retire/softdelete後→保存名保持、他clinic閲覧拒否、無変更save→event/projection不変、単一bulk→1event、システム0捏造なし。第10節のactor先行/rename両順、成功replayの現在view・氏名再読なし、projection各状態の不正遷移拒否も作成対象とし、実行はすべてNOT_RUN。

## 根拠

- [EMR-368](https://app.plane.so/baritechllc/browse/EMR-368/)
- [保存時氏名の業務正本](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/docs/spec/domains/medical-records/pages/medical-record-form.md#medical-record-main-author-name-display)
- [入力者・編集ログの業務正本](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/docs/spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-authorized-editor-audit)
- [Staffモデル](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/model/staff.go)
- [staff repository](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/staff/staff_repository.go)
- [カルテCRUD](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/medicalrecord/medical_record_crud.go)
- [認証actor guard](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/medicalrecord/medical_record_entered_by_actor.go)
- [既存route登録](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/medicalrecord/routes_records.go)

## 8. 追加調査：system経路と最小scope閉包

実在コード確認（同5f6ac6c0）：medical_record_auto_create.go L77–93は予約由来空作成後にCreateSubRecordsInput{}を渡す。medical_record_subrecords.goのupsertInquirySubRecordは入力なしなら保存せず、ensureClinicalPlanSubRecordは空row確保のみで、臨床fieldがない場合更新しない。この内部空生成/修復は実入力者への帰属対象外。自動生成した時点でunenteredを確立し、初回職員の実入力成功まではfirst/last staffを持たせない。既存rowに対する空修復はlegacy_unknownをunenteredに変えない。

同じCreateSubRecordsはHTTP新規作成の有値問診/診断にも呼ばれる。こちらはスタッフの実入力であり対象。現handlerはroot Create成功後にCreateSubRecordsを呼ぶため、root+clinical child+author event+必要監査の一成功単位への統合が必要。空作成と有値作成を「同じserviceだから全部system/全部入力」に分類しない。

lab_device_exam_persist.go persistMappedLabExams L289以降はLabExamPersistInputにPetID/JobID/検査値を渡すがMedicalRecordIDを渡さない。したがって確認した自動機器受信の永続化はペット単位の検査であり、カルテ本体staff最終へ自動帰属しない。lab_import_examination_service.goにはMedicalRecordIDを受理する汎用型はあるが、型があることだけで本番systemのカルテ編集が実在すると判断しない。fixture/import経路は出所・job監査のまま、現在実行スタッフの氏名を過去入力者にbackfillしない。後日のスタッフによる明示カルテ紐付け/検査訂正は既存D07の認可・actor付き操作として別扱い。独立検査の受信を契機に自動first/last更新しない。

最小の帰属scopeは「対象medical_recordに対する、現在認証スタッフの許可された実入力/実編集として成功した業務操作」。全DB更新ではない。root更新・有値root+子作成・子単独更新/許可削除・追記・実差分を持つclear/copy/undoが対象。医師確認/終了/訂正が同じrecordに対するスタッフ業務編集として既存ownerで保存される場合は、その既存操作を1単位に対応付ける。確認印や終了eventだけから過去の初回実入力者を遡及確定しない。

対象外：GET/一覧/印刷/版表示/キャッシュ、現在staff改名、account/権限/所属変更、予約空自動生成、空子row確保、baseline snapshot作成、編集権acquire/take-over/release、単なるversion/digest/監査/receipt/集計派生値更新、機器受信/移行原文の保存。対象操作内で派生処理が複数回書いても追加帰属eventにしない。実入力元の履歴をコピーしても今回はcopy実行actorであり元作者を偽装しない。

操作単位の最小一覧：

1. root-onlyの有効なheader/recommendation差分：1 record × 1成功operation。
2. root+問診+診断等の有値作成：共通外側成功operationで1event、child別に増殖させない。
3. inquiry/clinical plan/vital/treatment/checkup/vaccination/examination等の単独子編集：保存対象からserverで元recordを解決し、同operationで1event。recordlessは対象外。
4. bulk/clear/copy/undo：既存ownerが定めた原子的操作単位で各record1event、内部N子でもNeventにしない。別request/別成功operationは別event。
5. addenda/画像/許可削除/専用訂正：既存の本人操作・成功単位を保持し、独立履歴を再実装しない。内容未確認の任意バックグラウンド処理はここへ追加しない。

## 9. 移行不変条件と既存履歴の再利用

- 切替前の既存rowはentered_byのNULL/非NULLに関係なく証拠未確認ならlegacy_unknown。作成日が新しい、子rowが空、受付がある等ではunenteredと推定しない。
- unenteredは新方式producerが新しい空recordを生成した同TXで明示したものだけ。default unenteredを既存全行に適用しない。新方式未統合producerが作ったrowはprojection欠落→legacy_unknownで止め、次編集者を初回にしない。
- 新規有値保存は作成と同TXにfirst/lastをrecordedまで進める。root空作成だけ成功・子実入力失敗ならfirst/lastを進めない。一操作として成功を返す範囲は同TXへ統合する。
- recorded→unentered/legacy_unknownへの通常writeを許さず、first_eventは一度確立した後不変。旧unknownの後日編集はfirst unknown保持＋lastだけ確定。旧既知IDは消さず別legacy根拠として返す。
- 旧staff IDから現在nameを取得して既存snapshotへ移行しない。移行時の最終者もupdated_atや最大IDから推測しない。

再利用できるもの：494/各consumerのoperation binding/元target/receipt/idempotence、子領域の既存immutable revision IDとowner版、既存必須audit_logs、326 addenda本文履歴。新author eventは本文の第二正本を持たず、それらの成功operationと版を参照して氏名証拠だけを補う。

再利用だけでは不足するもの：medical_record_consultation_eventsのevent_kindはended/end_undone限定で、全編集の履歴に転用不可。recorded_byはあるが氏名snapshotなし。examination_revisionsのdisplay_snapshotはDoctorName等の表示masterでありactor_nameではなく、ActorIDとDoctorIDを混同できない。MedicalRecordAddendumはAuthorUserID＋現在AuthorStaff参照で、氏名保存方式の全体証明にはならない。audit_logsは全writerの同TX/fail-closed参加が未完であり、最新監査timestamp検索だけで本体first/lastを導出しない。

既存immutable revisionのactor ID＋厳密に同操作の保存時氏名が検証できた範囲では、その証拠参照を再利用できるが、採択済みevent＋projectionを別方式へ置き換える意味ではない。今回読んだ表には全writer共通の氏名根拠がない。このため採択snapshot方式は「全revision/bodyを新設」でなく、操作単位の狭い帰属証拠＋projectionに限定する。以下第10節の接続契約を実装前引継ぎとする。本人へ新しいsystem業務質問を要求する状態ではない。

追加根拠（同固定SHA）：

- [子保存](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/medicalrecord/medical_record_subrecords.go)、[空自動生成](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/medicalrecord/medical_record_auto_create.go)
- [機器検査永続化](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/medicalrecord/lab_device_exam_persist.go)、[検査取込service](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/medicalrecord/lab_import_examination_service.go)
- [検査revisionモデル](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/model/examination_revision.go)、[追記モデル](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/backend/internal/model/medical_record_addendum.go)
- [診療終了event](https://github.com/MinoruSoga/AnimalEkarte/blob/5f6ac6c03500e63d18165b14c22631e586f044ed/docs/architecture/erd-detail/medical_record_consultation_events.md)

## 10. 選定接続契約：狭い同TX participant

新privateファイルは backend/internal/medicalrecord/medical_record_author_attribution.go（resolver/RecordSuccessfulEdit）、medical_record_author_repository.go（2表）、medical_record_author_response.go（projection）とする。新汎用監査framework/別worker/outboxを作らない。これらは後続実装の配置契約であり、本PRでファイルを作成しない。

RecordSuccessfulEdit(ctx, immutable attribution input) はambient TX必須。入力はserver解決したclinic/record、trusted actor、C06 operation_id、同operation内の正規化済み変更集合 changes[]（各要素owner_kind/owner_ref/committed_owner_version/change_kind）、正規化後の実差分判定だけで、body指定名・別captured actor・新たなoperation_keyを受けない。親record lock保持を必須とする。event INSERTとprojection UPDATEを実行し、成功event IDを返す。業務本文やbefore/afterの第二コピーは持たない。snapshotはTX内でlock済staffから読み、外側participantが全子保存結果を集約し、changes[]を(owner_kind,owner_ref,change_kind)順に整列・重複排除して最後に1回だけ呼ぶ。子repositoryから独立にeventを作らない。同record同operationの再呼出は全変更集合とactorが一致した場合だけ同eventを返す。異owner/ref/版の正当な複数子は集合の別要素であって競合ではない。途中の部分集合でeventを確定した後に別子を追記する方式は禁止。

C06参加：494既存record owner/元anchor/lookupを再利用し、root/childの外側業務操作に統合する。新帰属専用HTTP write route、別帰属claim、別receiptは作らない。実差分保存→既存必須履歴/audit→帰属event/projection→既存C06成功結果の順を同ambient TX内に置く（auditとeventの前後が保存結果を変えないよう、どちらも成功必須）。親recordに対するevent順はそのrecord lockで直列化し、timestamp/MAX(id)でfirst/lastを再推定しない。no-opはevent/projection更新なし。既存consumerがno-op receipt/auditを必要とする場合はそれを維持する。

同key成功replayはApplyを呼ばず元receipt、actor名再読なし。ResultResponseへ新氏名本文を入れず、保存後の既存detail readで取得する。既存成功refsへauthor_event_idを出す必要はなく、業務receiptのshape変更を今回必須にしない。event.operation_id→bindingと同TXreceiptが同一成功の証拠になる。unknown→lookupは元consumerの契約そのまま。

### 10.1 root有値createの成功境界

medical_record_handler.CreateMedicalRecordの現在の Create → CreateSubRecords の二段呼出を、同じ外側create participant内にまとめる。medical_record_crud.createMedicalRecordInTx → medical_record_subrecords.upsertInquirySubRecord/ensureClinicalPlanSubRecord は同ambient TXを受け、内側が独立commitしない。root insert・有値子insert/update・必要監査・帰属・receipt全部成功して初めて成功応答。既存POST /medical-recordsの互換adapterは現行201＋既存resource DTOを維持する。内部C06のsucceeded/replayedをresource読取へ変換し、replayでCreateSubRecordsや帰属を再実行しない。新C06 consumer/共有lookupは既存494の200 ResultResponseを使い、既存POSTの201と同じ公開DTOだとは扱わない。既存POSTの失敗/結果不明では201を返さず、operation_key付き再試行・lookupはそのconsumerの明示契約に接続する。子失敗ならrootもrollback、既存recordヒットを新実入力として再記録しない。空createではunenteredのみを同TX作成し、氏名eventなし。LSTEP通知など既存外部post-commit作用をこのDB TXに取り込まない。

既存Createのentered_byはHTTP作成者履歴として残っていても、author_attribution.firstをそこから埋めない。rootだけで実入力と判定できる有効なheader/推奨理由差分は、正規化済の利用者実入力として保存する場合のみ対象。サーバ初期値・予約由来日付・自動担当配置だけなら空生成扱い。既存staff/clinic/owner/pet検証・作成重複防止・対象変更禁止を弱めない。

### 10.2 hookとPR単位（実装未完の必要subset）

| slice | 正確な接続開始点 | 必要閉包/完成条件 |
|---|---|---|
| A storage/reader | 上記3private files、model、API detail DTO、StickyHeader | 2表/制約/RLS、移行legacy_unknown、nullable DTO、GETだけ。writer未接続なら全履歴完成と表示しない |
| B root作成/更新 | medical_record_handler.CreateMedicalRecord / UpdateMedicalRecord / UpdateMedicalRecordRecommendationReason → medical_record_crud.go Create/Update/UpdateRecommendationReason → medical_record_subrecords.go | root有値createと子同TX、actor=nil廃止、best-effort必要監査を同TXへ、C06/494 shared必要subset。空自動create marker参加もこのsliceで揃える |
| C 問診/診察 | inquiry_service.Save → inquiry_repository.SaveByMedicalRecordID、clinical_plan_service.Update/Delete → repository | 494/495/496 ownerの同ambient participantへ帰属hook1回。GetOrCreateのみ除外。既存auditUpdateTx/auditDeleteTxを重複追加しない |
| D 子臨床 | routes_records.goのvitals/treatments/checkups/field-results/vaccinations/examinations/items/unconfirm、497新recordbound | 各ownerのroot record解決、既存TX/CAS/目的別監査の後にhook。具体service内部の全分岐閉包は未確認なので必要subsetとして各ownerにhandoffし、表への掲載だけで完了判定しない |
| E 追記/画像/目的別 | addenda/image upload/delete/既存削除・訂正consumer | 自操作/実差分/record対応を確認し同TX hook。326独立author表示/画像binary処理を再実装しない。blob外部保存の成功だけでDB帰属を先行させない |

C/D/Eの未参加writerが残る間は「全成功編集を反映済み」へ切替しない。A/BだけのPRは限定scopeとして完結できるが、EMR-368全体受入完了ではない。新しい個々の臨床ルールを聞き直さず、既存ownerの技術閉包確認として扱う。

### 10.3 lockとrename競合

実コードでmedical_record_entered_by_actor.goのcreate guardはstaff→assignment SHAREをrecord作成より前に取得する。一方sharedkernel.LockDraftMedicalRecord（medical_record_lock.goからdelegate）は親record FOR UPDATEを取得するため、R3の「lifecycle→staff」はrecord→staffになり得る。これを撤回し、actor lockをlifecycleの前へ置く。

選定順序：既存必要graph gate取得→binding lock/元target照会・現在開示認可→未成功Applyに限るtrusted actor staff SHAREと既存認可に必要なassignment/account SHARE→既存lifecycle保護（linked appointment/record lockを含む）→494 edit row→record（既取得なら再利用）→子owner安定順→author projection/event。単なるclaim前の非locking認可とは別に、外側業務TXでgraphとbinding取得後、保存済み成功の現在viewによるreplayを先に分岐し、未成功Applyに限ってlifecycle/業務row lockより前にactor resolverを実行する。成功replayに新しい編集権・在職・現在氏名取得を追加要求しない。現guardのstaff→assignment順を再利用し、同じstaffからnameを読む。actor/name結果は同TX contextのtyped private値で保持し、別TXやmutable participant fieldへ持ち越さない。graph contextを必要とするconsumerは489/494のContextGate subsetを使い、actor resolverはその返却contextを継承する。

これによりcreateの既存actor先行とupdateのlifecycle先行を、帰属対象consumerではactor先行へ統一する。すでにlifecycle内でrecordをlockした後から氏名resolverを呼ぶhookは採用しない。claim metadata段階でstaff/record業務FK lockを追加しない。graph取得の実装が既に同actor/assignmentをlockしている場合は同じ結果を再利用し、actor以外のstaffを後段で追加lockしない。multi-actor bulkは本票で新設せず、1operationの認証actorは1人。対象workerへ新規staff範囲を広げない。

通常name変更のstaff_repository.updateは同staff row UPDATEと競合するので、rename先行なら新名、保存側先行なら旧名でcommitまで一貫する。改名処理にrecord lockや全履歴更新を足さない。staff無効化/所属変更等の既存graph gateを必要とする経路はその先行gateを維持し、clinical側がgateより先にstaffを取る順序は禁止。現在の全owner実装がこの順序へ統合済みとは主張せず、slice B/Cで入口変更、D/Eは同じ順序への参加を必須subsetとする。deadlock/serialization失敗は全rollbackと元key回復、lockなし現在名fallbackなし。

### 10.4 projection状態・DTOの厳密意味

遷移：新方式空create→(first=unentered,lastなし)。初回実差分成功→(first=recorded[first_event],last=同event)。後続実差分成功→first不変,last=新event。legacy_unknown→first unknown不変,lastだけ新event。no-op/失敗/replay/改名/派生更新→projection不変。first recordedから別eventへの変更はrepository/DB guardで拒否し、記録者訂正を本票の通常writeで解禁しない。projection欠落は読取上legacy_unknown、GETで行を作らない。既存recordの最初の新方式編集ではrecord lock下で INSERT ... ON CONFLICT DO NOTHING によりfirst_state=legacy_unknown、first_event_id=NULLのprojectionだけを用意し、今回last_eventを設定する。欠落したからunenteredとしてINSERTすることは禁止。新create branchだけが同create TX内でunenteredを明示できる。

DB相互制約：first_state=recordedならfirst_event_id/last_event_idともNOT NULL。unenteredなら両event NULL。legacy_unknownならfirst_event NULL、last_eventはNULLまたは今回以降の実保存event。各event FKは同record/clinic、first recordedの再更新は拒否する。最後のeventをNULLへ戻す通常UPDATE、recorded→unknown/unentered、unknown→unenteredを禁止する。first recorded→recordedではfirst_event同一、lastだけ許可。unknown→recordedは通常編集では禁止し、証拠付き移行専用経路をこの票で追加しない。新unentered→recordedはfirst=last=今回eventでのみ成立。新しいevent順はrecord lockで決まるので、id/timestamp大小制約で代替しない。

公開detailに追加するauthor_attributionはclosed {first,last}。firstはclosed {state,staff_id,name_at_save,saved_at,operation_key}、lastも同じ5field。stateはrecorded/unentered/legacy_unknown。recordedでは5fieldすべてnonnull（IDは新fieldの正十進string、nameは保存済text、saved_at RFC3339、operation_key UUID）。unenteredではstate以外全null。legacy_unknownでは保存済み旧IDが確実な場合のみstaff_idをstringで保持し、残りはnull。旧lastに確かなIDがない場合は全null。新unentered recordのlastはunentered、旧欠落recordのlastはlegacy_unknown。新成功編集後のlastだけrecordedへ。

新wireのstateは氏名根拠の状態であってスタッフ在職/取込処理の状態ではない。新snapshot欠損をlegacy_unknownで隠して成功返却しない。recordedなのにevent/氏名が欠ける不整合と読取失敗は既存非2xx/error、DTOにfabricated nameを返さない。既存entered_by/doctor等のnumber wireとEnteredByStaff互換fieldはこの追加で破壊せず、FEの新本体氏名表示はauthor_attributionだけを読む。旧preloadを保存時名fallbackへ使用しない。

物理方式とno-opは採択済み。残るD/Eの具体hook/C06閉包証跡は実装準備subsetであり、ユーザー業務判断待ちではない。runtime/DDL適用/試験はNOT_RUN。
