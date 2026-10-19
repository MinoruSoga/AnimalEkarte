# EMR-446 / EMR-466 統合DB契約

この文書はDB保存契約とconsumer引渡しを管理する。実行状況は各行で区別し、schemaの存在をAPI/UI/認可/全writer接続・実DB適用の証明にしない。main向けDB PRは既存 #727 を再利用し、#728の成果は取り込み一致確認後に重複整理する。merge・force push・既存DB適用は今回行わない。

## 正本・範囲

基準main eb589ed1cde9291ce0e787c9aef074d9e8614729。保険4列はPR727 2707af868、再診S1はPR728 af2bc69c1。9単位は予約/施術/会計の先行調査の件数で、全DB要求の網羅数ではない。追加で非表示会話要約、証明書発行版、EMR464の預り金を照合する。未承認の小数精度/税規則・Phase2の実機能を混ぜない。

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
| A3 | 483、459/465 | 既存complete UUID/hashは再利用。金額了承before/after・対象行/会計/算定版と結果を結ぶ | 理由を新必須にしない。停止支払は新受領consumerの変更でDB追加不要。全金融receipt一括再設計なし |
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

[coverage JSON](unified-db-coverage.json)に各単位・資源・残consumerを列挙する。今回の選定は技術的な保存契約であり、新しい業務採用ではない。全246表。先行164表から確定不足を追加（詳細はsource trace）。S1の原文・receipt・対象世代は保持した。

- T1/T2の初期version=1は機構開始時の現在状態counterで、旧master履歴を捏造しない。rule子変更は新rule set親を先にlockしてから行い、親のversionを増分する。コース内包identity・master/rule構成の検証とsnapshot captureは同TXのconsumerが所有する。
- T3は予約pet/構成・course基礎時間/option時間/元版を保存し、option同ID一回・時間合計と件数を延期制約で照合。T4は安定source行と施術改訂、採用元価格/今回価格/値引き適用段階を保存する。既存価格の精度・税/値引き計算順を新しく決めず、0円とNULLを区別。JSON detail_valuesは明示DTOのsnapshotであり既存row全体や任意client属性を保存しない。
- A1はrecord/billing別の臨床世代・集合digest、確認時の全record集合とID/identity/数量snapshotを保持する。既存治療writerから親世代を自動更新するtriggerは作らない。薬剤/処置の集合変更writer・会計直接変更・確認・確定の全参加と現在集合検証は462で必要。単価/割引/表示順や検査/物販だけで医師再確認へ拡張しない。
- A2はtyped source/最後適用版/原値/手修正状態とbinding版を保存。古い手修正状態はNULL=unknown。追補でlifetime uniqueをactive claim uniqueへ訂正し、取消済の元履歴と理由付き解除を保持した選択再請求に対応する。同期/手編集はbilling行を先にlockし両版CAS、確定後自動同期禁止。481契約→482取込→481三分岐検証の順を維持する。
- A3の金額了承は対象行・会計/算定版・before/after・操作key/hash・actor/時刻を保持する。同操作の必要監査・成功結果と同TXで保存し、通常単価/割引理由を新必須にしない。プラン4種明細の価格変更監査は既存audit_logsへconsumerが保存する。その他行/確定会計訂正等の既存必須理由は維持。
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
