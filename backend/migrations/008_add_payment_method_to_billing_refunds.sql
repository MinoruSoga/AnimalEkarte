-- 008: billing_refunds に支払方法別返金の記録カラムを追加する。
--
-- 背景 (issue #60): payment_splits 導入で1会計に現金/カード/電子マネー混在支払いが
-- 可能になったが、billing_refunds は billing 単位の単一金額のみで「どの支払方法へ
-- 返金したか」を記録できなかった。混在支払いの返金追跡（帳票精度）のため記録カラムを追加する。
--
-- payments / payment_splits と同じ dual maintain パターンに揃える:
--   payment_method    : payment_method ENUM (cash / credit_card / electronic_money)
--   payment_method_id : payment_methods マスタ (FEAT-368) への FK
-- どちらも nullable。既存返金行・単一支払い会計の返金は支払方法なしで成立する（後方互換）。
-- 支払方法別の返金上限バリデーションは Phase 2。本 migration は記録のみ。

ALTER TABLE billing_refunds ADD COLUMN payment_method payment_method;
ALTER TABLE billing_refunds ADD COLUMN payment_method_id bigint REFERENCES payment_methods(id);

-- Phase 2 の支払方法別返金集計に備え、設定済み行のみの部分インデックスを付与する。
CREATE INDEX idx_billing_refunds_payment_method_id
    ON billing_refunds (payment_method_id)
    WHERE payment_method_id IS NOT NULL;
