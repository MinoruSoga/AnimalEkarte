# レジ締めの業務契約

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.cash-register.business-rules／業務仕様 |
| 決定状態 | 既存append-only・締め境界を継承。実入出金と請求の分離はD10承認範囲。開始残・補充／回収は承認C1。2026-10-15追加確認8の権限者・理由付き実査誤入力訂正を反映。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-15 |

## 目的・対象範囲

D12は期間ごとの現金実査、理論値との過不足、締めsnapshot・訂正差分・締め境界設定を扱う。請求額・未納・預り金残高はD10、読取レポートはD16。締め作成を会計確定・税務申告・金融機関への精算に代用しない。[既存締め仕様](../../cash-register.md)の旧集計式は現行実装説明として残し、本書のD10連携要件と区別する。

## 担当者・権限・前提

cash-register-close の view／create は別権限。同医院内の締め全体が対象で、担当者本人のみへの限定はない。確定締め原本の直接更新・削除・取消／再開は誰にも許可しない。実査誤入力は、当該訂正操作を許可された担当者が理由付きの追記で訂正する。閲覧・締め作成権限だけから訂正許可を推定しない。締め後会計訂正は accounting-post-close-edit と理由を要する。時刻設定は closing-settings の該当操作権限を要し、LINE予約の設定権限を代用しない。

## 契機・入口・出口

/accounting/close で営業日・AM／PM／EMGを選び、プレビュー→実査→過不足確認→締め保存。/accounting/close/history で当時snapshotを参照する。設定は /settings/closing-time。出力の保存はブラウザ印刷／PDFで、帳票の印刷が締め確定を発生させるものではない。

## 状態遷移・引継ぎデータ

<a id="requirement-fr-closing-period-boundary"></a>
### FR-closing-period-boundary：営業日と連続した区分

医院のJST境界によりAM=[開始,午前午後境界)、PM=[午前午後境界,終了)、EMG=[終了,翌開始)とする。午前0時からAM開始前は前営業日EMG。境界ぴったりは次区分に一度だけ含まれる。特別期間は標準AM開始を引き継ぎ境界／終了を上書きする。現行AM開始は09:00固定で画面・PATCHでは変更できない。既定値・説明例と#252の全院裁定値は別。今回医院設定を書き換えない。

確定済み締めは後日の境界変更で再計算しない。設定の休診日／特別期間は締めの責務であり、LINE予約の休診設定へ伝播しない。特別期間重複・不正な時刻順序の拒否、どの設定版を採用するかは現行検証／技術設計で明示する。未確認の優先順位を新しい業務合意にしない。

<a id="requirement-fr-closing-movement-consumer"></a>
### FR-closing-movement-consumer：実入出金を一度だけ集計

