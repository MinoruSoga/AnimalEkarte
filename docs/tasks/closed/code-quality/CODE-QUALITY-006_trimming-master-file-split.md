# CODE-QUALITY-006: trimming_master ファイル分割（Course / Option 単一責任化）

## 概要

`trimming_master_handler.go` / `trimming_master_service.go` / `trimming_master_repository.go` に
`TrimmingCourse` と `TrimmingOption` という2つの完全に独立したエンティティが同一ファイルに混在している。  
両者は FK 参照も依存関係もなく、1エンティティ = 1ファイルのプロジェクト規約に違反している。

## 優先度

MEDIUM

## 影響ファイル

| 現状ファイル | 行数 | 問題 |
|---|---|---|
| `backend/internal/handler/trimming_master_handler.go` | 282行 | Course + Option の2エンティティを含む |
| `backend/internal/handler/trimming_master_request.go` | 41行 | Course + Option のリクエスト型を含む |
| `backend/internal/handler/trimming_master_response.go` | 71行 | Course + Option のレスポンス型を含む |
| `backend/internal/service/trimming_master_service.go` | 328行 | Course + Option の2インターフェース・2実装を含む |
| `backend/internal/repository/trimming_master_repository.go` | 187行 | Course + Option の2インターフェース・2実装を含む |

---

## 分割後の構成

### Handler 層

```
handler/
├── trimming_course_handler.go      (List, Get, Create, Update, Delete, Reorder)
├── trimming_course_request.go
├── trimming_course_response.go
├── trimming_option_handler.go      (List, Get, Create, Update, Delete, Reorder)
├── trimming_option_request.go
└── trimming_option_response.go
```

### Service 層

```
service/
├── trimming_course_service.go      (TrimmingCourseService interface + 実装)
└── trimming_option_service.go      (TrimmingOptionService interface + 実装)
```

### Repository 層

```
repository/
├── trimming_course_repository.go   (TrimmingCourseRepository interface + 実装)
└── trimming_option_repository.go   (TrimmingOptionRepository interface + 実装)
```

---

## 修正方針

1. 各ファイルから Course / Option 関連コードを抽出し、新ファイルに移動
2. `service/service.go` の `Services` struct のフィールドは変更不要（既に分離済みの可能性あり）
3. `cmd/api/main.go` の DI 配線でファイル名変更による影響がないことを確認
4. ルート登録（`staff_handler.go` 内の `RegisterMasterRoutes`）の変更は不要（ハンドラメソッドの名前は維持）
5. テストファイルも同様に分割:
   - `trimming_master_handler_test.go` → `trimming_course_handler_test.go` + `trimming_option_handler_test.go`
   - `trimming_master_service_test.go` → `trimming_course_service_test.go` + `trimming_option_service_test.go`

---

## 対象外（分割不要）

`diagnosis_handler.go` / `diagnosis_service.go` / `diagnosis_repository.go` は
`DiagnosisType → DiagnosisName` の FK 親子関係があり、`diagnosisNameService` が `typeRepo` を依存として持つため、
現状の1ファイル配置が合理的。分割対象外とする。

---

## 規約参照

- `.claude/rules/naming-conventions.md`: 「原則1エンティティ = 1ファイルセット」
- `.claude/rules/code-style.md`: ファイル命名規則（snake_case）

## テスト

- 分割後も全既存テスト（917行の handler_test + 792行の service_test）が PASS することを確認
- DI 配線が正しく動作することを統合テストで確認
