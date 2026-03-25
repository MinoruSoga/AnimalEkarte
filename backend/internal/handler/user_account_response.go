package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// userResponse はユーザーアカウントの API レスポンス
// password_hash は絶対に含めない
type userResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	UserType    string    `json:"user_type"`
	StaffID     *string   `json:"staff_id,omitempty"`
	AvatarURL   string    `json:"avatar_url"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// userMembershipResponse は所属クリニック情報の API レスポンス
type userMembershipResponse struct {
	ClinicID string `json:"clinic_id"`
	IsMain   bool   `json:"is_main"`
}

// userDetailResponse はユーザー詳細（所属クリニック含む）の API レスポンス
type userDetailResponse struct {
	userResponse
	Memberships []userMembershipResponse `json:"memberships"`
}

// userPermissionResponse は権限の API レスポンス
type userPermissionResponse struct {
	Permission string `json:"permission"`
}

// toUserResponse は model.UserAccount を userResponse に変換する
func toUserResponse(a *model.UserAccount) userResponse {
	r := userResponse{
		ID:          strconv.FormatUint(a.ID, 10),
		Email:       a.Email,
		DisplayName: a.DisplayName,
		UserType:    string(a.UserType),
		AvatarURL:   a.AvatarURL,
		Status:      string(a.Status),
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
	if a.StaffID != nil {
		s := strconv.FormatUint(*a.StaffID, 10)
		r.StaffID = &s
	}
	return r
}
