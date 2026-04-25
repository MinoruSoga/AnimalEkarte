-- LSTEP-BE-002: owners に LINE User ID カラムを追加する
-- 飼主とLINEアカウントを紐づけることで、Lステップタグ同期・LINE通知送信を可能にする。
-- 同一クリニック内で LINE User ID の重複は許可しない（UNIQUE制約）。
-- NULL は「LINEアカウント未連携」を表す（任意連携）。

ALTER TABLE owners
    ADD COLUMN line_user_id text;

-- 同一クリニック内で line_user_id の重複を防ぐ
-- NULL は一意性制約の対象外（複数の未連携飼主を許容する）
CREATE UNIQUE INDEX uk_owners_clinic_line_user_id
    ON owners(clinic_id, line_user_id)
    WHERE line_user_id IS NOT NULL AND deleted_at IS NULL;

-- 検索用インデックス（line_user_id 単体での参照も想定）
CREATE INDEX idx_owners_line_user_id
    ON owners(line_user_id)
    WHERE line_user_id IS NOT NULL AND deleted_at IS NULL;

COMMENT ON COLUMN owners.line_user_id IS 'LINE User ID（Lステップ連携・LINE通知用）。NULL = 未連携。';
