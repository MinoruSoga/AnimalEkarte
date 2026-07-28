-- TASK-ADR003: payments / payment_splits の method ⇔ payment_methods.system_key 一致を DB 境界で fail-closed にする。
--
-- Scope:
--   1. app_private.enforce_payment_method_system_key_match() を追加
--   2. payment_splits / payments に BEFORE INSERT OR UPDATE トリガーを付与
--
-- Rules:
--   - payment_method_id IS NULL の行はレガシー互換として許可（MATCH SIMPLE 相当）
--   - payment_method_id が非 NULL のとき、payment_methods に
--       id = payment_method_id AND system_key = method::text AND clinic_id = NEW.clinic_id
--     の行が存在することを要求する
--   - 削除時の連鎖削除制約は導入しない（RESTRICT / 既定のみ）
--
-- 前提: 005 により payments.clinic_id が存在する。
-- 既存の不整合行を洗い出す場合（適用前の任意確認）:
--
-- SELECT p.id, p.clinic_id, p.method, p.payment_method_id, pm.system_key, pm.clinic_id AS pm_clinic_id
-- FROM payments p
-- LEFT JOIN payment_methods pm ON pm.id = p.payment_method_id
-- WHERE p.payment_method_id IS NOT NULL
--   AND (
--     pm.id IS NULL
--     OR pm.system_key IS DISTINCT FROM p.method::text
--     OR pm.clinic_id IS DISTINCT FROM p.clinic_id
--   );
--
-- SELECT s.id, s.clinic_id, s.method, s.payment_method_id, pm.system_key, pm.clinic_id AS pm_clinic_id
-- FROM payment_splits s
-- LEFT JOIN payment_methods pm ON pm.id = s.payment_method_id
-- WHERE s.payment_method_id IS NOT NULL
--   AND (
--     pm.id IS NULL
--     OR pm.system_key IS DISTINCT FROM s.method::text
--     OR pm.clinic_id IS DISTINCT FROM s.clinic_id
--   );

CREATE SCHEMA IF NOT EXISTS app_private;

CREATE OR REPLACE FUNCTION app_private.enforce_payment_method_system_key_match()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- レガシー行: payment_method_id 未設定は method ENUM のみで運用する（MATCH SIMPLE 相当）。
    IF NEW.payment_method_id IS NULL THEN
        RETURN NEW;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM payment_methods AS pm
        WHERE pm.id = NEW.payment_method_id
          AND pm.system_key IS NOT DISTINCT FROM NEW.method::text
          AND pm.clinic_id = NEW.clinic_id
    ) THEN
        RAISE EXCEPTION
            '支払方法の不整合: payment_method_id=% の system_key が method=% と一致しません (clinic_id=%)',
            NEW.payment_method_id,
            NEW.method,
            NEW.clinic_id
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

COMMENT ON FUNCTION app_private.enforce_payment_method_system_key_match() IS
    'TASK-ADR003: payments/payment_splits の method と payment_methods.system_key の一致を fail-closed で強制する';

DROP TRIGGER IF EXISTS trg_payment_splits_method_system_key_match ON payment_splits;
CREATE TRIGGER trg_payment_splits_method_system_key_match
    BEFORE INSERT OR UPDATE ON payment_splits
    FOR EACH ROW
    EXECUTE FUNCTION app_private.enforce_payment_method_system_key_match();

DROP TRIGGER IF EXISTS trg_payments_method_system_key_match ON payments;
CREATE TRIGGER trg_payments_method_system_key_match
    BEFORE INSERT OR UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION app_private.enforce_payment_method_system_key_match();

GRANT USAGE ON SCHEMA app_private TO PUBLIC;
GRANT EXECUTE ON FUNCTION app_private.enforce_payment_method_system_key_match() TO PUBLIC;
