# D07 検査結果 mutation 契約（D4）

識別: EXAM-RESULT-MUTATION-CONTRACT-20261023-V1  
採択日: 2026-10-23 UTC  
状態: **D4 技術契約は限定採択済み。供給・実装・試験・機能 activation は未完。**  
配置先: `docs/design/domains/examinations/result-mutation-contract.md`  
製品source確認基準: `04a9078d93c3899705ce9d420eae75e89709e2e3`  
チケット・migration規約確認基準: `bd309de7b9680a46fdf3ed4a89491c198624288c`

## 適用範囲と優先関係

本書は EMR-495 D4 の検査結果更新CASを EMR-501 C が単一取りまとめし、EMR-516 S1/S1a・E-linkおよび各writer ownerへ受け渡す技術正本である。結果専用counter、atomic read、closed write、固定target付きC06、writer参加閉包、両CORS preflight受領契約を採択済みとして固定する。必要DDLの内容は§10に定義するが、配置・採番・適用は現migration規約と446/466の単一ownerに従う。

[D07 lifecycle](../../../spec/domains/examinations/lifecycle.md) と [D07 technical-spec](./technical-spec.md) の臨床状態・確定・確認解除・revision・認可の意味は維持する。D4対象のwire・CAS・receipt・writer参加については本書を正本とし、旧wire/旧unguarded経路は移行対象とする。既採択の `exam-link-source-v1` はlink専用のままであり、本書のcounterへ置換しない。

**限定採択に含めないもの:** Detachのexact job target/key/成功・unknown回復契約、Revert回復のC06同等性確認、501独立A/Bのprovenance・感染症summary等の未選定wire、正式実行ID/claim、schema・adapter・生成物の供給、実装接続、試験実行、DB適用、本番activation。詳細は§11・§13。契約採択はこれらの完了証拠ではない。

本書の受入caseは期待値であり、fixture code作成・試験実行の証拠ではない。試験状態は **NOT_RUN / 判定UNKNOWN**。新規臨床業務判断は追加しない。

## 1. 採択契約の要点
1. **単一取りまとめ機能ownerは EMR-501 内の既存C（CAS/receipt/履歴）**。D07本文/項目/revisionの同一aggregateのため。516は既存S1/S1a coordinatorとE-linkの参加契約を受渡し、495はD4の呼出側/旧guard移行を受領する。個人の担当変更・新票・新stageは作らない。
2. `exams.result_state_version` という**結果aggregate専用の単調counter**を新設する。既存 `current_revision_version` はworking/official履歴のまま、`exam-link-source-v1` はlink-onlyのままにする。
3. GET detail/itemsが返す同一snapshotのcounterを、既存PUT/PATCH/unconfirm/DELETEの必須header `X-Exam-Result-State-Version` にそのまま送る。各writeは `Idempotency-Key` 必須、C06のflat ResultResponseを返す。旧unguarded callerは値補完せず拒否する。
4. 全existing-result writerを、同じambient transaction参加capabilityへ移す。job→exam順、親record/planよりexamを先取りしない順序、lock後のidentity/版/依存再読、結果counterの1回更新、履歴・usage・監査・success receiptの同一commitを要求する。
5. 新規作成は旧examを上書きせず初期counter=1。同時機器受信を一律D03編集session待ちにしない。機器detach/revert、revision restore、inline PATCH itemsをD4外の抜け道にしない。
6. C06未成功lookupに必要な `examination_operation_targets` を追加する。binding+clinicだけFKを持つ不変claim metadataであり、別receiptや臨床本文保存先ではない。

**本書の契約採択だけで「結果PUT安全化済み」「516のactivation可能」としない。**

## 2. 既存契約との境界

| 論点 | 既存契約・確認基準 | 本書で固定する境界 |
|---|---|---|
| 旧カルテfinalized | 診療終了だけを一律編集禁止にしない。D07独立writerに人のD03 sessionを一律要求しない（495 D1） | D07 owner guardと、実際にD03 aggregateを編集するときだけ必要なfenceを分ける |
| 結果の保護 | confirmed/official、revision-less completed、過去official、専用unconfirm、現在認可を保持（D4/D07） | counter一致でもこれらの保護を通過したことにはしない |
| 旧PUT | missing/nullと明示[]を区別し、無期待値で開放しない（D4） | missing/nullは400、[]のみ明示clear。全mutatorに必須header |
| CAS媒体 | D4は従来未選定だった。generic exams.versionやDDLを独自追加しない | 本書はresult_state_versionとtarget metadataを採択。実装・DDL供給・適用は別途必要 |
| 履歴版 | current_revision_versionはNULL又は実存revisionの連番。working/official更新を持つ | 履歴の意味を変えず、別counterを同TXで進める |
| 516 link | 固定revision又は真unversioned digest、6操作、link-only grant、plan版更新、実owner不変 | link payload/guardを変えずowner-lock参加を受渡す。result counterはlink操作で進めない |
| C06 | Scope、immutable Content、Preconditions、ClaimMetadataWriter、現在認可付きlookup/replay | D07 consumerのtarget metadata、Content projection、receipt keyを具体化 |
| Q | 固定mainにVerifiedCommand実装が存在 | 実装の再作成待ちにしない。D07 projector/adapterだけが必要 |
| 臨床判断 | 感染症code mapping、結果正常/異常、他院由来、不明日、死亡後専用処理は既存ownerの仕様 | 本契約はこれらを新造・再定義しない。未選定の501 wireをCAS採択で決定済みにしない |

495 D4原コメント: `f7411e88-eb92-409a-a8a0-041b429c2d03`。本書はその結果mutation CASの未選定を技術契約として解消する。供給・実装・activation完了までを意味しない。

## 3. 基準sourceの実状と移行対象

- `model/examination_record.go`: status/date/一部itemsはnullable、CurrentRevisionVersion nullable。aggregate CAS列なし。旧request/response/serviceにはpointer・ExactNumeric移行前の扱いが残る。543のcompile移行とは別に、**NULLを0/空/normal/pendingへ補完しない**。
- `examination_items.go:38-109`: ReplaceItemsはexam FOR UPDATEから始まり、期待値を受けない。親recordを後からlockする。revisionありはworkingをappendするが、unversionedの同時編集を検出するread→write比較はない。
- `examination_request.go`: PUTのnil sliceを[]にしてclearする旧仕様。parent PATCHのitems経由も同じreplace helperへ入る。
- `examination_service_update.go`: 親更新、inline items、pet変更時のitems再評価、初回confirm、working append、reconfirmが結果aggregateを変更する。split PUTだけ直しても閉じない。
- `examination_revision_workflow_repository.go`: mutable parent/子項目のrestoreとcurrent pointer CASがある。履歴counterのNULLとcompleted sealには意味があり、無理に全保存をrevision化できない。
- `lab_import_revert_service.go` / `lab_import_revert_tx.go`: job→exams→usageで補償、親record FOR SHAREを後から取る旧順序。採択516のcoordinator順へ参加が必要。
- `lab_import_usage_tracker.go`: clinical readはjob→examの短TX、manual mutationはambientでreceipt INSERT。manual経路が現状でjob row lockを必ず取得している、とは読めない。noop trackerはproduction安全性の証拠にしない。
- `examination_service.GetByID` はexam読取後にusage用TXへ進む。単に最後にcounterを読むだけでは、古いitemsと新counterの混合が可能。新read DTOのsnapshot一貫性を必要条件とする。
- frontendはGET detailとGET itemsを別cache、PUTはitems配列、PATCH/unconfirmはexam DTO、DELETEはvoidを期待し、key/expected counterを送らない。C06結果へのcaller切替が必要。
- repositoryのReplaceItemsは親lockの後にhard-delete/insertする。parent→itemsの順は維持するが、別の独立TXを開始できる旧汎用関数を安全な参加capabilityと同一視しない。

基準sourceの静的確認であり、稼働DB・全composition rootの有効化状態・現claimの証拠ではない。

## 4. 結果aggregate専用counter

