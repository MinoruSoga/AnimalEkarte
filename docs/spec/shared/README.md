# 共通用語・権限・マスタ

共通事項の本文を複製せず、既存の正本へ案内します。

| 調べたい事項 | 参照先 |
|:---|:---|
| appointment・来院・記録の用語と source of truth | [横断フロー §1](../workflows/reservation-to-record-flow.md) |
| 全体要件・臨床安全原則 | [specification.md（§2.1）](../specification.md) |
| RBAC・医院分離の設計 | [認証・権限設計](../../architecture/auth.md) |
| 各操作の resource / action | 各ドメインの pages/ にある API 連携・権限表（[ドメイン入口](../README.md)） |
| マスタ設定 | [settings/ 索引](../screens/settings/README.md)・[マスタ設定ポータル](../screens/20-master-settings.md) |
| 共通患者選択・ダイアログ | [患者選択](../screens/00-pet-selection.md)・[共通ダイアログ](../screens/common-dialogs.md) |
| 共通 UI | [design-system.md](../design-system.md)・[components/](../components/README.md) |

共通マスタは当面 screens/settings/ に保持します。医院やトリミング固有のルールも各既存仕様を参照してください。未決事項の承認や実機受入の完了は、この索引への掲載からは判断しません。

[仕様全体の入口](../README.md)
