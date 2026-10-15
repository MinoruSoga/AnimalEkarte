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

2026-10-14、ユーザー Sentinel_9d85a7955b608191985bb02c2c9356a5 が当面自動削除しない扱いを承認した。その後2026-10-15、質問 Sentinel_75a3a1bf81b481919ba4c34aa633feb4 の第6項と再確認 Sentinel_8d0032ca45b881919b3d2ceb5339f792 に対する回答 Sentinel_7a67fd30e47c819192cd36d416f0a2bf「ok」で、以下の製品範囲を確定した。

業務履歴と必要監査を保持し、通常操作は取消・非表示・論理削除までとする。保存年数による自動削除・一括物理削除は今回の製品仕様に含めない。必要時は対象・保存義務・承認・復旧方法を確認する別手続きで扱う。具体的な法定保存年数、法的無期限保存、法令適合や個別物理削除実行を承認した意味ではない。[運用資料U11](../../delivery/DELIVERY_PACKAGE.md)も同じ境界を参照する。

業務履歴・必要監査・予約再送の最小対応を本方針へ接続する。既存の取消・非表示・論理削除・復元・履歴参照の合意を保持し、保存することを一般公開や通常一覧への復帰と同義にしない。必要時の別手続きは履歴・監査・再送キー非再利用の整合を確認する。[予約の技術保存契約](../../design/domains/reservations/technical-spec.md#reservation-storage-details)を維持する。

API／インフラログ等の既存プラットフォーム保持を今回の業務履歴方針から変更しない。期限付き配送コピー・署名URLも業務原本の保持とは区別する。ただし任意Uploadを一律に一時コピーと推定せず、原本所在・保持と一時分類が確認できない対象は自動cleanupへ含めない。現行143時間cleanupとの差・確認までの安全側停止条件・URL期限との違いは[保存監査の整理](retention-audit.md#requirement-fr-common-retention-delivery-copy)を参照する。文書変更でruntimeの停止／設定変更を行った意味ではない。

## 全6共通領域の契約

| ID | 正本 | 責務 |
|:---|:---|:---|
| C01 | [共通UI・選択](ui-contracts.md) | 非同期／入力保持／遅延応答・印刷と共通部品 |
| C02 | [用語・文書状態](document-contracts.md) | 意味ID・関係・決定／実装／検証の区別 |
| C03 | [臨床安全](clinical-safety.md) | 患者・確定・数値・訂正・副作用の受入 |
| C04 | [認証・権限・医院scope](access-control.md) | 現在認可・scope例外・回復と権限変更 |
| C05 | [保存・監査](retention-audit.md) | 業務原本／必要監査保持・通常論理操作と別削除手続 |
| C06 | [技術接続](../../design/shared/integration-contracts.md) | owner・保存単位・意味型／版・再送／旧値・証跡 |

各領域の全受入を共通原則だけで代替せず、[全23領域](../domains/README.md)・[横断8接続](../workflows/reservation-to-record-flow.md#cross-domain-review-20261015)の具体契約と照合する。

## 全ownerへの適用照合

[全23領域の具体的な意味ID・保存・安全・保持・失敗条件との対応](owner-contract-coverage.md)を参照する。原則の再掲だけで個別操作の受入を満たしたとは扱わず、引用先のFR／AC・技術契約と照合する。

## 2026-10-15会議4点の最新入口

自動保存維持・終了後通常編集・タブ開始時復元／複製直前取消・終了履歴、トリミング開始／終了・料金、勤務複製／CSVの合意を追加整理。[全体進捗](../domains/SPECIFICATION_PROGRESS.md#meeting-four-decisions-20261015)と各owner本文を参照。途中の保存ボタン方式は撤回済み。今回差分は独立内容・構造レビューで重大指摘なしを確認し、文書再整合済み。全過去承認の監査完了・実装完了とは扱わない。
