# トリミング

本領域の読む順は、[業務ライフサイクル・受入条件](lifecycle.md) → [用語と関係](../../ontology/trimming.md) → [技術契約・現在差分](../../../design/domains/trimming/technical-spec.md)。

画面詳細の既存正本は[16-trimming-list](../../screens/16-trimming-list.md)・[17-trimming-form](../../screens/17-trimming-form.md)に保持する。業務FR7件／AC47件（既存42件＋終了条件5件）は既存継承・合意継承・未確認を本文で区別し、FR採番を新しい一括承認としない。D01等の継承要件は承認件数へ二重加算しない。

## 対象と境界

施術の専用記録・画像・コース／オプション／コース種別を所有する。appointmentsはD01 writer、受付表示はD02、通常カルテはD03、未請求取込・会計確定はD10。完了後専用訂正は2026-10-15追加回答で承認。[最新第5項合意](lifecycle.md#trimming-partial-save-undecided)では担当・コース未定の途中保存を可能にする。[終了条件の追加承認](lifecycle.md#trimming-end-recorded-performance)により、終了までに実担当と施術内容を記録する。コースなしでも既存オプション等の施術明細があれば終了可能。実装適合は未確認。

## 文書状態と確認範囲

文書確認日2026-10-15。既存本文の決定状態・実装状態を継承し、限定ソースを静的照合した。会議追加2・3より前の差分は独立内容レビュー／リンク検査を経て文書整理完了と判定した。今回の終了条件・旧制限同期の完了判定は[全体の有限チェック](../SPECIFICATION_PROGRESS.md#all-domain-closure-20261015)に従う。実装・移行・実権限変更・実機受入・本番稼働を含めない。現行差と未決を消して完了にしない。

[全体進捗](../SPECIFICATION_PROGRESS.md)・[共通方針](../../shared/README.md)・[記述規約](../../authoring-guide.md#status)。

## 2026-10-15会議追加2・3

[施術開始列・専用画面終了](lifecycle.md#trimming-meeting-start-end)と[カテゴリUI・施術料金調整](lifecycle.md#trimming-meeting-pricing-category)を反映する。開始後は診療中列、専用画面の終了から会計待ち、会計待ちカードは同じ既存記録へ接続する。料金手修正・値引きを会計pullへ渡し、D10の既決計算順・二重適用防止・master／確定請求保護を維持する。コース／オプション行の調整をD10明細scopeへ対応付け、会計全体値引き・追加後差分確認／明示再適用もD10既決を継承する。新しいscope・初期値・値引順序を設けない。過去のレビュー完了を今回追加差分へ読み替えず、実装・実機受入は未実施。
