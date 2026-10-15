# 受付

当日の患者進行・日付切替・予約から記録への遷移を確認する入口です。

## 読む順・業務ルールの所在

まず [全体要件と臨床安全原則](../../specification.md) §2.1 を確認し、次に以下の既存本文を読みます。

[横断フロー](../../workflows/reservation-to-record-flow.md) §2.3・§2.4・§5 → [当日の受付](pages/reception-board.md)の「主要機能」「操作詳細」を参照します。入口に業務ルールの本文を重複掲載しません。

## 受入条件の所在

画面固有の挙動・バリデーション・保護・API 権限は上記本文、具体的な確認手順は [S40 受付の戻り先](../../../ops/testing/scenarios/S40-reception-return-navigation.md)・[S01 死亡ペットのガード](../../../ops/testing/scenarios/S01-deceased-pet-guard.md)を参照します。これらは条件・手順の所在であり、今回の整理で実機受入済みになったことを示しません。実行状態と証跡の入口は [docs/work/](../../../work/README.md)です。

## 未決事項と実装状態

横断フロー §7 の未決・既知ギャップを確認します。通常診療とトリミングの既存遷移ルールは画面本文と横断フローを参照します。 新しい要望は決定前に既存の確定ルールへ混ぜず、決定状態と実装・検証状態を分けて扱います。

## 共通事項

関連マスタは [予約種別](../../screens/settings/master-reservation-type.md)・[スタッフ](../../screens/settings/master-staff.md)を参照します。[共通用語・権限](../../shared/README.md)・[仕様全体の入口](../../README.md)。

## このフォルダの役割と使い方

受付領域の業務ルール・受入条件・未決事項を辿る入口。

pagesに画面詳細を置き、保存契約など領域固有本文は既存の配置を維持する。横断フロー・共通マスタ・語彙をここで再定義しない。

子フォルダ：[pages/](pages/README.md)

主要文書：主要文書は本書の既存索引・上記の読む順を参照します。

読む順：本書の「読む順・業務ルールの所在」 → [画面詳細](pages/README.md) → 関連する横断契約・マスタ。

