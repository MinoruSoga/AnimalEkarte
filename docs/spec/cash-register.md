# レジ締め・売上集計業務 仕様書 (Cash Register Operations)

> **目的**: レジ締め・日次/月次売上集計の業務仕様(AM/PM/EMG境界含む)を定義する。
> **読者**: 会計機能の実装者・経理担当。
> **タイミング**: レジ締めロジックの実装・仕様確認時。

> **Animal Ekarte**: 正確な現金管理と透明性の高い会計フロー
> **最新更新**: 2026-09-06 | **ステータス**: repo の集計・append-only 契約を照合。実環境の受入結果は [UAT 集計](../ops/testing/UAT-DOMAIN-STATUS.md) を参照する。

画面操作（締め実行・履歴の既定期間・区分フィルタのスコープ）は [screens/29-closing-aggregation.md](screens/29-closing-aggregation.md) を正本とする。履歴 UI の既定は JST 当月。

---

## 1. 業務の定義

本システムにおける「締め」とは、一定期間の売上実績を確定させ、物理的な現金（レジ内）とシステム上の理論値を照合するプロセスを指します。

| 単位 | 名称 | 説明 |
|:---|:---|:---|
| **シフト単位** | AM締め / PM締め / EMG (夜間・緊急) 締め | 各時間帯（AM/PM/EMG）終了時の現金の受け渡し・売上確定。 |
| **日次単位** | 日計 | 一日の最終的な売上確定と、翌日への繰り越し準備。 |
| **月次単位** | 月計 | 一ヶ月の総売上、客単価、支払方法の内訳分析。 |

---

## 2. 集計の境界 (Closing Boundaries)

売上を「どの日・どのシフト」に計上するかは、締め時間設定およびクリニック設定の AM開始時刻 (`closing_am_start`, 既定 `09:00`) に基づきます。
以下の時刻例は計算説明用である。全院投入の裁定値は [#252](https://github.com/MinoruSoga/AnimalEkarte/issues/252) の AM 開始 09:00・AM/PM 境界 12:00・平日/日曜終了 18:30（[納品パッケージ](../delivery/DELIVERY_PACKAGE.md)）を参照し、例を投入値として流用しない。
- **シフトの区分境界**:
  - **AM**: `[am_start, pm_start)` (例: `09:00`〜`14:00`)
  - **PM**: `[pm_start, pm_end)` (例: `14:00`〜`19:00`)
  - **EMG (夜間・緊急)**: `[pm_end, 翌日 am_start)` (例: `19:00`〜`翌日 09:00`)
- **日付の帰属ルール**: 深夜 `00:00` から `am_start` (例: `09:00`) までの時間帯に行われた緊急会計は、カレンダー上の日付とは異なり、業務上は**前日の EMG (夜間・緊急)**に帰属します。これにより深夜診療の売上集計を同一営業日として扱います。
- **例** (AM境界 `09:00`, PM境界 `14:00`, 夜間開始 `19:00`):
  - 13:59 の会計 ➔ 当日の **AM売上**
  - 14:01 の会計 ➔ 当日の **PM売上**
  - 20:00 の会計 ➔ 当日の **EMG売上**
  - 翌日 02:00 の会計 ➔ 当日（前日扱い）の **EMG売上**

---

## 3. レジ締めフロー

1.  **プレビュー**: 指定期間（AM/PM/EMG）内の `completed_at` を持つ完了会計（`status = completed`、未削除）を集計し、理論上の売上高を表示。下書き・未確定会計は集計しない。
2.  **実査**: レジ内の現金を実際に数え、実査金額の合計を入力（金種別の内訳入力欄はない）。
3.  **過不足確認**: 
    - `現金決済の会計金額合計 ー 返金合計 ＝ 理論現金残高`(前回繰越残高・入金・支出の管理項目は現状なし)
    - 理論残高と実際の現金を比較し、差異（過不足）を算出。
4.  **確定とロック**: 「締め実行」により記録を保存。レジ締めレコードは **append-only** で、登録済み route には更新・削除・soft-delete 再開・巻き戻し（reverse）API がない（W-013）。
5.  **締め後の会計訂正**: 会計データ自体は `accounting-post-close-edit` 権限を持つスタッフであれば締め済み期間でも編集可能。編集時は理由必須で、(a) 監査ログ（#115）と (b) `cash_register_close_adjustments` への追記を **同一 transaction** で fail-closed に記録する。締めレコード自体を取り消して「開き直す」運用は productize しない。

---

## 4. 帳票と分析

- **日次レジ締め報告書**: 担当者名、支払方法・カテゴリ別内訳、過不足、売上構成を A4 形式で出力。
- **支払方法別レポート**: 現金、クレジットカード、電子マネーの利用比率を可視化。
- **部門 × 支払方法**: [#247](https://github.com/MinoruSoga/AnimalEkarte/issues/247) の DEC-16⑥に従い、会計単位で支払実額を配賦する。最大剰余法で行・列・総額を円単位で保存し、件数は会計 distinct。支払列は医院マスタから生成し、期間内実績のある無効・履歴 method も残す。月次・締めは `backend/internal/billing/allocation.go` の共通契約を使用する。

---

## 5. 技術仕様

- **整合性の担保（append-only）**:
  - 締め確定は集計スナップショットの **Create のみ**。登録済み API route は Create / List / Get / Preview のみで、取消・再開・巻き戻しの実装および公開契約はない。
  - `(clinic_id, close_date, period)` は **完全 UNIQUE**（`deleted_at` を見ない）。soft-delete で同じ区分を再締めする経路は塞がれている。
  - DB 層でも `cash_register_closes` / `cash_register_close_adjustments` の UPDATE/DELETE を immutability trigger で拒否する（`backend/migrations/001_init.sql` の append-only 統合ブロック。コメント上の旧 migration 003）。
- **締め後訂正モデル**: 会計編集の差分は `cash_register_close_adjustments`（`close_id` 参照・NO CASCADE DELETE）へ append-only 追記。`accounting_delta` は合計変更が分かる場合の best-effort 差分、会計のみの訂正では `cash_movement_amount=0`。
- **権限**: `cash-register-close` 権限(`view`/`create`)はクリニック単位で付与され、この権限を持つスタッフは同一クリニックの全締め記録を閲覧・作成できる(担当者本人の記録に限定する制御はない)。確定済み close 自体は権限の有無にかかわらず誰も修正・取消できない。締め後会計編集は `accounting-post-close-edit`。

---
