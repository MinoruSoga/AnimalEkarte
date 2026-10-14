# 共通用語・権限・マスタ

共通事項の本文を複製せず、既存の正本へ案内します。予約区分・予約区分グループ（表示用）・予約区分の業務分類は別概念として、下記の用語正本を先に確認します。

| 調べたい事項 | 参照先 |
|:---|:---|
| 予約ドメインの語彙・概念関係と現行モデル対応 | [予約オントロジー正本](../ontology/reservations.md#reservation-ontology) |
| 全体要件・臨床安全原則 | [specification.md（§2.1）](../specification.md) |
| RBAC・医院分離の設計 | [認証・権限設計](../../architecture/auth.md) |
| 各操作の resource / action | 各ドメインの pages/ にある API 連携・権限表（[ドメイン入口](../README.md)） |
| マスタ設定 | [settings/ 索引](../screens/settings/README.md)・[マスタ設定ポータル](../screens/20-master-settings.md) |
| 共通患者選択・ダイアログ | [患者選択](../screens/00-pet-selection.md)・[共通ダイアログ](../screens/common-dialogs.md) |
| 共通 UI | [design-system.md](../design-system.md)・[components/](../components/README.md) |

共通マスタは当面 screens/settings/ に保持します。医院やトリミング固有のルールも各既存仕様を参照してください。未決事項の承認や実機受入の完了は、この索引への掲載からは判断しません。

[仕様全体の入口](../README.md)

予約の用語定義を本索引へ複製せず、[ontology/](../ontology/README.md)へ参照する。確定業務と旧資料・設計の優先関係は[予約入口](../domains/reservations/README.md#reservation-document-priority)を参照する。

## このフォルダの役割と使い方

共通用語・権限・マスタ・共通UIの参照入口。

共有事項の正本への案内を置く。語彙定義はontology、マスタ本文はscreens/settings、技術的認可設計はarchitectureを参照し、本索引に第二の定義を作らない。

子フォルダ：現在、子フォルダはありません。

主要文書：主要文書は本書の既存索引・上記の読む順を参照します。

読む順：本索引 → [ontology](../ontology/README.md)または[マスタ](../screens/settings/README.md) → 関連する領域FR。

正本・状態：[決定状態と検証状態の規約](../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。


<a id="interim-retention-policy"></a>
## 保存・廃棄の当面の共通方針

2026-10-14、ユーザー Sentinel_9d85a7955b608191985bb02c2c9356a5 が、当面は自動削除せず、具体的保存期間と削除手続きは全ドメイン共通の方針として別途確定する扱いを承認した。業務履歴・必要監査・予約再送の最小対応を、この当面の扱いへ接続する。

これは具体的保存年数・廃棄手続きの最終決定、無期限保存の法的義務、法令適合確認済みという意味ではない。共通の保存期間・削除手続きは確定待ちで、[運用資料U11](../../delivery/DELIVERY_PACKAGE.md)にもその境界を記録する。API／インフラログ等の既存プラットフォーム保持を今回の業務履歴方針から変更しない。

既存の取消・非表示・論理削除・復元・履歴参照の合意を保持する。自動削除しないことを、論理削除済み情報の一般公開や通常一覧への復帰と同義にしない。将来の削除手続きは履歴・監査・再送キー非再利用の整合も含めて確定する。予約の技術保存契約は[§3.2](../../design/domains/reservations/technical-spec.md#reservation-storage-details)を参照する。
