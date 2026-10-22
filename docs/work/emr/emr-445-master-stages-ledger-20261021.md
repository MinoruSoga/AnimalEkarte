# EMR-445 マスタ整備：①契約・②全マスタ・③共通UIの段階台帳

2026-10-21。packet `DOCS-MASTER-STAGES-20261021`。本人09:18「1ステップずつ」「1,2,3は最優先に全てを完璧にしてから4」、09:21続行承認に基づく優先順の具体化。計画整備と実装接続と試験による品質証明は別判定。今回は文書・対照表のみで、API登録・DDL/model・製品コード・試験・環境・公開flagを変更していない。試験実行は全件 NOT_RUN。④の新規着手を許可しない。既存④成果を壊さず、進行中sessionを停止・再割当しない。

調査基準main `755f77abd2e6d63a2fc608fcfe6a518a61afcfff`。業務正本は [マスタ画面索引](../../spec/screens/settings/README.md) と各domainの確定FR/AC。現コードは差分確認の証拠であり、コードに合うよう確定仕様を縮小しない。実行状態・担当・正式依存の正本はPlane。旧[445引渡し索引](emr-445-plan-handoff-20261020.md)の全512票/229票は当時の対象数で、現在の全票契約・実装・品質完了数ではない。

## 2026-10-21 詳細着手プランへの入口

