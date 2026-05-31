package handler

import "github.com/animal-ekarte/backend/internal/service"

type lineSendRequest struct {
	MessageType string  `json:"message_type" binding:"required"`
	Text        string  `json:"text"`
	FileID      *uint64 `json:"file_id"`
	FileName    string  `json:"file_name"`
	Purpose     string  `json:"purpose"`
}

func (r lineSendRequest) toServiceInput(ownerID, staffID uint64) *service.SendLineMessageInput {
	purpose := r.Purpose
	if purpose == "" {
		purpose = "other"
	}

	return &service.SendLineMessageInput{
		OwnerID:     ownerID,
		StaffID:     staffID,
		MessageType: normalizeMessageType(r.MessageType),
		Text:        r.Text,
		FileID:      r.FileID,
		FileName:    r.FileName,
		Purpose:     purpose,
	}
}
