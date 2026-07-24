package owner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- mock OwnerService ----

type mockOwnerService struct {
	listFn                    func(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error)
	getByIDFn                 func(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	getByIDForClinicsFn       func(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error)
	createWithPetsFn          func(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error)
	updateFn                  func(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error)
	deleteFn                  func(ctx context.Context, clinicID, id uint64) error
	linkLineUserIDFn          func(ctx context.Context, clinicID, id uint64, lineUserID *string, actorUserID *uint64) error
	updateDeliveryExclusionFn func(ctx context.Context, clinicID, id uint64, input UpdateDeliveryExclusionInput) (*model.Owner, error)
	updateDeliveryCautionFn   func(ctx context.Context, clinicID, id uint64, input UpdateDeliveryCautionInput) (*model.Owner, error)
	updateTransferStatusFn    func(ctx context.Context, clinicID, id uint64, input UpdateTransferStatusInput) (*model.Owner, error)
	confirmLineIDFn           func(ctx context.Context, clinicID, id uint64, actorUserID *uint64) (*model.Owner, error)
}

func (m *mockOwnerService) List(ctx context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
	return m.listFn(ctx, clinicIDs, page, limit, search)
}

func (m *mockOwnerService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error) {
	return m.getByIDFn(ctx, clinicID, id)
}

func (m *mockOwnerService) GetByIDForClinics(ctx context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
	if m.getByIDForClinicsFn != nil {
		return m.getByIDForClinicsFn(ctx, clinicIDs, id)
	}
	return nil, nil
}

func (m *mockOwnerService) CreateWithPets(ctx context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
	return m.createWithPetsFn(ctx, clinicID, input)
}

func (m *mockOwnerService) Update(ctx context.Context, clinicID, id uint64, input *UpdateOwnerInput) (*model.Owner, error) {
	return m.updateFn(ctx, clinicID, id, input)
}

func (m *mockOwnerService) Delete(ctx context.Context, clinicID, id uint64) error {
	return m.deleteFn(ctx, clinicID, id)
}

func (m *mockOwnerService) LinkLineUserID(ctx context.Context, clinicID, id uint64, lineUserID *string, actorUserID *uint64) error {
	if m.linkLineUserIDFn != nil {
		return m.linkLineUserIDFn(ctx, clinicID, id, lineUserID, actorUserID)
	}
	return nil
}

func (m *mockOwnerService) UpdateDeliveryExclusion(ctx context.Context, clinicID, id uint64, input UpdateDeliveryExclusionInput) (*model.Owner, error) {
	if m.updateDeliveryExclusionFn != nil {
		return m.updateDeliveryExclusionFn(ctx, clinicID, id, input)
	}
	return &model.Owner{ID: id, ClinicID: clinicID}, nil
}

func (m *mockOwnerService) UpdateTransferStatus(ctx context.Context, clinicID, id uint64, input UpdateTransferStatusInput) (*model.Owner, error) {
	if m.updateTransferStatusFn != nil {
		return m.updateTransferStatusFn(ctx, clinicID, id, input)
	}
	return &model.Owner{ID: id, ClinicID: clinicID}, nil
}

func (m *mockOwnerService) UpdateDeliveryCaution(ctx context.Context, clinicID, id uint64, input UpdateDeliveryCautionInput) (*model.Owner, error) {
	if m.updateDeliveryCautionFn != nil {
		return m.updateDeliveryCautionFn(ctx, clinicID, id, input)
	}
	return &model.Owner{ID: id, ClinicID: clinicID}, nil
}

func (m *mockOwnerService) ConfirmLineID(ctx context.Context, clinicID, id uint64, actorUserID *uint64) (*model.Owner, error) {
	if m.confirmLineIDFn != nil {
		return m.confirmLineIDFn(ctx, clinicID, id, actorUserID)
	}
	return &model.Owner{ID: id, ClinicID: clinicID}, nil
}

// ---- test helpers ----

func newHandlerWithOwnerSvc(svc Service) *Handler {
	return NewHandler(svc, &mockOwnerDeletionLifecycle{}, allowPermission, allowPermissionCheck)
}

type mockOwnerDeletionLifecycle struct{}

func (*mockOwnerDeletionLifecycle) HandleOwnerDeletion(context.Context, uint64, uint64) error {
	return nil
}

func allowPermission(string, string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func allowPermissionCheck(*gin.Context, string, string) bool {
	return true
}

type mockEffectivePermissionService struct {
	getEffectivePermissionsFn func(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)
}

func (m *mockEffectivePermissionService) GetEffectivePermissions(
	ctx context.Context,
	staffID, clinicID uint64,
) ([]model.PermissionGroupRule, error) {
	if m.getEffectivePermissionsFn == nil {
		return nil, nil
	}
	return m.getEffectivePermissionsFn(ctx, staffID, clinicID)
}

func setNonSystemAdmin(c *gin.Context) {
	c.Set("is_system_admin", false)
	c.Set("user_id", "1")
	c.Set("clinic_id", "1")
}

// setClinicID は gin.Context に clinic_id を設定するヘルパー
func setClinicID(c *gin.Context) {
	c.Set("clinic_id", "1")
}

// ---- ListOwners ----

func TestListOwners(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(c *gin.Context)
		svc        *mockOwnerService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns paginated owners",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, clinicIDs []uint64, page, limit int, search string) ([]model.Owner, int64, error) {
					return []model.Owner{{ID: 1, Name: "田中太郎"}}, 1, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_name":"田中太郎"`,
		},
		{
			name:     "returns empty list when no owners",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, _ []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					return []model.Owner{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"total":0`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			query:      "",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 on invalid pagination",
			query:      "page=0",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "passes search param to service",
			query:    "page=1&limit=10&search=田中",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, _ []uint64, _, _ int, search string) ([]model.Owner, int64, error) {
					assert.Equal(t, "田中", search)
					return []model.Owner{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "returns 500 on service error",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, _ []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					return nil, 0, fmt.Errorf("unexpected db failure")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
		// ---- #86: 拠点横断一覧 (clinic_ids クエリ) の所属検証 ----
		{
			name:     "defaults to current clinic when clinic_ids absent",
			query:    "page=1&limit=10",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				listFn: func(_ context.Context, clinicIDs []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
					return []model.Owner{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "passes requested clinic_ids when all assigned",
			query: "page=1&limit=10&clinic_ids=1,2",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", false)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockOwnerService{
				listFn: func(_ context.Context, clinicIDs []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					assert.Equal(t, []uint64{1, 2}, clinicIDs)
					return []model.Owner{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "returns 403 when clinic_ids contains unassigned clinic",
			query: "page=1&limit=10&clinic_ids=1,99",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", false)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:  "system_admin can list an active clinic from trusted authority",
			query: "page=1&limit=10&clinic_ids=2",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockOwnerService{
				listFn: func(_ context.Context, clinicIDs []uint64, _, _ int, _ string) ([]model.Owner, int64, error) {
					assert.Equal(t, []uint64{2}, clinicIDs)
					return []model.Owner{}, 0, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "system_admin cannot list a clinic outside trusted active authority",
			query: "page=1&limit=10&clinic_ids=99",
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "returns 400 for malformed clinic_ids",
			query:      "page=1&limit=10&clinic_ids=abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOwnerSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, http.NoBody)
			tt.setupCtx(c)

			h.ListOwners(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- GetOwner ----

func TestGetOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		paramID    string
		setupCtx   func(c *gin.Context)
		svc        *mockOwnerService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "returns owner for valid id",
			paramID:  "42",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				getByIDForClinicsFn: func(_ context.Context, clinicIDs []uint64, id uint64) (*model.Owner, error) {
					assert.Equal(t, []uint64{1}, clinicIDs)
					assert.Equal(t, uint64(42), id)
					return &model.Owner{ID: 42, Name: "佐藤花子"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_name":"佐藤花子"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "abc",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when owner not found",
			paramID:  "999",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				getByIDForClinicsFn: func(_ context.Context, _ []uint64, _ uint64) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOwnerSvc(tt.svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.GetOwner(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- CreateOwner ----

func TestCreateOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := func() map[string]any {
		return map[string]any{
			"owner_name":      "山田一郎",
			"owner_name_kana": "ヤマダイチロウ",
			"email":           "yamada@example.com",
			"phone":           "090-1234-5678",
			"membership_type": "non_member",
			"pets":            []any{},
		}
	}

	tests := []struct {
		name       string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockOwnerService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "creates owner successfully",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				createWithPetsFn: func(_ context.Context, _ uint64, input *CreateOwnerInput) (*model.Owner, error) {
					return &model.Owner{ID: 1, Name: input.OwnerName}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"owner_name":"山田一郎"`,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			body:       validBody(),
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for invalid JSON",
			body:       "invalid json",
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 409 on duplicate email",
			body:     validBody(),
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				createWithPetsFn: func(_ context.Context, _ uint64, _ *CreateOwnerInput) (*model.Owner, error) {
					return nil, apperrors.WrapAlreadyExists("owner", "yamada@example.com")
				},
			},
			wantStatus: http.StatusConflict,
		},
		// ---- #84: 登録時の医院指定 (clinic_id) の所属検証 ----
		{
			name: "creates owner in specified assigned clinic",
			body: func() map[string]any {
				b := validBody()
				b["clinic_id"] = 2
				return b
			}(),
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", false)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockOwnerService{
				createWithPetsFn: func(_ context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
					assert.Equal(t, uint64(2), clinicID)
					return &model.Owner{ID: 1, ClinicID: clinicID, Name: input.OwnerName}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "returns 403 when specified clinic is not assigned",
			body: func() map[string]any {
				b := validBody()
				b["clinic_id"] = 99
				return b
			}(),
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", false)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "system_admin can create owner in an active clinic from trusted authority",
			body: func() map[string]any {
				b := validBody()
				b["clinic_id"] = 2
				return b
			}(),
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc: &mockOwnerService{
				createWithPetsFn: func(_ context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
					assert.Equal(t, uint64(2), clinicID)
					return &model.Owner{ID: 1, ClinicID: clinicID, Name: input.OwnerName}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "system_admin cannot create owner outside trusted active authority",
			body: func() map[string]any {
				b := validBody()
				b["clinic_id"] = 99
				return b
			}(),
			setupCtx: func(c *gin.Context) {
				setClinicID(c)
				c.Set("is_system_admin", true)
				c.Set("clinic_ids", []uint64{1, 2})
			},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "specified clinic equal to context clinic skips assignment check",
			body: func() map[string]any {
				b := validBody()
				b["clinic_id"] = 1
				return b
			}(),
			// clinic_ids / is_system_admin 未設定でも JWT clinic と同一なら検証不要で通る
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				createWithPetsFn: func(_ context.Context, clinicID uint64, input *CreateOwnerInput) (*model.Owner, error) {
					assert.Equal(t, uint64(1), clinicID)
					return &model.Owner{ID: 1, ClinicID: clinicID, Name: input.OwnerName}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOwnerSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			tt.setupCtx(c)

			h.CreateOwner(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

// ---- UpdateOwner ----

func TestUpdateOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ptrStr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		paramID    string
		body       any
		setupCtx   func(c *gin.Context)
		svc        *mockOwnerService
		wantStatus int
		wantBody   string
	}{
		{
			name:     "updates owner successfully",
			paramID:  "1",
			body:     map[string]any{"owner_name": "田中次郎"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateOwnerInput) (*model.Owner, error) {
					return &model.Owner{ID: 1, Name: *input.OwnerName}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"owner_name":"田中次郎"`,
		},
		{
			name:     "partial update with only phone",
			paramID:  "1",
			body:     map[string]any{"phone": "080-9999-0000"},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				updateFn: func(_ context.Context, _, _ uint64, input *UpdateOwnerInput) (*model.Owner, error) {
					assert.Equal(t, "080-9999-0000", *input.Phone)
					assert.Nil(t, input.OwnerName)
					return &model.Owner{ID: 1, Phone: *input.Phone}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 401 when clinic_id is missing",
			paramID:    "1",
			body:       map[string]any{},
			setupCtx:   func(_ *gin.Context) {},
			svc:        &mockOwnerService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "returns 400 for non-numeric id",
			paramID:    "xyz",
			body:       map[string]any{"owner_name": "テスト"},
			setupCtx:   func(c *gin.Context) { setClinicID(c) },
			svc:        &mockOwnerService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when owner not found",
			paramID:  "999",
			body:     map[string]any{"owner_name": ptrStr("テスト")},
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockOwnerService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateOwnerInput) (*model.Owner, error) {
					return nil, apperrors.WrapNotFound("owner", "999")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithOwnerSvc(tt.svc)

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			tt.setupCtx(c)

			h.UpdateOwner(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}
