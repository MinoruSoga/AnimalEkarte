-- support_chat_messages: ヘルプチャットの会話履歴（clinic_id × staff_id で個人スコープ）。
-- POST /support/chat 成功時に「質問+回答」のペアを保存し、GET/DELETE /support/chat/history で
-- 参照・リセットする。操作サポート用途だが入力内容は任意テキストのため個人情報を含み得る。
-- Do not edit applied migrations. User must run: make migrate

CREATE TABLE support_chat_messages (
    id         BIGSERIAL   PRIMARY KEY,
    clinic_id  bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    staff_id   bigint      NOT NULL REFERENCES staffs(id)  ON DELETE RESTRICT,
    role       varchar(16) NOT NULL CHECK (role IN ('user', 'assistant')),
    content    text        NOT NULL,
    sources    jsonb,        -- assistant ターンの参照記事 [{title,category,slug}]（user ターンは NULL）
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_support_chat_messages_clinic_staff_created
    ON support_chat_messages (clinic_id, staff_id, created_at, id)
    WHERE deleted_at IS NULL;

COMMENT ON TABLE support_chat_messages IS 'ヘルプチャットの会話履歴（認証スタッフ個人 × 選択clinic単位）';
COMMENT ON COLUMN support_chat_messages.sources IS 'assistant 回答の根拠となったマニュアル記事のスナップショット';