### 4.1 名前・意味・型
- DB: `exams.result_state_version BIGINT NOT NULL DEFAULT 1 CHECK(result_state_version >= 1)`。
- Go: positive int64範囲の非nullable値。wireはcanonical decimal string（1〜9223372036854775807、先頭0/+/-/小数/指数なし）。JS Number経由は禁止。
- これは「同じexamの編集可能な結果状態の世代」であり、臨床確定番号、検査回数、実施回数、plan版、医師確認、C06成功件数ではない。
- 既存rowは導入baseline=1。過去の更新回数を推測しない。current_revision_versionがNULLのrowもNULLのまま。GET/lookupで初期化しない。
- 新規exam＋初期items＋同create内初回officialは、一つの新aggregateとして最終counter=1。初回revision番号は既存D07契約に従う。途中のpendingは外へ公開しない。

### 4.2 counter参加範囲
既存examの次の変更はすべて同じcounterに参加する。
- parent: medical_record_id、pet_id、exam_type_id、doctor_id、date、performed_date_precision、status、result_summary、machine、job_id、deleted_at、current_revision_version。
- 子: exam_resultsの追加/更新/削除/並び順/判定値/基準値snapshot/nullable master参照変更。
- revision: 新working/officialと現在headの進行、mutable restore。過去immutable revisionを編集する新権限は作らない。
- 501の採択後のprovenance等が保存されるwriterも、このaggregateの変更として参加する。CAS契約だけでその業務payload/保存mappingは決めない。

以下はcounterを進めない。
- GET、lookup、replay、lab clinical-use receiptだけ、別表のmaster設定変更だけ。
- 516 link-onlyのlink/replace/unlink。plan版だけを進め、結果counter/current_revision_version/本文を変えない。
- 重複importを「既存examの書換なし」としてskipする経路。
- 失敗/rollback。C06の成功が不明な間は再実行しない。

**no-op規則:** 既存examに対し新しいkeyのmutating Applyが正常実行されたら、値が偶然同じでもcounterをちょうど1進める。空itemsから明示[]への置換も+1。複雑な値同値判定を隠れた採択条件にしない。同key replayは0回。
- counter上限ならwriteを行わずvalidation。wrap/負数/0への再初期化なし。
- child N行、parent update、revision appendの各helperが別々に加算しない。最外のD07 aggregate mutation participantが一回だけ消費する。
- SQL CASは `clinic_id,id,deleted_at IS NULL,result_state_version=expected` を必ず条件に持ち、`result_state_version=expected+1` と必要parent変更を一回のUPDATEに集約しRETURNINGをreceiptへ使う。status/pointer等の既存追加CASも落とさない。
- soft deleteは同UPDATEでdeleted_atとcounterを進め、削除前のscoped証跡からreceiptを確定する。hard-delete新設なし。
- child置換/履歴appendが先に実行されても、最終CAS失敗・監査失敗なら最外TX全rollback。二段commitは禁止。

### 4.3 採らない代案
- current_revision_versionだけ: unversionedでNULL、working/officialの意味とsealを壊すため不採用。
- updated_atだけ: 子更新参加と時刻精度/ABAを保証しないため不採用。
- link専用digestだけ: 採択用途違い、ABA、結果writerの契約がないため不採用。
- 全writerで新しい共通exams.versionだけ: 意味を広げず専用名に絞る。
- 新counterの導入だけ: 旧raw writerが残ると偽安全なので不可。§8の閉包と一体で受領/有効化する。

## 5. Exact read → write wire

### 5.1 GET
固定mainではdetailは `examinationResponse` object、itemsは既に `examItemsResponse={items:[...]}` object。**backend itemsが裸arrayなのではなく、frontend get-examination-items.tsがenvelopeを捨てExamResult[]へ変換している。** 新shapeでは両objectへ下記required contextを追加する。さらにnullable/ID/ExactNumeric/date型の明示移行を伴うため、wire全体を「純粋な加算変更だけ」とは呼ばない。

```json
"result_write_context": {
  "schema": "exam-result-write-v1",
  "examination_id": "123",
  "result_state_version": "7"
}
```

- contextの3keyはexact、追加不明keyなし。trusted clinicは認証/選択医院から決め、client clinicを採用しない。
- contextは権限tokenではなくCAS値。viewのみのcallerにも閲覧可能だが、書込grantを付与しない。
- **detailのD4 scalar shape**は required keys `id,clinic_id,medical_record_id,pet_id,exam_type_id,doctor_id,date,performed_date_precision,result_summary,machine,status,current_revision_version,created_at,updated_at,items,result_write_context`。ID/版はcanonical decimal string、nullable ID/版は明示NULL、dateはDB dateのYYYY-MM-DD又はNULL、precisionはday/unknown/NULL、statusは既存5値又はNULL、summary/machineは保存string、timestampはUTC RFC3339Nanoの保存時刻。追加で既存表示relation `pet,doctor,exam_type` はoptional、各objectの非D4表示fieldは固定source examination_response.goと各SummaryResponseの形を維持し、本D4 commandへ取り込まない。既存516の `plan_link_source` は受領時のみ同採択schemaでoptionalに加わる。未実装なのに空placeholderを返さない。
- **items endpointのexact top-level shape**は `{items: ExamResultD4[], result_write_context: ResultWriteContextD4}`。itemsはrequired array、0件は[]、NULL不可。detailのitemsと同型。裸arrayへの変更なし。
- **ExamResultD4のexact keys**は `id,exam_id,exam_type_field_id,name,inspection_value,normal_value,result,unit,reference_value,ref_min,ref_max,qualitative_min,qualitative_max,is_assessed,is_abnormal,status,sort_order,created_at,updated_at`。IDはstring、field IDはstring/NULL、ref_min/maxはExactNumericの正規decimal string/NULL、qualitative boundsはstring/NULL、is_abnormalはbool/NULL、statusはnormal/high/low/NULL、sort_orderは整数/NULL。is_assessedは既存判定helperの結果であり、NULLをnormalへ置換しない。他の保存stringとtimestampは上記どおり。schema外enumを受領したら契約不整合として編集を停止し、新正常値へ変換しない。
- read-only既存表示callerへの互換影響は、旧number ID→string、DateTime→date/NULL、ref_min/max number→decimal string/NULL、optional省略→明示NULLという型変更。OpenAPI/generated型/transformを同固定版で受領するまで新shapeを有効化しない。FE生成物を手編集・agent生成しない。
- detailは同snapshotの `items` を必ず含める加算変更とし、編集画面はdetail一応答を編集baselineとする。items endpointもitemsとcontextを同snapshotで返す。list/printの値からwrite tokenを合成しない。
- job候補を非lockで読んだ後、関連jobを既決順にlock→examをSHARE（必要なら既存usage participantに合わせUPDATE）→再読→items読取→必要clinical-use receipt保存、同TXで応答素材を確定する。jobIDが変われば応答を作らずやり直す。child writerはexam UPDATE lockを取るため、itemsとcounterの途中変更を防ぐ。
- stale detailと新itemsを同一編集baselineへ混ぜない。分割readを使うcallerは両contextのexam/version一致を検査し、不一致なら全baselineを再取得する。FEでcounterのみ新値へ差替えない。
- FE cacheは `['examinations','result-read-v1',clinic_id,exam_id,'detail'|'items']` というschema分離keyへ切替し、valueはenvelopeとcontextを保持する。既存配列cacheを新schemaとしてhydrateしない。ID/clinic/schema/version不一致の遅延応答は現在editorへ適用しない。編集baselineはdirty状態なら通常refetchで上書きせず、利用者が差分確認した後に一式交換する。旧cacheの削除/新cache無効化は成功/医院切替時のadapter責務。
- nullable status/date/items値は正本どおり保持。unknown statusや参照切れはtokenの存在を根拠に保存可能としない。writerが既存保護で拒否する。
- 現model/DTOのnullable・ExactNumeric移行は501の必要adapter subsetに含める。enum/nullable不整合をゼロ値で隠してこのDTOだけ追加しない。

### 5.2 PUT items（D4の主対象）
pathは既存 `PUT /api/v1/examinations/{id}/items` を維持する。

必須headers:
- `Idempotency-Key: <canonical lowercase nonnil UUID>`。C06 ParseOperationKeyを使う。
- `X-Exam-Result-State-Version: 7`。直前の利用者が確認したbaseline contextの値。client生成/サーバ自動補完不可。

