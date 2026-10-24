package commonbasic

import (
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

// sqlAliasPattern は JOIN の別名として許す平易な SQL 識別子。別名は SQL 文字列へ直接埋め込むため、
// 呼び出し側はコンパイル時定数だけを渡す。
var sqlAliasPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// JoinCurrentOwnerBasic はローカル owner（ownerAlias）から binding → subject → 現行 basic revision
// （revisionAlias）への JOIN を追加する。revision は subject の現行ポインタが指し、かつ同じ subject に
// 属する行だけに絞る（ON 句で保証する）。ownerAlias の JOIN は本関数より前に済ませること。
//
// left が true なら LEFT JOIN で、binding・subject・現行 revision の無い owner も行として残り、revisionAlias の
// 列が NULL になる。一覧や検索で未解決を黙って落とさず Hydrate* で検出するための形で、表示系はこちらを使う。
// false は未解決の owner を結果から除く。表示に使うと黙った除外になるため、未解決が無関係な絞り込みだけに使う。
//
// 別名は SQL 文字列へ直接埋め込むため、コンパイル時定数だけを渡すこと。識別子でない別名は呼び出し側の誤りとして
// panic する。
func JoinCurrentOwnerBasic(tx *gorm.DB, ownerAlias, revisionAlias string, left bool) *gorm.DB {
	return joinCurrentBasic(tx, ownerSubjectKind, ownerAlias, revisionAlias, left)
}

// JoinCurrentPetBasic は JoinCurrentOwnerBasic の pet 版。internal/pet の JoinCurrentCommonBasicName
// （INNER JOIN のみ）を一般化したもので、別名の接尾辞が異なるため同じクエリで併用できる。
func JoinCurrentPetBasic(tx *gorm.DB, petAlias, revisionAlias string, left bool) *gorm.DB {
	return joinCurrentBasic(tx, petSubjectKind, petAlias, revisionAlias, left)
}

// joinCurrentBasic は owner / pet 共通の JOIN 組み立て。
func joinCurrentBasic(tx *gorm.DB, kind subjectKind, localAlias, revisionAlias string, left bool) *gorm.DB {
	if !sqlAliasPattern.MatchString(localAlias) || !sqlAliasPattern.MatchString(revisionAlias) {
		panic(fmt.Sprintf("commonbasic: join alias must be a plain SQL identifier (local=%q revision=%q)", localAlias, revisionAlias))
	}
	joinKeyword := "JOIN"
	if left {
		joinKeyword = "LEFT JOIN"
	}
	bindingAlias := localAlias + kind.aliasPrefix + "b"
	subjectAlias := localAlias + kind.aliasPrefix + "s"
	return tx.
		Joins(fmt.Sprintf(
			"%s %s %s ON %s.clinic_id = %s.clinic_id AND %s.%s = %s.id",
			joinKeyword, kind.bindingTable, bindingAlias, bindingAlias, localAlias, bindingAlias, kind.bindingLocalColumn, localAlias)).
		Joins(fmt.Sprintf(
			"%s %s %s ON %s.id = %s.%s",
			joinKeyword, kind.subjectTable, subjectAlias, subjectAlias, bindingAlias, kind.bindingSubjectColumn)).
		Joins(fmt.Sprintf(
			"%s %s %s ON %s.id = %s.current_basic_revision_id AND %s.%s = %s.id",
			joinKeyword, kind.revisionTable, revisionAlias, revisionAlias, subjectAlias, revisionAlias, kind.revisionSubjectColumn, subjectAlias))
}
