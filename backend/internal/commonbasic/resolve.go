package commonbasic

import (
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// maxResolveRefs は 1 回の解決で扱う (clinic_id, ローカル ID) の組の上限。1 組あたり 2 パラメータで、
// PostgreSQL のバインド上限（65535）に収まる値にする。これを超える呼び出しは設計ミスとして失敗させる。
const maxResolveRefs = 30000

// subjectKind は owner / pet で共通の解決処理が参照するテーブルと列。
// 識別子はすべてコンパイル時定数で、入力から組み立てない。aliasPrefix は JOIN 別名の接頭辞で、
// internal/pet の JoinCurrentCommonBasicName（__cpcb / __cps）と同じ別名を作らない値にする。
type subjectKind struct {
	label                 string
	bindingTable          string
	bindingLocalColumn    string
	bindingSubjectColumn  string
	subjectTable          string
	revisionTable         string
	revisionSubjectColumn string
	aliasPrefix           string
}

var ownerSubjectKind = subjectKind{
	label:                 "owner",
	bindingTable:          "common_owner_clinic_bindings",
	bindingLocalColumn:    "local_owner_id",
	bindingSubjectColumn:  "common_owner_id",
	subjectTable:          "common_owner_subjects",
	revisionTable:         "common_owner_basic_revisions",
	revisionSubjectColumn: "common_owner_id",
	aliasPrefix:           "__o",
}

var petSubjectKind = subjectKind{
	label:                 "pet",
	bindingTable:          "common_pet_clinic_bindings",
	bindingLocalColumn:    "local_pet_id",
	bindingSubjectColumn:  "common_pet_id",
	subjectTable:          "common_pet_subjects",
	revisionTable:         "common_pet_basic_revisions",
	revisionSubjectColumn: "common_pet_id",
	aliasPrefix:           "__p",
}

// localRef は (clinic_id, ローカル ID) の組。ローカル ID 単独では解決しない。
type localRef struct {
	clinicID uint64
	localID  uint64
}

// subjectRow は binding と subject の結合結果。LEFT JOIN のため subject 側の列は NULL になり得る。
type subjectRow struct {
	ClinicID          uint64     `gorm:"column:clinic_id"`
	LocalID           uint64     `gorm:"column:local_id"`
	SubjectID         *uuid.UUID `gorm:"column:subject_id"`
	Version           *int64     `gorm:"column:subject_version"`
	CurrentRevisionID *uint64    `gorm:"column:current_revision_id"`
}

// uniqueRefs は組の重複を畳む（順序は保つ）。
func uniqueRefs(refs []localRef) []localRef {
	seen := make(map[localRef]struct{}, len(refs))
	out := make([]localRef, 0, len(refs))
	for _, ref := range refs {
		if _, dup := seen[ref]; dup {
			continue
		}
		seen[ref] = struct{}{}
		out = append(out, ref)
	}
	return out
}

// requireClinicInScope は呼び出し側が認可した医院の外にある組を unresolved として拒否する。
// 認可範囲外の組は binding があっても読まない（他院の基本情報を引かない）。
func requireClinicInScope(kind subjectKind, clinicIDs []uint64, clinicID, localID uint64) error {
	if slices.Contains(clinicIDs, clinicID) {
		return nil
	}
	return unresolved("%s %d/%d: clinic is outside the authorized scope", kind.label, clinicID, localID)
}

// resolveSubjects は refs の全組を binding → subject まで 1 文で解決する（version と現行ポインタは同一文で読む）。
// 1 組でも binding・subject・現行ポインタが欠ければ unresolved で、部分結果は返さない。
func resolveSubjects(tx *gorm.DB, kind subjectKind, refs []localRef) (map[localRef]model.CommonSubjectRef, error) {
	resolved := make(map[localRef]model.CommonSubjectRef, len(refs))
	if len(refs) == 0 {
		return resolved, nil
	}
	if len(refs) > maxResolveRefs {
		return nil, apperrors.WrapInternalServerError(fmt.Sprintf("commonbasic: %d refs exceed the resolve limit %d", len(refs), maxResolveRefs))
	}
	placeholders := strings.TrimSuffix(strings.Repeat("(?,?),", len(refs)), ",")
	args := make([]any, 0, len(refs)*2)
	for _, ref := range refs {
		args = append(args, ref.clinicID, ref.localID)
	}
	var rows []subjectRow
	err := tx.Table(kind.bindingTable+" AS b").
		Select(
			"b.clinic_id AS clinic_id",
			"b."+kind.bindingLocalColumn+" AS local_id",
			"s.id AS subject_id",
			"s.version AS subject_version",
			"s.current_basic_revision_id AS current_revision_id",
		).
		Joins("LEFT JOIN "+kind.subjectTable+" s ON s.id = b."+kind.bindingSubjectColumn).
		Where("(b.clinic_id, b."+kind.bindingLocalColumn+") IN ("+placeholders+")", args...).
		Scan(&rows).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to resolve common "+kind.label+" clinic bindings")
	}
	for _, row := range rows {
		ref := localRef{clinicID: row.ClinicID, localID: row.LocalID}
		switch {
		case row.SubjectID == nil || row.Version == nil:
			return nil, unresolved("%s %d/%d: common subject is missing", kind.label, ref.clinicID, ref.localID)
		case row.CurrentRevisionID == nil:
			return nil, unresolved("%s %d/%d: common subject has no current revision", kind.label, ref.clinicID, ref.localID)
		}
		resolved[ref] = model.CommonSubjectRef{
			CommonID:          *row.SubjectID,
			Version:           *row.Version,
			CurrentRevisionID: *row.CurrentRevisionID,
		}
	}
	for _, ref := range refs {
		if _, ok := resolved[ref]; !ok {
			return nil, unresolved("%s %d/%d: no common clinic binding", kind.label, ref.clinicID, ref.localID)
		}
	}
	return resolved, nil
}

// loadRevisions は subjects が指す現行 revision の行を ID で読む（revision は不変行）。
// 戻り値は revision ID をキーにした map。同一 subject の確認は呼び出し側が行う。
func loadRevisions[T any](tx *gorm.DB, subjects map[localRef]model.CommonSubjectRef, idOf func(T) uint64) (map[uint64]T, error) {
	idSet := make(map[uint64]struct{}, len(subjects))
	for _, subject := range subjects {
		idSet[subject.CurrentRevisionID] = struct{}{}
	}
	ids := make([]uint64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var rows []T
	if err := tx.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to load common basic revisions")
	}
	byID := make(map[uint64]T, len(rows))
	for _, row := range rows {
		byID[idOf(row)] = row
	}
	return byID, nil
}
