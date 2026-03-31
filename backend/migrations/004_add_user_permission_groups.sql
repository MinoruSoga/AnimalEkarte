-- 004_add_user_permission_groups.sql
-- user_permission_groups テーブルを追加（001_init.sql に追加されたが適用済みのためスキップされた）

CREATE TABLE IF NOT EXISTS user_permission_groups (
    user_id  bigint NOT NULL REFERENCES user_accounts(id) ON DELETE CASCADE,
    group_id bigint NOT NULL REFERENCES permission_groups(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_user_permission_groups_user ON user_permission_groups(user_id);

COMMENT ON TABLE user_permission_groups IS 'ユーザー→権限グループ紐付け';
