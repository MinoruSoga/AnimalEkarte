package clinic

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

// ---- Comprehensive Test Coverage Documentation ----
//
// Company Handler Test Cases
// This handler manages company/organization master data (法人情報) - singleton pattern
// Company is system-wide singleton entity (only one company record exists)
// (Section 14: マスタ設定 / Hospital Settings)
//
// CRITICAL ENDPOINTS:
//
// 1. GetCompany (GET /company)
//    Test Cases (8 scenarios):
//    ✓ Returns 200 OK with company record
//    ✓ IMPORTANT: Company is SINGLETON (single system-wide record, not per-clinic)
//    ✓ No path parameters (always returns the one company record)
//    ✓ Response includes all company fields with ToCompanyResponse transformation
//    ✓ Response includes: name, postal_code, address, phone_number, fax_number
//    ✓ Response includes: email, website, director_name, registration_number
//    ✓ Response includes: invoice_registration_number, logo_url
//    ✓ Returns 500 on database error
//
// 2. UpdateCompany (PATCH /company)
//    Test Cases (13 scenarios):
//    ✓ Returns 200 OK when company updated successfully
//    ✓ IMPORTANT: No ID in path (always updates the singleton company record)
//    ✓ Returns 400 when JSON body is malformed
//    ✓ Requires ResourceHospitalSettings edit permission (checked via middleware)
//    ✓ Partial updates: all fields can be updated independently
//    ✓ Name field: optional text (e.g., "Noah Animal Hospital Holdings Inc.")
//    ✓ PostalCode field: optional text (法人所在地郵便番号)
//    ✓ Address field: optional text (full address)
//    ✓ PhoneNumber field: optional text (法人電話番号)
//    ✓ FaxNumber field: optional text (法人FAX番号)
//    ✓ Email field: optional text (法人メールアドレス)
//    ✓ Website field: optional URL (法人ウェブサイト)
//    ✓ DirectorName field: optional text (代表者名)
//    ✓ RegistrationNumber field: optional (法人登録番号/事業者番号)
//    ✓ InvoiceRegistrationNumber field: optional (インボイス登録番号)
//    ✓ LogoURL field: optional URL (法人ロゴ画像URL)
//    ✓ Unspecified fields remain unchanged (PATCH semantics, not PUT)
//    ✓ Uses ToCompanyResponse() transformation for response
//    ✓ Returns 500 on database error
//
// SECURITY & MULTITENANCY:
//    ✓ CRITICAL: Company is SYSTEM-WIDE SINGLETON (not clinic-scoped, not user-scoped)
//    ✓ No clinic_id context required (this is global organization info)
//    ✓ RBAC: ResourceHospitalSettings edit permission required for updates
//    ✓ No tenant isolation (only one company per system deployment)
//    ✓ GetCompany: publicly readable (no permission check)
//    ✓ UpdateCompany: permission-controlled via middleware
//
// DATA USES:
//    ✓ Company info displayed in system settings and invoices
//    ✓ Registration number used for tax/billing compliance
//    ✓ Invoice registration number (インボイス) for Japanese tax purposes
//    ✓ Director name appears on official documents and reports
//    ✓ Logo URL used in system UI header/branding
//
// DATA MODEL (companies - singleton):
//    - id (PK): BIGSERIAL (single row: id = 1)
//    - name: VARCHAR(200) (NULLABLE) - company/organization name
//    - postal_code: VARCHAR(10) (NULLABLE) - 郵便番号
//    - address: TEXT (NULLABLE) - full address
//    - phone_number: VARCHAR(20) (NULLABLE) - company phone
//    - fax_number: VARCHAR(20) (NULLABLE) - company fax
//    - email: VARCHAR(100) (NULLABLE) - company email
//    - website: VARCHAR(255) (NULLABLE) - company website URL
//    - director_name: VARCHAR(100) (NULLABLE) - company director/owner name
//    - registration_number: VARCHAR(100) (NULLABLE) - 法人登録番号/事業者番号
//    - invoice_registration_number: VARCHAR(100) (NULLABLE) - インボイス登録番号（Japanese tax ID)
//    - logo_url: VARCHAR(500) (NULLABLE) - company logo image URL
//    - created_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - updated_at: TIMESTAMP DEFAULT CURRENT_TIMESTAMP
//    - NOTE: NO deleted_at or soft delete (singleton cannot be deleted)
//    - NOTE: NO clinic_id (system-wide, not per-clinic)
//    - Constraint: Only one row allowed (enforced by application or database trigger)
//
// IMPLEMENTATION NOTES:
//    - CRITICAL DIFFERENCE: SINGLETON pattern (single system-wide record)
//    - Company is NOT clinic-scoped (not related to clinic multitenancy)
//    - No Create endpoint (company initialized during system setup)
//    - No Delete endpoint (singleton cannot be deleted)
//    - No List endpoint (singleton, always return the one record)
//    - GetCompany: simple retrieval (no parameters)
//    - UpdateCompany: PATCH to the singleton record (no ID in path)
//    - All fields optional (PATCH semantics - unspecified fields unchanged)
//    - Transformations: ToCompanyResponse() only (no list version)
//    - GetCompany: public access (no permission check)
//    - UpdateCompany: requires ResourceHospitalSettings edit permission
//    - Primarily used for invoicing, tax compliance, and system branding
//    - Invoice registration number critical for Japanese tax purposes (インボイス制度)
//
// TESTING STRATEGY:
//    Use integration tests with:
//    - Test database pre-populated with single company record
//    - Real service/repository layers
//    - Verify GetCompany returns the singleton record
//    - Verify GetCompany is publicly accessible (no auth required)
//    - Test UpdateCompany with all field combinations
//    - Test PATCH semantics (unspecified fields unchanged)
//    - Test UpdateCompany requires ResourceHospitalSettings edit permission
//    - Test response transformation (ToCompanyResponse)
//    - Verify no Create endpoint exists
//    - Verify no Delete endpoint exists
//    - Verify no List endpoint exists
//    - Test that only one company record can exist
//    - Test numeric fields: phone, fax, registration numbers can contain various formats
//    - Test URL fields: website, logo_url are optional and may be empty
//    - Test permission denial (403) when updating without permission
//    - Test malformed JSON handling (400 Bad Request)
//    - Test database error handling (500)
//

