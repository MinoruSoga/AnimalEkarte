-- EMR-306 レビュー指摘対応: migrations/CLAUDE.md の必須チェック「clinic_id を含む
-- 複合インデックスを追加する」に合わせ、025 で追加した単一列インデックスを複合化する。
-- 全 consumer（CountMergedMembers / ClearMergedInto / LockCageGroupForUpdate 等）が
-- clinic_id スコープで検索するため先頭に clinic_id を置く。
DROP INDEX IF EXISTS idx_cages_merged_into_cage_id;

CREATE INDEX idx_cages_merged_into_cage_id
    ON cages (clinic_id, merged_into_cage_id)
    WHERE merged_into_cage_id IS NOT NULL AND deleted_at IS NULL;
