-- LSTEP-BE-021: LINE User ID 紐付けトークンテーブル
-- 受付スタッフが生成した一時トークンを QR コード化して飼い主に提示し、
-- LIFF アプリ経由で LINE User ID を owners に紐付けるために使用する。

CREATE TABLE line_link_tokens (
    id          BIGSERIAL   PRIMARY KEY,
    clinic_id   BIGINT      NOT NULL REFERENCES clinics(id),
    owner_id    BIGINT      NOT NULL REFERENCES owners(id),
    token       VARCHAR(64) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- トークン検索用（未使用のみを高速検索）
CREATE INDEX idx_line_link_tokens_token ON line_link_tokens (token)
    WHERE used_at IS NULL;

-- 飼い主ごとの発行履歴確認用
CREATE INDEX idx_line_link_tokens_owner ON line_link_tokens (clinic_id, owner_id);

COMMENT ON TABLE line_link_tokens IS 'LINE User ID 紐付け用の一時トークン（24時間有効、1回限り使用）。';
