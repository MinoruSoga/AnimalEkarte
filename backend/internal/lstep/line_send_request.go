package lstep

// normalizeMessageType は FE から送られる message_type エイリアスをサービス内部値に変換する（ISSUE-002）。
func normalizeMessageType(t string) string {
	switch t {
	case "pdf":
		return "pdf_url"
	case "image":
		return "image_url"
	default:
		return t
	}
}

type lineSendRequest struct {
	MessageType string  `json:"message_type" binding:"required"`
	Text        string  `json:"text" binding:"omitempty,max=5000"`
	FileID      *uint64 `json:"file_id"`
	FileName    string  `json:"file_name" binding:"omitempty,max=255"`
	Purpose     string  `json:"purpose" binding:"omitempty,max=255"`
}

func (r lineSendRequest) toServiceInput(ownerID, staffID uint64) *SendLineMessageInput {
	purpose := r.Purpose
	if purpose == "" {
		purpose = "other"
	}

	return &SendLineMessageInput{
		OwnerID:     ownerID,
		StaffID:     staffID,
		MessageType: normalizeMessageType(r.MessageType),
		Text:        r.Text,
		FileID:      r.FileID,
		FileName:    r.FileName,
		Purpose:     purpose,
	}
}
