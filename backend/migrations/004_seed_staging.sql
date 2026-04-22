-- =============================================================================
-- Animal Ekarte - Staging Additions (clinic_id=4,5 exclusive data)
-- PostgreSQL 18
-- 依存: 001_init.sql → 002_seed_master.sql → 003_seed_demo.sql
-- 内容: clinic_id=4,5 新規追加クリニック用の最小限データセット
-- =============================================================================

-- Staging clinics: clinic_id=4,5 (master only)
-- デモ用permission_groups・rules は GORM CreateClinic が動的生成するため手動定義不要

-- NOTE: clinic, company, permission_groups, permission_group_rules, staff_clinic_assignments
-- などは 003_seed_demo.sql で clinic_id=1,2,3 のみ処理。clinic_id=4,5 の運用データ
-- が必要な場合は、以降のマイグレーション（005以降）で段階的に追加可能。
-- 現在の004は「最小限のデータセットで動作確認可能」な状態を目指す。
