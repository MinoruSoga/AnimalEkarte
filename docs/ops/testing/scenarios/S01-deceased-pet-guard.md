# S01: 死亡ペット誤操作の物理ブロック

> **目的**: 死亡登録されたペットに対する新規の臨床・会計・配信操作が物理的にブロックされ、死亡した個体への誤診療・誤請求・誤配信が発生しないことを納品前に証明する。
> **所要目安**: 15分 / **深度**: 深い
> **仕様正本**: [specification.md §2.1](../../../spec/specification.md)・[screens/00-pet-selection.md](../../../spec/screens/00-pet-selection.md)・[screens/02-reservations.md](../../../spec/screens/02-reservations.md)・[screens/04-owners-form.md](../../../spec/screens/04-owners-form.md)・[screens/09-hospitalization-form.md](../../../spec/screens/09-hospitalization-form.md)・[line/lstep-integration.md §4](../../../spec/line/lstep-integration.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: admin ロール。
- 対象ペット: ペット検索で「ステータス=生存」の犬を 1 頭選ぶ（未来日の予約・入院中レコードを持たない個体を選ぶこと）。
- 依存シナリオ: なし（S02〜S06 より先に実行し、対象ペットは本シナリオ終了時に生存へ戻す）。
- ローカルでは Lステップへの実配信は発生しない前提で、手順 6 は画面・ログで観測可能な範囲のみ確認する。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 飼主詳細のペット編集で対象ペットの臨床ステータスを「死亡」に変更し、死亡日（当日）・理由を入力して保存（[04-owners-form.md §1.2](../../../spec/screens/04-owners-form.md)） | 保存成功。ペット一覧・検索結果で【死亡】バッジが表示され、行がグレーアウトする（[00-pet-selection.md](../../../spec/screens/00-pet-selection.md)） |
| 2 | 新規予約: 予約管理の予約登録フォームで患者検索し対象ペットを表示 | **既定検索は死亡ペットを結果に含めない**（includeDeceased なし + deceased_at IS NULL）。グレーアウトは手順 3〜5 の共通ペット選択 |
| 3 | 新規カルテ: カルテ新規作成のペット選択画面で対象ペットを検索 | 選択ボタンが「選択不可」に切り替わり、クリックが無効化される（[00-pet-selection.md](../../../spec/screens/00-pet-selection.md)「臨床安全ガード」） |
| 4 | 新規会計: 会計新規作成のペット選択画面で対象ペットを検索 | 手順 3 と同様に「選択不可」で無効化される（同上。会計もペット選択共通画面を使用） |
| 5 | 新規入院: 入院新規登録のペット選択画面で対象ペットを検索 | 「選択不可」で無効化され、死亡済みペットへの新規登録が物理的にブロックされる（[09-hospitalization-form.md](../../../spec/screens/09-hospitalization-form.md)） |
| 6 | Lステップ配信対象: 対象ペットの飼主に紐づくリマインド予定を配信監視で確認 | 死亡と同時に関連リマインドが破棄され、配信対象から外れる（[lstep-integration.md §4](../../../spec/line/lstep-integration.md)）。【要実測】ローカルでの観測方法（死亡は配信直前チェックではなくライフサイクル処理のため、`excluded` 行として現れない — [34-lstep-delivery-monitor.md](../../../spec/screens/34-lstep-delivery-monitor.md)） |
| 7 | 死亡解除: 手順 1 と同じ編集画面でステータスを「生存」に戻し保存 | 保存成功。手順 2〜5 の各導線で対象ペットが再び選択可能になる（各ガードの条件は「ステータスが死亡 (deceased)」であるため） |

## 確認観点

- 生存判定はペットの `deceased_at` / ステータス列に基づく。フロントの無効化実装は `PatientSelectionTable.tsx`（予約）と `PetSelectionResultsTable`（ペット選択共通）。
- ブロックの見え方は経路で異なる（2026-07-17 実測）: デフォルト一覧では「グレーアウト＋選択不可」で見えるが、**飼主 No 検索の結果からは死亡ペットが除外され 0 件になる**（見えない）。検索経路で 0 件でも異常ではない。
- 死亡登録はサブダイアログの確定時に API 保存され、外側のペット編集フォームにも保存結果が同期される。死亡登録のために外側フォームの「更新」を重ねて押す必要はない（`PetDeceasedDialog.tsx`・`PetDeceasedRecordButton`）。
- 死亡タグ除去のバックエンド処理は `HandlePetDeath`（Lステップ ライフサイクル処理）。
- 死亡登録・解除の変更が `audit_logs` に記録されること（[specification.md §2.1](../../../spec/specification.md)）— DB 参照は USER 実施。
- 全操作が同一クリニック内で完結すること（clinic_id 隔離）。
- 転院（transferred）も Lステップ配信破棄の対象（[lstep-integration.md §4](../../../spec/line/lstep-integration.md)）だが、本シナリオでは死亡のみを扱う。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 未来日の予約を持つペットを死亡登録する | 【要実測】既存予約の扱い（自動キャンセルされるか、残存して受付カンバンに出続けるか）。仕様文書に明記なし |
| A2 | 死亡登録済みペットの編集画面を再度開く | 死亡日・理由が保持されて表示される（[04-owners-form.md §1.2](../../../spec/screens/04-owners-form.md)「死亡時は日付と理由を記録」） |
| A3 | 死亡登録済みペットの過去のカルテ・会計履歴を参照する | 【要実測】過去記録の閲覧は可能であること（ブロック対象は新規データ作成 — [00-pet-selection.md](../../../spec/screens/00-pet-selection.md)。閲覧可否の明記なし） |
