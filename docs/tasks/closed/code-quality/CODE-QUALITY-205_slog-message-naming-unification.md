# CODE-QUALITY-205: slog メッセージ命名統一（アンダースコア混在・ドメイン名不統一）

## 概要

Service 層の `slog.InfoContext` メッセージに命名の不統一が複数存在する。
「アンダースコア区切り」と「スペース区切り」の混在、
`"service type"` と `"reservation_type"` の混在など。

## 優先度

MEDIUM

## 影響ファイル

| ファイル | 問題箇所 | 問題内容 |
|---------|---------|---------|
| `backend/internal/service/animal_species_service.go` | L147 | `"animal_species reordered"` → アンダースコア混在 |
| `backend/internal/service/exam_type_service.go` | L167 | `"exam_types reordered"` → アンダースコア混在 |
| `backend/internal/service/checkup_type_service.go` | L183 | `"checkup_types reordered"` → アンダースコア混在 |
| `backend/internal/service/chief_complaint_service.go` | L78, L105, L136 | `"category"` / `"type"` 混在 |
| `backend/internal/service/reservation_type_service.go` | L207,215,261,279,294,262,281,296,307 | `"service type"` / `"reservation_type"` 混在 |

---

## ルール（標準形式）

```
"[entity] [operation]"  // スペース区切り・単数形・小英字
```

例:
- `"exam type created"`
- `"exam type updated"`
- `"exam type deleted"`
- `"exam type reordered"`

---

## 修正内容

### 1. animal_species_service.go:147

```go
// 修正前
slog.InfoContext(ctx, "animal_species reordered", ...)
// 修正後
slog.InfoContext(ctx, "animal species reordered", ...)
```

---

### 2. exam_type_service.go:167

```go
// 修正前
slog.InfoContext(ctx, "exam_types reordered", ...)
// 修正後
slog.InfoContext(ctx, "exam type reordered", ...)
```

### 派生: exam_type_repository.go:88 — reorderByClinicID の resource 名

```go
// 修正前
reorderByClinicID(ctx, r.db, &model.ExaminationType{}, "exam_type", ...)
// 修正後
reorderByClinicID(ctx, r.db, &model.ExaminationType{}, "examination_type", ...)
// 理由: 同ファイルの他メソッドは全て "examination_type" を使用
```

---

### 3. checkup_type_service.go:183

```go
// 修正前
slog.InfoContext(ctx, "checkup_types reordered", ...)
// 修正後
slog.InfoContext(ctx, "checkup type reordered", ...)
```

---

### 4. chief_complaint_service.go — "category" と "type" の混在

```go
// L78 修正前
return nil, apperrors.Wrap(err, "failed to list chief complaint categories")
// 修正後
return nil, apperrors.Wrap(err, "failed to list chief complaint types")

// L105 修正前
slog.InfoContext(ctx, "chief complaint category created", ...)
// 修正後
slog.InfoContext(ctx, "chief complaint type created", ...)

// L136 修正前
slog.InfoContext(ctx, "chief_complaint_types reordered", ...)
// 修正後
slog.InfoContext(ctx, "chief complaint type reordered", ...)
```

---

### 5. reservation_type_service.go — "service type" を "reservation type" に全置換

```
"failed to list service types"    → "failed to list reservation types"
"failed to get service type"      → "failed to get reservation type"
"failed to create service type"   → "failed to create reservation type"
"failed to update service type"   → "failed to update reservation type"
"failed to delete service type"   → "failed to delete reservation type"
"service type created"            → "reservation type created"
"service type updated"            → "reservation type updated"
"service type deleted"            → "reservation type deleted"
"reservation_types reordered"     → "reservation type reordered"
```

---

## 規約参照

- `.claude/CLAUDE.md`: 構造化ログ `log/slog` の使用規約
- `.claude/rules/go-language.md`: 命名規則（6節）

## テスト

テスト不要（ログメッセージの修正のみ。動作に影響なし）。
ただし grep で修正漏れがないことを確認すること。
