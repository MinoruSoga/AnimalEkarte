# EMR-485 非表示会話要約: storage-only CAS採用契約

2026-10-21。独立静的レビューを経て技術方式をselectedとして固定する。既存非公開adapterを再利用するが、以下の厳密CAS・空本文・破損検知・ログ境界の適合変更は未実装。コード/試験/DB/実入力変更なし、全試験NOT_RUN。

## 先に結論
既存の3列だけでなく、conversation_summary_repository.go と conversation_summary_service.go が既に存在する。新table・新HTTP API・新producerは不要。今回の完結単位を「非公開storage adapterの型/NULL/CAS/非露出と局所保存契約」に限定し、入力サービス選定/業務利用開始/C06 operation連携を完成要件へ勝手に広げない。

F監査の「adapter未具体化」は、この追加コード照合で訂正する。adapterを新規作成するのではなく、現実装のLWW/内容一致再送判定などの採否と保証範囲を明記する。現コードの存在は実DB/制約/認可受入PASSを意味しない。

## 1. 静的確認済み実装
基準SHA: 0de244ba69f9287a0bacff6abeddabb50a712aac。
- medical_records: conversation_summary text NULL、conversation_summary_revision bigint NULL、conversation_summary_source text NULL。3列全NULL、またはsummary非NULL・revision非NULL正数・source非NULL非空の既存metadata CHECK。
- model.MedicalRecordは3列すべてjson:"-"。一般JSON公開の対象ではない。
- 非公開ConversationSummaryService.Get/Save、Repository.FindByMedicalRecordID/Save、型ConversationSummary/SaveConversationSummaryInput/Commandが実在。
- repositoryはclinic_id + id + deleted_at条件、DBOrTx、3列だけのUPDATE RETURNING。通常カルテupdated_at/version/clinical set・所見診断方針には書かない。
- 現実装はExpectedRevision=nilでLWW、0で未保存、正数で期待版。但し内容一致なら期待版99等でもno-op成功する。これはoperation keyによる元成功照会ではない。
- serviceはsourceをTrimSpace、本文は保存内容維持。blank本文拒否はrepositoryコメント自身が未承認の防御判断と明記。revision/source以外の型/長さ制約は未選定。
- serviceのslogはclinic/record/revisionだけ、本文/sourceはログ出力しない。actor/監査record/operation receipt/fenceを保存する処理ではない。
- constructorのGitHub code searchは各定義+testの2件のみ、incomplete_results=false。検索indexは別SHA edf71515...を示したため全treeの厳密な不存在証明とはしないが、確認範囲でproduction composition/handler/入力producerは発見していない。既存serviceコメントも公開未接続を明記。

## 2. 採用する最小contract

### 2.1 入出力とidentity
既存内部型を再利用。HTTP route/DTO/OpenAPI/画面を作らない。
- Get(ctx, clinicID, medicalRecordID) -> ConversationSummary。
- Save(ctx, SaveConversationSummaryInput) -> ConversationSummary。
- clinicID/medicalRecordIDは正のDB ID範囲。呼出元の自由なclinic切替を許可する意図ではなく、将来の認可済み内部入口から渡す対象。storageのscope条件はどちらも必須。
- 他院/不存在/論理削除は同じNotFound。record作成、他患者への付替え、他recordからの暗黙コピーをしない。
- recordは元の保存先identityであり、source stringから対象recordを選ばない。pet/ownerをbodyから更新しない。
- Getが返すIDは要求recordと同じ。3列部分NULL/非正revisionなど制約外の旧破損を見つけた場合はunknownの空欄補完やwriteで修復せず、不整合として非成功。これは採用する技術防御。

