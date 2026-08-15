package lstep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
	lstepapi "github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// injectTagClient bypasses the deploy write gate for unit tests that exercise
// AddOwnerTag/RemoveOwnerTag success paths (TASK-610: default LSTEP_WRITE_API_ENABLED off).
func injectTagClient(svc LstepTagService, client lstepapi.Client) {
	if s, ok := svc.(*lstepTagService); ok {
		s.buildClientFn = func(_ context.Context, _ uint64) (lstepapi.Client, error) {
			return client, nil
		}
	}
}

// ---- mocks ----

type mockLstepTagConfigRepository struct {
	findAllAutoManagedPrefixesFn    func(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error)
	createAutoManagedPrefixFn       func(ctx context.Context, m *model.LstepAutoManagedPrefix) error
	deleteAutoManagedPrefixFn       func(ctx context.Context, id uint64) error
	findAllConditionTagMappingsFn   func(ctx context.Context) ([]*model.LstepConditionTagMapping, error)
	createConditionTagMappingFn     func(ctx context.Context, m *model.LstepConditionTagMapping) error
	deleteConditionTagMappingFn     func(ctx context.Context, id uint64) error
	findAllSendPurposeTagPrefixesFn func(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error)
	createSendPurposeTagPrefixFn    func(ctx context.Context, m *model.LstepSendPurposeTagPrefix) error
	deleteSendPurposeTagPrefixFn    func(ctx context.Context, id uint64) error
}

func (m *mockLstepTagConfigRepository) FindAllAutoManagedPrefixes(ctx context.Context) ([]*model.LstepAutoManagedPrefix, error) {
	if m.findAllAutoManagedPrefixesFn != nil {
		return m.findAllAutoManagedPrefixesFn(ctx)
	}
	return nil, nil
}

func (m *mockLstepTagConfigRepository) CreateAutoManagedPrefix(ctx context.Context, rec *model.LstepAutoManagedPrefix) error {
	if m.createAutoManagedPrefixFn != nil {
		return m.createAutoManagedPrefixFn(ctx, rec)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) DeleteAutoManagedPrefix(ctx context.Context, id uint64) error {
	if m.deleteAutoManagedPrefixFn != nil {
		return m.deleteAutoManagedPrefixFn(ctx, id)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) FindAllConditionTagMappings(ctx context.Context) ([]*model.LstepConditionTagMapping, error) {
	if m.findAllConditionTagMappingsFn != nil {
		return m.findAllConditionTagMappingsFn(ctx)
	}
	return nil, nil
}

func (m *mockLstepTagConfigRepository) CreateConditionTagMapping(ctx context.Context, rec *model.LstepConditionTagMapping) error {
	if m.createConditionTagMappingFn != nil {
		return m.createConditionTagMappingFn(ctx, rec)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) DeleteConditionTagMapping(ctx context.Context, id uint64) error {
	if m.deleteConditionTagMappingFn != nil {
		return m.deleteConditionTagMappingFn(ctx, id)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) FindAllSendPurposeTagPrefixes(ctx context.Context) ([]*model.LstepSendPurposeTagPrefix, error) {
	if m.findAllSendPurposeTagPrefixesFn != nil {
		return m.findAllSendPurposeTagPrefixesFn(ctx)
	}
	return nil, nil
}

func (m *mockLstepTagConfigRepository) CreateSendPurposeTagPrefix(ctx context.Context, rec *model.LstepSendPurposeTagPrefix) error {
	if m.createSendPurposeTagPrefixFn != nil {
		return m.createSendPurposeTagPrefixFn(ctx, rec)
	}
	return nil
}

func (m *mockLstepTagConfigRepository) DeleteSendPurposeTagPrefix(ctx context.Context, id uint64) error {
	if m.deleteSendPurposeTagPrefixFn != nil {
		return m.deleteSendPurposeTagPrefixFn(ctx, id)
	}
	return nil
}

// ---- tests ----
// Reuses: mockLstepSettingsService (lstep_lifecycle_service_test.go)
//         mockOwnerRepository       (owner_service_test.go)
//         mockLstepTagCacheRepository (lstep_lifecycle_service_test.go)
//         mockAuditService           (lstep_lifecycle_service_test.go)

func TestGetOwnerTags_NotLinked(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1}, nil
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.False(t, res.IsLinked)
	assert.Empty(t, res.Tags)
}

func TestGetOwnerTags_OwnerNotFound(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, apperrors.WrapNotFound("owner", "1")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	_, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.Error(t, err)
}

func TestGetOwnerTags_CacheFallback(t *testing.T) {
	lineID := "U123"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return []*model.LstepTagCache{{TagName: "my_tag"}}, nil
		},
	}
	// GetRawCredentials returns empty apiKey → client is nil → falls back to cache
	settingsSvc := &mockLstepSettingsService{}
	svc := NewLstepTagService(settingsSvc, ownerRepo, tagCache, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.True(t, res.IsLinked)
	assert.Equal(t, []string{"my_tag"}, res.Tags)
}

func TestAddOwnerTag_AutoManagedTag(t *testing.T) {
	svc := NewLstepTagService(&mockLstepSettingsService{}, &mockOwnerRepository{}, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "cpm_stage_1", nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestIsAutoManagedTag_NewSpecPrefixes(t *testing.T) {
	tags := []string{
		"CPM_01_出会い",
		"LTV_上位20",
		"LTV_フード購入あり",
		"VISIT_120日超",
		"PET_犬あり",
		"HLTH_健診あり",
		"PREV_ワクチン期限",
		exclTagDeliveryStop,
	}

	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			assert.True(t, isSystemManagedTag(tag))
		})
	}
}

func TestAddOwnerTag_OwnerNotFound(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "my_tag", nil)
	assert.Error(t, err)
}

