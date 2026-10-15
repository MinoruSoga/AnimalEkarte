# 見積・締め・在庫・分析の用語と境界

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | ontology.finance-operations／語彙正本 |
| 決定状態 | 各業務正本の決定状態を継承。用語整理は新しい業務承認ではない。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-15 |

## 定義

| 推奨語 | 意味／混同しないもの | 業務正本 |
|:---|:---|:---|
| 見積 | 診療等の予想費用・提示条件を記録するもの。請求／領収書でない | [D11](../domains/estimates/business-rules.md) |
| 見積承認／却下 | 原見積をロックする状態。臨床実施・請求確定を意味しない | [状態](../domains/estimates/business-rules.md#requirement-fr-estimates-state-protection) |
| 後継見積 | 元見積を保持し別ID／番号で修正内容を作る下書き。unlock／同番号再発行でない | [後継](../domains/estimates/business-rules.md#requirement-fr-estimates-successor) |
| 見積の概算保険額 | 提示時点の概算。窓口の確認済み保険負担額や保険会社実入金ではない | [金額](../domains/estimates/business-rules.md#requirement-fr-estimates-money-source) |
| 営業日／区分 | AM開始から翌AM開始までをAM／PM／EMGに分割した帰属。JST暦日と別 | [締め境界](../domains/cash-register/business-rules.md#requirement-fr-closing-period-boundary) |
| 実入出金 | 実受領／実返金の資金移動。請求・充当・残高復元・配分変更と別 | [会計語彙](accounting.md)・[締めconsumer](../domains/cash-register/business-rules.md#requirement-fr-closing-movement-consumer) |
| 実査／理論現金／過不足 | 実際に数えた額／対象移動から導く額／実査−理論。売上を同義語にしない | [実査](../domains/cash-register/business-rules.md#requirement-fr-closing-cash-reconciliation) |
| 元締め／訂正差分 | 当時snapshot／後日判明した変更。元締めの再計算・再開ではない | [append-only](../domains/cash-register/business-rules.md#requirement-fr-closing-append-only) |
| 実査誤入力の訂正 | 元締めを保持する権限者・理由付きの追記。有効実査／過不足は変わるが、現金移動・売上を作らない | [実査訂正](../domains/cash-register/business-rules.md#requirement-fr-closing-actual-input-correction) |
| 在庫品目／管理単位 | 理論数量を管理する医院内の物品と単位。価格masterや臨床投与単位とは別 | [D13](../domains/inventory/business-rules.md) |
| 現在庫／実績調整 | 現時点の理論数量／実物に基づく増減・棚卸差異。後者はI1の承認範囲のみ | [調整](../domains/inventory/business-rules.md#requirement-fr-inventory-adjustment) |
| 残少／在庫切れ | 正の数量が正閾値以下／数量0以下。閾値0は在庫切れを解除しない | [導出](../domains/inventory/business-rules.md#requirement-fr-inventory-quantity-status) |
| 実物返却 | 実際に戻した物品数量。カルテや請求の取消だけで成立しない | [調整](../domains/inventory/business-rules.md#requirement-fr-inventory-adjustment) |
| 売上／入金額／返金控除後 | 別指標。選択期間の親請求帰属と実移動日帰属を必ず区別 | [D16金額](../domains/analytics/business-rules.md#requirement-fr-analytics-money-separation) |
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
