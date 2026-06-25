-- 009_add_payment_method_system_key.sql
-- Issue #197: payment_methods に system_key（安定識別子）を追加して rename 耐性を持たせる。
--
-- 背景: クリニックが「現金」等の既定 name を PATCH /v1/payment-methods/:id で改名できる。
-- 現状の name 文字列依存実装では:
--   - 書込み側: resolvePaymentMethodMasterID が「支払方法マスタ「現金」が見つかりません」で会計確定を停止
--   - 集計側: findCashMethodID が現金を見失いレジ締め・月次集計で現金を過少計上（データ品質バグ）
-- system_key を真実の源泉とすることで rename 後も正しく動作する。

-- 1. カラム追加
ALTER TABLE payment_methods ADD COLUMN system_key varchar(50);

-- 2. 既存行の backfill（標準4種）
UPDATE payment_methods SET system_key = 'cash'             WHERE name = '現金';
UPDATE payment_methods SET system_key = 'credit_card'      WHERE name = 'クレジットカード';
UPDATE payment_methods SET system_key = 'electronic_money' WHERE name = '電子マネー';
UPDATE payment_methods SET system_key = 'bank_transfer'    WHERE name = '銀行振込';

-- 3. 新規クリニック作成時のデフォルト支払方法に system_key をセット
--    (CREATE OR REPLACE: 001_init.sql のオリジナル定義を上書きする。既存トリガーはそのまま使用)
CREATE OR REPLACE FUNCTION create_default_payment_methods()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO payment_methods (clinic_id, name, system_key, display_order, is_active)
    VALUES
        (NEW.id, '現金',            'cash',             1, true),
        (NEW.id, 'クレジットカード', 'credit_card',      2, true),
        (NEW.id, '電子マネー',       'electronic_money', 3, true),
        (NEW.id, '銀行振込',         'bank_transfer',    4, true);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 4. 部分 UNIQUE INDEX（非標準 NULL 行は対象外）
CREATE UNIQUE INDEX idx_payment_methods_clinic_system_key
    ON payment_methods (clinic_id, system_key)
    WHERE system_key IS NOT NULL AND deleted_at IS NULL;
