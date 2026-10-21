# EMR-498 マスタ側 DB/API 選定契約 R2

2026-10-21。親レビュー採択。本稿は①DB/API契約の技術選定。②マスタ実装開始ではない。先のconsumer案は④として分離し、本稿の完了条件に含めない。基準main=755f77abd2e6d63a2fc608fcfe6a518a61afcfff。

## 現状と範囲

DDLは複数pattern、medicine_dose_pattern_revisions不変履歴、治療側revision FKを収録済み。DB適用は未確認。medicine_dose_param Go modelはpattern列/versionを保持するがJSON非公開、revision modelは `backend/internal/model/storage_clinical.go:276` の MedicineDosePatternRevisionとして存在する。既存modelを再作成せず新reader/writerへ再利用する。api.yamlはmedicine/species単一upsert/deleteと旧double DTOだけで、複数pattern/revision DTOは未収録。旧実装を新仕様の正本にしない。
今回の閉じたmaster subset: IDで区別したpattern作成/全項目更新/論理削除、同TXの不変revision append、現況/履歴reader、C06 lookup。臨床計算・treatment明示採用・488/494/484consumer接続は④。実際の医師承認登録は既存#261の証拠に従う別subsetで、今回の編集成功を臨床承認としない。

## 1. API

全pathは/api/v1 prefixを除く。
GET /masters/medicines/{medicine_id}/dose-patterns
GET /masters/medicines/{medicine_id}/dose-patterns/{pattern_id}
GET /masters/medicines/{medicine_id}/dose-patterns/{pattern_id}/revisions/{revision}
POST /masters/medicines/{medicine_id}/dose-patterns
PUT /masters/medicines/{medicine_id}/dose-patterns/{pattern_id}
DELETE /masters/medicines/{medicine_id}/dose-patterns/{pattern_id}
GET /medicine-dose-pattern-operations/{operation_key}?intent=create|update|delete

既存権限を再利用。GET/lookup=master-medical:view、POST/PUT=master-medical:edit（旧upsertによる作成/更新がedit）、DELETE=master-medical:delete。新しいapprove/doctor-only/restore権限を作らない。医院は認証contextから確定しmedicine/pattern/版の医院・親薬剤を各requestで検査。
一覧queryはspecies任意(dog/cat)、include_deleted任意boolean(default false)。薬剤内の全該当patternをpattern_id昇順で返す非ページング初回契約。暗黙limit/先頭1件/名前重複排除を禁止する。途中取得失敗を部分成功にしない。個別/固定版readは論理削除後も現在view権限下で可能。GETはwrite/旧履歴生成をしない。
GET応答は {medicine_id, items:[Pattern]} 又は単一Pattern/Revision、404等は既存ErrorResponse。結果照会は共通ResultResponseそのまま。全read no-store。

Pattern closed fields:
pattern_id, medicine_id, species, version, pattern_name, purpose, administration_route, parameters, source_reference, deleted_at, current_revision。
current_revision=nullは旧行にimmutable revision未作成の場合だけ。欠損・不整合を最新成功版へ推測しない。存在する場合はRevision完全形を返す。versionとcurrent_revision.revisionの一致を検査する。現行rowとrevisionの投影不一致はsource-unresolvedで失敗する。通常更新後は必ず一致。
Revision closed fields:
pattern_revision_id, pattern_id, medicine_id, revision, pattern_name, purpose, administration_route, parameter_snapshot, source_reference, approval_reference, approved_by, approved_at, is_active, recorded_at。
approvalの3列は独立に省略せずnull。is_activeはその改訂時点の状態で、今の採用適格性や現在薬剤状態の保証ではない。古い履歴を現在値から補完しない。

## 2. Write DTO / exact型

POST bodyはoperation_keyと以下の全business field:
species, pattern_name, purpose, administration_route, parameters, source_reference。
PUTは同fieldにexpected_version。speciesも全置換項目だがmedicine_id/clinic_id/pattern_idの変更は受け付けない。既存の種別・数量・丸めvalidationを維持し、同薬同種の複数目的/経路を許す。名称/目的/経路の一意制約や必須性を勝手に増やさず、DBどおりnullable。空文字とNULLの既存差を保つ。
DELETE body={operation_key,expected_version}。論理削除のみ。restoreは本subsetに含めず、削除済みへのPUTを復元代用にしない。
parameters={dose_basis,dose_per_kg,min_mg_per_kg,max_mg_per_kg,absolute_max_dose,rounding_step,rounding_mode,notes}。すべてkey必須、nullable列はnull可。dose_basisはper_administration/per_day、rounding_modeはup/down/nearest又はnull。notesは既存text NOT NULLで空文字可。
ID/version/revisionは正signed64 canonical decimal string、最大9223372036854775807をexact検査。operation_keyは共通canonical UUID。投与量はcanonical decimal string（指数/非数/余分な0/負の0禁止）でbinary floatを通さず、dose/min/max numeric(10,6)、絶対上限/roundingstep numeric(10,4)を丸めなしで検査。正負/相関/組合せ条件は既存dose validatorを仕様に照合して維持し、精度方式変更から新臨床値を導かない。unknown fields拒否、omitted/nullを混同しない。
parameter_snapshot={schema:"dose-pattern-parameters-v1",medicine_id,species,...parameters}。入力をstrict decodeしたimmutable commandからserverが生成し、clientにJSON snapshotやapproved_by/at/reference/is_activeを指定させない。source_referenceは正確text又はnull、medicine_dose_params.provenanceへ同値を保存する。

