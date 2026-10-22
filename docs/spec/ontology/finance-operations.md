# 見積・締め・在庫・分析の用語と境界

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | ontology.finance-operations／語彙正本 |
| 決定状態 | 各業務正本の決定状態を継承。用語整理は新しい業務承認ではない。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-20 |

<a id="estimate-ontology"></a>
## 定義

| 推奨語 | 意味／混同しないもの | 業務正本 |
|:---|:---|:---|
| 見積 | 診療等の予想費用・提示条件を記録するもの。請求／領収書でない | [D11](../domains/estimates/business-rules.md) |
| 提示済み原本 | 飼主へ示した見積の内容版。提示時点から返答待ちでも金額・明細を直接上書きせず、変更は元版を保持する後継・再提示で扱う | [提示版の保護](../domains/estimates/business-rules.md#requirement-fr-estimates-state-protection) |
| 見積承認／却下 | 当該提示版への返答を記録する状態。原本保護の開始とは別で、臨床実施・請求確定を意味しない | [回答状態と内容保護の分離](../domains/estimates/business-rules.md#requirement-fr-estimates-state-protection) |
| 後継見積 | 元見積を保持し別ID／番号で修正内容を作る下書き。unlock／同番号再発行でない | [後継](../domains/estimates/business-rules.md#requirement-fr-estimates-successor) |
| 見積の概算保険額 | 提示時点の概算。窓口の確認済み保険負担額や保険会社実入金ではない | [金額](../domains/estimates/business-rules.md#requirement-fr-estimates-money-source) |
| 営業日／区分 | AM開始から翌AM開始までをAM／PM／EMGに分割した帰属。JST暦日と別 | [締め境界](../domains/cash-register/business-rules.md#requirement-fr-closing-period-boundary) |
| 実入出金 | 実受領／実返金の資金移動。請求・充当・残高復元・配分変更と別 | [会計語彙](accounting.md)・[締めconsumer](../domains/cash-register/business-rules.md#requirement-fr-closing-movement-consumer) |
| 実査／理論現金／過不足 | 実際に数えた額／対象移動から導く額／実査−理論。売上を同義語にしない | [実査](../domains/cash-register/business-rules.md#requirement-fr-closing-cash-reconciliation) |
| 元締め／訂正差分 | 当時snapshot／後日判明した変更。元締めの再計算・再開ではない | [append-only](../domains/cash-register/business-rules.md#requirement-fr-closing-append-only) |
| 実査誤入力の訂正 | 元締めを保持する権限者・理由付きの追記。有効実査／過不足は変わるが、現金移動・売上を作らない | [実査訂正](../domains/cash-register/business-rules.md#requirement-fr-closing-actual-input-correction) |
| 在庫品目／管理単位 | 理論数量を管理する医院内の物品と単位。価格masterや臨床投与単位とは別 | [D13](../domains/inventory/business-rules.md) |
| 現在庫／実績調整 | 現時点の理論数量／実物に基づく増減・棚卸差異。後者はI1と[2026-10-20出庫確定回答](../domains/inventory/business-rules.md#inventory-outflow-confirmation-20261020)の承認範囲 | [調整](../domains/inventory/business-rules.md#requirement-fr-inventory-adjustment) |
| 残少／在庫切れ | 正の数量が正閾値以下／数量0以下。閾値0は在庫切れを解除しない | [導出](../domains/inventory/business-rules.md#requirement-fr-inventory-quantity-status) |
| 実物返却 | 実際に戻した物品数量。カルテや請求の取消だけで成立しない | [調整](../domains/inventory/business-rules.md#requirement-fr-inventory-adjustment) |
| 売上／入金額／返金控除後 | 別指標。選択期間の親請求帰属と実移動日帰属を必ず区別 | [D16金額](../domains/analytics/business-rules.md#requirement-fr-analytics-money-separation) |
| 月次売上の主表示基準 | 主表示は税抜、消費税と税込売上合計を併記。定義は会計語彙に置き、ここでは二重定義しない。別指標の区別は上記を維持 | [会計語彙](accounting.md)・[D16](../domains/analytics/business-rules.md#analytics-sales-tax-basis-20261022) |
| 来院回数（承認済み目標） | 確認済み実来院日の数。同医院・飼主・同日は1回、ペット単位は各1回。予約数・カルテ数・会計数でない | [計数](../domains/analytics/business-rules.md#requirement-fr-analytics-owner-counts) |
| 旧記録日数（現行顧客SQL） | draftを含む未削除カルテ日distinct。実来院回数へ無条件変換しない | [現行差](../domains/analytics/business-rules.md#requirement-fr-analytics-owner-counts) |
| CPM／最終来院バケット | V1等の継続性分類／90・180・365日固定分類。別軸 | [分類](../domains/analytics/business-rules.md#requirement-fr-analytics-segmentation) |
| 配信後来院率（現行） | firedログ単位の来院観測比。ユニーク飼主率・外部到達率・配信効果の因果推定でない | [観測](../domains/analytics/business-rules.md#requirement-fr-analytics-lstep-observation) |

## 関係・多重度・制約

| 関係 | 多重度／制約 |
|:---|:---|
| 原見積→後継 | 元を参照する別見積。原本不変。同じ再送intentの重複は設計で防ぐが、全履歴で後継1件限定という新ルールは置かない |
| 見積→会計draft | 確認した対象明細を受ける別owner。見積承認だけでは生成しない。飼主／ペット・元行参照を維持 |
| 実移動→営業日／区分 | 実日時から一つに帰属。元請求日と同じとは限らない。同じ移動IDを配分数で複製しない |
| 締め→訂正 | 一つの不変な締めに0以上の追記。削除・巻戻し・再締めなし |
| 在庫→臨床／master参照 | 参照元は複数あり得る。名前一致による同一性判定を一般契約にしない |
| 在庫実績→元実績／理由 | I1承認範囲で履歴を関連付け、返却・訂正・再送を二重増減しない |
| 飼主／ペット→実来院日／会計 | 同日複数ペットでも飼主1回、ペット各1回。同一source由来の再受付・転記を重複加算せず、会計は別IDで計数 |
| 分析母集団→CPM7区分 | 区分フィルタ・ページ前の母集団へ排他的に割当、未分類を落とさない |
| 配信ログ→来院観測 | 一ログにつき窓内来院有無。同一来院が複数ログの観測に寄与し得る |

## 表示・モデルとの照合

旧『年間』『純売上』『配信数』『十分』というラベルだけで意味を確定しない。識別子・APIを改名せず、[技術整理](../../design/domains/finance-operations/technical-spec.md)の取得式と照合する。会計の実移動・預り金等は[accounting](accounting.md)を再利用し、ここで定義を上書きしない。

## 見積の提示・単価保証（2026-10-15追加承認）

提示済み原本と承認／却下の意味は[定義表](#estimate-ontology)を参照する。「期限内の提示単価」は承認済み・同内容・同数量を条件に使う単価であり、保険負担や総額の無条件固定ではない。2026-10-15追加承認：承認済み・期限内で内容と数量が同じなら提示単価を使う。追加処置や内容・数量変更は再説明し、提示版の変更は同じ原本保護・後継の契約へ従う。税・保険・割引は既存適用条件を維持し、総額の無条件固定や自動請求確定に拡張しない。 [D11正本](../domains/estimates/business-rules.md#estimate-price-presentation-agreement-20261015)。

## 締め区分・差額メモの追加合意

締め済み保護は実際に締めた区分だけに適用し、午前締めだけで未締めの午後・夜間を保護対象にしない。確定済み会計の訂正条件は維持する。現金差額が残っていても締められるが、非0差額ではメモ必須とし「原因不明・確認中」も許容する。差額合わせの架空入出金は作らない。 [D12正本](../domains/cash-register/business-rules.md#closing-segment-difference-agreement-20261015)。区分未締めは会計未確定と同義ではない。過不足メモは実査結果の説明で、実入出金イベントではない。

## 日次集計の追加内訳

[原資料採用正本・追加2AC](../screens/29-closing-aggregation.md#closing-source-breakdown)へ従い、往診／トレセンを独立表示し、病院／トリミング別の保存済み税・値引きを表示する。分類・税・値引き・支払実額は別概念で、表示配賦で元会計総額を変えない。金額配賦の契約本文は画面29の唯一正本へ維持し、旧分類数を目標の固定数にしない。

<a id="discharge-report-and-accounting-time"></a>
## 日次退院情報・一覧日付・金銭日付

[滞在の時点](hospitalization.md#stay-identity-and-time-concept)で定義する実退院日時のJST暦日と、締めのAM/PM/EMG営業日・請求日・実受領日・実返金日を分離する。[退院情報](../domains/cash-register/business-rules.md#closing-daily-discharge-information-20261017)の退院日時/飼主名/ペット名/入院日という4項目は読取projectionで、会計がないこと/別日の入金を理由に別の退院日へ帰属させない。AC-closing-discharge-actual-jst-day-interval／AC-closing-discharge-calendar-day-not-closing-period／AC-closing-discharge-four-fields。退院情報行を追加しても売上・入出金やホテル頭数を増減しない。

[一覧条件](../shared/ui-contracts.md#requirement-fr-common-ui-list-filter)は表示コンテキストである。日付付き対象業務一覧はメニューから当日、詳細から戻る場合は日付・患者等検索条件を保持。「このペットで絞り込む」は全期間にして医院filterを保持。台帳/マスタ/入院/見積を当日対象へ一律化せず、受付の選択日と未納繰越の既存契約も保持する。

月次の実現金・現金差額の非掲載は[D16](../domains/analytics/business-rules.md#requirement-fr-analytics-matrix-report)、AC-analytics-monthly-no-cash-differenceへ。日次締めの実査/差額と月次売上内訳は別projection。CPM分類語彙や年間/犬種等の旧集計語彙があることを、今期対象外画面の提供承認へ読み替えない。今期の提供範囲は[D16業務正本](../domains/analytics/business-rules.md)を参照する。

## 出庫確定と請求準備の分離（2026-10-20）

[D13の本人回答](../domains/inventory/business-rules.md#inventory-outflow-confirmation-20261020)に従い、院内投薬の実施時、持帰薬・物販の実引渡し時に担当スタッフが出庫確定する。D10のready／準備完了は請求対象の条件で、実引渡し・出庫確定ではない。カルテ保存・請求確定だけでは在庫を減らさず、旧治療create自動減算は現行差として区別する。最小管理単位整数・不足拒否・履歴を維持し、小数臨床量の推測換算・新係数・丸めを追加しない。

[医院営業時間・休診日の共通基準](../domains/clinic-settings/README.md#clinic-hours-common-basis-20261020)とLINE範囲内別設定は、レジのAM／PM／EMG営業日区分と別である。
