# TASK-055: diagnosis_handler — parsePagination の結果を使用していないデッドコード

## 優先度

LOW

---

## 概要

`diagnosis_handler.go` の `ListDiagnosisTypes` / `ListDiagnosisNames` は `parsePagination(c)` を呼び出しているが、  
取得した `page` / `limit` を service の `List()` に渡していない。結果として pagination は機能せず、  
`parsePagination` の呼び出しがデッドコードになっている。

他のマスタハンドラはページネーション不使用（全件返却が仕様）のため `parsePagination` を呼ばないが、  
`diagnosis_handler` のみが呼び出しと実際の動作が乖離している。

---

## 問題箇所

### ファイル
`backend/internal/handler/diagnosis_handler.go`

```go
// ❌ ListDiagnosisTypes（L35-52 概算）
func (h *DiagnosisHandler) ListDiagnosisTypes(c *gin.Context) {
    ctx := c.Request.Context()
    clinicID, _ := extractClinicID(c)
    page, limit := parsePagination(c)  // ← 呼び出しているが...
    
    diagnosisTypes, _, err := h.service.ListDiagnosisTypes(ctx, clinicID)  // ← page/limit を渡していない
    // ...
}

// ❌ ListDiagnosisNames（L167-201 概算）も同様
```

---

## 修正方針

### 選択肢 A: diagnosis もページネーション不使用の仕様であれば `parsePagination` 呼び出しを削除

```go
// ✅ 修正後（ページネーション不使用）
func (h *DiagnosisHandler) ListDiagnosisTypes(c *gin.Context) {
    ctx := c.Request.Context()
    clinicID, _ := extractClinicID(c)
    // parsePagination の呼び出しを削除

    diagnosisTypes, err := h.service.ListDiagnosisTypes(ctx, clinicID)
    // ...
    c.JSON(http.StatusOK, mapSlice(diagnosisTypes, toDiagnosisTypeResponse))
}
```

### 選択肢 B: ページネーションを正しく実装する（service シグネチャも変更）

他のマスタと同様に全件返却が仕様であれば**選択肢 A が正しい**。  
実装意図を確認してから対応すること。

---

## 備考

`mapSlice` を使った統一レスポンスへの変更も同時に行うと、TASK-056 との合わせ技になる。
