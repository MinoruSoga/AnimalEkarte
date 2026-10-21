package pet

import (
	"fmt"

	"gorm.io/gorm"
)

// JoinCurrentCommonBasicName はローカル pet（petAlias）から common_pet_clinic_bindings →
// common_pet_subjects → common_pet_basic_revisions（現行 revision）への JOIN を追加する
// EMR-491 の「現在の正式名 reader」部分。binding は immutable で is_active /
// mapping_version 等の旧列を持たない（001_init.sql で廃止済み・参照しない）。
// revisionAlias は subject.current_basic_revision_id と同じ common pet に属する行だけに
// 絞る（rev.common_pet_id = subject.id の同体性を ON 句で保証する）。
// petAlias の JOIN は本関数の前に行うこと（参照解決のため）。
func JoinCurrentCommonBasicName(tx *gorm.DB, petAlias, revisionAlias string) *gorm.DB {
	bindingAlias := petAlias + "__cpcb"
	subjectAlias := petAlias + "__cps"
	return tx.
		Joins(fmt.Sprintf(
			"JOIN common_pet_clinic_bindings %s ON %s.clinic_id = %s.clinic_id AND %s.local_pet_id = %s.id",
			bindingAlias, bindingAlias, petAlias, bindingAlias, petAlias)).
		Joins(fmt.Sprintf(
			"JOIN common_pet_subjects %s ON %s.id = %s.common_pet_id",
			subjectAlias, subjectAlias, bindingAlias)).
		Joins(fmt.Sprintf(
			"JOIN common_pet_basic_revisions %s ON %s.id = %s.current_basic_revision_id AND %s.common_pet_id = %s.id",
			revisionAlias, revisionAlias, subjectAlias, revisionAlias, subjectAlias))
}
