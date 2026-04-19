# TASK-045: slog clinic_id 欠落 — checkup_type / trimming_course / trimming_option

## 優先度

MEDIUM

---

## 概要

TASK-040 で insurance / exam_type / procedure / cage の Update slog clinic_id 欠落を指摘した。
同パターンが checkup_type と trimming ドメインの Update・Delete にも存在する（5箇所）。

---

## 問題 1: checkup_type_service — Update slog に clinic_id なし

### ファイル
`backend/internal/service/checkup_type_service.go:93`

```go
// L93 Update（clinic_id なし）
slog.InfoContext(ctx, "checkup type updated", slog.Uint64("checkup_type_id", id))

// L114 Delete（clinic_id あり・正しい）
slog.InfoContext(ctx, "checkup type deleted", slog.Uint64("checkup_type_id", id), slog.Uint64("clinic_id", clinicID))
```

### 修正案
```go
slog.InfoContext(ctx, "checkup type updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("checkup_type_id", id))
```

---

## 問題 2: trimming_course_service — Update slog に clinic_id なし

### ファイル
`backend/internal/service/trimming_master_service.go:94`

```go
// L94 Update（clinic_id なし）
slog.InfoContext(ctx, "trimming course updated", slog.Uint64("trimming_course_id", id))

// L77-79 Create（clinic_id あり・正しい）
slog.InfoContext(ctx, "trimming course created",
    slog.Uint64("trimming_course_id", course.ID),
    slog.Uint64("clinic_id", clinicID))
```

### 修正案
```go
slog.InfoContext(ctx, "trimming course updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("trimming_course_id", id))
```

---

## 問題 3: trimming_course_service — Delete slog に clinic_id なし

### ファイル
`backend/internal/service/trimming_master_service.go:108`

```go
// L108 Delete（clinic_id なし）
slog.InfoContext(ctx, "trimming course deleted", slog.Uint64("trimming_course_id", id))
```

### 修正案
```go
slog.InfoContext(ctx, "trimming course deleted",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("trimming_course_id", id))
```

---

## 問題 4: trimming_option_service — Update slog に clinic_id なし

### ファイル
`backend/internal/service/trimming_master_service.go:238`

```go
// L238 Update（clinic_id なし）
slog.InfoContext(ctx, "trimming option updated", slog.Uint64("trimming_option_id", id))
```

### 修正案
```go
slog.InfoContext(ctx, "trimming option updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("trimming_option_id", id))
```

---

## 問題 5: trimming_option_service — Delete slog に clinic_id なし

### ファイル
`backend/internal/service/trimming_master_service.go:252`

```go
// L252 Delete（clinic_id なし）
slog.InfoContext(ctx, "trimming option deleted", slog.Uint64("trimming_option_id", id))
```

### 修正案
```go
slog.InfoContext(ctx, "trimming option deleted",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("trimming_option_id", id))
```
