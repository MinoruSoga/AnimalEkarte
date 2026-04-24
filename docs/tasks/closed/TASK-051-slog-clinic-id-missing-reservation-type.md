# TASK-051: slog clinic_id 欠落 — reservation_type / reservation_type_group（TASK-045 補完）

## 優先度

MEDIUM

---

## 概要

TASK-040・TASK-045 で複数ドメインの slog clinic_id 欠落を修正した。同パターンが `reservation_type_service.go` の Create/Update と `reservation_type_group_service.go` の Create に残存している（3 箇所）。

---

## 問題 1: reservation_type_service — Create slog に clinic_id なし

### ファイル
`backend/internal/service/reservation_type_service.go` L248（概算）

```go
// ❌ clinic_id なし
slog.InfoContext(ctx, "service type created", slog.Uint64("reservation_type_id", st.ID))
```

### 修正案
```go
slog.InfoContext(ctx, "reservation type created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("reservation_type_id", st.ID))
```

---

## 問題 2: reservation_type_service — Update slog に clinic_id なし

### ファイル
`backend/internal/service/reservation_type_service.go` L267（概算）

```go
// ❌ clinic_id なし
slog.InfoContext(ctx, "reservation type updated", slog.Uint64("reservation_type_id", id))
```

### 修正案
```go
slog.InfoContext(ctx, "reservation type updated",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("reservation_type_id", id))
```

---

## 問題 3: reservation_type_group_service — Create slog に clinic_id なし

### ファイル
`backend/internal/service/reservation_type_group_service.go` L89-91（概算）

```go
// ❌ clinic_id なし
slog.InfoContext(ctx, "reservation type group created",
    slog.Uint64("reservation_type_group_id", group.ID))
```

### 修正案
```go
slog.InfoContext(ctx, "reservation type group created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("reservation_type_group_id", group.ID))
```

---

## 備考

- reservation_type_service.go の Delete（L286）はすでに clinic_id を含んでいる。
- reservation_type_group_service.go の Update（L114-116）はすでに clinic_id を含んでいる。
- 3 箇所まとめて 1 コミットで対応してよい。
