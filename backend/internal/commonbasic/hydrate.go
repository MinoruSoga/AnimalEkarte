package commonbasic

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// HydrateOwners は owners の現行 basic revision と subject を Basic / BasicSubject へ載せる。
// clinicIDs は呼び出し側が認可した医院で、owner.ClinicID がその外にあれば unresolved にする。
// R-1: 1 件でも解決できなければ全体を失敗させ、owners は更新しない。
func HydrateOwners(ctx context.Context, db *gorm.DB, clinicIDs []uint64, owners []*model.Owner) error {
	if len(owners) == 0 {
		return nil
	}
	refs, err := ownerRefs(clinicIDs, owners)
	if err != nil {
		return err
	}

	tx := persistence.DBOrTx(ctx, db)
	subjects, err := resolveSubjects(tx, ownerSubjectKind, refs)
	if err != nil {
		return err
	}
	revisions, err := loadRevisions(tx, subjects, func(revision model.CommonOwnerBasicRevision) uint64 { return revision.ID })
	if err != nil {
		return err
	}

	// 全件の確認が済んでから代入する（途中で失敗しても owners を部分的に更新しない）。
	basics := make([]model.CommonOwnerBasicRevision, len(owners))
	subjectRefs := make([]model.CommonSubjectRef, len(owners))
	for i, owner := range owners {
		subject := subjects[localRef{clinicID: owner.ClinicID, localID: owner.ID}]
		revision, ok := revisions[subject.CurrentRevisionID]
		if !ok {
			return unresolved("owner %d/%d: current revision row is missing", owner.ClinicID, owner.ID)
		}
		if revision.CommonOwnerID != subject.CommonID {
			return unresolved("owner %d/%d: current revision belongs to another subject", owner.ClinicID, owner.ID)
		}
		basics[i] = revision
		subjectRefs[i] = subject
	}
	for i, owner := range owners {
		owner.Basic = &basics[i]
		owner.BasicSubject = &subjectRefs[i]
	}
	return nil
}

// ownerRefs は owners を (clinic_id, ローカル ID) の組へ変換する。nil 要素と認可範囲外の医院はここで拒否する。
func ownerRefs(clinicIDs []uint64, owners []*model.Owner) ([]localRef, error) {
	refs := make([]localRef, 0, len(owners))
	for _, owner := range owners {
		if owner == nil {
			return nil, apperrors.WrapInternalServerError("commonbasic: nil owner passed to HydrateOwners")
		}
		if err := requireClinicInScope(ownerSubjectKind, clinicIDs, owner.ClinicID, owner.ID); err != nil {
			return nil, err
		}
		refs = append(refs, localRef{clinicID: owner.ClinicID, localID: owner.ID})
	}
	return uniqueRefs(refs), nil
}

// HydratePets は pets の現行 basic revision・subject・動物種を Basic / BasicSubject / AnimalSpecies へ載せる。
// 規則は HydrateOwners と同じ（R-1: 1 件でも解決できなければ全体を失敗させ、pets は更新しない）。
// 動物種は revision の animal_species_id から引く。マスタに無ければ unresolved にする。
func HydratePets(ctx context.Context, db *gorm.DB, clinicIDs []uint64, pets []*model.Pet) error {
	if len(pets) == 0 {
		return nil
	}
	refs, err := petRefs(clinicIDs, pets)
	if err != nil {
		return err
	}

	tx := persistence.DBOrTx(ctx, db)
	subjects, err := resolveSubjects(tx, petSubjectKind, refs)
	if err != nil {
		return err
	}
	revisions, err := loadRevisions(tx, subjects, func(revision model.CommonPetBasicRevision) uint64 { return revision.ID })
	if err != nil {
		return err
	}
	species, err := loadAnimalSpecies(tx, revisions)
	if err != nil {
		return err
	}

	// 全件の確認が済んでから代入する（途中で失敗しても pets を部分的に更新しない）。
	basics := make([]model.CommonPetBasicRevision, len(pets))
	subjectRefs := make([]model.CommonSubjectRef, len(pets))
	speciesRows := make([]model.AnimalSpecies, len(pets))
	for i, pet := range pets {
		subject := subjects[localRef{clinicID: pet.ClinicID, localID: pet.ID}]
		revision, ok := revisions[subject.CurrentRevisionID]
		if !ok {
			return unresolved("pet %d/%d: current revision row is missing", pet.ClinicID, pet.ID)
		}
		if revision.CommonPetID != subject.CommonID {
			return unresolved("pet %d/%d: current revision belongs to another subject", pet.ClinicID, pet.ID)
		}
		animalSpecies, ok := species[revision.AnimalSpeciesID]
		if !ok {
			return unresolved("pet %d/%d: animal species %d is missing", pet.ClinicID, pet.ID, revision.AnimalSpeciesID)
		}
		basics[i] = revision
		subjectRefs[i] = subject
		speciesRows[i] = animalSpecies
	}
	for i, pet := range pets {
		pet.Basic = &basics[i]
		pet.BasicSubject = &subjectRefs[i]
		pet.AnimalSpecies = &speciesRows[i]
	}
	return nil
}

// petRefs は pets を (clinic_id, ローカル ID) の組へ変換する。nil 要素と認可範囲外の医院はここで拒否する。
func petRefs(clinicIDs []uint64, pets []*model.Pet) ([]localRef, error) {
	refs := make([]localRef, 0, len(pets))
	for _, pet := range pets {
		if pet == nil {
			return nil, apperrors.WrapInternalServerError("commonbasic: nil pet passed to HydratePets")
		}
		if err := requireClinicInScope(petSubjectKind, clinicIDs, pet.ClinicID, pet.ID); err != nil {
			return nil, err
		}
		refs = append(refs, localRef{clinicID: pet.ClinicID, localID: pet.ID})
	}
	return uniqueRefs(refs), nil
}

// loadAnimalSpecies は revision が参照する動物種をマスタから ID で読む。
func loadAnimalSpecies(tx *gorm.DB, revisions map[uint64]model.CommonPetBasicRevision) (map[uint64]model.AnimalSpecies, error) {
	idSet := make(map[uint64]struct{}, len(revisions))
	for _, revision := range revisions {
		idSet[revision.AnimalSpeciesID] = struct{}{}
	}
	ids := make([]uint64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var rows []model.AnimalSpecies
	if err := tx.Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "failed to load animal species for common pet revisions")
	}
	byID := make(map[uint64]model.AnimalSpecies, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	return byID, nil
}
