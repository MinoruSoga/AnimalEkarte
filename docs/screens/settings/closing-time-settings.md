# 締め時間設定 仕様書 (Closing Time Settings)

## 概要
- **画面の目的**: 日次および月次の売上集計に使用する境界時刻、休診日、および特別な集計期間の管理。
- **URLパターン**: `/settings/closing-time`
- **アクセス権限**: 締め時間設定管理権限が必要（`ResourceClosingSettings`）

---

## 画面構成

### 1. 標準境界時刻の定義
日常的な営業における「一日の区切り」を設定します。集計レンジは AM / PM / EMG（緊急）の 3 区分です（#215）。
- **AM 開始時刻**: AM レンジの開始時刻（`closing_am_start`、デフォルト 09:00）。
- **AM/PM 境界**: 午前診療と午後診療を分かつ時刻（デフォルト 14:00）。
- **終了時刻 (Weekday/Sunday)**: PM レンジの終端（デフォルト平日 18:30 / 日曜 17:30）。これを過ぎた会計は「翌日の売上」ではなく、**当日の EMG（緊急）レンジ** `[終了時刻, 翌日の AM 開始時刻)` に計上されます（深夜 0:00〜AM 開始の会計は前日 EMG 帰属）。
- **レンジプレビュー**: 設定値から算出した AM/PM/EMG の実レンジが画面上に表示されます（#151）。

### 2. 定例休診日の管理
曜日ごとの休診フラグを設定。カレンダーや予約システム上の「非稼働日」のデフォルトとなります。

### 3. 特別期間 (Special Periods)
年末年始、臨時休診、短縮営業などの例外的な営業時間を日付範囲で登録します。

---

## 主要な機能

### 1. 売上計上の自動判定
システムは、会計が完了（`completed_at`）した時刻と、本マスタの設定値を照合し、「何日の、AM/PM/EMG いずれの売上か」を自動的に決定します（AM/PM/EMG は連続・非重複で 24 時間を被覆）。これにより、深夜に及ぶ緊急診療等でも正確な日次集計が可能です。

### 2. 予約枠計算への波及
定例休診日（曜日単位の `closed_weekdays`）は LINE 予約システムの空き枠計算エンジンに反映され、該当曜日は予約不可となります（祝日除外は LINE 予約設定側の祝日フラグで制御）。一方、特別期間（短縮営業等）と日付単位の休診日登録は集計専用であり、空き枠計算には反映されません。

---

## 技術仕様

### 使用コンポーネント
- **`StandardClosingTimeSection`**: 境界時刻の編集（ネイティブ `<input type="time">`、分単位）とレンジプレビュー。
- **`SpecialPeriodSection`**: 登録済み例外期間の管理。
- **`HolidaySection`**: 定例休診日の管理。

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

