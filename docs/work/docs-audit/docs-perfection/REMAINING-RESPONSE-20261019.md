# docs残件対応の照合記録（2026-10-19）

基準main `563353287ebb08af55056e3fbfba2aca3531e599`。本人の2026-10-19 04:01:31 UTC「ok, 対応してください そしてmainに対してPRを作成して」が委任入力に示した承認範囲で、文書精査・内部設計具体化・静的根拠と未実行手順を整備した記録。実行状態の正本であるPlaneは変更していない。業務正本・設計・受入手順への案内であり、第二の業務仕様や製品Done台帳を作らない。

## 件数と完了の読み方

入力は確定文書不整合1件・その他75項目・別枠7項目。途中の74はONT-R01の割当補完後75へ訂正済み。75の**元調査時点の状態内訳**は未精査24・設計未選定20・未検証31、目的別は文書/受入定義精査16・内部/運用設計20・限定静的/資料/データ調査8・実装/実機等31。これは75個の製品不具合、75個の原子的変更、75件完了の意味ではない。

- 確定REM-OPS-01の旧Lock必須を修正。精査中に判明した慢性疾患の収録漏れ/ゼロgap断定も修正した。実機で製品不具合を再現した記録ではない。
- 24項目は下表に精査した範囲と残工程を区別。全原資料・実データの未確認を解消済みにしない。
- 20項目は既存案へ候補資源/列/キー/制約/adapter/照会/移行検査等を具体化。内部候補であり、すべての物理方式を本人承認・最終採用済みにはしていない。
- 31項目は限定静的観察・既存証跡・[対象/操作/期待/停止・復旧](REMAINING-VERIFICATION-20261019.md)を整理。今回の追加runtime実行0。過去の部分PASS/観測は対象版・範囲付きで同引継ぎに反映し、全面未実施やユーザー未回答としない。最新ケース全体の合格・コード適合・Go-live完了へ繰り上げない。

## 確認した文書修正

| 対象 | 修正と保持した境界 |
|:---|:---|
| [UAT254 Flow1](../../../ops/testing/scenarios/UAT-254-CLOSE-CHECKLIST.md)（REM-OPS-01） | 旧医師確定Lockを必須から外し、通常編集/編集権/履歴/監査と診療終了/医師確認/請求確定を分離。5flow、子owner固有保護、実LINE/token、別sign-off、締めを保持 |
| [フォーム棚卸し](../../../ops/testing/scenarios/FORM-FIELD-INVENTORY.md)と[V03](../../../ops/testing/scenarios/V03-owner-pet-staff-forms.md)（REM-OPS-09内） | 既存送出の慢性疾患5keyを収録し、未収録残件2箇所と全体ゼロgap断定を訂正。旧残余26記載と表件数25の差1件は未照合として保持。動的定義idと実行完了は保証しない |

## 項目別の今回実施と残工程

各IDは元調査の作業単位を保持する。P1は該当機能の切替/臨床・金銭・運用利用前、P2は順次精査、P3は低優先の照合。下表の『実施』は文書作業であり、記載した残工程がなくなったという意味ではない。根拠・元完了条件・依存は[機械可読の照合記録](EVIDENCE/remaining-response-20261019.json)も参照。

