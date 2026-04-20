# TASK-163: shift_template_handler.go — Handler 内のビジネスロジック混入

## 概要
`CreateShiftTemplate` ハンドラ内でデフォルト値の設定・型変換・nil ガードが実施されており、
Handler の責務である「リクエスト解析 + Service 委譲」を超えている。
これらのロジックは Service 層（`Create` メソッド内）に移すべき。

## 優先度
Medium（責務分離）

## 対象ファイル
`backend/internal/handler/shift_template_handler.go`（行 128〜170）

---

## 問題の詳細

### 現状コード（CreateShiftTemplate、行 128〜170）
```go
func (h *Handler) CreateShiftTemplate(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    var req createShiftTemplateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }

    // ❌ デフォルト値の設定ロジックが Handler にある
    isActive := true
    if req.IsActive != nil {
        isActive = *req.IsActive
    }

    // ❌ []shiftTemplateBreakRequest → []service.ShiftBreakTemplateInput の変換ロジックが Handler にある
    breaks := make([]service.ShiftBreakTemplateInput, 0, len(req.Breaks))
    for _, b := range req.Breaks {
        breaks = append(breaks, service.ShiftBreakTemplateInput{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
    }

    // ❌ *string へのラップロジックが Handler にある
    var startTime, endTime *string
    if req.StartTime != "" {
        startTime = &req.StartTime
    }
    if req.EndTime != "" {
        endTime = &req.EndTime
    }

    tpl, err := h.svc.ShiftTemplate.Create(c.Request.Context(), clinicID, &service.CreateShiftTemplateInput{
        Name:      req.Name,
        ShiftType: model.ShiftType(req.ShiftType),  // ❌ モデル型への変換も Handler で実施
        StartTime: startTime,
        EndTime:   endTime,
        Notes:     req.Notes,
        SortOrder: req.SortOrder,
        IsActive:  isActive,
        Breaks:    breaks,
    })
    // ...
}
```

### 修正方針
`CreateShiftTemplateInput` のフィールド型を Handler から渡しやすい raw 型（string / []ShiftBreakTemplateInput）に統一し、
デフォルト値・型変換・nil ガードはすべて Service の `Create` メソッド内で処理する。

### 修正後コード

**handler/shift_template_handler.go**
```go
func (h *Handler) CreateShiftTemplate(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    var req createShiftTemplateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }
    breaks := make([]service.ShiftBreakTemplateInput, 0, len(req.Breaks))
    for _, b := range req.Breaks {
        breaks = append(breaks, service.ShiftBreakTemplateInput{BreakStart: b.BreakStart, BreakEnd: b.BreakEnd})
    }
    tpl, err := h.svc.ShiftTemplate.Create(c.Request.Context(), clinicID, &service.CreateShiftTemplateInput{
        Name:      req.Name,
        ShiftType: req.ShiftType,    // string のまま渡す
        StartTime: req.StartTime,    // string のまま渡す（空文字許容）
        EndTime:   req.EndTime,      // string のまま渡す
        Notes:     req.Notes,
        SortOrder: req.SortOrder,
        IsActive:  req.IsActive,     // *bool のまま渡す（nil = デフォルト true をサービス側で処理）
        Breaks:    breaks,
    })
    if err != nil {
        RespondError(c, err)
        return
    }
    c.Header("Location", fmt.Sprintf("/v1/masters/shift-templates/%d", tpl.ID))
    c.JSON(http.StatusCreated, toShiftTemplateResponse(tpl))
}
```

**service/shift_template_service.go（CreateShiftTemplateInput の変更）**
```go
// CreateShiftTemplateInput はシフトテンプレート作成の入力DTO
type CreateShiftTemplateInput struct {
    Name      string
    ShiftType string    // model 変換はサービス内で行う
    StartTime string    // 空文字 = nil 扱い（サービス内で normalizeTimeString を適用）
    EndTime   string
    Notes     string
    SortOrder int
    IsActive  *bool     // nil = デフォルト true
    Breaks    []ShiftBreakTemplateInput
}

func (s *shiftTemplateService) Create(ctx context.Context, clinicID uint64, input *CreateShiftTemplateInput) (*model.ShiftTemplate, error) {
    // デフォルト値処理はここで行う
    isActive := true
    if input.IsActive != nil {
        isActive = *input.IsActive
    }
    startTime := normalizeTimeString(&input.StartTime)
    endTime := normalizeTimeString(&input.EndTime)
    if err := validateShiftTimes(model.ShiftType(input.ShiftType), startTime, endTime); err != nil {
        return nil, err
    }
    tpl := &model.ShiftTemplate{
        ClinicID:  clinicID,
        Name:      input.Name,
        ShiftType: model.ShiftType(input.ShiftType),
        StartTime: startTime,
        EndTime:   endTime,
        Notes:     input.Notes,
        SortOrder: input.SortOrder,
        IsActive:  isActive,
    }
    // ... 以降は変更なし
}
```

## 備考
`UpdateShiftTemplate` ハンドラの `model.ShiftType(*req.ShiftType)` 変換（行 196〜198）についても同様に
Service 層（`UpdateShiftTemplateInput.ShiftType` を `*string` にして変換をサービス内で実施）への移動が望ましいが、
Update は既に `*model.ShiftType` 型を受け取る設計になっているため、影響範囲を考慮して任意対応とする。
