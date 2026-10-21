package reservation

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// liffSettingsResponse はLIFF向け公開設定レスポンス（機密フィールド除外）。
type liffSettingsResponse struct {
	Status               string `json:"status"`
	HeaderText           string `json:"header_text"`
	ReservationNotice    string `json:"reservation_notice"`
	CancelNotice         string `json:"cancel_notice"`
	PrivacyPolicy        string `json:"privacy_policy"`
	PhoneNumber          string `json:"phone_number"`
	RequestExample       string `json:"request_example"`
	LiffID               string `json:"liff_id"`
	ShowNoStaffOption    bool   `json:"show_no_staff_option"`
	BookingWindowMaxDays int    `json:"booking_window"`
}

func toLiffSettingsResponse(s *model.LineReservationSetting) liffSettingsResponse {
	return liffSettingsResponse{
		Status:               s.Status,
		HeaderText:           s.HeaderText,
		ReservationNotice:    s.ReservationNotice,
		CancelNotice:         s.CancelNotice,
		PrivacyPolicy:        s.PrivacyPolicy,
		PhoneNumber:          s.PhoneNumber,
		RequestExample:       s.RequestExample,
		LiffID:               s.LiffID,
		ShowNoStaffOption:    s.ShowNoStaffOption,
		BookingWindowMaxDays: s.BookingWindowMaxDays,
	}
}

// liffCourseResponse はLIFF向けコースレスポンス。
type liffCourseResponse struct {
	ID                  uint64 `json:"id"`
	Name                string `json:"name"`
	ShortName           string `json:"short_name"`
	ShowShortName       bool   `json:"show_short_name"`
	DurationMinutes     int    `json:"duration_minutes"`
	ReservationComment  string `json:"reservation_comment"`
	ReservationImageURL string `json:"reservation_image_url"`
	SortOrder           int    `json:"sort_order"`
	// Category (P1-6, PR #186 review): 予約フローが course.category === 'trimming' で
	// トリミング詳細ステップの表示を分岐する。欠落時は FE 側が 'general' にフォールバックし
	// トリミング予約でも詳細入力がスキップされていた。
	Category string `json:"category"`
}

func toLiffCourseResponse(st *model.ReservationType) liffCourseResponse {
	// 名前のフォールバックチェーン:
	// 1. ReservationDisplayName（LINE表示名）
	// 2. ShortName（ShowShortName が true の場合）
	// 3. Name（元の名称）
	name := st.ReservationDisplayName
	if name == "" {
		if st.ShowShortName && st.ShortName != "" {
			name = st.ShortName
		} else {
			name = st.Name
		}
	}
	return liffCourseResponse{
		ID:                  st.ID,
		Name:                name,
		ShortName:           st.ShortName,
		ShowShortName:       st.ShowShortName,
		DurationMinutes:     st.DurationMinutes,
		ReservationComment:  st.ReservationComment,
		ReservationImageURL: st.ReservationImageURL,
		SortOrder:           st.SortOrder,
		Category:            string(st.Category),
	}
}

// liffStaffResponse はLIFF向けスタッフレスポンス。
type liffStaffResponse struct {
	ID                  uint64 `json:"id"`
	Name                string `json:"name"`
	ReservationComment  string `json:"reservation_comment"`
	ReservationImageURL string `json:"reservation_image_url"`
	SortOrder           int    `json:"sort_order"`
}

func toLiffStaffResponse(st *model.Staff) liffStaffResponse {
	name := st.ReservationDisplayName
	if name == "" {
		name = st.Name
	}
	return liffStaffResponse{
		ID:                  st.ID,
		Name:                name,
		ReservationComment:  st.ReservationComment,
		ReservationImageURL: st.ReservationImageURL,
		SortOrder:           st.SortOrder,
	}
}