body exact:
```json
{"items":[{"exam_type_field_id":"42","name":"項目名","inspection_value":"原値","normal_value":"","unit":"","reference_value":"","sort_order":0}]}
```
- items key必須、array必須。欠落/nullはvalidation/400でzero-write。[]は明示clear。
- 各itemのaccepted business fieldsは既存 `upsertExamItemRequest` の7fieldsだけ。field IDはnullable canonical ID、name必須、他の既存string fieldsとsort_orderは既存意味を維持。JSON number IDからcanonical ID stringへの移行はOpenAPI/FE同batchで明示する。
- status/is_abnormal/ref_min/ref_max/qualitative bounds/result等のserver派生・非入力fieldは受け付けない。未知keyも拒否し、無視した別payloadが同一操作に見えないようにする。
- sort_orderはDB integer範囲の整数又は明示NULL。明示NULLは現在modelのNULLを保つ値として、そのままcommand/storageへ渡す。items配列順をそのままcanonical commandに保持し、sort_order順へ勝手にsort/dedupしない。
- 完全DTO内では既存の省略可能stringを空文字、field ID省略をNULL、sort_order省略を0へ正規化してからhashする。sort_orderの明示NULLは0へしない。既存結果を編集するFEはreadしたNULLを保持し、NULLを省略して既存defaultへ落とさない。これはPUT item旧入力のdefaultと現在nullable modelの区別であり、read保存値のNULL補完ではない。items自体のmissing/nullだけは絶対に[]へしない。
- 変更理由を新たな臨床要件として追加しない。既存unconfirm等の理由要件は別に維持。

### 5.3 他の既存exam mutation
同じ2headersを `PATCH /examinations/{id}`、`POST /examinations/{id}/unconfirm`、`DELETE /examinations/{id}` に適用する。grantと保護は現在の各操作のまま。
- PATCHのparentのみ、inline items、status/confirmは同じ一回のaggregate mutation。items省略/nullは採択D07 technical-specの既存契約どおり「変更しない」、明示[]はclear。PUTのmissing/null拒否と混同しない。PATCHのclosed payloadは下表に固定する。
- unconfirmのreasonは既存必須・長さ制約を維持しContentへ含める。official保護の迂回を認めない。
- DELETEはbody追加不要。headerで期待counter/keyを受け、成功も204/voidではなくC06 receiptを返す。
- 初回createに存在しないexamのexpected counterは要求しない。新規作成の重複防止/C06はそのcreate ownerの既存契約で扱い、本契約は既存examへのunguarded updateへ流用させない。

#### PATCH `exam-result-command-v1` のclosed field/presence表
top-level allowlistは下表9fieldだけ。unknown/duplicate keyは拒否。payloadは**指定された変更だけ**を持つsparse objectでJCS化し、未指定を空/NULLへ埋めない。正規化後空objectはvalidation/400。claim後にdefaultを足さない。

| field | wireの許容値 | omitted / explicit NULL | canonical payloadと適用 |
|---|---|---|---|
| medical_record_id | canonical positive ID string | omitted=不変更、NULL=不変更（現pointer契約維持） | 値ありだけ同keyを保持。元/先record/pet相関とrevisionあり変更禁止をlock下検査 |
| pet_id | canonical positive ID string | omitted/NULL=不変更 | 値ありだけ保持。正当なunversioned変更だけ、既存再評価/死亡保護を維持 |
| exam_type_id | canonical positive ID string | omitted=不変更、NULLはvalidation | 0/空/数値/負数を不変更へ丸めない。master相関検査 |
| doctor_id | canonical positive ID string | omitted/NULL=不変更（現pointer契約維持） | NULLを明示担当解除へ新解釈しない。explicit clearの未選定wireは501 Aへ残す |
| date | valid Gregorian YYYY-MM-DD string | omitted/NULL=不変更（現pointer契約維持） | 正規date stringを保持しDB dateへ格納。時刻/zone付き旧DateTimeはこのv1では拒否しFE変換を同batch化。stored precisionがNULL/dayなら維持。unknown+NULLの結果を既知日に変更するprecision同時更新は501 Aの別wire受領まで対象外としてvalidation、今日補完なし |
| result_summary | string（空文字可） | omitted/NULL=不変更 | 指定stringをUTF-8そのまま保持。trim/空→NULLなし |
| machine | string（空文字可） | omitted/NULL=不変更 | 指定stringをそのまま保持 |
| status | pending / in_progress / result_entered / completed / confirmed | omitted/NULL=不変更 | 指定値そのまま。enum未知は拒否。現在NULLをpending化しない。遷移・seal条件は既存D07のまま |
| items | §5.2のarray（[]可） | omitted/NULL=不変更 | NULLはcanonical payloadから除去。配列は順序保持、各itemは§5.2正規化。[]は必ず残す |

したがって `{"items":null,"result_summary":"x"}` と `{"result_summary":"x"}` は同じcommand、`{"items":[],"result_summary":"x"}` は別command。全fieldがNULL/omittedなら空変更として拒否する。doctor解除・date precision/provenance追加をこの旧互換NULLから推測しない。保存済みnullable fieldはreadにそのまま返し、不変更fieldはDBでも保持する。

PET/record変更の既存D07認可・関連ルールは変更しない。許容された変更を行った成功直後のlookupは§6.2のsuccess targetで扱う。reconfirmが既存revisionありの場合はcanonical payloadが `{status:'confirmed'}` だけであることを必須とし、inline変更との同送拒否を維持する。

unconfirm payloadはexact `{reason:string}`。既存TrimSpace後の非空・500 rune以内を検査し、**保存するtrim後文字列とhash対象を一致**させる。delete payloadはexact `{}`。この正規化方針は技術契約であり、理由の業務要件追加ではない。

### 5.4 応答と旧caller切替
- mutation成功は `operation.ResultResponse` のflat形。最新exam/itemsを同receiptに混ぜない。
- valid keyがある場合のowner rejectionはC06 status。key欠落/不正、未認証、path parse等のC06前段は既存HTTP error envelope。
- missing/不正expected counterは400。**既存rowを読んでexpectedへ代入する互換shimは禁止。**
- 成功後はdetail/items/print/該当record projectionをinvalidateしrefetchする。refetch失敗を保存失敗へ変換して再POSTせず、「保存receiptは成功・再表示取得失敗」を分ける。
- 旧frontendのPUT items-array返却期待、PATCH/unconfirm exam返却期待、DELETE void期待を同リリースでC06 adapterへ切替。二重routeを作り旧unguarded pathを生存させない。
- 全caller移行の完了まで新wireのactivationは行わない。準備PRは独立受入可能。旧callerの400は想定拒否であり、自動新key/自動版更新で救済しない。

