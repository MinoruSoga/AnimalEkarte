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

type stubPetLifeStateReader struct {
	state PetLifeState
	err   error
	calls int
}

func (s *stubPetLifeStateReader) ReadPetLifeState(_ context.Context, _, _ uint64) (PetLifeState, error) {
	s.calls++
	return s.state, s.err
}

func petStatusPtr(status model.PetStatus) *model.PetStatus { return &status }

// TestClassifyPetLifeState は EMR-491 D6 真理値表の純粋関数テスト。
// deceased_at != nil OR status=deceased -> deceased。status=alive かつ deceased_at=nil だけが alive。
// それ以外（NULL / 未知 status かつ deceased_at=nil）は alive にせず unresolved エラー。
func TestClassifyPetLifeState(t *testing.T) {
	deceasedAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name       string
		status     *model.PetStatus
		deceasedAt *time.Time
		want       PetLifeState
		wantErr    bool
	}{
		{name: "alive / nil date is alive", status: petStatusPtr(model.PetStatusAlive), want: PetLifeStateAlive},
		{name: "deceased / nil date is deceased", status: petStatusPtr(model.PetStatusDeceased), want: PetLifeStateDeceased},
		{name: "deceased / dated is deceased", status: petStatusPtr(model.PetStatusDeceased), deceasedAt: &deceasedAt, want: PetLifeStateDeceased},
		{name: "alive + date is deceased (inconsistent row is not alive)", status: petStatusPtr(model.PetStatusAlive), deceasedAt: &deceasedAt, want: PetLifeStateDeceased},
		{name: "NULL status + date is deceased", deceasedAt: &deceasedAt, want: PetLifeStateDeceased},
		{name: "unknown status + date is deceased", status: petStatusPtr(model.PetStatus("unknown")), deceasedAt: &deceasedAt, want: PetLifeStateDeceased},
		{name: "NULL status + nil date is unresolved", wantErr: true},
		{name: "unknown status + nil date is unresolved", status: petStatusPtr(model.PetStatus("unknown")), wantErr: true},
		{name: "empty status + nil date is unresolved", status: petStatusPtr(model.PetStatus("")), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ClassifyPetLifeState(tc.status, tc.deceasedAt)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrPetLifeStateUnresolved)
				assert.Empty(t, got, "unresolved must not carry a life state")
				assert.NotEqual(t, PetLifeStateAlive, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestValidatePetNotDeceased(t *testing.T) {
	msg := "死亡したペットは予約できません"

	t.Run("alive passes", func(t *testing.T) {
		reader := &stubPetLifeStateReader{state: PetLifeStateAlive}
		assert.NoError(t, ValidatePetNotDeceased(context.Background(), reader, 1, 5, msg))
		assert.Equal(t, 1, reader.calls)
	})

	t.Run("deceased is invalid input with the caller message", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetLifeStateReader{state: PetLifeStateDeceased}, 1, 5, msg)
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Contains(t, err.Error(), msg)
	})

	t.Run("not found stays not found", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetLifeStateReader{err: apperrors.WrapNotFound("pet", "5")}, 1, 5, msg)
		require.Error(t, err)
		assert.True(t, apperrors.IsNotFound(err))
	})

	t.Run("unresolved is an error and never passes", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetLifeStateReader{
			err: apperrors.Wrap(ErrPetLifeStateUnresolved, "pet status is not recorded"),
		}, 1, 5, msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPetLifeStateUnresolved)
		assert.False(t, apperrors.IsInvalidInput(err), "unknown is not a deceased business rejection")
		assert.NotContains(t, err.Error(), msg)
	})

	t.Run("reader failure is wrapped and propagated", func(t *testing.T) {
		cause := errors.New("db down")
		err := ValidatePetNotDeceased(context.Background(), &stubPetLifeStateReader{err: cause}, 1, 5, msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, cause)
		assert.Contains(t, err.Error(), "failed to verify pet status")
	})

	t.Run("empty state is rejected rather than treated as alive", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetLifeStateReader{}, 1, 5, msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPetLifeStateUnresolved)
	})

	t.Run("unrecognized state is rejected rather than treated as alive", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), &stubPetLifeStateReader{state: PetLifeState("alive ")}, 1, 5, msg)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPetLifeStateUnresolved)
	})

	t.Run("nil reader fails closed", func(t *testing.T) {
		err := ValidatePetNotDeceased(context.Background(), nil, 1, 5, msg)
		require.Error(t, err)
		assert.False(t, apperrors.IsInvalidInput(err))
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
	// EMR-495 P1: 旧 "finalized parent returns conflict" ケースは廃止した（medical_records.status 列は
	// 存在せず、conflictMsg は使われない）。ロックされた親が存在する限り、メッセージの有無に関わらず通る。
	t.Run("existing locked parent passes regardless of the retired conflict message", func(t *testing.T) {
		err := LockDraftMedicalRecord(context.Background(),
			&stubLocker{rec: &model.MedicalRecord{}}, 1, 10, "find failed", "編集できません")
		assert.NoError(t, err)
		err = LockDraftMedicalRecord(context.Background(),
			&stubLocker{rec: &model.MedicalRecord{}}, 1, 10, "find failed", "")
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