// liffAvailableDateResponse は空き日付レスポンス。
type liffAvailableDateResponse struct {
	Date      string `json:"date"`
	Weekday   int    `json:"weekday"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// liffAvailableDatesResponse は空き日付一覧レスポンス。
type liffAvailableDatesResponse struct {
	Dates  []liffAvailableDateResponse `json:"dates"`
	Window BookingWindow               `json:"window"`
}

// liffTimeSlotResponse は時間枠レスポンス。
type liffTimeSlotResponse struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// liffReservationResponse は予約レスポンス。
// EMR-519: ID は正の int64 範囲の十進文字列で返す。pet_id / pet_name は常にキーを出し、
// 両方 nonnull か両方 null（値の決定は liffReservationPet に集約）。
type liffReservationResponse struct {
	ID         string    `json:"id"`
	Date       string    `json:"date"`
	StartTime  string    `json:"start_time"` // "HH:MM"
	EndTime    string    `json:"end_time"`   // "HH:MM"
	CourseName string    `json:"course_name"`
	PetID      *string   `json:"pet_id"`
	PetName    *string   `json:"pet_name"`
	StaffName  string    `json:"staff_name,omitempty"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type liffProfileOwnerPetResponse struct {
	ID            uint64 `json:"id"`
	Name          string `json:"name"`
	AnimalSpecies *struct {
		Name string `json:"name"`
	} `json:"animal_species,omitempty"`
}

type liffProfileOwnerResponse struct {
	OwnerName string                        `json:"owner_name"`
	Phone     string                        `json:"phone"`
	Pets      []liffProfileOwnerPetResponse `json:"pets,omitempty"`
}

type liffProfileResponse struct {
	LineUserID       string                    `json:"line_user_id"`
	DisplayName      string                    `json:"display_name"`
	AdditionalFields json.RawMessage           `json:"additional_fields"`
	Owner            *liffProfileOwnerResponse `json:"owner,omitempty"`
}

func toLiffProfileResponse(c *model.LineCustomer) liffProfileResponse {
	resp := liffProfileResponse{
		LineUserID:       c.LineUserID,
		DisplayName:      c.DisplayName,
		AdditionalFields: json.RawMessage(c.AdditionalFields),
	}
	if c.Owner != nil {
		// EMR-290: プロフィールの pets は予約画面の選択肢として使われるため生存ペットに
		// 限定する（livingPets と同じ OR 契約）。死亡ペットを選ぶと resolveReservationPetID が
		// attach を拒否し、pet_id なしの予約が作られる。
		living := livingPets(c.Owner.Pets)
		pets := make([]liffProfileOwnerPetResponse, 0, len(living))
		for i := range living {
			pet := liffProfileOwnerPetResponse{
				ID:   living[i].ID,
				Name: living[i].Name,
			}
			if living[i].AnimalSpecies != nil {
				pet.AnimalSpecies = &struct {
					Name string `json:"name"`
				}{Name: living[i].AnimalSpecies.Name}
			}
			pets = append(pets, pet)
		}
		resp.Owner = &liffProfileOwnerResponse{
			OwnerName: c.Owner.Name,
			Phone:     c.Owner.Phone,
			Pets:      pets,
		}
	}
	return resp
}

func toLiffReservationResponse(r *model.Reservation) liffReservationResponse {
	petID, petName := liffReservationPet(r)
	res := liffReservationResponse{
		ID:        strconv.FormatUint(r.ID, 10),
		Date:      r.StartTime.In(time.Local).Format(time.DateOnly),
		StartTime: r.StartTime.In(time.Local).Format("15:04"),
		EndTime:   r.EndTime.In(time.Local).Format("15:04"),
		PetID:     petID,
		PetName:   petName,
		Status:    string(r.Status),
		Notes:     r.Notes,
		CreatedAt: r.CreatedAt.In(time.Local),
	}
	if r.ReservationType != nil {
		res.CourseName = liffCourseName(r.ReservationType)
	}
	if r.Doctor != nil {
		res.StaffName = r.Doctor.ReservationDisplayName
		if res.StaffName == "" {
			res.StaffName = r.Doctor.Name
		}
	}
	return res
}

func liffCourseName(rt *model.ReservationType) string {
	if rt.ReservationDisplayName != "" {
		return rt.ReservationDisplayName
	}
	if rt.ShowShortName && rt.ShortName != "" {
		return rt.ShortName
	}
	return rt.Name
}

// liffReservationPet は一覧応答の pet_id / pet_name を返す。
// 「両方 nonnull か両方 null」の不変条件はこの関数の戻り値だけで保証し、
// 片方だけ値がある形は返さない。Stage 1 では常に (nil, nil) を返す。
// どのペット情報・関係を読むかの reader 接続は Stage 2 で行う。
func liffReservationPet(r *model.Reservation) (petID *string, petName *string) {
	return nil, nil
}

// liffReservationCreatedResponse はLIFF予約作成（201）専用レスポンス。
// 一覧用の liffReservationResponse とは別型（id/notes の2フィールドのみ露出）。
type liffReservationCreatedResponse struct {
	ID    string `json:"id"`
	Notes string `json:"notes"`
}

func toLiffReservationCreatedResponse(r *model.Reservation) liffReservationCreatedResponse {
	return liffReservationCreatedResponse{
		ID:    strconv.FormatUint(r.ID, 10),
		Notes: r.Notes,
	}
}

// liffHealthCardVaccineResponse はLIFF向け健康手帳ワクチン接種履歴レスポンス。
// 日付は last_visit_date と同じくローカル日付 (YYYY-MM-DD) のみを露出する。
type liffHealthCardVaccineResponse struct {
	VaccineName  string  `json:"vaccine_name"`
	VaccinatedAt string  `json:"vaccinated_at"`
	NextDueAt    *string `json:"next_due_at"`
}

// liffHealthCardPetResponse はLIFF向け健康手帳ペットレスポンス。
type liffHealthCardPetResponse struct {
	PetID         string                          `json:"pet_id"`
	PetName       string                          `json:"pet_name"`
	Species       string                          `json:"species"`
	Breed         string                          `json:"breed"`
	Vaccines      []liffHealthCardVaccineResponse `json:"vaccines"`
	LastVisitDate *string                         `json:"last_visit_date"`
}

// liffHealthCardResponse はLIFF向け健康手帳レスポンス。
type liffHealthCardResponse struct {
	OwnerName string                      `json:"owner_name"`
	Pets      []liffHealthCardPetResponse `json:"pets"`
}

// liffLocalDateString はローカル日付 (YYYY-MM-DD) を返す。ゼロ値は空文字にし、
// 0001-01-01 のような意味を持たない日付を露出しない。
func liffLocalDateString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(time.Local).Format(time.DateOnly)
}

