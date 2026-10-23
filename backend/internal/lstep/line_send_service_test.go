package lstep

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// EMR-504: 旧 line_send_logs は DROP 済みで、置換先の配送記録 writer が未接続の間、Send は
// 入力にかかわらず固定の 422 (line_dispatch_pending) で fail-closed し、成功を返さない。
func TestSend_FailsClosedWithDispatchPending(t *testing.T) {
	fileID := uint64(9)
	inputs := map[string]*SendLineMessageInput{
		"text":      {OwnerID: 1, StaffID: 2, MessageType: "text", Text: "hello", Purpose: "other"},
		"pdf_url":   {OwnerID: 1, StaffID: 2, MessageType: "pdf_url", FileID: &fileID, FileName: "a.pdf"},
		"image_url": {OwnerID: 1, StaffID: 2, MessageType: "image_url", FileID: &fileID},
		"invalid":   {OwnerID: 1, StaffID: 2, MessageType: "unknown"},
	}
	svc := NewLineSendService()
	for name, input := range inputs {
		t.Run(name, func(t *testing.T) {
			result, err := svc.Send(context.Background(), 1, input)

			require.Error(t, err)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, apperrors.ErrUnprocessable)
			assert.True(t, apperrors.HasCode(err, apperrors.CodeLineDispatchPending))
		})
	}
}
