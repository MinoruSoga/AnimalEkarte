-- 005_add_billing_item_discount.sql
-- #85: 会計明細(billing_items)に項目別割引(割引率・割引額)を追加する。
-- treatments と同型(discount_rate numeric(5,2) / discount_amount bigint)。
-- 後方互換のため NOT NULL DEFAULT 0。
-- 合計・税計算への反映(recalculateTotals / FE 合計計算 / CalculateTaxAmount)は
-- 税計算順序の確定後に別段階で対応する。本 migration はフィールド追加のみ。

ALTER TABLE billing_items
    ADD COLUMN discount_rate   numeric(5,2) NOT NULL DEFAULT 0,
    ADD COLUMN discount_amount bigint       NOT NULL DEFAULT 0;

COMMENT ON COLUMN billing_items.discount_rate   IS '#85: 項目別割引率(%)';
COMMENT ON COLUMN billing_items.discount_amount IS '#85: 項目別割引額(円)';
