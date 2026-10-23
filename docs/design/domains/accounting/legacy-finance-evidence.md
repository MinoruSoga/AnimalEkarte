# EMR-547 旧会計・見積の根拠保全と履歴閲覧：設計方針・未受領契約

| 管理項目 | 記載 |
|---|---|
| 文書ID・種別 | design.accounting.legacy-finance-evidence／移行保存・閲覧の技術設計候補 |
| 決定状態 | 業務保全原則は会計/見積正本を継承。技術設計は未完、closed schema/DDL/API未採択 |
| 実装状態 | 未実装。mapping設計入力はproducer報告として受領、新sidecar bindingは未受領 |
| 実機検証状態 | tests/build/codegen/DB/import/Docker/browser NOT_RUN |
| 文書確認日 | 2026-10-23（構造/参照確認。実行受入ではない） |

2026-10-23。[EMR-547](https://app.plane.so/baritechllc/browse/EMR-547/)（親349、Backlog、担当/期限未指定）の7ACに対応する製品側保存/閲覧契約の設計文書。**設計未完・実装未着手**。設計文書公開と未受領事項の明確化が今回の成果で、closed schema/物理DDL/APIの採択や通常会計移行の許可ではない。

調査根拠はmain `42a3890a22d645857ad8e27c3b2d7bbfc9c2e038`、docs作成baseは最新main `621d58a0834aa6984ee3c82401e3478efef45d2e`。差はClaude許可設定のcommitのみで、本作業では取り扱わない。old_db実コード・実行結果・実原本行は本consumer側では未確認。2026-10-23にproducer静的調査報告のmapping設計入力を受領した（[受領範囲/回答](#producer-static-report-and-preservation-response)）。報告されたコードSHAと既出artifact manifest/digestの一致は独立未検証で、新sidecarのclosed mapping/物理schema/APIは未採択・未受領。

今回docsのみ。製品コード/API YAML/DDL/model/producer SQL/実データ/設定/Planeは変更せず、tests/build/codegen/DB/import/Docker/browserはNOT_RUN。実PHI/原値行/ローカル実path/原本ファイルをGit・票・通常ログ・チャットへ転載しない。merge・レビュー依頼は行わない。

<a id="legacy-evidence-purpose-and-ownership"></a>
## 1. 根拠・業務目的と単一owner

業務判断窓口はMinoru Soga。旧請求の意味不明状態と旧見積の割引原値を失わず、製品内で原値/変換/不明を区別して履歴参照する。確認のために通常請求を捏造したり二重台帳へ転記したりする工程を除く。計画指標は7ACと保存→read/UI・担当・未受領契約が対応すること、未説明差分0。実業務の時間/操作数改善は未測定、保存/画面受入PASSではない。

| 参照 | 継承する境界 |
|---|---|
| [221：証拠範囲](../../../work/emr/emr-221-import-contract-design/DESIGN.md#source-evidence-scope)・[契約変更案](../../../work/emr/emr-221-import-contract-design/DESIGN.md#proposed-import-contract-boundaries)・[正式判定/耐久receipt](../../../work/emr/emr-221-import-contract-design/DESIGN.md#formal-eligibility-and-durable-receipt) | 原本identity/digestを固定、完全性と独立適格性を分ける提案、別revision原本保持、耐久receipt。同案を実装済み/正式import承認へ昇格しない |
| [formal安全境界](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#formal-import-safety-boundaries)・[再開gate](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#formal-import-producer-resume-gate)・[作業責務/実行binding補足](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#producer-work-ownership-and-runtime-binding) | 既存formal v1/profileのstrict検査、MIG-41がproducer作業項目を所有する既存境界、quarantine/PHI保護、fresh-only/band保護。未知版fallback/自己申告modeを追加しない |
| [会計正本](../../../spec/domains/accounting/lifecycle.md)・[会計技術仕様](technical-spec.md) | 通常請求確定・実受領/配賦・返金・原子receiptは正規根拠とownerに従い、証拠参照で代替しない |
| [見積正本](../../../spec/domains/estimates/business-rules.md)・[507/508契約](../../../work/emr/emr-507-508-selected-wire-contract-20261021.md) | 提示/計算revisionの意味が定まった型を維持し、未知旧割引値を無定義JSONへ詰めない |

| 作業項目owner | 本件の必要subset・除外 |
|---|---|
| 547 | 製品側immutable evidence保存、限定importer接続要求、read-only一覧/詳細UI。既存票/実装sessionへ無断追加しない |
| 349 | 配置・bundle受領・preflight/importの唯一の作業項目owner。547が別import実行経路を作らず、349という票を実行主体に読み替えない |
| 221 | 版付きevidence-only受領契約、identity/digests/独立適格性/receiptとの整合 |
| 446/466 | 必要物理保存schema、clinic複合FK/unique/RLS/immutable/index/model/ERDを受領・統合。架空物理名を予約しない |
| MIG-41 | producer作業項目ownerとして候補生成・source mapping/原本bytes/pins/receipt引渡しを追跡。既存producer責務はold_db HAC-CSV-1/HAC-PIPELINE-1にあるが、新evidence-onlyへの対応は候補で、現行producer main `a38e71f904460ba3d33194e817bafcbb78886d0c`と静的mapping設計入力は報告として受領済み。新sidecarのclosed mapping/実artifact binding/実行receiptは未受領。MIG-41という票自体が生成処理を実行する意味ではない |
| 361 | 同一runのN1画面/復旧実機受入。製品保存/read API/UIの実装ownerではない |
| 505/508 | 正式会計/見積writer/reader・未請求候補/集計からの分離、lineageと許可されたリンクの境界 |

既存runbookに記録されたold_db HAC-CSV-1/HAC-PIPELINE-1のproducer責務と、AnimalEkarteの現行csvimportの配置/preflight/import経路は既存範囲の事実として区別する。これらを新しいEMR-221案/547 evidence-only保存へ接続する対応は候補であり、現行producer SHAと静的mapping設計入力は[報告として受領済み](#producer-static-report-and-preservation-response)。新sidecarのclosed mapping/実artifact binding/実行receipt/producer-consumer接続は未受領。既存CLINIC_CSV_IMPORTのscript/consumer経路が実在することだけでは、新契約への接続済み・実行済みを証明しない。実際の操作者・対象環境・個別実行許可は同runで確認が必要で、今回未確認。作業項目ownerやdocs上の責務を実行済み証跡/操作者名の代わりにしない。

<a id="legacy-evidence-storage-roles"></a>
## 2. 保存方針：別immutable evidenceの役割

通常`billings`/`estimates`/計算revisionの空欄、claim_state、未定義JSONへ押し込まず、別のimmutable legacy finance evidenceとして耐久保存する。以下は**論理的な役割候補**で、物理table名/列型/JSON exact/制約DDLは未採択。

| 論理役割 | 保持する意味・不変条件 |
|---|---|
| batch | clinic/source、run/revision/stage build、source snapshot identity、原本/bundle/mapping digests、contract version、import identity、耐久成功receiptの結び付き。別医院/別snapshotのreceiptを流用しない |
| source dataset snapshotとsource record occurrence（原行証拠） | dataset snapshotは検証済み取得集合、occurrenceはその中の個々の原行。kind=billing又はestimate、snapshotを含むstable evidence identity、構造化原キー（判明分）、対象原値payload/原bytes digestを保持。集合と原行を混同せず、再送でoccurrenceを複製しない |
| derived projection revisionと入力lineage | 各原行由来projection revisionはinput occurrence 1..manyを参照し、各occurrenceはprojection revision 0..manyから参照される多対多。集約入力の全occurrenceと規則を保持し、単一原行へ縮約/集約projectionの原行数分複製をしない。mapping/transform/encoding版別の派生値・意味検証・保留理由をimmutableに追記。通常請求/見積の計算revisionではない |
| verified-target-link event | 同院owner/pet/record/billing/estimate間の関係まで検証済みのnullable link。期待link revision・前event参照・理由・actor/timeと追加/訂正/撤回をappend-onlyで保持。原snapshotを変更せず、最新リンクを無検証再計算しない |

446/466への受領はclinic付き複合FK、identityのunique、clinic RLS、immutable（通常UPDATE/DELETE拒否）、検索index、storage model/ERDとproducer/consumer同一版を含む。business tableへの検証なしlinkや、名前/日付/金額一致だけの新ID推定は禁止。link不明はnull＋理由、後続検証は元snapshot不変の追記event。通常appへ任意作成/更新/削除APIを出さない。実際の取込権限・link追記経路/監査/取消方式はclosed contractの受領残。

<a id="legacy-evidence-target-link-validation"></a>
### 2.1 nullable target linkの検証条件

医院の複合FK/RLSだけでは同院の別患者への誤linkを防げない。[会計の保存帰属](../../../spec/domains/accounting/lifecycle.md#accounting-owner-transfer-financial-attribution)と[患者譲渡時の保全](../../../spec/domains/owners-pets/lifecycle.md#owner-transfer-draft-retention-20261020)、[billings](../../../architecture/erd-detail/billings.md)／[estimates](../../../architecture/erd-detail/estimates.md)／[medical_records](../../../architecture/erd-detail/medical_records.md)の正規関連を根拠に、埋めた全targetの組合せを検証する。以下は必要な論理条件で、具体reader・履歴証明・lock/closed schemaは受領残。

| 値を持つnullable target | 検証する関係（全対象の同clinic・実在確認に追加） |
|---|---|
| ownerのみ | 検証済みsource対応が当該ownerの履歴帰属を証明すること。氏名/住所の類似から作らない |
| pet | 検証済みsourceが同petを指すこと。ownerも指定した場合は対象履歴に正当なowner/pet帰属を証明する。現在のpet ownerとの一致だけでは過去帰属の証明にも否定にもならない |
| record | source対応と当該recordの保存済みpet/owner関連を検査し、指定したpet/ownerとも履歴基準で整合すること。別患者のrecordへ付け替えない |
| billing | source対応と当該billingの保存済みowner/pet/medical_recordに加えhospitalizationおよび明細source_pet_id/source_record_idのlineageを照合する。supplied owner/pet/recordの非null値が、証拠の指す正しい保存帰属/対象明細に属することを検証。下記の複数頭条件を守り、ヘッダNULLから無関係患者へのlinkを許容しない |
| estimate | source対応と当該estimateの保存済みowner/pet_id/medical_record関連を検査する。supplied petが非nullのときだけ、非nullの保存pet_idとの直接一致を必須とし、矛盾は拒否。supplied petがnullでも、当該estimateのsource対応・保存帰属が検証できるestimate-only linkは許容し、架空petの指定を要求しない。null petをverified petとして表示/filter使用しない。recordがある場合は追加でrecord側のpet/帰属とも整合を検査する。record不在だけで正しい独立見積のpet linkを拒否せず、保存pet_idがnullならそれ自体を患者の証明にせず別の検証済み正規関連が必要。billingも併記する場合は両者が同じ検証済み履歴関係に属する根拠を検査し、同院/同名だけでは結ばない |

[billing header](../../../architecture/erd-detail/billings.md)のhospitalization_idがある場合は[入院](../../../architecture/erd-detail/hospitalizations.md)の保存済みowner/petと、source evidenceが結ぶ対象の帰属を検査する。[billing items](../../../architecture/erd-detail/billing_items.md)のsource_pet_id/source_record_idがある場合は対象明細の患者/元recordと、証拠・supplied targetの関係を検査する。headerのNULLは、入院/元record/対象明細との矛盾を無視する根拠にならない。

複数頭会計では全明細petが単一petへ一致する条件を新設しない。証拠がheader全体を指すのか特定明細/元record/入院の部分集合を指すのか、保存済みlineageで検証する。会計に無関係な別明細のpetを含むだけで証拠をその患者filterへ露出させず、対象との結合を明示できる関係のみverifiedにする。履歴上のowner帰属を現在pet.ownerで置換しない。組合せの意味/対象粒度/具体readerを閉じられない場合はverifiedとせず未リンク保留。これは既存multi-pet会計を禁止する仕様ではない。

targetがnullならその関係を確認済みとは扱わない。他の非null target同士の矛盾をnull欄で隠さない。無関係な同院患者や矛盾する組合せは拒否し、証明不足は当該linkを未リンク保留にする。部分的に証明できたlinkを採る場合も、採用する全組合せが整合することを要し、未証明targetをverifiedに含めない。譲渡前の正当なowner帰属を現在ownerへ自動移管せず、履歴上の関係証明とread時の現在認可を別検査にする。

追加/訂正/撤回は期待link revisionと現在の前eventを同TXで照合し、競合時は追記も有効link切替も行わない。原eventをUPDATE/DELETEせず、撤回も新eventで残す。期待版/競合error等の具体wireは未採択。認可付きowner/pet filterはこの検証を通った有効な関係だけへ適用し、未リンク保留や撤回済みlinkを他患者への検索入口に使わない。

全原本bundleは既存の保護領域へ保管する。同時に画面に必要な対象原値をDBへ耐久保存し、外部bundle/fileの喪失で詳細画面が空になる設計にしない。raw payload bytesを外部参照だけにせず、DBに保存する対象範囲・encoding/bytes digest・typed projectionとの対応を選定する必要がある。大容量の全bundleを無制限にDBへ複製する採択ではなく、サイズ/保持/backupの具体設計は未受領。ローカルquarantineに置くだけで製品閲覧完成としない。

<a id="legacy-evidence-raw-derived-separation"></a>
## 3. 原値・意味検証・変換の分離

presence（値/NULL/空/未取得という概念上の区別）、文字列としての元値、意味の検証状態を別軸で保持する。closed tag enumは未採択で、[追加技術設計候補](#legacy-evidence-additional-design-candidates)では空文字もvalue内に保持する案を分けて記載する。値が不明なことと未取得を混同せず、実際の0/空文字/NULLを互換のために潰さない。raw、derived、transform rule/版/根拠を分け、rawをbinary floatへ通したり不明を0補完したりしない。exact JSON/decimal文法/文字コード/サイズ上限はproducer-consumer共通schemaで未採択。

| 対象 | 保存/表示する原値と由来 | 禁止する意味補完 |
|---|---|---|
| 旧billing | 旧status、TSeisan_Flg、KKarte_Kbn（原列表記/変換は静的報告受領済み、実artifact照合待ち）、各日付、請求/現金/カード額を独立原値として保持。元の意味確認状態/保留理由を併記 | 意味不明TSeisan=1やshell生成定数1だけでcompleted認定しない。旧pending/waiting等を一律draft/finalizedへ変換しない |
| 旧精算証拠 | KKarte_Kbn=4は旧精算完了の証拠として区別し、旧精算日を原値として保持 | 旧精算日=新請求確定日とせず、不明な新確定日/終了event/入金配賦を生成しない。定数finalizedから終了eventを生成しない |
| 旧estimate割引 | header discountのraw値と丸め後値、変換規則/版/出典をそれぞれ保持 | 丸め後だけ残して原値を捨てず、割合/金額/適用順を推定しない。`estimate_calculation_revisions`の未定義JSONへ投入しない |
| 旧estimate保険 | insurance=0が生成定数ならgenerated_constant由来を保持して表示 | 原保険額が既知0円だったと解釈しない。生成定数/未取得/意味不明/確認済み値を分ける |
| 元PK不明 | 将来の検証済みartifact locator採用候補で原bytes中の場所と出典を辿る。現producerには元バイト位置locatorなし（[producer報告](#producer-static-report-and-preservation-response)） | 架空source keyを作らず原PK表示しない。artifact digest＋locatorを元レコードの一意性証明と呼ばない |

元table/column・変換式・構造化PKのmapping設計入力は[producer報告](#producer-static-report-and-preservation-response)として受領済み。新sidecarのclosed mapping、意味未確定列、snapshot/原キー/取得成果物の結合と実bytes対応は受領残。old_dbコードをconsumer側で検証済みとは書かず、推測でproducer SQLを修正しない。

<a id="legacy-evidence-formal-import-separation"></a>
## 4. evidence-only保全受入とformal importは別

| 判定 | 本件の扱い |
|---|---|
| clinic/provenance/bytes digestが検証済み、業務意味だけ不明 | evidence-only保全の対象にできる。通常台帳はHOLD。保存/閲覧できることをformal適格性/完全性PASSへ変換しない |
| provenance自体不明、bytes digest不一致、医院不明 | quarantine維持。製品保存受入で正当化しない |
| 通常台帳へのformal import | 既存v1/profileのstrict完全性/税/親子/FK/payment/clinic/target等の全条件と既存owner判断を維持。本書から許可しない |

evidence-onlyは**別名・版付きの明示契約**を221/MIG-41の作業項目ownerが実producer/consumer双方で整合させてから接続する。具体version文字列/manifest exact/parserは未採択。既存v1を緩めず、manifest自己申告や不明payloadの内容からmodeを推定しない。rehearsal pinや既存partial formal拒否をevidence-onlyの実装済み根拠にしない。

fresh-only/ID band/既存import prefix/lockの規約は対象profileの既存契約を維持し、本件保存で変更しない。通常台帳/証拠保存/除外集合/受入receiptの母集団を合算せず、source不足の事実を消さない。

<a id="legacy-evidence-identity-retry-and-atomicity"></a>
## 5. 再送・訂正・原子保存と結果不明

元entity（clinic/source system・系譜/source table/構造化された元複合キー）と、検証済みsource snapshot identityを含むstable source evidence identityを分ける。現在のsnapshotを含む証拠identityを弱めず、contract/mapping/transform/encoding版・run/build/stage生成IDを原本identityへ混ぜない。同stable identityの同じ原本bytesは同一証拠へ解決し、原本bytes不一致はconflictで上書きしない。digestをidentityへ足して異内容を新規化したり、契約版の更新だけで原本snapshotを重複生成したりしない。原本自体の変化は別source snapshot/revisionの証拠として関連付け、旧原本を保持する。

同じ検証済み入力occurrence集合へのmapping/transform/encoding変更は別のimmutable derived projection revisionとして追記する。入力が複数なら全occurrenceへのlineageを保持し、単一原行への従属として扱わない。派生値の差を原本改変conflictと誤判定せず、同じ派生版/内容の再送は既存結果を参照する。派生版の識別・同版異内容conflict・具体番号/encoding/closed payloadは221/MIG-41との受領残。原本identity、派生revision、各受領batchの実行identity/成功receiptは別概念であり、再処理batchやreceiptが新しくても原本を増殖させない。

元key不明時の検証済みartifact locatorは将来設計の条件付き案であり、[producer報告](#producer-static-report-and-preservation-response)では元バイト位置locatorは存在しない。snapshot/原キー/取得成果物を結ぶ方式は未選定。locatorを将来採用できても**証拠位置**の同定であって、元レコード一意性は保証できない。artifact bytes digestは整合検査に使い、不一致を別identityへ逃がさない。別artifact間の同一人物/取引や同内容の重複を推測統合しない。不明identityでどこまで受けられるか、並行取込のlock/unique・conflict応答はclosed consumer契約の受領条件。

設計目標はbatch/dataset snapshot/原行occurrences/初期projection revisions/全入力への多対多lineage/初期verified links/耐久成功receiptを同transactionで保存すること。既存原本を再処理するときは原本を参照し、必要な派生revisionと当該batchの結果だけを追記する。現在のimporterが実現済みとせず、必要adapter/監査sinkを221/349/547と保存ownerが受領する。成功はreceiptの耐久確認後、commit結果不明は同batch実行identityと保存済み原本/派生revisionの対応を照合する。unknown中の新identity/再importで救済せず、ローカルreportだけをDB保存成功証明にしない。

大batchを分割する場合のtransaction単位・部分成功receipt・再開/未知判定は未選定。全体原子性を架空保証せず、既存formal importerのatomicity/profileを変更しない。後続link追加/訂正/撤回は[target link検証条件](#legacy-evidence-target-link-validation)の期待link revision/前event照合・現在認可・actor/理由・監査を必要とし、具体結果参照を選定する必要があり、通常resourceの更新APIで代用しない。

<a id="legacy-evidence-read-api-ui-candidates"></a>
## 6. read-only API/UIの候補（未採択）

候補は既存prefix規則へ合わせる相対path：`GET /accountings/legacy-evidence`と`GET /accountings/legacy-evidence/{id}`、`GET /estimates/legacy-evidence`と`GET /estimates/legacy-evidence/{id}`。これは実在/収録済みrouteではない。動的normal `/:id`との登録順、typed legacy identity、closed DTO/ID型/カーソル/limit/sort/enum/具体errorは未採択。

- 認証clinic固定。会計証拠は既存accounting:view、見積証拠は既存estimates:viewを維持し、新共通grantを作らない。[現routes](https://github.com/MinoruSoga/AnimalEkarte/blob/42a3890a22d645857ad8e27c3b2d7bbfc9c2e038/backend/internal/billing/routes.go#L92)の別資源として検査する。他kindの原値を同一権限で混在返却しない。
- owner/pet filterはverified linkだけへ適用し、raw source名/番号の一致から現在対象を推定しない。未リンク証拠は当該clinicの専用一覧に置き、他院/未認可resourceへ横断しない。link訂正後も履歴の元帰属は保持する。
- 一覧候補はsnapshot ID/kind、原日付とその意味、旧状態原文、保留理由、link状態、証拠版。詳細候補は対象原値/derived/変換由来/関連証拠/全link履歴。raw payload bytesやlocal pathを一括返却せず、許されたprojectionだけへ。ID/金額は文字列を候補とし、日付不明を補完しない。
- HTTP候補は200、400 invalid filter/cursor、403 grant不足、404 missing/他院、500非機密read error。未認証は既存認証処理、具体closed error/非開示順はAPI契約で固定する。driver error/原値/PHIをエラーへ含めない。reader障害を空成功/原値なしへ偽装しない。

UIには「移行履歴参照専用」「通常請求への反映保留」を明示する。旧精算完了の証拠と新請求確定日不明を別表示し、原値/変換値、生成定数/不明/未取得を区別する。通常編集/回収/再計算/帳票/再請求ボタンを付けない。serverもlegacy IDをnormal billing/estimate IDとして受けず、通常writerの型付き入口とread lineageで拒否する。UI非表示だけでは保証しない。

通常未請求候補/会計/売上/見積集計へ旧明細が自動混入しないことを505/508と確認する。verified linkがあっても通常取込保留を解除しない。製品側保存→認可付きread→一覧/詳細UI→同identityの受入証拠が必要で、quarantine保存だけではAC5/6/7完成としない。

<a id="legacy-evidence-acceptance-fixtures"></a>
## 7. EMR-547の7AC対応と合成fixture（全NOT_RUN）

以下の番号は547本文の最小受入条件1–7と、[会計FR/6AC](../../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-legacy-evidence-preservation)・[見積FR/1AC](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-legacy-evidence-preservation)を対応付ける。業務条件は各specのみを正とし、この表は技術受領/fixture対応である。既存票を実装完了認定しない。

| 547 AC / 業務正本 | 技術受領/fixtureの接点 | 主な受領責務 |
|---|---|---|
| 1 [旧billing原値](../../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-legacy-evidence-preservation) / AC-accounting-legacy-evidence-originals | raw/derived/source statusのprojection・保持・表示fixture | MIG-41 mapping→221→547、505境界 |
| 2 [見積原値](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-legacy-evidence-preservation) / AC-estimates-legacy-evidence-discount-provenance | 割引raw/丸め/規則と生成定数のlineage・表示fixture | MIG-41、547/508 |
| 3 関連履歴 / AC-accounting-legacy-evidence-related-history | dataset snapshot/occurrence/全target関係検証・filter fixture | 221/349、547/446/466 |
| 4 再送/改訂 / AC-accounting-legacy-evidence-replay-history | identity/派生版/receipt/link期待版の保存trace | 547、221、446/466 |
| 5 認可read / AC-accounting-legacy-evidence-authorized-read | 現在認可・一覧/詳細reader・失敗と空の分離fixture | 547/API owner、既存認可 |
| 6 通常処理分離 / AC-accounting-legacy-evidence-normal-ledger-separation | normal writer/集計/legacy ID型境界のfixture | 547/505/508、349 |
| 7 合成受入 / AC-accounting-legacy-evidence-safe-acceptance | 同証拠の受領→保存→UI/復旧と機密保護の合成fixture | 547実装、361同run実機受入 |

fixtureを用意する条件と実行を分ける。非PHIの合成値で、bytes改変/provenance不明はquarantine、意味だけ不明は証拠候補/通常台帳HOLD、外部原本file欠損後もDB保存対象原値の表示、異snapshot/clinicのreceipt流用拒否、link撤回後の旧履歴保持、通常writer/集計非混入、receipt障害/unknown再送停止を記述する。追加fixtureは、契約版だけ変更して原本重複なし、同原本のmapping変更で派生revision追記、同identity原本bytes不一致conflict、同院別患者のowner/pet/record/billing/estimate組合せ拒否、譲渡前の正当帰属保持と現在認可、関係不明の未リンク保留、矛盾link拒否、同じ期待link revisionの並行訂正は片方のみ、estimate-onlyかつsupplied pet nullは架空pet要求なし/nullのverified表示なし、recordなし独立見積のsupplied petが非nullなら保存pet一致は許容/不一致拒否、recordありで直接petとrecordが矛盾すれば拒否、譲渡後も過去owner不変、billing header NULLでも入院/対象明細の帰属矛盾は保留/拒否、正当な複数頭会計を単一pet不一致だけで拒否しない、無関係明細から患者filterへ露出しない、集約input occurrence全件lineageと複数派生版/再送重複なし、撤回後filter非混入を含む。今回はfixtureファイル作成/実行も行わない。実データreceipt/producer適合/STG/本番受入とは別証拠。

<a id="legacy-evidence-unreceived-contracts"></a>
## 8. 設計残・次の限定引渡し

| 受領状況/残工程 | 閉じる対象 | 完了と誤認しないこと |
|---|---|---|
| MIG-41 mapping設計入力受領済み／closed sidecar mapping残 | [producer報告](#producer-static-report-and-preservation-response)のコードSHA・旧列/変換/PK概要はproducer報告として受領。未取得列・型/時刻・集約入力列挙、snapshot/構造化原キー/取得成果物の結合、新sidecar exactと実bundle同一性は残 | consumer独立コード/DB検証、現行19表artifact適合、producer生成済みを意味しない |
| 221 producer-consumer版付き契約 | evidence-only名/版・strict payload/bytes・identityと欠損identity・独立保全適格性/receipt・formal隔離 | accepted-exclusions提案やmanifest自己申告だけで有効化しない |
| 446/466物理保存 | batch/dataset snapshot/record occurrence/projection revision/全入力occurrenceとの多対多lineage/link履歴の実名/型/FK/unique/RLS/immutable/index/model/ERD、transaction/大batch方針 | 現normal tableのJSON欄へ暫定投入しない。DDL実装/適用未完 |
| 547 read API/UI | closed DTO/permission map・filter/cursor・error・projectionと履歴、namespace境界、保存後read/画面 | 本稿の候補routeを実在扱いしない。read-only受入未完 |
| 349/361同run受入 | 唯一import ownerの受領/配置/preflight/適用と耐久receipt、同N1/identityで製品画面・復旧照合 | docs公開を実import/復旧PASS/通常会計GOにしない |

設計整備の続きはこの既存責務に沿って閉じる。今回文書公開だけでBacklogを実装完了へ変更せず、全移行・全会計の設計/実装完成とも報告しない。


<a id="producer-static-report-and-preservation-response"></a>
## 9. 2026-10-23 producer静的調査報告の受領と回答

<a id="producer-report-evidence-boundary"></a>
### 受領根拠と検証境界

Library添付「貼り付けたテキスト(5).txt」のproducer回答を受領した。以下のold_db所見は**producer報告**であり、本consumer側で実コード/実DB/既出bundleを独立検証した結果ではない。報告のmainは`a38e71f904460ba3d33194e817bafcbb78886d0c`。主な調査対象は`sql/migration/raw-to-intermediate/020_canonical.sql`、`sql/migration/intermediate-to-mapping/030_stage.sql`、`scripts/lib/intermediate-to-mapping/stage-csv-columns.mjs`、SQL Server DDL。これはmappingの設計入力受領であり、新evidence-onlyのclosed schema/DDL/API全体の採択や実装開始許可ではない。

producerは現行コードのdigestを報告しているが、実artifact manifestのsource/run/build/contract/mapping digestが現在コード・出力bytesに一致するかは未検証。旧21表のreceiptを現行19表の適合証拠に転用しない。八王子の正式intakeはINCOMPLETE、城東候補は旧21表時代の報告であり、今回正式取込GOへ昇格しない。

<a id="producer-report-loss-and-sidecar"></a>
### 失われる情報と別sidecarの保存方針

| producer報告の所見 | consumer側の回答・保全境界 |
|---|---|
| 現行19 CSVはsource_table/source_pk、原flag、discount_raw、run/confidence等を出力しない。canonical段階でもmax/sum/trim/COALESCE/WHERE除外がある | 既存19 CSVを変更/旧形式へ戻すのではなく、対象の旧会計/見積に別の版付きevidence-only sidecarを設計する。生成/取込はまだ許可しない |
| 多数原行からcanonicalを集約、stage/CSVでは列や対象行がさらに落ちる | 原始source行、derived canonical/stage/CSV、それぞれへの入力原行集合と規則のlineage、除外理由を区別して保全。集約済み1行から原行を復元したとは扱わない |
| 同一カルテへの複数請求の一部はstageでmedical_record_idをNULLへ降格。元から未リンクの行もある | link降格と元から未リンクを別の由来として保持。canonicalの候補関係を無検証でverified targetへ昇格しない。[target link検証条件](#legacy-evidence-target-link-validation)の帰属/現在認可検査を維持 |
| 生成insurance=0、rate=0、tax=0.10、見積draft、補完1900-01-01、KNJO shell settle_flg=1 | generated_constant/default/補完としてrawと分ける。確認済みの保険額/率/税率/業務状態/日付/精算完了を意味しない |
| Mit_Nbiki/Mim_Nbikiのrawと丸め値が異なる。負値もあり得る。unit_priceが割後の実効単価になる経路がある | raw/丸め値/規則を分離し、負値を捨てず、不明0補完や割引の二重適用をしない。正規計算revisionへ未知値を詰めず通常処理HOLD |
| Kkarte_Kbn=4は旧精算状態の証拠。settle_flgの定数や意味不明値からcompletedが生成され得る | Kbn4・精算日・額は独立原値。新finalize/allocate/終了eventを生成せず、waiting/pendingからdraft/finalizedへ一律変換しない |
| Kkarte_Kbn 1/2（報告上の注釈は受付済/入力途中）からvisit_type first/revisitを生成するmappingがある | 初再診の根拠には採用できない。547で原値を保全するだけでは解消しない別のproducer/consumer修正責務として引き渡す。今回SQL変更/新しい初再診判定は行わない |

<a id="producer-report-identity-axes"></a>
### identityの3軸と未選定の結合

| 軸 | 必要な意味と混同禁止 |
|---|---|
| 元entity | 医院/source system・系譜/table/構造化された元複合キー。trim済み`pet|sno`だけで元の複合キーを代替せず、列ごとの原値/presenceを保持。stage IDENTITYや生成IDは原本identityではない |
| snapshot evidence revision | 元entityと検証済みsnapshotを結ぶ証拠版。同snapshot・同原本identityの異なる内容はconflict。別の検証済みsnapshotにおける実変更は旧証拠を上書きせず別revisionへ関連付ける。snapshot間の同一entityは検証済みキー/系譜を要し、名前類似で推測統合しない |
| delivery / derived projection | run/build/契約版/mapping版/encoding等は受領・生成の別軸。同原本に別変換が必要ならimmutable派生projection revisionへ追記。原本を二重生成したり派生差を原本改変と誤判定したりしない |

producerには元バイト位置locatorが存在しないという報告であり、snapshotと原キー・取得成果物を結ぶ方法は未選定。原キー欠落に架空key/locatorを作らず、stage生成IDやrun変更で同一原本の異内容を別新規へ逃がさない。[identity/再送条件](#legacy-evidence-identity-retry-and-atomicity)のsnapshotを含む証拠identity・conflict原則は維持する。producer報告内の「clinic+source_pk+契約版」案は、本書の原本identityへ契約版を戻す採択ではない。

<a id="producer-report-sql-and-csv-distinction"></a>
### SQLの情報欠落とCSV表現の区別

資料のCSV説明は[PostgreSQL 17 COPY公式仕様](https://www.postgresql.org/docs/17/sql-copy.html)に照合して訂正する。COPY CSVの既定NULLは非引用空欄、空文字は引用付き空文字であり、CSVという形式だけで両者が同一になるわけではない。実出力のCOPY option/bytesとの適合は別途確認する。事前のCOALESCE/trim/集約/WHERE除外で消えた区別は、このCSV表現から復元できない。SQL変換の情報欠落、CSV表現、未取得を別問題として保全する。

<a id="legacy-evidence-additional-design-candidates"></a>
### 追加の技術設計候補（closed contract未採択）

以下は設計レビュー用の候補であり、実データexporter/validator・取込の着手可能宣言ではない。既決の原値不変/正式取込隔離を弱めず、具体JSON/型/digest/物理保存/APIへの採択は別途行う。

| 候補 | 境界・未選定事項 |
|---|---|
| raw cellの必須presence tag | value（textの空文字・空白も含む）/null（SQL NULL確認済み）/missing（source schemaに列不存在を確認済み）/unavailable（取得不能・理由あり）を分ける。必須JSON項目の省略はmissing値ではなくvalidation error。生成定数はraw cellへ混ぜずderivedのconstantに置く。exact型は未選定 |
| bytesと値のdigest分離 | artifact bytes digestとsource-value digestを分ける。原値のtrim/Unicode正規化/丸めをせず、JSON escape/encodingだけの差を原値競合にしない候補。指定artifactのbytes不一致は整合検査失敗のまま、正規原値digestの仕様・比較層/変換証明は未選定。表現変更を原本変更と混同しない |
| 複数原行のlineage | 集約projectionから複数source occurrenceを参照できる形を検討する。入力集合digestだけで個々の元行への追跡を省略しない。機械可読mapping/原始取得位置と列挙方法は未受領 |
| PK不明の保全 | 将来、検証済みartifact locatorがあれば元entity未解決の証拠として保全可能な候補。PK/locator両方なければ集合の測定証拠として区別し、架空原行や元entityを生成しない。現producerにlocatorはなく、結合方式は未選定 |
| 大規模取込 | 固定manifestとchunk receipt、全chunk/lineageの検証後のbatch完了receipt、その後の履歴公開を候補とする。同chunk同内容は既存receipt、異内容conflict、未完batchを全体成功として表示しない。chunk上限/物理保存/再開wireは未選定で、現formal importerを変更しない |
| readとwrite認可分離 | read権限をimport/link訂正権限に流用しない。後続link訂正のwrite grant/実行経路は未選定。閲覧可能というだけで元帰属の追記訂正を許可しない |

<a id="producer-report-remaining-receipts"></a>
### 残る受領・実装条件（全NOT_RUN）

mapping設計入力は受領済みだが、実DBの型、日時の時刻成分、未取得原列、原キー欠落の実態、集約入力原行の列挙、現在コード/digestと実artifact manifestの一致は未調査/未検証を維持する。source provenance自体が不明なものはquarantine、意味だけ不明なものの保全とformal import適格性は分離する。21表時代のreceipt、producer報告、今回の文書公開を現行19表または新sidecarの実行receiptにしない。

次の設計受領はMIG-41のsidecar生成契約/原行lineage、221の版付きconsumer契約、446/466の物理保存subset、547の限定read/UI・通常writer分離である。実装可能なclosed範囲の確定は整合レビュー後に別途記録し、現時点では生成/取込/通常会計移行/全wire実装可能とはしない。実データ・producer SQL・API YAML・DDLは変更せず、DB操作/再生成/試験はNOT_RUN。