| ID・正本/詳細 | 優先/元状態 | 今回の文書作業 | 残る工程 |
|:---|:---|:---|:---|
| [CL-R01](../../../design/domains/medical-records/technical-spec.md#clinical-resource-fit-review-20261019) 臨床writerのowner別版・監査・receipt・共通gate適合台帳 | P1／未検証 | route群の初期配賦・gate/fence分離・receipt切替検査を設計草案化 | 全writer service/旧client/機器/移行/batch監査と競合実行は未了 |
| [CL-R02](../../../design/domains/medical-records/technical-spec.md#clinical-resource-fit-review-20261019) 編集権の候補API・session/epochを実資源と全経路へ具体化 | P1／設計未選定 | edit-right候補行・列・unique・epoch・現route差を具体化 | schema/DTO/認証対応・全経路適合・実行は未了 |
| [CL-R03](../../../design/domains/medical-records/technical-spec.md#clinical-resource-fit-review-20261019) 本人専用下書きの物理保管・端末保護方式の確定（保持方針は既決） | P1／設計未選定 | 本人draft資源・世代・比較選択UI・認可/鍵停止・未達試験を具体化 | offline鍵方式選定と同origin/端末脅威・quota実証は未了。保持業務は既決 |
| [CL-R04](../../../design/domains/medical-records/technical-spec.md#clinical-resource-fit-review-20261019) 統合入力のowner別保存・clear/copy/Undoと飼主原文 | P1／未検証 | dirty key/二基準/copy成功subset/遅延世代/原文非干渉を具体化 | IR09回答ref受渡しとDC-M01/02実行は未了 |
| [CL-R05](../../../design/domains/medical-records/technical-spec.md#clinical-resource-fit-review-20261019) 通常編集・専用訂正/過去入力/死亡後生前分と旧版対応 | P1／未検証 | addendum実requestとの差から目的別元先/版/確認/receipt候補を配賦 | 専用API適合と生前分・両順競合・監査失敗実行は未了 |
| [CL-R06](../../../design/domains/medical-records/technical-spec.md#clinical-resource-fit-review-20261019) 薬量の具体的臨床値と単位・複数パターンの対応 | P1／未検証 | 現model/PUTの単一種別値と目的経路承認パターン版の差を具体化 | #261具体値臨床承認と計算実証は未了 |
| [CL-R07](../../../design/domains/hospitalization/technical-spec.md#hospitalization-change-receipt-candidate-20261019) 入院中プラン変更・完了後訂正・ホテル料金の保存接続 | P2／設計未選定 | 安定plan ref/current projection/不変変更event/料金確認refとD10分離を具体化 | service全writer/旧履歴の移行/金銭連動実証は未了 |
| [CL-R08](../../../design/domains/trimming/technical-spec.md#trimming-price-version-candidate-20261019) トリミングの版/再送・既決終了条件と料金引継ぎの適合 | P2／未検証 | detail版と調整根拠/取込元行識別/終了との版結合/再送検査を具体化 | 全writer/保存物理対応/財務実証は未了 |
| [CL-R09](../../../design/domains/checkups/technical-spec.md#checkup-stage-recovery-candidate-20261019) 健診の段階保存復帰・旧field/値復元・監査適合 | P2／未検証 | 親子結果のstage成功ID/receipt/同ID再開資源を現hook再読で具体化 | GET復元配線/fieldサービス全件/実行は未了 |
| [CL-R10](../../../design/domains/examinations/technical-spec.md#examination-static-validation-handoff-20261019) 検査選択印刷/感染症投影のwire・旧版・mapping | P2／設計未選定 | 固定版print現model/旧版なしfailclosedを確認し選択refと感染症fixtureへ配賦 | 感染症実projectionの全経路と旧データmapping/実印刷は未了 |
| [CL-R11](../../../design/domains/vaccinations/technical-spec.md#vaccination-date-certificate-review-20261019) 接種次回日設定/当時日保持の両入口と旧snapshot適合 | P2／未検証 | 埋込hook現初期/複製clearと独立modelを分離し比較fixture表を具体化 | 医師interval・設定参照と両入口保存再読取の実行は未了 |
| [CL-R12](../../../design/domains/vaccinations/technical-spec.md#vaccination-date-certificate-review-20261019) 証明書専用実装所在と固定発行版/出力資源の適合 | P2／未精査 | 保存済みresume-certificate-review.jsonの原XLS SHA/3見本全項目・配置対応を現lifecycle115–151へ照合し一致。repo全体の実装拡張子検索を追加、専用発行経路は未特定。 | 専用実装所在と適合、旧版移行、描画golden、実印刷は未検証。元見本項目・配置の文書照合を未精査へ戻さない。 |
| [CL-R13](../../../design/domains/examinations/technical-spec.md#examination-static-validation-handoff-20261019) 機器待機の当日対象カルテ条件のserver適合 | P1／未検証 | PutWait264–313/todayVisits371–384/persist入口の再読で当日対象検査の静的不足を確認 | 修正禁止により実装と実機は未了。当日条件を未決へ戻さない |
| [CL-R14](../../../design/domains/examinations/technical-spec.md#examination-static-validation-handoff-20261019) lab実機/医院別配線・PIMS/パリティ・臨床範囲/rollout | P1／未検証 | 機器毎receiptと日常/decoder/PIMS/legacy別行、故障/復旧の検査配賦を具体化 | 配線/機種写真同定/医院実機/臨床値/運用承認は未了 |
| [CL-R15](../../../spec/lab-device-integration/README.md) 外注/画像/八王子・旧ID/版の最低継続要件調査 | P2／未精査 | 既存解析の痕跡と現利用未知を再照合、医院別六欄の継続確認票を草案化 | 外注/画像/八王子の現利用証拠・旧ID対応等は未取得。未精査完全解消ではない |
| [CL-R16](../../../design/shared/integration-contracts.md) 臨床AC/DCの意味対応・旧観察の最新性を限定再精査 | P3／未精査 | D03補足24ACとD08補足4ACを各正本AC期待値へIDごと照合、一致。DC-M01–09/DC-V01–08/C01–03と今回候補の適用を対応。 | 今回範囲の文書意味照合は完了。全臨床FR/AC・全旧コード観察の全面再監査は行っておらず、DCの実行証跡は別工程。 |
| [FIN-R01](../../../design/domains/accounting/technical-spec.md#finance-ac-execution-trace-20261019) 181本文AC＋subpage固有20ACの項目単位受入網羅 | P2／未精査 | 本文181ACと固有20ACをID別の検証群/個別結果キーへ対応。6行の割当修正を再確認 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R02](../../../ops/testing/scenarios/S08-accounting-corrections.md#s08-current-target-overlay-20261019) S08全10手順を現正本へ項目単位で接続 | P2／未精査 | S08全10手順について旧観察と現目標、fixture・対応AC・必要証跡/停止を対照 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R03](../../../ops/testing/scenarios/V02-accounting-reservation-forms.md#finance-field-targets-20261019) V02／項目棚卸し：会計精算フォーム | P2／未精査 | 精算フォーム5fieldにF0〜F6の適用/理由と支払・部分入金等の現ACを対応 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R04](../../../ops/testing/scenarios/V02-accounting-reservation-forms.md#finance-field-targets-20261019) V02／項目棚卸し：明細行編集 | P2／未精査 | 明細編集3fieldに数量/単価/認可/保存再読取の現期待を対応 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R05](../../../ops/testing/scenarios/V02-accounting-reservation-forms.md#finance-field-targets-20261019) V02／項目棚卸し：明細追加 | P2／未精査 | 明細追加7fieldに必須/形式/空クリア/その他理由と現ACを対応 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R06](../../../ops/testing/scenarios/V02-accounting-reservation-forms.md#finance-field-targets-20261019) V02／項目棚卸し：クレジット訂正 | P2／未精査 | クレジット訂正4fieldに訂正理由/元資金・方法/金額の現期待を対応 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R07](../../../ops/testing/scenarios/V02-accounting-reservation-forms.md#finance-field-targets-20261019) V02／項目棚卸し：返金 | P2／未精査 | 返金3fieldに理由/実方法/受領残上限/再読取を対応 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R08](../../../ops/testing/scenarios/V02-accounting-reservation-forms.md#finance-field-targets-20261019) V02／項目棚卸し：レジ締め | P2／未精査 | レジ締め3fieldに差額メモ条件/区分/現closing ACを対応 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R09](../../../ops/testing/scenarios/V02-accounting-reservation-forms.md#finance-field-targets-20261019) V02／項目棚卸し：見積作成編集 | P2／未精査 | 見積13fieldを現estimates ACとF観点へ対応し原本/後継を分離 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R10](../../../ops/testing/scenarios/V02-accounting-reservation-forms.md#finance-field-targets-20261019) V02／項目棚卸し：在庫品目 | P2／未精査 | 在庫9fieldを現inventory ACとF観点へ対応し旧値/出庫と属性保存を分離 | 実UI/DTO全項目差分、各AC固有fixtureと実行receiptの配置、全適用Fの実機結果 |
| [FIN-R11](../../../design/domains/accounting/technical-spec.md#finance-physical-candidates-20261019) 医師確認世代・同期由来の物理契約具体化 | P1／設計未選定 | 確認世代2資源・医師確認receipt・source binding/差異のキー/列/CASと確定時照合を候補化 | 候補の独立設計レビュー・実wire/schema/grant/実値との適合、性能/移行/実機 |
| [FIN-R12](../../../design/domains/accounting/technical-spec.md#finance-physical-candidates-20261019) 全writer／consumerのgate参加と認可の実装適合 | P1／未検証 | 現main固定版でroute/医師確認/退院取消分岐/在庫repoを限定静的照合し候補と残検査へ接続 | 全writer/consumerの呼出追跡、原記録と履歴可達、実DB並行・障害復旧。限定読取を全経路合格にしない |
| [FIN-R13](../../../design/domains/accounting/technical-spec.md#finance-physical-candidates-20261019) 金銭旧データ移行・資金対応の照合計画 | P1／未検証 | 旧値分類manifest/前後金額件数照合・認可参照境界の候補と不足時停止を草案化 | 実データ/全grant/旧欠落設定/全保存読取経路の限定調査と移行。必要な解釈だけ根拠を添えて質問 |
| [FIN-R14](../../../design/domains/accounting/technical-spec.md#finance-physical-candidates-20261019) 算定精度・snapshot形式・採番/帳票版の具体設計 | P2／設計未選定 | 十進係数/有理算定、snapshot、配賦安定順、発行版と数値fixtureを具体化 | 候補の独立設計レビュー・実wire/schema/grant/実値との適合、性能/移行/実機 |
| [FIN-R15](../../../design/domains/accounting/technical-spec.md#finance-physical-candidates-20261019) 品目保険設定・停止支払方法・割引認可の全入口適合 | P1／未精査 | 旧値分類manifest/前後金額件数照合・認可参照境界の候補と不足時停止を草案化 | 実データ/全grant/旧欠落設定/全保存読取経路の限定調査と移行。必要な解釈だけ根拠を添えて質問 |
| [FIN-R16](../../../design/domains/finance-operations/technical-spec.md) 見積全入口・提示原本・後継・確認転記の実装照合 | P1／未検証 | 既存意味契約へ候補資源キー/列/制約/投影/失敗・検証期待を追加具体化。内部提案のまま | 候補の独立設計レビュー・実wire/schema/grant/実値との適合、性能/移行/実機 |
| [FIN-R17](../../../design/domains/finance-operations/technical-spec.md) D12締め／D16月次への新金銭event adapter適合 | P1／未検証 | 既存意味契約へ候補資源キー/列/制約/投影/失敗・検証期待を追加具体化。内部提案のまま | 候補の独立設計レビュー・実wire/schema/grant/実値との適合、性能/移行/実機 |
| [FIN-R18](../../../design/domains/finance-operations/technical-spec.md) 月末未納・返金等を含む表示/cutoffの詳細接続 | P2／設計未選定 | JST半開cutoff/元event別調整と完全性を候補化し、限定式と実受領帰属差を分離 | 候補の独立設計レビュー・実wire/schema/grant/実値との適合、性能/移行/実機 |
| [FIN-R19](../../../design/domains/finance-operations/technical-spec.md) 退院情報の行粒度・固定読取・紙面配置 | P2／設計未選定 | 退院read応答/全ページ世代/1滞在有効退院事実の行と4列改頁配置を候補化 | 候補の独立設計レビュー・実wire/schema/grant/実値との適合、性能/移行/実機 |
| [FIN-R20](../../../design/domains/accounting/technical-spec.md#finance-physical-candidates-20261019) 退院会計の取消旧会計置換経路とD10再請求接続 | P1／未精査 | 現main固定版でroute/医師確認/退院取消分岐/在庫repoを限定静的照合し候補と残検査へ接続 | 全writer/consumerの呼出追跡、原記録と履歴可達、実DB並行・障害復旧。限定読取を全経路合格にしない |
| [FIN-R22](../../../design/domains/finance-operations/technical-spec.md) 在庫数量版・実績台帳・臨床小数量の具体接続 | P1／未検証 | 現main固定版でroute/医師確認/退院取消分岐/在庫repoを限定静的照合し候補と残検査へ接続 | 全writer/consumerの呼出追跡、原記録と履歴可達、実DB並行・障害復旧。限定読取を全経路合格にしない |
| [FIN-R23](../../../design/domains/finance-operations/technical-spec.md) 医師確認・全金銭intentの実API/DB並行・復旧UAT | P1／未検証 | 承認済み専用環境の前提、各intent実行/帳票matrix/外部確認の証跡単位と未実行判定を整理 | 実API/DB2session・実印刷/医院運用・税務/実契約確認。今回は禁止範囲なので未実行 |
| [FIN-R24](../../../design/domains/finance-operations/technical-spec.md) 帳票全項目のPDF・実印刷・医院運用受入 | P1／未検証 | 承認済み専用環境の前提、各intent実行/帳票matrix/外部確認の証跡単位と未実行判定を整理 | 実API/DB2session・実印刷/医院運用・税務/実契約確認。今回は禁止範囲なので未実行 |
| [FIN-R25](../../../design/domains/finance-operations/technical-spec.md) 税務分類・適格請求書全体・決済規約の外部確認 | P1／未検証 | 承認済み専用環境の前提、各intent実行/帳票matrix/外部確認の証跡単位と未実行判定を整理 | 実API/DB2session・実印刷/医院運用・税務/実契約確認。今回は禁止範囲なので未実行 |
| [IR-01](../../../design/domains/owners-pets/technical-spec.md#identity-physical-adapter-candidates-20261019) 共通subject/local医院利用mappingの内部方式選定 | P1／設計未選定 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 採用する内部方式と代替を選定理由付きで明示し、共通更新と元院履歴不変/無認可非公開/二重create防止の全接続が追える。 |
| [IR-02](../../../design/domains/owners-pets/technical-spec.md#identity-physical-adapter-candidates-20261019) 既存患者のmapping矛盾・移行対象の実在照合 | P1／未精査 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 対象件数・一意性/矛盾一覧・由来・未分類保護・切替排出/復旧計画が揃う。矛盾未発生の段階ではユーザー待ちにしない。 |
| [IR-03](../../../design/domains/owners-pets/technical-spec.md#identity-physical-adapter-candidates-20261019) 主飼主移管preview・集合変化・版/receiptの具体契約 | P1／設計未選定 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 新設予約/受付/JST境界/LINE変更の両先行順、既存条件付きcustomer付替え、過去資金不変を保持する設計/検証表ができる。独立LINE解除へ統合しない。 |
| [IR-04](../../../design/domains/owners-pets/technical-spec.md#identity-physical-adapter-candidates-20261019) 台帳一般更新・重複警告登録の版と結果照会 | P1／設計未選定 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 既決競合動作・same-tx・結果不明停止を満たす方式と各ACへの証跡計画があり、同電話候補をreceipt代用しない。 |
| [IR-05](../../../design/domains/owners-pets/technical-spec.md#identity-physical-adapter-candidates-20261019) LINE binding解除とdispatch claimの線形化点 | P1／設計未選定 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | DC-I05/06に対し両先行順・worker再起動・遅延HTTPで解除完了後の旧版新発信防止を検証可能。発信済み回収や不明自動再送を保証しない。 |
| [IR-06](../../../design/domains/owners-pets/technical-spec.md#identity-physical-adapter-candidates-20261019) 再連携後の予約選択再開adapterと独立停止の保持 | P1／設計未選定 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 予約/患者/参照履歴保持、未選択未再開、独立停止/結果不明不変、再連携のみの再開なし、確認保存だけの即時送信なしが検証表で追える。 |
| [IR-08](../../../spec/screens/38-liff-pet-health.md) LIFF限定公開と単回linkの現在適合・不明結果 | P1／未検証 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 公開allowlistを拡張せず、親finalizedを追加要件にせず、原子失敗と二重実行収束を検証できる。実機結果は実行担当の別証跡にする。 |
| [IR-09](../../../design/domains/reservations/technical-spec.md#reservation-intake-course-adapters-20261019) 予約再診回答target世代・版・既存予約導入方式 | P1／設計未選定 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 複数頭別回答・受付前訂正・日時引継ぎ・A→B→A未回答・旧原文保持・copy/clear非干渉が追える。Q-C1は将来要求時のみで現blockerにしない。 |
| [IR-10](../../../design/domains/reservations/technical-spec.md#reservation-intake-course-adapters-20261019) コース/option対応・時間snapshotを閉schemaへ追加設計 | P1／設計未選定 | コース0..1/option一意配列のclosed DTO、採用時master版/時間snapshotと0分の区別を候補化 | 既決6AC等と全consumer対応が揃い、未設定と明示0分/価格責務を混同せず、旧closed表で新選択を拒否しない。 |
| [IR-11](../../../design/domains/reservations/technical-spec.md#reservation-intake-course-adapters-20261019) 予約の選定済gate・schema・participantを現writerへ適合照合 | P1／未精査 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 文書選定済を未決へ戻さず、入口ごとに実装差/未精査/証跡/停止条件が判別可能。実DB/並行試験は別実行許可と担当へ引渡す。 |
| [IR-12](../../../design/domains/reservations/technical-spec.md#reservation-intake-course-adapters-20261019) 予約master/能力/名称履歴等の旧実データ状態 | P1／未精査 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 既決の分類/能力算法/保存方針を再質問せず、未設定とfallback・旧実効能力を区別した移行検証計画がある。 |
| [IR-13](../../../design/domains/reservations/technical-spec.md#reservation-intake-course-adapters-20261019) 受付snapshot/状態intent/別軸医師確認/browser設定の検証 | P1／未検証 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 表示設定が予約/会計writeを起動せず、poll/日付医院変更/結果不明/全件不足を誤成功にせず、各AC結果と限界を残す。 |
| [IR-14](../../../design/domains/staff-workforce/technical-spec.md#staff-bulk-physical-adapter-candidates-20261019) 複製/CSV勤務の物理仕様・整合性/receipt設計 | P1／設計未選定 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 通常単日API逐次部分commitで代用せず、予約保護/医院間勤務/現在grantを維持する設計と失敗注入/通信断の検証計画がある。 |
| [IR-15](../../../design/domains/staff-workforce/technical-spec.md#staff-bulk-physical-adapter-candidates-20261019) 職種/役職/雇用と休業期間の保存・候補接続 | P1／設計未選定 | 職種と独立する役職/雇用属性、医院別休業期間の候補列・適用入口・非自動副作用を具体化 | 追加3ACと作成/編集/複製/CSVの対応が揃い、復帰自動勤務作成/role連動/予約自動取消なし。略称を推測展開しない。 |
| [IR-16](../../../design/domains/reservations/technical-spec.md#reservation-intake-course-adapters-20261019) 個人/医院別capable移行と一院指名停止の現在適合 | P1／未精査 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 変換で既存有効範囲を無根拠拡大せず、未来/診療中保護・過去履歴保持・他院継続と現在grantを確認可能。 |
| [IR-17](../../../design/domains/owners-pets/technical-spec.md#identity-physical-adapter-candidates-20261019) 現在grant・医院割当・失効/cacheの実証 | P1／未検証 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 非所属/操作不足/失効/lookup障害/最後の管理権限/復旧logout競合で拒否と不変の証跡がある。新grant付与/自動poll追加を要求しない。 |
| [IR-18](../../../design/domains/clinic-settings/technical-spec.md#clinic-settings-evidence-plan-20261019) 医院・法人設定の操作別結果照合と既発行不変 | P2／未検証 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | scope訂正の再作業を数えず、全未終了業務解消という未採用条件を追加せず、D10 snapshotと各設定値が区別される。 |
| [IR-19](../../../design/domains/owners-pets/technical-spec.md#identity-physical-adapter-candidates-20261019) サイグラム派生結果保存と導入の具体設計 | P2／設計未選定 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | 9ACと固定表正本へtraceしDOB変更片側保存なし。将来ルール改定や検索/集計UIを現必須判断へ加算しない。 |
| [IR-20](../../../ops/testing/scenarios/S23-multi-clinic-scope.md) S14/S23修正済み目標の実画面受入証跡 | P2／未検証 | 具体内部候補・限定静的根拠・失敗/移行/照会/実機手順へ接続 | T1–T5/T-A1–4と検索各条件の機密除去済み証跡がある。自院利用採否やPR722文書修正を重複残件にしない。 |
| [REM-OPS-02](../../../ops/infra/production/setup.md) 本番provider構築と権限・DB実状態の確認 | P1／未検証 | 承認記録と実行receiptを分離し、対象同一性・head・互換性と停止/復旧条件を具体化。方式・責任者・保持年数・作業窓は推測補完しない。 | 専用account/zone/billing/DBrole/R2/DNS/証明書/Production reviewersとbranchrule/保護secret名/DB-backed health/実フロントtargetを承認済み担当者が確認。 |
| [REM-OPS-03](../../../ops/infra/production/runbook.md) backup取得方式・保持・restore責任の確定 | P1／設計未選定 | 承認記録と実行receiptを分離し、対象同一性・head・互換性と停止/復旧条件を具体化。方式・責任者・保持年数・作業窓は推測補完しない。 | owner、managed backupか承認pg_dump、正確なtarget、保護保存先、retention/expiry、取得timestamp/size/hash/receipt/restorelinkを選定記録。 |
| [REM-OPS-04](../../../ops/infra/production/runbook.md) 隔離restoreとCF-only rollback実測 | P1／未検証 | 承認記録と実行receiptを分離し、対象同一性・head・互換性と停止/復旧条件を具体化。方式・責任者・保持年数・作業窓は推測補完しない。 | 隔離target同一性を確認し承認済みrestore rehearsal、schema/代表値/アプリ互換/時間/cleanupとGOOD_SHA・migration互換・実runheadを確認。 |
| [REM-OPS-05](../../../delivery/GOLIVE_RUNBOOK.md) Go-live作業窓・authorityと初期準備のreceipt | P1／未検証 | 承認記録と実行receiptを分離し、対象同一性・head・互換性と停止/復旧条件を具体化。方式・責任者・保持年数・作業窓は推測補完しない。 | 当日手順前にcutover window/判断者/連絡経路、個人account、clinic scope、全院締め時間と初期master事後確認を収集。 |
| [REM-OPS-06](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md) formal producer/payment sourceとcandidate pinsの受領 | P1／未検証 | source/run/build/hash/profile/医院/target/6seedの同一candidate照合順と、同一execution IDでの結果照会・復旧手順を具体化。 | producer唯一ownerから同一source/run/build/hash全CSV/manifest/検証receiptを受領し、21baselineと22candidateのprofile/activationを分離して照合。 |
| [REM-OPS-07](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md) 移行税率/小数/NULL/薬剤profileの残契約確認 | P1／設計未選定 | source/run/build/hash/profile/医院/target/6seedの同一candidate照合順と、同一execution IDでの結果照会・復旧手順を具体化。 | MIG-34/37/41・EMR-343/345/364/365の最新承認とproducer/consumer反映を照合。未決値だけ判断者へ具体案と影響を提示。 |
| [REM-OPS-08](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md) 移行長時間/実規模・応答消失・復旧の検証 | P1／未検証 | source/run/build/hash/profile/医院/target/6seedの同一candidate照合順と、同一execution IDでの結果照会・復旧手順を具体化。 | 承認されたfresh隔離targetで実規模のrehearsal、timeout/切断/verify失敗/crash、同一executionIDでbackup/restore/画面照合。 |
| [REM-OPS-09](../../../ops/testing/scenarios/FORM-FIELD-INVENTORY.md) V inventoryの残定義ID/要実測の再構築 | P2／未精査 | run inventoryの固定列と動的定義id展開、F/Cの期待・理由・証跡を具体化し、未収録と未実行を分離。。慢性疾患5wire keyを既存FE builderから収録しV03へ接続、ゼロgap断定を訂正。 | 承認済み健診/検査/マスタの実定義idはfixture receipt受領後の展開待ち。旧26endpointと表25件の差1件は元母集団との再対応待ち。全request builder再監査や全field実行は未実施。 |
| [REM-OPS-10](../../../ops/testing/scenarios/FIELD-LEVEL-PROTOCOL.md) F0–F6/C1–C3と5flowの最新実行証跡確認 | P1／未検証 | run inventoryの固定列と動的定義id展開、F/Cの期待・理由・証跡を具体化し、未収録と未実行を分離。 | 証跡保管者へ最新runと実行refを確認しinventory×F結果を照合、不足だけ承認UATで実施。5flow/実LINE/token/DBaudit/別owner signoff/cleanupを揃える。 |
| [REM-OPS-11](../../../ops/infra/CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md) 外部連携console/送信/再送とscheduler運用資格 | P1／未検証 | deployed設定・外部console・実受信・停止のreceipt表と、schedulerのnatural cron/restart別検証を接続。対象Phaseを維持。 | 承認済みtestaccountだけでdeployedgate/console/allowedhost/送受信/duplicate/失敗監視を担当者確認。scheduler naturalcron/pause/revision/restart耐久は別releasequalificationを照合。 |
| [REM-OPS-13](../../../ops/agent-harness.md) agent runtimeの実効権限/Hookロード受入 | P1／未検証 | 起動4経路・親override・hook失敗とhost拒否・OS実効権限を分ける、非危険fixtureの受入票を具体化。 | 運用者が非危険fixtureでlauncher/root/backend/空白worktree各設定、親override、hookロード/失敗を検証しOS/runner権限を確認。 |
| [REM-OPS-14](../../../ops/testing/STG-PERFORMANCE-CHECKLIST.md) 性能SLO採用値と実負荷/回復測定 | P2／未検証 | 既存runの所在/条件照合を先にし、採用SLO・提案例・実測値と中止条件を区別する手順を追補。 | 既存run所在を確認、必要なら対象/負荷/期間/中止基準を運用判断して承認測定。3VU10分health/clinicsだけを全業務/多医院/長時間保証にしない。 |
| [REM-OPS-15](../../../architecture/erd-detail/index.md) architecture/生成ERD/toolingの意味監査残範囲 | P2／未精査 | 旧coverageを先に参照し、不足architecture主要本文・ADR005/006/索引・doc test4本の意味と検査限界を補読。36DDL/3生成器hash一致、歴史CSV761行/分類一致を静的確認。。inquiries/billing_confirmationsの現DDLと新設計の責務を分け、生成ページの手修正や実装済み認定を防ぐ追補を作成。 | 歴史測定annex be9-2a-boundary-map.md:121–516、CSV理由列9–762の各行意味、生成138テーブル全列の意味・旧DDL全文・稼働DBは未確認。歴史資料を現不具合へ再分類せず、他architectureの既存全文証跡を再利用。 |
| [REM-OPS-16](../../../ops/testing/scenarios/README.md) 残るS/V領域の現FR/ACとの項目別照合（財務・臨床担当分を除く） | P2／未精査 | V03全203行/V05全243行、S30/S31を補読。横断S/V8行、V03現目標4行、V05現目標6行へ既決の操作とFR/ACを具体接続。 | 文書上の接続を追補、実行は未実施。S34–39は見出し/操作表の限定確認で本文全体・実viewport未確認。V01臨床・V02財務は各担当の結果を再利用し、全request builder再監査や全field受入は主張しない。 |
| [SRC-R01](APPROVAL-LOCATORS-20261019.md) 現在参照される承認IDの保存原文所在・役割・後続訂正の対応補完 | P2-document-provenance／元状態：未精査 | 基準231 ID/523出現を保持。初回204に加え、親取得の原文引用で残27を連結し未連結0。助手質問11/本人回答15/本人補足1・採用内容16群、回答15文字列一致・意味矛盾0。receipt14ファイルを保持し原文時刻とQ/A・後続訂正を追記。 | この27件の所在不足は解消。全231の現採用条文と全後続訂正の網羅的意味照合は別で、SRC-R01全完了とはしない。過去業務回答を新実行承認へ転用しない。 |
| [ONT-R01](../../../spec/ontology/cross-domain-contracts.md) 既存ontologyの未追従3文書の差分・未確認節だけを正本へ意味照合 | P2／未精査 | README73行、cross-domain117行、medical-records146行の全内容を複数小出力で補完読了。既決/内部案、認可/多重度/時点/状態/表示実効果をD03/D08正本・sourceに照合、新矛盾なし。 | 物理実装/移行/実機は元から文書意味精査の対象外。既存証跡未結合を未読とせず、既確認8文書再読不要。 |

## 別枠7項目と重複の維持

| ID | 発生条件/扱い |
|:---|:---|
| SRC-R06 | ATT018の具体的な採用条文不足が疑われた場合だけ原画像確認。健診専用作図を再採用しない |
| SRC-R08 | 現在UIの検証で必要なroute/状態/viewportだけ過去合成画像と比較。旧画像を現在合格にしない |
| SRC-R09 | ZIP再配布/具体的破損疑い時だけpayload CRC再計算。完了済み一覧/PNG索引照合を再残件化しない |
| SRC-R10 | 実取得/共有時だけ対象版/ACL/読者を確認。URLを許可とみなさない |
| IR-07 | LSTEP/DM対象Phase開始時に目的/caller/DM/独立停止を対応。今期必須化しない |
| IR-D01 | 対象Phaseのenable/少数実送信/cron停止/件数証跡。実送信なし |
| IR-D02 | 対象Phaseの初診/健診followup adapter。実施確認/医師目的は既決、記録作成だけで送らない |

FIN-R21→CL-R08、REM-OPS-12→CL-R14の統合を維持。CL-R01/FIN-R12/IR-11は臨床/金融/予約・受付の責任を分け、共通gate証跡は再利用。IR-09の回答target競合とCL-R04のreader非干渉を重ねて別件数にしない。ATT001/002未発見skip、動画→議事録代替、DB backup4/recording ZIP対象外を維持。Q-C1/Q-I1は具体的条件が発生した時だけで現在の必須未決ではない。

## カバレッジ・出典の訂正と限界

入力時点の1246ファイル/691 UTF8/551 PNG/4 ZIPと、全文証跡の再利用下限221（既存202+今回前段19）、部分47、旧版追従12、証跡未結合411は**基準mainの調査結果**。旧116は過去の選択範囲で総意味確認数ではない。411を未読、470を一括再読件数とはしない。本差分後に新しい全文確認総数へ機械加算しない。

PNG551は過去合成UI証跡で新pixel目視0。ZIP4は同画像配布包の一覧/索引照合済みでpayload再計算は条件付き。外部212出現/raw149/正規化130資源は役割を分け一律再取得しない。Slack397本文/571条・Backlog103行/309条・過去承認資料の既読を維持。旧証跡の時点を現稼働に変換しない。

承認IDは[保存原文所在・質問/回答の役割索引](APPROVAL-LOCATORS-20261019.md)へ接続。助手の提案IDを本人承認へ昇格させず、今回27の未連結は解消。取得不足を未承認/要件なしと断定しない。原文本文・患者情報をこの記録へ大量転載しない。

## 責任者へ確認する事項

以下は送信済みの未回答質問一覧ではない。既存承認/保存receiptをこちらで先に照合し、資料調査・候補整理・実行証跡の所在確認を進める。具体的な不足を提示しても本人/医師/運用責任者の判断が必要なものだけを限定確認する。既決の3業務や予約の既選定技術方式を再承認待ちに戻さない。

| 責任者/出典 | こちらで先に調査・整理し、なお不足する場合の確認 |
|:---|:---|
| 医師・CL-R06、D03用量/原資料#261 | 犬猫の単純体重計算・目的経路パターン選択は本人2026-10-15 18:29:00 UTC `Sentinel_e1272a0f140c8191a2cd465e822364d4`「1 yes / 2 yes」で採用済み。こちらで原資料の値/単位/対象/版を整理し、具体量・上限等をどの医師承認へ対応するかを確認。役職は本人回答どおり「医師」 |
| 医院責任者・CL-R15、旧外注/画像/八王子の既存解析 | こちらで旧手順・既存資料・医院別利用の根拠を調査。保護VaultのSlack保存ログ `noah-カルテ開発_曽我-export.txt`（6/27 23:32 JST、青山）のDrワン/外注/院内検査運用の記載は既存事実で、個別の大阪血清/PicJointの利用・初日必須性までは証明しない。差が残る医院・業務だけ継続方法等を確認。専用新連携採用とは別 |
| 運用責任者・REM-OPS-03、production runbook | こちらでE1の既存backup/隔離復元PARTIAL PASSと方式候補・保護先/保持の差を整理し、将来の本番運用責任/採用を確認。旧DB backup4件資料調査の対象外判断を撤回しない |
| 切替判断者・REM-OPS-02/05、Go-live runbook | 同Slack保存ログの後続export（10/9 00:23:01 JST、Kentaro Kuroki、ts `1791472981.355199`）の「10/20構築開始→旧新並行→全医院準備→最終旧backup/import→新のみ」の既存方針を保持。こちらで記録/連絡先/適用receiptを照会し、現在の切替窓・Go/No-Go権限の不足だけ確認。連絡先を判断者と推測しない |
| 移行責任者・REM-OPS-07、CSV import契約 | 最新MIG/EMR承認との照合後に残った列/値/profile差をどれに揃えるか。具体差と影響を提示してから確認 |
| 外部連携/受入責任者・REM-OPS-10/11 | 最新run所在、承認testaccount/医院、実行担当・別確認者、実送信/停止/再送範囲は何か |
| 性能責任者・REM-OPS-14、性能チェックリスト | こちらで既存測定と採用基準を照会。2/15別会話の同時10・月1deployという既存想定（本人 `8b4f9a2c-bf6e-485b-9714-7342857d774c` / `da0abb9e-21a1-4b5f-8313-21219f587fcd`、親の原文引用）を保持し、現在の業務SLO採用とは分離。不足時だけ対象/負荷/期間/p95等・失敗率・回復/中止基準を確認 |

税務・適格請求書全体・決済実契約の適合はFIN-R25の確認者/根拠/判定を必要とし、仕様採否の未決へ戻さない。原文27件の個別ID連結は解消し、出典索引へ補足した。未確認の具体値・実契約・実行結果は影響する単位だけ残し、業務採否の未回答へ混ぜない。

## 文書検査の結果と対象外

文書差分の空白検査、75ID/別7・24/20/31の保持、201財務AC集合、47field一意性、31項目の追加未実行手順と既存証跡の対応、基準版231承認IDの索引一致、既存FR/AC/anchor保持を検査した。docs内ローカル参照・fragment計11,529件の欠損候補0。実行したscoped verificationは文書JSONの構文確認でPASS、製品runtime検査は対象外である。

汎用docs-governanceの全repository監査はexit 1・48指摘。基準mainと今回の出力が全文一致し、新規指摘0だった。既定の正本名、Git ref/ホーム内設定のパス扱い、docs外manual画像、ignored歴史report等の候補を含む。これを全repository PASSへ読み替えず、今回のdocs参照検査と分けて保管した。対象外ファイルや規約を検査通過だけの目的で変更していない。検査ログ・差分hash・独立レビューは作業領域の `full-audit/remaining-response-20261019/` に保管する。

公開一致・CI・PRの状態は公開後の報告を参照し、この時点の文書検査や作者の既存CI報告から推定しない。

## 後続承認2点の文書追加（10/19）

前段の75項目・別7・231 IDは上記基準mainの調査記録として保持する。後続の[質問と本人回答receipt](APPROVAL-LOCATORS-20261019.md#後続2点の本人承認基準231-idとは別の追補)に基づき、[空きケージの新規選択3AC](../../../spec/domains/hospitalization/lifecycle.md#cage-available-selection-20261019)と[飼主・ペット一覧/選択用の検索ボタン実行4AC](../../../spec/shared/ui-contracts.md#owner-pet-explicit-search-20261019)を正本へ追加した。既存の期間/医院/現在割当・任意、検索述語/filter/復帰/認可を保持する。旧実装の使用中選択/300ms自動発火との差は未検証で、実装・実機完了へ加算しない。メモの表記統一/詳細取得エラーは今回承認に含めない。

## 後続横断承認の反映（10/19 05:52 UTC）

[本人承認と共通FR/8AC](../../../spec/shared/ui-contracts.md#interaction-consistency-20261019)と[適用台帳](../../../spec/shared/interaction-scope.md)で業務一覧/飼主pet選択の発火を統一し、同種操作の意味と用途別例外を個別正本・設計・ontology・S14へ接続した。旧リアクティブ原則/予約の反映漏れを解消し、旧75/別7/231 IDは元調査基準として保持。新8ACと実装追従・全caller/実機適合は未検証。


## 原文連結・既存検証報告の追補

27対象の原文取得は親からの引用によるもので、取得元の `partial=false` は親の報告。ローカルで全会話を再取得したとはしない。回答済みの機能採否、こちらの資料/設計調査、医師の具体値・運用責任者の将来採用、実装/実機検証を区別した。今回確認できた範囲で新たな業務再質問はないが、全履歴で質問/回答が不存在との断定もしない。

旧DB backup4件の追加調査は2026-10-16 04:03:16 UTCの本人 `Sentinel_c877d7f8eeb48191b1024a7ac160f451`「DBバックアップは範囲外じゃない？」と後続整理に従い対象外のまま。将来の本番backup設計・隔離復元/切替受入が不要という意味ではない。31項目は既存の限定PASS等を[対象版付きの根拠と未検証部分](REMAINING-VERIFICATION-20261019.md#既存証跡の対象版成立範囲)へ接続し、今回新たな実行はしていない。