// liffLocalDatePtr はローカル日付 (YYYY-MM-DD) を返す。nil/ゼロ値は nil を返し、
// JSON では null として出力される（FE の em-dash プレースホルダーに対応）。
func liffLocalDatePtr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.In(time.Local).Format(time.DateOnly)
	return &s
}

func toLiffHealthCardResponse(r *HealthCardResult) liffHealthCardResponse {
	pets := make([]liffHealthCardPetResponse, 0, len(r.Pets))
	for _, p := range r.Pets {
		vaccines := make([]liffHealthCardVaccineResponse, 0, len(p.Vaccines))
		for _, v := range p.Vaccines {
			vaccines = append(vaccines, liffHealthCardVaccineResponse{
				VaccineName:  v.VaccineName,
				VaccinatedAt: liffLocalDateString(v.VaccinatedAt),
				NextDueAt:    liffLocalDatePtr(v.NextDueAt),
			})
		}
		var lastVisitDate *string
		if p.LastVisitDate != nil {
			s := p.LastVisitDate.In(time.Local).Format(time.DateOnly)
			lastVisitDate = &s
		}
		pets = append(pets, liffHealthCardPetResponse{
			PetID:         strconv.FormatUint(p.PetID, 10),
			PetName:       p.PetName,
			Species:       p.Species,
			Breed:         p.Breed,
			Vaccines:      vaccines,
			LastVisitDate: lastVisitDate,
		})
	}
	return liffHealthCardResponse{OwnerName: r.OwnerName, Pets: pets}
}

// liffTrimmingCourseResponse はLIFF向けトリミングコースレスポンス。
type liffTrimmingCourseResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       *int64 `json:"price"`
	SortOrder   int    `json:"sort_order"`
}

func toLiffTrimmingCourseResponse(c *model.TrimmingCourse) liffTrimmingCourseResponse {
	return liffTrimmingCourseResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Price:       c.Price,
		SortOrder:   c.SortOrder,
	}
}

// liffTrimmingOptionResponse はLIFF向けトリミングオプションレスポンス。
type liffTrimmingOptionResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       *int64 `json:"price"`
	SortOrder   int    `json:"sort_order"`
}

func toLiffTrimmingOptionResponse(o *model.TrimmingOption) liffTrimmingOptionResponse {
	return liffTrimmingOptionResponse{
		ID:          o.ID,
		Name:        o.Name,
		Description: o.Description,
		Price:       o.Price,
		SortOrder:   o.SortOrder,
	}
}
