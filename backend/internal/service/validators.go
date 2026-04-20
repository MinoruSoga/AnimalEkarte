package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MasterNameMaxLength はマスタ系エンティティの name カラムに許容する
// 最大文字数 (UTF-8 rune count)。BUG-379 対応。
const MasterNameMaxLength = 255

// 日本語エラーメッセージ定数 (BUG-385)
const (
	ErrMsgAtLeastOneField = "少なくとも1つのフィールドを指定してください"
	ErrMsgIDsNotEmpty     = "並び順のIDリストが空です"
	ErrMsgInputNotNil     = "更新内容が指定されていません"
)

// RFC 5322簡易的なメール形式パターン
var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// 電話番号パターン: 03-1234-5678, 090-1234-5678, 09012345678 等
var phonePattern = regexp.MustCompile(`^0\d{1,4}-?\d{1,4}-?\d{4}$`)

// 郵便番号パターン: 123-4567 または 1234567
var postalCodePattern = regexp.MustCompile(`^\d{3}-?\d{4}$`)

// validatePetGender は性別の値がドメイン上有効かを検証する
func validatePetGender(gender string) error {
	if gender == "" {
		return nil
	}
	switch model.PetGender(gender) {
	case model.PetGenderMale, model.PetGenderFemale, model.PetGenderUnknown:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid gender: %s", gender))
	}
}

// validatePetStatus は生存ステータスの値がドメイン上有効かを検証する
func validatePetStatus(status string) error {
	if status == "" {
		return nil
	}
	switch model.PetStatus(status) {
	case model.PetStatusAlive, model.PetStatusDeceased:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid status: %s", status))
	}
}

// validatePetAcquisitionType は入手経路の値がドメイン上有効かを検証する
func validatePetAcquisitionType(t string) error {
	if t == "" {
		return nil
	}
	switch model.AcquisitionType(t) {
	case model.AcquisitionTypePurchase, model.AcquisitionTypeTransfer,
		model.AcquisitionTypeProtected, model.AcquisitionTypeOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid acquisition_type: %s", t))
	}
}

// validatePetDangerLevel は危険度の値がドメイン上有効かを検証する
func validatePetDangerLevel(level string) error {
	if level == "" {
		return nil
	}
	switch model.DangerLevel(level) {
	case model.DangerLevelLow, model.DangerLevelMedium, model.DangerLevelHigh:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid danger_level: %s", level))
	}
}

// validateRequiredName は必須名前フィールドのバリデーションを行う。
// スペースのみ・NULL バイト・制御文字・255文字超が含まれていないかを検証する。
func validateRequiredName(name string) error {
	if strings.TrimSpace(name) == "" {
		return apperrors.WrapInvalidInput("名前を入力してください")
	}
	if utf8.RuneCountInString(name) > MasterNameMaxLength {
		return apperrors.WrapInvalidInput(fmt.Sprintf("名前は%d文字以内で入力してください", MasterNameMaxLength))
	}
	for _, r := range name {
		if r == '\u0000' {
			return apperrors.WrapInvalidInput("名前に無効な文字が含まれています")
		}
		if r < 0x20 && r != '\t' && r != '\n' {
			return apperrors.WrapInvalidInput("名前に無効な文字が含まれています")
		}
	}
	return nil
}

// validateOptionalName は nil 許容の名称バリデーション。
// PATCH 系で nil の場合は更新しない意味なのでスキップ、非 nil の場合のみ検証する。
func validateOptionalName(name *string) error {
	if name == nil {
		return nil
	}
	return validateRequiredName(*name)
}

// validateDiscountRate は割引率が 0〜100 の範囲内かを検証する
func validateDiscountRate(rate float64) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("割引率は0〜100の範囲で入力してください")
	}
	return nil
}

// validateMedicalImageType は診療画像種別がドメイン上有効かを検証する
func validateMedicalImageType(t string) error {
	if t == "" {
		return nil
	}
	switch model.MedicalImageType(t) {
	case model.MedicalImageTypeXray, model.MedicalImageTypeEcho, model.MedicalImageTypePhoto,
		model.MedicalImageTypeEndoscope, model.MedicalImageTypeCT, model.MedicalImageTypeMRI,
		model.MedicalImageTypeMicroscope, model.MedicalImageTypeOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid image_type: %s", t))
	}
}

// validateMembershipType は会員種別がドメイン上有効かを検証する
func validateMembershipType(t model.MembershipType) error {
	if t == "" {
		return nil
	}
	switch t {
	case model.MembershipTypeNonMember, model.MembershipTypeMember,
		model.MembershipTypeDeceased, model.MembershipTypeTransferred:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid membership_type: %s", t))
	}
}

// validateEmailFormat はメールアドレス形式が有効かを検証する（空の場合はスキップ）
func validateEmailFormat(email string) error {
	if email == "" {
		return nil
	}
	if !emailPattern.MatchString(email) {
		return apperrors.WrapInvalidInput("メールアドレスの形式が正しくありません")
	}
	return nil
}

