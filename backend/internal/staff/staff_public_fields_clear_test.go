package staff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// EMR-448: スタッフ公開4項目（免許番号・LINE表示名・説明文・画像URL）の "" は空で保存する。
// LINE表示名が空のときの氏名 fallback は表示側（reservation/appointment_notification_service.go・liff_response.go）。
func TestBuildStaffUpdate_EmptyPublicFieldsAreWritten(t *testing.T) {
	empty := ""
	fields := buildStaffUpdate(&UpdateStaffInput{
		LicenseNumber:          &empty,
		ReservationDisplayName: &empty,
		ReservationComment:     &empty,
		ReservationImageURL:    &empty,
	})
	assert.Equal(t, map[string]any{
		"license_number":           "",
		"reservation_display_name": "",
		"reservation_comment":      "",
		"reservation_image_url":    "",
	}, fields)
	assert.Empty(t, buildStaffUpdate(&UpdateStaffInput{}))
}