[D10の締め境界](../accounting/lifecycle.md#requirement-fr-accounting-post-close-separate-movement)を継承。会計の請求確定・請求差分と、実際の入金／返金イベントを分離する。実受領・実返金はそれぞれ実日時・方法・元取引参照を保持し、実日時に対応する営業日／区分へ計上する。元請求日を今日に書き換えない。

- 過去未納の今日回収、部分入金、預り金の実受領は今日の実入金。過去未納回収を新売上へ重複計上しない。
- 預り金の充当／解除／残高復元、取消元残入金の再配分は金銭移動0。再受領や実返金として足さない。
- 返金は実際に返した方法へ計上する。元支払いがカードでも現金返金なら当日の現金減少である。返金必要額だけでは減少しない。
- 料金訂正のみは請求差分で現金移動0。保険負担額の記録は保険会社からの実入金でない。
- 現金は釣銭後の実受領を用い、釣銭を含む預かった紙幣額で増収にしない。複合決済は各実額を一度ずつ扱い会計件数を分割数で増やさない。

イベントID／訂正参照／医院／実日時／方法／金額が不明なデータは0円やcompleted会計総額に置換しない。再読取・重複検査・差異確認の対象とする。D10新契約のconsumerは未実装／未検証であり、旧payment集計の名称変更だけで対応済みとしない。

<a id="requirement-fr-closing-cash-reconciliation"></a>
### FR-closing-cash-reconciliation：実査と過不足

既存継承：実査額は有限・0以上、金種別入力はない。差額は実査−理論。旧仕様の理論値は現金決済会計額−返金で開始残・補充・回収を持たない。この旧理論値を物理的なレジ全残高と同義にしない。

承認済みC1：開始時のレジ現金、レジへの補充／レジからの回収を理由・担当・日時付きで記録し、理論現金＝開始残＋現金実入金−現金実返金＋補充−回収。実入金には現金での預り金実受領を含む。補充／回収は売上・飼主入金でない。開始残と前区分の終了残の二重加算を防ぎ、どの残高を開始値に採用したか確認する。実査差額が存在するだけで架空の入出金・売上を作らない。

<a id="requirement-fr-closing-append-only"></a>
### FR-closing-append-only：原締めと訂正差分

既存append-onlyを継承。同医院・同営業日・同区分は一つの原締めを持ち、削除して再締めする経路を設けない。締め後会計訂正は理由付きで元締めへの差分を追加し、監査と差分が保存できない場合は訂正を成功扱いしない。現金移動0の料金訂正と実入出金を区別する。

実際の入出金が後日なら後日の区分へ、原請求との関連は保持する。閉じた期間について遅れて判明した訂正は原締めを保持して追跡し、当時snapshotを現在値へ置換しない。訂正イベントが元期間に属する場合の表示は元snapshot・追記・現在の再照合を区別し、二重計上を避ける。記録日時と実際の発生日時を混同しない。

<a id="requirement-fr-closing-actual-input-correction"></a>
### FR-closing-actual-input-correction：実査額の誤入力訂正

実際に数えた現金額を誤って入力した場合、元の締めを残し、権限のある担当者が訂正理由を付けて正しい実査額を追記できる。対象医院・元締めID／営業日／区分、訂正前の有効実査額、訂正後の実査額、理由・担当者・訂正日時を保持する。元の実査時点と訂正入力時点を区別し、今日の現金を数え直した値で過去の実査を置換しない。訂正後の値も既存の有限・0以上の入力条件に従う。

原snapshot・原実査額・原差額は不変で、履歴では元の値、各訂正、有効な訂正後実査額／過不足を区別する。実査誤入力の訂正は同じ照合基準の理論現金を変えず、訂正後実査−その基準の理論現金で差額を表示する。元理論10,000円・誤入力実査9,900円を10,000円へ訂正した例では、原不足100円を保持し、訂正後の差額0円を示す。100円の補充・売上・飼主入金を生成しない。

実際の現金過不足、レジ補充／回収、会計訂正や後日の実入出金とは異なる操作である。訂正によって差額が残る場合も理由だけで差額を0へ合わせず、金銭移動を作らない。元締めの取消／再開／再締めや会計ロック解除は行わない。すでに後続区分で採用した開始残snapshotを無言で書き換えない。

訂正は最新の有効実査額と版を再確認し、別担当者の先行訂正があれば差分を示して再確認する。訂正・必要監査・操作結果は同時成功とし、応答不明時は同じ操作を照合する。再送や二重クリックで訂正を重複追加しない。具体的な保存／照会・認可対応は[選定契約](../../../design/domains/finance-operations/technical-spec.md#closing-actual-input-correction-contract)へ従う。

現行差：既存公開routeはCreate／List／Get／Preview中心で、会計後訂正のappend-only差分はあるが、実査誤入力の専用訂正を実装済みとは扱わない。原closeのUPDATE禁止を解除する変更ではない。

<a id="requirement-fr-closing-projection"></a>
### FR-closing-projection：配賦・帳票・履歴

部門×支払方法の既存配賦契約は[画面29の唯一の契約本文](../../screens/29-closing-aggregation.md#closing-allocation-contract)。最大剰余法、行／列／総額の円単位保存、返金の独立負値行、会計distinct、無効／削除済み方法の過去列保持を継承する。D10の割引配賦とD12の支払配賦は異なる処理で、同じ値引きを二重控除しない。

新実入出金を取り込む時、預り金受領・レジ補充など請求明細を持たない移動を架空の診療部門へ配賦しない。売上部門集計と現金増減の差は内訳で説明する。現行マトリクスと旧KPIは返金帰属日が異なるため無条件に一致すると表示しない。D16の暦日報告と営業日締めも同じ日付ラベルだけでは突合できない。

## 異常系・保存・取消

- 二重締めは同医院／日／区分で原子的に拒否。応答消失時は履歴を読み直し、失敗と決め付け再作成しない。
- 非現金方法を現金に変更する訂正と、別方法での実返金を区別し、それぞれD10の履歴を受ける。
- プレビュー後に入出金が変わった時、古い表示を確定値と誤認しない。保存時の再集計／版確認・差額再確認は[選定技術契約](../../../design/domains/finance-operations/technical-spec.md#closing-selected-contract)に従い、不一致なら確定を止める。
- 権限拒否・取得失敗・0件・既に締め済みを区別。取得失敗を0円として締めない。
- 履歴はJST当月が既定、?date=YYYY-MM-DDは当日。区分絞込は取得ページ内という現行範囲を保つ。

## 受入条件

| 条件ID | 前提／操作（Given / When） | 期待結果（Then） | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-closing-boundaries | AM開始／境界／終了ぴったりと1秒前のイベント | 各イベントが一つの区分だけへ入る | [FR-closing-period-boundary](#requirement-fr-closing-period-boundary)。4境界と越日を照合、未実施 |
| AC-closing-midnight | 翌02時の実入金を前日営業日で集計 | 前日EMGへ一度計上 | [FR-closing-period-boundary](#requirement-fr-closing-period-boundary)。実日時と営業日を照合、未実施 |
| AC-closing-settings-history | 締め済み後に時刻設定変更 | 原snapshot不変、LINE予約設定も不変 | [FR-closing-period-boundary](#requirement-fr-closing-period-boundary)。新旧設定／履歴比較、未実施 |
| AC-closing-ar-collection | 過去請求へ今日3,000円入金 | 今日の入金3,000、過去売上の再生成なし | [FR-closing-movement-consumer](#requirement-fr-closing-movement-consumer)。請求・移動両参照、未実施 |
| AC-closing-deposit-use | 昨日預り受領5,000、今日充当5,000 | 昨日受領1回、今日の再受領0 | [FR-closing-movement-consumer](#requirement-fr-closing-movement-consumer)。受領と充当IDを照合、未実施 |
| AC-closing-deposit-restore | 取消に伴い充当解除・残高復元 | 現金入出金は0、返金は別実行 | [FR-closing-movement-consumer](#requirement-fr-closing-movement-consumer)。元充当・復元参照、未実施 |
| AC-closing-method-refund | 過去カード払いへ今日現金1,000円返金 | 今日現金−1,000、過去カード受領は保持 | [FR-closing-movement-consumer](#requirement-fr-closing-movement-consumer)。実方法／日時を照合、未実施 |
| AC-closing-no-money-adjust | 請求だけ1,000円減額 | 料金差分−1,000、現金移動0 | [FR-closing-append-only](#requirement-fr-closing-append-only)。差分と移動を照合、未実施 |
| AC-closing-event-repeat | 同じ実入金イベントを再読取 | 重複計上なし | [FR-closing-movement-consumer](#requirement-fr-closing-movement-consumer)。イベントID別件数、未実施 |
| AC-closing-invalid-actual | 実査額に負値／NaN相当を入力 | 保存拒否し原記録不変 | [FR-closing-cash-reconciliation](#requirement-fr-closing-cash-reconciliation)。入力境界、未実施 |
| AC-closing-no-reopen | 締め取消／削除／同区分再締め要求 | 原締め維持し拒否 | [FR-closing-append-only](#requirement-fr-closing-append-only)。DB／API結果を照合、未実施 |
| AC-closing-adjust-audit | 締め後訂正の監査または差分保存失敗 | 訂正全体を成功扱いしない | [FR-closing-append-only](#requirement-fr-closing-append-only)。原本と差分を照合、未実施 |
| AC-closing-actual-input-correction | 原理論10,000・実査9,900、権限者が誤入力を理由付きで10,000へ訂正 | 原値・不足100を保持し、訂正後実査10,000／差額0を別表示。現金移動・売上0 | [FR-closing-actual-input-correction](#requirement-fr-closing-actual-input-correction)。前後値と履歴照合、未実施 |
| AC-closing-actual-correction-guard | 訂正権限なし／理由空／負値・非有限値／別医院の締めを指定 | 拒否し原締め・有効実査額を変更しない | [FR-closing-actual-input-correction](#requirement-fr-closing-actual-input-correction)。認可・入力境界、未実施 |
| AC-closing-actual-correction-conflict | 同じ実査版を2人が訂正 | 先行成功のみ有効、後続は最新値と差分を再確認。上書きしない | [FR-closing-actual-input-correction](#requirement-fr-closing-actual-input-correction)。版競合、未実施 |
| AC-closing-actual-correction-retry | 訂正成功後応答不明・同一操作再送、または監査保存失敗 | 同一成功は同じ訂正へ収束。監査失敗では訂正も未保存 | [FR-closing-actual-input-correction](#requirement-fr-closing-actual-input-correction)。履歴・receipt比較、未実施 |
| AC-closing-actual-correction-separation | 訂正後も実査不足が残り、後続区分が開始残を保存済み | 不足を保持し架空補充／入金なし。元締め再開・再締め・後続snapshotの自動変更なし | [FR-closing-actual-input-correction](#requirement-fr-closing-actual-input-correction)。会計／現金／後続履歴比較、未実施 |
| AC-closing-actual-correction-history | 同じ原締めの実査を2回訂正 | 各前後値・理由・担当・時刻を保持し、最新有効実査を採用。訂正後絶対値を加算しない | [FR-closing-actual-input-correction](#requirement-fr-closing-actual-input-correction)。原本／訂正／現在投影比較、未実施 |
| AC-closing-refund-allocation | 同一支払と返金が期間内に共存 | 返金独立負値のまま配賦、各合計一致 | [FR-closing-projection](#requirement-fr-closing-projection)。画面29正本の回帰条件、未実施 |
| AC-closing-zero-v-error | 集計取得失敗後に締め操作 | 取得失敗を0円とみなして確定しない | 異常系。失敗／0件を比較、未実施 |
| AC-closing-opening-cash | C1の目標で、開始10,000・入金3,000・返金1,000・補充2,000・回収4,000 | 理論10,000、実査9,900なら不足100 | 承認C1。各移動と差額を照合、未実施 |
| AC-closing-no-fake-revenue | C1の目標で、レジ補充／回収・過不足を登録 | 診療売上・飼主入金を捏造しない | 承認C1。売上／現金内訳を照合、未実施 |

## 未決事項・関連仕様・根拠

C1（開始残高・現金実入出金・補充／回収による理論現金と実査比較、預り金充当を再入金にしない）は2026-10-15に依頼者承認済み。質問 Sentinel_75a3a1bf81b481919ba4c34aa633feb4 の「3. 在庫とレジ締め」、未回答群の確認 Sentinel_8d0032ca45b881919b3d2ceb5339f792 に対する Sentinel_7a67fd30e47c819192cd36d416f0a2bf「ok」が根拠。cash-register.md・画面29・締め時間設定の既存契約、D10の承認済み実移動境界を根拠にする。締めと月次の既存配賦は無条件に新台帳へ互換とせず、[技術整理](../../../design/domains/finance-operations/technical-spec.md#closing-consumer-gap)でadapterと検査を定義する。月次・税務適合、実決済との照合を完了扱いしない。

実査誤入力の訂正は2026-10-15追加確認8（Sentinel_a408975ce6408191be7a25734aae0aab）へのユーザー「ok」（Sentinel_4f050b1c0ed08191b5ca67deeaeec843）に基づく。承認範囲は元締めを保持する権限者・理由付き訂正で、取消／再開や現金移動の自動生成は含まない。
