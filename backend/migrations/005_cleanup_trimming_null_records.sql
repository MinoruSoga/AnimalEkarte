-- BUG-TRIMMING-ZERO-DATE-ROWS: pet_id が NULL のトリミング予約レコードを削除
-- appointment_trimming_details は ON DELETE CASCADE で自動削除される
DELETE FROM appointments
WHERE pet_id IS NULL
  AND reservation_type_id IN (
    SELECT id FROM reservation_types WHERE category = 'trimming'
  );
