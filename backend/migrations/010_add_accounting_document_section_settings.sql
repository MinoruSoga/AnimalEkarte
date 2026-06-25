-- 010_add_accounting_document_section_settings.sql
-- Issue #190: 明細兼領収書の帳票セクション表示トグルと表示順を病院単位で設定できるようにする。
-- migration 008 で追加したロゴ/警告/カテゴリ/フッター設定の続きとして clinics テーブルを拡張する。
--
-- 設計方針:
--   - セクション単位の表示/非表示は個別の boolean カラムで保持（NULL 不可・DEFAULT true）。
--   - 表示順は text[] として保持。空配列 '{}' はデフォルト順（自然順）を意味する。
--   - ADD COLUMN IF NOT EXISTS で冪等。既存行はすべて DEFAULT 値が適用される。

ALTER TABLE clinics
    ADD COLUMN IF NOT EXISTS accounting_document_show_clinic_header   boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS accounting_document_show_owner_pet_info  boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS accounting_document_show_items_table     boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS accounting_document_show_payment_summary boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS accounting_document_section_order        text[]  NOT NULL DEFAULT '{}';

COMMENT ON COLUMN clinics.accounting_document_show_clinic_header   IS '明細兼領収書に病院ヘッダーセクションを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_owner_pet_info  IS '明細兼領収書に飼主・ペット情報セクションを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_items_table     IS '明細兼領収書に明細表セクションを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_payment_summary IS '明細兼領収書に合計・支払セクションを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_section_order        IS '明細兼領収書のセクション表示順（空配列=デフォルト順）。';
