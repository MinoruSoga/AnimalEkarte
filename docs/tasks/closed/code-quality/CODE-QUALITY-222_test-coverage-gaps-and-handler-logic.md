# CODE-QUALITY-222: テストケース欠落 + handler 日付パースロジック混入

## 概要

Round 5 スキャンで発見した複数の軽微な問題をまとめて起票する。

---

## 問題1: handler 日付パースロジック混入

**ファイル:** `backend/internal/handler/reservation_type_handler.go:183-189`

### 現状コード

```go
// CreateUnavailableTime() 内
if req.SpecificDate != nil {
    t, err := time.Parse("2006-01-02", *req.SpecificDate)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("specific_date は YYYY-MM-DD 形式で入力してください"))
        return
    }
    input.SpecificDate = &t
}
```

### 問題

日付文字列のパースと変換ロジックが handler 層に存在している。
handler の責務は「リクエストのバインド → service 呼び出し → レスポンス変換」であり、
フォーマット検証・型変換は service 層（または shared helper）の責務。

### 修正方針

`response.go` の `parseDateQuery` ヘルパー関数と同様に、
`request.go` 側に `parseDateInput(s string) (time.Time, error)` を追加し handler から呼ぶか、
service が `string` で受け取ってパース・バリデーションを行う。

**優先度:** MEDIUM（extends CODE-QUALITY-201）

---

## 問題2: service_test.go の GetByID テスト欠落

| ファイル | 欠落テスト |
|---------|-----------|
| `backend/internal/service/reservation_type_group_service_test.go` | `TestReservationTypeGroupService_GetByID` |
| `backend/internal/service/permission_group_service_test.go` | `TestPermissionGroupService_GetByID` |

### 問題

他のマスタサービスは GetByID のテストケース（正常 / 404 not found）を持つが、
上記2ファイルは GetByID テストが欠落している。

### 修正方針

参照実装: `backend/internal/service/exam_type_service_test.go` の `TestExamTypeService_GetByID`

```go
func TestReservationTypeGroupService_GetByID(t *testing.T) {
    tests := []struct {
        name    string
        repoFn  func(*mockReservationTypeGroupRepository)
        id      uint64
        wantErr bool
    }{
        {
            name: "正常: グループを返す",
            repoFn: func(m *mockReservationTypeGroupRepository) {
                m.findByIDFn = func(...) (*model.ReservationTypeGroup, error) {
                    return &model.ReservationTypeGroup{ID: 1}, nil
                }
            },
            id:      1,
            wantErr: false,
        },
        {
            name: "404: 存在しないID",
            repoFn: func(m *mockReservationTypeGroupRepository) {
                m.findByIDFn = func(...) (*model.ReservationTypeGroup, error) {
                    return nil, apperrors.WrapNotFound("not found")
                }
            },
            id:      999,
            wantErr: true,
        },
    }
}
```

**優先度:** MEDIUM

---

## 問題3: inquiry_template_service_test.go の Reorder テスト欠落

**ファイル:** `backend/internal/service/inquiry_template_service_test.go`

他の service テストは `TestXxxService_Reorder` で「空IDsはエラー」テストを持つが、
`inquiry_template_service_test.go` には Reorder テストが存在しない。

**修正方針:** `TestInquiryTemplateService_Reorder` を追加（空IDs → `ErrMsgIDsNotEmpty`、正常ケース）

**優先度:** LOW

---

## 問題4: trimming_course_response.go の TargetSize omitempty 欠落

**ファイル:** `backend/internal/handler/trimming_course_response.go:16`

```go
type trimmingCourseResponse struct {
    // ...
    Price      *int64  `json:"price,omitempty"`    // ✅ omitempty あり
    Duration   *int    `json:"duration,omitempty"` // ✅ omitempty あり
    TargetSize *string `json:"target_size"`         // ❌ omitempty なし（不整合）
    // ...
}
```

nullable なポインタフィールドで `Price`/`Duration` は `omitempty` を持つのに、
`TargetSize` だけ欠落している（CODE-QUALITY-214 での修正時に付け忘れ）。

nil の TargetSize が `"target_size": null` として JSON に出力される。
クライアントが null と省略を区別して処理していれば問題ないが、
他フィールドとの一貫性を保つべき。

**修正案:**
```go
TargetSize *string `json:"target_size,omitempty"`
```

**優先度:** LOW

---

## まとめ

| 問題 | ファイル | 優先度 |
|------|---------|--------|
| 日付パースロジック混入 | reservation_type_handler.go:183 | MEDIUM |
| GetByID テスト欠落 | reservation_type_group_service_test.go | MEDIUM |
| GetByID テスト欠落 | permission_group_service_test.go | MEDIUM |
| Reorder テスト欠落 | inquiry_template_service_test.go | LOW |
| TargetSize omitempty 欠落 | trimming_course_response.go:16 | LOW |
