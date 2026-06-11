-- Feature 1: 親子階層
ALTER TABLE reservation_types
  ADD COLUMN parent_id bigint REFERENCES reservation_types(id) ON DELETE RESTRICT;

COMMENT ON COLUMN reservation_types.parent_id IS
  '親予約区分ID。NULL はルートノード（トップレベル）';

CREATE INDEX idx_reservation_types_parent
  ON reservation_types(parent_id)
  WHERE parent_id IS NOT NULL AND deleted_at IS NULL;

-- Feature 2: 同時受付上限
ALTER TABLE reservation_types
  ADD COLUMN max_concurrent integer CHECK (max_concurrent > 0);

COMMENT ON COLUMN reservation_types.max_concurrent IS
  '同一開始時刻の最大同時受付件数。NULL は無制限（従来動作）';