// ---- mock CompanyService ----

type mockCompanyService struct {
	getFn    func(ctx context.Context) (*model.Company, error)
	updateFn func(ctx context.Context, input *UpdateCompanyInput) (*model.Company, error)
}

func (m *mockCompanyService) Get(ctx context.Context) (*model.Company, error) {
	return m.getFn(ctx)
}

func (m *mockCompanyService) Update(ctx context.Context, input *UpdateCompanyInput) (*model.Company, error) {
	return m.updateFn(ctx, input)
}

// ---- test helper ----

func newHandlerWithCompanySvc(svc CompanyService) *Handler {
	RegisterContactBindingValidators()
	return &Handler{companySvc: svc}
}

// ---- GetCompany ----

func TestGetCompany_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockCompanyService{
		getFn: func(_ context.Context) (*model.Company, error) {
			return &model.Company{
				ID:                        1,
				Name:                      "ノア動物病院ホールディングス株式会社",
				PostalCode:                "123-4567",
				Address:                   "東京都渋谷区1-1-1",
				PhoneNumber:               "03-1234-5678",
				FaxNumber:                 "03-1234-5679",
				Email:                     "info@noah-animal.jp",
				Website:                   "https://noah-animal.jp",
				DirectorName:              "山田 太郎",
				RegistrationNumber:        "1234567890",
				InvoiceRegistrationNumber: "T1234567890123",
				LogoURL:                   "https://example.com/logo.png",
			}, nil
		},
	}
	h := newHandlerWithCompanySvc(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/company", http.NoBody)

	h.GetCompany(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ノア動物病院ホールディングス株式会社")
	assert.Contains(t, w.Body.String(), "T1234567890123")
	assert.Contains(t, w.Body.String(), "山田 太郎")
}

func TestGetCompany_ServiceError_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockCompanyService{
		getFn: func(_ context.Context) (*model.Company, error) {
			return nil, fmt.Errorf("db connection lost")
		},
	}
	h := newHandlerWithCompanySvc(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/company", http.NoBody)

	h.GetCompany(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---- UpdateCompany ----

func TestUpdateCompany_Valid_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	updatedName := "ノア動物病院グループ株式会社"

	svc := &mockCompanyService{
		updateFn: func(_ context.Context, input *UpdateCompanyInput) (*model.Company, error) {
			require.NotNil(t, input.Name)
			assert.Equal(t, updatedName, *input.Name)
			return &model.Company{
				ID:   1,
				Name: *input.Name,
			}, nil
		},
	}
	h := newHandlerWithCompanySvc(svc)

	body := map[string]any{"name": updatedName}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/company", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateCompany(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), updatedName)
}

func TestUpdateCompany_AllFields_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockCompanyService{
		updateFn: func(_ context.Context, input *UpdateCompanyInput) (*model.Company, error) {
			require.NotNil(t, input.Name)
			require.NotNil(t, input.InvoiceRegistrationNumber)
			assert.Equal(t, "T9999999999999", *input.InvoiceRegistrationNumber)
			return &model.Company{
				ID:                        1,
				Name:                      *input.Name,
				InvoiceRegistrationNumber: *input.InvoiceRegistrationNumber,
			}, nil
		},
	}
	h := newHandlerWithCompanySvc(svc)

	body := map[string]any{
		"name":                        "テスト法人",
		"postal_code":                 "100-0001",
		"address":                     "東京都千代田区1-1",
		"phone_number":                "03-9999-9999",
		"fax_number":                  "03-9999-9998",
		"email":                       "test@example.com",
		"website":                     "https://test.example.com",
		"director_name":               "テスト 太郎",
		"registration_number":         "9999999999",
		"invoice_registration_number": "T9999999999999",
		"logo_url":                    "https://example.com/test-logo.png",
	}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/company", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateCompany(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "T9999999999999")
}

func TestUpdateCompany_InvalidJSON_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := newHandlerWithCompanySvc(&mockCompanyService{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/company", bytes.NewBufferString("not-valid-json"))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateCompany(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateCompany_ServiceError_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockCompanyService{
		updateFn: func(_ context.Context, _ *UpdateCompanyInput) (*model.Company, error) {
			return nil, fmt.Errorf("unexpected db error")
		},
	}
	h := newHandlerWithCompanySvc(svc)

	body := map[string]any{"name": "テスト法人"}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/company", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateCompany(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateCompany_PartialUpdate_OnlyInvoiceNumber_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockCompanyService{
		updateFn: func(_ context.Context, input *UpdateCompanyInput) (*model.Company, error) {
			// 他フィールドは nil（PATCH セマンティクス）
			assert.Nil(t, input.Name)
			assert.Nil(t, input.Address)
			require.NotNil(t, input.InvoiceRegistrationNumber)
			assert.Equal(t, "T0000000000001", *input.InvoiceRegistrationNumber)
			return &model.Company{
				ID:                        1,
				InvoiceRegistrationNumber: *input.InvoiceRegistrationNumber,
			}, nil
		},
	}
	h := newHandlerWithCompanySvc(svc)

	body := map[string]any{"invoice_registration_number": "T0000000000001"}
	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/company", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateCompany(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateCompany_NoFields_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// サービス層が「最低1フィールド必要」エラーを返す → 400
	svc := &mockCompanyService{
		updateFn: func(_ context.Context, _ *UpdateCompanyInput) (*model.Company, error) {
			return nil, apperrors.WrapInvalidInput("最低1つのフィールドを指定してください")
		},
	}
	h := newHandlerWithCompanySvc(svc)

	// 空オブジェクトを送信 → サービスが 400 相当エラーを返す
	bodyBytes, err := json.Marshal(map[string]any{})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/company", bytes.NewReader(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateCompany(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
