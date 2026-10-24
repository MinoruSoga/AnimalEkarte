package lstep

type lineSendResponse struct {
	Sent     bool   `json:"sent"`
	SentAt   string `json:"sent_at"`
	TagAdded string `json:"tag_added,omitempty"`
}
