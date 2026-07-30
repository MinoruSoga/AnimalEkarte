package lstep

import (
	"context"
	"errors"
	"fmt"
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

type mockLstepTagCodeMappingTransactor struct {
	withTxFn func(ctx context.Context, fn func(context.Context) error) error
}

func (m *mockLstepTagCodeMappingTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m.withTxFn != nil {
		return m.withTxFn(ctx, fn)
	}
	return fn(ctx)
}

func newTestLstepTagCodeMappingService(repo LstepTagCodeMappingRepository) LstepTagCodeMappingService {
	return NewLstepTagCodeMappingService(repo, &mockLstepTagCodeMappingTransactor{})
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
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			findAllByClinicIDFn: func(_ context.Context, _ uint64) ([]*model.LstepTagCodeMapping, error) {
				return want, nil
			},
		})

		got, err := svc.ListMappings(context.Background(), 10)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("wraps repository error", func(t *testing.T) {
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
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
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
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
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{})
		got, err := svc.PutMappingsForTag(context.Background(), 10, "NOT_ALLOWED", nil)
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("rejects invalid code_type", func(t *testing.T) {
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{})
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: "not_a_type", Codes: []string{"X"}},
		})
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("rejects invalid species_scope", func(t *testing.T) {
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{})
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_01"}, SpeciesScope: "bird"},
		})
		assert.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("returns wrapped error when soft delete fails", func(t *testing.T) {
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
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
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
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
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
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
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
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

	t.Run("runs soft delete and every create in one transaction context", func(t *testing.T) {
		type txMarkerKey struct{}
		var operations []string
		txCalls := 0
		repo := &mockLstepTagCodeMappingRepositoryForCodeSettings{
			softDeleteByClinicIDAndTagNameFn: func(ctx context.Context, _ uint64, _ string) error {
				assert.Equal(t, "replacement-tx", ctx.Value(txMarkerKey{}))
				operations = append(operations, "soft-delete")
				return nil
			},
			createFn: func(ctx context.Context, _ *model.LstepTagCodeMapping) error {
				assert.Equal(t, "replacement-tx", ctx.Value(txMarkerKey{}))
				operations = append(operations, "create")
				return nil
			},
		}
		transactor := &mockLstepTagCodeMappingTransactor{
			withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				txCalls++
				return fn(context.WithValue(ctx, txMarkerKey{}, "replacement-tx"))
			},
		}
		svc := NewLstepTagCodeMappingService(repo, transactor)

		_, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_01"}},
			{CodeType: model.CodeTypePrescription, Codes: []string{"RX_01"}},
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, txCalls)
		assert.Equal(t, []string{"soft-delete", "create", "create"}, operations)
	})

	t.Run("does not write when transaction start fails", func(t *testing.T) {
		txErr := errors.New("transaction unavailable")
		writeCalls := 0
		repo := &mockLstepTagCodeMappingRepositoryForCodeSettings{
			softDeleteByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) error {
				writeCalls++
				return nil
			},
			createFn: func(_ context.Context, _ *model.LstepTagCodeMapping) error {
				writeCalls++
				return nil
			},
		}
		transactor := &mockLstepTagCodeMappingTransactor{
			withTxFn: func(context.Context, func(context.Context) error) error {
				return txErr
			},
		}
		svc := NewLstepTagCodeMappingService(repo, transactor)

		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_01"}},
		})

		assert.ErrorIs(t, err, txErr)
		assert.Nil(t, got)
		assert.Zero(t, writeCalls)
	})

	t.Run("fails closed when transaction dependency is missing", func(t *testing.T) {
		writeCalls := 0
		repo := &mockLstepTagCodeMappingRepositoryForCodeSettings{
			softDeleteByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) error {
				writeCalls++
				return nil
			},
			createFn: func(_ context.Context, _ *model.LstepTagCodeMapping) error {
				writeCalls++
				return nil
			},
		}
		svc := NewLstepTagCodeMappingService(repo, nil)

		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: []string{"CHK_01"}},
		})

		assert.ErrorContains(t, err, "transaction dependency is required")
		assert.Nil(t, got)
		assert.Zero(t, writeCalls)
	})

	// SEC-CS-F11: over-limit cardinality must fail before WithTx / SoftDelete / Create.
	t.Run("rejects over MaxEntries before SoftDelete or Create", func(t *testing.T) {
		txCalls := 0
		softDeleteCalls := 0
		createCalls := 0
		repo := &mockLstepTagCodeMappingRepositoryForCodeSettings{
			softDeleteByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) error {
				softDeleteCalls++
				return nil
			},
			createFn: func(_ context.Context, _ *model.LstepTagCodeMapping) error {
				createCalls++
				return nil
			},
		}
		transactor := &mockLstepTagCodeMappingTransactor{
			withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				txCalls++
				return fn(ctx)
			},
		}
		svc := NewLstepTagCodeMappingService(repo, transactor)

		entries := make([]PutMappingEntry, MaxTagCodeMappingEntries+1)
		for i := range entries {
			entries[i] = PutMappingEntry{
				CodeType: model.CodeTypeCheckupType,
				Codes:    []string{fmt.Sprintf("C%02d", i)},
			}
		}
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, entries)

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, txCalls, "over-limit must not open a transaction")
		assert.Zero(t, softDeleteCalls, "over-limit must not SoftDelete")
		assert.Zero(t, createCalls, "over-limit must not Create")
	})

	t.Run("rejects over MaxCodesPerEntry before SoftDelete or Create", func(t *testing.T) {
		txCalls := 0
		softDeleteCalls := 0
		createCalls := 0
		repo := &mockLstepTagCodeMappingRepositoryForCodeSettings{
			softDeleteByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) error {
				softDeleteCalls++
				return nil
			},
			createFn: func(_ context.Context, _ *model.LstepTagCodeMapping) error {
				createCalls++
				return nil
			},
		}
		transactor := &mockLstepTagCodeMappingTransactor{
			withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				txCalls++
				return fn(ctx)
			},
		}
		svc := NewLstepTagCodeMappingService(repo, transactor)

		codes := make([]string, MaxTagCodeMappingCodesPerEntry+1)
		for i := range codes {
			codes[i] = fmt.Sprintf("CODE_%03d", i)
		}
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, []PutMappingEntry{
			{CodeType: model.CodeTypeCheckupType, Codes: codes},
		})

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, txCalls)
		assert.Zero(t, softDeleteCalls)
		assert.Zero(t, createCalls)
	})

	t.Run("rejects over MaxTotalCodes before SoftDelete or Create", func(t *testing.T) {
		txCalls := 0
		softDeleteCalls := 0
		createCalls := 0
		repo := &mockLstepTagCodeMappingRepositoryForCodeSettings{
			softDeleteByClinicIDAndTagNameFn: func(_ context.Context, _ uint64, _ string) error {
				softDeleteCalls++
				return nil
			},
			createFn: func(_ context.Context, _ *model.LstepTagCodeMapping) error {
				createCalls++
				return nil
			},
		}
		transactor := &mockLstepTagCodeMappingTransactor{
			withTxFn: func(ctx context.Context, fn func(context.Context) error) error {
				txCalls++
				return fn(ctx)
			},
		}
		svc := NewLstepTagCodeMappingService(repo, transactor)

		// 3 entries × 70 codes = 210 > MaxTotalCodes(200), each under MaxCodesPerEntry(100).
		entries := make([]PutMappingEntry, 3)
		for i := range entries {
			codes := make([]string, 70)
			for j := range codes {
				codes[j] = fmt.Sprintf("E%d_C%02d", i, j)
			}
			entries[i] = PutMappingEntry{CodeType: model.CodeTypeCheckupType, Codes: codes}
		}
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, entries)

		assert.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, txCalls)
		assert.Zero(t, softDeleteCalls)
		assert.Zero(t, createCalls)
	})

	t.Run("accepts boundary MaxEntries MaxCodesPerEntry and MaxTotalCodes", func(t *testing.T) {
		// 2 entries × 100 codes = 200 total — exactly at all three caps that apply.
		createCalls := 0
		svc := newTestLstepTagCodeMappingService(&mockLstepTagCodeMappingRepositoryForCodeSettings{
			createFn: func(_ context.Context, _ *model.LstepTagCodeMapping) error {
				createCalls++
				return nil
			},
		})
		entries := make([]PutMappingEntry, 2)
		for i := range entries {
			codes := make([]string, MaxTagCodeMappingCodesPerEntry)
			for j := range codes {
				codes[j] = fmt.Sprintf("B%d_%03d", i, j)
			}
			entries[i] = PutMappingEntry{CodeType: model.CodeTypeCheckupType, Codes: codes}
		}
		got, err := svc.PutMappingsForTag(context.Background(), 10, HlthHealthcheckDoneTag, entries)
		assert.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, 2, createCalls)
	})
}