### 5.5 ブラウザpreflightの必要subset
固定04a9078の [backend/internal/middleware/cors.go:36](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/middleware/cors.go#L36) と [backend/worker/index.ts:432–438](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/worker/index.ts#L432-L438) は、ともにIdempotency-Keyを許可済みだがX-Exam-Result-State-Versionは未許可。WorkerはOPTIONSをContainerへ送らずedgeで即応答するため、Go側だけの変更ではブラウザの新writeが到達しない。

- **D4-WIRE-01の共有owner受領物**として、両Allow-Headers定数へ `X-Exam-Result-State-Version` を1項目だけ追加する最小差分を要求する。両者の値は一致させる。既存headerを削らず、origin allowlist、Allow-Credentials、CSRF/X-Requested-With、Allow-Methods、Max-Age、Expose-Headersの方針を変更しない。wildcard追加・origin拡大・権限追加は本契約外。
- 実編集は当該Go middleware/Workerの現共有claim ownerが行い、D4-WIRE-01のAPI/FE ownerへ**両変更の同一採用SHA**を受渡す。共有owner未確認/片側未受領を「backendは対応済み」として飛ばさない。本書の採択時点では両設定の変更・受領は未完。
- **Go側preflight fixture**: 既存許可Origin、Access-Control-Request-MethodをPUT/PATCH/POST/DELETEそれぞれ、Access-Control-Request-HeadersにContent-Type/Idempotency-Key/X-Exam-Result-State-Version/X-Requested-Withを指定。OPTIONSは204、既存origin/credentials/CSRF/methods契約が不変で、新headerがAllow-Headersに含まれることを期待する。
- **Worker側preflight fixture**: 同じ入力・期待header/204に加え、Container fetchを呼ばず即応答する既存境界を確認する。Go/WorkerのAllow-Headersが同値であるfixtureも用意する。両側とも非許可OriginにAllow-Originを追加しない負例を保持し、preflight成功をmutationの認可/CSRF/CAS成功へ読み替えない。
- fixture設計/作成と実行は別。現段階では上記は期待値のみ、fixture code未作成、実行NOT_RUN・判定UNKNOWN。試験実行は別途指示時に行い、実行延期を設計/PR作成の新しい禁止条件へ読み替えない。
- **ブラウザactivation gate**: 新headerを要求するwriteのブラウザ有効化には、GoとWorker両方の最小差分・API/FE caller・schemaが同版で受領されていることを必要とする。片側不足の間はactivationしない。curl等の非ブラウザ直接APIだけが動く結果をブラウザ接続完了としない。CORSを回避する別origin/無guard fallbackを用意しない。

## 6. C06 consumer契約

### 6.1 Scope / Content / Q / Preconditions
Scope:
- owner_kind=`examinations`
- intent allowlist=`examinations.replace_items`, `examinations.update`, `examinations.unconfirm`, `examinations.delete`
- clinic_id/actor_kind/actor_refはtrusted現在認証context。通常UIはstaff。callerからactorを採用しない。

immutable command:
```json
{
  "schema": "exam-result-command-v1",
  "action": "replace_items",
  "examination_id": "123",
  "payload": {"items": []}
}
```
- actionは4値。update payloadは§5.3のclosed field/presence表に従う正規形、unconfirmはreason、deleteは空object。
- serverがstrict decode/validateしてcanonical JSON（JCS）bytesを一度作る。UTF-8原文字列保持、binary floatの結果値化なし。
- `NewVerifiedCommand` によるpure projectorを必須化。ApplyはQのbytesからのみdecodeする。別closureのDTO/配列を使わない。
- Content exact keys=`command_schema,action,examination_id,command_sha256`。hashは上の全command bytesのSHA256 lowerhex。Contentに巨大な結果本文を入れず、Qが全payloadのcommitmentを検証する。CAS/fenceをhashに入れない。
- PreconditionsのD4 core exact keyは `expected_result_state_version`。D03 aggregateを実際に編集する操作は既存494/495のfence/dependency subsetを別typed adapterで追加する。独立D07結果PUTに人のsession/fenceを作らない。
- client bodyに同じkeyを二重指定、unknownfield、duplicate JSON key、非canonical IDはclaim前拒否。business内容が変われば旧key流用せず、旧結果が確定した後で新たな明示操作にする。

### 6.2 未成功lookupの永続target
新表 `examination_operation_targets`:
| column | 契約 |
|---|---|
| binding_id | BIGINT NOT NULL、PK、operation binding identity |
| clinic_id | BIGINT NOT NULL、bindingと複合FK |
| examination_id | BIGINT NOT NULL CHECK >0、business FKなし |
| medical_record_id | BIGINT NULL CHECK NULL又は>0、claim時のscope anchor、business FKなし |
| pet_id | BIGINT NULL CHECK NULL又は>0、claim時のscope anchor、business FKなし |
| action | TEXT NOT NULL、4action CHECK |
| command_sha256 | BYTEA NOT NULL、octet_length=32 |
| created_at | TIMESTAMPTZ NOT NULL DEFAULT now() |

- FKは `(clinic_id,binding_id)→operation_bindings(clinic_id,id)` のみ、削除restrict。業務exam/record/petへのFKを付けず、短claim TXから業務row lockへ伸ばさない。
- 既存bindingの有無をtrusted Scope/keyで先に確認する。既存bindingなら保存metadataを使う現在認可付きlookup/replay/key-conflict判定へ進み、新規examのactive状態・現在write grant・新preflightを前段条件にしない。bindingのない新規操作だけ、claim前scoped preflightでexamとidentityを認可付きreadする。並行claimの敗者は勝者の保存metadataへ戻る。claim winnerだけ、bindingとmetadataを同TX INSERT。callback失敗は両方rollback。全列immutable、UPDATE/upsert修復/通常DELETEなし。
- expected counterや現在fenceはclaim targetへ埋め込まない。同じbusiness commandの人によるprecondition再確認はC06既定どおり可能。**未成功の新Applyだけ**はclaim時before identityと現在identityの一致を要求する。成功後replayにこのbefore一致を適用しない。正当な自己変更後を扱う規則は下記success targetに固定する。
- 494/516のrecord/plan target表を流用しない。record=NULLの独立examもあり、plan ownerでもないため。
- ResolveLookupTargetは必ずbinding+clinicのclaim metadataを先に読む。incoming URL/body/receipt.refsから欠落claim metadataを補完しない。intent/action/hashの整合を検査する。未成功の場合はbefore target、成功の場合は下記の固定success targetを追加検証する。
- AuthorizeLookupは元targetの**現在view認可**を検査する。edit失権のみで元targetのviewが残る場合はreceipt照会可。claim時に許可されたことだけを現在許可へ流用しない。
- soft-deleted examも、scoped historical authorizerにより固定元targetの現在認可を立証できればreceiptだけを開示可。実本文の復活/GET修復なし。metadata欠落/不整合、元targetの現在認可を立証できない場合はnot-authorized/403でrefs非開示。
- 未成功bindingも同じtarget解決を行う。C06 success refsがないことを理由に無認可lookupへ落とさない。
- owner側でこのresolverが受領されるまでresult disclosure接続を完了扱いにしない。

#### 正当な自己identity変更と後日移動を区別する固定success target
claim metadataは変更しない。成功時だけC06 `Result.EffectSummary` に次の**内部専用exact string-map**を同最外TXで保存する。既存C06のJSON列を使い、新しい成功target表を追加しない。

`target_schema='exam-operation-target-v1', examination_id, clinic_id, action, command_sha256, before_medical_record_id, before_pet_id, after_medical_record_id, after_pet_id, before_result_state_version, after_result_state_version`

- ID/版はcanonical decimal string。nullable record/petは文字列 `'null'` で符号化し、空文字/0と区別する。after値は**当該Applyが保存したrowを同TXで読んだ値**だけを使う。現在rowからlookup時に再計算しない。deleteは削除直前のidentityをafterとして保持し、deleted_atをidentity変更としない。
- before identity/action/hashはclaim metadataと一致、exam/clinicはclaim/result_refsと一致、after counterはcommitted_versions.result_stateと一致を必須とする。必要監査にも同じbefore/after identity・counter・binding IDを保存する。欠落/不整合はfail closed。counterが進んだことだけで正当なreparentを推認しない。
- consumerのLookupTargetは `Kind='examination-operation'` とbinding参照を用い、保存metadata＋上記成功resultから元対象を解決する。複数targetをincoming DTOに捕捉してAuthorizeLookupへ渡さない。既存operation interfaceを破壊しないowner-private resolverとして実装する。
- **未成功:** before targetの現在view認可を検査。現在examがclaim時と同identityならnot-succeededとして同key再確認可。第三者等がidentityを変更した場合、歴史targetのviewが証明できればnot-succeededの照会自体は可能だが、新Applyはdependency-changedで停止。新identityへmetadataを上書きしない。
- **成功・identity不変:** 固定before=after targetに対する現在viewを検査し保存receiptを返す。
- **成功・当該操作でA→Bへ正当に変更:** 固定before Aとfixed after Bの両方について現在viewが成立すれば元receiptを返す。現在examがBにあることをclaim before A不一致として拒否しない。before Aの認可が失われた場合はBだけ見えても403。新たなwrite grantは不要。
- **成功後、第三者等がB→Cへ移動:** success targetはA/Bのまま。現在rowのCをreceiptへ混ぜない。scoped historical authorizerが、保存claim＋成功EffectSummary＋同TX必須監査に基づく元A/Bへの現在viewを立証できれば元receiptだけ返す。Cの本文/identityは返さない。元A又はBの認可が消失、又はhistory anchorを立証できなければ403。現在Cの閲覧権限だけで元A/B結果を開示しない。
- before/afterのNULL identityは、当該clinicにおける独立/未関連examという明示anchorとして扱う。患者やrecordを架空に補完しない。NULL targetにはそのclinicのexaminations:viewと元exam anchor、非NULL targetには該当patient/recordの現在scopeも検査する。
- historical authorizerはこのD4 consumerの**同PR必要subset**で、current exam identity一致だけの旧readerを流用しない。存在しない場合は成功reparent pathを有効化しない。record/petに新しいrole/grantの追加ではなく、元対象に対する既存view scopeの適用である。

claim時identityと現在identityが違うことだけを理由に、成功後lookupを一律拒否してはならない。保存済み結果を最新状態から作り直してはならない。

### 6.3 receipt
success/replayのresult_refs exact:
- 必須 `examination_id`。
- その操作が新revisionを作った場合だけ `examination_revision_id`。既存revisionを新規成果として偽記載しない。
committed_versions exact:
- 必須 `result_state` = 実DB RETURNINGのcounter。
- その操作がnew revisionを作った場合だけ `examination_revision` = そのversion。
RecordedAtは成功保存時刻。新revisionの有無はaction/statusだけで推測せず実write結果から決める。新revisionなしのcounter値は架空official版ではない。

- owner mutation、counter、items、必要revision、manual usage、必須audit、C06 success resultを同一最外TXでcommit。
- 保存receiptのrefs/version/timeだけをreplayし、最新current counterに差替えない。
- 同key同Content成功済みはApplyしない。stale counter、statusが今confirmed、後日削除等の現在write条件はreplay可否に用いず、現在view認可だけを再検査する。
- same key different Contentはoperation-key-conflict。別URLのexamへ振り替えて元metadataを更新しない。
- 失敗確定後のcounter/fence再確認は同key可。ただし利用者のreread/差分確認を必要とし、送信待ちpayloadに新counterを自動添付しない。
- timeout/parse失敗/5xx/unknown時は `GET /api/v1/examinations/{id}/operations/{operation_key}?intent=<allowlist>` をlookup routeとする。Scopeは現在actor/clinic、targetは保存metadata。UIは元key/intentを保持する。
- HTTP: succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。未認証401は既存前段。
- unknownは自動新key、再create、反対操作、結果clearで解消しない。C06既存unknown解決責務を使う。

## 7. lock / 再読 / fence / 原子性

### 7.1 共通順序
採択516の順序を維持する。新規bindingの短claimに業務lockはない。最外business TXは:
1. Runner Gate（対象ownerの既存shared coordinatorのみ。無関係reservation lockを新設しない）
2. binding NOWAIT、現在認可付きreplay/unknown判定、savepoint
3. 対象lifecycle → 必要なedit-right → 関連record ID昇順 → 関連plan ID昇順 → plan_items dependency
4. 関連job UUIDの固定昇順 → exam ID昇順 → revision/childの安定順
5. 参照master/species等の現在scope/入力依存を既存規約の順で検証・必要lock
6. business write、必要dependent source更新、counter CAS、履歴/usage/audit/receipt、commit

- **D07独立writerも、既存linkやrecordに影響する場合は同coordinatorの必要lockに参加するが、人のD03 edit-right保有を一律要求しない。** view scope、D07 edit/delete/unconfirm grant、D07保護は必須。
- D03本文/planそのものを変更する複合操作だけ494/495の現在holder/epoch/fenceを要求する。link-only grantでresult PUTを代行できない。
- 旧exam→record、job→exam→record、exam取得後の後付けcoordinator進入を残さない。既存サービスのWithTxを呼ぶだけでは参加にならない。
- 関連record/plan/job集合はpreflightで計画し、lock後に実examのJobID/record/pet、active link集合、dependency版を再読。集合追加/変更が分かったら現在TXをrollbackして上位から再計画し、下位lock保持のまま上位lockを追加しない。
- record=NULL・plan関連なしなら存在しないD03 anchorをGET/Applyで作らない。未接続のbaselineはdependency-changedとして対象操作を停止する。
- lock orderで正当な共通順がまだ提供されていないcallerは、該当入口をactivation対象外としunguarded fallbackなし。待ちを501全体や全backendへ拡大しない。

### 7.2 Applyの必須再検査
lock後に、trusted clinic/current actor/grant、active/deleted、元identity、expected counter、現在status、current revisionの存在/種別、専用seal、元/先master相関、関連plan/source dependencyを読み直す。
- counter不一致はversion-conflict、同counterでもscope/依存変化は既存規約でdependency-changed又はvalidation。現在認可不成立はnot-authorized。
- confirmedは本文/結果write不可。revision-less completedのseal、revisionありのreparent/delete拒否、unconfirmの理由/権限等を保持。未知nullable statusをpendingとみなさない。
- 親診療終了だけをdraft拒否にしない。既存目的別訂正/死亡後入力/実績保護の代替をここで作らない。
- exam type/field/species/master rangeの同clinic照合と既存server派生を維持。result_state counterは**保存済み結果aggregate**のCASで、別master設定自体の版を表すとは約束しない。保存時の現在master検証/導出は既存D07の意味を変更しない。
- 実存484 sourceのqualifying contextに影響する場合だけ、採択484/516のauthoritative最終context更新を同TXで1回行う。sourceがなければ生成しない。結果修正を自動実施/請求/通知へ変換しない。

### 7.3 application参加capability
`ExaminationMutationParticipant` は本書で要求する新規consumer adapterの契約名であり、基準sourceに既存の型として扱わない。
- callerは認可済みcommand/expected、対象集合、ambient TXを渡す。participantがlock済みexamと一回消費のmutation handleを返す。
- raw child/restore helperはambient TXとそのhandleを必須化する。nil/未配線/別exam/別TX/二回finishはfail closed。
- repositoryのCreate/Update/Delete/ReplaceItems、revision restore/CAS、SoftDeleteExamByJobを安全なfacade内からのみ呼べるinterfaceへ狭め、外部callerに旧unguarded mutation methodを残さない。移行中のadapterがexpectedをDBから自動充填して旧人入力を通すことは禁止。
- DB triggerで全行更新ごとにcounter加算する方式は本契約では採らない。単一logical operationで一回、revision pointer CASの意味を保持するため。従ってruntime writer閉包の証拠が必須であり、列追加だけで有効化できない。
- 管理用SQL/既存FK cascadeも単調性の外に隠さない。§8のoffline/非writer境界を維持し、live参入させる変更はこのcapability参加又は別採択を要する。

## 8. 全結果writer / 競合参加matrix

| 区分・実入口 | 基準sourceの実write/依存 | 本契約の参加と引渡し |
|---|---|---|
| ReplaceItems / PUT items | exam_results delete/insert、working append、manual usage | 501 C取りまとめ。必須read counter/key、item presence、同TX audit/receipt、1回counter |
| Update / PATCH | 親本文、関係、status、inline items、pet変更再評価 | 同counter。親だけの共有FE hookも対象。items helper内で二重bumpしない |
| Create + initial items/confirm | exam新規、items、場合によりofficial1 | baseline1を一回で作成。途中状態非公開。既存作成key/重複契約を勝手に置換しない |
| confirmFirst / appendWorking / reconfirm | revision append、pointer/status CAS | 元pointer/status CASを保持しresult counterも同TX一回。過去版編集不可 |
| Unconfirm | working from official、mutable restore/pointer | 専用grant/reason保持、同counter、C06。既存確認解除と機器revertを混同しない |
| Delete | soft-delete、履歴/子依存判定 | 削除保護維持、counter+deleted_at同UPDATE、現在view付き削除receipt lookup |
| LabImport PersistExam/PersistBatch | 新規exam/items、duplicate skip | 作成baseline1。既存examの上書きなし。機器/future importが更新を導入するなら必ず参加 |
| LabDevice ReceiveFrames/Attach/PersistLinkedJob | job経由で新規examを1検査種別ごと作成 | job ownerの現在認可/既存受信idempotency保持。D03 session一律なし。top-levelからcoordinatorを先に計画 |
| LabDevice Detach/DetachDeviceJob | job→exam soft-delete、retraction、job遷移 | 削除対象全examを安定集合でlock/再読、各counter一回、同TX。元job補償receiptとC06を二重成功commitにしない |
| LabImport Revert | job補償、usage再検査、retraction、SoftDeleteExamByJob | 上記同順序。既存理由/idempotency/補償保護維持。既存job成功receiptはjob ownerで維持しD4共通receiptへ無断置換しない |
| revision repository direct restore / ReplaceItemsByExamID / SoftDeleteExamByJob | 低層raw writer | public安全入口ではない。ambient capability必須、親lockなし直接呼出し拒否 |
| 516 E-link / replace / unlink | plan関連/版、usage、場合によりsource generation | 516既決contract。結果counterは不変。examとchildの同snapshotをlockで検査し、result writerも同coord参加 |
| clinical GET/items/print/lab report | usage receiptのINSERT、結果read | 結果counter不変。job→examと現在scopeでrevertと直列化。新GET tokenのsnapshot整合を確認 |
| checkup結果 / image.exam_id / billing参照 | 別owner row・結果への参照 | exam結果writerではないが、削除/付替え/active link判定の依存writer。必要coordinator/関係保護を維持し、D4で内容ownerを奪わない |
| exam_type_fields / reference ranges / staff | master更新、結果の参照検証 | 保存済み結果の書換は行わない。field削除は結果/range使用中拒否、staffはsoft-delete。master変更を全結果再計算へ広げない |
| FK ON DELETE SET NULL / CASCADE | ERDにはdoctor、field、record/exam cascadeが存在 | 通常runtimeからのphysical deleteを既存拒否/soft-delete境界で閉じる。liveで暗黙結果変更を許す入口が見つかれば参加設計前にactivation不可。FKを無断変更しない |
| csvimport/import、cutover、synthetic fixtures | SQL/COPY/DELETEのout-of-band writer | disposable/offline管理用途として通常API並行writerへ数え替えない。稼働環境で同時実行不可を運用ownerが証拠化。復元/cutover後は旧counter/key/cacheを持ち越す運用を禁止する設計引渡し |
| 501他院provenance/将来採択writer | 現wire未選定部分あり | 既存A/Cの受入に本capabilityを追加し、自前counter/別保存台帳を作らない |

機器detach/revertは、人のstaleフォームの期待値とは異なる**server-side guarded batch**として扱う。job lockと既存job preconditionを取得後、同一TXの安定対象集合を読み、各examのcounterを取得して条件付き更新する。手動編集によるmanual_mutation receipt/confirmed/既存downstream useなら既存の拒否を維持する。これは人のPUTへのexpected自動補完を許す例外ではない。

上表は基準sourceのwriter調査に基づく。任意の動的SQLや実装branchで追加された経路まで網羅したという証拠ではない。activation前に実装ownerが対象SHAのcallsite/SQL一覧を更新し、各経路を参加済み又は明示非対象へ分類する。未分類を0にする。

## 9. 必須監査
既存audit_logsを再利用し、結果PUTの「削除行があった時だけ」のログで済ませない。初回insert、同値replace、明示clear、parent/child複合にも同成功単位の監査を残す。
- resourceは既存examination。actorは認証staff又は採択済みsystem principalであり、nil actorで人操作を通さない。
- 新actionラベルは `examination.replace_items`（基準sourceの既存定数として扱わない）。他actionは既存値を維持できる。
- old/newは保存された結果aggregateのsnapshot又はimmutable revision参照。unversionedでも旧項目集合が復元可能な値をauditへ残す。新しい臨床revisionを捏造しない。
- metadata exact D4追加subset: schema_version='1', operation_binding_id, operation_key, command_sha256, before_result_state_version, after_result_state_version, previous_revision_version, new_revision_version。非該当revisionはNULL。既存operation_type/reasonを保持する。
- snapshot対象は§4.2のparent/全items、ID/NULL/ExactNumericを保存形どおり扱い、現在masterから後日再構成しない。secret/raw device frame/接続情報を追加しない。
- mandatory auditとmanual_mutation usage receiptの失敗はbusiness/counter/成功receiptをrollback。slogやnil-safe auditを代用しない。
- C06 receipt自体に検査値全文を保存しない。監査と結果照会権限を混同しない。

## 10. 最小DDL / model / API handoff（未実施）
既存DB ownerの446専有・466取りまとめへ提示する単体exam consumerの最小差分は次の2単位だけ。
1. examsへresult_state_version列/check/default。対応typed model/read columns/DTO。既存current_revision FK・NULL・job FK等を変更しない。
2. examination_operation_targets表とbinding複合FK、immutable制約、tenant/RLS適用。business FKなし。既存operation3表や516plan target tableを重複作成しない。

- 新DDLの配置/採番は、**001統合後の最新mainのmigration規約と446専有・466取りまとめの単一owner**へ従う。採番直前に最新mainと並行branchを照合し、本書ではファイル番号を予約しない。bd309de時点の [migrations/README.md](https://github.com/MinoruSoga/AnimalEkarte/blob/bd309de7b9680a46fdf3ed4a89491c198624288c/backend/migrations/README.md) / [migrations/CLAUDE.md](https://github.com/MinoruSoga/AnimalEkarte/blob/bd309de7b9680a46fdf3ed4a89491c198624288c/backend/migrations/CLAUDE.md) は003〜010統合済み、次の新規番号は011からという基準を持つが、これを将来の確定採番とはしない。
- **001_init.sqlへの無条件追記、旧003〜010の復活、独自migration系の追加を指示しない。** 凍結前の001整合は当該範囲の明示承認がある場合だけ、凍結後は新連番という現規約を維持する。DB ownerが最新の凍結/承認状態と適用履歴を確認して配置を決める。適用済みchecksumの書換え・自動resetで回避しない。ERDはownerが後工程で再生成する。本書の採択はDDLパッチ作成・生成・適用・採番の実施を意味しない。
- existingrows baseline1とnewcreate1の移行、既存data/nullable維持、counterpositive、上限、FK/RLS/metadataimmutabilityを同schema受入に入れる。
- 稼働DBに既に列があるとは主張しない。データ量/適用時間/rolling deployの安全性は未確認。active旧writerを残したままcounterを信頼するdeploymentは不可。
- OpenAPI/router/codegen共有ownerには§5/§6のschema・4write＋lookup・read加算変更のexact差分を渡す。同時にfrontend feature API、shared use-update-examination、types/transformsとunknown回復adapterを切り替える。
- 501が全shared pathを直接編集する意味ではない。現claimの実編集ownerと順序は452/445窓口で確認する。本書だけではclaimは付与されない。

## 11. 既存stageの独立受入とactivation
新票/新stageは作らず既存責務のchecklistとして分ける。

| 既存受渡し先 | 本契約で受け持つ最小subset | 独立成果と次の受渡し |
|---|---|---|
| 501 C（単一取りまとめ） | D4 counter/guard/全writer inventory/receipt/監査 | 技術契約・fixture期待値の先行PRは可。501 Aの他院wire、B感染症summary全完成待ちにしない |
| 446/466 | §10の2DDL単位とmodel/ERD単一編集 | schema成果を501 Cへ返す。既存727等と重複しない |
| 489 | 現C06/Q/claim hooks/lookup/unknown必要subset | helper再実装ではなくD07 adapter適合レビュー。C06全consumer Done不要 |
| 516 S1/S1a | coordinator・reverse dependency・owner ambient参加 | E-linkが依存するresult writer閉包だけを受領。接種/健診全完了を待たない |
| 516 E-link | 既存link guard/read/usage/owner closure | link-only contractは再選定しない。D4 result counterはplan版と別 |
| 495 D4/D1 | 旧draft helper置換caller、必要目的別guard | 本契約の必要部分だけ受領。D1/D3既決作業を止めない |
| 現D07/機器owner（297/113/268等の既存境界） | create baseline、detach/revert、usageとlock参加 | 機器仕様/臨床承認/実機受入を再実装しない |
| API/FE shared owner | exact wire＋caller切替＋result lookup | 合成契約成果と本番activationを分ける |

activation条件:
- counter+metadataのschema/model/readが同版である。
- §8のlive writerの参加又は明示非対象証拠が揃う。未参加入口はfail closedであり、unsafe互換がない。
- §5のread snapshot、必須guard、caller、C06 lookup/unknown、audit/usageの同TXが接続される。
- ブラウザwriteは§5.5のGo/Worker両preflight最小差分とAPI/FEを同版受領する。両preflight fixtureの期待値・未実行statusを明示し、片側だけの受領や非ブラウザ成功をactivation証拠にしない。
- 必要coordinator subsetだけが受領され、現在scope/identity再読がある。
- 受入試験を後日本人指示時のmainローカルで実施する。設計・実装・試験状態を別に管理し、現時点の試験状態はNOT_RUNとする。NOT_RUNだけで「コード未実装」とは断定しないが、試験PASS/運用activationの証拠にも使わない。

### 11.1 限定実行ID・対象AC・残ACの対応
下記IDは**実行packetの候補識別子**で、正式Plane IDでもclaim取得済みIDでもない。新票/新stageを作らず、既存stageへ対応付ける。取りまとめownerとclaim窓口が限定scopeと残AC対応を確認し、既存ownerの実行ID/claimへ正式対応を記録するまでは実装claim/Running/PR作成を開始しない。親EMR-501/516/495の全ACを限定成果PRの完了対象にしない。特にPRタイトル/関連付けで親番号を自動完了相当へ流す運用は不可。

AC番号は§12のcase番号を `D4-AC-01`〜`D4-AC-23` として参照する。これは実行PASSではなく期待値の識別子。

| 実行packet候補 / 既存stage | 成果型・閉じる範囲 | 完了証拠 / 対象AC | 閉じない残AC・受領owner |
|---|---|---|---|
| D4-DESIGN-R3 / 501 C | 本書の採択済み契約。exact wire/保存/lock/全writer参加/identity replay/CORS受領と未選定境界 | 本書の固定版、§14の根拠、全ACの期待値 | schema/adapter/runtime/全試験/activationは未完。後続=下記packet。識別子は正式実行IDではない |
| D4-SCHEMA-01 / 446専有・466統合 | result counter＋exam claim targetとmodel/ERDの明示先行成果 | 2DDL単位、NULL/FK/RLS/immutable/初期値/上限、供給SHA、D4-AC-16/17のfixture作成 | DB適用/consumer接続/機能公開は含まない。501 Cが受領。正規生成は本人実行/成果受領 |
| D4-OWNER-01 / 501 C、516 S1の必要受領subset | atomic read・counter participant・historical lookup authorizer・低層writer閉包のlibrary先行成果 | 実file/symbol/ambient contract、全callsite matrix、D4-AC-03/04/08/09/10/12/14/16/19–21のfixture作成 | HTTP/FE旧caller切替、job回復未接続なら機能activation不可。501 C/516 E-linkへ引渡し |
| D4-WIRE-01 / 501 Cと共有API/FE・Go CORS/Workerの現owner | 4mutation＋lookup＋readschema＋全旧caller切替、§5.5の両Allow-Headers最小追加を同版受領する縦割り機能成果 | handler/router/DI/typed Q/C06/FE cache・unknown回復接続、D4-AC-01/02/05/06/07/15/22/23のfixture、旧無guard到達0、両preflight差分の採用SHA | D4-OWNER/SCHEMA必要subset待ち、job未参加の該当入口は停止。Go/Worker片側不足ならブラウザactivation停止。501 A/Bの新provenance/summaryは残る |
| D4-JOB-01 / 現機器job owner＋501 C受領 | job→exam counter/lock参加、既存RevertとDetachの成功/unknown契約 | D4-AC-11、§13.1 exact契約の受領SHA、全対象exam同TX/監査、existingreceiptとC06責務の対応 | §13.1は未選定。受領前に機器解除までclosedとしない。独立受信/閲覧全体は待ちにしない |
| D4-LINK-01 / 516既存S1/E-link | 既採択link-onlyとD4参加閉包の結合 | 516 exact契約不変、D4-AC-13/18、関連writer closure/usage/必要source generation | 接種517/健診500/501 A/B全体のDoneは不要。機器解除など必要未受領だけactivation gate |

**execution IDの最終対応は未選定:** 本書は新規Plane子票や既存親票のscope変更を指示しない。上表の候補をそのまま `claim/<候補>` に使わない。452/445の現claim窓口と取りまとめownerが、既存適切実行票又は明示別実行IDへ対応を記録する必要がある。対応未記録の部分は実装開始条件未成立であり、NOT_RUNへ隠さない。

### 11.2 契約採択と必要subsetの供給状態
| subset | 現在状態 | 固定根拠 / 供給・受領 / 停止範囲 |
|---|---|---|
| C06 Scope/Content/Result/claim hooks/Q | 存在済み | 04a9078 operation/contract.go,store.go,runner.go,verified_command.go。D07接続済みとはしない |
| D07 seal/revision/storage意味、516 link-only | 採択契約は存在済み | 04a9078 lifecycle/technical/link契約。現在runtimeとの差は§3。閉じた業務契約を未選定へ戻さない |
| D4専用counter/target metadata/header/PATCH正規形/成功target/CORS受領 | 採択済み | 本書の固定契約。実装済み・供給済み・activation済みとはしない |
| schema/model/ERD必要差分 | 別owner待ち（D4-SCHEMA-01候補の供給範囲） | 446/466、現受領SHAなし。counterを使うruntime停止。架空columnを前提にしない |
| D07 read/participant/history authorizer | 契約採択済み・必要実装未供給 | D4-OWNER-01候補の同PR必要scope。正式実行ID/claimとの対応は未完。500/495の既存helper名だけを実在証拠にしない |
| 516 coordinator/dependency/source participant必要subset | 別owner待ち | 516 S1/S1a、必要runtime受領SHA未確認。影響あるwriter/link activationだけ停止 |
| machine API/FE/generated正規成果 | 別owner待ち | shared API ownerと本人codegen、受領SHAなし。手書き生成型/stubで代用しない |
| Go/Worker両preflightの新header許可 | 別owner待ち（D4-WIRE-01必要subset） | 04a9078では両方未許可。現共有ownerへ§5.5の最小差分を要求し、同一採用SHAと両fixtureを受領するまでブラウザactivation停止。変更・受領は未完 |
| Detach job target/key/unknown契約 | 未選定 | §13.1。現job ownerへ限定設計要求。revert receipt存在を代用しない |
| 検証 | NOT_RUN / 判定UNKNOWN | 試験実行は延期中。fixtureは本書に期待値を記述しただけでコードfixture未作成。試験未実行と未実装/未接続を分ける |

本書は契約成果であり、実装PRの供給証拠は含まない。実装ownerへ渡す際は本書の採用SHA・正式実行ID・限定scope・供給/受領owner・残ACを固定し、各subsetを「存在済み／同PR内に実装予定／別owner待ち／未選定」に分類して更新する。既に採択した契約を供給待ちのため未選定へ戻さない。

## 12. 固定受入case（期待値。未実行）
1. unversioned v7を二人が読む。同一itemsで別keyを同時PUT。先行v8のみ成功、後行version-conflict。last-write-winsなし。
2. PUTのmissing/nullは400・zero-write。[]は許容statusだけでclearしv+1。completed revision-less/confirmedは拒否。
3. parent-only PATCH、inline items、pet変更再評価、working appendのどれも旧PUT tokenをstaleにする。複合PATCHは一回だけv+1。
4. 初回confirm/unconfirm/reconfirm/restoreは既存revision履歴と新counterが同TX。revision番号とcounterが異なっても正常。過去official不変。
5. 同key成功後、current headが進んだ/confirmed/soft-deleteされた後でも、現在viewが許せば元receiptだけreplayしv不変。
6. 同key別payload/別examはkey-conflict、metadata不変。未成功bindingでも現在viewを検査し漏えいなし。
7. response loss→lookup succeeded/replayed。同key再送でitems/revision/usage/auditが重複しない。unknownは新keyを発行しない。
8. audit/usage/revision/result保存の各故障で全business/counter rollback。bindingだけはC06既定の失敗状態として残り得る。
9. item child raw writerが親lock/capabilityなしなら拒否。親だけcounter進行・子だけ更新の片落ち不可。
10. GETとPUT/restore競合で、返却itemsとcounterが同snapshot。別cache世代混合時は全reread、tokenだけ更新不可。
11. job-linked manual PUTとdetach/revertが同時。job→exam順とmanual receiptで直列化、既存使用済み補償拒否を維持。新たなD03 session要求なし。
12. jobID/record/active link集合がpreflight後に変わる。下位lock保持で上位追加せずrollback/replan。別患者結果へ誤適用なし。
13. link-onlyはplan版のみ進みresult counter不変。結果PUTはlink digestを受け付けない。旧fixed revision参照と現在headを混同しない。
14. master/field他院参照、unknown status、欠落revision、nil audit/coordinatorで成功しない。NULL/ExactNumericの意味を保存→再読で保持。
15. missingexpected旧FE/direct HTTP、shared parent-only hook、inline PATCH、DELETE/unconfirm旧呼出しがguardを迂回できない。
16. 同値writeは新keyで+1、replayは+0、duplicate import skipは+0、新createは1、counter上限拒否。
17. schema導入時unversionedはrevisionNULLのままcounter1。read/lookupでDDL的初期化・repairなし。
18. read-only/contract/library PRは成果として評価できるが、全writer参加/機能activationと別チェック。実機受入/臨床mapping承認へ機械fixtureを流用しない。
19. 正当なunversioned PATCHでA→Bへ変更、同じ送信の再送/lookup。A/B両方の現在viewありなら元成功receipt、counter/revisionは再進行なし。before一致を再要求しない。
20. 上記成功後の第三者B→C移動。固定A/B view+history anchorありなら元receiptだけ、Cの値は非開示。Cだけview可能又はAのview失権なら403、元receipt再計算なし。
21. 未成功bindingの間にA→B移動。元A viewがあればnot-succeeded照会可、同key新Applyはdependency-changed。metadataをBへ自動更新しない。
22. PATCH全9fieldのomitted/null/値、items NULL対[]、unsupportedfield、duplicate key、非canonical ID、未知日付precisionを組み合わせ、canonical commandと実writeが一致する。read cacheのschema/clinic/versionを混ぜない。
23. Go/Workerの両preflightで新expected-version headerを許可し、既存origin/credentials/CSRF/methodsは不変。WorkerのOPTIONSはContainerを呼ばない。片側だけ許可した構成はブラウザactivation不可。§5.5の許可Origin/非許可Origin・4method fixtureを各側に用意し、実行はNOT_RUNのまま区別する。

## 13. 移行上の差分・未完・未選定
**採択済みの技術契約**
- 501 CをD4単一取りまとめownerとする。個人担当・現claim・正式実行IDは別途確認する。
- 専用counter、単体exam consumerの2DDL単位、no-op規則、exact read/write/receipt/lookup、PATCH正規形。
- 不変claim metadataと当該成功時の固定before/after target、現在view認可、履歴anchorによるreplay。
- atomic read、全mutable writerの一回finish capability、lock/再読/監査/receiptの同一commit、旧unguarded route閉鎖。
- Go/Worker両preflightの最小header追加を共有ownerから受領する契約、両fixture期待値、ブラウザactivation条件。

**残る実行handoff（本書の採択だけでは完了しない）**
- 452/445の現在path claim、446/466の具体的schema差分受領、API/codegen ownerとリリース順。
- 516 S1 coordinatorのruntime受領、全live writerの実配線/ambient TX/現在認可・未知結果照会の適合。
- 機器Detachの操作key/耐久成功結果/unknown照会は現sourceでは未提供。Revertのlab_import_revert_receiptsをDetachにも存在すると見なさない。詳細は下記§13.1。
- model移行残（543）とD07nullable/ExactNumeric adapterの必要subset。compile/testsは未実行。
- 実DB適用状況、運用import/復元経路の稼働隔離、権限/RLS/本番activation。実環境の確認証拠なし。

**D4から切り離した未選定**
- 501 A/Bの他院provenance exact wire、mapping管理、summary endpoint/cache世代の最終選定。D4 counterはそれらの臨床意味を決めない。
- 495 D2見積削除等の本人業務判断。本契約のブロッカーへ混ぜない。
- 不明結果の患者/record修復、死亡後新入力、新しい感染症判定の業務操作。本契約で新造しない。

**既存記述との衝突**
- old PUT nil=>clearと既存FE配列応答は本書のwireと非互換。明示cutoverが必要。
- sourceのexam→record / job→exam→record順は採択516 coordinator順と非互換。helper追加だけでは解消しない。
- currentrevisionと新resultcounterを統一したくなるが、clinical semanticsが違うため不可。
- 516のlink digestに新counterを足す/置換する案は含めない。採択済みexam-link-source-v1のexact projectionを変えない。
- Q未実装とする旧待ちを引き継がない。固定mainのVerifiedCommandは存在する。ただしD07 consumer adapter完成とは別。

### 13.1 job ownerへの狭い残契約要求（未選定を維持）
固定sourceの `labDeviceReceiveService.Detach` はjob cardを返し、`DetachDeviceJob` は成功receiptを挿入しない。一方 `Revert` は `lab_import_revert_receipts` を保存する。従って「全機器解除にも既にC06 replayがある」とする記述は誤りになる。

- D4から引き渡す決定候補は、**結果を消すDetachもjob ownerの単一最外operationに含め、examのcounter/補償snapshot/監査と耐久成功結果を一度にcommitすること**。各examで別Runner.Executeを開始しない。
- job ownerが受け取る入力: trusted clinic/actor、元job UUID、操作action、元idempotency key、現在jobの許可状態。exam集合/期待counterはjob lock下のscoped再読から決める。返す成果: 元jobへの不変target解決、成功時のretracted exam ID集合と各counterの固定参照、現在認可付きsame-key lookup、unknown時の停止。外部機器送信なし。
- exact job Scope/claim metadata/revert既存receiptとの写像を本D4だけで独自採択しない。特にexamination_operation_targetsはexam単体用であり、複数examやexamゼロのjob targetへ捏造しない。job用metadataを増やすなら既存job ownerと446/466へ別の限定差分として戻す。§10の「最小2DDL」は単体D4 consumerの差分であり、job owner追加schemaまで閉じた主張ではない。
- 採択/受領前は、この未接続Detach経路をD4/516の安全なactivation証拠に含めない。**該当Detachだけfail closedのsubsetを選んで単体結果編集を先行する案**を提示する。独立機器受信/閲覧/新規保存を一律停止する案ではない。
- Revert既存receiptについても、現在認可付きlookupの実route/未成功target/unknown回復の同等性は未検証。既存receiptの存在だけでC06全契約適合としない。counter/lock参加は本契約で指定できるが、結果回復の完全適合はjob owner受領のチェックを残す。

この残点は技術調査不足を隠すための「実装時に検討」ではなく、単体exam consumerとjob補償ownerの境界にある具体的な採択待ちである。取りまとめownerと現job ownerが、job側の次の限定設計対象又は当該入口を停止した先行subsetを明示して選ぶ必要がある。

## 14. 根拠索引
基準sourceと既存正本・Plane記録。sourceは実装差分の根拠であり、臨床業務仕様を上書きするものではない。

- [495 D4/既決D1の票](https://app.plane.so/baritechllc/browse/EMR-495/) コメント f7411e88-eb92-409a-a8a0-041b429c2d03。
- [501責務](https://app.plane.so/baritechllc/browse/EMR-501/) / [516責務](https://app.plane.so/baritechllc/browse/EMR-516/)。
- [D07 lifecycle](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/spec/domains/examinations/lifecycle.md)
- [D07 technical spec](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/design/domains/examinations/technical-spec.md)
- [516/517採択link契約 §§4–9](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md)
- [C06 participation](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/design/shared/operation-participation.md)
- [結果model](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/model/examination_record.go)
- [結果PUT](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/medicalrecord/examination_items.go)
- [親/inline/revision update](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/medicalrecord/examination_service_update.go)
- [revision workflow](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/medicalrecord/examination_revision_workflow_repository.go)
- [lab補償](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/medicalrecord/lab_import_revert_service.go)
- [usage tracker](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/medicalrecord/lab_import_usage_tracker.go)
- [Q実装](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/operation/verified_command.go)
- [C06 target/participant interface](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/backend/internal/operation/contract.go)
- [exams ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/architecture/erd-detail/exams.md)
- [exam_results ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/04a9078d93c3899705ce9d420eae75e89709e2e3/docs/architecture/erd-detail/exam_results.md)
