# 経営・顧客集計の業務契約

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.analytics.business-rules／業務仕様 |
| 決定状態 | 既存計数を継承し現行SQLとの相違を明示。D10の請求／実移動分離を継承。外部到達・来院因果の保証なし。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-15 |

## 目的・対象範囲

D16は月次／任意期間レポート、飼主集計、Lステップ分析の読取・出力を扱う。請求／入金・預り金残高の更新はD10、締め確定はD12、LINE配信・タグ変更・CSV取込はD15、臨床記録はD03等。集計閲覧やCSV出力を自動配信の承認にしない。税務申告や外部会計ソフトへの適合を保証しない。

## 担当者・権限・前提

/accounting/reports は accounting-reports:view、/aggregation は owners:view、/lstep/analytics は lstep-analytics:view。医院scopeを維持する。締め履歴ドリルダウンには cash-register-close:view を追加確認。税率設定リンクは対象設定の閲覧権限が必要。同じ分析画面のCSV取込には独立した lstep-csv-import 権限が必要で、分析閲覧から許可を派生しない。

## 契機・入口・出口

[月次画面](../../screens/32-accounting-reports.md)、[顧客集計画面](../../screens/36-aggregation-dashboard.md)、[Lステップ分析入口](../../screens/31-lstep-integration.md)が入口。出力時は対象医院・期間・金額基準・取得範囲を保持する。飼主詳細へのリンクは対象飼主を参照し、ページ集計から新しい患者・請求を作らない。

## 状態遷移・引継ぎデータ

<a id="requirement-fr-analytics-period-basis"></a>
### FR-analytics-period-basis：指標と期間を対にする

月次レポートはJST当月を既定とし、月または開始／終了日を選ぶ。現行日次明細はcompleted_atのJST暦日で、D12のAM開始を日界にした営業日ではない。深夜の会計は両者で別の日付行へ入る。日付ラベルだけで締めと月次の一致を要求しない。

顧客売上タブの既定は当暦年、明示期間があればそれを使う。annual_visit_countはrolling365日で暦年売上と別。現行SQLでは金額・会計件数の期間帰属がカルテ日（なければ会計予定日）であり、旧文書の『その年に完了した会計』をcompleted_at期間条件と誤読しない。completed状態フィルタと日付帰属の列は別である。D10の新請求・実入金に対応する読取adapterは未検証で、現行SQLを新台帳の正本としない。

<a id="requirement-fr-analytics-money-separation"></a>
### FR-analytics-money-separation：売上・回収・返金・預り金

D10の既決区分を継承。請求確定額、今回の実入金、実返金、未納、預り金受領／充当／復元を区別する。未納回収・預り金受領は新しい診療売上ではなく、充当／復元／配分変更は新しい実入金でない。請求取消／減額だけで実返金を作らない。保険負担額は保険会社から受け取った額でない。

現行顧客集計には売上総額／入金額／返金控除後の切替があるが、選択期間の会計に紐付いた入金・返金を集計する旧SQLであり、その期間に現金が動いた額とは限らない。累計欄と選択期間欄も区別する。D12の実入出金イベントと突合するときは実日時基準を別に表示・照合し、旧指標へ無言で混入しない。新台帳への切替時のラベル・互換表示は設計事項。

<a id="requirement-fr-analytics-matrix-report"></a>
### FR-analytics-matrix-report：同一条件での表示・出力

月次の日次明細・部門×支払方法表・KPI／内訳、CSV／印刷は同じ選択条件を使う。C7-7の既決に従い、午前／午後の件数・売上4列は廃止対象で、会計件数・日計へ集約しAM／PM締め状態は保持する。EMG集計を含めた日次件数を『午前＋午後』という表示文言だけで誤定義しない。現行合算のEMG網羅性は設計検査へ。

