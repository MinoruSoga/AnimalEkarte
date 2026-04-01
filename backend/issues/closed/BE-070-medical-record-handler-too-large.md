# BE-070: medical_record_handler の CreateMedicalRecord が 137 行で肥大化

**Status**: Closed
**Priority**: Medium
**Affects**: `backend/internal/handler/medical_record_handler.go`
**Date Created**: 2026-03-26
**Related**: BE-067

## Summary

`CreateMedicalRecord` ハンドラ（:106-243）が 137 行に膨れ上がっており、日付解決・record_no 生成・ID 型変換・バリデーション・ClinicalPlan の原子的作成まで handler 内で行っている。これらはすべて service の責務であり、handler はリクエストバインドとレスポンス返却のみに絞るべき。BE-067（generateRecordNo 移動）完了後に対応する。

## 現状のコード（概略）

```go
// backend/internal/handler/medical_record_handler.go:106-243
func (h *MedicalRecordHandler) CreateMedicalRecord(c *gin.Context) {
    // 1. ID型変換（uint64 → 複数）
    // 2. visit_date 解決ロジック（今日日付フォールバック）
    // 3. generateRecordNo() 呼び出し
    // 4. model.MedicalRecord 直接組み立て
    // 5. model.ClinicalPlan 直接組み立て
    // 6. repo への直接呼び出し（service をバイパス？）
    // 7. 複数の c.JSON(StatusBadRequest, ...) 呼び出し（RespondError 未使用）
}
```

## 必要な変更

### handler の責務（あるべき姿）

```go
// backend/internal/handler/medical_record_handler.go
func (h *MedicalRecordHandler) CreateMedicalRecord(c *gin.Context) {
	clinicID := middleware.GetClinicID(c)

	var req CreateMedicalRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, apperrors.WrapInvalidInput(err.Error()))
		return
	}

	record, err := h.service.Create(c.Request.Context(), toCreateMedicalRecordInput(clinicID, req))
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toMedicalRecordResponse(record))
}
```

### service 層に移動すべき処理

1. `visit_date` がゼロ値の場合に今日日付を使うロジック
2. `generateRecordNo()` 呼び出し（BE-067 で service 移動済みが前提）
3. `ClinicalPlan` の原子的作成（トランザクション）
4. ID 型変換・バリデーション

### 依存関係

BE-067（generateRecordNo service 移動）を先に完了してから着手する。

## 完了条件

- [x] `CreateMedicalRecord` ハンドラを 30 行以内にスリム化（22行）
- [x] 日付解決・ID変換・モデル組立を `buildMedicalRecord()` 純粋関数に抽出
- [x] ClinicalPlan ベストエフォート作成を `createClinicalPlanIfNeeded()` に抽出
- [x] handler 内の全 `c.JSON(StatusBadRequest, ...)` を `RespondError` に統一
- [x] `docker compose exec backend go test ./... -v` がパス

## クローズ情報

- **Closed At**: 2026-03-26
- **変更ファイル**: `backend/internal/handler/medical_record_handler.go`