### 2.2 NULL/空文字/長さ
- 未保存は3列全NULLをそのまま返す。NULL→""/0変換なし。
- SaveのSummaryはstring、Sourceはstring、ExpectedRevisionは必須の0以上int64へ限定する。既存pointerを互換上残す場合もnilはinvalidで、LWW bypassは公開しない。
- storageだけの契約は既存DBに合わせsummaryの空文字/空白を有効な非NULL値として保持する。空だから未保存/削除/正常な診療要約と判定しない。既存blank拒否は業務要件ではないため、その維持を確定仕様扱いにしない。
- sourceはserviceで前後空白を除いた非空文字列、本文は無断trim/要約再生成しない。DBに格納不能な文字列はvalidation/storage errorとして失敗、黙って置換しない。
- 新しい文字数上限や任意12KB等の制限は作らない。既存text保存型の範囲で、切捨てなし。将来の外部入力サイズ制限はそのproducer契約で別途選定し、storageの業務意味へ混ぜない。
- clear/delete操作は追加しない。Saveにnullを持たせず、3列全NULLは未保存状態としてのみ扱う。将来の明示削除は別判断。

### 2.3 版・CAS・同値
- conversation_summary_revisionはこの保存値の改訂番号。元会話/音声/sourceの版やmedical_records.versionではない。
- ExpectedRevision=0は3列全NULLの初期状態、n>0は現在summary revision=n。
- 初回異値保存はrevision1、その後summary/source差分の保存で+1。同じ内容かつ期待版一致ならno-op、版不変。int64最大値なら増分せずconflict/overflowとして拒否、wrap/負数なし。
- **期待版不一致は内容一致でもconflict**とする採用技術是正。同値であることは「元の操作がcommitした」証明ではない。期待未来版、他writerが同値を書いた例、A→B→Aの区別を消さない。
- 応答不明後のGetは現在状態の確認のみ。元operation成功か不明なら、その事実は不明と残す。現在値一致を元receiptの代用にしない。再送/新たな更新は最新状態比較後の明示判断であり、storageが自動retry/LWWしない。
- 同TXで3列のCAS/update/RETURNINGを行い、保存後の別reader値を今回の保存結果と偽らない。現UPDATE RETURNINGを再利用する。0更新後はscope/削除/版不一致を分類し、再読込時の偶然の内容一致を成功へ昇格しない。

### 2.4 sourceとsource version
- 現列Sourceは非空の保存元識別文字列だけ。現testの"lstep"はfixtureでありLSTEP連携採用の根拠ではない。今回のfixtureは非実在のstorage-test等でよく、実provider名を既定値にしない。
- source欄へ本文/認証情報/患者情報を混ぜない。source revision、元会話のID/版、source authenticityをこの3列だけで保証したと主張しない。
- source versionをsummary revisionへ転用しない。今回producerなしなので元source versionの列・JSON詰込・別tableは追加しない。
- 将来producer採用時は、保存元の安定identity/版・対象record対応・旧入力重複判定をそのownerの既存保存資源/operation contractに定義する。これは今回のstorage sliceを止める業務質問ではなく、Phase2利用開始の未選定事項。

### 2.5 現在認可・非露出
- 非表示は単なるUI非表示だけでなく、一般record DTO/JSON、検索、印刷/レポート、複写、CSV、ログ/エラーに自動追加しない。json:"-"を保持。storage structをHTTPレスポンスへ直接使わない。
- 現serviceのclinic条件はtenant分離であり現在staff grant検査の代用ではない。現時点で現在認可/fence/actor auditが完成したと報告しない。
- 本sliceではproduction入口を作らず、認可未成立の内部batchやcronからも呼び出さない。コードの非公開保存能力と利用開始を分ける。
- 将来のread/write入口はそれぞれ現在の既存medical-record閲覧/編集権と対象clinic/recordを検査し、書込は必要な494編集session/fenceに従う。ただし未承認の無人service principalや新permissionは今回発行・選定しない。認可取りまとめは将来producer/入口ownerの責任として残す。

