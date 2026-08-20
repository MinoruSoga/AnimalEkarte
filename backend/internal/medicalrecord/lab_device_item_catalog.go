package medicalrecord

import "github.com/animal-ekarte/backend/internal/model"

// labDeviceItemCatalogRow is the embedded 城東観測カタログ (25 rows).
// DisplayName is a seed hint only. It is not written to the database.
type labDeviceItemCatalogRow struct {
	SourceType     model.LabImportSourceType
	DeviceItemCode string
	DisplayName    string
	Unit           string
	ValueShape     string
	SortOrder      int
}

// LabDeviceItemCatalogCount is the accepted 城東 seed size (NX600 16 + AU10V 1 + 尿 8 + IDEXX 11).
const LabDeviceItemCatalogCount = 36

type labDeviceDefault struct {
	SourceType model.LabImportSourceType
	Name       string
	SortOrder  int
}

func labDeviceDefaults() []labDeviceDefault {
	return []labDeviceDefault{
		{model.LabImportSourceTypeFujiNX600, "NX600", 10},
		{model.LabImportSourceTypeFujiAU10V, "AU10V", 20},
		{model.LabImportSourceTypeArkrayPU4010, "尿（PU-4010）", 30},
		{model.LabImportSourceTypeIDEXXVetLab, "IDEXX VetLab", 40},
	}
}

func isLabDeviceSourceType(sourceType string) bool {
	switch model.LabImportSourceType(sourceType) {
	case model.LabImportSourceTypeFujiNX600,
		model.LabImportSourceTypeFujiAU10V,
		model.LabImportSourceTypeArkrayPU4010,
		model.LabImportSourceTypeIDEXXVetLab:
		return true
	default:
		return false
	}
}

func isLabDeviceValueShape(shape string) bool {
	switch shape {
	case model.LabDeviceValueShapeNumeric,
		model.LabDeviceValueShapeInequality,
		model.LabDeviceValueShapeQualAndNum,
		model.LabDeviceValueShapeDash,
		model.LabDeviceValueShapeText:
		return true
	default:
		return false
	}
}