正本・状態：[決定状態と検証状態の規約](../../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。


<a id="reception-agreements-current"></a>
## 受付3点の現在の合意と残り

2026-10-14、Sentinel_ea8569c91d8c8191b70e7b2cf8df8db6 の2・3への回答と Sentinel_daeb79f859f48191a38d3a61f4f725d8 の1への「ok」を、[画面の業務正本・受入条件](pages/reception-board.md#reception-agreement-20261014)へ反映した。

読む順は[分類別確認・専用管理](pages/reception-board.md#requirement-fr-reception-business-class-handoff) → [選択日限定](pages/reception-board.md#requirement-fr-reception-selected-day-only) → [会計済の片付け・再表示](pages/reception-board.md#requirement-fr-reception-completed-card-visibility) → [受入条件](pages/reception-board.md#acceptance)。当日の残件案内・持越し表示という先の提案は不採用。予約で確定した担当者条件・記録時点・自動状態変更禁止・同じ対象IDを維持する。

会計済カードの非表示設定は Sentinel_67648c4b0030819193d97aacb34e9b16 と、ブラウザ保存を指定した Sentinel_7fc2cd6f8c708191b6ab1d9160cb836b により、医院×スタッフ別に同じブラウザで再ログイン後も保持する。別端末／別ブラウザへ同期せず、他スタッフの表示へ影響させない。「片付けたカードも表示」で対象日の確認・表示復元ができる。保存は最小表示ID／状態だけで、患者／認証情報を含めず、ブラウザ削除／保存不可で設定が失われても業務データは不変。[正本と受入条件](pages/reception-board.md#requirement-fr-reception-completed-card-visibility)へ参照する。再表示UI・分類別列／区画と通信異常・競合は[画面具体化](pages/reception-board.md#reception-display-layout)と[受付技術仕様](../../../design/domains/reception/technical-spec.md)で選定。専用管理内部の記録本文・業務遷移・担当人数は各ownerへ残す。今回の3点反映・文書レビューを受付全体の文書完了、API反映、実装／実機受入済みとは扱わない。未承認内部を一括確定せず、共通の保存期間・削除手続きも[当面方針](../../shared/README.md#interim-retention-policy)と分ける。


<a id="reception-remaining-decisions-current"></a>
### 受付全体の残件判定と次の整理案

ブラウザ保存の最新合意までを既存の予約・横断・受付正本へ照合した範囲では、追加のユーザー判断が必要な受付固有の業務残件は見つからない。旧「未決」案内を新たな質問へ転用せず、次をまとまり単位で整理する。これは新仕様の一括承認や受付全体の文書完成ではない。

| まとまり | 現在状態／推奨する整理 |
|:---|:---|
| 分類・表示日・片付け | 4分類確認／専用管理、選択日限定、医院×スタッフ別ブラウザ保持は確定。列・区画・filter・確認／復元配置は画面具体化で選定済み。実機の操作・accessibility検証は別 |
| 受付・開始／終了・訂正 | 同時成功、同じ記録再利用、明示開始／終了、Undo、会計確定境界、その他逆遷移の記録確認は既合意。古い案内を正本参照へ整理し再質問しない |
| 権限・集計・通信・保存 | 既存grantと選択日集計を継承。型・認可・戻り先・取得失敗／競合／結果不明・browser保存不可／値検査は受付技術仕様へ選定し、受入条件を画面へ対応。実装／検証は別。実装・実機状態は未確認 |
| 専用内部・共通保存 | 記録本文、世話担当の職種／人数、会計訂正／精算内部は各ownerの完成条件。具体的保存期間／削除手続きは全ドメイン共通で別途確定。受付だけで決めない |

全ドメイン文書確定後の連携再照合を残し、その前に実装へ着手しない。表示設計の具体化で既合意の業務結果・役割・状態を変える選択が必要になった場合だけ、新しい業務判断として分離する。


<a id="reception-document-completion-current"></a>
## 受付の表示・異常時契約の文書整理完了境界

2026-10-14、Sentinel_6bdf54904d3881918fb3e884820adfea のdocs限定承認で、[配置とfilter](pages/reception-board.md#reception-display-layout)、[異常時表示・入力保持](pages/reception-board.md#reception-network-contract)、[受入条件](pages/reception-board.md#acceptance)、[技術方式](../../../design/domains/reception/technical-spec.md)を一括具体化した。技術フォルダの対象・読む順は[説明README](../../../design/domains/reception/README.md)へ参照する。

受付の既合意と今回の表示・通信／競合・browser設定・受入対応は文書として揃った。追加の受付固有業務判断待ちは発見していない。文書整理完了の最終判定は独立レビューと限定文書検査の結果を添えて報告し、アプリ／API反映・実機受入・利用開始を完了にしない。

関連領域の専用記録本文・施術／宿泊内部状態・世話担当人数・会計訂正／精算・共通保存期間／削除手続きはそれぞれの完成条件へ引き継ぐ。旧データ・分類／専用ref・全件snapshot・grant・browser保存・遅延応答／再送・並行処理・性能の実装前検証と、全ドメイン確定後の連携再照合を残す。その前に実装しない。

## 2026-10-15会議追加2

[開始列・各記録画面終了・会計待ちの記録導線](pages/reception-board.md#reception-meeting-start-end)を新しい業務合意として反映する。トリミングの開始後は診療中列へ、通常カルテ／専用トリミング画面の終了から同じ予約を会計待ちへ進める。会計待ちでは既存記録を開き、終了だけで記録・請求・入金を確定しない。旧表示設計の「受付済の施術中badge」は置き換える。既存4FR／21ACのIDを維持して5ACを追加し、4FR／26ACを文書候補とする。過去の文書整理完了は今回差分のレビュー完了を示さず、実装・実機検証は未実施。

### 初診／再診・次回予定の接続（2026-10-15）

[ペットごとの受診歴による候補](../reservations/pages/reservation-management.md#requirement-fr-reservations-pet-history-visit-type)と[日付のみの次回予定・明示消化・取消後の手動再設定](../../workflows/reservation-to-record-flow.md#next-visit-intent-agreement-20261015)を参照する。画面は[受付への反映](pages/reception-board.md#next-visit-intent-reception)、技術接続は[予約§4.8](../../../design/domains/reservations/technical-spec.md#reservation-visit-history-intent-contract)へ。4FR／受入条件は予約側で数え、受付へ重複定義しない。受診歴は通常診療・ワクチン・健診を含み、トリミング・ホテルだけの利用は除外。実装・実機未確認。
