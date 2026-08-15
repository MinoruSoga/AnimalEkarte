package pet

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/audit"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type petOwnerServiceRepositoryDouble struct {
	links              []model.PetOwner
	sharedPets         []SharedPet
	sharedPetsErr      error
	sharedPetsCalls    int
	sharedPetsClinicID uint64
	sharedPetsOwnerID  uint64
	replaceCalls       int
}

func (r *petOwnerServiceRepositoryDouble) FindByPetID(
	context.Context,
	uint64,
	uint64,
) ([]model.PetOwner, error) {
	return append([]model.PetOwner(nil), r.links...), nil
}

func (r *petOwnerServiceRepositoryDouble) FindSharedPetsByOwnerID(
	_ context.Context,
	clinicID, ownerID uint64,
) ([]SharedPet, error) {
	r.sharedPetsCalls++
	r.sharedPetsClinicID = clinicID
	r.sharedPetsOwnerID = ownerID
	return append([]SharedPet(nil), r.sharedPets...), r.sharedPetsErr
}

func (r *petOwnerServiceRepositoryDouble) ReplaceForPet(
	_ context.Context,
	clinicID, petID uint64,
	links []model.PetOwner,
	_ *int,
) error {
	r.replaceCalls++
	r.links = make([]model.PetOwner, len(links))
	for i, link := range links {
		r.links[i] = model.PetOwner{
			ClinicID:     clinicID,
			PetID:        petID,
			OwnerID:      link.OwnerID,
			Relationship: link.Relationship,
		}
	}
	return nil
}

func (*petOwnerServiceRepositoryDouble) CountByOwnerID(
	context.Context,
	uint64,
	uint64,
) (int64, error) {
	return 0, nil
}

type petOwnerServicePetFinderDouble struct {
	pet *model.Pet
	err error
}

func (f *petOwnerServicePetFinderDouble) FindByID(
	context.Context,
	uint64,
	uint64,
) (*model.Pet, error) {
	return f.pet, f.err
}

type petOwnerServiceOwnerFinderDouble struct {
	err error
}

func (f *petOwnerServiceOwnerFinderDouble) FindByID(
	_ context.Context,
	clinicID, ownerID uint64,
) (*model.Owner, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &model.Owner{ID: ownerID, ClinicID: clinicID}, nil
}

type petOwnerServiceTransactorDouble struct{}

func (petOwnerServiceTransactorDouble) WithTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	return fn(ctx)
}

type petOwnerServiceAuditDouble struct {
	entries        []*audit.Entry
	err            error
	sawTransaction bool
}

func (l *petOwnerServiceAuditDouble) LogEntryTx(
	ctx context.Context,
	entry *audit.Entry,
) error {
	l.sawTransaction = persistence.TxFromContext(ctx) != nil
	l.entries = append(l.entries, entry)
	return l.err
}

type petOwnerServiceDBOwnerFinder struct {
	db *gorm.DB
}

func (f petOwnerServiceDBOwnerFinder) FindByID(
	ctx context.Context,
	clinicID, ownerID uint64,
) (*model.Owner, error) {
	var result model.Owner
	if err := persistence.DBOrTx(ctx, f.db).
		Where("clinic_id = ? AND id = ?", clinicID, ownerID).
		First(&result).Error; err != nil {
		return nil, apperrors.FromGORM(err, "owner", "")
	}
	return &result, nil
}

