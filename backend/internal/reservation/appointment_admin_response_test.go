package reservation

import (
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToReservationSummaryResponse(t *testing.T) {
	start := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC)

	tests := []struct {
		name             string
		reservation      *model.Reservation
		wantCustomerName string
		wantCourseName   string
		wantStaffName    string
	}{
		{
			name: "prefers LineCustomer RealName over DisplayName",
			reservation: &model.Reservation{
				ID:        1,
				StartTime: start,
				EndTime:   end,
				LineCustomer: &model.LineCustomer{
					DisplayName: "LINE表示名",
					RealName:    "本名太郎",
				},
				Source: model.ReservationSourceLine,
				Status: model.ReservationStatusPending,
			},
			wantCustomerName: "本名太郎",
		},
		{
			name: "falls back to LineCustomer DisplayName when RealName empty",
			reservation: &model.Reservation{
				ID:        2,
				StartTime: start,
				EndTime:   end,
				LineCustomer: &model.LineCustomer{
					DisplayName: "LINE表示名",
					RealName:    "",
				},
			},
			wantCustomerName: "LINE表示名",
		},
		{
			name: "uses Owner name when LineCustomer is nil",
			reservation: &model.Reservation{
				ID:        3,
				StartTime: start,
				EndTime:   end,
				Owner:     &model.Owner{Name: "飼主 花子"},
			},
			wantCustomerName: "飼主 花子",
		},
		{
			name: "customer name is empty when neither LineCustomer nor Owner set",
			reservation: &model.Reservation{
				ID:        4,
				StartTime: start,
				EndTime:   end,
			},
			wantCustomerName: "",
		},
		{
			name: "uses ReservationType ShortName when present",
			reservation: &model.Reservation{
				ID:              5,
				StartTime:       start,
				EndTime:         end,
				ReservationType: &model.ReservationType{Name: "一般診療", ShortName: "一般"},
			},
			wantCourseName: "一般",
		},
		{
			name: "falls back to ReservationType Name when ShortName empty",
			reservation: &model.Reservation{
				ID:              6,
				StartTime:       start,
				EndTime:         end,
				ReservationType: &model.ReservationType{Name: "一般診療", ShortName: ""},
			},
			wantCourseName: "一般診療",
		},
		{
			name: "course name empty when ReservationType nil",
			reservation: &model.Reservation{
				ID:        7,
				StartTime: start,
				EndTime:   end,
			},
			wantCourseName: "",
		},
		{
			name: "uses Doctor name when present",
			reservation: &model.Reservation{
				ID:        8,
				StartTime: start,
				EndTime:   end,
				Doctor:    &model.Staff{Name: "佐藤 医師"},
			},
			wantStaffName: "佐藤 医師",
		},
		{
			name: "staff name empty when Doctor nil",
			reservation: &model.Reservation{
				ID:        9,
				StartTime: start,
				EndTime:   end,
			},
			wantStaffName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toReservationSummaryResponse(tt.reservation)

			if resp.ID != tt.reservation.ID {
				t.Errorf("ID = %d, want %d", resp.ID, tt.reservation.ID)
			}
			if !resp.StartTime.Equal(start) {
				t.Errorf("StartTime = %v, want %v", resp.StartTime, start)
			}
			if !resp.EndTime.Equal(end) {
				t.Errorf("EndTime = %v, want %v", resp.EndTime, end)
			}
			if resp.CustomerName != tt.wantCustomerName {
				t.Errorf("CustomerName = %q, want %q", resp.CustomerName, tt.wantCustomerName)
			}
			if resp.CourseShortName != tt.wantCourseName {
				t.Errorf("CourseShortName = %q, want %q", resp.CourseShortName, tt.wantCourseName)
			}
			if resp.StaffName != tt.wantStaffName {
				t.Errorf("StaffName = %q, want %q", resp.StaffName, tt.wantStaffName)
			}
			if resp.Source != string(tt.reservation.Source) {
				t.Errorf("Source = %q, want %q", resp.Source, string(tt.reservation.Source))
			}
			if resp.Status != string(tt.reservation.Status) {
				t.Errorf("Status = %q, want %q", resp.Status, string(tt.reservation.Status))
			}
			if resp.IsStaffDelegated != tt.reservation.IsStaffDelegated {
				t.Errorf("IsStaffDelegated = %v, want %v", resp.IsStaffDelegated, tt.reservation.IsStaffDelegated)
			}
		})
	}
}

