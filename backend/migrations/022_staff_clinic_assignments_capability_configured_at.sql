-- Migration: add durable marker for explicit reservation-capability configuration.
--
-- staff_reservation_capabilities は「行が存在する staff/clinic だけ設定済み」とみなせない。
-- PUT /v1/staffs/:id/reservation-capabilities で全区分を空選択した場合、
-- update_path が能力 junction を全削除するため行は残らないが、意味は
-- 「未設定」ではなく「明示的に全部位対応不可」。この区別を残すため
-- staff_clinic_assignments にタイムスタンプ marker を持たせ、
-- 後続の backfill(migration 023) が NULL(未設定)の院割当だけを対象にできるようにする。
--
-- 用途:
--   スタッフの予約受付可能区分を「未設定(=既定で全部位可)」から
--   「設定済み(=junction行を authoritative に解釈)」へ不可逆に昇格させるための
--   書込み経路 marker。 backfill は NULL 行のみ触る。
--
-- Do not edit 001_init.sql. User must run: make migrate

ALTER TABLE staff_clinic_assignments
    ADD COLUMN capabilities_configured_at TIMESTAMPTZ;

COMMENT ON COLUMN staff_clinic_assignments.capabilities_configured_at IS
    'Timestamp when reservation capabilities were explicitly configured via PUT /v1/staffs/:id/reservation-capabilities (including an explicit empty selection). NULL means never configured: backfill seeds capabilities from clinic defaults. Set only by the write path; never cleared.';
