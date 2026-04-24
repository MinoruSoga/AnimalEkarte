# CODE-QUALITY-219: handler_test.go がコンパイルテストのみで実テスト未実装

## 概要

`inquiry_template_handler_test.go` と `permission_group_handler_test.go` の2ファイルが、
150行以上のテスト仕様コメントを持ちながら実テスト関数がゼロ（コンパイル確認用スタブのみ）。
handler 層のテストカバレッジが実質ゼロの状態。

## 該当ファイル

### 1. `inquiry_template_handler_test.go`

現状:
```go
func TestInquiryTemplateHandlerCompiles(t *testing.T) {
    // コンパイル確認のみ
}
```

実装されるべきテスト仕様（ファイル内コメントより）:
- `TestInquiryTemplateHandler_List` — 正常, サービスエラー
- `TestInquiryTemplateHandler_GetByID` — 正常, 404, 不正ID
- `TestInquiryTemplateHandler_Create` — 正常, バリデーションエラー, バインドエラー
- `TestInquiryTemplateHandler_Update` — 正常, 404, バインドエラー
- `TestInquiryTemplateHandler_Delete` — 正常, 409(依存チェック), 404
- `TestInquiryTemplateHandler_Reorder` — 正常, 空IDsエラー

### 2. `permission_group_handler_test.go`

現状:
```go
func TestPermissionGroupHandlerCompiles(t *testing.T) {
    // コンパイル確認のみ
}
```

実装されるべきテスト仕様（ファイル内コメントより）:
- `TestPermissionGroupHandler_List` — 正常, サービスエラー
- `TestPermissionGroupHandler_GetByID` — 正常, 404
- `TestPermissionGroupHandler_Create` — 正常, バリデーションエラー
- `TestPermissionGroupHandler_Update` — 正常, 404
- `TestPermissionGroupHandler_Delete` — 正常, 409(スタッフ割当済み)
- `TestPermissionGroupHandler_Reorder` — 正常, 空IDsエラー
- `TestPermissionGroupHandler_SetRules` — 正常, BUG-140(自グループ編集権限削除防止), BUG-146(リソース名バリデーション)
- `TestPermissionGroupHandler_GetEffectivePermissions` — 正常

特に `SetRules` は BUG-140・BUG-146 のリグレッション検証が必要。

## 参照実装

`trimming_course_handler_test.go` が完全な実装の参考になる:
```go
func TestTrimmingCourseHandler_List(t *testing.T) {
    tests := []struct {
        name       string
        svcFn      func(*mockTrimmingCourseService)
        wantStatus int
    }{
        {
            name: "正常: 一覧を返す",
            svcFn: func(m *mockTrimmingCourseService) {
                m.listFn = func(...) ([]model.TrimmingCourse, error) {
                    return []model.TrimmingCourse{{ID: 1, Name: "スタンダード"}}, nil
                }
            },
            wantStatus: http.StatusOK,
        },
        // ...
    }
}
```

## 優先度

HIGH — コメントで仕様が書かれているにもかかわらず実装されていない。
permission_group の `SetRules` は BUG-140・BUG-146 という重要な業務ルールを含むため、
handler 層の入力検証テストが欠落していることは回帰リスクが高い。