func TestRemoveOwnerTag_AutoManagedTag(t *testing.T) {
	tagConfigRepo := &mockLstepTagConfigRepository{
		findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
			return []*model.LstepAutoManagedPrefix{{Prefix: "vaccine_", Category: "C2"}}, nil
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, &mockOwnerRepository{}, &mockLstepTagCacheRepository{}, &mockAuditService{}, tagConfigRepo)
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "vaccine_rabies", nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestRemoveOwnerTag_NotLinked(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1}, nil
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "my_tag", nil)
	assert.NoError(t, err)
}

// ---- isAutoManagedTagWithPrefixes (pure function) ----

func TestIsAutoManagedTagWithPrefixes(t *testing.T) {
	prefixes := []*model.LstepAutoManagedPrefix{
		{Prefix: "vaccine_", Category: "C2"},
		{Prefix: "exam_result", Category: "C3"},
	}
	tests := []struct {
		name       string
		tagName    string
		dbPrefixes []*model.LstepAutoManagedPrefix
		want       bool
	}{
		{name: "system managed prefix short-circuits", tagName: "CPM_01", dbPrefixes: nil, want: true},
		{name: "exact match against db prefix", tagName: "vaccine_", dbPrefixes: prefixes, want: true},
		{name: "prefix match against db prefix", tagName: "vaccine_rabies", dbPrefixes: prefixes, want: true},
		{name: "no match returns false", tagName: "manual_note", dbPrefixes: prefixes, want: false},
		{name: "empty db prefixes and non system tag returns false", tagName: "manual_note", dbPrefixes: nil, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAutoManagedTagWithPrefixes(tt.tagName, tt.dbPrefixes))
		})
	}
}

// ---- isAutoManagedTag (indirect via AddOwnerTag/RemoveOwnerTag) ----

func TestAddOwnerTag_NotManaged_NilTagConfigRepo_Proceeds(t *testing.T) {
	lineID := "U999"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", "http://example.invalid", "", nil
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	injectTagClient(svc, &mockLstepAPIClient{})
	err := svc.AddOwnerTag(context.Background(), 1, 1, "manual_note", nil)
	assert.NoError(t, err)
}

func TestAddOwnerTag_IsAutoManagedTag_RepoError(t *testing.T) {
	tagConfigRepo := &mockLstepTagConfigRepository{
		findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, &mockOwnerRepository{}, &mockLstepTagCacheRepository{}, &mockAuditService{}, tagConfigRepo)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "manual_note", nil)
	assert.Error(t, err)
}

