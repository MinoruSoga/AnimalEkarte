package main

import (
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/owner"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/pet"
)

type ownerPetAuditLogger interface {
	owner.AuditLogger
	pet.PetOwnerAuditLogger
	// AnimalSpeciesAuditLogger enables fail-closed audit on global animal_species writes (POC-07).
	pet.AnimalSpeciesAuditLogger
}

type ownerPetCompositionDependencies struct {
	Insurance            owner.InsuranceFinder
	MedicalRecords       pet.MedicalRecordReader
	OwnerTags            owner.TagNotifier
	PetTags              pet.PetTagSynchronizer
	ChronicConditionTags pet.ChronicConditionTagSynchronizer
	Audit                ownerPetAuditLogger
}

// ownerPetRepositories are created before LSTEP so its application can receive
// the canonical owner/pet stores. Services are composed only after LSTEP makes
// its tag synchronizers available.
type ownerPetRepositories struct {
	Owner             owner.Repository
	Pet               pet.CompleteRepository
	PetOwners         pet.PetOwnerRepository
	Transactor        pet.PetOwnerTransactor
	AnimalSpecies     pet.AnimalSpeciesRepository
	ChronicConditions pet.ChronicConditionRepository
}

type ownerPetComposition struct {
	OwnerRepository   owner.Repository
	PetRepository     pet.CompleteRepository
	OwnerService      owner.Service
	PetService        pet.Service
	PetOwnerService   pet.PetOwnerService
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
		PetOwners:         pet.NewPetOwnerRepository(db),
		Transactor:        persistence.NewTransactor(db),
		AnimalSpecies:     pet.NewAnimalSpeciesRepository(db),
		ChronicConditions: pet.NewChronicConditionRepository(db),
	}
}

func newOwnerPetComposition(
	repositories ownerPetRepositories,
	dependencies ownerPetCompositionDependencies,
) ownerPetComposition {
	ownerService := owner.NewServiceWithPetOwnerDeleteGuard(
		repositories.Owner,
		dependencies.Insurance,
		dependencies.OwnerTags,
		dependencies.Audit,
		repositories.Owner,
		repositories.PetOwners,
		repositories.Transactor,
	)
	petService := pet.NewServiceWithPetOwnerReader(
		repositories.Pet,
		repositories.Owner,
		dependencies.Insurance,
		dependencies.MedicalRecords,
		dependencies.PetTags,
		repositories.PetOwners,
		repositories.Transactor,
	)
	petOwnerService := pet.NewPetOwnerService(
		repositories.Pet,
		repositories.Owner,
		repositories.PetOwners,
		repositories.Transactor,
		dependencies.Audit,
	)

	return ownerPetComposition{
		OwnerRepository: repositories.Owner,
		PetRepository:   repositories.Pet,
		OwnerService:    ownerService,
		PetService:      petService,
		PetOwnerService: petOwnerService,
		animalSpecies: pet.NewAnimalSpeciesServiceWithAudit(
			repositories.AnimalSpecies,
			repositories.Pet,
			dependencies.Audit,
			repositories.Transactor,
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
		Pet: pet.NewHandlerWithPetOwners(
			c.PetService,
			c.animalSpecies,
			c.chronicConditions,
			c.PetOwnerService,
			c.OwnerRepository,
			pet.PermissionMiddleware(requirePermission),
		),
	}
}
