package pet

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/httpapi"
)

// EMR-491: ペットの名前・動物種・性別・生年月日などの基本情報と生死は、pets の列ではなく共通 basic revision の項目になった。
// その改訂 writer・死亡記録 writer・新規登録 writer が接続されるまで、これらを伴う書き込みは成功を装わず固定の 422 で拒否する。
// 拒否は書き込み・採番・外部通知より前に行う。基本情報を含まない更新（主飼主の変更・保険・ペット番号）は従来の経路を通す。

const (
	// petBasicRevisionPendingMessage は基本情報の保存を拒否する固定 422（F4）の文言。
	petBasicRevisionPendingMessage = "ペットの基本情報（名前・動物種・性別・生年月日など）の保存は、共通基本情報の改訂機能の準備中のため現在できません"
	// petLifecycleWritePendingMessage は死亡記録・取り消しを拒否する固定 422（F5）の文言。
	petLifecycleWritePendingMessage = "ペットの死亡記録と取り消しは、共通基本情報の改訂機能の準備中のため現在できません"
	// petRegistrationPendingMessage は既存飼主へのペット直接登録を拒否する固定 422（F6）の文言。
	petRegistrationPendingMessage = "ペットの新規登録は、共通基本情報の登録機能の準備中のため現在できません"
)

// petBasicRequestKeys は PATCH /pets/:id の JSON キーのうち、共通 basic revision が持つ基本情報の項目。
// JSON キーは列名（buildPetUpdate の map キー）と一致するため、repository の列 map の検査にも使う。
// last_visit は pets.last_visit の廃止後に書き込み先が無いため、同じ理由でここに含める。
var petBasicRequestKeys = []string{
	"name", "name_kana", "animal_species_id", "gender", "birth_date", "breed", "color", "blood_type",
	"microchip_number", "weight", "neutered_date", "neutered_status", "acquisition_type",
	"danger_level", "danger_reason", "food", "environment", "phone", "remarks", "name_origin",
	"meeting_story", "last_visit",
}

// petLifecycleFieldKeys は生死を表す列名。pets から廃止され、死亡記録 writer の担当になった。
var petLifecycleFieldKeys = []string{"status", "deceased_at", "deceased_reason"}

// errPetBasicRevisionPending は基本情報の保存を拒否する固定 422（F4）。
func errPetBasicRevisionPending() error {
	return apperrors.WrapUnprocessableWithCode(apperrors.CodePetBasicRevisionPending, petBasicRevisionPendingMessage)
}

// errPetLifecycleWritePending は死亡記録・取り消しを拒否する固定 422（F5）。
func errPetLifecycleWritePending() error {
	return apperrors.WrapUnprocessableWithCode(apperrors.CodePetLifecycleWritePending, petLifecycleWritePendingMessage)
}

// errPetRegistrationPending は既存飼主へのペット直接登録を拒否する固定 422（F6）。
func errPetRegistrationPending() error {
	return apperrors.WrapUnprocessableWithCode(apperrors.CodePetRegistrationPending, petRegistrationPendingMessage)
}

// containsFold は names のいずれかが key と大文字小文字を区別せず一致するかを返す。
func containsFold(names []string, key string) bool {
	for _, name := range names {
		if strings.EqualFold(key, name) {
			return true
		}
	}
	return false
}

// rejectPresentPetBasicKeys は petBasicRequestKeys のいずれかが JSON のキーとして存在すれば、値（null を含む）に
// 関わらず F4 を返す。`*T` へのバインドでは null が未送信に見えるため、存在判定はキーで行う
// （owner の rejectPresentOwnerBasicKeys と同じ方針）。Go の JSON デコーダと同様に大文字小文字を区別しない。
func rejectPresentPetBasicKeys(keys map[string]json.RawMessage) error {
	for key := range keys {
		if containsFold(petBasicRequestKeys, key) {
			return errPetBasicRevisionPending()
		}
	}
	return nil
}

// bindUpdatePetRequest は PATCH /pets/:id の body を読み、基本情報のキーが 1 つでもあれば F4 で拒否してから
// updatePetRequest へバインドする。JSON として読めない body はキー判定せず、従来どおり ShouldBindJSON の
// 検証（400）に任せる。
func bindUpdatePetRequest(c *gin.Context, dst *updatePetRequest) error {
	if c.Request != nil && c.Request.Body != nil {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return apperrors.WrapInvalidInput("リクエストボディを読み取れません")
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		var keys map[string]json.RawMessage
		if json.Unmarshal(body, &keys) == nil {
			if err := rejectPresentPetBasicKeys(keys); err != nil {
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
// HTTP 境界の rejectPresentPetBasicKeys に対する service 層の二重防御で、廃止列への書き込みを防ぐ。
// 明示的な null（DangerReason / NameOrigin / MeetingStory の tri-state）も指定ありとして扱う。
func updateInputTouchesBasic(input *UpdatePetInput) bool {
	if input == nil {
		return false
	}
	return input.AnimalSpeciesID != nil || input.Name != nil || input.PetNameKana != nil || input.Gender != nil ||
		input.BirthDate != nil || input.Breed != nil || input.Color != nil || input.BloodType != nil ||
		input.MicrochipNumber != nil || input.Weight != nil || input.NeuteredDate != nil ||
		input.NeuteredStatus != nil || input.AcquisitionType != nil || input.DangerLevel != nil ||
		input.DangerReason != nil || input.Food != nil || input.Environment != nil || input.Phone != nil ||
		input.LastVisit != nil || input.Remarks != nil || input.NameOrigin != nil || input.MeetingStory != nil
}

// rejectRetiredPetFields は列 map（buildPetUpdate の結果）が pets から廃止された列を含むとき固定 422 を返す。
// 生死の列は F5、基本情報の列は F4。map の走査順に依存しないよう F5 を先に判定する。
// 呼び出しは書き込み・行ロックより前に行うこと。
func rejectRetiredPetFields(fields map[string]any) error {
	for key := range fields {
		if containsFold(petLifecycleFieldKeys, key) {
			return errPetLifecycleWritePending()
		}
	}
	for key := range fields {
		if containsFold(petBasicRequestKeys, key) {
			return errPetBasicRevisionPending()
		}
	}
	return nil
}

// isCommonBasicFixedError は共通基本情報の固定 422（F1/F4/F5/F6）かを返す。
func isCommonBasicFixedError(err error) bool {
	return apperrors.HasCode(err, apperrors.CodeCommonBasicUnresolved) ||
		apperrors.HasCode(err, apperrors.CodePetBasicRevisionPending) ||
		apperrors.HasCode(err, apperrors.CodePetLifecycleWritePending) ||
		apperrors.HasCode(err, apperrors.CodePetRegistrationPending)
}

// respondPetError は共通基本情報の固定 422（F1/F4/F5/F6）にだけ安定 code を応答へ載せ、
// それ以外は従来どおりの応答にする（httpapi.RespondError は code を出さない）。
func respondPetError(c *gin.Context, err error) {
	if isCommonBasicFixedError(err) {
		httpapi.RespondErrorWithExtras(c, err, nil)
		return
	}
	httpapi.RespondError(c, err)
}
