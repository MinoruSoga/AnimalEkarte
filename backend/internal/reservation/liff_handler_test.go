package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"errors"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ================================================================
// mockLiffService — LiffService インターフェースのモック実装
// ================================================================

type mockLiffService struct {
	getSettingsFn        func(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	getProfileFn         func(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error)
	getCoursesFn         func(ctx context.Context, clinicID uint64) ([]model.ReservationType, error)
	getTrimmingCoursesFn func(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
	getTrimmingOptionsFn func(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
	getStaffsFn          func(ctx context.Context, clinicID, typeID uint64) ([]model.Staff, error)
	getAvailableDatesFn  func(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error)
	getAvailableTimesFn  func(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error)
	createReservationFn  func(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Reservation, error)
	getMyReservationsFn  func(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error)
	cancelReservationFn  func(ctx context.Context, clinicID, customerID, reservationID uint64) error
	getHealthCardFn      func(ctx context.Context, clinicID, customerID uint64) (*HealthCardResult, error)
}

func (m *mockLiffService) GetSettings(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	if m.getSettingsFn != nil {
		return m.getSettingsFn(ctx, clinicID)
	}
	return &model.LineReservationSetting{ClinicID: clinicID, Status: "active"}, nil
}

func (m *mockLiffService) GetProfile(ctx context.Context, clinicID, customerID uint64) (*model.LineCustomer, error) {
	if m.getProfileFn != nil {
		return m.getProfileFn(ctx, clinicID, customerID)
	}
	return &model.LineCustomer{ID: customerID, ClinicID: clinicID}, nil
}

func (m *mockLiffService) GetCourses(ctx context.Context, clinicID uint64) ([]model.ReservationType, error) {
	if m.getCoursesFn != nil {
		return m.getCoursesFn(ctx, clinicID)
	}
	return []model.ReservationType{}, nil
}

func (m *mockLiffService) GetTrimmingCourses(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
	if m.getTrimmingCoursesFn != nil {
		return m.getTrimmingCoursesFn(ctx, clinicID)
	}
	return []model.TrimmingCourse{}, nil
}

func (m *mockLiffService) GetTrimmingOptions(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
	if m.getTrimmingOptionsFn != nil {
		return m.getTrimmingOptionsFn(ctx, clinicID)
	}
	return []model.TrimmingOption{}, nil
}

func (m *mockLiffService) GetStaffs(ctx context.Context, clinicID, typeID uint64) ([]model.Staff, error) {
	if m.getStaffsFn != nil {
		return m.getStaffsFn(ctx, clinicID, typeID)
	}
	return []model.Staff{}, nil
}

func (m *mockLiffService) GetAvailableDates(ctx context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error) {
	if m.getAvailableDatesFn != nil {
		return m.getAvailableDatesFn(ctx, clinicID, typeID, staffID)
	}
	return []AvailableDateResult{}, BookingWindow{}, nil
}

func (m *mockLiffService) GetAvailableTimes(ctx context.Context, clinicID, typeID, staffID uint64, date time.Time) ([]TimeSlot, error) {
	if m.getAvailableTimesFn != nil {
		return m.getAvailableTimesFn(ctx, clinicID, typeID, staffID, date)
	}
	return []TimeSlot{}, nil
}

func (m *mockLiffService) CreateReservation(ctx context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Reservation, error) {
	if m.createReservationFn != nil {
		return m.createReservationFn(ctx, clinicID, customerID, input)
	}
	return &model.Reservation{ID: 1, ClinicID: clinicID}, nil
}

func (m *mockLiffService) GetMyReservations(ctx context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
	if m.getMyReservationsFn != nil {
		return m.getMyReservationsFn(ctx, clinicID, customerID)
	}
	return []model.Reservation{}, nil
}

func (m *mockLiffService) CancelReservation(ctx context.Context, clinicID, customerID, reservationID uint64) error {
	if m.cancelReservationFn != nil {
		return m.cancelReservationFn(ctx, clinicID, customerID, reservationID)
	}
	return nil
}

func (m *mockLiffService) GetHealthCard(ctx context.Context, clinicID, customerID uint64) (*HealthCardResult, error) {
	if m.getHealthCardFn != nil {
		return m.getHealthCardFn(ctx, clinicID, customerID)
	}
	return &HealthCardResult{Pets: []PetHealthCard{}}, nil
}

// ================================================================
// テスト用ヘルパー
// ================================================================

// newLiffHandler は LIFF テスト用の Handler を返す。
func newLiffHandler(liff LiffService) *LiffHandler {
	return NewLiffHandler(liff, &mockStaffClinicAssignmentService{})
}

// newLiffRouter は LIFF ハンドラーのテスト用 Gin ルーターを返す。
// customerID=1 を context にセットした状態（LiffAuth 通過済み想定）。
func newLiffRouter(h *LiffHandler, withCustomer bool) *gin.Engine {
	r := gin.New()

	// LiffAuth 通過済みを模倣するミドルウェア
	authMW := func(c *gin.Context) {
		if withCustomer {
			c.Set("liff_customer_id", uint64(1)) // = middleware.LiffCustomerIDKey()（契約: extractLiffCustomerID 複製と同一キー）
		}
		c.Next()
	}

	liff := r.Group("/api/liff/:clinicId")
	liff.Use(authMW)
	liff.GET("/settings", h.GetLiffSettings)
	liff.GET("/profile", h.GetLiffProfile)
	liff.GET("/courses", h.GetLiffTypes)
	liff.GET("/trimming-courses", h.GetLiffTrimmingCourses)
	liff.GET("/trimming-options", h.GetLiffTrimmingOptions)
	liff.GET("/staffs", h.GetLiffStaffs)
	liff.GET("/available-dates", h.GetLiffAvailableDates)
	liff.GET("/available-times", h.GetLiffAvailableTimes)
	liff.POST("/reservations", h.CreateLiffReservation)
	liff.GET("/my-reservations", h.GetLiffMyReservations)
	liff.DELETE("/my-reservations/:id", h.CancelLiffReservation)
	liff.GET("/health-card", h.GetLiffHealthCard)
	return r
}

// doLiffRequest はテスト用 HTTP リクエストを発行し、レコーダーを返す。
func doLiffRequest(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ================================================================
// GetLiffSettings テスト
// ================================================================

func TestGetLiffSettings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: 200 + 設定情報", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getSettingsFn: func(_ context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
				assert.Equal(t, uint64(3), clinicID)
				return &model.LineReservationSetting{ClinicID: 3, Status: "active", PhoneNumber: "052-000-0000"}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/settings", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "active")
	})

	t.Run("clinicId が数値でない → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/abc/settings", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getSettingsFn: func(_ context.Context, _ uint64) (*model.LineReservationSetting, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/settings", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ================================================================
// GetLiffProfile テスト
// ================================================================

func TestGetLiffProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: 200 + プロフィール", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getProfileFn: func(_ context.Context, clinicID, customerID uint64) (*model.LineCustomer, error) {
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(1), customerID)
				return &model.LineCustomer{
					ID:               1,
					LineUserID:       "line-user-1",
					DisplayName:      "田中太郎",
					AdditionalFields: []byte(`{"phone":"090-1234-5678","owner_name":"田中太郎","pets":[{"name":"ポチ","type":"柴犬","is_new":true}]}`),
					Owner: &model.Owner{
						Name:  "田中太郎",
						Phone: "090-1234-5678",
						Pets: []model.Pet{
							{
								ID:   10,
								Name: "ポチ",
								AnimalSpecies: &model.AnimalSpecies{
									Name: "犬",
								},
							},
						},
					},
				}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodGet, "/api/liff/3/profile", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "田中太郎")
		assert.Contains(t, w.Body.String(), `"additional_fields":{"phone":"090-1234-5678"`)
		assert.Contains(t, w.Body.String(), `"owner_name":"田中太郎"`)
		assert.Contains(t, w.Body.String(), `"animal_species":{"name":"犬"}`)
	})

	t.Run("customerID が context にない → 401", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/profile", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("サービスが NotFound → 404", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getProfileFn: func(_ context.Context, _, _ uint64) (*model.LineCustomer, error) {
				return nil, apperrors.ErrNotFound
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodGet, "/api/liff/3/profile", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// ================================================================
// GetLiffTypes テスト
// ================================================================

func TestGetLiffTypes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: コース一覧を返す", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getCoursesFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
				return []model.ReservationType{
					{ID: 1, Name: "一般診察"},
					{ID: 2, Name: "ワクチン接種"},
				}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/courses", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp []map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Len(t, resp, 2)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getCoursesFn: func(_ context.Context, _ uint64) ([]model.ReservationType, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/courses", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("clinicId が数値でない → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/abc/courses", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ================================================================
// GetLiffTrimmingCourses テスト
// ================================================================

func TestGetLiffTrimmingCourses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: トリミングコース一覧を返す", func(t *testing.T) {
		price := int64(5000)
		h := newLiffHandler(&mockLiffService{
			getTrimmingCoursesFn: func(_ context.Context, clinicID uint64) ([]model.TrimmingCourse, error) {
				assert.Equal(t, uint64(3), clinicID)
				return []model.TrimmingCourse{
					{ID: 1, Name: "フルコース", Price: &price},
				}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/trimming-courses", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "フルコース")
	})

	t.Run("clinicId が数値でない → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/abc/trimming-courses", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getTrimmingCoursesFn: func(_ context.Context, _ uint64) ([]model.TrimmingCourse, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/trimming-courses", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ================================================================
// GetLiffTrimmingOptions テスト
// ================================================================

func TestGetLiffTrimmingOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: トリミングオプション一覧を返す", func(t *testing.T) {
		price := int64(800)
		h := newLiffHandler(&mockLiffService{
			getTrimmingOptionsFn: func(_ context.Context, clinicID uint64) ([]model.TrimmingOption, error) {
				assert.Equal(t, uint64(3), clinicID)
				return []model.TrimmingOption{
					{ID: 1, Name: "爪切り", Price: &price},
				}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/trimming-options", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "爪切り")
	})

	t.Run("clinicId が数値でない → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/abc/trimming-options", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getTrimmingOptionsFn: func(_ context.Context, _ uint64) ([]model.TrimmingOption, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/trimming-options", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ================================================================
// GetLiffStaffs テスト
// ================================================================

func TestGetLiffStaffs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: スタッフ一覧を返す", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getStaffsFn: func(_ context.Context, clinicID, typeID uint64) ([]model.Staff, error) {
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(1), typeID)
				return []model.Staff{{ID: 10, Name: "林先生"}}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/staffs?courseId=1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "林先生")
	})

	t.Run("courseId が未指定 → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/staffs", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("courseId がゼロ → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/staffs?courseId=0", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("courseId が文字列 → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/staffs?courseId=abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getStaffsFn: func(_ context.Context, _, _ uint64) ([]model.Staff, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/staffs?courseId=1", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ================================================================
// GetLiffAvailableDates テスト
// ================================================================

func TestGetLiffAvailableDates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: 予約可能日付一覧を返す", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getAvailableDatesFn: func(_ context.Context, clinicID, typeID, staffID uint64) ([]AvailableDateResult, BookingWindow, error) {
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(1), typeID)
				assert.Equal(t, uint64(0), staffID) // 指名なし
				return []AvailableDateResult{
					{Date: "2026-05-01", Weekday: 4, Available: true}, // 4=木曜日
				}, BookingWindow{}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-dates?courseId=1", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "2026-05-01")
	})

	t.Run("staffId=5 を指定したとき正しく渡る", func(t *testing.T) {
		var capturedStaffID uint64
		h := newLiffHandler(&mockLiffService{
			getAvailableDatesFn: func(_ context.Context, _, _, staffID uint64) ([]AvailableDateResult, BookingWindow, error) {
				capturedStaffID = staffID
				return nil, BookingWindow{}, nil
			},
		})
		doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-dates?courseId=1&staffId=5", nil)
		assert.Equal(t, uint64(5), capturedStaffID)
	})

	t.Run("courseId が未指定 → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-dates", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getAvailableDatesFn: func(_ context.Context, _, _, _ uint64) ([]AvailableDateResult, BookingWindow, error) {
				return nil, BookingWindow{}, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-dates?courseId=1", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ================================================================
// GetLiffAvailableTimes テスト
// ================================================================

func TestGetLiffAvailableTimes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: 時間枠一覧を返す", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getAvailableTimesFn: func(_ context.Context, _, _, _ uint64, date time.Time) ([]TimeSlot, error) {
				assert.Equal(t, 2026, date.Year())
				assert.Equal(t, time.May, date.Month())
				assert.Equal(t, 1, date.Day())
				return []TimeSlot{
					{StartTime: "1000", EndTime: "1015"},
					{StartTime: "1015", EndTime: "1030"},
				}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-times?courseId=1&date=2026-05-01", nil)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "1000")
	})

	t.Run("courseId が未指定 → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-times?date=2026-05-01", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("date が不正フォーマット → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-times?courseId=1&date=20260501", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("date が未指定 → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-times?courseId=1", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getAvailableTimesFn: func(_ context.Context, _, _, _ uint64, _ time.Time) ([]TimeSlot, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/available-times?courseId=1&date=2026-05-01", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ================================================================
// CreateLiffReservation テスト
// ================================================================

func TestCreateLiffReservation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validBody := map[string]any{
		"course_id":  1,
		"staff_id":   10,
		"date":       "2026-05-01",
		"start_time": "1000",
		"end_time":   "1015",
		"customer_fields": map[string]string{
			"phone": "090-1234-5678",
		},
	}

	t.Run("正常系: 201 + 予約ID", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			createReservationFn: func(_ context.Context, clinicID, customerID uint64, input *CreateReservationInput) (*model.Reservation, error) {
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(1), customerID)
				assert.Equal(t, uint64(10), input.StaffID)
				return &model.Reservation{ID: 42}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodPost, "/api/liff/3/reservations", validBody)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "/api/v1/reservations/42", w.Header().Get("Location"))

		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, map[string]any{"id": float64(42), "notes": ""}, body)
	})

	t.Run("customerID が context にない → 401", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodPost, "/api/liff/3/reservations", validBody)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("リクエストボディが不正 → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		req := httptest.NewRequest(http.MethodPost, "/api/liff/3/reservations", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		newLiffRouter(h, true).ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("date が不正フォーマット → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		body := map[string]any{"course_id": 1, "date": "20260501", "start_time": "1000", "end_time": "1015"}
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodPost, "/api/liff/3/reservations", body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ReservationLimitError → 409 + redirect_step", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			createReservationFn: func(_ context.Context, _, _ uint64, _ *CreateReservationInput) (*model.Reservation, error) {
				return nil, &ReservationLimitError{Code: "SLOT_TAKEN", Message: "満員", RedirectStep: 5}
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodPost, "/api/liff/3/reservations", validBody)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "SLOT_TAKEN")
		assert.Contains(t, w.Body.String(), "redirect_step")
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			createReservationFn: func(_ context.Context, _, _ uint64, _ *CreateReservationInput) (*model.Reservation, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodPost, "/api/liff/3/reservations", validBody)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("clinicId が数値でない → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodPost, "/api/liff/abc/reservations", validBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("request_text が長過ぎる → 400 (BUG-LINE-012)", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		body := map[string]any{
			"course_id":    1,
			"date":         "2026-05-01",
			"start_time":   "1000",
			"end_time":     "1015",
			"request_text": strings.Repeat("a", maxRequestTextLength+1),
		}
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodPost, "/api/liff/3/reservations", body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("指名スタッフが当該クリニックに未所属 → 400", func(t *testing.T) {
		// mockStaffClinicAssignmentService はクリニック 1, 3 のみ所属を返すため、
		// clinicId=99 を指定すると checkDoctorClinicAssignment が失敗する。
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodPost, "/api/liff/99/reservations", validBody)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ================================================================
// GetLiffMyReservations テスト
// ================================================================

func TestGetLiffMyReservations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: 予約一覧を返す", func(t *testing.T) {
		now := time.Now()
		h := newLiffHandler(&mockLiffService{
			getMyReservationsFn: func(_ context.Context, clinicID, customerID uint64) ([]model.Reservation, error) {
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(1), customerID)
				return []model.Reservation{
					{ID: 10, ClinicID: 3, StartTime: now, EndTime: now.Add(15 * time.Minute)},
				}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodGet, "/api/liff/3/my-reservations", nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp []map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Len(t, resp, 1)
	})

	t.Run("customerID が context にない → 401", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/my-reservations", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getMyReservationsFn: func(_ context.Context, _, _ uint64) ([]model.Reservation, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodGet, "/api/liff/3/my-reservations", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ================================================================
// CancelLiffReservation テスト
// ================================================================

func TestCancelLiffReservation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: 204 No Content", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			cancelReservationFn: func(_ context.Context, clinicID, customerID, reservationID uint64) error {
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(1), customerID)
				assert.Equal(t, uint64(10), reservationID)
				return nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodDelete, "/api/liff/3/my-reservations/10", nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("customerID が context にない → 401", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodDelete, "/api/liff/3/my-reservations/10", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("id が数値でない → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodDelete, "/api/liff/3/my-reservations/abc", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("id がゼロ → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodDelete, "/api/liff/3/my-reservations/0", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("予約が見つからない → 404", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			cancelReservationFn: func(_ context.Context, _, _, _ uint64) error {
				return apperrors.ErrNotFound
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodDelete, "/api/liff/3/my-reservations/999", nil)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			cancelReservationFn: func(_ context.Context, _, _, _ uint64) error {
				return errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodDelete, "/api/liff/3/my-reservations/10", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ================================================================
// GetLiffHealthCard テスト
// ================================================================

func TestGetLiffHealthCard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("正常系: 200 + 健康手帳データ（pet_id は文字列、nullable フィールドは明示 null）", func(t *testing.T) {
		vaccinatedAt := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		h := newLiffHandler(&mockLiffService{
			getHealthCardFn: func(_ context.Context, clinicID, customerID uint64) (*HealthCardResult, error) {
				assert.Equal(t, uint64(3), clinicID)
				assert.Equal(t, uint64(1), customerID)
				return &HealthCardResult{
					OwnerName: "田中太郎",
					Pets: []PetHealthCard{
						{
							PetID:         10,
							PetName:       "ポチ",
							Species:       "犬",
							Breed:         "柴犬",
							LastVisitDate: nil,
							Vaccines: []VaccineRecord{
								{VaccineName: "狂犬病", VaccinatedAt: vaccinatedAt, NextDueAt: nil},
							},
						},
					},
				}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodGet, "/api/liff/3/health-card", nil)
		require.Equal(t, http.StatusOK, w.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "田中太郎", body["owner_name"])

		pets, ok := body["pets"].([]any)
		require.True(t, ok)
		require.Len(t, pets, 1)
		pet, ok := pets[0].(map[string]any)
		require.True(t, ok)

		// pet_id は数値ではなく文字列でシリアライズされる（uint64精度落ち対策）
		petID, ok := pet["pet_id"].(string)
		require.True(t, ok, "pet_id should be a JSON string, got %T", pet["pet_id"])
		assert.Equal(t, "10", petID)
		assert.Equal(t, "ポチ", pet["pet_name"])
		assert.Equal(t, "犬", pet["species"])
		assert.Equal(t, "柴犬", pet["breed"])

		// nullable フィールドはキー省略ではなく明示 null であること
		lastVisitVal, hasLastVisit := pet["last_visit_date"]
		assert.True(t, hasLastVisit, "last_visit_date key must be present")
		assert.Nil(t, lastVisitVal)

		vaccines, ok := pet["vaccines"].([]any)
		require.True(t, ok)
		require.Len(t, vaccines, 1)
		vaccine, ok := vaccines[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "狂犬病", vaccine["vaccine_name"])
		nextDueVal, hasNextDue := vaccine["next_due_at"]
		assert.True(t, hasNextDue, "next_due_at key must be present")
		assert.Nil(t, nextDueVal)

		// raw body でも null が省略されず出力されていることを確認
		assert.Contains(t, w.Body.String(), `"next_due_at":null`)
		assert.Contains(t, w.Body.String(), `"last_visit_date":null`)
	})

	t.Run("正常系: last_visit_date は date-only 文字列で返る（RFC3339 生表示にならない）", func(t *testing.T) {
		lastVisit := time.Date(2026, 5, 1, 12, 34, 56, 0, time.UTC)
		h := newLiffHandler(&mockLiffService{
			getHealthCardFn: func(_ context.Context, _, _ uint64) (*HealthCardResult, error) {
				return &HealthCardResult{
					OwnerName: "田中太郎",
					Pets: []PetHealthCard{
						{
							PetID:         10,
							PetName:       "ポチ",
							Species:       "犬",
							Breed:         "柴犬",
							LastVisitDate: &lastVisit,
						},
					},
				}, nil
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodGet, "/api/liff/3/health-card", nil)
		require.Equal(t, http.StatusOK, w.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		pets, ok := body["pets"].([]any)
		require.True(t, ok)
		require.Len(t, pets, 1)
		pet, ok := pets[0].(map[string]any)
		require.True(t, ok)

		assert.Equal(t, "2026-05-01", pet["last_visit_date"], "last_visit_date は YYYY-MM-DD の日付のみで、時刻を含んではならない")
	})

	t.Run("customerID が context にない → 401", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, false), http.MethodGet, "/api/liff/3/health-card", nil)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("サービスエラー → 500", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{
			getHealthCardFn: func(_ context.Context, _, _ uint64) (*HealthCardResult, error) {
				return nil, errors.New("internal error")
			},
		})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodGet, "/api/liff/3/health-card", nil)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("clinicId が数値でない → 400", func(t *testing.T) {
		h := newLiffHandler(&mockLiffService{})
		w := doLiffRequest(t, newLiffRouter(h, true), http.MethodGet, "/api/liff/abc/health-card", nil)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
