package lstep

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLineSendRequest_ToServiceInput(t *testing.T) {
	fileID := uint64(10)
	req := lineSendRequest{
		MessageType: "pdf",
		Text:        "text",
		FileID:      &fileID,
		FileName:    "file.pdf",
	}

	input := req.toServiceInput(1, 2)

	if input.OwnerID != 1 {
		t.Errorf("OwnerID = %d, want 1", input.OwnerID)
	}
	if input.StaffID != 2 {
		t.Errorf("StaffID = %d, want 2", input.StaffID)
	}
	if input.MessageType != "pdf_url" {
		t.Errorf("MessageType = %q, want pdf_url", input.MessageType)
	}
	if input.Purpose != "other" {
		t.Errorf("Purpose = %q, want other", input.Purpose)
	}
	if input.FileID == nil || *input.FileID != fileID {
		t.Errorf("FileID = %v, want %d", input.FileID, fileID)
	}
}

func TestLineSendRequest_ToServiceInput_ExplicitPurpose(t *testing.T) {
	req := lineSendRequest{MessageType: "text", Purpose: "follow_up"}

	input := req.toServiceInput(1, 2)

	if input.Purpose != req.Purpose {
		t.Errorf("Purpose = %q, want %q", input.Purpose, req.Purpose)
	}
}

func TestLinkAccountRequest_ToServiceInput(t *testing.T) {
	req := linkAccountRequest{
		LinkToken:   "token",
		LineIDToken: "line-token",
	}

	input := req.toServiceInput()

	if input.LinkToken != req.LinkToken {
		t.Errorf("LinkToken = %q, want %q", input.LinkToken, req.LinkToken)
	}
	if input.LineIDToken != req.LineIDToken {
		t.Errorf("LineIDToken = %q, want %q", input.LineIDToken, req.LineIDToken)
	}
}

func bindJSONRequest(t *testing.T, body any, dest any) error {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")
	return c.ShouldBindJSON(dest)
}

func TestLineSendRequest_BindingMax(t *testing.T) {
	tests := []struct {
		name    string
		body    map[string]any
		wantErr bool
	}{
		{
			name:    "text at LINE max 5000 is accepted",
			body:    map[string]any{"message_type": "text", "text": strings.Repeat("a", 5000)},
			wantErr: false,
		},
		{
			name:    "text over 5000 is rejected",
			body:    map[string]any{"message_type": "text", "text": strings.Repeat("a", 5001)},
			wantErr: true,
		},
		{
			name:    "file_name at 255 is accepted",
			body:    map[string]any{"message_type": "pdf", "file_name": strings.Repeat("a", 255)},
			wantErr: false,
		},
		{
			name:    "file_name over 255 is rejected",
			body:    map[string]any{"message_type": "pdf", "file_name": strings.Repeat("a", 256)},
			wantErr: true,
		},
		{
			name:    "purpose at 255 is accepted",
			body:    map[string]any{"message_type": "text", "purpose": strings.Repeat("a", 255)},
			wantErr: false,
		},
		{
			name:    "purpose over 255 is rejected",
			body:    map[string]any{"message_type": "text", "purpose": strings.Repeat("a", 256)},
			wantErr: true,
		},
		{
			name:    "purpose 101 is accepted (not probe max=100)",
			body:    map[string]any{"message_type": "text", "purpose": strings.Repeat("a", 101)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req lineSendRequest
			err := bindJSONRequest(t, tt.body, &req)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ShouldBindJSON error = nil, want over-max rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("ShouldBindJSON error = %v, want nil", err)
			}
		})
	}
}
