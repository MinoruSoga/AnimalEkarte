package handler

import (
	"github.com/animal-ekarte/backend/internal/service"
)

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
// results が nil の場合は空配列として扱う（全削除と等価）。max は 1 パッケージあたりの
// フィールド数の現実的上限（過大配列による無駄な処理を防ぐ）。
type replaceCheckupFieldResultsRequest struct {
	Results []upsertCheckupFieldResultRequest `json:"results" binding:"max=200,dive"`
}

func (r replaceCheckupFieldResultsRequest) toServiceInput() []service.UpsertCheckupFieldResultInput {
	inputs := make([]service.UpsertCheckupFieldResultInput, 0, len(r.Results))
	for _, item := range r.Results {
		inputs = append(inputs, service.UpsertCheckupFieldResultInput{
			CheckupTypeFieldID: item.CheckupTypeFieldID,
			ValueNumber:        item.ValueNumber,
			ValueText:          item.ValueText,
			ValueBool:          item.ValueBool,
			ValueList:          item.ValueList,
		})
	}
	return inputs
}
