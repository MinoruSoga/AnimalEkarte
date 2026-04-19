# TASK-052: ENUM 入力値バリデーター欠落 — vaccine / procedure / cage

## 優先度

MEDIUM

---

## 概要

`validators.go` には `validatePetGender()`・`validatePetAcquisitionType()` 等のペット系 ENUM バリデーターが定義されているが、マスタ系 ENUM フィールドに対応するバリデーターが存在しない。

不正な文字列が DB に渡ると GORM が PostgreSQL の ENUM 型エラーを返し、500 Internal Server Error として露出する。サービス層で弾いて 400 Bad Request を返すべきだ。

---

## 対象フィールドと現状

| ドメイン | フィールド | 型 | バリデーター |
|---------|-----------|-----|------------|
| vaccine | `Species` | `model.VaccineSpecies` | ❌ なし |
| procedure | `Anesthesia` | `model.AnesthesiaType` | ❌ なし |
| procedure | `TaxType` | `model.TaxType` | ❌ なし |
| cage | `CageType` | `model.CageType` | ❌ なし |
| cage | `CageSize` | `model.CageSize` | ❌ なし |

---

## 問題

```go
// validators.go ✅ ペット系は存在する
func validatePetGender(gender string) error {
    switch model.PetGender(gender) {
    case model.PetGenderMale, model.PetGenderFemale, model.PetGenderUnknown:
        return nil
    default:
        return apperrors.WrapInvalidInput(fmt.Errorf("invalid gender: %s", gender))
    }
}

// ❌ マスタ系 ENUM は存在しない
// validateVaccineSpecies() → 未定義
// validateAnesthesiaType() → 未定義
// validateCageType()       → 未定義
// validateCageSize()       → 未定義
// validateTaxType()        → 未定義（vaccine/procedure/cage で共通）
```

---

## 修正方針

### 1. validators.go に ENUM バリデーター関数を追加

```go
// vaccine
func validateVaccineSpecies(species string) error {
    switch model.VaccineSpecies(species) {
    case model.VaccineSpeciesDog, model.VaccineSpeciesCat, /* ... 全値 */ :
        return nil
    default:
        return apperrors.WrapInvalidInput(fmt.Errorf("invalid vaccine species: %s", species))
    }
}

// procedure
func validateAnesthesiaType(anesthesia string) error { /* 同様 */ }
func validateTaxType(taxType string) error { /* 同様 */ }

// cage
func validateCageType(cageType string) error { /* 同様 */ }
func validateCageSize(cageSize string) error { /* 同様 */ }
```

### 2. サービス層の Create/Update でバリデーターを呼び出す

```go
// vaccine_service.go Create/Update
func (s *vaccineService) Create(ctx context.Context, clinicID uint64, input CreateVaccineInput) (*model.Vaccine, error) {
    if input.Species != "" {
        if err := validateVaccineSpecies(input.Species); err != nil {
            return nil, err
        }
    }
    // ...
}
```

---

## 備考

- TASK-049（procedure/cage の handler model 型変換 → service 移動）の実施後、ENUM 値は `*string` として受け取るため、変換前にバリデーションを実施すると自然に組み込める。
- `validateTaxType` は medicine_service.go で独自に ENUM チェックを行っている場合、共通化して `validators.go` に移動することも検討すること。
- `TaxType` は procedure と cage で共通の ENUM のため、バリデーター関数は 1 つ定義すれば両方で使い回せる。
