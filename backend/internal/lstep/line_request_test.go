package lstep

import "testing"

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
