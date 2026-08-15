package reservation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

// ---- toLiffCourseResponse ----

func TestToLiffCourseResponse(t *testing.T) {
	tests := []struct {
		name     string
		st       *model.ReservationType
		wantName string
	}{
		{
			name: "uses ReservationDisplayName when set",
			st: &model.ReservationType{
				ID:                     1,
				Name:                   "一般診察",
				ShortName:              "一般",
				ShowShortName:          true,
				ReservationDisplayName: "LINE表示名",
				DurationMinutes:        15,
				ReservationComment:     "comment",
				ReservationImageURL:    "https://example.com/x.png",
				SortOrder:              2,
				Category:               model.ReservationTypeCategoryGeneral,
			},
			wantName: "LINE表示名",
		},
		{
			name: "falls back to ShortName when ReservationDisplayName is empty and ShowShortName true",
			st: &model.ReservationType{
				ID:            2,
				Name:          "一般診察",
				ShortName:     "一般",
				ShowShortName: true,
				Category:      model.ReservationTypeCategoryGeneral,
			},
			wantName: "一般",
		},
		{
			name: "falls back to Name when ReservationDisplayName empty and ShowShortName false",
			st: &model.ReservationType{
				ID:            3,
				Name:          "一般診察",
				ShortName:     "一般",
				ShowShortName: false,
				Category:      model.ReservationTypeCategoryGeneral,
			},
			wantName: "一般診察",
		},
		{
			name: "falls back to Name when ShortName is also empty",
			st: &model.ReservationType{
				ID:            4,
				Name:          "ワクチン接種",
				ShortName:     "",
				ShowShortName: true,
				Category:      model.ReservationTypeCategoryGeneral,
			},
			wantName: "ワクチン接種",
		},
		{
			// P1-6 (PR #186 review): category was omitted from the response entirely,
			// so the LIFF reservation flow's course.category === 'trimming' check always
			// fell back to 'general' and skipped the trimming-details step.
			name: "propagates trimming category",
			st: &model.ReservationType{
				ID:            5,
				Name:          "トリミング",
				ShowShortName: false,
				Category:      model.ReservationTypeCategoryTrimming,
			},
			wantName: "トリミング",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toLiffCourseResponse(tt.st)
			assert.Equal(t, tt.st.ID, got.ID)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.st.ShortName, got.ShortName)
			assert.Equal(t, tt.st.ShowShortName, got.ShowShortName)
			assert.Equal(t, tt.st.DurationMinutes, got.DurationMinutes)
			assert.Equal(t, tt.st.ReservationComment, got.ReservationComment)
			assert.Equal(t, tt.st.ReservationImageURL, got.ReservationImageURL)
			assert.Equal(t, tt.st.SortOrder, got.SortOrder)
			assert.Equal(t, string(tt.st.Category), got.Category)
		})
	}
}

// ---- toLiffReservationResponse ----

