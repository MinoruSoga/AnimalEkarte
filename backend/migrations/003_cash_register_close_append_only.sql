-- 003_cash_register_close_append_only.sql
-- W-013 FINAL B: レジ締め append-only 強化 + 締め後訂正（adjustment）テーブル。
-- 適用: USER が make migrate を手動実行すること（エージェントは auto-apply しない）。
-- CASCADE DELETE 禁止。001_init.sql は編集しない。
--
-- 方針:
--   1. cash_register_closes の soft-delete 再オープン経路を塞ぐ（partial UNIQUE → 完全 UNIQUE）
--   2. soft-deleted 行は active 行と衝突しなければ revive、衝突すれば migration を fail-closed
--   3. deleted_at 列は残すが app は soft-delete しない（UPDATE/DELETE は immutability trigger で拒否）
--   4. 締め後の会計訂正は cash_register_close_adjustments への append-only 追記で表現する

-- ---------------------------------------------------------------------------
-- 1. soft-deleted 行の整理（完全 UNIQUE 化の前提）
-- ---------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM cash_register_closes d
        WHERE d.deleted_at IS NOT NULL
          AND EXISTS (
              SELECT 1
              FROM cash_register_closes a
              WHERE a.deleted_at IS NULL
                AND a.clinic_id = d.clinic_id
                AND a.close_date = d.close_date
                AND a.period = d.period
          )
    ) THEN
        RAISE EXCEPTION
            '003_cash_register_close_append_only: soft-deleted cash_register_closes conflict with active rows for same (clinic_id, close_date, period); resolve manually before migrate';
    END IF;

    UPDATE cash_register_closes
    SET deleted_at = NULL
    WHERE deleted_at IS NOT NULL;
END $$;

-- ---------------------------------------------------------------------------
-- 2. partial UNIQUE → 完全 UNIQUE（soft-delete 再オープン不可）
-- ---------------------------------------------------------------------------
DROP INDEX IF EXISTS uq_cash_register_closes_date_period;

CREATE UNIQUE INDEX uq_cash_register_closes_date_period
    ON cash_register_closes (clinic_id, close_date, period);

COMMENT ON TABLE cash_register_closes IS
    'レジ締めレコード（FEAT-368）。append-only。更新・削除・soft-delete 再開は不可。締め後訂正は cash_register_close_adjustments へ追記する（W-013）。';

-- ---------------------------------------------------------------------------
-- 3. cash_register_close_adjustments（締め後訂正の append-only 台帳）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cash_register_close_adjustments (
    id                   BIGSERIAL    PRIMARY KEY,
    clinic_id            BIGINT       NOT NULL REFERENCES clinics(id),
    close_id             BIGINT       NOT NULL REFERENCES cash_register_closes(id),
    billing_id           BIGINT       NOT NULL,
    accounting_delta     BIGINT       NOT NULL DEFAULT 0,
    cash_movement_amount BIGINT       NOT NULL DEFAULT 0,
    reason               TEXT         NOT NULL,
    actor_id             BIGINT,
    executed_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cash_register_close_adjustments_close
    ON cash_register_close_adjustments (clinic_id, close_id);

CREATE INDEX IF NOT EXISTS idx_cash_register_close_adjustments_executed
    ON cash_register_close_adjustments (clinic_id, executed_at);

COMMENT ON TABLE cash_register_close_adjustments IS
    'レジ締め後の会計訂正台帳（W-013 append-only）。close 自体の reverse/取消は productize しない。';
COMMENT ON COLUMN cash_register_close_adjustments.accounting_delta IS
    '会計合計の増減（円）。取得できない場合は 0 でも reason は必須。';
COMMENT ON COLUMN cash_register_close_adjustments.cash_movement_amount IS
    '現金移動額（円）。会計のみの訂正では 0。';

-- ---------------------------------------------------------------------------
-- 4. immutability triggers（UPDATE/DELETE を DB 層で拒否）
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION app_private.prevent_cash_register_close_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'cash_register_closes is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

DROP TRIGGER IF EXISTS trg_cash_register_closes_immutable ON cash_register_closes;
CREATE TRIGGER trg_cash_register_closes_immutable
    BEFORE UPDATE OR DELETE ON cash_register_closes
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_cash_register_close_mutation();

CREATE OR REPLACE FUNCTION app_private.prevent_cash_register_close_adjustment_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'cash_register_close_adjustments is append-only; UPDATE/DELETE are not allowed'
        USING ERRCODE = 'check_violation';
END;
$$;

DROP TRIGGER IF EXISTS trg_cash_register_close_adjustments_immutable ON cash_register_close_adjustments;
CREATE TRIGGER trg_cash_register_close_adjustments_immutable
    BEFORE UPDATE OR DELETE ON cash_register_close_adjustments
    FOR EACH ROW
    EXECUTE FUNCTION app_private.prevent_cash_register_close_adjustment_mutation();