func labDeviceItemCatalog() []labDeviceItemCatalogRow {
	return []labDeviceItemCatalogRow{
		{model.LabImportSourceTypeFujiNX600, "Na-P", "Na", "mEq/l", model.LabDeviceValueShapeNumeric, 10},
		{model.LabImportSourceTypeFujiNX600, "K-P", "K", "mEq/l", model.LabDeviceValueShapeNumeric, 20},
		{model.LabImportSourceTypeFujiNX600, "Cl-P", "Cl", "mEq/l", model.LabDeviceValueShapeNumeric, 30},
		{model.LabImportSourceTypeFujiNX600, "LIP-P", "LIP-P", "U/l", model.LabDeviceValueShapeNumeric, 40},
		{model.LabImportSourceTypeFujiNX600, "TP-P", "TP", "g/dl", model.LabDeviceValueShapeNumeric, 50},
		{model.LabImportSourceTypeFujiNX600, "ALB-P", "ALB", "g/dl", model.LabDeviceValueShapeNumeric, 60},
		{model.LabImportSourceTypeFujiNX600, "ALPi-P", "ALP", "U/l", model.LabDeviceValueShapeNumeric, 70},
		{model.LabImportSourceTypeFujiNX600, "GLU-P", "Glu", "mg/dl", model.LabDeviceValueShapeNumeric, 80},
		{model.LabImportSourceTypeFujiNX600, "TBIL-P", "TBIL", "mg/dl", model.LabDeviceValueShapeNumeric, 90},
		{model.LabImportSourceTypeFujiNX600, "IP-P", "IP", "mg/dl", model.LabDeviceValueShapeNumeric, 100},
		{model.LabImportSourceTypeFujiNX600, "TCHO-P", "CHOL", "mg/dl", model.LabDeviceValueShapeNumeric, 110},
		{model.LabImportSourceTypeFujiNX600, "GGT-P", "GGT", "U/l", model.LabDeviceValueShapeNumeric, 120},
		{model.LabImportSourceTypeFujiNX600, "GPT-P", "GPT", "U/l", model.LabDeviceValueShapeNumeric, 130},
		{model.LabImportSourceTypeFujiNX600, "Ca-P", "Ca", "mg/dl", model.LabDeviceValueShapeNumeric, 140},
		{model.LabImportSourceTypeFujiNX600, "CRE-P", "Cre", "mg/dl", model.LabDeviceValueShapeNumeric, 150},
		{model.LabImportSourceTypeFujiNX600, "BUN-P", "BUN", "mg/dl", model.LabDeviceValueShapeNumeric, 160},
		{model.LabImportSourceTypeFujiAU10V, "vf-SAA", "vf-SAA", "ug/mL", model.LabDeviceValueShapeInequality, 10},
		{model.LabImportSourceTypeArkrayPU4010, "GLU", "尿糖", "mg/dL", model.LabDeviceValueShapeDash, 10},
		{model.LabImportSourceTypeArkrayPU4010, "PRO", "尿蛋白", "mg/dL", model.LabDeviceValueShapeQualAndNum, 20},
		{model.LabImportSourceTypeArkrayPU4010, "BIL", "ビリルビン", "mg/dL", model.LabDeviceValueShapeDash, 30},
		{model.LabImportSourceTypeArkrayPU4010, "URO", "ウロビリノーゲン", "mg/dL", model.LabDeviceValueShapeText, 40},
		{model.LabImportSourceTypeArkrayPU4010, "PH", "pH", "", model.LabDeviceValueShapeNumeric, 50},
		{model.LabImportSourceTypeArkrayPU4010, "BLD", "潜血", "mg/dL", model.LabDeviceValueShapeQualAndNum, 60},
		{model.LabImportSourceTypeArkrayPU4010, "KET", "尿ケトン", "mg/dL", model.LabDeviceValueShapeDash, 70},
		{model.LabImportSourceTypeArkrayPU4010, "NIT", "NIT", "mg/dL", model.LabDeviceValueShapeDash, 80},
		// IDEXX VetLab Station: blood-count labels observed in 2026-08-19 capture (LAB_DEVICE_CONNECTIVITY.md §IDEXX)
		{model.LabImportSourceTypeIDEXXVetLab, "WBC", "WBC", "K/uL", model.LabDeviceValueShapeNumeric, 10},
		{model.LabImportSourceTypeIDEXXVetLab, "RBC", "RBC", "M/uL", model.LabDeviceValueShapeNumeric, 20},
		{model.LabImportSourceTypeIDEXXVetLab, "HCT", "HCT", "%", model.LabDeviceValueShapeNumeric, 30},
		{model.LabImportSourceTypeIDEXXVetLab, "HGB", "HGB", "g/dL", model.LabDeviceValueShapeNumeric, 40},
		{model.LabImportSourceTypeIDEXXVetLab, "PLT", "PLT", "K/uL", model.LabDeviceValueShapeNumeric, 50},
		{model.LabImportSourceTypeIDEXXVetLab, "NEU", "NEU", "K/uL", model.LabDeviceValueShapeNumeric, 60},
		{model.LabImportSourceTypeIDEXXVetLab, "LYM", "LYM", "K/uL", model.LabDeviceValueShapeNumeric, 70},
		{model.LabImportSourceTypeIDEXXVetLab, "MONO", "MONO", "K/uL", model.LabDeviceValueShapeNumeric, 80},
		{model.LabImportSourceTypeIDEXXVetLab, "EOS", "EOS", "K/uL", model.LabDeviceValueShapeNumeric, 90},
		{model.LabImportSourceTypeIDEXXVetLab, "BASO", "BASO", "K/uL", model.LabDeviceValueShapeNumeric, 100},
		{model.LabImportSourceTypeIDEXXVetLab, "RETIC", "RETIC", "K/uL", model.LabDeviceValueShapeNumeric, 110},
	}
}