### 2.6 TX、audit、結果照会
- repositoryはambient DBOrTxに参加し、自分でcommitしない。単独保存は単一statement、外側TXがある場合Save返値は「そのTX内の保存結果」であって外側commit成功証明ではない。
- 3列の部分commitなし。外側rollbackならsummaryもrollback。通常診療本文/診断/方針や臨床世代・医師確認・会計・在庫・飼主原文を変更しない。
- 現slogは非機微の診断ログで、transactional auditでも成功receiptでもない。外側rollback前に出たログを永続成功証拠にしない。成功語を使う場合はcommit責任側でのみ出す、またはTX内更新試行と明示する採用方式。
- actor/audit/operation ID列をsummaryへ増設しない。今回production writeなしのstorage-onlyでは新C06 family・HTTP結果照会を作らない。
- 将来の業務writerは必要な現在認可、監査と成功receiptを同最外TXで既存共通資源へ保存し、元key照会で成功/replayを返す。内容一致だけの今回adapterがそれを提供したとはしない。外部I/O/LLM実行はDB TXへ入れないが、そもそも今回選定/実行しない。

### 2.7 更新時点・保管
- 今回の本番更新トリガーは**なし**。通常カルテ保存・タブ表示・診療終了・接種/検査・請求確定・再読込ではsummaryを作らない/更新しない。
- 内部storage adapterを明示呼出したときだけ3列変更。既存通常record.updated_at/versionは変更しない。これは要約版から臨床更新日時を捏造しないための技術分離。
- 保存値は既存recordに保持し、独自TTL/自動削除/job/生音声保存を追加しない。親論理削除後の通常Get/Saveは拒否し、物理行の要約を勝手にpurgeしない。
- 3列は現在値storageで、過去summary本文のimmutable履歴tableではない。全履歴/法定保持/自動消去を保証しない。必要な将来の履歴/保管期間はproducer採用時の範囲確認へ残す。

## 3. 内部エラー分類
routeがないためHTTP statusを新設しない。既存apperrors分類を維持する。
- invalid_input: ID0/範囲外、source空、ExpectedRevision nil/負数、DB非対応文字列。
- not_found: 対象なし/別clinic/論理削除。存在有無の追加説明なし。
- conflict: 期待summary版不一致、版増分上限。current本文をエラーへ含めない。
- internal/dependency: repo未設定、DB/TX失敗、metadata不整合。NULL未保存へfallbackしない。

## 4. 固定fixture条件（実行NOT_RUN）
1. 旧3NULLのGetは3NULL、readがINSERTしない。
2. 初回expected0→rev1、同値かつexpected1→rev1、異値expected1→rev2。
3. nil/未来99/古い版は拒否。同値未来99も拒否。A→B→A後の古いrequestは元成功と認定しない。
4. summary空文字を非NULLとして保存可能、source空は拒否。NULL clear操作なし。
5. 別clinic/不存在/soft-deleteは同じNotFound、対象ID付替えなし。
6. 更新3列のみ。record updated_at/version/identity/clinical generation、ClinicalPlan全列/飼主原文不変。
7. ambient TX rollbackで3列も戻る。返値/ログをcommit証明にしない。
8. current Findはoperation lookupでない。応答不明→現在値一致だけで元key成功と表示しない。
9. JSON/model/通常DTO/CSV/印刷/ログにsummaryを露出しない。producer/constructorのproduction呼出なしを静的確認。
10. 正式001相当CHECK適用で3NULL/完全metadata/破損拒否、旧AutoMigrateだけのfixtureを本制約PASSとしない。既存testのstatus列参照は最終DDLと再照合し旧列を復活させない。

## 5. 本当に残る業務入力の未確定
今、ユーザーへ追加回答を要求するものはない。今回の承認範囲は非表示保存項目で完結できる。
将来、誰/何がどの会話から要約を作るか、生成/更新タイミング、source版/履歴・保管/削除の要件、外部サービス/API/人の確認方法は未選定でPhase2に残る。技術C06/fenceの不足をこれら業務決定と混ぜず、今回勝手に手動/LLM/LSTEP連携を追加しない。

