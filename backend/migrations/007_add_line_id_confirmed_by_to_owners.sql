-- Q22: LINE ID 紐付け確認者を記録するカラムを owners テーブルに追加する。
-- ON DELETE SET NULL: 担当スタッフが削除されても所有者レコードは保持する。
ALTER TABLE owners
    ADD COLUMN line_id_confirmed_by BIGINT NULL
        REFERENCES staffs(id) ON DELETE SET NULL;

CREATE INDEX idx_owners_line_id_confirmed_by
    ON owners (line_id_confirmed_by)
    WHERE line_id_confirmed_by IS NOT NULL;