func TestRemoveOwnerTag_IsAutoManagedTag_RepoError(t *testing.T) {
	tagConfigRepo := &mockLstepTagConfigRepository{
		findAllAutoManagedPrefixesFn: func(_ context.Context) ([]*model.LstepAutoManagedPrefix, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, &mockOwnerRepository{}, &mockLstepTagCacheRepository{}, &mockAuditService{}, tagConfigRepo)
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "manual_note", nil)
	assert.Error(t, err)
}

// ---- buildClient (indirect via GetOwnerTags/AddOwnerTag) ----

func TestGetOwnerTags_BuildClientError(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
			return false, errors.New("settings error")
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestGetOwnerTags_GetRawCredentialsError(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "", "", "", errors.New("credentials error")
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Nil(t, res)
}

// ---- GetOwnerTags: cache fallback error ----

func TestGetOwnerTags_CacheFallbackError(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	tagCache := &mockLstepTagCacheRepository{
		findByOwnerFn: func(_ context.Context, _, _ uint64) ([]*model.LstepTagCache, error) {
			return nil, errors.New("cache db error")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, tagCache, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Nil(t, res)
}

// ---- GetOwnerTags: real client via httptest server (GetUserTags is not disabled) ----

func TestGetOwnerTags_ClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tags":["tag1","tag2"]}`))
	}))
	defer server.Close()

	lineID := "U123"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, []string{"tag1", "tag2"}, res.Tags)
}

func TestGetOwnerTags_ClientUserNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	lineID := "U404"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.Empty(t, res.Tags)
}

func TestGetOwnerTags_ClientServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	lineID := "U500"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestGetOwnerTags_ClientTagsNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", server.URL, "", nil
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	res, err := svc.GetOwnerTags(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.Empty(t, res.Tags)
}

// ---- AddOwnerTag: remaining branches ----

func TestAddOwnerTag_LstepOptOut(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LstepOptOut: true, LineUserID: &lineID}, nil
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.Error(t, err)
}

func TestAddOwnerTag_LineUserIDNil(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1}, nil
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}

func TestAddOwnerTag_BuildClientError(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
			return false, errors.New("settings error")
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.Error(t, err)
}

func TestAddOwnerTag_ClientNil(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.AddOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.Error(t, err)
	assert.True(t, apperrors.IsInvalidInput(err))
}

func TestAddOwnerTag_Success_CacheUpsertErrorIsBestEffort(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", "http://example.invalid", "", nil
		},
	}
	tagCache := &mockLstepTagCacheRepository{
		upsertTagFn: func(_ context.Context, _, _ uint64, _, _, _ string) error {
			return errors.New("cache error")
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, tagCache, &mockAuditService{}, nil)
	injectTagClient(svc, &mockLstepAPIClient{})
	err := svc.AddOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.NoError(t, err)
}

// ---- RemoveOwnerTag: remaining branches ----

func TestRemoveOwnerTag_OwnerNotFound(t *testing.T) {
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewLstepTagService(&mockLstepSettingsService{}, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.Error(t, err)
}

func TestRemoveOwnerTag_BuildClientError(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) {
			return false, errors.New("settings error")
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.Error(t, err)
}

func TestRemoveOwnerTag_ClientNil_ReturnsNilNoError(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		isSyncEnabledFn: func(_ context.Context, _ uint64) (bool, error) { return false, nil },
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, &mockLstepTagCacheRepository{}, &mockAuditService{}, nil)
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.NoError(t, err)
}

func TestRemoveOwnerTag_Success_CacheDeleteErrorIsBestEffort(t *testing.T) {
	lineID := "U1"
	ownerRepo := &mockOwnerRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.Owner, error) {
			return &model.Owner{ID: 1, LineUserID: &lineID}, nil
		},
	}
	settingsSvc := &mockLstepSettingsService{
		getRawCredentialsFn: func(_ context.Context, _ uint64) (string, string, string, error) {
			return "test-api-key", "http://example.invalid", "", nil
		},
	}
	tagCache := &mockLstepTagCacheRepository{
		deleteTagFn: func(_ context.Context, _, _ uint64, _ string) error {
			return errors.New("cache error")
		},
	}
	svc := NewLstepTagService(settingsSvc, ownerRepo, tagCache, &mockAuditService{}, nil)
	injectTagClient(svc, &mockLstepAPIClient{})
	err := svc.RemoveOwnerTag(context.Background(), 1, 1, "manual_tag", nil)
	assert.NoError(t, err)
}