本台帳の段階境界を具体化した公開計画は、[①DB/API・24管理行（297-A設定subset追加）と生成型独立後続](emr-445-stage1-contract-start-plan-20261021.md)、[②settings21本文＋追加領域の保存・参照・管理画面](emr-445-stage2-master-start-plan-20261021.md)、[③検索72用途・共通UI](emr-450-stage3-common-ui-start-plan-20261021.md)、[③非検索操作family](emr-450-stage3-common-ui-start-plan-20261021-operations.md)を参照する。業務正本は[docs/spec/domains](../../spec/domains/README.md)と採用済み共通契約であり、本書群はその実現計画。具体caller・無修正根拠・必要差分・将来受領を分け、計画完成を実装接続や品質証明と混同しない。生成型は旧session3への固定割当ではなく独立後続の単一claim、管理票は①生成型[EMR-527](https://app.plane.so/baritechllc/browse/EMR-527/)（親445）、②入院plan保存[EMR-528](https://app.plane.so/baritechllc/browse/EMR-528/)（親448）、②診断Type/Name備考clear[EMR-529](https://app.plane.so/baritechllc/browse/EMR-529/)（親447）、②金銭補助master数値adapter[EMR-530](https://app.plane.so/baritechllc/browse/EMR-530/)（親449）。新規管理4票527–530ともReady・個人担当未割当であり、起票を実装・生成実行・試験の開始承認にしない。④未来接続とPhase1最後のマーケは先行gateへ戻さない。新実装の開始許可ではなく、実装・生成・試験はNOT_RUN。

## 1. 段階と完了の意味

| 段階 | 揃える成果 | 現時点の判定・後続への境界 |
|---|---|---|
| ① DB/API契約 | 対象操作とclosed wire、保存/版/NULL/権限/失敗/再送/原子性の選定、API正本収録、必要schema/modelとの対応、owner受領 | 下の4列を別判定。文書選定だけでは①全完了ではない。API/schema反映は別ownerの必要subsetへ渡す |
| ② 全マスタDB/API/UI | 確定属性のCRUD・保存後再取得、廃止/削除と過去参照保護、現在認可、管理画面・旧writer閉包 | ①の必要契約を使い、④consumer全機能完成を逆依存させない。既存予約/履歴を壊さない最小保護interfaceは必要 |
| ③ 共通UI | C01適用coverageとM01–21 consumer接続、共通部品再利用、具体残差/静的未確認の解消記録 | M26行は限定coverage。G1/G2とP/Uを区別。全業務画面完成をこの台帳だけで宣言しない |
| ④ 業務ドメイン | masterの型付き値/版を既存業務ownerへ接続 | ①②③の実装接続が揃うまで新規着手を増やさない。既存成果・進行中ownerの範囲は保全 |
| ⑤ 試験・品質証明 | fixture/期待値を用いた実行結果、実DB/実UI等の証拠 | 本人の後日指示でmainローカル。隔離環境や無断実行なし。①②③の実装接続判定に⑤の先行PASSを要求する循環は作らず、未試験・品質未証明を残す |

段階ごとの「計画作成」「採用コード/公開契約/owner受領と接続coverageが揃った」「試験で適合を実証した」を分ける。Ready/UAT/merge/過去PASSの一つで全段階を完了へ変えない。未知や資料待ちは該当行・該当subsetへ限定し、独立したmaster read/設定/fixtureの全停止へ広げない。

## 2. 対象管理台帳（直接22票）

22票は親/調査/追跡/実装票を含む管理範囲であり、22個の独立PR・完成数ではない。状態は今回変更していない。受領は型・採用commit・担当側確認の証拠を指し、本文に宛先があるだけでは受領済みにしない。下表の「未照合」は不適合の断定ではない。

| 票・限定責務 | 選定文書あり | API正本へ収録済み | 必要schema/modelが適合 | 後続ownerへ受け渡し済み |
|---|---|---|---|---|
| 445 親集約 | 本書・既存引渡し索引 | 直接routeなし。領域行で確認 | 直接schemaなし | 本PR公開が集約。個別受領は各行 |
| 447 臨床6種・用量調査 | 各master正本、454引渡し、498の①②subset | 基本CRUDあり。追加属性/用量pattern差分の照合残 | master存在と全属性適合を分離、用量pattern残契約 | 454/455/456/498へ必要subset。全caller受領未確認 |
| 448 予約/勤務/施術・ケージ調査 | [CSV](emr-448-493-shift-csv-selected-contract-20261021.md)、[コピー](emr-448-493-shift-copy-contract-20261021.md) | CSV/copy新wireは未収録。既存勤務CRUDとは別 | CSV共通metadata/版/exclusionの必要subset未実装 | 493/489/446/467の範囲を明示、runtime受領未確認 |
| 449 金銭調査 | [金銭限定契約](emr-449-459-465-payment-master-contract-20261021.md) | 新公開route不要。既存master read/update再利用 | fund/event型あり、旧payments参照の接続残 | 465-M/465-Cへ分割、459は追跡 |
| 450 共通UI/管理画面調査 | [M consumer台帳](emr-450-master-ui-consumer-ledger-20261021.md) | G1/G2は既存APIの状態伝播。新検索API不要 | DB変更不要 | 492/493の既存ownerと統合、G/P/Uの解消証拠残 |
| 453 明示検索の共通成果 | C01正本、既存183d004f2採用 | 新API不要 | 非該当 | main包含はPlane追補確認。再実装なし、consumer coverageは450 |
| 454 診察/処置・共通保険adapter | [CAS引渡し](../emr454-master-cas-handoff.md)、会計FR/AC | 基本CRUD/CASは既採用。C06/6種adapter差分照合残 | 既存専有成果/UAT、全必要schema適合を今回未認定 | 455/456共有、481/505は④、483は警告owner |
| 455 予防接種品目保険 | 454共有契約＋接種master正本 | 専有API既採用、現仕様全差分未照合 | 既存成果/UAT、実DB適用・受入の証明とは別 | 共有hunk454、専有writer/reader455。④引渡し別 |
| 456 健診種別保険 | 454共有契約＋健診master正本 | 専有API既採用、現仕様全差分未照合 | 既存成果/UAT、測定値/基準値の変更なし | 共有hunk454。健診保存から自動会計を新設しない |
| 457 コース属性/画像 | trimming正本と既存731成果 | 既存専有APIと画像routeを操作別照合 | 既存成果/UAT、cleanup分類証拠不明はcleanupだけ停止 | 458設定、480/482④へ型付き属性を渡す |
| 458 option対応設定 | [T2選定](emr-458-option-assignment-selected-contract-20261021.md) main採用 | 選定追加route未収録 | T2既存形状あり。anchor/digest/全CRUD参加は必要subset残 | A設定②、B採用consumer④。未接続を選定待ちへ戻さない |
| 492 予約区分/group設定 | [選定wire](emr-490-492-selected-wire-contract-20261021.md) §2以降 | 基本CRUDあり。rename/restore/C06追加差分は収録残 | 4種/版/名称stampの必要差分、全適合未照合 | 設定②、予約writer④。M09状態伝播は③ |
| 493 休業/指名停止/勤務 | [休業](emr-493-staff-leave-selected-contract-20261021.md)、[eligibility](emr-493-reservation-eligibility-contract-20261021.md)、CSV/copy | 既存勤務CRUD/lookupあり。選定休業/eligibility/CSV/copy追加未収録 | 既存assignment/positive集合あり。aggregate版/digest等追加未実装 | staff単一統合、467既存保護subset。新予約全体は④ |
| 293 全無効化入口・原子性追跡 | [監査契約](emr-293-deactivation-audit.md)＋366確定回答 | 既存staff操作、専用route不要 | 同TX監査既存、trusted actor全入口の残確認 | Delete actorは366、既存Update保持。再承認不要 |
| 366 trusted staff actor配管 | 同監査契約・Plane ACTOR-DECISION-20261021-366 | wire変更不要、client actor不可 | handler→service→repositoryの必要実装残 | 293と一つの監査経路。IP/UA追加なし、fixture責務確定 |
| 365 ワクチン小数単価 | [R2選定](emr-365-vaccine-price-decimal-contract-20261021.md) #784 main採用 | price文字列wireの段階互換は未反映 | numeric(23,4)/ExactNumeric選定、現在bigintとの差分残 | producer/consumer同版receipt・pin・実applyは別受領、未設計扱いしない |
| 369 税率変更回答 | D10非自動追従/明示再適用 | 新route不要、回答確認票 | 非該当。実医院設定/権限未確認 | 465のgateにしない。保存済会計への自動反映なし |
| 459 停止支払方法追跡 | 金銭限定契約・465と同一 | 465行に集約 | 465行に集約 | 別実装owner/別PRを作らない |
| 465 新受領guard | 金銭限定契約§2–6 | 内部guard shape選定、専用APIなし | 既存method/fund型再利用、旧usage参照の修正残 | 465-M②と465-C④。7AC全合格ではない |
| 518 package親契約 | [storage](emr-518-storage-contract-20261020.md)・[API](emr-518-package-api-contract-20261020.md) | #761で13操作/9path/58schema採用 | S0は#775採用。S1/S1a/F0は④apply依存と分離 | 521/523成果、522/524 runtimeは④、親全体Doneではない |
| 521 package master BE | 上記契約＋#775採用 | master6操作は#775 routes/DI接続 | S0/clinic初期化採用、生成型同期受領は未完 | input ae65ffd4aの旧session3向け引渡しは履歴。生成型は①詳細計画の独立後続・単一claimへ、受領commit照合まで型整合未完 |
| 523 package master FE | 上記API契約＋#775採用 | 専有client既採用 | generated/models.ts同期は独立後続の単一claim owner、担当未割当・手編集なし | 管理画面採用。実BE結合・⑤試験を製品完成と偽らない |

### 498のmaster subset（直接22票とは別の必要coverage）

498は単純な「対象外consumer」としない。同一票の用量patternマスタCRUD/revision契約は①②、治療計算と職員の明示採用は④に分ける。臨床量・数値例の資料待ちをmasterのidentity/保存/履歴の契約選定待ちと混同しない。[用量pattern契約](emr-498-dose-pattern-master-contract-20261021.md)を親採択済みとして同じ文書PRへ統合した。

| subset | 選定文書あり | API正本へ収録済み | 必要schema/modelが適合 | 後続ownerへ受け渡し済み |
|---|---|---|---|---|
| 498 pattern CRUD/revision（①②） | 同契約R2 | 複数pattern/revision DTOは未収録、旧species APIとは別 | DDL複数pattern/不変revision/FKあり。revision modelはstorage_clinical.goに既存。新anchor/reader/writer接続・必要な正規生成は未完、DB適用未確認 | master一式の必要subsetを同契約§6へ。runtime受領未確認 |
| 498 治療計算/明示採用（④） | dose-source正本、固定版/承認検査は維持 | 本PRで収録しない | 過去治療のFK/snapshotを変更しない | 488/494/484へ別接続、①②へ逆依存しない |

## 3. ②の実装単位と接続判定

| 実装単位/既存owner | 保存・再取得 | 廃止/過去参照保護 | 管理画面と受領物 |
|---|---|---|---|
| 臨床master447/454–456 | 保険boolean/null/省略・価格等の採択値と版の往復、共通save/transform単一owner | 既採用元値/版・旧欠落を現在masterで補完しない | 専有UI既存成果を再利用。481/505計算/受領は④へ |
| 施術457/458 | 属性・画像key・T2設定を現在認可で保存し再取得 | 内包service identity/unknown/設定済み空集合を保持、使用中参照を保護 | 457共有hunk、458設定A。480/482予約時点値/会計採用は④ |
| 予約区分/勤務492/493/448 | group/type設定、医院別eligibility、休業、勤務aggregate、CSV/copyの必要subset | 他院・過去履歴保持、既存保護予約は同TX現在reader/gateで守る | 明示設定・全件preview・結果照会。新予約入口④の完成を②へ逆依存させない |
| 金銭449/465 | 465-Mの許可された停止/再有効化、同TX監査、再読込 | fund/event/correction参照を同row lockで守り旧payments参照を修正 | 新候補/履歴のread契約をhandoff、PaymentCardは④465-C |
| staff/権限293/366 | trusted actor必須の既存無効化/削除、監査同TX | HTTP actor欠落をsystemへ補完せず既存正当system経路を推測削除しない | ログイン/所属/個人能力/医院別pauseを別操作として保持 |
| 小数365 | vaccine専用exact値→保存→文字列wire→表示の段階互換 | 旧未知・不正値を0や丸めで救済しない | producer実receipt待ちは限定boundary、全master停止なし |
| package521/523 | #775のCRUD/構成/医院別順序/初期化/FE管理を再利用 | 固定revision/参照保護を既存契約へ対応 | 型生成受領残のみ単一ownerへ。524 mockはruntime apply完成でない |
| その他master（M01–21各正本） | C01 M行と設定索引を各CRUD/許可属性へ対応 | 停止・削除・履歴の各正本を維持 | ③台帳は検索接続の証拠、CRUD全属性適合の代替ではない |

②の各ownerはPRに「属性の保存値/再取得」「過去参照/停止/削除」「画面と現在認可」「旧writer・shared型の受領commit」「未実行fixture」を同じ操作単位で載せる。①②③の実装残を一つの全backend完了や臨床資料待ちにまとめない。担当・claim・進行中PRは着手直前に確認し、同じhunkは単一ownerへ渡す。本書は新session割当を行わない。

## 4. ③C01・共通UIのcoverage

[M台帳](emr-450-master-ui-consumer-ledger-20261021.md)にM01–21を26行へ展開済み。既存共通成果453/409/411/247/284を再実装しない。G1は予約区分/group、G2は勤務templateのloading/error/真の0件の未接続。P/Uは実装欠陥と断定せず、scope/race/retry/復帰/未保存panelの確認残へ置く。N/Eへ検索欄を新設しない。packageにはO/S等の一般UIを適用し検索用途は非適用。

③全体はC01のL/H/Sや検索以外の同種操作も持つため、このM台帳だけで全画面完成を宣言しない。現在段階ではcoverageの未確認を残し、④新規着手の許可へ変換しない。後段⑤の試験品質証明を先行実施せず、未試験を明示する。

## 5. 隣接領域と④の境界

| 集合 | 限定的な扱い |
|---|---|
| 患者scope451/463、登録490/491 | 463はPlane09:14報告でUAT、773/796採用。OWNERS_CLINIC_SCOPE_ACTIVEをfalseへ戻さない。一覧scopeを登録の個別認可代替にせずmaster全体gateへ混入しない。490は#784選定済み、typed decimal/coverage/ContextGate等はruntime残 |
| 移行351/364/375/376 | 351の236列監査を再実装しない。[CSV列契約](emr-364-375-376-csv-column-contract-20261021.md)選定済み。producer共同receipt、実source/pin/実apply、未知履歴だけを限定残件へ |
| 488/498/480/482/461/499/514/522/524 | 既存成果を保持。498のmaster subsetは①②に含め治療採用④と分離。514既存日時CRUDをCSV/copyで再実装しない。522/524の実業務applyは④、mockは成果として別記 |
| 459/465 | 一つの実装owner、追跡票を重複PR数へ算入しない |
| 446/466/489等の共有owner | 必要schema/model/正規生成/operation subsetを型とcommitで受領。全owner Doneの一括gateや独自共通基盤を作らない |

## 6. 固定証拠・未完の更新方法

- #784採用 `b48a67d9908a8172400956edf38d3860138d9b7f`、#798採用 `582178a28961a7953f4ea8346c9629d0bb51478f`、#801採用 `755f77abd2e6d63a2fc608fcfe6a518a61afcfff`。旧「公開待ち/未merge」を現停止条件に戻さない。
- [#775](https://github.com/MinoruSoga/AnimalEkarte/pull/775) API読戻し: merged=true、2026-10-21T08:31:44Z、merge `9a1d50f9dcc7d275456f1214507ef0a55441ac98`。BE6操作/S0/clinic初期化/FE管理/524mockは採用済み。生成型同期は同PR説明と親引渡しで未完を確認。merge≠DB適用/実結合/試験PASS。
- 直接22票の最新本文を読取。454–457はUATとの親報告とPlane状態を照合したが、その状態を最新実行合格へ読み替えない。293/366のactor-only/trusted staff方針は確定、再承認不要。365旧「未採用/候補作成」は#784採用範囲では履歴。
- APIの静的根拠: `backend/docs/api.yaml`に既存勤務/lookup、capable-reservation-types、予約区分/group基本CRUD、package schemaと6master操作を確認。選定CSV/copy/eligibility/休業/458追加routeは現基準に未収録。構文やroute名だけでは各schema全属性適合を証明しない。
- 今回はdocsのみ。API具体反映は候補レビュー後の別指示で同集約PRへ必要hunkだけを追加する。DDL/model/生成物/試験/runtime/公開flagを先取りしない。受領待ち行の更新には受領commit・固定source・担当側確認を付け、宛先名だけで受領済みにしない。

## 今回5枠の限定追補

本節の基準mainは `06213d5778807f9ecbc9a8501dd579e8190a0d3b`（#804採用）。冒頭の `755f77abd2e6d63a2fc608fcfe6a518a61afcfff` は元の履歴調査基準であり、本節の追加受領範囲と区別する。

[①A〜E引渡し表](emr-445-stage1-contract-start-plan-20261021.md#stage-current-five-20261021)で今回の共有契約/生成/限定master hunkと次便の既選定subsetを区別する。直接22票＋498 master＋297-A設定subsetの①24管理行は②26resource・③検索72用途とは別集計で、24行すべて実装完了という意味ではない。493の休業元targetと297-Aの設定mapping版は既選定保存契約の①受領漏れ補正。A採用modelと同一固定commitの本人生成結果をBが最終受領し、A前の生成結果だけで①全体完了にしない。未割当を未設計とせず、計画追補を実装開始・全親Done・品質証明へ換算しない。試験NOT_RUN。

## 2026-10-22 coverage監査と完了判定の留保

[監査記録](emr-stage123-coverage-audit-20261022.md)を①②③の対照表へ接続した。監査基準は#805の `4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34`、docs作成基準は#806/#807後の `9eef73feac19c25571724372b0b19da160ba5cfb`。#806 API収録・#807保存型/DDL/ERDは限定採用、#808/#809は確認時open。今回所見と解消済み/意図した後続を混算しない。

①24管理行、②26resource、③検索72用途（57＋15）/非検索D01–D23/P01/7primitive＋18storiesは別のcoverage。調査票/親票を独立実装数に数えない。全体割引除外設定、担当者別LINE開始時刻、医院共通営業時間/休診日の保存writerは既存ownerへ技術設計残を受領し、現5セッションへ暗黙追加しない。計画記載済み、MECE/責務の閉鎖、実装接続、試験品質証明は別判定であり、現時点で①〜③を整備/実装とも全完了とは宣言しない。今回docs-only、実装・試験・生成はNOT_RUN。