## 3. Revision append / 承認境界

POST: version=1でpattern作成、同revision=1、不変snapshot、is_active=true、source_reference、operation_idを同TX。
PUT: lock済みpattern.versionをvとしてexpected_version=vを検査する。内容変更時はentity_version_guardが確定した保存後version（v+1）をRETURNING等で取得し、新revision番号にその保存後versionを用いて完全な新snapshotをappendする。現行revisionがある場合は事前にcurrent_revision.revision=vを要求する。source/計算field/名前/目的/経路を含む内容変更の新revisionはapproval_reference/approved_by/approved_atをすべてnullにする。旧承認を別内容にコピーしない。これは新しい承認権限でなく『承認は固定版に結び付く』既決の実現。
DELETE: 未削除patternのlock済みversion=vと期待版を照合しdeleted_atを変更、entity_version_guardが確定した保存後version（v+1）を新revision番号にして最終snapshot/is_active=falseを同TXでappendする。現行revisionがある場合はcurrent_revision.revision=vを要求する。歴史の旧承認は旧revisionに保持。削除revisionは新しい臨床承認としないためapproval3列null。既存治療FK/snapshotを削除・付替えしない。
現行revisionがある完全no-op PUT／既に削除済みで現行revisionを照合できるDELETEは版増分/新履歴なしで現在の結果をreceiptへ。削除済みlegacyでrevisionが無ければsource-unresolvedとしてzero-writeにし、架空のrevision参照や成功receiptを作らない。source_referenceやnull/空の差は実変更。
旧行(current_revision=null)へ明示PUTし内容が同値の場合は、既存entity_version_guardを維持し、patternの現在versionを変更せず明示PUT直前のlock済みpattern.version=vを初回revision番号として不変revisionをappendする（行versionもrevisionもvであり、1やv+1へ補完しない）。これは今回の操作で確認した現況の未承認snapshotであり、recorded_atは今回のserver時刻、operation_idは今回のbindingとする。過去時点の履歴をbackfillしたとは扱わず、昔から承認済みと見せない。内容も変わるlegacy PUTは通常どおり保存後version=v+1をrevision番号として新内容をappendし、旧状態のrevisionは捏造しない。同時要求はmedicine→pattern lockと既存UNIQUE(clinic_id,pattern_id,revision)で直列化し、二重初版を作らない。移行時の一括backfill/ダミー承認/架空operationはしない。
採番の唯一の基準はそのpatternの保存後versionであり、MAX(revision)+1やnullable current_revision+1ではない。entity_version_guardを唯一のpattern版writerとして保持し、アプリで二重加算しない。現行revisionとpattern版の不一致・同じ番号の異内容・版overflowはzero-writeで拒否する。現在revisionを照合できるno-opは既存版/既存revisionを返し、legacy同値PUTの初回現況保存はpattern UPDATEなし/revision=vのINSERTだけという限定例外である。

既にある承認済revisionをreaderが返すことと新承認を登録できることを分ける。承認資料の真正性/医師/単位/具体値は#261。新writerには承認登録APIを含めず、新規/変更patternは臨床自動計算にまだ使えない。④consumerは現在の承認・scope・固定版を必ず検査する。

## 4. C06 / 元対象anchor / DB最小差分

