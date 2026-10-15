# 入院・ホテル

本領域の読む順は、[業務ライフサイクル・受入条件](lifecycle.md) → [用語と関係](../../ontology/hospitalization.md) → [技術契約・現在差分](../../../design/domains/hospitalization/technical-spec.md)。

画面詳細の既存正本は[07-hospitalization-list](../../screens/07-hospitalization-list.md)・[08-hospitalization-detail](../../screens/08-hospitalization-detail.md)・[09-hospitalization-form](../../screens/09-hospitalization-form.md)に保持する。業務FR8件／AC29件は既存継承・合意継承・未確認を本文で区別し、FR採番を新しい一括承認としない。D01等の継承要件は承認件数へ二重加算しない。

## 対象と境界

臨床入院とホテル宿泊の記録・実滞在・ケージ・プラン・退院／退室を所有する。予約頭数はD01の正本を継承し、会計確定はD10へ。担当＝担当医師／指名非必須は2026-10-15ユーザー修正を反映。責任担当1名必須・非医師可の提案は不採用。専用訂正とホテル料金運用、入院中の元履歴保持による治療プラン追加・変更・中止は2026-10-15追加回答で承認。死亡状態を戻さない生前分後日入力・誤記訂正・退院処理／精算も追加承認を反映。実装適合は未確認。

## 文書状態と確認範囲

文書確認日2026-10-15。既存本文の決定状態・実装状態を継承し、限定ソースを静的照合した。独立レビュー／リンク検査後に限定文書整理の完了を判定する。実装・移行・実権限変更・実機受入・本番稼働を含めない。現行差と未決を消して完了にしない。

[全体進捗](../SPECIFICATION_PROGRESS.md)・[共通方針](../../shared/README.md)・[記述規約](../../authoring-guide.md#status)。

