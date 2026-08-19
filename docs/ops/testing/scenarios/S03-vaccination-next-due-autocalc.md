# S03: ワクチン接種→次回予定自動計算

> **目的**: ワクチン接種登録時に次回予定日が接種間隔ルールから自動計算され、その結果が画面上で常時確認できること（自動化には監視が伴うこと — [product-philosophy.md ⑤](../../../product-philosophy.md)）を納品前に証明する。
> **所要目安**: 15分 / **深度**: 深い
> **仕様正本**: [specification.md §2.1](../../../spec/specification.md)・[screens/14-vaccinations-list.md](../../../spec/screens/14-vaccinations-list.md)・[screens/15-vaccinations-form.md](../../../spec/screens/15-vaccinations-form.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: vet ロール。
- 対象ペット: ペット検索で「ステータス=生存」の犬を 1 頭選ぶ。
- ワクチンマスタ: 犬に適合するワクチン（名称に「ワクチン」を含む・対象種が犬または両方）と、フィラリア予防（名称に「フィラリア」を含む）が存在すること。
- 依存シナリオ: なし。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 予防接種一覧から新規登録 → ペット選択 → 接種フォームを開く | 実施日のデフォルトが当日である（[15 §1.1](../../../spec/screens/15-vaccinations-form.md)）。クエリ `petId`（camelCase）でペットが引き継がれる場合あり |
| 2 | 犬用ワクチンをマスタから選択し、ロット番号を 1 件入力する | 選択・入力が反映される。ロット入力欄は lot1〜lot4 の最大 4 つ（[15 §1.2](../../../spec/screens/15-vaccinations-form.md)）。ワクチン選択肢は `useGetAllVaccinesMaster` の active 行（種フィルタなし — 犬/猫製品が混在しうる） |
| 3 | 次回予定で標準間隔「3週後」（value=`3weeks`）を選択する | 次回予定日が実施日 + 3 週で自動計算されて表示される（`calculateNextDate` — [15 §1.3](../../../spec/screens/15-vaccinations-form.md)「自動算出」） |
| 4 | 標準間隔を「1年後」（value=`1year`）に切り替える | 次回予定日が実施日 + 1 年で再計算される（[15 §1.3](../../../spec/screens/15-vaccinations-form.md)）。フォーム既定の nextScheduleType も `1year` |
| 5 | 「以外（手動）」（value=`other`）を選び、カレンダーから任意の日付を直接指定する | 臨床判断による手動調整が可能で、指定した日付が保持される。payload の `next_schedule_type` は `other`（`custom` ではない） |
| 6 | 保存して予防接種一覧に戻る | 一覧の「次回予定」列に登録した次回予定日が表示され、自動計算結果を画面上でいつでも確認できる（[14 §1.2/§2](../../../spec/screens/14-vaccinations-list.md)） |
| 7 | 同じペットにフィラリア予防を新規登録する | フィラリア等の予防も同じ予防接種機能で管理される（[14 概要](../../../spec/screens/14-vaccinations-list.md)「フィラリア・ノミダニ予防等」）。**runtime PASS（2026-08-01 browser）**: フィラリア薬選択後も次回予定候補は **3週後 / 4週後 / 1年後 / 以外（手動）**（ワクチンと同一セット・既定 1年後）。保存は未実施 |

## 確認観点

- 接種間隔の計算ロジックはフロントの `calculateNextDate`。選択肢 UI は `3weeks` / `4weeks` / `1year` / `other`。独立画面の既定は `1year`。カルテ予防接種タブも同じ算出を配線済み（既定は `4weeks`）。
- **#125 / BUG-401 回帰**: ワクチンマスタを実クエリし、保存後の記録のワクチン名称が選択どおりであること（ハードコード id 表は廃止）。seed 003_demo は修正済み（`backend/migrations/seeds/003_demo/vaccinations.csv`）。Lステップのフィラリアリマインドが混合ワクチン接種に対して発火しないこと。
- 種によるマスタ絞り込みは未実装（BUG-408 残置）— 犬/猫製品が combobox に混在しうるのは現状仕様。
- 接種期限接近の飼主通知は LINE 配信トリガー `vaccine_deadline_60d` / `vaccine_deadline_30d` が担う（[14 §2](../../../spec/screens/14-vaccinations-list.md)。受付画面での自動アラート表示はない）。
- 登録・更新が `audit_logs` に記録されること（[specification.md §2.1](../../../spec/specification.md)）— DB 参照は USER 実施。
- カルテ「予防接種」タブも `NextScheduleField` + `calculateNextDate` を使う。本シナリオの手順は独立画面を正とし、カルテ経路はラジオ切替で日付が変わることだけ確認する。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 実施日に未来日を指定しようとする | DatePicker で未来日は選択不可（`disabledDays: { after: 当日 JST }`）。保存時バリデーションでも未来日は fieldError（BUG-024） |
| A2 | 猫のペットで犬専用ワクチンを選択しようとする | 種フィルタなしのため選択肢に犬製品が出る（現状仕様・BUG-408）。接種 create はペット種とワクチン種の不一致を拒否しない |
| A3 | ロット番号を 5 つ目まで追加しようとする | 入力欄が lot1〜lot4 の 4 つのみで 5 つ目を追加する UI はない（[15 §1.2](../../../spec/screens/15-vaccinations-form.md)） |

---

## 実装突合

- 突合日: 2026-08-07
- HEAD: `844e43f69`
- 変更サマリ:
  - 次回予定 value（`3weeks`/`4weeks`/`1year`/`other`）。独立画面既定 `1year`、カルテタブ既定 `4weeks`
  - マスタ種フィルタなし（BUG-408）・`petId` クエリ・未来日 `disabledDays` を追記
