package medicalrecord

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-353: 次回予定日 (next_visit) の request/validation/service/DTO 契約。
// cage clear 系 (EMR-316) と同じ PATCH 規約: 省略は変更なし、clear は NULL、
// clear+値の同時指定は request と service の両方で拒否する。

func TestCreateHospitalizationRequest_NextVisit(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, 8, 10, 0, 0, 0, 0, time.Local)

	newRequest := func(nextVisit *string) *createHospitalizationRequest {
		return &createHospitalizationRequest{
			OwnerID:             1,
			PetID:               2,
			HospitalizationType: string(model.HospitalizationTypeInpatient),
			StartDate:           start,
			EndDate:             end,
			NextVisit:           nextVisit,
		}
	}

	t.Run("YYYY-MM-DD は受理され service input の NextVisit へ入る", func(t *testing.T) {
		value := "2026-11-20"
		input, err := newRequest(&value).toServiceInput()
		require.NoError(t, err)
		require.NotNil(t, input.NextVisit)
		assert.Equal(t, "2026-11-20", input.NextVisit.Format(time.DateOnly))
	})

	t.Run("省略は NextVisit nil のまま", func(t *testing.T) {
		input, err := newRequest(nil).toServiceInput()
		require.NoError(t, err)
		assert.Nil(t, input.NextVisit)
	})

	invalids := []struct {
		name  string
		value string
	}{
		{"空文字", ""},
		{"存在しない日付", "2026-13-40"},
		{"スラッシュ区切り", "2026/11/20"},
		{"RFC3339", "2026-11-20T00:00:00+09:00"},
		{"文字列", "next-month"},
	}
	for _, tt := range invalids {
		t.Run(tt.name+"は invalid input", func(t *testing.T) {
			_, err := newRequest(&tt.value).toServiceInput()
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}

func TestUpdateHospitalizationRequest_NextVisit(t *testing.T) {
	t.Run("next_visit を受理して input.NextVisit へ入れる", func(t *testing.T) {
		value := "2026-11-20"
		input, err := (&updateHospitalizationRequest{NextVisit: &value}).toServiceInput()
		require.NoError(t, err)
		require.NotNil(t, input.NextVisit)
		assert.Equal(t, "2026-11-20", input.NextVisit.Format(time.DateOnly))
		assert.False(t, input.ClearNextVisit)
	})

	t.Run("clear_next_visit のみ受理して ClearNextVisit へ入れる", func(t *testing.T) {
		input, err := (&updateHospitalizationRequest{ClearNextVisit: true}).toServiceInput()
		require.NoError(t, err)
		assert.True(t, input.ClearNextVisit)
		assert.Nil(t, input.NextVisit)
	})

	t.Run("clear_next_visit と next_visit の同時指定は拒否", func(t *testing.T) {
		value := "2026-11-20"
		_, err := (&updateHospitalizationRequest{
			ClearNextVisit: true,
			NextVisit:      &value,
		}).toServiceInput()
		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
	})

	t.Run("省略は何も変えない", func(t *testing.T) {
		input, err := (&updateHospitalizationRequest{}).toServiceInput()
		require.NoError(t, err)
		assert.Nil(t, input.NextVisit)
		assert.False(t, input.ClearNextVisit)
	})

	invalids := []struct {
		name  string
		value string
	}{
		{"空文字", ""},
		{"存在しない日付", "2026-13-40"},
		{"RFC3339", "2026-11-20T00:00:00+09:00"},
	}
	for _, tt := range invalids {
		t.Run(tt.name+"は invalid input", func(t *testing.T) {
			_, err := (&updateHospitalizationRequest{NextVisit: &tt.value}).toServiceInput()
			require.Error(t, err)
			assert.True(t, apperrors.IsInvalidInput(err))
		})
	}
}

func TestBuildHospitalizationUpdate_NextVisit(t *testing.T) {
	nextVisit := time.Date(2026, 11, 20, 0, 0, 0, 0, time.Local)

	t.Run("next_visit 指定は date 値として更新フィールドに入る", func(t *testing.T) {
		fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{NextVisit: &nextVisit})
		got, ok := fields["next_visit"]
		require.True(t, ok)
		assert.Equal(t, nextVisit, got)
	})

	t.Run("clear_next_visit は NULL で更新フィールドに入る", func(t *testing.T) {
		fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{ClearNextVisit: true})
		got, ok := fields["next_visit"]
		require.True(t, ok)
		assert.Nil(t, got)
	})

	t.Run("どちらも未指定なら更新対象外", func(t *testing.T) {
		fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{})
		_, ok := fields["next_visit"]
		assert.False(t, ok)
	})

	t.Run("is_insurance と併用しても next_visit が落ちない", func(t *testing.T) {
		isInsurance := true
		fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{
			IsInsurance: &isInsurance,
			NextVisit:   &nextVisit,
		})
		got, ok := fields["next_visit"]
		require.True(t, ok)
		assert.Equal(t, nextVisit, got)
	})

	t.Run("is_insurance false と併用しても next_visit 設定が落ちない", func(t *testing.T) {
		isInsurance := false
		fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{
			IsInsurance: &isInsurance,
			NextVisit:   &nextVisit,
		})
		got, ok := fields["next_visit"]
		require.True(t, ok)
		assert.Equal(t, nextVisit, got)
		assert.Nil(t, fields["insurance_company_name"])
	})

	t.Run("is_insurance false と併用しても clear_next_visit が落ちない", func(t *testing.T) {
		isInsurance := false
		fields := buildHospitalizationUpdate(&UpdateHospitalizationInput{
			IsInsurance:    &isInsurance,
			ClearNextVisit: true,
		})
		got, ok := fields["next_visit"]
		require.True(t, ok)
		assert.Nil(t, got)
		assert.Nil(t, fields["insurance_number"])
	})
}