## 6. 業務仕様と票への引継ぎ要旨
D03 lifecycleの会話要約節への引継ぎ:
「外部サービス選定と公開API・入力経路は未選定。会話要約の専用保存列・metadata制約と非公開storage adapterは既存成果を再利用する。今回の対象は非表示の来院別保存契約に限り、入力producerの追加、業務利用開始、現在認可/監査/operation結果照会の全経路接続・検証完了を意味しない。」

EMR-485への引継ぎ:
「既存conversation_summary_service/repositoryと3列の非公開保存を再利用し、新規table/公開API/入力producerを追加しない。現adapterはclinic分離・DBOrTx・3列UPDATE RETURNINGを持つが、任意CAS/LWWと内容一致による再送判定は操作成功証明ではなく、現在staff認可・TX監査・成功receiptは未接続。storage-onlyの技術採用表で厳密CAS/NULL/非露出/返値のcommit境界を確定し、通常カルテ保存や診療終了から自動呼出ししない。保存能力と業務利用開始を分け、未試験はNOT_RUN。既存成果を新規実装し直さない。」

## Sources
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/medicalrecord/conversation_summary_repository.go : L45–80入力/未承認blank、L101–118read、L120–205SQL/CAS/no-op。
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/medicalrecord/conversation_summary_service.go : L38–44認可未接続、L72–99trim/ログ。
- 同SHA backend/internal/model/medical_record.go :3列json非公開。
- 同SHA docs/architecture/erd-detail/medical_records.md L30–32/134。
- 同SHA docs/spec/domains/medical-records/lifecycle.md L506/523。
- 同SHA docs/design/domains/medical-records/technical-spec.md : 来院別会話要約/D03「今回は非表示保存項目だけ。本文自動更新・全文音声・手動貼付UI・外部取込writerを追加しない」。
- 検索 https://api.github.com/search/code?q=NewConversationSummaryService+repo:MinoruSoga/AnimalEkarte およびRepository、2026-10-21 04:38 UTC。index SHA差を上記のとおり限定。

## 7. 独立レビュー追補（2026-10-21 05:35 UTC）

最新main `0de244ba69f9287a0bacff6abeddabb50a712aac` のservice/repository/model、001 metadata CHECK、D03業務・技術仕様を再取得し、上記実装事実を照合した。旧検索indexによるconstructor件数は過去の限定調査としてのみ保持し、最新main全域でproduction呼出が絶対不存在という証明に使わない。公開接続を追加しないことは本採用契約の制約であり、配備済み環境の実測結果ではない。

model.MedicalRecordの3列は既存json:"-"で非露出。一方、ConversationSummary内部structのexported fieldsはjsonタグがなくてもGo標準JSONで出力できるため、タグがないことをセキュリティ境界と説明しない。内部structを一般responseへ直接渡さない、一般DTO/serializer/exportへの投影を作らないことを必要条件とする。

厳密CASは現SQLの期待版述語にある「OR 内容一致」と、0行更新後の内容一致成功fallbackの両方を撤去する必要がある。nil拒否はserviceだけでなくrepository command検証でも行い、内部callerの迂回を残さない。ExpectedRevision=0はCOALESCEだけに頼らず3列全NULLを確認する。正常な保存済みmetadataで期待版一致のときだけ同値no-opを認める。部分NULL/非正revisionはGetおよびSaveの検証対象で、制約が未適用のfixtureを正常旧未保存とみなさない。

revision最大値でも同値かつ期待版一致のno-opは増分しないため可能。異値更新だけをoverflow conflictとして拒否する。既存本文空白拒否を外し非NULL空文字を保存可能にするのはDB CHECKと整合させるstorage表現の選定であり、空要約を臨床上有効と認める判断ではない。producer/UI・通常カルテ保存への接続は追加しない。

serviceの現「conversation_summary saved」ログは外側TXのcommit前にも出る。採用後は保存能力の返値とcommit証明を区別し、service内で残すログはTX内更新試行と明示する。成功audit/receiptへ格上げしない。source/本文のログ出力禁止、通常record.updated_at/version不変、その他の診療本文不変を保持する。
