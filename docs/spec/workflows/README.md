# 横断業務フロー

[予約からカルテ入力までの統合フロー](reservation-to-record-flow.md)が予約・受付・記録連携の正本です。§1 の用語、§2 の入口別フロー、§5 の lifecycle / transaction safety、§7 の未決・既知ギャップの順に読みます。§5 には会計完了時の appointment 更新も含まれます。

会計・未納の画面操作と保護は [会計ドメイン](../domains/accounting/README.md)、保存ルーティングの契約は [カルテ lifecycle §2](../domains/medical-records/lifecycle.md)を参照します。レジ締めは [cash-register.md](../cash-register.md)の既存仕様を参照します。

横断フロー内の未実装・deferred・履歴の記載はそのまま保持します。記載された契約と、実装・受入・リリースの確認状況は分けて扱います。

[仕様全体の入口](../README.md)
