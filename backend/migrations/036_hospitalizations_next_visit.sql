-- EMR-353: 入院の次回予定日。カレンダー情報のみ（予約意図・LSTEP 連携なし）。
ALTER TABLE hospitalizations
    ADD COLUMN IF NOT EXISTS next_visit DATE;
