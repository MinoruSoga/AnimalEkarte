# TASK-028: その他マスタ MEDIUM 問題 6件

## 優先度

MEDIUM

---

## 問題 1: Reorder の HTTP ステータスが cage=204 / 他=200+body で不統一

### ファイル
`backend/internal/handler/cage_handler.go:142`（204 を返す）
他: vaccine / medicine / trimming / occupation / insurance / animal_species / procedure / consultation ハンドラの Reorder（200 + `{"message": "reordered"}` を返す）

### 問題
全マスタ系の Reorder で HTTP ステータスが混在。フロントエンドのエラーハンドリングが分岐する。

### 修正案
全ドメインを `c.Status(http.StatusNoContent)` に統一する（reservation_type_group と cage が正しい実装）。

---

## 問題 2: Procedure / Consultation ハンドラでデフォルト値（税種・税率）を組み立て

### ファイル
`backend/internal/handler/procedure_handler.go:69-90`
`backend/internal/handler/consultation_handler.go:69-88`

### 問題
```go
taxType := model.TaxTypeExcluded
if req.TaxType != nil { taxType = *req.TaxType }
taxRate := 0.10
if req.TaxRate != nil { taxRate = *req.TaxRate }
```
デフォルト値適用はビジネスロジックであり service 層の責務。`medicine_service.go` が正しい参照実装（service 内でデフォルト設定）。

### 修正案
`CreateProcedureInput` / `CreateConsultationInput` のフィールドをポインタ型にして service 層でデフォルト補完する。

---

## 問題 3: cage / procedure / consultation ハンドラが生 model を返す（Response DTO なし）

### ファイル
`backend/internal/handler/cage_handler.go:31`
`backend/internal/handler/procedure_handler.go:31`
`backend/internal/handler/consultation_handler.go:31`

### 問題
`c.JSON(http.StatusOK, cage)` と生モデルを返している。medicine / insurance / occupation / trimming_course / inquiry_template は `toXxxResponse()` を通している。DB カラム追加時に API コントラクトが意図せず変化するリスクがある。

### 修正案
`toCageResponse()` / `toProcedureResponse()` / `toConsultationResponse()` を各 `*_response.go` に追加し、全レスポンスで使用する。

---

## 問題 4: animal_species_repository の FindAll が is_active=true を強制フィルタ

### ファイル
`backend/internal/repository/animal_species_repository.go:31-39`

### 問題
```go
Where("is_active = ?", true)
```
他の全ドメイン（vaccine, medicine, occupation 等）は全件返してフロントでフィルタするか、is_active フィルタなし。animal_species だけが repository レベルで強制フィルタをかけており、管理画面で is_active=false の種別を表示・編集できない。

### 修正案
`FindAll` の WHERE 条件から `is_active = true` を削除し、フロントまたは service 層でのフィルタに委ねる。

---

## 問題 5: medicine_service のカテゴリ削除に WrapConflict でなく WrapInvalidInput を使用

### ファイル
`backend/internal/service/medicine_service.go:276-283`

### 問題
```go
// 現状: 400 Bad Request を返す
return apperrors.WrapInvalidInput(fmt.Sprintf("..."))
```
他ドメインの FK 依存チェックは全て `apperrors.WrapConflict`（409 Conflict）。medicine カテゴリだけ 400 が返り、フロントエンドのエラーハンドリングが非対称になる。

### 修正案
```go
return apperrors.WrapConflict(
    fmt.Sprintf("このカテゴリには%d件の薬剤が含まれています。先に薬剤を移動または削除してください", count))
```

---

## 問題 6: inquiry_template の Delete に FK 依存チェックなし

### ファイル
`backend/internal/service/inquiry_template_service.go:91-98`

### 問題
問診テンプレートが使用中でも削除できてしまう可能性がある。procedure / consultation / vaccine / medicine は全て削除前の依存チェックを実装済み。

### 修正案
`inquiry_answers` や予約との紐付けテーブルが存在する場合は `CountUsageByTemplateID` を追加してチェックし、`apperrors.WrapConflict` で 409 を返す。依存テーブルが存在しない場合は「設計上の意図」としてコードコメントで明示する。
