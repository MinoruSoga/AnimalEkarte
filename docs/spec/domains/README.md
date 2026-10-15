# domains — フォルダの説明


## このフォルダの役割と使い方

業務領域ごとの仕様入口をまとめる。

今回の全体範囲・領域数・完了境界と残件は[仕様整理の全体進捗](SPECIFICATION_PROGRESS.md)で管理する。実装は今回の対象外。

領域READMEに業務ルール・受入条件・未決の所在を示し、pagesに画面詳細、必要な保存契約を領域内に置く。複数領域のフロー、共通語彙、共通マスタの本文を複製しない。

子フォルダ：[accounting/](accounting/README.md)、[medical-records/](medical-records/README.md)、[reception/](reception/README.md)、[reservations/](reservations/README.md)

主要文書：各領域READMEが、その領域の主要文書・受入条件の入口です。

読む順：[全体要件](../specification.md) → [予約](reservations/README.md) → [受付](reception/README.md) → [カルテ](medical-records/README.md) → [会計・未納](accounting/README.md)。横断する条件は[workflows](../workflows/README.md)、共通概念は[ontology](../ontology/README.md)・[shared](../shared/README.md)。

正本・状態：[決定状態と検証状態の規約](../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。
