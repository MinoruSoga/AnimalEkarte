# TASK-095: slog — モデルフィールド使用 + "species_id" 命名違反（inquiry_template / animal_species）

## 優先度

LOW

---

## 概要

2 つの service ファイルで slog に関する規約違反が存在する。

1. `inquiry_template_service.go`: `clinicID` パラメータの代わりにモデルフィールド（`template.ClinicID`）を使用（TASK-081 と同パターンだが未起票）
2. `animal_species_service.go`: エンティティ ID フィールド名が `"species_id"` で、規約の `"animal_species_id"` と不一致（TASK-089 の追加対象）

---

## 問題箇所

### 1. inquiry_template_service.go:83

```go
// ❌ template.ClinicID（モデルフィールド）を使用
slog.InfoContext(ctx, "inquiry template created",
    slog.Uint64("clinic_id", template.ClinicID),  // ← clinicID パラメータを使うべき
    slog.Uint64("inquiry_template_id", template.ID))

// ✅ 修正後: パラメータを直接使用
slog.InfoContext(ctx, "inquiry template created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("inquiry_template_id", template.ID))
```

**理由**: Create 前に `template.ClinicID = clinicID` を設定しているため実害はないが、
パラメータ直接参照が規約（TASK-081 参照: occupation/cage 同パターン）。

---

### 2. animal_species_service.go: 全 slog 呼び出し

```go
// ❌ L84-86 (Create): "species_id" — "animal_species_id" であるべき
slog.InfoContext(ctx, "animal species created",
    slog.Uint64("species_id", species.ID),   // ← 非準拠
    slog.String("name", species.Name))

// ❌ L102-103 (Update): "species_id"
slog.InfoContext(ctx, "animal species updated",
    slog.Uint64("species_id", id))           // ← 非準拠

// ❌ L118 (Delete): "species_id"
slog.InfoContext(ctx, "animal species deleted",
    slog.Uint64("species_id", id))           // ← 非準拠
```

規約: エンティティ ID フィールド名は `{ドメイン名}_id` 形式（TASK-089 参照）。
`animal_species` ドメインなら `"animal_species_id"` が正しい。

また `Create` ログの `slog.String("name", ...)` は TASK-089 で問題視されたパターンと同様
（entity_id 以外の余分なフィールド）。

---

## 修正方針

```go
// ✅ inquiry_template_service.go:83
slog.Uint64("clinic_id", clinicID),  // モデルフィールド→パラメータ

// ✅ animal_species_service.go 全箇所
slog.Uint64("animal_species_id", species.ID)  // "species_id"→"animal_species_id"
slog.Uint64("animal_species_id", id)          // 同上
// name フィールドは削除（entity_id のみが規約）
```

---

## 修正ファイル

| ファイル | 行 | 修正内容 |
|---------|-----|---------|
| `service/inquiry_template_service.go` | L83 | `template.ClinicID` → `clinicID` |
| `service/animal_species_service.go` | L85,103,118 | `"species_id"` → `"animal_species_id"` |
| `service/animal_species_service.go` | L86 | `slog.String("name", ...)` 削除 |

---

## 関連

- TASK-081: occupation/cage の slog でモデルフィールド使用（同パターン）
- TASK-089: slog エンティティ ID フィールド命名不統一（複数サービス）
