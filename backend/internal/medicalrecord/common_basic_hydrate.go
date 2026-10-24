package medicalrecord

import (
	"context"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/commonbasic"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-491: owner / pet の名前・動物種・生死などの基本情報は owners / pets の列ではなく現行 basic revision にある。
// 入れ子要約（toOwnerSummary / toPetSummary）は hydrate 済みの Basic を読むため、Owner / Pet を preload する
// 読み取り経路は、この helper で現行 basic revision を載せてから返す。1 件でも解決できなければ全体が
// unresolved（422）で、空値や廃止列へのフォールバックはしない。clinicIDs は owners / pets を読んだ認可集合と
// 同じものを渡す。

// basicTargets は hydrate 対象の owner / pet を集める。nil は集めない
// （preload が対象外にした関連の不在は、基本情報の未解決ではない）。
type basicTargets struct {
	owners []*model.Owner
	pets   []*model.Pet
}

func (t *basicTargets) addOwner(owner *model.Owner) {
	if owner != nil {
		t.owners = append(t.owners, owner)
	}
}

// addPet は pet と、preload 済みの pet.Owner を対象に加える。
func (t *basicTargets) addPet(pet *model.Pet) {
	if pet == nil {
		return
	}
	t.pets = append(t.pets, pet)
	t.addOwner(pet.Owner)
}

// hydrate は集めた owner / pet に現行 basic revision を載せる（動物種は HydratePets が revision から引く）。
func (t *basicTargets) hydrate(ctx context.Context, db *gorm.DB, clinicIDs []uint64) error {
	if err := commonbasic.HydrateOwners(ctx, db, clinicIDs, t.owners); err != nil {
		return err
	}
	return commonbasic.HydratePets(ctx, db, clinicIDs, t.pets)
}

// hydrateMedicalRecordBasics はカルテの Owner / Pet に現行 basic revision を載せる。
func hydrateMedicalRecordBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, records []*model.MedicalRecord) error {
	var targets basicTargets
	for _, record := range records {
		if record == nil {
			continue
		}
		targets.addOwner(record.Owner)
		targets.addPet(record.Pet)
	}
	return targets.hydrate(ctx, db, clinicIDs)
}

// hydrateHospitalizationBasics は入院の Owner / Pet に現行 basic revision を載せる。
func hydrateHospitalizationBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, items []*model.Hospitalization) error {
	var targets basicTargets
	for _, item := range items {
		if item == nil {
			continue
		}
		targets.addOwner(item.Owner)
		targets.addPet(item.Pet)
	}
	return targets.hydrate(ctx, db, clinicIDs)
}

// hydrateVaccinationBasics はワクチン接種の Pet（と preload 済みの Pet.Owner）に現行 basic revision を載せる。
func hydrateVaccinationBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, items []*model.Vaccination) error {
	var targets basicTargets
	for _, item := range items {
		if item == nil {
			continue
		}
		targets.addPet(item.Pet)
	}
	return targets.hydrate(ctx, db, clinicIDs)
}

// hydrateExaminationBasics は検査の Pet（と preload 済みの Pet.Owner）に現行 basic revision を載せる。
func hydrateExaminationBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, items []*model.Examination) error {
	var targets basicTargets
	for _, item := range items {
		if item == nil {
			continue
		}
		targets.addPet(item.Pet)
	}
	return targets.hydrate(ctx, db, clinicIDs)
}

// hydrateCheckupBasics は健診の MedicalRecord.Pet（と preload 済みの Pet.Owner）に現行 basic revision を載せる。
func hydrateCheckupBasics(ctx context.Context, db *gorm.DB, clinicIDs []uint64, items []*model.Checkup) error {
	var targets basicTargets
	for _, item := range items {
		if item == nil || item.MedicalRecord == nil {
			continue
		}
		targets.addPet(item.MedicalRecord.Pet)
	}
	return targets.hydrate(ctx, db, clinicIDs)
}
