# CODE-QUALITY-005: Service 層命名・インターフェース整合性修正

## 概要

Service 層の命名規則とインターフェース設計に不整合がある。  
`ExamTypeService` の型名、`occupation_service.go` のコード構造順序、`diagnosis_service.go` の定数重複などを修正。

## 優先度

MEDIUM

## 影響ファイル

| ファイル | 問題 |
|---------|-----|
| `backend/internal/service/service.go` | L33: ExaminationType フィールドと ExamTypeService 型名の不整合 |
| `backend/internal/service/occupation_service.go` | L136以降: buildUpdateFields が末尾に配置（構造逆順） |
| `backend/internal/service/diagnosis_service.go` | L53-67: 列名定数の重複定義 |
| `backend/internal/service/diagnosis_service.go` | L241-252: ListNames の limit=10000 マジックナンバー |

---

## 問題一覧

### 1. `service.go:33` — ExaminationType フィールドと ExamTypeService 型名の不整合

```go
// 現状: フィールド名（フル）と型名（略称）が不一致
type Services struct {
    ExaminationType ExamTypeService  // フィールド:フル、型:略称
    // 他は全て一致
    TrimmingCourse TrimmingCourseService
    ...
}
```

**修正案A（推奨）**: インターフェース型名を `ExaminationTypeService` に統一。
```go
// exam_type_service.go
type ExaminationTypeService interface { ... }

// service.go
ExaminationType ExaminationTypeService  // 一致
```

**修正案B**: フィールド名を型名に合わせて `ExamType` に変更（handler.go も全変更が必要）。

影響ファイル: `exam_type_service.go`、`service.go`、`exam_type_handler.go`（参照変更）

---

### 2. `occupation_service.go:136以降` — buildUpdateFields がファイル末尾に逆順配置

他の全サービスの構造:
```
定数ブロック → buildXxxUpdateFields → インターフェース → struct → コンストラクタ → メソッド群
```

`occupation_service.go` のみ `buildOccupationUpdateFields` と列名定数がメソッド群の**後**に配置されており、
コードを上から読む際に「関数が参照する定数・ヘルパーがどこにあるか」が把握しづらい。

**修正方針**: 定数ブロックと `buildOccupationUpdateFields` を `UpdateOccupationInput` struct 定義の直後に移動。

---

### 3. `diagnosis_service.go:53-67` — 列名定数の重複定義

`DiagnosisType` と `DiagnosisName` でそれぞれ `colXxx` 定数を定義しているが、
`name`、`is_active`、`description`、`sort_order` の4つが重複している。

```go
// 現状
const (
    colDiagnosisTypeName      = "name"        // 重複
    colDiagnosisTypeIsActive  = "is_active"   // 重複
    colDiagnosisTypeSortOrder = "sort_order"  // 重複
    ...
)
const (
    colDiagnosisNameName      = "name"        // 重複
    colDiagnosisNameIsActive  = "is_active"   // 重複
    colDiagnosisNameSortOrder = "sort_order"  // 重複
    ...
)
```

**修正方針**: 重複する定数はファイルスコープの共通定数として1箇所に定義する。
```go
// 共通列名（DiagnosisType / DiagnosisName 共通）
const (
    colName      = "name"
    colIsActive  = "is_active"
    colSortOrder = "sort_order"
    colDescription = "description"
)
```

---

### 4. `diagnosis_service.go:241-252` — `ListNames` の limit=10000 マジックナンバー

```go
// 現状: limit=10000 のハードコード（データが10,000件超でサイレント欠落）
items, _, err := s.repo.FindByCategoryID(ctx, clinicID, *typeID, 1, 10000)
```

診断名データが 10,000 件を超えた場合にサイレントなデータ欠落が発生する。

**修正方針**: `DiagnosisNameRepository` に全件取得用メソッドを追加する。
```go
// diagnosis_repository.go に追加
FindAllActive(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)

// diagnosis_service.go
items, err := s.repo.FindAllActive(ctx, clinicID, typeID)
```

---

## 規約参照

- `.claude/rules/naming-conventions.md`: Go 層の命名規則
- `.claude/rules/go-language.md`: インターフェース設計

## テスト

- `ExaminationTypeService` 型名変更後も全ハンドラテストが通ることを確認
- `ListNames` で大量データ（10,000件超）の全件取得が正しく動作することを確認
