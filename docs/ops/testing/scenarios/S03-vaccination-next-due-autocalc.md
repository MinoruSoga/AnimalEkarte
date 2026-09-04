# S03: ワクチン接種→次回予定自動計算

> **目的**: ワクチン接種登録時に次回予定日が接種間隔ルールから自動計算され、その結果が画面上で常時確認できること（自動化には監視が伴うこと — [product-philosophy.md ⑤](../../../product-philosophy.md)）を納品前に証明する。
> **所要目安**: 15分 / **深度**: 深い
> **仕様正本**: [specification.md §2.1](../../../spec/specification.md)・[screens/14-vaccinations-list.md](../../../spec/screens/14-vaccinations-list.md)・[screens/15-vaccinations-form.md](../../../spec/screens/15-vaccinations-form.md)

## 前提条件

- ローカルの使い捨て clinic に、medical-records create/edit 権限を持つ attached account と生存犬を作成する。
- 犬向けワクチンマスタとフィラリア予防マスタを承認済み fixture/import 手順で作成する。各マスタの接種間隔を記録し、固定 ID や retired demo seed を仮定しない。
- 試験後に作成した接種記録と専用 fixture を削除する。
- 依存シナリオ: なし。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 予防接種一覧から新規登録 → ペット選択 → 接種フォームを開く | 実施日のデフォルトが当日である（[15 §1.1](../../../spec/screens/15-vaccinations-form.md)）。クエリ `petId`（camelCase）でペットが引き継がれる場合あり |
| 2 | 犬用ワクチンをマスタから選択し、ロット番号を 1 件入力する | 選択・入力が反映される。ロット入力欄は lot1〜lot4 の最大 4 つ（[15 §1.2](../../../spec/screens/15-vaccinations-form.md)）。ワクチン選択肢は `useGetAllVaccinesMaster` の active 行（種フィルタなし — 犬/猫製品が混在しうる） |
| 3 | 次回予定で標準間隔「3週後」（value=`3weeks`）を選択する | 次回予定日が実施日 + 3 週で自動計算されて表示される（`calculateNextDate` — [15 §1.3](../../../spec/screens/15-vaccinations-form.md)「自動算出」） |
| 4 | 別のワクチンマスタを選び、間隔候補を確認する | 選択したワクチンマスタに間隔が設定されている場合、その値へ `next_schedule_type` と次回予定日が更新される。常に `1year` のままとは仮定しない。手動で `1year` を選べる場合は実施日 + 1 年になる |
| 5 | 「以外（手動）」（value=`other`）を選び、カレンダーから任意の日付を直接指定する | 臨床判断による手動調整が可能で、指定した日付が保持される。payload の `next_schedule_type` は `other`（`custom` ではない） |
| 6 | 保存して予防接種一覧に戻る | 一覧の「次回予定」列に登録した次回予定日が表示され、自動計算結果を画面上でいつでも確認できる（[14 §1.2/§2](../../../spec/screens/14-vaccinations-list.md)） |
| 7 | 同じペットにフィラリア予防を新規登録する | フィラリア等の予防も同じ予防接種機能で管理される（[14 概要](../../../spec/screens/14-vaccinations-list.md)「フィラリア・ノミダニ予防等」）。フィラリア薬選択後も次回予定候補は **3週後 / 4週後 / 1年後 / 以外（手動）**（ワクチンと同一セット・既定 1年後）。保存は未実施 |

## 確認観点

- 接種間隔の計算ロジックはフロントの `calculateNextDate`。選択肢 UI は `3weeks` / `4weeks` / `1year` / `other`。独立画面の既定は `1year`。カルテ予防接種タブも同じ算出を配線済み（既定は `4weeks`）。
- **#125 / BUG-401 回帰**: API から取得したワクチンマスタを選び、保存後の名称が選択どおりであること。固定 ID 表は使わない。フィラリアリマインドが混合ワクチン接種に対して発火しないこと。
- 種によるマスタ絞り込みは未実装（BUG-408 残置）— 犬/猫製品が combobox に混在しうるのは現状仕様。
- 接種期限接近の飼主通知は LINE 配信トリガー `vaccine_deadline_60d` / `vaccine_deadline_30d` が担う（[14 §2](../../../spec/screens/14-vaccinations-list.md)。受付画面での自動アラート表示はない）。
- 登録・更新が `audit_logs` に記録されること（[specification.md §2.1](../../../spec/specification.md)）— DB 参照は USER 実施。
- カルテ「予防接種」タブも `NextScheduleField` + `calculateNextDate` を使う。本シナリオの手順は独立画面を正とし、カルテ経路はラジオ切替で日付が変わることだけ確認する。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 実施日に未来日を指定しようとする | DatePicker で未来日は選択不可（`disabledDays: { after: 当日 JST }`）。保存時バリデーションでも未来日は fieldError（BUG-024） |
| A2 | 猫のペットで犬専用ワクチンを選択しようとする | 種フィルタなしのため選択肢に犬製品が出る（現状仕様・BUG-408）。接種 create はペット種とワクチン種の不一致を拒否しない |
| A3 | 次回予定日を接種日と同日または前日にして保存する | `next_date <= vaccination date` として拒否され、記録は保存されない |
| A4 | ロット番号を 5 つ目まで追加しようとする | 入力欄が lot1〜lot4 の 4 つのみで 5 つ目を追加する UI はない（[15 §1.2](../../../spec/screens/15-vaccinations-form.md)） |

---

## 実装突合

- 変更サマリ:
  - 次回予定 value（`3weeks`/`4weeks`/`1year`/`other`）。独立画面既定 `1year`、カルテタブ既定 `4weeks`
  - マスタ種フィルタなし（BUG-408）・`petId` クエリ・未来日 `disabledDays` を追記
