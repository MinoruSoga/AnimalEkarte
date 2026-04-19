# TASK-040: slog clinic_id 欠落 — insurance / exam_type / procedure / cage Update

## 優先度

MEDIUM

---

## 概要

Create・Delete には `clinic_id` 付き slog があるのに、Update のみ `clinic_id` が欠落しているドメインが4件存在する。マルチテナント環境での障害調査・監査時に「どのクリニックで発生したか」の追跡が困難になる。

---

## 問題 1: insurance_service — Create / Update slog に clinic_id なし

### ファイル
`backend/internal/service/insurance_service.go:72, 87`

```go
// L72 Create（clinic_id なし）
slog.InfoContext(ctx, "insurance created", slog.Uint64("insurance_id", insurance.ID))

// L87 Update（clinic_id なし）
slog.InfoContext(ctx, "insurance updated", slog.Uint64("insurance_id", id))

// L101 Delete（clinic_id あり・正しい）
slog.InfoContext(ctx, "insurance deleted", slog.Uint64("insurance_id", id), slog.Uint64("clinic_id", clinicID))
```

### 修正案
```go
slog.InfoContext(ctx, "insurance created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("insurance_id", insurance.ID))

slog.InfoContext(ctx, "insurance updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("insurance_id", id))
```

---

## 問題 2: exam_type_service — Update slog に clinic_id なし

### ファイル
`backend/internal/service/exam_type_service.go:87`

```go
// L70-72 Create（clinic_id あり・正しい）
slog.InfoContext(ctx, "exam type created",
    slog.Uint64("exam_type_id", exType.ID),
    slog.Uint64("clinic_id", clinicID))

// L87 Update（clinic_id なし）
slog.InfoContext(ctx, "exam type updated", slog.Uint64("exam_type_id", id))

// L108 Delete（clinic_id あり・正しい）
slog.InfoContext(ctx, "exam type deleted", slog.Uint64("exam_type_id", id), slog.Uint64("clinic_id", clinicID))
```

### 修正案
```go
slog.InfoContext(ctx, "exam type updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("exam_type_id", id))
```

---

## 問題 3: procedure_service — Update slog に clinic_id なし

### ファイル
`backend/internal/service/procedure_service.go:110`

```go
// L92 Create（clinic_id あり・正しい）
slog.InfoContext(ctx, "procedure created", slog.Uint64("procedure_id", procedure.ID), slog.Uint64("clinic_id", clinicID))

// L110 Update（clinic_id なし）
slog.InfoContext(ctx, "procedure updated", slog.Uint64("procedure_id", id))

// L124 Delete（clinic_id あり・正しい）
slog.InfoContext(ctx, "procedure deleted", slog.Uint64("procedure_id", id), slog.Uint64("clinic_id", clinicID))
```

### 修正案
```go
slog.InfoContext(ctx, "procedure updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("procedure_id", id))
```

---

## 問題 4: cage_service — Update slog に clinic_id なし

### ファイル
`backend/internal/service/cage_service.go:92`

```go
// L75-77 Create（clinic_id あり・正しい）
slog.InfoContext(ctx, "cage created",
    slog.Uint64("cage_id", cage.ID),
    slog.Uint64("clinic_id", cage.ClinicID))

// L92 Update（clinic_id なし）
slog.InfoContext(ctx, "cage updated", slog.Uint64("cage_id", id))

// L106-108 Delete（clinic_id あり・正しい）
slog.InfoContext(ctx, "cage deleted",
    slog.Uint64("cage_id", id),
    slog.Uint64("clinic_id", clinicID))
```

### 修正案
```go
slog.InfoContext(ctx, "cage updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("cage_id", id))
```
