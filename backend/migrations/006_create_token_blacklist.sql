-- refresh_token のサーバーサイド失効を実現する JTI ブラックリストテーブル。
-- ログアウト時または不正検知時に JTI を登録し、RefreshToken エンドポイントで照合する。
-- expires_at を過ぎたエントリは定期バッチで物理削除してよい（有効期限切れ JTI は自然失効）。

CREATE TABLE token_blacklist (
    jti        TEXT        NOT NULL PRIMARY KEY,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE token_blacklist IS
    'ログアウト・失効済み refresh_token の JTI ブラックリスト';
COMMENT ON COLUMN token_blacklist.jti IS
    'JWT ID クレーム (uuid v4)。PRIMARY KEY なので一意性は保証される';
COMMENT ON COLUMN token_blacklist.expires_at IS
    '元 refresh_token の有効期限。これ以降は照合対象から除外してよい（バッチ削除の目安）';

-- 期限切れエントリを効率的に削除するためのインデックス
CREATE INDEX idx_token_blacklist_expires_at ON token_blacklist(expires_at);
