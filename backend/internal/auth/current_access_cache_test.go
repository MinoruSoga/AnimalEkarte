package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

type countingCurrentAccessResolver struct {
	access *CurrentAccess
	err    error
	calls  int
}

func (r *countingCurrentAccessResolver) Resolve(
	context.Context,
	uint64,
) (*CurrentAccess, error) {
	r.calls++
	if r.err != nil {
		return nil, r.err
	}
	return cloneCurrentAccess(r.access), nil
}

func TestCachedCurrentAccessResolver_HitsWithinTTL(t *testing.T) {
	inner := &countingCurrentAccessResolver{access: &CurrentAccess{
		StaffID:      17,
		AccountEpoch: 9,
		MainClinicID: "23",
		ClinicIDs:    []uint64{23, 24},
	}}
	resolver := NewCachedCurrentAccessResolver(inner, 50*time.Millisecond)

	first, err := resolver.Resolve(context.Background(), 17)
	require.NoError(t, err)
	second, err := resolver.Resolve(context.Background(), 17)
	require.NoError(t, err)

	assert.Equal(t, 1, inner.calls)
	assert.Equal(t, first.ClinicIDs, second.ClinicIDs)
	first.ClinicIDs[0] = 99
	assert.Equal(t, uint64(23), second.ClinicIDs[0])
}

func TestCachedCurrentAccessResolver_ExpiresAndDoesNotCacheErrors(t *testing.T) {
	inner := &countingCurrentAccessResolver{
		err: apperrors.WrapForbidden("staff account is no longer active"),
	}
	resolver := NewCachedCurrentAccessResolver(inner, 50*time.Millisecond)

	_, err := resolver.Resolve(context.Background(), 17)
	require.Error(t, err)
	_, err = resolver.Resolve(context.Background(), 17)
	require.Error(t, err)
	assert.Equal(t, 2, inner.calls)

	inner.err = nil
	inner.access = &CurrentAccess{StaffID: 17, AccountEpoch: 1, MainClinicID: "1", ClinicIDs: []uint64{1}}
	first, err := resolver.Resolve(context.Background(), 17)
	require.NoError(t, err)
	assert.Equal(t, uint64(17), first.StaffID)

	time.Sleep(60 * time.Millisecond)
	_, err = resolver.Resolve(context.Background(), 17)
	require.NoError(t, err)
	assert.Equal(t, 4, inner.calls)
}

func TestCachedCurrentAccessResolver_NilInnerError(t *testing.T) {
	inner := &countingCurrentAccessResolver{err: errors.New("db down")}
	resolver := NewCachedCurrentAccessResolver(inner, time.Second)
	_, err := resolver.Resolve(context.Background(), 17)
	require.EqualError(t, err, "db down")
}
