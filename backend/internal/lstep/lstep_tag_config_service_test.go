package lstep

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func newTagConfigSvc(repo *mockLstepTagConfigRepository) LstepTagConfigService {
	return NewLstepTagConfigService(repo)
}

// --- auto_managed_prefixes ---

func TestListAutoManagedPrefixes_OK(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
			return []*model.LstepAutoManagedPrefix{{ID: 1, Prefix: "vaccine_", Category: "C2"}}, nil
		},
	}
	items, err := newTagConfigSvc(repo).ListAutoManagedPrefixes(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "vaccine_", items[0].Prefix)
}

func TestListAutoManagedPrefixes_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
			return nil, errors.New("db error")
		},
	}
	_, err := newTagConfigSvc(repo).ListAutoManagedPrefixes(context.Background())
	assert.Error(t, err)
}

func TestCreateAutoManagedPrefix_OK(t *testing.T) {
	var captured *model.LstepAutoManagedPrefix
	repo := &mockLstepTagConfigRepository{
		createAutoManagedPrefixFn: func(_ context.Context, m *model.LstepAutoManagedPrefix) error {
			captured = m
			m.ID = 10
			return nil
		},
	}
	desc := "ワクチン系"
	item, err := newTagConfigSvc(repo).CreateAutoManagedPrefix(context.Background(), CreateAutoManagedPrefixInput{
		Prefix:      "vaccine_",
		Category:    "C2",
		Description: &desc,
	})
	require.NoError(t, err)
	assert.Equal(t, "vaccine_", captured.Prefix)
	assert.Equal(t, "C2", captured.Category)
	assert.Equal(t, uint64(10), item.ID)
}

func TestCreateAutoManagedPrefix_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		createAutoManagedPrefixFn: func(_ context.Context, _ *model.LstepAutoManagedPrefix) error {
			return errors.New("db error")
		},
	}
	_, err := newTagConfigSvc(repo).CreateAutoManagedPrefix(context.Background(), CreateAutoManagedPrefixInput{Prefix: "x", Category: "C1"})
	assert.Error(t, err)
}

func TestCreateAutoManagedPrefix_RejectsInvalidCategory(t *testing.T) {
	created := false
	repo := &mockLstepTagConfigRepository{
		createAutoManagedPrefixFn: func(_ context.Context, _ *model.LstepAutoManagedPrefix) error {
			created = true
			return nil
		},
	}
	_, err := newTagConfigSvc(repo).CreateAutoManagedPrefix(context.Background(), CreateAutoManagedPrefixInput{Prefix: "x", Category: "c2"})
	assert.Error(t, err, "LSA-14: lowercase category must be rejected")
	assert.False(t, created, "invalid category must not reach repository")
}

func TestDeleteAutoManagedPrefix_OK(t *testing.T) {
	var deletedID uint64
	repo := &mockLstepTagConfigRepository{
		deleteAutoManagedPrefixFn: func(_ context.Context, id uint64) error {
			deletedID = id
			return nil
		},
	}
	err := newTagConfigSvc(repo).DeleteAutoManagedPrefix(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, uint64(5), deletedID)
}

func TestDeleteAutoManagedPrefix_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		deleteAutoManagedPrefixFn: func(_ context.Context, _ uint64) error {
			return errors.New("not found")
		},
	}
	err := newTagConfigSvc(repo).DeleteAutoManagedPrefix(context.Background(), 99)
	assert.Error(t, err)
}

// --- condition_tag_mappings ---

func TestListConditionTagMappings_OK(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		findAllConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
			return []*model.LstepConditionTagMapping{{ID: 1, ConditionCode: "DM", TagName: "CHRON_DM"}}, nil
		},
	}
	items, err := newTagConfigSvc(repo).ListConditionTagMappings(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestListConditionTagMappings_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		findAllConditionTagMappingsFn: func(_ context.Context) ([]*model.LstepConditionTagMapping, error) {
			return nil, errors.New("db error")
		},
	}
	_, err := newTagConfigSvc(repo).ListConditionTagMappings(context.Background())
	assert.Error(t, err)
}

func TestCreateConditionTagMapping_OK(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		createConditionTagMappingFn: func(_ context.Context, m *model.LstepConditionTagMapping) error {
			m.ID = 20
			return nil
		},
	}
	item, err := newTagConfigSvc(repo).CreateConditionTagMapping(context.Background(), CreateConditionTagMappingInput{
		ConditionCode: "DM",
		TagName:       "CHRON_DM",
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(20), item.ID)
}

