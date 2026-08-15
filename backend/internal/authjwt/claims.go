package authjwt

import "github.com/golang-jwt/jwt/v5"

// Claims は JWT のペイロード（access / refresh 共通）。
type Claims struct {
	UserID          string   `json:"user_id"`
	ClinicID        string   `json:"clinic_id"`
	IsSystemAdmin   bool     `json:"is_system_admin"`
	ClinicIDs       []uint64 `json:"clinic_ids,omitempty"`
	AccountEpoch    int64    `json:"account_epoch"`
	RefreshFamilyID string   `json:"refresh_family_id,omitempty"`
	jwt.RegisteredClaims
}
