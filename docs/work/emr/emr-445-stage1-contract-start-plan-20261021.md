# EMR-445 段階① DB・API 契約の着手プラン

2026-10-21 選定済みの着手プラン。基準は PR #802 の採用 commit `a88fa4add58d43cff643590f9f507c8ab1a3c085` と `docs/work/emr/emr-445-master-stages-ledger-20261021.md`。

段階① DB・API 契約 → ②全マスタ → ③共通UIを優先し、これらの完了前に④業務機能の新規着手を増やさない。既存④成果や進行中の担当は停止・再割当しない。

本書で選定した計画と移行DTOの境界は採択済み。本書に基づく新規実装は未開始で、必要な API・schema・model・generated の反映・受領は未完了。共通生成型の管理票は[EMR-527](https://app.plane.so/baritechllc/browse/EMR-527/)（親445、Ready、個人担当未割当）として作成済み。起票は実装・生成実行・試験の開始承認ではなく、技術方式の未選定でもない。正規 `make codegen` は本人実行の手順であり、agentは自動実行しない。本人の生成実行と成果受領は未実施。実装・生成・DB適用・試験は本書の作成では実施しておらず、試験は `NOT_RUN`。

<a id="requirement-validation-20261021"></a>
## 要件検証：既決仕様を実現する計画の採用境界

[AGENTS.md](../../../AGENTS.md)→[.claude/CLAUDE.md](../../../.claude/CLAUDE.md)→[product-philosophy](../../product-philosophy.md)の要件検証に従う。要件判断の窓口は本人 **Minoru Soga**。これは既存確定仕様への準拠と着手プラン整備を求めた本人指示（「改善して」「1,2,3は最優先に全てを完璧にしてから、4に進みたい」）の反映で、新しい業務要求・KPI・実装開始の承認を作るものではない。各実装の担当は別途単一claimで決まり、要件判断窓口を未割当の実装担当へ置換しない。②全マスタ・③検索/非検索操作の詳細計画にも本節を適用する。

| 設計思想の順序 | この計画で確認した目的・削減対象・境界 |
|---|---|
| ①要件を疑う | 目的は確定仕様と保存/API/画面の不整合、未説明の接続不足を具体化して、後段で同じ判断・転記・調査を繰り返す工程を減らすこと。業務正本docs/spec/domains/と採択済み共通契約を起点にし、現コードの不足に合わせて要件を縮小しない。臨床値・新業務判断は追加しない |
| ②不要工程・入力を除く | #775等の既採用CRUD・既存共通UIの再開発、同一共有hunkの二重実装、同じmaster値の別管理、keywordのない用途への検索欄新設、入力補完の一括置換を計画から除く。無差分は採用証拠で閉じ、空PRや試験だけの新規開発票を増やさない。患者情報の追加入力・未承認の自動連携は増やさない |
| ③残る工程を簡素化 | ①24管理行の正本/変更path/責務/受領条件、②26領域の保存/参照/画面、③72用途と非検索familyの既存接点を一つの文書PRと既存索引で辿れる形にする。新しい二重台帳・独自generator・別receipt方式を必須にしない |
| ④サイクル短縮の確認方法 | 計画指標は各対象行に正本・変更path（又は無修正理由）・完了/受領条件・担当責務があること、差分を未説明のまま残さないこと（目標は未説明差分0）。U/未受領/NOT_RUNは理由付き残件として可視化し、欠陥や完成へ換算しない。これらは計画coverageの指標で、診療時間/クリック数/売上等の本人未承認の業務KPIではない。業務実績のbaseline・改善量は未測定で、効果達成を宣言しない |
| ⑤自動化は最後 | 本稿は自動化の新規選定や実行をしない。生成型も既存正規generatorと入力保全の必要範囲だけで、実行許可は別。既存の現在認可・監査・unknown停止・同TX保護を維持し、安全な保護を空stubへ置換しない |

以上は計画の要件検証結果であり、実装接続・品質証明・業務効果測定の完了ではない。実装・生成・試験はNOT_RUN。本人の新しい業務判断が必要な変更が見つかった場合は該当項目だけを判断窓口へ返し、既決内容の再承認や全領域の一括停止へ広げない。

## 1. 計画の配置と所有境界

- EMR-445：API正本の単一統合窓口・下表の引渡し台帳。API専有hunk準備は並行可能だが api.yaml の最終統合は一人。新しい個人担当を本計画で指定しない。
- EMR-446/466：必要なDDL/model/ERDの共通編集窓口。各領域の必要subsetだけを同じ採用版に整合。全backend修復・実DB適用を要求しない。
- 領域原票：対象操作・wire・schema差分・依存契約の提出責務を保持。①の実装PRは関連票を束ねてよく、一票一PRを要求しない。
- 共通生成型：[EMR-527](https://app.plane.so/baritechllc/browse/EMR-527/)（親445、Ready、個人担当未割当）。責務 owner は「この票をclaimした単一実装session」。既存session3へ自動割当しない。
- 2026-10-21 10:33 UTC の確認時点で、#778のmanual edit / ClaimMetadataWriter / 短claimTX / migration005 / manual-corpus変更は現在の実装ownerが所有。新しい①担当が同じhunkを重複実装しない。後発確認では#778は2026-10-21 11:48:30 UTC、`eb137881c013f9bd7b0335c01c62ede31429cb23` でmain採用済み。ClaimMetadataWriterの限定採用と、生成同期EMR-527の本人実行・成果受領および個別consumer接続の未完を分ける。
- 試験実行・DB適用・新実装開始の許可は計画とは別。必要な実行許可を開始時条件として明示する。担当が未claimであることと、対象・完了条件が未定義であることを混同しない。

## 2. 追跡項目と完了の意味

後掲の24行の対照表は、S＝固定仕様と対象操作、W＝公開wire/DTO、D＝保存列・制約・model、X＝変更ファイルと担当境界、P＝必要共有interface、E＝受領証拠を記す。非対象は範囲表で明示する。該当しない差分は理由を残す。

計画完成は作業と境界が定義されていること。段階①の反映完了は API・schema・model・generated の必要差分と受領証拠が揃うことであり、文書の採択だけでは完了しない。実装PRの1つの対照表で追跡できればよく、別の hash 台帳・CSV・allowlist ファイルを必須成果物にしない。試験は段階⑤で、未試験を品質保証として扱わない。

## 3. 対象範囲：直接22票＋EMR-498 master＋297-A設定subset（24管理行）

各行の「次作業」は①限定。固定仕様・現状の値・受領証拠は第5節の対照表を使う。同一成果を複数票で二重実装しない。

| ID | ①対象・次作業 | 変更先/必要依存 | 非対象・行の閉じ方 |
|---|---|---|---|
|445|全行の差分と採用/受領を統合|本台帳、backend/docs/api.yaml の単一統合|runtime実装は各原票。全行Eへのリンクが必要|
|447|臨床6種の確定属性と既存CRUD/CAS/保険値の型対応を照合|api.yaml、454-456専有request/response/model、498契約|薬剤/検査の会計接続は④。未確認callerを適合済みにしない|
|448|CSV/copyのclosed wireとmetadata共用を登録|api.yaml、staff DTO、shift_bulk_operation_targets、446/466、493/489|parser/実保存/UIは②。copy用の第二metadata表を作らない|
|449|金銭masterと465内部guardの型・既存API対応表|採択payment契約、465、既存payment-method model/request/response|新公開route不要。459と単一実装|
|450|共通UIが受け取る既存loading/error/empty等のAPI状態を確認|UI台帳G1/G2、492/493 wire|新検索API/DB不要。UI接続は③|
|453|採用済み共通検索のAPI変更不要を記録|183d004f2採用、C01、450|再実装不要。③consumer差分へ引渡し|
|454|診察/処置・共有保険adapterの確定属性/CASのAPI-model整合表|既存emr454-master-cas-handoff、api.yaml、専有request/response/model|既採用CRUDは再作成しない。不適合hunkのみ①、会計consumer④|
|455|予防接種品目の保険属性のwire/NULL/省略整合|454共有契約、vaccine request/response/model、365|保険と価格を別差分。既存UATを全属性適合証拠にしない|
|456|健診種別の保険属性のwire/NULL/省略整合|454共有契約、checkup type request/response/model|測定値/基準値/自動会計生成を変更しない|
|457|コース属性/画像keyの既採用wire-model対応照合|trimming_master.go、専有request/response、731採用、458|画像cleanup/予約採用は別subset。既採用成果再実装なし|
|458|T2追加wire、設定state/digestと元anchorの保存型を収録|api.yaml、model/trimming_master.go、storage_reservations.go、trimming_configuration_operation_targets、489|設定CRUD・全writerlock参加②、480/482採用④。既存T2型を再作成しない|
|492|区分/groupの追加操作・名称版・C06 wireとanchor整合|api.yaml、reservation_type/group model、reservation_master_operation_anchors、489|490患者登録は混ぜない。設定writer②、予約変更④|
|493|休業/eligibility/CSV/copyのwireと保存版整合|api.yaml、model/staff.go・account.go、staff_leave_operation_targets、eligibility_version/digest、staff_eligibility_operation_targets、shift_bulk共用、489/467必要interface|実保存・旧writer切替②、予約consumer④。休業とCSV/copyを別familyとして保持|
|293|全無効化入口の既存actor/監査interface確認|emr-293-deactivation-audit、366のchain|新API不要。実actor配管②、IP/UA再判断なし|
|366|trusted staff actorの必須型と既存認証エラー契約を対応付け|staff_handler.go→service.Delete→credential_audit.go|client actor新設なし。実配管②、system推測補完禁止|
|365|price文字列段階互換とnumeric(23,4)/ExactNumeric対応を収録|api.yaml、model/vaccine.go、vaccine_request.go、vaccine_response.go、DDL/ERD|保存処理・FE②、実source/pin受領は限定境界。精度再選定なし|
|369|保存済会計への非自動追従を既存仕様へ参照|D10既決。新API/schemaなし|465のgateにしない。実医院設定確認を①設計残にしない|
|459|465単一owner/成果参照|465行|新PR・新guardを二重作成しない|
|465|内部guardのtyped input/既存master APIと保存型の整合|payment契約、既存method/fund/event型、489/506/464必要interface|専用公開API不要。停止writer②、新受領consumer④|
|518|package13操作API/S0の既採用範囲と生成型受領を集約|761、775採用、521/523、生成型後続|522/524 runtime④。親全体Doneとはしない|
|521|BE master6操作/S0の採用commitを生成型入力へ引渡す|775 merge9a1d50f9、input ae65ffd4a、生成型後続|CRUD再実装なし。生成物受領Eが必要|
|523|専有client契約と共通生成型の参照差分を受領|frontend/src/types/generated/models.ts、生成型後続|画面機能追加なし。実BE結合は②、試験⑤|
|498 master|複数pattern/固定revision DTOと最小anchor保存型を登録|api.yaml、model/medicine_dose_param.go、既存storage_clinical.goのMedicineDosePatternRevision、medicine_dose_pattern_operation_anchors、489|新revision model重複作成なし。CRUD②、治療採用/計算④|
|297-A設定|既選定mapping版/confirm元targetとcontextを①へ受領|445/446/466、lab_device_mapping_heads・lab_device_mapping_revisions・confirm用lab_mapping_operation_targets、draft版/context・採用field定義snapshot|②設定writer参加は447/297。daily application/wait/resolve/persistは④、今回A〜Eのruntime範囲に追加しない|

保存制約の反映に全writer同時接続が必須なものは、①でDDL/model差分を成果化しても、writer未参加のまま稼働適用・有効化しない。API宣言の収録は実handler登録済みの意味ではない。readonlyのphantom成功stubを増やさず、未実装を明記する。

## 4. EMR-527 共通生成型同期の管理責務と実装境界

管理票：[EMR-527](https://app.plane.so/baritechllc/browse/EMR-527/)（親445、Ready、個人担当未割当）。題名：① 共通model生成型同期と既存caller型互換の整合
親：EMR-445。担当：この票をclaimした単一実装session。既存の実装ownerを無断で再割当しない。

目的：#729以降の採用modelとfrontend共通生成型を一致させ、#775のpackage生成型受領を閉じる。①の型境界整合だけを行い、②CRUD/UI・④業務機能を追加しない。

### A/B間の最終生成入力と受領

A（445/446/466）の必要model入力変更を採用した**同一固定commit**を基準に、本人実行 `make codegen` の最終結果をB（527）が受領する。Bの先行caller整理は可能だが、A採用前の生成結果で①全体完了にしない。既存11出力（8TS＋3corpus）と手書き入力保全の範囲は以下のまま、agentのcodegen自動実行はしない。A/Bの型整合を489のQ/ContextGate等のruntime受領済みの証拠にしない。未実装subsetには既存担当があり担当漏れとは数えず、必要参加の受領は別管理する。実行NOT_RUN。

### 根拠と重複確認

- PR775はmerged、9a1d50f9dcc7d275456f1214507ef0a55441ac98。採用input ae65ffd4aは原引渡し記録の短SHAであり、開始時に完全commitへ解決して記録する。
- PR775コメント https://github.com/MinoruSoga/AnimalEkarte/pull/775#issuecomment-6310971025 は型同期を独立後続へ分離、未完・後続別承認と明記。
- Plane502 comment39f523d7-142c-4f9a-9d18-c287fa0700bbは共通生成型と6consumerの分離、担当・票・codegen実行許可未確定を記録。
- PR778本文「残件・対象外」はmodels.tsを変更しないと明記。manual-corpus修正は既存ownerのcorpus.go/repository.goで、別担当が重複実装しない。
- 2026-10-21 10:33のPlane検索「生成型」「型同期」は0件、「codegen」はDoneのEMR-198だけ。既取得526票の件名照合も198のみ。題名検索では本文のみの別票を完全否定できないため、claim前に445/502/521の最新引渡しとopen PRを再照合する。当時のopen PR取得では778のみ。この検索記録は起票前の履歴であり、現在はEMR-527を管理票とする。198は別対象の完了票なので再開しない。

### 固定入力・正規生成（調査済み）

参照mainは a88fa4add58d43cff643590f9f507c8ab1a3c085。実装開始時はこの版を含む最新採用commitをPRに記載し、#778のmanual-corpus差分を受領したcommitを別記する。#778は上記merge commitで採用済みであり、採用待ちへ戻さない。EMR-527の生成実行/成果受領は引き続き独立後続で未完。入力hash専用ファイル・別CSV・別allowlistは必須にせず、1枚のPR対照表で固定commit、source path、変更前後、受領先を追跡できればよい。

正規コマンドは既存 `make codegen`（.claude/CLAUDE.mdのuser-run指定。本人が実行し、agentは成果を確認・取り込む）。Makefileから docker-compose.yml のcodegen serviceへ進み、backend/tygo.yamlを出力先置換して `go run github.com/gzuidhof/tygo@v0.2.21 generate --config /tmp/tygo-ci.yaml`、続いて `go run ./cmd/manual-corpus --frontend /frontend --backend /app`。Go imageはgolang:1.26-alpine。ここを無断の手書き生成や独自generatorへ置換しない。DB serviceを起動するターゲットではないが、実ターゲット/実行先を本人へ提示し、本人が生成環境で実行する。agentの自動実行へ委任する計画にしない。

既存出力11件（tygoの8 TS＋manual-corpusの3件）：frontend/src/types/generated/ 配下の models.ts、auth-responses.ts、trimming-responses.ts、pet-responses.ts、owner-responses.ts、medicalrecord-responses.ts、hospitalization-responses.ts、identitylink-responses.ts、および backend/internal/manualarticle/corpus.generated.json、backend/internal/support/search_dict.go、frontend/src/constants/manual-corpus.generated.ts。frontend/src/constants/manual-search-dict.jsonはmanual-corpus/main.goがreadManualSourceで読む手書き入力であり、生成物として置換・手編集しない。本同期では入力を保全し、codegen-checkに検査対象として含まれることを出力の証拠にしない。codegen-checkのdiff対象はこの全体と同じとは限らないため、生成が触れた11出力をPRで確認する。生成物手編集・失敗出力の採用禁止。今回の静的調査は生成成功の証明ではない。

### 具体consumerと採用する置換境界

以下は同じ固定mainの実ファイルを読取済み。「6件」は過去の観測であり、現在の完了数や上限には使わない。対象symbolのgenerated/models直接importは9productionファイルを全frontend/srcで静的照合し、下表に列挙した。廃止型/保存model構造への依存と、JSDocのみ2ファイルを分離した。構造依存の下流transformは別に明示する。

| 実path（frontend/src/以下） | 元型/利用 | ①で選ぶ置換先・根拠 | ②/④へ残すもの |
|---|---|---|---|
| features/accounting/api/transforms.ts:2,46,68,84 | Payment / PaymentSplit / BillingRefund | 同feature api/legacy-wire.ts（新規非永続移行DTO）へ LegacyPaymentResponse、LegacyPaymentSplitResponse、LegacyRefundResponseを定義しimport差替。wire列はbackend/internal/billing/accounting_response.goのpaymentResponse/paymentSplitResponse、refund_request.goのRefundResponse。number/string/null/省略とstaff名をそのまま対応させる | 新PaymentFund/FundConsumptionEventへの意味変換・残高計算をここで足さない。新会計契約consumer実装時にこのlegacy adapterを置換する |
| features/accounting/api/get-refunds.ts:5 | axios.get<BillingRefund[]> | 上記LegacyRefundResponse[]。GET /v1/accountings/{id}/refunds の既存wire専用 | 新返金event readerの④接続 |
| features/accounting/api/create-refund.ts:2 | axios.post<BillingRefund> | 上記LegacyRefundResponse。PaymentMethod enumは残るため維持 | 新返金C06/資金選択への変更は④ |
| features/master/api/reservation-type-occupations.ts:7 | ReservationTypeOccupationRaw | 同file private LegacyReservationTypeOccupationResponse。backend/internal/reservation/reservation_type_response.go:77–84 の {id,clinic_id,reservation_type_id,occupation_id,occupation?:{id,name},created_at} を厳密に写す。保存モデルを新設しない | ②492/493で廃止gate UI/handlerを置換。正本 master-reservation-type.md#requirement-fr-reservation-types-occupation-gate-retirement を明記し、旧API存在を仕様承認にしない |
| features/accounting/api/types.ts:10,17 | Omit<Billing>, Omit<BillingItem> | LegacyAccountingResponse/LegacyAccountingItemResponse を上記legacy-wireへ。accounting_response.goのaccountingResponseと既存BillingItemResponseのHTTP列を明示し、保存modelのrelations/ExactNumericを自動継承しない | 新billing revision/確定金額の④。petSpecies参照はHTTPsummaryに無いので既存欠落のままoptional扱い、架空補完禁止 |
| features/estimates/api/types.ts:5,8,12,20 | Estimate/EstimateItem/Pick | 同feature api/legacy-wire.ts の LegacyEstimateResponse/LegacyEstimateItemResponse/LegacyEstimateItemRequestへ。billing/estimate_response.go・estimate_request.goの移行HTTP列が基準。requestとresponseをPickで混同せず独立列挙 | 新確定仕様の版/原文/金銭計算への切替は④。既に採択した新API consumerにはlegacy DTOを持ち込まない |
| features/medical-records/api/save-estimate.ts:6 | axios response Estimate | estimatesの同じLegacyEstimateResponseをtype-only参照。TaxType enum維持 | 見積作成機能や自動会計への新接続なし |
| features/medical-records/components/MedicalRecordEstimate.tsx:29 | Estimate / EstimateItemの直接import、toTreatmentItemの引数・読取 | estimates/api/legacy-wire.tsの同じLegacyEstimateResponse/LegacyEstimateItemResponseをtype-only参照し、toTreatmentItemもresponse item型へ。save-estimate.ts・estimates/api/types.tsと同じ移行HTTP列を使用、保存modelを復活させない | この①変更で見積の既存計算・未設定既定値・保存意味を変更せず、新固定版/会計接続は④ |
| types/pet.ts:1,10 | Pick<model.Pet,...>によるwrite | 現createPetRequest/updatePetRequestに対応する明示HTTP inputへ。既存backend/internal/pet/pet_request.go:108–133,177–206 と採択490/491の旧公開field互換に従う。DB-only列・関係配列は入れず、PATCHのstatus除外、name_origin等tri-state維持 | common subject/revision writerは④490491。体重をstorage ExactNumeric型のままwireへ漏らさず、新登録のexact decode契約は既採択490側を優先 |
| features/accounting/types/index.ts:64,81付近 | JSDoc Payment/PaymentSplit | JSDoc参照だけ上記HTTP DTOへ変更。UI PaymentInfo/PaymentSplitInfoは既存型を維持 | これはruntime import破損件数に算入しない |
| features/medical-records/types/index.ts:5 | JSDoc BillingConfirmation | 生成保存型への誤リンクを除去し、同fileの現移行BillingConfirmationとD03/D10 doctor-review契約の境界を明記 | ローカルBillingConfirmationを新doctor_review_sets/recordsに見せかけない。④494/会計ownerが意味接続 |

付随caller：features/estimates/api/transforms.ts はLegacyEstimate型経由でそのまま型整合させ、計算/未設定既定値を改変しない。accounting/transforms.tsも同様。types/owner.tsは既にcreateOwnerRequest互換の明示DTO、lib/transforms/pet.tsは既にgenerated/pet-responses.PetResponseを使用するため、生成model削除を理由とする再置換は不要。Owner/Petの全画面をscopeへ広げない。

採用判断の根拠は model-alignment-20261020/README.md:24–32 の保存model廃止と非永続DTO分離。現HTTPは新業務正本ではなく移行境界だけである。共通生成models.tsに旧Payment等を復活させない、any/type assertionで意味差を隠さない、UIの既定値や計算を新正本と称さない。API正本に新wireがあるconsumerは新wireを優先し、旧互換DTOで上書きしない。新操作への機能切替はこのPRの完了条件に含めない。

具体境界判断：①で全旧backend callerを新財務modelへ移植することは不可。正規生成が#778受領後も旧billing packageコンパイル依存で停止する場合、失敗outputを採用せず、その実エラーのpathだけを既存domain ownerへ渡す。これを架空Payment保存型の復活や全backend実装へ拡大して解消しない。既存tygoのinclude_files方式で生成可能なDTOだけを扱う場合も、generator変更は共有ownerの単一hunkとし、公開意味を勝手に増やさない。

### 実行開始条件

AGENTS.mdが参照する.claude/CLAUDE.mdのuser-run規約に従い、本人へ正規make codegenの対象commit・設定・実行先・付随処理を提示し、本人が実行する。agentは自動実行せず、本人の実行結果と11出力・手書き入力保全を確認して取り込む。過去の「型生成後続分離」や実装claimを本人実行済みの証拠にしない。生成が未実行/失敗なら失敗出力を採用せず、そのpath/結果を受領待ちとして保持する。試験/type-check/build/DB適用・隔離環境は別でNOT_RUN。許可範囲外のDB/外部送信/試験がターゲットに含まれる場合は本人へ該当内容を報告し、その実行を先取りしない。通常generator compileと製品試験の証明を区別する。

### 完了条件

- 入力採用commit・上記source path・正規コマンド・11出力の変更有無と手書き辞書入力の保全を1つのPR対照表で確認できる。
- 上の具体consumer表を満たし、削除型の直接importと保存model由来のHTTP Pick/Omit依存が指定境界から除かれる。新API側を旧互換型へ戻さない。
- 正規生成の成功結果だけを採用し、生成物手編集がない。
- 必要最小caller修正を同じPRで回収。②/④機能を追加しない。
- 採用commit、生成コマンド/設定、生成前後のsymbol/field差分をPRに記録。別hash台帳は任意。
- 521/523/445へ同じ採用commitを引渡し、受領を確認する。
- 未実行試験はNOT_RUN。生成成功を製品compile/test/品質PASSとしない。

## 5. 固定仕様・現wire・保存型・依存・受領条件の24行対照表

本節の全pathは固定main a88fa4add58d43cff643590f9f507c8ab1a3c085 に対する参照。URLは https://github.com/MinoruSoga/AnimalEkarte/blob/a88fa4add58d43cff643590f9f507c8ab1a3c085/ に当該pathを連結する。相対routeの共通prefixは /api/v1。api.yaml の server は http://localhost:8080/api/v1、frontend axios の既定 baseURL は /api と caller の /v1/... を組み合わせるため、二重に /api/v1 を付け足さない。Sの採択とW/Dの実反映を区別する。Eは新しい実装成果をまだ作っていないため「未受領」であり、空欄でなく次の受領条件を示す。

### 5.1. 既存契約を再実装しない行

| 票 | S 正本・節 | W 現wire/DTO | D/X 現保存・変更対象 | PとEの現在値 |
|---|---|---|---|---|
|445|docs/work/emr/emr-445-master-stages-ledger-20261021.md 段階①|新API無し、下記5.2のapi.yaml単一統合|台帳のみ、共通編集を445へ集約|#778採用済みmanual差分を保持（上記merge commit）。E=5.2全family＋生成型の採用commitと未実装表示の対照表。現未受領|
|447|docs/work/emr454-master-cas-handoff.md、stage ledger臨床群|下4票の既存CRUD＋498追加wire|新しい447独自schemaなし|454/455/456/457/498の同一仕様版を参照。E=下位成果リンク、親を二重実装しない|
|454|emr454-master-cas-handoff.md CAS/保険境界|medicalrecord/routes_masters.go: /masters/consultations と /procedures のGET/POST/PATCH/DELETE/reorder。create/updateConsultationRequest・ProcedureRequest。PATCH version正int64必須、boolCreateField/boolUpdateField|model/consultation.go・procedure.goにversion int64、is_non_insurance *boolあり。各request/response、004_master_insurance_versions.sql採用範囲。version/保険はjson:"-"のmodel生成に頼らずHTTP DTOから読む|追加DB/API設計不要。E=現APIの既採用CAS/保険と専有responseの型対応をAPI対照表へ記録。②既存writerの再実装を要求しない|
|455|上記vaccine consumer節、365小数契約|/masters/vaccines CRUD。create/updateVaccineRequest。PATCH version必須、is_non_insurance省略維持/null拒否|model/vaccine.go version int64/*bool既存。priceは*int64で365差分あり。vaccine_request/response.go|保険/CASは既採用、価格だけ5.2へ。E=保険部分nochange根拠＋365受領。price以外のfieldを勝手に厳格化しない|
|456|上記checkup consumer節|/masters/checkup-types CRUD。create/updateCheckupTypeRequest、checkupTypeInsuranceField、PATCH Version *int64必須（min1）|model/checkup_type.go版/*bool既存、checkup_type_request/response.go|保険/CAS追加設計不要、検査結果/健診fieldは本票対象外。E=現在のHTTP版/null契約対応の記録|
|457|stage ledger、既採択コース契約|trimming/routes.go: /masters/trimming-courses、options、course-types CRUD/reorder、course image upload/discard。create/updateTrimmingCourseRequest|model/trimming_master.go・trimming_course_request/response.go。ShortName/ImageObjectKey/AllowedWeekdays既存。T2拡張は458が所有|画像/通常属性を再作成しない。E=現属性wireのnochange記録＋458T2差分受領。新予約snapshotは480別|
|449|docs/work/emr/emr-449-459-465-payment-master-contract-20261021.md|465に集約、新449APIなし|新449表なし|E=465の単一成果参照|
|459|同上|465へ同じ操作/ownerで集約|新459表/guard無し|E=465成果、二重PR不要|
|465|同上 停止判定/同TX/再送節|既存billing/routes.go payment-methods CRUDとpayment_method_master_request/response.go。新公開route無し|model/payment_method_master.go、storage_finance.go PaymentFund/FundConsumptionEvent/FundRecordCorrectionを再利用|内部guard入力は選定契約の受領用途/元参照/現認可、停止writerと新受領は同master行lock。①=型/現API対照のみ。②停止writer、④受領consumerを分離。E=新API不要の理由と採択guard signatureへの参照|
|293|docs/work/emr/emr-293-deactivation-audit.md|既存staff変更/削除入口の認証actor。新公開request actor無し|staff_handler.go→service.Delete→credential_audit.go、監査保存再利用|366と同じchain。E=trusted actor必須型と全該当入口一覧の既存計画参照、実配管は②|
|366|上記＋Plane366の採択actor-only優先追補|新wire/新grant無し。認証actor欠落は既存認証エラーで変更前拒否|既存credential auditのみ、IP/UA列追加なし|E=actor AがBを変更、body actor不採用、欠落時副作用0、audit rollback fixture定義。実行NOT_RUN|
|369|docs/spec/domains/accounting/lifecycle.md マスタ再適用/保存snapshot|新route無し。既存保存会計を現在税設定で更新しない|新DB差分無し|E=①非対象の明記。実医院値の確認を①設計の待ち条件へ戻さない|
|450|docs/work/emr/emr-450-master-ui-consumer-ledger-20261021.md G1/G2|新検索API不要。G2各consumerが採用する上記wireへ接続|新DB無し|①=API結果形の引渡し、③=loading/error/empty/失権/競合UI。E=G2の採用API ref、UI完成を①へ混ぜない|
|453|上記G1、共通検索採用183d004f2|既存C01検索で新API/schema無し|既採用共通component保持|E=main採用参照。再実装を依頼しない|
|518|stage ledger package群、#775採用|既存package13操作契約/S0|保存型既採用、generated未受領|E=521/523の同commit生成型受領。親全体runtimeは別|
|521|#775 merge9a1d50f9dcc7d275456f1214507ef0a55441ac98|medicalrecord/routes_masters.go /masters/treatment-packages GET一覧/詳細、POST、PATCH、DELETE、PATCH reorder。treatment_package_request/response.go|package model/S0採用済み、DDL再作成なし|E=共通生成型後続の採用commit。旧sessionへ自動再割当なし|
|523|同上、専有package FE contract|同6master操作client、generated/models.ts受領|DBなし|E=生成型更新で既採用package型が残る対照表。画面新機能/④展開は含めない|

366は実在確認済み emr-293-deactivation-audit.md と本人actor-only承認のPlane最新追補を引用する。未確認名の新しい仕様pathを作らない。

### 5.2. API/schema の必要差分を選定済みの行

以下の選定routeは現api.yaml未収録であることを10:32固定main監査で確認。記載は実装済みではない。X共通はbackend/docs/api.yaml、各ownerのrequest/response、model、migration/ERD。DDL適用はしない。

| 票 | S 固定採択doc（docs/work/emr/以下） | W追加familyとclosed型 | D/X差分とP必要subset | E受領条件 |
|---|---|---|---|---|
|448|emr-448-493-shift-csv-selected-contract-20261021.md、emr-448-493-shift-copy-contract-20261021.md|CSV template、POST /shifts/csv-imports/preview と /shifts/csv-imports/apply、GET /shifts/csv-operations/{operation_key} と同pathの/details・/reconciliation。copyはPOST /shifts/copies/preview と /shifts/copies/apply、GET /shifts/copy-operations/{operation_key} と同pathの/details・/reconciliation。PreviewResponse/Preconditions/Details、CopySelection/CopyProjection/CopyPreconditionsは各closed定義をそのままschema化|`shift_bulk_operation_targets` 1表を両family共有、staff.go既存ShiftEntry/Break、aggregate版・deferred exclusion差分。489 Q/ContextGate必要subsetと#778採用済みClaimMetadataWriterへのconsumer参加。hookの二重実装なし。CSVとcopyのowner rowlockはbinding後Applyで統一|全route/schema/認可/C06status/details shapeを対照。metadata既存化の二重作成なし、scope/版/null fixture。consumer実装や実CSV入力は②/⑤|
|493|emr-493-staff-leave-selected-contract-20261021.md、emr-493-reservation-eligibility-contract-20261021.md＋上記CSV/copy|leave-periods/leave-operations選定family、GET/PATCH masters/staffs/{id}/reservation-eligibility、GET reservation-eligibility-operations/{key}。PATCH={operation_key,expected_version,changes}、changes=nomination_paused/reservation_type_ids。旧capable-reservation-types PUTへ版/keyを捏造しない|staff_leave_periods/changes既存、既選定staff_leave_operation_targets（PK/FK/RLS/不変性/winner claim同TXは休業契約の原対象anchor節どおり）、assignment eligibility_version/digest追加、staff_eligibility_operation_targets。model/account.go・staff.go再利用。467予約保護の最小participantだけ。旧休業row/CASなしwriteは切替時の停止adapterが②必須|API整合＋selected DTO/modelの対応。unknown/未設定/明示空を区別し、結果別再送表に一致。旧API停止の記述を含めるがこのPRでactivateしない|
|458|emr-458-option-assignment-selected-contract-20261021.md §3,8|計10 method/path（prefix /api/v1を除く）：GET /masters/trimming-service-identities、POST /masters/trimming-service-identities、GET /masters/trimming-options/{id}/service-identity、POST /masters/trimming-options/{id}/service-identity、GET /masters/trimming-courses/{id}/included-services、POST /masters/trimming-courses/{id}/included-services、GET /masters/trimming-option-rule-sets、POST /masters/trimming-option-rule-sets、GET /masters/trimming-option-candidates、GET /masters/trimming-configuration-operations/{operation_key}?intent={採択4種}。scope_kind/course_id条件は採択§3のまま、lookup intent allowlistは§5のtrimming.service_identity.create / trimming.option_identity.replace / trimming.course_included.replace / trimming.option_rules.replaceに従う。4resource_kindのContent/success branchを固定|既存T2表/model再利用、course included_services_configured/digest、設定元anchor。CHECKでdigest IS NOT NULL明示。T2設定全writerのinterface/lock契約とValidateAndCaptureConfiguration、489 Q/ClaimMetadata/ContextGate必要subset|未設定/明示空、no_course refs省略、identitycreate versions={}、scope digest CASをschema/fixtureで対応。実writer閉包②、480/482採用④|
|492|emr-490-492-selected-wire-contract-20261021.md 492節|既存type/group CRUD/reorderのoperation_key/CAS envelope、restore POST2本、business-kind change/confirm、reservation-master-operations lookup。renameは既存PATCH name branch|model/reservation_type.go・reservation_type_group.go、name revision・reservation_master_operation_anchorsを既存保存型へ対応。予約保護interfaceと489必要subset|型/NULL/旧wire移行/失敗をapi.yamlへ収録。490登録wireはこのsubsetから除外。保護participant実接続②|
|365|emr-365-vaccine-price-decimal-contract-20261021.md price cutover節|現Vaccine.price number/double、create/update integer/int64から採択string段階互換。ID/version数値維持、旧undefined/null読取・旧負値読取保持|vaccines.price bigint→numeric(23,4)、model/vaccine.go *ExactNumeric、medicalrecord/vaccine_request.go,response.go、DDL/ERD。旧price以外PATCHで再送しない|price精度・NULL/0/signed旧read・新非負・無round・cutover順を固定。最小producer型adapterが必要なら②との同時受領条件でありモデルだけ変更して完了不可|
|498 master|emr-498-dose-pattern-master-contract-20261021.md §1–6|GET medicines/{id}/dose-patterns一覧/詳細/revisions、POST/PUT/DELETE、medicine-dose-pattern-operations lookup。decimal文字列DTOとnullable元根拠|medicine_dose_param.go既存、storage_clinical.go MedicineDosePatternRevision既存、medicine_dose_pattern_operation_anchors追加。489 Q/ClaimTarget/ResultResponse必要subset|保存後pattern.versionをrevisionへ採用、legacy同値PUT版維持、元target/replay、NULL/未設定と新明示を対応。実計算/臨床値承認は④へ分離|
|297-A設定|emr-297-device-mapping-bridge-contract-20261021.md §6/7/7a/7b|既存設定PATCH/configuration/specとdraft-confirmの版参加に必要なcontext・採用field定義snapshot。confirmは既選定DTO/C06を再利用|lab_device_mapping_heads、lab_device_mapping_revisions、confirm用lab_mapping_operation_targets、draft version/context。既存model/lab_device_item_master.go・lab_device_mapping_draft.goを対応、追加modelの物理ファイル名は未指定。ContextGate/ClaimMetadataWriter/VerifiedCommand必要subsetを再利用|445/446/466がDDL/model/ERD・型を受領し、447/297の②全設定writer参加へ引渡す。lab_import_mapping_applications・wait・resolve/persistは④。今回A〜Eに297-A runtimeを加えない|

5.2の①PRは選定契約からAPI/schema/modelへ反映する差分で、業務handlerの新しい成功実装を含めない。Xの既存domain request/responseファイルがそのままruntimeへ接続される場合は、互換adapterを保持し、新wire有効化と旧writer閉包を②に明示する。stubで実装済み化しない。

449のcampaign.discount_value/merchandise.tax_rateは②詳細計画§3C/C2のexact adapter差分を①API正本へ受領する。既存number/double DTOをmodelのExactNumericへ直代入する方式は適合済みにせず、採択483/405のexact文字列境界と旧client切替を当該2fieldだけで明記する。API YAML編集・生成は今回未実施。

### 5.3. 未受領と未設計を分離した最終判定

- 24行の①対象/非対象、固定参照、変更先、必要共有subset、②④境界、受領条件は上記で明示した。実際のAPI/schema/generated変更は未受領で、①完了ではない。
- 生成型票EMR-527は作成済み、実装claimは未実施。特定担当の選択を技術設計の未決として利用者へ返さず、単一claim owner方式とする。
- #778は上記merge commitで採用済み。採用hook/corpusを並列担当が作り直さず、生成同期はEMR-527の単一claimで本人実行→成果受領を管理する。回復用接続や全C06 runtimeの完成を意味しない。
- 第4節の非永続移行HTTP DTOは採択済みmodel-alignmentの境界に準拠し、技術選定として採択済み。ここに新規設計残はない。生成型票EMR-527の作成は完了し、実装・生成実行・試験の開始承認は別途であり、本書の採択を実行許可とみなさない。本書では臨床業務値を追加選定しない。

### 5.4. 操作回復Slice Aの採択技術追補（API正本未収録・runtime未実装）

[回復契約§12の採択技術追補](emr-474-487-operation-recovery-contract-20261021.md#recovery-query-gap-selected-20261021)で、接種単位の本人reprint discovery、completeとcursor終端の同値、base64url/padding/strict typed decodeと2048文字上限の3点を具体化した。#803の既存計画と2queryのwireを保持し、§12が§4の走査終端/cursor詳細を補足する。監査追補とは節を分ける。対象は474/487と①445の限定引渡しであり、24管理行の範囲や業務ownerを増やさない。

| ①の確認列 | 今回の状態 | 次の受領条件 |
|---|---|---|
| 選定文書 | §12の3点技術追補は採択済み | API収録/実接続の受領を別に追跡 |
| API正本収録 | 新read/追加cursor説明は未収録。既存Slice A query2本は採用済み | 445単一ownerが新readと既存query説明の最小OpenAPI差分を反映・受領 |
| schema/model | 本追補はcontrol/DB/model追加なし、runtime適合完了とはしない | 既存binding/元targetから未成功を含む列挙と現在認可を実接続する際に照合 |
| 後続owner受渡し | 文書契約の採択まで。生成同期527とは別残件 | API固定後に474/487の純read adapterへ渡し、Slice B・新キー再開を開始しない |

試験はNOT_RUN、API YAML・生成物・runtime・設定・公開flagは未変更。計画・採択・収録・実装・⑤品質証明の完了を混同しない。

### 5.5. ①既採用範囲の限定静的監査（監査追補）

固定main `c86c8dd2530bbea3d1810e07badb72f3ed1a07e1` に対し、463、521/523、519/520、486読取、502/489のroutes・DI・主要境界を確認した範囲では、新規の確定抜けを見つけていない。全属性・全C06 callerの適合、実行成功、①完成の宣言ではない。予約ID producerの文字列化後続、certificate font/store/writer等の未接続、EMR-527本人生成/受領、522/524の④applyは意図した残件として保持し、新たな担当漏れに数えない。今回の回復契約3点は別の採択技術追補として§5.4に分離する。527–530のReady訂正は管理状態の是正であり、実装開始許可やclosesの根拠ではない。runtime・API YAML・生成物・試験は未変更/NOT_RUN。

<a id="stage-current-five-20261021"></a>
### 5.6. 今回5枠と後続subsetの引渡し

この追補は `06213d5778807f9ecbc9a8501dd579e8190a0d3b` 基準の計画接続補正。要件窓口はMinoru Soga、既決仕様を実現する受領条件の具体化で、新業務要求・新票・個人担当の割当ではない。既存5枠の今回実装scopeを下表に固定し、親全体完了や全票同時着手可能とはしない。A〜Eの実装sessionはそれぞれのclaim範囲で進行する。本書冒頭・旧節の「未開始/未claim/個人担当未割当」は計画採択当時の記録であり、現在の各sessionのclaimや実績を上書きしない。進捗・採用commitは各担当の証拠で別受領し、このdocs追補で実装完了を宣言しない。

| 枠 | 今回の責務 | 明示する受領・除外 |
|---|---|---|
| A | 445/446/466の共有API/DDL/model | 493のstaff_leave_operation_targetsを原対象anchor節のPK/FK/RLS/不変性/winner claim同TXまでDDL/model/ERD受領チェックへ追加。297-A設定の既選定保存/contextも①引渡しへ追加するが297-A runtimeや②医院FEを暗黙追加しない |
| B | 527 generated/direct caller | Aの必要model入力採用済み同一固定commit→本人make codegen→B最終成果受領。先行caller整理と最終生成受領を分け、11出力/手書き入力保全。agent実行禁止・NOT_RUN |
| C | 447/529 | 診断・主訴・臨床5tabのdescription clear、medicine base exact。498 pattern/revision・488製剤量は別 |
| D | 448/528 | hospitalizationのnull/省略/0/type、cage/occupation/staff公開4項目/trimming clear。492設定writerを今回へ追加しない |
| E | 449/530 | campaign/merch exactとinsurance description/contact_phone clear。医院税率・薬剤は530へ混ぜない |

②後続は458-A T2、492設定writer、493休業/eligibility/CSV/copy、366 actor、365 vaccine小数producer、498 pattern/revision、297-A mapping、465-M guard。②医院設定専有subsetは445の単一claim待ちで、Aの共有API/DDL担当へFE実装を暗黙追加しない。③450/523は次段階。これらは既選定の後続であり、未割当を未設計と読み替えず、A〜Eへ全面実装を追加しない。①24管理行、②26resource、③検索72用途と非検索familyは別集計。試験は⑤、今回NOT_RUN、製品コード/API YAML/DDL/model/生成物は未変更。
