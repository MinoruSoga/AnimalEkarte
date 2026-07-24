package main

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/pet"
)

type ownerPetCompositionDependencies struct {
	Insurance            owner.InsuranceFinder
	MedicalRecords       pet.MedicalRecordReader
	OwnerTags            owner.TagNotifier
	PetTags              pet.PetTagSynchronizer
	ChronicConditionTags pet.ChronicConditionTagSynchronizer
	Audit                owner.AuditLogger
}

// ownerPetRepositories are created before LSTEP so its application can receive
// the canonical owner/pet stores. Services are composed only after LSTEP makes
// its tag synchronizers available.
type ownerPetRepositories struct {
	Owner             owner.Repository
	Pet               pet.CompleteRepository
	AnimalSpecies     pet.AnimalSpeciesRepository
	ChronicConditions pet.ChronicConditionRepository
}

type ownerPetComposition struct {
	OwnerRepository   owner.Repository
	PetRepository     pet.CompleteRepository
	OwnerService      owner.Service
	PetService        pet.Service
	animalSpecies     pet.AnimalSpeciesService
	chronicConditions pet.ChronicConditionService
}

type ownerPetHandlers struct {
	Owner *owner.Handler
	Pet   *pet.Handler
}

func newOwnerPetRepositories(db *gorm.DB) ownerPetRepositories {
	petWriter := pet.NewWriter(db)
	ownerRepository := owner.NewRepository(
		db,
		pet.NewOwnerRegistrationAdapter(petWriter),
	)
	return ownerPetRepositories{
		Owner:             ownerRepository,
		Pet:               pet.NewRepositoryWithWriter(db, petWriter),
		AnimalSpecies:     pet.NewAnimalSpeciesRepository(db),
		ChronicConditions: pet.NewChronicConditionRepository(db),
	}
}

func newOwnerPetComposition(
	repositories ownerPetRepositories,
	dependencies ownerPetCompositionDependencies,
) ownerPetComposition {
	ownerService := owner.NewService(
		repositories.Owner,
		dependencies.Insurance,
		dependencies.OwnerTags,
		dependencies.Audit,
	)
	petService := pet.NewService(
		repositories.Pet,
		repositories.Owner,
		dependencies.Insurance,
		dependencies.MedicalRecords,
		dependencies.PetTags,
	)

	return ownerPetComposition{
		OwnerRepository: repositories.Owner,
		PetRepository:   repositories.Pet,
		OwnerService:    ownerService,
		PetService:      petService,
		animalSpecies: pet.NewAnimalSpeciesService(
			repositories.AnimalSpecies,
			repositories.Pet,
		),
		chronicConditions: pet.NewChronicConditionService(
			repositories.ChronicConditions,
			repositories.Pet,
			dependencies.ChronicConditionTags,
		),
	}
}

func (c ownerPetComposition) newHandlers(
	deletionLifecycle owner.DeletionLifecycle,
	requirePermission owner.PermissionMiddleware,
	hasPermission owner.PermissionChecker,
) ownerPetHandlers {
	return ownerPetHandlers{
		Owner: owner.NewHandler(
			c.OwnerService,
			deletionLifecycle,
			requirePermission,
			hasPermission,
		),
		Pet: pet.NewHandler(
			c.PetService,
			c.animalSpecies,
			c.chronicConditions,
			pet.PermissionMiddleware(requirePermission),
		),
	}
}
