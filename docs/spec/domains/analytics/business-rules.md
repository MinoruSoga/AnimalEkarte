# 経営・顧客集計の業務契約

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.analytics.business-rules／業務仕様 |
| 決定状態 | 2026-10-20追加本人回答は末尾の該当節を参照。2026-10-15の承認6により来院回数は実来院日基準へ変更。旧記録日数との差とD10請求／実移動分離を保持。外部到達・来院因果の保証なし。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-20（本差分。従来：2026-10-15） |

## 目的・対象範囲

D16は月次／任意期間レポート、飼主集計、Lステップ分析の読取・出力を扱う。請求／入金・預り金残高の更新はD10、締め確定はD12、LINE配信・タグ変更・CSV取込はD15、臨床記録はD03等。集計閲覧やCSV出力を自動配信の承認にしない。税務申告や外部会計ソフトへの適合を保証しない。

### 今期の提供範囲

[Backlog Phase1 No.25・27 F26:G26／F28:G28](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F26:G28)の2026-10-15曽我回答により、年間売上・来院回数ランキングと犬種別集計の画面表示／CSV出力、集計画面のCPMセグメント別人数／一覧表示は今期対象外。以前の「対応済」やサーバエラー修正待ちという記述を今期提供契約へ戻さない。月次売上レポートを今期の集計画面の中心とする（No.28 F29:G29）。以下の対象外機能の計数・母集団・CSVの記述とACは将来再開時の保持条件・既存説明として残し、今期の表示／出力提供や新しい実装完了を意味しない。実来院日を数える後続承認は指標の意味を定めたもので、対象外画面の公開承認ではない。

## 担当者・権限・前提

/accounting/reports は accounting-reports:view、/aggregation は owners:view、/lstep/analytics は lstep-analytics:view。医院scopeを維持する。締め履歴ドリルダウンには cash-register-close:view を追加確認。税率設定リンクは対象設定の閲覧権限が必要。同じ分析画面のCSV取込には独立した lstep-csv-import 権限が必要で、分析閲覧から許可を派生しない。

## 契機・入口・出口

[月次画面](../../screens/32-accounting-reports.md)、[顧客集計画面](../../screens/36-aggregation-dashboard.md)、[Lステップ分析入口](../../screens/31-lstep-integration.md)が入口。出力時は対象医院・期間・金額基準・取得範囲を保持する。飼主詳細へのリンクは対象飼主を参照し、ページ集計から新しい患者・請求を作らない。

## 状態遷移・引継ぎデータ

<a id="requirement-fr-analytics-period-basis"></a>
### FR-analytics-period-basis：指標と期間を対にする

