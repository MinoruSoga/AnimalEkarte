package medicalrecord

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// legacyDiagnosisFieldNames は EMR-495 A で廃止した単一診断フィールド（clinical_plans の診断 4 列と、
// カルテ作成 body の diagnosis_1_* / diagnosis_2_* 系）。FE が旧名のまま送ると、Gin の既定バインドは
// 未知キーを黙って捨てて 200/201 を返し、診断が保存されないまま成功に見える（SPEC-01）。
// 既存の binder に DisallowUnknownFields は使わず、この旧名だけを狙って拒否する。
var legacyDiagnosisFieldNames = map[string]struct{}{
	"diagnosis_category_id":   {},
	"diagnosis_type_id":       {},
	"diagnosis_name_id":       {},
	"diagnosis_1_category_id": {},
	"diagnosis_1_type_id":     {},
	"diagnosis_1_name_id":     {},
	"diagnosis1_category_id":  {},
	"diagnosis1_type_id":      {},
	"diagnosis1_name_id":      {},
	"diagnosis_2_type_id":     {},
	"diagnosis_2_name_id":     {},
	"diagnosis2_type_id":      {},
	"diagnosis2_name_id":      {},
}

// legacyDiagnosisFieldsMessage は旧フィールドを送った request への 400 文言。
const legacyDiagnosisFieldsMessage = "旧来の単一診断フィールドは廃止されました。診断は diagnoses[]（diagnosis_slot 必須）で送信してください"

// rejectLegacyDiagnosisFields は JSON object の最上位キーに旧単一診断フィールドが含まれていれば 400 を返す。
// 値が null でも拒否する（null は「クリア」の意図であり、無視すると診断が消えないまま成功に見える）。
// JSON として解釈できない body はここでは扱わず、後続の標準バインドに任せる。
func rejectLegacyDiagnosisFields(body []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(body, &keys); err != nil {
		return nil
	}
	found := make([]string, 0)
	for key := range keys {
		if _, legacy := legacyDiagnosisFieldNames[key]; legacy {
			found = append(found, key)
		}
	}
	if len(found) == 0 {
		return nil
	}
	sort.Strings(found)
	return apperrors.WrapInvalidInput(legacyDiagnosisFieldsMessage + "（" + strings.Join(found, ", ") + "）")
}

// bindJSONRejectingLegacyDiagnosis は c.ShouldBindJSON の前に旧単一診断フィールドを拒否する。
// body は読み戻して標準バインドへ渡す。
func bindJSONRejectingLegacyDiagnosis(c *gin.Context, dst any) error {
	if c.Request == nil || c.Request.Body == nil {
		return c.ShouldBindJSON(dst)
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err := rejectLegacyDiagnosisFields(body); err != nil {
		return err
	}
	return c.ShouldBindJSON(dst)
}

// legacyAwareBindError は bindJSONRejectingLegacyDiagnosis の error を HTTP 応答用に正規化する。
// 旧フィールド拒否は既に InvalidInput（固定文言）なのでそのまま、それ以外は標準のバインドエラー整形を通す。
func legacyAwareBindError(err error) error {
	if apperrors.IsInvalidInput(err) {
		return err
	}
	return apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
}
