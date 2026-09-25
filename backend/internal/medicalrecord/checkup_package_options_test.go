package medicalrecord

// checkup_package_options_test.go — manifest options → checkup_type_fields.options の
// 永続化形状を固定するユニットテスト（DB 不要）。
//
// 契約: options jsonb は [{"value":..,"label":..}] オブジェクト配列。
// - FE 動的フォーム DynamicCheckupFields が opt.value / opt.label を参照する
// - 結果値バリデーション checkupFieldOption が "value" を unmarshal する
// manifest 側は []string なので、importCheckupFields 側で正規化する必要がある。

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestMarshalCheckupFieldOptions_ValueLabelObjects(t *testing.T) {
	raw, err := marshalCheckupFieldOptions([]string{"0", "1", "5"})
	require.NoError(t, err)

	var rows []map[string]any
	require.NoError(t, json.Unmarshal(raw, &rows))
	require.Len(t, rows, 3)
	for i, want := range []string{"0", "1", "5"} {
		assert.Equal(t, want, rows[i]["value"], "option %d must expose value", i)
		assert.Equal(t, want, rows[i]["label"], "option %d must expose label", i)
	}
}

func TestMarshalCheckupFieldOptions_EmptyMarshalsAsArray(t *testing.T) {
	raw, err := marshalCheckupFieldOptions(nil)
	require.NoError(t, err)
	assert.JSONEq(t, `[]`, string(raw), "options must stay an array (parseCheckupOptionValues unmarshals []checkupFieldOption)")
}

func TestMarshalCheckupFieldOptions_RoundTripThroughResultValidation(t *testing.T) {
	raw, err := marshalCheckupFieldOptions([]string{"0", "1", "2", "3", "4", "5"})
	require.NoError(t, err)

	field := &model.CheckupTypeField{
		Name:      "アドプリットレベル",
		FieldType: model.CheckupFieldTypeSingleSelect,
		Options:   datatypes.JSON(raw),
	}
	allowed, err := parseCheckupOptionValues(field)
	require.NoError(t, err, "stored options must be readable by the production result validator")
	assert.Len(t, allowed, 6)

	assert.NoError(t, validateCheckupFieldValue(field, UpsertCheckupFieldResultInput{ValueText: "3"}))
	assert.Error(t, validateCheckupFieldValue(field, UpsertCheckupFieldResultInput{ValueText: "6"}))
}
