package lstep

// linkAccountRequest は LinkLiffAccount のリクエスト。
type linkAccountRequest struct {
	LinkToken   string `json:"link_token"`
	LineIDToken string `json:"line_id_token"`
}

// toServiceInput は同一 package 化により型変換で足りる（field 名・型は 1:1・
// 対応は line_request_test.go の TestLinkAccountRequest_ToServiceInput が担保）。
func (r linkAccountRequest) toServiceInput() LinkAccountInput {
	return LinkAccountInput(r)
}
