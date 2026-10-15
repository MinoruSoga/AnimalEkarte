# トリミング

本領域の読む順は、[業務ライフサイクル・受入条件](lifecycle.md) → [用語と関係](../../ontology/trimming.md) → [技術契約・現在差分](../../../design/domains/trimming/technical-spec.md)。

画面詳細の既存正本は[16-trimming-list](../../screens/16-trimming-list.md)・[17-trimming-form](../../screens/17-trimming-form.md)に保持する。業務FR7件／AC23件は既存継承・合意継承・未確認を本文で区別し、FR採番を新しい一括承認としない。D01等の継承要件は承認件数へ二重加算しない。

## 対象と境界

施術の専用記録・画像・コース／オプション／コース種別を所有する。appointmentsはD01 writer、受付表示はD02、通常カルテはD03、未請求取込・会計確定はD10。完了後専用訂正は2026-10-15追加回答で承認。実装適合は未確認。

## 文書状態と確認範囲

文書確認日2026-10-15。既存本文の決定状態・実装状態を継承し、限定ソースを静的照合した。独立レビュー／リンク検査後に限定文書整理の完了を判定する。実装・移行・実権限変更・実機受入・本番稼働を含めない。現行差と未決を消して完了にしない。

[全体進捗](../SPECIFICATION_PROGRESS.md)・[共通方針](../../shared/README.md)・[記述規約](../../authoring-guide.md#status)。

