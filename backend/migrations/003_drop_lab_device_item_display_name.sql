-- Fable MASTER M2: display_name は exam_type_fields.name と二重のため削除する。
-- 表示は未対応=電文コード、対応済=exam_type_fields.name。

ALTER TABLE lab_device_item_masters
  DROP COLUMN IF EXISTS display_name;
