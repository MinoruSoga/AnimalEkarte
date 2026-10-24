# EMR-364 / 375 / 376：履歴CSVと通常更新を分離する共通契約

<a id="compatibility-scope-20261024"></a>
## 2026-10-24 現在の適用範囲（364）

§3.1/3.5/7/8の旧profile意味・既定/pin・表別commit/prefix resume・互換回帰を恒久維持する義務だけを撤回する。履歴tax_rateと375/376の保全・更新意味は変更しない。現在19 CSVへの適合を推定せず、新profileの全source hash・同一artifact/run/医院/target・独立検証・両側receipt・fresh/band保護・原子性・UNKNOWN照合・実操作承認と正式HOLDを維持する。旧成果物は証拠保全し、退役会計CSV/列を戻さない。 詳細根拠は[全票監査の採用是正方針](backward-compatibility-audit-20261024.md#accepted-corrections)。以下の旧日付の記録はこの限定追補と区別する。製品置換・実行受入済みではない。

状態：root技術採択済みの選定契約。業務方針は既決を再利用。producer共同合意・実source承認・実装・DB適用・runtime受入は未完。NOT_RUN。文書公開と実装完了を混同しない。
確認版：main `e71ddf03462a0ee7a9864ad2ce6854b243591770`、2026-10-21。旧351監査のed04c0237・236列は比較inventoryであり、現writerの証明ではない。

## 1. 採用境界と責任

- 364はCSVの履歴tax_rate輸送だけ。税計算、ワクチン小数365、薬剤分類343、実bundle生成MIG41を変更しない。
- 375は通常更新による既存import列の脱落防止。原文の編集権限を新設せず、見積の提示済み原本を更新しない。
- 376は履歴のread/保持と現在のcreate/PATCH/訂正を分ける列裁定。インポート可能だから現在の新規操作にも同じ値を許す、という推論は禁止。
- 3票は本稿の同じmatrixを参照し、互いの票全体完了を着手条件にしない。364の税rate新profileは375の旧列保持に依存せず、375の非金銭保持は364のreceiptを待たず、376のreader保持も新profileを待たない。
- 正本で未採択の個別金銭値・臨床値・producerの変換は本稿で承認しない。未決subsetだけ停止し、それ以外のPRを止めない。

## 2. 現在コードの確認事実

1. `cutover_profiles.go`は明示選択の`legacy-21`（空選択の既定）と`medicines-22-candidate`。後者はrehearsalモードかつREHEARSAL_ONLY artifactだけ。表数やdigestからprofileを推定しない。
2. `cutover_contract.go`のbilling_itemsは現在10列でtax_rateなし。estimate_itemsにはtax_type/tax_rateと参照列がある。medicine_idはlegacy-21では空必須、medicines候補ではCSV親参照、consultation_idは空必須。全部を旧21表の制約で扱わない。
3. 見積DTOは現在tax_type/tax_rate/existing_item_idを受ける。`applyMatchedEstimateItemTaxDefaults`は同一見積の一意の旧行だけに照合し、省略税設定を保持、重複・未知・他見積IDを400。旧375の「税未bind」は失効。ただし`estimateItemsFromInput`はmaster参照を転記しないため、税保持だけで参照保持完了としない。
4. 問診upsertは現在chief_complaint/notes/chief_complaint_type_idのみを更新。parent行lock・FirstOrCreate・送信済みkeyだけUpdates・再読取が同TX。省略値は保持。旧finalized/draftガードは現コードの事実であり、D03の新しい通常カルテ方針へ再採用しない。
5. procedure PATCHはversion必須、明示field mapと既存TX/CASを使い、is_surgeryはDTOにない。is_non_insuranceの省略/明示false/nullは専用field型で区別される。is_surgeryを保険設定や麻酔値から合成しない。
6. COPYは明示列・`FORMAT csv, HEADER true, NULL ''`とForceNotNull列の既存契約。RLS対象はTEMP staging→INSERT経路を使用し、権限設定を解除しない。エラーはtable/column/row等に限定し、生のDB/COPYエラーや原値を出さない。

## 3. 364：履歴税率の選定wire

### 3.1 単一の新規candidate

追加実装する明示profile名を`billing-tax-21-candidate`に選定する。表構成はlegacy-21と同じ、billing_itemsだけ次の11列に変更する。既存legacy-21とmedicines-22-candidateの意味・既定・pinを変更しない。medicines追加との合成profileは本票の必須成果にしない。これにより365/343の未決を巻き込まず税率輸送を閉じる。

`id,clinic_id,billing_id,category,name,unit_price,quantity,tax_type,tax_rate,is_insurance_applicable,sort_order`

これは新candidateの実装選定であり正式cutoverの採用ではない。名称をmanifest.schemaVersionへ混用しない。現schemaVersionは`animalekarte-cutover-v1`を維持し、operator-selected profileとmanifestの表/列集合・provenanceを現在同様に束縛する。

税candidateも明示rehearsal＋REHEARSAL_ONLY以外は拒否する。正式有効化は同じartifactに対するproducer/consumer receiptが揃った後の別変更。既存のrehearsal許容を理由に10列/11列を混用せず、ヘッダ順も本candidateではexact順に固定する。

### 3.2 bytes / 型 / NULL

- 新candidate producer出力はUTF-8（BOMなし）、カンマ区切り、LF改行、必要時double quote、quote内のquoteは2重化。CSVセルをJSON encodeせず、文字列の前後trim・改行置換・Unicode正規化を行わない。既存profileのencoding受入を変更しない。
- セルtax_rateはASCIIの`0.05`/`0.08`/`0.10`のいずれか。required/non-null。空、引用空、`0.1`、`0.080`、符号/指数/周囲空白/NaN/Inf/未知率は新candidateで拒否。float往復で正規化して受理しない。
- storageは現在のbilling_items.tax_rateのexact numeric契約を利用。runtime public DTOへstorage modelを直出ししない。preflightとCOPY直前検証に同じlexical predicateを使い、target schemaの型/nullabilityチェックにも新列を含める。DDL scaleで丸めて収める経路は禁止。
- 他の10列の型、ID帯、FK、nullable、ForceNotNullは選択元profileと同じ。bare空をSQL NULLへ、ForceNotNull文字列の空を空文字へ扱う現在の差を維持。0/falseをmissingへ変換しない。

### 3.3 日付と既決写像

canonical精算日の日本暦日で、2014-04-01未満0.05、2014-04-01〜2019-09-30は0.08、2019-10-01以後0.10。精算日NULL/欠落は0.10をproducerが明示出力する。非NULLの不正/変換不能はreview/rejectであって既定適用ではない。exemptを率0に変更しない。新しい原請求日定義へ旧精算日写像を置換しない。

consumerには元精算日の全根拠が無いため、rateのlexical検証をproducerの日付写像検証済みと偽装しない。MIG44が匿名境界fixtureと写像版をreceiptへ記録する。consumerは元の請求総額/入金額/税snapshotをrateから再計算・上書きしない。

### 3.4 pinと共同receipt

新しいhash値はsourceが未生成なので本稿に捏造しない。実装担当が採用producer sourceファイル全体のSHA256を計算し、新candidate専用にcompileする。既存global pinを使い回したりheaderだけhashしない。stageMappingSha256、csvContractSha256、各CSVのbytes/hash/rowcount、sourceRunID/stageBuildID/医院band/target bindingを現在のprovenanceと同時検証する。

共同receipt必須項目：profile名、正本のcommit/path、ordered header、型/NULL規則、source全体SHA、mapping SHA、採用producer/consumer各commit、同一4院candidate/run参照、legacy互換範囲、rehearsal限定/正式可否、未実装/NOT_RUN。364がconsumer部分を一度作成しMIG44/37/41へ同じ参照を渡す。MIG側に新しい業務税率判断を求めない。

### 3.5 既存医院・apply

本profileはfresh-onlyとする。対象医院のimport対象ID bandは全21表で空であることを、最初のBegin/COPY前と既存atomic経路のTX内検査で確認する。医院自体やseedの存在は禁止しないが、移行対象bandに行がある医院は拒否する。既存import済み行のtax_rate自動backfill・削除・置換・prefix継続・現在10%行の日付による一括書換えは行わない。新candidate選択は空bandや再取込可能の証拠ではない。通常APIで移行行を作り直さない。

**全21表のdata row書込みをsingle TXへ閉じる方式を採択する。** 現`cutover_import_resume.go`は`CutoverProfileMedicinesCandidate`のexact比較でatomic経路へ分岐し、それ以外はlegacy表別commit/prefix resumeへ進む。enumと列specの追加だけでは本契約を満たさない。364-Aの必須実装subsetを次のcaller閉包とする。

1. `cutover_profiles.go`の明示selector、profile別spec・pin・provenance・rehearsal/REHEARSAL_ONLY gateにtax候補を追加する。候補か否かの共通predicateは既知のmedicine候補とtax候補だけを列挙し、未知profileを候補扱いして受理しない。
2. `PreflightCutoverTargetAllowingResumeWithProvenance`はtax候補にも`validateCutoverBandEmpty`を使う。`validateCutoverBandResumable`へ落とさない。prefixが実manifestと一致してもtax候補の再開根拠にしない。
3. `ApplyCutoverCommittingEachTable`→`applyCutoverCommittingEachTable`のtax分岐は、既存の同一acquired connection・session advisory lockを保ち、空band検査、当該profileのunique/FK/target検査後に`applyCutoverWithBegin`へ渡す。別connectionでBeginしない。全21表のCOPY・検証と成功判定は既存atomic経路のTX/lock契約に参加し、legacyの`importOneCutoverTable` loop/finalize別TXへ流さない。medicines専用name検査はtax候補へ誤適用せず、共通master検査は省略しない。pre-Begin probeだけで並行writerを排除できたと扱わない。
4. 実caller `backend/cmd/csv-import-stg-uat/main.go`の`pgxCutoverTarget.Preflight/Apply`、`--cutover-profile`選択/help、bundle/provenance/target検査、applyログ、audit statusとfailure stageのcandidate predicateをtax候補に対応させる。候補失敗は`ClassifyCutoverApplyFailure`へ、legacyは既存`ClassifyCutoverTableFailure`のまま。候補のatomic化前に『全行rollback』と表示してはならない。直接のformal apply入口にもcandidate activation拒否を残す。
5. legacy-21の表別commit・complete-prefix resumeとmedicines-22-candidateの既存契約は変更しない。tax候補に途中prefixがあれば自動削除も途中続行もせず拒否する。長時間TXの実運用耐性は未検証であり、タイムアウトを理由に無断で表別commitへ切り替えない。

preflight成功とapply成功は別。`ErrCutoverTransactionNotStarted`、確定rollback、`ErrCutoverCommitOutcomeUnknown`を区別する。commit結果不明は成功/rollback確定にせず、**同一profile・manifest/hash・run・医院band・targetに対するread-only verifyとaudit/receipt照合だけ**へ進む。照合が未確定ならapply/retry/restoreを実行しない。全件一致を確認しても新規importやreceiptの捏造で成功扱いせず、既存の照合結果として記録する。空bandでも旧unknownを無視して自動再試行しない。sequenceが進んだ可能性をdata row全件commitの証拠にしない。

現在の21表specには退役済み`billings.status/completed_at`等のlegacy列が残る。tax_rate追加だけで最新main DDLへ全21表apply可能になったとはしない。target schema mismatchはfail closedとし、退役物理列を復活させず、通常writerで不足を補わない。必要なtarget adapterの接続は364-Aの実apply可能subsetとして明示し、未実装ならcandidateの契約/検証PRと実apply接続完了を分けて報告する。

## 4. 375：通常更新での原値保持matrix

| 対象 | 採用する保存/更新契約 | 実装責務 |
|---|---|---|
| inquiries.owner_observations/history/allergy_info/current_medications | legacy欄を独立保持。通常chief/notes更新で未送信列を上書きしない。再診回答や飼主原文へ自動変換/複写しない | 375は送信field mapとreaderの保持確認。496/477の原文APIは変更しない |
| inquiries.staff_id | 記録時点の参照。現在の入力actorで置換しない。現在actorは編集履歴側 | 375保持、376actor受入profile |
| inquiries.chief_complaint/notes/chief_complaint_type_id | 既存DTOと明示nullable契約を再利用。省略保持、空文字はその欄の既存許可規則、type_idの明示null解除は既存規則 | 375。D03終了/reopenの差分は495に委譲し旧finalizedを増設しない |
| procedures.is_surgery | imported保存値は通常更新で保持。anesthesia/categoryから再計算しない。既存boolean既定と未供給を別証跡にする | 375は保持。新しい編集UI/APIを本稿で増やさない |
| procedure保険/価格/税/階層 | 454等の既存field契約/CASを維持。変更しない列を0/false/nullで潰さない | 375は回帰のみ、各master機能ownerが変更writer |
| estimate_items.tax_type/tax_rate | 現existing_item_idの同一見積・一意照合を再利用。省略は旧税値保持。明示税変更は見積自身の状態/権限/版と505/508の規則 | 375は既存処理を再実装せず回帰ケース |
| estimate_items.procedure_id/merchandise_item_id/medicine_id/consultation_id | 現全行置換時も既存行identityで照合できる行には保存済み参照を保持。現在masterから再選択/再分類しない。既存参照を別項目へ転用しない | 375の主修正。下記identity・変更検知条件を必須 |
| estimates/estimate_itemsの提示済み原本 | 直接変更しない。改訂は理由付き後継と原本参照。税/価格は元版、実施・請求・入金を自動生成しない | 508の既存専用契約を利用。375独自訂正routeなし |
| items省略 / 明示[] | 省略は全行保持、[]は既存の明示全行置換要求として扱い対象状態・版に従う。空を省略にしない | 375。許可されない原本への[]も拒否 |

参照保持の具体条件：`existing_item_id`は同一見積内一対一でのみ使う。照合対象のcategory/master identity相当が変わる要求を、同じ元参照のまま新種別へ移すな。現DTOにmaster再選択契約がないため、その行の意味変更が必要なときは508の選択済み明細契約が実装済みの経路へ渡し、未実装なら当該変更だけ明示拒否する。参照を黙ってNULLにして保存成功する代替は禁止。名称はsnapshotで、現master名との一致を保持条件にしない。新規行（ヒントなし）は旧行を位置/名称/同額で推測して接続しない。

現serviceのtax引継ぎはコピーの正規化だが、参照引継ぎを追加するときは旧行の取得・状態/版・対応検証・全行置換・合計保存を同じ所有TXに閉じる。旧行読取後の別TX更新は禁止。税だけの修正と扱って金額再計算を旧float実装へ新規固定しない。505内税subset未閉鎖の操作だけ止め、外税通常/非金銭保持を止めない。

## 5. 376：履歴受入と現在writerの裁定matrix

| 列群 | 履歴read/保持 | create/PATCH/訂正・停止条件 |
|---|---|---|
| record_no/estimate_no、既存ID | 保存原識別子を保持。現在採番へ書換えない | 別identityの後継は既存生成規則。履歴readerから新番号を作らない |
| entered_by/paid_by/created_by/staff_id/doctor_id | 同profileの実在参照・医院条件を検証。過去actorと現在入力actorを分離 | 廃職/非activeの歴史actorを現在の操作権限へ変換しない。未存在actorを現在操作者やdefault doctorで補完しない。144/145の本番assignment receiptを代用しない |
| staffs.is_active/staff_typeと省略assignment | CSV供給のactive値を保持。非CSVのdoctor defaultは移行時DB挙動で、臨床資格承認ではない | 現職種/医院/権限は権威対応表。現在権限の拡大/自動assignなし |
| appointment status/source、medical_record status | 旧表現の出典を履歴として保持。importで現在ワークフローを再生しない | D01/D03の現在状態遷移・495 adapterへ。旧finalizedだけで全domainロック/開放しない。通常終了を新たに自動実行しない |
| billing status/completed_at、estimate status | 元状態・日時・提示版・実入金を別保持 | D10/D11の専用訂正/後継のみ。importを実精算/通知/在庫出庫にしない |
| pets.status/deceased_at | 原記録を保持。NULLから生存/死亡や時刻を捏造しない | D04死亡/訂正と生前分専用経路。履歴閲覧のために復活させない |
| birth/date/valid_until/recorded_at/created_at等 | 暦日と瞬間を区別し元値を保持。省略timestampを実施時刻と表示しない | 376の別版profile/日時adapter候補の形式は後述。tax候補へ自動適用しない。currentAPIの変更対象値だけ現validator。無関係PATCHで古い形式を再送/正規化しない |
| owners電話/郵便/メール、pets旧体重等 | 有効に保存された旧原値のreaderと無関係項目更新は保持 | 新規/当該値の変更は現入力規則。歴史的原値を現規則に合わすため無断trim/補完/0化しない。現DB floor違反はreader保持方針で突破しない |
| appointments start=end | 同値という事実を保持・観測件数と失敗件数を分離 | 374の共同profileと現時間writerを使用。1分加算等の無断修正なし |
| 課税率/割引率/insurance_ratio | 保存済み原値と適用根拠を保持、旧schemaが受けたことを新許容値にしない | 新入力の率範囲とexact numericを既存ownerへ。insurance_ratioはimportの0..1検証を維持。新しいAPI許容拡張なし |
| 歴史負額 | 既存に記録された原値を保持し現在master負価格禁止を遡及適用しない | AE-MIG-NEG-1のbillings.total_amount/payment_splits.amount以外へ新しい負額受入を推定しない。billing_items.unit_price、payments.total_amount/discount_amount、負の見積を新profileへ受入可能とする判断は未確定subsetとして明示保留 |
| payments/splits.method | 既存profileのcash/credit_cardを保持 | DB enumに他方式があってもimport受入を広げない。未知旧方式をcashへ変換しない。追加方式はproducer実値/既決写像の根拠が出るまで当該subset保留 |
| master参照なし/非CSV列 | 当該profileの空必須と型/DB既定を区別 | 値不明を現在master由来や実施済みにしない。legacy21のmedicine_id空必須を22candidateへ誤用しない |

日時の技術選定は**376の将来の別版profile/当該日時column adapter候補だけ**へ限定する。`billing-tax-21-candidate`はtax_rateだけの列差分であり、他列の日時受入形式・意味・変換は選択元legacy契約のまま。この候補のreceiptに日時正規化の承認を含めない。

376で明示版を作る際の選定形式はdateが厳密YYYY-MM-DD（存在する暦日）、timestampがRFC3339でoffset必須・小数秒はPostgreSQL保持可能な6桁以下。producer側で元の暦日/瞬間を保持し、consumerは欠落timezoneをJST/UTCで補わない。変更する表/列・元表現からの写像・source全体hash・profile名/版・双方receiptをその別版で列挙するまで現profileの受入を変更しない。元timezone/日付意味が不明な行だけproducer reviewへ戻す。単なる表記変更でもsource hashを変更する。376の日時別版は364-A/375-A/B/376-Aの前提ではない。

空/null/既定の共通判定：CSV列なし、bare空、ForceNotNull空文字、0、false、SQL NULL、API省略は別。既定で作られた値を『原値あり』と表示しない。新たなprovenanceが必要でも本稿で架空の旧値を作らない。既存保存値の保護・表示は今始められ、source非供給列の由来は別のreceipt欄に残す。

## 6. エラー・原子性・既存医院

- readerは現在の医院/患者/記録の認可を毎回検査する。履歴actorがinactiveだからreaderを全拒否する変更も、過去に見えたから現権限を免除する変更もしない。
- 通常APIは既存400（不正入力/他見積ヒント）、403/404（認可/対象）、409（版/保護状態）へ接続。新たな成功fallback、無権限列の黙示変更、部分保存は不可。
- importerはsource identity/profile/header/digest/FK/ID band/同医院・親子照合/targetDDLのどの段階で失敗したかを非機密coordinateで記録。患者原文・金額実値・生SQLエラーをログへ複写しない。
- importは通常writerを呼ばず、通常writerはimport例外を利用しない。TX内の再読取/版検査/親子関係と既存監査の契約を守る。今回全writerへ新しい監査必須条件を一律新設しない。専用訂正で既に必須の監査失敗はrollback。
- 既存医院のrowは読み取りだけで修正しない。新profileは再実行/切替/移行済みrow修復の自動承認ではない。source由来不明な列へ最新masterを埋めるrepairは禁止。

## 7. PR単位と非循環handoff

1. **364-A consumer candidate**：新profile selector、11列、strict rate validator、target scalar規則、hashのprofile別束縛、§3.5のfresh-only/全21表atomic/CLI・resume・failure分類caller閉包、匿名fixture/test記述、docs。旧profile回帰。producer receipt未受領ならcandidate限定のままPRを閉じ、正式有効化しない。実bundle/実DB試験は別。
2. **375-A 非金銭保持**：inquiry未送信列、staff原参照、procedure.is_surgeryを通常更新で保持するfixture/test記述と必要最小修正。新編集UI/APIなし。375-B/364を待たない。
3. **375-B 見積既存行保持**：existing_item_idの同一見積一対一検査を再利用し、元参照をsame-TXで保存。現在税保持は再実装しない。提示原本不変、items省略/[]/不一致/種別変更拒否、505/508 adapterとの責任を明示。金額writerの未実装subsetをfeatureフラグ等で勝手に回避しない。
4. **376-A 履歴read/保持adapter**：本matrixの現在認可・原actor/原日時/状態/空の意味をreaderと無関係PATCHで保持。新しいimport値域を開かない。375-A/Bの保持機構を利用する箇所だけ具体依存にし、376全票を前提にしない。
5. **376-B 日時別版候補**：対象日時列と元意味の根拠が揃うsubsetに限定し、明示profile/column adapter・全source hash・双方receiptを別途固定。tax候補や現在profileへ暗黙適用しない。未知timezoneを埋めず、その行だけ保留。
6. **共同receipt/正式activation**：364-A後にproducer同一artifact署名・参照が揃った時だけMIG44/37/41/349へ返却。実取込は別運用。365や未知負額採用を暗黙に含めない。

## 8. 固定fixture（作成/実行を混同しない）

全件NOT_RUN。既存匿名matrix/fixtureの構造を再利用し、実患者rowを取り込まない。
- 税率：両日付境界、NULL日付、不正非NULL日付、exempt率維持、10列を新candidateで拒否、11列を旧profileで拒否、未知rate/0.1/空/float表記拒否、改変hash/旧pin拒否。
- apply境界：tax候補の全21表空band、途中prefix拒否、既存行不変更、後半表の失敗でdata row部分commitなし、実CLIからatomic経路到達、legacy prefix互換、unknown時read-only照合のみ、未知profile拒否。
- 原値保持：owner原文に改行/quote/空、未送信field、false/0/NULL、staff旧参照。通常notes更新で他列byte/NULL不変。
- 見積：同一行税省略→旧税率、他見積/重複/未知ID拒否、順番変更はidentityで対応、元参照保持、ヒントなし新行の旧参照推定なし、提示済み直接変更拒否、items省略と[]区別。unknown/internal failureで旧items削除だけ残らない。
- 履歴：inactive actorを現操作主体にしない、旧終了状態でのread、current権限取消、376別版のdate/timestamp境界（tax候補の日時形式は変更なし）、未知timezone/負額/未知支払方式は非捏造、既存医院ID衝突、commit結果不明の再取込禁止。

## 9. 残るもの（新質問を今送らない）

- producer source/pinは未来の実ファイルに束縛するため、現在未計算。正式共同receiptと実bundle受領はこの設計文書で代行できない。
- 365小数採用は別。具体臨床値は47/48/49の承認境界。これらを3票全体の設計待ちにしない。
- 新profileで広げる負金額の個別列、現cash/card以外の旧支払写像、元timezone不明だけ、根拠取得後に最小追加裁定。現DBが通る/過去import済みだけを採用根拠にしない。
- 既存コードが目標と異なる部分（旧finalized、見積float計算、参照落ち）は明記済み。文書選定はruntime完成ではない。

## 10. 根拠

固定mainから確認：backend/internal/csvimport/{cutover_contract.go,cutover_profiles.go,cutover_contract_validate.go,cutover_reference_graph.go,cutover_import.go,cutover_import_resume.go,cutover_import_copy.go,cutover_ddl_scalar.go}、medicalrecord/{inquiry_request.go,inquiry_service.go,inquiry_repository.go,procedure_request.go,procedure_service.go,procedure_repository.go}、billing/{estimate_request.go,estimate_service.go}、model/storage_finance.go、backend/cmd/csv-import-stg-uat/main.go。

既決正本：docs/spec/domains/accounting/lifecycle.md、medical-records/pages/medical-record-form.md、estimates/business-rules.md。旧調査：docs/work/emr-351-api-csv-audit/{AUDIT.md,column-matrix.csv}。旧提案の現在所在はdocs/work/emr/emr-342-medicines-contract-design/DESIGN.md §8/9（旧emr直下パスは404）。
Plane：EMR-364 comment c12cf950-1c2b-4513-ae5b-7997e47ae3f5/4e239b7a-3c88-4960-b38b-bd3b86fe8f5e、EMR-375/376の2026-10-21優先追補と2026-10-20コメント。本文/全コメント取得、外部更新なし。

固定版根拠リンク：
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/csvimport/cutover_profiles.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/billing/estimate_service.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/work/emr/emr-342-medicines-contract-design/DESIGN.md

本matrixは3票に帰属する差分列群の裁定であり、旧351の236列全writerの再監査PASSではない。旧inventoryの列を削除せず、実装PRで選択profileの現列集合と担当subsetを固定する。未割当の旧調査列を「他票待ち」へ送らず、376が同じ列matrixに現在検証/保持/別契約の区分を記録する。

追加固定根拠：
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/csvimport/cutover_import_resume.go#L121-L201
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/cmd/csv-import-stg-uat/main.go
