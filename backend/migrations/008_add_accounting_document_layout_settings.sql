-- 008_add_accounting_document_layout_settings.sql
-- Issue #179 follow-up ①: 明細兼領収書の病院単位表示設定を追加する。
--
-- 帳票設定は病院マスタ（clinics）に属する表示方針であり、
-- 既存の AccountingDocument は /me の user.clinic を正本として描画している。
-- そのため clinics に最小表示設定を追加し、既存帳票はデフォルト挙動を維持する。

ALTER TABLE clinics
    ADD COLUMN IF NOT EXISTS accounting_document_show_logo boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS accounting_document_show_registration_warning boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS accounting_document_show_item_category boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS accounting_document_footer_note text NOT NULL DEFAULT '';

COMMENT ON COLUMN clinics.accounting_document_show_logo IS '明細兼領収書に病院ロゴを表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_registration_warning IS '登録番号未設定時の帳票警告を表示するか。';
COMMENT ON COLUMN clinics.accounting_document_show_item_category IS '明細兼領収書の項目カテゴリ行を表示するか。';
COMMENT ON COLUMN clinics.accounting_document_footer_note IS '明細兼領収書のフッター備考文言。';
