-- 005_add_appointment_checked_in_at.sql
-- 受付ヘッダー テレメトリ表示（change-ui.md Phase 2）: 待ち時間算出のための受付済み遷移時刻を記録する。
-- appointments.updated_at は autoUpdateTime のため予約編集全般でリセットされ待ち時間算出に流用できない。
-- そのため checked_in ステータスへ遷移した時点の時刻専用カラムを新設する。

ALTER TABLE appointments
    ADD COLUMN checked_in_at timestamptz;