func TestHospitalizationService_Update_NextVisit(t *testing.T) {
	nextVisit := time.Date(2026, 11, 20, 0, 0, 0, 0, time.Local)

	t.Run("clear_next_visit と next_visit の同時指定は direct caller でも拒否", func(t *testing.T) {
		repo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Hospitalization, error) {
				t.Fatal("repository must not be reached for contradictory clear+set")
				return nil, nil
			},
		}
		svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

		hosp, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{
			ClearNextVisit: true,
			NextVisit:      &nextVisit,
		})

		require.Error(t, err)
		assert.True(t, apperrors.IsInvalidInput(err))
		assert.Nil(t, hosp)
	})

	t.Run("next_visit は repo へそのまま伝播する", func(t *testing.T) {
		var gotCmd UpdateHospitalizationInput
		repo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{
					ID:        id,
					ClinicID:  clinicID,
					Status:    model.HospitalizationStatusAdmitted,
					StartDate: nextVisit.AddDate(0, -1, 0),
					EndDate:   nextVisit.AddDate(0, 1, 0),
				}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
				gotCmd = cmd
				return &model.Hospitalization{ID: 1, ClinicID: 1, NextVisit: &nextVisit}, nil
			},
		}
		svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

		hosp, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{
			NextVisit: &nextVisit,
		})

		require.NoError(t, err)
		require.NotNil(t, hosp)
		require.NotNil(t, gotCmd.NextVisit)
		assert.Equal(t, nextVisit, *gotCmd.NextVisit)
	})

	t.Run("clear_next_visit は repo へそのまま伝播する", func(t *testing.T) {
		var gotCmd UpdateHospitalizationInput
		repo := &mockHospitalizationRepository{
			findByIDFn: func(_ context.Context, clinicID, id uint64) (*model.Hospitalization, error) {
				return &model.Hospitalization{
					ID:        id,
					ClinicID:  clinicID,
					Status:    model.HospitalizationStatusAdmitted,
					StartDate: nextVisit.AddDate(0, -1, 0),
					EndDate:   nextVisit.AddDate(0, 1, 0),
					NextVisit: &nextVisit,
				}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, cmd UpdateHospitalizationInput) (*model.Hospitalization, error) {
				gotCmd = cmd
				return &model.Hospitalization{ID: 1, ClinicID: 1}, nil
			},
		}
		svc := NewHospitalizationService(repo, nil, nil, nil, nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

		hosp, err := svc.Update(context.Background(), 1, 1, &UpdateHospitalizationInput{
			ClearNextVisit: true,
		})

		require.NoError(t, err)
		require.NotNil(t, hosp)
		assert.True(t, gotCmd.ClearNextVisit)
	})
}

