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
| 1 | 飼主詳細のペット編集で `PetDeceasedRecordButton` / `PetDeceasedDialog` から死亡日（当日）・理由を入力して死亡登録（[04-owners-form.md §1.2](../../../spec/screens/04-owners-form.md)）。死亡は generic PATCH ではなく `PATCH/POST …/pets/:id/death` 系（監査付き） | 保存成功（ダイアログ確定時に API 保存）。ペット一覧・検索結果で死亡バッジ／ステータス「死亡」が表示され、行がグレーアウトする（[00-pet-selection.md](../../../spec/screens/00-pet-selection.md)） |
| 2 | 新規予約: 予約管理の予約登録フォームで患者検索し対象ペットを表示 | **既定検索は死亡ペットを結果に含めない**（`PatientSelectionTable` の `includeDeceased` 既定 false → API `include_deceased` 未指定 → BE `deceased_at IS NULL`）。死亡を含める経路は手順 3〜5 |
| 3 | 新規カルテ: カルテ新規作成のペット選択画面で対象ペットを検索 | 共通選択（`usePetSelectionPage` は `includeDeceased: true`）で死亡個体も一覧に出るが、選択ボタンが「選択不可」で無効化される（`PetSelectionResultsTable`・[00-pet-selection.md](../../../spec/screens/00-pet-selection.md)「臨床安全ガード」） |
| 4 | 新規会計: 会計新規作成のペット選択画面で対象ペットを検索 | 手順 3 と同様に「選択不可」で無効化される（同上。会計もペット選択共通画面を使用） |
| 5 | 新規入院: 入院新規登録のペット選択画面で対象ペットを検索 | 「選択不可」で無効化され、死亡済みペットへの新規登録が物理的にブロックされる（[09-hospitalization-form.md](../../../spec/screens/09-hospitalization-form.md)） |
| 6 | Lステップ配信対象: 対象ペットの飼主に紐づくリマインド予定を配信監視で確認 | 死亡と同時に関連リマインドが破棄され、配信対象から外れる（[lstep-integration.md §4](../../../spec/line/lstep-integration.md)）。【要実測】ローカルでの観測方法（死亡は配信直前チェックではなくライフサイクル処理のため、`excluded` 行として現れない — [34-lstep-delivery-monitor.md](../../../spec/screens/34-lstep-delivery-monitor.md)）。**runtime 2026-08-01**: `/lstep/delivery-monitor` に除外カウンタはあるが、死亡→除外行の end-to-end は未実施（DEFER） |
| 7 | 死亡解除: 同じ編集導線で死亡記録を解除（生存へ戻す） | 保存成功。手順 2〜5 の各導線で対象ペットが再び選択可能になる（ガードは `status === "死亡"` / `deceased`） |

## 確認観点

- 生存判定はペットの `deceased_at` / ステータス列に基づく。フロントの無効化実装は `PatientSelectionTable.tsx`（予約・既定で死亡除外）と `PetSelectionResultsTable`（カルテ/会計/入院などの共通選択・`includeDeceased: true` で sentinel 表示＋「選択不可」）。
- ブロックの見え方は経路で異なる: **予約の既定検索**は死亡を結果に出さない。**共通ペット選択**は死亡を出し「グレーアウト＋選択不可」。**飼主 No 検索**で 0 件になる経路もあり、0 件でも異常ではない（2026-07-17 実測）。
- 死亡登録はサブダイアログの確定時に API 保存され、外側のペット編集フォームにも保存結果が同期される。死亡登録のために外側フォームの「更新」を重ねて押す必要はない（`PetDeceasedDialog.tsx`・`PetDeceasedRecordButton`）。generic `PATCH /pets/:id` では status を送らない（死亡/復活は `/:id/death` に一本化）。
- 死亡タグ除去のバックエンド処理は `HandlePetDeath`（`backend/internal/lstep/lstep_lifecycle_service.go`）。解除は `HandlePetRevival`。
- 死亡登録・解除の変更が `audit_logs` に記録されること（[specification.md §2.1](../../../spec/specification.md)）— DB 参照は USER 実施。
- 全操作が同一クリニック内で完結すること（clinic_id 隔離）。
- 転院（transferred）も Lステップ配信破棄の対象（[lstep-integration.md §4](../../../spec/line/lstep-integration.md)）だが、本シナリオでは死亡のみを扱う。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | 未来日の予約を持つペットを死亡登録する | 【要実測】既存予約の扱い（自動キャンセルされるか、残存して受付カンバンに出続けるか）。仕様文書に明記なし。**runtime 2026-08-01 BLOCKED**: 死亡登録は復元未証明のため submit せずキャンセル（清掃リスク） |
| A2 | 死亡登録済みペットの編集画面を再度開く | 死亡日・理由が保持されて表示される（[04-owners-form.md §1.2](../../../spec/screens/04-owners-form.md)「死亡時は日付と理由を記録」） |
| A3 | 死亡登録済みペットの過去のカルテ・会計履歴を参照する | **runtime PASS（2026-08-01 browser）**: 死亡ペットを含む飼主の会計履歴から過去会計詳細をオープン可能（閲覧ブロックなし。ブロック対象は新規データ作成 — [00-pet-selection.md](../../../spec/screens/00-pet-selection.md)） |

---

## 実装突合

- 突合日: 2026-08-07
- HEAD: `844e43f69`
- 変更サマリ:
  - 死亡登録導線を `PetDeceasedDialog` / `/:id/death` に明記（generic PATCH status ではない）
  - 予約既定検索は死亡除外、共通ペット選択は `includeDeceased: true` +「選択不可」と経路差を正確化
  - LSTEP は `HandlePetDeath` / `HandlePetRevival` をソースパス付きで記載

- runtime 2026-08-07: **BLOCKED** — authenticated UI requires E2E_LOGIN_* or non-empty DEV_ADMIN_* in host `.env.local` (currently empty). Stack healthy :3003/:8080.
