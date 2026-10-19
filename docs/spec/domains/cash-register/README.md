# レジ締め（D12）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.cash-register.index／入口 |
| 決定状態 | 既存記載を継承（決定根拠未確認の部分を含む）。C1の目標は依頼者2026-10-15承認（Sentinel_7a67fd30e47c819192cd36d416f0a2bf）。追加確認8の実査誤入力は権限者・理由付き追記で訂正可能（Sentinel_4f050b1c0ed08191b5ca67deeaeec843）。原締め不変。既存詳細全体の再承認ではない。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-15 |

## 読む順と正本

1. [業務契約・FR／受入条件・例外・未決](business-rules.md)
2. [語彙と他領域境界](../../ontology/finance-operations.md)
3. [締め・履歴](../../screens/29-closing-aggregation.md)・[締め時間](../../screens/settings/closing-time-settings.md)
4. [現行実装差・技術契約](../../../design/domains/finance-operations/technical-spec.md)

業務契約と現行画面の説明は別軸。既存画面本文は削除せず、目標との相違を明示して参照する。未承認の提案を合意済みへ格上げしない。仕様整理、実装完了、実機受入を別に判定する。

## 区分保護と差額メモの追加合意

締め済み保護は実際に締めた区分だけに適用し、午前締めだけで未締めの午後・夜間を保護対象にしない。確定済み会計の訂正条件は維持する。現金差額が残っていても締められるが、非0差額ではメモ必須とし「原因不明・確認中」も許容する。差額合わせの架空入出金は作らない。 [根拠・受入条件](business-rules.md#closing-segment-difference-agreement-20261015)。FR6件を維持し、AC22件へ6件を追加した28件に、Backlog No.30の締め履歴フィルタ1件を加えた計29件に、日次／月次A4横出力の1ACを補完した計30件に、[No.4の退院情報掲載・印刷](business-rules.md#closing-daily-discharge-information-20261017)の3ACを加えた33件に、[実退院日時のJST暦日](business-rules.md#closing-discharge-actual-jst-day-20261017)の2ACを加えた35件に、[退院一覧の4項目](business-rules.md#closing-discharge-four-fields-20261017)の1ACを加えた計36件。実装適合は別工程。
