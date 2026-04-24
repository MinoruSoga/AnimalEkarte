# TASK-099: Create ハンドラ — Location ヘッダー欠落（animal_species / shift_template / reservation_type / consultation）

## 優先度

MEDIUM

---

## 概要

以下の Create ハンドラが 201 Created を返すが、`Location` ヘッダーを設定していない。
TASK-073 でクローズ済みの 11 ハンドラとは別に残存していた未対応箇所。

---

## 問題箇所

### animal_species_handler.go:54

```go
// ❌ Location ヘッダーなしで 201 返却
c.JSON(http.StatusCreated, toAnimalSpeciesResponse(species))
```

### shift_template_handler.go:155

```go
// ❌ Location ヘッダーなしで 201 返却
c.JSON(http.StatusCreated, toShiftTemplateResponse(tpl))
```

### reservation_type_handler.go:81（CreateReservationType）

```go
// ❌ Location ヘッダーなしで 201 返却
c.JSON(http.StatusCreated, toReservationTypeResponse(st))
```

### reservation_type_handler.go:194（CreateUnavailableTime）

```go
// ❌ Location ヘッダーなしで 201 返却
c.JSON(http.StatusCreated, resp)
```

### consultation_handler.go:83（CreateConsultation）

```go
// ❌ Location ヘッダーなしで 201 返却
c.JSON(http.StatusCreated, toConsultationResponse(consultation))
```

---

## 参照実装（exam_type_handler.go 等）

```go
// ✅ 正しい実装パターン
examType, err := h.svc.ExamType.Create(ctx, clinicID, &service.CreateExamTypeInput{...})
if err != nil {
    RespondError(c, err)
    return
}
c.Header("Location", fmt.Sprintf("/v1/masters/exam-types/%d", examType.ID))
c.JSON(http.StatusCreated, toExamTypeResponse(examType))
```

---

## 修正方針

各 Create ハンドラで `c.JSON` の直前に `c.Header("Location", ...)` を追加する。

```go
// ✅ animal_species_handler.go
c.Header("Location", fmt.Sprintf("/v1/masters/animal-species/%d", species.ID))
c.JSON(http.StatusCreated, toAnimalSpeciesResponse(species))

// ✅ shift_template_handler.go
c.Header("Location", fmt.Sprintf("/v1/shift-templates/%d", tpl.ID))
c.JSON(http.StatusCreated, toShiftTemplateResponse(tpl))

// ✅ reservation_type_handler.go (CreateReservationType)
c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d", st.ID))
c.JSON(http.StatusCreated, toReservationTypeResponse(st))

// ✅ reservation_type_handler.go (CreateUnavailableTime)
c.Header("Location", fmt.Sprintf("/v1/masters/reservation-types/%d/unavailable-times/%d", reservationTypeID, result.ID))
c.JSON(http.StatusCreated, resp)
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `handler/animal_species_handler.go` | CreateAnimalSpecies に Location ヘッダー追加 |
| `handler/shift_template_handler.go` | CreateShiftTemplate に Location ヘッダー追加 |
| `handler/reservation_type_handler.go` | CreateReservationType と CreateUnavailableTime に Location ヘッダー追加 |
| `handler/consultation_handler.go` | CreateConsultation に Location ヘッダー追加 |

---

## 関連

- TASK-073: 他の 11 マスタ Create ハンドラの同種問題（クローズ済み）
