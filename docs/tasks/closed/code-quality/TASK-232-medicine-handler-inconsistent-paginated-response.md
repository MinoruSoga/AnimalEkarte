# TASK-232: medicine_handler.go — マスタ List が PaginatedResponse を返しており他マスタと不統一

## 優先度
Medium

## 対象ファイル
- `backend/internal/handler/medicine_handler.go`

## 問題概要
`ListMedicines`（行32）が `newPaginatedResponse(...)` でラップしたレスポンスを返しているが、
同じマスタ系 handler（vaccine, procedure, payment_method_master）は全て生配列 `mapSlice(...)` を直接返す設計になっている。

フロントエンドが `response.data` でアクセスするか `response[0]` でアクセスするかが
ドメインによって異なってしまい、フロントエンド側の実装に混乱をもたらす。

## 現状コード（medicine_handler.go:32）

```go
// ❌ PaginatedResponse でラップ（他マスタと不統一）
c.JSON(http.StatusOK, newPaginatedResponse(mapSlice(medicines, toMedicineResponse), total, page, limit))
```

## 比較（他マスタの実装）

```go
// vaccine_handler.go:31 ✅ 生配列
c.JSON(http.StatusOK, mapSlice(vaccines, toVaccineResponse))

// procedure_handler.go:27 ✅ 生配列
c.JSON(http.StatusOK, mapSlice(procedures, toProcedureResponse))

// payment_method_master_handler.go:26 ✅ 生配列
c.JSON(http.StatusOK, mapSlice(ms, toPaymentMethodResponse))
```

## あるべき姿

マスタ系は全件返却（ページネーション不要）で統一する。

```go
// medicine_handler.go
func (h *Handler) ListMedicines(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    medicines, err := h.svc.Medicine.List(c.Request.Context(), clinicID)
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, mapSlice(medicines, toMedicineResponse))
}
```

medicine service の `List` が `([]model.Medicine, int64, error)` を返す場合は
`total` を削除し `([]model.Medicine, error)` に変更するか、
`total` を無視して生配列を返すよう handler を修正する。

## 確認事項
- フロントエンドが `medicine` の一覧を `response.data` で参照している場合、フロント修正も必要

## 完了条件
- [ ] `ListMedicines` の応答を生配列に変更
- [ ] フロントエンド側の medicine 一覧取得コードを確認・修正
- [ ] `go test ./backend/internal/...` がパス