// validatePhoneFormat は電話番号形式が有効かを検証する（空の場合はスキップ）
// 許可形式: 03-1234-5678, 090-1234-5678, 09012345678, 0312345678
func validatePhoneFormat(phone string) error {
	if phone == "" {
		return nil
	}
	if !phonePattern.MatchString(phone) {
		return apperrors.WrapInvalidInput("電話番号の形式が正しくありません（例：090-1234-5678 または 09012345678）")
	}
	return nil
}

// validatePostalCodeFormat は郵便番号形式が有効かを検証する（空の場合はスキップ）
// 許可形式: 123-4567 または 1234567
func validatePostalCodeFormat(postalCode string) error {
	if postalCode == "" {
		return nil
	}
	if !postalCodePattern.MatchString(postalCode) {
		return apperrors.WrapInvalidInput("郵便番号の形式が正しくありません（例：123-4567 または 1234567）")
	}
	return nil
}

// validateVaccineSpecies はワクチン対象種別がドメイン上有効かを検証する
func validateVaccineSpecies(species string) error {
	if species == "" {
		return nil
	}
	switch model.VaccineSpecies(species) {
	case model.VaccineSpeciesDog, model.VaccineSpeciesCat, model.VaccineSpeciesBoth:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid vaccine species: %s", species))
	}
}

// validateAnesthesiaType は麻酔種別がドメイン上有効かを検証する
func validateAnesthesiaType(anesthesia string) error {
	if anesthesia == "" {
		return nil
	}
	switch model.AnesthesiaType(anesthesia) {
	case model.AnesthesiaTypeNone, model.AnesthesiaTypeLocal,
		model.AnesthesiaTypeSedation, model.AnesthesiaTypeGeneral:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid anesthesia_type: %s", anesthesia))
	}
}

// validateTaxType は税種別がドメイン上有効かを検証する
func validateTaxType(taxType string) error {
	if taxType == "" {
		return nil
	}
	switch model.TaxType(taxType) {
	case model.TaxTypeIncluded, model.TaxTypeExcluded, model.TaxTypeExempt:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid tax_type: %s", taxType))
	}
}

// validateItemCategory は明細カテゴリ文字列がドメイン上有効かを検証する
func validateItemCategory(v string) error {
	switch model.ItemCategory(v) {
	case model.ItemCategoryExamination, model.ItemCategoryTest, model.ItemCategoryProcedure,
		model.ItemCategorySurgery, model.ItemCategoryMedicine, model.ItemCategoryFood,
		model.ItemCategoryGoods, model.ItemCategoryOther:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid category: %s", v))
	}
}

// validateItemSource は明細ソース文字列がドメイン上有効かを検証する
func validateItemSource(v string) error {
	switch model.ItemSource(v) {
	case model.ItemSourceMedicalRecord, model.ItemSourceManual, model.ItemSourceHospitalization:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid source: %s", v))
	}
}

// validateCageType はケージ種別がドメイン上有効かを検証する
func validateCageType(cageType string) error {
	if cageType == "" {
		return nil
	}
	switch model.CageType(cageType) {
	case model.CageTypeICU, model.CageTypeDog, model.CageTypeCat, model.CageTypeGeneral:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid cage_type: %s", cageType))
	}
}

// validateNonNegativePrice は価格フィールドが 0 以上かを検証する（nil の場合はスキップ）(BUG-380)
func validateNonNegativePrice(price *int64, fieldName string) error {
	if price == nil {
		return nil
	}
	if *price < 0 {
		return apperrors.WrapInvalidInput(fmt.Sprintf("%sは0以上を入力してください", fieldName))
	}
	return nil
}

// validateCoverageRate は保険補償率が 0〜100 の範囲内かを検証する (BUG-398)
func validateCoverageRate(rate int) error {
	if rate < 0 || rate > 100 {
		return apperrors.WrapInvalidInput("補償率は0〜100の範囲で入力してください")
	}
	return nil
}

// validateOptionalCoverageRate は nil 許容の保険補償率バリデーション (BUG-398)
func validateOptionalCoverageRate(rate *int) error {
	if rate == nil {
		return nil
	}
	return validateCoverageRate(*rate)
}

// validatePassword はパスワード複雑性を検証する（BUG-139）。
// 8文字以上、英字1文字以上、数字1文字以上が必須。
func validatePassword(pw string) error {
	if len(pw) < 8 {
		return apperrors.WrapInvalidInput("パスワードは8文字以上で入力してください")
	}
	var hasLetter, hasDigit bool
	for _, r := range pw {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return apperrors.WrapInvalidInput("パスワードは英字と数字の両方を含めてください")
	}
	return nil
}

// validateCageSize はケージサイズがドメイン上有効かを検証する
func validateCageSize(cageSize string) error {
	if cageSize == "" {
		return nil
	}
	switch model.CageSize(cageSize) {
	case model.CageSizeSmall, model.CageSizeMedium, model.CageSizeLarge:
		return nil
	default:
		return apperrors.WrapInvalidInput(fmt.Sprintf("invalid cage_size: %s", cageSize))
	}
}