func TestPetOwnerService(t *testing.T) {
	t.Run("get shared pets validates owner clinic and returns repository projection", func(t *testing.T) {
		repo := &petOwnerServiceRepositoryDouble{
			sharedPets: []SharedPet{{
				ID:           10,
				Name:         "共同飼育ペット",
				Relationship: "家族",
			}},
		}
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{},
			&petOwnerServiceOwnerFinderDouble{},
			repo,
			petOwnerServiceTransactorDouble{},
			&petOwnerServiceAuditDouble{},
		)

		got, err := svc.GetSharedPetsByOwnerID(context.Background(), 1, 20)

		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "共同飼育ペット", got[0].Name)
		assert.Equal(t, "家族", got[0].Relationship)
		assert.Equal(t, 1, repo.sharedPetsCalls)
		assert.Equal(t, uint64(1), repo.sharedPetsClinicID)
		assert.Equal(t, uint64(20), repo.sharedPetsOwnerID)
	})

	t.Run("get shared pets returns a non-nil empty collection", func(t *testing.T) {
		repo := &petOwnerServiceRepositoryDouble{}
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{},
			&petOwnerServiceOwnerFinderDouble{},
			repo,
			petOwnerServiceTransactorDouble{},
			&petOwnerServiceAuditDouble{},
		)

		got, err := svc.GetSharedPetsByOwnerID(context.Background(), 1, 20)

		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("get shared pets rejects an owner outside the clinic before repository access", func(t *testing.T) {
		repo := &petOwnerServiceRepositoryDouble{}
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{},
			&petOwnerServiceOwnerFinderDouble{err: apperrors.WrapNotFound("owner", "20")},
			repo,
			petOwnerServiceTransactorDouble{},
			&petOwnerServiceAuditDouble{},
		)

		got, err := svc.GetSharedPetsByOwnerID(context.Background(), 1, 20)

		require.Error(t, err)
		assert.Nil(t, got)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Zero(t, repo.sharedPetsCalls)
	})

	t.Run("missing audit dependency fails closed before replacement", func(t *testing.T) {
		repo := &petOwnerServiceRepositoryDouble{}
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{pet: &model.Pet{ID: 10, ClinicID: 1, OwnerID: 20}},
			&petOwnerServiceOwnerFinderDouble{},
			repo,
			petOwnerServiceTransactorDouble{},
			nil,
		)

		err := svc.ReplaceForPet(context.Background(), 1, 10, &ReplacePetOwnersInput{
			Links: []PetOwnerLinkInput{{OwnerID: 21, Relationship: "家族"}},
		})

		require.Error(t, err)
		assert.Zero(t, repo.replaceCalls)
	})

	t.Run("missing pet in clinic returns not found", func(t *testing.T) {
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{err: apperrors.WrapNotFound("pet", "99")},
			&petOwnerServiceOwnerFinderDouble{},
			&petOwnerServiceRepositoryDouble{},
			petOwnerServiceTransactorDouble{},
			&petOwnerServiceAuditDouble{},
		)

		err := svc.ReplaceForPet(context.Background(), 1, 99, &ReplacePetOwnersInput{})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("get returns not found when pet is outside the clinic", func(t *testing.T) {
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{err: apperrors.WrapNotFound("pet", "99")},
			&petOwnerServiceOwnerFinderDouble{},
			&petOwnerServiceRepositoryDouble{},
			petOwnerServiceTransactorDouble{},
			&petOwnerServiceAuditDouble{},
		)

		links, err := svc.GetByPetID(context.Background(), 1, 99)

		require.Error(t, err)
		assert.Nil(t, links)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("get returns a non-nil empty collection for a pet in clinic", func(t *testing.T) {
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{pet: &model.Pet{ID: 10, ClinicID: 1}},
			&petOwnerServiceOwnerFinderDouble{},
			&petOwnerServiceRepositoryDouble{},
			petOwnerServiceTransactorDouble{},
			&petOwnerServiceAuditDouble{},
		)

		links, err := svc.GetByPetID(context.Background(), 1, 10)

		require.NoError(t, err)
		assert.NotNil(t, links)
		assert.Empty(t, links)
	})

	t.Run("primary owner is rejected as a sub-owner", func(t *testing.T) {
		repo := &petOwnerServiceRepositoryDouble{}
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{pet: &model.Pet{ID: 10, ClinicID: 1, OwnerID: 20}},
			&petOwnerServiceOwnerFinderDouble{},
			repo,
			petOwnerServiceTransactorDouble{},
			&petOwnerServiceAuditDouble{},
		)

		err := svc.ReplaceForPet(context.Background(), 1, 10, &ReplacePetOwnersInput{
			Links: []PetOwnerLinkInput{{OwnerID: 20, Relationship: "本人"}},
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Zero(t, repo.replaceCalls)
	})

	t.Run("duplicate owner IDs are rejected", func(t *testing.T) {
		repo := &petOwnerServiceRepositoryDouble{}
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{pet: &model.Pet{ID: 10, ClinicID: 1, OwnerID: 20}},
			&petOwnerServiceOwnerFinderDouble{},
			repo,
			petOwnerServiceTransactorDouble{},
			&petOwnerServiceAuditDouble{},
		)

		err := svc.ReplaceForPet(context.Background(), 1, 10, &ReplacePetOwnersInput{
			Links: []PetOwnerLinkInput{
				{OwnerID: 21, Relationship: "家族"},
				{OwnerID: 21, Relationship: "親族"},
			},
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Zero(t, repo.replaceCalls)
	})

	t.Run("relationship length is measured after trim in Unicode runes", func(t *testing.T) {
		tests := []struct {
			name         string
			relationship string
			wantError    bool
			wantStored   string
		}{
			{name: "empty", relationship: "", wantError: true},
			{name: "whitespace only", relationship: " \t\n ", wantError: true},
			{name: "51 runes", relationship: strings.Repeat("親", 51), wantError: true},
			{
				name:         "exactly 50 runes is accepted and trimmed",
				relationship: "  " + strings.Repeat("親", 50) + "  ",
				wantStored:   strings.Repeat("親", 50),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := &petOwnerServiceRepositoryDouble{}
				svc := NewPetOwnerService(
					&petOwnerServicePetFinderDouble{pet: &model.Pet{ID: 10, ClinicID: 1, OwnerID: 20}},
					&petOwnerServiceOwnerFinderDouble{},
					repo,
					petOwnerServiceTransactorDouble{},
					&petOwnerServiceAuditDouble{},
				)

				err := svc.ReplaceForPet(context.Background(), 1, 10, &ReplacePetOwnersInput{
					Links: []PetOwnerLinkInput{{OwnerID: 21, Relationship: tt.relationship}},
				})

				if tt.wantError {
					require.Error(t, err)
					assert.True(t, apperrors.IsInvalidInput(err))
					assert.Zero(t, repo.replaceCalls)
					return
				}
				require.NoError(t, err)
				require.Len(t, repo.links, 1)
				assert.Equal(t, tt.wantStored, repo.links[0].Relationship)
			})
		}
	})

	t.Run("empty replacement still writes one audit entry", func(t *testing.T) {
		repo := &petOwnerServiceRepositoryDouble{
			links: []model.PetOwner{{OwnerID: 21, Relationship: "家族"}},
		}
		auditLog := &petOwnerServiceAuditDouble{}
		svc := NewPetOwnerService(
			&petOwnerServicePetFinderDouble{pet: &model.Pet{ID: 10, ClinicID: 1, OwnerID: 20}},
			&petOwnerServiceOwnerFinderDouble{},
			repo,
			petOwnerServiceTransactorDouble{},
			auditLog,
		)

		err := svc.ReplaceForPet(context.Background(), 1, 10, &ReplacePetOwnersInput{
			Links:     []PetOwnerLinkInput{},
			ActorID:   ptrUint64(7),
			ActorType: model.AuditActorTypeStaff,
		})

		require.NoError(t, err)
		require.Len(t, auditLog.entries, 1)
		assert.Equal(t, model.AuditActionPetOwnerReplace, auditLog.entries[0].Action)
		assert.Equal(t, "pet", auditLog.entries[0].Resource)
		assert.Equal(t, []petOwnerAuditValue{}, auditLog.entries[0].NewValue)
	})

	t.Run("cross-clinic owner is rejected without creating a link", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		const clinicA, clinicB = uint64(1), uint64(2)
		primaryOwner := makeTestOwner(t, db, clinicA, "自院主飼主")
		foreignOwner := makeTestOwner(t, db, clinicB, "他院副飼主")
		pet := makeSpeciesAndPet(t, db, clinicA, primaryOwner.ID, "越境拒否ペット")
		svc := NewPetOwnerService(
			NewRepository(db),
			petOwnerServiceDBOwnerFinder{db: db},
			NewPetOwnerRepository(db),
			persistence.NewTransactor(db),
			&petOwnerServiceAuditDouble{},
		)

		err := svc.ReplaceForPet(context.Background(), clinicA, pet.ID, &ReplacePetOwnersInput{
			Links: []PetOwnerLinkInput{{OwnerID: foreignOwner.ID, Relationship: "越境"}},
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		var count int64
		require.NoError(t, db.Model(&model.PetOwner{}).
			Where("clinic_id = ? AND pet_id = ?", clinicA, pet.ID).
			Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("cross-clinic pet is rejected without changing either clinic", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		const clinicA, clinicB = uint64(1), uint64(2)
		primaryA := makeTestOwner(t, db, clinicA, "自院主飼主")
		primaryB := makeTestOwner(t, db, clinicB, "他院主飼主")
		subOwnerB := makeTestOwner(t, db, clinicB, "他院副飼主")
		petA := makeSpeciesAndPet(t, db, clinicA, primaryA.ID, "自院ペット")
		petB := makeSpeciesAndPet(t, db, clinicB, primaryB.ID, "他院ペット")
		makePetOwnerLink(t, db, clinicB, petB.ID, subOwnerB.ID, "家族")
		beforeA, linksA := loadPetOwnerState(t, db, clinicA, petA.ID)
		beforeB, linksB := loadPetOwnerState(t, db, clinicB, petB.ID)
		auditLog := &petOwnerServiceAuditDouble{}
		svc := NewPetOwnerService(
			NewRepository(db),
			petOwnerServiceDBOwnerFinder{db: db},
			NewPetOwnerRepository(db),
			persistence.NewTransactor(db),
			auditLog,
		)

		err := svc.ReplaceForPet(context.Background(), clinicA, petB.ID, &ReplacePetOwnersInput{
			Links: []PetOwnerLinkInput{},
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
		assert.Empty(t, auditLog.entries)
		afterA, afterLinksA := loadPetOwnerState(t, db, clinicA, petA.ID)
		afterB, afterLinksB := loadPetOwnerState(t, db, clinicB, petB.ID)
		assert.Equal(t, beforeA.Version, afterA.Version)
		assert.Equal(t, linksA, afterLinksA)
		assert.Equal(t, beforeB.Version, afterB.Version)
		assert.Equal(t, linksB, afterLinksB)
	})

	t.Run("audit failure rolls back replacement and version", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		const clinicID = uint64(1)
		primaryOwner := makeTestOwner(t, db, clinicID, "監査主飼主")
		oldOwner := makeTestOwner(t, db, clinicID, "監査旧副飼主")
		newOwner := makeTestOwner(t, db, clinicID, "監査新副飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "監査失敗ペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "変更前")
		beforePet, beforeLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		auditFailure := errors.New("audit unavailable")
		auditLog := &petOwnerServiceAuditDouble{err: auditFailure}
		svc := NewPetOwnerService(
			NewRepository(db),
			petOwnerServiceDBOwnerFinder{db: db},
			NewPetOwnerRepository(db),
			persistence.NewTransactor(db),
			auditLog,
		)

		err := svc.ReplaceForPet(context.Background(), clinicID, pet.ID, &ReplacePetOwnersInput{
			Version: &beforePet.Version,
			Links:   []PetOwnerLinkInput{{OwnerID: newOwner.ID, Relationship: "変更後"}},
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, auditFailure)
		assert.True(t, auditLog.sawTransaction)
		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, beforePet.Version, afterPet.Version)
		assert.Equal(t, beforeLinks, afterLinks)
	})

	t.Run("stale version returns conflict without changing links or version", func(t *testing.T) {
		db := setupPetOwnerRepositoryTestDB(t)
		const clinicID = uint64(1)
		primaryOwner := makeTestOwner(t, db, clinicID, "CAS主飼主")
		oldOwner := makeTestOwner(t, db, clinicID, "CAS旧副飼主")
		newOwner := makeTestOwner(t, db, clinicID, "CAS新副飼主")
		pet := makeSpeciesAndPet(t, db, clinicID, primaryOwner.ID, "CASペット")
		makePetOwnerLink(t, db, clinicID, pet.ID, oldOwner.ID, "変更前")
		beforePet, beforeLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		staleVersion := beforePet.Version + 1
		auditLog := &petOwnerServiceAuditDouble{}
		svc := NewPetOwnerService(
			NewRepository(db),
			petOwnerServiceDBOwnerFinder{db: db},
			NewPetOwnerRepository(db),
			persistence.NewTransactor(db),
			auditLog,
		)

		err := svc.ReplaceForPet(context.Background(), clinicID, pet.ID, &ReplacePetOwnersInput{
			Version: &staleVersion,
			Links:   []PetOwnerLinkInput{{OwnerID: newOwner.ID, Relationship: "変更後"}},
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Empty(t, auditLog.entries)
		afterPet, afterLinks := loadPetOwnerState(t, db, clinicID, pet.ID)
		assert.Equal(t, beforePet.Version, afterPet.Version)
		assert.Equal(t, beforeLinks, afterLinks)
	})
}
