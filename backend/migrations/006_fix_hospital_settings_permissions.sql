-- =============================================================================
-- 006_fix_hospital_settings_permissions.sql
-- -----------------------------------------------------------------------------
-- BUG-377 / BUG-378: 医院マスタ (hospital-settings) の作成・削除権限を
-- permission_group_rules から剥奪する。
--
-- 目的: Frontend の usePermission() と Backend の RequirePermission() を
--       単一の真実とし、is_system_admin 二重ガードを廃止する。
--       is_system_admin=true は hasPermission() でバイパスされるため、
--       能力は維持される（admin@noavet.jp 等は従来通り作成・削除可）。
-- -----------------------------------------------------------------------------
UPDATE permission_group_rules
   SET can_create = false,
       can_delete = false
 WHERE resource = 'hospital-settings';
