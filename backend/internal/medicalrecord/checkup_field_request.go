package medicalrecord

// upsertCheckupFieldResultRequest は健診結果値 1 件分のバインド struct。
// status / is_abnormal はサーバ側で導出するため受け付けない（信頼境界はサーバ）。
type upsertCheckupFieldResultRequest struct {
	CheckupTypeFieldID *uint64  `json:"checkup_type_field_id" binding:"required"`
	ValueNumber        *float64 `json:"value_number"`
	ValueText          string   `json:"value_text"`
	ValueBool          *bool    `json:"value_bool"`
	ValueList          []string `json:"value_list"`
}

// replaceCheckupFieldResultsRequest は健診結果値の一括置換 PUT リクエスト。
// #211: results の省略（JSON に results キーが無い／null）は nil として service へ伝播し拒否される。
// 患者検診結果値の全削除は意図的な空配列 [] の明示送信を要求し、空ボディ/壊れた request による
// 偶発的・silent な全消去を防止する。max は 1 パッケージあたりのフィールド数の現実的上限。
type replaceCheckupFieldResultsRequest struct {
	Results []upsertCheckupFieldResultRequest `json:"results" binding:"max=200,dive"`
}

func (r replaceCheckupFieldResultsRequest) toServiceInput() []UpsertCheckupFieldResultInput {
	// nil（results 省略）は nil のまま伝播させ、service 層で必須エラーにする（#211 偶発全消去防御）。
	// 明示的な空配列 [] は非 nil の空 slice として伝播し、意図的な全削除（監査される）となる。
	if r.Results == nil {
		return nil
	}
	inputs := make([]UpsertCheckupFieldResultInput, 0, len(r.Results))
	for _, item := range r.Results {
		inputs = append(inputs, UpsertCheckupFieldResultInput(item))
	}
	return inputs
}
