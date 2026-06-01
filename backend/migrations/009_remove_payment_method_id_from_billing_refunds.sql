-- 009: billing_refunds から payment_method_id を除去する (#60 Phase 1 是正)。
--
-- 背景: 008 で billing_refunds に payment_method(ENUM) + payment_method_id(payment_methods
-- マスタ FK) を dual maintain で追加したが、会計の payment_splits は method(ENUM) 体系で
-- payment_method_id を持たない(会計UI PaymentCard が cash/credit_card/electronic_money の
-- ENUM 固定)。返金だけマスタ駆動にすると会計と突合できず、支払方法別の返金上限(Phase 2)が
-- 成立しない。返金を会計と同じ ENUM 体系に一本化するため payment_method_id を撤回する。
--
-- 008 は適用済みの可能性があるため編集せず、本 009 で打ち消す(DROP IF EXISTS で冪等)。
-- マスタ駆動への移行は会計UIの payment_methods マスタ化(FEAT-368 完遂)と一括で別途行う。

DROP INDEX IF EXISTS idx_billing_refunds_payment_method_id;
ALTER TABLE billing_refunds DROP COLUMN IF EXISTS payment_method_id;
