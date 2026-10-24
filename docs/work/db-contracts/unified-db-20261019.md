# EMR-446 / EMR-466 統合DB契約

<a id="compatibility-scope-20261024"></a>
## 2026-10-24 現在の適用範囲（466）

既存consumer互換の統合再検証は、確定新仕様consumerへの適合と履歴保全の確認へ読み替える。現行readerの形状保持を受入条件にせず、必要callerを同じ完結単位で置換する。checksum/ERD整合・院別安全境界と試験NOT_RUNは維持する。 詳細根拠は[全票監査の採用是正方針](../emr/backward-compatibility-audit-20261024.md#accepted-corrections)。以下の旧日付の記録はこの限定追補と区別する。製品置換・実行受入済みではない。

この文書はDB保存契約とconsumer引渡しを管理する。実行状況は各行で区別し、schemaの存在をAPI/UI/認可/全writer接続・実DB適用の証明にしない。PR #727再利用・#728統合は作成時の経緯。現在は#727/#729がmainへmerge済みで、基準mainは`5c5219965ad44258b8ccd918ffe52179a6139ade`。以下の過去の公開／統合手順は履歴であり、閉じたPRへの追加を要求しない。consumer接続・実DB受入は別。2026-10-20のdocs同期ではmerge・SQL／model変更・DB適用は行わない。

## 正本・範囲

作成時の基準main eb589ed1cde9291ce0e787c9aef074d9e8614729。保険4列はPR727 2707af868、再診S1はPR728 af2bc69c1。9単位は予約/施術/会計の先行調査の件数で、全DB要求の網羅数ではない。追加で非表示会話要約、証明書発行版、EMR464の預り金を照合する。未承認の小数精度/税規則・Phase2の実機能を混ぜない。

- 再診: [業務FRと6AC](../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-line-revisit-intake)、[詳細契約](revisit-intake.md)、[19writer/4reader](revisit-writer-reader-matrix.json)、[C01–C26](revisit-acceptance-matrix.json)。
- 施術: [コース属性](../../spec/screens/settings/master-trimming.md)、[予約構成/時間](../../spec/domains/reservations/pages/reservation-management.md#booking-course-option-duration-20261017)、[CL-R08/FIN-R21](../../design/domains/trimming/technical-spec.md#trimming-price-version-candidate-20261019)。
- 会計: [医師確認/同期/請求対象](../../spec/domains/accounting/lifecycle.md#doctor-review-sync-billing-20261017)、[版・binding契約](../../design/domains/accounting/technical-spec.md#review-version-sync-proposal)。
- 預り金: [FR-accounting-reservation-advance](../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-reservation-advance)、[資金移動と配賦](../../design/domains/finance-operations/technical-spec.md)。
- 証明書: [不変発行内容・訂正](../../spec/domains/vaccinations/lifecycle.md#certificate-editions-20261017)、[技術設計](../../design/domains/vaccinations/technical-spec.md)。
- 会話要約: [D03独立summary項目](../../spec/domains/medical-records/lifecycle.md#integrated-plan-20261017)。今回は保存項目だけで画面非表示、全文音声/手動貼付/外部writerなし。

## 保存対象の区別

「予約採用snapshot」は選択時の構成/時間、「施術記録改訂」は実際に記録した内容/調整、「請求取込元」は安定source行とその適用版、「医師確認集合」は薬剤/処置identityと数量の集合である。予定を実施へ、表示版を臨床世代へ、充当を再受領へ変換しない。

## 契約設計（採用・実装・検証を別に記録）

以下の保存資源は統合001へ作成済みの未試験候補。[001](../../../backend/migrations/001_init.sql)・[全項目ERD](../../architecture/erd-detail/index.md)・[coverage](unified-db-coverage.json)を対応させる。API/UI・既存writer参加・受入完了とは区別する。9単位/旧25起票候補は先行subsetの件数で、追加確定分を含む全要求の網羅数ではない。

| 単位 | 関連票 | 保存契約・再利用 | consumer条件 |
|---|---|---|---|
| 保険4列 | 446/454/455/456 | consultations/procedures/vaccines/checkup_typesのnullable is_non_insurance、defaultなし、旧値補完なし。既存薬剤/検査は維持 | 新規明示初期値はAPI、資格/補償保証と別。共有型owner454 |
| R1 S1 | 460/467–475/478/479 | PR728の3表/世代/8guardを再利用。NULL statusも拒否、immutable原文/receipt、RLS ENABLE非FORCE | 共通gate/全writer/認可/期待版/hash/再送は467–475。DB16件とC01–C26完了は別 |
| R1 S2 | 476/477 | medical_recordsのnullable target参照・clinic相関FK/index/guard。旧NULLを現在予約から補完しない | capture W14とreadを同最外TX契約へ。source不明と未回答を分ける |
| T1 | 448/457 | 任意短縮名、private画像key、曜日集合。既存description/price/duration/is_active/sort_order/clinic_idを再利用 | 曜日未設定は全曜日だが休診/shift/空き枠を免除しない。既存tenant scopeを勝手に全院共有へ拡張しない |
| T2 | 448/458 | 論理施術identity、course内包identity、course/コースなし別の設定資源とoption対応。clinic相関FK/対応一意 | 名前一致で同一施術判定しない。設定なしunknownと設定済空集合を区別。旧対応なしを全許可にしない |
| T3 | 480 | 予約採用時の構成・元master/根拠版・course基礎時間/option時間snapshot。option同ID一回 | server算定と枠検査は480。旧start/endは維持し現masterから旧内訳を再構成しない |
| T4 | 482 | 安定施術source行、detail改訂と価格/値引き適用段階・採用元価格/今回価格、immutable改訂snapshot | 0円を保持。二重値引なし、保存と終了・請求は別。共通binding契約を先に受領 |
| A1 | 462 | record/billingの臨床世代と確認receipt、確認対象record集合と各世代。既存confirmation履歴を維持 | 対象は既決の薬剤/処置だけ。検査/物販まで再確認範囲を拡張しない。全writer参加なしで有効化しない |
| A2 | 481 | typed source/最後適用版・原値・行版/manual_overrideと差異。旧unknownをfalseにしない | 共通binding契約固定→482取込→481三分岐統合検証。T4完了を契約固定の着手依存にしない |
| A3 | 483、459/465 | 金額了承before/after計算revisionはContent、会計／行版はCAS。旧complete UUID/hash列を復活させず489のbinding/resultへ接続 | 理由不要了承のevent契約とrequest_hash二重管理整理は[限定DB依存](#price-ack-db-handoff-20261020)。停止支払は465の新受領gateを再利用。全金融receipt一括再設計なし |
| A4 | 484 | source別の実施/準備ready・actor/時刻/取消の履歴。既存statusの意味と現billing参照を保持 | 未実施/後日/中止と予定額を区別。持帰準備≠引渡/出庫。個別中止料契約を消さない |
| 隠し要約 | D03既存票要照合 | record別の独立summary・版・出所、旧行NULL | Phase2 writer/UIを追加しない。スタッフ診療記録を上書きしない |
| 証明書 | D08既存票要照合 | 接種/証明種別の系譜、immutable edition、元接種版/基本情報snapshot、layout別artifact/固定描画版、操作結果 | 通常縦横は同じ最新版、狂犬病は別種。訂正commit後の旧版有効出力拒否、旧原本保持。発行で接種を自動作成しない |
| 預り金 | 464、関連257 | clinic/元owner口座、受領/充当/解除/実返金の追記ledger、元受領・元充当参照、任意予約関連、操作結果 | 同院同ownerのみ、同owner別予約/petは確認付き。自動失効/超過入金自動振替なし。解除は新受領/実返金でない |

## 預り金の一括DB要求（464との調整用）

- owner口座はclinic+ownerで一意。残高の出所は「受領累計−有効充当−未充当実返金」で追跡し、キャッシュだけを正本にしない。
- 受領は元方法/実受領日時/実額/actor/任意予約参照。充当は元受領・請求・確認額、解除は元充当への参照、実返金は元受領・実方法/日時/額を持つ。履歴を上書き/削除しない。
- 解除・減額・取消で元owner残高へ一度だけ戻す。混合支払の内訳はstaff確認であり固定優先順をDBに埋め込まない。
- 口座rowの直列化、元受領ごとの残額、元充当を超えない解除、合計非負、同scope操作key/hash、必要監査/成功結果の原子性を検証する。入院専用の別残高を二重定義しない。
- 既存payments/payment_splitsの受領を再作成せず、預り金充当を新cash受領へ写さない。残高0への推測backfillなし。旧金融データの適合/実DB移行は別。

## 共通境界と互換

既存writerの省略は旧値unknownとして表現し、新たに確認済み/実施済み/手修正なしへ自動変換しない。新たなFKはclinic相関とRESTRICTを基本とし、臨床/金融原本の連鎖削除を作らない。新規RLSは既存helperのENABLE非FORCE境界に合わせ、appのclinic/患者/操作認可を代替しない。staff actorは主所属医院と操作医院が同じとは限らないため、既存所属/権限検査をconsumerで維持する。

現時点でstaff/system監査とLINE原文監査は別sink。原文をstaff/systemへ偽装せず、金融・証明発行では既存staff actorと必要監査を使う。実DB旧値件数/seed/全API/実機受入は未検証。

## 未採用・対象外を明示

EMR365のワクチン小数精度はStartedの契約検討で、具体的scale/DDL採用は未承認と本文に明記されている。全master decimal化はしない。税率/丸めや臨床数値を新しく決めない。EMR257のCompletedは入院前金の設計完了で、464の実装完了を意味しない。459/465のDB追加不要を、預り金専用DBも不要という意味にしない。

## 統合検証と公開

正規生成詳細ERD、001 checksumと旧main/単独727/単独728のfail-closed、既存consumer互換を一体で確認する。最新本人指示「試験はしなくていい」「マイグレーションファイルをそのまま」により、新規隔離DB試験および任意テストは実行しない。ERD生成と差分読取はファイル作成として行う。実環境DB/保持データ/旧dump/共有volumeに触れない。統合先で内容一致後に728の重複を整理し、未確認の内容を失わない。

## 作成した物理契約とconsumerの未完境界

[coverage JSON](unified-db-coverage.json)に各単位・資源・残consumerを列挙する。今回の選定は技術的な保存契約であり、新しい業務採用ではない。現行241表（不要・重複定義の整理後）。先行164表から確定不足を追加（詳細はsource trace）。S1の原文・receipt・対象世代は保持した。

- T1/T2の初期version=1は機構開始時の現在状態counterで、旧master履歴を捏造しない。rule子変更は新rule set親を先にlockしてから行い、親のversionを増分する。コース内包identity・master/rule構成の検証とsnapshot captureは同TXのconsumerが所有する。
- T3は予約pet/構成・course基礎時間/option時間/元版を保存し、option同ID一回・時間合計と件数を延期制約で照合。T4は安定source行と施術改訂、採用元価格/今回価格/値引き適用段階を保存する。既存価格の精度・税/値引き計算順を新しく決めず、0円とNULLを区別。JSON detail_valuesは明示DTOのsnapshotであり既存row全体や任意client属性を保存しない。
- A1はrecord/billing別の臨床世代・集合digest、確認時の全record集合とID/identity/数量snapshotを保持する。既存治療writerから親世代を自動更新するtriggerは作らない。薬剤/処置の集合変更writer・会計直接変更・確認・確定の全参加と現在集合検証は462で必要。単価/割引/表示順や検査/物販だけで医師再確認へ拡張しない。
- A2はtyped source/最後適用版/原値/手修正状態とbinding版を保存。古い手修正状態はNULL=unknown。追補でlifetime uniqueをactive claim uniqueへ訂正し、取消済の元履歴と理由付き解除を保持した選択再請求に対応する。同期/手編集はbilling行を先にlockし両版CAS、確定後自動同期禁止。481契約→482取込→481三分岐検証の順を維持する。
- A3の金額了承は対象行・before/after計算revision／snapshot・提示額／条件と、検査した会計／行CAS版・actor／時刻を分離して保持する。calculation_versionは了承AFTER revisionのmapping案、保存後版は共通結果へ。request_hashの整理は下記限定依存であり、独立した再送判定や無条件HMAC複写を新設しない。同操作の必要監査・成功結果と同TXで保存し、通常単価/割引理由を新必須にしない。プラン4種明細の価格変更監査は既存audit_logsへconsumerが保存する。その他行/確定会計訂正等の既存必須理由は維持。
- A4はsource別世代と実施/準備/取消履歴を保存する。self-row世代は品目/数量/実施状態等の変更で増分し、単価だけでは準備を失効させない。最新世代と履歴、未実施/中止・診療種別の請求対象判定をconsumerで照合する。履歴の存在だけで請求可能にしない。
- 薬剤3値は1回製剤量＋単位、用法文、処方総量＋単位のnullable欄。既存dose_amount_mgはbasis次第で1日量を含み、quantityは会計数量なので、旧値を推測コピーしない。新数値の既定/計算式/テンプレートはない。既存犬猫の承認済み用量安全条件は変更しない。
- 証明書は接種＋種別のseries、追記edition、固定snapshot、様式/template/renderer/digest/private object、操作結果。通常縦横は同じseriesの最新版。current viewは元接種版一致＋最新editionのみを返すが、認可やダウンロードURLを与えない。訂正writerは接種versionを必ず進め、出力直前に現在認可/current/source版を同じ整合readで再確認する。render/uploadはDB lock外。物理印刷完了をDB成功結果と同一化しない。
- 預り金は口座row lock、連続account revision、元受領残額/元充当解除上限、原子操作件数・残高照合で保存する。元ownerは履歴から変えない。applyは現在請求ownerを照合、releaseは元applyを参照するため請求owner後日訂正で元資金を別ownerへ移さない。混合支払の返金優先順・保険/請求残額・同owner別予約利用の確認・監査はconsumerが所有する。操作receiptは紙の領収書ではない。

## 今回の実行状態

001 hashはcoverage JSONに記録する。旧main/単独727/単独728/旧統合coreのchecksumを自動補修せず拒否する既存runnerを維持し、期待hash fixtureだけ更新した。既存のS1試験ソース・parserのtable pin/S2存在条件は新catalogへ整合させたが、再実行していない。旧保険/S1 coreの試験成功はその旧hash限定で、拡張001の試験結果はNOT_RUN。公開時の通常hook・自動CIは別記録にし、手動再実行しない。

独立机上レビュー後に#727へ通常pushし、公開blobで728由来DDL/試験を保持したことを確認してから728を重複整理する。merge・実環境DB適用・reset・既存volume操作はしない。

## 全体不足の再照合と追補

先行18単位の自己照合だけでは、全確定仕様のDB対応完了とは言えなかった。連絡先/院内予約拒否・4業務種別・名称履歴・休業/日跨ぎ・構造化初診問診・編集権/draft・予防処方/年間投薬・用量pattern・滞在/入院開始・算定/資金/締め/在庫/提示見積・検査出所・健診段階再開・院内マニュアル公開履歴・support/LINE配送保存・実来院根拠・共通患者UUID対応を追加した。[FR/AC→表列→票の追補](spec-storage-trace.md)と[完全列JSON](spec-storage-trace.json)を参照する。旧18単位のreuse記載だけを完成根拠にしない。

共通UUID方式は本人の追加okで採用済み。独立UUID/元医院別ID保持、D19非流用、医院別臨床/金銭分離を維持する。ID統合・患者移行・unlink方式の再設計や実データ操作はない。旧PR728は統合保持の公開確認後にclose済み、branch/worktree/claimは保持する。

## 不要列/旧並行正本の整理

[削除・統合対照](schema-cleanup-audit.md)が、旧表・列を保持する前提の記載に優先する。共通基本情報・予約制限・シフト日時・固定計算・単一fund・本文版・外部配送の保存責任を一元化。医院内ID/診療会計履歴/採用済みCygram/非表示要約は維持し、実データ移行は行っていない。

## 金融帳票版のNULL参照整合（レビュー訂正）

`financial_document_editions.billing_id`は計算版がない場合も医院付きFKで実在/医院帰属を検査し、常に親document identityのbilling_idとNULLを含め一致させる。計算版があればbilling_idを必須にし、計算版/医院/請求の既存複合FKを有効にする。受領証・預り受領証・返金帳票は計算版を持たない正当な場合があるため、請求関連を保持するだけで計算版まで一律必須にはしない。元fund/返金eventは親identityで保持する。新しい帳票種別・番号規則・実データ移行は追加しない。

S1および指摘されたmigration/CSV fixtureの入力を現行DDLへ合わせた。これは試験ソースの整合であり、追加試験・DB適用は実行していない。既存backend consumerの切替は引き続き別担当。

<a id="price-ack-db-handoff-20261020"></a>
## 2026-10-20：483と480からの限定DB条件（未実装）

[483の技術mapping](../../design/domains/accounting/technical-spec.md#price-acknowledgement-handoff-20261020)を446が保存契約へ固定し、466が同一採用revisionのDDL／ERD／model／fixtureを統合する。計画の受付と実装・適用・runtime合格は別である。

- A：現行billing_price_acknowledgements.request_hashはbytea NOT NULL／32 bytes制約。共通binding fingerprintを内容同一性と再送判定の唯一の正本とし、既存operation_key＋完全scope参照で十分か、共通binding明示参照が必要かを比較する。operation_id追加／request_hash削除は未決。clinic・actor・intent scope、既存(clinic_id,billing_id,operation_key,billing_item_id) unique、FK／immutable／RLS、旧履歴／seed、writer／reader／API、成功receipt保証の前後差を公開する。旧値の再解釈・ゼロ埋め・無目的なHMAC複写はしない。現ack表に共通成功receipt必須triggerがあると仮定しない。
- B：現行billing_change_events.intentはfinalize/correct/cancel/collect/refund/reallocate。価格だけの了承に理由を新必須としない適合event経路が必要で、専用price_acknowledgement intentは第一候補に留める。correct/cancel/refundの非空reason要件を維持し、偽finalize・ダミー理由では回避しない。snapshot／ack／event／必要監査／成功receiptの同TX、旧event reader／enum／API互換を同時に示す。
- T3：EMR-480がコースなし時の区分基礎時間と出所／版を保持する限定要求とfixtureを446／466へ渡す。現guardはcourse_id NULLならcourse_duration_minutes NULL必須で、時間合計はoption＋COALESCE(course_duration_minutes,0)。区分30分＋option10分＝40分を満たせず、親／子guard両方が対象。course欄へ区分値を偽装せず、guard無効化・基礎時間省略もしない。具体列／方式は未採用。

受入fixtureはCASだけ更新した同keyと了承版変更の競合、成功応答喪失／並列再送の重複なし、scope誤対応・grant失効、理由なしprice-only成功とcorrect等の理由欠落拒否、snapshot／監査／receipt／遅延制約失敗の全rollback、旧履歴保持。実001相当の制約・RLS・deferred triggerを使う承認済み隔離DBの証跡が必要で、AutoMigrate double・過去PRのPASSでは代替しない。今回は未実施。

共有試験基盤は446が最終storage/model対応、466がschema_drift_test.go旧登録と必要最小限のinternal/testdb整合、489がoperation固有fixtureと実001／RLS／deferred制約受入を所有する。全backendの一括修復や削除model復活へ広げない。

001凍結は2026-10-20という暦日ではなく、別途承認した初回本番切替時。その後は連番migration。旧STG復旧は新main／001受入と別で、過去のreset／seedを再実行しない。全件点検の責任境界・未取得資料・引渡し条件は[445索引](../emr/emr-445-plan-handoff-20261020.md)を参照。全owner Doneを必要subsetの着手gateにしない。