func TestToReservationDetailResponse(t *testing.T) {
	start := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC)
	ownerID := uint64(10)
	petID := uint64(20)
	doctorID := uint64(30)
	createdBy := uint64(40)

	t.Run("populates nested relation-derived fields when all relations present", func(t *testing.T) {
		reservation := &model.Reservation{
			ID:                1,
			StartTime:         start,
			EndTime:           end,
			OwnerID:           &ownerID,
			PetID:             &petID,
			VisitType:         model.VisitTypeFirst,
			ReservationTypeID: 7,
			DoctorID:          &doctorID,
			IsDesignated:      true,
			IsStaffDelegated:  true,
			Notes:             "備考",
			CreatedBy:         &createdBy,
			LineCustomer:      &model.LineCustomer{DisplayName: "LINE名", RealName: "本名"},
			ReservationType:   &model.ReservationType{Name: "一般診療", ShortName: "一般"},
			Doctor:            &model.Staff{Name: "佐藤 医師"},
			CreatedByStaff:    &model.Staff{Name: "受付 太郎"},
		}

		resp := toReservationDetailResponse(reservation)

		if resp.ID != 1 {
			t.Errorf("ID = %d, want 1", resp.ID)
		}
		if resp.OwnerID == nil || *resp.OwnerID != ownerID {
			t.Errorf("OwnerID = %v, want %d", resp.OwnerID, ownerID)
		}
		if resp.PetID == nil || *resp.PetID != petID {
			t.Errorf("PetID = %v, want %d", resp.PetID, petID)
		}
		if resp.DoctorID == nil || *resp.DoctorID != doctorID {
			t.Errorf("DoctorID = %v, want %d", resp.DoctorID, doctorID)
		}
		if resp.CreatedBy == nil || *resp.CreatedBy != createdBy {
			t.Errorf("CreatedBy = %v, want %d", resp.CreatedBy, createdBy)
		}
		if resp.CustomerName != "本名" {
			t.Errorf("CustomerName = %q, want %q", resp.CustomerName, "本名")
		}
		if resp.CourseShortName != "一般" {
			t.Errorf("CourseShortName = %q, want %q", resp.CourseShortName, "一般")
		}
		if resp.StaffName != "佐藤 医師" {
			t.Errorf("StaffName = %q, want %q", resp.StaffName, "佐藤 医師")
		}
		if resp.CreatedByName != "受付 太郎" {
			t.Errorf("CreatedByName = %q, want %q", resp.CreatedByName, "受付 太郎")
		}
		if resp.VisitType != string(model.VisitTypeFirst) {
			t.Errorf("VisitType = %q, want %q", resp.VisitType, string(model.VisitTypeFirst))
		}
		if !resp.IsDesignated {
			t.Error("IsDesignated = false, want true")
		}
		if !resp.IsStaffDelegated {
			t.Error("IsStaffDelegated = false, want true")
		}
		if resp.Notes != "備考" {
			t.Errorf("Notes = %q, want %q", resp.Notes, "備考")
		}
	})

	t.Run("falls back to empty/zero values when relations are nil", func(t *testing.T) {
		reservation := &model.Reservation{
			ID:                2,
			StartTime:         start,
			EndTime:           end,
			ReservationTypeID: 8,
		}

		resp := toReservationDetailResponse(reservation)

		if resp.OwnerID != nil {
			t.Errorf("OwnerID = %v, want nil", resp.OwnerID)
		}
		if resp.PetID != nil {
			t.Errorf("PetID = %v, want nil", resp.PetID)
		}
		if resp.DoctorID != nil {
			t.Errorf("DoctorID = %v, want nil", resp.DoctorID)
		}
		if resp.CreatedBy != nil {
			t.Errorf("CreatedBy = %v, want nil", resp.CreatedBy)
		}
		if resp.CustomerName != "" {
			t.Errorf("CustomerName = %q, want empty", resp.CustomerName)
		}
		if resp.CourseShortName != "" {
			t.Errorf("CourseShortName = %q, want empty", resp.CourseShortName)
		}
		if resp.StaffName != "" {
			t.Errorf("StaffName = %q, want empty", resp.StaffName)
		}
		if resp.CreatedByName != "" {
			t.Errorf("CreatedByName = %q, want empty", resp.CreatedByName)
		}
	})

	t.Run("prefers Owner over LineCustomer priority and ShortName fallback to Name", func(t *testing.T) {
		reservation := &model.Reservation{
			ID:                3,
			StartTime:         start,
			EndTime:           end,
			ReservationTypeID: 9,
			Owner:             &model.Owner{Name: "飼主 花子"},
			ReservationType:   &model.ReservationType{Name: "一般診療", ShortName: ""},
		}

		resp := toReservationDetailResponse(reservation)

		if resp.CustomerName != "飼主 花子" {
			t.Errorf("CustomerName = %q, want %q", resp.CustomerName, "飼主 花子")
		}
		if resp.CourseShortName != "一般診療" {
			t.Errorf("CourseShortName = %q, want %q", resp.CourseShortName, "一般診療")
		}
	})
}
