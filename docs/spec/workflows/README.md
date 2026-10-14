# 横断業務フロー

[予約からカルテ入力までの統合フロー](reservation-to-record-flow.md)が予約・受付・記録連携の正本です。用語は[ontology](../ontology/README.md)、次に本フローの§2の入口別契約と§5のlifecycle / transaction safetyを読み、未決・既知ギャップは§7の移設案内から[design](../../work/design/reservations/pending-and-design.md#workflow-gaps)へ辿ります。§5 には会計完了時の appointment 更新も含まれます。

会計・未納の画面操作と保護は [会計ドメイン](../domains/accounting/README.md)、保存ルーティングの契約は [カルテ lifecycle §2](../domains/medical-records/lifecycle.md)を参照します。レジ締めは [cash-register.md](../cash-register.md)の既存仕様を参照します。

横断フロー内の未実装・deferred・履歴の記載はそのまま保持します。記載された契約と、実装・受入・リリースの確認状況は分けて扱います。

[仕様全体の入口](../README.md)

## このフォルダの役割と使い方

複数業務領域をまたぐ状態・引継ぎ・保存境界の正本入口。

横断契約を置く。個別画面の詳細、語彙正本、カルテ保存ルーティング、旧記録・設計の本文を複製しない。

子フォルダ：現在、子フォルダはありません。

主要文書：主要文書は本書の既存索引・上記の読む順を参照します。

読む順：[予約入口](../domains/reservations/README.md)・[語彙](../ontology/README.md) → [予約から記録のフロー](reservation-to-record-flow.md) → 各領域の画面・lifecycle。未決／既知ギャップの移設先はフローの案内を辿る。

正本・状態：[決定状態と検証状態の規約](../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。
