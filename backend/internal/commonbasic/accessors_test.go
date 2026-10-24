package commonbasic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// accessors_test.go — DB を使わない accessor / 別名検証のテストと、テスト共通のヘルパー。

// requireUnresolved は err が R-1 の unresolved（422・安定 code）であることを確認する。
func requireUnresolved(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	assert.True(t, apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved), "got %v", err)
	assert.ErrorIs(t, err, apperrors.ErrUnprocessable)
}

// recoverPanicError は fn が panic することを確認し、その panic 値（error）を返す。
func recoverPanicError(t *testing.T, fn func()) (recovered error) {
	t.Helper()
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected a panic")
		err, ok := r.(error)
		require.True(t, ok, "panic value must be an error, got %T", r)
		recovered = err
	}()
	fn()
	return nil
}

func TestOwnerOfAndPetOf_FailClosedWhenNotHydrated(t *testing.T) {
	ownerBasic, err := OwnerOf(&model.Owner{ID: 1, ClinicID: 2})
	requireUnresolved(t, err)
	assert.Nil(t, ownerBasic)

	_, err = OwnerOf(nil)
	requireUnresolved(t, err)

	petBasic, err := PetOf(&model.Pet{ID: 3, ClinicID: 2})
	requireUnresolved(t, err)
	assert.Nil(t, petBasic)

	_, err = PetOf(nil)
	requireUnresolved(t, err)
}

func TestOwnerOfAndPetOf_ReturnHydratedBasic(t *testing.T) {
	owner := &model.Owner{Basic: &model.CommonOwnerBasicRevision{Name: "山田"}}
	pet := &model.Pet{Basic: &model.CommonPetBasicRevision{Name: "ポチ"}}

	ownerBasic, err := OwnerOf(owner)
	require.NoError(t, err)
	assert.Equal(t, "山田", ownerBasic.Name)

	petBasic, err := PetOf(pet)
	require.NoError(t, err)
	assert.Equal(t, "ポチ", petBasic.Name)
}

func TestMustBasic_PanicsWithUnresolvedWhenNotHydrated(t *testing.T) {
	requireUnresolved(t, recoverPanicError(t, func() { MustOwnerBasic(&model.Owner{ID: 1}) }))
	requireUnresolved(t, recoverPanicError(t, func() { MustPetBasic(&model.Pet{ID: 1}) }))

	assert.NotPanics(t, func() { MustOwnerBasic(&model.Owner{Basic: &model.CommonOwnerBasicRevision{}}) })
	assert.NotPanics(t, func() { MustPetBasic(&model.Pet{Basic: &model.CommonPetBasicRevision{}}) })
}

func TestPtrs_PointsAtSliceElements(t *testing.T) {
	owners := []model.Owner{{ID: 1}, {ID: 2}}

	ptrs := Ptrs(owners)

	require.Len(t, ptrs, 2)
	ptrs[1].Remarks = "updated"
	assert.Equal(t, "updated", owners[1].Remarks)
}

func TestJoinCurrentBasic_RejectsNonIdentifierAlias(t *testing.T) {
	assert.Panics(t, func() { JoinCurrentOwnerBasic(nil, "owners; DROP TABLE owners", "rev", false) })
	assert.Panics(t, func() { JoinCurrentPetBasic(nil, "pets", "rev --", true) })
	assert.Panics(t, func() { JoinCurrentPetBasic(nil, "", "rev", true) })
}
