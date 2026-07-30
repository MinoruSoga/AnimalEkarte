package identitylink

// MemberRef is a clinic-scoped parent identity reference used in link requests.
// Both ClinicID and entity ID must resolve inside the actor's verified clinicIDs.
type MemberRef struct {
	ClinicID uint64 `json:"clinic_id"`
	ID       uint64 `json:"id"`
}

// OwnerMemberRef identifies an owners row.
type OwnerMemberRef struct {
	ClinicID uint64 `json:"clinic_id" binding:"required"`
	OwnerID  uint64 `json:"owner_id" binding:"required"`
}

// PetMemberRef identifies a pets row.
type PetMemberRef struct {
	ClinicID uint64 `json:"clinic_id" binding:"required"`
	PetID    uint64 `json:"pet_id" binding:"required"`
}

// ActorContext carries verified request-time identity for scope checks.
type ActorContext struct {
	StaffID         uint64
	HomeClinicID    uint64
	VerifiedClinics []uint64
	IPAddress       string
	UserAgent       string
}

// ClinicPetPair is a correlated (clinic_id, pet_id) used for linked history queries.
// Never expand into independent IN lists.
type ClinicPetPair struct {
	ClinicID uint64
	PetID    uint64
}

// LinkedTreatmentHistoryItem is a minimal treatment history row for the Phase 1 slice.
type LinkedTreatmentHistoryItem struct {
	ClinicID        uint64  `json:"clinic_id"`
	PetID           uint64  `json:"pet_id"`
	MedicalRecordID uint64  `json:"medical_record_id"`
	RecordNo        string  `json:"record_no"`
	RecordDate      string  `json:"record_date"`
	TreatmentID     uint64  `json:"treatment_id"`
	ItemType        string  `json:"item_type"`
	Content         string  `json:"content"`
	UnitPrice       int64   `json:"unit_price"`
	Quantity        float64 `json:"quantity"`
}
