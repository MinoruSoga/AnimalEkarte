-- BE-014: ノーショウ検知 — reservation_status ENUM に no_show を追加
ALTER TYPE reservation_status ADD VALUE IF NOT EXISTS 'no_show';
