package sharedkernel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type stubPetByIDFinder struct {
	pet *model.Pet
	err error
}

func (s *stubPetByIDFinder) FindByID(_ context.Context, _, _ uint64) (*model.Pet, error) {
	return s.pet, s.err
}

func TestValidatePetNotDeceased(t *testing.T) {
	deceasedAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	msg := "死亡したペットは予約できません"

	t.Run("living pet passes", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetByIDFinder{
			pet: &model.Pet{ID: 5, Status: model.PetStatusAlive},
		}, 1, 5, msg)
		assert.NoError(t, err)
	})

	t.Run("deceased pet is invalid input", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetByIDFinder{
			pet: &model.Pet{ID: 5, Status: model.PetStatusDeceased, DeceasedAt: &deceasedAt},
		}, 1, 5, msg)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), msg)
	})

	t.Run("nil pet is not found", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetByIDFinder{pet: nil}, 1, 5, msg)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("find error is wrapped", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetByIDFinder{
			err: errors.New("db down"),
		}, 1, 5, msg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify pet status")
	})
}

type stubLocker struct {
	rec *model.MedicalRecord
	err error
}

func (s *stubLocker) LockByIDForUpdate(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
	return s.rec, s.err
}

// TestLockDraftMedicalRecord_NilParentFailsClosed は fail-closed 契約（A-5: nil 親 = NotFound）の
// kernel 直接テスト。service/medicalrecord 両 package の同名テストは delegate 経由で同じ実装を
// 検証し続ける（昇格で検証が薄まらないことの三重化）。
func TestLockDraftMedicalRecord_NilParentFailsClosed(t *testing.T) {
	err := LockDraftMedicalRecord(context.Background(), &stubLocker{rec: nil}, 1, 10, "find failed", "conflict msg")
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err), "nil parent は NotFound (fail-closed): %v", err)
}

func TestLockDraftMedicalRecord_Branches(t *testing.T) {
	t.Run("finalized parent returns conflict with given message", func(t *testing.T) {
		err := LockDraftMedicalRecord(context.Background(),
			&stubLocker{rec: &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}}, 1, 10, "find failed", "編集できません")
		require.Error(t, err)
		assert.True(t, apperrors.IsConflict(err))
		assert.Contains(t, err.Error(), "編集できません")
	})
	t.Run("draft parent passes", func(t *testing.T) {
		err := LockDraftMedicalRecord(context.Background(),
			&stubLocker{rec: &model.MedicalRecord{Status: model.MedicalRecordStatusDraft}}, 1, 10, "find failed", "conflict msg")
		assert.NoError(t, err)
	})
	t.Run("lock error is wrapped with findErrMsg", func(t *testing.T) {
		err := LockDraftMedicalRecord(context.Background(),
			&stubLocker{err: errors.New("db down")}, 1, 10, "find failed", "conflict msg")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "find failed")
	})
}

func TestSharedKernelValidators(t *testing.T) {
	t.Run("ValidateRequiredName rejects empty/too-long/control chars", func(t *testing.T) {
		assert.Error(t, ValidateRequiredName("  "))
		assert.Error(t, ValidateRequiredName(string(make([]rune, 0))+""))
		long := make([]rune, MasterNameMaxLength+1)
		for i := range long {
			long[i] = 'あ'
		}
		assert.Error(t, ValidateRequiredName(string(long)))
		assert.Error(t, ValidateRequiredName("bad\u0000name"))
		assert.NoError(t, ValidateRequiredName("正常な名前"))
	})
	t.Run("ValidateNonNegativePrice", func(t *testing.T) {
		neg := int64(-1)
		zero := int64(0)
		assert.NoError(t, ValidateNonNegativePrice(nil))
		assert.NoError(t, ValidateNonNegativePrice(&zero))
		assert.Error(t, ValidateNonNegativePrice(&neg))
	})
	t.Run("ValidateDiscountRate bounds", func(t *testing.T) {
		assert.NoError(t, ValidateDiscountRate(0))
		assert.NoError(t, ValidateDiscountRate(100))
		assert.Error(t, ValidateDiscountRate(-0.1))
		assert.Error(t, ValidateDiscountRate(100.1))
	})
	t.Run("AuditActorTypeFor", func(t *testing.T) {
		id := uint64(3)
		assert.Equal(t, model.AuditActorTypeStaff, AuditActorTypeFor(&id))
		assert.Equal(t, model.AuditActorTypeSystem, AuditActorTypeFor(nil))
	})
}

// TestValidatePetEnums は POC-11: owner/pet で複製されていたペット列挙検証の単一正本。
func TestValidatePetEnums(t *testing.T) {
	tests := []struct {
		name     string
		validate func(string) error
		valid    string
		invalid  string
	}{
		{
			name:     "gender",
			validate: ValidatePetGender,
			valid:    string(model.PetGenderMale),
			invalid:  "invalid_gender",
		},
		{
			name:     "status",
			validate: ValidatePetStatus,
			valid:    string(model.PetStatusAlive),
			invalid:  "invalid_status",
		},
		{
			name:     "acquisition type",
			validate: ValidatePetAcquisitionType,
			valid:    string(model.AcquisitionTypePurchase),
			invalid:  "invalid_acquisition",
		},
		{
			name:     "danger level",
			validate: ValidatePetDangerLevel,
			valid:    string(model.DangerLevelLow),
			invalid:  "invalid_danger",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, tt.validate(""))
			assert.NoError(t, tt.validate(tt.valid))
			assert.Error(t, tt.validate(tt.invalid))
		})
	}
}
