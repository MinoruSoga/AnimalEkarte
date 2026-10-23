package medicalrecord

import (
	"bytes"
	"encoding/json"
	"errors"
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

// invalidJSONObjectBodyMessage は body が「ちょうど 1 つの JSON object」でないときの 400 文言。
const invalidJSONObjectBodyMessage = "リクエストボディは単一の JSON オブジェクトである必要があります"

// parseSingleJSONObject は body が「ちょうど 1 つの JSON object + EOF」であることを検証し、最上位キーを返す。
// 末尾に別の値やゴミが続く body（例: `{"a":1} {}`）、最上位が配列・スカラー・null の body は 400。
// 標準バインド（ShouldBindJSON）は最初の値しか読まないため、事前検査をここに一本化して fail-open を防ぐ。
func parseSingleJSONObject(body []byte) (map[string]json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	var keys map[string]json.RawMessage
	if err := dec.Decode(&keys); err != nil || keys == nil {
		return nil, apperrors.WrapInvalidInput(invalidJSONObjectBodyMessage)
	}
	var trailing json.RawMessage
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, apperrors.WrapInvalidInput(invalidJSONObjectBodyMessage)
	}
	return keys, nil
}

// rejectLegacyDiagnosisFields は最上位キーに旧単一診断フィールドが含まれていれば 400 を返す。
// 値が null でも拒否する（null は「クリア」の意図であり、無視すると診断が消えないまま成功に見える）。
func rejectLegacyDiagnosisFields(keys map[string]json.RawMessage) error {
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

// rejectStatusKeyOnUpdate は更新 body に status キーが存在すれば、値（null・文字列・数値・真偽・配列・object）に
// 関わらず固定エラー（422 medical_record_status_retired）を返す。存在判定はキーで行い、値の型や null には依存しない
// （`*string` へのバインドでは null が未送信に見え、数値等はバインド 400 になるため。R2-01）。
func rejectStatusKeyOnUpdate(keys map[string]json.RawMessage) error {
	if _, present := keys["status"]; present {
		return apperrors.WrapMedicalRecordStatusRetired(medicalRecordStatusFieldRetiredMessage)
	}
	return nil
}

// bindJSONGuarded は body を 1 度だけ読み、「単一 JSON object」検証と guard 群を ShouldBindJSON の前に実行する。
// 空 body は従来どおり標準バインドのエラー（400）に任せる。body は読み戻して標準バインドへ渡す。
func bindJSONGuarded[T any](c *gin.Context, dst *T, guards ...func(map[string]json.RawMessage) error) error {
	if c.Request == nil || c.Request.Body == nil {
		return c.ShouldBindJSON(dst)
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if len(bytes.TrimSpace(body)) == 0 {
		return c.ShouldBindJSON(dst)
	}
	keys, err := parseSingleJSONObject(body)
	if err != nil {
		return err
	}
	for _, guard := range guards {
		if err := guard(keys); err != nil {
			return err
		}
	}
	return c.ShouldBindJSON(dst)
}

// legacyAwareBindError は bindJSONGuarded の error を HTTP 応答用に正規化する。
// guard 由来の AppError（400 InvalidInput / 422 固定エラー）はそのまま、それ以外は標準のバインドエラー整形を通す。
func legacyAwareBindError(err error) error {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return err
	}
	return apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
}