func TestHospitalizationService_Create_PropagatesNextVisit(t *testing.T) {
	now := time.Now()
	nextVisit := time.Date(2026, 11, 20, 0, 0, 0, 0, time.Local)
	var created *model.Hospitalization

	repo := &mockHospitalizationRepository{
		createFn: func(_ context.Context, h *model.Hospitalization) error {
			created = h
			return nil
		},
	}
	svc := NewHospitalizationService(repo, &mockReservationRepository{
		assertOwnerInClinicFn: func(_ context.Context, _, _ uint64) error { return nil },
		findPetOwnerInClinicFn: func(_ context.Context, _, _ uint64) (uint64, error) {
			return 2, nil
		},
	}, &mockPetRepository{
		findByIDFn: func(_ context.Context, _, id uint64) (*model.Pet, error) {
			return &model.Pet{ID: id}, nil
		},
	}, acceptAnyCageRepo(), nil, nil, nil, &mockTransactor{}, WithPetLifeStateReader(stubPetLifeReader{}))

	hosp, err := svc.Create(context.Background(), 1, &CreateHospitalizationInput{
		OwnerID:             2,
		PetID:               5,
		HospitalizationType: model.HospitalizationTypeInpatient,
		StartDate:           now,
		EndDate:             now.AddDate(0, 0, 7),
		NextVisit:           &nextVisit,
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.NextVisit)
	assert.Equal(t, nextVisit, *created.NextVisit)
	require.NotNil(t, hosp.NextVisit)
	assert.Equal(t, nextVisit, *hosp.NextVisit)
}

func TestToHospitalizationResponse_NextVisit(t *testing.T) {
	t.Run("DATE 値を date-only 文字列で返す", func(t *testing.T) {
		nextVisit := time.Date(2026, 11, 20, 0, 0, 0, 0, time.Local)
		res := toHospitalizationResponse(&model.Hospitalization{NextVisit: &nextVisit})
		require.NotNil(t, res.NextVisit)
		assert.Equal(t, "2026-11-20", *res.NextVisit)
	})

	t.Run("UTC midnight の DATE でもカレンダー日がずれない", func(t *testing.T) {
		nextVisit := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
		res := toHospitalizationResponse(&model.Hospitalization{NextVisit: &nextVisit})
		require.NotNil(t, res.NextVisit)
		assert.Equal(t, "2026-11-20", *res.NextVisit)
	})

	t.Run("未設定なら nil（wire では省略）", func(t *testing.T) {
		res := toHospitalizationResponse(&model.Hospitalization{})
		assert.Nil(t, res.NextVisit)
	})
}

// EMR-353 R353-02: HTTP binding/error mapping 経路の回帰。request struct の直接テストは
// ShouldBindJSON の型不一致・handler の RespondError 写像を通さないため、ここで
// POST/PATCH の実 wire JSON から service input と response JSON までを確認する。

func TestCreateHospitalization_NextVisit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 有効な必須項目を固定した生 JSON に、制御されたフィールドスニペット
	// （例: `,"next_visit":"2026-11-20"`）を連結する。実 wire をそのまま送るため
	// null/数値/オブジェクトの型不一致も実際の JSON で表現できる。
	const baseBodyJSON = `{"owner_id":1,"pet_id":2,"hospitalization_type":"hospitalization","start_date":"2026-05-28T00:00:00Z","end_date":"2026-05-30T00:00:00Z"}`

	newBody := func(fieldSnippet string) string {
		return baseBodyJSON[:len(baseBodyJSON)-1] + fieldSnippet + "}"
	}

	newRequest := func(body string) (*httptest.ResponseRecorder, *gin.Context) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		setClinicID(c)
		return w, c
	}

	t.Run("POST next_visit は service input へ届き response は date-only", func(t *testing.T) {
		nextVisit := time.Date(2026, 11, 20, 0, 0, 0, 0, time.Local)
		svc := &mockHospitalizationService{
			createFn: func(_ context.Context, clinicID uint64, input *CreateHospitalizationInput) (*model.Hospitalization, error) {
				assert.Equal(t, uint64(1), clinicID)
				require.NotNil(t, input.NextVisit)
				assert.Equal(t, "2026-11-20", input.NextVisit.Format(time.DateOnly))
				return &model.Hospitalization{ID: 10, ClinicID: clinicID, NextVisit: &nextVisit}, nil
			},
		}
		h := newHandlerWithHospitalizationSvc(svc)

		w, c := newRequest(newBody(`,"next_visit":"2026-11-20"`))
		h.CreateHospitalization(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"next_visit":"2026-11-20"`)
	})

	t.Run("next_visit を返さない response は key 自体を含まない", func(t *testing.T) {
		svc := &mockHospitalizationService{
			createFn: func(_ context.Context, clinicID uint64, _ *CreateHospitalizationInput) (*model.Hospitalization, error) {
				return &model.Hospitalization{ID: 11, ClinicID: clinicID}, nil
			},
		}
		h := newHandlerWithHospitalizationSvc(svc)

		w, c := newRequest(newBody(""))
		h.CreateHospitalization(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.NotContains(t, w.Body.String(), "next_visit")
	})

	t.Run("POST next_visit が空文字なら 400 で service は呼ばれない", func(t *testing.T) {
		svc := &mockHospitalizationService{
			createFn: func(_ context.Context, _ uint64, _ *CreateHospitalizationInput) (*model.Hospitalization, error) {
				t.Fatal("service.Create must not run for invalid next_visit")
				return nil, nil
			},
		}
		h := newHandlerWithHospitalizationSvc(svc)

		w, c := newRequest(newBody(`,"next_visit":""`))
		h.CreateHospitalization(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateHospitalization_NextVisit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRequest := func(body string) (*httptest.ResponseRecorder, *gin.Context) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "1"}}
		setClinicID(c)
		return w, c
	}

	t.Run("PATCH next_visit=null は「変更なし」: NextVisit nil / ClearNextVisit false", func(t *testing.T) {
		nextVisit := time.Date(2026, 11, 20, 0, 0, 0, 0, time.Local)
		svc := &mockHospitalizationService{
			updateFn: func(_ context.Context, clinicID, id uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.Equal(t, uint64(1), id)
				assert.Nil(t, input.NextVisit)
				assert.False(t, input.ClearNextVisit)
				require.NotNil(t, input.Memo)
				assert.Equal(t, "経過メモ", *input.Memo)
				return &model.Hospitalization{ID: id, ClinicID: clinicID, NextVisit: &nextVisit}, nil
			},
		}
		h := newHandlerWithHospitalizationSvc(svc)

		w, c := newRequest(`{"next_visit":null,"memo":"経過メモ"}`)
		h.UpdateHospitalization(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"next_visit":"2026-11-20"`)
	})

	t.Run("PATCH next_visit=日付 は service input へ届く", func(t *testing.T) {
		svc := &mockHospitalizationService{
			updateFn: func(_ context.Context, clinicID, _ uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error) {
				assert.Equal(t, uint64(1), clinicID)
				require.NotNil(t, input.NextVisit)
				assert.Equal(t, "2026-12-01", input.NextVisit.Format(time.DateOnly))
				assert.False(t, input.ClearNextVisit)
				return &model.Hospitalization{ID: 1, ClinicID: clinicID, NextVisit: input.NextVisit}, nil
			},
		}
		h := newHandlerWithHospitalizationSvc(svc)

		w, c := newRequest(`{"next_visit":"2026-12-01"}`)
		h.UpdateHospitalization(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"next_visit":"2026-12-01"`)
	})

	t.Run("PATCH clear_next_visit=true は ClearNextVisit へ届く", func(t *testing.T) {
		svc := &mockHospitalizationService{
			updateFn: func(_ context.Context, clinicID, _ uint64, input *UpdateHospitalizationInput) (*model.Hospitalization, error) {
				assert.Equal(t, uint64(1), clinicID)
				assert.True(t, input.ClearNextVisit)
				assert.Nil(t, input.NextVisit)
				return &model.Hospitalization{ID: 1, ClinicID: clinicID}, nil
			},
		}
		h := newHandlerWithHospitalizationSvc(svc)

		w, c := newRequest(`{"clear_next_visit":true}`)
		h.UpdateHospitalization(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), "next_visit")
	})

	// 不正入力は 400 で service.Update を呼ばないことも固定する。
	invalids := []struct {
		name string
		body string
	}{
		{"clear_next_visit と next_visit の同時指定", `{"clear_next_visit":true,"next_visit":"2026-11-20"}`},
		{"空文字", `{"next_visit":""}`},
		{"スラッシュ区切り", `{"next_visit":"2026/11/20"}`},
		{"RFC3339", `{"next_visit":"2026-11-20T00:00:00+09:00"}`},
		{"数値", `{"next_visit":20261120}`},
		{"オブジェクト", `{"next_visit":{"year":2026}}`},
	}
	for _, tt := range invalids {
		t.Run(tt.name+"は 400 で service は呼ばれない", func(t *testing.T) {
			svc := &mockHospitalizationService{
				updateFn: func(_ context.Context, _, _ uint64, _ *UpdateHospitalizationInput) (*model.Hospitalization, error) {
					t.Fatal("service.Update must not run for invalid next_visit")
					return nil, nil
				},
			}
			h := newHandlerWithHospitalizationSvc(svc)

			w, c := newRequest(tt.body)
			h.UpdateHospitalization(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}
