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

2026-10-15追加承認：締め済み保護の対象は実際に締めた区分だけとする。午前を締めても、未締めの午後・夜間を締め済み扱いにしない。午前09:00〜12:00・午後12:00〜18:30・夜間18:30〜翌09:00という確認済み区分と、当該区分で採用した境界に従って判定する。時刻設定を今回変更する操作ではない。会計自体の確定後訂正条件は区分が未締めでも維持する。旧実装の日全体保護をこの目標の例外にしない。

確定済み締めは後日の境界変更で再計算しない。設定の休診日／特別期間は締めの責務であり、LINE予約の休診設定へ伝播しない。特別期間重複・不正な時刻順序の拒否、どの設定版を採用するかは現行検証／技術設計で明示する。未確認の優先順位を新しい業務合意にしない。

<a id="requirement-fr-closing-movement-consumer"></a>
### FR-closing-movement-consumer：実入出金を一度だけ集計

[D10の締め境界](../accounting/lifecycle.md#requirement-fr-accounting-post-close-separate-movement)を継承。会計の請求確定・請求差分と、実際の入金／返金イベントを分離する。実受領・実返金はそれぞれ実日時・方法・元取引参照を保持し、実日時に対応する営業日／区分へ計上する。元請求日を今日に書き換えない。

- 過去未納の今日回収、部分入金、預り金の実受領は今日の実入金。過去未納回収を新売上へ重複計上しない。
- 預り金の充当／解除／残高復元、取消元残入金の再配分は金銭移動0。再受領や実返金として足さない。
- 返金は実際に返した方法へ計上する。元支払いがカードでも現金返金なら当日の現金減少である。返金必要額だけでは減少しない。
- 料金訂正のみは請求差分で現金移動0。保険負担額の記録は保険会社からの実入金でない。
- 現金は釣銭後の実受領を用い、釣銭を含む預かった紙幣額で増収にしない。[No.30の手動釣銭修正](../accounting/lifecycle.md#accounting-manual-change-reconciliation-20261017)でも、未確認の過受領を残して完了せず、金額だけを自動補正して売上・入出金を作らない。[過受領の訂正・実返金・了承付き預り金登録](../accounting/lifecycle.md#accounting-excess-receipt-resolution-20261017)において、同じ現金移動は一度だけ計上し、預り金への対応づけや充当を架空の追加入金にしない。複合決済は各実額を一度ずつ扱い会計件数を分割数で増やさない。

イベントID／訂正参照／医院／実日時／方法／金額が不明なデータは0円やcompleted会計総額に置換しない。再読取・重複検査・差異確認の対象とする。D10新契約のconsumerは未実装／未検証であり、旧payment集計の名称変更だけで対応済みとしない。

<a id="requirement-fr-closing-cash-reconciliation"></a>
### FR-closing-cash-reconciliation：実査と過不足

既存継承：実査額は有限・0以上、金種別入力はない。差額は実査−理論。旧仕様の理論値は現金決済会計額−返金で開始残・補充・回収を持たない。この旧理論値を物理的なレジ全残高と同義にしない。

承認済みC1：開始時のレジ現金、レジへの補充／レジからの回収を理由・担当・日時付きで記録し、理論現金＝開始残＋現金実入金−現金実返金＋補充−回収。実入金には現金での預り金実受領を含む。補充／回収は売上・飼主入金でない。開始残と前区分の終了残の二重加算を防ぎ、どの残高を開始値に採用したか確認する。実査差額が存在するだけで架空の入出金・売上を作らない。2026-10-15追加承認：現金の差額が残っていても、その差額を記録して締められる。差額が0でないときはメモを必須とし、「原因不明」「確認中」も受け付ける。原因の解明や架空の補充／回収での差額解消を締めの前提にしない。差額0の場合に新たなメモ必須条件は追加しない。

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

締め履歴は年・月・区分（AM／PM／EMG）を指定して確認できる。既存の区分フィルタは取得ページ内という制限を保持し、全期間の横断検索と表示しない。実現金・差額はレジ締め画面・締め履歴・日次締め帳票で確認し、月次集計レポートには記載しない。[Backlog Phase1 No.30 F31:G31](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F31:G31)の回答が根拠。実査訂正後もこの出力境界を維持し、月次売上へ過不足を加算しない。

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
| AC-closing-history-filters | 年・月とAM／PM／EMGを指定して締め履歴を開く | 指定年月・区分の対象を確認でき、区分は取得ページ内の絞込みとして区別。閲覧認可・元締め保護を維持 | FR-closing-projection、Phase1 No.30。未実施 |
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
| AC-closing-closed-segment-only | 午前だけ締め済みで同営業日の午後・夜間を扱う | 未締め区分を締め済み扱いにしない。午前原締め保護は維持 | [FR-closing-period-boundary](#requirement-fr-closing-period-boundary)。追加承認目標、未実施 |
| AC-closing-closed-segment-finalized-invoice | 未締め区分に確定済み会計がある | 区分が未締めでも会計自体の確定後訂正条件を維持する | [FR-closing-period-boundary](#requirement-fr-closing-period-boundary)。追加承認目標、未実施 |
| AC-closing-cash-difference-note | 理論10,000円・実査9,900円で「原因不明・確認中」を記載して締める | 不足100円とメモを保持して締められる。架空入出金は生成しない | [FR-closing-cash-reconciliation](#requirement-fr-closing-cash-reconciliation)。追加承認目標、未実施 |
| AC-closing-cash-difference-note-empty | 差額が0でなくメモが空／空白だけ | 保存を拒否しメモ入力を求める。原因判明までは要求しない | [FR-closing-cash-reconciliation](#requirement-fr-closing-cash-reconciliation)。追加承認目標、未実施 |
| AC-closing-cash-difference-zero-note | 差額0でメモなし | 他の締め条件を満たせば、差額メモ不足を理由に拒否しない | [FR-closing-cash-reconciliation](#requirement-fr-closing-cash-reconciliation)。追加承認目標、未実施 |
| AC-closing-cash-difference-repreview | プレビュー後の入出金更新で差額が変化する | 再集計と差額再確認を行い、保存時の非0差額にメモを必須とする | [FR-closing-cash-reconciliation](#requirement-fr-closing-cash-reconciliation)。追加承認目標、未実施 |

## 未決事項・関連仕様・根拠

C1（開始残高・現金実入出金・補充／回収による理論現金と実査比較、預り金充当を再入金にしない）は2026-10-15に依頼者承認済み。質問 Sentinel_75a3a1bf81b481919ba4c34aa633feb4 の「3. 在庫とレジ締め」、未回答群の確認 Sentinel_8d0032ca45b881919b3d2ceb5339f792 に対する Sentinel_7a67fd30e47c819192cd36d416f0a2bf「ok」が根拠。cash-register.md・画面29・締め時間設定の既存契約、D10の承認済み実移動境界を根拠にする。締めと月次の既存配賦は無条件に新台帳へ互換とせず、[技術整理](../../../design/domains/finance-operations/technical-spec.md#closing-consumer-gap)でadapterと検査を定義する。月次・税務適合、実決済との照合を完了扱いしない。

実査誤入力の訂正は2026-10-15追加確認8（Sentinel_a408975ce6408191be7a25734aae0aab）へのユーザー「ok」（Sentinel_4f050b1c0ed08191b5ca67deeaeec843）に基づく。承認範囲は元締めを保持する権限者・理由付き訂正で、取消／再開や現金移動の自動生成は含まない。

<a id="closing-segment-difference-agreement-20261015"></a>
### 追加合意の根拠

2026-10-15、質問 Sentinel_c1016545a11c8191b7149324731079ff の締めた区分だけの保護・非0現金差額のメモ必須に対し、依頼者 Sentinel_f90aa49cf1c0819185758f699732539b が「1 yes / 2 yes」と承認。原因不明／確認中でも締められる。既存の確定済み会計訂正、原締め不変、権限者・理由付き実査誤入力訂正、実入出金分離、保存時再集計は維持する。

## 日次集計の追加内訳

[原資料採用正本・追加2AC](../../screens/29-closing-aggregation.md#closing-source-breakdown)へ従い、往診／トレセンを独立表示し、病院／トリミング別の保存済み税・値引きを表示する。分類・税・値引き・支払実額は別概念で、表示配賦で元会計総額を変えない。金額配賦の契約本文は画面29の唯一正本へ維持し、旧分類数を目標の固定数にしない。

## Backlog原条文の補足受入条件

保存済み原Q/Aと既存業務契約の照合による補足。対応済みという旧回答は実装・実機検証の合格証明に用いない。後続の本人合意、医院scope、権限、履歴・会計訂正の既存条件を維持する。

### FR-closing-projection／日次と月次の印刷

| 条件ID | 受入条件 | 根拠 | 検証状態 |
|:---|:---|:---|:---|
| AC-closing-report-a4-landscape | 日次締めプレビューと月次レポートに既存部門明細がある／それぞれ印刷・PDF保存する／A4横で必要な全部門が欠落・横切れせず出力され、表示と同じ集計対象・円単位合計を使う。 | Phase1 F31:G31 | 未実施 |


<a id="closing-daily-discharge-information-20261017"></a>
## 日次集計・レジ締め帳票への当日の退院情報掲載（Backlog No.4対象確認）

日次集計・レジ締めの帳票へ、その対象日に退院した情報を掲載し、帳票とともに印刷できるようにする。Backlog No.4の退院一覧はこの帳票への掲載を指すものとして整理し、入院管理ページの独立した退院一覧印刷機能を追加する承認とはしない。

[Backlog Phase1 No.4 F5:G5](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F5:G5)の原文は「退院も一覧に載せるほしい」「その一覧は印刷できるようにしてほしい」。同じ質問は現金／クレジット・部門別の締め集計とレジ現金突合の文脈に置かれ、回答にも「退院一覧・印刷：対応済み」がある。対象を区別した質問Sentinel_fceda778e8a081918bd49d1d9b0fcd07のB「日次集計・レジ締めの帳票に、その日の退院情報を掲載」に対し、本人Sentinel_5fcc573467388191a06621dd871dcffeが2026-10-17 22:02:55 UTC「B のことじゃね？」と判断したことを、今回の対象特定の根拠にする。旧回答の対応済み表記を、現在の実装・実機検証の証拠にはしない。

退院情報は[後続のJST実退院日承認](#closing-discharge-actual-jst-day-20261017)に従い実際の退院日時の日本時間暦日で集め、会計の支払日・精算日へ置き換えない。退院したという理由だけで未精算分を売上・入金へ算入せず、別日に支払われた会計を退院日に再計上しない。既存の部門別・支払方法別集計、金銭発生日・精算日等の帰属、返金・未納・預り金、円単位の配賦・総額契約は変更しない。退院情報の掲載と金額集計を区別し、掲載を会計作成／入金／締め確定の副作用にしない。

既存医院scope・帳票参照権限・患者情報保護を維持する。金銭集計の営業日／締め区分の既存契約は変えず、退院情報には後続承認のJST暦日を独立して適用する。対象日の画面表示・帳票上の具体配置は設計であり、退院日をレジ締め期間へそろえる変更はしない。

退院情報の掲載項目は[22:15:33 UTCの4項目承認](#closing-discharge-four-fields-20261017)に従う。並び・具体配置、退院情報の収集／表示単位は設計事項とし、担当医・金額等の追加掲載やホテル退室への対象拡張をこの承認から採用しない。退院情報に当該帳票日の支払があることを掲載の必須条件にしない。既存の会計表の固定列を勝手に差し替えず、具体配置はこの目標と整合させて設計する。

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-closing-daily-discharge-information | 帳票対象日に退院した情報がある日次集計・レジ締め帳票を表示する | 対象日の退院情報を帳票へ掲載でき、入院管理独立印刷へ要件を置き換えない。日付は後続承認の実退院日時JST暦日に従い、掲載項目は後続承認の退院日時・飼主名・ペット名・入院日に従い、不明値を推測補完しない | 未実施 |
| AC-closing-discharge-payment-date-separation | 退院日と支払日が異なる例、未精算で退院した例を帳票と金額集計で比較する | 退院情報は退院の事実を掲載し、支払日へ置換しない。掲載で売上・入金の追加／二重計上をせず、既存金銭集計契約を維持する | 未実施 |
| AC-closing-discharge-information-print | 対象日の退院情報を含む日次集計・レジ締め帳票を印刷する | 掲載した退院情報を帳票とともに出力でき、医院・参照権限を維持する。印刷で退院・会計・入金・締め確定を変更しない | 未実施 |


<a id="closing-discharge-actual-jst-day-20261017"></a>
### 退院情報の対象日は実退院日時のJST暦日

本人Sentinel_83fab0e44c748191914ba5941ba22e8eの2026-10-17 22:09:11 UTC「退院一覧は『実際の退院日（日本時間）』で集める形で」を根拠に、退院対象日Dは日本時間の当日0時以上・翌日0時未満、すなわち実退院日時が[D 00:00 JST, D+1 00:00 JST)に入る情報とする。当日0時ちょうどを含み、翌日0時ちょうどは翌日へ掲載する。保存時刻の表現が異なる場合も、対象判定は日本時間の実退院日時に照合し、予定退院日・入力日・支払日で代用しない。

退院情報の抽出は、支払日・精算日・レジ締め対象期間から独立する。AM／PM／EMGや営業日の金銭集計期間が暦日と異なっても、その金銭期間を退院情報の抽出条件に流用しない。例えば午前0時以降で金銭上は前営業日EMGに属する日時の退院でも、退院情報は実退院日時のJST暦日Dへ掲載する。前日支払・後日支払・未精算によって退院対象日を移動させず、退院情報掲載による売上・入金の追加／二重計上を行わない。

22:09:11 UTCの回答自体は日付基準のみであり、掲載項目はその後の[22:15:33 UTC承認](#closing-discharge-four-fields-20261017)で別に確定した。配置・行単位、実退院日時の欠損／訂正の具体処理、画面と帳票の対象日表示・取得／保存方法は別の設計・確認事項。欠損を支払日等から推測補完せず、日次集計・レジ締め帳票への掲載という対象と、入院管理独立印刷を追加しない既決は維持する。

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-closing-discharge-actual-jst-day-interval | 退院対象日Dの直前、Dの0時ちょうど、Dの末尾、翌日0時ちょうどに実退院した例を確認する | 実退院日時をJSTで照合し、Dの0時以上・翌日0時未満だけをDへ掲載する。翌日0時は翌日、予定日・入力日・支払日では代用しない | 未実施 |
| AC-closing-discharge-calendar-day-not-closing-period | 日付Dの0時以降に退院し、金銭の締め区分上は前営業日EMGに該当する例と、支払日が別日の例を確認する | 退院情報は実退院日時の暦日Dへ掲載し、締め期間や支払日へ移動しない。金銭帰属は既存契約のままで、退院掲載による売上・入金の二重計上なし | 未実施 |


<a id="closing-discharge-four-fields-20261017"></a>
### 退院一覧の掲載4項目

質問Sentinel_d78002cbb96c8191b029b3d431530dc3「掲載項目は、先ほどの『退院日時・飼主名・ペット名・入院日』の4つでよいですか？」への本人Sentinel_12d62cbb3e9c8191ba5300b05f4ff19eの2026-10-17 22:15:33 UTC「ok」を根拠に、日次集計・レジ締め帳票内の退院一覧には、退院日時・飼主名・ペット名・入院日の4項目を掲載し印刷する。

退院日時は上記の実退院日時のJST基準と整合させ、支払／精算日時で代用しない。入院日は該当する既存の入院記録を参照し、退院日や支払日を転記して埋めない。具体的な参照項目・表示書式や欠損／訂正時の処理は設計で照合し、不明値を推測生成しない。今回の項目承認で担当医・金額・請求額・入金額等の掲載を追加しない。

実退院日時のJST当日0時以上・翌日0時未満という抽出基準、支払日／締め期間との独立、金銭の既存帰属と二重計上防止、医院scope・参照権限、帳票への掲載という既決を維持する。4項目の掲載・印刷は入院状態・会計・締めの変更を兼ねない。具体的な並び順・紙面配置・行単位は未確定のままで、入院管理独立印刷は追加しない。

| 条件ID | 前提／操作 | 期待結果 | 検証状態 |
|:---|:---|:---|:---|
| AC-closing-discharge-four-fields | 実退院日JST基準で抽出した情報を日次集計・レジ締め帳票の退院一覧に表示し印刷する | 退院日時・飼主名・ペット名・入院日の4項目を確認できる。退院／入院情報を支払日で代用せず、担当医や金額項目を追加承認したとはしない。掲載で売上・入金を二重計上しない | 未実施 |
