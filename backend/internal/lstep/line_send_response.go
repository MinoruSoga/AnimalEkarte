package lstep

import (
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type lineSendResponse struct {
	Sent     bool   `json:"sent"`
	SentAt   string `json:"sent_at"`
	TagAdded string `json:"tag_added,omitempty"`
}

type lineSendLogResponse struct {
	ID             uint64  `json:"id"`
	MessageType    string  `json:"message_type"`
	ContentSummary string  `json:"content_summary"`
	Status         string  `json:"status"`
	ErrorMessage   *string `json:"error_message,omitempty"`
	SentAt         string  `json:"sent_at"`
}

type lineSendLogListResponse struct {
	Items []lineSendLogResponse `json:"items"`
}

func toLineSendLogResponse(l *model.LineSendLog) lineSendLogResponse {
	return lineSendLogResponse{
		ID:             l.ID,
		MessageType:    l.MessageType,
		ContentSummary: l.ContentSummary,
		Status:         l.Status,
		ErrorMessage:   l.ErrorMessage,
		SentAt:         httpapi.LocalTimeRFC3339(l.SentAt),
	}
}

func toLineSendLogListResponse(logs []*model.LineSendLog) []lineSendLogResponse {
	out := make([]lineSendLogResponse, 0, len(logs))
	for _, l := range logs {
		out = append(out, toLineSendLogResponse(l))
	}
	return out
}
