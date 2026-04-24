# BUG-388: マスタハンドラの関数定義順序が不統一（Get が List より先に定義されている）

## 概要
15ファイルのマスタハンドラのうち、8ファイルで `GetXxx` が `ListXxx` より先に定義されている。プロジェクトの標準順序は `List → Get → Create → Update → Delete → Reorder` だが、多数のハンドラが `Get → List → ...` の順になっており、コードの一貫性が失われている。機能的な問題はないが、コードレビューと IDE ナビゲーションの効率が下がる。

## 再現手順
コードレビューで確認可能。

## 期待する動作
- 全マスタハンドラで `List → Get → Create → Update → Delete → Reorder` の順序に統一すること

## 現状コード

### 問題のあるハンドラ（Get が先）
```
checkup_type_handler.go:  GetCheckupType(16行目), ListCheckupTypes(34行目)
chief_complaint_handler.go: GetChiefComplaint(16行目), ListChiefComplaints(34行目)
diagnosis_handler.go:     GetDiagnosisType(17行目), ListDiagnosisTypes(35行目)
exam_type_handler.go:     GetExaminationType(16行目), ListExaminationTypes(34行目)
occupation_handler.go:    GetOccupation(16行目), ListOccupations(34行目)
procedure_handler.go:     GetProcedure(16行目), ListProcedures(34行目)
reservation_type_handler.go: GetReservationType(17行目), ListReservationTypes(35行目)
vaccine_handler.go:       GetVaccine(16行目), ListVaccines(34行目)
trimming_master_handler.go: GetTrimmingCourse(16行目), ListTrimmingCourses(34行目)
```

### 正しい順序のハンドラ（参照実装）
```
animal_species_handler.go:  ListAnimalSpecies(14行目), GetAnimalSpecies(24行目) ✓
insurance_handler.go:       ListInsurances(16行目), GetInsurance(30行目) ✓
merchandise_item_handler.go: ListMerchandiseItems(14行目), GetMerchandiseItem(34行目) ✓
reservation_type_group_handler.go: ListReservationTypeGroups(13行目), GetReservationTypeGroup(27行目) ✓
```

## 影響範囲

| 対象ファイル | 変更内容 | 影響 |
|------------|---------|------|
| checkup_type_handler.go | Get/List 順序入れ替え | 行番号変更のみ、機能変更なし |
| chief_complaint_handler.go | 同上 | 同上 |
| diagnosis_handler.go | 同上 | 同上 |
| exam_type_handler.go | 同上 | 同上 |
| occupation_handler.go | 同上 | 同上 |
| procedure_handler.go | 同上 | 同上 |
| reservation_type_handler.go | 同上 | 同上 |
| vaccine_handler.go | 同上 | 同上 |
| trimming_master_handler.go | 同上（Course/Option 両方） | 同上 |

## 修正方針

各ファイルで `GetXxx` 関数全体（godoc コメントを含む）を `ListXxx` の後に移動する。

### 例: `checkup_type_handler.go`
```go
// 修正前の順序
func (h *Handler) GetCheckupType(c *gin.Context) { ... }   // 16行目
func (h *Handler) ListCheckupTypes(c *gin.Context) { ... } // 34行目

// 修正後の順序
func (h *Handler) ListCheckupTypes(c *gin.Context) { ... } // 先に定義
func (h *Handler) GetCheckupType(c *gin.Context) { ... }   // 後に定義
```

機能コードは一切変更しない。関数の順序の入れ替えのみ。

## 準拠すべきプロジェクト規約・ベストプラクティス

### プロジェクト標準順序（animal_species_handler.go 参照実装）
`List → Get → Create → Update → Delete → Reorder`
— RESTful リソース操作の論理的な順序（コレクション操作 → 単体操作）

### プロジェクト内参照実装
`backend/internal/handler/animal_species_handler.go` — 標準順序の正しい実装

## 優先度
**Low** — 機能上の問題なし。コードの一貫性・可読性の問題。他の修正と合わせて一括対応が効率的。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/checkup_type_handler.go:16,34` — 修正対象
- `backend/internal/handler/chief_complaint_handler.go:16,34` — 修正対象
- `backend/internal/handler/diagnosis_handler.go:17,35` — 修正対象
- `backend/internal/handler/exam_type_handler.go:16,34` — 修正対象
- `backend/internal/handler/occupation_handler.go:16,34` — 修正対象
- `backend/internal/handler/procedure_handler.go:16,34` — 修正対象
- `backend/internal/handler/reservation_type_handler.go:17,35` — 修正対象
- `backend/internal/handler/vaccine_handler.go:16,34` — 修正対象
- `backend/internal/handler/trimming_master_handler.go` — 修正対象
