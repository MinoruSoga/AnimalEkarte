-- LSTEP-BE-021: owners に LINE フォロー・ブロック日時カラムを追加する
-- フォローイベントで line_followed_at を、ブロック(unfollow)イベントで line_blocked_at を更新する。

ALTER TABLE owners
    ADD COLUMN IF NOT EXISTS line_followed_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS line_blocked_at   TIMESTAMPTZ;

COMMENT ON COLUMN owners.line_followed_at IS 'LINE フォロー日時（最終フォロー時刻）。Webhook follow イベントで更新。';
COMMENT ON COLUMN owners.line_blocked_at  IS 'LINE ブロック日時。Webhook unfollow イベントで更新。再フォロー時に NULL にリセット。';
