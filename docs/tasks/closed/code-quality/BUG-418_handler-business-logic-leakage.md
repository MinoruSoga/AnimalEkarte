# BUG-418: ハンドラ層へのビジネスロジック漏洩（diagnosis / consultation / trimming）

## 概要

`diagnosis_handler.go`、`consultation_handler.go`、`trimming_handler.go` の3ファイルで、
本来 Service 層が担うべきロジックがハンドラ層に実装されている。
責任分離（Clean Architecture）に違反しており、ハンドラのテスト難度が増す。

## 問題箇所

### 1. `diagnosis_handler.go:181-200` — ListDiagnosisNames の条件分岐

```go
// diagnosis_handler.go:181-200
func (h *Handler) ListDiagnosisNames(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    typeIDStr := c.Query("diagnosis_type_id")
    var names []model.DiagnosisName
    var err error
    if typeIDStr != "" {                           // ← ハンドラ内で分岐（サービス層で処理すべき）
        typeID, parseErr := strconv.ParseUint(...)
        if parseErr != nil { ... }
        names, err = h.svc.Diagnosis.ListNamesByTypeID(c.Request.Context(), clinicID, typeID)
    } else {
        names, err = h.svc.Diagnosis.ListAllNames(c.Request.Context(), clinicID)
    }
    // ...
}
```

**問題**: `typeIDStr` の有無判定によるサービス呼び分けはハンドラの責務ではない。
サービス層に `ListNames(ctx, clinicID, typeID *uint64)` として統合すべき。

### 2. `consultation_handler.go:61-65` — TaxType の変換ロジック

```go
// consultation_handler.go:61-65
var taxType *string
if input.TaxType != "" {                           // ← ハンドラ内で変換ロジック
    taxType = &input.TaxType
}
svcInput := &service.CreateConsultationInput{
    TaxType: taxType,
    ...
}
```

**問題**: `input.TaxType` の空文字チェックとポインタ変換はサービス層の Input DTO または
サービスメソッド内で処理すべき。ハンドラは Request → Service Input の単純マッピングに留めるべき。

### 3. `trimming_handler.go:63-66` — 手動ループによるマッピング

```go
// trimming_handler.go:63-66（Create リクエストのオプション変換）
var optionIDs []uint64
for _, id := range input.OptionIDs {              // ← mapSlice 未使用、手動ループ
    optionIDs = append(optionIDs, id)
}
```

**問題**: 他のマスタハンドラが `mapSlice(items, toXxxResponse)` で統一しているのに対し、
trimming_handler は手動ループを使用。さらにロジックが混在している。

## 修正方針

### diagnosis_handler.go
```go
// サービス側に統合（nullable typeID を受け取る）
// service interface を変更:
ListNames(ctx context.Context, clinicID uint64, typeID *uint64) ([]model.DiagnosisName, error)

// ハンドラはクエリパース→呼び出しのみ
var typeID *uint64
if s := c.Query("diagnosis_type_id"); s != "" {
    id, err := strconv.ParseUint(s, 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid diagnosis_type_id"))
        return
    }
    typeID = &id
}
names, err := h.svc.Diagnosis.ListNames(c.Request.Context(), clinicID, typeID)
```

### consultation_handler.go
```go
// Request 構造体の TaxType をポインタ型に変更するか、サービス Input に委譲
// ハンドラ:
svcInput := &service.CreateConsultationInput{
    TaxType: nilIfEmpty(input.TaxType),   // ヘルパー化
    ...
}
```

### trimming_handler.go
```go
// optionIDs は直接 input.OptionIDs を渡す（uint64 スライスなら型変換不要）
svcInput := &service.CreateTrimmingInput{
    OptionIDs: input.OptionIDs,
    ...
}
```

## 影響ファイル

- `backend/internal/handler/diagnosis_handler.go` — 行 181-200
- `backend/internal/handler/consultation_handler.go` — 行 61-65
- `backend/internal/handler/trimming_handler.go` — 行 63-66

## 優先度

**Medium** — アーキテクチャ規約違反。テスト難度・保守性に影響する。

## 関連規約

- `.claude/CLAUDE.md` — handler → service → repository の軽量レイヤード徹底
