# TASK-082: WrapAlreadyExists vs WrapConflict — 重複エラーパターン不統一

## 優先度

MEDIUM

---

## 概要

重複キー違反（Unique 制約違反）発生時のエラー生成関数が、リポジトリによって
`WrapAlreadyExists(resource, identifier)` と `WrapConflict("同じ名称が既に登録されています")` に
分裂している。

クライアントが受け取る JSON エラーメッセージが日本語統一されないため、FE での
エラーメッセージ表示が不統一になる。

---

## 現状

### WrapConflict を使う（✅ 正しいパターン）

| ファイル | メッセージ |
|---------|---------|
| `occupation_repository.go` | `"同じ名称が既に登録されています"` |
| `procedure_repository.go` | `"同じ名称が既に登録されています"` |
| `exam_type_repository.go` | `"同じ名称が既に登録されています"` |
| `checkup_type_repository.go` | `"同じ名称が既に登録されています"` |
| `insurance_repository.go` | `"同じ名称が既に登録されています"` |
| `diagnosis_repository.go` | `"同じ名称が既に登録されています"` |
| `permission_group_repository.go` | `"同じ名称が既に登録されています"` |
| `consultation_repository.go` | `"同じ名称が既に登録されています"` |
| `reservation_type_group_repository.go` | `"同じ名称のグループが既に登録されています"` |

### WrapAlreadyExists を使う（❌ 不統一）

| ファイル | 問題 |
|---------|------|
| `inquiry_template_repository.go:56` | `apperrors.WrapAlreadyExists("inquiry_template", template.Title)` |
| `shift_template_repository.go:60,74` | `apperrors.WrapAlreadyExists("shift_template", ...)` |

---

## 問題の影響

`WrapAlreadyExists` は英語形式のメッセージ（例: `"inquiry_template already exists: テンプレート名"`）を
生成する。`WrapConflict` は日本語メッセージを生成する。

FE のエラーハンドリングがメッセージ文字列に依存している場合、表示の一貫性が失われる。

---

## 修正方針

```go
// ❌ 修正前: inquiry_template_repository.go:56
return apperrors.WrapAlreadyExists("inquiry_template", template.Title)

// ✅ 修正後
return apperrors.WrapConflict("同じ名称が既に登録されています")
```

```go
// ❌ 修正前: shift_template_repository.go:60
return apperrors.WrapAlreadyExists("shift_template", tpl.Name)

// ✅ 修正後
return apperrors.WrapConflict("同じ名称が既に登録されています")
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `inquiry_template_repository.go` | `WrapAlreadyExists` → `WrapConflict("同じ名称が既に登録されています")` |
| `shift_template_repository.go` | 同上（2箇所） |
