package handler

import (
	"testing"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestToReservationResponseIncludesNestedReservationType(t *testing.T) {
	groupID := uint64(2)
	ownerID := uint64(10)
	petID := uint64(20)
	doctorID := uint64(30)

	resp := toReservationResponse(&model.Reservation{
		ID:                1,
		ClinicID:          1,
		OwnerID:           &ownerID,
		PetID:             &petID,
		ReservationTypeID: 5,
		DoctorID:          &doctorID,
		Owner: &model.Owner{
			ID:   ownerID,
			Name: "山田 太郎",
		},
		Pet: &model.Pet{
			ID:   petID,
			Name: "ポチ",
		},
		ReservationType: &model.ReservationType{
			ID:       5,
			ClinicID: 1,
			Name:     "一般診療",
			GroupID:  &groupID,
			Group: &model.ReservationTypeGroup{
				ID:    groupID,
				Name:  "診療",
				Color: "#2563EB",
			},
		},
		Doctor: &model.Staff{
			ID:   doctorID,
			Name: "佐藤 医師",
		},
	})

	if resp.ReservationType == nil {
		t.Fatal("reservation_type is nil")
	}
	if resp.ReservationType.Name != "一般診療" {
		t.Fatalf("reservation_type.name = %q, want %q", resp.ReservationType.Name, "一般診療")
	}
	if resp.ReservationType.Group == nil {
		t.Fatal("reservation_type.group is nil")
	}
	if resp.ReservationType.Group.Name != "診療" {
		t.Fatalf("reservation_type.group.name = %q, want %q", resp.ReservationType.Group.Name, "診療")
	}
	if resp.Owner == nil || resp.Owner.OwnerName != "山田 太郎" {
		t.Fatalf("owner summary = %#v, want owner_name %q", resp.Owner, "山田 太郎")
	}
	if resp.Pet == nil || resp.Pet.Name != "ポチ" {
		t.Fatalf("pet summary = %#v, want name %q", resp.Pet, "ポチ")
	}
	if resp.Doctor == nil || resp.Doctor.Name != "佐藤 医師" {
		t.Fatalf("doctor summary = %#v, want name %q", resp.Doctor, "佐藤 医師")
	}
}