以下のcompleted_at／completedは旧consumer SQLの観察であり、001末尾DROP後の現存DDLではない。[請求日と最終保存正本の対応](../accounting/lifecycle.md#accounting-bill-date-agreement-20261020)を参照し、旧列復活や新暦日adapterの実装済みを前提にしない。

月次レポートはJST当月を既定とし、月または開始／終了日を選ぶ。現行日次明細はcompleted_atのJST暦日で、D12のAM開始を日界にした営業日ではない。深夜の会計は両者で別の日付行へ入る。日付ラベルだけで締めと月次の一致を要求しない。

顧客売上タブの既定は当暦年、明示期間があればそれを使う。annual_visit_countはrolling365日で暦年売上と別。現行SQLでは金額・会計件数の期間帰属がカルテ日（なければ会計予定日）であり、旧文書の『その年に完了した会計』をcompleted_at期間条件と誤読しない。completed状態フィルタと日付帰属の列は別である。D10の新請求・実入金に対応する読取adapterは未検証で、現行SQLを新台帳の正本としない。

<a id="requirement-fr-analytics-money-separation"></a>
### FR-analytics-money-separation：売上・回収・返金・預り金

D10の既決区分を継承。請求確定額、今回の実入金、実返金、未納、預り金受領／充当／復元を区別する。未納回収・預り金受領は新しい診療売上ではなく、充当／復元／配分変更は新しい実入金でない。請求取消／減額だけで実返金を作らない。保険負担額は保険会社から受け取った額でない。

現行顧客集計には売上総額／入金額／返金控除後の切替があるが、選択期間の会計に紐付いた入金・返金を集計する旧SQLであり、その期間に現金が動いた額とは限らない。累計欄と選択期間欄も区別する。D12の実入出金イベントと突合するときは実日時基準を別に表示・照合し、旧指標へ無言で混入しない。新台帳への切替時のラベル・互換表示は設計事項。

<a id="requirement-fr-analytics-matrix-report"></a>
### FR-analytics-matrix-report：同一条件での表示・出力

月次レポートの画面・印刷／PDF・CSVにはレジの実現金・現金差額を記載しない。これらはレジ締め画面・締め履歴・日次締め帳票で参照する。[Backlog Phase1 No.30 F31:G31](https://docs.google.com/spreadsheets/d/17jaFBv7NmbPzk88lZx94vFxa_G1shGNX2O_lFGt4QdQ/edit#gid=0&range=F31:G31)が根拠で、実査誤入力訂正や非0差額メモの後続合意でもこの出力範囲は変更しない。

月次の日次明細・部門×支払方法表・KPI／内訳、CSV／印刷は同じ選択条件を使う。C7-7の既決に従い、午前／午後の件数・売上4列は廃止対象で、会計件数・日計へ集約しAM／PM締め状態は保持する。EMG集計を含めた日次件数を『午前＋午後』という表示文言だけで誤定義しない。現行合算のEMG網羅性は設計検査へ。

部門×支払方法の唯一の配賦正本は[画面29](../../screens/29-closing-aggregation.md#closing-allocation-contract)。返金は発生日負値、旧consumerの支払帰属はcompleted_at基準、件数は会計distinct、行件数の単純合計を総会計件数にしない。旧KPIの返金は親会計完了日帰属で、差が起こり得る。この差を表示上の丸め誤差として消さない。D10移行後の期間基準は新イベント／請求の意味を維持する。

<a id="requirement-fr-analytics-owner-counts"></a>
### FR-analytics-owner-counts：飼主と来院日・母集団

承認済み目標：来院回数は、当該医院へ実際に来院した日を基準に数える。同医院・同飼主・同日を1回とし、同日の複数ペットでも飼主の回数を増やさない。ペット単位では同医院・同ペット・同日を1回ずつ数える。同日の再受付・複数予約・複数カルテ・複数会計も、同じ単位の日数を増やす根拠にしない。期間内・累計・年間の各来院回数は同じ実来院sourceを使い、選択期間とrolling365日の違いは維持する。

通常来院の根拠は[D01の確定した受付結果と履歴](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-atomicity)を照合する。予約の存在・予定時刻到達・カルテdraft／確定・会計の存在だけから来院を推定しない。実際に来院して受付が成立している場合、カルテがdraftでもその理由だけで来院を除外しない。診療終了／会計完了を新しい計数必要条件にしない。[明示診療開始・終了](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-manual-consultation-state)も受付とは別の状態であり、状態文字列だけを来院履歴の代用にしない。

[D03の過去入力](../medical-records/lifecycle.md#requirement-fr-medical-records-retrospective-no-card-no-billing)では、既存の確認・由来から当該医院への実来院まで確認できる場合だけ実来院日へ帰属させ、登録日を新たな来院日にしない。医師確認済みというだけで外部資料・転記・日付のある臨床記録一般を来院へ変換しない。集計のための追加医師承認を全来院に要求する変更でもない。[患者転記](../medical-records/lifecycle.md#requirement-fr-medical-records-wrong-patient-linked-transfer)は新たな来院ではない。日付／患者訂正は元sourceと訂正関係を照合し、元・先を二重加算しない。

実来院の根拠・日付・医院／患者の対応が不明なら、実来院回数を確定したり「来院なし」「0回」で補完したりしない。旧期間は照合状況と旧基準を明示し、確認できる新基準の結果と混在させない。飼主変更時の帰属はD04の対象確認と履歴境界へ従い、現在のpet.owner_idだけで過去の来院を自動付け替える新方針は追加しない。

現行差：既存SQLは同医院・現在のpet.owner_idに結び付く未削除カルテのDISTINCT mr.dateで、finalized限定条件がなく、正確には「記録のある日数」。同日複数ペットを1記録日とする点は共通でも、新しい実来院回数への適合済みとはしない。旧値を表示・出力する間は「旧記録日数」等で基準を区別する。具体的source照合・再受付時刻上書きの制約は[選定read契約](../../../design/domains/finance-operations/technical-spec.md#analytics-actual-visit-source)を参照する。

実来院source未照合の飼主を無断で母集団から除外せず、必要な分類が確定できない応答は不完全と示す。以下の7区分合計の保証は分類に必要なsourceが揃った応答に適用し、不明を未分類や0件へ機械的に振り分けて完全な集計に見せない。

CPM人数は一般条件適用後、CPM区分フィルタとページ分割前の母集団を使う。一覧totalはCPM絞込み後。全7区分の合計と母集団が一致し、0件でも7キーを0で返す。売上タブの0円除外を来院タブに無条件適用せず、最終来院軸の『来院なし』条件を他軸へ無断適用しない。

<a id="requirement-fr-analytics-segmentation"></a>
### FR-analytics-segmentation：CPM・最終来院・配信設定

顧客集計のCPMはV1の6区分＋未分類を用いる。配信設定がV2でも分析が自動的にV2へ切り替わる現行契約ではない。同期側6区分と分析の残余区分を混同しない。V1閾値設定と固定の最終来院バケットは別軸。

同じ実来院指標から導く最終来院日は最後の確認済み実来院日とし、不明データを来院なしへ落とさない。最終来院の境界は90日未満、90〜179日、180〜364日、365日以上、来院なし。『3ヶ月』等のラベルを暦月差で再実装しない。タイムゾーン・日付差の再現条件を検証する。CPMや売上ランキングは配信同意／送信可能性を示さない。

<a id="requirement-fr-analytics-lstep-observation"></a>
### FR-analytics-lstep-observation：配信統計と来院観測

現行Lステップ月次統計はscheduled_atのJST月でトリガー種別×状態のログ件数を集計。来院率はfired_atが当月・fired状態・優先度抑止でないログを母数とする。UIは配信後30日、APIは1〜365日。現在の来院判定はfired_atのJST日から指定日後まで両端を含むカルテ日が同医院・飼主にあること。これは旧記録日ベースの現行差。目標の来院観測は同じ窓内に確認済み実来院日があることを上記sourceで照合し、期間・母数・窓の境界は変更しない。根拠不明時は観測不明とし、未来院や率0を確定しない。分析の変更だけで配信トリガーや送信許可を追加しない。

単位は配信ログで、ユニーク飼主数ではない。同じ飼主への複数ログが各々母数になり、一つの来院が複数ログに対応し得る。firedは外部サービスへの最終到達／既読の証明でなく、来院相関は配信による因果効果を証明しない。観測窓未経過の当月値は後日変わり得る。母数0の率0と、通信失敗で値不明を区別する。

## 異常系・保存・取消

読取失敗時は旧データ／0件／取得中と区別する。対象期間の不正入力はAPIエラーのまま扱い、無言の別期間集計をしない。検索条件変更後に古い応答が新条件の値として出ないことを検査する。

顧客CSVは現在のサーバーページで選択した行のみ。未選択時は出力不可で、全検索結果を出力したと表示しない。月次CSVの対象は選択した集計期間であり同じ出力契約ではない。友だち属性CSV取込・外部タグ同期はD15の別操作・別権限・保存結果を参照する。分析画面の表示成功を取込成功にしない。

## 受入条件

| 条件ID | 前提／操作（Given / When） | 期待結果（Then） | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-analytics-current-scope | 今期の集計画面・出力対象を確認する | 月次売上レポートと対象外のランキング／犬種別画面・CSV、CPM人数／一覧を区別する。旧対応済報告から公開済みと判定しない | 上記「今期の提供範囲」、No.25・27・28。文書／画面照合、未実施 |
| AC-analytics-midnight-basis | 翌02時の会計を月次と営業日締めで比較 | JST暦日と前日EMGの差を説明できる | [FR-analytics-period-basis](#requirement-fr-analytics-period-basis)。両基準を照合、未実施 |
| AC-analytics-year-v-365 | 暦年とrolling365日が異なる日で集計 | 売上期間と年次来院の別定義を維持 | [FR-analytics-period-basis](#requirement-fr-analytics-period-basis)。年初境界、未実施 |
| AC-analytics-old-period-column | カルテ日と完了日が別月の既存会計 | 現行顧客SQLのカルテ日帰属と完了状態を区別 | [FR-analytics-period-basis](#requirement-fr-analytics-period-basis)。旧SQL契約比較、未実施 |
| AC-analytics-old-ar | 過去未納を今日回収 | 今日入金と過去請求、売上再生成なし | [FR-analytics-money-separation](#requirement-fr-analytics-money-separation)。D10移動と請求比較、未実施 |
| AC-analytics-deposit | 預り受領→充当→取消復元 | 売上／実入金を各段階で重複加算しない | [FR-analytics-money-separation](#requirement-fr-analytics-money-separation)。元資金参照、未実施 |
| AC-analytics-refund-dates | 前月会計へ当月返金 | 発生日matrixと親会計帰属KPIの差を保持 | [FR-analytics-matrix-report](#requirement-fr-analytics-matrix-report)。正本の例を比較、未実施 |
| AC-analytics-monthly-no-cash-difference | 現金差額を含む締め・実査訂正がある月を画面／PDF／CSVへ出力する | 月次には実現金・差額を載せず、売上を差額で補正しない。締め／履歴／日次締め帳票では区別して確認できる | FR-analytics-matrix-report、Phase1 No.30。未実施 |
| AC-analytics-export-consistency | 同一医院／期間で画面・CSV・印刷 | 同じ集計対象・条件と円単位合計 | [FR-analytics-matrix-report](#requirement-fr-analytics-matrix-report)。出力3種比較、未実施 |
| AC-analytics-distinct-billing | 複合決済と複数部門を持つ会計 | 総件数をsplit数・部門行件数の合計にしない | [FR-analytics-matrix-report](#requirement-fr-analytics-matrix-report)。元ID比較、未実施 |
| AC-analytics-same-day-pets | 同飼主の2ペットが同日に実来院し、各々複数予約／カルテがある | 飼主1回、各ペット1回。予約・記録・会計の数で増やさない | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。医院／飼主／日と医院／ペット／日を比較、未実施 |
| AC-analytics-record-days-not-visits | 未削除のdraftカルテと確定カルテが異なる日に存在し、実来院根拠は未照合 | 旧SQLは2記録日。新指標の実来院2回にも0回にも確定せず旧基準／未照合を明示 | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。SQL母集団とラベルの意味を照合、未実施 |
| AC-analytics-actual-source | 未来の予約だけ、実来院根拠のないdraft／確定記録だけ、会計だけを比較 | いずれも単独では実来院にしない。実来院受付済み・draftの患者はdraftだけを理由に除外しない | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。D01受付結果と照合、未実施 |
| AC-analytics-visit-recheckin | 同一ペットの同日再受付・複数診療記録と、別日の実来院を集計 | 同日は1回、別日の確認済み実来院は別1回。再受付時刻の上書きで元日の事実を失わせない | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。source ID／履歴／日付照合、未実施 |
| AC-analytics-retrospective-visit | 当該医院への過去実来院が確認できる専用入力と、外部資料転記を比較 | 前者は実来院日に1回、登録日加算なし。後者は医師確認や日付だけで実来院にしない | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。D03の確認済み由来と照合、未実施 |
| AC-analytics-visit-unknown | 旧記録に実来院日／患者対応根拠の欠落、またはsource取得失敗 | 旧記録日数／未照合／取得失敗を示し、0回・来院なし・確定済み実来院値に補完しない | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。不明fixture、未実施 |
| AC-analytics-visit-consumer-basis | 同じ医院・期間・実来院sourceを画面／CSV／CPM入力／来院観測で参照 | 回数の基準・集計単位を揃え、旧値と区別。閾値・窓・母数・通知許可を変えない | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)・[観測](#requirement-fr-analytics-lstep-observation)。consumer照合、未実施 |
| AC-analytics-population | CPM一つを選択しページを移動 | 人数サマリーは区分絞込前母集団、一覧totalは後 | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。全件／ページ比較、未実施 |
| AC-analytics-empty-counts | 母集団0件を取得 | 7区分0、通信失敗とは別 | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。0件レスポンス、未実施 |
| AC-analytics-cpm-versions | 配信設定をV2にした状態で分析閲覧 | 顧客分析V1と配信V2を別として表示 | [FR-analytics-segmentation](#requirement-fr-analytics-segmentation)。版と分類照合、未実施 |
| AC-analytics-last-visit-boundary | 89／90／179／180／364／365日・来院なし | 固定境界の各区分に一意に分類 | [FR-analytics-segmentation](#requirement-fr-analytics-segmentation)。境界fixture、未実施 |
| AC-analytics-selected-export | 現在ページで2行選択しCSV | 選択2行のみ、未選択は出力不可 | 異常系。別ページと比較、未実施 |
| AC-analytics-permission | 分析閲覧のみでCSV取込やタグ変更を要求 | 独立操作権限で拒否、閲覧から権限拡張なし | 担当者節。D15境界照合、未実施 |
| AC-analytics-lstep-denominator | 同飼主へ2つのfiredログと一つの来院 | ログ単位で計数しunique飼主率と呼ばない | [FR-analytics-lstep-observation](#requirement-fr-analytics-lstep-observation)。元ログと対応比較、未実施 |
| AC-analytics-lstep-window | 配信当日と30日後、31日後の来院 | 現行30日窓は当日／30日後を含み31日後を除外 | [FR-analytics-lstep-observation](#requirement-fr-analytics-lstep-observation)。JST日付境界、未実施 |
| AC-analytics-lstep-zero-error | 配信母数0／通信失敗を比較 | 前者の率0と後者の不明を区別 | [FR-analytics-lstep-observation](#requirement-fr-analytics-lstep-observation)。応答／表示比較、未実施 |

## 未決事項・関連仕様・根拠

来院集計の変更は2026-10-15の確認6（Sentinel_4230d91e7dd88191b67f30714ed073ba）へのユーザー「ok」（Sentinel_ae272463dc248191b51fbef540e04109）が根拠。実来院日、飼主は同日1回、ペットは各1回を承認範囲とする。既存の他計数とD10の金銭境界を保持し、D10新台帳の読取adapter、税務・外部会計ソフトへの適合と実運用受入は設計／検証残件として分離する。既存計数を新方針の業務承認にせず、変更が必要なら別の明示決定を残す。[技術整理](../../../design/domains/finance-operations/technical-spec.md#analytics-current-gap)が現行ソースと差分の参照先。

<a id="analytics-metric-dictionary"></a>
## 指標辞書・帰属日の参照表

以下は本書の既存FRと正本の接続索引であり、新指標や統一日界の採用ではない。現行読取との差・未検証を保持する。

| 指標／単位 | 帰属・区別 | 正本 |
|:---|:---|:---|
| 請求確定額／請求 | 実入金・未納回収・預り金と別。[2026-10-20本人回答](../accounting/lifecycle.md#accounting-bill-date-agreement-20261020)で原請求日は請求確定日。確定前は未存在で、作成日／会計予定日を代用しない。診療日・実入金日を別保持 | [金銭分離](#requirement-fr-analytics-money-separation)・[D10](../accounting/lifecycle.md#accounting-open-decisions) |
| 月次日次明細／会計 | 旧consumer観察はcompleted_atのJST暦日（最終DDLの存続列ではない）。旧顧客売上SQLのカルテ日／予定日帰属・完了状態filterと別 | [期間](#requirement-fr-analytics-period-basis) |
| 実受領・実返金／金銭イベント | 実日時と元資金・請求を区別。保険負担は保険会社の実入金ではなく実受取額へ加算しない | [金銭分離](#requirement-fr-analytics-money-separation)・[D10引継ぎ](../../../design/domains/accounting/technical-spec.md#accounting-handoff-contract) |
| 部門×方法／配賦・会計distinct | 返金は発生日負値、旧consumerの支払帰属はcompleted_at（退役列）。最終契約では[D10の元確定event／実入金と日付projectionの対応](../accounting/lifecycle.md#accounting-bill-date-agreement-20261020)を別途照合し、adapter実装済みとしない。行件数を総件数へ単純合算しない | [matrix](#requirement-fr-analytics-matrix-report)・[配賦正本](../../screens/29-closing-aggregation.md#closing-allocation-contract) |
| レジ営業日／AM・PM・EMG | am_start前の深夜は前日EMG。月次JST暦日と差があり日付ラベルだけで一致要求しない | [レジ正本](../../cash-register.md)・[既存境界AC](#requirement-fr-analytics-period-basis) |
| 来院回数／医院・飼主・日、医院・ペット・日 | 確認済み実来院日。同日複数頭・再受付の重複加算なし。年次rolling365日と暦年売上は別 | [来院FR／既存AC](#requirement-fr-analytics-owner-counts) |

今期対象外の顧客ランキング・CPM等をこの表で公開範囲へ戻さず、欠損／不明を0や来院なしに変換しない。

<a id="analytics-sales-tax-basis-20261022"></a>
## 月次売上の主表示（2026-10-22本人回答）

決定者はユーザー（MinoruSoga、2026-10-22の直接回答「現行カルテに合わせる」。受付記録は[PR #815](https://github.com/MinoruSoga/AnimalEkarte/pull/815)の本文）。現行システムの日次集計一覧表・現金受払一覧・精算画面は、部門別と純売上を税抜で示し、消費税・値引金・税込の売上合計を別列で併記する。月次売上レポートもこれに合わせ、主表示は税抜（純売上・部門別）、消費税と税込売上合計を併記する。保険負担分を実受取へ加算しない既決、[請求日＝請求確定日](../accounting/lifecycle.md#accounting-bill-date-agreement-20261020)、[金銭の分離](#requirement-fr-analytics-money-separation)は維持する。列の配置・丸め・CSV形式は設計へ分け、実装・実機は未確認。

| 条件ID | 前提／操作 | 期待結果 | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-analytics-sales-tax-basis | 税抜明細と税込明細が混在する月の売上レポートを表示・出力する | 主表示は税抜の純売上・部門別で、消費税と税込売上合計を併記する。税込入力の元額は保持し、保険負担分を実受取へ加算しない | 現行帳票の列構成と照合。未実施 |