func TestToLiffReservationResponse(t *testing.T) {
	start := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 1, 10, 15, 0, 0, time.UTC)
	created := time.Date(2026, 4, 25, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		r             *model.Reservation
		wantCourse    string
		wantStaffName string
	}{
		{
			name: "no ReservationType and no Doctor relations",
			r: &model.Reservation{
				ID:        1,
				StartTime: start,
				EndTime:   end,
				Status:    model.ReservationStatusPending,
				Notes:     "初診",
				CreatedAt: created,
			},
			wantCourse:    "",
			wantStaffName: "",
		},
		{
			name: "ReservationType uses ReservationDisplayName",
			r: &model.Reservation{
				ID:        2,
				StartTime: start,
				EndTime:   end,
				Status:    model.ReservationStatusConfirmed,
				ReservationType: &model.ReservationType{
					Name:                   "一般診察",
					ShortName:              "一般",
					ShowShortName:          true,
					ReservationDisplayName: "LINE表示名",
				},
			},
			wantCourse:    "LINE表示名",
			wantStaffName: "",
		},
		{
			name: "ReservationType falls back to ShortName",
			r: &model.Reservation{
				ID:        3,
				StartTime: start,
				EndTime:   end,
				ReservationType: &model.ReservationType{
					Name:          "一般診察",
					ShortName:     "一般",
					ShowShortName: true,
				},
			},
			wantCourse:    "一般",
			wantStaffName: "",
		},
		{
			name: "ReservationType falls back to Name when ShowShortName false",
			r: &model.Reservation{
				ID:        4,
				StartTime: start,
				EndTime:   end,
				ReservationType: &model.ReservationType{
					Name:          "一般診察",
					ShortName:     "一般",
					ShowShortName: false,
				},
			},
			wantCourse:    "一般診察",
			wantStaffName: "",
		},
		{
			name: "Doctor uses ReservationDisplayName",
			r: &model.Reservation{
				ID:        5,
				StartTime: start,
				EndTime:   end,
				Doctor: &model.Staff{
					Name:                   "山田太郎",
					ReservationDisplayName: "山田先生",
				},
			},
			wantCourse:    "",
			wantStaffName: "山田先生",
		},
		{
			name: "Doctor falls back to Name when ReservationDisplayName empty",
			r: &model.Reservation{
				ID:        6,
				StartTime: start,
				EndTime:   end,
				Doctor: &model.Staff{
					Name: "山田太郎",
				},
			},
			wantCourse:    "",
			wantStaffName: "山田太郎",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toLiffReservationResponse(tt.r)
			assert.Equal(t, tt.r.ID, got.ID)
			assert.Equal(t, tt.r.StartTime.In(time.Local).Format("2006-01-02"), got.Date)
			assert.Equal(t, tt.r.StartTime.In(time.Local).Format("15:04"), got.StartTime)
			assert.Equal(t, tt.r.EndTime.In(time.Local).Format("15:04"), got.EndTime)
			assert.Equal(t, string(tt.r.Status), got.Status)
			assert.Equal(t, tt.r.Notes, got.Notes)
			assert.Equal(t, tt.wantCourse, got.CourseName)
			assert.Equal(t, tt.wantStaffName, got.StaffName)
		})
	}
}

// ---- toLiffTrimmingCourseResponse ----

func TestToLiffTrimmingCourseResponse(t *testing.T) {
	price := int64(5000)

	tests := []struct {
		name string
		c    *model.TrimmingCourse
	}{
		{
			name: "with price",
			c: &model.TrimmingCourse{
				ID:          1,
				Name:        "フルコース",
				Description: "シャンプー+カット",
				Price:       &price,
				SortOrder:   1,
			},
		},
		{
			name: "nil price",
			c: &model.TrimmingCourse{
				ID:          2,
				Name:        "シャンプーのみ",
				Description: "",
				Price:       nil,
				SortOrder:   2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toLiffTrimmingCourseResponse(tt.c)
			assert.Equal(t, tt.c.ID, got.ID)
			assert.Equal(t, tt.c.Name, got.Name)
			assert.Equal(t, tt.c.Description, got.Description)
			assert.Equal(t, tt.c.SortOrder, got.SortOrder)
			if tt.c.Price == nil {
				assert.Nil(t, got.Price)
			} else {
				require.NotNil(t, got.Price)
				assert.Equal(t, *tt.c.Price, *got.Price)
			}
		})
	}
}

// ---- toLiffTrimmingOptionResponse ----

func TestToLiffTrimmingOptionResponse(t *testing.T) {
	price := int64(800)

	tests := []struct {
		name string
		o    *model.TrimmingOption
	}{
		{
			name: "with price",
			o: &model.TrimmingOption{
				ID:          1,
				Name:        "爪切り",
				Description: "爪切りオプション",
				Price:       &price,
				SortOrder:   1,
			},
		},
		{
			name: "nil price",
			o: &model.TrimmingOption{
				ID:          2,
				Name:        "耳掃除",
				Description: "",
				Price:       nil,
				SortOrder:   2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toLiffTrimmingOptionResponse(tt.o)
			assert.Equal(t, tt.o.ID, got.ID)
			assert.Equal(t, tt.o.Name, got.Name)
			assert.Equal(t, tt.o.Description, got.Description)
			assert.Equal(t, tt.o.SortOrder, got.SortOrder)
			if tt.o.Price == nil {
				assert.Nil(t, got.Price)
			} else {
				require.NotNil(t, got.Price)
				assert.Equal(t, *tt.o.Price, *got.Price)
			}
		})
	}
}
