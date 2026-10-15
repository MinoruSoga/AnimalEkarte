# 飼主・ペット

本領域の読む順は、[業務ライフサイクル・受入条件](lifecycle.md) → [用語と関係](../../ontology/owners-pets.md) → [技術契約・現在差分](../../../design/domains/owners-pets/technical-spec.md)。

画面詳細の既存正本は[03-owners-list](../../screens/03-owners-list.md)・[04-owners-form](../../screens/04-owners-form.md)に保持する。業務FR7件／AC17件は既存継承・合意継承・未確認を本文で区別し、FR採番を新しい一括承認としない。D01等の継承要件は承認件数へ二重加算しない。

## 対象と境界

飼主／患者の基本台帳、主飼主・副飼主、動物種マスタを所有する。identity linkはD19、LINE／健康手帳はD20、スタッフ専用飼主レポートはD21、請求／預り金はD10、外部配信はD15。台帳変更で過去履歴・資金帰属を自動移管しない。

## 文書状態と確認範囲

文書確認日2026-10-15。既存本文の決定状態・実装状態を継承し、限定ソースを静的照合した。独立レビュー／リンク検査後に限定文書整理の完了を判定する。実装・移行・実権限変更・実機受入・本番稼働を含めない。現行差と未決を消して完了にしない。

[全体進捗](../SPECIFICATION_PROGRESS.md)・[共通方針](../../shared/README.md)・[記述規約](../../authoring-guide.md#status)。

