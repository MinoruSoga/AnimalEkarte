-- 飼主番号カラム追加
-- 人間が読みやすい連番形式の飼主識別子
-- 30001から開始して連番を割り当て

-- シーケンス作成（30001から開始）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = 'owner_number_seq') THEN
        CREATE SEQUENCE owner_number_seq START WITH 30001;
    END IF;
END $$;

-- 飼主番号カラム追加
ALTER TABLE owners ADD COLUMN IF NOT EXISTS owner_number INTEGER UNIQUE;

-- 既存データに連番を割り当て（まだ割り当てられていないもの）
UPDATE owners SET owner_number = nextval('owner_number_seq') WHERE owner_number IS NULL;

-- NOT NULL制約を追加
ALTER TABLE owners ALTER COLUMN owner_number SET NOT NULL;

-- デフォルト値を設定（新規作成時に自動採番）
ALTER TABLE owners ALTER COLUMN owner_number SET DEFAULT nextval('owner_number_seq');

-- インデックス作成
CREATE INDEX IF NOT EXISTS idx_owner_number ON owners(owner_number);

-- name_kana カラム追加（モデルと同期）
ALTER TABLE owners ADD COLUMN IF NOT EXISTS name_kana VARCHAR(100);

-- notes カラム追加（モデルと同期）
ALTER TABLE owners ADD COLUMN IF NOT EXISTS notes TEXT;
