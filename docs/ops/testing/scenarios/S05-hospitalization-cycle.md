# S05: 入院サイクル（ケア記録→退院会計）

> **目的**: 入院登録からデイリーケア記録、ケージ稼働の可視化、退院時の会計統合までの入院業務一巡が成立し、退院済み患者への二重退院が拒否されることを納品前に証明する。
> **所要目安**: 25分 / **深度**: 深い
> **仕様正本**: [specification.md §3.1](../../../spec/specification.md)・[screens/07-hospitalization-list.md](../../../spec/screens/07-hospitalization-list.md)・[screens/08-hospitalization-detail.md](../../../spec/screens/08-hospitalization-detail.md)・[screens/09-hospitalization-form.md](../../../spec/screens/09-hospitalization-form.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: vet ロール（hospitalization の create/edit 権限あり）。
- 対象ペット: ペット検索で「ステータス=生存」のペットを 1 頭選ぶ。
- ケージマスタに、現在入院中（active）の患者が割り当てられていないケージが 2 つ以上あること（ケージ移動の確認に使用）。
- 依存シナリオ: なし。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 入院管理から新規入院登録 → ペット選択 → 入院タイプ「入院」・予定期間・空きケージ・主訴を設定し、ケアプランに給餌指示と投薬指示を各 1 件登録して保存（[09 §1/§2](../../../spec/screens/09-hospitalization-form.md)） | 保存成功。入院一覧の「入院中」タブに表示される（[07 §2](../../../spec/screens/07-hospitalization-list.md)） |
| 2 | 一覧をボードビューに切り替える | 割り当てたケージの位置に患者カードが表示され、ケージ稼働状況が可視化される（[07 §1](../../../spec/screens/07-hospitalization-list.md)・[specification.md §3.1](../../../spec/specification.md)) |
| 2b | リストビューへ切り替える | 入院期間・主訴・担当医を含む表形式で当該入院が表示される（[07 §1](../../../spec/screens/07-hospitalization-list.md)） |
| 3 | 入院詳細を開き、デイリーログにバイタル 1 件（時刻・測定値）とケアログ 1 件を追加する | それぞれ独立に記録され、日付単位の時系列で「バイタル」「ケアログ」セクションに表示される（[08 §2/主要機能1](../../../spec/screens/08-hospitalization-detail.md)） |
| 4 | ボードビューで患者カードを別の空きケージへドラッグする | 収容ケージが直感的に変更され、移動先ケージにカードが表示される（[07 主要なユーザーアクション](../../../spec/screens/07-hospitalization-list.md)） |
| 5 | 入院詳細から退院処理（会計あり）を実行する | 確認ダイアログを経てステータスが退院（discharged）になり、入院中に蓄積された明細を含む会計詳細画面または新規会計作成フォームへ自動遷移する（[08 §2 退院プロセスと会計連携](../../../spec/screens/08-hospitalization-detail.md)）。会計明細にケアプラン由来の項目が含まれる |
| 6 | 入院一覧に戻り、タブとボードを確認する | 当該患者が「入院中」タブから消え「退院済」タブに表示される（[07 §2](../../../spec/screens/07-hospitalization-list.md)）。【要実測】ボードビューで当該ケージが空きに戻ること（仕様に明記なし） |
| 7 | 2 件目の入院を作成し、会計を作成せずに退院する経路を実行する | 【要実測】会計なし退院の UI 経路と挙動（バックエンドは会計作成の有無をフラグで受けるが、UI 導線は仕様に明記なし）。退院ステータスのみ更新され会計レコードが作られないこと |

## 確認観点

- 退院と会計生成は同一トランザクションで原子的に行われる（片方だけ成立しない）— `backend/internal/service/hospitalization_service.go` の `DischargeWithBilling`。
- 退院前の最終確認は `DischargeAlertDialog` が担う（[08 技術仕様](../../../spec/screens/08-hospitalization-detail.md)）。
- 二重退院はバックエンドが行ロック（FOR UPDATE）で直列化し「already discharged」で拒否する（同上・直近修正）。
- 入院登録・ケア記録・退院が `audit_logs` に記録されること（[specification.md §2.1](../../../spec/specification.md)）— DB 参照は USER 実施。
- 入院フォームのケージ割り当ては空き状況によるフィルタリングを行わない（[09 §1](../../../spec/screens/09-hospitalization-form.md)）— 占有中ケージが選択できても仕様どおりであり逸脱ではない。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 退院済み入院に対して再度退院操作を試みる（詳細画面の連打含む） | 【要実測】退院済みでは退院導線が出ない、または実行時にエラーが 1 回だけ表示される。UI 連打でも会計レコードが二重生成されないこと（仕様に明記なし・バックエンドは拒否を保証） |
| A2 | ケアプランが紐付いたままの入院レコードを一覧から削除する | 【要実測】ケアプラン紐付き入院は削除が拒否される（バックエンド実装のガード。仕様に明記なし） |
