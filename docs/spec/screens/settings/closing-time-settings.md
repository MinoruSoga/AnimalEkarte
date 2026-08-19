# 締め時間設定 仕様書 (Closing Time Settings)

## 概要
- **画面の目的**: 日次および月次の売上集計に使用する境界時刻、休診日、および特別な集計期間の管理。
- **URLパターン**: `/settings/closing-time`
- **アクセス権限**: 締め時間設定管理権限が必要（`ResourceClosingSettings`）

---

## 画面構成

### 1. 標準境界時刻の定義
日常的な営業における「一日の区切り」を設定します。集計レンジは AM / PM / EMG（緊急）の 3 区分です（#215）。
- **AM 開始時刻**: AM レンジの開始時刻（`closing_am_start`、固定 09:00）。**編集不可**（PATCH API・画面のいずれにも入力フィールドがなく、DB default から変更する経路が存在しない。レンジプレビューの算出にのみ使用）。
- **AM/PM 境界**: 午前診療と午後診療を分かつ時刻（デフォルト 14:00）。
- **終了時刻 (Weekday/Sunday)**: PM レンジの終端（デフォルト平日 18:30 / 日曜 17:30）。これを過ぎた会計は「翌日の売上」ではなく、**当日の EMG（緊急）レンジ** `[終了時刻, 翌日の AM 開始時刻)` に計上されます（深夜 0:00〜AM 開始の会計は前日 EMG 帰属）。
- **レンジプレビュー**: 設定値から算出した AM/PM/EMG の実レンジが画面上に表示されます（#151）。

### 2. 定例休診日の管理
曜日ごとの休診フラグ（`closed_weekdays`）を設定します。`StandardClosingTimeSection` 内のチェックボックス群として実装されており、独立したセクションではありません。

### 3. 特別期間 (Special Periods)
年末年始、短縮営業などの例外的な集計境界時刻を日付範囲で登録します（`SpecialPeriodSection`）。

### 4. 個別休診日 (Holidays)
特定の1日を休診日として登録します（日付＋理由メモ、`HolidaySection`）。定例休診日（曜日単位）や特別期間（境界時刻の上書き）とは独立した、日付単位の登録です。`POST` は INSERT のみで同一日を UPSERT しない。既存日の再追加は 409（`uk_clinic_holidays_clinic_date`）。理由変更は削除してから再追加する。

---

## 主要な機能

### 1. 売上計上の自動判定
システムは、会計が完了（`completed_at`）した時刻と、本マスタの設定値を照合し、「何日の、AM/PM/EMG いずれの売上か」を自動的に決定します（AM/PM/EMG は連続・非重複で 24 時間を被覆）。これにより、深夜に及ぶ緊急診療等でも正確な日次集計が可能です。

### 2. 予約枠計算とは非連動
本画面の休診設定（定例休診日 `closed_weekdays`・特別期間・日付単位の休診日）はいずれも `clinic_settings` / `closing_special_periods` / `clinic_holidays` テーブルに保存され、**売上集計専用**です。LINE 予約の空き枠計算エンジン（[28-line-reservation.md](../28-line-reservation.md)）は完全に別モデルの `line_reservation_settings.closed_weekdays`（LINE 予約設定画面で個別設定）を参照しており、本画面の設定とは連動しません。曜日休診をLINE予約にも反映したい場合は、LINE 予約設定側でも別途設定する必要があります。

---

## 技術仕様

### 使用コンポーネント
- **`StandardClosingTimeSection`**: 境界時刻の編集（ネイティブ `<input type="time">`、分単位）とレンジプレビュー、および定例休診日（曜日チェックボックス）。
- **`SpecialPeriodSection`**: 登録済み特別期間（境界時刻の日付範囲上書き）の管理。
- **`HolidaySection`**: 個別休診日（日付単位）の管理。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/closing-settings` | 現在の時刻・休診設定の取得 | `closing-settings` | `view` |
| PATCH | `/api/v1/closing-settings` | 境界時刻の更新 | `closing-settings` | `edit` |
| GET | `/api/v1/closing-settings/special-periods` | 登録済み例外期間の一覧取得 | `closing-settings` | `view` |
| POST | `/api/v1/closing-settings/special-periods` | 新規例外期間の登録 | `closing-settings` | `create` |
| PATCH | `/api/v1/closing-settings/special-periods/:id` | 例外期間の更新（BE実装済みだが編集UIは未実装のため未呼出。変更は削除+再登録で運用） | `closing-settings` | `edit` |
| DELETE | `/api/v1/closing-settings/special-periods/:id` | 例外期間の削除 | `closing-settings` | `delete` |
| GET | `/api/v1/closing-settings/holidays` | 休診日一覧の取得 | `closing-settings` | `view` |
| POST | `/api/v1/closing-settings/holidays` | 休診日の登録 | `closing-settings` | `create` |
| DELETE | `/api/v1/closing-settings/holidays/:date` | 休診日の削除 | `closing-settings` | `delete` |

---

