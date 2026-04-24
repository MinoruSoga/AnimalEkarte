# CODE-QUALITY-231: service Delete テストの 409 Conflict ケース欠落

## 概要

`cage_service_test.go` / `insurance_service_test.go` / `occupation_service_test.go` の
`Delete` テストに、依存レコードが存在する場合（FK 依存チェックで弾かれるケース）の
テストが欠落している。

---

## 対象ファイルと欠落ケース

### 1. cage_service_test.go

**ファイル:** `backend/internal/service/cage_service_test.go:344`

**現状:** `TestCageService_Delete` に正常削除・Not Found のみ

**欠落ケース:**
```go
{
    name: "使用中のケージは削除できない",
    setup: func(m *mockCageRepository) {
        m.On("FindByID", ...).Return(&model.Cage{ID: 1}, nil)
        m.On("CountRecordsByCageID", ..., uint64(1)).Return(int64(3), nil)
    },
    wantErr: true,
    wantErrType: apperrors.ErrConflict,
},
```

### 2. insurance_service_test.go

**ファイル:** `backend/internal/service/insurance_service_test.go:382`

**現状:** `TestInsuranceService_Delete` に正常削除・Not Found のみ

**欠落ケース:**
```go
{
    name: "使用中の保険は削除できない",
    setup: func(m *mockInsuranceRepository) {
        m.On("FindByID", ...).Return(&model.Insurance{ID: 1}, nil)
        m.On("CountPetsByInsuranceID", ..., uint64(1)).Return(int64(2), nil)
    },
    wantErr: true,
    wantErrType: apperrors.ErrConflict,
},
```

### 3. occupation_service_test.go

**ファイル:** `backend/internal/service/occupation_service_test.go:326`

**現状:** `TestOccupationService_Delete` に正常削除・Not Found のみ

**欠落ケース:**
```go
{
    name: "スタッフが所属している職種は削除できない",
    setup: func(m *mockOccupationRepository) {
        m.On("FindByID", ...).Return(&model.Occupation{ID: 1}, nil)
        m.On("CountStaffsByOccupationID", ..., uint64(1)).Return(int64(5), nil)
    },
    wantErr: true,
    wantErrType: apperrors.ErrConflict,
},
```

---

## 比較（正しい実装例: medicine_service_test.go）

```go
// medicine_service_test.go
{
    name: "使用中の薬品は削除できない",
    setup: func(m *mockMedicineRepository) {
        m.On("FindByID", ...).Return(&model.Medicine{ID: 1}, nil)
        m.On("CountUsageByMedicineID", ..., uint64(1)).Return(int64(1), nil)
    },
    wantErr:     true,
    wantErrType: apperrors.ErrConflict,
},
```

---

## 修正方針

各 `TestXxxService_Delete` のテーブルに `count > 0 → 409 Conflict` ケースを追加する。
service 実装自体は FK チェックが正しく実装されているため、テストの追加のみで対応可能。

---

## 優先度

MEDIUM — 実装バグではなくテストカバレッジ欠落。
FK 依存チェックのリグレッションを検知できないリスクあり。
