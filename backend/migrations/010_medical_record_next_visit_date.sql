-- BE-006: 次回来院推奨日フィールドを medical_records に追加
ALTER TABLE medical_records
  ADD COLUMN next_visit_recommended_date DATE NULL;
