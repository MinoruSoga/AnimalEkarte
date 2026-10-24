package owner

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-491: 氏名・住所・連絡先などの飼主の基本情報は owners の列ではなく、共通 basic revision の項目になった。
// その改訂 writer と新規登録 writer が接続されるまで、基本情報を伴う書き込みは成功を装わず固定の 422 で拒否する。
// 拒否は書き込み・採番・外部通知より前に行う。基本情報を含まない更新（備考・割引・会員区分など）は従来の経路を通す。

const (
	ownerBasicRevisionPendingMessage = "飼主の基本情報（氏名・住所・連絡先など）の保存は、共通基本情報の改訂機能の準備中のため現在できません"
	ownerRegistrationPendingMessage  = "飼主とペットの新規登録は、共通基本情報の登録機能の準備中のため現在できません"
)

// ownerBasicRequestKeys は PATCH /owners/:id の JSON キーのうち、共通 basic revision が持つ基本情報の項目。
var ownerBasicRequestKeys = []string{
	"owner_name", "owner_name_kana", "birth_date", "company", "postal_code", "address1", "address2",
	"home_postal_code", "home_address1", "home_address2", "phone", "company_phone", "email",
}

// errOwnerBasicRevisionPending は基本情報の保存を拒否する固定 422（F2）。
func errOwnerBasicRevisionPending() error {
	return apperrors.WrapUnprocessableWithCode(apperrors.CodeOwnerBasicRevisionPending, ownerBasicRevisionPendingMessage)
}

// errOwnerRegistrationPending は飼主・ペットの新規登録を拒否する固定 422（F3）。
func errOwnerRegistrationPending() error {
	return apperrors.WrapUnprocessableWithCode(apperrors.CodeOwnerRegistrationPending, ownerRegistrationPendingMessage)
}

// rejectPresentOwnerBasicKeys は ownerBasicRequestKeys のいずれかが JSON のキーとして存在すれば、値（null を含む）に
// 関わらず F2 を返す。`*T` へのバインドでは null が未送信に見えるため、存在判定はキーで行う
// （medicalrecord の rejectPresentHospitalizationKeys と同じ方針）。Go の JSON デコーダと同様に大文字小文字を区別しない。
func rejectPresentOwnerBasicKeys(keys map[string]json.RawMessage) error {
	for _, name := range ownerBasicRequestKeys {
		for key := range keys {
			if strings.EqualFold(key, name) {
				return errOwnerBasicRevisionPending()
			}
		}
	}
	return nil
}

// bindUpdateOwnerRequest は PATCH /owners/:id の body を読み、基本情報のキーが 1 つでもあれば F2 で拒否してから
// updateOwnerRequest へバインドする。JSON として読めない body はキー判定せず、従来どおり ShouldBindJSON の
// 検証（400）に任せる。
func bindUpdateOwnerRequest(c *gin.Context, dst *updateOwnerRequest) error {
	if c.Request != nil && c.Request.Body != nil {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return apperrors.WrapInvalidInput("リクエストボディを読み取れません")
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		var keys map[string]json.RawMessage
		if json.Unmarshal(body, &keys) == nil {
			if err := rejectPresentOwnerBasicKeys(keys); err != nil {
				return err
			}
		}
	}
	if err := c.ShouldBindJSON(dst); err != nil {
		return apperrors.WrapInvalidInput(httpapi.ParseBindError(err))
	}
	return nil
}

// updateInputTouchesBasic は更新入力が基本情報の項目を 1 つでも指定しているかを返す。
// HTTP 境界の rejectPresentOwnerBasicKeys に対する service 層の二重防御で、廃止列への書き込みを防ぐ。
func updateInputTouchesBasic(input *UpdateOwnerInput) bool {
	if input == nil {
		return false
	}
	return input.OwnerName != nil || input.OwnerNameKana != nil || input.BirthDate != nil ||
		input.Company != nil || input.PostalCode != nil || input.Address1 != nil || input.Address2 != nil ||
		input.HomePostalCode != nil || input.HomeAddress1 != nil || input.HomeAddress2 != nil ||
		input.Phone != nil || input.CompanyPhone != nil || input.Email != nil
}

// respondOwnerError は共通基本情報の固定 422（F1/F2/F3）にだけ安定 code を応答へ載せ、
// それ以外は従来どおりの応答にする（httpapi.RespondError は code を出さない）。
func respondOwnerError(c *gin.Context, err error) {
	if apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved) ||
		apperrors.HasCode(err, apperrors.CodeOwnerBasicRevisionPending) ||
		apperrors.HasCode(err, apperrors.CodeOwnerRegistrationPending) {
		httpapi.RespondErrorWithExtras(c, err, nil)
		return
	}
	httpapi.RespondError(c, err)
}

// respondOwner は飼主を応答 DTO へ変換して返す。基本情報が未解決なら空の応答を返さず F1（422）で失敗させる。
func respondOwner(c *gin.Context, status int, owner *model.Owner) {
	resp, err := toOwnerResponse(owner)
	if err != nil {
		respondOwnerError(c, err)
		return
	}
	c.JSON(status, resp)
}

// toOwnerResponses は一覧を応答 DTO へ変換する。1 件でも未解決なら一覧全体を失敗させる（行を黙って落とさない）。
func toOwnerResponses(owners []model.Owner) ([]OwnerResponse, error) {
	items := make([]OwnerResponse, 0, len(owners))
	for i := range owners {
		item, err := toOwnerResponse(&owners[i])
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
