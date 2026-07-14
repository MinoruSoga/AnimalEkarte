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

		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{
				CodeType: model.CodeTypeCheckupType,
				Codes:    []string{"CHK_01", "CHK_02"},
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, HlthHealthcheckDoneTag, deletedTag)
		if assert.Len(t, created, 1) {
			assert.Equal(t, uint64(10), created[0].ClinicID)
			assert.Equal(t, HlthHealthcheckDoneTag, created[0].TagName)
			assert.Equal(t, model.CodeTypeCheckupType, created[0].CodeType)
			assert.Equal(t, pq.StringArray{"CHK_01", "CHK_02"}, created[0].Codes)
			assert.Nil(t, created[0].SpeciesScope)
			assert.Nil(t, created[0].AgeMin)
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

	t.Run("returns wrapped error when soft delete fails", func(t *testing.T) {
		svc := NewLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			softDeleteByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) error {
				return errors.New("db error")
			},
		})
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_01"}},
		})
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("returns wrapped error when create fails", func(t *testing.T) {
		svc := NewLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			createFn: func(_ context.Context, _ *model.LstepTagCodeMapping) error {
				return errors.New("db error")
			},
		})
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_01"}},
		})
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("stops on first create failure and does not create remaining entries", func(t *testing.T) {
		callCount := 0
		svc := NewLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			createFn: func(_ context.Context, _ *model.LstepTagCodeMapping) error {
				callCount++
				if callCount == 1 {
					return errors.New("db error")
				}
				return nil
			},
		})
		_, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_01"}},
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_02"}},
		})
		assert.Error(t, err)
		assert.Equal(t, 1, callCount, "loop must stop at the first failing Create call")
	})

	t.Run("sets SpeciesScope and AgeMin when provided, and handles multiple entries", func(t *testing.T) {
		ageMin := 5
		var created []*model.LstepTagCodeMapping
		svc := NewLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			createFn: func(_ context.Context, mapping *model.LstepTagCodeMapping) error {
				created = append(created, mapping)
				mapping.ID = uint64(len(created))
				return nil
			},
		})

		got, err := svc.PutMappingsForTag(context.Background(), 10, PrevFilariaTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_01"}, SpeciesScope: "dog", AgeMin: &ageMin},
			{CodeType: model.CodeTypePrescription, Codes: []string{"RX_01"}},
		})

		assert.NoError(t, err)
		if assert.Len(t, created, 2) {
			if assert.NotNil(t, created[0].SpeciesScope) {
				assert.Equal(t, "dog", *created[0].SpeciesScope)
			}
			if assert.NotNil(t, created[0].AgeMin) {
				assert.Equal(t, ageMin, *created[0].AgeMin)
			}
			assert.Nil(t, created[1].SpeciesScope)
			assert.Nil(t, created[1].AgeMin)
		}
		assert.Len(t, got, 2)
	})
}