部門×支払方法の唯一の配賦正本は[画面29](../../screens/29-closing-aggregation.md#closing-allocation-contract)。返金は発生日負値、支払は現行completed_at基準、件数は会計distinct、行件数の単純合計を総会計件数にしない。旧KPIの返金は親会計完了日帰属で、差が起こり得る。この差を表示上の丸め誤差として消さない。D10移行後の期間基準は新イベント／請求の意味を維持する。

<a id="requirement-fr-analytics-owner-counts"></a>
### FR-analytics-owner-counts：飼主と来院日・母集団

飼主単位の来院回数は現行SQLで同医院・同飼主の重複しないカルテ日。複数ペットが同日に来ても同じ日付は1回で、会計件数・予約件数・診療明細数とは違う。診療記録への過去入力は臨床ownerの契約に従い、分析のために実来院を捏造しない。この既存指標は正確には「記録のある日数」で、母集団は同医院・対象飼主のペットに結び付いた未削除カルテ。現SQLにはfinalized限定の条件がなく、受付済／診療中／実来院済みの判定を行う指標へ拡張しない。「来院回数」という現行ラベルから実来院・完了回数を推定しない。

[D03の過去入力](../medical-records/lifecycle.md#requirement-fr-medical-records-retrospective-no-card-no-billing)は実診療日と登録日時を分けるため、過去記録を今日の来院として扱わない。[患者転記](../medical-records/lifecycle.md#requirement-fr-medical-records-wrong-patient-linked-transfer)は新しい実施でなく、カルテ確定も全治療項目の実施完了ではない。これらの臨床的意味と既存SQLの記録日数を区別し、今回新しい「実来院件数」への集計変更は追加しない。

CPM人数は一般条件適用後、CPM区分フィルタとページ分割前の母集団を使う。一覧totalはCPM絞込み後。全7区分の合計と母集団が一致し、0件でも7キーを0で返す。売上タブの0円除外を来院タブに無条件適用せず、最終来院軸の『来院なし』条件を他軸へ無断適用しない。

<a id="requirement-fr-analytics-segmentation"></a>
### FR-analytics-segmentation：CPM・最終来院・配信設定

顧客集計のCPMはV1の6区分＋未分類を用いる。配信設定がV2でも分析が自動的にV2へ切り替わる現行契約ではない。同期側6区分と分析の残余区分を混同しない。V1閾値設定と固定の最終来院バケットは別軸。

最終来院の境界は90日未満、90〜179日、180〜364日、365日以上、来院なし。『3ヶ月』等のラベルを暦月差で再実装しない。タイムゾーン・日付差の再現条件を検証する。CPMや売上ランキングは配信同意／送信可能性を示さない。

<a id="requirement-fr-analytics-lstep-observation"></a>
### FR-analytics-lstep-observation：配信統計と来院観測

現行Lステップ月次統計はscheduled_atのJST月でトリガー種別×状態のログ件数を集計。来院率はfired_atが当月・fired状態・優先度抑止でないログを母数とする。UIは配信後30日、APIは1〜365日。現在の来院判定はfired_atのJST日から指定日後まで両端を含むカルテ日が同医院・飼主にあること。

単位は配信ログで、ユニーク飼主数ではない。同じ飼主への複数ログが各々母数になり、一つの来院が複数ログに対応し得る。firedは外部サービスへの最終到達／既読の証明でなく、来院相関は配信による因果効果を証明しない。観測窓未経過の当月値は後日変わり得る。母数0の率0と、通信失敗で値不明を区別する。

## 異常系・保存・取消

読取失敗時は旧データ／0件／取得中と区別する。対象期間の不正入力はAPIエラーのまま扱い、無言の別期間集計をしない。検索条件変更後に古い応答が新条件の値として出ないことを検査する。

顧客CSVは現在のサーバーページで選択した行のみ。未選択時は出力不可で、全検索結果を出力したと表示しない。月次CSVの対象は選択した集計期間であり同じ出力契約ではない。友だち属性CSV取込・外部タグ同期はD15の別操作・別権限・保存結果を参照する。分析画面の表示成功を取込成功にしない。

## 受入条件

| 条件ID | 前提／操作（Given / When） | 期待結果（Then） | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-analytics-midnight-basis | 翌02時の会計を月次と営業日締めで比較 | JST暦日と前日EMGの差を説明できる | [FR-analytics-period-basis](#requirement-fr-analytics-period-basis)。両基準を照合、未実施 |
| AC-analytics-year-v-365 | 暦年とrolling365日が異なる日で集計 | 売上期間と年次来院の別定義を維持 | [FR-analytics-period-basis](#requirement-fr-analytics-period-basis)。年初境界、未実施 |
| AC-analytics-old-period-column | カルテ日と完了日が別月の既存会計 | 現行顧客SQLのカルテ日帰属と完了状態を区別 | [FR-analytics-period-basis](#requirement-fr-analytics-period-basis)。旧SQL契約比較、未実施 |
| AC-analytics-old-ar | 過去未納を今日回収 | 今日入金と過去請求、売上再生成なし | [FR-analytics-money-separation](#requirement-fr-analytics-money-separation)。D10移動と請求比較、未実施 |
| AC-analytics-deposit | 預り受領→充当→取消復元 | 売上／実入金を各段階で重複加算しない | [FR-analytics-money-separation](#requirement-fr-analytics-money-separation)。元資金参照、未実施 |
| AC-analytics-refund-dates | 前月会計へ当月返金 | 発生日matrixと親会計帰属KPIの差を保持 | [FR-analytics-matrix-report](#requirement-fr-analytics-matrix-report)。正本の例を比較、未実施 |
| AC-analytics-export-consistency | 同一医院／期間で画面・CSV・印刷 | 同じ集計対象・条件と円単位合計 | [FR-analytics-matrix-report](#requirement-fr-analytics-matrix-report)。出力3種比較、未実施 |
| AC-analytics-distinct-billing | 複合決済と複数部門を持つ会計 | 総件数をsplit数・部門行件数の合計にしない | [FR-analytics-matrix-report](#requirement-fr-analytics-matrix-report)。元ID比較、未実施 |
| AC-analytics-same-day-pets | 同飼主の2ペットに同日カルテ | 現行記録日数は1、会計件数とは別 | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。飼主／日付distinct、未実施 |
| AC-analytics-record-days-not-visits | 未削除のdraftカルテと確定カルテが異なる日に存在 | 現行指標は2記録日。実来院2回・完了2回とは説明しない | [FR-analytics-owner-counts](#requirement-fr-analytics-owner-counts)。SQL母集団とラベルの意味を照合、未実施 |
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

D16の新たな重大業務判断は追加しない。既存の計数とD10の金銭境界を整理し、D10新台帳の読取adapter、税務・外部会計ソフトへの適合と実運用受入は設計／検証残件として分離する。既存計数を新方針の業務承認にせず、変更が必要なら別の明示決定を残す。[技術整理](../../../design/domains/finance-operations/technical-spec.md#analytics-current-gap)が現行ソースと差分の参照先。
