# TASK-081: slog Create — model フィールド参照（clinicID パラメータ未使用）— occupation / cage

## 優先度

LOW

---

## 概要

`occupation_service.go` と `cage_service.go` の Create メソッドで、
slog の `clinic_id` フィールドに関数パラメータ `clinicID` ではなく
生成されたエンティティの `xxx.ClinicID` フィールドを使っている。

値は同一のはずだが、他のサービス（medicine, exam_type 等）はすべて
パラメータ `clinicID` を直接参照しており、スタイルが不統一。

---

## 問題箇所

### occupation_service.go:79-81

```go
// ❌ occupation.ClinicID（モデルフィールド参照）
slog.InfoContext(ctx, "occupation created",
    slog.Uint64("clinic_id", occupation.ClinicID),  // ← パラメータを使うべき
    slog.Uint64("occupation_id", occupation.ID))
```

### cage_service.go:81-83

```go
// ❌ cage.ClinicID（モデルフィールド参照）
slog.InfoContext(ctx, "cage created",
    slog.Uint64("clinic_id", cage.ClinicID),  // ← パラメータを使うべき
    slog.Uint64("cage_id", cage.ID))
```

---

## 参照実装（medicine_service.go）

```go
// ✅ 関数パラメータを直接参照
slog.InfoContext(ctx, "medicine created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("medicine_id", medicine.ID))
```

---

## 修正方針

```go
// ✅ occupation_service.go 修正後
slog.InfoContext(ctx, "occupation created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("occupation_id", occupation.ID))

// ✅ cage_service.go 修正後
slog.InfoContext(ctx, "cage created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("cage_id", cage.ID))
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `occupation_service.go` | slog Create の `occupation.ClinicID` → `clinicID` |
| `cage_service.go` | slog Create の `cage.ClinicID` → `clinicID` |
