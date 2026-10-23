# EMR-547 旧会計・見積の根拠保全と履歴閲覧：設計方針・未受領契約

2026-10-23。[EMR-547](https://app.plane.so/baritechllc/browse/EMR-547/)（親349、Backlog、担当/期限未指定）の7ACに対応する製品側保存/閲覧契約の設計文書。**設計未完・実装未着手**。設計文書公開と未受領事項の明確化が今回の成果で、closed schema/物理DDL/APIの採択や通常会計移行の許可ではない。

調査根拠はmain `42a3890a22d645857ad8e27c3b2d7bbfc9c2e038`、docs作成baseは最新main `621d58a0834aa6984ee3c82401e3478efef45d2e`。差はClaude許可設定のcommitのみで、本作業では取り扱わない。old_db実コード・実行結果・実原本行は今回未確認。以下の原値名/意味は受領した保全条件であり、正確なsource table/columnのclosed mappingはMIG-41から未受領。

今回docsのみ。製品コード/API YAML/DDL/model/producer SQL/実データ/設定/Planeは変更せず、tests/build/codegen/DB/import/Docker/browserはNOT_RUN。実PHI/原値行/ローカル実path/原本ファイルをGit・票・通常ログ・チャットへ転載しない。merge・レビュー依頼は行わない。

## 1. 根拠・業務目的と単一owner

業務判断窓口はMinoru Soga。旧請求の意味不明状態と旧見積の割引原値を失わず、製品内で原値/変換/不明を区別して履歴参照する。確認のために通常請求を捏造したり二重台帳へ転記したりする工程を除く。計画指標は7ACと保存→read/UI・担当・未受領契約が対応すること、未説明差分0。実業務の時間/操作数改善は未測定、保存/画面受入PASSではない。

| 参照 | 継承する境界 |
|---|---|
| [221：証拠範囲](emr-221-import-contract-design/DESIGN.md#source-evidence-scope)・[契約変更案](emr-221-import-contract-design/DESIGN.md#proposed-import-contract-boundaries)・[正式判定/耐久receipt](emr-221-import-contract-design/DESIGN.md#formal-eligibility-and-durable-receipt) | 原本identity/digestを固定、完全性と独立適格性を分ける提案、別revision原本保持、耐久receipt。同案を実装済み/正式import承認へ昇格しない |
| [formal安全境界](../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#formal-import-safety-boundaries)・[再開gate](../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#formal-import-producer-resume-gate)・[作業責務/実行binding補足](../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#producer-work-ownership-and-runtime-binding) | 既存formal v1/profileのstrict検査、MIG-41がproducer作業項目を所有する既存境界、quarantine/PHI保護、fresh-only/band保護。未知版fallback/自己申告modeを追加しない |
| [会計正本](../../spec/domains/accounting/lifecycle.md)・[会計技術仕様](../../design/domains/accounting/technical-spec.md) | 通常請求確定・実受領/配賦・返金・原子receiptは正規根拠とownerに従い、証拠参照で代替しない |
| [見積正本](../../spec/domains/estimates/business-rules.md)・[507/508契約](emr-507-508-selected-wire-contract-20261021.md) | 提示/計算revisionの意味が定まった型を維持し、未知旧割引値を無定義JSONへ詰めない |

| 作業項目owner | 本件の必要subset・除外 |
|---|---|
| 547 | 製品側immutable evidence保存、限定importer接続要求、read-only一覧/詳細UI。既存票/実装sessionへ無断追加しない |
| 349 | 配置・bundle受領・preflight/importの唯一の作業項目owner。547が別import実行経路を作らず、349という票を実行主体に読み替えない |
| 221 | 版付きevidence-only受領契約、identity/digests/独立適格性/receiptとの整合 |
| 446/466 | 必要物理保存schema、clinic複合FK/unique/RLS/immutable/index/model/ERDを受領・統合。架空物理名を予約しない |
| MIG-41 | producer作業項目ownerとして候補生成・source mapping/原本bytes/pins/receipt引渡しを追跡。既存producer責務はold_db HAC-CSV-1/HAC-PIPELINE-1にあるが、新evidence-onlyへの対応は候補で、正確なproducer SHA/mapping/実行receipt/bindingは未受領。MIG-41という票自体が生成処理を実行する意味ではない |
| 361 | 同一runのN1画面/復旧実機受入。製品保存/read API/UIの実装ownerではない |
| 505/508 | 正式会計/見積writer/reader・未請求候補/集計からの分離、lineageと許可されたリンクの境界 |

既存runbookに記録されたold_db HAC-CSV-1/HAC-PIPELINE-1のproducer責務と、AnimalEkarteの現行csvimportの配置/preflight/import経路は既存範囲の事実として区別する。これらを新しいEMR-221案/547 evidence-only保存へ接続する対応は候補であり、正確なproducer SHA/source mapping/実行receipt/producer-consumer bindingは未受領。既存CLINIC_CSV_IMPORTのscript/consumer経路が実在することだけでは、新契約への接続済み・実行済みを証明しない。実際の操作者・対象環境・個別実行許可は同runで確認が必要で、今回未確認。作業項目ownerやdocs上の責務を実行済み証跡/操作者名の代わりにしない。

## 2. 保存方針：別immutable evidenceの役割

通常`billings`/`estimates`/計算revisionの空欄、claim_state、未定義JSONへ押し込まず、別のimmutable legacy finance evidenceとして耐久保存する。以下は**論理的な役割候補**で、物理table名/列型/JSON exact/制約DDLは未採択。

| 論理役割 | 保持する意味・不変条件 |
|---|---|
| batch | clinic/source、run/revision/stage build、source snapshot identity、原本/bundle/mapping digests、contract version、import identity、耐久成功receiptの結び付き。別医院/別snapshotのreceiptを流用しない |
| snapshot（原本証拠） | kind=billing又はestimate、stable source evidence identity、source table/key（判明分）、対象原値payloadと原bytes digest。元値を更新・上書きしない |
| derived projection revision（snapshotに従属） | 同じ原本証拠に対するmapping/transform/encoding版別の派生値・変換由来・意味の検証状態・保留理由をimmutableに追記。原本の複製や通常請求/見積の計算revisionではない |
| verified-target-link event | 同院owner/pet/record/billing/estimate間の関係まで検証済みのnullable link。期待link revision・前event参照・理由・actor/timeと追加/訂正/撤回をappend-onlyで保持。原snapshotを変更せず、最新リンクを無検証再計算しない |

446/466への受領はclinic付き複合FK、identityのunique、clinic RLS、immutable（通常UPDATE/DELETE拒否）、検索index、storage model/ERDとproducer/consumer同一版を含む。business tableへの検証なしlinkや、名前/日付/金額一致だけの新ID推定は禁止。link不明はnull＋理由、後続検証は元snapshot不変の追記event。通常appへ任意作成/更新/削除APIを出さない。実際の取込権限・link追記経路/監査/取消方式はclosed contractの受領残。

### 2.1 nullable target linkの検証条件

医院の複合FK/RLSだけでは同院の別患者への誤linkを防げない。[会計の保存帰属](../../spec/domains/accounting/lifecycle.md#主飼主変更時の会計帰属)と[患者譲渡時の保全](../../spec/domains/owners-pets/lifecycle.md#owner-transfer-draft-retention-20261020)、[billings](../../architecture/erd-detail/billings.md)／[estimates](../../architecture/erd-detail/estimates.md)／[medical_records](../../architecture/erd-detail/medical_records.md)の正規関連を根拠に、埋めた全targetの組合せを検証する。以下は必要な論理条件で、具体reader・履歴証明・lock/closed schemaは受領残。

| 値を持つnullable target | 検証する関係（全対象の同clinic・実在確認に追加） |
|---|---|
| ownerのみ | 検証済みsource対応が当該ownerの履歴帰属を証明すること。氏名/住所の類似から作らない |
| pet | 検証済みsourceが同petを指すこと。ownerも指定した場合は対象履歴に正当なowner/pet帰属を証明する。現在のpet ownerとの一致だけでは過去帰属の証明にも否定にもならない |
| record | source対応と当該recordの保存済みpet/owner関連を検査し、指定したpet/ownerとも履歴基準で整合すること。別患者のrecordへ付け替えない |
| billing | source対応と当該billingの保存済みowner/pet/medical_record関連を検査し、併記したowner/pet/recordがその保存帰属と整合すること。任意関連の欠損を推測補完しない |
| estimate | source対応と当該estimateの保存済みowner/medical_record等の正規関連を検査する。併記petは正規record等を通した証明が必要。billingも併記する場合は両者が同じ検証済み履歴関係に属する根拠を検査し、同院/同名だけでは結ばない |

targetがnullならその関係を確認済みとは扱わない。他の非null target同士の矛盾をnull欄で隠さない。無関係な同院患者や矛盾する組合せは拒否し、証明不足は当該linkを未リンク保留にする。部分的に証明できたlinkを採る場合も、採用する全組合せが整合することを要し、未証明targetをverifiedに含めない。譲渡前の正当なowner帰属を現在ownerへ自動移管せず、履歴上の関係証明とread時の現在認可を別検査にする。

追加/訂正/撤回は期待link revisionと現在の前eventを同TXで照合し、競合時は追記も有効link切替も行わない。原eventをUPDATE/DELETEせず、撤回も新eventで残す。期待版/競合error等の具体wireは未採択。認可付きowner/pet filterはこの検証を通った有効な関係だけへ適用し、未リンク保留や撤回済みlinkを他患者への検索入口に使わない。

全原本bundleは既存の保護領域へ保管する。同時に画面に必要な対象原値をDBへ耐久保存し、外部bundle/fileの喪失で詳細画面が空になる設計にしない。raw payload bytesを外部参照だけにせず、DBに保存する対象範囲・encoding/bytes digest・typed projectionとの対応を選定する必要がある。大容量の全bundleを無制限にDBへ複製する採択ではなく、サイズ/保持/backupの具体設計は未受領。ローカルquarantineに置くだけで製品閲覧完成としない。

## 3. 原値・意味検証・変換の分離

presence（value/null/empty/unavailable）、文字列としての元値、意味の検証状態を別軸で保持する。値が不明なことと未取得を混同せず、実際の0/空文字/NULLを互換のために潰さない。raw、derived、transform rule/版/根拠を分け、rawをbinary floatへ通したり不明を0補完したりしない。exact JSON/decimal文法/文字コード/サイズ上限はproducer-consumer共通schemaで未採択。

| 対象 | 保存/表示する原値と由来 | 禁止する意味補完 |
|---|---|---|
| 旧billing | 旧status、TSeisan_Flg、KKarte_Kbn（受領資料のKkarteKbn表記との対応はmapping待ち）、各日付、請求/現金/カード額を独立原値として保持。元の意味確認状態/保留理由を併記 | 意味不明TSeisan=1やshell生成定数1だけでcompleted認定しない。旧pending/waiting等を一律draft/finalizedへ変換しない |
| 旧精算証拠 | KKarte_Kbn=4は旧精算完了の証拠として区別し、旧精算日を原値として保持 | 旧精算日=新請求確定日とせず、不明な新確定日/終了event/入金配賦を生成しない。定数finalizedから終了eventを生成しない |
| 旧estimate割引 | header discountのraw値と丸め後値、変換規則/版/出典をそれぞれ保持 | 丸め後だけ残して原値を捨てず、割合/金額/適用順を推定しない。`estimate_calculation_revisions`の未定義JSONへ投入しない |
| 旧estimate保険 | insurance=0が生成定数ならgenerated_constant由来を保持して表示 | 原保険額が既知0円だったと解釈しない。生成定数/未取得/意味不明/確認済み値を分ける |
| 元PK不明 | 検証済みartifact locatorで原bytes中の場所と出典を辿れるようにする | 架空source keyを作らず原PK表示しない。artifact digest＋locatorを元レコードの一意性証明と呼ばない |

元table/column名・型、source status enumと各flag/date/discountの意味、PK判明範囲、bytesと派生値の対応はMIG-41のclosed mapping受領残。old_dbコードを今回検証済みとは書かず、推測でproducer SQLを修正しない。

## 4. evidence-only保全受入とformal importは別

| 判定 | 本件の扱い |
|---|---|
| clinic/provenance/bytes digestが検証済み、業務意味だけ不明 | evidence-only保全の対象にできる。通常台帳はHOLD。保存/閲覧できることをformal適格性/完全性PASSへ変換しない |
| provenance自体不明、bytes digest不一致、医院不明 | quarantine維持。製品保存受入で正当化しない |
| 通常台帳へのformal import | 既存v1/profileのstrict完全性/税/親子/FK/payment/clinic/target等の全条件と既存owner判断を維持。本書から許可しない |

evidence-onlyは**別名・版付きの明示契約**を221/MIG-41の作業項目ownerが実producer/consumer双方で整合させてから接続する。具体version文字列/manifest exact/parserは未採択。既存v1を緩めず、manifest自己申告や不明payloadの内容からmodeを推定しない。rehearsal pinや既存partial formal拒否をevidence-onlyの実装済み根拠にしない。

fresh-only/ID band/既存import prefix/lockの規約は対象profileの既存契約を維持し、本件保存で変更しない。通常台帳/証拠保存/除外集合/受入receiptの母集団を合算せず、source不足の事実を消さない。

## 5. 再送・訂正・原子保存と結果不明

stable source evidence identityはclinic/source system/source snapshot identity/source table/source keyへ束縛し、contract/mapping/transform/encoding版を混ぜない。同stable identityの同じ原本bytesは同一証拠へ解決し、原本bytes不一致はconflictで上書きしない。digestをidentityへ足して異内容を新規化したり、契約版の更新だけで原本snapshotを重複生成したりしない。原本自体の変化は別source snapshot/revisionの証拠として関連付け、旧原本を保持する。

同じ原本へのmapping/transform/encoding変更は別のimmutable derived projection revisionとして追記する。派生値の差を原本改変conflictと誤判定せず、同じ派生版/内容の再送は既存結果を参照する。派生版の識別・同版異内容conflict・具体番号/encoding/closed payloadは221/MIG-41との受領残。原本identity、派生revision、各受領batchの実行identity/成功receiptは別概念であり、再処理batchやreceiptが新しくても原本を増殖させない。

元keyが不明な場合は検証済みartifact locator（artifact identityとその中の位置）でsource keyの代わりに**証拠位置**を同定するが、元レコード一意性は保証できない。artifact bytes digestは整合検査に使い、不一致を別identityへ逃がさない。別artifact間の同一人物/取引や同内容の重複を推測統合しない。不明identityでどこまで受けられるか、並行取込のlock/unique・conflict応答はclosed consumer契約の受領条件。

設計目標はbatch/原本snapshot/初期派生projection revision/初期verified links/耐久成功receiptを同transactionで保存すること。既存原本を再処理するときは原本を参照し、必要な派生revisionと当該batchの結果だけを追記する。現在のimporterが実現済みとせず、必要adapter/監査sinkを221/349/547と保存ownerが受領する。成功はreceiptの耐久確認後、commit結果不明は同batch実行identityと保存済み原本/派生revisionの対応を照合する。unknown中の新identity/再importで救済せず、ローカルreportだけをDB保存成功証明にしない。

大batchを分割する場合のtransaction単位・部分成功receipt・再開/未知判定は未選定。全体原子性を架空保証せず、既存formal importerのatomicity/profileを変更しない。後続link追加/訂正/撤回は§2.1の期待link revision/前event照合・現在認可・actor/理由・監査を必要とし、具体結果参照を選定する必要があり、通常resourceの更新APIで代用しない。

## 6. read-only API/UIの候補（未採択）

候補は既存prefix規則へ合わせる相対path：`GET /accountings/legacy-evidence`と`GET /accountings/legacy-evidence/{id}`、`GET /estimates/legacy-evidence`と`GET /estimates/legacy-evidence/{id}`。これは実在/収録済みrouteではない。動的normal `/:id`との登録順、typed legacy identity、closed DTO/ID型/カーソル/limit/sort/enum/具体errorは未採択。

- 認証clinic固定。会計証拠は既存accounting:view、見積証拠は既存estimates:viewを維持し、新共通grantを作らない。[現routes](https://github.com/MinoruSoga/AnimalEkarte/blob/42a3890a22d645857ad8e27c3b2d7bbfc9c2e038/backend/internal/billing/routes.go#L92)の別資源として検査する。他kindの原値を同一権限で混在返却しない。
- owner/pet filterはverified linkだけへ適用し、raw source名/番号の一致から現在対象を推定しない。未リンク証拠は当該clinicの専用一覧に置き、他院/未認可resourceへ横断しない。link訂正後も履歴の元帰属は保持する。
- 一覧候補はsnapshot ID/kind、原日付とその意味、旧状態原文、保留理由、link状態、証拠版。詳細候補は対象原値/derived/変換由来/関連証拠/全link履歴。raw payload bytesやlocal pathを一括返却せず、許されたprojectionだけへ。ID/金額は文字列を候補とし、日付不明を補完しない。
- HTTP候補は200、400 invalid filter/cursor、403 grant不足、404 missing/他院、500非機密read error。未認証は既存認証処理、具体closed error/非開示順はAPI契約で固定する。driver error/原値/PHIをエラーへ含めない。reader障害を空成功/原値なしへ偽装しない。

UIには「移行履歴参照専用」「通常請求への反映保留」を明示する。旧精算完了の証拠と新請求確定日不明を別表示し、原値/変換値、生成定数/不明/未取得を区別する。通常編集/回収/再計算/帳票/再請求ボタンを付けない。serverもlegacy IDをnormal billing/estimate IDとして受けず、通常writerの型付き入口とread lineageで拒否する。UI非表示だけでは保証しない。

通常未請求候補/会計/売上/見積集計へ旧明細が自動混入しないことを505/508と確認する。verified linkがあっても通常取込保留を解除しない。製品側保存→認可付きread→一覧/詳細UI→同identityの受入証拠が必要で、quarantine保存だけではAC5/6/7完成としない。

## 7. EMR-547の7AC対応と合成fixture（全NOT_RUN）

以下の番号は547本文の最小受入条件1–7に対応し、既存票の実装/受入成果を再実装・完了認定しない。

| AC | 必要成果・期待結果 | 主な受領責務 |
|---|---|---|
| 1 旧billing状態/原値保全 | status/flag/date/各額のraw・意味確認状態・保留理由を保持。意味不明/生成定数から通常claim/finalizeを生成しない | MIG-41 mapping→221→547、505境界 |
| 2 estimate割引/保険由来 | raw/丸め値/規則版を別保持、generated_constantのinsurance=0を原本既知0と表示しない | MIG-41、547/508 |
| 3 identity/検証済みlink | clinic/source/snapshot/run/revision/build/各digestを束縛、保存帰属/全指定target間の関係も検証したlinkだけ採用、不明null＋理由。架空PKなし | 221/349、547/446/466 |
| 4 再送/改訂/訂正 | stable原本identity再送は二重保存なし、原本bytes不一致conflict、派生版変更は追記/別source revision旧保持、link期待版競合拒否・追記履歴 | 547、221、446/466 |
| 5 認可/一覧詳細 | 医院分離とaccounting/estimatesの別view grant、原値/変換/不明/由来/保留/関連証拠/履歴をread-only表示 | 547/API owner、既存認可 |
| 6 通常処理との分離 | 通常請求/回収/帳票/再計算/未請求候補へ混入なし、legacy IDの通常writer利用拒否、閲覧保存≠formal保留解除 | 547/505/508、349 |
| 7 合成受入/機密 | 欠損/null/空/実0/unavailable、他院、競合/重複/改訂/履歴継続、権限拒否を合成fixtureで定義。受領→保存→UIを同identityで照合、PHI転載なし | 547実装、361同run実機受入 |

fixtureを用意する条件と実行を分ける。非PHIの合成値で、bytes改変/provenance不明はquarantine、意味だけ不明は証拠候補/通常台帳HOLD、外部原本file欠損後もDB保存対象原値の表示、異snapshot/clinicのreceipt流用拒否、link撤回後の旧履歴保持、通常writer/集計非混入、receipt障害/unknown再送停止を記述する。追加fixtureは、契約版だけ変更して原本重複なし、同原本のmapping変更で派生revision追記、同identity原本bytes不一致conflict、同院別患者のowner/pet/record/billing/estimate組合せ拒否、譲渡前の正当帰属保持と現在認可、関係不明の未リンク保留、矛盾link拒否、同じ期待link revisionの並行訂正は片方のみ、撤回後filter非混入を含む。今回はfixtureファイル作成/実行も行わない。実データreceipt/producer適合/STG/本番受入とは別証拠。

## 8. 設計残・次の限定引渡し

| 未受領 | 閉じる対象 | 完了と誤認しないこと |
|---|---|---|
| MIG-41 closed source mapping | 旧table/column/PK、各flag/date/status/割引/保険定数のraw・derivedと変換版、bundleの同一snapshot証明 | 本文の保全原則だけでold_dbコード検証/producer生成済みにしない |
| 221 producer-consumer版付き契約 | evidence-only名/版・strict payload/bytes・identityと欠損identity・独立保全適格性/receipt・formal隔離 | accepted-exclusions提案やmanifest自己申告だけで有効化しない |
| 446/466物理保存 | batch/原本snapshot/従属派生projection revision/link履歴の実名/型/FK/unique/RLS/immutable/index/model/ERD、transaction/大batch方針 | 現normal tableのJSON欄へ暫定投入しない。DDL実装/適用未完 |
| 547 read API/UI | closed DTO/permission map・filter/cursor・error・projectionと履歴、namespace境界、保存後read/画面 | 本稿の候補routeを実在扱いしない。read-only受入未完 |
| 349/361同run受入 | 唯一import ownerの受領/配置/preflight/適用と耐久receipt、同N1/identityで製品画面・復旧照合 | docs公開を実import/復旧PASS/通常会計GOにしない |

設計整備の続きはこの既存責務に沿って閉じる。今回文書公開だけでBacklogを実装完了へ変更せず、全移行・全会計の設計/実装完成とも報告しない。