func TestCreateConditionTagMapping_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		createConditionTagMappingFn: func(_ context.Context, _ *model.LstepConditionTagMapping) error {
			return errors.New("db error")
		},
	}
	_, err := newTagConfigSvc(repo).CreateConditionTagMapping(context.Background(), CreateConditionTagMappingInput{
		ConditionCode: "DM",
		TagName:       "CHRON_DM",
	})
	assert.Error(t, err)
}

func TestCreateConditionTagMapping_RemapsAlreadyExistsToConditionCode(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		createConditionTagMappingFn: func(_ context.Context, _ *model.LstepConditionTagMapping) error {
			return apperrors.WrapAlreadyExists("lstep_condition_tag_mapping", "")
		},
	}
	_, err := newTagConfigSvc(repo).CreateConditionTagMapping(context.Background(), CreateConditionTagMappingInput{
		ConditionCode: "CKD",
		TagName:       "CHRON_CKD",
	})
	require.Error(t, err)
	require.True(t, apperrors.IsAlreadyExists(err))
	assert.Contains(t, err.Error(), "CKD")
}

func TestDeleteConditionTagMapping_OK(t *testing.T) {
	var deletedID uint64
	repo := &mockLstepTagConfigRepository{
		deleteConditionTagMappingFn: func(_ context.Context, id uint64) error {
			deletedID = id
			return nil
		},
	}
	err := newTagConfigSvc(repo).DeleteConditionTagMapping(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, uint64(7), deletedID)
}

func TestDeleteConditionTagMapping_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		deleteConditionTagMappingFn: func(_ context.Context, _ uint64) error {
			return errors.New("not found")
		},
	}
	err := newTagConfigSvc(repo).DeleteConditionTagMapping(context.Background(), 99)
	assert.Error(t, err)
}

// --- send_purpose_tag_prefixes ---

func TestListSendPurposeTagPrefixes_OK(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		findAllSendPurposeTagPrefixesFn: func(_ context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
			return []*model.LstepSendPurposeTagPrefix{{ID: 1, Purpose: "vaccine_reminder", TagPrefix: "PREV_"}}, nil
		},
	}
	items, err := newTagConfigSvc(repo).ListSendPurposeTagPrefixes(context.Background())
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestListSendPurposeTagPrefixes_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		findAllSendPurposeTagPrefixesFn: func(_ context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
			return nil, errors.New("db error")
		},
	}
	_, err := newTagConfigSvc(repo).ListSendPurposeTagPrefixes(context.Background())
	assert.Error(t, err)
}

func TestCreateSendPurposeTagPrefix_OK(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		createSendPurposeTagPrefixFn: func(_ context.Context, m *model.LstepSendPurposeTagPrefix) error {
			m.ID = 30
			return nil
		},
	}
	item, err := newTagConfigSvc(repo).CreateSendPurposeTagPrefix(context.Background(), CreateSendPurposeTagPrefixInput{
		Purpose:   "vaccine_reminder",
		TagPrefix: "PREV_",
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(30), item.ID)
}

func TestCreateSendPurposeTagPrefix_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		createSendPurposeTagPrefixFn: func(_ context.Context, _ *model.LstepSendPurposeTagPrefix) error {
			return errors.New("db error")
		},
	}
	_, err := newTagConfigSvc(repo).CreateSendPurposeTagPrefix(context.Background(), CreateSendPurposeTagPrefixInput{
		Purpose:   "vaccine_reminder",
		TagPrefix: "PREV_",
	})
	assert.Error(t, err)
}

func TestDeleteSendPurposeTagPrefix_OK(t *testing.T) {
	var deletedID uint64
	repo := &mockLstepTagConfigRepository{
		deleteSendPurposeTagPrefixFn: func(_ context.Context, id uint64) error {
			deletedID = id
			return nil
		},
	}
	err := newTagConfigSvc(repo).DeleteSendPurposeTagPrefix(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), deletedID)
}

func TestDeleteSendPurposeTagPrefix_Error(t *testing.T) {
	repo := &mockLstepTagConfigRepository{
		deleteSendPurposeTagPrefixFn: func(_ context.Context, _ uint64) error {
			return errors.New("db error")
		},
	}
	err := newTagConfigSvc(repo).DeleteSendPurposeTagPrefix(context.Background(), 99)
	assert.Error(t, err)
}
