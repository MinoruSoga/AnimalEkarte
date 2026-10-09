package reservation

import (
	"encoding/json"
	"testing"
	"time"

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

// EMR-301/EMR-319 レビュー対応: reservation の nested pet summary は
// birth_date（誕生月バッジ）と deceased_at（死亡OR契約）を wire へ含める。
func TestToPetSummarySerializesBirthDateAndDeceasedAt(t *testing.T) {
	birthDate := time.Date(2019, time.October, 2, 0, 0, 0, 0, time.UTC)
	deceasedAt := time.Date(2026, time.July, 10, 12, 0, 0, 0, time.UTC)

	resp := toPetSummary(&model.Pet{
		ID:         20,
		Name:       "ポチ",
		Status:     model.PetStatusAlive,
		BirthDate:  &birthDate,
		DeceasedAt: &deceasedAt,
	})

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["birth_date"] != "2019-10-02" {
		t.Fatalf("birth_date = %v, want %q", payload["birth_date"], "2019-10-02")
	}
	if payload["deceased_at"] == nil {
		t.Fatal("deceased_at is missing; deceasedAt-only pets must propagate to the wire")
	}
}

func TestToReservationResponseIncludesPetDangerLevel(t *testing.T) {
	resp := toReservationResponse(&model.Reservation{
		ID:       1,
		ClinicID: 1,
		Pet: &model.Pet{
			ID:          20,
			Name:        "ポチ",
			DangerLevel: model.DangerLevelHigh,
		},
	})

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var payload struct {
		Pet struct {
			DangerLevel *string `json:"danger_level"`
		} `json:"pet"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if resp.Pet == nil {
		t.Fatal("pet summary is nil")
	}
	if got, want := resp.Pet.DangerLevel, string(model.DangerLevelHigh); got != want {
		t.Fatalf("response pet.danger_level = %q, want %q", got, want)
	}
	if payload.Pet.DangerLevel == nil {
		t.Fatalf("reservation JSON = %s, want pet.danger_level", body)
	}
	if got, want := *payload.Pet.DangerLevel, string(model.DangerLevelHigh); got != want {
		t.Fatalf("pet.danger_level = %q, want %q", got, want)
	}
}

// EMR-173: 受付カンバンの飼主「危険人物」マーク用に、owner サマリへ is_dangerous を
// 載せる。reservation.owner と reservation.pet.owner の両方が同じ toOwnerSummary を通る。
func TestToReservationResponseIncludesOwnerIsDangerous(t *testing.T) {
	owner := &model.Owner{
		ID:          10,
		Name:        "山田 太郎",
		IsDangerous: true,
	}
	resp := toReservationResponse(&model.Reservation{
		ID:       1,
		ClinicID: 1,
		Owner:    owner,
		Pet: &model.Pet{
			ID:    20,
			Name:  "ポチ",
			Owner: owner,
		},
	})

	body, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var payload struct {
		Owner struct {
			IsDangerous *bool `json:"is_dangerous"`
		} `json:"owner"`
		Pet struct {
			Owner struct {
				IsDangerous *bool `json:"is_dangerous"`
			} `json:"owner"`
		} `json:"pet"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if resp.Owner == nil || !resp.Owner.IsDangerous {
		t.Fatalf("response owner summary = %#v, want IsDangerous true", resp.Owner)
	}
	if resp.Pet == nil || resp.Pet.Owner == nil || !resp.Pet.Owner.IsDangerous {
		t.Fatalf("response pet.owner summary = %#v, want IsDangerous true", resp.Pet)
	}
	if payload.Owner.IsDangerous == nil || !*payload.Owner.IsDangerous {
		t.Fatalf("reservation JSON = %s, want owner.is_dangerous true", body)
	}
	if payload.Pet.Owner.IsDangerous == nil || !*payload.Pet.Owner.IsDangerous {
		t.Fatalf("reservation JSON = %s, want pet.owner.is_dangerous true", body)
	}
}

// 受付ヘッダー テレメトリ（change-ui.md Phase 2）: checked_in_at が
// レスポンス DTO に正しくマッピングされることを保証する。
func TestToReservationResponseIncludesCheckedInAt(t *testing.T) {
	checkedInAt := time.Date(2026, 7, 6, 9, 30, 0, 0, time.UTC)

	resp := toReservationResponse(&model.Reservation{
		ID:          1,
		ClinicID:    1,
		Status:      model.ReservationStatusCheckedIn,
		CheckedInAt: &checkedInAt,
	})

	if resp.CheckedInAt == nil {
		t.Fatal("checked_in_at is nil, want non-nil")
	}
	if !resp.CheckedInAt.Equal(checkedInAt) {
		t.Fatalf("checked_in_at = %v, want %v", resp.CheckedInAt, checkedInAt)
	}
}

// checked_in_at が未設定（受付前）の場合は nil のまま omitempty で省略されることを保証する。
func TestToReservationResponseOmitsCheckedInAtWhenNil(t *testing.T) {
	resp := toReservationResponse(&model.Reservation{
		ID:       1,
		ClinicID: 1,
		Status:   model.ReservationStatusPending,
	})

	if resp.CheckedInAt != nil {
		t.Fatalf("checked_in_at = %v, want nil", resp.CheckedInAt)
	}
}
