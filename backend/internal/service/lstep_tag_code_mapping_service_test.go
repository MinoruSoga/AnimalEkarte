package service

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type mockLstepTagCodeMappingRepositoryForCodeSettings struct {
	findAllByClinicIDFn              func(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error)
	softDeleteByClinicIDAndTagNameFn func(ctx context.Context, clinicID uint64, tagName string) error
	createFn                         func(ctx context.Context, mapping *model.LstepTagCodeMapping) error
}

func (m *mockLstepTagCodeMappingRepositoryForCodeSettings) FindAllByClinicID(ctx context.Context, clinicID uint64) ([]*model.LstepTagCodeMapping, error) {
	if m.findAllByClinicIDFn != nil {
		return m.findAllByClinicIDFn(ctx, clinicID)
	}
	return nil, nil
}

func (m *mockLstepTagCodeMappingRepositoryForCodeSettings) FindByClinicIDAndTagName(ctx context.Context, clinicID uint64, tagName string) ([]*model.LstepTagCodeMapping, error) {
	return nil, nil
}

func (m *mockLstepTagCodeMappingRepositoryForCodeSettings) Create(ctx context.Context, mapping *model.LstepTagCodeMapping) error {
	if m.createFn != nil {
		return m.createFn(ctx, mapping)
	}
	return nil
}

func (m *mockLstepTagCodeMappingRepositoryForCodeSettings) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.LstepTagCodeMapping, error) {
	return nil, nil
}

func (m *mockLstepTagCodeMappingRepositoryForCodeSettings) SoftDelete(ctx context.Context, clinicID, id uint64) error {
	return nil
}

func (m *mockLstepTagCodeMappingRepositoryForCodeSettings) SoftDeleteByClinicIDAndTagName(ctx context.Context, clinicID uint64, tagName string) error {
	if m.softDeleteByClinicIDAndTagNameFn != nil {
		return m.softDeleteByClinicIDAndTagNameFn(ctx, clinicID, tagName)
	}
	return nil
}

func TestLstepTagCodeMappingService_ListMappings(t *testing.T) {
	t.Run("returns mappings", func(t *testing.T) {
		want := []*model.LstepTagCodeMapping{
			{
				ID:       1,
				ClinicID: 10,
				TagName:  LtvFoodPurchaseTag,
				CodeType: model.CodeTypeMerchandiseItem,
				Codes:    pq.StringArray{"FOOD_A"},
			},
		}
		svc := NewLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			findAllByClinicIDFn: func(_ context.Context, _ uint64) ([]*model.LstepTagCodeMapping, error) {
				return want, nil
			},
		})

		got, err := svc.ListMappings(context.Background(), 10)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		svc := NewLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			findAllByClinicIDFn: func(_ context.Context, _ uint64) ([]*model.LstepTagCodeMapping, error) {
				return nil, errors.New("db error")
			},
		})

		got, err := svc.ListMappings(context.Background(), 10)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestLstepTagCodeMappingService_PutMappingsForTag(t *testing.T) {
	t.Run("creates entries after soft delete", func(t *testing.T) {
		var deletedTag string
		var created []*model.LstepTagCodeMapping
		svc := NewLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			softDeleteByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, tagName string) error {
				deletedTag = tagName
				return nil
			},
			createFn: func(_ context.Context, mapping *model.LstepTagCodeMapping) error {
				created = append(created, mapping)
				mapping.ID = uint64(len(created))
				return nil
			},
		})

		ageMin := 8
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthSpecialCheckupCandidateTag, []PutMappingEntry{
			{
				CodeType:     model.CodeTypeSpecialtyOphthalmology,
				Codes:        []string{"EYE_01", "EYE_02"},
				SpeciesScope: model.SpeciesScopeDog,
				AgeMin:       &ageMin,
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, HlthSpecialCheckupCandidateTag, deletedTag)
		if assert.Len(t, created, 1) {
			assert.Equal(t, uint64(10), created[0].ClinicID)
			assert.Equal(t, HlthSpecialCheckupCandidateTag, created[0].TagName)
			assert.Equal(t, model.CodeTypeSpecialtyOphthalmology, created[0].CodeType)
			assert.Equal(t, pq.StringArray{"EYE_01", "EYE_02"}, created[0].Codes)
			if assert.NotNil(t, created[0].SpeciesScope) {
				assert.Equal(t, model.SpeciesScopeDog, *created[0].SpeciesScope)
			}
			if assert.NotNil(t, created[0].AgeMin) {
				assert.Equal(t, ageMin, *created[0].AgeMin)
			}
		}
		assert.Len(t, got, 1)
		assert.Equal(t, created[0], got[0])
	})

	t.Run("rejects non-configurable tag", func(t *testing.T) {
		svc := NewLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{})
		got, err := svc.PutMappingsForTag(context.Background(), 10, "NOT_ALLOWED", nil)
		assert.Error(t, err)
		assert.Nil(t, got)
	})
}
