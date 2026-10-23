# EMR-507 / EMR-508 固定原本・提示返答の採択技術契約

更新: 2026-10-21 03:36 UTC（独立review採用補正反映）
状態: 採択技術契約。runtime未実装・DB未適用・試験NOT_RUN。
適用範囲: 507 invoice初回/訂正発行、固定edition reader、同版再印刷、artifact段階。508 提示・特定提示版への初回承認/却下。旧設計の候補/未決記述より本契約を優先するが、既存業務正本を拡張しない。

## 1. 決定根拠と範囲

ユーザーの2026-10-21 03:30 UTC回答 Sentinel_4fae2c73bdd48191836bf897944922ec「1」により、新しい請求書固定版を発行する職員は「会計の編集権限がある職員のみ」。保存済み帳票の閲覧/再印刷は会計閲覧権限を維持。初回発行と訂正済sourceからの新版発行に適用する。権限rule/roleの付与変更自体は本作業で行わない。

読取基準main: 19855942f762f31f0403b5de0ddbc61b761fd94e。

- [会計正本](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/spec/domains/accounting/lifecycle.md): 帳票分離、元番号/旧版保持、飼主宛名/pet別/共通料金、請求日は原確定日。
- [見積正本](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/spec/domains/estimates/business-rules.md): 提示原本保護・版別返答・後継・会計分離。
- [財務操作設計](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/design/domains/finance-operations/technical-spec.md#finance-selected-operation-contract): Content/CAS・現在認可・receipt・TX意味。
- [C06参加表](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/design/shared/operation-participation.md)およびbackend/internal/operation/{contract,dto,runner}.go。
- [502 claim hook採択](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/work/emr/emr-502-manual-edit-contract-20261021.md): 原target anchor・claim原子性・immutable command。
- [現行API](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/backend/docs/api.yaml) blob 9459cb23c968975417f4a44ea6f8d31ae701176f を全文読取。新設routeは未実装。既存accounting/estimate CRUD DTOを全面置換しない。ただし見積status直接遷移は§9.3の移行境界に従う。
- [EMR-507](https://app.plane.so/baritechllc/browse/EMR-507/)本文・全2コメント、[EMR-508](https://app.plane.so/baritechllc/browse/EMR-508/)本文・全1コメントを照合済み。507 PR-A=固定edition、PR-B=artifact分離を維持。

## 2. Route・権限・intent

以下はOpenAPI相対path。実配信prefixは既存 /api/v1 を用い、別prefixを新設しない。

| Method/path | intent | 現在権限 |
|---|---|---|
| POST /accountings/{billing_id}/document-editions | financial_document.issue_invoice | accounting:edit + 対象view |
| POST /financial-documents/{document_id}/editions/{edition}/corrections | financial_document.issue_corrected_invoice | accounting:edit + 対象view |
| GET /accountings/{billing_id}/document-editions | invoice edition発見（§9.6） | accounting:view |
| GET /financial-documents/{document_id}/editions/{edition} | 固定reader | accounting:view |
| POST /financial-documents/{document_id}/editions/{edition}/reprints | financial_document.reprint_invoice | accounting:view |
| POST /financial-documents/{document_id}/editions/{edition}/artifact | 同版materialization/再開、C06業務Apply外 | accounting:view |
| GET /financial-documents/{document_id}/editions/{edition}/artifact | 同版artifact取得 | accounting:view |
| GET /accountings/{billing_id}/document-operations/{operation_key}?intent=... | 指定済familyのlookup | accounting:view |
| POST /estimates/{estimate_id}/presentations | estimate.present | estimates:edit + 対象view |
| POST /estimates/{estimate_id}/presentations/{presentation_id}/replies | estimate.reply | estimates:edit + 対象view |
| GET /estimates/{estimate_id}/presentations/{presentation_id} | 固定提示reader | estimates:view |
| GET /estimates/{estimate_id}/operations/{operation_key}?intent=... | 指定済familyのlookup | estimates:view |

lookup intent queryは上表の当該rootの既知family enumのみ。owner_kind/actor/clinicをqueryから選べない。owner_kind=billingは今回507/508共通のwrite-ownerとして選定する。共通ownerでもintent・原target anchor・現在grantは操作別であり、invoice権限をestimateへ流用しない。actorは認証staff、clinicは選択認証context。query enum→固定server intentへ変換し、既存bindingの原anchorを先に認可する。

pathの{edition}は文書内版番号であり、edition_id（financial_document_editions.idの行PK）ではない。trusted clinic + document_id + editionで一意行を解決し、その行PKをimmutable Contentとresult_refsのedition_idに使用する。committed_versions.editionは文書内版番号。path版番号を行PKとして検索/保存しない。

初回/訂正発行POSTは同key成功replayでも入口accounting:editを要求する。editを失ったview-only職員がPOSTを使って成功結果を取得する抜け道は作らない。view-onlyはGET lookup/readで保存済み結果を取得する。同様に提示/返答POSTもedit入口を維持する。view失効はlookup/replay/busy/unknown/key-conflictも非開示。

再印刷は同editionの監査/receipt記録だけ。新edition発行や元会計訂正ではない。artifactは同じ原本の形式生成に限定し宛先・金額・layout変更/共有範囲変更/外部送信を含めない。実プリンタ成功は主張しない。割引変更を提示/返答commandへ混ぜない。

## 3. Wire共通

- operation_key: canonical lowercase hyphenated non-nil UUID。既存ParseOperationKey使用。内部binding ID bigintとは別。
- 新専用DTOのID/版: 正のsigned bigint範囲のcanonical decimal string（先頭ゼロ/符号/空白なし）。既存numeric CRUD wireは維持。
- SHA256 digest: lowercase hex 64文字。DB byteaは32bytes。layout_versionは既存rendererが認識する固定version。client自由なrenderer/object_keyを受けない。
- actor/clinic/issued_at/presented_at/recorded_atはserver決定。bodyで指定不可。
- requestはadditionalProperties=false、重複JSON member・不正UTF-8/lone surrogate・decoded U+0000・非有限値を拒否。silent trim/NFC/改行変換なし。raw sanitizerに内容を変更される前に新対象route限定guardで拒否し、既存body上限は維持。
- ExactNumericは505選定の正規十進文字列/計算契約へ合わせる。binary float経由禁止。欠落/NULL/ゼロを分け、旧不明を架空値で補完しない。
- 全明細集合のCASは見積root版と確認済全行digestの両方で検査する。存在未確認のper-line version列を必須に新設しない。明細追加/削除/順序/内容をdigestに含める。

### 3.1 immutable commandとhash

502と同じ単一factoryが、検証済DTOからprivate immutable command、C06 Content、Preconditions、participantを同時生成する。map/slice/pointerをdefensive copyしsetterを公開しない。Applyはwriter前に同commandからdigestを再計算しreq.Contentと完全一致を検証する。異なるcaptured payloadや再parse差込みは禁止。

本節のimmutable typed commandは[採択済みQの共通契約](../../design/shared/operation-participation.md#verified-command-selected-20261021)に従い、Applyがreq.VerifiedCommandだけから再構成する。Qはrequest-memoryのみでowner-private耐久保管とは別、CAS/fenceはPreconditions、Contentは既定の短いstring mapを維持する。489のQ必要subsetは未実装依存であり、別command/DTO/pointerのcaptureで迂回せず、既存Content-only consumerは変更しない。


Canonical hash = lowerhex SHA256(UTF8(JCS(projection)))。JCSの文字列/keysはRFC8785、数値意味のあるID/金額/数量/版は正規decimal stringとして投影し、float誤差を持ち込まない。object key順はJCS、集合はID順、表示行順はordinalとIDを値として保持する。表示上の行順変更を勝手に同内容へ潰さない。

- invoice source projection schema=financial-invoice-source-v1: billing_id、document_kind=invoice、calculation_id、calculation_revision、原取引番号、元owner/元pet・共通料金区分、原請求日、固定source_snapshot/display_snapshot、layout_version。snapshotは当該sourceの固定実値のみ。
- estimate presentation projection schema=estimate-presentation-v1: estimate_id、固定header_snapshot（§9.5）、固定recipient_snapshot、valid_until（日付YYYY-MM-DD又はnull）、全presentation lineのestimate_item_id/ordinal/source_values/unit_price/quantity、固定計算内容とrule_version。採番後calculation/presentation IDやserver時刻は含めない。既存calculationを参照する場合はそのID/revisionを別Contentとして固定。
- invoice command projection schema=financial-document-command-v1: intent、billing_id、original_document_id/null、original_edition_id/null、source_digest、layout_version。
- presentation command projection schema=estimate-present-command-v1: intent、estimate_id、content_digest、確認済calculation_id/revision（新設ならnullとそのprepared計算digest）。
- reply command projection schema=estimate-reply-command-v1: intent、estimate_id、presentation_id、confirmed_presentation_revision、confirmed_content_digest、response、response_note。
- reprint command projection schema=financial-document-reprint-v1: intent、billing_id、document_id、edition_id。印刷時刻はserver audit、内容は同edition。

C06 Contentはbilling_id又はestimate_id、対象document/edition/presentationの必要ID、command_digest、source/content_digest、確認した計算ID/revision等の小さなstringのみ。256-byte/fieldの既存上限を維持し、大snapshot/全行配列はdigest commitmentで束縛したtyped commandに保持する。全projectionにoperation_key/actor/clinic/server時刻/CASは含めない（認証scopeで別束縛）。確認した計算revision・提示revisionは業務Content、現在のexpected_*_versionはPreconditionsで区別する。

## 4. 507 invoice固定発行

### 4.1 初回request/保存
request: operation_key、source_calculation_id、source_calculation_revision、confirmed_source_digest、layout_version、expected_billing_version。
sourceは505の確認済固定計算。client snapshotを無検証で保存せず、source取得→prepared immutable command作成→digest照合を行う。初回invoice identityはdocument_kind=invoice/billing_id必須、fund/refund_eventはNULL。ownerは元sourceから固定し、元番号は§9.1のinvoice専用選定により生成・保存する。請求未確定なら拒否し未確定previewを原本にしない。

billing root lock下でidentity既存なしとsource状態/版、current_calculation_id一致（§9.4）、原請求日event対応（§9.2）を検査。financial_document_identities + editions + 必要監査 + operation_resultsを同TX保存。editionのcalculation_idは対応billing_id付き複合FKに整合。source_snapshot/display_snapshotはobject、layout/issued actor/time/operation_idを固定。
成功refs: billing_id, document_id, edition_id, calculation_id。versions: edition, source_calculation_revision, billing_version。

### 4.2 訂正発行request/保存
request: operation_key、billing_id、source_calculation_id/revision、confirmed_source_digest、layout_version、expected_billing_version、expected_latest_edition。pathは元document/editionを指す。
既存の専用会計訂正で成立済みの別sourceを要求し、§9.4の現在計算と訂正event対応を検証する。発行POST自身では会計/請求/入金を訂正しない。元document/owner/番号を保持し別editionを追加。過去版・過去artifactを上書きしない。同sourceで別layout/宛名だけを変える任意新版はこのslice外。

### 4.3 別keyの重複排除
全invoice発行writerはbilling FOR UPDATEを共有し、(clinic, invoice, billing)のidentityを検査。初回に既存identityがある、または同documentに同calculation_idのeditionが既存なら新editionを作らず409 dependency-changed。別keyをreplayedと偽装しない。既存原本は認可済readerで取得。新key binding/anchorは成功なしで残る。
identityが複数あれば曖昧として停止。番号はroot lock下で最新+1かつexpected_latest_edition一致。既存edition uniqueだけでsource重複保証と称さない。旧/バッチwriterも同gate参加を必須とし、未参加なら当該保存経路を利用開始しない。追加business uniqueは本契約では必要とせず既存root lockで直列化する。

### 4.4 reader/再印刷/artifact
readerは元owner/番号/source refs、固定snapshot、固定計算ref、layout/発行時刻を返す。現在master/氏名/税率で再構築しない。欠落は不明/表示不能を明示。請求日は§9.2の原確定eventから固定した日付、印刷日/予定日/診療日を代入しない。
reprint requestはoperation_key、billing_idのみ（document/editionはpath）。path版番号から解決したedition行PKをContentへ束縛し、同editionへの監査とreceiptのみを同TX保存。成功result_refsはbilling_id/document_id/edition_id、committed_versionsはedition。新keyの明示再印刷は別監査を残すが同edition。同keyは同履歴へ収束。表示だけのGETでは再印刷履歴を作らない。
artifactはDB固定版成立後の別段階。server rendererが不変editionだけから生成し、外部生成/保存はTX外。最後の短いTXで(clinic,edition_id)一意参照を保存。同hashは同artifactへ収束、異hashは上書き拒否。生成成功/DB参照保存失敗は同版/hashの照合で再開し、issue新keyを作らない。固定版成立≠artifact ready≠印刷/送信成功。PR-AとPR-Bを分離する。

## 5. 508 提示・初回返答

presentation request: operation_key、expected_estimate_version、confirmed_content_digest、confirmed_calculation_id/revision。
505がprepared計算を同TX新設する場合はcalculation ID/revisionを入力で捏造せず、別variant prepared_calculation_digestを必須にして前記固定Contentに束縛。両variant同時指定は禁止。明細/recipient/期限/計算実値は確認済sourceからimmutable commandへ固定し現在値に黙って置換しない。

estimate FOR UPDATEで通常編集/提示/返答/後継を直列化し、root版と全明細digest（追加/削除含む）、未提示編集可能状態・現在認可・医院/飼主/pet/親関係を検査。計算revision/presentation/全固定lines/監査/receiptを同TX保存。親カルテ確定だけで未提示編集を拒否する旧gateは適用しない。通常削除も[2026-10-23本人判断](../../spec/domains/estimates/business-rules.md#estimates-delete-parent-state-20261023)に従い親未確定条件だけを外す。見積自身の削除条件/履歴保全を維持し、診療終了/医師確認/会計確定を代替条件にしない。
成功refs: estimate_id,presentation_id,calculation_id。versions: estimate_version,presentation_revision,calculation_revision。
reader: 固定header_snapshot、§9.7のnullable reply、固定recipient、nullable valid_until、digest、line_countと順序/元line refs/source_values/ExactNumeric値、固定計算ref・当時値、presented_by/at。提示保存は外部送信成功ではない。

reply request: operation_key、response=approved|rejected、response_note=null又はstring、expected_estimate_version、confirmed_presentation_revision、confirmed_content_digest。response_noteはrequired nullableとし、null=記録なし、空stringは空stringのまま保存（trim/空→null変換なし）。
path estimate/presentationの元関係とContentを一致させ、同estimate lock下で「当該提示に未返答」を検査。reply/許可された状態投影/監査/receipt同TX。旧提示への返答はその版だけへ記録し後継版の状態/承認へ継承しない。
成功refs: estimate_id,presentation_id,reply_id。versions: estimate_version,presentation_revision。

既返答同keyはreplay。別keyは同値でも新replyを作らず拒否し、相反回答も旧原本を保持する。withdrawnはDB enumに存在するがUI/業務正本に操作許可なし。本sliceは初回approve/rejectに限定し、撤回/既返答訂正/承認復活を対象外とする。後継/後継編集/選択会計転記は別intent・key・receipt。提示/返答だけで請求・受領・在庫・臨床実施を生成しない。承認済み期限内同内容同数量の単価保証と税/保険/割引条件は別。

## 6. 成功なしbindingの元target anchor

502 claim hookを再利用する新owner metadata（DDL設計選定、未実装）:

- financial_document_operation_targets: clinic_id,binding_id,billing_id
- estimate_operation_targets: clinic_id,binding_id,estimate_id

全列正のsigned bigint、NOT NULL、CHECK >0。PK(clinic_id,binding_id)。即時非遅延FK(binding_id,clinic_id)→operation_bindings(id,clinic_id)、ON DELETE RESTRICT。
対象billing/estimate FKは置かずclaim中の業務行KEY SHARE先取りを避ける。対象unique/別clinic FK/生成ID/default/時刻/actor/版/本文/statusは不要。

INSERT/SELECTのみ。表local BEFORE UPDATE OR DELETE拒否trigger、ordinary invoker。既存apply_rls_policy/has_clinic_accessによるFOR ALL USING/WITH CHECKとENABLEを再利用。FORCE/bypass/SECURITY DEFINER/role追加なし。queryはtrusted selected clinic+binding限定、binding JOINでbilling owner・当該family・actor scopeを検査する。RLSは現在認可の代用ではない。

ClaimMetadataWriter.CreateClaimMetadataはcanonical req.Contentからroot IDを取得し、同短いclaim TXでbinding新規winnerだけanchor作成。INSERT DO NOTHING RETURNING/RowsAffectedでwinner判定。callback失敗はbinding+anchor両rollback。loserはwinnerを待ち元anchorを読みcallbackを再実行しない。業務行read/lock、nested TX、domain history/receipt、外部作用はhook内禁止。欠落anchorの推定backfill/補修/上書きなし。

順序: 認証preflight→短いclaim(binding+anchor)TX終了→必要な既存owner Gate→binding NOWAIT→原anchor/current authでreplay/unknown判定→savepoint→Apply current write auth→billing又はestimate root FOR UPDATE→関連source/lineを安定ID順→CAS/条件→domain/監査/成功receipt→commit。不要な予約graph gateは追加せず、既存共通gateが必要な経路はその既存順へ参加し逆順を作らない。
claim commit応答不明は停止し同key lookup。Apply rollbackでもanchorは残り成功なしlookup可。old binding欠落anchor、元target消失/曖昧/認可不能は非開示。

lookupはbinding→保存anchor→元target現在view→route root一致→結果開示。C06到達前のURL不一致は元target現在認可後のみ既存409 ErrorResponse {error:string}。同keyのcanonical Content fingerprint不一致はoperation-key-conflict。fingerprint一致後のApply対象整合/CAS以外の依存変化はdependency-changed、CASはversion-conflictへ写像し、いずれも既存409とする。元target現在認可前に詳細を開示せず、target_mismatch独自codeは追加しない。incoming target Bの権限で元Aを開示しない。presentation/document/editionの参照も原rootへ一致検証する。507のinvoice-only anchorを将来fund帳票へ流用しない。

## 7. 公開receipt・error adapter

新route専用にC06 ResultResponse互換domain schemaを選定。既存finance CRUD/ManualEdit/他consumer DTOを変更しない。

- 200成功: status=succeeded|replayed、operation_key、allowlist result_refs/committed_versions、recorded_at。任意内部mapを無検査で公開しない。
- GET200 not-succeeded: status/operation_keyのみ。
- 今回507/508の新consumerは既公開C06 HTTP mappingを採用する。succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。202は処理中であり成功ではない。not-succeeded200も保存成功とは扱わない。
- C06非成功はHTTP値を問わず{status,operation_key}のみ。error/code/result_refs/committed_versions/recorded_atや原本bodyを混ぜない。C06到達前の認証・構文・body validation等だけ既存{error:string}形と既存HTTPを維持する。401は既存middleware。
- C06到達前のURL不一致は元target現在認可後に既存409 ErrorResponse {error:string}。同keyのcanonical Content fingerprint不一致はoperation-key-conflict409。fingerprint一致後のApply対象整合/CAS以外の依存変化はdependency-changed409、CASはversion-conflict409。元target現在認可前に詳細を開示せず、target_mismatch独自codeは追加しない。同source別key発行もdependency-changed409、既返答別keyはC06 validation400。
- artifact生成/取得失敗はC06業務Apply外の既存エラー形{error:string}と503で返し、C06 statusや成功receiptを捏造しない。
- EffectSummaryはinternalで公開しない。成功receiptのrefから固定readerへ進み、原文/金額等はそこで現在認可して返す。
- not-succeeded/not-found/unknownを初回不実行の証明にせず同key照会/人の再確認。同key同内容CAS-only再確認と確認source内容変更を区別。unknown/busyを新keyで救済しない。

## 8. 共有所有・実装順・受入

507 PR-A: invoice edition writer/reader/表示adapter。PR-B: artifact。受領/預り/返金は後続別source slice。
508: 提示→初回返答/既存後継→別段階の選択転記。481全体完了を提示準備のgateにしない。
505=固定算定、489=claim hook/C06、481=転記typed source、462=共有Complete、464=預り、506=fund、446/466=DDL/ERD。共有routes/OpenAPI/生成modelは単一編集ownerへhunk提出。除去済status/可変合計/旧payments等を復活させない。

fixture計画（全NOT_RUN）:

- editなし発行POST拒否（成功replayも拒否）、view-only GET lookup/版read/同版再印刷可、view失効で全結果非開示。
- claim winner1 anchor/loser非上書き、callback rollback、成功なし/unknown/busy/key-conflictの原target解決、BでAを認可しない。
- same key同Content1結果、別Content拒否、CAS-only再確認、別key同source edition重複なし。
- 編集/提示、返答/後継の両順、全行追加削除の競合、旧版返答を新提示へ継承しない。
- required監査/receipt失敗全rollback、応答消失lookup、sticky unknown停止。
- master/税/氏名/主飼主変更後の原本不変、原請求日と印刷日分離。
- artifact生成失敗/hash不一致/参照保存失敗をedition成立と別結果にし同版再開。
- 請求10,000/実受領3,000、共通料金、旧header見積、単価保証と税保険割引の区別。

試験コード/fixture準備と実装PR準備を後段の試験実行待ちにしない。実行は本人指示後mainローカルでまとめて行い、隔離/使い捨てDBは使用しない。必須CI/保護ルールは無効化しない。本作業では外部書込・コード変更・DB適用・試験・PR操作・実金銭・外部送信なし。



## 9. 独立reviewに基づく採択技術補正

本節のinvoice採番・原請求日暦日方式は2026-10-21のdot側の技術選定であり、本人が番号書式そのものを指定・承認したという意味ではない。旧版の実装済み事実ではなく、税・割引の業務算式も変更しない。

### 9.1 invoice原取引番号

今回invoiceに限り、表示番号を `INV-{clinic_id}-{billing_id}`（両ID canonical decimal）に固定する技術選定とする。初回identity作成TXで original_transaction_number へ保存し、以後はその保存値だけを読む。文書内edition/operation keyを元番号にしない。別billingへの再請求は別番号、同billing訂正editionと再印刷は同番号。既存identityの番号を改名しない。旧番号を持つ移行対象は別mappingで保持し、この規則で上書きしない。

初回source projectionの番号はこの明示規則から生成し、既存identityなら保存番号を使う。Applyのbilling lock下で再一致検査する。新しい年度連番registry・欠番禁止・法的採番適合保証を追加せず、fund/refundの後続帳票へ流用しない。旧文の『元sourceに番号が存在する』を実装前提にせず、この生成・保存契約に置き換える。

### 9.2 原請求日の正本と暦日adapter

462の請求確定TXは、当該billingの原 `billing_change_events.intent=finalize` と、同clinic/billing_change_event_idの `finance_source_events` を一体保存する。後者のoccurred_atを業務上の確定成立時点としてserverが固定し、recorded_atやDB commit時刻の代用とはしない。成功receiptと別commitしない。source_ownerはbilling、source_event_keyは当該finalize操作に固定したUUIDとし、同操作で再生成しない。既存列の帰属・digest・金銭非移動の契約へ従い、別資金イベントを生成しない。

507は対応する一意の原finalizeイベントを参照し、clinic/billing/source帰属を検査する。日付adapterの採択技術方式は明示 `Asia/Tokyo`、invoice_dateはoccurred_atを当該zoneへ変換したYYYY-MM-DD。実行環境TZに依存せず、source/display snapshotへ closed `{finalization_event_id,finance_source_event_id,occurred_at,calendar_timezone,invoice_date}` を固定しsource digestへ全量拘束する。IDは正decimal string、occurred_atはRFC3339、calendar_timezoneはこのadapter版ではAsia/Tokyo、invoice_dateは日付string。

訂正editionでは原日付を維持し、新billing再請求は新billingのfinalizeを参照する。欠落/複数/帰属不整合/旧根拠不明は発行不可で、印刷日/予定日/診療日/記録時刻へfallbackしない。旧日付の修復操作や現在時刻によるbackfillは本契約外。暦日adapterは本節で選定し、462 writer接続が完了するまでは実装済みとしない。

### 9.3 旧CRUDからの提示・返答迂回を廃止

新提示/返答契約の有効化時、旧CRUD status入力でのsent/approved/rejectedへの遷移は禁止する。createは未提示draftのみ、updateはstatus省略又は未提示draftの同値維持のみ。提示済みをdraftへ戻すことも禁止。通常編集は未提示/編集可の既決gate下だけで行う。

専用presentation/reply writerだけが対応する保存原本と同TXで許可された状態projectionを更新する。旧UI/APIの両方を同sliceで移行し、旧status入力を新操作への暗黙変換に使わない。C06未到達の旧CRUD拒否は既存エラー形を維持し業務書込なし。未移行writerが残る間は新提示契約を全writer適合済みとして利用開始しない。

### 9.4 新editionのsource鮮度

初回/訂正の新edition作成ではbilling lock下で source_calculation_id = billings.current_calculation_id を必須とし、clinic/billing帰属とsource revisionも一致させる。取消済み/確定前は拒否。訂正版は同billingの成立済みcorrect eventから現在計算への対応を検証し、`correction_event_id`（正IDstring、初回null）と `correction_event_revision`（正版string、初回null）をsource projectionに加えdigestへ拘束する。該当correct eventのsource_calculation_idだけでbefore/afterの意味を推定せず、505の当該correct adapterの固定after計算参照が現在計算と一致することを条件とする。未接続なら訂正editionだけを有効化しない。

不一致はdependency-changed409、業務保存なし。確認済sourceの黙った最新置換は禁止。成功replay・保存版GET・同版再印刷には現在計算一致条件を適用せず、過去固定版の取得を保つ。

### 9.5 提示header固定

estimate_presentationsにimmutable header_snapshot JSON objectを追加するDDL依存を選定する。closed `{schema_version:1,title:string,comment:string}`を新writer必須で保存し、元タイトル/飼主向けコメントの確認値を変更せず保持する。空commentは空のまま。staff-only notesは含めない。header全量をpresentation projection/digest、同TX保存、固定readerへ追加する。recipient_snapshotへ異なる概念を混入しない。

既存presentationのheader欠落はunknownとして表示し、現在見積から再構成しない。既存行へ架空値をbackfillせず、旧行nullableと新writer必須の移行境界をDDL/model/readerで明示する。

### 9.6 billingからのedition発見

GET /accountings/{billing_id}/document-editions はaccounting:viewとtrusted clinic/元billing現在viewを要求し、actor-scoped operation bindingへ依存しない。invoice限定。responseはclosed `{billing_id,document_id,latest_edition,editions,next_before_edition}`。未発行はdocument_id/latest_edition=null、editions=[]。edition要素はclosed `{edition,edition_id,calculation_id,source_calculation_revision,issued_at}`、ID/版は正decimal string、issued_atはRFC3339。next_before_editionは次page境界の正版string又はnull。

queryはbefore_edition（任意正版string）、limit（任意整数、技術候補default20/max100）。edition降順keysetとし、最新参照と返却pageを同読取snapshotで整合させる。権限なしは既存非開示、identity曖昧なら任意選択しない。0件は未実行の証明ではなく、unknown操作の新key救済に使わない。GETで新edition/印刷監査を作らない。

### 9.7 固定提示readerのreply

提示GETへrequired nullable replyを追加する。未返答null、返答ありはclosed `{reply_id,response,response_note,recorded_by,recorded_at}`。reply_id/recorded_byは正IDstring、response=approved|rejected、response_noteはnullable string、recorded_atはRFC3339。trusted clinic/estimate/presentationの元関係一致かつ現在estimates:viewで取得する。既存estimate_repliesを使用し新表なし。

別提示の返答やaggregate estimate.statusで補完しない。複数/不整合な旧記録は任意の最新行や未返答nullへ潰さず整合エラーとする。replyの追加で固定提示content_digestを変化させず、返答は原提示を参照する別記録のまま保持する。
