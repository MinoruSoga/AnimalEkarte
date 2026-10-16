# D14 職員・シフト

> 2026-10-15の仕様整理候補。PR #670の基準版から既存合意・既存仕様・未承認案を分離した。業務合意と実装／実機検証を別管理する。D22/D23は2026-10-15の一括承認を反映。技術設計案は業務承認と分ける。

2026-10-15の[シフト複製・CSV取込の追加合意](../../screens/24-shift-calendar.md#shift-copy-csv-agreement-20261015)を接続する。期間・スタッフ選択、Excel作成CSV、反映前の差分・エラー・予約影響確認、無断上書き禁止が業務合意の範囲。明示選択した置換だけを対象とし、エラー／未解決衝突は確認範囲全体を止め、選び直しは再プレビューする。相対日付保持・確定範囲の原子保存を設計へ具体化し、CSV等の物理設計と実装／実機受入は別管理。

## 目的・範囲

P1：職員の同一性と予約に使う勤務基礎を維持する。勤怠打刻・給与計算・有休残数管理は今回の仕様範囲外（追加範囲は別合意）。

## 正本と優先順

- [docs/spec/screens/settings/master-staff.md](../../screens/settings/master-staff.md)
- [docs/spec/screens/settings/master-occupation.md](../../screens/settings/master-occupation.md)
- [docs/spec/screens/settings/master-shift-template.md](../../screens/settings/master-shift-template.md)
- [docs/spec/screens/24-shift-calendar.md](../../screens/24-shift-calendar.md)

最新の明示合意が適用範囲内の旧記述に優先する。コードは実装事実の根拠で、業務承認の代替ではない。既存FR／ACの本文は元正本に維持し、本書の接続要件で改番・複製しない。

## 業務要件

<a id="requirement-fr-staff-workforce-identity"></a>
### FR-staff-workforce-identity — 職員とログインの分離

同一人物は1アカウント、兼務は医院所属追加で表す。ログインを持たない職員登録も認める。旧移行の複数staff行は権威ある本人対応に従い、氏名一致だけで統合しない。職種は分類であり、権限や対応能力を自動付与しない。

決定状態・根拠：既存仕様継承。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-staff-workforce-workforce-maintenance"></a>
### FR-staff-workforce-workforce-maintenance — 勤務入力とテンプレート

勤務・複数休憩・休日・有休を医院別に管理する。休日／有休は終日非勤務で時刻・休憩を保存しない。有効テンプレートから入力値をコピーし、テンプレート改訂を既存勤務への一括変更としない。日跨ぎ勤務・医院間重複禁止・移動時間の手動確保はシフト正本の既決に従う。

決定状態・根拠：既存仕様継承。テンプレートから入力への値コピー契約に基づく非遡及の境界整理（実装確認は別）。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

追加の入力経路は[FR-shifts-copy-selected-range](../../screens/24-shift-calendar.md#requirement-fr-shifts-copy-selected-range)、[FR-shifts-excel-csv-import](../../screens/24-shift-calendar.md#requirement-fr-shifts-excel-csv-import)、[FR-shifts-bulk-preview-confirm](../../screens/24-shift-calendar.md#requirement-fr-shifts-bulk-preview-confirm)を正本とする。複製／CSVも既存勤務のowner・医院scope・現在権限・予約保護を通り、反映前確認と明示反映を分離する。新しいFR／ACの本文を本書へ複製しない。

<a id="requirement-fr-staff-workforce-history-and-access"></a>
### FR-staff-workforce-history-and-access — 利用終了と履歴

職員無効化・一医院の勤務終了は区別し、他医院の勤務・履歴を消さない。将来・未終了の既指名予約保護は既存FRを参照し定義を複製しない。過去勤務の訂正は勤務情報の訂正に限定し、予約・カルテ・会計を自動変更しない。実際の退職者・メール・所属名簿は運用入力であり仕様決定で補完しない。

決定状態・根拠：予約保護は承認済み。過去勤務訂正から他owner履歴を自動改変しない境界整理。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-staff-workforce-booking-contract-index"></a>
### FR-staff-workforce-booking-contract-index — 予約保護の適用範囲

職員の無効化・全体退職、特定医院での勤務終了、医院別対応区分解除、勤務短縮・休憩追加・シフト削除は、該当する未来／未終了の既指名予約を解消するまで確定しない。自動取消・自動指名解除はしない。予定終了超過でも診療中は保護する。完了／取消／未来院の扱いと再受付時の適格性再確認は予約横断正本を参照し、対象集合を独断で拡張しない。B院の所属・勤務をA院の停止で消さない。

決定状態・根拠：既存master-staff／shift／予約横断FRの接続索引。元FRを再定義しない。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

## 受入条件

全条件は文書上の期待結果。実機未実施。技術方式の提案は実装設計として区別する。

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-staff-workforce-01 | 同一人物がA院に登録済み | B院勤務を追加 | 既存人物へ所属を追加し別ログインを増やさない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-identity](#requirement-fr-staff-workforce-identity)。実機未実施 |
| AC-staff-workforce-02 | 同姓同名の別人候補がある | 統合しようとする | 名称一致だけでは統合しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-identity](#requirement-fr-staff-workforce-identity)。実機未実施 |
| AC-staff-workforce-03 | 時刻と休憩を入力済み | 休日または有休へ変更する | 終日非勤務となり時刻・休憩を勤務扱いしない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-workforce-maintenance](#requirement-fr-staff-workforce-workforce-maintenance)。実機未実施 |
| AC-staff-workforce-04 | テンプレートで勤務登録済み | 後日テンプレートを改訂する | 既存勤務を無断変更しない。テンプレートは入力値のコピー元であり、登録済み勤務の一括改訂命令ではない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-workforce-maintenance](#requirement-fr-staff-workforce-workforce-maintenance)。実機未実施 |
| AC-staff-workforce-05 | 完了済み診療と過去勤務がある | 過去勤務を訂正する | 診療・会計・予約の履歴を自動変更しない。各ownerの明示訂正契約を維持する | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-history-and-access](#requirement-fr-staff-workforce-history-and-access)。実機未実施 |
| AC-staff-workforce-06 | A院の未来指名予約とB院勤務がある | A院での勤務終了を確定 | A院予約の解消を要求しB院勤務・予約・過去履歴は保持する | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-booking-contract-index](#requirement-fr-staff-workforce-booking-contract-index)。実機未実施 |
| AC-staff-workforce-07 | 診療中で予定終了を過ぎた指名予約がある | 担当者を退職無効化する | 該当予約の解消まで拒否し自動終了・取消をしない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-booking-contract-index](#requirement-fr-staff-workforce-booking-contract-index)。実機未実施 |
| AC-staff-workforce-08 | 過去完了の履歴だけがある | 職員を無効化 | 過去完了履歴だけを理由に永久に拒否せず履歴を保持する | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-booking-contract-index](#requirement-fr-staff-workforce-booking-contract-index)。実機未実施 |
| AC-staff-workforce-09 | 未登録勤務と他院勤務が混在 | 指名候補・勤務を参照 | 未登録を休日と推測せず他院勤務を自院勤務として扱わない。詳細判定は予約正本へ従う | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-staff-workforce-workforce-maintenance](#requirement-fr-staff-workforce-workforce-maintenance)。実機未実施 |

追加操作の受入はシフト画面の `AC-shifts-copy-selected-scope`、`AC-shifts-bulk-*`、`AC-shifts-csv-identity-errors`、`AC-shifts-template-copy-boundary` と[D14技術trace](../../../design/domains/staff-workforce/technical-spec.md#shift-bulk-acceptance-trace)へ接続する。既存AC-staff-workforce-01〜09と既存シフト保護ACを置換・改番しない。全条件の実機受入は未実施。

## 未決・責務境界

既存テンプレート改訂と過去勤務訂正の境界整理は新しい業務質問を生まない。追加合意の複製／CSVについては[選定設計と完成条件](../../screens/24-shift-calendar.md#shift-copy-csv-open-details)を別に管理する。日付対応・明示置換・確認範囲全体の拒否／原子保存と再プレビューは合意を守る安全な設計として具体化し、新たな業務許可を追加しない。CSVの具体仕様・上限等は後続設計へ残す。各画面の実装方式・既存予約不整合の移行救済・医院別能力保存・日跨ぎ表現・並行保存は引き続き設計／移行残件。

## レビュー状態

基準版の本文と関連する静的根拠を照合。独立レビュー・統合リンク検査は統合担当で実施する。仕様記述を実装完了、受入PASS、実環境変更承認へ変換しない。

## 操作・状態・例外の全域点検

| 操作群 | 参照する契約と境界 |
|:---|:---|
| 職員一覧・検索・職種絞込 | master-staffの一覧、職種分類。有効／無効・検索結果0件を未登録勤務や権限不足と同一にしない |
| 職員登録・アカウント追加 | 1人1account、無ログイン職員可、既存staffへのaccount attachはauth正本のsystem_admin専用。実発行・本人メールは今回未実施 |
| プロフィール・所属・権限group変更 | 選択医院scope、自己無効化／最後の管理者保護を保持。主所属・操作医院・勤務先を区別する。役職や氏名から自動権限付与しない |
| 退職／無効化・医院別指名停止・削除 | 予約保護を適用。無効化と削除を同一視せず、過去参照の削除制約を継承。実装は未来予約ガード等に未追従／未検証あり |
| 個人能力・LINE公開プロフィール | 肯定形の医院別対応区分と既決のLINE公開条件。新規区分は自動許可しない。画像はURL入力でアップロードを既存機能としない |
| 職種CRUD・有効無効 | 分類用途。職種の推奨groupという説明は自動付与契約ではない。並び替えAPIの存在を画面UI実装としない |
| 勤務作成・更新・削除・月切替 | 登録済み勤務と未登録を区別。予約保護・医院間重複拒否、日跨ぎ、複数休憩、休憩端点。過去訂正から診療履歴を改変しない |
| 期間・スタッフ複製／Excel作成CSV取込 | 入力案→差分・エラー・予約影響確認→明示反映。既存勤務の無断上書き・削除・除外なし。予約保護、現在scope／権限／版、原子保存・結果不明照会・部分結果の区別へ接続 |
| 休診日設定・解除 | 既存予約一覧の警告確認で設定可、取消なら未確定。シフト変更の強制保護と区別。予約／シフトを自動削除しない |
| template CRUD・並び替え・選択 | 医院内名称一意、有効templateだけ選択、休日／有休は時刻なし。入力コピーと登録済み勤務の再計算を区別 |
| 失敗・競合・再保存 | 保存対象を保持して理由と予約修正入口を提示。最新状態で再評価。別医院に自動再送しない。担当設定側の入力案保持や競合方式は設計残件で、既存シフト案保持の承認を勝手に拡大しない |

予約FR/ACの正本はstaff／shift画面と予約横断フロー。本節は確認範囲の索引であり新しい業務条件を追加しない。

根拠の詳細：AC04は[シフトtemplateの入力反映](../../screens/settings/master-shift-template.md#22-入力コストの削減)を参照する。登録済み勤務へ自動反映する機能の承認はなく、本案も追加しない。AC05は[職員無効化時の履歴保持](../../screens/settings/master-staff.md#requirement-fr-staff-deactivation-booking-guard)と各owner分離を参照する。過去勤務を訂正できる権限や専用workflowを新設する要件ではなく、既存操作が他owner履歴へ自動波及しない境界のみを扱う。

<a id="staff-source-adoption-20261015"></a>
## 社員リストからの追加採用（2026-10-15）

[社員リスト原資料](https://drive.google.com/file/d/13NLADSLBr9w1w_zQrfP45rvvW1Drbymn/view)とユーザーの職員属性・休業の採用回答に基づく追加目標。既存の人物同一性・医院scope・明示権限設定・予約保護を継承する。資料の全行を初期登録承認したものではなく、個人情報・未確定略称の補完を行わない。

- 職種・役職・雇用区分は別の属性として管理する。役職を職種へ混在させず、いずれの変更からも認可role／grantや診療能力を自動設定しない。ログインを持たない職員・複数医院所属も既存契約のまま扱う。
- 休業は在籍を保持した期間として管理する。対象医院・期間に重なる新規シフト登録と新しい指名候補から除外し、復帰後は同一人物・所属・履歴を保持して通常の候補条件へ戻す。復帰だけで勤務や指名を作らない。
- 期間変更時に既存勤務・指名予約との衝突を検知し、既存の勤務変更／予約保護へ渡す。既存予約を自動取消・移動せず、退職・アカウント停止へ変換しない。未登録勤務や他院勤務を自院の勤務可能と推測しない。

既存FR-staff-workforce-identity／workforce-maintenance／history-and-accessの追加受入（実機未実施）：

| AC | Given / When | Then |
|:---|:---|:---|
| AC-staff-source-attributes | 職種・役職・雇用区分を別々に変更する | 他属性を潰さず保存し、認可role／grantと過去履歴は自動変更されない |
| AC-staff-source-leave | 休業期間内と期間外に新規勤務・指名を設定する | 期間内の対象医院では除外し、期間外は既存条件で判定する。既存予約は取消されない |
| AC-staff-source-return | 休業期間を変更／終了する | 同一人物・所属・過去履歴を保持し、既存勤務／予約衝突は明示処理へ渡す。復帰による自動勤務作成はない |