共通Runner/[VerifiedCommand](../../design/shared/operation-participation.md#verified-command-selected-20261021)/ClaimTarget/ResultResponseを再利用。owner_kind=medicine_dose_pattern、actor_kind=staff、intent=medicine_dose_pattern.create|update|delete。共有identityはoperation_keyのみ。
新規は親medicineとC06 bindingが元target、既存はmedicine＋patternが元target。成功receiptが無い失敗でも認可元を解決するため、owner-private metadata表 medicine_dose_pattern_operation_anchors を最小追加候補とする。他のownerのanchor表を改名・流用しない。
列: operation_binding_id bigint PRIMARY KEY, clinic_id bigint NOT NULL, medicine_id bigint NOT NULL CHECK>0, pattern_id bigint NULL CHECK(NULL or >0)。UNIQUE(operation_binding_id,clinic_id)、FK(binding,clinic)→operation_bindings(id,clinic_id) ON DELETE RESTRICTのみ。業務medicine/pattern FKはclaim lock順を壊すため付けない。createならpattern_id=NULL、update/deleteは正値をclaim factoryで検査。RLSは既存医院context、UPDATE/DELETE拒否、通常INSERT/SELECTのみ。既存revision表の複製ではなく未成功lookupの不変metadata。
claim winnerだけがbinding+anchorを短い同TXで作る。loser/既存bindingのanchorをincoming対象で補完しない。lookupはbinding→元anchor→元medicine/patternの現在scope/view→receipt、元actor/familyも共通規約で照合。論理削除済みは原行で認可。元anchor欠落・解決不能等で元対象の現在認可を成立させられない場合は共通not-authorizedで非開示。元対象の現在認可が成立した後に結果データ破損・取得不能を検出した場合だけresult-unavailableとする。incoming対象や成功receiptから元anchorを捏造しない。
Contentはschema="dose-pattern-command-v1",medicine_ref,pattern_ref(createはliteral "new"),command_sha256の短いstring map。immutable commandはintent,medicine_ref,pattern_ref,business payload（deleteはnull）を正確JCS化しdigest。operation_key/actor/clinic/期待版/時刻/生成IDはdigestへ入れずscope/Preconditionsへ。VerifiedCommandはmemoryのみ、既存bindingの再送はclientから同commandを再提出しhash検査。耐久command保管や第二request IDを要求しない。
判定は元対象現在認可→同Content replay/unknown→Apply再検査。unknown/in-progressは別key救済不可。CASだけの再確認とbusiness変更を区別する。共通既存12status/HTTPを変更しない。本書のsource-unresolvedは新しいC06 statusではない。GETで現況とrevisionの不整合等を検出した場合は409の既存ErrorResponse {error:"source-unresolved"}。Applyで同条件を検出した場合は共通status=validation（HTTP 400）の内部理由へ写像し、公開ResultResponseへ独自status/reason fieldを追加しない。元対象認可が成立しない場合はこの理由よりnot-authorizedを優先する。

## 5. 同一TX・lock・audit

短いclaim終了後、業務TXで既存契約が要求するContextGate→C06 binding NOWAIT→元対象現在認可/replay→savepoint→medicine row→pattern rowの順。必要なContextGateは必ずbindingより前に取得し、そのderived ctxを後段へ渡す。予約技術仕様§6.4等で関連clinical masterに既に要求されるgate参加を、今回treatmentへ直接writeしない理由で解除しない。一方、既存契約が要求しない経路へ新たな一律graph gateや編集権を追加しない。親medicine rowを新pattern作成/全pattern更新共通のserialization pointにする。clinical reader/writerは④の上位coordinator保護を先に取り、その後medicine→patternに合わせる。master側はtreatment/plan rowを後取りせず、過去snapshotを更新しない。
現在親存在/医院/薬剤設定・pattern親/削除状態・CASを同TXで検査。既存medicine設定writerもmedicine lockを共有しない限り競合安全なactivationとしない。整数overflowはzero-write失敗。
pattern UPDATE/INSERT、revision INSERT、必須audit、成功receiptを同一TX。audit action=medicine_dose_pattern.create/update/delete、resource=medicine_dose_pattern、resource_id=pattern bigint、old/newの版・revision ID・snapshot digest・元親を既存audit JSONに格納。actorは認証から、失敗で全rollback。audit schemaの資源ID型を変更しない。GET/replayで履歴/auditを増やさない。
refs={medicine_id,pattern_id,pattern_revision_id}、versions={dose_pattern, dose_pattern_revision}をcanonical文字列で返す。no-opは既存revision参照、成功effectがないのに新revisionを生成しない。

## 6. 互換・1PR境界・fixture

旧species単位PUT/DELETEと旧自動先頭選択を新master cohortの書込迂回に残さない。新FEのID単位CRUD/strict DTOへの移行と同じactivationで旧writeを明示非対応にし、再読込を促す。互換用に勝手にpatternやkey/期待版を生成しない。旧readは移行対象を明記し、旧DTOで複数を単一に縮約しない。
master一式PRはapi.yaml selected定義、既存revision modelの再利用・既存DDLとの整合、新DTO/handler/reader/writer、anchor最小DDL/model、C06共有subset受領、設定FE、fixture記述まで。治療計算/採用writerをこのPRに混ぜない。未実装共有hookがあればAPI/型部分先行とruntime接続待ちを明記し、fake成功を実装完了にしない。
fixture: 同薬同種複数pattern、ID別更新/削除、旧revision-null→明示保存、null/空/0、精度超過/overflow、他院/他薬剤、CAS、同keyreplay/異内容、unknown/auditrollback、最新headとsnapshot一致、旧承認を新内容へコピーしない、過去治療参照不変、旧species writerの迂回不可。
DB適用、実データ承認、実試験はこの技術契約に含まれずNOT_RUN。本route/DTO/anchor設計は親採択済み、未承認の臨床値を本人へ再質問する必要はない。選定とAPI収録/schema・runtime実装の完了は別である。

## 固定根拠

- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/docs/spec/domains/medical-records/lifecycle.md#dose-source-adoption-20261015
- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/backend/migrations/001_init.sql#L9125-L9163
- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/backend/internal/model/medicine_dose_param.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/backend/docs/api.yaml
- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/docs/work/emr/emr-490-492-selected-wire-contract-20261021.md
